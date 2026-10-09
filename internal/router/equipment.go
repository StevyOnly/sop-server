package router

import (
	"github.com/gin-gonic/gin"

	"sop/internal/api"
	"sop/internal/middleware"
)

// initEquipmentRouter 注册设备机型相关路由
func initEquipmentRouter(group *gin.RouterGroup) {
	equipmentApi := api.EquipmentApi{}
	equipmentRouter := group.Group("/equipment") // 设备机型接口路径
	equipmentRouter.Use(middleware.JWTAuth())
	{
		equipmentRouter.POST("list", equipmentApi.GetEquipmentList) // 获取全部设备机型（无分页）
	}
}
