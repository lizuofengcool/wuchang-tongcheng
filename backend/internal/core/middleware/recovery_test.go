// Package middleware Recovery 崩溃恢复中间件单元测试
// 覆盖 Recovery() panic 捕获（string/error/struct 三类 panic 值）+ 日志写入（Error 级别 + error/path/method 字段）
// + 响应封装（HTTP 200 + 业务码 500 + "服务器内部错误"）+ Abort 短路（后续 handler 不执行）。
// 使用 gin.TestMode + httptest + zaptest/observer，无外部依赖。
package middleware

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// newRecoveryRouter 构造测试路由：Recovery(logger) → GET /
// panicValue 非 nil 时 handler 主动 panic。
// Abort 短路语义的验证见 TestRecovery_Abort_StopsSubsequentMiddleware（独立构造中间件链）。
func newRecoveryRouter(logger *zap.Logger, panicValue interface{}) *gin.Engine {
	r := gin.New()
	r.Use(Recovery(logger))
	r.GET("/", func(c *gin.Context) {
		if panicValue != nil {
			panic(panicValue)
		}
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})
	return r
}

// TestRecovery_NoPanic handler 正常返回 → 不写 Error 日志，不触发 recovery 响应
func TestRecovery_NoPanic(t *testing.T) {
	logger, recorded := newLoggerWithObserver(t)
	r := newRecoveryRouter(logger, nil)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"ok":true`)
	// 不应有 Error 级别日志（panic 才会写 Error）
	for _, e := range recorded.All() {
		assert.NotEqual(t, zap.ErrorLevel, e.Level, "不应写 Error 日志")
	}
	assert.Empty(t, recorded.FilterLevelExact(zap.ErrorLevel).All())
}

// TestRecovery_PanicString handler panic 字符串 → 200 + 业务码 500 + 服务器内部错误
func TestRecovery_PanicString(t *testing.T) {
	logger, recorded := newLoggerWithObserver(t)
	r := newRecoveryRouter(logger, "boom in handler")
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	// Recovery 内部 recover() 捕获 panic，不应向外抛
	require.NotPanics(t, func() {
		r.ServeHTTP(w, req)
	})

	require.Equal(t, http.StatusOK, w.Code, "HTTP 状态码应为 200（业务码 500 在 body 内）")
	body := w.Body.String()
	assert.Contains(t, body, `"code":500`)
	assert.Contains(t, body, `"服务器内部错误"`)
	assert.Contains(t, body, `"data":null`)
	// 写入 1 条 Error 日志
	errEntries := recorded.FilterLevelExact(zap.ErrorLevel).All()
	require.Len(t, errEntries, 1, "应写入 1 条 Error 日志")
	assert.Equal(t, "Panic recovered", errEntries[0].Message)
}

// TestRecovery_PanicError handler panic error 对象 → 同样被捕获，响应一致
func TestRecovery_PanicError(t *testing.T) {
	logger, recorded := newLoggerWithObserver(t)
	r := newRecoveryRouter(logger, errors.New("db connection lost"))
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	require.NotPanics(t, func() {
		r.ServeHTTP(w, req)
	})

	require.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"code":500`)
	assert.Contains(t, w.Body.String(), `"服务器内部错误"`)
	errEntries := recorded.FilterLevelExact(zap.ErrorLevel).All()
	require.Len(t, errEntries, 1)
}

// TestRecovery_PanicStruct handler panic 任意结构体 → zap.Any 序列化，同样被捕获
func TestRecovery_PanicStruct(t *testing.T) {
	type customError struct {
		Code int
		Msg  string
	}
	logger, recorded := newLoggerWithObserver(t)
	r := newRecoveryRouter(logger, customError{Code: 42, Msg: "weird"})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	require.NotPanics(t, func() {
		r.ServeHTTP(w, req)
	})

	require.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"code":500`)
	errEntries := recorded.FilterLevelExact(zap.ErrorLevel).All()
	require.Len(t, errEntries, 1)
	// zap.Any 序列化结构体到 error 字段
	ctx := errEntries[0].ContextMap()
	errField, ok := ctx["error"]
	assert.True(t, ok, "error 字段应存在")
	assert.NotNil(t, errField)
}

// TestRecovery_LogFields panic → Error 日志含 error/method/path 三个字段
// 使用独立 router 注册 /api/v1/news 路径以触发 panic（newRecoveryRouter 默认仅注册 / 路由）
func TestRecovery_LogFields(t *testing.T) {
	logger, recorded := newLoggerWithObserver(t)
	r := gin.New()
	r.Use(Recovery(logger))
	r.GET("/api/v1/news", func(c *gin.Context) {
		panic("explode")
	})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/news", nil)
	require.NotPanics(t, func() {
		r.ServeHTTP(w, req)
	})

	require.Equal(t, http.StatusOK, w.Code)
	errEntries := recorded.FilterLevelExact(zap.ErrorLevel).All()
	require.Len(t, errEntries, 1)
	ctx := errEntries[0].ContextMap()
	assert.Equal(t, "GET", ctx["method"], "method 字段应为 GET")
	assert.Equal(t, "/api/v1/news", ctx["path"], "path 字段应为请求路径")
	assert.Equal(t, "explode", ctx["error"], "error 字段应为 panic 值")
}

// TestRecovery_ResponseFormat 验证响应体完整 JSON 结构（code/message/data 三字段）
func TestRecovery_ResponseFormat(t *testing.T) {
	logger, _ := newLoggerWithObserver(t)
	r := newRecoveryRouter(logger, "fmt error")
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	// 完整结构断言（顺序无关，用 map 解析）
	assert.Equal(t, "application/json; charset=utf-8", w.Header().Get("Content-Type"))
	body := w.Body.String()
	assert.Contains(t, body, `"code":500`)
	assert.Contains(t, body, `"message":"服务器内部错误"`)
	assert.Contains(t, body, `"data":null`)
}

// TestRecovery_Abort_StopsSubsequentMiddleware panic 被 Recovery 捕获 → 内层中间件 defer 仍执行 + 响应为 Recovery 的 500
// 验证 panic 包含语义：Recovery → SecondMiddleware(defer 观察) → Handler(panic)。
// panic 经 c.Next() 展开栈时跳过普通语句，但 defer 仍会执行（LIFO，内层 defer 先于 Recovery defer）；
// Recovery 的 defer 随后 recover 并写响应。最终响应应为 Recovery 的 code:500（而非 gin 默认 500 空响应）。
func TestRecovery_Abort_StopsSubsequentMiddleware(t *testing.T) {
	logger, _ := newLoggerWithObserver(t)
	var (
		deferRan      bool
		handlerCalled bool
	)

	r := gin.New()
	r.Use(Recovery(logger))
	// 在 Recovery 之后注册一个中间件，用 defer 观察展开后的状态
	r.Use(func(c *gin.Context) {
		defer func() {
			deferRan = true
		}()
		c.Next()
	})
	r.GET("/", func(c *gin.Context) {
		handlerCalled = true
		panic("boom")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	require.NotPanics(t, func() {
		r.ServeHTTP(w, req)
	})

	require.Equal(t, http.StatusOK, w.Code)
	assert.True(t, handlerCalled, "panic handler 本身被调用过（panic 发生在其内部）")
	assert.True(t, deferRan, "内层中间件 defer 在 panic 展开时应执行")
	assert.Contains(t, w.Body.String(), `"code":500`, "响应应为 Recovery 写入的统一 500")
}

// TestRecovery_AfterRecoveryMiddleware_NotCalled panic 发生在第一个 handler → 后续路由 handler 不会被调用
// 验证 Recovery 在路由组中时，panic 后不会继续匹配其他路由
func TestRecovery_AfterRecoveryMiddleware_NotCalled(t *testing.T) {
	logger, _ := newLoggerWithObserver(t)
	r := gin.New()
	r.Use(Recovery(logger))
	secondHandlerCalled := false
	r.GET("/", func(c *gin.Context) {
		panic("first handler boom")
	})
	// 同路径只能注册一个 handler，此处用 NoRoute 验证 panic 后流程不会进入 404 handler
	r.NoRoute(func(c *gin.Context) {
		secondHandlerCalled = true
		c.JSON(http.StatusNotFound, gin.H{"err": "not found"})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	require.NotPanics(t, func() {
		r.ServeHTTP(w, req)
	})

	require.Equal(t, http.StatusOK, w.Code, "应返回 Recovery 的 200，而非 404")
	assert.False(t, secondHandlerCalled, "不应进入 NoRoute handler")
	assert.Contains(t, w.Body.String(), `"code":500`)
}

// TestRecovery_PostMethod panic 发生在 POST 请求 → method 字段为 POST，响应仍为统一格式
func TestRecovery_PostMethod(t *testing.T) {
	logger, recorded := newLoggerWithObserver(t)
	r := gin.New()
	r.Use(Recovery(logger))
	r.POST("/", func(c *gin.Context) {
		panic("post boom")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	require.NotPanics(t, func() {
		r.ServeHTTP(w, req)
	})

	require.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"code":500`)
	errEntries := recorded.FilterLevelExact(zap.ErrorLevel).All()
	require.Len(t, errEntries, 1)
	assert.Equal(t, "POST", errEntries[0].ContextMap()["method"])
}

// TestRecovery_PanicWithEmptyPath 根路径 panic → path 字段为 "/"
func TestRecovery_PanicWithEmptyPath(t *testing.T) {
	logger, recorded := newLoggerWithObserver(t)
	r := newRecoveryRouter(logger, "any panic")
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	errEntries := recorded.FilterLevelExact(zap.ErrorLevel).All()
	require.Len(t, errEntries, 1)
	assert.Equal(t, "/", errEntries[0].ContextMap()["path"])
}
