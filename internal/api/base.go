package api

import (
	"github.com/gin-gonic/gin"

	"sop/internal/service"
	"sop/pkg/response"
)

// BaseApi 基础接口处理器（登录等认证相关接口）
type BaseApi struct{}

// Login 用户登录
// 同时支持 application/json 与 multipart/form-data 两种请求体格式
func (b BaseApi) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBind(&req); err != nil {
		bindFail(c, err)
		return
	}

	result, err := service.Login(req.Name, req.Password)
	if err != nil {
		response.FailError(c, err)
		return
	}

	response.Result(c, gin.H{
		"user":      result.User,
		"token":     result.Token,
		"expire_at": result.ExpiresAt,
	}, "登录成功")
}

// AppLogin APP 端免密码登录（userId + appKey 换取永久 AppToken）。
// 仅支持 JSON 请求体；appKey 校验失败、用户未启用等均不会签发 token。
func (b BaseApi) AppLogin(c *gin.Context) {
	var req AppLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		bindFail(c, err)
		return
	}

	result, err := service.AppLogin(req.UserID, req.AppKey)
	if err != nil {
		response.FailError(c, err)
		return
	}

	response.Result(c, gin.H{
		"user":      result.User,
		"token":     result.Token,
		"expire_at": result.ExpiresAt,
	}, "登录成功")
}
