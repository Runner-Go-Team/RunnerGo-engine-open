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
	"math"
	"sync"
	"time"
)

// RTModel 响应时间模式
func RTModel(scene model.Scene, configuration *model.Configuration, reportMsg *model.ResultDataMsg, resultDataMsgCh chan *model.ResultDataMsg, requestCollection *mongo.Collection) (msg string) {
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
	currentWg := &sync.WaitGroup{}
	concurrentMap := new(sync.Map)
	// 只要开始时间+持续时长大于当前时间就继续循环
	targetTime, startTime, endTime := time.Now().Unix(), time.Now().Unix(), time.Now().Unix()
	switch scene.ConfigTask.ControlMode {
	case constant.CentralizedMode:
		for startTime+stepRunTime > endTime {
			scene.Debug = debug

			res := model.QueryReportData(key)
			if res != "" {
				var result = new(model.RedisSceneTestResultDataMsg)
				err := json.Unmarshal([]byte(res), result)
				if err != nil {
					break
				}
				for _, resultData := range result.Results {
					switch resultData.PercentAge {
					case 50:
						times := int64(math.Ceil(resultData.FiftyRequestTimelineValue / float64(time.Millisecond)))
						if resultData.ResponseThreshold > 0 && times >= resultData.ResponseThreshold {
							return fmt.Sprintf("最大并发数：%d， 总运行时长%ds,  接口：%s: %d 响应时间线大于等于阈值%d, 任务结束！", concurrent, endTime-targetTime, resultData.Name, resultData.PercentAge, resultData.RequestThreshold)
						}
					case 90:
						times := int64(math.Ceil(resultData.NinetyRequestTimeLineValue / float64(time.Millisecond)))
						if resultData.ResponseThreshold > 0 && times >= resultData.ResponseThreshold {
							return fmt.Sprintf("最大并发数：%d， 总运行时长%ds,  接口：%s: %d 响应时间线大于等于阈值%d, 任务结束！", concurrent, endTime-targetTime, resultData.Name, resultData.PercentAge, resultData.RequestThreshold)
						}
					case 95:
						times := int64(math.Ceil(resultData.NinetyFiveRequestTimeLineValue / float64(time.Millisecond)))
						if resultData.ResponseThreshold > 0 && times >= resultData.ResponseThreshold {
							return fmt.Sprintf("最大并发数：%d， 总运行时长%ds,  接口：%s: %d 响应时间线大于等于阈值%d, 任务结束！", concurrent, endTime-targetTime, resultData.Name, resultData.PercentAge, resultData.RequestThreshold)
						}
					case 99:
						times := int64(math.Ceil(resultData.NinetyNineRequestTimeLineValue / float64(time.Millisecond)))
						if resultData.ResponseThreshold > 0 && times >= resultData.ResponseThreshold {
							return fmt.Sprintf("最大并发数：%d， 总运行时长%ds,  接口：%s: %d 响应时间线大于等于阈值%d, 任务结束！", concurrent, endTime-targetTime, resultData.Name, resultData.PercentAge, resultData.RequestThreshold)
						}
					case 100:
						times := int64(math.Ceil(resultData.MaxRequestTime / float64(time.Millisecond)))
						if resultData.ResponseThreshold > 0 && times >= resultData.ResponseThreshold {
							return fmt.Sprintf("最大并发数：%d， 总运行时长%ds,  接口：%s: %d 响应时间线大于等于阈值%d, 任务结束！", concurrent, endTime-targetTime, resultData.Name, resultData.PercentAge, resultData.RequestThreshold)
						}
					case 101:
						times := int64(math.Ceil(resultData.AvgRequestTime / float64(time.Millisecond)))
						if resultData.ResponseThreshold > 0 && times >= resultData.ResponseThreshold {
							return fmt.Sprintf("最大并发数：%d， 总运行时长%ds,  接口：%s: %d 响应时间线大于等于阈值%d, 任务结束！", concurrent, endTime-targetTime, resultData.Name, resultData.PercentAge, resultData.RequestThreshold)
						}
					default:
						if resultData.PercentAge == resultData.CustomRequestTimeLine {
							times := int64(math.Ceil(resultData.CustomRequestTimeLineValue / float64(time.Millisecond)))
							if resultData.ResponseThreshold > 0 && times >= resultData.ResponseThreshold {
								return fmt.Sprintf("最大并发数：%d， 总运行时长%ds,  接口：%s: %d 响应时间线大于等于阈值%d, 任务结束！", concurrent, endTime-targetTime, resultData.Name, resultData.PercentAge, resultData.RequestThreshold)
							}
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
				for i := int64(0); i < concurrent; i++ {
					if _, isOk := concurrentMap.Load(i); isOk {
						continue
					}

					currentWg.Add(1)
					go func(concurrentId, concurrent int64, useConfiguration *model.Configuration) {
						defer currentWg.Done()
						defer concurrentMap.Delete(concurrentId)
						golink.DisposeScene(constant.PlanType, scene, useConfiguration, reportMsg, resultDataMsgCh, requestCollection, concurrentId, concurrent)
					}(i, concurrent, configuration)
				}
				currentWg.Wait()
			}

			endTime = time.Now().Unix()
			// 当此时的并发等于最大并发数时，并且持续时长等于稳定持续时长且当前运行时长大于等于此时时结束
			if concurrent == maxConcurrent && stepRunTime == stableDuration && startTime+stepRunTime <= endTime {
				return fmt.Sprintf("到达最大并发数：%d, 总运行时长%d秒, 任务正常结束！", maxConcurrent, endTime-targetTime)
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
			scene.Debug = debug

			res := model.QueryReportData(key)
			if res != "" {
				var result = new(model.RedisSceneTestResultDataMsg)
				err := json.Unmarshal([]byte(res), result)
				if err != nil {
					break
				}
				for _, resultData := range result.Results {
					switch resultData.PercentAge {
					case 50:
						times := int64(math.Ceil(resultData.FiftyRequestTimelineValue / float64(time.Millisecond)))
						if resultData.ResponseThreshold > 0 && times >= resultData.ResponseThreshold {
							return fmt.Sprintf("最大并发数：%d， 总运行时长%ds,  接口：%s: %d 响应时间线大于等于阈值%d, 任务结束！", concurrent, endTime-targetTime, resultData.Name, resultData.PercentAge, resultData.RequestThreshold)
						}
					case 90:
						times := int64(math.Ceil(resultData.NinetyRequestTimeLineValue / float64(time.Millisecond)))
						if resultData.ResponseThreshold > 0 && times >= resultData.ResponseThreshold {
							return fmt.Sprintf("最大并发数：%d， 总运行时长%ds,  接口：%s: %d 响应时间线大于等于阈值%d, 任务结束！", concurrent, endTime-targetTime, resultData.Name, resultData.PercentAge, resultData.RequestThreshold)
						}
					case 95:
						times := int64(math.Ceil(resultData.NinetyFiveRequestTimeLineValue / float64(time.Millisecond)))
						if resultData.ResponseThreshold > 0 && times >= resultData.ResponseThreshold {
							return fmt.Sprintf("最大并发数：%d， 总运行时长%ds,  接口：%s: %d 响应时间线大于等于阈值%d, 任务结束！", concurrent, endTime-targetTime, resultData.Name, resultData.PercentAge, resultData.RequestThreshold)
						}
					case 99:
						times := int64(math.Ceil(resultData.NinetyNineRequestTimeLineValue / float64(time.Millisecond)))
						if resultData.ResponseThreshold > 0 && times >= resultData.ResponseThreshold {
							return fmt.Sprintf("最大并发数：%d， 总运行时长%ds,  接口：%s: %d 响应时间线大于等于阈值%d, 任务结束！", concurrent, endTime-targetTime, resultData.Name, resultData.PercentAge, resultData.RequestThreshold)
						}
					case 100:
						times := int64(math.Ceil(resultData.MaxRequestTime / float64(time.Millisecond)))
						if resultData.ResponseThreshold > 0 && times >= resultData.ResponseThreshold {
							return fmt.Sprintf("最大并发数：%d， 总运行时长%ds,  接口：%s: %d 响应时间线大于等于阈值%d, 任务结束！", concurrent, endTime-targetTime, resultData.Name, resultData.PercentAge, resultData.RequestThreshold)
						}
					case 101:
						times := int64(math.Ceil(resultData.AvgRequestTime / float64(time.Millisecond)))
						if resultData.ResponseThreshold > 0 && times >= resultData.ResponseThreshold {
							return fmt.Sprintf("最大并发数：%d， 总运行时长%ds,  接口：%s: %d 响应时间线大于等于阈值%d, 任务结束！", concurrent, endTime-targetTime, resultData.Name, resultData.PercentAge, resultData.RequestThreshold)
						}
					default:
						if resultData.PercentAge == resultData.CustomRequestTimeLine {
							times := int64(math.Ceil(resultData.CustomRequestTimeLineValue / float64(time.Millisecond)))
							if resultData.ResponseThreshold > 0 && times >= resultData.ResponseThreshold {
								return fmt.Sprintf("最大并发数：%d， 总运行时长%ds,  接口：%s: %d 响应时间线大于等于阈值%d, 任务结束！", concurrent, endTime-targetTime, resultData.Name, resultData.PercentAge, resultData.RequestThreshold)
							}
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
				for i := int64(0); i < concurrent; i++ {
					if _, isOk := concurrentMap.Load(i); isOk {
						continue
					}
					concurrentMap.Store(i, true)
					currentWg.Add(1)
					go func(concurrentId int64, useConfiguration *model.Configuration) {
						defer currentWg.Done()
						defer concurrentMap.Delete(concurrentId)
						for startTime+stepRunTime > endTime {
							if _, ok := concurrentMap.Load(i); !ok {
								break
							}
							golink.DisposeScene(constant.PlanType, scene, useConfiguration, reportMsg, resultDataMsgCh, requestCollection, concurrentId, concurrent)
						}

					}(i, configuration)
				}

			}

			endTime = time.Now().Unix()
			// 当此时的并发等于最大并发数时，并且持续时长等于稳定持续时长且当前运行时长大于等于此时时结束
			if concurrent == maxConcurrent && stepRunTime == stableDuration && startTime+stepRunTime <= endTime {
				return fmt.Sprintf("到达最大并发数：%d, 总运行时长%d秒, 任务正常结束！", maxConcurrent, endTime-targetTime)
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

	return fmt.Sprintf("到达最大并发数：%d, 总运行时长%d秒, 任务非正常结束！", maxConcurrent, endTime-targetTime)

}
