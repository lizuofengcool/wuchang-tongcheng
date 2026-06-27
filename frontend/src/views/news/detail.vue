<template>
  <div class="app-container">
    <div v-loading="loading" class="news-detail page-card">
      <div class="detail-header">
        <el-button :icon="ArrowLeft" @click="goBack">返回列表</el-button>
      </div>

      <template v-if="news">
        <h1 class="news-title">{{ news.title }}</h1>

        <div class="news-meta">
          <el-tag :type="statusTagType(news.status)" size="small">{{ statusText(news.status) }}</el-tag>
          <span class="meta-item">
            <el-icon><User /></el-icon>{{ news.author_name || '佚名' }}
          </span>
          <span class="meta-item">
            <el-icon><Calendar /></el-icon>{{ formatTime(news.published_at || news.created_at) }}
          </span>
          <span class="meta-item">
            <el-icon><View /></el-icon>{{ news.view_count }} 浏览
          </span>
          <span class="meta-item">
            <el-icon><Star /></el-icon>{{ news.like_count }} 点赞
          </span>
          <span v-if="news.tags" class="meta-item">
            <el-icon><Collection /></el-icon>{{ news.tags }}
          </span>
        </div>

        <el-divider />

        <div v-if="news.cover_image" class="news-cover">
          <el-image :src="resolveImg(news.cover_image)" fit="cover" style="max-height: 360px; border-radius: 8px" />
        </div>

        <div class="news-content" v-html="news.content"></div>

        <el-divider />

        <div class="news-actions">
          <el-button
            :type="liked ? 'danger' : 'primary'"
            :icon="liked ? StarFilled : Star"
            :loading="likeLoading"
            size="large"
            @click="handleLike"
          >
            {{ liked ? '已点赞' : '点赞' }}（{{ likeCount }}）
          </el-button>
        </div>
      </template>

      <el-empty v-else-if="!loading" description="头条不存在或已删除" />
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { ArrowLeft, User, Calendar, View, Star, Collection, StarFilled } from '@element-plus/icons-vue'
import { getNews, likeNews, getNewsLikeStatus } from '@/api/news'
import { newsStatusText as statusText, newsStatusTagType as statusTagType, formatTime } from '@/utils/format'

const route = useRoute()
const router = useRouter()

const loading = ref(false)
const likeLoading = ref(false)
const news = ref(null)
const liked = ref(false)
const likeCount = ref(0)

// 后端返回的图片 URL 可能是相对路径（/uploads/...），需要拼接后端地址
const resolveImg = (url) => {
  if (!url) return ''
  if (/^https?:\/\//.test(url)) return url
  const base = import.meta.env.VITE_BACKEND_URL || ''
  return base + url
}

const loadNews = async () => {
  const id = route.params.id
  if (!id) {
    ElMessage.error('缺少头条ID')
    return
  }
  loading.value = true
  try {
    const res = await getNews(id)
    news.value = res.data || null
    likeCount.value = news.value?.like_count || 0
  } catch (e) {
    news.value = null
  } finally {
    loading.value = false
  }
}

const loadLikeStatus = async () => {
  const id = route.params.id
  if (!id) return
  try {
    const res = await getNewsLikeStatus(id)
    liked.value = res.data?.liked || false
    likeCount.value = res.data?.like_count ?? likeCount.value
  } catch (e) {
    // 未登录或接口异常，保持默认
  }
}

const handleLike = async () => {
  const id = route.params.id
  likeLoading.value = true
  try {
    const res = await likeNews(id)
    liked.value = res.data?.liked || false
    likeCount.value = res.data?.like_count ?? likeCount.value
    ElMessage.success(res.data?.liked ? '点赞成功' : '已取消点赞')
  } catch (e) {
    ElMessage.error('操作失败，请重试')
  } finally {
    likeLoading.value = false
  }
}

const goBack = () => {
  router.push({ name: 'News' })
}

onMounted(() => {
  loadNews()
  loadLikeStatus()
})
</script>

<style scoped>
.news-detail {
  max-width: 900px;
  margin: 0 auto;
  padding: 24px;
}

.detail-header {
  margin-bottom: 16px;
}

.news-title {
  font-size: 24px;
  font-weight: 600;
  margin: 0 0 12px;
  line-height: 1.4;
}

.news-meta {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 16px;
  color: #909399;
  font-size: 13px;
}

.meta-item {
  display: inline-flex;
  align-items: center;
  gap: 4px;
}

.news-cover {
  margin: 16px 0;
  text-align: center;
}

.news-content {
  font-size: 15px;
  line-height: 1.8;
  color: #303133;
  word-break: break-word;
}

.news-content :deep(img) {
  max-width: 100%;
  height: auto;
  border-radius: 4px;
}

.news-content :deep(p) {
  margin: 0 0 12px;
}

.news-actions {
  text-align: center;
  padding: 16px 0;
}
</style>
