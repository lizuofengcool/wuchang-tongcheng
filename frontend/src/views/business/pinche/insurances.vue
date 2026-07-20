<template>
  <div class="app-container">
    <!-- 顶部统计卡片 -->
    <el-row :gutter="16" class="stat-row">
      <el-col :xs="12" :sm="8" :md="6">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-icon" style="background: #409eff">
            <el-icon :size="22"><Umbrella /></el-icon>
          </div>
          <div class="stat-content">
            <div class="stat-value">{{ stats.total }}</div>
            <div class="stat-label">总保单数</div>
          </div>
        </el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="6">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-icon" style="background: #67c23a">
            <el-icon :size="22"><Money /></el-icon>
          </div>
          <div class="stat-content">
            <div class="stat-value">¥{{ stats.totalAmount.toFixed(2) }}</div>
            <div class="stat-label">总保额</div>
          </div>
        </el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="6">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-icon" style="background: #e6a23c">
            <el-icon :size="22"><Clock /></el-icon>
          </div>
          <div class="stat-content">
            <div class="stat-value">{{ stats.active }}</div>
            <div class="stat-label">保障中</div>
          </div>
        </el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="6">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-icon" style="background: #f56c6c">
            <el-icon :size="22"><Warning /></el-icon>
          </div>
          <div class="stat-content">
            <div class="stat-value">{{ stats.claiming }}</div>
            <div class="stat-label">理赔中</div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <div class="page-card">
      <!-- 筛选区 -->
      <el-form :inline="true" :model="filters" class="filter-form" @submit.prevent>
        <el-form-item label="保单号">
          <el-input v-model="filters.policy_no" placeholder="保单号" clearable style="width: 180px" @keyup.enter="onSearch" />
        </el-form-item>
        <el-form-item label="预订ID">
          <el-input v-model="filters.booking_id" placeholder="预订ID" clearable style="width: 120px" @keyup.enter="onSearch" />
        </el-form-item>
        <el-form-item label="保险类型">
          <el-select v-model="filters.insurance_type" placeholder="全部" clearable style="width: 140px" @change="onSearch">
            <el-option label="人身意外险" value="accident" />
            <el-option label="车辆险" value="vehicle" />
            <el-option label="拼车责任险" value="liability" />
            <el-option label="行程险" value="trip" />
          </el-select>
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="filters.status" placeholder="全部" clearable style="width: 130px" @change="onSearch">
            <el-option label="待生效" :value="0" />
            <el-option label="保障中" :value="1" />
            <el-option label="已结束" :value="2" />
            <el-option label="理赔中" :value="3" />
            <el-option label="已理赔" :value="4" />
            <el-option label="已取消" :value="5" />
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
        <el-table-column prop="policy_no" label="保单号" width="180" />
        <el-table-column prop="booking_id" label="预订ID" width="90" />
        <el-table-column label="保险类型" width="120">
          <template #default="{ row }">
            <el-tag :type="typeTagType(row.insurance_type)" size="small">{{ typeText(row.insurance_type) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="保额" width="120">
          <template #default="{ row }">
            <span class="price">¥{{ Number(row.coverage_amount || 0).toFixed(2) }}</span>
          </template>
        </el-table-column>
        <el-table-column label="保费" width="110">
          <template #default="{ row }">¥{{ Number(row.premium || 0).toFixed(2) }}</template>
        </el-table-column>
        <el-table-column label="投保人" width="140">
          <template #default="{ row }">
            <div>{{ row.insured_name || `#${row.insured_id}` }}</div>
          </template>
        </el-table-column>
        <el-table-column label="保障期间" width="220">
          <template #default="{ row }">
            <div>{{ formatTime(row.start_time, 'YYYY-MM-DD HH:mm') }}</div>
            <div class="text-muted">至 {{ formatTime(row.end_time, 'YYYY-MM-DD HH:mm') }}</div>
          </template>
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
            <el-button v-if="row.status === 1" type="warning" link size="small" @click="onClaim(row)">理赔</el-button>
            <el-button v-if="row.status === 0" type="danger" link size="small" @click="onCancel(row)">取消</el-button>
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
    <el-dialog v-model="detailVisible" title="保单详情" width="700px">
      <el-descriptions v-if="detail" :column="2" border>
        <el-descriptions-item label="ID">{{ detail.id }}</el-descriptions-item>
        <el-descriptions-item label="保单号">{{ detail.policy_no }}</el-descriptions-item>
        <el-descriptions-item label="预订ID">{{ detail.booking_id }}</el-descriptions-item>
        <el-descriptions-item label="行程ID">{{ detail.trip_id || '-' }}</el-descriptions-item>
        <el-descriptions-item label="拼车ID">{{ detail.pinche_id || '-' }}</el-descriptions-item>
        <el-descriptions-item label="保险类型">
          <el-tag :type="typeTagType(detail.insurance_type)" size="small">{{ typeText(detail.insurance_type) }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="保额">¥{{ Number(detail.coverage_amount || 0).toFixed(2) }}</el-descriptions-item>
        <el-descriptions-item label="保费">¥{{ Number(detail.premium || 0).toFixed(2) }}</el-descriptions-item>
        <el-descriptions-item label="投保人">{{ detail.insured_name || `#${detail.insured_id}` }}</el-descriptions-item>
        <el-descriptions-item label="保险公司">{{ detail.insurer || '-' }}</el-descriptions-item>
        <el-descriptions-item label="保障开始">{{ formatTime(detail.start_time) }}</el-descriptions-item>
        <el-descriptions-item label="保障结束">{{ formatTime(detail.end_time) }}</el-descriptions-item>
        <el-descriptions-item label="状态">
          <el-tag :type="statusTagType(detail.status)" size="small">{{ statusText(detail.status) }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="理赔金额">{{ detail.claim_amount ? '¥' + Number(detail.claim_amount).toFixed(2) : '-' }}</el-descriptions-item>
        <el-descriptions-item v-if="detail.claim_reason" label="理赔原因" :span="2">{{ detail.claim_reason }}</el-descriptions-item>
        <el-descriptions-item v-if="detail.claim_time" label="理赔时间">{{ formatTime(detail.claim_time) }}</el-descriptions-item>
        <el-descriptions-item label="创建时间">{{ formatTime(detail.created_at) }}</el-descriptions-item>
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
  Umbrella, Money, Clock, Warning
} from '@element-plus/icons-vue'
import { formatTime } from '@/utils/format'

const loading = ref(false)
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)
const list = ref([])

const stats = reactive({ total: 0, totalAmount: 0, active: 0, claiming: 0 })

const filters = reactive({
  policy_no: '', booking_id: '', insurance_type: '',
  status: null, dateRange: null
})

const statusText = (s) => ({ 0: '待生效', 1: '保障中', 2: '已结束', 3: '理赔中', 4: '已理赔', 5: '已取消' }[s] || '-')
const statusTagType = (s) => ({ 0: 'info', 1: 'success', 2: 'info', 3: 'warning', 4: 'primary', 5: 'danger' }[s] || 'info')
const typeText = (t) => ({ accident: '人身意外险', vehicle: '车辆险', liability: '拼车责任险', trip: '行程险' }[t] || '-')
const typeTagType = (t) => ({ accident: 'danger', vehicle: 'warning', liability: 'primary', trip: 'success' }[t] || 'info')

const onSearch = () => { page.value = 1; loadList() }
const onReset = () => {
  filters.policy_no = ''; filters.booking_id = ''; filters.insurance_type = ''
  filters.status = null; filters.dateRange = null
  page.value = 1; loadList()
}

const loadList = async () => {
  loading.value = true
  try {
    const params = {
      page: page.value,
      page_size: pageSize.value,
      policy_no: filters.policy_no || undefined,
      booking_id: filters.booking_id || undefined,
      insurance_type: filters.insurance_type || undefined,
      status: filters.status === null || filters.status === '' ? undefined : filters.status
    }
    if (filters.dateRange && filters.dateRange.length === 2) {
      params.start_date = filters.dateRange[0]
      params.end_date = filters.dateRange[1]
    }
    const res = await request.get('/pinche/admin/insurances', { params })
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
    const res = await request.get(`/pinche/insurances/${row.id}`)
    if (res.data) detail.value = res.data
  } catch (e) { /* ignore */ }
}

const onClaim = async (row) => {
  try {
    const { value } = await ElMessageBox.prompt('请输入理赔原因', '理赔', {
      inputType: 'textarea',
      inputPlaceholder: '理赔原因',
      inputValidator: (v) => !!v || '请输入理赔原因'
    })
    await request.put(`/pinche/admin/insurances/${row.id}/status`, { status: 3, claim_reason: value })
    ElMessage.success('已进入理赔流程')
    await loadList()
  } catch (e) { /* cancel */ }
}

const onCancel = async (row) => {
  try {
    await ElMessageBox.confirm(`确认取消保单 "${row.policy_no}"？`, '提示', { type: 'warning' })
    await request.put(`/pinche/admin/insurances/${row.id}/status`, { status: 5 })
    ElMessage.success('保单已取消')
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

.price { color: #f56c6c; font-weight: 600; }
.text-muted { color: #909399; font-size: 12px; }
.pagination-wrap { margin-top: 16px; display: flex; justify-content: flex-end; }
</style>
