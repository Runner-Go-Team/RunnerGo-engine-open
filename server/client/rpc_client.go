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

func NewRpcClient(dubbo model.DubboDetail) {
	rpcServer, err := newRpcServer(dubbo)
	if err != nil {
		return
	}
	parameterTypes, parameterValues := []string{}, []hessian.Object{}

	for _, parame := range dubbo.DubboParam {
		if parame.IsChecked != model.Open {
			break
		}
		parameterTypes = append(parameterTypes, parame.ParamType)
		parameterValues = append(parameterValues, parame.Val)
	}
	resp, err := rpcServer.(*generic.GenericService).Invoke(
		context.TODO(),
		dubbo.FunctionName,
		parameterTypes,
		parameterValues, // 实参
	)
	if err != nil {
		fmt.Println("请求错误 :   ", err.Error())
	}

	fmt.Println("resp:    ", resp)

}

func newRpcServer(dubbo model.DubboDetail) (rpcServer common.RPCService, err error) {
	defer tools.DeferPanic("初始化dubbo配置失败")
	registryConfig := &dubboConfig.RegistryConfig{
		Protocol: dubbo.DubboConfig.RegistrationCenterName,
		Address:  dubbo.DubboConfig.RegistrationCenterAddress,
	}

	var zk string

	if dubbo.DubboConfig.RegistrationCenterName == "zookeeper" {
		zk = "zk"
	}

	refConf := &dubboConfig.ReferenceConfig{
		InterfaceName: dubbo.ApiName, // 服务接口名
		Cluster:       "failover",
		RegistryIDs:   []string{zk},        // 注册中心
		Protocol:      dubbo.DubboProtocol, // dubbo  或 tri（triple）
		Generic:       "true",
		Version:       dubbo.Version,
	}
	if dubbo.DubboConfig.RegistrationCenterName != "zookeeper" {
		refConf.RegistryIDs = append(refConf.RegistryIDs, dubbo.DubboConfig.RegistrationCenterName)
	}

	rootConfig := dubboConfig.NewRootConfigBuilder().AddRegistry("zk", registryConfig).Build()
	if err = dubboConfig.Load(dubboConfig.WithRootConfig(rootConfig)); err != nil {
		return
	}

	if err = rootConfig.Init(); err != nil {
		return
	}

	if err = refConf.Init(rootConfig); err != nil {
		fmt.Println("refConfig 初始化失败：  ", err.Error())
		return
	}

	refConf.GenericLoad(dubbo.ApiName)
	rpcServer = refConf.GetRPCService()
	return
}
