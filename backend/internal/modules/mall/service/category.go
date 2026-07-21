// Package service 同城商城业务逻辑层 - 商品分类
// 树形分类（parent_id 自关联），支持多级分类
package service

import (
	"errors"
	"strconv"

	"wuchang-tongcheng/internal/modules/mall/dto"
	"wuchang-tongcheng/internal/modules/mall/model"
	"wuchang-tongcheng/internal/modules/mall/repository"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

var (
	ErrCategoryNotFound      = errors.New("分类不存在")
	ErrCategoryHasProduct    = errors.New("分类下存在商品，无法删除")
	ErrCategoryStatusInvalid = errors.New("分类状态不允许此操作")
)

// CategoryService 商品分类业务接口
type CategoryService interface {
	Create(regionID uint, req *dto.CreateCategoryRequest) (*dto.CategoryInfo, error)
	Update(id uint, req *dto.UpdateCategoryRequest) error
	Delete(id uint) error
	GetByID(id uint) (*dto.CategoryInfo, error)
	List(regionID uint, req *dto.CategoryListRequest) (*utils.Pagination, []dto.CategoryInfo, error)
	ListTree(regionID uint) ([]dto.CategoryInfo, error)
	ListByParent(parentID uint) ([]dto.CategoryInfo, error)
	ListEnabled(regionID uint) ([]dto.CategoryInfo, error)
	UpdateStatus(id uint, status int) error
}

type categoryService struct {
	repo repository.CategoryRepository
}

// NewCategoryService 创建分类 service 实例
func NewCategoryService(repo repository.CategoryRepository) CategoryService {
	return &categoryService{repo: repo}
}

// categoryStatusText 状态文本
func categoryStatusText(s int) string {
	switch s {
	case model.CategoryStatusDisabled:
		return "禁用"
	case model.CategoryStatusEnabled:
		return "启用"
	}
	return ""
}

// toCategoryInfo model -> dto
func toCategoryInfo(c *model.Category) *dto.CategoryInfo {
	return &dto.CategoryInfo{
		ID:           c.ID,
		ParentID:     c.ParentID,
		Name:         c.Name,
		Icon:         c.Icon,
		Cover:        c.Cover,
		Level:        c.Level,
		Path:         c.Path,
		Sort:         c.Sort,
		Status:       c.Status,
		StatusText:   categoryStatusText(c.Status),
		IsShow:       c.IsShow,
		ProductCount: c.ProductCount,
		Keywords:     c.Keywords,
		Description:  c.Description,
		RegionID:     c.RegionID,
		CreatedAt:    c.CreatedAt,
		UpdatedAt:    c.UpdatedAt,
	}
}

// buildCategoryTree 将扁平列表构建为树形结构
func buildCategoryTree(list []model.Category, parentID uint) []dto.CategoryInfo {
	tree := make([]dto.CategoryInfo, 0)
	for i := range list {
		if list[i].ParentID == parentID {
			info := toCategoryInfo(&list[i])
			info.Children = buildCategoryTree(list, list[i].ID)
			tree = append(tree, *info)
		}
	}
	return tree
}

// Create 创建分类
func (s *categoryService) Create(regionID uint, req *dto.CreateCategoryRequest) (*dto.CategoryInfo, error) {
	level := 1
	path := ""
	if req.ParentID > 0 {
		parent, err := s.repo.FindByID(req.ParentID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, ErrCategoryNotFound
			}
			return nil, err
		}
		level = parent.Level + 1
		path = parent.Path
	}

	status := req.Status
	if status == 0 {
		status = model.CategoryStatusEnabled
	}

	c := &model.Category{
		Name:        req.Name,
		Icon:        req.Icon,
		Cover:       req.Cover,
		ParentID:    req.ParentID,
		Level:       level,
		Sort:        req.Sort,
		Status:      status,
		Keywords:    req.Keywords,
		Description: req.Description,
	}
	c.RegionID = regionID
	if req.IsShow != nil {
		c.IsShow = *req.IsShow
	} else {
		c.IsShow = true
	}

	if err := s.repo.Create(c); err != nil {
		return nil, err
	}

	// 更新 path（包含自身 ID）
	idStr := strconv.FormatUint(uint64(c.ID), 10)
	if path == "" {
		path = idStr
	} else {
		path = path + "," + idStr
	}
	_ = s.repo.UpdateFields(c.ID, map[string]interface{}{"path": path})
	c.Path = path

	return toCategoryInfo(c), nil
}

// Update 更新分类
func (s *categoryService) Update(id uint, req *dto.UpdateCategoryRequest) error {
	if _, err := s.repo.FindByID(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrCategoryNotFound
		}
		return err
	}

	fields := make(map[string]interface{})
	if req.Name != nil {
		fields["name"] = *req.Name
	}
	if req.Icon != nil {
		fields["icon"] = *req.Icon
	}
	if req.Cover != nil {
		fields["cover"] = *req.Cover
	}
	if req.Sort != nil {
		fields["sort"] = *req.Sort
	}
	if req.Status != nil {
		fields["status"] = *req.Status
	}
	if req.IsShow != nil {
		fields["is_show"] = *req.IsShow
	}
	if req.Keywords != nil {
		fields["keywords"] = *req.Keywords
	}
	if req.Description != nil {
		fields["description"] = *req.Description
	}

	if len(fields) == 0 {
		return nil
	}
	return s.repo.UpdateFields(id, fields)
}

// Delete 删除分类
func (s *categoryService) Delete(id uint) error {
	c, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrCategoryNotFound
		}
		return err
	}
	if c.ProductCount > 0 {
		return ErrCategoryHasProduct
	}
	return s.repo.Delete(id)
}

// GetByID 获取分类详情
func (s *categoryService) GetByID(id uint) (*dto.CategoryInfo, error) {
	c, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCategoryNotFound
		}
		return nil, err
	}
	return toCategoryInfo(c), nil
}

// List 分类列表
func (s *categoryService) List(regionID uint, req *dto.CategoryListRequest) (*utils.Pagination, []dto.CategoryInfo, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	opts := repository.CategoryListOptions{
		ParentID: req.ParentID,
		Level:    req.Level,
		Status:   req.Status,
		Keyword:  req.Keyword,
	}
	list, total, err := s.repo.List(opts)
	if err != nil {
		return nil, nil, err
	}

	if req.IsTree {
		tree := buildCategoryTree(list, 0)
		pagination.Total = total
		return pagination, tree, nil
	}

	infos := make([]dto.CategoryInfo, 0, len(list))
	for i := range list {
		infos = append(infos, *toCategoryInfo(&list[i]))
	}
	pagination.Total = total
	return pagination, infos, nil
}

// ListTree 分类树形列表
func (s *categoryService) ListTree(regionID uint) ([]dto.CategoryInfo, error) {
	list, err := s.repo.ListTree(regionID)
	if err != nil {
		return nil, err
	}
	return buildCategoryTree(list, 0), nil
}

// ListByParent 按父级查询子分类
func (s *categoryService) ListByParent(parentID uint) ([]dto.CategoryInfo, error) {
	list, err := s.repo.ListByParent(parentID)
	if err != nil {
		return nil, err
	}
	infos := make([]dto.CategoryInfo, 0, len(list))
	for i := range list {
		infos = append(infos, *toCategoryInfo(&list[i]))
	}
	return infos, nil
}

// ListEnabled 启用中的分类列表
func (s *categoryService) ListEnabled(regionID uint) ([]dto.CategoryInfo, error) {
	list, err := s.repo.ListEnabled()
	if err != nil {
		return nil, err
	}
	infos := make([]dto.CategoryInfo, 0, len(list))
	for i := range list {
		if regionID > 0 && list[i].RegionID != regionID {
			continue
		}
		infos = append(infos, *toCategoryInfo(&list[i]))
	}
	return infos, nil
}

// UpdateStatus 更新分类状态
func (s *categoryService) UpdateStatus(id uint, status int) error {
	if status != model.CategoryStatusDisabled && status != model.CategoryStatusEnabled {
		return ErrCategoryStatusInvalid
	}
	return s.repo.UpdateFields(id, map[string]interface{}{"status": status})
}
