// Package service DIY 前端页面中台业务逻辑层 - 页面（page 子域）
// 依据架构设计 4.12：拖拽生成首页/专题页/店铺页/活动页
// 依据需求文档 1.10：4 维数据隔离（region_id + user_id）
package service

import (
	"errors"
	"time"

	"wuchang-tongcheng/internal/modules/diy/dto"
	"wuchang-tongcheng/internal/modules/diy/model"
	"wuchang-tongcheng/internal/modules/diy/repository"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

// PageService 页面业务接口
type PageService interface {
	// C 端
	Create(regionID uint, userID uint, req *dto.CreatePageRequest) (*dto.PageInfo, error)
	Update(id uint, operatorID uint, req *dto.UpdatePageRequest) error
	Delete(id uint, operatorID uint) error
	GetByID(id uint) (*dto.PageInfo, error)
	GetBySlug(regionID uint, slug string) (*dto.PageInfo, error)
	List(regionID uint, req *dto.PageListRequest) (*utils.Pagination, []dto.PageInfo, error)
	ListMine(userID uint, req *dto.PageListRequest) (*utils.Pagination, []dto.PageInfo, error)

	// 状态管理
	Publish(id uint, operatorID uint) error
	Offline(id uint, operatorID uint) error

	// 复制
	Copy(id uint, operatorID uint, newTitle string) (*dto.PageInfo, error)

	// M 端
	AdminList(req *dto.PageListAdminRequest) (*utils.Pagination, []dto.PageInfo, error)
	AdminGetByID(id uint) (*dto.PageInfo, error)
	AdminUpdateStatus(id uint, status int) error
}

type pageService struct {
	repo repository.PageRepository
}

// NewPageService 创建页面 service 实例
func NewPageService(repo repository.PageRepository) PageService {
	return &pageService{repo: repo}
}

// pageTypeText 页面类型文本
func pageTypeText(t string) string {
	switch t {
	case model.PageTypeHome:
		return "首页"
	case model.PageTypeTopic:
		return "专题页"
	case model.PageTypeShop:
		return "店铺页"
	case model.PageTypeActivity:
		return "活动页"
	}
	return ""
}

// pageStatusText 页面状态文本
func pageStatusText(s int) string {
	switch s {
	case model.PageStatusDraft:
		return "草稿"
	case model.PageStatusPublish:
		return "已发布"
	case model.PageStatusOffline:
		return "已下线"
	}
	return ""
}

// toPageInfo model -> dto
func toPageInfo(p *model.Page) *dto.PageInfo {
	info := &dto.PageInfo{
		ID:          p.ID,
		RegionID:    p.RegionID,
		Title:       p.Title,
		Type:        p.Type,
		TypeText:    pageTypeText(p.Type),
		Slug:        p.Slug,
		Status:      p.Status,
		StatusText:  pageStatusText(p.Status),
		UserID:      p.UserID,
		BizID:       p.BizID,
		PublishedAt: p.PublishedAt,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
	}
	if p.Components != nil {
		info.Components = p.Components
	}
	if p.Settings != nil {
		info.Settings = p.Settings
	}
	return info
}

// Create 创建页面
func (s *pageService) Create(regionID uint, userID uint, req *dto.CreatePageRequest) (*dto.PageInfo, error) {
	pageType := req.Type
	if pageType == "" {
		pageType = model.PageTypeHome
	}
	status := req.Status
	if status == 0 {
		status = model.PageStatusDraft
	}
	// 发布状态必须设置 slug
	if status == model.PageStatusPublish && req.Slug == "" {
		return nil, ErrPageSlugEmpty
	}

	p := &model.Page{
		Title:  req.Title,
		Type:   pageType,
		Slug:   req.Slug,
		Status: status,
		UserID: userID,
		BizID:  req.BizID,
	}
	p.RegionID = regionID
	if status == model.PageStatusPublish {
		now := time.Now()
		p.PublishedAt = &now
	}

	// JSONB 字段处理
	if req.Components != nil {
		if b, err := model.FromJSON(req.Components); err == nil {
			p.Components = b
		}
	}
	if req.Settings != nil {
		if b, err := model.FromJSON(req.Settings); err == nil {
			p.Settings = b
		}
	}

	if err := s.repo.Create(p); err != nil {
		return nil, err
	}
	return toPageInfo(p), nil
}

// Update 更新页面（仅创建者本人）
func (s *pageService) Update(id uint, operatorID uint, req *dto.UpdatePageRequest) error {
	p, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrPageNotFound
		}
		return err
	}
	if p.UserID != operatorID {
		return ErrPageNoPermission
	}

	fields := make(map[string]interface{})
	if req.Title != nil {
		fields["title"] = *req.Title
	}
	if req.Type != nil {
		fields["type"] = *req.Type
	}
	if req.Slug != nil {
		fields["slug"] = *req.Slug
	}
	if req.BizID != nil {
		fields["biz_id"] = *req.BizID
	}
	if req.Components != nil {
		if b, err := model.FromJSON(req.Components); err == nil {
			fields["components"] = b
		}
	}
	if req.Settings != nil {
		if b, err := model.FromJSON(req.Settings); err == nil {
			fields["settings"] = b
		}
	}
	if req.Status != nil {
		fields["status"] = *req.Status
		// 状态变更为已发布时记录发布时间
		if *req.Status == model.PageStatusPublish && p.PublishedAt == nil {
			now := time.Now()
			fields["published_at"] = &now
		}
		// 发布状态必须设置 slug
		if *req.Status == model.PageStatusPublish && p.Slug == "" {
			if req.Slug == nil || *req.Slug == "" {
				return ErrPageSlugEmpty
			}
		}
	}

	if len(fields) == 0 {
		return nil
	}
	return s.repo.UpdateFields(id, fields)
}

// Delete 删除页面（仅创建者本人）
func (s *pageService) Delete(id uint, operatorID uint) error {
	p, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrPageNotFound
		}
		return err
	}
	if p.UserID != operatorID {
		return ErrPageNoPermission
	}
	return s.repo.Delete(id)
}

// GetByID 获取页面详情
func (s *pageService) GetByID(id uint) (*dto.PageInfo, error) {
	p, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrPageNotFound
		}
		return nil, err
	}
	return toPageInfo(p), nil
}

// GetBySlug 按 slug 获取已发布页面
func (s *pageService) GetBySlug(regionID uint, slug string) (*dto.PageInfo, error) {
	p, err := s.repo.FindBySlug(regionID, slug)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrPageNotFound
		}
		return nil, err
	}
	return toPageInfo(p), nil
}

// List 页面列表（C 端仅返回已发布）
func (s *pageService) List(regionID uint, req *dto.PageListRequest) (*utils.Pagination, []dto.PageInfo, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	status := model.PageStatusPublish
	opts := repository.PageListOptions{
		UserID:  req.UserID,
		Type:    req.Type,
		Status:  &status,
		Keyword: req.Keyword,
	}
	// 如果用户筛选自己的页面（mine 调用），不限制状态
	if req.UserID > 0 && req.Status == nil {
		opts.Status = nil
	} else if req.Status != nil {
		opts.Status = req.Status
	}
	list, total, err := s.repo.List(regionID, opts, pagination)
	if err != nil {
		return nil, nil, err
	}
	infos := make([]dto.PageInfo, 0, len(list))
	for i := range list {
		infos = append(infos, *toPageInfo(&list[i]))
	}
	pagination.Total = total
	return pagination, infos, nil
}

// ListMine 我的页面列表
func (s *pageService) ListMine(userID uint, req *dto.PageListRequest) (*utils.Pagination, []dto.PageInfo, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	opts := repository.PageListOptions{
		Type:    req.Type,
		Status:  req.Status,
		Keyword: req.Keyword,
	}
	list, total, err := s.repo.ListByUser(userID, opts, pagination)
	if err != nil {
		return nil, nil, err
	}
	infos := make([]dto.PageInfo, 0, len(list))
	for i := range list {
		infos = append(infos, *toPageInfo(&list[i]))
	}
	pagination.Total = total
	return pagination, infos, nil
}

// Publish 发布页面（仅创建者本人）
func (s *pageService) Publish(id uint, operatorID uint) error {
	p, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrPageNotFound
		}
		return err
	}
	if p.UserID != operatorID {
		return ErrPageNoPermission
	}
	if p.Slug == "" {
		return ErrPageSlugEmpty
	}
	fields := map[string]interface{}{
		"status": model.PageStatusPublish,
	}
	if p.PublishedAt == nil {
		now := time.Now()
		fields["published_at"] = &now
	}
	return s.repo.UpdateFields(id, fields)
}

// Offline 下线页面（仅创建者本人）
func (s *pageService) Offline(id uint, operatorID uint) error {
	p, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrPageNotFound
		}
		return err
	}
	if p.UserID != operatorID {
		return ErrPageNoPermission
	}
	return s.repo.UpdateFields(id, map[string]interface{}{
		"status": model.PageStatusOffline,
	})
}

// Copy 复制页面（仅创建者本人可复制自己的页面）
func (s *pageService) Copy(id uint, operatorID uint, newTitle string) (*dto.PageInfo, error) {
	p, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrPageNotFound
		}
		return nil, err
	}
	if p.UserID != operatorID {
		return nil, ErrPageNoPermission
	}
	title := newTitle
	if title == "" {
		title = p.Title + " - 副本"
	}
	newPage := &model.Page{
		Title:  title,
		Type:   p.Type,
		Slug:   "", // 副本不复制 slug，避免冲突
		Status: model.PageStatusDraft,
		UserID: operatorID,
		BizID:  p.BizID,
	}
	newPage.RegionID = p.RegionID
	if p.Components != nil {
		newPage.Components = make(model.JSONB, len(p.Components))
		copy(newPage.Components, p.Components)
	}
	if p.Settings != nil {
		newPage.Settings = make(model.JSONB, len(p.Settings))
		copy(newPage.Settings, p.Settings)
	}
	if err := s.repo.Create(newPage); err != nil {
		return nil, err
	}
	return toPageInfo(newPage), nil
}

// AdminList 管理后台列表
func (s *pageService) AdminList(req *dto.PageListAdminRequest) (*utils.Pagination, []dto.PageInfo, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	opts := repository.PageListOptions{
		UserID:  req.UserID,
		Type:    req.Type,
		Status:  req.Status,
		Keyword: req.Keyword,
	}
	list, total, err := s.repo.List(req.RegionID, opts, pagination)
	if err != nil {
		return nil, nil, err
	}
	infos := make([]dto.PageInfo, 0, len(list))
	for i := range list {
		infos = append(infos, *toPageInfo(&list[i]))
	}
	pagination.Total = total
	return pagination, infos, nil
}

// AdminGetByID 管理后台获取详情
func (s *pageService) AdminGetByID(id uint) (*dto.PageInfo, error) {
	p, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrPageNotFound
		}
		return nil, err
	}
	return toPageInfo(p), nil
}

// AdminUpdateStatus 管理后台更新状态
func (s *pageService) AdminUpdateStatus(id uint, status int) error {
	fields := map[string]interface{}{"status": status}
	if status == model.PageStatusPublish {
		p, err := s.repo.FindByID(id)
		if err == nil && p.PublishedAt == nil {
			now := time.Now()
			fields["published_at"] = &now
		}
	}
	return s.repo.UpdateFields(id, fields)
}
