<template>
  <div class="app-container">
    <div class="page-card">
      <div class="toolbar">
        <div class="toolbar-left">
          <el-button type="primary" :icon="Plus" @click="openCreate(0)">新建根类目</el-button>
          <el-button :icon="Refresh" @click="loadTree">刷新</el-button>
          <el-button :icon="Tickets" @click="toggleMode">{{ treeMode ? '切换为列表' : '切换为树形' }}</el-button>
        </div>
        <div class="toolbar-right">
          <el-input
            v-model="search"
            placeholder="类目名"
            clearable
            style="width: 200px"
            :prefix-icon="Search"
            @keyup.enter="onSearch"
            @clear="onSearch"
          />
          <el-select
            v-model="statusFilter"
            placeholder="状态"
            clearable
            style="width: 120px; margin-left: 8px"
            @change="onSearch"
          >
            <el-option label="启用" :value="1" />
            <el-option label="禁用" :value="0" />
          </el-select>
          <el-button type="primary" :icon="Search" style="margin-left: 8px" @click="onSearch">搜索</el-button>
        </div>
      </div>

      <!-- 树形展示 -->
      <el-table
        v-if="treeMode"
        v-loading="loading"
        :data="treeData"
        row-key="id"
        border
        stripe
        default-expand-all
        :tree-props="{ children: 'children' }"
      >
        <el-table-column prop="id" label="ID" width="70" />
        <el-table-column prop="name" label="类目名" min-width="200">
          <template #default="{ row }">
            <el-icon v-if="row.icon"><Picture /></el-icon>
            <span style="margin-left: 4px;">{{ row.name || '-' }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="parent_id" label="父ID" width="80" />
        <el-table-column prop="sort" label="排序" width="80" />
        <el-table-column label="状态" width="90">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'info'" size="small">
              {{ row.status_text || (row.status === 1 ? '启用' : '禁用') }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="创建时间" width="170" />
        <el-table-column label="操作" width="280" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" size="small" @click="openCreate(row.id)">添加子类目</el-button>
            <el-button link type="warning" size="small" @click="openEdit(row)">编辑</el-button>
            <el-button
              link
              :type="row.status === 1 ? 'info' : 'success'"
              size="small"
              @click="onToggleStatus(row)"
            >{{ row.status === 1 ? '禁用' : '启用' }}</el-button>
            <el-button link type="danger" size="small" @click="onDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>

      <!-- 列表展示 -->
      <el-table v-else v-loading="loading" :data="list" border stripe>
        <el-table-column prop="id" label="ID" width="70" />
        <el-table-column prop="parent_id" label="父ID" width="80" />
        <el-table-column prop="name" label="类目名" min-width="180" />
        <el-table-column prop="icon" label="图标" min-width="160" show-overflow-tooltip />
        <el-table-column prop="sort" label="排序" width="80" />
        <el-table-column label="状态" width="90">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'info'" size="small">
              {{ row.status_text || (row.status === 1 ? '启用' : '禁用') }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="创建时间" width="170" />
        <el-table-column label="操作" width="280" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" size="small" @click="openCreate(row.id)">添加子类目</el-button>
            <el-button link type="warning" size="small" @click="openEdit(row)">编辑</el-button>
            <el-button
              link
              :type="row.status === 1 ? 'info' : 'success'"
              size="small"
              @click="onToggleStatus(row)"
            >{{ row.status === 1 ? '禁用' : '启用' }}</el-button>
            <el-button link type="danger" size="small" @click="onDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>

      <div v-if="!treeMode" class="pagination-wrap">
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
    <el-dialog v-model="formVisible" :title="isEdit ? '编辑类目' : '新建类目'" width="520px" @closed="resetForm">
      <el-form ref="formRef" :model="form" :rules="rules" label-width="100px">
        <el-form-item label="父类目" prop="parent_id">
          <el-input-number v-model="form.parent_id" :min="0" :disabled="isEdit" style="width: 100%" />
          <div class="tip">0 表示根类目</div>
        </el-form-item>
        <el-form-item label="类目名" prop="name">
          <el-input v-model="form.name" maxlength="64" placeholder="请输入类目名" />
        </el-form-item>
        <el-form-item label="图标" prop="icon">
          <el-input v-model="form.icon" maxlength="255" placeholder="图标 URL" />
        </el-form-item>
        <el-form-item label="排序" prop="sort">
          <el-input-number v-model="form.sort" :min="0" style="width: 100%" />
        </el-form-item>
        <el-form-item label="状态" prop="status">
          <el-radio-group v-model="form.status">
            <el-radio :value="1">启用</el-radio>
            <el-radio :value="0">禁用</el-radio>
          </el-radio-group>
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
import { Plus, Refresh, Search, Tickets, Picture } from '@element-plus/icons-vue'
import {
  getMerchantCategoryTree,
  getMerchantCategoryList,
  createMerchantCategory,
  updateMerchantCategory,
  deleteMerchantCategory,
  updateMerchantCategoryStatus
} from '@/api/merchant'

const loading = ref(false)
const treeMode = ref(true)
const treeData = ref([])
const list = ref([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(10)
const search = ref('')
const statusFilter = ref(null)

const formVisible = ref(false)
const isEdit = ref(false)
const submitting = ref(false)
const formRef = ref(null)
const form = reactive({
  id: 0,
  parent_id: 0,
  name: '',
  icon: '',
  sort: 0,
  status: 1
})
const rules = {
  name: [{ required: true, message: '请输入类目名', trigger: 'blur' }]
}

function toggleMode() {
  treeMode.value = !treeMode.value
  if (treeMode.value) {
    loadTree()
  } else {
    loadList()
  }
}

async function loadTree() {
  loading.value = true
  try {
    const res = await getMerchantCategoryTree()
    let data = res.data || []
    // 应用搜索过滤
    if (search.value) {
      data = filterTree(data, search.value)
    }
    if (statusFilter.value !== null && statusFilter.value !== '') {
      data = filterTreeByStatus(data, statusFilter.value)
    }
    treeData.value = data
  } catch (e) {
    ElMessage.error('加载类目树失败')
  } finally {
    loading.value = false
  }
}

function filterTree(nodes, keyword) {
  const result = []
  for (const node of nodes) {
    const children = node.children ? filterTree(node.children, keyword) : []
    if (node.name && node.name.includes(keyword)) {
      result.push({ ...node, children })
    } else if (children.length > 0) {
      result.push({ ...node, children })
    }
  }
  return result
}

function filterTreeByStatus(nodes, status) {
  const result = []
  for (const node of nodes) {
    const children = node.children ? filterTreeByStatus(node.children, status) : []
    if (node.status === status) {
      result.push({ ...node, children })
    } else if (children.length > 0) {
      result.push({ ...node, children })
    }
  }
  return result
}

async function loadList() {
  loading.value = true
  try {
    const params = { page: page.value, page_size: pageSize.value, keyword: search.value }
    if (statusFilter.value !== null && statusFilter.value !== '') params.status = statusFilter.value
    const res = await getMerchantCategoryList(params)
    list.value = res.data?.list || []
    total.value = res.data?.total || 0
  } catch (e) {
    ElMessage.error('加载类目列表失败')
  } finally {
    loading.value = false
  }
}

function onSearch() {
  if (treeMode.value) {
    loadTree()
  } else {
    page.value = 1
    loadList()
  }
}

function openCreate(parentId) {
  isEdit.value = false
  resetForm()
  form.parent_id = parentId || 0
  formVisible.value = true
}

function openEdit(row) {
  isEdit.value = true
  Object.assign(form, {
    id: row.id,
    parent_id: row.parent_id,
    name: row.name,
    icon: row.icon,
    sort: row.sort,
    status: row.status
  })
  formVisible.value = true
}

function resetForm() {
  Object.assign(form, {
    id: 0, parent_id: 0, name: '', icon: '', sort: 0, status: 1
  })
}

async function onSubmit() {
  if (!formRef.value) return
  await formRef.value.validate(async (valid) => {
    if (!valid) return
    submitting.value = true
    try {
      const payload = {
        parent_id: form.parent_id,
        name: form.name,
        icon: form.icon,
        sort: form.sort,
        status: form.status
      }
      if (isEdit.value) {
        await updateMerchantCategory(form.id, payload)
        ElMessage.success('更新成功')
      } else {
        await createMerchantCategory(payload)
        ElMessage.success('创建成功')
      }
      formVisible.value = false
      if (treeMode.value) loadTree()
      else loadList()
    } catch (e) {
      ElMessage.error(e.message || '操作失败')
    } finally {
      submitting.value = false
    }
  })
}

async function onToggleStatus(row) {
  const next = row.status === 1 ? 0 : 1
  try {
    await ElMessageBox.confirm(`确定要${next === 1 ? '启用' : '禁用'}类目「${row.name}」吗？`, '提示', { type: 'warning' })
    await updateMerchantCategoryStatus(row.id, { status: next })
    ElMessage.success('状态已更新')
    if (treeMode.value) loadTree()
    else loadList()
  } catch (e) {
    if (e !== 'cancel') ElMessage.error('操作失败')
  }
}

async function onDelete(row) {
  try {
    await ElMessageBox.confirm(`确定要删除类目「${row.name}」吗？删除后子类目也将一并删除。`, '警告', { type: 'warning' })
    await deleteMerchantCategory(row.id)
    ElMessage.success('删除成功')
    if (treeMode.value) loadTree()
    else loadList()
  } catch (e) {
    if (e !== 'cancel') ElMessage.error('删除失败')
  }
}

onMounted(() => loadTree())
</script>

<style scoped>
.pagination-wrap { margin-top: 16px; display: flex; justify-content: flex-end; }
.tip { color: #909399; font-size: 12px; margin-top: 4px; }
</style>
