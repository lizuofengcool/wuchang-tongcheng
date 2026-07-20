// Package repository love 相亲交友数据访问层 - 访客
package repository

import (
	"wuchang-tongcheng/internal/modules/love/model"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

// LoveVisitRepository 访客仓储接口
type LoveVisitRepository interface {
	Create(v *model.LoveVisit) error
	FindByID(id uint) (*model.LoveVisit, error)
	Update(v *model.LoveVisit) error
	UpdateFields(id uint, fields map[string]interface{}) error
	Delete(id uint) error
	List(pagination *utils.Pagination, opts LoveVisitListOptions) ([]model.LoveVisit, int64, error)
	ListByUser(userID uint, pagination *utils.Pagination) ([]model.LoveVisit, int64, error)
	ListByVisitor(visitorUserID uint, pagination *utils.Pagination) ([]model.LoveVisit, int64, error)
	ListUnread(userID uint, pagination *utils.Pagination) ([]model.LoveVisit, int64, error)
	MarkRead(id uint) error
	MarkAllRead(userID uint) error
	Upsert(v *model.LoveVisit) error
	CountByUser(userID uint) (int64, error)
	CountTodayByUser(userID uint) (int64, error)
	CountUnreadByUser(userID uint) (int64, error)
	CountWeeklyByUser(userID uint) (int64, error)
	CountMonthlyByUser(userID uint) (int64, error)
}

// LoveVisitListOptions 访客列表过滤
type LoveVisitListOptions struct {
	UserID        uint
	VisitorUserID uint
	VisitType     string
	IsRead        *bool
}

type loveVisitRepository struct {
	db *gorm.DB
}

// NewLoveVisitRepository 创建访客仓储
func NewLoveVisitRepository(db *gorm.DB) LoveVisitRepository {
	return &loveVisitRepository{db: db}
}

func (r *loveVisitRepository) Create(v *model.LoveVisit) error {
	return r.db.Create(v).Error
}

func (r *loveVisitRepository) FindByID(id uint) (*model.LoveVisit, error) {
	var v model.LoveVisit
	if err := r.db.First(&v, id).Error; err != nil {
		return nil, err
	}
	return &v, nil
}

func (r *loveVisitRepository) Update(v *model.LoveVisit) error {
	return r.db.Save(v).Error
}

func (r *loveVisitRepository) UpdateFields(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.LoveVisit{}).Where("id = ?", id).Updates(fields).Error
}

func (r *loveVisitRepository) Delete(id uint) error {
	return r.db.Delete(&model.LoveVisit{}, id).Error
}

func (r *loveVisitRepository) List(pagination *utils.Pagination, opts LoveVisitListOptions) ([]model.LoveVisit, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.LoveVisit
	var total int64

	query := r.db.Model(&model.LoveVisit{})
	if opts.UserID > 0 {
		query = query.Where("user_id = ?", opts.UserID)
	}
	if opts.VisitorUserID > 0 {
		query = query.Where("visitor_user_id = ?", opts.VisitorUserID)
	}
	if opts.VisitType != "" {
		query = query.Where("visit_type = ?", opts.VisitType)
	}
	if opts.IsRead != nil {
		query = query.Where("is_read = ?", *opts.IsRead)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(pagination)).Order("id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *loveVisitRepository) ListByUser(userID uint, pagination *utils.Pagination) ([]model.LoveVisit, int64, error) {
	return r.List(pagination, LoveVisitListOptions{UserID: userID})
}

func (r *loveVisitRepository) ListByVisitor(visitorUserID uint, pagination *utils.Pagination) ([]model.LoveVisit, int64, error) {
	return r.List(pagination, LoveVisitListOptions{VisitorUserID: visitorUserID})
}

func (r *loveVisitRepository) ListUnread(userID uint, pagination *utils.Pagination) ([]model.LoveVisit, int64, error) {
	unread := false
	return r.List(pagination, LoveVisitListOptions{UserID: userID, IsRead: &unread})
}

func (r *loveVisitRepository) MarkRead(id uint) error {
	return r.db.Model(&model.LoveVisit{}).Where("id = ?", id).Updates(map[string]interface{}{
		"is_read": true,
		"updated_at": gorm.Expr("NOW()"),
	}).Error
}

func (r *loveVisitRepository) MarkAllRead(userID uint) error {
	return r.db.Model(&model.LoveVisit{}).Where("user_id = ? AND is_read = ?", userID, false).Updates(map[string]interface{}{
		"is_read": true,
		"updated_at": gorm.Expr("NOW()"),
	}).Error
}

func (r *loveVisitRepository) Upsert(v *model.LoveVisit) error {
	// 唯一约束 (user_id, visitor_user_id) - 若已存在则更新时间，否则插入
	result := r.db.Where("user_id = ? AND visitor_user_id = ?", v.UserID, v.VisitorUserID).FirstOrCreate(v)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		// 已存在记录，更新时间
		return r.db.Model(&model.LoveVisit{}).Where("user_id = ? AND visitor_user_id = ?", v.UserID, v.VisitorUserID).Updates(map[string]interface{}{
			"visit_type": v.VisitType,
			"source":     v.Source,
			"duration":   v.Duration,
			"updated_at": gorm.Expr("NOW()"),
			"is_read":    false,
		}).Error
	}
	return nil
}

func (r *loveVisitRepository) CountByUser(userID uint) (int64, error) {
	var count int64
	err := r.db.Model(&model.LoveVisit{}).Where("user_id = ?", userID).Count(&count).Error
	return count, err
}

func (r *loveVisitRepository) CountTodayByUser(userID uint) (int64, error) {
	var count int64
	err := r.db.Model(&model.LoveVisit{}).Where("user_id = ? AND created_at >= DATE_TRUNC('day', NOW())", userID).Count(&count).Error
	return count, err
}

func (r *loveVisitRepository) CountUnreadByUser(userID uint) (int64, error) {
	var count int64
	err := r.db.Model(&model.LoveVisit{}).Where("user_id = ? AND is_read = ?", userID, false).Count(&count).Error
	return count, err
}

func (r *loveVisitRepository) CountWeeklyByUser(userID uint) (int64, error) {
	var count int64
	err := r.db.Model(&model.LoveVisit{}).Where("user_id = ? AND created_at >= DATE_TRUNC('week', NOW())", userID).Count(&count).Error
	return count, err
}

func (r *loveVisitRepository) CountMonthlyByUser(userID uint) (int64, error) {
	var count int64
	err := r.db.Model(&model.LoveVisit{}).Where("user_id = ? AND created_at >= DATE_TRUNC('month', NOW())", userID).Count(&count).Error
	return count, err
}
