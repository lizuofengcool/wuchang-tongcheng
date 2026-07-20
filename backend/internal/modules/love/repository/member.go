// Package repository love 相亲交友数据访问层 - 会员
package repository

import (
	"wuchang-tongcheng/internal/modules/love/model"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

// LoveMemberLevelRepository 会员等级仓储接口
type LoveMemberLevelRepository interface {
	Create(l *model.LoveMemberLevel) error
	FindByID(id uint) (*model.LoveMemberLevel, error)
	FindByLevelCode(code string) (*model.LoveMemberLevel, error)
	FindByLevel(level int) (*model.LoveMemberLevel, error)
	Update(l *model.LoveMemberLevel) error
	UpdateFields(id uint, fields map[string]interface{}) error
	Delete(id uint) error
	List(pagination *utils.Pagination, opts LoveMemberLevelListOptions) ([]model.LoveMemberLevel, int64, error)
	ListAll() ([]model.LoveMemberLevel, error)
}

// LoveMemberLevelListOptions 会员等级列表过滤
type LoveMemberLevelListOptions struct {
	Status *int
}

// LoveMembershipRepository 会员订阅仓储接口
type LoveMembershipRepository interface {
	Create(m *model.LoveMembership) error
	FindByID(id uint) (*model.LoveMembership, error)
	FindBySubNo(no string) (*model.LoveMembership, error)
	FindByUserActive(userID uint) (*model.LoveMembership, error)
	Update(m *model.LoveMembership) error
	UpdateFields(id uint, fields map[string]interface{}) error
	Delete(id uint) error
	List(pagination *utils.Pagination, opts LoveMembershipListOptions) ([]model.LoveMembership, int64, error)
	ListByUser(userID uint, pagination *utils.Pagination) ([]model.LoveMembership, int64, error)
	Cancel(id uint, reason string) error
	Refund(id uint, amount float64, reason string) error
	MarkPaid(id uint, payMethod, payOrderNo string) error
	UpdateStatus(id uint, status int) error
	CountActiveByUser(userID uint) (int64, error)
}

// LoveMembershipListOptions 会员订阅列表过滤
type LoveMembershipListOptions struct {
	UserID    uint
	LoveID    uint
	LevelCode string
	Plan      string
	Status    *int
}

type loveMemberLevelRepository struct {
	db *gorm.DB
}

// NewLoveMemberLevelRepository 创建会员等级仓储
func NewLoveMemberLevelRepository(db *gorm.DB) LoveMemberLevelRepository {
	return &loveMemberLevelRepository{db: db}
}

func (r *loveMemberLevelRepository) Create(l *model.LoveMemberLevel) error {
	return r.db.Create(l).Error
}

func (r *loveMemberLevelRepository) FindByID(id uint) (*model.LoveMemberLevel, error) {
	var l model.LoveMemberLevel
	if err := r.db.First(&l, id).Error; err != nil {
		return nil, err
	}
	return &l, nil
}

func (r *loveMemberLevelRepository) FindByLevelCode(code string) (*model.LoveMemberLevel, error) {
	var l model.LoveMemberLevel
	if err := r.db.Where("level_code = ?", code).First(&l).Error; err != nil {
		return nil, err
	}
	return &l, nil
}

func (r *loveMemberLevelRepository) FindByLevel(level int) (*model.LoveMemberLevel, error) {
	var l model.LoveMemberLevel
	if err := r.db.Where("level = ?", level).First(&l).Error; err != nil {
		return nil, err
	}
	return &l, nil
}

func (r *loveMemberLevelRepository) Update(l *model.LoveMemberLevel) error {
	return r.db.Save(l).Error
}

func (r *loveMemberLevelRepository) UpdateFields(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.LoveMemberLevel{}).Where("id = ?", id).Updates(fields).Error
}

func (r *loveMemberLevelRepository) Delete(id uint) error {
	return r.db.Delete(&model.LoveMemberLevel{}, id).Error
}

func (r *loveMemberLevelRepository) List(pagination *utils.Pagination, opts LoveMemberLevelListOptions) ([]model.LoveMemberLevel, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.LoveMemberLevel
	var total int64

	query := r.db.Model(&model.LoveMemberLevel{})
	if opts.Status != nil {
		query = query.Where("status = ?", *opts.Status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(pagination)).Order("level ASC, id ASC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *loveMemberLevelRepository) ListAll() ([]model.LoveMemberLevel, error) {
	var list []model.LoveMemberLevel
	err := r.db.Where("status = ?", 1).Order("level ASC").Find(&list).Error
	return list, err
}

type loveMembershipRepository struct {
	db *gorm.DB
}

// NewLoveMembershipRepository 创建会员订阅仓储
func NewLoveMembershipRepository(db *gorm.DB) LoveMembershipRepository {
	return &loveMembershipRepository{db: db}
}

func (r *loveMembershipRepository) Create(m *model.LoveMembership) error {
	return r.db.Create(m).Error
}

func (r *loveMembershipRepository) FindByID(id uint) (*model.LoveMembership, error) {
	var m model.LoveMembership
	if err := r.db.First(&m, id).Error; err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *loveMembershipRepository) FindBySubNo(no string) (*model.LoveMembership, error) {
	var m model.LoveMembership
	if err := r.db.Where("sub_no = ?", no).First(&m).Error; err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *loveMembershipRepository) FindByUserActive(userID uint) (*model.LoveMembership, error) {
	var m model.LoveMembership
	err := r.db.Where("user_id = ? AND status = ? AND end_at > NOW()", userID, 1).First(&m).Error
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *loveMembershipRepository) Update(m *model.LoveMembership) error {
	return r.db.Save(m).Error
}

func (r *loveMembershipRepository) UpdateFields(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.LoveMembership{}).Where("id = ?", id).Updates(fields).Error
}

func (r *loveMembershipRepository) Delete(id uint) error {
	return r.db.Delete(&model.LoveMembership{}, id).Error
}

func (r *loveMembershipRepository) List(pagination *utils.Pagination, opts LoveMembershipListOptions) ([]model.LoveMembership, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.LoveMembership
	var total int64

	query := r.db.Model(&model.LoveMembership{})
	if opts.UserID > 0 {
		query = query.Where("user_id = ?", opts.UserID)
	}
	if opts.LoveID > 0 {
		query = query.Where("love_id = ?", opts.LoveID)
	}
	if opts.LevelCode != "" {
		query = query.Where("level_code = ?", opts.LevelCode)
	}
	if opts.Plan != "" {
		query = query.Where("plan = ?", opts.Plan)
	}
	if opts.Status != nil {
		query = query.Where("status = ?", *opts.Status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(pagination)).Order("id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *loveMembershipRepository) ListByUser(userID uint, pagination *utils.Pagination) ([]model.LoveMembership, int64, error) {
	return r.List(pagination, LoveMembershipListOptions{UserID: userID})
}

func (r *loveMembershipRepository) Cancel(id uint, reason string) error {
	return r.db.Model(&model.LoveMembership{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":        2,
		"cancel_at":     gorm.Expr("NOW()"),
		"cancel_reason": reason,
	}).Error
}

func (r *loveMembershipRepository) Refund(id uint, amount float64, reason string) error {
	return r.db.Model(&model.LoveMembership{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":         3,
		"refund_amount":  amount,
		"refund_at":      gorm.Expr("NOW()"),
		"refund_reason":  reason,
	}).Error
}

func (r *loveMembershipRepository) MarkPaid(id uint, payMethod, payOrderNo string) error {
	return r.db.Model(&model.LoveMembership{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":        1,
		"pay_method":    payMethod,
		"pay_order_no":  payOrderNo,
		"pay_at":        gorm.Expr("NOW()"),
	}).Error
}

func (r *loveMembershipRepository) UpdateStatus(id uint, status int) error {
	return r.db.Model(&model.LoveMembership{}).Where("id = ?", id).Update("status", status).Error
}

func (r *loveMembershipRepository) CountActiveByUser(userID uint) (int64, error) {
	var count int64
	err := r.db.Model(&model.LoveMembership{}).Where("user_id = ? AND status = ? AND end_at > NOW()", userID, 1).Count(&count).Error
	return count, err
}
