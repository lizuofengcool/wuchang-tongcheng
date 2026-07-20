// Package service 同城零工兼职业务逻辑层 - 数据统计
// 对标斗米/青团：岗位/雇主/求职者/技能/地区/平台 多维度日统计
// 4 维数据隔离（region_id + user_id）
package service

import (
	"errors"
	"time"

	"wuchang-tongcheng/internal/modules/linggong/dto"
	"wuchang-tongcheng/internal/modules/linggong/model"
	"wuchang-tongcheng/internal/modules/linggong/repository"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

var (
	ErrStatisticNotFound  = errors.New("统计记录不存在")
	ErrStatDateInvalid    = errors.New("统计日期区间无效")
	ErrStatTypeInvalid    = errors.New("统计类型无效")
)

// StatisticService 数据统计业务接口
type StatisticService interface {
	// 查询
	GetByID(id uint) (*dto.StatisticInfo, error)
	List(regionID uint, req *dto.StatListRequest) (*utils.Pagination, []dto.StatisticInfo, error)
	ListByUser(userID uint, page, pageSize int) (*utils.Pagination, []dto.StatisticInfo, error)
	ListByDateRange(regionID uint, statType string, targetID uint, startDate, endDate string) (*utils.Pagination, []dto.StatisticInfo, error)
	ListByTarget(statType string, targetID uint, startDate, endDate string, page, pageSize int) (*utils.Pagination, []dto.StatisticInfo, error)

	// 汇总
	Overview(regionID uint, statType string, targetID uint, startDate, endDate string) (*dto.StatOverviewResponse, error)

	// M 端管理
	Upsert(regionID uint, req *dto.StatisticUpsertRequest) (*dto.StatisticInfo, error)
	Delete(id uint) error
	UpdateStatus(id uint, status int) error
}

type statisticService struct {
	repo repository.StatisticRepository
}

// NewStatisticService 创建数据统计 service 实例
func NewStatisticService(repo repository.StatisticRepository) StatisticService {
	return &statisticService{repo: repo}
}

// statTypeText 统计类型文本
func statTypeText(t string) string {
	switch t {
	case model.StatTypeLinggong:
		return "岗位统计"
	case model.StatTypeEmployer:
		return "雇主统计"
	case model.StatTypeWorker:
		return "求职者统计"
	case model.StatTypeSkill:
		return "技能统计"
	case model.StatTypeRegion:
		return "地区统计"
	case model.StatTypePlatform:
		return "平台统计"
	case model.StatTypeTask:
		return "任务统计"
	case model.StatTypeCategory:
		return "分类统计"
	}
	return ""
}

// toStatisticInfo model -> dto
func toStatisticInfo(s *model.LinggongStatistic) *dto.StatisticInfo {
	return &dto.StatisticInfo{
		ID:                s.ID,
		StatDate:          s.StatDate,
		StatType:          s.StatType,
		StatTypeText:      statTypeText(s.StatType),
		TargetID:          s.TargetID,
		TargetName:        s.TargetName,
		ImpressionCount:   s.ImpressionCount,
		ClickCount:        s.ClickCount,
		FavCount:          s.FavCount,
		ContactCount:      s.ContactCount,
		ApplicationCount:  s.ApplicationCount,
		HiredCount:        s.HiredCount,
		CompletedCount:    s.CompletedCount,
		DealCount:         s.DealCount,
		ConversionRate:    s.ConversionRate,
		TotalSalary:       s.TotalSalary,
		AvgSalary:         s.AvgSalary,
		AvgDealDays:       s.AvgDealDays,
		RegionID:          s.RegionID,
	}
}

// parseStatDate 解析日期字符串（YYYY-MM-DD）
func parseStatDate(s string) (time.Time, error) {
	if s == "" {
		return time.Time{}, nil
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
func (s *statisticService) List(regionID uint, req *dto.StatListRequest) (*utils.Pagination, []dto.StatisticInfo, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	query := repository.StatListQuery{
		RegionID: regionID,
		StatType: req.StatType,
		TargetID: req.TargetID,
		Keyword:  req.Keyword,
	}
	if req.StartDate != "" {
		if sd, err := parseStatDate(req.StartDate); err == nil {
			query.StartDate = sd
		}
	}
	if req.EndDate != "" {
		if ed, err := parseStatDate(req.EndDate); err == nil {
			query.EndDate = ed
		}
	}
	list, total, err := s.repo.List(query, pagination)
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

// ListByUser 按用户列出统计（兼容接口，按 regionID=0 查询平台统计）
func (s *statisticService) ListByUser(userID uint, page, pageSize int) (*utils.Pagination, []dto.StatisticInfo, error) {
	pagination := utils.NewPagination(page, pageSize)
	// linggong_statistics 表无 user_id 字段，按平台类型查询并兼容接口
	query := repository.StatListQuery{
		StatType: model.StatTypePlatform,
		TargetID: userID,
	}
	list, total, err := s.repo.List(query, pagination)
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
func (s *statisticService) ListByDateRange(regionID uint, statType string, targetID uint, startDate, endDate string) (*utils.Pagination, []dto.StatisticInfo, error) {
	pagination := utils.NewPagination(1, 100)
	sd, err := parseStatDate(startDate)
	if err != nil {
		return nil, nil, ErrStatDateInvalid
	}
	ed, err := parseStatDate(endDate)
	if err != nil {
		return nil, nil, ErrStatDateInvalid
	}
	if !sd.IsZero() && !ed.IsZero() && ed.Before(sd) {
		return nil, nil, ErrStatDateInvalid
	}
	query := repository.StatListQuery{
		RegionID:  regionID,
		StatType:  statType,
		TargetID:  targetID,
		StartDate: sd,
		EndDate:   ed,
	}
	list, total, err := s.repo.List(query, pagination)
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

// ListByTarget 按 target 反查
func (s *statisticService) ListByTarget(statType string, targetID uint, startDate, endDate string, page, pageSize int) (*utils.Pagination, []dto.StatisticInfo, error) {
	pagination := utils.NewPagination(page, pageSize)
	sd, err := parseStatDate(startDate)
	if err != nil {
		return nil, nil, ErrStatDateInvalid
	}
	ed, err := parseStatDate(endDate)
	if err != nil {
		return nil, nil, ErrStatDateInvalid
	}
	list, total, err := s.repo.ListByTarget(statType, targetID, sd, ed, pagination)
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

// Overview 平台总览统计（指定日期区间）
func (s *statisticService) Overview(regionID uint, statType string, targetID uint, startDate, endDate string) (*dto.StatOverviewResponse, error) {
	sd, err := parseStatDate(startDate)
	if err != nil {
		return nil, ErrStatDateInvalid
	}
	ed, err := parseStatDate(endDate)
	if err != nil {
		return nil, ErrStatDateInvalid
	}
	if !sd.IsZero() && !ed.IsZero() && ed.Before(sd) {
		return nil, ErrStatDateInvalid
	}
	if statType == "" {
		statType = model.StatTypePlatform
	}
	summary, err := s.repo.SumByRange(statType, targetID, sd, ed)
	if err != nil {
		return nil, err
	}
	resp := &dto.StatOverviewResponse{
		ImpressionCount:  summary.ImpressionCount,
		ClickCount:       summary.ClickCount,
		FavCount:         summary.FavCount,
		ContactCount:     summary.ContactCount,
		ApplicationCount: summary.ApplicationCount,
		HiredCount:       summary.HiredCount,
		CompletedCount:   summary.CompletedCount,
		DealCount:        summary.DealCount,
		TotalSalary:      summary.TotalSalary,
		AvgSalary:        summary.AvgSalary,
	}
	if summary.ImpressionCount > 0 {
		resp.ConversionRate = float64(summary.DealCount) / float64(summary.ImpressionCount)
	}
	_ = regionID
	return resp, nil
}

// Upsert 创建/更新统计（M 端维护）
func (s *statisticService) Upsert(regionID uint, req *dto.StatisticUpsertRequest) (*dto.StatisticInfo, error) {
	statType := req.StatType
	if statType == "" {
		statType = model.StatTypeLinggong
	}
	stat := &model.LinggongStatistic{
		StatDate:         req.StatDate,
		StatType:         statType,
		TargetID:         req.TargetID,
		TargetName:       req.TargetName,
		ImpressionCount:  req.ImpressionCount,
		ClickCount:       req.ClickCount,
		FavCount:         req.FavCount,
		ContactCount:     req.ContactCount,
		ApplicationCount: req.ApplicationCount,
		HiredCount:       req.HiredCount,
		CompletedCount:   req.CompletedCount,
		DealCount:        req.DealCount,
		ConversionRate:   req.ConversionRate,
		TotalSalary:      req.TotalSalary,
		AvgSalary:        req.AvgSalary,
		AvgDealDays:      req.AvgDealDays,
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

// UpdateStatus 占位实现：统计表无 status 字段，保留接口以兼容 handler
func (s *statisticService) UpdateStatus(id uint, status int) error {
	_ = status
	_ = id
	return nil
}
