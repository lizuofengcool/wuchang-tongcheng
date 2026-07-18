// Package plugin 插件系统核心定义
// 提供插件接口、注册机制和生命周期管理
package plugin

import (
	"context"
	"io"
	"sync"

	"gorm.io/gorm"
)

// Plugin 插件接口，所有业务模块插件必须实现此接口
type Plugin interface {
	// Name 返回插件名称，必须唯一
	Name() string
	// Version 返回插件版本号
	Version() string
	// Init 初始化插件，在服务启动时调用
	Init(ctx context.Context) error
	// RegisterRoutes 注册插件的路由
	RegisterRoutes(router RouterGroup)
	// Close 关闭插件，在服务停止时调用
	Close() error
}

// Pluggable 可插拔插件接口（可选实现，用于模块管理）
// 实现此接口的模块可通过模块管理界面进行安装/卸载
type Pluggable interface {
	// Description 模块描述
	Description() string
	// Category 模块分类（如：系统、业务、工具）
	Category() string
	// Dependencies 依赖的其他模块名称
	Dependencies() []string
	// Install 安装模块（建表、初始化数据等）
	Install(db *gorm.DB) error
	// Uninstall 卸载模块（删表、清理数据等）
	Uninstall(db *gorm.DB) error
}

// RouterGroup 路由组接口，插件通过此接口注册路由
type RouterGroup interface {
	// Group 创建子路由组
	Group(relativePath string, handlers ...HandlerFunc) RouterGroup
	// GET 注册GET请求
	GET(relativePath string, handlers ...HandlerFunc)
	// POST 注册POST请求
	POST(relativePath string, handlers ...HandlerFunc)
	// PUT 注册PUT请求
	PUT(relativePath string, handlers ...HandlerFunc)
	// DELETE 注册DELETE请求
	DELETE(relativePath string, handlers ...HandlerFunc)
	// PATCH 注册PATCH请求
	PATCH(relativePath string, handlers ...HandlerFunc)
}

// HandlerFunc 处理函数类型定义
type HandlerFunc func(ctx Context)

// Context 上下文接口，抽象HTTP请求上下文
type Context interface {
	// JSON 返回JSON响应
	JSON(code int, obj interface{})
	// Param 获取URL参数
	Param(key string) string
	// Query 获取Query参数
	Query(key string) string
	// DefaultQuery 获取Query参数，带默认值
	DefaultQuery(key, defaultValue string) string
	// PostForm 获取表单参数
	PostForm(key string) string
	// Bind 绑定请求数据
	Bind(obj interface{}) error
	// Set 设置上下文值
	Set(key string, value interface{})
	// Get 获取上下文值
	Get(key string) (interface{}, bool)
	// GetHeader 获取请求头
	GetHeader(key string) string
	// Status 设置响应状态码
	Status(code int)
	// Writer 获取响应写入器
	Writer() ResponseWriter
	// Request 获取请求对象
	Request() *Request
	// FormFile 获取上传的文件（multipart/form-data）
	FormFile() (FileHeader, error)
}

// FileHeader 上传文件信息
type FileHeader interface {
	// Filename 原始文件名
	Filename() string
	// Size 文件大小（字节）
	Size() int64
	// Open 打开文件读取流，用完需Close
	Open() (io.ReadCloser, error)
}

// ResponseWriter 响应写入器接口
type ResponseWriter interface {
	Write([]byte) (int, error)
	WriteHeader(statusCode int)
	Header() map[string][]string
}

// Request 请求对象接口
type Request interface {
	Method() string
	URL() string
	Header() map[string][]string
}

// Manager 插件管理器
type Manager struct {
	mu      sync.RWMutex
	plugins map[string]Plugin
	order   []string // 保持插件注册顺序
}

var (
	instance *Manager
	once     sync.Once
)

// GetManager 获取插件管理器单例
func GetManager() *Manager {
	once.Do(func() {
		instance = &Manager{
			plugins: make(map[string]Plugin),
			order:   make([]string, 0),
		}
	})
	return instance
}

// Register 注册插件
func (m *Manager) Register(plugin Plugin) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	name := plugin.Name()
	if _, exists := m.plugins[name]; exists {
		return ErrPluginAlreadyExists
	}

	m.plugins[name] = plugin
	m.order = append(m.order, name)
	return nil
}

// Get 获取指定名称的插件
func (m *Manager) Get(name string) (Plugin, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	p, ok := m.plugins[name]
	return p, ok
}

// List 获取所有已注册的插件列表（按注册顺序）
func (m *Manager) List() []Plugin {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]Plugin, 0, len(m.order))
	for _, name := range m.order {
		if p, ok := m.plugins[name]; ok {
			result = append(result, p)
		}
	}
	return result
}

// InitAll 初始化所有插件
func (m *Manager) InitAll(ctx context.Context) error {
	plugins := m.List()
	for _, p := range plugins {
		if err := p.Init(ctx); err != nil {
			return &PluginInitError{
				PluginName: p.Name(),
				Err:        err,
			}
		}
	}
	return nil
}

// RegisterAllRoutes 注册所有插件的路由
func (m *Manager) RegisterAllRoutes(router RouterGroup) {
	plugins := m.List()
	for _, p := range plugins {
		pluginGroup := router.Group("/api/v1/" + p.Name())
		p.RegisterRoutes(pluginGroup)
	}
}

// CloseAll 关闭所有插件
func (m *Manager) CloseAll() error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var firstErr error
	// 逆序关闭
	for i := len(m.order) - 1; i >= 0; i-- {
		name := m.order[i]
		if p, ok := m.plugins[name]; ok {
			if err := p.Close(); err != nil && firstErr == nil {
				firstErr = err
			}
		}
	}
	return firstErr
}

// Unregister 注销插件
func (m *Manager) Unregister(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.plugins[name]; !exists {
		return ErrPluginNotFound
	}

	delete(m.plugins, name)
	// 从顺序列表中移除
	for i, n := range m.order {
		if n == name {
			m.order = append(m.order[:i], m.order[i+1:]...)
			break
		}
	}
	return nil
}

// MustRegister 注册插件（幂等，已存在则跳过）
// 用于 init() 自动注册场景，避免重复注册报错
func (m *Manager) MustRegister(p Plugin) {
	m.mu.Lock()
	defer m.mu.Unlock()
	name := p.Name()
	if _, exists := m.plugins[name]; exists {
		return
	}
	m.plugins[name] = p
	m.order = append(m.order, name)
}

// ListPluggable 返回所有实现了 Pluggable 接口的插件
func (m *Manager) ListPluggable() []Pluggable {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make([]Pluggable, 0)
	for _, name := range m.order {
		if p, ok := m.plugins[name]; ok {
			if pl, ok := p.(Pluggable); ok {
				result = append(result, pl)
			}
		}
	}
	return result
}

// AutoRegister 自动注册插件（幂等，在模块 init() 中调用）
// 用法：在模块的 plugin.go 中添加 init() { AutoRegister(NewPlugin()) }
func AutoRegister(p Plugin) {
	GetManager().MustRegister(p)
}
