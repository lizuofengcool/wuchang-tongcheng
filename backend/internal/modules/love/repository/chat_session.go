// Package repository love 相亲交友数据访问层 - 聊天会话
package repository

import (
	"wuchang-tongcheng/internal/modules/love/model"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

// LoveChatSessionRepository 聊天会话仓储接口
type LoveChatSessionRepository interface {
	Create(s *model.LoveChatSession) error
	FindByID(id uint) (*model.LoveChatSession, error)
	FindBySessionNo(no string) (*model.LoveChatSession, error)
	FindByMatchID(matchID uint) (*model.LoveChatSession, error)
	Update(s *model.LoveChatSession) error
	UpdateFields(id uint, fields map[string]interface{}) error
	Delete(id uint) error
	List(pagination *utils.Pagination, opts LoveChatSessionListOptions) ([]model.LoveChatSession, int64, error)
	ListByUser(userID uint, pagination *utils.Pagination) ([]model.LoveChatSession, int64, error)
	ListByUserAndStatus(userID uint, status int, pagination *utils.Pagination) ([]model.LoveChatSession, int64, error)

	UpdateLastMessage(id uint, msgID uint, content, msgType string, senderID uint) error
	IncrUnreadCount(id uint, side string) error
	ResetUnreadCount(id uint, side string) error
	SetMuted(id uint, side string, muted bool) error
	SetPinned(id uint, side string, pinned bool) error
	SetDeleted(id uint, side string, deleted bool) error
	Dissolve(id uint, byUserID uint, reason string) error

	IncrMessageCount(id uint) error
	IncrGiftCount(id uint) error

	CountActiveByUser(userID uint) (int64, error)
}

// LoveChatSessionListOptions 会话列表过滤
type LoveChatSessionListOptions struct {
	UserID uint
	Status *int
}

type loveChatSessionRepository struct {
	db *gorm.DB
}

// NewLoveChatSessionRepository 创建聊天会话仓储
func NewLoveChatSessionRepository(db *gorm.DB) LoveChatSessionRepository {
	return &loveChatSessionRepository{db: db}
}

func (r *loveChatSessionRepository) Create(s *model.LoveChatSession) error {
	return r.db.Create(s).Error
}

func (r *loveChatSessionRepository) FindByID(id uint) (*model.LoveChatSession, error) {
	var s model.LoveChatSession
	if err := r.db.First(&s, id).Error; err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *loveChatSessionRepository) FindBySessionNo(no string) (*model.LoveChatSession, error) {
	var s model.LoveChatSession
	if err := r.db.Where("session_no = ?", no).First(&s).Error; err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *loveChatSessionRepository) FindByMatchID(matchID uint) (*model.LoveChatSession, error) {
	var s model.LoveChatSession
	if err := r.db.Where("match_id = ?", matchID).First(&s).Error; err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *loveChatSessionRepository) Update(s *model.LoveChatSession) error {
	return r.db.Save(s).Error
}

func (r *loveChatSessionRepository) UpdateFields(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.LoveChatSession{}).Where("id = ?", id).Updates(fields).Error
}

func (r *loveChatSessionRepository) Delete(id uint) error {
	return r.db.Delete(&model.LoveChatSession{}, id).Error
}

func (r *loveChatSessionRepository) List(pagination *utils.Pagination, opts LoveChatSessionListOptions) ([]model.LoveChatSession, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.LoveChatSession
	var total int64

	query := r.db.Model(&model.LoveChatSession{})
	if opts.UserID > 0 {
		query = query.Where("user_id_a = ? OR user_id_b = ?", opts.UserID, opts.UserID)
	}
	if opts.Status != nil {
		query = query.Where("status = ?", *opts.Status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(pagination)).Order("last_message_at DESC NULLS LAST, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *loveChatSessionRepository) ListByUser(userID uint, pagination *utils.Pagination) ([]model.LoveChatSession, int64, error) {
	return r.List(pagination, LoveChatSessionListOptions{UserID: userID})
}

func (r *loveChatSessionRepository) ListByUserAndStatus(userID uint, status int, pagination *utils.Pagination) ([]model.LoveChatSession, int64, error) {
	st := status
	return r.List(pagination, LoveChatSessionListOptions{UserID: userID, Status: &st})
}

func (r *loveChatSessionRepository) UpdateLastMessage(id uint, msgID uint, content, msgType string, senderID uint) error {
	return r.db.Model(&model.LoveChatSession{}).Where("id = ?", id).Updates(map[string]interface{}{
		"last_message_id":      msgID,
		"last_message_content": content,
		"last_message_type":    msgType,
		"last_message_at":      gorm.Expr("NOW()"),
		"last_sender_id":       senderID,
	}).Error
}

func (r *loveChatSessionRepository) IncrUnreadCount(id uint, side string) error {
	col := "unread_count_a"
	if side == "b" {
		col = "unread_count_b"
	}
	return r.db.Model(&model.LoveChatSession{}).Where("id = ?", id).UpdateColumn(col, gorm.Expr(col+" + 1")).Error
}

func (r *loveChatSessionRepository) ResetUnreadCount(id uint, side string) error {
	col := "unread_count_a"
	if side == "b" {
		col = "unread_count_b"
	}
	return r.db.Model(&model.LoveChatSession{}).Where("id = ?", id).Update(col, 0).Error
}

func (r *loveChatSessionRepository) SetMuted(id uint, side string, muted bool) error {
	col := "muted_a"
	if side == "b" {
		col = "muted_b"
	}
	return r.db.Model(&model.LoveChatSession{}).Where("id = ?", id).Update(col, muted).Error
}

func (r *loveChatSessionRepository) SetPinned(id uint, side string, pinned bool) error {
	col := "pinned_a"
	if side == "b" {
		col = "pinned_b"
	}
	return r.db.Model(&model.LoveChatSession{}).Where("id = ?", id).Update(col, pinned).Error
}

func (r *loveChatSessionRepository) SetDeleted(id uint, side string, deleted bool) error {
	col := "deleted_a"
	if side == "b" {
		col = "deleted_b"
	}
	return r.db.Model(&model.LoveChatSession{}).Where("id = ?", id).Update(col, deleted).Error
}

func (r *loveChatSessionRepository) Dissolve(id uint, byUserID uint, reason string) error {
	return r.db.Model(&model.LoveChatSession{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":          model.ChatSessionStatusDissolved,
		"dissolved_at":    gorm.Expr("NOW()"),
		"dissolve_by":     byUserID,
		"dissolve_reason": reason,
	}).Error
}

func (r *loveChatSessionRepository) IncrMessageCount(id uint) error {
	return r.db.Model(&model.LoveChatSession{}).Where("id = ?", id).UpdateColumn("message_count", gorm.Expr("message_count + 1")).Error
}

func (r *loveChatSessionRepository) IncrGiftCount(id uint) error {
	return r.db.Model(&model.LoveChatSession{}).Where("id = ?", id).UpdateColumn("gift_count", gorm.Expr("gift_count + 1")).Error
}

func (r *loveChatSessionRepository) CountActiveByUser(userID uint) (int64, error) {
	var count int64
	err := r.db.Model(&model.LoveChatSession{}).Where(
		"(user_id_a = ? OR user_id_b = ?) AND status = ?",
		userID, userID, model.ChatSessionStatusActive,
	).Count(&count).Error
	return count, err
}
