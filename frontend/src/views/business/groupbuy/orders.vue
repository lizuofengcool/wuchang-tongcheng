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
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value text-info">{{ stats.used }}</div><div class="stat-label">已使用</div></div></el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="4">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value text-success">{{ stats.completed }}</div><div class="stat-label">已完成</div></div></el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="4">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value text-danger">{{ stats.canceled }}</div><div class="stat-label">已取消</div></div></el-card>
      </el-col>
    </el-row>

    <div class="page-card">
      <el-form :inline="true" :model="filters" class="filter-form" @submit.prevent>
        <el-form-item label="订单号">
          <el-input v-model="filters.order_no" placeholder="订单号" clearable style="width: 200px" :prefix-icon="Search" @keyup.enter="onSearch" />
        </el-form-item>
        <el-form-item label="团购ID">
          <el-input-number v-model="filters.groupbuy_id" :controls="false" placeholder="团购ID" style="width: 140px" />
        </el-form-item>
        <el-form-item label="用户ID">
          <el-input-number v-model="filters.user_id" :controls="false" placeholder="用户ID" style="width: 140px" />
        </el-form-item>
        <el-form-item label="店铺ID">
          <el-input-number v-model="filters.shop_id" :controls="false" placeholder="店铺ID" style="width: 140px" />
        </el-form-item>
        <el-form-item label="订单状态">
          <el-select v-model="filters.status" placeholder="全部" clearable style="width: 120px" @change="onSearch">
            <el-option label="待付款" :value="0" />
            <el-option label="已付款" :value="1" />
            <el-option label="已使用" :value="2" />
            <el-option label="已完成" :value="3" />
            <el-option label="已取消" :value="4" />
            <el-option label="已退款" :value="5" />
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
        </div>
      </div>

      <el-table v-loading="loading" :data="list" border stripe @sort-change="onSortChange">
        <el-table-column prop="id" label="ID" width="70" sortable="custom" fixed="left" />
        <el-table-column label="订单号" min-width="200">
          <template #default="{ row }">
            <el-link type="primary" :underline="false" @click="openDetail(row)">{{ row.order_no }}</el-link>
          </template>
        </el-table-column>
        <el-table-column label="团购商品" min-width="200">
          <template #default="{ row }">
            <div>{{ row.groupbuy_title || `#${row.groupbuy_id}` }}</div>
            <div class="text-muted text-xs">¥{{ Number(row.unit_price || 0).toFixed(2) }} × {{ row.quantity || 1 }}</div>
          </template>
        </el-table-column>
        <el-table-column label="店铺" width="120">
          <template #default="{ row }">{{ row.shop_name || `#${row.shop_id}` }}</template>
        </el-table-column>
        <el-table-column label="买家" width="140">
          <template #default="{ row }">
            <div>{{ row.buyer_name || `用户#${row.user_id}` }}</div>
            <div class="text-muted text-xs">{{ row.buyer_phone || '-' }}</div>
          </template>
        </el-table-column>
        <el-table-column label="总金额" width="110">
          <template #default="{ row }">¥{{ Number(row.total_amount || 0).toFixed(2) }}</template>
        </el-table-column>
        <el-table-column label="实付" width="110">
          <template #default="{ row }">¥{{ Number(row.pay_amount || 0).toFixed(2) }}</template>
        </el-table-column>
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="statusTagType(row.status)" size="small">{{ statusText(row.status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="创建时间" width="160" prop="created_at" sortable="custom">
          <template #default="{ row }">{{ formatTime(row.created_at) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="160" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="openDetail(row)">详情</el-button>
            <el-button v-if="row.status < 4 && row.status !== 5" type="danger" link size="small" @click="onClose(row)">关闭</el-button>
          </template>
        </el-table-column>
      </el-table>

      <div class="pagination-wrap">
        <el-pagination v-model:current-page="page" v-model:page-size="pageSize" :total="total" :page-sizes="[10, 20, 50, 100]" layout="total, sizes, prev, pager, next, jumper" background @current-change="loadList" @size-change="loadList" />
      </div>
    </div>

    <!-- 详情弹窗 -->
    <el-dialog v-model="detailVisible" title="订单详情" width="800px" destroy-on-close>
      <el-descriptions v-if="detail" :column="3" border>
        <el-descriptions-item label="订单号" :span="2">{{ detail.order_no }}</el-descriptions-item>
        <el-descriptions-item label="订单ID">{{ detail.id }}</el-descriptions-item>
        <el-descriptions-item label="团购商品">{{ detail.groupbuy_title || `#${detail.groupbuy_id}` }}</el-descriptions-item>
        <el-descriptions-item label="店铺">{{ detail.shop_name || `#${detail.shop_id}` }}</el-descriptions-item>
        <el-descriptions-item label="买家">{{ detail.buyer_name || `用户#${detail.user_id}` }}</el-descriptions-item>
        <el-descriptions-item label="买家电话">{{ detail.buyer_phone || '-' }}</el-descriptions-item>
        <el-descriptions-item label="单价">¥{{ Number(detail.unit_price || 0).toFixed(2) }}</el-descriptions-item>
        <el-descriptions-item label="数量">{{ detail.quantity || 1 }}</el-descriptions-item>
        <el-descriptions-item label="总金额">¥{{ Number(detail.total_amount || 0).toFixed(2) }}</el-descriptions-item>
        <el-descriptions-item label="实付金额">¥{{ Number(detail.pay_amount || 0).toFixed(2) }}</el-descriptions-item>
        <el-descriptions-item label="优惠金额">¥{{ Number(detail.discount_amount || 0).toFixed(2) }}</el-descriptions-item>
        <el-descriptions-item label="状态">
          <el-tag :type="statusTagType(detail.status)" size="small">{{ statusText(detail.status) }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="支付方式">{{ detail.pay_method || '-' }}</el-descriptions-item>
        <el-descriptions-item label="收货人">{{ detail.receiver_name || '-' }}</el-descriptions-item>
        <el-descriptions-item label="收货电话">{{ detail.receiver_phone || '-' }}</el-descriptions-item>
        <el-descriptions-item label="收货地址" :span="3">{{ detail.receiver_address || '-' }}</el-descriptions-item>
        <el-descriptions-item label="买家留言" :span="3">{{ detail.buyer_remark || '-' }}</el-descriptions-item>
        <el-descriptions-item label="创建时间">{{ formatTime(detail.created_at) }}</el-descriptions-item>
        <el-descriptions-item label="支付时间">{{ formatTime(detail.paid_at) }}</el-descriptions-item>
        <el-descriptions-item label="使用时间">{{ formatTime(detail.used_at) }}</el-descriptions-item>
        <el-descriptions-item label="完成时间">{{ formatTime(detail.completed_at) }}</el-descriptions-item>
        <el-descriptions-item label="取消时间">{{ formatTime(detail.canceled_at) }}</el-descriptions-item>
        <el-descriptions-item label="更新时间">{{ formatTime(detail.updated_at) }}</el-descriptions-item>
      </el-descriptions>
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
import { getGroupbuyOrderList, getGroupbuyOrderDetail, closeGroupbuyOrder } from '@/api/groupbuy'

const loading = ref(false)
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)
const list = ref([])
const sortField = ref('created_at')
const sortOrder = ref('descending')

const stats = reactive({ total: 0, pending: 0, paid: 0, used: 0, completed: 0, canceled: 0 })

const filters = reactive({ order_no: '', groupbuy_id: null, user_id: null, shop_id: null, status: null, dateRange: null })

const onSearch = () => { page.value = 1; loadList() }
const onReset = () => {
  filters.order_no = ''
  filters.groupbuy_id = null
  filters.user_id = null
  filters.shop_id = null
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

const statusText = (s) => ({ 0: '待付款', 1: '已付款', 2: '已使用', 3: '已完成', 4: '已取消', 5: '已退款' }[s] || '-')
const statusTagType = (s) => ({ 0: 'warning', 1: 'primary', 2: 'info', 3: 'success', 4: 'danger', 5: 'danger' }[s] || 'info')

const loadList = async () => {
  loading.value = true
  try {
    const params = {
      page: page.value,
      page_size: pageSize.value,
      order_no: filters.order_no.trim() || undefined,
      groupbuy_id: filters.groupbuy_id || undefined,
      user_id: filters.user_id || undefined,
      shop_id: filters.shop_id || undefined,
      status: filters.status === null || filters.status === '' ? undefined : filters.status,
      sort: sortField.value,
      order: sortOrder.value
    }
    if (filters.dateRange && filters.dateRange.length === 2) {
      params.start_date = filters.dateRange[0]
      params.end_date = filters.dateRange[1]
    }
    const res = await getGroupbuyOrderList(params)
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
  stats.used = all.filter((r) => r.status === 2).length
  stats.completed = all.filter((r) => r.status === 3).length
  stats.canceled = all.filter((r) => r.status === 4 || r.status === 5).length
}

const detailVisible = ref(false)
const detail = ref(null)

const openDetail = async (row) => {
  try {
    const res = await getGroupbuyOrderDetail(row.id)
    detail.value = res.data || null
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
    await closeGroupbuyOrder(closeForm.id, { close_reason: closeForm.close_reason })
    ElMessage.success('订单已关闭')
    closeVisible.value = false
    await loadList()
  } catch (e) {
    ElMessage.error('关闭失败')
  } finally {
    closeLoading.value = false
  }
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
</style>
