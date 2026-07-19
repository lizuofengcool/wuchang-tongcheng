// Package model 技能标签 + 福利标签（对标 BOSS直聘）
// Java/Python/PM/UI设计/软技能/语言 + 五险一金/年终奖/弹性工作
package model

import (
	"wuchang-tongcheng/internal/pkg/database"
)

// === 技能分类常量 ===
const (
	SkillCategoryTechnical  = "technical"  // 技术类
	SkillCategoryDesign     = "design"      // 设计类
	SkillCategoryProduct    = "product"     // 产品类
	SkillCategoryOperation  = "operation"   // 运营类
	SkillCategoryMarketing  = "marketing"   // 市场营销
	SkillCategoryFinance    = "finance"     // 财务类
	SkillCategoryHR         = "hr"          // 人力资源
	SkillCategoryLegal      = "legal"       // 法务
	SkillCategorySoftSkill  = "soft_skill"  // 软技能
	SkillCategoryLanguage   = "language"    // 语言能力
	SkillCategoryOther      = "other"       // 其他
)

// === 技能状态常量 ===
const (
	SkillStatusDisabled = 0 // 禁用
	SkillStatusEnabled  = 1 // 启用
)

// JobSkill 技能标签表（对标 BOSS直聘）
type JobSkill struct {
	database.BaseModel // id/created_at/updated_at/deleted_at
	Name        string `gorm:"size:64;not null;uniqueIndex:uniq_job_skills_name_category" json:"name"`        // 技能名称
	Code        string `gorm:"size:64" json:"code"`                                                                  // 技能编码
	Category    string `gorm:"size:32;default:'technical';index;uniqueIndex:uniq_job_skills_name_category" json:"category"` // 技能分类
	Description string `gorm:"size:500" json:"description"`                                                          // 描述
	Icon        string `gorm:"size:64" json:"icon"`                                                                   // 图标
	Color       string `gorm:"size:16;default:'#409EFF'" json:"color"`                                                // 颜色
	Status      int    `gorm:"default:1;index" json:"status"`                                                        // 0禁用 1启用
	Sort        int    `gorm:"default:0;index" json:"sort"`                                                            // 排序
	UseCount    int    `gorm:"default:0" json:"use_count"`                                                            // 使用次数
	IsHot       bool   `gorm:"default:false;index" json:"is_hot"`                                                    // 是否热门
	CreatorID   uint   `gorm:"index" json:"creator_id"`                                                                // 创建人 ID
}

// TableName 表名（job_ 前缀）
func (JobSkill) TableName() string { return "job_skills" }

// === 福利分类常量 ===
const (
	BenefitCategoryWelfare  = "welfare"  // 福利保障
	BenefitCategoryBonus    = "bonus"    // 奖金分红
	BenefitCategorySchedule = "schedule" // 工作时间
	BenefitCategoryLeave    = "leave"     // 假期
	BenefitCategoryTraining = "training" // 培训发展
	BenefitCategoryFacility = "facility" // 公司设施
	BenefitCategoryOther    = "other"    // 其他
)

// === 福利状态常量 ===
const (
	BenefitStatusDisabled = 0 // 禁用
	BenefitStatusEnabled  = 1 // 启用
)

// JobBenefit 福利标签表（对标 BOSS直聘）
type JobBenefit struct {
	database.BaseModel // id/created_at/updated_at/deleted_at
	Name        string `gorm:"size:64;not null;uniqueIndex:uniq_job_benefits_code" json:"name"`              // 福利名称
	Code        string `gorm:"size:64;not null;uniqueIndex:uniq_job_benefits_code" json:"code"`              // 福利编码
	Icon        string `gorm:"size:64" json:"icon"`                                                        // 图标
	Color       string `gorm:"size:16;default:'#67C23A'" json:"color"`                                     // 颜色
	Background  string `gorm:"size:32" json:"background"`                                                  // 背景色
	Category    string `gorm:"size:32;default:'welfare';index" json:"category"`                            // 福利分类
	Description string `gorm:"size:500" json:"description"`                                              // 描述
	Status      int    `gorm:"default:1;index" json:"status"`                                            // 0禁用 1启用
	Sort        int    `gorm:"default:0;index" json:"sort"`                                                // 排序
	UseCount    int    `gorm:"default:0" json:"use_count"`                                                // 使用次数
	IsHot       bool   `gorm:"default:false;index" json:"is_hot"`                                        // 是否热门
	CreatorID   uint   `gorm:"index" json:"creator_id"`                                                    // 创建人 ID
}

// TableName 表名（job_ 前缀）
func (JobBenefit) TableName() string { return "job_benefits" }
