<template>
  <div class="app-container">
    <!-- 顶部统计卡片 -->
    <el-row :gutter="16" class="stat-row">
      <el-col :xs="12" :sm="8" :md="6">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value">{{ stats.total }}</div><div class="stat-label">技能总数</div></div></el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="6">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value text-success">{{ stats.enabled }}</div><div class="stat-label">启用</div></div></el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="6">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value text-primary">{{ stats.categories }}</div><div class="stat-label">分类数</div></div></el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="6">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value text-warning">{{ stats.hot }}</div><div class="stat-label">热门数</div></div></el-card>
      </el-col>
    </el-row>

    <div class="page-card">
      <el-form :inline="true" :model="filters" class="filter-form" @submit.prevent>
        <el-form-item label="关键词">
          <el-input v-model="filters.keyword" placeholder="技能名" clearable style="width: 200px" @keyup.enter="onSearch" />
        </el-form-item>
        <el-form-item label="分类">
          <el-input v-model="filters.category" placeholder="分类" clearable style="width: 160px" @keyup.enter="onSearch" />
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
          <el-button type="danger" :icon="Delete" :disabled="!selection.length" @click="onBatchDelete">批量删除</el-button>
        </div>
        <div class="toolbar-right">
          <el-button type="primary" :icon="Plus" @click="openCreate">新建技能</el-button>
        </div>
      </div>

      <el-table v-loading="loading" :data="list" border stripe @selection-change="onSelectionChange">
        <el-table-column type="selection" width="44" fixed="left" />
        <el-table-column prop="id" label="ID" width="70" fixed="left" />
        <el-table-column prop="name" label="技能名" min-width="160" />
        <el-table-column prop="category" label="分类" width="120" />
        <el-table-column prop="parent_id" label="父级ID" width="90" />
        <el-table-column prop="icon" label="图标" width="80" />
        <el-table-column prop="sort" label="排序" width="80" />
        <el-table-column prop="usage_count" label="使用数" width="90" />
        <el-table-column label="热门" width="80">
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
        <el-table-column prop="description" label="描述" min-width="180" show-overflow-tooltip />
        <el-table-column label="操作" width="160" fixed="right">
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
        <el-form-item label="技能名" prop="name">
          <el-input v-model="form.name" maxlength="64" />
        </el-form-item>
        <el-form-item label="分类">
          <el-input v-model="form.category" maxlength="64" placeholder="如：餐饮/物流/IT" />
        </el-form-item>
        <el-form-item label="父级ID">
          <el-input-number v-model="form.parent_id" :min="0" :controls="false" />
        </el-form-item>
        <el-form-item label="图标">
          <el-input v-model="form.icon" maxlength="64" placeholder="图标 URL 或 class" />
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
import { Refresh, RefreshLeft, Search, Plus, Delete } from '@element-plus/icons-vue'
import {
  adminListLinggongSkills, createLinggongSkill, updateLinggongSkill,
  deleteLinggongSkill, adminUpdateLinggongSkillStatus
} from '@/api/linggong'

const loading = ref(false)
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)
const list = ref([])
const selection = ref([])

const filters = reactive({ keyword: '', category: '', status: null })

const stats = computed(() => {
  const total = list.value.length
  const enabled = list.value.filter((r) => r.status === 1).length
  const categories = new Set(list.value.map((r) => r.category).filter(Boolean)).size
  const hot = list.value.filter((r) => r.is_hot).length
  return { total, enabled, categories, hot }
})

const onSearch = () => { page.value = 1; loadList() }
const onReset = () => {
  filters.keyword = ''
  filters.category = ''
  filters.status = null
  page.value = 1
  loadList()
}

const onSelectionChange = (rows) => {
  selection.value = rows
}

const loadList = async () => {
  loading.value = true
  try {
    const res = await adminListLinggongSkills({
      page: page.value,
      page_size: pageSize.value,
      keyword: filters.keyword || undefined,
      category: filters.category || undefined,
      status: filters.status === null || filters.status === '' ? undefined : filters.status
    })
    const data = res.data || {}
    list.value = data.list || []
    total.value = data.total || 0
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
const formTitle = computed(() => isEdit.value ? '编辑技能' : '新建技能')
const form = reactive({
  id: null, name: '', category: '', parent_id: 0, icon: '',
  sort: 0, is_hot: false, status: 1, description: ''
})
const rules = {
  name: [{ required: true, message: '请输入技能名', trigger: 'blur' }]
}

const openCreate = () => {
  isEdit.value = false
  Object.assign(form, {
    id: null, name: '', category: '', parent_id: 0, icon: '',
    sort: 0, is_hot: false, status: 1, description: ''
  })
  formVisible.value = true
}

const openEdit = (row) => {
  isEdit.value = true
  Object.assign(form, row)
  formVisible.value = true
}

const onSubmit = async () => {
  try {
    await formRef.value.validate()
    formLoading.value = true
    if (isEdit.value) {
      await updateLinggongSkill(form.id, form)
    } else {
      await createLinggongSkill(form)
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
    await ElMessageBox.confirm(`确认删除技能 "${row.name}"？`, '提示', { type: 'warning' })
    await deleteLinggongSkill(row.id)
    ElMessage.success('删除成功')
    await loadList()
  } catch (e) { /* cancel */ }
}

const onBatchDelete = async () => {
  try {
    await ElMessageBox.confirm(`确认批量删除 ${selection.value.length} 个技能？`, '批量删除', { type: 'warning' })
    await Promise.all(selection.value.map((r) => deleteLinggongSkill(r.id)))
    ElMessage.success('批量删除完成')
    await loadList()
  } catch (e) { /* cancel */ }
}

const onToggle = async (row, val) => {
  try {
    await adminUpdateLinggongSkillStatus(row.id, val ? 1 : 0)
    row.status = val ? 1 : 0
    ElMessage.success('状态已更新')
  } catch (e) { /* ignore */ }
}

onMounted(() => loadList())
</script>

<style scoped>
.stat-row { margin-bottom: 16px; }
.stat-item { text-align: center; padding: 16px; }
.stat-value { font-size: 22px; font-weight: 600; color: #303133; }
.text-primary { color: #409eff; }
.text-warning { color: #e6a23c; }
.text-success { color: #67c23a; }
.text-muted { color: #909399; }
.filter-form { margin-bottom: 12px; padding: 12px 16px; background: #fafafa; border-radius: 4px; }
.filter-form :deep(.el-form-item) { margin-bottom: 8px; margin-right: 12px; }
.toolbar {
  display: flex; justify-content: space-between; align-items: center;
  flex-wrap: wrap; gap: 8px; margin-bottom: 12px;
}
.toolbar-left { display: flex; gap: 8px; flex-wrap: wrap; }
.toolbar-right { display: flex; gap: 8px; }
.pagination-wrap { margin-top: 16px; display: flex; justify-content: flex-end; }
</style>
