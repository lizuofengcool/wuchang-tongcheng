// Package service 看房预约业务逻辑层
// 依据 v3.2.1 架构方案第五章：7 状态机
package service

import (
	"errors"
	"fmt"
	"time"

	"wuchang-tongcheng/internal/modules/house/dto"
	"wuchang-tongcheng/internal/modules/house/model"
	"wuchang-tongcheng/internal/modules/house/repository"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

var (
	ErrViewingNotFound     = errors.New("看房预约不存在")
	ErrViewingNoPermission = errors.New("无权操作此预约")
	ErrViewingStatus       = errors.New("预约状态不允许此操作")
)

// ViewingService 看房预约业务接口
type ViewingService interface {
	// C 端
	Create(regionID uint, userID uint, userName string, userPhone string, userAvatar string, req *dto.ViewingCreateRequest) (*dto.ViewingResponse, error)
	Update(id uint, operatorID uint, req *dto.ViewingUpdateRequest) error
	GetByID(id uint, userID uint) (*dto.ViewingResponse, error)
	List(regionID uint, req *dto.ViewingListQuery) (*utils.Pagination, []dto.ViewingResponse, error)
	ListMine(userID uint, page, pageSize int) (*utils.Pagination, []dto.ViewingResponse, error)
	Confirm(id uint, userID uint, req *dto.ViewingConfirmRequest) error
	Cancel(id uint, userID uint, req *dto.ViewingCancelRequest) error
	Reschedule(id uint, userID uint, req *dto.ViewingRescheduleRequest) error
	Complete(id uint, userID uint, req *dto.ViewingCompleteRequest) error

	// M 端
	AdminList(req *dto.ViewingAdminListQuery) (*utils.Pagination, []dto.ViewingResponse, error)
}

type viewingService struct {
	repo repository.ViewingRepository
}

// NewViewingService 创建 service 实例
func NewViewingService(repo repository.ViewingRepository) ViewingService {
	return &viewingService{repo: repo}
}

// toViewingInfo model -> dto
func toViewingInfo(v *model.HouseViewing) *dto.ViewingResponse {
	return &dto.ViewingResponse{
		ID:              v.ID,
		ViewingNo:       v.ViewingNo,
		HouseID:         v.HouseID,
		ListingID:       v.ListingID,
		CommunityID:     v.CommunityID,
		UserID:          v.UserID,
		UserName:        v.UserName,
		UserPhone:       v.UserPhone,
		UserAvatar:      v.UserAvatar,
		AgentID:         v.AgentID,
		AgentName:       v.AgentName,
		AgentPhone:      v.AgentPhone,
		ScheduledAt:     v.ScheduledAt,
		DurationMinutes: v.DurationMinutes,
		ViewingType:     v.ViewingType,
		ViewingTypeText: viewingTypeText(v.ViewingType),
		OnlineURL:       v.OnlineURL,
		OnlinePassword:  v.OnlinePassword,
		MeetLocation:    v.MeetLocation,
		Remark:          v.Remark,
		Status:          v.Status,
		StatusText:      viewingStatusText(v.Status),
		Result:          v.Result,
		ResultText:      viewingResultText(v.Result),
		Feedback:        v.Feedback,
		Rating:          v.Rating,
		AttendedAt:      v.AttendedAt,
		CompletedAt:     v.CompletedAt,
		CanceledAt:      v.CanceledAt,
		CanceledReason:  v.CanceledReason,
		CanceledBy:      v.CanceledBy,
		ReminderSent:    v.ReminderSent,
		ReminderSentAt:  v.ReminderSentAt,
		RegionID:        v.RegionID,
		CreatedAt:       v.CreatedAt,
		UpdatedAt:       v.UpdatedAt,
	}
}

func viewingTypeText(s string) string {
	switch s {
	case model.ViewingTypeOffline:
		return "线下看房"
	case model.ViewingTypeOnline:
		return "在线看房"
	case model.ViewingTypeVR:
		return "VR看房"
	}
	return "线下看房"
}

func viewingStatusText(s int) string {
	switch s {
	case model.ViewingStatusPending:
		return "待确认"
	case model.ViewingStatusConfirmed:
		return "已确认"
	case model.ViewingStatusInProgress:
		return "进行中"
	case model.ViewingStatusCompleted:
		return "已完成"
	case model.ViewingStatusCanceled:
		return "已取消"
	case model.ViewingStatusNoShow:
		return "已爽约"
	}
	return "待确认"
}

func viewingResultText(s string) string {
	switch s {
	case model.ViewingResultPending:
		return "待反馈"
	case model.ViewingResultSatisfied:
		return "满意"
	case model.ViewingResultNeutral:
		return "一般"
	case model.ViewingResultDissatisfied:
		return "不满意"
	case model.ViewingResultNoShow:
		return "爽约"
	}
	return "待反馈"
}

// generateViewingNo 生成预约单号
func generateViewingNo() string {
	return fmt.Sprintf("VW%s", time.Now().Format("20060102150405.000"))
}

// ===== C 端 =====

func (s *viewingService) Create(regionID uint, userID uint, userName string, userPhone string, userAvatar string, req *dto.ViewingCreateRequest) (*dto.ViewingResponse, error) {
	v := &model.HouseViewing{
		ViewingNo:       generateViewingNo(),
		HouseID:         req.HouseID,
		ListingID:       req.ListingID,
		AgentID:         req.AgentID,
		UserID:          userID,
		UserName:        userName,
		UserPhone:       userPhone,
		UserAvatar:      userAvatar,
		ScheduledAt:     req.ScheduledAt,
		DurationMinutes: req.DurationMinutes,
		ViewingType:     req.ViewingType,
		OnlineURL:       req.OnlineURL,
		OnlinePassword:  req.OnlinePassword,
		MeetLocation:    req.MeetLocation,
		Remark:          req.Remark,
		Status:          model.ViewingStatusPending,
	}
	v.RegionID = regionID

	if v.ViewingType == "" {
		v.ViewingType = model.ViewingTypeOffline
	}
	if v.DurationMinutes == 0 {
		v.DurationMinutes = 30
	}

	if err := s.repo.Create(v); err != nil {
		return nil, err
	}
	return toViewingInfo(v), nil
}

func (s *viewingService) Update(id uint, operatorID uint, req *dto.ViewingUpdateRequest) error {
	v, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrViewingNotFound
		}
		return err
	}
	// 仅预约人或经纪人在确认前可改
	if v.UserID != operatorID && v.AgentID != operatorID {
		return ErrViewingNoPermission
	}
	if v.Status != model.ViewingStatusPending {
		return ErrViewingStatus
	}

	fields := map[string]interface{}{}
	if req.ScheduledAt != nil {
		fields["scheduled_at"] = req.ScheduledAt
	}
	if req.DurationMinutes > 0 {
		fields["duration_minutes"] = req.DurationMinutes
	}
	if req.ViewingType != "" {
		fields["viewing_type"] = req.ViewingType
	}
	if req.OnlineURL != "" {
		fields["online_url"] = req.OnlineURL
	}
	if req.OnlinePassword != "" {
		fields["online_password"] = req.OnlinePassword
	}
	if req.MeetLocation != "" {
		fields["meet_location"] = req.MeetLocation
	}
	if req.Remark != "" {
		fields["remark"] = req.Remark
	}

	if len(fields) == 0 {
		return nil
	}
	return s.repo.UpdateFields(id, fields)
}

func (s *viewingService) GetByID(id uint, userID uint) (*dto.ViewingResponse, error) {
	v, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrViewingNotFound
		}
		return nil, err
	}
	if userID > 0 && v.UserID != userID && v.AgentID != userID {
		return nil, ErrViewingNoPermission
	}
	return toViewingInfo(v), nil
}

func (s *viewingService) List(regionID uint, req *dto.ViewingListQuery) (*utils.Pagination, []dto.ViewingResponse, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	opts := repository.ViewingListOptions{
		HouseID:     req.HouseID,
		ListingID:   req.ListingID,
		AgentID:     req.AgentID,
		UserID:      req.UserID,
		ViewingType: req.ViewingType,
		Status:      req.Status,
		Result:      req.Result,
		StartDate:   req.StartDate,
		EndDate:     req.EndDate,
	}
	list, total, err := s.repo.List(regionID, pagination, opts)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total

	result := make([]dto.ViewingResponse, 0, len(list))
	for i := range list {
		result = append(result, *toViewingInfo(&list[i]))
	}
	return pagination, result, nil
}

func (s *viewingService) ListMine(userID uint, page, pageSize int) (*utils.Pagination, []dto.ViewingResponse, error) {
	pagination := utils.NewPagination(page, pageSize)
	list, total, err := s.repo.ListByUser(userID, pagination)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total

	result := make([]dto.ViewingResponse, 0, len(list))
	for i := range list {
		result = append(result, *toViewingInfo(&list[i]))
	}
	return pagination, result, nil
}

// Confirm 经纪人确认预约
func (s *viewingService) Confirm(id uint, userID uint, req *dto.ViewingConfirmRequest) error {
	v, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrViewingNotFound
		}
		return err
	}
	if v.AgentID != userID && v.UserID != userID {
		return ErrViewingNoPermission
	}
	if v.Status != model.ViewingStatusPending {
		return ErrViewingStatus
	}
	fields := map[string]interface{}{
		"status": model.ViewingStatusConfirmed,
	}
	if req.MeetLocation != "" {
		fields["meet_location"] = req.MeetLocation
	}
	if req.Remark != "" {
		fields["remark"] = req.Remark
	}
	return s.repo.UpdateFields(id, fields)
}

// Cancel 取消预约
func (s *viewingService) Cancel(id uint, userID uint, req *dto.ViewingCancelRequest) error {
	v, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrViewingNotFound
		}
		return err
	}
	if v.UserID != userID && v.AgentID != userID {
		return ErrViewingNoPermission
	}
	if v.Status == model.ViewingStatusCompleted || v.Status == model.ViewingStatusCanceled {
		return ErrViewingStatus
	}
	now := time.Now()
	return s.repo.UpdateFields(id, map[string]interface{}{
		"status":          model.ViewingStatusCanceled,
		"canceled_at":     &now,
		"canceled_reason": req.Reason,
		"canceled_by":     userID,
	})
}

// Reschedule 改期
func (s *viewingService) Reschedule(id uint, userID uint, req *dto.ViewingRescheduleRequest) error {
	v, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrViewingNotFound
		}
		return err
	}
	if v.UserID != userID && v.AgentID != userID {
		return ErrViewingNoPermission
	}
	if v.Status == model.ViewingStatusCompleted || v.Status == model.ViewingStatusCanceled {
		return ErrViewingStatus
	}
	fields := map[string]interface{}{
		"scheduled_at": req.ScheduledAt,
		"status":       model.ViewingStatusPending, // 改期后回到待确认
	}
	if req.DurationMinutes > 0 {
		fields["duration_minutes"] = req.DurationMinutes
	}
	if req.Remark != "" {
		fields["remark"] = req.Remark
	}
	return s.repo.UpdateFields(id, fields)
}

// Complete 完成看房
func (s *viewingService) Complete(id uint, userID uint, req *dto.ViewingCompleteRequest) error {
	v, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrViewingNotFound
		}
		return err
	}
	if v.UserID != userID && v.AgentID != userID {
		return ErrViewingNoPermission
	}
	if v.Status != model.ViewingStatusConfirmed && v.Status != model.ViewingStatusInProgress {
		return ErrViewingStatus
	}
	now := time.Now()
	result := req.Result
	if result == "" {
		result = model.ViewingResultPending
	}
	fields := map[string]interface{}{
		"status":       model.ViewingStatusCompleted,
		"completed_at": &now,
		"result":       result,
		"feedback":     req.Feedback,
		"rating":       req.Rating,
	}
	if v.AttendedAt == nil {
		fields["attended_at"] = &now
	}
	return s.repo.UpdateFields(id, fields)
}

// ===== M 端 =====

func (s *viewingService) AdminList(req *dto.ViewingAdminListQuery) (*utils.Pagination, []dto.ViewingResponse, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	opts := repository.ViewingAdminListOptions{
		RegionID: req.RegionID,
		HouseID:  req.HouseID,
		UserID:   req.UserID,
		AgentID:  req.AgentID,
		Status:   req.Status,
	}
	list, total, err := s.repo.AdminList(pagination, opts)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total

	result := make([]dto.ViewingResponse, 0, len(list))
	for i := range list {
		result = append(result, *toViewingInfo(&list[i]))
	}
	return pagination, result, nil
}
