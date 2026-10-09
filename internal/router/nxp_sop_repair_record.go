package router

import (
	"github.com/gin-gonic/gin"

	"sop/internal/api"
	"sop/internal/middleware"
)

// initRepairRecordRouter 注册维修记录相关路由
func initRepairRecordRouter(group *gin.RouterGroup) {
	repairApi := api.RepairRecordApi{}
	repairRouter := group.Group("/repairRecords")
	repairRouter.Use(middleware.JWTAuth())
	{
		// 【app+web使用】维修记录查询
		repairRouter.POST("list", repairApi.ListRepairRecords) // 分页获取维修记录列表
		// 【APP端使用】APP 端提交与批量查询
		repairRouter.POST("saveRepairRecordWithDetails", middleware.RequireActiveUser(), repairApi.SaveRepairRecordWithDetails) // 保存维修记录（完工/结束提交）
		repairRouter.POST("getNxpRepairRecordByIds", repairApi.GetNxpRepairRecordByIds)                                         // 按 ID 列表批量获取维修记录
	}
}
