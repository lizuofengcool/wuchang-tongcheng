// Package middleware CORS 跨域中间件单元测试
// 覆盖 CORS() 头注入（Origin 回显 / Allow-Methods / Allow-Headers / Allow-Credentials / Max-Age）
// 与 OPTIONS 预检短路（204 NoContent 且不调用后续 handler）。
// 使用 gin.TestMode + httptest，无外部依赖。
package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newCORSRouter 构造测试路由：CORS() → GET / 与 OPTIONS /
// handlerCalled 用于断言后续 handler 是否被触发。
func newCORSRouter(handlerCalled *bool) *gin.Engine {
	r := gin.New()
	r.Use(CORS())
	r.Handle(http.MethodGet, "/", func(c *gin.Context) {
		*handlerCalled = true
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})
	// 注册 OPTIONS 路由（实际生产由 CORS 中间件短路，此处仅占位让路由表识别）
	r.Handle(http.MethodOptions, "/", func(c *gin.Context) {
		*handlerCalled = true
		c.JSON(http.StatusOK, gin.H{"ok": "options"})
	})
	return r
}

// TestCORS_WithOrigin_GET GET 请求携带 Origin → 回显 Origin 并写入全部 CORS 头，后续 handler 正常执行
func TestCORS_WithOrigin_GET(t *testing.T) {
	called := false
	r := newCORSRouter(&called)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Origin", "https://example.com")
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.True(t, called, "GET 请求应放行到后续 handler")
	assert.Equal(t, "https://example.com", w.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "GET, POST, PUT, DELETE, PATCH, OPTIONS", w.Header().Get("Access-Control-Allow-Methods"))
	assert.Equal(t, "Content-Type, Authorization, X-Requested-With, X-Token, X-Region-ID", w.Header().Get("Access-Control-Allow-Headers"))
	assert.Equal(t, "true", w.Header().Get("Access-Control-Allow-Credentials"))
	assert.Equal(t, "86400", w.Header().Get("Access-Control-Max-Age"))
}

// TestCORS_WithoutOrigin_GET GET 请求无 Origin → 不写入任何 CORS 头，handler 正常执行
func TestCORS_WithoutOrigin_GET(t *testing.T) {
	called := false
	r := newCORSRouter(&called)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.True(t, called, "无 Origin 不应阻断请求")
	assert.Empty(t, w.Header().Get("Access-Control-Allow-Origin"))
	assert.Empty(t, w.Header().Get("Access-Control-Allow-Methods"))
	assert.Empty(t, w.Header().Get("Access-Control-Allow-Headers"))
	assert.Empty(t, w.Header().Get("Access-Control-Allow-Credentials"))
	assert.Empty(t, w.Header().Get("Access-Control-Max-Age"))
}

// TestCORS_OptionsPreflight_WithOrigin OPTIONS 预检携带 Origin → 写入 CORS 头并短路返回 204，handler 不执行
func TestCORS_OptionsPreflight_WithOrigin(t *testing.T) {
	called := false
	r := newCORSRouter(&called)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodOptions, "/", nil)
	req.Header.Set("Origin", "https://app.example.com")
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusNoContent, w.Code, "OPTIONS 预检应返回 204 NoContent")
	require.False(t, called, "OPTIONS 预检应短路，不调用后续 handler")
	assert.Equal(t, "https://app.example.com", w.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "GET, POST, PUT, DELETE, PATCH, OPTIONS", w.Header().Get("Access-Control-Allow-Methods"))
	assert.Equal(t, "Content-Type, Authorization, X-Requested-With, X-Token, X-Region-ID", w.Header().Get("Access-Control-Allow-Headers"))
	assert.Equal(t, "true", w.Header().Get("Access-Control-Allow-Credentials"))
	assert.Equal(t, "86400", w.Header().Get("Access-Control-Max-Age"))
	// 204 响应体应为空
	assert.Empty(t, w.Body.String())
}

// TestCORS_OptionsPreflight_WithoutOrigin OPTIONS 预检无 Origin → 不写 CORS 头但依旧短路 204
// 即便没有 Origin，OPTIONS 仍按预检处理（避免后续 handler 收到未预期的 OPTIONS）
func TestCORS_OptionsPreflight_WithoutOrigin(t *testing.T) {
	called := false
	r := newCORSRouter(&called)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodOptions, "/", nil)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusNoContent, w.Code)
	require.False(t, called, "OPTIONS 预检应短路，不调用后续 handler")
	assert.Empty(t, w.Header().Get("Access-Control-Allow-Origin"))
	assert.Empty(t, w.Header().Get("Access-Control-Allow-Methods"))
}

// TestCORS_OriginEchoed_DifferentValues 不同 Origin 值均原样回显（不做白名单校验，由部署层 Nginx/网关控制）
func TestCORS_OriginEchoed_DifferentValues(t *testing.T) {
	origins := []string{
		"http://localhost:5173",
		"https://wuchang.example.cn",
		"https://admin.example.com:8080",
	}
	for _, origin := range origins {
		called := false
		r := newCORSRouter(&called)
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Origin", origin)
		r.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Code, "Origin=%s", origin)
		assert.Equal(t, origin, w.Header().Get("Access-Control-Allow-Origin"))
	}
}

// TestCORS_AllowHeaders_IncludesRegionAndToken 验证 Allow-Headers 包含项目自定义头 X-Region-ID 与 X-Token
// 这两个头是前后端约定的关键头（地区隔离 + 兼容 token），缺失会导致前端请求被浏览器 CORS 拦截。
func TestCORS_AllowHeaders_IncludesRegionAndToken(t *testing.T) {
	called := false
	r := newCORSRouter(&called)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Origin", "https://test.example.com")
	r.ServeHTTP(w, req)

	allowHeaders := w.Header().Get("Access-Control-Allow-Headers")
	require.Contains(t, allowHeaders, "X-Region-ID", "必须允许地区隔离头")
	require.Contains(t, allowHeaders, "X-Token", "必须允许兼容 token 头")
	require.Contains(t, allowHeaders, "Authorization", "必须允许标准鉴权头")
	require.Contains(t, allowHeaders, "Content-Type", "必须允许 Content-Type")
}

// TestCORS_PostRequest POST 请求携带 Origin → CORS 头注入且 handler 执行
// 验证非 GET/OPTIONS 方法（POST/PUT/DELETE/PATCH）同样走头注入 + 放行路径
func TestCORS_PostRequest(t *testing.T) {
	called := false
	r := gin.New()
	r.Use(CORS())
	r.POST("/", func(c *gin.Context) {
		called = true
		c.JSON(http.StatusCreated, gin.H{"created": true})
	})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("Origin", "https://post.example.com")
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusCreated, w.Code)
	require.True(t, called)
	assert.Equal(t, "https://post.example.com", w.Header().Get("Access-Control-Allow-Origin"))
}

// TestCORS_NextHandlerExecuted_GET GET 请求 handler 写入的响应体应被透传
// 验证 CORS 中间件不吞响应体（c.Next() 后 handler 正常写 JSON）
func TestCORS_NextHandlerExecuted_GET(t *testing.T) {
	called := false
	r := newCORSRouter(&called)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Origin", "https://example.com")
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"ok":true`)
}
