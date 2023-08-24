// Package execution -----------------------------
// @file      : roundModel.go
// @author    : 被测试耽误的大厨
// @contact   : 13383088061@163.com
// @time      : 2023/6/26 16:12
// -------------------------------------------
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

// RoundModel 轮次模式
func RoundModel(scene model.Scene, configuration *model.Configuration, reportMsg *model.ResultDataMsg, resultDataMsgCh chan *model.ResultDataMsg, requestCollection *mongo.Collection) string {

	concurrent := scene.ConfigTask.ModeConf.Concurrency
	// 订阅redis中消息  任务状态：包括：报告停止；debug日志状态；任务配置变更
	adjustKey := fmt.Sprintf("SubscriptionStressPlanStatusChange:%s", reportMsg.ReportId)
	pubSub := model.SubscribeMsg(adjustKey)
	statusCh := pubSub.Channel()
	defer pubSub.Close()
	debug := scene.ConfigTask.DebugMode
	currentWg := &sync.WaitGroup{}
	// 定义一个map，管理并发
	concurrentMap := new(sync.Map)
	currentTime := time.Now().UnixMilli()
	// 按轮次压测
	if scene.ConfigTask.ModeConf.Concurrency == 0 || scene.ConfigTask.ModeConf.RoundNum == 0 {
		return fmt.Sprintf("轮次模式参数错误：无并发数或无运行轮次！无法运行！")
	}
	log.Logger.Info(fmt.Sprintf("机器ip:%s, 开始性能测试,轮次 %d轮", middlewares.LocalIp, scene.ConfigTask.ModeConf.RoundNum))
	rounds := scene.ConfigTask.ModeConf.RoundNum
	targetTime := time.Now().Unix()
	switch scene.ConfigTask.ControlMode {
	case constant.CentralizedMode:
		for i := int64(0); i < rounds; i++ {
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
						return fmt.Sprintf("并发数：%d， 运行了%d轮次, 任务手动结束！", concurrent, i-1)
					}
				case constant.DebugStatus:
					debug = subscriptionStressPlanStatusChange.Debug
				case constant.ReportChange:
					MachineModeConf := subscriptionStressPlanStatusChange.MachineModeConf
					if MachineModeConf.Machine != middlewares.LocalIp {
						continue
					}
					modeConf := MachineModeConf.ModeConf
					if modeConf.RoundNum > 0 {
						rounds = modeConf.RoundNum
					}
					if modeConf.Concurrency > 0 && modeConf.Concurrency != concurrent {
						// 如果修改后的并发小于当前并发
						if modeConf.Concurrency < concurrent {
							diff := concurrent - modeConf.Concurrency
							// 将最后的几个并发从map中去掉
							for j := int64(0); j < diff; j++ {
								concurrentMap.Delete(concurrent - 1 - j)
							}
						}
						concurrent = modeConf.Concurrency
					}
				}
			default:
				for j := int64(0); j < concurrent; j++ {
					if _, isOk := concurrentMap.Load(j); isOk {
						continue
					}
					concurrentMap.Store(j, true)
					currentWg.Add(1)
					scene.Debug = debug
					go func(concurrentId, concurrent int64, useConfiguration *model.Configuration, currentScene model.Scene) {
						defer currentWg.Done()
						defer concurrentMap.Delete(concurrentId)
						golink.DisposeScene(constant.PlanType, currentScene, useConfiguration, reportMsg, resultDataMsgCh, requestCollection, concurrentId, concurrent, currentTime)

					}(j, concurrent, configuration, scene)
				}
				currentWg.Wait()
			}
		}
		return fmt.Sprintf("并发数：%d， 运行了%d轮次, 任务正常结束！", concurrent, rounds)

	case constant.AloneMode:
		var stop bool
		for !stop {
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
						return fmt.Sprintf("并发数：%d， 运行了%ds, 任务手动结束！", concurrent, time.Now().Unix()-targetTime)
					}
				case constant.DebugStatus:
					debug = subscriptionStressPlanStatusChange.Debug
				case constant.ReportChange:
					MachineModeConf := subscriptionStressPlanStatusChange.MachineModeConf
					if MachineModeConf.Machine != middlewares.LocalIp {
						continue
					}
					modeConf := MachineModeConf.ModeConf
					if modeConf.RoundNum > 0 {
						rounds = modeConf.RoundNum
					}
					if modeConf.Concurrency > 0 && modeConf.Concurrency != concurrent {
						// 如果修改后的并发小于当前并发
						if modeConf.Concurrency < concurrent {
							diff := concurrent - modeConf.Concurrency
							// 将最后的几个并发从map中去掉
							for j := int64(0); j < diff; j++ {
								concurrentMap.Store(concurrent-1-j, false)
							}
						}
						concurrent = modeConf.Concurrency
					}
				}
			default:
				for i := int64(0); i < concurrent; i++ {
					if _, isOk := concurrentMap.Load(i); isOk {
						continue
					} else {
						concurrentMap.Store(i, true)
					}
					currentWg.Add(1)
					go func(concurrentId int64, useConfiguration *model.Configuration, currentScene model.Scene) {
						defer currentWg.Done()
						defer concurrentMap.Delete(concurrentId)
						for j := int64(0); j < rounds; j++ {
							if status, isOk := concurrentMap.Load(concurrentId); !isOk {
								break
							} else {
								if !status.(bool) {
									break
								}
							}
							currentScene.Debug = debug
							golink.DisposeScene(constant.PlanType, currentScene, useConfiguration, reportMsg, resultDataMsgCh, requestCollection, concurrentId, concurrent, currentTime)

						}
					}(i, configuration, scene)
				}
			}
			time.Sleep(50 * time.Millisecond)
			concurrentMapLen := 0
			concurrentMap.Range(func(key, value any) bool {
				if value.(bool) {
					concurrentMapLen++
				}
				return true
			})
			if concurrentMapLen == 0 {
				stop = true
			}
		}
		currentWg.Wait()

	}
	return fmt.Sprintf("并发数：%d, 总运行时长%ds, 任务正常结束!", concurrent, time.Now().Unix()-targetTime)

}
