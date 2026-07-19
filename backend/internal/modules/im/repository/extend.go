// Package repository IM 中台扩展数据访问层
package repository

import (
	"wuchang-tongcheng/internal/modules/im/model"

	"gorm.io/gorm"
)

// IMExtendRepository IM 扩展仓储接口
type IMExtendRepository interface {
	// 消息已读
	CreateMessageRead(r *model.MessageRead) error
	FindMessageRead(messageID, userID uint) (*model.MessageRead, error)
	ListMessageReadsByUser(userID uint, page, pageSize int) ([]model.MessageRead, int64, error)

	// 会话用户
	CreateSessionUser(su *model.SessionUser) error
	FindSessionUser(sessionID string, userID uint) (*model.SessionUser, error)
	ListSessionUsers(sessionID string) ([]model.SessionUser, error)
	UpdateSessionUserFields(id uint, fields map[string]interface{}) error

	// 用户设置
	FindOrCreateUserSetting(userID uint, regionID uint) (*model.UserSetting, error)
	UpdateUserSettingFields(id uint, fields map[string]interface{}) error

	// 群组
	CreateGroup(g *model.Group) error
	FindGroupByID(groupID string) (*model.Group, error)
	ListGroupsByUser(userID uint, page, pageSize int) ([]model.Group, int64, error)
	UpdateGroupFields(id uint, fields map[string]interface{}) error

	// 群成员
	CreateGroupMember(m *model.GroupMember) error
	FindGroupMember(groupID string, userID uint) (*model.GroupMember, error)
	ListGroupMembers(groupID string) ([]model.GroupMember, error)
	UpdateGroupMemberFields(id uint, fields map[string]interface{}) error
	DeleteGroupMember(id uint) error
	CountGroupMembers(groupID string) (int64, error)

	// 统计
	StatTotalSessions() (int64, error)
	StatTotalMessages() (int64, error)
	StatTotalNotifications() (int64, error)
	StatUnreadNotifications(userID uint) (int64, error)
	StatTotalGroups() (int64, error)
	StatTotalGroupMembers() (int64, error)
}

type imExtendRepository struct {
	db *gorm.DB
}

// NewIMExtendRepository 创建扩展仓储实例
func NewIMExtendRepository(db *gorm.DB) IMExtendRepository {
	return &imExtendRepository{db: db}
}

// ===== 消息已读 =====

func (r *imExtendRepository) CreateMessageRead(read *model.MessageRead) error {
	return r.db.Create(read).Error
}

func (r *imExtendRepository) FindMessageRead(messageID, userID uint) (*model.MessageRead, error) {
	var m model.MessageRead
	if err := r.db.Where("message_id = ? AND user_id = ?", messageID, userID).First(&m).Error; err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *imExtendRepository) ListMessageReadsByUser(userID uint, page, pageSize int) ([]model.MessageRead, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	var list []model.MessageRead
	var total int64
	q := r.db.Model(&model.MessageRead{}).Where("user_id = ?", userID)
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := q.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// ===== 会话用户 =====

func (r *imExtendRepository) CreateSessionUser(su *model.SessionUser) error {
	return r.db.Create(su).Error
}

func (r *imExtendRepository) FindSessionUser(sessionID string, userID uint) (*model.SessionUser, error) {
	var su model.SessionUser
	if err := r.db.Where("session_id = ? AND user_id = ?", sessionID, userID).First(&su).Error; err != nil {
		return nil, err
	}
	return &su, nil
}

func (r *imExtendRepository) ListSessionUsers(sessionID string) ([]model.SessionUser, error) {
	var list []model.SessionUser
	if err := r.db.Where("session_id = ? AND status = ?", sessionID, 1).Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *imExtendRepository) UpdateSessionUserFields(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.SessionUser{}).Where("id = ?", id).Updates(fields).Error
}

// ===== 用户设置 =====

func (r *imExtendRepository) FindOrCreateUserSetting(userID uint, regionID uint) (*model.UserSetting, error) {
	var us model.UserSetting
	err := r.db.Where("user_id = ?", userID).First(&us).Error
	if err == nil {
		return &us, nil
	}
	if err != gorm.ErrRecordNotFound {
		return nil, err
	}
	us = model.UserSetting{
		RegionID:     regionID,
		UserID:       userID,
		OnlineStatus: model.OnlineStatusOnline,
		Extra:        "{}",
	}
	if err := r.db.Create(&us).Error; err != nil {
		return nil, err
	}
	return &us, nil
}

func (r *imExtendRepository) UpdateUserSettingFields(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.UserSetting{}).Where("id = ?", id).Updates(fields).Error
}

// ===== 群组 =====

func (r *imExtendRepository) CreateGroup(g *model.Group) error {
	return r.db.Create(g).Error
}

func (r *imExtendRepository) FindGroupByID(groupID string) (*model.Group, error) {
	var g model.Group
	if err := r.db.Where("group_id = ?", groupID).First(&g).Error; err != nil {
		return nil, err
	}
	return &g, nil
}

func (r *imExtendRepository) ListGroupsByUser(userID uint, page, pageSize int) ([]model.Group, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	var list []model.Group
	var total int64
	q := r.db.Model(&model.Group{}).
		Where("id IN (SELECT id FROM im_group_members WHERE user_id = ? AND status = 1)", userID)
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := q.Order("updated_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *imExtendRepository) UpdateGroupFields(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.Group{}).Where("id = ?", id).Updates(fields).Error
}

// ===== 群成员 =====

func (r *imExtendRepository) CreateGroupMember(m *model.GroupMember) error {
	return r.db.Create(m).Error
}

func (r *imExtendRepository) FindGroupMember(groupID string, userID uint) (*model.GroupMember, error) {
	var m model.GroupMember
	if err := r.db.Where("group_id = ? AND user_id = ?", groupID, userID).First(&m).Error; err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *imExtendRepository) ListGroupMembers(groupID string) ([]model.GroupMember, error) {
	var list []model.GroupMember
	if err := r.db.Where("group_id = ? AND status = ?", groupID, 1).Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *imExtendRepository) UpdateGroupMemberFields(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.GroupMember{}).Where("id = ?", id).Updates(fields).Error
}

func (r *imExtendRepository) DeleteGroupMember(id uint) error {
	return r.db.Delete(&model.GroupMember{}, id).Error
}

func (r *imExtendRepository) CountGroupMembers(groupID string) (int64, error) {
	var count int64
	if err := r.db.Model(&model.GroupMember{}).
		Where("group_id = ? AND status = ?", groupID, 1).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

// ===== 统计 =====

func (r *imExtendRepository) StatTotalSessions() (int64, error) {
	var count int64
	err := r.db.Model(&model.Session{}).Count(&count).Error
	return count, err
}

func (r *imExtendRepository) StatTotalMessages() (int64, error) {
	var count int64
	err := r.db.Model(&model.Message{}).Count(&count).Error
	return count, err
}

func (r *imExtendRepository) StatTotalNotifications() (int64, error) {
	var count int64
	err := r.db.Model(&model.SystemNotification{}).Count(&count).Error
	return count, err
}

func (r *imExtendRepository) StatUnreadNotifications(userID uint) (int64, error) {
	var count int64
	err := r.db.Model(&model.SystemNotification{}).
		Where("user_id = ? AND is_read = ?", userID, false).Count(&count).Error
	return count, err
}

func (r *imExtendRepository) StatTotalGroups() (int64, error) {
	var count int64
	err := r.db.Model(&model.Group{}).Where("status = ?", 1).Count(&count).Error
	return count, err
}

func (r *imExtendRepository) StatTotalGroupMembers() (int64, error) {
	var count int64
	err := r.db.Model(&model.GroupMember{}).Where("status = ?", 1).Count(&count).Error
	return count, err
}
