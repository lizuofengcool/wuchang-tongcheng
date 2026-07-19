// Package service 店铺业务逻辑层
// 依据 v3.2.1 架构方案：对标转转商家版
package service

import (
	"errors"
	"time"

	"wuchang-tongcheng/internal/modules/ershou/dto"
	"wuchang-tongcheng/internal/modules/ershou/model"
	"wuchang-tongcheng/internal/modules/ershou/repository"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

var (
	ErrShopNotFound     = errors.New("店铺不存在")
	ErrShopExists       = errors.New("您已开通店铺")
	ErrShopNoPermission = errors.New("无权操作此店铺")
	ErrShopStatus       = errors.New("店铺状态不允许此操作")
)

// ShopService 店铺业务接口
type ShopService interface {
	Create(userID uint, req *dto.ShopCreateRequest) (*dto.ShopResponse, error)
	GetByID(shopID uint) (*dto.ShopResponse, error)
	GetByUserID(userID uint) (*dto.ShopResponse, error)
	Update(shopID, userID uint, req *dto.ShopUpdateRequest) (*dto.ShopResponse, error)
	List(query dto.ShopListQuery) (*utils.Pagination, []dto.ShopResponse, error)
	Audit(shopID, adminID uint, approved bool, rejectReason string) (*dto.ShopResponse, error)
	UpdateStatus(shopID, adminID uint, status int) (*dto.ShopResponse, error)

	// 关注/取关
	Follow(shopID, userID uint, req *dto.ShopFollowRequest) error
	Unfollow(shopID, userID uint) error
	ListFollowers(shopID uint, pagination *utils.Pagination) (*utils.Pagination, []model.ErshouShopFollower, error)
	ListUserFollowing(userID uint, pagination *utils.Pagination) (*utils.Pagination, []model.ErshouShopFollower, error)
}

type shopService struct {
	repo            repository.ShopRepository
	followerRepo    repository.ShopFollowerRepository
}

// NewShopService 创建店铺 service 实例
func NewShopService(
	repo repository.ShopRepository,
	followerRepo repository.ShopFollowerRepository,
) ShopService {
	return &shopService{repo: repo, followerRepo: followerRepo}
}

func shopLevelText(level int) string {
	switch level {
	case model.ShopLevelNormal:
		return "普通商家"
	case model.ShopLevelVerified:
		return "认证商家"
	case model.ShopLevelGold:
		return "金牌商家"
	case model.ShopLevelDiamond:
		return "钻石商家"
	}
	return "普通商家"
}

func shopStatusText(status int) string {
	switch status {
	case model.ShopStatusPending:
		return "待审核"
	case model.ShopStatusApproved:
		return "已通过"
	case model.ShopStatusRejected:
		return "已拒绝"
	case model.ShopStatusFrozen:
		return "已冻结"
	case model.ShopStatusClosed:
		return "已关闭"
	}
	return "未知"
}

// Create 用户开通店铺
func (s *shopService) Create(userID uint, req *dto.ShopCreateRequest) (*dto.ShopResponse, error) {
	// 同一用户只能有一个店铺
	if existing, _ := s.repo.FindByUserID(userID); existing != nil {
		return nil, ErrShopExists
	}
	shop := &model.ErshouShop{
		UserID:          userID,
		ShopName:        req.ShopName,
		Logo:            req.Logo,
		Banner:          req.Banner,
		Description:     req.Description,
		Level:           model.ShopLevelNormal,
		Status:          model.ShopStatusPending,
		ContactName:     req.ContactName,
		ContactPhone:    req.ContactPhone,
		ContactWechat:   req.ContactWechat,
		Address:         req.Address,
		Latitude:        req.Latitude,
		Longitude:       req.Longitude,
		BusinessLicense: req.BusinessLicense,
		LicenseNo:       req.LicenseNo,
		IDCardFront:     req.IDCardFront,
		IDCardBack:      req.IDCardBack,
		Deposit:         req.Deposit,
	}
	if len(req.Tags) > 0 {
		if jb, err := model.FromJSON(req.Tags); err == nil {
			shop.Tags = jb
		}
	}
	if err := s.repo.Create(shop); err != nil {
		return nil, err
	}
	return s.toShopResponse(shop, false), nil
}

func (s *shopService) GetByID(shopID uint) (*dto.ShopResponse, error) {
	shop, err := s.repo.FindByID(shopID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrShopNotFound
		}
		return nil, err
	}
	return s.toShopResponse(shop, false), nil
}

func (s *shopService) GetByUserID(userID uint) (*dto.ShopResponse, error) {
	shop, err := s.repo.FindByUserID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrShopNotFound
		}
		return nil, err
	}
	return s.toShopResponse(shop, false), nil
}

func (s *shopService) Update(shopID, userID uint, req *dto.ShopUpdateRequest) (*dto.ShopResponse, error) {
	shop, err := s.repo.FindByID(shopID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrShopNotFound
		}
		return nil, err
	}
	if shop.UserID != userID {
		return nil, ErrShopNoPermission
	}
	fields := map[string]interface{}{}
	if req.ShopName != "" {
		fields["shop_name"] = req.ShopName
	}
	if req.Logo != "" {
		fields["logo"] = req.Logo
	}
	if req.Banner != "" {
		fields["banner"] = req.Banner
	}
	if req.Description != "" {
		fields["description"] = req.Description
	}
	if req.Address != "" {
		fields["address"] = req.Address
	}
	if req.Latitude != 0 {
		fields["latitude"] = req.Latitude
	}
	if req.Longitude != 0 {
		fields["longitude"] = req.Longitude
	}
	if req.Tags != nil {
		if jb, err := model.FromJSON(req.Tags); err == nil {
			fields["tags"] = jb
		}
	}
	if len(fields) == 0 {
		return s.toShopResponse(shop, false), nil
	}
	if err := s.repo.Update(shopID, fields); err != nil {
		return nil, err
	}
	updated, _ := s.repo.FindByID(shopID)
	return s.toShopResponse(updated, false), nil
}

func (s *shopService) List(query dto.ShopListQuery) (*utils.Pagination, []dto.ShopResponse, error) {
	pagination := utils.NewPagination(query.Page, query.PageSize)
	list, total, err := s.repo.List(repository.ShopListQuery{
		UserID:  query.UserID,
		Status:  query.Status,
		Level:   query.Level,
		Keyword: query.Keyword,
	}, pagination)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	result := make([]dto.ShopResponse, 0, len(list))
	for i := range list {
		result = append(result, *s.toShopResponse(&list[i], false))
	}
	return pagination, result, nil
}

// Audit 店铺审核（M端审核员）
func (s *shopService) Audit(shopID, adminID uint, approved bool, rejectReason string) (*dto.ShopResponse, error) {
	shop, err := s.repo.FindByID(shopID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrShopNotFound
		}
		return nil, err
	}
	if shop.Status != model.ShopStatusPending {
		return nil, ErrShopStatus
	}
	now := time.Now()
	fields := map[string]interface{}{}
	if approved {
		fields["status"] = model.ShopStatusApproved
		fields["approved_at"] = &now
		fields["verified_at"] = &now
		fields["rejected_reason"] = ""
	} else {
		fields["status"] = model.ShopStatusRejected
		fields["rejected_reason"] = rejectReason
	}
	if err := s.repo.Update(shopID, fields); err != nil {
		return nil, err
	}
	updated, _ := s.repo.FindByID(shopID)
	return s.toShopResponse(updated, false), nil
}

// UpdateStatus 更新店铺状态（冻结/关闭/恢复）
func (s *shopService) UpdateStatus(shopID, adminID uint, status int) (*dto.ShopResponse, error) {
	if _, err := s.repo.FindByID(shopID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrShopNotFound
		}
		return nil, err
	}
	fields := map[string]interface{}{"status": status}
	if status == model.ShopStatusClosed {
		now := time.Now()
		fields["closed_at"] = &now
	}
	if err := s.repo.Update(shopID, fields); err != nil {
		return nil, err
	}
	updated, _ := s.repo.FindByID(shopID)
	return s.toShopResponse(updated, false), nil
}

// Follow 关注店铺
func (s *shopService) Follow(shopID, userID uint, req *dto.ShopFollowRequest) error {
	if _, err := s.repo.FindByID(shopID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrShopNotFound
		}
		return err
	}
	// 已关注则幂等返回
	if following, _ := s.followerRepo.IsFollowing(shopID, userID); following {
		return nil
	}
	f := &model.ErshouShopFollower{
		ShopID: shopID,
		UserID: userID,
		Notify: req.Notify,
	}
	if err := s.followerRepo.Follow(f); err != nil {
		return err
	}
	_ = s.followerRepo.IncrFollowerCount(shopID, 1)
	return nil
}

// Unfollow 取关店铺
func (s *shopService) Unfollow(shopID, userID uint) error {
	if err := s.followerRepo.Unfollow(shopID, userID); err != nil {
		return err
	}
	_ = s.followerRepo.IncrFollowerCount(shopID, -1)
	return nil
}

func (s *shopService) ListFollowers(shopID uint, pagination *utils.Pagination) (*utils.Pagination, []model.ErshouShopFollower, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 20)
	}
	list, total, err := s.followerRepo.ListFollowers(shopID, pagination)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	return pagination, list, nil
}

func (s *shopService) ListUserFollowing(userID uint, pagination *utils.Pagination) (*utils.Pagination, []model.ErshouShopFollower, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 20)
	}
	list, total, err := s.followerRepo.ListUserFollowing(userID, pagination)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	return pagination, list, nil
}

func (s *shopService) toShopResponse(shop *model.ErshouShop, isFollowing bool) *dto.ShopResponse {
	resp := &dto.ShopResponse{
		ID:              shop.ID,
		UserID:          shop.UserID,
		ShopName:        shop.ShopName,
		Logo:            shop.Logo,
		Banner:          shop.Banner,
		Description:     shop.Description,
		Level:           shop.Level,
		LevelText:       shopLevelText(shop.Level),
		Status:          shop.Status,
		StatusText:      shopStatusText(shop.Status),
		ContactName:     shop.ContactName,
		ContactPhone:    shop.ContactPhone,
		ContactWechat:   shop.ContactWechat,
		Address:         shop.Address,
		Latitude:        shop.Latitude,
		Longitude:       shop.Longitude,
		BusinessLicense: shop.BusinessLicense,
		LicenseNo:       shop.LicenseNo,
		VerifiedAt:      shop.VerifiedAt,
		FollowerCount:   shop.FollowerCount,
		ItemCount:       shop.ItemCount,
		SoldCount:       shop.SoldCount,
		TotalAmount:     shop.TotalAmount,
		GoodRate:        shop.GoodRate,
		Deposit:         shop.Deposit,
		Tags:            []string{},
		ApprovedAt:      shop.ApprovedAt,
		RejectedReason:  shop.RejectedReason,
		IsFollowing:     isFollowing,
		CreatedAt:       shop.CreatedAt,
	}
	if shop.Tags != nil {
		var tags []string
		_ = shop.Tags.Parse(&tags)
		if tags != nil {
			resp.Tags = tags
		}
	}
	return resp
}
