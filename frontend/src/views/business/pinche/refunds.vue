<template>
  <div class="app-container">
    <!-- 顶部统计卡片 -->
    <el-row :gutter="16" class="stat-row">
      <el-col :xs="12" :sm="8" :md="4">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-icon" style="background: #409eff">
            <el-icon :size="22"><RefreshLeft /></el-icon>
          </div>
          <div class="stat-content">
            <div class="stat-value">{{ stats.total }}</div>
            <div class="stat-label">总退款数</div>
          </div>
        </el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="4">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-icon" style="background: #f56c6c">
            <el-icon :size="22"><Money /></el-icon>
          </div>
          <div class="stat-content">
            <div class="stat-value">¥{{ stats.totalAmount.toFixed(2) }}</div>
            <div class="stat-label">退款总额</div>
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
            <div class="stat-label">待处理</div>
          </div>
        </el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="4">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-icon" style="background: #722ed1">
            <el-icon :size="22"><Loading /></el-icon>
          </div>
          <div class="stat-content">
            <div class="stat-value">{{ stats.processing }}</div>
            <div class="stat-label">处理中</div>
          </div>
        </el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="4">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-icon" style="background: #67c23a">
            <el-icon :size="22"><CircleCheckFilled /></el-icon>
          </div>
          <div class="stat-content">
            <div class="stat-value">{{ stats.success }}</div>
            <div class="stat-label">已退款</div>
          </div>
        </el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="4">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-icon" style="background: #909399">
            <el-icon :size="22"><CircleCloseFilled /></el-icon>
          </div>
          <div class="stat-content">
            <div class="stat-value">{{ stats.rejected }}</div>
            <div class="stat-label">已驳回</div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <div class="page-card">
      <!-- 筛选区 -->
      <el-form :inline="true" :model="filters" class="filter-form" @submit.prevent>
        <el-form-item label="退款单号">
          <el-input
            v-model="filters.refund_no"
            placeholder="退款单号"
            clearable
            style="width: 180px"
            @keyup.enter="onSearch"
          />
        </el-form-item>
        <el-form-item label="订单ID">
          <el-input v-model="filters.order_id" placeholder="订单 ID" clearable style="width: 130px" @keyup.enter="onSearch" />
        </el-form-item>
        <el-form-item label="预订ID">
          <el-input v-model="filters.booking_id" placeholder="预订 ID" clearable style="width: 130px" @keyup.enter="onSearch" />
        </el-form-item>
        <el-form-item label="退款类型">
          <el-select v-model="filters.refund_type" placeholder="全部" clearable style="width: 140px" @change="onSearch">
            <el-option label="取消预订" value="cancel_booking" />
            <el-option label="行程取消" value="trip_cancel" />
            <el-option label="车主违约" value="driver_breach" />
            <el-option label="乘客违约" value="passenger_breach" />
            <el-option label="平台介入" value="platform_intervene" />
            <el-option label="其他" value="other" />
          </el-select>
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="filters.status" placeholder="全部" clearable style="width: 120px" @change="onSearch">
            <el-option label="待处理" :value="0" />
            <el-option label="处理中" :value="1" />
            <el-option label="已退款" :value="2" />
            <el-option label="已驳回" :value="3" />
            <el-option label="已撤销" :value="4" />
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
          <el-button type="success" :icon="Check" :disabled="!selection.length" @click="onBatchApprove">批量批准</el-button>
          <el-button type="warning" :icon="Close" :disabled="!selection.length" @click="onBatchReject">批量驳回</el-button>
        </div>
        <div class="toolbar-right">
          <el-button type="primary" :icon="Download" @click="onExport">导出</el-button>
        </div>
      </div>

      <!-- 表格 -->
      <el-table
        v-loading="loading"
        :data="list"
        border
        stripe
        @selection-change="onSelectionChange"
      >
        <el-table-column type="selection" width="44" fixed="left" />
        <el-table-column prop="id" label="ID" width="70" fixed="left" />
        <el-table-column prop="refund_no" label="退款单号" width="180" />
        <el-table-column prop="order_id" label="订单ID" width="90" />
        <el-table-column prop="booking_id" label="预订ID" width="90" />
        <el-table-column label="退款类型" width="120">
          <template #default="{ row }">
            <el-tag :type="typeTagType(row.refund_type)" size="small">{{ typeText(row.refund_type) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="退款金额" width="120">
          <template #default="{ row }">
            <span class="price">¥{{ Number(row.refund_amount || 0).toFixed(2) }}</span>
          </template>
        </el-table-column>
        <el-table-column label="申请人" width="140">
          <template #default="{ row }">
            <div>{{ row.applicant_name || `用户#${row.applicant_id}` }}</div>
          </template>
        </el-table-column>
        <el-table-column label="退款原因" min-width="220">
          <template #default="{ row }">
            <div class="reason-text">{{ row.reason }}</div>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="statusTagType(row.status)" size="small">{{ statusText(row.status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="申请时间" width="160" prop="created_at" sortable>
          <template #default="{ row }">{{ formatTime(row.created_at) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="220" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="openDetail(row)">详情</el-button>
            <el-button
              v-if="row.status === 0 || row.status === 1"
              type="success"
              link
              size="small"
              @click="onApprove(row)"
            >批准</el-button>
            <el-button
              v-if="row.status === 0 || row.status === 1"
              type="danger"
              link
              size="small"
              @click="onReject(row)"
            >驳回</el-button>
            <el-button
              v-if="row.status === 2"
              type="warning"
              link
              size="small"
              @click="onMarkPaid(row)"
            >标记已打款</el-button>
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
    <el-dialog v-model="detailVisible" title="退款详情" width="800px" destroy-on-close>
      <el-descriptions v-if="detail" :column="2" border>
        <el-descriptions-item label="ID">{{ detail.id }}</el-descriptions-item>
        <el-descriptions-item label="退款单号">{{ detail.refund_no }}</el-descriptions-item>
        <el-descriptions-item label="订单ID">{{ detail.order_id }}</el-descriptions-item>
        <el-descriptions-item label="预订ID">{{ detail.booking_id || '-' }}</el-descriptions-item>
        <el-descriptions-item label="行程ID">{{ detail.trip_id || '-' }}</el-descriptions-item>
        <el-descriptions-item label="拼车ID">{{ detail.pinche_id || '-' }}</el-descriptions-item>
        <el-descriptions-item label="退款类型">
          <el-tag :type="typeTagType(detail.refund_type)" size="small">{{ typeText(detail.refund_type) }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="退款金额">
          <span class="price">¥{{ Number(detail.refund_amount || 0).toFixed(2) }}</span>
        </el-descriptions-item>
        <el-descriptions-item label="原支付金额">¥{{ Number(detail.original_amount || 0).toFixed(2) }}</el-descriptions-item>
        <el-descriptions-item label="退款方式">{{ refundMethodText(detail.refund_method) }}</el-descriptions-item>
        <el-descriptions-item label="申请人">{{ detail.applicant_name || `用户#${detail.applicant_id}` }}</el-descriptions-item>
        <el-descriptions-item label="联系电话">{{ detail.applicant_phone || '-' }}</el-descriptions-item>
        <el-descriptions-item label="收款账户">{{ detail.payee_account || '-' }}</el-descriptions-item>
        <el-descriptions-item label="状态">
          <el-tag :type="statusTagType(detail.status)" size="small">{{ statusText(detail.status) }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="申请时间">{{ formatTime(detail.created_at) }}</el-descriptions-item>
        <el-descriptions-item v-if="detail.handled_at" label="处理时间">{{ formatTime(detail.handled_at) }}</el-descriptions-item>
        <el-descriptions-item v-if="detail.paid_at" label="打款时间">{{ formatTime(detail.paid_at) }}</el-descriptions-item>
        <el-descriptions-item v-if="detail.handler_id" label="处理人">管理员 #{{ detail.handler_id }}</el-descriptions-item>
        <el-descriptions-item label="退款原因" :span="2">{{ detail.reason }}</el-descriptions-item>
        <el-descriptions-item v-if="detail.handle_note" label="处理记录" :span="2">{{ detail.handle_note }}</el-descriptions-item>
        <el-descriptions-item v-if="detail.reject_reason" label="驳回原因" :span="2">{{ detail.reject_reason }}</el-descriptions-item>
      </el-descriptions>
      <template #footer>
        <el-button @click="detailVisible = false">关闭</el-button>
        <el-button v-if="detail && (detail.status === 0 || detail.status === 1)" type="success" @click="onApprove(detail)">立即批准</el-button>
      </template>
    </el-dialog>

    <!-- 批准弹窗 -->
    <el-dialog v-model="approveVisible" title="批准退款" width="600px" destroy-on-close>
      <el-form :model="approveForm" label-width="110px">
        <el-form-item label="退款单号">
          <span>{{ approveForm.refund_no }}</span>
        </el-form-item>
        <el-form-item label="退款金额">
          <span class="price">¥{{ Number(approveForm.refund_amount || 0).toFixed(2) }}</span>
        </el-form-item>
        <el-form-item label="退款方式">
          <el-select v-model="approveForm.refund_method" placeholder="请选择" style="width: 100%">
            <el-option label="原路退回" value="original" />
            <el-option label="余额退款" value="balance" />
            <el-option label="银行转账" value="bank" />
            <el-option label="线下退款" value="offline" />
          </el-select>
        </el-form-item>
        <el-form-item label="实际退款金额">
          <el-input-number
            v-model="approveForm.actual_amount"
            :min="0"
            :precision="2"
            controls-position="right"
            style="width: 100%"
          />
        </el-form-item>
        <el-form-item label="处理说明">
          <el-input
            v-model="approveForm.handle_note"
            type="textarea"
            :rows="4"
            placeholder="请填写处理说明"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="approveVisible = false">取消</el-button>
        <el-button type="primary" @click="submitApprove">确认批准</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import request from '@/utils/request'
import {
  Refresh, RefreshLeft, Search, Refresh as RefreshIcon, Check, Close, Download,
  Money, Clock, Loading, CircleCheckFilled, CircleCloseFilled
} from '@element-plus/icons-vue'
import { formatTime } from '@/utils/format'

const loading = ref(false)
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)
const list = ref([])
const selection = ref([])

const stats = reactive({
  total: 0, totalAmount: 0, pending: 0, processing: 0, success: 0, rejected: 0
})

const filters = reactive({
  refund_no: '', order_id: '', booking_id: '',
  refund_type: '', status: null, dateRange: null
})

const statusText = (s) => ({ 0: '待处理', 1: '处理中', 2: '已退款', 3: '已驳回', 4: '已撤销' }[s] || '-')
const statusTagType = (s) => ({ 0: 'warning', 1: '', 2: 'success', 3: 'danger', 4: 'info' }[s] || 'info')
const typeText = (t) => ({
  cancel_booking: '取消预订', trip_cancel: '行程取消', driver_breach: '车主违约',
  passenger_breach: '乘客违约', platform_intervene: '平台介入', other: '其他'
}[t] || '-')
const typeTagType = (t) => ({
  cancel_booking: 'info', trip_cancel: 'warning', driver_breach: 'danger',
  passenger_breach: 'warning', platform_intervene: 'primary', other: 'info'
}[t] || 'info')
const refundMethodText = (m) => ({
  original: '原路退回', balance: '余额退款', bank: '银行转账', offline: '线下退款'
}[m] || '-')

const onSearch = () => { page.value = 1; loadList() }
const onReset = () => {
  filters.refund_no = ''; filters.order_id = ''; filters.booking_id = ''
  filters.refund_type = ''; filters.status = null; filters.dateRange = null
  page.value = 1; loadList()
}

const onSelectionChange = (rows) => { selection.value = rows }

const loadList = async () => {
  loading.value = true
  try {
    const params = {
      page: page.value,
      page_size: pageSize.value,
      refund_no: filters.refund_no || undefined,
      order_id: filters.order_id || undefined,
      booking_id: filters.booking_id || undefined,
      refund_type: filters.refund_type || undefined,
      status: filters.status === null || filters.status === '' ? undefined : filters.status
    }
    if (filters.dateRange && filters.dateRange.length === 2) {
      params.start_date = filters.dateRange[0]
      params.end_date = filters.dateRange[1]
    }
    const res = await request.get('/pinche/admin/refunds', { params })
    const data = res.data || {}
    list.value = data.list || []
    total.value = data.total || 0
    if (data.stats) {
      stats.total = data.stats.total || 0
      stats.totalAmount = data.stats.total_amount || 0
      stats.pending = data.stats.pending || 0
      stats.processing = data.stats.processing || 0
      stats.success = data.stats.success || 0
      stats.rejected = data.stats.rejected || 0
    } else {
      stats.total = total.value
      stats.totalAmount = list.value.reduce((s, i) => s + (i.refund_amount || 0), 0)
      stats.pending = list.value.filter(i => i.status === 0).length
      stats.processing = list.value.filter(i => i.status === 1).length
      stats.success = list.value.filter(i => i.status === 2).length
      stats.rejected = list.value.filter(i => i.status === 3).length
    }
  } catch (e) {
    list.value = []; total.value = 0
  } finally {
    loading.value = false
  }
}

const detailVisible = ref(false)
const detail = ref(null)
const openDetail = async (row) => {
  detail.value = { ...row }
  detailVisible.value = true
}

const approveVisible = ref(false)
const approveForm = reactive({
  id: 0, refund_no: '', refund_amount: 0,
  refund_method: 'original', actual_amount: 0, handle_note: ''
})

const onApprove = (row) => {
  approveForm.id = row.id
  approveForm.refund_no = row.refund_no || `#${row.id}`
  approveForm.refund_amount = row.refund_amount || 0
  approveForm.refund_method = row.refund_method || 'original'
  approveForm.actual_amount = row.refund_amount || 0
  approveForm.handle_note = ''
  approveVisible.value = true
}

const submitApprove = async () => {
  if (approveForm.actual_amount <= 0) {
    ElMessage.warning('请输入有效的退款金额')
    return
  }
  try {
    await request.put(`/pinche/admin/refunds/${approveForm.id}/process`, {
      action: 'approve',
      refund_method: approveForm.refund_method,
      actual_amount: approveForm.actual_amount,
      handle_note: approveForm.handle_note,
      status: 2
    })
    ElMessage.success('退款已批准')
    approveVisible.value = false
    await loadList()
  } catch (e) { /* ignore */ }
}

const onReject = async (row) => {
  try {
    const { value } = await ElMessageBox.prompt('请输入驳回原因', '驳回退款', {
      inputType: 'textarea',
      inputPlaceholder: '驳回原因',
      inputValidator: (v) => !!v || '请输入驳回原因'
    })
    await request.put(`/pinche/admin/refunds/${row.id}/process`, {
      action: 'reject',
      reject_reason: value,
      status: 3
    })
    ElMessage.success('退款已驳回')
    await loadList()
  } catch (e) { /* cancel */ }
}

const onMarkPaid = async (row) => {
  try {
    await ElMessageBox.confirm(
      `确认退款单 "${row.refund_no}" 已实际打款到用户账户？`, '标记已打款',
      { type: 'warning' }
    )
    await request.put(`/pinche/admin/refunds/${row.id}/process`, {
      action: 'mark_paid',
      status: 2,
      paid_at: new Date().toISOString()
    })
    ElMessage.success('已标记为打款成功')
    await loadList()
  } catch (e) { /* cancel */ }
}

const onBatchApprove = async () => {
  if (!selection.value.length) return
  try {
    const { value } = await ElMessageBox.prompt(
      `确认批量批准 ${selection.value.length} 笔退款？请填写统一处理说明`,
      '批量批准',
      { inputType: 'textarea', inputPlaceholder: '处理说明', inputValidator: (v) => !!v || '请输入处理说明' }
    )
    await Promise.all(selection.value.map(i =>
      request.put(`/pinche/admin/refunds/${i.id}/process`, {
        action: 'approve',
        refund_method: i.refund_method || 'original',
        actual_amount: i.refund_amount || 0,
        handle_note: value,
        status: 2
      })
    ))
    ElMessage.success(`已批量批准 ${selection.value.length} 笔退款`)
    await loadList()
  } catch (e) { /* cancel */ }
}

const onBatchReject = async () => {
  if (!selection.value.length) return
  try {
    const { value } = await ElMessageBox.prompt(
      `确认批量驳回 ${selection.value.length} 笔退款？请填写驳回原因`,
      '批量驳回',
      { inputType: 'textarea', inputPlaceholder: '驳回原因', inputValidator: (v) => !!v || '请输入驳回原因' }
    )
    await Promise.all(selection.value.map(i =>
      request.put(`/pinche/admin/refunds/${i.id}/process`, {
        action: 'reject',
        reject_reason: value,
        status: 3
      })
    ))
    ElMessage.success(`已批量驳回 ${selection.value.length} 笔退款`)
    await loadList()
  } catch (e) { /* cancel */ }
}

const onExport = () => {
  ElMessage.success('退款报表已导出（模拟）')
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
.text-muted { color: #909399; font-size: 12px; }
.reason-text {
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
  text-overflow: ellipsis;
  word-break: break-all;
}
.pagination-wrap { margin-top: 16px; display: flex; justify-content: flex-end; }
</style>
