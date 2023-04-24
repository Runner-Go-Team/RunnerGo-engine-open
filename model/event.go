package model

import (
	uuid "github.com/satori/go.uuid"
)

type Event struct {
	Id                string    `json:"id" bson:"id"`
	PlanId            string    `json:"plan_id" bson:"plan_id"`
	CaseId            string    `json:"case_id" bson:"case_id"`
	SceneId           string    `json:"scene_id" bson:"scene_id"` // 场景Id
	ParentId          string    `json:"parentId" bson:"parent_id"`
	ReportId          string    `json:"report_id" bson:"report_id"`
	TeamId            string    `json:"team_id" bson:"team_id"`
	IsCheck           bool      `json:"is_check" bson:"is_check"`
	Uuid              uuid.UUID `json:"uuid" bson:"uuid"`
	Type              string    `json:"type" bson:"type"` //   事件类型 "request" "controller"
	PreList           []string  `json:"pre_list" bson:"pre_list"`
	NextList          []string  `json:"next_list"   bson:"next_list"`
	Tag               bool      `json:"tag" bson:"tag"` // Tps模式下，该标签代表以该接口为准
	Debug             string    `json:"debug" bson:"debug"`
	Mode              int64     `json:"mode"`                 // 模式类型
	RequestThreshold  int64     `json:"request_threshold"`    // Rps（每秒请求数）阈值
	ResponseThreshold int64     `json:"response_threshold"`   // 响应时间阈值
	ErrorThreshold    float32   `json:"error_threshold"`      // 错误率阈值
	PercentAge        int64     `json:"percent_age"`          // 响应时间线
	Weight            int64     `json:"weight" bson:"weight"` // 权重，并发分配的比例
	Api               Api       `json:"api"`
	Var               string    `json:"var"`     // if控制器key，值某个变量
	Compare           string    `json:"compare"` // 逻辑运算符
	Val               string    `json:"val"`     // key对应的值
	Name              string    `json:"name"`    // 控制器名称
	WaitTime          int       `json:"wait_ms"` // 等待时长，ms

}

type EventStatus struct {
	EventType string `json:"eventType" bson:"eventType"`
	EventId   string `json:"eventId" bson:"eventId"`
	Status    bool   `json:"status" bson:"status"`
}

func (api *Api) GlobalToRequest() {
	if api.ApiVariable.Cookie != nil && len(api.ApiVariable.Cookie.Parameter) > 0 {
		if api.Request.Cookie == nil {
			api.Request.Cookie = new(Cookie)
		}
		if api.Request.Cookie.Parameter == nil {
			api.Request.Cookie.Parameter = []*VarForm{}
		}
		for _, parameter := range api.ApiVariable.Cookie.Parameter {
			if parameter.IsChecked != Open {
				continue
			}
			var isExist bool
			for _, value := range api.Request.Cookie.Parameter {
				if value.IsChecked == Open && parameter.Key == value.Key && parameter.Value == value.Value {
					isExist = true
				}
			}
			if isExist {
				continue
			}
			api.Request.Cookie.Parameter = append(api.Request.Cookie.Parameter, parameter)
		}
	}
	if api.ApiVariable.Header != nil && len(api.ApiVariable.Header.Parameter) > 0 {
		if api.Request.Header == nil {
			api.Request.Header = new(Header)
		}
		if api.Request.Header.Parameter == nil {
			api.Request.Header.Parameter = []*VarForm{}
		}
		for _, parameter := range api.ApiVariable.Header.Parameter {
			if parameter.IsChecked != Open {
				continue
			}
			var isExist bool
			for _, value := range api.Request.Header.Parameter {
				if value.IsChecked == Open && parameter.Key == value.Key && parameter.Value == parameter.Value {
					isExist = true
				}
			}
			if isExist {
				continue
			}
			api.Request.Header.Parameter = append(api.Request.Header.Parameter, parameter)

		}
	}

	if api.ApiVariable.Assert != nil && len(api.ApiVariable.Assert) > 0 {
		if api.Assert == nil {
			api.Assert = []*AssertionText{}
		}
		for _, parameter := range api.ApiVariable.Assert {
			if parameter.IsChecked != Open {
				continue
			}
			var isExist bool
			for _, asser := range api.Assert {
				if asser.IsChecked == Open && parameter.ResponseType == asser.ResponseType && parameter.Compare == asser.Compare && parameter.Val == asser.Val && parameter.Var == asser.Var {
					isExist = true
				}
			}
			if isExist {
				continue
			}
			api.Assert = append(api.Assert, parameter)

		}
	}

}
