package client

import (
	"dubbo.apache.org/dubbo-go/v3/common"
	dubboConfig "dubbo.apache.org/dubbo-go/v3/config"
	_ "dubbo.apache.org/dubbo-go/v3/imports"
	_ "dubbo.apache.org/dubbo-go/v3/metadata/service/local"
	"fmt"
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
	}

	refConf := &dubboConfig.ReferenceConfig{
		InterfaceName: dubbo.ApiName, // 服务接口名
		Cluster:       "failover",
		RegistryIDs:   []string{zk},        // 注册中心
		Protocol:      dubbo.DubboProtocol, // dubbo  或 tri（triple）
		Generic:       "true",
		Version:       dubbo.DubboConfig.Version,
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
