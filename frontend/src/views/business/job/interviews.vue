<template>
  <div class="app-container">
    <el-row :gutter="16" class="stat-row">
      <el-col :xs="12" :sm="6"><el-card shadow="hover"><div class="stat-content"><div class="stat-value">{{ stats.total }}</div><div class="stat-label">总面试</div></div></el-card></el-col>
      <el-col :xs="12" :sm="6"><el-card shadow="hover"><div class="stat-content"><div class="stat-value">{{ stats.pending }}</div><div class="stat-label">待确认</div></div></el-card></el-col>
      <el-col :xs="12" :sm="6"><el-card shadow="hover"><div class="stat-content"><div class="stat-value">{{ stats.completed }}</div><div class="stat-label">已完成</div></div></el-card></el-col>
      <el-col :xs="12" :sm="6"><el-card shadow="hover"><div class="stat-content"><div class="stat-value">{{ stats.cancelled }}</div><div class="stat-label">已取消</div></div></el-card></el-col>
    </el-row>

    <div class="page-card">
      <el-form :inline="true" :model="filters" class="filter-form" @submit.prevent>
        <el-form-item label="关键词">
          <el-input v-model="filters.keyword" placeholder="职位/求职者" clearable style="width: 200px" @keyup.enter="onSearch" />
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="filters.status" placeholder="全部" clearable style="width: 130px" @change="onSearch">
            <el-option label="待确认" :value="0" />
            <el-option label="已确认" :value="1" />
            <el-option label="已改期" :value="2" />
            <el-option label="已完成" :value="3" />
            <el-option label="已取消" :value="4" />
            <el-option label="已拒绝" :value="5" />
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
        <el-table-column label="职位" min-width="140">
          <template #default="{ row }">{{ row.job_title || `职位#${row.job_id}` }}</template>
        </el-table-column>
        <el-table-column label="求职者" width="120">
          <template #default="{ row }">{{ row.applicant_name || `用户#${row.user_id}` }}</template>
        </el-table-column>
        <el-table-column label="面试时间" width="160">
          <template #default="{ row }">{{ formatTime(row.interview_time || row.scheduled_at) }}</template>
        </el-table-column>
        <el-table-column label="方式" width="80">
          <template #default="{ row }">{{ row.interview_type || row.type || '-' }}</template>
        </el-table-column>
        <el-table-column label="地点/链接" min-width="160" show-overflow-tooltip>
          <template #default="{ row }">{{ row.location || row.online_url || '-' }}</template>
        </el-table-column>
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="statusType(row.status)" size="small">{{ statusText(row.status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="220" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="openDetail(row)">详情</el-button>
            <el-button v-if="row.status === 0" type="success" link size="small" @click="handleAction(row, 'accept')">接受</el-button>
            <el-button v-if="row.status === 0" type="danger" link size="small" @click="handleAction(row, 'reject')">拒绝</el-button>
            <el-button v-if="row.status === 1 || row.status === 2" type="warning" link size="small" @click="openFeedback(row)">反馈</el-button>
          </template>
        </el-table-column>
      </el-table>

      <div class="pagination-wrap">
        <el-pagination v-model:current-page="page" v-model:page-size="pageSize" :total="total" :page-sizes="[10, 20, 50]" layout="total, sizes, prev, pager, next, jumper" background @current-change="loadList" @size-change="loadList" />
      </div>
    </div>

    <el-dialog v-model="detailVisible" title="面试详情" width="700px">
      <div v-loading="detailLoading">
        <el-descriptions v-if="detail" :column="2" border>
          <el-descriptions-item label="ID">{{ detail.id }}</el-descriptions-item>
          <el-descriptions-item label="状态"><el-tag :type="statusType(detail.status)" size="small">{{ statusText(detail.status) }}</el-tag></el-descriptions-item>
          <el-descriptions-item label="职位">{{ detail.job_title || `职位#${detail.job_id}` }}</el-descriptions-item>
          <el-descriptions-item label="求职者">{{ detail.applicant_name || `用户#${detail.user_id}` }}</el-descriptions-item>
          <el-descriptions-item label="面试时间">{{ formatTime(detail.interview_time || detail.scheduled_at) }}</el-descriptions-item>
          <el-descriptions-item label="方式">{{ detail.interview_type || detail.type || '-' }}</el-descriptions-item>
          <el-descriptions-item label="地点/链接" :span="2">{{ detail.location || detail.online_url || '-' }}</el-descriptions-item>
          <el-descriptions-item v-if="detail.feedback" label="反馈" :span="2">{{ detail.feedback }}</el-descriptions-item>
        </el-descriptions>
      </div>
      <template #footer><el-button @click="detailVisible = false">关闭</el-button></template>
    </el-dialog>

    <el-dialog v-model="feedbackVisible" title="面试反馈" width="560px">
      <el-form :model="feedbackForm" label-width="80px">
        <el-form-item label="评分">
          <el-rate v-model="feedbackForm.rating" />
        </el-form-item>
        <el-form-item label="反馈">
          <el-input v-model="feedbackForm.feedback" type="textarea" :rows="4" placeholder="面试反馈" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="feedbackVisible = false">取消</el-button>
        <el-button type="primary" :loading="feedbackLoading" @click="onFeedback">提交</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Refresh, RefreshLeft, Search } from '@element-plus/icons-vue'
import { listInterviews, getInterview, interviewAction, interviewFeedback, getInterviewStats } from '@/api/job'
import { formatTime } from '@/utils/format'

const loading = ref(false)
const page = ref(1); const pageSize = ref(20); const total = ref(0); const list = ref([])
const filters = reactive({ keyword: '', status: null })
const stats = reactive({ total: 0, pending: 0, completed: 0, cancelled: 0 })

const statusText = (s) => ({ 0: '待确认', 1: '已确认', 2: '已改期', 3: '已完成', 4: '已取消', 5: '已拒绝' }[s] || '-')
const statusType = (s) => ({ 0: 'warning', 1: 'success', 2: 'primary', 3: 'success', 4: 'info', 5: 'danger' }[s] || 'info')

const onSearch = () => { page.value = 1; loadList() }
const onReset = () => { Object.assign(filters, { keyword: '', status: null }); page.value = 1; loadList() }

const loadStats = async () => {
  try { const res = await getInterviewStats(); const d = res.data || {}; Object.assign(stats, { total: d.total || 0, pending: d.pending || 0, completed: d.completed || 0, cancelled: d.cancelled || 0 }) } catch (e) { /* */ }
}

const loadList = async () => {
  loading.value = true
  try {
    const params = { page: page.value, page_size: pageSize.value, keyword: filters.keyword || undefined, status: filters.status === null || filters.status === '' ? undefined : filters.status }
    const res = await listInterviews(params)
    const data = res.data || {}
    list.value = data.list || []; total.value = data.total || 0
  } catch (e) { list.value = []; total.value = 0 } finally { loading.value = false }
}

const detailVisible = ref(false); const detailLoading = ref(false); const detail = ref(null)
const openDetail = async (row) => {
  detailVisible.value = true; detailLoading.value = true
  try { const res = await getInterview(row.id); detail.value = res.data || row } catch (e) { detail.value = row } finally { detailLoading.value = false }
}

const handleAction = async (row, action) => {
  try {
    await ElMessageBox.confirm(`确定${action === 'accept' ? '接受' : '拒绝'}该面试吗？`, '提示', { type: 'warning' })
    await interviewAction(row.id, { action })
    ElMessage.success('操作成功'); await loadList()
  } catch (e) { /* */ }
}

const feedbackVisible = ref(false); const feedbackLoading = ref(false)
const feedbackForm = reactive({ id: null, rating: 3, feedback: '' })
const openFeedback = (row) => { Object.assign(feedbackForm, { id: row.id, rating: row.rating || 3, feedback: row.feedback || '' }); feedbackVisible.value = true }
const onFeedback = async () => {
  try {
    feedbackLoading.value = true
    await interviewFeedback(feedbackForm.id, { rating: feedbackForm.rating, feedback: feedbackForm.feedback })
    ElMessage.success('反馈成功'); feedbackVisible.value = false; await loadList()
  } catch (e) { /* */ } finally { feedbackLoading.value = false }
}

onMounted(async () => { await Promise.all([loadStats(), loadList()]) })
</script>

<style scoped>
.stat-row { margin-bottom: 16px; }
.stat-content { padding: 12px 16px; }
.stat-value { font-size: 24px; font-weight: 600; color: #303133; }
.stat-label { font-size: 13px; color: #909399; margin-top: 4px; }
.filter-form { margin-bottom: 12px; padding: 12px 16px; background: #fafafa; border-radius: 4px; }
.filter-form :deep(.el-form-item) { margin-bottom: 8px; margin-right: 12px; }
.toolbar { display: flex; justify-content: space-between; margin-bottom: 12px; }
.pagination-wrap { margin-top: 16px; display: flex; justify-content: flex-end; }
</style>
