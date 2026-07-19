// Package service 同城车辆买卖业务逻辑层 - 车况检测
// 依据需求文档 1.10：4 维数据隔离（region_id + user_id）
// 依据 v3.2.1 架构方案：对标瓜子 254 项检测
package service

import (
	"errors"
	"fmt"
	"math/rand"
	"time"

	"wuchang-tongcheng/internal/modules/car/dto"
	"wuchang-tongcheng/internal/modules/car/model"
	"wuchang-tongcheng/internal/modules/car/repository"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

var (
	ErrInspectionNotFound      = errors.New("检测单不存在")
	ErrInspectionNoPermission  = errors.New("无权操作此检测单")
	ErrInspectionStatusInvalid = errors.New("检测单状态不允许此操作")
)

// InspectionService 车况检测业务接口
type InspectionService interface {
	// C 端
	Create(regionID uint, userID uint, req *dto.CreateInspectionRequest) (*dto.InspectionInfo, error)
	Update(id uint, operatorID uint, req *dto.UpdateInspectionRequest) error
	Delete(id uint, operatorID uint) error
	GetByID(id uint) (*dto.InspectionInfo, error)
	GetByCarID(carID uint) (*dto.InspectionInfo, error)
	List(regionID uint, req *dto.InspectionListRequest) (*utils.Pagination, []dto.InspectionInfo, error)
	ListByInspector(inspectorID uint, page, pageSize int) (*utils.Pagination, []dto.InspectionInfo, error)

	// M 端管理
	AdminList(req *dto.InspectionListRequest) (*utils.Pagination, []dto.InspectionInfo, error)
	AdminGetByID(id uint) (*dto.InspectionInfo, error)
	Review(id uint, req *dto.InspectionReviewRequest) error
	UpdateStatus(id uint, status int) error
}

type inspectionService struct {
	repo repository.InspectionRepository
}

// NewInspectionService 创建车况检测 service 实例
func NewInspectionService(repo repository.InspectionRepository) InspectionService {
	return &inspectionService{repo: repo}
}

// inspectionStatusText 检测状态文本
func inspectionStatusText(status int) string {
	switch status {
	case model.InspectionStatusPending:
		return "待检测"
	case model.InspectionStatusInProgress:
		return "检测中"
	case model.InspectionStatusCompleted:
		return "已完成"
	case model.InspectionStatusReviewed:
		return "已复核"
	case model.InspectionStatusCanceled:
		return "已取消"
	}
	return ""
}

// toInspectionInfo model -> dto
func toInspectionInfo(i *model.CarInspection) *dto.InspectionInfo {
	info := &dto.InspectionInfo{
		ID:               i.ID,
		InspectionNo:     i.InspectionNo,
		CarID:            i.CarID,
		ListingID:        i.ListingID,
		InspectorID:      i.InspectorID,
		InspectorName:    i.InspectorName,
		InspectorLevel:   i.InspectorLevel,
		InspectionType:   i.InspectionType,
		TotalItems:       i.TotalItems,
		PassedItems:      i.PassedItems,
		FailedItems:      i.FailedItems,
		WarningItems:     i.WarningItems,
		OverallScore:     i.OverallScore,
		ConditionLevel:   i.ConditionLevel,
		ExteriorScore:    i.ExteriorScore,
		InteriorScore:    i.InteriorScore,
		EngineScore:      i.EngineScore,
		ChassisScore:     i.ChassisScore,
		ElectronicsScore: i.ElectronicsScore,
		SafetyScore:      i.SafetyScore,
		HasAccident:      i.HasAccident,
		HasFlood:         i.HasFlood,
		HasFire:          i.HasFire,
		HasOverhaul:      i.HasOverhaul,
		ReportURL:        i.ReportURL,
		StartedAt:        i.StartedAt,
		CompletedAt:      i.CompletedAt,
		ReviewedBy:       i.ReviewedBy,
		ReviewedAt:       i.ReviewedAt,
		Status:           i.Status,
		StatusText:       inspectionStatusText(i.Status),
		Remark:           i.Remark,
		RegionID:         i.RegionID,
		CreatedAt:        i.CreatedAt,
		UpdatedAt:        i.UpdatedAt,
	}
	// JSONB 字段透传
	if i.Items != nil {
		info.Items = i.Items
	}
	if i.AccidentHistory != nil {
		info.AccidentHistory = i.AccidentHistory
	}
	if i.ReportImages != nil {
		info.ReportImages = i.ReportImages
	}
	return info
}

// genInspectionNo 生成检测单号：IN + yyyyMMddHHmmss + 6 位随机
func genInspectionNo() string {
	return fmt.Sprintf("IN%s%06d", time.Now().Format("20060102150405"), rand.Intn(1000000))
}

// ===== C 端 =====

// Create 创建检测单
func (s *inspectionService) Create(regionID uint, userID uint, req *dto.CreateInspectionRequest) (*dto.InspectionInfo, error) {
	i := &model.CarInspection{
		InspectionNo:     genInspectionNo(),
		CarID:            req.CarID,
		ListingID:        req.ListingID,
		InspectorID:      req.InspectorID,
		InspectorName:    req.InspectorName,
		InspectorLevel:   req.InspectorLevel,
		InspectionType:   req.InspectionType,
		TotalItems:       req.TotalItems,
		PassedItems:      req.PassedItems,
		FailedItems:      req.FailedItems,
		WarningItems:     req.WarningItems,
		OverallScore:     req.OverallScore,
		ConditionLevel:   req.ConditionLevel,
		ExteriorScore:    req.ExteriorScore,
		InteriorScore:    req.InteriorScore,
		EngineScore:      req.EngineScore,
		ChassisScore:     req.ChassisScore,
		ElectronicsScore: req.ElectronicsScore,
		SafetyScore:      req.SafetyScore,
		HasAccident:      req.HasAccident,
		HasFlood:         req.HasFlood,
		HasFire:          req.HasFire,
		HasOverhaul:      req.HasOverhaul,
		ReportURL:        req.ReportURL,
		Remark:           req.Remark,
		Status:           model.InspectionStatusPending,
	}
	i.RegionID = regionID

	// 默认值兜底
	if i.InspectorLevel == "" {
		i.InspectorLevel = model.InspectorLevelJunior
	}
	if i.InspectionType == "" {
		i.InspectionType = model.InspectionTypeStandard
	}
	if i.ConditionLevel == "" {
		i.ConditionLevel = model.InspectionConditionA
	}
	if i.TotalItems == 0 {
		i.TotalItems = 254
	}

	// JSONB 字段
	if req.Items != nil {
		if jb, err := model.FromJSON(req.Items); err == nil {
			i.Items = jb
		}
	}
	if req.AccidentHistory != nil {
		if jb, err := model.FromJSON(req.AccidentHistory); err == nil {
			i.AccidentHistory = jb
		}
	}
	if req.ReportImages != nil {
		if jb, err := model.FromJSON(req.ReportImages); err == nil {
			i.ReportImages = jb
		}
	}

	if err := s.repo.Create(i); err != nil {
		return nil, err
	}
	return toInspectionInfo(i), nil
}

// Update 更新检测单
func (s *inspectionService) Update(id uint, operatorID uint, req *dto.UpdateInspectionRequest) error {
	i, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrInspectionNotFound
		}
		return err
	}
	if i.InspectorID != operatorID && operatorID != 0 {
		// operatorID = 0 表示 M 端强制操作
		return ErrInspectionNoPermission
	}

	fields := map[string]interface{}{}
	if req.InspectorID != nil {
		fields["inspector_id"] = *req.InspectorID
	}
	if req.InspectorName != nil {
		fields["inspector_name"] = *req.InspectorName
	}
	if req.InspectorLevel != nil {
		fields["inspector_level"] = *req.InspectorLevel
	}
	if req.InspectionType != nil {
		fields["inspection_type"] = *req.InspectionType
	}
	if req.TotalItems != nil {
		fields["total_items"] = *req.TotalItems
	}
	if req.PassedItems != nil {
		fields["passed_items"] = *req.PassedItems
	}
	if req.FailedItems != nil {
		fields["failed_items"] = *req.FailedItems
	}
	if req.WarningItems != nil {
		fields["warning_items"] = *req.WarningItems
	}
	if req.OverallScore != nil {
		fields["overall_score"] = *req.OverallScore
	}
	if req.ConditionLevel != nil {
		fields["condition_level"] = *req.ConditionLevel
	}
	if req.ExteriorScore != nil {
		fields["exterior_score"] = *req.ExteriorScore
	}
	if req.InteriorScore != nil {
		fields["interior_score"] = *req.InteriorScore
	}
	if req.EngineScore != nil {
		fields["engine_score"] = *req.EngineScore
	}
	if req.ChassisScore != nil {
		fields["chassis_score"] = *req.ChassisScore
	}
	if req.ElectronicsScore != nil {
		fields["electronics_score"] = *req.ElectronicsScore
	}
	if req.SafetyScore != nil {
		fields["safety_score"] = *req.SafetyScore
	}
	if req.Items != nil {
		if jb, err := model.FromJSON(req.Items); err == nil {
			fields["items"] = jb
		}
	}
	if req.AccidentHistory != nil {
		if jb, err := model.FromJSON(req.AccidentHistory); err == nil {
			fields["accident_history"] = jb
		}
	}
	if req.HasAccident != nil {
		fields["has_accident"] = *req.HasAccident
	}
	if req.HasFlood != nil {
		fields["has_flood"] = *req.HasFlood
	}
	if req.HasFire != nil {
		fields["has_fire"] = *req.HasFire
	}
	if req.HasOverhaul != nil {
		fields["has_overhaul"] = *req.HasOverhaul
	}
	if req.ReportURL != nil {
		fields["report_url"] = *req.ReportURL
	}
	if req.ReportImages != nil {
		if jb, err := model.FromJSON(req.ReportImages); err == nil {
			fields["report_images"] = jb
		}
	}
	if req.Remark != nil {
		fields["remark"] = *req.Remark
	}

	// 状态变更
	if req.Status != nil {
		now := time.Now()
		switch *req.Status {
		case model.InspectionStatusInProgress:
			fields["status"] = model.InspectionStatusInProgress
			if i.StartedAt == nil {
				fields["started_at"] = &now
			}
		case model.InspectionStatusCompleted:
			fields["status"] = model.InspectionStatusCompleted
			if i.StartedAt == nil {
				fields["started_at"] = &now
			}
			fields["completed_at"] = &now
		case model.InspectionStatusCanceled:
			fields["status"] = model.InspectionStatusCanceled
		default:
			fields["status"] = *req.Status
		}
	}

	if len(fields) == 0 {
		return nil
	}
	return s.repo.Update(id, fields)
}

// Delete 删除检测单
func (s *inspectionService) Delete(id uint, operatorID uint) error {
	i, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrInspectionNotFound
		}
		return err
	}
	if i.InspectorID != operatorID && operatorID != 0 {
		return ErrInspectionNoPermission
	}
	return s.repo.Delete(id)
}

// GetByID 获取详情
func (s *inspectionService) GetByID(id uint) (*dto.InspectionInfo, error) {
	i, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInspectionNotFound
		}
		return nil, err
	}
	return toInspectionInfo(i), nil
}

// GetByCarID 按车源查询最新检测单
func (s *inspectionService) GetByCarID(carID uint) (*dto.InspectionInfo, error) {
	i, err := s.repo.FindByCarID(carID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInspectionNotFound
		}
		return nil, err
	}
	return toInspectionInfo(i), nil
}

// List C 端列表查询（地区隔离）
func (s *inspectionService) List(regionID uint, req *dto.InspectionListRequest) (*utils.Pagination, []dto.InspectionInfo, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	opts := repository.InspectionListOptions{
		CarID:           req.CarID,
		ListingID:       req.ListingID,
		InspectorID:     req.InspectorID,
		InspectionType:  req.InspectionType,
		ConditionLevel:  req.ConditionLevel,
		HasAccident:     req.HasAccident,
		Status:          req.Status,
		Keyword:         req.Keyword,
	}
	list, total, err := s.repo.List(regionID, pagination, opts)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	result := make([]dto.InspectionInfo, 0, len(list))
	for i := range list {
		result = append(result, *toInspectionInfo(&list[i]))
	}
	return pagination, result, nil
}

// ListByInspector 检测师自己的检测单
func (s *inspectionService) ListByInspector(inspectorID uint, page, pageSize int) (*utils.Pagination, []dto.InspectionInfo, error) {
	pagination := utils.NewPagination(page, pageSize)
	list, total, err := s.repo.ListByInspector(inspectorID, pagination)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	result := make([]dto.InspectionInfo, 0, len(list))
	for i := range list {
		result = append(result, *toInspectionInfo(&list[i]))
	}
	return pagination, result, nil
}

// ===== M 端管理 =====

func (s *inspectionService) AdminList(req *dto.InspectionListRequest) (*utils.Pagination, []dto.InspectionInfo, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	opts := repository.InspectionAdminListOptions{
		CarID:          req.CarID,
		InspectorID:    req.InspectorID,
		InspectionType: req.InspectionType,
		ConditionLevel: req.ConditionLevel,
		HasAccident:    req.HasAccident,
		Status:         req.Status,
		Keyword:        req.Keyword,
	}
	list, total, err := s.repo.AdminList(pagination, opts)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	result := make([]dto.InspectionInfo, 0, len(list))
	for i := range list {
		result = append(result, *toInspectionInfo(&list[i]))
	}
	return pagination, result, nil
}

func (s *inspectionService) AdminGetByID(id uint) (*dto.InspectionInfo, error) {
	i, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInspectionNotFound
		}
		return nil, err
	}
	return toInspectionInfo(i), nil
}

// Review M 端复核检测报告
func (s *inspectionService) Review(id uint, req *dto.InspectionReviewRequest) error {
	i, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrInspectionNotFound
		}
		return err
	}
	// 只有已完成的检测单可以复核
	if i.Status != model.InspectionStatusCompleted && i.Status != model.InspectionStatusReviewed {
		return ErrInspectionStatusInvalid
	}

	now := time.Now()
	fields := map[string]interface{}{
		"reviewed_by": req.ReviewedBy,
		"reviewed_at": &now,
		"status":      req.Status,
	}
	if req.Remark != "" {
		fields["remark"] = req.Remark
	}
	return s.repo.Update(id, fields)
}

// UpdateStatus M 端强制更新状态
func (s *inspectionService) UpdateStatus(id uint, status int) error {
	fields := map[string]interface{}{
		"status": status,
	}
	now := time.Now()
	switch status {
	case model.InspectionStatusInProgress:
		fields["started_at"] = &now
	case model.InspectionStatusCompleted:
		fields["completed_at"] = &now
	}
	return s.repo.Update(id, fields)
}
