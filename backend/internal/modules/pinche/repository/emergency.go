// Package repository 同城拼车出行数据访问层 - 紧急联系人/一键报警
package repository

import (
	"wuchang-tongcheng/internal/modules/pinche/model"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

// EmergencyListOptions 紧急联系人/报警列表过滤条件
type EmergencyListOptions struct {
	UserID      uint
	PincheID    uint
	TripID      uint
	AlertType   string
	AlertStatus *int
}

// EmergencyRepository 紧急联系人/报警仓储接口
type EmergencyRepository interface {
	Create(e *model.PincheEmergency) error
	FindByID(id uint) (*model.PincheEmergency, error)
	Update(id uint, fields map[string]interface{}) error
	Delete(id uint) error

	// 紧急联系人列表
	ListContacts(userID uint, pagination *utils.Pagination) ([]model.PincheEmergency, int64, error)
	// 报警列表
	ListAlerts(regionID uint, pagination *utils.Pagination, opts EmergencyListOptions) ([]model.PincheEmergency, int64, error)
	// 按行程查询
	ListByPinche(pincheID uint, pagination *utils.Pagination) ([]model.PincheEmergency, int64, error)
	ListByTrip(tripID uint, pagination *utils.Pagination) ([]model.PincheEmergency, int64, error)

	UpdateAlertStatus(id uint, status int) error
	UpdateHandleResult(id uint, handlerID uint, result string) error
	CountAlertsByStatus(regionID uint, status int) (int64, error)
}

type emergencyRepository struct {
	db *gorm.DB
}

// NewEmergencyRepository 创建紧急联系人/报警仓储实例
func NewEmergencyRepository(db *gorm.DB) EmergencyRepository {
	return &emergencyRepository{db: db}
}

func (r *emergencyRepository) Create(e *model.PincheEmergency) error {
	return r.db.Create(e).Error
}

func (r *emergencyRepository) FindByID(id uint) (*model.PincheEmergency, error) {
	var e model.PincheEmergency
	if err := r.db.First(&e, id).Error; err != nil {
		return nil, err
	}
	return &e, nil
}

func (r *emergencyRepository) Update(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.PincheEmergency{}).Where("id = ?", id).Updates(fields).Error
}

func (r *emergencyRepository) Delete(id uint) error {
	return r.db.Delete(&model.PincheEmergency{}, id).Error
}

func (r *emergencyRepository) ListContacts(userID uint, pagination *utils.Pagination) ([]model.PincheEmergency, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.PincheEmergency
	var total int64

	// 紧急联系人：alert_type 为空，contact_name 不为空
	query := r.db.Model(&model.PincheEmergency{}).
		Where("user_id = ? AND contact_name != '' AND alert_time IS NULL", userID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(pagination)).
		Order("is_primary DESC, created_at DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *emergencyRepository) ListAlerts(regionID uint, pagination *utils.Pagination, opts EmergencyListOptions) ([]model.PincheEmergency, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.PincheEmergency
	var total int64

	query := r.db.Model(&model.PincheEmergency{}).Where("alert_time IS NOT NULL")
	if regionID > 0 {
		query = query.Where("region_id = ?", regionID)
	}
	if opts.UserID > 0 {
		query = query.Where("user_id = ?", opts.UserID)
	}
	if opts.PincheID > 0 {
		query = query.Where("pinche_id = ?", opts.PincheID)
	}
	if opts.TripID > 0 {
		query = query.Where("trip_id = ?", opts.TripID)
	}
	if opts.AlertType != "" {
		query = query.Where("alert_type = ?", opts.AlertType)
	}
	if opts.AlertStatus != nil {
		query = query.Where("alert_status = ?", *opts.AlertStatus)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(pagination)).
		Order("alert_time DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *emergencyRepository) ListByPinche(pincheID uint, pagination *utils.Pagination) ([]model.PincheEmergency, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.PincheEmergency
	var total int64

	query := r.db.Model(&model.PincheEmergency{}).Where("pinche_id = ?", pincheID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(pagination)).
		Order("created_at DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *emergencyRepository) ListByTrip(tripID uint, pagination *utils.Pagination) ([]model.PincheEmergency, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.PincheEmergency
	var total int64

	query := r.db.Model(&model.PincheEmergency{}).Where("trip_id = ?", tripID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(pagination)).
		Order("created_at DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *emergencyRepository) UpdateAlertStatus(id uint, status int) error {
	return r.db.Model(&model.PincheEmergency{}).Where("id = ?", id).Update("alert_status", status).Error
}

func (r *emergencyRepository) UpdateHandleResult(id uint, handlerID uint, result string) error {
	return r.db.Model(&model.PincheEmergency{}).Where("id = ?", id).
		Updates(map[string]interface{}{"handler_id": handlerID, "handle_result": result}).Error
}

func (r *emergencyRepository) CountAlertsByStatus(regionID uint, status int) (int64, error) {
	var count int64
	q := r.db.Model(&model.PincheEmergency{}).Where("alert_status = ? AND alert_time IS NOT NULL", status)
	if regionID > 0 {
		q = q.Where("region_id = ?", regionID)
	}
	if err := q.Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}
