<template>
  <div class="app-container">
    <!-- 统计卡片 -->
    <el-row :gutter="16" class="stat-row">
      <el-col :xs="12" :sm="8" :md="6">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value">{{ stats.total }}</div><div class="stat-label">总推荐数</div></div></el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="6">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value text-success">{{ stats.active }}</div><div class="stat-label">进行中</div></div></el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="6">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value text-warning">{{ stats.clicked }}</div><div class="stat-label">已点击</div></div></el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="6">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value text-primary">{{ stats.contacted }}</div><div class="stat-label">已联系</div></div></el-card>
      </el-col>
    </el-row>

    <div class="page-card">
      <el-form :inline="true" :model="filters" class="filter-form" @submit.prevent>
        <el-form-item label="推荐类型">
          <el-select v-model="filters.rec_type" placeholder="全部" clearable style="width: 160px" @change="onSearch">
            <el-option label="首页推荐" value="home" />
            <el-option label="频道推荐" value="channel" />
            <el-option label="搜索推荐" value="search" />
            <el-option label="附近推荐" value="nearby" />
            <el-option label="分类推荐" value="category" />
          </el-select>
        </el-form-item>
        <el-form-item label="商户ID">
          <el-input v-model="filters.dh114_id" placeholder="商户ID" clearable style="width: 140px" @keyup.enter="onSearch" />
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="filters.status" placeholder="全部" clearable style="width: 120px" @change="onSearch">
            <el-option label="启用" :value="1" />
            <el-option label="禁用" :value="0" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" :icon="Search" @click="onSearch">搜索</el-button>
          <el-button :icon="RefreshLeft" @click="onReset">重置</el-button>
        </el-form-item>
      </el-form>

      <div class="toolbar">
        <el-button :icon="Refresh" @click="loadList">刷新</el-button>
        <el-button type="primary" :icon="Plus" @click="openCreate">新建推荐</el-button>
      </div>

      <el-table v-loading="loading" :data="list" border stripe>
        <el-table-column prop="id" label="ID" width="70" />
        <el-table-column prop="dh114_id" label="商户ID" width="90" />
        <el-table-column label="商户名" min-width="160">
          <template #default="{ row }">{{ row.business_name || `商户#${row.dh114_id}` }}</template>
        </el-table-column>
        <el-table-column label="推荐类型" width="120">
          <template #default="{ row }">
            <el-tag :type="recTypeTag(row.rec_type)" size="small">{{ recTypeText(row.rec_type) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="sort" label="排序" width="80" />
        <el-table-column label="点击/联系" width="110">
          <template #default="{ row }">{{ row.click_count || 0 }}/{{ row.contact_count || 0 }}</template>
        </el-table-column>
        <el-table-column label="状态" width="90">
          <template #default="{ row }">
            <el-switch :model-value="row.status === 1" @change="(val) => onToggle(row, val)" />
          </template>
        </el-table-column>
        <el-table-column label="开始-结束" width="220">
          <template #default="{ row }">
            <div class="time-cell">
              <div>{{ formatTime(row.start_time, 'YYYY-MM-DD HH:mm') }}</div>
              <div>至 {{ formatTime(row.end_time, 'YYYY-MM-DD HH:mm') }}</div>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="备注" min-width="160" show-overflow-tooltip>
          <template #default="{ row }">{{ row.remark || '-' }}</template>
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
    <el-dialog v-model="formVisible" :title="formTitle" width="600px" destroy-on-close>
      <el-form ref="formRef" :model="form" :rules="rules" label-width="100px">
        <el-form-item label="商户ID" prop="dh114_id">
          <el-input-number v-model="form.dh114_id" :min="1" :controls="false" style="width: 100%" />
        </el-form-item>
        <el-form-item label="推荐类型" prop="rec_type">
          <el-select v-model="form.rec_type" style="width: 100%">
            <el-option label="首页推荐" value="home" />
            <el-option label="频道推荐" value="channel" />
            <el-option label="搜索推荐" value="search" />
            <el-option label="附近推荐" value="nearby" />
            <el-option label="分类推荐" value="category" />
          </el-select>
        </el-form-item>
        <el-form-item label="有效期">
          <el-date-picker
            v-model="form.dateRange"
            type="datetimerange"
            range-separator="至"
            start-placeholder="开始"
            end-placeholder="结束"
            value-format="YYYY-MM-DD HH:mm:ss"
            style="width: 100%"
          />
        </el-form-item>
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="排序">
              <el-input-number v-model="form.sort" :min="0" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="状态">
              <el-switch v-model="form.status" :active-value="1" :inactive-value="0" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-form-item label="备注">
          <el-input v-model="form.remark" type="textarea" :rows="2" maxlength="200" />
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

const stats = reactive({ total: 0, active: 0, clicked: 0, contacted: 0 })

const filters = reactive({ rec_type: '', dh114_id: '', status: null })

const onSearch = () => { page.value = 1; loadList() }
const onReset = () => {
  filters.rec_type = ''
  filters.dh114_id = ''
  filters.status = null
  page.value = 1
  loadList()
}

const recTypeText = (t) => ({ home: '首页', channel: '频道', search: '搜索', nearby: '附近', category: '分类' }[t] || '-')
const recTypeTag = (t) => ({ home: 'danger', channel: 'warning', search: 'primary', nearby: 'success', category: 'info' }[t] || 'info')

const loadList = async () => {
  loading.value = true
  try {
    const params = {
      page: page.value,
      page_size: pageSize.value,
      rec_type: filters.rec_type || undefined,
      dh114_id: filters.dh114_id || undefined,
      status: filters.status === null || filters.status === '' ? undefined : filters.status
    }
    const res = await request.get('/dh114/admin/recommendations', { params })
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
  const active = list.value.filter((r) => r.status === 1 && (!r.end_time || new Date(r.end_time).getTime() >= now)).length
  const clicked = list.value.reduce((s, r) => s + (r.click_count || 0), 0)
  const contacted = list.value.reduce((s, r) => s + (r.contact_count || 0), 0)
  Object.assign(stats, { total, active, clicked, contacted })
}

// ===== 表单 =====
const formVisible = ref(false)
const formLoading = ref(false)
const formRef = ref(null)
const isEdit = ref(false)
const formTitle = computed(() => isEdit.value ? '编辑推荐' : '新建推荐')
const form = reactive({
  id: null, dh114_id: undefined, rec_type: 'home',
  dateRange: [], sort: 0, status: 1, remark: ''
})
const rules = {
  dh114_id: [{ required: true, message: '请输入商户ID', trigger: 'blur' }],
  rec_type: [{ required: true, message: '请选择类型', trigger: 'change' }]
}

const resetForm = () => {
  Object.assign(form, {
    id: null, dh114_id: undefined, rec_type: 'home',
    dateRange: [], sort: 0, status: 1, remark: ''
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
    id: row.id, dh114_id: row.dh114_id, rec_type: row.rec_type || 'home',
    dateRange: row.start_time && row.end_time ? [row.start_time, row.end_time] : [],
    sort: row.sort || 0, status: row.status || 1, remark: row.remark || ''
  })
  formVisible.value = true
}

const onSubmit = async () => {
  try {
    await formRef.value.validate()
    formLoading.value = true
    const payload = {
      dh114_id: form.dh114_id, rec_type: form.rec_type,
      start_time: form.dateRange?.[0] || undefined,
      end_time: form.dateRange?.[1] || undefined,
      sort: form.sort, status: form.status, remark: form.remark
    }
    if (isEdit.value) {
      await request.put(`/dh114/recommendations/${form.id}`, payload)
    } else {
      await request.post('/dh114/admin/recommendations', payload)
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
    await ElMessageBox.confirm('确认删除该推荐？', '提示', { type: 'warning' })
    await request.delete(`/dh114/admin/recommendations/${row.id}`)
    ElMessage.success('删除成功')
    await loadList()
  } catch (e) { /* cancel */ }
}

const onToggle = async (row, val) => {
  try {
    await request.put(`/dh114/recommendations/${row.id}`, { status: val ? 1 : 0 })
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
.time-cell { font-size: 12px; color: #606266; line-height: 1.6; }
.pagination-wrap { margin-top: 16px; display: flex; justify-content: flex-end; }
</style>
