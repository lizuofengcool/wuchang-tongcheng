// Package service IM 消息中台精简版业务逻辑层
// 依据 ershou 模块依赖：私聊 + 系统通知 + 隐私号码
// 暴露 IMService 接口供其他模块直接 import 调用（不通过 HTTP）
package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"wuchang-tongcheng/internal/modules/im/dto"
	"wuchang-tongcheng/internal/modules/im/model"
	"wuchang-tongcheng/internal/modules/im/repository"

	"gorm.io/gorm"
)

var (
	ErrSessionNotFound  = errors.New("会话不存在")
	ErrNotParticipant   = errors.New("非会话参与者")
	ErrPrivacyNotFound  = errors.New("隐私号码不存在")
	ErrPrivacyUnbound   = errors.New("隐私号码已解绑")
)

// IMService IM 中台业务接口
// 暴露给其他模块直接 import 调用，不通过 HTTP
type IMService interface {
	// 会话
	CreateSession(regionID uint, userID uint, req *dto.CreateSessionRequest) (*dto.SessionInfo, error)
	GetSession(sessionID string) (*dto.SessionInfo, error)
	ListSessions(userID uint, page, pageSize int) ([]dto.SessionInfo, int64, error)

	// 消息
	SendMessage(regionID uint, userID uint, req *dto.SendMessageRequest) (*dto.MessageInfo, error)
	GetHistory(sessionID string, userID uint, page, pageSize int) ([]dto.MessageInfo, int64, error)
	MarkRead(sessionID string, userID uint) error

	// 系统通知
	PushNotification(regionID uint, req *dto.PushNotificationRequest) (*dto.NotificationInfo, error)
	ListNotifications(userID uint, page, pageSize int) ([]dto.NotificationInfo, int64, error)
	ListUnreadNotifications(userID uint) ([]dto.NotificationInfo, int64, error)
	MarkAllNotificationsRead(userID uint) error

	// 隐私号码
	BindPrivacyNumber(regionID uint, req *dto.BindPrivacyNumberRequest) (*dto.PrivacyNumberInfo, error)
	UnbindPrivacyNumber(req *dto.UnbindPrivacyNumberRequest) error
}

type imService struct {
	repo repository.IMRepository
}

// NewIMService 创建 service 实例
func NewIMService(repo repository.IMRepository) IMService {
	return &imService{repo: repo}
}

// ===== 会话 =====

// CreateSession 创建会话
func (s *imService) CreateSession(regionID uint, userID uint, req *dto.CreateSessionRequest) (*dto.SessionInfo, error) {
	// 校验：发起人必须在参与者中
	if !contains(req.Participants, userID) {
		return nil, ErrNotParticipant
	}

	participantsJSON, _ := json.Marshal(req.Participants)
	session := &model.Session{
		SessionID:    generateSessionID(),
		SessionType:  req.SessionType,
		Participants: string(participantsJSON),
		LastMessage:  "{}",
		UnreadCount:  "{}",
		Status:       1,
	}
	session.RegionID = regionID

	if err := s.repo.CreateSession(session); err != nil {
		return nil, err
	}
	return toSessionInfo(session), nil
}

// GetSession 查询会话
func (s *imService) GetSession(sessionID string) (*dto.SessionInfo, error) {
	session, err := s.repo.FindSessionByID(sessionID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrSessionNotFound
		}
		return nil, err
	}
	return toSessionInfo(session), nil
}

// ListSessions 用户的会话列表
func (s *imService) ListSessions(userID uint, page, pageSize int) ([]dto.SessionInfo, int64, error) {
	list, total, err := s.repo.ListSessionsByUser(userID, page, pageSize)
	if err != nil {
		return nil, 0, err
	}
	result := make([]dto.SessionInfo, 0, len(list))
	for i := range list {
		result = append(result, *toSessionInfo(&list[i]))
	}
	return result, total, nil
}

// ===== 消息 =====

// SendMessage 发送消息
// 流程：校验会话 → 写入消息 → 更新会话 last_message
func (s *imService) SendMessage(regionID uint, userID uint, req *dto.SendMessageRequest) (*dto.MessageInfo, error) {
	session, err := s.repo.FindSessionByID(req.SessionID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrSessionNotFound
		}
		return nil, err
	}

	// 校验：发送者必须是会话参与者
	var participants []uint
	if err := json.Unmarshal([]byte(session.Participants), &participants); err != nil {
		return nil, err
	}
	if !contains(participants, userID) {
		return nil, ErrNotParticipant
	}

	msg := &model.Message{
		SessionID:  req.SessionID,
		SenderID:   userID,
		MsgType:    req.MsgType,
		Content:    req.Content,
		Extra:      defaultJSON(req.Extra),
		ReadStatus: fmt.Sprintf(`{"%d":true}`, userID),
	}
	msg.RegionID = regionID

	if err := s.repo.CreateMessage(msg); err != nil {
		return nil, err
	}

	// 更新会话 last_message 和未读数
	now := time.Now()
	lastMsg, _ := json.Marshal(map[string]interface{}{
		"msg_id":   msg.ID,
		"sender":   msg.SenderID,
		"msg_type": msg.MsgType,
		"content":  truncate(msg.Content, 100),
		"at":       now,
	})

	// 更新其他参与者未读数 +1
	unreadMap := map[uint]int{}
	if err := json.Unmarshal([]byte(session.UnreadCount), &unreadMap); err != nil {
		unreadMap = map[uint]int{}
	}
	for _, uid := range participants {
		if uid != userID {
			unreadMap[uid]++
		}
	}
	unreadJSON, _ := json.Marshal(unreadMap)

	_ = s.repo.UpdateSessionFields(session.ID, map[string]interface{}{
		"last_message":    string(lastMsg),
		"last_message_at": &now,
		"unread_count":    string(unreadJSON),
	})

	return toMessageInfo(msg), nil
}

// GetHistory 获取历史消息（分页）
func (s *imService) GetHistory(sessionID string, userID uint, page, pageSize int) ([]dto.MessageInfo, int64, error) {
	list, total, err := s.repo.ListMessages(sessionID, page, pageSize)
	if err != nil {
		return nil, 0, err
	}
	result := make([]dto.MessageInfo, 0, len(list))
	for i := range list {
		result = append(result, *toMessageInfo(&list[i]))
	}
	return result, total, nil
}

// MarkRead 标记会话已读（清零该用户的未读数）
func (s *imService) MarkRead(sessionID string, userID uint) error {
	session, err := s.repo.FindSessionByID(sessionID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrSessionNotFound
		}
		return err
	}
	// 更新未读数
	unreadMap := map[uint]int{}
	if err := json.Unmarshal([]byte(session.UnreadCount), &unreadMap); err != nil {
		unreadMap = map[uint]int{}
	}
	unreadMap[userID] = 0
	unreadJSON, _ := json.Marshal(unreadMap)
	return s.repo.UpdateSessionFields(session.ID, map[string]interface{}{
		"unread_count": string(unreadJSON),
	})
}

// ===== 系统通知 =====

// PushNotification 推送系统通知
// 其他模块（pay/ershou 等）可直接 import 调用此方法
func (s *imService) PushNotification(regionID uint, req *dto.PushNotificationRequest) (*dto.NotificationInfo, error) {
	n := &model.SystemNotification{
		UserID:     req.UserID,
		NotifyType: req.NotifyType,
		Title:      req.Title,
		Content:    req.Content,
		JumpURL:    req.JumpURL,
		Extra:      defaultJSON(req.Extra),
		IsRead:     false,
	}
	n.RegionID = regionID
	if err := s.repo.CreateNotification(n); err != nil {
		return nil, err
	}
	return toNotificationInfo(n), nil
}

// ListNotifications 用户通知列表
func (s *imService) ListNotifications(userID uint, page, pageSize int) ([]dto.NotificationInfo, int64, error) {
	list, total, err := s.repo.ListNotifications(userID, page, pageSize)
	if err != nil {
		return nil, 0, err
	}
	result := make([]dto.NotificationInfo, 0, len(list))
	for i := range list {
		result = append(result, *toNotificationInfo(&list[i]))
	}
	return result, total, nil
}

// ListUnreadNotifications 未读通知
func (s *imService) ListUnreadNotifications(userID uint) ([]dto.NotificationInfo, int64, error) {
	list, total, err := s.repo.ListUnreadNotifications(userID)
	if err != nil {
		return nil, 0, err
	}
	result := make([]dto.NotificationInfo, 0, len(list))
	for i := range list {
		result = append(result, *toNotificationInfo(&list[i]))
	}
	return result, total, nil
}

// MarkAllNotificationsRead 标记全部通知已读
func (s *imService) MarkAllNotificationsRead(userID uint) error {
	return s.repo.MarkAllRead(userID)
}

// ===== 隐私号码 =====

// BindPrivacyNumber 绑定隐私号码
func (s *imService) BindPrivacyNumber(regionID uint, req *dto.BindPrivacyNumberRequest) (*dto.PrivacyNumberInfo, error) {
	p := &model.PrivacyNumber{
		PrivacyNo:   generatePrivacyNo(),
		RealNoA:     req.RealNoA,
		RealNoB:     req.RealNoB,
		UserIDA:     req.UserIDA,
		UserIDB:     req.UserIDB,
		BizModule:   req.BizModule,
		BizID:       req.BizID,
		CallRecords: "[]",
		BoundAt:     time.Now(),
		Status:      model.PrivacyStatusBound,
	}
	p.RegionID = regionID
	if err := s.repo.CreatePrivacyNumber(p); err != nil {
		return nil, err
	}
	return toPrivacyNumberInfo(p), nil
}

// UnbindPrivacyNumber 解绑隐私号码
func (s *imService) UnbindPrivacyNumber(req *dto.UnbindPrivacyNumberRequest) error {
	p, err := s.repo.FindPrivacyNumber(req.PrivacyNo)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrPrivacyNotFound
		}
		return err
	}
	if p.Status == model.PrivacyStatusUnbound {
		return ErrPrivacyUnbound
	}
	now := time.Now()
	return s.repo.UpdatePrivacyNumberFields(p.ID, map[string]interface{}{
		"status":     model.PrivacyStatusUnbound,
		"unbound_at": &now,
	})
}

// ===== 工具函数 =====

func generateSessionID() string {
	return fmt.Sprintf("S%s%08d", time.Now().Format("20060102150405"), time.Now().UnixNano()%100000000)
}

func generatePrivacyNo() string {
	// 170 开头的虚拟号码 + 8 位随机
	return fmt.Sprintf("170%08d", time.Now().UnixNano()%100000000)
}

func contains(arr []uint, v uint) bool {
	for _, a := range arr {
		if a == v {
			return true
		}
	}
	return false
}

func defaultJSON(s string) string {
	if s == "" {
		return "{}"
	}
	return s
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}

// toSessionInfo model → dto
func toSessionInfo(s *model.Session) *dto.SessionInfo {
	return &dto.SessionInfo{
		ID:            s.ID,
		SessionID:     s.SessionID,
		SessionType:   s.SessionType,
		Participants:  s.Participants,
		LastMessage:   s.LastMessage,
		LastMessageAt: s.LastMessageAt,
		UnreadCount:   s.UnreadCount,
		Status:        s.Status,
		CreatedAt:     s.CreatedAt,
		UpdatedAt:     s.UpdatedAt,
	}
}

// toMessageInfo model → dto
func toMessageInfo(m *model.Message) *dto.MessageInfo {
	return &dto.MessageInfo{
		ID:         m.ID,
		SessionID:  m.SessionID,
		SenderID:   m.SenderID,
		MsgType:    m.MsgType,
		Content:    m.Content,
		Extra:      m.Extra,
		ReadStatus: m.ReadStatus,
		IsRecalled: m.IsRecalled,
		CreatedAt:  m.CreatedAt,
	}
}

// toNotificationInfo model → dto
func toNotificationInfo(n *model.SystemNotification) *dto.NotificationInfo {
	return &dto.NotificationInfo{
		ID:         n.ID,
		UserID:     n.UserID,
		NotifyType: n.NotifyType,
		Title:      n.Title,
		Content:    n.Content,
		JumpURL:    n.JumpURL,
		Extra:      n.Extra,
		IsRead:     n.IsRead,
		ReadAt:     n.ReadAt,
		CreatedAt:  n.CreatedAt,
	}
}

// toPrivacyNumberInfo model → dto
func toPrivacyNumberInfo(p *model.PrivacyNumber) *dto.PrivacyNumberInfo {
	return &dto.PrivacyNumberInfo{
		ID:          p.ID,
		PrivacyNo:   p.PrivacyNo,
		RealNoA:     p.RealNoA,
		RealNoB:     p.RealNoB,
		UserIDA:     p.UserIDA,
		UserIDB:     p.UserIDB,
		BizModule:   p.BizModule,
		BizID:       p.BizID,
		CallRecords: p.CallRecords,
		BoundAt:     p.BoundAt,
		UnboundAt:   p.UnboundAt,
		Status:      p.Status,
	}
}
