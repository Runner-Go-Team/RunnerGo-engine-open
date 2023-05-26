package model

import uuid "github.com/satori/go.uuid"

type TCP struct {
	TargetId       string          `json:"target_id"`
	Uuid           uuid.UUID       `json:"uuid"`
	Name           string          `json:"name"`
	TeamId         string          `json:"team_id"`
	TargetType     string          `json:"target_type"` // api/webSocket/tcp/grpc
	Timeout        int64           `json:"timeout"`     // 请求超时时间
	Debug          string          `json:"debug"`       // 是否开启Debug模式
	Url            string          `json:"url"`
	SendMessage    string          `bson:"send_message"`
	TcpConfig      TcpConfig       `json:"tcp_config"`
	Configuration  *Configuration  `json:"configuration"`
	SqlVariable    *GlobalVariable `json:"sql_variable"`    // 全局变量
	GlobalVariable *GlobalVariable `json:"global_variable"` // 全局变量
}

type TcpConfig struct {
	ConnectType         int `bson:"connect_type"`           // 连接类型：1-长连接，2-短连接
	ConnectTimeoutTime  int `bson:"connect_timeout_time"`   // 连接超时时间，单位：毫秒
	RetryNum            int `bson:"retry_num"`              // 重连次数
	RetryInterval       int `bson:"retry_interval"`         // 重连间隔时间，单位：毫秒
	ConnectDurationTime int `json:"connect_duration_time"`  // 连接持续时长
	SendMsgDurationTime int `json:"send_msg_duration_time"` // 发送消息间隔时长
}
