// Package repository 同城拼车出行数据访问层 - 投诉
package repository

import (
	"wuchang-tongcheng/internal/modules/pinche/model"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

// ComplaintListOptions 投诉列表过滤条件
type ComplaintListOptions struct {
	PincheID       uint
	BookingID      uint
	TripID         uint
	ComplainantID  uint
	RespondentID   uint
	ComplaintType  string
	Status         *int
	HandlerID      uint
	Keyword        string
}

// ComplaintRepository 投诉仓储接口
type ComplaintRepository interface {
	Create(c *model.PincheComplaint) error
	FindByID(id uint) (*model.PincheComplaint, error)
	Update(id uint, fields map[string]interface{}) error
	Delete(id uint) error

	List(regionID uint, pagination *utils.Pagination, opts ComplaintListOptions) ([]model.PincheComplaint, int64, error)
	ListByComplainant(complainantID uint, pagination *utils.Pagination) ([]model.PincheComplaint, int64, error)
	ListByRespondent(respondentID uint, pagination *utils.Pagination) ([]model.PincheComplaint, int64, error)
	ListByPinche(pincheID uint, pagination *utils.Pagination) ([]model.PincheComplaint, int64, error)
	ListPending(pagination *utils.Pagination) ([]model.PincheComplaint, int64, error)

	UpdateStatus(id uint, status int) error
	UpdateHandleResult(id uint, handlerID uint, handlerName, handleResult string) error
	UpdateAppeal(id uint, appealResult string, appealHandlerID uint) error
	CountByStatus(regionID uint, status int) (int64, error)
	CountByRespondent(respondentID uint, days int) (int64, error)
}

type complaintRepository struct {
	db *gorm.DB
}

// NewComplaintRepository 创建投诉仓储实例
func NewComplaintRepository(db *gorm.DB) ComplaintRepository {
	return &complaintRepository{db: db}
}

func (r *complaintRepository) Create(c *model.PincheComplaint) error {
	return r.db.Create(c).Error
}

func (r *complaintRepository) FindByID(id uint) (*model.PincheComplaint, error) {
	var c model.PincheComplaint
	if err := r.db.First(&c, id).Error; err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *complaintRepository) Update(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.PincheComplaint{}).Where("id = ?", id).Updates(fields).Error
}

func (r *complaintRepository) Delete(id uint) error {
	return r.db.Delete(&model.PincheComplaint{}, id).Error
}

func (r *complaintRepository) List(regionID uint, pagination *utils.Pagination, opts ComplaintListOptions) ([]model.PincheComplaint, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.PincheComplaint
	var total int64

	query := r.db.Model(&model.PincheComplaint{})
	if regionID > 0 {
		query = query.Where("region_id = ?", regionID)
	}
	if opts.PincheID > 0 {
		query = query.Where("pinche_id = ?", opts.PincheID)
	}
	if opts.BookingID > 0 {
		query = query.Where("booking_id = ?", opts.BookingID)
	}
	if opts.TripID > 0 {
		query = query.Where("trip_id = ?", opts.TripID)
	}
	if opts.ComplainantID > 0 {
		query = query.Where("complainant_id = ?", opts.ComplainantID)
	}
	if opts.RespondentID > 0 {
		query = query.Where("respondent_id = ?", opts.RespondentID)
	}
	if opts.ComplaintType != "" {
		query = query.Where("complaint_type = ?", opts.ComplaintType)
	}
	if opts.Status != nil {
		query = query.Where("status = ?", *opts.Status)
	}
	if opts.HandlerID > 0 {
		query = query.Where("handler_id = ?", opts.HandlerID)
	}
	if opts.Keyword != "" {
		query = query.Where("description ILIKE ? OR complaint_reason ILIKE ?", "%"+opts.Keyword+"%", "%"+opts.Keyword+"%")
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

func (r *complaintRepository) ListByComplainant(complainantID uint, pagination *utils.Pagination) ([]model.PincheComplaint, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.PincheComplaint
	var total int64

	query := r.db.Model(&model.PincheComplaint{}).Where("complainant_id = ?", complainantID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(pagination)).
		Order("created_at DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *complaintRepository) ListByRespondent(respondentID uint, pagination *utils.Pagination) ([]model.PincheComplaint, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.PincheComplaint
	var total int64

	query := r.db.Model(&model.PincheComplaint{}).Where("respondent_id = ?", respondentID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(pagination)).
		Order("created_at DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *complaintRepository) ListByPinche(pincheID uint, pagination *utils.Pagination) ([]model.PincheComplaint, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.PincheComplaint
	var total int64

	query := r.db.Model(&model.PincheComplaint{}).Where("pinche_id = ?", pincheID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(pagination)).
		Order("created_at DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *complaintRepository) ListPending(pagination *utils.Pagination) ([]model.PincheComplaint, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.PincheComplaint
	var total int64

	query := r.db.Model(&model.PincheComplaint{}).
		Where("status IN ?", []int{model.ComplaintStatusPending, model.ComplaintStatusHandling})
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(pagination)).
		Order("created_at ASC, id ASC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *complaintRepository) UpdateStatus(id uint, status int) error {
	return r.db.Model(&model.PincheComplaint{}).Where("id = ?", id).
		Update("status", status).Error
}

func (r *complaintRepository) UpdateHandleResult(id uint, handlerID uint, handlerName, handleResult string) error {
	return r.db.Model(&model.PincheComplaint{}).Where("id = ?", id).
		Updates(map[string]interface{}{
			"handler_id":    handlerID,
			"handler_name":  handlerName,
			"handle_result": handleResult,
			"status":        model.ComplaintStatusHandled,
			"handled_at":    gorm.Expr("NOW()"),
		}).Error
}

func (r *complaintRepository) UpdateAppeal(id uint, appealResult string, appealHandlerID uint) error {
	return r.db.Model(&model.PincheComplaint{}).Where("id = ?", id).
		Updates(map[string]interface{}{
			"appeal_result":      appealResult,
			"appeal_handler_id":  appealHandlerID,
			"appealed_at":        gorm.Expr("NOW()"),
		}).Error
}

func (r *complaintRepository) CountByStatus(regionID uint, status int) (int64, error) {
	var count int64
	q := r.db.Model(&model.PincheComplaint{}).Where("status = ?", status)
	if regionID > 0 {
		q = q.Where("region_id = ?", regionID)
	}
	if err := q.Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (r *complaintRepository) CountByRespondent(respondentID uint, days int) (int64, error) {
	var count int64
	q := r.db.Model(&model.PincheComplaint{}).Where("respondent_id = ?", respondentID)
	if days > 0 {
		q = q.Where("created_at >= NOW() - INTERVAL ?", gorm.Expr("? || ' days'", days))
	}
	if err := q.Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}
