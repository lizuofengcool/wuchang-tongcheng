// Package repository 多租户分站数据访问层 - 员工
package repository

import (
	"wuchang-tongcheng/internal/modules/tenant/model"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

// StaffListOptions 员工列表过滤条件
type StaffListOptions struct {
	StationID uint
	UserID    uint
	Role      string
	Status    *int
}

// StaffRepository 员工仓储接口
type StaffRepository interface {
	Create(s *model.Staff) error
	FindByID(id uint) (*model.Staff, error)
	FindByStationAndUser(stationID, userID uint) (*model.Staff, error)
	Update(s *model.Staff) error
	UpdateFields(id uint, fields map[string]interface{}) error
	Delete(id uint) error
	List(pagination *utils.Pagination, opts StaffListOptions) ([]model.Staff, int64, error)
	ListByStation(stationID uint) ([]model.Staff, error)
	ListByUser(userID uint) ([]model.Staff, error)
}

type staffRepository struct {
	db *gorm.DB
}

// NewStaffRepository 创建员工仓储实例
func NewStaffRepository(db *gorm.DB) StaffRepository {
	return &staffRepository{db: db}
}

func (r *staffRepository) Create(s *model.Staff) error {
	return r.db.Create(s).Error
}

func (r *staffRepository) FindByID(id uint) (*model.Staff, error) {
	var s model.Staff
	if err := r.db.First(&s, id).Error; err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *staffRepository) FindByStationAndUser(stationID, userID uint) (*model.Staff, error) {
	var s model.Staff
	if err := r.db.Where("station_id = ? AND user_id = ?", stationID, userID).First(&s).Error; err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *staffRepository) Update(s *model.Staff) error {
	return r.db.Save(s).Error
}

func (r *staffRepository) UpdateFields(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.Staff{}).Where("id = ?", id).Updates(fields).Error
}

func (r *staffRepository) Delete(id uint) error {
	return r.db.Delete(&model.Staff{}, id).Error
}

func (r *staffRepository) List(pagination *utils.Pagination, opts StaffListOptions) ([]model.Staff, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.Staff
	var total int64

	query := r.db.Model(&model.Staff{})
	if opts.StationID > 0 {
		query = query.Where("station_id = ?", opts.StationID)
	}
	if opts.UserID > 0 {
		query = query.Where("user_id = ?", opts.UserID)
	}
	if opts.Role != "" {
		query = query.Where("role = ?", opts.Role)
	}
	if opts.Status != nil {
		query = query.Where("status = ?", *opts.Status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(pagination)).
		Order("created_at DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *staffRepository) ListByStation(stationID uint) ([]model.Staff, error) {
	var list []model.Staff
	if err := r.db.Where("station_id = ?", stationID).
		Order("created_at DESC, id DESC").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *staffRepository) ListByUser(userID uint) ([]model.Staff, error) {
	var list []model.Staff
	if err := r.db.Where("user_id = ?", userID).
		Order("created_at DESC, id DESC").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}
