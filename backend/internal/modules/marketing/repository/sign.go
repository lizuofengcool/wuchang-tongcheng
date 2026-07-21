// Package repository 营销活动中台数据访问层 - 签到（sign 子域）
package repository

import (
	"time"

	"wuchang-tongcheng/internal/modules/marketing/model"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

// SignRepository 签到仓储接口
type SignRepository interface {
	// 签到记录
	CreateRecord(r *model.SignRecord) error
	FindTodayRecord(userID uint, date time.Time) (*model.SignRecord, error)
	FindLatestRecord(userID uint) (*model.SignRecord, error)
	ListRecordsByMonth(userID uint, start, end time.Time) ([]model.SignRecord, error)
	SumPoints(userID uint) (int, error)

	// 签到规则
	CreateRule(rule *model.SignRule) error
	FindRuleByID(id uint) (*model.SignRule, error)
	FindRuleByDay(day int) (*model.SignRule, error)
	UpdateRule(id uint, fields map[string]interface{}) error
	DeleteRule(id uint) error
	ListRules(query SignRuleListQuery, pagination *utils.Pagination) ([]model.SignRule, int64, error)
	ListEnabledRules() ([]model.SignRule, error)
}

// SignRuleListQuery 签到规则列表查询
type SignRuleListQuery struct {
	Status *int
}

type signRepository struct {
	db *gorm.DB
}

// NewSignRepository 创建签到仓储实例
func NewSignRepository(db *gorm.DB) SignRepository {
	return &signRepository{db: db}
}

// ===== 签到记录 =====

func (r *signRepository) CreateRecord(rec *model.SignRecord) error {
	return r.db.Create(rec).Error
}

func (r *signRepository) FindTodayRecord(userID uint, date time.Time) (*model.SignRecord, error) {
	var rec model.SignRecord
	// date 为当日 00:00，查询区间 [date, date+24h)
	next := date.AddDate(0, 0, 1)
	if err := r.db.Where("user_id = ? AND sign_date >= ? AND sign_date < ?", userID, date, next).
		First(&rec).Error; err != nil {
		return nil, err
	}
	return &rec, nil
}

func (r *signRepository) FindLatestRecord(userID uint) (*model.SignRecord, error) {
	var rec model.SignRecord
	if err := r.db.Where("user_id = ?", userID).Order("sign_date DESC, id DESC").First(&rec).Error; err != nil {
		return nil, err
	}
	return &rec, nil
}

func (r *signRepository) ListRecordsByMonth(userID uint, start, end time.Time) ([]model.SignRecord, error) {
	var list []model.SignRecord
	err := r.db.Where("user_id = ? AND sign_date >= ? AND sign_date < ?", userID, start, end).
		Order("sign_date ASC, id ASC").Find(&list).Error
	return list, err
}

func (r *signRepository) SumPoints(userID uint) (int, error) {
	var sum int
	err := r.db.Model(&model.SignRecord{}).
		Where("user_id = ?", userID).
		Select("COALESCE(SUM(points), 0)").Scan(&sum).Error
	return sum, err
}

// ===== 签到规则 =====

func (r *signRepository) CreateRule(rule *model.SignRule) error {
	return r.db.Create(rule).Error
}

func (r *signRepository) FindRuleByID(id uint) (*model.SignRule, error) {
	var rule model.SignRule
	if err := r.db.First(&rule, id).Error; err != nil {
		return nil, err
	}
	return &rule, nil
}

func (r *signRepository) FindRuleByDay(day int) (*model.SignRule, error) {
	var rule model.SignRule
	if err := r.db.Where("day = ?", day).First(&rule).Error; err != nil {
		return nil, err
	}
	return &rule, nil
}

func (r *signRepository) UpdateRule(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.SignRule{}).Where("id = ?", id).Updates(fields).Error
}

func (r *signRepository) DeleteRule(id uint) error {
	return r.db.Delete(&model.SignRule{}, id).Error
}

func (r *signRepository) ListRules(query SignRuleListQuery, pagination *utils.Pagination) ([]model.SignRule, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.SignRule
	var total int64

	q := r.db.Model(&model.SignRule{})
	if query.Status != nil {
		q = q.Where("status = ?", *query.Status)
	}

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := q.Scopes(utils.Paginate(pagination)).
		Order("day ASC, id ASC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *signRepository) ListEnabledRules() ([]model.SignRule, error) {
	var list []model.SignRule
	err := r.db.Where("status = ?", model.SignRuleStatusEnabled).
		Order("day ASC").Find(&list).Error
	return list, err
}
