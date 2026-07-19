<template>
  <div class="app-container">
    <div class="page-card">
      <el-form :inline="true" :model="filters" class="filter-form" @submit.prevent>
        <el-form-item label="名称">
          <el-input v-model="filters.keyword" placeholder="标签名" clearable style="width: 180px" @keyup.enter="onSearch" />
        </el-form-item>
        <el-form-item label="类型">
          <el-select v-model="filters.type" placeholder="全部" clearable style="width: 140px" @change="onSearch">
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
        <el-button :icon="Refresh" @click="loadList">刷新</el-button>
        <el-button type="danger" :icon="Delete" :disabled="!selection.length" @click="onBatchDelete">批量删除</el-button>
        <div class="toolbar-right">
          <el-button type="primary" :icon="Plus" @click="openCreate">新建标签</el-button>
        </div>
      </div>

      <el-table v-loading="loading" :data="list" border stripe @selection-change="(rows) => selection = rows">
        <el-table-column type="selection" width="44" />
        <el-table-column prop="id" label="ID" width="70" />
        <el-table-column label="名称" min-width="140">
          <template #default="{ row }">
            <el-tag :color="row.background" :style="{ color: row.color }" effect="dark">{{ row.name }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="类型" width="100">
          <template #default="{ row }">
            <el-tag size="small" :type="typeTagType(row.type)">{{ typeMap[row.type] || row.type }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="颜色" width="100">
          <template #default="{ row }">
            <div class="color-cell">
              <div class="color-block" :style="{ background: row.color }"></div>
              <span>{{ row.color || '-' }}</span>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="use_count" label="使用次数" width="100" />
        <el-table-column prop="sort" label="排序" width="80" />
        <el-table-column label="热门" width="70">
          <template #default="{ row }">
            <el-tag v-if="row.is_hot" type="danger" size="small">热</el-tag>
            <span v-else class="text-muted">-</span>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="90">
          <template #default="{ row }">
            <el-switch :model-value="row.status === 1" @change="(val) => onToggle(row, val)" />
          </template>
        </el-table-column>
        <el-table-column label="创建时间" width="160">
          <template #default="{ row }">{{ formatTime(row.created_at) }}</template>
        </el-table-column>
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

    <!-- 新建/编辑弹窗 -->
    <el-dialog v-model="formVisible" :title="formTitle" width="500px">
      <el-form ref="formRef" :model="form" :rules="rules" label-width="100px">
        <el-form-item label="名称" prop="name">
          <el-input v-model="form.name" maxlength="64" show-word-limit placeholder="标签名" />
        </el-form-item>
        <el-form-item label="类型" prop="type">
          <el-select v-model="form.type" placeholder="请选择" style="width: 100%">
            <el-option v-for="(label, val) in typeMap" :key="val" :label="label" :value="val" />
          </el-select>
        </el-form-item>
        <el-form-item label="颜色">
          <el-color-picker v-model="form.color" />
          <el-input v-model="form.color" placeholder="#409eff" style="width: 140px; margin-left: 8px" />
        </el-form-item>
        <el-form-item label="背景色">
          <el-color-picker v-model="form.background" />
          <el-input v-model="form.background" placeholder="#409eff" style="width: 140px; margin-left: 8px" />
        </el-form-item>
        <el-form-item label="图标">
          <el-input v-model="form.icon" placeholder="图标名" />
        </el-form-item>
        <el-form-item label="排序">
          <el-input-number v-model="form.sort" :min="0" />
        </el-form-item>
        <el-form-item label="热门">
          <el-switch v-model="form.is_hot" />
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
  listErshouTags, createErshouTag, updateErshouTag, deleteErshouTag
} from '@/api/ershou'
import { formatTime } from '@/utils/format'

const loading = ref(false)
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)
const list = ref([])
const selection = ref([])

const filters = reactive({ keyword: '', type: '', status: null })

const typeMap = { smart: '智能标签', operation: '运营标签', custom: '自定义' }
const typeTagType = (t) => ({ smart: 'primary', operation: 'success', custom: 'info' }[t] || 'info')

const onSearch = () => { page.value = 1; loadList() }
const onReset = () => { filters.keyword = ''; filters.type = ''; filters.status = null; page.value = 1; loadList() }

const loadList = async () => {
  loading.value = true
  try {
    const res = await listErshouTags({
      page: page.value,
      page_size: pageSize.value,
      keyword: filters.keyword || undefined,
      type: filters.type || undefined,
      status: filters.status === null || filters.status === '' ? undefined : filters.status
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

// ===== 表单 =====
const formVisible = ref(false)
const formLoading = ref(false)
const formRef = ref(null)
const isEdit = ref(false)
const formTitle = computed(() => isEdit.value ? '编辑标签' : '新建标签')
const form = reactive({
  id: null, name: '', type: 'custom', color: '#409eff', background: '',
  icon: '', sort: 0, is_hot: false, status: 1
})
const rules = {
  name: [{ required: true, message: '请输入名称', trigger: 'blur' }],
  type: [{ required: true, message: '请选择类型', trigger: 'change' }]
}

const openCreate = () => {
  isEdit.value = false
  Object.assign(form, { id: null, name: '', type: 'custom', color: '#409eff', background: '', icon: '', sort: 0, is_hot: false, status: 1 })
  formVisible.value = true
}

const openEdit = (row) => {
  isEdit.value = true
  Object.assign(form, {
    id: row.id, name: row.name, type: row.type || 'custom',
    color: row.color || '#409eff', background: row.background || '',
    icon: row.icon || '', sort: row.sort || 0, is_hot: !!row.is_hot, status: row.status
  })
  formVisible.value = true
}

const onSubmit = async () => {
  try {
    await formRef.value.validate()
    formLoading.value = true
    if (isEdit.value) {
      await updateErshouTag(form.id, { ...form })
    } else {
      await createErshouTag({ ...form })
    }
    ElMessage.success(isEdit.value ? '更新成功' : '创建成功')
    formVisible.value = false
    await loadList()
  } catch (e) {
    // 校验或接口失败
  } finally {
    formLoading.value = false
  }
}

const onDelete = async (row) => {
  try {
    await ElMessageBox.confirm(`确认删除标签 "${row.name}"？`, '提示', { type: 'warning' })
    await deleteErshouTag(row.id)
    ElMessage.success('删除成功')
    await loadList()
  } catch (e) { /* cancel */ }
}

const onBatchDelete = async () => {
  try {
    await ElMessageBox.confirm(`确认批量删除 ${selection.value.length} 个标签？`, '提示', { type: 'warning' })
    await Promise.all(selection.value.map((r) => deleteErshouTag(r.id)))
    ElMessage.success('批量删除成功')
    await loadList()
  } catch (e) { /* cancel */ }
}

const onToggle = async (row, val) => {
  try {
    await updateErshouTag(row.id, { status: val ? 1 : 0 })
    row.status = val ? 1 : 0
    ElMessage.success('状态已更新')
  } catch (e) { /* ignore */ }
}

onMounted(() => loadList())
</script>

<style scoped>
.filter-form { margin-bottom: 12px; padding: 12px 16px; background: #fafafa; border-radius: 4px; }
.filter-form :deep(.el-form-item) { margin-bottom: 8px; margin-right: 12px; }
.toolbar { display: flex; justify-content: space-between; margin-bottom: 12px; }
.toolbar-right { display: flex; gap: 8px; }
.color-cell { display: flex; align-items: center; gap: 6px; }
.color-block { width: 16px; height: 16px; border-radius: 2px; border: 1px solid #ebeef5; }
.text-muted { color: #909399; }
.pagination-wrap { margin-top: 16px; display: flex; justify-content: flex-end; }
</style>
