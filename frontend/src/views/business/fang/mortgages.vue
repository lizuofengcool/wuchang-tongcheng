<template>
  <div class="app-container">
    <div class="page-card">
      <el-form :inline="true" :model="filters" class="filter-form" @submit.prevent>
        <el-form-item label="关键词">
          <el-input v-model="filters.keyword" placeholder="方案名/银行" clearable style="width: 200px" :prefix-icon="Search" @keyup.enter="onSearch" />
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
        <el-button type="primary" :icon="Plus" @click="openCreate">新建方案</el-button>
      </div>

      <el-table v-loading="loading" :data="list" border stripe>
        <el-table-column prop="id" label="ID" width="70" fixed="left" />
        <el-table-column prop="name" label="方案名称" min-width="180" show-overflow-tooltip />
        <el-table-column prop="bank" label="银行" width="140" />
        <el-table-column label="利率" width="100">
          <template #default="{ row }">{{ row.interest_rate ? `${(row.interest_rate * 100).toFixed(2)}%` : '-' }}</template>
        </el-table-column>
        <el-table-column label="首付比例" width="100">
          <template #default="{ row }">{{ row.down_payment_ratio ? `${(row.down_payment_ratio * 100).toFixed(0)}%` : '-' }}</template>
        </el-table-column>
        <el-table-column label="最长年限" width="100">
          <template #default="{ row }">{{ row.max_years ? `${row.max_years}年` : '-' }}</template>
        </el-table-column>
        <el-table-column prop="apply_count" label="申请数" width="90" />
        <el-table-column label="状态" width="90">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'info'" size="small">{{ row.status === 1 ? '启用' : '禁用' }}</el-tag>
          </template>
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

    <el-dialog v-model="formVisible" :title="formMode === 'create' ? '新建房贷方案' : '编辑房贷方案'" width="600px" destroy-on-close>
      <el-form ref="formRef" :model="form" :rules="rules" label-width="120px">
        <el-form-item label="方案名称" prop="name">
          <el-input v-model="form.name" placeholder="如：首套商业贷款" />
        </el-form-item>
        <el-form-item label="银行" prop="bank">
          <el-input v-model="form.bank" placeholder="如：中国银行" />
        </el-form-item>
        <el-form-item label="利率">
          <el-input-number v-model="form.interest_rate" :min="0" :max="1" :step="0.001" :precision="4" :controls="false" style="width: 200px" />
          <span style="margin-left: 8px; color: #909399">（0-1，如0.045表示4.5%）</span>
        </el-form-item>
        <el-form-item label="首付比例">
          <el-input-number v-model="form.down_payment_ratio" :min="0" :max="1" :step="0.05" :precision="2" :controls="false" style="width: 200px" />
        </el-form-item>
        <el-form-item label="最长年限">
          <el-input-number v-model="form.max_years" :min="0" :max="50" :controls="false" style="width: 200px" />
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
import { ElMessage, ElMessageBox } from 'element-plus'
import { Search, RefreshLeft, Refresh, Plus } from '@element-plus/icons-vue'
import { listMortgages, createMortgage, updateMortgage, deleteMortgage, updateMortgageStatus } from '@/api/house'

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
const form = reactive({ id: null, name: '', bank: '', interest_rate: 0, down_payment_ratio: 0, max_years: 30, description: '' })
const rules = {
  name: [{ required: true, message: '请输入方案名称', trigger: 'blur' }],
  bank: [{ required: true, message: '请输入银行', trigger: 'blur' }]
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
    const res = await listMortgages(buildParams())
    const d = res.data || {}
    list.value = d.list || []
    total.value = d.total || 0
  } catch (e) { list.value = [] } finally { loading.value = false }
}

const onSearch = () => { page.value = 1; loadList() }
const onReset = () => { Object.assign(filters, { keyword: '', status: '' }); onSearch() }

const resetForm = () => Object.assign(form, { id: null, name: '', bank: '', interest_rate: 0, down_payment_ratio: 0, max_years: 30, description: '' })

const openCreate = () => { formMode.value = 'create'; resetForm(); formVisible.value = true }

const openEdit = (row) => {
  formMode.value = 'edit'
  resetForm()
  Object.assign(form, row)
  formVisible.value = true
}

const handleSubmit = async () => {
  try {
    await formRef.value.validate()
    submitting.value = true
    if (formMode.value === 'create') {
      await createMortgage({ ...form })
      ElMessage.success('创建成功')
    } else {
      await updateMortgage(form.id, { ...form })
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
    await ElMessageBox.confirm(`确定${status === 1 ? '启用' : '禁用'}该方案吗？`, '提示', { type: 'warning' })
    await updateMortgageStatus(row.id, { status })
    ElMessage.success('操作成功')
    loadList()
  } catch (e) { /* cancel */ }
}

const handleDelete = async (row) => {
  try {
    await ElMessageBox.confirm('确定删除该方案吗？', '危险操作', { type: 'error' })
    await deleteMortgage(row.id)
    ElMessage.success('已删除')
    loadList()
  } catch (e) { /* cancel */ }
}

onMounted(() => loadList())
</script>
