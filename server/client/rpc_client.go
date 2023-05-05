package client

import (
	"context"
	"dubbo.apache.org/dubbo-go/v3/common"
	dubboConfig "dubbo.apache.org/dubbo-go/v3/config"
	"dubbo.apache.org/dubbo-go/v3/config/generic"
	_ "dubbo.apache.org/dubbo-go/v3/imports"
	_ "dubbo.apache.org/dubbo-go/v3/metadata/service/local"
	"fmt"
	"github.com/Runner-Go-Team/RunnerGo-engine-open/model"
	"github.com/Runner-Go-Team/RunnerGo-engine-open/tools"
	hessian "github.com/apache/dubbo-go-hessian2"
)

func NewRpcClient(rpc model.DubboRequest) {
	rpcServer := newRpcServer(rpc)
	if rpcServer == nil {
		return
	}
	resp, err := rpcServer.(*generic.GenericService).Invoke(
		context.TODO(),
		rpc.Method,
		rpc.ParameterType,               // 参数类型
		[]hessian.Object{rpc.Parameter}, // 实参
	)
	if err != nil {
		fmt.Println("请求错误 :   ", err.Error())
	}

	fmt.Println("resp:    ", resp)

}

func newRpcServer(rpc model.DubboRequest) (rpcServer common.RPCService) {
	defer tools.DeferPanic("初始化dubbo配置失败")
	registryConfig := &dubboConfig.RegistryConfig{
		Protocol: rpc.Registry,
		Address:  rpc.RegistryAddress,
	}

	refConf := &dubboConfig.ReferenceConfig{
		InterfaceName: rpc.Iface, // 服务接口名
		Cluster:       "failover",
		RegistryIDs:   []string{"zk"}, // 注册中心
		Protocol:      rpc.Protocol,   // dubbo  或 tri（triple）
		Generic:       "true",
	}
	if rpc.Registry != "zookeeper" {
		refConf.RegistryIDs = append(refConf.RegistryIDs, rpc.Registry)
	}

	rootConfig := dubboConfig.NewRootConfigBuilder().AddRegistry("zk", registryConfig).Build()
	if err := dubboConfig.Load(dubboConfig.WithRootConfig(rootConfig)); err != nil {
		fmt.Println("rootconfig 错误：", err)
		return
	}

	if err := rootConfig.Init(); err != nil {
		fmt.Println("rootConfig 初始化失败：  ", err.Error())
		return
	}

	if err := refConf.Init(rootConfig); err != nil {
		fmt.Println("refConfig 初始化失败：  ", err.Error())
		return
	}

	refConf.GenericLoad(rpc.AppName)
	rpcServer = refConf.GetRPCService()
	return
}
