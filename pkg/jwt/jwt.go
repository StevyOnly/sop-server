package jwt

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// 签名密钥与有效期（默认值，可通过 Init 用配置覆盖）
var (
	secretKey   = []byte("sop-dev-secret-change-me")
	expiresDays = 7
)

// Init 使用配置初始化 JWT 签名密钥与有效期
func Init(secret string, days int) {
	if secret != "" {
		secretKey = []byte(secret)
	}
	if days > 0 {
		expiresDays = days
	}
}

// Claims 自定义 JWT 载荷
type Claims struct {
	UserID uint   `json:"userId"`
	Type   string `json:"type,omitempty"` // 令牌类型：空=登录 JWT；"app"=APP 端永久 AppToken
	jwt.RegisteredClaims
}

// GenerateToken 根据用户 ID 生成 JWT Token（UserID 对应 nxp_user.id）
func GenerateToken(userID uint) (tokenString string, expiresAt int64, err error) {
	expiresAt = time.Now().Add(time.Duration(expiresDays) * 24 * time.Hour).UnixMilli()

	claims := Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.UnixMilli(expiresAt)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "sop",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err = token.SignedString(secretKey)
	return tokenString, expiresAt, err
}

// appTokenExpYears APP 端永久令牌的固定有效期长度（年）。
// 采用足够大的固定期限而非真正永久，规避部分 JWT 库/中间件对极端时间戳的处理异常，
// 同时保证 token 在一段时间内实际等效永久（期间可通过更换 appKey 全局吊销）。
const appTokenExpYears = 100

// TokenTypeApp APP 端永久令牌在 Claims.Type 中的标记值
const TokenTypeApp = "app"

// GenerateAppToken 根据用户 ID 生成 APP 端永久令牌（AppToken）。
// 与 GenerateToken 同一签名密钥，验签路径一致；载荷 Type 标记为 "app"，
// ExpiresAt 设为固定大期限近似永久。调用前必须已校验 appKey 与用户启用状态。
func GenerateAppToken(userID uint) (tokenString string, err error) {
	expiresAt := time.Now().AddDate(appTokenExpYears, 0, 0)

	claims := Claims{
		UserID: userID,
		Type:   TokenTypeApp,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "sop",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secretKey)
}

// ParseToken 解析并校验 JWT Token，返回自定义 Claims
func ParseToken(tokenString string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return secretKey, nil
	})
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}
