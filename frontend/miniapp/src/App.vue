<script>
export default {
  onLaunch() {
    console.log('五常同城小程序启动')
    // 初始化默认地区（武汉市=2）
    if (!uni.getStorageSync('regionId')) {
      uni.setStorageSync('regionId', 2)
    }
  },
  onShow() {},
  onHide() {},
}
</script>

<style>
/* 全局样式 */
page {
  background: #f5f5f5;
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'PingFang SC', 'Hiragino Sans GB', 'Microsoft YaHei', sans-serif;
  font-size: 14px;
  color: #1f2937;
}

/* 通用容器 */
.container {
  padding: 12px;
}

/* 卡片 */
.card {
  background: #fff;
  border-radius: 8px;
  padding: 12px;
  margin-bottom: 10px;
  box-shadow: 0 1px 2px rgba(0,0,0,0.04);
}

/* 标题 */
.title-lg {
  font-size: 18px;
  font-weight: bold;
}
.title-md {
  font-size: 16px;
  font-weight: bold;
}

/* 品牌色 */
.text-brand { color: #dc2626; }
.bg-brand { background: #dc2626; }

/* 文本省略 */
.ellipsis-2 {
  overflow: hidden;
  text-overflow: ellipsis;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
}

/* flex 工具类 */
.flex { display: flex; }
.flex-between { display: flex; justify-content: space-between; align-items: center; }
.flex-center { display: flex; justify-content: center; align-items: center; }
.gap-2 { gap: 8px; }
.gap-3 { gap: 12px; }

.text-sm { font-size: 12px; color: #6b7280; }
.mt-2 { margin-top: 8px; }
.mt-4 { margin-top: 16px; }
.mb-2 { margin-bottom: 8px; }
.mb-4 { margin-bottom: 16px; }
</style>
