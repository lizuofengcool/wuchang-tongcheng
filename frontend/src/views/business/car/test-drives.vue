<template>
  <div class="app-container">
    <el-row :gutter="16" class="stat-row">
      <el-col :xs="12" :sm="8" :md="6">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value">{{ stats.total }}</div><div class="stat-label">总预约</div></div></el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="6">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value text-warning">{{ stats.pending }}</div><div class="stat-label">待确认</div></div></el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="6">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value text-primary">{{ stats.confirmed }}</div><div class="stat-label">已确认</div></div></el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="6">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value text-success">{{ stats.completed }}</div><div class="stat-label">已完成</div></div></el-card>
      </el-col>
    </el-row>

    <div class="page-card">
      <el-form :inline="true" :model="filters" class="filter-form" @submit.prevent>
        <el-form-item label="关键词">
          <el-input v-model="filters.keyword" placeholder="车源/用户" clearable style="width: 200px" :prefix-icon="Search" @keyup.enter="onSearch" />
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="filters.status" placeholder="全部" clearable style="width: 140px" @change="onSearch">
            <el-option label="待确认" :value="0" />
            <el-option label="已确认" :value="1" />
            <el-option label="已完成" :value="2" />
            <el-option label="已取消" :value="3" />
            <el-option label="已爽约" :value="4" />
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
        <el-table-column prop="booking_no" label="预约单号" width="180" />
        <el-table-column label="车源" min-width="160" show-overflow-tooltip>
          <template #default="{ row }">{{ row.car_title || `车源#${row.car_id}` }}</template>
        </el-table-column>
        <el-table-column label="预约人" width="140">
          <template #default="{ row }">{{ row.user_name || `#${row.user_id}` }}</template>
        </el-table-column>
        <el-table-column label="联系电话" width="120">
          <template #default="{ row }">{{ row.phone || '-' }}</template>
        </el-table-column>
        <el-table-column label="试驾时间" width="160">
          <template #default="{ row }">{{ formatTime(row.drive_at) }}</template>
        </el-table-column>
        <el-table-column label="门店" width="140">
          <template #default="{ row }">{{ row.dealer_name || row.dealer || '-' }}</template>
        </el-table-column>
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="statusTagType(row.status)" size="small">{{ statusText(row.status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="220" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="openDetail(row)">详情</el-button>
            <el-button v-if="row.status === 0" type="success" link size="small" @click="handleAction(row, 1)">确认</el-button>
            <el-button v-if="row.status === 0" type="danger" link size="small" @click="handleAction(row, 3)">取消</el-button>
            <el-button v-if="row.status === 1" type="primary" link size="small" @click="handleAction(row, 2)">完成</el-button>
          </template>
        </el-table-column>
      </el-table>

      <div class="pagination-wrap">
        <el-pagination v-model:current-page="page" v-model:page-size="pageSize" :total="total" :page-sizes="[10, 20, 50]" layout="total, sizes, prev, pager, next, jumper" background @current-change="loadList" @size-change="loadList" />
      </div>
    </div>

    <el-dialog v-model="detailVisible" title="试驾预约详情" width="700px" destroy-on-close>
      <el-descriptions v-if="detail" :column="2" border>
        <el-descriptions-item label="ID">{{ detail.id }}</el-descriptions-item>
        <el-descriptions-item label="预约单号">{{ detail.booking_no }}</el-descriptions-item>
        <el-descriptions-item label="车源">{{ detail.car_title || `车源#${detail.car_id}` }}</el-descriptions-item>
        <el-descriptions-item label="预约人">{{ detail.user_name || `#${detail.user_id}` }}</el-descriptions-item>
        <el-descriptions-item label="联系电话">{{ detail.phone || '-' }}</el-descriptions-item>
        <el-descriptions-item label="试驾时间">{{ formatTime(detail.drive_at) }}</el-descriptions-item>
        <el-descriptions-item label="门店">{{ detail.dealer_name || detail.dealer || '-' }}</el-descriptions-item>
        <el-descriptions-item label="销售">{{ detail.sales_name || '-' }}</el-descriptions-item>
        <el-descriptions-item label="驾照号">{{ detail.license_no || '-' }}</el-descriptions-item>
        <el-descriptions-item label="状态">{{ statusText(detail.status) }}</el-descriptions-item>
        <el-descriptions-item label="创建时间">{{ formatTime(detail.created_at) }}</el-descriptions-item>
        <el-descriptions-item label="更新时间">{{ formatTime(detail.updated_at) }}</el-descriptions-item>
        <el-descriptions-item v-if="detail.remark" label="备注" :span="2">{{ detail.remark }}</el-descriptions-item>
      </el-descriptions>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Search, RefreshLeft, Refresh } from '@element-plus/icons-vue'
import { adminListTestDrives, adminGetTestDrive, adminUpdateTestDriveStatus } from '@/api/car'
import { formatTime } from '@/utils/format'

const loading = ref(false)
const list = ref([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const filters = reactive({ keyword: '', status: '' })

const detailVisible = ref(false)
const detail = ref(null)

const statusText = (s) => ({ 0: '待确认', 1: '已确认', 2: '已完成', 3: '已取消', 4: '已爽约' }[s] || '-')
const statusTagType = (s) => ({ 0: 'warning', 1: 'primary', 2: 'success', 3: 'info', 4: 'danger' }[s] || 'info')

const stats = computed(() => {
  const total = list.value.length
  const pending = list.value.filter((r) => r.status === 0).length
  const confirmed = list.value.filter((r) => r.status === 1).length
  const completed = list.value.filter((r) => r.status === 2).length
  return { total, pending, confirmed, completed }
})

const buildParams = () => {
  const p = { page: page.value, page_size: pageSize.value }
  if (filters.keyword) p.keyword = filters.keyword
  if (filters.status !== '' && filters.status !== null) p.status = filters.status
  return p
}

const loadList = async () => {
  loading.value = true
  try {
    const res = await adminListTestDrives(buildParams())
    const d = res.data || {}
    list.value = d.list || []
    total.value = d.total || 0
  } catch (e) { list.value = [] } finally { loading.value = false }
}

const onSearch = () => { page.value = 1; loadList() }
const onReset = () => { Object.assign(filters, { keyword: '', status: '' }); onSearch() }

const openDetail = async (row) => {
  try {
    const res = await adminGetTestDrive(row.id)
    detail.value = res.data || row
  } catch (e) { detail.value = row }
  detailVisible.value = true
}

const actionMap = { 1: '确认', 2: '完成', 3: '取消' }
const handleAction = async (row, status) => {
  try {
    await ElMessageBox.confirm(`确认${actionMap[status]}该试驾预约？`, '提示', { type: 'warning' })
    await adminUpdateTestDriveStatus(row.id, { status })
    ElMessage.success('操作成功')
    loadList()
  } catch (e) {
    if (e !== 'cancel' && e?.message) ElMessage.error(e.message)
  }
}

onMounted(() => loadList())
</script>

<style scoped>
.stat-row { margin-bottom: 16px; }
.stat-item { text-align: center; padding: 16px; }
.stat-value { font-size: 24px; font-weight: 600; color: #409eff; }
.text-warning { color: #e6a23c; }
.text-primary { color: #409eff; }
.text-success { color: #67c23a; }
.page-card { background: #fff; padding: 16px; border-radius: 4px; box-shadow: 0 1px 4px rgba(0,0,0,0.04); }
.filter-form { margin-bottom: 12px; padding: 12px 16px; background: #fafafa; border-radius: 4px; }
.filter-form :deep(.el-form-item) { margin-bottom: 8px; margin-right: 12px; }
.toolbar { display: flex; justify-content: flex-start; margin-bottom: 12px; }
.pagination-wrap { margin-top: 16px; display: flex; justify-content: flex-end; }
</style>
