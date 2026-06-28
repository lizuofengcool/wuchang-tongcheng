// Package jwt JWT 工具单元测试
// 覆盖 GenerateToken/ParseToken 的成功/失败/签名方法/配置变更等路径。
package jwt

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGenerateToken_Success 生成 token 非空且唯一
func TestGenerateToken_Success(t *testing.T) {
	token1, err := GenerateToken(1, "alice")
	require.NoError(t, err)
	require.NotEmpty(t, token1)

	token2, err := GenerateToken(2, "bob")
	require.NoError(t, err)
	require.NotEmpty(t, token2)
	assert.NotEqual(t, token1, token2, "不同用户应生成不同 token")
}

// TestParseToken_Valid 有效 token 解析回 claims
func TestParseToken_Valid(t *testing.T) {
	token, err := GenerateToken(123, "alice")
	require.NoError(t, err)

	claims, err := ParseToken(token)
	require.NoError(t, err)
	require.NotNil(t, claims)
	assert.Equal(t, uint(123), claims.UserID)
	assert.Equal(t, "alice", claims.Username)
	assert.Equal(t, "wuchang-tongcheng", claims.Issuer)
}

// TestParseToken_Invalid 非法字符串返回错误
func TestParseToken_Invalid(t *testing.T) {
	_, err := ParseToken("not.a.valid.jwt")
	require.Error(t, err)
}

// TestParseToken_Empty 空字符串返回错误
func TestParseToken_Empty(t *testing.T) {
	_, err := ParseToken("")
	require.Error(t, err)
}

// TestParseToken_WrongSecret 用不同 secret 签发的 token 解析失败
func TestParseToken_WrongSecret(t *testing.T) {
	// 用默认 secret 生成
	token, err := GenerateToken(1, "alice")
	require.NoError(t, err)

	// 改变 secret 后再解析，应失败
	origSecret := secretKey
	secretKey = []byte("a-completely-different-secret")
	defer func() { secretKey = origSecret }()

	_, err = ParseToken(token)
	require.Error(t, err, "改 secret 后旧 token 应无法解析")
}

// TestParseToken_WrongSigningMethod 用 RS256 等非 HMAC 算法签发的 token 应被拒绝
func TestParseToken_WrongSigningMethod(t *testing.T) {
	// 用 HS512 等非 HS256 HMAC 算法签发（HMAC 都被允许，但 HS256 是项目约定）
	// 这里用完全无效的签名构造 token，触发签名方法校验
	claims := Claims{
		UserID:   1,
		Username: "alice",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}
	// 用 None 签名方法（应被拒绝）
	token := jwt.NewWithClaims(jwt.SigningMethodNone, claims)
	signed, err := token.SignedString(jwt.UnsafeAllowNoneSignatureType)
	require.NoError(t, err)

	_, err = ParseToken(signed)
	require.Error(t, err, "None 签名方法应被拒绝")
}

// TestInit_ConfigChange Init 改变 secret 后旧 token 失效、新 token 有效
func TestInit_ConfigChange(t *testing.T) {
	// 备份原配置
	origSecret := secretKey
	origExpire := expireHour
	origIssuer := issuer
	defer func() {
		secretKey = origSecret
		expireHour = origExpire
		issuer = origIssuer
	}()

	// 旧配置生成 token
	oldToken, err := GenerateToken(1, "alice")
	require.NoError(t, err)

	// 切换配置
	Init("new-secret-xxx", 48)
	assert.Equal(t, "new-secret-xxx", string(secretKey))
	assert.Equal(t, 48, expireHour)

	// 旧 token 解析失败
	_, err = ParseToken(oldToken)
	require.Error(t, err, "切换 secret 后旧 token 应失效")

	// 新 token 解析成功
	newToken, err := GenerateToken(2, "bob")
	require.NoError(t, err)
	claims, err := ParseToken(newToken)
	require.NoError(t, err)
	assert.Equal(t, uint(2), claims.UserID)
}

// TestInit_EmptySecretKeepsDefault Init 传空值保持默认
func TestInit_EmptySecretKeepsDefault(t *testing.T) {
	origSecret := secretKey
	origExpire := expireHour
	defer func() {
		secretKey = origSecret
		expireHour = origExpire
	}()

	Init("", 0)
	assert.Equal(t, origSecret, secretKey, "空 secret 应保持原值")
	assert.Equal(t, origExpire, expireHour, "0 expire 应保持原值")
}

// TestInit_NegativeExpireKeepsDefault 负数 expire 保持默认
func TestInit_NegativeExpireKeepsDefault(t *testing.T) {
	origExpire := expireHour
	defer func() { expireHour = origExpire }()

	Init("", -1)
	assert.Equal(t, origExpire, expireHour)
}

// TestGenerateToken_ContainsExpectedFields token 包含标准字段
func TestGenerateToken_ContainsExpectedFields(t *testing.T) {
	token, err := GenerateToken(999, "tester")
	require.NoError(t, err)

	claims, err := ParseToken(token)
	require.NoError(t, err)
	assert.Equal(t, uint(999), claims.UserID)
	assert.Equal(t, "tester", claims.Username)
	assert.Equal(t, "wuchang-tongcheng", claims.Issuer)
	require.NotNil(t, claims.ExpiresAt)
	require.NotNil(t, claims.IssuedAt)
	require.NotNil(t, claims.NotBefore)
	// ExpiresAt 应晚于 IssuedAt
	assert.True(t, claims.ExpiresAt.After(claims.IssuedAt.Time))
}
