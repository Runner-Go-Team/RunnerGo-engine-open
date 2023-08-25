package execution

import (
	"encoding/json"
	"fmt"
	"github.com/Runner-Go-Team/RunnerGo-engine-open/constant"
	"github.com/Runner-Go-Team/RunnerGo-engine-open/log"
	"github.com/Runner-Go-Team/RunnerGo-engine-open/middlewares"
	"github.com/Runner-Go-Team/RunnerGo-engine-open/model"
	"github.com/Runner-Go-Team/RunnerGo-engine-open/server/golink"
	"go.mongodb.org/mongo-driver/mongo"
	"sync"
	"time"
)

// RPSModel 响应时间模式
func RPSModel(scene model.Scene, configuration *model.Configuration, reportMsg *model.ResultDataMsg, resultDataMsgCh chan *model.ResultDataMsg, requestCollection *mongo.Collection) string {
	startConcurrent := scene.ConfigTask.ModeConf.StartConcurrency
	step := scene.ConfigTask.ModeConf.Step
	maxConcurrent := scene.ConfigTask.ModeConf.MaxConcurrency
	stepRunTime := scene.ConfigTask.ModeConf.StepRunTime
	stableDuration := scene.ConfigTask.ModeConf.Duration
	concurrent := startConcurrent

	target := 0

	adjustKey := fmt.Sprintf("SubscriptionStressPlanStatusChange:%s", reportMsg.ReportId)
	pubSub := model.SubscribeMsg(adjustKey)
	statusCh := pubSub.Channel()
	defer pubSub.Close()
	debug := scene.ConfigTask.DebugMode
	key := fmt.Sprintf("reportData:%s", reportMsg.ReportId)
	// 创建es客户端
	concurrentMap := new(sync.Map)
	currentWg := &sync.WaitGroup{}
	targetTime, startTime, endTime := time.Now().Unix(), time.Now().Unix(), time.Now().Unix()
	rpsTag := false
	rpsMap := make(map[string]bool)
	for _, nodes := range scene.NodesRound {
		if nodes != nil && len(nodes) > 0 {
			for _, node := range nodes {
				if node.RequestThreshold > 0 {
					rpsMap[node.Id] = true
				}
			}
		}

	}
	switch scene.ConfigTask.ControlMode {
	case constant.CentralizedMode:
		for startTime+stepRunTime > endTime {
			// 如果所有的接口rps都达到阈值，则不在进行查询当前rps
			if !rpsTag {
				res := model.QueryReportData(key)
				if res != "" {
					var result = new(model.RedisSceneTestResultDataMsg)
					err := json.Unmarshal([]byte(res), result)
					if err != nil {
						break
					}

					for _, resultData := range result.Results {
						if _, ok := rpsMap[resultData.EventId]; ok {
							if resultData.Rps > float64(resultData.RequestThreshold) {
								delete(rpsMap, resultData.EventId)
							}
						}
					}
					// 如果所有的接口rps都达到阈值，那么直接进入最大并发数
					if len(rpsMap) == 0 {
						concurrent = maxConcurrent
						stepRunTime = stableDuration
						rpsTag = true
					}
				}

			}

			select {
			case c := <-statusCh:
				log.Logger.Debug("接收到manage消息：  ", c.String())
				var subscriptionStressPlanStatusChange = new(model.SubscriptionStressPlanStatusChange)
				_ = json.Unmarshal([]byte(c.Payload), subscriptionStressPlanStatusChange)
				if subscriptionStressPlanStatusChange.MachineModeConf == nil {
					continue
				}
				log.Logger.Debug(fmt.Sprintf("%s报告，手动修改为：  %s", scene.ReportId, c.Payload))
				switch subscriptionStressPlanStatusChange.Type {
				case constant.StopPlan:
					if subscriptionStressPlanStatusChange.StopPlan == "stop" {
						return fmt.Sprintf("最大并发数：%d， 总运行时长%ds, 任务手动结束！", concurrent, endTime-targetTime)
					}
				case constant.DebugStatus:
					debug = subscriptionStressPlanStatusChange.Debug
				case constant.ReportChange:
					MachineModeConf := subscriptionStressPlanStatusChange.MachineModeConf
					if MachineModeConf.Machine != middlewares.LocalIp {
						continue
					}
					modeConf := MachineModeConf.ModeConf
					if modeConf.StartConcurrency > 0 {
						concurrent = modeConf.StartConcurrency
					}
					if modeConf.StepRunTime > 0 {
						stepRunTime = modeConf.StepRunTime
						startTime = time.Now().Unix()
					}
					if modeConf.Step > 0 {
						step = modeConf.Step
					}
					if modeConf.MaxConcurrency > 0 {
						maxConcurrent = modeConf.MaxConcurrency
						if maxConcurrent < concurrent {
							diff := concurrent - modeConf.Concurrency
							// 将最后的几个并发从map中去掉
							for i := int64(0); i < diff; i++ {
								//stopGo := fmt.Sprintf("stop:%d", concurrent-1-i)
								//concurrentMap.Store(stopGo, true)
								concurrentMap.Delete(concurrent - 1 - i)
							}
						}
					}
					if modeConf.MaxConcurrency > 0 {
						stableDuration = modeConf.MaxConcurrency
					}
					if modeConf.Duration > 0 {
						stableDuration = modeConf.Duration
					}
				}

			default:
				scene.Debug = debug
				for i := int64(0); i < concurrent; i++ {
					if _, isOk := concurrentMap.Load(i); isOk {
						continue
					}
					concurrentMap.Store(i, true)
					currentWg.Add(1)
					go func(concurrentId, concurrent int64, useConfiguration *model.Configuration, currentScene model.Scene) {
						defer currentWg.Done()
						defer concurrentMap.Delete(concurrentId)
						golink.DisposeScene(constant.PlanType, scene, useConfiguration, reportMsg, resultDataMsgCh, requestCollection, concurrentId, concurrent)

					}(i, concurrent, configuration, scene)
				}
				currentWg.Wait()
			}

			// 如果发送的并发数时间小于1000ms，那么休息剩余的时间;也就是说每秒只发送concurrent个请求
			endTime = time.Now().Unix()
			if concurrent == maxConcurrent && stepRunTime == stableDuration && startTime+stepRunTime <= endTime {
				return fmt.Sprintf("最大并发数：%d， 总运行时长%ds, 任务正常结束！", concurrent, endTime-targetTime)
			}

			// 如果当前并发数小于最大并发数，
			if concurrent < maxConcurrent {
				if endTime-startTime >= stepRunTime {
					// 从开始时间算起，加上持续时长。如果大于现在是的时间，说明已经运行了持续时长的时间，那么就要将开始时间的值，变为现在的时间值
					concurrent = concurrent + step
					if concurrent > maxConcurrent {
						concurrent = maxConcurrent
					}

					if concurrent < maxConcurrent {
						startTime = endTime

					}
				}

			}
			if concurrent == maxConcurrent {
				if target == 0 {
					stepRunTime = stableDuration
					startTime = endTime
				}
				target++
			}

		}
	case constant.AloneMode:
		for startTime+stepRunTime > endTime {
			// 如果所有的接口rps都达到阈值，则不在进行查询当前rps
			if !rpsTag {
				res := model.QueryReportData(key)
				if res != "" {
					var result = new(model.RedisSceneTestResultDataMsg)
					err := json.Unmarshal([]byte(res), result)
					if err != nil {
						break
					}

					for _, resultData := range result.Results {
						if _, ok := rpsMap[resultData.EventId]; ok {
							if resultData.Rps > float64(resultData.RequestThreshold) {
								delete(rpsMap, resultData.EventId)
							}
						}
					}
					// 如果所有的接口rps都达到阈值，那么直接进入最大并发数
					if len(rpsMap) == 0 {
						concurrent = maxConcurrent
						stepRunTime = stableDuration
						rpsTag = true
					}
				}

			}

			select {
			case c := <-statusCh:
				log.Logger.Debug("接收到manage消息：  ", c.String())
				var subscriptionStressPlanStatusChange = new(model.SubscriptionStressPlanStatusChange)
				_ = json.Unmarshal([]byte(c.Payload), subscriptionStressPlanStatusChange)
				if subscriptionStressPlanStatusChange.MachineModeConf == nil {
					continue
				}
				log.Logger.Debug(fmt.Sprintf("%s报告，手动修改为：  %s", scene.ReportId, c.Payload))
				switch subscriptionStressPlanStatusChange.Type {
				case constant.StopPlan:
					if subscriptionStressPlanStatusChange.StopPlan == "stop" {
						concurrentMap.Range(func(key, value any) bool {
							concurrentMap.Delete(key)
							return true
						})
						return fmt.Sprintf("最大并发数：%d， 总运行时长%ds, 任务手动结束！", concurrent, endTime-targetTime)
					}
				case constant.DebugStatus:
					debug = subscriptionStressPlanStatusChange.Debug
				case constant.ReportChange:
					MachineModeConf := subscriptionStressPlanStatusChange.MachineModeConf
					if MachineModeConf.Machine != middlewares.LocalIp {
						continue
					}
					modeConf := MachineModeConf.ModeConf
					if modeConf.StartConcurrency > 0 {
						concurrent = modeConf.StartConcurrency
					}
					if modeConf.StepRunTime > 0 {
						stepRunTime = modeConf.StepRunTime
						startTime = time.Now().Unix()
					}
					if modeConf.Step > 0 {
						step = modeConf.Step
					}
					if modeConf.MaxConcurrency > 0 {
						maxConcurrent = modeConf.MaxConcurrency
						if maxConcurrent < concurrent {
							diff := concurrent - modeConf.Concurrency
							// 将最后的几个并发从map中去掉
							for i := int64(0); i < diff; i++ {
								//stopGo := fmt.Sprintf("stop:%d", concurrent-1-i)
								//concurrentMap.Store(stopGo, true)
								concurrentMap.Delete(concurrent - 1 - i)
							}
						}
					}
					if modeConf.MaxConcurrency > 0 {
						stableDuration = modeConf.MaxConcurrency
					}
					if modeConf.Duration > 0 {
						stableDuration = modeConf.Duration
					}
				}

			default:
				scene.Debug = debug
				for i := int64(0); i < concurrent; i++ {
					if _, isOk := concurrentMap.Load(i); isOk {
						continue
					}
					concurrentMap.Store(i, true)
					currentWg.Add(1)
					go func(concurrentId int64, useConfiguration *model.Configuration, currentScene model.Scene) {
						defer currentWg.Done()
						defer concurrentMap.Delete(concurrentId)
						for startTime+stepRunTime > endTime {
							if _, ok := concurrentMap.Load(concurrentId); !ok {
								break
							}
							golink.DisposeScene(constant.PlanType, scene, useConfiguration, reportMsg, resultDataMsgCh, requestCollection, concurrentId, concurrent)

						}
					}(i, configuration, scene)
				}
			}

			// 如果发送的并发数时间小于1000ms，那么休息剩余的时间;也就是说每秒只发送concurrent个请求
			endTime = time.Now().Unix()
			if concurrent == maxConcurrent && stepRunTime == stableDuration && startTime+stepRunTime <= endTime {
				return fmt.Sprintf("最大并发数：%d， 总运行时长%ds, 任务正常结束！", concurrent, endTime-targetTime)
			}

			// 如果当前并发数小于最大并发数，
			if concurrent < maxConcurrent {
				if endTime-startTime >= stepRunTime {
					// 从开始时间算起，加上持续时长。如果大于现在是的时间，说明已经运行了持续时长的时间，那么就要将开始时间的值，变为现在的时间值
					concurrent = concurrent + step
					if concurrent > maxConcurrent {
						concurrent = maxConcurrent
					}

					if concurrent < maxConcurrent {
						startTime = endTime

					}
				}

			}
			if concurrent == maxConcurrent {
				if target == 0 {
					stepRunTime = stableDuration
					startTime = endTime
				}
				target++
			}

		}
		currentWg.Wait()

	}
	// 只要开始时间+持续时长大于当前时间就继续循环

	return fmt.Sprintf("最大并发数：%d， 总运行时长%ds, 任务非正常结束！", concurrent, time.Now().Unix()-targetTime)

}
