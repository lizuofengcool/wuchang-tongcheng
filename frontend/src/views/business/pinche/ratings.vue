<template>
  <div class="app-container">
    <!-- 顶部统计卡片 -->
    <el-row :gutter="16" class="stat-row">
      <el-col :xs="12" :sm="8" :md="6">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-icon" style="background: #409eff">
            <el-icon :size="22"><ChatDotRound /></el-icon>
          </div>
          <div class="stat-content">
            <div class="stat-value">{{ stats.total }}</div>
            <div class="stat-label">总评价数</div>
          </div>
        </el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="6">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-icon" style="background: #67c23a">
            <el-icon :size="22"><Star /></el-icon>
          </div>
          <div class="stat-content">
            <div class="stat-value">{{ stats.avgRating.toFixed(1) }} ⭐</div>
            <div class="stat-label">平均评分</div>
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
            <div class="stat-label">待审核</div>
          </div>
        </el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="6">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-icon" style="background: #f56c6c">
            <el-icon :size="22"><Warning /></el-icon>
          </div>
          <div class="stat-content">
            <div class="stat-value">{{ stats.reported }}</div>
            <div class="stat-label">被举报</div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <div class="page-card">
      <!-- 筛选区 -->
      <el-form :inline="true" :model="filters" class="filter-form" @submit.prevent>
        <el-form-item label="拼车ID">
          <el-input v-model="filters.pinche_id" placeholder="拼车ID" clearable style="width: 120px" @keyup.enter="onSearch" />
        </el-form-item>
        <el-form-item label="评分">
          <el-select v-model="filters.rating" placeholder="全部" clearable style="width: 120px" @change="onSearch">
            <el-option label="1星" :value="1" />
            <el-option label="2星" :value="2" />
            <el-option label="3星" :value="3" />
            <el-option label="4星" :value="4" />
            <el-option label="5星" :value="5" />
          </el-select>
        </el-form-item>
        <el-form-item label="评价类型">
          <el-select v-model="filters.rating_type" placeholder="全部" clearable style="width: 130px" @change="onSearch">
            <el-option label="乘客评价司机" value="passenger_to_driver" />
            <el-option label="司机评价乘客" value="driver_to_passenger" />
            <el-option label="评价行程" value="trip" />
          </el-select>
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="filters.status" placeholder="全部" clearable style="width: 130px" @change="onSearch">
            <el-option label="待审核" :value="0" />
            <el-option label="已发布" :value="1" />
            <el-option label="已隐藏" :value="2" />
            <el-option label="已删除" :value="3" />
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
      </div>

      <!-- 表格 -->
      <el-table v-loading="loading" :data="list" border stripe>
        <el-table-column prop="id" label="ID" width="70" fixed="left" />
        <el-table-column prop="pinche_id" label="拼车ID" width="90" />
        <el-table-column label="评分" width="160">
          <template #default="{ row }">
            <el-rate :model-value="row.rating || 0" disabled />
          </template>
        </el-table-column>
        <el-table-column label="评价内容" min-width="240" show-overflow-tooltip>
          <template #default="{ row }">{{ row.content || '-' }}</template>
        </el-table-column>
        <el-table-column label="评价人" width="140">
          <template #default="{ row }">
            <div>{{ row.rater_name || `#${row.rater_id}` }}</div>
            <div class="text-muted">{{ ratingTypeText(row.rating_type) }}</div>
          </template>
        </el-table-column>
        <el-table-column label="被评价人" width="140">
          <template #default="{ row }">
            <div>{{ row.ratee_name || `#${row.ratee_id}` }}</div>
          </template>
        </el-table-column>
        <el-table-column prop="like_count" label="点赞" width="80" />
        <el-table-column prop="reply_count" label="回复" width="80" />
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
            <el-button v-if="row.status === 0" type="success" link size="small" @click="onUpdateStatus(row, 1)">通过</el-button>
            <el-button v-if="row.status === 0 || row.status === 1" type="warning" link size="small" @click="onUpdateStatus(row, 2)">隐藏</el-button>
            <el-button type="danger" link size="small" @click="onDelete(row)">删除</el-button>
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
    <el-dialog v-model="detailVisible" title="评价详情" width="700px">
      <el-descriptions v-if="detail" :column="2" border>
        <el-descriptions-item label="ID">{{ detail.id }}</el-descriptions-item>
        <el-descriptions-item label="拼车ID">{{ detail.pinche_id }}</el-descriptions-item>
        <el-descriptions-item label="行程ID">{{ detail.trip_id || '-' }}</el-descriptions-item>
        <el-descriptions-item label="预订ID">{{ detail.booking_id || '-' }}</el-descriptions-item>
        <el-descriptions-item label="评分" :span="2">
          <el-rate :model-value="detail.rating || 0" disabled />
        </el-descriptions-item>
        <el-descriptions-item label="评价内容" :span="2">{{ detail.content }}</el-descriptions-item>
        <el-descriptions-item label="评价人">{{ detail.rater_name || `#${detail.rater_id}` }}</el-descriptions-item>
        <el-descriptions-item label="被评价人">{{ detail.ratee_name || `#${detail.ratee_id}` }}</el-descriptions-item>
        <el-descriptions-item label="评价类型">{{ ratingTypeText(detail.rating_type) }}</el-descriptions-item>
        <el-descriptions-item label="状态">
          <el-tag :type="statusTagType(detail.status)" size="small">{{ statusText(detail.status) }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="点赞数">{{ detail.like_count }}</el-descriptions-item>
        <el-descriptions-item label="回复数">{{ detail.reply_count }}</el-descriptions-item>
        <el-descriptions-item v-if="detail.tags && detail.tags.length" label="标签" :span="2">
          <el-tag v-for="t in detail.tags" :key="t" size="small" style="margin-right: 4px">{{ t }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item v-if="detail.reply" label="回复" :span="2">{{ detail.reply }}</el-descriptions-item>
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
  Refresh, RefreshLeft, Search,
  ChatDotRound, Star, Clock, Warning
} from '@element-plus/icons-vue'
import { formatTime } from '@/utils/format'

const loading = ref(false)
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)
const list = ref([])

const stats = reactive({ total: 0, avgRating: 0, pending: 0, reported: 0 })

const filters = reactive({
  pinche_id: '', rating: null, rating_type: '',
  status: null, dateRange: null
})

const statusText = (s) => ({ 0: '待审核', 1: '已发布', 2: '已隐藏', 3: '已删除' }[s] || '-')
const statusTagType = (s) => ({ 0: 'warning', 1: 'success', 2: 'info', 3: 'danger' }[s] || 'info')
const ratingTypeText = (t) => ({
  passenger_to_driver: '乘客评价司机',
  driver_to_passenger: '司机评价乘客',
  trip: '评价行程'
}[t] || '-')

const onSearch = () => { page.value = 1; loadList() }
const onReset = () => {
  filters.pinche_id = ''; filters.rating = null; filters.rating_type = ''
  filters.status = null; filters.dateRange = null
  page.value = 1; loadList()
}

const loadList = async () => {
  loading.value = true
  try {
    const params = {
      page: page.value,
      page_size: pageSize.value,
      pinche_id: filters.pinche_id || undefined,
      rating: filters.rating === null || filters.rating === '' ? undefined : filters.rating,
      rating_type: filters.rating_type || undefined,
      status: filters.status === null || filters.status === '' ? undefined : filters.status
    }
    if (filters.dateRange && filters.dateRange.length === 2) {
      params.start_date = filters.dateRange[0]
      params.end_date = filters.dateRange[1]
    }
    const res = await request.get('/pinche/admin/ratings', { params })
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
    const res = await request.get(`/pinche/ratings/${row.id}`)
    if (res.data) detail.value = res.data
  } catch (e) { /* ignore */ }
}

const onUpdateStatus = async (row, status) => {
  try {
    const label = statusText(status)
    await ElMessageBox.confirm(`确认将评价设为「${label}」？`, '提示', { type: 'warning' })
    await request.put(`/pinche/admin/ratings/${row.id}/status`, { status })
    ElMessage.success('状态更新成功')
    await loadList()
  } catch (e) { /* cancel */ }
}

const onDelete = async (row) => {
  try {
    await ElMessageBox.confirm('确认删除该评价？删除后不可恢复！', '危险操作', {
      type: 'error', confirmButtonText: '确认删除', cancelButtonText: '取消'
    })
    await request.delete(`/pinche/ratings/${row.id}`)
    ElMessage.success('删除成功')
    await loadList()
  } catch (e) { /* cancel */ }
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

.text-muted { color: #909399; font-size: 12px; }
.pagination-wrap { margin-top: 16px; display: flex; justify-content: flex-end; }
</style>
