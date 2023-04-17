package golink

import (
	"encoding/json"
	"fmt"
	"github.com/Runner-Go-Team/RunnerGo-engine-open/middlewares"
	"github.com/Runner-Go-Team/RunnerGo-engine-open/model"
	"github.com/Runner-Go-Team/RunnerGo-engine-open/tools"
	"go.mongodb.org/mongo-driver/mongo"
	"math"
	"strings"
	"sync"
	"time"
)

// DisposeScene 对场景进行处理
func DisposeScene(wg, sceneWg *sync.WaitGroup, runType string, scene model.Scene, configuration *model.Configuration, reportMsg *model.ResultDataMsg, resultDataMsgCh chan *model.ResultDataMsg, requestCollection *mongo.Collection, options ...int64) {
	sceneBy, _ := json.Marshal(scene)
	var tempScene model.Scene
	json.Unmarshal(sceneBy, &tempScene)
	nodesList := tempScene.NodesRound
	if configuration.ParameterizedFile.VariableNames != nil && configuration.ParameterizedFile.VariableNames.VarMapList != nil {
		configuration.Mu.Lock()
		kvList := configuration.VarToSceneKV()
		configuration.Mu.Unlock()
		if kvList != nil {
			for _, v := range kvList {
				varForm := new(model.VarForm)
				varForm.IsChecked = model.Open
				varForm.Key = v.Key
				varForm.Value = v.Value
				scene.Configuration.SceneVariable.Variable = append(scene.Configuration.SceneVariable.Variable, varForm)
			}
		}
	}

	var globalVar, preNodeMap = new(sync.Map), new(sync.Map)
	for _, par := range scene.Configuration.SceneVariable.Variable {
		globalVar.Store(par.Key, par.Value)
	}

	for _, v := range configuration.SceneVariable.Variable {
		if _, ok := globalVar.Load(v.Key); ok {
			continue
		}
		globalVar.Store(v.Key, v.Value)
	}
	globalVar.Range(func(key, value any) bool {
		if value == nil {
			return true
		}
		values := tools.FindAllDestStr(value.(string), "{{(.*?)}}")
		if values == nil {
			return true
		}

		for _, v := range values {
			if len(v) < 2 {
				return true
			}
			realVar := tools.ParsFunc(v[1])
			if realVar != v[1] {
				value = strings.Replace(value.(string), v[0], realVar, -1)
				globalVar.Store(key, value)
			}
		}
		return true
	})
	for _, nodes := range nodesList {
		for _, node := range nodes {
			node.Uuid = scene.Uuid
			wg.Add(1)
			sceneWg.Add(1)
			switch runType {
			case model.PlanType:
				node.TeamId = reportMsg.TeamId
				node.PlanId = reportMsg.PlanId
				node.ReportId = reportMsg.ReportId
				node.Debug = scene.Debug
				go disposePlanNode(preNodeMap, tempScene, globalVar, node, wg, sceneWg, reportMsg, resultDataMsgCh, requestCollection, options...)
			case model.SceneType:
				node.TeamId = scene.TeamId
				node.PlanId = scene.PlanId
				node.CaseId = scene.CaseId
				node.SceneId = scene.SceneId
				node.ReportId = scene.ReportId
				node.ParentId = scene.ParentId
				go disposeDebugNode(preNodeMap, tempScene, globalVar, node, wg, sceneWg, reportMsg, resultDataMsgCh, requestCollection)
			default:
				wg.Done()
				sceneWg.Done()
			}
		}
		sceneWg.Wait()
	}

}

// disposePlanNode 处理node节点
func disposePlanNode(preNodeMap *sync.Map, scene model.Scene, globalVar *sync.Map, event model.Event, wg, sceneWg *sync.WaitGroup, reportMsg *model.ResultDataMsg, resultDataMsgCh chan *model.ResultDataMsg, requestCollection *mongo.Collection, disOptions ...int64) {
	defer wg.Done()
	defer sceneWg.Done()

	var (
		goroutineId int64 // 启动的第几个协程
	)
	var eventResult = model.EventResult{}

	// 如果该事件上一级有事件，那么就一直查询上一级事件的状态，直到上一级所有事件全部完成
	if event.PreList != nil && len(event.PreList) > 0 {
		var preMaxConcurrent = int64(0) // 上一级最大并发数
		var preMaxWeight = int64(0)
		// 将上一级事件放入一个map中进行维护
		for _, eventId := range event.PreList {
			if preCh, ok := preNodeMap.Load(eventId); ok {
				if preCh == nil {
					continue
				}
				preEventStatus := preCh.(model.EventResult)
				switch preEventStatus.Status {
				case model.End:
					goroutineId = disOptions[0]
					if preEventStatus.Concurrent >= preMaxConcurrent {
						preMaxConcurrent = preEventStatus.Concurrent
					}

					if event.Type == model.IfControllerType || event.Type == model.WaitControllerType {
						if preEventStatus.Weight >= preMaxWeight {
							preMaxWeight = preEventStatus.Weight
						}
					}
				case model.NotRun:
					eventResult.Status = model.NotRun
					eventResult.Weight = event.Weight
					if event.NextList != nil && len(event.NextList) >= 1 {
						//for _, _ = range event.NextList {
						//	nodeCh <- eventResult
						//}
						preNodeMap.Store(event.Id, eventResult)
					}
					return
				case model.NotHit:
					eventResult.Status = model.NotRun
					eventResult.Weight = event.Weight
					if event.NextList != nil && len(event.NextList) >= 1 {
						//for _, _ = range event.NextList {
						//	nodeCh <- eventResult
						//}
						preNodeMap.Store(event.Id, eventResult)
					}
					return
				}
			}

		}

		if event.Type == model.WaitControllerType || event.Type == model.IfControllerType {
			event.Weight = preMaxWeight
		}
		if event.Weight > 0 && event.Weight < 100 {
			eventResult.Concurrent = int64(math.Ceil(float64(event.Weight) * float64(preMaxConcurrent) / 100))
		}

		if event.Weight == 100 {
			eventResult.Concurrent = preMaxConcurrent
		}

		// 如果该事件上一级有事件, 并且上一级事件中的第一个事件的权重不等于100，那么并发数就等于上一级的并发*权重

	} else {
		if event.Type == model.WaitControllerType || event.Type == model.IfControllerType {
			event.Weight = 100
		}
		if disOptions != nil && len(disOptions) > 1 {
			if event.Weight == 100 {
				eventResult.Concurrent = disOptions[1]
			}
			if event.Weight > 0 && event.Weight < 100 {
				eventResult.Concurrent = int64(math.Ceil(float64(disOptions[1]) * (float64(event.Weight) / float64(100))))
			}
		}

	}

	if eventResult.Concurrent == 0 {
		eventResult.Status = model.NotRun
		eventResult.Weight = event.Weight
		if event.NextList != nil && len(event.NextList) >= 1 {
			//for _, _ = range event.NextList {
			//	nodeCh <- eventResult
			//}
			preNodeMap.Store(event.Id, eventResult)
		}
		return
	}

	if goroutineId > eventResult.Concurrent && event.Type == model.RequestType {
		eventResult.Status = model.NotRun
		eventResult.Weight = event.Weight
		if event.NextList != nil && len(event.NextList) >= 1 {
			//for _, _ = range event.NextList {
			//	nodeCh <- eventResult
			//}
			preNodeMap.Store(event.Id, eventResult)
		}
		return
	}

	event.TeamId = scene.TeamId
	event.Debug = scene.Debug
	event.ReportId = scene.ReportId
	if scene.GlobalVariable != nil {
		scene.GlobalVariable.GlobalToRequest(event.Api)
	}

	switch event.Type {
	case model.RequestType:
		event.Api.Uuid = scene.Uuid
		var requestResults = &model.ResultDataMsg{}
		DisposeRequest(reportMsg, resultDataMsgCh, requestResults, globalVar, event, requestCollection, goroutineId, eventResult.Concurrent)
		eventResult.Status = model.End
		eventResult.Weight = event.Weight
		if event.NextList != nil && len(event.NextList) >= 1 {
			//for _, _ = range event.NextList {
			//	nodeCh <- eventResult
			//}
			preNodeMap.Store(event.Id, eventResult)
		}

	case model.IfControllerType:
		keys := tools.FindAllDestStr(event.Var, "{{(.*?)}}")
		if len(keys) > 0 {
			for _, val := range keys {
				if value, ok := globalVar.Load(val[1]); ok {
					if value != nil {
						str := ""
						switch fmt.Sprintf("%T", value) {
						case "string":
							str = value.(string)
						case "float64":
							str = fmt.Sprintf("%f", value)
							if strings.HasSuffix(str, ".000000") {
								str = strings.Split(str, ".")[0]
							}
						case "bool":
							str = fmt.Sprintf("%t", value)
						case "int":
							str = fmt.Sprintf("%d", value)
						}
						event.Var = strings.Replace(event.Var, val[0], str, -1)
					}

				}
			}
		}
		values := tools.FindAllDestStr(event.Val, "{{(.*?)}}")
		if len(values) > 0 {
			for _, val := range values {
				if value, ok := globalVar.Load(val[1]); ok {
					if value == nil {
						continue
					}

					str := ""
					switch fmt.Sprintf("%T", value) {
					case "string":
						str = value.(string)
					case "float64":
						str = fmt.Sprintf("%f", value)
						if strings.HasSuffix(str, ".000000") {
							str = strings.Split(str, ".")[0]
						}
					case "bool":
						str = fmt.Sprintf("%t", value)
					case "int":
						str = fmt.Sprintf("%d", value)
					}
					event.Val = strings.Replace(event.Val, val[0], str, -1)

				}
			}
		}
		var result = model.Failed
		var temp = false
		globalVar.Range(func(key, value any) bool {
			if key == event.Var {
				temp = true
				if value != nil {
					str := ""
					switch fmt.Sprintf("%T", value) {
					case "string":
						str = value.(string)
					case "float64":
						str = fmt.Sprintf("%f", value)
					case "bool":
						str = fmt.Sprintf("%t", value)
					case "int":
						str = fmt.Sprintf("%d", value)
					}
					result, _ = event.PerForm(str)
				}
			}
			return true
		})
		if temp == false {
			result, _ = event.PerForm(event.Var)
		}
		if result == model.Failed {
			eventResult.Status = model.NotHit
			eventResult.Weight = event.Weight
		} else {
			eventResult.Status = model.End
			eventResult.Weight = event.Weight
		}
		//for _, _ = range event.NextList {
		//	nodeCh <- eventResult
		//}
		preNodeMap.Store(event.Id, eventResult)
	case model.WaitControllerType:
		time.Sleep(time.Duration(event.WaitTime) * time.Millisecond)
		eventResult.Status = model.End
		eventResult.Weight = event.Weight
		if event.NextList != nil && len(event.NextList) >= 1 {
			//for _, _ = range event.NextList {
			//	nodeCh <- eventResult
			//}
			preNodeMap.Store(event.Id, eventResult)
		}
	}
}

func disposeDebugNode(preNodeMap *sync.Map, scene model.Scene, globalVar *sync.Map, event model.Event, wg, sceneWg *sync.WaitGroup, reportMsg *model.ResultDataMsg, resultDataMsgCh chan *model.ResultDataMsg, requestCollection *mongo.Collection) {
	defer wg.Done()
	defer sceneWg.Done()
	//defer close(nodeCh)
	var eventResult = model.EventResult{}
	// 如果该事件上一级有事件，那么就一直查询上一级事件的状态，直到上一级所有事件全部完成
	if event.PreList != nil && len(event.PreList) > 0 {
		// 将上一级事件放入一个map中进行维护
		for _, preEventId := range event.PreList {
			if preCh, ok := preNodeMap.Load(preEventId); ok {
				if preCh == nil {
					continue
				}
				preEventStatus := preCh.(model.EventResult)
				switch preEventStatus.Status {
				case model.NotRun:
					eventResult.Status = model.NotRun
					//for _, _ = range event.NextList {
					//	nodeCh <- eventResult
					//}
					preNodeMap.Store(event.Id, eventResult)
					debugMsg := make(map[string]interface{})
					debugMsg["team_id"] = event.TeamId
					debugMsg["plan_id"] = event.PlanId
					debugMsg["report_id"] = event.ReportId
					debugMsg["scene_id"] = event.SceneId
					debugMsg["parent_id"] = event.ParentId
					debugMsg["case_id"] = event.CaseId
					debugMsg["uuid"] = event.Uuid.String()
					debugMsg["event_id"] = event.Id
					debugMsg["status"] = model.NotRun
					debugMsg["msg"] = "未运行"
					debugMsg["type"] = event.Type
					switch event.Type {
					case model.RequestType:
						debugMsg["api_name"] = event.Api.Name
						debugMsg["api_id"] = event.Api.TargetId
					case model.IfControllerType:
						debugMsg["api_name"] = model.IfControllerType
					case model.WaitControllerType:
						debugMsg["api_name"] = model.IfControllerType
					}
					debugMsg["next_list"] = event.NextList
					if requestCollection != nil {
						model.Insert(requestCollection, debugMsg, middlewares.LocalIp)
					}
					return
				case model.NotHit:
					eventResult.Status = model.NotRun
					//for _, _ = range event.NextList {
					//	nodeCh <- eventResult
					//}
					preNodeMap.Store(event.Id, eventResult)
					debugMsg := make(map[string]interface{})
					debugMsg["team_id"] = event.TeamId
					debugMsg["plan_id"] = event.PlanId
					debugMsg["report_id"] = event.ReportId
					debugMsg["scene_id"] = event.SceneId
					debugMsg["parent_id"] = event.ParentId
					debugMsg["case_id"] = event.CaseId
					debugMsg["uuid"] = event.Uuid.String()
					debugMsg["event_id"] = event.Id
					debugMsg["status"] = model.NotRun
					debugMsg["msg"] = "未运行"
					debugMsg["type"] = event.Type
					switch event.Type {
					case model.RequestType:
						debugMsg["api_name"] = event.Api.Name
						debugMsg["api_id"] = event.Api.TargetId
					case model.IfControllerType:
						debugMsg["api_name"] = model.IfControllerType
					case model.WaitControllerType:
						debugMsg["api_name"] = model.IfControllerType
					}
					debugMsg["next_list"] = event.NextList
					if requestCollection != nil {
						model.Insert(requestCollection, debugMsg, middlewares.LocalIp)
					}

					return
				}
			}
		}
	}
	event.TeamId = scene.TeamId
	event.Debug = scene.Debug
	event.ReportId = scene.ReportId
	if scene.GlobalVariable != nil {
		scene.GlobalVariable.GlobalToRequest(event.Api)
	}
	switch event.Type {
	case model.RequestType:
		event.Api.Uuid = scene.Uuid
		event.CaseId = scene.CaseId
		DisposeRequest(reportMsg, resultDataMsgCh, nil, globalVar, event, requestCollection)
		eventResult.Status = model.End

		preNodeMap.Store(event.Id, eventResult)
	case model.IfControllerType:
		keys := tools.FindAllDestStr(event.Var, "{{(.*?)}}")
		if len(keys) > 0 {
			for _, val := range keys {
				if value, ok := globalVar.Load(val[1]); ok {
					if value == nil {
						continue
					}
					str := ""
					switch fmt.Sprintf("%T", value) {
					case "string":
						str = value.(string)
					case "float64":
						str = fmt.Sprintf("%f", value)
						if strings.HasSuffix(str, ".000000") {
							str = strings.Split(str, ".")[0]
						}
					case "bool":
						str = fmt.Sprintf("%t", value)
					case "int":
						str = fmt.Sprintf("%d", value)
					}
					event.Var = strings.Replace(event.Var, val[0], str, -1)

				}
			}
		}
		values := tools.FindAllDestStr(event.Val, "{{(.*?)}}")
		if len(values) > 0 {
			for _, val := range values {
				if value, ok := globalVar.Load(val[1]); ok {
					if value == nil {
						continue
					}
					str := ""
					switch fmt.Sprintf("%T", value) {
					case "string":
						str = value.(string)
					case "float64":
						str = fmt.Sprintf("%f", value)
						if strings.HasSuffix(str, ".000000") {
							str = strings.Split(str, ".")[0]
						}
					case "bool":
						str = fmt.Sprintf("%t", value)
					case "int":
						str = fmt.Sprintf("%d", value)

					}
					event.Val = strings.Replace(event.Val, val[0], str, -1)

				}
			}
		}
		var result = model.Failed
		var msg = ""

		var temp = false
		globalVar.Range(func(key, value any) bool {
			if key == event.Var {
				temp = true
				if value != nil {
					str := ""
					switch fmt.Sprintf("%T", value) {
					case "string":
						str = value.(string)
					case "float64":
						str = fmt.Sprintf("%f", value)
					case "bool":
						str = fmt.Sprintf("%t", value)
					case "int":
						str = fmt.Sprintf("%d", value)

					}
					result, msg = event.PerForm(str)
				}
			}
			return true
		})
		if temp == false {
			result, msg = event.PerForm(event.Var)
		}

		setControllerDebugMsg(preNodeMap, eventResult, scene, event, requestCollection, msg, result, model.IfControllerType)

	case model.WaitControllerType:
		time.Sleep(time.Duration(event.WaitTime) * time.Millisecond)
		msg := fmt.Sprintf("等待了 %d 毫秒", event.WaitTime)
		setControllerDebugMsg(preNodeMap, eventResult, scene, event, requestCollection, msg, model.Success, model.WaitControllerType)
	}
}

// DisposeRequest 开始对请求进行处理
func DisposeRequest(reportMsg *model.ResultDataMsg, resultDataMsgCh chan *model.ResultDataMsg, requestResults *model.ResultDataMsg, globalVar *sync.Map,
	event model.Event, mongoCollection *mongo.Collection, options ...int64) {
	//api := &event.Api
	api := event.Api
	api.TeamId = event.TeamId

	api.Debug = event.Debug

	if requestResults != nil {
		requestResults.PlanId = reportMsg.PlanId
		requestResults.PlanName = reportMsg.PlanName
		requestResults.EventId = event.Id
		requestResults.PercentAge = event.PercentAge
		requestResults.ResponseThreshold = event.ResponseThreshold
		requestResults.TeamId = event.TeamId
		requestResults.SceneId = reportMsg.SceneId
		requestResults.MachineIp = reportMsg.MachineIp
		requestResults.Concurrency = options[1]
		requestResults.SceneName = reportMsg.SceneName
		requestResults.ReportId = reportMsg.ReportId
		requestResults.ReportName = reportMsg.ReportName
		requestResults.PercentAge = event.PercentAge
		requestResults.RequestThreshold = event.RequestThreshold
		requestResults.ResponseThreshold = event.ResponseThreshold
		requestResults.ErrorThreshold = event.ErrorThreshold
		requestResults.TargetId = api.TargetId
		requestResults.Name = api.Name
		requestResults.MachineNum = reportMsg.MachineNum
	}
	api.Request.PreUrl = strings.TrimSpace(api.Request.PreUrl)
	api.Request.URL = api.Request.PreUrl + api.Request.URL
	// 请求中所有的变量替换成真正的值
	api.ReplaceQueryParameterizes(globalVar)

	var (
		isSucceed          = false
		errCode            = int64(0)
		requestTime        = uint64(0)
		sendBytes          = float64(0)
		receivedBytes      = float64(0)
		errMsg             = ""
		startTime, endTime = time.Time{}, time.Time{}
	)
	switch api.TargetType {
	case model.FormTypeHTTP:
		isSucceed, errCode, requestTime, sendBytes, receivedBytes, errMsg, startTime, endTime = HttpSend(event, api, globalVar, mongoCollection)
	case model.FormTypeWebSocket:
		isSucceed, errCode, requestTime, sendBytes, receivedBytes = webSocketSend(api)
	case model.FormTypeGRPC:
		//isSucceed, errCode, requestTime, sendBytes, contentLength := rpcSend(request)
	default:
		return
	}

	if resultDataMsgCh != nil {
		requestResults.Name = api.Name
		requestResults.RequestTime = requestTime
		requestResults.ErrorType = errCode
		requestResults.IsSucceed = isSucceed
		requestResults.SendBytes = sendBytes
		requestResults.ReceivedBytes = receivedBytes
		requestResults.ErrorMsg = errMsg
		requestResults.Timestamp = time.Now().UnixMilli()
		requestResults.StartTime = startTime.UnixMilli()
		requestResults.EndTime = endTime.UnixMilli()
		resultDataMsgCh <- requestResults
	}

}

func setControllerDebugMsg(preNodeMap *sync.Map, eventResult model.EventResult, scene model.Scene, event model.Event, collection *mongo.Collection, msg, status, controllerType string) {
	if scene.Debug != "" {
		debugMsg := make(map[string]interface{})
		debugMsg["team_id"] = event.TeamId
		debugMsg["plan_id"] = event.PlanId
		debugMsg["report_id"] = event.ReportId
		debugMsg["api_name"] = controllerType
		debugMsg["scene_id"] = event.SceneId
		debugMsg["parent_id"] = event.ParentId
		debugMsg["case_id"] = event.CaseId
		debugMsg["uuid"] = event.Uuid.String()
		debugMsg["event_id"] = event.Id
		debugMsg["status"] = status
		debugMsg["type"] = controllerType
		debugMsg["msg"] = msg
		debugMsg["next_list"] = event.NextList
		if collection != nil {
			model.Insert(collection, debugMsg, middlewares.LocalIp)
		}
	}
	if status == model.Failed {
		eventResult.Status = model.NotHit
	} else {
		eventResult.Status = model.End
	}
	preNodeMap.Store(event.Id, eventResult)
	//for _, _ = range event.NextList {
	//	nodeCh <- eventResult
	//}
}
