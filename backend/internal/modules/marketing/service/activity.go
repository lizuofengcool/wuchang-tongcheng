// Package service 营销活动中台业务逻辑层 - 营销活动（activity 子域）
package service

import (
	"errors"
	"time"

	"wuchang-tongcheng/internal/modules/marketing/dto"
	"wuchang-tongcheng/internal/modules/marketing/model"
	"wuchang-tongcheng/internal/modules/marketing/repository"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

// ActivityService 营销活动业务接口
type ActivityService interface {
	// 活动 CRUD
	Create(regionID uint, req *dto.CreateActivityRequest) (*dto.ActivityInfo, error)
	Update(id uint, req *dto.UpdateActivityRequest) error
	Delete(id uint) error
	GetByID(id uint) (*dto.ActivityInfo, error)
	List(regionID uint, req *dto.ActivityListRequest) (*utils.Pagination, []dto.ActivityInfo, error)
	ListOngoing(regionID uint, page, pageSize int) (*utils.Pagination, []dto.ActivityInfo, error)
	ListUpcoming(regionID uint, page, pageSize int) (*utils.Pagination, []dto.ActivityInfo, error)
	ListEnded(regionID uint, page, pageSize int) (*utils.Pagination, []dto.ActivityInfo, error)

	// 状态管理
	UpdateStatus(id uint, status int) error
	AutoUpdateStatus() (int64, error)

	// 统计
	Statistics(regionID uint) (*dto.ActivityStatistics, error)
}

type activityService struct {
	repo repository.ActivityRepository
}

// NewActivityService 创建营销活动 service 实例
func NewActivityService(repo repository.ActivityRepository) ActivityService {
	return &activityService{repo: repo}
}

// activityTypeText 活动类型文本
func activityTypeText(t string) string {
	switch t {
	case model.ActivityTypeGroupBuy:
		return "拼团"
	case model.ActivityTypeBargain:
		return "砍价"
	case model.ActivityTypeSeckill:
		return "秒杀"
	case model.ActivityTypeLottery:
		return "抽奖"
	}
	return ""
}

// activityStatusText 活动状态文本
func activityStatusText(s int) string {
	switch s {
	case model.ActivityStatusDisabled:
		return "禁用"
	case model.ActivityStatusPending:
		return "待开始"
	case model.ActivityStatusOngoing:
		return "进行中"
	case model.ActivityStatusEnded:
		return "已结束"
	case model.ActivityStatusCancelled:
		return "已取消"
	}
	return ""
}

// toActivityInfo model -> dto
func toActivityInfo(a *model.Activity) *dto.ActivityInfo {
	info := &dto.ActivityInfo{
		ID:          a.ID,
		RegionID:    a.RegionID,
		Title:       a.Title,
		Type:        a.Type,
		TypeText:    activityTypeText(a.Type),
		Description: a.Description,
		CoverImage:  a.CoverImage,
		StartAt:     a.StartAt,
		EndAt:       a.EndAt,
		Status:      a.Status,
		StatusText:  activityStatusText(a.Status),
		CreatedAt:   a.CreatedAt,
		UpdatedAt:   a.UpdatedAt,
	}
	if a.Config != nil && len(a.Config) > 0 {
		var cfg interface{}
		if err := a.Config.Parse(&cfg); err == nil {
			info.Config = cfg
		} else {
			info.Config = a.Config.String()
		}
	}
	return info
}

// Create 创建营销活动
func (s *activityService) Create(regionID uint, req *dto.CreateActivityRequest) (*dto.ActivityInfo, error) {
	status := req.Status
	if status == 0 {
		status = model.ActivityStatusPending
	}
	a := &model.Activity{
		Title:       req.Title,
		Type:        req.Type,
		Description: req.Description,
		CoverImage:  req.CoverImage,
		StartAt:     req.StartAt,
		EndAt:       req.EndAt,
		Status:      status,
	}
	a.RegionID = regionID
	if req.Config != nil {
		if b, err := model.FromJSON(req.Config); err == nil {
			a.Config = b
		}
	}
	if err := s.repo.Create(a); err != nil {
		return nil, err
	}
	return toActivityInfo(a), nil
}

// Update 更新营销活动
func (s *activityService) Update(id uint, req *dto.UpdateActivityRequest) error {
	if _, err := s.repo.FindByID(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrActivityNotFound
		}
		return err
	}
	fields := make(map[string]interface{})
	if req.Title != nil {
		fields["title"] = *req.Title
	}
	if req.Type != nil {
		fields["type"] = *req.Type
	}
	if req.Description != nil {
		fields["description"] = *req.Description
	}
	if req.CoverImage != nil {
		fields["cover_image"] = *req.CoverImage
	}
	if req.StartAt != nil {
		fields["start_at"] = req.StartAt
	}
	if req.EndAt != nil {
		fields["end_at"] = req.EndAt
	}
	if req.Status != nil {
		fields["status"] = *req.Status
	}
	if req.Config != nil {
		if b, err := model.FromJSON(req.Config); err == nil {
			fields["config"] = b
		}
	}
	if len(fields) == 0 {
		return nil
	}
	return s.repo.Update(id, fields)
}

// Delete 删除营销活动
func (s *activityService) Delete(id uint) error {
	if _, err := s.repo.FindByID(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrActivityNotFound
		}
		return err
	}
	return s.repo.Delete(id)
}

// GetByID 获取营销活动详情
func (s *activityService) GetByID(id uint) (*dto.ActivityInfo, error) {
	a, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrActivityNotFound
		}
		return nil, err
	}
	return toActivityInfo(a), nil
}

// List 营销活动列表
func (s *activityService) List(regionID uint, req *dto.ActivityListRequest) (*utils.Pagination, []dto.ActivityInfo, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	query := repository.ActivityListQuery{
		Type:    req.Type,
		Status:  req.Status,
		Keyword: req.Keyword,
	}
	list, total, err := s.repo.List(regionID, query, pagination)
	if err != nil {
		return nil, nil, err
	}
	infos := make([]dto.ActivityInfo, 0, len(list))
	for i := range list {
		infos = append(infos, *toActivityInfo(&list[i]))
	}
	pagination.Total = total
	return pagination, infos, nil
}

// ListOngoing 进行中的活动
func (s *activityService) ListOngoing(regionID uint, page, pageSize int) (*utils.Pagination, []dto.ActivityInfo, error) {
	pagination := utils.NewPagination(page, pageSize)
	list, total, err := s.repo.ListOngoing(regionID, pagination)
	if err != nil {
		return nil, nil, err
	}
	infos := make([]dto.ActivityInfo, 0, len(list))
	for i := range list {
		infos = append(infos, *toActivityInfo(&list[i]))
	}
	pagination.Total = total
	return pagination, infos, nil
}

// ListUpcoming 即将开始的活动
func (s *activityService) ListUpcoming(regionID uint, page, pageSize int) (*utils.Pagination, []dto.ActivityInfo, error) {
	pagination := utils.NewPagination(page, pageSize)
	list, total, err := s.repo.ListUpcoming(regionID, pagination)
	if err != nil {
		return nil, nil, err
	}
	infos := make([]dto.ActivityInfo, 0, len(list))
	for i := range list {
		infos = append(infos, *toActivityInfo(&list[i]))
	}
	pagination.Total = total
	return pagination, infos, nil
}

// ListEnded 已结束的活动
func (s *activityService) ListEnded(regionID uint, page, pageSize int) (*utils.Pagination, []dto.ActivityInfo, error) {
	pagination := utils.NewPagination(page, pageSize)
	list, total, err := s.repo.ListEnded(regionID, pagination)
	if err != nil {
		return nil, nil, err
	}
	infos := make([]dto.ActivityInfo, 0, len(list))
	for i := range list {
		infos = append(infos, *toActivityInfo(&list[i]))
	}
	pagination.Total = total
	return pagination, infos, nil
}

// UpdateStatus 手动更新活动状态
func (s *activityService) UpdateStatus(id uint, status int) error {
	if _, err := s.repo.FindByID(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrActivityNotFound
		}
		return err
	}
	return s.repo.Update(id, map[string]interface{}{"status": status})
}

// AutoUpdateStatus 根据时间自动推进活动状态（定时任务调用）
func (s *activityService) AutoUpdateStatus() (int64, error) {
	return s.repo.UpdateStatusByTime(time.Now())
}

// Statistics 活动统计
func (s *activityService) Statistics(regionID uint) (*dto.ActivityStatistics, error) {
	stats := &dto.ActivityStatistics{}

	// 总数
	_, total, err := s.repo.List(regionID, repository.ActivityListQuery{}, utils.NewPagination(1, 1))
	if err != nil {
		return nil, err
	}
	stats.TotalActivities = total

	// 进行中
	ongoingStatus := model.ActivityStatusOngoing
	_, ongoing, err := s.repo.List(regionID, repository.ActivityListQuery{Status: &ongoingStatus}, utils.NewPagination(1, 1))
	if err != nil {
		return nil, err
	}
	stats.OngoingActivities = ongoing

	// 待开始
	pendingStatus := model.ActivityStatusPending
	_, pending, err := s.repo.List(regionID, repository.ActivityListQuery{Status: &pendingStatus}, utils.NewPagination(1, 1))
	if err != nil {
		return nil, err
	}
	stats.PendingActivities = pending

	// 已结束
	endedStatus := model.ActivityStatusEnded
	_, ended, err := s.repo.List(regionID, repository.ActivityListQuery{Status: &endedStatus}, utils.NewPagination(1, 1))
	if err != nil {
		return nil, err
	}
	stats.EndedActivities = ended

	return stats, nil
}
