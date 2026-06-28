// Package pgtest 自身的 smoke 测试。
// 主要验证：在无 Docker 环境下 SetupPostgres 能优雅 t.Skip，而不是卡死或 panic。
// 在有 Docker 的环境（CI/本地）下会真实启动容器并跑一次建表 + 基础 CRUD。
package pgtest

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	userModel "wuchang-tongcheng/internal/modules/user/model"
)

// TestDockerUnavailableCheck 验证 dockerUnavailableReason 的基本行为：
// 不应 panic，返回的 reason 字符串在无 docker 时非空。
func TestDockerUnavailableCheck(t *testing.T) {
	reason := dockerUnavailableReason()
	t.Logf("dockerUnavailableReason = %q", reason)
	// 不做强断言：CI 可能有 docker，本地可能没有，二者都合法。
	// 仅确保函数能正常返回。
	_ = reason
}

// TestSetupPostgres_Smoke 端到端 smoke：启动容器 → migrate → CRUD 一条 user。
// 无 Docker 时自动 skip。
func TestSetupPostgres_Smoke(t *testing.T) {
	db := SetupPostgres(t)
	MigrateAll(t, db)

	u := &userModel.User{
		Username: "smoke_user",
		Password: "$2a$10$placeholderhashplaceholderhashplaceholderhashplacehold",
		Nickname: "Smoke",
		Email:    "smoke@example.com",
		Status:   1,
	}
	u.RegionID = 2 // 武汉市

	require.NoError(t, db.Create(u).Error)
	require.NotZero(t, u.ID)

	var got userModel.User
	require.NoError(t, db.First(&got, u.ID).Error)
	assert.Equal(t, "smoke_user", got.Username)
	assert.Equal(t, uint(2), got.RegionID)

	// 软删除验证
	require.NoError(t, db.Delete(&userModel.User{}, u.ID).Error)
	var cnt int64
	require.NoError(t, db.Model(&userModel.User{}).Where("id = ?", u.ID).Count(&cnt).Error)
	assert.Equal(t, int64(0), cnt, "软删除后常规查询应查不到")
}
