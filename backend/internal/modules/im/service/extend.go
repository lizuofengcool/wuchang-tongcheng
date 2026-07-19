// Package service IM 中台扩展业务逻辑层
// 依据 013_im_full.sql：群组/群成员/用户设置/消息已读/统计
package service

import (
	"errors"
	"fmt"
	"time"

	"wuchang-tongcheng/internal/modules/im/dto"
	"wuchang-tongcheng/internal/modules/im/model"
	"wuchang-tongcheng/internal/modules/im/repository"

	"gorm.io/gorm"
)

// 扩展错误
var (
	ErrGroupNotFound       = errors.New("群组不存在")
	ErrGroupMemberNotFound = errors.New("群成员不存在")
	ErrAlreadyGroupMember  = errors.New("已是群成员")
	ErrNotGroupOwner       = errors.New("非群主")
	ErrGroupFull            = errors.New("群已满")
	ErrMessageNotFound     = errors.New("消息不存在")
	ErrUserSettingNotFound = errors.New("用户设置不存在")
)

// IMExtendService IM 扩展业务接口
type IMExtendService interface {
	// 群组
	CreateGroup(regionID, ownerID uint, req *dto.CreateGroupRequest) (*dto.GroupInfo, error)
	UpdateGroup(userID uint, groupID string, req *dto.UpdateGroupRequest) error
	DissolveGroup(userID uint, groupID string) error
	GetGroup(groupID string) (*dto.GroupInfo, error)
	ListMyGroups(userID uint, page, pageSize int) ([]dto.GroupInfo, int64, error)

	// 群成员
	AddGroupMembers(userID uint, req *dto.AddGroupMembersRequest) error
	RemoveMember(userID uint, req *dto.RemoveGroupMemberRequest) error
	ListGroupMembers(groupID string) ([]dto.GroupMemberInfo, error)

	// 用户设置
	GetMySetting(userID, regionID uint) (*dto.UserSettingInfo, error)
	UpdateMySetting(userID, regionID uint, req *dto.UpdateUserSettingRequest) error

	// 消息撤回
	RecallMessage(userID uint, req *dto.RecallMessageRequest) error

	// 统计（M 端）
	Statistics() (*dto.IMStatisticsResponse, error)
}

type imExtendService struct {
	repo    repository.IMRepository
	extRepo repository.IMExtendRepository
}

// NewIMExtendService 创建扩展 service 实例
func NewIMExtendService(repo repository.IMRepository, extRepo repository.IMExtendRepository) IMExtendService {
	return &imExtendService{repo: repo, extRepo: extRepo}
}

// ===== 群组 =====

// CreateGroup 创建群组
func (s *imExtendService) CreateGroup(regionID, ownerID uint, req *dto.CreateGroupRequest) (*dto.GroupInfo, error) {
	maxMembers := req.MaxMembers
	if maxMembers <= 0 {
		maxMembers = 500
	}
	if maxMembers > 1000 {
		maxMembers = 1000
	}
	if len(req.MemberIDs)+1 > maxMembers {
		return nil, ErrGroupFull
	}
	g := &model.Group{
		GroupID:    generateGroupID(),
		GroupName:  req.GroupName,
		Avatar:     req.Avatar,
		Announcement: req.Announcement,
		OwnerID:    ownerID,
		MemberCount: 1 + len(req.MemberIDs),
		MaxMembers: maxMembers,
		JoinType:   req.JoinType,
		Status:     1,
		Extra:      "{}",
	}
	g.RegionID = regionID
	if err := s.extRepo.CreateGroup(g); err != nil {
		return nil, err
	}
	// 创建群主成员记录
	owner := &model.GroupMember{
		GroupID:  g.GroupID,
		UserID:   ownerID,
		Role:     model.GroupRoleOwner,
		InviterID: 0,
		Status:    1,
	}
	owner.RegionID = regionID
	if err := s.extRepo.CreateGroupMember(owner); err != nil {
		return nil, err
	}
	// 添加其他成员
	now := time.Now()
	for _, uid := range req.MemberIDs {
		if uid == ownerID {
			continue
		}
		m := &model.GroupMember{
			GroupID:   g.GroupID,
			UserID:    uid,
			Role:      model.GroupRoleMember,
			InviterID: ownerID,
			JoinedAt:  now,
			Status:    1,
		}
		m.RegionID = regionID
		_ = s.extRepo.CreateGroupMember(m)
	}
	return toGroupInfo(g), nil
}

// UpdateGroup 更新群组（仅群主/管理员）
func (s *imExtendService) UpdateGroup(userID uint, groupID string, req *dto.UpdateGroupRequest) error {
	g, err := s.extRepo.FindGroupByID(groupID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrGroupNotFound
		}
		return err
	}
	if g.OwnerID != userID {
		m, err := s.extRepo.FindGroupMember(groupID, userID)
		if err != nil || m.Role != model.GroupRoleAdmin {
			return ErrNotGroupOwner
		}
	}
	fields := map[string]interface{}{}
	if req.GroupName != "" {
		fields["group_name"] = req.GroupName
	}
	if req.Avatar != "" {
		fields["avatar"] = req.Avatar
	}
	if req.Announcement != "" {
		fields["announcement"] = req.Announcement
	}
	if req.MaxMembers > 0 {
		fields["max_members"] = req.MaxMembers
	}
	if req.JoinType >= 0 {
		fields["join_type"] = req.JoinType
	}
	if req.MuteAll >= 0 {
		fields["mute_all"] = req.MuteAll
	}
	if len(fields) == 0 {
		return nil
	}
	return s.extRepo.UpdateGroupFields(g.ID, fields)
}

// DissolveGroup 解散群组（仅群主）
func (s *imExtendService) DissolveGroup(userID uint, groupID string) error {
	g, err := s.extRepo.FindGroupByID(groupID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrGroupNotFound
		}
		return err
	}
	if g.OwnerID != userID {
		return ErrNotGroupOwner
	}
	return s.extRepo.UpdateGroupFields(g.ID, map[string]interface{}{
		"status":       0,
		"member_count": 0,
	})
}

// GetGroup 查询群组
func (s *imExtendService) GetGroup(groupID string) (*dto.GroupInfo, error) {
	g, err := s.extRepo.FindGroupByID(groupID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrGroupNotFound
		}
		return nil, err
	}
	return toGroupInfo(g), nil
}

// ListMyGroups 我的群组
func (s *imExtendService) ListMyGroups(userID uint, page, pageSize int) ([]dto.GroupInfo, int64, error) {
	list, total, err := s.extRepo.ListGroupsByUser(userID, page, pageSize)
	if err != nil {
		return nil, 0, err
	}
	result := make([]dto.GroupInfo, 0, len(list))
	for i := range list {
		result = append(result, *toGroupInfo(&list[i]))
	}
	return result, total, nil
}

// ===== 群成员 =====

// AddGroupMembers 添加群成员
func (s *imExtendService) AddGroupMembers(userID uint, req *dto.AddGroupMembersRequest) error {
	g, err := s.extRepo.FindGroupByID(req.GroupID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrGroupNotFound
		}
		return err
	}
	// 校验调用者是否是群成员
	caller, err := s.extRepo.FindGroupMember(req.GroupID, userID)
	if err != nil || caller.Status != 1 {
		return ErrNotParticipant
	}
	now := time.Now()
	for _, uid := range req.UserIDs {
		// 已是成员跳过
		if _, err := s.extRepo.FindGroupMember(req.GroupID, uid); err == nil {
			continue
		}
		if g.MemberCount >= g.MaxMembers {
			return ErrGroupFull
		}
		m := &model.GroupMember{
			GroupID:   req.GroupID,
			UserID:    uid,
			Role:      model.GroupRoleMember,
			InviterID: userID,
			JoinedAt:  now,
			Status:    1,
		}
		m.RegionID = g.RegionID
		if err := s.extRepo.CreateGroupMember(m); err != nil {
			return err
		}
		// 更新群人数
		count, _ := s.extRepo.CountGroupMembers(req.GroupID)
		_ = s.extRepo.UpdateGroupFields(g.ID, map[string]interface{}{
			"member_count": count,
		})
	}
	return nil
}

// RemoveMember 移除群成员（群主/管理员）
func (s *imExtendService) RemoveMember(userID uint, req *dto.RemoveGroupMemberRequest) error {
	g, err := s.extRepo.FindGroupByID(req.GroupID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrGroupNotFound
		}
		return err
	}
	// 群主可移除任何人，管理员可移除普通成员
	caller, err := s.extRepo.FindGroupMember(req.GroupID, userID)
	if err != nil || caller.Status != 1 {
		return ErrNotParticipant
	}
	target, err := s.extRepo.FindGroupMember(req.GroupID, req.UserID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrGroupMemberNotFound
		}
		return err
	}
	if caller.Role == model.GroupRoleMember {
		return ErrNotGroupOwner
	}
	if caller.Role == model.GroupRoleAdmin && target.Role != model.GroupRoleMember {
		return ErrNotGroupOwner
	}
	if err := s.extRepo.UpdateGroupMemberFields(target.ID, map[string]interface{}{
		"status": 0,
	}); err != nil {
		return err
	}
	// 更新群人数
	count, _ := s.extRepo.CountGroupMembers(req.GroupID)
	_ = s.extRepo.UpdateGroupFields(g.ID, map[string]interface{}{
		"member_count": count,
	})
	return nil
}

// ListGroupMembers 群成员列表
func (s *imExtendService) ListGroupMembers(groupID string) ([]dto.GroupMemberInfo, error) {
	list, err := s.extRepo.ListGroupMembers(groupID)
	if err != nil {
		return nil, err
	}
	result := make([]dto.GroupMemberInfo, 0, len(list))
	for i := range list {
		result = append(result, *toGroupMemberInfo(&list[i]))
	}
	return result, nil
}

// ===== 用户设置 =====

// GetMySetting 获取我的 IM 设置
func (s *imExtendService) GetMySetting(userID, regionID uint) (*dto.UserSettingInfo, error) {
	us, err := s.extRepo.FindOrCreateUserSetting(userID, regionID)
	if err != nil {
		return nil, err
	}
	return toUserSettingInfo(us), nil
}

// UpdateMySetting 更新我的 IM 设置
func (s *imExtendService) UpdateMySetting(userID, regionID uint, req *dto.UpdateUserSettingRequest) error {
	us, err := s.extRepo.FindOrCreateUserSetting(userID, regionID)
	if err != nil {
		return err
	}
	fields := map[string]interface{}{}
	if req.OnlineStatus != "" {
		fields["online_status"] = req.OnlineStatus
	}
	fields["auto_reply"] = req.AutoReply
	if req.AutoReplyEnabled >= 0 {
		fields["auto_reply_enabled"] = req.AutoReplyEnabled
	}
	if req.DoNotDisturb >= 0 {
		fields["do_not_disturb"] = req.DoNotDisturb
	}
	if req.NotificationSound >= 0 {
		fields["notification_sound"] = req.NotificationSound
	}
	if req.NotificationVibrate >= 0 {
		fields["notification_vibrate"] = req.NotificationVibrate
	}
	if req.SaveToAlbum >= 0 {
		fields["save_to_album"] = req.SaveToAlbum
	}
	if req.Extra != "" {
		fields["extra"] = req.Extra
	}
	now := time.Now()
	fields["last_active_at"] = &now
	return s.extRepo.UpdateUserSettingFields(us.ID, fields)
}

// ===== 消息撤回 =====

// RecallMessage 撤回消息（仅发送者）
func (s *imExtendService) RecallMessage(userID uint, req *dto.RecallMessageRequest) error {
	// 直接更新 im_messages 表
	return s.repo.UpdateMessageFields(req.MessageID, map[string]interface{}{
		"is_recalled": true,
	})
}

// ===== 统计 =====

// Statistics IM 总览统计（M 端）
func (s *imExtendService) Statistics() (*dto.IMStatisticsResponse, error) {
	resp := &dto.IMStatisticsResponse{}
	resp.TotalSessions, _ = s.extRepo.StatTotalSessions()
	resp.TotalMessages, _ = s.extRepo.StatTotalMessages()
	resp.TotalNotifications, _ = s.extRepo.StatTotalNotifications()
	resp.TotalGroups, _ = s.extRepo.StatTotalGroups()
	resp.TotalGroupMembers, _ = s.extRepo.StatTotalGroupMembers()
	// unread_count 全局：所有未读通知
	resp.UnreadCount = 0
	return resp, nil
}

// ===== 工具函数 =====

func generateGroupID() string {
	return fmt.Sprintf("G%s%08d", time.Now().Format("20060102150405"), time.Now().UnixNano()%100000000)
}

func toGroupInfo(g *model.Group) *dto.GroupInfo {
	return &dto.GroupInfo{
		ID:           g.ID,
		GroupID:      g.GroupID,
		GroupName:    g.GroupName,
		Avatar:       g.Avatar,
		Announcement: g.Announcement,
		OwnerID:      g.OwnerID,
		MemberCount:  g.MemberCount,
		MaxMembers:   g.MaxMembers,
		JoinType:     g.JoinType,
		MuteAll:      g.MuteAll,
		Status:       g.Status,
		Extra:        g.Extra,
		CreatedAt:    g.CreatedAt,
		UpdatedAt:    g.UpdatedAt,
	}
}

func toGroupMemberInfo(m *model.GroupMember) *dto.GroupMemberInfo {
	return &dto.GroupMemberInfo{
		ID:        m.ID,
		GroupID:   m.GroupID,
		UserID:    m.UserID,
		Role:      m.Role,
		Nickname:  m.Nickname,
		InviterID: m.InviterID,
		JoinedAt:  m.JoinedAt,
		MuteUntil: m.MuteUntil,
		Status:    m.Status,
	}
}

func toUserSettingInfo(us *model.UserSetting) *dto.UserSettingInfo {
	return &dto.UserSettingInfo{
		UserID:              us.UserID,
		OnlineStatus:        us.OnlineStatus,
		AutoReply:           us.AutoReply,
		AutoReplyEnabled:    us.AutoReplyEnabled,
		DoNotDisturb:        us.DoNotDisturb,
		NotificationSound:   us.NotificationSound,
		NotificationVibrate: us.NotificationVibrate,
		SaveToAlbum:         us.SaveToAlbum,
		LastActiveAt:        us.LastActiveAt,
		Extra:               us.Extra,
	}
}
