// Package dto 同城114数据传输对象 - 菜单/服务项目
package dto

import (
	"time"

	"wuchang-tongcheng/internal/pkg/utils"
)

// MenuInfo 菜单/服务项目响应
type MenuInfo struct {
	ID            uint       `json:"id"`
	Dh114ID       uint       `json:"dh114_id"`
	BusinessID    uint       `json:"business_id"`
	MenuType      string     `json:"menu_type"`
	MenuTypeText  string     `json:"menu_type_text"`
	Name          string     `json:"name"`
	Description   string     `json:"description"`
	Price         float64    `json:"price"`
	OriginalPrice float64    `json:"original_price"`
	Image         string     `json:"image"`
	Unit          string     `json:"unit"`
	Sort          int        `json:"sort"`
	Status        int        `json:"status"`
	StatusText    string     `json:"status_text"`
	OrderCount    int        `json:"order_count"`
	Tags          interface{} `json:"tags"`
	IsSignature   bool       `json:"is_signature"`
	RegionID      uint       `json:"region_id"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

// CreateMenuRequest 创建菜单请求
type CreateMenuRequest struct {
	Dh114ID       uint       `json:"dh114_id" binding:"required"`
	MenuType      string     `json:"menu_type" binding:"omitempty,oneof=dish service"`
	Name          string     `json:"name" binding:"required,max=128"`
	Description   string     `json:"description" binding:"max=500"`
	Price         float64    `json:"price" binding:"min=0"`
	OriginalPrice float64    `json:"original_price" binding:"min=0"`
	Image         string     `json:"image" binding:"max=255"`
	Unit          string     `json:"unit" binding:"max=32"`
	Sort          int        `json:"sort"`
	Status        int        `json:"status" binding:"omitempty,oneof=0 1"`
	Tags          interface{} `json:"tags"`
	IsSignature   bool       `json:"is_signature"`
}

// UpdateMenuRequest 更新菜单请求
type UpdateMenuRequest struct {
	MenuType      *string `json:"menu_type" binding:"omitempty,oneof=dish service"`
	Name          *string `json:"name" binding:"max=128"`
	Description   *string `json:"description" binding:"max=500"`
	Price         *float64 `json:"price" binding:"min=0"`
	OriginalPrice *float64 `json:"original_price" binding:"min=0"`
	Image         *string `json:"image" binding:"max=255"`
	Unit          *string `json:"unit" binding:"max=32"`
	Sort          *int    `json:"sort"`
	Status        *int    `json:"status" binding:"omitempty,oneof=0 1"`
	Tags          interface{} `json:"tags"`
	IsSignature   *bool   `json:"is_signature"`
}

// MenuListRequest 菜单列表请求
type MenuListRequest struct {
	Dh114ID     uint   `form:"dh114_id" json:"dh114_id"`
	MenuType    string `form:"menu_type" json:"menu_type"`
	Status      *int   `form:"status" json:"status"`
	IsSignature *bool  `form:"is_signature" json:"is_signature"`
	Keyword     string `form:"keyword" json:"keyword"`
	utils.Pagination
}

// BatchReplaceMenusRequest 批量替换菜单请求
type BatchReplaceMenusRequest struct {
	Dh114ID uint                 `json:"dh114_id" binding:"required"`
	Menus   []CreateMenuRequest  `json:"menus" binding:"required,min=1"`
}

// RecommendationInfo 推荐商家响应
type RecommendationInfo struct {
	ID            uint       `json:"id"`
	UserID        uint       `json:"user_id"`
	Dh114ID       uint       `json:"dh114_id"`
	BusinessID    uint       `json:"business_id"`
	RecommendType string     `json:"recommend_type"`
	Position      int        `json:"position"`
	Score         float64    `json:"score"`
	Reason        string     `json:"reason"`
	CategoryID    *uint      `json:"category_id"`
	ExpireAt      *time.Time `json:"expire_at"`
	Status        int        `json:"status"`
	StatusText    string     `json:"status_text"`
	ClickedAt     *time.Time `json:"clicked_at"`
	ContactedAt   *time.Time `json:"contacted_at"`
	DismissedAt   *time.Time `json:"dismissed_at"`
	RegionID      uint       `json:"region_id"`
	CreatedAt     time.Time  `json:"created_at"`
}

// RecommendationListRequest 推荐列表请求
type RecommendationListRequest struct {
	UserID        uint   `form:"user_id" json:"user_id"`
	Dh114ID       uint   `form:"dh114_id" json:"dh114_id"`
	RecommendType string `form:"recommend_type" json:"recommend_type"`
	CategoryID    uint   `form:"category_id" json:"category_id"`
	Status        *int   `form:"status" json:"status"`
	utils.Pagination
}

// AuditRuleInfo 审核规则响应
type AuditRuleInfo struct {
	ID          uint       `json:"id"`
	RuleName    string     `json:"rule_name"`
	RuleType    string     `json:"rule_type"`
	RuleTypeText string    `json:"rule_type_text"`
	RuleKey     string     `json:"rule_key"`
	Pattern     string     `json:"pattern"`
	Threshold   interface{} `json:"threshold"`
	Action      string     `json:"action"`
	ActionText  string     `json:"action_text"`
	PenaltyType string     `json:"penalty_type"`
	Severity    int        `json:"severity"`
	Status      int        `json:"status"`
	StatusText  string     `json:"status_text"`
	Description string     `json:"description"`
	Sort        int        `json:"sort"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// CreateAuditRuleRequest 创建审核规则请求
type CreateAuditRuleRequest struct {
	RuleName    string      `json:"rule_name" binding:"required,max=128"`
	RuleType    string      `json:"rule_type" binding:"required,oneof=sensitive_word prohibited contact price_check frequency"`
	RuleKey     string      `json:"rule_key" binding:"max=64"`
	Pattern     string      `json:"pattern"`
	Threshold   interface{} `json:"threshold"`
	Action      string      `json:"action" binding:"omitempty,oneof=reject approval filter limit"`
	PenaltyType string      `json:"penalty_type" binding:"max=32"`
	Severity    int         `json:"severity" binding:"omitempty,min=1,max=5"`
	Status      int         `json:"status" binding:"omitempty,oneof=0 1"`
	Description string      `json:"description" binding:"max=500"`
	Sort        int         `json:"sort"`
}

// UpdateAuditRuleRequest 更新审核规则请求
type UpdateAuditRuleRequest struct {
	RuleName    *string `json:"rule_name" binding:"max=128"`
	RuleKey     *string `json:"rule_key" binding:"max=64"`
	Pattern     *string `json:"pattern"`
	Threshold   interface{} `json:"threshold"`
	Action      *string `json:"action" binding:"omitempty,oneof=reject approval filter limit"`
	PenaltyType *string `json:"penalty_type" binding:"max=32"`
	Severity    *int    `json:"severity" binding:"omitempty,min=1,max=5"`
	Status      *int    `json:"status" binding:"omitempty,oneof=0 1"`
	Description *string `json:"description" binding:"max=500"`
	Sort        *int    `json:"sort"`
}

// AuditRuleListRequest 审核规则列表请求
type AuditRuleListRequest struct {
	RuleType string `form:"rule_type" json:"rule_type"`
	RuleKey  string `form:"rule_key" json:"rule_key"`
	Status   *int   `form:"status" json:"status"`
	Keyword  string `form:"keyword" json:"keyword"`
	utils.Pagination
}

// StatisticInfo 统计信息响应
type StatisticInfo struct {
	ID              uint      `json:"id"`
	StatDate        string    `json:"stat_date"`
	StatType        string    `json:"stat_type"`
	StatTypeText    string    `json:"stat_type_text"`
	Dh114ID         uint      `json:"dh114_id"`
	BusinessID      uint      `json:"business_id"`
	CategoryID      *uint     `json:"category_id"`
	ViewCount       int64     `json:"view_count"`
	FavCount        int64     `json:"fav_count"`
	CallCount       int64     `json:"call_count"`
	ShareCount      int64     `json:"share_count"`
	ContactCount    int64     `json:"contact_count"`
	VisitCount      int64     `json:"visit_count"`
	ReviewCount     int64     `json:"review_count"`
	NewReviewCount  int64     `json:"new_review_count"`
	AvgRating       float64   `json:"avg_rating"`
	GoodRate        float64   `json:"good_rate"`
	GroupbuySold    int64     `json:"groupbuy_sold"`
	GroupbuyAmount  float64   `json:"groupbuy_amount"`
	CouponIssued    int64     `json:"coupon_issued"`
	CouponUsed      int64     `json:"coupon_used"`
	OrderCount      int64     `json:"order_count"`
	OrderAmount     float64   `json:"order_amount"`
	RegionID        uint      `json:"region_id"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// OverviewResponse 总览统计响应
type OverviewResponse struct {
	TotalBusiness int64   `json:"total_business"`
	TotalView     int64   `json:"total_view"`
	TotalFav      int64   `json:"total_fav"`
	TotalCall     int64   `json:"total_call"`
	TotalReview   int64   `json:"total_review"`
	TotalGroupbuy int64   `json:"total_groupbuy"`
	TotalCoupon   int64   `json:"total_coupon"`
	TotalOrder    int64   `json:"total_order"`
	TotalAmount   float64 `json:"total_amount"`
	AvgRating     float64 `json:"avg_rating"`
}

// HotBusinessResponse 热门商户响应
type HotBusinessResponse struct {
	Dh114ID     uint    `json:"dh114_id"`
	Title       string  `json:"title"`
	ViewCount   int64   `json:"view_count"`
	FavCount    int64   `json:"fav_count"`
	CallCount   int64   `json:"call_count"`
	ReviewCount int64   `json:"review_count"`
	Rating      float64 `json:"rating"`
}

// HotCategoryResponse 热门分类响应
type HotCategoryResponse struct {
	CategoryID    uint   `json:"category_id"`
	CategoryName  string `json:"category_name"`
	BusinessCount int    `json:"business_count"`
	TotalViews    int64  `json:"total_views"`
	TotalReviews  int64  `json:"total_reviews"`
}

// BusinessHourInfo 营业时间信息
type BusinessHourInfo struct {
	ID         uint   `json:"id"`
	Dh114ID    uint   `json:"dh114_id"`
	BusinessID uint   `json:"business_id"`
	Weekday    int    `json:"weekday"`
	WeekdayText string `json:"weekday_text"`
	OpenTime   string `json:"open_time"`
	CloseTime  string `json:"close_time"`
	IsOpen     bool   `json:"is_open"`
	Is24H      bool   `json:"is_24h"`
}

// BusinessHourItem 营业时间项（用于批量设置）
type BusinessHourItem struct {
	Weekday  int    `json:"weekday" binding:"required,min=1,max=7"`
	OpenTime string `json:"open_time" binding:"max=8"`
	CloseTime string `json:"close_time" binding:"max=8"`
	IsOpen   bool   `json:"is_open"`
	Is24H    bool   `json:"is_24h"`
}

// BatchReplaceHoursRequest 批量替换营业时间请求
type BatchReplaceHoursRequest struct {
	Dh114ID uint                `json:"dh114_id" binding:"required"`
	Hours   []BusinessHourItem  `json:"hours" binding:"required,min=1"`
}

// TagInfo 标签信息
type TagInfo struct {
	ID        uint   `json:"id"`
	TagType   string `json:"tag_type"`
	TagTypeText string `json:"tag_type_text"`
	Name      string `json:"name"`
	Code      string `json:"code"`
	Color     string `json:"color"`
	Icon      string `json:"icon"`
	Sort      int    `json:"sort"`
	Status    int    `json:"status"`
	StatusText string `json:"status_text"`
	UseCount  int    `json:"use_count"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}
