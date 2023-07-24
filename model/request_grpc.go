// Package model -----------------------------
// @file      : request_grpc.go
// @author    : 被测试耽误的大厨
// @contact   : 13383088061@163.com
// @time      : 2023/6/27 10:02
// -------------------------------------------
package model

import uuid "github.com/satori/go.uuid"

type RequestGrpc struct {
	TargetId     string               `json:"target_id"`
	Uuid         uuid.UUID            `json:"uuid"`
	Name         string               `json:"name"`
	TeamId       string               `json:"team_id"`
	Debug        string               `json:"debug"`
	Parameter    []*VarForm           `json:"parameter"`
	Header       *Header              `json:"header"` // Headers
	Query        *Query               `json:"query"`
	Body         *Body                `json:"body"`
	Auth         *Auth                `json:"auth"`
	Cookie       *Cookie              `json:"cookie"`
	HttpApiSetup *HttpApiSetup        `json:"http_api_setup"`
	Assert       []*AssertionText     `json:"assert"` // 验证的方法(断言)
	Regex        []*RegularExpression `json:"regex"`  // 正则表达式
}
