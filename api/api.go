package api

import (
	"RunnerGo-engine/global"
	"RunnerGo-engine/log"
	"RunnerGo-engine/middlewares"
	"RunnerGo-engine/model"
	auto2 "RunnerGo-engine/model/auto"
	"fmt"

	"RunnerGo-engine/server"
	"RunnerGo-engine/server/auto"
	"encoding/json"
	"github.com/gin-gonic/gin"
	uuid "github.com/satori/go.uuid"
	"net/http"
)

func RunPlan(c *gin.Context) {
	var planInstance = model.Plan{}
	err := c.ShouldBindJSON(&planInstance)

	if err != nil {
		global.ReturnMsg(c, http.StatusBadRequest, "数据格式不正确", err.Error())
		return
	}

	requestJson, err := json.Marshal(planInstance)
	if err != nil {
		log.Logger.Info(fmt.Sprintf("机器ip:%s, 测试计划，结构体转json失败：   ", middlewares.LocalIp), err)
		global.ReturnMsg(c, http.StatusBadRequest, "数据格式不正确", err.Error())
		return
	}

	log.Logger.Info(fmt.Sprintf("机器ip:%s, 开始执行计划", middlewares.LocalIp), string(requestJson))

	server.DisposeTask(&planInstance, c)
	return

}

func RunAutoPlan(c *gin.Context) {
	var autoPlan = auto2.Plan{}
	err := c.ShouldBindJSON(&autoPlan)
	if err != nil {
		global.ReturnMsg(c, http.StatusBadRequest, "数据格式不正确", err.Error())
		return
	}

	requestJson, err := json.Marshal(autoPlan)
	if err != nil {
		log.Logger.Info(fmt.Sprintf("机器ip:%s, 测试计划，结构体转json失败：   ", middlewares.LocalIp), err)
		global.ReturnMsg(c, http.StatusBadRequest, "数据格式不正确", err.Error())
		return
	}
	log.Logger.Info(fmt.Sprintf("机器ip:%s, 开始执行自动化计划! ", middlewares.LocalIp), string(requestJson))
	auto.DisposeAutoPlan(&autoPlan, c)
	return
}

func RunScene(c *gin.Context) {
	var scene model.Scene
	err := c.ShouldBindJSON(&scene)
	if err != nil {
		global.ReturnMsg(c, http.StatusBadRequest, "数据格式不正确", err.Error())
		return
	}

	uid := uuid.NewV4()
	scene.Uuid = uid
	requestJson, _ := json.Marshal(scene)
	log.Logger.Info(fmt.Sprintf("机器ip:%s, 调试场景", middlewares.LocalIp), string(requestJson))
	go server.DebugScene(scene)
	global.ReturnMsg(c, http.StatusOK, "调式场景", uid)
}

func RunApi(c *gin.Context) {
	var runApi = model.Api{}
	err := c.ShouldBindJSON(&runApi)

	if err != nil {
		global.ReturnMsg(c, http.StatusBadRequest, "数据格式不正确", err.Error())
		return
	}

	uid := uuid.NewV4()
	runApi.Uuid = uid
	runApi.Debug = model.All

	requestJson, _ := json.Marshal(&runApi)

	log.Logger.Info(fmt.Sprintf("机器ip:%s, 调试接口", middlewares.LocalIp), string(requestJson))
	_, _ = json.Marshal(runApi.Request.Body.Mode)
	go server.DebugApi(runApi)
	global.ReturnMsg(c, http.StatusOK, "调试接口", uid)
}

//func Stop(c *gin.Context) {
//	var stop model.Stop
//
//	if err := c.ShouldBindJSON(&stop); err != nil {
//		global.ReturnMsg(c, http.StatusBadRequest, "数据格式不正确", err.Error())
//		return
//	}
//	if stop.ReportIds == nil || len(stop.ReportIds) < 1 {
//		global.ReturnMsg(c, http.StatusBadRequest, "数据格式不正确", "报告列表不能为空")
//		return
//	}
//	go func(stop model.Stop) {
//		for _, reportId := range stop.ReportIds {
//			key := fmt.Sprintf("%d:%d:%s:status", stop.TeamId, stop.PlanId, reportId)
//			log.Logger.Info("停止任务：   ", key)
//			err := model.InsertStatus(key, "stop", -1)
//			if err != nil {
//				log.Logger.Error("向redis写入任务状态失败：", err)
//			}
//		}
//	}(stop)
//	global.ReturnMsg(c, http.StatusOK, "停止任务", nil)
//}

//func StopScene(c *gin.Context) {
//	var stop model.StopScene
//	if err := c.ShouldBindJSON(&stop); err != nil {
//		global.ReturnMsg(c, http.StatusBadRequest, "数据格式不正确", err.Error())
//		return
//	}
//
//	if stop.SceneId == 0 {
//		global.ReturnMsg(c, http.StatusBadRequest, "scene_id不正确", stop.SceneId)
//		return
//	}
//	go func(stop model.StopScene) {
//		err := model.InsertStatus(fmt.Sprintf("StopScene:%d:%d", stop.TeamId, stop.SceneId), "stop", 20)
//		if err != nil {
//			log.Logger.Error("向redis写入任务状态失败：", err)
//		}
//	}(stop)
//
//	global.ReturnMsg(c, http.StatusOK, "停止成功", nil)
//}
