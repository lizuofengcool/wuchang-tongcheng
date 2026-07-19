// Package service 标签/品牌/型号/分类属性业务逻辑层
// 依据 v3.2.1 架构方案：对标转转
package service

import (
	"errors"

	"wuchang-tongcheng/internal/modules/ershou/dto"
	"wuchang-tongcheng/internal/modules/ershou/model"
	"wuchang-tongcheng/internal/modules/ershou/repository"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

var (
	ErrTagNotFound          = errors.New("标签不存在")
	ErrTagNameDuplicate     = errors.New("标签名称已存在")
	ErrBrandNotFound        = errors.New("品牌不存在")
	ErrBrandNameDuplicate   = errors.New("品牌名称已存在")
	ErrModelNotFound        = errors.New("型号不存在")
	ErrCategoryAttrNotFound = errors.New("分类属性不存在")
)

// ===== TagService =====

// TagService 标签业务接口
type TagService interface {
	Create(creatorID uint, req *dto.TagCreateRequest) (*dto.TagResponse, error)
	GetByID(id uint) (*dto.TagResponse, error)
	Update(id uint, req *dto.TagCreateRequest) (*dto.TagResponse, error)
	Delete(id uint) error
	List(tagType string, status *int, isHot *bool, page, pageSize int) (*utils.Pagination, []dto.TagResponse, error)
	ListHot(limit int) ([]dto.TagResponse, error)
}

type tagService struct {
	repo repository.TagRepository
}

// NewTagService 创建标签 service 实例
func NewTagService(repo repository.TagRepository) TagService {
	return &tagService{repo: repo}
}

func (s *tagService) Create(creatorID uint, req *dto.TagCreateRequest) (*dto.TagResponse, error) {
	tag := &model.ErshouTag{
		Name:       req.Name,
		Type:       req.Type,
		Color:      req.Color,
		Icon:       req.Icon,
		Background: req.Background,
		Status:     req.Status,
		Sort:       req.Sort,
		IsHot:      req.IsHot,
		CreatorID:  creatorID,
	}
	if tag.Type == "" {
		tag.Type = model.TagTypeCustom
	}
	if tag.Color == "" {
		tag.Color = "#409EFF"
	}
	if tag.Status == 0 {
		tag.Status = 1
	}
	if err := s.repo.Create(tag); err != nil {
		return nil, err
	}
	return s.toTagResponse(tag), nil
}

func (s *tagService) GetByID(id uint) (*dto.TagResponse, error) {
	t, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTagNotFound
		}
		return nil, err
	}
	return s.toTagResponse(t), nil
}

func (s *tagService) Update(id uint, req *dto.TagCreateRequest) (*dto.TagResponse, error) {
	if _, err := s.repo.FindByID(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTagNotFound
		}
		return nil, err
	}
	fields := map[string]interface{}{
		"name":       req.Name,
		"type":       req.Type,
		"color":      req.Color,
		"icon":       req.Icon,
		"background": req.Background,
		"status":     req.Status,
		"sort":       req.Sort,
		"is_hot":     req.IsHot,
	}
	if err := s.repo.Update(id, fields); err != nil {
		return nil, err
	}
	updated, _ := s.repo.FindByID(id)
	return s.toTagResponse(updated), nil
}

func (s *tagService) Delete(id uint) error {
	if _, err := s.repo.FindByID(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrTagNotFound
		}
		return err
	}
	return s.repo.Delete(id)
}

func (s *tagService) List(tagType string, status *int, isHot *bool, page, pageSize int) (*utils.Pagination, []dto.TagResponse, error) {
	pagination := utils.NewPagination(page, pageSize)
	list, total, err := s.repo.List(repository.TagListQuery{
		Type:   tagType,
		Status: status,
		IsHot:  isHot,
	}, pagination)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	result := make([]dto.TagResponse, 0, len(list))
	for i := range list {
		result = append(result, *s.toTagResponse(&list[i]))
	}
	return pagination, result, nil
}

func (s *tagService) ListHot(limit int) ([]dto.TagResponse, error) {
	list, err := s.repo.ListHot(limit)
	if err != nil {
		return nil, err
	}
	result := make([]dto.TagResponse, 0, len(list))
	for i := range list {
		result = append(result, *s.toTagResponse(&list[i]))
	}
	return result, nil
}

func (s *tagService) toTagResponse(t *model.ErshouTag) *dto.TagResponse {
	return &dto.TagResponse{
		ID:         t.ID,
		Name:       t.Name,
		Type:       t.Type,
		Color:      t.Color,
		Icon:       t.Icon,
		Background: t.Background,
		Status:     t.Status,
		Sort:       t.Sort,
		UseCount:   t.UseCount,
		IsHot:      t.IsHot,
		CreatorID:  t.CreatorID,
		CreatedAt:  t.CreatedAt,
	}
}

// ===== BrandService =====

// BrandService 品牌业务接口
type BrandService interface {
	Create(operatorID uint, req *dto.BrandCreateRequest) (*dto.BrandResponse, error)
	GetByID(id uint) (*dto.BrandResponse, error)
	Update(id uint, req *dto.BrandCreateRequest) (*dto.BrandResponse, error)
	Delete(id uint) error
	List(keyword string, status *int, page, pageSize int) (*utils.Pagination, []dto.BrandResponse, error)
}

type brandService struct {
	repo repository.BrandRepository
}

// NewBrandService 创建品牌 service 实例
func NewBrandService(repo repository.BrandRepository) BrandService {
	return &brandService{repo: repo}
}

func (s *brandService) Create(operatorID uint, req *dto.BrandCreateRequest) (*dto.BrandResponse, error) {
	// 同名品牌查重
	if existing, _ := s.repo.FindByName(req.Name); existing != nil {
		return nil, ErrBrandNameDuplicate
	}
	brand := &model.ErshouBrand{
		Name:             req.Name,
		Logo:             req.Logo,
		EnglishName:      req.EnglishName,
		Description:      req.Description,
		Country:          req.Country,
		OfficialVerified: req.OfficialVerified,
		OfficialURL:      req.OfficialURL,
		Status:           req.Status,
		Sort:             req.Sort,
	}
	if brand.Status == 0 {
		brand.Status = 1
	}
	if len(req.CategoryIDs) > 0 {
		if jb, err := model.FromJSON(req.CategoryIDs); err == nil {
			brand.CategoryIDs = jb
		}
	}
	if err := s.repo.Create(brand); err != nil {
		return nil, err
	}
	return s.toBrandResponse(brand), nil
}

func (s *brandService) GetByID(id uint) (*dto.BrandResponse, error) {
	b, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrBrandNotFound
		}
		return nil, err
	}
	return s.toBrandResponse(b), nil
}

func (s *brandService) Update(id uint, req *dto.BrandCreateRequest) (*dto.BrandResponse, error) {
	if _, err := s.repo.FindByID(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrBrandNotFound
		}
		return nil, err
	}
	fields := map[string]interface{}{
		"name":              req.Name,
		"logo":              req.Logo,
		"english_name":      req.EnglishName,
		"description":       req.Description,
		"country":           req.Country,
		"official_verified": req.OfficialVerified,
		"official_url":      req.OfficialURL,
		"status":            req.Status,
		"sort":              req.Sort,
	}
	if req.CategoryIDs != nil {
		if jb, err := model.FromJSON(req.CategoryIDs); err == nil {
			fields["category_ids"] = jb
		}
	}
	if err := s.repo.Update(id, fields); err != nil {
		return nil, err
	}
	updated, _ := s.repo.FindByID(id)
	return s.toBrandResponse(updated), nil
}

func (s *brandService) Delete(id uint) error {
	if _, err := s.repo.FindByID(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrBrandNotFound
		}
		return err
	}
	return s.repo.Delete(id)
}

func (s *brandService) List(keyword string, status *int, page, pageSize int) (*utils.Pagination, []dto.BrandResponse, error) {
	pagination := utils.NewPagination(page, pageSize)
	list, total, err := s.repo.List(repository.BrandListQuery{
		Keyword: keyword,
		Status:  status,
	}, pagination)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	result := make([]dto.BrandResponse, 0, len(list))
	for i := range list {
		result = append(result, *s.toBrandResponse(&list[i]))
	}
	return pagination, result, nil
}

func (s *brandService) toBrandResponse(b *model.ErshouBrand) *dto.BrandResponse {
	resp := &dto.BrandResponse{
		ID:                b.ID,
		Name:              b.Name,
		Logo:              b.Logo,
		EnglishName:       b.EnglishName,
		Description:       b.Description,
		Country:           b.Country,
		OfficialVerified:  b.OfficialVerified,
		OfficialURL:       b.OfficialURL,
		CategoryIDs:       []uint{},
		Status:            b.Status,
		Sort:              b.Sort,
		UseCount:          b.UseCount,
		CreatedAt:         b.CreatedAt,
	}
	if b.CategoryIDs != nil {
		var ids []uint
		_ = b.CategoryIDs.Parse(&ids)
		if ids != nil {
			resp.CategoryIDs = ids
		}
	}
	return resp
}

// ===== ModelService =====

// ModelService 型号业务接口
type ModelService interface {
	Create(brandID, operatorID uint, req *dto.ModelCreateRequest) (*dto.ModelResponse, error)
	GetByID(id uint) (*dto.ModelResponse, error)
	Update(id uint, req *dto.ModelCreateRequest) (*dto.ModelResponse, error)
	Delete(id uint) error
	ListByBrandID(brandID uint, page, pageSize int) (*utils.Pagination, []dto.ModelResponse, error)
	List(brandID uint, keyword string, status *int, page, pageSize int) (*utils.Pagination, []dto.ModelResponse, error)
}

type modelService struct {
	repo repository.ModelRepository
}

// NewModelService 创建型号 service 实例
func NewModelService(repo repository.ModelRepository) ModelService {
	return &modelService{repo: repo}
}

func (s *modelService) Create(brandID, operatorID uint, req *dto.ModelCreateRequest) (*dto.ModelResponse, error) {
	m := &model.ErshouModel{
		BrandID:        brandID,
		Name:           req.Name,
		FullName:       req.FullName,
		Image:          req.Image,
		ReleaseDate:    req.ReleaseDate,
		Status:         req.Status,
		Sort:           req.Sort,
		ReferencePrice: req.ReferencePrice,
	}
	if m.Status == 0 {
		m.Status = 1
	}
	if req.Specifications != nil {
		if jb, err := model.FromJSON(req.Specifications); err == nil {
			m.Specifications = jb
		}
	}
	if err := s.repo.Create(m); err != nil {
		return nil, err
	}
	return s.toModelResponse(m), nil
}

func (s *modelService) GetByID(id uint) (*dto.ModelResponse, error) {
	m, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrModelNotFound
		}
		return nil, err
	}
	return s.toModelResponse(m), nil
}

func (s *modelService) Update(id uint, req *dto.ModelCreateRequest) (*dto.ModelResponse, error) {
	if _, err := s.repo.FindByID(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrModelNotFound
		}
		return nil, err
	}
	fields := map[string]interface{}{
		"name":            req.Name,
		"full_name":       req.FullName,
		"image":           req.Image,
		"release_date":    req.ReleaseDate,
		"status":          req.Status,
		"sort":            req.Sort,
		"reference_price": req.ReferencePrice,
	}
	if req.Specifications != nil {
		if jb, err := model.FromJSON(req.Specifications); err == nil {
			fields["specifications"] = jb
		}
	}
	if err := s.repo.Update(id, fields); err != nil {
		return nil, err
	}
	updated, _ := s.repo.FindByID(id)
	return s.toModelResponse(updated), nil
}

func (s *modelService) Delete(id uint) error {
	if _, err := s.repo.FindByID(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrModelNotFound
		}
		return err
	}
	return s.repo.Delete(id)
}

func (s *modelService) ListByBrandID(brandID uint, page, pageSize int) (*utils.Pagination, []dto.ModelResponse, error) {
	pagination := utils.NewPagination(page, pageSize)
	list, total, err := s.repo.ListByBrandID(brandID, pagination)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	result := make([]dto.ModelResponse, 0, len(list))
	for i := range list {
		result = append(result, *s.toModelResponse(&list[i]))
	}
	return pagination, result, nil
}

func (s *modelService) List(brandID uint, keyword string, status *int, page, pageSize int) (*utils.Pagination, []dto.ModelResponse, error) {
	pagination := utils.NewPagination(page, pageSize)
	list, total, err := s.repo.List(repository.ModelListQuery{
		BrandID: brandID,
		Keyword: keyword,
		Status:  status,
	}, pagination)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	result := make([]dto.ModelResponse, 0, len(list))
	for i := range list {
		result = append(result, *s.toModelResponse(&list[i]))
	}
	return pagination, result, nil
}

func (s *modelService) toModelResponse(m *model.ErshouModel) *dto.ModelResponse {
	resp := &dto.ModelResponse{
		ID:              m.ID,
		BrandID:         m.BrandID,
		Name:            m.Name,
		FullName:        m.FullName,
		Image:           m.Image,
		ReleaseDate:     m.ReleaseDate,
		Status:          m.Status,
		Sort:            m.Sort,
		UseCount:        m.UseCount,
		ReferencePrice:  m.ReferencePrice,
		CreatedAt:       m.CreatedAt,
		Specifications:  map[string]interface{}{},
	}
	if m.Specifications != nil {
		var sp map[string]interface{}
		_ = m.Specifications.Parse(&sp)
		if sp != nil {
			resp.Specifications = sp
		}
	}
	return resp
}

// ===== CategoryAttrService =====

// CategoryAttrService 分类属性业务接口
type CategoryAttrService interface {
	Create(operatorID uint, req *dto.CategoryAttrCreateRequest) (*dto.CategoryAttrResponse, error)
	GetByID(id uint) (*dto.CategoryAttrResponse, error)
	Update(id uint, req *dto.CategoryAttrCreateRequest) (*dto.CategoryAttrResponse, error)
	Delete(id uint) error
	ListByCategoryID(categoryID uint) ([]dto.CategoryAttrResponse, error)
	List(categoryID uint, status *int, page, pageSize int) (*utils.Pagination, []dto.CategoryAttrResponse, error)
}

type categoryAttrService struct {
	repo repository.CategoryAttrRepository
}

// NewCategoryAttrService 创建分类属性 service 实例
func NewCategoryAttrService(repo repository.CategoryAttrRepository) CategoryAttrService {
	return &categoryAttrService{repo: repo}
}

func (s *categoryAttrService) Create(operatorID uint, req *dto.CategoryAttrCreateRequest) (*dto.CategoryAttrResponse, error) {
	a := &model.ErshouCategoryAttr{
		CategoryID:   req.CategoryID,
		AttrName:     req.AttrName,
		AttrKey:      req.AttrKey,
		AttrType:     req.AttrType,
		Unit:         req.Unit,
		IsRequired:   req.IsRequired,
		IsFilterable: req.IsFilterable,
		IsSearchable: req.IsSearchable,
		DefaultValue: req.DefaultValue,
		Placeholder:  req.Placeholder,
		Description:  req.Description,
		Status:       req.Status,
		Sort:         req.Sort,
	}
	if a.AttrType == "" {
		a.AttrType = model.CategoryAttrTypeString
	}
	if a.Status == 0 {
		a.Status = 1
	}
	if len(req.Options) > 0 {
		if jb, err := model.FromJSON(req.Options); err == nil {
			a.Options = jb
		}
	}
	if err := s.repo.Create(a); err != nil {
		return nil, err
	}
	return s.toCategoryAttrResponse(a), nil
}

func (s *categoryAttrService) GetByID(id uint) (*dto.CategoryAttrResponse, error) {
	a, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCategoryAttrNotFound
		}
		return nil, err
	}
	return s.toCategoryAttrResponse(a), nil
}

func (s *categoryAttrService) Update(id uint, req *dto.CategoryAttrCreateRequest) (*dto.CategoryAttrResponse, error) {
	if _, err := s.repo.FindByID(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCategoryAttrNotFound
		}
		return nil, err
	}
	fields := map[string]interface{}{
		"category_id":   req.CategoryID,
		"attr_name":     req.AttrName,
		"attr_key":      req.AttrKey,
		"attr_type":     req.AttrType,
		"unit":          req.Unit,
		"is_required":   req.IsRequired,
		"is_filterable": req.IsFilterable,
		"is_searchable": req.IsSearchable,
		"default_value": req.DefaultValue,
		"placeholder":   req.Placeholder,
		"description":   req.Description,
		"status":        req.Status,
		"sort":          req.Sort,
	}
	if req.Options != nil {
		if jb, err := model.FromJSON(req.Options); err == nil {
			fields["options"] = jb
		}
	}
	if err := s.repo.Update(id, fields); err != nil {
		return nil, err
	}
	updated, _ := s.repo.FindByID(id)
	return s.toCategoryAttrResponse(updated), nil
}

func (s *categoryAttrService) Delete(id uint) error {
	if _, err := s.repo.FindByID(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrCategoryAttrNotFound
		}
		return err
	}
	return s.repo.Delete(id)
}

func (s *categoryAttrService) ListByCategoryID(categoryID uint) ([]dto.CategoryAttrResponse, error) {
	list, err := s.repo.ListByCategoryID(categoryID)
	if err != nil {
		return nil, err
	}
	result := make([]dto.CategoryAttrResponse, 0, len(list))
	for i := range list {
		result = append(result, *s.toCategoryAttrResponse(&list[i]))
	}
	return result, nil
}

func (s *categoryAttrService) List(categoryID uint, status *int, page, pageSize int) (*utils.Pagination, []dto.CategoryAttrResponse, error) {
	pagination := utils.NewPagination(page, pageSize)
	list, total, err := s.repo.List(repository.CategoryAttrListQuery{
		CategoryID: categoryID,
		Status:     status,
	}, pagination)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	result := make([]dto.CategoryAttrResponse, 0, len(list))
	for i := range list {
		result = append(result, *s.toCategoryAttrResponse(&list[i]))
	}
	return pagination, result, nil
}

func (s *categoryAttrService) toCategoryAttrResponse(a *model.ErshouCategoryAttr) *dto.CategoryAttrResponse {
	resp := &dto.CategoryAttrResponse{
		ID:            a.ID,
		CategoryID:    a.CategoryID,
		AttrName:      a.AttrName,
		AttrKey:       a.AttrKey,
		AttrType:      a.AttrType,
		Unit:          a.Unit,
		IsRequired:    a.IsRequired,
		IsFilterable:  a.IsFilterable,
		IsSearchable:  a.IsSearchable,
		DefaultValue:  a.DefaultValue,
		Placeholder:   a.Placeholder,
		Description:   a.Description,
		Status:        a.Status,
		Sort:          a.Sort,
		CreatedAt:     a.CreatedAt,
		Options:       []string{},
	}
	if a.Options != nil {
		var opts []string
		_ = a.Options.Parse(&opts)
		if opts != nil {
			resp.Options = opts
		}
	}
	return resp
}
