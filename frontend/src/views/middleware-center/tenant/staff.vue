<template>
  <div class="app-container">
    <div class="page-card">
      <div class="toolbar">
        <div class="toolbar-left">
          <el-button type="primary" :icon="Plus" @click="openCreate">添加员工</el-button>
          <el-button :icon="Refresh" @click="loadList">刷新</el-button>
        </div>
        <div class="toolbar-right">
          <el-input
            v-model="search"
            placeholder="姓名/手机号"
            clearable
            style="width: 180px"
            :prefix-icon="Search"
            @keyup.enter="onSearch"
            @clear="onSearch"
          />
          <el-select
            v-model="stationFilter"
            placeholder="所属分站"
            clearable
            filterable
            style="width: 200px; margin-left: 8px"
            @change="onSearch"
          >
            <el-option v-for="s in stations" :key="s.id" :label="s.name" :value="s.id" />
          </el-select>
          <el-select
            v-model="roleFilter"
            placeholder="角色"
            clearable
            style="width: 140px; margin-left: 8px"
            @change="onSearch"
          >
            <el-option label="运营人员" :value="1" />
            <el-option label="管理员" :value="2" />
          </el-select>
          <el-select
            v-model="statusFilter"
            placeholder="状态"
            clearable
            style="width: 120px; margin-left: 8px"
            @change="onSearch"
          >
            <el-option label="已启用" :value="1" />
            <el-option label="已停用" :value="0" />
          </el-select>
          <el-button type="primary" :icon="Search" style="margin-left: 8px" @click="onSearch">搜索</el-button>
        </div>
      </div>

      <el-table v-loading="loading" :data="list" border stripe>
        <el-table-column prop="id" label="ID" width="70" />
        <el-table-column label="所属分站" min-width="160">
          <template #default="{ row }">{{ stationName(row.station_id) }}</template>
        </el-table-column>
        <el-table-column prop="user_id" label="用户ID" width="100" />
        <el-table-column label="姓名" min-width="120">
          <template #default="{ row }">{{ row.username || '-' }}</template>
        </el-table-column>
        <el-table-column label="手机号" min-width="130">
          <template #default="{ row }">{{ row.phone || '-' }}</template>
        </el-table-column>
        <el-table-column label="角色" width="110">
          <template #default="{ row }">
            <el-tag :type="row.role === 2 ? 'danger' : 'warning'" size="small">{{ roleText(row.role) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="权限项" min-width="160" show-overflow-tooltip>
          <template #default="{ row }">
            <span v-if="row.permissions && row.permissions.length">{{ row.permissions.join(', ') }}</span>
            <span v-else class="text-muted">-</span>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="90">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'info'" size="small">{{ row.status === 1 ? '启用' : '停用' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="创建时间" width="170" />
        <el-table-column label="操作" width="220" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" size="small" @click="openEdit(row)">编辑</el-button>
            <el-button link :type="row.status === 1 ? 'warning' : 'success'" size="small" @click="toggleStatus(row)">
              {{ row.status === 1 ? '停用' : '启用' }}
            </el-button>
            <el-button link type="danger" size="small" @click="onDelete(row)">删除</el-button>
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
          @size-change="loadList"
          @current-change="loadList"
        />
      </div>
    </div>

    <!-- 新建/编辑弹窗 -->
    <el-dialog v-model="formVisible" :title="isEdit ? '编辑员工' : '添加员工'" width="560px" @closed="resetForm">
      <el-form ref="formRef" :model="form" :rules="rules" label-width="100px">
        <el-form-item label="所属分站" prop="station_id">
          <el-select v-model="form.station_id" filterable placeholder="请选择分站" style="width: 100%" :disabled="isEdit">
            <el-option v-for="s in stations" :key="s.id" :label="`${s.name}（ID:${s.id}）`" :value="s.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="用户ID" prop="user_id">
          <el-input-number v-model="form.user_id" :min="1" :disabled="isEdit" style="width: 100%" />
        </el-form-item>
        <el-form-item label="角色" prop="role">
          <el-radio-group v-model="form.role">
            <el-radio :value="1">运营人员</el-radio>
            <el-radio :value="2">管理员</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="状态" prop="status">
          <el-radio-group v-model="form.status">
            <el-radio :value="1">启用</el-radio>
            <el-radio :value="0">停用</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="权限项" prop="permission_input">
          <el-input
            v-model="form.permission_input"
            type="textarea"
            :rows="3"
            placeholder="多个权限用英文逗号分隔，如 station:read, staff:manage"
          />
          <div class="permission-tip">提示：管理员默认拥有全部权限，运营人员按勾选权限操作。</div>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="formVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="onSubmit">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Refresh, Search } from '@element-plus/icons-vue'
import {
  getTenantStaffList,
  createTenantStaff,
  updateTenantStaff,
  deleteTenantStaff,
  getTenantStationAdminList
} from '@/api/tenant'

const loading = ref(false)
const list = ref([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(10)
const search = ref('')
const stationFilter = ref(null)
const roleFilter = ref(null)
const statusFilter = ref(null)
const stations = ref([])

const formVisible = ref(false)
const isEdit = ref(false)
const submitting = ref(false)
const formRef = ref(null)
const form = reactive({
  id: 0,
  station_id: null,
  user_id: 1,
  role: 1,
  status: 1,
  permission_input: ''
})
const rules = {
  station_id: [{ required: true, message: '请选择分站', trigger: 'change' }],
  user_id: [{ required: true, message: '请输入用户ID', trigger: 'blur' }],
  role: [{ required: true, message: '请选择角色', trigger: 'change' }]
}

function roleText(role) {
  return role === 2 ? '管理员' : role === 1 ? '运营人员' : '-'
}

function stationName(id) {
  const s = stations.value.find((x) => x.id === id)
  return s ? s.name : `#${id}`
}

async function loadStations() {
  try {
    const res = await getTenantStationAdminList({ page: 1, page_size: 200 })
    stations.value = res.data?.list || []
  } catch (e) {
    stations.value = []
  }
}

async function loadList() {
  loading.value = true
  try {
    const params = { page: page.value, page_size: pageSize.value, keyword: search.value }
    if (stationFilter.value) params.station_id = stationFilter.value
    if (roleFilter.value !== null && roleFilter.value !== '') params.role = roleFilter.value
    if (statusFilter.value !== null && statusFilter.value !== '') params.status = statusFilter.value
    const res = await getTenantStaffList(params)
    list.value = res.data?.list || []
    total.value = res.data?.total || 0
  } catch (e) {
    ElMessage.error('加载员工列表失败')
  } finally {
    loading.value = false
  }
}

function onSearch() {
  page.value = 1
  loadList()
}

function openCreate() {
  isEdit.value = false
  resetForm()
  formVisible.value = true
}

function openEdit(row) {
  isEdit.value = true
  const perms = Array.isArray(row.permissions) ? row.permissions : []
  Object.assign(form, {
    id: row.id,
    station_id: row.station_id,
    user_id: row.user_id,
    role: row.role,
    status: row.status,
    permission_input: perms.join(', ')
  })
  formVisible.value = true
}

function resetForm() {
  Object.assign(form, {
    id: 0, station_id: null, user_id: 1, role: 1, status: 1, permission_input: ''
  })
}

function parsePermissions(text) {
  if (!text || !text.trim()) return []
  return text.split(',').map((s) => s.trim()).filter(Boolean)
}

async function onSubmit() {
  if (!formRef.value) return
  await formRef.value.validate(async (valid) => {
    if (!valid) return
    submitting.value = true
    try {
      const payload = {
        station_id: form.station_id,
        user_id: form.user_id,
        role: form.role,
        status: form.status,
        permissions: parsePermissions(form.permission_input)
      }
      if (isEdit.value) {
        await updateTenantStaff(form.id, payload)
        ElMessage.success('更新成功')
      } else {
        await createTenantStaff(payload)
        ElMessage.success('添加成功')
      }
      formVisible.value = false
      loadList()
    } catch (e) {
      ElMessage.error(e.message || '操作失败')
    } finally {
      submitting.value = false
    }
  })
}

async function toggleStatus(row) {
  const next = row.status === 1 ? 0 : 1
  try {
    await ElMessageBox.confirm(`确定要${next === 1 ? '启用' : '停用'}员工「${row.username || row.user_id}」吗？`, '提示', { type: 'warning' })
    await updateTenantStaff(row.id, { status: next, role: row.role, permissions: row.permissions || [] })
    ElMessage.success('状态已更新')
    loadList()
  } catch (e) {
    if (e !== 'cancel') ElMessage.error('操作失败')
  }
}

async function onDelete(row) {
  try {
    await ElMessageBox.confirm(`确定要删除员工「${row.username || row.user_id}」吗？`, '警告', { type: 'warning' })
    await deleteTenantStaff(row.id)
    ElMessage.success('删除成功')
    loadList()
  } catch (e) {
    if (e !== 'cancel') ElMessage.error('删除失败')
  }
}

onMounted(async () => {
  await loadStations()
  loadList()
})
</script>

<style scoped>
.pagination-wrap { margin-top: 16px; display: flex; justify-content: flex-end; }
.permission-tip { color: #909399; font-size: 12px; margin-top: 4px; }
.text-muted { color: #c0c4cc; }
</style>
