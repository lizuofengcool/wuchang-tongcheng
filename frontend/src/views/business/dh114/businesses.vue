<template>
  <div class="app-container">
    <!-- 统计卡片 -->
    <el-row :gutter="16" class="stat-row">
      <el-col :xs="12" :sm="8" :md="6">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value">{{ stats.total }}</div><div class="stat-label">总资料数</div></div></el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="6">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value text-success">{{ stats.verified }}</div><div class="stat-label">已认证</div></div></el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="6">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value text-warning">{{ stats.pending }}</div><div class="stat-label">待审核</div></div></el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="6">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value text-danger">{{ stats.failed }}</div><div class="stat-label">认证失败</div></div></el-card>
      </el-col>
    </el-row>

    <div class="page-card">
      <el-form :inline="true" :model="filters" class="filter-form" @submit.prevent>
        <el-form-item label="商户ID">
          <el-input v-model="filters.dh114_id" placeholder="商户ID" clearable style="width: 140px" @keyup.enter="onSearch" />
        </el-form-item>
        <el-form-item label="关键词">
          <el-input v-model="filters.keyword" placeholder="法人/执照号" clearable style="width: 200px" @keyup.enter="onSearch" />
        </el-form-item>
        <el-form-item label="认证状态">
          <el-select v-model="filters.verification_status" placeholder="全部" clearable style="width: 140px" @change="onSearch">
            <el-option label="未认证" :value="0" />
            <el-option label="待审核" :value="1" />
            <el-option label="已认证" :value="2" />
            <el-option label="认证失败" :value="3" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" :icon="Search" @click="onSearch">搜索</el-button>
          <el-button :icon="RefreshLeft" @click="onReset">重置</el-button>
        </el-form-item>
      </el-form>

      <div class="toolbar">
        <el-button :icon="Refresh" @click="loadList">刷新</el-button>
        <el-button type="primary" :icon="Plus" @click="openCreate">新建资料</el-button>
      </div>

      <el-table v-loading="loading" :data="list" border stripe>
        <el-table-column prop="id" label="ID" width="70" />
        <el-table-column prop="dh114_id" label="商户ID" width="90" />
        <el-table-column label="法人" width="120">
          <template #default="{ row }">{{ row.legal_person || '-' }}</template>
        </el-table-column>
        <el-table-column label="执照号" width="180">
          <template #default="{ row }">{{ row.license_no || '-' }}</template>
        </el-table-column>
        <el-table-column label="执照图片" width="90">
          <template #default="{ row }">
            <el-image v-if="row.license_image" :src="row.license_image" fit="cover" class="thumb" preview-teleported :preview-src-list="[row.license_image]" />
            <span v-else class="text-muted">-</span>
          </template>
        </el-table-column>
        <el-table-column label="营业面积" width="100">
          <template #default="{ row }">{{ row.area ? row.area + '㎡' : '-' }}</template>
        </el-table-column>
        <el-table-column label="员工数" width="80">
          <template #default="{ row }">{{ row.employee_count || '-' }}</template>
        </el-table-column>
        <el-table-column label="认证状态" width="100">
          <template #default="{ row }">
            <el-tag :type="verifyTagType(row.verification_status)" size="small">{{ verifyText(row.verification_status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="描述" min-width="180" show-overflow-tooltip>
          <template #default="{ row }">{{ row.description || '-' }}</template>
        </el-table-column>
        <el-table-column label="更新时间" width="160">
          <template #default="{ row }">{{ formatTime(row.updated_at) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="openEdit(row)">编辑</el-button>
            <el-button v-if="row.verification_status !== 2" type="success" link size="small" @click="onVerify(row, 2)">通过</el-button>
            <el-button v-if="row.verification_status !== 3" type="danger" link size="small" @click="onVerify(row, 3)">拒绝</el-button>
            <el-button type="danger" link size="small" @click="onDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>

      <div class="pagination-wrap">
        <el-pagination
          v-model:current-page="page"
          v-model:page-size="pageSize"
          :total="total"
          :page-sizes="[10, 20, 50]"
          layout="total, sizes, prev, pager, next, jumper"
          background
          @current-change="loadList"
          @size-change="loadList"
        />
      </div>
    </div>

    <!-- 营业时间面板 -->
    <el-card class="section-card">
      <template #header>
        <div class="card-header-flex">
          <span class="section-title">营业时间管理（{{ hoursList.length }}）</span>
          <el-button type="primary" link size="small" @click="openHoursDialog">配置营业时间</el-button>
        </div>
      </template>
      <el-table :data="hoursList" border size="small">
        <el-table-column prop="dh114_id" label="商户ID" width="90" />
        <el-table-column label="星期" width="100">
          <template #default="{ row }">{{ weekText(row.day_of_week) }}</template>
        </el-table-column>
        <el-table-column prop="open_time" label="开门" width="100" />
        <el-table-column prop="close_time" label="关门" width="100" />
        <el-table-column label="状态" width="80">
          <template #default="{ row }">
            <el-tag :type="row.is_closed ? 'danger' : 'success'" size="small">{{ row.is_closed ? '休息' : '营业' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="remark" label="备注" min-width="160" />
      </el-table>
    </el-card>

    <!-- 表单弹窗 -->
    <el-dialog v-model="formVisible" :title="formTitle" width="640px" destroy-on-close>
      <el-form ref="formRef" :model="form" :rules="rules" label-width="100px">
        <el-form-item label="商户ID" prop="dh114_id">
          <el-input-number v-model="form.dh114_id" :min="1" :controls="false" style="width: 100%" />
        </el-form-item>
        <el-form-item label="法人">
          <el-input v-model="form.legal_person" maxlength="64" />
        </el-form-item>
        <el-form-item label="执照号">
          <el-input v-model="form.license_no" maxlength="64" />
        </el-form-item>
        <el-form-item label="执照图片">
          <el-input v-model="form.license_image" placeholder="图片 URL" />
        </el-form-item>
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="营业面积">
              <el-input-number v-model="form.area" :min="0" :controls="false" style="width: 100%" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="员工数">
              <el-input-number v-model="form.employee_count" :min="0" :controls="false" style="width: 100%" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-form-item label="描述">
          <el-input v-model="form.description" type="textarea" :rows="3" maxlength="500" show-word-limit />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="formVisible = false">取消</el-button>
        <el-button type="primary" :loading="formLoading" @click="onSubmit">确认</el-button>
      </template>
    </el-dialog>

    <!-- 营业时间弹窗 -->
    <el-dialog v-model="hoursVisible" title="配置营业时间" width="720px" destroy-on-close>
      <el-form label-width="100px">
        <el-form-item label="商户ID">
          <el-input-number v-model="hoursForm.dh114_id" :min="1" :controls="false" style="width: 200px" />
        </el-form-item>
      </el-form>
      <el-table :data="hoursForm.items" border size="small">
        <el-table-column label="星期" width="120">
          <template #default="{ row }">{{ weekText(row.day_of_week) }}</template>
        </el-table-column>
        <el-table-column label="开门" width="160">
          <template #default="{ row }">
            <el-time-picker v-model="row.open_time" value-format="HH:mm:ss" format="HH:mm:ss" style="width: 100%" />
          </template>
        </el-table-column>
        <el-table-column label="关门" width="160">
          <template #default="{ row }">
            <el-time-picker v-model="row.close_time" value-format="HH:mm:ss" format="HH:mm:ss" style="width: 100%" />
          </template>
        </el-table-column>
        <el-table-column label="休息" width="80">
          <template #default="{ row }">
            <el-switch v-model="row.is_closed" :active-value="1" :inactive-value="0" />
          </template>
        </el-table-column>
        <el-table-column label="备注" min-width="160">
          <template #default="{ row }">
            <el-input v-model="row.remark" size="small" />
          </template>
        </el-table-column>
      </el-table>
      <template #footer>
        <el-button @click="hoursVisible = false">取消</el-button>
        <el-button type="primary" :loading="hoursLoading" @click="onSaveHours">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Refresh, RefreshLeft, Search, Plus } from '@element-plus/icons-vue'
import request from '@/utils/request'
import { formatTime } from '@/utils/format'

const loading = ref(false)
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)
const list = ref([])

const stats = reactive({ total: 0, verified: 0, pending: 0, failed: 0 })

const filters = reactive({ dh114_id: '', keyword: '', verification_status: null })

const onSearch = () => { page.value = 1; loadList() }
const onReset = () => {
  filters.dh114_id = ''
  filters.keyword = ''
  filters.verification_status = null
  page.value = 1
  loadList()
}

const verifyText = (s) => ({ 0: '未认证', 1: '待审核', 2: '已认证', 3: '认证失败' }[s] || '-')
const verifyTagType = (s) => ({ 0: 'info', 1: 'warning', 2: 'success', 3: 'danger' }[s] || 'info')
const weekText = (d) => ({ 1: '周一', 2: '周二', 3: '周三', 4: '周四', 5: '周五', 6: '周六', 0: '周日' }[d] || '-')

const loadList = async () => {
  loading.value = true
  try {
    const params = {
      page: page.value,
      page_size: pageSize.value,
      dh114_id: filters.dh114_id || undefined,
      keyword: filters.keyword || undefined,
      verification_status: filters.verification_status === null || filters.verification_status === '' ? undefined : filters.verification_status
    }
    const res = await request.get('/dh114/businesses', { params })
    const data = res.data || {}
    list.value = data.list || []
    total.value = data.total || 0
    calcStats()
  } catch (e) {
    list.value = []
    total.value = 0
  } finally {
    loading.value = false
  }
}

const calcStats = () => {
  const total = list.value.length
  const verified = list.value.filter((r) => r.verification_status === 2).length
  const pending = list.value.filter((r) => r.verification_status === 1).length
  const failed = list.value.filter((r) => r.verification_status === 3).length
  Object.assign(stats, { total, verified, pending, failed })
}

// ===== 表单 =====
const formVisible = ref(false)
const formLoading = ref(false)
const formRef = ref(null)
const isEdit = ref(false)
const formTitle = computed(() => isEdit.value ? '编辑资料' : '新建资料')
const form = reactive({
  id: null, dh114_id: undefined, legal_person: '', license_no: '', license_image: '',
  area: undefined, employee_count: undefined, description: ''
})
const rules = {
  dh114_id: [{ required: true, message: '请输入商户ID', trigger: 'blur' }]
}

const resetForm = () => {
  Object.assign(form, {
    id: null, dh114_id: undefined, legal_person: '', license_no: '', license_image: '',
    area: undefined, employee_count: undefined, description: ''
  })
}

const openCreate = () => {
  isEdit.value = false
  resetForm()
  formVisible.value = true
}

const openEdit = (row) => {
  isEdit.value = true
  Object.assign(form, {
    id: row.id, dh114_id: row.dh114_id, legal_person: row.legal_person || '',
    license_no: row.license_no || '', license_image: row.license_image || '',
    area: row.area, employee_count: row.employee_count, description: row.description || ''
  })
  formVisible.value = true
}

const onSubmit = async () => {
  try {
    await formRef.value.validate()
    formLoading.value = true
    const payload = { ...form }
    if (isEdit.value) {
      await request.put(`/dh114/businesses/${form.id}`, payload)
    } else {
      await request.post('/dh114/businesses', payload)
    }
    ElMessage.success(isEdit.value ? '更新成功' : '创建成功')
    formVisible.value = false
    await loadList()
  } catch (e) {
    // 校验或接口失败
  } finally {
    formLoading.value = false
  }
}

const onDelete = async (row) => {
  try {
    await ElMessageBox.confirm(`确认删除该资料？`, '提示', { type: 'warning' })
    await request.delete(`/dh114/businesses/${row.id}`)
    ElMessage.success('删除成功')
    await loadList()
  } catch (e) { /* cancel */ }
}

const onVerify = async (row, status) => {
  try {
    const label = verifyText(status)
    await ElMessageBox.confirm(`确认将该资料设为「${label}」？`, '提示', { type: 'warning' })
    await request.put(`/dh114/${row.dh114_id}/business/verification-status`, { verification_status: status })
    ElMessage.success('操作成功')
    await loadList()
  } catch (e) { /* cancel */ }
}

// ===== 营业时间 =====
const hoursList = ref([])
const hoursVisible = ref(false)
const hoursLoading = ref(false)
const hoursForm = reactive({
  dh114_id: undefined,
  items: Array.from({ length: 7 }, (_, i) => ({
    day_of_week: i === 6 ? 0 : i + 1,
    open_time: '09:00:00',
    close_time: '18:00:00',
    is_closed: 0,
    remark: ''
  }))
})

const loadHours = async () => {
  try {
    const res = await request.get('/dh114/business-hours', { params: { page: 1, page_size: 100 } })
    const data = res.data || {}
    hoursList.value = data.list || data || []
  } catch (e) {
    hoursList.value = []
  }
}

const openHoursDialog = () => {
  hoursForm.dh114_id = undefined
  hoursForm.items = Array.from({ length: 7 }, (_, i) => ({
    day_of_week: i === 6 ? 0 : i + 1,
    open_time: '09:00:00',
    close_time: '18:00:00',
    is_closed: 0,
    remark: ''
  }))
  hoursVisible.value = true
}

const onSaveHours = async () => {
  if (!hoursForm.dh114_id) {
    ElMessage.warning('请输入商户ID')
    return
  }
  hoursLoading.value = true
  try {
    await request.put(`/dh114/${hoursForm.dh114_id}/business-hours`, { items: hoursForm.items })
    ElMessage.success('保存成功')
    hoursVisible.value = false
    await loadHours()
  } catch (e) {
    // 失败已提示
  } finally {
    hoursLoading.value = false
  }
}

onMounted(async () => {
  await Promise.all([loadList(), loadHours()])
})
</script>

<style scoped>
.stat-row { margin-bottom: 16px; }
.stat-item { text-align: center; padding: 16px; }
.stat-value { font-size: 22px; font-weight: 600; color: #409eff; }
.text-success { color: #67c23a; }
.text-warning { color: #e6a23c; }
.text-danger { color: #f56c6c; }
.filter-form { margin-bottom: 12px; padding: 12px 16px; background: #fafafa; border-radius: 4px; }
.filter-form :deep(.el-form-item) { margin-bottom: 8px; margin-right: 12px; }
.toolbar { display: flex; justify-content: space-between; margin-bottom: 12px; }
.text-muted { color: #909399; }
.thumb { width: 50px; height: 50px; border-radius: 4px; border: 1px solid #ebeef5; }
.pagination-wrap { margin-top: 16px; display: flex; justify-content: flex-end; }
.section-card { margin-top: 16px; }
.section-title { font-weight: 600; color: #303133; }
.card-header-flex { display: flex; justify-content: space-between; align-items: center; }
</style>
