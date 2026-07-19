// Package plugin 插件系统核心定义
// 提供插件接口、注册机制和生命周期管理
package plugin

import (
	"context"
	"io"
	"sync"

	"gorm.io/gorm"
)

// PluginMeta 插件元信息，描述模块的展示与分类等元数据
// 用于 modules 模块注册表同步、模块总控面板展示、依赖关系分析
type PluginMeta struct {
	Name         string   // 模块唯一标识（如 ershou）
	DisplayName  string   // 中文展示名（如"同城二手"）
	Category     string   // 分类：system/business/marketing/user/community/middleware
	Description  string   // 模块描述
	Version      string   // 版本号
	Dependencies []string // 依赖的其他模块名
	Icon         string   // 图标 URL 或 icon class
	Author       string   // 作者
	Homepage     string   // 主页 URL
}

// Plugin 插件接口，所有业务模块插件必须实现此接口
type Plugin interface {
	// Name 返回插件名称，必须唯一
	Name() string
	// Version 返回插件版本号
	Version() string
	// Meta 返回插件元信息（展示名、分类、依赖等），用于 modules 注册表同步
	Meta() PluginMeta
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

// metaProvider 内部元信息提供者接口，用于 MetaFromPlugin 的类型断言
// 保留此接口便于外部包装器（proxy/decorator）按需扩展，即便 Plugin 已强制要求 Meta()
type metaProvider interface {
	Meta() PluginMeta
}

// MetaFromPlugin 通过类型断言获取插件的元信息。
// 由于 Plugin 接口已包含 Meta() 方法，常规插件直接调用即可；
// 对于通过反射/代理加载的外部插件（可能未实现 Meta()），返回基于 Name()/Version() 的默认 PluginMeta。
func MetaFromPlugin(p Plugin) PluginMeta {
	if mp, ok := p.(metaProvider); ok {
		return mp.Meta()
	}
	return PluginMeta{
		Name:    p.Name(),
		Version: p.Version(),
	}
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
// 新增 db 参数用于查询 modules 表，跳过 enabled=false 的模块（不注册路由）
// 若 db 为 nil（如单元测试场景），则按原行为注册全部插件路由
func (m *Manager) RegisterAllRoutes(db *gorm.DB, router RouterGroup) {
	plugins := m.List()
	for _, p := range plugins {
		// 查询 modules 表，跳过已禁用的模块
		if db != nil && isModuleDisabled(db, p.Name()) {
			continue
		}
		pluginGroup := router.Group("/api/v1/" + p.Name())
		p.RegisterRoutes(pluginGroup)
	}
}

// isModuleDisabled 查询 modules 表判断模块是否被禁用
// 表不存在或记录不存在视为启用（不阻塞未同步的模块）
func isModuleDisabled(db *gorm.DB, name string) bool {
	type moduleStatus struct {
		Enabled bool
	}
	var status moduleStatus
	err := db.Table("modules").
		Select("enabled").
		Where("name = ?", name).
		Limit(1).
		First(&status).Error
	if err != nil {
		// 记录不存在或表不存在 → 视为启用
		return false
	}
	return !status.Enabled
}

// RegisterRoutesFor 运行时动态注册指定模块的路由（用于启用某模块时调用）
// 调用方应先在 modules 表中将该模块置为 enabled=true，再调用此方法
func (m *Manager) RegisterRoutesFor(db *gorm.DB, name string, router RouterGroup) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	p, ok := m.plugins[name]
	if !ok {
		return ErrPluginNotFound
	}
	pluginGroup := router.Group("/api/v1/" + p.Name())
	p.RegisterRoutes(pluginGroup)
	return nil
}

// UnregisterRoutesFor 运行时动态禁用指定模块的路由
// P0 实现说明：Gin 路由树不支持运行时删除路由，此方法仅返回 nil；
// 实际拦截由 ModuleCheck 中间件在请求层完成（查询 modules 表，禁用模块返回 403）
func (m *Manager) UnregisterRoutesFor(name string) error {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if _, ok := m.plugins[name]; !ok {
		return ErrPluginNotFound
	}
	return nil
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
