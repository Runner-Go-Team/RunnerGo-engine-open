package golink

import (
	"encoding/json"
	"fmt"
	"github.com/Runner-Go-Team/RunnerGo-engine-open/constant"
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
				varForm.IsChecked = constant.Open
				varForm.Key = v.Key
				varForm.Value = v.Value
				scene.Configuration.SceneVariable.Variable = append(scene.Configuration.SceneVariable.Variable, varForm)
			}
		}
	}

	var globalVar, preNodeMap = new(sync.Map), new(sync.Map)
	for _, par := range scene.Configuration.SceneVariable.Variable {
		if par.IsChecked != constant.Open {
			continue
		}
		globalVar.Store(par.Key, par.Value)
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
			case constant.PlanType:
				node.TeamId = reportMsg.TeamId
				node.PlanId = reportMsg.PlanId
				node.ReportId = reportMsg.ReportId
				node.Debug = scene.Debug
				go disposePlanNode(preNodeMap, tempScene, globalVar, node, wg, sceneWg, reportMsg, resultDataMsgCh, requestCollection, options...)
			case constant.SceneType:
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
				case constant.End:
					goroutineId = disOptions[0]
					if preEventStatus.Concurrent >= preMaxConcurrent {
						preMaxConcurrent = preEventStatus.Concurrent
					}

					if event.Type == constant.IfControllerType || event.Type == constant.WaitControllerType {
						if preEventStatus.Weight >= preMaxWeight {
							preMaxWeight = preEventStatus.Weight
						}
					}
				case constant.NotRun:
					eventResult.Status = constant.NotRun
					eventResult.Weight = event.Weight
					if event.NextList != nil && len(event.NextList) >= 1 {
						//for _, _ = range event.NextList {
						//	nodeCh <- eventResult
						//}
						preNodeMap.Store(event.Id, eventResult)
					}
					return
				case constant.NotHit:
					eventResult.Status = constant.NotRun
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

		if event.Type == constant.WaitControllerType || event.Type == constant.IfControllerType {
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
		if event.Type == constant.WaitControllerType || event.Type == constant.IfControllerType {
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
		eventResult.Status = constant.NotRun
		eventResult.Weight = event.Weight
		if event.NextList != nil && len(event.NextList) >= 1 {
			//for _, _ = range event.NextList {
			//	nodeCh <- eventResult
			//}
			preNodeMap.Store(event.Id, eventResult)
		}
		return
	}

	if goroutineId > eventResult.Concurrent && event.Type == constant.RequestType {
		eventResult.Status = constant.NotRun
		eventResult.Weight = event.Weight
		if event.NextList != nil && len(event.NextList) >= 1 {
			preNodeMap.Store(event.Id, eventResult)
		}
		return
	}

	event.TeamId = scene.TeamId
	event.Debug = scene.Debug
	event.ReportId = scene.ReportId
	if scene.Configuration.SceneVariable != nil {
		event.Api.ApiVariable = new(model.GlobalVariable)
		scene.Configuration.SceneVariable.SupToSub(event.Api.ApiVariable)
		event.Api.ApiVariable.InitReplace()
	}
	switch event.Type {
	case constant.RequestType:
		event.Api.Uuid = scene.Uuid
		var requestResults = &model.ResultDataMsg{}
		DisposeRequest(reportMsg, resultDataMsgCh, requestResults, globalVar, event, requestCollection, goroutineId, eventResult.Concurrent)
		eventResult.Status = constant.End
		eventResult.Weight = event.Weight
		if event.NextList != nil && len(event.NextList) >= 1 {
			//for _, _ = range event.NextList {
			//	nodeCh <- eventResult
			//}
			preNodeMap.Store(event.Id, eventResult)
		}

	case constant.IfControllerType:
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
		var result = constant.Failed
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
		if result == constant.Failed {
			eventResult.Status = constant.NotHit
			eventResult.Weight = event.Weight
		} else {
			eventResult.Status = constant.End
			eventResult.Weight = event.Weight
		}
		preNodeMap.Store(event.Id, eventResult)
	case constant.WaitControllerType:
		time.Sleep(time.Duration(event.WaitTime) * time.Millisecond)
		eventResult.Status = constant.End
		eventResult.Weight = event.Weight
		if event.NextList != nil && len(event.NextList) >= 1 {
			preNodeMap.Store(event.Id, eventResult)
		}
	case constant.SqlType:

	case constant.MqttType:
	case constant.KafkaType:
	case constant.RedisType:

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
				case constant.NotRun:
					eventResult.Status = constant.NotRun
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
					debugMsg["status"] = constant.NotRun
					debugMsg["msg"] = "未运行"
					debugMsg["type"] = event.Type
					switch event.Type {
					case constant.RequestType:
						debugMsg["api_name"] = event.Api.Name
						debugMsg["api_id"] = event.Api.TargetId
					case constant.IfControllerType:
						debugMsg["api_name"] = constant.IfControllerType
					case constant.WaitControllerType:
						debugMsg["api_name"] = constant.IfControllerType
					}
					debugMsg["next_list"] = event.NextList
					if requestCollection != nil {
						model.Insert(requestCollection, debugMsg, middlewares.LocalIp)
					}
					return
				case constant.NotHit:
					eventResult.Status = constant.NotRun
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
					debugMsg["status"] = constant.NotRun
					debugMsg["msg"] = "未运行"
					debugMsg["type"] = event.Type
					switch event.Type {
					case constant.RequestType:
						debugMsg["api_name"] = event.Api.Name
						debugMsg["api_id"] = event.Api.TargetId
					case constant.IfControllerType:
						debugMsg["api_name"] = constant.IfControllerType
					case constant.WaitControllerType:
						debugMsg["api_name"] = constant.IfControllerType
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

	if scene.Configuration != nil || scene.Configuration.SceneVariable != nil {
		if event.Api.ApiVariable == nil {
			event.Api.ApiVariable = new(model.GlobalVariable)
		}
		event.Api.ApiVariable.InitReplace()
		scene.Configuration.SceneVariable.SupToSub(event.Api.ApiVariable)
	}

	switch event.Type {
	case constant.RequestType:
		event.Api.Uuid = scene.Uuid
		event.CaseId = scene.CaseId
		DisposeRequest(reportMsg, resultDataMsgCh, nil, globalVar, event, requestCollection)
		eventResult.Status = constant.End
		preNodeMap.Store(event.Id, eventResult)
	case constant.IfControllerType:
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
		var result = constant.Failed
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

		setControllerDebugMsg(preNodeMap, eventResult, scene, event, requestCollection, msg, result, constant.IfControllerType)

	case constant.WaitControllerType:
		time.Sleep(time.Duration(event.WaitTime) * time.Millisecond)
		msg := fmt.Sprintf("等待了 %d 毫秒", event.WaitTime)
		setControllerDebugMsg(preNodeMap, eventResult, scene, event, requestCollection, msg, constant.Success, constant.WaitControllerType)
	}
}

// DisposeRequest 开始对请求进行处理
func DisposeRequest(reportMsg *model.ResultDataMsg, resultDataMsgCh chan *model.ResultDataMsg, requestResults *model.ResultDataMsg, globalVar *sync.Map,
	event model.Event, mongoCollection *mongo.Collection, options ...int64) {
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

	var (
		isSucceed          = false
		errCode            = int64(0)
		requestTime        = uint64(0)
		sendBytes          = float64(0)
		receivedBytes      = float64(0)
		errMsg             = ""
		startTime, endTime = time.Time{}, time.Time{}
	)
	if event.Prepositions != nil && len(event.Prepositions) > 0 {
		for _, preposition := range event.Prepositions {
			preposition.Exec()
		}
	}

	var debugMsg = make(map[string]interface{})
	if api.Debug != "" && api.Debug != "stop" {
		debugMsg["team_id"] = event.TeamId
		debugMsg["plan_id"] = event.PlanId
		debugMsg["report_id"] = event.ReportId
		debugMsg["scene_id"] = event.SceneId
		debugMsg["parent_id"] = event.ParentId
		debugMsg["case_id"] = event.CaseId
		debugMsg["uuid"] = api.Uuid.String()
		debugMsg["event_id"] = event.Id
		debugMsg["api_id"] = api.TargetId
		debugMsg["api_name"] = api.Name
		debugMsg["next_list"] = event.NextList
	}

	switch api.TargetType {
	case constant.FormTypeHTTP:
		api.Request.PreUrl = strings.TrimSpace(api.Request.PreUrl)
		api.Request.URL = api.Request.PreUrl + api.Request.URL

		if api.ApiVariable != nil {
			api.GlobalToRequest()
		}
		// 请求中所有的变量替换成真正的值
		api.ReplaceQueryParameterizes(globalVar)
		isSucceed, errCode, requestTime, sendBytes, receivedBytes, errMsg, startTime, endTime = api.Request.Send(api.Debug, debugMsg, mongoCollection, globalVar)
	case constant.FormTypeWebSocket:
		isSucceed, errCode, requestTime, sendBytes, receivedBytes = api.Ws.Send(api.Debug, debugMsg, mongoCollection, globalVar)
	case constant.FormTypeDubbo:
		//isSucceed, errCode, requestTime, sendBytes, contentLength := rpcSend(request)
	case constant.FormTypeTcp:

	case constant.FormTypeSql:
		isSucceed, requestTime, startTime, endTime = api.SQL.Send(event.Api.Debug, debugMsg, mongoCollection, globalVar)
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

// DisposeSql 开始对请求进行处理
func DisposeSql(reportMsg *model.ResultDataMsg, resultDataMsgCh chan *model.ResultDataMsg, requestResults *model.ResultDataMsg, globalVar *sync.Map,
	event model.Event, mongoCollection *mongo.Collection, options ...int64) {
	sql := event.Api.SQL

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
		requestResults.TargetId = event.Api.TargetId
		requestResults.Name = event.Api.Name
		requestResults.MachineNum = reportMsg.MachineNum
	}

	var (
		isSucceed          = false
		errCode            = int64(0)
		requestTime        = uint64(0)
		sendBytes          = float64(0)
		receivedBytes      = float64(0)
		errMsg             = ""
		startTime, endTime = time.Time{}, time.Time{}
	)
	if event.Prepositions != nil && len(event.Prepositions) > 0 {
		for _, preposition := range event.Prepositions {
			preposition.Exec()
		}
	}
	debugMsg := make(map[string]interface{})
	debugMsg["team_id"] = event.Api.TeamId
	debugMsg["sql_name"] = event.Api.Name
	debugMsg["target_id"] = event.Api.TargetId
	debugMsg["uuid"] = event.Api.Uuid.String()
	//isSucceed, requestTime, startTime, endTime = SqlSend(sql, sqlInfo, mongoCollection, globalVar)
	isSucceed, requestTime, startTime, endTime = sql.Send(event.Api.Debug, debugMsg, mongoCollection, globalVar)
	if resultDataMsgCh != nil {
		requestResults.Name = event.Api.Name
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

// DisposeTcp 开始对请求进行处理
func DisposeTcp(reportMsg *model.ResultDataMsg, resultDataMsgCh chan *model.ResultDataMsg, requestResults *model.ResultDataMsg, globalVar *sync.Map,
	event model.Event, mongoCollection *mongo.Collection, options ...int64) {
	tcp := event.Api.TCP
	tcp.TeamId = event.TeamId

	tcp.Debug = event.Debug

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
		requestResults.TargetId = tcp.TargetId
		requestResults.Name = tcp.Name
		requestResults.MachineNum = reportMsg.MachineNum
	}

	var (
		isSucceed          = false
		errCode            = int64(0)
		requestTime        = uint64(0)
		sendBytes          = float64(0)
		receivedBytes      = float64(0)
		errMsg             = ""
		startTime, endTime = time.Time{}, time.Time{}
	)
	if event.Prepositions != nil && len(event.Prepositions) > 0 {
		for _, preposition := range event.Prepositions {
			preposition.Exec()
		}
	}

	TcpConnection(tcp, mongoCollection)
	//isSucceed, requestTime, startTime, endTime = TcpConnection(tcp)

	if resultDataMsgCh != nil {
		requestResults.Name = tcp.Name
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
	if status == constant.Failed {
		eventResult.Status = constant.NotHit
	} else {
		eventResult.Status = constant.End
	}
	preNodeMap.Store(event.Id, eventResult)
	//for _, _ = range event.NextList {
	//	nodeCh <- eventResult
	//}
}

// DisposeWs 开始对请求进行处理
func DisposeWs(reportMsg *model.ResultDataMsg, resultDataMsgCh chan *model.ResultDataMsg, requestResults *model.ResultDataMsg, globalVar *sync.Map,
	event model.Event, mongoCollection *mongo.Collection, options ...int64) {
	ws := event.Api.Ws
	ws.TeamId = event.TeamId

	ws.Debug = event.Debug

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
		requestResults.TargetId = ws.TargetId
		requestResults.Name = ws.Name
		requestResults.MachineNum = reportMsg.MachineNum
	}

	var (
		isSucceed          = false
		errCode            = int64(0)
		requestTime        = uint64(0)
		sendBytes          = float64(0)
		receivedBytes      = float64(0)
		errMsg             = ""
		startTime, endTime = time.Time{}, time.Time{}
	)
	if event.Prepositions != nil && len(event.Prepositions) > 0 {
		for _, preposition := range event.Prepositions {
			preposition.Exec()
		}
	}

	webSocketSend(ws, mongoCollection)
	//isSucceed, requestTime, startTime, endTime = TcpConnection(tcp)

	if resultDataMsgCh != nil {
		requestResults.Name = ws.Name
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

// DisposeDubbo 开始对请求进行处理
func DisposeDubbo(reportMsg *model.ResultDataMsg, resultDataMsgCh chan *model.ResultDataMsg, requestResults *model.ResultDataMsg, globalVar *sync.Map,
	event model.Event, mongoCollection *mongo.Collection, options ...int64) {
	dubbo := event.Api.DubboDetail
	dubbo.TeamId = event.TeamId
	dubbo.Debug = event.Debug
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
		requestResults.TargetId = dubbo.TargetId
		requestResults.Name = dubbo.Name
		requestResults.MachineNum = reportMsg.MachineNum
	}

	var (
		isSucceed          = false
		errCode            = int64(0)
		requestTime        = uint64(0)
		sendBytes          = float64(0)
		receivedBytes      = float64(0)
		errMsg             = ""
		startTime, endTime = time.Time{}, time.Time{}
	)
	if event.Prepositions != nil && len(event.Prepositions) > 0 {
		for _, preposition := range event.Prepositions {
			preposition.Exec()
		}
	}

	SendDubbo(dubbo, mongoCollection)
	//isSucceed, requestTime, startTime, endTime = TcpConnection(tcp)

	if resultDataMsgCh != nil {
		requestResults.Name = dubbo.Name
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

// DisposeMqtt 开始对请求进行处理
//func DisposeMqtt(reportMsg *model.ResultDataMsg, resultDataMsgCh chan *model.ResultDataMsg, requestResults *model.ResultDataMsg, globalVar *sync.Map,
//	event model.Event, mongoCollection *mongo.Collection, options ...int64) {
//	mqtt := event.MQTT
//	mqtt.TeamId = event.TeamId
//
//	mqtt.Debug = event.Debug
//
//	if requestResults != nil {
//		requestResults.PlanId = reportMsg.PlanId
//		requestResults.PlanName = reportMsg.PlanName
//		requestResults.EventId = event.Id
//		requestResults.PercentAge = event.PercentAge
//		requestResults.ResponseThreshold = event.ResponseThreshold
//		requestResults.TeamId = event.TeamId
//		requestResults.SceneId = reportMsg.SceneId
//		requestResults.MachineIp = reportMsg.MachineIp
//		requestResults.Concurrency = options[1]
//		requestResults.SceneName = reportMsg.SceneName
//		requestResults.ReportId = reportMsg.ReportId
//		requestResults.ReportName = reportMsg.ReportName
//		requestResults.PercentAge = event.PercentAge
//		requestResults.RequestThreshold = event.RequestThreshold
//		requestResults.ResponseThreshold = event.ResponseThreshold
//		requestResults.ErrorThreshold = event.ErrorThreshold
//		requestResults.TargetId = mqtt.TargetId
//		requestResults.Name = mqtt.Name
//		requestResults.MachineNum = reportMsg.MachineNum
//	}
//
//	var (
//		isSucceed          = false
//		errCode            = int64(0)
//		requestTime        = uint64(0)
//		sendBytes          = float64(0)
//		receivedBytes      = float64(0)
//		errMsg             = ""
//		startTime, endTime = time.Time{}, time.Time{}
//	)
//	if event.Prepositions != nil && len(event.Prepositions) > 0 {
//		for _, preposition := range event.Prepositions {
//			preposition.Exec(globalVar, mqtt.GlobalVariable)
//		}
//	}
//
//	SendMqtt(mqtt, mongoCollection)
//	//isSucceed, requestTime, startTime, endTime = TcpConnection(tcp)
//
//	if resultDataMsgCh != nil {
//		requestResults.Name = mqtt.Name
//		requestResults.RequestTime = requestTime
//		requestResults.ErrorType = errCode
//		requestResults.IsSucceed = isSucceed
//		requestResults.SendBytes = sendBytes
//		requestResults.ReceivedBytes = receivedBytes
//		requestResults.ErrorMsg = errMsg
//		requestResults.Timestamp = time.Now().UnixMilli()
//		requestResults.StartTime = startTime.UnixMilli()
//		requestResults.EndTime = endTime.UnixMilli()
//		resultDataMsgCh <- requestResults
//	}
//
//}
