<template>
  <div class="app-container">
    <div class="page-card">
      <div class="toolbar">
        <div class="toolbar-left">
          <el-button type="primary" :icon="Plus" @click="openCreate">新建角色</el-button>
          <el-button :icon="Refresh" @click="loadRoles">刷新</el-button>
        </div>
        <div class="toolbar-right">
          <el-input
            v-model="search"
            placeholder="搜索角色名称/编码"
            clearable
            style="width: 240px"
            :prefix-icon="Search"
          />
        </div>
      </div>

      <el-table v-loading="loading" :data="filteredRoles" border stripe>
        <el-table-column type="index" label="#" width="50" />
        <el-table-column prop="id" label="ID" width="70" />
        <el-table-column prop="name" label="角色名称" min-width="140" />
        <el-table-column prop="code" label="角色编码" min-width="140" />
        <el-table-column prop="description" label="描述" min-width="200" show-overflow-tooltip />
        <el-table-column prop="sort" label="排序" width="80" />
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'info'">
              {{ row.status === 1 ? '启用' : '禁用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="240" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="openAssignPerms(row)">
              分配权限
            </el-button>
            <el-button type="warning" link size="small" @click="openEdit(row)">编辑</el-button>
            <el-button type="danger" link size="small" @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </div>

    <!-- 新建/编辑角色 -->
    <el-dialog v-model="dialogVisible" :title="isEdit ? '编辑角色' : '新建角色'" width="500px">
      <el-form ref="formRef" :model="form" :rules="formRules" label-width="80px">
        <el-form-item label="角色名称" prop="name">
          <el-input v-model="form.name" placeholder="请输入角色名称" />
        </el-form-item>
        <el-form-item v-if="!isEdit" label="角色编码" prop="code">
          <el-input v-model="form.code" placeholder="如 admin / editor" />
        </el-form-item>
        <el-form-item label="描述" prop="description">
          <el-input v-model="form.description" type="textarea" :rows="2" />
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

    <!-- 分配权限 -->
    <el-dialog v-model="permVisible" :title="`分配权限 - ${currentRow?.name || ''}`" width="560px">
      <el-input v-model="permSearch" placeholder="搜索权限名称/编码" clearable style="margin-bottom: 12px" />
      <div class="perm-list">
        <el-checkbox-group v-model="selectedPermIds">
          <div v-for="p in filteredPerms" :key="p.id" class="perm-check-item">
            <el-checkbox :value="p.id" :label="p.id">
              <span>{{ p.name }}</span>
              <el-tag size="small" type="info" style="margin-left: 8px">{{ typeText(p.type) }}</el-tag>
              <span class="perm-code">{{ p.code }}</span>
            </el-checkbox>
          </div>
        </el-checkbox-group>
      </div>
      <template #footer>
        <el-button @click="permVisible = false">取消</el-button>
        <el-button type="primary" :loading="assigning" @click="handleAssign">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Refresh, Search } from '@element-plus/icons-vue'
import {
  listRoles,
  createRole,
  updateRole,
  deleteRole,
  listPermissions,
  assignPermissions,
  getRolePermissions
} from '@/api/permission'

const loading = ref(false)
const search = ref('')
const roles = ref([])

const filteredRoles = computed(() => {
  const kw = search.value.trim().toLowerCase()
  if (!kw) return roles.value
  return roles.value.filter(
    (r) =>
      (r.name || '').toLowerCase().includes(kw) ||
      (r.code || '').toLowerCase().includes(kw)
  )
})

const loadRoles = async () => {
  loading.value = true
  try {
    const res = await listRoles()
    roles.value = res.data || []
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
  description: '',
  sort: 0,
  status: 1
})
const formRules = {
  name: [{ required: true, message: '请输入角色名称', trigger: 'blur' }],
  code: [{ required: true, message: '请输入角色编码', trigger: 'blur' }]
}

const openCreate = () => {
  isEdit.value = false
  Object.assign(form, { id: 0, name: '', code: '', description: '', sort: 0, status: 1 })
  dialogVisible.value = true
}

const openEdit = (row) => {
  isEdit.value = true
  Object.assign(form, {
    id: row.id,
    name: row.name,
    code: row.code,
    description: row.description || '',
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
        await updateRole(form.id, {
          name: form.name,
          description: form.description,
          sort: form.sort,
          status: form.status
        })
        ElMessage.success('更新成功')
      } else {
        await createRole({
          name: form.name,
          code: form.code,
          description: form.description,
          sort: form.sort,
          status: form.status
        })
        ElMessage.success('创建成功')
      }
      dialogVisible.value = false
      await loadRoles()
    } catch (e) {
      // ignore
    } finally {
      submitting.value = false
    }
  })
}

const handleDelete = async (row) => {
  try {
    await ElMessageBox.confirm(`确定删除角色 "${row.name}" 吗？`, '提示', { type: 'warning' })
    await deleteRole(row.id)
    ElMessage.success('删除成功')
    await loadRoles()
  } catch (e) {
    // 取消
  }
}

// ===== 分配权限 =====
const allPerms = ref([])
const permVisible = ref(false)
const permSearch = ref('')
const assigning = ref(false)
const currentRow = ref(null)
const selectedPermIds = ref([])

const filteredPerms = computed(() => {
  const kw = permSearch.value.trim().toLowerCase()
  if (!kw) return allPerms.value
  return allPerms.value.filter(
    (p) =>
      (p.name || '').toLowerCase().includes(kw) ||
      (p.code || '').toLowerCase().includes(kw)
  )
})

const typeText = (t) => ({ 1: '菜单', 2: '按钮', 3: '接口' }[t] || '-')

const openAssignPerms = async (row) => {
  currentRow.value = row
  selectedPermIds.value = []
  permSearch.value = ''
  permVisible.value = true
  // 并行加载所有权限 + 角色已分配的权限（用于回显）
  const tasks = []
  if (allPerms.value.length === 0) {
    tasks.push(
      listPermissions()
        .then((res) => (allPerms.value = res.data || []))
        .catch(() => (allPerms.value = []))
    )
  }
  tasks.push(
    getRolePermissions(row.id)
      .then((res) => {
        selectedPermIds.value = (res.data || []).map((p) => p.id)
      })
      .catch(() => (selectedPermIds.value = []))
  )
  await Promise.all(tasks)
}

const handleAssign = async () => {
  assigning.value = true
  try {
    await assignPermissions({
      role_id: currentRow.value.id,
      permission_ids: selectedPermIds.value
    })
    ElMessage.success('权限分配成功')
    permVisible.value = false
  } catch (e) {
    // ignore
  } finally {
    assigning.value = false
  }
}

onMounted(loadRoles)
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
.perm-list {
  max-height: 420px;
  overflow-y: auto;
  border: 1px solid #ebeef5;
  border-radius: 4px;
  padding: 8px 12px;
}
.perm-check-item {
  padding: 6px 0;
}
.perm-code {
  color: #909399;
  font-size: 12px;
  margin-left: 8px;
}
</style>
