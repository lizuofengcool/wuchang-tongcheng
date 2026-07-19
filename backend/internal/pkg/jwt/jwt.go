// Package jwt JWT工具
// 提供Token的生成、解析和校验功能
package jwt

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Claims 自定义JWT声明
type Claims struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	Phone    string `json:"phone,omitempty"`  // 手机号（冗余携带，便于业务模块冗余存储，避免每次发布都查 users 表）
	Avatar   string `json:"avatar,omitempty"` // 头像 URL（同上）
	jwt.RegisteredClaims
}

// 配置（由Init设置）
var (
	secretKey  = []byte("wuchang-tongcheng-secret-key")
	expireHour = 24 // 默认24小时
	issuer     = "wuchang-tongcheng"
)

// Init 初始化JWT配置
func Init(secret string, expire int) {
	if secret != "" {
		secretKey = []byte(secret)
	}
	if expire > 0 {
		expireHour = expire
	}
}

// GenerateToken 生成Token（仅携带基础信息，兼容旧调用方与测试）
// 需要携带 phone/avatar 时请使用 GenerateTokenWithProfile。
func GenerateToken(userID uint, username string) (string, error) {
	return GenerateTokenWithProfile(userID, username, "", "")
}

// GenerateTokenWithProfile 生成携带 phone/avatar 的 Token
// 用于登录场景，业务模块（如 ershou 发布）可从 Context 直接取这些冗余字段，
// 避免每次发布都额外查询 users 表。
func GenerateTokenWithProfile(userID uint, username, phone, avatar string) (string, error) {
	claims := Claims{
		UserID:   userID,
		Username: username,
		Phone:    phone,
		Avatar:   avatar,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    issuer,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(expireHour) * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secretKey)
}

// ParseToken 解析Token
func ParseToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return secretKey, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}
