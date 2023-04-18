// Package server 压测启动
package server

import (
	"context"
	"fmt"
	"github.com/Runner-Go-Team/RunnerGo-engine-open/config"
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

	var wg = &sync.WaitGroup{}

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
		scene.Configuration.ParameterizedFile.VariableNames.VarMapList = make(map[string][]string)
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
		plan.GlobalVariable.SupToSub(scene.Configuration.SceneVariable)
		scene.Configuration.SceneVariable.InitReplace()
	}

	// 分解任务
	TaskDecomposition(plan, wg, resultDataMsgCh, mongoClient, requestCollection)
}

// TaskDecomposition 分解任务
func TaskDecomposition(plan *model.Plan, wg *sync.WaitGroup, resultDataMsgCh chan *model.ResultDataMsg, mongoClient *mongo.Client, mongoCollection *mongo.Collection) {
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
	if scene.Configuration.ParameterizedFile.VariableNames.VarMapList == nil {
		scene.Configuration.ParameterizedFile.VariableNames.VarMapList = make(map[string][]string)
	}
	configuration := scene.Configuration
	if configuration.ParameterizedFile != nil {
		p := configuration.ParameterizedFile
		p.VariableNames.Mu = sync.Mutex{}
		//teamId := strconv.FormatInt(plan.TeamId, 10)
		//p.DownLoadFile(teamId, plan.ReportId)
		p.UseFile()

	}

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
	var msg string
	switch scene.ConfigTask.Mode {
	case model.ConcurrentModel:
		msg = execution.ConcurrentModel(wg, scene, configuration, reportMsg, resultDataMsgCh, mongoCollection)
	case model.ErrorRateModel:
		msg = execution.ErrorRateModel(wg, scene, configuration, reportMsg, resultDataMsgCh, mongoCollection)
	case model.LadderModel:
		msg = execution.LadderModel(wg, scene, configuration, reportMsg, resultDataMsgCh, mongoCollection)
	case model.RTModel:
		msg = execution.RTModel(wg, scene, configuration, reportMsg, resultDataMsgCh, mongoCollection)
	case model.RpsModel:
		msg = execution.RPSModel(wg, scene, configuration, reportMsg, resultDataMsgCh, mongoCollection)
	default:
		var machines []string
		msg = "任务类型不存在"
		machine := reportMsg.MachineIp
		machines = append(machines, machine)
		tools.SendStopStressReport(machines, plan.TeamId, plan.PlanId, plan.ReportId)
	}
	wg.Wait()
	// 发送结束消息时间戳
	startMsg.Start = false
	startMsg.End = true
	startMsg.Timestamp = time.Now().UnixMilli()
	resultDataMsgCh <- startMsg
	debugMsg := make(map[string]interface{})
	debugMsg["team_id"] = plan.TeamId
	debugMsg["plan_id"] = plan.PlanId
	debugMsg["report_id"] = plan.ReportId
	debugMsg["end"] = true
	model.Insert(mongoCollection, debugMsg, middlewares.LocalIp)
	log.Logger.Info(fmt.Sprintf("机器ip:%s, 团队: %s, 计划：%s， 报告： %s, %s", middlewares.LocalIp, plan.TeamId, plan.PlanId, plan.ReportId, msg))
}

// DebugScene 场景调试
func DebugScene(scene model.Scene) {
	wg := &sync.WaitGroup{}
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
		scene.Configuration.ParameterizedFile.VariableNames.VarMapList = make(map[string][]string)
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
		scene.GlobalVariable.SupToSub(scene.Configuration.SceneVariable)
		scene.Configuration.SceneVariable.InitReplace()
	}

	scene.Configuration.SceneVariable.InitReplace()

	configuration := scene.Configuration
	if configuration.ParameterizedFile != nil {
		p := scene.Configuration.ParameterizedFile
		p.VariableNames.Mu = sync.Mutex{}
		//teamId := strconv.FormatInt(plan.TeamId, 10)
		//p.DownLoadFile(teamId, plan.ReportId)
		p.UseFile()
	}
	scene.Debug = model.All
	defer mongoClient.Disconnect(context.TODO())
	mongoCollection := model.NewCollection(config.Conf.Mongo.DataBase, config.Conf.Mongo.SceneDebugTable, mongoClient)
	var sceneWg = &sync.WaitGroup{}
	golink.DisposeScene(wg, sceneWg, model.SceneType, scene, configuration, nil, nil, mongoCollection)
	wg.Wait()
	sceneWg.Wait()
	log.Logger.Info(fmt.Sprintf("机器ip:%s, 团队: %s, 场景：%s, 调试结束！", middlewares.LocalIp, scene.TeamId, scene.SceneName))

}

// DebugApi api调试
func DebugApi(debugApi model.Api) {

	var globalVar = new(sync.Map)

	if debugApi.GlobalVariable != nil {
		if debugApi.GlobalVariable.Variable != nil {
			for _, kv := range debugApi.GlobalVariable.Variable {
				if kv.IsChecked == model.Open {
					globalVar.Store(kv.Key, kv.Value)
				}
			}
		}

	}

	if debugApi.Configuration != nil {
		debugApi.Configuration.SceneVariable.SupToSub(debugApi.GlobalVariable)
		debugApi.GlobalVariable.InitReplace()
		if debugApi.GlobalVariable != nil && debugApi.GlobalVariable.Variable != nil {
			for _, kv := range debugApi.GlobalVariable.Variable {
				if kv.IsChecked != model.Open {
					continue
				}
				globalVar.Store(kv.Key, kv.Value)
			}
		}
		if debugApi.Configuration.ParameterizedFile != nil {
			if debugApi.Configuration.ParameterizedFile.VariableNames == nil {
				debugApi.Configuration.ParameterizedFile.VariableNames = new(model.VariableNames)
			}
			debugApi.Configuration.ParameterizedFile.UseFile()

			if debugApi.Configuration.ParameterizedFile.VariableNames.VarMapList != nil {
				for k, v := range debugApi.Configuration.ParameterizedFile.VariableNames.VarMapList {
					globalVar.Store(k, v[0])
				}
			}
		}
	}

	event := model.Event{}
	event.Api = debugApi
	event.TeamId = debugApi.TeamId
	event.Weight = 100
	event.Id = "接口调试"
	event.Debug = model.All
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
	log.Logger.Info(fmt.Sprintf("机器ip:%s, 团队：%s, 接口：%s, 调试结束！", middlewares.LocalIp, debugApi.TeamId, debugApi.Name))

}
