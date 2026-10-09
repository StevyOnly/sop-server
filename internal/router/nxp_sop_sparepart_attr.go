package router

import (
	"github.com/gin-gonic/gin"

	"sop/internal/api"
	"sop/internal/middleware"
)

// initSparepartAttrRouter 注册备件 SOP 属性相关路由
func initSparepartAttrRouter(group *gin.RouterGroup) {
	attrApi := api.SparepartAttrApi{}
	attrRouter := group.Group("/sparepartAttr") // 备件属性接口路径
	attrRouter.Use(middleware.JWTAuth())
	{
		// 【web端使用】（后台备件属性管理）
		attrRouter.POST("list", attrApi.GetSparepartAttrList)                                  // 分页获取备件属性列表
		attrRouter.POST("getSparepartAttrBySId", attrApi.GetSparepartAttrBySID)                // 按备件 sId 获取属性
		attrRouter.POST("create", middleware.RequireActiveUser(), attrApi.CreateSparepartAttr) // 新增备件属性
		attrRouter.POST("update", middleware.RequireActiveUser(), attrApi.UpdateSparepartAttr) // 更新备件属性
		attrRouter.POST("delete", middleware.RequireActiveUser(), attrApi.DeleteSparepartAttr) // 删除备件属性
	}
}
