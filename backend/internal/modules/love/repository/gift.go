// Package repository love 相亲交友数据访问层 - 礼物
package repository

import (
	"wuchang-tongcheng/internal/modules/love/model"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

// LoveGiftRepository 礼物仓储接口
type LoveGiftRepository interface {
	Create(g *model.LoveGift) error
	FindByID(id uint) (*model.LoveGift, error)
	FindByCode(code string) (*model.LoveGift, error)
	Update(g *model.LoveGift) error
	UpdateFields(id uint, fields map[string]interface{}) error
	Delete(id uint) error
	List(pagination *utils.Pagination, opts LoveGiftListOptions) ([]model.LoveGift, int64, error)
	ListAvailable(memberLevel int) ([]model.LoveGift, error)
	BatchUpdateStatus(ids []uint, status int) error
}

// LoveGiftListOptions 礼物列表过滤
type LoveGiftListOptions struct {
	Category    string
	MemberLevel *int
	Status      *int
}

// LoveGiftRecordRepository 送礼记录仓储接口（love_gift_records 表如有则用，本实现复用 gift 主表，预留扩展）
type LoveGiftRecordRepository interface {
	Create(record interface{}) error
	ListByFromUser(userID uint, pagination *utils.Pagination) ([]model.LoveGift, int64, error)
	ListByToUser(userID uint, pagination *utils.Pagination) ([]model.LoveGift, int64, error)
	CountTodayByUser(userID uint) (int64, error)
}

type loveGiftRepository struct {
	db *gorm.DB
}

// NewLoveGiftRepository 创建礼物仓储
func NewLoveGiftRepository(db *gorm.DB) LoveGiftRepository {
	return &loveGiftRepository{db: db}
}

func (r *loveGiftRepository) Create(g *model.LoveGift) error {
	return r.db.Create(g).Error
}

func (r *loveGiftRepository) FindByID(id uint) (*model.LoveGift, error) {
	var g model.LoveGift
	if err := r.db.First(&g, id).Error; err != nil {
		return nil, err
	}
	return &g, nil
}

func (r *loveGiftRepository) FindByCode(code string) (*model.LoveGift, error) {
	var g model.LoveGift
	if err := r.db.Where("gift_code = ?", code).First(&g).Error; err != nil {
		return nil, err
	}
	return &g, nil
}

func (r *loveGiftRepository) Update(g *model.LoveGift) error {
	return r.db.Save(g).Error
}

func (r *loveGiftRepository) UpdateFields(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.LoveGift{}).Where("id = ?", id).Updates(fields).Error
}

func (r *loveGiftRepository) Delete(id uint) error {
	return r.db.Delete(&model.LoveGift{}, id).Error
}

func (r *loveGiftRepository) List(pagination *utils.Pagination, opts LoveGiftListOptions) ([]model.LoveGift, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.LoveGift
	var total int64

	query := r.db.Model(&model.LoveGift{})
	if opts.Category != "" {
		query = query.Where("category = ?", opts.Category)
	}
	if opts.MemberLevel != nil {
		query = query.Where("member_level <= ?", *opts.MemberLevel)
	}
	if opts.Status != nil {
		query = query.Where("status = ?", *opts.Status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(pagination)).Order("sort ASC, id ASC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *loveGiftRepository) ListAvailable(memberLevel int) ([]model.LoveGift, error) {
	var list []model.LoveGift
	err := r.db.Where("status = ? AND member_level <= ?", 1, memberLevel).Order("sort ASC, id ASC").Find(&list).Error
	return list, err
}

func (r *loveGiftRepository) BatchUpdateStatus(ids []uint, status int) error {
	return r.db.Model(&model.LoveGift{}).Where("id IN ?", ids).Update("status", status).Error
}

type loveGiftRecordRepository struct {
	db *gorm.DB
}

// NewLoveGiftRecordRepository 创建送礼记录仓储
func NewLoveGiftRecordRepository(db *gorm.DB) LoveGiftRecordRepository {
	return &loveGiftRecordRepository{db: db}
}

func (r *loveGiftRecordRepository) Create(record interface{}) error {
	return r.db.Create(record).Error
}

func (r *loveGiftRecordRepository) ListByFromUser(userID uint, pagination *utils.Pagination) ([]model.LoveGift, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.LoveGift
	var total int64
	query := r.db.Model(&model.LoveGift{})
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(pagination)).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *loveGiftRecordRepository) ListByToUser(userID uint, pagination *utils.Pagination) ([]model.LoveGift, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.LoveGift
	var total int64
	query := r.db.Model(&model.LoveGift{})
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(pagination)).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *loveGiftRecordRepository) CountTodayByUser(userID uint) (int64, error) {
	var count int64
	err := r.db.Model(&model.LoveGift{}).Count(&count).Error
	return count, err
}
