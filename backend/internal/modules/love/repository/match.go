// Package repository love 相亲交友数据访问层 - 匹配记录
package repository

import (
	"wuchang-tongcheng/internal/modules/love/model"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

// LoveMatchRepository 匹配记录仓储接口
type LoveMatchRepository interface {
	Create(m *model.LoveMatch) error
	FindByID(id uint) (*model.LoveMatch, error)
	FindByMatchNo(matchNo string) (*model.LoveMatch, error)
	FindByUserPair(userA, userB uint) (*model.LoveMatch, error)
	Update(m *model.LoveMatch) error
	UpdateFields(id uint, fields map[string]interface{}) error
	Delete(id uint) error
	List(pagination *utils.Pagination, opts LoveMatchListOptions) ([]model.LoveMatch, int64, error)
	ListByUser(userID uint, pagination *utils.Pagination) ([]model.LoveMatch, int64, error)
	ListByUserAndStatus(userID uint, status int, pagination *utils.Pagination) ([]model.LoveMatch, int64, error)
	Dissolve(id uint, byUserID uint, reason string) error
	UpdateLastMessage(id uint, msgID uint, content, msgType string, senderID uint) error
	UpdateUnreadCount(id uint, side string, count int) error
	UpdateChatSessionID(id uint, sessionID uint) error
	CountByUser(userID uint) (int64, error)
	CountTodayByUser(userID uint) (int64, error)
}

// LoveMatchListOptions 匹配列表过滤
type LoveMatchListOptions struct {
	UserID    uint
	Status    *int
	MatchType string
}

type loveMatchRepository struct {
	db *gorm.DB
}

// NewLoveMatchRepository 创建匹配记录仓储
func NewLoveMatchRepository(db *gorm.DB) LoveMatchRepository {
	return &loveMatchRepository{db: db}
}

func (r *loveMatchRepository) Create(m *model.LoveMatch) error {
	return r.db.Create(m).Error
}

func (r *loveMatchRepository) FindByID(id uint) (*model.LoveMatch, error) {
	var m model.LoveMatch
	if err := r.db.First(&m, id).Error; err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *loveMatchRepository) FindByMatchNo(matchNo string) (*model.LoveMatch, error) {
	var m model.LoveMatch
	if err := r.db.Where("match_no = ?", matchNo).First(&m).Error; err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *loveMatchRepository) FindByUserPair(userA, userB uint) (*model.LoveMatch, error) {
	var m model.LoveMatch
	err := r.db.Where(
		"(user_id_a = ? AND user_id_b = ?) OR (user_id_a = ? AND user_id_b = ?)",
		userA, userB, userB, userA,
	).First(&m).Error
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *loveMatchRepository) Update(m *model.LoveMatch) error {
	return r.db.Save(m).Error
}

func (r *loveMatchRepository) UpdateFields(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.LoveMatch{}).Where("id = ?", id).Updates(fields).Error
}

func (r *loveMatchRepository) Delete(id uint) error {
	return r.db.Delete(&model.LoveMatch{}, id).Error
}

func (r *loveMatchRepository) List(pagination *utils.Pagination, opts LoveMatchListOptions) ([]model.LoveMatch, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.LoveMatch
	var total int64

	query := r.db.Model(&model.LoveMatch{})
	if opts.UserID > 0 {
		query = query.Where("user_id_a = ? OR user_id_b = ?", opts.UserID, opts.UserID)
	}
	if opts.Status != nil {
		query = query.Where("status = ?", *opts.Status)
	}
	if opts.MatchType != "" {
		query = query.Where("match_type = ?", opts.MatchType)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(pagination)).Order("matched_at DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *loveMatchRepository) ListByUser(userID uint, pagination *utils.Pagination) ([]model.LoveMatch, int64, error) {
	return r.List(pagination, LoveMatchListOptions{UserID: userID})
}

func (r *loveMatchRepository) ListByUserAndStatus(userID uint, status int, pagination *utils.Pagination) ([]model.LoveMatch, int64, error) {
	st := status
	return r.List(pagination, LoveMatchListOptions{UserID: userID, Status: &st})
}

func (r *loveMatchRepository) Dissolve(id uint, byUserID uint, reason string) error {
	return r.db.Model(&model.LoveMatch{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":          model.MatchStatusDissolved,
		"dissolved_at":    gorm.Expr("NOW()"),
		"dissolve_by":     byUserID,
		"dissolve_reason": reason,
	}).Error
}

func (r *loveMatchRepository) UpdateLastMessage(id uint, msgID uint, content, msgType string, senderID uint) error {
	return r.db.Model(&model.LoveMatch{}).Where("id = ?", id).Updates(map[string]interface{}{
		"last_message_id":      msgID,
		"last_message_content": content,
		"last_message_type":    msgType,
		"last_message_at":      gorm.Expr("NOW()"),
		"last_sender_id":       senderID,
	}).Error
}

func (r *loveMatchRepository) UpdateUnreadCount(id uint, side string, count int) error {
	col := "unread_count_a"
	if side == "b" {
		col = "unread_count_b"
	}
	return r.db.Model(&model.LoveMatch{}).Where("id = ?", id).Update(col, count).Error
}

func (r *loveMatchRepository) UpdateChatSessionID(id uint, sessionID uint) error {
	return r.db.Model(&model.LoveMatch{}).Where("id = ?", id).Update("chat_session_id", sessionID).Error
}

func (r *loveMatchRepository) CountByUser(userID uint) (int64, error) {
	var count int64
	err := r.db.Model(&model.LoveMatch{}).Where(
		"(user_id_a = ? OR user_id_b = ?) AND status = ?",
		userID, userID, model.MatchStatusActive,
	).Count(&count).Error
	return count, err
}

func (r *loveMatchRepository) CountTodayByUser(userID uint) (int64, error) {
	var count int64
	err := r.db.Model(&model.LoveMatch{}).Where(
		"(user_id_a = ? OR user_id_b = ?) AND matched_at >= DATE_TRUNC('day', NOW())",
		userID, userID,
	).Count(&count).Error
	return count, err
}
