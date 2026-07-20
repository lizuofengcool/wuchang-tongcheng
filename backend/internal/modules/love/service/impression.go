// Package service love 相亲交友业务逻辑层 - 印象标签
package service

import (
	"errors"

	"wuchang-tongcheng/internal/modules/love/dto"
	"wuchang-tongcheng/internal/modules/love/model"
	"wuchang-tongcheng/internal/modules/love/repository"
	"wuchang-tongcheng/internal/pkg/utils"
)

var (
	ErrLoveImpressionNotFound     = errors.New("印象记录不存在")
	ErrLoveImpressionNoPermission = errors.New("无权操作此印象")
)

// LoveImpressionService 印象标签业务接口
type LoveImpressionService interface {
	Create(loveID, userID uint, req *dto.CreateLoveImpressionRequest) (*dto.LoveImpressionInfo, error)
	Delete(id uint, userID uint) error
	GetByID(id uint) (*dto.LoveImpressionInfo, error)
	List(req *dto.LoveImpressionListRequest) (*utils.Pagination, []dto.LoveImpressionInfo, error)
	ListByLoveID(loveID uint, req *dto.LoveImpressionListRequest) (*utils.Pagination, []dto.LoveImpressionInfo, error)
	ListByFromUser(fromUserID uint, req *dto.LoveImpressionListRequest) (*utils.Pagination, []dto.LoveImpressionInfo, error)
	Stats(loveID uint) (*dto.LoveImpressionStatsResponse, error)
}

type loveImpressionService struct {
	repo repository.LoveImpressionRepository
}

// NewLoveImpressionService 创建印象标签 service
func NewLoveImpressionService(repo repository.LoveImpressionRepository) LoveImpressionService {
	return &loveImpressionService{repo: repo}
}

func toLoveImpressionInfo(i *model.LoveImpression) dto.LoveImpressionInfo {
	return dto.LoveImpressionInfo{
		ID:                i.ID,
		LoveID:            i.LoveID,
		UserID:            i.UserID,
		FromUserID:        i.FromUserID,
		FromUserNickname:  i.FromUserNickname,
		FromUserAvatar:    i.FromUserAvatar,
		Tag:               i.Tag,
		Content:           i.Content,
		Anonymous:        i.Anonymous,
		IsAnonymous:       i.IsAnonymous,
		MatchID:           i.MatchID,
		Status:            i.Status,
		CreatedAt:         i.CreatedAt,
	}
}

func (s *loveImpressionService) Create(loveID, userID uint, req *dto.CreateLoveImpressionRequest) (*dto.LoveImpressionInfo, error) {
	i := &model.LoveImpression{
		LoveID:         req.TargetLoveID,
		UserID:         req.TargetUserID,
		FromUserID:     userID,
		Tag:            req.Tag,
		Content:        req.Content,
		Anonymous:      req.Anonymous,
		IsAnonymous:    req.Anonymous,
		MatchID:        req.MatchID,
		Status:         1,
	}
	_ = loveID
	if err := s.repo.Create(i); err != nil {
		return nil, err
	}
	info := toLoveImpressionInfo(i)
	return &info, nil
}

func (s *loveImpressionService) Delete(id uint, userID uint) error {
	i, err := s.repo.FindByID(id)
	if err != nil {
		return ErrLoveImpressionNotFound
	}
	if i.FromUserID != userID {
		return ErrLoveImpressionNoPermission
	}
	return s.repo.Delete(id)
}

func (s *loveImpressionService) GetByID(id uint) (*dto.LoveImpressionInfo, error) {
	i, err := s.repo.FindByID(id)
	if err != nil {
		return nil, ErrLoveImpressionNotFound
	}
	info := toLoveImpressionInfo(i)
	return &info, nil
}

func (s *loveImpressionService) List(req *dto.LoveImpressionListRequest) (*utils.Pagination, []dto.LoveImpressionInfo, error) {
	opts := repository.LoveImpressionListOptions{
		LoveID:     req.LoveID,
		UserID:     req.UserID,
		FromUserID: req.FromUserID,
		Tag:        req.Tag,
	}
	list, total, err := s.repo.List(&req.Pagination, opts)
	if err != nil {
		return nil, nil, err
	}
	infos := make([]dto.LoveImpressionInfo, 0, len(list))
	for i := range list {
		infos = append(infos, toLoveImpressionInfo(&list[i]))
	}
	req.Pagination.Total = total
	return &req.Pagination, infos, nil
}

func (s *loveImpressionService) ListByLoveID(loveID uint, req *dto.LoveImpressionListRequest) (*utils.Pagination, []dto.LoveImpressionInfo, error) {
	req.LoveID = loveID
	return s.List(req)
}

func (s *loveImpressionService) ListByFromUser(fromUserID uint, req *dto.LoveImpressionListRequest) (*utils.Pagination, []dto.LoveImpressionInfo, error) {
	req.FromUserID = fromUserID
	return s.List(req)
}

func (s *loveImpressionService) Stats(loveID uint) (*dto.LoveImpressionStatsResponse, error) {
	total, err := s.repo.CountByLoveID(loveID)
	if err != nil {
		return nil, err
	}
	tags, err := s.repo.ListTopTagsByLoveID(loveID, 20)
	if err != nil {
		return nil, err
	}
	tagStats := make([]dto.LoveImpressionTagStat, 0, len(tags))
	for i := range tags {
		tagStats = append(tagStats, dto.LoveImpressionTagStat{
			Tag:   tags[i].Tag,
			Count: tags[i].Count,
		})
	}
	return &dto.LoveImpressionStatsResponse{
		TotalImpressions: total,
		TopTags:          tagStats,
	}, nil
}
