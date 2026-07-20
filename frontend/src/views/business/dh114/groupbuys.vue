<template>
  <div class="app-container">
    <!-- 统计卡片 -->
    <el-row :gutter="16" class="stat-row">
      <el-col :xs="12" :sm="8" :md="4">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value">{{ stats.total }}</div><div class="stat-label">团购总数</div></div></el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="4">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value text-success">{{ stats.active }}</div><div class="stat-label">进行中</div></div></el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="4">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value text-warning">{{ stats.pending }}</div><div class="stat-label">未开始</div></div></el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="4">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value text-danger">{{ stats.expired }}</div><div class="stat-label">已结束</div></div></el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="4">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value text-primary">{{ stats.soldTotal }}</div><div class="stat-label">总销量</div></div></el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="4">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value">¥{{ totalAmount.toFixed(2) }}</div><div class="stat-label">总销售额</div></div></el-card>
      </el-col>
    </el-row>

    <div class="page-card">
      <el-form :inline="true" :model="filters" class="filter-form" @submit.prevent>
        <el-form-item label="商户ID">
          <el-input v-model="filters.dh114_id" placeholder="商户ID" clearable style="width: 140px" @keyup.enter="onSearch" />
        </el-form-item>
        <el-form-item label="关键词">
          <el-input v-model="filters.keyword" placeholder="团购名" clearable style="width: 200px" @keyup.enter="onSearch" />
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="filters.status" placeholder="全部" clearable style="width: 120px" @change="onSearch">
            <el-option label="启用" :value="1" />
            <el-option label="禁用" :value="0" />
          </el-select>
        </el-form-item>
        <el-form-item label="时间范围">
          <el-date-picker
            v-model="filters.dateRange"
            type="daterange"
            range-separator="至"
            start-placeholder="开始"
            end-placeholder="结束"
            value-format="YYYY-MM-DD"
            style="width: 240px"
          />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" :icon="Search" @click="onSearch">搜索</el-button>
          <el-button :icon="RefreshLeft" @click="onReset">重置</el-button>
        </el-form-item>
      </el-form>

      <div class="toolbar">
        <el-button :icon="Refresh" @click="loadList">刷新</el-button>
        <el-button type="primary" :icon="Plus" @click="openCreate">新建团购</el-button>
      </div>

      <el-table v-loading="loading" :data="list" border stripe>
        <el-table-column prop="id" label="ID" width="70" />
        <el-table-column prop="dh114_id" label="商户ID" width="90" />
        <el-table-column label="封面" width="70">
          <template #default="{ row }">
            <el-image v-if="row.cover_image" :src="row.cover_image" fit="cover" class="thumb" preview-teleported :preview-src-list="[row.cover_image]" />
            <div v-else class="thumb thumb-empty">无图</div>
          </template>
        </el-table-column>
        <el-table-column label="团购名" min-width="200">
          <template #default="{ row }">
            <span class="gb-name">{{ row.name }}</span>
            <el-tag v-if="row.is_hot" type="danger" size="small" effect="dark" style="margin-left: 6px">HOT</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="团购价" width="100">
          <template #default="{ row }">
            <span class="price">¥{{ Number(row.groupbuy_price || 0).toFixed(2) }}</span>
          </template>
        </el-table-column>
        <el-table-column label="原价" width="100">
          <template #default="{ row }">
            <span v-if="row.original_price > 0" class="original-price">¥{{ Number(row.original_price).toFixed(2) }}</span>
            <span v-else class="text-muted">-</span>
          </template>
        </el-table-column>
        <el-table-column label="销量/库存" width="110">
          <template #default="{ row }">{{ row.sold_count || 0 }}/{{ row.total_count || '-' }}</template>
        </el-table-column>
        <el-table-column label="每人限购" width="90">
          <template #default="{ row }">{{ row.limit_per_user || '-' }}</template>
        </el-table-column>
        <el-table-column label="有效期" width="200">
          <template #default="{ row }">
            <div class="time-cell">
              <div>{{ formatTime(row.start_time, 'YYYY-MM-DD') }}</div>
              <div>至 {{ formatTime(row.end_time, 'YYYY-MM-DD') }}</div>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="90">
          <template #default="{ row }">
            <el-switch :model-value="row.status === 1" @change="(val) => onToggle(row, val)" />
          </template>
        </el-table-column>
        <el-table-column label="操作" width="160" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="openEdit(row)">编辑</el-button>
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

    <!-- 表单弹窗 -->
    <el-dialog v-model="formVisible" :title="formTitle" width="720px" destroy-on-close>
      <el-form ref="formRef" :model="form" :rules="rules" label-width="100px">
        <el-form-item label="商户ID" prop="dh114_id">
          <el-input-number v-model="form.dh114_id" :min="1" :controls="false" style="width: 100%" />
        </el-form-item>
        <el-form-item label="团购名" prop="name">
          <el-input v-model="form.name" maxlength="128" />
        </el-form-item>
        <el-form-item label="封面图">
          <el-input v-model="form.cover_image" placeholder="图片 URL" />
        </el-form-item>
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="团购价" prop="groupbuy_price">
              <el-input-number v-model="form.groupbuy_price" :min="0" :precision="2" :controls="false" style="width: 100%" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="原价">
              <el-input-number v-model="form.original_price" :min="0" :precision="2" :controls="false" style="width: 100%" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="总量">
              <el-input-number v-model="form.total_count" :min="0" :controls="false" style="width: 100%" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="每人限购">
              <el-input-number v-model="form.limit_per_user" :min="1" :controls="false" style="width: 100%" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-form-item label="有效期">
          <el-date-picker
            v-model="form.dateRange"
            type="datetimerange"
            range-separator="至"
            start-placeholder="开始时间"
            end-placeholder="结束时间"
            value-format="YYYY-MM-DD HH:mm:ss"
            style="width: 100%"
          />
        </el-form-item>
        <el-row :gutter="16">
          <el-col :span="8">
            <el-form-item label="热门">
              <el-switch v-model="form.is_hot" :active-value="1" :inactive-value="0" />
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item label="状态">
              <el-switch v-model="form.status" :active-value="1" :inactive-value="0" />
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item label="排序">
              <el-input-number v-model="form.sort" :min="0" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-form-item label="描述">
          <el-input v-model="form.description" type="textarea" :rows="3" maxlength="1000" show-word-limit />
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

const stats = reactive({ total: 0, active: 0, pending: 0, expired: 0, soldTotal: 0 })
const totalAmount = ref(0)

const filters = reactive({ dh114_id: '', keyword: '', status: null, dateRange: null })

const onSearch = () => { page.value = 1; loadList() }
const onReset = () => {
  filters.dh114_id = ''
  filters.keyword = ''
  filters.status = null
  filters.dateRange = null
  page.value = 1
  loadList()
}

const loadList = async () => {
  loading.value = true
  try {
    const params = {
      page: page.value,
      page_size: pageSize.value,
      dh114_id: filters.dh114_id || undefined,
      keyword: filters.keyword || undefined,
      status: filters.status === null || filters.status === '' ? undefined : filters.status
    }
    if (filters.dateRange && filters.dateRange.length === 2) {
      params.start_date = filters.dateRange[0]
      params.end_date = filters.dateRange[1]
    }
    const res = await request.get('/dh114/admin/groupbuys', { params })
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
  const now = Date.now()
  const total = list.value.length
  const active = list.value.filter((r) => r.status === 1 && new Date(r.start_time).getTime() <= now && new Date(r.end_time).getTime() >= now).length
  const pending = list.value.filter((r) => new Date(r.start_time).getTime() > now).length
  const expired = list.value.filter((r) => new Date(r.end_time).getTime() < now).length
  const soldTotal = list.value.reduce((s, r) => s + (r.sold_count || 0), 0)
  const amount = list.value.reduce((s, r) => s + (r.sold_count || 0) * Number(r.groupbuy_price || 0), 0)
  Object.assign(stats, { total, active, pending, expired, soldTotal })
  totalAmount.value = amount
}

// ===== 表单 =====
const formVisible = ref(false)
const formLoading = ref(false)
const formRef = ref(null)
const isEdit = ref(false)
const formTitle = computed(() => isEdit.value ? '编辑团购' : '新建团购')
const form = reactive({
  id: null, dh114_id: undefined, name: '', cover_image: '',
  groupbuy_price: 0, original_price: 0, total_count: 0, limit_per_user: 1,
  dateRange: [], is_hot: 0, status: 1, sort: 0, description: ''
})
const rules = {
  dh114_id: [{ required: true, message: '请输入商户ID', trigger: 'blur' }],
  name: [{ required: true, message: '请输入团购名', trigger: 'blur' }],
  groupbuy_price: [{ required: true, message: '请输入团购价', trigger: 'blur' }]
}

const resetForm = () => {
  Object.assign(form, {
    id: null, dh114_id: undefined, name: '', cover_image: '',
    groupbuy_price: 0, original_price: 0, total_count: 0, limit_per_user: 1,
    dateRange: [], is_hot: 0, status: 1, sort: 0, description: ''
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
    id: row.id, dh114_id: row.dh114_id, name: row.name || '',
    cover_image: row.cover_image || '', groupbuy_price: row.groupbuy_price || 0,
    original_price: row.original_price || 0, total_count: row.total_count || 0,
    limit_per_user: row.limit_per_user || 1,
    dateRange: row.start_time && row.end_time ? [row.start_time, row.end_time] : [],
    is_hot: row.is_hot || 0, status: row.status || 1, sort: row.sort || 0,
    description: row.description || ''
  })
  formVisible.value = true
}

const onSubmit = async () => {
  try {
    await formRef.value.validate()
    formLoading.value = true
    const payload = {
      dh114_id: form.dh114_id, name: form.name, cover_image: form.cover_image,
      groupbuy_price: form.groupbuy_price, original_price: form.original_price,
      total_count: form.total_count, limit_per_user: form.limit_per_user,
      start_time: form.dateRange?.[0] || undefined,
      end_time: form.dateRange?.[1] || undefined,
      is_hot: form.is_hot, status: form.status, sort: form.sort, description: form.description
    }
    if (isEdit.value) {
      await request.put(`/dh114/groupbuys/${form.id}`, payload)
    } else {
      await request.post('/dh114/groupbuys', payload)
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
    await ElMessageBox.confirm(`确认删除团购 "${row.name}"？`, '提示', { type: 'warning' })
    await request.delete(`/dh114/groupbuys/${row.id}`)
    ElMessage.success('删除成功')
    await loadList()
  } catch (e) { /* cancel */ }
}

const onToggle = async (row, val) => {
  try {
    await request.put(`/dh114/admin/groupbuys/${row.id}/status`, { status: val ? 1 : 0 })
    row.status = val ? 1 : 0
    ElMessage.success('状态已更新')
  } catch (e) { /* ignore */ }
}

onMounted(() => loadList())
</script>

<style scoped>
.stat-row { margin-bottom: 16px; }
.stat-item { text-align: center; padding: 16px; }
.stat-value { font-size: 22px; font-weight: 600; color: #409eff; }
.text-success { color: #67c23a; }
.text-warning { color: #e6a23c; }
.text-danger { color: #f56c6c; }
.text-primary { color: #409eff; }
.filter-form { margin-bottom: 12px; padding: 12px 16px; background: #fafafa; border-radius: 4px; }
.filter-form :deep(.el-form-item) { margin-bottom: 8px; margin-right: 12px; }
.toolbar { display: flex; justify-content: space-between; margin-bottom: 12px; }
.text-muted { color: #909399; }
.thumb { width: 50px; height: 50px; border-radius: 4px; border: 1px solid #ebeef5; }
.thumb-empty { display: flex; align-items: center; justify-content: center; background: #fafafa; color: #909399; font-size: 12px; }
.gb-name { color: #303133; font-weight: 500; }
.price { color: #f56c6c; font-weight: 600; }
.original-price { color: #c0c4cc; text-decoration: line-through; font-size: 12px; }
.time-cell { font-size: 12px; color: #606266; line-height: 1.6; }
.pagination-wrap { margin-top: 16px; display: flex; justify-content: flex-end; }
</style>
