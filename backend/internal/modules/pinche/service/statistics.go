// Package service 同城拼车出行业务逻辑层 - 统计
// 依据 v3.2.1 架构方案：日统计/周统计/月统计/总统计
// 对标哈啰出行/嘀嗒出行 数据分析
package service

import (
	"errors"
	"time"

	"wuchang-tongcheng/internal/modules/pinche/dto"
	"wuchang-tongcheng/internal/modules/pinche/model"
	"wuchang-tongcheng/internal/modules/pinche/repository"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

var (
	ErrStatisticNotFound = errors.New("统计记录不存在")
	ErrStatDateInvalid   = errors.New("统计日期区间无效")
)

// StatisticService 统计业务接口
type StatisticService interface {
	// 查询
	GetByID(id uint) (*dto.StatisticInfo, error)
	List(regionID uint, req *dto.StatisticListRequest) (*utils.Pagination, []dto.StatisticInfo, error)
	ListByUser(userID uint, page, pageSize int) (*utils.Pagination, []dto.StatisticInfo, error)
	ListByDateRange(regionID uint, startDate, endDate time.Time, statType string) ([]dto.StatisticInfo, error)

	// 汇总查询
	Overview(regionID uint, startDate, endDate time.Time) (*dto.StatisticOverviewResponse, error)

	// M 端管理
	Upsert(regionID uint, req *dto.StatisticUpsertRequest) (*dto.StatisticInfo, error)
	Delete(id uint) error
	UpdateStatus(id uint, status int) error
}

type statisticService struct {
	repo repository.StatisticRepository
}

// NewStatisticService 创建统计 service 实例
func NewStatisticService(repo repository.StatisticRepository) StatisticService {
	return &statisticService{repo: repo}
}

// statTypeText 统计类型文本
func statTypeText(t string) string {
	switch t {
	case model.StatTypeDaily:
		return "日统计"
	case model.StatTypeWeekly:
		return "周统计"
	case model.StatTypeMonthly:
		return "月统计"
	case model.StatTypeTotal:
		return "总统计"
	}
	return ""
}

// toStatisticInfo model -> dto
func toStatisticInfo(s *model.PincheStatistic) *dto.StatisticInfo {
	return &dto.StatisticInfo{
		ID:                s.ID,
		RegionID:          s.RegionID,
		StatDate:          s.StatDate,
		StatType:          s.StatType,
		StatTypeText:      statTypeText(s.StatType),
		UserID:            s.UserID,
		TotalTrips:        s.TotalTrips,
		CompletedTrips:    s.CompletedTrips,
		CancelledTrips:    s.CancelledTrips,
		TotalBookings:     s.TotalBookings,
		CompletedBookings: s.CompletedBookings,
		TotalRevenue:      s.TotalRevenue,
		TotalRefund:       s.TotalRefund,
		AvgRating:         s.AvgRating,
		AvgPrice:          s.AvgPrice,
		TotalDistance:     s.TotalDistance,
		TotalDuration:     s.TotalDuration,
		TotalPassengers:   s.TotalPassengers,
		TotalDrivers:      s.TotalDrivers,
		CreatedAt:         s.CreatedAt,
	}
}

// parseDate 解析日期字符串（YYYY-MM-DD）
func parseDate(s string) (time.Time, error) {
	if s == "" {
		return time.Time{}, errors.New("日期为空")
	}
	return time.ParseInLocation("2006-01-02", s, time.Local)
}

// GetByID 获取统计详情
func (s *statisticService) GetByID(id uint) (*dto.StatisticInfo, error) {
	stat, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrStatisticNotFound
		}
		return nil, err
	}
	return toStatisticInfo(stat), nil
}

// List 统计列表
func (s *statisticService) List(regionID uint, req *dto.StatisticListRequest) (*utils.Pagination, []dto.StatisticInfo, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	opts := repository.StatisticListOptions{
		StatType: req.StatType,
		UserID:   req.UserID,
	}
	if req.StartDate != "" {
		if sd, err := parseDate(req.StartDate); err == nil {
			opts.StartDate = &sd
		}
	}
	if req.EndDate != "" {
		if ed, err := parseDate(req.EndDate); err == nil {
			opts.EndDate = &ed
		}
	}
	list, total, err := s.repo.List(regionID, pagination, opts)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	result := make([]dto.StatisticInfo, 0, len(list))
	for i := range list {
		result = append(result, *toStatisticInfo(&list[i]))
	}
	return pagination, result, nil
}

// ListByUser 按用户列出统计
func (s *statisticService) ListByUser(userID uint, page, pageSize int) (*utils.Pagination, []dto.StatisticInfo, error) {
	pagination := utils.NewPagination(page, pageSize)
	list, total, err := s.repo.ListByUser(userID, pagination)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	result := make([]dto.StatisticInfo, 0, len(list))
	for i := range list {
		result = append(result, *toStatisticInfo(&list[i]))
	}
	return pagination, result, nil
}

// ListByDateRange 按日期区间列出统计
func (s *statisticService) ListByDateRange(regionID uint, startDate, endDate time.Time, statType string) ([]dto.StatisticInfo, error) {
	if endDate.Before(startDate) {
		return nil, ErrStatDateInvalid
	}
	list, err := s.repo.ListByDateRange(regionID, startDate, endDate, statType)
	if err != nil {
		return nil, err
	}
	result := make([]dto.StatisticInfo, 0, len(list))
	for i := range list {
		result = append(result, *toStatisticInfo(&list[i]))
	}
	return result, nil
}

// Overview 平台总览统计（指定日期区间）
func (s *statisticService) Overview(regionID uint, startDate, endDate time.Time) (*dto.StatisticOverviewResponse, error) {
	if endDate.Before(startDate) {
		return nil, ErrStatDateInvalid
	}
	// 优先使用 total 类型，回退到 daily
	totalTrips, completedTrips, cancelledTrips, totalRevenue, totalRefund, err := s.repo.SumByDateRange(regionID, startDate, endDate, model.StatTypeDaily)
	if err != nil {
		return nil, err
	}
	resp := &dto.StatisticOverviewResponse{
		TotalTrips:     totalTrips,
		CompletedTrips: completedTrips,
		CancelledTrips: cancelledTrips,
		TotalRevenue:   totalRevenue,
		TotalRefund:    totalRefund,
	}
	if totalTrips > 0 {
		resp.CompletionRate = float64(completedTrips) / float64(totalTrips)
		resp.CancellationRate = float64(cancelledTrips) / float64(totalTrips)
	}
	return resp, nil
}

// Upsert 创建/更新统计（M 端维护）
func (s *statisticService) Upsert(regionID uint, req *dto.StatisticUpsertRequest) (*dto.StatisticInfo, error) {
	statType := req.StatType
	if statType == "" {
		statType = model.StatTypeDaily
	}
	stat := &model.PincheStatistic{
		StatDate:          req.StatDate,
		StatType:          statType,
		UserID:            req.UserID,
		TotalTrips:        req.TotalTrips,
		CompletedTrips:    req.CompletedTrips,
		CancelledTrips:    req.CancelledTrips,
		TotalBookings:     req.TotalBookings,
		CompletedBookings: req.CompletedBookings,
		TotalRevenue:      req.TotalRevenue,
		TotalRefund:       req.TotalRefund,
		AvgRating:         req.AvgRating,
		AvgPrice:          req.AvgPrice,
		TotalDistance:     req.TotalDistance,
		TotalDuration:     req.TotalDuration,
		TotalPassengers:   req.TotalPassengers,
		TotalDrivers:      req.TotalDrivers,
	}
	stat.RegionID = regionID
	if err := s.repo.Upsert(stat); err != nil {
		return nil, err
	}
	return toStatisticInfo(stat), nil
}

// Delete 删除统计记录
func (s *statisticService) Delete(id uint) error {
	return s.repo.Delete(id)
}

// UpdateStatus 占位实现：统计记录无 status 字段，保留接口以兼容 handler
func (s *statisticService) UpdateStatus(id uint, status int) error {
	_ = status
	_ = id
	return nil
}
