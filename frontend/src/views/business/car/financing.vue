<template>
  <div class="app-container">
    <div class="page-card">
      <el-form :inline="true" :model="filters" class="filter-form" @submit.prevent>
        <el-form-item label="关键词">
          <el-input v-model="filters.keyword" placeholder="方案名/金融机构" clearable style="width: 200px" :prefix-icon="Search" @keyup.enter="onSearch" />
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="filters.status" placeholder="全部" clearable style="width: 140px" @change="onSearch">
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
        <el-button type="primary" :icon="Plus" @click="openCreate">新建方案</el-button>
      </div>

      <el-table v-loading="loading" :data="list" border stripe>
        <el-table-column prop="id" label="ID" width="70" fixed="left" />
        <el-table-column prop="name" label="方案名称" min-width="160" />
        <el-table-column prop="institution" label="金融机构" width="140" />
        <el-table-column label="年利率" width="100">
          <template #default="{ row }">{{ row.interest_rate != null ? (row.interest_rate * 100).toFixed(2) + '%' : '-' }}</template>
        </el-table-column>
        <el-table-column label="首付比例" width="100">
          <template #default="{ row }">{{ row.down_payment_ratio != null ? (row.down_payment_ratio * 100).toFixed(0) + '%' : '-' }}</template>
        </el-table-column>
        <el-table-column label="最长年限" width="100">
          <template #default="{ row }">{{ row.max_years ? row.max_years + '年' : '-' }}</template>
        </el-table-column>
        <el-table-column label="状态" width="80">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'info'" size="small">{{ row.status === 1 ? '启用' : '禁用' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="openEdit(row)">编辑</el-button>
            <el-button :type="row.status === 1 ? 'warning' : 'success'" link size="small" @click="handleStatus(row)">{{ row.status === 1 ? '禁用' : '启用' }}</el-button>
            <el-popconfirm title="确认删除？" @confirm="handleDelete(row)">
              <template #reference><el-button type="danger" link size="small">删除</el-button></template>
            </el-popconfirm>
          </template>
        </el-table-column>
      </el-table>

      <div class="pagination-wrap">
        <el-pagination v-model:current-page="page" v-model:page-size="pageSize" :total="total" :page-sizes="[10, 20, 50]" layout="total, sizes, prev, pager, next, jumper" background @current-change="loadList" @size-change="loadList" />
      </div>
    </div>

    <el-dialog v-model="formVisible" :title="formMode === 'create' ? '新建分期方案' : '编辑分期方案'" width="600px" destroy-on-close>
      <el-form ref="formRef" :model="form" :rules="rules" label-width="100px">
        <el-form-item label="方案名称" prop="name">
          <el-input v-model="form.name" placeholder="方案名称" />
        </el-form-item>
        <el-form-item label="金融机构" prop="institution">
          <el-input v-model="form.institution" placeholder="如：XX银行" />
        </el-form-item>
        <el-form-item label="年利率" prop="interest_rate">
          <el-input-number v-model="form.interest_rate" :min="0" :max="1" :step="0.001" :precision="4" style="width: 100%" />
          <span class="tip-text">输入小数，如0.05表示5%</span>
        </el-form-item>
        <el-form-item label="首付比例" prop="down_payment_ratio">
          <el-input-number v-model="form.down_payment_ratio" :min="0" :max="1" :step="0.05" :precision="2" style="width: 100%" />
          <span class="tip-text">输入小数，如0.3表示30%</span>
        </el-form-item>
        <el-form-item label="最长年限" prop="max_years">
          <el-input-number v-model="form.max_years" :min="1" :max="10" style="width: 100%" />
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="form.description" type="textarea" :rows="3" />
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
import { ElMessage } from 'element-plus'
import { Search, RefreshLeft, Refresh, Plus } from '@element-plus/icons-vue'
import { adminListFinancings, createFinancing, updateFinancing, deleteFinancing, adminUpdateFinancingStatus } from '@/api/car'
import { formatTime } from '@/utils/format'

const loading = ref(false)
const list = ref([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const filters = reactive({ keyword: '', status: '' })

const formVisible = ref(false)
const formMode = ref('create')
const submitting = ref(false)
const formRef = ref(null)
const form = reactive({ id: null, name: '', institution: '', interest_rate: 0.05, down_payment_ratio: 0.3, max_years: 5, description: '' })
const rules = {
  name: [{ required: true, message: '请输入方案名称', trigger: 'blur' }],
  institution: [{ required: true, message: '请输入金融机构', trigger: 'blur' }]
}

const buildParams = () => {
  const p = { page: page.value, page_size: pageSize.value }
  if (filters.keyword) p.keyword = filters.keyword
  if (filters.status !== '' && filters.status !== null) p.status = filters.status
  return p
}

const loadList = async () => {
  loading.value = true
  try {
    const res = await adminListFinancings(buildParams())
    const d = res.data || {}
    list.value = d.list || []
    total.value = d.total || 0
  } catch (e) { list.value = [] } finally { loading.value = false }
}

const onSearch = () => { page.value = 1; loadList() }
const onReset = () => { Object.assign(filters, { keyword: '', status: '' }); onSearch() }

const openCreate = () => {
  Object.assign(form, { id: null, name: '', institution: '', interest_rate: 0.05, down_payment_ratio: 0.3, max_years: 5, description: '' })
  formMode.value = 'create'
  formVisible.value = true
}

const openEdit = (row) => {
  Object.assign(form, { id: row.id, name: row.name, institution: row.institution, interest_rate: row.interest_rate, down_payment_ratio: row.down_payment_ratio, max_years: row.max_years, description: row.description || '' })
  formMode.value = 'edit'
  formVisible.value = true
}

const handleSubmit = async () => {
  try {
    await formRef.value.validate()
    submitting.value = true
    if (formMode.value === 'create') {
      await createFinancing({ ...form })
      ElMessage.success('创建成功')
    } else {
      await updateFinancing(form.id, { ...form })
      ElMessage.success('更新成功')
    }
    formVisible.value = false
    loadList()
  } catch (e) {
    if (e?.message) ElMessage.error(e.message)
  } finally { submitting.value = false }
}

const handleStatus = async (row) => {
  try {
    await adminUpdateFinancingStatus(row.id, { status: row.status === 1 ? 0 : 1 })
    ElMessage.success('操作成功')
    loadList()
  } catch (e) {
    if (e?.message) ElMessage.error(e.message)
  }
}

const handleDelete = async (row) => {
  try {
    await deleteFinancing(row.id)
    ElMessage.success('删除成功')
    loadList()
  } catch (e) {
    if (e?.message) ElMessage.error(e.message)
  }
}

onMounted(() => loadList())
</script>

<style scoped>
.page-card { background: #fff; padding: 16px; border-radius: 4px; box-shadow: 0 1px 4px rgba(0,0,0,0.04); }
.filter-form { margin-bottom: 12px; padding: 12px 16px; background: #fafafa; border-radius: 4px; }
.filter-form :deep(.el-form-item) { margin-bottom: 8px; margin-right: 12px; }
.toolbar { display: flex; justify-content: flex-start; gap: 8px; margin-bottom: 12px; }
.pagination-wrap { margin-top: 16px; display: flex; justify-content: flex-end; }
.tip-text { font-size: 12px; color: #909399; }
</style>
