// Package dto 商家模块数据传输对象
package dto

import "time"

// ShopInfo 店铺信息
type ShopInfo struct {
	ID            uint      `json:"id"`
	Name          string    `json:"name"`
	Logo          string    `json:"logo"`
	Description   string    `json:"description"`
	Phone         string    `json:"phone"`
	Address       string    `json:"address"`
	Longitude     float64   `json:"longitude"`
	Latitude      float64   `json:"latitude"`
	CategoryID    uint      `json:"category_id"`
	BusinessHours string    `json:"business_hours"`
	Status        int       `json:"status"`
	AuditStatus   int       `json:"audit_status"`
	Rating        float32   `json:"rating"`
	Views         int       `json:"views"`
	IsRecommend   int       `json:"is_recommend"`
	Sort          int       `json:"sort"`
	UserID        uint      `json:"user_id"`
	RegionID      uint      `json:"region_id"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// ShopImageInfo 店铺图片信息
type ShopImageInfo struct {
	ID        uint      `json:"id"`
	ShopID    uint      `json:"shop_id"`
	ImageURL  string    `json:"image_url"`
	Sort      int       `json:"sort"`
	CreatedAt time.Time `json:"created_at"`
}

// ShopReviewInfo 店铺评价信息
type ShopReviewInfo struct {
	ID        uint       `json:"id"`
	ShopID    uint       `json:"shop_id"`
	UserID    uint       `json:"user_id"`
	Rating    int        `json:"rating"`
	Content   string     `json:"content"`
	Reply     string     `json:"reply"`
	ReplyAt   *time.Time `json:"reply_at"`
	Status    int        `json:"status"`
	CreatedAt time.Time  `json:"created_at"`
}

// ApplyShopRequest 商家入驻申请
type ApplyShopRequest struct {
	Name          string `json:"name" binding:"required,max=100"`
	Logo          string `json:"logo" binding:"max=255"`
	Description   string `json:"description"`
	Phone         string `json:"phone" binding:"required,max=30"`
	Address       string `json:"address" binding:"max=255"`
	Longitude     float64 `json:"longitude"`
	Latitude      float64 `json:"latitude"`
	CategoryID    uint   `json:"category_id"`
	BusinessHours string `json:"business_hours" binding:"max=50"`
}

// UpdateShopRequest 编辑我的店铺
type UpdateShopRequest struct {
	Name          string `json:"name" binding:"max=100"`
	Logo          string `json:"logo" binding:"max=255"`
	Description   string `json:"description"`
	Phone         string `json:"phone" binding:"max=30"`
	Address       string `json:"address" binding:"max=255"`
	Longitude     float64 `json:"longitude"`
	Latitude      float64 `json:"latitude"`
	CategoryID    uint   `json:"category_id"`
	BusinessHours string `json:"business_hours" binding:"max=50"`
	Status        int    `json:"status" binding:"omitempty,oneof=0 1 2"` // 营业状态
}

// AddShopImageRequest 上传店铺图片（图片URL由文件模块上传后返回）
type AddShopImageRequest struct {
	ImageURL string `json:"image_url" binding:"required,max=255"`
	Sort     int    `json:"sort"`
}

// CreateReviewRequest 发表评价
type CreateReviewRequest struct {
	Rating  int    `json:"rating" binding:"required,min=1,max=5"`
	Content string `json:"content" binding:"max=1000"`
}

// ShopListRequest 店铺列表查询（公开）
type ShopListRequest struct {
	Page       int    `form:"page"`
	PageSize   int    `form:"page_size"`
	CategoryID uint   `form:"category_id"`
	Keyword    string `form:"keyword"`
	RegionID   uint   `form:"region_id"`
	IsRecommend int   `form:"is_recommend"` // -1不筛选 0否 1是
}

// AdminShopListRequest 管理端店铺列表查询
type AdminShopListRequest struct {
	Page        int    `form:"page"`
	PageSize    int    `form:"page_size"`
	CategoryID  uint   `form:"category_id"`
	Keyword     string `form:"keyword"`
	RegionID    uint   `form:"region_id"`
	AuditStatus int    `form:"audit_status"` // -1不筛选
	Status      int    `form:"status"`       // -1不筛选
	IsRecommend int    `form:"is_recommend"` // -1不筛选
}

// AuditShopRequest 审核店铺
type AuditShopRequest struct {
	AuditStatus int    `json:"audit_status" binding:"required,oneof=1 2"` // 1通过 2拒绝
	Reason      string `json:"reason" binding:"max=255"`                  // 拒绝原因（可选，存入简介后缀或日志）
}

// UpdateShopStatusRequest 修改营业状态
type UpdateShopStatusRequest struct {
	Status int `json:"status" binding:"required,oneof=0 1 2"`
}

// SetRecommendRequest 设置推荐
type SetRecommendRequest struct {
	IsRecommend int `json:"is_recommend" binding:"required,oneof=0 1"`
}

// ReviewListRequest 店铺评价列表查询（公开，仅返回已通过）
type ReviewListRequest struct {
	Page     int `form:"page"`
	PageSize int `form:"page_size"`
}

// AdminReviewListRequest 管理端评价列表查询
type AdminReviewListRequest struct {
	Page     int    `form:"page"`
	PageSize int    `form:"page_size"`
	ShopID   uint   `form:"shop_id"`
	Status   int    `form:"status"` // -1不筛选
}

// AuditReviewRequest 审核评价
type AuditReviewRequest struct {
	Status int `json:"status" binding:"required,oneof=1 2"` // 1通过 2拒绝
}
