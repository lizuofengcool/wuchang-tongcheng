// Package service DIY 前端页面中台业务逻辑层 - 组件（component 子域）
package service

import (
	"errors"

	"wuchang-tongcheng/internal/modules/diy/dto"
	"wuchang-tongcheng/internal/modules/diy/model"
	"wuchang-tongcheng/internal/modules/diy/repository"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

// ComponentService 组件业务接口
type ComponentService interface {
	Create(req *dto.CreateComponentRequest) (*dto.ComponentInfo, error)
	Update(id uint, req *dto.UpdateComponentRequest) error
	Delete(id uint) error
	GetByID(id uint) (*dto.ComponentInfo, error)
	GetByCode(code string) (*dto.ComponentInfo, error)
	List(req *dto.ComponentListRequest) (*utils.Pagination, []dto.ComponentInfo, error)
	ListByCategory(category string, page, pageSize int) (*utils.Pagination, []dto.ComponentInfo, error)
}

type componentService struct {
	repo repository.ComponentRepository
}

// NewComponentService 创建组件 service 实例
func NewComponentService(repo repository.ComponentRepository) ComponentService {
	return &componentService{repo: repo}
}

// componentCategoryText 组件分类文本
func componentCategoryText(c string) string {
	switch c {
	case model.ComponentCategoryBasic:
		return "基础组件"
	case model.ComponentCategoryLayout:
		return "布局组件"
	case model.ComponentCategoryBusiness:
		return "业务组件"
	}
	return ""
}

// componentStatusText 组件状态文本
func componentStatusText(s int) string {
	switch s {
	case model.ComponentStatusDisabled:
		return "禁用"
	case model.ComponentStatusEnabled:
		return "启用"
	}
	return ""
}

// toComponentInfo model -> dto
func toComponentInfo(c *model.Component) *dto.ComponentInfo {
	info := &dto.ComponentInfo{
		ID:           c.ID,
		Name:         c.Name,
		Code:         c.Code,
		Category:     c.Category,
		CategoryText: componentCategoryText(c.Category),
		Description:  c.Description,
		Thumbnail:    c.Thumbnail,
		Status:       c.Status,
		StatusText:   componentStatusText(c.Status),
		CreatedAt:    c.CreatedAt,
		UpdatedAt:    c.UpdatedAt,
	}
	if c.Config != nil {
		info.Config = c.Config
	}
	return info
}

// Create 创建组件
func (s *componentService) Create(req *dto.CreateComponentRequest) (*dto.ComponentInfo, error) {
	category := req.Category
	if category == "" {
		category = model.ComponentCategoryBasic
	}
	status := req.Status
	if status == 0 {
		status = model.ComponentStatusEnabled
	}

	// 校验 code 唯一
	if existing, err := s.repo.FindByCode(req.Code); err == nil && existing != nil {
		return nil, ErrComponentCodeConflict
	} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	c := &model.Component{
		Name:        req.Name,
		Code:        req.Code,
		Category:    category,
		Description: req.Description,
		Thumbnail:   req.Thumbnail,
		Status:      status,
	}
	if req.Config != nil {
		if b, err := model.FromJSON(req.Config); err == nil {
			c.Config = b
		}
	}
	if err := s.repo.Create(c); err != nil {
		return nil, err
	}
	return toComponentInfo(c), nil
}

// Update 更新组件
func (s *componentService) Update(id uint, req *dto.UpdateComponentRequest) error {
	if _, err := s.repo.FindByID(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrComponentNotFound
		}
		return err
	}
	fields := make(map[string]interface{})
	if req.Name != nil {
		fields["name"] = *req.Name
	}
	if req.Category != nil {
		fields["category"] = *req.Category
	}
	if req.Description != nil {
		fields["description"] = *req.Description
	}
	if req.Thumbnail != nil {
		fields["thumbnail"] = *req.Thumbnail
	}
	if req.Config != nil {
		if b, err := model.FromJSON(req.Config); err == nil {
			fields["config"] = b
		}
	}
	if req.Status != nil {
		fields["status"] = *req.Status
	}
	if len(fields) == 0 {
		return nil
	}
	return s.repo.UpdateFields(id, fields)
}

// Delete 删除组件
func (s *componentService) Delete(id uint) error {
	if _, err := s.repo.FindByID(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrComponentNotFound
		}
		return err
	}
	return s.repo.Delete(id)
}

// GetByID 组件详情
func (s *componentService) GetByID(id uint) (*dto.ComponentInfo, error) {
	c, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrComponentNotFound
		}
		return nil, err
	}
	return toComponentInfo(c), nil
}

// GetByCode 按 code 获取组件
func (s *componentService) GetByCode(code string) (*dto.ComponentInfo, error) {
	c, err := s.repo.FindByCode(code)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrComponentNotFound
		}
		return nil, err
	}
	return toComponentInfo(c), nil
}

// List 组件列表
func (s *componentService) List(req *dto.ComponentListRequest) (*utils.Pagination, []dto.ComponentInfo, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	opts := repository.ComponentListOptions{
		Category: req.Category,
		Status:   req.Status,
		Keyword:  req.Keyword,
	}
	list, total, err := s.repo.List(opts, pagination)
	if err != nil {
		return nil, nil, err
	}
	infos := make([]dto.ComponentInfo, 0, len(list))
	for i := range list {
		infos = append(infos, *toComponentInfo(&list[i]))
	}
	pagination.Total = total
	return pagination, infos, nil
}

// ListByCategory 按分类获取组件
func (s *componentService) ListByCategory(category string, page, pageSize int) (*utils.Pagination, []dto.ComponentInfo, error) {
	pagination := utils.NewPagination(page, pageSize)
	list, total, err := s.repo.ListByCategory(category, pagination)
	if err != nil {
		return nil, nil, err
	}
	infos := make([]dto.ComponentInfo, 0, len(list))
	for i := range list {
		infos = append(infos, *toComponentInfo(&list[i]))
	}
	pagination.Total = total
	return pagination, infos, nil
}
