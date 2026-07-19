// Package repository IM 消息中台精简版数据访问层
package repository

import (
	"wuchang-tongcheng/internal/modules/im/model"

	"gorm.io/gorm"
)

// IMRepository IM 中台仓储接口
type IMRepository interface {
	// 会话
	CreateSession(s *model.Session) error
	FindSessionByID(sessionID string) (*model.Session, error)
	ListSessionsByUser(userID uint, page, pageSize int) ([]model.Session, int64, error)
	UpdateSessionFields(id uint, fields map[string]interface{}) error

	// 消息
	CreateMessage(m *model.Message) error
	ListMessages(sessionID string, page, pageSize int) ([]model.Message, int64, error)
	UpdateMessageFields(id uint, fields map[string]interface{}) error

	// 系统通知
	CreateNotification(n *model.SystemNotification) error
	ListNotifications(userID uint, page, pageSize int) ([]model.SystemNotification, int64, error)
	ListUnreadNotifications(userID uint) ([]model.SystemNotification, int64, error)
	UpdateNotificationFields(id uint, fields map[string]interface{}) error
	MarkAllRead(userID uint) error

	// 隐私号码
	CreatePrivacyNumber(p *model.PrivacyNumber) error
	FindPrivacyNumber(no string) (*model.PrivacyNumber, error)
	UpdatePrivacyNumberFields(id uint, fields map[string]interface{}) error
}

type imRepository struct {
	db *gorm.DB
}

// NewIMRepository 创建仓储实例
func NewIMRepository(db *gorm.DB) IMRepository {
	return &imRepository{db: db}
}

// ===== 会话 =====

func (r *imRepository) CreateSession(s *model.Session) error {
	return r.db.Create(s).Error
}

func (r *imRepository) FindSessionByID(sessionID string) (*model.Session, error) {
	var s model.Session
	if err := r.db.Where("session_id = ?", sessionID).First(&s).Error; err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *imRepository) ListSessionsByUser(userID uint, page, pageSize int) ([]model.Session, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	var list []model.Session
	var total int64
	// 参与者 ID 存在 jsonb 数组中，使用 @> 包含查询
	q := r.db.Model(&model.Session{}).Where("participants @ ?::jsonb", `["`+uintToStr(userID)+`"]`)
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := q.Order("last_message_at DESC NULLS LAST, id DESC").
		Offset((page - 1) * pageSize).Limit(pageSize).
		Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *imRepository) UpdateSessionFields(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.Session{}).Where("id = ?", id).Updates(fields).Error
}

// ===== 消息 =====

func (r *imRepository) CreateMessage(m *model.Message) error {
	return r.db.Create(m).Error
}

func (r *imRepository) ListMessages(sessionID string, page, pageSize int) ([]model.Message, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 50
	}
	var list []model.Message
	var total int64
	q := r.db.Model(&model.Message{}).Where("session_id = ?", sessionID)
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := q.Order("created_at ASC, id ASC").
		Offset((page - 1) * pageSize).Limit(pageSize).
		Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *imRepository) UpdateMessageFields(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.Message{}).Where("id = ?", id).Updates(fields).Error
}

// ===== 系统通知 =====

func (r *imRepository) CreateNotification(n *model.SystemNotification) error {
	return r.db.Create(n).Error
}

func (r *imRepository) ListNotifications(userID uint, page, pageSize int) ([]model.SystemNotification, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	var list []model.SystemNotification
	var total int64
	q := r.db.Model(&model.SystemNotification{}).Where("user_id = ?", userID)
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := q.Order("created_at DESC, id DESC").
		Offset((page - 1) * pageSize).Limit(pageSize).
		Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *imRepository) ListUnreadNotifications(userID uint) ([]model.SystemNotification, int64, error) {
	var list []model.SystemNotification
	var total int64
	q := r.db.Model(&model.SystemNotification{}).Where("user_id = ? AND is_read = false", userID)
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := q.Order("created_at DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *imRepository) UpdateNotificationFields(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.SystemNotification{}).Where("id = ?", id).Updates(fields).Error
}

func (r *imRepository) MarkAllRead(userID uint) error {
	return r.db.Model(&model.SystemNotification{}).
		Where("user_id = ? AND is_read = false", userID).
		Update("is_read", true).Error
}

// ===== 隐私号码 =====

func (r *imRepository) CreatePrivacyNumber(p *model.PrivacyNumber) error {
	return r.db.Create(p).Error
}

func (r *imRepository) FindPrivacyNumber(no string) (*model.PrivacyNumber, error) {
	var p model.PrivacyNumber
	if err := r.db.Where("privacy_no = ?", no).First(&p).Error; err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *imRepository) UpdatePrivacyNumberFields(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.PrivacyNumber{}).Where("id = ?", id).Updates(fields).Error
}

// uintToStr uint 转字符串
func uintToStr(n uint) string {
	if n == 0 {
		return "0"
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[i:])
}
