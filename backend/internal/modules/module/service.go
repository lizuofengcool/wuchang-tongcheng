// Package module 模块注册表业务服务层
package module

import (
	"context"
	"errors"
	"fmt"
	"time"

	"wuchang-tongcheng/internal/core/plugin"
	redispkg "wuchang-tongcheng/internal/pkg/redis"
)

// 模块开关 Redis 缓存键前缀与 TTL
// 中间件与 service 共享同一套键，启停时由 service 失效缓存
const (
	moduleCacheKeyPrefix = "module:enabled:"
	moduleCacheTTL       = 60 * time.Second
)

// moduleCacheKey 构造模块开关缓存的键
func moduleCacheKey(name string) string {
	return moduleCacheKeyPrefix + name
}

// invalidateModuleCache 失效指定模块的开关缓存（Redis 不可用时 no-op）
func invalidateModuleCache(name string) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_ = redispkg.Del(ctx, moduleCacheKey(name))
}

// Service 模块业务服务接口
type Service interface {
	ListModules() ([]ModuleInfo, error)
	GetModule(name string) (*ModuleInfo, error)
	EnableModule(name string) error
	DisableModule(name string) error
	UpdateModule(name string, req *UpdateModuleRequest) error
	RegisterFromPlugin(p plugin.Plugin) error
	SyncAllFromManager() error
}

type service struct {
	repo Repository
}

// NewService 创建模块服务
func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) ListModules() ([]ModuleInfo, error) {
	modules, err := s.repo.List()
	if err != nil {
		return nil, err
	}
	result := make([]ModuleInfo, 0, len(modules))
	for i := range modules {
		result = append(result, toInfo(&modules[i]))
	}
	return result, nil
}

func (s *service) GetModule(name string) (*ModuleInfo, error) {
	m, err := s.repo.GetByName(name)
	if err != nil {
		return nil, err
	}
	info := toInfo(m)
	return &info, nil
}

func (s *service) EnableModule(name string) error {
	if _, err := s.repo.GetByName(name); err != nil {
		return err
	}
	if err := s.repo.Enable(name); err != nil {
		return err
	}
	// 失效缓存，下次请求重新从 DB 读取
	invalidateModuleCache(name)
	return nil
}

func (s *service) DisableModule(name string) error {
	if _, err := s.repo.GetByName(name); err != nil {
		return err
	}
	if err := s.repo.Disable(name); err != nil {
		return err
	}
	invalidateModuleCache(name)
	return nil
}

func (s *service) UpdateModule(name string, req *UpdateModuleRequest) error {
	m, err := s.repo.GetByName(name)
	if err != nil {
		return err
	}
	if req.DisplayName != nil {
		m.DisplayName = *req.DisplayName
	}
	if req.Description != nil {
		m.Description = *req.Description
	}
	if req.Icon != nil {
		m.Icon = *req.Icon
	}
	if req.Author != nil {
		m.Author = *req.Author
	}
	if req.Homepage != nil {
		m.Homepage = *req.Homepage
	}
	return s.repo.Update(m)
}

// RegisterFromPlugin 将单个插件的元信息同步到 modules 表
// 已存在则更新元信息（不修改 enabled），不存在则新增（默认 enabled=true）
func (s *service) RegisterFromPlugin(p plugin.Plugin) error {
	meta := plugin.MetaFromPlugin(p)
	return s.syncOne(meta)
}

// SyncAllFromManager 从 plugin.GetManager() 读取所有已注册插件，同步到 modules 表
// 启动时调用：新增或更新元信息，不修改 enabled 状态
func (s *service) SyncAllFromManager() error {
	manager := plugin.GetManager()
	plugins := manager.List()
	for _, p := range plugins {
		meta := plugin.MetaFromPlugin(p)
		if err := s.syncOne(meta); err != nil {
			return fmt.Errorf("sync module %s failed: %w", meta.Name, err)
		}
	}
	return nil
}

// syncOne 同步单个插件元信息到 modules 表
func (s *service) syncOne(meta plugin.PluginMeta) error {
	existing, err := s.repo.GetByName(meta.Name)
	if err != nil {
		if !errors.Is(err, ErrModuleNotFound) {
			return err
		}
		// 新增模块记录，默认启用
		m := &Module{
			Name:         meta.Name,
			DisplayName:  meta.DisplayName,
			Category:     meta.Category,
			Description:  meta.Description,
			Version:      meta.Version,
			Dependencies: toJSON(meta.Dependencies),
			Icon:         meta.Icon,
			Author:       meta.Author,
			Homepage:     meta.Homepage,
			Enabled:      true,
		}
		return s.repo.Create(m)
	}
	// 已存在：更新元信息，保留 enabled 状态
	existing.DisplayName = meta.DisplayName
	existing.Category = meta.Category
	existing.Description = meta.Description
	existing.Version = meta.Version
	existing.Dependencies = toJSON(meta.Dependencies)
	existing.Icon = meta.Icon
	existing.Author = meta.Author
	existing.Homepage = meta.Homepage
	return s.repo.Update(existing)
}
