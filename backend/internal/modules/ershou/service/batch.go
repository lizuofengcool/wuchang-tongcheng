// Package service 批量操作业务逻辑层
// 依据 v3.2.1 架构方案：M 端批量审核/状态变更/删除/导出
package service

import (
	"wuchang-tongcheng/internal/modules/ershou/dto"
	"wuchang-tongcheng/internal/modules/ershou/model"
	"wuchang-tongcheng/internal/modules/ershou/repository"

	"gorm.io/gorm"
)

// BatchService 批量操作业务接口
type BatchService interface {
	// Audit 批量审核（M端）
	Audit(adminID uint, req *dto.BatchAuditRequest) (*dto.BatchResultResponse, error)
	// UpdateStatus 批量状态变更（M端）
	UpdateStatus(adminID uint, req *dto.BatchStatusUpdateRequest) (*dto.BatchResultResponse, error)
	// Delete 批量删除（M端软删除）
	Delete(adminID uint, req *dto.BatchDeleteRequest) (*dto.BatchResultResponse, error)
	// Export 导出 Excel/CSV（返回行数据，由 Handler 转为文件流）
	Export(req *dto.ExportRequest) ([]map[string]interface{}, error)
}

type batchService struct {
	db  *gorm.DB
	repo repository.ErshouRepository
}

// NewBatchService 创建批量操作 service 实例
func NewBatchService(db *gorm.DB, repo repository.ErshouRepository) BatchService {
	return &batchService{db: db, repo: repo}
}

// Audit 批量审核
func (s *batchService) Audit(adminID uint, req *dto.BatchAuditRequest) (*dto.BatchResultResponse, error) {
	result := &dto.BatchResultResponse{
		Total:     len(req.IDs),
		FailedIDs: []uint{},
	}
	for _, id := range req.IDs {
		fields := map[string]interface{}{
			"audit_status": req.AuditStatus,
		}
		if req.AuditReason != "" {
			fields["audit_reason"] = req.AuditReason
		}
		// 审核通过则同步发布状态
		if req.AuditStatus == model.AuditApproved {
			fields["status"] = model.StatusPublished
		}
		if err := s.repo.UpdateFields(id, fields); err != nil {
			result.Failed++
			result.FailedIDs = append(result.FailedIDs, id)
			continue
		}
		result.Success++
	}
	return result, nil
}

// UpdateStatus 批量状态变更
func (s *batchService) UpdateStatus(adminID uint, req *dto.BatchStatusUpdateRequest) (*dto.BatchResultResponse, error) {
	result := &dto.BatchResultResponse{
		Total:     len(req.IDs),
		FailedIDs: []uint{},
	}
	for _, id := range req.IDs {
		if err := s.repo.UpdateFields(id, map[string]interface{}{
			"status": req.Status,
		}); err != nil {
			result.Failed++
			result.FailedIDs = append(result.FailedIDs, id)
			continue
		}
		result.Success++
	}
	return result, nil
}

// Delete 批量删除（软删除）
func (s *batchService) Delete(adminID uint, req *dto.BatchDeleteRequest) (*dto.BatchResultResponse, error) {
	result := &dto.BatchResultResponse{
		Total:     len(req.IDs),
		FailedIDs: []uint{},
	}
	for _, id := range req.IDs {
		if err := s.repo.Delete(id); err != nil {
			result.Failed++
			result.FailedIDs = append(result.FailedIDs, id)
			continue
		}
		result.Success++
	}
	return result, nil
}

// Export 导出数据（返回行数据）
func (s *batchService) Export(req *dto.ExportRequest) ([]map[string]interface{}, error) {
	query := s.db.Model(&model.Ershou{})
	if req.Status != nil {
		query = query.Where("status = ?", *req.Status)
	}
	if req.CategoryID > 0 {
		query = query.Where("category_id = ?", req.CategoryID)
	}
	if req.UserID > 0 {
		query = query.Where("user_id = ?", req.UserID)
	}
	if req.Keyword != "" {
		query = query.Where("title ILIKE ?", "%"+req.Keyword+"%")
	}
	var list []model.Ershou
	if err := query.Order("created_at DESC, id DESC").Limit(10000).Find(&list).Error; err != nil {
		return nil, err
	}
	rows := make([]map[string]interface{}, 0, len(list))
	for _, e := range list {
		row := map[string]interface{}{
			"id":            e.ID,
			"title":         e.Title,
			"price":         e.Price,
			"condition":     e.Condition,
			"brand":         e.Brand,
			"user_id":       e.UserID,
			"user_name":     e.UserName,
			"user_phone":    e.UserPhone,
			"category_id":   e.CategoryID,
			"category_name": e.CategoryName,
			"status":        e.Status,
			"audit_status":  e.AuditStatus,
			"view_count":    e.ViewCount,
			"fav_count":     e.FavCount,
			"message_count": e.MessageCount,
			"region_id":     e.RegionID,
			"created_at":    e.CreatedAt,
		}
		rows = append(rows, row)
	}
	return rows, nil
}
