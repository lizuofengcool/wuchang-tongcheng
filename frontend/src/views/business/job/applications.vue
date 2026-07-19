<template>
  <div class="app-container">
    <el-row :gutter="16" class="stat-row">
      <el-col :xs="12" :sm="6"><el-card shadow="hover"><div class="stat-content"><div class="stat-value">{{ stats.total }}</div><div class="stat-label">总投递</div></div></el-card></el-col>
      <el-col :xs="12" :sm="6"><el-card shadow="hover"><div class="stat-content"><div class="stat-value">{{ stats.pending }}</div><div class="stat-label">待处理</div></div></el-card></el-col>
      <el-col :xs="12" :sm="6"><el-card shadow="hover"><div class="stat-content"><div class="stat-value">{{ stats.interviewing }}</div><div class="stat-label">面试中</div></div></el-card></el-col>
      <el-col :xs="12" :sm="6"><el-card shadow="hover"><div class="stat-content"><div class="stat-value">{{ stats.hired }}</div><div class="stat-label">已录用</div></div></el-card></el-col>
    </el-row>

    <div class="page-card">
      <el-form :inline="true" :model="filters" class="filter-form" @submit.prevent>
        <el-form-item label="关键词">
          <el-input v-model="filters.keyword" placeholder="职位/求职者" clearable style="width: 200px" @keyup.enter="onSearch" />
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="filters.status" placeholder="全部" clearable style="width: 130px" @change="onSearch">
            <el-option label="已投递" :value="0" />
            <el-option label="已查看" :value="1" />
            <el-option label="已邀面试" :value="2" />
            <el-option label="面试中" :value="3" />
            <el-option label="已录用" :value="4" />
            <el-option label="已拒绝" :value="5" />
            <el-option label="已撤回" :value="6" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" :icon="Search" @click="onSearch">搜索</el-button>
          <el-button :icon="RefreshLeft" @click="onReset">重置</el-button>
        </el-form-item>
      </el-form>

      <div class="toolbar">
        <el-button :icon="Refresh" @click="loadList">刷新</el-button>
        <el-button type="warning" :disabled="!selection.length" @click="onBatch('viewed')">批量标记已查看</el-button>
      </div>

      <el-table v-loading="loading" :data="list" border stripe @selection-change="(r) => selection = r">
        <el-table-column type="selection" width="44" />
        <el-table-column prop="id" label="ID" width="70" />
        <el-table-column label="职位" min-width="160">
          <template #default="{ row }">{{ row.job_title || `职位#${row.job_id}` }}</template>
        </el-table-column>
        <el-table-column label="求职者" width="130">
          <template #default="{ row }">{{ row.applicant_name || `用户#${row.user_id}` }}</template>
        </el-table-column>
        <el-table-column label="简历" width="120">
          <template #default="{ row }">{{ row.resume_title || `简历#${row.resume_id}` }}</template>
        </el-table-column>
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="statusType(row.status)" size="small">{{ statusText(row.status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="投递时间" width="160" prop="created_at" sortable>
          <template #default="{ row }">{{ formatTime(row.created_at) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="openDetail(row)">详情</el-button>
            <el-dropdown trigger="click" @command="(cmd) => handleStatus(row, cmd)">
              <el-button type="warning" link size="small">状态变更<el-icon class="el-icon--right"><ArrowDown /></el-icon></el-button>
              <template #dropdown>
                <el-dropdown-menu>
                  <el-dropdown-item :command="1">已查看</el-dropdown-item>
                  <el-dropdown-item :command="2">邀面试</el-dropdown-item>
                  <el-dropdown-item :command="4">录用</el-dropdown-item>
                  <el-dropdown-item :command="5">拒绝</el-dropdown-item>
                </el-dropdown-menu>
              </template>
            </el-dropdown>
          </template>
        </el-table-column>
      </el-table>

      <div class="pagination-wrap">
        <el-pagination v-model:current-page="page" v-model:page-size="pageSize" :total="total" :page-sizes="[10, 20, 50]" layout="total, sizes, prev, pager, next, jumper" background @current-change="loadList" @size-change="loadList" />
      </div>
    </div>

    <el-dialog v-model="detailVisible" title="投递详情" width="700px">
      <div v-loading="detailLoading">
        <el-descriptions v-if="detail" :column="2" border>
          <el-descriptions-item label="ID">{{ detail.id }}</el-descriptions-item>
          <el-descriptions-item label="状态">
            <el-tag :type="statusType(detail.status)" size="small">{{ statusText(detail.status) }}</el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="职位">{{ detail.job_title || `职位#${detail.job_id}` }}</el-descriptions-item>
          <el-descriptions-item label="求职者">{{ detail.applicant_name || `用户#${detail.user_id}` }}</el-descriptions-item>
          <el-descriptions-item label="简历">{{ detail.resume_title || `简历#${detail.resume_id}` }}</el-descriptions-item>
          <el-descriptions-item label="投递时间">{{ formatTime(detail.created_at) }}</el-descriptions-item>
          <el-descriptions-item v-if="detail.cover_letter" label="求职信" :span="2">{{ detail.cover_letter }}</el-descriptions-item>
        </el-descriptions>
      </div>
      <template #footer><el-button @click="detailVisible = false">关闭</el-button></template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Refresh, RefreshLeft, Search, ArrowDown } from '@element-plus/icons-vue'
import { listApplications, getApplication, updateApplicationStatus, batchActionApplications, getApplicationStats } from '@/api/job'
import { formatTime } from '@/utils/format'

const loading = ref(false)
const page = ref(1); const pageSize = ref(20); const total = ref(0); const list = ref([])
const selection = ref([])
const filters = reactive({ keyword: '', status: null })
const stats = reactive({ total: 0, pending: 0, interviewing: 0, hired: 0 })

const statusText = (s) => ({ 0: '已投递', 1: '已查看', 2: '已邀面试', 3: '面试中', 4: '已录用', 5: '已拒绝', 6: '已撤回' }[s] || '-')
const statusType = (s) => ({ 0: 'info', 1: '', 2: 'warning', 3: 'primary', 4: 'success', 5: 'danger', 6: 'info' }[s] || 'info')

const onSearch = () => { page.value = 1; loadList() }
const onReset = () => { Object.assign(filters, { keyword: '', status: null }); page.value = 1; loadList() }

const loadStats = async () => {
  try { const res = await getApplicationStats(); const d = res.data || {}; Object.assign(stats, { total: d.total || 0, pending: d.pending || 0, interviewing: d.interviewing || 0, hired: d.hired || 0 }) } catch (e) { /* */ }
}

const loadList = async () => {
  loading.value = true
  try {
    const params = { page: page.value, page_size: pageSize.value, keyword: filters.keyword || undefined, status: filters.status === null || filters.status === '' ? undefined : filters.status }
    const res = await listApplications(params)
    const data = res.data || {}
    list.value = data.list || []; total.value = data.total || 0
  } catch (e) { list.value = []; total.value = 0 } finally { loading.value = false }
}

const detailVisible = ref(false); const detailLoading = ref(false); const detail = ref(null)
const openDetail = async (row) => {
  detailVisible.value = true; detailLoading.value = true
  try { const res = await getApplication(row.id); detail.value = res.data || row } catch (e) { detail.value = row } finally { detailLoading.value = false }
}

const handleStatus = async (row, status) => {
  try {
    const label = statusText(status)
    await ElMessageBox.confirm(`确定设为「${label}」吗？`, '提示', { type: 'warning' })
    await updateApplicationStatus(row.id, { status })
    ElMessage.success('状态更新成功'); await loadList()
  } catch (e) { /* */ }
}

const onBatch = async (action) => {
  try {
    await ElMessageBox.confirm(`确认批量操作 ${selection.value.length} 条投递记录？`, '批量操作', { type: 'warning' })
    await batchActionApplications({ ids: selection.value.map((r) => r.id), action })
    ElMessage.success('批量操作完成'); await loadList()
  } catch (e) { /* */ }
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
