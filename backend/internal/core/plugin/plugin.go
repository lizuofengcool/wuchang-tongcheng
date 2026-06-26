// Package plugin 插件系统核心定义
// 提供插件接口、注册机制和生命周期管理
package plugin

import (
	"context"
	"io"
	"sync"
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
