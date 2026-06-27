<template>
  <div class="app-container">
    <div class="page-card">
      <div class="toolbar">
        <div class="toolbar-left">
          <el-button type="primary" :icon="Plus" @click="openCreate">新建用户</el-button>
          <el-button :icon="Refresh" @click="loadUsers">刷新</el-button>
        </div>
        <div class="toolbar-right">
          <el-input
            v-model="search"
            placeholder="用户名/昵称"
            clearable
            style="width: 200px"
            :prefix-icon="Search"
            @keyup.enter="onSearch"
            @clear="onSearch"
          />
          <el-select v-model="statusFilter" placeholder="状态" clearable style="width: 120px; margin-left: 8px" @change="onSearch">
            <el-option label="正常" :value="1" />
            <el-option label="禁用" :value="0" />
          </el-select>
          <el-button type="primary" :icon="Search" style="margin-left: 8px" @click="onSearch">搜索</el-button>
        </div>
      </div>

      <el-table v-loading="loading" :data="users" border stripe>
        <el-table-column prop="id" label="ID" width="70" />
        <el-table-column prop="username" label="用户名" min-width="120" />
        <el-table-column prop="nickname" label="昵称" min-width="120" />
        <el-table-column prop="phone" label="手机号" min-width="120" />
        <el-table-column prop="email" label="邮箱" min-width="160" show-overflow-tooltip />
        <el-table-column label="性别" width="80">
          <template #default="{ row }">{{ genderText(row.gender) }}</template>
        </el-table-column>
        <el-table-column label="角色" min-width="180">
          <template #default="{ row }">
            <el-tag
              v-for="r in row._roles"
              :key="r.id"
              type="success"
              size="small"
              style="margin-right: 4px"
            >{{ r.name }}</el-tag>
            <span v-if="!row._roles?.length" class="text-muted">未分配</span>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-switch
              :model-value="row.status === 1"
              @change="(val) => handleToggleStatus(row, val)"
            />
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="创建时间" width="170">
          <template #default="{ row }">{{ formatTime(row.created_at) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="240" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="openAssignRoles(row)">分配角色</el-button>
            <el-button type="warning" link size="small" @click="openEdit(row)">编辑</el-button>
            <el-button type="info" link size="small" @click="openResetPwd(row)">重置密码</el-button>
            <el-button type="danger" link size="small" @click="handleDelete(row)">删除</el-button>
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
          @current-change="loadUsers"
          @size-change="loadUsers"
        />
      </div>
    </div>

    <!-- 新建/编辑用户 -->
    <el-dialog v-model="dialogVisible" :title="isEdit ? '编辑用户' : '新建用户'" width="500px">
      <el-form ref="formRef" :model="form" :rules="formRules" label-width="80px">
        <el-form-item label="用户名" prop="username">
          <el-input v-model="form.username" :disabled="isEdit" placeholder="3-50位字符" />
        </el-form-item>
        <el-form-item v-if="!isEdit" label="密码" prop="password">
          <el-input v-model="form.password" type="password" show-password placeholder="6-50位" />
        </el-form-item>
        <el-form-item label="昵称" prop="nickname">
          <el-input v-model="form.nickname" />
        </el-form-item>
        <el-form-item label="手机号" prop="phone">
          <el-input v-model="form.phone" />
        </el-form-item>
        <el-form-item label="邮箱" prop="email">
          <el-input v-model="form.email" />
        </el-form-item>
        <el-form-item label="性别" prop="gender">
          <el-radio-group v-model="form.gender">
            <el-radio :value="0">未知</el-radio>
            <el-radio :value="1">男</el-radio>
            <el-radio :value="2">女</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item v-if="!isEdit" label="状态" prop="status">
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

    <!-- 重置密码 -->
    <el-dialog v-model="pwdVisible" :title="`重置密码 - ${currentRow?.username || ''}`" width="420px">
      <el-form ref="pwdFormRef" :model="pwdForm" :rules="pwdRules" label-width="90px">
        <el-form-item label="新密码" prop="new_password">
          <el-input v-model="pwdForm.new_password" type="password" show-password placeholder="6-50位" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="pwdVisible = false">取消</el-button>
        <el-button type="primary" :loading="resetting" @click="handleResetPwd">确定</el-button>
      </template>
    </el-dialog>

    <!-- 分配角色 -->
    <el-dialog v-model="assignVisible" :title="`分配角色 - ${currentRow?.username || ''}`" width="480px">
      <el-checkbox-group v-model="selectedRoleIds">
        <div v-for="r in allRoles" :key="r.id" class="role-check-item">
          <el-checkbox :value="r.id" :label="r.id">
            {{ r.name }}
            <span class="role-code">{{ r.code }}</span>
          </el-checkbox>
        </div>
      </el-checkbox-group>
      <template #footer>
        <el-button @click="assignVisible = false">取消</el-button>
        <el-button type="primary" :loading="assigning" @click="handleAssign">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Refresh, Search } from '@element-plus/icons-vue'
import {
  listUsers,
  adminCreateUser,
  adminUpdateUser,
  updateUserStatus,
  resetUserPassword,
  deleteUser
} from '@/api/user'
import { listRoles, assignRoles, getUserRoles } from '@/api/permission'
import { formatTime } from '@/utils/format'

const loading = ref(false)
const search = ref('')
const statusFilter = ref(null)
const page = ref(1)
const pageSize = ref(10)
const total = ref(0)
const users = ref([])

const genderText = (g) => ({ 0: '未知', 1: '男', 2: '女' }[g] || '未知')

const onSearch = () => {
  page.value = 1
  loadUsers()
}

const loadUsers = async () => {
  loading.value = true
  try {
    const res = await listUsers({
      page: page.value,
      page_size: pageSize.value,
      keyword: search.value.trim(),
      status: statusFilter.value === null || statusFilter.value === '' ? -1 : statusFilter.value
    })
    const data = res.data || {}
    users.value = data.list || []
    total.value = data.total || 0
    // 拉取每个用户的角色（用于展示）
    await Promise.all(
      users.value.map(async (u) => {
        try {
          const r = await getUserRoles(u.id)
          u._roles = r.data || []
        } catch (e) {
          u._roles = []
        }
      })
    )
    users.value = [...users.value]
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
  username: '',
  password: '',
  nickname: '',
  phone: '',
  email: '',
  gender: 0,
  status: 1
})
const formRules = {
  username: [
    { required: true, message: '请输入用户名', trigger: 'blur' },
    { min: 3, max: 50, message: '3-50位字符', trigger: 'blur' }
  ],
  password: [
    { required: true, message: '请输入密码', trigger: 'blur' },
    { min: 6, max: 50, message: '6-50位', trigger: 'blur' }
  ],
  phone: [{ pattern: /^1\d{10}$/, message: '请输入正确的手机号', trigger: 'blur' }],
  email: [{ type: 'email', message: '请输入正确的邮箱', trigger: 'blur' }]
}

const openCreate = () => {
  isEdit.value = false
  Object.assign(form, { id: 0, username: '', password: '', nickname: '', phone: '', email: '', gender: 0, status: 1 })
  dialogVisible.value = true
}

const openEdit = (row) => {
  isEdit.value = true
  Object.assign(form, {
    id: row.id,
    username: row.username,
    password: '',
    nickname: row.nickname || '',
    phone: row.phone || '',
    email: row.email || '',
    gender: row.gender ?? 0,
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
        await adminUpdateUser(form.id, {
          nickname: form.nickname,
          phone: form.phone,
          email: form.email,
          gender: form.gender
        })
        ElMessage.success('更新成功')
      } else {
        await adminCreateUser({
          username: form.username,
          password: form.password,
          nickname: form.nickname,
          phone: form.phone,
          email: form.email,
          gender: form.gender,
          status: form.status
        })
        ElMessage.success('创建成功')
      }
      dialogVisible.value = false
      await loadUsers()
    } catch (e) {
      // ignore
    } finally {
      submitting.value = false
    }
  })
}

// ===== 状态切换 =====
const handleToggleStatus = async (row, val) => {
  const status = val ? 1 : 0
  try {
    await updateUserStatus(row.id, status)
    row.status = status
    ElMessage.success(status === 1 ? '已启用' : '已禁用')
  } catch (e) {
    // ignore
  }
}

// ===== 重置密码 =====
const pwdVisible = ref(false)
const pwdFormRef = ref(null)
const pwdForm = reactive({ new_password: '' })
const pwdRules = {
  new_password: [
    { required: true, message: '请输入新密码', trigger: 'blur' },
    { min: 6, max: 50, message: '6-50位', trigger: 'blur' }
  ]
}
const resetting = ref(false)
const openResetPwd = (row) => {
  currentRow.value = row
  pwdForm.new_password = ''
  pwdVisible.value = true
}
const handleResetPwd = async () => {
  if (!pwdFormRef.value) return
  await pwdFormRef.value.validate(async (valid) => {
    if (!valid) return
    resetting.value = true
    try {
      await resetUserPassword(currentRow.value.id, pwdForm.new_password)
      ElMessage.success('密码重置成功')
      pwdVisible.value = false
    } catch (e) {
      // ignore
    } finally {
      resetting.value = false
    }
  })
}

// ===== 分配角色 =====
const allRoles = ref([])
const assignVisible = ref(false)
const assigning = ref(false)
const currentRow = ref(null)
const selectedRoleIds = ref([])

const openAssignRoles = (row) => {
  currentRow.value = row
  selectedRoleIds.value = (row._roles || []).map((r) => r.id)
  assignVisible.value = true
}

const handleAssign = async () => {
  assigning.value = true
  try {
    await assignRoles({
      user_id: currentRow.value.id,
      role_ids: selectedRoleIds.value
    })
    ElMessage.success('角色分配成功')
    assignVisible.value = false
    await loadUsers()
  } catch (e) {
    // ignore
  } finally {
    assigning.value = false
  }
}

const handleDelete = async (row) => {
  try {
    await ElMessageBox.confirm(`确定删除用户 "${row.username}" 吗？`, '提示', { type: 'warning' })
    await deleteUser(row.id)
    ElMessage.success('删除成功')
    await loadUsers()
  } catch (e) {
    // 取消
  }
}

const loadRoles = async () => {
  try {
    const res = await listRoles()
    allRoles.value = res.data || []
  } catch (e) {
    allRoles.value = []
  }
}

onMounted(async () => {
  await loadRoles()
  await loadUsers()
})
</script>

<style scoped>
.role-check-item {
  padding: 6px 0;
}
.role-code {
  color: #909399;
  font-size: 12px;
  margin-left: 8px;
}
.text-muted {
  color: #c0c4cc;
}
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
