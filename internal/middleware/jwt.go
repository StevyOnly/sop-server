package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"

	bizerr "sop/pkg/errors"
	"sop/pkg/jwt"
)

// JWTAuthKey 在 gin 上下文中保存 JWT 载荷的键
const JWTAuthKey = "jwt_claims"

// 认证失败响应辅助
func authFail(c *gin.Context, msg string) {
	c.AbortWithStatusJSON(401, gin.H{"code": bizerr.CodeUnauthorized, "msg": msg, "data": nil})
}

// JWTAuth 返回一个校验 Bearer Token 的 gin 中间件
// 通过后将解析出的 Claims 存入上下文，可用 GetJWTClaims(c) 取出。
// 兼容两类令牌（验签后不分来源统一提交给上层）：
//   - 登录 JWT（无 Type 标记）
//   - APP 端永久 AppToken（Claims.Type = "app"）
func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" {
			authFail(c, "未提供认证信息")
			return
		}

		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			authFail(c, "认证格式错误")
			return
		}

		claims, err := jwt.ParseToken(strings.TrimSpace(parts[1]))
		if err != nil {
			authFail(c, "Token 无效或已过期")
			return
		}

		c.Set(JWTAuthKey, claims)
		c.Next()
	}
}

// AppTokenAuth 校验特定类型的 Bearer Token 是否为本服务签发的 APP 永久 AppToken，
// 并校验其 Claims.Type == "app"。通过后同样将 Claims 存入上下文。
// 需要单独限定"仅接受 APP 永久 token"时使用（如 appLogin 之外要求更高隔离的接口）；
// 一般情况下直接使用 JWTAuth 即可（同时兼容登录 JWT 与 AppToken）。
func AppTokenAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" {
			authFail(c, "未提供认证信息")
			return
		}

		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			authFail(c, "认证格式错误")
			return
		}

		claims, err := jwt.ParseToken(strings.TrimSpace(parts[1]))
		if err != nil || claims.Type != jwt.TokenTypeApp {
			authFail(c, "Token 无效或已过期")
			return
		}

		c.Set(JWTAuthKey, claims)
		c.Next()
	}
}

// GetJWTClaims 从 gin 上下文中取出 JWT 载荷
func GetJWTClaims(c *gin.Context) (*jwt.Claims, bool) {
	val, ok := c.Get(JWTAuthKey)
	if !ok {
		return nil, false
	}
	claims, ok := val.(*jwt.Claims)
	return claims, ok
}
