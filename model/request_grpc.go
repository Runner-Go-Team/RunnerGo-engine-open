// Package model -----------------------------
// @file      : request_grpc.go
// @author    : 被测试耽误的大厨
// @contact   : 13383088061@163.com
// @time      : 2023/6/27 10:02
// -------------------------------------------
package model

type RequestGrpc struct {
	FilePath     string               `json:"file_path"`
	ServiceName  string               `json:"service"`
	Method       string               `json:"method"` // 方法 GET/POST/PUT
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

func initGrpc() {
	//var parser protoparse.Parser
	//
	//fileDescriptors, _ := parser.ParseFiles("./")
}
