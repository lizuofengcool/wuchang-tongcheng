<template>
  <div class="app-container">
    <div class="page-card">
      <el-form :inline="true" :model="filters" class="filter-form" @submit.prevent>
        <el-form-item label="关键词">
          <el-input v-model="filters.keyword" placeholder="合同号/房源/用户" clearable style="width: 200px" :prefix-icon="Search" @keyup.enter="onSearch" />
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="filters.status" placeholder="全部" clearable style="width: 140px" @change="onSearch">
            <el-option label="待签署" :value="0" />
            <el-option label="已签署" :value="1" />
            <el-option label="已终止" :value="2" />
            <el-option label="已过期" :value="3" />
          </el-select>
        </el-form-item>
        <el-form-item label="类型">
          <el-select v-model="filters.contract_type" placeholder="全部" clearable style="width: 120px" @change="onSearch">
            <el-option label="出售" value="sale" />
            <el-option label="出租" value="rent" />
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
        <el-table-column label="房源" min-width="180" show-overflow-tooltip>
          <template #default="{ row }">{{ row.house_title || `房源#${row.house_id}` }}</template>
        </el-table-column>
        <el-table-column label="类型" width="80">
          <template #default="{ row }">
            <el-tag size="small" :type="row.contract_type === 'rent' ? 'warning' : 'primary'">{{ row.contract_type === 'rent' ? '出租' : '出售' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="甲方" width="140">
          <template #default="{ row }">{{ row.party_a_name || `#${row.party_a_id}` }}</template>
        </el-table-column>
        <el-table-column label="乙方" width="140">
          <template #default="{ row }">{{ row.party_b_name || `#${row.party_b_id}` }}</template>
        </el-table-column>
        <el-table-column label="金额" width="120">
          <template #default="{ row }">¥{{ Number(row.amount || 0).toFixed(2) }}</template>
        </el-table-column>
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="statusTagType(row.status)" size="small">{{ statusText(row.status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="签署时间" width="160">
          <template #default="{ row }">{{ formatTime(row.signed_at) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="180" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="openDetail(row)">详情</el-button>
            <el-button v-if="row.status === 0" type="success" link size="small" @click="handleAction(row, 'sign')">签署</el-button>
            <el-button v-if="row.status === 1" type="warning" link size="small" @click="handleAction(row, 'terminate')">终止</el-button>
          </template>
        </el-table-column>
      </el-table>

      <div class="pagination-wrap">
        <el-pagination v-model:current-page="page" v-model:page-size="pageSize" :total="total" :page-sizes="[10, 20, 50]" layout="total, sizes, prev, pager, next, jumper" background @current-change="loadList" @size-change="loadList" />
      </div>
    </div>

    <el-dialog v-model="detailVisible" title="合同详情" width="800px" destroy-on-close>
      <el-descriptions v-if="detail" :column="2" border>
        <el-descriptions-item label="ID">{{ detail.id }}</el-descriptions-item>
        <el-descriptions-item label="合同号">{{ detail.contract_no }}</el-descriptions-item>
        <el-descriptions-item label="房源">{{ detail.house_title || `房源#${detail.house_id}` }}</el-descriptions-item>
        <el-descriptions-item label="类型">{{ detail.contract_type === 'rent' ? '出租' : '出售' }}</el-descriptions-item>
        <el-descriptions-item label="甲方">{{ detail.party_a_name || `#${detail.party_a_id}` }}</el-descriptions-item>
        <el-descriptions-item label="乙方">{{ detail.party_b_name || `#${detail.party_b_id}` }}</el-descriptions-item>
        <el-descriptions-item label="金额">¥{{ Number(detail.amount || 0).toFixed(2) }}</el-descriptions-item>
        <el-descriptions-item label="状态">{{ statusText(detail.status) }}</el-descriptions-item>
        <el-descriptions-item label="起始">{{ formatTime(detail.start_date) }}</el-descriptions-item>
        <el-descriptions-item label="结束">{{ formatTime(detail.end_date) }}</el-descriptions-item>
        <el-descriptions-item label="签署时间">{{ formatTime(detail.signed_at) }}</el-descriptions-item>
        <el-descriptions-item label="创建时间">{{ formatTime(detail.created_at) }}</el-descriptions-item>
        <el-descriptions-item label="条款" :span="2">{{ detail.terms || '-' }}</el-descriptions-item>
      </el-descriptions>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Search, RefreshLeft, Refresh } from '@element-plus/icons-vue'
import { adminListContracts, getContract, signContract, terminateContract } from '@/api/house'
import { formatTime } from '@/utils/format'

const loading = ref(false)
const list = ref([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const filters = reactive({ keyword: '', status: '', contract_type: '' })

const detailVisible = ref(false)
const detail = ref(null)

const statusText = (s) => ({ 0: '待签署', 1: '已签署', 2: '已终止', 3: '已过期' }[s] || '-')
const statusTagType = (s) => ({ 0: 'warning', 1: 'success', 2: 'danger', 3: 'info' }[s] || 'info')

const buildParams = () => {
  const p = { page: page.value, page_size: pageSize.value }
  if (filters.keyword) p.keyword = filters.keyword
  if (filters.status !== '' && filters.status !== null) p.status = filters.status
  if (filters.contract_type) p.contract_type = filters.contract_type
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
const onReset = () => { Object.assign(filters, { keyword: '', status: '', contract_type: '' }); onSearch() }

const openDetail = async (row) => {
  try {
    const res = await getContract(row.id)
    detail.value = res.data || row
  } catch (e) { detail.value = row }
  detailVisible.value = true
}

const handleAction = async (row, action) => {
  const actionMap = { sign: signContract, terminate: terminateContract }
  const labelMap = { sign: '签署', terminate: '终止' }
  try {
    await ElMessageBox.confirm(`确定${labelMap[action]}该合同吗？`, '提示', { type: 'warning' })
    await actionMap[action](row.id)
    ElMessage.success('操作成功')
    loadList()
  } catch (e) { /* cancel */ }
}

onMounted(() => loadList())
</script>
