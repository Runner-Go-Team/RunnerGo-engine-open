package client

import (
	"dubbo.apache.org/dubbo-go/v3/common"
	dubboConfig "dubbo.apache.org/dubbo-go/v3/config"
	_ "dubbo.apache.org/dubbo-go/v3/imports"
	_ "dubbo.apache.org/dubbo-go/v3/metadata/service/local"
	"github.com/Runner-Go-Team/RunnerGo-engine-open/model"
	"github.com/Runner-Go-Team/RunnerGo-engine-open/tools"
)

func NewRpcServer(dubbo model.DubboDetail) (rpcServer common.RPCService, err error) {
	defer tools.DeferPanic("初始化dubbo配置失败")
	registryConfig := &dubboConfig.RegistryConfig{
		Protocol: dubbo.DubboConfig.RegistrationCenterName,
		Address:  dubbo.DubboConfig.RegistrationCenterAddress,
	}
	var zk string
	if dubbo.DubboConfig.RegistrationCenterName == "zookeeper" {
		zk = "zk"
	} else {
		zk = dubbo.DubboConfig.RegistrationCenterName
	}
	refConf := &dubboConfig.ReferenceConfig{
		InterfaceName:  dubbo.ApiName, // 服务接口名，如：org.apache.dubbo.sample.UserProvider
		Cluster:        "failover",
		RegistryIDs:    []string{zk},              // 注册中心
		Protocol:       dubbo.DubboProtocol,       // dubbo  或 tri（triple）  使用的协议
		Generic:        "true",                    // true: 使用泛化调用；false: 不适用泛化调用
		Version:        dubbo.DubboConfig.Version, // 版本号
		RequestTimeout: "3",
		Serialization:  "hessian2",
	}

	// 构造 Root 配置，引入注册中心模块
	rootConfig := dubboConfig.NewRootConfigBuilder().AddRegistry(zk, registryConfig).Build()
	if err = dubboConfig.Load(dubboConfig.WithRootConfig(rootConfig)); err != nil {
		return
	}

	//if err = rootConfig.Init(); err != nil {
	//	return
	//}

	// Reference 配置初始化，因为需要使用注册中心进行服务发现，需要传入经过配置的 rootConfig
	if err = refConf.Init(rootConfig); err != nil {
		return
	}

	refConf.GenericLoad(dubbo.ApiName)
	rpcServer = refConf.GetRPCService()
	return
}
