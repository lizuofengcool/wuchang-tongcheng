// Package service love 相亲交友业务逻辑层 - 举报
// 依据需求文档 1.10：4 维数据隔离（region_id + user_id）
package service

import (
	"errors"
	"fmt"
	"time"

	"wuchang-tongcheng/internal/modules/love/dto"
	"wuchang-tongcheng/internal/modules/love/model"
	"wuchang-tongcheng/internal/modules/love/repository"
	"wuchang-tongcheng/internal/pkg/utils"
)

var (
	ErrLoveReportNotFound       = errors.New("举报记录不存在")
	ErrLoveReportNoPermission   = errors.New("无权操作此举报")
	ErrLoveReportAlreadyHandled = errors.New("举报已处理")
	ErrLoveReportAppealExists   = errors.New("已提交过申诉")
)

// LoveReportService 举报业务接口
type LoveReportService interface {
	// C 端
	Create(reporterUserID, reporterLoveID uint, reporterNickname, reporterAvatar string, req *dto.CreateLoveReportRequest) (*dto.LoveReportInfo, error)
	GetByID(id uint, userID uint) (*dto.LoveReportInfo, error)
	ListByReporter(userID uint, req *dto.LoveReportListRequest) (*utils.Pagination, []dto.LoveReportInfo, error)
	Appeal(req *dto.LoveReportAppealRequest, userID uint) error
	// M 端
	List(req *dto.LoveReportListRequest) (*utils.Pagination, []dto.LoveReportInfo, error)
	ListPending(req *dto.LoveReportListRequest) (*utils.Pagination, []dto.LoveReportInfo, error)
	ListByTarget(userID uint, req *dto.LoveReportListRequest) (*utils.Pagination, []dto.LoveReportInfo, error)
	Handle(req *dto.LoveReportHandleRequest, handledBy uint) error
	HandleAppeal(req *dto.LoveReportAppealHandleRequest, handledBy uint) error
	UpdateRiskScore(id uint, score int) error
	Delete(id uint) error
	Stats() (*dto.LoveReportStatsResponse, error)
}

type loveReportService struct {
	repo repository.LoveReportRepository
}

// NewLoveReportService 创建举报 service
func NewLoveReportService(repo repository.LoveReportRepository) LoveReportService {
	return &loveReportService{repo: repo}
}

// reportStatusText 状态文本
func reportStatusText(s int) string {
	switch s {
	case model.ReportStatusPending:
		return "待处理"
	case model.ReportStatusHandling:
		return "处理中"
	case model.ReportStatusHandled:
		return "已处理"
	case model.ReportStatusRejected:
		return "已驳回"
	}
	return ""
}

// appealStatusText 申诉状态文本
func appealStatusText(s int) string {
	switch s {
	case model.AppealStatusNone:
		return "未申诉"
	case model.AppealStatusPending:
		return "申诉中"
	case model.AppealStatusApproved:
		return "申诉成功"
	case model.AppealStatusRejected:
		return "申诉驳回"
	}
	return ""
}

// toLoveReportInfo model -> dto
func toLoveReportInfo(r *model.LoveReport) dto.LoveReportInfo {
	return dto.LoveReportInfo{
		ID:               r.ID,
		ReportNo:         r.ReportNo,
		ReporterUserID:   r.ReporterUserID,
		ReporterLoveID:   r.ReporterLoveID,
		ReporterNickname: r.ReporterNickname,
		ReporterAvatar:   r.ReporterAvatar,
		TargetType:       r.TargetType,
		TargetUserID:     r.TargetUserID,
		TargetLoveID:     r.TargetLoveID,
		TargetNickname:   r.TargetNickname,
		TargetAvatar:     r.TargetAvatar,
		TargetID:         r.TargetID,
		ReasonType:       r.ReasonType,
		ReasonDetail:     r.ReasonDetail,
		EvidenceImages:   r.EvidenceImages,
		EvidenceVideos:   r.EvidenceVideos,
		EvidenceText:     r.EvidenceText,
		Status:           r.Status,
		StatusText:       reportStatusText(r.Status),
		HandledBy:        r.HandledBy,
		HandledAt:        r.HandledAt,
		HandleResult:     r.HandleResult,
		HandleRemark:     r.HandleRemark,
		PenaltyType:      r.PenaltyType,
		PenaltyDuration:  r.PenaltyDuration,
		PenaltyExpiredAt: r.PenaltyExpiredAt,
		AppealStatus:     r.AppealStatus,
		AppealReason:     r.AppealReason,
		AppealedAt:       r.AppealedAt,
		AppealResult:     r.AppealResult,
		RiskScore:        r.RiskScore,
		CreatedAt:        r.CreatedAt,
		UpdatedAt:        r.UpdatedAt,
	}
}

// ===== C 端 =====

func (s *loveReportService) Create(reporterUserID, reporterLoveID uint, reporterNickname, reporterAvatar string, req *dto.CreateLoveReportRequest) (*dto.LoveReportInfo, error) {
	r := &model.LoveReport{
		ReportNo:         fmt.Sprintf("LOVE-RPT-%d-%d", reporterUserID, time.Now().UnixNano()/1e6),
		ReporterUserID:   reporterUserID,
		ReporterLoveID:   reporterLoveID,
		ReporterNickname: reporterNickname,
		ReporterAvatar:   reporterAvatar,
		TargetType:       req.TargetType,
		TargetUserID:     req.TargetUserID,
		TargetLoveID:     req.TargetLoveID,
		TargetID:         req.TargetID,
		ReasonType:       req.ReasonType,
		ReasonDetail:     req.ReasonDetail,
		EvidenceText:     req.EvidenceText,
		Status:           model.ReportStatusPending,
	}
	if req.EvidenceImages != nil {
		if jb, err := model.FromJSON(req.EvidenceImages); err == nil {
			r.EvidenceImages = jb
		}
	}
	if req.EvidenceVideos != nil {
		if jb, err := model.FromJSON(req.EvidenceVideos); err == nil {
			r.EvidenceVideos = jb
		}
	}
	if r.TargetType == "" {
		r.TargetType = "user"
	}
	if err := s.repo.Create(r); err != nil {
		return nil, err
	}
	info := toLoveReportInfo(r)
	return &info, nil
}

func (s *loveReportService) GetByID(id uint, userID uint) (*dto.LoveReportInfo, error) {
	r, err := s.repo.FindByID(id)
	if err != nil {
		return nil, ErrLoveReportNotFound
	}
	// 举报人本人或管理员可查
	if r.ReporterUserID != userID && r.Status == model.ReportStatusPending {
		// 非举报人查看他人待处理举报 → 拒绝（M 端通过 admin 接口查询）
		return nil, ErrLoveReportNoPermission
	}
	info := toLoveReportInfo(r)
	return &info, nil
}

func (s *loveReportService) ListByReporter(userID uint, req *dto.LoveReportListRequest) (*utils.Pagination, []dto.LoveReportInfo, error) {
	list, total, err := s.repo.ListByReporter(userID, &req.Pagination)
	if err != nil {
		return nil, nil, err
	}
	infos := make([]dto.LoveReportInfo, 0, len(list))
	for i := range list {
		infos = append(infos, toLoveReportInfo(&list[i]))
	}
	req.Pagination.Total = total
	return &req.Pagination, infos, nil
}

// Appeal 举报人提交申诉
func (s *loveReportService) Appeal(req *dto.LoveReportAppealRequest, userID uint) error {
	r, err := s.repo.FindByID(req.ID)
	if err != nil {
		return ErrLoveReportNotFound
	}
	if r.ReporterUserID != userID {
		return ErrLoveReportNoPermission
	}
	if r.AppealStatus != model.AppealStatusNone {
		return ErrLoveReportAppealExists
	}
	return s.repo.UpdateAppeal(req.ID, req.AppealReason)
}

// ===== M 端 =====

func (s *loveReportService) List(req *dto.LoveReportListRequest) (*utils.Pagination, []dto.LoveReportInfo, error) {
	opts := repository.LoveReportListOptions{
		ReporterUserID: req.ReporterUserID,
		TargetUserID:   req.TargetUserID,
		TargetType:     req.TargetType,
		ReasonType:     req.ReasonType,
	}
	if req.Status != nil {
		opts.Status = req.Status
	}
	list, total, err := s.repo.List(&req.Pagination, opts)
	if err != nil {
		return nil, nil, err
	}
	infos := make([]dto.LoveReportInfo, 0, len(list))
	for i := range list {
		infos = append(infos, toLoveReportInfo(&list[i]))
	}
	req.Pagination.Total = total
	return &req.Pagination, infos, nil
}

func (s *loveReportService) ListPending(req *dto.LoveReportListRequest) (*utils.Pagination, []dto.LoveReportInfo, error) {
	list, total, err := s.repo.ListPending(&req.Pagination)
	if err != nil {
		return nil, nil, err
	}
	infos := make([]dto.LoveReportInfo, 0, len(list))
	for i := range list {
		infos = append(infos, toLoveReportInfo(&list[i]))
	}
	req.Pagination.Total = total
	return &req.Pagination, infos, nil
}

func (s *loveReportService) ListByTarget(userID uint, req *dto.LoveReportListRequest) (*utils.Pagination, []dto.LoveReportInfo, error) {
	list, total, err := s.repo.ListByTarget(userID, &req.Pagination)
	if err != nil {
		return nil, nil, err
	}
	infos := make([]dto.LoveReportInfo, 0, len(list))
	for i := range list {
		infos = append(infos, toLoveReportInfo(&list[i]))
	}
	req.Pagination.Total = total
	return &req.Pagination, infos, nil
}

// Handle 处理举报
func (s *loveReportService) Handle(req *dto.LoveReportHandleRequest, handledBy uint) error {
	r, err := s.repo.FindByID(req.ID)
	if err != nil {
		return ErrLoveReportNotFound
	}
	if r.Status == model.ReportStatusHandled {
		return ErrLoveReportAlreadyHandled
	}
	opts := repository.LoveReportHandleOptions{
		HandleResult:    req.HandleResult,
		HandleRemark:    req.HandleRemark,
		PenaltyType:     req.PenaltyType,
		PenaltyDuration: req.PenaltyDuration,
		HandledBy:       handledBy,
	}
	return s.repo.Handle(req.ID, opts)
}

// HandleAppeal 处理申诉
func (s *loveReportService) HandleAppeal(req *dto.LoveReportAppealHandleRequest, handledBy uint) error {
	r, err := s.repo.FindByID(req.ID)
	if err != nil {
		return ErrLoveReportNotFound
	}
	if r.AppealStatus != model.AppealStatusPending {
		return ErrLoveReportAppealExists
	}
	return s.repo.HandleAppeal(req.ID, req.AppealResult, req.AppealRemark, handledBy)
}

func (s *loveReportService) UpdateRiskScore(id uint, score int) error {
	if _, err := s.repo.FindByID(id); err != nil {
		return ErrLoveReportNotFound
	}
	return s.repo.UpdateRiskScore(id, score)
}

func (s *loveReportService) Delete(id uint) error {
	if _, err := s.repo.FindByID(id); err != nil {
		return ErrLoveReportNotFound
	}
	return s.repo.Delete(id)
}

// Stats 举报统计
func (s *loveReportService) Stats() (*dto.LoveReportStatsResponse, error) {
	today, _ := s.repo.CountToday()
	pending, _ := s.repo.CountByStatus(model.ReportStatusPending)
	handled, _ := s.repo.CountByStatus(model.ReportStatusHandled)
	appeals, _ := s.repo.CountAppeals()
	return &dto.LoveReportStatsResponse{
		TotalReports:   today + handled,
		TodayReports:   today,
		PendingReports: pending,
		HandledReports: handled,
		AppealReports:  appeals,
	}, nil
}
