// Package repository 素材存储中台精简版数据访问层
package repository

import (
	"wuchang-tongcheng/internal/modules/material/model"

	"gorm.io/gorm"
)

// MaterialRepository 素材中台仓储接口
type MaterialRepository interface {
	// 文件
	CreateFile(f *model.File) error
	FindFileByID(fileID string) (*model.File, error)
	FindFileByHash(hash string) (*model.File, error)
	ListFiles(userID uint, fileType string, category string, page, pageSize int) ([]model.File, int64, error)
	UpdateFileFields(id uint, fields map[string]interface{}) error
	DeleteFile(id uint) error

	// 图片
	CreateImage(img *model.Image) error
	FindImageByFileID(fileID string) (*model.Image, error)
	UpdateImageFields(id uint, fields map[string]interface{}) error

	// 视频
	CreateVideo(v *model.Video) error
	FindVideoByFileID(fileID string) (*model.Video, error)
	UpdateVideoFields(id uint, fields map[string]interface{}) error

	// 图片特征
	CreateImageFeature(f *model.ImageFeature) error
	FindImageFeatureByImageID(imageID uint) (*model.ImageFeature, error)
	FindSimilarByPHash(phash string, limit int) ([]model.ImageFeature, error)
}

type materialRepository struct {
	db *gorm.DB
}

// NewMaterialRepository 创建仓储实例
func NewMaterialRepository(db *gorm.DB) MaterialRepository {
	return &materialRepository{db: db}
}

// ===== 文件 =====

func (r *materialRepository) CreateFile(f *model.File) error {
	return r.db.Create(f).Error
}

func (r *materialRepository) FindFileByID(fileID string) (*model.File, error) {
	var f model.File
	if err := r.db.Where("file_id = ?", fileID).First(&f).Error; err != nil {
		return nil, err
	}
	return &f, nil
}

func (r *materialRepository) FindFileByHash(hash string) (*model.File, error) {
	var f model.File
	if err := r.db.Where("file_hash = ?", hash).First(&f).Error; err != nil {
		return nil, err
	}
	return &f, nil
}

func (r *materialRepository) ListFiles(userID uint, fileType string, category string, page, pageSize int) ([]model.File, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	var list []model.File
	var total int64
	q := r.db.Model(&model.File{})
	if userID > 0 {
		q = q.Where("user_id = ?", userID)
	}
	if fileType != "" {
		q = q.Where("file_type = ?", fileType)
	}
	if category != "" {
		q = q.Where("category = ?", category)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := q.Order("created_at DESC, id DESC").
		Offset((page - 1) * pageSize).Limit(pageSize).
		Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *materialRepository) UpdateFileFields(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.File{}).Where("id = ?", id).Updates(fields).Error
}

func (r *materialRepository) DeleteFile(id uint) error {
	return r.db.Delete(&model.File{}, id).Error
}

// ===== 图片 =====

func (r *materialRepository) CreateImage(img *model.Image) error {
	return r.db.Create(img).Error
}

func (r *materialRepository) FindImageByFileID(fileID string) (*model.Image, error) {
	var img model.Image
	if err := r.db.Where("file_id = ?", fileID).First(&img).Error; err != nil {
		return nil, err
	}
	return &img, nil
}

func (r *materialRepository) UpdateImageFields(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.Image{}).Where("id = ?", id).Updates(fields).Error
}

// ===== 视频 =====

func (r *materialRepository) CreateVideo(v *model.Video) error {
	return r.db.Create(v).Error
}

func (r *materialRepository) FindVideoByFileID(fileID string) (*model.Video, error) {
	var v model.Video
	if err := r.db.Where("file_id = ?", fileID).First(&v).Error; err != nil {
		return nil, err
	}
	return &v, nil
}

func (r *materialRepository) UpdateVideoFields(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.Video{}).Where("id = ?", id).Updates(fields).Error
}

// ===== 图片特征 =====

func (r *materialRepository) CreateImageFeature(f *model.ImageFeature) error {
	return r.db.Create(f).Error
}

func (r *materialRepository) FindImageFeatureByImageID(imageID uint) (*model.ImageFeature, error) {
	var f model.ImageFeature
	if err := r.db.Where("image_id = ?", imageID).First(&f).Error; err != nil {
		return nil, err
	}
	return &f, nil
}

// FindSimilarByPHash 基于 pHash 查找相似图片（精简版：前 16 字符相同视为相似）
// 实际可使用汉明距离算法
func (r *materialRepository) FindSimilarByPHash(phash string, limit int) ([]model.ImageFeature, error) {
	if limit <= 0 {
		limit = 10
	}
	var list []model.ImageFeature
	// 精简版：使用 LIKE 匹配 pHash 前缀
	prefix := phash
	if len(prefix) > 16 {
		prefix = prefix[:16]
	}
	if err := r.db.Where("phash LIKE ?", prefix+"%").
		Order("phash ASC").
		Limit(limit).
		Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}
