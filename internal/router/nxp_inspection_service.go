package router

import (
	"github.com/gin-gonic/gin"

	"sop/internal/api"
	"sop/internal/middleware"
)

// initInspectionServiceRouter 注册维修服务相关路由（/inspectionService）
func initInspectionServiceRouter(group *gin.RouterGroup) {
	serviceApi := api.InspectionServiceApi{}
	serviceRouter := group.Group("/inspectionService")
	serviceRouter.Use(middleware.JWTAuth())
	{
		// 【web端使用】备件更换列表
		serviceRouter.POST("sparepartReplaceList", serviceApi.SparepartReplaceList)               // 备件更换列表
		serviceRouter.POST("sparepartReplaceExcelExport", serviceApi.SparepartReplaceExcelExport) // 备件更换台账导出 Excel
		serviceRouter.POST("getSModelFromNxpInspectionService", serviceApi.SparepartModelOptions) // 备件型号下拉
	}
}
