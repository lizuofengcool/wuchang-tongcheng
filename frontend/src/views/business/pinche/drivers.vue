<template>
  <div class="app-container">
    <!-- 顶部统计卡片 -->
    <el-row :gutter="16" class="stat-row">
      <el-col :xs="12" :sm="8" :md="6">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-icon" style="background: #409eff">
            <el-icon :size="22"><User /></el-icon>
          </div>
          <div class="stat-content">
            <div class="stat-value">{{ stats.total }}</div>
            <div class="stat-label">总车主数</div>
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
          <div class="stat-icon" style="background: #67c23a">
            <el-icon :size="22"><CircleCheck /></el-icon>
          </div>
          <div class="stat-content">
            <div class="stat-value">{{ stats.approved }}</div>
            <div class="stat-label">已认证</div>
          </div>
        </el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="6">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-icon" style="background: #f56c6c">
            <el-icon :size="22"><CircleClose /></el-icon>
          </div>
          <div class="stat-content">
            <div class="stat-value">{{ stats.rejected }}</div>
            <div class="stat-label">已拒绝</div>
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
            placeholder="姓名/手机号/车牌号"
            clearable
            style="width: 220px"
            @keyup.enter="onSearch"
          />
        </el-form-item>
        <el-form-item label="审核状态">
          <el-select v-model="filters.audit_status" placeholder="全部" clearable style="width: 120px" @change="onSearch">
            <el-option label="待审核" :value="0" />
            <el-option label="已通过" :value="1" />
            <el-option label="已拒绝" :value="2" />
          </el-select>
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="filters.status" placeholder="全部" clearable style="width: 120px" @change="onSearch">
            <el-option label="正常" :value="1" />
            <el-option label="禁用" :value="0" />
            <el-option label="封禁" :value="2" />
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
          <el-button type="primary" :icon="Plus" @click="openCreate">新建车主</el-button>
        </div>
      </div>

      <!-- 表格 -->
      <el-table v-loading="loading" :data="list" border stripe>
        <el-table-column prop="id" label="ID" width="70" fixed="left" />
        <el-table-column label="车主" min-width="160">
          <template #default="{ row }">
            <div class="driver-cell">
              <div class="driver-name">{{ row.real_name || row.user_name || `#${row.user_id}` }}</div>
              <div class="text-muted">{{ maskPhone(row.phone) }}</div>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="身份证" width="180">
          <template #default="{ row }">{{ maskIDCard(row.id_card) }}</template>
        </el-table-column>
        <el-table-column label="驾驶证" width="120">
          <template #default="{ row }">
            <span v-if="row.license_no">{{ row.license_no }}</span>
            <span v-else class="text-muted">-</span>
          </template>
        </el-table-column>
        <el-table-column label="车牌号" width="120">
          <template #default="{ row }">
            <span v-if="row.plate_no" class="plate">{{ row.plate_no }}</span>
            <span v-else class="text-muted">-</span>
          </template>
        </el-table-column>
        <el-table-column prop="vehicle_count" label="车辆数" width="80" />
        <el-table-column prop="trip_count" label="行程数" width="80" />
        <el-table-column prop="rating_avg" label="评分" width="90">
          <template #default="{ row }">
            <span v-if="row.rating_avg">{{ Number(row.rating_avg).toFixed(1) }} ⭐</span>
            <span v-else class="text-muted">-</span>
          </template>
        </el-table-column>
        <el-table-column label="审核" width="100">
          <template #default="{ row }">
            <el-tag :type="auditTagType(row.audit_status)" size="small">{{ auditText(row.audit_status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="90">
          <template #default="{ row }">
            <el-tag :type="statusTagType(row.status)" size="small" effect="plain">{{ statusText(row.status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="注册时间" width="160">
          <template #default="{ row }">{{ formatTime(row.created_at) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="240" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="openDetail(row)">详情</el-button>
            <el-button v-if="row.audit_status === 0" type="success" link size="small" @click="openReview(row, 1)">通过</el-button>
            <el-button v-if="row.audit_status === 0" type="danger" link size="small" @click="openReview(row, 2)">拒绝</el-button>
            <el-dropdown trigger="click" @command="(cmd) => handleStatusCommand(row, cmd)">
              <el-button type="warning" link size="small">
                更多<el-icon class="el-icon--right"><ArrowDown /></el-icon>
              </el-button>
              <template #dropdown>
                <el-dropdown-menu>
                  <el-dropdown-item :command="1" :disabled="row.status === 1">启用</el-dropdown-item>
                  <el-dropdown-item :command="0" :disabled="row.status === 0">禁用</el-dropdown-item>
                  <el-dropdown-item :command="2" :disabled="row.status === 2">封禁</el-dropdown-item>
                </el-dropdown-menu>
              </template>
            </el-dropdown>
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
    <el-dialog v-model="detailVisible" title="车主详情" width="800px">
      <el-descriptions v-if="detail" :column="2" border>
        <el-descriptions-item label="ID">{{ detail.id }}</el-descriptions-item>
        <el-descriptions-item label="用户ID">{{ detail.user_id }}</el-descriptions-item>
        <el-descriptions-item label="真实姓名">{{ detail.real_name || '-' }}</el-descriptions-item>
        <el-descriptions-item label="手机号">{{ maskPhone(detail.phone) }}</el-descriptions-item>
        <el-descriptions-item label="身份证">{{ maskIDCard(detail.id_card) }}</el-descriptions-item>
        <el-descriptions-item label="驾驶证号">{{ detail.license_no || '-' }}</el-descriptions-item>
        <el-descriptions-item label="车牌号">{{ detail.plate_no || '-' }}</el-descriptions-item>
        <el-descriptions-item label="车辆类型">{{ detail.vehicle_type || '-' }}</el-descriptions-item>
        <el-descriptions-item label="车辆数">{{ detail.vehicle_count }}</el-descriptions-item>
        <el-descriptions-item label="行程数">{{ detail.trip_count }}</el-descriptions-item>
        <el-descriptions-item label="评分">{{ detail.rating_avg ? Number(detail.rating_avg).toFixed(1) : '-' }}</el-descriptions-item>
        <el-descriptions-item label="累计接单">{{ detail.total_orders || 0 }}</el-descriptions-item>
        <el-descriptions-item label="累计里程">{{ detail.total_distance ? detail.total_distance + ' km' : '-' }}</el-descriptions-item>
        <el-descriptions-item label="累计收入">¥{{ Number(detail.total_income || 0).toFixed(2) }}</el-descriptions-item>
        <el-descriptions-item label="审核状态">
          <el-tag :type="auditTagType(detail.audit_status)" size="small">{{ auditText(detail.audit_status) }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="状态">
          <el-tag :type="statusTagType(detail.status)" size="small" effect="plain">{{ statusText(detail.status) }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item v-if="detail.audit_reason" label="审核原因" :span="2">{{ detail.audit_reason }}</el-descriptions-item>
        <el-descriptions-item v-if="detail.location" label="最近位置" :span="2">
          {{ detail.location.address || '-' }} ({{ detail.location.lng }}, {{ detail.location.lat }})
        </el-descriptions-item>
        <el-descriptions-item label="注册时间">{{ formatTime(detail.created_at) }}</el-descriptions-item>
        <el-descriptions-item label="更新时间">{{ formatTime(detail.updated_at) }}</el-descriptions-item>
      </el-descriptions>
      <template #footer><el-button @click="detailVisible = false">关闭</el-button></template>
    </el-dialog>

    <!-- 审核弹窗 -->
    <el-dialog v-model="reviewVisible" :title="reviewForm.audit_status === 1 ? '审核通过' : '审核拒绝'" width="500px">
      <el-form :model="reviewForm" label-width="100px">
        <el-form-item label="车主">{{ reviewForm.driver_name }}</el-form-item>
        <el-form-item label="审核结果">
          <el-radio-group v-model="reviewForm.audit_status">
            <el-radio :value="1">通过</el-radio>
            <el-radio :value="2">拒绝</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item v-if="reviewForm.audit_status === 2" label="拒绝原因">
          <el-input v-model="reviewForm.audit_reason" type="textarea" :rows="3" placeholder="拒绝原因" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="reviewVisible = false">取消</el-button>
        <el-button type="primary" :loading="reviewLoading" @click="onReview">确认</el-button>
      </template>
    </el-dialog>

    <!-- 新建弹窗 -->
    <el-dialog v-model="formVisible" title="新建车主" width="640px">
      <el-form ref="formRef" :model="form" :rules="rules" label-width="100px">
        <el-form-item label="用户ID" prop="user_id">
          <el-input-number v-model="form.user_id" :min="1" style="width: 100%" />
        </el-form-item>
        <el-form-item label="真实姓名" prop="real_name">
          <el-input v-model="form.real_name" maxlength="32" />
        </el-form-item>
        <el-form-item label="手机号" prop="phone">
          <el-input v-model="form.phone" maxlength="20" />
        </el-form-item>
        <el-form-item label="身份证" prop="id_card">
          <el-input v-model="form.id_card" maxlength="18" />
        </el-form-item>
        <el-form-item label="驾驶证号">
          <el-input v-model="form.license_no" maxlength="32" />
        </el-form-item>
        <el-form-item label="车牌号">
          <el-input v-model="form.plate_no" maxlength="20" />
        </el-form-item>
        <el-form-item label="车辆类型">
          <el-select v-model="form.vehicle_type" style="width: 100%">
            <el-option label="轿车" value="sedan" />
            <el-option label="SUV" value="suv" />
            <el-option label="MPV" value="mpv" />
            <el-option label="商务车" value="commercial" />
            <el-option label="其他" value="other" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="formVisible = false">取消</el-button>
        <el-button type="primary" :loading="formLoading" @click="onSubmit">确认</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import request from '@/utils/request'
import {
  Refresh, RefreshLeft, Search, Plus, ArrowDown,
  User, Clock, CircleCheck, CircleClose
} from '@element-plus/icons-vue'
import { formatTime } from '@/utils/format'

const loading = ref(false)
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)
const list = ref([])

const stats = reactive({ total: 0, pending: 0, approved: 0, rejected: 0 })

const filters = reactive({
  keyword: '', audit_status: null, status: null, dateRange: null
})

const auditText = (s) => ({ 0: '待审核', 1: '已通过', 2: '已拒绝' }[s] || '-')
const auditTagType = (s) => ({ 0: 'warning', 1: 'success', 2: 'danger' }[s] || 'info')
const statusText = (s) => ({ 0: '禁用', 1: '正常', 2: '封禁' }[s] || '-')
const statusTagType = (s) => ({ 0: 'info', 1: 'success', 2: 'danger' }[s] || 'info')

const maskPhone = (p) => {
  if (!p) return '-'
  const s = String(p)
  if (s.length < 7) return s
  return s.slice(0, 3) + '****' + s.slice(-4)
}
const maskIDCard = (c) => {
  if (!c) return '-'
  const s = String(c)
  if (s.length < 8) return s
  return s.slice(0, 4) + '********' + s.slice(-4)
}

const onSearch = () => { page.value = 1; loadList() }
const onReset = () => {
  filters.keyword = ''; filters.audit_status = null; filters.status = null; filters.dateRange = null
  page.value = 1; loadList()
}

const loadList = async () => {
  loading.value = true
  try {
    const params = {
      page: page.value,
      page_size: pageSize.value,
      keyword: filters.keyword.trim() || undefined,
      audit_status: filters.audit_status === null || filters.audit_status === '' ? undefined : filters.audit_status,
      status: filters.status === null || filters.status === '' ? undefined : filters.status
    }
    if (filters.dateRange && filters.dateRange.length === 2) {
      params.start_date = filters.dateRange[0]
      params.end_date = filters.dateRange[1]
    }
    const res = await request.get('/pinche/admin/drivers', { params })
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
    const res = await request.get(`/pinche/drivers/${row.id}`)
    if (res.data) detail.value = res.data
  } catch (e) { /* ignore */ }
}

// 审核
const reviewVisible = ref(false)
const reviewLoading = ref(false)
const reviewForm = reactive({ id: null, driver_name: '', audit_status: 1, audit_reason: '' })

const openReview = (row, status) => {
  reviewForm.id = row.id
  reviewForm.driver_name = row.real_name || row.user_name || `#${row.user_id}`
  reviewForm.audit_status = status
  reviewForm.audit_reason = ''
  reviewVisible.value = true
}

const onReview = async () => {
  reviewLoading.value = true
  try {
    await request.put(`/pinche/admin/drivers/${reviewForm.id}/review`, {
      audit_status: reviewForm.audit_status,
      audit_reason: reviewForm.audit_reason
    })
    ElMessage.success('审核完成')
    reviewVisible.value = false
    await loadList()
  } catch (e) { /* fail */ } finally {
    reviewLoading.value = false
  }
}

const handleStatusCommand = async (row, cmd) => {
  try {
    const label = statusText(cmd)
    await ElMessageBox.confirm(`确定将车主设为「${label}」吗？`, '提示', { type: 'warning' })
    await request.put(`/pinche/admin/drivers/${row.id}/status`, { status: cmd })
    ElMessage.success('状态更新成功')
    await loadList()
  } catch (e) { /* cancel */ }
}

// 新建
const formVisible = ref(false)
const formLoading = ref(false)
const formRef = ref(null)
const form = reactive({
  user_id: null, real_name: '', phone: '', id_card: '',
  license_no: '', plate_no: '', vehicle_type: 'sedan'
})
const rules = {
  user_id: [{ required: true, message: '请输入用户ID', trigger: 'blur' }],
  real_name: [{ required: true, message: '请输入真实姓名', trigger: 'blur' }],
  phone: [{ required: true, message: '请输入手机号', trigger: 'blur' }],
  id_card: [{ required: true, message: '请输入身份证', trigger: 'blur' }]
}

const openCreate = () => {
  Object.assign(form, {
    user_id: null, real_name: '', phone: '', id_card: '',
    license_no: '', plate_no: '', vehicle_type: 'sedan'
  })
  formVisible.value = true
}

const onSubmit = async () => {
  try {
    await formRef.value.validate()
    formLoading.value = true
    await request.post('/pinche/drivers', form)
    ElMessage.success('创建成功')
    formVisible.value = false
    await loadList()
  } catch (e) { /* fail */ } finally {
    formLoading.value = false
  }
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

.driver-cell { font-size: 13px; }
.driver-name { color: #303133; }
.text-muted { color: #909399; font-size: 12px; }
.plate {
  display: inline-block; padding: 2px 8px;
  background: #f0f9ff; color: #303133;
  border: 1px solid #d0e8ff; border-radius: 4px;
  font-weight: 500;
}
.pagination-wrap { margin-top: 16px; display: flex; justify-content: flex-end; }
</style>
