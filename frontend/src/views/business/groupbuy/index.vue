<template>
  <div class="app-container">
    <!-- 顶部统计 -->
    <el-row :gutter="16" class="stat-row">
      <el-col :xs="12" :sm="8" :md="6">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-icon" style="background: #409eff"><el-icon :size="22"><Present /></el-icon></div>
          <div class="stat-content"><div class="stat-value">{{ stats.total }}</div><div class="stat-label">团购总数</div></div>
        </el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="6">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-icon" style="background: #67c23a"><el-icon :size="22"><CircleCheck /></el-icon></div>
          <div class="stat-content"><div class="stat-value">{{ stats.online }}</div><div class="stat-label">上架中</div></div>
        </el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="6">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-icon" style="background: #e6a23c"><el-icon :size="22"><Clock /></el-icon></div>
          <div class="stat-content"><div class="stat-value">{{ stats.pendingAudit }}</div><div class="stat-label">待审核</div></div>
        </el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="6">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-icon" style="background: #f56c6c"><el-icon :size="22"><Warning /></el-icon></div>
          <div class="stat-content"><div class="stat-value">{{ stats.soldOut }}</div><div class="stat-label">已售罄</div></div>
        </el-card>
      </el-col>
    </el-row>

    <div class="page-card">
      <!-- 筛选区 -->
      <el-form :inline="true" :model="filters" class="filter-form" @submit.prevent>
        <el-form-item label="关键词">
          <el-input v-model="filters.keyword" placeholder="标题/副标题" clearable style="width: 200px" :prefix-icon="Search" @keyup.enter="onSearch" />
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="filters.status" placeholder="全部" clearable style="width: 120px" @change="onSearch">
            <el-option label="下架" :value="0" />
            <el-option label="上架" :value="1" />
            <el-option label="售罄" :value="2" />
          </el-select>
        </el-form-item>
        <el-form-item label="审核状态">
          <el-select v-model="filters.audit_status" placeholder="全部" clearable style="width: 120px" @change="onSearch">
            <el-option label="待审核" :value="0" />
            <el-option label="已通过" :value="1" />
            <el-option label="已拒绝" :value="2" />
          </el-select>
        </el-form-item>
        <el-form-item label="店铺ID">
          <el-input-number v-model="filters.shop_id" :controls="false" placeholder="店铺ID" style="width: 140px" />
        </el-form-item>
        <el-form-item label="时间范围">
          <el-date-picker v-model="filters.dateRange" type="daterange" range-separator="至" start-placeholder="开始" end-placeholder="结束" value-format="YYYY-MM-DD" style="width: 240px" />
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
          <el-button type="primary" :icon="Plus" @click="openCreate">新增团购</el-button>
        </div>
      </div>

      <!-- 表格 -->
      <el-table v-loading="loading" :data="list" border stripe @sort-change="onSortChange">
        <el-table-column prop="id" label="ID" width="70" sortable="custom" fixed="left" />
        <el-table-column label="封面" width="80">
          <template #default="{ row }">
            <el-image v-if="row.cover_image" :src="row.cover_image" fit="cover" style="width: 50px; height: 38px; border-radius: 4px" />
            <span v-else class="text-muted">-</span>
          </template>
        </el-table-column>
        <el-table-column label="标题" min-width="220">
          <template #default="{ row }">
            <el-link type="primary" :underline="false" @click="openDetail(row)">{{ row.title }}</el-link>
            <div class="text-muted text-xs">{{ row.sub_title || '-' }}</div>
          </template>
        </el-table-column>
        <el-table-column label="店铺" width="140">
          <template #default="{ row }">{{ row.shop_name || `#${row.shop_id || '-'}` }}</template>
        </el-table-column>
        <el-table-column label="团购价" width="100">
          <template #default="{ row }">¥{{ Number(row.price || 0).toFixed(2) }}</template>
        </el-table-column>
        <el-table-column label="原价" width="100">
          <template #default="{ row }">¥{{ Number(row.origin_price || 0).toFixed(2) }}</template>
        </el-table-column>
        <el-table-column label="销量" width="80" prop="sales_count" />
        <el-table-column label="库存" width="80" prop="stock" />
        <el-table-column label="状态" width="90">
          <template #default="{ row }">
            <el-tag :type="statusTagType(row.status)" size="small" effect="plain">{{ statusText(row.status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="审核" width="90">
          <template #default="{ row }">
            <el-tag :type="auditTagType(row.audit_status)" size="small">{{ auditText(row.audit_status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="创建时间" width="160" prop="created_at" sortable="custom">
          <template #default="{ row }">{{ formatTime(row.created_at) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="240" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="openDetail(row)">详情</el-button>
            <el-button v-if="row.audit_status === 0 || row.audit_status === 2" type="success" link size="small" @click="handleAudit(row, 1)">通过</el-button>
            <el-button v-if="row.audit_status === 0 || row.audit_status === 1" type="danger" link size="small" @click="handleAudit(row, 2)">拒绝</el-button>
            <el-button v-if="row.status === 0" type="primary" link size="small" @click="handleStatus(row, 1)">上架</el-button>
            <el-button v-if="row.status === 1" type="warning" link size="small" @click="handleStatus(row, 0)">下架</el-button>
            <el-button type="danger" link size="small" @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>

      <!-- 分页 -->
      <div class="pagination-wrap">
        <el-pagination v-model:current-page="page" v-model:page-size="pageSize" :total="total" :page-sizes="[10, 20, 50, 100]" layout="total, sizes, prev, pager, next, jumper" background @current-change="loadList" @size-change="loadList" />
      </div>
    </div>

    <!-- 新增/编辑弹窗 -->
    <el-dialog v-model="formVisible" :title="formType === 'create' ? '新增团购' : '编辑团购'" width="640px" destroy-on-close>
      <el-form ref="formRef" :model="form" :rules="formRules" label-width="100px">
        <el-form-item label="标题" prop="title">
          <el-input v-model="form.title" placeholder="请输入团购标题" maxlength="100" show-word-limit />
        </el-form-item>
        <el-form-item label="副标题" prop="sub_title">
          <el-input v-model="form.sub_title" placeholder="请输入副标题" maxlength="200" />
        </el-form-item>
        <el-form-item label="封面图URL" prop="cover_image">
          <el-input v-model="form.cover_image" placeholder="请输入封面图URL" />
        </el-form-item>
        <el-form-item label="店铺ID" prop="shop_id">
          <el-input-number v-model="form.shop_id" :min="0" :controls="false" style="width: 200px" />
        </el-form-item>
        <el-form-item label="分类ID">
          <el-input-number v-model="form.category_id" :min="0" :controls="false" style="width: 200px" />
        </el-form-item>
        <el-form-item label="团购价" prop="price">
          <el-input-number v-model="form.price" :min="0" :precision="2" :controls="false" style="width: 200px" />
        </el-form-item>
        <el-form-item label="原价">
          <el-input-number v-model="form.origin_price" :min="0" :precision="2" :controls="false" style="width: 200px" />
        </el-form-item>
        <el-form-item label="库存" prop="stock">
          <el-input-number v-model="form.stock" :min="0" :controls="false" style="width: 200px" />
        </el-form-item>
        <el-form-item label="限购">
          <el-input-number v-model="form.limit_per_user" :min="0" :controls="false" style="width: 200px" />
          <span class="text-muted text-xs" style="margin-left: 8px">0 表示不限购</span>
        </el-form-item>
        <el-form-item label="开始时间" prop="start_time">
          <el-date-picker v-model="form.start_time" type="datetime" value-format="YYYY-MM-DD HH:mm:ss" style="width: 240px" />
        </el-form-item>
        <el-form-item label="结束时间" prop="end_time">
          <el-date-picker v-model="form.end_time" type="datetime" value-format="YYYY-MM-DD HH:mm:ss" style="width: 240px" />
        </el-form-item>
        <el-form-item label="详情描述">
          <el-input v-model="form.description" type="textarea" :rows="4" placeholder="请输入详情描述" maxlength="2000" show-word-limit />
        </el-form-item>
        <el-form-item label="精选">
          <el-switch v-model="form.is_featured" :active-value="1" :inactive-value="0" />
        </el-form-item>
        <el-form-item label="热门">
          <el-switch v-model="form.is_hot" :active-value="1" :inactive-value="0" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="formVisible = false">取消</el-button>
        <el-button type="primary" :loading="formLoading" @click="onFormSubmit">确认</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Refresh, RefreshLeft, Search, Plus, Present, Clock, CircleCheck, Warning } from '@element-plus/icons-vue'
import { formatTime } from '@/utils/format'
import { getGroupbuyList, createGroupbuy, updateGroupbuy, deleteGroupbuy, auditGroupbuy, updateGroupbuyStatus } from '@/api/groupbuy'

const router = useRouter()

const loading = ref(false)
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)
const list = ref([])
const sortField = ref('created_at')
const sortOrder = ref('descending')

const stats = reactive({ total: 0, online: 0, pendingAudit: 0, soldOut: 0 })

const filters = reactive({ keyword: '', status: null, audit_status: null, shop_id: null, dateRange: null })

const onSearch = () => { page.value = 1; loadList() }
const onReset = () => {
  filters.keyword = ''
  filters.status = null
  filters.audit_status = null
  filters.shop_id = null
  filters.dateRange = null
  page.value = 1
  loadList()
}
const onSortChange = ({ prop, order }) => {
  sortField.value = prop || 'created_at'
  sortOrder.value = order || 'descending'
  loadList()
}

const statusText = (s) => ({ 0: '下架', 1: '上架', 2: '售罄' }[s] || '-')
const statusTagType = (s) => ({ 0: 'info', 1: 'success', 2: 'warning' }[s] || 'info')
const auditText = (s) => ({ 0: '待审核', 1: '已通过', 2: '已拒绝' }[s] || '-')
const auditTagType = (s) => ({ 0: 'warning', 1: 'success', 2: 'danger' }[s] || 'info')

const loadList = async () => {
  loading.value = true
  try {
    const params = {
      page: page.value,
      page_size: pageSize.value,
      keyword: filters.keyword.trim() || undefined,
      status: filters.status === null || filters.status === '' ? undefined : filters.status,
      audit_status: filters.audit_status === null || filters.audit_status === '' ? undefined : filters.audit_status,
      shop_id: filters.shop_id || undefined,
      sort: sortField.value,
      order: sortOrder.value
    }
    if (filters.dateRange && filters.dateRange.length === 2) {
      params.start_date = filters.dateRange[0]
      params.end_date = filters.dateRange[1]
    }
    const res = await getGroupbuyList(params)
    const data = res.data || {}
    list.value = data.list || []
    total.value = data.total || 0
    calcStats()
  } catch (e) {
    ElMessage.error('加载团购列表失败')
    list.value = []
    total.value = 0
  } finally {
    loading.value = false
  }
}

const calcStats = () => {
  const all = list.value || []
  stats.total = total.value
  stats.online = all.filter((r) => r.status === 1).length
  stats.pendingAudit = all.filter((r) => r.audit_status === 0).length
  stats.soldOut = all.filter((r) => r.status === 2).length
}

const openDetail = (row) => {
  router.push(`/marketing/groupbuy/detail/${row.id}`)
}

// 新增/编辑
const formVisible = ref(false)
const formLoading = ref(false)
const formType = ref('create')
const formRef = ref(null)
const form = reactive({
  id: null, title: '', sub_title: '', cover_image: '', shop_id: 0, category_id: 0,
  price: 0, origin_price: 0, stock: 0, limit_per_user: 0,
  start_time: null, end_time: null, description: '', is_featured: 0, is_hot: 0
})
const formRules = {
  title: [{ required: true, message: '请输入标题', trigger: 'blur' }],
  cover_image: [{ required: true, message: '请输入封面图URL', trigger: 'blur' }],
  price: [{ required: true, message: '请输入团购价', trigger: 'blur' }],
  stock: [{ required: true, message: '请输入库存', trigger: 'blur' }],
  start_time: [{ required: true, message: '请选择开始时间', trigger: 'change' }],
  end_time: [{ required: true, message: '请选择结束时间', trigger: 'change' }]
}

const openCreate = () => {
  formType.value = 'create'
  Object.assign(form, {
    id: null, title: '', sub_title: '', cover_image: '', shop_id: 0, category_id: 0,
    price: 0, origin_price: 0, stock: 0, limit_per_user: 0,
    start_time: null, end_time: null, description: '', is_featured: 0, is_hot: 0
  })
  formVisible.value = true
}

const onFormSubmit = async () => {
  if (!formRef.value) return
  try {
    await formRef.value.validate()
    formLoading.value = true
    const payload = { ...form }
    delete payload.id
    if (formType.value === 'create') {
      await createGroupbuy(payload)
      ElMessage.success('创建成功')
    } else {
      await updateGroupbuy(form.id, payload)
      ElMessage.success('更新成功')
    }
    formVisible.value = false
    await loadList()
  } catch (e) {
    if (e && e.message) ElMessage.error(e.message)
  } finally {
    formLoading.value = false
  }
}

const handleAudit = async (row, auditStatus) => {
  try {
    if (auditStatus === 2) {
      const { value } = await ElMessageBox.prompt('请输入拒绝原因（可选）', '拒绝审核', {
        confirmButtonText: '确定', cancelButtonText: '取消', inputType: 'textarea', inputPlaceholder: '拒绝原因'
      })
      await auditGroupbuy(row.id, { audit_status: auditStatus, audit_reason: value || '' })
    } else {
      await ElMessageBox.confirm(`确定通过团购 "${row.title}" 的审核吗？`, '提示', { type: 'warning' })
      await auditGroupbuy(row.id, { audit_status: auditStatus })
    }
    ElMessage.success('操作成功')
    await loadList()
  } catch (e) { /* cancel */ }
}

const handleStatus = async (row, status) => {
  try {
    const action = status === 1 ? '上架' : '下架'
    await ElMessageBox.confirm(`确定${action}团购 "${row.title}" 吗？`, '提示', { type: 'warning' })
    await updateGroupbuyStatus(row.id, status)
    ElMessage.success(`${action}成功`)
    await loadList()
  } catch (e) { /* cancel */ }
}

const handleDelete = async (row) => {
  try {
    await ElMessageBox.confirm(`确定删除团购 "${row.title}" 吗？此操作不可恢复！`, '危险操作', { type: 'warning', confirmButtonText: '确认删除', cancelButtonText: '取消' })
    await deleteGroupbuy(row.id)
    ElMessage.success('删除成功')
    await loadList()
  } catch (e) { /* cancel */ }
}

onMounted(() => loadList())
</script>

<style scoped>
.stat-row { margin-bottom: 16px; }
.stat-card { display: flex; align-items: center; }
.stat-card :deep(.el-card__body) { display: flex; align-items: center; gap: 14px; padding: 16px; width: 100%; }
.stat-icon { width: 44px; height: 44px; border-radius: 8px; display: flex; align-items: center; justify-content: center; color: #fff; flex-shrink: 0; }
.stat-content { flex: 1; min-width: 0; }
.stat-value { font-size: 22px; font-weight: 600; color: #303133; }
.stat-label { font-size: 12px; color: #909399; margin-top: 2px; }
.filter-form { margin-bottom: 12px; padding: 12px 16px; background: #fafafa; border-radius: 4px; }
.filter-form :deep(.el-form-item) { margin-bottom: 8px; margin-right: 12px; }
.toolbar { display: flex; justify-content: space-between; align-items: center; flex-wrap: wrap; gap: 8px; margin-bottom: 12px; }
.toolbar-left { display: flex; gap: 8px; flex-wrap: wrap; }
.pagination-wrap { margin-top: 16px; display: flex; justify-content: flex-end; }
.text-muted { color: #909399; }
.text-xs { font-size: 12px; }
</style>
