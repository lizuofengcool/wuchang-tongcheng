<template>
  <div class="app-container">
    <div class="page-card">
      <el-form :inline="true" :model="filters" class="filter-form" @submit.prevent>
        <el-form-item label="分类ID">
          <el-input v-model="filters.category_id" placeholder="分类ID" clearable style="width: 140px" @keyup.enter="onSearch" />
        </el-form-item>
        <el-form-item label="属性名">
          <el-input v-model="filters.keyword" placeholder="属性名" clearable style="width: 180px" @keyup.enter="onSearch" />
        </el-form-item>
        <el-form-item label="类型">
          <el-select v-model="filters.attr_type" placeholder="全部" clearable style="width: 140px" @change="onSearch">
            <el-option v-for="(label, val) in typeMap" :key="val" :label="label" :value="val" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" :icon="Search" @click="onSearch">搜索</el-button>
          <el-button :icon="RefreshLeft" @click="onReset">重置</el-button>
        </el-form-item>
      </el-form>

      <div class="toolbar">
        <el-button :icon="Refresh" @click="loadList">刷新</el-button>
        <el-button type="primary" :icon="Plus" @click="openCreate">新建属性</el-button>
      </div>

      <el-table v-loading="loading" :data="list" border stripe>
        <el-table-column prop="id" label="ID" width="70" />
        <el-table-column prop="category_id" label="分类ID" width="90" />
        <el-table-column prop="attr_name" label="属性名" min-width="140" />
        <el-table-column prop="attr_key" label="属性键" width="120" />
        <el-table-column label="类型" width="110">
          <template #default="{ row }">
            <el-tag size="small" :type="typeTagType(row.attr_type)">{{ typeMap[row.attr_type] || row.attr_type }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="可选值" min-width="180">
          <template #default="{ row }">
            <span v-if="row.options && row.options.length">
              <el-tag v-for="(opt, i) in row.options" :key="i" size="small" class="opt-tag">{{ opt }}</el-tag>
            </span>
            <span v-else class="text-muted">-</span>
          </template>
        </el-table-column>
        <el-table-column prop="unit" label="单位" width="80" />
        <el-table-column label="必填" width="70">
          <template #default="{ row }">
            <el-tag v-if="row.is_required" type="danger" size="small">必</el-tag>
            <span v-else class="text-muted">-</span>
          </template>
        </el-table-column>
        <el-table-column label="可筛选" width="80">
          <template #default="{ row }">
            <el-tag v-if="row.is_filterable" type="success" size="small">筛</el-tag>
            <span v-else class="text-muted">-</span>
          </template>
        </el-table-column>
        <el-table-column label="可搜索" width="80">
          <template #default="{ row }">
            <el-tag v-if="row.is_searchable" type="primary" size="small">搜</el-tag>
            <span v-else class="text-muted">-</span>
          </template>
        </el-table-column>
        <el-table-column prop="sort" label="排序" width="80" />
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
    <el-dialog v-model="formVisible" :title="formTitle" width="600px">
      <el-form ref="formRef" :model="form" :rules="rules" label-width="100px">
        <el-form-item label="分类ID" prop="category_id">
          <el-input-number v-model="form.category_id" :min="1" :controls="false" style="width: 100%" />
        </el-form-item>
        <el-form-item label="属性名" prop="attr_name">
          <el-input v-model="form.attr_name" maxlength="64" />
        </el-form-item>
        <el-form-item label="属性键">
          <el-input v-model="form.attr_key" maxlength="64" placeholder="英文键名" />
        </el-form-item>
        <el-form-item label="类型" prop="attr_type">
          <el-select v-model="form.attr_type" style="width: 100%">
            <el-option v-for="(label, val) in typeMap" :key="val" :label="label" :value="val" />
          </el-select>
        </el-form-item>
        <el-form-item v-if="['select', 'multi_select'].includes(form.attr_type)" label="可选值">
          <div class="options-editor">
            <div v-for="(opt, i) in form.options" :key="i" class="option-row">
              <el-input v-model="form.options[i]" placeholder="选项值" />
              <el-button type="danger" :icon="Delete" circle size="small" @click="form.options.splice(i, 1)" />
            </div>
            <el-button :icon="Plus" size="small" @click="form.options.push('')">添加选项</el-button>
          </div>
        </el-form-item>
        <el-form-item label="单位">
          <el-input v-model="form.unit" maxlength="32" />
        </el-form-item>
        <el-form-item label="默认值">
          <el-input v-model="form.default_value" />
        </el-form-item>
        <el-form-item label="占位符">
          <el-input v-model="form.placeholder" />
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="form.description" type="textarea" :rows="2" />
        </el-form-item>
        <el-form-item label="排序">
          <el-input-number v-model="form.sort" :min="0" />
        </el-form-item>
        <el-form-item label="必填">
          <el-switch v-model="form.is_required" />
        </el-form-item>
        <el-form-item label="可筛选">
          <el-switch v-model="form.is_filterable" />
        </el-form-item>
        <el-form-item label="可搜索">
          <el-switch v-model="form.is_searchable" />
        </el-form-item>
        <el-form-item label="状态">
          <el-switch v-model="form.status" :active-value="1" :inactive-value="0" />
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
import { Refresh, RefreshLeft, Search, Plus, Delete } from '@element-plus/icons-vue'
import {
  listErshouCategoryAttrs, createErshouCategoryAttr, updateErshouCategoryAttr, deleteErshouCategoryAttr
} from '@/api/ershou'

const loading = ref(false)
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)
const list = ref([])

const filters = reactive({ category_id: '', keyword: '', attr_type: '' })

const typeMap = {
  string: '字符串', number: '数字', select: '单选', multi_select: '多选', date: '日期', boolean: '布尔'
}
const typeTagType = (t) => ({
  string: '', number: 'success', select: 'warning', multi_select: 'danger', date: 'info', boolean: 'primary'
}[t] || 'info')

const onSearch = () => { page.value = 1; loadList() }
const onReset = () => { filters.category_id = ''; filters.keyword = ''; filters.attr_type = ''; page.value = 1; loadList() }

const loadList = async () => {
  loading.value = true
  try {
    const res = await listErshouCategoryAttrs({
      page: page.value,
      page_size: pageSize.value,
      category_id: filters.category_id || undefined,
      keyword: filters.keyword || undefined,
      attr_type: filters.attr_type || undefined
    })
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
const formTitle = computed(() => isEdit.value ? '编辑属性' : '新建属性')
const form = reactive({
  id: null, category_id: undefined, attr_name: '', attr_key: '',
  attr_type: 'string', options: [], unit: '', is_required: false,
  is_filterable: false, is_searchable: false, default_value: '',
  placeholder: '', description: '', status: 1, sort: 0
})
const rules = {
  category_id: [{ required: true, message: '请输入分类ID', trigger: 'blur' }],
  attr_name: [{ required: true, message: '请输入属性名', trigger: 'blur' }],
  attr_type: [{ required: true, message: '请选择类型', trigger: 'change' }]
}

const openCreate = () => {
  isEdit.value = false
  Object.assign(form, {
    id: null, category_id: undefined, attr_name: '', attr_key: '',
    attr_type: 'string', options: [], unit: '', is_required: false,
    is_filterable: false, is_searchable: false, default_value: '',
    placeholder: '', description: '', status: 1, sort: 0
  })
  formVisible.value = true
}

const openEdit = (row) => {
  isEdit.value = true
  Object.assign(form, {
    ...row,
    options: row.options ? [...row.options] : []
  })
  formVisible.value = true
}

const onSubmit = async () => {
  try {
    await formRef.value.validate()
    formLoading.value = true
    const payload = { ...form, options: form.options.filter((o) => o) }
    if (isEdit.value) {
      await updateErshouCategoryAttr(form.id, payload)
    } else {
      await createErshouCategoryAttr(payload)
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
    await ElMessageBox.confirm(`确认删除属性 "${row.attr_name}"？`, '提示', { type: 'warning' })
    await deleteErshouCategoryAttr(row.id)
    ElMessage.success('删除成功')
    await loadList()
  } catch (e) { /* cancel */ }
}

onMounted(() => loadList())
</script>

<style scoped>
.filter-form { margin-bottom: 12px; padding: 12px 16px; background: #fafafa; border-radius: 4px; }
.filter-form :deep(.el-form-item) { margin-bottom: 8px; margin-right: 12px; }
.toolbar { display: flex; justify-content: space-between; margin-bottom: 12px; }
.text-muted { color: #909399; }
.opt-tag { margin-right: 4px; margin-bottom: 4px; }
.options-editor { width: 100%; }
.option-row { display: flex; gap: 8px; margin-bottom: 8px; }
.pagination-wrap { margin-top: 16px; display: flex; justify-content: flex-end; }
</style>
