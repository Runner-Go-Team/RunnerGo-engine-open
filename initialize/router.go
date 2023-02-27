package initialize

import (
	"github.com/Runner-Go-Team/RunnerGo-engine-open/middlewares"
	"github.com/Runner-Go-Team/RunnerGo-engine-open/routers"
	"github.com/gin-gonic/gin"
)

func Routers() *gin.Engine {

	Routers := gin.Default()
	// 配置跨域
	Routers.Use(middlewares.Cors())

	groups := Routers.Group("runner")
	routers.InitRouter(groups)

	return Routers
}
