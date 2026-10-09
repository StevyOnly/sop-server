package router

import (
	"github.com/gin-gonic/gin"

	"sop/internal/api"
	"sop/internal/middleware"
)

// initExecutionOptionRouter 注册执行说明选项相关路由
func initExecutionOptionRouter(group *gin.RouterGroup) {
	execOptionApi := api.ExecutionOptionApi{}
	execOptionRouter := group.Group("/executionOption")
	execOptionRouter.Use(middleware.JWTAuth())
	{
		// 【web端使用】（后台执行说明管理）
		execOptionRouter.POST("getExecutionOptionList", execOptionApi.GetExecutionOptionList)                               // 分页获取执行说明选项列表
		execOptionRouter.POST("createExecutionOption", middleware.RequireActiveUser(), execOptionApi.CreateExecutionOption) // 新增执行说明选项
		execOptionRouter.POST("updateExecutionOption", middleware.RequireActiveUser(), execOptionApi.UpdateExecutionOption) // 更新执行说明选项
		execOptionRouter.POST("deleteExecutionOption", middleware.RequireActiveUser(), execOptionApi.DeleteExecutionOption) // 删除执行说明选项
		// 【APP端使用】供 APP 下拉
		execOptionRouter.POST("getActiveExecutionOptions", execOptionApi.GetActiveExecutionOptions) // 获取启用选项
	}
}
