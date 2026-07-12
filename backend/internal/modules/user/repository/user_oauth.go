// user_oauth.go 第三方账号绑定关系仓储
//
// 与 UserRepository 分离：避免给已有 UserRepository 接口加方法导致测试桩件全量改写，
// 同时保持单一职责。两个仓储共享同一个 *gorm.DB 连接。

package repository

import (
	"wuchang-tongcheng/internal/modules/user/model"

	"gorm.io/gorm"
)

// UserOAuthRepository 第三方账号绑定仓储接口
type UserOAuthRepository interface {
	// FindByProviderOpenID 按 (provider, open_id) 查绑定，未找到返回 gorm.ErrRecordNotFound
	FindByProviderOpenID(provider, openID string) (*model.UserOAuth, error)
	// Create 创建绑定记录
	Create(binding *model.UserOAuth) error
}

type userOAuthRepository struct {
	db *gorm.DB
}

// NewUserOAuthRepository 创建第三方账号绑定仓储
func NewUserOAuthRepository(db *gorm.DB) UserOAuthRepository {
	return &userOAuthRepository{db: db}
}

// FindByProviderOpenID 按 (provider, open_id) 查绑定
func (r *userOAuthRepository) FindByProviderOpenID(provider, openID string) (*model.UserOAuth, error) {
	var b model.UserOAuth
	if err := r.db.Where("provider = ? AND open_id = ?", provider, openID).First(&b).Error; err != nil {
		return nil, err
	}
	return &b, nil
}

// Create 创建绑定记录
func (r *userOAuthRepository) Create(binding *model.UserOAuth) error {
	return r.db.Create(binding).Error
}
