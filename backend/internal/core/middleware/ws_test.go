// Package middleware WebSocket 鉴权中间件单元测试
// 仅覆盖握手前的鉴权失败路径（无 token / 无效 token / hub 不可用），
// 不进入 Upgrade（Upgrade 会启动阻塞的读写泵，需在集成测试中覆盖）。
package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"wuchang-tongcheng/internal/pkg/jwt"
	"wuchang-tongcheng/internal/pkg/ws"
)

func newWSRouter() *gin.Engine {
	r := gin.New()
	r.GET("/ws", WebSocketHandler())
	return r
}

// TestWS_NoToken 缺少 token 返回 401
func TestWS_NoToken(t *testing.T) {
	r := newWSRouter()
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/ws", nil)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "缺少 token")
}

// TestWS_InvalidToken 无效 token 返回 401
func TestWS_InvalidToken(t *testing.T) {
	r := newWSRouter()
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/ws?token=invalid.jwt", nil)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "token 无效或已过期")
}

// TestWS_HubUnavailable hub 未初始化返回 503
// 仅在 hub 未初始化时验证；hub 已初始化（如其他测试调用过 ws.Init）时跳过。
func TestWS_HubUnavailable(t *testing.T) {
	if ws.IsAvailable() {
		t.Skip("Hub 已初始化，跳过 hub 不可用测试")
	}

	token, err := jwt.GenerateToken(1, "alice")
	require.NoError(t, err)

	r := newWSRouter()
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/ws?token="+token, nil)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusServiceUnavailable, w.Code)
	assert.Contains(t, w.Body.String(), "实时服务不可用")
}
