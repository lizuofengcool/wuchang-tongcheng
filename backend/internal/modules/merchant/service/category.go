// Package service 商户中台业务逻辑层 - 类目
// 树形 CRUD
package service

import (
	"errors"

	"wuchang-tongcheng/internal/modules/merchant/dto"
	"wuchang-tongcheng/internal/modules/merchant/model"
	"wuchang-tongcheng/internal/modules/merchant/repository"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

var (
	ErrCategoryNotFound     = errors.New("类目不存在")
	ErrCategoryHasChildren  = errors.New("该类目下有子类目，无法删除")
	ErrCategoryStatusInvalid = errors.New("类目状态无效")
)

// CategoryService 类目业务接口
type CategoryService interface {
	Create(req *dto.CreateCategoryRequest) (*dto.CategoryInfo, error)
	Update(id uint, req *dto.UpdateCategoryRequest) error
	Delete(id uint) error
	GetByID(id uint) (*dto.CategoryInfo, error)
	List(req *dto.CategoryListRequest) (*utils.Pagination, []dto.CategoryInfo, error)
	Tree() ([]dto.CategoryInfo, error)
	UpdateStatus(id uint, status int) error
}

type categoryService struct {
	repo repository.CategoryRepository
}

// NewCategoryService 创建类目 service 实例
func NewCategoryService(repo repository.CategoryRepository) CategoryService {
	return &categoryService{repo: repo}
}

// categoryStatusText 类目状态文本
func categoryStatusText(status int) string {
	switch status {
	case model.CategoryStatusDisabled:
		return "禁用"
	case model.CategoryStatusEnabled:
		return "启用"
	}
	return ""
}

// Create 创建类目
func (s *categoryService) Create(req *dto.CreateCategoryRequest) (*dto.CategoryInfo, error) {
	status := req.Status
	if status == 0 {
		status = model.CategoryStatusEnabled
	}
	c := &model.Category{
		ParentID: req.ParentID,
		Name:     req.Name,
		Icon:     req.Icon,
		Sort:     req.Sort,
		Status:   status,
	}
	if err := s.repo.Create(c); err != nil {
		return nil, err
	}
	return s.toInfo(c), nil
}

// Update 更新类目
func (s *categoryService) Update(id uint, req *dto.UpdateCategoryRequest) error {
	_, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrCategoryNotFound
		}
		return err
	}
	fields := make(map[string]interface{})
	if req.ParentID != nil {
		fields["parent_id"] = *req.ParentID
	}
	if req.Name != nil {
		fields["name"] = *req.Name
	}
	if req.Icon != nil {
		fields["icon"] = *req.Icon
	}
	if req.Sort != nil {
		fields["sort"] = *req.Sort
	}
	if req.Status != nil {
		fields["status"] = *req.Status
	}
	if len(fields) == 0 {
		return nil
	}
	return s.repo.UpdateFields(id, fields)
}

// Delete 删除类目
func (s *categoryService) Delete(id uint) error {
	_, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrCategoryNotFound
		}
		return err
	}
	// 检查是否有子类目
	children, err := s.repo.FindByParent(id)
	if err != nil {
		return err
	}
	if len(children) > 0 {
		return ErrCategoryHasChildren
	}
	return s.repo.Delete(id)
}

// GetByID 类目详情
func (s *categoryService) GetByID(id uint) (*dto.CategoryInfo, error) {
	c, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCategoryNotFound
		}
		return nil, err
	}
	return s.toInfo(c), nil
}

// List 类目列表
func (s *categoryService) List(req *dto.CategoryListRequest) (*utils.Pagination, []dto.CategoryInfo, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	opts := repository.CategoryListOptions{
		ParentID: req.ParentID,
		Status:   req.Status,
		Keyword:  req.Keyword,
	}
	list, total, err := s.repo.List(pagination, opts)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	infos := make([]dto.CategoryInfo, 0, len(list))
	for i := range list {
		infos = append(infos, *s.toInfo(&list[i]))
	}
	return pagination, infos, nil
}

// Tree 类目树
func (s *categoryService) Tree() ([]dto.CategoryInfo, error) {
	all, err := s.repo.FindAll()
	if err != nil {
		return nil, err
	}
	infos := make([]dto.CategoryInfo, 0, len(all))
	for i := range all {
		infos = append(infos, *s.toInfo(&all[i]))
	}
	// 按 parent_id 分组
	childrenMap := make(map[uint][]dto.CategoryInfo)
	for i := range infos {
		pid := infos[i].ParentID
		childrenMap[pid] = append(childrenMap[pid], infos[i])
	}
	// 递归填充 children
	var fill func(node *dto.CategoryInfo)
	fill = func(node *dto.CategoryInfo) {
		if subs, ok := childrenMap[node.ID]; ok {
			for j := range subs {
				fill(&subs[j])
			}
			node.Children = subs
		}
	}
	roots := childrenMap[0]
	for i := range roots {
		fill(&roots[i])
	}
	return roots, nil
}

// UpdateStatus 更新状态
func (s *categoryService) UpdateStatus(id uint, status int) error {
	if status != model.CategoryStatusDisabled && status != model.CategoryStatusEnabled {
		return ErrCategoryStatusInvalid
	}
	_, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrCategoryNotFound
		}
		return err
	}
	return s.repo.UpdateFields(id, map[string]interface{}{"status": status})
}

// toInfo 模型转 DTO
func (s *categoryService) toInfo(m *model.Category) *dto.CategoryInfo {
	return &dto.CategoryInfo{
		ID:          m.ID,
		ParentID:    m.ParentID,
		Name:        m.Name,
		Icon:        m.Icon,
		Sort:        m.Sort,
		Status:      m.Status,
		StatusText:  categoryStatusText(m.Status),
		CreatedAt:   m.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:   m.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}
