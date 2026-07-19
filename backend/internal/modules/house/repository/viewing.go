// Package repository 看房预约数据访问层
package repository

import (
	"wuchang-tongcheng/internal/modules/house/model"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

// ViewingRepository 看房预约仓储接口
type ViewingRepository interface {
	Create(v *model.HouseViewing) error
	FindByID(id uint) (*model.HouseViewing, error)
	FindByViewingNo(no string) (*model.HouseViewing, error)
	Update(v *model.HouseViewing) error
	UpdateFields(id uint, fields map[string]interface{}) error
	Delete(id uint) error

	List(regionID uint, req *utils.Pagination, opts ViewingListOptions) ([]model.HouseViewing, int64, error)
	AdminList(req *utils.Pagination, opts ViewingAdminListOptions) ([]model.HouseViewing, int64, error)
	ListByUser(userID uint, req *utils.Pagination) ([]model.HouseViewing, int64, error)
	ListByAgent(agentID uint, req *utils.Pagination) ([]model.HouseViewing, int64, error)
	ListByHouse(houseID uint, req *utils.Pagination) ([]model.HouseViewing, int64, error)

	// 状态机
	MarkReminderSent(id uint) error
	UpdateReminderSentAt(id uint) error
}

// ViewingListOptions C 端列表过滤条件
type ViewingListOptions struct {
	HouseID     uint
	ListingID   uint
	AgentID     uint
	UserID      uint
	ViewingType string
	Status      *int
	Result      string
	StartDate   string
	EndDate     string
}

// ViewingAdminListOptions M 端管理列表过滤条件
type ViewingAdminListOptions struct {
	RegionID uint
	HouseID  uint
	UserID   uint
	AgentID  uint
	Status   *int
}

type viewingRepository struct {
	db *gorm.DB
}

// NewViewingRepository 创建仓储实例
func NewViewingRepository(db *gorm.DB) ViewingRepository {
	return &viewingRepository{db: db}
}

func (r *viewingRepository) Create(v *model.HouseViewing) error {
	return r.db.Create(v).Error
}

func (r *viewingRepository) FindByID(id uint) (*model.HouseViewing, error) {
	var v model.HouseViewing
	if err := r.db.First(&v, id).Error; err != nil {
		return nil, err
	}
	return &v, nil
}

func (r *viewingRepository) FindByViewingNo(no string) (*model.HouseViewing, error) {
	var v model.HouseViewing
	if err := r.db.Where("viewing_no = ?", no).First(&v).Error; err != nil {
		return nil, err
	}
	return &v, nil
}

func (r *viewingRepository) Update(v *model.HouseViewing) error {
	return r.db.Save(v).Error
}

func (r *viewingRepository) UpdateFields(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.HouseViewing{}).Where("id = ?", id).Updates(fields).Error
}

func (r *viewingRepository) Delete(id uint) error {
	return r.db.Delete(&model.HouseViewing{}, id).Error
}

func (r *viewingRepository) List(regionID uint, req *utils.Pagination, opts ViewingListOptions) ([]model.HouseViewing, int64, error) {
	if req == nil {
		req = utils.NewPagination(1, 10)
	}
	var list []model.HouseViewing
	var total int64

	query := r.db.Model(&model.HouseViewing{})
	if regionID > 0 {
		query = query.Where("region_id = ?", regionID)
	}
	if opts.HouseID > 0 {
		query = query.Where("house_id = ?", opts.HouseID)
	}
	if opts.ListingID > 0 {
		query = query.Where("listing_id = ?", opts.ListingID)
	}
	if opts.AgentID > 0 {
		query = query.Where("agent_id = ?", opts.AgentID)
	}
	if opts.UserID > 0 {
		query = query.Where("user_id = ?", opts.UserID)
	}
	if opts.ViewingType != "" {
		query = query.Where("viewing_type = ?", opts.ViewingType)
	}
	if opts.Status != nil {
		query = query.Where("status = ?", *opts.Status)
	}
	if opts.Result != "" {
		query = query.Where("result = ?", opts.Result)
	}
	if opts.StartDate != "" {
		query = query.Where("scheduled_at >= ?", opts.StartDate)
	}
	if opts.EndDate != "" {
		query = query.Where("scheduled_at <= ?", opts.EndDate+" 23:59:59")
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(req)).Order("scheduled_at DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *viewingRepository) AdminList(req *utils.Pagination, opts ViewingAdminListOptions) ([]model.HouseViewing, int64, error) {
	if req == nil {
		req = utils.NewPagination(1, 10)
	}
	var list []model.HouseViewing
	var total int64

	query := r.db.Model(&model.HouseViewing{})
	if opts.RegionID > 0 {
		query = query.Where("region_id = ?", opts.RegionID)
	}
	if opts.HouseID > 0 {
		query = query.Where("house_id = ?", opts.HouseID)
	}
	if opts.UserID > 0 {
		query = query.Where("user_id = ?", opts.UserID)
	}
	if opts.AgentID > 0 {
		query = query.Where("agent_id = ?", opts.AgentID)
	}
	if opts.Status != nil {
		query = query.Where("status = ?", *opts.Status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(req)).Order("created_at DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *viewingRepository) ListByUser(userID uint, req *utils.Pagination) ([]model.HouseViewing, int64, error) {
	if req == nil {
		req = utils.NewPagination(1, 10)
	}
	var list []model.HouseViewing
	var total int64

	query := r.db.Model(&model.HouseViewing{}).Where("user_id = ?", userID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(req)).Order("scheduled_at DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *viewingRepository) ListByAgent(agentID uint, req *utils.Pagination) ([]model.HouseViewing, int64, error) {
	if req == nil {
		req = utils.NewPagination(1, 10)
	}
	var list []model.HouseViewing
	var total int64

	query := r.db.Model(&model.HouseViewing{}).Where("agent_id = ?", agentID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(req)).Order("scheduled_at DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *viewingRepository) ListByHouse(houseID uint, req *utils.Pagination) ([]model.HouseViewing, int64, error) {
	if req == nil {
		req = utils.NewPagination(1, 10)
	}
	var list []model.HouseViewing
	var total int64

	query := r.db.Model(&model.HouseViewing{}).Where("house_id = ?", houseID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(req)).Order("scheduled_at DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *viewingRepository) MarkReminderSent(id uint) error {
	return r.db.Model(&model.HouseViewing{}).Where("id = ?", id).
		Update("reminder_sent", true).Error
}

func (r *viewingRepository) UpdateReminderSentAt(id uint) error {
	return r.db.Model(&model.HouseViewing{}).Where("id = ?", id).
		Updates(map[string]interface{}{
			"reminder_sent":    true,
			"reminder_sent_at": gorm.Expr("NOW()"),
		}).Error
}
