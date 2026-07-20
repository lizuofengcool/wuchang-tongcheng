<template>
  <div class="app-container">
    <!-- 顶部统计卡片 -->
    <el-row :gutter="16" class="stat-row">
      <el-col :xs="12" :sm="8" :md="4">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-icon" style="background: #409eff">
            <el-icon :size="22"><Van /></el-icon>
          </div>
          <div class="stat-content">
            <div class="stat-value">{{ stats.total }}</div>
            <div class="stat-label">总行程数</div>
          </div>
        </el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="4">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-icon" style="background: #e6a23c">
            <el-icon :size="22"><Clock /></el-icon>
          </div>
          <div class="stat-content">
            <div class="stat-value">{{ stats.inProgress }}</div>
            <div class="stat-label">进行中</div>
          </div>
        </el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="4">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-icon" style="background: #67c23a">
            <el-icon :size="22"><CircleCheck /></el-icon>
          </div>
          <div class="stat-content">
            <div class="stat-value">{{ stats.completed }}</div>
            <div class="stat-label">已完成</div>
          </div>
        </el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="4">
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
      <el-col :xs="12" :sm="8" :md="4">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-icon" style="background: #722ed1">
            <el-icon :size="22"><User /></el-icon>
          </div>
          <div class="stat-content">
            <div class="stat-value">{{ stats.totalPassengers }}</div>
            <div class="stat-label">总乘客数</div>
          </div>
        </el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="4">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-icon" style="background: #13c2c2">
            <el-icon :size="22"><Money /></el-icon>
          </div>
          <div class="stat-content">
            <div class="stat-value">¥{{ stats.totalAmount.toFixed(2) }}</div>
            <div class="stat-label">总金额</div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <div class="page-card">
      <!-- 筛选区 -->
      <el-form :inline="true" :model="filters" class="filter-form" @submit.prevent>
        <el-form-item label="行程单号">
          <el-input v-model="filters.trip_no" placeholder="行程单号" clearable style="width: 180px" @keyup.enter="onSearch" />
        </el-form-item>
        <el-form-item label="关键词">
          <el-input v-model="filters.keyword" placeholder="路线/乘客" clearable style="width: 200px" @keyup.enter="onSearch" />
        </el-form-item>
        <el-form-item label="行程状态">
          <el-select v-model="filters.status" placeholder="全部" clearable style="width: 130px" @change="onSearch">
            <el-option label="待出发" :value="0" />
            <el-option label="进行中" :value="1" />
            <el-option label="已完成" :value="2" />
            <el-option label="已取消" :value="3" />
            <el-option label="异常" :value="4" />
          </el-select>
        </el-form-item>
        <el-form-item label="司机ID">
          <el-input v-model="filters.driver_id" placeholder="司机ID" clearable style="width: 120px" @keyup.enter="onSearch" />
        </el-form-item>
        <el-form-item label="拼车ID">
          <el-input v-model="filters.pinche_id" placeholder="拼车ID" clearable style="width: 120px" @keyup.enter="onSearch" />
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
          <el-button type="primary" :icon="Plus" @click="onCreate">新建行程</el-button>
        </div>
      </div>

      <!-- 表格 -->
      <el-table v-loading="loading" :data="list" border stripe>
        <el-table-column prop="id" label="ID" width="70" fixed="left" />
        <el-table-column prop="trip_no" label="行程单号" width="180" />
        <el-table-column label="路线" min-width="220">
          <template #default="{ row }">
            <div class="route-text">
              <span class="from">{{ row.origin }}</span>
              <el-icon class="arrow"><Right /></el-icon>
              <span class="to">{{ row.destination }}</span>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="司机" width="140">
          <template #default="{ row }">
            <div>{{ row.driver_name || `#${row.driver_id}` }}</div>
            <div class="text-muted">{{ maskPhone(row.driver_phone) }}</div>
          </template>
        </el-table-column>
        <el-table-column label="乘客数" width="90">
          <template #default="{ row }">{{ row.passenger_count }} / {{ row.seats_total }}</template>
        </el-table-column>
        <el-table-column label="金额" width="110">
          <template #default="{ row }">
            <span class="price">¥{{ Number(row.total_amount || 0).toFixed(2) }}</span>
          </template>
        </el-table-column>
        <el-table-column label="出发时间" width="160">
          <template #default="{ row }">{{ formatTime(row.departure_time) }}</template>
        </el-table-column>
        <el-table-column label="到达时间" width="160">
          <template #default="{ row }">{{ formatTime(row.arrival_time) }}</template>
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
            <el-button v-if="row.status === 0" type="success" link size="small" @click="onStart(row)">开始行程</el-button>
            <el-button v-if="row.status === 1" type="warning" link size="small" @click="onComplete(row)">完成行程</el-button>
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
    <el-dialog v-model="detailVisible" title="行程详情" width="800px">
      <el-descriptions v-if="detail" :column="2" border>
        <el-descriptions-item label="ID">{{ detail.id }}</el-descriptions-item>
        <el-descriptions-item label="行程单号">{{ detail.trip_no }}</el-descriptions-item>
        <el-descriptions-item label="出发地" :span="2">{{ detail.origin }}</el-descriptions-item>
        <el-descriptions-item label="目的地" :span="2">{{ detail.destination }}</el-descriptions-item>
        <el-descriptions-item label="司机">{{ detail.driver_name || `#${detail.driver_id}` }}</el-descriptions-item>
        <el-descriptions-item label="车辆">{{ detail.vehicle_info || '-' }}</el-descriptions-item>
        <el-descriptions-item label="乘客数">{{ detail.passenger_count }} / {{ detail.seats_total }}</el-descriptions-item>
        <el-descriptions-item label="金额">
          <span class="price">¥{{ Number(detail.total_amount || 0).toFixed(2) }}</span>
        </el-descriptions-item>
        <el-descriptions-item label="出发时间">{{ formatTime(detail.departure_time) }}</el-descriptions-item>
        <el-descriptions-item label="到达时间">{{ formatTime(detail.arrival_time) }}</el-descriptions-item>
        <el-descriptions-item label="实际出发">{{ formatTime(detail.actual_departure_time) }}</el-descriptions-item>
        <el-descriptions-item label="实际到达">{{ formatTime(detail.actual_arrival_time) }}</el-descriptions-item>
        <el-descriptions-item label="状态">
          <el-tag :type="statusTagType(detail.status)" size="small">{{ statusText(detail.status) }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="拼车ID">{{ detail.pinche_id }}</el-descriptions-item>
        <el-descriptions-item label="距离">{{ detail.distance ? detail.distance + ' km' : '-' }}</el-descriptions-item>
        <el-descriptions-item label="时长">{{ detail.duration ? detail.duration + ' 分钟' : '-' }}</el-descriptions-item>
        <el-descriptions-item v-if="detail.cancel_reason" label="取消原因" :span="2">{{ detail.cancel_reason }}</el-descriptions-item>
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
import request from '@/utils/request'
import {
  Refresh, RefreshLeft, Search, Plus, Right,
  Van, Clock, CircleCheck, CircleClose, User, Money
} from '@element-plus/icons-vue'
import { formatTime } from '@/utils/format'

const loading = ref(false)
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)
const list = ref([])

const stats = reactive({
  total: 0, inProgress: 0, completed: 0, cancelled: 0, totalPassengers: 0, totalAmount: 0
})

const filters = reactive({
  trip_no: '', keyword: '', status: null,
  driver_id: '', pinche_id: '', dateRange: null
})

const statusText = (s) => ({ 0: '待出发', 1: '进行中', 2: '已完成', 3: '已取消', 4: '异常' }[s] || '-')
const statusTagType = (s) => ({ 0: 'info', 1: 'warning', 2: 'success', 3: 'danger', 4: 'danger' }[s] || 'info')

const maskPhone = (p) => {
  if (!p) return '-'
  const s = String(p)
  if (s.length < 7) return s
  return s.slice(0, 3) + '****' + s.slice(-4)
}

const onSearch = () => { page.value = 1; loadList() }
const onReset = () => {
  filters.trip_no = ''; filters.keyword = ''; filters.status = null
  filters.driver_id = ''; filters.pinche_id = ''; filters.dateRange = null
  page.value = 1; loadList()
}

const loadList = async () => {
  loading.value = true
  try {
    const params = {
      page: page.value,
      page_size: pageSize.value,
      trip_no: filters.trip_no || undefined,
      keyword: filters.keyword.trim() || undefined,
      status: filters.status === null || filters.status === '' ? undefined : filters.status,
      driver_id: filters.driver_id || undefined,
      pinche_id: filters.pinche_id || undefined
    }
    if (filters.dateRange && filters.dateRange.length === 2) {
      params.start_date = filters.dateRange[0]
      params.end_date = filters.dateRange[1]
    }
    const res = await request.get('/pinche/trips', { params })
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
    const res = await request.get(`/pinche/trips/${row.id}`)
    if (res.data) detail.value = res.data
  } catch (e) { /* ignore */ }
}

const onStart = async (row) => {
  try {
    await ElMessageBox.confirm(`确认开始行程 "${row.trip_no}"？`, '提示', { type: 'warning' })
    await request.put(`/pinche/admin/trips/${row.id}/status`, { status: 1 })
    ElMessage.success('行程已开始')
    await loadList()
  } catch (e) { /* cancel */ }
}

const onComplete = async (row) => {
  try {
    await ElMessageBox.confirm(`确认完成行程 "${row.trip_no}"？`, '提示', { type: 'warning' })
    await request.put(`/pinche/trips/${row.id}/complete`)
    ElMessage.success('行程已完成')
    await loadList()
  } catch (e) { /* cancel */ }
}

const onCancel = async (row) => {
  try {
    const { value } = await ElMessageBox.prompt('请输入取消原因（可选）', '取消行程', {
      inputType: 'textarea',
      inputPlaceholder: '取消原因（可不填）'
    })
    await request.put(`/pinche/admin/trips/${row.id}/status`, { status: 3, cancel_reason: value || '' })
    ElMessage.success('行程已取消')
    await loadList()
  } catch (e) { /* cancel */ }
}

const onCreate = () => {
  ElMessage.info('新建行程功能开发中，请使用 C 端发起')
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
