// Package service 营销活动中台业务逻辑层 - 优惠券（coupon 子域）
package service

import (
	"errors"
	"time"

	"wuchang-tongcheng/internal/modules/marketing/dto"
	"wuchang-tongcheng/internal/modules/marketing/model"
	"wuchang-tongcheng/internal/modules/marketing/repository"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

// CouponService 优惠券业务接口
type CouponService interface {
	// 优惠券 CRUD
	Create(regionID uint, req *dto.CreateCouponRequest) (*dto.CouponInfo, error)
	Update(id uint, req *dto.UpdateCouponRequest) error
	Delete(id uint) error
	GetByID(id uint) (*dto.CouponInfo, error)
	List(regionID uint, req *dto.CouponListRequest) (*utils.Pagination, []dto.CouponInfo, error)
	ListAvailable(regionID uint, page, pageSize int) (*utils.Pagination, []dto.CouponInfo, error)

	// 领取/使用/退还
	Receive(userID uint, couponID uint, source string) error
	Use(userCouponID uint, orderID uint) error
	Refund(userCouponID uint) error

	// 我的优惠券
	ListMine(userID uint, req *dto.UserCouponListRequest) (*utils.Pagination, []dto.UserCouponInfo, error)

	// 统计 & 过期检查
	Statistics(regionID uint) (*dto.CouponStatistics, error)
	ExpireCoupons() (int64, error)
}

type couponService struct {
	repo repository.CouponRepository
}

// NewCouponService 创建优惠券 service 实例
func NewCouponService(repo repository.CouponRepository) CouponService {
	return &couponService{repo: repo}
}

// couponTypeText 优惠券类型文本
func couponTypeText(t string) string {
	switch t {
	case model.CouponTypeDiscount:
		return "折扣券"
	case model.CouponTypeReduce:
		return "满减券"
	case model.CouponTypeExchange:
		return "兑换券"
	}
	return ""
}

// couponStatusText 优惠券状态文本
func couponStatusText(s int) string {
	switch s {
	case model.CouponStatusDisabled:
		return "禁用"
	case model.CouponStatusActive:
		return "进行中"
	case model.CouponStatusDraft:
		return "草稿"
	case model.CouponStatusOffline:
		return "已下架"
	case model.CouponStatusExpired:
		return "已过期"
	case model.CouponStatusSoldOut:
		return "已抢完"
	}
	return ""
}

// userCouponStatusText 用户优惠券状态文本
func userCouponStatusText(s string) string {
	switch s {
	case model.UserCouponStatusUnused:
		return "未使用"
	case model.UserCouponStatusUsed:
		return "已使用"
	case model.UserCouponStatusExpired:
		return "已过期"
	}
	return ""
}

// userCouponSourceText 用户优惠券来源文本
func userCouponSourceText(s string) string {
	switch s {
	case model.UserCouponSourceReceive:
		return "主动领取"
	case model.UserCouponSourceGift:
		return "系统赠送"
	case model.UserCouponSourceActivity:
		return "活动奖励"
	case model.UserCouponSourceNewUser:
		return "新人礼包"
	}
	return ""
}

// toCouponInfo model -> dto
func toCouponInfo(c *model.Coupon) *dto.CouponInfo {
	return &dto.CouponInfo{
		ID:            c.ID,
		RegionID:      c.RegionID,
		Title:         c.Title,
		Type:          c.Type,
		TypeText:      couponTypeText(c.Type),
		Amount:        c.Amount,
		Threshold:     c.Threshold,
		TotalCount:    c.TotalCount,
		ReceivedCount: c.ReceivedCount,
		StartAt:       c.StartAt,
		EndAt:         c.EndAt,
		Status:        c.Status,
		StatusText:    couponStatusText(c.Status),
		CreatedAt:     c.CreatedAt,
		UpdatedAt:     c.UpdatedAt,
	}
}

// toUserCouponInfo model -> dto
func toUserCouponInfo(uc *model.UserCoupon) *dto.UserCouponInfo {
	return &dto.UserCouponInfo{
		ID:         uc.ID,
		UserID:     uc.UserID,
		CouponID:   uc.CouponID,
		Status:     uc.Status,
		StatusText: userCouponStatusText(uc.Status),
		Source:     uc.Source,
		SourceText: userCouponSourceText(uc.Source),
		UsedAt:     uc.UsedAt,
		OrderID:    uc.OrderID,
		CreatedAt:  uc.CreatedAt,
		UpdatedAt:  uc.UpdatedAt,
	}
}

// Create 创建优惠券
func (s *couponService) Create(regionID uint, req *dto.CreateCouponRequest) (*dto.CouponInfo, error) {
	status := req.Status
	if status == 0 {
		status = model.CouponStatusActive
	}
	c := &model.Coupon{
		Title:      req.Title,
		Type:       req.Type,
		Amount:     req.Amount,
		Threshold:  req.Threshold,
		TotalCount: req.TotalCount,
		StartAt:    req.StartAt,
		EndAt:      req.EndAt,
		Status:     status,
	}
	c.RegionID = regionID
	if err := s.repo.Create(c); err != nil {
		return nil, err
	}
	return toCouponInfo(c), nil
}

// Update 更新优惠券
func (s *couponService) Update(id uint, req *dto.UpdateCouponRequest) error {
	if _, err := s.repo.FindByID(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrCouponNotFound
		}
		return err
	}
	fields := make(map[string]interface{})
	if req.Title != nil {
		fields["title"] = *req.Title
	}
	if req.Type != nil {
		fields["type"] = *req.Type
	}
	if req.Amount != nil {
		fields["amount"] = *req.Amount
	}
	if req.Threshold != nil {
		fields["threshold"] = *req.Threshold
	}
	if req.TotalCount != nil {
		fields["total_count"] = *req.TotalCount
	}
	if req.StartAt != nil {
		fields["start_at"] = req.StartAt
	}
	if req.EndAt != nil {
		fields["end_at"] = req.EndAt
	}
	if req.Status != nil {
		fields["status"] = *req.Status
	}
	if len(fields) == 0 {
		return nil
	}
	return s.repo.Update(id, fields)
}

// Delete 删除优惠券
func (s *couponService) Delete(id uint) error {
	if _, err := s.repo.FindByID(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrCouponNotFound
		}
		return err
	}
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
		Type:    req.Type,
		Status:  req.Status,
		Keyword: req.Keyword,
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

// ListAvailable 可领取优惠券列表
func (s *couponService) ListAvailable(regionID uint, page, pageSize int) (*utils.Pagination, []dto.CouponInfo, error) {
	pagination := utils.NewPagination(page, pageSize)
	list, total, err := s.repo.ListAvailable(regionID, pagination)
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

// Receive 领取优惠券
func (s *couponService) Receive(userID uint, couponID uint, source string) error {
	c, err := s.repo.FindByID(couponID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrCouponNotFound
		}
		return err
	}
	if c.Status != model.CouponStatusActive {
		return ErrCouponStatusInvalid
	}
	now := time.Now()
	if c.StartAt != nil && c.StartAt.After(now) {
		return ErrCouponNotStarted
	}
	if c.EndAt != nil && c.EndAt.Before(now) {
		return ErrCouponExpired
	}
	// 库存检查
	if c.TotalCount > 0 && c.ReceivedCount >= c.TotalCount {
		return ErrCouponSoldOut
	}
	// 每人限领 1 张（同一张券）
	count, err := s.repo.CountUserCoupon(userID, couponID)
	if err != nil {
		return err
	}
	if count > 0 {
		return ErrCouponAlreadyRecv
	}
	// 落库用户券
	if source == "" {
		source = model.UserCouponSourceReceive
	}
	uc := &model.UserCoupon{
		UserID:   userID,
		CouponID: couponID,
		Status:   model.UserCouponStatusUnused,
		Source:   source,
	}
	if err := s.repo.CreateUserCoupon(uc); err != nil {
		return err
	}
	// 增加已领取数
	return s.repo.IncrReceivedCount(couponID)
}

// Use 使用优惠券（user_coupon_id）
func (s *couponService) Use(userCouponID uint, orderID uint) error {
	uc, err := s.repo.FindUserCouponByID(userCouponID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrUserCouponNotFound
		}
		return err
	}
	if uc.Status == model.UserCouponStatusUsed {
		return ErrUserCouponUsed
	}
	if uc.Status == model.UserCouponStatusExpired {
		return ErrUserCouponExpired
	}
	now := time.Now()
	fields := map[string]interface{}{
		"status":   model.UserCouponStatusUsed,
		"used_at":  &now,
		"order_id": orderID,
	}
	return s.repo.UpdateUserCoupon(userCouponID, fields)
}

// Refund 退还优惠券（订单退款时调用）
func (s *couponService) Refund(userCouponID uint) error {
	uc, err := s.repo.FindUserCouponByID(userCouponID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrUserCouponNotFound
		}
		return err
	}
	if uc.Status != model.UserCouponStatusUsed {
		return ErrUserCouponUsed
	}
	fields := map[string]interface{}{
		"status":  model.UserCouponStatusUnused,
		"used_at": nil,
	}
	return s.repo.UpdateUserCoupon(userCouponID, fields)
}

// ListMine 我的优惠券列表
func (s *couponService) ListMine(userID uint, req *dto.UserCouponListRequest) (*utils.Pagination, []dto.UserCouponInfo, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	query := repository.UserCouponListQuery{
		UserID: userID,
		Status: req.Status,
	}
	list, total, err := s.repo.ListUserCoupons(query, pagination)
	if err != nil {
		return nil, nil, err
	}
	infos := make([]dto.UserCouponInfo, 0, len(list))
	for i := range list {
		info := *toUserCouponInfo(&list[i])
		// 冗余优惠券信息
		if c, err := s.repo.FindByID(list[i].CouponID); err == nil {
			info.CouponTitle = c.Title
			info.CouponType = c.Type
			info.CouponAmount = c.Amount
		}
		infos = append(infos, info)
	}
	pagination.Total = total
	return pagination, infos, nil
}

// Statistics 优惠券领取统计
func (s *couponService) Statistics(regionID uint) (*dto.CouponStatistics, error) {
	// 总数
	_, total, err := s.repo.List(regionID, repository.CouponListQuery{}, utils.NewPagination(1, 1))
	if err != nil {
		return nil, err
	}
	// 进行中
	activeStatus := model.CouponStatusActive
	_, active, err := s.repo.List(regionID, repository.CouponListQuery{Status: &activeStatus}, utils.NewPagination(1, 1))
	if err != nil {
		return nil, err
	}
	// 领取/使用总数：遍历进行中优惠券汇总（简化实现）
	allList, _, err := s.repo.List(regionID, repository.CouponListQuery{}, utils.NewPagination(1, 100))
	if err != nil {
		return nil, err
	}
	var received, used int64
	for i := range allList {
		received += int64(allList[i].ReceivedCount)
		used += int64(allList[i].ReceivedCount) // MVP：使用数以用户券已使用为准，这里用 received 近似占位
	}
	// 实际使用数：查询 user_coupons 已使用
	usedList, usedCount, err := s.repo.ListUserCoupons(repository.UserCouponListQuery{Status: model.UserCouponStatusUsed}, utils.NewPagination(1, 1))
	if err == nil {
		used = usedCount
	}
	_ = usedList

	stats := &dto.CouponStatistics{
		TotalCoupons:  total,
		ActiveCoupons: active,
		TotalReceived: received,
		TotalUsed:     used,
	}
	if total > 0 {
		stats.ReceiveRate = float64(received) / float64(total)
	}
	if received > 0 {
		stats.UsageRate = float64(used) / float64(received)
	}
	return stats, nil
}

// ExpireCoupons 过期检查（定时任务调用）
// - 将已过期的优惠券状态置为已过期
// - 将未使用且对应优惠券已过期的用户券标记为已过期
func (s *couponService) ExpireCoupons() (int64, error) {
	now := time.Now()
	// 1. 优惠券过期：状态为进行中且 end_at < now → 已过期
	list, _, err := s.repo.List(0, repository.CouponListQuery{Status: intPtrMarketingStatus(model.CouponStatusActive)}, utils.NewPagination(1, 100))
	if err != nil {
		return 0, err
	}
	var affected int64
	for i := range list {
		if list[i].EndAt != nil && list[i].EndAt.Before(now) {
			if err := s.repo.Update(list[i].ID, map[string]interface{}{"status": model.CouponStatusExpired}); err == nil {
				affected++
			}
		}
	}
	// 2. 用户券过期
	n, err := s.repo.ExpireUserCoupons(now)
	if err != nil {
		return affected, err
	}
	affected += n
	return affected, nil
}
