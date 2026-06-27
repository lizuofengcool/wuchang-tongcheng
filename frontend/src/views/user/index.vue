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
            placeholder="搜索用户名/昵称"
            clearable
            style="width: 240px"
            :prefix-icon="Search"
            @keyup.enter="loadUsers"
            @clear="loadUsers"
          />
        </div>
      </div>

      <el-table v-loading="loading" :data="filteredUsers" border stripe>
        <el-table-column type="index" label="#" width="50" />
        <el-table-column prop="id" label="ID" width="70" />
        <el-table-column prop="username" label="用户名" min-width="120" />
        <el-table-column prop="nickname" label="昵称" min-width="120" />
        <el-table-column prop="phone" label="手机号" min-width="120" />
        <el-table-column prop="email" label="邮箱" min-width="160" />
        <el-table-column label="性别" width="80">
          <template #default="{ row }">
            {{ genderText(row.gender) }}
          </template>
        </el-table-column>
        <el-table-column label="角色" min-width="180">
          <template #default="{ row }">
            <el-tag
              v-for="r in row._roles"
              :key="r.id"
              type="success"
              size="small"
              style="margin-right: 4px"
            >
              {{ r.name }}
            </el-tag>
            <span v-if="!row._roles?.length" class="text-muted">未分配</span>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="220" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="openAssignRoles(row)">
              分配角色
            </el-button>
            <el-button type="danger" link size="small" @click="handleDelete(row)">
              删除
            </el-button>
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
        />
      </div>
    </div>

    <!-- 新建用户 -->
    <el-dialog v-model="createVisible" title="新建用户" width="480px">
      <el-form ref="createFormRef" :model="createForm" :rules="createRules" label-width="80px">
        <el-form-item label="用户名" prop="username">
          <el-input v-model="createForm.username" placeholder="3-50位字符" />
        </el-form-item>
        <el-form-item label="密码" prop="password">
          <el-input v-model="createForm.password" type="password" show-password placeholder="6-50位" />
        </el-form-item>
        <el-form-item label="昵称" prop="nickname">
          <el-input v-model="createForm.nickname" />
        </el-form-item>
        <el-form-item label="手机号" prop="phone">
          <el-input v-model="createForm.phone" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="createVisible = false">取消</el-button>
        <el-button type="primary" :loading="creating" @click="handleCreate">确定</el-button>
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
import { ref, reactive, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Refresh, Search } from '@element-plus/icons-vue'
import { register } from '@/api/user'
import { listRoles, assignRoles, getUserRoles } from '@/api/permission'

const loading = ref(false)
const search = ref('')
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)

// 后端 user 模块当前未提供列表接口，本地维护注册过的用户列表（id 自增）
// 这里使用 localStorage 缓存已注册用户，避免重复拉取
const users = ref([])

const filteredUsers = computed(() => {
  const kw = search.value.trim().toLowerCase()
  if (!kw) return users.value
  return users.value.filter(
    (u) =>
      (u.username || '').toLowerCase().includes(kw) ||
      (u.nickname || '').toLowerCase().includes(kw)
  )
})

const genderText = (g) => ({ 0: '未知', 1: '男', 2: '女' }[g] || '未知')

const loadUsers = async () => {
  loading.value = true
  try {
    // 从 localStorage 读取本地维护的用户清单
    const cached = JSON.parse(localStorage.getItem('admin_users') || '[]')
    users.value = cached
    total.value = cached.length
    // 拉取每个用户的角色
    await Promise.all(
      users.value.map(async (u) => {
        try {
          const res = await getUserRoles(u.id)
          u._roles = res.data || []
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

// ===== 新建用户 =====
const createVisible = ref(false)
const creating = ref(false)
const createFormRef = ref(null)
const createForm = reactive({
  username: '',
  password: '',
  nickname: '',
  phone: ''
})
const createRules = {
  username: [
    { required: true, message: '请输入用户名', trigger: 'blur' },
    { min: 3, max: 50, message: '3-50位字符', trigger: 'blur' }
  ],
  password: [
    { required: true, message: '请输入密码', trigger: 'blur' },
    { min: 6, max: 50, message: '6-50位', trigger: 'blur' }
  ],
  phone: [{ pattern: /^1\d{10}$/, message: '请输入正确的手机号', trigger: 'blur' }]
}

const openCreate = () => {
  Object.assign(createForm, { username: '', password: '', nickname: '', phone: '' })
  createVisible.value = true
}

const handleCreate = async () => {
  if (!createFormRef.value) return
  await createFormRef.value.validate(async (valid) => {
    if (!valid) return
    creating.value = true
    try {
      const res = await register({
        username: createForm.username,
        password: createForm.password,
        nickname: createForm.nickname,
        phone: createForm.phone
      })
      // 后端返回的 data 是 user_info（含 id 等）
      const newUser = res.data || {}
      newUser._roles = []
      const list = JSON.parse(localStorage.getItem('admin_users') || '[]')
      list.push(newUser)
      localStorage.setItem('admin_users', JSON.stringify(list))
      ElMessage.success('用户创建成功')
      createVisible.value = false
      await loadUsers()
    } catch (e) {
      // ignore
    } finally {
      creating.value = false
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

// ===== 删除（本地清单移除） =====
const handleDelete = async (row) => {
  try {
    await ElMessageBox.confirm(
      `确定从管理列表移除用户 "${row.username}" 吗？（不会删除后端账号）`,
      '提示',
      { type: 'warning' }
    )
    const list = JSON.parse(localStorage.getItem('admin_users') || '[]')
    const next = list.filter((u) => u.id !== row.id)
    localStorage.setItem('admin_users', JSON.stringify(next))
    ElMessage.success('已从列表移除')
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
}
.toolbar-left {
  display: flex;
  gap: 8px;
}
</style>
