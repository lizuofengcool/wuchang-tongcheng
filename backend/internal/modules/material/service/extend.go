// Package service 素材中台扩展业务逻辑层
// 依据 014_material_full.sql：分类/标签/图片标签/搜索历史/相似结果/OCR
package service

import (
	"errors"
	"time"

	"wuchang-tongcheng/internal/modules/material/dto"
	"wuchang-tongcheng/internal/modules/material/model"
	"wuchang-tongcheng/internal/modules/material/repository"

	"gorm.io/gorm"
)

// 扩展错误
var (
	ErrCategoryNotFound     = errors.New("素材分类不存在")
	ErrTagNotFound           = errors.New("素材标签不存在")
	ErrImageTagNotFound     = errors.New("图片标签关联不存在")
	ErrOCRResultNotFound    = errors.New("OCR 结果不存在")
)

// MaterialExtendService 素材扩展业务接口
type MaterialExtendService interface {
	// 分类
	CreateCategory(regionID uint, req *dto.CreateCategoryRequest) (*dto.CategoryInfo, error)
	UpdateCategory(id uint, req *dto.UpdateCategoryRequest) error
	DeleteCategory(id uint) error
	ListCategories(parentID uint, status, page, pageSize int) ([]dto.CategoryInfo, int64, error)

	// 标签
	CreateTag(regionID uint, req *dto.CreateTagRequest) (*dto.TagInfo, error)
	UpdateTag(id uint, fields map[string]interface{}) error
	DeleteTag(id uint) error
	ListTags(tagType string, page, pageSize int) ([]dto.TagInfo, int64, error)

	// 图片标签
	AddImageTags(regionID uint, req *dto.AddImageTagsRequest) error
	RemoveImageTag(req *dto.RemoveImageTagRequest) error
	ListImageTags(imageID uint) ([]dto.TagInfo, error)
	ListImagesByTag(tagID uint, page, pageSize int) ([]dto.SimilarResultInfo, int64, error)

	// 搜索历史
	ListMySearchHistory(userID uint, page, pageSize int) ([]dto.SearchHistoryInfo, int64, error)
	RecordSearch(userID, regionID uint, searchType string, queryImageID uint, queryText string, resultCount int, topSim float64, costMs int) error

	// 相似结果
	ListSimilarResults(sourceImageID uint, limit int) ([]dto.SimilarResultInfo, error)

	// OCR
	RecognizeOCR(regionID uint, req *dto.OCRRequest) (*dto.OCRResultInfo, error)
	GetOCRByImageID(imageID uint) (*dto.OCRResultInfo, error)
	ListOCRResults(page, pageSize int) ([]dto.OCRResultInfo, int64, error)

	// 统计（M 端）
	Statistics() (*dto.MaterialStatisticsResponse, error)
}

type materialExtendService struct {
	repo    repository.MaterialRepository
	extRepo repository.MaterialExtendRepository
}

// NewMaterialExtendService 创建扩展 service 实例
func NewMaterialExtendService(repo repository.MaterialRepository, extRepo repository.MaterialExtendRepository) MaterialExtendService {
	return &materialExtendService{repo: repo, extRepo: extRepo}
}

// ===== 分类 =====

func (s *materialExtendService) CreateCategory(regionID uint, req *dto.CreateCategoryRequest) (*dto.CategoryInfo, error) {
	level := 1
	if req.ParentID > 0 {
		parent, err := s.extRepo.FindCategoryByID(req.ParentID)
		if err != nil {
			return nil, ErrCategoryNotFound
		}
		level = parent.Level + 1
	}
	status := req.Status
	if status == 0 {
		status = 1
	}
	c := &model.Category{
		Name:        req.Name,
		ParentID:    req.ParentID,
		Level:       level,
		Icon:        req.Icon,
		Description: req.Description,
		Sort:        req.Sort,
		Status:      status,
	}
	c.RegionID = regionID
	if err := s.extRepo.CreateCategory(c); err != nil {
		return nil, err
	}
	return toCategoryInfo(c), nil
}

func (s *materialExtendService) UpdateCategory(id uint, req *dto.UpdateCategoryRequest) error {
	c, err := s.extRepo.FindCategoryByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrCategoryNotFound
		}
		return err
	}
	_ = c
	fields := map[string]interface{}{}
	if req.Name != "" {
		fields["name"] = req.Name
	}
	if req.Icon != "" {
		fields["icon"] = req.Icon
	}
	if req.Description != "" {
		fields["description"] = req.Description
	}
	if req.Sort != 0 {
		fields["sort"] = req.Sort
	}
	if req.Status != 0 {
		fields["status"] = req.Status
	}
	if len(fields) == 0 {
		return nil
	}
	return s.extRepo.UpdateCategoryFields(id, fields)
}

func (s *materialExtendService) DeleteCategory(id uint) error {
	if _, err := s.extRepo.FindCategoryByID(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrCategoryNotFound
		}
		return err
	}
	return s.extRepo.DeleteCategory(id)
}

func (s *materialExtendService) ListCategories(parentID uint, status, page, pageSize int) ([]dto.CategoryInfo, int64, error) {
	list, total, err := s.extRepo.ListCategories(parentID, status, page, pageSize)
	if err != nil {
		return nil, 0, err
	}
	result := make([]dto.CategoryInfo, 0, len(list))
	for i := range list {
		result = append(result, *toCategoryInfo(&list[i]))
	}
	return result, total, nil
}

// ===== 标签 =====

func (s *materialExtendService) CreateTag(regionID uint, req *dto.CreateTagRequest) (*dto.TagInfo, error) {
	tagType := req.TagType
	if tagType == "" {
		tagType = model.TagTypeGeneral
	}
	status := req.Status
	if status == 0 {
		status = 1
	}
	t := &model.Tag{
		Name:        req.Name,
		TagType:     tagType,
		Description: req.Description,
		Icon:        req.Icon,
		Status:      status,
	}
	t.RegionID = regionID
	if err := s.extRepo.CreateTag(t); err != nil {
		return nil, err
	}
	return toTagInfo(t), nil
}

func (s *materialExtendService) UpdateTag(id uint, fields map[string]interface{}) error {
	if _, err := s.extRepo.FindTagByID(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrTagNotFound
		}
		return err
	}
	return s.extRepo.UpdateTagFields(id, fields)
}

func (s *materialExtendService) DeleteTag(id uint) error {
	if _, err := s.extRepo.FindTagByID(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrTagNotFound
		}
		return err
	}
	return s.extRepo.DeleteTag(id)
}

func (s *materialExtendService) ListTags(tagType string, page, pageSize int) ([]dto.TagInfo, int64, error) {
	list, total, err := s.extRepo.ListTags(tagType, page, pageSize)
	if err != nil {
		return nil, 0, err
	}
	result := make([]dto.TagInfo, 0, len(list))
	for i := range list {
		result = append(result, *toTagInfo(&list[i]))
	}
	return result, total, nil
}

// ===== 图片标签 =====

func (s *materialExtendService) AddImageTags(regionID uint, req *dto.AddImageTagsRequest) error {
	source := req.Source
	if source == "" {
		source = model.TagSourceManual
	}
	for _, tagID := range req.TagIDs {
		it := &model.ImageTag{
			ImageID:    req.ImageID,
			TagID:      tagID,
			Source:     source,
			Confidence: 1.0,
		}
		it.RegionID = regionID
		_ = s.extRepo.CreateImageTag(it)
		_ = s.extRepo.IncrTagUsageCount(tagID)
	}
	_ = s.extRepo.IncrCategoryImageCount(0)
	return nil
}

func (s *materialExtendService) RemoveImageTag(req *dto.RemoveImageTagRequest) error {
	return s.extRepo.DeleteImageTag(req.ImageID, req.TagID)
}

func (s *materialExtendService) ListImageTags(imageID uint) ([]dto.TagInfo, error) {
	list, err := s.extRepo.ListImageTagsByImage(imageID)
	if err != nil {
		return nil, err
	}
	result := make([]dto.TagInfo, 0, len(list))
	for _, it := range list {
		t, err := s.extRepo.FindTagByID(it.TagID)
		if err != nil {
			continue
		}
		result = append(result, *toTagInfo(t))
	}
	return result, nil
}

func (s *materialExtendService) ListImagesByTag(tagID uint, page, pageSize int) ([]dto.SimilarResultInfo, int64, error) {
	list, total, err := s.extRepo.ListImageTagsByTag(tagID, page, pageSize)
	if err != nil {
		return nil, 0, err
	}
	result := make([]dto.SimilarResultInfo, 0, len(list))
	for _, it := range list {
		result = append(result, dto.SimilarResultInfo{
			ID:            it.ID,
			SourceImageID: 0,
			TargetImageID: it.ImageID,
			Similarity:    it.Confidence,
			FeatureAlgo:   it.Source,
		})
	}
	return result, total, nil
}

// ===== 搜索历史 =====

func (s *materialExtendService) ListMySearchHistory(userID uint, page, pageSize int) ([]dto.SearchHistoryInfo, int64, error) {
	list, total, err := s.extRepo.ListSearchHistory(userID, page, pageSize)
	if err != nil {
		return nil, 0, err
	}
	result := make([]dto.SearchHistoryInfo, 0, len(list))
	for i := range list {
		result = append(result, *toSearchHistoryInfo(&list[i]))
	}
	return result, total, nil
}

func (s *materialExtendService) RecordSearch(userID, regionID uint, searchType string, queryImageID uint, queryText string, resultCount int, topSim float64, costMs int) error {
	h := &model.SearchHistory{
		UserID:         userID,
		SearchType:    searchType,
		QueryImageID:  queryImageID,
		QueryText:     queryText,
		ResultCount:   resultCount,
		TopSimilarity: topSim,
		CostMs:        costMs,
	}
	h.RegionID = regionID
	return s.extRepo.CreateSearchHistory(h)
}

// ===== 相似结果 =====

func (s *materialExtendService) ListSimilarResults(sourceImageID uint, limit int) ([]dto.SimilarResultInfo, error) {
	list, err := s.extRepo.ListSimilarResults(sourceImageID, limit)
	if err != nil {
		return nil, err
	}
	result := make([]dto.SimilarResultInfo, 0, len(list))
	for i := range list {
		result = append(result, *toSimilarResultInfo(&list[i]))
	}
	return result, nil
}

// ===== OCR =====

func (s *materialExtendService) RecognizeOCR(regionID uint, req *dto.OCRRequest) (*dto.OCRResultInfo, error) {
	engine := req.Engine
	if engine == "" {
		engine = model.OCREngineLocal
	}
	language := req.Language
	if language == "" {
		language = "zh"
	}
	// 模拟 OCR 识别过程（实际应调用 aliyun/tencent OCR）
	start := time.Now()
	textContent := "" // TODO: 接入真实 OCR 引擎
	costMs := int(time.Since(start).Milliseconds())
	r := &model.OCRResult{
		ImageID:     req.ImageID,
		OCREngine:   engine,
		TextContent: textContent,
		TextBlocks:  "[]",
		Language:    language,
		Confidence:  0.0,
		CostMs:      costMs,
	}
	r.RegionID = regionID
	if err := s.extRepo.CreateOCRResult(r); err != nil {
		return nil, err
	}
	return toOCRResultInfo(r), nil
}

func (s *materialExtendService) GetOCRByImageID(imageID uint) (*dto.OCRResultInfo, error) {
	r, err := s.extRepo.FindOCRResultByImageID(imageID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrOCRResultNotFound
		}
		return nil, err
	}
	return toOCRResultInfo(r), nil
}

func (s *materialExtendService) ListOCRResults(page, pageSize int) ([]dto.OCRResultInfo, int64, error) {
	list, total, err := s.extRepo.ListOCRResults(page, pageSize)
	if err != nil {
		return nil, 0, err
	}
	result := make([]dto.OCRResultInfo, 0, len(list))
	for i := range list {
		result = append(result, *toOCRResultInfo(&list[i]))
	}
	return result, total, nil
}

// ===== 统计 =====

func (s *materialExtendService) Statistics() (*dto.MaterialStatisticsResponse, error) {
	resp := &dto.MaterialStatisticsResponse{}
	resp.TotalFiles, _ = s.extRepo.StatTotalFiles()
	resp.TotalImages, _ = s.extRepo.StatTotalImages()
	resp.TotalVideos, _ = s.extRepo.StatTotalVideos()
	resp.TotalCategories, _ = s.extRepo.StatTotalCategories()
	resp.TotalTags, _ = s.extRepo.StatTotalTags()
	resp.TotalSearches, _ = s.extRepo.StatTotalSearches()
	resp.TotalOCR, _ = s.extRepo.StatTotalOCR()
	resp.StorageSize, _ = s.extRepo.StatStorageSize()
	return resp, nil
}

// ===== 工具函数 =====

func toCategoryInfo(c *model.Category) *dto.CategoryInfo {
	return &dto.CategoryInfo{
		ID:          c.ID,
		Name:        c.Name,
		ParentID:    c.ParentID,
		Level:       c.Level,
		Icon:        c.Icon,
		Description: c.Description,
		Sort:        c.Sort,
		ImageCount:  c.ImageCount,
		Status:      c.Status,
		CreatedAt:   c.CreatedAt,
	}
}

func toTagInfo(t *model.Tag) *dto.TagInfo {
	return &dto.TagInfo{
		ID:          t.ID,
		Name:        t.Name,
		TagType:     t.TagType,
		Description: t.Description,
		Icon:        t.Icon,
		UsageCount:  t.UsageCount,
		Status:      t.Status,
		CreatedAt:   t.CreatedAt,
	}
}

func toSearchHistoryInfo(h *model.SearchHistory) *dto.SearchHistoryInfo {
	return &dto.SearchHistoryInfo{
		ID:             h.ID,
		UserID:         h.UserID,
		SearchType:     h.SearchType,
		QueryImageID:   h.QueryImageID,
		QueryText:      h.QueryText,
		ResultCount:    h.ResultCount,
		TopSimilarity:  h.TopSimilarity,
		CostMs:         h.CostMs,
		CreatedAt:      h.CreatedAt,
	}
}

func toSimilarResultInfo(r *model.SimilarResult) *dto.SimilarResultInfo {
	return &dto.SimilarResultInfo{
		ID:            r.ID,
		SourceImageID: r.SourceImageID,
		TargetImageID: r.TargetImageID,
		Similarity:    r.Similarity,
		FeatureAlgo:   r.FeatureAlgo,
	}
}

func toOCRResultInfo(r *model.OCRResult) *dto.OCRResultInfo {
	return &dto.OCRResultInfo{
		ID:          r.ID,
		ImageID:     r.ImageID,
		FileID:      r.FileID,
		OCREngine:   r.OCREngine,
		TextContent: r.TextContent,
		TextBlocks:  r.TextBlocks,
		Language:    r.Language,
		Confidence:  r.Confidence,
		CostMs:      r.CostMs,
		CreatedAt:   r.CreatedAt,
	}
}
