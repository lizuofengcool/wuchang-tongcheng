<template>
  <div class="app-container">
    <div class="page-card">
      <el-form :inline="true" :model="filters" class="filter-form" @submit.prevent>
        <el-form-item label="关键词">
          <el-input v-model="filters.keyword" placeholder="发布单号/标题" clearable style="width: 200px" :prefix-icon="Search" @keyup.enter="onSearch" />
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="filters.status" placeholder="全部" clearable style="width: 140px" @change="onSearch">
            <el-option label="待审核" :value="0" />
            <el-option label="已发布" :value="1" />
            <el-option label="已下架" :value="2" />
            <el-option label="已拒绝" :value="3" />
          </el-select>
        </el-form-item>
        <el-form-item label="房源类型">
          <el-select v-model="filters.house_type" placeholder="全部" clearable style="width: 120px" @change="onSearch">
            <el-option label="出售" value="sale" />
            <el-option label="出租" value="rent" />
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
        <el-table-column prop="id" label="ID" width="70" fixed="left" />
        <el-table-column prop="listing_no" label="发布单号" width="160" />
        <el-table-column label="房源" min-width="180" show-overflow-tooltip>
          <template #default="{ row }">
            <div>{{ row.title || `房源#${row.house_id}` }}</div>
            <div class="text-muted text-xs">{{ row.house_name || '-' }}</div>
          </template>
        </el-table-column>
        <el-table-column label="类型" width="80">
          <template #default="{ row }">
            <el-tag size="small" :type="row.house_type === 'rent' ? 'warning' : 'primary'">{{ row.house_type === 'rent' ? '出租' : '出售' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="发布者" width="140">
          <template #default="{ row }">{{ row.user_name || `#${row.user_id}` }}</template>
        </el-table-column>
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="statusTagType(row.status)" size="small">{{ statusText(row.status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="view_count" label="浏览" width="80" />
        <el-table-column prop="refresh_count" label="刷新" width="80" />
        <el-table-column label="发布时间" width="160">
          <template #default="{ row }">{{ formatTime(row.created_at) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="openDetail(row)">详情</el-button>
            <el-button v-if="row.status === 0" type="success" link size="small" @click="handleAudit(row, 1)">通过</el-button>
            <el-button v-if="row.status === 0" type="danger" link size="small" @click="handleAudit(row, 3)">拒绝</el-button>
            <el-button v-if="row.status === 1" type="warning" link size="small" @click="handleStatus(row, 2)">下架</el-button>
            <el-button v-if="row.status === 2" type="primary" link size="small" @click="handleStatus(row, 1)">上架</el-button>
          </template>
        </el-table-column>
      </el-table>

      <div class="pagination-wrap">
        <el-pagination v-model:current-page="page" v-model:page-size="pageSize" :total="total" :page-sizes="[10, 20, 50]" layout="total, sizes, prev, pager, next, jumper" background @current-change="loadList" @size-change="loadList" />
      </div>
    </div>

    <el-dialog v-model="detailVisible" title="发布单详情" width="700px" destroy-on-close>
      <el-descriptions v-if="detail" :column="2" border>
        <el-descriptions-item label="ID">{{ detail.id }}</el-descriptions-item>
        <el-descriptions-item label="发布单号">{{ detail.listing_no }}</el-descriptions-item>
        <el-descriptions-item label="房源">{{ detail.title || `房源#${detail.house_id}` }}</el-descriptions-item>
        <el-descriptions-item label="类型">{{ detail.house_type === 'rent' ? '出租' : '出售' }}</el-descriptions-item>
        <el-descriptions-item label="发布者">{{ detail.user_name || `#${detail.user_id}` }}</el-descriptions-item>
        <el-descriptions-item label="状态">{{ statusText(detail.status) }}</el-descriptions-item>
        <el-descriptions-item label="浏览">{{ detail.view_count || 0 }}</el-descriptions-item>
        <el-descriptions-item label="刷新次数">{{ detail.refresh_count || 0 }}</el-descriptions-item>
        <el-descriptions-item label="创建时间">{{ formatTime(detail.created_at) }}</el-descriptions-item>
        <el-descriptions-item label="发布时间">{{ formatTime(detail.published_at) }}</el-descriptions-item>
        <el-descriptions-item v-if="detail.audit_reason" label="审核原因" :span="2">{{ detail.audit_reason }}</el-descriptions-item>
      </el-descriptions>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Search, RefreshLeft, Refresh } from '@element-plus/icons-vue'
import { adminListListings, auditListing, adminUpdateListingStatus, getListing } from '@/api/house'
import { formatTime } from '@/utils/format'

const loading = ref(false)
const list = ref([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const filters = reactive({ keyword: '', status: '', house_type: '', dateRange: null })

const detailVisible = ref(false)
const detail = ref(null)

const statusText = (s) => ({ 0: '待审核', 1: '已发布', 2: '已下架', 3: '已拒绝' }[s] || '-')
const statusTagType = (s) => ({ 0: 'warning', 1: 'success', 2: 'info', 3: 'danger' }[s] || 'info')

const buildParams = () => {
  const p = { page: page.value, page_size: pageSize.value }
  if (filters.keyword) p.keyword = filters.keyword
  if (filters.status !== '' && filters.status !== null) p.status = filters.status
  if (filters.house_type) p.house_type = filters.house_type
  if (filters.dateRange && filters.dateRange.length === 2) {
    p.start_date = filters.dateRange[0]
    p.end_date = filters.dateRange[1]
  }
  return p
}

const loadList = async () => {
  loading.value = true
  try {
    const res = await adminListListings(buildParams())
    const d = res.data || {}
    list.value = d.list || []
    total.value = d.total || 0
  } catch (e) {
    list.value = []
  } finally {
    loading.value = false
  }
}

const onSearch = () => { page.value = 1; loadList() }
const onReset = () => {
  Object.assign(filters, { keyword: '', status: '', house_type: '', dateRange: null })
  onSearch()
}

const openDetail = async (row) => {
  try {
    const res = await getListing(row.id)
    detail.value = res.data || row
    detailVisible.value = true
  } catch (e) {
    detail.value = row
    detailVisible.value = true
  }
}

const handleAudit = async (row, status) => {
  try {
    if (status === 3) {
      const { value } = await ElMessageBox.prompt('请输入拒绝原因', '拒绝审核', { confirmButtonText: '确定', cancelButtonText: '取消' })
      await auditListing(row.id, { status, audit_reason: value || '' })
    } else {
      await ElMessageBox.confirm('确定审核通过吗？', '提示', { type: 'warning' })
      await auditListing(row.id, { status })
    }
    ElMessage.success('操作成功')
    loadList()
  } catch (e) { /* cancel */ }
}

const handleStatus = async (row, status) => {
  try {
    const label = statusText(status)
    await ElMessageBox.confirm(`确定设为「${label}」吗？`, '提示', { type: 'warning' })
    await adminUpdateListingStatus(row.id, { status })
    ElMessage.success('操作成功')
    loadList()
  } catch (e) { /* cancel */ }
}

onMounted(() => loadList())
</script>

<style scoped>
.text-muted { color: #909399; }
.text-xs { font-size: 12px; }
</style>
