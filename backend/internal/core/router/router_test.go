// Package router 路由封装单元测试
// 覆盖 Router（Gin 引擎封装）/ RouterGroup（plugin.RouterGroup 适配器）/ Context（plugin.Context 适配器）
// / fileHeader（multipart.FileHeader 适配器）/ responseWriter（plugin.ResponseWriter 适配器）
// / convertHandlers（plugin.HandlerFunc→gin.HandlerFunc 闭包捕获）/ WrapGin（gin 中间件→plugin.HandlerFunc）
// 全方法全分支，使用 gin.TestMode + httptest，无 DB/Redis/Docker 依赖。
package router

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"wuchang-tongcheng/internal/core/plugin"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMain 设置 gin 测试模式，抑制路由注册日志噪声。
func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	m.Run()
}

// ============================ Router ============================

// TestNewRouter NewRouter 返回非 nil Router 且 Engine() 返回非 nil gin.Engine
func TestNewRouter(t *testing.T) {
	r := NewRouter()
	require.NotNil(t, r)
	require.NotNil(t, r.Engine(), "Engine() 不应为 nil")
}

// TestRouter_Engine 同一实例返回同一 gin.Engine 指针
func TestRouter_Engine(t *testing.T) {
	r := NewRouter()
	e1 := r.Engine()
	e2 := r.Engine()
	require.Same(t, e1, e2, "多次 Engine() 调用应返回同一指针")
}

// TestRouter_GET 注册 GET 路由 → 命中 handler 返回 200 JSON
func TestRouter_GET(t *testing.T) {
	r := NewRouter()
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"msg": "pong"})
	})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `{"msg":"pong"}`, w.Body.String())
}

// TestRouter_POST 注册 POST 路由 → 命中 handler 读取 body 返回 201
func TestRouter_POST(t *testing.T) {
	r := NewRouter()
	r.POST("/items", func(c *gin.Context) {
		var body map[string]string
		require.NoError(t, c.ShouldBindJSON(&body))
		c.JSON(http.StatusCreated, body)
	})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/items", strings.NewReader(`{"name":"a"}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusCreated, w.Code)
	assert.JSONEq(t, `{"name":"a"}`, w.Body.String())
}

// TestRouter_PUT 注册 PUT 路由 → 命中 handler
func TestRouter_PUT(t *testing.T) {
	r := NewRouter()
	r.PUT("/items/:id", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"id": c.Param("id")})
	})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/items/42", nil)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `{"id":"42"}`, w.Body.String())
}

// TestRouter_DELETE 注册 DELETE 路由 → 命中 handler 返回 204
func TestRouter_DELETE(t *testing.T) {
	r := NewRouter()
	r.DELETE("/items/:id", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/items/9", nil)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusNoContent, w.Code)
	assert.Empty(t, w.Body.String())
}

// TestRouter_PATCH 注册 PATCH 路由 → 命中 handler
func TestRouter_PATCH(t *testing.T) {
	r := NewRouter()
	r.PATCH("/items/:id", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"patched": c.Param("id")})
	})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPatch, "/items/7", nil)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `{"patched":"7"}`, w.Body.String())
}

// TestRouter_Any Any 注册所有 HTTP 方法 → GET/POST/PUT/DELETE/PATCH 均命中同一 handler
func TestRouter_Any(t *testing.T) {
	r := NewRouter()
	r.Any("/all", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"m": c.Request.Method})
	})
	for _, m := range []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodPatch} {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(m, "/all", nil)
		r.ServeHTTP(w, req)
		require.Equal(t, http.StatusOK, w.Code, "方法 %s 应命中", m)
		assert.JSONEq(t, `{"m":"`+m+`"}`, w.Body.String(), "方法 %s 响应体", m)
	}
}

// TestRouter_Use Use 注册全局中间件 → 中间件对每个请求注入响应头
func TestRouter_Use(t *testing.T) {
	r := NewRouter()
	r.Use(func(c *gin.Context) {
		c.Header("X-Mw", "1")
		c.Next()
	})
	r.GET("/x", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "1", w.Header().Get("X-Mw"))
}

// TestRouter_Group_GroupFunc Group 创建根路由组、GroupFunc 返回原生 gin.RouterGroup，均可注册子路由
func TestRouter_Group_GroupFunc(t *testing.T) {
	r := NewRouter()

	// Group 返回 *RouterGroup（plugin 适配器），通过其 gin 路由注册
	rg := r.Group("/v1")
	require.NotNil(t, rg)
	// RouterGroup 暴露的 GET 注册 plugin.HandlerFunc
	rg.GET("/plugin", func(ctx plugin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"src": "plugin"})
	})

	// GroupFunc 返回原生 *gin.RouterGroup
	gg := r.GroupFunc("/v2")
	require.NotNil(t, gg)
	gg.GET("/native", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"src": "native"})
	})

	// 验证 /v1/plugin
	w1 := httptest.NewRecorder()
	r.ServeHTTP(w1, httptest.NewRequest(http.MethodGet, "/v1/plugin", nil))
	require.Equal(t, http.StatusOK, w1.Code)
	assert.JSONEq(t, `{"src":"plugin"}`, w1.Body.String())

	// 验证 /v2/native
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, httptest.NewRequest(http.MethodGet, "/v2/native", nil))
	require.Equal(t, http.StatusOK, w2.Code)
	assert.JSONEq(t, `{"src":"native"}`, w2.Body.String())
}

// TestRouter_ServeHTTP Router 实现 http.Handler 接口，可直接作为 http.Server.Handler 使用
func TestRouter_ServeHTTP(t *testing.T) {
	r := NewRouter()
	r.GET("/h", func(c *gin.Context) {
		c.String(http.StatusOK, "hello")
	})

	var _ http.Handler = r // 编译期断言实现 http.Handler

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/h", nil)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "hello", w.Body.String())
}

// TestRouter_Run_InvalidAddr Run 对非法端口地址立即返回错误（不阻塞）
func TestRouter_Run_InvalidAddr(t *testing.T) {
	r := NewRouter()
	// 端口超出 65535 为非法端口，ListenAndServe 会立即返回错误
	err := r.Run("127.0.0.1:70000")
	require.Error(t, err, "非法端口应返回错误")
}

// TestRouter_404 未注册路由返回 404
func TestRouter_404(t *testing.T) {
	r := NewRouter()
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/nope", nil)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusNotFound, w.Code)
}

// ============================ RouterGroup（plugin.HandlerFunc） ============================

// TestRouterGroup_Methods RouterGroup 通过 plugin.HandlerFunc 注册 GET/POST/PUT/DELETE/PATCH 路由
func TestRouterGroup_Methods(t *testing.T) {
	r := NewRouter()
	rg := r.Group("/api")

	rg.GET("/g", func(ctx plugin.Context) { ctx.JSON(http.StatusOK, gin.H{"m": "GET"}) })
	rg.POST("/g", func(ctx plugin.Context) { ctx.JSON(http.StatusCreated, gin.H{"m": "POST"}) })
	rg.PUT("/g", func(ctx plugin.Context) { ctx.JSON(http.StatusOK, gin.H{"m": "PUT"}) })
	rg.DELETE("/g", func(ctx plugin.Context) { ctx.Status(http.StatusNoContent) })
	rg.PATCH("/g", func(ctx plugin.Context) { ctx.JSON(http.StatusOK, gin.H{"m": "PATCH"}) })

	cases := []struct {
		method string
		code   int
		body   string
	}{
		{http.MethodGet, http.StatusOK, `{"m":"GET"}`},
		{http.MethodPost, http.StatusCreated, `{"m":"POST"}`},
		{http.MethodPut, http.StatusOK, `{"m":"PUT"}`},
		{http.MethodDelete, http.StatusNoContent, ``},
		{http.MethodPatch, http.StatusOK, `{"m":"PATCH"}`},
	}
	for _, c := range cases {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(c.method, "/api/g", nil)
		r.ServeHTTP(w, req)
		require.Equal(t, c.code, w.Code, "方法 %s 期望码 %d", c.method, c.code)
		if c.body != "" {
			assert.JSONEq(t, c.body, w.Body.String(), "方法 %s 响应体", c.method)
		}
	}
}

// TestRouterGroup_Group_Nested RouterGroup.Group 创建嵌套子路由组，子组路由可命中
func TestRouterGroup_Group_Nested(t *testing.T) {
	r := NewRouter()
	rg := r.Group("/api")
	sub := rg.Group("/v1")
	require.NotNil(t, sub)
	sub.GET("/users", func(ctx plugin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"list": []string{"a", "b"}})
	})

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/users", nil))
	require.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `{"list":["a","b"]}`, w.Body.String())
}

// TestRouterGroup_Group_WithMiddleware RouterGroup.Group 透传中间件 handler，中间件先于业务 handler 执行
func TestRouterGroup_Group_WithMiddleware(t *testing.T) {
	r := NewRouter()
	rg := r.Group("/api")
	mwCalled := false
	sub := rg.Group("/v1", func(ctx plugin.Context) {
		mwCalled = true
		ctx.Set("mw", "yes")
	})
	sub.GET("/ping", func(ctx plugin.Context) {
		v, _ := ctx.Get("mw")
		ctx.JSON(http.StatusOK, gin.H{"mw": v})
	})

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/ping", nil))
	require.Equal(t, http.StatusOK, w.Code)
	require.True(t, mwCalled, "组中间件应被调用")
	assert.JSONEq(t, `{"mw":"yes"}`, w.Body.String())
}

// TestConvertHandlers_ClosureCapture convertHandlers 内部用 `handler := h` 独立捕获每个 handler，
// 验证多 handler 闭包变量隔离（不退化为最后一个）且多 handler 路由串行执行。
func TestConvertHandlers_ClosureCapture(t *testing.T) {
	// convertHandlers 内部对每个 handler 用 `handler := h` 独立捕获，
	// 这里直接验证：3 个 handler 各自设置独立 key，调用后互不干扰（验证闭包变量隔离，非循环变量退化）
	ghs := convertHandlers(
		func(ctx plugin.Context) { ctx.Set("a", 1) },
		func(ctx plugin.Context) { ctx.Set("b", 2) },
		func(ctx plugin.Context) { ctx.Set("c", 3) },
	)
	require.Len(t, ghs, 3)

	// 逐个调用，验证每个 handler 设置各自的 key（验证闭包变量隔离）
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	for _, gh := range ghs {
		gh(c)
	}
	assert.Equal(t, 1, c.Value("a"))
	assert.Equal(t, 2, c.Value("b"))
	assert.Equal(t, 3, c.Value("c"))

	// 端到端验证：通过 RouterGroup 注册一条带多个 handler 的路由，全部串行执行
	r := NewRouter()
	rg := r.Group("/api")
	rg.GET("/chain",
		func(ctx plugin.Context) { ctx.Set("step1", true) },
		func(ctx plugin.Context) { ctx.Set("step2", true) },
		func(ctx plugin.Context) { ctx.JSON(http.StatusOK, gin.H{"done": true}) },
	)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/chain", nil))
	require.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `{"done":true}`, w.Body.String())
}

// TestConvertHandlers_Empty convertHandlers 空入参返回长度 0 的非 nil 切片
func TestConvertHandlers_Empty(t *testing.T) {
	ghs := convertHandlers()
	require.Len(t, ghs, 0)
}

// ============================ WrapGin ============================

// TestWrapGin_WithContext WrapGin 包装 gin 中间件，经 RouterGroup 注册后中间件正常执行
func TestWrapGin_WithContext(t *testing.T) {
	r := NewRouter()
	rg := r.Group("/api")
	mwCalled := false

	rg.GET("/secret", WrapGin(func(c *gin.Context) {
		mwCalled = true
		c.Header("X-Auth", "ok")
		c.Next()
	}), func(ctx plugin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"ok": true})
	})

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/secret", nil))
	require.Equal(t, http.StatusOK, w.Code)
	require.True(t, mwCalled, "WrapGin 包装的 gin 中间件应被调用")
	assert.Equal(t, "ok", w.Header().Get("X-Auth"))
	assert.JSONEq(t, `{"ok":true}`, w.Body.String())
}

// TestWrapGin_NonContextContext WrapGin 返回的 func 传入非 *Context 的 plugin.Context 时直接返回（不 panic、不调用 mw）
func TestWrapGin_NonContextContext(t *testing.T) {
	mwCalled := false
	wrapped := WrapGin(func(c *gin.Context) {
		mwCalled = true
	})
	// 传入一个非 *Context 的 plugin.Context 实现
	wrapped(fakeContext{})
	require.False(t, mwCalled, "非 *Context 入参时中间件不应被调用")
}

// ============================ Context 适配器 ============================

// newTestRouterWith 构造测试 Router 并注册单条 plugin.HandlerFunc 路由，
// 返回 Router 与指向捕获指针的地址（**Context）。调用方在 ServeHTTP 之后再解引用 *ptr 取得 *Context，
// 因为闭包内对 captured 的赋值需通过指针回传给调用方（直接返回 *Context 值拷贝会丢失）。
func newTestRouterWith(t *testing.T, method, path string, h plugin.HandlerFunc) (*Router, **Context) {
	t.Helper()
	r := NewRouter()
	var captured *Context
	rg := r.Group("/api")
	wrap := func(ctx plugin.Context) {
		captured = ctx.(*Context)
		h(ctx)
	}
	switch method {
	case http.MethodGet:
		rg.GET(path, wrap)
	case http.MethodPost:
		rg.POST(path, wrap)
	case http.MethodPut:
		rg.PUT(path, wrap)
	default:
		t.Fatalf("unsupported method %s", method)
	}
	return r, &captured
}

// TestContext_JSON JSON 返回 JSON 响应，状态码与 body 透传
func TestContext_JSON(t *testing.T) {
	r, _ := newTestRouterWith(t, http.MethodGet, "/j", func(ctx plugin.Context) {
		ctx.JSON(http.StatusAccepted, gin.H{"k": "v"})
	})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/j", nil))
	require.Equal(t, http.StatusAccepted, w.Code)
	assert.JSONEq(t, `{"k":"v"}`, w.Body.String())
	assert.Equal(t, "application/json; charset=utf-8", w.Header().Get("Content-Type"))
}

// TestContext_Param Param 读取 URL 路径参数
func TestContext_Param(t *testing.T) {
	r, _ := newTestRouterWith(t, http.MethodGet, "/u/:id", func(ctx plugin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"id": ctx.Param("id")})
	})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/u/77", nil))
	require.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `{"id":"77"}`, w.Body.String())
}

// TestContext_Query_DefaultQuery Query 读取 query 参数；DefaultQuery 缺省时回退默认值
func TestContext_Query_DefaultQuery(t *testing.T) {
	r, _ := newTestRouterWith(t, http.MethodGet, "/q", func(ctx plugin.Context) {
		ctx.JSON(http.StatusOK, gin.H{
			"q":  ctx.Query("q"),
			"d":  ctx.DefaultQuery("d", "def"),
			"d2": ctx.DefaultQuery("d2", "def2"),
		})
	})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/q?q=hello&d=set", nil))
	require.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `{"q":"hello","d":"set","d2":"def2"}`, w.Body.String())
}

// TestContext_PostForm PostForm 读取表单字段
func TestContext_PostForm(t *testing.T) {
	r, _ := newTestRouterWith(t, http.MethodPost, "/f", func(ctx plugin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"name": ctx.PostForm("name")})
	})
	body := strings.NewReader("name=alice")
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/f", body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `{"name":"alice"}`, w.Body.String())
}

// TestContext_Bind Bind 绑定 JSON body 到结构体
func TestContext_Bind(t *testing.T) {
	type payload struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}
	r, _ := newTestRouterWith(t, http.MethodPost, "/b", func(ctx plugin.Context) {
		var p payload
		require.NoError(t, ctx.Bind(&p))
		ctx.JSON(http.StatusOK, p)
	})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/b", strings.NewReader(`{"name":"bob","age":30}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `{"name":"bob","age":30}`, w.Body.String())
}

// TestContext_Bind_Error Bind 绑定非法 JSON 返回错误
func TestContext_Bind_Error(t *testing.T) {
	type payload struct {
		Name string `json:"name"`
	}
	r, _ := newTestRouterWith(t, http.MethodPost, "/b", func(ctx plugin.Context) {
		var p payload
		err := ctx.Bind(&p)
		require.Error(t, err)
		ctx.Status(http.StatusBadRequest)
	})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/b", strings.NewReader(`{bad json`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusBadRequest, w.Code)
}

// TestContext_Set_Get Set/Get 读写上下文值；Get 未命中返回 ok=false
func TestContext_Set_Get(t *testing.T) {
	r, capturedPtr := newTestRouterWith(t, http.MethodGet, "/s", func(ctx plugin.Context) {
		ctx.Set("k1", "v1")
		ctx.JSON(http.StatusOK, gin.H{})
	})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/s", nil))
	require.Equal(t, http.StatusOK, w.Code)
	captured := *capturedPtr
	require.NotNil(t, captured)
	v, ok := captured.Get("k1")
	require.True(t, ok)
	assert.Equal(t, "v1", v)
	_, ok = captured.Get("missing")
	assert.False(t, ok, "未设置的 key 应返回 ok=false")
}

// TestContext_GetHeader GetHeader 读取请求头
func TestContext_GetHeader(t *testing.T) {
	r, capturedPtr := newTestRouterWith(t, http.MethodGet, "/h", func(ctx plugin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"x": ctx.GetHeader("X-Custom")})
	})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/h", nil)
	req.Header.Set("X-Custom", "abc")
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `{"x":"abc"}`, w.Body.String())
	captured := *capturedPtr
	require.NotNil(t, captured)
	assert.Equal(t, "abc", captured.GetHeader("X-Custom"))
	assert.Empty(t, captured.GetHeader("X-Missing"))
}

// TestContext_Status Status 设置响应状态码（无 body）
func TestContext_Status(t *testing.T) {
	r, _ := newTestRouterWith(t, http.MethodGet, "/st", func(ctx plugin.Context) {
		ctx.Status(http.StatusTeapot)
	})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/st", nil))
	require.Equal(t, http.StatusTeapot, w.Code)
	assert.Empty(t, w.Body.String())
}

// TestContext_Writer Writer() 返回的 ResponseWriter 支持 Write/WriteHeader/Header
func TestContext_Writer(t *testing.T) {
	r, _ := newTestRouterWith(t, http.MethodGet, "/w", func(ctx plugin.Context) {
		rw := ctx.Writer()
		// responseWriter.Header() 返回原生 map[string][]string（非 http.Header），直接赋值
		rw.Header()["X-Test"] = []string{"1"}
		rw.WriteHeader(http.StatusCreated)
		_, _ = rw.Write([]byte("raw"))
	})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/w", nil))
	require.Equal(t, http.StatusCreated, w.Code)
	assert.Equal(t, "raw", w.Body.String())
	assert.Equal(t, "1", w.Header().Get("X-Test"))
}

// TestContext_Request Request() 始终返回 nil（设计上返回原生 gin.Context 取代）
func TestContext_Request(t *testing.T) {
	r, _ := newTestRouterWith(t, http.MethodGet, "/r", func(ctx plugin.Context) {
		// Request() 按设计返回 nil，在 handler 内断言并通过响应体回传结果
		req := ctx.Request()
		ctx.JSON(http.StatusOK, gin.H{"nil": req == nil})
	})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/r", nil))
	require.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `{"nil":true}`, w.Body.String(), "Request() 应返回 nil")
}

// TestContext_GinContext GinContext 返回底层 *gin.Context（同一指针）
func TestContext_GinContext(t *testing.T) {
	r, _ := newTestRouterWith(t, http.MethodGet, "/g", func(ctx plugin.Context) {
		// GinContext 为 *Context 特有方法（非 plugin.Context 接口方法），需类型断言后调用
		gc := ctx.(*Context).GinContext()
		ctx.JSON(http.StatusOK, gin.H{
			"path": gc.Request.URL.Path,
			"ok":   gc != nil,
		})
	})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/g", nil))
	require.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `{"path":"/api/g","ok":true}`, w.Body.String())
}

// TestNewContext NewContext 从 *gin.Context 构造 *Context，二者 GinContext 指向同一实例
func TestNewContext(t *testing.T) {
	r := NewRouter()
	var raw *gin.Context
	r.GET("/n", func(c *gin.Context) {
		raw = c
		c.Status(http.StatusOK)
	})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/n", nil))
	require.NotNil(t, raw)

	ctx := NewContext(raw)
	require.NotNil(t, ctx)
	require.Same(t, raw, ctx.GinContext(), "NewContext 后 GinContext 应返回同一 gin.Context 指针")
}

// ============================ Context.FormFile / fileHeader ============================

// newMultipartBody 构造 multipart/form-data 请求体，包含单个 "file" 字段，返回 body 与 content-type
func newMultipartBody(t *testing.T, fieldName, filename, content string) (*bytes.Buffer, string) {
	t.Helper()
	body := &bytes.Buffer{}
	mw := multipart.NewWriter(body)
	fw, err := mw.CreateFormFile(fieldName, filename)
	require.NoError(t, err)
	_, err = fw.Write([]byte(content))
	require.NoError(t, err)
	require.NoError(t, mw.Close())
	return body, mw.FormDataContentType()
}

// TestContext_FormFile FormFile 读取上传文件，返回的 fileHeader 暴露 Filename/Size/Open
func TestContext_FormFile(t *testing.T) {
	r, _ := newTestRouterWith(t, http.MethodPost, "/up", func(ctx plugin.Context) {
		fh, err := ctx.FormFile()
		require.NoError(t, err)
		require.NotNil(t, fh)
		// 验证 fileHeader 适配器三方法
		assert.Equal(t, "a.txt", fh.Filename())
		assert.Equal(t, int64(len("hello-content")), fh.Size())
		rc, err := fh.Open()
		require.NoError(t, err)
		buf := make([]byte, len("hello-content"))
		n, err := rc.Read(buf)
		require.NoError(t, err)
		assert.Equal(t, len("hello-content"), n)
		assert.Equal(t, "hello-content", string(buf))
		require.NoError(t, rc.Close())
		ctx.JSON(http.StatusOK, gin.H{"name": fh.Filename(), "size": fh.Size()})
	})

	body, ct := newMultipartBody(t, "file", "a.txt", "hello-content")
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/up", body)
	req.Header.Set("Content-Type", ct)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `{"name":"a.txt","size":13}`, w.Body.String())
}

// TestContext_FormFile_Error 请求不含 "file" 字段时 FormFile 返回错误
func TestContext_FormFile_Error(t *testing.T) {
	r, _ := newTestRouterWith(t, http.MethodPost, "/up", func(ctx plugin.Context) {
		_, err := ctx.FormFile()
		require.Error(t, err)
		ctx.Status(http.StatusBadRequest)
	})
	// 构造一个 multipart 请求但字段名不是 "file"
	body := &bytes.Buffer{}
	mw := multipart.NewWriter(body)
	require.NoError(t, mw.WriteField("other", "x"))
	require.NoError(t, mw.Close())
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/up", body)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusBadRequest, w.Code)
}

// TestFileHeader_Open 验证 fileHeader.Open 返回的 ReaderCloser 可读取全部内容
func TestFileHeader_Open(t *testing.T) {
	r, _ := newTestRouterWith(t, http.MethodPost, "/up", func(ctx plugin.Context) {
		fh, err := ctx.FormFile()
		require.NoError(t, err)
		rc, err := fh.Open()
		require.NoError(t, err)
		defer rc.Close()
		// 读取全部
		buf := new(bytes.Buffer)
		_, err = buf.ReadFrom(rc)
		require.NoError(t, err)
		ctx.JSON(http.StatusOK, gin.H{"content": buf.String()})
	})
	body, ct := newMultipartBody(t, "file", "b.txt", "ABCDEF")
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/up", body)
	req.Header.Set("Content-Type", ct)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `{"content":"ABCDEF"}`, w.Body.String())
}

// ============================ 接口编译期断言 ============================

// TestInterfaceAssertions 编译期断言：RouterGroup 实现 plugin.RouterGroup，
// Context 实现 plugin.Context，fileHeader 实现 plugin.FileHeader，
// responseWriter 实现 plugin.ResponseWriter。
func TestInterfaceAssertions(t *testing.T) {
	var _ plugin.RouterGroup = (*RouterGroup)(nil)
	var _ plugin.Context = (*Context)(nil)
	var _ plugin.FileHeader = (*fileHeader)(nil)
	var _ plugin.ResponseWriter = (*responseWriter)(nil)
	// http.Handler 断言（Router）
	var _ http.Handler = (*Router)(nil)
	t.Log("所有接口编译期断言通过")
}

// ============================ fakeContext（仅用于 WrapGin 非 *Context 分支） ============================

// fakeContext 一个不作为 *router.Context 的 plugin.Context 实现，
// 仅供 TestWrapGin_NonContextContext 触发 WrapGin 的类型断言失败早返回分支。
type fakeContext struct{}

func (fakeContext) JSON(int, interface{})                              {}
func (fakeContext) Param(string) string                                { return "" }
func (fakeContext) Query(string) string                                { return "" }
func (fakeContext) DefaultQuery(string, string) string                 { return "" }
func (fakeContext) PostForm(string) string                             { return "" }
func (fakeContext) Bind(interface{}) error                             { return nil }
func (fakeContext) Set(string, interface{})                            {}
func (fakeContext) Get(string) (interface{}, bool)                     { return nil, false }
func (fakeContext) GetHeader(string) string                            { return "" }
func (fakeContext) Status(int)                                         {}
func (fakeContext) Writer() plugin.ResponseWriter                      { return nil }
func (fakeContext) Request() *plugin.Request                           { return nil }
func (fakeContext) FormFile() (plugin.FileHeader, error)               { return nil, nil }

// 编译期断言 fakeContext 实现 plugin.Context
var _ plugin.Context = fakeContext{}
