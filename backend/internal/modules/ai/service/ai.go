// Package service AI 智能中台精简版业务逻辑层
// 依据 ershou 模块依赖：图文审核 + 标题优化 + 价格建议 + 描述生成
// 暴露 AIService 接口供其他模块直接 import 调用（不通过 HTTP）
package service

import (
	"encoding/json"
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

var (
	ErrTaskNotFound      = errors.New("AI任务不存在")
	ErrModelNotFound     = errors.New("AI模型不存在")
	ErrPromptNotFound    = errors.New("提示词模板不存在")
	ErrGenerationNotFound = errors.New("生成记录不存在")
	ErrModelDisabled     = errors.New("AI模型已禁用")
	ErrUnsupportedType   = errors.New("不支持的AI任务类型")
)

// AIService AI 中台业务接口
// 暴露给其他模块直接 import 调用，不通过 HTTP
type AIService interface {
	// 任务
	CreateTask(regionID uint, req *dto.CreateTaskRequest) (*dto.TaskInfo, error)
	GetTask(taskID string) (*dto.TaskInfo, error)
	ListTasks(req *dto.TaskListRequest) ([]dto.TaskInfo, int64, error)
	// 任务执行（精简版：直接同步执行，不调用外部 AI API）
	RunTask(taskID string) (*dto.TaskInfo, error)

	// 模型管理
	AddModel(req *dto.ModelRequest) error
	ListModels(provider, modelType string, page, pageSize int) ([]model.Model, int64, error)
	UpdateModelStatus(id uint, status int) error

	// 提示词模板
	AddPrompt(req *dto.PromptRequest) error
	ListPrompts(templateType string, page, pageSize int) ([]model.Prompt, int64, error)
	RenderPrompt(templateName string, variables map[string]interface{}) (string, error)

	// 生成记录
	ListMyGenerations(userID uint, page, pageSize int) ([]dto.GenerationInfo, int64, error)
	RateGeneration(userID uint, req *dto.RateGenerationRequest) error

	// 高级接口（供其他模块直接调用，内部创建任务并执行）
	OptimizeTitle(regionID, userID uint, req *dto.OptimizeTitleRequest) (*dto.OptimizeTitleResponse, error)
	GenerateDescription(regionID, userID uint, req *dto.GenerateDescriptionRequest) (*dto.GenerateDescriptionResponse, error)
	SuggestPrice(regionID, userID uint, req *dto.SuggestPriceRequest) (*dto.SuggestPriceResponse, error)
}

type aiService struct {
	repo repository.AIRepository
}

// NewAIService 创建 service 实例
func NewAIService(repo repository.AIRepository) AIService {
	return &aiService{repo: repo}
}

// ===== 任务 =====

// CreateTask 创建任务（不立即执行）
func (s *aiService) CreateTask(regionID uint, req *dto.CreateTaskRequest) (*dto.TaskInfo, error) {
	taskID := generateTaskID()
	inputJSON, _ := json.Marshal(req.Input)

	t := &model.Task{
		TaskID:    taskID,
		TaskType:  req.TaskType,
		UserID:    req.UserID,
		Input:     string(inputJSON),
		Status:    model.TaskStatusPending,
		ModelName: req.ModelName,
	}
	t.RegionID = regionID

	if err := s.repo.CreateTask(t); err != nil {
		return nil, err
	}
	return toTaskInfo(t), nil
}

// GetTask 查询任务
func (s *aiService) GetTask(taskID string) (*dto.TaskInfo, error) {
	t, err := s.repo.FindTaskByTaskID(taskID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTaskNotFound
		}
		return nil, err
	}
	return toTaskInfo(t), nil
}

// ListTasks 任务列表
func (s *aiService) ListTasks(req *dto.TaskListRequest) ([]dto.TaskInfo, int64, error) {
	q := &repository.ListTasksQuery{
		UserID:   req.UserID,
		TaskType: req.TaskType,
		Status:   req.Status,
		Page:     req.Page,
		PageSize: req.PageSize,
	}
	list, total, err := s.repo.ListTasks(q)
	if err != nil {
		return nil, 0, err
	}
	result := make([]dto.TaskInfo, 0, len(list))
	for i := range list {
		result = append(result, *toTaskInfo(&list[i]))
	}
	return result, total, nil
}

// RunTask 执行任务（精简版：不调用真实 AI API，返回模板化结果）
func (s *aiService) RunTask(taskID string) (*dto.TaskInfo, error) {
	t, err := s.repo.FindTaskByTaskID(taskID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTaskNotFound
		}
		return nil, err
	}

	start := time.Now()

	// 标记为处理中
	_ = s.repo.UpdateTaskFields(t.ID, map[string]interface{}{
		"status":     model.TaskStatusRunning,
		"started_at": start,
	})

	// 精简版：根据 task_type 生成 mock 结果（实际场景应调用 LLM API）
	var output map[string]interface{}
	var runErr error
	switch t.TaskType {
	case model.TaskTypeOptimizeTitle:
		output, runErr = s.mockOptimizeTitle(t.Input)
	case model.TaskTypeGenerateDesc:
		output, runErr = s.mockGenerateDescription(t.Input)
	case model.TaskTypeSuggestPrice:
		output, runErr = s.mockSuggestPrice(t.Input)
	case model.TaskTypeAuditText:
		output, runErr = s.mockAuditText(t.Input)
	case model.TaskTypeAuditImage:
		output = map[string]interface{}{"passed": true, "risk_level": "safe"}
	case model.TaskTypeGenerateSummary:
		output, runErr = s.mockGenerateSummary(t.Input)
	default:
		runErr = ErrUnsupportedType
	}

	costMs := int(time.Since(start).Milliseconds())

	if runErr != nil {
		_ = s.repo.UpdateTaskFields(t.ID, map[string]interface{}{
			"status":    model.TaskStatusFailed,
			"error_msg": runErr.Error(),
			"cost_ms":   costMs,
		})
		return nil, runErr
	}

	outputJSON, _ := json.Marshal(output)
	_ = s.repo.UpdateTaskFields(t.ID, map[string]interface{}{
		"status": model.TaskStatusSucceeded,
		"output": string(outputJSON),
		"cost_ms": costMs,
	})

	t.Output = string(outputJSON)
	t.Status = model.TaskStatusSucceeded
	t.CostMs = costMs
	return toTaskInfo(t), nil
}

// ===== 模型管理 =====

// AddModel 添加模型
func (s *aiService) AddModel(req *dto.ModelRequest) error {
	// 幂等：已存在则跳过
	if existing, err := s.repo.FindModelByName(req.ModelName); err == nil && existing != nil {
		return nil
	}

	configJSON := "{}"
	if req.Config != nil {
		b, _ := json.Marshal(req.Config)
		configJSON = string(b)
	}

	m := &model.Model{
		ModelName: req.ModelName,
		Provider:  req.Provider,
		ModelType: req.ModelType,
		APIKey:    req.APIKey,
		Endpoint:  req.Endpoint,
		Config:    configJSON,
		Status:    1,
	}
	return s.repo.CreateModel(m)
}

// ListModels 模型列表
func (s *aiService) ListModels(provider, modelType string, page, pageSize int) ([]model.Model, int64, error) {
	return s.repo.ListModels(provider, modelType, page, pageSize)
}

// UpdateModelStatus 更新模型状态
func (s *aiService) UpdateModelStatus(id uint, status int) error {
	return s.repo.UpdateModelFields(id, map[string]interface{}{"status": status})
}

// ===== 提示词模板 =====

// AddPrompt 添加提示词模板
func (s *aiService) AddPrompt(req *dto.PromptRequest) error {
	if existing, err := s.repo.FindPromptByName(req.TemplateName); err == nil && existing != nil {
		return nil
	}

	variablesJSON := "[]"
	if len(req.Variables) > 0 {
		b, _ := json.Marshal(req.Variables)
		variablesJSON = string(b)
	}

	p := &model.Prompt{
		TemplateName: req.TemplateName,
		TemplateType: req.TemplateType,
		Content:      req.Content,
		Variables:    variablesJSON,
		Description:  req.Description,
		Status:       1,
	}
	return s.repo.CreatePrompt(p)
}

// ListPrompts 模板列表
func (s *aiService) ListPrompts(templateType string, page, pageSize int) ([]model.Prompt, int64, error) {
	return s.repo.ListPrompts(templateType, page, pageSize)
}

// RenderPrompt 渲染提示词
// 精简版：使用简单的 {{var}} 模板替换
func (s *aiService) RenderPrompt(templateName string, variables map[string]interface{}) (string, error) {
	p, err := s.repo.FindPromptByName(templateName)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", ErrPromptNotFound
		}
		return "", err
	}

	content := p.Content
	for k, v := range variables {
		placeholder := fmt.Sprintf("{{%s}}", k)
		content = strings.ReplaceAll(content, placeholder, fmt.Sprintf("%v", v))
	}
	return content, nil
}

// ===== 生成记录 =====

// ListMyGenerations 我的生成记录
func (s *aiService) ListMyGenerations(userID uint, page, pageSize int) ([]dto.GenerationInfo, int64, error) {
	list, total, err := s.repo.ListGenerationsByUser(userID, page, pageSize)
	if err != nil {
		return nil, 0, err
	}
	result := make([]dto.GenerationInfo, 0, len(list))
	for i := range list {
		result = append(result, *toGenerationInfo(&list[i]))
	}
	return result, total, nil
}

// RateGeneration 评分
func (s *aiService) RateGeneration(userID uint, req *dto.RateGenerationRequest) error {
	g, err := s.repo.FindGenerationByID(req.GenerationID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrGenerationNotFound
		}
		return err
	}
	if g.UserID != userID {
		return errors.New("无权评价他人生成记录")
	}
	return s.repo.UpdateGenerationFields(g.ID, map[string]interface{}{
		"rating":   req.Rating,
		"feedback": req.Feedback,
	})
}

// ===== 高级接口 =====

// OptimizeTitle 标题优化
func (s *aiService) OptimizeTitle(regionID, userID uint, req *dto.OptimizeTitleRequest) (*dto.OptimizeTitleResponse, error) {
	input, _ := json.Marshal(map[string]interface{}{
		"title":    req.Title,
		"category": req.Category,
		"brand":    req.Brand,
	})

	t := &model.Task{
		TaskID:    generateTaskID(),
		TaskType:  model.TaskTypeOptimizeTitle,
		UserID:    userID,
		Input:     string(input),
		Status:    model.TaskStatusPending,
		ModelName: "default-llm",
	}
	t.RegionID = regionID
	if err := s.repo.CreateTask(t); err != nil {
		return nil, err
	}

	if _, err := s.RunTask(t.TaskID); err != nil {
		return nil, err
	}

	// 读取结果
	updated, _ := s.repo.FindTaskByTaskID(t.TaskID)
	var output map[string]interface{}
	_ = json.Unmarshal([]byte(updated.Output), &output)

	optimized, _ := output["optimized"].(string)
	alts, _ := output["alternatives"].([]interface{})
	altStrs := make([]string, 0, len(alts))
	for _, a := range alts {
		if s, ok := a.(string); ok {
			altStrs = append(altStrs, s)
		}
	}

	// 记录 Generation
	_ = s.recordGeneration(regionID, userID, t.TaskID, model.GenerationTypeTitle, input, []byte(updated.Output))

	return &dto.OptimizeTitleResponse{
		OriginalTitle: req.Title,
		Optimized:     optimized,
		Alternatives:  altStrs,
		TaskID:        t.TaskID,
	}, nil
}

// GenerateDescription 描述生成
func (s *aiService) GenerateDescription(regionID, userID uint, req *dto.GenerateDescriptionRequest) (*dto.GenerateDescriptionResponse, error) {
	input, _ := json.Marshal(map[string]interface{}{
		"title":     req.Title,
		"category":  req.Category,
		"brand":     req.Brand,
		"condition": req.Condition,
		"keywords":  req.Keywords,
	})

	t := &model.Task{
		TaskID:    generateTaskID(),
		TaskType:  model.TaskTypeGenerateDesc,
		UserID:    userID,
		Input:     string(input),
		Status:    model.TaskStatusPending,
		ModelName: "default-llm",
	}
	t.RegionID = regionID
	if err := s.repo.CreateTask(t); err != nil {
		return nil, err
	}
	if _, err := s.RunTask(t.TaskID); err != nil {
		return nil, err
	}

	updated, _ := s.repo.FindTaskByTaskID(t.TaskID)
	var output map[string]interface{}
	_ = json.Unmarshal([]byte(updated.Output), &output)

	desc, _ := output["description"].(string)
	alts, _ := output["alternatives"].([]interface{})
	altStrs := make([]string, 0, len(alts))
	for _, a := range alts {
		if s, ok := a.(string); ok {
			altStrs = append(altStrs, s)
		}
	}

	_ = s.recordGeneration(regionID, userID, t.TaskID, model.GenerationTypeDescription, input, []byte(updated.Output))

	return &dto.GenerateDescriptionResponse{
		Description:  desc,
		Alternatives: altStrs,
		TaskID:       t.TaskID,
	}, nil
}

// SuggestPrice 价格建议
func (s *aiService) SuggestPrice(regionID, userID uint, req *dto.SuggestPriceRequest) (*dto.SuggestPriceResponse, error) {
	input, _ := json.Marshal(map[string]interface{}{
		"title":          req.Title,
		"category":       req.Category,
		"brand":          req.Brand,
		"condition":      req.Condition,
		"original_price": req.OriginalPrice,
	})

	t := &model.Task{
		TaskID:    generateTaskID(),
		TaskType:  model.TaskTypeSuggestPrice,
		UserID:    userID,
		Input:     string(input),
		Status:    model.TaskStatusPending,
		ModelName: "default-llm",
	}
	t.RegionID = regionID
	if err := s.repo.CreateTask(t); err != nil {
		return nil, err
	}
	if _, err := s.RunTask(t.TaskID); err != nil {
		return nil, err
	}

	updated, _ := s.repo.FindTaskByTaskID(t.TaskID)
	var output map[string]interface{}
	_ = json.Unmarshal([]byte(updated.Output), &output)

	suggested, _ := output["suggested_price"].(float64)
	minP, _ := output["min_price"].(float64)
	maxP, _ := output["max_price"].(float64)
	reason, _ := output["reason"].(string)

	_ = s.recordGeneration(regionID, userID, t.TaskID, model.GenerationTypePrice, input, []byte(updated.Output))

	return &dto.SuggestPriceResponse{
		SuggestedPrice: suggested,
		MinPrice:       minP,
		MaxPrice:       maxP,
		Reason:         reason,
		TaskID:         t.TaskID,
	}, nil
}

// recordGeneration 记录生成结果
func (s *aiService) recordGeneration(regionID, userID uint, taskID, genType string, input, output []byte) error {
	g := &model.Generation{
		TaskID:         taskID,
		UserID:         userID,
		GenerationType: genType,
		Input:          string(input),
		Output:         string(output),
	}
	g.RegionID = regionID
	return s.repo.CreateGeneration(g)
}

// ===== Mock 实现（精简版） =====

// mockOptimizeTitle 模拟标题优化
func (s *aiService) mockOptimizeTitle(inputJSON string) (map[string]interface{}, error) {
	var input struct {
		Title    string `json:"title"`
		Category string `json:"category"`
		Brand    string `json:"brand"`
	}
	_ = json.Unmarshal([]byte(inputJSON), &input)

	title := strings.TrimSpace(input.Title)
	if title == "" {
		title = "闲置好物转让"
	}

	optimized := title
	if input.Brand != "" {
		optimized = fmt.Sprintf("【%s】%s", input.Brand, title)
	}
	if input.Category != "" && !strings.Contains(title, input.Category) {
		optimized = fmt.Sprintf("%s | %s", optimized, input.Category)
	}

	return map[string]interface{}{
		"optimized":    optimized,
		"alternatives": []string{title + "（急转）", title + "（9成新）"},
	}, nil
}

// mockGenerateDescription 模拟描述生成
func (s *aiService) mockGenerateDescription(inputJSON string) (map[string]interface{}, error) {
	var input struct {
		Title     string   `json:"title"`
		Category  string   `json:"category"`
		Brand     string   `json:"brand"`
		Condition string   `json:"condition"`
		Keywords  []string `json:"keywords"`
	}
	_ = json.Unmarshal([]byte(inputJSON), &input)

	condition := input.Condition
	if condition == "" {
		condition = "9成新"
	}
	keywords := strings.Join(input.Keywords, "、")

	desc := fmt.Sprintf("【%s】%s，%s，使用一段时间，成色良好，无损坏无维修史。", input.Brand, input.Title, condition)
	if keywords != "" {
		desc += fmt.Sprintf("关键词：%s。", keywords)
	}
	desc += "支持当面验货，同城自提，价格可议。"

	return map[string]interface{}{
		"description": desc,
		"alternatives": []string{
			desc + " 欢迎私聊详询。",
			"闲置转让 " + input.Title + "，" + condition + "，需要的联系。",
		},
	}, nil
}

// mockSuggestPrice 模拟价格建议
func (s *aiService) mockSuggestPrice(inputJSON string) (map[string]interface{}, error) {
	var input struct {
		Title          string  `json:"title"`
		Category       string  `json:"category"`
		Brand          string  `json:"brand"`
		Condition      string  `json:"condition"`
		OriginalPrice  float64 `json:"original_price"`
	}
	_ = json.Unmarshal([]byte(inputJSON), &input)

	// 精简版：基于原价按成色打折
	if input.OriginalPrice <= 0 {
		input.OriginalPrice = 100 // 默认 100 元
	}

	conditionFactor := 0.6
	switch input.Condition {
	case "全新":
		conditionFactor = 0.9
	case "9成新":
		conditionFactor = 0.7
	case "8成新":
		conditionFactor = 0.55
	case "7成新":
		conditionFactor = 0.4
	case "旧":
		conditionFactor = 0.25
	}

	suggested := input.OriginalPrice * conditionFactor
	minP := suggested * 0.85
	maxP := suggested * 1.15

	reason := fmt.Sprintf("基于原价 %.2f 元，成色 %s 折算", input.OriginalPrice, input.Condition)

	return map[string]interface{}{
		"suggested_price": suggested,
		"min_price":       minP,
		"max_price":       maxP,
		"reason":          reason,
	}, nil
}

// mockAuditText 模拟文本审核（精简版：不调用真实审核 API）
func (s *aiService) mockAuditText(inputJSON string) (map[string]interface{}, error) {
	return map[string]interface{}{
		"passed":      true,
		"risk_level":  "safe",
		"suggestion":  "通过",
	}, nil
}

// mockGenerateSummary 模拟摘要生成
func (s *aiService) mockGenerateSummary(inputJSON string) (map[string]interface{}, error) {
	var input struct {
		Content string `json:"content"`
	}
	_ = json.Unmarshal([]byte(inputJSON), &input)

	content := input.Content
	if len([]rune(content)) > 50 {
		content = string([]rune(content)[:50]) + "..."
	}
	return map[string]interface{}{
		"summary": content,
	}, nil
}

// ===== 工具函数 =====

// generateTaskID 生成任务ID
func generateTaskID() string {
	return fmt.Sprintf("AI%s%s", time.Now().Format("20060102150405"), utils.RandomNumber(6))
}

// toTaskInfo model → dto
func toTaskInfo(t *model.Task) *dto.TaskInfo {
	return &dto.TaskInfo{
		ID:        t.ID,
		TaskID:    t.TaskID,
		TaskType:  t.TaskType,
		UserID:    t.UserID,
		Input:     t.Input,
		Output:    t.Output,
		Status:    t.Status,
		ModelName: t.ModelName,
		CostMs:    t.CostMs,
		ErrorMsg:  t.ErrorMsg,
		CreatedAt: t.CreatedAt,
		UpdatedAt: t.UpdatedAt,
	}
}

// toGenerationInfo model → dto
func toGenerationInfo(g *model.Generation) *dto.GenerationInfo {
	return &dto.GenerationInfo{
		ID:             g.ID,
		TaskID:         g.TaskID,
		UserID:         g.UserID,
		GenerationType: g.GenerationType,
		Input:          g.Input,
		Output:         g.Output,
		Rating:         g.Rating,
		Feedback:       g.Feedback,
		CreatedAt:      g.CreatedAt,
	}
}
