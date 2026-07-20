<template>
  <div class="app-container">
    <!-- 顶部统计卡片 -->
    <el-row :gutter="16" class="stat-row">
      <el-col :xs="12" :sm="8" :md="6">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-icon" style="background: #409eff">
            <el-icon :size="22"><Tickets /></el-icon>
          </div>
          <div class="stat-content">
            <div class="stat-value">{{ stats.total }}</div>
            <div class="stat-label">总预订数</div>
          </div>
        </el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="6">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-icon" style="background: #e6a23c">
            <el-icon :size="22"><Clock /></el-icon>
          </div>
          <div class="stat-content">
            <div class="stat-value">{{ stats.pending }}</div>
            <div class="stat-label">待确认</div>
          </div>
        </el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="6">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-icon" style="background: #67c23a">
            <el-icon :size="22"><CircleCheck /></el-icon>
          </div>
          <div class="stat-content">
            <div class="stat-value">{{ stats.confirmed }}</div>
            <div class="stat-label">已确认</div>
          </div>
        </el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="6">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-icon" style="background: #f56c6c">
            <el-icon :size="22"><CircleClose /></el-icon>
          </div>
          <div class="stat-content">
            <div class="stat-value">{{ stats.cancelled }}</div>
            <div class="stat-label">已取消</div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <div class="page-card">
      <!-- 筛选区 -->
      <el-form :inline="true" :model="filters" class="filter-form" @submit.prevent>
        <el-form-item label="预订单号">
          <el-input v-model="filters.booking_no" placeholder="预订单号" clearable style="width: 180px" @keyup.enter="onSearch" />
        </el-form-item>
        <el-form-item label="拼车ID">
          <el-input v-model="filters.pinche_id" placeholder="拼车ID" clearable style="width: 120px" @keyup.enter="onSearch" />
        </el-form-item>
        <el-form-item label="乘客ID">
          <el-input v-model="filters.passenger_id" placeholder="乘客ID" clearable style="width: 120px" @keyup.enter="onSearch" />
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="filters.status" placeholder="全部" clearable style="width: 130px" @change="onSearch">
            <el-option label="待确认" :value="0" />
            <el-option label="已确认" :value="1" />
            <el-option label="已上车" :value="2" />
            <el-option label="已完成" :value="3" />
            <el-option label="已取消" :value="4" />
            <el-option label="已违约" :value="5" />
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
          <el-button type="primary" :icon="Plus" @click="onCreate">新建预订</el-button>
        </div>
      </div>

      <!-- 表格 -->
      <el-table v-loading="loading" :data="list" border stripe>
        <el-table-column prop="id" label="ID" width="70" fixed="left" />
        <el-table-column prop="booking_no" label="预订单号" width="180" />
        <el-table-column label="路线" min-width="220">
          <template #default="{ row }">
            <div class="route-text">
              <span class="from">{{ row.origin }}</span>
              <el-icon class="arrow"><Right /></el-icon>
              <span class="to">{{ row.destination }}</span>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="乘客" width="140">
          <template #default="{ row }">
            <div>{{ row.passenger_name || `#${row.passenger_id}` }}</div>
            <div class="text-muted">{{ maskPhone(row.passenger_phone) }}</div>
          </template>
        </el-table-column>
        <el-table-column prop="seats" label="座位数" width="80" />
        <el-table-column label="金额" width="110">
          <template #default="{ row }">
            <span class="price">¥{{ Number(row.total_amount || 0).toFixed(2) }}</span>
          </template>
        </el-table-column>
        <el-table-column label="上车地点" min-width="140" show-overflow-tooltip>
          <template #default="{ row }">{{ row.boarding_point || '-' }}</template>
        </el-table-column>
        <el-table-column label="下车地点" min-width="140" show-overflow-tooltip>
          <template #default="{ row }">{{ row.dropoff_point || '-' }}</template>
        </el-table-column>
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="statusTagType(row.status)" size="small">{{ statusText(row.status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="创建时间" width="160">
          <template #default="{ row }">{{ formatTime(row.created_at) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="240" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="openDetail(row)">详情</el-button>
            <el-button v-if="row.status === 0" type="success" link size="small" @click="onConfirm(row)">确认</el-button>
            <el-button v-if="row.status === 1" type="warning" link size="small" @click="onBoarding(row)">上车</el-button>
            <el-button v-if="row.status === 2" type="success" link size="small" @click="onComplete(row)">完成</el-button>
            <el-button v-if="row.status === 0 || row.status === 1" type="danger" link size="small" @click="onCancel(row)">取消</el-button>
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
    <el-dialog v-model="detailVisible" title="预订详情" width="700px">
      <el-descriptions v-if="detail" :column="2" border>
        <el-descriptions-item label="ID">{{ detail.id }}</el-descriptions-item>
        <el-descriptions-item label="预订单号">{{ detail.booking_no }}</el-descriptions-item>
        <el-descriptions-item label="拼车ID">{{ detail.pinche_id }}</el-descriptions-item>
        <el-descriptions-item label="行程ID">{{ detail.trip_id || '-' }}</el-descriptions-item>
        <el-descriptions-item label="乘客">{{ detail.passenger_name || `#${detail.passenger_id}` }}</el-descriptions-item>
        <el-descriptions-item label="乘客电话">{{ maskPhone(detail.passenger_phone) }}</el-descriptions-item>
        <el-descriptions-item label="座位数">{{ detail.seats }}</el-descriptions-item>
        <el-descriptions-item label="金额">
          <span class="price">¥{{ Number(detail.total_amount || 0).toFixed(2) }}</span>
        </el-descriptions-item>
        <el-descriptions-item label="上车地点" :span="2">{{ detail.boarding_point || '-' }}</el-descriptions-item>
        <el-descriptions-item label="下车地点" :span="2">{{ detail.dropoff_point || '-' }}</el-descriptions-item>
        <el-descriptions-item label="状态">
          <el-tag :type="statusTagType(detail.status)" size="small">{{ statusText(detail.status) }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="支付状态">{{ detail.payment_status || '-' }}</el-descriptions-item>
        <el-descriptions-item label="上车时间">{{ formatTime(detail.boarding_time) }}</el-descriptions-item>
        <el-descriptions-item label="完成时间">{{ formatTime(detail.completed_time) }}</el-descriptions-item>
        <el-descriptions-item v-if="detail.cancel_reason" label="取消原因" :span="2">{{ detail.cancel_reason }}</el-descriptions-item>
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
  Refresh, RefreshLeft, Search, Plus, Right,
  Tickets, Clock, CircleCheck, CircleClose
} from '@element-plus/icons-vue'
import { formatTime } from '@/utils/format'

const loading = ref(false)
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)
const list = ref([])

const stats = reactive({ total: 0, pending: 0, confirmed: 0, cancelled: 0 })

const filters = reactive({
  booking_no: '', pinche_id: '', passenger_id: '',
  status: null, dateRange: null
})

const statusText = (s) => ({ 0: '待确认', 1: '已确认', 2: '已上车', 3: '已完成', 4: '已取消', 5: '已违约' }[s] || '-')
const statusTagType = (s) => ({ 0: 'warning', 1: 'primary', 2: 'info', 3: 'success', 4: 'danger', 5: 'danger' }[s] || 'info')

const maskPhone = (p) => {
  if (!p) return '-'
  const s = String(p)
  if (s.length < 7) return s
  return s.slice(0, 3) + '****' + s.slice(-4)
}

const onSearch = () => { page.value = 1; loadList() }
const onReset = () => {
  filters.booking_no = ''; filters.pinche_id = ''; filters.passenger_id = ''
  filters.status = null; filters.dateRange = null
  page.value = 1; loadList()
}

const loadList = async () => {
  loading.value = true
  try {
    const params = {
      page: page.value,
      page_size: pageSize.value,
      booking_no: filters.booking_no || undefined,
      pinche_id: filters.pinche_id || undefined,
      passenger_id: filters.passenger_id || undefined,
      status: filters.status === null || filters.status === '' ? undefined : filters.status
    }
    if (filters.dateRange && filters.dateRange.length === 2) {
      params.start_date = filters.dateRange[0]
      params.end_date = filters.dateRange[1]
    }
    const res = await request.get('/pinche/bookings', { params })
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
    const res = await request.get(`/pinche/bookings/${row.id}`)
    if (res.data) detail.value = res.data
  } catch (e) { /* ignore */ }
}

const onConfirm = async (row) => {
  try {
    await ElMessageBox.confirm(`确认预订 "${row.booking_no}"？`, '提示', { type: 'warning' })
    await request.put(`/pinche/admin/bookings/${row.id}/status`, { status: 1 })
    ElMessage.success('预订已确认')
    await loadList()
  } catch (e) { /* cancel */ }
}

const onBoarding = async (row) => {
  try {
    await ElMessageBox.confirm(`确认乘客已上车 "${row.booking_no}"？`, '提示', { type: 'warning' })
    await request.post(`/pinche/bookings/${row.id}/boarding`)
    ElMessage.success('已确认上车')
    await loadList()
  } catch (e) { /* cancel */ }
}

const onComplete = async (row) => {
  try {
    await ElMessageBox.confirm(`确认完成预订 "${row.booking_no}"？`, '提示', { type: 'warning' })
    await request.put(`/pinche/admin/bookings/${row.id}/status`, { status: 3 })
    ElMessage.success('预订已完成')
    await loadList()
  } catch (e) { /* cancel */ }
}

const onCancel = async (row) => {
  try {
    const { value } = await ElMessageBox.prompt('请输入取消原因（可选）', '取消预订', {
      inputType: 'textarea',
      inputPlaceholder: '取消原因（可不填）'
    })
    await request.post(`/pinche/bookings/${row.id}/cancel`, { cancel_reason: value || '' })
    ElMessage.success('预订已取消')
    await loadList()
  } catch (e) { /* cancel */ }
}

const onCreate = () => {
  ElMessage.info('新建预订功能开发中，请使用 C 端发起')
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

.route-text { font-weight: 500; color: #303133; display: flex; align-items: center; gap: 6px; }
.from, .to { color: #303133; }
.arrow { color: #909399; }
.price { color: #f56c6c; font-weight: 600; }
.text-muted { color: #909399; font-size: 12px; }
.pagination-wrap { margin-top: 16px; display: flex; justify-content: flex-end; }
</style>
