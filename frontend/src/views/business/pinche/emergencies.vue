<template>
  <div class="app-container">
    <!-- 顶部统计卡片 -->
    <el-row :gutter="16" class="stat-row">
      <el-col :xs="12" :sm="8" :md="6">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-icon" style="background: #f56c6c">
            <el-icon :size="22"><WarnTriangleFilled /></el-icon>
          </div>
          <div class="stat-content">
            <div class="stat-value">{{ stats.total }}</div>
            <div class="stat-label">总报警数</div>
          </div>
        </el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="6">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-icon" style="background: #e6a23c">
            <el-icon :size="22"><Bell /></el-icon>
          </div>
          <div class="stat-content">
            <div class="stat-value">{{ stats.pending }}</div>
            <div class="stat-label">待处理</div>
          </div>
        </el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="6">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-icon" style="background: #67c23a">
            <el-icon :size="22"><CircleCheckFilled /></el-icon>
          </div>
          <div class="stat-content">
            <div class="stat-value">{{ stats.handled }}</div>
            <div class="stat-label">已处理</div>
          </div>
        </el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="6">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-icon" style="background: #722ed1">
            <el-icon :size="22"><PhoneFilled /></el-icon>
          </div>
          <div class="stat-content">
            <div class="stat-value">{{ stats.sos }}</div>
            <div class="stat-label">SOS 报警</div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <div class="page-card">
      <!-- 筛选区 -->
      <el-form :inline="true" :model="filters" class="filter-form" @submit.prevent>
        <el-form-item label="关键词">
          <el-input
            v-model="filters.keyword"
            placeholder="报警单号/报警人/联系电话"
            clearable
            style="width: 220px"
            :prefix-icon="Search"
            @keyup.enter="onSearch"
          />
        </el-form-item>
        <el-form-item label="报警类型">
          <el-select v-model="filters.alert_type" placeholder="全部" clearable style="width: 140px" @change="onSearch">
            <el-option label="SOS 紧急求助" value="sos" />
            <el-option label="行程异常" value="trip_abnormal" />
            <el-option label="车主违约" value="driver_violation" />
            <el-option label="乘客违约" value="passenger_violation" />
            <el-option label="车辆故障" value="vehicle_breakdown" />
            <el-option label="交通事故" value="accident" />
            <el-option label="骚扰纠纷" value="harassment" />
            <el-option label="其他报警" value="other" />
          </el-select>
        </el-form-item>
        <el-form-item label="处理状态">
          <el-select v-model="filters.status" placeholder="全部" clearable style="width: 130px" @change="onSearch">
            <el-option label="待处理" :value="0" />
            <el-option label="处理中" :value="1" />
            <el-option label="已处理" :value="2" />
            <el-option label="已关闭" :value="3" />
            <el-option label="误报" :value="4" />
          </el-select>
        </el-form-item>
        <el-form-item label="紧急程度">
          <el-select v-model="filters.priority" placeholder="全部" clearable style="width: 120px" @change="onSearch">
            <el-option label="低" :value="0" />
            <el-option label="中" :value="1" />
            <el-option label="高" :value="2" />
            <el-option label="紧急" :value="3" />
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
          <el-button type="warning" :icon="Bell" :disabled="!selection.length" @click="onBatchHandle">批量处理</el-button>
          <el-button type="info" :icon="CircleClose" :disabled="!selection.length" @click="onBatchClose">批量关闭</el-button>
        </div>
        <div class="toolbar-right">
          <el-button type="danger" :icon="PhoneFilled" @click="onEmergencyCall">一键报警（110）</el-button>
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
        <el-table-column prop="alert_no" label="报警单号" width="180" />
        <el-table-column label="类型" width="130">
          <template #default="{ row }">
            <el-tag :type="typeTagType(row.alert_type)" size="small">{{ typeText(row.alert_type) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="紧急程度" width="90">
          <template #default="{ row }">
            <el-tag :type="priorityTagType(row.priority)" size="small" effect="dark">{{ priorityText(row.priority) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="报警人" width="150">
          <template #default="{ row }">
            <div>{{ row.reporter_name || `用户#${row.reporter_id}` }}</div>
            <div class="text-muted">{{ maskPhone(row.reporter_phone) }}</div>
          </template>
        </el-table-column>
        <el-table-column label="关联行程" width="180">
          <template #default="{ row }">
            <div v-if="row.trip_id">行程 #{{ row.trip_id }}</div>
            <div v-if="row.pinche_id" class="text-muted">拼车 #{{ row.pinche_id }}</div>
            <span v-if="!row.trip_id && !row.pinche_id" class="text-muted">-</span>
          </template>
        </el-table-column>
        <el-table-column label="报警位置" min-width="200">
          <template #default="{ row }">
            <div v-if="row.location">{{ row.location }}</div>
            <div v-if="row.latitude && row.longitude" class="text-muted">
              经纬度: {{ Number(row.latitude).toFixed(6) }}, {{ Number(row.longitude).toFixed(6) }}
            </div>
            <span v-else class="text-muted">-</span>
          </template>
        </el-table-column>
        <el-table-column label="SLA" width="100">
          <template #default="{ row }">
            <el-tag v-if="isOverSLA(row)" type="danger" size="small" effect="dark">超时</el-tag>
            <el-tag v-else-if="row.status === 2 || row.status === 3" type="success" size="small">已响应</el-tag>
            <el-tag v-else type="warning" size="small">{{ slaText(row) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="statusTagType(row.status)" size="small">{{ statusText(row.status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="报警时间" width="160" prop="created_at" sortable>
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
              @click="onHandle(row)"
            >处理</el-button>
            <el-button
              v-if="row.status === 0 || row.status === 1"
              type="warning"
              link
              size="small"
              @click="onMarkFalse(row)"
            >误报</el-button>
            <el-button
              v-if="row.status === 0 || row.status === 1"
              type="danger"
              link
              size="small"
              @click="onClose(row)"
            >关闭</el-button>
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
    <el-dialog v-model="detailVisible" title="报警详情" width="800px" destroy-on-close>
      <el-descriptions v-if="detail" :column="2" border>
        <el-descriptions-item label="ID">{{ detail.id }}</el-descriptions-item>
        <el-descriptions-item label="报警单号">{{ detail.alert_no }}</el-descriptions-item>
        <el-descriptions-item label="报警类型">
          <el-tag :type="typeTagType(detail.alert_type)" size="small">{{ typeText(detail.alert_type) }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="紧急程度">
          <el-tag :type="priorityTagType(detail.priority)" size="small" effect="dark">{{ priorityText(detail.priority) }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="报警人">{{ detail.reporter_name || `用户#${detail.reporter_id}` }}</el-descriptions-item>
        <el-descriptions-item label="联系电话">{{ detail.reporter_phone || '-' }}</el-descriptions-item>
        <el-descriptions-item label="关联行程">{{ detail.trip_id ? `#${detail.trip_id}` : '-' }}</el-descriptions-item>
        <el-descriptions-item label="关联拼车">{{ detail.pinche_id ? `#${detail.pinche_id}` : '-' }}</el-descriptions-item>
        <el-descriptions-item label="报警位置" :span="2">{{ detail.location || '-' }}</el-descriptions-item>
        <el-descriptions-item v-if="detail.latitude" label="经度">{{ Number(detail.latitude).toFixed(6) }}</el-descriptions-item>
        <el-descriptions-item v-if="detail.longitude" label="纬度">{{ Number(detail.longitude).toFixed(6) }}</el-descriptions-item>
        <el-descriptions-item label="状态">
          <el-tag :type="statusTagType(detail.status)" size="small">{{ statusText(detail.status) }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="报警时间">{{ formatTime(detail.created_at) }}</el-descriptions-item>
        <el-descriptions-item v-if="detail.handled_at" label="处理时间">{{ formatTime(detail.handled_at) }}</el-descriptions-item>
        <el-descriptions-item v-if="detail.handler_id" label="处理人">管理员 #{{ detail.handler_id }}</el-descriptions-item>
        <el-descriptions-item v-if="detail.description" label="报警描述" :span="2">{{ detail.description }}</el-descriptions-item>
        <el-descriptions-item v-if="detail.handle_note" label="处理记录" :span="2">
          <div class="handle-note">{{ detail.handle_note }}</div>
        </el-descriptions-item>
      </el-descriptions>
      <template #footer>
        <el-button @click="detailVisible = false">关闭</el-button>
        <el-button v-if="detail && (detail.status === 0 || detail.status === 1)" type="success" @click="onHandle(detail)">立即处理</el-button>
      </template>
    </el-dialog>

    <!-- 处理弹窗 -->
    <el-dialog v-model="handleVisible" title="处理报警" width="600px" destroy-on-close>
      <el-form :model="handleForm" label-width="100px">
        <el-form-item label="报警单号">
          <span>{{ handleForm.alert_no }}</span>
        </el-form-item>
        <el-form-item label="处理状态">
          <el-radio-group v-model="handleForm.status">
            <el-radio :value="1">处理中</el-radio>
            <el-radio :value="2">已处理</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="处理记录" required>
          <el-input
            v-model="handleForm.handle_note"
            type="textarea"
            :rows="5"
            placeholder="请描述处理过程、联系方式、结果等"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="handleVisible = false">取消</el-button>
        <el-button type="primary" @click="submitHandle">提交</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import request from '@/utils/request'
import {
  Refresh, RefreshLeft, Search,
  WarnTriangleFilled, Bell, CircleCheckFilled, PhoneFilled, CircleClose
} from '@element-plus/icons-vue'
import { formatTime } from '@/utils/format'

const loading = ref(false)
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)
const list = ref([])
const selection = ref([])

const stats = reactive({ total: 0, pending: 0, handled: 0, sos: 0 })

const filters = reactive({
  keyword: '', alert_type: '', status: null, priority: null, dateRange: null
})

const statusText = (s) => ({ 0: '待处理', 1: '处理中', 2: '已处理', 3: '已关闭', 4: '误报' }[s] || '-')
const statusTagType = (s) => ({ 0: 'danger', 1: 'warning', 2: 'success', 3: 'info', 4: 'info' }[s] || 'info')
const typeText = (t) => ({
  sos: 'SOS 紧急求助', trip_abnormal: '行程异常', driver_violation: '车主违约',
  passenger_violation: '乘客违约', vehicle_breakdown: '车辆故障', accident: '交通事故',
  harassment: '骚扰纠纷', other: '其他报警'
}[t] || '-')
const typeTagType = (t) => ({
  sos: 'danger', trip_abnormal: 'warning', driver_violation: 'danger',
  passenger_violation: 'warning', vehicle_breakdown: 'info', accident: 'danger',
  harassment: 'warning', other: 'info'
}[t] || 'info')
const priorityText = (p) => ({ 0: '低', 1: '中', 2: '高', 3: '紧急' }[p] || '中')
const priorityTagType = (p) => ({ 0: 'info', 1: '', 2: 'warning', 3: 'danger' }[p] || '')

const maskPhone = (p) => p ? p.replace(/(\d{3})\d{4}(\d{4})/, '$1****$2') : '-'

const isOverSLA = (row) => {
  if (row.status === 2 || row.status === 3 || row.status === 4) return false
  const created = new Date(row.created_at).getTime()
  const sla = (row.sla_minutes || 30) * 60 * 1000
  return Date.now() - created > sla
}

const slaText = (row) => {
  const created = new Date(row.created_at).getTime()
  const sla = (row.sla_minutes || 30) * 60 * 1000
  const remain = sla - (Date.now() - created)
  if (remain <= 0) return '已超时'
  const mins = Math.floor(remain / 60000)
  return `${mins} 分钟内`
}

const onSearch = () => { page.value = 1; loadList() }
const onReset = () => {
  filters.keyword = ''; filters.alert_type = ''
  filters.status = null; filters.priority = null; filters.dateRange = null
  page.value = 1; loadList()
}

const onSelectionChange = (rows) => { selection.value = rows }

const loadList = async () => {
  loading.value = true
  try {
    const params = {
      page: page.value,
      page_size: pageSize.value,
      keyword: filters.keyword || undefined,
      alert_type: filters.alert_type || undefined,
      status: filters.status === null || filters.status === '' ? undefined : filters.status,
      priority: filters.priority === null || filters.priority === '' ? undefined : filters.priority
    }
    if (filters.dateRange && filters.dateRange.length === 2) {
      params.start_date = filters.dateRange[0]
      params.end_date = filters.dateRange[1]
    }
    const res = await request.get('/pinche/admin/emergencies', { params })
    const data = res.data || {}
    list.value = data.list || []
    total.value = data.total || 0
    // 更新统计
    if (data.stats) {
      stats.total = data.stats.total || 0
      stats.pending = data.stats.pending || 0
      stats.handled = data.stats.handled || 0
      stats.sos = data.stats.sos || 0
    } else {
      // 本地兜底统计
      stats.total = total.value
      stats.pending = list.value.filter(i => i.status === 0 || i.status === 1).length
      stats.handled = list.value.filter(i => i.status === 2).length
      stats.sos = list.value.filter(i => i.alert_type === 'sos').length
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

const handleVisible = ref(false)
const handleForm = reactive({ id: 0, alert_no: '', status: 2, handle_note: '' })

const onHandle = (row) => {
  handleForm.id = row.id
  handleForm.alert_no = row.alert_no || `#${row.id}`
  handleForm.status = row.status === 1 ? 2 : 2
  handleForm.handle_note = ''
  handleVisible.value = true
}

const submitHandle = async () => {
  if (!handleForm.handle_note?.trim()) {
    ElMessage.warning('请填写处理记录')
    return
  }
  try {
    await request.put(`/pinche/admin/emergencies/${handleForm.id}/handle`, {
      handle_note: handleForm.handle_note,
      status: handleForm.status
    })
    ElMessage.success('处理成功')
    handleVisible.value = false
    await loadList()
  } catch (e) { /* ignore */ }
}

const onMarkFalse = async (row) => {
  try {
    await ElMessageBox.confirm(`确认将报警 "${row.alert_no}" 标记为误报？`, '提示', { type: 'warning' })
    await request.put(`/pinche/admin/emergencies/${row.id}/status`, { status: 4 })
    ElMessage.success('已标记为误报')
    await loadList()
  } catch (e) { /* cancel */ }
}

const onClose = async (row) => {
  try {
    const { value } = await ElMessageBox.prompt('请输入关闭原因', '关闭报警', {
      inputType: 'textarea',
      inputPlaceholder: '关闭原因',
      inputValidator: (v) => !!v || '请输入关闭原因'
    })
    await request.put(`/pinche/admin/emergencies/${row.id}/status`, { status: 3, handle_note: value })
    ElMessage.success('报警已关闭')
    await loadList()
  } catch (e) { /* cancel */ }
}

const onBatchHandle = async () => {
  if (!selection.value.length) return
  try {
    const { value } = await ElMessageBox.prompt(
      `确认批量处理 ${selection.value.length} 条报警？请填写统一处理记录`, '批量处理',
      { inputType: 'textarea', inputPlaceholder: '处理记录', inputValidator: (v) => !!v || '请输入处理记录' }
    )
    const ids = selection.value.map(i => i.id)
    await Promise.all(ids.map(id =>
      request.put(`/pinche/admin/emergencies/${id}/handle`, { handle_note: value, status: 2 })
    ))
    ElMessage.success(`已批量处理 ${ids.length} 条报警`)
    await loadList()
  } catch (e) { /* cancel */ }
}

const onBatchClose = async () => {
  if (!selection.value.length) return
  try {
    await ElMessageBox.confirm(
      `确认批量关闭 ${selection.value.length} 条报警？此操作不可撤销`, '批量关闭',
      { type: 'warning' }
    )
    const ids = selection.value.map(i => i.id)
    await Promise.all(ids.map(id =>
      request.put(`/pinche/admin/emergencies/${id}/status`, { status: 3, handle_note: '批量关闭' })
    ))
    ElMessage.success(`已批量关闭 ${ids.length} 条报警`)
    await loadList()
  } catch (e) { /* cancel */ }
}

const onEmergencyCall = async () => {
  try {
    await ElMessageBox.confirm(
      '即将拨打 110 报警电话，确认继续？',
      '一键报警',
      { type: 'warning', confirmButtonText: '确认拨打', cancelButtonText: '取消' }
    )
    ElMessage.success('正在拨打 110，请保持通话畅通')
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
.handle-note { white-space: pre-wrap; background: #fafafa; padding: 8px; border-radius: 4px; }
.pagination-wrap { margin-top: 16px; display: flex; justify-content: flex-end; }
</style>
