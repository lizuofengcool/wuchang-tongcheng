// Package service 同城零工兼职业务逻辑层 - 技能标签
// 对标猪八戒威客：技能分类 + 认证 + 评分 + 关联岗位
// 注意：BaseModel 无 region_id，全局共享
package service

import (
	"errors"

	"wuchang-tongcheng/internal/modules/linggong/dto"
	"wuchang-tongcheng/internal/modules/linggong/model"
	"wuchang-tongcheng/internal/modules/linggong/repository"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

var (
	ErrSkillNotFound     = errors.New("技能不存在")
	ErrSkillDuplicate    = errors.New("技能名已存在")
	ErrSkillHasChildren  = errors.New("技能下存在子技能，不能删除")
)

// SkillService 技能标签业务接口
type SkillService interface {
	// C 端
	GetByID(id uint) (*dto.SkillInfo, error)
	List(req *dto.SkillListRequest) (*utils.Pagination, []dto.SkillInfo, error)
	ListByCategory(category string, page, pageSize int) (*utils.Pagination, []dto.SkillInfo, error)
	ListByParent(parentID uint, page, pageSize int) (*utils.Pagination, []dto.SkillInfo, error)
	ListHot(page, pageSize int) (*utils.Pagination, []dto.SkillInfo, error)

	// M 端管理
	Create(req *dto.CreateSkillRequest) (*dto.SkillInfo, error)
	Update(id uint, req *dto.UpdateSkillRequest) error
	Delete(id uint) error
	UpdateStatus(id uint, status int) error
}

type skillService struct {
	repo repository.SkillRepository
}

// NewSkillService 创建技能标签 service 实例
func NewSkillService(repo repository.SkillRepository) SkillService {
	return &skillService{repo: repo}
}

// skillCategoryText 技能类型文本
func skillCategoryText(c string) string {
	switch c {
	case model.SkillTypeTechnical:
		return "技术类"
	case model.SkillTypeService:
		return "服务类"
	case model.SkillTypeArt:
		return "艺术类"
	case model.SkillTypeLanguage:
		return "语言类"
	case model.SkillTypeSports:
		return "体育类"
	case model.SkillTypeDriver:
		return "驾驶类"
	case model.SkillTypeLabor:
		return "劳务类"
	case model.SkillTypeDesign:
		return "设计类"
	case model.SkillTypeMarketing:
		return "营销类"
	case model.SkillTypeWriting:
		return "文案类"
	case model.SkillTypeCatering:
		return "餐饮类"
	case model.SkillTypeOther:
		return "其他"
	}
	return ""
}

// skillStatusText 技能状态文本
func skillStatusText(s int) string {
	switch s {
	case 0:
		return "禁用"
	case 1:
		return "启用"
	}
	return ""
}

// toSkillInfo model -> dto
func toSkillInfo(s *model.LinggongSkill) *dto.SkillInfo {
	return &dto.SkillInfo{
		ID:            s.ID,
		Name:          s.Name,
		Code:          s.Code,
		Category:      s.Category,
		CategoryText:  skillCategoryText(s.Category),
		ParentID:      s.ParentID,
		Level:         s.Level,
		Icon:          s.Icon,
		Color:         s.Color,
		Description:   s.Description,
		WorkerCount:   s.WorkerCount,
		LinggongCount: s.LinggongCount,
		AvgSalary:     s.AvgSalary,
		HotScore:      s.HotScore,
		Status:        s.Status,
		StatusText:    skillStatusText(s.Status),
		Sort:          s.Sort,
		CreatedAt:     s.CreatedAt,
		UpdatedAt:     s.UpdatedAt,
	}
}

// ===== C 端 =====

// GetByID 获取技能详情
func (s *skillService) GetByID(id uint) (*dto.SkillInfo, error) {
	sk, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrSkillNotFound
		}
		return nil, err
	}
	return toSkillInfo(sk), nil
}

// List 技能列表
func (s *skillService) List(req *dto.SkillListRequest) (*utils.Pagination, []dto.SkillInfo, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	opts := repository.SkillListOptions{
		Category: req.Category,
		ParentID: req.ParentID,
		Level:    req.Level,
		Status:   req.Status,
		Keyword:  req.Keyword,
	}
	list, total, err := s.repo.List(pagination, opts)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	result := make([]dto.SkillInfo, 0, len(list))
	for i := range list {
		result = append(result, *toSkillInfo(&list[i]))
	}
	return pagination, result, nil
}

// ListByCategory 按分类查询技能
func (s *skillService) ListByCategory(category string, page, pageSize int) (*utils.Pagination, []dto.SkillInfo, error) {
	pagination := utils.NewPagination(page, pageSize)
	list, total, err := s.repo.ListByCategory(category, pagination)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	result := make([]dto.SkillInfo, 0, len(list))
	for i := range list {
		result = append(result, *toSkillInfo(&list[i]))
	}
	return pagination, result, nil
}

// ListByParent 按父技能查询子技能
func (s *skillService) ListByParent(parentID uint, page, pageSize int) (*utils.Pagination, []dto.SkillInfo, error) {
	pagination := utils.NewPagination(page, pageSize)
	list, total, err := s.repo.ListByParent(parentID, pagination)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	result := make([]dto.SkillInfo, 0, len(list))
	for i := range list {
		result = append(result, *toSkillInfo(&list[i]))
	}
	return pagination, result, nil
}

// ListHot 热门技能列表
func (s *skillService) ListHot(page, pageSize int) (*utils.Pagination, []dto.SkillInfo, error) {
	pagination := utils.NewPagination(page, pageSize)
	list, total, err := s.repo.ListHot(pagination)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	result := make([]dto.SkillInfo, 0, len(list))
	for i := range list {
		result = append(result, *toSkillInfo(&list[i]))
	}
	return pagination, result, nil
}

// ===== M 端管理 =====

// Create 创建技能
func (s *skillService) Create(req *dto.CreateSkillRequest) (*dto.SkillInfo, error) {
	// 名称唯一性校验
	if existing, err := s.repo.FindByName(req.Name); err == nil && existing.ID > 0 {
		return nil, ErrSkillDuplicate
	} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	sk := &model.LinggongSkill{
		Name:        req.Name,
		Code:        req.Code,
		Category:    req.Category,
		ParentID:    req.ParentID,
		Level:       req.Level,
		Icon:        req.Icon,
		Color:       req.Color,
		Description: req.Description,
		Sort:        req.Sort,
		Status:      1,
	}

	// 默认值兜底
	if sk.Category == "" {
		sk.Category = model.SkillTypeOther
	}
	if sk.Level == 0 {
		sk.Level = 1
	}

	if err := s.repo.Create(sk); err != nil {
		return nil, err
	}
	return toSkillInfo(sk), nil
}

// Update 更新技能
func (s *skillService) Update(id uint, req *dto.UpdateSkillRequest) error {
	if _, err := s.repo.FindByID(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrSkillNotFound
		}
		return err
	}

	fields := map[string]interface{}{}
	if req.Name != nil {
		fields["name"] = *req.Name
	}
	if req.Code != nil {
		fields["code"] = *req.Code
	}
	if req.Category != nil {
		fields["category"] = *req.Category
	}
	if req.ParentID != nil {
		fields["parent_id"] = *req.ParentID
	}
	if req.Level != nil {
		fields["level"] = *req.Level
	}
	if req.Icon != nil {
		fields["icon"] = *req.Icon
	}
	if req.Color != nil {
		fields["color"] = *req.Color
	}
	if req.Description != nil {
		fields["description"] = *req.Description
	}
	if req.HotScore != nil {
		fields["hot_score"] = *req.HotScore
	}
	if req.Status != nil {
		fields["status"] = *req.Status
	}
	if req.Sort != nil {
		fields["sort"] = *req.Sort
	}

	if len(fields) == 0 {
		return nil
	}
	return s.repo.Update(id, fields)
}

// Delete 删除技能
func (s *skillService) Delete(id uint) error {
	if _, err := s.repo.FindByID(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrSkillNotFound
		}
		return err
	}
	// 检查是否有子技能
	children, total, err := s.repo.ListByParent(id, utils.NewPagination(1, 1))
	if err != nil {
		return err
	}
	if total > 0 && len(children) > 0 {
		return ErrSkillHasChildren
	}
	return s.repo.Delete(id)
}

// UpdateStatus 启用/禁用技能
func (s *skillService) UpdateStatus(id uint, status int) error {
	if _, err := s.repo.FindByID(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrSkillNotFound
		}
		return err
	}
	return s.repo.Update(id, map[string]interface{}{"status": status})
}
