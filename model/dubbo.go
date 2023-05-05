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
