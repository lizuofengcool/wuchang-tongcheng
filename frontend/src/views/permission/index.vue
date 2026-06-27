<template>
  <div class="app-container">
    <div class="page-card">
      <div class="toolbar">
        <div class="toolbar-left">
          <el-button type="primary" :icon="Plus" @click="openCreate">新建权限</el-button>
          <el-button :icon="Refresh" @click="loadPerms">刷新</el-button>
        </div>
        <div class="toolbar-right">
          <el-input
            v-model="search"
            placeholder="搜索权限名称/编码"
            clearable
            style="width: 240px"
            :prefix-icon="Search"
          />
          <el-select v-model="filterType" placeholder="类型" clearable style="width: 120px; margin-left: 8px">
            <el-option label="菜单" :value="1" />
            <el-option label="按钮" :value="2" />
            <el-option label="接口" :value="3" />
          </el-select>
        </div>
      </div>

      <el-table v-loading="loading" :data="filteredPerms" border stripe>
        <el-table-column type="index" label="#" width="50" />
        <el-table-column prop="id" label="ID" width="70" />
        <el-table-column prop="name" label="权限名称" min-width="140" />
        <el-table-column prop="code" label="权限编码" min-width="160" />
        <el-table-column label="类型" width="90">
          <template #default="{ row }">
            <el-tag :type="typeTag(row.type)" size="small">{{ typeText(row.type) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="path" label="路径" min-width="180" show-overflow-tooltip />
        <el-table-column prop="method" label="方法" width="90">
          <template #default="{ row }">
            <el-tag v-if="row.method" size="small" effect="plain">{{ row.method }}</el-tag>
            <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column prop="sort" label="排序" width="80" />
        <el-table-column label="状态" width="90">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'info'">
              {{ row.status === 1 ? '启用' : '禁用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="160" fixed="right">
          <template #default="{ row }">
            <el-button type="warning" link size="small" @click="openEdit(row)">编辑</el-button>
            <el-button type="danger" link size="small" @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </div>

    <!-- 新建/编辑权限 -->
    <el-dialog v-model="dialogVisible" :title="isEdit ? '编辑权限' : '新建权限'" width="520px">
      <el-form ref="formRef" :model="form" :rules="formRules" label-width="90px">
        <el-form-item label="权限名称" prop="name">
          <el-input v-model="form.name" placeholder="如 用户管理" />
        </el-form-item>
        <el-form-item v-if="!isEdit" label="权限编码" prop="code">
          <el-input v-model="form.code" placeholder="如 system:user:list" />
        </el-form-item>
        <el-form-item v-else label="权限编码">
          <el-input :model-value="form.code" disabled />
        </el-form-item>
        <el-form-item label="类型" prop="type">
          <el-radio-group v-model="form.type" :disabled="isEdit">
            <el-radio :value="1">菜单</el-radio>
            <el-radio :value="2">按钮</el-radio>
            <el-radio :value="3">接口</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="路径" prop="path">
          <el-input v-model="form.path" placeholder="/api/v1/xxx" />
        </el-form-item>
        <el-form-item label="请求方法" prop="method">
          <el-select v-model="form.method" placeholder="请选择" clearable style="width: 100%">
            <el-option label="GET" value="GET" />
            <el-option label="POST" value="POST" />
            <el-option label="PUT" value="PUT" />
            <el-option label="DELETE" value="DELETE" />
            <el-option label="PATCH" value="PATCH" />
          </el-select>
        </el-form-item>
        <el-form-item label="排序" prop="sort">
          <el-input-number v-model="form.sort" :min="0" :max="9999" />
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
import { ref, reactive, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Refresh, Search } from '@element-plus/icons-vue'
import { listPermissions, createPermission, updatePermission, deletePermission } from '@/api/permission'

const loading = ref(false)
const search = ref('')
const filterType = ref(null)
const perms = ref([])

const filteredPerms = computed(() => {
  const kw = search.value.trim().toLowerCase()
  return perms.value.filter((p) => {
    const matchKw =
      !kw ||
      (p.name || '').toLowerCase().includes(kw) ||
      (p.code || '').toLowerCase().includes(kw)
    const matchType = !filterType.value || p.type === filterType.value
    return matchKw && matchType
  })
})

const typeText = (t) => ({ 1: '菜单', 2: '按钮', 3: '接口' }[t] || '-')
const typeTag = (t) => ({ 1: 'success', 2: 'warning', 3: 'primary' }[t] || 'info')

const loadPerms = async () => {
  loading.value = true
  try {
    const res = await listPermissions()
    perms.value = res.data || []
  } finally {
    loading.value = false
  }
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
  type: 3,
  parent_id: 0,
  path: '',
  method: '',
  sort: 0,
  status: 1
})
const formRules = {
  name: [{ required: true, message: '请输入权限名称', trigger: 'blur' }],
  code: [{ required: true, message: '请输入权限编码', trigger: 'blur' }],
  type: [{ required: true, message: '请选择类型', trigger: 'change' }]
}

const openCreate = () => {
  isEdit.value = false
  Object.assign(form, {
    id: 0,
    name: '',
    code: '',
    type: 3,
    parent_id: 0,
    path: '',
    method: '',
    sort: 0,
    status: 1
  })
  dialogVisible.value = true
}

const openEdit = (row) => {
  isEdit.value = true
  Object.assign(form, {
    id: row.id,
    name: row.name || '',
    code: row.code || '',
    type: row.type || 3,
    parent_id: row.parent_id || 0,
    path: row.path || '',
    method: row.method || '',
    sort: row.sort || 0,
    status: row.status ?? 1
  })
  dialogVisible.value = true
}

const handleSubmit = async () => {
  if (!formRef.value) return
  await formRef.value.validate(async (valid) => {
    if (!valid) return
    submitting.value = true
    try {
      if (isEdit.value) {
        await updatePermission(form.id, {
          name: form.name,
          path: form.path,
          method: form.method,
          sort: form.sort,
          status: form.status
        })
        ElMessage.success('更新成功')
      } else {
        await createPermission({
          name: form.name,
          code: form.code,
          type: form.type,
          parent_id: form.parent_id,
          path: form.path,
          method: form.method,
          sort: form.sort,
          status: form.status
        })
        ElMessage.success('创建成功')
      }
      dialogVisible.value = false
      await loadPerms()
    } catch (e) {
      // ignore
    } finally {
      submitting.value = false
    }
  })
}

const handleDelete = async (row) => {
  try {
    await ElMessageBox.confirm(`确定删除权限 "${row.name}" 吗？`, '提示', { type: 'warning' })
    await deletePermission(row.id)
    ElMessage.success('删除成功')
    await loadPerms()
  } catch (e) {
    // 取消
  }
}

onMounted(loadPerms)
</script>

<style scoped>
.toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.toolbar-left {
  display: flex;
  gap: 8px;
}
.toolbar-right {
  display: flex;
  align-items: center;
}
</style>
