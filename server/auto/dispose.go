package auto

import (
	"context"
	"fmt"
	"github.com/Runner-Go-Team/RunnerGo-engine-open/config"
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
	"strings"
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

	if plan.Variable != nil {
		for _, value := range plan.Variable {
			values := tools.FindAllDestStr(value.Val, "{{(.*?)}}")
			if values == nil {
				continue
			}

			for _, v := range values {
				if len(v) < 2 {
					continue
				}
				realVar := tools.ParsFunc(v[1])
				if realVar != v[1] {
					value.Val = strings.Replace(value.Val, v[0], realVar, -1)
				}
			}
		}
	}

	// 循环读全局变量，如果场景变量不存在则将添加到场景变量中，如果有参数化数据则，将其替换

	for _, scene := range plan.Scenes {
		if scene.Configuration == nil {
			scene.Configuration = new(model.Configuration)
		}
		if scene.Configuration.Variable == nil {
			scene.Configuration.Variable = []*model.KV{}
		}
		if plan.Variable != nil {
			for _, value := range plan.Variable {
				var target = false
				for _, kv := range scene.Configuration.Variable {
					if value.Var == kv.Key {
						target = true
						continue
					}
				}
				if !target {
					var variable = new(model.KV)
					variable.Key = value.Var
					variable.Value = value.Val
					scene.Configuration.Variable = append(scene.Configuration.Variable, variable)
				}
			}
		}
		if scene.Configuration.ParameterizedFile == nil {
			scene.Configuration.ParameterizedFile = new(model.ParameterizedFile)
		}
		if scene.Configuration.ParameterizedFile.VariableNames == nil {
			scene.Configuration.ParameterizedFile.VariableNames = new(model.VariableNames)
		}
		if scene.Configuration.ParameterizedFile.VariableNames.VarMapList == nil {
			scene.Configuration.ParameterizedFile.VariableNames.VarMapList = make(map[string][]string)
		}
		if scene.Configuration.ParameterizedFile != nil {
			p := scene.Configuration.ParameterizedFile
			p.VariableNames.Mu = sync.Mutex{}
			p.UseFile()
		}
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

	// 设置接收数据缓存
	//resultDataMsgCh := make(chan *model.ResultDataMsg, 10000)

	var wg = &sync.WaitGroup{}
	collection := model.NewCollection(config.Conf.Mongo.DataBase, config.Conf.Mongo.AutoTable, mongoClient)

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
	case model.AuToOrderMode:
		for _, scene := range plan.Scenes {
			key := fmt.Sprintf("StopAutoPlan:%s:%s:%d", scene.TeamId, scene.PlanId, scene.ReportId)
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
	case model.AuToSameMode:
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
	log.Logger.Info(fmt.Sprintf("机器ip:%s, 团队: %s, 自动化计划: %s, 报告: %s  已完成, 运行：%d", middlewares.LocalIp, plan.TeamId, plan.PlanId, plan.ReportId, time.Now().UnixMilli()-startTime))

}

func disposeCase(scene model.Scene, sceneRunMode, caseMode int64, wg *sync.WaitGroup, configuration *model.Configuration, reportMsg *model.ResultDataMsg, resultDataMsgCh chan *model.ResultDataMsg, collection *mongo.Collection) {
	if sceneRunMode == model.AuToSameMode {
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
		if c.IsChecked != model.Open {
			continue
		}
		c.PlanId = scene.PlanId
		c.TeamId = scene.TeamId
		c.ReportId = scene.ReportId
		c.ParentId = scene.SceneId
		c.Debug = model.All
		if c.Configuration == nil {
			c.Configuration = new(model.Configuration)
		}
		if c.Configuration.Variable == nil {
			c.Configuration.Variable = []*model.KV{}
		}
		var currentWg = new(sync.WaitGroup)
		switch caseMode {
		case model.AuToOrderMode:
			var sceneWg = &sync.WaitGroup{}
			golink.DisposeScene(wg, currentWg, sceneWg, model.SceneType, c, configuration, reportMsg, resultDataMsgCh, collection)
			sceneWg.Wait()
		}
		//wg.Wait()
		currentWg.Wait()
	}
}
