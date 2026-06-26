// Package file 文件存储模块插件
// 提供文件上传功能，支持本地存储，预留MinIO/七牛云
package file

import (
	"context"

	"wuchang-tongcheng/internal/core/plugin"
	"wuchang-tongcheng/internal/modules/file/handler"
	"wuchang-tongcheng/internal/modules/file/model"
	"wuchang-tongcheng/internal/modules/file/service"
	"wuchang-tongcheng/internal/pkg/config"
	"wuchang-tongcheng/internal/pkg/database"
	"wuchang-tongcheng/internal/pkg/storage"
)

// Plugin 文件模块插件
type Plugin struct {
	name    string
	version string
	handler *handler.Handler
}

// NewPlugin 创建文件模块插件
func NewPlugin() *Plugin {
	return &Plugin{name: "file", version: "1.0.0"}
}

// Name 返回插件名称
func (p *Plugin) Name() string { return p.name }

// Version 返回插件版本号
func (p *Plugin) Version() string { return p.version }

// Init 初始化插件
func (p *Plugin) Init(ctx context.Context) error {
	db := database.GetDB()

	// 自动迁移文件表
	if err := db.AutoMigrate(&model.FileUpload{}); err != nil {
		return err
	}

	// 初始化存储（使用全局配置）
	cfg := config.Get()
	if err := storage.Init(&cfg.Storage); err != nil {
		// 存储初始化失败不阻塞启动，降级为本地存储
		_ = storage.Init(nil)
	}

	// 初始化依赖链
	fileService := service.NewFileService(db)
	p.handler = handler.NewHandler(fileService)

	return nil
}

// RegisterRoutes 注册插件路由
func (p *Plugin) RegisterRoutes(router plugin.RouterGroup) {
	router.POST("/upload", p.handler.Upload)
}

// Close 关闭插件
func (p *Plugin) Close() error { return nil }

// 确保Plugin实现了plugin.Plugin接口
var _ plugin.Plugin = (*Plugin)(nil)
