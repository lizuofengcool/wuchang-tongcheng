// Package service love 相亲交友业务逻辑层 - 匹配
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
	ErrLoveMatchNotFound = errors.New("匹配记录不存在")
	ErrLoveMatchNoPermission = errors.New("无权操作此匹配")
)

// LoveMatchService 匹配业务接口
type LoveMatchService interface {
	GetByID(id uint, userID uint) (*dto.LoveMatchInfo, error)
	List(req *dto.LoveMatchListRequest, userID uint) (*utils.Pagination, []dto.LoveMatchInfo, error)
	ListByUser(userID uint, req *dto.LoveMatchListRequest) (*utils.Pagination, []dto.LoveMatchInfo, error)
	Dissolve(id uint, userID uint, req *dto.DissolveMatchRequest) error
	CountByUser(userID uint) (int64, error)
	CountTodayByUser(userID uint) (int64, error)
}

type loveMatchService struct {
	repo repository.LoveMatchRepository
}

// NewLoveMatchService 创建匹配 service
func NewLoveMatchService(repo repository.LoveMatchRepository) LoveMatchService {
	return &loveMatchService{repo: repo}
}

func matchStatusText(s int) string {
	switch s {
	case model.MatchStatusActive:
		return "活跃"
	case model.MatchStatusUnmuted:
		return "解除匹配"
	case model.MatchStatusDissolved:
		return "已解除"
	case model.MatchStatusBlocked:
		return "已拉黑"
	}
	return ""
}

func toLoveMatchInfo(m *model.LoveMatch) dto.LoveMatchInfo {
	return dto.LoveMatchInfo{
		ID:                 m.ID,
		MatchNo:            m.MatchNo,
		UserIDA:            m.UserIDA,
		UserIDB:            m.UserIDB,
		LoveIDA:            m.LoveIDA,
		LoveIDB:            m.LoveIDB,
		MatchScore:         m.MatchScore,
		MatchType:          m.MatchType,
		InterestMatch:      m.InterestMatch,
		PersonalityMatch:   m.PersonalityMatch,
		ValueMatch:         m.ValueMatch,
		LocationMatch:      m.LocationMatch,
		AgeMatch:           m.AgeMatch,
		MatchedAt:          m.MatchedAt,
		ChatSessionID:      m.ChatSessionID,
		Status:             m.Status,
		StatusText:         matchStatusText(m.Status),
		LastMessageAt:      m.LastMessageAt,
		LastMessageContent: m.LastMessageContent,
		LastMessageType:    m.LastMessageType,
		DissolvedAt:        m.DissolvedAt,
		DissolveReason:     m.DissolveReason,
		CreatedAt:          m.CreatedAt,
		UpdatedAt:          m.UpdatedAt,
	}
}

func (s *loveMatchService) GetByID(id uint, userID uint) (*dto.LoveMatchInfo, error) {
	m, err := s.repo.FindByID(id)
	if err != nil {
		return nil, ErrLoveMatchNotFound
	}
	if m.UserIDA != userID && m.UserIDB != userID {
		return nil, ErrLoveMatchNoPermission
	}
	info := toLoveMatchInfo(m)
	// 对方信息
	if m.UserIDA == userID {
		info.UnreadCount = m.UnreadCountA
	} else {
		info.UnreadCount = m.UnreadCountB
	}
	return &info, nil
}

func (s *loveMatchService) List(req *dto.LoveMatchListRequest, userID uint) (*utils.Pagination, []dto.LoveMatchInfo, error) {
	return s.ListByUser(userID, req)
}

func (s *loveMatchService) ListByUser(userID uint, req *dto.LoveMatchListRequest) (*utils.Pagination, []dto.LoveMatchInfo, error) {
	opts := repository.LoveMatchListOptions{
		UserID:    userID,
		MatchType: req.MatchType,
	}
	if req.Status != nil {
		opts.Status = req.Status
	}
	list, total, err := s.repo.List(&req.Pagination, opts)
	if err != nil {
		return nil, nil, err
	}
	infos := make([]dto.LoveMatchInfo, 0, len(list))
	for i := range list {
		infos = append(infos, toLoveMatchInfo(&list[i]))
	}
	req.Pagination.Total = total
	return &req.Pagination, infos, nil
}

func (s *loveMatchService) Dissolve(id uint, userID uint, req *dto.DissolveMatchRequest) error {
	m, err := s.repo.FindByID(id)
	if err != nil {
		return ErrLoveMatchNotFound
	}
	if m.UserIDA != userID && m.UserIDB != userID {
		return ErrLoveMatchNoPermission
	}
	if m.Status == model.MatchStatusDissolved {
		return errors.New("匹配已解除")
	}
	return s.repo.Dissolve(id, userID, req.Reason)
}

func (s *loveMatchService) CountByUser(userID uint) (int64, error) {
	return s.repo.CountByUser(userID)
}

func (s *loveMatchService) CountTodayByUser(userID uint) (int64, error) {
	return s.repo.CountTodayByUser(userID)
}

// generateMatchNo 生成匹配单号
func generateMatchNo(prefix string, id uint) string {
	return fmt.Sprintf("%s%s%08d", prefix, time.Now().Format("20060102150405"), id%100000000)
}
