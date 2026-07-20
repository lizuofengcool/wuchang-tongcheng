// Package service 同城零工兼职业务逻辑层 - 纠纷
// 对标闲鱼/瓜子纠纷处理：工单 + 证据 + 调解 + 仲裁 + 申诉
// 4 维数据隔离（region_id + user_id）
package service

import (
	"errors"
	"fmt"
	"math/rand"
	"time"

	"wuchang-tongcheng/internal/modules/linggong/dto"
	"wuchang-tongcheng/internal/modules/linggong/model"
	"wuchang-tongcheng/internal/modules/linggong/repository"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

var (
	ErrDisputeNotFound     = errors.New("纠纷不存在")
	ErrDisputeNoPermission = errors.New("无权操作此纠纷")
	ErrDisputeStatusInvalid = errors.New("纠纷状态不允许此操作")
)

// DisputeService 纠纷业务接口
type DisputeService interface {
	// C 端
	Create(regionID uint, userID uint, req *dto.CreateDisputeRequest) (*dto.DisputeInfo, error)
	GetByID(id uint) (*dto.DisputeInfo, error)
	GetByDisputeNo(no string) (*dto.DisputeInfo, error)
	List(regionID uint, req *dto.DisputeListRequest) (*utils.Pagination, []dto.DisputeInfo, error)
	ListByLinggong(linggongID uint, page, pageSize int) (*utils.Pagination, []dto.DisputeInfo, error)
	ListByApplicant(applicantID uint, page, pageSize int) (*utils.Pagination, []dto.DisputeInfo, error)
	ListByRespondent(respondentID uint, page, pageSize int) (*utils.Pagination, []dto.DisputeInfo, error)

	// 处理（M 端调解/仲裁）
	Handle(id uint, handlerID uint, handlerName string, req *dto.DisputeHandleRequest) error

	// M 端管理
	AdminList(req *dto.DisputeListRequest) (*utils.Pagination, []dto.DisputeInfo, error)
}

type disputeService struct {
	repo repository.DisputeRepository
}

// NewDisputeService 创建纠纷 service 实例
func NewDisputeService(repo repository.DisputeRepository) DisputeService {
	return &disputeService{repo: repo}
}

// disputeStatusText 纠纷状态文本
func disputeStatusText(s int) string {
	switch s {
	case model.DisputeStatusPending:
		return "待处理"
	case model.DisputeStatusProcessing:
		return "处理中"
	case model.DisputeStatusMediated:
		return "已调解"
	case model.DisputeStatusArbitrated:
		return "已仲裁"
	case model.DisputeStatusResolved:
		return "已解决"
	case model.DisputeStatusRejected:
		return "已驳回"
	case model.DisputeStatusCanceled:
		return "已取消"
	case model.DisputeStatusAppealed:
		return "申诉中"
	}
	return ""
}

// disputeTypeText 纠纷类型文本
func disputeTypeText(t string) string {
	switch t {
	case model.DisputeTypeSalary:
		return "薪资纠纷"
	case model.DisputeTypeQuality:
		return "工作质量纠纷"
	case model.DisputeTypeAttendance:
		return "考勤纠纷"
	case model.DisputeTypeBreach:
		return "违约"
	case model.DisputeTypeDiscrimination:
		return "歧视"
	case model.DisputeTypeHarassment:
		return "骚扰"
	case model.DisputeTypeFraud:
		return "欺诈"
	case model.DisputeTypeSafety:
		return "安全问题"
	case model.DisputeTypeOther:
		return "其他"
	}
	return ""
}

// disputeApplicantTypeText 申请人类型文本
func disputeApplicantTypeText(t string) string {
	switch t {
	case model.DisputeApplicantWorker:
		return "求职者"
	case model.DisputeApplicantEmployer:
		return "雇主"
	case model.DisputeApplicantPlatform:
		return "平台"
	}
	return ""
}

// disputeFinalResultText 最终结果文本
func disputeFinalResultText(r string) string {
	switch r {
	case model.DisputeResultWorkerWin:
		return "求职者胜"
	case model.DisputeResultEmployerWin:
		return "雇主胜"
	case model.DisputeResultCompromise:
		return "协商解决"
	case model.DisputeResultPlatformDecision:
		return "平台裁决"
	case model.DisputeResultReject:
		return "驳回"
	}
	return ""
}

// toDisputeInfo model -> dto
func toDisputeInfo(d *model.LinggongDispute) *dto.DisputeInfo {
	return &dto.DisputeInfo{
		ID:                 d.ID,
		DisputeNo:          d.DisputeNo,
		LinggongID:         d.LinggongID,
		TaskID:             d.TaskID,
		ApplicationID:      d.ApplicationID,
		ContractID:         d.ContractID,
		PaymentID:          d.PaymentID,
		DisputeType:        d.DisputeType,
		DisputeTypeText:    disputeTypeText(d.DisputeType),
		ApplicantType:      d.ApplicantType,
		ApplicantTypeText:  disputeApplicantTypeText(d.ApplicantType),
		ApplicantID:        d.ApplicantID,
		ApplicantName:      d.ApplicantName,
		RespondentID:       d.RespondentID,
		RespondentName:     d.RespondentName,
		Title:              d.Title,
		Description:        d.Description,
		EvidenceImages:     d.EvidenceImages,
		EvidenceVideos:     d.EvidenceVideos,
		EvidenceDocs:       d.EvidenceDocs,
		ClaimAmount:        d.ClaimAmount,
		Status:             d.Status,
		StatusText:         disputeStatusText(d.Status),
		HandlerID:          d.HandlerID,
		HandlerName:        d.HandlerName,
		MediationResult:    d.MediationResult,
		ArbitrationResult:  d.ArbitrationResult,
		FinalResult:        d.FinalResult,
		FinalResultText:    disputeFinalResultText(d.FinalResult),
		CompensationAmount: d.CompensationAmount,
		SLADeadline:        d.SLADeadline,
		HandledAt:          d.HandledAt,
		ResolvedAt:         d.ResolvedAt,
		ClosedAt:           d.ClosedAt,
		AppealReason:       d.AppealReason,
		AppealedAt:         d.AppealedAt,
		AppealResult:       d.AppealResult,
		AppealHandlerID:    d.AppealHandlerID,
		AppealHandledAt:    d.AppealHandledAt,
		RegionID:           d.RegionID,
		CreatedAt:          d.CreatedAt,
		UpdatedAt:          d.UpdatedAt,
	}
}

// genDisputeNo 生成纠纷单号：DG + yyyyMMddHHmmss + 6 位随机
func genDisputeNo() string {
	return fmt.Sprintf("DG%s%06d", time.Now().Format("20060102150405"), rand.Intn(1000000))
}

// ===== C 端 =====

// Create 创建纠纷
func (s *disputeService) Create(regionID uint, userID uint, req *dto.CreateDisputeRequest) (*dto.DisputeInfo, error) {
	d := &model.LinggongDispute{
		DisputeNo:     genDisputeNo(),
		LinggongID:    req.LinggongID,
		TaskID:        req.TaskID,
		ApplicationID: req.ApplicationID,
		ContractID:    req.ContractID,
		PaymentID:     req.PaymentID,
		DisputeType:   req.DisputeType,
		ApplicantType: req.ApplicantType,
		ApplicantID:   req.ApplicantID,
		ApplicantName: req.ApplicantName,
		RespondentID:  req.RespondentID,
		RespondentName: req.RespondentName,
		Title:         req.Title,
		Description:   req.Description,
		ClaimAmount:   req.ClaimAmount,
		Status:        model.DisputeStatusPending,
	}
	d.RegionID = regionID

	// 默认值兜底
	if d.DisputeType == "" {
		d.DisputeType = model.DisputeTypeOther
	}
	if d.ApplicantType == "" {
		d.ApplicantType = model.DisputeApplicantWorker
	}

	// JSONB 字段
	if req.EvidenceImages != nil {
		if jb, err := model.FromJSON(req.EvidenceImages); err == nil {
			d.EvidenceImages = jb
		}
	}
	if req.EvidenceVideos != nil {
		if jb, err := model.FromJSON(req.EvidenceVideos); err == nil {
			d.EvidenceVideos = jb
		}
	}
	if req.EvidenceDocs != nil {
		if jb, err := model.FromJSON(req.EvidenceDocs); err == nil {
			d.EvidenceDocs = jb
		}
	}

	// SLA 截止时间：默认 3 天
	sla := time.Now().Add(72 * time.Hour)
	d.SLADeadline = &sla

	_ = userID

	if err := s.repo.Create(d); err != nil {
		return nil, err
	}
	return toDisputeInfo(d), nil
}

// GetByID 获取纠纷详情
func (s *disputeService) GetByID(id uint) (*dto.DisputeInfo, error) {
	d, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrDisputeNotFound
		}
		return nil, err
	}
	return toDisputeInfo(d), nil
}

// GetByDisputeNo 按纠纷单号查询
func (s *disputeService) GetByDisputeNo(no string) (*dto.DisputeInfo, error) {
	d, err := s.repo.FindByDisputeNo(no)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrDisputeNotFound
		}
		return nil, err
	}
	return toDisputeInfo(d), nil
}

// List C 端列表查询
func (s *disputeService) List(regionID uint, req *dto.DisputeListRequest) (*utils.Pagination, []dto.DisputeInfo, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	opts := repository.DisputeListOptions{
		LinggongID:   req.LinggongID,
		DisputeType:  req.DisputeType,
		ApplicantID:  req.ApplicantID,
		RespondentID: req.RespondentID,
		Status:       req.Status,
		Keyword:      req.Keyword,
	}
	list, total, err := s.repo.List(regionID, pagination, opts)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	result := make([]dto.DisputeInfo, 0, len(list))
	for i := range list {
		result = append(result, *toDisputeInfo(&list[i]))
	}
	return pagination, result, nil
}

// ListByLinggong 按岗位反查
func (s *disputeService) ListByLinggong(linggongID uint, page, pageSize int) (*utils.Pagination, []dto.DisputeInfo, error) {
	pagination := utils.NewPagination(page, pageSize)
	list, total, err := s.repo.ListByLinggong(linggongID, pagination)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	result := make([]dto.DisputeInfo, 0, len(list))
	for i := range list {
		result = append(result, *toDisputeInfo(&list[i]))
	}
	return pagination, result, nil
}

// ListByApplicant 按申请人反查
func (s *disputeService) ListByApplicant(applicantID uint, page, pageSize int) (*utils.Pagination, []dto.DisputeInfo, error) {
	pagination := utils.NewPagination(page, pageSize)
	list, total, err := s.repo.ListByApplicant(applicantID, pagination)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	result := make([]dto.DisputeInfo, 0, len(list))
	for i := range list {
		result = append(result, *toDisputeInfo(&list[i]))
	}
	return pagination, result, nil
}

// ListByRespondent 按被申请人反查
func (s *disputeService) ListByRespondent(respondentID uint, page, pageSize int) (*utils.Pagination, []dto.DisputeInfo, error) {
	pagination := utils.NewPagination(page, pageSize)
	list, total, err := s.repo.ListByRespondent(respondentID, pagination)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	result := make([]dto.DisputeInfo, 0, len(list))
	for i := range list {
		result = append(result, *toDisputeInfo(&list[i]))
	}
	return pagination, result, nil
}

// Handle 处理纠纷（调解/仲裁/解决/驳回）
// 状态机：待处理 → 处理中 → 已调解/已仲裁 → 已解决 → 已关闭
func (s *disputeService) Handle(id uint, handlerID uint, handlerName string, req *dto.DisputeHandleRequest) error {
	d, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrDisputeNotFound
		}
		return err
	}

	// 状态机校验：已解决/驳回/取消状态不可再处理
	if d.Status == model.DisputeStatusResolved ||
		d.Status == model.DisputeStatusRejected ||
		d.Status == model.DisputeStatusCanceled {
		return ErrDisputeStatusInvalid
	}

	now := time.Now()
	fields := map[string]interface{}{
		"status":        req.Status,
		"handler_id":    handlerID,
		"handler_name":  handlerName,
		"handled_at":    &now,
	}

	switch req.Status {
	case model.DisputeStatusProcessing:
		// 进入处理中
	case model.DisputeStatusMediated:
		// 已调解
		fields["mediation_result"] = req.MediationResult
		fields["final_result"] = req.FinalResult
		fields["compensation_amount"] = req.CompensationAmount
		fields["resolved_at"] = &now
	case model.DisputeStatusArbitrated:
		// 已仲裁
		fields["arbitration_result"] = req.ArbitrationResult
		fields["final_result"] = req.FinalResult
		fields["compensation_amount"] = req.CompensationAmount
		fields["resolved_at"] = &now
	case model.DisputeStatusResolved:
		// 已解决（关闭）
		fields["final_result"] = req.FinalResult
		fields["compensation_amount"] = req.CompensationAmount
		fields["resolved_at"] = &now
		fields["closed_at"] = &now
	case model.DisputeStatusRejected:
		// 已驳回（关闭）
		fields["final_result"] = model.DisputeResultReject
		fields["closed_at"] = &now
	case model.DisputeStatusCanceled:
		// 已取消（关闭）
		fields["closed_at"] = &now
	case model.DisputeStatusAppealed:
		// 申诉中
		fields["appealed_at"] = &now
	}

	return s.repo.Update(id, fields)
}

// ===== M 端管理 =====

// AdminList M 端纠纷列表（跨地区）
func (s *disputeService) AdminList(req *dto.DisputeListRequest) (*utils.Pagination, []dto.DisputeInfo, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	opts := repository.DisputeListOptions{
		LinggongID:   req.LinggongID,
		DisputeType:  req.DisputeType,
		ApplicantID:  req.ApplicantID,
		RespondentID: req.RespondentID,
		Status:       req.Status,
		Keyword:      req.Keyword,
	}
	// 跨地区：regionID=0
	list, total, err := s.repo.List(0, pagination, opts)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	result := make([]dto.DisputeInfo, 0, len(list))
	for i := range list {
		result = append(result, *toDisputeInfo(&list[i]))
	}
	return pagination, result, nil
}
