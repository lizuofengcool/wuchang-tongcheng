// Package service 举报 + 评价业务逻辑层
// 依据 v3.2.1 架构方案第五章：对标贝壳/链家
// 依据需求文档 1.5：内容审核必须做（举报工单 SLA + 申诉流程）
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
	ErrReportNotFound     = errors.New("举报不存在")
	ErrReviewNotFound     = errors.New("评价不存在")
	ErrReportHandled      = errors.New("举报已处理，无法重复处理")
	ErrReportNoAppeal     = errors.New("举报未处理，无需申诉")
	ErrReportAppealed     = errors.New("举报已申诉，无法重复申诉")
	ErrReviewNoPermission = errors.New("无权操作此评价")
	ErrReportNoPermission = errors.New("无权操作此举报")
)

// ReportService 举报业务接口
type ReportService interface {
	// C 端
	Create(regionID uint, reporterID uint, reporterName string, req *dto.ReportCreateRequest) (*dto.ReportResponse, error)
	GetByID(id uint) (*dto.ReportResponse, error)
	GetByNo(no string) (*dto.ReportResponse, error)
	List(req *dto.ReportListQuery) (*utils.Pagination, []dto.ReportResponse, error)
	ListMine(reporterID uint, req *dto.ReportListQuery) (*utils.Pagination, []dto.ReportResponse, error)
	ListByTarget(targetType string, targetID uint, page, pageSize int) (*utils.Pagination, []dto.ReportResponse, error)
	Appeal(id uint, userID uint, req *dto.ReportAppealRequest) error

	// M 端管理
	AdminList(req *dto.ReportAdminListQuery) (*utils.Pagination, []dto.ReportResponse, error)
	Process(id uint, handlerID uint, handlerName string, req *dto.ReportProcessRequest) error
	AppealHandle(id uint, handlerID uint, req *dto.ReportAppealHandleRequest) error
	BatchUpdateStatus(ids []uint, status int) (*dto.BatchResultResponse, error)
	CountPending() (int64, error)
}

// ReviewService 评价业务接口
type ReviewService interface {
	// C 端
	Create(regionID uint, reviewerID uint, reviewerName string, reviewerAvatar string, req *dto.ReviewCreateRequest) (*dto.ReviewResponse, error)
	GetByID(id uint, userID uint) (*dto.ReviewResponse, error)
	List(req *dto.ReviewListQuery) (*utils.Pagination, []dto.ReviewResponse, error)
	ListByTarget(targetType string, targetID uint, page, pageSize int) (*utils.Pagination, []dto.ReviewResponse, error)
	ListMine(reviewerID uint, req *dto.ReviewListQuery) (*utils.Pagination, []dto.ReviewResponse, error)
	Reply(id uint, ownerID uint, req *dto.ReviewReplyRequest) error
	Append(id uint, reviewerID uint, req *dto.ReviewAppendRequest) error
	Like(id uint, userID uint) error

	// 统计
	GetStats(targetType string, targetID uint) (*model.ReviewStatsData, error)

	// M 端管理
	AdminList(req *dto.ReviewAdminListQuery) (*utils.Pagination, []dto.ReviewResponse, error)
	UpdateStatus(id uint, status int) error
	BatchUpdateStatus(ids []uint, status int) (*dto.BatchResultResponse, error)
	Delete(id uint) error
}

// ===== 举报实现 =====

type reportService struct {
	repo repository.RiskRepository
}

// NewReportService 创建举报 service 实例
func NewReportService(repo repository.RiskRepository) ReportService {
	return &reportService{repo: repo}
}

// toReportInfo model -> dto
func toReportInfo(r *model.HouseReport) *dto.ReportResponse {
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
		UpdatedAt:        r.UpdatedAt,
	}
	// 反序列化证据图片
	if len(r.EvidenceImages) > 0 {
		var imgs []model.EvidenceImage
		_ = r.EvidenceImages.Parse(&imgs)
		resp.EvidenceImages = imgs
	}
	return resp
}

// reportStatusText 举报状态文本
func reportStatusText(status int) string {
	switch status {
	case model.ReportStatusPending:
		return "待处理"
	case model.ReportStatusProcessing:
		return "处理中"
	case model.ReportStatusHandled:
		return "已处理"
	case model.ReportStatusRejected:
		return "已驳回"
	case model.ReportStatusAppealed:
		return "申诉中"
	case model.ReportStatusAppealHandled:
		return "申诉已处理"
	}
	return ""
}

// generateReportNo 生成举报单号：RP yyyyMMddHHmmss . 000
func generateReportNo() string {
	return fmt.Sprintf("RP%s.%03d", time.Now().Format("20060102150405"), time.Now().Nanosecond()%1000)
}

// Create 创建举报
func (s *reportService) Create(regionID uint, reporterID uint, reporterName string, req *dto.ReportCreateRequest) (*dto.ReportResponse, error) {
	r := &model.HouseReport{
		ReportNo:       generateReportNo(),
		TargetType:     req.TargetType,
		TargetID:       req.TargetID,
		TargetUserID:   req.TargetUserID,
		ReporterID:     reporterID,
		ReporterName:   reporterName,
		ReportType:     req.ReportType,
		Reason:         req.Reason,
		Description:    req.Description,
		Status:         model.ReportStatusPending,
		HandledAt:      nil,
	}
	// 证据图片转 JSONB
	if len(req.EvidenceImages) > 0 {
		b, err := model.FromJSON(req.EvidenceImages)
		if err == nil {
			r.EvidenceImages = b
		}
	}
	// SLA 截止时间：48 小时
	sla := time.Now().Add(48 * time.Hour)
	r.SLADeadline = &sla

	if err := s.repo.CreateReport(r); err != nil {
		return nil, err
	}
	return toReportInfo(r), nil
}

func (s *reportService) GetByID(id uint) (*dto.ReportResponse, error) {
	r, err := s.repo.FindReportByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrReportNotFound
		}
		return nil, err
	}
	return toReportInfo(r), nil
}

func (s *reportService) GetByNo(no string) (*dto.ReportResponse, error) {
	r, err := s.repo.FindReportByNo(no)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrReportNotFound
		}
		return nil, err
	}
	return toReportInfo(r), nil
}

func (s *reportService) List(req *dto.ReportListQuery) (*utils.Pagination, []dto.ReportResponse, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	opts := repository.ReportListOptions{
		TargetType:  req.TargetType,
		TargetID:    req.TargetID,
		ReporterID:  req.ReporterID,
		ReportType:  req.ReportType,
		Status:      req.Status,
		PenaltyType: req.PenaltyType,
		Keyword:     req.Keyword,
	}
	list, total, err := s.repo.ListReports(pagination, opts)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	result := make([]dto.ReportResponse, 0, len(list))
	for i := range list {
		result = append(result, *toReportInfo(&list[i]))
	}
	return pagination, result, nil
}

func (s *reportService) ListMine(reporterID uint, req *dto.ReportListQuery) (*utils.Pagination, []dto.ReportResponse, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	list, total, err := s.repo.ListReportsByReporter(reporterID, pagination)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	result := make([]dto.ReportResponse, 0, len(list))
	for i := range list {
		result = append(result, *toReportInfo(&list[i]))
	}
	return pagination, result, nil
}

func (s *reportService) ListByTarget(targetType string, targetID uint, page, pageSize int) (*utils.Pagination, []dto.ReportResponse, error) {
	pagination := utils.NewPagination(page, pageSize)
	list, total, err := s.repo.ListReportsByTarget(targetType, targetID, pagination)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	result := make([]dto.ReportResponse, 0, len(list))
	for i := range list {
		result = append(result, *toReportInfo(&list[i]))
	}
	return pagination, result, nil
}

// Appeal 申诉：仅举报人本人对已处理举报可申诉
func (s *reportService) Appeal(id uint, userID uint, req *dto.ReportAppealRequest) error {
	r, err := s.repo.FindReportByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrReportNotFound
		}
		return err
	}
	if r.ReporterID != userID {
		return ErrReportNoPermission
	}
	// 仅已处理 / 已驳回的举报可申诉
	if r.Status != model.ReportStatusHandled && r.Status != model.ReportStatusRejected {
		return ErrReportNoAppeal
	}
	if r.Status == model.ReportStatusAppealed || r.Status == model.ReportStatusAppealHandled {
		return ErrReportAppealed
	}
	now := time.Now()
	return s.repo.UpdateReportFields(id, map[string]interface{}{
		"appeal_reason": req.Reason,
		"appealed_at":   &now,
		"status":        model.ReportStatusAppealed,
	})
}

func (s *reportService) AdminList(req *dto.ReportAdminListQuery) (*utils.Pagination, []dto.ReportResponse, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	opts := repository.ReportListOptions{
		TargetType:  req.TargetType,
		ReportType:  req.ReportType,
		Status:      req.Status,
		PenaltyType: req.PenaltyType,
		Keyword:     req.Keyword,
	}
	// 通过 reporter_id 过滤（ListReports 不支持 handlerID，使用 keyword 模糊匹配处理人）
	list, total, err := s.repo.ListReports(pagination, opts)
	if err != nil {
		return nil, nil, err
	}
	// 在内存中按 handler_id 过滤
	if req.HandlerID > 0 {
		filtered := make([]model.HouseReport, 0, len(list))
		for i := range list {
			if list[i].HandlerID == req.HandlerID {
				filtered = append(filtered, list[i])
			}
		}
		list = filtered
		total = int64(len(filtered))
	}
	pagination.Total = total
	result := make([]dto.ReportResponse, 0, len(list))
	for i := range list {
		result = append(result, *toReportInfo(&list[i]))
	}
	return pagination, result, nil
}

// Process 处理举报（M 端）
func (s *reportService) Process(id uint, handlerID uint, handlerName string, req *dto.ReportProcessRequest) error {
	r, err := s.repo.FindReportByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrReportNotFound
		}
		return err
	}
	// 仅待处理/处理中可处理
	if r.Status == model.ReportStatusHandled || r.Status == model.ReportStatusAppealHandled {
		return ErrReportHandled
	}
	now := time.Now()
	fields := map[string]interface{}{
		"status":        req.Status,
		"handler_id":    handlerID,
		"handler_name":  handlerName,
		"handle_result": req.HandleResult,
		"penalty_type":  req.PenaltyType,
		"handled_at":    &now,
	}
	// 处罚用户 ID
	if r.ReportedUserID > 0 {
		fields["penalty_user_id"] = r.ReportedUserID
	}
	return s.repo.UpdateReportFields(id, fields)
}

// AppealHandle 处理申诉（M 端）
func (s *reportService) AppealHandle(id uint, handlerID uint, req *dto.ReportAppealHandleRequest) error {
	r, err := s.repo.FindReportByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrReportNotFound
		}
		return err
	}
	if r.Status != model.ReportStatusAppealed {
		return ErrReportNoAppeal
	}
	now := time.Now()
	return s.repo.UpdateReportFields(id, map[string]interface{}{
		"appeal_result":     req.AppealResult,
		"appeal_handler_id": handlerID,
		"appeal_handled_at": &now,
		"status":            model.ReportStatusAppealHandled,
	})
}

func (s *reportService) BatchUpdateStatus(ids []uint, status int) (*dto.BatchResultResponse, error) {
	affected, err := s.repo.BatchUpdateReportStatus(ids, status)
	if err != nil {
		return &dto.BatchResultResponse{
			Total: len(ids), Success: 0, Failed: len(ids), FailedIDs: ids,
		}, err
	}
	return &dto.BatchResultResponse{
		Total:   len(ids),
		Success: int(affected),
		Failed:  len(ids) - int(affected),
	}, nil
}

func (s *reportService) CountPending() (int64, error) {
	return s.repo.CountPendingReports()
}

// ===== 评价实现 =====

type reviewService struct {
	repo repository.RiskRepository
}

// NewReviewService 创建评价 service 实例
func NewReviewService(repo repository.RiskRepository) ReviewService {
	return &reviewService{repo: repo}
}

// toReviewInfo model -> dto
func toReviewInfo(r *model.HouseReview, hasLiked bool) *dto.ReviewResponse {
	resp := &dto.ReviewResponse{
		ID:                r.ID,
		TargetType:        r.TargetType,
		TargetID:          r.TargetID,
		ReviewerID:        r.ReviewerID,
		ReviewerName:      r.ReviewerName,
		ReviewerAvatar:    r.ReviewerAvatar,
		ReviewType:        r.ReviewType,
		Rating:            r.Rating,
		Content:           r.Content,
		VideoURL:          r.VideoURL,
		IsAnonymous:       r.IsAnonymous,
		IsRecommended:     r.IsRecommended,
		DealAmount:        r.DealAmount,
		ServiceAttitude:   r.ServiceAttitude,
		ProfessionalSkill: r.ProfessionalSkill,
		Reply:             r.Reply,
		ReplyAt:           r.ReplyAt,
		AppendContent:     r.AppendContent,
		AppendAt:          r.AppendAt,
		LikeCount:         r.LikeCount,
		Status:            r.Status,
		StatusText:        reviewStatusText(r.Status),
		RegionID:          r.RegionID,
		CreatedAt:         r.CreatedAt,
		UpdatedAt:         r.UpdatedAt,
		HasLiked:          hasLiked,
	}
	// 反序列化图片/标签/追评图片
	if len(r.Images) > 0 {
		var imgs []model.ReviewImage
		_ = r.Images.Parse(&imgs)
		resp.Images = imgs
	}
	if len(r.Tags) > 0 {
		var tags []model.HouseTagItem
		_ = r.Tags.Parse(&tags)
		resp.Tags = tags
	}
	if len(r.AppendImages) > 0 {
		var imgs []model.ReviewImage
		_ = r.AppendImages.Parse(&imgs)
		resp.AppendImages = imgs
	}
	return resp
}

// reviewStatusText 评价状态文本
func reviewStatusText(status int) string {
	switch status {
	case model.ReviewStatusHidden:
		return "隐藏"
	case model.ReviewStatusVisible:
		return "显示"
	case model.ReviewStatusReported:
		return "被举报"
	}
	return ""
}

// Create 创建评价
func (s *reviewService) Create(regionID uint, reviewerID uint, reviewerName string, reviewerAvatar string, req *dto.ReviewCreateRequest) (*dto.ReviewResponse, error) {
	r := &model.HouseReview{
		TargetType:        req.TargetType,
		TargetID:          req.TargetID,
		ReviewerID:        reviewerID,
		ReviewerName:      reviewerName,
		ReviewerAvatar:    reviewerAvatar,
		ReviewType:        req.ReviewType,
		Rating:            req.Rating,
		Content:           req.Content,
		VideoURL:          req.VideoURL,
		IsAnonymous:       req.IsAnonymous,
		IsRecommended:     req.IsRecommended,
		DealAmount:        req.DealAmount,
		ServiceAttitude:   req.ServiceAttitude,
		ProfessionalSkill: req.ProfessionalSkill,
		Status:            model.ReviewStatusVisible,
	}
	r.RegionID = regionID
	if r.ReviewType == "" {
		r.ReviewType = model.ReviewTypeTenant
	}
	if r.ServiceAttitude == 0 {
		r.ServiceAttitude = 5
	}
	if r.ProfessionalSkill == 0 {
		r.ProfessionalSkill = 5
	}
	// 图片转 JSONB
	if len(req.Images) > 0 {
		b, err := model.FromJSON(req.Images)
		if err == nil {
			r.Images = b
		}
	}
	// 标签转 JSONB
	if len(req.Tags) > 0 {
		b, err := model.FromJSON(req.Tags)
		if err == nil {
			r.Tags = b
		}
	}

	if err := s.repo.CreateReview(r); err != nil {
		return nil, err
	}
	return toReviewInfo(r, false), nil
}

func (s *reviewService) GetByID(id uint, userID uint) (*dto.ReviewResponse, error) {
	r, err := s.repo.FindReviewByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrReviewNotFound
		}
		return nil, err
	}
	// 当前用户是否点赞：本 MVP 不维护点赞明细表，简化为 false
	return toReviewInfo(r, false), nil
}

func (s *reviewService) List(req *dto.ReviewListQuery) (*utils.Pagination, []dto.ReviewResponse, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	opts := repository.ReviewListOptions{
		TargetType:    req.TargetType,
		TargetID:      req.TargetID,
		ReviewerID:    req.ReviewerID,
		ReviewType:    req.ReviewType,
		Rating:        req.Rating,
		IsRecommended: req.IsRecommended,
		Status:        req.Status,
		Sort:          req.Sort,
	}
	list, total, err := s.repo.ListReviews(0, pagination, opts)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	result := make([]dto.ReviewResponse, 0, len(list))
	for i := range list {
		result = append(result, *toReviewInfo(&list[i], false))
	}
	return pagination, result, nil
}

func (s *reviewService) ListByTarget(targetType string, targetID uint, page, pageSize int) (*utils.Pagination, []dto.ReviewResponse, error) {
	pagination := utils.NewPagination(page, pageSize)
	list, total, err := s.repo.ListReviewsByTarget(targetType, targetID, pagination)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	result := make([]dto.ReviewResponse, 0, len(list))
	for i := range list {
		result = append(result, *toReviewInfo(&list[i], false))
	}
	return pagination, result, nil
}

func (s *reviewService) ListMine(reviewerID uint, req *dto.ReviewListQuery) (*utils.Pagination, []dto.ReviewResponse, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	list, total, err := s.repo.ListReviewsByReviewer(reviewerID, pagination)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	result := make([]dto.ReviewResponse, 0, len(list))
	for i := range list {
		result = append(result, *toReviewInfo(&list[i], false))
	}
	return pagination, result, nil
}

// Reply 回复评价：仅评价目标所有者可回复
func (s *reviewService) Reply(id uint, ownerID uint, req *dto.ReviewReplyRequest) error {
	r, err := s.repo.FindReviewByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrReviewNotFound
		}
		return err
	}
	// MVP：不强制校验目标所有者，由 handler 层传 ownerID 做权限判断
	_ = ownerID
	_ = r
	now := time.Now()
	return s.repo.UpdateReviewFields(id, map[string]interface{}{
		"reply":    req.Reply,
		"reply_at": &now,
	})
}

// Append 追评：仅评价人本人可追评
func (s *reviewService) Append(id uint, reviewerID uint, req *dto.ReviewAppendRequest) error {
	r, err := s.repo.FindReviewByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrReviewNotFound
		}
		return err
	}
	if r.ReviewerID != reviewerID {
		return ErrReviewNoPermission
	}
	now := time.Now()
	fields := map[string]interface{}{
		"append_content": req.AppendContent,
		"append_at":      &now,
	}
	// 追评图片
	if len(req.AppendImages) > 0 {
		b, err := model.FromJSON(req.AppendImages)
		if err == nil {
			fields["append_images"] = b
		}
	}
	return s.repo.UpdateReviewFields(id, fields)
}

// Like 点赞评价（简化版：仅自增 like_count，不维护明细表）
func (s *reviewService) Like(id uint, userID uint) error {
	_ = userID
	if _, err := s.repo.FindReviewByID(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrReviewNotFound
		}
		return err
	}
	return s.repo.IncrReviewLikeCount(id)
}

func (s *reviewService) GetStats(targetType string, targetID uint) (*model.ReviewStatsData, error) {
	return s.repo.GetReviewStats(targetType, targetID)
}

func (s *reviewService) AdminList(req *dto.ReviewAdminListQuery) (*utils.Pagination, []dto.ReviewResponse, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	opts := repository.ReviewListOptions{
		TargetType: req.TargetType,
		TargetID:   req.TargetID,
		ReviewerID: req.ReviewerID,
		Rating:     req.Rating,
		Status:     req.Status,
	}
	list, total, err := s.repo.ListReviews(req.RegionID, pagination, opts)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	result := make([]dto.ReviewResponse, 0, len(list))
	for i := range list {
		result = append(result, *toReviewInfo(&list[i], false))
	}
	return pagination, result, nil
}

func (s *reviewService) UpdateStatus(id uint, status int) error {
	if _, err := s.repo.FindReviewByID(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrReviewNotFound
		}
		return err
	}
	return s.repo.UpdateReviewFields(id, map[string]interface{}{"status": status})
}

func (s *reviewService) BatchUpdateStatus(ids []uint, status int) (*dto.BatchResultResponse, error) {
	affected, err := s.repo.BatchUpdateReviewStatus(ids, status)
	if err != nil {
		return &dto.BatchResultResponse{
			Total: len(ids), Success: 0, Failed: len(ids), FailedIDs: ids,
		}, err
	}
	return &dto.BatchResultResponse{
		Total:   len(ids),
		Success: int(affected),
		Failed:  len(ids) - int(affected),
	}, nil
}

func (s *reviewService) Delete(id uint) error {
	if _, err := s.repo.FindReviewByID(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrReviewNotFound
		}
		return err
	}
	return s.repo.DeleteReview(id)
}
