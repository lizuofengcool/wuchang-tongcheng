// Package repository 素材中台扩展数据访问层
package repository

import (
	"wuchang-tongcheng/internal/modules/material/model"

	"gorm.io/gorm"
)

// MaterialExtendRepository 素材扩展仓储接口
type MaterialExtendRepository interface {
	// 分类
	CreateCategory(c *model.Category) error
	FindCategoryByID(id uint) (*model.Category, error)
	ListCategories(parentID uint, status int, page, pageSize int) ([]model.Category, int64, error)
	UpdateCategoryFields(id uint, fields map[string]interface{}) error
	DeleteCategory(id uint) error
	IncrCategoryImageCount(id uint) error

	// 标签
	CreateTag(t *model.Tag) error
	FindTagByName(name string) (*model.Tag, error)
	FindTagByID(id uint) (*model.Tag, error)
	ListTags(tagType string, page, pageSize int) ([]model.Tag, int64, error)
	UpdateTagFields(id uint, fields map[string]interface{}) error
	DeleteTag(id uint) error
	IncrTagUsageCount(id uint) error

	// 图片标签
	CreateImageTag(it *model.ImageTag) error
	DeleteImageTag(imageID, tagID uint) error
	ListImageTagsByImage(imageID uint) ([]model.ImageTag, error)
	ListImageTagsByTag(tagID uint, page, pageSize int) ([]model.ImageTag, int64, error)

	// 搜索历史
	CreateSearchHistory(h *model.SearchHistory) error
	ListSearchHistory(userID uint, page, pageSize int) ([]model.SearchHistory, int64, error)

	// 相似结果
	CreateSimilarResult(r *model.SimilarResult) error
	BatchCreateSimilarResults(list []model.SimilarResult) error
	ListSimilarResults(sourceImageID uint, limit int) ([]model.SimilarResult, error)

	// OCR 结果
	CreateOCRResult(r *model.OCRResult) error
	FindOCRResultByImageID(imageID uint) (*model.OCRResult, error)
	ListOCRResults(page, pageSize int) ([]model.OCRResult, int64, error)
	UpdateOCRResultFields(id uint, fields map[string]interface{}) error

	// 统计
	StatTotalFiles() (int64, error)
	StatTotalImages() (int64, error)
	StatTotalVideos() (int64, error)
	StatTotalCategories() (int64, error)
	StatTotalTags() (int64, error)
	StatTotalSearches() (int64, error)
	StatTotalOCR() (int64, error)
	StatStorageSize() (int64, error)
}

type materialExtendRepository struct {
	db *gorm.DB
}

// NewMaterialExtendRepository 创建扩展仓储实例
func NewMaterialExtendRepository(db *gorm.DB) MaterialExtendRepository {
	return &materialExtendRepository{db: db}
}

// ===== 分类 =====

func (r *materialExtendRepository) CreateCategory(c *model.Category) error {
	return r.db.Create(c).Error
}

func (r *materialExtendRepository) FindCategoryByID(id uint) (*model.Category, error) {
	var c model.Category
	if err := r.db.First(&c, id).Error; err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *materialExtendRepository) ListCategories(parentID uint, status int, page, pageSize int) ([]model.Category, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 50
	}
	var list []model.Category
	var total int64
	q := r.db.Model(&model.Category{})
	if parentID > 0 {
		q = q.Where("parent_id = ?", parentID)
	} else if parentID == 0 {
		// parentID=0 表示根分类；status=-1 表示不限制
		q = q.Where("parent_id = 0")
	}
	if status >= 0 {
		q = q.Where("status = ?", status)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := q.Order("sort ASC, id ASC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *materialExtendRepository) UpdateCategoryFields(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.Category{}).Where("id = ?", id).Updates(fields).Error
}

func (r *materialExtendRepository) DeleteCategory(id uint) error {
	return r.db.Delete(&model.Category{}, id).Error
}

func (r *materialExtendRepository) IncrCategoryImageCount(id uint) error {
	return r.db.Model(&model.Category{}).Where("id = ?", id).
		UpdateColumn("image_count", gorm.Expr("image_count + 1")).Error
}

// ===== 标签 =====

func (r *materialExtendRepository) CreateTag(t *model.Tag) error {
	return r.db.Create(t).Error
}

func (r *materialExtendRepository) FindTagByName(name string) (*model.Tag, error) {
	var t model.Tag
	if err := r.db.Where("name = ?", name).First(&t).Error; err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *materialExtendRepository) FindTagByID(id uint) (*model.Tag, error) {
	var t model.Tag
	if err := r.db.First(&t, id).Error; err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *materialExtendRepository) ListTags(tagType string, page, pageSize int) ([]model.Tag, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 50
	}
	var list []model.Tag
	var total int64
	q := r.db.Model(&model.Tag{}).Where("status = ?", 1)
	if tagType != "" {
		q = q.Where("tag_type = ?", tagType)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := q.Order("usage_count DESC, id ASC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *materialExtendRepository) UpdateTagFields(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.Tag{}).Where("id = ?", id).Updates(fields).Error
}

func (r *materialExtendRepository) DeleteTag(id uint) error {
	return r.db.Delete(&model.Tag{}, id).Error
}

func (r *materialExtendRepository) IncrTagUsageCount(id uint) error {
	return r.db.Model(&model.Tag{}).Where("id = ?", id).
		UpdateColumn("usage_count", gorm.Expr("usage_count + 1")).Error
}

// ===== 图片标签 =====

func (r *materialExtendRepository) CreateImageTag(it *model.ImageTag) error {
	return r.db.Create(it).Error
}

func (r *materialExtendRepository) DeleteImageTag(imageID, tagID uint) error {
	return r.db.Where("image_id = ? AND tag_id = ?", imageID, tagID).Delete(&model.ImageTag{}).Error
}

func (r *materialExtendRepository) ListImageTagsByImage(imageID uint) ([]model.ImageTag, error) {
	var list []model.ImageTag
	if err := r.db.Where("image_id = ?", imageID).Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *materialExtendRepository) ListImageTagsByTag(tagID uint, page, pageSize int) ([]model.ImageTag, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 50
	}
	var list []model.ImageTag
	var total int64
	q := r.db.Model(&model.ImageTag{}).Where("tag_id = ?", tagID)
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := q.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// ===== 搜索历史 =====

func (r *materialExtendRepository) CreateSearchHistory(h *model.SearchHistory) error {
	return r.db.Create(h).Error
}

func (r *materialExtendRepository) ListSearchHistory(userID uint, page, pageSize int) ([]model.SearchHistory, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	var list []model.SearchHistory
	var total int64
	q := r.db.Model(&model.SearchHistory{}).Where("user_id = ?", userID)
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := q.Order("created_at DESC, id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// ===== 相似结果 =====

func (r *materialExtendRepository) CreateSimilarResult(r2 *model.SimilarResult) error {
	return r.db.Create(r2).Error
}

func (r *materialExtendRepository) BatchCreateSimilarResults(list []model.SimilarResult) error {
	if len(list) == 0 {
		return nil
	}
	return r.db.CreateInBatches(list, 100).Error
}

func (r *materialExtendRepository) ListSimilarResults(sourceImageID uint, limit int) ([]model.SimilarResult, error) {
	if limit <= 0 {
		limit = 20
	}
	var list []model.SimilarResult
	if err := r.db.Where("source_image_id = ?", sourceImageID).
		Order("similarity DESC, id ASC").Limit(limit).Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

// ===== OCR 结果 =====

func (r *materialExtendRepository) CreateOCRResult(r2 *model.OCRResult) error {
	return r.db.Create(r2).Error
}

func (r *materialExtendRepository) FindOCRResultByImageID(imageID uint) (*model.OCRResult, error) {
	var r2 model.OCRResult
	if err := r.db.Where("image_id = ?", imageID).Order("id DESC").First(&r2).Error; err != nil {
		return nil, err
	}
	return &r2, nil
}

func (r *materialExtendRepository) ListOCRResults(page, pageSize int) ([]model.OCRResult, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	var list []model.OCRResult
	var total int64
	q := r.db.Model(&model.OCRResult{})
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := q.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *materialExtendRepository) UpdateOCRResultFields(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.OCRResult{}).Where("id = ?", id).Updates(fields).Error
}

// ===== 统计 =====

func (r *materialExtendRepository) StatTotalFiles() (int64, error) {
	var count int64
	err := r.db.Model(&model.File{}).Count(&count).Error
	return count, err
}

func (r *materialExtendRepository) StatTotalImages() (int64, error) {
	var count int64
	err := r.db.Model(&model.Image{}).Count(&count).Error
	return count, err
}

func (r *materialExtendRepository) StatTotalVideos() (int64, error) {
	var count int64
	err := r.db.Model(&model.Video{}).Count(&count).Error
	return count, err
}

func (r *materialExtendRepository) StatTotalCategories() (int64, error) {
	var count int64
	err := r.db.Model(&model.Category{}).Where("status = ?", 1).Count(&count).Error
	return count, err
}

func (r *materialExtendRepository) StatTotalTags() (int64, error) {
	var count int64
	err := r.db.Model(&model.Tag{}).Where("status = ?", 1).Count(&count).Error
	return count, err
}

func (r *materialExtendRepository) StatTotalSearches() (int64, error) {
	var count int64
	err := r.db.Model(&model.SearchHistory{}).Count(&count).Error
	return count, err
}

func (r *materialExtendRepository) StatTotalOCR() (int64, error) {
	var count int64
	err := r.db.Model(&model.OCRResult{}).Count(&count).Error
	return count, err
}

func (r *materialExtendRepository) StatStorageSize() (int64, error) {
	var sum int64
	err := r.db.Model(&model.File{}).Select("COALESCE(SUM(file_size),0)").Scan(&sum).Error
	return sum, err
}
