package router

import (
	"github.com/gin-gonic/gin"

	"sop/internal/api"
	"sop/internal/middleware"
)

// initSparepartRouter 注册备件相关路由
func initSparepartRouter(group *gin.RouterGroup) {
	sparepartApi := api.SparepartApi{}
	sparepartRouter := group.Group("/sparepart") // 备件接口路径
	sparepartRouter.Use(middleware.JWTAuth())
	{
		sparepartRouter.POST("list", sparepartApi.GetSparepartList)                 // 分页获取备件列表 —— 【APP端使用】（APP 领料；后台端未调用）
		sparepartRouter.POST("listWithAttr", sparepartApi.GetSparepartListWithAttr) // 分页获取备件列表（含 SOP 属性 standardCycle/safetyFactor）
	}
}
