<template>
  <div class="app-container">
    <!-- 支付统计 -->
    <el-row :gutter="16" class="stat-row">
      <el-col :xs="12" :sm="8" :md="6">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value">{{ stats.total_count }}</div><div class="stat-label">支付笔数</div></div></el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="6">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value text-success">¥{{ Number(stats.total_amount || 0).toFixed(2) }}</div><div class="stat-label">支付总额</div></div></el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="6">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value text-warning">¥{{ Number(stats.pending_amount || 0).toFixed(2) }}</div><div class="stat-label">待支付</div></div></el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="6">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value text-danger">¥{{ Number(stats.refund_amount || 0).toFixed(2) }}</div><div class="stat-label">已退款</div></div></el-card>
      </el-col>
    </el-row>

    <div class="page-card">
      <el-form :inline="true" :model="filters" class="filter-form" @submit.prevent>
        <el-form-item label="支付单号">
          <el-input v-model="filters.payment_no" placeholder="支付单号" clearable style="width: 200px" :prefix-icon="Search" @keyup.enter="onSearch" />
        </el-form-item>
        <el-form-item label="订单号">
          <el-input v-model="filters.order_no" placeholder="订单号" clearable style="width: 200px" @keyup.enter="onSearch" />
        </el-form-item>
        <el-form-item label="店铺ID">
          <el-input-number v-model="filters.shop_id" :controls="false" placeholder="店铺ID" style="width: 140px" />
        </el-form-item>
        <el-form-item label="支付状态">
          <el-select v-model="filters.status" placeholder="全部" clearable style="width: 120px" @change="onSearch">
            <el-option label="待支付" :value="0" />
            <el-option label="已支付" :value="1" />
            <el-option label="已关闭" :value="2" />
            <el-option label="已退款" :value="3" />
          </el-select>
        </el-form-item>
        <el-form-item label="支付方式">
          <el-select v-model="filters.pay_method" placeholder="全部" clearable style="width: 120px" @change="onSearch">
            <el-option label="微信" value="wechat" />
            <el-option label="支付宝" value="alipay" />
            <el-option label="余额" value="balance" />
          </el-select>
        </el-form-item>
        <el-form-item label="时间范围">
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

      <el-table v-loading="loading" :data="list" border stripe @sort-change="onSortChange">
        <el-table-column prop="id" label="ID" width="70" sortable="custom" fixed="left" />
        <el-table-column label="支付单号" min-width="200">
          <template #default="{ row }">
            <el-link type="primary" :underline="false" @click="openDetail(row)">{{ row.payment_no }}</el-link>
          </template>
        </el-table-column>
        <el-table-column label="订单号" width="180">
          <template #default="{ row }">{{ row.order_no || `#${row.order_id}` }}</template>
        </el-table-column>
        <el-table-column label="店铺ID" width="90" prop="shop_id" />
        <el-table-column label="用户ID" width="90" prop="user_id" />
        <el-table-column label="支付金额" width="120">
          <template #default="{ row }">¥{{ Number(row.amount || 0).toFixed(2) }}</template>
        </el-table-column>
        <el-table-column label="支付方式" width="100">
          <template #default="{ row }">
            <el-tag size="small">{{ payMethodText(row.pay_method) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="statusTagType(row.status)" size="small">{{ statusText(row.status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="交易号" width="180" show-overflow-tooltip>
          <template #default="{ row }">{{ row.transaction_id || '-' }}</template>
        </el-table-column>
        <el-table-column label="支付时间" width="160">
          <template #default="{ row }">{{ formatTime(row.paid_at) }}</template>
        </el-table-column>
        <el-table-column label="创建时间" width="160" prop="created_at" sortable="custom">
          <template #default="{ row }">{{ formatTime(row.created_at) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="160" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="openDetail(row)">详情</el-button>
            <el-button v-if="row.status === 1" type="warning" link size="small" @click="openRefund(row)">触发退款</el-button>
          </template>
        </el-table-column>
      </el-table>

      <div class="pagination-wrap">
        <el-pagination v-model:current-page="page" v-model:page-size="pageSize" :total="total" :page-sizes="[10, 20, 50, 100]" layout="total, sizes, prev, pager, next, jumper" background @current-change="loadList" @size-change="loadList" />
      </div>
    </div>

    <!-- 详情弹窗 -->
    <el-dialog v-model="detailVisible" title="支付详情" width="720px" destroy-on-close>
      <el-descriptions v-if="detail" :column="2" border>
        <el-descriptions-item label="ID">{{ detail.id }}</el-descriptions-item>
        <el-descriptions-item label="支付单号">{{ detail.payment_no }}</el-descriptions-item>
        <el-descriptions-item label="订单号">{{ detail.order_no || `#${detail.order_id}` }}</el-descriptions-item>
        <el-descriptions-item label="店铺ID">{{ detail.shop_id }}</el-descriptions-item>
        <el-descriptions-item label="用户ID">{{ detail.user_id }}</el-descriptions-item>
        <el-descriptions-item label="支付方式">{{ payMethodText(detail.pay_method) }}</el-descriptions-item>
        <el-descriptions-item label="支付金额">¥{{ Number(detail.amount || 0).toFixed(2) }}</el-descriptions-item>
        <el-descriptions-item label="已退款金额">¥{{ Number(detail.refund_amount || 0).toFixed(2) }}</el-descriptions-item>
        <el-descriptions-item label="状态">
          <el-tag :type="statusTagType(detail.status)" size="small">{{ statusText(detail.status) }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="交易号">{{ detail.transaction_id || '-' }}</el-descriptions-item>
        <el-descriptions-item label="支付时间">{{ formatTime(detail.paid_at) }}</el-descriptions-item>
        <el-descriptions-item label="关闭时间">{{ formatTime(detail.closed_at) }}</el-descriptions-item>
        <el-descriptions-item label="创建时间">{{ formatTime(detail.created_at) }}</el-descriptions-item>
        <el-descriptions-item label="更新时间">{{ formatTime(detail.updated_at) }}</el-descriptions-item>
        <el-descriptions-item label="第三方回执" :span="2">{{ detail.callback_data || '-' }}</el-descriptions-item>
      </el-descriptions>
    </el-dialog>

    <!-- 触发退款弹窗 -->
    <el-dialog v-model="refundVisible" title="触发退款" width="480px" destroy-on-close>
      <el-form :model="refundForm" label-width="100px">
        <el-form-item label="支付单号">{{ refundForm.payment_no }}</el-form-item>
        <el-form-item label="支付金额">¥{{ Number(refundForm.amount || 0).toFixed(2) }}</el-form-item>
        <el-form-item label="退款原因">
          <el-input v-model="refundForm.reason" type="textarea" :rows="3" placeholder="退款原因" />
        </el-form-item>
        <el-form-item>
          <el-alert type="info" :closable="false" title="提示：触发退款将跳转到退款管理页面创建退款单" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="refundVisible = false">取消</el-button>
        <el-button type="primary" @click="goToRefund">前往退款管理</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Refresh, RefreshLeft, Search } from '@element-plus/icons-vue'
import { formatTime } from '@/utils/format'
import { getMallPaymentList, getMallPaymentDetail, getMallPaymentStats } from '@/api/mall'

const router = useRouter()

const loading = ref(false)
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)
const list = ref([])
const sortField = ref('created_at')
const sortOrder = ref('descending')

const stats = reactive({ total_count: 0, total_amount: 0, pending_amount: 0, refund_amount: 0 })

const filters = reactive({ payment_no: '', order_no: '', shop_id: null, status: null, pay_method: '', dateRange: null })

const onSearch = () => { page.value = 1; loadList() }
const onReset = () => {
  filters.payment_no = ''
  filters.order_no = ''
  filters.shop_id = null
  filters.status = null
  filters.pay_method = ''
  filters.dateRange = null
  page.value = 1
  loadList()
}
const onSortChange = ({ prop, order }) => {
  sortField.value = prop || 'created_at'
  sortOrder.value = order || 'descending'
  loadList()
}

const statusText = (s) => ({ 0: '待支付', 1: '已支付', 2: '已关闭', 3: '已退款' }[s] || '-')
const statusTagType = (s) => ({ 0: 'warning', 1: 'success', 2: 'info', 3: 'danger' }[s] || 'info')
const payMethodText = (m) => ({ wechat: '微信', alipay: '支付宝', balance: '余额' }[m] || m || '-')

const loadList = async () => {
  loading.value = true
  try {
    const params = {
      page: page.value,
      page_size: pageSize.value,
      payment_no: filters.payment_no.trim() || undefined,
      order_no: filters.order_no.trim() || undefined,
      shop_id: filters.shop_id || undefined,
      status: filters.status === null || filters.status === '' ? undefined : filters.status,
      pay_method: filters.pay_method || undefined,
      sort: sortField.value,
      order: sortOrder.value
    }
    if (filters.dateRange && filters.dateRange.length === 2) {
      params.start_date = filters.dateRange[0]
      params.end_date = filters.dateRange[1]
    }
    const res = await getMallPaymentList(params)
    const data = res.data || {}
    list.value = data.list || []
    total.value = data.total || 0
  } catch (e) {
    ElMessage.error('加载支付列表失败')
    list.value = []
    total.value = 0
  } finally {
    loading.value = false
  }
}

const loadStats = async () => {
  try {
    const params = {}
    if (filters.dateRange && filters.dateRange.length === 2) {
      params.start_date = filters.dateRange[0]
      params.end_date = filters.dateRange[1]
    }
    const res = await getMallPaymentStats(params)
    Object.assign(stats, res.data || {})
  } catch (e) { /* ignore */ }
}

const detailVisible = ref(false)
const detail = ref(null)

const openDetail = async (row) => {
  try {
    const res = await getMallPaymentDetail(row.id)
    detail.value = res.data || null
    detailVisible.value = true
  } catch (e) {
    ElMessage.error('加载详情失败')
  }
}

const refundVisible = ref(false)
const refundForm = reactive({ payment_no: '', amount: 0, reason: '' })

const openRefund = (row) => {
  Object.assign(refundForm, { payment_no: row.payment_no, amount: row.amount, reason: '' })
  refundVisible.value = true
}

const goToRefund = () => {
  refundVisible.value = false
  router.push('/business/mall/refunds')
}

onMounted(async () => {
  await Promise.all([loadList(), loadStats()])
})
</script>

<style scoped>
.stat-row { margin-bottom: 16px; }
.stat-item { text-align: center; padding: 16px; }
.stat-value { font-size: 20px; font-weight: 600; color: #303133; }
.text-success { color: #67c23a; }
.text-warning { color: #e6a23c; }
.text-danger { color: #f56c6c; }
.filter-form { margin-bottom: 12px; padding: 12px 16px; background: #fafafa; border-radius: 4px; }
.filter-form :deep(.el-form-item) { margin-bottom: 8px; margin-right: 12px; }
.toolbar { display: flex; justify-content: space-between; margin-bottom: 12px; }
.pagination-wrap { margin-top: 16px; display: flex; justify-content: flex-end; }
</style>
