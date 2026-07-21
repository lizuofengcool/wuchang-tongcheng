<template>
  <div class="app-container">
    <!-- 退款统计 -->
    <el-row :gutter="16" class="stat-row">
      <el-col :xs="12" :sm="8" :md="6">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value">{{ stats.total_count }}</div><div class="stat-label">退款单总数</div></div></el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="6">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value text-warning">¥{{ Number(stats.total_amount || 0).toFixed(2) }}</div><div class="stat-label">退款总额</div></div></el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="6">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value text-primary">{{ stats.pending_count }}</div><div class="stat-label">待处理</div></div></el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="6">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value text-success">{{ stats.approved_count }}</div><div class="stat-label">已通过</div></div></el-card>
      </el-col>
    </el-row>

    <div class="page-card">
      <el-form :inline="true" :model="filters" class="filter-form" @submit.prevent>
        <el-form-item label="退款单号">
          <el-input v-model="filters.refund_no" placeholder="退款单号" clearable style="width: 200px" :prefix-icon="Search" @keyup.enter="onSearch" />
        </el-form-item>
        <el-form-item label="订单号">
          <el-input v-model="filters.order_no" placeholder="订单号" clearable style="width: 200px" @keyup.enter="onSearch" />
        </el-form-item>
        <el-form-item label="店铺ID">
          <el-input-number v-model="filters.shop_id" :controls="false" placeholder="店铺ID" style="width: 140px" />
        </el-form-item>
        <el-form-item label="退款状态">
          <el-select v-model="filters.status" placeholder="全部" clearable style="width: 130px" @change="onSearch">
            <el-option label="待卖家处理" :value="0" />
            <el-option label="卖家同意" :value="1" />
            <el-option label="卖家拒绝" :value="2" />
            <el-option label="待买家发货" :value="3" />
            <el-option label="退款中" :value="4" />
            <el-option label="退款成功" :value="5" />
            <el-option label="退款关闭" :value="6" />
          </el-select>
        </el-form-item>
        <el-form-item label="退款类型">
          <el-select v-model="filters.refund_type" placeholder="全部" clearable style="width: 120px" @change="onSearch">
            <el-option label="仅退款" value="refund_only" />
            <el-option label="退货退款" value="return_refund" />
          </el-select>
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
        <el-table-column label="退款单号" min-width="200">
          <template #default="{ row }">
            <el-link type="primary" :underline="false" @click="openDetail(row)">{{ row.refund_no }}</el-link>
          </template>
        </el-table-column>
        <el-table-column label="订单号" width="180">
          <template #default="{ row }">{{ row.order_no || `#${row.order_id}` }}</template>
        </el-table-column>
        <el-table-column label="店铺ID" width="90" prop="shop_id" />
        <el-table-column label="用户ID" width="90" prop="user_id" />
        <el-table-column label="退款金额" width="120">
          <template #default="{ row }">¥{{ Number(row.refund_amount || 0).toFixed(2) }}</template>
        </el-table-column>
        <el-table-column label="退款类型" width="100">
          <template #default="{ row }">
            <el-tag size="small">{{ refundTypeText(row.refund_type) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="110">
          <template #default="{ row }">
            <el-tag :type="statusTagType(row.status)" size="small">{{ statusText(row.status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="申请时间" width="160" prop="created_at" sortable="custom">
          <template #default="{ row }">{{ formatTime(row.created_at) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="openDetail(row)">详情</el-button>
            <el-button v-if="row.status === 0 || row.status === 4" type="success" link size="small" @click="openProcess(row, 5)">同意退款</el-button>
            <el-button v-if="row.status === 0 || row.status === 4" type="danger" link size="small" @click="openProcess(row, 6)">拒绝退款</el-button>
          </template>
        </el-table-column>
      </el-table>

      <div class="pagination-wrap">
        <el-pagination v-model:current-page="page" v-model:page-size="pageSize" :total="total" :page-sizes="[10, 20, 50, 100]" layout="total, sizes, prev, pager, next, jumper" background @current-change="loadList" @size-change="loadList" />
      </div>
    </div>

    <!-- 详情弹窗 -->
    <el-dialog v-model="detailVisible" title="退款详情" width="780px" destroy-on-close>
      <el-descriptions v-if="detail" :column="2" border>
        <el-descriptions-item label="ID">{{ detail.id }}</el-descriptions-item>
        <el-descriptions-item label="退款单号">{{ detail.refund_no }}</el-descriptions-item>
        <el-descriptions-item label="订单号">{{ detail.order_no || `#${detail.order_id}` }}</el-descriptions-item>
        <el-descriptions-item label="订单项ID">{{ detail.order_item_id || '-' }}</el-descriptions-item>
        <el-descriptions-item label="店铺ID">{{ detail.shop_id }}</el-descriptions-item>
        <el-descriptions-item label="用户ID">{{ detail.user_id }}</el-descriptions-item>
        <el-descriptions-item label="退款金额">¥{{ Number(detail.refund_amount || 0).toFixed(2) }}</el-descriptions-item>
        <el-descriptions-item label="退款类型">{{ refundTypeText(detail.refund_type) }}</el-descriptions-item>
        <el-descriptions-item label="退款原因" :span="2">{{ detail.reason || '-' }}</el-descriptions-item>
        <el-descriptions-item label="状态">
          <el-tag :type="statusTagType(detail.status)" size="small">{{ statusText(detail.status) }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="卖家处理时间">{{ formatTime(detail.seller_processed_at) }}</el-descriptions-item>
        <el-descriptions-item label="卖家备注" :span="2">{{ detail.seller_remark || '-' }}</el-descriptions-item>
        <el-descriptions-item label="管理员处理时间">{{ formatTime(detail.admin_processed_at) }}</el-descriptions-item>
        <el-descriptions-item label="管理员">{{ detail.admin_name || '-' }}</el-descriptions-item>
        <el-descriptions-item label="管理员备注" :span="2">{{ detail.admin_remark || '-' }}</el-descriptions-item>
        <el-descriptions-item label="退货物流单号" :span="2">{{ detail.return_tracking_no || '-' }}</el-descriptions-item>
        <el-descriptions-item label="创建时间">{{ formatTime(detail.created_at) }}</el-descriptions-item>
        <el-descriptions-item label="更新时间">{{ formatTime(detail.updated_at) }}</el-descriptions-item>
      </el-descriptions>
    </el-dialog>

    <!-- 处理退款弹窗 -->
    <el-dialog v-model="processVisible" title="处理退款" width="480px" destroy-on-close>
      <el-form :model="processForm" label-width="100px">
        <el-form-item label="退款单号">{{ processForm.refund_no }}</el-form-item>
        <el-form-item label="退款金额">¥{{ Number(processForm.refund_amount || 0).toFixed(2) }}</el-form-item>
        <el-form-item label="处理结果">
          <el-tag :type="processForm.status === 5 ? 'success' : 'danger'">{{ processForm.status === 5 ? '同意退款' : '拒绝退款' }}</el-tag>
        </el-form-item>
        <el-form-item label="处理备注">
          <el-input v-model="processForm.admin_remark" type="textarea" :rows="3" placeholder="处理备注" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="processVisible = false">取消</el-button>
        <el-button type="primary" :loading="processLoading" @click="onProcessSubmit">确认处理</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { Refresh, RefreshLeft, Search } from '@element-plus/icons-vue'
import { formatTime } from '@/utils/format'
import { getMallRefundList, getMallRefundDetail, adminProcessMallRefund, getMallRefundStats } from '@/api/mall'

const loading = ref(false)
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)
const list = ref([])
const sortField = ref('created_at')
const sortOrder = ref('descending')

const stats = reactive({ total_count: 0, total_amount: 0, pending_count: 0, approved_count: 0 })

const filters = reactive({ refund_no: '', order_no: '', shop_id: null, status: null, refund_type: '' })

const onSearch = () => { page.value = 1; loadList() }
const onReset = () => {
  filters.refund_no = ''
  filters.order_no = ''
  filters.shop_id = null
  filters.status = null
  filters.refund_type = ''
  page.value = 1
  loadList()
}
const onSortChange = ({ prop, order }) => {
  sortField.value = prop || 'created_at'
  sortOrder.value = order || 'descending'
  loadList()
}

const statusText = (s) => ({ 0: '待卖家处理', 1: '卖家同意', 2: '卖家拒绝', 3: '待买家发货', 4: '退款中', 5: '退款成功', 6: '退款关闭' }[s] || '-')
const statusTagType = (s) => ({ 0: 'warning', 1: 'primary', 2: 'danger', 3: 'info', 4: 'warning', 5: 'success', 6: 'info' }[s] || 'info')
const refundTypeText = (t) => ({ refund_only: '仅退款', return_refund: '退货退款' }[t] || t || '-')

const loadList = async () => {
  loading.value = true
  try {
    const params = {
      page: page.value,
      page_size: pageSize.value,
      refund_no: filters.refund_no.trim() || undefined,
      order_no: filters.order_no.trim() || undefined,
      shop_id: filters.shop_id || undefined,
      status: filters.status === null || filters.status === '' ? undefined : filters.status,
      refund_type: filters.refund_type || undefined,
      sort: sortField.value,
      order: sortOrder.value
    }
    const res = await getMallRefundList(params)
    const data = res.data || {}
    list.value = data.list || []
    total.value = data.total || 0
  } catch (e) {
    ElMessage.error('加载退款列表失败')
    list.value = []
    total.value = 0
  } finally {
    loading.value = false
  }
}

const loadStats = async () => {
  try {
    const res = await getMallRefundStats()
    Object.assign(stats, res.data || {})
  } catch (e) { /* ignore */ }
}

const detailVisible = ref(false)
const detail = ref(null)

const openDetail = async (row) => {
  try {
    const res = await getMallRefundDetail(row.id)
    detail.value = res.data || null
    detailVisible.value = true
  } catch (e) {
    ElMessage.error('加载详情失败')
  }
}

const processVisible = ref(false)
const processLoading = ref(false)
const processForm = reactive({ id: null, refund_no: '', refund_amount: 0, status: 5, admin_remark: '' })

const openProcess = (row, status) => {
  Object.assign(processForm, {
    id: row.id,
    refund_no: row.refund_no,
    refund_amount: row.refund_amount,
    status,
    admin_remark: ''
  })
  processVisible.value = true
}

const onProcessSubmit = async () => {
  try {
    processLoading.value = true
    await adminProcessMallRefund(processForm.id, {
      status: processForm.status,
      admin_remark: processForm.admin_remark
    })
    ElMessage.success('处理成功')
    processVisible.value = false
    await loadList()
  } catch (e) {
    ElMessage.error('处理失败')
  } finally {
    processLoading.value = false
  }
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
.text-primary { color: #409eff; }
.filter-form { margin-bottom: 12px; padding: 12px 16px; background: #fafafa; border-radius: 4px; }
.filter-form :deep(.el-form-item) { margin-bottom: 8px; margin-right: 12px; }
.toolbar { display: flex; justify-content: space-between; margin-bottom: 12px; }
.pagination-wrap { margin-top: 16px; display: flex; justify-content: flex-end; }
</style>
