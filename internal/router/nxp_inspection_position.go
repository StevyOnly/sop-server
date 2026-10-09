package router

import (
	"github.com/gin-gonic/gin"

	"sop/internal/api"
	"sop/internal/middleware"
)

// initInspectionPositionRouter 注册维修位置相关路由
func initInspectionPositionRouter(group *gin.RouterGroup) {
	api := api.InspectionPositionApi{}
	r := group.Group("/inspectionPosition")
	r.Use(middleware.JWTAuth())
	{
		r.POST("getInspectionPositionList", api.GetInspectionPositionList) // 分页获取维修位置列表 —— 【app+web使用】
	}
}
