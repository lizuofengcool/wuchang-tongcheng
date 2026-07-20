<template>
  <div class="app-container">
    <!-- 顶部统计卡片 -->
    <el-row :gutter="16" class="stat-row">
      <el-col :xs="12" :sm="8" :md="6">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value">{{ stats.total }}</div><div class="stat-label">等级总数</div></div></el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="6">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value text-success">{{ stats.enabled }}</div><div class="stat-label">启用中</div></div></el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="6">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value text-primary">{{ stats.totalSubscribers }}</div><div class="stat-label">总订阅数</div></div></el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="6">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value text-warning">¥{{ stats.totalRevenue }}</div><div class="stat-label">总收入</div></div></el-card>
      </el-col>
    </el-row>

    <div class="page-card">
      <el-form :inline="true" :model="filters" class="filter-form" @submit.prevent>
        <el-form-item label="关键词">
          <el-input v-model="filters.keyword" placeholder="等级名/编码" clearable style="width: 200px" :prefix-icon="Search" @keyup.enter="onSearch" />
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
        <el-button type="primary" :icon="Plus" @click="openCreate">新建等级</el-button>
      </div>

      <el-table v-loading="loading" :data="list" border stripe @sort-change="onSortChange">
        <el-table-column prop="id" label="ID" width="70" sortable="custom" />
        <el-table-column prop="level_code" label="等级编码" width="120" />
        <el-table-column prop="level_name" label="等级名" min-width="160" />
        <el-table-column label="价格" width="120">
          <template #default="{ row }">
            <span class="price">¥{{ Number(row.price || 0).toFixed(2) }}</span>
            <span class="price-unit">/{{ periodText(row.period_unit) }}</span>
          </template>
        </el-table-column>
        <el-table-column label="原价" width="100">
          <template #default="{ row }">
            <span v-if="row.original_price > 0" class="original-price">¥{{ Number(row.original_price).toFixed(2) }}</span>
            <span v-else class="text-muted">-</span>
          </template>
        </el-table-column>
        <el-table-column label="权益" min-width="200">
          <template #default="{ row }">
            <el-tag v-for="(p, i) in parsePerks(row.perks)" :key="i" size="small" class="tag-item">{{ p }}</el-tag>
            <span v-if="!parsePerks(row.perks).length" class="text-muted">-</span>
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
        <el-form-item label="等级编码" prop="level_code">
          <el-input v-model="form.level_code" maxlength="32" placeholder="如：VIP1, VIP2" />
        </el-form-item>
        <el-form-item label="等级名" prop="level_name">
          <el-input v-model="form.level_name" maxlength="64" />
        </el-form-item>
        <el-form-item label="价格" prop="price">
          <el-input-number v-model="form.price" :min="0" :precision="2" :step="1" style="width: 180px" />
          <el-select v-model="form.period_unit" style="width: 120px; margin-left: 8px">
            <el-option label="每月" value="month" />
            <el-option label="每季" value="quarter" />
            <el-option label="每年" value="year" />
            <el-option label="永久" value="forever" />
          </el-select>
        </el-form-item>
        <el-form-item label="原价">
          <el-input-number v-model="form.original_price" :min="0" :precision="2" :step="1" style="width: 180px" />
        </el-form-item>
        <el-form-item label="权益列表">
          <el-input v-model="form.perksStr" type="textarea" :rows="4" placeholder='JSON 数组，如 ["无限喜欢","查看访客","专属徽章"]' />
        </el-form-item>
        <el-form-item label="图标">
          <el-input v-model="form.icon" placeholder="图标URL" />
        </el-form-item>
        <el-form-item label="颜色">
          <el-color-picker v-model="form.color" />
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
import { Refresh, RefreshLeft, Search, Plus } from '@element-plus/icons-vue'
import request from '@/utils/request'

const loading = ref(false)
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)
const list = ref([])
const sortField = ref('sort')
const sortOrder = ref('ascending')

const filters = reactive({ keyword: '', status: null })

const stats = reactive({ total: 0, enabled: 0, totalSubscribers: 0, totalRevenue: 0 })

const periodText = (u) => ({ month: '月', quarter: '季', year: '年', forever: '永久' }[u] || '-')

const parsePerks = (p) => {
  if (!p) return []
  if (Array.isArray(p)) return p
  try { return JSON.parse(p) || [] } catch (e) { return String(p).split(/[,，]/).filter(Boolean) }
}

const onSearch = () => { page.value = 1; loadList() }
const onReset = () => { filters.keyword = ''; filters.status = null; page.value = 1; loadList() }
const onSortChange = ({ prop, order }) => {
  sortField.value = prop || 'sort'
  sortOrder.value = order || 'ascending'
  loadList()
}

const loadList = async () => {
  loading.value = true
  try {
    const params = {
      page: page.value,
      page_size: pageSize.value,
      keyword: filters.keyword.trim() || undefined,
      status: filters.status === null || filters.status === '' ? undefined : filters.status,
      sort: sortField.value,
      order: sortOrder.value
    }
    const res = await request.get('/love/admin/member-levels', { params })
    const data = res.data || {}
    list.value = data.list || data || []
    total.value = data.total || list.value.length
    calcStats()
  } catch (e) {
    list.value = []
    total.value = 0
  } finally {
    loading.value = false
  }
}

const calcStats = () => {
  stats.total = list.value.length
  stats.enabled = list.value.filter((r) => r.status === 1).length
  stats.totalSubscribers = list.value.reduce((s, r) => s + (r.subscriber_count || 0), 0)
  stats.totalRevenue = list.value.reduce((s, r) => s + Number(r.revenue || 0), 0).toFixed(2)
}

// 表单
const formVisible = ref(false)
const formLoading = ref(false)
const formRef = ref(null)
const isEdit = ref(false)
const formTitle = computed(() => isEdit.value ? '编辑等级' : '新建等级')
const form = reactive({
  id: null, level_code: '', level_name: '', price: 0, original_price: 0,
  period_unit: 'month', perksStr: '', icon: '', color: '#409eff',
  sort: 0, status: 1, description: ''
})
const rules = {
  level_code: [{ required: true, message: '请输入等级编码', trigger: 'blur' }],
  level_name: [{ required: true, message: '请输入等级名', trigger: 'blur' }],
  price: [{ required: true, message: '请输入价格', trigger: 'blur' }]
}

const openCreate = () => {
  isEdit.value = false
  Object.assign(form, {
    id: null, level_code: '', level_name: '', price: 0, original_price: 0,
    period_unit: 'month', perksStr: '', icon: '', color: '#409eff',
    sort: 0, status: 1, description: ''
  })
  formVisible.value = true
}

const openEdit = (row) => {
  isEdit.value = true
  Object.assign(form, {
    ...row,
    perksStr: row.perks ? JSON.stringify(parsePerks(row.perks), null, 2) : ''
  })
  formVisible.value = true
}

const onSubmit = async () => {
  try {
    await formRef.value.validate()
    let perks = []
    if (form.perksStr) {
      try { perks = JSON.parse(form.perksStr) }
      catch (e) {
        ElMessage.error('权益列表 JSON 格式错误')
        return
      }
    }
    const payload = { ...form, perks }
    delete payload.perksStr
    formLoading.value = true
    if (isEdit.value) {
      await request.put(`/love/admin/member-levels/${form.id}`, payload)
    } else {
      await request.post('/love/admin/member-levels', payload)
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
    await ElMessageBox.confirm(`确认删除等级 "${row.level_name}"？`, '提示', { type: 'warning' })
    await request.delete(`/love/admin/member-levels/${row.id}`)
    ElMessage.success('删除成功')
    await loadList()
  } catch (e) { /* cancel */ }
}

const onToggle = async (row, val) => {
  try {
    await request.put(`/love/admin/member-levels/${row.id}`, { status: val ? 1 : 0 })
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
.text-success { color: #67c23a; }
.text-warning { color: #e6a23c; }
.text-primary { color: #409eff; }
.filter-form { margin-bottom: 12px; padding: 12px 16px; background: #fafafa; border-radius: 4px; }
.filter-form :deep(.el-form-item) { margin-bottom: 8px; margin-right: 12px; }
.toolbar { display: flex; justify-content: space-between; margin-bottom: 12px; }
.text-muted { color: #909399; }
.price { color: #f56c6c; font-weight: 600; }
.original-price { color: #c0c4cc; text-decoration: line-through; font-size: 12px; }
.price-unit { color: #909399; font-size: 12px; margin-left: 2px; }
.tag-item { margin-right: 4px; margin-bottom: 4px; }
.pagination-wrap { margin-top: 16px; display: flex; justify-content: flex-end; }
</style>
