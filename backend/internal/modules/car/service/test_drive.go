// Package service 同城车辆买卖业务逻辑层 - 试驾预约
// 依据需求文档 1.10：4 维数据隔离（region_id + user_id）
// 依据 v3.2.1 架构方案：对标懂车帝/汽车之家 试驾/看车/上门
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
	ErrTestDriveNotFound      = errors.New("试驾预约不存在")
	ErrTestDriveNoPermission  = errors.New("无权操作此试驾预约")
	ErrTestDriveStatusInvalid = errors.New("试驾预约状态不允许此操作")
)

// TestDriveService 试驾预约业务接口
type TestDriveService interface {
	// C 端
	Create(regionID uint, userID uint, userName string, userPhone string, userAvatar string, req *dto.CreateTestDriveRequest) (*dto.TestDriveInfo, error)
	Update(id uint, operatorID uint, req *dto.UpdateTestDriveRequest) error
	Cancel(id uint, userID uint, reason string) error
	GetByID(id uint) (*dto.TestDriveInfo, error)
	List(regionID uint, req *dto.TestDriveListRequest) (*utils.Pagination, []dto.TestDriveInfo, error)
	ListByUser(userID uint, page, pageSize int) (*utils.Pagination, []dto.TestDriveInfo, error)
	ListBySales(salesID uint, page, pageSize int) (*utils.Pagination, []dto.TestDriveInfo, error)
	ListByDealer(dealerID uint, page, pageSize int) (*utils.Pagination, []dto.TestDriveInfo, error)
	UploadLicense(id uint, userID uint, req *dto.TestDriveLicenseUploadRequest) error

	// M 端管理
	AdminList(req *dto.TestDriveListRequest) (*utils.Pagination, []dto.TestDriveInfo, error)
	AdminGetByID(id uint) (*dto.TestDriveInfo, error)
	UpdateStatus(id uint, req *dto.TestDriveStatusUpdateRequest) error
}

type testDriveService struct {
	repo repository.TestDriveRepository
}

// NewTestDriveService 创建试驾预约 service 实例
func NewTestDriveService(repo repository.TestDriveRepository) TestDriveService {
	return &testDriveService{repo: repo}
}

// testDriveStatusText 试驾状态文本
func testDriveStatusText(status int) string {
	switch status {
	case model.TestDriveStatusPending:
		return "待确认"
	case model.TestDriveStatusConfirmed:
		return "已确认"
	case model.TestDriveStatusInProgress:
		return "试驾中"
	case model.TestDriveStatusCompleted:
		return "已完成"
	case model.TestDriveStatusCanceled:
		return "已取消"
	case model.TestDriveStatusNoShow:
		return "未到店"
	}
	return ""
}

// toTestDriveInfo model -> dto
func toTestDriveInfo(t *model.CarTestDrive) *dto.TestDriveInfo {
	info := &dto.TestDriveInfo{
		ID:              t.ID,
		DriveNo:         t.DriveNo,
		CarID:           t.CarID,
		ListingID:       t.ListingID,
		UserID:          t.UserID,
		UserName:        t.UserName,
		UserPhone:       t.UserPhone,
		UserAvatar:      t.UserAvatar,
		DealerID:        t.DealerID,
		DealerName:      t.DealerName,
		SalesID:         t.SalesID,
		SalesName:       t.SalesName,
		AppointmentDate: t.AppointmentDate,
		AppointmentTime: t.AppointmentTime,
		Address:         t.Address,
		Latitude:        t.Latitude,
		Longitude:       t.Longitude,
		DriveType:       t.DriveType,
		LicenseStatus:   t.LicenseStatus,
		Remark:          t.Remark,
		Status:          t.Status,
		StatusText:      testDriveStatusText(t.Status),
		CancelReason:    t.CancelReason,
		Result:          t.Result,
		ResultRemark:    t.ResultRemark,
		StartedAt:       t.StartedAt,
		CompletedAt:     t.CompletedAt,
		CanceledAt:      t.CanceledAt,
		RegionID:        t.RegionID,
		CreatedAt:       t.CreatedAt,
		UpdatedAt:       t.UpdatedAt,
	}
	if t.LicenseImages != nil {
		info.LicenseImages = t.LicenseImages
	}
	return info
}

// genDriveNo 生成预约单号：TD + yyyyMMddHHmmss + 6 位随机
func genDriveNo() string {
	return fmt.Sprintf("TD%s%06d", time.Now().Format("20060102150405"), rand.Intn(1000000))
}

// parseDateRange 解析日期范围（YYYY-MM-DD）
func parseDateRange(startDate, endDate string) (*time.Time, *time.Time) {
	var st, et *time.Time
	if startDate != "" {
		if t, err := time.ParseInLocation("2006-01-02", startDate, time.Local); err == nil {
			st = &t
		}
	}
	if endDate != "" {
		if t, err := time.ParseInLocation("2006-01-02", endDate, time.Local); err == nil {
			// 包含整天
			end := t.Add(24*time.Hour - 1*time.Second)
			et = &end
		}
	}
	return st, et
}

// ===== C 端 =====

// Create 创建试驾预约
func (s *testDriveService) Create(regionID uint, userID uint, userName string, userPhone string, userAvatar string, req *dto.CreateTestDriveRequest) (*dto.TestDriveInfo, error) {
	t := &model.CarTestDrive{
		DriveNo:         genDriveNo(),
		CarID:           req.CarID,
		ListingID:       req.ListingID,
		UserID:          userID,
		UserName:        userName,
		UserPhone:       userPhone,
		UserAvatar:      userAvatar,
		DealerID:        req.DealerID,
		DealerName:      req.DealerName,
		SalesID:         req.SalesID,
		SalesName:       req.SalesName,
		AppointmentDate: req.AppointmentDate,
		AppointmentTime: req.AppointmentTime,
		Address:         req.Address,
		Latitude:        req.Latitude,
		Longitude:       req.Longitude,
		DriveType:       req.DriveType,
		Remark:          req.Remark,
		LicenseStatus:   model.LicenseStatusUnsubmitted,
		Status:          model.TestDriveStatusPending,
	}
	t.RegionID = regionID

	// 默认值兜底
	if t.DriveType == "" {
		t.DriveType = model.TestDriveTypeTestDrive
	}

	if err := s.repo.Create(t); err != nil {
		return nil, err
	}
	return toTestDriveInfo(t), nil
}

// Update 更新试驾预约
func (s *testDriveService) Update(id uint, operatorID uint, req *dto.UpdateTestDriveRequest) error {
	t, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrTestDriveNotFound
		}
		return err
	}
	if t.UserID != operatorID && operatorID != 0 {
		return ErrTestDriveNoPermission
	}

	fields := map[string]interface{}{}
	if req.SalesID != nil {
		fields["sales_id"] = *req.SalesID
	}
	if req.SalesName != nil {
		fields["sales_name"] = *req.SalesName
	}
	if req.AppointmentDate != nil {
		fields["appointment_date"] = req.AppointmentDate
	}
	if req.AppointmentTime != nil {
		fields["appointment_time"] = *req.AppointmentTime
	}
	if req.Address != nil {
		fields["address"] = *req.Address
	}
	if req.Latitude != nil {
		fields["latitude"] = *req.Latitude
	}
	if req.Longitude != nil {
		fields["longitude"] = *req.Longitude
	}
	if req.DriveType != nil {
		fields["drive_type"] = *req.DriveType
	}
	if req.LicenseStatus != nil {
		fields["license_status"] = *req.LicenseStatus
	}
	if req.LicenseImages != nil {
		if jb, err := model.FromJSON(req.LicenseImages); err == nil {
			fields["license_images"] = jb
		}
	}
	if req.Remark != nil {
		fields["remark"] = *req.Remark
	}
	if req.Status != nil {
		now := time.Now()
		switch *req.Status {
		case model.TestDriveStatusCanceled:
			fields["status"] = model.TestDriveStatusCanceled
			fields["canceled_at"] = &now
		case model.TestDriveStatusInProgress:
			fields["status"] = model.TestDriveStatusInProgress
			if t.StartedAt == nil {
				fields["started_at"] = &now
			}
		case model.TestDriveStatusCompleted:
			fields["status"] = model.TestDriveStatusCompleted
			fields["completed_at"] = &now
		default:
			fields["status"] = *req.Status
		}
	}

	if len(fields) == 0 {
		return nil
	}
	return s.repo.Update(id, fields)
}

// Cancel 用户取消预约
func (s *testDriveService) Cancel(id uint, userID uint, reason string) error {
	t, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrTestDriveNotFound
		}
		return err
	}
	if t.UserID != userID {
		return ErrTestDriveNoPermission
	}
	// 已完成/已取消的不能取消
	if t.Status == model.TestDriveStatusCompleted || t.Status == model.TestDriveStatusCanceled {
		return ErrTestDriveStatusInvalid
	}

	now := time.Now()
	fields := map[string]interface{}{
		"cancel_reason": reason,
		"canceled_at":   &now,
	}
	return s.repo.UpdateStatus(id, model.TestDriveStatusCanceled, fields)
}

// GetByID 获取详情
func (s *testDriveService) GetByID(id uint) (*dto.TestDriveInfo, error) {
	t, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTestDriveNotFound
		}
		return nil, err
	}
	return toTestDriveInfo(t), nil
}

// List C 端列表查询
func (s *testDriveService) List(regionID uint, req *dto.TestDriveListRequest) (*utils.Pagination, []dto.TestDriveInfo, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	startDate, endDate := parseDateRange(req.StartDate, req.EndDate)
	opts := repository.TestDriveListOptions{
		CarID:     req.CarID,
		ListingID: req.ListingID,
		UserID:    req.UserID,
		DealerID:  req.DealerID,
		SalesID:   req.SalesID,
		DriveType: req.DriveType,
		Status:    req.Status,
		StartDate: startDate,
		EndDate:   endDate,
	}
	list, total, err := s.repo.List(regionID, pagination, opts)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	result := make([]dto.TestDriveInfo, 0, len(list))
	for i := range list {
		result = append(result, *toTestDriveInfo(&list[i]))
	}
	return pagination, result, nil
}

// ListByUser 用户的预约
func (s *testDriveService) ListByUser(userID uint, page, pageSize int) (*utils.Pagination, []dto.TestDriveInfo, error) {
	pagination := utils.NewPagination(page, pageSize)
	list, total, err := s.repo.ListByUser(userID, pagination)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	result := make([]dto.TestDriveInfo, 0, len(list))
	for i := range list {
		result = append(result, *toTestDriveInfo(&list[i]))
	}
	return pagination, result, nil
}

// ListBySales 销售的预约
func (s *testDriveService) ListBySales(salesID uint, page, pageSize int) (*utils.Pagination, []dto.TestDriveInfo, error) {
	pagination := utils.NewPagination(page, pageSize)
	list, total, err := s.repo.ListBySales(salesID, pagination)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	result := make([]dto.TestDriveInfo, 0, len(list))
	for i := range list {
		result = append(result, *toTestDriveInfo(&list[i]))
	}
	return pagination, result, nil
}

// ListByDealer 车商的预约
func (s *testDriveService) ListByDealer(dealerID uint, page, pageSize int) (*utils.Pagination, []dto.TestDriveInfo, error) {
	pagination := utils.NewPagination(page, pageSize)
	list, total, err := s.repo.ListByDealer(dealerID, pagination)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	result := make([]dto.TestDriveInfo, 0, len(list))
	for i := range list {
		result = append(result, *toTestDriveInfo(&list[i]))
	}
	return pagination, result, nil
}

// UploadLicense 上传驾照
func (s *testDriveService) UploadLicense(id uint, userID uint, req *dto.TestDriveLicenseUploadRequest) error {
	t, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrTestDriveNotFound
		}
		return err
	}
	if t.UserID != userID {
		return ErrTestDriveNoPermission
	}

	var jb model.JSONB
	if req.LicenseImages != nil {
		var err error
		jb, err = model.FromJSON(req.LicenseImages)
		if err != nil {
			return err
		}
	}
	return s.repo.UpdateLicenseStatus(id, model.LicenseStatusPending, jb)
}

// ===== M 端管理 =====

func (s *testDriveService) AdminList(req *dto.TestDriveListRequest) (*utils.Pagination, []dto.TestDriveInfo, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	startDate, endDate := parseDateRange(req.StartDate, req.EndDate)
	opts := repository.TestDriveAdminListOptions{
		CarID:     req.CarID,
		UserID:    req.UserID,
		DealerID:  req.DealerID,
		SalesID:   req.SalesID,
		DriveType: req.DriveType,
		Status:    req.Status,
		StartDate: startDate,
		EndDate:   endDate,
	}
	list, total, err := s.repo.AdminList(pagination, opts)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	result := make([]dto.TestDriveInfo, 0, len(list))
	for i := range list {
		result = append(result, *toTestDriveInfo(&list[i]))
	}
	return pagination, result, nil
}

func (s *testDriveService) AdminGetByID(id uint) (*dto.TestDriveInfo, error) {
	t, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTestDriveNotFound
		}
		return nil, err
	}
	return toTestDriveInfo(t), nil
}

// UpdateStatus M 端更新状态（含确认/取消/试驾中/完成/未到店）
func (s *testDriveService) UpdateStatus(id uint, req *dto.TestDriveStatusUpdateRequest) error {
	t, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrTestDriveNotFound
		}
		return err
	}

	now := time.Now()
	fields := map[string]interface{}{}
	switch req.Status {
	case model.TestDriveStatusCanceled:
		fields["cancel_reason"] = req.CancelReason
		fields["canceled_at"] = &now
	case model.TestDriveStatusInProgress:
		if t.StartedAt == nil {
			fields["started_at"] = &now
		}
	case model.TestDriveStatusCompleted:
		fields["completed_at"] = &now
		if req.Result != "" {
			fields["result"] = req.Result
		}
		if req.ResultRemark != "" {
			fields["result_remark"] = req.ResultRemark
		}
	case model.TestDriveStatusNoShow:
		// 未到店，无需额外字段
	}
	return s.repo.UpdateStatus(id, req.Status, fields)
}
