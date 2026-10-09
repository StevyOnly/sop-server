package router

import (
	"github.com/gin-gonic/gin"

	"sop/internal/api"
	"sop/internal/middleware"
)

// initInspectionQuestionTypeRouter 注册问题类型相关路由
func initInspectionQuestionTypeRouter(group *gin.RouterGroup) {
	api := api.InspectionQuestionTypeApi{}
	r := group.Group("/inspectionQuestionType")
	r.Use(middleware.JWTAuth())
	{
		r.POST("getInspectionQuestionTypeList", api.GetInspectionQuestionTypeList) // 分页获取问题类型列表 —— 【web端使用】（后台数据字典）
	}
}
