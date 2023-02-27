package routers

import (
	"github.com/Runner-Go-Team/RunnerGo-engine-open/api"
	"github.com/gin-gonic/gin"
)

func InitRouter(Router *gin.RouterGroup) {
	{
		Router.POST("/run_plan/", api.RunPlan)
		Router.POST("/run_api/", api.RunApi)
		Router.POST("/run_scene/", api.RunScene)
		Router.POST("run_auto_plan/", api.RunAutoPlan)
		//Router.POST("/stop/", api.Stop)
		//Router.POST("/stop_scene/", api.StopScene)
	}
}
