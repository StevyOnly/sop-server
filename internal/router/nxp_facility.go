package router

import (
	"github.com/gin-gonic/gin"

	"sop/internal/api"
	"sop/internal/middleware"
)

// initFacilityRouter 注册设备（设施）相关路由
func initFacilityRouter(group *gin.RouterGroup) {
	facilityApi := api.FacilityApi{}
	facilityRouter := group.Group("/facility") // 设备（设施）接口路径
	facilityRouter.Use(middleware.JWTAuth())
	{
		// 【web端使用】（后台设备管理）
		facilityRouter.POST("list", facilityApi.GetFacilityList) // 分页获取设备列表
		// 【app+web使用】按 ID 列表批量获取设备
		facilityRouter.POST("getFacilitiesByIDs", facilityApi.GetFacilitiesByIDs)
	}
}
