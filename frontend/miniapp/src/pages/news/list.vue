<template>
  <view class="container">
    <!-- 分类筛选 -->
    <scroll-view scroll-x class="cat-bar">
      <view
        class="cat-item"
        :class="{ active: !categoryId }"
        @tap="filterByCategory(null)"
      >全部</view>
      <view
        v-for="c in categories"
        :key="c.id"
        class="cat-item"
        :class="{ active: categoryId === c.id }"
        @tap="filterByCategory(c.id)"
      >{{ c.name }}</view>
    </scroll-view>

    <!-- 列表 -->
    <view v-if="loading && newsList.length === 0" class="text-sm flex-center" style="padding:20px;">加载中…</view>
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
        <view class="flex gap-3">
          <image
            v-if="n.cover_image"
            :src="n.cover_image"
            mode="aspectFill"
            style="width:100px;height:80px;border-radius:4px;flex-shrink:0;"
          />
          <view style="flex:1;">
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
      </view>

      <view class="text-sm flex-center" style="padding:12px;">
        {{ loading ? '加载中…' : (hasMore ? '上拉加载更多' : '没有更多了') }}
      </view>
    </view>
  </view>
</template>

<script>
import { listNews, listCategories } from '@/api/news'

export default {
  data() {
    return {
      newsList: [],
      categories: [],
      categoryId: null,
      page: 1,
      pageSize: 10,
      total: 0,
      loading: false,
      hasMore: true,
    }
  },
  onLoad() {
    const cid = uni.getStorageSync('filterCategoryId')
    if (cid) {
      this.categoryId = cid
      uni.removeStorageSync('filterCategoryId')
    }
    this.loadCategories()
    this.refresh()
  },
  onPullDownRefresh() {
    this.refresh().finally(() => uni.stopPullDownRefresh())
  },
  onReachBottom() {
    this.loadMore()
  },
  methods: {
    async loadCategories() {
      try {
        this.categories = (await listCategories()) || []
      } catch (e) {}
    },
    async refresh() {
      this.page = 1
      this.newsList = []
      this.hasMore = true
      return this.loadMore()
    },
    async loadMore() {
      if (this.loading || !this.hasMore) return
      this.loading = true
      try {
        const res = await listNews({
          page: this.page,
          pageSize: this.pageSize,
          categoryId: this.categoryId || undefined,
        })
        this.newsList.push(...(res.list || []))
        this.total = res.total || 0
        this.hasMore = this.newsList.length < this.total
        if (this.hasMore) this.page++
      } catch (e) {} finally {
        this.loading = false
      }
    },
    filterByCategory(cid) {
      this.categoryId = cid
      this.refresh()
    },
    goDetail(id) {
      uni.navigateTo({ url: `/pages/news/detail?id=${id}` })
    },
  },
}
</script>

<style>
.cat-bar {
  white-space: nowrap;
  padding: 8px 0;
  margin-bottom: 8px;
}
.cat-item {
  display: inline-block;
  padding: 4px 12px;
  margin-right: 6px;
  font-size: 13px;
  border-radius: 14px;
  background: #fff;
  color: #374151;
}
.cat-item.active {
  background: #dc2626;
  color: #fff;
}
</style>
