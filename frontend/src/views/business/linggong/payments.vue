<template>
  <div class="app-container">
    <!-- 顶部统计卡片 -->
    <el-row :gutter="16" class="stat-row">
      <el-col :xs="12" :sm="8" :md="6">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value">{{ stats.total }}</div><div class="stat-label">支付总数</div></div></el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="6">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value text-primary">¥{{ formatAmount(stats.totalAmount) }}</div><div class="stat-label">总金额</div></div></el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="6">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value text-warning">¥{{ formatAmount(stats.pendingAmount) }}</div><div class="stat-label">待结算</div></div></el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="6">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value text-success">¥{{ formatAmount(stats.settledAmount) }}</div><div class="stat-label">已结算</div></div></el-card>
      </el-col>
    </el-row>

    <div class="page-card">
      <el-form :inline="true" :model="filters" class="filter-form" @submit.prevent>
        <el-form-item label="关键词">
          <el-input v-model="filters.keyword" placeholder="支付编号" clearable style="width: 200px" @keyup.enter="onSearch" />
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="filters.status" placeholder="全部" clearable style="width: 140px" @change="onSearch">
            <el-option v-for="(label, val) in statusMap" :key="val" :label="label" :value="Number(val)" />
          </el-select>
        </el-form-item>
        <el-form-item label="时间范围">
          <el-date-picker
            v-model="filters.dateRange"
            type="daterange"
            range-separator="至"
            start-placeholder="开始日期"
            end-placeholder="结束日期"
            value-format="YYYY-MM-DD"
            style="width: 240px"
          />
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
        <el-table-column prop="payment_no" label="支付编号" width="180" />
        <el-table-column label="合同" min-width="140">
          <template #default="{ row }">
            <span>{{ row.contract_no || `合同#${row.contract_id}` }}</span>
          </template>
        </el-table-column>
        <el-table-column label="雇主" width="120">
          <template #default="{ row }">{{ row.employer_name || `#${row.employer_id}` }}</template>
        </el-table-column>
        <el-table-column label="工人" width="120">
          <template #default="{ row }">{{ row.worker_name || `#${row.worker_id}` }}</template>
        </el-table-column>
        <el-table-column label="金额" width="120">
          <template #default="{ row }"><span class="price">¥{{ Number(row.amount || 0).toFixed(2) }}</span></template>
        </el-table-column>
        <el-table-column label="手续费" width="100">
          <template #default="{ row }">¥{{ Number(row.fee || 0).toFixed(2) }}</template>
        </el-table-column>
        <el-table-column label="实发" width="100">
          <template #default="{ row }">¥{{ Number(row.net_amount || 0).toFixed(2) }}</template>
        </el-table-column>
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="statusTagType(row.status)" size="small">{{ statusMap[row.status] || '-' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="支付方式" width="100">
          <template #default="{ row }">{{ payMethodText(row.pay_method) }}</template>
        </el-table-column>
        <el-table-column label="创建时间" width="160">
          <template #default="{ row }">{{ formatTime(row.created_at) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="180" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="openDetail(row)">详情</el-button>
            <el-button v-if="row.status === 1 || row.status === 2" type="success" link size="small" @click="onSettle(row)">结算</el-button>
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
    <el-dialog v-model="detailVisible" title="支付详情" width="700px">
      <el-descriptions v-if="detail" :column="2" border>
        <el-descriptions-item label="ID">{{ detail.id }}</el-descriptions-item>
        <el-descriptions-item label="支付编号">{{ detail.payment_no }}</el-descriptions-item>
        <el-descriptions-item label="合同" :span="2">{{ detail.contract_no || `合同#${detail.contract_id}` }}</el-descriptions-item>
        <el-descriptions-item label="雇主">{{ detail.employer_name || `#${detail.employer_id}` }}</el-descriptions-item>
        <el-descriptions-item label="工人">{{ detail.worker_name || `#${detail.worker_id}` }}</el-descriptions-item>
        <el-descriptions-item label="金额">¥{{ Number(detail.amount || 0).toFixed(2) }}</el-descriptions-item>
        <el-descriptions-item label="手续费">¥{{ Number(detail.fee || 0).toFixed(2) }}</el-descriptions-item>
        <el-descriptions-item label="实发金额">¥{{ Number(detail.net_amount || 0).toFixed(2) }}</el-descriptions-item>
        <el-descriptions-item label="支付方式">{{ payMethodText(detail.pay_method) }}</el-descriptions-item>
        <el-descriptions-item label="状态">
          <el-tag :type="statusTagType(detail.status)" size="small">{{ statusMap[detail.status] || '-' }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item v-if="detail.third_party_no" label="三方单号">{{ detail.third_party_no }}</el-descriptions-item>
        <el-descriptions-item v-if="detail.remark" label="备注" :span="2">{{ detail.remark }}</el-descriptions-item>
        <el-descriptions-item v-if="detail.paid_at" label="支付时间">{{ formatTime(detail.paid_at) }}</el-descriptions-item>
        <el-descriptions-item v-if="detail.settled_at" label="结算时间">{{ formatTime(detail.settled_at) }}</el-descriptions-item>
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
import { adminListLinggongPayments, getLinggongPayment, settleLinggongPayment } from '@/api/linggong'
import { formatTime } from '@/utils/format'

const loading = ref(false)
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)
const list = ref([])

const filters = reactive({ keyword: '', status: null, dateRange: null })

const statusMap = {
  0: '待支付', 1: '已支付', 2: '结算中',
  3: '已结算', 4: '已退款', 5: '已失败'
}
const statusTagType = (s) => ({
  0: 'info', 1: 'success', 2: 'warning',
  3: 'success', 4: 'primary', 5: 'danger'
}[s] || 'info')

const payMethodText = (m) => ({
  wechat: '微信', alipay: '支付宝', bank: '银行卡',
  balance: '余额', offline: '线下'
}[m] || m || '-')

const formatAmount = (n) => Number(n || 0).toFixed(2)

const stats = computed(() => {
  const total = list.value.length
  let totalAmount = 0
  let pendingAmount = 0
  let settledAmount = 0
  list.value.forEach((r) => {
    totalAmount += Number(r.amount || 0)
    if (r.status === 1 || r.status === 2) pendingAmount += Number(r.amount || 0)
    if (r.status === 3) settledAmount += Number(r.amount || 0)
  })
  return { total, totalAmount, pendingAmount, settledAmount }
})

const onSearch = () => { page.value = 1; loadList() }
const onReset = () => {
  filters.keyword = ''
  filters.status = null
  filters.dateRange = null
  page.value = 1
  loadList()
}

const loadList = async () => {
  loading.value = true
  try {
    const params = {
      page: page.value,
      page_size: pageSize.value,
      keyword: filters.keyword || undefined,
      status: filters.status === null || filters.status === '' ? undefined : filters.status
    }
    if (filters.dateRange && filters.dateRange.length === 2) {
      params.start_date = filters.dateRange[0]
      params.end_date = filters.dateRange[1]
    }
    const res = await adminListLinggongPayments(params)
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
    const res = await getLinggongPayment(row.id)
    detail.value = res.data || row
  } catch (e) {
    detail.value = row
  }
  detailVisible.value = true
}

const onSettle = async (row) => {
  try {
    await ElMessageBox.confirm(`确认结算该笔支付 ¥${Number(row.amount || 0).toFixed(2)} 吗？`, '结算', { type: 'warning' })
    await settleLinggongPayment(row.id, {})
    ElMessage.success('结算成功')
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
.price { color: #f56c6c; font-weight: 600; }
.filter-form { margin-bottom: 12px; padding: 12px 16px; background: #fafafa; border-radius: 4px; }
.filter-form :deep(.el-form-item) { margin-bottom: 8px; margin-right: 12px; }
.toolbar { display: flex; justify-content: flex-start; margin-bottom: 12px; }
.pagination-wrap { margin-top: 16px; display: flex; justify-content: flex-end; }
</style>
