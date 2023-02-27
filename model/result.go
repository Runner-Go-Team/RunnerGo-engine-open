package model

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/Runner-Go-Team/RunnerGo-engine-open/log"
	"github.com/Runner-Go-Team/RunnerGo-engine-open/middlewares"
)

/*
 测试结果
*/

// SceneTestResultDataMsg 场景的测试结果

type SceneTestResultDataMsg struct {
	TeamId      string                                `json:"team_id" bson:"team_id"`
	MachineIp   string                                `json:"machine_ip" bson:"machine_ip"`
	MachineName string                                `json:"machine_name" bson:"machine_name"`
	ReportId    string                                `json:"report_id" bson:"report_id"`
	ReportName  string                                `json:"report_name" bson:"report_name"`
	PlanId      string                                `json:"plan_id" bson:"plan_id"`     // 任务ID
	PlanName    string                                `json:"plan_name" bson:"plan_name"` //
	SceneId     string                                `json:"scene_id" bson:"scene_id"`   // 场景
	SceneName   string                                `json:"scene_name" bson:"scene_name"`
	Results     map[interface{}]*ApiTestResultDataMsg `json:"results" bson:"results"`
}

// ApiTestResultDataMsg 接口测试数据经过计算后的测试结果
type ApiTestResultDataMsg struct {
	EventId                        string  `json:"event_id"`
	Name                           string  `json:"name"`
	PlanId                         string  `json:"plan_id"`   // 任务ID
	PlanName                       string  `json:"plan_name"` //
	SceneId                        string  `json:"scene_id"`  // 场景
	SceneName                      string  `json:"scene_name"`
	Concurrency                    int64   `json:"concurrency"`
	TotalRequestNum                int64   `json:"total_request_num"`  // 总请求数
	TotalRequestTime               int64   `json:"total_request_time"` // 总响应时间
	SuccessNum                     int64   `json:"success_num"`
	ErrorRate                      float64 `json:"error_rate"`
	ErrorNum                       int64   `json:"error_num"`          // 错误数
	PercentAge                     int64   `json:"percent_age"`        // 响应时间线
	ErrorThreshold                 float64 `json:"error_threshold"`    // 自定义错误率
	RequestThreshold               int64   `json:"request_threshold"`  // Rps（每秒请求数）阈值
	ResponseThreshold              int64   `json:"response_threshold"` // 响应时间阈值
	AvgRequestTime                 float64 `json:"avg_request_time"`   // 平均响应事件
	MaxRequestTime                 float64 `json:"max_request_time"`
	MinRequestTime                 float64 `json:"min_request_time"` // 毫秒
	CustomRequestTimeLine          int64   `json:"custom_request_time_line"`
	FiftyRequestTimeline           int64   `json:"fifty_request_time_line" `
	NinetyRequestTimeLine          int64   `json:"ninety_request_time_line"`
	NinetyFiveRequestTimeLine      int64   `json:"ninety_five_request_time_line"`
	NinetyNineRequestTimeLine      int64   `json:"ninety_nine_request_time_line"`
	FiftyRequestTimelineValue      float64 `json:"fifty_request_time_line_value"`
	CustomRequestTimeLineValue     float64 `json:"custom_request_time_line_value"`
	NinetyRequestTimeLineValue     float64 `json:"ninety_request_time_line_value"`
	NinetyFiveRequestTimeLineValue float64 `json:"ninety_five_request_time_line_value"`
	NinetyNineRequestTimeLineValue float64 `json:"ninety_nine_request_time_line_value"`
	SendBytes                      float64 `json:"send_bytes"`     // 发送字节数
	ReceivedBytes                  float64 `json:"received_bytes"` // 接收字节数
	Rps                            float64 `json:"rps"`
}

// ResultDataMsg 请求结果数据结构
type ResultDataMsg struct {
	Start             bool    `json:"start" bson:"start"`
	End               bool    `json:"end" bson:"end"` //最后一条记录标记
	TeamId            string  `json:"team_id" bson:"team_id"`
	Concurrency       int64   `json:"concurrency" bson:"concurrency"`
	MachineNum        int64   `json:"machine_num" bson:"machine_num"` // 机器数量
	MachineIp         string  `json:"machine_ip" bson:"machine_ip"`
	ReportId          string  `json:"report_id" bson:"report_id"`
	ReportName        string  `json:"report_name" bson:"report_name"`
	EventId           string  `json:"event_id" bson:"event_id"`
	PlanId            string  `json:"plan_id" bson:"plan_id"`     // 任务ID
	PlanName          string  `json:"plan_name" bson:"plan_name"` //
	SceneId           string  `json:"scene_id" bson:"scene_id"`   // 场景
	SceneName         string  `json:"sceneName" bson:"scene_name"`
	TargetId          string  `json:"target_id" bson:"target_id"`             // 接口ID
	Name              string  `json:"name" bson:"name"`                       // 接口名称
	RequestTime       uint64  `json:"request_time" bson:"request_time"`       // 请求响应时间
	ResponseThreshold int64   `json:"response_threshold"`                     // 响应时间阈值
	PercentAge        int64   `json:"percent_age" bson:"percent_age"`         // 响应时间线
	ErrorThreshold    float32 `json:"error_threshold" bson:"error_threshold"` // 自定义错误率
	RequestThreshold  int64   `json:"request_threshold"`                      // Rps（每秒请求数）阈值
	ErrorType         int64   `json:"error_type" bson:"error_type"`           // 错误类型：1. 请求错误；2. 断言错误
	IsSucceed         bool    `json:"is_succeed" bson:"is_succeed"`           // 请求是否有错：true / false   为了计数
	ErrorMsg          string  `json:"error_msg" bson:"error_msg"`             // 错误信息
	SendBytes         float64 `json:"send_bytes" bson:"send_bytes"`           // 发送字节数
	ReceivedBytes     float64 `json:"received_bytes" bson:"received_bytes"`   // 接收字节数
	Timestamp         int64   `json:"timestamp" bson:"timestamp"`
	StartTime         int64   `json:"start_time" bson:"start_time"`
	EndTime           int64   `json:"end_time" bson:"end_time"`
}

func (tr *ApiTestResultDataMsg) Encode() ([]byte, error) {
	buf := new(bytes.Buffer)
	if err := binary.Write(buf, binary.BigEndian, tr); err != nil {
		log.Logger.Error(fmt.Sprintf("机器ip:%s, 测试数据转字节失败: %s", middlewares.LocalIp, err.Error()))
		return nil, err
	}
	return buf.Bytes(), nil
}

func (tr *ApiTestResultDataMsg) Length() int {
	by, _ := tr.Encode()
	return len(by)
}

// EsApiTestResultDataMsg Es接口测试数据经过计算后的测试结果
type EsApiTestResultDataMsg struct {
	EventId                    string  `json:"event_id" bson:"event_id"`
	Name                       string  `json:"name" bson:"name"`
	PlanId                     string  `json:"plan_id" bson:"plan_id"`     // 任务ID
	PlanName                   string  `json:"plan_name" bson:"plan_name"` //
	SceneId                    string  `json:"scene_id" bson:"scene_id"`   // 场景
	SceneName                  string  `json:"scene_name" bson:"scene_name"`
	TotalRequestNum            uint64  `json:"total_request_num" bson:"total_request_num"`   // 总请求数
	TotalRequestTime           uint64  `json:"total_request_time" bson:"total_request_time"` // 总响应时间
	SuccessNum                 uint64  `json:"success_num" bson:"success_num"`
	ErrorNum                   uint64  `json:"error_num" bson:"error_num"`               // 错误数
	AvgRequestTime             uint64  `json:"avg_request_time" bson:"avg_request_time"` // 平均响应事件
	MaxRequestTime             uint64  `json:"max_request_time" bson:"max_request_time"`
	MinRequestTime             uint64  `json:"min_request_time" bson:"min_request_time"` // 毫秒
	CustomRequestTimeLine      int64   `json:"custom_request_time_line" bson:"custom_request_time_line"`
	CustomRequestTimeLineValue uint64  `json:"custom_request_time_line_value" bson:"custom_request_time_line_value"`
	NinetyRequestTimeLine      uint64  `json:"ninety_request_time_line" bson:"ninety_request_time_line"`
	NinetyFiveRequestTimeLine  uint64  `json:"ninety_five_request_time_line" bson:"ninety_five_request_time_line"`
	NinetyNineRequestTimeLine  uint64  `json:"ninety_nine_request_time_line" bson:"ninety_nine_request_time_line"`
	SendBytes                  uint64  `json:"send_bytes" bson:"send_bytes"`         // 发送字节数
	ReceivedBytes              uint64  `json:"received_bytes" bson:"received_bytes"` // 接收字节数
	Rps                        float64 `json:"rps" bson:"rps"`
}

// EsSceneTestResultDataMsg Es场景的测试结果

type RedisSceneTestResultDataMsg struct {
	End        bool                             `json:"end" bson:"end"`
	TeamId     string                           `json:"team_id" bson:"team_id"`
	ReportId   string                           `json:"report_id" bson:"report_id"`
	ReportName string                           `json:"report_name" bson:"report_name"`
	PlanId     string                           `json:"plan_id" bson:"plan_id"`     // 任务ID
	PlanName   string                           `json:"plan_name" bson:"plan_name"` //
	SceneId    string                           `json:"scene_id" bson:"scene_id"`   // 场景
	SceneName  string                           `json:"scene_name" bson:"scene_name"`
	Results    map[string]*ApiTestResultDataMsg `json:"results" bson:"results"`
	Machine    map[string]int64                 `json:"machine" bson:"machine"`
	TimeStamp  int64                            `json:"time_stamp" bson:"time_stamp"`
}
