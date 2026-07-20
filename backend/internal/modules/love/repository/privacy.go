// Package repository love 相亲交友数据访问层 - 隐私设置
package repository

import (
	"wuchang-tongcheng/internal/modules/love/model"

	"gorm.io/gorm"
)

// LovePrivacySettingRepository 隐私设置仓储接口
type LovePrivacySettingRepository interface {
	Create(p *model.LovePrivacySetting) error
	FindByID(id uint) (*model.LovePrivacySetting, error)
	FindByUserID(userID uint) (*model.LovePrivacySetting, error)
	FindByLoveID(loveID uint) (*model.LovePrivacySetting, error)
	Update(p *model.LovePrivacySetting) error
	UpdateFields(id uint, fields map[string]interface{}) error
	UpdateByUserID(userID uint, fields map[string]interface{}) error
	Delete(id uint) error
	Upsert(p *model.LovePrivacySetting) error
}

type lovePrivacySettingRepository struct {
	db *gorm.DB
}

// NewLovePrivacySettingRepository 创建隐私设置仓储
func NewLovePrivacySettingRepository(db *gorm.DB) LovePrivacySettingRepository {
	return &lovePrivacySettingRepository{db: db}
}

func (r *lovePrivacySettingRepository) Create(p *model.LovePrivacySetting) error {
	return r.db.Create(p).Error
}

func (r *lovePrivacySettingRepository) FindByID(id uint) (*model.LovePrivacySetting, error) {
	var p model.LovePrivacySetting
	if err := r.db.First(&p, id).Error; err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *lovePrivacySettingRepository) FindByUserID(userID uint) (*model.LovePrivacySetting, error) {
	var p model.LovePrivacySetting
	if err := r.db.Where("user_id = ?", userID).First(&p).Error; err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *lovePrivacySettingRepository) FindByLoveID(loveID uint) (*model.LovePrivacySetting, error) {
	var p model.LovePrivacySetting
	if err := r.db.Where("love_id = ?", loveID).First(&p).Error; err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *lovePrivacySettingRepository) Update(p *model.LovePrivacySetting) error {
	return r.db.Save(p).Error
}

func (r *lovePrivacySettingRepository) UpdateFields(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.LovePrivacySetting{}).Where("id = ?", id).Updates(fields).Error
}

func (r *lovePrivacySettingRepository) UpdateByUserID(userID uint, fields map[string]interface{}) error {
	return r.db.Model(&model.LovePrivacySetting{}).Where("user_id = ?", userID).Updates(fields).Error
}

func (r *lovePrivacySettingRepository) Delete(id uint) error {
	return r.db.Delete(&model.LovePrivacySetting{}, id).Error
}

func (r *lovePrivacySettingRepository) Upsert(p *model.LovePrivacySetting) error {
	result := r.db.Where("user_id = ?", p.UserID).FirstOrCreate(p)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return r.db.Model(&model.LovePrivacySetting{}).Where("user_id = ?", p.UserID).Updates(map[string]interface{}{
			"hide_online":            p.HideOnline,
			"hide_location":          p.HideLocation,
			"hide_age":               p.HideAge,
			"hide_distance":          p.HideDistance,
			"hide_constellation":     p.HideConstellation,
			"hide_hometown":          p.HideHometown,
			"hide_occupation":        p.HideOccupation,
			"hide_income":            p.HideIncome,
			"hide_last_active":       p.HideLastActive,
			"hide_visitors":          p.HideVisitors,
			"only_verified_can_see":   p.OnlyVerifiedCanSee,
			"only_verified_can_match": p.OnlyVerifiedCanMatch,
			"only_member_can_chat":    p.OnlyMemberCanChat,
			"block_strangers":        p.BlockStrangers,
			"block_same_city":        p.BlockSameCity,
			"allow_phone_lookup":     p.AllowPhoneLookup,
			"allow_contact_import":   p.AllowContactImport,
			"allow_recommendation":   p.AllowRecommendation,
			"allow_story":            p.AllowStory,
			"allow_impression":       p.AllowImpression,
			"distance_visibility":    p.DistanceVisibility,
			"age_visibility":         p.AgeVisibility,
			"updated_at":             gorm.Expr("NOW()"),
		}).Error
	}
	return nil
}
