package api

import (
	"sop/global"
	"sop/pkg/response"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// LoginRequest 登录请求（nxp_user.name + password）
type LoginRequest struct {
	Name     string `json:"name" form:"name" binding:"required"`
	Password string `json:"password" form:"password" binding:"required"`
}

// AppLoginRequest APP 端免密码登录请求（userId + appKey）
type AppLoginRequest struct {
	UserID uint   `json:"userId" binding:"required"`
	AppKey string `json:"appKey" binding:"required"`
}

// UserListRequest 分页获取用户列表请求
type UserListRequest struct {
	response.PageQuery
	Name     string `json:"name"`
	Tel      string `json:"tel"`
	Email    string `json:"email"`
	SFC      string `json:"sfcNumber"`
	FSL      string `json:"fslNumber"`
	Dept     string `json:"departmentNumber"`
	RoleID   *int   `json:"roleId"`
	Position *int   `json:"position"`
	Status   *int8  `json:"status"`
}

// bindFail 统一的请求参数绑定失败处理：
// 对外仅返回泛化提示，避免把校验器的内部错误原文（字段路径、tag 细节）泄露给客户端；
// 真实错误记入服务端日志便于排查。
func bindFail(c *gin.Context, err error) {
	global.Logger.Warn("request bind failed",
		zap.String("path", c.Request.URL.Path),
		zap.Error(err))
	response.Fail(c, "请求参数错误")
}
