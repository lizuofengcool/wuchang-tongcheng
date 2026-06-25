// Package user 用户模块插件
// 示例插件模板，所有业务模块插件都应遵循此结构
package user

import (
	"context"

	"wuchang-tongcheng/internal/core/plugin"
)

// Plugin 用户模块插件
type Plugin struct {
	name    string
	version string
}

// NewPlugin 创建用户模块插件
func NewPlugin() *Plugin {
	return &Plugin{
		name:    "user",
		version: "1.0.0",
	}
}

// Name 返回插件名称
func (p *Plugin) Name() string {
	return p.name
}

// Version 返回插件版本号
func (p *Plugin) Version() string {
	return p.version
}

// Init 初始化插件
func (p *Plugin) Init(ctx context.Context) error {
	// TODO: 初始化用户模块
	// - 初始化数据库表
	// - 初始化缓存
	// - 注册消息队列消费者
	// - 初始化定时任务

	return nil
}

// RegisterRoutes 注册插件路由
func (p *Plugin) RegisterRoutes(router plugin.RouterGroup) {
	// TODO: 注册用户模块路由
	// 示例：
	// router.GET("/info", p.GetUserInfo)
	// router.POST("/login", p.Login)
	// router.POST("/register", p.Register)
}

// Close 关闭插件
func (p *Plugin) Close() error {
	// TODO: 清理资源
	// - 关闭数据库连接
	// - 关闭消息队列连接
	// - 停止定时任务

	return nil
}

// 确保Plugin实现了plugin.Plugin接口
var _ plugin.Plugin = (*Plugin)(nil)
