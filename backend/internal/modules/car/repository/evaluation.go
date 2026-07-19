// Package repository 同城车辆买卖数据访问层 - 车辆评估
// 依据需求文档 1.10：4 维数据隔离（region_id + user_id）
package repository

import (
	"wuchang-tongcheng/internal/modules/car/model"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

// EvaluationRepository 车辆评估仓储接口
type EvaluationRepository interface {
	Create(e *model.CarEvaluation) error
	FindByID(id uint) (*model.CarEvaluation, error)
	FindByEvaluationNo(no string) (*model.CarEvaluation, error)
	FindByCarID(carID uint) (*model.CarEvaluation, error)
	Update(id uint, fields map[string]interface{}) error
	Delete(id uint) error

	// 列表（C 端，地区隔离）
	List(regionID uint, pagination *utils.Pagination, opts EvaluationListOptions) ([]model.CarEvaluation, int64, error)
	// 管理后台列表（M 端，可跨地区）
	AdminList(pagination *utils.Pagination, opts EvaluationAdminListOptions) ([]model.CarEvaluation, int64, error)
	// 评估师自己的评估单
	ListByEvaluator(evaluatorID uint, pagination *utils.Pagination) ([]model.CarEvaluation, int64, error)
	// 车辆历史评估
	ListByCarID(carID uint) ([]model.CarEvaluation, error)

	// 统计
	CountByStatus(regionID uint, status int) (int64, error)
	AvgPriceByModelID(modelID uint) (float64, error)
}

// EvaluationListOptions C 端评估列表过滤条件
type EvaluationListOptions struct {
	CarID          uint
	ModelID        uint
	EvaluatorID    uint
	EvaluationType string
	Status         *int
	Keyword        string
}

// EvaluationAdminListOptions M 端评估列表过滤条件
type EvaluationAdminListOptions struct {
	RegionID       uint
	CarID          uint
	EvaluatorID    uint
	EvaluationType string
	Status         *int
	Keyword        string
}

type evaluationRepository struct {
	db *gorm.DB
}

// NewEvaluationRepository 创建车辆评估仓储实例
func NewEvaluationRepository(db *gorm.DB) EvaluationRepository {
	return &evaluationRepository{db: db}
}

// ===== CRUD =====

func (r *evaluationRepository) Create(e *model.CarEvaluation) error {
	return r.db.Create(e).Error
}

func (r *evaluationRepository) FindByID(id uint) (*model.CarEvaluation, error) {
	var e model.CarEvaluation
	if err := r.db.First(&e, id).Error; err != nil {
		return nil, err
	}
	return &e, nil
}

func (r *evaluationRepository) FindByEvaluationNo(no string) (*model.CarEvaluation, error) {
	var e model.CarEvaluation
	if err := r.db.Where("evaluation_no = ?", no).First(&e).Error; err != nil {
		return nil, err
	}
	return &e, nil
}

func (r *evaluationRepository) FindByCarID(carID uint) (*model.CarEvaluation, error) {
	var e model.CarEvaluation
	if err := r.db.Where("car_id = ?", carID).Order("id DESC").First(&e).Error; err != nil {
		return nil, err
	}
	return &e, nil
}

func (r *evaluationRepository) Update(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.CarEvaluation{}).Where("id = ?", id).Updates(fields).Error
}

func (r *evaluationRepository) Delete(id uint) error {
	return r.db.Delete(&model.CarEvaluation{}, id).Error
}

// ===== 列表查询 =====

func (r *evaluationRepository) List(regionID uint, pagination *utils.Pagination, opts EvaluationListOptions) ([]model.CarEvaluation, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.CarEvaluation
	var total int64

	query := r.db.Model(&model.CarEvaluation{})

	if regionID > 0 {
		query = query.Where("region_id = ?", regionID)
	}
	if opts.CarID > 0 {
		query = query.Where("car_id = ?", opts.CarID)
	}
	if opts.ModelID > 0 {
		query = query.Where("model_id = ?", opts.ModelID)
	}
	if opts.EvaluatorID > 0 {
		query = query.Where("evaluator_id = ?", opts.EvaluatorID)
	}
	if opts.EvaluationType != "" {
		query = query.Where("evaluation_type = ?", opts.EvaluationType)
	}
	if opts.Status != nil {
		query = query.Where("status = ?", *opts.Status)
	}
	if opts.Keyword != "" {
		like := "%" + opts.Keyword + "%"
		query = query.Where("evaluation_no ILIKE ? OR evaluator_name ILIKE ?", like, like)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(pagination)).
		Order("created_at DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *evaluationRepository) AdminList(pagination *utils.Pagination, opts EvaluationAdminListOptions) ([]model.CarEvaluation, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.CarEvaluation
	var total int64

	query := r.db.Model(&model.CarEvaluation{})

	if opts.RegionID > 0 {
		query = query.Where("region_id = ?", opts.RegionID)
	}
	if opts.CarID > 0 {
		query = query.Where("car_id = ?", opts.CarID)
	}
	if opts.EvaluatorID > 0 {
		query = query.Where("evaluator_id = ?", opts.EvaluatorID)
	}
	if opts.EvaluationType != "" {
		query = query.Where("evaluation_type = ?", opts.EvaluationType)
	}
	if opts.Status != nil {
		query = query.Where("status = ?", *opts.Status)
	}
	if opts.Keyword != "" {
		like := "%" + opts.Keyword + "%"
		query = query.Where("evaluation_no ILIKE ? OR evaluator_name ILIKE ?", like, like)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(pagination)).
		Order("created_at DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *evaluationRepository) ListByEvaluator(evaluatorID uint, pagination *utils.Pagination) ([]model.CarEvaluation, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.CarEvaluation
	var total int64

	query := r.db.Model(&model.CarEvaluation{}).Where("evaluator_id = ?", evaluatorID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(pagination)).
		Order("created_at DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *evaluationRepository) ListByCarID(carID uint) ([]model.CarEvaluation, error) {
	var list []model.CarEvaluation
	if err := r.db.Where("car_id = ?", carID).Order("id DESC").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

// ===== 统计 =====

func (r *evaluationRepository) CountByStatus(regionID uint, status int) (int64, error) {
	var count int64
	q := r.db.Model(&model.CarEvaluation{}).Where("status = ?", status)
	if regionID > 0 {
		q = q.Where("region_id = ?", regionID)
	}
	if err := q.Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

// AvgPriceByModelID 查询某车型的均价（按 final_price 取平均，仅完成状态）
func (r *evaluationRepository) AvgPriceByModelID(modelID uint) (float64, error) {
	var avg float64
	if err := r.db.Model(&model.CarEvaluation{}).
		Where("model_id = ? AND status = ?", modelID, model.EvaluationStatusCompleted).
		Select("COALESCE(AVG(final_price), 0)").
		Scan(&avg).Error; err != nil {
		return 0, err
	}
	return avg, nil
}
