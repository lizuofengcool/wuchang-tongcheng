// Package repository AI 中台扩展数据访问层
// 依据 016_ai_full.sql：审核结果/推荐记录/模型配置/对话会话/对话消息/训练数据
package repository

import (
	"errors"

	"wuchang-tongcheng/internal/modules/ai/model"

	"gorm.io/gorm"
)

// AIExtendRepository AI 中台扩展仓储接口
type AIExtendRepository interface {
	// 审核结果
	CreateAuditResult(a *model.AuditResult) error
	FindAuditResultByTaskID(taskID string) (*model.AuditResult, error)
	ListAuditResults(level string, page, pageSize int) ([]model.AuditResult, int64, error)
	ListAuditResultsByLevel(level string, page, pageSize int) ([]model.AuditResult, int64, error)
	UpdateAuditResultFields(id uint, fields map[string]interface{}) error

	// 推荐记录
	CreateRecommendation(r *model.Recommendation) error
	ListRecommendationsByUser(userID uint, recType string, limit int) ([]model.Recommendation, error)
	UpdateRecommendationFields(id uint, fields map[string]interface{}) error
	StatClickedRecommendations() (int64, error)

	// 模型配置
	UpsertModelConfig(c *model.ModelConfig) error
	FindModelConfig(modelID uint, configKey string) (*model.ModelConfig, error)
	ListModelConfigs(modelType string, page, pageSize int) ([]model.ModelConfig, int64, error)
	DeleteModelConfig(id uint) error

	// 对话会话
	CreateChatSession(s *model.ChatSession) error
	FindChatSessionByID(sessionID string) (*model.ChatSession, error)
	ListChatSessionsByUser(userID uint, page, pageSize int) ([]model.ChatSession, int64, error)
	UpdateChatSessionFields(id uint, fields map[string]interface{}) error

	// 对话消息
	CreateChatMessage(m *model.ChatMessage) error
	ListChatMessagesBySession(sessionID string, page, pageSize int) ([]model.ChatMessage, int64, error)
	UpdateMessageFeedback(id uint, feedback int, feedbackText string) error
	StatTotalChatMessages() (int64, error)

	// 训练数据
	CreateTrainingData(t *model.TrainingData) error
	ListTrainingData(dataType string, page, pageSize int) ([]model.TrainingData, int64, error)
	MarkTrainingDataUsed(id uint, modelID uint) error

	// 统计
	StatTotalTasks() (int64, error)
	StatPassedAudit() (int64, error)
	StatBlockedAudit() (int64, error)
	StatTotalRecommendations() (int64, error)
	StatTotalChatSessions() (int64, error)
	StatTotalTrainingData() (int64, error)
}

type aiExtendRepository struct {
	db *gorm.DB
}

// NewAIExtendRepository 创建扩展仓储实例
func NewAIExtendRepository(db *gorm.DB) AIExtendRepository {
	return &aiExtendRepository{db: db}
}

// ===== 审核结果 =====

func (r *aiExtendRepository) CreateAuditResult(a *model.AuditResult) error {
	return r.db.Create(a).Error
}

func (r *aiExtendRepository) FindAuditResultByTaskID(taskID string) (*model.AuditResult, error) {
	var a model.AuditResult
	if err := r.db.Where("task_id = ?", taskID).First(&a).Error; err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *aiExtendRepository) ListAuditResults(level string, page, pageSize int) ([]model.AuditResult, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	var list []model.AuditResult
	var total int64
	q := r.db.Model(&model.AuditResult{})
	if level != "" {
		q = q.Where("risk_level = ?", level)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := q.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *aiExtendRepository) ListAuditResultsByLevel(level string, page, pageSize int) ([]model.AuditResult, int64, error) {
	return r.ListAuditResults(level, page, pageSize)
}

func (r *aiExtendRepository) UpdateAuditResultFields(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.AuditResult{}).Where("id = ?", id).Updates(fields).Error
}

// ===== 推荐记录 =====

func (r *aiExtendRepository) CreateRecommendation(rec *model.Recommendation) error {
	return r.db.Create(rec).Error
}

func (r *aiExtendRepository) ListRecommendationsByUser(userID uint, recType string, limit int) ([]model.Recommendation, error) {
	if limit <= 0 {
		limit = 20
	}
	var list []model.Recommendation
	q := r.db.Model(&model.Recommendation{}).Where("user_id = ?", userID)
	if recType != "" {
		q = q.Where("rec_type = ?", recType)
	}
	if err := q.Order("score DESC, id DESC").Limit(limit).Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *aiExtendRepository) UpdateRecommendationFields(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.Recommendation{}).Where("id = ?", id).Updates(fields).Error
}

func (r *aiExtendRepository) StatClickedRecommendations() (int64, error) {
	var count int64
	err := r.db.Model(&model.Recommendation{}).Where("is_clicked = ?", 1).Count(&count).Error
	return count, err
}

// ===== 模型配置 =====

func (r *aiExtendRepository) UpsertModelConfig(c *model.ModelConfig) error {
	// 按 model_id + config_key 唯一索引 upsert
	var existing model.ModelConfig
	err := r.db.Where("model_id = ? AND config_key = ?", c.ModelID, c.ConfigKey).First(&existing).Error
	if err == nil {
		// 已存在，更新
		fields := map[string]interface{}{
			"config_value": c.ConfigValue,
			"config_type":  c.ConfigType,
			"description":  c.Description,
			"status":       c.Status,
		}
		return r.db.Model(&model.ModelConfig{}).Where("id = ?", existing.ID).Updates(fields).Error
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	return r.db.Create(c).Error
}

func (r *aiExtendRepository) FindModelConfig(modelID uint, configKey string) (*model.ModelConfig, error) {
	var c model.ModelConfig
	if err := r.db.Where("model_id = ? AND config_key = ?", modelID, configKey).First(&c).Error; err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *aiExtendRepository) ListModelConfigs(modelType string, page, pageSize int) ([]model.ModelConfig, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	var list []model.ModelConfig
	var total int64
	q := r.db.Model(&model.ModelConfig{})
	if modelType != "" {
		// 通过 config_key 前缀简单过滤（model_type 未独立字段）
		q = q.Where("config_key LIKE ?", modelType+"%")
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := q.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *aiExtendRepository) DeleteModelConfig(id uint) error {
	return r.db.Delete(&model.ModelConfig{}, id).Error
}

// ===== 对话会话 =====

func (r *aiExtendRepository) CreateChatSession(s *model.ChatSession) error {
	return r.db.Create(s).Error
}

func (r *aiExtendRepository) FindChatSessionByID(sessionID string) (*model.ChatSession, error) {
	var s model.ChatSession
	if err := r.db.Where("session_id = ?", sessionID).First(&s).Error; err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *aiExtendRepository) ListChatSessionsByUser(userID uint, page, pageSize int) ([]model.ChatSession, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	var list []model.ChatSession
	var total int64
	q := r.db.Model(&model.ChatSession{}).Where("user_id = ?", userID)
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := q.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *aiExtendRepository) UpdateChatSessionFields(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.ChatSession{}).Where("id = ?", id).Updates(fields).Error
}

// ===== 对话消息 =====

func (r *aiExtendRepository) CreateChatMessage(m *model.ChatMessage) error {
	return r.db.Create(m).Error
}

func (r *aiExtendRepository) ListChatMessagesBySession(sessionID string, page, pageSize int) ([]model.ChatMessage, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 50
	}
	var list []model.ChatMessage
	var total int64
	q := r.db.Model(&model.ChatMessage{}).Where("session_id = ?", sessionID)
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := q.Order("id ASC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *aiExtendRepository) UpdateMessageFeedback(id uint, feedback int, feedbackText string) error {
	return r.db.Model(&model.ChatMessage{}).Where("id = ?", id).Updates(map[string]interface{}{
		"feedback":      feedback,
		"feedback_text": feedbackText,
	}).Error
}

func (r *aiExtendRepository) StatTotalChatMessages() (int64, error) {
	var count int64
	err := r.db.Model(&model.ChatMessage{}).Count(&count).Error
	return count, err
}

// ===== 训练数据 =====

func (r *aiExtendRepository) CreateTrainingData(t *model.TrainingData) error {
	return r.db.Create(t).Error
}

func (r *aiExtendRepository) ListTrainingData(dataType string, page, pageSize int) ([]model.TrainingData, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	var list []model.TrainingData
	var total int64
	q := r.db.Model(&model.TrainingData{})
	if dataType != "" {
		q = q.Where("data_type = ?", dataType)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := q.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *aiExtendRepository) MarkTrainingDataUsed(id uint, modelID uint) error {
	return r.db.Model(&model.TrainingData{}).Where("id = ?", id).Updates(map[string]interface{}{
		"is_used":        1,
		"used_model_id":  modelID,
	}).Error
}

// ===== 统计 =====

func (r *aiExtendRepository) StatTotalTasks() (int64, error) {
	var count int64
	err := r.db.Model(&model.Task{}).Count(&count).Error
	return count, err
}

func (r *aiExtendRepository) StatPassedAudit() (int64, error) {
	var count int64
	err := r.db.Model(&model.AuditResult{}).Where("passed = ?", 1).Count(&count).Error
	return count, err
}

func (r *aiExtendRepository) StatBlockedAudit() (int64, error) {
	var count int64
	err := r.db.Model(&model.AuditResult{}).Where("risk_level = ?", "block").Count(&count).Error
	return count, err
}

func (r *aiExtendRepository) StatTotalRecommendations() (int64, error) {
	var count int64
	err := r.db.Model(&model.Recommendation{}).Count(&count).Error
	return count, err
}

func (r *aiExtendRepository) StatTotalChatSessions() (int64, error) {
	var count int64
	err := r.db.Model(&model.ChatSession{}).Count(&count).Error
	return count, err
}

func (r *aiExtendRepository) StatTotalTrainingData() (int64, error) {
	var count int64
	err := r.db.Model(&model.TrainingData{}).Count(&count).Error
	return count, err
}
