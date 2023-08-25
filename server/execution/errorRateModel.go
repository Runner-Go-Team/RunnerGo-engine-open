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

type ErrorRateData struct {
	PlanId  string `json:"planId"`
	SceneId string `json:"sceneId"`
	Apis    []Apis `json:"apis"`
}

type Apis struct {
	ApiName   string  `json:"apiName"`
	Threshold float64 `json:"threshold"`
	Actual    float64 `json:"actual"`
}

// ErrorRateModel 错误率模式
func ErrorRateModel(scene model.Scene, configuration *model.Configuration, reportMsg *model.ResultDataMsg, resultDataMsgCh chan *model.ResultDataMsg, requestCollection *mongo.Collection) string {
	startConcurrent := scene.ConfigTask.ModeConf.StartConcurrency
	step := scene.ConfigTask.ModeConf.Step
	maxConcurrent := scene.ConfigTask.ModeConf.MaxConcurrency
	stepRunTime := scene.ConfigTask.ModeConf.StepRunTime
	stableDuration := scene.ConfigTask.ModeConf.Duration

	adjustKey := fmt.Sprintf("SubscriptionStressPlanStatusChange:%s", reportMsg.ReportId)
	pubSub := model.SubscribeMsg(adjustKey)
	statusCh := pubSub.Channel()
	defer pubSub.Close()
	debug := scene.ConfigTask.DebugMode
	// 定义一个chan, 从es中获取当前错误率与阈值分别是多少

	// preConcurrent 是为了回退,此功能后续开发
	//preConcurrent := startConcurrent
	concurrent := startConcurrent
	// 只要开始时间+持续时长大于当前时间就继续循环
	target := 0
	key := fmt.Sprintf("reportData:%s", reportMsg.ReportId)
	currentWg := &sync.WaitGroup{}
	concurrentMap := new(sync.Map)
	targetTime, startTime, endTime := time.Now().Unix(), time.Now().Unix(), time.Now().Unix()

	switch scene.ConfigTask.ControlMode {
	case constant.CentralizedMode:
		for startTime+stepRunTime > endTime {

			// 查询当前错误率时多少
			//GetErrorRate(planId+":"+sceneId+":"+"errorRate", errorRateData)
			res := model.QueryReportData(key)
			if res != "" {
				var result = new(model.RedisSceneTestResultDataMsg)
				err := json.Unmarshal([]byte(res), result)
				if err != nil {
					break
				}
				for _, resultData := range result.Results {
					if resultData.TotalRequestNum > 0 {
						errRate := float64(resultData.ErrorNum) / float64(resultData.TotalRequestNum)
						if errRate > resultData.ErrorThreshold {
							return fmt.Sprintf("最大并发数：%d， 总运行时长%ds, 接口：%s, 错误率为：%f, 大于等于阈值：%f， 任务结束！", concurrent, endTime-targetTime, resultData.Name, errRate, resultData.ErrorThreshold)
						}
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
					if modeConf.MaxConcurrency > 0 {
						maxConcurrent = modeConf.MaxConcurrency
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
						golink.DisposeScene(constant.PlanType, currentScene, useConfiguration, reportMsg, resultDataMsgCh, requestCollection, concurrentId, concurrent)

					}(i, concurrent, configuration, scene)
				}
				currentWg.Wait()

			}

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
					target++
					stepRunTime = stableDuration
					startTime = endTime
				}

			}

		}
	case constant.AloneMode:
		for startTime+stepRunTime > endTime {

			// 查询当前错误率时多少
			//GetErrorRate(planId+":"+sceneId+":"+"errorRate", errorRateData)
			res := model.QueryReportData(key)
			if res != "" {
				var result = new(model.RedisSceneTestResultDataMsg)
				err := json.Unmarshal([]byte(res), result)
				if err != nil {
					break
				}
				for _, resultData := range result.Results {
					if resultData.TotalRequestNum > 0 {
						errRate := float64(resultData.ErrorNum) / float64(resultData.TotalRequestNum)
						if errRate > resultData.ErrorThreshold {
							return fmt.Sprintf("最大并发数：%d， 总运行时长%ds, 接口：%s, 错误率为：%f, 大于等于阈值：%f， 任务结束！", concurrent, endTime-targetTime, resultData.Name, errRate, resultData.ErrorThreshold)
						}
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
						return fmt.Sprintf("并发数：%d, 总运行时长%ds, 任务手动结束！", concurrent, time.Now().Unix()-targetTime)
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
							if _, isOk := concurrentMap.Load(concurrentId); !isOk {
								break
							}
							currentScene.Debug = debug
							var sceneWg = &sync.WaitGroup{}
							golink.DisposeScene(constant.PlanType, currentScene, useConfiguration, reportMsg, resultDataMsgCh, requestCollection, concurrentId, concurrent)
							sceneWg.Wait()
						}

					}(i, configuration, scene)
				}

			}

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
					target++
					stepRunTime = stableDuration
					startTime = endTime
				}

			}

		}
		currentWg.Wait()
	}

	return fmt.Sprintf("最大并发数：%d， 总运行时长%ds, 任务非正常结束！", concurrent, time.Now().Unix()-targetTime)
}
