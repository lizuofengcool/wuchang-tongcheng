<template>
  <div class="app-container">
    <div class="page-card">
      <div class="toolbar">
        <div class="toolbar-left">
          <el-button type="primary" :icon="Plus" @click="openCreate(0)">新建地区</el-button>
          <el-button :icon="Refresh" @click="loadTree">刷新</el-button>
          <el-button :icon="Sort" @click="toggleExpandAll">{{ isExpandAll ? '折叠全部' : '展开全部' }}</el-button>
        </div>
        <div class="toolbar-right">
          <el-input
            v-model="search"
            placeholder="地区名称/编码"
            clearable
            style="width: 220px"
            :prefix-icon="Search"
            @keyup.enter="onSearch"
            @clear="onSearch"
          />
          <el-button type="primary" :icon="Search" style="margin-left: 8px" @click="onSearch">搜索</el-button>
        </div>
      </div>

      <el-table
        v-if="refreshTable"
        v-loading="loading"
        :data="filteredList"
        row-key="id"
        border
        stripe
        :default-expand-all="isExpandAll"
        :tree-props="{ children: 'children', hasChildren: 'hasChildren' }"
      >
        <el-table-column prop="name" label="地区名称" min-width="180" />
        <el-table-column prop="code" label="编码" width="120" />
        <el-table-column label="层级" width="90">
          <template #default="{ row }">{{ levelText(row.level) }}</template>
        </el-table-column>
        <el-table-column prop="sort" label="排序" width="80" />
        <el-table-column label="状态" width="90">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'info'" size="small">
              {{ row.status === 1 ? '启用' : '禁用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="240" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="openCreate(row.id)">添加子级</el-button>
            <el-button type="warning" link size="small" @click="openEdit(row)">编辑</el-button>
            <el-button
              type="success"
              link
              size="small"
              @click="toggleStatus(row)"
            >{{ row.status === 1 ? '禁用' : '启用' }}</el-button>
            <el-button type="danger" link size="small" @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </div>

    <!-- 新建/编辑 -->
    <el-dialog v-model="dialogVisible" :title="isEdit ? '编辑地区' : '新建地区'" width="560px" @close="onDialogClose">
      <el-form ref="formRef" :model="form" :rules="formRules" label-width="90px">
        <el-form-item label="父级地区">
          <el-tree-select
            v-model="form.parent_id"
            :data="parentOptions"
            :props="{ value: 'id', label: 'name', children: 'children' }"
            check-strictly
            clearable
            node-key="id"
            placeholder="不选则为顶级地区"
            style="width: 100%"
          />
        </el-form-item>
        <el-form-item label="地区名称" prop="name">
          <el-input v-model="form.name" placeholder="请输入地区名称" maxlength="50" show-word-limit />
        </el-form-item>
        <el-form-item label="地区编码" prop="code">
          <el-input v-model="form.code" placeholder="如 420100" maxlength="20" show-word-limit />
        </el-form-item>
        <el-form-item label="层级" prop="level">
          <el-radio-group v-model="form.level">
            <el-radio :value="1">省级</el-radio>
            <el-radio :value="2">市级</el-radio>
            <el-radio :value="3">区县</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="排序" prop="sort">
          <el-input-number v-model="form.sort" :min="0" :max="9999" controls-position="right" />
        </el-form-item>
        <el-form-item label="状态" prop="status">
          <el-radio-group v-model="form.status">
            <el-radio :value="1">启用</el-radio>
            <el-radio :value="0">禁用</el-radio>
          </el-radio-group>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="handleSubmit">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, computed, nextTick, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Refresh, Search, Sort } from '@element-plus/icons-vue'
import { getRegionTree, createRegion, updateRegion, deleteRegion } from '@/api/region'

const loading = ref(false)
const search = ref('')
const treeData = ref([])
const isExpandAll = ref(true)
const refreshTable = ref(true)

const levelText = (l) => ({ 1: '省级', 2: '市级', 3: '区县' }[l] || '-')

// 关键词过滤树
const filteredList = computed(() => {
  if (!search.value.trim()) return treeData.value
  const kw = search.value.trim().toLowerCase()
  const filterNode = (nodes) => {
    if (!Array.isArray(nodes)) return []
    const result = []
    nodes.forEach((n) => {
      const children = filterNode(n.children)
      const matched = (n.name || '').toLowerCase().includes(kw) || (n.code || '').toLowerCase().includes(kw)
      if (matched || children.length > 0) {
        result.push({ ...n, children })
      }
    })
    return result
  }
  return filterNode(treeData.value)
})

const parentOptions = computed(() => {
  return [{ id: 0, name: '根地区', children: treeData.value }]
})

const onSearch = () => {
  if (search.value.trim()) {
    isExpandAll.value = true
    refreshTable.value = false
    nextTick(() => (refreshTable.value = true))
  }
}

const loadTree = async () => {
  loading.value = true
  try {
    const res = await getRegionTree()
    treeData.value = res.data || []
  } catch (e) {
    treeData.value = []
  } finally {
    loading.value = false
  }
}

const toggleExpandAll = async () => {
  isExpandAll.value = !isExpandAll.value
  refreshTable.value = false
  await nextTick()
  refreshTable.value = true
}

// ===== 新建/编辑 =====
const dialogVisible = ref(false)
const isEdit = ref(false)
const submitting = ref(false)
const formRef = ref(null)
const form = reactive({
  id: 0,
  name: '',
  code: '',
  parent_id: 0,
  level: 1,
  sort: 0,
  status: 1
})
const formRules = {
  name: [{ required: true, message: '请输入地区名称', trigger: 'blur' }],
  code: [{ required: true, message: '请输入地区编码', trigger: 'blur' }],
  level: [{ required: true, message: '请选择层级', trigger: 'change' }]
}

const resetForm = () => {
  Object.assign(form, { id: 0, name: '', code: '', parent_id: 0, level: 1, sort: 0, status: 1 })
}

const findNode = (nodes, id) => {
  for (const n of nodes || []) {
    if (n.id === id) return n
    const f = findNode(n.children, id)
    if (f) return f
  }
  return null
}

const openCreate = (parentId) => {
  isEdit.value = false
  resetForm()
  form.parent_id = parentId || 0
  if (parentId) {
    const node = findNode(treeData.value, parentId)
    if (node) form.level = Math.min((node.level || 1) + 1, 3)
  }
  dialogVisible.value = true
}

const openEdit = (row) => {
  isEdit.value = true
  Object.assign(form, {
    id: row.id,
    name: row.name || '',
    code: row.code || '',
    parent_id: row.parent_id || 0,
    level: row.level || 1,
    sort: row.sort ?? 0,
    status: row.status ?? 1
  })
  dialogVisible.value = true
}

const onDialogClose = () => {
  formRef.value?.clearValidate()
}

const handleSubmit = async () => {
  if (!formRef.value) return
  await formRef.value.validate(async (valid) => {
    if (!valid) return
    submitting.value = true
    try {
      if (isEdit.value) {
        await updateRegion(form.id, {
          name: form.name,
          sort: form.sort,
          status: form.status
        })
        ElMessage.success('更新成功')
      } else {
        await createRegion({
          name: form.name,
          code: form.code,
          parent_id: form.parent_id || 0,
          level: form.level,
          sort: form.sort,
          status: form.status
        })
        ElMessage.success('创建成功')
      }
      dialogVisible.value = false
      await loadTree()
    } catch (e) {
      // ignore
    } finally {
      submitting.value = false
    }
  })
}

const toggleStatus = async (row) => {
  try {
    await updateRegion(row.id, { status: row.status === 1 ? 0 : 1 })
    ElMessage.success('状态已更新')
    await loadTree()
  } catch (e) {
    // ignore
  }
}

const handleDelete = async (row) => {
  try {
    await ElMessageBox.confirm(`确定删除地区 "${row.name}" 吗？子地区将一并处理。`, '提示', { type: 'warning' })
    await deleteRegion(row.id)
    ElMessage.success('删除成功')
    await loadTree()
  } catch (e) {
    // 取消
  }
}

onMounted(() => {
  loadTree()
})
</script>

<style scoped>
.toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex-wrap: wrap;
  gap: 8px;
}
.toolbar-left {
  display: flex;
  gap: 8px;
}
.toolbar-right {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
}
</style>
