// Package repository AI 智能中台精简版数据访问层
package repository

import (
	"wuchang-tongcheng/internal/modules/ai/model"

	"gorm.io/gorm"
)

// AIRepository AI 中台仓储接口
type AIRepository interface {
	// 任务
	CreateTask(t *model.Task) error
	FindTaskByID(id uint) (*model.Task, error)
	FindTaskByTaskID(taskID string) (*model.Task, error)
	ListTasks(req *ListTasksQuery) ([]model.Task, int64, error)
	UpdateTaskFields(id uint, fields map[string]interface{}) error

	// 模型
	CreateModel(m *model.Model) error
	FindModelByName(name string) (*model.Model, error)
	FindActiveModelByType(modelType string) (*model.Model, error)
	ListModels(provider, modelType string, page, pageSize int) ([]model.Model, int64, error)
	UpdateModelFields(id uint, fields map[string]interface{}) error

	// 提示词模板
	CreatePrompt(p *model.Prompt) error
	FindPromptByName(name string) (*model.Prompt, error)
	FindActivePromptByType(templateType string) (*model.Prompt, error)
	ListPrompts(templateType string, page, pageSize int) ([]model.Prompt, int64, error)
	UpdatePromptFields(id uint, fields map[string]interface{}) error

	// 生成记录
	CreateGeneration(g *model.Generation) error
	FindGenerationByID(id uint) (*model.Generation, error)
	ListGenerationsByUser(userID uint, page, pageSize int) ([]model.Generation, int64, error)
	ListGenerationsByTask(taskID string) ([]model.Generation, error)
	UpdateGenerationFields(id uint, fields map[string]interface{}) error
}

// ListTasksQuery 任务列表查询参数
type ListTasksQuery struct {
	RegionID uint
	UserID   uint
	TaskType string
	Status   int
	Page     int
	PageSize int
}

type aiRepository struct {
	db *gorm.DB
}

// NewAIRepository 创建仓储实例
func NewAIRepository(db *gorm.DB) AIRepository {
	return &aiRepository{db: db}
}

// ===== 任务 =====

func (r *aiRepository) CreateTask(t *model.Task) error {
	return r.db.Create(t).Error
}

func (r *aiRepository) FindTaskByID(id uint) (*model.Task, error) {
	var t model.Task
	if err := r.db.First(&t, id).Error; err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *aiRepository) FindTaskByTaskID(taskID string) (*model.Task, error) {
	var t model.Task
	if err := r.db.Where("task_id = ?", taskID).First(&t).Error; err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *aiRepository) ListTasks(req *ListTasksQuery) ([]model.Task, int64, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}
	var list []model.Task
	var total int64
	q := r.db.Model(&model.Task{})
	if req.RegionID > 0 {
		q = q.Where("region_id = ?", req.RegionID)
	}
	if req.UserID > 0 {
		q = q.Where("user_id = ?", req.UserID)
	}
	if req.TaskType != "" {
		q = q.Where("task_type = ?", req.TaskType)
	}
	if req.Status >= 0 {
		q = q.Where("status = ?", req.Status)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := q.Order("created_at DESC, id DESC").
		Offset((req.Page - 1) * req.PageSize).Limit(req.PageSize).
		Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *aiRepository) UpdateTaskFields(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.Task{}).Where("id = ?", id).Updates(fields).Error
}

// ===== 模型 =====

func (r *aiRepository) CreateModel(m *model.Model) error {
	return r.db.Create(m).Error
}

func (r *aiRepository) FindModelByName(name string) (*model.Model, error) {
	var m model.Model
	if err := r.db.Where("model_name = ?", name).First(&m).Error; err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *aiRepository) FindActiveModelByType(modelType string) (*model.Model, error) {
	var m model.Model
	if err := r.db.Where("model_type = ? AND status = ?", modelType, 1).
		Order("id ASC").First(&m).Error; err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *aiRepository) ListModels(provider, modelType string, page, pageSize int) ([]model.Model, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	var list []model.Model
	var total int64
	q := r.db.Model(&model.Model{})
	if provider != "" {
		q = q.Where("provider = ?", provider)
	}
	if modelType != "" {
		q = q.Where("model_type = ?", modelType)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := q.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *aiRepository) UpdateModelFields(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.Model{}).Where("id = ?", id).Updates(fields).Error
}

// ===== 提示词模板 =====

func (r *aiRepository) CreatePrompt(p *model.Prompt) error {
	return r.db.Create(p).Error
}

func (r *aiRepository) FindPromptByName(name string) (*model.Prompt, error) {
	var p model.Prompt
	if err := r.db.Where("template_name = ?", name).First(&p).Error; err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *aiRepository) FindActivePromptByType(templateType string) (*model.Prompt, error) {
	var p model.Prompt
	if err := r.db.Where("template_type = ? AND status = ?", templateType, 1).
		Order("id ASC").First(&p).Error; err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *aiRepository) ListPrompts(templateType string, page, pageSize int) ([]model.Prompt, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	var list []model.Prompt
	var total int64
	q := r.db.Model(&model.Prompt{})
	if templateType != "" {
		q = q.Where("template_type = ?", templateType)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := q.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *aiRepository) UpdatePromptFields(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.Prompt{}).Where("id = ?", id).Updates(fields).Error
}

// ===== 生成记录 =====

func (r *aiRepository) CreateGeneration(g *model.Generation) error {
	return r.db.Create(g).Error
}

func (r *aiRepository) FindGenerationByID(id uint) (*model.Generation, error) {
	var g model.Generation
	if err := r.db.First(&g, id).Error; err != nil {
		return nil, err
	}
	return &g, nil
}

func (r *aiRepository) ListGenerationsByUser(userID uint, page, pageSize int) ([]model.Generation, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	var list []model.Generation
	var total int64
	q := r.db.Model(&model.Generation{}).Where("user_id = ?", userID)
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := q.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *aiRepository) ListGenerationsByTask(taskID string) ([]model.Generation, error) {
	var list []model.Generation
	if err := r.db.Where("task_id = ?", taskID).Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *aiRepository) UpdateGenerationFields(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.Generation{}).Where("id = ?", id).Updates(fields).Error
}
