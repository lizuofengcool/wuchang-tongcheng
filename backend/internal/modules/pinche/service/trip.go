// Package service 同城拼车出行业务逻辑层 - 完成行程（含行程分享）
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
	ErrTripNotFound      = errors.New("行程不存在")
	ErrTripNoPermission  = errors.New("无权操作此行程")
	ErrTripStatusInvalid = errors.New("行程状态不允许此操作")
	ErrTripShareInvalid  = errors.New("行程分享已过期或不存在")
)

// TripService 完成行程业务接口
type TripService interface {
	// C 端
	Start(regionID uint, userID uint, req *dto.StartTripRequest) (*dto.TripInfo, error)
	Complete(id uint, operatorID uint, req *dto.CompleteTripRequest) (*dto.TripInfo, error)
	Confirm(id uint, operatorID uint, req *dto.ConfirmTripRequest) (*dto.TripInfo, error)
	GetByID(id uint) (*dto.TripInfo, error)
	GetByTripNo(tripNo string) (*dto.TripInfo, error)
	ListByUser(userID uint, page, pageSize int) (*utils.Pagination, []dto.TripInfo, error)
	ListByDriver(driverID uint, page, pageSize int) (*utils.Pagination, []dto.TripInfo, error)
	ListByPassenger(passengerID uint, page, pageSize int) (*utils.Pagination, []dto.TripInfo, error)
	ListByPinche(pincheID uint, page, pageSize int) (*utils.Pagination, []dto.TripInfo, error)

	// 行程分享
	Share(id uint, req *dto.ShareTripRequest) (*dto.TripShareResponse, error)
	GetByShareToken(token string) (*dto.TripInfo, error)

	// M 端
	AdminList(req *dto.TripListRequest) (*utils.Pagination, []dto.TripInfo, error)
	UpdateStatus(id uint, status int) error
}

type tripService struct {
	repo repository.TripRepository
}

// NewTripService 创建完成行程 service 实例
func NewTripService(repo repository.TripRepository) TripService {
	return &tripService{repo: repo}
}

// tripStatusText 状态文本
func tripStatusText(s int) string {
	switch s {
	case model.TripStatusOngoing:
		return "进行中"
	case model.TripStatusCompleted:
		return "已完成"
	case model.TripStatusAbnormal:
		return "异常结束"
	}
	return ""
}

// toTripInfo model -> dto
func toTripInfo(t *model.PincheTrip) *dto.TripInfo {
	return &dto.TripInfo{
		ID:                   t.ID,
		RegionID:             t.RegionID,
		PincheID:             t.PincheID,
		BookingID:            t.BookingID,
		TripNo:               t.TripNo,
		DriverID:             t.DriverID,
		DriverName:           t.DriverName,
		DriverPhone:          t.DriverPhone,
		PassengerID:          t.PassengerID,
		PassengerName:        t.PassengerName,
		PassengerPhone:       t.PassengerPhone,
		VehicleID:            t.VehicleID,
		PlateNo:              t.PlateNo,
		OriginAddress:        t.OriginAddress,
		OriginLat:            t.OriginLat,
		OriginLng:            t.OriginLng,
		DestinationAddress:   t.DestinationAddress,
		DestinationLat:       t.DestinationLat,
		DestinationLng:       t.DestinationLng,
		ActualPickupTime:     t.ActualPickupTime,
		ActualDropoffTime:    t.ActualDropoffTime,
		ActualDistanceKM:     t.ActualDistanceKM,
		ActualDurationMin:    t.ActualDurationMin,
		PassengersCount:      t.PassengersCount,
		FareAmount:           t.FareAmount,
		TollFee:              t.TollFee,
		TotalAmount:          t.TotalAmount,
		ShareToken:           t.ShareToken,
		ShareExpiresAt:       t.ShareExpiresAt,
		Status:               t.Status,
		StatusText:           tripStatusText(t.Status),
		DriverConfirmedAt:    t.DriverConfirmedAt,
		PassengerConfirmedAt: t.PassengerConfirmedAt,
		CompletedAt:          t.CompletedAt,
		CreatedAt:            t.CreatedAt,
	}
}

// genTripNo 生成行程单号 TR + yyyyMMddHHmmss + 8 位 hex
func genTripNo() string {
	b := make([]byte, 4)
	_, _ = rand.Read(b)
	return fmt.Sprintf("TR%s%s", time.Now().Format("20060102150405"), hex.EncodeToString(b))
}

// genShareToken 生成分享 token 32 位 hex
func genShareToken() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// Start 启动行程
func (s *tripService) Start(regionID uint, userID uint, req *dto.StartTripRequest) (*dto.TripInfo, error) {
	now := time.Now()
	tripNo := genTripNo()
	t := &model.PincheTrip{
		PincheID:         req.PincheID,
		BookingID:        req.BookingID,
		TripNo:           tripNo,
		DriverID:         userID,
		DriverName:       "",
		DriverPhone:      "",
		PassengerID:      0,
		PassengerName:    "",
		PassengerPhone:   "",
		ActualPickupTime: &now,
		Status:           model.TripStatusOngoing,
	}
	t.RegionID = regionID
	// 行程分享 token 默认 24 小时有效
	t.ShareToken = genShareToken()
	expiresAt := now.Add(24 * time.Hour)
	t.ShareExpiresAt = &expiresAt

	if err := s.repo.Create(t); err != nil {
		return nil, err
	}
	return toTripInfo(t), nil
}

// Complete 完成行程（仅车主可调用）
func (s *tripService) Complete(id uint, operatorID uint, req *dto.CompleteTripRequest) (*dto.TripInfo, error) {
	t, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTripNotFound
		}
		return nil, err
	}
	if t.DriverID != operatorID {
		return nil, ErrTripNoPermission
	}
	if t.Status != model.TripStatusOngoing {
		return nil, ErrTripStatusInvalid
	}
	now := time.Now()
	fields := map[string]interface{}{
		"actual_dropoff_time": &now,
		"actual_dropoff_lat":  req.ActualDropoffLat,
		"actual_dropoff_lng":  req.ActualDropoffLng,
		"actual_distance_km":  req.ActualDistanceKM,
		"actual_duration_min": req.ActualDurationMin,
		"toll_fee":            req.TollFee,
	}
	// 车主完成即视为已完成（也支持双方确认流程，简化处理）
	fields["status"] = model.TripStatusCompleted
	fields["completed_at"] = &now
	fields["driver_confirmed_at"] = &now

	if err := s.repo.Update(id, fields); err != nil {
		return nil, err
	}
	t.ActualDropoffTime = &now
	t.ActualDistanceKM = req.ActualDistanceKM
	t.ActualDurationMin = req.ActualDurationMin
	t.TollFee = req.TollFee
	t.Status = model.TripStatusCompleted
	t.CompletedAt = &now
	t.DriverConfirmedAt = &now
	return toTripInfo(t), nil
}

// Confirm 确认行程（车主/乘客）
func (s *tripService) Confirm(id uint, operatorID uint, req *dto.ConfirmTripRequest) (*dto.TripInfo, error) {
	t, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTripNotFound
		}
		return nil, err
	}
	now := time.Now()
	fields := map[string]interface{}{}
	switch req.ConfirmRole {
	case "driver":
		if t.DriverID != operatorID {
			return nil, ErrTripNoPermission
		}
		fields["driver_confirmed_at"] = &now
	case "passenger":
		if t.PassengerID != operatorID {
			return nil, ErrTripNoPermission
		}
		fields["passenger_confirmed_at"] = &now
	}
	if len(fields) == 0 {
		return nil, ErrTripStatusInvalid
	}
	// 双方都确认则更新为已完成
	if t.DriverConfirmedAt != nil && req.ConfirmRole == "passenger" {
		fields["status"] = model.TripStatusCompleted
		fields["completed_at"] = &now
	} else if t.PassengerConfirmedAt != nil && req.ConfirmRole == "driver" {
		fields["status"] = model.TripStatusCompleted
		fields["completed_at"] = &now
	}
	if err := s.repo.Update(id, fields); err != nil {
		return nil, err
	}
	// 重新查询以获得最新数据
	t2, err := s.repo.FindByID(id)
	if err == nil {
		return toTripInfo(t2), nil
	}
	return toTripInfo(t), nil
}

// GetByID 获取详情
func (s *tripService) GetByID(id uint) (*dto.TripInfo, error) {
	t, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTripNotFound
		}
		return nil, err
	}
	return toTripInfo(t), nil
}

// GetByTripNo 按单号查询
func (s *tripService) GetByTripNo(tripNo string) (*dto.TripInfo, error) {
	t, err := s.repo.FindByTripNo(tripNo)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTripNotFound
		}
		return nil, err
	}
	return toTripInfo(t), nil
}

// ListByUser 按用户查询（车主+乘客）
func (s *tripService) ListByUser(userID uint, page, pageSize int) (*utils.Pagination, []dto.TripInfo, error) {
	pagination := utils.NewPagination(page, pageSize)
	list, total, err := s.repo.ListByUser(userID, pagination)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	result := make([]dto.TripInfo, 0, len(list))
	for i := range list {
		result = append(result, *toTripInfo(&list[i]))
	}
	return pagination, result, nil
}

// ListByDriver 按车主查询
func (s *tripService) ListByDriver(driverID uint, page, pageSize int) (*utils.Pagination, []dto.TripInfo, error) {
	pagination := utils.NewPagination(page, pageSize)
	list, total, err := s.repo.ListByDriver(driverID, pagination)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	result := make([]dto.TripInfo, 0, len(list))
	for i := range list {
		result = append(result, *toTripInfo(&list[i]))
	}
	return pagination, result, nil
}

// ListByPassenger 按乘客查询
func (s *tripService) ListByPassenger(passengerID uint, page, pageSize int) (*utils.Pagination, []dto.TripInfo, error) {
	pagination := utils.NewPagination(page, pageSize)
	list, total, err := s.repo.ListByPassenger(passengerID, pagination)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	result := make([]dto.TripInfo, 0, len(list))
	for i := range list {
		result = append(result, *toTripInfo(&list[i]))
	}
	return pagination, result, nil
}

// ListByPinche 按拼车行程查询
func (s *tripService) ListByPinche(pincheID uint, page, pageSize int) (*utils.Pagination, []dto.TripInfo, error) {
	pagination := utils.NewPagination(page, pageSize)
	list, total, err := s.repo.ListByPinche(pincheID, pagination)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	result := make([]dto.TripInfo, 0, len(list))
	for i := range list {
		result = append(result, *toTripInfo(&list[i]))
	}
	return pagination, result, nil
}

// Share 生成行程分享
func (s *tripService) Share(id uint, req *dto.ShareTripRequest) (*dto.TripShareResponse, error) {
	t, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTripNotFound
		}
		return nil, err
	}
	hours := req.Hours
	if hours <= 0 {
		hours = 24
	}
	token := genShareToken()
	now := time.Now()
	expiresAt := now.Add(time.Duration(hours) * time.Hour)
	fields := map[string]interface{}{
		"share_token":      token,
		"share_expires_at": &expiresAt,
	}
	if err := s.repo.Update(id, fields); err != nil {
		return nil, err
	}
	resp := &dto.TripShareResponse{
		ShareToken:     token,
		ShareURL:       fmt.Sprintf("/pinche/share/%s", token),
		ShareExpiresAt: &expiresAt,
		TripNo:         t.TripNo,
		DriverName:     t.DriverName,
		PlateNo:        t.PlateNo,
		OriginAddress:  t.OriginAddress,
		DestinationAddress: t.DestinationAddress,
		Status:         t.Status,
	}
	return resp, nil
}

// GetByShareToken 通过分享 token 查询
func (s *tripService) GetByShareToken(token string) (*dto.TripInfo, error) {
	t, err := s.repo.FindByShareToken(token)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTripShareInvalid
		}
		return nil, err
	}
	// 检查过期
	if t.ShareExpiresAt != nil && t.ShareExpiresAt.Before(time.Now()) {
		return nil, ErrTripShareInvalid
	}
	return toTripInfo(t), nil
}

// AdminList 管理后台行程列表
func (s *tripService) AdminList(req *dto.TripListRequest) (*utils.Pagination, []dto.TripInfo, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	opts := repository.TripListOptions{
		UserID:      req.UserID,
		DriverID:    req.DriverID,
		PassengerID: req.PassengerID,
		PincheID:    req.PincheID,
		Status:      req.Status,
		TripNo:      req.TripNo,
		StartTime:   req.StartTime,
		EndTime:     req.EndTime,
	}
	// 跨地区：regionID=0
	list, total, err := s.repo.List(0, pagination, opts)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	result := make([]dto.TripInfo, 0, len(list))
	for i := range list {
		result = append(result, *toTripInfo(&list[i]))
	}
	return pagination, result, nil
}

// UpdateStatus 管理后台更新状态
func (s *tripService) UpdateStatus(id uint, status int) error {
	return s.repo.UpdateStatus(id, status)
}
