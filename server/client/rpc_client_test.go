package client

import (
	"github.com/Runner-Go-Team/RunnerGo-engine-open/model"
	"testing"
)

func TestNewRpcClient(t *testing.T) {
	dubbo := model.DubboDetail{
		DubboProtocol: "dubbo",
		ApiName:       "org.apache.dubbo.samples.api.GreetingsService",
		DubboConfig: model.DubboConfig{
			RegistrationCenterName:    "zookeeper",
			RegistrationCenterAddress: "172.17.101.188",
		},
		FunctionName: "sayHi",
		DubboParam: []model.DubboParam{
			model.DubboParam{
				IsChecked: 1,
				ParamType: "java.lang.String",
				Var:       "",
				Val:       "123",
			},
		},
	}
	NewRpcClient(dubbo)
	//rpc.Iface = "org.apache.dubbo.samples.api.GreetingsService"
	//rpc.ParameterType = []string{"java.lang.String"}
	//rpc.Parameter = "123"
	//rpc.RegistryAddress = "172.17.101.188:2181"
	//rpc.Registry = "zookeeper"
	//rpc.Protocol = "dubbo"
	//rpc.Method = "sayHi"
	//rpc.AppName = "demo"
	//NewRpcClient(rpc)
}
