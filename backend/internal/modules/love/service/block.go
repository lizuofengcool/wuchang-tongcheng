// Package service love 相亲交友业务逻辑层 - 拉黑
package service

import (
	"errors"

	"wuchang-tongcheng/internal/modules/love/dto"
	"wuchang-tongcheng/internal/modules/love/model"
	"wuchang-tongcheng/internal/modules/love/repository"
	"wuchang-tongcheng/internal/pkg/utils"
)

var (
	ErrLoveBlockNotFound     = errors.New("拉黑记录不存在")
	ErrLoveBlockNoPermission = errors.New("无权操作此拉黑记录")
	ErrLoveBlocked           = errors.New("已被对方拉黑")
)

// LoveBlockService 拉黑业务接口
type LoveBlockService interface {
	Block(loveID, userID uint, req *dto.CreateLoveBlockRequest) (*dto.LoveBlockInfo, error)
	Unblock(id uint, userID uint) error
	GetByID(id uint, userID uint) (*dto.LoveBlockInfo, error)
	List(req *dto.LoveBlockListRequest, userID uint) (*utils.Pagination, []dto.LoveBlockInfo, error)
	ListByUser(userID uint, req *dto.LoveBlockListRequest) (*utils.Pagination, []dto.LoveBlockInfo, error)
	HasBlocked(userID, blockedUserID uint) (bool, error)
	HasBlockedEither(userA, userB uint) (bool, error)
	CountByUser(userID uint) (int64, error)
}

type loveBlockService struct {
	repo repository.LoveBlockRepository
}

// NewLoveBlockService 创建拉黑 service
func NewLoveBlockService(repo repository.LoveBlockRepository) LoveBlockService {
	return &loveBlockService{repo: repo}
}

func toLoveBlockInfo(b *model.LoveBlock) dto.LoveBlockInfo {
	return dto.LoveBlockInfo{
		ID:              b.ID,
		UserID:          b.UserID,
		LoveID:          b.LoveID,
		BlockedUserID:   b.BlockedUserID,
		BlockedLoveID:   b.BlockedLoveID,
		BlockedNickname: b.BlockedNickname,
		BlockedAvatar:   b.BlockedAvatar,
		Reason:          b.Reason,
		ReportID:        b.ReportID,
		Status:          b.Status,
		CreatedAt:       b.CreatedAt,
		UpdatedAt:       b.UpdatedAt,
	}
}

func (s *loveBlockService) Block(loveID, userID uint, req *dto.CreateLoveBlockRequest) (*dto.LoveBlockInfo, error) {
	b := &model.LoveBlock{
		UserID:         userID,
		LoveID:         loveID,
		BlockedUserID:  req.BlockedUserID,
		BlockedLoveID:  req.BlockedLoveID,
		Reason:          req.Reason,
		ReportID:        req.ReportID,
		Status:          1,
	}
	if err := s.repo.Upsert(b); err != nil {
		return nil, err
	}
	info := toLoveBlockInfo(b)
	return &info, nil
}

func (s *loveBlockService) Unblock(id uint, userID uint) error {
	b, err := s.repo.FindByID(id)
	if err != nil {
		return ErrLoveBlockNotFound
	}
	if b.UserID != userID {
		return ErrLoveBlockNoPermission
	}
	return s.repo.UpdateFields(id, map[string]interface{}{"status": 0})
}

func (s *loveBlockService) GetByID(id uint, userID uint) (*dto.LoveBlockInfo, error) {
	b, err := s.repo.FindByID(id)
	if err != nil {
		return nil, ErrLoveBlockNotFound
	}
	if b.UserID != userID {
		return nil, ErrLoveBlockNoPermission
	}
	info := toLoveBlockInfo(b)
	return &info, nil
}

func (s *loveBlockService) List(req *dto.LoveBlockListRequest, userID uint) (*utils.Pagination, []dto.LoveBlockInfo, error) {
	return s.ListByUser(userID, req)
}

func (s *loveBlockService) ListByUser(userID uint, req *dto.LoveBlockListRequest) (*utils.Pagination, []dto.LoveBlockInfo, error) {
	opts := repository.LoveBlockListOptions{UserID: userID}
	list, total, err := s.repo.List(&req.Pagination, opts)
	if err != nil {
		return nil, nil, err
	}
	infos := make([]dto.LoveBlockInfo, 0, len(list))
	for i := range list {
		infos = append(infos, toLoveBlockInfo(&list[i]))
	}
	req.Pagination.Total = total
	return &req.Pagination, infos, nil
}

func (s *loveBlockService) HasBlocked(userID, blockedUserID uint) (bool, error) {
	return s.repo.HasBlocked(userID, blockedUserID)
}

func (s *loveBlockService) HasBlockedEither(userA, userB uint) (bool, error) {
	return s.repo.HasBlockedEither(userA, userB)
}

func (s *loveBlockService) CountByUser(userID uint) (int64, error) {
	return s.repo.CountByUser(userID)
}
