<template>
  <div class="app-container">
    <div class="page-card">
      <el-form :inline="true" :model="filters" class="filter-form" @submit.prevent>
        <el-form-item label="规则名">
          <el-input v-model="filters.keyword" placeholder="规则名" clearable style="width: 180px" @keyup.enter="onSearch" />
        </el-form-item>
        <el-form-item label="类型">
          <el-select v-model="filters.rule_type" placeholder="全部" clearable style="width: 160px" @change="onSearch">
            <el-option v-for="(label, val) in typeMap" :key="val" :label="label" :value="val" />
          </el-select>
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
        <div class="toolbar-left">
          <el-button :icon="Refresh" @click="loadList">刷新</el-button>
          <el-button type="primary" :icon="Plus" @click="openCreate">新建规则</el-button>
        </div>
        <div class="toolbar-right">
          <el-button type="success" :icon="Check" :disabled="!selection.length" @click="onBatchStatus(1)">批量启用</el-button>
          <el-button type="warning" :icon="Close" :disabled="!selection.length" @click="onBatchStatus(0)">批量禁用</el-button>
        </div>
      </div>

      <el-table v-loading="loading" :data="list" border stripe @selection-change="onSelectionChange">
        <el-table-column type="selection" width="44" fixed="left" />
        <el-table-column prop="id" label="ID" width="70" />
        <el-table-column prop="rule_name" label="规则名" min-width="160" />
        <el-table-column label="类型" width="120">
          <template #default="{ row }">
            <el-tag size="small" :type="typeTagType(row.rule_type)">{{ typeMap[row.rule_type] || row.rule_type }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="rule_key" label="规则键" width="120" />
        <el-table-column prop="pattern" label="匹配模式" min-width="180" show-overflow-tooltip />
        <el-table-column label="处置" width="100">
          <template #default="{ row }">
            <el-tag size="small" :type="actionTagType(row.action)">{{ actionMap[row.action] || row.action }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="严重程度" width="100">
          <template #default="{ row }">
            <el-rate v-model="row.severity" disabled />
          </template>
        </el-table-column>
        <el-table-column prop="sort" label="排序" width="80" />
        <el-table-column label="状态" width="90">
          <template #default="{ row }">
            <el-switch :model-value="row.status === 1" @change="(val) => onToggle(row, val)" />
          </template>
        </el-table-column>
        <el-table-column prop="description" label="描述" min-width="180" show-overflow-tooltip />
        <el-table-column label="操作" width="140" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="openEdit(row)">编辑</el-button>
            <el-button type="danger" link size="small" @click="onDelete(row)">删除</el-button>
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
          @current-change="loadList"
          @size-change="loadList"
        />
      </div>
    </div>

    <!-- 表单弹窗 -->
    <el-dialog v-model="formVisible" :title="formTitle" width="640px">
      <el-form ref="formRef" :model="form" :rules="rules" label-width="100px">
        <el-form-item label="规则名" prop="rule_name">
          <el-input v-model="form.rule_name" maxlength="128" />
        </el-form-item>
        <el-form-item label="类型" prop="rule_type">
          <el-select v-model="form.rule_type" style="width: 100%">
            <el-option v-for="(label, val) in typeMap" :key="val" :label="label" :value="val" />
          </el-select>
        </el-form-item>
        <el-form-item label="规则键">
          <el-input v-model="form.rule_key" maxlength="64" placeholder="英文键名" />
        </el-form-item>
        <el-form-item label="匹配模式">
          <el-input v-model="form.pattern" type="textarea" :rows="3" placeholder="正则表达式或敏感词列表（逗号分隔）" />
        </el-form-item>
        <el-form-item label="阈值">
          <el-input v-model="form.thresholdStr" type="textarea" :rows="3" placeholder='JSON 格式，如 {"min":0,"max":100}' />
        </el-form-item>
        <el-form-item label="处置" prop="action">
          <el-select v-model="form.action" style="width: 100%">
            <el-option v-for="(label, val) in actionMap" :key="val" :label="label" :value="val" />
          </el-select>
        </el-form-item>
        <el-form-item label="处罚类型">
          <el-select v-model="form.penalty_type" clearable style="width: 100%">
            <el-option label="警告" value="warning" />
            <el-option label="限制" value="limit" />
            <el-option label="封禁1天" value="ban1d" />
            <el-option label="封禁7天" value="ban7d" />
            <el-option label="永久封禁" value="banForever" />
          </el-select>
        </el-form-item>
        <el-form-item label="严重程度">
          <el-rate v-model="form.severity" />
        </el-form-item>
        <el-form-item label="排序">
          <el-input-number v-model="form.sort" :min="0" />
        </el-form-item>
        <el-form-item label="状态">
          <el-switch v-model="form.status" :active-value="1" :inactive-value="0" />
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="form.description" type="textarea" :rows="2" maxlength="500" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="formVisible = false">取消</el-button>
        <el-button type="primary" :loading="formLoading" @click="onSubmit">确认</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Refresh, RefreshLeft, Search, Plus, Check, Close } from '@element-plus/icons-vue'
import request from '@/utils/request'

const loading = ref(false)
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)
const list = ref([])
const selection = ref([])

const filters = reactive({ keyword: '', rule_type: '', status: null })

const typeMap = {
  sensitive_word: '敏感词', photo_check: '照片审核', info_check: '资料检查',
  age_check: '年龄检查', content: '内容审核', behavior: '行为审核'
}
const typeTagType = (t) => ({
  sensitive_word: 'danger', photo_check: 'warning', info_check: 'info',
  age_check: 'info', content: 'primary', behavior: 'warning'
}[t] || 'info')

const actionMap = { reject: '拒绝', approval: '通过', warning: '警告', manual: '人工审核' }
const actionTagType = (a) => ({
  reject: 'danger', approval: 'success', warning: 'warning', manual: 'info'
}[a] || 'info')

const onSearch = () => { page.value = 1; loadList() }
const onReset = () => { filters.keyword = ''; filters.rule_type = ''; filters.status = null; page.value = 1; loadList() }

const onSelectionChange = (rows) => { selection.value = rows }

const loadList = async () => {
  loading.value = true
  try {
    const params = {
      page: page.value,
      page_size: pageSize.value,
      keyword: filters.keyword || undefined,
      rule_type: filters.rule_type || undefined,
      status: filters.status === null || filters.status === '' ? undefined : filters.status
    }
    const res = await request.get('/love/admin/audit-rules', { params })
    const data = res.data || {}
    list.value = data.list || data || []
    total.value = data.total || list.value.length
  } catch (e) {
    list.value = []
    total.value = 0
  } finally {
    loading.value = false
  }
}

const formVisible = ref(false)
const formLoading = ref(false)
const formRef = ref(null)
const isEdit = ref(false)
const formTitle = computed(() => isEdit.value ? '编辑规则' : '新建规则')
const form = reactive({
  id: null, rule_name: '', rule_type: 'sensitive_word', rule_key: '',
  pattern: '', thresholdStr: '', action: 'manual', penalty_type: '',
  severity: 3, status: 1, description: '', sort: 0
})
const rules = {
  rule_name: [{ required: true, message: '请输入规则名', trigger: 'blur' }],
  rule_type: [{ required: true, message: '请选择类型', trigger: 'change' }],
  action: [{ required: true, message: '请选择处置', trigger: 'change' }]
}

const openCreate = () => {
  isEdit.value = false
  Object.assign(form, {
    id: null, rule_name: '', rule_type: 'sensitive_word', rule_key: '',
    pattern: '', thresholdStr: '', action: 'manual', penalty_type: '',
    severity: 3, status: 1, description: '', sort: 0
  })
  formVisible.value = true
}

const openEdit = (row) => {
  isEdit.value = true
  Object.assign(form, {
    ...row,
    thresholdStr: row.threshold ? JSON.stringify(row.threshold, null, 2) : ''
  })
  formVisible.value = true
}

const onSubmit = async () => {
  try {
    await formRef.value.validate()
    let threshold = null
    if (form.thresholdStr) {
      try {
        threshold = JSON.parse(form.thresholdStr)
      } catch (e) {
        ElMessage.error('阈值 JSON 格式错误')
        return
      }
    }
    const payload = { ...form, threshold }
    delete payload.thresholdStr
    formLoading.value = true
    if (isEdit.value) {
      await request.put(`/love/admin/audit-rules/${form.id}`, payload)
    } else {
      await request.post('/love/admin/audit-rules', payload)
    }
    ElMessage.success(isEdit.value ? '更新成功' : '创建成功')
    formVisible.value = false
    await loadList()
  } catch (e) {
    // ignore
  } finally {
    formLoading.value = false
  }
}

const onDelete = async (row) => {
  try {
    await ElMessageBox.confirm(`确认删除规则 "${row.rule_name}"？`, '提示', { type: 'warning' })
    await request.delete(`/love/admin/audit-rules/${row.id}`)
    ElMessage.success('删除成功')
    await loadList()
  } catch (e) { /* cancel */ }
}

const onToggle = async (row, val) => {
  try {
    await request.put(`/love/admin/audit-rules/${row.id}/status`, { status: val ? 1 : 0 })
    row.status = val ? 1 : 0
    ElMessage.success('状态已更新')
  } catch (e) { /* ignore */ }
}

const onBatchStatus = async (status) => {
  try {
    const label = status === 1 ? '启用' : '禁用'
    await ElMessageBox.confirm(`确认批量${label} ${selection.value.length} 个规则？`, '批量操作', { type: 'warning' })
    await request.put('/love/admin/audit-rules/batch-status', {
      ids: selection.value.map((r) => r.id),
      status
    })
    ElMessage.success('批量操作完成')
    await loadList()
  } catch (e) { /* cancel */ }
}

onMounted(() => loadList())
</script>

<style scoped>
.filter-form { margin-bottom: 12px; padding: 12px 16px; background: #fafafa; border-radius: 4px; }
.filter-form :deep(.el-form-item) { margin-bottom: 8px; margin-right: 12px; }
.toolbar { display: flex; justify-content: space-between; align-items: center; margin-bottom: 12px; flex-wrap: wrap; gap: 8px; }
.toolbar-left { display: flex; gap: 8px; }
.toolbar-right { display: flex; gap: 8px; }
.pagination-wrap { margin-top: 16px; display: flex; justify-content: flex-end; }
</style>
