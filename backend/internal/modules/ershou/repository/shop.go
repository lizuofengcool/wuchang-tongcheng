// Package repository 店铺+粉丝数据访问层
// 依据 v3.2.1 架构方案：对标转转商家版
package repository

import (
	"wuchang-tongcheng/internal/modules/ershou/model"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

// ===== Shop =====

// ShopRepository 店铺仓储接口
type ShopRepository interface {
	Create(s *model.ErshouShop) error
	FindByID(id uint) (*model.ErshouShop, error)
	FindByUserID(userID uint) (*model.ErshouShop, error)
	Update(id uint, fields map[string]interface{}) error
	Delete(id uint) error
	List(query ShopListQuery, pagination *utils.Pagination) ([]model.ErshouShop, int64, error)
	IncrItemCount(shopID uint, delta int) error
	IncrSoldCount(shopID uint, n int) error
	IncrTotalAmount(shopID uint, amount float64) error
	StatsByUserID(userID uint) (itemCount, soldCount int64, totalAmount float64, err error)
}

// ShopListQuery 店铺列表查询
type ShopListQuery struct {
	UserID   uint
	Status   *int
	Level    *int
	Keyword  string
}

type shopRepository struct {
	db *gorm.DB
}

// NewShopRepository 创建店铺仓储实例
func NewShopRepository(db *gorm.DB) ShopRepository {
	return &shopRepository{db: db}
}

func (r *shopRepository) Create(s *model.ErshouShop) error {
	return r.db.Create(s).Error
}

func (r *shopRepository) FindByID(id uint) (*model.ErshouShop, error) {
	var s model.ErshouShop
	if err := r.db.First(&s, id).Error; err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *shopRepository) FindByUserID(userID uint) (*model.ErshouShop, error) {
	var s model.ErshouShop
	if err := r.db.Where("user_id = ?", userID).First(&s).Error; err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *shopRepository) Update(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.ErshouShop{}).Where("id = ?", id).Updates(fields).Error
}

func (r *shopRepository) Delete(id uint) error {
	return r.db.Delete(&model.ErshouShop{}, id).Error
}

func (r *shopRepository) List(query ShopListQuery, pagination *utils.Pagination) ([]model.ErshouShop, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.ErshouShop
	var total int64

	q := r.db.Model(&model.ErshouShop{})
	if query.UserID > 0 {
		q = q.Where("user_id = ?", query.UserID)
	}
	if query.Status != nil {
		q = q.Where("status = ?", *query.Status)
	}
	if query.Level != nil {
		q = q.Where("level = ?", *query.Level)
	}
	if query.Keyword != "" {
		q = q.Where("shop_name ILIKE ?", "%"+query.Keyword+"%")
	}

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := q.Scopes(utils.Paginate(pagination)).
		Order("created_at DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *shopRepository) IncrItemCount(shopID uint, delta int) error {
	return r.db.Model(&model.ErshouShop{}).Where("id = ?", shopID).
		UpdateColumn("item_count", gorm.Expr("item_count + ?", delta)).Error
}

func (r *shopRepository) IncrSoldCount(shopID uint, n int) error {
	return r.db.Model(&model.ErshouShop{}).Where("id = ?", shopID).
		UpdateColumn("sold_count", gorm.Expr("sold_count + ?", n)).Error
}

func (r *shopRepository) IncrTotalAmount(shopID uint, amount float64) error {
	return r.db.Model(&model.ErshouShop{}).Where("id = ?", shopID).
		UpdateColumn("total_amount", gorm.Expr("total_amount + ?", amount)).Error
}

func (r *shopRepository) StatsByUserID(userID uint) (int64, int64, float64, error) {
	type stat struct {
		ItemCount   int64   `gorm:"column:item_count"`
		SoldCount   int64   `gorm:"column:sold_count"`
		TotalAmount float64 `gorm:"column:total_amount"`
	}
	var s stat
	err := r.db.Model(&model.ErshouShop{}).
		Select("item_count, sold_count, total_amount").
		Where("user_id = ?", userID).
		Scan(&s).Error
	return s.ItemCount, s.SoldCount, s.TotalAmount, err
}

// ===== ShopFollower =====

// ShopFollowerRepository 店铺粉丝仓储接口
type ShopFollowerRepository interface {
	Follow(f *model.ErshouShopFollower) error
	Unfollow(shopID, userID uint) error
	IsFollowing(shopID, userID uint) (bool, error)
	ListFollowers(shopID uint, pagination *utils.Pagination) ([]model.ErshouShopFollower, int64, error)
	ListUserFollowing(userID uint, pagination *utils.Pagination) ([]model.ErshouShopFollower, int64, error)
	IncrFollowerCount(shopID uint, delta int) error
}

type shopFollowerRepository struct {
	db *gorm.DB
}

// NewShopFollowerRepository 创建店铺粉丝仓储实例
func NewShopFollowerRepository(db *gorm.DB) ShopFollowerRepository {
	return &shopFollowerRepository{db: db}
}

func (r *shopFollowerRepository) Follow(f *model.ErshouShopFollower) error {
	return r.db.Create(f).Error
}

func (r *shopFollowerRepository) Unfollow(shopID, userID uint) error {
	return r.db.Where("shop_id = ? AND user_id = ?", shopID, userID).
		Delete(&model.ErshouShopFollower{}).Error
}

func (r *shopFollowerRepository) IsFollowing(shopID, userID uint) (bool, error) {
	var count int64
	err := r.db.Model(&model.ErshouShopFollower{}).
		Where("shop_id = ? AND user_id = ?", shopID, userID).
		Count(&count).Error
	return count > 0, err
}

func (r *shopFollowerRepository) ListFollowers(shopID uint, pagination *utils.Pagination) ([]model.ErshouShopFollower, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 20)
	}
	var list []model.ErshouShopFollower
	var total int64
	q := r.db.Model(&model.ErshouShopFollower{}).Where("shop_id = ?", shopID)
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := q.Scopes(utils.Paginate(pagination)).
		Order("created_at DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *shopFollowerRepository) ListUserFollowing(userID uint, pagination *utils.Pagination) ([]model.ErshouShopFollower, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 20)
	}
	var list []model.ErshouShopFollower
	var total int64
	q := r.db.Model(&model.ErshouShopFollower{}).Where("user_id = ?", userID)
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := q.Scopes(utils.Paginate(pagination)).
		Order("created_at DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *shopFollowerRepository) IncrFollowerCount(shopID uint, delta int) error {
	return r.db.Model(&model.ErshouShop{}).Where("id = ?", shopID).
		UpdateColumn("follower_count", gorm.Expr("follower_count + ?", delta)).Error
}
