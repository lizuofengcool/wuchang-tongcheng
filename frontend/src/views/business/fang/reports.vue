<template>
  <div class="app-container">
    <el-row :gutter="16" class="stat-row">
      <el-col :xs="12" :sm="8" :md="6">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value">{{ stats.total }}</div><div class="stat-label">总举报</div></div></el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="6">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value text-warning">{{ stats.pending }}</div><div class="stat-label">待处理</div></div></el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="6">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value text-success">{{ stats.handled }}</div><div class="stat-label">已处理</div></div></el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="6">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value text-danger">{{ stats.appeal }}</div><div class="stat-label">申诉中</div></div></el-card>
      </el-col>
    </el-row>

    <div class="page-card">
      <el-form :inline="true" :model="filters" class="filter-form" @submit.prevent>
        <el-form-item label="关键词">
          <el-input v-model="filters.keyword" placeholder="举报单号/原因" clearable style="width: 200px" :prefix-icon="Search" @keyup.enter="onSearch" />
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="filters.status" placeholder="全部" clearable style="width: 140px" @change="onSearch">
            <el-option label="待处理" :value="0" />
            <el-option label="处理中" :value="1" />
            <el-option label="已处理" :value="2" />
            <el-option label="已申诉" :value="3" />
          </el-select>
        </el-form-item>
        <el-form-item label="类型">
          <el-select v-model="filters.report_type" placeholder="全部" clearable style="width: 140px" @change="onSearch">
            <el-option label="虚假信息" value="fake" />
            <el-option label="价格欺诈" value="price" />
            <el-option label="违规内容" value="illegal" />
            <el-option label="其他" value="other" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" :icon="Search" @click="onSearch">搜索</el-button>
          <el-button :icon="RefreshLeft" @click="onReset">重置</el-button>
        </el-form-item>
      </el-form>

      <div class="toolbar"><el-button :icon="Refresh" @click="loadList">刷新</el-button></div>

      <el-table v-loading="loading" :data="list" border stripe>
        <el-table-column prop="id" label="ID" width="70" fixed="left" />
        <el-table-column prop="report_no" label="举报单号" width="180" />
        <el-table-column label="举报人" width="140">
          <template #default="{ row }">{{ row.reporter_name || `#${row.reporter_id}` }}</template>
        </el-table-column>
        <el-table-column label="被举报" width="140">
          <template #default="{ row }">{{ row.reported_user_name || `#${row.reported_user_id}` }}</template>
        </el-table-column>
        <el-table-column label="类型" width="100">
          <template #default="{ row }">
            <el-tag size="small">{{ typeText(row.report_type) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="reason" label="原因" min-width="180" show-overflow-tooltip />
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="statusTagType(row.status)" size="small">{{ statusText(row.status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="时间" width="160">
          <template #default="{ row }">{{ formatTime(row.created_at) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="openDetail(row)">详情</el-button>
            <el-button v-if="row.status === 0" type="warning" link size="small" @click="openProcess(row)">处理</el-button>
            <el-button v-if="row.status === 3" type="primary" link size="small" @click="openAppeal(row)">申诉处理</el-button>
          </template>
        </el-table-column>
      </el-table>

      <div class="pagination-wrap">
        <el-pagination v-model:current-page="page" v-model:page-size="pageSize" :total="total" :page-sizes="[10, 20, 50]" layout="total, sizes, prev, pager, next, jumper" background @current-change="loadList" @size-change="loadList" />
      </div>
    </div>

    <el-dialog v-model="detailVisible" title="举报详情" width="700px" destroy-on-close>
      <el-descriptions v-if="detail" :column="2" border>
        <el-descriptions-item label="ID">{{ detail.id }}</el-descriptions-item>
        <el-descriptions-item label="举报单号">{{ detail.report_no }}</el-descriptions-item>
        <el-descriptions-item label="举报人">{{ detail.reporter_name || `#${detail.reporter_id}` }}</el-descriptions-item>
        <el-descriptions-item label="被举报人">{{ detail.reported_user_name || `#${detail.reported_user_id}` }}</el-descriptions-item>
        <el-descriptions-item label="类型">{{ typeText(detail.report_type) }}</el-descriptions-item>
        <el-descriptions-item label="状态">{{ statusText(detail.status) }}</el-descriptions-item>
        <el-descriptions-item label="原因" :span="2">{{ detail.reason }}</el-descriptions-item>
        <el-descriptions-item label="处理结果" :span="2">{{ detail.handle_result || '-' }}</el-descriptions-item>
        <el-descriptions-item label="创建时间">{{ formatTime(detail.created_at) }}</el-descriptions-item>
        <el-descriptions-item label="处理时间">{{ formatTime(detail.handled_at) }}</el-descriptions-item>
      </el-descriptions>
    </el-dialog>

    <el-dialog v-model="processVisible" title="处理举报" width="500px" destroy-on-close>
      <el-form :model="processForm" label-width="100px">
        <el-form-item label="处理结果">
          <el-select v-model="processForm.handle_result" placeholder="选择处理结果" style="width: 100%">
            <el-option label="警告" value="warn" />
            <el-option label="下架" value="takedown" />
            <el-option label="封号" value="ban" />
            <el-option label="驳回" value="dismiss" />
          </el-select>
        </el-form-item>
        <el-form-item label="处理说明">
          <el-input v-model="processForm.handle_remark" type="textarea" :rows="3" placeholder="处理说明" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="processVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="handleProcess">确定</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="appealVisible" title="申诉处理" width="500px" destroy-on-close>
      <el-form :model="appealForm" label-width="100px">
        <el-form-item label="处理结果">
          <el-select v-model="appealForm.appeal_result" placeholder="选择处理结果" style="width: 100%">
            <el-option label="申诉成立" value="uphold" />
            <el-option label="申诉驳回" value="reject" />
          </el-select>
        </el-form-item>
        <el-form-item label="处理说明">
          <el-input v-model="appealForm.appeal_remark" type="textarea" :rows="3" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="appealVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="handleAppeal">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { Search, RefreshLeft, Refresh } from '@element-plus/icons-vue'
import { adminListReports, adminGetReport, processReport, processAppeal } from '@/api/house'
import { formatTime } from '@/utils/format'

const loading = ref(false)
const list = ref([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const filters = reactive({ keyword: '', status: '', report_type: '' })
const stats = reactive({ total: 0, pending: 0, handled: 0, appeal: 0 })

const detailVisible = ref(false)
const detail = ref(null)
const processVisible = ref(false)
const appealVisible = ref(false)
const submitting = ref(false)
const processForm = reactive({ id: null, handle_result: '', handle_remark: '' })
const appealForm = reactive({ id: null, appeal_result: '', appeal_remark: '' })

const statusText = (s) => ({ 0: '待处理', 1: '处理中', 2: '已处理', 3: '已申诉' }[s] || '-')
const statusTagType = (s) => ({ 0: 'warning', 1: 'primary', 2: 'success', 3: 'danger' }[s] || 'info')
const typeText = (t) => ({ fake: '虚假信息', price: '价格欺诈', illegal: '违规内容', other: '其他' }[t] || t || '-')

const buildParams = () => {
  const p = { page: page.value, page_size: pageSize.value }
  if (filters.keyword) p.keyword = filters.keyword
  if (filters.status !== '' && filters.status !== null) p.status = filters.status
  if (filters.report_type) p.report_type = filters.report_type
  return p
}

const loadList = async () => {
  loading.value = true
  try {
    const res = await adminListReports(buildParams())
    const d = res.data || {}
    list.value = d.list || []
    total.value = d.total || 0
    if (d.stats) Object.assign(stats, d.stats)
  } catch (e) { list.value = [] } finally { loading.value = false }
}

const onSearch = () => { page.value = 1; loadList() }
const onReset = () => { Object.assign(filters, { keyword: '', status: '', report_type: '' }); onSearch() }

const openDetail = async (row) => {
  try {
    const res = await adminGetReport(row.id)
    detail.value = res.data || row
  } catch (e) { detail.value = row }
  detailVisible.value = true
}

const openProcess = (row) => {
  Object.assign(processForm, { id: row.id, handle_result: '', handle_remark: '' })
  processVisible.value = true
}

const handleProcess = async () => {
  try {
    submitting.value = true
    await processReport(processForm.id, { handle_result: processForm.handle_result, handle_remark: processForm.handle_remark })
    ElMessage.success('处理成功')
    processVisible.value = false
    loadList()
  } catch (e) {
    if (e?.message) ElMessage.error(e.message)
  } finally { submitting.value = false }
}

const openAppeal = (row) => {
  Object.assign(appealForm, { id: row.id, appeal_result: '', appeal_remark: '' })
  appealVisible.value = true
}

const handleAppeal = async () => {
  try {
    submitting.value = true
    await processAppeal(appealForm.id, { appeal_result: appealForm.appeal_result, appeal_remark: appealForm.appeal_remark })
    ElMessage.success('处理成功')
    appealVisible.value = false
    loadList()
  } catch (e) {
    if (e?.message) ElMessage.error(e.message)
  } finally { submitting.value = false }
}

onMounted(() => loadList())
</script>

<style scoped>
.stat-row { margin-bottom: 16px; }
.stat-item { text-align: center; padding: 16px; }
.stat-value { font-size: 22px; font-weight: 600; color: #303133; }
.text-warning { color: #e6a23c; }
.text-success { color: #67c23a; }
.text-danger { color: #f56c6c; }
</style>
