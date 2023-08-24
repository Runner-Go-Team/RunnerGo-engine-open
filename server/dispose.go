// Package server 压测启动
package server

import (
	"context"
	"fmt"
	"github.com/Runner-Go-Team/RunnerGo-engine-open/config"
	"github.com/Runner-Go-Team/RunnerGo-engine-open/constant"
	"github.com/Runner-Go-Team/RunnerGo-engine-open/global"
	"github.com/Runner-Go-Team/RunnerGo-engine-open/log"
	"github.com/Runner-Go-Team/RunnerGo-engine-open/middlewares"
	"github.com/Runner-Go-Team/RunnerGo-engine-open/model"
	"github.com/Runner-Go-Team/RunnerGo-engine-open/server/execution"
	"github.com/Runner-Go-Team/RunnerGo-engine-open/server/golink"
	"github.com/Runner-Go-Team/RunnerGo-engine-open/tools"
	"github.com/Shopify/sarama"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	"net/http"
	"strings"
	"sync"
	"time"
)

// DisposeTask 处理任务
func DisposeTask(plan *model.Plan, c *gin.Context) {
	// 如果场景为空或者场景中的事件为空，直接结束该方法
	if plan.Scene.NodesRound == nil {
		global.ReturnMsg(c, http.StatusBadRequest, "执行计划失败：", "计划的场景不能为nil ")
		return
	}

	if plan.Scene.NodesRound[0] == nil {
		global.ReturnMsg(c, http.StatusBadRequest, "执行计划失败：", "计划的场景事件列表不能为空")
		return
	}
	if len(plan.Scene.NodesRound[0]) == 0 {
		global.ReturnMsg(c, http.StatusBadRequest, "执行计划失败：", "计划的场景事件不能为空")
		return
	}

	if plan.ReportId == "" {
		global.ReturnMsg(c, http.StatusBadRequest, "执行计划失败：", "reportId 不能为空 ")
		return
	}

	if plan.TeamId == "" {
		global.ReturnMsg(c, http.StatusBadRequest, "执行计划失败：", "teamId 不能为空 ")
		return
	}

	if plan.PlanId == "" {
		global.ReturnMsg(c, http.StatusBadRequest, "执行计划失败：", "planId 不能为空 ")
		return
	}

	if plan.PlanName == "" {
		global.ReturnMsg(c, http.StatusBadRequest, "执行计划失败：", "planName 不能为空 ")
		return
	}
	if plan.ConfigTask != nil {
		plan.Scene.ConfigTask = plan.ConfigTask
	} else {
		global.ReturnMsg(c, http.StatusBadRequest, "执行计划失败：", "计划任务不能为空")
		return
	}

	// 设置kafka消费者，目的是像kafka中发送测试结果
	kafkaProducer, err := model.NewKafkaProducer([]string{config.Conf.Kafka.Address})
	if err != nil {
		global.ReturnMsg(c, http.StatusBadRequest, "执行计划失败：", fmt.Sprintf("机器ip:%s, kafka连接失败: %s", middlewares.LocalIp, err.Error()))
		return
	}

	// 新建mongo客户端连接，用于发送debug数据
	mongoClient, err := model.NewMongoClient(
		config.Conf.Mongo.DSN,
		middlewares.LocalIp)
	if err != nil {
		global.ReturnMsg(c, http.StatusBadRequest, "执行计划失败：", fmt.Sprintf("机器ip:%s, 连接mongo错误：%s", middlewares.LocalIp, err.Error()))
		return
	}

	go func() {
		ExecutionPlan(plan, kafkaProducer, mongoClient)
	}()

	global.ReturnMsg(c, http.StatusOK, "开始执行计划", nil)
}

// ExecutionPlan 执行计划
func ExecutionPlan(plan *model.Plan, kafkaProducer sarama.SyncProducer, mongoClient *mongo.Client) {

	// 设置接收数据缓存
	resultDataMsgCh := make(chan *model.ResultDataMsg, 500000)

	// 向kafka发送消息

	topic := config.Conf.Kafka.TopIc
	partition := plan.Partition
	go model.SendKafkaMsg(kafkaProducer, resultDataMsgCh, topic, partition, middlewares.LocalIp)

	requestCollection := model.NewCollection(config.Conf.Mongo.DataBase, config.Conf.Mongo.StressDebugTable, mongoClient)
	//debugCollection := model.NewCollection(config.Conf.Mongo.DataBase, config.Conf.Mongo.DebugStatusTable, mongoClient)
	scene := plan.Scene

	// 如果场景中的任务配置勾选了全局任务配置，那么使用全局任务配置
	if scene.EnablePlanConfiguration == true {
		scene.ConfigTask = plan.ConfigTask
	}

	if scene.Configuration == nil {
		scene.Configuration = new(model.Configuration)
		scene.Configuration.Mu = sync.Mutex{}
	}

	if scene.Configuration.ParameterizedFile == nil {
		scene.Configuration.ParameterizedFile = new(model.ParameterizedFile)
	}
	if scene.Configuration.ParameterizedFile.VariableNames == nil {
		scene.Configuration.ParameterizedFile.VariableNames = new(model.VariableNames)
		scene.Configuration.ParameterizedFile.VariableNames.VarMapLists = make(map[string]*model.VarMapList)
	}
	if scene.Configuration.SceneVariable == nil {
		scene.Configuration.SceneVariable = new(model.GlobalVariable)
	}

	if scene.Configuration.SceneVariable.Variable == nil {
		scene.Configuration.SceneVariable.Variable = []*model.VarForm{}
	}
	if scene.Configuration.SceneVariable.Header == nil {
		scene.Configuration.SceneVariable.Header = new(model.Header)
	}
	if scene.Configuration.SceneVariable.Header.Parameter == nil {
		scene.Configuration.SceneVariable.Header.Parameter = []*model.VarForm{}
	}
	if scene.Configuration.SceneVariable.Cookie == nil {
		scene.Configuration.SceneVariable.Cookie = new(model.Cookie)
	}
	if scene.Configuration.SceneVariable.Cookie.Parameter == nil {
		scene.Configuration.SceneVariable.Cookie.Parameter = []*model.VarForm{}
	}

	if scene.Configuration.SceneVariable.Assert == nil {
		scene.Configuration.SceneVariable.Assert = []*model.AssertionText{}
	}

	if plan.GlobalVariable != nil {
		if scene.Configuration.SceneVariable == nil {
			scene.Configuration.SceneVariable = new(model.GlobalVariable)
		}
		plan.GlobalVariable.InitReplace()
		plan.GlobalVariable.SupToSub(scene.Configuration.SceneVariable)

	}

	if scene.Configuration != nil && scene.Configuration.SceneVariable != nil {
		scene.Configuration.SceneVariable.InitReplace()
	}

	// 分解任务
	TaskDecomposition(plan, resultDataMsgCh, mongoClient, requestCollection)
}

// TaskDecomposition 分解任务
func TaskDecomposition(plan *model.Plan, resultDataMsgCh chan *model.ResultDataMsg, mongoClient *mongo.Client, mongoCollection *mongo.Collection) {
	defer close(resultDataMsgCh)
	defer mongoClient.Disconnect(context.TODO())
	scene := plan.Scene
	scene.TeamId = plan.TeamId
	scene.ReportId = plan.ReportId

	if scene.Configuration.ParameterizedFile == nil {
		scene.Configuration.ParameterizedFile = new(model.ParameterizedFile)
	}
	if scene.Configuration.ParameterizedFile.VariableNames == nil {
		scene.Configuration.ParameterizedFile.VariableNames = new(model.VariableNames)
	}
	if scene.Configuration.ParameterizedFile.VariableNames.VarMapLists == nil {
		scene.Configuration.ParameterizedFile.VariableNames.VarMapLists = make(map[string]*model.VarMapList)
	}
	configuration := scene.Configuration
	if configuration.ParameterizedFile == nil {
		configuration.ParameterizedFile = new(model.ParameterizedFile)
	}

	p := configuration.ParameterizedFile
	p.VariableNames.Mu = sync.Mutex{}
	//teamId := strconv.FormatInt(plan.TeamId, 10)
	//p.DownLoadFile(teamId, plan.ReportId)
	p.UseFile()

	var sqlMap = new(sync.Map)
	if scene.Prepositions != nil && len(scene.Prepositions) > 0 {
		for _, preposition := range scene.Prepositions {
			if preposition.IsDisabled == 1 {
				continue
			}
			preposition.Exec(scene, mongoCollection, sqlMap)
		}
	}
	if p.VariableNames.VarMapLists == nil {
		p.VariableNames.VarMapLists = make(map[string]*model.VarMapList)
	}

	sqlMap.Range(func(key, value any) bool {
		if _, ok := p.VariableNames.VarMapLists[key.(string)]; !ok {
			p.VariableNames.VarMapLists[key.(string)] = new(model.VarMapList)
		}
		switch fmt.Sprintf("%T", value) {
		case "string":
			p.VariableNames.VarMapLists[key.(string)].Value = append(p.VariableNames.VarMapLists[key.(string)].Value, value.(string))
		case "[]string":
			p.VariableNames.VarMapLists[key.(string)].Value = append(p.VariableNames.VarMapLists[key.(string)].Value, value.([]string)...)
		}

		return true
	})

	var reportMsg = &model.ResultDataMsg{}
	if plan.MachineNum <= 0 {
		plan.MachineNum = 1
	}
	reportMsg.TeamId = plan.TeamId
	reportMsg.PlanId = plan.PlanId
	reportMsg.SceneId = scene.SceneId
	reportMsg.SceneName = scene.SceneName
	reportMsg.PlanName = plan.PlanName
	reportMsg.ReportId = plan.ReportId
	reportMsg.ReportName = plan.ReportName
	reportMsg.MachineNum = plan.MachineNum
	reportMsg.MachineIp = middlewares.LocalIp + fmt.Sprintf("_%d", config.Conf.Heartbeat.Port)

	var startMsg = &model.ResultDataMsg{}
	startMsg.TeamId = plan.TeamId
	startMsg.PlanId = plan.PlanId
	startMsg.SceneId = scene.SceneId
	startMsg.SceneName = scene.SceneName
	startMsg.PlanName = plan.PlanName
	startMsg.ReportId = plan.ReportId
	startMsg.ReportName = plan.ReportName
	startMsg.MachineNum = plan.MachineNum
	startMsg.Timestamp = time.Now().UnixMilli()
	startMsg.Start = true
	resultDataMsgCh <- startMsg

	defer func() {
		if err := recover(); err != nil {
			// 发送结束消息时间戳
			startMsg.Start = false
			startMsg.End = true
			startMsg.Timestamp = time.Now().UnixMilli()
			resultDataMsgCh <- startMsg

			debugMsg := &model.DebugMsg{}
			debugMsg.TeamId = plan.TeamId
			debugMsg.PlanId = plan.PlanId
			debugMsg.ReportId = plan.ReportId
			debugMsg.End = true
			model.Insert(mongoCollection, debugMsg, middlewares.LocalIp)
			log.Logger.Info(fmt.Sprintf("机器ip:%s, 团队: %s, 计划：%s， 报告： %s, %v", middlewares.LocalIp, plan.TeamId, plan.PlanId, plan.ReportId, err))
		}
	}()

	var msg string
	switch scene.ConfigTask.Mode {
	case model.ConcurrentModel:
		msg = execution.ConcurrentModel(scene, configuration, reportMsg, resultDataMsgCh, mongoCollection)
	case model.ErrorRateModel:
		msg = execution.ErrorRateModel(scene, configuration, reportMsg, resultDataMsgCh, mongoCollection)
	case model.LadderModel:
		msg = execution.LadderModel(scene, configuration, reportMsg, resultDataMsgCh, mongoCollection)
	case model.RTModel:
		msg = execution.RTModel(scene, configuration, reportMsg, resultDataMsgCh, mongoCollection)
	case model.RpsModel:
		msg = execution.RPSModel(scene, configuration, reportMsg, resultDataMsgCh, mongoCollection)
	case model.RoundModel:
		msg = execution.RoundModel(scene, configuration, reportMsg, resultDataMsgCh, mongoCollection)
	default:
		var machines []string
		msg = "任务类型不存在"
		machine := reportMsg.MachineIp
		machines = append(machines, machine)
		tools.SendStopStressReport(machines, plan.TeamId, plan.PlanId, plan.ReportId)
	}
	// 发送结束消息时间戳
	startMsg.Start = false
	startMsg.End = true
	startMsg.Timestamp = time.Now().UnixMilli()
	resultDataMsgCh <- startMsg

	debugMsg := new(model.DebugMsg)
	debugMsg.TeamId = plan.TeamId
	debugMsg.PlanId = plan.PlanId
	debugMsg.ReportId = plan.ReportId
	debugMsg.End = true
	model.Insert(mongoCollection, debugMsg, middlewares.LocalIp)
	log.Logger.Info(fmt.Sprintf("机器ip:%s, 团队: %s, 计划：%s， 报告： %s, %s", middlewares.LocalIp, plan.TeamId, plan.PlanId, plan.ReportId, msg))
}

// DebugScene 场景调试
func DebugScene(scene model.Scene) {
	mongoClient, err := model.NewMongoClient(
		config.Conf.Mongo.DSN,
		middlewares.LocalIp)
	if err != nil {
		log.Logger.Error(fmt.Sprintf("机器ip:%s, 连接mongo错误：%s", middlewares.LocalIp, err))
		return
	}

	if scene.Configuration == nil {
		scene.Configuration = new(model.Configuration)
		scene.Configuration.Mu = sync.Mutex{}
	}

	if scene.Configuration.ParameterizedFile == nil {
		scene.Configuration.ParameterizedFile = new(model.ParameterizedFile)
	}
	if scene.Configuration.ParameterizedFile.VariableNames == nil {
		scene.Configuration.ParameterizedFile.VariableNames = new(model.VariableNames)
		scene.Configuration.ParameterizedFile.VariableNames.VarMapLists = make(map[string]*model.VarMapList)
	}
	if scene.Configuration.SceneVariable == nil {
		scene.Configuration.SceneVariable = new(model.GlobalVariable)
	}

	if scene.Configuration.SceneVariable.Variable == nil {
		scene.Configuration.SceneVariable.Variable = []*model.VarForm{}
	}
	if scene.Configuration.SceneVariable.Header == nil {
		scene.Configuration.SceneVariable.Header = new(model.Header)
	}
	if scene.Configuration.SceneVariable.Header.Parameter == nil {
		scene.Configuration.SceneVariable.Header.Parameter = []*model.VarForm{}
	}
	if scene.Configuration.SceneVariable.Cookie == nil {
		scene.Configuration.SceneVariable.Cookie = new(model.Cookie)
	}
	if scene.Configuration.SceneVariable.Cookie.Parameter == nil {
		scene.Configuration.SceneVariable.Cookie.Parameter = []*model.VarForm{}
	}

	if scene.Configuration.SceneVariable.Assert == nil {
		scene.Configuration.SceneVariable.Assert = []*model.AssertionText{}
	}

	if scene.GlobalVariable != nil {
		scene.GlobalVariable.InitReplace()
		scene.GlobalVariable.SupToSub(scene.Configuration.SceneVariable)
	}
	if scene.Configuration.SceneVariable != nil {
		scene.Configuration.SceneVariable.InitReplace()
	}

	configuration := scene.Configuration
	if configuration.ParameterizedFile == nil {
		configuration.ParameterizedFile = new(model.ParameterizedFile)
	}
	p := scene.Configuration.ParameterizedFile
	p.VariableNames.Mu = sync.Mutex{}
	p.UseFile()

	scene.Debug = constant.All
	defer mongoClient.Disconnect(context.TODO())
	mongoCollection := model.NewCollection(config.Conf.Mongo.DataBase, config.Conf.Mongo.SceneDebugTable, mongoClient)

	var sqlMap = new(sync.Map)
	if scene.Prepositions != nil && len(scene.Prepositions) > 0 {
		for _, preposition := range scene.Prepositions {
			if preposition.IsDisabled == 1 {
				continue
			}
			preposition.Exec(scene, mongoCollection, sqlMap)
		}
	}

	if p.VariableNames.VarMapLists == nil {
		p.VariableNames.VarMapLists = make(map[string]*model.VarMapList)
	}
	sqlMap.Range(func(key, value any) bool {
		if _, ok := p.VariableNames.VarMapLists[key.(string)]; !ok {
			p.VariableNames.VarMapLists[key.(string)] = new(model.VarMapList)
		}
		switch fmt.Sprintf("%T", value) {
		case "string":
			p.VariableNames.VarMapLists[key.(string)].Value = append(p.VariableNames.VarMapLists[key.(string)].Value, value.(string))
		case "[]string":
			p.VariableNames.VarMapLists[key.(string)].Value = append(p.VariableNames.VarMapLists[key.(string)].Value, value.([]string)...)
		}

		return true
	})

	golink.DisposeScene(constant.SceneType, scene, configuration, nil, nil, mongoCollection)
	log.Logger.Info(fmt.Sprintf("机器ip:%s, 团队: %s, 场景：%s, 调试结束！", middlewares.LocalIp, scene.TeamId, scene.SceneName))

}

// DebugApi api调试
func DebugApi(debugApi model.Api) {

	var globalVar = new(sync.Map)

	if debugApi.GlobalVariable != nil {
		debugApi.GlobalVariable.InitReplace()
		if debugApi.Configuration == nil {
			debugApi.Configuration = new(model.Configuration)
		}
		if debugApi.Configuration.SceneVariable == nil {
			debugApi.Configuration.SceneVariable = new(model.GlobalVariable)
		}
		debugApi.GlobalVariable.SupToSub(debugApi.Configuration.SceneVariable)
		debugApi.ApiVariable = new(model.GlobalVariable)
		debugApi.Configuration.SceneVariable.SupToSub(debugApi.ApiVariable)
		debugApi.ApiVariable.InitReplace()
	} else {
		if debugApi.Configuration != nil && debugApi.Configuration.SceneVariable != nil {
			debugApi.Configuration.SceneVariable.InitReplace()
			debugApi.ApiVariable = new(model.GlobalVariable)
			debugApi.Configuration.SceneVariable.SupToSub(debugApi.ApiVariable)

		}
	}
	if debugApi.ApiVariable.Variable != nil {
		for _, variable := range debugApi.Configuration.SceneVariable.Variable {
			if variable.IsChecked != constant.Open {
				continue
			}
			globalVar.Store(variable.Key, variable.Value)
		}

	}

	if debugApi.Configuration != nil {
		if debugApi.Configuration.ParameterizedFile != nil {
			if debugApi.Configuration.ParameterizedFile.VariableNames == nil {
				debugApi.Configuration.ParameterizedFile.VariableNames = new(model.VariableNames)
			}
			debugApi.Configuration.ParameterizedFile.UseFile()

			if debugApi.Configuration.ParameterizedFile.VariableNames.VarMapLists != nil {
				for k, v := range debugApi.Configuration.ParameterizedFile.VariableNames.VarMapLists {
					globalVar.Store(k, v.Value[0])
				}
			}
		}
	}

	event := model.Event{}
	event.Api = debugApi
	event.TeamId = debugApi.TeamId
	event.Weight = 100
	event.Id = "接口调试"
	event.Debug = constant.All
	// 新建mongo客户端连接，用于发送debug数据
	mongoClient, err := model.NewMongoClient(
		config.Conf.Mongo.DSN,
		middlewares.LocalIp)
	if err != nil {
		log.Logger.Error(fmt.Sprintf("机器ip:%s, 连接mongo错误：%s", middlewares.LocalIp, err.Error()))
		return
	}
	globalVar.Range(func(key, value any) bool {
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
	defer mongoClient.Disconnect(context.TODO())
	mongoCollection := model.NewCollection(config.Conf.Mongo.DataBase, config.Conf.Mongo.ApiDebugTable, mongoClient)

	golink.DisposeRequest(nil, nil, nil, globalVar, event, mongoCollection)
	log.Logger.Info(fmt.Sprintf("机器ip:%s, 团队：%s, 测试对象：%s, 调试结束！", middlewares.LocalIp, debugApi.TeamId, debugApi.Name))

}

// DebugSql sql调试
//func DebugSql(debugSql model.SQLDetail) {
//
//	var globalVar = new(sync.Map)
//
//	if debugSql.GlobalVariable != nil {
//		debugSql.GlobalVariable.InitReplace()
//		if debugSql.Configuration == nil {
//			debugSql.Configuration = new(model.Configuration)
//		}
//		if debugSql.Configuration.SceneVariable == nil {
//			debugSql.Configuration.SceneVariable = new(model.GlobalVariable)
//		}
//		debugSql.GlobalVariable.SupToSub(debugSql.Configuration.SceneVariable)
//		debugSql.SqlVariable = new(model.GlobalVariable)
//		debugSql.Configuration.SceneVariable.SupToSub(debugSql.SqlVariable)
//		debugSql.SqlVariable.InitReplace()
//	} else {
//		if debugSql.Configuration != nil && debugSql.Configuration.SceneVariable != nil {
//			debugSql.Configuration.SceneVariable.InitReplace()
//			debugSql.SqlVariable = new(model.GlobalVariable)
//			debugSql.Configuration.SceneVariable.SupToSub(debugSql.SqlVariable)
//
//		}
//	}
//
//	if debugSql.Configuration != nil {
//		if debugSql.Configuration.ParameterizedFile != nil {
//			if debugSql.Configuration.ParameterizedFile.VariableNames == nil {
//				debugSql.Configuration.ParameterizedFile.VariableNames = new(model.VariableNames)
//			}
//			debugSql.Configuration.ParameterizedFile.UseFile()
//
//			if debugSql.Configuration.ParameterizedFile.VariableNames.VarMapList != nil {
//				for k, v := range debugSql.Configuration.ParameterizedFile.VariableNames.VarMapList {
//					globalVar.Store(k, v[0])
//				}
//			}
//		}
//	}
//
//	if debugSql.SqlVariable.Variable != nil {
//		for _, variable := range debugSql.Configuration.SceneVariable.Variable {
//			if variable.IsChecked != constant.Open {
//				continue
//			}
//			globalVar.Store(variable.Key, variable.Value)
//		}
//
//	}
//
//	event := model.Event{}
//	event.Api.SQL = debugSql
//	//event.TeamId = debugSql.TeamId
//	event.Weight = 100
//	event.Id = "接口调试"
//	event.Debug = constant.All
//	// 新建mongo客户端连接，用于发送debug数据
//	mongoClient, err := model.NewMongoClient(
//		config.Conf.Mongo.DSN,
//		middlewares.LocalIp)
//	if err != nil {
//		log.Logger.Error(fmt.Sprintf("机器ip:%s, 连接mongo错误：%s", middlewares.LocalIp, err.Error()))
//		return
//	}
//	globalVar.Range(func(key, value any) bool {
//		values := tools.FindAllDestStr(value.(string), "{{(.*?)}}")
//		if values == nil {
//			return true
//		}
//		for _, v := range values {
//			if len(v) < 2 {
//				return true
//			}
//			realVar := tools.ParsFunc(v[1])
//			if realVar != v[1] {
//				value = strings.Replace(value.(string), v[0], realVar, -1)
//				globalVar.Store(key, value)
//			}
//		}
//		return true
//	})
//	defer mongoClient.Disconnect(context.TODO())
//	mongoCollection := model.NewCollection(config.Conf.Mongo.DataBase, config.Conf.Mongo.SqlDebugTable, mongoClient)
//
//	golink.DisposeSql(nil, nil, nil, globalVar, event, mongoCollection)
//	//log.Logger.Info(fmt.Sprintf("机器ip:%s, 团队：%s, sql：%s, 调试结束！", middlewares.LocalIp, debugSql.TeamId, debugSql.Name))
//
//}

// DebugTcp tcp调试
//func DebugTcp(debugTcp model.TCPDetail) {
//
//	var globalVar = new(sync.Map)
//
//	if debugTcp.GlobalVariable != nil {
//		debugTcp.GlobalVariable.InitReplace()
//		if debugTcp.Configuration == nil {
//			debugTcp.Configuration = new(model.Configuration)
//		}
//		if debugTcp.Configuration.SceneVariable == nil {
//			debugTcp.Configuration.SceneVariable = new(model.GlobalVariable)
//		}
//		debugTcp.GlobalVariable.SupToSub(debugTcp.Configuration.SceneVariable)
//		debugTcp.SqlVariable = new(model.GlobalVariable)
//		debugTcp.Configuration.SceneVariable.SupToSub(debugTcp.SqlVariable)
//		debugTcp.SqlVariable.InitReplace()
//	} else {
//		if debugTcp.Configuration != nil && debugTcp.Configuration.SceneVariable != nil {
//			debugTcp.Configuration.SceneVariable.InitReplace()
//			debugTcp.SqlVariable = new(model.GlobalVariable)
//			debugTcp.Configuration.SceneVariable.SupToSub(debugTcp.SqlVariable)
//
//		}
//	}
//
//	if debugTcp.Configuration != nil {
//		if debugTcp.Configuration.ParameterizedFile != nil {
//			if debugTcp.Configuration.ParameterizedFile.VariableNames == nil {
//				debugTcp.Configuration.ParameterizedFile.VariableNames = new(model.VariableNames)
//			}
//			debugTcp.Configuration.ParameterizedFile.UseFile()
//
//			if debugTcp.Configuration.ParameterizedFile.VariableNames.VarMapList != nil {
//				for k, v := range debugTcp.Configuration.ParameterizedFile.VariableNames.VarMapList {
//					globalVar.Store(k, v[0])
//				}
//			}
//		}
//	}
//
//	if debugTcp.SqlVariable.Variable != nil {
//		for _, variable := range debugTcp.Configuration.SceneVariable.Variable {
//			if variable.IsChecked != constant.Open {
//				continue
//			}
//			globalVar.Store(variable.Key, variable.Value)
//		}
//
//	}
//
//	event := model.Event{}
//	event.Api.TCP = debugTcp
//	event.TeamId = debugTcp.TeamId
//	event.Weight = 100
//	event.Id = "接口调试"
//	event.Debug = constant.All
//	// 新建mongo客户端连接，用于发送debug数据
//	mongoClient, err := model.NewMongoClient(
//		config.Conf.Mongo.DSN,
//		middlewares.LocalIp)
//	if err != nil {
//		log.Logger.Error(fmt.Sprintf("机器ip:%s, 连接mongo错误：%s", middlewares.LocalIp, err.Error()))
//		return
//	}
//	globalVar.Range(func(key, value any) bool {
//		values := tools.FindAllDestStr(value.(string), "{{(.*?)}}")
//		if values == nil {
//			return true
//		}
//		for _, v := range values {
//			if len(v) < 2 {
//				return true
//			}
//			realVar := tools.ParsFunc(v[1])
//			if realVar != v[1] {
//				value = strings.Replace(value.(string), v[0], realVar, -1)
//				globalVar.Store(key, value)
//			}
//		}
//		return true
//	})
//	defer mongoClient.Disconnect(context.TODO())
//	mongoCollection := model.NewCollection(config.Conf.Mongo.DataBase, config.Conf.Mongo.TcpDebugTable, mongoClient)
//
//	golink.DisposeTcp(nil, nil, nil, globalVar, event, mongoCollection)
//	log.Logger.Info(fmt.Sprintf("机器ip:%s, 团队：%s, sql：%s, 调试结束！", middlewares.LocalIp, debugTcp.TeamId, debugTcp.Name))
//
//}

// DebugWs websocket调试
//func DebugWs(debugWs model.WebsocketDetail) {
//
//	var globalVar = new(sync.Map)
//
//	if debugWs.GlobalVariable != nil {
//		debugWs.GlobalVariable.InitReplace()
//		if debugWs.Configuration == nil {
//			debugWs.Configuration = new(model.Configuration)
//		}
//		if debugWs.Configuration.SceneVariable == nil {
//			debugWs.Configuration.SceneVariable = new(model.GlobalVariable)
//		}
//		debugWs.GlobalVariable.SupToSub(debugWs.Configuration.SceneVariable)
//		debugWs.WsVariable = new(model.GlobalVariable)
//		debugWs.Configuration.SceneVariable.SupToSub(debugWs.WsVariable)
//		debugWs.WsVariable.InitReplace()
//	} else {
//		if debugWs.Configuration != nil && debugWs.Configuration.SceneVariable != nil {
//			debugWs.Configuration.SceneVariable.InitReplace()
//			debugWs.WsVariable = new(model.GlobalVariable)
//			debugWs.Configuration.SceneVariable.SupToSub(debugWs.WsVariable)
//
//		}
//	}
//
//	if debugWs.Configuration != nil {
//		if debugWs.Configuration.ParameterizedFile != nil {
//			if debugWs.Configuration.ParameterizedFile.VariableNames == nil {
//				debugWs.Configuration.ParameterizedFile.VariableNames = new(model.VariableNames)
//			}
//			debugWs.Configuration.ParameterizedFile.UseFile()
//
//			if debugWs.Configuration.ParameterizedFile.VariableNames.VarMapList != nil {
//				for k, v := range debugWs.Configuration.ParameterizedFile.VariableNames.VarMapList {
//					globalVar.Store(k, v[0])
//				}
//			}
//		}
//	}
//
//	if debugWs.WsVariable.Variable != nil {
//		for _, variable := range debugWs.Configuration.SceneVariable.Variable {
//			if variable.IsChecked != constant.Open {
//				continue
//			}
//			globalVar.Store(variable.Key, variable.Value)
//		}
//
//	}
//
//	event := model.Event{}
//	event.Api.Ws = debugWs
//	event.TeamId = debugWs.TeamId
//	event.Weight = 100
//	event.Id = "接口调试"
//	event.Debug = constant.All
//	// 新建mongo客户端连接，用于发送debug数据
//	mongoClient, err := model.NewMongoClient(
//		config.Conf.Mongo.DSN,
//		middlewares.LocalIp)
//	if err != nil {
//		log.Logger.Error(fmt.Sprintf("机器ip:%s, 连接mongo错误：%s", middlewares.LocalIp, err.Error()))
//		return
//	}
//	globalVar.Range(func(key, value any) bool {
//		values := tools.FindAllDestStr(value.(string), "{{(.*?)}}")
//		if values == nil {
//			return true
//		}
//		for _, v := range values {
//			if len(v) < 2 {
//				return true
//			}
//			realVar := tools.ParsFunc(v[1])
//			if realVar != v[1] {
//				value = strings.Replace(value.(string), v[0], realVar, -1)
//				globalVar.Store(key, value)
//			}
//		}
//		return true
//	})
//	defer mongoClient.Disconnect(context.TODO())
//	mongoCollection := model.NewCollection(config.Conf.Mongo.DataBase, config.Conf.Mongo.WsDebugTable, mongoClient)
//
//	golink.DisposeWs(nil, nil, nil, globalVar, event, mongoCollection)
//	log.Logger.Info(fmt.Sprintf("机器ip:%s, 团队：%s, sql：%s, 调试结束！", middlewares.LocalIp, debugWs.TeamId, debugWs.Name))
//
//}

// DebugDubbo dubbo调试
//func DebugDubbo(dubbo model.DubboDetail) {
//
//	var globalVar = new(sync.Map)
//
//	if dubbo.GlobalVariable != nil {
//		dubbo.GlobalVariable.InitReplace()
//		if dubbo.Configuration == nil {
//			dubbo.Configuration = new(model.Configuration)
//		}
//		if dubbo.Configuration.SceneVariable == nil {
//			dubbo.Configuration.SceneVariable = new(model.GlobalVariable)
//		}
//		dubbo.GlobalVariable.SupToSub(dubbo.Configuration.SceneVariable)
//		dubbo.DubboVariable = new(model.GlobalVariable)
//		dubbo.Configuration.SceneVariable.SupToSub(dubbo.DubboVariable)
//		dubbo.DubboVariable.InitReplace()
//	} else {
//		if dubbo.Configuration != nil && dubbo.Configuration.SceneVariable != nil {
//			dubbo.Configuration.SceneVariable.InitReplace()
//			dubbo.DubboVariable = new(model.GlobalVariable)
//			dubbo.Configuration.SceneVariable.SupToSub(dubbo.DubboVariable)
//
//		}
//	}
//
//	if dubbo.Configuration != nil {
//		if dubbo.Configuration.ParameterizedFile != nil {
//			if dubbo.Configuration.ParameterizedFile.VariableNames == nil {
//				dubbo.Configuration.ParameterizedFile.VariableNames = new(model.VariableNames)
//			}
//			dubbo.Configuration.ParameterizedFile.UseFile()
//
//			if dubbo.Configuration.ParameterizedFile.VariableNames.VarMapList != nil {
//				for k, v := range dubbo.Configuration.ParameterizedFile.VariableNames.VarMapList {
//					globalVar.Store(k, v[0])
//				}
//			}
//		}
//	}
//
//	if dubbo.DubboVariable.Variable != nil {
//		for _, variable := range dubbo.Configuration.SceneVariable.Variable {
//			if variable.IsChecked != constant.Open {
//				continue
//			}
//			globalVar.Store(variable.Key, variable.Value)
//		}
//
//	}
//
//	event := model.Event{}
//	event.Api.DubboDetail = dubbo
//	event.TeamId = dubbo.TeamId
//	event.Weight = 100
//	event.Id = "接口调试"
//	event.Debug = constant.All
//	// 新建mongo客户端连接，用于发送debug数据
//	mongoClient, err := model.NewMongoClient(
//		config.Conf.Mongo.DSN,
//		middlewares.LocalIp)
//	if err != nil {
//		log.Logger.Error(fmt.Sprintf("机器ip:%s, 连接mongo错误：%s", middlewares.LocalIp, err.Error()))
//		return
//	}
//	globalVar.Range(func(key, value any) bool {
//		values := tools.FindAllDestStr(value.(string), "{{(.*?)}}")
//		if values == nil {
//			return true
//		}
//		for _, v := range values {
//			if len(v) < 2 {
//				return true
//			}
//			realVar := tools.ParsFunc(v[1])
//			if realVar != v[1] {
//				value = strings.Replace(value.(string), v[0], realVar, -1)
//				globalVar.Store(key, value)
//			}
//		}
//		return true
//	})
//	defer mongoClient.Disconnect(context.TODO())
//	mongoCollection := model.NewCollection(config.Conf.Mongo.DataBase, config.Conf.Mongo.DubboDebugTable, mongoClient)
//
//	golink.DisposeDubbo(nil, nil, nil, globalVar, event, mongoCollection)
//	log.Logger.Info(fmt.Sprintf("机器ip:%s, 团队：%s, sql：%s, 调试结束！", middlewares.LocalIp, dubbo.TeamId, dubbo.Name))
//
//}

// DebugMqtt mqtt调试
//func DebugMqtt(mqtt model.MQTT) {
//
//	var globalVar = new(sync.Map)
//
//	if mqtt.GlobalVariable != nil {
//		mqtt.GlobalVariable.InitReplace()
//		if mqtt.Configuration == nil {
//			mqtt.Configuration = new(model.Configuration)
//		}
//		if mqtt.Configuration.SceneVariable == nil {
//			mqtt.Configuration.SceneVariable = new(model.GlobalVariable)
//		}
//		mqtt.GlobalVariable.SupToSub(mqtt.Configuration.SceneVariable)
//		mqtt.MqttVariable = new(model.GlobalVariable)
//		mqtt.Configuration.SceneVariable.SupToSub(mqtt.MqttVariable)
//		mqtt.MqttVariable.InitReplace()
//	} else {
//		if mqtt.Configuration != nil && mqtt.Configuration.SceneVariable != nil {
//			mqtt.Configuration.SceneVariable.InitReplace()
//			mqtt.MqttVariable = new(model.GlobalVariable)
//			mqtt.Configuration.SceneVariable.SupToSub(mqtt.MqttVariable)
//
//		}
//	}
//
//	if mqtt.Configuration != nil {
//		if mqtt.Configuration.ParameterizedFile != nil {
//			if mqtt.Configuration.ParameterizedFile.VariableNames == nil {
//				mqtt.Configuration.ParameterizedFile.VariableNames = new(model.VariableNames)
//			}
//			mqtt.Configuration.ParameterizedFile.UseFile()
//
//			if mqtt.Configuration.ParameterizedFile.VariableNames.VarMapList != nil {
//				for k, v := range mqtt.Configuration.ParameterizedFile.VariableNames.VarMapList {
//					globalVar.Store(k, v[0])
//				}
//			}
//		}
//	}
//
//	if mqtt.MqttVariable.Variable != nil {
//		for _, variable := range mqtt.Configuration.SceneVariable.Variable {
//			if variable.IsChecked != model.Open {
//				continue
//			}
//			globalVar.Store(variable.Key, variable.Value)
//		}
//
//	}
//
//	event := model.Event{}
//	event.MQTT = mqtt
//	event.TeamId = mqtt.TeamId
//	event.Weight = 100
//	event.Id = "接口调试"
//	event.Debug = model.All
//	// 新建mongo客户端连接，用于发送debug数据
//	mongoClient, err := model.NewMongoClient(
//		config.Conf.Mongo.DSN,
//		middlewares.LocalIp)
//	if err != nil {
//		log.Logger.Error(fmt.Sprintf("机器ip:%s, 连接mongo错误：%s", middlewares.LocalIp, err.Error()))
//		return
//	}
//	globalVar.Range(func(key, value any) bool {
//		values := tools.FindAllDestStr(value.(string), "{{(.*?)}}")
//		if values == nil {
//			return true
//		}
//		for _, v := range values {
//			if len(v) < 2 {
//				return true
//			}
//			realVar := tools.ParsFunc(v[1])
//			if realVar != v[1] {
//				value = strings.Replace(value.(string), v[0], realVar, -1)
//				globalVar.Store(key, value)
//			}
//		}
//		return true
//	})
//	defer mongoClient.Disconnect(context.TODO())
//	mongoCollection := model.NewCollection(config.Conf.Mongo.DataBase, config.Conf.Mongo.TcpDebugTable, mongoClient)
//
//	golink.DisposeMqtt(nil, nil, nil, globalVar, event, mongoCollection)
//	log.Logger.Info(fmt.Sprintf("机器ip:%s, 团队：%s, sql：%s, 调试结束！", middlewares.LocalIp, mqtt.TeamId, mqtt.Name))
//
//}
