package client

import (
	"fmt"
	"github.com/samuel/go-zookeeper/zk"
	"testing"
	"time"
)

func TestNewRpcClient(t *testing.T) {
	//dubbo := model.DubboDetail{
	//	DubboProtocol: "dubbo",
	//	ApiName:       "org.apache.dubbo.samples.api.GreetingsService",
	//	DubboConfig: model.DubboConfig{
	//		RegistrationCenterName:    "zookeeper",
	//		RegistrationCenterAddress: "",
	//	},
	//	FunctionName: "sayHi",
	//	DubboParam: []model.DubboParam{
	//		model.DubboParam{
	//			IsChecked: 1,
	//			ParamType: "java.lang.String",
	//			Var:       "",
	//			Val:       "123",
	//		},
	//	},
	//}
	//NewRpcClient(dubbo)
	//rpc.Iface = "org.apache.dubbo.samples.api.GreetingsService"
	//rpc.ParameterType = []string{"java.lang.String"}
	//rpc.Parameter = "123"
	//rpc.RegistryAddress = ":2181"
	//rpc.Registry = "zookeeper"
	//rpc.Protocol = "dubbo"
	//rpc.Method = "sayHi"
	//rpc.AppName = "demo"
	//NewRpcClient(rpc)

	// 测试连接zk
	hosts := []string{":2181"}
	conn, event, err := zk.Connect(hosts, time.Second*5)
	defer conn.Close()

	if err != nil {
		fmt.Println("err:   ", err)
		return
	}
	fmt.Println("event:      ", event)
	fmt.Println("conn:       ", conn.Server())
}
