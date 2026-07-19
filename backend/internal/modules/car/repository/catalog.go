// Package repository 数据统计 + 车源图片数据访问层
// 依据 v3.2.1 架构方案：对标瓜子/懂车帝
package repository

import (
	"time"

	"wuchang-tongcheng/internal/modules/car/model"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

// ===== Statistic 数据统计 =====

// StatisticListOptions 统计列表过滤条件
type StatisticListOptions struct {
	RegionID    uint
	StatType    string // car/dealer/category/brand/region/platform
	TargetID    uint
	TargetName  string
	StartDate   *time.Time
	EndDate     *time.Time
	OrderByDesc string // 默认 stat_date DESC
}

// StatisticRepository 统计仓储接口
type StatisticRepository interface {
	Create(s *model.CarStatistic) error
	BatchCreate(stats []model.CarStatistic) error
	FindByID(id uint) (*model.CarStatistic, error)
	FindByDateTypeTarget(statDate time.Time, statType string, targetID uint) (*model.CarStatistic, error)
	Update(id uint, fields map[string]interface{}) error
	UpsertByDateTypeTarget(s *model.CarStatistic) error
	List(opts StatisticListOptions, pagination *utils.Pagination) ([]model.CarStatistic, int64, error)
	ListByType(statType string, startDate, endDate *time.Time, pagination *utils.Pagination) ([]model.CarStatistic, int64, error)
	ListByTarget(statType string, targetID uint, startDate, endDate *time.Time) ([]model.CarStatistic, error)
	ListByRegion(regionID uint, statType string, startDate, endDate *time.Time) ([]model.CarStatistic, error)
	ListTopByType(statType string, limit int, startDate, endDate *time.Time) ([]model.CarStatistic, error)
	SumByType(statType string, startDate, endDate *time.Time) (*model.CarStatistic, error)
	DeleteByDate(before time.Time) error
}

type statisticRepository struct {
	db *gorm.DB
}

// NewStatisticRepository 创建统计仓储实例
func NewStatisticRepository(db *gorm.DB) StatisticRepository {
	return &statisticRepository{db: db}
}

func (r *statisticRepository) Create(s *model.CarStatistic) error {
	return r.db.Create(s).Error
}

func (r *statisticRepository) BatchCreate(stats []model.CarStatistic) error {
	if len(stats) == 0 {
		return nil
	}
	return r.db.CreateInBatches(stats, 100).Error
}

func (r *statisticRepository) FindByID(id uint) (*model.CarStatistic, error) {
	var s model.CarStatistic
	if err := r.db.First(&s, id).Error; err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *statisticRepository) FindByDateTypeTarget(statDate time.Time, statType string, targetID uint) (*model.CarStatistic, error) {
	var s model.CarStatistic
	if err := r.db.Where("stat_date = ? AND stat_type = ? AND target_id = ?",
		statDate, statType, targetID).First(&s).Error; err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *statisticRepository) Update(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.CarStatistic{}).Where("id = ?", id).Updates(fields).Error
}

// UpsertByDateTypeTarget 按日期+类型+目标 ID 唯一键 upsert
// PostgreSQL ON CONFLICT 需要 (stat_date, stat_type, target_id) 唯一约束
func (r *statisticRepository) UpsertByDateTypeTarget(s *model.CarStatistic) error {
	return r.db.Where("stat_date = ? AND stat_type = ? AND target_id = ?",
		s.StatDate, s.StatType, s.TargetID).
		Assign(s).
		FirstOrCreate(s).Error
}

func (r *statisticRepository) List(opts StatisticListOptions, pagination *utils.Pagination) ([]model.CarStatistic, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.CarStatistic
	var total int64

	q := r.db.Model(&model.CarStatistic{})
	if opts.RegionID > 0 {
		q = q.Where("region_id = ?", opts.RegionID)
	}
	if opts.StatType != "" {
		q = q.Where("stat_type = ?", opts.StatType)
	}
	if opts.TargetID > 0 {
		q = q.Where("target_id = ?", opts.TargetID)
	}
	if opts.TargetName != "" {
		q = q.Where("target_name LIKE ?", "%"+opts.TargetName+"%")
	}
	if opts.StartDate != nil {
		q = q.Where("stat_date >= ?", *opts.StartDate)
	}
	if opts.EndDate != nil {
		q = q.Where("stat_date <= ?", *opts.EndDate)
	}

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	orderField := "stat_date DESC, id DESC"
	if opts.OrderByDesc != "" {
		orderField = opts.OrderByDesc
	}
	if err := q.Scopes(utils.Paginate(pagination)).
		Order(orderField).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *statisticRepository) ListByType(statType string, startDate, endDate *time.Time, pagination *utils.Pagination) ([]model.CarStatistic, int64, error) {
	return r.List(StatisticListOptions{
		StatType:  statType,
		StartDate: startDate,
		EndDate:   endDate,
	}, pagination)
}

func (r *statisticRepository) ListByTarget(statType string, targetID uint, startDate, endDate *time.Time) ([]model.CarStatistic, error) {
	q := r.db.Model(&model.CarStatistic{}).
		Where("stat_type = ? AND target_id = ?", statType, targetID)
	if startDate != nil {
		q = q.Where("stat_date >= ?", *startDate)
	}
	if endDate != nil {
		q = q.Where("stat_date <= ?", *endDate)
	}
	var list []model.CarStatistic
	if err := q.Order("stat_date DESC").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *statisticRepository) ListByRegion(regionID uint, statType string, startDate, endDate *time.Time) ([]model.CarStatistic, error) {
	q := r.db.Model(&model.CarStatistic{}).
		Where("region_id = ? AND stat_type = ?", regionID, statType)
	if startDate != nil {
		q = q.Where("stat_date >= ?", *startDate)
	}
	if endDate != nil {
		q = q.Where("stat_date <= ?", *endDate)
	}
	var list []model.CarStatistic
	if err := q.Order("stat_date DESC").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

// ListTopByType 按某类型统计取 Top N（默认按成交数 deal_count 降序）
func (r *statisticRepository) ListTopByType(statType string, limit int, startDate, endDate *time.Time) ([]model.CarStatistic, error) {
	if limit <= 0 || limit > 200 {
		limit = 20
	}
	q := r.db.Model(&model.CarStatistic{}).Where("stat_type = ?", statType)
	if startDate != nil {
		q = q.Where("stat_date >= ?", *startDate)
	}
	if endDate != nil {
		q = q.Where("stat_date <= ?", *endDate)
	}
	var list []model.CarStatistic
	if err := q.Order("deal_count DESC, click_count DESC, id DESC").Limit(limit).Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

// SumByType 按类型聚合求和（用于平台总览）
func (r *statisticRepository) SumByType(statType string, startDate, endDate *time.Time) (*model.CarStatistic, error) {
	q := r.db.Model(&model.CarStatistic{}).Where("stat_type = ?", statType)
	if startDate != nil {
		q = q.Where("stat_date >= ?", *startDate)
	}
	if endDate != nil {
		q = q.Where("stat_date <= ?", *endDate)
	}
	var s model.CarStatistic
	err := q.Select("COALESCE(SUM(impression_count),0) AS impression_count, " +
		"COALESCE(SUM(click_count),0) AS click_count, " +
		"COALESCE(SUM(fav_count),0) AS fav_count, " +
		"COALESCE(SUM(contact_count),0) AS contact_count, " +
		"COALESCE(SUM(test_drive_count),0) AS test_drive_count, " +
		"COALESCE(SUM(deal_count),0) AS deal_count, " +
		"COALESCE(AVG(avg_price),0) AS avg_price, " +
		"COALESCE(AVG(avg_deal_days),0) AS avg_deal_days").
		Scan(&s).Error
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *statisticRepository) DeleteByDate(before time.Time) error {
	return r.db.Where("stat_date < ?", before).Delete(&model.CarStatistic{}).Error
}

// ===== Image 车源图片 =====

// ImageListOptions 图片列表过滤条件
type ImageListOptions struct {
	CarID      uint
	ListingID  uint
	ImageType  string // exterior/interior/engine/chassis/accident/document/dashboard/wheel/trunk/other
	IsCover    *bool
	Tag        string
	Width      int
	Height     int
	MinSize    int
	MaxSize    int
}

// ImageRepository 车源图片仓储接口
type ImageRepository interface {
	Create(img *model.CarImage) error
	BatchCreate(imgs []model.CarImage) error
	FindByID(id uint) (*model.CarImage, error)
	Update(id uint, fields map[string]interface{}) error
	Delete(id uint) error
	DeleteByCarID(carID uint) error
	List(opts ImageListOptions, pagination *utils.Pagination) ([]model.CarImage, int64, error)
	ListByCarID(carID uint) ([]model.CarImage, error)
	ListByListingID(listingID uint) ([]model.CarImage, error)
	ListByCarIDAndType(carID uint, imageType string) ([]model.CarImage, error)
	GetCoverByCarID(carID uint) (*model.CarImage, error)
	SetCover(carID uint, imageID uint) error
	UpdateSort(id uint, sort int) error
	UpdateSortBatch(items []ImageSortItem) error
	CountByCarID(carID uint) (int64, error)
	CountByCarIDAndType(carID uint, imageType string) (int64, error)
}

// ImageSortItem 图片排序更新项
type ImageSortItem struct {
	ID   uint `json:"id"`
	Sort int  `json:"sort"`
}

type imageRepository struct {
	db *gorm.DB
}

// NewImageRepository 创建车源图片仓储实例
func NewImageRepository(db *gorm.DB) ImageRepository {
	return &imageRepository{db: db}
}

func (r *imageRepository) Create(img *model.CarImage) error {
	return r.db.Create(img).Error
}

func (r *imageRepository) BatchCreate(imgs []model.CarImage) error {
	if len(imgs) == 0 {
		return nil
	}
	return r.db.CreateInBatches(imgs, 100).Error
}

func (r *imageRepository) FindByID(id uint) (*model.CarImage, error) {
	var img model.CarImage
	if err := r.db.First(&img, id).Error; err != nil {
		return nil, err
	}
	return &img, nil
}

func (r *imageRepository) Update(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.CarImage{}).Where("id = ?", id).Updates(fields).Error
}

func (r *imageRepository) Delete(id uint) error {
	return r.db.Delete(&model.CarImage{}, id).Error
}

func (r *imageRepository) DeleteByCarID(carID uint) error {
	return r.db.Where("car_id = ?", carID).Delete(&model.CarImage{}).Error
}

func (r *imageRepository) List(opts ImageListOptions, pagination *utils.Pagination) ([]model.CarImage, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 20)
	}
	var list []model.CarImage
	var total int64

	q := r.db.Model(&model.CarImage{})
	if opts.CarID > 0 {
		q = q.Where("car_id = ?", opts.CarID)
	}
	if opts.ListingID > 0 {
		q = q.Where("listing_id = ?", opts.ListingID)
	}
	if opts.ImageType != "" {
		q = q.Where("image_type = ?", opts.ImageType)
	}
	if opts.IsCover != nil {
		q = q.Where("is_cover = ?", *opts.IsCover)
	}
	if opts.Tag != "" {
		q = q.Where("tag = ?", opts.Tag)
	}
	if opts.Width > 0 {
		q = q.Where("width = ?", opts.Width)
	}
	if opts.Height > 0 {
		q = q.Where("height = ?", opts.Height)
	}
	if opts.MinSize > 0 {
		q = q.Where("size >= ?", opts.MinSize)
	}
	if opts.MaxSize > 0 {
		q = q.Where("size <= ?", opts.MaxSize)
	}

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := q.Scopes(utils.Paginate(pagination)).
		Order("sort ASC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *imageRepository) ListByCarID(carID uint) ([]model.CarImage, error) {
	var list []model.CarImage
	if err := r.db.Where("car_id = ?", carID).Order("sort ASC, id DESC").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *imageRepository) ListByListingID(listingID uint) ([]model.CarImage, error) {
	var list []model.CarImage
	if err := r.db.Where("listing_id = ?", listingID).Order("sort ASC, id DESC").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *imageRepository) ListByCarIDAndType(carID uint, imageType string) ([]model.CarImage, error) {
	var list []model.CarImage
	if err := r.db.Where("car_id = ? AND image_type = ?", carID, imageType).
		Order("sort ASC, id DESC").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *imageRepository) GetCoverByCarID(carID uint) (*model.CarImage, error) {
	var img model.CarImage
	// 优先取 is_cover = true 的；若没有则取排序第一张
	if err := r.db.Where("car_id = ? AND is_cover = ?", carID, true).First(&img).Error; err == nil {
		return &img, nil
	}
	if err := r.db.Where("car_id = ?", carID).Order("sort ASC, id DESC").First(&img).Error; err != nil {
		return nil, err
	}
	return &img, nil
}

// SetCover 设置指定图片为封面，同时取消该车其他图片的封面标记（事务）
func (r *imageRepository) SetCover(carID uint, imageID uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// 1. 取消该车所有图片的封面标记
		if err := tx.Model(&model.CarImage{}).
			Where("car_id = ? AND is_cover = ?", carID, true).
			Update("is_cover", false).Error; err != nil {
			return err
		}
		// 2. 设置目标图片为封面
		return tx.Model(&model.CarImage{}).
			Where("id = ? AND car_id = ?", imageID, carID).
			Update("is_cover", true).Error
	})
}

func (r *imageRepository) UpdateSort(id uint, sort int) error {
	return r.db.Model(&model.CarImage{}).Where("id = ?", id).Update("sort", sort).Error
}

func (r *imageRepository) UpdateSortBatch(items []ImageSortItem) error {
	if len(items) == 0 {
		return nil
	}
	return r.db.Transaction(func(tx *gorm.DB) error {
		for _, it := range items {
			if err := tx.Model(&model.CarImage{}).
				Where("id = ?", it.ID).
				Update("sort", it.Sort).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *imageRepository) CountByCarID(carID uint) (int64, error) {
	var count int64
	if err := r.db.Model(&model.CarImage{}).Where("car_id = ?", carID).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (r *imageRepository) CountByCarIDAndType(carID uint, imageType string) (int64, error) {
	var count int64
	if err := r.db.Model(&model.CarImage{}).
		Where("car_id = ? AND image_type = ?", carID, imageType).
		Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}
