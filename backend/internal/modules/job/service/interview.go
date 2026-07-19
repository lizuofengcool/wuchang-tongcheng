// Package service 面试邀约业务逻辑层
// 依据 v3.2.1 架构方案第四章：对标 BOSS直聘多轮面试 + Offer
// 状态机：0待确认/1已确认/2已改期/3已参加/4已完成/5已取消/6未到面
// 结果：pending/pass/reject/next_round/offer
package service

import (
	"errors"
	"fmt"
	"math/rand"
	"time"

	"wuchang-tongcheng/internal/modules/job/dto"
	"wuchang-tongcheng/internal/modules/job/model"
	"wuchang-tongcheng/internal/modules/job/repository"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

var (
	ErrInterviewNotFound     = errors.New("面试邀约不存在")
	ErrInterviewNoPermission = errors.New("无权操作此面试")
	ErrInterviewStatus       = errors.New("面试状态不允许此操作")
)

// InterviewService 面试业务接口
type InterviewService interface {
	Create(regionID uint, recruiterID uint, req *dto.InterviewCreateRequest) (*dto.InterviewResponse, error)
	Update(id uint, operatorID uint, req *dto.InterviewUpdateRequest) error
	GetByID(id uint, userID uint) (*dto.InterviewResponse, error)
	List(userID uint, req *dto.InterviewListQuery) (*utils.Pagination, []dto.InterviewResponse, error)
	ListByApplication(applicationID uint, userID uint) ([]dto.InterviewResponse, error)
	Action(id uint, operatorID uint, req *dto.InterviewActionRequest) (*dto.InterviewResponse, error)
	Feedback(id uint, operatorID uint, req *dto.InterviewFeedbackRequest) (*dto.InterviewResponse, error)
	Stats(userID uint, role string) (*dto.InterviewStatsResponse, error)
}

type interviewService struct {
	repo            repository.InterviewRepository
	applicationRepo repository.ApplicationRepository
}

// NewInterviewService 创建面试 service 实例
func NewInterviewService(repo repository.InterviewRepository, applicationRepo repository.ApplicationRepository) InterviewService {
	return &interviewService{repo: repo, applicationRepo: applicationRepo}
}

// genInterviewNo 生成面试单号：JI + yyyyMMddHHmmss + 6位随机数
func genInterviewNo() string {
	return fmt.Sprintf("JI%s%06d", time.Now().Format("20060102150405"), rand.Intn(1000000))
}

// interviewStatusText 面试状态文本
func interviewStatusText(status int) string {
	switch status {
	case model.InterviewStatusPending:
		return "待确认"
	case model.InterviewStatusConfirmed:
		return "已确认"
	case model.InterviewStatusRescheduled:
		return "已改期"
	case model.InterviewStatusAttended:
		return "已参加"
	case model.InterviewStatusCompleted:
		return "已完成"
	case model.InterviewStatusCanceled:
		return "已取消"
	case model.InterviewStatusNoShow:
		return "未到面"
	}
	return "未知"
}

// interviewResultText 面试结果文本
func interviewResultText(result string) string {
	switch result {
	case model.InterviewResultPass:
		return "通过"
	case model.InterviewResultReject:
		return "不通过"
	case model.InterviewResultNextRound:
		return "进入下一轮"
	case model.InterviewResultOffer:
		return "发放Offer"
	}
	return "待定"
}

// interviewTypeText 面试方式文本
func interviewTypeText(t string) string {
	switch t {
	case model.InterviewTypeOnsite:
		return "现场面试"
	case model.InterviewTypeOnline:
		return "在线视频"
	case model.InterviewTypePhone:
		return "电话面试"
	case model.InterviewTypeVideo:
		return "视频面试"
	}
	return "现场面试"
}

// toInterviewResponse model -> dto
func toInterviewResponse(i *model.JobInterview) *dto.InterviewResponse {
	resp := &dto.InterviewResponse{
		ID:                  i.ID,
		InterviewNo:         i.InterviewNo,
		ApplicationID:       i.ApplicationID,
		JobID:               i.JobID,
		ApplicantID:         i.ApplicantID,
		RecruiterID:         i.RecruiterID,
		CompanyID:           i.CompanyID,
		Round:               i.Round,
		InterviewType:       i.InterviewType,
		InterviewTypeText:   interviewTypeText(i.InterviewType),
		ScheduledAt:         i.ScheduledAt,
		DurationMinutes:     i.DurationMinutes,
		Location:            i.Location,
		OnlineURL:           i.OnlineURL,
		OnlinePassword:      i.OnlinePassword,
		InterviewerName:     i.InterviewerName,
		InterviewerPosition: i.InterviewerPosition,
		ContactPhone:        i.ContactPhone,
		Status:              i.Status,
		StatusText:          interviewStatusText(i.Status),
		Result:              i.Result,
		ResultText:          interviewResultText(i.Result),
		Feedback:            i.Feedback,
		Rating:              i.Rating,
		SalaryOffered:       i.SalaryOffered,
		PositionOffered:     i.PositionOffered,
		EntryDate:           i.EntryDate,
		Attachments:         []map[string]interface{}{},
		ConfirmedAt:         i.ConfirmedAt,
		AttendedAt:          i.AttendedAt,
		CompletedAt:         i.CompletedAt,
		CanceledAt:          i.CanceledAt,
		CanceledReason:      i.CanceledReason,
		RegionID:            i.RegionID,
		CreatedAt:           i.CreatedAt,
		UpdatedAt:           i.UpdatedAt,
	}
	if i.Attachments != nil {
		var arr []map[string]interface{}
		_ = i.Attachments.Parse(&arr)
		if arr != nil {
			resp.Attachments = arr
		}
	}
	return resp
}

func (s *interviewService) Create(regionID uint, recruiterID uint, req *dto.InterviewCreateRequest) (*dto.InterviewResponse, error) {
	// 校验投递记录
	a, err := s.applicationRepo.FindByID(req.ApplicationID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrApplicationNotFound
		}
		return nil, err
	}
	if a.RecruiterID != recruiterID {
		return nil, ErrInterviewNoPermission
	}

	interviewType := req.InterviewType
	if interviewType == "" {
		interviewType = model.InterviewTypeOnsite
	}

	round := req.Round
	if round <= 0 {
		// 自动计算轮次：当前投递的面试次数 + 1
		count, _ := s.repo.CountByApplicationID(req.ApplicationID)
		round = int(count) + 1
	}

	duration := req.DurationMinutes
	if duration <= 0 {
		duration = 60
	}

	i := &model.JobInterview{
		InterviewNo:        genInterviewNo(),
		ApplicationID:      req.ApplicationID,
		JobID:              a.JobID,
		ApplicantID:        a.ApplicantID,
		RecruiterID:        a.RecruiterID,
		CompanyID:          a.CompanyID,
		Round:              round,
		InterviewType:      interviewType,
		ScheduledAt:        req.ScheduledAt,
		DurationMinutes:    duration,
		Location:           req.Location,
		OnlineURL:          req.OnlineURL,
		OnlinePassword:     req.OnlinePassword,
		InterviewerName:    req.InterviewerName,
		InterviewerPosition: req.InterviewerPosition,
		ContactPhone:       req.ContactPhone,
		Status:             model.InterviewStatusPending,
		Result:             model.InterviewResultPending,
	}
	i.RegionID = regionID

	if err := s.repo.Create(i); err != nil {
		return nil, err
	}

	// 同步更新投递状态为"面试邀约"
	_ = s.applicationRepo.Update(a.ID, map[string]interface{}{
		"status":    model.ApplicationStatusInterview,
		"replied_at": time.Now(),
	})
	// 投递记录面试次数 +1
	_ = s.applicationRepo.IncrInterviewCount(a.ID)

	return toInterviewResponse(i), nil
}

func (s *interviewService) Update(id uint, operatorID uint, req *dto.InterviewUpdateRequest) error {
	i, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrInterviewNotFound
		}
		return err
	}
	if i.RecruiterID != operatorID {
		return ErrInterviewNoPermission
	}
	// 仅待确认/已确认状态下可修改
	if i.Status != model.InterviewStatusPending && i.Status != model.InterviewStatusConfirmed {
		return ErrInterviewStatus
	}

	fields := make(map[string]interface{})
	if req.ScheduledAt != nil {
		fields["scheduled_at"] = req.ScheduledAt
	}
	if req.DurationMinutes > 0 {
		fields["duration_minutes"] = req.DurationMinutes
	}
	if req.Location != "" {
		fields["location"] = req.Location
	}
	if req.OnlineURL != "" {
		fields["online_url"] = req.OnlineURL
	}
	if req.OnlinePassword != "" {
		fields["online_password"] = req.OnlinePassword
	}
	if req.InterviewerName != "" {
		fields["interviewer_name"] = req.InterviewerName
	}
	if req.InterviewerPosition != "" {
		fields["interviewer_position"] = req.InterviewerPosition
	}
	if req.ContactPhone != "" {
		fields["contact_phone"] = req.ContactPhone
	}
	return s.repo.Update(id, fields)
}

func (s *interviewService) GetByID(id uint, userID uint) (*dto.InterviewResponse, error) {
	i, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInterviewNotFound
		}
		return nil, err
	}
	if userID > 0 && i.ApplicantID != userID && i.RecruiterID != userID {
		return nil, ErrInterviewNoPermission
	}
	return toInterviewResponse(i), nil
}

func (s *interviewService) List(userID uint, req *dto.InterviewListQuery) (*utils.Pagination, []dto.InterviewResponse, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	list, total, err := s.repo.List(repository.InterviewListQuery{
		UserID:        userID,
		Role:          req.Role,
		Status:        req.Status,
		Result:        req.Result,
		JobID:         req.JobID,
		ApplicationID: req.ApplicationID,
		CompanyID:     req.CompanyID,
		InterviewNo:   req.InterviewNo,
		StartTime:     req.StartTime,
		EndTime:       req.EndTime,
	}, pagination)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	result := make([]dto.InterviewResponse, 0, len(list))
	for j := range list {
		result = append(result, *toInterviewResponse(&list[j]))
	}
	return pagination, result, nil
}

func (s *interviewService) ListByApplication(applicationID uint, userID uint) ([]dto.InterviewResponse, error) {
	// 校验权限
	a, err := s.applicationRepo.FindByID(applicationID)
	if err != nil {
		return nil, ErrApplicationNotFound
	}
	if userID > 0 && a.ApplicantID != userID && a.RecruiterID != userID {
		return nil, ErrInterviewNoPermission
	}
	list, err := s.repo.ListByApplicationID(applicationID)
	if err != nil {
		return nil, err
	}
	result := make([]dto.InterviewResponse, 0, len(list))
	for i := range list {
		result = append(result, *toInterviewResponse(&list[i]))
	}
	return result, nil
}

// Action 面试操作
//   - confirm: 待确认→已确认（求职者）
//   - reschedule: 待确认/已确认→已改期（双方，需 NewScheduledAt）
//   - attend: 已确认→已参加（双方）
//   - complete: 已参加→已完成（招聘者）
//   - cancel: 任意状态→已取消（双方）
//   - noshow: 已确认/已参加→未到面（招聘者）
func (s *interviewService) Action(id uint, operatorID uint, req *dto.InterviewActionRequest) (*dto.InterviewResponse, error) {
	i, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInterviewNotFound
		}
		return nil, err
	}
	// 权限校验
	isApplicant := i.ApplicantID == operatorID
	isRecruiter := i.RecruiterID == operatorID
	if operatorID > 0 && !isApplicant && !isRecruiter {
		return nil, ErrInterviewNoPermission
	}

	now := time.Now()
	fields := make(map[string]interface{})

	switch req.Action {
	case "confirm":
		// 仅求职者可确认
		if operatorID > 0 && !isApplicant {
			return nil, ErrInterviewNoPermission
		}
		if i.Status != model.InterviewStatusPending {
			return nil, ErrInterviewStatus
		}
		fields["status"] = model.InterviewStatusConfirmed
		fields["confirmed_at"] = &now

	case "reschedule":
		if req.NewScheduledAt == nil {
			return nil, errors.New("改期需提供新的面试时间")
		}
		if i.Status != model.InterviewStatusPending && i.Status != model.InterviewStatusConfirmed {
			return nil, ErrInterviewStatus
		}
		fields["status"] = model.InterviewStatusRescheduled
		fields["scheduled_at"] = req.NewScheduledAt
		if req.Reason != "" {
			fields["canceled_reason"] = req.Reason
		}

	case "attend":
		if i.Status != model.InterviewStatusConfirmed && i.Status != model.InterviewStatusRescheduled {
			return nil, ErrInterviewStatus
		}
		fields["status"] = model.InterviewStatusAttended
		fields["attended_at"] = &now

	case "complete":
		// 仅招聘者可标记完成
		if operatorID > 0 && !isRecruiter {
			return nil, ErrInterviewNoPermission
		}
		if i.Status != model.InterviewStatusAttended {
			return nil, ErrInterviewStatus
		}
		fields["status"] = model.InterviewStatusCompleted
		fields["completed_at"] = &now

	case "cancel":
		if i.Status == model.InterviewStatusCompleted || i.Status == model.InterviewStatusCanceled {
			return nil, ErrInterviewStatus
		}
		fields["status"] = model.InterviewStatusCanceled
		fields["canceled_at"] = &now
		if req.Reason != "" {
			fields["canceled_reason"] = req.Reason
		}

	case "noshow":
		// 仅招聘者可标记未到面
		if operatorID > 0 && !isRecruiter {
			return nil, ErrInterviewNoPermission
		}
		if i.Status != model.InterviewStatusConfirmed && i.Status != model.InterviewStatusAttended {
			return nil, ErrInterviewStatus
		}
		fields["status"] = model.InterviewStatusNoShow
		if req.Reason != "" {
			fields["canceled_reason"] = req.Reason
		}

	default:
		return nil, errors.New("无效的操作类型")
	}

	if err := s.repo.Update(id, fields); err != nil {
		return nil, err
	}
	updated, _ := s.repo.FindByID(id)
	return toInterviewResponse(updated), nil
}

// Feedback 面试反馈（招聘者提交）
// result: pending/pass/reject/next_round/offer
func (s *interviewService) Feedback(id uint, operatorID uint, req *dto.InterviewFeedbackRequest) (*dto.InterviewResponse, error) {
	i, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInterviewNotFound
		}
		return nil, err
	}
	if operatorID > 0 && i.RecruiterID != operatorID {
		return nil, ErrInterviewNoPermission
	}
	// 仅已完成/已参加状态可反馈
	if i.Status != model.InterviewStatusCompleted && i.Status != model.InterviewStatusAttended {
		return nil, ErrInterviewStatus
	}

	now := time.Now()
	fields := map[string]interface{}{
		"result":   req.Result,
		"feedback": req.Feedback,
		"rating":   req.Rating,
	}
	if req.SalaryOffered > 0 {
		fields["salary_offered"] = req.SalaryOffered
	}
	if req.PositionOffered != "" {
		fields["position_offered"] = req.PositionOffered
	}
	if req.EntryDate != nil {
		fields["entry_date"] = req.EntryDate
	}

	// 若状态为"已参加"，反馈后同步置为"已完成"
	if i.Status == model.InterviewStatusAttended {
		fields["status"] = model.InterviewStatusCompleted
		fields["completed_at"] = &now
	}

	if err := s.repo.Update(id, fields); err != nil {
		return nil, err
	}

	// 同步投递记录状态
	switch req.Result {
	case model.InterviewResultPass, model.InterviewResultNextRound:
		// 通过/进入下一轮 → 投递状态保持"面试中"
		_ = s.applicationRepo.Update(i.ApplicationID, map[string]interface{}{
			"status": model.ApplicationStatusInterviewing,
		})
	case model.InterviewResultReject:
		// 不通过 → 投递状态置为"不合适"
		_ = s.applicationRepo.Update(i.ApplicationID, map[string]interface{}{
			"status":         model.ApplicationStatusUnsuitable,
			"rejected_at":    &now,
			"rejected_reason": "面试未通过",
		})
	case model.InterviewResultOffer:
		// 发放 Offer → 投递状态置为"已发Offer"
		_ = s.applicationRepo.Update(i.ApplicationID, map[string]interface{}{
			"status":  model.ApplicationStatusOffer,
			"offer_at": &now,
		})
		if req.SalaryOffered > 0 {
			_ = s.applicationRepo.SetOfferInfo(i.ApplicationID, req.SalaryOffered, &now)
		}
	}

	updated, _ := s.repo.FindByID(id)
	return toInterviewResponse(updated), nil
}

// Stats 面试统计
func (s *interviewService) Stats(userID uint, role string) (*dto.InterviewStatsResponse, error) {
	_, groups, err := s.repo.Stats(userID, role)
	if err != nil {
		return nil, err
	}
	resp := &dto.InterviewStatsResponse{
		PassRate: 0,
	}
	for status, count := range groups {
		resp.TotalInterviews += count
		switch status {
		case model.InterviewStatusPending:
			resp.PendingCount = count
		case model.InterviewStatusConfirmed, model.InterviewStatusRescheduled:
			resp.ConfirmedCount += count
		case model.InterviewStatusCompleted:
			resp.CompletedCount = count
		case model.InterviewStatusCanceled:
			resp.CanceledCount = count
		case model.InterviewStatusNoShow:
			resp.NoShowCount = count
		}
	}
	// Offer 数 = result=offer 的面试数（通过 Feedback 写入 application.offer_at 间接体现）
	// 简化：通过 application 状态为 Offer/Onboarded 的统计
	if resp.CompletedCount > 0 {
		resp.PassRate = float64(resp.CompletedCount) / float64(resp.TotalInterviews) * 100
	}
	return resp, nil
}
