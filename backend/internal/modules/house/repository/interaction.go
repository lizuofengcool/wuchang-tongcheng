// Package repository 互动数据访问层（VR + 推荐 + 浏览记录通用方法）
package repository

import (
	"wuchang-tongcheng/internal/modules/house/model"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

// InteractionRepository 互动仓储接口（VR + 推荐 + 通用浏览记录）
type InteractionRepository interface {
	// VR 看房
	CreateVR(v *model.HouseVRTour) error
	FindVRByID(id uint) (*model.HouseVRTour, error)
	FindVRByVRNo(no string) (*model.HouseVRTour, error)
	UpdateVR(v *model.HouseVRTour) error
	UpdateVRFields(id uint, fields map[string]interface{}) error
	DeleteVR(id uint) error
	ListVR(req *utils.Pagination, opts VRListOptions) ([]model.HouseVRTour, int64, error)
	IncrVRViewCount(id uint) error
	IncrVRShareCount(id uint) error

	// 推荐
	CreateRec(r *model.HouseRecommendation) error
	FindRecByID(id uint) (*model.HouseRecommendation, error)
	UpdateRec(r *model.HouseRecommendation) error
	UpdateRecFields(id uint, fields map[string]interface{}) error
	DeleteRec(id uint) error
	ListRec(req *utils.Pagination, opts RecListOptions) ([]model.HouseRecommendation, int64, error)
	ListRecByUser(userID uint, req *utils.Pagination) ([]model.HouseRecommendation, int64, error)
}

// VRListOptions VR 列表过滤条件
type VRListOptions struct {
	HouseID     uint
	ListingID   uint
	CommunityID uint
	VRType      string
	Status      *int
}

// RecListOptions 推荐列表过滤条件
type RecListOptions struct {
	UserID  uint
	RecType string
	Source  string
	Status  *int
}

type interactionRepository struct {
	db *gorm.DB
}

// NewInteractionRepository 创建仓储实例
func NewInteractionRepository(db *gorm.DB) InteractionRepository {
	return &interactionRepository{db: db}
}

// ===== VR 看房 =====

func (r *interactionRepository) CreateVR(v *model.HouseVRTour) error {
	return r.db.Create(v).Error
}

func (r *interactionRepository) FindVRByID(id uint) (*model.HouseVRTour, error) {
	var v model.HouseVRTour
	if err := r.db.First(&v, id).Error; err != nil {
		return nil, err
	}
	return &v, nil
}

func (r *interactionRepository) FindVRByVRNo(no string) (*model.HouseVRTour, error) {
	var v model.HouseVRTour
	if err := r.db.Where("vr_no = ?", no).First(&v).Error; err != nil {
		return nil, err
	}
	return &v, nil
}

func (r *interactionRepository) UpdateVR(v *model.HouseVRTour) error {
	return r.db.Save(v).Error
}

func (r *interactionRepository) UpdateVRFields(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.HouseVRTour{}).Where("id = ?", id).Updates(fields).Error
}

func (r *interactionRepository) DeleteVR(id uint) error {
	return r.db.Delete(&model.HouseVRTour{}, id).Error
}

func (r *interactionRepository) ListVR(req *utils.Pagination, opts VRListOptions) ([]model.HouseVRTour, int64, error) {
	if req == nil {
		req = utils.NewPagination(1, 10)
	}
	var list []model.HouseVRTour
	var total int64

	query := r.db.Model(&model.HouseVRTour{})
	if opts.HouseID > 0 {
		query = query.Where("house_id = ?", opts.HouseID)
	}
	if opts.ListingID > 0 {
		query = query.Where("listing_id = ?", opts.ListingID)
	}
	if opts.CommunityID > 0 {
		query = query.Where("community_id = ?", opts.CommunityID)
	}
	if opts.VRType != "" {
		query = query.Where("vr_type = ?", opts.VRType)
	}
	if opts.Status != nil {
		query = query.Where("status = ?", *opts.Status)
	} else {
		query = query.Where("status = ?", model.VRStatusPublished)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(req)).Order("published_at DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *interactionRepository) IncrVRViewCount(id uint) error {
	return r.db.Model(&model.HouseVRTour{}).Where("id = ?", id).
		UpdateColumn("view_count", gorm.Expr("view_count + 1")).Error
}

func (r *interactionRepository) IncrVRShareCount(id uint) error {
	return r.db.Model(&model.HouseVRTour{}).Where("id = ?", id).
		UpdateColumn("share_count", gorm.Expr("share_count + 1")).Error
}

// ===== 推荐 =====

func (r *interactionRepository) CreateRec(rec *model.HouseRecommendation) error {
	return r.db.Create(rec).Error
}

func (r *interactionRepository) FindRecByID(id uint) (*model.HouseRecommendation, error) {
	var rec model.HouseRecommendation
	if err := r.db.First(&rec, id).Error; err != nil {
		return nil, err
	}
	return &rec, nil
}

func (r *interactionRepository) UpdateRec(rec *model.HouseRecommendation) error {
	return r.db.Save(rec).Error
}

func (r *interactionRepository) UpdateRecFields(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.HouseRecommendation{}).Where("id = ?", id).Updates(fields).Error
}

func (r *interactionRepository) DeleteRec(id uint) error {
	return r.db.Delete(&model.HouseRecommendation{}, id).Error
}

func (r *interactionRepository) ListRec(req *utils.Pagination, opts RecListOptions) ([]model.HouseRecommendation, int64, error) {
	if req == nil {
		req = utils.NewPagination(1, 10)
	}
	var list []model.HouseRecommendation
	var total int64

	query := r.db.Model(&model.HouseRecommendation{})
	if opts.UserID > 0 {
		query = query.Where("user_id = ?", opts.UserID)
	}
	if opts.RecType != "" {
		query = query.Where("rec_type = ?", opts.RecType)
	}
	if opts.Source != "" {
		query = query.Where("source = ?", opts.Source)
	}
	if opts.Status != nil {
		query = query.Where("status = ?", *opts.Status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(req)).Order("score DESC, created_at DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *interactionRepository) ListRecByUser(userID uint, req *utils.Pagination) ([]model.HouseRecommendation, int64, error) {
	if req == nil {
		req = utils.NewPagination(1, 10)
	}
	var list []model.HouseRecommendation
	var total int64

	query := r.db.Model(&model.HouseRecommendation{}).Where("user_id = ?", userID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(req)).Order("score DESC, created_at DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}
