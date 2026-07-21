// Package repository 分销合伙人中台数据访问层 - 推广渠道
package repository

import (
	"wuchang-tongcheng/internal/modules/distribution/model"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

// ChannelRepository 渠道仓储接口
type ChannelRepository interface {
	Create(c *model.Channel) error
	FindByID(id uint) (*model.Channel, error)
	FindByCode(code string) (*model.Channel, error)
	Update(c *model.Channel) error
	UpdateFields(id uint, fields map[string]interface{}) error
	Delete(id uint) error

	List(pagination *utils.Pagination, opts ChannelListOptions) ([]model.Channel, int64, error)
	ListByPartner(partnerID uint) ([]model.Channel, error)

	// 统计
	IncrClick(id uint) error
	IncrRegister(id uint) error
	IncrOrder(id uint) error
	AddCommission(id uint, amount float64) error
	StatsByPartner(partnerID uint) (totalChannels, totalClicks, totalRegisters, totalOrders int, totalCommission float64, err error)
}

// ChannelListOptions 渠道列表过滤条件
type ChannelListOptions struct {
	PartnerID uint
	Code      string
	Keyword   string
}

type channelRepository struct {
	db *gorm.DB
}

// NewChannelRepository 创建渠道仓储实例
func NewChannelRepository(db *gorm.DB) ChannelRepository {
	return &channelRepository{db: db}
}

func (r *channelRepository) Create(c *model.Channel) error {
	return r.db.Create(c).Error
}

func (r *channelRepository) FindByID(id uint) (*model.Channel, error) {
	var c model.Channel
	if err := r.db.First(&c, id).Error; err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *channelRepository) FindByCode(code string) (*model.Channel, error) {
	var c model.Channel
	if err := r.db.Where("code = ?", code).First(&c).Error; err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *channelRepository) Update(c *model.Channel) error {
	return r.db.Save(c).Error
}

func (r *channelRepository) UpdateFields(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.Channel{}).Where("id = ?", id).Updates(fields).Error
}

func (r *channelRepository) Delete(id uint) error {
	return r.db.Delete(&model.Channel{}, id).Error
}

func (r *channelRepository) List(pagination *utils.Pagination, opts ChannelListOptions) ([]model.Channel, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.Channel
	var total int64

	query := r.db.Model(&model.Channel{})
	if opts.PartnerID > 0 {
		query = query.Where("partner_id = ?", opts.PartnerID)
	}
	if opts.Code != "" {
		query = query.Where("code = ?", opts.Code)
	}
	if opts.Keyword != "" {
		like := "%" + opts.Keyword + "%"
		query = query.Where("name ILIKE ? OR code ILIKE ?", like, like)
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

func (r *channelRepository) ListByPartner(partnerID uint) ([]model.Channel, error) {
	var list []model.Channel
	if err := r.db.Where("partner_id = ?", partnerID).
		Order("created_at DESC, id DESC").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *channelRepository) IncrClick(id uint) error {
	return r.db.Model(&model.Channel{}).Where("id = ?", id).
		UpdateColumn("click_count", gorm.Expr("click_count + 1")).Error
}

func (r *channelRepository) IncrRegister(id uint) error {
	return r.db.Model(&model.Channel{}).Where("id = ?", id).
		UpdateColumn("register_count", gorm.Expr("register_count + 1")).Error
}

func (r *channelRepository) IncrOrder(id uint) error {
	return r.db.Model(&model.Channel{}).Where("id = ?", id).
		UpdateColumn("order_count", gorm.Expr("order_count + 1")).Error
}

func (r *channelRepository) AddCommission(id uint, amount float64) error {
	return r.db.Model(&model.Channel{}).Where("id = ?", id).
		UpdateColumn("commission_amount", gorm.Expr("commission_amount + ?", amount)).Error
}

func (r *channelRepository) StatsByPartner(partnerID uint) (totalChannels, totalClicks, totalRegisters, totalOrders int, totalCommission float64, err error) {
	type statsResult struct {
		TotalChannels    int     `gorm:"column:total_channels"`
		TotalClicks      int     `gorm:"column:total_clicks"`
		TotalRegisters   int     `gorm:"column:total_registers"`
		TotalOrders      int     `gorm:"column:total_orders"`
		TotalCommission  float64 `gorm:"column:total_commission"`
	}
	var stats statsResult
	query := r.db.Model(&model.Channel{})
	if partnerID > 0 {
		query = query.Where("partner_id = ?", partnerID)
	}
	if err := query.Select(
		"COUNT(*) AS total_channels, " +
			"COALESCE(SUM(click_count),0) AS total_clicks, " +
			"COALESCE(SUM(register_count),0) AS total_registers, " +
			"COALESCE(SUM(order_count),0) AS total_orders, " +
			"COALESCE(SUM(commission_amount),0) AS total_commission",
	).Scan(&stats).Error; err != nil {
		return 0, 0, 0, 0, 0, err
	}
	return stats.TotalChannels, stats.TotalClicks, stats.TotalRegisters, stats.TotalOrders, stats.TotalCommission, nil
}
