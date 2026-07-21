// Package service DIY 前端页面中台业务逻辑层 - 模板（template 子域）
package service

import (
	"errors"

	"wuchang-tongcheng/internal/modules/diy/dto"
	"wuchang-tongcheng/internal/modules/diy/model"
	"wuchang-tongcheng/internal/modules/diy/repository"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

// TemplateService 模板业务接口
type TemplateService interface {
	Create(req *dto.CreateTemplateRequest) (*dto.TemplateInfo, error)
	Update(id uint, req *dto.UpdateTemplateRequest) error
	Delete(id uint) error
	GetByID(id uint) (*dto.TemplateInfo, error)
	List(req *dto.TemplateListRequest) (*utils.Pagination, []dto.TemplateInfo, error)

	// 应用模板：将模板的 pages 配置应用到新页面
	ApplyTemplate(templateID uint, regionID uint, userID uint, req *dto.ApplyTemplateRequest, pageSvc PageService) (*dto.PageInfo, error)
	// 保存为模板：将现有页面保存为模板
	SaveAsTemplate(pageID uint, name string, category string, thumbnail string, description string) (*dto.TemplateInfo, error)
}

type templateService struct {
	repo     repository.TemplateRepository
	pageRepo repository.PageRepository
}

// NewTemplateService 创建模板 service 实例
func NewTemplateService(repo repository.TemplateRepository, pageRepo repository.PageRepository) TemplateService {
	return &templateService{repo: repo, pageRepo: pageRepo}
}

// templateCategoryText 模板分类文本
func templateCategoryText(c string) string {
	switch c {
	case model.TemplateCategoryHome:
		return "首页模板"
	case model.TemplateCategoryTopic:
		return "专题页模板"
	case model.TemplateCategoryShop:
		return "店铺页模板"
	case model.TemplateCategoryActivity:
		return "活动页模板"
	}
	return ""
}

// templateStatusText 模板状态文本
func templateStatusText(s int) string {
	switch s {
	case model.TemplateStatusDisabled:
		return "禁用"
	case model.TemplateStatusEnabled:
		return "启用"
	}
	return ""
}

// toTemplateInfo model -> dto
func toTemplateInfo(t *model.Template) *dto.TemplateInfo {
	info := &dto.TemplateInfo{
		ID:           t.ID,
		Name:         t.Name,
		Thumbnail:    t.Thumbnail,
		Description:  t.Description,
		Category:     t.Category,
		CategoryText: templateCategoryText(t.Category),
		Status:       t.Status,
		StatusText:   templateStatusText(t.Status),
		CreatedAt:    t.CreatedAt,
		UpdatedAt:    t.UpdatedAt,
	}
	if t.Pages != nil {
		info.Pages = t.Pages
	}
	return info
}

// Create 创建模板
func (s *templateService) Create(req *dto.CreateTemplateRequest) (*dto.TemplateInfo, error) {
	category := req.Category
	if category == "" {
		category = model.TemplateCategoryHome
	}
	status := req.Status
	if status == 0 {
		status = model.TemplateStatusEnabled
	}
	t := &model.Template{
		Name:        req.Name,
		Thumbnail:   req.Thumbnail,
		Description: req.Description,
		Category:    category,
		Status:      status,
	}
	if req.Pages != nil {
		if b, err := model.FromJSON(req.Pages); err == nil {
			t.Pages = b
		}
	}
	if err := s.repo.Create(t); err != nil {
		return nil, err
	}
	return toTemplateInfo(t), nil
}

// Update 更新模板
func (s *templateService) Update(id uint, req *dto.UpdateTemplateRequest) error {
	if _, err := s.repo.FindByID(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrTemplateNotFound
		}
		return err
	}
	fields := make(map[string]interface{})
	if req.Name != nil {
		fields["name"] = *req.Name
	}
	if req.Thumbnail != nil {
		fields["thumbnail"] = *req.Thumbnail
	}
	if req.Description != nil {
		fields["description"] = *req.Description
	}
	if req.Category != nil {
		fields["category"] = *req.Category
	}
	if req.Pages != nil {
		if b, err := model.FromJSON(req.Pages); err == nil {
			fields["pages"] = b
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

// Delete 删除模板
func (s *templateService) Delete(id uint) error {
	if _, err := s.repo.FindByID(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrTemplateNotFound
		}
		return err
	}
	return s.repo.Delete(id)
}

// GetByID 模板详情
func (s *templateService) GetByID(id uint) (*dto.TemplateInfo, error) {
	t, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTemplateNotFound
		}
		return nil, err
	}
	return toTemplateInfo(t), nil
}

// List 模板列表
func (s *templateService) List(req *dto.TemplateListRequest) (*utils.Pagination, []dto.TemplateInfo, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	opts := repository.TemplateListOptions{
		Category: req.Category,
		Status:   req.Status,
		Keyword:  req.Keyword,
	}
	list, total, err := s.repo.List(opts, pagination)
	if err != nil {
		return nil, nil, err
	}
	infos := make([]dto.TemplateInfo, 0, len(list))
	for i := range list {
		infos = append(infos, *toTemplateInfo(&list[i]))
	}
	pagination.Total = total
	return pagination, infos, nil
}

// ApplyTemplate 应用模板创建新页面
func (s *templateService) ApplyTemplate(templateID uint, regionID uint, userID uint, req *dto.ApplyTemplateRequest, pageSvc PageService) (*dto.PageInfo, error) {
	t, err := s.repo.FindByID(templateID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTemplateNotFound
		}
		return nil, err
	}

	// 解析模板的 pages 配置，提取 components 和 settings
	var pagesConfig struct {
		Components interface{} `json:"components"`
		Settings   interface{} `json:"settings"`
	}
	if t.Pages != nil {
		_ = t.Pages.Parse(&pagesConfig)
	}

	pageType := req.Type
	if pageType == "" {
		pageType = t.Category
	}

	createReq := &dto.CreatePageRequest{
		Title:      req.Title,
		Type:       pageType,
		Slug:       req.Slug,
		BizID:      req.BizID,
		Components: pagesConfig.Components,
		Settings:   pagesConfig.Settings,
		Status:     req.Status,
	}
	return pageSvc.Create(regionID, userID, createReq)
}

// SaveAsTemplate 将现有页面保存为模板
func (s *templateService) SaveAsTemplate(pageID uint, name string, category string, thumbnail string, description string) (*dto.TemplateInfo, error) {
	p, err := s.pageRepo.FindByID(pageID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrPageNotFound
		}
		return nil, err
	}
	if category == "" {
		category = p.Type
	}
	// 构造 pages 配置
	pagesConfig := map[string]interface{}{
		"components": nil,
		"settings":   nil,
	}
	if p.Components != nil {
		var comp interface{}
		_ = p.Components.Parse(&comp)
		pagesConfig["components"] = comp
	}
	if p.Settings != nil {
		var sett interface{}
		_ = p.Settings.Parse(&sett)
		pagesConfig["settings"] = sett
	}

	t := &model.Template{
		Name:        name,
		Thumbnail:   thumbnail,
		Description: description,
		Category:    category,
		Status:      model.TemplateStatusEnabled,
	}
	if b, err := model.FromJSON(pagesConfig); err == nil {
		t.Pages = b
	}
	if err := s.repo.Create(t); err != nil {
		return nil, err
	}
	return toTemplateInfo(t), nil
}
