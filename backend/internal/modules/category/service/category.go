// Package service 分类信息业务逻辑层
package service

import (
	"errors"

	"wuchang-tongcheng/internal/modules/category/dto"
	"wuchang-tongcheng/internal/modules/category/model"
	"wuchang-tongcheng/internal/modules/category/repository"

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

// CategoryService 分类业务逻辑接口
type CategoryService interface {
	Create(regionID uint, req *dto.CreateCategoryRequest) (*dto.CategoryInfo, error)
	Update(id uint, req *dto.UpdateCategoryRequest) error
	Delete(id uint) error
	GetByID(id uint) (*dto.CategoryInfo, error)
	GetByParentID(parentID uint, regionID uint) ([]dto.CategoryInfo, error)
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
	return s.categoryRepo.UpdateFields(id, fields)
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
	return s.categoryRepo.Delete(id)
}

// GetByID 根据ID获取分类
func (s *categoryService) GetByID(id uint) (*dto.CategoryInfo, error) {
	category, err := s.categoryRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCategoryNotFound
		}
		return nil, err
	}
	return toCategoryInfo(category), nil
}

// GetByParentID 根据父级ID获取子分类
func (s *categoryService) GetByParentID(parentID uint, regionID uint) ([]dto.CategoryInfo, error) {
	categories, err := s.categoryRepo.FindByParentID(parentID, regionID)
	if err != nil {
		return nil, err
	}
	result := make([]dto.CategoryInfo, 0, len(categories))
	for i := range categories {
		result = append(result, *toCategoryInfo(&categories[i]))
	}
	return result, nil
}

// GetTree 获取分类树形结构
func (s *categoryService) GetTree(regionID uint) ([]dto.CategoryTree, error) {
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
				Children:     build(c.ID),
			}
			trees = append(trees, tree)
		}
		return trees
	}
	return build(0), nil
}
