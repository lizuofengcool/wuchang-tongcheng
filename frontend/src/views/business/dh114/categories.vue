<template>
  <div class="app-container">
    <!-- 统计卡片 -->
    <el-row :gutter="16" class="stat-row">
      <el-col :xs="12" :sm="8" :md="6">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value">{{ stats.total }}</div><div class="stat-label">总分类数</div></div></el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="6">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value text-success">{{ stats.enabled }}</div><div class="stat-label">启用</div></div></el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="6">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value text-warning">{{ stats.disabled }}</div><div class="stat-label">禁用</div></div></el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="6">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value text-primary">{{ stats.topLevel }}</div><div class="stat-label">一级分类</div></div></el-card>
      </el-col>
    </el-row>

    <div class="page-card">
      <el-form :inline="true" :model="filters" class="filter-form" @submit.prevent>
        <el-form-item label="关键词">
          <el-input v-model="filters.keyword" placeholder="分类名" clearable style="width: 200px" @keyup.enter="onSearch" />
        </el-form-item>
        <el-form-item label="商户类型">
          <el-select v-model="filters.business_type" placeholder="全部" clearable style="width: 160px" @change="onSearch">
            <el-option label="餐饮美食" value="food" />
            <el-option label="酒店住宿" value="hotel" />
            <el-option label="生活服务" value="life" />
            <el-option label="休闲娱乐" value="leisure" />
            <el-option label="购物" value="shopping" />
            <el-option label="教育培训" value="education" />
            <el-option label="医疗健康" value="medical" />
          </el-select>
        </el-form-item>
        <el-form-item label="层级">
          <el-select v-model="filters.level" placeholder="全部" clearable style="width: 120px" @change="onSearch">
            <el-option label="一级" :value="1" />
            <el-option label="二级" :value="2" />
            <el-option label="三级" :value="3" />
          </el-select>
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
        <el-button type="primary" :icon="Plus" @click="openCreate">新建分类</el-button>
      </div>

      <el-table v-loading="loading" :data="list" border stripe row-key="id" default-expand-all>
        <el-table-column prop="id" label="ID" width="70" />
        <el-table-column label="分类名" min-width="180">
          <template #default="{ row }">
            <span :style="{ fontWeight: row.level === 1 ? 600 : 400 }">{{ row.name }}</span>
          </template>
        </el-table-column>
        <el-table-column label="图标" width="70">
          <template #default="{ row }">
            <el-image v-if="row.icon" :src="row.icon" fit="cover" class="thumb" />
            <span v-else class="text-muted">-</span>
          </template>
        </el-table-column>
        <el-table-column prop="level" label="层级" width="80">
          <template #default="{ row }">
            <el-tag size="small">{{ row.level || 1 }}级</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="商户类型" width="120">
          <template #default="{ row }">{{ row.business_type || '-' }}</template>
        </el-table-column>
        <el-table-column prop="parent_id" label="父分类" width="90">
          <template #default="{ row }">{{ row.parent_id || '-' }}</template>
        </el-table-column>
        <el-table-column prop="sort" label="排序" width="80" />
        <el-table-column label="状态" width="90">
          <template #default="{ row }">
            <el-switch :model-value="row.status === 1" @change="(val) => onToggle(row, val)" />
          </template>
        </el-table-column>
        <el-table-column label="描述" min-width="180" show-overflow-tooltip>
          <template #default="{ row }">{{ row.description || '-' }}</template>
        </el-table-column>
        <el-table-column label="操作" width="180" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="openEdit(row)">编辑</el-button>
            <el-button type="success" link size="small" @click="openCreateChild(row)">添加子分类</el-button>
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
        <el-form-item label="分类名" prop="name">
          <el-input v-model="form.name" maxlength="64" />
        </el-form-item>
        <el-form-item label="商户类型">
          <el-select v-model="form.business_type" clearable style="width: 100%">
            <el-option label="餐饮美食" value="food" />
            <el-option label="酒店住宿" value="hotel" />
            <el-option label="生活服务" value="life" />
            <el-option label="休闲娱乐" value="leisure" />
            <el-option label="购物" value="shopping" />
            <el-option label="教育培训" value="education" />
            <el-option label="医疗健康" value="medical" />
          </el-select>
        </el-form-item>
        <el-form-item label="父分类">
          <el-select v-model="form.parent_id" clearable filterable style="width: 100%">
            <el-option v-for="c in parentOptions" :key="c.id" :label="c.name" :value="c.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="层级">
          <el-input-number v-model="form.level" :min="1" :max="5" />
        </el-form-item>
        <el-form-item label="图标">
          <el-input v-model="form.icon" placeholder="图标 URL" />
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
        <el-form-item label="描述">
          <el-input v-model="form.description" type="textarea" :rows="2" maxlength="500" show-word-limit />
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

const loading = ref(false)
const page = ref(1)
const pageSize = ref(50)
const total = ref(0)
const list = ref([])

const stats = reactive({ total: 0, enabled: 0, disabled: 0, topLevel: 0 })

const filters = reactive({ keyword: '', business_type: '', level: null, status: null })

const onSearch = () => { page.value = 1; loadList() }
const onReset = () => {
  filters.keyword = ''
  filters.business_type = ''
  filters.level = null
  filters.status = null
  page.value = 1
  loadList()
}

const loadList = async () => {
  loading.value = true
  try {
    const params = {
      page: page.value,
      page_size: pageSize.value,
      keyword: filters.keyword || undefined,
      business_type: filters.business_type || undefined,
      level: filters.level === null || filters.level === '' ? undefined : filters.level,
      status: filters.status === null || filters.status === '' ? undefined : filters.status
    }
    const res = await request.get('/dh114/categories', { params })
    const data = res.data || {}
    list.value = data.list || data || []
    total.value = data.total || list.value.length
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
  const enabled = list.value.filter((r) => r.status === 1).length
  const disabled = list.value.filter((r) => r.status === 0).length
  const topLevel = list.value.filter((r) => r.level === 1 || !r.parent_id).length
  Object.assign(stats, { total, enabled, disabled, topLevel })
}

const parentOptions = computed(() => list.value.filter((r) => r.level === 1 || !r.parent_id))

// ===== 表单 =====
const formVisible = ref(false)
const formLoading = ref(false)
const formRef = ref(null)
const isEdit = ref(false)
const formTitle = computed(() => isEdit.value ? '编辑分类' : '新建分类')
const form = reactive({
  id: null, name: '', business_type: '', parent_id: null,
  level: 1, icon: '', sort: 0, status: 1, description: ''
})
const rules = {
  name: [{ required: true, message: '请输入分类名', trigger: 'blur' }]
}

const resetForm = () => {
  Object.assign(form, {
    id: null, name: '', business_type: '', parent_id: null,
    level: 1, icon: '', sort: 0, status: 1, description: ''
  })
}

const openCreate = () => {
  isEdit.value = false
  resetForm()
  formVisible.value = true
}

const openCreateChild = (row) => {
  isEdit.value = false
  resetForm()
  form.parent_id = row.id
  form.level = (row.level || 1) + 1
  form.business_type = row.business_type || ''
  formVisible.value = true
}

const openEdit = (row) => {
  isEdit.value = true
  Object.assign(form, {
    id: row.id, name: row.name || '', business_type: row.business_type || '',
    parent_id: row.parent_id || null, level: row.level || 1,
    icon: row.icon || '', sort: row.sort || 0, status: row.status || 1,
    description: row.description || ''
  })
  formVisible.value = true
}

const onSubmit = async () => {
  try {
    await formRef.value.validate()
    formLoading.value = true
    const payload = { ...form }
    if (isEdit.value) {
      await request.put(`/dh114/categories/${form.id}`, payload)
    } else {
      await request.post('/dh114/categories', payload)
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
    await ElMessageBox.confirm(`确认删除分类 "${row.name}"？子分类也会受影响`, '提示', { type: 'warning' })
    await request.delete(`/dh114/categories/${row.id}`)
    ElMessage.success('删除成功')
    await loadList()
  } catch (e) { /* cancel */ }
}

const onToggle = async (row, val) => {
  try {
    await request.put(`/dh114/categories/${row.id}/status`, { status: val ? 1 : 0 })
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
.thumb { width: 40px; height: 40px; border-radius: 4px; border: 1px solid #ebeef5; }
.pagination-wrap { margin-top: 16px; display: flex; justify-content: flex-end; }
</style>
