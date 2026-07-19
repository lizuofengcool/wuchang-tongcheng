// Package model AI 智能中台精简版数据模型
// 依据 ershou 模块依赖：图文审核 + 标题优化 + 价格建议 + 描述生成
// 依据需求文档 1.10：4 维数据隔离（region_id + user_id）
package model

import (
	"time"

	"wuchang-tongcheng/internal/pkg/database"
)

// === 任务状态 ===
const (
	TaskStatusPending   = 0 // 待处理
	TaskStatusRunning   = 1 // 处理中
	TaskStatusSucceeded = 2 // 已完成
	TaskStatusFailed    = 3 // 失败
)

// === 任务类型 ===
const (
	TaskTypeAuditImage        = "audit_image"        // 图文审核
	TaskTypeAuditText         = "audit_text"         // 文本审核
	TaskTypeOptimizeTitle     = "optimize_title"     // 标题优化
	TaskTypeGenerateDesc      = "generate_description" // 描述生成
	TaskTypeSuggestPrice      = "suggest_price"      // 价格建议
	TaskTypeGenerateSummary   = "generate_summary"   // 摘要生成
)

// === 模型提供商 ===
const (
	ProviderAliyun  = "aliyun"  // 阿里云
	ProviderTencent = "tencent" // 腾讯云
	ProviderQwen    = "qwen"    // 通义千问
	ProviderWenxin  = "wenxin"  // 文心一言
	ProviderXFYun   = "xfyun"   // 讯飞星火
)

// === 模型类型 ===
const (
	ModelTypeAuditImage = "audit_image" // 图片审核
	ModelTypeAuditText  = "audit_text"  // 文本审核
	ModelTypeLLM        = "llm"         // 大语言模型
)

// === 模板类型 ===
const (
	TemplateTypeOptimizeTitle = "optimize_title"
	TemplateTypeGenerateDesc  = "generate_description"
	TemplateTypeSuggestPrice  = "suggest_price"
	TemplateTypeAuditText     = "audit_text"
)

// === 生成类型 ===
const (
	GenerationTypeTitle       = "title"
	GenerationTypeDescription = "description"
	GenerationTypePrice       = "price"
	GenerationTypeSummary     = "summary"
)

// Task AI 任务
type Task struct {
	database.RegionBaseModel

	TaskID    string `gorm:"size:64;not null;uniqueIndex" json:"task_id"`        // 任务ID
	TaskType  string `gorm:"size:32;index" json:"task_type"`                    // 任务类型
	UserID    uint   `gorm:"index" json:"user_id"`                              // 用户ID
	Input     string `gorm:"type:jsonb;default:'{}'::jsonb" json:"input"`       // 输入 JSON
	Output    string `gorm:"type:jsonb;default:'{}'::jsonb" json:"output"`      // 输出 JSON
	Status    int    `gorm:"default:0;index" json:"status"`                     // 状态
	ModelName string `gorm:"size:64" json:"model_name"`                         // 使用的模型
	CostMs    int    `gorm:"default:0" json:"cost_ms"`                          // 耗时（毫秒）
	ErrorMsg  string `gorm:"type:text" json:"error_msg"`                        // 错误信息
}

// TableName 表名
func (Task) TableName() string { return "ai_tasks" }

// Model AI 模型配置
type Model struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	ModelName string    `gorm:"size:64;not null;uniqueIndex" json:"model_name"` // 模型名
	Provider  string    `gorm:"size:32;index" json:"provider"`                 // 提供商
	ModelType string    `gorm:"size:32;index" json:"model_type"`               // 类型
	APIKey    string    `gorm:"size:256" json:"api_key"`                       // API Key
	Endpoint  string    `gorm:"size:256" json:"endpoint"`                      // Endpoint
	Config    string    `gorm:"type:jsonb;default:'{}'::jsonb" json:"config"`  // 配置 JSON
	Status    int       `gorm:"default:1;index" json:"status"`                 // 状态 1启用 0禁用
	CreatedAt time.Time `gorm:"not null;default:now()" json:"created_at"`
	UpdatedAt time.Time `gorm:"not null;default:now()" json:"updated_at"`
}

// TableName 表名
func (Model) TableName() string { return "ai_models" }

// Prompt 提示词模板
type Prompt struct {
	ID           uint      `gorm:"primarykey" json:"id"`
	TemplateName string    `gorm:"size:64;not null;uniqueIndex" json:"template_name"` // 模板名
	TemplateType string    `gorm:"size:32;index" json:"template_type"`                // 类型
	Content      string    `gorm:"type:text" json:"content"`                          // 模板内容
	Variables    string    `gorm:"type:jsonb;default:'[]'::jsonb" json:"variables"`   // 变量列表 JSON
	Description  string    `gorm:"size:256" json:"description"`                       // 描述
	Status       int       `gorm:"default:1;index" json:"status"`                     // 状态
	CreatedAt    time.Time `gorm:"not null;default:now()" json:"created_at"`
	UpdatedAt    time.Time `gorm:"not null;default:now()" json:"updated_at"`
}

// TableName 表名
func (Prompt) TableName() string { return "ai_prompts" }

// Generation AI 生成记录
type Generation struct {
	database.RegionBaseModel

	TaskID         string `gorm:"size:64;index" json:"task_id"`                         // 关联任务ID
	UserID         uint   `gorm:"index;not null" json:"user_id"`                        // 用户ID
	GenerationType string `gorm:"size:32;index" json:"generation_type"`                 // 生成类型
	Input          string `gorm:"type:jsonb;default:'{}'::jsonb" json:"input"`          // 输入 JSON
	Output         string `gorm:"type:jsonb;default:'{}'::jsonb" json:"output"`         // 输出 JSON
	Rating         int    `gorm:"default:0" json:"rating"`                              // 用户评分 1-5
	Feedback       string `gorm:"type:text" json:"feedback"`                            // 用户反馈
}

// TableName 表名
func (Generation) TableName() string { return "ai_generations" }
