// Package service 同城114业务逻辑层 - 优惠券
// 依据 v3.2.1 架构方案：对标大众点评/美团
// 满减/折扣/代金券/礼品券
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
	ErrCouponNotFound      = errors.New("优惠券不存在")
	ErrCouponNoPermission  = errors.New("无权操作此优惠券")
	ErrCouponStatusInvalid = errors.New("优惠券状态不允许此操作")
	ErrCouponSoldOut       = errors.New("优惠券已抢完")
	ErrCouponExpired       = errors.New("优惠券已过期")
)

// CouponService 优惠券业务接口
type CouponService interface {
	// C 端
	Create(regionID uint, userID uint, req *dto.CreateCouponRequest) (*dto.CouponInfo, error)
	Update(id uint, operatorID uint, req *dto.UpdateCouponRequest) error
	Delete(id uint, operatorID uint) error
	GetByID(id uint) (*dto.CouponInfo, error)
	List(regionID uint, req *dto.CouponListRequest) (*utils.Pagination, []dto.CouponInfo, error)
	ListByDh114(regionID uint, dh114ID uint, page, pageSize int) (*utils.Pagination, []dto.CouponInfo, error)
	ListHot(regionID uint, page, pageSize int) (*utils.Pagination, []dto.CouponInfo, error)

	// 领取/使用
	Receive(userID uint, couponID uint) error
	Use(userID uint, couponID uint) error

	// M 端管理
	AdminList(req *dto.CouponAdminListRequest) (*utils.Pagination, []dto.CouponInfo, error)
	Audit(id uint, auditStatus int, auditReason string) error
	BatchAudit(req *dto.BatchAuditRequest) (*dto.BatchResultResponse, error)
	AdminUpdateStatus(id uint, status int) error
}

type couponService struct {
	repo repository.CouponRepository
}

// NewCouponService 创建优惠券 service 实例
func NewCouponService(repo repository.CouponRepository) CouponService {
	return &couponService{repo: repo}
}

// couponStatusText 优惠券状态文本
func couponStatusText(s int) string {
	switch s {
	case model.CouponStatusDraft:
		return "草稿"
	case model.CouponStatusPublished:
		return "已发布"
	case model.CouponStatusSoldOut:
		return "已抢完"
	case model.CouponStatusOffline:
		return "已下架"
	case model.CouponStatusExpired:
		return "已过期"
	}
	return ""
}

// couponTypeText 优惠券类型文本
func couponTypeText(t string) string {
	switch t {
	case model.CouponTypeDiscount:
		return "折扣券"
	case model.CouponTypeFullReduction:
		return "满减券"
	case model.CouponTypeCash:
		return "代金券"
	case model.CouponTypeGift:
		return "礼品券"
	}
	return ""
}

// toCouponInfo model -> dto
func toCouponInfo(c *model.Dh114Coupon) *dto.CouponInfo {
	info := &dto.CouponInfo{
		ID:              c.ID,
		CouponNo:        c.CouponNo,
		Dh114ID:         c.Dh114ID,
		BusinessID:      c.BusinessID,
		Title:           c.Title,
		Description:     c.Description,
		CoverImage:      c.CoverImage,
		CouponType:      c.CouponType,
		CouponTypeText:  couponTypeText(c.CouponType),
		FaceValue:       c.FaceValue,
		Threshold:       c.Threshold,
		Discount:        c.Discount,
		MaxDiscount:     c.MaxDiscount,
		TotalCount:      c.TotalCount,
		IssuedCount:     c.IssuedCount,
		UsedCount:       c.UsedCount,
		PerUserLimit:    c.PerUserLimit,
		StartTime:       c.StartTime,
		EndTime:         c.EndTime,
		ValidStart:      c.ValidStart,
		ValidEnd:        c.ValidEnd,
		ValidDays:       c.ValidDays,
		UseThreshold:    c.UseThreshold,
		Status:          c.Status,
		StatusText:      couponStatusText(c.Status),
		AuditStatus:     c.AuditStatus,
		AuditReason:     c.AuditReason,
		PublishedAt:     c.PublishedAt,
		Featured:        c.Featured,
		RegionID:        c.RegionID,
		CreatedAt:       c.CreatedAt,
		UpdatedAt:       c.UpdatedAt,
	}
	if c.UseInstructions != nil {
		info.UseInstructions = c.UseInstructions
	}
	return info
}

// generateCouponNo 生成优惠券单号
func generateCouponNo() string {
	return fmt.Sprintf("DH114CP%s%06d", time.Now().Format("20060102150405"), rand.Intn(1000000))
}

// Create 创建优惠券
func (s *couponService) Create(regionID uint, userID uint, req *dto.CreateCouponRequest) (*dto.CouponInfo, error) {
	status := req.Status
	if status == 0 {
		status = model.CouponStatusDraft
	}
	perUserLimit := req.PerUserLimit
	if perUserLimit == 0 {
		perUserLimit = 1
	}

	c := &model.Dh114Coupon{
		CouponNo:     generateCouponNo(),
		Dh114ID:      req.Dh114ID,
		Title:        req.Title,
		Description:  req.Description,
		CoverImage:   req.CoverImage,
		CouponType:   req.CouponType,
		FaceValue:    req.FaceValue,
		Threshold:    req.Threshold,
		Discount:     req.Discount,
		MaxDiscount:  req.MaxDiscount,
		TotalCount:   req.TotalCount,
		PerUserLimit: perUserLimit,
		StartTime:    req.StartTime,
		EndTime:      req.EndTime,
		ValidStart:   req.ValidStart,
		ValidEnd:     req.ValidEnd,
		ValidDays:    req.ValidDays,
		UseThreshold: req.UseThreshold,
		Status:       status,
		AuditStatus:  model.AuditApproved, // MVP：发布即通过
	}
	c.RegionID = regionID

	if req.UseInstructions != nil {
		if b, err := model.FromJSON(req.UseInstructions); err == nil {
			c.UseInstructions = b
		}
	}

	if c.Status == model.CouponStatusPublished {
		now := time.Now()
		c.PublishedAt = &now
	}

	if err := s.repo.Create(c); err != nil {
		return nil, err
	}
	return toCouponInfo(c), nil
}

// Update 更新优惠券
func (s *couponService) Update(id uint, operatorID uint, req *dto.UpdateCouponRequest) error {
	c, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrCouponNotFound
		}
		return err
	}
	_ = operatorID
	_ = c

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
	if req.FaceValue != nil {
		fields["face_value"] = *req.FaceValue
	}
	if req.Threshold != nil {
		fields["threshold"] = *req.Threshold
	}
	if req.Discount != nil {
		fields["discount"] = *req.Discount
	}
	if req.MaxDiscount != nil {
		fields["max_discount"] = *req.MaxDiscount
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
	if req.ValidDays != nil {
		fields["valid_days"] = *req.ValidDays
	}
	if req.UseThreshold != nil {
		fields["use_threshold"] = *req.UseThreshold
	}
	if req.Status != nil {
		fields["status"] = *req.Status
		if *req.Status == model.CouponStatusPublished && c.PublishedAt == nil {
			now := time.Now()
			fields["published_at"] = &now
		}
	}
	if req.UseInstructions != nil {
		if b, err := model.FromJSON(req.UseInstructions); err == nil {
			fields["use_instructions"] = b
		}
	}

	if len(fields) == 0 {
		return nil
	}
	return s.repo.Update(id, fields)
}

// Delete 删除优惠券
func (s *couponService) Delete(id uint, operatorID uint) error {
	if _, err := s.repo.FindByID(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrCouponNotFound
		}
		return err
	}
	_ = operatorID
	return s.repo.Delete(id)
}

// GetByID 获取优惠券详情
func (s *couponService) GetByID(id uint) (*dto.CouponInfo, error) {
	c, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCouponNotFound
		}
		return nil, err
	}
	return toCouponInfo(c), nil
}

// List 优惠券列表
func (s *couponService) List(regionID uint, req *dto.CouponListRequest) (*utils.Pagination, []dto.CouponInfo, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	query := repository.CouponListQuery{
		Dh114ID:    req.Dh114ID,
		CouponType: req.CouponType,
		Featured:    req.Featured,
		Keyword:     req.Keyword,
	}
	if req.Status != nil {
		query.Status = req.Status
	}
	list, total, err := s.repo.List(regionID, query, pagination)
	if err != nil {
		return nil, nil, err
	}
	infos := make([]dto.CouponInfo, 0, len(list))
	for i := range list {
		infos = append(infos, *toCouponInfo(&list[i]))
	}
	pagination.Total = total
	return pagination, infos, nil
}

// ListByDh114 按商户列出优惠券
func (s *couponService) ListByDh114(regionID uint, dh114ID uint, page, pageSize int) (*utils.Pagination, []dto.CouponInfo, error) {
	pagination := utils.NewPagination(page, pageSize)
	list, total, err := s.repo.ListByDh114(regionID, dh114ID, pagination)
	if err != nil {
		return nil, nil, err
	}
	infos := make([]dto.CouponInfo, 0, len(list))
	for i := range list {
		infos = append(infos, *toCouponInfo(&list[i]))
	}
	pagination.Total = total
	return pagination, infos, nil
}

// ListHot 热门优惠券
func (s *couponService) ListHot(regionID uint, page, pageSize int) (*utils.Pagination, []dto.CouponInfo, error) {
	pagination := utils.NewPagination(page, pageSize)
	list, total, err := s.repo.ListHot(regionID, pagination)
	if err != nil {
		return nil, nil, err
	}
	infos := make([]dto.CouponInfo, 0, len(list))
	for i := range list {
		infos = append(infos, *toCouponInfo(&list[i]))
	}
	pagination.Total = total
	return pagination, infos, nil
}

// Receive 领取优惠券（增加已领取数）
func (s *couponService) Receive(userID uint, couponID uint) error {
	c, err := s.repo.FindByID(couponID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrCouponNotFound
		}
		return err
	}
	if c.Status != model.CouponStatusPublished {
		return ErrCouponStatusInvalid
	}
	// 检查是否过期
	now := time.Now()
	if c.EndTime != nil && c.EndTime.Before(now) {
		return ErrCouponExpired
	}
	// 检查库存
	if c.TotalCount > 0 && c.IssuedCount >= c.TotalCount {
		return ErrCouponSoldOut
	}
	_ = userID
	return s.repo.IncrIssuedCount(couponID)
}

// Use 使用优惠券（增加已使用数）
func (s *couponService) Use(userID uint, couponID uint) error {
	c, err := s.repo.FindByID(couponID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrCouponNotFound
		}
		return err
	}
	if c.Status != model.CouponStatusPublished {
		return ErrCouponStatusInvalid
	}
	_ = userID
	return s.repo.IncrUsedCount(couponID)
}

// AdminList 管理后台列表
func (s *couponService) AdminList(req *dto.CouponAdminListRequest) (*utils.Pagination, []dto.CouponInfo, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	query := repository.CouponAdminListQuery{
		Dh114ID:     req.Dh114ID,
		Status:      req.Status,
		AuditStatus: req.AuditStatus,
		Keyword:     req.Keyword,
	}
	list, total, err := s.repo.AdminList(query, pagination)
	if err != nil {
		return nil, nil, err
	}
	infos := make([]dto.CouponInfo, 0, len(list))
	for i := range list {
		infos = append(infos, *toCouponInfo(&list[i]))
	}
	pagination.Total = total
	return pagination, infos, nil
}

// Audit 审核优惠券
func (s *couponService) Audit(id uint, auditStatus int, auditReason string) error {
	if _, err := s.repo.FindByID(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrCouponNotFound
		}
		return err
	}
	fields := map[string]interface{}{
		"audit_status": auditStatus,
		"audit_reason": auditReason,
	}
	if auditStatus == model.AuditApproved {
		c, err := s.repo.FindByID(id)
		if err == nil && c.PublishedAt == nil {
			now := time.Now()
			fields["published_at"] = &now
			if c.Status == model.CouponStatusDraft {
				fields["status"] = model.CouponStatusPublished
			}
		}
	}
	return s.repo.Update(id, fields)
}

// BatchAudit 批量审核
func (s *couponService) BatchAudit(req *dto.BatchAuditRequest) (*dto.BatchResultResponse, error) {
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
func (s *couponService) AdminUpdateStatus(id uint, status int) error {
	if status < model.CouponStatusDraft || status > model.CouponStatusExpired {
		return ErrCouponStatusInvalid
	}
	fields := map[string]interface{}{"status": status}
	if status == model.CouponStatusPublished {
		c, err := s.repo.FindByID(id)
		if err == nil && c.PublishedAt == nil {
			now := time.Now()
			fields["published_at"] = &now
		}
	}
	return s.repo.Update(id, fields)
}
