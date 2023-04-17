// Package golink 连接
package golink

import (
	"github.com/Runner-Go-Team/RunnerGo-engine-open/middlewares"
	"github.com/Runner-Go-Team/RunnerGo-engine-open/model"
	"github.com/Runner-Go-Team/RunnerGo-engine-open/server/client"
	"github.com/valyala/fasthttp"
	"go.mongodb.org/mongo-driver/mongo"
	"net/url"
	"sync"
	"time"
)

// HttpSend 发送http请求
func HttpSend(event model.Event, api model.Api, globalVar *sync.Map, requestCollection *mongo.Collection) (bool, int64, uint64, float64, float64, string, time.Time, time.Time) {
	var (
		isSucceed       = true
		errCode         = model.NoError
		receivedBytes   = float64(0)
		errMsg          = ""
		assertNum       = 0
		assertFailedNum = 0
	)

	if api.HttpApiSetup == nil {
		api.HttpApiSetup = new(model.HttpApiSetup)
	}
	resp, req, requestTime, sendBytes, err, str, startTime, endTime := client.HTTPRequest(api.Method, api.Request.URL, api.Request.Body, api.Request.Query,
		api.Request.Header, api.Request.Cookie, api.Request.Auth, api.HttpApiSetup)
	defer fasthttp.ReleaseResponse(resp) // 用完需要释放资源
	defer fasthttp.ReleaseRequest(req)
	var regex []map[string]interface{}
	if api.Regex != nil {
		for _, regular := range api.Regex {
			if regular.IsChecked != model.Open {
				continue
			}
			reg := make(map[string]interface{})
			value := regular.Extract(resp, globalVar)
			if value == nil {
				continue
			}
			reg[regular.Var] = value
			regex = append(regex, reg)
		}
	}
	//log.Logger.Debug("api:::::     ", api.Name, "\nbody:      ", string(resp.Body()), "\nREGEX:    ", regex)
	if err != nil {
		isSucceed = false
		errMsg = err.Error()
	}
	if resp.StatusCode() != 200 {
		isSucceed = false
		errMsg = string(resp.Body())
	}
	var assertionMsgList []model.AssertionMsg
	// 断言验证

	if api.Assert != nil {
		var assertionMsg = model.AssertionMsg{}
		var (
			code    = int64(10000)
			succeed = true
			msg     = ""
		)
		for _, v := range api.Assert {
			if v.IsChecked != model.Open {
				continue
			}
			code, succeed, msg = v.VerifyAssertionText(resp)
			if succeed != true {
				errCode = code
				isSucceed = succeed
				errMsg = msg
				assertFailedNum++
			}
			assertionMsg.Code = code
			assertionMsg.IsSucceed = succeed
			assertionMsg.Msg = msg
			assertionMsgList = append(assertionMsgList, assertionMsg)
			assertNum++
		}
	}
	// 接收到的字节长度
	//contentLength = uint(resp.Header.ContentLength())

	receivedBytes = float64(resp.Header.ContentLength()) / 1024
	if receivedBytes <= 0 {
		receivedBytes = float64(len(resp.Body())) / 1024
	}
	// 开启debug模式后，将请求响应信息写入到mongodb中
	if api.Debug != "" && api.Debug != "stop" {
		var debugMsg = make(map[string]interface{})
		responseTime := endTime.Format("2006-01-02 15:04:05")
		insertDebugMsg(regex, debugMsg, api.Debug, event, api, resp, req, requestTime, responseTime, receivedBytes, errMsg, str, err, isSucceed, assertionMsgList, assertNum, assertFailedNum)
		if requestCollection != nil {
			model.Insert(requestCollection, debugMsg, middlewares.LocalIp)
		}
	}
	return isSucceed, errCode, requestTime, sendBytes, receivedBytes, errMsg, startTime, endTime
}

func insertDebugMsg(regex []map[string]interface{}, debugMsg map[string]interface{}, debugType string, event model.Event, api model.Api, resp *fasthttp.Response, req *fasthttp.Request, requestTime uint64, responseTime string, receivedBytes float64, errMsg, str string, err error, isSucceed bool, assertionMsgList []model.AssertionMsg, assertNum, assertFailedNum int) {
	switch debugType {
	case model.All:
		makeDebugMsg(regex, debugMsg, event, api, resp, req, requestTime, responseTime, receivedBytes, errMsg, str, err, isSucceed, assertionMsgList, assertNum, assertFailedNum)
	case model.OnlySuccess:
		if isSucceed == true {
			makeDebugMsg(regex, debugMsg, event, api, resp, req, requestTime, responseTime, receivedBytes, errMsg, str, err, isSucceed, assertionMsgList, assertNum, assertFailedNum)
		}

	case model.OnlyError:
		if isSucceed == false {
			makeDebugMsg(regex, debugMsg, event, api, resp, req, requestTime, responseTime, receivedBytes, errMsg, str, err, isSucceed, assertionMsgList, assertNum, assertFailedNum)
		}
	}
}

func makeDebugMsg(regex []map[string]interface{}, debugMsg map[string]interface{}, event model.Event, api model.Api, resp *fasthttp.Response, req *fasthttp.Request, requestTime uint64, responseTime string, receivedBytes float64, errMsg, str string, err error, isSucceed bool, assertionMsgList []model.AssertionMsg, assertNum, assertFailedNum int) {
	debugMsg["team_id"] = event.TeamId
	debugMsg["request_url"] = req.URI().String()
	debugMsg["plan_id"] = event.PlanId
	debugMsg["report_id"] = event.ReportId
	debugMsg["scene_id"] = event.SceneId
	debugMsg["parent_id"] = event.ParentId
	debugMsg["case_id"] = event.CaseId
	debugMsg["uuid"] = api.Uuid.String()
	debugMsg["event_id"] = event.Id
	debugMsg["api_id"] = api.TargetId
	debugMsg["api_name"] = api.Name
	debugMsg["type"] = model.RequestType
	debugMsg["request_time"] = requestTime / uint64(time.Millisecond)
	debugMsg["request_code"] = resp.StatusCode()
	debugMsg["request_header"] = req.Header.String()
	debugMsg["response_time"] = responseTime
	if string(req.Body()) != "" {
		var errBody error
		debugMsg["request_body"], errBody = url.QueryUnescape(string(req.Body()))
		if errBody != nil {
			debugMsg["request_body"] = string(req.Body())
		}
	} else {
		debugMsg["request_body"] = str
	}
	if string(resp.Body()) == "" && errMsg != "" {
		debugMsg["response_body"] = errMsg
	}

	debugMsg["response_header"] = resp.Header.String()

	debugMsg["response_bytes"] = receivedBytes
	if err != nil {
		debugMsg["response_body"] = err.Error()
	} else {
		debugMsg["response_body"] = string(resp.Body())
	}
	switch isSucceed {
	case false:
		debugMsg["status"] = model.Failed
	case true:
		debugMsg["status"] = model.Success
	}

	debugMsg["next_list"] = event.NextList
	debugMsg["assertion"] = assertionMsgList
	debugMsg["assertion_num"] = assertNum
	debugMsg["assertion_failed_num"] = assertFailedNum
	debugMsg["regex"] = regex
}
