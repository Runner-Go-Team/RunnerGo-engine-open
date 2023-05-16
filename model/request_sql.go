package model

import uuid "github.com/satori/go.uuid"

type SQL struct {
	TargetId          string               `json:"target_id"`
	Uuid              uuid.UUID            `json:"uuid"`
	Name              string               `json:"name"`
	TeamId            string               `json:"team_id"`
	TargetType        string               `json:"target_type"` // api/webSocket/tcp/grpc
	SqlString         string               `json:"sql_string"`
	MysqlDatabaseInfo MysqlDatabaseInfo    `json:"mysql_database_info"`
	Assert            []*AssertionText     `json:"assert"`  // 验证的方法(断言)
	Timeout           int64                `json:"timeout"` // 请求超时时间
	Regex             []*RegularExpression `json:"regex"`   // 正则表达式
	Debug             string               `json:"debug"`   // 是否开启Debug模式
	Configuration     *Configuration       `json:"configuration"`
	SqlVariable       *GlobalVariable      `json:"sql_variable"`    // 全局变量
	GlobalVariable    *GlobalVariable      `json:"global_variable"` // 全局变量
}

type MysqlDatabaseInfo struct {
	Type     string `json:"type"`
	Host     string `json:"host"`
	User     string `json:"user"`
	Password string `json:"password"`
	Port     int32  `json:"port"`
	DbName   string `json:"db_name"`
	Charset  string `json:"charset"`
}
