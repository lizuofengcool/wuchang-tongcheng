// Package dto 同城零工兼职数据传输对象 - 总览统计（M 端）
// 依据需求：GET /api/v1/linggong/admin/statistics/overview
// 聚合多张业务表（linggongs/employers/workers/applications/contracts/payments/skills）
// 字段对齐前端 statistics.vue（loadOverview -> overview.total_jobs/today_new/total_employers/...）
package dto

// OverviewStats 总览统计响应
// 字段命名优先对齐前端 statistics.vue 的 overview.total_jobs/today_new/total_employers/total_workers/total_amount/completion_rate
// 同时包含任务描述要求的 total_linggongs/total_applications/total_contracts/total_payments/pending_audits/week_new/month_new/hot_skills/top_employers
type OverviewStats struct {
	// ===== 计数类 =====
	TotalLinggongs    int64 `json:"total_linggongs"`    // linggongs 总数
	TotalJobs         int64 `json:"total_jobs"`         // 兼容前端字段（= total_linggongs）
	TotalEmployers    int64 `json:"total_employers"`   // linggong_employers 总数
	TotalWorkers      int64 `json:"total_workers"`     // linggong_workers 总数
	TotalApplications int64 `json:"total_applications"` // linggong_applications 总数
	TotalContracts    int64 `json:"total_contracts"`    // linggong_contracts 总数
	TotalPayments     int64 `json:"total_payments"`    // linggong_payments 总数
	PendingAudits    int64 `json:"pending_audits"`     // audit_status=0 的 linggongs 数量

	// ===== 时间趋势 =====
	TodayNew int64 `json:"today_new"` // linggongs 今日新增
	WeekNew  int64 `json:"week_new"`  // linggongs 本周新增
	MonthNew int64 `json:"month_new"` // linggongs 本月新增

	// ===== 业务汇总 =====
	TotalAmount    float64 `json:"total_amount"`     // 成交总额（linggong_payments 已支付金额累计）
	CompletionRate float64 `json:"completion_rate"`  // 完成率（linggongs status=7 / total_linggongs）

	// ===== 排行榜 =====
	HotSkills    []OverviewHotSkill    `json:"hot_skills"`     // 热门技能 Top10（按 hot_score 降序）
	TopEmployers []OverviewTopEmployer  `json:"top_employers"` // 雇主排行 Top10（按 published_count 降序）
}

// OverviewHotSkill 热门技能项
type OverviewHotSkill struct {
	ID           uint    `json:"id"`
	Name         string  `json:"name"`
	Category     string  `json:"category"`
	UsageCount   int     `json:"usage_count"`   // = linggong_count（关联岗位数）
	HotScore     int     `json:"hot_score"`
	WorkerCount  int     `json:"worker_count"`
}

// OverviewTopEmployer 雇主排行项
type OverviewTopEmployer struct {
	ID            uint    `json:"id"`
	UserID        uint    `json:"user_id"`
	CompanyName   string  `json:"company_name"`
	ContactName   string  `json:"contact_name"`
	JobCount      int     `json:"job_count"`      // = published_count
	AppliedCount  int     `json:"applied_count"`  // 兼容前端字段（暂用 total_workers 近似）
	Verified      bool    `json:"verified"`       // status=1 视为已认证
	Level         int     `json:"level"`
	AvgRating     float64 `json:"avg_rating"`
}
