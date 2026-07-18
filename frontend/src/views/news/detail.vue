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

        <!-- 评论区 -->
        <div class="comment-section">
          <h3 class="comment-title">
            评论 <span class="comment-count">({{ news.comment_count || 0 }})</span>
          </h3>

          <!-- 发表评论 -->
          <div v-if="userStore.isLoggedIn" class="comment-form">
            <el-avatar :size="36" :src="userStore.avatar">
              {{ userStore.nickname.charAt(0) }}
            </el-avatar>
            <div class="comment-form-body">
              <el-input
                v-model="commentContent"
                type="textarea"
                :rows="3"
                placeholder="说点什么吧…（最多 500 字）"
                maxlength="500"
                show-word-limit
                resize="none"
              />
              <div class="comment-form-footer">
                <el-button
                  type="primary"
                  size="small"
                  :loading="commentSubmitting"
                  :disabled="!commentContent.trim()"
                  @click="handleSubmitComment"
                >发表评论</el-button>
              </div>
            </div>
          </div>
          <el-alert
            v-else
            type="info"
            :closable="false"
            show-icon
            class="login-tip"
          >
            <router-link :to="{ name: 'Login', query: { redirect: route.fullPath } }">
              登录
            </router-link>
            后参与评论
          </el-alert>

          <!-- 评论列表 -->
          <div v-loading="commentLoading" class="comment-list">
            <div v-for="c in comments" :key="c.id" class="comment-item">
              <el-avatar :size="36" :src="c.avatar">
                {{ (c.user_name || '?').charAt(0) }}
              </el-avatar>
              <div class="comment-content">
                <div class="comment-meta">
                  <span class="comment-user">{{ c.user_name || '匿名用户' }}</span>
                  <span v-if="c.reply_to" class="comment-reply">
                    回复 <span class="comment-reply-to">@{{ c.reply_to }}</span>
                  </span>
                  <span class="comment-time">{{ formatTime(c.created_at) }}</span>
                </div>
                <div class="comment-text">{{ c.content }}</div>
                <div class="comment-ops">
                  <el-button
                    v-if="canDeleteComment(c)"
                    type="danger"
                    link
                    size="small"
                    @click="handleDeleteComment(c)"
                  >删除</el-button>
                </div>
              </div>
            </div>
            <el-empty
              v-if="!commentLoading && comments.length === 0"
              description="暂无评论，快来抢沙发吧"
              :image-size="80"
            />
            <div v-if="commentTotal > comments.length" class="comment-more">
              <el-button link type="primary" @click="loadMoreComments">加载更多</el-button>
            </div>
          </div>
        </div>
      </template>

      <el-empty v-else-if="!loading" description="头条不存在或已删除" />
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { ArrowLeft, User, Calendar, View, Star, Collection, StarFilled } from '@element-plus/icons-vue'
import { getNews, likeNews, getNewsLikeStatus, listComments, createComment, deleteComment } from '@/api/news'
import { newsStatusText as statusText, newsStatusTagType as statusTagType, formatTime } from '@/utils/format'
import { useUserStore } from '@/stores/user'

const route = useRoute()
const router = useRouter()
const userStore = useUserStore()

const loading = ref(false)
const likeLoading = ref(false)
const news = ref(null)
const liked = ref(false)
const likeCount = ref(0)

// ====== 评论 ======
const comments = ref([])
const commentTotal = ref(0)
const commentPage = ref(1)
const commentPageSize = 10
const commentLoading = ref(false)
const commentContent = ref('')
const commentSubmitting = ref(false)

const loadComments = async (reset = false) => {
  if (reset) {
    commentPage.value = 1
    comments.value = []
  }
  commentLoading.value = true
  try {
    const res = await listComments(route.params.id, {
      page: commentPage.value,
      page_size: commentPageSize
    })
    const data = res.data || {}
    const list = data.list || []
    // 加载更多追加，重置覆盖
    comments.value = reset ? list : [...comments.value, ...list]
    commentTotal.value = data.total || 0
  } catch (e) {
    // 评论加载失败静默
  } finally {
    commentLoading.value = false
  }
}

const loadMoreComments = () => {
  commentPage.value += 1
  loadComments(false)
}

const handleSubmitComment = async () => {
  const content = commentContent.value.trim()
  if (!content) return
  commentSubmitting.value = true
  try {
    const res = await createComment(route.params.id, { content })
    ElMessage.success('评论成功')
    commentContent.value = ''
    // 重新加载第一页评论，保证顺序最新
    await loadComments(true)
    // 同步评论数显示
    if (news.value) {
      news.value.comment_count = (news.value.comment_count || 0) + 1
    }
    // 顶层返回的 comment 可能未在 list 中，忽略 res.data
  } catch (e) {
    // 错误已由 request 拦截器提示
  } finally {
    commentSubmitting.value = false
  }
}

const canDeleteComment = (c) => {
  if (!userStore.isLoggedIn) return false
  // 超管或评论作者可删
  if (userStore.isSuperAdmin) return true
  return userStore.userInfo?.id && c.user_id === userStore.userInfo.id
}

const handleDeleteComment = async (c) => {
  try {
    await ElMessageBox.confirm('确定删除该评论吗？', '提示', { type: 'warning' })
    await deleteComment(c.id)
    ElMessage.success('已删除')
    comments.value = comments.value.filter((x) => x.id !== c.id)
    commentTotal.value = Math.max(0, commentTotal.value - 1)
    if (news.value) {
      news.value.comment_count = Math.max(0, (news.value.comment_count || 0) - 1)
    }
  } catch (e) {
    // 取消或失败
  }
}

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
  loadComments(true)
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

/* 评论区 */
.comment-section {
  margin-top: 8px;
  padding-top: 8px;
}

.comment-title {
  font-size: 18px;
  font-weight: 600;
  margin: 0 0 16px;
  color: #303133;
}

.comment-count {
  font-size: 14px;
  color: #909399;
  font-weight: normal;
}

.comment-form {
  display: flex;
  gap: 12px;
  margin-bottom: 24px;
  align-items: flex-start;
}

.comment-form-body {
  flex: 1;
}

.comment-form-footer {
  margin-top: 8px;
  text-align: right;
}

.login-tip {
  margin-bottom: 24px;
}

.login-tip a {
  color: #409eff;
  text-decoration: none;
  font-weight: 500;
}

.login-tip a:hover {
  text-decoration: underline;
}

.comment-list {
  min-height: 80px;
}

.comment-item {
  display: flex;
  gap: 12px;
  padding: 12px 0;
  border-bottom: 1px solid #f0f0f0;
}

.comment-item:last-child {
  border-bottom: none;
}

.comment-content {
  flex: 1;
  min-width: 0;
}

.comment-meta {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 13px;
  color: #909399;
  margin-bottom: 4px;
}

.comment-user {
  color: #303133;
  font-weight: 500;
}

.comment-reply {
  color: #909399;
}

.comment-reply-to {
  color: #409eff;
}

.comment-time {
  margin-left: auto;
}

.comment-text {
  font-size: 14px;
  color: #303133;
  line-height: 1.6;
  word-break: break-word;
  white-space: pre-wrap;
}

.comment-ops {
  margin-top: 4px;
}

.comment-more {
  text-align: center;
  padding: 12px 0;
}
</style>
