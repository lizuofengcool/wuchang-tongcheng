// Package service 同城商城业务逻辑层 - 评价
package service

import (
	"errors"
	"time"

	"wuchang-tongcheng/internal/modules/mall/dto"
	"wuchang-tongcheng/internal/modules/mall/model"
	"wuchang-tongcheng/internal/modules/mall/repository"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

var (
	ErrReviewNotFound = errors.New("评价不存在")
	ErrReviewNotOwner = errors.New("无权操作他人评价")
	ErrReviewExists   = errors.New("该订单明细已评价")
)

// ReviewService 评价业务接口
type ReviewService interface {
	Create(regionID, userID uint, userName string, req *dto.CreateReviewRequest) (*dto.ReviewInfo, error)
	Reply(id, replyUserID uint, req *dto.ReplyReviewRequest) error
	Append(id, userID uint, req *dto.AppendReviewRequest) error
	Delete(id uint) error
	GetByID(id uint) (*dto.ReviewInfo, error)
	List(req *dto.ReviewListRequest) (*utils.Pagination, []dto.ReviewInfo, error)
	AdminList(req *dto.AdminReviewListRequest) (*utils.Pagination, []dto.ReviewInfo, error)
	ListByProduct(productID uint, page, pageSize int) (*utils.Pagination, []dto.ReviewInfo, error)
	ListByShop(shopID uint, page, pageSize int) (*utils.Pagination, []dto.ReviewInfo, error)
	ListByUser(userID uint, page, pageSize int) (*utils.Pagination, []dto.ReviewInfo, error)
	ListByOrder(orderID uint) ([]dto.ReviewInfo, error)

	UpdateStatus(id uint, status int, reason string) error
	Stats(req *dto.ReviewStatsRequest) (*dto.ReviewStatsInfo, error)

	IncrLikeCount(id uint) error
	DecrLikeCount(id uint) error
	IncrDislikeCount(id uint) error
	DecrDislikeCount(id uint) error
}

type reviewService struct {
	repo          repository.ReviewRepository
	orderItemRepo repository.OrderItemRepository
}

// NewReviewService 创建评价 service 实例
func NewReviewService(repo repository.ReviewRepository, orderItemRepo repository.OrderItemRepository) ReviewService {
	return &reviewService{repo: repo, orderItemRepo: orderItemRepo}
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
func toReviewInfo(r *model.Review) *dto.ReviewInfo {
	info := &dto.ReviewInfo{
		ID:              r.ID,
		ProductID:       r.ProductID,
		SkuID:           r.SkuID,
		OrderID:         r.OrderID,
		OrderNo:         r.OrderNo,
		OrderItemID:     r.OrderItemID,
		UserID:          r.UserID,
		ShopID:          r.ShopID,
		UserName:        r.UserName,
		UserAvatar:      r.UserAvatar,
		IsAnonymous:     r.IsAnonymous,
		Rating:          r.Rating,
		Content:         r.Content,
		Video:           r.Video,
		SkuName:         r.SkuName,
		SkuSpecs:        r.SkuSpecs,
		Reply:           r.Reply,
		ReplyAt:         r.ReplyAt,
		ReplyUserID:     r.ReplyUserID,
		Status:          r.Status,
		StatusText:      reviewStatusText(r.Status),
		AuditReason:     r.AuditReason,
		HasSellerReply:  r.HasSellerReply,
		LikeCount:       r.LikeCount,
		DislikeCount:    r.DislikeCount,
		ReplyCount:      r.ReplyCount,
		AppendContent:   r.AppendContent,
		AppendAt:        r.AppendAt,
		ContentHash:     r.ContentHash,
		RiskScore:       r.RiskScore,
		RegionID:        r.RegionID,
		CreatedAt:       r.CreatedAt,
		UpdatedAt:       r.UpdatedAt,
	}
	// Type 字段在 model 是 string，在 dto 是 int（保持兼容：留空）
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
func (s *reviewService) Create(regionID, userID uint, userName string, req *dto.CreateReviewRequest) (*dto.ReviewInfo, error) {
	// 检查订单明细是否已评价
	item, err := s.orderItemRepo.FindByID(req.OrderItemID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("订单明细不存在")
		}
		return nil, err
	}
	if item.HasReview {
		return nil, ErrReviewExists
	}
	if item.UserID != userID {
		return nil, errors.New("无权评价他人订单")
	}

	r := &model.Review{
		ProductID:    req.ProductID,
		SkuID:        req.SkuID,
		OrderID:      req.OrderID,
		OrderItemID:  req.OrderItemID,
		UserID:       userID,
		ShopID:       item.ShopID,
		UserName:     userName,
		IsAnonymous:  req.IsAnonymous,
		Rating:       req.Rating,
		Content:      req.Content,
		Video:        req.Video,
		SkuName:      item.SkuName,
		SkuSpecs:     item.SkuSpecs,
		Status:       model.ReviewStatusPending,
	}
	r.RegionID = regionID
	if req.Images != nil {
		if b, err := model.FromJSON(req.Images); err == nil {
			r.Images = b
		}
	}
	if req.Tags != nil {
		if b, err := model.FromJSON(req.Tags); err == nil {
			r.Tags = b
		}
	}

	if err := s.repo.Create(r); err != nil {
		return nil, err
	}

	// 更新订单明细评价状态
	_ = s.orderItemRepo.UpdateReviewStatus(item.ID, true, r.ID)

	return toReviewInfo(r), nil
}

// Reply 商家回复评价
func (s *reviewService) Reply(id, replyUserID uint, req *dto.ReplyReviewRequest) error {
	if _, err := s.repo.FindByID(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrReviewNotFound
		}
		return err
	}
	now := time.Now()
	_ = s.repo.UpdateFields(id, map[string]interface{}{
		"reply_at": &now,
	})
	return s.repo.UpdateReply(id, req.Reply, replyUserID)
}

// Append 追加评价
func (s *reviewService) Append(id, userID uint, req *dto.AppendReviewRequest) error {
	r, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrReviewNotFound
		}
		return err
	}
	if r.UserID != userID {
		return ErrReviewNotOwner
	}

	var appendImages model.JSONB
	if req.AppendImages != nil {
		if b, err := model.FromJSON(req.AppendImages); err == nil {
			appendImages = b
		}
	}
	now := time.Now()
	_ = s.repo.UpdateFields(id, map[string]interface{}{
		"append_at": &now,
	})
	return s.repo.UpdateAppend(id, req.AppendContent, appendImages)
}

// Delete 删除评价
func (s *reviewService) Delete(id uint) error {
	if _, err := s.repo.FindByID(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrReviewNotFound
		}
		return err
	}
	return s.repo.Delete(id)
}

// GetByID 获取评价详情
func (s *reviewService) GetByID(id uint) (*dto.ReviewInfo, error) {
	r, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrReviewNotFound
		}
		return nil, err
	}
	return toReviewInfo(r), nil
}

// List 评价列表（C 端）
func (s *reviewService) List(req *dto.ReviewListRequest) (*utils.Pagination, []dto.ReviewInfo, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	opts := repository.ReviewListOptions{
		ProductID: req.ProductID,
		SkuID:     req.SkuID,
		ShopID:    req.ShopID,
		UserID:    req.UserID,
		OrderID:   req.OrderID,
		Rating:    req.Rating,
		Status:    req.Status,
		HasReply:  req.HasReply,
		HasImages: req.HasImages,
		Sort:      req.Sort,
		Keyword:   req.Keyword,
	}
	// 默认只看通过的评价
	if opts.Status == nil {
		approved := model.ReviewStatusApproved
		opts.Status = &approved
	}
	list, total, err := s.repo.List(opts, pagination)
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

// AdminList 管理后台评价列表
func (s *reviewService) AdminList(req *dto.AdminReviewListRequest) (*utils.Pagination, []dto.ReviewInfo, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	opts := repository.ReviewListOptions{
		ProductID: req.ProductID,
		ShopID:    req.ShopID,
		UserID:    req.UserID,
		Status:    req.Status,
		Rating:    req.Rating,
		Keyword:   req.Keyword,
	}
	list, total, err := s.repo.List(opts, pagination)
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

// ListByProduct 按商品列出评价
func (s *reviewService) ListByProduct(productID uint, page, pageSize int) (*utils.Pagination, []dto.ReviewInfo, error) {
	pagination := utils.NewPagination(page, pageSize)
	list, total, err := s.repo.ListByProduct(productID, pagination)
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

// ListByShop 按店铺列出评价
func (s *reviewService) ListByShop(shopID uint, page, pageSize int) (*utils.Pagination, []dto.ReviewInfo, error) {
	pagination := utils.NewPagination(page, pageSize)
	list, total, err := s.repo.ListByShop(shopID, pagination)
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

// ListByUser 按用户列出评价
func (s *reviewService) ListByUser(userID uint, page, pageSize int) (*utils.Pagination, []dto.ReviewInfo, error) {
	pagination := utils.NewPagination(page, pageSize)
	list, total, err := s.repo.ListByUser(userID, pagination)
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

// ListByOrder 按订单列出评价
func (s *reviewService) ListByOrder(orderID uint) ([]dto.ReviewInfo, error) {
	list, err := s.repo.ListByOrder(orderID)
	if err != nil {
		return nil, err
	}
	infos := make([]dto.ReviewInfo, 0, len(list))
	for i := range list {
		infos = append(infos, *toReviewInfo(&list[i]))
	}
	return infos, nil
}

// UpdateStatus 更新评价状态（管理后台审核）
func (s *reviewService) UpdateStatus(id uint, status int, reason string) error {
	fields := map[string]interface{}{}
	if reason != "" {
		fields["audit_reason"] = reason
	}
	return s.repo.UpdateStatus(id, status, fields)
}

// Stats 评价统计
func (s *reviewService) Stats(req *dto.ReviewStatsRequest) (*dto.ReviewStatsInfo, error) {
	var result *repository.ReviewStatsResult
	var err error
	if req.ProductID > 0 {
		result, err = s.repo.StatsByProduct(req.ProductID)
	} else if req.ShopID > 0 {
		result, err = s.repo.StatsByShop(req.ShopID)
	} else {
		return &dto.ReviewStatsInfo{}, nil
	}
	if err != nil {
		return nil, err
	}
	return &dto.ReviewStatsInfo{
		TotalCount:      result.TotalCount,
		AvgRating:       result.AvgRating,
		FiveStarCount:   result.FiveStarCount,
		FourStarCount:   result.FourStarCount,
		ThreeStarCount:  result.ThreeStarCount,
		TwoStarCount:    result.TwoStarCount,
		OneStarCount:    result.OneStarCount,
		HasImagesCount:  result.HasImagesCount,
		HasVideoCount:   result.HasVideoCount,
		GoodRate:        result.GoodRate,
	}, nil
}

// IncrLikeCount 点赞 +1
func (s *reviewService) IncrLikeCount(id uint) error {
	return s.repo.IncrLikeCount(id)
}

// DecrLikeCount 点赞 -1
func (s *reviewService) DecrLikeCount(id uint) error {
	return s.repo.DecrLikeCount(id)
}

// IncrDislikeCount 踩 +1
func (s *reviewService) IncrDislikeCount(id uint) error {
	return s.repo.IncrDislikeCount(id)
}

// DecrDislikeCount 踩 -1
func (s *reviewService) DecrDislikeCount(id uint) error {
	return s.repo.DecrDislikeCount(id)
}
