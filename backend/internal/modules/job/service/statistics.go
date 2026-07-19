// Package service 数据统计业务逻辑层
// 依据 v3.2.1 架构方案：M 端运营数据 + C 端招聘者/求职者数据
package service

import (
	"time"

	"wuchang-tongcheng/internal/modules/job/dto"
	"wuchang-tongcheng/internal/modules/job/model"

	"gorm.io/gorm"
)

// StatisticsService 数据统计业务接口
type StatisticsService interface {
	// M 端运营总览
	Overview(regionID uint) (*dto.StatisticsResponse, error)
	// C 端招聘者总览
	RecruiterOverview(userID uint) (*dto.RecruiterOverviewResponse, error)
	// C 端求职者总览
	ApplicantOverview(userID uint) (*dto.ApplicantOverviewResponse, error)
	// 热门职位 Top N
	HotJobs(regionID uint, limit int) ([]dto.HotJobResponse, error)
	// 薪资趋势
	SalaryTrend(categoryID uint, days int) (*dto.SalaryTrendResponse, error)
	// 转化漏斗
	ConversionStats(regionID uint, req *dto.DateTrendRequest) (*dto.ConversionStatsResponse, error)
	// 分类统计
	CategoryStats(regionID uint) ([]dto.CategoryStatResponse, error)
	// 地区统计
	RegionStats() ([]dto.RegionStatResponse, error)
	// 职位趋势
	JobTrend(regionID uint, days int) (*dto.JobTrendResponse, error)
}

type statisticsService struct {
	db *gorm.DB
}

// NewStatisticsService 创建数据统计 service 实例
func NewStatisticsService(db *gorm.DB) StatisticsService {
	return &statisticsService{db: db}
}

// Overview M 端运营总览
func (s *statisticsService) Overview(regionID uint) (*dto.StatisticsResponse, error) {
	resp := &dto.StatisticsResponse{}
	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	jobQuery := s.db.Model(&model.Job{})
	if regionID > 0 {
		jobQuery = jobQuery.Where("region_id = ?", regionID)
	}
	// 在招职位总数
	jobQuery.Where("status = ?", model.StatusPublished).Count(&resp.TotalJobs)
	// 今日新增职位
	q := s.db.Model(&model.Job{}).Where("created_at >= ?", todayStart)
	if regionID > 0 {
		q = q.Where("region_id = ?", regionID)
	}
	q.Count(&resp.TodayNewJobs)

	// 公司总数
	companyQuery := s.db.Model(&model.JobCompany{})
	if regionID > 0 {
		companyQuery = companyQuery.Where("region_id = ?", regionID)
	}
	companyQuery.Where("status = ?", model.CompanyStatusApproved).Count(&resp.TotalCompanies)

	// 活跃招聘者数
	activeQ := s.db.Model(&model.Job{}).Where("status = ?", model.StatusPublished)
	if regionID > 0 {
		activeQ = activeQ.Where("region_id = ?", regionID)
	}
	activeQ.Distinct("user_id").Count(&resp.ActiveRecruiters)

	// 简历总数
	resumeQ := s.db.Model(&model.JobResume{})
	if regionID > 0 {
		resumeQ = resumeQ.Where("region_id = ?", regionID)
	}
	resumeQ.Count(&resp.TotalResumes)

	// 投递统计
	appQuery := s.db.Model(&model.JobApplication{})
	if regionID > 0 {
		appQuery = appQuery.Where("region_id = ?", regionID)
	}
	appQuery.Count(&resp.TotalApplications)

	todayAppQ := s.db.Model(&model.JobApplication{}).Where("created_at >= ?", todayStart)
	if regionID > 0 {
		todayAppQ = todayAppQ.Where("region_id = ?", regionID)
	}
	todayAppQ.Count(&resp.TodayApplications)

	// 面试/Offer/入职总数
	interviewQ := s.db.Model(&model.JobInterview{})
	if regionID > 0 {
		interviewQ = interviewQ.Where("region_id = ?", regionID)
	}
	interviewQ.Count(&resp.TotalInterviews)

	offerQ := s.db.Model(&model.JobApplication{}).Where("status >= ?", model.ApplicationStatusOffer)
	if regionID > 0 {
		offerQ = offerQ.Where("region_id = ?", regionID)
	}
	offerQ.Count(&resp.TotalOffers)

	onboardQ := s.db.Model(&model.JobApplication{}).Where("status = ?", model.ApplicationStatusOnboarded)
	if regionID > 0 {
		onboardQ = onboardQ.Where("region_id = ?", regionID)
	}
	onboardQ.Count(&resp.TotalOnboarded)

	// 平均薪资
	var avgResult struct {
		AvgSalary float64
	}
	avgQ := s.db.Model(&model.Job{}).Where("status = ? AND salary_negotiable = false", model.StatusPublished)
	if regionID > 0 {
		avgQ = avgQ.Where("region_id = ?", regionID)
	}
	avgQ.Select("COALESCE(AVG(salary_monthly), 0) AS avg_salary").Scan(&avgResult)
	resp.AvgSalary = avgResult.AvgSalary

	// 转化率
	if resp.TotalApplications > 0 {
		resp.OfferRate = float64(resp.TotalOffers) / float64(resp.TotalApplications) * 100
	}
	if resp.TotalOffers > 0 {
		resp.OnboardRate = float64(resp.TotalOnboarded) / float64(resp.TotalOffers) * 100
	}

	// 举报统计
	reportQ := s.db.Model(&model.JobReport{})
	if regionID > 0 {
		reportQ = reportQ.Where("region_id = ?", regionID)
	}
	reportQ.Count(&resp.TotalReports)
	s.db.Model(&model.JobReport{}).Where("status = ?", model.ReportStatusPending).Count(&resp.PendingReports)

	return resp, nil
}

// RecruiterOverview C 端招聘者总览
func (s *statisticsService) RecruiterOverview(userID uint) (*dto.RecruiterOverviewResponse, error) {
	resp := &dto.RecruiterOverviewResponse{UserID: userID}
	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	// 发布职位数
	s.db.Model(&model.Job{}).Where("user_id = ?", userID).Count(&resp.TotalJobs)
	// 在招职位数
	s.db.Model(&model.Job{}).Where("user_id = ? AND status = ?", userID, model.StatusPublished).Count(&resp.ActiveJobs)
	// 收到投递数
	s.db.Model(&model.JobApplication{}).Where("recruiter_id = ?", userID).Count(&resp.TotalApplications)
	// 今日投递数
	s.db.Model(&model.JobApplication{}).Where("recruiter_id = ? AND created_at >= ?", userID, todayStart).Count(&resp.TodayApplications)
	// 未读投递数
	s.db.Model(&model.JobApplication{}).Where("recruiter_id = ? AND status = ?", userID, model.ApplicationStatusDelivered).Count(&resp.UnreadApplications)
	// 面试数
	s.db.Model(&model.JobInterview{}).Where("recruiter_id = ?", userID).Count(&resp.TotalInterviews)
	// 已发 Offer 数
	s.db.Model(&model.JobApplication{}).Where("recruiter_id = ? AND status >= ?", userID, model.ApplicationStatusOffer).Count(&resp.TotalOffers)
	// 入职数
	s.db.Model(&model.JobApplication{}).Where("recruiter_id = ? AND status = ?", userID, model.ApplicationStatusOnboarded).Count(&resp.TotalOnboarded)
	// 转化率
	if resp.TotalApplications > 0 {
		resp.OfferRate = float64(resp.TotalOffers) / float64(resp.TotalApplications) * 100
	}
	if resp.TotalOffers > 0 {
		resp.OnboardRate = float64(resp.TotalOnboarded) / float64(resp.TotalOffers) * 100
	}
	// 平均薪资
	var avgResult struct {
		AvgSalary float64
	}
	s.db.Model(&model.Job{}).Where("user_id = ? AND salary_negotiable = false", userID).
		Select("COALESCE(AVG(salary_monthly), 0) AS avg_salary").Scan(&avgResult)
	resp.AvgSalary = avgResult.AvgSalary
	// 职位被收藏数
	s.db.Model(&model.JobFavorite{}).Where("favorite_type = ?", model.FavoriteTypeJob).
		Joins("JOIN jobs ON jobs.id = job_favorites.job_id").
		Where("jobs.user_id = ?", userID).Count(&resp.TotalFavs)
	// 职位被浏览数
	s.db.Model(&model.JobView{}).Where("view_type = ?", model.ViewTypeJob).
		Joins("JOIN jobs ON jobs.id = job_views.job_id").
		Where("jobs.user_id = ?", userID).Count(&resp.TotalViews)

	return resp, nil
}

// ApplicantOverview C 端求职者总览
func (s *statisticsService) ApplicantOverview(userID uint) (*dto.ApplicantOverviewResponse, error) {
	resp := &dto.ApplicantOverviewResponse{UserID: userID}
	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	// 投递总数
	s.db.Model(&model.JobApplication{}).Where("applicant_id = ?", userID).Count(&resp.TotalApplications)
	// 今日投递数
	s.db.Model(&model.JobApplication{}).Where("applicant_id = ? AND created_at >= ?", userID, todayStart).Count(&resp.TodayApplications)
	// 面试数
	s.db.Model(&model.JobInterview{}).Where("applicant_id = ?", userID).Count(&resp.TotalInterviews)
	// 收到 Offer 数
	s.db.Model(&model.JobApplication{}).Where("applicant_id = ? AND status >= ?", userID, model.ApplicationStatusOffer).Count(&resp.TotalOffers)
	// 入职数
	s.db.Model(&model.JobApplication{}).Where("applicant_id = ? AND status = ?", userID, model.ApplicationStatusOnboarded).Count(&resp.TotalOnboarded)
	// 收藏职位数
	s.db.Model(&model.JobFavorite{}).Where("user_id = ? AND favorite_type = ?", userID, model.FavoriteTypeJob).Count(&resp.TotalFavs)
	// 浏览职位数
	s.db.Model(&model.JobView{}).Where("user_id = ? AND view_type = ?", userID, model.ViewTypeJob).Count(&resp.TotalViews)
	// 未读消息数
	s.db.Model(&model.JobMessage{}).Where("to_user_id = ? AND is_read = false AND status = ?", userID, model.MessageStatusNormal).Count(&resp.UnreadMessages)

	return resp, nil
}

// HotJobs 热门职位 Top N
func (s *statisticsService) HotJobs(regionID uint, limit int) ([]dto.HotJobResponse, error) {
	if limit <= 0 {
		limit = 10
	}
	var list []dto.HotJobResponse
	q := s.db.Model(&model.Job{}).
		Select("jobs.id AS job_id, jobs.title, jobs.work_city, jobs.salary_min, jobs.salary_max, jobs.salary_unit, "+
			"jobs.view_count, jobs.fav_count, jobs.deliver_count, jobs.interview_count, jobs.offer_count").
		Where("jobs.status = ?", model.StatusPublished).
		Order("jobs.view_count DESC, jobs.deliver_count DESC, jobs.id DESC").
		Limit(limit)
	if regionID > 0 {
		q = q.Where("jobs.region_id = ?", regionID)
	}
	if err := q.Scan(&list).Error; err != nil {
		return nil, err
	}
	// 填充公司名
	for i := range list {
		var companyName string
		s.db.Model(&model.JobCompany{}).
			Joins("JOIN jobs ON jobs.company_id = job_companies.id").
			Where("jobs.id = ?", list[i].JobID).
			Select("job_companies.name").
			Scan(&companyName)
		list[i].CompanyName = companyName
		list[i].Rank = i + 1
	}
	return list, nil
}

// SalaryTrend 薪资趋势
func (s *statisticsService) SalaryTrend(categoryID uint, days int) (*dto.SalaryTrendResponse, error) {
	if days <= 0 {
		days = 30
	}
	if days > 365 {
		days = 365
	}
	resp := &dto.SalaryTrendResponse{CategoryID: categoryID}
	if categoryID > 0 {
		// 暂不实现分类名查询（分类表为 job_categories）
	}
	now := time.Now()
	startDate := now.AddDate(0, 0, -days)

	type trendRow struct {
		Date      string
		AvgSalary float64
	}
	var rows []trendRow
	query := s.db.Model(&model.Job{}).
		Select("TO_CHAR(created_at, 'YYYY-MM-DD') AS date, COALESCE(AVG(salary_monthly), 0) AS avg_salary").
		Where("created_at >= ? AND status = ? AND salary_negotiable = false", startDate, model.StatusPublished).
		Group("TO_CHAR(created_at, 'YYYY-MM-DD')").
		Order("date ASC")
	if categoryID > 0 {
		query = query.Where("category_id = ?", categoryID)
	}
	if err := query.Scan(&rows).Error; err != nil {
		return nil, err
	}
	for _, r := range rows {
		resp.Dates = append(resp.Dates, r.Date)
		resp.AvgSalaries = append(resp.AvgSalaries, r.AvgSalary)
		resp.MedianSalaries = append(resp.MedianSalaries, r.AvgSalary) // MVP：暂以均值代替中位数
	}
	return resp, nil
}

// ConversionStats 转化漏斗
func (s *statisticsService) ConversionStats(regionID uint, req *dto.DateTrendRequest) (*dto.ConversionStatsResponse, error) {
	resp := &dto.ConversionStatsResponse{}
	startDate := req.StartDate
	endDate := req.EndDate
	if startDate.IsZero() {
		startDate = time.Now().AddDate(0, 0, -30)
	}
	if endDate.IsZero() {
		endDate = time.Now()
	}

	// 浏览数（impression 用 view_count 累加）
	viewQ := s.db.Model(&model.Job{}).
		Where("created_at >= ? AND created_at <= ?", startDate, endDate)
	if regionID > 0 {
		viewQ = viewQ.Where("region_id = ?", regionID)
	}
	var viewSum struct {
		Total int64
	}
	viewQ.Select("COALESCE(SUM(view_count), 0) AS total").Scan(&viewSum)
	resp.ImpressionCount = viewSum.Total

	// 收藏数
	favQ := s.db.Model(&model.Job{}).
		Where("created_at >= ? AND created_at <= ?", startDate, endDate)
	if regionID > 0 {
		favQ = favQ.Where("region_id = ?", regionID)
	}
	var favSum struct {
		Total int64
	}
	favQ.Select("COALESCE(SUM(fav_count), 0) AS total").Scan(&favSum)
	resp.FavCount = favSum.Total

	// 投递数 / 面试数 / Offer 数 / 入职数
	appQ := s.db.Model(&model.JobApplication{}).
		Where("created_at >= ? AND created_at <= ?", startDate, endDate)
	if regionID > 0 {
		appQ = appQ.Where("region_id = ?", regionID)
	}
	appQ.Count(&resp.DeliverCount)

	interviewQ := s.db.Model(&model.JobInterview{}).
		Where("created_at >= ? AND created_at <= ?", startDate, endDate)
	if regionID > 0 {
		interviewQ = interviewQ.Where("region_id = ?", regionID)
	}
	interviewQ.Count(&resp.InterviewCount)

	offerQ := s.db.Model(&model.JobApplication{}).
		Where("created_at >= ? AND created_at <= ? AND status >= ?", startDate, endDate, model.ApplicationStatusOffer)
	if regionID > 0 {
		offerQ = offerQ.Where("region_id = ?", regionID)
	}
	offerQ.Count(&resp.OfferCount)

	onboardQ := s.db.Model(&model.JobApplication{}).
		Where("created_at >= ? AND created_at <= ? AND status = ?", startDate, endDate, model.ApplicationStatusOnboarded)
	if regionID > 0 {
		onboardQ = onboardQ.Where("region_id = ?", regionID)
	}
	onboardQ.Count(&resp.OnboardingCount)

	// 转化率
	if resp.ImpressionCount > 0 {
		resp.DeliverRate = float64(resp.DeliverCount) / float64(resp.ImpressionCount) * 100
		resp.FavRate = float64(resp.FavCount) / float64(resp.ImpressionCount) * 100
	}
	if resp.DeliverCount > 0 {
		resp.InterviewRate = float64(resp.InterviewCount) / float64(resp.DeliverCount) * 100
		resp.OfferRate = float64(resp.OfferCount) / float64(resp.DeliverCount) * 100
	}
	if resp.OfferCount > 0 {
		resp.OnboardingRate = float64(resp.OnboardingCount) / float64(resp.OfferCount) * 100
	}

	return resp, nil
}

// CategoryStats 分类统计
func (s *statisticsService) CategoryStats(regionID uint) ([]dto.CategoryStatResponse, error) {
	var list []dto.CategoryStatResponse
	q := s.db.Model(&model.Job{}).
		Select("category_id, COUNT(*) AS job_count").
		Where("status = ?", model.StatusPublished).
		Group("category_id").
		Order("job_count DESC")
	if regionID > 0 {
		q = q.Where("region_id = ?", regionID)
	}
	if err := q.Scan(&list).Error; err != nil {
		return nil, err
	}
	// 填充分类名 + 投递数 + 平均薪资
	for i := range list {
		var appCount int64
		s.db.Model(&model.JobApplication{}).
			Joins("JOIN jobs ON jobs.id = job_applications.job_id").
			Where("jobs.category_id = ?", list[i].CategoryID).Count(&appCount)
		list[i].ApplicationCount = appCount

		var avgResult struct {
			AvgSalary float64
		}
		s.db.Model(&model.Job{}).
			Where("category_id = ? AND salary_negotiable = false", list[i].CategoryID).
			Select("COALESCE(AVG(salary_monthly), 0) AS avg_salary").Scan(&avgResult)
		list[i].AvgSalary = avgResult.AvgSalary
	}
	return list, nil
}

// RegionStats 地区统计
func (s *statisticsService) RegionStats() ([]dto.RegionStatResponse, error) {
	var list []dto.RegionStatResponse
	err := s.db.Model(&model.Job{}).
		Select("region_id, COUNT(*) AS job_count").
		Where("status = ?", model.StatusPublished).
		Group("region_id").
		Order("job_count DESC").
		Scan(&list).Error
	if err != nil {
		return nil, err
	}
	for i := range list {
		var companyCount int64
		s.db.Model(&model.JobCompany{}).Where("region_id = ?", list[i].RegionID).Count(&companyCount)
		list[i].CompanyCount = companyCount

		var appCount int64
		s.db.Model(&model.JobApplication{}).Where("region_id = ?", list[i].RegionID).Count(&appCount)
		list[i].ApplicationCount = appCount

		var avgResult struct {
			AvgSalary float64
		}
		s.db.Model(&model.Job{}).
			Where("region_id = ? AND salary_negotiable = false", list[i].RegionID).
			Select("COALESCE(AVG(salary_monthly), 0) AS avg_salary").Scan(&avgResult)
		list[i].AvgSalary = avgResult.AvgSalary
	}
	return list, nil
}

// JobTrend 职位趋势
func (s *statisticsService) JobTrend(regionID uint, days int) (*dto.JobTrendResponse, error) {
	if days <= 0 {
		days = 30
	}
	if days > 365 {
		days = 365
	}
	resp := &dto.JobTrendResponse{}
	now := time.Now()
	startDate := now.AddDate(0, 0, -days)

	// 按天聚合
	type row struct {
		Date string
		Cnt  int64
	}

	// 新增职位
	var jobRows []row
	jobQ := s.db.Model(&model.Job{}).
		Select("TO_CHAR(created_at, 'YYYY-MM-DD') AS date, COUNT(*) AS cnt").
		Where("created_at >= ?", startDate).
		Group("TO_CHAR(created_at, 'YYYY-MM-DD')").
		Order("date ASC")
	if regionID > 0 {
		jobQ = jobQ.Where("region_id = ?", regionID)
	}
	_ = jobQ.Scan(&jobRows).Error

	// 新增投递
	var appRows []row
	appQ := s.db.Model(&model.JobApplication{}).
		Select("TO_CHAR(created_at, 'YYYY-MM-DD') AS date, COUNT(*) AS cnt").
		Where("created_at >= ?", startDate).
		Group("TO_CHAR(created_at, 'YYYY-MM-DD')").
		Order("date ASC")
	if regionID > 0 {
		appQ = appQ.Where("region_id = ?", regionID)
	}
	_ = appQ.Scan(&appRows).Error

	// 新增 Offer
	var offerRows []row
	offerQ := s.db.Model(&model.JobApplication{}).
		Select("TO_CHAR(offer_at, 'YYYY-MM-DD') AS date, COUNT(*) AS cnt").
		Where("offer_at >= ? AND status >= ?", startDate, model.ApplicationStatusOffer).
		Group("TO_CHAR(offer_at, 'YYYY-MM-DD')").
		Order("date ASC")
	if regionID > 0 {
		offerQ = offerQ.Where("region_id = ?", regionID)
	}
	_ = offerQ.Scan(&offerRows).Error

	// 生成日期序列，确保所有日期都出现
	jobMap := map[string]int64{}
	appMap := map[string]int64{}
	offerMap := map[string]int64{}
	for _, r := range jobRows {
		jobMap[r.Date] = r.Cnt
	}
	for _, r := range appRows {
		appMap[r.Date] = r.Cnt
	}
	for _, r := range offerRows {
		offerMap[r.Date] = r.Cnt
	}
	for i := 0; i < days; i++ {
		d := startDate.AddDate(0, 0, i).Format("2006-01-02")
		resp.Dates = append(resp.Dates, d)
		resp.NewJobs = append(resp.NewJobs, jobMap[d])
		resp.NewApps = append(resp.NewApps, appMap[d])
		resp.NewOffers = append(resp.NewOffers, offerMap[d])
	}
	return resp, nil
}
