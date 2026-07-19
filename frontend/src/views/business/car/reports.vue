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
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value text-danger">{{ stats.appealing }}</div><div class="stat-label">申诉中</div></div></el-card>
      </el-col>
    </el-row>

    <div class="page-card">
      <el-form :inline="true" :model="filters" class="filter-form" @submit.prevent>
        <el-form-item label="关键词">
          <el-input v-model="filters.keyword" placeholder="举报单号/原因" clearable style="width: 200px" :prefix-icon="Search" @keyup.enter="onSearch" />
        </el-form-item>
        <el-form-item label="类型">
          <el-select v-model="filters.report_type" placeholder="全部" clearable style="width: 140px" @change="onSearch">
            <el-option v-for="(label, val) in typeMap" :key="val" :label="label" :value="val" />
          </el-select>
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="filters.status" placeholder="全部" clearable style="width: 140px" @change="onSearch">
            <el-option v-for="(label, val) in statusMap" :key="val" :label="label" :value="Number(val)" />
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
        <el-table-column label="车源" min-width="160" show-overflow-tooltip>
          <template #default="{ row }">{{ row.car_title || `车源#${row.car_id}` }}</template>
        </el-table-column>
        <el-table-column label="举报人" width="140">
          <template #default="{ row }">{{ row.reporter_name || `#${row.reporter_id}` }}</template>
        </el-table-column>
        <el-table-column label="类型" width="100">
          <template #default="{ row }"><el-tag size="small">{{ typeMap[row.report_type] || row.report_type }}</el-tag></template>
        </el-table-column>
        <el-table-column prop="reason" label="原因" min-width="180" show-overflow-tooltip />
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="statusTagType(row.status)" size="small">{{ statusMap[row.status] || '-' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="举报时间" width="160">
          <template #default="{ row }">{{ formatTime(row.created_at) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="220" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="openDetail(row)">详情</el-button>
            <el-button v-if="row.status === 0" type="warning" link size="small" @click="openProcess(row)">处理</el-button>
            <el-button v-if="row.status === 5" type="danger" link size="small" @click="openAppeal(row)">申诉处理</el-button>
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
        <el-descriptions-item label="车源">{{ detail.car_title || `车源#${detail.car_id}` }}</el-descriptions-item>
        <el-descriptions-item label="举报人">{{ detail.reporter_name || `#${detail.reporter_id}` }}</el-descriptions-item>
        <el-descriptions-item label="被举报人">{{ detail.reported_user_name || `#${detail.reported_user_id}` }}</el-descriptions-item>
        <el-descriptions-item label="类型">{{ typeMap[detail.report_type] || detail.report_type }}</el-descriptions-item>
        <el-descriptions-item label="状态">{{ statusMap[detail.status] }}</el-descriptions-item>
        <el-descriptions-item label="举报时间">{{ formatTime(detail.created_at) }}</el-descriptions-item>
        <el-descriptions-item label="原因" :span="2">{{ detail.reason }}</el-descriptions-item>
        <el-descriptions-item v-if="detail.description" label="详细描述" :span="2">{{ detail.description }}</el-descriptions-item>
        <el-descriptions-item v-if="detail.handle_result" label="处理结果" :span="2">{{ detail.handle_result }}</el-descriptions-item>
        <el-descriptions-item label="处理人">{{ detail.handler_name || `#${detail.handler_id}` }}</el-descriptions-item>
        <el-descriptions-item label="处理时间">{{ formatTime(detail.handled_at) }}</el-descriptions-item>
      </el-descriptions>
    </el-dialog>

    <el-dialog v-model="processVisible" title="处理举报" width="500px" destroy-on-close>
      <el-form :model="processForm" label-width="100px">
        <el-form-item label="处理结果">
          <el-radio-group v-model="processForm.status">
            <el-radio :value="1">核实警告</el-radio>
            <el-radio :value="2">下架车源</el-radio>
            <el-radio :value="3">封号</el-radio>
            <el-radio :value="4">驳回</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="处理说明">
          <el-input v-model="processForm.handle_result" type="textarea" :rows="3" placeholder="处理说明" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="processVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="onProcess">确认处理</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="appealVisible" title="申诉处理" width="500px" destroy-on-close>
      <el-form :model="appealForm" label-width="100px">
        <el-form-item label="申诉结果">
          <el-radio-group v-model="appealForm.appeal_result">
            <el-radio :value="1">维持原判</el-radio>
            <el-radio :value="2">申诉成立</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="申诉说明">
          <el-input v-model="appealForm.appeal_remark" type="textarea" :rows="3" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="appealVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="onAppeal">确认处理</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { Search, RefreshLeft, Refresh } from '@element-plus/icons-vue'
import { adminListReports, adminGetReport, processReport, processAppeal } from '@/api/car'
import { formatTime } from '@/utils/format'

const loading = ref(false)
const list = ref([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const filters = reactive({ keyword: '', report_type: '', status: '' })

const typeMap = { porn: '色情低俗', scam: '欺诈', fake: '虚假', accident: '事故车隐瞒', flooded: '泡水车隐瞒', burned: '火烧车隐瞒', price: '价格欺诈', other: '其他' }
const statusMap = { 0: '待处理', 1: '已核实警告', 2: '已下架', 3: '已封号', 4: '已驳回', 5: '申诉中', 6: '申诉已处理' }

const statusTagType = (s) => ({ 0: 'warning', 1: 'info', 2: 'danger', 3: 'danger', 4: 'success', 5: 'warning', 6: 'info' }[s] || 'info')

const stats = computed(() => ({
  total: list.value.length,
  pending: list.value.filter((r) => r.status === 0).length,
  handled: list.value.filter((r) => [1, 2, 3, 4, 6].includes(r.status)).length,
  appealing: list.value.filter((r) => r.status === 5).length
}))

const detailVisible = ref(false)
const detail = ref(null)
const processVisible = ref(false)
const appealVisible = ref(false)
const submitting = ref(false)
const processForm = reactive({ id: null, status: 1, handle_result: '' })
const appealForm = reactive({ id: null, appeal_result: 1, appeal_remark: '' })

const buildParams = () => {
  const p = { page: page.value, page_size: pageSize.value }
  if (filters.keyword) p.keyword = filters.keyword
  if (filters.report_type) p.report_type = filters.report_type
  if (filters.status !== '' && filters.status !== null) p.status = filters.status
  return p
}

const loadList = async () => {
  loading.value = true
  try {
    const res = await adminListReports(buildParams())
    const d = res.data || {}
    list.value = d.list || []
    total.value = d.total || 0
  } catch (e) { list.value = [] } finally { loading.value = false }
}

const onSearch = () => { page.value = 1; loadList() }
const onReset = () => { Object.assign(filters, { keyword: '', report_type: '', status: '' }); onSearch() }

const openDetail = async (row) => {
  try {
    const res = await adminGetReport(row.id)
    detail.value = res.data || row
  } catch (e) { detail.value = row }
  detailVisible.value = true
}

const openProcess = (row) => {
  Object.assign(processForm, { id: row.id, status: 1, handle_result: '' })
  processVisible.value = true
}

const openAppeal = (row) => {
  Object.assign(appealForm, { id: row.id, appeal_result: 1, appeal_remark: '' })
  appealVisible.value = true
}

const onProcess = async () => {
  try {
    submitting.value = true
    await processReport(processForm.id, { status: processForm.status, handle_result: processForm.handle_result })
    ElMessage.success('处理成功')
    processVisible.value = false
    loadList()
  } catch (e) {
    if (e?.message) ElMessage.error(e.message)
  } finally { submitting.value = false }
}

const onAppeal = async () => {
  try {
    submitting.value = true
    await processAppeal(appealForm.id, { appeal_result: appealForm.appeal_result, appeal_remark: appealForm.appeal_remark })
    ElMessage.success('申诉处理完成')
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
.stat-value { font-size: 24px; font-weight: 600; color: #409eff; }
.text-warning { color: #e6a23c; }
.text-success { color: #67c23a; }
.text-danger { color: #f56c6c; }
.page-card { background: #fff; padding: 16px; border-radius: 4px; box-shadow: 0 1px 4px rgba(0,0,0,0.04); }
.filter-form { margin-bottom: 12px; padding: 12px 16px; background: #fafafa; border-radius: 4px; }
.filter-form :deep(.el-form-item) { margin-bottom: 8px; margin-right: 12px; }
.toolbar { display: flex; justify-content: flex-start; margin-bottom: 12px; }
.pagination-wrap { margin-top: 16px; display: flex; justify-content: flex-end; }
</style>
