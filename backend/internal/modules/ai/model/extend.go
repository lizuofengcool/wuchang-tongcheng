// Package model AI 中台扩展数据模型
// 依据 016_ai_full.sql：审核结果/推荐记录/模型配置/对话会话/对话消息/训练数据
package model

import (
	"time"

	"wuchang-tongcheng/internal/pkg/database"
)

// === 审核内容类型 ===
const (
	ContentTypeText  = "text"
	ContentTypeImage = "image"
	ContentTypeVideo = "video"
)

// === 审核算法 ===
const (
	AlgoLocal   = "local"
	AlgoDFA     = "dfa"
	AlgoAliyun  = "aliyun"
	AlgoTencent = "tencent"
)

// === 审核建议 ===
const (
	SuggestionPass   = "pass"
	SuggestionReview  = "review"
	SuggestionBlock   = "block"
)

// === 推荐内容类型 ===
const (
	RecContentTypeItem = "item"
	RecContentTypeNews = "news"
	RecContentTypeVideo = "video"
	RecContentTypeAd   = "ad"
)

// === 推荐类型 ===
const (
	RecTypeHot          = "hot"
	RecTypeNew          = "new"
	RecTypePersonalized = "personalized"
	RecTypeSimilar      = "similar"
)

// === 模型参数类型 ===
const (
	ConfigTypeString  = "string"
	ConfigTypeNumber  = "number"
	ConfigTypeBoolean = "boolean"
	ConfigTypeJSON    = "json"
)

// === 对话角色 ===
const (
	ChatRoleUser      = "user"
	ChatRoleAssistant = "assistant"
	ChatRoleSystem    = "system"
)

// === 反馈类型 ===
const (
	FeedbackNegative = -1
	FeedbackNone     = 0
	FeedbackPositive = 1
)

// === 训练数据类型 ===
const (
	TrainingDataTypeText         = "text"
	TrainingDataTypeImage        = "image"
	TrainingDataTypeConversation = "conversation"
)

// AuditResult AI 审核结果
type AuditResult struct {
	ID          uint      `gorm:"primarykey" json:"id"`
	RegionID    uint      `gorm:"index;not null;default:1" json:"region_id"`
	TaskID      string    `gorm:"size:64;not null;default:'';index" json:"task_id"`
	BizModule   string    `gorm:"size:32;not null;default:'';index" json:"biz_module"`
	BizID       string    `gorm:"size:128;not null;default:''" json:"biz_id"`
	UserID      uint      `gorm:"default:0;index" json:"user_id"`
	ContentType string    `gorm:"size:32;not null;default:'text'" json:"content_type"`
	ContentHash string    `gorm:"size:64;not null;default:''" json:"content_hash"`
	Algorithm   string    `gorm:"size:32;not null;default:'local'" json:"algorithm"`
	RiskScore   float64   `gorm:"type:decimal(5,2);default:0.00" json:"risk_score"`
	RiskLevel   string    `gorm:"size:16;not null;default:'safe';index" json:"risk_level"`
	Labels      string    `gorm:"type:jsonb" json:"labels"`
	HitRules    string    `gorm:"type:jsonb" json:"hit_rules"`
	Suggestion  string    `gorm:"size:256;not null;default:''" json:"suggestion"`
	Passed      int       `gorm:"default:1;index" json:"passed"`
	CostMs      int       `gorm:"default:0" json:"cost_ms"`
	CreatedAt   time.Time `gorm:"not null;default:now()" json:"created_at"`
	UpdatedAt   time.Time `gorm:"not null;default:now()" json:"updated_at"`
}

// TableName 表名
func (AuditResult) TableName() string { return "ai_audit_results" }

// Recommendation AI 推荐记录
type Recommendation struct {
	ID          uint       `gorm:"primarykey" json:"id"`
	RegionID    uint       `gorm:"index;not null;default:1" json:"region_id"`
	UserID      uint       `gorm:"not null;index" json:"user_id"`
	BizModule   string     `gorm:"size:32;not null;default:''" json:"biz_module"`
	ContentType string     `gorm:"size:32;not null;default:'item'" json:"content_type"`
	ContentID   string     `gorm:"size:128;not null;default:''" json:"content_id"`
	RecType     string     `gorm:"size:32;not null;default:'hot';index" json:"rec_type"`
	Score       float64    `gorm:"type:decimal(8,2);default:0.00" json:"score"`
	Reason      string     `gorm:"size:256;not null;default:''" json:"reason"`
	IsClicked   int        `gorm:"default:0;index" json:"is_clicked"`
	ClickedAt   *time.Time `gorm:"index" json:"clicked_at"`
	IsLiked     int        `gorm:"default:0" json:"is_liked"`
	IsDisliked  int        `gorm:"default:0" json:"is_disliked"`
	DwellMs     int        `gorm:"default:0" json:"dwell_ms"`
	Extra        string `gorm:"type:jsonb" json:"extra"`
	CreatedAt   time.Time  `gorm:"not null;default:now();index" json:"created_at"`
}

// TableName 表名
func (Recommendation) TableName() string { return "ai_recommendations" }

// ModelConfig 模型参数配置
type ModelConfig struct {
	ID          uint      `gorm:"primarykey" json:"id"`
	RegionID    uint      `gorm:"index;not null;default:1" json:"region_id"`
	ModelID     uint      `gorm:"not null;index;uniqueIndex:uk_ai_model_configs,priority:1" json:"model_id"`
	ConfigKey   string    `gorm:"size:64;not null;uniqueIndex:uk_ai_model_configs,priority:2" json:"config_key"`
	ConfigValue string    `gorm:"size:256;not null;default:''" json:"config_value"`
	ConfigType  string    `gorm:"size:16;not null;default:'string'" json:"config_type"`
	Description string    `gorm:"size:256;not null;default:''" json:"description"`
	Status      int       `gorm:"default:1;index" json:"status"`
	CreatedAt   time.Time `gorm:"not null;default:now()" json:"created_at"`
	UpdatedAt   time.Time `gorm:"not null;default:now()" json:"updated_at"`
}

// TableName 表名
func (ModelConfig) TableName() string { return "ai_model_configs" }

// ChatSession AI 对话会话
type ChatSession struct {
	database.RegionBaseModel

	SessionID      string `gorm:"size:64;not null;uniqueIndex" json:"session_id"`
	UserID         uint   `gorm:"not null;index" json:"user_id"`
	Title          string `gorm:"size:128;not null;default:''" json:"title"`
	ModelName      string `gorm:"size:64;not null;default:''" json:"model_name"`
	SystemPrompt   string `gorm:"type:text;not null;default:''" json:"system_prompt"`
	ContextLength  int    `gorm:"default:10" json:"context_length"`
	TotalMessages  int    `gorm:"default:0" json:"total_messages"`
	TotalTokens    int    `gorm:"default:0" json:"total_tokens"`
	Status         int    `gorm:"default:1;index" json:"status"`
	Extra          string `gorm:"type:jsonb;default:'{}'::jsonb" json:"extra"`
}

// TableName 表名
func (ChatSession) TableName() string { return "ai_chat_sessions" }

// ChatMessage AI 对话消息
type ChatMessage struct {
	ID           uint      `gorm:"primarykey" json:"id"`
	RegionID     uint      `gorm:"index;not null;default:1" json:"region_id"`
	SessionID    string    `gorm:"size:64;not null;index" json:"session_id"`
	UserID       uint      `gorm:"not null;index" json:"user_id"`
	Role         string    `gorm:"size:16;not null;default:'user'" json:"role"`
	Content      string    `gorm:"type:text;not null;default:''" json:"content"`
	Tokens       int       `gorm:"default:0" json:"tokens"`
	ModelName    string    `gorm:"size:64;not null;default:''" json:"model_name"`
	CostMs       int       `gorm:"default:0" json:"cost_ms"`
	Images       string    `gorm:"type:jsonb;default:'[]'::jsonb" json:"images"`
	Feedback     int       `gorm:"default:0" json:"feedback"`
	FeedbackText string    `gorm:"type:text;not null;default:''" json:"feedback_text"`
	CreatedAt    time.Time `gorm:"not null;default:now();index" json:"created_at"`
}

// TableName 表名
func (ChatMessage) TableName() string { return "ai_chat_messages" }

// TrainingData 训练数据
type TrainingData struct {
	ID           uint      `gorm:"primarykey" json:"id"`
	RegionID     uint      `gorm:"index;not null;default:1" json:"region_id"`
	DataType     string    `gorm:"size:32;not null;default:'text';index" json:"data_type"`
	BizModule    string    `gorm:"size:32;not null;default:''" json:"biz_module"`
	BizID        string    `gorm:"size:128;not null;default:''" json:"biz_id"`
	UserID       uint      `gorm:"default:0" json:"user_id"`
	Input        string    `gorm:"type:text;not null;default:''" json:"input"`
	Output       string    `gorm:"type:text;not null;default:''" json:"output"`
	Label        string    `gorm:"size:64;not null;default:'';index" json:"label"`
	QualityScore float64   `gorm:"type:decimal(3,2);default:0.00" json:"quality_score"`
	IsUsed       int       `gorm:"default:0;index" json:"is_used"`
	UsedModelID  uint      `gorm:"default:0" json:"used_model_id"`
	Extra        string    `gorm:"type:jsonb;default:'{}'::jsonb" json:"extra"`
	CreatedAt    time.Time `gorm:"not null;default:now()" json:"created_at"`
	UpdatedAt    time.Time `gorm:"not null;default:now()" json:"updated_at"`
}

// TableName 表名
func (TrainingData) TableName() string { return "ai_training_data" }
