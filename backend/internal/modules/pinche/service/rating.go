// Package service 同城拼车出行业务逻辑层 - 评价
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
	ErrRatingNotFound      = errors.New("评价不存在")
	ErrRatingNoPermission  = errors.New("无权操作此评价")
	ErrRatingStatusInvalid = errors.New("评价状态不允许此操作")
	ErrRatingAlreadyExists = errors.New("已评价过该行程")
)

// RatingService 评价业务接口
type RatingService interface {
	// C 端
	Create(regionID uint, raterID uint, raterName, raterAvatar string, req *dto.CreateRatingRequest) (*dto.RatingInfo, error)
	Update(id uint, operatorID uint, req *dto.UpdateRatingRequest) error
	Delete(id uint, operatorID uint) error
	GetByID(id uint) (*dto.RatingInfo, error)
	ListByRater(raterID uint, page, pageSize int) (*utils.Pagination, []dto.RatingInfo, error)
	ListByRatee(rateeID uint, page, pageSize int) (*utils.Pagination, []dto.RatingInfo, error)
	ListByPinche(pincheID uint, page, pageSize int) (*utils.Pagination, []dto.RatingInfo, error)
	Reply(id uint, operatorID uint, req *dto.RatingReplyRequest) error
	Like(id uint) error
	Stats(rateeID uint) (*dto.RatingStatsResponse, error)

	// M 端
	AdminList(req *dto.RatingListRequest) (*utils.Pagination, []dto.RatingInfo, error)
	UpdateStatus(id uint, status int) error
}

type ratingService struct {
	repo repository.RatingRepository
}

// NewRatingService 创建评价 service 实例
func NewRatingService(repo repository.RatingRepository) RatingService {
	return &ratingService{repo: repo}
}

// ratingStatusText 状态文本
func ratingStatusText(status int) string {
	switch status {
	case model.RatingStatusPending:
		return "待审"
	case model.RatingStatusApproved:
		return "通过"
	case model.RatingStatusRejected:
		return "拒绝"
	case model.RatingStatusHidden:
		return "隐藏"
	}
	return ""
}

// toRatingInfo model -> dto
func toRatingInfo(r *model.PincheRating) *dto.RatingInfo {
	info := &dto.RatingInfo{
		ID:          r.ID,
		RegionID:    r.RegionID,
		PincheID:    r.PincheID,
		BookingID:   r.BookingID,
		TripID:      r.TripID,
		RaterID:     r.RaterID,
		RaterName:   r.RaterName,
		RaterAvatar: r.RaterAvatar,
		RateeID:     r.RateeID,
		RateeName:   r.RateeName,
		RateeAvatar: r.RateeAvatar,
		RatingType:  r.RatingType,
		Rating:      r.Rating,
		Content:     r.Content,
		IsAnonymous: r.IsAnonymous,
		Reply:       r.Reply,
		ReplyAt:     r.ReplyAt,
		LikeCount:   r.LikeCount,
		Status:      r.Status,
		StatusText:  ratingStatusText(r.Status),
		CreatedAt:   r.CreatedAt,
	}
	if r.Images != nil {
		info.Images = r.Images
	}
	if r.Tags != nil {
		info.Tags = r.Tags
	}
	// 匿名处理
	if r.IsAnonymous {
		info.RaterName = "匿名用户"
		info.RaterAvatar = ""
	}
	return info
}

// Create 创建评价
func (s *ratingService) Create(regionID uint, raterID uint, raterName, raterAvatar string, req *dto.CreateRatingRequest) (*dto.RatingInfo, error) {
	ratingType := req.RatingType
	if ratingType == "" {
		ratingType = model.RatingTypePassengerToDriver
	}
	// 重复评价检查
	if has, err := s.repo.HasRated(raterID, req.PincheID, ratingType); err == nil && has {
		return nil, ErrRatingAlreadyExists
	}
	r := &model.PincheRating{
		PincheID:    req.PincheID,
		BookingID:   req.BookingID,
		TripID:      req.TripID,
		RaterID:     raterID,
		RaterName:   raterName,
		RaterAvatar: raterAvatar,
		RateeID:     0, // 由调用方或上层补齐
		RatingType:  ratingType,
		Rating:      req.Rating,
		Content:     req.Content,
		IsAnonymous: req.IsAnonymous,
		Status:      model.RatingStatusPending,
	}
	r.RegionID = regionID
	if req.Images != nil {
		if jb, err := model.FromJSON(req.Images); err == nil {
			r.Images = jb
		}
	}
	if req.Tags != nil {
		if jb, err := model.FromJSON(req.Tags); err == nil {
			r.Tags = jb
		}
	}
	if err := s.repo.Create(r); err != nil {
		return nil, err
	}
	return toRatingInfo(r), nil
}

// Update 更新评价
func (s *ratingService) Update(id uint, operatorID uint, req *dto.UpdateRatingRequest) error {
	r, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrRatingNotFound
		}
		return err
	}
	if r.RaterID != operatorID {
		return ErrRatingNoPermission
	}
	fields := map[string]interface{}{}
	if req.Content != nil {
		fields["content"] = *req.Content
	}
	if req.IsAnonymous != nil {
		fields["is_anonymous"] = *req.IsAnonymous
	}
	if req.Status != nil {
		fields["status"] = *req.Status
	}
	if req.Images != nil {
		if jb, err := model.FromJSON(req.Images); err == nil {
			fields["images"] = jb
		}
	}
	if req.Tags != nil {
		if jb, err := model.FromJSON(req.Tags); err == nil {
			fields["tags"] = jb
		}
	}
	if len(fields) == 0 {
		return nil
	}
	return s.repo.Update(id, fields)
}

// Delete 删除评价
func (s *ratingService) Delete(id uint, operatorID uint) error {
	r, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrRatingNotFound
		}
		return err
	}
	if r.RaterID != operatorID {
		return ErrRatingNoPermission
	}
	return s.repo.Delete(id)
}

// GetByID 获取详情
func (s *ratingService) GetByID(id uint) (*dto.RatingInfo, error) {
	r, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrRatingNotFound
		}
		return nil, err
	}
	return toRatingInfo(r), nil
}

// ListByRater 按评价人查询
func (s *ratingService) ListByRater(raterID uint, page, pageSize int) (*utils.Pagination, []dto.RatingInfo, error) {
	pagination := utils.NewPagination(page, pageSize)
	list, total, err := s.repo.ListByRater(raterID, pagination)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	result := make([]dto.RatingInfo, 0, len(list))
	for i := range list {
		result = append(result, *toRatingInfo(&list[i]))
	}
	return pagination, result, nil
}

// ListByRatee 按被评价人查询
func (s *ratingService) ListByRatee(rateeID uint, page, pageSize int) (*utils.Pagination, []dto.RatingInfo, error) {
	pagination := utils.NewPagination(page, pageSize)
	list, total, err := s.repo.ListByRatee(rateeID, pagination)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	result := make([]dto.RatingInfo, 0, len(list))
	for i := range list {
		result = append(result, *toRatingInfo(&list[i]))
	}
	return pagination, result, nil
}

// ListByPinche 按行程查询
func (s *ratingService) ListByPinche(pincheID uint, page, pageSize int) (*utils.Pagination, []dto.RatingInfo, error) {
	pagination := utils.NewPagination(page, pageSize)
	list, total, err := s.repo.ListByPinche(pincheID, pagination)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	result := make([]dto.RatingInfo, 0, len(list))
	for i := range list {
		result = append(result, *toRatingInfo(&list[i]))
	}
	return pagination, result, nil
}

// Reply 回复评价
func (s *ratingService) Reply(id uint, operatorID uint, req *dto.RatingReplyRequest) error {
	r, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrRatingNotFound
		}
		return err
	}
	// 仅被评价人可回复
	if r.RateeID != operatorID {
		return ErrRatingNoPermission
	}
	now := time.Now()
	fields := map[string]interface{}{
		"reply":    req.Reply,
		"reply_at": &now,
	}
	return s.repo.Update(id, fields)
}

// Like 点赞
func (s *ratingService) Like(id uint) error {
	return s.repo.IncrLikeCount(id)
}

// Stats 评价统计
func (s *ratingService) Stats(rateeID uint) (*dto.RatingStatsResponse, error) {
	total, avg, good, medium, bad, err := s.repo.StatsByTarget(rateeID)
	if err != nil {
		return nil, err
	}
	resp := &dto.RatingStatsResponse{
		TotalReviews: int(total),
		AvgRating:    avg,
	}
	if total > 0 {
		resp.GoodRate = float64(good) / float64(total)
		resp.MediumRate = float64(medium) / float64(total)
		resp.BadRate = float64(bad) / float64(total)
	}
	// has_reply_rate 暂用 medium 占比近似
	if total > 0 {
		resp.HasReplyRate = float64(medium) / float64(total)
	}
	return resp, nil
}

// AdminList 管理后台评价列表
func (s *ratingService) AdminList(req *dto.RatingListRequest) (*utils.Pagination, []dto.RatingInfo, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	opts := repository.RatingListOptions{
		PincheID:   req.PincheID,
		BookingID:  req.BookingID,
		RaterID:    req.RaterID,
		RateeID:    req.RateeID,
		RatingType: req.RatingType,
		Rating:     req.Rating,
		Status:     req.Status,
		Keyword:    req.Keyword,
	}
	// 跨地区：regionID=0
	list, total, err := s.repo.List(0, pagination, opts)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	result := make([]dto.RatingInfo, 0, len(list))
	for i := range list {
		result = append(result, *toRatingInfo(&list[i]))
	}
	return pagination, result, nil
}

// UpdateStatus 管理后台更新状态
func (s *ratingService) UpdateStatus(id uint, status int) error {
	return s.repo.UpdateStatus(id, status)
}
