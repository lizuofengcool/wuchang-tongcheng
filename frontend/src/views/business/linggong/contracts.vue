<template>
  <div class="app-container">
    <!-- 顶部统计卡片 -->
    <el-row :gutter="16" class="stat-row">
      <el-col :xs="12" :sm="8" :md="6">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value">{{ stats.total }}</div><div class="stat-label">合同总数</div></div></el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="6">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value text-primary">{{ stats.pending }}</div><div class="stat-label">待签署</div></div></el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="6">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value text-warning">{{ stats.active }}</div><div class="stat-label">履行中</div></div></el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="6">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value text-success">{{ stats.completed }}</div><div class="stat-label">已完成</div></div></el-card>
      </el-col>
    </el-row>

    <div class="page-card">
      <el-form :inline="true" :model="filters" class="filter-form" @submit.prevent>
        <el-form-item label="关键词">
          <el-input v-model="filters.keyword" placeholder="合同编号" clearable style="width: 200px" @keyup.enter="onSearch" />
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="filters.status" placeholder="全部" clearable style="width: 140px" @change="onSearch">
            <el-option v-for="(label, val) in statusMap" :key="val" :label="label" :value="Number(val)" />
          </el-select>
        </el-form-item>
        <el-form-item label="雇主ID">
          <el-input v-model="filters.employer_id" placeholder="雇主ID" clearable style="width: 120px" @keyup.enter="onSearch" />
        </el-form-item>
        <el-form-item label="工人ID">
          <el-input v-model="filters.worker_id" placeholder="工人ID" clearable style="width: 120px" @keyup.enter="onSearch" />
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
        <el-table-column prop="id" label="ID" width="70" fixed="left" />
        <el-table-column prop="contract_no" label="合同编号" width="180" />
        <el-table-column label="岗位" min-width="160">
          <template #default="{ row }">
            <span>{{ row.linggong_title || `岗位#${row.linggong_id}` }}</span>
          </template>
        </el-table-column>
        <el-table-column label="雇主" width="140">
          <template #default="{ row }">
            <div>{{ row.employer_name || `#${row.employer_id}` }}</div>
          </template>
        </el-table-column>
        <el-table-column label="工人" width="140">
          <template #default="{ row }">
            <div>{{ row.worker_name || `#${row.worker_id}` }}</div>
          </template>
        </el-table-column>
        <el-table-column label="合同金额" width="110">
          <template #default="{ row }">¥{{ Number(row.amount || 0).toFixed(2) }}</template>
        </el-table-column>
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="statusTagType(row.status)" size="small">{{ statusMap[row.status] || '-' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="开始时间" width="120">
          <template #default="{ row }">{{ formatTime(row.start_date, 'YYYY-MM-DD') }}</template>
        </el-table-column>
        <el-table-column label="结束时间" width="120">
          <template #default="{ row }">{{ formatTime(row.end_date, 'YYYY-MM-DD') }}</template>
        </el-table-column>
        <el-table-column label="创建时间" width="160">
          <template #default="{ row }">{{ formatTime(row.created_at) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="160" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="openDetail(row)">详情</el-button>
            <el-button v-if="row.status === 0 || row.status === 1" type="warning" link size="small" @click="onUpdateStatus(row, 2)">终止</el-button>
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

    <!-- 详情弹窗 -->
    <el-dialog v-model="detailVisible" title="合同详情" width="700px">
      <el-descriptions v-if="detail" :column="2" border>
        <el-descriptions-item label="ID">{{ detail.id }}</el-descriptions-item>
        <el-descriptions-item label="合同编号">{{ detail.contract_no }}</el-descriptions-item>
        <el-descriptions-item label="岗位" :span="2">{{ detail.linggong_title || `岗位#${detail.linggong_id}` }}</el-descriptions-item>
        <el-descriptions-item label="雇主">{{ detail.employer_name || `#${detail.employer_id}` }}</el-descriptions-item>
        <el-descriptions-item label="工人">{{ detail.worker_name || `#${detail.worker_id}` }}</el-descriptions-item>
        <el-descriptions-item label="合同金额">¥{{ Number(detail.amount || 0).toFixed(2) }}</el-descriptions-item>
        <el-descriptions-item label="状态">
          <el-tag :type="statusTagType(detail.status)" size="small">{{ statusMap[detail.status] || '-' }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="开始日期">{{ formatTime(detail.start_date, 'YYYY-MM-DD') }}</el-descriptions-item>
        <el-descriptions-item label="结束日期">{{ formatTime(detail.end_date, 'YYYY-MM-DD') }}</el-descriptions-item>
        <el-descriptions-item v-if="detail.signed_at" label="签署时间">{{ formatTime(detail.signed_at) }}</el-descriptions-item>
        <el-descriptions-item v-if="detail.completed_at" label="完成时间">{{ formatTime(detail.completed_at) }}</el-descriptions-item>
        <el-descriptions-item v-if="detail.terms" label="合同条款" :span="2">
          <div class="content-box">{{ detail.terms }}</div>
        </el-descriptions-item>
        <el-descriptions-item label="创建时间">{{ formatTime(detail.created_at) }}</el-descriptions-item>
        <el-descriptions-item label="更新时间">{{ formatTime(detail.updated_at) }}</el-descriptions-item>
      </el-descriptions>
      <template #footer><el-button @click="detailVisible = false">关闭</el-button></template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Refresh, RefreshLeft, Search } from '@element-plus/icons-vue'
import { adminListLinggongContracts, getLinggongContract, updateLinggongContractStatus } from '@/api/linggong'
import { formatTime } from '@/utils/format'

const loading = ref(false)
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)
const list = ref([])

const filters = reactive({ keyword: '', status: null, employer_id: '', worker_id: '' })

const statusMap = {
  0: '待签署', 1: '已签署', 2: '已终止',
  3: '履行中', 4: '已完成', 5: '已违约'
}
const statusTagType = (s) => ({
  0: 'info', 1: 'success', 2: 'danger',
  3: 'warning', 4: 'success', 5: 'danger'
}[s] || 'info')

const stats = computed(() => {
  const total = list.value.length
  const pending = list.value.filter((r) => r.status === 0).length
  const active = list.value.filter((r) => r.status === 3 || r.status === 1).length
  const completed = list.value.filter((r) => r.status === 4).length
  return { total, pending, active, completed }
})

const onSearch = () => { page.value = 1; loadList() }
const onReset = () => {
  filters.keyword = ''
  filters.status = null
  filters.employer_id = ''
  filters.worker_id = ''
  page.value = 1
  loadList()
}

const loadList = async () => {
  loading.value = true
  try {
    const res = await adminListLinggongContracts({
      page: page.value,
      page_size: pageSize.value,
      keyword: filters.keyword || undefined,
      status: filters.status === null || filters.status === '' ? undefined : filters.status,
      employer_id: filters.employer_id || undefined,
      worker_id: filters.worker_id || undefined
    })
    const data = res.data || {}
    list.value = data.list || []
    total.value = data.total || 0
  } catch (e) {
    list.value = []
    total.value = 0
  } finally {
    loading.value = false
  }
}

const detailVisible = ref(false)
const detail = ref(null)

const openDetail = async (row) => {
  try {
    const res = await getLinggongContract(row.id)
    detail.value = res.data || row
  } catch (e) {
    detail.value = row
  }
  detailVisible.value = true
}

const onUpdateStatus = async (row, status) => {
  try {
    const label = statusMap[status]
    await ElMessageBox.confirm(`确定将合同设为「${label}」吗？`, '提示', { type: 'warning' })
    await updateLinggongContractStatus(row.id, status)
    ElMessage.success('状态更新成功')
    await loadList()
  } catch (e) {
    // 取消
  }
}

onMounted(() => loadList())
</script>

<style scoped>
.stat-row { margin-bottom: 16px; }
.stat-item { text-align: center; padding: 16px; }
.stat-value { font-size: 22px; font-weight: 600; color: #303133; }
.text-primary { color: #409eff; }
.text-warning { color: #e6a23c; }
.text-success { color: #67c23a; }
.filter-form { margin-bottom: 12px; padding: 12px 16px; background: #fafafa; border-radius: 4px; }
.filter-form :deep(.el-form-item) { margin-bottom: 8px; margin-right: 12px; }
.toolbar { display: flex; justify-content: flex-start; margin-bottom: 12px; }
.content-box { white-space: pre-wrap; word-break: break-all; max-height: 200px; overflow-y: auto; }
.pagination-wrap { margin-top: 16px; display: flex; justify-content: flex-end; }
</style>
