package middleware

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

const (
	// RegionIDKey 地区ID在上下文中的Key
	RegionIDKey = "region_id"
	// DefaultRegionID 默认地区ID（武汉市，由 seed 初始化为 id=2）
	DefaultRegionID = 2
)

// Region 地区数据隔离中间件
// 从请求头或Query参数中获取region_id，存入上下文
// 所有业务表都需要根据region_id进行数据隔离
func Region() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 优先从请求头获取
		regionIDStr := c.GetHeader("X-Region-ID")
		if regionIDStr == "" {
			// 从Query参数获取
			regionIDStr = c.Query("region_id")
		}

		var regionID uint
		if regionIDStr != "" {
			if id, err := strconv.ParseUint(regionIDStr, 10, 32); err == nil {
				regionID = uint(id)
			}
		}

		// 如果没有获取到，使用默认值
		if regionID == 0 {
			regionID = DefaultRegionID
		}

		// 将region_id存入上下文
		c.Set(RegionIDKey, regionID)
		c.Next()
	}
}

// GetRegionID 从上下文中获取地区ID
func GetRegionID(c *gin.Context) uint {
	if value, exists := c.Get(RegionIDKey); exists {
		if id, ok := value.(uint); ok {
			return id
		}
	}
	return DefaultRegionID
}
