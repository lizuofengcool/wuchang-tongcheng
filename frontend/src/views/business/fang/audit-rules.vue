<template>
  <div class="app-container">
    <div class="page-card">
      <el-form :inline="true" :model="filters" class="filter-form" @submit.prevent>
        <el-form-item label="关键词">
          <el-input v-model="filters.keyword" placeholder="规则名" clearable style="width: 200px" :prefix-icon="Search" @keyup.enter="onSearch" />
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
        <el-button type="primary" :icon="Plus" @click="openCreate">新建规则</el-button>
      </div>

      <el-table v-loading="loading" :data="list" border stripe>
        <el-table-column prop="id" label="ID" width="70" fixed="left" />
        <el-table-column prop="name" label="规则名称" min-width="160" />
        <el-table-column prop="target_type" label="目标类型" width="120" />
        <el-table-column prop="field" label="字段" width="120" />
        <el-table-column prop="operator" label="操作符" width="100" />
        <el-table-column prop="value" label="阈值" min-width="120" show-overflow-tooltip />
        <el-table-column prop="action" label="动作" width="100" />
        <el-table-column prop="priority" label="优先级" width="80" />
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

    <el-dialog v-model="formVisible" :title="formMode === 'create' ? '新建规则' : '编辑规则'" width="600px" destroy-on-close>
      <el-form ref="formRef" :model="form" :rules="rules" label-width="120px">
        <el-form-item label="规则名称" prop="name">
          <el-input v-model="form.name" placeholder="如：价格异常检测" />
        </el-form-item>
        <el-form-item label="目标类型" prop="target_type">
          <el-select v-model="form.target_type" placeholder="选择目标" style="width: 100%">
            <el-option label="房源" value="house" />
            <el-option label="发布单" value="listing" />
            <el-option label="评价" value="review" />
          </el-select>
        </el-form-item>
        <el-form-item label="字段" prop="field">
          <el-input v-model="form.field" placeholder="如：price/title" />
        </el-form-item>
        <el-form-item label="操作符" prop="operator">
          <el-select v-model="form.operator" placeholder="选择操作符" style="width: 100%">
            <el-option label="大于" value="gt" />
            <el-option label="小于" value="lt" />
            <el-option label="等于" value="eq" />
            <el-option label="包含" value="contains" />
            <el-option label="正则" value="regex" />
          </el-select>
        </el-form-item>
        <el-form-item label="阈值" prop="value">
          <el-input v-model="form.value" placeholder="对比值" />
        </el-form-item>
        <el-form-item label="动作">
          <el-select v-model="form.action" placeholder="选择动作" style="width: 100%">
            <el-option label="拒绝" value="reject" />
            <el-option label="人工审核" value="manual" />
            <el-option label="警告" value="warn" />
          </el-select>
        </el-form-item>
        <el-form-item label="优先级">
          <el-input-number v-model="form.priority" :min="0" :controls="false" style="width: 200px" />
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
import { listAuditRules, createAuditRule, updateAuditRule, deleteAuditRule, updateAuditRuleStatus } from '@/api/house'

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
const form = reactive({ id: null, name: '', target_type: 'house', field: '', operator: 'gt', value: '', action: 'reject', priority: 0 })
const rules = {
  name: [{ required: true, message: '请输入规则名称', trigger: 'blur' }],
  target_type: [{ required: true, message: '请选择目标类型', trigger: 'change' }],
  field: [{ required: true, message: '请输入字段', trigger: 'blur' }],
  operator: [{ required: true, message: '请选择操作符', trigger: 'change' }],
  value: [{ required: true, message: '请输入阈值', trigger: 'blur' }]
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
    const res = await listAuditRules(buildParams())
    const d = res.data || {}
    list.value = d.list || []
    total.value = d.total || 0
  } catch (e) { list.value = [] } finally { loading.value = false }
}

const onSearch = () => { page.value = 1; loadList() }
const onReset = () => { Object.assign(filters, { keyword: '', status: '' }); onSearch() }

const resetForm = () => Object.assign(form, { id: null, name: '', target_type: 'house', field: '', operator: 'gt', value: '', action: 'reject', priority: 0 })

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
      await createAuditRule({ ...form })
      ElMessage.success('创建成功')
    } else {
      await updateAuditRule(form.id, { ...form })
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
    await ElMessageBox.confirm(`确定${status === 1 ? '启用' : '禁用'}该规则吗？`, '提示', { type: 'warning' })
    await updateAuditRuleStatus(row.id, { status })
    ElMessage.success('操作成功')
    loadList()
  } catch (e) { /* cancel */ }
}

const handleDelete = async (row) => {
  try {
    await ElMessageBox.confirm('确定删除该规则吗？', '危险操作', { type: 'error' })
    await deleteAuditRule(row.id)
    ElMessage.success('已删除')
    loadList()
  } catch (e) { /* cancel */ }
}

onMounted(() => loadList())
</script>
