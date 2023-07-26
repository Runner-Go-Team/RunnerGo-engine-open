package model

import (
	uuid "github.com/satori/go.uuid"
)

type Event struct {
	Id                string         `json:"id"`
	PlanId            string         `json:"plan_id"`
	CaseId            string         `json:"case_id"`
	SceneId           string         `json:"scene_id"` // 场景Id
	ParentId          string         `json:"parentId"`
	ReportId          string         `json:"report_id"`
	TeamId            string         `json:"team_id"`
	IsDisabled        int            `json:"is_disabled"` // 0: 不禁用， 1: 禁用
	IsCheck           bool           `json:"is_check"`
	Uuid              uuid.UUID      `json:"uuid"`
	Type              string         `json:"type"` //   事件类型 "request" "controller"
	PreList           []string       `json:"pre_list"`
	NextList          []string       `json:"next_list"`
	Tag               bool           `json:"tag"` // Tps模式下，该标签代表以该接口为准
	Debug             string         `json:"debug"`
	Mode              int64          `json:"mode"`               // 模式类型
	RequestThreshold  int64          `json:"request_threshold"`  // Rps（每秒请求数）阈值
	ResponseThreshold int64          `json:"response_threshold"` // 响应时间阈值
	ErrorThreshold    float32        `json:"error_threshold"`    // 错误率阈值
	PercentAge        int64          `json:"percent_age"`        // 响应时间线
	Weight            int64          `json:"weight"`             // 权重，并发分配的比例
	Prepositions      []*Preposition `json:"prepositions"`       // 前置条件
	Api               Api            `json:"api"`

	Var      string `json:"var"`     // if控制器key，值某个变量
	Compare  string `json:"compare"` // 逻辑运算符
	Val      string `json:"val"`     // key对应的值
	Name     string `json:"name"`    // 控制器名称
	WaitTime int    `json:"wait_ms"` // 等待时长，ms

}

type EventStatus struct {
	EventType string `json:"eventType"`
	EventId   string `json:"eventId"`
	Status    bool   `json:"status"`
}
