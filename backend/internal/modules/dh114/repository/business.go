// Package repository 同城114数据访问层 - 商户详细信息 + 营业时间
package repository

import (
	"wuchang-tongcheng/internal/modules/dh114/model"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

// ===== BusinessRepository 商户详细信息 =====

// BusinessRepository 商户详细信息仓储接口
type BusinessRepository interface {
	Create(b *model.Dh114Business) error
	FindByID(id uint) (*model.Dh114Business, error)
	FindByDh114ID(dh114ID uint) (*model.Dh114Business, error)
	Update(id uint, fields map[string]interface{}) error
	Delete(id uint) error
	List(query BusinessListQuery, pagination *utils.Pagination) ([]model.Dh114Business, int64, error)
	UpdateVerificationStatus(dh114ID uint, status int, verifiedAt interface{}) error
}

// BusinessListQuery 商户详情列表查询
type BusinessListQuery struct {
	Dh114ID            uint
	VerificationStatus *int
	Keyword            string
}

type businessRepository struct {
	db *gorm.DB
}

// NewBusinessRepository 创建商户详情仓储实例
func NewBusinessRepository(db *gorm.DB) BusinessRepository {
	return &businessRepository{db: db}
}

func (r *businessRepository) Create(b *model.Dh114Business) error {
	return r.db.Create(b).Error
}

func (r *businessRepository) FindByID(id uint) (*model.Dh114Business, error) {
	var b model.Dh114Business
	if err := r.db.First(&b, id).Error; err != nil {
		return nil, err
	}
	return &b, nil
}

func (r *businessRepository) FindByDh114ID(dh114ID uint) (*model.Dh114Business, error) {
	var b model.Dh114Business
	if err := r.db.Where("dh114_id = ?", dh114ID).First(&b).Error; err != nil {
		return nil, err
	}
	return &b, nil
}

func (r *businessRepository) Update(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.Dh114Business{}).Where("id = ?", id).Updates(fields).Error
}

func (r *businessRepository) Delete(id uint) error {
	return r.db.Delete(&model.Dh114Business{}, id).Error
}

func (r *businessRepository) List(query BusinessListQuery, pagination *utils.Pagination) ([]model.Dh114Business, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.Dh114Business
	var total int64

	q := r.db.Model(&model.Dh114Business{})
	if query.Dh114ID > 0 {
		q = q.Where("dh114_id = ?", query.Dh114ID)
	}
	if query.VerificationStatus != nil {
		q = q.Where("verification_status = ?", *query.VerificationStatus)
	}
	if query.Keyword != "" {
		like := "%" + query.Keyword + "%"
		q = q.Where("business_name ILIKE ? OR license_no ILIKE ? OR legal_person ILIKE ?", like, like, like)
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

func (r *businessRepository) UpdateVerificationStatus(dh114ID uint, status int, verifiedAt interface{}) error {
	fields := map[string]interface{}{
		"verification_status": status,
		"verified_at":         verifiedAt,
	}
	return r.db.Model(&model.Dh114Business{}).Where("dh114_id = ?", dh114ID).Updates(fields).Error
}

// ===== BusinessHourRepository 营业时间 =====

// BusinessHourRepository 营业时间仓储接口
type BusinessHourRepository interface {
	Create(h *model.Dh114BusinessHour) error
	BatchCreate(hours []model.Dh114BusinessHour) error
	FindByDh114ID(dh114ID uint) ([]model.Dh114BusinessHour, error)
	UpdateByDh114AndWeekday(dh114ID uint, weekday int, fields map[string]interface{}) error
	DeleteByDh114(dh114ID uint) error
	ReplaceHours(dh114ID uint, hours []model.Dh114BusinessHour) error
}

type businessHourRepository struct {
	db *gorm.DB
}

// NewBusinessHourRepository 创建营业时间仓储实例
func NewBusinessHourRepository(db *gorm.DB) BusinessHourRepository {
	return &businessHourRepository{db: db}
}

func (r *businessHourRepository) Create(h *model.Dh114BusinessHour) error {
	return r.db.Create(h).Error
}

func (r *businessHourRepository) BatchCreate(hours []model.Dh114BusinessHour) error {
	if len(hours) == 0 {
		return nil
	}
	return r.db.CreateInBatches(hours, len(hours)).Error
}

func (r *businessHourRepository) FindByDh114ID(dh114ID uint) ([]model.Dh114BusinessHour, error) {
	var list []model.Dh114BusinessHour
	if err := r.db.Where("dh114_id = ?", dh114ID).Order("weekday ASC").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *businessHourRepository) UpdateByDh114AndWeekday(dh114ID uint, weekday int, fields map[string]interface{}) error {
	return r.db.Model(&model.Dh114BusinessHour{}).
		Where("dh114_id = ? AND weekday = ?", dh114ID, weekday).
		Updates(fields).Error
}

func (r *businessHourRepository) DeleteByDh114(dh114ID uint) error {
	return r.db.Where("dh114_id = ?", dh114ID).Delete(&model.Dh114BusinessHour{}).Error
}

func (r *businessHourRepository) ReplaceHours(dh114ID uint, hours []model.Dh114BusinessHour) error {
	tx := r.db.Begin()
	if tx.Error != nil {
		return tx.Error
	}
	if err := tx.Where("dh114_id = ?", dh114ID).Delete(&model.Dh114BusinessHour{}).Error; err != nil {
		tx.Rollback()
		return err
	}
	for i := range hours {
		hours[i].Dh114ID = dh114ID
		if err := tx.Create(&hours[i]).Error; err != nil {
			tx.Rollback()
			return err
		}
	}
	return tx.Commit().Error
}
