// Package repository 沟通消息数据访问层
// 依据 v3.2.1 架构方案：对标 BOSS直聘在线聊天
package repository

import (
	"wuchang-tongcheng/internal/modules/job/model"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

// MessageRepository 消息仓储接口
type MessageRepository interface {
	Create(m *model.JobMessage) error
	FindByID(id uint) (*model.JobMessage, error)
	Update(id uint, fields map[string]interface{}) error
	Delete(id uint) error
	List(query MessageListQuery, pagination *utils.Pagination) ([]model.JobMessage, int64, error)
	ListByConversation(conversationID string, pagination *utils.Pagination) ([]model.JobMessage, int64, error)
	ListByUser(userID uint, role string, pagination *utils.Pagination) ([]model.JobMessage, int64, error)
	ListConversations(userID uint) ([]model.JobMessage, error)
	MarkRead(fromUserID, toUserID uint) error
	MarkReadByConversation(conversationID string, toUserID uint) error
	CountUnread(userID uint) (int64, error)
	CountUnreadByUser(fromUserID, toUserID uint) (int64, error)
	BatchDelete(ids []uint) error
}

// MessageListQuery 消息列表查询
type MessageListQuery struct {
	UserID         uint
	Role           string // from/to/all
	JobID          uint
	ApplicationID  uint
	ConversationID string
	MessageType    string
	IsRead         *bool
	IsSystem       *bool
	Keyword        string
}

type messageRepository struct {
	db *gorm.DB
}

// NewMessageRepository 创建消息仓储实例
func NewMessageRepository(db *gorm.DB) MessageRepository {
	return &messageRepository{db: db}
}

func (r *messageRepository) Create(m *model.JobMessage) error {
	return r.db.Create(m).Error
}

func (r *messageRepository) FindByID(id uint) (*model.JobMessage, error) {
	var m model.JobMessage
	if err := r.db.First(&m, id).Error; err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *messageRepository) Update(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.JobMessage{}).Where("id = ?", id).Updates(fields).Error
}

func (r *messageRepository) Delete(id uint) error {
	return r.db.Delete(&model.JobMessage{}, id).Error
}

func (r *messageRepository) List(query MessageListQuery, pagination *utils.Pagination) ([]model.JobMessage, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 20)
	}
	var list []model.JobMessage
	var total int64

	q := r.db.Model(&model.JobMessage{})
	switch query.Role {
	case "from":
		q = q.Where("from_user_id = ?", query.UserID)
	case "to":
		q = q.Where("to_user_id = ?", query.UserID)
	case "all":
		q = q.Where("from_user_id = ? OR to_user_id = ?", query.UserID, query.UserID)
	}
	if query.JobID > 0 {
		q = q.Where("job_id = ?", query.JobID)
	}
	if query.ApplicationID > 0 {
		q = q.Where("application_id = ?", query.ApplicationID)
	}
	if query.ConversationID != "" {
		q = q.Where("conversation_id = ?", query.ConversationID)
	}
	if query.MessageType != "" {
		q = q.Where("message_type = ?", query.MessageType)
	}
	if query.IsRead != nil {
		q = q.Where("is_read = ?", *query.IsRead)
	}
	if query.IsSystem != nil {
		q = q.Where("is_system = ?", *query.IsSystem)
	}
	if query.Keyword != "" {
		q = q.Where("content ILIKE ?", "%"+query.Keyword+"%")
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

func (r *messageRepository) ListByConversation(conversationID string, pagination *utils.Pagination) ([]model.JobMessage, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 50)
	}
	var list []model.JobMessage
	var total int64

	q := r.db.Model(&model.JobMessage{}).Where("conversation_id = ? AND status = ?", conversationID, model.MessageStatusNormal)
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := q.Scopes(utils.Paginate(pagination)).
		Order("created_at ASC, id ASC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *messageRepository) ListByUser(userID uint, role string, pagination *utils.Pagination) ([]model.JobMessage, int64, error) {
	return r.List(MessageListQuery{UserID: userID, Role: role}, pagination)
}

// ListConversations 列出当前用户的会话列表（按最近一条消息排序）
func (r *messageRepository) ListConversations(userID uint) ([]model.JobMessage, error) {
	var list []model.JobMessage
	// 取每个 conversation 的最近一条消息
	subQuery := r.db.Model(&model.JobMessage{}).
		Select("MAX(id) as id").
		Where("from_user_id = ? OR to_user_id = ?", userID, userID).
		Where("status = ?", model.MessageStatusNormal).
		Group("conversation_id")
	if err := r.db.Where("id IN (?)", subQuery).
		Order("created_at DESC").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *messageRepository) MarkRead(fromUserID, toUserID uint) error {
	return r.db.Model(&model.JobMessage{}).
		Where("from_user_id = ? AND to_user_id = ? AND is_read = ?", fromUserID, toUserID, false).
		Updates(map[string]interface{}{
			"is_read": true,
			"read_at": gorm.Expr("NOW()"),
		}).Error
}

func (r *messageRepository) MarkReadByConversation(conversationID string, toUserID uint) error {
	return r.db.Model(&model.JobMessage{}).
		Where("conversation_id = ? AND to_user_id = ? AND is_read = ?", conversationID, toUserID, false).
		Updates(map[string]interface{}{
			"is_read": true,
			"read_at": gorm.Expr("NOW()"),
		}).Error
}

func (r *messageRepository) CountUnread(userID uint) (int64, error) {
	var count int64
	err := r.db.Model(&model.JobMessage{}).
		Where("to_user_id = ? AND is_read = ? AND status = ?", userID, false, model.MessageStatusNormal).
		Count(&count).Error
	return count, err
}

func (r *messageRepository) CountUnreadByUser(fromUserID, toUserID uint) (int64, error) {
	var count int64
	err := r.db.Model(&model.JobMessage{}).
		Where("from_user_id = ? AND to_user_id = ? AND is_read = ? AND status = ?", fromUserID, toUserID, false, model.MessageStatusNormal).
		Count(&count).Error
	return count, err
}

func (r *messageRepository) BatchDelete(ids []uint) error {
	return r.db.Where("id IN ?", ids).Delete(&model.JobMessage{}).Error
}
