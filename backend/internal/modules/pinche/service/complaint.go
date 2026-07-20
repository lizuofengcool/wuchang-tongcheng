// Package service 同城拼车出行业务逻辑层 - 举报（复用 pinche_complaints 表）
package service

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"wuchang-tongcheng/internal/modules/pinche/dto"
	"wuchang-tongcheng/internal/modules/pinche/model"
	"wuchang-tongcheng/internal/modules/pinche/repository"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

var (
	ErrComplaintNotFound = errors.New("举报不存在")
	ErrComplaintInvalid  = errors.New("举报参数无效")
)

// ComplaintService 举报业务接口
type ComplaintService interface {
	AdminList(req *dto.ComplaintListRequest) (*dto.ComplaintListResult, error)
	GetByID(id uint) (*dto.ComplaintInfo, error)
	Create(regionID, operatorID uint, operatorName string, req *dto.CreateComplaintRequest) (*dto.ComplaintInfo, error)
	Process(id, operatorID uint, operatorName string, req *dto.ProcessComplaintRequest) error
	Stats() (*dto.ComplaintStatsResponse, error)
}

type complaintService struct {
	repo repository.ComplaintRepository
}

// NewComplaintService 创建举报 service 实例
func NewComplaintService(repo repository.ComplaintRepository) ComplaintService {
	return &complaintService{repo: repo}
}

// complaintStatusText 状态文本
func complaintStatusText(status int) string {
	switch status {
	case model.ComplaintStatusPending:
		return "待处理"
	case model.ComplaintStatusHandling:
		return "处理中"
	case model.ComplaintStatusHandled:
		return "已处理"
	case model.ComplaintStatusRejected:
		return "已驳回"
	}
	return ""
}

// complaintReportNo 生成举报单号
func complaintReportNo(id uint) string {
	return fmt.Sprintf("RPT%08d", id)
}

// toComplaintInfo model -> dto
func toComplaintInfo(c *model.PincheComplaint) *dto.ComplaintInfo {
	info := &dto.ComplaintInfo{
		ID:             c.ID,
		RegionID:       c.RegionID,
		ReportNo:       complaintReportNo(c.ID),
		PincheID:       c.PincheID,
		BookingID:      c.BookingID,
		TripID:         c.TripID,
		ReporterID:     c.ComplainantID,
		ReporterName:   c.ComplainantName,
		RespondentID:   c.RespondentID,
		RespondentName: c.RespondentName,
		ReportType:     c.ComplaintType,
		Reason:         c.ComplaintReason,
		Description:    c.Description,
		Status:         c.Status,
		StatusText:     complaintStatusText(c.Status),
		HandlerID:      c.HandlerID,
		HandlerName:    c.HandlerName,
		HandleResult:   c.HandleResult,
		HandledAt:      c.HandledAt,
		PenaltyType:    c.PenaltyType,
		PenaltyUserID:  c.PenaltyUserID,
		SLADeadline:    c.SLADeadline,
		CreatedAt:      c.CreatedAt,
	}
	// 目标类型映射：优先使用 complaint_type 中的目标信息
	// 若 pinche_id > 0 视为拼车信息举报
	if c.PincheID > 0 {
		info.TargetType = "pinche"
		info.TargetID = c.PincheID
	} else if c.BookingID != nil && *c.BookingID > 0 {
		info.TargetType = "booking"
		info.TargetID = *c.BookingID
	} else if c.TripID != nil && *c.TripID > 0 {
		info.TargetType = "trip"
		info.TargetID = *c.TripID
	}
	// 证据图片处理
	if c.EvidenceImages != nil && len(c.EvidenceImages) > 0 {
		info.EvidenceImages = c.EvidenceImages
		info.EvidenceURLs = c.EvidenceImages
		// 计算证据数量
		var arr []interface{}
		if c.EvidenceImages.Parse(&arr) == nil {
			info.EvidenceCount = len(arr)
		}
	}
	return info
}

// AdminList 管理后台举报列表
func (s *complaintService) AdminList(req *dto.ComplaintListRequest) (*dto.ComplaintListResult, error) {
	if req == nil {
		req = &dto.ComplaintListRequest{}
	}
	page := req.Page
	if page <= 0 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	pagination := utils.NewPagination(page, pageSize)

	opts := repository.ComplaintListOptions{
		ComplaintType: req.ReportType,
		Status:        req.Status,
		Keyword:       req.Keyword,
	}
	// 跨地区：regionID=0
	list, total, err := s.repo.List(0, pagination, opts)
	if err != nil {
		return nil, err
	}

	result := make([]dto.ComplaintInfo, 0, len(list))
	for i := range list {
		// 日期过滤（在内存中过滤，repo 没有暴露日期过滤）
		if !matchComplaintDate(&list[i], req.StartDate, req.EndDate) {
			continue
		}
		// target_type 过滤
		if req.TargetType != "" {
			info := toComplaintInfo(&list[i])
			if info.TargetType != req.TargetType {
				continue
			}
			result = append(result, *info)
			continue
		}
		result = append(result, *toComplaintInfo(&list[i]))
	}

	// 重新计算过滤后的总数（近似：如果使用了日期或目标类型过滤）
	finalTotal := total
	if req.StartDate != "" || req.EndDate != "" || req.TargetType != "" {
		finalTotal = int64(len(result))
	}

	stats, err := s.Stats()
	if err != nil {
		return nil, err
	}

	return &dto.ComplaintListResult{
		List:     result,
		Total:    finalTotal,
		Page:     page,
		PageSize: pageSize,
		Stats:    *stats,
	}, nil
}

// matchComplaintDate 内存日期过滤
func matchComplaintDate(c *model.PincheComplaint, startDate, endDate string) bool {
	if startDate == "" && endDate == "" {
		return true
	}
	t := c.CreatedAt
	if startDate != "" {
		if st, err := time.Parse("2006-01-02", startDate); err == nil {
			if t.Before(st) {
				return false
			}
		}
	}
	if endDate != "" {
		if et, err := time.Parse("2006-01-02", endDate); err == nil {
			end := et.Add(24 * time.Hour)
			if t.After(end) {
				return false
			}
		}
	}
	return true
}

// GetByID 举报详情
func (s *complaintService) GetByID(id uint) (*dto.ComplaintInfo, error) {
	c, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrComplaintNotFound
		}
		return nil, err
	}
	return toComplaintInfo(c), nil
}

// Create 创建举报
func (s *complaintService) Create(regionID, operatorID uint, operatorName string, req *dto.CreateComplaintRequest) (*dto.ComplaintInfo, error) {
	if req == nil {
		return nil, ErrComplaintInvalid
	}
	if req.Reason == "" {
		return nil, ErrComplaintInvalid
	}
	reportType := req.ReportType
	if reportType == "" {
		reportType = "other"
	}
	c := &model.PincheComplaint{
		PincheID:        req.PincheID,
		BookingID:       req.BookingID,
		TripID:          req.TripID,
		ComplainantID:   operatorID,
		ComplainantName: operatorName,
		ComplaintType:   reportType,
		ComplaintReason: req.Reason,
		Description:     req.Reason,
		Status:          model.ComplaintStatusPending,
	}
	c.RegionID = regionID
	// 根据前端传入的 target_type / target_id 自动填充关联 ID
	if req.TargetType != "" && req.TargetID > 0 {
		switch strings.ToLower(req.TargetType) {
		case "pinche":
			c.PincheID = req.TargetID
		case "booking":
			bid := req.TargetID
			c.BookingID = &bid
		case "trip":
			tid := req.TargetID
			c.TripID = &tid
		}
	}
	if err := s.repo.Create(c); err != nil {
		return nil, err
	}
	return toComplaintInfo(c), nil
}

// Process 处理举报
func (s *complaintService) Process(id, operatorID uint, operatorName string, req *dto.ProcessComplaintRequest) error {
	if req == nil {
		return ErrComplaintInvalid
	}
	c, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrComplaintNotFound
		}
		return err
	}
	_ = c

	status := req.Status
	if status == 0 {
		// 根据 result 推断 status
		switch req.Result {
		case "valid", "partial":
			status = model.ComplaintStatusHandled
		case "invalid":
			status = model.ComplaintStatusRejected
		}
	}
	if status == 0 {
		status = model.ComplaintStatusHandled
	}

	fields := map[string]interface{}{
		"handler_id":    operatorID,
		"handler_name":  operatorName,
		"handle_result": req.HandleNote,
		"status":        status,
		"handled_at":    gorm.Expr("NOW()"),
	}
	if req.PenaltyType != "" {
		fields["penalty_type"] = req.PenaltyType
	}
	return s.repo.Update(id, fields)
}

// Stats 举报统计
func (s *complaintService) Stats() (*dto.ComplaintStatsResponse, error) {
	resp := &dto.ComplaintStatsResponse{}
	pending, err := s.repo.CountByStatus(0, model.ComplaintStatusPending)
	if err != nil {
		return nil, err
	}
	handling, err := s.repo.CountByStatus(0, model.ComplaintStatusHandling)
	if err != nil {
		return nil, err
	}
	handled, err := s.repo.CountByStatus(0, model.ComplaintStatusHandled)
	if err != nil {
		return nil, err
	}
	rejected, err := s.repo.CountByStatus(0, model.ComplaintStatusRejected)
	if err != nil {
		return nil, err
	}
	resp.Total = pending + handling + handled + rejected
	resp.Pending = pending + handling
	resp.Processed = handled
	// high_priority: 诈骗/安全问题（按 complaint_type 统计需要新接口，这里近似用 0）
	resp.HighPriority = 0
	return resp, nil
}
