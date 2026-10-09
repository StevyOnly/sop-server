package middleware

import (
	"github.com/gin-gonic/gin"

	"sop/internal/service"
	bizerr "sop/pkg/errors"
)

// ActiveUserKey 在 gin 上下文中保存「已校验启用状态的用户」的键
const ActiveUserKey = "active_user"

// RequireActiveUser 关键写操作前置校验中间件：反查 Onebe 确认当前用户存在且处于启用状态。
//
// 背景：JWTAuth 只验签不查库，Onebe 侧禁用/删除用户后，已签发的 token 在有效期内仍然可用。
// 必须挂在 JWTAuth() 之后使用。
//
// 兜底约定与 service.EnsureActiveUser 一致：仅在明确判定「用户不存在」或「已禁用」时拒绝；
// Onebe 查询发生基础设施错误（不可用/超时）时放行，避免老库抖动误伤正常写请求。
func RequireActiveUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, ok := GetJWTClaims(c)
		if !ok {
			c.AbortWithStatusJSON(401, gin.H{"code": bizerr.CodeUnauthorized, "msg": "未获取到用户信息", "data": nil})
			return
		}
		user, err := service.EnsureActiveUser(claims.UserID)
		if err != nil {
			httpStatus, code := bizerr.StatusOf(err)
			c.AbortWithStatusJSON(httpStatus, gin.H{"code": code, "msg": err.Error(), "data": nil})
			return
		}
		// 校验通过且成功反查到时存入上下文，供 handler 复用（如取用户名），避免二次查询
		if user != nil {
			c.Set(ActiveUserKey, user)
		}
		c.Next()
	}
}

// GetActiveUser 取出 RequireActiveUser 校验通过的用户；
// Onebe 抖动放行时为 nil，调用方不得依赖其非空。
func GetActiveUser(c *gin.Context) *service.UserNxp {
	val, ok := c.Get(ActiveUserKey)
	if !ok {
		return nil
	}
	user, _ := val.(*service.UserNxp)
	return user
}
