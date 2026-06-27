// Package repository 文件数据访问层
package repository

import (
	"wuchang-tongcheng/internal/modules/file/model"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

// FileRepository 文件仓储接口
type FileRepository interface {
	Create(record *model.FileUpload) error
	GetByID(id uint) (*model.FileUpload, error)
	List(pagination *utils.Pagination, fileType, keyword string) ([]model.FileUpload, int64, error)
	Delete(id uint) error
}

type fileRepository struct {
	db *gorm.DB
}

// NewFileRepository 创建文件仓储
func NewFileRepository(db *gorm.DB) FileRepository {
	return &fileRepository{db: db}
}

// Create 创建文件记录
func (r *fileRepository) Create(record *model.FileUpload) error {
	return r.db.Create(record).Error
}

// GetByID 根据ID获取文件
func (r *fileRepository) GetByID(id uint) (*model.FileUpload, error) {
	var record model.FileUpload
	if err := r.db.First(&record, id).Error; err != nil {
		return nil, err
	}
	return &record, nil
}

// List 文件分页列表（支持类型筛选、文件名关键词）
func (r *fileRepository) List(pagination *utils.Pagination, fileType, keyword string) ([]model.FileUpload, int64, error) {
	var list []model.FileUpload
	var total int64
	query := r.db.Model(&model.FileUpload{})
	if fileType != "" {
		query = query.Where("file_type = ?", fileType)
	}
	if keyword != "" {
		query = query.Where("file_name LIKE ?", "%"+keyword+"%")
	}
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(pagination)).Order("id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// Delete 删除文件记录
func (r *fileRepository) Delete(id uint) error {
	return r.db.Delete(&model.FileUpload{}, id).Error
}
