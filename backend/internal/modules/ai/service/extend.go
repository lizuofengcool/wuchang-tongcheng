// Package service AI 中台扩展业务逻辑层
// 依据 016_ai_full.sql：审核/推荐/对话/模型配置/训练数据
package service

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"wuchang-tongcheng/internal/modules/ai/dto"
	"wuchang-tongcheng/internal/modules/ai/model"
	"wuchang-tongcheng/internal/modules/ai/repository"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

// 扩展错误（与 ai.go 中已存在的错误变量不重复）
var (
	ErrAuditResultNotFound    = errors.New("审核结果不存在")
	ErrRecommendationNotFound = errors.New("推荐记录不存在")
	ErrChatSessionNotFound    = errors.New("对话会话不存在")
	ErrChatMessageNotFound    = errors.New("对话消息不存在")
	ErrTrainingDataNotFound   = errors.New("训练数据不存在")
	ErrModelConfigNotFound    = errors.New("模型配置不存在")
	ErrChatSessionClosed      = errors.New("对话会话已关闭")
)

// 简单敏感词表（精简版本地审核）
var sensitiveWords = []string{
	"赌博", "色情", "毒品", "枪支", "弹药", "爆炸物",
	"诈骗", "传销", "非法集资", "违禁药品",
	"反动", "邪教",
}

// AIExtendService AI 扩展业务接口
type AIExtendService interface {
	// 审核
	SubmitAudit(content string, contentType string, bizModule string, bizID uint) (*dto.AuditResultInfo, error)
	GetAuditResult(taskID string) (*dto.AuditResultInfo, error)
	ListAuditResults(level string, page, pageSize int) ([]dto.AuditResultInfo, int64, error)
	ReviewAudit(taskID string, passed bool, reviewer string, comment string) error

	// 推荐
	GetRecommendations(userID uint, recType string, limit int) ([]dto.RecommendationInfo, error)
	TrackClick(recID uint, req dto.TrackClickRequest) error
	TrackDwell(recID uint, req dto.TrackDwellRequest) error
	FeedbackRecommendation(recID uint, req dto.FeedbackRecommendationRequest) error

	// 模型配置
	ListModelConfigs(modelType string, page, pageSize int) ([]dto.ModelConfigInfo, int64, error)
	UpsertModelConfig(req dto.UpsertModelConfigRequest) (uint, error)
	DeleteModelConfig(id uint) error

	// 对话
	CreateChatSession(userID uint, req dto.CreateChatSessionRequest) (*dto.ChatSessionInfo, error)
	ListChatSessions(userID uint, page, pageSize int) ([]dto.ChatSessionInfo, int64, error)
	Chat(userID uint, req dto.ChatRequest) (*dto.ChatResponse, error)
	ListChatMessages(sessionID string, page, pageSize int) ([]dto.ChatMessageInfo, int64, error)
	MessageFeedback(messageID uint, req dto.MessageFeedbackRequest) error

	// 训练数据
	ListTrainingData(dataType string, page, pageSize int) ([]dto.TrainingDataInfo, int64, error)
	CreateTrainingData(req dto.CreateTrainingDataRequest) (uint, error)
	MarkTrainingDataUsed(id uint) error

	// 统计
	GetStatistics() (*dto.AIStatisticsResponse, error)
}

type aiExtendService struct {
	repo repository.AIExtendRepository
}

// NewAIExtendService 创建扩展 service 实例
func NewAIExtendService(repo repository.AIExtendRepository) AIExtendService {
	return &aiExtendService{repo: repo}
}

// ===== 审核 =====

// SubmitAudit 提交审核（精简版：本地规则匹配）
func (s *aiExtendService) SubmitAudit(content string, contentType string, bizModule string, bizID uint) (*dto.AuditResultInfo, error) {
	if content == "" {
		return nil, errors.New("审核内容不能为空")
	}
	if bizModule == "" {
		return nil, errors.New("业务模块不能为空")
	}
	if bizID == 0 {
		return nil, errors.New("业务ID不能为空")
	}

	start := time.Now()
	taskID := generateTaskID()

	algorithm := model.AlgoLocal
	riskScore, riskLevel, suggestion, hitRules := s.localAudit(content)

	passed := 1
	if riskLevel == "block" {
		passed = 0
	}

	a := &model.AuditResult{
		TaskID:      taskID,
		BizModule:   bizModule,
		BizID:       fmt.Sprintf("%d", bizID),
		ContentType: contentType,
		ContentHash: utils.MD5(content),
		Algorithm:   algorithm,
		RiskScore:   riskScore,
		RiskLevel:   riskLevel,
		Labels:      "[]",
		HitRules:    hitRules,
		Suggestion:  suggestion,
		Passed:      passed,
		CostMs:      int(time.Since(start).Milliseconds()),
	}
	if err := s.repo.CreateAuditResult(a); err != nil {
		return nil, err
	}
	return toAuditResultInfo(a), nil
}

// localAudit 本地规则审核
// 返回：riskScore, riskLevel, suggestion, hitRules(json)
func (s *aiExtendService) localAudit(content string) (float64, string, string, string) {
	lower := strings.ToLower(content)
	var hits []string
	for _, w := range sensitiveWords {
		if strings.Contains(lower, strings.ToLower(w)) {
			hits = append(hits, w)
		}
	}

	// 长度异常
	if len([]rune(content)) > 5000 {
		hits = append(hits, "content_too_long")
	}

	if len(hits) == 0 {
		return 0, "safe", model.SuggestionPass, "[]"
	}

	// 命中敏感词
	riskScore := float64(len(hits) * 30)
	if riskScore > 100 {
		riskScore = 100
	}
	if riskScore >= 60 {
		return riskScore, "block", model.SuggestionBlock, fmt.Sprintf(`["%s"]`, strings.Join(hits, `","`))
	}
	return riskScore, "review", model.SuggestionReview, fmt.Sprintf(`["%s"]`, strings.Join(hits, `","`))
}

func (s *aiExtendService) GetAuditResult(taskID string) (*dto.AuditResultInfo, error) {
	a, err := s.repo.FindAuditResultByTaskID(taskID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrAuditResultNotFound
		}
		return nil, err
	}
	return toAuditResultInfo(a), nil
}

func (s *aiExtendService) ListAuditResults(level string, page, pageSize int) ([]dto.AuditResultInfo, int64, error) {
	list, total, err := s.repo.ListAuditResults(level, page, pageSize)
	if err != nil {
		return nil, 0, err
	}
	result := make([]dto.AuditResultInfo, 0, len(list))
	for i := range list {
		result = append(result, *toAuditResultInfo(&list[i]))
	}
	return result, total, nil
}

func (s *aiExtendService) ReviewAudit(taskID string, passed bool, reviewer string, comment string) error {
	a, err := s.repo.FindAuditResultByTaskID(taskID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrAuditResultNotFound
		}
		return err
	}

	passedInt := 0
	if passed {
		passedInt = 1
	}
	level := "block"
	suggestion := comment
	if passed {
		level = "safe"
		suggestion = "人工复核通过"
	}
	if suggestion == "" {
		suggestion = "人工复核"
	}

	return s.repo.UpdateAuditResultFields(a.ID, map[string]interface{}{
		"passed":     passedInt,
		"risk_level": level,
		"suggestion": suggestion,
	})
}

// ===== 推荐 =====

func (s *aiExtendService) GetRecommendations(userID uint, recType string, limit int) ([]dto.RecommendationInfo, error) {
	if userID == 0 {
		return nil, errors.New("用户ID不能为空")
	}
	if limit <= 0 {
		limit = 10
	}

	list, err := s.repo.ListRecommendationsByUser(userID, recType, limit)
	if err != nil {
		return nil, err
	}

	// 精简版：若用户暂无推荐记录，生成 mock 推荐
	if len(list) == 0 {
		mockList := s.generateMockRecommendations(userID, recType, limit)
		for i := range mockList {
			_ = s.repo.CreateRecommendation(&mockList[i])
		}
		list = mockList
	}

	result := make([]dto.RecommendationInfo, 0, len(list))
	for i := range list {
		result = append(result, *toRecommendationInfo(&list[i]))
	}
	return result, nil
}

// generateMockRecommendations 生成 mock 推荐（精简版：热门 + 随机分）
func (s *aiExtendService) generateMockRecommendations(userID uint, recType string, limit int) []model.Recommendation {
	if recType == "" {
		recType = model.RecTypeHot
	}
	now := time.Now()
	list := make([]model.Recommendation, 0, limit)
	for i := 1; i <= limit; i++ {
		list = append(list, model.Recommendation{
			UserID:      userID,
			BizModule:   "ershou",
			ContentType: model.RecContentTypeItem,
			ContentID:   fmt.Sprintf("mock_%d_%d", userID, i),
			RecType:     recType,
			Score:       float64(100 - i*5),
			Reason:      "热门推荐",
			IsClicked:   0,
			CreatedAt:   now,
		})
	}
	return list
}

func (s *aiExtendService) TrackClick(recID uint, req dto.TrackClickRequest) error {
	if recID == 0 {
		return errors.New("推荐ID不能为空")
	}
	now := time.Now()
	return s.repo.UpdateRecommendationFields(recID, map[string]interface{}{
		"is_clicked": 1,
		"clicked_at": &now,
	})
}

func (s *aiExtendService) TrackDwell(recID uint, req dto.TrackDwellRequest) error {
	if recID == 0 {
		return errors.New("推荐ID不能为空")
	}
	return s.repo.UpdateRecommendationFields(recID, map[string]interface{}{
		"dwell_ms": req.DwellMs,
	})
}

func (s *aiExtendService) FeedbackRecommendation(recID uint, req dto.FeedbackRecommendationRequest) error {
	if recID == 0 {
		return errors.New("推荐ID不能为空")
	}
	fields := map[string]interface{}{}
	if req.Feedback > 0 {
		fields["is_liked"] = 1
		fields["is_disliked"] = 0
	} else {
		fields["is_liked"] = 0
		fields["is_disliked"] = 1
	}
	return s.repo.UpdateRecommendationFields(recID, fields)
}

// ===== 模型配置 =====

func (s *aiExtendService) ListModelConfigs(modelType string, page, pageSize int) ([]dto.ModelConfigInfo, int64, error) {
	list, total, err := s.repo.ListModelConfigs(modelType, page, pageSize)
	if err != nil {
		return nil, 0, err
	}
	result := make([]dto.ModelConfigInfo, 0, len(list))
	for i := range list {
		result = append(result, *toModelConfigInfo(&list[i]))
	}
	return result, total, nil
}

func (s *aiExtendService) UpsertModelConfig(req dto.UpsertModelConfigRequest) (uint, error) {
	configType := req.ConfigType
	if configType == "" {
		configType = model.ConfigTypeString
	}
	status := req.Status
	if status == 0 {
		status = 1
	}
	c := &model.ModelConfig{
		ModelID:     req.ModelID,
		ConfigKey:   req.ConfigKey,
		ConfigValue: req.ConfigValue,
		ConfigType:  configType,
		Description: req.Description,
		Status:      status,
	}
	if err := s.repo.UpsertModelConfig(c); err != nil {
		return 0, err
	}
	return c.ID, nil
}

func (s *aiExtendService) DeleteModelConfig(id uint) error {
	if id == 0 {
		return ErrModelConfigNotFound
	}
	return s.repo.DeleteModelConfig(id)
}

// ===== 对话 =====

func (s *aiExtendService) CreateChatSession(userID uint, req dto.CreateChatSessionRequest) (*dto.ChatSessionInfo, error) {
	if userID == 0 {
		return nil, errors.New("用户ID不能为空")
	}

	modelName := req.ModelName
	if modelName == "" {
		modelName = "default-llm"
	}
	contextLength := req.ContextLength
	if contextLength <= 0 {
		contextLength = 10
	}

	sessionID := generateChatSessionID()
	sess := &model.ChatSession{
		SessionID:     sessionID,
		UserID:        userID,
		Title:         req.Title,
		ModelName:     modelName,
		SystemPrompt:  req.SystemPrompt,
		ContextLength: contextLength,
		TotalMessages: 0,
		TotalTokens:   0,
		Status:        1,
		Extra:         "{}",
	}
	if err := s.repo.CreateChatSession(sess); err != nil {
		return nil, err
	}
	return toChatSessionInfo(sess), nil
}

func (s *aiExtendService) ListChatSessions(userID uint, page, pageSize int) ([]dto.ChatSessionInfo, int64, error) {
	list, total, err := s.repo.ListChatSessionsByUser(userID, page, pageSize)
	if err != nil {
		return nil, 0, err
	}
	result := make([]dto.ChatSessionInfo, 0, len(list))
	for i := range list {
		result = append(result, *toChatSessionInfo(&list[i]))
	}
	return result, total, nil
}

func (s *aiExtendService) Chat(userID uint, req dto.ChatRequest) (*dto.ChatResponse, error) {
	if userID == 0 {
		return nil, errors.New("用户ID不能为空")
	}
	if req.SessionID == "" {
		return nil, ErrChatSessionNotFound
	}
	if strings.TrimSpace(req.Content) == "" {
		return nil, errors.New("对话内容不能为空")
	}

	sess, err := s.repo.FindChatSessionByID(req.SessionID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrChatSessionNotFound
		}
		return nil, err
	}
	if sess.Status != 1 {
		return nil, ErrChatSessionClosed
	}
	if sess.UserID != userID {
		return nil, errors.New("无权操作他人在对话")
	}

	// 保存用户消息
	imagesJSON := "[]"
	if len(req.Images) > 0 {
		imagesJSON = fmt.Sprintf(`["%s"]`, strings.Join(req.Images, `","`))
	}
	userMsg := &model.ChatMessage{
		SessionID: sess.SessionID,
		UserID:    userID,
		Role:      model.ChatRoleUser,
		Content:   req.Content,
		ModelName: sess.ModelName,
		Images:    imagesJSON,
	}
	if err := s.repo.CreateChatMessage(userMsg); err != nil {
		return nil, err
	}

	// 精简版：未对接真实 LLM，返回提示性回复
	start := time.Now()
	reply := s.mockChatReply(req.Content)
	assistantMsg := &model.ChatMessage{
		SessionID: sess.SessionID,
		UserID:    userID,
		Role:      model.ChatRoleAssistant,
		Content:   reply,
		ModelName: sess.ModelName,
		CostMs:    int(time.Since(start).Milliseconds()),
	}
	if err := s.repo.CreateChatMessage(assistantMsg); err != nil {
		return nil, err
	}

	// 更新会话计数
	_ = s.repo.UpdateChatSessionFields(sess.ID, map[string]interface{}{
		"total_messages": sess.TotalMessages + 2,
	})

	return &dto.ChatResponse{
		Message: *toChatMessageInfo(assistantMsg),
	}, nil
}

// mockChatReply 模拟对话回复
func (s *aiExtendService) mockChatReply(content string) string {
	content = strings.TrimSpace(content)
	if content == "" {
		return "请输入您的问题"
	}
	return fmt.Sprintf("已收到您的提问：「%s」。当前为精简模式，未对接真实 LLM，请配置 AI 模型后使用。", content)
}

func (s *aiExtendService) ListChatMessages(sessionID string, page, pageSize int) ([]dto.ChatMessageInfo, int64, error) {
	if sessionID == "" {
		return nil, 0, ErrChatSessionNotFound
	}
	list, total, err := s.repo.ListChatMessagesBySession(sessionID, page, pageSize)
	if err != nil {
		return nil, 0, err
	}
	result := make([]dto.ChatMessageInfo, 0, len(list))
	for i := range list {
		result = append(result, *toChatMessageInfo(&list[i]))
	}
	return result, total, nil
}

func (s *aiExtendService) MessageFeedback(messageID uint, req dto.MessageFeedbackRequest) error {
	if messageID == 0 {
		return ErrChatMessageNotFound
	}
	return s.repo.UpdateMessageFeedback(messageID, req.Feedback, req.FeedbackText)
}

// ===== 训练数据 =====

func (s *aiExtendService) ListTrainingData(dataType string, page, pageSize int) ([]dto.TrainingDataInfo, int64, error) {
	list, total, err := s.repo.ListTrainingData(dataType, page, pageSize)
	if err != nil {
		return nil, 0, err
	}
	result := make([]dto.TrainingDataInfo, 0, len(list))
	for i := range list {
		result = append(result, *toTrainingDataInfo(&list[i]))
	}
	return result, total, nil
}

func (s *aiExtendService) CreateTrainingData(req dto.CreateTrainingDataRequest) (uint, error) {
	t := &model.TrainingData{
		DataType:     req.DataType,
		BizModule:    req.BizModule,
		BizID:        req.BizID,
		UserID:       req.UserID,
		Input:        req.Input,
		Output:       req.Output,
		Label:        req.Label,
		QualityScore: req.QualityScore,
		IsUsed:       0,
		Extra:        "{}",
	}
	if err := s.repo.CreateTrainingData(t); err != nil {
		return 0, err
	}
	return t.ID, nil
}

func (s *aiExtendService) MarkTrainingDataUsed(id uint) error {
	if id == 0 {
		return ErrTrainingDataNotFound
	}
	return s.repo.MarkTrainingDataUsed(id, 0)
}

// ===== 统计 =====

func (s *aiExtendService) GetStatistics() (*dto.AIStatisticsResponse, error) {
	resp := &dto.AIStatisticsResponse{}
	resp.TotalTasks, _ = s.repo.StatTotalTasks()
	resp.TotalAuditResults, _ = s.repo.StatTotalTasks() // 审核结果数复用任务统计近似
	resp.PassedAudit, _ = s.repo.StatPassedAudit()
	resp.BlockedAudit, _ = s.repo.StatBlockedAudit()
	resp.TotalRecommendations, _ = s.repo.StatTotalRecommendations()
	resp.ClickedRecommendations, _ = s.repo.StatClickedRecommendations()
	resp.TotalChatSessions, _ = s.repo.StatTotalChatSessions()
	resp.TotalChatMessages, _ = s.repo.StatTotalChatMessages()
	resp.TotalTrainingData, _ = s.repo.StatTotalTrainingData()
	return resp, nil
}

// ===== 工具函数 =====

// generateChatSessionID 生成对话会话ID
func generateChatSessionID() string {
	return fmt.Sprintf("CHAT%s%s", time.Now().Format("20060102150405"), utils.RandomNumber(6))
}

// toAuditResultInfo model → dto
func toAuditResultInfo(a *model.AuditResult) *dto.AuditResultInfo {
	return &dto.AuditResultInfo{
		ID:          a.ID,
		TaskID:      a.TaskID,
		BizModule:   a.BizModule,
		BizID:       a.BizID,
		UserID:      a.UserID,
		ContentType: a.ContentType,
		ContentHash: a.ContentHash,
		Algorithm:   a.Algorithm,
		RiskScore:   a.RiskScore,
		RiskLevel:   a.RiskLevel,
		Labels:      a.Labels,
		HitRules:    a.HitRules,
		Suggestion:  a.Suggestion,
		Passed:      a.Passed,
		CostMs:      a.CostMs,
		CreatedAt:   a.CreatedAt,
	}
}

// toRecommendationInfo model → dto
func toRecommendationInfo(r *model.Recommendation) *dto.RecommendationInfo {
	return &dto.RecommendationInfo{
		ID:          r.ID,
		UserID:      r.UserID,
		BizModule:   r.BizModule,
		ContentType: r.ContentType,
		ContentID:   r.ContentID,
		RecType:     r.RecType,
		Score:       r.Score,
		Reason:      r.Reason,
		IsClicked:   r.IsClicked,
		ClickedAt:   r.ClickedAt,
		IsLiked:     r.IsLiked,
		IsDisliked:  r.IsDisliked,
		DwellMs:     r.DwellMs,
		Extra:       r.Extra,
		CreatedAt:   r.CreatedAt,
	}
}

// toModelConfigInfo model → dto
func toModelConfigInfo(c *model.ModelConfig) *dto.ModelConfigInfo {
	return &dto.ModelConfigInfo{
		ID:          c.ID,
		ModelID:     c.ModelID,
		ConfigKey:   c.ConfigKey,
		ConfigValue: c.ConfigValue,
		ConfigType:  c.ConfigType,
		Description: c.Description,
		Status:      c.Status,
		CreatedAt:   c.CreatedAt,
		UpdatedAt:   c.UpdatedAt,
	}
}

// toChatSessionInfo model → dto
func toChatSessionInfo(s *model.ChatSession) *dto.ChatSessionInfo {
	return &dto.ChatSessionInfo{
		ID:            s.ID,
		SessionID:     s.SessionID,
		UserID:        s.UserID,
		Title:         s.Title,
		ModelName:     s.ModelName,
		SystemPrompt:  s.SystemPrompt,
		ContextLength: s.ContextLength,
		TotalMessages: s.TotalMessages,
		TotalTokens:   s.TotalTokens,
		Status:        s.Status,
		Extra:         s.Extra,
		CreatedAt:     s.CreatedAt,
		UpdatedAt:     s.UpdatedAt,
	}
}

// toChatMessageInfo model → dto
func toChatMessageInfo(m *model.ChatMessage) *dto.ChatMessageInfo {
	return &dto.ChatMessageInfo{
		ID:           m.ID,
		SessionID:    m.SessionID,
		UserID:       m.UserID,
		Role:         m.Role,
		Content:      m.Content,
		Tokens:       m.Tokens,
		ModelName:    m.ModelName,
		CostMs:       m.CostMs,
		Images:       m.Images,
		Feedback:     m.Feedback,
		FeedbackText: m.FeedbackText,
		CreatedAt:    m.CreatedAt,
	}
}

// toTrainingDataInfo model → dto
func toTrainingDataInfo(t *model.TrainingData) *dto.TrainingDataInfo {
	return &dto.TrainingDataInfo{
		ID:           t.ID,
		DataType:     t.DataType,
		BizModule:    t.BizModule,
		BizID:        t.BizID,
		UserID:       t.UserID,
		Input:        t.Input,
		Output:       t.Output,
		Label:        t.Label,
		QualityScore: t.QualityScore,
		IsUsed:       t.IsUsed,
		UsedModelID:  t.UsedModelID,
		Extra:        t.Extra,
		CreatedAt:    t.CreatedAt,
		UpdatedAt:    t.UpdatedAt,
	}
}
