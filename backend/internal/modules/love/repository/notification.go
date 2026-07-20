// Package repository love 相亲交友数据访问层 - 通知
package repository

import (
	"wuchang-tongcheng/internal/modules/love/model"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

// LoveNotificationRepository 通知仓储接口
type LoveNotificationRepository interface {
	Create(n *model.LoveNotification) error
	FindByID(id uint) (*model.LoveNotification, error)
	Update(n *model.LoveNotification) error
	UpdateFields(id uint, fields map[string]interface{}) error
	Delete(id uint) error
	List(pagination *utils.Pagination, opts LoveNotificationListOptions) ([]model.LoveNotification, int64, error)
	ListByUser(userID uint, pagination *utils.Pagination) ([]model.LoveNotification, int64, error)
	ListUnread(userID uint, pagination *utils.Pagination) ([]model.LoveNotification, int64, error)
	MarkRead(id uint) error
	MarkAllRead(userID uint) error
	BatchMarkRead(ids []uint) error
	UpdatePushStatus(id uint, status, errorMsg string) error
	CountUnreadByUser(userID uint) (int64, error)
	CountByUserAndType(userID uint, notifyType string) (int64, error)
}

// LoveNotificationListOptions 通知列表过滤
type LoveNotificationListOptions struct {
	UserID uint
	Type   string
	IsRead *bool
}

type loveNotificationRepository struct {
	db *gorm.DB
}

// NewLoveNotificationRepository 创建通知仓储
func NewLoveNotificationRepository(db *gorm.DB) LoveNotificationRepository {
	return &loveNotificationRepository{db: db}
}

func (r *loveNotificationRepository) Create(n *model.LoveNotification) error {
	return r.db.Create(n).Error
}

func (r *loveNotificationRepository) FindByID(id uint) (*model.LoveNotification, error) {
	var n model.LoveNotification
	if err := r.db.First(&n, id).Error; err != nil {
		return nil, err
	}
	return &n, nil
}

func (r *loveNotificationRepository) Update(n *model.LoveNotification) error {
	return r.db.Save(n).Error
}

func (r *loveNotificationRepository) UpdateFields(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.LoveNotification{}).Where("id = ?", id).Updates(fields).Error
}

func (r *loveNotificationRepository) Delete(id uint) error {
	return r.db.Delete(&model.LoveNotification{}, id).Error
}

func (r *loveNotificationRepository) List(pagination *utils.Pagination, opts LoveNotificationListOptions) ([]model.LoveNotification, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.LoveNotification
	var total int64

	query := r.db.Model(&model.LoveNotification{})
	if opts.UserID > 0 {
		query = query.Where("user_id = ?", opts.UserID)
	}
	if opts.Type != "" {
		query = query.Where("type = ?", opts.Type)
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

func (r *loveNotificationRepository) ListByUser(userID uint, pagination *utils.Pagination) ([]model.LoveNotification, int64, error) {
	return r.List(pagination, LoveNotificationListOptions{UserID: userID})
}

func (r *loveNotificationRepository) ListUnread(userID uint, pagination *utils.Pagination) ([]model.LoveNotification, int64, error) {
	unread := false
	return r.List(pagination, LoveNotificationListOptions{UserID: userID, IsRead: &unread})
}

func (r *loveNotificationRepository) MarkRead(id uint) error {
	return r.db.Model(&model.LoveNotification{}).Where("id = ?", id).Updates(map[string]interface{}{
		"is_read":  true,
		"read_at":  gorm.Expr("NOW()"),
		"updated_at": gorm.Expr("NOW()"),
	}).Error
}

func (r *loveNotificationRepository) MarkAllRead(userID uint) error {
	return r.db.Model(&model.LoveNotification{}).Where("user_id = ? AND is_read = ?", userID, false).Updates(map[string]interface{}{
		"is_read":  true,
		"read_at":  gorm.Expr("NOW()"),
		"updated_at": gorm.Expr("NOW()"),
	}).Error
}

func (r *loveNotificationRepository) BatchMarkRead(ids []uint) error {
	return r.db.Model(&model.LoveNotification{}).Where("id IN ?", ids).Updates(map[string]interface{}{
		"is_read":  true,
		"read_at":  gorm.Expr("NOW()"),
		"updated_at": gorm.Expr("NOW()"),
	}).Error
}

func (r *loveNotificationRepository) UpdatePushStatus(id uint, status, errorMsg string) error {
	return r.db.Model(&model.LoveNotification{}).Where("id = ?", id).Updates(map[string]interface{}{
		"is_pushed":   true,
		"pushed_at":   gorm.Expr("NOW()"),
		"push_status": status,
		"push_error":  errorMsg,
	}).Error
}

func (r *loveNotificationRepository) CountUnreadByUser(userID uint) (int64, error) {
	var count int64
	err := r.db.Model(&model.LoveNotification{}).Where("user_id = ? AND is_read = ?", userID, false).Count(&count).Error
	return count, err
}

func (r *loveNotificationRepository) CountByUserAndType(userID uint, notifyType string) (int64, error) {
	var count int64
	err := r.db.Model(&model.LoveNotification{}).Where("user_id = ? AND type = ?", userID, notifyType).Count(&count).Error
	return count, err
}
