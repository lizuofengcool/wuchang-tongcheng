// Package service 投递记录业务逻辑层
// 依据 v3.2.1 架构方案第四章：对标 BOSS直聘 9 状态机投递
// 状态机：0已投递/1已读/2不合适/3面试邀约/4面试中/5Offer/6已入职/7已撤回/8已过期
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
	ErrApplicationNotFound     = errors.New("投递记录不存在")
	ErrApplicationNoPermission = errors.New("无权操作此投递记录")
	ErrApplicationExists       = errors.New("已投递过此职位")
	ErrApplicationStatus       = errors.New("投递状态不允许此操作")
	ErrApplicationAction       = errors.New("无效的操作类型")
)

// ApplicationService 投递记录业务接口
type ApplicationService interface {
	Create(regionID uint, userID uint, userName string, userAvatar string, req *dto.ApplicationCreateRequest) (*dto.ApplicationResponse, error)
	GetByID(id uint, userID uint) (*dto.ApplicationResponse, error)
	List(userID uint, req *dto.ApplicationListQuery) (*utils.Pagination, []dto.ApplicationResponse, error)
	ListByJobID(jobID uint, userID uint, page, pageSize int) (*utils.Pagination, []dto.ApplicationResponse, error)
	StatusUpdate(id uint, operatorID uint, req *dto.ApplicationStatusUpdateRequest) (*dto.ApplicationResponse, error)
	BatchAction(operatorID uint, req *dto.ApplicationBatchActionRequest) (*dto.BatchResultResponse, error)
	Stats(userID uint, role string) (map[int]int64, error)
}

type applicationService struct {
	repo     repository.ApplicationRepository
	jobRepo  repository.JobRepository
	resumeRepo repository.ResumeRepository
}

// NewApplicationService 创建投递记录 service 实例
func NewApplicationService(repo repository.ApplicationRepository, jobRepo repository.JobRepository, resumeRepo repository.ResumeRepository) ApplicationService {
	return &applicationService{repo: repo, jobRepo: jobRepo, resumeRepo: resumeRepo}
}

// genApplicationNo 生成投递单号：JA + yyyyMMddHHmmss + 6位随机数
func genApplicationNo() string {
	return fmt.Sprintf("JA%s%06d", time.Now().Format("20060102150405"), rand.Intn(1000000))
}

// applicationStatusText 投递状态文本
func applicationStatusText(status int) string {
	switch status {
	case model.ApplicationStatusDelivered:
		return "已投递"
	case model.ApplicationStatusRead:
		return "已读"
	case model.ApplicationStatusUnsuitable:
		return "不合适"
	case model.ApplicationStatusInterview:
		return "面试邀约"
	case model.ApplicationStatusInterviewing:
		return "面试中"
	case model.ApplicationStatusOffer:
		return "已发Offer"
	case model.ApplicationStatusOnboarded:
		return "已入职"
	case model.ApplicationStatusWithdrawn:
		return "已撤回"
	case model.ApplicationStatusExpired:
		return "已过期"
	}
	return "未知"
}

// toApplicationResponse model -> dto
func toApplicationResponse(a *model.JobApplication) *dto.ApplicationResponse {
	resp := &dto.ApplicationResponse{
		ID:              a.ID,
		ApplicationNo:   a.ApplicationNo,
		JobID:           a.JobID,
		ResumeID:        a.ResumeID,
		ApplicantID:     a.ApplicantID,
		RecruiterID:     a.RecruiterID,
		CompanyID:       a.CompanyID,
		PositionName:    a.PositionName,
		Status:          a.Status,
		StatusText:      applicationStatusText(a.Status),
		Source:          a.Source,
		CoverLetter:     a.CoverLetter,
		Attachments:     []map[string]interface{}{},
		ReadAt:          a.ReadAt,
		RepliedAt:       a.RepliedAt,
		InterviewCount:  a.InterviewCount,
		OfferAt:         a.OfferAt,
		OfferAmount:     a.OfferAmount,
		RejectedReason:  a.RejectedReason,
		RejectedAt:      a.RejectedAt,
		WithdrawnAt:     a.WithdrawnAt,
		WithdrawnReason: a.WithdrawnReason,
		CompletedAt:     a.CompletedAt,
		ExpiredAt:       a.ExpiredAt,
		SLADeadline:     a.SLADeadline,
		RegionID:        a.RegionID,
		CreatedAt:       a.CreatedAt,
		UpdatedAt:       a.UpdatedAt,
	}
	if a.Attachments != nil {
		var arr []map[string]interface{}
		_ = a.Attachments.Parse(&arr)
		if arr != nil {
			resp.Attachments = arr
		}
	}
	return resp
}

func (s *applicationService) Create(regionID uint, userID uint, userName string, userAvatar string, req *dto.ApplicationCreateRequest) (*dto.ApplicationResponse, error) {
	// 校验职位
	job, err := s.jobRepo.FindByID(req.JobID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrJobNotFound
		}
		return nil, err
	}
	// 不能投递自己的职位
	if job.UserID == userID {
		return nil, ErrApplicationExists
	}
	// 重复投递检查
	if count, _ := s.repo.CountByJobAndApplicant(req.JobID, userID); count > 0 {
		return nil, ErrApplicationExists
	}

	// 选择简历：优先使用 req.ResumeID，否则取默认简历
	resumeID := req.ResumeID
	if resumeID == 0 {
		if r, err := s.resumeRepo.GetDefault(userID); err == nil {
			resumeID = r.ID
		}
	}
	if resumeID == 0 {
		return nil, errors.New("请先选择简历")
	}

	source := req.Source
	if source == "" {
		source = model.ApplicationSourceProactive
	}

	a := &model.JobApplication{
		ApplicationNo: genApplicationNo(),
		JobID:         req.JobID,
		ResumeID:      resumeID,
		ApplicantID:   userID,
		RecruiterID:   job.UserID,
		CompanyID:     job.CompanyID,
		PositionName:  job.Title,
		Status:        model.ApplicationStatusDelivered,
		Source:        source,
		CoverLetter:   req.CoverLetter,
	}
	a.RegionID = regionID

	// JSONB 字段
	if len(req.Attachments) > 0 {
		if jb, err := model.FromJSON(req.Attachments); err == nil {
			a.Attachments = jb
		}
	}

	// 职位快照（简化版）
	if snapshot, err := model.FromJSON(map[string]interface{}{
		"job_id":          job.ID,
		"title":           job.Title,
		"salary_min":      job.SalaryMin,
		"salary_max":      job.SalaryMax,
		"salary_unit":     job.SalaryUnit,
		"work_city":       job.WorkCity,
		"recruitment_type": job.RecruitmentType,
	}); err == nil {
		a.PositionSnapshot = snapshot
	}

	if err := s.repo.Create(a); err != nil {
		return nil, err
	}
	// 职位投递数 +1，简历投递数 +1
	_ = s.jobRepo.IncrDeliverCount(req.JobID)
	_ = s.resumeRepo.IncrDeliverCount(resumeID)

	_ = userName
	_ = userAvatar
	return toApplicationResponse(a), nil
}

func (s *applicationService) GetByID(id uint, userID uint) (*dto.ApplicationResponse, error) {
	a, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrApplicationNotFound
		}
		return nil, err
	}
	// 仅投递人 / 招聘者可查
	if userID > 0 && a.ApplicantID != userID && a.RecruiterID != userID {
		return nil, ErrApplicationNoPermission
	}
	resp := toApplicationResponse(a)
	// 冗余展示：职位标题、公司名
	if job, err := s.jobRepo.FindByID(a.JobID); err == nil {
		resp.JobTitle = job.Title
	}
	return resp, nil
}

func (s *applicationService) List(userID uint, req *dto.ApplicationListQuery) (*utils.Pagination, []dto.ApplicationResponse, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	list, total, err := s.repo.List(repository.ApplicationListQuery{
		UserID:        userID,
		Role:          req.Role,
		Status:        req.Status,
		JobID:         req.JobID,
		CompanyID:     req.CompanyID,
		ApplicationNo: req.ApplicationNo,
		Keyword:       req.Keyword,
	}, pagination)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	result := make([]dto.ApplicationResponse, 0, len(list))
	for i := range list {
		resp := toApplicationResponse(&list[i])
		// 冗余职位标题
		if job, err := s.jobRepo.FindByID(list[i].JobID); err == nil {
			resp.JobTitle = job.Title
		}
		result = append(result, *resp)
	}
	return pagination, result, nil
}

func (s *applicationService) ListByJobID(jobID uint, userID uint, page, pageSize int) (*utils.Pagination, []dto.ApplicationResponse, error) {
	pagination := utils.NewPagination(page, pageSize)
	// 校验：仅该职位的招聘者可查
	if userID > 0 {
		job, err := s.jobRepo.FindByID(jobID)
		if err != nil {
			return nil, nil, ErrJobNotFound
		}
		if job.UserID != userID {
			return nil, nil, ErrApplicationNoPermission
		}
	}
	list, total, err := s.repo.ListByJobID(jobID, pagination)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	result := make([]dto.ApplicationResponse, 0, len(list))
	for i := range list {
		result = append(result, *toApplicationResponse(&list[i]))
	}
	return pagination, result, nil
}

// StatusUpdate 状态机变更
// action 映射：
//   - read: 已投递→已读（招聘者操作）
//   - unsuitable: 已读/面试邀约→不合适（招聘者操作）
//   - interview: 已读/已投递→面试邀约（招聘者发起）
//   - interviewing: 面试邀约→面试中
//   - offer: 面试中→已发Offer（招聘者发Offer）
//   - onboard: 已发Offer→已入职（招聘者/系统）
//   - withdraw: 已投递/已读/面试邀约→已撤回（求职者撤回）
//   - reactivate: 已撤回/不合适/已过期→已投递（重新激活）
func (s *applicationService) StatusUpdate(id uint, operatorID uint, req *dto.ApplicationStatusUpdateRequest) (*dto.ApplicationResponse, error) {
	a, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrApplicationNotFound
		}
		return nil, err
	}

	now := time.Now()
	fields := make(map[string]interface{})

	switch req.Action {
	case "read":
		// 仅招聘者可操作
		if operatorID > 0 && a.RecruiterID != operatorID {
			return nil, ErrApplicationNoPermission
		}
		if a.Status != model.ApplicationStatusDelivered {
			return nil, ErrApplicationStatus
		}
		fields["status"] = model.ApplicationStatusRead
		fields["read_at"] = &now

	case "unsuitable":
		if operatorID > 0 && a.RecruiterID != operatorID {
			return nil, ErrApplicationNoPermission
		}
		if a.Status != model.ApplicationStatusRead && a.Status != model.ApplicationStatusInterview && a.Status != model.ApplicationStatusInterviewing {
			return nil, ErrApplicationStatus
		}
		fields["status"] = model.ApplicationStatusUnsuitable
		fields["rejected_at"] = &now
		fields["rejected_reason"] = req.Reason
		fields["replied_at"] = &now

	case "interview":
		if operatorID > 0 && a.RecruiterID != operatorID {
			return nil, ErrApplicationNoPermission
		}
		if a.Status != model.ApplicationStatusDelivered && a.Status != model.ApplicationStatusRead {
			return nil, ErrApplicationStatus
		}
		fields["status"] = model.ApplicationStatusInterview
		fields["replied_at"] = &now

	case "interviewing":
		if operatorID > 0 && a.RecruiterID != operatorID && a.ApplicantID != operatorID {
			return nil, ErrApplicationNoPermission
		}
		if a.Status != model.ApplicationStatusInterview {
			return nil, ErrApplicationStatus
		}
		fields["status"] = model.ApplicationStatusInterviewing

	case "offer":
		if operatorID > 0 && a.RecruiterID != operatorID {
			return nil, ErrApplicationNoPermission
		}
		if a.Status != model.ApplicationStatusInterview && a.Status != model.ApplicationStatusInterviewing {
			return nil, ErrApplicationStatus
		}
		fields["status"] = model.ApplicationStatusOffer
		fields["offer_at"] = &now
		fields["replied_at"] = &now
		if req.OfferAmount > 0 {
			fields["offer_amount"] = req.OfferAmount
		}

	case "onboard":
		if operatorID > 0 && a.RecruiterID != operatorID {
			return nil, ErrApplicationNoPermission
		}
		if a.Status != model.ApplicationStatusOffer {
			return nil, ErrApplicationStatus
		}
		fields["status"] = model.ApplicationStatusOnboarded
		fields["completed_at"] = &now
		// 简历 Offer 数 +1
		_ = s.resumeRepo.IncrOfferCount(a.ResumeID)

	case "withdraw":
		if operatorID > 0 && a.ApplicantID != operatorID {
			return nil, ErrApplicationNoPermission
		}
		if a.Status != model.ApplicationStatusDelivered && a.Status != model.ApplicationStatusRead && a.Status != model.ApplicationStatusInterview {
			return nil, ErrApplicationStatus
		}
		fields["status"] = model.ApplicationStatusWithdrawn
		fields["withdrawn_at"] = &now
		fields["withdrawn_reason"] = req.Reason

	case "reactivate":
		if operatorID > 0 && a.ApplicantID != operatorID {
			return nil, ErrApplicationNoPermission
		}
		if a.Status != model.ApplicationStatusWithdrawn && a.Status != model.ApplicationStatusUnsuitable && a.Status != model.ApplicationStatusExpired {
			return nil, ErrApplicationStatus
		}
		fields["status"] = model.ApplicationStatusDelivered
		fields["withdrawn_at"] = nil
		fields["rejected_at"] = nil
		fields["expired_at"] = nil

	default:
		return nil, ErrApplicationAction
	}

	if err := s.repo.Update(id, fields); err != nil {
		return nil, err
	}
	updated, _ := s.repo.FindByID(id)
	return toApplicationResponse(updated), nil
}

// BatchAction 批量操作（仅招聘者可批量处理自己收到的投递）
func (s *applicationService) BatchAction(operatorID uint, req *dto.ApplicationBatchActionRequest) (*dto.BatchResultResponse, error) {
	result := &dto.BatchResultResponse{
		Total:     len(req.IDs),
		FailedIDs: []uint{},
	}
	statusMap := map[string]int{
		"read":       model.ApplicationStatusRead,
		"unsuitable": model.ApplicationStatusUnsuitable,
		"offer":      model.ApplicationStatusOffer,
		"onboard":    model.ApplicationStatusOnboarded,
	}
	status, ok := statusMap[req.Action]
	if !ok {
		return nil, ErrApplicationAction
	}

	now := time.Now()
	fields := map[string]interface{}{
		"status": status,
	}
	switch req.Action {
	case "read":
		fields["read_at"] = &now
	case "unsuitable":
		fields["rejected_at"] = &now
		fields["rejected_reason"] = req.Reason
		fields["replied_at"] = &now
	case "offer":
		fields["offer_at"] = &now
		fields["replied_at"] = &now
	case "onboard":
		fields["completed_at"] = &now
	}

	success := 0
	for _, id := range req.IDs {
		// 校验权限
		a, err := s.repo.FindByID(id)
		if err != nil || a.RecruiterID != operatorID {
			result.FailedIDs = append(result.FailedIDs, id)
			continue
		}
		if err := s.repo.Update(id, fields); err != nil {
			result.FailedIDs = append(result.FailedIDs, id)
			continue
		}
		success++
	}
	result.Success = success
	result.Failed = len(result.FailedIDs)
	return result, nil
}

// Stats 按状态分组统计
func (s *applicationService) Stats(userID uint, role string) (map[int]int64, error) {
	result := make(map[int]int64)
	statuses := []int{
		model.ApplicationStatusDelivered,
		model.ApplicationStatusRead,
		model.ApplicationStatusUnsuitable,
		model.ApplicationStatusInterview,
		model.ApplicationStatusInterviewing,
		model.ApplicationStatusOffer,
		model.ApplicationStatusOnboarded,
		model.ApplicationStatusWithdrawn,
		model.ApplicationStatusExpired,
	}
	for _, st := range statuses {
		count, err := s.repo.CountByStatus(userID, role, st)
		if err != nil {
			return nil, err
		}
		result[st] = count
	}
	return result, nil
}
