package execution

import (
	"RunnerGo-engine/log"
	"RunnerGo-engine/middlewares"
	"RunnerGo-engine/model"
	"RunnerGo-engine/server/golink"
	"encoding/json"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"sync"
	"time"
)

// LadderModel 阶梯模式
func LadderModel(wg *sync.WaitGroup, scene model.Scene, configuration *model.Configuration, reportMsg *model.ResultDataMsg, resultDataMsgCh chan *model.ResultDataMsg, debugCollection, requestCollection *mongo.Collection) string {
	startConcurrent := scene.ConfigTask.ModeConf.StartConcurrency
	step := scene.ConfigTask.ModeConf.Step
	maxConcurrent := scene.ConfigTask.ModeConf.MaxConcurrency
	stepRunTime := scene.ConfigTask.ModeConf.StepRunTime
	stableDuration := scene.ConfigTask.ModeConf.Duration
	// 连接es，并查询当前错误率为多少，并将其放入到chan中

	// preConcurrent 是为了回退,此功能后续开发
	//preConcurrent := startConcurrent
	concurrent := startConcurrent
	target := 0
	currentWg := &sync.WaitGroup{}
	targetTime, startTime, endTime := time.Now().Unix(), time.Now().Unix(), time.Now().Unix()
	adjustKey := fmt.Sprintf("SubscriptionStressPlanStatusChange:%s:%s:%s", reportMsg.TeamId, reportMsg.PlanId, reportMsg.ReportId)
	pubSub := model.SubscribeMsg(adjustKey)
	statusCh := pubSub.Channel()
	defer pubSub.Close()
	debug := scene.Debug
	concurrentMap := new(sync.Map)
	// 只要开始时间+持续时长大于当前时间就继续循环
	switch scene.ConfigTask.ControlMode {
	case model.CentralizedMode:
		for startTime+stepRunTime > endTime {

			select {
			case c := <-statusCh:
				var subscriptionStressPlanStatusChange = new(model.SubscriptionStressPlanStatusChange)
				_ = json.Unmarshal([]byte(c.Payload), subscriptionStressPlanStatusChange)
				if subscriptionStressPlanStatusChange.MachineModeConf == nil {
					continue
				}
				log.Logger.Debug(fmt.Sprintf("%s报告，手动修改为：  %s", scene.ReportId, c.Payload))
				switch subscriptionStressPlanStatusChange.Type {
				case model.StopPlan:
					if subscriptionStressPlanStatusChange.StopPlan == "stop" {
						return fmt.Sprintf("最大并发数：%d， 总运行时长%ds, 任务手动结束！", concurrent, endTime-targetTime)
					}
				case model.DebugStatus:
					debug = subscriptionStressPlanStatusChange.Debug
				case model.ReportChange:
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
				// 查询是否在报告中对任务模式进行修改
				for i := int64(0); i < concurrent; i++ {
					if _, isOk := concurrentMap.Load(i); isOk {
						continue
					}
					concurrentMap.Store(i, true)
					wg.Add(1)
					currentWg.Add(1)
					go func(concurrentId, concurrent int64, wg *sync.WaitGroup, useConfiguration *model.Configuration) {
						var sceneWg = &sync.WaitGroup{}
						golink.DisposeScene(wg, currentWg, sceneWg, model.PlanType, scene, useConfiguration, reportMsg, resultDataMsgCh, requestCollection, concurrentId, concurrent)
						sceneWg.Wait()
						wg.Done()
						currentWg.Done()
						concurrentMap.Delete(concurrentId)
					}(i, concurrent, wg, configuration)
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
					// 从开始时间算起，加上持续时长。如果大于现在的时间，说明已经运行了持续时长的时间，那么就要将开始时间的值，变为现在的时间值
					concurrent = concurrent + step
					if concurrent > maxConcurrent {
						concurrent = maxConcurrent
					}
					if startTime+stepRunTime <= endTime && concurrent < maxConcurrent {
						startTime = endTime + stepRunTime
					}
				}
			}
			if concurrent == maxConcurrent {
				if target == 0 {
					target++
					stepRunTime = stableDuration
					startTime = endTime + stepRunTime
				}

			}

		}
	case model.AloneMode:
		for startTime+stepRunTime > endTime {
			select {
			case c := <-statusCh:
				var subscriptionStressPlanStatusChange = new(model.SubscriptionStressPlanStatusChange)
				_ = json.Unmarshal([]byte(c.Payload), subscriptionStressPlanStatusChange)
				if subscriptionStressPlanStatusChange.MachineModeConf == nil {
					continue
				}
				log.Logger.Debug(fmt.Sprintf("%s报告，手动修改为：  %s", scene.ReportId, c.Payload))
				switch subscriptionStressPlanStatusChange.Type {
				case model.StopPlan:
					if subscriptionStressPlanStatusChange.StopPlan == "stop" {
						return fmt.Sprintf("最大并发数：%d， 总运行时长%ds, 任务手动结束！", concurrent, endTime-targetTime)
					}
				case model.DebugStatus:
					debug = subscriptionStressPlanStatusChange.Debug
				case model.ReportChange:
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
				// 查询是否在报告中对任务模式进行修改
				for i := int64(0); i < concurrent; i++ {
					if _, isOk := concurrentMap.Load(i); isOk {
						continue
					}
					concurrentMap.Store(i, true)
					wg.Add(1)
					currentWg.Add(1)
					go func(concurrentId int64, wg *sync.WaitGroup, useConfiguration *model.Configuration, currentScene model.Scene) {

						for startTime+stepRunTime > endTime {
							if _, ok := concurrentMap.Load(concurrentId); !ok {
								break
							}
							var sceneWg = &sync.WaitGroup{}
							golink.DisposeScene(wg, currentWg, sceneWg, model.PlanType, currentScene, useConfiguration, reportMsg, resultDataMsgCh, requestCollection, i, concurrent)
							sceneWg.Wait()

						}
						concurrentMap.Delete(concurrentId)
						currentWg.Done()
						wg.Done()

					}(i, wg, configuration, scene)
				}
			}
			endTime = time.Now().Unix()
			if concurrent == maxConcurrent && stepRunTime == stableDuration && startTime+stepRunTime <= endTime {
				return fmt.Sprintf("最大并发数：%d， 总运行时长%ds, 任务正常结束！", concurrent, endTime-targetTime)
			}

			// 如果当前并发数小于最大并发数，
			if concurrent < maxConcurrent {
				if endTime-startTime >= stepRunTime {
					// 从开始时间算起，加上持续时长。如果大于现在的时间，说明已经运行了持续时长的时间，那么就要将开始时间的值，变为现在的时间值
					concurrent = concurrent + step
					if concurrent > maxConcurrent {
						concurrent = maxConcurrent
					}
					if startTime+stepRunTime <= endTime && concurrent < maxConcurrent {
						startTime = endTime + stepRunTime
					}
				}
			}
			if concurrent == maxConcurrent {
				if target == 0 {
					target++
					stepRunTime = stableDuration
					startTime = endTime + stepRunTime
				}

			}

		}
		currentWg.Wait()
	}
	return fmt.Sprintf("最大并发数：%d， 总运行时长%ds, 任务非正常结束！", concurrent, time.Now().Unix()-targetTime)

}
