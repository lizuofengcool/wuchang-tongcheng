// Package service love 相亲交友业务逻辑层 - 访客
// 依据需求文档 1.10：4 维数据隔离（region_id + user_id）
package service

import (
	"errors"

	"wuchang-tongcheng/internal/modules/love/dto"
	"wuchang-tongcheng/internal/modules/love/model"
	"wuchang-tongcheng/internal/modules/love/repository"
	"wuchang-tongcheng/internal/pkg/utils"
)

var (
	ErrLoveVisitNotFound       = errors.New("访客记录不存在")
	ErrLoveVisitNoPermission   = errors.New("无权操作此访客记录")
	ErrLoveVisitHidden         = errors.New("访客已开启隐身访问")
)

// LoveVisitService 访客业务接口
type LoveVisitService interface {
	// C 端：访问他人主页时调用
	Visit(visitorUserID, visitorLoveID uint, visitorNickname, visitorAvatar string, visitorGender int, req *dto.CreateLoveVisitRequest) (*dto.LoveVisitInfo, error)
	GetByID(id uint, userID uint) (*dto.LoveVisitInfo, error)
	List(req *dto.LoveVisitListRequest, userID uint) (*utils.Pagination, []dto.LoveVisitInfo, error)
	ListByVisitor(userID uint, req *dto.LoveVisitListRequest) (*utils.Pagination, []dto.LoveVisitInfo, error)
	ListUnread(userID uint, req *dto.LoveVisitListRequest) (*utils.Pagination, []dto.LoveVisitInfo, error)
	MarkRead(id uint, userID uint) error
	MarkAllRead(userID uint) error
	Stats(userID uint) (*dto.LoveVisitStatsResponse, error)
}

type loveVisitService struct {
	repo repository.LoveVisitRepository
}

// NewLoveVisitService 创建访客 service
func NewLoveVisitService(repo repository.LoveVisitRepository) LoveVisitService {
	return &loveVisitService{repo: repo}
}

// toLoveVisitInfo model -> dto
func toLoveVisitInfo(v *model.LoveVisit) dto.LoveVisitInfo {
	return dto.LoveVisitInfo{
		ID:              v.ID,
		LoveID:          v.LoveID,
		UserID:          v.UserID,
		VisitorUserID:   v.VisitorUserID,
		VisitorLoveID:   v.VisitorLoveID,
		VisitorNickname: v.VisitorNickname,
		VisitorAvatar:   v.VisitorAvatar,
		VisitorGender:   v.VisitorGender,
		VisitType:       v.VisitType,
		Source:          v.Source,
		Duration:        v.Duration,
		IsHidden:        v.IsHidden,
		IsRead:          v.IsRead,
		Status:          v.Status,
		CreatedAt:       v.CreatedAt,
	}
}

// Visit 访问他人主页
// 入参 req.TargetUserID / req.TargetLoveID 为被访者
// visitor* 为访客信息
func (s *loveVisitService) Visit(visitorUserID, visitorLoveID uint, visitorNickname, visitorAvatar string, visitorGender int, req *dto.CreateLoveVisitRequest) (*dto.LoveVisitInfo, error) {
	// 不允许访问自己
	if visitorUserID == req.TargetUserID {
		return nil, ErrLoveVisitNoPermission
	}
	visitType := req.VisitType
	if visitType == "" {
		visitType = "profile"
	}
	source := req.Source
	if source == "" {
		source = "profile"
	}
	v := &model.LoveVisit{
		UserID:           req.TargetUserID,
		LoveID:           req.TargetLoveID,
		VisitorUserID:    visitorUserID,
		VisitorLoveID:    visitorLoveID,
		VisitorNickname:  visitorNickname,
		VisitorAvatar:    visitorAvatar,
		VisitorGender:    visitorGender,
		VisitType:        visitType,
		Source:           source,
		Duration:         req.Duration,
		IsRead:           false,
		Status:           1,
	}
	// Upsert：同一访客记录则更新时间，否则插入
	if err := s.repo.Upsert(v); err != nil {
		return nil, err
	}
	info := toLoveVisitInfo(v)
	return &info, nil
}

func (s *loveVisitService) GetByID(id uint, userID uint) (*dto.LoveVisitInfo, error) {
	v, err := s.repo.FindByID(id)
	if err != nil {
		return nil, ErrLoveVisitNotFound
	}
	// 仅被访者本人或访客本人可查
	if v.UserID != userID && v.VisitorUserID != userID {
		return nil, ErrLoveVisitNoPermission
	}
	info := toLoveVisitInfo(v)
	return &info, nil
}

// List 当前用户的访客列表（别人来看我）
func (s *loveVisitService) List(req *dto.LoveVisitListRequest, userID uint) (*utils.Pagination, []dto.LoveVisitInfo, error) {
	opts := repository.LoveVisitListOptions{
		UserID:     userID,
		VisitType:  req.VisitType,
		IsRead:     req.IsRead,
	}
	list, total, err := s.repo.List(&req.Pagination, opts)
	if err != nil {
		return nil, nil, err
	}
	infos := make([]dto.LoveVisitInfo, 0, len(list))
	for i := range list {
		infos = append(infos, toLoveVisitInfo(&list[i]))
	}
	req.Pagination.Total = total
	return &req.Pagination, infos, nil
}

// ListByVisitor 当前用户作为访客的访问记录（我看过谁）
func (s *loveVisitService) ListByVisitor(userID uint, req *dto.LoveVisitListRequest) (*utils.Pagination, []dto.LoveVisitInfo, error) {
	list, total, err := s.repo.ListByVisitor(userID, &req.Pagination)
	if err != nil {
		return nil, nil, err
	}
	infos := make([]dto.LoveVisitInfo, 0, len(list))
	for i := range list {
		infos = append(infos, toLoveVisitInfo(&list[i]))
	}
	req.Pagination.Total = total
	return &req.Pagination, infos, nil
}

func (s *loveVisitService) ListUnread(userID uint, req *dto.LoveVisitListRequest) (*utils.Pagination, []dto.LoveVisitInfo, error) {
	list, total, err := s.repo.ListUnread(userID, &req.Pagination)
	if err != nil {
		return nil, nil, err
	}
	infos := make([]dto.LoveVisitInfo, 0, len(list))
	for i := range list {
		infos = append(infos, toLoveVisitInfo(&list[i]))
	}
	req.Pagination.Total = total
	return &req.Pagination, infos, nil
}

func (s *loveVisitService) MarkRead(id uint, userID uint) error {
	v, err := s.repo.FindByID(id)
	if err != nil {
		return ErrLoveVisitNotFound
	}
	if v.UserID != userID {
		return ErrLoveVisitNoPermission
	}
	return s.repo.MarkRead(id)
}

func (s *loveVisitService) MarkAllRead(userID uint) error {
	return s.repo.MarkAllRead(userID)
}

// Stats 访客统计
func (s *loveVisitService) Stats(userID uint) (*dto.LoveVisitStatsResponse, error) {
	total, err := s.repo.CountByUser(userID)
	if err != nil {
		return nil, err
	}
	today, _ := s.repo.CountTodayByUser(userID)
	unread, _ := s.repo.CountUnreadByUser(userID)
	weekly, _ := s.repo.CountWeeklyByUser(userID)
	monthly, _ := s.repo.CountMonthlyByUser(userID)
	return &dto.LoveVisitStatsResponse{
		TotalVisitors:   total,
		TodayVisitors:   today,
		UnreadVisitors:  unread,
		WeeklyVisitors:  weekly,
		MonthlyVisitors: monthly,
	}, nil
}
