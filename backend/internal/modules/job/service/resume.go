// Package service 简历业务逻辑层
// 依据 v3.2.1 架构方案第四章：对标 BOSS直聘简历
package service

import (
	"errors"

	"wuchang-tongcheng/internal/modules/job/dto"
	"wuchang-tongcheng/internal/modules/job/model"
	"wuchang-tongcheng/internal/modules/job/repository"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

var (
	ErrResumeNotFound     = errors.New("简历不存在")
	ErrResumeNoPermission = errors.New("无权操作此简历")
	ErrResumeStatus       = errors.New("简历状态不允许此操作")
)

// ResumeService 简历业务接口
type ResumeService interface {
	Create(regionID uint, userID uint, req *dto.ResumeCreateRequest) (*dto.ResumeResponse, error)
	Update(id uint, operatorID uint, req *dto.ResumeUpdateRequest) error
	Delete(id uint, operatorID uint) error
	GetByID(id uint, userID uint) (*dto.ResumeResponse, error)
	List(req *dto.ResumeListQuery) (*utils.Pagination, []dto.ResumeResponse, error)
	ListMine(userID uint, page, pageSize int) (*utils.Pagination, []dto.ResumeResponse, error)
	GetDefault(userID uint) (*dto.ResumeResponse, error)
	SetDefault(userID, resumeID uint) error
	UpdateStatus(id uint, operatorID uint, status int) error
}

type resumeService struct {
	repo repository.ResumeRepository
}

// NewResumeService 创建简历 service 实例
func NewResumeService(repo repository.ResumeRepository) ResumeService {
	return &resumeService{repo: repo}
}

// toResumeResponse model -> dto
func toResumeResponse(r *model.JobResume) *dto.ResumeResponse {
	resp := &dto.ResumeResponse{
		ID:                   r.ID,
		UserID:               r.UserID,
		Name:                 r.Name,
		Gender:               r.Gender,
		BirthDate:            r.BirthDate,
		Phone:                r.Phone,
		Email:                r.Email,
		Avatar:               r.Avatar,
		EducationLevel:       r.EducationLevel,
		School:               r.School,
		Major:                r.Major,
		GraduateDate:         r.GraduateDate,
		WorkYears:            r.WorkYears,
		CurrentCompany:       r.CurrentCompany,
		CurrentPosition:      r.CurrentPosition,
		CurrentSalary:        r.CurrentSalary,
		ExpectSalaryMin:      r.ExpectSalaryMin,
		ExpectSalaryMax:      r.ExpectSalaryMax,
		ExpectCity:           r.ExpectCity,
		ExpectPosition:       r.ExpectPosition,
		ExpectIndustry:       r.ExpectIndustry,
		ExpectJobType:        r.ExpectJobType,
		ExpectEmploymentType: r.ExpectEmploymentType,
		Status:               r.Status,
		Completeness:         r.Completeness,
		IsPublic:             r.IsPublic,
		IsDefault:            r.IsDefault,
		ViewCount:            r.ViewCount,
		DeliverCount:         r.DeliverCount,
		InterviewCount:       r.InterviewCount,
		OfferCount:           r.OfferCount,
		SelfIntroduction:     r.SelfIntroduction,
		Advantage:            r.Advantage,
		Disadvantage:         r.Disadvantage,
		Attachments:          []map[string]interface{}{},
		Educations:           []map[string]interface{}{},
		WorkExperiences:      []map[string]interface{}{},
		Projects:             []map[string]interface{}{},
		Skills:               []map[string]interface{}{},
		Certificates:         []map[string]interface{}{},
		Languages:            []map[string]interface{}{},
		Tags:                 []string{},
		RegionID:             r.RegionID,
		CreatedAt:            r.CreatedAt,
		UpdatedAt:            r.UpdatedAt,
	}
	if r.Attachments != nil {
		var arr []map[string]interface{}
		_ = r.Attachments.Parse(&arr)
		if arr != nil {
			resp.Attachments = arr
		}
	}
	if r.Educations != nil {
		var arr []map[string]interface{}
		_ = r.Educations.Parse(&arr)
		if arr != nil {
			resp.Educations = arr
		}
	}
	if r.WorkExperiences != nil {
		var arr []map[string]interface{}
		_ = r.WorkExperiences.Parse(&arr)
		if arr != nil {
			resp.WorkExperiences = arr
		}
	}
	if r.Projects != nil {
		var arr []map[string]interface{}
		_ = r.Projects.Parse(&arr)
		if arr != nil {
			resp.Projects = arr
		}
	}
	if r.Skills != nil {
		var arr []map[string]interface{}
		_ = r.Skills.Parse(&arr)
		if arr != nil {
			resp.Skills = arr
		}
	}
	if r.Certificates != nil {
		var arr []map[string]interface{}
		_ = r.Certificates.Parse(&arr)
		if arr != nil {
			resp.Certificates = arr
		}
	}
	if r.Languages != nil {
		var arr []map[string]interface{}
		_ = r.Languages.Parse(&arr)
		if arr != nil {
			resp.Languages = arr
		}
	}
	if r.Tags != nil {
		var tags []string
		_ = r.Tags.Parse(&tags)
		if tags != nil {
			resp.Tags = tags
		}
	}
	return resp
}

// calcCompleteness 计算简历完整度（0-100）
func calcCompleteness(req *dto.ResumeCreateRequest) int {
	score := 0
	if req.Name != "" {
		score += 10
	}
	if req.Gender != "" && req.Gender != "unlimited" {
		score += 5
	}
	if req.BirthDate != nil {
		score += 5
	}
	if req.Phone != "" {
		score += 10
	}
	if req.Email != "" {
		score += 5
	}
	if req.Avatar != "" {
		score += 5
	}
	if req.EducationLevel != "" && req.EducationLevel != "unlimited" {
		score += 10
	}
	if req.School != "" {
		score += 5
	}
	if req.Major != "" {
		score += 5
	}
	if req.WorkYears > 0 {
		score += 5
	}
	if req.CurrentCompany != "" {
		score += 5
	}
	if req.CurrentPosition != "" {
		score += 5
	}
	if req.ExpectSalaryMin > 0 || req.ExpectSalaryMax > 0 {
		score += 5
	}
	if req.ExpectCity != "" {
		score += 5
	}
	if req.ExpectPosition != "" {
		score += 5
	}
	if req.SelfIntroduction != "" {
		score += 5
	}
	if len(req.Educations) > 0 {
		score += 5
	}
	if len(req.WorkExperiences) > 0 {
		score += 5
	}
	if len(req.Projects) > 0 {
		score += 5
	}
	if score > 100 {
		score = 100
	}
	return score
}

func (s *resumeService) Create(regionID uint, userID uint, req *dto.ResumeCreateRequest) (*dto.ResumeResponse, error) {
	r := &model.JobResume{
		UserID:               userID,
		Name:                 req.Name,
		Gender:               req.Gender,
		BirthDate:            req.BirthDate,
		Phone:                req.Phone,
		Email:                req.Email,
		Avatar:               req.Avatar,
		EducationLevel:       req.EducationLevel,
		School:               req.School,
		Major:                req.Major,
		GraduateDate:         req.GraduateDate,
		WorkYears:            req.WorkYears,
		CurrentCompany:       req.CurrentCompany,
		CurrentPosition:      req.CurrentPosition,
		CurrentSalary:        req.CurrentSalary,
		ExpectSalaryMin:      req.ExpectSalaryMin,
		ExpectSalaryMax:      req.ExpectSalaryMax,
		ExpectCity:           req.ExpectCity,
		ExpectPosition:       req.ExpectPosition,
		ExpectIndustry:       req.ExpectIndustry,
		ExpectJobType:        req.ExpectJobType,
		ExpectEmploymentType: req.ExpectEmploymentType,
		Status:               req.Status,
		IsPublic:             req.IsPublic,
		IsDefault:            req.IsDefault,
		SelfIntroduction:     req.SelfIntroduction,
		Advantage:            req.Advantage,
		Disadvantage:         req.Disadvantage,
		Completeness:         calcCompleteness(req),
	}
	r.RegionID = regionID

	if r.Gender == "" {
		r.Gender = model.GenderUnlimited
	}
	if r.EducationLevel == "" {
		r.EducationLevel = model.EducationUnlimited
	}
	if r.ExpectJobType == "" {
		r.ExpectJobType = model.RecruitmentTypeFullTime
	}
	if r.ExpectEmploymentType == "" {
		r.ExpectEmploymentType = model.EmploymentTypeRegular
	}
	if r.Status == 0 {
		r.Status = model.ResumeStatusPublished
	}

	// JSONB 字段
	if len(req.Attachments) > 0 {
		if jb, err := model.FromJSON(req.Attachments); err == nil {
			r.Attachments = jb
		}
	}
	if len(req.Educations) > 0 {
		if jb, err := model.FromJSON(req.Educations); err == nil {
			r.Educations = jb
		}
	}
	if len(req.WorkExperiences) > 0 {
		if jb, err := model.FromJSON(req.WorkExperiences); err == nil {
			r.WorkExperiences = jb
		}
	}
	if len(req.Projects) > 0 {
		if jb, err := model.FromJSON(req.Projects); err == nil {
			r.Projects = jb
		}
	}
	if len(req.Skills) > 0 {
		if jb, err := model.FromJSON(req.Skills); err == nil {
			r.Skills = jb
		}
	}
	if len(req.Certificates) > 0 {
		if jb, err := model.FromJSON(req.Certificates); err == nil {
			r.Certificates = jb
		}
	}
	if len(req.Languages) > 0 {
		if jb, err := model.FromJSON(req.Languages); err == nil {
			r.Languages = jb
		}
	}
	if len(req.Tags) > 0 {
		if jb, err := model.FromJSON(req.Tags); err == nil {
			r.Tags = jb
		}
	}

	if err := s.repo.Create(r); err != nil {
		return nil, err
	}
	// 若设为默认，取消其他默认
	if r.IsDefault {
		_ = s.repo.SetDefault(userID, r.ID)
	}
	return toResumeResponse(r), nil
}

func (s *resumeService) Update(id uint, operatorID uint, req *dto.ResumeUpdateRequest) error {
	r, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrResumeNotFound
		}
		return err
	}
	if r.UserID != operatorID {
		return ErrResumeNoPermission
	}

	fields := make(map[string]interface{})
	if req.Name != "" {
		fields["name"] = req.Name
	}
	if req.Gender != "" {
		fields["gender"] = req.Gender
	}
	if req.BirthDate != nil {
		fields["birth_date"] = req.BirthDate
	}
	if req.Phone != "" {
		fields["phone"] = req.Phone
	}
	if req.Email != "" {
		fields["email"] = req.Email
	}
	if req.Avatar != "" {
		fields["avatar"] = req.Avatar
	}
	if req.EducationLevel != "" {
		fields["education_level"] = req.EducationLevel
	}
	if req.School != "" {
		fields["school"] = req.School
	}
	if req.Major != "" {
		fields["major"] = req.Major
	}
	if req.GraduateDate != nil {
		fields["graduate_date"] = req.GraduateDate
	}
	if req.WorkYears > 0 {
		fields["work_years"] = req.WorkYears
	}
	if req.CurrentCompany != "" {
		fields["current_company"] = req.CurrentCompany
	}
	if req.CurrentPosition != "" {
		fields["current_position"] = req.CurrentPosition
	}
	if req.CurrentSalary > 0 {
		fields["current_salary"] = req.CurrentSalary
	}
	if req.ExpectSalaryMin > 0 {
		fields["expect_salary_min"] = req.ExpectSalaryMin
	}
	if req.ExpectSalaryMax > 0 {
		fields["expect_salary_max"] = req.ExpectSalaryMax
	}
	if req.ExpectCity != "" {
		fields["expect_city"] = req.ExpectCity
	}
	if req.ExpectPosition != "" {
		fields["expect_position"] = req.ExpectPosition
	}
	if req.ExpectIndustry != "" {
		fields["expect_industry"] = req.ExpectIndustry
	}
	if req.ExpectJobType != "" {
		fields["expect_job_type"] = req.ExpectJobType
	}
	if req.ExpectEmploymentType != "" {
		fields["expect_employment_type"] = req.ExpectEmploymentType
	}
	if req.Status != nil {
		fields["status"] = *req.Status
	}
	if req.IsPublic != nil {
		fields["is_public"] = *req.IsPublic
	}
	if req.IsDefault != nil {
		fields["is_default"] = *req.IsDefault
	}
	if req.SelfIntroduction != "" {
		fields["self_introduction"] = req.SelfIntroduction
	}
	if req.Advantage != "" {
		fields["advantage"] = req.Advantage
	}
	if req.Disadvantage != "" {
		fields["disadvantage"] = req.Disadvantage
	}

	// JSONB 字段（nil 表示不更新，空数组表示清空）
	if req.Attachments != nil {
		if jb, err := model.FromJSON(req.Attachments); err == nil {
			fields["attachments"] = jb
		}
	}
	if req.Educations != nil {
		if jb, err := model.FromJSON(req.Educations); err == nil {
			fields["educations"] = jb
		}
	}
	if req.WorkExperiences != nil {
		if jb, err := model.FromJSON(req.WorkExperiences); err == nil {
			fields["work_experiences"] = jb
		}
	}
	if req.Projects != nil {
		if jb, err := model.FromJSON(req.Projects); err == nil {
			fields["projects"] = jb
		}
	}
	if req.Skills != nil {
		if jb, err := model.FromJSON(req.Skills); err == nil {
			fields["skills"] = jb
		}
	}
	if req.Certificates != nil {
		if jb, err := model.FromJSON(req.Certificates); err == nil {
			fields["certificates"] = jb
		}
	}
	if req.Languages != nil {
		if jb, err := model.FromJSON(req.Languages); err == nil {
			fields["languages"] = jb
		}
	}
	if req.Tags != nil {
		if jb, err := model.FromJSON(req.Tags); err == nil {
			fields["tags"] = jb
		}
	}

	if err := s.repo.UpdateFields(id, fields); err != nil {
		return err
	}
	// 若设为默认，取消其他默认
	if req.IsDefault != nil && *req.IsDefault {
		_ = s.repo.SetDefault(operatorID, id)
	}
	return nil
}

func (s *resumeService) Delete(id uint, operatorID uint) error {
	r, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrResumeNotFound
		}
		return err
	}
	if r.UserID != operatorID {
		return ErrResumeNoPermission
	}
	return s.repo.Delete(id)
}

func (s *resumeService) GetByID(id uint, userID uint) (*dto.ResumeResponse, error) {
	r, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrResumeNotFound
		}
		return nil, err
	}
	// 隐私保护：非本人查看，仅返回公开简历
	if userID > 0 && r.UserID != userID && !r.IsPublic {
		return nil, ErrResumeNotFound
	}
	// 浏览量 +1（异步容错）
	_ = s.repo.IncrViewCount(id)
	return toResumeResponse(r), nil
}

func (s *resumeService) List(req *dto.ResumeListQuery) (*utils.Pagination, []dto.ResumeResponse, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	list, total, err := s.repo.List(repository.ResumeListQuery{
		UserID:         req.UserID,
		Keyword:        req.Keyword,
		EducationLevel: req.EducationLevel,
		ExpectCity:     req.ExpectCity,
		ExpectPosition: req.ExpectPosition,
		IsPublic:       req.IsPublic,
		Status:         req.Status,
	}, pagination)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	result := make([]dto.ResumeResponse, 0, len(list))
	for i := range list {
		result = append(result, *toResumeResponse(&list[i]))
	}
	return pagination, result, nil
}

func (s *resumeService) ListMine(userID uint, page, pageSize int) (*utils.Pagination, []dto.ResumeResponse, error) {
	pagination := utils.NewPagination(page, pageSize)
	// 用户本人查看自己的所有简历（含草稿），将 Status 设为 nil 不限制
	list, total, err := s.repo.List(repository.ResumeListQuery{
		UserID: userID,
	}, pagination)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	result := make([]dto.ResumeResponse, 0, len(list))
	for i := range list {
		result = append(result, *toResumeResponse(&list[i]))
	}
	return pagination, result, nil
}

func (s *resumeService) GetDefault(userID uint) (*dto.ResumeResponse, error) {
	r, err := s.repo.GetDefault(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrResumeNotFound
		}
		return nil, err
	}
	return toResumeResponse(r), nil
}

func (s *resumeService) SetDefault(userID, resumeID uint) error {
	// 校验所有权
	r, err := s.repo.FindByID(resumeID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrResumeNotFound
		}
		return err
	}
	if r.UserID != userID {
		return ErrResumeNoPermission
	}
	return s.repo.SetDefault(userID, resumeID)
}

func (s *resumeService) UpdateStatus(id uint, operatorID uint, status int) error {
	r, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrResumeNotFound
		}
		return err
	}
	if r.UserID != operatorID {
		return ErrResumeNoPermission
	}
	return s.repo.UpdateFields(id, map[string]interface{}{"status": status})
}
