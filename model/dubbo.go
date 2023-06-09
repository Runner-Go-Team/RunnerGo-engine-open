package model

import uuid "github.com/satori/go.uuid"

type DubboDetail struct {
	TargetId string    `json:"target_id"`
	Uuid     uuid.UUID `json:"uuid"`
	Name     string    `json:"name"`
	TeamId   string    `json:"team_id"`
	Debug    string    `json:"debug"`

	DubboProtocol string `json:"dubbo_method"`
	ApiName       string `json:"api_name"`
	FunctionName  string `json:"dubbo_iface"`
	Version       string `json:"version"`

	DubboParam     []DubboParam    `json:"dubbo_param"`
	DubboAssert    []DubboAssert   `json:"dubbo_assert"`
	DubboRegex     []DubboRegex    `json:"dubbo_regex"`
	DubboConfig    DubboConfig     `json:"dubbo_config"`
	Configuration  *Configuration  `json:"configuration"`   // 场景设置
	GlobalVariable *GlobalVariable `json:"global_variable"` // 全局变量
	DubboVariable  *GlobalVariable `json:"dubbo_variable"`
}

type DubboConfig struct {
	RegistrationCenterName    string `json:"registration_center_name"`
	RegistrationCenterAddress string `json:"registration_center_address"`
}

type DubboAssert struct {
	IsChecked    int    `json:"is_checked"`
	ResponseType int32  `json:"response_type"`
	Var          string `json:"var"`
	Compare      string `json:"compare"`
	Val          string `json:"val"`
	Index        int    `json:"index"` // 正则时提取第几个值
}

type DubboRegex struct {
	IsChecked int    `json:"is_checked"` // 1 选中, -1未选
	Type      int    `json:"type"`       // 0 正则  1 json
	Var       string `json:"var"`
	Express   string `json:"express"`
	Val       string `json:"val"`
	Index     int    `json:"index"` // 正则时提取第几个值
}

type DubboParam struct {
	IsChecked int32  `json:"is_checked"`
	ParamType string `json:"param_type"`
	Var       string `json:"var"`
	Val       string `json:"val"`
}
