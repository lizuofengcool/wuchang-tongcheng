// Package service 同城拼车出行业务逻辑层 - 预订
package service

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"wuchang-tongcheng/internal/modules/pinche/dto"
	"wuchang-tongcheng/internal/modules/pinche/model"
	"wuchang-tongcheng/internal/modules/pinche/repository"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

var (
	ErrBookingNotFound      = errors.New("预订不存在")
	ErrBookingNoPermission  = errors.New("无权操作此预订")
	ErrBookingStatusInvalid = errors.New("预订状态不允许此操作")
	ErrBookingAlreadyExists = errors.New("已预订过该行程")
	ErrBookingCodeInvalid   = errors.New("上车码无效")
)

// BookingService 预订业务接口
type BookingService interface {
	// C 端
	Create(regionID uint, userID uint, userName, userPhone, userAvatar string, req *dto.CreateBookingRequest) (*dto.BookingInfo, error)
	Update(id uint, operatorID uint, req *dto.UpdateBookingRequest) error
	Cancel(id uint, operatorID uint, req *dto.CancelBookingRequest) error
	ConfirmBoarding(id uint, operatorID uint, req *dto.ConfirmBoardingRequest) (*dto.BoardingResponse, error)
	GetByID(id uint) (*dto.BookingInfo, error)
	List(regionID uint, req *dto.BookingListRequest) (*utils.Pagination, []dto.BookingInfo, error)
	ListByPassenger(userID uint, page, pageSize int) (*utils.Pagination, []dto.BookingInfo, error)
	ListByDriver(userID uint, page, pageSize int) (*utils.Pagination, []dto.BookingInfo, error)
	ListByPinche(pincheID uint, page, pageSize int) (*utils.Pagination, []dto.BookingInfo, error)

	// M 端
	AdminUpdateStatus(id uint, status int) error
}

type bookingService struct {
	repo     repository.BookingRepository
	pincheRepo repository.PincheRepository
}

// NewBookingService 创建预订 service 实例
func NewBookingService(repo repository.BookingRepository, pincheRepo repository.PincheRepository) BookingService {
	return &bookingService{repo: repo, pincheRepo: pincheRepo}
}

// bookingStatusText 状态文本
func bookingStatusText(status int) string {
	switch status {
	case model.BookingStatusPending:
		return "待支付"
	case model.BookingStatusPaid:
		return "已支付"
	case model.BookingStatusBoarded:
		return "已上车"
	case model.BookingStatusCompleted:
		return "已完成"
	case model.BookingStatusCancelled:
		return "已取消"
	case model.BookingStatusRefunded:
		return "已退款"
	}
	return ""
}

// toBookingInfo model -> dto
func toBookingInfo(b *model.PincheBooking) *dto.BookingInfo {
	return &dto.BookingInfo{
		ID:              b.ID,
		RegionID:        b.RegionID,
		PincheID:        b.PincheID,
		BookingNo:       b.BookingNo,
		PassengerID:     b.PassengerID,
		PassengerName:   b.PassengerName,
		PassengerPhone:  b.PassengerPhone,
		PassengerAvatar: b.PassengerAvatar,
		DriverID:        b.DriverID,
		DriverName:      b.DriverName,
		DriverPhone:     b.DriverPhone,
		Seats:           b.Seats,
		PickupLocation:  b.PickupLocation,
		PickupLat:       b.PickupLat,
		PickupLng:       b.PickupLng,
		DropoffLocation: b.DropoffLocation,
		DropoffLat:      b.DropoffLat,
		DropoffLng:      b.DropoffLng,
		UnitPrice:       b.UnitPrice,
		TotalAmount:     b.TotalAmount,
		InsuranceFee:    b.InsuranceFee,
		ServiceFee:      b.ServiceFee,
		Status:          b.Status,
		StatusText:      bookingStatusText(b.Status),
		PaymentID:       b.PaymentID,
		BoardingCode:    b.BoardingCode,
		PaidAt:          b.PaidAt,
		BoardedAt:       b.BoardedAt,
		CompletedAt:     b.CompletedAt,
		CancelledAt:     b.CancelledAt,
		CancelReason:    b.CancelReason,
		CancelledBy:     b.CancelledBy,
		CreatedAt:       b.CreatedAt,
	}
}

// genBookingNo 生成预订单号 BK + yyyyMMddHHmmss + 6位hex
func genBookingNo() string {
	return fmt.Sprintf("BK%s%s", time.Now().Format("20060102150405"), randomHex(3))
}

// genBoardingCode 生成 6 位上车码
func genBoardingCode() string {
	return randomHex(3)
}

// randomHex 生成 n 字节随机十六进制字符串
func randomHex(n int) string {
	b := make([]byte, n)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// Create 创建预订
func (s *bookingService) Create(regionID uint, userID uint, userName, userPhone, userAvatar string, req *dto.CreateBookingRequest) (*dto.BookingInfo, error) {
	// 校验拼车行程
	p, err := s.pincheRepo.FindByID(req.PincheID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrPincheNotFound
		}
		return nil, err
	}
	if p.Status != model.PincheStatusPublished && p.Status != model.PincheStatusOngoing {
		return nil, ErrPincheStatusInvalid
	}
	if p.AvailableSeats < req.Seats {
		return nil, ErrPincheNoSeats
	}
	// 检查是否已预订
	hasBooked, err := s.repo.HasBooked(userID, req.PincheID)
	if err != nil {
		return nil, err
	}
	if hasBooked {
		return nil, ErrBookingAlreadyExists
	}

	totalAmount := p.PricePerSeat * float64(req.Seats)
	insuranceFee := 0.0
	if req.BuyInsurance {
		insuranceFee = 5.0 * float64(req.Seats) // 简化：5元/座
	}

	b := &model.PincheBooking{
		PincheID:        req.PincheID,
		BookingNo:       genBookingNo(),
		PassengerID:     userID,
		PassengerName:   userName,
		PassengerPhone:  userPhone,
		PassengerAvatar: userAvatar,
		DriverID:        p.UserID,
		DriverName:      p.UserName,
		DriverPhone:     p.UserPhone,
		Seats:           req.Seats,
		PickupLocation:  req.PickupLocation,
		PickupLat:       req.PickupLat,
		PickupLng:       req.PickupLng,
		DropoffLocation: req.DropoffLocation,
		DropoffLat:      req.DropoffLat,
		DropoffLng:      req.DropoffLng,
		UnitPrice:       p.PricePerSeat,
		TotalAmount:     totalAmount,
		InsuranceFee:    insuranceFee,
		ServiceFee:      0,
		Status:          model.BookingStatusPending,
		BoardingCode:    genBoardingCode(),
	}
	b.RegionID = regionID

	if err := s.repo.Create(b); err != nil {
		return nil, err
	}

	// 更新主表已订座位
	bookedSeats := p.BookedSeats + req.Seats
	availableSeats := p.TotalSeats - bookedSeats
	_ = s.pincheRepo.Update(req.PincheID, map[string]interface{}{
		"booked_seats":    bookedSeats,
		"available_seats": availableSeats,
	})

	return toBookingInfo(b), nil
}

// Update 更新预订（仅乘客）
func (s *bookingService) Update(id uint, operatorID uint, req *dto.UpdateBookingRequest) error {
	b, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrBookingNotFound
		}
		return err
	}
	if b.PassengerID != operatorID {
		return ErrBookingNoPermission
	}
	if b.Status != model.BookingStatusPending && b.Status != model.BookingStatusPaid {
		return ErrBookingStatusInvalid
	}
	fields := map[string]interface{}{}
	if req.Seats != nil {
		fields["seats"] = *req.Seats
	}
	if req.PickupLocation != nil {
		fields["pickup_location"] = *req.PickupLocation
	}
	if req.PickupLat != nil {
		fields["pickup_lat"] = *req.PickupLat
	}
	if req.PickupLng != nil {
		fields["pickup_lng"] = *req.PickupLng
	}
	if req.DropoffLocation != nil {
		fields["dropoff_location"] = *req.DropoffLocation
	}
	if req.DropoffLat != nil {
		fields["dropoff_lat"] = *req.DropoffLat
	}
	if req.DropoffLng != nil {
		fields["dropoff_lng"] = *req.DropoffLng
	}
	if len(fields) == 0 {
		return nil
	}
	return s.repo.Update(id, fields)
}

// Cancel 取消预订
func (s *bookingService) Cancel(id uint, operatorID uint, req *dto.CancelBookingRequest) error {
	b, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrBookingNotFound
		}
		return err
	}
	if b.PassengerID != operatorID && b.DriverID != operatorID {
		return ErrBookingNoPermission
	}
	if b.Status == model.BookingStatusCancelled || b.Status == model.BookingStatusCompleted {
		return ErrBookingStatusInvalid
	}

	now := time.Now()
	fields := map[string]interface{}{
		"status":        model.BookingStatusCancelled,
		"cancelled_at":  &now,
		"cancel_reason": req.Reason,
		"cancelled_by":  operatorID,
	}
	if err := s.repo.Update(id, fields); err != nil {
		return err
	}

	// 恢复主表座位
	p, err := s.pincheRepo.FindByID(b.PincheID)
	if err == nil {
		bookedSeats := p.BookedSeats - b.Seats
		if bookedSeats < 0 {
			bookedSeats = 0
		}
		availableSeats := p.TotalSeats - bookedSeats
		_ = s.pincheRepo.Update(b.PincheID, map[string]interface{}{
			"booked_seats":    bookedSeats,
			"available_seats": availableSeats,
		})
	}
	return nil
}

// ConfirmBoarding 确认上车（凭上车码）
func (s *bookingService) ConfirmBoarding(id uint, operatorID uint, req *dto.ConfirmBoardingRequest) (*dto.BoardingResponse, error) {
	b, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrBookingNotFound
		}
		return nil, err
	}
	if b.DriverID != operatorID && b.PassengerID != operatorID {
		return nil, ErrBookingNoPermission
	}
	if b.BoardingCode != req.BoardingCode {
		return nil, ErrBookingCodeInvalid
	}
	if b.Status != model.BookingStatusPaid {
		return nil, ErrBookingStatusInvalid
	}

	now := time.Now()
	if err := s.repo.Update(id, map[string]interface{}{
		"status":     model.BookingStatusBoarded,
		"boarded_at": &now,
	}); err != nil {
		return nil, err
	}
	return &dto.BoardingResponse{
		BoardingCode: b.BoardingCode,
		BookingID:    b.ID,
	}, nil
}

// GetByID 获取预订详情
func (s *bookingService) GetByID(id uint) (*dto.BookingInfo, error) {
	b, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrBookingNotFound
		}
		return nil, err
	}
	return toBookingInfo(b), nil
}

// List 预订列表
func (s *bookingService) List(regionID uint, req *dto.BookingListRequest) (*utils.Pagination, []dto.BookingInfo, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	opts := repository.BookingListOptions{
		PincheID:    req.PincheID,
		PassengerID: req.PassengerID,
		DriverID:    req.DriverID,
		Status:      req.Status,
		Keyword:     req.Keyword,
	}
	list, total, err := s.repo.List(regionID, pagination, opts)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	result := make([]dto.BookingInfo, 0, len(list))
	for i := range list {
		result = append(result, *toBookingInfo(&list[i]))
	}
	return pagination, result, nil
}

// ListByPassenger 乘客的预订列表
func (s *bookingService) ListByPassenger(userID uint, page, pageSize int) (*utils.Pagination, []dto.BookingInfo, error) {
	pagination := utils.NewPagination(page, pageSize)
	list, total, err := s.repo.ListByPassenger(userID, pagination)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	result := make([]dto.BookingInfo, 0, len(list))
	for i := range list {
		result = append(result, *toBookingInfo(&list[i]))
	}
	return pagination, result, nil
}

// ListByDriver 车主的预订列表
func (s *bookingService) ListByDriver(userID uint, page, pageSize int) (*utils.Pagination, []dto.BookingInfo, error) {
	pagination := utils.NewPagination(page, pageSize)
	list, total, err := s.repo.ListByDriver(userID, pagination)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	result := make([]dto.BookingInfo, 0, len(list))
	for i := range list {
		result = append(result, *toBookingInfo(&list[i]))
	}
	return pagination, result, nil
}

// ListByPinche 行程的预订列表
func (s *bookingService) ListByPinche(pincheID uint, page, pageSize int) (*utils.Pagination, []dto.BookingInfo, error) {
	pagination := utils.NewPagination(page, pageSize)
	list, total, err := s.repo.ListByPinche(pincheID, pagination)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	result := make([]dto.BookingInfo, 0, len(list))
	for i := range list {
		result = append(result, *toBookingInfo(&list[i]))
	}
	return pagination, result, nil
}

// AdminUpdateStatus 管理后台强制变更状态
func (s *bookingService) AdminUpdateStatus(id uint, status int) error {
	return s.repo.UpdateStatus(id, status)
}
