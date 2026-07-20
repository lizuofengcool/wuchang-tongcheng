<template>
  <div class="app-container">
    <!-- 顶部统计卡片 -->
    <el-row :gutter="16" class="stat-row">
      <el-col :xs="12" :sm="8" :md="4">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-icon" style="background: #409eff">
            <el-icon :size="22"><Wallet /></el-icon>
          </div>
          <div class="stat-content">
            <div class="stat-value">{{ stats.total }}</div>
            <div class="stat-label">总支付数</div>
          </div>
        </el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="4">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-icon" style="background: #67c23a">
            <el-icon :size="22"><Money /></el-icon>
          </div>
          <div class="stat-content">
            <div class="stat-value">¥{{ stats.totalAmount.toFixed(2) }}</div>
            <div class="stat-label">总金额</div>
          </div>
        </el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="4">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-icon" style="background: #e6a23c">
            <el-icon :size="22"><Clock /></el-icon>
          </div>
          <div class="stat-content">
            <div class="stat-value">{{ stats.pending }}</div>
            <div class="stat-label">待支付</div>
          </div>
        </el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="4">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-icon" style="background: #13c2c2">
            <el-icon :size="22"><CircleCheck /></el-icon>
          </div>
          <div class="stat-content">
            <div class="stat-value">{{ stats.paid }}</div>
            <div class="stat-label">已支付</div>
          </div>
        </el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="4">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-icon" style="background: #722ed1">
            <el-icon :size="22"><Refresh /></el-icon>
          </div>
          <div class="stat-content">
            <div class="stat-value">{{ stats.refunded }}</div>
            <div class="stat-label">已退款</div>
          </div>
        </el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="4">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-icon" style="background: #f56c6c">
            <el-icon :size="22"><CircleClose /></el-icon>
          </div>
          <div class="stat-content">
            <div class="stat-value">{{ stats.failed }}</div>
            <div class="stat-label">支付失败</div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <div class="page-card">
      <!-- 筛选区 -->
      <el-form :inline="true" :model="filters" class="filter-form" @submit.prevent>
        <el-form-item label="支付单号">
          <el-input v-model="filters.payment_no" placeholder="支付单号" clearable style="width: 180px" @keyup.enter="onSearch" />
        </el-form-item>
        <el-form-item label="预订ID">
          <el-input v-model="filters.booking_id" placeholder="预订ID" clearable style="width: 120px" @keyup.enter="onSearch" />
        </el-form-item>
        <el-form-item label="支付方式">
          <el-select v-model="filters.payment_method" placeholder="全部" clearable style="width: 130px" @change="onSearch">
            <el-option label="微信" value="wechat" />
            <el-option label="支付宝" value="alipay" />
            <el-option label="银行卡" value="bank" />
            <el-option label="余额" value="balance" />
            <el-option label="ETC" value="etc" />
          </el-select>
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="filters.status" placeholder="全部" clearable style="width: 130px" @change="onSearch">
            <el-option label="待支付" :value="0" />
            <el-option label="已支付" :value="1" />
            <el-option label="已退款" :value="2" />
            <el-option label="已取消" :value="3" />
            <el-option label="支付失败" :value="4" />
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

      <!-- 工具栏 -->
      <div class="toolbar">
        <div class="toolbar-left">
          <el-button :icon="Refresh" @click="loadList">刷新</el-button>
        </div>
        <div class="toolbar-right">
          <el-button type="primary" :icon="Plus" @click="onCreate">新建支付</el-button>
        </div>
      </div>

      <!-- 表格 -->
      <el-table v-loading="loading" :data="list" border stripe>
        <el-table-column prop="id" label="ID" width="70" fixed="left" />
        <el-table-column prop="payment_no" label="支付单号" width="180" />
        <el-table-column prop="booking_id" label="预订ID" width="90" />
        <el-table-column label="金额" width="120">
          <template #default="{ row }">
            <span class="price">¥{{ Number(row.amount || 0).toFixed(2) }}</span>
          </template>
        </el-table-column>
        <el-table-column label="支付方式" width="110">
          <template #default="{ row }">
            <el-tag size="small">{{ methodText(row.payment_method) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="付款方" width="140">
          <template #default="{ row }">
            <div>{{ row.payer_name || `#${row.payer_id}` }}</div>
          </template>
        </el-table-column>
        <el-table-column label="收款方" width="140">
          <template #default="{ row }">
            <div>{{ row.payee_name || `#${row.payee_id}` }}</div>
          </template>
        </el-table-column>
        <el-table-column label="支付时间" width="160">
          <template #default="{ row }">{{ formatTime(row.paid_at) }}</template>
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
            <el-button v-if="row.status === 0" type="success" link size="small" @click="onMarkPaid(row)">标记已付</el-button>
            <el-button v-if="row.status === 1" type="warning" link size="small" @click="onRefund(row)">退款</el-button>
          </template>
        </el-table-column>
      </el-table>

      <!-- 分页 -->
      <div class="pagination-wrap">
        <el-pagination
          v-model:current-page="page"
          v-model:page-size="pageSize"
          :total="total"
          :page-sizes="[10, 20, 50, 100]"
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
        <el-descriptions-item label="支付单号">{{ detail.payment_no }}</el-descriptions-item>
        <el-descriptions-item label="预订ID">{{ detail.booking_id }}</el-descriptions-item>
        <el-descriptions-item label="金额">
          <span class="price">¥{{ Number(detail.amount || 0).toFixed(2) }}</span>
        </el-descriptions-item>
        <el-descriptions-item label="支付方式">
          <el-tag size="small">{{ methodText(detail.payment_method) }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="状态">
          <el-tag :type="statusTagType(detail.status)" size="small">{{ statusText(detail.status) }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="付款方">{{ detail.payer_name || `#${detail.payer_id}` }}</el-descriptions-item>
        <el-descriptions-item label="收款方">{{ detail.payee_name || `#${detail.payee_id}` }}</el-descriptions-item>
        <el-descriptions-item label="第三方流水号" :span="2">{{ detail.transaction_id || '-' }}</el-descriptions-item>
        <el-descriptions-item label="支付时间">{{ formatTime(detail.paid_at) }}</el-descriptions-item>
        <el-descriptions-item label="退款时间">{{ formatTime(detail.refunded_at) }}</el-descriptions-item>
        <el-descriptions-item v-if="detail.refund_amount" label="退款金额">¥{{ Number(detail.refund_amount).toFixed(2) }}</el-descriptions-item>
        <el-descriptions-item v-if="detail.refund_reason" label="退款原因" :span="2">{{ detail.refund_reason }}</el-descriptions-item>
        <el-descriptions-item v-if="detail.remark" label="备注" :span="2">{{ detail.remark }}</el-descriptions-item>
        <el-descriptions-item label="创建时间">{{ formatTime(detail.created_at) }}</el-descriptions-item>
        <el-descriptions-item label="更新时间">{{ formatTime(detail.updated_at) }}</el-descriptions-item>
      </el-descriptions>
      <template #footer><el-button @click="detailVisible = false">关闭</el-button></template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import request from '@/utils/request'
import {
  Refresh, RefreshLeft, Search, Plus,
  Wallet, Money, Clock, CircleCheck, CircleClose
} from '@element-plus/icons-vue'
import { formatTime } from '@/utils/format'

const loading = ref(false)
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)
const list = ref([])

const stats = reactive({
  total: 0, totalAmount: 0, pending: 0, paid: 0, refunded: 0, failed: 0
})

const filters = reactive({
  payment_no: '', booking_id: '', payment_method: '',
  status: null, dateRange: null
})

const statusText = (s) => ({ 0: '待支付', 1: '已支付', 2: '已退款', 3: '已取消', 4: '支付失败' }[s] || '-')
const statusTagType = (s) => ({ 0: 'warning', 1: 'success', 2: 'info', 3: 'info', 4: 'danger' }[s] || 'info')
const methodText = (m) => ({ wechat: '微信', alipay: '支付宝', bank: '银行卡', balance: '余额', etc: 'ETC' }[m] || m || '-')

const onSearch = () => { page.value = 1; loadList() }
const onReset = () => {
  filters.payment_no = ''; filters.booking_id = ''; filters.payment_method = ''
  filters.status = null; filters.dateRange = null
  page.value = 1; loadList()
}

const loadList = async () => {
  loading.value = true
  try {
    const params = {
      page: page.value,
      page_size: pageSize.value,
      payment_no: filters.payment_no || undefined,
      booking_id: filters.booking_id || undefined,
      payment_method: filters.payment_method || undefined,
      status: filters.status === null || filters.status === '' ? undefined : filters.status
    }
    if (filters.dateRange && filters.dateRange.length === 2) {
      params.start_date = filters.dateRange[0]
      params.end_date = filters.dateRange[1]
    }
    const res = await request.get('/pinche/admin/payments', { params })
    const data = res.data || {}
    list.value = data.list || []
    total.value = data.total || 0
  } catch (e) {
    list.value = []; total.value = 0
  } finally {
    loading.value = false
  }
}

const detailVisible = ref(false)
const detail = ref(null)
const openDetail = async (row) => {
  detail.value = row
  detailVisible.value = true
  try {
    const res = await request.get(`/pinche/payments/${row.id}`)
    if (res.data) detail.value = res.data
  } catch (e) { /* ignore */ }
}

const onMarkPaid = async (row) => {
  try {
    await ElMessageBox.confirm(`确认将支付 "${row.payment_no}" 标记为已支付？`, '提示', { type: 'warning' })
    await request.put(`/pinche/admin/payments/${row.id}/status`, { status: 1 })
    ElMessage.success('已标记为已支付')
    await loadList()
  } catch (e) { /* cancel */ }
}

const onRefund = async (row) => {
  try {
    const { value } = await ElMessageBox.prompt('请输入退款原因', '退款', {
      inputType: 'textarea',
      inputPlaceholder: '退款原因',
      inputValidator: (v) => !!v || '请输入退款原因'
    })
    await request.put(`/pinche/admin/payments/${row.id}/status`, { status: 2, refund_reason: value })
    ElMessage.success('已退款')
    await loadList()
  } catch (e) { /* cancel */ }
}

const onCreate = () => {
  ElMessage.info('新建支付功能开发中，请使用 C 端发起')
}

onMounted(() => loadList())
</script>

<style scoped>
.stat-row { margin-bottom: 16px; }
.stat-card { display: flex; align-items: center; }
.stat-card :deep(.el-card__body) {
  display: flex; align-items: center; gap: 14px; padding: 16px; width: 100%;
}
.stat-icon {
  width: 44px; height: 44px; border-radius: 8px;
  display: flex; align-items: center; justify-content: center;
  color: #fff; flex-shrink: 0;
}
.stat-content { flex: 1; min-width: 0; }
.stat-value { font-size: 22px; font-weight: 600; color: #303133; }
.stat-label { font-size: 12px; color: #909399; margin-top: 2px; }

.filter-form { margin-bottom: 12px; padding: 12px 16px; background: #fafafa; border-radius: 4px; }
.filter-form :deep(.el-form-item) { margin-bottom: 8px; margin-right: 12px; }

.toolbar { display: flex; justify-content: space-between; align-items: center; flex-wrap: wrap; gap: 8px; margin-bottom: 12px; }
.toolbar-left { display: flex; gap: 8px; flex-wrap: wrap; }
.toolbar-right { display: flex; gap: 8px; }

.price { color: #f56c6c; font-weight: 600; }
.pagination-wrap { margin-top: 16px; display: flex; justify-content: flex-end; }
</style>
