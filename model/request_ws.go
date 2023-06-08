package model

import uuid "github.com/satori/go.uuid"

type WebsocketDetail struct {
	TargetId       string          `json:"target_id"`
	Uuid           uuid.UUID       `json:"uuid"`
	Name           string          `json:"name"`
	TeamId         string          `json:"team_id"`
	Url            string          `json:"url"`
	Debug          string          `json:"debug"`
	SendMessage    string          `json:"send_message"`
	WsHeader       []WsQuery       `json:"ws_header"`
	WsParam        []WsQuery       `json:"ws_param"`
	WsEvent        []WsQuery       `json:"ws_event"`
	WsConfig       WsConfig        `json:"ws_config"`
	Configuration  *Configuration  `json:"configuration"`   // 场景设置
	GlobalVariable *GlobalVariable `json:"global_variable"` // 全局变量
	WsVariable     *GlobalVariable `json:"api_variable"`
}

type WsConfig struct {
	ConnectType         int `json:"connect_type"`           // 连接类型：1-长连接，2-短连接
	ConnectDurationTime int `json:"connect_duration_time"`  // 连接持续时长，单位：秒
	SendMsgDurationTime int `json:"send_msg_duration_time"` // 发送消息间隔时长，单位：毫秒
	ConnectTimeoutTime  int `json:"connect_timeout_time"`   // 连接超时时间，单位：毫秒
	RetryNum            int `json:"retry_num"`              // 重连次数
	RetryInterval       int `json:"retry_interval"`         // 重连间隔时间，单位：毫秒
}

type WsQuery struct {
	IsChecked int32  `json:"is_checked"`
	Var       string `json:"var"`
	Val       string `json:"val"`
}
