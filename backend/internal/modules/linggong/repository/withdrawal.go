// Package repository 同城零工兼职数据访问层 - 提现
package repository

import (
	"wuchang-tongcheng/internal/modules/linggong/model"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

// WithdrawalRepository 提现仓储接口
type WithdrawalRepository interface {
	Create(w *model.LinggongWithdrawal) error
	FindByID(id uint) (*model.LinggongWithdrawal, error)
	FindByWithdrawalNo(no string) (*model.LinggongWithdrawal, error)
	Update(id uint, fields map[string]interface{}) error
	Delete(id uint) error

	List(regionID uint, pagination *utils.Pagination, opts WithdrawalListOptions) ([]model.LinggongWithdrawal, int64, error)
	ListByUser(userID uint, pagination *utils.Pagination) ([]model.LinggongWithdrawal, int64, error)
}

// WithdrawalListOptions 提现列表过滤条件
type WithdrawalListOptions struct {
	UserID   uint
	UserType string
	Status   *int
	Method   string
	Keyword  string
}

type withdrawalRepository struct {
	db *gorm.DB
}

// NewWithdrawalRepository 创建提现仓储实例
func NewWithdrawalRepository(db *gorm.DB) WithdrawalRepository {
	return &withdrawalRepository{db: db}
}

func (r *withdrawalRepository) Create(w *model.LinggongWithdrawal) error {
	return r.db.Create(w).Error
}

func (r *withdrawalRepository) FindByID(id uint) (*model.LinggongWithdrawal, error) {
	var w model.LinggongWithdrawal
	if err := r.db.First(&w, id).Error; err != nil {
		return nil, err
	}
	return &w, nil
}

func (r *withdrawalRepository) FindByWithdrawalNo(no string) (*model.LinggongWithdrawal, error) {
	var w model.LinggongWithdrawal
	if err := r.db.Where("withdrawal_no = ?", no).First(&w).Error; err != nil {
		return nil, err
	}
	return &w, nil
}

func (r *withdrawalRepository) Update(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.LinggongWithdrawal{}).Where("id = ?", id).Updates(fields).Error
}

func (r *withdrawalRepository) Delete(id uint) error {
	return r.db.Delete(&model.LinggongWithdrawal{}, id).Error
}

func (r *withdrawalRepository) List(regionID uint, pagination *utils.Pagination, opts WithdrawalListOptions) ([]model.LinggongWithdrawal, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.LinggongWithdrawal
	var total int64

	query := r.db.Model(&model.LinggongWithdrawal{})
	if regionID > 0 {
		query = query.Where("region_id = ?", regionID)
	}
	if opts.UserID > 0 {
		query = query.Where("user_id = ?", opts.UserID)
	}
	if opts.UserType != "" {
		query = query.Where("user_type = ?", opts.UserType)
	}
	if opts.Status != nil {
		query = query.Where("status = ?", *opts.Status)
	}
	if opts.Method != "" {
		query = query.Where("method = ?", opts.Method)
	}
	if opts.Keyword != "" {
		like := "%" + opts.Keyword + "%"
		query = query.Where("withdrawal_no ILIKE ? OR user_name ILIKE ?", like, like)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(pagination)).
		Order("created_at DESC, id DESC").
		Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *withdrawalRepository) ListByUser(userID uint, pagination *utils.Pagination) ([]model.LinggongWithdrawal, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.LinggongWithdrawal
	var total int64
	query := r.db.Model(&model.LinggongWithdrawal{}).Where("user_id = ?", userID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(pagination)).
		Order("created_at DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}
