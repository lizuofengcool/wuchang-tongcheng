// Package service 商家模块业务逻辑层
package service

import (
	"errors"
	"math"

	"wuchang-tongcheng/internal/modules/shop/dto"
	"wuchang-tongcheng/internal/modules/shop/model"
	"wuchang-tongcheng/internal/modules/shop/repository"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

var (
	// ErrShopNotFound 店铺不存在
	ErrShopNotFound = errors.New("店铺不存在")
	// ErrShopAlreadyExists 已存在店铺
	ErrShopAlreadyExists = errors.New("您已申请过店铺")
	// ErrShopNoPermission 无权操作该店铺
	ErrShopNoPermission = errors.New("无权操作此店铺")
	// ErrShopImageNotFound 店铺图片不存在
	ErrShopImageNotFound = errors.New("店铺图片不存在")
	// ErrShopReviewNotFound 评价不存在
	ErrShopReviewNotFound = errors.New("评价不存在")
	// ErrShopNotApproved 店铺未审核通过
	ErrShopNotApproved = errors.New("店铺未审核通过")
)

// ShopService 商家业务逻辑接口
type ShopService interface {
	// 公开接口
	GetByID(id uint) (*dto.ShopInfo, error)
	List(regionID uint, req *dto.ShopListRequest) (*utils.Pagination, []dto.ShopInfo, error)
	GetImages(shopID uint) ([]dto.ShopImageInfo, error)
	GetReviews(shopID uint, req *dto.ReviewListRequest) (*utils.Pagination, []dto.ShopReviewInfo, error)

	// 用户接口
	Apply(regionID uint, userID uint, req *dto.ApplyShopRequest) (*dto.ShopInfo, error)
	GetMyShop(userID uint, regionID uint) (*dto.ShopInfo, error)
	UpdateMyShop(userID uint, regionID uint, req *dto.UpdateShopRequest) error
	AddImage(userID uint, regionID uint, req *dto.AddShopImageRequest) (*dto.ShopImageInfo, error)
	DeleteImage(userID uint, imageID uint) error
	CreateReview(regionID uint, userID uint, shopID uint, req *dto.CreateReviewRequest) (*dto.ShopReviewInfo, error)

	// 管理接口
	AdminList(regionID uint, req *dto.AdminShopListRequest) (*utils.Pagination, []dto.ShopInfo, error)
	AuditShop(id uint, req *dto.AuditShopRequest) error
	UpdateShopStatus(id uint, req *dto.UpdateShopStatusRequest) error
	SetRecommend(id uint, req *dto.SetRecommendRequest) error
	DeleteShop(id uint) error
	AdminReviewList(req *dto.AdminReviewListRequest) (*utils.Pagination, []dto.ShopReviewInfo, error)
	AuditReview(id uint, req *dto.AuditReviewRequest) error
}

type shopService struct {
	shopRepo  repository.ShopRepository
	imageRepo repository.ShopImageRepository
	reviewRepo repository.ShopReviewRepository
}

// NewShopService 创建商家服务
func NewShopService(shopRepo repository.ShopRepository, imageRepo repository.ShopImageRepository, reviewRepo repository.ShopReviewRepository) ShopService {
	return &shopService{
		shopRepo:   shopRepo,
		imageRepo:  imageRepo,
		reviewRepo: reviewRepo,
	}
}

func toShopInfo(s *model.Shop) *dto.ShopInfo {
	return &dto.ShopInfo{
		ID:            s.ID,
		Name:          s.Name,
		Logo:          s.Logo,
		Description:   s.Description,
		Phone:         s.Phone,
		Address:       s.Address,
		Longitude:     s.Longitude,
		Latitude:      s.Latitude,
		CategoryID:    s.CategoryID,
		BusinessHours: s.BusinessHours,
		Status:        s.Status,
		AuditStatus:   s.AuditStatus,
		Rating:        s.Rating,
		Views:         s.Views,
		IsRecommend:   s.IsRecommend,
		Sort:          s.Sort,
		UserID:        s.UserID,
		RegionID:      s.RegionID,
		CreatedAt:     s.CreatedAt,
		UpdatedAt:     s.UpdatedAt,
	}
}

func toImageInfo(img *model.ShopImage) *dto.ShopImageInfo {
	return &dto.ShopImageInfo{
		ID:        img.ID,
		ShopID:    img.ShopID,
		ImageURL:  img.ImageURL,
		Sort:      img.Sort,
		CreatedAt: img.CreatedAt,
	}
}

func toReviewInfo(r *model.ShopReview) *dto.ShopReviewInfo {
	return &dto.ShopReviewInfo{
		ID:        r.ID,
		ShopID:    r.ShopID,
		UserID:    r.UserID,
		Rating:    r.Rating,
		Content:   r.Content,
		Reply:     r.Reply,
		ReplyAt:   r.ReplyAt,
		Status:    r.Status,
		CreatedAt: r.CreatedAt,
	}
}

// ==================== 公开接口 ====================

// GetByID 获取店铺详情（同时增加浏览量）
func (s *shopService) GetByID(id uint) (*dto.ShopInfo, error) {
	shop, err := s.shopRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrShopNotFound
		}
		return nil, err
	}
	_ = s.shopRepo.IncrViews(id)
	shop.Views++
	return toShopInfo(shop), nil
}

// List 店铺列表
func (s *shopService) List(regionID uint, req *dto.ShopListRequest) (*utils.Pagination, []dto.ShopInfo, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	if req.RegionID > 0 {
		regionID = req.RegionID
	}
	isRecommend := req.IsRecommend
	list, total, err := s.shopRepo.List(regionID, pagination, req.CategoryID, isRecommend, req.Keyword)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total

	result := make([]dto.ShopInfo, 0, len(list))
	for i := range list {
		result = append(result, *toShopInfo(&list[i]))
	}
	return pagination, result, nil
}

// GetImages 获取店铺相册
func (s *shopService) GetImages(shopID uint) ([]dto.ShopImageInfo, error) {
	list, err := s.imageRepo.FindByShopID(shopID)
	if err != nil {
		return nil, err
	}
	result := make([]dto.ShopImageInfo, 0, len(list))
	for i := range list {
		result = append(result, *toImageInfo(&list[i]))
	}
	return result, nil
}

// GetReviews 获取店铺评价列表（仅已通过）
func (s *shopService) GetReviews(shopID uint, req *dto.ReviewListRequest) (*utils.Pagination, []dto.ShopReviewInfo, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	list, total, err := s.reviewRepo.FindApprovedByShopID(shopID, pagination)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total

	result := make([]dto.ShopReviewInfo, 0, len(list))
	for i := range list {
		result = append(result, *toReviewInfo(&list[i]))
	}
	return pagination, result, nil
}

// ==================== 用户接口 ====================

// Apply 商家入驻申请
func (s *shopService) Apply(regionID uint, userID uint, req *dto.ApplyShopRequest) (*dto.ShopInfo, error) {
	// 同一用户在同一地区仅可申请一个店铺
	if existing, err := s.shopRepo.FindByUserID(userID, regionID); err == nil && existing != nil {
		return nil, ErrShopAlreadyExists
	} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	shop := &model.Shop{
		Name:          req.Name,
		Logo:          req.Logo,
		Description:   req.Description,
		Phone:         req.Phone,
		Address:       req.Address,
		Longitude:     req.Longitude,
		Latitude:      req.Latitude,
		CategoryID:    req.CategoryID,
		BusinessHours: req.BusinessHours,
		Status:        model.ShopStatusClosed,
		AuditStatus:   model.AuditStatusPending,
		UserID:        userID,
	}
	shop.RegionID = regionID

	if err := s.shopRepo.Create(shop); err != nil {
		return nil, err
	}
	return toShopInfo(shop), nil
}

// GetMyShop 获取我的店铺
func (s *shopService) GetMyShop(userID uint, regionID uint) (*dto.ShopInfo, error) {
	shop, err := s.shopRepo.FindByUserID(userID, regionID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrShopNotFound
		}
		return nil, err
	}
	return toShopInfo(shop), nil
}

// UpdateMyShop 编辑我的店铺
func (s *shopService) UpdateMyShop(userID uint, regionID uint, req *dto.UpdateShopRequest) error {
	shop, err := s.shopRepo.FindByUserID(userID, regionID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrShopNotFound
		}
		return err
	}
	if shop.UserID != userID {
		return ErrShopNoPermission
	}

	fields := map[string]interface{}{}
	if req.Name != "" {
		fields["name"] = req.Name
	}
	if req.Logo != "" {
		fields["logo"] = req.Logo
	}
	if req.Description != "" {
		fields["description"] = req.Description
	}
	if req.Phone != "" {
		fields["phone"] = req.Phone
	}
	if req.Address != "" {
		fields["address"] = req.Address
	}
	fields["longitude"] = req.Longitude
	fields["latitude"] = req.Latitude
	fields["category_id"] = req.CategoryID
	if req.BusinessHours != "" {
		fields["business_hours"] = req.BusinessHours
	}
	// 商家可自行调整营业状态
	if req.Status >= 0 && req.Status <= 2 {
		fields["status"] = req.Status
	}

	if len(fields) == 0 {
		return nil
	}
	return s.shopRepo.UpdateFields(shop.ID, fields)
}

// AddImage 上传店铺图片（基于我的店铺）
func (s *shopService) AddImage(userID uint, regionID uint, req *dto.AddShopImageRequest) (*dto.ShopImageInfo, error) {
	shop, err := s.shopRepo.FindByUserID(userID, regionID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrShopNotFound
		}
		return nil, err
	}
	if shop.UserID != userID {
		return nil, ErrShopNoPermission
	}

	img := &model.ShopImage{
		ShopID:   shop.ID,
		ImageURL: req.ImageURL,
		Sort:     req.Sort,
	}
	img.RegionID = regionID

	if err := s.imageRepo.Create(img); err != nil {
		return nil, err
	}
	return toImageInfo(img), nil
}

// DeleteImage 删除店铺图片（仅限本人店铺）
func (s *shopService) DeleteImage(userID uint, imageID uint) error {
	img, err := s.imageRepo.FindByID(imageID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrShopImageNotFound
		}
		return err
	}
	shop, err := s.shopRepo.FindByID(img.ShopID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrShopNotFound
		}
		return err
	}
	if shop.UserID != userID {
		return ErrShopNoPermission
	}
	return s.imageRepo.Delete(imageID)
}

// CreateReview 发表评价
func (s *shopService) CreateReview(regionID uint, userID uint, shopID uint, req *dto.CreateReviewRequest) (*dto.ShopReviewInfo, error) {
	shop, err := s.shopRepo.FindByID(shopID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrShopNotFound
		}
		return nil, err
	}
	if shop.AuditStatus != model.AuditStatusApproved {
		return nil, ErrShopNotApproved
	}

	review := &model.ShopReview{
		ShopID:  shopID,
		UserID:  userID,
		Rating:  req.Rating,
		Content: req.Content,
		Status:  model.ReviewStatusPending,
	}
	review.RegionID = regionID

	if err := s.reviewRepo.Create(review); err != nil {
		return nil, err
	}
	return toReviewInfo(review), nil
}

// ==================== 管理接口 ====================

// AdminList 管理端店铺列表
func (s *shopService) AdminList(regionID uint, req *dto.AdminShopListRequest) (*utils.Pagination, []dto.ShopInfo, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	if req.RegionID > 0 {
		regionID = req.RegionID
	}
	list, total, err := s.shopRepo.AdminList(regionID, pagination, req.CategoryID, req.AuditStatus, req.Status, req.IsRecommend, req.Keyword)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total

	result := make([]dto.ShopInfo, 0, len(list))
	for i := range list {
		result = append(result, *toShopInfo(&list[i]))
	}
	return pagination, result, nil
}

// AuditShop 审核店铺
func (s *shopService) AuditShop(id uint, req *dto.AuditShopRequest) error {
	if _, err := s.shopRepo.FindByID(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrShopNotFound
		}
		return err
	}
	fields := map[string]interface{}{
		"audit_status": req.AuditStatus,
	}
	// 审核通过后默认置为营业中，拒绝则歇业
	if req.AuditStatus == model.AuditStatusApproved {
		fields["status"] = model.ShopStatusOpen
	} else {
		fields["status"] = model.ShopStatusClosed
	}
	return s.shopRepo.UpdateFields(id, fields)
}

// UpdateShopStatus 修改营业状态
func (s *shopService) UpdateShopStatus(id uint, req *dto.UpdateShopStatusRequest) error {
	if _, err := s.shopRepo.FindByID(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrShopNotFound
		}
		return err
	}
	return s.shopRepo.UpdateFields(id, map[string]interface{}{"status": req.Status})
}

// SetRecommend 设置推荐
func (s *shopService) SetRecommend(id uint, req *dto.SetRecommendRequest) error {
	if _, err := s.shopRepo.FindByID(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrShopNotFound
		}
		return err
	}
	return s.shopRepo.UpdateFields(id, map[string]interface{}{"is_recommend": req.IsRecommend})
}

// DeleteShop 删除店铺
func (s *shopService) DeleteShop(id uint) error {
	if _, err := s.shopRepo.FindByID(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrShopNotFound
		}
		return err
	}
	return s.shopRepo.Delete(id)
}

// AdminReviewList 管理端评价列表
func (s *shopService) AdminReviewList(req *dto.AdminReviewListRequest) (*utils.Pagination, []dto.ShopReviewInfo, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	list, total, err := s.reviewRepo.AdminList(pagination, req.ShopID, req.Status)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total

	result := make([]dto.ShopReviewInfo, 0, len(list))
	for i := range list {
		result = append(result, *toReviewInfo(&list[i]))
	}
	return pagination, result, nil
}

// AuditReview 审核评价，并重算店铺评分
func (s *shopService) AuditReview(id uint, req *dto.AuditReviewRequest) error {
	review, err := s.reviewRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrShopReviewNotFound
		}
		return err
	}
	if err := s.reviewRepo.UpdateFields(id, map[string]interface{}{"status": req.Status}); err != nil {
		return err
	}

	// 重新计算店铺评分（仅统计已通过评价）
	return s.recomputeShopRating(review.ShopID)
}

// recomputeShopRating 重新计算并更新店铺评分
func (s *shopService) recomputeShopRating(shopID uint) error {
	avg, _, err := s.reviewRepo.AvgRatingByShopID(shopID)
	if err != nil {
		return err
	}
	// 保留一位小数
	rating := float32(math.Round(float64(avg)*10) / 10)
	return s.shopRepo.UpdateRating(shopID, rating)
}
