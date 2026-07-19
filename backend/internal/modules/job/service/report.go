// Package service 举报 + 评价业务逻辑层
// 依据 v3.2.1 架构方案第四章：对标 BOSS直聘/看准
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
	ErrReportNotFound     = errors.New("举报不存在")
	ErrReportExists       = errors.New("已举报过此内容")
	ErrReportNoPermission = errors.New("无权操作此举报")
	ErrReportStatus       = errors.New("举报状态不允许此操作")

	ErrReviewNotFound     = errors.New("评价不存在")
	ErrReviewExists       = errors.New("已评价过此公司")
	ErrReviewNoPermission = errors.New("无权操作此评价")
	ErrReviewStatus       = errors.New("评价状态不允许此操作")
)

// ReportService 举报 + 评价业务接口
type ReportService interface {
	// 举报
	CreateReport(regionID uint, reporterID uint, reporterName string, req *dto.ReportCreateRequest) (*dto.ReportResponse, error)
	GetReport(id uint) (*dto.ReportResponse, error)
	ListReports(req *dto.ReportListQuery) (*utils.Pagination, []dto.ReportResponse, error)
	ListReportsByTarget(targetType string, targetID uint) ([]dto.ReportResponse, error)
	ProcessReport(id uint, handlerID uint, handlerName string, req *dto.ReportProcessRequest) (*dto.ReportResponse, error)
	AppealReport(id uint, reporterID uint, req *dto.ReportAppealRequest) (*dto.ReportResponse, error)
	ProcessAppeal(id uint, handlerID uint, req *dto.ReportAppealProcessRequest) (*dto.ReportResponse, error)

	// 评价
	CreateReview(regionID uint, reviewerID uint, reviewerName string, reviewerAvatar string, req *dto.ReviewCreateRequest) (*dto.ReviewResponse, error)
	UpdateReview(id uint, operatorID uint, req *dto.ReviewUpdateRequest) error
	GetReview(id uint) (*dto.ReviewResponse, error)
	ListReviews(req *dto.ReviewListQuery) (*utils.Pagination, []dto.ReviewResponse, error)
	ListReviewsByCompany(companyID uint, page, pageSize int) (*utils.Pagination, []dto.ReviewResponse, error)
	ReplyReview(id uint, operatorID uint, req *dto.ReviewReplyRequest) error
	AppendReview(id uint, operatorID uint, req *dto.ReviewAppendRequest) error
	DeleteReview(id uint, operatorID uint) error
	ReviewStats(companyID uint) (*dto.ReviewStatsResponse, error)
	LikeReview(id uint) error
	AuditReview(id uint, status int, reason string) error
}

type reportService struct {
	reportRepo repository.ReportRepository
	reviewRepo repository.ReviewRepository
}

// NewReportService 创建举报 + 评价 service 实例
func NewReportService(reportRepo repository.ReportRepository, reviewRepo repository.ReviewRepository) ReportService {
	return &reportService{reportRepo: reportRepo, reviewRepo: reviewRepo}
}

// genReportNo 生成举报单号：JR + yyyyMMddHHmmss + 6位随机数
func genReportNo() string {
	return fmt.Sprintf("JR%s%06d", time.Now().Format("20060102150405"), rand.Intn(1000000))
}

// reportStatusText 举报状态文本
func reportStatusText(status int) string {
	switch status {
	case model.ReportStatusPending:
		return "待处理"
	case model.ReportStatusProcessing:
		return "处理中"
	case model.ReportStatusResolved:
		return "已处理-成立"
	case model.ReportStatusRejected:
		return "已处理-不成立"
	case model.ReportStatusAppealed:
		return "申诉中"
	case model.ReportStatusClosed:
		return "已关闭"
	}
	return "未知"
}

// reviewStatusText 评价状态文本
func reviewStatusText(status int) string {
	switch status {
	case model.ReviewStatusPending:
		return "待审核"
	case model.ReviewStatusApproved:
		return "已通过"
	case model.ReviewStatusRejected:
		return "已拒绝"
	case model.ReviewStatusHidden:
		return "已隐藏"
	}
	return "未知"
}

// intPtr 工具：返回 int 指针
func intPtr(v int) *int {
	return &v
}

// toReportResponse model -> dto
func toReportResponse(r *model.JobReport) *dto.ReportResponse {
	resp := &dto.ReportResponse{
		ID:               r.ID,
		ReportNo:         r.ReportNo,
		TargetType:       r.TargetType,
		TargetID:         r.TargetID,
		TargetUserID:     r.TargetUserID,
		ReporterID:       r.ReporterID,
		ReporterName:     r.ReporterName,
		ReportedUserID:   r.ReportedUserID,
		ReportedUserName: r.ReportedUserName,
		ReportType:       r.ReportType,
		Reason:           r.Reason,
		Description:      r.Description,
		EvidenceImages:   []string{},
		Status:           r.Status,
		StatusText:       reportStatusText(r.Status),
		HandlerID:        r.HandlerID,
		HandlerName:      r.HandlerName,
		HandleResult:     r.HandleResult,
		PenaltyType:      r.PenaltyType,
		PenaltyUserID:    r.PenaltyUserID,
		SLADeadline:      r.SLADeadline,
		HandledAt:        r.HandledAt,
		AppealReason:     r.AppealReason,
		AppealedAt:       r.AppealedAt,
		AppealResult:     r.AppealResult,
		AppealHandlerID:  r.AppealHandlerID,
		AppealHandledAt:  r.AppealHandledAt,
		CreatedAt:        r.CreatedAt,
	}
	if r.EvidenceImages != nil {
		var arr []string
		_ = r.EvidenceImages.Parse(&arr)
		if arr != nil {
			resp.EvidenceImages = arr
		}
	}
	return resp
}

// toReviewResponse model -> dto
func toReviewResponse(r *model.JobReview) *dto.ReviewResponse {
	resp := &dto.ReviewResponse{
		ID:             r.ID,
		CompanyID:      r.CompanyID,
		ReviewerID:     r.ReviewerID,
		ReviewerName:   r.ReviewerName,
		ReviewerAvatar: r.ReviewerAvatar,
		ReviewType:     r.ReviewType,
		Rating:         r.Rating,
		Content:        r.Content,
		Images:         []string{},
		VideoURL:       r.VideoURL,
		IsAnonymous:    r.IsAnonymous,
		IsRecommended:  r.IsRecommended,
		Tags:           []string{},
		Position:       r.Position,
		Department:     r.Department,
		WorkDuration:   r.WorkDuration,
		SalaryRange:    r.SalaryRange,
		Pros:           r.Pros,
		Cons:           r.Cons,
		Advice:         r.Advice,
		Reply:          r.Reply,
		ReplyAt:        r.ReplyAt,
		AppendContent:  r.AppendContent,
		AppendImages:   []string{},
		AppendAt:       r.AppendAt,
		LikeCount:      r.LikeCount,
		Status:         r.Status,
		StatusText:     reviewStatusText(r.Status),
		RegionID:       r.RegionID,
		CreatedAt:      r.CreatedAt,
	}
	if r.Images != nil {
		var arr []string
		_ = r.Images.Parse(&arr)
		if arr != nil {
			resp.Images = arr
		}
	}
	if r.Tags != nil {
		var arr []string
		_ = r.Tags.Parse(&arr)
		if arr != nil {
			resp.Tags = arr
		}
	}
	if r.AppendImages != nil {
		var arr []string
		_ = r.AppendImages.Parse(&arr)
		if arr != nil {
			resp.AppendImages = arr
		}
	}
	// 匿名评价：隐藏评价人信息
	if r.IsAnonymous {
		resp.ReviewerName = "匿名用户"
		resp.ReviewerAvatar = ""
	}
	return resp
}

// ===== 举报 =====

func (s *reportService) CreateReport(regionID uint, reporterID uint, reporterName string, req *dto.ReportCreateRequest) (*dto.ReportResponse, error) {
	rep := &model.JobReport{
		ReportNo:        genReportNo(),
		TargetType:      req.TargetType,
		TargetID:        req.TargetID,
		TargetUserID:    req.TargetUserID,
		ReporterID:      reporterID,
		ReporterName:    reporterName,
		ReportedUserID:  req.TargetUserID,
		ReportedUserName: "",
		ReportType:      req.ReportType,
		Reason:          req.Reason,
		Description:     req.Description,
		Status:          model.ReportStatusPending,
	}
	// SLA：3 天内处理
	sla := time.Now().AddDate(0, 0, 3)
	rep.SLADeadline = &sla

	if len(req.EvidenceImages) > 0 {
		if jb, err := model.FromJSON(req.EvidenceImages); err == nil {
			rep.EvidenceImages = jb
		}
	}

	if err := s.reportRepo.Create(rep); err != nil {
		return nil, err
	}
	return toReportResponse(rep), nil
}

func (s *reportService) GetReport(id uint) (*dto.ReportResponse, error) {
	rep, err := s.reportRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrReportNotFound
		}
		return nil, err
	}
	return toReportResponse(rep), nil
}

func (s *reportService) ListReports(req *dto.ReportListQuery) (*utils.Pagination, []dto.ReportResponse, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	list, total, err := s.reportRepo.List(repository.ReportListQuery{
		Status:     req.Status,
		ReportType: req.ReportType,
		TargetType: req.TargetType,
		TargetID:   req.TargetID,
		ReporterID: req.ReporterID,
	}, pagination)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	result := make([]dto.ReportResponse, 0, len(list))
	for i := range list {
		result = append(result, *toReportResponse(&list[i]))
	}
	return pagination, result, nil
}

func (s *reportService) ListReportsByTarget(targetType string, targetID uint) ([]dto.ReportResponse, error) {
	list, err := s.reportRepo.ListByTarget(targetType, targetID)
	if err != nil {
		return nil, err
	}
	result := make([]dto.ReportResponse, 0, len(list))
	for i := range list {
		result = append(result, *toReportResponse(&list[i]))
	}
	return result, nil
}

func (s *reportService) ProcessReport(id uint, handlerID uint, handlerName string, req *dto.ReportProcessRequest) (*dto.ReportResponse, error) {
	rep, err := s.reportRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrReportNotFound
		}
		return nil, err
	}
	if rep.Status != model.ReportStatusPending && rep.Status != model.ReportStatusProcessing {
		return nil, ErrReportStatus
	}

	now := time.Now()
	fields := map[string]interface{}{
		"status":        req.Status,
		"handler_id":    handlerID,
		"handler_name":  handlerName,
		"handle_result": req.HandleResult,
		"penalty_type":  req.PenaltyType,
	}
	if req.Status == model.ReportStatusResolved || req.Status == model.ReportStatusRejected {
		fields["handled_at"] = &now
	}
	if req.Status == model.ReportStatusProcessing {
		// 处理中：不更新 handled_at
		delete(fields, "handled_at")
	}

	if err := s.reportRepo.Update(id, fields); err != nil {
		return nil, err
	}
	updated, _ := s.reportRepo.FindByID(id)
	return toReportResponse(updated), nil
}

func (s *reportService) AppealReport(id uint, reporterID uint, req *dto.ReportAppealRequest) (*dto.ReportResponse, error) {
	rep, err := s.reportRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrReportNotFound
		}
		return nil, err
	}
	if rep.ReporterID != reporterID {
		return nil, ErrReportNoPermission
	}
	// 仅已处理（成立/不成立）状态可申诉
	if rep.Status != model.ReportStatusResolved && rep.Status != model.ReportStatusRejected {
		return nil, ErrReportStatus
	}
	now := time.Now()
	fields := map[string]interface{}{
		"status":        model.ReportStatusAppealed,
		"appeal_reason": req.AppealReason,
		"appealed_at":   &now,
	}
	if err := s.reportRepo.Update(id, fields); err != nil {
		return nil, err
	}
	updated, _ := s.reportRepo.FindByID(id)
	return toReportResponse(updated), nil
}

func (s *reportService) ProcessAppeal(id uint, handlerID uint, req *dto.ReportAppealProcessRequest) (*dto.ReportResponse, error) {
	rep, err := s.reportRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrReportNotFound
		}
		return nil, err
	}
	if rep.Status != model.ReportStatusAppealed {
		return nil, ErrReportStatus
	}
	now := time.Now()
	fields := map[string]interface{}{
		"status":            model.ReportStatusClosed,
		"appeal_result":     req.AppealResult,
		"appeal_handler_id": handlerID,
		"appeal_handled_at": &now,
	}
	if err := s.reportRepo.Update(id, fields); err != nil {
		return nil, err
	}
	updated, _ := s.reportRepo.FindByID(id)
	return toReportResponse(updated), nil
}

// ===== 评价 =====

func (s *reportService) CreateReview(regionID uint, reviewerID uint, reviewerName string, reviewerAvatar string, req *dto.ReviewCreateRequest) (*dto.ReviewResponse, error) {
	// 检查是否已评价过该公司（每个用户对每家公司仅能评价一次）
	list, _, _ := s.reviewRepo.List(repository.ReviewListQuery{
		CompanyID:  req.CompanyID,
		ReviewerID: reviewerID,
		Status:     intPtr(model.ReviewStatusPending),
	}, utils.NewPagination(1, 1))
	if len(list) > 0 {
		// 状态为待审/通过/拒绝/隐藏的任一存在即视为已评价
	}
	// 简化：通过查询所有状态判断是否已评价
	for _, st := range []int{model.ReviewStatusPending, model.ReviewStatusApproved, model.ReviewStatusHidden} {
		existList, _, _ := s.reviewRepo.List(repository.ReviewListQuery{
			CompanyID:  req.CompanyID,
			ReviewerID: reviewerID,
			Status:     &st,
		}, utils.NewPagination(1, 1))
		if len(existList) > 0 {
			return nil, ErrReviewExists
		}
	}

	reviewType := req.ReviewType
	if reviewType == "" {
		reviewType = model.ReviewTypeEmployee
	}

	rv := &model.JobReview{
		CompanyID:     req.CompanyID,
		ReviewerID:    reviewerID,
		ReviewerName:  reviewerName,
		ReviewerAvatar: reviewerAvatar,
		ReviewType:    reviewType,
		Rating:        req.Rating,
		Content:       req.Content,
		VideoURL:      req.VideoURL,
		IsAnonymous:   req.IsAnonymous,
		IsRecommended: req.IsRecommended,
		Position:      req.Position,
		Department:    req.Department,
		WorkDuration:  req.WorkDuration,
		SalaryRange:   req.SalaryRange,
		Pros:          req.Pros,
		Cons:          req.Cons,
		Advice:        req.Advice,
		Status:        model.ReviewStatusApproved, // MVP 简化：直接通过
	}
	rv.RegionID = regionID

	if len(req.Images) > 0 {
		if jb, err := model.FromJSON(req.Images); err == nil {
			rv.Images = jb
		}
	}
	if len(req.Tags) > 0 {
		if jb, err := model.FromJSON(req.Tags); err == nil {
			rv.Tags = jb
		}
	}

	if err := s.reviewRepo.Create(rv); err != nil {
		return nil, err
	}
	return toReviewResponse(rv), nil
}

func (s *reportService) UpdateReview(id uint, operatorID uint, req *dto.ReviewUpdateRequest) error {
	rv, err := s.reviewRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrReviewNotFound
		}
		return err
	}
	if rv.ReviewerID != operatorID {
		return ErrReviewNoPermission
	}

	fields := make(map[string]interface{})
	if req.Rating != nil {
		fields["rating"] = *req.Rating
	}
	if req.Content != "" {
		fields["content"] = req.Content
	}
	if req.Images != nil {
		if jb, err := model.FromJSON(req.Images); err == nil {
			fields["images"] = jb
		}
	}
	if req.VideoURL != "" {
		fields["video_url"] = req.VideoURL
	}
	if req.IsAnonymous != nil {
		fields["is_anonymous"] = *req.IsAnonymous
	}
	if req.IsRecommended != nil {
		fields["is_recommended"] = *req.IsRecommended
	}
	if req.Tags != nil {
		if jb, err := model.FromJSON(req.Tags); err == nil {
			fields["tags"] = jb
		}
	}
	if req.Pros != "" {
		fields["pros"] = req.Pros
	}
	if req.Cons != "" {
		fields["cons"] = req.Cons
	}
	if req.Advice != "" {
		fields["advice"] = req.Advice
	}
	return s.reviewRepo.Update(id, fields)
}

func (s *reportService) GetReview(id uint) (*dto.ReviewResponse, error) {
	rv, err := s.reviewRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrReviewNotFound
		}
		return nil, err
	}
	return toReviewResponse(rv), nil
}

func (s *reportService) ListReviews(req *dto.ReviewListQuery) (*utils.Pagination, []dto.ReviewResponse, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	list, total, err := s.reviewRepo.List(repository.ReviewListQuery{
		CompanyID:  req.CompanyID,
		ReviewerID: req.ReviewerID,
		ReviewType: req.ReviewType,
		Rating:     req.Rating,
		Status:     req.Status,
	}, pagination)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	result := make([]dto.ReviewResponse, 0, len(list))
	for i := range list {
		result = append(result, *toReviewResponse(&list[i]))
	}
	return pagination, result, nil
}

func (s *reportService) ListReviewsByCompany(companyID uint, page, pageSize int) (*utils.Pagination, []dto.ReviewResponse, error) {
	pagination := utils.NewPagination(page, pageSize)
	list, total, err := s.reviewRepo.ListByCompanyID(companyID, pagination)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	result := make([]dto.ReviewResponse, 0, len(list))
	for i := range list {
		result = append(result, *toReviewResponse(&list[i]))
	}
	return pagination, result, nil
}

func (s *reportService) ReplyReview(id uint, operatorID uint, req *dto.ReviewReplyRequest) error {
	rv, err := s.reviewRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrReviewNotFound
		}
		return err
	}
	// TODO: 校验 operatorID 是否为公司所有者（此处简化）
	_ = operatorID
	if rv.Status != model.ReviewStatusApproved {
		return ErrReviewStatus
	}
	now := time.Now()
	return s.reviewRepo.Update(id, map[string]interface{}{
		"reply":    req.Reply,
		"reply_at": &now,
	})
}

func (s *reportService) AppendReview(id uint, operatorID uint, req *dto.ReviewAppendRequest) error {
	rv, err := s.reviewRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrReviewNotFound
		}
		return err
	}
	if rv.ReviewerID != operatorID {
		return ErrReviewNoPermission
	}
	if rv.AppendContent != "" {
		return errors.New("已追评，不可重复追评")
	}
	now := time.Now()
	fields := map[string]interface{}{
		"append_content": req.AppendContent,
		"append_at":      &now,
	}
	if len(req.AppendImages) > 0 {
		if jb, err := model.FromJSON(req.AppendImages); err == nil {
			fields["append_images"] = jb
		}
	}
	return s.reviewRepo.Update(id, fields)
}

func (s *reportService) DeleteReview(id uint, operatorID uint) error {
	rv, err := s.reviewRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrReviewNotFound
		}
		return err
	}
	if rv.ReviewerID != operatorID {
		return ErrReviewNoPermission
	}
	return s.reviewRepo.Delete(id)
}

func (s *reportService) ReviewStats(companyID uint) (*dto.ReviewStatsResponse, error) {
	total, avgRating, good, medium, bad, recommend, err := s.reviewRepo.StatsByCompanyID(companyID)
	if err != nil {
		return nil, err
	}
	resp := &dto.ReviewStatsResponse{
		CompanyID:      companyID,
		TotalReviews:   total,
		AvgRating:      avgRating,
		RecommendCount: recommend,
	}
	if total > 0 {
		resp.GoodRate = float64(good) / float64(total) * 100
		resp.MediumRate = float64(medium) / float64(total) * 100
		resp.BadRate = float64(bad) / float64(total) * 100
	}
	return resp, nil
}

// LikeReview 点赞评价（+1）
func (s *reportService) LikeReview(id uint) error {
	rv, err := s.reviewRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrReviewNotFound
		}
		return err
	}
	return s.reviewRepo.Update(id, map[string]interface{}{
		"like_count": rv.LikeCount + 1,
	})
}

// AuditReview M 端审核评价
func (s *reportService) AuditReview(id uint, status int, reason string) error {
	rv, err := s.reviewRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrReviewNotFound
		}
		return err
	}
	if rv.Status == status {
		return ErrReviewStatus
	}
	fields := map[string]interface{}{
		"status": status,
	}
	_ = reason // 暂不存储拒绝原因（可扩展）
	return s.reviewRepo.Update(id, fields)
}
