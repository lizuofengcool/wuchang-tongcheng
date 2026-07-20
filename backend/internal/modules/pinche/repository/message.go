// Package repository 同城拼车出行数据访问层 - 行程内消息
package repository

import (
	"time"

	"wuchang-tongcheng/internal/modules/pinche/model"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

// MessageListOptions 消息列表过滤条件
type MessageListOptions struct {
	PincheID    uint
	BookingID   uint
	TripID      uint
	SenderID    uint
	ReceiverID  uint
	MessageType string
	IsRead      *bool
	Keyword     string
}

// MessageRepository 行程内消息仓储接口
type MessageRepository interface {
	Create(m *model.PincheMessage) error
	FindByID(id uint) (*model.PincheMessage, error)
	Update(id uint, fields map[string]interface{}) error
	Delete(id uint) error

	List(regionID uint, pagination *utils.Pagination, opts MessageListOptions) ([]model.PincheMessage, int64, error)
	ListByPinche(pincheID uint, pagination *utils.Pagination) ([]model.PincheMessage, int64, error)
	ListByConversation(senderID, receiverID uint, pagination *utils.Pagination) ([]model.PincheMessage, int64, error)
	ListUnread(receiverID uint, pagination *utils.Pagination) ([]model.PincheMessage, int64, error)

	MarkRead(id uint) error
	MarkAllRead(receiverID, pincheID uint) error
	CountUnread(receiverID uint) (int64, error)
}

type messageRepository struct {
	db *gorm.DB
}

// NewMessageRepository 创建行程内消息仓储实例
func NewMessageRepository(db *gorm.DB) MessageRepository {
	return &messageRepository{db: db}
}

func (r *messageRepository) Create(m *model.PincheMessage) error {
	return r.db.Create(m).Error
}

func (r *messageRepository) FindByID(id uint) (*model.PincheMessage, error) {
	var m model.PincheMessage
	if err := r.db.First(&m, id).Error; err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *messageRepository) Update(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.PincheMessage{}).Where("id = ?", id).Updates(fields).Error
}

func (r *messageRepository) Delete(id uint) error {
	return r.db.Delete(&model.PincheMessage{}, id).Error
}

func (r *messageRepository) List(regionID uint, pagination *utils.Pagination, opts MessageListOptions) ([]model.PincheMessage, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.PincheMessage
	var total int64

	query := r.db.Model(&model.PincheMessage{})
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
	if opts.SenderID > 0 {
		query = query.Where("sender_id = ?", opts.SenderID)
	}
	if opts.ReceiverID > 0 {
		query = query.Where("receiver_id = ?", opts.ReceiverID)
	}
	if opts.MessageType != "" {
		query = query.Where("message_type = ?", opts.MessageType)
	}
	if opts.IsRead != nil {
		query = query.Where("is_read = ?", *opts.IsRead)
	}
	if opts.Keyword != "" {
		query = query.Where("content ILIKE ?", "%"+opts.Keyword+"%")
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

func (r *messageRepository) ListByPinche(pincheID uint, pagination *utils.Pagination) ([]model.PincheMessage, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.PincheMessage
	var total int64

	query := r.db.Model(&model.PincheMessage{}).Where("pinche_id = ?", pincheID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(pagination)).
		Order("created_at ASC, id ASC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *messageRepository) ListByConversation(senderID, receiverID uint, pagination *utils.Pagination) ([]model.PincheMessage, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.PincheMessage
	var total int64

	query := r.db.Model(&model.PincheMessage{}).
		Where("(sender_id = ? AND receiver_id = ?) OR (sender_id = ? AND receiver_id = ?)",
			senderID, receiverID, receiverID, senderID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(pagination)).
		Order("created_at DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *messageRepository) ListUnread(receiverID uint, pagination *utils.Pagination) ([]model.PincheMessage, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.PincheMessage
	var total int64

	query := r.db.Model(&model.PincheMessage{}).
		Where("receiver_id = ? AND is_read = false", receiverID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(pagination)).
		Order("created_at DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *messageRepository) MarkRead(id uint) error {
	return r.db.Model(&model.PincheMessage{}).Where("id = ?", id).
		Updates(map[string]interface{}{
			"is_read": true,
			"read_at": time.Now(),
		}).Error
}

func (r *messageRepository) MarkAllRead(receiverID, pincheID uint) error {
	q := r.db.Model(&model.PincheMessage{}).
		Where("receiver_id = ? AND is_read = false", receiverID)
	if pincheID > 0 {
		q = q.Where("pinche_id = ?", pincheID)
	}
	return q.Updates(map[string]interface{}{
		"is_read": true,
		"read_at": time.Now(),
	}).Error
}

func (r *messageRepository) CountUnread(receiverID uint) (int64, error) {
	var count int64
	if err := r.db.Model(&model.PincheMessage{}).
		Where("receiver_id = ? AND is_read = false", receiverID).
		Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}
