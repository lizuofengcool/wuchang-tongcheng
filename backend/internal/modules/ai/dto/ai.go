// Package dto AI 智能中台精简版数据传输对象
package dto

import "time"

// CreateTaskRequest 创建 AI 任务请求
type CreateTaskRequest struct {
	TaskType  string                 `json:"task_type" binding:"required,oneof=audit_image audit_text optimize_title generate_description suggest_price generate_summary"`
	UserID    uint                   `json:"user_id"`
	Input     map[string]interface{} `json:"input" binding:"required"`
	ModelName string                 `json:"model_name"` // 指定模型，空则按 task_type 选默认
}

// TaskInfo AI 任务信息
type TaskInfo struct {
	ID        uint      `json:"id"`
	TaskID    string    `json:"task_id"`
	TaskType  string    `json:"task_type"`
	UserID    uint      `json:"user_id"`
	Input     string    `json:"input"`
	Output    string    `json:"output"`
	Status    int       `json:"status"`
	ModelName string    `json:"model_name"`
	CostMs    int       `json:"cost_ms"`
	ErrorMsg  string    `json:"error_msg"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TaskListRequest 任务列表请求
type TaskListRequest struct {
	UserID   uint   `form:"user_id" json:"user_id"`
	TaskType string `form:"task_type" json:"task_type"`
	Status   int    `form:"status" json:"status"`
	Page     int    `form:"page" json:"page"`
	PageSize int    `form:"page_size" json:"page_size"`
}

// ModelRequest 模型管理请求
type ModelRequest struct {
	ModelName string                 `json:"model_name" binding:"required,max=64"`
	Provider  string                 `json:"provider" binding:"required,oneof=aliyun tencent qwen wenxin xfyun"`
	ModelType string                 `json:"model_type" binding:"required,oneof=audit_image audit_text llm"`
	APIKey    string                 `json:"api_key"`
	Endpoint  string                 `json:"endpoint"`
	Config    map[string]interface{} `json:"config"`
}

// ModelInfo 模型信息
type ModelInfo struct {
	ID        uint      `json:"id"`
	ModelName string    `json:"model_name"`
	Provider  string    `json:"provider"`
	ModelType string    `json:"model_type"`
	APIKey    string    `json:"api_key"` // 注意：列表接口应做脱敏
	Endpoint  string    `json:"endpoint"`
	Config    string    `json:"config"`
	Status    int       `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

// PromptRequest 提示词模板请求
type PromptRequest struct {
	TemplateName string                 `json:"template_name" binding:"required,max=64"`
	TemplateType string                 `json:"template_type" binding:"required,oneof=optimize_title generate_description suggest_price audit_text"`
	Content      string                 `json:"content" binding:"required"`
	Variables    []string               `json:"variables"`
	Description  string                 `json:"description"`
}

// PromptInfo 提示词模板信息
type PromptInfo struct {
	ID           uint      `json:"id"`
	TemplateName string    `json:"template_name"`
	TemplateType string    `json:"template_type"`
	Content      string    `json:"content"`
	Variables    string    `json:"variables"`
	Description  string    `json:"description"`
	Status       int       `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
}

// RenderPromptRequest 渲染提示词请求
type RenderPromptRequest struct {
	TemplateName string                 `json:"template_name" binding:"required"`
	Variables    map[string]interface{} `json:"variables" binding:"required"`
}

// RenderPromptResponse 渲染提示词响应
type RenderPromptResponse struct {
	Content string `json:"content"` // 渲染后的提示词
}

// GenerationInfo 生成记录信息
type GenerationInfo struct {
	ID             uint      `json:"id"`
	TaskID         string    `json:"task_id"`
	UserID         uint      `json:"user_id"`
	GenerationType string    `json:"generation_type"`
	Input          string    `json:"input"`
	Output         string    `json:"output"`
	Rating         int       `json:"rating"`
	Feedback       string    `json:"feedback"`
	CreatedAt      time.Time `json:"created_at"`
}

// RateGenerationRequest 评分请求
type RateGenerationRequest struct {
	GenerationID uint   `json:"generation_id" binding:"required"`
	Rating       int    `json:"rating" binding:"required,min=1,max=5"`
	Feedback     string `json:"feedback" binding:"max=512"`
}

// OptimizeTitleRequest 标题优化请求
type OptimizeTitleRequest struct {
	Title    string `json:"title" binding:"required,max=128"`
	Category string `json:"category"`
	Brand    string `json:"brand"`
}

// OptimizeTitleResponse 标题优化响应
type OptimizeTitleResponse struct {
	OriginalTitle string   `json:"original_title"`
	Optimized     string   `json:"optimized"`     // 优化后标题
	Alternatives  []string `json:"alternatives"`  // 备选
	TaskID        string   `json:"task_id"`
}

// GenerateDescriptionRequest 描述生成请求
type GenerateDescriptionRequest struct {
	Title    string `json:"title" binding:"required,max=128"`
	Category string `json:"category"`
	Brand    string `json:"brand"`
	Condition string `json:"condition"` // 成色
	Keywords []string `json:"keywords"`
}

// GenerateDescriptionResponse 描述生成响应
type GenerateDescriptionResponse struct {
	Description string   `json:"description"`
	Alternatives []string `json:"alternatives"`
	TaskID       string   `json:"task_id"`
}

// SuggestPriceRequest 价格建议请求
type SuggestPriceRequest struct {
	Title    string  `json:"title" binding:"required,max=128"`
	Category string  `json:"category"`
	Brand    string  `json:"brand"`
	Condition string `json:"condition"`
	OriginalPrice float64 `json:"original_price"` // 原价（可选）
}

// SuggestPriceResponse 价格建议响应
type SuggestPriceResponse struct {
	SuggestedPrice float64 `json:"suggested_price"`
	MinPrice       float64 `json:"min_price"`
	MaxPrice       float64 `json:"max_price"`
	Reason         string  `json:"reason"`
	TaskID         string  `json:"task_id"`
}
