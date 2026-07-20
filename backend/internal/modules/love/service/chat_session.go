// Package service love 相亲交友业务逻辑层 - 聊天会话
// 依据需求文档 1.10：4 维数据隔离（region_id + user_id）
package service

import (
	"errors"
	"fmt"
	"time"

	"wuchang-tongcheng/internal/modules/love/dto"
	"wuchang-tongcheng/internal/modules/love/model"
	"wuchang-tongcheng/internal/modules/love/repository"
	"wuchang-tongcheng/internal/pkg/utils"
)

var (
	ErrLoveChatSessionNotFound     = errors.New("聊天会话不存在")
	ErrLoveChatSessionNoPermission = errors.New("无权操作此会话")
	ErrLoveChatSessionInvalidAction = errors.New("无效的会话操作")
	ErrLoveChatSessionDissolved    = errors.New("会话已解散")
)

// LoveChatSessionService 聊天会话业务接口
type LoveChatSessionService interface {
	// 内部调用（match service 创建匹配后调用）
	CreateByMatch(matchID uint, userIDA, userIDB, loveIDA, loveIDB uint, nicknameA, nicknameB, avatarA, avatarB string) (*dto.LoveChatSessionInfo, error)
	GetByID(id uint, userID uint) (*dto.LoveChatSessionInfo, error)
	GetByMatchID(matchID uint) (*dto.LoveChatSessionInfo, error)
	List(req *dto.LoveChatSessionListRequest, userID uint) (*utils.Pagination, []dto.LoveChatSessionInfo, error)
	Action(req *dto.LoveChatSessionActionRequest, userID uint) error
	MarkRead(id uint, userID uint) error
	UpdateLastMessage(id uint, msgID uint, content, msgType string, senderID uint) error
	IncrMessageCount(id uint) error
	IncrGiftCount(id uint) error
	CountActiveByUser(userID uint) (int64, error)
}

type loveChatSessionService struct {
	repo repository.LoveChatSessionRepository
}

// NewLoveChatSessionService 创建聊天会话 service
func NewLoveChatSessionService(repo repository.LoveChatSessionRepository) LoveChatSessionService {
	return &loveChatSessionService{repo: repo}
}

// chatSessionStatusText 状态文本
func chatSessionStatusText(s int) string {
	switch s {
	case model.ChatSessionStatusActive:
		return "活跃"
	case model.ChatSessionStatusMuted:
		return "免打扰"
	case model.ChatSessionStatusDissolved:
		return "已解散"
	case model.ChatSessionStatusBlocked:
		return "已拉黑"
	}
	return ""
}

// toLoveChatSessionInfo model -> dto
// 自动判断当前用户视角，填充 PartnerNickname 等
func toLoveChatSessionInfo(s *model.LoveChatSession, viewerUserID uint) dto.LoveChatSessionInfo {
	info := dto.LoveChatSessionInfo{
		ID:                 s.ID,
		SessionNo:          s.SessionNo,
		MatchID:            s.MatchID,
		UserIDA:            s.UserIDA,
		UserIDB:            s.UserIDB,
		LoveIDA:            s.LoveIDA,
		LoveIDB:            s.LoveIDB,
		NicknameA:          s.NicknameA,
		NicknameB:          s.NicknameB,
		AvatarA:            s.AvatarA,
		AvatarB:            s.AvatarB,
		LastMessageID:      s.LastMessageID,
		LastMessageContent: s.LastMessageContent,
		LastMessageType:    s.LastMessageType,
		LastMessageAt:      s.LastMessageAt,
		LastSenderID:       s.LastSenderID,
		Status:             s.Status,
		StatusText:         chatSessionStatusText(s.Status),
		MessageCount:       s.MessageCount,
		GiftCount:          s.GiftCount,
		DissolvedAt:        s.DissolvedAt,
		DissolveReason:     s.DissolveReason,
		CreatedAt:          s.CreatedAt,
		UpdatedAt:          s.UpdatedAt,
	}

	// 视角相关：未读数 / mute / pin / 对方信息
	if viewerUserID == s.UserIDA {
		info.UnreadCount = s.UnreadCountA
		info.Muted = s.MutedA
		info.Pinned = s.PinnedA
		info.PartnerUserID = s.UserIDB
		info.PartnerLoveID = s.LoveIDB
		info.PartnerNickname = s.NicknameB
		info.PartnerAvatar = s.AvatarB
	} else if viewerUserID == s.UserIDB {
		info.UnreadCount = s.UnreadCountB
		info.Muted = s.MutedB
		info.Pinned = s.PinnedB
		info.PartnerUserID = s.UserIDA
		info.PartnerLoveID = s.LoveIDA
		info.PartnerNickname = s.NicknameA
		info.PartnerAvatar = s.AvatarA
	}
	return info
}

// CreateByMatch 匹配成功后创建会话（如已存在则返回已存在的）
func (s *loveChatSessionService) CreateByMatch(matchID uint, userIDA, userIDB, loveIDA, loveIDB uint, nicknameA, nicknameB, avatarA, avatarB string) (*dto.LoveChatSessionInfo, error) {
	// 已存在直接返回
	if existing, err := s.repo.FindByMatchID(matchID); err == nil && existing != nil {
		info := toLoveChatSessionInfo(existing, 0)
		return &info, nil
	}
	session := &model.LoveChatSession{
		SessionNo:  fmt.Sprintf("LOVE-CHAT-%d-%d", matchID, time.Now().UnixNano()/1e6),
		MatchID:    matchID,
		UserIDA:    userIDA,
		UserIDB:    userIDB,
		LoveIDA:    loveIDA,
		LoveIDB:    loveIDB,
		NicknameA:  nicknameA,
		NicknameB:  nicknameB,
		AvatarA:    avatarA,
		AvatarB:    avatarB,
		Status:     model.ChatSessionStatusActive,
	}
	if err := s.repo.Create(session); err != nil {
		return nil, err
	}
	info := toLoveChatSessionInfo(session, 0)
	return &info, nil
}

func (s *loveChatSessionService) GetByID(id uint, userID uint) (*dto.LoveChatSessionInfo, error) {
	session, err := s.repo.FindByID(id)
	if err != nil {
		return nil, ErrLoveChatSessionNotFound
	}
	if session.UserIDA != userID && session.UserIDB != userID {
		return nil, ErrLoveChatSessionNoPermission
	}
	info := toLoveChatSessionInfo(session, userID)
	return &info, nil
}

func (s *loveChatSessionService) GetByMatchID(matchID uint) (*dto.LoveChatSessionInfo, error) {
	session, err := s.repo.FindByMatchID(matchID)
	if err != nil {
		return nil, ErrLoveChatSessionNotFound
	}
	info := toLoveChatSessionInfo(session, 0)
	return &info, nil
}

func (s *loveChatSessionService) List(req *dto.LoveChatSessionListRequest, userID uint) (*utils.Pagination, []dto.LoveChatSessionInfo, error) {
	list, total, err := s.repo.ListByUser(userID, &req.Pagination)
	if err != nil {
		return nil, nil, err
	}
	infos := make([]dto.LoveChatSessionInfo, 0, len(list))
	for i := range list {
		infos = append(infos, toLoveChatSessionInfo(&list[i], userID))
	}
	req.Pagination.Total = total
	return &req.Pagination, infos, nil
}

// Action 会话操作（mute/unmute/pin/unpin/delete/dissolve）
func (s *loveChatSessionService) Action(req *dto.LoveChatSessionActionRequest, userID uint) error {
	session, err := s.repo.FindByID(req.ID)
	if err != nil {
		return ErrLoveChatSessionNotFound
	}
	if session.UserIDA != userID && session.UserIDB != userID {
		return ErrLoveChatSessionNoPermission
	}
	if session.Status == model.ChatSessionStatusDissolved {
		return ErrLoveChatSessionDissolved
	}
	side := "a"
	if session.UserIDB == userID {
		side = "b"
	}
	switch req.Action {
	case "mute":
		return s.repo.SetMuted(req.ID, side, true)
	case "unmute":
		return s.repo.SetMuted(req.ID, side, false)
	case "pin":
		return s.repo.SetPinned(req.ID, side, true)
	case "unpin":
		return s.repo.SetPinned(req.ID, side, false)
	case "delete":
		// 软删除：仅当前用户视角不可见
		return s.repo.SetDeleted(req.ID, side, true)
	case "dissolve":
		// 解散：双方都不可见，会话终止
		return s.repo.Dissolve(req.ID, userID, req.Reason)
	}
	return ErrLoveChatSessionInvalidAction
}

// MarkRead 标记已读（重置当前用户视角未读数）
func (s *loveChatSessionService) MarkRead(id uint, userID uint) error {
	session, err := s.repo.FindByID(id)
	if err != nil {
		return ErrLoveChatSessionNotFound
	}
	if session.UserIDA != userID && session.UserIDB != userID {
		return ErrLoveChatSessionNoPermission
	}
	side := "a"
	if session.UserIDB == userID {
		side = "b"
	}
	return s.repo.ResetUnreadCount(id, side)
}

// UpdateLastMessage 由消息发送方调用，更新会话最后一条消息
func (s *loveChatSessionService) UpdateLastMessage(id uint, msgID uint, content, msgType string, senderID uint) error {
	if err := s.repo.UpdateLastMessage(id, msgID, content, msgType, senderID); err != nil {
		return err
	}
	// 给接收方未读数 +1
	session, err := s.repo.FindByID(id)
	if err != nil {
		return nil
	}
	side := "b"
	if session.UserIDB == senderID {
		side = "a"
	}
	_ = s.repo.IncrUnreadCount(id, side)
	return nil
}

func (s *loveChatSessionService) IncrMessageCount(id uint) error {
	return s.repo.IncrMessageCount(id)
}

func (s *loveChatSessionService) IncrGiftCount(id uint) error {
	return s.repo.IncrGiftCount(id)
}

func (s *loveChatSessionService) CountActiveByUser(userID uint) (int64, error) {
	return s.repo.CountActiveByUser(userID)
}
