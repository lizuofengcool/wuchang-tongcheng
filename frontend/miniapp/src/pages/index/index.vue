<template>
  <view class="container">
    <!-- 地区切换 -->
    <view class="card flex-between" @tap="showRegionPicker = true">
      <view>
        <text class="text-sm">当前地区</text>
        <view class="title-md text-brand">{{ currentRegionName }}</view>
      </view>
      <text class="text-sm">切换 ▾</text>
    </view>

    <!-- 分类导航 -->
    <view class="card">
      <view class="title-md mb-2">分类导航</view>
      <view class="flex" style="flex-wrap:wrap;gap:8px;">
        <view
          v-for="c in categories"
          :key="c.id"
          class="cat-tag"
          @tap="goNewsList(c.id)"
        >{{ c.name }}</view>
        <view v-if="categories.length === 0" class="text-sm">暂无分类</view>
      </view>
    </view>

    <!-- 最新头条 -->
    <view class="flex-between mb-2">
      <text class="title-md">最新同城头条</text>
      <text class="text-sm text-brand" @tap="goNewsList()">查看全部 →</text>
    </view>

    <view v-if="loading" class="text-sm flex-center" style="padding:20px;">加载中…</view>
    <view v-else-if="newsList.length === 0" class="card text-sm flex-center">
      暂无头条
    </view>
    <view v-else>
      <view
        v-for="n in newsList"
        :key="n.id"
        class="card"
        @tap="goDetail(n.id)"
      >
        <image
          v-if="n.cover_image"
          :src="n.cover_image"
          mode="aspectFill"
          style="width:100%;height:140px;border-radius:6px;margin-bottom:8px;"
        />
        <view class="title-md ellipsis-2">{{ n.title }}</view>
        <view v-if="n.summary" class="text-sm ellipsis-2 mt-2">{{ n.summary }}</view>
        <view class="flex-between mt-2 text-sm">
          <text>{{ n.author_name }}</text>
          <view class="flex gap-2">
            <text>👁 {{ n.view_count }}</text>
            <text>❤ {{ n.like_count }}</text>
          </view>
        </view>
      </view>
    </view>

    <!-- 地区选择 picker -->
    <uni-popup v-if="showRegionPicker" />
  </view>
</template>

<script>
import { listNews, listCategories, listRegions } from '@/api/news'

export default {
  data() {
    return {
      newsList: [],
      categories: [],
      regions: [],
      currentRegionName: '武汉市',
      loading: true,
      showRegionPicker: false,
    }
  },
  onShow() {
    const rid = uni.getStorageSync('regionId') || 2
    const r = this.regions.find((x) => x.id === rid)
    if (r) this.currentRegionName = r.name
  },
  onLoad() {
    this.loadAll()
  },
  onPullDownRefresh() {
    this.loadAll().finally(() => uni.stopPullDownRefresh())
  },
  methods: {
    async loadAll() {
      this.loading = true
      try {
        const [news, cats, regions] = await Promise.all([
          listNews({ page: 1, pageSize: 6 }),
          listCategories(),
          listRegions(),
        ])
        this.newsList = news.list || []
        this.categories = cats || []
        this.regions = regions || []
        const rid = uni.getStorageSync('regionId') || 2
        const r = this.regions.find((x) => x.id === rid)
        if (r) this.currentRegionName = r.name
      } catch (e) {
        // 错误已在 request 中提示
      } finally {
        this.loading = false
      }
    },
    goDetail(id) {
      uni.navigateTo({ url: `/pages/news/detail?id=${id}` })
    },
    goNewsList(categoryId) {
      const q = categoryId ? `?category_id=${categoryId}` : ''
      uni.switchTab({ url: '/pages/news/list' }).catch(() => {})
      // switchTab 不支持参数，改用 reLaunch
      if (categoryId) {
        uni.setStorageSync('filterCategoryId', categoryId)
      }
    },
  },
}
</script>

<style>
.cat-tag {
  padding: 4px 12px;
  background: #f3f4f6;
  border-radius: 14px;
  font-size: 12px;
  color: #374151;
}
</style>
