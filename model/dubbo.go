package model

type DubboRequest struct {
	PreUrl          string      `json:"pre_url"`
	URL             string      `json:"url"`
	AppName         string      `json:"app_name"`
	Method          string      `json:"method"`           // 方法名
	Iface           string      `json:"iface"`            // 接口名称
	Protocol        string      `json:"protocol"`         // 协议dubbo 或 triple （tri）
	Registry        string      `json:"registry"`         // 注册中心 如 "zookeeper"
	RegistryAddress string      `json:"registry_address"` // 注册中心地址： 如： “127.0.0.1:2181”
	ParameterType   []string    `json:"parameter_type"`
	Parameter       interface{} `json:"parameter"`
}

type DubboDetail struct {
	DubboProtocol string `json:"dubbo_method"`
	ApiName       string `json:"api_name"`
	FunctionName  string `json:"dubbo_iface"`
	Version       string `json:"version"`

	DubboParam  []DubboParam  `json:"dubbo_param"`
	DubboAssert []DubboAssert `json:"dubbo_assert"`
	DubboRegex  []DubboRegex  `json:"dubbo_regex"`
	DubboConfig DubboConfig   `json:"dubbo_config"`
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
