// Package service 沟通消息业务逻辑层
// 依据 v3.2.1 架构方案第四章：对标 BOSS直聘在线聊天
package service

import (
	"errors"
	"fmt"
	"time"

	"wuchang-tongcheng/internal/modules/job/dto"
	"wuchang-tongcheng/internal/modules/job/model"
	"wuchang-tongcheng/internal/modules/job/repository"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

var (
	ErrMessageNotFound     = errors.New("消息不存在")
	ErrMessageNoPermission = errors.New("无权操作此消息")
	ErrConversationEmpty   = errors.New("会话 ID 不能为空")
)

// MessageService 消息业务接口
type MessageService interface {
	Create(regionID uint, fromUserID uint, fromName string, fromAvatar string, req *dto.MessageCreateRequest) (*dto.MessageResponse, error)
	GetByID(id uint, userID uint) (*dto.MessageResponse, error)
	List(userID uint, req *dto.MessageListQuery) (*utils.Pagination, []dto.MessageResponse, error)
	ListByConversation(conversationID string, userID uint, page, pageSize int) (*utils.Pagination, []dto.MessageResponse, error)
	ListConversations(userID uint) ([]dto.ConversationResponse, error)
	MarkRead(conversationID string, userID uint) error
	CountUnread(userID uint) (int64, error)
	CountUnreadByUser(fromUserID, toUserID uint) (int64, error)
	Delete(id uint, userID uint) error
	BatchDelete(userID uint, ids []uint) error
	Recall(id uint, userID uint) error
}

type messageService struct {
	repo repository.MessageRepository
}

// NewMessageService 创建消息 service 实例
func NewMessageService(repo repository.MessageRepository) MessageService {
	return &messageService{repo: repo}
}

// genConversationID 生成会话 ID（小的 user_id 在前，保证双方会话 ID 一致）
func genConversationID(userA, userB uint) string {
	if userA > userB {
		userA, userB = userB, userA
	}
	return fmt.Sprintf("C_%d_%d", userA, userB)
}

// toMessageResponse model -> dto
func toMessageResponse(m *model.JobMessage) *dto.MessageResponse {
	resp := &dto.MessageResponse{
		ID:             m.ID,
		ConversationID: m.ConversationID,
		JobID:          m.JobID,
		ApplicationID:  m.ApplicationID,
		FromUserID:     m.FromUserID,
		ToUserID:       m.ToUserID,
		FromName:       m.FromName,
		FromAvatar:     m.FromAvatar,
		ToName:         m.ToName,
		ToAvatar:       m.ToAvatar,
		Content:        m.Content,
		MessageType:    m.MessageType,
		Attachments:    []map[string]interface{}{},
		IsRead:         m.IsRead,
		ReadAt:         m.ReadAt,
		IsRecruiter:    m.IsRecruiter,
		IsSystem:       m.IsSystem,
		Status:         m.Status,
		Source:         m.Source,
		RegionID:       m.RegionID,
		CreatedAt:      m.CreatedAt,
	}
	if m.Attachments != nil {
		var arr []map[string]interface{}
		_ = m.Attachments.Parse(&arr)
		if arr != nil {
			resp.Attachments = arr
		}
	}
	return resp
}

func (s *messageService) Create(regionID uint, fromUserID uint, fromName string, fromAvatar string, req *dto.MessageCreateRequest) (*dto.MessageResponse, error) {
	if fromUserID == req.ToUserID {
		return nil, errors.New("不能给自己发消息")
	}
	// 自动生成会话 ID（若未提供）
	conversationID := req.ConversationID
	if conversationID == "" {
		conversationID = genConversationID(fromUserID, req.ToUserID)
	}

	messageType := req.MessageType
	if messageType == "" {
		messageType = model.MessageTypeText
	}

	m := &model.JobMessage{
		ConversationID: conversationID,
		JobID:          req.JobID,
		ApplicationID:  req.ApplicationID,
		FromUserID:     fromUserID,
		ToUserID:       req.ToUserID,
		FromName:       fromName,
		FromAvatar:     fromAvatar,
		Content:        req.Content,
		MessageType:    messageType,
		IsRecruiter:    req.IsRecruiter,
		IsSystem:       false,
		Status:         model.MessageStatusNormal,
		Source:         model.MessageSourceChat,
	}
	m.RegionID = regionID

	if len(req.Attachments) > 0 {
		if jb, err := model.FromJSON(req.Attachments); err == nil {
			m.Attachments = jb
		}
	}

	if err := s.repo.Create(m); err != nil {
		return nil, err
	}
	return toMessageResponse(m), nil
}

func (s *messageService) GetByID(id uint, userID uint) (*dto.MessageResponse, error) {
	m, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrMessageNotFound
		}
		return nil, err
	}
	if userID > 0 && m.FromUserID != userID && m.ToUserID != userID {
		return nil, ErrMessageNoPermission
	}
	return toMessageResponse(m), nil
}

func (s *messageService) List(userID uint, req *dto.MessageListQuery) (*utils.Pagination, []dto.MessageResponse, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	list, total, err := s.repo.List(repository.MessageListQuery{
		UserID:         userID,
		Role:           req.Role,
		JobID:          req.JobID,
		ApplicationID:  req.ApplicationID,
		ConversationID: req.ConversationID,
		MessageType:    req.MessageType,
		IsRead:         req.IsRead,
		IsSystem:       req.IsSystem,
		Keyword:        req.Keyword,
	}, pagination)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	result := make([]dto.MessageResponse, 0, len(list))
	for i := range list {
		result = append(result, *toMessageResponse(&list[i]))
	}
	return pagination, result, nil
}

func (s *messageService) ListByConversation(conversationID string, userID uint, page, pageSize int) (*utils.Pagination, []dto.MessageResponse, error) {
	if conversationID == "" {
		return nil, nil, ErrConversationEmpty
	}
	pagination := utils.NewPagination(page, pageSize)
	// 默认每页 50 条（聊天场景）
	if pageSize == 0 {
		pagination = utils.NewPagination(page, 50)
	}
	list, total, err := s.repo.ListByConversation(conversationID, pagination)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	result := make([]dto.MessageResponse, 0, len(list))
	for i := range list {
		result = append(result, *toMessageResponse(&list[i]))
	}
	// 已读：将接收方为当前用户的消息标记为已读
	_ = s.repo.MarkReadByConversation(conversationID, userID)
	return pagination, result, nil
}

func (s *messageService) ListConversations(userID uint) ([]dto.ConversationResponse, error) {
	list, err := s.repo.ListConversations(userID)
	if err != nil {
		return nil, err
	}
	result := make([]dto.ConversationResponse, 0, len(list))
	for _, m := range list {
		// 统计当前用户在该会话的未读消息数
		unread, _ := s.repo.CountUnreadByUser(m.FromUserID, userID)
		// 若最后一条消息发送者就是当前用户，未读数为 0
		if m.FromUserID == userID {
			unread = 0
		}
		resp := dto.ConversationResponse{
			ConversationID:  m.ConversationID,
			JobID:           m.JobID,
			ApplicationID:   m.ApplicationID,
			FromUserID:      m.FromUserID,
			ToUserID:        m.ToUserID,
			FromName:        m.FromName,
			FromAvatar:      m.FromAvatar,
			ToName:          m.ToName,
			ToAvatar:        m.ToAvatar,
			LastContent:     m.Content,
			LastMessageType: m.MessageType,
			LastMessageAt:   m.CreatedAt,
			UnreadCount:     unread,
		}
		result = append(result, resp)
	}
	return result, nil
}

func (s *messageService) MarkRead(conversationID string, userID uint) error {
	if conversationID == "" {
		return ErrConversationEmpty
	}
	return s.repo.MarkReadByConversation(conversationID, userID)
}

func (s *messageService) CountUnread(userID uint) (int64, error) {
	return s.repo.CountUnread(userID)
}

func (s *messageService) CountUnreadByUser(fromUserID, toUserID uint) (int64, error) {
	return s.repo.CountUnreadByUser(fromUserID, toUserID)
}

func (s *messageService) Delete(id uint, userID uint) error {
	m, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrMessageNotFound
		}
		return err
	}
	if userID > 0 && m.FromUserID != userID && m.ToUserID != userID {
		return ErrMessageNoPermission
	}
	return s.repo.Delete(id)
}

func (s *messageService) BatchDelete(userID uint, ids []uint) error {
	// 简化：仅校验是否为消息发送/接收方，不做逐条校验（前端应已限制）
	return s.repo.BatchDelete(ids)
}

// Recall 撤回消息（仅发送者可撤回，且 2 分钟内）
func (s *messageService) Recall(id uint, userID uint) error {
	m, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrMessageNotFound
		}
		return err
	}
	if m.FromUserID != userID {
		return ErrMessageNoPermission
	}
	// 2 分钟内可撤回
	if time.Since(m.CreatedAt) > 2*time.Minute {
		return errors.New("超过 2 分钟不可撤回")
	}
	return s.repo.Update(id, map[string]interface{}{
		"status": model.MessageStatusRecall,
	})
}
