// Package service 同城拼车出行业务逻辑层 - 车主认证
package service

import (
	"errors"
	"strings"
	"time"

	"wuchang-tongcheng/internal/modules/pinche/dto"
	"wuchang-tongcheng/internal/modules/pinche/model"
	"wuchang-tongcheng/internal/modules/pinche/repository"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

var (
	ErrDriverNotFound      = errors.New("车主认证不存在")
	ErrDriverNoPermission  = errors.New("无权操作此车主认证")
	ErrDriverStatusInvalid = errors.New("车主认证状态不允许此操作")
	ErrDriverAlreadyExists = errors.New("已提交过车主认证")
	ErrDriverNotApproved   = errors.New("车主认证未通过")
)

// DriverService 车主认证业务接口
type DriverService interface {
	// C 端
	Create(regionID uint, userID uint, userName, userPhone, userAvatar string, req *dto.CreateDriverRequest) (*dto.DriverInfo, error)
	Update(id uint, operatorID uint, req *dto.UpdateDriverRequest) error
	GetByID(id uint) (*dto.DriverInfo, error)
	GetByUserID(userID uint) (*dto.DriverInfo, error)
	List(regionID uint, req *dto.DriverListRequest) (*utils.Pagination, []dto.DriverInfo, error)

	// M 端
	AdminList(req *dto.DriverListRequest) (*utils.Pagination, []dto.DriverInfo, error)
	Review(id uint, auditorID uint, req *dto.DriverReviewRequest) error
	UpdateStatus(id uint, status int) error
}

type driverService struct {
	repo repository.DriverRepository
}

// NewDriverService 创建车主认证 service 实例
func NewDriverService(repo repository.DriverRepository) DriverService {
	return &driverService{repo: repo}
}

// driverStatusText 状态文本
func driverStatusText(status int) string {
	switch status {
	case model.DriverStatusPending:
		return "待审"
	case model.DriverStatusApproved:
		return "通过"
	case model.DriverStatusRejected:
		return "拒绝"
	case model.DriverStatusExpired:
		return "已过期"
	}
	return ""
}

// maskIDCard 身份证脱敏
func maskIDCard(idCard string) string {
	if len(idCard) < 11 {
		return strings.Repeat("*", len(idCard))
	}
	return idCard[:3] + strings.Repeat("*", 11) + idCard[len(idCard)-4:]
}

// toDriverInfo model -> dto
func toDriverInfo(d *model.PincheDriver) *dto.DriverInfo {
	return &dto.DriverInfo{
		ID:                  d.ID,
		RegionID:            d.RegionID,
		UserID:              d.UserID,
		UserName:            d.UserName,
		UserPhone:           d.UserPhone,
		UserAvatar:          d.UserAvatar,
		RealName:            d.RealName,
		IDCardNo:            maskIDCard(d.IDCardNo),
		IDCardFront:         d.IDCardFront,
		IDCardBack:          d.IDCardBack,
		DriverLicenseNo:     d.DriverLicenseNo,
		DriverLicenseFront:  d.DriverLicenseFront,
		DriverLicenseBack:   d.DriverLicenseBack,
		LicenseIssueDate:    d.LicenseIssueDate,
		LicenseExpiryDate:   d.LicenseExpiryDate,
		VehicleLicenseNo:    d.VehicleLicenseNo,
		VehicleLicenseFront: d.VehicleLicenseFront,
		VehicleLicenseBack:  d.VehicleLicenseBack,
		CarPhoto:            d.CarPhoto,
		Status:              d.Status,
		StatusText:          driverStatusText(d.Status),
		AuditReason:         d.AuditReason,
		AuditedAt:           d.AuditedAt,
		AuditorID:           d.AuditorID,
		Verified:            d.Verified,
		VerifiedAt:          d.VerifiedAt,
		RatingAvg:           d.RatingAvg,
		TripCount:           d.TripCount,
		TotalIncome:         d.TotalIncome,
		CreatedAt:           d.CreatedAt,
	}
}

// Create 提交车主认证
func (s *driverService) Create(regionID uint, userID uint, userName, userPhone, userAvatar string, req *dto.CreateDriverRequest) (*dto.DriverInfo, error) {
	// 检查是否已认证
	if existing, err := s.repo.FindByUserID(userID); err == nil && existing != nil {
		return nil, ErrDriverAlreadyExists
	}

	d := &model.PincheDriver{
		UserID:              userID,
		UserName:            userName,
		UserPhone:           userPhone,
		UserAvatar:          userAvatar,
		RealName:            req.RealName,
		IDCardNo:            req.IDCardNo,
		IDCardFront:         req.IDCardFront,
		IDCardBack:          req.IDCardBack,
		DriverLicenseNo:     req.DriverLicenseNo,
		DriverLicenseFront:  req.DriverLicenseFront,
		DriverLicenseBack:   req.DriverLicenseBack,
		LicenseIssueDate:    req.LicenseIssueDate,
		LicenseExpiryDate:   req.LicenseExpiryDate,
		VehicleLicenseNo:    req.VehicleLicenseNo,
		VehicleLicenseFront: req.VehicleLicenseFront,
		VehicleLicenseBack:  req.VehicleLicenseBack,
		CarPhoto:            req.CarPhoto,
		Status:              model.DriverStatusPending,
	}
	d.RegionID = regionID
	if err := s.repo.Create(d); err != nil {
		return nil, err
	}
	return toDriverInfo(d), nil
}

// Update 更新车主认证（仅本人）
func (s *driverService) Update(id uint, operatorID uint, req *dto.UpdateDriverRequest) error {
	d, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrDriverNotFound
		}
		return err
	}
	if d.UserID != operatorID {
		return ErrDriverNoPermission
	}
	fields := map[string]interface{}{}
	if req.IDCardFront != nil {
		fields["id_card_front"] = *req.IDCardFront
	}
	if req.IDCardBack != nil {
		fields["id_card_back"] = *req.IDCardBack
	}
	if req.DriverLicenseFront != nil {
		fields["driver_license_front"] = *req.DriverLicenseFront
	}
	if req.DriverLicenseBack != nil {
		fields["driver_license_back"] = *req.DriverLicenseBack
	}
	if req.LicenseExpiryDate != nil {
		fields["license_expiry_date"] = *req.LicenseExpiryDate
	}
	if req.VehicleLicenseFront != nil {
		fields["vehicle_license_front"] = *req.VehicleLicenseFront
	}
	if req.VehicleLicenseBack != nil {
		fields["vehicle_license_back"] = *req.VehicleLicenseBack
	}
	if req.CarPhoto != nil {
		fields["car_photo"] = *req.CarPhoto
	}
	if len(fields) == 0 {
		return nil
	}
	// 修改后重新审核
	fields["status"] = model.DriverStatusPending
	return s.repo.Update(id, fields)
}

// GetByID 获取详情
func (s *driverService) GetByID(id uint) (*dto.DriverInfo, error) {
	d, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrDriverNotFound
		}
		return nil, err
	}
	return toDriverInfo(d), nil
}

// GetByUserID 按用户 ID 查询
func (s *driverService) GetByUserID(userID uint) (*dto.DriverInfo, error) {
	d, err := s.repo.FindByUserID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrDriverNotFound
		}
		return nil, err
	}
	return toDriverInfo(d), nil
}

// List 车主列表（C 端：仅展示已通过）
func (s *driverService) List(regionID uint, req *dto.DriverListRequest) (*utils.Pagination, []dto.DriverInfo, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	opts := repository.DriverListOptions{
		UserID:   req.UserID,
		Keyword:  req.Keyword,
	}
	approved := model.DriverStatusApproved
	opts.Status = &approved
	if req.Verified != nil {
		opts.Verified = req.Verified
	}
	list, total, err := s.repo.List(regionID, pagination, opts)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	result := make([]dto.DriverInfo, 0, len(list))
	for i := range list {
		result = append(result, *toDriverInfo(&list[i]))
	}
	return pagination, result, nil
}

// AdminList 管理后台车主列表
func (s *driverService) AdminList(req *dto.DriverListRequest) (*utils.Pagination, []dto.DriverInfo, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	opts := repository.DriverListOptions{
		UserID:   req.UserID,
		Status:   req.Status,
		Keyword:  req.Keyword,
	}
	list, total, err := s.repo.AdminList(pagination, opts)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	result := make([]dto.DriverInfo, 0, len(list))
	for i := range list {
		result = append(result, *toDriverInfo(&list[i]))
	}
	return pagination, result, nil
}

// Review 审核
func (s *driverService) Review(id uint, auditorID uint, req *dto.DriverReviewRequest) error {
	d, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrDriverNotFound
		}
		return err
	}
	if d.Status != model.DriverStatusPending {
		return ErrDriverStatusInvalid
	}
	now := time.Now()
	fields := map[string]interface{}{
		"status":       req.Status,
		"audit_reason": req.Reason,
		"audited_at":   &now,
		"auditor_id":   auditorID,
	}
	if req.Status == model.DriverStatusApproved {
		fields["verified"] = true
		fields["verified_at"] = &now
	}
	return s.repo.Update(id, fields)
}

// UpdateStatus 管理后台更新状态
func (s *driverService) UpdateStatus(id uint, status int) error {
	return s.repo.UpdateStatus(id, status)
}
