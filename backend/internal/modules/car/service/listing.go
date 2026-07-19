// Package service 同城车辆买卖业务逻辑层 - 车源发布单
// 依据需求文档 1.10：4 维数据隔离（region_id + user_id）
// 依据需求文档 1.5：内容审核必须做（MVP 简化为发布即通过，M 端可手动审核/下架）
package service

import (
	"errors"
	"fmt"
	"math/rand"
	"time"

	"wuchang-tongcheng/internal/modules/car/dto"
	"wuchang-tongcheng/internal/modules/car/model"
	"wuchang-tongcheng/internal/modules/car/repository"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

var (
	ErrListingNotFound      = errors.New("发布单不存在")
	ErrListingNoPermission  = errors.New("无权操作此发布单")
	ErrListingAudited       = errors.New("已审核的发布单不能重复审核")
	ErrListingStatusInvalid = errors.New("发布单状态不允许此操作")
)

// ListingService 发布单业务接口
type ListingService interface {
	// C 端
	Create(regionID uint, userID uint, userName string, userAvatar string, req *dto.CreateListingRequest) (*dto.ListingInfo, error)
	Update(id uint, operatorID uint, req *dto.UpdateListingRequest) error
	Delete(id uint, operatorID uint) error
	GetByID(id uint) (*dto.ListingInfo, error)
	List(regionID uint, req *dto.ListingListRequest) (*utils.Pagination, []dto.ListingInfo, error)
	ListMine(userID uint, page, pageSize int) (*utils.Pagination, []dto.ListingInfo, error)

	// M 端管理
	AdminList(req *dto.ListingAdminListRequest) (*utils.Pagination, []dto.ListingInfo, error)
	AdminGetByID(id uint) (*dto.ListingInfo, error)
	Audit(id uint, req *dto.ListingAuditRequest) error
	UpdateStatus(id uint, status int) error
	UpdateInspectionStatus(id uint, req *dto.InspectionStatusUpdateRequest) error
	UpdateRealCarVerified(id uint, verified bool) error
}

type listingService struct {
	repo repository.ListingRepository
}

// NewListingService 创建发布单 service 实例
func NewListingService(repo repository.ListingRepository) ListingService {
	return &listingService{repo: repo}
}

// listingStatusText 状态文本
func listingStatusText(status int) string {
	switch status {
	case model.ListingStatusDraft:
		return "草稿"
	case model.ListingStatusPublished:
		return "已发布"
	case model.ListingStatusOffline:
		return "已下架"
	case model.ListingStatusExpired:
		return "已过期"
	case model.ListingStatusSold:
		return "已售出"
	}
	return ""
}

// listingAuditStatusText 审核状态文本
func listingAuditStatusText(s int) string {
	switch s {
	case model.ListingAuditPending:
		return "待审"
	case model.ListingAuditApproved:
		return "通过"
	case model.ListingAuditRejected:
		return "拒绝"
	}
	return ""
}

// listingInspectionStatusText 检测状态文本
func listingInspectionStatusText(s int) string {
	switch s {
	case model.ListingInspectionNone:
		return "未检测"
	case model.ListingInspectionPending:
		return "待检测"
	case model.ListingInspectionInProgress:
		return "检测中"
	case model.ListingInspectionPassed:
		return "检测通过"
	case model.ListingInspectionFailed:
		return "检测不通过"
	}
	return ""
}

// toListingInfo model -> dto
func toListingInfo(l *model.CarListing) *dto.ListingInfo {
	return &dto.ListingInfo{
		ID:                   l.ID,
		ListingNo:            l.ListingNo,
		CarID:                l.CarID,
		ModelID:              l.ModelID,
		PublisherID:          l.PublisherID,
		PublisherName:        l.PublisherName,
		PublisherAvatar:      l.PublisherAvatar,
		PublisherType:        l.PublisherType,
		DealerID:             l.DealerID,
		DealerName:           l.DealerName,
		ListingType:          l.ListingType,
		Title:                l.Title,
		Description:          l.Description,
		Price:                l.Price,
		OriginalPrice:        l.OriginalPrice,
		PriceNegotiable:      l.PriceNegotiable,
		Status:               l.Status,
		StatusText:           listingStatusText(l.Status),
		AuditStatus:          l.AuditStatus,
		AuditStatusText:      listingAuditStatusText(l.AuditStatus),
		AuditReason:          l.AuditReason,
		PublishedAt:          l.PublishedAt,
		OfflineAt:            l.OfflineAt,
		ExpiredAt:            l.ExpiredAt,
		SoldAt:               l.SoldAt,
		ViewCount:            l.ViewCount,
		FavCount:             l.FavCount,
		ContactCount:         l.ContactCount,
		TestDriveCount:       l.TestDriveCount,
		InspectionStatus:     l.InspectionStatus,
		InspectionStatusText: listingInspectionStatusText(l.InspectionStatus),
		InspectionID:         l.InspectionID,
		RealCarVerified:      l.RealCarVerified,
		Featured:             l.Featured,
		PromotionLevel:       l.PromotionLevel,
		RegionID:             l.RegionID,
		CreatedAt:            l.CreatedAt,
		UpdatedAt:            l.UpdatedAt,
	}
}

// genListingNo 生成发布单号：LS + yyyyMMddHHmmss + 6 位随机
func genListingNo() string {
	return fmt.Sprintf("LS%s%06d", time.Now().Format("20060102150405"), rand.Intn(1000000))
}

// ===== C 端 =====

// Create 发布车源
func (s *listingService) Create(regionID uint, userID uint, userName string, userAvatar string, req *dto.CreateListingRequest) (*dto.ListingInfo, error) {
	l := &model.CarListing{
		ListingNo:       genListingNo(),
		CarID:           req.CarID,
		ModelID:         req.ModelID,
		PublisherID:     userID,
		PublisherName:   userName,
		PublisherAvatar: userAvatar,
		PublisherType:   req.PublisherType,
		DealerID:        req.DealerID,
		DealerName:      req.DealerName,
		ListingType:     req.ListingType,
		Title:           req.Title,
		Description:     req.Description,
		Price:           req.Price,
		OriginalPrice:   req.OriginalPrice,
		PriceNegotiable: req.PriceNegotiable,
		Status:          model.ListingStatusDraft,
		AuditStatus:     model.ListingAuditApproved, // MVP 简化：默认通过
	}
	l.RegionID = regionID

	// 默认值兜底
	if l.PublisherType == "" {
		l.PublisherType = model.PublisherTypePersonal
	}
	if l.ListingType == "" {
		l.ListingType = model.ListingTypeUsed
	}

	if err := s.repo.Create(l); err != nil {
		return nil, err
	}
	return toListingInfo(l), nil
}

// Update 更新发布单（仅发布者本人）
func (s *listingService) Update(id uint, operatorID uint, req *dto.UpdateListingRequest) error {
	l, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrListingNotFound
		}
		return err
	}
	if l.PublisherID != operatorID {
		return ErrListingNoPermission
	}

	fields := map[string]interface{}{}
	if req.Title != nil {
		fields["title"] = *req.Title
	}
	if req.Description != nil {
		fields["description"] = *req.Description
	}
	if req.Price != nil {
		fields["price"] = *req.Price
	}
	if req.OriginalPrice != nil {
		fields["original_price"] = *req.OriginalPrice
	}
	if req.PriceNegotiable != nil {
		fields["price_negotiable"] = *req.PriceNegotiable
	}

	// 状态变更
	if req.Status != nil {
		now := time.Now()
		switch *req.Status {
		case model.ListingStatusPublished:
			if l.Status != model.ListingStatusPublished {
				fields["status"] = model.ListingStatusPublished
				fields["published_at"] = &now
				fields["audit_status"] = model.ListingAuditApproved
			}
		case model.ListingStatusOffline:
			fields["status"] = model.ListingStatusOffline
			fields["offline_at"] = &now
		case model.ListingStatusSold:
			fields["status"] = model.ListingStatusSold
			fields["sold_at"] = &now
		default:
			fields["status"] = *req.Status
		}
	}

	if len(fields) == 0 {
		return nil
	}
	return s.repo.Update(id, fields)
}

// Delete 删除发布单（仅发布者本人）
func (s *listingService) Delete(id uint, operatorID uint) error {
	l, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrListingNotFound
		}
		return err
	}
	if l.PublisherID != operatorID {
		return ErrListingNoPermission
	}
	return s.repo.Delete(id)
}

// GetByID 获取详情（同时增加浏览量）
func (s *listingService) GetByID(id uint) (*dto.ListingInfo, error) {
	l, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrListingNotFound
		}
		return nil, err
	}
	_ = s.repo.IncrViewCount(id)
	l.ViewCount++
	return toListingInfo(l), nil
}

// List C 端列表查询（地区隔离）
func (s *listingService) List(regionID uint, req *dto.ListingListRequest) (*utils.Pagination, []dto.ListingInfo, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	opts := repository.ListingListOptions{
		CarID:            req.CarID,
		PublisherID:      req.PublisherID,
		DealerID:         req.DealerID,
		ListingType:      req.ListingType,
		PublisherType:    req.PublisherType,
		Status:           req.Status,
		AuditStatus:      req.AuditStatus,
		InspectionStatus: req.InspectionStatus,
		Featured:         req.Featured,
		RealCarVerified:  req.RealCarVerified,
		Keyword:          req.Keyword,
	}

	// C 端默认仅展示已发布+审核通过
	if opts.Status == nil {
		published := model.ListingStatusPublished
		opts.Status = &published
	}
	if opts.AuditStatus == nil {
		approved := model.ListingAuditApproved
		opts.AuditStatus = &approved
	}

	list, total, err := s.repo.List(regionID, pagination, opts)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	result := make([]dto.ListingInfo, 0, len(list))
	for i := range list {
		result = append(result, *toListingInfo(&list[i]))
	}
	return pagination, result, nil
}

// ListMine 我的发布
func (s *listingService) ListMine(userID uint, page, pageSize int) (*utils.Pagination, []dto.ListingInfo, error) {
	pagination := utils.NewPagination(page, pageSize)
	list, total, err := s.repo.ListByUser(userID, pagination)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	result := make([]dto.ListingInfo, 0, len(list))
	for i := range list {
		result = append(result, *toListingInfo(&list[i]))
	}
	return pagination, result, nil
}

// ===== M 端管理 =====

func (s *listingService) AdminList(req *dto.ListingAdminListRequest) (*utils.Pagination, []dto.ListingInfo, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	opts := repository.ListingAdminListOptions{
		RegionID:         req.RegionID,
		PublisherID:      req.PublisherID,
		DealerID:         req.DealerID,
		ListingType:      req.ListingType,
		Status:           req.Status,
		AuditStatus:      req.AuditStatus,
		InspectionStatus: req.InspectionStatus,
		Keyword:          req.Keyword,
	}
	list, total, err := s.repo.AdminList(pagination, opts)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	result := make([]dto.ListingInfo, 0, len(list))
	for i := range list {
		result = append(result, *toListingInfo(&list[i]))
	}
	return pagination, result, nil
}

func (s *listingService) AdminGetByID(id uint) (*dto.ListingInfo, error) {
	l, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrListingNotFound
		}
		return nil, err
	}
	return toListingInfo(l), nil
}

// Audit M 端审核
func (s *listingService) Audit(id uint, req *dto.ListingAuditRequest) error {
	l, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrListingNotFound
		}
		return err
	}

	fields := map[string]interface{}{
		"audit_status": req.AuditStatus,
		"audit_reason": req.AuditReason,
	}

	now := time.Now()
	// 审核通过且当前为草稿：自动发布
	if req.AuditStatus == model.ListingAuditApproved && l.Status == model.ListingStatusDraft {
		fields["status"] = model.ListingStatusPublished
		fields["published_at"] = &now
	}
	// 审核拒绝：强制下架
	if req.AuditStatus == model.ListingAuditRejected && l.Status == model.ListingStatusPublished {
		fields["status"] = model.ListingStatusOffline
		fields["offline_at"] = &now
	}

	return s.repo.Update(id, fields)
}

// UpdateStatus M 端强制下架/恢复
func (s *listingService) UpdateStatus(id uint, status int) error {
	now := time.Now()
	fields := map[string]interface{}{
		"status": status,
	}
	switch status {
	case model.ListingStatusPublished:
		fields["published_at"] = &now
		fields["audit_status"] = model.ListingAuditApproved
	case model.ListingStatusOffline:
		fields["offline_at"] = &now
	case model.ListingStatusSold:
		fields["sold_at"] = &now
	}
	return s.repo.Update(id, fields)
}

// UpdateInspectionStatus 更新检测状态
func (s *listingService) UpdateInspectionStatus(id uint, req *dto.InspectionStatusUpdateRequest) error {
	return s.repo.UpdateInspectionStatus(id, req.InspectionStatus, req.InspectionID)
}

// UpdateRealCarVerified 更新真车认证
func (s *listingService) UpdateRealCarVerified(id uint, verified bool) error {
	return s.repo.UpdateRealCarVerified(id, verified)
}
