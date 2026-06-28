// Package middleware 鉴权中间件单元测试
// 覆盖 Auth() 游客/Bearer/query token 解析、AuthRequired() 强制登录、
// GetUserID/GetUsername 辅助函数。使用 gin.TestMode + httptest，无外部依赖。
package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"wuchang-tongcheng/internal/pkg/jwt"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// newAuthRouter 构造测试路由：Auth() → {public, auth: AuthRequired()}
func newAuthRouter() *gin.Engine {
	r := gin.New()
	r.Use(Auth())
	public := r.Group("/pub")
	public.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"uid": GetUserID(c), "name": GetUsername(c)})
	})
	auth := r.Group("/auth", AuthRequired())
	auth.GET("/me", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"uid": GetUserID(c)})
	})
	return r
}

// TestAuth_NoToken_Guest 无 token 视为游客，用户ID=0，username=guest
func TestAuth_NoToken_Guest(t *testing.T) {
	r := newAuthRouter()
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/pub/", nil)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"uid":0`)
	assert.Contains(t, w.Body.String(), `"name":"guest"`)
}

// TestAuth_BearerToken 解析 Bearer token 写入上下文
func TestAuth_BearerToken(t *testing.T) {
	token, err := jwt.GenerateToken(123, "alice")
	require.NoError(t, err)

	r := newAuthRouter()
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/pub/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"uid":123`)
	assert.Contains(t, w.Body.String(), `"name":"alice"`)
}

// TestAuth_QueryToken 兼容 query 参数 ?token=<JWT>（WS 场景）
func TestAuth_QueryToken(t *testing.T) {
	token, err := jwt.GenerateToken(456, "bob")
	require.NoError(t, err)

	r := newAuthRouter()
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/pub/?token="+token, nil)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"uid":456`)
	assert.Contains(t, w.Body.String(), `"name":"bob"`)
}

// TestAuth_InvalidToken 非法 token 视为游客，不返回 401
func TestAuth_InvalidToken(t *testing.T) {
	r := newAuthRouter()
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/pub/", nil)
	req.Header.Set("Authorization", "Bearer not.a.valid.jwt")
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"uid":0`)
	assert.Contains(t, w.Body.String(), `"name":"guest"`)
}

// TestAuth_BearerPrefixMissing 直接传 token 不带 Bearer 前缀也应解析（TrimPrefix 兼容）
func TestAuth_BearerPrefixMissing(t *testing.T) {
	token, err := jwt.GenerateToken(789, "charlie")
	require.NoError(t, err)

	r := newAuthRouter()
	w := httptest.NewRecorder()
	// 不带 Bearer 前缀，TrimPrefix 不影响 token 本身
	req := httptest.NewRequest("GET", "/pub/", nil)
	req.Header.Set("Authorization", token)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	// 直接传 token 也能被 ParseToken 解析（jwt 包不要求前缀）
	assert.Contains(t, w.Body.String(), `"uid":789`)
}

// TestAuthRequired_NoToken 需登录路由无 token 返回 401（HTTP 200 + 业务码 401）
func TestAuthRequired_NoToken(t *testing.T) {
	r := newAuthRouter()
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/auth/me", nil)
	r.ServeHTTP(w, req)

	// 中间件用 c.JSON(200, response.Unauthorized) 返回业务码 401
	require.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"code":401`)
	assert.Contains(t, w.Body.String(), "请先登录")
}

// TestAuthRequired_WithToken 有效 token 通过 AuthRequired
func TestAuthRequired_WithToken(t *testing.T) {
	token, err := jwt.GenerateToken(999, "dave")
	require.NoError(t, err)

	r := newAuthRouter()
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/auth/me", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"uid":999`)
}

// TestGetUserID_NoContext 上下文无 user_id 时返回 0
func TestGetUserID_NoContext(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"uid": GetUserID(c), "name": GetUsername(c)})
	})
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"uid":0`)
	assert.Contains(t, w.Body.String(), `"name":""`)
}
