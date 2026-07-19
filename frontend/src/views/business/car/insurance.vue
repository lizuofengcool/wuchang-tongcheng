<template>
  <div class="app-container">
    <div class="page-card">
      <el-form :inline="true" :model="filters" class="filter-form" @submit.prevent>
        <el-form-item label="关键词">
          <el-input v-model="filters.keyword" placeholder="险种名/保险公司" clearable style="width: 200px" :prefix-icon="Search" @keyup.enter="onSearch" />
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
        <el-button type="primary" :icon="Plus" @click="openCreate">新建险种</el-button>
      </div>

      <el-table v-loading="loading" :data="list" border stripe>
        <el-table-column prop="id" label="ID" width="70" fixed="left" />
        <el-table-column prop="name" label="险种名称" min-width="160" />
        <el-table-column prop="company" label="保险公司" width="140" />
        <el-table-column prop="type" label="险种类型" width="100">
          <template #default="{ row }">{{ row.type || '-' }}</template>
        </el-table-column>
        <el-table-column label="基础保费" width="120">
          <template #default="{ row }">¥{{ Number(row.base_premium || 0).toFixed(2) }}</template>
        </el-table-column>
        <el-table-column label="保额" width="120">
          <template #default="{ row }">¥{{ Number(row.coverage || 0).toFixed(2) }}</template>
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

    <el-dialog v-model="formVisible" :title="formMode === 'create' ? '新建险种' : '编辑险种'" width="600px" destroy-on-close>
      <el-form ref="formRef" :model="form" :rules="rules" label-width="100px">
        <el-form-item label="险种名称" prop="name">
          <el-input v-model="form.name" placeholder="如：交强险" />
        </el-form-item>
        <el-form-item label="保险公司" prop="company">
          <el-input v-model="form.company" placeholder="如：人保财险" />
        </el-form-item>
        <el-form-item label="险种类型" prop="type">
          <el-select v-model="form.type" placeholder="选择类型" style="width: 100%">
            <el-option label="交强险" value="compulsory" />
            <el-option label="商业险" value="commercial" />
            <el-option label="第三者责任险" value="third_party" />
            <el-option label="车损险" value="damage" />
            <el-option label="盗抢险" value="theft" />
            <el-option label="其他" value="other" />
          </el-select>
        </el-form-item>
        <el-form-item label="基础保费" prop="base_premium">
          <el-input-number v-model="form.base_premium" :min="0" :precision="2" style="width: 100%" />
        </el-form-item>
        <el-form-item label="保额" prop="coverage">
          <el-input-number v-model="form.coverage" :min="0" :precision="2" style="width: 100%" />
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
import { adminListInsurances, createInsurance, updateInsurance, deleteInsurance, adminUpdateInsuranceStatus } from '@/api/car'

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
const form = reactive({ id: null, name: '', company: '', type: '', base_premium: 0, coverage: 0, description: '' })
const rules = {
  name: [{ required: true, message: '请输入险种名称', trigger: 'blur' }],
  company: [{ required: true, message: '请输入保险公司', trigger: 'blur' }],
  type: [{ required: true, message: '请选择险种类型', trigger: 'change' }]
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
    const res = await adminListInsurances(buildParams())
    const d = res.data || {}
    list.value = d.list || []
    total.value = d.total || 0
  } catch (e) { list.value = [] } finally { loading.value = false }
}

const onSearch = () => { page.value = 1; loadList() }
const onReset = () => { Object.assign(filters, { keyword: '', status: '' }); onSearch() }

const openCreate = () => {
  Object.assign(form, { id: null, name: '', company: '', type: '', base_premium: 0, coverage: 0, description: '' })
  formMode.value = 'create'
  formVisible.value = true
}

const openEdit = (row) => {
  Object.assign(form, { id: row.id, name: row.name, company: row.company, type: row.type, base_premium: row.base_premium, coverage: row.coverage, description: row.description || '' })
  formMode.value = 'edit'
  formVisible.value = true
}

const handleSubmit = async () => {
  try {
    await formRef.value.validate()
    submitting.value = true
    if (formMode.value === 'create') {
      await createInsurance({ ...form })
      ElMessage.success('创建成功')
    } else {
      await updateInsurance(form.id, { ...form })
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
    await adminUpdateInsuranceStatus(row.id, { status: row.status === 1 ? 0 : 1 })
    ElMessage.success('操作成功')
    loadList()
  } catch (e) {
    if (e?.message) ElMessage.error(e.message)
  }
}

const handleDelete = async (row) => {
  try {
    await deleteInsurance(row.id)
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
</style>
