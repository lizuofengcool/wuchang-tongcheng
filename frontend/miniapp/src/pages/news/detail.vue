<template>
  <view class="container">
    <view v-if="loading" class="text-sm flex-center" style="padding:40px;">加载中…</view>
    <view v-else-if="!news" class="card text-sm flex-center">头条不存在</view>
    <view v-else>
      <!-- 标题 -->
      <view class="title-lg" style="font-size:20px;line-height:1.4;">{{ news.title }}</view>

      <!-- 元信息 -->
      <view class="flex-between mt-2 mb-4 text-sm" style="padding-bottom:8px;border-bottom:1px solid #eee;">
        <text>{{ news.author_name }}</text>
        <view class="flex gap-2">
          <text>👁 {{ news.view_count }}</text>
          <text>❤ {{ news.like_count }}</text>
        </view>
      </view>

      <!-- 封面图 -->
      <image
        v-if="news.cover_image"
        :src="news.cover_image"
        mode="widthFix"
        style="width:100%;border-radius:6px;margin-bottom:12px;"
      />

      <!-- 正文（rich-text 渲染富文本 HTML） -->
      <rich-text :nodes="news.content" style="line-height:1.8;font-size:15px;"></rich-text>

      <!-- 标签 -->
      <view v-if="news.tags" class="flex mt-4" style="flex-wrap:wrap;gap:6px;padding-top:8px;border-top:1px solid #eee;">
        <text
          v-for="(tag, i) in news.tags.split(',').filter(t => t.trim())"
          :key="i"
          class="tag"
        >#{{ tag.trim() }}</text>
      </view>

      <!-- 点赞按钮 -->
      <view class="flex-center mt-4">
        <button
          @tap="onLike"
          :type="liked ? 'warn' : 'default'"
          size="mini"
          style="padding:0 30px;border-radius:20px;"
        >
          {{ liked ? '❤ 已点赞' : '🤍 点赞' }}（{{ likeCount }}）
        </button>
      </view>
    </view>
  </view>
</template>

<script>
import { getNewsDetail, getLikeStatus, toggleLike } from '@/api/news'

export default {
  data() {
    return {
      news: null,
      loading: true,
      liked: false,
      likeCount: 0,
    }
  },
  onLoad(options) {
    this.newsId = Number(options.id)
    if (!this.newsId) {
      uni.showToast({ title: '无效的头条 ID', icon: 'none' })
      return
    }
    this.load()
  },
  methods: {
    async load() {
      this.loading = true
      try {
        const news = await getNewsDetail(this.newsId)
        this.news = news
        this.likeCount = news.like_count
        // 已登录则查询点赞状态
        if (uni.getStorageSync('token')) {
          try {
            const status = await getLikeStatus(this.newsId)
            this.liked = status.liked
            this.likeCount = status.like_count
          } catch (e) {}
        }
      } catch (e) {} finally {
        this.loading = false
      }
    },
    async onLike() {
      if (!uni.getStorageSync('token')) {
        uni.showModal({
          title: '提示',
          content: '点赞需先登录，是否前往登录？',
          success: (res) => {
            if (res.confirm) uni.switchTab({ url: '/pages/user/index' })
          },
        })
        return
      }
      try {
        const res = await toggleLike(this.newsId)
        this.liked = res.liked
        this.likeCount = res.like_count
        uni.showToast({
          title: res.liked ? '点赞成功' : '已取消点赞',
          icon: 'none',
        })
      } catch (e) {}
    },
  },
}
</script>

<style>
.tag {
  padding: 2px 8px;
  font-size: 11px;
  background: #f3f4f6;
  color: #6b7280;
  border-radius: 4px;
}
</style>
