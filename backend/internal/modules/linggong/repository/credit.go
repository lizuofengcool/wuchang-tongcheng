// Package repository 同城零工兼职数据访问层 - 信用分
package repository

import (
	"wuchang-tongcheng/internal/modules/linggong/model"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

// CreditRepository 信用分仓储接口
type CreditRepository interface {
	Create(c *model.LinggongCredit) error
	FindByID(id uint) (*model.LinggongCredit, error)
	Update(id uint, fields map[string]interface{}) error
	Delete(id uint) error

	List(regionID uint, pagination *utils.Pagination, opts CreditListOptions) ([]model.LinggongCredit, int64, error)
	ListByUser(userID uint, userType string, pagination *utils.Pagination) ([]model.LinggongCredit, int64, error)
	GetLatestScore(userID uint, userType string) (int, error)
}

// CreditListOptions 信用分列表过滤条件
type CreditListOptions struct {
	UserID     uint
	UserType   string
	Reason     string
	ChangeType string
}

type creditRepository struct {
	db *gorm.DB
}

// NewCreditRepository 创建信用分仓储实例
func NewCreditRepository(db *gorm.DB) CreditRepository {
	return &creditRepository{db: db}
}

func (r *creditRepository) Create(c *model.LinggongCredit) error {
	return r.db.Create(c).Error
}

func (r *creditRepository) FindByID(id uint) (*model.LinggongCredit, error) {
	var c model.LinggongCredit
	if err := r.db.First(&c, id).Error; err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *creditRepository) Update(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.LinggongCredit{}).Where("id = ?", id).Updates(fields).Error
}

func (r *creditRepository) Delete(id uint) error {
	return r.db.Delete(&model.LinggongCredit{}, id).Error
}

func (r *creditRepository) List(regionID uint, pagination *utils.Pagination, opts CreditListOptions) ([]model.LinggongCredit, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.LinggongCredit
	var total int64

	query := r.db.Model(&model.LinggongCredit{})
	if regionID > 0 {
		query = query.Where("region_id = ?", regionID)
	}
	if opts.UserID > 0 {
		query = query.Where("user_id = ?", opts.UserID)
	}
	if opts.UserType != "" {
		query = query.Where("user_type = ?", opts.UserType)
	}
	if opts.Reason != "" {
		query = query.Where("reason = ?", opts.Reason)
	}
	if opts.ChangeType != "" {
		query = query.Where("change_type = ?", opts.ChangeType)
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

func (r *creditRepository) ListByUser(userID uint, userType string, pagination *utils.Pagination) ([]model.LinggongCredit, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.LinggongCredit
	var total int64
	query := r.db.Model(&model.LinggongCredit{}).Where("user_id = ? AND user_type = ?", userID, userType)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(pagination)).
		Order("created_at DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *creditRepository) GetLatestScore(userID uint, userType string) (int, error) {
	var c model.LinggongCredit
	err := r.db.Where("user_id = ? AND user_type = ?", userID, userType).
		Order("created_at DESC, id DESC").First(&c).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return 100, nil // 默认信用分 100
		}
		return 0, err
	}
	return c.AfterScore, nil
}
