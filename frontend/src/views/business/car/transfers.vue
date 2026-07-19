<template>
  <div class="app-container">
    <div class="page-card">
      <el-form :inline="true" :model="filters" class="filter-form" @submit.prevent>
        <el-form-item label="关键词">
          <el-input v-model="filters.keyword" placeholder="过户单号/车源" clearable style="width: 200px" :prefix-icon="Search" @keyup.enter="onSearch" />
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="filters.status" placeholder="全部" clearable style="width: 140px" @change="onSearch">
            <el-option label="待办理" :value="0" />
            <el-option label="办理中" :value="1" />
            <el-option label="已完成" :value="2" />
            <el-option label="已取消" :value="3" />
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
        <el-table-column prop="transfer_no" label="过户单号" width="180" />
        <el-table-column label="车源" min-width="160" show-overflow-tooltip>
          <template #default="{ row }">{{ row.car_title || `车源#${row.car_id}` }}</template>
        </el-table-column>
        <el-table-column label="原车主" width="140">
          <template #default="{ row }">{{ row.seller_name || `#${row.seller_id}` }}</template>
        </el-table-column>
        <el-table-column label="新车主" width="140">
          <template #default="{ row }">{{ row.buyer_name || `#${row.buyer_id}` }}</template>
        </el-table-column>
        <el-table-column label="成交价" width="120">
          <template #default="{ row }">¥{{ Number(row.deal_price || 0).toFixed(2) }}万</template>
        </el-table-column>
        <el-table-column label="过户费" width="120">
          <template #default="{ row }">¥{{ Number(row.fee || 0).toFixed(2) }}</template>
        </el-table-column>
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="statusTagType(row.status)" size="small">{{ statusText(row.status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="申请时间" width="160">
          <template #default="{ row }">{{ formatTime(row.created_at) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="220" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="openDetail(row)">详情</el-button>
            <el-button v-if="row.status === 0" type="success" link size="small" @click="handleStatus(row, 1)">办理</el-button>
            <el-button v-if="row.status === 1" type="primary" link size="small" @click="handleStatus(row, 2)">完成</el-button>
            <el-button v-if="row.status === 0 || row.status === 1" type="danger" link size="small" @click="handleStatus(row, 3)">取消</el-button>
          </template>
        </el-table-column>
      </el-table>

      <div class="pagination-wrap">
        <el-pagination v-model:current-page="page" v-model:page-size="pageSize" :total="total" :page-sizes="[10, 20, 50]" layout="total, sizes, prev, pager, next, jumper" background @current-change="loadList" @size-change="loadList" />
      </div>
    </div>

    <el-dialog v-model="detailVisible" title="过户详情" width="700px" destroy-on-close>
      <el-descriptions v-if="detail" :column="2" border>
        <el-descriptions-item label="ID">{{ detail.id }}</el-descriptions-item>
        <el-descriptions-item label="过户单号">{{ detail.transfer_no }}</el-descriptions-item>
        <el-descriptions-item label="车源">{{ detail.car_title || `车源#${detail.car_id}` }}</el-descriptions-item>
        <el-descriptions-item label="原车主">{{ detail.seller_name || `#${detail.seller_id}` }}</el-descriptions-item>
        <el-descriptions-item label="新车主">{{ detail.buyer_name || `#${detail.buyer_id}` }}</el-descriptions-item>
        <el-descriptions-item label="成交价">¥{{ Number(detail.deal_price || 0).toFixed(2) }}万</el-descriptions-item>
        <el-descriptions-item label="过户费">¥{{ Number(detail.fee || 0).toFixed(2) }}</el-descriptions-item>
        <el-descriptions-item label="状态">{{ statusText(detail.status) }}</el-descriptions-item>
        <el-descriptions-item label="车管所">{{ detail.vehicle_office || '-' }}</el-descriptions-item>
        <el-descriptions-item label="办理人">{{ detail.handler || '-' }}</el-descriptions-item>
        <el-descriptions-item label="申请时间">{{ formatTime(detail.created_at) }}</el-descriptions-item>
        <el-descriptions-item label="完成时间">{{ formatTime(detail.completed_at) }}</el-descriptions-item>
        <el-descriptions-item v-if="detail.remark" label="备注" :span="2">{{ detail.remark }}</el-descriptions-item>
      </el-descriptions>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Search, RefreshLeft, Refresh } from '@element-plus/icons-vue'
import { adminListTransfers, adminGetTransfer, adminUpdateTransferStatus } from '@/api/car'
import { formatTime } from '@/utils/format'

const loading = ref(false)
const list = ref([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const filters = reactive({ keyword: '', status: '' })

const detailVisible = ref(false)
const detail = ref(null)

const statusText = (s) => ({ 0: '待办理', 1: '办理中', 2: '已完成', 3: '已取消' }[s] || '-')
const statusTagType = (s) => ({ 0: 'warning', 1: 'primary', 2: 'success', 3: 'info' }[s] || 'info')

const buildParams = () => {
  const p = { page: page.value, page_size: pageSize.value }
  if (filters.keyword) p.keyword = filters.keyword
  if (filters.status !== '' && filters.status !== null) p.status = filters.status
  return p
}

const loadList = async () => {
  loading.value = true
  try {
    const res = await adminListTransfers(buildParams())
    const d = res.data || {}
    list.value = d.list || []
    total.value = d.total || 0
  } catch (e) { list.value = [] } finally { loading.value = false }
}

const onSearch = () => { page.value = 1; loadList() }
const onReset = () => { Object.assign(filters, { keyword: '', status: '' }); onSearch() }

const openDetail = async (row) => {
  try {
    const res = await adminGetTransfer(row.id)
    detail.value = res.data || row
  } catch (e) { detail.value = row }
  detailVisible.value = true
}

const actionTextMap = { 1: '办理', 2: '完成', 3: '取消' }
const handleStatus = async (row, status) => {
  try {
    await ElMessageBox.confirm(`确认${actionTextMap[status]}该过户申请？`, '提示', { type: 'warning' })
    await adminUpdateTransferStatus(row.id, { status })
    ElMessage.success('操作成功')
    loadList()
  } catch (e) {
    if (e !== 'cancel' && e?.message) ElMessage.error(e.message)
  }
}

onMounted(() => loadList())
</script>

<style scoped>
.page-card { background: #fff; padding: 16px; border-radius: 4px; box-shadow: 0 1px 4px rgba(0,0,0,0.04); }
.filter-form { margin-bottom: 12px; padding: 12px 16px; background: #fafafa; border-radius: 4px; }
.filter-form :deep(.el-form-item) { margin-bottom: 8px; margin-right: 12px; }
.toolbar { display: flex; justify-content: flex-start; margin-bottom: 12px; }
.pagination-wrap { margin-top: 16px; display: flex; justify-content: flex-end; }
</style>
