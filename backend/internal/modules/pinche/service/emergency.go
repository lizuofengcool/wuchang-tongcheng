// Package service 同城拼车出行业务逻辑层 - 紧急联系人/一键报警
package service

import (
	"errors"
	"time"

	"wuchang-tongcheng/internal/modules/pinche/dto"
	"wuchang-tongcheng/internal/modules/pinche/model"
	"wuchang-tongcheng/internal/modules/pinche/repository"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

var (
	ErrEmergencyNotFound      = errors.New("紧急联系人/报警不存在")
	ErrEmergencyNoPermission  = errors.New("无权操作此紧急联系人/报警")
	ErrEmergencyStatusInvalid = errors.New("报警状态不允许此操作")
)

// EmergencyService 紧急联系人/报警业务接口
type EmergencyService interface {
	// C 端 - 紧急联系人
	CreateContact(regionID uint, userID uint, req *dto.CreateEmergencyContactRequest) (*dto.EmergencyInfo, error)
	UpdateContact(id uint, operatorID uint, req *dto.UpdateEmergencyContactRequest) error
	DeleteContact(id uint, operatorID uint) error
	ListContacts(userID uint, page, pageSize int) (*utils.Pagination, []dto.EmergencyInfo, error)

	// C 端 - 一键报警
	SOS(regionID uint, userID uint, req *dto.SOSAlertRequest) (*dto.EmergencyInfo, error)
	GetAlert(id uint) (*dto.EmergencyInfo, error)
	ListMyAlerts(userID uint, page, pageSize int) (*utils.Pagination, []dto.EmergencyInfo, error)
	ListAlertsByPinche(pincheID uint, page, pageSize int) (*utils.Pagination, []dto.EmergencyInfo, error)
	ListAlertsByTrip(tripID uint, page, pageSize int) (*utils.Pagination, []dto.EmergencyInfo, error)

	// M 端
	AdminListAlerts(req *dto.EmergencyListRequest) (*utils.Pagination, []dto.EmergencyInfo, error)
	HandleAlert(id uint, handlerID uint, req *dto.HandleAlertRequest) error
	UpdateAlertStatus(id uint, status int) error
}

type emergencyService struct {
	repo repository.EmergencyRepository
}

// NewEmergencyService 创建紧急联系人/报警 service 实例
func NewEmergencyService(repo repository.EmergencyRepository) EmergencyService {
	return &emergencyService{repo: repo}
}

// alertTypeText 类型文本
func alertTypeText(t string) string {
	switch t {
	case model.AlertTypeSOS:
		return "一键报警"
	case model.AlertTypeShare:
		return "行程分享"
	case model.AlertTypePeriodic:
		return "定时上报"
	}
	return ""
}

// alertStatusText 状态文本
func alertStatusText(s int) string {
	switch s {
	case model.AlertStatusPending:
		return "未处理"
	case model.AlertStatusHandling:
		return "处理中"
	case model.AlertStatusHandled:
		return "已处理"
	}
	return ""
}

// toEmergencyInfo model -> dto
func toEmergencyInfo(e *model.PincheEmergency) *dto.EmergencyInfo {
	info := &dto.EmergencyInfo{
		ID:                e.ID,
		RegionID:          e.RegionID,
		UserID:            e.UserID,
		ContactName:       e.ContactName,
		ContactPhone:      e.ContactPhone,
		ContactRelation:   e.ContactRelation,
		IsPrimary:         e.IsPrimary,
		PincheID:          e.PincheID,
		TripID:            e.TripID,
		AlertType:         e.AlertType,
		AlertTypeText:     alertTypeText(e.AlertType),
		AlertStatus:       e.AlertStatus,
		AlertStatusText:   alertStatusText(e.AlertStatus),
		AlertTime:         e.AlertTime,
		AlertLocationLat:  e.AlertLocationLat,
		AlertLocationLng:  e.AlertLocationLng,
		AlertAddress:      e.AlertAddress,
		AlertDescription:  e.AlertDescription,
		HandledAt:         e.HandledAt,
		HandlerID:         e.HandlerID,
		HandleResult:      e.HandleResult,
		CreatedAt:         e.CreatedAt,
	}
	if e.AlertEvidence != nil {
		info.AlertEvidence = e.AlertEvidence
	}
	return info
}

// CreateContact 创建紧急联系人
func (s *emergencyService) CreateContact(regionID uint, userID uint, req *dto.CreateEmergencyContactRequest) (*dto.EmergencyInfo, error) {
	e := &model.PincheEmergency{
		UserID:          userID,
		ContactName:     req.ContactName,
		ContactPhone:    req.ContactPhone,
		ContactRelation: req.ContactRelation,
		IsPrimary:       req.IsPrimary,
	}
	e.RegionID = regionID
	if err := s.repo.Create(e); err != nil {
		return nil, err
	}
	return toEmergencyInfo(e), nil
}

// UpdateContact 更新紧急联系人
func (s *emergencyService) UpdateContact(id uint, operatorID uint, req *dto.UpdateEmergencyContactRequest) error {
	e, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrEmergencyNotFound
		}
		return err
	}
	if e.UserID != operatorID {
		return ErrEmergencyNoPermission
	}
	fields := map[string]interface{}{}
	if req.ContactName != nil {
		fields["contact_name"] = *req.ContactName
	}
	if req.ContactPhone != nil {
		fields["contact_phone"] = *req.ContactPhone
	}
	if req.ContactRelation != nil {
		fields["contact_relation"] = *req.ContactRelation
	}
	if req.IsPrimary != nil {
		fields["is_primary"] = *req.IsPrimary
	}
	if len(fields) == 0 {
		return nil
	}
	return s.repo.Update(id, fields)
}

// DeleteContact 删除紧急联系人
func (s *emergencyService) DeleteContact(id uint, operatorID uint) error {
	e, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrEmergencyNotFound
		}
		return err
	}
	if e.UserID != operatorID {
		return ErrEmergencyNoPermission
	}
	return s.repo.Delete(id)
}

// ListContacts 紧急联系人列表
func (s *emergencyService) ListContacts(userID uint, page, pageSize int) (*utils.Pagination, []dto.EmergencyInfo, error) {
	pagination := utils.NewPagination(page, pageSize)
	list, total, err := s.repo.ListContacts(userID, pagination)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	result := make([]dto.EmergencyInfo, 0, len(list))
	for i := range list {
		result = append(result, *toEmergencyInfo(&list[i]))
	}
	return pagination, result, nil
}

// SOS 一键报警
func (s *emergencyService) SOS(regionID uint, userID uint, req *dto.SOSAlertRequest) (*dto.EmergencyInfo, error) {
	alertType := req.AlertType
	if alertType == "" {
		alertType = model.AlertTypeSOS
	}
	now := time.Now()
	e := &model.PincheEmergency{
		UserID:            userID,
		PincheID:          uintToPtr(req.PincheID),
		TripID:            req.TripID,
		AlertType:         alertType,
		AlertStatus:       model.AlertStatusPending,
		AlertTime:         &now,
		AlertLocationLat:  req.AlertLocationLat,
		AlertLocationLng:  req.AlertLocationLng,
		AlertAddress:      req.AlertAddress,
		AlertDescription:  req.AlertDescription,
	}
	e.RegionID = regionID
	if req.AlertEvidence != nil {
		if jb, err := model.FromJSON(req.AlertEvidence); err == nil {
			e.AlertEvidence = jb
		}
	}
	if err := s.repo.Create(e); err != nil {
		return nil, err
	}
	return toEmergencyInfo(e), nil
}

// uintToPtr 将 uint 转换为 *uint（若为 0 返回 nil）
func uintToPtr(v uint) *uint {
	if v == 0 {
		return nil
	}
	return &v
}

// intPtr 已在 pinche.go service 中定义，此处避免重复定义
// 若未定义，则启用下方定义
// var intPtr = func(v uint) *uint { return &v }

// GetAlert 获取报警详情
func (s *emergencyService) GetAlert(id uint) (*dto.EmergencyInfo, error) {
	e, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrEmergencyNotFound
		}
		return nil, err
	}
	return toEmergencyInfo(e), nil
}

// ListMyAlerts 我的报警列表
func (s *emergencyService) ListMyAlerts(userID uint, page, pageSize int) (*utils.Pagination, []dto.EmergencyInfo, error) {
	pagination := utils.NewPagination(page, pageSize)
	opts := repository.EmergencyListOptions{UserID: userID}
	list, total, err := s.repo.ListAlerts(0, pagination, opts)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	result := make([]dto.EmergencyInfo, 0, len(list))
	for i := range list {
		result = append(result, *toEmergencyInfo(&list[i]))
	}
	return pagination, result, nil
}

// ListAlertsByPinche 按行程查询报警
func (s *emergencyService) ListAlertsByPinche(pincheID uint, page, pageSize int) (*utils.Pagination, []dto.EmergencyInfo, error) {
	pagination := utils.NewPagination(page, pageSize)
	list, total, err := s.repo.ListByPinche(pincheID, pagination)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	result := make([]dto.EmergencyInfo, 0, len(list))
	for i := range list {
		result = append(result, *toEmergencyInfo(&list[i]))
	}
	return pagination, result, nil
}

// ListAlertsByTrip 按行程完成记录查询报警
func (s *emergencyService) ListAlertsByTrip(tripID uint, page, pageSize int) (*utils.Pagination, []dto.EmergencyInfo, error) {
	pagination := utils.NewPagination(page, pageSize)
	list, total, err := s.repo.ListByTrip(tripID, pagination)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	result := make([]dto.EmergencyInfo, 0, len(list))
	for i := range list {
		result = append(result, *toEmergencyInfo(&list[i]))
	}
	return pagination, result, nil
}

// AdminListAlerts 管理后台报警列表
func (s *emergencyService) AdminListAlerts(req *dto.EmergencyListRequest) (*utils.Pagination, []dto.EmergencyInfo, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	opts := repository.EmergencyListOptions{
		UserID:      req.UserID,
		PincheID:    req.PincheID,
		TripID:      req.TripID,
		AlertType:   req.AlertType,
		AlertStatus: req.AlertStatus,
	}
	// 跨地区：regionID=0
	list, total, err := s.repo.ListAlerts(0, pagination, opts)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	result := make([]dto.EmergencyInfo, 0, len(list))
	for i := range list {
		result = append(result, *toEmergencyInfo(&list[i]))
	}
	return pagination, result, nil
}

// HandleAlert 处理报警
func (s *emergencyService) HandleAlert(id uint, handlerID uint, req *dto.HandleAlertRequest) error {
	e, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrEmergencyNotFound
		}
		return err
	}
	if e.AlertStatus == model.AlertStatusHandled {
		return ErrEmergencyStatusInvalid
	}
	now := time.Now()
	fields := map[string]interface{}{
		"alert_status":  model.AlertStatusHandled,
		"handler_id":    &handlerID,
		"handle_result": req.HandleResult,
		"handled_at":    &now,
	}
	return s.repo.Update(id, fields)
}

// UpdateAlertStatus 管理后台更新报警状态
func (s *emergencyService) UpdateAlertStatus(id uint, status int) error {
	return s.repo.UpdateAlertStatus(id, status)
}
