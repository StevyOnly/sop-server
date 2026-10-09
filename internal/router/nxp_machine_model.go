package router

import (
	"github.com/gin-gonic/gin"

	"sop/internal/api"
	"sop/internal/middleware"
)

// initMachineModelRouter 注册机型相关路由
func initMachineModelRouter(group *gin.RouterGroup) {
	machineModelApi := api.MachineModelApi{}
	machineModelRouter := group.Group("/machineModel")
	machineModelRouter.Use(middleware.JWTAuth())
	{
		// 【app+web使用】机型列表查询
		machineModelRouter.POST("getMachineModelList", machineModelApi.GetMachineModelList) // 分页获取机型列表
		// 【APP端使用】按 ID 批量获取机型
		machineModelRouter.POST("getMachineModelsByIDs", machineModelApi.GetMachineModelsByIDs)
	}
}
