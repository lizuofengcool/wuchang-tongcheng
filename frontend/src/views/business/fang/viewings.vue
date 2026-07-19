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
          <el-input v-model="filters.keyword" placeholder="房源/用户" clearable style="width: 200px" :prefix-icon="Search" @keyup.enter="onSearch" />
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="filters.status" placeholder="全部" clearable style="width: 140px" @change="onSearch">
            <el-option label="待确认" :value="0" />
            <el-option label="已确认" :value="1" />
            <el-option label="已取消" :value="2" />
            <el-option label="已完成" :value="3" />
            <el-option label="已过期" :value="4" />
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

      <div class="toolbar"><el-button :icon="Refresh" @click="loadList">刷新</el-button></div>

      <el-table v-loading="loading" :data="list" border stripe>
        <el-table-column prop="id" label="ID" width="70" fixed="left" />
        <el-table-column prop="viewing_no" label="预约单号" width="160" />
        <el-table-column label="房源" min-width="180" show-overflow-tooltip>
          <template #default="{ row }">{{ row.house_title || `房源#${row.house_id}` }}</template>
        </el-table-column>
        <el-table-column label="预约人" width="140">
          <template #default="{ row }">
            <div>{{ row.user_name || `#${row.user_id}` }}</div>
            <div class="text-muted text-xs">{{ row.user_phone || '-' }}</div>
          </template>
        </el-table-column>
        <el-table-column label="看房时间" width="160">
          <template #default="{ row }">{{ formatTime(row.viewing_time) }}</template>
        </el-table-column>
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="statusTagType(row.status)" size="small">{{ statusText(row.status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="创建时间" width="160">
          <template #default="{ row }">{{ formatTime(row.created_at) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="openDetail(row)">详情</el-button>
            <el-button v-if="row.status === 0" type="success" link size="small" @click="handleAction(row, 'confirm')">确认</el-button>
            <el-button v-if="row.status === 0 || row.status === 1" type="warning" link size="small" @click="handleAction(row, 'cancel')">取消</el-button>
            <el-button v-if="row.status === 1" type="primary" link size="small" @click="handleAction(row, 'complete')">完成</el-button>
          </template>
        </el-table-column>
      </el-table>

      <div class="pagination-wrap">
        <el-pagination v-model:current-page="page" v-model:page-size="pageSize" :total="total" :page-sizes="[10, 20, 50]" layout="total, sizes, prev, pager, next, jumper" background @current-change="loadList" @size-change="loadList" />
      </div>
    </div>

    <el-dialog v-model="detailVisible" title="看房预约详情" width="700px" destroy-on-close>
      <el-descriptions v-if="detail" :column="2" border>
        <el-descriptions-item label="ID">{{ detail.id }}</el-descriptions-item>
        <el-descriptions-item label="预约单号">{{ detail.viewing_no }}</el-descriptions-item>
        <el-descriptions-item label="房源">{{ detail.house_title || `房源#${detail.house_id}` }}</el-descriptions-item>
        <el-descriptions-item label="预约人">{{ detail.user_name || `#${detail.user_id}` }}</el-descriptions-item>
        <el-descriptions-item label="手机">{{ detail.user_phone || '-' }}</el-descriptions-item>
        <el-descriptions-item label="看房时间">{{ formatTime(detail.viewing_time) }}</el-descriptions-item>
        <el-descriptions-item label="状态">{{ statusText(detail.status) }}</el-descriptions-item>
        <el-descriptions-item label="经纪人">{{ detail.agent_name || '-' }}</el-descriptions-item>
        <el-descriptions-item label="备注" :span="2">{{ detail.remark || '-' }}</el-descriptions-item>
      </el-descriptions>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Search, RefreshLeft, Refresh } from '@element-plus/icons-vue'
import { adminListViewings, getViewing, confirmViewing, cancelViewing, completeViewing } from '@/api/house'
import { formatTime } from '@/utils/format'

const loading = ref(false)
const list = ref([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const filters = reactive({ keyword: '', status: '', dateRange: null })
const stats = reactive({ total: 0, pending: 0, confirmed: 0, completed: 0 })

const detailVisible = ref(false)
const detail = ref(null)

const statusText = (s) => ({ 0: '待确认', 1: '已确认', 2: '已取消', 3: '已完成', 4: '已过期' }[s] || '-')
const statusTagType = (s) => ({ 0: 'warning', 1: 'primary', 2: 'info', 3: 'success', 4: 'danger' }[s] || 'info')

const buildParams = () => {
  const p = { page: page.value, page_size: pageSize.value }
  if (filters.keyword) p.keyword = filters.keyword
  if (filters.status !== '' && filters.status !== null) p.status = filters.status
  if (filters.dateRange && filters.dateRange.length === 2) {
    p.start_date = filters.dateRange[0]
    p.end_date = filters.dateRange[1]
  }
  return p
}

const loadList = async () => {
  loading.value = true
  try {
    const res = await adminListViewings(buildParams())
    const d = res.data || {}
    list.value = d.list || []
    total.value = d.total || 0
    if (d.stats) Object.assign(stats, d.stats)
  } catch (e) {
    list.value = []
  } finally {
    loading.value = false
  }
}

const onSearch = () => { page.value = 1; loadList() }
const onReset = () => {
  Object.assign(filters, { keyword: '', status: '', dateRange: null })
  onSearch()
}

const openDetail = async (row) => {
  try {
    const res = await getViewing(row.id)
    detail.value = res.data || row
  } catch (e) {
    detail.value = row
  }
  detailVisible.value = true
}

const handleAction = async (row, action) => {
  const actionMap = { confirm: confirmViewing, cancel: cancelViewing, complete: completeViewing }
  const labelMap = { confirm: '确认', cancel: '取消', complete: '完成' }
  try {
    await ElMessageBox.confirm(`确定${labelMap[action]}该预约吗？`, '提示', { type: 'warning' })
    await actionMap[action](row.id)
    ElMessage.success('操作成功')
    loadList()
  } catch (e) { /* cancel */ }
}

onMounted(() => loadList())
</script>

<style scoped>
.stat-row { margin-bottom: 16px; }
.stat-item { text-align: center; padding: 16px; }
.stat-value { font-size: 22px; font-weight: 600; color: #303133; }
.text-warning { color: #e6a23c; }
.text-primary { color: #409eff; }
.text-success { color: #67c23a; }
.text-muted { color: #909399; }
.text-xs { font-size: 12px; }
</style>
