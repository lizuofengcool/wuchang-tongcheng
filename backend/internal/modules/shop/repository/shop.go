// Package repository 商家模块数据访问层
package repository

import (
	"wuchang-tongcheng/internal/modules/shop/model"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

// ==================== 店铺仓储 ====================

// ShopRepository 店铺仓储接口
type ShopRepository interface {
	Create(shop *model.Shop) error
	FindByID(id uint) (*model.Shop, error)
	FindByUserID(userID uint, regionID uint) (*model.Shop, error)
	Update(shop *model.Shop) error
	UpdateFields(id uint, fields map[string]interface{}) error
	Delete(id uint) error
	List(regionID uint, pagination *utils.Pagination, categoryID uint, isRecommend int, keyword string) ([]model.Shop, int64, error)
	AdminList(regionID uint, pagination *utils.Pagination, categoryID uint, auditStatus int, status int, isRecommend int, keyword string) ([]model.Shop, int64, error)
	IncrViews(id uint) error
	UpdateRating(id uint, rating float32) error
}

type shopRepository struct {
	db *gorm.DB
}

// NewShopRepository 创建店铺仓储
func NewShopRepository(db *gorm.DB) ShopRepository {
	return &shopRepository{db: db}
}

func (r *shopRepository) Create(shop *model.Shop) error {
	return r.db.Create(shop).Error
}

func (r *shopRepository) FindByID(id uint) (*model.Shop, error) {
	var shop model.Shop
	if err := r.db.First(&shop, id).Error; err != nil {
		return nil, err
	}
	return &shop, nil
}

func (r *shopRepository) FindByUserID(userID uint, regionID uint) (*model.Shop, error) {
	var shop model.Shop
	query := r.db.Where("user_id = ?", userID)
	if regionID > 0 {
		query = query.Where("region_id = ?", regionID)
	}
	if err := query.First(&shop).Error; err != nil {
		return nil, err
	}
	return &shop, nil
}

func (r *shopRepository) Update(shop *model.Shop) error {
	return r.db.Save(shop).Error
}

func (r *shopRepository) UpdateFields(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.Shop{}).Where("id = ?", id).Updates(fields).Error
}

func (r *shopRepository) Delete(id uint) error {
	return r.db.Delete(&model.Shop{}, id).Error
}

func (r *shopRepository) List(regionID uint, pagination *utils.Pagination, categoryID uint, isRecommend int, keyword string) ([]model.Shop, int64, error) {
	var list []model.Shop
	var total int64

	query := r.db.Model(&model.Shop{})

	// 公开列表仅返回审核通过
	query = query.Where("audit_status = ?", model.AuditStatusApproved)

	if regionID > 0 {
		query = query.Where("region_id = ?", regionID)
	}
	if categoryID > 0 {
		query = query.Where("category_id = ?", categoryID)
	}
	if isRecommend == 0 || isRecommend == 1 {
		query = query.Where("is_recommend = ?", isRecommend)
	}
	if keyword != "" {
		query = query.Where("name LIKE ?", "%"+keyword+"%")
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 推荐优先，再按排序与浏览量倒序
	if err := query.Scopes(utils.Paginate(pagination)).
		Order("is_recommend DESC, sort DESC, views DESC, id DESC").
		Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

func (r *shopRepository) AdminList(regionID uint, pagination *utils.Pagination, categoryID uint, auditStatus int, status int, isRecommend int, keyword string) ([]model.Shop, int64, error) {
	var list []model.Shop
	var total int64

	query := r.db.Model(&model.Shop{})

	if regionID > 0 {
		query = query.Where("region_id = ?", regionID)
	}
	if categoryID > 0 {
		query = query.Where("category_id = ?", categoryID)
	}
	if auditStatus >= 0 && auditStatus <= 2 {
		query = query.Where("audit_status = ?", auditStatus)
	}
	if status >= 0 && status <= 2 {
		query = query.Where("status = ?", status)
	}
	if isRecommend == 0 || isRecommend == 1 {
		query = query.Where("is_recommend = ?", isRecommend)
	}
	if keyword != "" {
		query = query.Where("name LIKE ?", "%"+keyword+"%")
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Scopes(utils.Paginate(pagination)).
		Order("id DESC").
		Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

func (r *shopRepository) IncrViews(id uint) error {
	return r.db.Model(&model.Shop{}).Where("id = ?", id).
		UpdateColumn("views", gorm.Expr("views + 1")).Error
}

func (r *shopRepository) UpdateRating(id uint, rating float32) error {
	return r.db.Model(&model.Shop{}).Where("id = ?", id).
		Update("rating", rating).Error
}

// ==================== 店铺相册仓储 ====================

// ShopImageRepository 店铺相册仓储接口
type ShopImageRepository interface {
	Create(img *model.ShopImage) error
	FindByID(id uint) (*model.ShopImage, error)
	FindByShopID(shopID uint) ([]model.ShopImage, error)
	Delete(id uint) error
}

type shopImageRepository struct {
	db *gorm.DB
}

// NewShopImageRepository 创建店铺相册仓储
func NewShopImageRepository(db *gorm.DB) ShopImageRepository {
	return &shopImageRepository{db: db}
}

func (r *shopImageRepository) Create(img *model.ShopImage) error {
	return r.db.Create(img).Error
}

func (r *shopImageRepository) FindByID(id uint) (*model.ShopImage, error) {
	var img model.ShopImage
	if err := r.db.First(&img, id).Error; err != nil {
		return nil, err
	}
	return &img, nil
}

func (r *shopImageRepository) FindByShopID(shopID uint) ([]model.ShopImage, error) {
	var list []model.ShopImage
	if err := r.db.Where("shop_id = ?", shopID).Order("sort ASC, id ASC").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *shopImageRepository) Delete(id uint) error {
	return r.db.Delete(&model.ShopImage{}, id).Error
}

// ==================== 店铺评价仓储 ====================

// ShopReviewRepository 店铺评价仓储接口
type ShopReviewRepository interface {
	Create(review *model.ShopReview) error
	FindByID(id uint) (*model.ShopReview, error)
	FindApprovedByShopID(shopID uint, pagination *utils.Pagination) ([]model.ShopReview, int64, error)
	AdminList(pagination *utils.Pagination, shopID uint, status int) ([]model.ShopReview, int64, error)
	UpdateFields(id uint, fields map[string]interface{}) error
	AvgRatingByShopID(shopID uint) (float32, int64, error)
}

type shopReviewRepository struct {
	db *gorm.DB
}

// NewShopReviewRepository 创建店铺评价仓储
func NewShopReviewRepository(db *gorm.DB) ShopReviewRepository {
	return &shopReviewRepository{db: db}
}

func (r *shopReviewRepository) Create(review *model.ShopReview) error {
	return r.db.Create(review).Error
}

func (r *shopReviewRepository) FindByID(id uint) (*model.ShopReview, error) {
	var review model.ShopReview
	if err := r.db.First(&review, id).Error; err != nil {
		return nil, err
	}
	return &review, nil
}

func (r *shopReviewRepository) FindApprovedByShopID(shopID uint, pagination *utils.Pagination) ([]model.ShopReview, int64, error) {
	var list []model.ShopReview
	var total int64

	query := r.db.Model(&model.ShopReview{}).Where("shop_id = ?", shopID).
		Where("status = ?", model.ReviewStatusApproved)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Scopes(utils.Paginate(pagination)).
		Order("id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

func (r *shopReviewRepository) AdminList(pagination *utils.Pagination, shopID uint, status int) ([]model.ShopReview, int64, error) {
	var list []model.ShopReview
	var total int64

	query := r.db.Model(&model.ShopReview{})

	if shopID > 0 {
		query = query.Where("shop_id = ?", shopID)
	}
	if status >= 0 && status <= 2 {
		query = query.Where("status = ?", status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Scopes(utils.Paginate(pagination)).
		Order("id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

func (r *shopReviewRepository) UpdateFields(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.ShopReview{}).Where("id = ?", id).Updates(fields).Error
}

// AvgRatingByShopID 计算店铺已通过评价的平均分与数量
func (r *shopReviewRepository) AvgRatingByShopID(shopID uint) (float32, int64, error) {
	var result struct {
		Avg   float32
		Count int64
	}
	err := r.db.Model(&model.ShopReview{}).
		Where("shop_id = ? AND status = ?", shopID, model.ReviewStatusApproved).
		Select("COALESCE(AVG(rating), 0) AS avg, COUNT(*) AS count").
		Scan(&result).Error
	if err != nil {
		return 0, 0, err
	}
	return result.Avg, result.Count, nil
}
