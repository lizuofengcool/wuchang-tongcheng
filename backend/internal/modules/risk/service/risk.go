// Package service 风控审核中台精简版业务逻辑层
// 依据 ershou 模块依赖：举报 + 敏感词 DFA + 内容审核 + 黑名单 + 用户风险分 + 违规处罚
// 暴露 RiskService 接口供其他模块直接 import 调用（不通过 HTTP）
package service

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"wuchang-tongcheng/internal/modules/risk/dto"
	"wuchang-tongcheng/internal/modules/risk/model"
	"wuchang-tongcheng/internal/modules/risk/repository"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

var (
	ErrReportNotFound      = errors.New("举报不存在")
	ErrWordNotFound        = errors.New("敏感词不存在")
	ErrRuleNotFound        = errors.New("审核规则不存在")
	ErrBlacklistNotFound   = errors.New("黑名单记录不存在")
	ErrScoreNotFound       = errors.New("用户风险分不存在")
	ErrViolationNotFound   = errors.New("违规处罚不存在")
	ErrAlreadyInBlacklist  = errors.New("已在黑名单中")
	ErrUserBanned          = errors.New("用户已被封禁")
	ErrContentRejected     = errors.New("内容审核未通过")
)

// RiskService 风控中台业务接口
// 暴露给其他模块直接 import 调用，不通过 HTTP
type RiskService interface {
	// 举报
	CreateReport(regionID, reporterID uint, req *dto.ReportRequest) (*dto.ReportInfo, error)
	GetReport(reportID uint) (*dto.ReportInfo, error)
	ListReports(req *dto.ReportListRequest) ([]dto.ReportInfo, int64, error)
	HandleReport(handlerID uint, req *dto.HandleReportRequest) error

	// 敏感词管理
	AddSensitiveWord(req *dto.SensitiveWordRequest) error
	DeleteSensitiveWord(id uint) error
	ListSensitiveWords(wordType string, page, pageSize int) ([]model.SensitiveWord, int64, error)
	ReloadSensitiveWords() error // 重新加载 DFA

	// 文本审核
	CheckText(req *dto.CheckTextRequest) (*dto.CheckTextResponse, error)

	// 内容审核（综合：敏感词 + 黑名单 + 风险分）
	AuditContent(regionID uint, req *dto.AuditResultRequest) (*dto.AuditResultResponse, error)

	// 黑名单
	AddToBlacklist(operatorID uint, req *dto.BlacklistRequest) error
	CheckBlacklist(targetType, targetValue string) (bool, error)
	ListBlacklist(targetType string, page, pageSize int) ([]model.Blacklist, int64, error)

	// 用户风险分
	GetUserScore(userID uint) (*dto.UserScoreInfo, error)
	DeductUserScore(userID, regionID uint, points int, reason string) error

	// 违规处罚
	CreateViolation(regionID uint, userID uint, violationType, level, reason, bizModule, bizID string, reportID uint, penaltyMinutes int) (*dto.ViolationInfo, error)
	ListUserViolations(userID uint, page, pageSize int) ([]dto.ViolationInfo, int64, error)
	AppealViolation(userID uint, req *dto.AppealRequest) error
}

type riskService struct {
	repo repository.RiskRepository

	// DFA 敏感词树
	mu      sync.RWMutex
	dfaRoot *dfaNode
	loaded  bool
}

// dfaNode DFA 节点
type dfaNode struct {
	children map[rune]*dfaNode
	isEnd    bool
	word     string // 完整敏感词（isEnd=true 时有效）
}

func newDFANode() *dfaNode {
	return &dfaNode{children: make(map[rune]*dfaNode)}
}

// NewRiskService 创建 service 实例
func NewRiskService(repo repository.RiskRepository) RiskService {
	s := &riskService{repo: repo}
	// 启动时尝试加载敏感词，失败不阻塞
	_ = s.ReloadSensitiveWords()
	return s
}

// ===== 举报 =====

// CreateReport 创建举报
func (s *riskService) CreateReport(regionID, reporterID uint, req *dto.ReportRequest) (*dto.ReportInfo, error) {
	reportNo := generateReportNo()
	evidenceJSON := "[]"
	if len(req.EvidenceImages) > 0 {
		evidenceJSON = strings.Join(req.EvidenceImages, "\",\"")
		evidenceJSON = fmt.Sprintf(`["%s"]`, evidenceJSON)
	}

	// SLA 24 小时
	sla := time.Now().Add(24 * time.Hour)

	rep := &model.Report{
		ReportNo:          reportNo,
		ReporterID:        reporterID,
		ReportedUserID:    req.ReportedUserID,
		ReportedBizModule: req.BizModule,
		ReportedBizID:     req.BizID,
		ReportType:        req.ReportType,
		Reason:            req.Reason,
		EvidenceImages:    evidenceJSON,
		Status:            model.ReportStatusPending,
		SLADeadline:       &sla,
	}
	rep.RegionID = regionID

	if err := s.repo.CreateReport(rep); err != nil {
		return nil, err
	}

	// 增加被举报人的被举报次数
	if req.ReportedUserID > 0 {
		_, _ = s.repo.GetOrCreateUserScore(req.ReportedUserID, regionID)
		// 直接 SQL 风格递增
		_ = s.repo.UpdateUserScoreFields(req.ReportedUserID, map[string]interface{}{
			"report_count": gorm.Expr("report_count + 1"),
		})
	}

	return toReportInfo(rep), nil
}

// GetReport 查询举报
func (s *riskService) GetReport(reportID uint) (*dto.ReportInfo, error) {
	rep, err := s.repo.FindReportByID(reportID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrReportNotFound
		}
		return nil, err
	}
	return toReportInfo(rep), nil
}

// ListReports 举报列表
func (s *riskService) ListReports(req *dto.ReportListRequest) ([]dto.ReportInfo, int64, error) {
	q := &repository.ListReportsQuery{
		Status:     req.Status,
		ReportType: req.ReportType,
		BizModule:  req.BizModule,
		ReporterID: req.ReporterID,
		Page:       req.Page,
		PageSize:   req.PageSize,
	}
	list, total, err := s.repo.ListReports(q)
	if err != nil {
		return nil, 0, err
	}
	result := make([]dto.ReportInfo, 0, len(list))
	for i := range list {
		result = append(result, *toReportInfo(&list[i]))
	}
	return result, total, nil
}

// HandleReport 处理举报
func (s *riskService) HandleReport(handlerID uint, req *dto.HandleReportRequest) error {
	rep, err := s.repo.FindReportByID(req.ReportID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrReportNotFound
		}
		return err
	}

	now := time.Now()
	fields := map[string]interface{}{
		"status":        model.ReportStatusHandled,
		"handle_result": req.HandleResult,
		"handle_remark": req.HandleRemark,
		"handler_id":    handlerID,
		"handled_at":    now,
	}
	if err := s.repo.UpdateReportFields(rep.ID, fields); err != nil {
		return err
	}

	// 处罚被举报人
	if req.NeedPenalty && rep.ReportedUserID > 0 && req.PenaltyLevel != "" {
		violationType := mapHandleResultToViolationType(req.HandleResult)
		_, err := s.CreateViolation(
			rep.RegionID,
			rep.ReportedUserID,
			violationType,
			req.PenaltyLevel,
			req.HandleRemark,
			rep.ReportedBizModule,
			rep.ReportedBizID,
			rep.ID,
			req.PenaltyMinutes,
		)
		if err != nil {
			return err
		}
	}

	return nil
}

// ===== 敏感词管理 =====

// AddSensitiveWord 添加敏感词
func (s *riskService) AddSensitiveWord(req *dto.SensitiveWordRequest) error {
	word := strings.TrimSpace(req.Word)
	if word == "" {
		return ErrWordNotFound
	}

	// 幂等：已存在则跳过
	if existing, err := s.repo.FindSensitiveWordByWord(word); err == nil && existing != nil {
		return nil
	}

	replacement := req.Replacement
	if replacement == "" {
		replacement = "***"
	}

	w := &model.SensitiveWord{
		Word:        word,
		WordType:    req.WordType,
		Category:    req.Category,
		Replacement: replacement,
		Status:      1,
	}
	if err := s.repo.CreateSensitiveWord(w); err != nil {
		return err
	}

	// 增量加入 DFA
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.dfaRoot == nil {
		s.dfaRoot = newDFANode()
	}
	insertWordToDFA(s.dfaRoot, word)
	return nil
}

// DeleteSensitiveWord 删除敏感词
func (s *riskService) DeleteSensitiveWord(id uint) error {
	if err := s.repo.DeleteSensitiveWord(id); err != nil {
		return err
	}
	// 删除后重新加载 DFA
	return s.ReloadSensitiveWords()
}

// ListSensitiveWords 列表
func (s *riskService) ListSensitiveWords(wordType string, page, pageSize int) ([]model.SensitiveWord, int64, error) {
	return s.repo.ListSensitiveWords(wordType, page, pageSize)
}

// ReloadSensitiveWords 重新加载敏感词 DFA
func (s *riskService) ReloadSensitiveWords() error {
	words, err := s.repo.ListAllActiveSensitiveWords()
	if err != nil {
		return err
	}

	root := newDFANode()
	for _, w := range words {
		insertWordToDFA(root, w.Word)
	}

	s.mu.Lock()
	s.dfaRoot = root
	s.loaded = true
	s.mu.Unlock()
	return nil
}

// ===== 文本审核 =====

// CheckText 敏感词检测
func (s *riskService) CheckText(req *dto.CheckTextRequest) (*dto.CheckTextResponse, error) {
	replacement := req.Replacement
	if replacement == "" {
		replacement = "***"
	}

	s.mu.RLock()
	root := s.dfaRoot
	s.mu.RUnlock()

	if root == nil {
		// 未加载，尝试加载
		_ = s.ReloadSensitiveWords()
		s.mu.RLock()
		root = s.dfaRoot
		s.mu.RUnlock()
	}

	if root == nil || len(root.children) == 0 {
		return &dto.CheckTextResponse{
			Passed:      true,
			HitWords:    []string{},
			CleanedText: req.Text,
			HitCount:    0,
		}, nil
	}

	hitWords, cleanedText := searchDFA(root, req.Text, replacement)
	hitCount := len(hitWords)

	return &dto.CheckTextResponse{
		Passed:      hitCount == 0,
		HitWords:    hitWords,
		CleanedText: cleanedText,
		HitCount:    hitCount,
	}, nil
}

// ===== 综合内容审核 =====

// AuditContent 综合内容审核
// 1. 黑名单检查
// 2. 敏感词检查
// 3. 用户风险分检查
func (s *riskService) AuditContent(regionID uint, req *dto.AuditResultRequest) (*dto.AuditResultResponse, error) {
	resp := &dto.AuditResultResponse{
		Passed:    true,
		HitWords:  []string{},
		Reasons:   []string{},
		RiskLevel: model.RiskLevelSafe,
	}

	// 1. 黑名单检查（user 类型）
	if req.UserID > 0 {
		inBlack, err := s.CheckBlacklist(model.BlacklistTargetUser, fmt.Sprintf("%d", req.UserID))
		if err == nil && inBlack {
			resp.Passed = false
			resp.RiskLevel = model.RiskLevelDanger
			resp.Reasons = append(resp.Reasons, "用户已在黑名单中")
			resp.Suggestion = "禁止发布"
			return resp, nil
		}
	}

	// 2. 敏感词检查
	text := req.Title + " " + req.Content
	if strings.TrimSpace(text) != "" {
		checkResp, err := s.CheckText(&dto.CheckTextRequest{Text: text})
		if err == nil && !checkResp.Passed {
			resp.Passed = false
			resp.HitWords = checkResp.HitWords
			resp.RiskLevel = model.RiskLevelWarning
			resp.Reasons = append(resp.Reasons, fmt.Sprintf("命中敏感词 %d 个", checkResp.HitCount))
			resp.Suggestion = "请修改内容后重新提交"
		}
	}

	// 3. 用户风险分检查
	if req.UserID > 0 {
		score, err := s.GetUserScore(req.UserID)
		if err == nil {
			if score.Score < 30 {
				resp.Passed = false
				resp.RiskLevel = model.RiskLevelDanger
				resp.Reasons = append(resp.Reasons, fmt.Sprintf("用户风险分过低（%d）", score.Score))
				resp.Suggestion = "限制发布"
			} else if score.Score < 60 && resp.RiskLevel == model.RiskLevelSafe {
				resp.RiskLevel = model.RiskLevelWarning
			}
		}
	}

	return resp, nil
}

// ===== 黑名单 =====

// AddToBlacklist 加入黑名单
func (s *riskService) AddToBlacklist(operatorID uint, req *dto.BlacklistRequest) error {
	// 幂等：已在黑名单则跳过
	if existing, err := s.repo.FindActiveBlacklist(req.TargetType, req.TargetValue); err == nil && existing != nil {
		return nil
	}

	b := &model.Blacklist{
		TargetType:  req.TargetType,
		TargetValue: req.TargetValue,
		Reason:      req.Reason,
		OperatorID:  operatorID,
		ExpireAt:    req.ExpireAt,
		Status:      1,
	}
	return s.repo.CreateBlacklist(b)
}

// CheckBlacklist 检查是否在黑名单
func (s *riskService) CheckBlacklist(targetType, targetValue string) (bool, error) {
	b, err := s.repo.FindActiveBlacklist(targetType, targetValue)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}
	return b != nil, nil
}

// ListBlacklist 黑名单列表
func (s *riskService) ListBlacklist(targetType string, page, pageSize int) ([]model.Blacklist, int64, error) {
	return s.repo.ListBlacklist(targetType, page, pageSize)
}

// ===== 用户风险分 =====

// GetUserScore 查询用户风险分
func (s *riskService) GetUserScore(userID uint) (*dto.UserScoreInfo, error) {
	us, err := s.repo.FindUserScore(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrScoreNotFound
		}
		return nil, err
	}
	return &dto.UserScoreInfo{
		UserID:          us.UserID,
		Score:           us.Score,
		Level:           us.Level,
		ViolationCount:  us.ViolationCount,
		ReportCount:     us.ReportCount,
		LastViolationAt: us.LastViolationAt,
	}, nil
}

// DeductUserScore 扣减用户风险分
func (s *riskService) DeductUserScore(userID, regionID uint, points int, reason string) error {
	us, err := s.repo.GetOrCreateUserScore(userID, regionID)
	if err != nil {
		return err
	}

	newScore := us.Score - points
	if newScore < 0 {
		newScore = 0
	}

	level := model.RiskLevelSafe
	switch {
	case newScore < 30:
		level = model.RiskLevelDanger
	case newScore < 60:
		level = model.RiskLevelWarning
	}

	now := time.Now()
	return s.repo.UpdateUserScoreFields(us.ID, map[string]interface{}{
		"score":             newScore,
		"level":             level,
		"last_violation_at": now,
	})
}

// ===== 违规处罚 =====

// CreateViolation 创建违规处罚
func (s *riskService) CreateViolation(regionID uint, userID uint, violationType, level, reason, bizModule, bizID string, reportID uint, penaltyMinutes int) (*dto.ViolationInfo, error) {
	now := time.Now()
	var penaltyEnd *time.Time
	if level != model.PenaltyLevelWarning && level != model.PenaltyLevelLimit {
		if level == model.PenaltyLevelBanForever {
			// 永久封禁不设结束时间
		} else {
			minutes := penaltyMinutes
			if minutes <= 0 {
				switch level {
				case model.PenaltyLevelMute:
					minutes = 60
				case model.PenaltyLevelBan1d:
					minutes = 1440
				case model.PenaltyLevelBan7d:
					minutes = 10080
				}
			}
			end := now.Add(time.Duration(minutes) * time.Minute)
			penaltyEnd = &end
		}
	}

	v := &model.Violation{
		UserID:         userID,
		ViolationType:  violationType,
		Level:          level,
		Reason:         reason,
		BizModule:      bizModule,
		BizID:          bizID,
		ReportID:       reportID,
		PenaltyStart:   &now,
		PenaltyEnd:     penaltyEnd,
		Status:         model.ViolationStatusActive,
		AppealStatus:   model.AppealStatusNone,
	}
	v.RegionID = regionID

	if err := s.repo.CreateViolation(v); err != nil {
		return nil, err
	}

	// 同步扣分 + 累计违规次数
	_ = s.DeductUserScore(userID, regionID, penaltyToScoreDeduction(level), reason)
	if score, err := s.repo.GetOrCreateUserScore(userID, regionID); err == nil {
		_ = s.repo.UpdateUserScoreFields(score.ID, map[string]interface{}{
			"violation_count": gorm.Expr("violation_count + 1"),
		})
	}

	return toViolationInfo(v), nil
}

// ListUserViolations 用户违规列表
func (s *riskService) ListUserViolations(userID uint, page, pageSize int) ([]dto.ViolationInfo, int64, error) {
	list, total, err := s.repo.ListUserViolations(userID, page, pageSize)
	if err != nil {
		return nil, 0, err
	}
	result := make([]dto.ViolationInfo, 0, len(list))
	for i := range list {
		result = append(result, *toViolationInfo(&list[i]))
	}
	return result, total, nil
}

// AppealViolation 申诉
func (s *riskService) AppealViolation(userID uint, req *dto.AppealRequest) error {
	v, err := s.repo.FindViolationByID(req.ViolationID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrViolationNotFound
		}
		return err
	}
	if v.UserID != userID {
		return errors.New("无权申诉他人违规")
	}
	return s.repo.UpdateViolationFields(v.ID, map[string]interface{}{
		"appeal_status": model.AppealStatusProcessing,
		"appeal_remark": req.Remark,
	})
}

// ===== 工具函数 =====

// generateReportNo 生成举报单号
func generateReportNo() string {
	return fmt.Sprintf("R%s%s", time.Now().Format("20060102150405"), utils.RandomNumber(6))
}

// mapHandleResultToViolationType 处理结果转违规类型
func mapHandleResultToViolationType(result string) string {
	switch result {
	case model.HandleResultBan:
		return model.ViolationTypeSpam
	default:
		return model.ViolationTypeSpam
	}
}

// penaltyToScoreDeduction 处罚级别对应扣分
func penaltyToScoreDeduction(level string) int {
	switch level {
	case model.PenaltyLevelWarning:
		return 5
	case model.PenaltyLevelLimit:
		return 10
	case model.PenaltyLevelMute:
		return 20
	case model.PenaltyLevelBan1d:
		return 30
	case model.PenaltyLevelBan7d:
		return 50
	case model.PenaltyLevelBanForever:
		return 100
	}
	return 5
}

// toReportInfo model → dto
func toReportInfo(r *model.Report) *dto.ReportInfo {
	return &dto.ReportInfo{
		ID:                r.ID,
		ReportNo:          r.ReportNo,
		ReporterID:        r.ReporterID,
		ReportedUserID:    r.ReportedUserID,
		ReportedBizModule: r.ReportedBizModule,
		ReportedBizID:     r.ReportedBizID,
		ReportType:        r.ReportType,
		Reason:            r.Reason,
		EvidenceImages:    r.EvidenceImages,
		Status:            r.Status,
		HandleResult:      r.HandleResult,
		HandleRemark:      r.HandleRemark,
		HandlerID:         r.HandlerID,
		HandledAt:         r.HandledAt,
		SLADeadline:       r.SLADeadline,
		CreatedAt:         r.CreatedAt,
	}
}

// toViolationInfo model → dto
func toViolationInfo(v *model.Violation) *dto.ViolationInfo {
	return &dto.ViolationInfo{
		ID:            v.ID,
		UserID:        v.UserID,
		ViolationType: v.ViolationType,
		Level:         v.Level,
		Reason:        v.Reason,
		BizModule:     v.BizModule,
		BizID:         v.BizID,
		ReportID:      v.ReportID,
		PenaltyStart:  v.PenaltyStart,
		PenaltyEnd:    v.PenaltyEnd,
		Status:        v.Status,
		AppealStatus:  v.AppealStatus,
		CreatedAt:     v.CreatedAt,
	}
}

// ===== DFA 算法 =====

// insertWordToDFA 将词插入 DFA
func insertWordToDFA(root *dfaNode, word string) {
	if root == nil || word == "" {
		return
	}
	node := root
	for _, ch := range word {
		child, ok := node.children[ch]
		if !ok {
			child = newDFANode()
			node.children[ch] = child
		}
		node = child
	}
	node.isEnd = true
	node.word = word
}

// searchDFA 在文本中搜索敏感词并替换
func searchDFA(root *dfaNode, text string, replacement string) ([]string, string) {
	if root == nil || text == "" {
		return nil, text
	}
	runes := []rune(text)
	n := len(runes)
	hitSet := make(map[string]struct{})
	var hitWords []string

	result := make([]rune, 0, n)
	i := 0
	for i < n {
		node := root
		j := i
		lastHitEnd := -1
		lastHitWord := ""
		for j < n {
			child, ok := node.children[runes[j]]
			if !ok {
				break
			}
			node = child
			if node.isEnd {
				lastHitEnd = j
				lastHitWord = node.word
			}
			j++
		}
		if lastHitEnd >= 0 {
			// 命中：i..lastHitEnd 替换为 replacement
			if _, exists := hitSet[lastHitWord]; !exists {
				hitSet[lastHitWord] = struct{}{}
				hitWords = append(hitWords, lastHitWord)
			}
			result = append(result, []rune(replacement)...)
			i = lastHitEnd + 1
		} else {
			result = append(result, runes[i])
			i++
		}
	}
	return hitWords, string(result)
}
