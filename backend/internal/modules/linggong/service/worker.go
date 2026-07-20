// Package service 同城零工兼职业务逻辑层 - 求职者档案
// 对标斗米/兼职猫：求职意向 + 工作经历 + 教育背景 + 技能认证
// 4 维数据隔离（region_id + user_id）
package service

import (
	"errors"
	"time"

	"wuchang-tongcheng/internal/modules/linggong/dto"
	"wuchang-tongcheng/internal/modules/linggong/model"
	"wuchang-tongcheng/internal/modules/linggong/repository"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

var (
	ErrWorkerNotFound     = errors.New("求职者档案不存在")
	ErrWorkerNoPermission = errors.New("无权操作此求职者档案")
	ErrWorkerDuplicate    = errors.New("求职者档案已存在")
	ErrWorkerStatusInvalid = errors.New("求职者档案状态不允许此操作")
)

// WorkerService 求职者档案业务接口
type WorkerService interface {
	// C 端
	Create(regionID uint, userID uint, req *dto.CreateWorkerRequest) (*dto.WorkerInfo, error)
	Update(id uint, operatorID uint, req *dto.UpdateWorkerRequest) error
	Delete(id uint, operatorID uint) error
	GetByID(id uint) (*dto.WorkerInfo, error)
	GetByUserID(userID uint) (*dto.WorkerInfo, error)
	List(regionID uint, req *dto.WorkerListRequest) (*utils.Pagination, []dto.WorkerInfo, error)

	// M 端管理
	AdminList(req *dto.WorkerAdminListRequest) (*utils.Pagination, []dto.WorkerInfo, error)
	Audit(id uint, status int, rejectReason string) error
	UpdateStatus(id uint, status int) error

	// 内部调用（评价/信用/统计）
	UpdateRating(id uint, avgRating float64, ratingCount int) error
	UpdateCreditScore(id uint, score int) error
	IncrAppliedCount(id uint) error
	IncrCompletedCount(id uint) error
	IncrTotalWorkHours(id uint, hours int) error
	IncrTotalEarnings(id uint, amount float64) error
}

type workerService struct {
	repo repository.WorkerRepository
}

// NewWorkerService 创建求职者档案 service 实例
func NewWorkerService(repo repository.WorkerRepository) WorkerService {
	return &workerService{repo: repo}
}

// workerStatusText 求职者档案状态文本
func workerStatusText(s int) string {
	switch s {
	case model.WorkerStatusInactive:
		return "不活跃"
	case model.WorkerStatusActive:
		return "活跃"
	case model.WorkerStatusBanned:
		return "已封禁"
	}
	return ""
}

// workerJobIntentionText 求职意向文本
func workerJobIntentionText(s string) string {
	switch s {
	case model.WorkerIntentFullTime:
		return "全职"
	case model.WorkerIntentPartTime:
		return "兼职"
	case model.WorkerIntentTemp:
		return "临时"
	case model.WorkerIntentRemote:
		return "远程"
	case model.WorkerIntentAny:
		return "任意"
	}
	return ""
}

// workerEducationText 学历文本
func workerEducationText(s string) string {
	switch s {
	case model.EducationBelowHigh:
		return "高中以下"
	case model.EducationHigh:
		return "高中"
	case model.EducationCollege:
		return "大专"
	case model.EducationBachelor:
		return "本科"
	case model.EducationMaster:
		return "硕士"
	case model.EducationDoctor:
		return "博士"
	}
	return ""
}

// workerAgeFromBirthday 根据生日计算年龄
func workerAgeFromBirthday(birthday *time.Time) int {
	if birthday == nil {
		return 0
	}
	now := time.Now()
	age := now.Year() - birthday.Year()
	if now.YearDay() < birthday.YearDay() {
		age--
	}
	if age < 0 {
		age = 0
	}
	return age
}

// toWorkerInfo model -> dto
func toWorkerInfo(w *model.LinggongWorker) *dto.WorkerInfo {
	return &dto.WorkerInfo{
		ID:                   w.ID,
		UserID:               w.UserID,
		RealName:             w.RealName,
		RealNameVerified:     w.RealNameVerified,
		Gender:               w.Gender,
		Birthday:             w.Birthday,
		Age:                  w.Age,
		IDCard:               w.IDCard,
		IDCardVerified:       w.IDCardVerified,
		IDCardFrontURL:       w.IDCardFrontURL,
		IDCardBackURL:        w.IDCardBackURL,
		IDCardHandURL:        w.IDCardHandURL,
		Nickname:             w.Nickname,
		Avatar:               w.Avatar,
		Phone:                w.Phone,
		Email:                w.Email,
		Wechat:               w.Wechat,
		Province:             w.Province,
		City:                 w.City,
		District:             w.District,
		Address:              w.Address,
		Latitude:             w.Latitude,
		Longitude:            w.Longitude,
		Education:            w.Education,
		School:               w.School,
		Major:                w.Major,
		GraduationYear:       w.GraduationYear,
		JobIntention:         w.JobIntention,
		JobIntentionText:     workerJobIntentionText(w.JobIntention),
		ExpectedSalary:       w.ExpectedSalary,
		AvailableTime:        w.AvailableTime,
		AvailableNow:         w.AvailableNow,
		HealthCertURL:        w.HealthCertURL,
		HealthCertValidUntil: w.HealthCertValidUntil,
		HasCriminalRecord:    w.HasCriminalRecord,
		CriminalRecordURL:    w.CriminalRecordURL,
		BankAccount:          w.BankAccount,
		BankName:             w.BankName,
		AlipayAccount:        w.AlipayAccount,
		WechatPayAccount:     w.WechatPayAccount,
		Bio:                  w.Bio,
		SkillTags:            w.SkillTags,
		CategoryTags:         w.CategoryTags,
		WorkExperience:       w.WorkExperience,
		EducationHistory:     w.EducationHistory,
		Portfolio:            w.Portfolio,
		CreditScore:          w.CreditScore,
		Level:                w.Level,
		Status:               w.Status,
		StatusText:           workerStatusText(w.Status),
		AppliedCount:         w.AppliedCount,
		CompletedCount:       w.CompletedCount,
		TotalWorkHours:       w.TotalWorkHours,
		TotalEarnings:        w.TotalEarnings,
		AvgRating:            w.AvgRating,
		RatingCount:          w.RatingCount,
		PunctualityRate:      w.PunctualityRate,
		CompletionRate:       w.CompletionRate,
		RegionID:             w.RegionID,
		CreatedAt:            w.CreatedAt,
		UpdatedAt:            w.UpdatedAt,
	}
}

// ===== C 端 =====

// Create 创建求职者档案（每用户仅一份）
func (s *workerService) Create(regionID uint, userID uint, req *dto.CreateWorkerRequest) (*dto.WorkerInfo, error) {
	// 唯一性校验
	if existing, err := s.repo.FindByUserID(userID); err == nil && existing != nil {
		return nil, ErrWorkerDuplicate
	} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	w := &model.LinggongWorker{
		UserID:              userID,
		RealName:            req.RealName,
		Gender:              req.Gender,
		Birthday:            req.Birthday,
		IDCard:              req.IDCard,
		IDCardFrontURL:      req.IDCardFrontURL,
		IDCardBackURL:       req.IDCardBackURL,
		IDCardHandURL:       req.IDCardHandURL,
		Nickname:            req.Nickname,
		Avatar:              req.Avatar,
		Phone:               req.Phone,
		Email:               req.Email,
		Wechat:              req.Wechat,
		Province:            req.Province,
		City:                req.City,
		District:            req.District,
		Address:             req.Address,
		Latitude:            req.Latitude,
		Longitude:           req.Longitude,
		Education:           req.Education,
		School:              req.School,
		Major:               req.Major,
		GraduationYear:      req.GraduationYear,
		JobIntention:        req.JobIntention,
		ExpectedSalary:      req.ExpectedSalary,
		AvailableTime:       req.AvailableTime,
		AvailableNow:        req.AvailableNow,
		HealthCertURL:       req.HealthCertURL,
		HealthCertValidUntil: req.HealthCertValidUntil,
		HasCriminalRecord:   req.HasCriminalRecord,
		CriminalRecordURL:   req.CriminalRecordURL,
		BankAccount:         req.BankAccount,
		BankName:            req.BankName,
		AlipayAccount:       req.AlipayAccount,
		WechatPayAccount:    req.WechatPayAccount,
		Bio:                 req.Bio,
	}
	w.RegionID = regionID

	// 默认值
	if w.Gender == "" {
		w.Gender = "unknown"
	}
	if w.JobIntention == "" {
		w.JobIntention = model.WorkerIntentAny
	}
	w.Age = workerAgeFromBirthday(req.Birthday)
	w.CreditScore = 100
	w.Level = 1
	w.Status = model.WorkerStatusActive

	// JSONB 字段
	if req.SkillTags != nil {
		if jb, err := model.FromJSON(req.SkillTags); err == nil {
			w.SkillTags = jb
		}
	}
	if req.CategoryTags != nil {
		if jb, err := model.FromJSON(req.CategoryTags); err == nil {
			w.CategoryTags = jb
		}
	}
	if req.WorkExperience != nil {
		if jb, err := model.FromJSON(req.WorkExperience); err == nil {
			w.WorkExperience = jb
		}
	}
	if req.EducationHistory != nil {
		if jb, err := model.FromJSON(req.EducationHistory); err == nil {
			w.EducationHistory = jb
		}
	}
	if req.Portfolio != nil {
		if jb, err := model.FromJSON(req.Portfolio); err == nil {
			w.Portfolio = jb
		}
	}

	if err := s.repo.Create(w); err != nil {
		return nil, err
	}
	return toWorkerInfo(w), nil
}

// Update 更新求职者档案（仅本人）
func (s *workerService) Update(id uint, operatorID uint, req *dto.UpdateWorkerRequest) error {
	w, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrWorkerNotFound
		}
		return err
	}
	if w.UserID != operatorID {
		return ErrWorkerNoPermission
	}
	if w.Status == model.WorkerStatusBanned {
		return ErrWorkerStatusInvalid
	}

	fields := map[string]interface{}{}
	if req.RealName != nil {
		fields["real_name"] = *req.RealName
	}
	if req.Gender != nil {
		fields["gender"] = *req.Gender
	}
	if req.Birthday != nil {
		fields["birthday"] = req.Birthday
		fields["age"] = workerAgeFromBirthday(req.Birthday)
	}
	if req.Nickname != nil {
		fields["nickname"] = *req.Nickname
	}
	if req.Avatar != nil {
		fields["avatar"] = *req.Avatar
	}
	if req.Phone != nil {
		fields["phone"] = *req.Phone
	}
	if req.Email != nil {
		fields["email"] = *req.Email
	}
	if req.Wechat != nil {
		fields["wechat"] = *req.Wechat
	}
	if req.Province != nil {
		fields["province"] = *req.Province
	}
	if req.City != nil {
		fields["city"] = *req.City
	}
	if req.District != nil {
		fields["district"] = *req.District
	}
	if req.Address != nil {
		fields["address"] = *req.Address
	}
	if req.Latitude != nil {
		fields["latitude"] = *req.Latitude
	}
	if req.Longitude != nil {
		fields["longitude"] = *req.Longitude
	}
	if req.Education != nil {
		fields["education"] = *req.Education
	}
	if req.School != nil {
		fields["school"] = *req.School
	}
	if req.Major != nil {
		fields["major"] = *req.Major
	}
	if req.GraduationYear != nil {
		fields["graduation_year"] = *req.GraduationYear
	}
	if req.JobIntention != nil {
		fields["job_intention"] = *req.JobIntention
	}
	if req.ExpectedSalary != nil {
		fields["expected_salary"] = *req.ExpectedSalary
	}
	if req.AvailableTime != nil {
		fields["available_time"] = *req.AvailableTime
	}
	if req.AvailableNow != nil {
		fields["available_now"] = *req.AvailableNow
	}
	if req.HealthCertURL != nil {
		fields["health_cert_url"] = *req.HealthCertURL
	}
	if req.HealthCertValidUntil != nil {
		fields["health_cert_valid_until"] = req.HealthCertValidUntil
	}
	if req.BankAccount != nil {
		fields["bank_account"] = *req.BankAccount
	}
	if req.BankName != nil {
		fields["bank_name"] = *req.BankName
	}
	if req.AlipayAccount != nil {
		fields["alipay_account"] = *req.AlipayAccount
	}
	if req.WechatPayAccount != nil {
		fields["wechat_pay_account"] = *req.WechatPayAccount
	}
	if req.Bio != nil {
		fields["bio"] = *req.Bio
	}
	if req.SkillTags != nil {
		if jb, err := model.FromJSON(req.SkillTags); err == nil {
			fields["skill_tags"] = jb
		}
	}
	if req.CategoryTags != nil {
		if jb, err := model.FromJSON(req.CategoryTags); err == nil {
			fields["category_tags"] = jb
		}
	}
	if req.WorkExperience != nil {
		if jb, err := model.FromJSON(req.WorkExperience); err == nil {
			fields["work_experience"] = jb
		}
	}
	if req.EducationHistory != nil {
		if jb, err := model.FromJSON(req.EducationHistory); err == nil {
			fields["education_history"] = jb
		}
	}
	if req.Portfolio != nil {
		if jb, err := model.FromJSON(req.Portfolio); err == nil {
			fields["portfolio"] = jb
		}
	}

	if len(fields) == 0 {
		return nil
	}
	return s.repo.Update(id, fields)
}

// Delete 删除求职者档案（仅本人）
func (s *workerService) Delete(id uint, operatorID uint) error {
	w, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrWorkerNotFound
		}
		return err
	}
	if w.UserID != operatorID {
		return ErrWorkerNoPermission
	}
	return s.repo.Delete(id)
}

// GetByID 获取求职者档案详情
func (s *workerService) GetByID(id uint) (*dto.WorkerInfo, error) {
	w, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrWorkerNotFound
		}
		return nil, err
	}
	return toWorkerInfo(w), nil
}

// GetByUserID 按 user_id 查询（C 端"我的档案"）
func (s *workerService) GetByUserID(userID uint) (*dto.WorkerInfo, error) {
	w, err := s.repo.FindByUserID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrWorkerNotFound
		}
		return nil, err
	}
	return toWorkerInfo(w), nil
}

// List C 端求职者列表（默认仅展示活跃）
func (s *workerService) List(regionID uint, req *dto.WorkerListRequest) (*utils.Pagination, []dto.WorkerInfo, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	opts := repository.WorkerListOptions{
		JobIntention:   req.JobIntention,
		Education:      req.Education,
		City:           req.City,
		AvailableNow:   req.AvailableNow,
		SkillID:        req.SkillID,
		MinCreditScore: req.MinCreditScore,
		Status:         req.Status,
		Keyword:        req.Keyword,
	}
	// C 端默认仅展示活跃用户
	if opts.Status == nil {
		active := model.WorkerStatusActive
		opts.Status = &active
	}
	list, total, err := s.repo.List(regionID, pagination, opts)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	result := make([]dto.WorkerInfo, 0, len(list))
	for i := range list {
		result = append(result, *toWorkerInfo(&list[i]))
	}
	return pagination, result, nil
}

// ===== M 端管理 =====

// AdminList M 端求职者列表（跨地区）
func (s *workerService) AdminList(req *dto.WorkerAdminListRequest) (*utils.Pagination, []dto.WorkerInfo, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	opts := repository.WorkerAdminListOptions{
		RegionID:     req.RegionID,
		UserID:       req.UserID,
		Status:       req.Status,
		JobIntention: req.JobIntention,
		Keyword:      req.Keyword,
	}
	list, total, err := s.repo.AdminList(pagination, opts)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	result := make([]dto.WorkerInfo, 0, len(list))
	for i := range list {
		result = append(result, *toWorkerInfo(&list[i]))
	}
	return pagination, result, nil
}

// Audit 求职者档案审核（占位实现：求职者档案无审核字段，使用 status 控制）
func (s *workerService) Audit(id uint, status int, rejectReason string) error {
	if _, err := s.repo.FindByID(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrWorkerNotFound
		}
		return err
	}
	_ = rejectReason
	return s.repo.Update(id, map[string]interface{}{"status": status})
}

// UpdateStatus 更新求职者档案状态（活跃/不活跃/封禁）
func (s *workerService) UpdateStatus(id uint, status int) error {
	if _, err := s.repo.FindByID(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrWorkerNotFound
		}
		return err
	}
	return s.repo.Update(id, map[string]interface{}{"status": status})
}

// ===== 内部调用 =====

// UpdateRating 更新评分
func (s *workerService) UpdateRating(id uint, avgRating float64, ratingCount int) error {
	return s.repo.UpdateRating(id, avgRating, ratingCount)
}

// UpdateCreditScore 更新信用分
func (s *workerService) UpdateCreditScore(id uint, score int) error {
	return s.repo.UpdateCreditScore(id, score)
}

// IncrAppliedCount 报名次数 +1
func (s *workerService) IncrAppliedCount(id uint) error {
	return s.repo.IncrAppliedCount(id)
}

// IncrCompletedCount 完成单数 +1
func (s *workerService) IncrCompletedCount(id uint) error {
	return s.repo.IncrCompletedCount(id)
}

// IncrTotalWorkHours 累计工作时长
func (s *workerService) IncrTotalWorkHours(id uint, hours int) error {
	return s.repo.IncrTotalWorkHours(id, hours)
}

// IncrTotalEarnings 累计收入
func (s *workerService) IncrTotalEarnings(id uint, amount float64) error {
	return s.repo.IncrTotalEarnings(id, amount)
}
