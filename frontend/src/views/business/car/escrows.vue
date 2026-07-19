<template>
  <div class="app-container">
    <div class="page-card">
      <el-form :inline="true" :model="filters" class="filter-form" @submit.prevent>
        <el-form-item label="关键词">
          <el-input v-model="filters.keyword" placeholder="担保号/车源" clearable style="width: 200px" :prefix-icon="Search" @keyup.enter="onSearch" />
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="filters.status" placeholder="全部" clearable style="width: 140px" @change="onSearch">
            <el-option label="待付款" :value="0" />
            <el-option label="已付款" :value="1" />
            <el-option label="已放款" :value="2" />
            <el-option label="已退款" :value="3" />
            <el-option label="争议中" :value="4" />
            <el-option label="已取消" :value="5" />
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
        <el-table-column prop="escrow_no" label="担保号" width="180" />
        <el-table-column label="车源" min-width="160" show-overflow-tooltip>
          <template #default="{ row }">{{ row.car_title || `车源#${row.car_id}` }}</template>
        </el-table-column>
        <el-table-column label="付款方" width="140">
          <template #default="{ row }">{{ row.payer_name || `#${row.payer_id}` }}</template>
        </el-table-column>
        <el-table-column label="收款方" width="140">
          <template #default="{ row }">{{ row.payee_name || `#${row.payee_id}` }}</template>
        </el-table-column>
        <el-table-column label="金额" width="120">
          <template #default="{ row }">¥{{ Number(row.amount || 0).toFixed(2) }}万</template>
        </el-table-column>
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="statusTagType(row.status)" size="small">{{ statusText(row.status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="创建时间" width="160">
          <template #default="{ row }">{{ formatTime(row.created_at) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="220" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="openDetail(row)">详情</el-button>
            <el-select v-if="row.status === 4" v-model="row._action" placeholder="仲裁" size="small" style="width: 100px" @change="(val) => handleArbitrate(row, val)">
              <el-option label="放款" value="release" />
              <el-option label="退款" value="refund" />
            </el-select>
          </template>
        </el-table-column>
      </el-table>

      <div class="pagination-wrap">
        <el-pagination v-model:current-page="page" v-model:page-size="pageSize" :total="total" :page-sizes="[10, 20, 50]" layout="total, sizes, prev, pager, next, jumper" background @current-change="loadList" @size-change="loadList" />
      </div>
    </div>

    <el-dialog v-model="detailVisible" title="担保详情" width="700px" destroy-on-close>
      <el-descriptions v-if="detail" :column="2" border>
        <el-descriptions-item label="ID">{{ detail.id }}</el-descriptions-item>
        <el-descriptions-item label="担保号">{{ detail.escrow_no }}</el-descriptions-item>
        <el-descriptions-item label="车源">{{ detail.car_title || `车源#${detail.car_id}` }}</el-descriptions-item>
        <el-descriptions-item label="金额">¥{{ Number(detail.amount || 0).toFixed(2) }}万</el-descriptions-item>
        <el-descriptions-item label="付款方">{{ detail.payer_name || `#${detail.payer_id}` }}</el-descriptions-item>
        <el-descriptions-item label="收款方">{{ detail.payee_name || `#${detail.payee_id}` }}</el-descriptions-item>
        <el-descriptions-item label="状态">{{ statusText(detail.status) }}</el-descriptions-item>
        <el-descriptions-item label="创建时间">{{ formatTime(detail.created_at) }}</el-descriptions-item>
        <el-descriptions-item label="付款时间">{{ formatTime(detail.paid_at) }}</el-descriptions-item>
        <el-descriptions-item label="放款时间">{{ formatTime(detail.released_at) }}</el-descriptions-item>
        <el-descriptions-item label="备注" :span="2">{{ detail.remark || '-' }}</el-descriptions-item>
      </el-descriptions>
    </el-dialog>

    <el-dialog v-model="arbitrateVisible" title="担保仲裁" width="500px" destroy-on-close>
      <el-form :model="arbitrateForm" label-width="100px">
        <el-form-item label="仲裁结果">
          <el-select v-model="arbitrateForm.result" placeholder="选择仲裁结果" style="width: 100%">
            <el-option label="放款给收款方" value="release" />
            <el-option label="退款给付款方" value="refund" />
          </el-select>
        </el-form-item>
        <el-form-item label="仲裁说明">
          <el-input v-model="arbitrateForm.remark" type="textarea" :rows="3" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="arbitrateVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="handleArbitrateSubmit">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { Search, RefreshLeft, Refresh } from '@element-plus/icons-vue'
import { adminListEscrows, adminGetEscrow, adminUpdateEscrowStatus } from '@/api/car'
import { formatTime } from '@/utils/format'

const loading = ref(false)
const list = ref([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const filters = reactive({ keyword: '', status: '' })

const detailVisible = ref(false)
const detail = ref(null)
const arbitrateVisible = ref(false)
const submitting = ref(false)
const arbitrateForm = reactive({ id: null, result: '', remark: '' })

const statusText = (s) => ({ 0: '待付款', 1: '已付款', 2: '已放款', 3: '已退款', 4: '争议中', 5: '已取消' }[s] || '-')
const statusTagType = (s) => ({ 0: 'warning', 1: 'primary', 2: 'success', 3: 'info', 4: 'danger', 5: 'info' }[s] || 'info')

const buildParams = () => {
  const p = { page: page.value, page_size: pageSize.value }
  if (filters.keyword) p.keyword = filters.keyword
  if (filters.status !== '' && filters.status !== null) p.status = filters.status
  return p
}

const loadList = async () => {
  loading.value = true
  try {
    const res = await adminListEscrows(buildParams())
    const d = res.data || {}
    list.value = d.list || []
    total.value = d.total || 0
  } catch (e) { list.value = [] } finally { loading.value = false }
}

const onSearch = () => { page.value = 1; loadList() }
const onReset = () => { Object.assign(filters, { keyword: '', status: '' }); onSearch() }

const openDetail = async (row) => {
  try {
    const res = await adminGetEscrow(row.id)
    detail.value = res.data || row
  } catch (e) { detail.value = row }
  detailVisible.value = true
}

const handleArbitrate = (row, result) => {
  Object.assign(arbitrateForm, { id: row.id, result, remark: '' })
  arbitrateVisible.value = true
  row._action = null
}

const handleArbitrateSubmit = async () => {
  try {
    submitting.value = true
    const status = arbitrateForm.result === 'release' ? 2 : 3
    await adminUpdateEscrowStatus(arbitrateForm.id, { status, remark: arbitrateForm.remark })
    ElMessage.success('仲裁完成')
    arbitrateVisible.value = false
    loadList()
  } catch (e) {
    if (e?.message) ElMessage.error(e.message)
  } finally { submitting.value = false }
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
