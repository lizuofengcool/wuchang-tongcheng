<template>
  <div class="app-container">
    <div class="page-card">
      <el-form :inline="true" :model="filters" class="filter-form" @submit.prevent>
        <el-form-item label="关键词">
          <el-input v-model="filters.keyword" placeholder="合同号/车源" clearable style="width: 200px" :prefix-icon="Search" @keyup.enter="onSearch" />
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="filters.status" placeholder="全部" clearable style="width: 140px" @change="onSearch">
            <el-option label="待签署" :value="0" />
            <el-option label="已签署" :value="1" />
            <el-option label="已生效" :value="2" />
            <el-option label="已终止" :value="3" />
            <el-option label="已作废" :value="4" />
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
        <el-table-column prop="contract_no" label="合同号" width="180" />
        <el-table-column label="车源" min-width="160" show-overflow-tooltip>
          <template #default="{ row }">{{ row.car_title || `车源#${row.car_id}` }}</template>
        </el-table-column>
        <el-table-column label="买方" width="140">
          <template #default="{ row }">{{ row.buyer_name || `#${row.buyer_id}` }}</template>
        </el-table-column>
        <el-table-column label="卖方" width="140">
          <template #default="{ row }">{{ row.seller_name || `#${row.seller_id}` }}</template>
        </el-table-column>
        <el-table-column label="金额" width="120">
          <template #default="{ row }">¥{{ Number(row.amount || 0).toFixed(2) }}万</template>
        </el-table-column>
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="statusTagType(row.status)" size="small">{{ statusText(row.status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="签署时间" width="160">
          <template #default="{ row }">{{ formatTime(row.signed_at || row.created_at) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="160" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="openDetail(row)">详情</el-button>
            <el-button v-if="row.status === 0" type="success" link size="small" @click="handleAction(row, 'sign')">签署</el-button>
            <el-button v-if="row.status === 1 || row.status === 2" type="warning" link size="small" @click="handleAction(row, 'terminate')">终止</el-button>
          </template>
        </el-table-column>
      </el-table>

      <div class="pagination-wrap">
        <el-pagination v-model:current-page="page" v-model:page-size="pageSize" :total="total" :page-sizes="[10, 20, 50]" layout="total, sizes, prev, pager, next, jumper" background @current-change="loadList" @size-change="loadList" />
      </div>
    </div>

    <el-dialog v-model="detailVisible" title="合同详情" width="700px" destroy-on-close>
      <el-descriptions v-if="detail" :column="2" border>
        <el-descriptions-item label="ID">{{ detail.id }}</el-descriptions-item>
        <el-descriptions-item label="合同号">{{ detail.contract_no }}</el-descriptions-item>
        <el-descriptions-item label="车源">{{ detail.car_title || `车源#${detail.car_id}` }}</el-descriptions-item>
        <el-descriptions-item label="买方">{{ detail.buyer_name || `#${detail.buyer_id}` }}</el-descriptions-item>
        <el-descriptions-item label="卖方">{{ detail.seller_name || `#${detail.seller_id}` }}</el-descriptions-item>
        <el-descriptions-item label="金额">¥{{ Number(detail.amount || 0).toFixed(2) }}万</el-descriptions-item>
        <el-descriptions-item label="状态">{{ statusText(detail.status) }}</el-descriptions-item>
        <el-descriptions-item label="签署时间">{{ formatTime(detail.signed_at) }}</el-descriptions-item>
        <el-descriptions-item label="生效时间">{{ formatTime(detail.effective_at) }}</el-descriptions-item>
        <el-descriptions-item label="到期时间">{{ formatTime(detail.expire_at) }}</el-descriptions-item>
        <el-descriptions-item label="创建时间">{{ formatTime(detail.created_at) }}</el-descriptions-item>
        <el-descriptions-item v-if="detail.remark" label="备注" :span="2">{{ detail.remark }}</el-descriptions-item>
      </el-descriptions>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Search, RefreshLeft, Refresh } from '@element-plus/icons-vue'
import { adminListContracts, adminGetContract, adminUpdateContractStatus } from '@/api/car'
import { formatTime } from '@/utils/format'

const loading = ref(false)
const list = ref([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const filters = reactive({ keyword: '', status: '' })

const detailVisible = ref(false)
const detail = ref(null)

const statusText = (s) => ({ 0: '待签署', 1: '已签署', 2: '已生效', 3: '已终止', 4: '已作废' }[s] || '-')
const statusTagType = (s) => ({ 0: 'warning', 1: 'primary', 2: 'success', 3: 'info', 4: 'danger' }[s] || 'info')

const buildParams = () => {
  const p = { page: page.value, page_size: pageSize.value }
  if (filters.keyword) p.keyword = filters.keyword
  if (filters.status !== '' && filters.status !== null) p.status = filters.status
  return p
}

const loadList = async () => {
  loading.value = true
  try {
    const res = await adminListContracts(buildParams())
    const d = res.data || {}
    list.value = d.list || []
    total.value = d.total || 0
  } catch (e) { list.value = [] } finally { loading.value = false }
}

const onSearch = () => { page.value = 1; loadList() }
const onReset = () => { Object.assign(filters, { keyword: '', status: '' }); onSearch() }

const openDetail = async (row) => {
  try {
    const res = await adminGetContract(row.id)
    detail.value = res.data || row
  } catch (e) { detail.value = row }
  detailVisible.value = true
}

const actionMap = { sign: 1, terminate: 3 }
const actionTextMap = { sign: '签署', terminate: '终止' }
const handleAction = async (row, action) => {
  try {
    await ElMessageBox.confirm(`确认${actionTextMap[action]}该合同？`, '提示', { type: 'warning' })
    await adminUpdateContractStatus(row.id, { status: actionMap[action] })
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
