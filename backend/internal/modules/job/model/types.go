// Package model job 招聘求职模块通用类型定义
// 提供 JSONB 字段类型与简历专用结构（教育/工作/项目/技能/期望）
// 兼容 GORM 与 PostgreSQL jsonb 类型
package model

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

// JSONB 包装 []byte 以便与 PostgreSQL jsonb 类型交互
// 实现 driver.Valuer 与 sql.Scanner 接口，支持 GORM 自动映射
// 空值落库为 NULL，非空值落库为合法 JSON
type JSONB []byte

// Value 实现 driver.Valuer 接口
func (j JSONB) Value() (driver.Value, error) {
	if j == nil || len(j) == 0 {
		return nil, nil
	}
	return []byte(j), nil
}

// Scan 实现 sql.Scanner 接口
func (j *JSONB) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}
	switch v := value.(type) {
	case []byte:
		*j = append((*j)[:0], v...)
		return nil
	case string:
		*j = []byte(v)
		return nil
	}
	return errors.New("job.JSONB.Scan: unsupported source type")
}

// MarshalJSON 实现 json.Marshaler
func (j JSONB) MarshalJSON() ([]byte, error) {
	if j == nil || len(j) == 0 {
		return []byte("null"), nil
	}
	return j, nil
}

// UnmarshalJSON 实现 json.Unmarshaler
func (j *JSONB) UnmarshalJSON(data []byte) error {
	if j == nil {
		return errors.New("job.JSONB.UnmarshalJSON: nil pointer")
	}
	*j = append((*j)[:0], data...)
	return nil
}

// Bytes 返回底层字节切片的只读副本
func (j JSONB) Bytes() []byte {
	if j == nil {
		return nil
	}
	out := make([]byte, len(j))
	copy(out, j)
	return out
}

// String 返回字符串形式（用于日志/打印）
func (j JSONB) String() string {
	if j == nil || len(j) == 0 {
		return ""
	}
	return string(j)
}

// Parse 尝试反序列化为目标对象
func (j JSONB) Parse(v interface{}) error {
	if j == nil || len(j) == 0 {
		return nil
	}
	return json.Unmarshal(j, v)
}

// FromJSON 从任意 Go 对象构造 JSONB
func FromJSON(v interface{}) (JSONB, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	return JSONB(b), nil
}

// === 简历专用结构（存储为 JSONB） ===

// ResumeEducation 简历教育经历（存储在 job_resumes.educations JSONB 数组中）
type ResumeEducation struct {
	School       string `json:"school"`         // 学校名称
	Major        string `json:"major"`           // 专业
	Degree       string `json:"degree"`          // 学历：junior_high/high_school/college/bachelor/master/phd
	StartDate    string `json:"start_date"`      // 入学时间 YYYY-MM
	EndDate      string `json:"end_date"`        // 毕业/结业时间 YYYY-MM
	IsFullTime   bool   `json:"is_full_time"`   // 是否全日制
	Description  string `json:"description"`    // 描述
}

// WorkExperience 简历工作经历（存储在 job_resumes.work_experiences JSONB 数组中）
type WorkExperience struct {
	Company      string  `json:"company"`       // 公司名称
	Department   string  `json:"department"`    // 部门
	Position     string  `json:"position"`      // 职位
	StartDate    string  `json:"start_date"`    // 入职时间 YYYY-MM
	EndDate      string  `json:"end_date"`      // 离职时间 YYYY-MM（空表示在职）
	IsCurrent    bool    `json:"is_current"`    // 是否在职
	Salary       float64 `json:"salary"`        // 薪资（选填）
	Description  string  `json:"description"`   // 工作描述
	Achievements string  `json:"achievements"`  // 主要业绩
}

// ProjectExperience 简历项目经历（存储在 job_resumes.projects JSONB 数组中）
type ProjectExperience struct {
	Name        string   `json:"name"`         // 项目名称
	Role        string   `json:"role"`         // 担任角色
	StartDate   string   `json:"start_date"`   // 开始时间 YYYY-MM
	EndDate     string   `json:"end_date"`     // 结束时间 YYYY-MM
	Description string   `json:"description"`  // 项目描述
	TechStack   []string `json:"tech_stack"`   // 技术栈
	URL         string   `json:"url"`          // 项目链接（选填）
}

// ResumeSkill 简历技能（存储在 job_resumes.skills JSONB 数组中）
type ResumeSkill struct {
	SkillID  uint   `json:"skill_id"`  // 关联 job_skills ID
	Name     string `json:"name"`      // 技能名称
	Level    string `json:"level"`     // 掌握程度：beginner/intermediate/advanced/expert
	Years    int    `json:"years"`     // 使用年限
}

// ResumeCertificate 简历证书（存储在 job_resumes.certificates JSONB 数组中）
type ResumeCertificate struct {
	Name        string `json:"name"`         // 证书名称
	Issuer      string `json:"issuer"`       // 颁发机构
	IssueDate   string `json:"issue_date"`   // 颁发日期 YYYY-MM
	ExpireDate  string `json:"expire_date"`  // 过期日期 YYYY-MM（空表示长期有效）
	CredentialID string `json:"credential_id"` // 证书编号
	URL         string `json:"url"`          // 证书图片 URL
}

// ResumeLanguage 简历语言能力（存储在 job_resumes.languages JSONB 数组中）
type ResumeLanguage struct {
	Language string `json:"language"` // 语言：english/japanese/korean/french/german/spanish/other
	Level    string `json:"level"`    // 等级：basic/conversational/professional/fluent/native
	CertName string `json:"cert_name"` // 证书名（如 CET-6/IELTS）
	CertScore string `json:"cert_score"` // 证书分数
}

// SalaryRange 薪资范围结构（冗余存储便于序列化）
type SalaryRange struct {
	Min     float64 `json:"min"`      // 最低
	Max     float64 `json:"max"`      // 最高
	Unit    string  `json:"unit"`     // 单位：month/year/hour/day
	Currency string `json:"currency"` // 币种 CNY/USD
}

// PositionSnapshot 职位快照（投递时冗余存储，便于追溯）
type PositionSnapshot struct {
	Title          string   `json:"title"`           // 职位标题
	CompanyName    string   `json:"company_name"`    // 公司名
	SalaryMin      float64 `json:"salary_min"`      // 薪资下限
	SalaryMax      float64 `json:"salary_max"`      // 薪资上限
	SalaryUnit     string   `json:"salary_unit"`     // 薪资单位
	Education      string   `json:"education"`       // 学历要求
	WorkYearMin    int      `json:"work_year_min"`   // 经验下限
	WorkYearMax    int      `json:"work_year_max"`   // 经验上限
	WorkCity       string   `json:"work_city"`       // 工作城市
	RecruitmentType string  `json:"recruitment_type"` // 招聘类型
	Benefits       []uint   `json:"benefits"`        // 福利 ID
	Skills         []uint   `json:"skills"`          // 技能 ID
	Tags           []string `json:"tags"`            // 标签
}

// ResumeSnapshot 简历快照（投递时冗余存储，便于追溯）
type ResumeSnapshot struct {
	Name           string `json:"name"`            // 姓名
	Phone          string `json:"phone"`           // 手机
	Email          string `json:"email"`          // 邮箱
	EducationLevel string `json:"education_level"` // 学历
	School         string `json:"school"`         // 学校
	Major          string `json:"major"`          // 专业
	WorkYears      int    `json:"work_years"`     // 工作年限
	CurrentCompany string `json:"current_company"` // 当前公司
	CurrentPosition string `json:"current_position"` // 当前职位
}

// ApplicationItem 投递附加项（投递时的额外信息，存储在 job_applications.attachments JSONB）
// 用于支持一次投递附带多个附件/简历版本
type ApplicationItem struct {
	Type        string `json:"type"`         // 类型：resume/cover_letter/portfolio/certificate
	Name        string `json:"name"`         // 文件名
	URL         string `json:"url"`          // 文件 URL
	Size        int64  `json:"size"`          // 文件大小（字节）
	ContentType string `json:"content_type"` // MIME 类型
	IsPrimary   bool   `json:"is_primary"`   // 是否主简历
}

// InterviewAttachment 面试附件（存储在 job_interviews.attachments JSONB）
type InterviewAttachment struct {
	Type string `json:"type"` // 类型：online_test/result/feedback/offer
	Name string `json:"name"` // 文件名
	URL  string `json:"url"`  // 文件 URL
}
