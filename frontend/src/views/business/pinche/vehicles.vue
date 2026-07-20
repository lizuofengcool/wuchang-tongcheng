<template>
  <div class="app-container">
    <!-- 顶部统计卡片 -->
    <el-row :gutter="16" class="stat-row">
      <el-col :xs="12" :sm="8" :md="6">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-icon" style="background: #409eff">
            <el-icon :size="22"><Van /></el-icon>
          </div>
          <div class="stat-content">
            <div class="stat-value">{{ stats.total }}</div>
            <div class="stat-label">总车辆数</div>
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
          <div class="stat-icon" style="background: #722ed1">
            <el-icon :size="22"><Star /></el-icon>
          </div>
          <div class="stat-content">
            <div class="stat-value">{{ stats.defaultCount }}</div>
            <div class="stat-label">默认车辆</div>
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
            placeholder="车牌号/品牌/车主"
            clearable
            style="width: 220px"
            @keyup.enter="onSearch"
          />
        </el-form-item>
        <el-form-item label="车型">
          <el-select v-model="filters.vehicle_type" placeholder="全部" clearable style="width: 130px" @change="onSearch">
            <el-option label="轿车" value="sedan" />
            <el-option label="SUV" value="suv" />
            <el-option label="MPV" value="mpv" />
            <el-option label="商务车" value="commercial" />
            <el-option label="其他" value="other" />
          </el-select>
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
          <el-button type="primary" :icon="Plus" @click="openCreate">新建车辆</el-button>
        </div>
      </div>

      <!-- 表格 -->
      <el-table v-loading="loading" :data="list" border stripe>
        <el-table-column prop="id" label="ID" width="70" fixed="left" />
        <el-table-column label="车牌号" width="120">
          <template #default="{ row }">
            <span class="plate">{{ row.plate_no }}</span>
            <el-tag v-if="row.is_default" type="warning" size="small" effect="dark" style="margin-left: 4px">默认</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="车型" width="100">
          <template #default="{ row }">
            <el-tag :type="typeTagType(row.vehicle_type)" size="small" effect="plain">{{ typeText(row.vehicle_type) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="品牌型号" min-width="160">
          <template #default="{ row }">
            <div>{{ row.brand || '-' }} {{ row.model || '' }}</div>
            <div class="text-muted">{{ row.color || '-' }}</div>
          </template>
        </el-table-column>
        <el-table-column label="座位数" width="80">
          <template #default="{ row }">{{ row.seats }}</template>
        </el-table-column>
        <el-table-column label="车主" width="140">
          <template #default="{ row }">
            <div>{{ row.driver_name || `#${row.driver_id}` }}</div>
            <div class="text-muted">{{ maskPhone(row.driver_phone) }}</div>
          </template>
        </el-table-column>
        <el-table-column prop="trip_count" label="行程数" width="80" />
        <el-table-column label="审核" width="100">
          <template #default="{ row }">
            <el-tag :type="auditTagType(row.audit_status)" size="small">{{ auditText(row.audit_status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="90">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'info'" size="small" effect="plain">{{ row.status === 1 ? '正常' : '禁用' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="注册时间" width="160">
          <template #default="{ row }">{{ formatTime(row.created_at) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="240" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="openDetail(row)">详情</el-button>
            <el-button type="warning" link size="small" @click="openEdit(row)">编辑</el-button>
            <el-button v-if="row.audit_status === 0" type="success" link size="small" @click="openReview(row, 1)">通过</el-button>
            <el-button v-if="row.audit_status === 0" type="danger" link size="small" @click="openReview(row, 2)">拒绝</el-button>
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
    <el-dialog v-model="detailVisible" title="车辆详情" width="700px">
      <el-descriptions v-if="detail" :column="2" border>
        <el-descriptions-item label="ID">{{ detail.id }}</el-descriptions-item>
        <el-descriptions-item label="车主ID">{{ detail.driver_id }}</el-descriptions-item>
        <el-descriptions-item label="车牌号" :span="2">
          <span class="plate">{{ detail.plate_no }}</span>
          <el-tag v-if="detail.is_default" type="warning" size="small" effect="dark" style="margin-left: 4px">默认</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="车型">
          <el-tag :type="typeTagType(detail.vehicle_type)" size="small" effect="plain">{{ typeText(detail.vehicle_type) }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="品牌型号">{{ detail.brand }} {{ detail.model }}</el-descriptions-item>
        <el-descriptions-item label="颜色">{{ detail.color || '-' }}</el-descriptions-item>
        <el-descriptions-item label="座位数">{{ detail.seats }}</el-descriptions-item>
        <el-descriptions-item label="行驶证号">{{ detail.driving_license_no || '-' }}</el-descriptions-item>
        <el-descriptions-item label="车辆识别码">{{ detail.vin || '-' }}</el-descriptions-item>
        <el-descriptions-item label="注册日期">{{ detail.register_date || '-' }}</el-descriptions-item>
        <el-descriptions-item label="保险到期">{{ formatTime(detail.insurance_expiry) }}</el-descriptions-item>
        <el-descriptions-item label="年检到期">{{ formatTime(detail.inspection_expiry) }}</el-descriptions-item>
        <el-descriptions-item label="行程数">{{ detail.trip_count }}</el-descriptions-item>
        <el-descriptions-item label="累计里程">{{ detail.total_distance ? detail.total_distance + ' km' : '-' }}</el-descriptions-item>
        <el-descriptions-item label="审核状态">
          <el-tag :type="auditTagType(detail.audit_status)" size="small">{{ auditText(detail.audit_status) }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="状态">
          <el-tag :type="detail.status === 1 ? 'success' : 'info'" size="small" effect="plain">{{ detail.status === 1 ? '正常' : '禁用' }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item v-if="detail.audit_reason" label="审核原因" :span="2">{{ detail.audit_reason }}</el-descriptions-item>
        <el-descriptions-item label="创建时间">{{ formatTime(detail.created_at) }}</el-descriptions-item>
        <el-descriptions-item label="更新时间">{{ formatTime(detail.updated_at) }}</el-descriptions-item>
      </el-descriptions>
      <template #footer><el-button @click="detailVisible = false">关闭</el-button></template>
    </el-dialog>

    <!-- 审核/编辑/新建弹窗 -->
    <el-dialog v-model="formVisible" :title="formTitle" width="640px">
      <el-form ref="formRef" :model="form" :rules="rules" label-width="100px">
        <el-form-item label="车主ID" prop="driver_id">
          <el-input-number v-model="form.driver_id" :min="1" style="width: 100%" />
        </el-form-item>
        <el-form-item label="车牌号" prop="plate_no">
          <el-input v-model="form.plate_no" maxlength="20" />
        </el-form-item>
        <el-form-item label="车型" prop="vehicle_type">
          <el-select v-model="form.vehicle_type" style="width: 100%">
            <el-option label="轿车" value="sedan" />
            <el-option label="SUV" value="suv" />
            <el-option label="MPV" value="mpv" />
            <el-option label="商务车" value="commercial" />
            <el-option label="其他" value="other" />
          </el-select>
        </el-form-item>
        <el-form-item label="品牌">
          <el-input v-model="form.brand" maxlength="32" />
        </el-form-item>
        <el-form-item label="型号">
          <el-input v-model="form.model" maxlength="32" />
        </el-form-item>
        <el-form-item label="颜色">
          <el-input v-model="form.color" maxlength="16" />
        </el-form-item>
        <el-form-item label="座位数" prop="seats">
          <el-input-number v-model="form.seats" :min="1" :max="60" />
        </el-form-item>
        <el-form-item label="行驶证号">
          <el-input v-model="form.driving_license_no" maxlength="32" />
        </el-form-item>
        <el-form-item label="车辆识别码">
          <el-input v-model="form.vin" maxlength="32" />
        </el-form-item>
        <el-form-item label="状态">
          <el-switch v-model="form.status" :active-value="1" :inactive-value="0" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="formVisible = false">取消</el-button>
        <el-button type="primary" :loading="formLoading" @click="onSubmit">确认</el-button>
      </template>
    </el-dialog>

    <!-- 审核弹窗 -->
    <el-dialog v-model="reviewVisible" title="车辆审核" width="500px">
      <el-form :model="reviewForm" label-width="100px">
        <el-form-item label="车辆">{{ reviewForm.vehicle_name }}</el-form-item>
        <el-form-item label="审核结果">
          <el-radio-group v-model="reviewForm.audit_status">
            <el-radio :value="1">通过</el-radio>
            <el-radio :value="2">拒绝</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item v-if="reviewForm.audit_status === 2" label="拒绝原因">
          <el-input v-model="reviewForm.audit_reason" type="textarea" :rows="3" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="reviewVisible = false">取消</el-button>
        <el-button type="primary" :loading="reviewLoading" @click="onReview">确认</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import request from '@/utils/request'
import {
  Refresh, RefreshLeft, Search, Plus,
  Van, Clock, CircleCheck, Star
} from '@element-plus/icons-vue'
import { formatTime } from '@/utils/format'

const loading = ref(false)
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)
const list = ref([])

const stats = reactive({ total: 0, pending: 0, approved: 0, defaultCount: 0 })

const filters = reactive({
  keyword: '', vehicle_type: '', audit_status: null, status: null, dateRange: null
})

const auditText = (s) => ({ 0: '待审核', 1: '已通过', 2: '已拒绝' }[s] || '-')
const auditTagType = (s) => ({ 0: 'warning', 1: 'success', 2: 'danger' }[s] || 'info')
const typeText = (t) => ({ sedan: '轿车', suv: 'SUV', mpv: 'MPV', commercial: '商务车', other: '其他' }[t] || '-')
const typeTagType = (t) => ({ sedan: 'primary', suv: 'success', mpv: 'warning', commercial: 'info', other: 'info' }[t] || 'info')

const maskPhone = (p) => {
  if (!p) return '-'
  const s = String(p)
  if (s.length < 7) return s
  return s.slice(0, 3) + '****' + s.slice(-4)
}

const onSearch = () => { page.value = 1; loadList() }
const onReset = () => {
  filters.keyword = ''; filters.vehicle_type = ''
  filters.audit_status = null; filters.status = null; filters.dateRange = null
  page.value = 1; loadList()
}

const loadList = async () => {
  loading.value = true
  try {
    const params = {
      page: page.value,
      page_size: pageSize.value,
      keyword: filters.keyword.trim() || undefined,
      vehicle_type: filters.vehicle_type || undefined,
      audit_status: filters.audit_status === null || filters.audit_status === '' ? undefined : filters.audit_status,
      status: filters.status === null || filters.status === '' ? undefined : filters.status
    }
    if (filters.dateRange && filters.dateRange.length === 2) {
      params.start_date = filters.dateRange[0]
      params.end_date = filters.dateRange[1]
    }
    const res = await request.get('/pinche/admin/vehicles', { params })
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
    const res = await request.get(`/pinche/vehicles/${row.id}`)
    if (res.data) detail.value = res.data
  } catch (e) { /* ignore */ }
}

// 新建/编辑
const formVisible = ref(false)
const formLoading = ref(false)
const formRef = ref(null)
const isEdit = ref(false)
const formTitle = computed(() => isEdit.value ? '编辑车辆' : '新建车辆')
const form = reactive({
  id: null, driver_id: null, plate_no: '', vehicle_type: 'sedan',
  brand: '', model: '', color: '', seats: 5,
  driving_license_no: '', vin: '', status: 1
})
const rules = {
  driver_id: [{ required: true, message: '请输入车主ID', trigger: 'blur' }],
  plate_no: [{ required: true, message: '请输入车牌号', trigger: 'blur' }],
  vehicle_type: [{ required: true, message: '请选择车型', trigger: 'change' }],
  seats: [{ required: true, message: '请输入座位数', trigger: 'blur' }]
}

const openCreate = () => {
  isEdit.value = false
  Object.assign(form, {
    id: null, driver_id: null, plate_no: '', vehicle_type: 'sedan',
    brand: '', model: '', color: '', seats: 5,
    driving_license_no: '', vin: '', status: 1
  })
  formVisible.value = true
}

const openEdit = (row) => {
  isEdit.value = true
  Object.assign(form, row)
  formVisible.value = true
}

const onSubmit = async () => {
  try {
    await formRef.value.validate()
    formLoading.value = true
    if (isEdit.value) {
      await request.put(`/pinche/vehicles/${form.id}`, form)
    } else {
      await request.post('/pinche/vehicles', form)
    }
    ElMessage.success(isEdit.value ? '更新成功' : '创建成功')
    formVisible.value = false
    await loadList()
  } catch (e) { /* fail */ } finally {
    formLoading.value = false
  }
}

const onDelete = async (row) => {
  try {
    await ElMessageBox.confirm(`确认删除车辆 "${row.plate_no}"？`, '提示', { type: 'warning' })
    await request.delete(`/pinche/vehicles/${row.id}`)
    ElMessage.success('删除成功')
    await loadList()
  } catch (e) { /* cancel */ }
}

// 审核
const reviewVisible = ref(false)
const reviewLoading = ref(false)
const reviewForm = reactive({ id: null, vehicle_name: '', audit_status: 1, audit_reason: '' })

const openReview = (row, status) => {
  reviewForm.id = row.id
  reviewForm.vehicle_name = `${row.brand || ''} ${row.model || ''} ${row.plate_no}`
  reviewForm.audit_status = status
  reviewForm.audit_reason = ''
  reviewVisible.value = true
}

const onReview = async () => {
  reviewLoading.value = true
  try {
    await request.put(`/pinche/admin/vehicles/${reviewForm.id}/review`, {
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

.plate {
  display: inline-block; padding: 2px 8px;
  background: #f0f9ff; color: #303133;
  border: 1px solid #d0e8ff; border-radius: 4px;
  font-weight: 500;
}
.text-muted { color: #909399; font-size: 12px; }
.pagination-wrap { margin-top: 16px; display: flex; justify-content: flex-end; }
</style>
