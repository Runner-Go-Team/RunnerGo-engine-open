package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/Runner-Go-Team/RunnerGo-engine-open/middlewares"
	"github.com/Runner-Go-Team/RunnerGo-engine-open/model"
	"github.com/Runner-Go-Team/RunnerGo-engine-open/tools"
	"github.com/gin-gonic/gin"
	"github.com/pyroscope-io/client/pyroscope"
	"go.uber.org/zap"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Runner-Go-Team/RunnerGo-engine-open/config"
	"github.com/Runner-Go-Team/RunnerGo-engine-open/initialize"
	"github.com/Runner-Go-Team/RunnerGo-engine-open/log"
	"github.com/Runner-Go-Team/RunnerGo-engine-open/server/heartbeat"
	_ "net/http/pprof"
)

var (
	GinRouter *gin.Engine
	mode      int
)

func initService() {

	// 初始化配置文件

	if mode == 0 {
		// 读取配置文件方式
		config.InitConfig()
	} else {
		// 读取环境变量方式
		config.EnvInitConfig()
	}

	// 初始化logger
	zap.S().Debug("初始化logger")
	log.InitLogger()

	// 获取本机地址
	heartbeat.InitLocalIp()
	// 初始化redis客户端
	log.Logger.Info(fmt.Sprintf("机器ip:%s,初始化redis客户端", middlewares.LocalIp))
	if err := model.InitRedisClient(
		config.Conf.ReportRedis.Address,
		config.Conf.ReportRedis.Password,
		config.Conf.ReportRedis.DB,
		config.Conf.Redis.Address,
		config.Conf.Redis.Password,
		config.Conf.Redis.DB,
	); err != nil {
		log.Logger.Error(fmt.Sprintf("机器ip:%s, redis连接失败:", middlewares.LocalIp), err)
		panic("redis 连接失败")
		return
	}

	//3. 初始化routers
	log.Logger.Debug(fmt.Sprintf("机器ip:%s, 初始化routers", middlewares.LocalIp))
	GinRouter = initialize.Routers()

	// 语言转换
	if err := initialize.InitTrans("zh"); err != nil {
		log.Logger.Error(fmt.Sprintf("机器ip:%s", middlewares.LocalIp), err)
	}

	// 注册服务
	log.Logger.Debug(fmt.Sprintf("机器ip:%s, 注册服务", middlewares.LocalIp))
	kpRunnerService := &http.Server{
		Addr:           config.Conf.Http.Address,
		Handler:        GinRouter,
		ReadTimeout:    config.Conf.Http.ReadTimeout * time.Millisecond,
		WriteTimeout:   config.Conf.Http.WriteTimeout * time.Millisecond,
		MaxHeaderBytes: 1 << 20,
	}

	// 初始化全局函数
	tools.InitPublicFunc()

	go func() {
		if err := kpRunnerService.ListenAndServe(); err != nil {
			log.Logger.Error(fmt.Sprintf("机器ip:%s, kpRunnerService:", middlewares.LocalIp), err)
			return
		}
	}()

	// pprof 监控
	//go func() {
	//	pprof.Register(GinRouter)
	//	GinRouter.Run(":8003")
	//}()
	// 注册并发送心跳数据
	field := middlewares.LocalIp + "_" + fmt.Sprintf("%d", config.Conf.Heartbeat.Port) + "_" + config.Conf.Heartbeat.Region
	go func() {
		//heartbeat.SendHeartBeat(config.Conf.Heartbeat.GrpcHost, config.Conf.Heartbeat.Duration)
		heartbeat.SendHeartBeatRedis(field, config.Conf.Heartbeat.Duration)
	}()
	// 资源监控数据
	go func() {
		heartbeat.SendMachineResources(config.Conf.Heartbeat.Resources)
	}()
	/// 接收终止信号
	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Logger.Info(fmt.Sprintf("机器ip:%s, 注销成功", middlewares.LocalIp))

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := kpRunnerService.Shutdown(ctx); err != nil {
		log.Logger.Info(fmt.Sprintf("机器ip:%s, 注销成功", middlewares.LocalIp))
	}
}

func main() {
	flag.IntVar(&mode, "m", 0, "读取环境变量还是读取配置文件")
	flag.Parse()
	// 性能分析
	pyroscope.Start(
		pyroscope.Config{
			ApplicationName: "RunnerGo-engine-open",
			ServerAddress:   "http://192.168.1.205:4040/",
			//Logger:          pyroscope.StandardLogger,
			ProfileTypes: []pyroscope.ProfileType{
				pyroscope.ProfileCPU,
				pyroscope.ProfileAllocObjects,
				pyroscope.ProfileAllocSpace,
				pyroscope.ProfileInuseObjects,
				pyroscope.ProfileInuseSpace,
			},
		})
	initService()
}
