// Package repository 同城114数据访问层 - 电话拨打记录
// 一键拨号核心数据，统计拨打次数与设备信息
package repository

import (
	"wuchang-tongcheng/internal/modules/dh114/model"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

// PhoneCallRepository 电话拨打记录仓储接口
type PhoneCallRepository interface {
	Create(c *model.Dh114PhoneCall) error
	FindByID(id uint) (*model.Dh114PhoneCall, error)
	List(query PhoneCallListQuery, pagination *utils.Pagination) ([]model.Dh114PhoneCall, int64, error)
	ListByDh114(dh114ID uint, pagination *utils.Pagination) ([]model.Dh114PhoneCall, int64, error)
	ListByCaller(callerID uint, pagination *utils.Pagination) ([]model.Dh114PhoneCall, int64, error)
	CountByDh114(dh114ID uint) (int64, error)
	CountTodayByDh114(dh114ID uint) (int64, error)
}

// PhoneCallListQuery 电话拨打列表查询
type PhoneCallListQuery struct {
	Dh114ID   uint
	CallerID  uint
	CallType  string
	Status    string
	Device    string
}

type phoneCallRepository struct {
	db *gorm.DB
}

// NewPhoneCallRepository 创建电话拨打记录仓储实例
func NewPhoneCallRepository(db *gorm.DB) PhoneCallRepository {
	return &phoneCallRepository{db: db}
}

func (r *phoneCallRepository) Create(c *model.Dh114PhoneCall) error {
	return r.db.Create(c).Error
}

func (r *phoneCallRepository) FindByID(id uint) (*model.Dh114PhoneCall, error) {
	var c model.Dh114PhoneCall
	if err := r.db.First(&c, id).Error; err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *phoneCallRepository) List(query PhoneCallListQuery, pagination *utils.Pagination) ([]model.Dh114PhoneCall, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 20)
	}
	var list []model.Dh114PhoneCall
	var total int64

	q := r.db.Model(&model.Dh114PhoneCall{})
	if query.Dh114ID > 0 {
		q = q.Where("dh114_id = ?", query.Dh114ID)
	}
	if query.CallerID > 0 {
		q = q.Where("caller_id = ?", query.CallerID)
	}
	if query.CallType != "" {
		q = q.Where("call_type = ?", query.CallType)
	}
	if query.Status != "" {
		q = q.Where("status = ?", query.Status)
	}
	if query.Device != "" {
		q = q.Where("device = ?", query.Device)
	}

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := q.Scopes(utils.Paginate(pagination)).
		Order("called_at DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *phoneCallRepository) ListByDh114(dh114ID uint, pagination *utils.Pagination) ([]model.Dh114PhoneCall, int64, error) {
	return r.List(PhoneCallListQuery{Dh114ID: dh114ID}, pagination)
}

func (r *phoneCallRepository) ListByCaller(callerID uint, pagination *utils.Pagination) ([]model.Dh114PhoneCall, int64, error) {
	return r.List(PhoneCallListQuery{CallerID: callerID}, pagination)
}

func (r *phoneCallRepository) CountByDh114(dh114ID uint) (int64, error) {
	var count int64
	if err := r.db.Model(&model.Dh114PhoneCall{}).
		Where("dh114_id = ? AND status = ?", dh114ID, model.CallStatusSuccess).
		Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (r *phoneCallRepository) CountTodayByDh114(dh114ID uint) (int64, error) {
	var count int64
	if err := r.db.Model(&model.Dh114PhoneCall{}).
		Where("dh114_id = ? AND status = ? AND called_at::date = CURRENT_DATE", dh114ID, model.CallStatusSuccess).
		Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}
