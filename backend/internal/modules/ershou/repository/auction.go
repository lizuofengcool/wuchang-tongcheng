// Package repository 拍卖数据访问层
// 依据 v3.2.1 架构方案：对标闲鱼拍卖
package repository

import (
	"wuchang-tongcheng/internal/modules/ershou/model"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

// AuctionRepository 拍卖仓储接口
type AuctionRepository interface {
	Create(auction *model.ErshouAuction) error
	FindByID(id uint) (*model.ErshouAuction, error)
	FindByErshouID(ershouID uint) (*model.ErshouAuction, error)
	Update(id uint, fields map[string]interface{}) error
	List(query AuctionListQuery, pagination *utils.Pagination) ([]model.ErshouAuction, int64, error)
	ListActive(regionID uint) ([]model.ErshouAuction, error)
	ListEndingSoon(minutes int) ([]model.ErshouAuction, error)

	// 出价记录
	CreateBid(bid *model.ErshouAuctionBid) error
	ListBids(auctionID uint, limit int) ([]model.ErshouAuctionBid, error)
	ListUserBids(userID uint, pagination *utils.Pagination) ([]model.ErshouAuctionBid, int64, error)
	HasUserBid(auctionID, userID uint) (bool, error)
}

// AuctionListQuery 拍卖列表查询
type AuctionListQuery struct {
	ErshouID uint
	Status   *int
	UserID   uint
	RegionID uint
}

type auctionRepository struct {
	db *gorm.DB
}

// NewAuctionRepository 创建拍卖仓储实例
func NewAuctionRepository(db *gorm.DB) AuctionRepository {
	return &auctionRepository{db: db}
}

func (r *auctionRepository) Create(auction *model.ErshouAuction) error {
	return r.db.Create(auction).Error
}

func (r *auctionRepository) FindByID(id uint) (*model.ErshouAuction, error) {
	var a model.ErshouAuction
	if err := r.db.First(&a, id).Error; err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *auctionRepository) FindByErshouID(ershouID uint) (*model.ErshouAuction, error) {
	var a model.ErshouAuction
	if err := r.db.Where("ershou_id = ?", ershouID).First(&a).Error; err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *auctionRepository) Update(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.ErshouAuction{}).Where("id = ?", id).Updates(fields).Error
}

func (r *auctionRepository) List(query AuctionListQuery, pagination *utils.Pagination) ([]model.ErshouAuction, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.ErshouAuction
	var total int64

	q := r.db.Model(&model.ErshouAuction{})
	if query.ErshouID > 0 {
		q = q.Where("ershou_id = ?", query.ErshouID)
	}
	if query.Status != nil {
		q = q.Where("status = ?", *query.Status)
	}
	if query.RegionID > 0 {
		q = q.Where("region_id = ?", query.RegionID)
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

func (r *auctionRepository) ListActive(regionID uint) ([]model.ErshouAuction, error) {
	var list []model.ErshouAuction
	q := r.db.Where("status = ?", model.AuctionStatusActive)
	if regionID > 0 {
		q = q.Where("region_id = ?", regionID)
	}
	if err := q.Order("end_time ASC").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *auctionRepository) ListEndingSoon(minutes int) ([]model.ErshouAuction, error) {
	var list []model.ErshouAuction
	// 查找状态为进行中、且截拍时间在 N 分钟内的拍卖
	if err := r.db.Where("status = ? AND end_time <= NOW() + INTERVAL '?' MINUTES",
		model.AuctionStatusActive, minutes).
		Order("end_time ASC").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *auctionRepository) CreateBid(bid *model.ErshouAuctionBid) error {
	return r.db.Create(bid).Error
}

func (r *auctionRepository) ListBids(auctionID uint, limit int) ([]model.ErshouAuctionBid, error) {
	var list []model.ErshouAuctionBid
	q := r.db.Where("auction_id = ? AND is_invalid = false", auctionID)
	if limit > 0 {
		q = q.Limit(limit)
	}
	if err := q.Order("bid_price DESC, bid_time DESC").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *auctionRepository) ListUserBids(userID uint, pagination *utils.Pagination) ([]model.ErshouAuctionBid, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.ErshouAuctionBid
	var total int64
	q := r.db.Model(&model.ErshouAuctionBid{}).Where("user_id = ?", userID)
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := q.Scopes(utils.Paginate(pagination)).
		Order("bid_time DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *auctionRepository) HasUserBid(auctionID, userID uint) (bool, error) {
	var count int64
	err := r.db.Model(&model.ErshouAuctionBid{}).
		Where("auction_id = ? AND user_id = ? AND is_invalid = false", auctionID, userID).
		Count(&count).Error
	return count > 0, err
}
