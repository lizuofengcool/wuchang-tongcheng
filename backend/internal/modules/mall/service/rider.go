// Package service 同城商城业务逻辑层 - 骑手
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
	// ErrRiderNotFound 骑手不存在
	ErrRiderNotFound = errors.New("骑手不存在")
	// ErrRiderNotVerified 骑手未认证
	ErrRiderNotVerified = errors.New("骑手未认证")
	// ErrRiderFrozen 骑手已冻结
	ErrRiderFrozen = errors.New("骑手已冻结")
	// ErrRiderOnline 骑手已上线
	ErrRiderOnline = errors.New("骑手已上线")
	// ErrRiderOffline 骑手已下线
	ErrRiderOffline = errors.New("骑手已下线")
	// ErrRiderBusy 骑手配送中
	ErrRiderBusy = errors.New("骑手配送中")
	// ErrRiderInsufficientCredit 信用分不足
	ErrRiderInsufficientCredit = errors.New("信用分不足")
)

// RiderService 骑手业务接口
type RiderService interface {
	Apply(regionID, userID uint, req *dto.RiderApplyRequest) (*dto.RiderInfo, error)
	Update(id, userID uint, req *dto.RiderUpdateRequest) error
	GetByID(id uint) (*dto.RiderInfo, error)
	GetByUserID(userID uint) (*dto.RiderInfo, error)
	List(regionID uint, req *dto.RiderListRequest) (*utils.Pagination, []dto.RiderInfo, error)
	AdminList(req *dto.RiderListRequest) (*utils.Pagination, []dto.RiderInfo, error)

	Online(id, userID uint) error
	Offline(id, userID uint) error
	Audit(id uint, req *dto.RiderAuditRequest) error
	UpdateStatus(id uint, req *dto.RiderStatusUpdateRequest) error

	// 统计
	Earnings(userID uint) (*dto.RiderEarningsResponse, error)
}

type riderService struct {
	repo          repository.RiderRepository
	deliveryRepo  repository.DeliveryRepository
}

// NewRiderService 创建骑手 service 实例
func NewRiderService(repo repository.RiderRepository, deliveryRepo repository.DeliveryRepository) RiderService {
	return &riderService{repo: repo, deliveryRepo: deliveryRepo}
}

func riderStatusText(s int) string {
	switch s {
	case model.RiderStatusPending:
		return "待审核"
	case model.RiderStatusApproved:
		return "已通过"
	case model.RiderStatusRejected:
		return "已拒绝"
	case model.RiderStatusFrozen:
		return "已冻结"
	}
	return ""
}

func riderOnlineStatusText(s int) string {
	switch s {
	case model.RiderOffline:
		return "下线"
	case model.RiderOnline:
		return "在线"
	case model.RiderDelivering:
		return "配送中"
	}
	return ""
}

func riderVehicleTypeText(t string) string {
	switch t {
	case model.VehicleTypeElectric:
		return "电动车"
	case model.VehicleTypeMotor:
		return "摩托车"
	case model.VehicleTypeBicycle:
		return "自行车"
	case model.VehicleTypeCar:
		return "汽车"
	}
	return ""
}

func toRiderInfo(r *model.Rider) *dto.RiderInfo {
	return &dto.RiderInfo{
		ID:              r.ID,
		UserID:          r.UserID,
		ShopID:          r.ShopID,
		RealName:        r.RealName,
		Phone:           r.Phone,
		IDCard:          r.IDCard,
		Avatar:          r.Avatar,
		VehicleType:     r.VehicleType,
		VehicleTypeText: riderVehicleTypeText(r.VehicleType),
		VehiclePlate:    r.VehiclePlate,
		LicenseURL:      r.LicenseURL,
		Status:          r.Status,
		StatusText:      riderStatusText(r.Status),
		CreditScore:     r.CreditScore,
		Level:           r.Level,
		TotalOrders:     r.TotalOrders,
		TotalEarnings:   r.TotalEarnings,
		OnlineStatus:    r.OnlineStatus,
		OnlineStatusText: riderOnlineStatusText(r.OnlineStatus),
		AuditReason:     r.AuditReason,
		RegionID:        r.RegionID,
		CreatedAt:       r.CreatedAt,
		UpdatedAt:       r.UpdatedAt,
	}
}

// Apply 骑手申请
func (s *riderService) Apply(regionID, userID uint, req *dto.RiderApplyRequest) (*dto.RiderInfo, error) {
	// 检查是否已申请过
	existing, err := s.repo.FindByUserID(userID)
	if err == nil && existing != nil {
		// 已存在，返回当前状态
		return toRiderInfo(existing), nil
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	rider := &model.Rider{
		UserID:       userID,
		ShopID:       req.ShopID,
		RealName:     req.RealName,
		Phone:        req.Phone,
		IDCard:       req.IDCard,
		Avatar:       req.Avatar,
		VehicleType:  req.VehicleType,
		VehiclePlate: req.VehiclePlate,
		LicenseURL:   req.LicenseURL,
		Status:       model.RiderStatusPending,
		CreditScore:  100,
		Level:        1,
		OnlineStatus: model.RiderOffline,
	}
	rider.RegionID = regionID

	if err := s.repo.Create(rider); err != nil {
		return nil, err
	}
	return toRiderInfo(rider), nil
}

// Update 更新骑手资料
func (s *riderService) Update(id, userID uint, req *dto.RiderUpdateRequest) error {
	rider, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrRiderNotFound
		}
		return err
	}
	if rider.UserID != userID {
		return errors.New("无权操作他人骑手资料")
	}

	fields := make(map[string]interface{})
	if req.RealName != nil {
		fields["real_name"] = *req.RealName
	}
	if req.Phone != nil {
		fields["phone"] = *req.Phone
	}
	if req.IDCard != nil {
		fields["id_card"] = *req.IDCard
	}
	if req.Avatar != nil {
		fields["avatar"] = *req.Avatar
	}
	if req.VehicleType != nil {
		fields["vehicle_type"] = *req.VehicleType
	}
	if req.VehiclePlate != nil {
		fields["vehicle_plate"] = *req.VehiclePlate
	}
	if req.LicenseURL != nil {
		fields["license_url"] = *req.LicenseURL
	}
	if len(fields) == 0 {
		return nil
	}
	return s.repo.UpdateFields(id, fields)
}

// GetByID 获取骑手详情
func (s *riderService) GetByID(id uint) (*dto.RiderInfo, error) {
	rider, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrRiderNotFound
		}
		return nil, err
	}
	return toRiderInfo(rider), nil
}

// GetByUserID 根据用户 ID 获取骑手资料
func (s *riderService) GetByUserID(userID uint) (*dto.RiderInfo, error) {
	rider, err := s.repo.FindByUserID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrRiderNotFound
		}
		return nil, err
	}
	return toRiderInfo(rider), nil
}

// List 骑手列表
func (s *riderService) List(regionID uint, req *dto.RiderListRequest) (*utils.Pagination, []dto.RiderInfo, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	opts := repository.RiderListOptions{
		Keyword:      req.Keyword,
		Status:       req.Status,
		OnlineStatus: req.OnlineStatus,
		ShopID:       req.ShopID,
		UserID:       req.UserID,
	}
	list, total, err := s.repo.List(regionID, pagination, opts)
	if err != nil {
		return nil, nil, err
	}
	infos := make([]dto.RiderInfo, 0, len(list))
	for i := range list {
		infos = append(infos, *toRiderInfo(&list[i]))
	}
	pagination.Total = total
	return pagination, infos, nil
}

// AdminList 管理后台骑手列表
func (s *riderService) AdminList(req *dto.RiderListRequest) (*utils.Pagination, []dto.RiderInfo, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	opts := repository.RiderAdminListOptions{
		RegionID:     req.RegionID,
		UserID:       req.UserID,
		ShopID:       req.ShopID,
		Status:       req.Status,
		OnlineStatus: req.OnlineStatus,
		Keyword:      req.Keyword,
	}
	list, total, err := s.repo.AdminList(pagination, opts)
	if err != nil {
		return nil, nil, err
	}
	infos := make([]dto.RiderInfo, 0, len(list))
	for i := range list {
		infos = append(infos, *toRiderInfo(&list[i]))
	}
	pagination.Total = total
	return pagination, infos, nil
}

// Online 骑手上线
func (s *riderService) Online(id, userID uint) error {
	rider, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrRiderNotFound
		}
		return err
	}
	if rider.UserID != userID {
		return errors.New("无权操作他人骑手资料")
	}
	if rider.Status == model.RiderStatusFrozen {
		return ErrRiderFrozen
	}
	if rider.Status != model.RiderStatusApproved {
		return ErrRiderNotVerified
	}
	if rider.OnlineStatus == model.RiderOnline {
		return ErrRiderOnline
	}
	if rider.OnlineStatus == model.RiderDelivering {
		return ErrRiderBusy
	}
	return s.repo.UpdateOnlineStatus(id, model.RiderOnline)
}

// Offline 骑手下线
func (s *riderService) Offline(id, userID uint) error {
	rider, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrRiderNotFound
		}
		return err
	}
	if rider.UserID != userID {
		return errors.New("无权操作他人骑手资料")
	}
	if rider.OnlineStatus == model.RiderOffline {
		return ErrRiderOffline
	}
	if rider.OnlineStatus == model.RiderDelivering {
		return ErrRiderBusy
	}
	return s.repo.UpdateOnlineStatus(id, model.RiderOffline)
}

// Audit 骑手审核
func (s *riderService) Audit(id uint, req *dto.RiderAuditRequest) error {
	rider, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrRiderNotFound
		}
		return err
	}
	_ = rider
	return s.repo.UpdateStatus(id, req.Status, req.AuditReason)
}

// UpdateStatus 管理后台骑手状态更新（冻结/解冻）
func (s *riderService) UpdateStatus(id uint, req *dto.RiderStatusUpdateRequest) error {
	_, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrRiderNotFound
		}
		return err
	}
	return s.repo.UpdateStatus(id, req.Status, "")
}

// Earnings 骑手收益统计
func (s *riderService) Earnings(userID uint) (*dto.RiderEarningsResponse, error) {
	rider, err := s.repo.FindByUserID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrRiderNotFound
		}
		return nil, err
	}

	resp := &dto.RiderEarningsResponse{
		TotalOrders:   int64(rider.TotalOrders),
		TotalEarnings: rider.TotalEarnings,
	}

	// 本月统计
	now := time.Now()
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	monthEnd := monthStart.AddDate(0, 1, 0)
	monthEarnings, monthOrders, err := s.deliveryRepo.SumEarnings(rider.ID, monthStart, monthEnd)
	if err == nil {
		resp.MonthEarnings = monthEarnings
		resp.MonthOrders = monthOrders
	}

	// 今日统计
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	todayEnd := todayStart.AddDate(0, 0, 1)
	todayEarnings, todayOrders, err := s.deliveryRepo.SumEarnings(rider.ID, todayStart, todayEnd)
	if err == nil {
		resp.TodayEarnings = todayEarnings
		resp.TodayOrders = todayOrders
	}

	return resp, nil
}
