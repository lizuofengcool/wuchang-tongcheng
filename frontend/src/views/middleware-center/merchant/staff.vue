<template>
  <div class="app-container">
    <div class="page-card">
      <div class="toolbar">
        <div class="toolbar-left">
          <el-button type="primary" :icon="Plus" @click="openCreate">添加员工</el-button>
          <el-button :icon="Refresh" @click="loadList">刷新</el-button>
        </div>
        <div class="toolbar-right">
          <el-input-number
            v-model="shopFilter"
            :min="0"
            placeholder="店铺ID"
            style="width: 130px"
            @change="onSearch"
          />
          <el-input-number
            v-model="userFilter"
            :min="0"
            placeholder="用户ID"
            style="width: 130px; margin-left: 8px"
            @change="onSearch"
          />
          <el-select
            v-model="roleFilter"
            placeholder="角色"
            clearable
            style="width: 140px; margin-left: 8px"
            @change="onSearch"
          >
            <el-option label="店主" value="owner" />
            <el-option label="管理员" value="manager" />
            <el-option label="店员" value="clerk" />
          </el-select>
          <el-select
            v-model="statusFilter"
            placeholder="状态"
            clearable
            style="width: 120px; margin-left: 8px"
            @change="onSearch"
          >
            <el-option label="在职" :value="1" />
            <el-option label="停用" :value="2" />
          </el-select>
          <el-button type="primary" :icon="Search" style="margin-left: 8px" @click="onSearch">搜索</el-button>
        </div>
      </div>

      <el-table v-loading="loading" :data="list" border stripe>
        <el-table-column prop="id" label="ID" width="70" />
        <el-table-column prop="shop_id" label="店铺ID" width="100" />
        <el-table-column prop="user_id" label="用户ID" width="100" />
        <el-table-column label="角色" width="110">
          <template #default="{ row }">
            <el-tag :type="roleTagType(row.role)" size="small">{{ row.role_text || roleText(row.role) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="权限" min-width="200" show-overflow-tooltip>
          <template #default="{ row }">
            <span v-if="row.permissions && Object.keys(row.permissions).length">
              {{ formatPermissions(row.permissions) }}
            </span>
            <span v-else class="text-muted">-</span>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="90">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'info'" size="small">
              {{ row.status_text || (row.status === 1 ? '在职' : '停用') }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="创建时间" width="170" />
        <el-table-column label="操作" width="280" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" size="small" @click="openEdit(row)">编辑</el-button>
            <el-button link type="warning" size="small" @click="openPermission(row)">权限</el-button>
            <el-button link type="success" size="small" @click="openRoleSwitch(row)">切换角色</el-button>
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
        <el-form-item label="所属店铺" prop="shop_id">
          <el-input-number v-model="form.shop_id" :min="1" :disabled="isEdit" style="width: 100%" />
        </el-form-item>
        <el-form-item label="用户ID" prop="user_id">
          <el-input-number v-model="form.user_id" :min="1" :disabled="isEdit" style="width: 100%" />
        </el-form-item>
        <el-form-item label="角色" prop="role">
          <el-radio-group v-model="form.role">
            <el-radio value="owner">店主</el-radio>
            <el-radio value="manager">管理员</el-radio>
            <el-radio value="clerk">店员</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="状态" prop="status">
          <el-radio-group v-model="form.status">
            <el-radio :value="1">在职</el-radio>
            <el-radio :value="2">停用</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="权限配置" prop="permission_input">
          <el-input
            v-model="form.permission_input"
            type="textarea"
            :rows="4"
            placeholder='JSON 格式，如 [{"code":"order:read","name":"查看订单","scope":"read"}]'
          />
          <div class="tip">提示：JSON 格式，每项包含 code/name/scope 字段</div>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="formVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="onSubmit">确定</el-button>
      </template>
    </el-dialog>

    <!-- 权限分配弹窗 -->
    <el-dialog v-model="permVisible" title="权限分配" width="560px">
      <el-form :model="permForm" label-width="100px">
        <el-form-item label="员工">
          <span>#{{ permForm.id }} (店铺 #{{ permForm.shop_id }}, 用户 #{{ permForm.user_id }})</span>
        </el-form-item>
        <el-form-item label="权限配置">
          <el-input
            v-model="permForm.permission_input"
            type="textarea"
            :rows="8"
            placeholder='JSON 格式，如 [{"code":"order:read","name":"查看订单","scope":"read"}]'
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="permVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="onSubmitPermission">确定</el-button>
      </template>
    </el-dialog>

    <!-- 角色切换弹窗 -->
    <el-dialog v-model="roleVisible" title="切换角色" width="420px">
      <el-form :model="roleForm" label-width="80px">
        <el-form-item label="员工">
          <span>#{{ roleForm.id }}</span>
        </el-form-item>
        <el-form-item label="当前角色">
          <el-tag :type="roleTagType(roleForm.current)" size="small">{{ roleText(roleForm.current) }}</el-tag>
        </el-form-item>
        <el-form-item label="新角色">
          <el-radio-group v-model="roleForm.role">
            <el-radio value="owner">店主</el-radio>
            <el-radio value="manager">管理员</el-radio>
            <el-radio value="clerk">店员</el-radio>
          </el-radio-group>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="roleVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="onSubmitRole">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Refresh, Search } from '@element-plus/icons-vue'
import {
  getMerchantStaffList,
  createMerchantStaff,
  updateMerchantStaff,
  deleteMerchantStaff,
  assignMerchantStaffPermissions,
  switchMerchantStaffRole
} from '@/api/merchant'

const loading = ref(false)
const list = ref([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(10)
const shopFilter = ref(0)
const userFilter = ref(0)
const roleFilter = ref('')
const statusFilter = ref(null)

const formVisible = ref(false)
const isEdit = ref(false)
const submitting = ref(false)
const formRef = ref(null)
const form = reactive({
  id: 0,
  shop_id: 1,
  user_id: 1,
  role: 'clerk',
  status: 1,
  permission_input: ''
})
const rules = {
  shop_id: [{ required: true, message: '请输入店铺ID', trigger: 'blur' }],
  user_id: [{ required: true, message: '请输入用户ID', trigger: 'blur' }],
  role: [{ required: true, message: '请选择角色', trigger: 'change' }]
}

const permVisible = ref(false)
const permForm = reactive({ id: 0, shop_id: 0, user_id: 0, permission_input: '' })

const roleVisible = ref(false)
const roleForm = reactive({ id: 0, current: 'clerk', role: 'clerk' })

function roleText(role) {
  return role === 'owner' ? '店主' : role === 'manager' ? '管理员' : role === 'clerk' ? '店员' : '-'
}

function roleTagType(role) {
  return role === 'owner' ? 'danger' : role === 'manager' ? 'warning' : 'info'
}

function formatPermissions(perms) {
  if (!perms) return '-'
  if (Array.isArray(perms)) {
    return perms.map((p) => (typeof p === 'string' ? p : p.code || p.name || '')).filter(Boolean).join(', ')
  }
  if (typeof perms === 'object') {
    return Object.keys(perms).join(', ')
  }
  return String(perms)
}

function parsePermissions(text) {
  if (!text || !text.trim()) return []
  try {
    const parsed = JSON.parse(text)
    return Array.isArray(parsed) ? parsed : parsed
  } catch (e) {
    return text.split(',').map((s) => s.trim()).filter(Boolean)
  }
}

function stringifyPermissions(perms) {
  if (!perms) return ''
  if (Array.isArray(perms) || typeof perms === 'object') {
    try {
      return JSON.stringify(perms, null, 2)
    } catch (e) {
      return ''
    }
  }
  return String(perms)
}

async function loadList() {
  loading.value = true
  try {
    const params = { page: page.value, page_size: pageSize.value }
    if (shopFilter.value > 0) params.shop_id = shopFilter.value
    if (userFilter.value > 0) params.user_id = userFilter.value
    if (roleFilter.value) params.role = roleFilter.value
    if (statusFilter.value !== null && statusFilter.value !== '') params.status = statusFilter.value
    const res = await getMerchantStaffList(params)
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
  Object.assign(form, {
    id: row.id,
    shop_id: row.shop_id,
    user_id: row.user_id,
    role: row.role,
    status: row.status,
    permission_input: stringifyPermissions(row.permissions)
  })
  formVisible.value = true
}

function resetForm() {
  Object.assign(form, {
    id: 0, shop_id: 1, user_id: 1, role: 'clerk', status: 1, permission_input: ''
  })
}

async function onSubmit() {
  if (!formRef.value) return
  await formRef.value.validate(async (valid) => {
    if (!valid) return
    submitting.value = true
    try {
      const payload = {
        shop_id: form.shop_id,
        user_id: form.user_id,
        role: form.role,
        status: form.status,
        permissions: parsePermissions(form.permission_input)
      }
      if (isEdit.value) {
        await updateMerchantStaff(form.id, payload)
        ElMessage.success('更新成功')
      } else {
        await createMerchantStaff(payload)
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

function openPermission(row) {
  permForm.id = row.id
  permForm.shop_id = row.shop_id
  permForm.user_id = row.user_id
  permForm.permission_input = stringifyPermissions(row.permissions)
  permVisible.value = true
}

async function onSubmitPermission() {
  submitting.value = true
  try {
    await assignMerchantStaffPermissions(permForm.id, {
      permissions: parsePermissions(permForm.permission_input)
    })
    ElMessage.success('权限已更新')
    permVisible.value = false
    loadList()
  } catch (e) {
    ElMessage.error(e.message || '操作失败')
  } finally {
    submitting.value = false
  }
}

function openRoleSwitch(row) {
  roleForm.id = row.id
  roleForm.current = row.role
  roleForm.role = row.role
  roleVisible.value = true
}

async function onSubmitRole() {
  submitting.value = true
  try {
    await switchMerchantStaffRole(roleForm.id, { role: roleForm.role })
    ElMessage.success('角色已切换')
    roleVisible.value = false
    loadList()
  } catch (e) {
    ElMessage.error(e.message || '操作失败')
  } finally {
    submitting.value = false
  }
}

async function onDelete(row) {
  try {
    await ElMessageBox.confirm(`确定要删除员工「#${row.id}」吗？`, '警告', { type: 'warning' })
    await deleteMerchantStaff(row.id)
    ElMessage.success('删除成功')
    loadList()
  } catch (e) {
    if (e !== 'cancel') ElMessage.error('删除失败')
  }
}

onMounted(() => loadList())
</script>

<style scoped>
.pagination-wrap { margin-top: 16px; display: flex; justify-content: flex-end; }
.tip { color: #909399; font-size: 12px; margin-top: 4px; }
.text-muted { color: #c0c4cc; }
</style>
