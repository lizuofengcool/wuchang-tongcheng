package utils

import (
	"math"
	"strconv"

	"wuchang-tongcheng/internal/core/response"

	"gorm.io/gorm"
)

const (
	// DefaultPage 默认页码
	DefaultPage = 1
	// DefaultPageSize 默认每页大小
	DefaultPageSize = 10
	// MaxPageSize 最大每页大小
	MaxPageSize = 100
)

// Pagination 分页参数
type Pagination struct {
	Page     int `json:"page" form:"page"`
	PageSize int `json:"page_size" form:"page_size"`
	Total    int64 `json:"total"`
}

// NewPagination 创建分页对象
func NewPagination(page, pageSize int) *Pagination {
	if page <= 0 {
		page = DefaultPage
	}
	if pageSize <= 0 {
		pageSize = DefaultPageSize
	}
	if pageSize > MaxPageSize {
		pageSize = MaxPageSize
	}
	return &Pagination{
		Page:     page,
		PageSize: pageSize,
	}
}

// ParsePagination 从字符串解析分页参数
func ParsePagination(pageStr, pageSizeStr string) *Pagination {
	page, _ := strconv.Atoi(pageStr)
	pageSize, _ := strconv.Atoi(pageSizeStr)
	return NewPagination(page, pageSize)
}

// Offset 获取偏移量
func (p *Pagination) Offset() int {
	return (p.Page - 1) * p.PageSize
}

// Limit 获取限制数量
func (p *Pagination) Limit() int {
	return p.PageSize
}

// TotalPages 获取总页数
func (p *Pagination) TotalPages() int {
	if p.Total == 0 {
		return 0
	}
	return int(math.Ceil(float64(p.Total) / float64(p.PageSize)))
}

// HasNext 是否有下一页
func (p *Pagination) HasNext() bool {
	return p.Page < p.TotalPages()
}

// HasPrev 是否有上一页
func (p *Pagination) HasPrev() bool {
	return p.Page > 1
}

// Paginate GORM分页查询
// 使用示例：
//
//	var users []User
//	var total int64
//	pagination := utils.NewPagination(1, 10)
//	db.Scopes(utils.Paginate(pagination)).Find(&users)
//	db.Model(&User{}).Count(&total)
//	pagination.Total = total
func Paginate(p *Pagination) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Offset(p.Offset()).Limit(p.Limit())
	}
}

// PageResult 生成分页结果
func PageResult(list interface{}, pagination *Pagination) *response.PageResult {
	return response.NewPageResult(list, pagination.Total, pagination.Page, pagination.PageSize)
}

// SortParams 排序参数
type SortParams struct {
	Field string `json:"sort_field" form:"sort_field"`
	Order string `json:"sort_order" form:"sort_order"` // asc, desc
}

// NewSortParams 创建排序参数
func NewSortParams(field, order string) *SortParams {
	if order != "asc" && order != "desc" {
		order = "desc"
	}
	return &SortParams{
		Field: field,
		Order: order,
	}
}

// OrderString 获取排序字符串
func (s *SortParams) OrderString() string {
	if s.Field == "" {
		return ""
	}
	return s.Field + " " + s.Order
}

// SortScope GORM排序作用域
func SortScope(s *SortParams, defaultField string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		field := s.Field
		if field == "" {
			field = defaultField
		}
		order := s.Order
		if order == "" {
			order = "desc"
		}
		return db.Order(field + " " + order)
	}
}
