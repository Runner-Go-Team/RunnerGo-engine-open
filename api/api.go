package api

import (
	"fmt"
	"github.com/Runner-Go-Team/RunnerGo-engine-open/constant"
	"github.com/Runner-Go-Team/RunnerGo-engine-open/global"
	"github.com/Runner-Go-Team/RunnerGo-engine-open/log"
	"github.com/Runner-Go-Team/RunnerGo-engine-open/middlewares"
	"github.com/Runner-Go-Team/RunnerGo-engine-open/model"
	auto2 "github.com/Runner-Go-Team/RunnerGo-engine-open/model/auto"
	"github.com/Runner-Go-Team/RunnerGo-engine-open/server/client"

	"encoding/json"
	"github.com/Runner-Go-Team/RunnerGo-engine-open/server"
	"github.com/Runner-Go-Team/RunnerGo-engine-open/server/auto"
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
	runApi.Debug = constant.All

	requestJson, _ := json.Marshal(&runApi)

	log.Logger.Info(fmt.Sprintf("机器ip:%s, 调试测试对象", middlewares.LocalIp), string(requestJson))
	_, _ = json.Marshal(runApi.Request.Body.Mode)
	go server.DebugApi(runApi)
	global.ReturnMsg(c, http.StatusOK, "调试接口", uid)
}

func RunMysqlConnection(c *gin.Context) {

	var connection = model.SqlDatabaseInfo{}
	err := c.ShouldBindJSON(&connection)
	if err != nil {
		global.ReturnMsg(c, http.StatusBadRequest, "数据格式不正确", err.Error())
		return
	}
	connJson, _ := json.Marshal(&connection)
	log.Logger.Info(fmt.Sprintf("机器ip:%s, 数据库测试链接：    ", middlewares.LocalIp), string(connJson))
	db, err := client.TestConnection(connection)
	if db != nil {
		defer db.Close()
	}
	if err != nil {
		global.ReturnMsg(c, http.StatusBadRequest, "mysql链接数据不正确", err.Error())
		return
	}

	global.ReturnMsg(c, http.StatusOK, "测试链接成功", "success")
}
