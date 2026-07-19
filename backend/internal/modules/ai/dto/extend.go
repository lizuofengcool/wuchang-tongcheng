// Package dto AI 中台扩展数据传输对象
package dto

import "time"

// AuditResultInfo 审核结果信息
type AuditResultInfo struct {
	ID          uint      `json:"id"`
	TaskID      string    `json:"task_id"`
	BizModule   string    `json:"biz_module"`
	BizID       string    `json:"biz_id"`
	UserID      uint      `json:"user_id"`
	ContentType string    `json:"content_type"`
	ContentHash string    `json:"content_hash"`
	Algorithm   string    `json:"algorithm"`
	RiskScore   float64   `json:"risk_score"`
	RiskLevel   string    `json:"risk_level"`
	Labels      string    `json:"labels"`
	HitRules    string    `json:"hit_rules"`
	Suggestion  string    `json:"suggestion"`
	Passed      int       `json:"passed"`
	CostMs      int       `json:"cost_ms"`
	CreatedAt   time.Time `json:"created_at"`
}

// AuditContentRequest AI 审核请求
type AuditContentRequest struct {
	BizModule   string `json:"biz_module" binding:"required"`
	BizID       string `json:"biz_id" binding:"required"`
	UserID      uint   `json:"user_id"`
	ContentType string `json:"content_type" binding:"required,oneof=text image video"`
	ContentHash string `json:"content_hash"`
	Algorithm   string `json:"algorithm" binding:"omitempty,oneof=local dfa aliyun tencent"`
}

// RecommendationInfo 推荐信息
type RecommendationInfo struct {
	ID          uint       `json:"id"`
	UserID      uint       `json:"user_id"`
	BizModule   string     `json:"biz_module"`
	ContentType string     `json:"content_type"`
	ContentID   string     `json:"content_id"`
	RecType     string     `json:"rec_type"`
	Score       float64    `json:"score"`
	Reason      string     `json:"reason"`
	IsClicked   int        `json:"is_clicked"`
	ClickedAt   *time.Time `json:"clicked_at"`
	IsLiked     int        `json:"is_liked"`
	IsDisliked  int        `json:"is_disliked"`
	DwellMs     int        `json:"dwell_ms"`
	Extra       string     `json:"extra"`
	CreatedAt   time.Time  `json:"created_at"`
}

// TrackClickRequest 推荐点击追踪请求
type TrackClickRequest struct {
	RecommendationID uint `json:"recommendation_id" binding:"required"`
}

// TrackDwellRequest 停留时长追踪请求
type TrackDwellRequest struct {
	RecommendationID uint `json:"recommendation_id" binding:"required"`
	DwellMs          int  `json:"dwell_ms" binding:"required"`
}

// FeedbackRecommendationRequest 推荐反馈请求
type FeedbackRecommendationRequest struct {
	RecommendationID uint   `json:"recommendation_id" binding:"required"`
	Feedback         int    `json:"feedback" binding:"required,oneof=-1 1"`
	FeedbackText     string `json:"feedback_text" binding:"max=256"`
}

// ModelConfigInfo 模型配置信息
type ModelConfigInfo struct {
	ID          uint      `json:"id"`
	ModelID     uint      `json:"model_id"`
	ConfigKey   string    `json:"config_key"`
	ConfigValue string    `json:"config_value"`
	ConfigType  string    `json:"config_type"`
	Description string    `json:"description"`
	Status      int       `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// UpsertModelConfigRequest 创建/更新模型配置请求
type UpsertModelConfigRequest struct {
	ModelID     uint   `json:"model_id" binding:"required"`
	ConfigKey   string `json:"config_key" binding:"required,max=64"`
	ConfigValue string `json:"config_value" binding:"required,max=256"`
	ConfigType  string `json:"config_type" binding:"omitempty,oneof=string number boolean json"`
	Description string `json:"description" binding:"max=256"`
	Status      int    `json:"status"`
}

// ChatSessionInfo 对话会话信息
type ChatSessionInfo struct {
	ID             uint      `json:"id"`
	SessionID      string    `json:"session_id"`
	UserID         uint      `json:"user_id"`
	Title          string    `json:"title"`
	ModelName      string    `json:"model_name"`
	SystemPrompt   string    `json:"system_prompt"`
	ContextLength  int       `json:"context_length"`
	TotalMessages  int       `json:"total_messages"`
	TotalTokens    int       `json:"total_tokens"`
	Status         int       `json:"status"`
	Extra          string    `json:"extra"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// CreateChatSessionRequest 创建对话会话请求
type CreateChatSessionRequest struct {
	Title         string `json:"title" binding:"max=128"`
	ModelName     string `json:"model_name" binding:"max=64"`
	SystemPrompt  string `json:"system_prompt"`
	ContextLength int    `json:"context_length"`
}

// ChatMessageInfo 对话消息信息
type ChatMessageInfo struct {
	ID           uint      `json:"id"`
	SessionID    string    `json:"session_id"`
	UserID       uint      `json:"user_id"`
	Role         string    `json:"role"`
	Content      string    `json:"content"`
	Tokens       int       `json:"tokens"`
	ModelName    string    `json:"model_name"`
	CostMs       int       `json:"cost_ms"`
	Images       string    `json:"images"`
	Feedback     int       `json:"feedback"`
	FeedbackText string    `json:"feedback_text"`
	CreatedAt    time.Time `json:"created_at"`
}

// ChatRequest 对话请求
type ChatRequest struct {
	SessionID string   `json:"session_id" binding:"required"`
	Content  string   `json:"content" binding:"required"`
	Images   []string `json:"images"`
}

// ChatResponse 对话响应
type ChatResponse struct {
	Message ChatMessageInfo `json:"message"`
}

// MessageFeedbackRequest 消息反馈请求
type MessageFeedbackRequest struct {
	MessageID     uint   `json:"message_id" binding:"required"`
	Feedback      int    `json:"feedback" binding:"required,oneof=-1 1"`
	FeedbackText  string `json:"feedback_text" binding:"max=256"`
}

// TrainingDataInfo 训练数据信息
type TrainingDataInfo struct {
	ID           uint      `json:"id"`
	DataType     string    `json:"data_type"`
	BizModule    string    `json:"biz_module"`
	BizID        string    `json:"biz_id"`
	UserID       uint      `json:"user_id"`
	Input        string    `json:"input"`
	Output       string    `json:"output"`
	Label        string    `json:"label"`
	QualityScore float64   `json:"quality_score"`
	IsUsed       int       `json:"is_used"`
	UsedModelID  uint      `json:"used_model_id"`
	Extra        string    `json:"extra"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// CreateTrainingDataRequest 创建训练数据请求
type CreateTrainingDataRequest struct {
	DataType     string `json:"data_type" binding:"required,oneof=text image conversation"`
	BizModule    string `json:"biz_module" binding:"max=32"`
	BizID        string `json:"biz_id" binding:"max=128"`
	UserID       uint   `json:"user_id"`
	Input        string `json:"input" binding:"required"`
	Output       string `json:"output" binding:"required"`
	Label        string `json:"label" binding:"max=64"`
	QualityScore float64 `json:"quality_score"`
	Extra        string `json:"extra"`
}

// AIStatisticsResponse AI 统计响应
type AIStatisticsResponse struct {
	TotalTasks          int64 `json:"total_tasks"`
	TotalAuditResults   int64 `json:"total_audit_results"`
	PassedAudit         int64 `json:"passed_audit"`
	BlockedAudit        int64 `json:"blocked_audit"`
	TotalRecommendations int64 `json:"total_recommendations"`
	ClickedRecommendations int64 `json:"clicked_recommendations"`
	TotalChatSessions   int64 `json:"total_chat_sessions"`
	TotalChatMessages   int64 `json:"total_chat_messages"`
	TotalTrainingData   int64 `json:"total_training_data"`
}
