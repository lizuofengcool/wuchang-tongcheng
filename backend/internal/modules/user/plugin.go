// Package user 用户模块插件
// 实现用户注册、登录、个人信息管理等业务
package user

import (
	"context"

	"wuchang-tongcheng/internal/core/middleware"
	"wuchang-tongcheng/internal/core/plugin"
	coreRouter "wuchang-tongcheng/internal/core/router"
	"wuchang-tongcheng/internal/modules/user/handler"
	"wuchang-tongcheng/internal/modules/user/model"
	"wuchang-tongcheng/internal/modules/user/repository"
	"wuchang-tongcheng/internal/modules/user/service"
	"wuchang-tongcheng/internal/pkg/config"
	"wuchang-tongcheng/internal/pkg/database"
	oauthpkg "wuchang-tongcheng/internal/pkg/oauth"
	"wuchang-tongcheng/internal/pkg/sms"
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

// Meta 返回插件元信息
func (p *Plugin) Meta() plugin.PluginMeta {
	return plugin.PluginMeta{
		Name:         "user",
		DisplayName:  "用户中心",
		Category:     "user",
		Description:  "实现用户注册、登录、个人信息管理等业务",
		Version:      p.version,
		Dependencies: []string{},
		Author:       "wuchang",
	}
}

// Init 初始化插件
func (p *Plugin) Init(ctx context.Context) error {
	db := database.GetDB()

	// 自动迁移用户表 + 第三方账号绑定表
	if err := db.AutoMigrate(&model.User{}, &model.UserOAuth{}); err != nil {
		return err
	}

	// 初始化依赖链: repository -> service -> handler
	userRepo := repository.NewUserRepository(db)
	oauthRepo := repository.NewUserOAuthRepository(db)
	// 短信验证码服务（按全局 SMS 配置创建；provider 未配置时走 mock，Redis 不可用降级内存存储）
	smsSvc := sms.NewService(&config.Get().SMS)
	// 第三方 OAuth 登录服务（按全局 OAuth 配置创建；未配置时所有 provider 不注册，OAuthLogin 返回未启用）
	oauthSvc := oauthpkg.NewService(&config.Get().OAuth)
	userService := service.NewUserService(userRepo, smsSvc, oauthRepo, oauthSvc)
	p.handler = handler.NewHandler(userService)

	return nil
}

// RegisterRoutes 注册插件路由
func (p *Plugin) RegisterRoutes(router plugin.RouterGroup) {
	// 公开接口（无需登录）
	// 登录限流：单 IP 每分钟最多 5 次，防止暴力破解
	loginLimiter := coreRouter.WrapGin(middleware.RateLimit(5, 60, "login"))
	// 短信验证码限流：单 IP 每分钟最多 5 次，防止短信轰炸
	smsLimiter := coreRouter.WrapGin(middleware.RateLimit(5, 60, "sms"))
	router.POST("/register", p.handler.Register)
	router.POST("/login", loginLimiter, p.handler.Login)
	// 短信验证码登录：发送验证码 + 验证码登录
	router.POST("/sms/code", smsLimiter, p.handler.SendSMSCode)
	router.POST("/login/sms", loginLimiter, p.handler.SMSLogin)
	// 第三方 OAuth 登录：前端从授权回调拿到 code 后换取 JWT
	router.POST("/login/oauth/:provider", loginLimiter, p.handler.OAuthLogin)

	// 需要登录的接口
	auth := coreRouter.WrapGin(middleware.AuthRequired())
	router.GET("/info", auth, p.handler.GetUserInfo)
	router.PUT("/profile", auth, p.handler.UpdateProfile)
	router.PUT("/password", auth, p.handler.ChangePassword)

	// 管理后台接口（需要登录 + 权限）
	admin := router.Group("/admin")
	admin.GET("/users", coreRouter.WrapGin(middleware.RequirePermission("user:read")), p.handler.ListUsers)
	admin.POST("/users", coreRouter.WrapGin(middleware.RequirePermission("user:create")), p.handler.AdminCreateUser)
	admin.GET("/users/:id", coreRouter.WrapGin(middleware.RequirePermission("user:read")), p.handler.AdminGetUser)
	admin.PUT("/users/:id", coreRouter.WrapGin(middleware.RequirePermission("user:update")), p.handler.AdminUpdateUser)
	admin.PUT("/users/:id/status", coreRouter.WrapGin(middleware.RequirePermission("user:update")), p.handler.UpdateUserStatus)
	admin.PUT("/users/:id/password", coreRouter.WrapGin(middleware.RequirePermission("user:reset_password")), p.handler.ResetPassword)
	admin.DELETE("/users/:id", coreRouter.WrapGin(middleware.RequirePermission("user:delete")), p.handler.DeleteUser)
}

// Close 关闭插件
func (p *Plugin) Close() error {
	return nil
}

// init 自动注册插件（幂等）
func init() {
	plugin.AutoRegister(NewPlugin())
}

// 确保Plugin实现了plugin.Plugin接口
var _ plugin.Plugin = (*Plugin)(nil)
