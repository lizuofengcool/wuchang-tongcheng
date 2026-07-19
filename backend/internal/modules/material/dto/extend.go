// Package dto 素材中台扩展数据传输对象
package dto

import "time"

// CategoryInfo 分类信息
type CategoryInfo struct {
	ID          uint      `json:"id"`
	Name        string    `json:"name"`
	ParentID    uint      `json:"parent_id"`
	Level       int       `json:"level"`
	Icon        string    `json:"icon"`
	Description string    `json:"description"`
	Sort        int       `json:"sort"`
	ImageCount  int       `json:"image_count"`
	Status      int       `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
}

// CreateCategoryRequest 创建分类请求
type CreateCategoryRequest struct {
	Name        string `json:"name" binding:"required,max=64"`
	ParentID    uint   `json:"parent_id"`
	Icon        string `json:"icon" binding:"max=256"`
	Description string `json:"description" binding:"max=256"`
	Sort        int    `json:"sort"`
	Status      int    `json:"status"`
}

// UpdateCategoryRequest 更新分类请求
type UpdateCategoryRequest struct {
	Name        string `json:"name" binding:"max=64"`
	Icon        string `json:"icon" binding:"max=256"`
	Description string `json:"description" binding:"max=256"`
	Sort        int    `json:"sort"`
	Status      int    `json:"status"`
}

// TagInfo 标签信息
type TagInfo struct {
	ID          uint      `json:"id"`
	Name        string    `json:"name"`
	TagType     string    `json:"tag_type"`
	Description string    `json:"description"`
	Icon        string    `json:"icon"`
	UsageCount  int       `json:"usage_count"`
	Status      int       `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
}

// CreateTagRequest 创建标签请求
type CreateTagRequest struct {
	Name        string `json:"name" binding:"required,max=64"`
	TagType     string `json:"tag_type" binding:"omitempty,oneof=general object scene color shape"`
	Description string `json:"description" binding:"max=256"`
	Icon        string `json:"icon" binding:"max=256"`
	Status      int    `json:"status"`
}

// AddImageTagsRequest 给图片打标签请求
type AddImageTagsRequest struct {
	ImageID uint   `json:"image_id" binding:"required"`
	TagIDs  []uint `json:"tag_ids" binding:"required,min=1"`
	Source  string `json:"source" binding:"omitempty,oneof=manual auto"`
}

// RemoveImageTagRequest 移除图片标签请求
type RemoveImageTagRequest struct {
	ImageID uint `json:"image_id" binding:"required"`
	TagID   uint `json:"tag_id" binding:"required"`
}

// ListImageTagsRequest 查询图片的标签
type ListImageTagsRequest struct {
	ImageID uint `form:"image_id" json:"image_id"`
	TagID   uint `form:"tag_id" json:"tag_id"`
}

// SearchHistoryInfo 搜索历史信息
type SearchHistoryInfo struct {
	ID             uint      `json:"id"`
	UserID         uint      `json:"user_id"`
	SearchType     string    `json:"search_type"`
	QueryImageID   uint      `json:"query_image_id"`
	QueryText      string    `json:"query_text"`
	ResultCount    int       `json:"result_count"`
	TopSimilarity  float64   `json:"top_similarity"`
	CostMs         int       `json:"cost_ms"`
	CreatedAt      time.Time `json:"created_at"`
}

// SimilarResultInfo 相似结果信息
type SimilarResultInfo struct {
	ID              uint    `json:"id"`
	SourceImageID   uint    `json:"source_image_id"`
	TargetImageID   uint    `json:"target_image_id"`
	Similarity      float64 `json:"similarity"`
	FeatureAlgo     string  `json:"feature_algo"`
}

// OCRRequest OCR 识别请求
type OCRRequest struct {
	ImageID  uint   `json:"image_id" binding:"required"`
	Engine   string `json:"engine" binding:"omitempty,oneof=local aliyun tencent baidu"`
	Language string `json:"language" binding:"omitempty,max=16"`
}

// OCRResultInfo OCR 结果信息
type OCRResultInfo struct {
	ID          uint      `json:"id"`
	ImageID     uint      `json:"image_id"`
	FileID      string    `json:"file_id"`
	OCREngine   string    `json:"ocr_engine"`
	TextContent string    `json:"text_content"`
	TextBlocks  string    `json:"text_blocks"`
	Language    string    `json:"language"`
	Confidence  float64   `json:"confidence"`
	CostMs      int       `json:"cost_ms"`
	CreatedAt   time.Time `json:"created_at"`
}

// MaterialStatisticsResponse 素材统计响应
type MaterialStatisticsResponse struct {
	TotalFiles       int64 `json:"total_files"`
	TotalImages      int64 `json:"total_images"`
	TotalVideos      int64 `json:"total_videos"`
	TotalCategories  int64 `json:"total_categories"`
	TotalTags        int64 `json:"total_tags"`
	TotalSearches    int64 `json:"total_searches"`
	TotalOCR         int64 `json:"total_ocr"`
	StorageSize      int64 `json:"storage_size"`
}
