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
        <el-form-item label="商户ID">
          <el-input v-model="filters.dh114_id" placeholder="商户ID" clearable style="width: 140px" @keyup.enter="onSearch" />
        </el-form-item>
        <el-form-item label="评分">
          <el-select v-model="filters.rating" placeholder="全部" clearable style="width: 120px" @change="onSearch">
            <el-option label="5星" :value="5" />
            <el-option label="4星" :value="4" />
            <el-option label="3星" :value="3" />
            <el-option label="2星" :value="2" />
            <el-option label="1星" :value="1" />
          </el-select>
        </el-form-item>
        <el-form-item label="审核状态">
          <el-select v-model="filters.audit_status" placeholder="全部" clearable style="width: 120px" @change="onSearch">
            <el-option label="待审核" :value="0" />
            <el-option label="已通过" :value="1" />
            <el-option label="已拒绝" :value="2" />
          </el-select>
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

      <div class="toolbar">
        <el-button :icon="Refresh" @click="loadList">刷新</el-button>
        <el-button type="primary" :icon="Plus" @click="openCreate">新建评价</el-button>
      </div>

      <el-table v-loading="loading" :data="list" border stripe>
        <el-table-column prop="id" label="ID" width="70" />
        <el-table-column prop="dh114_id" label="商户ID" width="90" />
        <el-table-column label="评价人" width="140">
          <template #default="{ row }">
            <div>{{ row.is_anonymous ? '匿名用户' : (row.reviewer_name || `#${row.reviewer_id}`) }}</div>
          </template>
        </el-table-column>
        <el-table-column label="评分" width="140">
          <template #default="{ row }">
            <el-rate :model-value="row.rating" disabled />
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
        <el-table-column label="审核" width="90">
          <template #default="{ row }">
            <el-tag :type="auditTagType(row.audit_status)" size="small">{{ auditText(row.audit_status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="回复" width="80">
          <template #default="{ row }">
            <el-tag v-if="row.reply" type="success" size="small">已回</el-tag>
            <span v-else class="text-muted">未回</span>
          </template>
        </el-table-column>
        <el-table-column label="评价时间" width="160">
          <template #default="{ row }">{{ formatTime(row.created_at) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="openReply(row)">回复</el-button>
            <el-button v-if="row.audit_status !== 1" type="success" link size="small" @click="onAudit(row, 1)">通过</el-button>
            <el-button v-if="row.audit_status !== 2" type="danger" link size="small" @click="onAudit(row, 2)">拒绝</el-button>
            <el-button type="danger" link size="small" @click="onDelete(row)">删除</el-button>
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
    <el-dialog v-model="replyVisible" title="评价回复" width="600px" destroy-on-close>
      <div v-if="current" class="reply-original">
        <div class="reply-header">
          <el-rate :model-value="current.rating" disabled />
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

    <!-- 新建评价弹窗 -->
    <el-dialog v-model="formVisible" title="新建评价" width="600px" destroy-on-close>
      <el-form ref="formRef" :model="form" :rules="rules" label-width="100px">
        <el-form-item label="商户ID" prop="dh114_id">
          <el-input-number v-model="form.dh114_id" :min="1" :controls="false" style="width: 100%" />
        </el-form-item>
        <el-form-item label="评分" prop="rating">
          <el-rate v-model="form.rating" />
        </el-form-item>
        <el-form-item label="内容" prop="content">
          <el-input v-model="form.content" type="textarea" :rows="4" maxlength="500" show-word-limit />
        </el-form-item>
        <el-form-item label="匿名">
          <el-switch v-model="form.is_anonymous" :active-value="1" :inactive-value="0" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="formVisible = false">取消</el-button>
        <el-button type="primary" :loading="formLoading" @click="onSubmit">确认</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Refresh, RefreshLeft, Search, Plus, Picture } from '@element-plus/icons-vue'
import request from '@/utils/request'
import { formatTime } from '@/utils/format'

const loading = ref(false)
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)
const list = ref([])

const filters = reactive({ dh114_id: '', rating: null, audit_status: null, sort: 'latest' })

const ratingStats = reactive({ total: 0, avg: 0, five: 0, four: 0, three: 0, two: 0, one: 0 })

const auditText = (s) => ({ 0: '待审核', 1: '已通过', 2: '已拒绝' }[s] || '-')
const auditTagType = (s) => ({ 0: 'warning', 1: 'success', 2: 'danger' }[s] || 'info')

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
const onReset = () => {
  filters.dh114_id = ''
  filters.rating = null
  filters.audit_status = null
  filters.sort = 'latest'
  page.value = 1
  loadList()
}

const loadList = async () => {
  loading.value = true
  try {
    const params = {
      page: page.value,
      page_size: pageSize.value,
      dh114_id: filters.dh114_id || undefined,
      rating: filters.rating === null || filters.rating === '' ? undefined : filters.rating,
      audit_status: filters.audit_status === null || filters.audit_status === '' ? undefined : filters.audit_status,
      sort: filters.sort
    }
    const res = await request.get('/dh114/reviews', { params })
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

// ===== 回复 =====
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
    await request.post(`/dh114/reviews/${replyForm.id}/replies`, { reply: replyForm.reply })
    ElMessage.success('回复成功')
    replyVisible.value = false
    await loadList()
  } catch (e) {
    // 失败已提示
  } finally {
    replyLoading.value = false
  }
}

// ===== 审核 =====
const onAudit = async (row, status) => {
  try {
    const label = auditText(status)
    await ElMessageBox.confirm(`确认将该评价设为「${label}」？`, '提示', { type: 'warning' })
    await request.put(`/dh114/admin/reviews/${row.id}/audit`, { audit_status: status })
    ElMessage.success('操作成功')
    await loadList()
  } catch (e) { /* cancel */ }
}

const onDelete = async (row) => {
  try {
    await ElMessageBox.confirm('确认删除该评价？', '提示', { type: 'warning' })
    await request.delete(`/dh114/reviews/${row.id}`)
    ElMessage.success('删除成功')
    await loadList()
  } catch (e) { /* cancel */ }
}

// ===== 新建 =====
const formVisible = ref(false)
const formLoading = ref(false)
const formRef = ref(null)
const form = reactive({ dh114_id: undefined, rating: 5, content: '', is_anonymous: 0 })
const rules = {
  dh114_id: [{ required: true, message: '请输入商户ID', trigger: 'blur' }],
  rating: [{ required: true, message: '请选择评分', trigger: 'change' }],
  content: [{ required: true, message: '请输入内容', trigger: 'blur' }]
}

const openCreate = () => {
  Object.assign(form, { dh114_id: undefined, rating: 5, content: '', is_anonymous: 0 })
  formVisible.value = true
}

const onSubmit = async () => {
  try {
    await formRef.value.validate()
    formLoading.value = true
    await request.post('/dh114/reviews', { ...form })
    ElMessage.success('创建成功')
    formVisible.value = false
    await loadList()
  } catch (e) {
    // 校验或接口失败
  } finally {
    formLoading.value = false
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
.toolbar { display: flex; justify-content: space-between; margin-bottom: 12px; }
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
