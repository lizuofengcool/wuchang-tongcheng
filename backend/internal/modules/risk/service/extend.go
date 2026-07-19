// Package service 风控中台扩展业务逻辑层
// 依据 015_risk_full.sql：证据/申诉/规则/评分记录/审核日志
package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"wuchang-tongcheng/internal/modules/risk/dto"
	"wuchang-tongcheng/internal/modules/risk/model"
	"wuchang-tongcheng/internal/modules/risk/repository"

	"gorm.io/gorm"
)

// 扩展错误
// 注：ErrRuleNotFound 已在 risk.go 中声明，此处仅新增其他错误
var (
	ErrAppealNotFound      = errors.New("申诉不存在")
	ErrRuleAlreadyExists   = errors.New("规则名已存在")
	ErrEvidenceNotFound    = errors.New("证据不存在")
	ErrAppealAlreadyHandled = errors.New("申诉已处理")
)

// RiskExtendService 风控扩展业务接口
type RiskExtendService interface {
	// 证据
	AddEvidence(regionID, uploaderID uint, req *dto.AddEvidenceRequest) (*dto.EvidenceInfo, error)
	ListEvidenceByReport(reportID uint) ([]dto.EvidenceInfo, error)
	DeleteEvidence(id uint) error

	// 申诉
	CreateAppeal(regionID, userID uint, req *dto.CreateAppealRequest) (*dto.AppealInfo, error)
	ListMyAppeals(userID uint, page, pageSize int) ([]dto.AppealInfo, int64, error)
	ListAppeals(status, page, pageSize int) ([]dto.AppealInfo, int64, error)
	HandleAppeal(handlerID uint, req *dto.HandleAppealRequest) error

	// 风控规则
	CreateRule(regionID uint, req *dto.CreateRuleRequest) (*dto.RuleInfo, error)
	UpdateRule(id uint, req *dto.UpdateRuleRequest) error
	DeleteRule(id uint) error
	ListRules(ruleType string, page, pageSize int) ([]dto.RuleInfo, int64, error)
	GetRule(id uint) (*dto.RuleInfo, error)

	// 风险评分记录
	RecordUserScore(regionID, userID uint, score int, level, action string, reasons []string, ruleIDs []uint) error
	ListMyScoreRecords(userID uint, page, pageSize int) ([]dto.ScoreRecordInfo, int64, error)
	ListScoreRecordsByLevel(level string, page, pageSize int) ([]dto.ScoreRecordInfo, int64, error)

	// 审核日志
	CreateAuditLog(regionID, auditorID uint, auditorName, action, targetType string, targetID uint, bizModule, bizID, beforeStatus, afterStatus, remark, ip, userAgent string) error
	ListAuditLogs(auditorID uint, action, targetType string, page, pageSize int) ([]dto.AuditLogInfo, int64, error)

	// 统计
	Statistics() (*dto.RiskStatisticsResponse, error)
}

type riskExtendService struct {
	repo    repository.RiskRepository
	extRepo repository.RiskExtendRepository
}

// NewRiskExtendService 创建扩展 service 实例
func NewRiskExtendService(repo repository.RiskRepository, extRepo repository.RiskExtendRepository) RiskExtendService {
	return &riskExtendService{repo: repo, extRepo: extRepo}
}

// ===== 证据 =====

func (s *riskExtendService) AddEvidence(regionID, uploaderID uint, req *dto.AddEvidenceRequest) (*dto.EvidenceInfo, error) {
	e := &model.ReportEvidence{
		ReportID:     req.ReportID,
		EvidenceType: req.EvidenceType,
		URL:          req.URL,
		Description: req.Description,
		UploaderID:  uploaderID,
		FileSize:    req.FileSize,
		FileHash:    req.FileHash,
		Extra:       defaultJSONStr(req.Extra),
	}
	e.RegionID = regionID
	if err := s.extRepo.CreateEvidence(e); err != nil {
		return nil, err
	}
	return toEvidenceInfo(e), nil
}

func (s *riskExtendService) ListEvidenceByReport(reportID uint) ([]dto.EvidenceInfo, error) {
	list, err := s.extRepo.ListEvidenceByReport(reportID)
	if err != nil {
		return nil, err
	}
	result := make([]dto.EvidenceInfo, 0, len(list))
	for i := range list {
		result = append(result, *toEvidenceInfo(&list[i]))
	}
	return result, nil
}

func (s *riskExtendService) DeleteEvidence(id uint) error {
	return s.extRepo.DeleteEvidence(id)
}

// ===== 申诉 =====

func (s *riskExtendService) CreateAppeal(regionID, userID uint, req *dto.CreateAppealRequest) (*dto.AppealInfo, error) {
	v, err := s.repo.FindViolationByID(req.ViolationID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrRuleNotFound
		}
		return nil, err
	}
	if v.UserID != userID {
		return nil, errors.New("无权申诉他人违规")
	}
	imagesJSON := "[]"
	if len(req.EvidenceImages) > 0 {
		b, _ := json.Marshal(req.EvidenceImages)
		imagesJSON = string(b)
	}
	a := &model.Appeal{
		AppealNo:       generateAppealNo(),
		ViolationID:    req.ViolationID,
		UserID:         userID,
		Reason:         req.Reason,
		EvidenceImages: imagesJSON,
		Status:         model.AppealStatusPending,
	}
	a.RegionID = regionID
	if err := s.extRepo.CreateAppeal(a); err != nil {
		return nil, err
	}
	// 更新违规的申诉状态
	_ = s.repo.UpdateViolationFields(v.ID, map[string]interface{}{
		"appeal_status": model.AppealStatusProcessing,
	})
	return toAppealInfo(a), nil
}

func (s *riskExtendService) ListMyAppeals(userID uint, page, pageSize int) ([]dto.AppealInfo, int64, error) {
	list, total, err := s.extRepo.ListUserAppeals(userID, page, pageSize)
	if err != nil {
		return nil, 0, err
	}
	result := make([]dto.AppealInfo, 0, len(list))
	for i := range list {
		result = append(result, *toAppealInfo(&list[i]))
	}
	return result, total, nil
}

func (s *riskExtendService) ListAppeals(status, page, pageSize int) ([]dto.AppealInfo, int64, error) {
	list, total, err := s.extRepo.ListAppeals(status, page, pageSize)
	if err != nil {
		return nil, 0, err
	}
	result := make([]dto.AppealInfo, 0, len(list))
	for i := range list {
		result = append(result, *toAppealInfo(&list[i]))
	}
	return result, total, nil
}

func (s *riskExtendService) HandleAppeal(handlerID uint, req *dto.HandleAppealRequest) error {
	a, err := s.extRepo.FindAppealByID(req.AppealID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrAppealNotFound
		}
		return err
	}
	if a.Status != model.AppealStatusPending {
		return ErrAppealAlreadyHandled
	}
	now := time.Now()
	if err := s.extRepo.UpdateAppealFields(a.ID, map[string]interface{}{
		"status":        req.Status,
		"handler_id":    handlerID,
		"handle_remark": req.HandleRemark,
		"handled_at":    &now,
	}); err != nil {
		return err
	}
	// 申诉成功 → 撤销违规
	if req.Status == model.AppealStatusApproved {
		_ = s.repo.UpdateViolationFields(a.ViolationID, map[string]interface{}{
			"status":        model.ViolationStatusAppealed,
			"appeal_status": model.AppealStatusSuccess,
		})
	} else {
		_ = s.repo.UpdateViolationFields(a.ViolationID, map[string]interface{}{
			"appeal_status": model.AppealStatusFailed,
		})
	}
	return nil
}

// ===== 风控规则 =====

func (s *riskExtendService) CreateRule(regionID uint, req *dto.CreateRuleRequest) (*dto.RuleInfo, error) {
	if _, err := s.extRepo.FindRuleByName(req.RuleName); err == nil {
		return nil, ErrRuleAlreadyExists
	}
	priority := req.Priority
	if priority == 0 {
		priority = 100
	}
	status := req.Status
	if status == 0 {
		status = 1
	}
	r := &model.Rule{
		RuleName:    req.RuleName,
		RuleType:    req.RuleType,
		Description: req.Description,
		Config:      defaultJSONStr(req.Config),
		Action:      req.Action,
		Priority:    priority,
		Status:      status,
	}
	r.RegionID = regionID
	if err := s.extRepo.CreateRule(r); err != nil {
		return nil, err
	}
	return toRuleInfo(r), nil
}

func (s *riskExtendService) UpdateRule(id uint, req *dto.UpdateRuleRequest) error {
	if _, err := s.extRepo.FindRuleByID(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrRuleNotFound
		}
		return err
	}
	fields := map[string]interface{}{}
	if req.Description != "" {
		fields["description"] = req.Description
	}
	if req.Config != "" {
		fields["config"] = req.Config
	}
	if req.Action != "" {
		fields["action"] = req.Action
	}
	if req.Priority != 0 {
		fields["priority"] = req.Priority
	}
	if req.Status != 0 {
		fields["status"] = req.Status
	}
	if len(fields) == 0 {
		return nil
	}
	return s.extRepo.UpdateRuleFields(id, fields)
}

func (s *riskExtendService) DeleteRule(id uint) error {
	if _, err := s.extRepo.FindRuleByID(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrRuleNotFound
		}
		return err
	}
	return s.extRepo.DeleteRule(id)
}

func (s *riskExtendService) ListRules(ruleType string, page, pageSize int) ([]dto.RuleInfo, int64, error) {
	list, total, err := s.extRepo.ListRules(ruleType, page, pageSize)
	if err != nil {
		return nil, 0, err
	}
	result := make([]dto.RuleInfo, 0, len(list))
	for i := range list {
		result = append(result, *toRuleInfo(&list[i]))
	}
	return result, total, nil
}

func (s *riskExtendService) GetRule(id uint) (*dto.RuleInfo, error) {
	r, err := s.extRepo.FindRuleByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrRuleNotFound
		}
		return nil, err
	}
	return toRuleInfo(r), nil
}

// ===== 风险评分记录 =====

func (s *riskExtendService) RecordUserScore(regionID, userID uint, score int, level, action string, reasons []string, ruleIDs []uint) error {
	reasonsJSON := "[]"
	if len(reasons) > 0 {
		b, _ := json.Marshal(reasons)
		reasonsJSON = string(b)
	}
	ruleIDsJSON := "[]"
	if len(ruleIDs) > 0 {
		b, _ := json.Marshal(ruleIDs)
		ruleIDsJSON = string(b)
	}
	rec := &model.ScoreRecord{
		UserID:      userID,
		TargetType:  model.ScoreTargetUser,
		TargetValue: fmt.Sprintf("%d", userID),
		Score:       score,
		Level:       level,
		Reasons:     reasonsJSON,
		RuleIDs:     ruleIDsJSON,
		ActionTaken: action,
	}
	rec.RegionID = regionID
	if err := s.extRepo.CreateScoreRecord(rec); err != nil {
		return err
	}
	// 同步更新用户风险分
	us, err := s.repo.GetOrCreateUserScore(userID, regionID)
	if err == nil {
		newScore := us.Score + score
		if newScore < 0 {
			newScore = 0
		}
		if newScore > 100 {
			newScore = 100
		}
		newLevel := model.RiskLevelSafe
		if newScore >= 70 {
			newLevel = model.RiskLevelDanger
		} else if newScore >= 40 {
			newLevel = model.RiskLevelWarning
		}
		_ = s.repo.UpdateUserScoreFields(us.ID, map[string]interface{}{
			"score": newScore,
			"level": newLevel,
		})
	}
	return nil
}

func (s *riskExtendService) ListMyScoreRecords(userID uint, page, pageSize int) ([]dto.ScoreRecordInfo, int64, error) {
	list, total, err := s.extRepo.ListScoreRecordsByUser(userID, page, pageSize)
	if err != nil {
		return nil, 0, err
	}
	result := make([]dto.ScoreRecordInfo, 0, len(list))
	for i := range list {
		result = append(result, *toScoreRecordInfo(&list[i]))
	}
	return result, total, nil
}

func (s *riskExtendService) ListScoreRecordsByLevel(level string, page, pageSize int) ([]dto.ScoreRecordInfo, int64, error) {
	list, total, err := s.extRepo.ListScoreRecordsByLevel(level, page, pageSize)
	if err != nil {
		return nil, 0, err
	}
	result := make([]dto.ScoreRecordInfo, 0, len(list))
	for i := range list {
		result = append(result, *toScoreRecordInfo(&list[i]))
	}
	return result, total, nil
}

// ===== 审核日志 =====

func (s *riskExtendService) CreateAuditLog(regionID, auditorID uint, auditorName, action, targetType string, targetID uint, bizModule, bizID, beforeStatus, afterStatus, remark, ip, userAgent string) error {
	l := &model.AuditLog{
		AuditorID:    auditorID,
		AuditorName:  auditorName,
		Action:       action,
		TargetType:   targetType,
		TargetID:     targetID,
		BizModule:    bizModule,
		BizID:        bizID,
		BeforeStatus: beforeStatus,
		AfterStatus:  afterStatus,
		Remark:       remark,
		IP:           ip,
		UserAgent:    userAgent,
	}
	l.RegionID = regionID
	return s.extRepo.CreateAuditLog(l)
}

func (s *riskExtendService) ListAuditLogs(auditorID uint, action, targetType string, page, pageSize int) ([]dto.AuditLogInfo, int64, error) {
	list, total, err := s.extRepo.ListAuditLogs(auditorID, action, targetType, page, pageSize)
	if err != nil {
		return nil, 0, err
	}
	result := make([]dto.AuditLogInfo, 0, len(list))
	for i := range list {
		result = append(result, *toAuditLogInfo(&list[i]))
	}
	return result, total, nil
}

// ===== 统计 =====

func (s *riskExtendService) Statistics() (*dto.RiskStatisticsResponse, error) {
	resp := &dto.RiskStatisticsResponse{}
	resp.TotalReports, _ = s.extRepo.StatTotalReports()
	resp.PendingReports, _ = s.extRepo.StatPendingReports()
	resp.HandledReports, _ = s.extRepo.StatHandledReports()
	resp.TotalAppeals, _ = s.extRepo.StatTotalAppeals()
	resp.PendingAppeals, _ = s.extRepo.StatPendingAppeals()
	resp.TotalViolations, _ = s.extRepo.StatTotalViolations()
	resp.ActiveViolations, _ = s.extRepo.StatActiveViolations()
	resp.BlacklistCount, _ = s.extRepo.StatBlacklistCount()
	resp.SensitiveWords, _ = s.extRepo.StatSensitiveWords()
	resp.RulesCount, _ = s.extRepo.StatRulesCount()
	resp.AuditLogsCount, _ = s.extRepo.StatAuditLogsCount()
	return resp, nil
}

// ===== 工具函数 =====

func generateAppealNo() string {
	return fmt.Sprintf("AP%s%08d", time.Now().Format("20060102150405"), time.Now().UnixNano()%100000000)
}

func defaultJSONStr(s string) string {
	if s == "" {
		return "{}"
	}
	return s
}

func toEvidenceInfo(e *model.ReportEvidence) *dto.EvidenceInfo {
	return &dto.EvidenceInfo{
		ID:           e.ID,
		ReportID:     e.ReportID,
		EvidenceType: e.EvidenceType,
		URL:          e.URL,
		Description:  e.Description,
		UploaderID:   e.UploaderID,
		FileSize:     e.FileSize,
		FileHash:     e.FileHash,
		Extra:        e.Extra,
		CreatedAt:    e.CreatedAt,
	}
}

func toAppealInfo(a *model.Appeal) *dto.AppealInfo {
	return &dto.AppealInfo{
		ID:             a.ID,
		AppealNo:       a.AppealNo,
		ViolationID:    a.ViolationID,
		UserID:         a.UserID,
		Reason:         a.Reason,
		EvidenceImages: a.EvidenceImages,
		Status:         a.Status,
		HandlerID:      a.HandlerID,
		HandleRemark:   a.HandleRemark,
		HandledAt:      a.HandledAt,
		CreatedAt:      a.CreatedAt,
	}
}

func toRuleInfo(r *model.Rule) *dto.RuleInfo {
	return &dto.RuleInfo{
		ID:          r.ID,
		RuleName:    r.RuleName,
		RuleType:    r.RuleType,
		Description: r.Description,
		Config:      r.Config,
		Action:      r.Action,
		Priority:    r.Priority,
		Status:      r.Status,
		HitCount:    r.HitCount,
		LastHitAt:   r.LastHitAt,
		CreatedAt:   r.CreatedAt,
	}
}

func toScoreRecordInfo(r *model.ScoreRecord) *dto.ScoreRecordInfo {
	return &dto.ScoreRecordInfo{
		ID:          r.ID,
		UserID:      r.UserID,
		TargetType:  r.TargetType,
		TargetValue: r.TargetValue,
		ContentType: r.ContentType,
		ContentID:   r.ContentID,
		Score:       r.Score,
		Level:       r.Level,
		Reasons:     r.Reasons,
		RuleIDs:     r.RuleIDs,
		ActionTaken: r.ActionTaken,
		CreatedAt:   r.CreatedAt,
	}
}

func toAuditLogInfo(l *model.AuditLog) *dto.AuditLogInfo {
	return &dto.AuditLogInfo{
		ID:           l.ID,
		AuditorID:    l.AuditorID,
		AuditorName:  l.AuditorName,
		Action:       l.Action,
		TargetType:   l.TargetType,
		TargetID:     l.TargetID,
		BizModule:    l.BizModule,
		BizID:        l.BizID,
		BeforeStatus: l.BeforeStatus,
		AfterStatus:  l.AfterStatus,
		Remark:       l.Remark,
		IP:           l.IP,
		UserAgent:    l.UserAgent,
		CreatedAt:    l.CreatedAt,
	}
}
