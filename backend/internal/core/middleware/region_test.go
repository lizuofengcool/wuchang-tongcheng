// Package middleware 地区隔离中间件单元测试
// 覆盖 Region() 优先级（Header > Query > 默认值）、GetRegionID 辅助函数。
package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newRegionRouter() *gin.Engine {
	r := gin.New()
	r.Use(Region())
	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"region_id": GetRegionID(c)})
	})
	return r
}

// TestRegion_HeaderFirst Header 优先级最高
func TestRegion_HeaderFirst(t *testing.T) {
	r := newRegionRouter()
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/?region_id=99", nil)
	req.Header.Set("X-Region-ID", "5")
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"region_id":5`)
}

// TestRegion_QueryFallback Header 缺失时用 Query
func TestRegion_QueryFallback(t *testing.T) {
	r := newRegionRouter()
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/?region_id=7", nil)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"region_id":7`)
}

// TestRegion_Default 都缺失时使用默认值 2
func TestRegion_Default(t *testing.T) {
	r := newRegionRouter()
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"region_id":2`)
}

// TestRegion_InvalidHeader 非数字 Header 走默认值
func TestRegion_InvalidHeader(t *testing.T) {
	r := newRegionRouter()
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("X-Region-ID", "abc")
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"region_id":2`)
}

// TestRegion_InvalidQuery 非数字 Query 走默认值
func TestRegion_InvalidQuery(t *testing.T) {
	r := newRegionRouter()
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/?region_id=xyz", nil)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"region_id":2`)
}

// TestRegion_ZeroHeader 0 视为未设置，走默认值
func TestRegion_ZeroHeader(t *testing.T) {
	r := newRegionRouter()
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("X-Region-ID", "0")
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"region_id":2`)
}

// TestGetRegionID_NoContext 无上下文时返回 DefaultRegionID
func TestGetRegionID_NoContext(t *testing.T) {
	r := gin.New()
	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"region_id": GetRegionID(c)})
	})
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"region_id":2`)
}

// TestRegion_Constants 验证常量值
func TestRegion_Constants(t *testing.T) {
	assert.Equal(t, "region_id", RegionIDKey)
	assert.Equal(t, 2, DefaultRegionID)
}
