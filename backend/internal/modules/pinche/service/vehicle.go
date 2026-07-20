// Package service 同城拼车出行业务逻辑层 - 车辆
package service

import (
	"errors"

	"wuchang-tongcheng/internal/modules/pinche/dto"
	"wuchang-tongcheng/internal/modules/pinche/model"
	"wuchang-tongcheng/internal/modules/pinche/repository"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

var (
	ErrVehicleNotFound      = errors.New("车辆不存在")
	ErrVehicleNoPermission  = errors.New("无权操作此车辆")
	ErrVehicleStatusInvalid = errors.New("车辆状态不允许此操作")
)

// VehicleService 车辆业务接口
type VehicleService interface {
	// C 端
	Create(regionID uint, driverID, userID uint, req *dto.CreateVehicleRequest) (*dto.VehicleInfo, error)
	Update(id uint, operatorID uint, req *dto.UpdateVehicleRequest) error
	Delete(id uint, operatorID uint) error
	GetByID(id uint) (*dto.VehicleInfo, error)
	ListByDriver(driverID uint, page, pageSize int) (*utils.Pagination, []dto.VehicleInfo, error)
	ListByUser(userID uint, page, pageSize int) (*utils.Pagination, []dto.VehicleInfo, error)
	SetDefault(id uint, operatorID uint) error

	// M 端
	AdminList(req *dto.VehicleListRequest) (*utils.Pagination, []dto.VehicleInfo, error)
	Review(id uint, req *dto.VehicleReviewRequest) error
	UpdateStatus(id uint, status int, reason string) error
}

type vehicleService struct {
	repo repository.VehicleRepository
}

// NewVehicleService 创建车辆 service 实例
func NewVehicleService(repo repository.VehicleRepository) VehicleService {
	return &vehicleService{repo: repo}
}

// vehicleStatusText 状态文本
func vehicleStatusText(status int) string {
	switch status {
	case model.VehicleStatusPending:
		return "待审"
	case model.VehicleStatusApproved:
		return "通过"
	case model.VehicleStatusRejected:
		return "拒绝"
	}
	return ""
}

// toVehicleInfo model -> dto
func toVehicleInfo(v *model.PincheVehicle) *dto.VehicleInfo {
	info := &dto.VehicleInfo{
		ID:                  v.ID,
		RegionID:            v.RegionID,
		DriverID:            v.DriverID,
		UserID:              v.UserID,
		PlateNo:             v.PlateNo,
		Brand:               v.Brand,
		Model:               v.Model,
		Year:                v.Year,
		Color:               v.Color,
		SeatCount:           v.SeatCount,
		VehicleType:         v.VehicleType,
		FuelType:            v.FuelType,
		VehicleLicensePhoto: v.VehicleLicensePhoto,
		InsurancePhoto:      v.InsurancePhoto,
		Status:              v.Status,
		StatusText:          vehicleStatusText(v.Status),
		AuditStatus:         v.AuditStatus,
		AuditReason:         v.AuditReason,
		IsDefault:           v.IsDefault,
	}
	if v.VehiclePhotos != nil {
		info.VehiclePhotos = v.VehiclePhotos
	}
	return info
}

// Create 添加车辆
func (s *vehicleService) Create(regionID uint, driverID, userID uint, req *dto.CreateVehicleRequest) (*dto.VehicleInfo, error) {
	v := &model.PincheVehicle{
		DriverID:            driverID,
		UserID:              userID,
		PlateNo:             req.PlateNo,
		Brand:               req.Brand,
		Model:               req.Model,
		Year:                req.Year,
		Color:               req.Color,
		SeatCount:           req.SeatCount,
		VehicleType:         req.VehicleType,
		FuelType:            req.FuelType,
		VehicleLicensePhoto: req.VehicleLicensePhoto,
		InsurancePhoto:      req.InsurancePhoto,
		Status:              model.VehicleStatusPending,
	}
	v.RegionID = regionID
	if req.VehiclePhotos != nil {
		if jb, err := model.FromJSON(req.VehiclePhotos); err == nil {
			v.VehiclePhotos = jb
		}
	}
	if v.VehicleType == "" {
		v.VehicleType = "sedan"
	}
	if v.FuelType == "" {
		v.FuelType = "gasoline"
	}

	if err := s.repo.Create(v); err != nil {
		return nil, err
	}
	// 若标记为默认，则覆盖同车主其他车辆
	if req.IsDefault {
		_ = s.repo.SetDefault(driverID, v.ID)
	}
	return toVehicleInfo(v), nil
}

// Update 更新车辆
func (s *vehicleService) Update(id uint, operatorID uint, req *dto.UpdateVehicleRequest) error {
	v, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrVehicleNotFound
		}
		return err
	}
	if v.UserID != operatorID {
		return ErrVehicleNoPermission
	}
	fields := map[string]interface{}{}
	if req.PlateNo != nil {
		fields["plate_no"] = *req.PlateNo
	}
	if req.Brand != nil {
		fields["brand"] = *req.Brand
	}
	if req.Model != nil {
		fields["model"] = *req.Model
	}
	if req.Year != nil {
		fields["year"] = *req.Year
	}
	if req.Color != nil {
		fields["color"] = *req.Color
	}
	if req.SeatCount != nil {
		fields["seat_count"] = *req.SeatCount
	}
	if req.VehicleType != nil {
		fields["vehicle_type"] = *req.VehicleType
	}
	if req.FuelType != nil {
		fields["fuel_type"] = *req.FuelType
	}
	if req.VehicleLicensePhoto != nil {
		fields["vehicle_license_photo"] = *req.VehicleLicensePhoto
	}
	if req.InsurancePhoto != nil {
		fields["insurance_photo"] = *req.InsurancePhoto
	}
	if req.IsDefault != nil {
		fields["is_default"] = *req.IsDefault
	}
	if req.VehiclePhotos != nil {
		if jb, err := model.FromJSON(req.VehiclePhotos); err == nil {
			fields["vehicle_photos"] = jb
		}
	}
	if len(fields) == 0 {
		return nil
	}
	// 修改后重置为待审
	fields["status"] = model.VehicleStatusPending
	if err := s.repo.Update(id, fields); err != nil {
		return err
	}
	// 默认车辆切换
	if req.IsDefault != nil && *req.IsDefault {
		_ = s.repo.SetDefault(v.DriverID, id)
	}
	return nil
}

// Delete 删除车辆
func (s *vehicleService) Delete(id uint, operatorID uint) error {
	v, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrVehicleNotFound
		}
		return err
	}
	if v.UserID != operatorID {
		return ErrVehicleNoPermission
	}
	return s.repo.Delete(id)
}

// GetByID 获取详情
func (s *vehicleService) GetByID(id uint) (*dto.VehicleInfo, error) {
	v, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrVehicleNotFound
		}
		return nil, err
	}
	return toVehicleInfo(v), nil
}

// ListByDriver 按车主查询
func (s *vehicleService) ListByDriver(driverID uint, page, pageSize int) (*utils.Pagination, []dto.VehicleInfo, error) {
	pagination := utils.NewPagination(page, pageSize)
	list, total, err := s.repo.ListByDriver(driverID, pagination)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	result := make([]dto.VehicleInfo, 0, len(list))
	for i := range list {
		result = append(result, *toVehicleInfo(&list[i]))
	}
	return pagination, result, nil
}

// ListByUser 按用户查询
func (s *vehicleService) ListByUser(userID uint, page, pageSize int) (*utils.Pagination, []dto.VehicleInfo, error) {
	pagination := utils.NewPagination(page, pageSize)
	list, total, err := s.repo.ListByUser(userID, pagination)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	result := make([]dto.VehicleInfo, 0, len(list))
	for i := range list {
		result = append(result, *toVehicleInfo(&list[i]))
	}
	return pagination, result, nil
}

// SetDefault 设为默认车辆
func (s *vehicleService) SetDefault(id uint, operatorID uint) error {
	v, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrVehicleNotFound
		}
		return err
	}
	if v.UserID != operatorID {
		return ErrVehicleNoPermission
	}
	return s.repo.SetDefault(v.DriverID, id)
}

// AdminList 管理后台车辆列表
func (s *vehicleService) AdminList(req *dto.VehicleListRequest) (*utils.Pagination, []dto.VehicleInfo, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	opts := repository.VehicleListOptions{
		DriverID: req.DriverID,
		UserID:   req.UserID,
		Status:   req.Status,
		Keyword:  req.Keyword,
	}
	// AdminList 跨地区，regionID=0
	list, total, err := s.repo.List(0, pagination, opts)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	result := make([]dto.VehicleInfo, 0, len(list))
	for i := range list {
		result = append(result, *toVehicleInfo(&list[i]))
	}
	return pagination, result, nil
}

// Review 审核
func (s *vehicleService) Review(id uint, req *dto.VehicleReviewRequest) error {
	v, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrVehicleNotFound
		}
		return err
	}
	if v.Status != model.VehicleStatusPending {
		return ErrVehicleStatusInvalid
	}
	return s.repo.UpdateStatus(id, req.Status, req.Reason)
}

// UpdateStatus 管理后台更新状态
func (s *vehicleService) UpdateStatus(id uint, status int, reason string) error {
	return s.repo.UpdateStatus(id, status, reason)
}
