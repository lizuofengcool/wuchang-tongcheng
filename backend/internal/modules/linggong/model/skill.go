// Package model 技能标签表（对标猪八戒威客）
// 技能分类 + 认证 + 评分 + 关联岗位
package model

import (
	"wuchang-tongcheng/internal/pkg/database"
)

// === 技能等级常量 ===
const (
	SkillLevelBeginner     = "beginner"     // 入门
	SkillLevelIntermediate = "intermediate" // 中级
	SkillLevelAdvanced     = "advanced"     // 高级
	SkillLevelExpert       = "expert"       // 专家
	SkillLevelMaster       = "master"       // 大师
)

// === 技能认证状态常量 ===
const (
	SkillCertStatusPending  = 0 // 待审
	SkillCertStatusApproved = 1 // 通过
	SkillCertStatusRejected = 2 // 拒绝
)

// === 技能类型常量 ===
const (
	SkillTypeTechnical   = "technical"   // 技术类
	SkillTypeService     = "service"     // 服务类
	SkillTypeArt         = "art"         // 艺术类
	SkillTypeLanguage    = "language"    // 语言类
	SkillTypeSports      = "sports"      // 体育类
	SkillTypeDriver      = "driver"      // 驾驶类
	SkillTypeLabor      = "labor"       // 劳务类
	SkillTypeDesign      = "design"      // 设计类
	SkillTypeMarketing   = "marketing"   // 营销类
	SkillTypeWriting     = "writing"     // 文案类
	SkillTypeCatering   = "catering"   // 餐饮类
	SkillTypeOther       = "other"       // 其他
)

// LinggongSkill 技能标签表
type LinggongSkill struct {
	database.BaseModel // id/created_at/updated_at/deleted_at
	Name         string  `gorm:"size:64;not null;index;uniqueIndex:uniq_linggong_skills_name" json:"name"` // 技能名
	Code         string  `gorm:"size:64;not null;default:'';index" json:"code"`                       // 编码
	Category     string  `gorm:"size:32;not null;default:'other';index" json:"category"`               // 技能类型
	ParentID     uint    `gorm:"not null;default:0;index" json:"parent_id"`                            // 父技能 ID
	Level        int     `gorm:"not null;default:1" json:"level"`                                     // 层级 1/2/3
	Icon         string  `gorm:"size:255;not null;default:''" json:"icon"`                              // 图标
	Color        string  `gorm:"size:32;not null;default:''" json:"color"`                            // 颜色
	Description  string  `gorm:"size:500;not null;default:''" json:"description"`                     // 描述
	WorkerCount  int     `gorm:"not null;default:0" json:"worker_count"`                              // 拥有此技能的求职者数
	LinggongCount int    `gorm:"not null;default:0" json:"linggong_count"`                            // 关联岗位数
	AvgSalary    float64 `gorm:"type:decimal(12,2);default:0" json:"avg_salary"`                        // 平均薪资
	HotScore     int     `gorm:"not null;default:0;index" json:"hot_score"`                            // 热度
	Status       int     `gorm:"default:1;index" json:"status"`                                       // 0禁用 1启用
	Sort         int     `gorm:"not null;default:0;index" json:"sort"`                                // 排序
}

// TableName 表名（linggong_ 前缀）
func (LinggongSkill) TableName() string { return "linggong_skills" }
