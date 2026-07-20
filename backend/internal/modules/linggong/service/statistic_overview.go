// Package service 同城零工兼职业务逻辑层 - 总览统计（M 端）
// 依据需求：GET /api/v1/linggong/admin/statistics/overview
// 聚合多张业务表，单条 SQL 计数 + 排行榜查询
// 不复用 linggong_statistics 表（该表用于按日/类型/目标的明细统计），此服务直接跨表聚合
package service

import (
	"wuchang-tongcheng/internal/modules/linggong/dto"

	"gorm.io/gorm"
)

// OverviewService 总览统计服务接口
type OverviewService interface {
	Overview() (*dto.OverviewStats, error)
}

type overviewService struct {
	db *gorm.DB
}

// NewOverviewService 创建总览统计服务实例
func NewOverviewService(db *gorm.DB) OverviewService {
	return &overviewService{db: db}
}

// Overview 总览统计
// 聚合表：linggongs / linggong_employers / linggong_workers / linggong_applications /
//
//	linggong_contracts / linggong_payments / linggong_skills
func (s *overviewService) Overview() (*dto.OverviewStats, error) {
	stats := &dto.OverviewStats{}

	// ===== 计数类（单条 SQL 完成） =====
	type countResult struct {
		TotalLinggongs    int64
		TotalEmployers    int64
		TotalWorkers      int64
		TotalApplications int64
		TotalContracts    int64
		TotalPayments     int64
		PendingAudits     int64
		TodayNew         int64
		WeekNew           int64
		MonthNew         int64
		CompletedLinggongs int64
	}
	var cr countResult
	// 使用子查询聚合各表计数（PostgreSQL 支持）
	// deleted_at IS NULL 过滤软删除（GORM 软删除字段在 raw SQL 中需手动处理）
	if err := s.db.Raw(`
		SELECT
			(SELECT COUNT(*) FROM linggongs WHERE deleted_at IS NULL) AS total_linggongs,
			(SELECT COUNT(*) FROM linggong_employers WHERE deleted_at IS NULL) AS total_employers,
			(SELECT COUNT(*) FROM linggong_workers WHERE deleted_at IS NULL) AS total_workers,
			(SELECT COUNT(*) FROM linggong_applications WHERE deleted_at IS NULL) AS total_applications,
			(SELECT COUNT(*) FROM linggong_contracts WHERE deleted_at IS NULL) AS total_contracts,
			(SELECT COUNT(*) FROM linggong_payments WHERE deleted_at IS NULL) AS total_payments,
			(SELECT COUNT(*) FROM linggongs WHERE deleted_at IS NULL AND audit_status = 0) AS pending_audits,
			(SELECT COUNT(*) FROM linggongs WHERE deleted_at IS NULL AND created_at >= date_trunc('day', NOW())) AS today_new,
			(SELECT COUNT(*) FROM linggongs WHERE deleted_at IS NULL AND created_at >= date_trunc('week', NOW())) AS week_new,
			(SELECT COUNT(*) FROM linggongs WHERE deleted_at IS NULL AND created_at >= date_trunc('month', NOW())) AS month_new,
			(SELECT COUNT(*) FROM linggongs WHERE deleted_at IS NULL AND status = 7) AS completed_linggongs
	`).Scan(&cr).Error; err != nil {
		return nil, err
	}

	stats.TotalLinggongs = cr.TotalLinggongs
	stats.TotalJobs = cr.TotalLinggongs // 兼容前端字段
	stats.TotalEmployers = cr.TotalEmployers
	stats.TotalWorkers = cr.TotalWorkers
	stats.TotalApplications = cr.TotalApplications
	stats.TotalContracts = cr.TotalContracts
	stats.TotalPayments = cr.TotalPayments
	stats.PendingAudits = cr.PendingAudits
	stats.TodayNew = cr.TodayNew
	stats.WeekNew = cr.WeekNew
	stats.MonthNew = cr.MonthNew

	// ===== 成交总额（linggong_payments status>=1 已支付累计） =====
	// 任务描述要求 total_payments 是数量，total_amount 是金额（前端字段）
	var totalAmount float64
	if err := s.db.Raw(`
		SELECT COALESCE(SUM(amount), 0)
		FROM linggong_payments
		WHERE deleted_at IS NULL AND status >= 1
	`).Scan(&totalAmount).Error; err != nil {
		return nil, err
	}
	stats.TotalAmount = totalAmount

	// ===== 完成率（status=7 已完成 / 总数） =====
	if stats.TotalLinggongs > 0 {
		stats.CompletionRate = float64(cr.CompletedLinggongs) / float64(stats.TotalLinggongs) * 100
	}

	// ===== 热门技能 Top10（按 hot_score 降序） =====
	var hotSkills []dto.OverviewHotSkill
	if err := s.db.Raw(`
		SELECT id, name, category, linggong_count AS usage_count, hot_score, worker_count
		FROM linggong_skills
		WHERE deleted_at IS NULL AND status = 1
		ORDER BY hot_score DESC, linggong_count DESC, id ASC
		LIMIT 10
	`).Scan(&hotSkills).Error; err != nil {
		return nil, err
	}
	if hotSkills == nil {
		hotSkills = []dto.OverviewHotSkill{}
	}
	stats.HotSkills = hotSkills

	// ===== 雇主排行 Top10（按 published_count 降序） =====
	var topEmployers []dto.OverviewTopEmployer
	if err := s.db.Raw(`
		SELECT id, user_id, company_name, contact_name,
		       published_count AS job_count,
		       total_workers AS applied_count,
		       (status = 1) AS verified,
		       level, avg_rating
		FROM linggong_employers
		WHERE deleted_at IS NULL
		ORDER BY published_count DESC, id ASC
		LIMIT 10
	`).Scan(&topEmployers).Error; err != nil {
		return nil, err
	}
	if topEmployers == nil {
		topEmployers = []dto.OverviewTopEmployer{}
	}
	stats.TopEmployers = topEmployers

	return stats, nil
}
