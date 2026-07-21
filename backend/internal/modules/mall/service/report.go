// Package service 同城商城业务逻辑层 - 举报
package service

import (
	"errors"
	"fmt"
	"math/rand"
	"time"

	"wuchang-tongcheng/internal/modules/mall/dto"
	"wuchang-tongcheng/internal/modules/mall/model"
	"wuchang-tongcheng/internal/modules/mall/repository"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

var (
	ErrReportNotFound = errors.New("举报记录不存在")
)

// 举报处理 SLA 时长（48 小时）
const reportSLADuration = 48 * time.Hour

// ReportService 举报业务接口
type ReportService interface {
	Create(regionID uint, userID uint, userName string, req *dto.CreateReportRequest) (*dto.ReportInfo, error)
	GetByID(id uint) (*dto.ReportInfo, error)
	Process(id uint, handlerID uint, handlerName string, req *dto.ProcessReportRequest) error
	Delete(id uint) error
	List(regionID uint, req *dto.ReportListRequest) (*utils.Pagination, []dto.ReportInfo, *dto.ReportStats, error)
	ListByUser(userID uint, page, pageSize int) (*utils.Pagination, []dto.ReportInfo, error)
	ListByTarget(targetType string, targetID uint) ([]dto.ReportInfo, error)
	Stats(regionID uint) (*dto.ReportStats, error)
}

type reportService struct {
	repo repository.ReportRepository
}

// NewReportService 创建举报 service 实例
func NewReportService(repo repository.ReportRepository) ReportService {
	return &reportService{repo: repo}
}

// reportStatusText 举报状态文本
func reportStatusText(s int) string {
	switch s {
	case model.ReportStatusPending:
		return "待处理"
	case model.ReportStatusWarning:
		return "已核实警告"
	case model.ReportStatusTakedown:
		return "已下架"
	case model.ReportStatusBanned:
		return "已封号"
	case model.ReportStatusDismissed:
		return "已驳回"
	case model.ReportStatusTransferred:
		return "已转交"
	}
	return ""
}

// toReportInfo model -> dto
func toReportInfo(r *model.Report) *dto.ReportInfo {
	info := &dto.ReportInfo{
		ID:              r.ID,
		ReportNo:        r.ReportNo,
		ReporterID:      r.ReporterID,
		ReporterName:    r.ReporterName,
		TargetType:      r.TargetType,
		TargetID:        r.TargetID,
		TargetName:      r.TargetName,
		ReportType:      r.ReportType,
		ReportReason:    r.ReportReason,
		Description:     r.Description,
		ContactInfo:     r.ContactInfo,
		Status:          r.Status,
		StatusText:      reportStatusText(r.Status),
		HandlerID:       r.HandlerID,
		HandlerName:     r.HandlerName,
		HandleResult:    r.HandleResult,
		HandledAt:       r.HandledAt,
		PenaltyType:     r.PenaltyType,
		PenaltyTargetID: r.PenaltyTargetID,
		RegionID:        r.RegionID,
		CreatedAt:       r.CreatedAt,
		UpdatedAt:       r.UpdatedAt,
	}
	if r.EvidenceImages != nil {
		info.EvidenceImages = r.EvidenceImages
	}
	return info
}

// generateReportNo 生成举报单号
func generateReportNo() string {
	return fmt.Sprintf("MALRP%s%06d", time.Now().Format("20060102150405"), rand.Intn(1000000))
}

// Create 创建举报
func (s *reportService) Create(regionID uint, userID uint, userName string, req *dto.CreateReportRequest) (*dto.ReportInfo, error) {
	if req.TargetID == 0 {
		return nil, errors.New("被举报对象 ID 不能为空")
	}

	rp := &model.Report{
		ReportNo:     generateReportNo(),
		ReporterID:   userID,
		ReporterName: userName,
		TargetType:   req.TargetType,
		TargetID:     req.TargetID,
		TargetName:   req.TargetName,
		ReportType:   req.ReportType,
		ReportReason: req.ReportReason,
		Description:  req.Description,
		ContactInfo:  req.ContactInfo,
		Status:       model.ReportStatusPending,
	}
	rp.RegionID = regionID
	if req.EvidenceImages != nil {
		if b, err := model.FromJSON(req.EvidenceImages); err == nil {
			rp.EvidenceImages = b
		}
	}

	if err := s.repo.Create(rp); err != nil {
		return nil, err
	}
	return toReportInfo(rp), nil
}

// GetByID 获取举报详情
func (s *reportService) GetByID(id uint) (*dto.ReportInfo, error) {
	rp, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrReportNotFound
		}
		return nil, err
	}
	return toReportInfo(rp), nil
}

// Process 处理举报
func (s *reportService) Process(id uint, handlerID uint, handlerName string, req *dto.ProcessReportRequest) error {
	if _, err := s.repo.FindByID(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrReportNotFound
		}
		return err
	}

	// 合并 handle_result 与 handle_note（兼容前端）
	handleResult := req.HandleResult
	if handleResult == "" && req.HandleNote != "" {
		handleResult = req.HandleNote
	}

	now := time.Now()
	fields := map[string]interface{}{
		"status":        req.Status,
		"handler_id":    handlerID,
		"handler_name":  handlerName,
		"handle_result": handleResult,
		"handled_at":    &now,
	}
	if req.PenaltyType != "" {
		fields["penalty_type"] = req.PenaltyType
	}
	return s.repo.UpdateFields(id, fields)
}

// Delete 删除举报
func (s *reportService) Delete(id uint) error {
	if _, err := s.repo.FindByID(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrReportNotFound
		}
		return err
	}
	return s.repo.Delete(id)
}

// List 举报列表
func (s *reportService) List(regionID uint, req *dto.ReportListRequest) (*utils.Pagination, []dto.ReportInfo, *dto.ReportStats, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	opts := repository.ReportListOptions{
		TargetType: req.TargetType,
		TargetID:   req.TargetID,
		ReporterID: req.ReporterID,
		HandlerID:  req.HandlerID,
		Status:     req.Status,
		ReportType: req.ReportType,
		Keyword:    req.Keyword,
		StartDate:  req.StartDate,
		EndDate:    req.EndDate,
		RegionID:   regionID,
	}
	list, total, err := s.repo.List(opts, pagination)
	if err != nil {
		return nil, nil, nil, err
	}
	infos := make([]dto.ReportInfo, 0, len(list))
	for i := range list {
		infos = append(infos, *toReportInfo(&list[i]))
	}
	pagination.Total = total

	// 统计信息
	stats, err := s.Stats(regionID)
	if err != nil {
		stats = &dto.ReportStats{}
	}
	return pagination, infos, stats, nil
}

// ListByUser 按举报人列出
func (s *reportService) ListByUser(userID uint, page, pageSize int) (*utils.Pagination, []dto.ReportInfo, error) {
	pagination := utils.NewPagination(page, pageSize)
	list, total, err := s.repo.ListByUser(userID, pagination)
	if err != nil {
		return nil, nil, err
	}
	infos := make([]dto.ReportInfo, 0, len(list))
	for i := range list {
		infos = append(infos, *toReportInfo(&list[i]))
	}
	pagination.Total = total
	return pagination, infos, nil
}

// ListByTarget 按被举报对象列出
func (s *reportService) ListByTarget(targetType string, targetID uint) ([]dto.ReportInfo, error) {
	list, err := s.repo.ListByTarget(targetType, targetID)
	if err != nil {
		return nil, err
	}
	infos := make([]dto.ReportInfo, 0, len(list))
	for i := range list {
		infos = append(infos, *toReportInfo(&list[i]))
	}
	return infos, nil
}

// Stats 举报统计
func (s *reportService) Stats(regionID uint) (*dto.ReportStats, error) {
	opts := repository.ReportStatsOptions{
		RegionID: regionID,
	}
	result, err := s.repo.Stats(opts)
	if err != nil {
		return nil, err
	}
	return &dto.ReportStats{
		Total:     result.Total,
		Pending:   result.Pending,
		Processed: result.Processed,
		Valid:     result.Valid,
		Invalid:   result.Invalid,
	}, nil
}
