package router

import (
	"github.com/gin-gonic/gin"

	"sop/internal/api"
	"sop/internal/middleware"
)

// initInspectionContentRouter 注册维修内容相关路由
func initInspectionContentRouter(group *gin.RouterGroup) {
	api := api.InspectionContentApi{}
	r := group.Group("/inspectionContent")
	r.Use(middleware.JWTAuth())
	{
		r.POST("getInspectionContentList", api.GetInspectionContentList) // 分页获取维修内容列表 —— 【app+web使用】
	}
}
