// Package service 同城114业务逻辑层 - 举报
package service

import (
	"errors"
	"fmt"
	"math/rand"
	"time"

	"wuchang-tongcheng/internal/modules/dh114/dto"
	"wuchang-tongcheng/internal/modules/dh114/model"
	"wuchang-tongcheng/internal/modules/dh114/repository"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

var (
	ErrReportNotFound = errors.New("举报记录不存在")
)

// ReportService 举报业务接口
type ReportService interface {
	Create(regionID uint, userID uint, userName string, req *dto.CreateReportRequest) (*dto.ReportInfo, error)
	GetByID(id uint) (*dto.ReportInfo, error)
	Process(id uint, handlerID uint, handlerName string, req *dto.ProcessReportRequest) error
	Delete(id uint) error
	List(regionID uint, req *dto.ReportListRequest) (*utils.Pagination, []dto.ReportInfo, *dto.ReportStats, error)
	Stats(regionID uint) (*dto.ReportStats, error)
}

type reportService struct {
	repo repository.ReportRepository
}

// NewReportService 创建举报 service 实例
func NewReportService(repo repository.ReportRepository) ReportService {
	return &reportService{repo: repo}
}

// reportStatusText 举报状态文本（与前端 statusMap 对齐）
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

// slaDuration SLA 时长（48 小时）
const slaDuration = 48 * time.Hour

// toReportInfo model -> dto
func toReportInfo(r *model.Dh114Report) *dto.ReportInfo {
	info := &dto.ReportInfo{
		ID:             r.ID,
		ReportNo:       r.ReportNo,
		ReporterID:     r.ReporterID,
		ReporterName:   r.ReporterName,

		TargetType:     r.TargetType,
		TargetID:       r.TargetID,
		TargetName:     r.TargetName,

		ReportType:     r.ReportType,
		Reason:         r.ReportReason,
		Description:    r.Description,
		ContactInfo:    r.ContactInfo,

		Status:         r.Status,
		StatusText:     reportStatusText(r.Status),
		HandlerID:      r.HandlerID,
		HandlerName:    r.HandlerName,
		HandleResult:   r.HandleResult,
		HandledAt:      r.HandledAt,
		PenaltyType:    r.PenaltyType,

		RegionID:       r.RegionID,
		CreatedAt:      r.CreatedAt,
		UpdatedAt:      r.UpdatedAt,
	}
	// 被举报对象映射：dh114 类型时填充 dh114_id；user 类型时填充 reported_user_id/reported_user_name
	switch r.TargetType {
	case model.ReportTargetDh114:
		info.Dh114ID = r.TargetID
	case model.ReportTargetUser:
		info.ReportedUserID = r.TargetID
		info.ReportedUserName = r.TargetName
	}
	if r.EvidenceImages != nil {
		info.EvidenceImages = r.EvidenceImages
	}
	// SLA 截止时间 = 创建时间 + 48h
	if r.Status == model.ReportStatusPending {
		deadline := r.CreatedAt.Add(slaDuration)
		info.SlaDeadline = &deadline
	}
	return info
}

// generateReportNo 生成举报单号
func generateReportNo() string {
	return fmt.Sprintf("DH114RP%s%06d", time.Now().Format("20060102150405"), rand.Intn(1000000))
}

// Create 创建举报
func (s *reportService) Create(regionID uint, userID uint, userName string, req *dto.CreateReportRequest) (*dto.ReportInfo, error) {
	// 确定被举报对象类型与 ID
	targetType := req.TargetType
	if targetType == "" {
		targetType = model.ReportTargetDh114
	}
	targetID := req.TargetID
	if targetID == 0 {
		targetID = req.Dh114ID
	}
	if targetID == 0 {
		return nil, errors.New("被举报对象 ID 不能为空")
	}

	rp := &model.Dh114Report{
		ReportNo:     generateReportNo(),
		ReporterID:   userID,
		ReporterName: userName,
		TargetType:   targetType,
		TargetID:     targetID,
		TargetName:   req.TargetName,
		ReportType:   req.ReportType,
		ReportReason: req.Reason,
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
	rp, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrReportNotFound
		}
		return err
	}
	_ = rp

	// 合并 handle_result 与 handle_note（兼容前端与任务规范）
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
	return s.repo.Update(id, fields)
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
	// 前端使用 dh114_id 作为筛选条件，等价于 target_type='dh114' AND target_id=?
	query := repository.ReportListQuery{
		Keyword:    req.Keyword,
		Status:     req.Status,
		ReportType: req.ReportType,
		TargetType: req.TargetType,
	}
	if req.Dh114ID > 0 {
		query.TargetType = model.ReportTargetDh114
		query.TargetID = req.Dh114ID
	}

	list, total, err := s.repo.List(regionID, query, pagination)
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
		// 统计失败不影响列表，返回零值
		stats = &dto.ReportStats{}
	}
	return pagination, infos, stats, nil
}

// Stats 举报统计
func (s *reportService) Stats(regionID uint) (*dto.ReportStats, error) {
	total, pending, processed, err := s.repo.CountByStatus(regionID)
	if err != nil {
		return nil, err
	}
	return &dto.ReportStats{
		Total:     total,
		Pending:   pending,
		Processed: processed,
	}, nil
}
