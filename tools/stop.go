package tools

import (
	"encoding/json"
	"fmt"
	"github.com/Runner-Go-Team/RunnerGo-engine-open/config"
	"github.com/Runner-Go-Team/RunnerGo-engine-open/log"
	"github.com/Runner-Go-Team/RunnerGo-engine-open/middlewares"
	"io/ioutil"
	"net/http"
	"strings"
)

type StopMsg struct {
	TeamId   string   `json:"team_id"`
	PlanId   string   `json:"plan_id"`
	ReportId string   `json:"report_id"`
	Machines []string `json:"machines"`
}

func SendStopStressReport(machines []string, teamId, planId, reportId string) {
	sm := StopMsg{
		TeamId:   teamId,
		PlanId:   planId,
		ReportId: reportId,
	}
	sm.Machines = machines

	body, err := json.Marshal(&sm)
	if err != nil {
		log.Logger.Error(fmt.Sprintf("机器ip:%s, json转化失败：%s", middlewares.LocalIp, err.Error()))
	}
	res, err := http.Post(config.Conf.Management.NotifyStopStress, "application/json", strings.NewReader(string(body)))

	if err != nil {
		log.Logger.Error(fmt.Sprintf("机器ip:%s, http请求建立链接失败：%s", middlewares.LocalIp, err.Error()))
		return
	}
	defer res.Body.Close()

	responseBody, err := ioutil.ReadAll(res.Body)

	if err != nil {
		log.Logger.Error(fmt.Sprintf("机器ip:%s, 团队：%s， 计划：%s， 报告：%s， 发送停止任务失败，http请求失败: %s", middlewares.LocalIp, teamId, planId, reportId, err.Error()))
		return
	}
	if strings.Contains(string(responseBody), "\"code\":0,") {
		log.Logger.Info(fmt.Sprintf("%s 团队，自动化计划：%s, %s 测试报告停止成功, 请求体：%s,  响应体： %s", teamId, planId, reportId, string(body), string(responseBody)))
	} else {
		log.Logger.Info(fmt.Sprintf("%s 团队，自动化计划：%s, %s 测试报告停止失败, 请求体：%s,  响应体： %s", teamId, planId, reportId, string(body), string(responseBody)))
	}

}

type NotifyRunFinishReq struct {
	TeamID          string `json:"team_id" binding:"required"`
	PlanID          string `json:"plan_id" binding:"required"`
	ReportID        string `json:"report_id" binding:"required"`
	RunDurationTime int64  `json:"run_duration_time"`
}

func SendStopAutoReport(teamId, planId, reportId string, duration int64) {
	sm := NotifyRunFinishReq{
		ReportID:        reportId,
		TeamID:          teamId,
		PlanID:          planId,
		RunDurationTime: duration,
	}

	body, err := json.Marshal(&sm)
	if err != nil {
		log.Logger.Error(fmt.Sprintf("机器ip:%s, json转化失败：%s", middlewares.LocalIp, err.Error()))
	}
	res, err := http.Post(config.Conf.Management.NotifyRunFinish, "application/json", strings.NewReader(string(body)))

	if err != nil {
		log.Logger.Error(fmt.Sprintf("机器ip:%s, http请求建立链接失败：%s", middlewares.LocalIp, err.Error()))
		return
	}
	defer res.Body.Close()

	responseBody, err := ioutil.ReadAll(res.Body)

	if err != nil {
		log.Logger.Error(fmt.Sprintf("机器ip:%s, 团队:%s, 计划: %s, 报告: %s, 发送停止自动化任务失败，http请求失败: %s", middlewares.LocalIp, teamId, planId, reportId, err.Error()))
		return
	}

	if strings.Contains(string(responseBody), "\"code\":0,") {
		log.Logger.Info(fmt.Sprintf("%s 团队，自动化计划：%s, %s 测试报告停止成功, 请求体：%s,  响应体： %s", teamId, planId, reportId, string(body), string(responseBody)))
	} else {
		log.Logger.Info(fmt.Sprintf("%s 团队，自动化计划：%s, %s 测试报告停止失败, 请求体：%s,  响应体： %s", teamId, planId, reportId, string(body), string(responseBody)))
	}

}
