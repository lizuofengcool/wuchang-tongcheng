<template>
  <div class="app-container">
    <!-- 物流统计 -->
    <el-row :gutter="16" class="stat-row">
      <el-col :xs="12" :sm="8" :md="6">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value">{{ stats.total }}</div><div class="stat-label">物流单总数</div></div></el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="6">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value text-primary">{{ stats.shipped }}</div><div class="stat-label">运输中</div></div></el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="6">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value text-success">{{ stats.delivered }}</div><div class="stat-label">已签收</div></div></el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="6">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value text-warning">{{ stats.pending }}</div><div class="stat-label">待发货</div></div></el-card>
      </el-col>
    </el-row>

    <div class="page-card">
      <el-form :inline="true" :model="filters" class="filter-form" @submit.prevent>
        <el-form-item label="物流单号">
          <el-input v-model="filters.tracking_no" placeholder="物流单号" clearable style="width: 200px" @keyup.enter="onSearch" />
        </el-form-item>
        <el-form-item label="订单ID">
          <el-input-number v-model="filters.order_id" :controls="false" :min="1" placeholder="订单ID" style="width: 140px" />
        </el-form-item>
        <el-form-item label="店铺ID">
          <el-input-number v-model="filters.shop_id" :controls="false" :min="1" placeholder="店铺ID" style="width: 140px" />
        </el-form-item>
        <el-form-item label="物流公司">
          <el-input v-model="filters.company" placeholder="物流公司" clearable style="width: 160px" @keyup.enter="onSearch" />
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="filters.status" placeholder="全部" clearable style="width: 130px" @change="onSearch">
            <el-option label="待发货" :value="0" />
            <el-option label="已发货" :value="1" />
            <el-option label="运输中" :value="2" />
            <el-option label="已签收" :value="3" />
            <el-option label="已拒收" :value="4" />
            <el-option label="异常" :value="5" />
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
        <el-table-column label="物流单号" min-width="180">
          <template #default="{ row }">
            <el-link type="primary" :underline="false" @click="openDetail(row)">{{ row.tracking_no || '-' }}</el-link>
          </template>
        </el-table-column>
        <el-table-column label="订单ID" width="100" prop="order_id" />
        <el-table-column label="店铺ID" width="90" prop="shop_id" />
        <el-table-column label="物流公司" width="140" prop="company" />
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="statusTagType(row.status)" size="small">{{ statusText(row.status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="发货时间" width="160" prop="shipped_at" sortable="custom">
          <template #default="{ row }">{{ formatTime(row.shipped_at) }}</template>
        </el-table-column>
        <el-table-column label="签收时间" width="160" prop="delivered_at" sortable="custom">
          <template #default="{ row }">{{ formatTime(row.delivered_at) }}</template>
        </el-table-column>
        <el-table-column label="创建时间" width="160" prop="created_at" sortable="custom">
          <template #default="{ row }">{{ formatTime(row.created_at) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="openDetail(row)">详情</el-button>
            <el-button v-if="row.status === 0" type="success" link size="small" @click="onUpdateStatus(row, 1)">标记发货</el-button>
            <el-button v-if="row.status === 1 || row.status === 2" type="success" link size="small" @click="onUpdateStatus(row, 3)">标记签收</el-button>
          </template>
        </el-table-column>
      </el-table>

      <div class="pagination-wrap">
        <el-pagination v-model:current-page="page" v-model:page-size="pageSize" :total="total" :page-sizes="[10, 20, 50, 100]" layout="total, sizes, prev, pager, next, jumper" background @current-change="loadList" @size-change="loadList" />
      </div>
    </div>

    <!-- 详情弹窗 -->
    <el-dialog v-model="detailVisible" title="物流详情" width="780px" destroy-on-close>
      <el-descriptions v-if="detail" :column="2" border>
        <el-descriptions-item label="ID">{{ detail.id }}</el-descriptions-item>
        <el-descriptions-item label="物流单号">{{ detail.tracking_no || '-' }}</el-descriptions-item>
        <el-descriptions-item label="订单ID">{{ detail.order_id }}</el-descriptions-item>
        <el-descriptions-item label="店铺ID">{{ detail.shop_id }}</el-descriptions-item>
        <el-descriptions-item label="物流公司">{{ detail.company || '-' }}</el-descriptions-item>
        <el-descriptions-item label="物流公司代码">{{ detail.company_code || '-' }}</el-descriptions-item>
        <el-descriptions-item label="发货人">{{ detail.sender_name || '-' }}</el-descriptions-item>
        <el-descriptions-item label="发货手机">{{ detail.sender_phone || '-' }}</el-descriptions-item>
        <el-descriptions-item label="发货地址" :span="2">{{ detail.sender_address || '-' }}</el-descriptions-item>
        <el-descriptions-item label="收货人">{{ detail.receiver_name || '-' }}</el-descriptions-item>
        <el-descriptions-item label="收货手机">{{ detail.receiver_phone || '-' }}</el-descriptions-item>
        <el-descriptions-item label="收货地址" :span="2">{{ detail.receiver_address || '-' }}</el-descriptions-item>
        <el-descriptions-item label="重量(kg)">{{ detail.weight || '-' }}</el-descriptions-item>
        <el-descriptions-item label="运费">¥{{ Number(detail.freight || 0).toFixed(2) }}</el-descriptions-item>
        <el-descriptions-item label="状态">
          <el-tag :type="statusTagType(detail.status)" size="small">{{ statusText(detail.status) }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="备注">{{ detail.remark || '-' }}</el-descriptions-item>
        <el-descriptions-item label="发货时间">{{ formatTime(detail.shipped_at) }}</el-descriptions-item>
        <el-descriptions-item label="签收时间">{{ formatTime(detail.delivered_at) }}</el-descriptions-item>
        <el-descriptions-item label="创建时间">{{ formatTime(detail.created_at) }}</el-descriptions-item>
        <el-descriptions-item label="更新时间">{{ formatTime(detail.updated_at) }}</el-descriptions-item>
      </el-descriptions>

      <!-- 物流轨迹 -->
      <div v-if="detail && detail.tracks && detail.tracks.length" class="tracks-section">
        <h4 class="tracks-title">物流轨迹</h4>
        <el-timeline>
          <el-timeline-item
            v-for="(t, idx) in detail.tracks"
            :key="idx"
            :timestamp="formatTime(t.timestamp || t.time || t.created_at)"
            placement="top"
          >
            <div class="track-content">{{ t.content || t.description || t.desc || '-' }}</div>
            <div class="track-location" v-if="t.location">📍 {{ t.location }}</div>
          </el-timeline-item>
        </el-timeline>
      </div>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Refresh, RefreshLeft, Search } from '@element-plus/icons-vue'
import { formatTime } from '@/utils/format'
import {
  getMallLogisticsList, getMallLogisticsDetail,
  getMallLogisticsByOrder, getMallLogisticsByTrackingNo,
  updateMallLogisticsStatus
} from '@/api/mall'

const loading = ref(false)
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)
const list = ref([])
const sortField = ref('created_at')
const sortOrder = ref('descending')

const stats = reactive({ total: 0, shipped: 0, delivered: 0, pending: 0 })

const filters = reactive({ tracking_no: '', order_id: null, shop_id: null, company: '', status: null })

const onSearch = () => { page.value = 1; loadList() }
const onReset = () => {
  filters.tracking_no = ''
  filters.order_id = null
  filters.shop_id = null
  filters.company = ''
  filters.status = null
  page.value = 1
  loadList()
}
const onSortChange = ({ prop, order }) => {
  sortField.value = prop || 'created_at'
  sortOrder.value = order || 'descending'
  loadList()
}

const statusText = (s) => ({ 0: '待发货', 1: '已发货', 2: '运输中', 3: '已签收', 4: '已拒收', 5: '异常' }[s] ?? '-')
const statusTagType = (s) => ({ 0: 'info', 1: 'primary', 2: 'warning', 3: 'success', 4: 'danger', 5: 'danger' }[s] || 'info')

const loadList = async () => {
  loading.value = true
  try {
    let data = {}
    // 优先按物流单号精确查询
    if (filters.tracking_no.trim()) {
      const res = await getMallLogisticsByTrackingNo(filters.tracking_no.trim())
      const item = res.data
      list.value = item ? [item] : []
      total.value = list.value.length
    } else if (filters.order_id) {
      const res = await getMallLogisticsByOrder(filters.order_id)
      data = res.data || {}
      list.value = data.list || data || []
      total.value = data.total || list.value.length
    } else {
      const params = {
        page: page.value,
        page_size: pageSize.value,
        shop_id: filters.shop_id || undefined,
        company: filters.company.trim() || undefined,
        status: filters.status === null || filters.status === '' ? undefined : filters.status,
        sort: sortField.value,
        order: sortOrder.value
      }
      const res = await getMallLogisticsList(params)
      data = res.data || {}
      list.value = data.list || []
      total.value = data.total || 0
    }
    computeStats()
  } catch (e) {
    ElMessage.error('加载物流列表失败')
    list.value = []
    total.value = 0
  } finally {
    loading.value = false
  }
}

const computeStats = () => {
  const total = list.value.length
  const shipped = list.value.filter((r) => r.status === 1 || r.status === 2).length
  const delivered = list.value.filter((r) => r.status === 3).length
  const pending = list.value.filter((r) => r.status === 0).length
  Object.assign(stats, { total, shipped, delivered, pending })
}

const detailVisible = ref(false)
const detail = ref(null)

const openDetail = async (row) => {
  try {
    const res = await getMallLogisticsDetail(row.id)
    detail.value = res.data || null
    detailVisible.value = true
  } catch (e) {
    ElMessage.error('加载详情失败')
  }
}

const onUpdateStatus = async (row, status) => {
  const action = status === 1 ? '标记发货' : (status === 3 ? '标记签收' : '更新状态')
  try {
    await ElMessageBox.confirm(`确认${action}：${row.tracking_no || `#${row.id}`}？`, '提示', { type: 'warning' })
    await updateMallLogisticsStatus(row.id, status)
    ElMessage.success(`${action}成功`)
    await loadList()
  } catch (e) { /* cancel */ }
}

onMounted(() => loadList())
</script>

<style scoped>
.stat-row { margin-bottom: 16px; }
.stat-item { text-align: center; padding: 16px; }
.stat-value { font-size: 20px; font-weight: 600; color: #303133; }
.text-success { color: #67c23a; }
.text-warning { color: #e6a23c; }
.text-primary { color: #409eff; }
.filter-form { margin-bottom: 12px; padding: 12px 16px; background: #fafafa; border-radius: 4px; }
.filter-form :deep(.el-form-item) { margin-bottom: 8px; margin-right: 12px; }
.toolbar { display: flex; justify-content: space-between; margin-bottom: 12px; }
.pagination-wrap { margin-top: 16px; display: flex; justify-content: flex-end; }
.tracks-section { margin-top: 20px; }
.tracks-title { margin: 0 0 12px 0; font-size: 14px; color: #303133; }
.track-content { font-size: 13px; color: #303133; }
.track-location { font-size: 12px; color: #909399; margin-top: 4px; }
</style>
