// Package service 分类信息业务逻辑层
package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"wuchang-tongcheng/internal/modules/category/dto"
	"wuchang-tongcheng/internal/modules/category/model"
	"wuchang-tongcheng/internal/modules/category/repository"
	rediscache "wuchang-tongcheng/internal/pkg/redis"

	"gorm.io/gorm"
)

var (
	ErrCategoryNotFound    = errors.New("分类不存在")
	ErrCategoryHasChildren  = errors.New("该分类存在子分类，无法删除")
	ErrCategoryMaxLevel    = errors.New("分类层级已达上限（最多3级）")
	ErrCategoryParentInvalid = errors.New("父分类不存在")
)

// MaxCategoryLevel 分类最大层级
const MaxCategoryLevel = 3

// 缓存键前缀与 TTL（分类数据变更少，TTL 30min，写操作按前缀失效整组）
const (
	categoryCachePrefix = "cache:category:"
	categoryCacheTTL    = 30 * time.Minute
)

func categoryCacheKeyTree(regionID uint) string {
	return fmt.Sprintf(categoryCachePrefix+"tree:%d", regionID)
}
func categoryCacheKeyByID(id uint) string {
	return fmt.Sprintf(categoryCachePrefix+"id:%d", id)
}
func categoryCacheKeyByParent(parentID, regionID uint) string {
	return fmt.Sprintf(categoryCachePrefix+"parent:%d:%d", parentID, regionID)
}

// invalidateCategoryCache 失效整组分类缓存（SCAN+DEL，Redis 不可用时 no-op）
func invalidateCategoryCache() {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_ = rediscache.DelByPrefix(ctx, categoryCachePrefix)
}

// CategoryService 分类业务逻辑接口
type CategoryService interface {
	Create(regionID uint, req *dto.CreateCategoryRequest) (*dto.CategoryInfo, error)
	Update(id uint, req *dto.UpdateCategoryRequest) error
	Delete(id uint) error
	GetByID(id uint) (*dto.CategoryInfo, error)
	GetByParentID(parentID uint, regionID uint) ([]dto.CategoryInfo, error)
	GetAll(regionID uint) ([]dto.CategoryInfo, error)
	GetTree(regionID uint) ([]dto.CategoryTree, error)
}

type categoryService struct {
	categoryRepo repository.CategoryRepository
}

// NewCategoryService 创建分类服务
func NewCategoryService(categoryRepo repository.CategoryRepository) CategoryService {
	return &categoryService{categoryRepo: categoryRepo}
}

func toCategoryInfo(c *model.Category) *dto.CategoryInfo {
	return &dto.CategoryInfo{
		ID:       c.ID,
		Name:     c.Name,
		Icon:     c.Icon,
		ParentID: c.ParentID,
		Level:    c.Level,
		Sort:     c.Sort,
		Status:   c.Status,
	}
}

// Create 创建分类
// Level 自动根据 ParentID 计算：ParentID=0 时为 1，否则为 父分类 Level+1
// 最大层级受 MaxCategoryLevel 限制，超过返回错误
func (s *categoryService) Create(regionID uint, req *dto.CreateCategoryRequest) (*dto.CategoryInfo, error) {
	status := req.Status
	if status == 0 {
		status = 1
	}

	// 根据父分类计算层级
	level := 1
	if req.ParentID > 0 {
		parent, err := s.categoryRepo.FindByID(req.ParentID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, ErrCategoryParentInvalid
			}
			return nil, err
		}
		level = parent.Level + 1
		if level > MaxCategoryLevel {
			return nil, ErrCategoryMaxLevel
		}
	}

	category := &model.Category{
		Name:     req.Name,
		Icon:     req.Icon,
		ParentID: req.ParentID,
		Level:    level,
		Sort:     req.Sort,
		Status:   status,
	}
	category.RegionID = regionID

	if err := s.categoryRepo.Create(category); err != nil {
		return nil, err
	}
	invalidateCategoryCache()
	return toCategoryInfo(category), nil
}

// Update 更新分类
func (s *categoryService) Update(id uint, req *dto.UpdateCategoryRequest) error {
	if _, err := s.categoryRepo.FindByID(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrCategoryNotFound
		}
		return err
	}

	fields := map[string]interface{}{}
	if req.Name != "" {
		fields["name"] = req.Name
	}
	if req.Icon != "" {
		fields["icon"] = req.Icon
	}
	fields["sort"] = req.Sort
	if req.Status == 0 || req.Status == 1 {
		fields["status"] = req.Status
	}
	if err := s.categoryRepo.UpdateFields(id, fields); err != nil {
		return err
	}
	invalidateCategoryCache()
	return nil
}

// Delete 删除分类
func (s *categoryService) Delete(id uint) error {
	if _, err := s.categoryRepo.FindByID(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrCategoryNotFound
		}
		return err
	}

	// 检查子分类
	category, _ := s.categoryRepo.FindByID(id)
	regionID := uint(0)
	if category != nil {
		regionID = category.RegionID
	}
	children, err := s.categoryRepo.FindByParentID(id, regionID)
	if err != nil {
		return err
	}
	if len(children) > 0 {
		return ErrCategoryHasChildren
	}
	if err := s.categoryRepo.Delete(id); err != nil {
		return err
	}
	invalidateCategoryCache()
	return nil
}

// GetByID 根据ID获取分类（cache-aside）
func (s *categoryService) GetByID(id uint) (*dto.CategoryInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	var cached dto.CategoryInfo
	if hit, _ := rediscache.GetJSON(ctx, categoryCacheKeyByID(id), &cached); hit {
		return &cached, nil
	}

	category, err := s.categoryRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCategoryNotFound
		}
		return nil, err
	}
	info := toCategoryInfo(category)
	_ = rediscache.SetJSON(ctx, categoryCacheKeyByID(id), info, categoryCacheTTL)
	return info, nil
}

// GetByParentID 根据父级ID获取子分类（cache-aside）
func (s *categoryService) GetByParentID(parentID uint, regionID uint) ([]dto.CategoryInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	var cached []dto.CategoryInfo
	if hit, _ := rediscache.GetJSON(ctx, categoryCacheKeyByParent(parentID, regionID), &cached); hit {
		return cached, nil
	}

	categories, err := s.categoryRepo.FindByParentID(parentID, regionID)
	if err != nil {
		return nil, err
	}
	result := make([]dto.CategoryInfo, 0, len(categories))
	for i := range categories {
		result = append(result, *toCategoryInfo(&categories[i]))
	}
	_ = rediscache.SetJSON(ctx, categoryCacheKeyByParent(parentID, regionID), result, categoryCacheTTL)
	return result, nil
}

// GetAll 获取全部分类平铺列表（cache-aside：供 PC/小程序门户使用）
func (s *categoryService) GetAll(regionID uint) ([]dto.CategoryInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	allCacheKey := categoryCachePrefix + "all:" + fmt.Sprintf("%d", regionID)
	var cached []dto.CategoryInfo
	if hit, _ := rediscache.GetJSON(ctx, allCacheKey, &cached); hit {
		return cached, nil
	}

	all, err := s.categoryRepo.FindByRegionID(regionID)
	if err != nil {
		return nil, err
	}
	result := make([]dto.CategoryInfo, 0, len(all))
	for i := range all {
		result = append(result, *toCategoryInfo(&all[i]))
	}
	_ = rediscache.SetJSON(ctx, allCacheKey, result, categoryCacheTTL)
	return result, nil
}

// GetTree 获取分类树形结构（cache-aside：分类树变更少，前端导航热点读）
func (s *categoryService) GetTree(regionID uint) ([]dto.CategoryTree, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	var cached []dto.CategoryTree
	if hit, _ := rediscache.GetJSON(ctx, categoryCacheKeyTree(regionID), &cached); hit {
		return cached, nil
	}

	all, err := s.categoryRepo.FindByRegionID(regionID)
	if err != nil {
		return nil, err
	}

	// 构建父ID到子节点映射
	childrenMap := make(map[uint][]model.Category)
	for _, c := range all {
		childrenMap[c.ParentID] = append(childrenMap[c.ParentID], c)
	}

	var build func(parentID uint) []dto.CategoryTree
	build = func(parentID uint) []dto.CategoryTree {
		children := childrenMap[parentID]
		trees := make([]dto.CategoryTree, 0, len(children))
		for i := range children {
			c := children[i]
			tree := dto.CategoryTree{
				CategoryInfo: *toCategoryInfo(&c),
				Children:   build(c.ID),
			}
			trees = append(trees, tree)
		}
		return trees
	}
	tree := build(0)
	_ = rediscache.SetJSON(ctx, categoryCacheKeyTree(regionID), tree, categoryCacheTTL)
	return tree, nil
}
