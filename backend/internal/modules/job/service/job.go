// Package service 同城招聘求职业务逻辑层 - 职位主表
// 依据需求文档 1.10：4 维数据隔离（region_id + user_id）
// 依据需求文档 1.5：内容审核必须做（MVP 简化：发布即通过，M 端可手动审核/下架）
package service

import (
	"errors"
	"time"

	"wuchang-tongcheng/internal/modules/job/dto"
	"wuchang-tongcheng/internal/modules/job/model"
	"wuchang-tongcheng/internal/modules/job/repository"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

var (
	ErrJobNotFound     = errors.New("职位不存在")
	ErrJobNoPermission = errors.New("无权操作此职位")
	ErrJobAudited      = errors.New("已审核的职位不能重复审核")
	ErrJobStatus       = errors.New("职位状态不允许此操作")
)

// JobService 职位业务逻辑接口
type JobService interface {
	// C 端
	Create(regionID uint, userID uint, userName string, userPhone string, userAvatar string, req *dto.CreateJobRequest) (*dto.JobInfo, error)
	Update(id uint, operatorID uint, req *dto.UpdateJobRequest) error
	Delete(id uint, operatorID uint) error
	GetByID(id uint, userID uint) (*dto.JobDetailResponse, error)
	List(regionID uint, req *dto.JobListRequest) (*utils.Pagination, []dto.JobInfo, error)
	ListNearby(regionID uint, req *dto.JobNearbyRequest) (*utils.Pagination, []dto.JobInfo, error)
	Search(regionID uint, req *dto.JobSearchRequest) (*utils.Pagination, []dto.JobInfo, error)
	AdvancedSearch(regionID uint, req *dto.AdvancedSearchRequest) (*utils.Pagination, []dto.JobInfo, error)
	ListMine(userID uint, page, pageSize int) (*utils.Pagination, []dto.JobInfo, error)
	ListSimilar(jobID uint, limit int) ([]dto.SimilarJobResponse, error)
	UpdateStatus(id uint, operatorID uint, status int) error

	// 收藏（职位收藏）
	Fav(userID, jobID uint) (*dto.FavResponse, error)
	Unfav(userID, jobID uint) (*dto.FavResponse, error)
	FavStatus(userID, jobID uint) (*dto.FavResponse, error)
	ListFavs(userID uint, page, pageSize int) (*utils.Pagination, []dto.JobInfo, error)

	// 推广
	Promotion(id uint, operatorID uint, req *dto.JobPromotionRequest) error

	// M 端管理
	AdminList(req *dto.JobAdminListRequest) (*utils.Pagination, []dto.JobInfo, error)
	AdminGetByID(id uint) (*dto.JobDetailResponse, error)
	Audit(id uint, auditStatus int, auditReason string) error
	AdminUpdateStatus(id uint, status int) error
}

type jobService struct {
	repo     repository.JobRepository
	favRepo  repository.InteractionRepository
}

// NewJobService 创建职位 service 实例
func NewJobService(repo repository.JobRepository, favRepo repository.InteractionRepository) JobService {
	return &jobService{repo: repo, favRepo: favRepo}
}

// toJobInfo model -> dto（含图片列表拼装）
func toJobInfo(j *model.Job, images []string) *dto.JobInfo {
	info := &dto.JobInfo{
		ID:                  j.ID,
		Title:               j.Title,
		Content:             j.Content,
		Summary:             j.Summary,
		UserID:              j.UserID,
		UserName:            j.UserName,
		UserPhone:           j.UserPhone,
		UserAvatar:          j.UserAvatar,
		SalaryMin:           j.SalaryMin,
		SalaryMax:           j.SalaryMax,
		SalaryUnit:          j.SalaryUnit,
		SalaryMonthly:       j.SalaryMonthly,
		SalaryNegotiable:    j.SalaryNegotiable,
		SalaryRangeID:       j.SalaryRangeID,
		ShowSalary:          j.ShowSalary,
		Education:           j.Education,
		WorkYearMin:         j.WorkYearMin,
		WorkYearMax:         j.WorkYearMax,
		ExperienceText:      j.ExperienceText,
		WorkAddress:         j.WorkAddress,
		WorkLatitude:        j.WorkLatitude,
		WorkLongitude:       j.WorkLongitude,
		WorkCity:            j.WorkCity,
		WorkDistrict:        j.WorkDistrict,
		WorkBusinessDistrict: j.WorkBusinessDistrict,
		RecruitmentType:     j.RecruitmentType,
		EmploymentType:      j.EmploymentType,
		HiringCount:         j.HiringCount,
		Department:          j.Department,
		PositionTemplateID:  j.PositionTemplateID,
		CategoryID:          j.CategoryID,
		CompanyID:           j.CompanyID,
		RecruiterID:         j.RecruiterID,
		RecruiterName:       j.RecruiterName,
		RecruiterAvatar:     j.RecruiterAvatar,
		RecruiterPosition:   j.RecruiterPosition,
		IsUrgent:            j.IsUrgent,
		UrgentExpire:        j.UrgentExpire,
		IsTop:               j.IsTop,
		TopExpire:           j.TopExpire,
		AgeMin:              j.AgeMin,
		AgeMax:              j.AgeMax,
		GenderRequirement:   j.GenderRequirement,
		Major:               j.Major,
		LanguageRequirement: j.LanguageRequirement,
		CertificateRequirement: j.CertificateRequirement,
		TravelFrequency:     j.TravelFrequency,
		ProbationMonths:     j.ProbationMonths,
		ProbationSalaryRatio: j.ProbationSalaryRatio,
		HasSocialInsurance:  j.HasSocialInsurance,
		HasHousingFund:      j.HasHousingFund,
		WorkSchedule:        j.WorkSchedule,
		OvertimeStatus:      j.OvertimeStatus,
		AllowRemote:         j.AllowRemote,
		ContactName:         j.ContactName,
		ContactPhone:        j.ContactPhone,
		ContactEmail:        j.ContactEmail,
		ContactWechat:       j.ContactWechat,
		ApplicationDeadline: j.ApplicationDeadline,
		NeedBgCheck:         j.NeedBgCheck,
		NeedHealthCheck:     j.NeedHealthCheck,
		ExpiryTime:          j.ExpiryTime,
		ViewCount:           j.ViewCount,
		FavCount:            j.FavCount,
		DeliverCount:        j.DeliverCount,
		InterviewCount:      j.InterviewCount,
		OfferCount:          j.OfferCount,
		MessageCount:        j.MessageCount,
		Status:              j.Status,
		AuditStatus:         j.AuditStatus,
		AuditReason:         j.AuditReason,
		PublishedAt:         j.PublishedAt,
		RegionID:            j.RegionID,
		CreatedAt:           j.CreatedAt,
		UpdatedAt:           j.UpdatedAt,
		VideoURL:            j.VideoURL,
		VideoCover:          j.VideoCover,
		Featured:            j.Featured,
		Picked:              j.Picked,
		Verified:            j.Verified,
		PromotionLevel:      j.PromotionLevel,
		TrafficWeight:       j.TrafficWeight,
		Images:              images,
		Distance:            j.Distance,
	}
	if info.SalaryUnit == "" {
		info.SalaryUnit = model.SalaryUnitMonth
	}
	if info.Education == "" {
		info.Education = model.EducationUnlimited
	}
	if info.GenderRequirement == "" {
		info.GenderRequirement = model.GenderUnlimited
	}
	if info.Images == nil {
		info.Images = []string{}
	}
	// 解析 JSONB 数组字段
	if j.Benefits != nil {
		var arr []uint
		_ = j.Benefits.Parse(&arr)
		if arr != nil {
			info.Benefits = arr
		}
	}
	if j.Skills != nil {
		var arr []uint
		_ = j.Skills.Parse(&arr)
		if arr != nil {
			info.Skills = arr
		}
	}
	if j.Tags != nil {
		var arr []string
		_ = j.Tags.Parse(&arr)
		if arr != nil {
			info.Tags = arr
		}
	}
	if j.WelfareTags != nil {
		var arr []string
		_ = j.WelfareTags.Parse(&arr)
		if arr != nil {
			info.WelfareTags = arr
		}
	}
	if j.Allowances != nil {
		var m map[string]interface{}
		_ = j.Allowances.Parse(&m)
		if m != nil {
			info.Allowances = m
		}
	}
	if j.PromotionChannels != nil {
		var m map[string]interface{}
		_ = j.PromotionChannels.Parse(&m)
		if m != nil {
			info.PromotionChannels = m
		}
	}
	return info
}

// ===== C 端 =====

func (s *jobService) Create(regionID uint, userID uint, userName string, userPhone string, userAvatar string, req *dto.CreateJobRequest) (*dto.JobInfo, error) {
	expireDays := req.ExpireDays
	if expireDays <= 0 {
		expireDays = 30
	}
	expiryTime := time.Now().AddDate(0, 0, expireDays)

	j := &model.Job{
		Title:                req.Title,
		Content:              req.Content,
		Summary:              req.Summary,
		UserID:               userID,
		UserName:             userName,
		UserPhone:            userPhone,
		UserAvatar:           userAvatar,
		SalaryMin:            req.SalaryMin,
		SalaryMax:            req.SalaryMax,
		SalaryUnit:           req.SalaryUnit,
		SalaryNegotiable:     req.SalaryNegotiable,
		SalaryRangeID:        req.SalaryRangeID,
		ShowSalary:           req.ShowSalary,
		Education:            req.Education,
		WorkYearMin:          req.WorkYearMin,
		WorkYearMax:          req.WorkYearMax,
		ExperienceText:       req.ExperienceText,
		WorkAddress:          req.WorkAddress,
		WorkLatitude:         req.WorkLatitude,
		WorkLongitude:        req.WorkLongitude,
		WorkCity:             req.WorkCity,
		WorkDistrict:         req.WorkDistrict,
		WorkBusinessDistrict: req.WorkBusinessDistrict,
		RecruitmentType:      req.RecruitmentType,
		EmploymentType:       req.EmploymentType,
		HiringCount:          req.HiringCount,
		Department:           req.Department,
		PositionTemplateID:   req.PositionTemplateID,
		CategoryID:           req.CategoryID,
		CompanyID:            req.CompanyID,
		RecruiterID:          req.RecruiterID,
		RecruiterName:        req.RecruiterName,
		RecruiterAvatar:      req.RecruiterAvatar,
		RecruiterPosition:    req.RecruiterPosition,
		IsUrgent:             req.IsUrgent,
		AgeMin:               req.AgeMin,
		AgeMax:               req.AgeMax,
		GenderRequirement:    req.GenderRequirement,
		Major:                req.Major,
		LanguageRequirement:  req.LanguageRequirement,
		CertificateRequirement: req.CertificateRequirement,
		TravelFrequency:      req.TravelFrequency,
		ProbationMonths:      req.ProbationMonths,
		ProbationSalaryRatio: req.ProbationSalaryRatio,
		HasSocialInsurance:   req.HasSocialInsurance,
		HasHousingFund:       req.HasHousingFund,
		WorkSchedule:         req.WorkSchedule,
		OvertimeStatus:       req.OvertimeStatus,
		AllowRemote:          req.AllowRemote,
		ContactName:          req.ContactName,
		ContactPhone:         req.ContactPhone,
		ContactEmail:         req.ContactEmail,
		ContactWechat:        req.ContactWechat,
		ApplicationDeadline:  req.ApplicationDeadline,
		NeedBgCheck:          req.NeedBgCheck,
		NeedHealthCheck:      req.NeedHealthCheck,
		ExpiryTime:           &expiryTime,
		Status:               req.Status,
		AuditStatus:          model.AuditApproved,
		VideoURL:             req.VideoURL,
		VideoCover:           req.VideoCover,
	}
	j.RegionID = regionID

	if j.SalaryUnit == "" {
		j.SalaryUnit = model.SalaryUnitMonth
	}
	if j.Education == "" {
		j.Education = model.EducationUnlimited
	}
	if j.RecruitmentType == "" {
		j.RecruitmentType = model.RecruitmentTypeFullTime
	}
	if j.EmploymentType == "" {
		j.EmploymentType = model.EmploymentTypeRegular
	}
	if j.GenderRequirement == "" {
		j.GenderRequirement = model.GenderUnlimited
	}
	if j.TravelFrequency == "" {
		j.TravelFrequency = model.TravelNone
	}
	if j.OvertimeStatus == "" {
		j.OvertimeStatus = model.OvertimeUnknown
	}
	// 计算月薪展示值
	if j.SalaryMonthly == 0 && j.SalaryMin > 0 {
		switch j.SalaryUnit {
		case model.SalaryUnitYear:
			j.SalaryMonthly = j.SalaryMin / 12
		case model.SalaryUnitMonth:
			j.SalaryMonthly = j.SalaryMin
		case model.SalaryUnitDay:
			j.SalaryMonthly = j.SalaryMin * 22
		case model.SalaryUnitHour:
			j.SalaryMonthly = j.SalaryMin * 8 * 22
		}
	}

	if req.Status == model.StatusPublished {
		now := time.Now()
		j.PublishedAt = &now
	}

	// JSONB 字段
	if len(req.Benefits) > 0 {
		if jb, err := model.FromJSON(req.Benefits); err == nil {
			j.Benefits = jb
		}
	}
	if len(req.Skills) > 0 {
		if jb, err := model.FromJSON(req.Skills); err == nil {
			j.Skills = jb
		}
	}
	if len(req.Tags) > 0 {
		if jb, err := model.FromJSON(req.Tags); err == nil {
			j.Tags = jb
		}
	}
	if len(req.WelfareTags) > 0 {
		if jb, err := model.FromJSON(req.WelfareTags); err == nil {
			j.WelfareTags = jb
		}
	}
	if req.Allowances != nil {
		if jb, err := model.FromJSON(req.Allowances); err == nil {
			j.Allowances = jb
		}
	}
	if req.PromotionChannels != nil {
		if jb, err := model.FromJSON(req.PromotionChannels); err == nil {
			j.PromotionChannels = jb
		}
	}

	if err := s.repo.Create(j); err != nil {
		return nil, err
	}

	if len(req.Images) > 0 {
		_ = s.repo.ReplaceImages(j.ID, req.Images)
	}

	return toJobInfo(j, req.Images), nil
}

func (s *jobService) Update(id uint, operatorID uint, req *dto.UpdateJobRequest) error {
	j, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrJobNotFound
		}
		return err
	}
	if j.UserID != operatorID {
		return ErrJobNoPermission
	}

	fields := make(map[string]interface{})
	if req.Title != "" {
		fields["title"] = req.Title
	}
	if req.Content != "" {
		fields["content"] = req.Content
	}
	if req.Summary != "" {
		fields["summary"] = req.Summary
	}
	if req.CategoryID > 0 {
		fields["category_id"] = req.CategoryID
	}
	if req.CompanyID > 0 {
		fields["company_id"] = req.CompanyID
	}
	if req.SalaryMin >= 0 {
		fields["salary_min"] = req.SalaryMin
	}
	if req.SalaryMax >= 0 {
		fields["salary_max"] = req.SalaryMax
	}
	if req.SalaryUnit != "" {
		fields["salary_unit"] = req.SalaryUnit
	}
	if req.SalaryNegotiable != nil {
		fields["salary_negotiable"] = *req.SalaryNegotiable
	}
	if req.SalaryRangeID > 0 {
		fields["salary_range_id"] = req.SalaryRangeID
	}
	if req.ShowSalary != nil {
		fields["show_salary"] = *req.ShowSalary
	}
	if req.Education != "" {
		fields["education"] = req.Education
	}
	if req.WorkYearMin >= 0 {
		fields["work_year_min"] = req.WorkYearMin
	}
	if req.WorkYearMax >= 0 {
		fields["work_year_max"] = req.WorkYearMax
	}
	if req.ExperienceText != "" {
		fields["experience_text"] = req.ExperienceText
	}
	if req.WorkAddress != "" {
		fields["work_address"] = req.WorkAddress
	}
	if req.WorkLatitude != 0 {
		fields["work_latitude"] = req.WorkLatitude
	}
	if req.WorkLongitude != 0 {
		fields["work_longitude"] = req.WorkLongitude
	}
	if req.WorkCity != "" {
		fields["work_city"] = req.WorkCity
	}
	if req.WorkDistrict != "" {
		fields["work_district"] = req.WorkDistrict
	}
	if req.WorkBusinessDistrict != "" {
		fields["work_business_district"] = req.WorkBusinessDistrict
	}
	if req.RecruitmentType != "" {
		fields["recruitment_type"] = req.RecruitmentType
	}
	if req.EmploymentType != "" {
		fields["employment_type"] = req.EmploymentType
	}
	if req.HiringCount > 0 {
		fields["hiring_count"] = req.HiringCount
	}
	if req.Department != "" {
		fields["department"] = req.Department
	}
	if req.PositionTemplateID > 0 {
		fields["position_template_id"] = req.PositionTemplateID
	}
	if req.RecruiterID > 0 {
		fields["recruiter_id"] = req.RecruiterID
	}
	if req.RecruiterName != "" {
		fields["recruiter_name"] = req.RecruiterName
	}
	if req.RecruiterAvatar != "" {
		fields["recruiter_avatar"] = req.RecruiterAvatar
	}
	if req.RecruiterPosition != "" {
		fields["recruiter_position"] = req.RecruiterPosition
	}
	if req.IsUrgent != nil {
		fields["is_urgent"] = *req.IsUrgent
	}
	if req.AgeMin >= 0 {
		fields["age_min"] = req.AgeMin
	}
	if req.AgeMax >= 0 {
		fields["age_max"] = req.AgeMax
	}
	if req.GenderRequirement != "" {
		fields["gender_requirement"] = req.GenderRequirement
	}
	if req.Major != "" {
		fields["major"] = req.Major
	}
	if req.LanguageRequirement != "" {
		fields["language_requirement"] = req.LanguageRequirement
	}
	if req.CertificateRequirement != "" {
		fields["certificate_requirement"] = req.CertificateRequirement
	}
	if req.TravelFrequency != "" {
		fields["travel_frequency"] = req.TravelFrequency
	}
	if req.ProbationMonths >= 0 {
		fields["probation_months"] = req.ProbationMonths
	}
	if req.ProbationSalaryRatio > 0 {
		fields["probation_salary_ratio"] = req.ProbationSalaryRatio
	}
	if req.HasSocialInsurance != nil {
		fields["has_social_insurance"] = *req.HasSocialInsurance
	}
	if req.HasHousingFund != nil {
		fields["has_housing_fund"] = *req.HasHousingFund
	}
	if req.WorkSchedule != "" {
		fields["work_schedule"] = req.WorkSchedule
	}
	if req.OvertimeStatus != "" {
		fields["overtime_status"] = req.OvertimeStatus
	}
	if req.AllowRemote != nil {
		fields["allow_remote"] = *req.AllowRemote
	}
	if req.ContactName != "" {
		fields["contact_name"] = req.ContactName
	}
	if req.ContactPhone != "" {
		fields["contact_phone"] = req.ContactPhone
	}
	if req.ContactEmail != "" {
		fields["contact_email"] = req.ContactEmail
	}
	if req.ContactWechat != "" {
		fields["contact_wechat"] = req.ContactWechat
	}
	if req.ApplicationDeadline != nil {
		fields["application_deadline"] = req.ApplicationDeadline
	}
	if req.NeedBgCheck != nil {
		fields["need_bg_check"] = *req.NeedBgCheck
	}
	if req.NeedHealthCheck != nil {
		fields["need_health_check"] = *req.NeedHealthCheck
	}
	if req.VideoURL != "" {
		fields["video_url"] = req.VideoURL
	}
	if req.VideoCover != "" {
		fields["video_cover"] = req.VideoCover
	}
	if req.Status != 0 {
		fields["status"] = req.Status
		if req.Status == model.StatusPublished && j.PublishedAt == nil {
			now := time.Now()
			fields["published_at"] = &now
		}
	}
	if req.ExpireDays > 0 {
		expiryTime := time.Now().AddDate(0, 0, req.ExpireDays)
		fields["expiry_time"] = &expiryTime
	}

	// JSONB 字段
	if req.Benefits != nil {
		if jb, err := model.FromJSON(req.Benefits); err == nil {
			fields["benefits"] = jb
		}
	}
	if req.Skills != nil {
		if jb, err := model.FromJSON(req.Skills); err == nil {
			fields["skills"] = jb
		}
	}
	if req.Tags != nil {
		if jb, err := model.FromJSON(req.Tags); err == nil {
			fields["tags"] = jb
		}
	}
	if req.WelfareTags != nil {
		if jb, err := model.FromJSON(req.WelfareTags); err == nil {
			fields["welfare_tags"] = jb
		}
	}
	if req.Allowances != nil {
		if jb, err := model.FromJSON(req.Allowances); err == nil {
			fields["allowances"] = jb
		}
	}
	if req.PromotionChannels != nil {
		if jb, err := model.FromJSON(req.PromotionChannels); err == nil {
			fields["promotion_channels"] = jb
		}
	}

	if err := s.repo.UpdateFields(id, fields); err != nil {
		return err
	}

	if req.Images != nil {
		_ = s.repo.ReplaceImages(id, req.Images)
	}
	return nil
}

func (s *jobService) Delete(id uint, operatorID uint) error {
	j, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrJobNotFound
		}
		return err
	}
	if j.UserID != operatorID {
		return ErrJobNoPermission
	}
	return s.repo.Delete(id)
}

func (s *jobService) GetByID(id uint, userID uint) (*dto.JobDetailResponse, error) {
	j, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrJobNotFound
		}
		return nil, err
	}
	// 浏览量 +1（异步容错）
	_ = s.repo.IncrViewCount(id)

	images, _ := s.repo.ListImages(id)
	urls := make([]string, 0, len(images))
	for _, img := range images {
		urls = append(urls, img.URL)
	}
	info := toJobInfo(j, urls)

	// 收藏状态
	if userID > 0 {
		if exists, _ := s.favRepo.FavExists(userID, model.FavoriteTypeJob, id); exists {
			info.HasFaved = true
		}
	}

	return &dto.JobDetailResponse{JobInfo: *info}, nil
}

func (s *jobService) List(regionID uint, req *dto.JobListRequest) (*utils.Pagination, []dto.JobInfo, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	list, total, err := s.repo.List(regionID, pagination, repository.JobListOptions{
		CategoryID:      req.CategoryID,
		Keyword:         req.Keyword,
		RecruitmentType: req.RecruitmentType,
		EmploymentType:  req.EmploymentType,
		Education:       req.Education,
		WorkYearMin:     req.WorkYearMin,
		WorkYearMax:     req.WorkYearMax,
		SalaryMin:       req.SalaryMin,
		SalaryMax:       req.SalaryMax,
		WorkCity:        req.WorkCity,
		CompanyID:       req.CompanyID,
		AllowRemote:     req.AllowRemote,
		IsUrgent:        req.IsUrgent,
		Featured:        req.Featured,
		Verified:        req.Verified,
		Sort:            req.Sort,
	})
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	result := make([]dto.JobInfo, 0, len(list))
	for i := range list {
		images, _ := s.repo.ListImages(list[i].ID)
		urls := make([]string, 0, len(images))
		for _, img := range images {
			urls = append(urls, img.URL)
		}
		result = append(result, *toJobInfo(&list[i], urls))
	}
	return pagination, result, nil
}

func (s *jobService) ListNearby(regionID uint, req *dto.JobNearbyRequest) (*utils.Pagination, []dto.JobInfo, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	list, total, err := s.repo.ListNearby(regionID, pagination, req.Latitude, req.Longitude, req.RadiusKm, repository.JobListOptions{})
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	result := make([]dto.JobInfo, 0, len(list))
	for i := range list {
		result = append(result, *toJobInfo(&list[i], nil))
	}
	return pagination, result, nil
}

func (s *jobService) Search(regionID uint, req *dto.JobSearchRequest) (*utils.Pagination, []dto.JobInfo, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	list, total, err := s.repo.Search(regionID, pagination, req.Keyword)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	result := make([]dto.JobInfo, 0, len(list))
	for i := range list {
		result = append(result, *toJobInfo(&list[i], nil))
	}
	return pagination, result, nil
}

func (s *jobService) AdvancedSearch(regionID uint, req *dto.AdvancedSearchRequest) (*utils.Pagination, []dto.JobInfo, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	list, total, err := s.repo.AdvancedSearch(regionID, pagination, repository.JobAdvancedSearchOptions{
		Keyword:         req.Keyword,
		CategoryID:      req.CategoryID,
		RecruitmentType: req.RecruitmentType,
		EmploymentType:  req.EmploymentType,
		Education:       req.Education,
		WorkYearMin:     req.WorkYearMin,
		WorkYearMax:     req.WorkYearMax,
		SalaryMin:       req.SalaryMin,
		SalaryMax:       req.SalaryMax,
		WorkCity:        req.WorkCity,
		CompanyID:       req.CompanyID,
		SkillIDs:        req.SkillIDs,
		BenefitIDs:      req.BenefitIDs,
		AllowRemote:     req.AllowRemote,
		IsUrgent:        req.IsUrgent,
		Featured:        req.Featured,
		Verified:        req.Verified,
		Sort:            req.Sort,
		Latitude:        req.Latitude,
		Longitude:       req.Longitude,
		RadiusKm:        req.RadiusKm,
	})
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	result := make([]dto.JobInfo, 0, len(list))
	for i := range list {
		result = append(result, *toJobInfo(&list[i], nil))
	}
	return pagination, result, nil
}

func (s *jobService) ListMine(userID uint, page, pageSize int) (*utils.Pagination, []dto.JobInfo, error) {
	pagination := utils.NewPagination(page, pageSize)
	list, total, err := s.repo.ListByUser(userID, pagination)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	result := make([]dto.JobInfo, 0, len(list))
	for i := range list {
		result = append(result, *toJobInfo(&list[i], nil))
	}
	return pagination, result, nil
}

func (s *jobService) ListSimilar(jobID uint, limit int) ([]dto.SimilarJobResponse, error) {
	list, err := s.repo.ListSimilar(jobID, limit)
	if err != nil {
		return nil, err
	}
	result := make([]dto.SimilarJobResponse, 0, len(list))
	for i := range list {
		result = append(result, dto.SimilarJobResponse{
			JobID:       list[i].ID,
			Title:       list[i].Title,
			SalaryMin:   list[i].SalaryMin,
			SalaryMax:   list[i].SalaryMax,
			SalaryUnit:  list[i].SalaryUnit,
			WorkCity:    list[i].WorkCity,
			Similarity:  0.8, // 简化：固定值
		})
	}
	return result, nil
}

func (s *jobService) UpdateStatus(id uint, operatorID uint, status int) error {
	j, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrJobNotFound
		}
		return err
	}
	if j.UserID != operatorID {
		return ErrJobNoPermission
	}
	fields := map[string]interface{}{
		"status": status,
	}
	if status == model.StatusPublished && j.PublishedAt == nil {
		now := time.Now()
		fields["published_at"] = &now
	}
	return s.repo.UpdateFields(id, fields)
}

// ===== 收藏 =====

func (s *jobService) Fav(userID, jobID uint) (*dto.FavResponse, error) {
	exists, err := s.favRepo.FavExists(userID, model.FavoriteTypeJob, jobID)
	if err != nil {
		return nil, err
	}
	if !exists {
		fav := &model.JobFavorite{
			UserID:       userID,
			JobID:        jobID,
			FavoriteType: model.FavoriteTypeJob,
			Notify:       true,
		}
		if err := s.favRepo.CreateFav(fav); err != nil {
			return nil, err
		}
		_ = s.repo.IncrFavCount(jobID)
		exists = true
	}
	// 取最新收藏数
	j, _ := s.repo.FindByID(jobID)
	count := 0
	if j != nil {
		count = j.FavCount
	}
	return &dto.FavResponse{HasFaved: exists, FavCount: count}, nil
}

func (s *jobService) Unfav(userID, jobID uint) (*dto.FavResponse, error) {
	exists, _ := s.favRepo.FavExists(userID, model.FavoriteTypeJob, jobID)
	if exists {
		if err := s.favRepo.DeleteFav(userID, model.FavoriteTypeJob, jobID); err != nil {
			return nil, err
		}
		_ = s.repo.DecrFavCount(jobID)
		exists = false
	}
	j, _ := s.repo.FindByID(jobID)
	count := 0
	if j != nil {
		count = j.FavCount
	}
	return &dto.FavResponse{HasFaved: exists, FavCount: count}, nil
}

func (s *jobService) FavStatus(userID, jobID uint) (*dto.FavResponse, error) {
	exists, err := s.favRepo.FavExists(userID, model.FavoriteTypeJob, jobID)
	if err != nil {
		return nil, err
	}
	j, _ := s.repo.FindByID(jobID)
	count := 0
	if j != nil {
		count = j.FavCount
	}
	return &dto.FavResponse{HasFaved: exists, FavCount: count}, nil
}

func (s *jobService) ListFavs(userID uint, page, pageSize int) (*utils.Pagination, []dto.JobInfo, error) {
	pagination := utils.NewPagination(page, pageSize)
	favs, total, err := s.favRepo.ListFavs(userID, model.FavoriteTypeJob, pagination)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	result := make([]dto.JobInfo, 0, len(favs))
	for _, fav := range favs {
		j, err := s.repo.FindByID(fav.JobID)
		if err != nil {
			continue
		}
		result = append(result, *toJobInfo(j, nil))
	}
	return pagination, result, nil
}

// ===== 推广 =====

func (s *jobService) Promotion(id uint, operatorID uint, req *dto.JobPromotionRequest) error {
	j, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrJobNotFound
		}
		return err
	}
	if j.UserID != operatorID {
		return ErrJobNoPermission
	}

	fields := map[string]interface{}{
		"promotion_level": req.PromotionLevel,
		"traffic_weight":  req.TrafficWeight,
		"featured":        req.Featured,
		"picked":          req.Picked,
		"verified":        req.Verified,
		"is_top":          req.IsTop,
		"is_urgent":       req.IsUrgent,
	}
	now := time.Now()
	if req.IsTop && req.TopDays > 0 {
		topExpire := now.AddDate(0, 0, req.TopDays)
		fields["top_expire"] = &topExpire
	}
	if req.IsUrgent && req.UrgentDays > 0 {
		urgentExpire := now.AddDate(0, 0, req.UrgentDays)
		fields["urgent_expire"] = &urgentExpire
	}
	return s.repo.UpdateFields(id, fields)
}

// ===== M 端管理 =====

func (s *jobService) AdminList(req *dto.JobAdminListRequest) (*utils.Pagination, []dto.JobInfo, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	list, total, err := s.repo.AdminList(pagination, repository.JobAdminListOptions{
		RegionID:    req.RegionID,
		UserID:      req.UserID,
		CategoryID:  req.CategoryID,
		CompanyID:   req.CompanyID,
		Status:      req.Status,
		AuditStatus: req.AuditStatus,
		Keyword:     req.Keyword,
	})
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	result := make([]dto.JobInfo, 0, len(list))
	for i := range list {
		result = append(result, *toJobInfo(&list[i], nil))
	}
	return pagination, result, nil
}

func (s *jobService) AdminGetByID(id uint) (*dto.JobDetailResponse, error) {
	j, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrJobNotFound
		}
		return nil, err
	}
	images, _ := s.repo.ListImages(id)
	urls := make([]string, 0, len(images))
	for _, img := range images {
		urls = append(urls, img.URL)
	}
	info := toJobInfo(j, urls)
	return &dto.JobDetailResponse{JobInfo: *info}, nil
}

func (s *jobService) Audit(id uint, auditStatus int, auditReason string) error {
	j, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrJobNotFound
		}
		return err
	}
	if j.AuditStatus == auditStatus {
		return ErrJobAudited
	}
	fields := map[string]interface{}{
		"audit_status": auditStatus,
		"audit_reason": auditReason,
	}
	// 审核通过 → 同步发布状态
	if auditStatus == model.AuditApproved && j.Status == model.StatusDraft {
		fields["status"] = model.StatusPublished
		now := time.Now()
		if j.PublishedAt == nil {
			fields["published_at"] = &now
		}
	}
	// 审核拒绝 → 同步下架
	if auditStatus == model.AuditRejected {
		fields["status"] = model.StatusOffline
	}
	return s.repo.UpdateFields(id, fields)
}

func (s *jobService) AdminUpdateStatus(id uint, status int) error {
	fields := map[string]interface{}{
		"status": status,
	}
	if status == model.StatusPublished {
		now := time.Now()
		fields["published_at"] = &now
	}
	return s.repo.UpdateFields(id, fields)
}
