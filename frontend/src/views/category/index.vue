<template>
  <div class="app-container">
    <div class="page-card">
      <div class="toolbar">
        <div class="toolbar-left">
          <el-button type="primary" :icon="Plus" @click="openCreate(0)">新建分类</el-button>
          <el-button :icon="Refresh" @click="loadTree">刷新</el-button>
          <el-button :icon="Sort" @click="toggleExpandAll">{{ isExpandAll ? '折叠全部' : '展开全部' }}</el-button>
        </div>
        <div class="toolbar-right">
          <el-input
            v-model="search"
            placeholder="分类名称"
            clearable
            style="width: 200px"
            :prefix-icon="Search"
            @keyup.enter="loadTree"
            @clear="loadTree"
          />
          <el-button type="primary" :icon="Search" style="margin-left: 8px" @click="loadTree">搜索</el-button>
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
        <el-table-column prop="name" label="分类名称" min-width="200" />
        <el-table-column prop="icon" label="图标" width="120" show-overflow-tooltip />
        <el-table-column label="层级" width="80">
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
    <el-dialog v-model="dialogVisible" :title="isEdit ? '编辑分类' : '新建分类'" width="560px" @close="onDialogClose">
      <el-form ref="formRef" :model="form" :rules="formRules" label-width="90px">
        <el-form-item label="父级分类">
          <el-tree-select
            v-model="form.parent_id"
            :data="parentOptions"
            :props="{ value: 'id', label: 'name', children: 'children' }"
            check-strictly
            clearable
            node-key="id"
            placeholder="不选则为顶级分类"
            style="width: 100%"
          />
        </el-form-item>
        <el-form-item label="分类名称" prop="name">
          <el-input v-model="form.name" placeholder="请输入分类名称" maxlength="50" show-word-limit />
        </el-form-item>
        <el-form-item label="图标" prop="icon">
          <el-input v-model="form.icon" placeholder="图标URL或类名，可选" />
        </el-form-item>
        <el-form-item label="层级" prop="level">
          <el-radio-group v-model="form.level">
            <el-radio :value="1">一级</el-radio>
            <el-radio :value="2">二级</el-radio>
            <el-radio :value="3">三级</el-radio>
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
import { getCategoryTree, createCategory, updateCategory, deleteCategory } from '@/api/category'

const loading = ref(false)
const search = ref('')
const treeData = ref([])
const isExpandAll = ref(true)
const refreshTable = ref(true)

const levelText = (l) => ({ 1: '一级', 2: '二级', 3: '三级' }[l] || '-')

// 关键词过滤树
const filteredList = computed(() => {
  if (!search.value.trim()) return treeData.value
  const kw = search.value.trim()
  const filterNode = (nodes) => {
    if (!Array.isArray(nodes)) return []
    const result = []
    nodes.forEach((n) => {
      const children = filterNode(n.children)
      if (n.name.includes(kw) || children.length > 0) {
        result.push({ ...n, children })
      }
    })
    return result
  }
  return filterNode(treeData.value)
})

// 父级选项树（顶级加"根分类"）
const parentOptions = computed(() => {
  return [{ id: 0, name: '根分类', children: treeData.value }]
})

const loadTree = async () => {
  loading.value = true
  try {
    const res = await getCategoryTree()
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
  icon: '',
  parent_id: 0,
  level: 1,
  sort: 0,
  status: 1
})
const formRules = {
  name: [{ required: true, message: '请输入分类名称', trigger: 'blur' }],
  level: [{ required: true, message: '请选择层级', trigger: 'change' }]
}

const resetForm = () => {
  Object.assign(form, { id: 0, name: '', icon: '', parent_id: 0, level: 1, sort: 0, status: 1 })
}

const openCreate = (parentId) => {
  isEdit.value = false
  resetForm()
  form.parent_id = parentId || 0
  // 根据父级自动推算层级
  if (parentId) {
    const node = findNode(treeData.value, parentId)
    if (node) form.level = Math.min((node.level || 1) + 1, 3)
  }
  dialogVisible.value = true
}

const findNode = (nodes, id) => {
  for (const n of nodes || []) {
    if (n.id === id) return n
    const f = findNode(n.children, id)
    if (f) return f
  }
  return null
}

const openEdit = (row) => {
  isEdit.value = true
  Object.assign(form, {
    id: row.id,
    name: row.name || '',
    icon: row.icon || '',
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
        await updateCategory(form.id, {
          name: form.name,
          icon: form.icon,
          sort: form.sort,
          status: form.status
        })
        ElMessage.success('更新成功')
      } else {
        await createCategory({
          name: form.name,
          icon: form.icon,
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
    await updateCategory(row.id, { status: row.status === 1 ? 0 : 1 })
    ElMessage.success('状态已更新')
    await loadTree()
  } catch (e) {
    // ignore
  }
}

const handleDelete = async (row) => {
  try {
    await ElMessageBox.confirm(`确定删除分类 "${row.name}" 吗？子分类将一并处理。`, '提示', { type: 'warning' })
    await deleteCategory(row.id)
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
