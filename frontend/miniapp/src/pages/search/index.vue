<template>
  <view class="container">
    <!-- 搜索框 -->
    <view class="search-bar">
      <input
        v-model="keyword"
        class="search-input"
        placeholder="搜索头条、商家、服务…"
        confirm-type="search"
        @confirm="doSearch"
      />
      <view class="search-btn" @tap="doSearch">搜索</view>
    </view>

    <!-- 结果 -->
    <view v-if="loading" class="text-sm flex-center" style="padding:20px;">搜索中…</view>
    <view v-else-if="searched">
      <view v-if="results.length === 0" class="card text-sm flex-center">
        未找到与「{{ searchedKeyword }}」相关的头条
      </view>
      <view v-else>
        <view class="text-sm" style="padding:8px 0;">共找到 {{ total }} 条结果</view>
        <view
          v-for="n in results"
          :key="n.id"
          class="card"
          @tap="goDetail(n.id)"
        >
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
        <view v-if="results.length < total" class="text-sm flex-center" style="padding:12px;">
          <view class="search-btn" @tap="loadMore">加载更多</view>
        </view>
      </view>
    </view>
    <view v-else class="card text-sm flex-center">
      请输入关键词搜索头条
    </view>
  </view>
</template>

<script>
import { searchNews } from '@/api/news'

export default {
  data() {
    return {
      keyword: '',
      searchedKeyword: '',
      results: [],
      total: 0,
      page: 1,
      pageSize: 10,
      loading: false,
      searched: false,
    }
  },
  methods: {
    async doSearch() {
      const kw = this.keyword.trim()
      if (!kw) {
        uni.showToast({ title: '请输入关键词', icon: 'none' })
        return
      }
      this.searchedKeyword = kw
      this.page = 1
      this.results = []
      this.searched = true
      await this.load()
    },
    async load() {
      this.loading = true
      try {
        const res = await searchNews({
          keyword: this.searchedKeyword,
          page: this.page,
          pageSize: this.pageSize,
        })
        this.results.push(...(res.list || []))
        this.total = res.total || 0
      } catch (e) {} finally {
        this.loading = false
      }
    },
    loadMore() {
      this.page++
      this.load()
    },
    goDetail(id) {
      uni.navigateTo({ url: `/pages/news/detail?id=${id}` })
    },
  },
}
</script>

<style>
.search-bar {
  display: flex;
  gap: 8px;
  padding: 8px 0;
  margin-bottom: 8px;
}
.search-input {
  flex: 1;
  background: #fff;
  border: 1px solid #e5e7eb;
  border-radius: 6px;
  padding: 8px 12px;
  font-size: 14px;
}
.search-btn {
  background: #dc2626;
  color: #fff;
  padding: 8px 16px;
  border-radius: 6px;
  font-size: 14px;
}
</style>
