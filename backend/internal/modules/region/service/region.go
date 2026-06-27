// Package service 地区业务逻辑层
package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"wuchang-tongcheng/internal/modules/region/dto"
	"wuchang-tongcheng/internal/modules/region/model"
	"wuchang-tongcheng/internal/modules/region/repository"
	rediscache "wuchang-tongcheng/internal/pkg/redis"

	"gorm.io/gorm"
)

var (
	ErrRegionNotFound      = errors.New("地区不存在")
	ErrRegionCodeExists    = errors.New("地区编码已存在")
	ErrRegionHasChildren   = errors.New("该地区存在子地区，无法删除")
	ErrRegionMaxLevel      = errors.New("地区层级已达上限（最多3级：省/市/区县）")
	ErrRegionParentInvalid = errors.New("父地区不存在")
)

// MaxRegionLevel 地区最大层级（1省 2市 3区县）
const MaxRegionLevel = 3

// 缓存键前缀与 TTL（地区数据变更极少，TTL 30min，写操作按前缀失效整组）
const (
	regionCachePrefix = "cache:region:"
	regionCacheTTL    = 30 * time.Minute
)

// regionCacheKey 各读路径的缓存键
func regionCacheKeyTree() string          { return regionCachePrefix + "tree" }
func regionCacheKeyByID(id uint) string   { return fmt.Sprintf(regionCachePrefix+"id:%d", id) }
func regionCacheKeyByParent(id uint) string {
	return fmt.Sprintf(regionCachePrefix+"parent:%d", id)
}

// invalidateRegionCache 失效整组地区缓存（SCAN+DEL，Redis 不可用时 no-op）
func invalidateRegionCache() {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_ = rediscache.DelByPrefix(ctx, regionCachePrefix)
}

// RegionService 地区业务逻辑接口
type RegionService interface {
	Create(req *dto.CreateRegionRequest) (*dto.RegionInfo, error)
	Update(id uint, req *dto.UpdateRegionRequest) error
	Delete(id uint) error
	GetByID(id uint) (*dto.RegionInfo, error)
	GetByParentID(parentID uint) ([]dto.RegionInfo, error)
	GetTree() ([]dto.RegionTree, error)
}

type regionService struct {
	regionRepo repository.RegionRepository
}

// NewRegionService 创建地区服务
func NewRegionService(regionRepo repository.RegionRepository) RegionService {
	return &regionService{regionRepo: regionRepo}
}

func toRegionInfo(region *model.Region) *dto.RegionInfo {
	return &dto.RegionInfo{
		ID:       region.ID,
		Name:     region.Name,
		Code:     region.Code,
		ParentID: region.ParentID,
		Level:    region.Level,
		Sort:     region.Sort,
		Status:   region.Status,
	}
}

// Create 创建地区
// Level 自动根据 ParentID 计算：ParentID=0 时为 1（省），否则为 父地区 Level+1
// 最大层级受 MaxRegionLevel 限制（省/市/区县 3 级），超过返回错误
func (s *regionService) Create(req *dto.CreateRegionRequest) (*dto.RegionInfo, error) {
	// 检查编码是否重复
	if _, err := s.regionRepo.FindByCode(req.Code); err == nil {
		return nil, ErrRegionCodeExists
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	status := req.Status
	if status == 0 {
		status = 1 // 默认正常
	}

	// 根据父地区计算层级
	level := 1
	if req.ParentID > 0 {
		parent, err := s.regionRepo.FindByID(req.ParentID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, ErrRegionParentInvalid
			}
			return nil, err
		}
		level = parent.Level + 1
		if level > MaxRegionLevel {
			return nil, ErrRegionMaxLevel
		}
	}

	region := &model.Region{
		Name:     req.Name,
		Code:     req.Code,
		ParentID: req.ParentID,
		Level:    level,
		Sort:     req.Sort,
		Status:   status,
	}

	if err := s.regionRepo.Create(region); err != nil {
		return nil, err
	}
	invalidateRegionCache()
	return toRegionInfo(region), nil
}

// Update 更新地区
func (s *regionService) Update(id uint, req *dto.UpdateRegionRequest) error {
	if _, err := s.regionRepo.FindByID(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrRegionNotFound
		}
		return err
	}

	fields := map[string]interface{}{}
	if req.Name != "" {
		fields["name"] = req.Name
	}
	fields["sort"] = req.Sort
	if req.Status == 0 || req.Status == 1 {
		fields["status"] = req.Status
	}
	if err := s.regionRepo.UpdateFields(id, fields); err != nil {
		return err
	}
	invalidateRegionCache()
	return nil
}

// Delete 删除地区
func (s *regionService) Delete(id uint) error {
	if _, err := s.regionRepo.FindByID(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrRegionNotFound
		}
		return err
	}

	// 检查是否有子地区
	children, err := s.regionRepo.FindByParentID(id)
	if err != nil {
		return err
	}
	if len(children) > 0 {
		return ErrRegionHasChildren
	}

	if err := s.regionRepo.Delete(id); err != nil {
		return err
	}
	invalidateRegionCache()
	return nil
}

// GetByID 根据ID获取地区（cache-aside：Redis 命中直接返回，miss 回源 DB 并回填）
func (s *regionService) GetByID(id uint) (*dto.RegionInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	var cached dto.RegionInfo
	if hit, _ := rediscache.GetJSON(ctx, regionCacheKeyByID(id), &cached); hit {
		return &cached, nil
	}

	region, err := s.regionRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrRegionNotFound
		}
		return nil, err
	}
	info := toRegionInfo(region)
	_ = rediscache.SetJSON(ctx, regionCacheKeyByID(id), info, regionCacheTTL)
	return info, nil
}

// GetByParentID 根据父级ID获取子地区列表（cache-aside）
func (s *regionService) GetByParentID(parentID uint) ([]dto.RegionInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	var cached []dto.RegionInfo
	if hit, _ := rediscache.GetJSON(ctx, regionCacheKeyByParent(parentID), &cached); hit {
		return cached, nil
	}

	regions, err := s.regionRepo.FindByParentID(parentID)
	if err != nil {
		return nil, err
	}
	result := make([]dto.RegionInfo, 0, len(regions))
	for i := range regions {
		result = append(result, *toRegionInfo(&regions[i]))
	}
	_ = rediscache.SetJSON(ctx, regionCacheKeyByParent(parentID), result, regionCacheTTL)
	return result, nil
}

// GetTree 获取地区树形结构（cache-aside：地区树变更极少，热点读）
func (s *regionService) GetTree() ([]dto.RegionTree, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	var cached []dto.RegionTree
	if hit, _ := rediscache.GetJSON(ctx, regionCacheKeyTree(), &cached); hit {
		return cached, nil
	}

	all, err := s.regionRepo.FindAll()
	if err != nil {
		return nil, err
	}
	// 构建ID到子节点映射
	childrenMap := make(map[uint][]model.Region)
	for _, r := range all {
		childrenMap[r.ParentID] = append(childrenMap[r.ParentID], r)
	}

	// 递归构建树
	var build func(parentID uint) []dto.RegionTree
	build = func(parentID uint) []dto.RegionTree {
		children := childrenMap[parentID]
		trees := make([]dto.RegionTree, 0, len(children))
		for i := range children {
			r := children[i]
			tree := dto.RegionTree{
				RegionInfo: *toRegionInfo(&r),
				Children:   build(r.ID),
			}
			trees = append(trees, tree)
		}
		return trees
	}

	tree := build(0)
	_ = rediscache.SetJSON(ctx, regionCacheKeyTree(), tree, regionCacheTTL)
	return tree, nil
}
