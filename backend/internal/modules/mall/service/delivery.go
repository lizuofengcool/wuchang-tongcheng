// Package service 同城商城业务逻辑层 - 配送单
package service

import (
	"errors"
	"time"

	"wuchang-tongcheng/internal/modules/mall/dto"
	"wuchang-tongcheng/internal/modules/mall/model"
	"wuchang-tongcheng/internal/modules/mall/repository"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

var (
	// ErrDeliveryNotFound 配送单不存在
	ErrDeliveryNotFound = errors.New("配送单不存在")
	// ErrDeliveryGrabbed 配送单已被抢
	ErrDeliveryGrabbed = errors.New("配送单已被抢")
	// ErrDeliveryStatusInvalid 配送状态不允许此操作
	ErrDeliveryStatusInvalid = errors.New("配送状态不允许此操作")
)

// DeliveryService 配送单业务接口
type DeliveryService interface {
	Create(regionID uint, req *dto.DeliveryCreateRequest) (*dto.DeliveryInfo, error)
	GetByID(id uint) (*dto.DeliveryInfo, error)
	GetByDeliveryNo(no string) (*dto.DeliveryInfo, error)
	List(regionID uint, req *dto.DeliveryListRequest) (*utils.Pagination, []dto.DeliveryInfo, error)
	ListByRider(userID uint, req *dto.DeliveryListRequest) (*utils.Pagination, []dto.DeliveryInfo, error)
	AdminList(req *dto.DeliveryListRequest) (*utils.Pagination, []dto.DeliveryInfo, error)

	// 状态流转（C 端骑手操作）
	Grab(id, userID uint) error
	ArriveShop(id, userID uint) error
	Pickup(id, userID uint) error
	Deliver(id, userID uint) error
	Complete(id, userID uint) error
	Cancel(id, userID uint, req *dto.DeliveryCancelRequest) error

	// 统计
	Stats(userID uint) (*dto.DeliveryStatsResponse, error)
}

type deliveryService struct {
	repo repository.DeliveryRepository
	riderRepo repository.RiderRepository
}

// NewDeliveryService 创建配送单 service 实例
func NewDeliveryService(repo repository.DeliveryRepository, riderRepo repository.RiderRepository) DeliveryService {
	return &deliveryService{repo: repo, riderRepo: riderRepo}
}

func deliveryStatusText(s int) string {
	switch s {
	case model.DeliveryStatusPending:
		return "待接单"
	case model.DeliveryStatusAccepted:
		return "已接单"
	case model.DeliveryStatusArrived:
		return "已到店"
	case model.DeliveryStatusPicked:
		return "已取货"
	case model.DeliveryStatusDelivering:
		return "配送中"
	case model.DeliveryStatusDelivered:
		return "已送达"
	case model.DeliveryStatusCancelled:
		return "已取消"
	}
	return ""
}

func toDeliveryInfo(d *model.Delivery) *dto.DeliveryInfo {
	return &dto.DeliveryInfo{
		ID:              d.ID,
		OrderID:         d.OrderID,
		RiderID:         d.RiderID,
		ShopID:          d.ShopID,
		UserID:          d.UserID,
		DeliveryNo:      d.DeliveryNo,
		Status:          d.Status,
		StatusText:      deliveryStatusText(d.Status),
		PickupAddress:   d.PickupAddress,
		PickupLat:       d.PickupLat,
		PickupLng:       d.PickupLng,
		DeliveryAddress: d.DeliveryAddress,
		DeliveryLat:     d.DeliveryLat,
		DeliveryLng:     d.DeliveryLng,
		Distance:        d.Distance,
		DeliveryFee:     d.DeliveryFee,
		Tip:             d.Tip,
		AcceptedAt:      d.AcceptedAt,
		PickedAt:        d.PickedAt,
		DeliveredAt:     d.DeliveredAt,
		CancelReason:    d.CancelReason,
		RegionID:        d.RegionID,
		CreatedAt:       d.CreatedAt,
		UpdatedAt:       d.UpdatedAt,
	}
}

// Create 创建配送单
func (s *deliveryService) Create(regionID uint, req *dto.DeliveryCreateRequest) (*dto.DeliveryInfo, error) {
	d := &model.Delivery{
		OrderID:         req.OrderID,
		ShopID:          req.ShopID,
		UserID:          req.UserID,
		DeliveryNo:      req.DeliveryNo,
		Status:          model.DeliveryStatusPending,
		PickupAddress:   req.PickupAddress,
		PickupLat:       req.PickupLat,
		PickupLng:       req.PickupLng,
		DeliveryAddress: req.DeliveryAddress,
		DeliveryLat:     req.DeliveryLat,
		DeliveryLng:     req.DeliveryLng,
		Distance:        req.Distance,
		DeliveryFee:     req.DeliveryFee,
		Tip:             req.Tip,
	}
	d.RegionID = regionID
	if err := s.repo.Create(d); err != nil {
		return nil, err
	}
	return toDeliveryInfo(d), nil
}

// GetByID 配送单详情
func (s *deliveryService) GetByID(id uint) (*dto.DeliveryInfo, error) {
	d, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrDeliveryNotFound
		}
		return nil, err
	}
	return toDeliveryInfo(d), nil
}

// GetByDeliveryNo 根据配送单号查询
func (s *deliveryService) GetByDeliveryNo(no string) (*dto.DeliveryInfo, error) {
	d, err := s.repo.FindByDeliveryNo(no)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrDeliveryNotFound
		}
		return nil, err
	}
	return toDeliveryInfo(d), nil
}

// List 配送单列表（抢单大厅）
func (s *deliveryService) List(regionID uint, req *dto.DeliveryListRequest) (*utils.Pagination, []dto.DeliveryInfo, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	opts := repository.DeliveryListOptions{
		Keyword:    req.Keyword,
		Status:     req.Status,
		ShopID:     req.ShopID,
		UserID:     req.UserID,
		DeliveryNo: req.DeliveryNo,
	}
	list, total, err := s.repo.List(regionID, pagination, opts)
	if err != nil {
		return nil, nil, err
	}
	infos := make([]dto.DeliveryInfo, 0, len(list))
	for i := range list {
		infos = append(infos, *toDeliveryInfo(&list[i]))
	}
	pagination.Total = total
	return pagination, infos, nil
}

// ListByRider 我的配送单（按当前用户→骑手 ID 查询）
func (s *deliveryService) ListByRider(userID uint, req *dto.DeliveryListRequest) (*utils.Pagination, []dto.DeliveryInfo, error) {
	rider, err := s.riderRepo.FindByUserID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, ErrRiderNotFound
		}
		return nil, nil, err
	}
	pagination := utils.NewPagination(req.Page, req.PageSize)
	opts := repository.DeliveryListOptions{
		Keyword:    req.Keyword,
		Status:     req.Status,
		DeliveryNo: req.DeliveryNo,
	}
	list, total, err := s.repo.ListByRider(rider.ID, pagination, opts)
	if err != nil {
		return nil, nil, err
	}
	infos := make([]dto.DeliveryInfo, 0, len(list))
	for i := range list {
		infos = append(infos, *toDeliveryInfo(&list[i]))
	}
	pagination.Total = total
	return pagination, infos, nil
}

// AdminList 管理后台配送单列表
func (s *deliveryService) AdminList(req *dto.DeliveryListRequest) (*utils.Pagination, []dto.DeliveryInfo, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	opts := repository.DeliveryAdminListOptions{
		RegionID:   req.RegionID,
		RiderID:    req.RiderID,
		ShopID:     req.ShopID,
		UserID:     req.UserID,
		Status:     req.Status,
		Keyword:    req.Keyword,
		DeliveryNo: req.DeliveryNo,
	}
	list, total, err := s.repo.AdminList(pagination, opts)
	if err != nil {
		return nil, nil, err
	}
	infos := make([]dto.DeliveryInfo, 0, len(list))
	for i := range list {
		infos = append(infos, *toDeliveryInfo(&list[i]))
	}
	pagination.Total = total
	return pagination, infos, nil
}

// Grab 骑手抢单
// 业务规则：检查骑手在线 + 信用分 > 50，更新 delivery.status=1，rider.online_status=2
func (s *deliveryService) Grab(id, userID uint) error {
	rider, err := s.riderRepo.FindByUserID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrRiderNotFound
		}
		return err
	}
	if rider.Status != model.RiderStatusApproved {
		return ErrRiderNotVerified
	}
	if rider.Status == model.RiderStatusFrozen {
		return ErrRiderFrozen
	}
	if rider.OnlineStatus != model.RiderOnline {
		return ErrRiderOffline
	}
	if rider.CreditScore <= 50 {
		return ErrRiderInsufficientCredit
	}

	d, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrDeliveryNotFound
		}
		return err
	}
	if d.Status != model.DeliveryStatusPending {
		return ErrDeliveryGrabbed
	}

	// 事务：更新配送单状态 + 骑手在线状态
	now := time.Now()
	fields := map[string]interface{}{
		"rider_id":    rider.ID,
		"accepted_at": &now,
	}
	if err := s.repo.UpdateStatus(id, model.DeliveryStatusAccepted, fields); err != nil {
		return err
	}
	if err := s.riderRepo.UpdateOnlineStatus(rider.ID, model.RiderDelivering); err != nil {
		// 回滚配送单状态
		_ = s.repo.UpdateStatus(id, model.DeliveryStatusPending, map[string]interface{}{"rider_id": nil, "accepted_at": nil})
		return err
	}
	return nil
}

// ArriveShop 骑手到店
func (s *deliveryService) ArriveShop(id, userID uint) error {
	d, riderID, err := s.checkRiderDelivery(id, userID)
	if err != nil {
		return err
	}
	if d.Status != model.DeliveryStatusAccepted {
		return ErrDeliveryStatusInvalid
	}
	_ = riderID
	return s.repo.UpdateStatus(id, model.DeliveryStatusArrived, nil)
}

// Pickup 骑手取货
// 业务规则：更新 delivery.status=3，记录 picked_at
func (s *deliveryService) Pickup(id, userID uint) error {
	d, riderID, err := s.checkRiderDelivery(id, userID)
	if err != nil {
		return err
	}
	if d.Status != model.DeliveryStatusArrived {
		return ErrDeliveryStatusInvalid
	}
	_ = riderID
	now := time.Now()
	fields := map[string]interface{}{
		"picked_at": &now,
	}
	return s.repo.UpdateStatus(id, model.DeliveryStatusPicked, fields)
}

// Deliver 开始配送
func (s *deliveryService) Deliver(id, userID uint) error {
	d, riderID, err := s.checkRiderDelivery(id, userID)
	if err != nil {
		return err
	}
	if d.Status != model.DeliveryStatusPicked {
		return ErrDeliveryStatusInvalid
	}
	_ = riderID
	return s.repo.UpdateStatus(id, model.DeliveryStatusDelivering, nil)
}

// Complete 送达
// 业务规则：更新 delivery.status=5，记录 delivered_at，更新 rider.total_orders + total_earnings
func (s *deliveryService) Complete(id, userID uint) error {
	d, riderID, err := s.checkRiderDelivery(id, userID)
	if err != nil {
		return err
	}
	if d.Status != model.DeliveryStatusDelivering {
		return ErrDeliveryStatusInvalid
	}

	now := time.Now()
	fields := map[string]interface{}{
		"delivered_at": &now,
	}
	if err := s.repo.UpdateStatus(id, model.DeliveryStatusDelivered, fields); err != nil {
		return err
	}

	// 更新骑手累计订单 + 收益（配送费 + 小费）
	earnings := d.DeliveryFee + d.Tip
	if err := s.riderRepo.IncrTotalOrders(riderID, earnings); err != nil {
		return err
	}

	// 骑手回到在线状态
	_ = s.riderRepo.UpdateOnlineStatus(riderID, model.RiderOnline)
	return nil
}

// Cancel 取消配送单
func (s *deliveryService) Cancel(id, userID uint, req *dto.DeliveryCancelRequest) error {
	d, riderID, err := s.checkRiderDelivery(id, userID)
	if err != nil {
		return err
	}
	if d.Status == model.DeliveryStatusDelivered || d.Status == model.DeliveryStatusCancelled {
		return ErrDeliveryStatusInvalid
	}
	reason := req.CancelReason
	fields := map[string]interface{}{
		"cancel_reason": reason,
	}
	if err := s.repo.UpdateStatus(id, model.DeliveryStatusCancelled, fields); err != nil {
		return err
	}
	// 骑手回到在线状态
	_ = s.riderRepo.UpdateOnlineStatus(riderID, model.RiderOnline)
	return nil
}

// Stats 配送统计
func (s *deliveryService) Stats(userID uint) (*dto.DeliveryStatsResponse, error) {
	rider, err := s.riderRepo.FindByUserID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrRiderNotFound
		}
		return nil, err
	}

	resp := &dto.DeliveryStatsResponse{
		TotalOrders:   int64(rider.TotalOrders),
		TotalEarnings: rider.TotalEarnings,
	}

	// 各状态计数
	if cnt, err := s.repo.CountByStatus(rider.ID, model.DeliveryStatusDelivered); err == nil {
		resp.CompletedCount = cnt
	}
	if cnt, err := s.repo.CountByStatus(rider.ID, model.DeliveryStatusCancelled); err == nil {
		resp.CancelledCount = cnt
	}
	// 进行中订单数（已接单/到店/取货/配送中）
	for _, st := range []int{
		model.DeliveryStatusAccepted,
		model.DeliveryStatusArrived,
		model.DeliveryStatusPicked,
		model.DeliveryStatusDelivering,
	} {
		if cnt, err := s.repo.CountByStatus(rider.ID, st); err == nil {
			resp.PendingCount += cnt
		}
	}

	// 今日统计
	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	todayEnd := todayStart.AddDate(0, 0, 1)
	todayEarnings, todayOrders, err := s.repo.SumEarnings(rider.ID, todayStart, todayEnd)
	if err == nil {
		resp.TodayEarnings = todayEarnings
		resp.TodayOrders = todayOrders
	}

	return resp, nil
}

// checkRiderDelivery 校验当前用户为指定配送单的认领骑手
func (s *deliveryService) checkRiderDelivery(id, userID uint) (*model.Delivery, uint, error) {
	rider, err := s.riderRepo.FindByUserID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, 0, ErrRiderNotFound
		}
		return nil, 0, err
	}
	d, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, 0, ErrDeliveryNotFound
		}
		return nil, 0, err
	}
	if d.RiderID == nil || *d.RiderID != rider.ID {
		return nil, 0, errors.New("无权操作他人配送单")
	}
	return d, rider.ID, nil
}
