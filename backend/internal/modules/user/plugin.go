// Package user 用户模块插件
// 实现用户注册、登录、个人信息管理等业务
package user

import (
	"context"

	"wuchang-tongcheng/internal/core/plugin"
	"wuchang-tongcheng/internal/modules/user/handler"
	"wuchang-tongcheng/internal/modules/user/model"
	"wuchang-tongcheng/internal/modules/user/repository"
	"wuchang-tongcheng/internal/modules/user/service"
	"wuchang-tongcheng/internal/pkg/database"
)

// Plugin 用户模块插件
type Plugin struct {
	name    string
	version string
	handler *handler.Handler
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
	db := database.GetDB()

	// 自动迁移用户表
	if err := db.AutoMigrate(&model.User{}); err != nil {
		return err
	}

	// 初始化依赖链: repository -> service -> handler
	userRepo := repository.NewUserRepository(db)
	userService := service.NewUserService(userRepo)
	p.handler = handler.NewHandler(userService)

	return nil
}

// RegisterRoutes 注册插件路由
func (p *Plugin) RegisterRoutes(router plugin.RouterGroup) {
	// 公开接口（无需登录）
	router.POST("/register", p.handler.Register)
	router.POST("/login", p.handler.Login)

	// 需要登录的接口
	router.GET("/info", p.handler.GetUserInfo)
	router.PUT("/profile", p.handler.UpdateProfile)
	router.PUT("/password", p.handler.ChangePassword)
}

// Close 关闭插件
func (p *Plugin) Close() error {
	return nil
}

// 确保Plugin实现了plugin.Plugin接口
var _ plugin.Plugin = (*Plugin)(nil)
