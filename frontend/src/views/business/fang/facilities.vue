<template>
  <div class="app-container">
    <div class="page-card">
      <el-form :inline="true" :model="filters" class="filter-form" @submit.prevent>
        <el-form-item label="关键词">
          <el-input v-model="filters.keyword" placeholder="设施名" clearable style="width: 200px" :prefix-icon="Search" @keyup.enter="onSearch" />
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
        <el-button :icon="Refresh" @click="loadList">刷新</el-button>
        <el-button type="primary" :icon="Plus" @click="openCreate">新建设施</el-button>
      </div>

      <el-table v-loading="loading" :data="list" border stripe>
        <el-table-column prop="id" label="ID" width="70" fixed="left" />
        <el-table-column prop="name" label="设施名称" min-width="160" />
        <el-table-column label="图标" width="80">
          <template #default="{ row }">
            <el-image v-if="row.icon" :src="row.icon" fit="cover" style="width: 30px; height: 30px" />
            <span v-else class="text-muted">-</span>
          </template>
        </el-table-column>
        <el-table-column prop="category" label="分类" width="120" />
        <el-table-column prop="sort" label="排序" width="80" />
        <el-table-column prop="house_count" label="使用数" width="90" />
        <el-table-column label="状态" width="90">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'info'" size="small">{{ row.status === 1 ? '启用' : '禁用' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="创建时间" width="160">
          <template #default="{ row }">{{ formatTime(row.created_at) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="openEdit(row)">编辑</el-button>
            <el-button v-if="row.status === 1" type="warning" link size="small" @click="handleStatus(row, 0)">禁用</el-button>
            <el-button v-else type="success" link size="small" @click="handleStatus(row, 1)">启用</el-button>
            <el-button type="danger" link size="small" @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>

      <div class="pagination-wrap">
        <el-pagination v-model:current-page="page" v-model:page-size="pageSize" :total="total" :page-sizes="[10, 20, 50]" layout="total, sizes, prev, pager, next, jumper" background @current-change="loadList" @size-change="loadList" />
      </div>
    </div>

    <el-dialog v-model="formVisible" :title="formMode === 'create' ? '新建设施' : '编辑设施'" width="500px" destroy-on-close>
      <el-form ref="formRef" :model="form" :rules="rules" label-width="100px">
        <el-form-item label="设施名称" prop="name">
          <el-input v-model="form.name" placeholder="请输入设施名称" />
        </el-form-item>
        <el-form-item label="分类">
          <el-input v-model="form.category" placeholder="如：基础/家电/家具" />
        </el-form-item>
        <el-form-item label="图标URL">
          <el-input v-model="form.icon" placeholder="图标URL（可选）" />
        </el-form-item>
        <el-form-item label="排序">
          <el-input-number v-model="form.sort" :min="0" :controls="false" style="width: 200px" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="formVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="handleSubmit">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Search, RefreshLeft, Refresh, Plus } from '@element-plus/icons-vue'
import { listFacilities, createFacility, updateFacility, deleteFacility, updateFacilityStatus, getFacility } from '@/api/house'
import { formatTime } from '@/utils/format'

const loading = ref(false)
const list = ref([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const filters = reactive({ keyword: '', status: '' })

const formVisible = ref(false)
const formMode = ref('create')
const formRef = ref(null)
const submitting = ref(false)
const form = reactive({ id: null, name: '', category: '', icon: '', sort: 0 })
const rules = { name: [{ required: true, message: '请输入设施名称', trigger: 'blur' }] }

const buildParams = () => {
  const p = { page: page.value, page_size: pageSize.value }
  if (filters.keyword) p.keyword = filters.keyword
  if (filters.status !== '' && filters.status !== null) p.status = filters.status
  return p
}

const loadList = async () => {
  loading.value = true
  try {
    const res = await listFacilities(buildParams())
    const d = res.data || {}
    list.value = d.list || []
    total.value = d.total || 0
  } catch (e) { list.value = [] } finally { loading.value = false }
}

const onSearch = () => { page.value = 1; loadList() }
const onReset = () => { Object.assign(filters, { keyword: '', status: '' }); onSearch() }

const resetForm = () => Object.assign(form, { id: null, name: '', category: '', icon: '', sort: 0 })

const openCreate = () => { formMode.value = 'create'; resetForm(); formVisible.value = true }

const openEdit = async (row) => {
  formMode.value = 'edit'
  resetForm()
  try {
    const res = await getFacility(row.id)
    Object.assign(form, res.data || row)
  } catch (e) { Object.assign(form, row) }
  formVisible.value = true
}

const handleSubmit = async () => {
  try {
    await formRef.value.validate()
    submitting.value = true
    if (formMode.value === 'create') {
      await createFacility({ ...form })
      ElMessage.success('创建成功')
    } else {
      await updateFacility(form.id, { ...form })
      ElMessage.success('更新成功')
    }
    formVisible.value = false
    loadList()
  } catch (e) {
    if (e?.message) ElMessage.error(e.message)
  } finally { submitting.value = false }
}

const handleStatus = async (row, status) => {
  try {
    await ElMessageBox.confirm(`确定${status === 1 ? '启用' : '禁用'}该设施吗？`, '提示', { type: 'warning' })
    await updateFacilityStatus(row.id, { status })
    ElMessage.success('操作成功')
    loadList()
  } catch (e) { /* cancel */ }
}

const handleDelete = async (row) => {
  try {
    await ElMessageBox.confirm('确定删除该设施吗？', '危险操作', { type: 'error' })
    await deleteFacility(row.id)
    ElMessage.success('已删除')
    loadList()
  } catch (e) { /* cancel */ }
}

onMounted(() => loadList())
</script>

<style scoped>
.text-muted { color: #909399; }
</style>
