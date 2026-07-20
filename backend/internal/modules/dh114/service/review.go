// Package service 同城114业务逻辑层 - 评价 + 商家回复
// 依据 v3.2.1 架构方案：对标大众点评 5 星评价体系
package service

import (
	"errors"
	"fmt"
	"math/rand"
	"time"

	"wuchang-tongcheng/internal/modules/dh114/dto"
	"wuchang-tongcheng/internal/modules/dh114/model"
	"wuchang-tongcheng/internal/modules/dh114/repository"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

var (
	ErrReviewNotFound      = errors.New("评价不存在")
	ErrReviewNoPermission   = errors.New("无权操作此评价")
	ErrReviewAlreadyExists  = errors.New("已评价过此商户")
	ErrReviewReplyNotFound  = errors.New("回复不存在")
)

// ReviewService 评价业务接口
type ReviewService interface {
	// 评价 CRUD
	Create(regionID uint, userID uint, userName string, userPhone string, userAvatar string, req *dto.CreateReviewRequest) (*dto.ReviewInfo, error)
	Update(id uint, operatorID uint, req *dto.UpdateReviewRequest) error
	Delete(id uint, operatorID uint) error
	GetByID(id uint) (*dto.ReviewInfo, error)
	List(regionID uint, req *dto.ReviewListRequest) (*utils.Pagination, []dto.ReviewInfo, error)
	ListByDh114(regionID uint, dh114ID uint, page, pageSize int) (*utils.Pagination, []dto.ReviewInfo, error)
	ListByReviewer(reviewerID uint, page, pageSize int) (*utils.Pagination, []dto.ReviewInfo, error)

	// 评价统计
	StatsByDh114(dh114ID uint) (*dto.ReviewStats, error)
	StatsByReviewer(reviewerID uint) (int64, float64, error)

	// 商家回复
	Reply(reviewID uint, replierID uint, replierName string, replierAvatar string, req *dto.ReviewReplyRequest) error
	ListReplies(reviewID uint) ([]model.Dh114ReviewReply, error)
	DeleteReply(replyID uint, operatorID uint) error

	// 互动
	LikeReview(id uint) error
	IncrLikeCount(id uint) error

	// M 端管理
	AuditReview(id uint, status int, auditReason string) error
	BatchAuditReviews(req *dto.BatchAuditRequest) (*dto.BatchResultResponse, error)
}

type reviewService struct {
	repo         repository.ReviewRepository
	replyRepo    repository.ReviewReplyRepository
	dh114Repo    repository.Dh114Repository
}

// NewReviewService 创建评价 service 实例
func NewReviewService(
	repo repository.ReviewRepository,
	replyRepo repository.ReviewReplyRepository,
	dh114Repo repository.Dh114Repository,
) ReviewService {
	return &reviewService{
		repo:      repo,
		replyRepo: replyRepo,
		dh114Repo: dh114Repo,
	}
}

// reviewStatusText 评价状态文本
func reviewStatusText(s int) string {
	switch s {
	case model.ReviewStatusPending:
		return "待审"
	case model.ReviewStatusApproved:
		return "通过"
	case model.ReviewStatusRejected:
		return "拒绝"
	case model.ReviewStatusHidden:
		return "隐藏"
	}
	return ""
}

// toReviewInfo model -> dto
func toReviewInfo(r *model.Dh114Review) *dto.ReviewInfo {
	info := &dto.ReviewInfo{
		ID:                r.ID,
		ReviewNo:          r.ReviewNo,
		Dh114ID:           r.Dh114ID,
		BusinessID:        r.BusinessID,
		ReviewerID:        r.ReviewerID,
		ReviewerName:      r.ReviewerName,
		ReviewerAvatar:    r.ReviewerAvatar,
		Rating:            r.Rating,
		TasteRating:       r.TasteRating,
		ServiceRating:     r.ServiceRating,
		EnvironmentRating: r.EnvironmentRating,
		Content:           r.Content,
		VideoURL:          r.VideoURL,
		VideoCover:       r.VideoCover,
		Reply:             r.Reply,
		RepliedAt:         r.RepliedAt,
		HasReply:          r.HasReply,
		LikeCount:         r.LikeCount,
		Status:            r.Status,
		StatusText:        reviewStatusText(r.Status),
		AuditStatus:       r.AuditStatus,
		AuditReason:       r.AuditReason,
		OrderID:           r.OrderID,
		ConsumedAt:        r.ConsumedAt,
		ReviewType:        r.ReviewType,
		RegionID:          r.RegionID,
		CreatedAt:         r.CreatedAt,
		UpdatedAt:         r.UpdatedAt,
	}
	if r.Images != nil {
		info.Images = r.Images
	}
	if r.Tags != nil {
		info.Tags = r.Tags
	}
	return info
}

// generateReviewNo 生成评价单号
func generateReviewNo() string {
	return fmt.Sprintf("DH114RV%s%06d", time.Now().Format("20060102150405"), rand.Intn(1000000))
}

// Create 创建评价
func (s *reviewService) Create(regionID uint, userID uint, userName string, userPhone string, userAvatar string, req *dto.CreateReviewRequest) (*dto.ReviewInfo, error) {
	// 检查是否已评价过
	if exists, _ := s.repo.HasReviewed(userID, req.Dh114ID); exists {
		return nil, ErrReviewAlreadyExists
	}

	rating := req.Rating
	if rating < 1 {
		rating = 5
	}
	if rating > 5 {
		rating = 5
	}
	tasteRating := req.TasteRating
	if tasteRating == 0 {
		tasteRating = rating
	}
	serviceRating := req.ServiceRating
	if serviceRating == 0 {
		serviceRating = rating
	}
	environmentRating := req.EnvironmentRating
	if environmentRating == 0 {
		environmentRating = rating
	}
	reviewType := req.ReviewType
	if reviewType == "" {
		reviewType = "general"
	}

	rv := &model.Dh114Review{
		ReviewNo:          generateReviewNo(),
		Dh114ID:           req.Dh114ID,
		ReviewerID:        userID,
		ReviewerName:      userName,
		ReviewerAvatar:    userAvatar,
		ReviewerPhone:     userPhone,
		Rating:            rating,
		TasteRating:       tasteRating,
		ServiceRating:     serviceRating,
		EnvironmentRating: environmentRating,
		Content:           req.Content,
		VideoURL:          req.VideoURL,
		VideoCover:        req.VideoCover,
		OrderID:           req.OrderID,
		ConsumedAt:        req.ConsumedAt,
		ReviewType:        reviewType,
		Status:            model.ReviewStatusApproved, // MVP：评价即通过
		AuditStatus:       model.AuditApproved,
	}
	rv.RegionID = regionID
	if req.Images != nil {
		if b, err := model.FromJSON(req.Images); err == nil {
			rv.Images = b
		}
	}
	if req.Tags != nil {
		if b, err := model.FromJSON(req.Tags); err == nil {
			rv.Tags = b
		}
	}

	if err := s.repo.Create(rv); err != nil {
		return nil, err
	}

	// 更新商户评分和评价数（异步执行简单可靠）
	s.updateDh114Rating(req.Dh114ID)

	return toReviewInfo(rv), nil
}

// updateDh114Rating 更新商户评分（简化版：直接重算平均分）
func (s *reviewService) updateDh114Rating(dh114ID uint) {
	total, avg, _, _, _, err := s.repo.StatsByDh114(dh114ID)
	if err != nil || total == 0 {
		return
	}
	_ = s.dh114Repo.UpdateFields(dh114ID, map[string]interface{}{
		"rating":       avg,
		"review_count": total,
	})
}

// Update 更新评价
func (s *reviewService) Update(id uint, operatorID uint, req *dto.UpdateReviewRequest) error {
	rv, err := s.repo.FindByID(id)
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
	if req.TasteRating != nil {
		fields["taste_rating"] = *req.TasteRating
	}
	if req.ServiceRating != nil {
		fields["service_rating"] = *req.ServiceRating
	}
	if req.EnvironmentRating != nil {
		fields["environment_rating"] = *req.EnvironmentRating
	}
	if req.Content != nil {
		fields["content"] = *req.Content
	}
	if req.VideoURL != nil {
		fields["video_url"] = *req.VideoURL
	}
	if req.VideoCover != nil {
		fields["video_cover"] = *req.VideoCover
	}
	if req.Images != nil {
		if b, err := model.FromJSON(req.Images); err == nil {
			fields["images"] = b
		}
	}
	if req.Tags != nil {
		if b, err := model.FromJSON(req.Tags); err == nil {
			fields["tags"] = b
		}
	}

	if len(fields) == 0 {
		return nil
	}
	if err := s.repo.Update(id, fields); err != nil {
		return err
	}

	// 更新商户评分
	s.updateDh114Rating(rv.Dh114ID)
	return nil
}

// Delete 删除评价
func (s *reviewService) Delete(id uint, operatorID uint) error {
	rv, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrReviewNotFound
		}
		return err
	}
	if rv.ReviewerID != operatorID {
		return ErrReviewNoPermission
	}
	if err := s.repo.Delete(id); err != nil {
		return err
	}

	// 更新商户评分
	s.updateDh114Rating(rv.Dh114ID)
	return nil
}

// GetByID 获取评价详情
func (s *reviewService) GetByID(id uint) (*dto.ReviewInfo, error) {
	rv, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrReviewNotFound
		}
		return nil, err
	}
	return toReviewInfo(rv), nil
}

// List 评价列表
func (s *reviewService) List(regionID uint, req *dto.ReviewListRequest) (*utils.Pagination, []dto.ReviewInfo, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	query := repository.ReviewListQuery{
		Dh114ID:    req.Dh114ID,
		ReviewerID: req.ReviewerID,
		Rating:     req.Rating,
		Status:     req.Status,
		HasReply:   req.HasReply,
		Keyword:    req.Keyword,
		Sort:       req.Sort,
	}
	list, total, err := s.repo.List(regionID, query, pagination)
	if err != nil {
		return nil, nil, err
	}
	infos := make([]dto.ReviewInfo, 0, len(list))
	for i := range list {
		infos = append(infos, *toReviewInfo(&list[i]))
	}
	pagination.Total = total
	return pagination, infos, nil
}

// ListByDh114 按商户列出评价
func (s *reviewService) ListByDh114(regionID uint, dh114ID uint, page, pageSize int) (*utils.Pagination, []dto.ReviewInfo, error) {
	pagination := utils.NewPagination(page, pageSize)
	list, total, err := s.repo.ListByDh114(regionID, dh114ID, pagination)
	if err != nil {
		return nil, nil, err
	}
	infos := make([]dto.ReviewInfo, 0, len(list))
	for i := range list {
		infos = append(infos, *toReviewInfo(&list[i]))
	}
	pagination.Total = total
	return pagination, infos, nil
}

// ListByReviewer 按评价人列出评价
func (s *reviewService) ListByReviewer(reviewerID uint, page, pageSize int) (*utils.Pagination, []dto.ReviewInfo, error) {
	pagination := utils.NewPagination(page, pageSize)
	list, total, err := s.repo.ListByReviewer(reviewerID, pagination)
	if err != nil {
		return nil, nil, err
	}
	infos := make([]dto.ReviewInfo, 0, len(list))
	for i := range list {
		infos = append(infos, *toReviewInfo(&list[i]))
	}
	pagination.Total = total
	return pagination, infos, nil
}

// StatsByDh114 评价统计
func (s *reviewService) StatsByDh114(dh114ID uint) (*dto.ReviewStats, error) {
	total, avg, good, medium, bad, err := s.repo.StatsByDh114(dh114ID)
	if err != nil {
		return nil, err
	}
	stats := &dto.ReviewStats{
		TotalReviews: int(total),
		AvgRating:    avg,
	}
	if total > 0 {
		stats.GoodRate = float64(good) / float64(total)
		stats.MediumRate = float64(medium) / float64(total)
		stats.BadRate = float64(bad) / float64(total)
	}
	return stats, nil
}

// StatsByReviewer 评价人统计
func (s *reviewService) StatsByReviewer(reviewerID uint) (int64, float64, error) {
	return s.repo.StatsByReviewer(reviewerID)
}

// Reply 商家回复评价
func (s *reviewService) Reply(reviewID uint, replierID uint, replierName string, replierAvatar string, req *dto.ReviewReplyRequest) error {
	rv, err := s.repo.FindByID(reviewID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrReviewNotFound
		}
		return err
	}

	// 创建回复记录
	reply := &model.Dh114ReviewReply{
		ReviewID:      reviewID,
		Dh114ID:       rv.Dh114ID,
		ReplierID:     replierID,
		ReplierName:   replierName,
		ReplierAvatar: replierAvatar,
		ReplierType:   "merchant",
		Content:       req.Content,
		Status:        1,
	}
	if req.Images != nil {
		if b, err := model.FromJSON(req.Images); err == nil {
			reply.Images = b
		}
	}
	if err := s.replyRepo.Create(reply); err != nil {
		return err
	}

	// 同时更新评价表的回复冗余字段
	now := time.Now()
	return s.repo.UpdateReply(reviewID, req.Content, &now)
}

// ListReplies 列出评价的回复
func (s *reviewService) ListReplies(reviewID uint) ([]model.Dh114ReviewReply, error) {
	return s.replyRepo.ListByReview(reviewID)
}

// DeleteReply 删除回复
func (s *reviewService) DeleteReply(replyID uint, operatorID uint) error {
	reply, err := s.replyRepo.FindByID(replyID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrReviewReplyNotFound
		}
		return err
	}
	if reply.ReplierID != operatorID {
		return ErrReviewNoPermission
	}
	return s.replyRepo.Delete(replyID)
}

// LikeReview 点赞评价
func (s *reviewService) LikeReview(id uint) error {
	return s.repo.IncrLikeCount(id)
}

// IncrLikeCount 增加点赞数
func (s *reviewService) IncrLikeCount(id uint) error {
	return s.repo.IncrLikeCount(id)
}

// AuditReview 审核评价
func (s *reviewService) AuditReview(id uint, status int, auditReason string) error {
	rv, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrReviewNotFound
		}
		return err
	}
	fields := map[string]interface{}{
		"status":       status,
		"audit_reason": auditReason,
	}
	if err := s.repo.Update(id, fields); err != nil {
		return err
	}

	// 重新计算商户评分
	s.updateDh114Rating(rv.Dh114ID)
	return nil
}

// BatchAuditReviews 批量审核评价
func (s *reviewService) BatchAuditReviews(req *dto.BatchAuditRequest) (*dto.BatchResultResponse, error) {
	result := &dto.BatchResultResponse{Total: len(req.IDs)}
	failedIDs := make([]uint, 0)
	for _, id := range req.IDs {
		if err := s.AuditReview(id, req.AuditStatus, req.AuditReason); err != nil {
			failedIDs = append(failedIDs, id)
		} else {
			result.Success++
		}
	}
	result.Failed = len(failedIDs)
	result.FailedIDs = failedIDs
	return result, nil
}
