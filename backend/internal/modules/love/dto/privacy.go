// Package dto love 相亲交友数据传输对象 - 隐私设置
package dto

// LovePrivacySettingInfo 隐私设置响应
type LovePrivacySettingInfo struct {
	ID                  uint `json:"id"`
	UserID              uint `json:"user_id"`
	LoveID              uint `json:"love_id"`
	HideOnline          bool `json:"hide_online"`
	HideLocation        bool `json:"hide_location"`
	HideAge             bool `json:"hide_age"`
	HideDistance        bool `json:"hide_distance"`
	HideConstellation   bool `json:"hide_constellation"`
	HideHometown        bool `json:"hide_hometown"`
	HideOccupation      bool `json:"hide_occupation"`
	HideIncome          bool `json:"hide_income"`
	HideLastActive      bool `json:"hide_last_active"`
	HideVisitors        bool `json:"hide_visitors"`
	OnlyVerifiedCanSee   bool `json:"only_verified_can_see"`
	OnlyVerifiedCanMatch bool `json:"only_verified_can_match"`
	OnlyMemberCanChat    bool `json:"only_member_can_chat"`
	BlockStrangers      bool `json:"block_strangers"`
	BlockSameCity       bool `json:"block_same_city"`
	AllowPhoneLookup    bool `json:"allow_phone_lookup"`
	AllowContactImport  bool `json:"allow_contact_import"`
	AllowRecommendation bool `json:"allow_recommendation"`
	AllowStory          bool `json:"allow_story"`
	AllowImpression     bool `json:"allow_impression"`
	DistanceVisibility  int  `json:"distance_visibility"`
	AgeVisibility       int  `json:"age_visibility"`
	Status              int  `json:"status"`
}

// UpdateLovePrivacySettingRequest 更新隐私设置请求
type UpdateLovePrivacySettingRequest struct {
	HideOnline          *bool `json:"hide_online"`
	HideLocation        *bool `json:"hide_location"`
	HideAge             *bool `json:"hide_age"`
	HideDistance        *bool `json:"hide_distance"`
	HideConstellation   *bool `json:"hide_constellation"`
	HideHometown        *bool `json:"hide_hometown"`
	HideOccupation      *bool `json:"hide_occupation"`
	HideIncome          *bool `json:"hide_income"`
	HideLastActive      *bool `json:"hide_last_active"`
	HideVisitors        *bool `json:"hide_visitors"`
	OnlyVerifiedCanSee   *bool `json:"only_verified_can_see"`
	OnlyVerifiedCanMatch *bool `json:"only_verified_can_match"`
	OnlyMemberCanChat    *bool `json:"only_member_can_chat"`
	BlockStrangers      *bool `json:"block_strangers"`
	BlockSameCity       *bool `json:"block_same_city"`
	AllowPhoneLookup    *bool `json:"allow_phone_lookup"`
	AllowContactImport  *bool `json:"allow_contact_import"`
	AllowRecommendation *bool `json:"allow_recommendation"`
	AllowStory          *bool `json:"allow_story"`
	AllowImpression     *bool `json:"allow_impression"`
	DistanceVisibility  *int  `json:"distance_visibility" binding:"omitempty,oneof=0 1 2 3"`
	AgeVisibility       *int  `json:"age_visibility" binding:"omitempty,oneof=0 1 2 3"`
}
