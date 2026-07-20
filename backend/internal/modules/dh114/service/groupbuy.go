// Package service 同城114业务逻辑层 - 团购
// 依据 v3.2.1 架构方案：对标大众点评/美团 限时抢购
// 限时抢购/数量限制/使用规则/有效期
package service

import (
	"errors"
	"fmt"
	"math/rand"
	"time"

	"wuchang-tongcheng/internal/modules/dh114/dto"
	"wuchang-tongcheng/internal/modules/dh114/model"
	"wuchang-tongcheng/internal/modules/dh114/repository"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

var (
	ErrGroupbuyNotFound      = errors.New("团购不存在")
	ErrGroupbuyNoPermission  = errors.New("无权操作此团购")
	ErrGroupbuyStatusInvalid = errors.New("团购状态不允许此操作")
	ErrGroupbuySoldOut       = errors.New("团购已售罄")
	ErrGroupbuyExpired       = errors.New("团购已过期")
)

// GroupbuyService 团购业务接口
type GroupbuyService interface {
	// C 端
	Create(regionID uint, userID uint, req *dto.CreateGroupbuyRequest) (*dto.GroupbuyInfo, error)
	Update(id uint, operatorID uint, req *dto.UpdateGroupbuyRequest) error
	Delete(id uint, operatorID uint) error
	GetByID(id uint) (*dto.GroupbuyInfo, error)
	List(regionID uint, req *dto.GroupbuyListRequest) (*utils.Pagination, []dto.GroupbuyInfo, error)
	ListByDh114(regionID uint, dh114ID uint, page, pageSize int) (*utils.Pagination, []dto.GroupbuyInfo, error)
	ListHot(regionID uint, page, pageSize int) (*utils.Pagination, []dto.GroupbuyInfo, error)

	// 互动
	IncrViewCount(id uint) error
	IncrFavCount(id uint) error
	IncrSoldCount(id uint, count int) error

	// M 端管理
	AdminList(req *dto.GroupbuyAdminListRequest) (*utils.Pagination, []dto.GroupbuyInfo, error)
	Audit(id uint, auditStatus int, auditReason string) error
	BatchAudit(req *dto.BatchAuditRequest) (*dto.BatchResultResponse, error)
	AdminUpdateStatus(id uint, status int) error
}

type groupbuyService struct {
	repo repository.GroupbuyRepository
}

// NewGroupbuyService 创建团购 service 实例
func NewGroupbuyService(repo repository.GroupbuyRepository) GroupbuyService {
	return &groupbuyService{repo: repo}
}

// groupbuyStatusText 团购状态文本
func groupbuyStatusText(s int) string {
	switch s {
	case model.GroupbuyStatusDraft:
		return "草稿"
	case model.GroupbuyStatusPublished:
		return "已发布"
	case model.GroupbuyStatusSoldOut:
		return "已售罄"
	case model.GroupbuyStatusOffline:
		return "已下架"
	case model.GroupbuyStatusExpired:
		return "已过期"
	}
	return ""
}

// toGroupbuyInfo model -> dto
func toGroupbuyInfo(g *model.Dh114Groupbuy) *dto.GroupbuyInfo {
	info := &dto.GroupbuyInfo{
		ID:              g.ID,
		GroupbuyNo:      g.GroupbuyNo,
		Dh114ID:         g.Dh114ID,
		BusinessID:      g.BusinessID,
		Title:           g.Title,
		Description:     g.Description,
		CoverImage:      g.CoverImage,
		OriginalPrice:   g.OriginalPrice,
		GroupbuyPrice:   g.GroupbuyPrice,
		Discount:        g.Discount,
		TotalCount:      g.TotalCount,
		SoldCount:       g.SoldCount,
		PerUserLimit:    g.PerUserLimit,
		StartTime:       g.StartTime,
		EndTime:         g.EndTime,
		ValidStart:      g.ValidStart,
		ValidEnd:        g.ValidEnd,
		NeedReservation: g.NeedReservation,
		ViewCount:       g.ViewCount,
		FavCount:        g.FavCount,
		Status:          g.Status,
		StatusText:      groupbuyStatusText(g.Status),
		AuditStatus:     g.AuditStatus,
		AuditReason:     g.AuditReason,
		PublishedAt:     g.PublishedAt,
		Featured:        g.Featured,
		RegionID:        g.RegionID,
		CreatedAt:       g.CreatedAt,
		UpdatedAt:       g.UpdatedAt,
	}
	if g.Images != nil {
		info.Images = g.Images
	}
	if g.ValidWeekdays != nil {
		info.ValidWeekdays = g.ValidWeekdays
	}
	if g.UseInstructions != nil {
		info.UseInstructions = g.UseInstructions
	}
	if g.UseTimeRanges != nil {
		info.UseTimeRanges = g.UseTimeRanges
	}
	return info
}

// generateGroupbuyNo 生成团购单号
func generateGroupbuyNo() string {
	return fmt.Sprintf("DH114GB%s%06d", time.Now().Format("20060102150405"), rand.Intn(1000000))
}

// Create 创建团购
func (s *groupbuyService) Create(regionID uint, userID uint, req *dto.CreateGroupbuyRequest) (*dto.GroupbuyInfo, error) {
	status := req.Status
	if status == 0 {
		status = model.GroupbuyStatusDraft
	}

	g := &model.Dh114Groupbuy{
		GroupbuyNo:      generateGroupbuyNo(),
		Dh114ID:         req.Dh114ID,
		Title:           req.Title,
		Description:     req.Description,
		CoverImage:      req.CoverImage,
		OriginalPrice:   req.OriginalPrice,
		GroupbuyPrice:   req.GroupbuyPrice,
		TotalCount:      req.TotalCount,
		PerUserLimit:    req.PerUserLimit,
		StartTime:       req.StartTime,
		EndTime:         req.EndTime,
		ValidStart:      req.ValidStart,
		ValidEnd:        req.ValidEnd,
		NeedReservation: req.NeedReservation,
		Status:          status,
		AuditStatus:     model.AuditApproved, // MVP：发布即通过
	}
	g.RegionID = regionID

	// 计算折扣
	if g.OriginalPrice > 0 && g.GroupbuyPrice > 0 {
		g.Discount = g.GroupbuyPrice / g.OriginalPrice
	}

	// JSONB 字段处理
	if req.Images != nil {
		if b, err := model.FromJSON(req.Images); err == nil {
			g.Images = b
		}
	}
	if req.ValidWeekdays != nil {
		if b, err := model.FromJSON(req.ValidWeekdays); err == nil {
			g.ValidWeekdays = b
		}
	}
	if req.UseInstructions != nil {
		if b, err := model.FromJSON(req.UseInstructions); err == nil {
			g.UseInstructions = b
		}
	}
	if req.UseTimeRanges != nil {
		if b, err := model.FromJSON(req.UseTimeRanges); err == nil {
			g.UseTimeRanges = b
		}
	}

	// 状态为已发布时记录发布时间
	if g.Status == model.GroupbuyStatusPublished {
		now := time.Now()
		g.PublishedAt = &now
	}

	if err := s.repo.Create(g); err != nil {
		return nil, err
	}
	return toGroupbuyInfo(g), nil
}

// Update 更新团购
func (s *groupbuyService) Update(id uint, operatorID uint, req *dto.UpdateGroupbuyRequest) error {
	g, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrGroupbuyNotFound
		}
		return err
	}
	// MVP 简化：不做归属校验，由 handler 层注入的 userID 控制权限
	_ = operatorID
	_ = g

	fields := make(map[string]interface{})
	if req.Title != nil {
		fields["title"] = *req.Title
	}
	if req.Description != nil {
		fields["description"] = *req.Description
	}
	if req.CoverImage != nil {
		fields["cover_image"] = *req.CoverImage
	}
	if req.OriginalPrice != nil {
		fields["original_price"] = *req.OriginalPrice
	}
	if req.GroupbuyPrice != nil {
		fields["groupbuy_price"] = *req.GroupbuyPrice
	}
	if req.TotalCount != nil {
		fields["total_count"] = *req.TotalCount
	}
	if req.PerUserLimit != nil {
		fields["per_user_limit"] = *req.PerUserLimit
	}
	if req.StartTime != nil {
		fields["start_time"] = req.StartTime
	}
	if req.EndTime != nil {
		fields["end_time"] = req.EndTime
	}
	if req.ValidStart != nil {
		fields["valid_start"] = req.ValidStart
	}
	if req.ValidEnd != nil {
		fields["valid_end"] = req.ValidEnd
	}
	if req.NeedReservation != nil {
		fields["need_reservation"] = *req.NeedReservation
	}
	if req.Status != nil {
		fields["status"] = *req.Status
		if *req.Status == model.GroupbuyStatusPublished && g.PublishedAt == nil {
			now := time.Now()
			fields["published_at"] = &now
		}
	}

	// JSONB 字段处理
	if req.Images != nil {
		if b, err := model.FromJSON(req.Images); err == nil {
			fields["images"] = b
		}
	}
	if req.ValidWeekdays != nil {
		if b, err := model.FromJSON(req.ValidWeekdays); err == nil {
			fields["valid_weekdays"] = b
		}
	}
	if req.UseInstructions != nil {
		if b, err := model.FromJSON(req.UseInstructions); err == nil {
			fields["use_instructions"] = b
		}
	}
	if req.UseTimeRanges != nil {
		if b, err := model.FromJSON(req.UseTimeRanges); err == nil {
			fields["use_time_ranges"] = b
		}
	}

	// 重算折扣
	if req.OriginalPrice != nil || req.GroupbuyPrice != nil {
		originalPrice := g.OriginalPrice
		if req.OriginalPrice != nil {
			originalPrice = *req.OriginalPrice
		}
		groupbuyPrice := g.GroupbuyPrice
		if req.GroupbuyPrice != nil {
			groupbuyPrice = *req.GroupbuyPrice
		}
		if originalPrice > 0 && groupbuyPrice > 0 {
			fields["discount"] = groupbuyPrice / originalPrice
		}
	}

	if len(fields) == 0 {
		return nil
	}
	return s.repo.Update(id, fields)
}

// Delete 删除团购
func (s *groupbuyService) Delete(id uint, operatorID uint) error {
	if _, err := s.repo.FindByID(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrGroupbuyNotFound
		}
		return err
	}
	_ = operatorID
	return s.repo.Delete(id)
}

// GetByID 获取团购详情
func (s *groupbuyService) GetByID(id uint) (*dto.GroupbuyInfo, error) {
	g, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrGroupbuyNotFound
		}
		return nil, err
	}
	// 增加浏览量
	_ = s.repo.IncrViewCount(id)
	return toGroupbuyInfo(g), nil
}

// List 团购列表
func (s *groupbuyService) List(regionID uint, req *dto.GroupbuyListRequest) (*utils.Pagination, []dto.GroupbuyInfo, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	query := repository.GroupbuyListQuery{
		Dh114ID:  req.Dh114ID,
		Featured: req.Featured,
		MinPrice: req.MinPrice,
		MaxPrice: req.MaxPrice,
		Keyword:  req.Keyword,
		Sort:     req.Sort,
	}
	if req.Status != nil {
		query.Status = req.Status
	}
	list, total, err := s.repo.List(regionID, query, pagination)
	if err != nil {
		return nil, nil, err
	}
	infos := make([]dto.GroupbuyInfo, 0, len(list))
	for i := range list {
		infos = append(infos, *toGroupbuyInfo(&list[i]))
	}
	pagination.Total = total
	return pagination, infos, nil
}

// ListByDh114 按商户列出团购
func (s *groupbuyService) ListByDh114(regionID uint, dh114ID uint, page, pageSize int) (*utils.Pagination, []dto.GroupbuyInfo, error) {
	pagination := utils.NewPagination(page, pageSize)
	list, total, err := s.repo.ListByDh114(regionID, dh114ID, pagination)
	if err != nil {
		return nil, nil, err
	}
	infos := make([]dto.GroupbuyInfo, 0, len(list))
	for i := range list {
		infos = append(infos, *toGroupbuyInfo(&list[i]))
	}
	pagination.Total = total
	return pagination, infos, nil
}

// ListHot 热门团购
func (s *groupbuyService) ListHot(regionID uint, page, pageSize int) (*utils.Pagination, []dto.GroupbuyInfo, error) {
	pagination := utils.NewPagination(page, pageSize)
	list, total, err := s.repo.ListHot(regionID, pagination)
	if err != nil {
		return nil, nil, err
	}
	infos := make([]dto.GroupbuyInfo, 0, len(list))
	for i := range list {
		infos = append(infos, *toGroupbuyInfo(&list[i]))
	}
	pagination.Total = total
	return pagination, infos, nil
}

// IncrViewCount 增加浏览数
func (s *groupbuyService) IncrViewCount(id uint) error {
	return s.repo.IncrViewCount(id)
}

// IncrFavCount 增加收藏数
func (s *groupbuyService) IncrFavCount(id uint) error {
	return s.repo.IncrFavCount(id)
}

// IncrSoldCount 增加已售数（下单时调用）
func (s *groupbuyService) IncrSoldCount(id uint, count int) error {
	if count <= 0 {
		count = 1
	}
	return s.repo.IncrSoldCount(id, count)
}

// AdminList 管理后台列表
func (s *groupbuyService) AdminList(req *dto.GroupbuyAdminListRequest) (*utils.Pagination, []dto.GroupbuyInfo, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	query := repository.GroupbuyAdminListQuery{
		Dh114ID:     req.Dh114ID,
		Status:      req.Status,
		AuditStatus: req.AuditStatus,
		Keyword:     req.Keyword,
	}
	list, total, err := s.repo.AdminList(query, pagination)
	if err != nil {
		return nil, nil, err
	}
	infos := make([]dto.GroupbuyInfo, 0, len(list))
	for i := range list {
		infos = append(infos, *toGroupbuyInfo(&list[i]))
	}
	pagination.Total = total
	return pagination, infos, nil
}

// Audit 审核团购
func (s *groupbuyService) Audit(id uint, auditStatus int, auditReason string) error {
	if _, err := s.repo.FindByID(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrGroupbuyNotFound
		}
		return err
	}
	fields := map[string]interface{}{
		"audit_status": auditStatus,
		"audit_reason": auditReason,
	}
	// 审核通过且未发布过，自动发布
	if auditStatus == model.AuditApproved {
		g, err := s.repo.FindByID(id)
		if err == nil && g.PublishedAt == nil {
			now := time.Now()
			fields["published_at"] = &now
			if g.Status == model.GroupbuyStatusDraft {
				fields["status"] = model.GroupbuyStatusPublished
			}
		}
	}
	return s.repo.Update(id, fields)
}

// BatchAudit 批量审核
func (s *groupbuyService) BatchAudit(req *dto.BatchAuditRequest) (*dto.BatchResultResponse, error) {
	result := &dto.BatchResultResponse{Total: len(req.IDs)}
	failedIDs := make([]uint, 0)
	for _, id := range req.IDs {
		if err := s.Audit(id, req.AuditStatus, req.AuditReason); err != nil {
			failedIDs = append(failedIDs, id)
		} else {
			result.Success++
		}
	}
	result.Failed = len(failedIDs)
	result.FailedIDs = failedIDs
	return result, nil
}

// AdminUpdateStatus 管理后台更新状态
func (s *groupbuyService) AdminUpdateStatus(id uint, status int) error {
	if status < model.GroupbuyStatusDraft || status > model.GroupbuyStatusExpired {
		return ErrGroupbuyStatusInvalid
	}
	fields := map[string]interface{}{"status": status}
	if status == model.GroupbuyStatusPublished {
		g, err := s.repo.FindByID(id)
		if err == nil && g.PublishedAt == nil {
			now := time.Now()
			fields["published_at"] = &now
		}
	}
	return s.repo.Update(id, fields)
}
