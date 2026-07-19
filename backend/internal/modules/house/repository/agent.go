// Package repository 经纪人数据访问层
package repository

import (
	"wuchang-tongcheng/internal/modules/house/model"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

// AgentRepository 经纪人仓储接口
type AgentRepository interface {
	Create(a *model.HouseAgent) error
	FindByID(id uint) (*model.HouseAgent, error)
	FindByUserID(userID uint) (*model.HouseAgent, error)
	FindByPhone(phone string) (*model.HouseAgent, error)
	Update(a *model.HouseAgent) error
	UpdateFields(id uint, fields map[string]interface{}) error
	Delete(id uint) error

	List(regionID uint, req *utils.Pagination, opts AgentListOptions) ([]model.HouseAgent, int64, error)
	AdminList(req *utils.Pagination, opts AgentAdminListOptions) ([]model.HouseAgent, int64, error)

	// 关注
	FollowExists(userID, agentID uint) (bool, error)
	CreateFollow(fav *model.HouseFavorite) error
	DeleteFollow(userID, agentID uint) error
	IncrFollowerCount(id uint) error
	DecrFollowerCount(id uint) error

	// 统计
	IncrListingCount(id uint) error
	DecrListingCount(id uint) error
	IncrDealCount(id uint) error
	UpdateRating(id uint, rating float64, ratingCount int) error
	UpdateLastActiveAt(id uint) error

	// 审核
	BatchAudit(ids []uint, status int, reason string) (int64, error)
}

// AgentListOptions C 端列表过滤条件
type AgentListOptions struct {
	City         string
	Company      string
	StoreID      uint
	Level        *int
	Status       *int
	OnlineStatus *int
	Keyword      string
	Sort         string // latest/rating/deal_count/listing_count
}

// AgentAdminListOptions M 端管理列表过滤条件
type AgentAdminListOptions struct {
	RegionID uint
	UserID   uint
	Status   *int
	Level    *int
	Keyword  string
}

type agentRepository struct {
	db *gorm.DB
}

// NewAgentRepository 创建仓储实例
func NewAgentRepository(db *gorm.DB) AgentRepository {
	return &agentRepository{db: db}
}

func (r *agentRepository) Create(a *model.HouseAgent) error {
	return r.db.Create(a).Error
}

func (r *agentRepository) FindByID(id uint) (*model.HouseAgent, error) {
	var a model.HouseAgent
	if err := r.db.First(&a, id).Error; err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *agentRepository) FindByUserID(userID uint) (*model.HouseAgent, error) {
	var a model.HouseAgent
	if err := r.db.Where("user_id = ?", userID).First(&a).Error; err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *agentRepository) FindByPhone(phone string) (*model.HouseAgent, error) {
	var a model.HouseAgent
	if err := r.db.Where("phone = ?", phone).First(&a).Error; err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *agentRepository) Update(a *model.HouseAgent) error {
	return r.db.Save(a).Error
}

func (r *agentRepository) UpdateFields(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.HouseAgent{}).Where("id = ?", id).Updates(fields).Error
}

func (r *agentRepository) Delete(id uint) error {
	return r.db.Delete(&model.HouseAgent{}, id).Error
}

func (r *agentRepository) List(regionID uint, req *utils.Pagination, opts AgentListOptions) ([]model.HouseAgent, int64, error) {
	if req == nil {
		req = utils.NewPagination(1, 10)
	}
	var list []model.HouseAgent
	var total int64

	query := r.db.Model(&model.HouseAgent{}).Where("status = ?", model.AgentStatusApproved)
	if regionID > 0 {
		query = query.Where("region_id = ?", regionID)
	}
	if opts.City != "" {
		query = query.Where("company ILIKE ?", "%"+opts.City+"%")
	}
	if opts.Company != "" {
		query = query.Where("company ILIKE ?", "%"+opts.Company+"%")
	}
	if opts.StoreID > 0 {
		query = query.Where("store_id = ?", opts.StoreID)
	}
	if opts.Level != nil {
		query = query.Where("level = ?", *opts.Level)
	}
	if opts.OnlineStatus != nil {
		query = query.Where("online_status = ?", *opts.OnlineStatus)
	}
	if opts.Keyword != "" {
		like := "%" + opts.Keyword + "%"
		query = query.Where("name ILIKE ? OR company ILIKE ? OR store_name ILIKE ?", like, like, like)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	orderClause := "created_at DESC, id DESC"
	switch opts.Sort {
	case "rating":
		orderClause = "rating DESC, id DESC"
	case "deal_count":
		orderClause = "deal_count DESC, id DESC"
	case "listing_count":
		orderClause = "listing_count DESC, id DESC"
	}
	if err := query.Scopes(utils.Paginate(req)).Order(orderClause).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *agentRepository) AdminList(req *utils.Pagination, opts AgentAdminListOptions) ([]model.HouseAgent, int64, error) {
	if req == nil {
		req = utils.NewPagination(1, 10)
	}
	var list []model.HouseAgent
	var total int64

	query := r.db.Model(&model.HouseAgent{})
	if opts.RegionID > 0 {
		query = query.Where("region_id = ?", opts.RegionID)
	}
	if opts.UserID > 0 {
		query = query.Where("user_id = ?", opts.UserID)
	}
	if opts.Status != nil {
		query = query.Where("status = ?", *opts.Status)
	}
	if opts.Level != nil {
		query = query.Where("level = ?", *opts.Level)
	}
	if opts.Keyword != "" {
		like := "%" + opts.Keyword + "%"
		query = query.Where("name ILIKE ? OR phone ILIKE ? OR company ILIKE ?", like, like, like)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(req)).Order("created_at DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// ===== 关注 =====

func (r *agentRepository) FollowExists(userID, agentID uint) (bool, error) {
	var count int64
	err := r.db.Model(&model.HouseFavorite{}).
		Where("user_id = ? AND agent_id = ? AND favorite_type = ?", userID, agentID, model.FavoriteTypeAgent).
		Count(&count).Error
	return count > 0, err
}

func (r *agentRepository) CreateFollow(fav *model.HouseFavorite) error {
	return r.db.Create(fav).Error
}

func (r *agentRepository) DeleteFollow(userID, agentID uint) error {
	return r.db.Where("user_id = ? AND agent_id = ? AND favorite_type = ?", userID, agentID, model.FavoriteTypeAgent).
		Delete(&model.HouseFavorite{}).Error
}

func (r *agentRepository) IncrFollowerCount(id uint) error {
	return r.db.Model(&model.HouseAgent{}).Where("id = ?", id).
		UpdateColumn("follower_count", gorm.Expr("follower_count + 1")).Error
}

func (r *agentRepository) DecrFollowerCount(id uint) error {
	return r.db.Model(&model.HouseAgent{}).Where("id = ? AND follower_count > 0", id).
		UpdateColumn("follower_count", gorm.Expr("follower_count - 1")).Error
}

// ===== 统计 =====

func (r *agentRepository) IncrListingCount(id uint) error {
	return r.db.Model(&model.HouseAgent{}).Where("id = ?", id).
		UpdateColumn("listing_count", gorm.Expr("listing_count + 1")).Error
}

func (r *agentRepository) DecrListingCount(id uint) error {
	return r.db.Model(&model.HouseAgent{}).Where("id = ? AND listing_count > 0", id).
		UpdateColumn("listing_count", gorm.Expr("listing_count - 1")).Error
}

func (r *agentRepository) IncrDealCount(id uint) error {
	return r.db.Model(&model.HouseAgent{}).Where("id = ?", id).
		UpdateColumn("deal_count", gorm.Expr("deal_count + 1")).Error
}

func (r *agentRepository) UpdateRating(id uint, rating float64, ratingCount int) error {
	return r.db.Model(&model.HouseAgent{}).Where("id = ?", id).
		Updates(map[string]interface{}{
			"rating":       rating,
			"rating_count": ratingCount,
		}).Error
}

func (r *agentRepository) UpdateLastActiveAt(id uint) error {
	return r.db.Model(&model.HouseAgent{}).Where("id = ?", id).
		Update("last_active_at", gorm.Expr("NOW()")).Error
}

// ===== 审核 =====

func (r *agentRepository) BatchAudit(ids []uint, status int, reason string) (int64, error) {
	fields := map[string]interface{}{
		"status": status,
	}
	if status == model.AgentStatusRejected {
		fields["rejected_reason"] = reason
	}
	result := r.db.Model(&model.HouseAgent{}).Where("id IN ?", ids).Updates(fields)
	return result.RowsAffected, result.Error
}
