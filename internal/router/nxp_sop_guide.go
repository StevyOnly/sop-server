package router

import (
	"github.com/gin-gonic/gin"

	"sop/internal/api"
	"sop/internal/middleware"
)

// initGuideRouter 注册指南相关路由（均需 JWT 认证）
func initGuideRouter(group *gin.RouterGroup) {
	guideApi := api.GuideApi{}
	guideRouter := group.Group("/guides")
	guideRouter.Use(middleware.JWTAuth())
	{
		// 【app+web使用】指南查询
		guideRouter.POST("list", guideApi.GetGuideList) // 分页获取指南列表
		// 【APP端使用】指南按条件查询（APP 端执行 SOP 时使用）
		guideRouter.POST("getGuideByIds", middleware.ResourceURL(), guideApi.GetGuideByIds)                                     // 按指南 ID 列表分页获取指南（含步骤）
		guideRouter.POST("getGuideSampleByMachineModelIds", middleware.ResourceURL(), guideApi.GetGuideSampleByMachineModelIds) // 按设备 ID 列表分页获取指南精简视图（不含步骤）。入参 ids 为设备 id，即 nxp_facility.id；因 app 前端不方便修改，仅后端调整逻辑：经 nxp_facility.model 中转后查询指南（仅返回启用状态 Enabled=true）
		guideRouter.POST("getStepsByFaultCode", middleware.ResourceURL(), guideApi.GetStepsByFaultCode)                         // 按故障码获取对应步骤（data 为扁平步骤数组）
		// 【web端使用】（后台 SOP 库管理）
		guideRouter.POST("save", middleware.RequireActiveUser(), guideApi.SaveGuide)           // 保存指南
		guideRouter.POST("status", middleware.RequireActiveUser(), guideApi.UpdateGuideStatus) // 更新发布状态
		guideRouter.POST("delete", middleware.RequireActiveUser(), guideApi.DeleteGuide)       // 删除指南
	}
}
