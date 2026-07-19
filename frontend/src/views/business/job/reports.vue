<template>
  <div class="app-container">
    <div class="page-card">
      <el-form :inline="true" :model="filters" class="filter-form" @submit.prevent>
        <el-form-item label="关键词">
          <el-input v-model="filters.keyword" placeholder="举报单号/原因" clearable style="width: 200px" @keyup.enter="onSearch" />
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="filters.status" placeholder="全部" clearable style="width: 120px" @change="onSearch">
            <el-option label="待处理" :value="0" />
            <el-option label="处理中" :value="1" />
            <el-option label="已处理" :value="2" />
            <el-option label="已申诉" :value="3" />
          </el-select>
        </el-form-item>
        <el-form-item label="类型">
          <el-select v-model="filters.report_type" placeholder="全部" clearable style="width: 140px" @change="onSearch">
            <el-option label="职位举报" value="job" />
            <el-option label="公司举报" value="company" />
            <el-option label="评价举报" value="review" />
          </el-select>
        </el-form-item>
        <el-form-item label="时间">
          <el-date-picker v-model="filters.dateRange" type="daterange" range-separator="至" start-placeholder="开始" end-placeholder="结束" value-format="YYYY-MM-DD" style="width: 240px" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" :icon="Search" @click="onSearch">搜索</el-button>
          <el-button :icon="RefreshLeft" @click="onReset">重置</el-button>
        </el-form-item>
      </el-form>

      <div class="toolbar">
        <el-button :icon="Refresh" @click="loadList">刷新</el-button>
      </div>

      <el-table v-loading="loading" :data="list" border stripe>
        <el-table-column prop="id" label="ID" width="70" />
        <el-table-column prop="report_no" label="举报单号" width="160" />
        <el-table-column label="举报类型" width="100">
          <template #default="{ row }">
            <el-tag size="small">{{ row.report_type || '举报' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="举报原因" min-width="200" show-overflow-tooltip>
          <template #default="{ row }">{{ row.reason || row.description || '-' }}</template>
        </el-table-column>
        <el-table-column label="举报人" width="120">
          <template #default="{ row }">{{ row.reporter_name || `用户#${row.reporter_id}` }}</template>
        </el-table-column>
        <el-table-column label="被举报对象" width="160">
          <template #default="{ row }">{{ row.target_title || `#${row.target_id}` }}</template>
        </el-table-column>
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="reportStatusType(row.status)" size="small">{{ reportStatusText(row.status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="举报时间" width="160" prop="created_at" sortable>
          <template #default="{ row }">{{ formatTime(row.created_at) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="openDetail(row)">详情</el-button>
            <el-button v-if="row.status === 0 || row.status === 1" type="success" link size="small" @click="openProcess(row)">处理</el-button>
            <el-button v-if="row.status === 3" type="warning" link size="small" @click="openAppeal(row)">申诉处理</el-button>
          </template>
        </el-table-column>
      </el-table>

      <div class="pagination-wrap">
        <el-pagination v-model:current-page="page" v-model:page-size="pageSize" :total="total" :page-sizes="[10, 20, 50]" layout="total, sizes, prev, pager, next, jumper" background @current-change="loadList" @size-change="loadList" />
      </div>
    </div>

    <!-- 详情弹窗 -->
    <el-dialog v-model="detailVisible" title="举报详情" width="700px">
      <div v-loading="detailLoading">
        <el-descriptions v-if="detail" :column="2" border>
          <el-descriptions-item label="举报单号">{{ detail.report_no }}</el-descriptions-item>
          <el-descriptions-item label="状态">
            <el-tag :type="reportStatusType(detail.status)" size="small">{{ reportStatusText(detail.status) }}</el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="举报人">{{ detail.reporter_name || `用户#${detail.reporter_id}` }}</el-descriptions-item>
          <el-descriptions-item label="举报时间">{{ formatTime(detail.created_at) }}</el-descriptions-item>
          <el-descriptions-item label="举报原因" :span="2">{{ detail.reason || detail.description || '-' }}</el-descriptions-item>
          <el-descriptions-item v-if="detail.evidence" label="证据" :span="2">{{ detail.evidence }}</el-descriptions-item>
          <el-descriptions-item v-if="detail.handle_result" label="处理结果" :span="2">{{ detail.handle_result }}</el-descriptions-item>
        </el-descriptions>
      </div>
      <template #footer>
        <el-button @click="detailVisible = false">关闭</el-button>
      </template>
    </el-dialog>

    <!-- 处理弹窗 -->
    <el-dialog v-model="processVisible" title="处理举报" width="560px">
      <el-form :model="processForm" label-width="100px">
        <el-form-item label="处理结果">
          <el-select v-model="processForm.result" style="width: 100%">
            <el-option label="警告" value="warning" />
            <el-option label="下架内容" value="takedown" />
            <el-option label="封禁用户" value="ban" />
            <el-option label="驳回举报" value="dismiss" />
          </el-select>
        </el-form-item>
        <el-form-item label="处理说明">
          <el-input v-model="processForm.note" type="textarea" :rows="3" placeholder="处理说明" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="processVisible = false">取消</el-button>
        <el-button type="primary" :loading="processLoading" @click="onProcess">确认处理</el-button>
      </template>
    </el-dialog>

    <!-- 申诉处理弹窗 -->
    <el-dialog v-model="appealVisible" title="申诉处理" width="560px">
      <el-form :model="appealForm" label-width="100px">
        <el-form-item label="处理结果">
          <el-select v-model="appealForm.result" style="width: 100%">
            <el-option label="申诉成立" value="uphold" />
            <el-option label="申诉驳回" value="reject" />
          </el-select>
        </el-form-item>
        <el-form-item label="处理说明">
          <el-input v-model="appealForm.note" type="textarea" :rows="3" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="appealVisible = false">取消</el-button>
        <el-button type="primary" :loading="appealLoading" @click="onAppeal">确认</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { Refresh, RefreshLeft, Search } from '@element-plus/icons-vue'
import { adminListReports, adminGetReport, processReport, processAppeal } from '@/api/job'
import { formatTime } from '@/utils/format'

const loading = ref(false)
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)
const list = ref([])
const filters = reactive({ keyword: '', status: null, report_type: '', dateRange: null })

const reportStatusText = (s) => ({ 0: '待处理', 1: '处理中', 2: '已处理', 3: '已申诉' }[s] || '-')
const reportStatusType = (s) => ({ 0: 'warning', 1: 'primary', 2: 'success', 3: 'danger' }[s] || 'info')

const onSearch = () => { page.value = 1; loadList() }
const onReset = () => { Object.assign(filters, { keyword: '', status: null, report_type: '', dateRange: null }); page.value = 1; loadList() }

const loadList = async () => {
  loading.value = true
  try {
    const params = {
      page: page.value, page_size: pageSize.value,
      keyword: filters.keyword || undefined,
      status: filters.status === null || filters.status === '' ? undefined : filters.status,
      report_type: filters.report_type || undefined
    }
    if (filters.dateRange?.length === 2) { params.start_date = filters.dateRange[0]; params.end_date = filters.dateRange[1] }
    const res = await adminListReports(params)
    const data = res.data || {}
    list.value = data.list || []
    total.value = data.total || 0
  } catch (e) { list.value = []; total.value = 0 } finally { loading.value = false }
}

const detailVisible = ref(false)
const detailLoading = ref(false)
const detail = ref(null)
const openDetail = async (row) => {
  detailVisible.value = true; detailLoading.value = true
  try { const res = await adminGetReport(row.id); detail.value = res.data || row } catch (e) { detail.value = row } finally { detailLoading.value = false }
}

const processVisible = ref(false)
const processLoading = ref(false)
const processForm = reactive({ id: null, result: 'warning', note: '' })
const openProcess = (row) => { Object.assign(processForm, { id: row.id, result: 'warning', note: '' }); processVisible.value = true }
const onProcess = async () => {
  try {
    processLoading.value = true
    await processReport(processForm.id, { result: processForm.result, note: processForm.note })
    ElMessage.success('处理成功'); processVisible.value = false; await loadList()
  } catch (e) { /* */ } finally { processLoading.value = false }
}

const appealVisible = ref(false)
const appealLoading = ref(false)
const appealForm = reactive({ id: null, result: 'reject', note: '' })
const openAppeal = (row) => { Object.assign(appealForm, { id: row.id, result: 'reject', note: '' }); appealVisible.value = true }
const onAppeal = async () => {
  try {
    appealLoading.value = true
    await processAppeal(appealForm.id, { result: appealForm.result, note: appealForm.note })
    ElMessage.success('处理成功'); appealVisible.value = false; await loadList()
  } catch (e) { /* */ } finally { appealLoading.value = false }
}

onMounted(() => loadList())
</script>

<style scoped>
.filter-form { margin-bottom: 12px; padding: 12px 16px; background: #fafafa; border-radius: 4px; }
.filter-form :deep(.el-form-item) { margin-bottom: 8px; margin-right: 12px; }
.toolbar { display: flex; justify-content: space-between; margin-bottom: 12px; }
.pagination-wrap { margin-top: 16px; display: flex; justify-content: flex-end; }
</style>
