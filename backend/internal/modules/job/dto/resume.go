// Package dto 简历相关 DTO
// 依据 v3.2.1 架构方案第四章：对标 BOSS直聘简历
package dto

import (
	"time"

	"wuchang-tongcheng/internal/pkg/utils"
)

// ResumeResponse 简历详情响应
type ResumeResponse struct {
	ID                   uint       `json:"id"`
	UserID               uint       `json:"user_id"`
	Name                 string     `json:"name"`
	Gender               string     `json:"gender"`
	BirthDate            *time.Time `json:"birth_date"`
	Phone                string     `json:"phone"`
	Email                string     `json:"email"`
	Avatar               string     `json:"avatar"`
	EducationLevel       string     `json:"education_level"`
	School               string     `json:"school"`
	Major                string     `json:"major"`
	GraduateDate         *time.Time `json:"graduate_date"`
	WorkYears            int        `json:"work_years"`
	CurrentCompany       string     `json:"current_company"`
	CurrentPosition      string     `json:"current_position"`
	CurrentSalary        float64    `json:"current_salary"`
	ExpectSalaryMin      float64    `json:"expect_salary_min"`
	ExpectSalaryMax      float64    `json:"expect_salary_max"`
	ExpectCity           string     `json:"expect_city"`
	ExpectPosition       string     `json:"expect_position"`
	ExpectIndustry       string     `json:"expect_industry"`
	ExpectJobType        string     `json:"expect_job_type"`
	ExpectEmploymentType string     `json:"expect_employment_type"`
	Status               int        `json:"status"`
	Completeness         int        `json:"completeness"`
	IsPublic             bool       `json:"is_public"`
	IsDefault            bool       `json:"is_default"`
	ViewCount            int        `json:"view_count"`
	DeliverCount         int        `json:"deliver_count"`
	InterviewCount       int        `json:"interview_count"`
	OfferCount           int        `json:"offer_count"`
	SelfIntroduction     string     `json:"self_introduction"`
	Advantage            string     `json:"advantage"`
	Disadvantage         string     `json:"disadvantage"`
	Attachments          []map[string]interface{} `json:"attachments"`
	Educations           []map[string]interface{} `json:"educations"`
	WorkExperiences      []map[string]interface{} `json:"work_experiences"`
	Projects             []map[string]interface{} `json:"projects"`
	Skills               []map[string]interface{} `json:"skills"`
	Certificates         []map[string]interface{} `json:"certificates"`
	Languages            []map[string]interface{} `json:"languages"`
	Tags                 []string   `json:"tags"`
	RegionID             uint       `json:"region_id"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`
}

// ResumeCreateRequest 创建简历请求
type ResumeCreateRequest struct {
	Name                 string     `json:"name" binding:"required,max=50"`
	Gender               string     `json:"gender" binding:"omitempty,oneof=unlimited male female"`
	BirthDate            *time.Time `json:"birth_date"`
	Phone                string     `json:"phone" binding:"max=20"`
	Email                string     `json:"email" binding:"max=128"`
	Avatar               string     `json:"avatar" binding:"max=255"`
	EducationLevel       string     `json:"education_level" binding:"omitempty,oneof=unlimited junior_high high_school college bachelor master phd"`
	School               string     `json:"school" binding:"max=128"`
	Major                string     `json:"major" binding:"max=128"`
	GraduateDate         *time.Time `json:"graduate_date"`
	WorkYears            int        `json:"work_years" binding:"gte=0"`
	CurrentCompany       string     `json:"current_company" binding:"max=128"`
	CurrentPosition      string     `json:"current_position" binding:"max=64"`
	CurrentSalary        float64    `json:"current_salary" binding:"gte=0"`
	ExpectSalaryMin      float64    `json:"expect_salary_min" binding:"gte=0"`
	ExpectSalaryMax      float64    `json:"expect_salary_max" binding:"gte=0"`
	ExpectCity           string     `json:"expect_city" binding:"max=64"`
	ExpectPosition       string     `json:"expect_position" binding:"max=128"`
	ExpectIndustry       string     `json:"expect_industry" binding:"max=64"`
	ExpectJobType        string     `json:"expect_job_type" binding:"omitempty,oneof=full_time part_time internship temp outsource gig"`
	ExpectEmploymentType string     `json:"expect_employment_type" binding:"omitempty,oneof=regular labor_dispatch outsourcing freelance"`
	Status               int        `json:"status" binding:"oneof=0 1 2"`
	IsPublic             bool       `json:"is_public"`
	IsDefault            bool       `json:"is_default"`
	SelfIntroduction     string     `json:"self_introduction"`
	Advantage            string     `json:"advantage"`
	Disadvantage         string     `json:"disadvantage"`
	Attachments          []map[string]interface{} `json:"attachments"`
	Educations           []map[string]interface{} `json:"educations"`
	WorkExperiences      []map[string]interface{} `json:"work_experiences"`
	Projects             []map[string]interface{} `json:"projects"`
	Skills               []map[string]interface{} `json:"skills"`
	Certificates         []map[string]interface{} `json:"certificates"`
	Languages            []map[string]interface{} `json:"languages"`
	Tags                 []string   `json:"tags"`
}

// ResumeUpdateRequest 更新简历请求
type ResumeUpdateRequest struct {
	Name                 string     `json:"name" binding:"max=50"`
	Gender               string     `json:"gender" binding:"omitempty,oneof=unlimited male female"`
	BirthDate            *time.Time `json:"birth_date"`
	Phone                string     `json:"phone" binding:"max=20"`
	Email                string     `json:"email" binding:"max=128"`
	Avatar               string     `json:"avatar" binding:"max=255"`
	EducationLevel       string     `json:"education_level" binding:"omitempty,oneof=unlimited junior_high high_school college bachelor master phd"`
	School               string     `json:"school" binding:"max=128"`
	Major                string     `json:"major" binding:"max=128"`
	GraduateDate         *time.Time `json:"graduate_date"`
	WorkYears            int        `json:"work_years"`
	CurrentCompany       string     `json:"current_company" binding:"max=128"`
	CurrentPosition      string     `json:"current_position" binding:"max=64"`
	CurrentSalary        float64    `json:"current_salary"`
	ExpectSalaryMin      float64    `json:"expect_salary_min"`
	ExpectSalaryMax      float64    `json:"expect_salary_max"`
	ExpectCity           string     `json:"expect_city" binding:"max=64"`
	ExpectPosition       string     `json:"expect_position" binding:"max=128"`
	ExpectIndustry       string     `json:"expect_industry" binding:"max=64"`
	ExpectJobType        string     `json:"expect_job_type" binding:"omitempty,oneof=full_time part_time internship temp outsource gig"`
	ExpectEmploymentType string     `json:"expect_employment_type" binding:"omitempty,oneof=regular labor_dispatch outsourcing freelance"`
	Status               *int       `json:"status" binding:"omitempty,oneof=0 1 2 3"`
	IsPublic             *bool      `json:"is_public"`
	IsDefault            *bool      `json:"is_default"`
	SelfIntroduction     string     `json:"self_introduction"`
	Advantage            string     `json:"advantage"`
	Disadvantage         string     `json:"disadvantage"`
	Attachments          []map[string]interface{} `json:"attachments"`
	Educations           []map[string]interface{} `json:"educations"`
	WorkExperiences      []map[string]interface{} `json:"work_experiences"`
	Projects             []map[string]interface{} `json:"projects"`
	Skills               []map[string]interface{} `json:"skills"`
	Certificates         []map[string]interface{} `json:"certificates"`
	Languages            []map[string]interface{} `json:"languages"`
	Tags                 []string   `json:"tags"`
}

// ResumeListQuery 简历列表查询
type ResumeListQuery struct {
	UserID         uint   `form:"user_id" json:"user_id"`
	Keyword        string `form:"keyword" json:"keyword"`
	EducationLevel string `form:"education_level" json:"education_level"`
	ExpectCity     string `form:"expect_city" json:"expect_city"`
	ExpectPosition string `form:"expect_position" json:"expect_position"`
	IsPublic       *bool  `form:"is_public" json:"is_public"`
	Status         *int   `form:"status" json:"status"`
	utils.Pagination
}
