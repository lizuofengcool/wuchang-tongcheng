// Package router 路由封装
// 基于Gin的路由封装，实现插件系统的路由接口
package router

import (
	"io"
	"mime/multipart"
	"net/http"

	"wuchang-tongcheng/internal/core/plugin"

	"github.com/gin-gonic/gin"
)

// Router 路由引擎
type Router struct {
	engine *gin.Engine
}

// NewRouter 创建新的路由引擎
func NewRouter() *Router {
	engine := gin.New()
	return &Router{
		engine: engine,
	}
}

// Engine 获取Gin引擎
func (r *Router) Engine() *gin.Engine {
	return r.engine
}

// Group 创建根路由组
func (r *Router) Group(relativePath string, handlers ...gin.HandlerFunc) *RouterGroup {
	return &RouterGroup{
		group: r.engine.Group(relativePath, handlers...),
	}
}

// Use 添加全局中间件
func (r *Router) Use(middleware ...gin.HandlerFunc) {
	r.engine.Use(middleware...)
}

// GET 注册GET请求
func (r *Router) GET(relativePath string, handlers ...gin.HandlerFunc) {
	r.engine.GET(relativePath, handlers...)
}

// POST 注册POST请求
func (r *Router) POST(relativePath string, handlers ...gin.HandlerFunc) {
	r.engine.POST(relativePath, handlers...)
}

// PUT 注册PUT请求
func (r *Router) PUT(relativePath string, handlers ...gin.HandlerFunc) {
	r.engine.PUT(relativePath, handlers...)
}

// DELETE 注册DELETE请求
func (r *Router) DELETE(relativePath string, handlers ...gin.HandlerFunc) {
	r.engine.DELETE(relativePath, handlers...)
}

// PATCH 注册PATCH请求
func (r *Router) PATCH(relativePath string, handlers ...gin.HandlerFunc) {
	r.engine.PATCH(relativePath, handlers...)
}

// Any 注册所有HTTP方法
func (r *Router) Any(relativePath string, handlers ...gin.HandlerFunc) {
	r.engine.Any(relativePath, handlers...)
}

// GroupFunc 返回原生 gin.RouterGroup 用于更灵活的路由注册
func (r *Router) GroupFunc(relativePath string, handlers ...gin.HandlerFunc) *gin.RouterGroup {
	return r.engine.Group(relativePath, handlers...)
}

// Run 启动HTTP服务
func (r *Router) Run(addr string) error {
	return r.engine.Run(addr)
}

// ServeHTTP 实现http.Handler接口
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.engine.ServeHTTP(w, req)
}

// RouterGroup 路由组适配器，实现plugin.RouterGroup接口
type RouterGroup struct {
	group *gin.RouterGroup
}

// 确保RouterGroup实现了plugin.RouterGroup接口
var _ plugin.RouterGroup = (*RouterGroup)(nil)

// Group 创建子路由组
func (rg *RouterGroup) Group(relativePath string, handlers ...plugin.HandlerFunc) plugin.RouterGroup {
	ginHandlers := convertHandlers(handlers...)
	return &RouterGroup{
		group: rg.group.Group(relativePath, ginHandlers...),
	}
}

// GET 注册GET请求
func (rg *RouterGroup) GET(relativePath string, handlers ...plugin.HandlerFunc) {
	ginHandlers := convertHandlers(handlers...)
	rg.group.GET(relativePath, ginHandlers...)
}

// POST 注册POST请求
func (rg *RouterGroup) POST(relativePath string, handlers ...plugin.HandlerFunc) {
	ginHandlers := convertHandlers(handlers...)
	rg.group.POST(relativePath, ginHandlers...)
}

// PUT 注册PUT请求
func (rg *RouterGroup) PUT(relativePath string, handlers ...plugin.HandlerFunc) {
	ginHandlers := convertHandlers(handlers...)
	rg.group.PUT(relativePath, ginHandlers...)
}

// DELETE 注册DELETE请求
func (rg *RouterGroup) DELETE(relativePath string, handlers ...plugin.HandlerFunc) {
	ginHandlers := convertHandlers(handlers...)
	rg.group.DELETE(relativePath, ginHandlers...)
}

// PATCH 注册PATCH请求
func (rg *RouterGroup) PATCH(relativePath string, handlers ...plugin.HandlerFunc) {
	ginHandlers := convertHandlers(handlers...)
	rg.group.PATCH(relativePath, ginHandlers...)
}

// convertHandlers 将plugin.HandlerFunc转换为gin.HandlerFunc
func convertHandlers(handlers ...plugin.HandlerFunc) []gin.HandlerFunc {
	result := make([]gin.HandlerFunc, len(handlers))
	for i, h := range handlers {
		handler := h
		result[i] = func(c *gin.Context) {
			ctx := &Context{c: c}
			handler(ctx)
		}
	}
	return result
}

// WrapGin 将 gin.HandlerFunc 中间件转换为 plugin.HandlerFunc，便于插件路由复用现有 gin 中间件
// 用法：router.POST("/admin/users", router.WrapGin(middleware.AuthRequired()), p.handler.Create)
func WrapGin(mw gin.HandlerFunc) plugin.HandlerFunc {
	return func(ctx plugin.Context) {
		c, ok := ctx.(*Context)
		if !ok {
			return
		}
		mw(c.GinContext())
	}
}

// Context 上下文适配器，实现plugin.Context接口
type Context struct {
	c *gin.Context
}

// 确保Context实现了plugin.Context接口
var _ plugin.Context = (*Context)(nil)

// JSON 返回JSON响应
func (ctx *Context) JSON(code int, obj interface{}) {
	ctx.c.JSON(code, obj)
}

// Param 获取URL参数
func (ctx *Context) Param(key string) string {
	return ctx.c.Param(key)
}

// Query 获取Query参数
func (ctx *Context) Query(key string) string {
	return ctx.c.Query(key)
}

// PostForm 获取表单参数
func (ctx *Context) PostForm(key string) string {
	return ctx.c.PostForm(key)
}

// Bind 绑定请求数据
func (ctx *Context) Bind(obj interface{}) error {
	return ctx.c.ShouldBind(obj)
}

// Set 设置上下文值
func (ctx *Context) Set(key string, value interface{}) {
	ctx.c.Set(key, value)
}

// Get 获取上下文值
func (ctx *Context) Get(key string) (interface{}, bool) {
	return ctx.c.Get(key)
}

// GetHeader 获取请求头
func (ctx *Context) GetHeader(key string) string {
	return ctx.c.GetHeader(key)
}

// Status 设置响应状态码
func (ctx *Context) Status(code int) {
	ctx.c.Status(code)
}

// Writer 获取响应写入器
func (ctx *Context) Writer() plugin.ResponseWriter {
	return &responseWriter{w: ctx.c.Writer}
}

// Request 获取请求对象
func (ctx *Context) Request() *plugin.Request {
	// 返回nil，因为Gin的Request是具体类型
	// 实际使用时可以通过GinContext()获取原始上下文
	return nil
}

// FormFile 获取上传的文件
func (ctx *Context) FormFile() (plugin.FileHeader, error) {
	fh, err := ctx.c.FormFile("file")
	if err != nil {
		return nil, err
	}
	return &fileHeader{fh: fh}, nil
}

// GinContext 获取原始Gin上下文
func (ctx *Context) GinContext() *gin.Context {
	return ctx.c
}

// fileHeader 适配 gin 的 *multipart.FileHeader 到 plugin.FileHeader 接口
type fileHeader struct {
	fh *multipart.FileHeader
}

// Filename 原始文件名
func (h *fileHeader) Filename() string { return h.fh.Filename }

// Size 文件大小
func (h *fileHeader) Size() int64 { return h.fh.Size }

// Open 打开文件
func (h *fileHeader) Open() (io.ReadCloser, error) {
	f, err := h.fh.Open()
	if err != nil {
		return nil, err
	}
	return f, nil
}

// responseWriter 响应写入器适配器
type responseWriter struct {
	w gin.ResponseWriter
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	return rw.w.Write(b)
}

func (rw *responseWriter) WriteHeader(statusCode int) {
	rw.w.WriteHeader(statusCode)
}

func (rw *responseWriter) Header() map[string][]string {
	return rw.w.Header()
}

// NewContext 从Gin上下文创建插件上下文
func NewContext(c *gin.Context) *Context {
	return &Context{c: c}
}
