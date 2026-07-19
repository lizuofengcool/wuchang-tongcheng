// Package model 职位模板 + 职位分类（对标 BOSS直聘）
// 标准职位库 + 分类层级（互联网/金融/制造/教育/医疗等）
package model

import (
	"wuchang-tongcheng/internal/pkg/database"
)

// === 职位模板状态常量 ===
const (
	PositionTemplateStatusDisabled = 0 // 禁用
	PositionTemplateStatusEnabled  = 1 // 启用
)

// === 分类状态常量 ===
const (
	CategoryStatusDisabled = 0 // 禁用
	CategoryStatusEnabled  = 1 // 启用
)

// JobPosition 职位模板表（对标 BOSS直聘标准职位库）
type JobPosition struct {
	database.BaseModel // id/created_at/updated_at/deleted_at
	Name                  string  `gorm:"size:128;not null;uniqueIndex:uniq_job_positions_name_code" json:"name"`              // 职位名称
	Code                  string  `gorm:"size:64;uniqueIndex:uniq_job_positions_name_code" json:"code"`                          // 职位编码
	CategoryID            uint    `gorm:"index" json:"category_id"`                                                              // 关联分类 ID
	Department            string  `gorm:"size:64" json:"department"`                                                            // 默认部门
	Description           string  `gorm:"type:text" json:"description"`                                                          // 默认描述
	Requirements          string  `gorm:"type:text" json:"requirements"`                                                         // 默认要求
	Responsibilities      string  `gorm:"type:text" json:"responsibilities"`                                                     // 默认职责
	DefaultSalaryMin      float64 `gorm:"type:decimal(12,2);default:0" json:"default_salary_min"`                                // 默认薪资下限
	DefaultSalaryMax      float64 `gorm:"type:decimal(12,2);default:0" json:"default_salary_max"`                                // 默认薪资上限
	DefaultEducation      string  `gorm:"size:32;default:'unlimited'" json:"default_education"`                                 // 默认学历
	DefaultWorkYearMin    int     `gorm:"default:0" json:"default_work_year_min"`                                                // 默认经验下限
	DefaultWorkYearMax    int     `gorm:"default:0" json:"default_work_year_max"`                                                // 默认经验上限
	DefaultRecruitmentType string  `gorm:"size:32;default:'full_time'" json:"default_recruitment_type"`                          // 默认招聘类型
	DefaultBenefits       JSONB   `gorm:"type:jsonb" json:"default_benefits"`                                                    // 默认福利 ID 数组
	DefaultSkills         JSONB   `gorm:"type:jsonb" json:"default_skills"`                                                       // 默认技能 ID 数组
	Status                int     `gorm:"default:1;index" json:"status"`                                                          // 0禁用 1启用
	Sort                  int     `gorm:"default:0;index" json:"sort"`                                                          // 排序
	UseCount              int     `gorm:"default:0" json:"use_count"`                                                            // 使用次数
	CreatorID             uint    `gorm:"index" json:"creator_id"`                                                                // 创建人 ID
}

// TableName 表名（job_ 前缀）
func (JobPosition) TableName() string { return "job_positions" }

// JobCategory 职位分类表（对标 BOSS直聘）
type JobCategory struct {
	database.BaseModel // id/created_at/updated_at/deleted_at
	Name        string `gorm:"size:64;not null" json:"name"`             // 分类名称
	Code        string `gorm:"size:64;not null;uniqueIndex" json:"code"` // 分类编码（唯一）
	ParentID    uint   `gorm:"index" json:"parent_id"`                   // 父分类 ID（0=顶级）
	Level       int    `gorm:"default:1;index" json:"level"`             // 层级 1/2/3
	Icon        string `gorm:"size:64" json:"icon"`                       // 图标
	Color       string `gorm:"size:16;default:'#409EFF'" json:"color"`   // 颜色
	Description string `gorm:"size:500" json:"description"`              // 描述
	Sort        int    `gorm:"default:0;index" json:"sort"`              // 排序
	Status      int    `gorm:"default:1;index" json:"status"`            // 0禁用 1启用
	JobCount    int    `gorm:"default:0" json:"job_count"`               // 关联职位数
}

// TableName 表名（job_ 前缀）
func (JobCategory) TableName() string { return "job_categories" }
