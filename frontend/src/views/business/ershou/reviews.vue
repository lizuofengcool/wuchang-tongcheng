<template>
  <div class="app-container">
    <!-- 评分分布 -->
    <el-row :gutter="16" class="stat-row">
      <el-col :xs="12" :sm="8" :md="4">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value">{{ ratingStats.total }}</div><div class="stat-label">总评价</div></div></el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="4">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value text-success">{{ ratingStats.avg }}</div><div class="stat-label">平均分</div></div></el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="4">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value text-success">{{ ratingStats.five }}/{{ ratingStats.total || 1 }}</div><div class="stat-label">5星占比</div></div></el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="4">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value">{{ ratingStats.four }}/{{ ratingStats.total || 1 }}</div><div class="stat-label">4星占比</div></div></el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="4">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value text-warning">{{ ratingStats.three }}/{{ ratingStats.total || 1 }}</div><div class="stat-label">3星占比</div></div></el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="4">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value text-danger">{{ ratingStats.one + ratingStats.two }}/{{ ratingStats.total || 1 }}</div><div class="stat-label">差评数</div></div></el-card>
      </el-col>
    </el-row>

    <div class="page-card">
      <el-form :inline="true" :model="filters" class="filter-form" @submit.prevent>
        <el-form-item label="评分">
          <el-select v-model="filters.rating" placeholder="全部" clearable style="width: 120px" @change="onSearch">
            <el-option label="5星" :value="5" />
            <el-option label="4星" :value="4" />
            <el-option label="3星" :value="3" />
            <el-option label="2星" :value="2" />
            <el-option label="1星" :value="1" />
          </el-select>
        </el-form-item>
        <el-form-item label="商品ID">
          <el-input v-model="filters.ershou_id" placeholder="商品ID" clearable style="width: 140px" @keyup.enter="onSearch" />
        </el-form-item>
        <el-form-item label="排序">
          <el-select v-model="filters.sort" placeholder="最新" style="width: 140px" @change="onSearch">
            <el-option label="最新" value="latest" />
            <el-option label="高分优先" value="highest" />
            <el-option label="低分优先" value="lowest" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" :icon="Search" @click="onSearch">搜索</el-button>
          <el-button :icon="RefreshLeft" @click="onReset">重置</el-button>
        </el-form-item>
      </el-form>

      <div class="toolbar"><el-button :icon="Refresh" @click="loadList">刷新</el-button></div>

      <el-table v-loading="loading" :data="list" border stripe>
        <el-table-column prop="id" label="ID" width="70" />
        <el-table-column prop="order_id" label="订单ID" width="90" />
        <el-table-column prop="ershou_id" label="商品ID" width="90" />
        <el-table-column label="评价人" width="140">
          <template #default="{ row }">
            <div>{{ row.is_anonymous ? '匿名用户' : (row.reviewer_name || `#${row.reviewer_id}`) }}</div>
          </template>
        </el-table-column>
        <el-table-column label="评分" width="140">
          <template #default="{ row }">
            <el-rate v-model="row.rating" disabled />
          </template>
        </el-table-column>
        <el-table-column label="内容" min-width="240">
          <template #default="{ row }">
            <div class="review-content">{{ row.content || '-' }}</div>
            <div v-if="row.tags && row.tags.length" class="review-tags">
              <el-tag v-for="(t, i) in row.tags" :key="i" size="small" type="info" class="tag-item">{{ t }}</el-tag>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="图片" width="80">
          <template #default="{ row }">
            <el-badge v-if="row.images && row.images.length" :value="row.images.length" type="primary">
              <el-icon><Picture /></el-icon>
            </el-badge>
            <span v-else class="text-muted">-</span>
          </template>
        </el-table-column>
        <el-table-column label="推荐" width="70">
          <template #default="{ row }">
            <el-tag v-if="row.is_recommended" type="success" size="small">推荐</el-tag>
            <span v-else class="text-muted">-</span>
          </template>
        </el-table-column>
        <el-table-column label="回复" width="80">
          <template #default="{ row }">
            <el-tag v-if="row.reply" type="success" size="small">已回</el-tag>
            <span v-else class="text-muted">未回</span>
          </template>
        </el-table-column>
        <el-table-column label="追评" width="80">
          <template #default="{ row }">
            <el-tag v-if="row.append_content" type="warning" size="small">有</el-tag>
            <span v-else class="text-muted">-</span>
          </template>
        </el-table-column>
        <el-table-column label="评价时间" width="160">
          <template #default="{ row }">{{ formatTime(row.created_at) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="120" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="openReply(row)">回复</el-button>
          </template>
        </el-table-column>
      </el-table>

      <div class="pagination-wrap">
        <el-pagination
          v-model:current-page="page"
          v-model:page-size="pageSize"
          :total="total"
          :page-sizes="[10, 20, 50]"
          layout="total, sizes, prev, pager, next, jumper"
          background
          @current-change="loadList"
          @size-change="loadList"
        />
      </div>
    </div>

    <!-- 回复弹窗 -->
    <el-dialog v-model="replyVisible" title="评价回复" width="600px">
      <div v-if="current" class="reply-original">
        <div class="reply-header">
          <el-rate v-model="current.rating" disabled />
          <span class="reply-time">{{ formatTime(current.created_at) }}</span>
        </div>
        <div class="reply-content">{{ current.content }}</div>
      </div>
      <el-input v-model="replyForm.reply" type="textarea" :rows="4" placeholder="请输入回复内容" maxlength="500" show-word-limit />
      <template #footer>
        <el-button @click="replyVisible = false">取消</el-button>
        <el-button type="primary" :loading="replyLoading" @click="onReply">确认回复</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { Refresh, RefreshLeft, Search, Picture } from '@element-plus/icons-vue'
import { listErshouReviews, replyErshouReview } from '@/api/ershou'
import { formatTime } from '@/utils/format'

const loading = ref(false)
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)
const list = ref([])

const filters = reactive({ rating: null, ershou_id: '', sort: 'latest' })

const ratingStats = reactive({ total: 0, avg: 0, five: 0, four: 0, three: 0, two: 0, one: 0 })

const calcStats = () => {
  const stats = { total: list.value.length, avg: 0, five: 0, four: 0, three: 0, two: 0, one: 0 }
  let sum = 0
  list.value.forEach((r) => {
    sum += r.rating || 0
    if (r.rating === 5) stats.five++
    else if (r.rating === 4) stats.four++
    else if (r.rating === 3) stats.three++
    else if (r.rating === 2) stats.two++
    else if (r.rating === 1) stats.one++
  })
  stats.avg = stats.total ? (sum / stats.total).toFixed(1) : '0.0'
  Object.assign(ratingStats, stats)
}

const onSearch = () => { page.value = 1; loadList() }
const onReset = () => { filters.rating = null; filters.ershou_id = ''; filters.sort = 'latest'; page.value = 1; loadList() }

const loadList = async () => {
  loading.value = true
  try {
    const res = await listErshouReviews({
      page: page.value,
      page_size: pageSize.value,
      rating: filters.rating === null || filters.rating === '' ? undefined : filters.rating,
      ershou_id: filters.ershou_id || undefined,
      sort: filters.sort
    })
    const data = res.data || {}
    list.value = data.list || []
    total.value = data.total || 0
    calcStats()
  } catch (e) {
    list.value = []
    total.value = 0
  } finally {
    loading.value = false
  }
}

const replyVisible = ref(false)
const replyLoading = ref(false)
const current = ref(null)
const replyForm = reactive({ id: null, reply: '' })

const openReply = (row) => {
  current.value = row
  replyForm.id = row.id
  replyForm.reply = row.reply || ''
  replyVisible.value = true
}

const onReply = async () => {
  if (!replyForm.reply.trim()) {
    ElMessage.warning('请输入回复内容')
    return
  }
  replyLoading.value = true
  try {
    await replyErshouReview(replyForm.id, { reply: replyForm.reply })
    ElMessage.success('回复成功')
    replyVisible.value = false
    await loadList()
  } catch (e) {
    // 失败已提示
  } finally {
    replyLoading.value = false
  }
}

onMounted(() => loadList())
</script>

<style scoped>
.stat-row { margin-bottom: 16px; }
.stat-item { text-align: center; padding: 16px; }
.stat-value { font-size: 22px; font-weight: 600; color: #409eff; }
.text-success { color: #67c23a; }
.text-warning { color: #e6a23c; }
.text-danger { color: #f56c6c; }
.filter-form { margin-bottom: 12px; padding: 12px 16px; background: #fafafa; border-radius: 4px; }
.filter-form :deep(.el-form-item) { margin-bottom: 8px; margin-right: 12px; }
.toolbar { display: flex; justify-content: flex-start; margin-bottom: 12px; }
.text-muted { color: #909399; }
.review-content { color: #303133; word-break: break-all; }
.review-tags { margin-top: 4px; }
.tag-item { margin-right: 4px; }
.reply-original {
  padding: 12px; background: #fafafa; border-radius: 4px; margin-bottom: 12px;
}
.reply-header {
  display: flex; justify-content: space-between; align-items: center;
  margin-bottom: 8px;
}
.reply-time { font-size: 12px; color: #909399; }
.reply-content { color: #303133; word-break: break-all; }
.pagination-wrap { margin-top: 16px; display: flex; justify-content: flex-end; }
</style>
