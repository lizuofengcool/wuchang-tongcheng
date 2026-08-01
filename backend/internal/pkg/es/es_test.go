// Package es Elasticsearch 封装单元测试。
//
// 全部用例离线运行（无外部 ES 依赖），聚焦以下可离线验证的行为：
//   - Init 配置缺失时优雅跳过（nil cfg / 空 Addresses 均不初始化、不报错）
//   - 未初始化时 IsAvailable=false、GetClient=nil、Close 幂等返回 nil
//   - 未初始化时所有文档/索引操作统一返回 ErrNotAvailable（业务侧据此降级）
//   - IDToStr 纯函数主键转字符串
//
// 与 pkg/sts、pkg/redis 的离线测试风格保持一致。
package es

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"wuchang-tongcheng/internal/pkg/config"
)

// resetClient 测试辅助：清空全局 client，确保用例间状态隔离。
func resetClient() {
	mu.Lock()
	client = nil
	mu.Unlock()
}

// ----- Init 优雅跳过 -----

// TestInit_NilConfig cfg 为 nil 时跳过初始化且不报错
func TestInit_NilConfig(t *testing.T) {
	resetClient()
	t.Cleanup(resetClient)

	err := Init(nil)
	assert.NoError(t, err)
	assert.False(t, IsAvailable(), "nil cfg 不应初始化客户端")
	assert.Nil(t, GetClient())
}

// TestInit_EmptyAddresses 无 Addresses 时跳过初始化且不报错
func TestInit_EmptyAddresses(t *testing.T) {
	resetClient()
	t.Cleanup(resetClient)

	err := Init(&config.ESConfig{Username: "u", Password: "p"})
	assert.NoError(t, err)
	assert.False(t, IsAvailable(), "空 Addresses 不应初始化客户端")
	assert.Nil(t, GetClient())
}

// ----- 未初始化状态访问 -----

// TestIsAvailable_NotInitialized 未初始化时返回 false
func TestIsAvailable_NotInitialized(t *testing.T) {
	resetClient()
	t.Cleanup(resetClient)

	assert.False(t, IsAvailable())
}

// TestGetClient_NotInitialized 未初始化时返回 nil（而非 panic）
func TestGetClient_NotInitialized(t *testing.T) {
	resetClient()
	t.Cleanup(resetClient)

	assert.Nil(t, GetClient())
}

// TestClose_NotInitialized 未初始化时 Close 幂等返回 nil
func TestClose_NotInitialized(t *testing.T) {
	resetClient()
	t.Cleanup(resetClient)

	assert.NoError(t, Close())
	assert.False(t, IsAvailable())
}

// ----- 未初始化时操作统一降级 ErrNotAvailable -----

// TestOps_NotAvailable 未初始化时所有文档/索引操作返回 ErrNotAvailable
func TestOps_NotAvailable(t *testing.T) {
	resetClient()
	t.Cleanup(resetClient)

	ctx := context.Background()

	t.Run("IndexDoc", func(t *testing.T) {
		err := IndexDoc(ctx, "idx", "1", map[string]string{"k": "v"})
		assert.ErrorIs(t, err, ErrNotAvailable)
	})
	t.Run("DeleteDoc", func(t *testing.T) {
		err := DeleteDoc(ctx, "idx", "1")
		assert.ErrorIs(t, err, ErrNotAvailable)
	})
	t.Run("SearchByQuery", func(t *testing.T) {
		res, err := SearchByQuery(ctx, "idx", `{"query":{"match_all":{}}}`, 0, 10)
		assert.ErrorIs(t, err, ErrNotAvailable)
		assert.Nil(t, res)
	})
	t.Run("CreateIndexIfNotExists", func(t *testing.T) {
		err := CreateIndexIfNotExists(ctx, "idx", "")
		assert.ErrorIs(t, err, ErrNotAvailable)
	})
}

// ----- IDToStr 纯函数 -----

func TestIDToStr(t *testing.T) {
	cases := []struct {
		name string
		id   uint
		want string
	}{
		{"zero", 0, "0"},
		{"one", 1, "1"},
		{"small", 7, "7"},
		{"two_digits", 42, "42"},
		{"three_digits", 255, "255"},
		{"power_of_two", 4096, "4096"},
		{"max_uint32", 4294967295, "4294967295"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			assert.Equal(t, c.want, IDToStr(c.id))
		})
	}
}

// TestErrNotAvailableMessage 错误信息可读
func TestErrNotAvailableMessage(t *testing.T) {
	assert.Equal(t, "elasticsearch not available", ErrNotAvailable.Error())
}
