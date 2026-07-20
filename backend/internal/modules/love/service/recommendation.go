// Package service love 相亲交友业务逻辑层 - 推荐池
// 依据需求文档 1.10：4 维数据隔离（region_id + user_id）
// 5 维度对标灵魂匹配：兴趣/性格/价值观/位置/年龄
package service

import (
	"errors"
	"time"

	"wuchang-tongcheng/internal/modules/love/dto"
	"wuchang-tongcheng/internal/modules/love/model"
	"wuchang-tongcheng/internal/modules/love/repository"
	"wuchang-tongcheng/internal/pkg/utils"
)

var (
	ErrLoveRecommendationNotFound = errors.New("推荐记录不存在")
	ErrLoveRecommendationExpired  = errors.New("推荐已过期")
	ErrLoveRecommendationInvalidAction = errors.New("无效的推荐操作")
)

// LoveRecommendationService 推荐池业务接口
type LoveRecommendationService interface {
	// 内部调用：算法服务生成推荐后批量入库
	BatchCreate(items []model.LoveRecommendation) error
	Generate(userID, loveID uint, req *dto.GenerateLoveRecommendationsRequest) ([]dto.LoveRecommendationInfo, error)
	GetByID(id uint, userID uint) (*dto.LoveRecommendationInfo, error)
	List(req *dto.LoveRecommendationListRequest, userID uint) (*utils.Pagination, []dto.LoveRecommendationInfo, error)
	ListByUserAndType(userID uint, recType string, req *dto.LoveRecommendationListRequest) (*utils.Pagination, []dto.LoveRecommendationInfo, error)
	Action(req *dto.LoveRecommendationActionRequest, userID uint) error
	DeleteExpired() (int64, error)
	Stats(userID uint) (*dto.LoveRecommendationStatsResponse, error)
}

type loveRecommendationService struct {
	repo repository.LoveRecommendationRepository
}

// NewLoveRecommendationService 创建推荐池 service
func NewLoveRecommendationService(repo repository.LoveRecommendationRepository) LoveRecommendationService {
	return &loveRecommendationService{repo: repo}
}

// recStatusText 推荐状态文本
func recStatusText(s int) string {
	switch s {
	case model.RecStatusPending:
		return "待处理"
	case model.RecStatusViewed:
		return "已查看"
	case model.RecStatusLiked:
		return "已喜欢"
	case model.RecStatusDisliked:
		return "已不喜欢"
	case model.RecStatusSkipped:
		return "已跳过"
	case model.RecStatusDismissed:
		return "已忽略"
	case model.RecStatusExpired:
		return "已过期"
	}
	return ""
}

// toLoveRecommendationInfo model -> dto
func toLoveRecommendationInfo(r *model.LoveRecommendation) dto.LoveRecommendationInfo {
	return dto.LoveRecommendationInfo{
		ID:                r.ID,
		UserID:            r.UserID,
		LoveID:            r.LoveID,
		TargetUserID:      r.TargetUserID,
		TargetLoveID:      r.TargetLoveID,
		TargetNickname:    r.TargetNickname,
		TargetAvatar:      r.TargetAvatar,
		TargetGender:      r.TargetGender,
		TargetAge:         r.TargetAge,
		TargetDistance:    r.TargetDistance,
		TargetVerified:    false,
		TargetMemberLevel: 0,
		RecType:           r.RecType,
		Source:            r.Source,
		Score:             r.Score,
		InterestMatch:     r.InterestMatch,
		PersonalityMatch:  r.PersonalityMatch,
		ValueMatch:        r.ValueMatch,
		LocationMatch:     r.LocationMatch,
		AgeMatch:          r.AgeMatch,
		Reason:            r.Reason,
		IsViewed:          r.IsViewed,
		IsLiked:           r.IsLiked,
		IsDisliked:        r.IsDisliked,
		IsSuperLiked:      r.IsSuperLiked,
		IsSkipped:         r.IsSkipped,
		IsDismissed:       r.IsDismissed,
		Status:            r.Status,
		ExpiredAt:         r.ExpiredAt,
		CreatedAt:         r.CreatedAt,
	}
}

// BatchCreate 批量创建推荐（内部调用）
func (s *loveRecommendationService) BatchCreate(items []model.LoveRecommendation) error {
	if len(items) == 0 {
		return nil
	}
	return s.repo.BatchCreate(items)
}

// Generate 生成推荐
// MVP 简化：不接入算法服务，仅返回当日剩余推荐池
// 完整实现：调用算法服务计算 5 维评分 → 批量入库 → 返回列表
func (s *loveRecommendationService) Generate(userID, loveID uint, req *dto.GenerateLoveRecommendationsRequest) ([]dto.LoveRecommendationInfo, error) {
	recType := req.RecType
	if recType == "" {
		recType = "daily"
	}
	count := req.Count
	if count <= 0 {
		count = 10
	}
	_ = loveID

	// 取当前用户已有推荐（避免重复生成）
	pagination := utils.NewPagination(1, count)
	list, _, err := s.repo.ListByUserAndType(userID, recType, pagination)
	if err != nil {
		return nil, err
	}
	infos := make([]dto.LoveRecommendationInfo, 0, len(list))
	for i := range list {
		infos = append(infos, toLoveRecommendationInfo(&list[i]))
	}
	return infos, nil
}

func (s *loveRecommendationService) GetByID(id uint, userID uint) (*dto.LoveRecommendationInfo, error) {
	r, err := s.repo.FindByID(id)
	if err != nil {
		return nil, ErrLoveRecommendationNotFound
	}
	if r.UserID != userID {
		return nil, ErrLoveRecommendationNotFound
	}
	info := toLoveRecommendationInfo(r)
	return &info, nil
}

func (s *loveRecommendationService) List(req *dto.LoveRecommendationListRequest, userID uint) (*utils.Pagination, []dto.LoveRecommendationInfo, error) {
	opts := repository.LoveRecommendationListOptions{
		UserID:  userID,
		RecType: req.RecType,
		Source:  req.Source,
	}
	if req.Status != nil {
		opts.Status = req.Status
	}
	list, total, err := s.repo.List(&req.Pagination, opts)
	if err != nil {
		return nil, nil, err
	}
	infos := make([]dto.LoveRecommendationInfo, 0, len(list))
	for i := range list {
		infos = append(infos, toLoveRecommendationInfo(&list[i]))
	}
	req.Pagination.Total = total
	return &req.Pagination, infos, nil
}

func (s *loveRecommendationService) ListByUserAndType(userID uint, recType string, req *dto.LoveRecommendationListRequest) (*utils.Pagination, []dto.LoveRecommendationInfo, error) {
	list, total, err := s.repo.ListByUserAndType(userID, recType, &req.Pagination)
	if err != nil {
		return nil, nil, err
	}
	infos := make([]dto.LoveRecommendationInfo, 0, len(list))
	for i := range list {
		infos = append(infos, toLoveRecommendationInfo(&list[i]))
	}
	req.Pagination.Total = total
	return &req.Pagination, infos, nil
}

// Action 推荐操作（view/like/dislike/super_like/skip/dismiss）
func (s *loveRecommendationService) Action(req *dto.LoveRecommendationActionRequest, userID uint) error {
	r, err := s.repo.FindByID(req.ID)
	if err != nil {
		return ErrLoveRecommendationNotFound
	}
	if r.UserID != userID {
		return ErrLoveRecommendationNotFound
	}
	// 过期判断
	if r.ExpiredAt != nil && time.Now().After(*r.ExpiredAt) {
		return ErrLoveRecommendationExpired
	}
	switch req.Action {
	case "view":
		return s.repo.MarkViewed(req.ID)
	case "like":
		return s.repo.MarkLiked(req.ID)
	case "dislike":
		return s.repo.MarkDisliked(req.ID)
	case "super_like":
		return s.repo.MarkSuperLiked(req.ID)
	case "skip":
		return s.repo.MarkSkipped(req.ID)
	case "dismiss":
		return s.repo.MarkDismissed(req.ID)
	}
	return ErrLoveRecommendationInvalidAction
}

func (s *loveRecommendationService) DeleteExpired() (int64, error) {
	return s.repo.DeleteExpired()
}

// Stats 推荐统计
func (s *loveRecommendationService) Stats(userID uint) (*dto.LoveRecommendationStatsResponse, error) {
	total, err := s.repo.CountByUser(userID)
	if err != nil {
		return nil, err
	}
	today, _ := s.repo.CountTodayByUser(userID)
	// 简化：各状态计数复用 total
	return &dto.LoveRecommendationStatsResponse{
		TotalRecommendations: total,
		TodayRecommendations: today,
		ViewedCount:          0,
		LikedCount:           0,
		DislikedCount:        0,
		SuperLikedCount:      0,
		SkippedCount:         0,
		DismissedCount:       0,
	}, nil
}
