// Package service 同城114业务逻辑层 - 商家分类
// 依据 v3.2.1 架构方案：对标大众点评/美团/58同城
// 全行业分类（餐饮/零售/服务/娱乐/酒店/医疗/教育/生活服务）
package service

import (
	"errors"

	"wuchang-tongcheng/internal/modules/dh114/dto"
	"wuchang-tongcheng/internal/modules/dh114/model"
	"wuchang-tongcheng/internal/modules/dh114/repository"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

var (
	ErrCategoryNotFound      = errors.New("分类不存在")
	ErrCategoryCodeExists    = errors.New("分类编码已存在")
	ErrCategoryHasBusiness   = errors.New("分类下存在商户，无法删除")
	ErrCategoryStatusInvalid = errors.New("分类状态不允许此操作")
)

// CategoryService 分类业务接口
type CategoryService interface {
	Create(req *dto.CreateCategoryRequest) (*dto.CategoryInfo, error)
	Update(id uint, req *dto.UpdateCategoryRequest) error
	Delete(id uint) error
	GetByID(id uint) (*dto.CategoryInfo, error)
	List(req *dto.CategoryListRequest) (*utils.Pagination, []dto.CategoryInfo, error)
	ListByParent(parentID uint) ([]dto.CategoryInfo, error)
	ListByLevel(level int) ([]dto.CategoryInfo, error)
	ListByBusinessType(businessType string) ([]dto.CategoryInfo, error)
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
func categoryStatusText(status int) string {
	switch status {
	case model.CategoryStatusDraft:
		return "草稿"
	case model.CategoryStatusPublished:
		return "已发布"
	case model.CategoryStatusOffline:
		return "已下架"
	}
	return ""
}

// toCategoryInfo model -> dto
func toCategoryInfo(c *model.Dh114Category) *dto.CategoryInfo {
	return &dto.CategoryInfo{
		ID:            c.ID,
		Name:          c.Name,
		Code:          c.Code,
		ParentID:      c.ParentID,
		Level:         c.Level,
		BusinessType:  c.BusinessType,
		Icon:          c.Icon,
		Color:         c.Color,
		Description:   c.Description,
		Sort:          c.Sort,
		Status:        c.Status,
		StatusText:    categoryStatusText(c.Status),
		BusinessCount: c.BusinessCount,
		CreatedAt:     c.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:     c.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

// Create 创建分类
func (s *categoryService) Create(req *dto.CreateCategoryRequest) (*dto.CategoryInfo, error) {
	// 检查编码唯一性
	if existing, err := s.repo.FindByCode(req.Code); err == nil && existing != nil {
		return nil, ErrCategoryCodeExists
	}

	level := req.Level
	if level == 0 {
		level = 1
	}
	status := req.Status
	if status == 0 {
		status = model.CategoryStatusPublished
	}

	c := &model.Dh114Category{
		Name:         req.Name,
		Code:         req.Code,
		ParentID:     req.ParentID,
		Level:        level,
		BusinessType: req.BusinessType,
		Icon:         req.Icon,
		Color:        req.Color,
		Description:  req.Description,
		Sort:         req.Sort,
		Status:       status,
	}
	if c.BusinessType == "" {
		c.BusinessType = model.BusinessTypeOther
	}

	if err := s.repo.Create(c); err != nil {
		return nil, err
	}
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
	if req.Code != nil {
		// 检查编码唯一性
		if existing, err := s.repo.FindByCode(*req.Code); err == nil && existing != nil && existing.ID != id {
			return ErrCategoryCodeExists
		}
		fields["code"] = *req.Code
	}
	if req.ParentID != nil {
		fields["parent_id"] = *req.ParentID
	}
	if req.Level != nil {
		fields["level"] = *req.Level
	}
	if req.BusinessType != nil {
		fields["business_type"] = *req.BusinessType
	}
	if req.Icon != nil {
		fields["icon"] = *req.Icon
	}
	if req.Color != nil {
		fields["color"] = *req.Color
	}
	if req.Description != nil {
		fields["description"] = *req.Description
	}
	if req.Sort != nil {
		fields["sort"] = *req.Sort
	}
	if req.Status != nil {
		fields["status"] = *req.Status
	}

	if len(fields) == 0 {
		return nil
	}
	return s.repo.Update(id, fields)
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
	if c.BusinessCount > 0 {
		return ErrCategoryHasBusiness
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
func (s *categoryService) List(req *dto.CategoryListRequest) (*utils.Pagination, []dto.CategoryInfo, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	query := repository.CategoryListQuery{
		ParentID:     req.ParentID,
		Level:        req.Level,
		BusinessType: req.BusinessType,
		Status:       req.Status,
		Keyword:      req.Keyword,
	}
	list, total, err := s.repo.List(query, pagination)
	if err != nil {
		return nil, nil, err
	}
	infos := make([]dto.CategoryInfo, 0, len(list))
	for i := range list {
		infos = append(infos, *toCategoryInfo(&list[i]))
	}
	pagination.Total = total
	return pagination, infos, nil
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

// ListByLevel 按层级查询分类
func (s *categoryService) ListByLevel(level int) ([]dto.CategoryInfo, error) {
	list, err := s.repo.ListByLevel(level)
	if err != nil {
		return nil, err
	}
	infos := make([]dto.CategoryInfo, 0, len(list))
	for i := range list {
		infos = append(infos, *toCategoryInfo(&list[i]))
	}
	return infos, nil
}

// ListByBusinessType 按业务类型查询分类
func (s *categoryService) ListByBusinessType(businessType string) ([]dto.CategoryInfo, error) {
	list, err := s.repo.ListByBusinessType(businessType)
	if err != nil {
		return nil, err
	}
	infos := make([]dto.CategoryInfo, 0, len(list))
	for i := range list {
		infos = append(infos, *toCategoryInfo(&list[i]))
	}
	return infos, nil
}

// UpdateStatus 更新分类状态
func (s *categoryService) UpdateStatus(id uint, status int) error {
	if status != model.CategoryStatusDraft && status != model.CategoryStatusPublished && status != model.CategoryStatusOffline {
		return ErrCategoryStatusInvalid
	}
	return s.repo.Update(id, map[string]interface{}{"status": status})
}
