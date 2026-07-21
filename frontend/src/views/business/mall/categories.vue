<template>
  <div class="app-container">
    <!-- 分类统计 -->
    <el-row :gutter="16" class="stat-row">
      <el-col :xs="12" :sm="8" :md="6">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value">{{ stats.total }}</div><div class="stat-label">分类总数</div></div></el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="6">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value text-success">{{ stats.enabled }}</div><div class="stat-label">启用</div></div></el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="6">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value text-warning">{{ stats.disabled }}</div><div class="stat-label">禁用</div></div></el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="6">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value text-primary">{{ stats.max_level }}</div><div class="stat-label">最大层级</div></div></el-card>
      </el-col>
    </el-row>

    <div class="page-card">
      <el-form :inline="true" :model="filters" class="filter-form" @submit.prevent>
        <el-form-item label="分类名">
          <el-input v-model="filters.keyword" placeholder="分类名" clearable style="width: 200px" @keyup.enter="onSearch" />
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
        <div class="toolbar-left">
          <el-button :icon="Refresh" @click="loadTree">刷新</el-button>
          <el-button type="primary" :icon="Plus" @click="openCreate(null)">新建顶级分类</el-button>
        </div>
        <div class="toolbar-right">
          <el-switch v-model="expandAll" active-text="展开全部" inactive-text="折叠全部" @change="onExpandChange" />
        </div>
      </div>

      <el-table
        v-loading="loading"
        :data="treeData"
        row-key="id"
        border
        stripe
        :default-expand-all="expandAll"
        :tree-props="{ children: 'children', hasChildren: 'hasChildren' }"
      >
        <el-table-column prop="name" label="分类名" min-width="220" />
        <el-table-column label="图标" width="80">
          <template #default="{ row }">
            <el-image v-if="row.icon" :src="row.icon" fit="cover" style="width: 32px; height: 32px" />
            <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column prop="sort" label="排序" width="80" />
        <el-table-column label="层级" width="80">
          <template #default="{ row }">
            <el-tag size="small" :type="levelTagType(row.level)">L{{ row.level }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="90">
          <template #default="{ row }">
            <el-switch :model-value="row.status === 1" @change="(val) => onToggle(row, val)" />
          </template>
        </el-table-column>
        <el-table-column prop="description" label="描述" min-width="180" show-overflow-tooltip />
        <el-table-column label="操作" width="220" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="openCreate(row)">新增子分类</el-button>
            <el-button type="primary" link size="small" @click="openEdit(row)">编辑</el-button>
            <el-button type="danger" link size="small" @click="onDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </div>

    <!-- 表单弹窗 -->
    <el-dialog v-model="formVisible" :title="formTitle" width="560px" destroy-on-close>
      <el-form ref="formRef" :model="form" :rules="rules" label-width="100px">
        <el-form-item label="父分类">
          <el-input v-if="!form.parent_id" value="（顶级分类）" disabled />
          <el-input v-else :value="parentName" disabled />
        </el-form-item>
        <el-form-item label="分类名" prop="name">
          <el-input v-model="form.name" maxlength="64" placeholder="分类名" />
        </el-form-item>
        <el-form-item label="图标">
          <el-input v-model="form.icon" placeholder="图标 URL" />
        </el-form-item>
        <el-form-item label="排序">
          <el-input-number v-model="form.sort" :min="0" :max="9999" />
        </el-form-item>
        <el-form-item label="状态">
          <el-switch v-model="form.status" :active-value="1" :inactive-value="0" />
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="form.description" type="textarea" :rows="2" maxlength="255" />
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
import {
  getMallCategoryTree, getMallCategoryList,
  createMallCategory, updateMallCategory, deleteMallCategory, updateMallCategoryStatus
} from '@/api/mall'

const loading = ref(false)
const treeData = ref([])
const flatList = ref([])
const expandAll = ref(true)

const stats = reactive({ total: 0, enabled: 0, disabled: 0, max_level: 0 })

const filters = reactive({ keyword: '', status: null })

const onSearch = () => loadTree()
const onReset = () => {
  filters.keyword = ''
  filters.status = null
  loadTree()
}

const onExpandChange = (val) => {
  // el-table 的 default-expand-all 仅在初次渲染生效，需重新加载
  reloadTree()
}

const reloadTree = async () => {
  loading.value = true
  try {
    const data = await fetchTree()
    treeData.value = data
    flatList.value = flatten(data)
    computeStats()
  } finally {
    loading.value = false
  }
}

const loadTree = async () => {
  await reloadTree()
}

const fetchTree = async () => {
  try {
    const res = await getMallCategoryTree()
    const data = res.data
    let list = data?.list || data || []
    if (filters.status !== null && filters.status !== '') {
      list = filterByStatus(list, filters.status)
    }
    if (filters.keyword) {
      list = filterByKeyword(list, filters.keyword.trim())
    }
    return list
  } catch (e) {
    ElMessage.error('加载分类树失败')
    return []
  }
}

const filterByStatus = (nodes, status) => {
  return nodes
    .map((n) => {
      const children = n.children ? filterByStatus(n.children, status) : []
      return { ...n, children }
    })
    .filter((n) => n.status === status || (n.children && n.children.length > 0))
}

const filterByKeyword = (nodes, kw) => {
  return nodes
    .map((n) => {
      const children = n.children ? filterByKeyword(n.children, kw) : []
      return { ...n, children }
    })
    .filter((n) => (n.name || '').includes(kw) || (n.children && n.children.length > 0))
}

const flatten = (nodes, level = 1, result = []) => {
  for (const n of nodes) {
    result.push({ ...n, level })
    if (n.children && n.children.length > 0) {
      flatten(n.children, level + 1, result)
    }
  }
  return result
}

const computeStats = () => {
  const total = flatList.value.length
  const enabled = flatList.value.filter((r) => r.status === 1).length
  const disabled = flatList.value.filter((r) => r.status === 0).length
  const maxLevel = flatList.value.reduce((m, r) => Math.max(m, r.level || 1), 1)
  Object.assign(stats, { total, enabled, disabled, max_level: maxLevel })
}

const levelTagType = (lvl) => ({ 1: 'primary', 2: 'success', 3: 'warning' }[lvl] || 'info')

// ===== 表单 =====
const formVisible = ref(false)
const formLoading = ref(false)
const formRef = ref(null)
const isEdit = ref(false)
const formTitle = computed(() => isEdit.value ? '编辑分类' : '新建分类')
const parentName = ref('')
const form = reactive({ id: null, parent_id: null, name: '', icon: '', sort: 0, status: 1, description: '' })
const rules = {
  name: [{ required: true, message: '请输入分类名', trigger: 'blur' }]
}

const openCreate = (parent) => {
  isEdit.value = false
  Object.assign(form, { id: null, parent_id: parent?.id || null, name: '', icon: '', sort: 0, status: 1, description: '' })
  parentName.value = parent?.name || ''
  formVisible.value = true
}

const openEdit = (row) => {
  isEdit.value = true
  Object.assign(form, {
    id: row.id,
    parent_id: row.parent_id || null,
    name: row.name || '',
    icon: row.icon || '',
    sort: row.sort || 0,
    status: row.status === undefined ? 1 : row.status,
    description: row.description || ''
  })
  const parent = flatList.value.find((r) => r.id === row.parent_id)
  parentName.value = parent?.name || ''
  formVisible.value = true
}

const onSubmit = async () => {
  try {
    await formRef.value.validate()
    formLoading.value = true
    const payload = {
      parent_id: form.parent_id || 0,
      name: form.name,
      icon: form.icon,
      sort: form.sort,
      status: form.status,
      description: form.description
    }
    if (isEdit.value) {
      await updateMallCategory(form.id, payload)
    } else {
      await createMallCategory(payload)
    }
    ElMessage.success(isEdit.value ? '更新成功' : '创建成功')
    formVisible.value = false
    await loadTree()
  } catch (e) {
    // 校验或请求失败
  } finally {
    formLoading.value = false
  }
}

const onDelete = async (row) => {
  try {
    await ElMessageBox.confirm(`确认删除分类 "${row.name}"？子分类将一并解除关联，请谨慎操作。`, '提示', { type: 'warning' })
    await deleteMallCategory(row.id)
    ElMessage.success('删除成功')
    await loadTree()
  } catch (e) { /* cancel */ }
}

const onToggle = async (row, val) => {
  try {
    await updateMallCategoryStatus(row.id, val ? 1 : 0)
    row.status = val ? 1 : 0
    ElMessage.success('状态已更新')
    computeStats()
  } catch (e) { /* ignore */ }
}

onMounted(() => loadTree())
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
.toolbar { display: flex; justify-content: space-between; align-items: center; margin-bottom: 12px; }
.toolbar-left { display: flex; gap: 8px; }
</style>
