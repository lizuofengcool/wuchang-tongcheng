// Package service love 相亲交友业务逻辑层 - 喜欢/不喜欢/心动信号
package service

import (
	"errors"
	"time"

	"wuchang-tongcheng/internal/modules/love/dto"
	"wuchang-tongcheng/internal/modules/love/model"
	"wuchang-tongcheng/internal/modules/love/repository"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

var (
	ErrLoveLikeNotFound      = errors.New("喜欢记录不存在")
	ErrLoveLikeNoPermission  = errors.New("无权操作此喜欢记录")
	ErrLoveLikeAlreadyExists = errors.New("已经操作过该用户")
	ErrLoveLikeInvalidAction = errors.New("无效的操作类型")
)

// LoveLikeService 喜欢业务接口
type LoveLikeService interface {
	// Act 执行喜欢/不喜欢/跳过/心动信号动作（双向喜欢则匹配）
	Act(loveID, userID uint, req *dto.CreateLoveLikeRequest) (*dto.LoveLikeInfo, error)
	Undo(id uint, userID uint, req *dto.UndoLoveLikeRequest) error
	GetByID(id uint) (*dto.LoveLikeInfo, error)
	List(req *dto.LoveLikeListRequest, userID uint) (*utils.Pagination, []dto.LoveLikeInfo, error)
	ListByUser(userID uint, req *dto.LoveLikeListRequest) (*utils.Pagination, []dto.LoveLikeInfo, error)
	ListMatched(userID uint, req *dto.LoveLikeListRequest) (*utils.Pagination, []dto.LoveLikeInfo, error)
	HasLiked(userID, targetUserID uint) (bool, error)
	TodayStats(userID uint, memberLevel int) (*dto.LoveLikeTodayStatsResponse, error)
}

type loveLikeService struct {
	repo repository.LoveLikeRepository
}

// NewLoveLikeService 创建喜欢 service
func NewLoveLikeService(repo repository.LoveLikeRepository) LoveLikeService {
	return &loveLikeService{repo: repo}
}

func toLoveLikeInfo(l *model.LoveLike) dto.LoveLikeInfo {
	return dto.LoveLikeInfo{
		ID:             l.ID,
		UserID:         l.UserID,
		LoveID:         l.LoveID,
		TargetUserID:   l.TargetUserID,
		TargetLoveID:   l.TargetLoveID,
		TargetNickname: l.TargetNickname,
		TargetAvatar:   l.TargetAvatar,
		TargetGender:   l.TargetGender,
		Action:         l.Action,
		SuperLike:      l.SuperLike,
		MatchScore:     l.MatchScore,
		Source:         l.Source,
		IsMatched:      l.IsMatched,
		MatchID:        l.MatchID,
		MatchedAt:      l.MatchedAt,
		Status:         l.Status,
		CreatedAt:      l.CreatedAt,
	}
}

func (s *loveLikeService) Act(loveID, userID uint, req *dto.CreateLoveLikeRequest) (*dto.LoveLikeInfo, error) {
	// 已存在记录则更新动作
	existing, err := s.repo.FindByUserTarget(userID, req.TargetUserID)
	if err == nil && existing != nil {
		// 更新动作
		fields := map[string]interface{}{
			"action":      req.Action,
			"super_like":  req.Action == model.LikeActionSuper,
			"source":      req.Source,
			"undone_at":   nil,
			"undo_reason": "",
			"status":      1,
			"updated_at":  time.Now(),
		}
		if err := s.repo.UpdateFields(existing.ID, fields); err != nil {
			return nil, err
		}
		existing.Action = req.Action
		existing.SuperLike = req.Action == model.LikeActionSuper
		existing.Source = req.Source
		info := toLoveLikeInfo(existing)
		return &info, nil
	} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	like := &model.LoveLike{
		UserID:         userID,
		LoveID:         loveID,
		TargetUserID:   req.TargetUserID,
		TargetLoveID:   req.TargetLoveID,
		Action:         req.Action,
		SuperLike:      req.Action == model.LikeActionSuper,
		Source:         req.Source,
		Status:         1,
	}
	if req.Source == "" {
		like.Source = "recommend"
	}
	if err := s.repo.Create(like); err != nil {
		return nil, err
	}

	// 双向喜欢则匹配
	if req.Action == model.LikeActionLike || req.Action == model.LikeActionSuper {
		_ = s.tryMatch(userID, req.TargetUserID, like)
	}

	info := toLoveLikeInfo(like)
	return &info, nil
}

// tryMatch 尝试匹配，若对方也喜欢我则标记 is_matched
func (s *loveLikeService) tryMatch(userA, userB uint, likeA *model.LoveLike) error {
	likeB, err := s.repo.FindByUserTarget(userB, userA)
	if err != nil {
		return nil
	}
	if likeB.Action != model.LikeActionLike && likeB.Action != model.LikeActionSuper {
		return nil
	}
	now := time.Now()
	_ = s.repo.MarkMatched(likeA.ID, 0)
	_ = s.repo.MarkMatched(likeB.ID, 0)
	_ = s.repo.UpdateFields(likeA.ID, map[string]interface{}{"is_matched": true, "matched_at": now})
	_ = s.repo.UpdateFields(likeB.ID, map[string]interface{}{"is_matched": true, "matched_at": now})
	likeA.IsMatched = true
	likeA.MatchedAt = &now
	return nil
}

func (s *loveLikeService) Undo(id uint, userID uint, req *dto.UndoLoveLikeRequest) error {
	l, err := s.repo.FindByID(id)
	if err != nil {
		return ErrLoveLikeNotFound
	}
	if l.UserID != userID {
		return ErrLoveLikeNoPermission
	}
	return s.repo.Undo(id, req.Reason)
}

func (s *loveLikeService) GetByID(id uint) (*dto.LoveLikeInfo, error) {
	l, err := s.repo.FindByID(id)
	if err != nil {
		return nil, ErrLoveLikeNotFound
	}
	info := toLoveLikeInfo(l)
	return &info, nil
}

func (s *loveLikeService) List(req *dto.LoveLikeListRequest, userID uint) (*utils.Pagination, []dto.LoveLikeInfo, error) {
	return s.ListByUser(userID, req)
}

func (s *loveLikeService) ListByUser(userID uint, req *dto.LoveLikeListRequest) (*utils.Pagination, []dto.LoveLikeInfo, error) {
	opts := repository.LoveLikeListOptions{
		UserID: userID,
		Action: req.Action,
	}
	if req.SuperLike != nil {
		opts.SuperLike = req.SuperLike
	}
	if req.IsMatched != nil {
		opts.IsMatched = req.IsMatched
	}
	list, total, err := s.repo.List(&req.Pagination, opts)
	if err != nil {
		return nil, nil, err
	}
	infos := make([]dto.LoveLikeInfo, 0, len(list))
	for i := range list {
		infos = append(infos, toLoveLikeInfo(&list[i]))
	}
	req.Pagination.Total = total
	return &req.Pagination, infos, nil
}

func (s *loveLikeService) ListMatched(userID uint, req *dto.LoveLikeListRequest) (*utils.Pagination, []dto.LoveLikeInfo, error) {
	matched := true
	opts := repository.LoveLikeListOptions{
		UserID:    userID,
		IsMatched: &matched,
	}
	list, total, err := s.repo.List(&req.Pagination, opts)
	if err != nil {
		return nil, nil, err
	}
	infos := make([]dto.LoveLikeInfo, 0, len(list))
	for i := range list {
		infos = append(infos, toLoveLikeInfo(&list[i]))
	}
	req.Pagination.Total = total
	return &req.Pagination, infos, nil
}

func (s *loveLikeService) HasLiked(userID, targetUserID uint) (bool, error) {
	return s.repo.HasLiked(userID, targetUserID)
}

// likeLimitByMemberLevel 依据会员等级返回每日限额（likes/superLikes）
func likeLimitByMemberLevel(memberLevel int) (limitLikes, limitSuperLikes int) {
	switch memberLevel {
	case model.MemberLevelNone:
		return 20, 0
	case model.MemberLevelBasic:
		return 50, 1
	case model.MemberLevelAdvanced:
		return 100, 3
	case model.MemberLevelVIP:
		return 200, 5
	case model.MemberLevelPremium:
		return 500, 10
	}
	return 20, 0
}

func (s *loveLikeService) TodayStats(userID uint, memberLevel int) (*dto.LoveLikeTodayStatsResponse, error) {
	limitLikes, limitSuperLikes := likeLimitByMemberLevel(memberLevel)
	likeCount, _ := s.repo.CountTodayByAction(userID, model.LikeActionLike)
	dislikeCount, _ := s.repo.CountTodayByAction(userID, model.LikeActionDislike)
	superCount, _ := s.repo.CountTodayByAction(userID, model.LikeActionSuper)
	skipCount, _ := s.repo.CountTodayByAction(userID, model.LikeActionSkip)
	matchedCount, _ := s.repo.CountMatchedByUser(userID)

	remainingLikes := limitLikes - likeCount
	if remainingLikes < 0 {
		remainingLikes = 0
	}
	remainingSuper := limitSuperLikes - superCount
	if remainingSuper < 0 {
		remainingSuper = 0
	}
	return &dto.LoveLikeTodayStatsResponse{
		TotalLikes:           likeCount,
		TotalDislikes:        dislikeCount,
		TotalSuperLikes:      superCount,
		TotalSkips:           skipCount,
		TotalMatches:         int(matchedCount),
		LimitLikes:           limitLikes,
		LimitSuperLikes:      limitSuperLikes,
		RemainingLikes:       remainingLikes,
		RemainingSuperLikes:  remainingSuper,
	}, nil
}
