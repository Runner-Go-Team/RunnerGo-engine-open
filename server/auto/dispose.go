package auto

import (
	"context"
	"fmt"
	"github.com/Runner-Go-Team/RunnerGo-engine-open/config"
	"github.com/Runner-Go-Team/RunnerGo-engine-open/constant"
	"github.com/Runner-Go-Team/RunnerGo-engine-open/global"
	"github.com/Runner-Go-Team/RunnerGo-engine-open/log"
	"github.com/Runner-Go-Team/RunnerGo-engine-open/middlewares"
	"github.com/Runner-Go-Team/RunnerGo-engine-open/model"
	"github.com/Runner-Go-Team/RunnerGo-engine-open/model/auto"
	"github.com/Runner-Go-Team/RunnerGo-engine-open/server/golink"
	"github.com/Runner-Go-Team/RunnerGo-engine-open/tools"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	"net/http"
	"sync"
	"time"
)

// DisposeAutoPlan 执行计划
func DisposeAutoPlan(plan *auto.Plan, c *gin.Context) {

	// 如果场景为空或者场景中的事件为空，直接结束该方法
	if plan.Scenes == nil || len(plan.Scenes) < 1 {
		global.ReturnMsg(c, http.StatusBadRequest, "自动化测试-计划的场景不能为空", plan.Scenes)
		return
	}

	if plan.ReportId == "" {
		global.ReturnMsg(c, http.StatusBadRequest, "reportId不能为空", plan)
		return
	}

	if plan.PlanId == "" {
		global.ReturnMsg(c, http.StatusBadRequest, "planId不能为空", plan)
		return
	}

	if plan.TeamId == "" {
		global.ReturnMsg(c, http.StatusBadRequest, "teamId不能为空", plan)
		return
	}

	if plan.GlobalVariable != nil {
		plan.GlobalVariable.InitReplace()
	}

	if plan.ConfigTask == nil {
		log.Logger.Error(fmt.Sprintf("机器ip:%s, 任务配置不能为空", middlewares.LocalIp), plan)
		return
	}

	// 新建mongo客户端连接，用于发送debug数据
	mongoClient, err := model.NewMongoClient(
		config.Conf.Mongo.DSN,
		middlewares.LocalIp)
	if err != nil {
		log.Logger.Error(fmt.Sprintf("机器ip:%s, 连接mongo错误：%s", middlewares.LocalIp, err))
		return
	}

	var wg = &sync.WaitGroup{}
	collection := model.NewCollection(config.Conf.Mongo.DataBase, config.Conf.Mongo.AutoTable, mongoClient)

	if collection == nil {
		log.Logger.Error(fmt.Sprintf("机器ip:%s, mongo创建失败：%s", middlewares.LocalIp, err))
		return
	}
	// 循环读全局变量，如果场景变量不存在则将添加到场景变量中，如果有参数化数据则，将其替换
	for _, scene := range plan.Scenes {
		if scene.Configuration == nil {
			scene.Configuration = new(model.Configuration)
		}
		if plan.GlobalVariable != nil {
			plan.GlobalVariable.SupToSub(scene.Configuration.SceneVariable)
			scene.Configuration.SceneVariable.InitReplace()
		}

		if scene.Configuration.SceneVariable == nil && plan.GlobalVariable != nil {
			scene.Configuration.SceneVariable = plan.GlobalVariable
		}

		if scene.Configuration.ParameterizedFile == nil {
			scene.Configuration.ParameterizedFile = new(model.ParameterizedFile)
		}
		if scene.Configuration.ParameterizedFile.VariableNames == nil {
			scene.Configuration.ParameterizedFile.VariableNames = new(model.VariableNames)
		}
		if scene.Configuration.ParameterizedFile.VariableNames.VarMapLists == nil {
			scene.Configuration.ParameterizedFile.VariableNames.VarMapLists = make(map[string]*model.VarMapList)
		}

		p := scene.Configuration.ParameterizedFile
		p.VariableNames.Mu = sync.Mutex{}
		p.UseFile()

		var sqlMap = new(sync.Map)
		if scene.Prepositions != nil && len(scene.Prepositions) > 0 {
			for _, preposition := range scene.Prepositions {
				if preposition.IsDisabled == 1 {
					continue
				}
				preposition.Exec(scene, collection, sqlMap)
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
	}

	// 设置接收数据缓存
	//resultDataMsgCh := make(chan *model.ResultDataMsg, 10000)
	var reportMsg = new(model.ResultDataMsg)
	reportMsg.TeamId = plan.TeamId
	reportMsg.PlanId = plan.PlanId
	reportMsg.ReportId = plan.ReportId
	go func() {
		sceneDecomposition(plan, wg, reportMsg, nil, collection, mongoClient)
	}()
	global.ReturnMsg(c, http.StatusOK, "开始执行计划", nil)

}

// SceneDecomposition 分解
func sceneDecomposition(plan *auto.Plan, wg *sync.WaitGroup, reportMsg *model.ResultDataMsg, resultDataMsgCh chan *model.ResultDataMsg, collection *mongo.Collection, mongoClient *mongo.Client) {
	defer mongoClient.Disconnect(context.TODO())
	startTime := time.Now().UnixMilli()
	switch plan.ConfigTask.SceneRunMode {
	case constant.AuToOrderMode:
		for _, scene := range plan.Scenes {
			key := fmt.Sprintf("StopAutoPlan:%s:%s:%s", scene.TeamId, scene.PlanId, scene.ReportId)
			err, stop := model.QueryPlanStatus(key)
			if err == nil && stop == "stop" {
				return
			}
			scene.PlanId = plan.PlanId
			scene.TeamId = plan.TeamId
			scene.ReportId = plan.ReportId
			configuration := scene.Configuration

			disposeCase(scene, plan.ConfigTask.SceneRunMode, plan.ConfigTask.CaseRunMode, wg, configuration, reportMsg, resultDataMsgCh, collection)
		}
	case constant.AuToSameMode:
		for _, scene := range plan.Scenes {
			key := fmt.Sprintf("StopAutoPlan:%s:%s:%s", scene.TeamId, scene.PlanId, scene.ReportId)
			err, stop := model.QueryPlanStatus(key)
			if err == nil && stop == "stop" {
				return
			}
			scene.PlanId = plan.PlanId
			scene.TeamId = plan.TeamId
			scene.ReportId = plan.ReportId
			configuration := scene.Configuration
			wg.Add(1)
			//currentWg.Add(1)
			go disposeCase(scene, plan.ConfigTask.SceneRunMode, plan.ConfigTask.CaseRunMode, wg, configuration, reportMsg, resultDataMsgCh, collection)
		}
	}
	wg.Wait()
	tools.SendStopAutoReport(plan.TeamId, plan.PlanId, plan.ReportId, time.Now().UnixMilli()-startTime)
	log.Logger.Info(fmt.Sprintf("机器ip:%s, 团队: %s, 自动化计划: %s, 报告: %s  已完成, 运行：%d毫秒", middlewares.LocalIp, plan.TeamId, plan.PlanId, plan.ReportId, time.Now().UnixMilli()-startTime))

}

func disposeCase(scene model.Scene, sceneRunMode, caseMode int64, wg *sync.WaitGroup, configuration *model.Configuration, reportMsg *model.ResultDataMsg, resultDataMsgCh chan *model.ResultDataMsg, collection *mongo.Collection) {
	if sceneRunMode == constant.AuToSameMode {
		defer wg.Done()
	}
	if scene.Cases == nil || len(scene.Cases) < 1 {
		log.Logger.Debug(fmt.Sprintf("机器ip:%s, 团队: %s, 自动化计划：%s, 报告：%s  自动化测试场景中，用例集不能为空", middlewares.LocalIp, scene.TeamId, scene.PlanId, scene.ReportId))
		return
	}
	for _, c := range scene.Cases {
		key := fmt.Sprintf("StopAutoPlan:%s:%s:%s", scene.TeamId, scene.PlanId, scene.ReportId)
		err, stop := model.QueryPlanStatus(key)
		if err == nil && stop == "stop" {
			return
		}
		if c.IsChecked != constant.Open {
			continue
		}
		c.PlanId = scene.PlanId
		c.TeamId = scene.TeamId
		c.ReportId = scene.ReportId
		c.ParentId = scene.SceneId
		c.Debug = constant.All
		if scene.Configuration != nil {
			c.Configuration = scene.Configuration
		}
		switch caseMode {
		case constant.AuToOrderMode:
			golink.DisposeScene(constant.SceneType, c, configuration, reportMsg, resultDataMsgCh, collection)
		}
	}
}
