// Package middleware Logger 日志中间件单元测试
// 覆盖 Logger() 在 c.Next() 后写入日志的字段透传（status/method/path/query/ip/user_agent/cost/errors）。
// 使用 go.uber.org/zap/zaptest/observer 捕获 zap 日志条目做断言，无外部依赖。
package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

// newLoggerWithObserver 构造一个可观测的 zap.Logger，返回 logger 与捕获日志的 observer。
// 级别设为 Info 与生产 Logger() 默认输出级别一致（Logger 内部仅调用 Info）。
func newLoggerWithObserver(t *testing.T) (*zap.Logger, *observer.ObservedLogs) {
	core, recorded := observer.New(zapcore.InfoLevel)
	return zap.New(core), recorded
}

// newLoggerRouter 构造测试路由：Logger(logger) → GET/POST /
// handler 可定制响应状态码与是否注入 gin error。
func newLoggerRouter(logger *zap.Logger, status int, withError bool) *gin.Engine {
	r := gin.New()
	r.Use(Logger(logger))
	r.Handle(http.MethodGet, "/", func(c *gin.Context) {
		if withError {
			_ = c.Error(http.ErrAbortHandler) // 注入一个 private error
		}
		if status != 0 {
			c.Status(status)
			return
		}
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})
	r.Handle(http.MethodPost, "/", func(c *gin.Context) {
		c.JSON(http.StatusCreated, gin.H{"created": true})
	})
	return r
}

// TestLogger_BasicRequest GET / 默认 200 → 写入 1 条日志，字段 status/method/path/query 齐全
func TestLogger_BasicRequest(t *testing.T) {
	logger, recorded := newLoggerWithObserver(t)
	r := newLoggerRouter(logger, 0, false)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	entries := recorded.All()
	require.Len(t, entries, 1, "应写入 1 条日志")
	assert.Equal(t, zap.InfoLevel, entries[0].Level)
	assert.Equal(t, "HTTP Request", entries[0].Message)

	ctx := entries[0].ContextMap()
	assert.Equal(t, int64(200), ctx["status"])
	assert.Equal(t, "GET", ctx["method"])
	assert.Equal(t, "/", ctx["path"])
	assert.Equal(t, "", ctx["query"])
	assert.Equal(t, "", ctx["errors"])
	_, hasCost := ctx["cost"]
	assert.True(t, hasCost, "cost 字段应存在")
	_, hasIP := ctx["ip"]
	assert.True(t, hasIP, "ip 字段应存在")
}

// TestLogger_WithQuery 携带 query string → query 字段透传 RawQuery
func TestLogger_WithQuery(t *testing.T) {
	logger, recorded := newLoggerWithObserver(t)
	r := newLoggerRouter(logger, 0, false)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/?foo=bar&page=2", nil)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	entries := recorded.All()
	require.Len(t, entries, 1)
	ctx := entries[0].ContextMap()
	assert.Equal(t, "foo=bar&page=2", ctx["query"])
	assert.Equal(t, "/", ctx["path"])
}

// TestLogger_NonDefaultStatus handler 返回 201 → status 字段为 201
// 验证 status 取自 c.Writer.Status()（在 c.Next() 之后读取，反映 handler 实际写入的状态码）
func TestLogger_NonDefaultStatus(t *testing.T) {
	logger, recorded := newLoggerWithObserver(t)
	r := newLoggerRouter(logger, http.StatusCreated, false)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusCreated, w.Code)
	entries := recorded.All()
	require.Len(t, entries, 1)
	ctx := entries[0].ContextMap()
	assert.Equal(t, int64(201), ctx["status"])
}

// TestLogger_PostMethod POST 请求 → method 字段为 POST
func TestLogger_PostMethod(t *testing.T) {
	logger, recorded := newLoggerWithObserver(t)
	r := newLoggerRouter(logger, 0, false)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusCreated, w.Code)
	entries := recorded.All()
	require.Len(t, entries, 1)
	ctx := entries[0].ContextMap()
	assert.Equal(t, "POST", ctx["method"])
	assert.Equal(t, int64(201), ctx["status"])
}

// TestLogger_UserAgent 携带 User-Agent → user_agent 字段透传
func TestLogger_UserAgent(t *testing.T) {
	logger, recorded := newLoggerWithObserver(t)
	r := newLoggerRouter(logger, 0, false)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Test Runner)")
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	entries := recorded.All()
	require.Len(t, entries, 1)
	ctx := entries[0].ContextMap()
	assert.Equal(t, "Mozilla/5.0 (Test Runner)", ctx["user_agent"])
}

// TestLogger_WithErrors handler 注入 gin error → errors 字段非空
// 验证 c.Errors.ByType(gin.ErrorTypePrivate).String() 被正确采集
func TestLogger_WithErrors(t *testing.T) {
	logger, recorded := newLoggerWithObserver(t)
	r := newLoggerRouter(logger, 0, true)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	r.ServeHTTP(w, req)

	// handler 注入 error 后未显式写状态码，c.Status 默认 200
	require.Equal(t, http.StatusOK, w.Code)
	entries := recorded.All()
	require.Len(t, entries, 1)
	ctx := entries[0].ContextMap()
	errStr, _ := ctx["errors"].(string)
	assert.NotEmpty(t, errStr, "errors 字段应非空")
}

// TestLogger_CostPositive cost 字段为正数（time.Since(start)）
// 验证 cost 类型为 time.Duration 且 > 0
func TestLogger_CostPositive(t *testing.T) {
	logger, recorded := newLoggerWithObserver(t)
	r := newLoggerRouter(logger, 0, false)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	entries := recorded.All()
	require.Len(t, entries, 1)
	ctx := entries[0].ContextMap()
	cost, ok := ctx["cost"].(time.Duration)
	require.True(t, ok, "cost 应为 time.Duration 类型，实际类型=%T", ctx["cost"])
	assert.True(t, cost >= 0, "cost 应 >= 0（极速完成可能为 0）")
}

// TestLogger_AfterNext 日志在 c.Next() 之后写入 → 即使 handler 慢，日志仍能采集到最终状态
// 通过 sleep handler 验证日志写入时机在 handler 完成后
func TestLogger_AfterNext(t *testing.T) {
	logger, recorded := newLoggerWithObserver(t)
	r := gin.New()
	r.Use(Logger(logger))
	r.GET("/", func(c *gin.Context) {
		time.Sleep(5 * time.Millisecond)
		c.Status(http.StatusOK)
	})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	entries := recorded.All()
	require.Len(t, entries, 1, "日志应在 handler 完成后写入")
	ctx := entries[0].ContextMap()
	assert.Equal(t, int64(200), ctx["status"])
	cost, _ := ctx["cost"].(time.Duration)
	assert.GreaterOrEqual(t, cost, 5*time.Millisecond, "cost 应至少 5ms")
}

// TestLogger_MultipleRequests 多次请求 → 每次写一条日志，互不干扰
// 验证 logger 闭包捕获的 start/path/query 是请求级局部变量而非共享
func TestLogger_MultipleRequests(t *testing.T) {
	logger, recorded := newLoggerWithObserver(t)
	r := newLoggerRouter(logger, 0, false)

	for i := 0; i < 3; i++ {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/?i="+string(rune('0'+i)), nil)
		r.ServeHTTP(w, req)
		require.Equal(t, http.StatusOK, w.Code)
	}

	entries := recorded.All()
	require.Len(t, entries, 3, "3 次请求应写 3 条日志")
	// 每条 query 不同，验证请求级隔离
	assert.Equal(t, "i=0", entries[0].ContextMap()["query"])
	assert.Equal(t, "i=1", entries[1].ContextMap()["query"])
	assert.Equal(t, "i=2", entries[2].ContextMap()["query"])
}

// TestLogger_PathCaptured 路径透传 → path 字段为 URL.Path（不含 query）
func TestLogger_PathCaptured(t *testing.T) {
	logger, recorded := newLoggerWithObserver(t)
	r := gin.New()
	r.Use(Logger(logger))
	r.GET("/api/v1/news", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/news?keyword=foo", nil)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	entries := recorded.All()
	require.Len(t, entries, 1)
	ctx := entries[0].ContextMap()
	assert.Equal(t, "/api/v1/news", ctx["path"])
	assert.Equal(t, "keyword=foo", ctx["query"])
}
