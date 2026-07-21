<template>
  <div class="app-container">
    <!-- 顶部统计 -->
    <el-row :gutter="16" class="stat-row">
      <el-col :xs="12" :sm="8" :md="4">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value">{{ stats.total }}</div><div class="stat-label">订单总数</div></div></el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="4">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value text-warning">{{ stats.pending }}</div><div class="stat-label">待付款</div></div></el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="4">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value text-primary">{{ stats.paid }}</div><div class="stat-label">已付款</div></div></el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="4">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value text-info">{{ stats.shipped }}</div><div class="stat-label">已发货</div></div></el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="4">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value text-success">{{ stats.completed }}</div><div class="stat-label">已完成</div></div></el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="4">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value text-danger">{{ stats.closed }}</div><div class="stat-label">已关闭</div></div></el-card>
      </el-col>
    </el-row>

    <div class="page-card">
      <el-form :inline="true" :model="filters" class="filter-form" @submit.prevent>
        <el-form-item label="订单号">
          <el-input v-model="filters.order_no" placeholder="订单号" clearable style="width: 200px" :prefix-icon="Search" @keyup.enter="onSearch" />
        </el-form-item>
        <el-form-item label="店铺ID">
          <el-input-number v-model="filters.shop_id" :controls="false" placeholder="店铺ID" style="width: 140px" />
        </el-form-item>
        <el-form-item label="用户ID">
          <el-input-number v-model="filters.user_id" :controls="false" placeholder="用户ID" style="width: 140px" />
        </el-form-item>
        <el-form-item label="订单状态">
          <el-select v-model="filters.status" placeholder="全部" clearable style="width: 120px" @change="onSearch">
            <el-option label="待付款" :value="0" />
            <el-option label="已付款" :value="1" />
            <el-option label="已发货" :value="2" />
            <el-option label="已收货" :value="3" />
            <el-option label="已完成" :value="4" />
            <el-option label="已取消" :value="5" />
            <el-option label="已关闭" :value="6" />
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
        <div class="toolbar-left">
          <el-button :icon="Refresh" @click="loadList">刷新</el-button>
          <el-button type="warning" @click="onAutoClose">自动关闭超时订单</el-button>
          <el-button type="success" @click="onAutoConfirm">自动确认收货</el-button>
          <el-button type="primary" @click="onAutoReview">自动评价</el-button>
        </div>
      </div>

      <el-table v-loading="loading" :data="list" border stripe @sort-change="onSortChange">
        <el-table-column prop="id" label="ID" width="70" sortable="custom" fixed="left" />
        <el-table-column label="订单号" min-width="200">
          <template #default="{ row }">
            <el-link type="primary" :underline="false" @click="openDetail(row)">{{ row.order_no }}</el-link>
          </template>
        </el-table-column>
        <el-table-column label="店铺" width="120">
          <template #default="{ row }">{{ row.shop_name || `#${row.shop_id}` }}</template>
        </el-table-column>
        <el-table-column label="买家" width="120">
          <template #default="{ row }">
            <div>{{ row.buyer_name || `用户#${row.user_id}` }}</div>
            <div class="text-muted text-xs">{{ row.buyer_phone || '-' }}</div>
          </template>
        </el-table-column>
        <el-table-column label="商品数" width="80" prop="item_count" />
        <el-table-column label="总金额" width="120">
          <template #default="{ row }">¥{{ Number(row.total_amount || 0).toFixed(2) }}</template>
        </el-table-column>
        <el-table-column label="实付" width="120">
          <template #default="{ row }">¥{{ Number(row.pay_amount || 0).toFixed(2) }}</template>
        </el-table-column>
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="statusTagType(row.status)" size="small">{{ statusText(row.status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="支付方式" width="100">
          <template #default="{ row }">{{ row.pay_method || '-' }}</template>
        </el-table-column>
        <el-table-column label="创建时间" width="160" prop="created_at" sortable="custom">
          <template #default="{ row }">{{ formatTime(row.created_at) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="180" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="openDetail(row)">详情</el-button>
            <el-button v-if="row.status < 6 && row.status !== 4" type="danger" link size="small" @click="onClose(row)">关闭</el-button>
          </template>
        </el-table-column>
      </el-table>

      <div class="pagination-wrap">
        <el-pagination v-model:current-page="page" v-model:page-size="pageSize" :total="total" :page-sizes="[10, 20, 50, 100]" layout="total, sizes, prev, pager, next, jumper" background @current-change="loadList" @size-change="loadList" />
      </div>
    </div>

    <!-- 详情弹窗 -->
    <el-dialog v-model="detailVisible" title="订单详情" width="900px" destroy-on-close>
      <el-descriptions v-if="detail" :column="3" border>
        <el-descriptions-item label="订单号" :span="2">{{ detail.order_no }}</el-descriptions-item>
        <el-descriptions-item label="订单ID">{{ detail.id }}</el-descriptions-item>
        <el-descriptions-item label="店铺">{{ detail.shop_name || `#${detail.shop_id}` }}</el-descriptions-item>
        <el-descriptions-item label="买家">{{ detail.buyer_name || `用户#${detail.user_id}` }}</el-descriptions-item>
        <el-descriptions-item label="买家电话">{{ detail.buyer_phone || '-' }}</el-descriptions-item>
        <el-descriptions-item label="商品数">{{ detail.item_count || 0 }}</el-descriptions-item>
        <el-descriptions-item label="总金额">¥{{ Number(detail.total_amount || 0).toFixed(2) }}</el-descriptions-item>
        <el-descriptions-item label="运费">¥{{ Number(detail.shipping_fee || 0).toFixed(2) }}</el-descriptions-item>
        <el-descriptions-item label="优惠金额">¥{{ Number(detail.discount_amount || 0).toFixed(2) }}</el-descriptions-item>
        <el-descriptions-item label="实付金额">¥{{ Number(detail.pay_amount || 0).toFixed(2) }}</el-descriptions-item>
        <el-descriptions-item label="支付方式">{{ detail.pay_method || '-' }}</el-descriptions-item>
        <el-descriptions-item label="状态">
          <el-tag :type="statusTagType(detail.status)" size="small">{{ statusText(detail.status) }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="收货人">{{ detail.receiver_name || '-' }}</el-descriptions-item>
        <el-descriptions-item label="收货电话">{{ detail.receiver_phone || '-' }}</el-descriptions-item>
        <el-descriptions-item label="收货地址" :span="3">{{ detail.receiver_address || '-' }}</el-descriptions-item>
        <el-descriptions-item label="买家留言" :span="3">{{ detail.buyer_remark || '-' }}</el-descriptions-item>
        <el-descriptions-item label="卖家备注" :span="3">{{ detail.seller_remark || '-' }}</el-descriptions-item>
        <el-descriptions-item v-if="detail.close_reason" label="关闭原因" :span="3">{{ detail.close_reason }}</el-descriptions-item>
        <el-descriptions-item label="创建时间">{{ formatTime(detail.created_at) }}</el-descriptions-item>
        <el-descriptions-item label="支付时间">{{ formatTime(detail.paid_at) }}</el-descriptions-item>
        <el-descriptions-item label="发货时间">{{ formatTime(detail.shipped_at) }}</el-descriptions-item>
        <el-descriptions-item label="完成时间">{{ formatTime(detail.completed_at) }}</el-descriptions-item>
        <el-descriptions-item label="更新时间" :span="2">{{ formatTime(detail.updated_at) }}</el-descriptions-item>
      </el-descriptions>

      <h4 class="sub-title">订单明细</h4>
      <el-table :data="orderItems" border size="small">
        <el-table-column prop="id" label="ID" width="70" />
        <el-table-column label="商品" min-width="200">
          <template #default="{ row }">
            <div>{{ row.product_name || `#${row.product_id}` }}</div>
            <div class="text-muted text-xs">{{ row.sku_name || row.sku_no || '-' }}</div>
          </template>
        </el-table-column>
        <el-table-column label="单价" width="100">
          <template #default="{ row }">¥{{ Number(row.unit_price || 0).toFixed(2) }}</template>
        </el-table-column>
        <el-table-column prop="quantity" label="数量" width="80" />
        <el-table-column label="小计" width="120">
          <template #default="{ row }">¥{{ Number(row.total_price || 0).toFixed(2) }}</template>
        </el-table-column>
      </el-table>
    </el-dialog>

    <!-- 关闭订单弹窗 -->
    <el-dialog v-model="closeVisible" title="关闭订单" width="480px" destroy-on-close>
      <el-form :model="closeForm" label-width="100px">
        <el-form-item label="订单号">{{ closeForm.order_no }}</el-form-item>
        <el-form-item label="关闭原因">
          <el-input v-model="closeForm.close_reason" type="textarea" :rows="3" placeholder="请输入关闭原因" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="closeVisible = false">取消</el-button>
        <el-button type="danger" :loading="closeLoading" @click="onCloseSubmit">确认关闭</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Refresh, RefreshLeft, Search } from '@element-plus/icons-vue'
import { formatTime } from '@/utils/format'
import {
  getMallOrderList, getMallOrderDetail, closeMallOrder,
  autoCloseMallOrders, autoConfirmMallOrders, autoReviewMallOrders,
  getMallOrderItemsByOrder
} from '@/api/mall'

const loading = ref(false)
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)
const list = ref([])
const sortField = ref('created_at')
const sortOrder = ref('descending')

const stats = reactive({ total: 0, pending: 0, paid: 0, shipped: 0, completed: 0, closed: 0 })

const filters = reactive({ order_no: '', shop_id: null, user_id: null, status: null, dateRange: null })

const onSearch = () => { page.value = 1; loadList() }
const onReset = () => {
  filters.order_no = ''
  filters.shop_id = null
  filters.user_id = null
  filters.status = null
  filters.dateRange = null
  page.value = 1
  loadList()
}
const onSortChange = ({ prop, order }) => {
  sortField.value = prop || 'created_at'
  sortOrder.value = order || 'descending'
  loadList()
}

const statusText = (s) => ({ 0: '待付款', 1: '已付款', 2: '已发货', 3: '已收货', 4: '已完成', 5: '已取消', 6: '已关闭' }[s] || '-')
const statusTagType = (s) => ({ 0: 'warning', 1: 'primary', 2: 'info', 3: 'info', 4: 'success', 5: 'danger', 6: 'danger' }[s] || 'info')

const loadList = async () => {
  loading.value = true
  try {
    const params = {
      page: page.value,
      page_size: pageSize.value,
      order_no: filters.order_no.trim() || undefined,
      shop_id: filters.shop_id || undefined,
      user_id: filters.user_id || undefined,
      status: filters.status === null || filters.status === '' ? undefined : filters.status,
      sort: sortField.value,
      order: sortOrder.value
    }
    if (filters.dateRange && filters.dateRange.length === 2) {
      params.start_date = filters.dateRange[0]
      params.end_date = filters.dateRange[1]
    }
    const res = await getMallOrderList(params)
    const data = res.data || {}
    list.value = data.list || []
    total.value = data.total || 0
    calcStats()
  } catch (e) {
    ElMessage.error('加载订单列表失败')
    list.value = []
    total.value = 0
  } finally {
    loading.value = false
  }
}

const calcStats = () => {
  const all = list.value || []
  stats.total = total.value
  stats.pending = all.filter((r) => r.status === 0).length
  stats.paid = all.filter((r) => r.status === 1).length
  stats.shipped = all.filter((r) => r.status === 2).length
  stats.completed = all.filter((r) => r.status === 4).length
  stats.closed = all.filter((r) => r.status === 6 || r.status === 5).length
}

const detailVisible = ref(false)
const detail = ref(null)
const orderItems = ref([])

const openDetail = async (row) => {
  try {
    const [detailRes, itemsRes] = await Promise.all([
      getMallOrderDetail(row.id),
      getMallOrderItemsByOrder(row.id)
    ])
    detail.value = detailRes.data || null
    const itemsData = itemsRes.data
    orderItems.value = Array.isArray(itemsData) ? itemsData : (itemsData?.list || [])
    detailVisible.value = true
  } catch (e) {
    ElMessage.error('加载详情失败')
  }
}

const closeVisible = ref(false)
const closeLoading = ref(false)
const closeForm = reactive({ id: null, order_no: '', close_reason: '' })

const onClose = (row) => {
  Object.assign(closeForm, { id: row.id, order_no: row.order_no, close_reason: '' })
  closeVisible.value = true
}

const onCloseSubmit = async () => {
  try {
    closeLoading.value = true
    await closeMallOrder(closeForm.id, { close_reason: closeForm.close_reason })
    ElMessage.success('订单已关闭')
    closeVisible.value = false
    await loadList()
  } catch (e) {
    ElMessage.error('关闭失败')
  } finally {
    closeLoading.value = false
  }
}

const onAutoClose = async () => {
  try {
    await ElMessageBox.confirm('确定执行自动关闭超时未付款订单？', '提示', { type: 'warning' })
    const res = await autoCloseMallOrders()
    ElMessage.success(`已关闭 ${res.data?.closed_count || 0} 个订单`)
    await loadList()
  } catch (e) { /* cancel */ }
}

const onAutoConfirm = async () => {
  try {
    await ElMessageBox.confirm('确定执行自动确认收货？', '提示', { type: 'warning' })
    const res = await autoConfirmMallOrders()
    ElMessage.success(`已确认 ${res.data?.confirmed_count || 0} 个订单`)
    await loadList()
  } catch (e) { /* cancel */ }
}

const onAutoReview = async () => {
  try {
    await ElMessageBox.confirm('确定执行自动评价？', '提示', { type: 'warning' })
    const res = await autoReviewMallOrders()
    ElMessage.success(`已评价 ${res.data?.reviewed_count || 0} 个订单`)
    await loadList()
  } catch (e) { /* cancel */ }
}

onMounted(() => loadList())
</script>

<style scoped>
.stat-row { margin-bottom: 16px; }
.stat-item { text-align: center; padding: 16px; }
.stat-value { font-size: 22px; font-weight: 600; color: #303133; }
.text-success { color: #67c23a; }
.text-warning { color: #e6a23c; }
.text-danger { color: #f56c6c; }
.text-primary { color: #409eff; }
.text-info { color: #909399; }
.filter-form { margin-bottom: 12px; padding: 12px 16px; background: #fafafa; border-radius: 4px; }
.filter-form :deep(.el-form-item) { margin-bottom: 8px; margin-right: 12px; }
.toolbar { display: flex; justify-content: space-between; align-items: center; margin-bottom: 12px; flex-wrap: wrap; gap: 8px; }
.toolbar-left { display: flex; gap: 8px; flex-wrap: wrap; }
.pagination-wrap { margin-top: 16px; display: flex; justify-content: flex-end; }
.text-muted { color: #909399; }
.text-xs { font-size: 12px; }
.sub-title { margin: 16px 0 8px; font-weight: 600; color: #303133; }
</style>
