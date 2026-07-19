<template>
  <div class="app-container">
    <div class="page-card">
      <el-form :inline="true" :model="filters" class="filter-form" @submit.prevent>
        <el-form-item label="关键词">
          <el-input v-model="filters.keyword" placeholder="成交号/房源" clearable style="width: 200px" :prefix-icon="Search" @keyup.enter="onSearch" />
        </el-form-item>
        <el-form-item label="类型">
          <el-select v-model="filters.deal_type" placeholder="全部" clearable style="width: 120px" @change="onSearch">
            <el-option label="出售" value="sale" />
            <el-option label="出租" value="rent" />
          </el-select>
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="filters.status" placeholder="全部" clearable style="width: 140px" @change="onSearch">
            <el-option label="待确认" :value="0" />
            <el-option label="已确认" :value="1" />
            <el-option label="已完成" :value="2" />
            <el-option label="已取消" :value="3" />
          </el-select>
        </el-form-item>
        <el-form-item label="时间">
          <el-date-picker v-model="filters.date_range" type="daterange" value-format="YYYY-MM-DD" range-separator="至" start-placeholder="开始" end-placeholder="结束" style="width: 240px" @change="onSearch" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" :icon="Search" @click="onSearch">搜索</el-button>
          <el-button :icon="RefreshLeft" @click="onReset">重置</el-button>
        </el-form-item>
      </el-form>

      <div class="toolbar"><el-button :icon="Refresh" @click="loadList">刷新</el-button></div>

      <el-table v-loading="loading" :data="list" border stripe>
        <el-table-column prop="id" label="ID" width="70" fixed="left" />
        <el-table-column prop="deal_no" label="成交号" width="180" />
        <el-table-column label="房源" min-width="180" show-overflow-tooltip>
          <template #default="{ row }">{{ row.house_title || `房源#${row.house_id}` }}</template>
        </el-table-column>
        <el-table-column label="类型" width="80">
          <template #default="{ row }">
            <el-tag :type="row.deal_type === 'rent' ? 'primary' : 'success'" size="small">{{ row.deal_type === 'rent' ? '租' : '售' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="买方/租客" width="140">
          <template #default="{ row }">{{ row.buyer_name || `#${row.buyer_id}` }}</template>
        </el-table-column>
        <el-table-column label="卖方/房东" width="140">
          <template #default="{ row }">{{ row.seller_name || `#${row.seller_id}` }}</template>
        </el-table-column>
        <el-table-column label="成交金额" width="120">
          <template #default="{ row }">{{ formatPrice(row) }}</template>
        </el-table-column>
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="statusTagType(row.status)" size="small">{{ statusText(row.status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="成交时间" width="160">
          <template #default="{ row }">{{ formatTime(row.deal_at || row.created_at) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="160" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="openDetail(row)">详情</el-button>
            <el-button v-if="row.status === 1" type="success" link size="small" @click="handleComplete(row)">完成</el-button>
          </template>
        </el-table-column>
      </el-table>

      <div class="pagination-wrap">
        <el-pagination v-model:current-page="page" v-model:page-size="pageSize" :total="total" :page-sizes="[10, 20, 50]" layout="total, sizes, prev, pager, next, jumper" background @current-change="loadList" @size-change="loadList" />
      </div>
    </div>

    <el-dialog v-model="detailVisible" title="成交详情" width="700px" destroy-on-close>
      <el-descriptions v-if="detail" :column="2" border>
        <el-descriptions-item label="ID">{{ detail.id }}</el-descriptions-item>
        <el-descriptions-item label="成交号">{{ detail.deal_no }}</el-descriptions-item>
        <el-descriptions-item label="房源">{{ detail.house_title || `房源#${detail.house_id}` }}</el-descriptions-item>
        <el-descriptions-item label="类型">{{ detail.deal_type === 'rent' ? '出租' : '出售' }}</el-descriptions-item>
        <el-descriptions-item label="买方/租客">{{ detail.buyer_name || `#${detail.buyer_id}` }}</el-descriptions-item>
        <el-descriptions-item label="卖方/房东">{{ detail.seller_name || `#${detail.seller_id}` }}</el-descriptions-item>
        <el-descriptions-item label="成交金额">{{ formatPrice(detail) }}</el-descriptions-item>
        <el-descriptions-item label="状态">{{ statusText(detail.status) }}</el-descriptions-item>
        <el-descriptions-item label="成交时间">{{ formatTime(detail.deal_at) }}</el-descriptions-item>
        <el-descriptions-item label="创建时间">{{ formatTime(detail.created_at) }}</el-descriptions-item>
        <el-descriptions-item label="经纪人">{{ detail.agent_name || '-' }}</el-descriptions-item>
        <el-descriptions-item label="佣金">¥{{ Number(detail.commission || 0).toFixed(2) }}</el-descriptions-item>
        <el-descriptions-item v-if="detail.remark" label="备注" :span="2">{{ detail.remark }}</el-descriptions-item>
      </el-descriptions>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Search, RefreshLeft, Refresh } from '@element-plus/icons-vue'
import { adminListDeals, getDeal, completeDeal } from '@/api/house'
import { formatTime } from '@/utils/format'

const loading = ref(false)
const list = ref([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const filters = reactive({ keyword: '', deal_type: '', status: '', date_range: [] })

const detailVisible = ref(false)
const detail = ref(null)

const statusText = (s) => ({ 0: '待确认', 1: '已确认', 2: '已完成', 3: '已取消' }[s] || '-')
const statusTagType = (s) => ({ 0: 'warning', 1: 'primary', 2: 'success', 3: 'info' }[s] || 'info')

const formatPrice = (row) => {
  if (!row) return '-'
  const amount = Number(row.amount || row.deal_price || 0)
  if (row.deal_type === 'rent') return `¥${amount.toFixed(2)}/月`
  return amount >= 10000 ? `¥${(amount / 10000).toFixed(2)}万` : `¥${amount.toFixed(2)}`
}

const buildParams = () => {
  const p = { page: page.value, page_size: pageSize.value }
  if (filters.keyword) p.keyword = filters.keyword
  if (filters.deal_type) p.deal_type = filters.deal_type
  if (filters.status !== '' && filters.status !== null) p.status = filters.status
  if (filters.date_range && filters.date_range.length === 2) {
    p.start_date = filters.date_range[0]
    p.end_date = filters.date_range[1]
  }
  return p
}

const loadList = async () => {
  loading.value = true
  try {
    const res = await adminListDeals(buildParams())
    const d = res.data || {}
    list.value = d.list || []
    total.value = d.total || 0
  } catch (e) { list.value = [] } finally { loading.value = false }
}

const onSearch = () => { page.value = 1; loadList() }
const onReset = () => { Object.assign(filters, { keyword: '', deal_type: '', status: '', date_range: [] }); onSearch() }

const openDetail = async (row) => {
  try {
    const res = await getDeal(row.id)
    detail.value = res.data || row
  } catch (e) { detail.value = row }
  detailVisible.value = true
}

const handleComplete = async (row) => {
  try {
    await ElMessageBox.confirm('确认将该成交标记为已完成？', '提示', { type: 'warning' })
    await completeDeal(row.id)
    ElMessage.success('已完成')
    loadList()
  } catch (e) {
    if (e !== 'cancel' && e?.message) ElMessage.error(e.message)
  }
}

onMounted(() => loadList())
</script>

<style scoped>
.filter-form { margin-bottom: 12px; padding: 12px 16px; background: #fafafa; border-radius: 4px; }
.filter-form :deep(.el-form-item) { margin-bottom: 8px; margin-right: 12px; }
.toolbar { display: flex; justify-content: flex-start; margin-bottom: 12px; }
.pagination-wrap { margin-top: 16px; display: flex; justify-content: flex-end; }
</style>
