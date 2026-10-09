package router

import (
	"github.com/gin-gonic/gin"

	"sop/internal/api"
	"sop/internal/middleware"
)

// initFaultCategoryRouter 注册故障分类相关路由
func initFaultCategoryRouter(group *gin.RouterGroup) {
	faultCategoryApi := api.FaultCategoryApi{}
	faultCategoryRouter := group.Group("/faultCategory")
	faultCategoryRouter.Use(middleware.JWTAuth())
	{
		// 【app+web使用】故障分类查询
		faultCategoryRouter.POST("getFaultCategoryList", faultCategoryApi.GetFaultCategoryList) // 分页获取故障分类列表
		// 【APP端使用】APP 端故障分类查询
		faultCategoryRouter.POST("getFaultCategoryByIDs", faultCategoryApi.GetFaultCategoryByIDs)             // 按 ID 列表批量获取故障分类
		faultCategoryRouter.POST("getFaultCategoriesByModelId", faultCategoryApi.GetFaultCategoriesByModelId) // 按机型 ID 分页查询关联故障分类及报警明细
		// 【web端使用】（后台故障分类管理）
		faultCategoryRouter.POST("createFaultCategory", middleware.RequireActiveUser(), faultCategoryApi.CreateFaultCategory) // 新增故障分类
		faultCategoryRouter.POST("updateFaultCategory", middleware.RequireActiveUser(), faultCategoryApi.UpdateFaultCategory) // 更新故障分类
		faultCategoryRouter.POST("deleteFaultCategory", middleware.RequireActiveUser(), faultCategoryApi.DeleteFaultCategory) // 删除故障分类
	}
}
