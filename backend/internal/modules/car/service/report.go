// Package service 同城车辆买卖业务逻辑层 - 举报 + 评价
// CarReport 举报（全局工单，BaseModel 无 region_id）
// CarReview 评价（RegionBaseModel，含 region_id）
package service

import (
	"errors"
	"math/rand"
	"time"

	"wuchang-tongcheng/internal/modules/car/dto"
	"wuchang-tongcheng/internal/modules/car/model"
	"wuchang-tongcheng/internal/modules/car/repository"
	"wuchang-tongcheng/internal/pkg/utils"
)

var (
	ErrReportNotFound     = errors.New("举报工单不存在")
	ErrReportStatusInvalid = errors.New("举报状态不允许此操作")
	ErrReportNoPermission = errors.New("无权操作此举报工单")

	ErrReviewNotFound     = errors.New("评价不存在")
	ErrReviewNoPermission = errors.New("无权操作此评价")
	ErrReviewAlreadyExist = errors.New("已评价过该目标")
	ErrReviewStatusInvalid = errors.New("评价状态不允许此操作")
)

// ===== ReportService 举报业务接口 =====

// ReportService 举报业务接口（举报为全局工单，无 region_id 隔离）
type ReportService interface {
	// C 端
	Create(reporterID uint, reporterName string, req *dto.CreateReportRequest) (*dto.ReportInfo, error)
	GetByID(id uint) (*dto.ReportInfo, error)
	List(req *dto.ReportListRequest) (*utils.Pagination, []dto.ReportInfo, error)
	ListByReporter(reporterID uint, page, pageSize int) (*utils.Pagination, []dto.ReportInfo, error)
	ListByReportedUser(reportedUserID uint, page, pageSize int) (*utils.Pagination, []dto.ReportInfo, error)
	ListByTarget(targetType string, targetID uint) ([]dto.ReportInfo, error)
	// 申诉
	Appeal(id uint, reporterID uint, req *dto.AppealReportRequest) error

	// M 端
	AdminGetByID(id uint) (*dto.ReportInfo, error)
	Process(id uint, handlerID uint, handlerName string, req *dto.ProcessReportRequest) error
	ProcessAppeal(id uint, handlerID uint, req *dto.ProcessAppealRequest) error
	UpdateStatus(id uint, status int) error
}

type reportService struct {
	repo repository.ReportRepository
}

// NewReportService 创建举报 service 实例
func NewReportService(repo repository.ReportRepository) ReportService {
	return &reportService{repo: repo}
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
	case model.ReportStatusCanceled:
		return "已取消"
	}
	return ""
}

// genReportNo 生成举报单号：RP + yyyyMMddHHmmss + 6 位随机
func genReportNo() string {
	return "RP" + time.Now().Format("20060102150405") + randomDigits(6)
}

// randomDigits 生成 n 位随机数字串
func randomDigits(n int) string {
	if n <= 0 {
		return ""
	}
	const digits = "0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = digits[rand.Intn(len(digits))]
	}
	return string(b)
}

// toReportInfo model -> dto
func toReportInfo(r *model.CarReport) *dto.ReportInfo {
	info := &dto.ReportInfo{
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
	}
	if r.EvidenceImages != nil {
		info.EvidenceImages = r.EvidenceImages
	}
	return info
}

// Create 创建举报（C 端用户发起）
func (s *reportService) Create(reporterID uint, reporterName string, req *dto.CreateReportRequest) (*dto.ReportInfo, error) {
	// SLA 默认 48 小时
	sla := time.Now().Add(48 * time.Hour)
	rep := &model.CarReport{
		ReportNo:         genReportNo(),
		TargetType:       req.TargetType,
		TargetID:         req.TargetID,
		TargetUserID:     req.TargetUserID,
		ReporterID:       reporterID,
		ReporterName:     reporterName,
		ReportedUserID:   req.TargetUserID,
		ReportedUserName: "",
		ReportType:       req.ReportType,
		Reason:           req.Reason,
		Description:      req.Description,
		Status:           model.ReportStatusPending,
		SLADeadline:      &sla,
	}
	// 证据图片 JSONB
	if req.EvidenceImages != nil {
		if jb, err := model.FromJSON(req.EvidenceImages); err == nil {
			rep.EvidenceImages = jb
		}
	}
	if err := s.repo.Create(rep); err != nil {
		return nil, err
	}
	return toReportInfo(rep), nil
}

// GetByID 查询举报详情
func (s *reportService) GetByID(id uint) (*dto.ReportInfo, error) {
	rep, err := s.repo.FindByID(id)
	if err != nil {
		return nil, ErrReportNotFound
	}
	return toReportInfo(rep), nil
}

// List 举报列表（M 端）
func (s *reportService) List(req *dto.ReportListRequest) (*utils.Pagination, []dto.ReportInfo, error) {
	pagination := utils.NewPagination(int(req.Page), int(req.PageSize))
	query := repository.ReportListQuery{
		TargetType: req.TargetType,
		TargetID:   req.TargetID,
		ReporterID: req.ReporterID,
		ReportType: req.ReportType,
		Status:     req.Status,
		Keyword:    req.Keyword,
	}
	list, total, err := s.repo.List(query, pagination)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	infos := make([]dto.ReportInfo, 0, len(list))
	for i := range list {
		infos = append(infos, *toReportInfo(&list[i]))
	}
	return pagination, infos, nil
}

// ListByReporter 按举报人反查
func (s *reportService) ListByReporter(reporterID uint, page, pageSize int) (*utils.Pagination, []dto.ReportInfo, error) {
	pagination := utils.NewPagination(page, pageSize)
	list, total, err := s.repo.ListByReporter(reporterID, pagination)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	infos := make([]dto.ReportInfo, 0, len(list))
	for i := range list {
		infos = append(infos, *toReportInfo(&list[i]))
	}
	return pagination, infos, nil
}

// ListByReportedUser 按被举报人反查
func (s *reportService) ListByReportedUser(reportedUserID uint, page, pageSize int) (*utils.Pagination, []dto.ReportInfo, error) {
	pagination := utils.NewPagination(page, pageSize)
	list, total, err := s.repo.ListByReportedUser(reportedUserID, pagination)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	infos := make([]dto.ReportInfo, 0, len(list))
	for i := range list {
		infos = append(infos, *toReportInfo(&list[i]))
	}
	return pagination, infos, nil
}

// ListByTarget 按目标反查
func (s *reportService) ListByTarget(targetType string, targetID uint) ([]dto.ReportInfo, error) {
	list, err := s.repo.ListByTarget(targetType, targetID)
	if err != nil {
		return nil, err
	}
	infos := make([]dto.ReportInfo, 0, len(list))
	for i := range list {
		infos = append(infos, *toReportInfo(&list[i]))
	}
	return infos, nil
}

// Appeal 用户对处理结果申诉
func (s *reportService) Appeal(id uint, reporterID uint, req *dto.AppealReportRequest) error {
	rep, err := s.repo.FindByID(id)
	if err != nil {
		return ErrReportNotFound
	}
	if rep.ReporterID != reporterID {
		return ErrReportNoPermission
	}
	// 仅已处理/已驳回的工单可以申诉
	if rep.Status != model.ReportStatusHandled && rep.Status != model.ReportStatusRejected {
		return ErrReportStatusInvalid
	}
	now := time.Now()
	return s.repo.Update(id, map[string]interface{}{
		"appeal_reason": req.AppealReason,
		"appealed_at":   &now,
		"status":        model.ReportStatusProcessing,
	})
}

// AdminGetByID M 端查询
func (s *reportService) AdminGetByID(id uint) (*dto.ReportInfo, error) {
	return s.GetByID(id)
}

// Process M 端处理举报
func (s *reportService) Process(id uint, handlerID uint, handlerName string, req *dto.ProcessReportRequest) error {
	rep, err := s.repo.FindByID(id)
	if err != nil {
		return ErrReportNotFound
	}
	// 仅待处理/处理中的工单可以处理
	if rep.Status != model.ReportStatusPending && rep.Status != model.ReportStatusProcessing {
		return ErrReportStatusInvalid
	}
	now := time.Now()
	penaltyUserID := req.PenaltyUserID
	if penaltyUserID == 0 {
		penaltyUserID = rep.ReportedUserID
	}
	fields := map[string]interface{}{
		"status":          req.Status,
		"handler_id":      handlerID,
		"handler_name":    handlerName,
		"handle_result":   req.HandleResult,
		"penalty_type":    req.PenaltyType,
		"penalty_user_id": penaltyUserID,
		"handled_at":      &now,
	}
	return s.repo.Update(id, fields)
}

// ProcessAppeal M 端处理申诉
func (s *reportService) ProcessAppeal(id uint, handlerID uint, req *dto.ProcessAppealRequest) error {
	rep, err := s.repo.FindByID(id)
	if err != nil {
		return ErrReportNotFound
	}
	if rep.AppealedAt == nil {
		return ErrReportStatusInvalid
	}
	now := time.Now()
	return s.repo.Update(id, map[string]interface{}{
		"appeal_result":     req.AppealResult,
		"appeal_handler_id": handlerID,
		"appeal_handled_at": &now,
		"status":            model.ReportStatusHandled,
	})
}

// UpdateStatus M 端更新状态
func (s *reportService) UpdateStatus(id uint, status int) error {
	if _, err := s.repo.FindByID(id); err != nil {
		return ErrReportNotFound
	}
	return s.repo.Update(id, map[string]interface{}{"status": status})
}

// ===== ReviewService 评价业务接口 =====

// ReviewService 评价业务接口（评价含 region_id 隔离）
type ReviewService interface {
	// C 端
	Create(regionID uint, userID uint, userName string, userAvatar string, req *dto.CreateReviewRequest) (*dto.ReviewInfo, error)
	Update(id uint, operatorID uint, req *dto.UpdateReviewRequest) error
	Delete(id uint, operatorID uint) error
	GetByID(id uint) (*dto.ReviewInfo, error)
	List(regionID uint, req *dto.ReviewListRequest) (*utils.Pagination, []dto.ReviewInfo, error)
	ListByTarget(regionID uint, targetType string, targetID uint, page, pageSize int) (*utils.Pagination, []dto.ReviewInfo, error)
	ListByReviewer(reviewerID uint, page, pageSize int) (*utils.Pagination, []dto.ReviewInfo, error)
	Reply(id uint, operatorID uint, req *dto.ReviewReplyRequest) error
	Append(id uint, operatorID uint, req *dto.ReviewAppendRequest) error
	Stats(targetType string, targetID uint) (*dto.ReviewStatsResponse, error)
	Like(id uint) error

	// M 端
	AdminList(req *dto.ReviewListRequest) (*utils.Pagination, []dto.ReviewInfo, error)
	UpdateStatus(id uint, status int) error
}

type reviewService struct {
	repo repository.ReviewRepository
}

// NewReviewService 创建评价 service 实例
func NewReviewService(repo repository.ReviewRepository) ReviewService {
	return &reviewService{repo: repo}
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
	return ""
}

// toReviewInfo model -> dto
func toReviewInfo(r *model.CarReview) *dto.ReviewInfo {
	info := &dto.ReviewInfo{
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
		ExteriorRating:    r.ExteriorRating,
		InteriorRating:    r.InteriorRating,
		EngineRating:      r.EngineRating,
		PaperworkRating:   r.PaperworkRating,
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
	}
	if r.Images != nil {
		info.Images = r.Images
	}
	if r.Tags != nil {
		info.Tags = r.Tags
	}
	if r.AppendImages != nil {
		info.AppendImages = r.AppendImages
	}
	return info
}

// Create 创建评价
func (s *reviewService) Create(regionID uint, userID uint, userName string, userAvatar string, req *dto.CreateReviewRequest) (*dto.ReviewInfo, error) {
	// 同一目标只能评价一次
	has, err := s.repo.HasReviewed(userID, req.TargetID, req.TargetType)
	if err != nil {
		return nil, err
	}
	if has {
		return nil, ErrReviewAlreadyExist
	}
	targetType := req.TargetType
	if targetType == "" {
		targetType = model.ReviewTargetTypeDealer
	}
	reviewType := req.ReviewType
	if reviewType == "" {
		reviewType = model.ReviewerTypeBuyer
	}
	rv := &model.CarReview{
		TargetType:        targetType,
		TargetID:          req.TargetID,
		ReviewerID:        userID,
		ReviewerName:      userName,
		ReviewerAvatar:    userAvatar,
		ReviewType:        reviewType,
		Rating:            req.Rating,
		Content:           req.Content,
		VideoURL:          req.VideoURL,
		IsAnonymous:       req.IsAnonymous,
		IsRecommended:     req.IsRecommended,
		DealAmount:        req.DealAmount,
		ExteriorRating:    req.ExteriorRating,
		InteriorRating:    req.InteriorRating,
		EngineRating:      req.EngineRating,
		PaperworkRating:   req.PaperworkRating,
		ServiceAttitude:   req.ServiceAttitude,
		ProfessionalSkill: req.ProfessionalSkill,
		Status:            model.ReviewStatusApproved, // MVP 简化：默认通过
	}
	rv.RegionID = regionID
	if req.Images != nil {
		if jb, err := model.FromJSON(req.Images); err == nil {
			rv.Images = jb
		}
	}
	if req.Tags != nil {
		if jb, err := model.FromJSON(req.Tags); err == nil {
			rv.Tags = jb
		}
	}
	if err := s.repo.Create(rv); err != nil {
		return nil, err
	}
	return toReviewInfo(rv), nil
}

// Update 更新评价
func (s *reviewService) Update(id uint, operatorID uint, req *dto.UpdateReviewRequest) error {
	rv, err := s.repo.FindByID(id)
	if err != nil {
		return ErrReviewNotFound
	}
	if rv.ReviewerID != operatorID {
		return ErrReviewNoPermission
	}
	fields := map[string]interface{}{}
	if req.Content != nil {
		fields["content"] = *req.Content
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
	if req.Status != nil {
		fields["status"] = *req.Status
	}
	if len(fields) == 0 {
		return nil
	}
	return s.repo.Update(id, fields)
}

// Delete 删除评价
func (s *reviewService) Delete(id uint, operatorID uint) error {
	rv, err := s.repo.FindByID(id)
	if err != nil {
		return ErrReviewNotFound
	}
	// operatorID == 0 表示 M 端强制操作
	if operatorID != 0 && rv.ReviewerID != operatorID {
		return ErrReviewNoPermission
	}
	return s.repo.Delete(id)
}

// GetByID 查询评价详情
func (s *reviewService) GetByID(id uint) (*dto.ReviewInfo, error) {
	rv, err := s.repo.FindByID(id)
	if err != nil {
		return nil, ErrReviewNotFound
	}
	return toReviewInfo(rv), nil
}

// List 评价列表
func (s *reviewService) List(regionID uint, req *dto.ReviewListRequest) (*utils.Pagination, []dto.ReviewInfo, error) {
	pagination := utils.NewPagination(int(req.Page), int(req.PageSize))
	query := repository.ReviewListQuery{
		TargetType: req.TargetType,
		TargetID:   req.TargetID,
		ReviewerID: req.ReviewerID,
		ReviewType: req.ReviewType,
		Rating:     req.Rating,
		Status:     req.Status,
		HasReply:   req.HasReply,
		Keyword:    req.Keyword,
	}
	list, total, err := s.repo.List(regionID, query, pagination)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	infos := make([]dto.ReviewInfo, 0, len(list))
	for i := range list {
		infos = append(infos, *toReviewInfo(&list[i]))
	}
	return pagination, infos, nil
}

// ListByTarget 按目标反查
func (s *reviewService) ListByTarget(regionID uint, targetType string, targetID uint, page, pageSize int) (*utils.Pagination, []dto.ReviewInfo, error) {
	pagination := utils.NewPagination(page, pageSize)
	list, total, err := s.repo.ListByTarget(regionID, targetType, targetID, pagination)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	infos := make([]dto.ReviewInfo, 0, len(list))
	for i := range list {
		infos = append(infos, *toReviewInfo(&list[i]))
	}
	return pagination, infos, nil
}

// ListByReviewer 按评价人反查
func (s *reviewService) ListByReviewer(reviewerID uint, page, pageSize int) (*utils.Pagination, []dto.ReviewInfo, error) {
	pagination := utils.NewPagination(page, pageSize)
	list, total, err := s.repo.ListByReviewer(reviewerID, pagination)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	infos := make([]dto.ReviewInfo, 0, len(list))
	for i := range list {
		infos = append(infos, *toReviewInfo(&list[i]))
	}
	return pagination, infos, nil
}

// Reply 评价回复（卖家/车商回复评价）
func (s *reviewService) Reply(id uint, operatorID uint, req *dto.ReviewReplyRequest) error {
	rv, err := s.repo.FindByID(id)
	if err != nil {
		return ErrReviewNotFound
	}
	_ = rv // operatorID 暂不强制校验（回复者身份由 handler 通过上下文判定）
	now := time.Now()
	return s.repo.Update(id, map[string]interface{}{
		"reply":    req.Reply,
		"reply_at": &now,
	})
}

// Append 评价追评
func (s *reviewService) Append(id uint, operatorID uint, req *dto.ReviewAppendRequest) error {
	rv, err := s.repo.FindByID(id)
	if err != nil {
		return ErrReviewNotFound
	}
	if rv.ReviewerID != operatorID {
		return ErrReviewNoPermission
	}
	now := time.Now()
	fields := map[string]interface{}{
		"append_content": req.AppendContent,
		"append_at":      &now,
	}
	if req.AppendImages != nil {
		if jb, err := model.FromJSON(req.AppendImages); err == nil {
			fields["append_images"] = jb
		}
	}
	return s.repo.Update(id, fields)
}

// Stats 评价统计
func (s *reviewService) Stats(targetType string, targetID uint) (*dto.ReviewStatsResponse, error) {
	total, avg, good, medium, bad, err := s.repo.StatsByTarget(targetType, targetID)
	if err != nil {
		return nil, err
	}
	resp := &dto.ReviewStatsResponse{
		TotalReviews: int(total),
		AvgRating:    avg,
	}
	if total > 0 {
		resp.GoodRate = float64(good) / float64(total)
		resp.MediumRate = float64(medium) / float64(total)
		resp.BadRate = float64(bad) / float64(total)
		// 回复率：需另行查询，MVP 简化为 0
		resp.HasReplyRate = 0
	}
	return resp, nil
}

// Like 点赞评价
func (s *reviewService) Like(id uint) error {
	if _, err := s.repo.FindByID(id); err != nil {
		return ErrReviewNotFound
	}
	return s.repo.IncrLikeCount(id)
}

// AdminList M 端列表
func (s *reviewService) AdminList(req *dto.ReviewListRequest) (*utils.Pagination, []dto.ReviewInfo, error) {
	// M 端跨地区：regionID=0
	return s.List(0, req)
}

// UpdateStatus M 端更新状态
func (s *reviewService) UpdateStatus(id uint, status int) error {
	if _, err := s.repo.FindByID(id); err != nil {
		return ErrReviewNotFound
	}
	return s.repo.Update(id, map[string]interface{}{"status": status})
}
