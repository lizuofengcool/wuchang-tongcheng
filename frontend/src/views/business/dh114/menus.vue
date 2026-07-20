<template>
  <div class="app-container">
    <!-- 统计卡片 -->
    <el-row :gutter="16" class="stat-row">
      <el-col :xs="12" :sm="8" :md="4">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value">{{ stats.total }}</div><div class="stat-label">总菜单数</div></div></el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="4">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value text-warning">{{ stats.signature }}</div><div class="stat-label">招牌菜数</div></div></el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="4">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value text-success">{{ stats.onSale }}</div><div class="stat-label">在售</div></div></el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="4">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value text-danger">{{ stats.offShelf }}</div><div class="stat-label">下架</div></div></el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="4">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value text-primary">¥{{ avgPrice.toFixed(2) }}</div><div class="stat-label">平均价</div></div></el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="4">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value">{{ stats.soldTotal }}</div><div class="stat-label">总销量</div></div></el-card>
      </el-col>
    </el-row>

    <div class="page-card">
      <el-form :inline="true" :model="filters" class="filter-form" @submit.prevent>
        <el-form-item label="商户ID">
          <el-input v-model="filters.dh114_id" placeholder="商户ID" clearable style="width: 140px" @keyup.enter="onSearch" />
        </el-form-item>
        <el-form-item label="关键词">
          <el-input v-model="filters.keyword" placeholder="菜名" clearable style="width: 200px" @keyup.enter="onSearch" />
        </el-form-item>
        <el-form-item label="招牌">
          <el-select v-model="filters.is_signature" placeholder="全部" clearable style="width: 120px" @change="onSearch">
            <el-option label="招牌菜" :value="1" />
            <el-option label="普通菜" :value="0" />
          </el-select>
        </el-form-item>
        <el-form-item label="价格区间">
          <el-input-number v-model="filters.min_price" :min="0" :controls="false" placeholder="最低" style="width: 100px" />
          <span style="margin: 0 4px">-</span>
          <el-input-number v-model="filters.max_price" :min="0" :controls="false" placeholder="最高" style="width: 100px" />
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
        <el-button type="primary" :icon="Plus" @click="openCreate">新建菜单</el-button>
      </div>

      <el-table v-loading="loading" :data="list" border stripe @selection-change="(rows) => selection = rows">
        <el-table-column type="selection" width="44" fixed="left" />
        <el-table-column prop="id" label="ID" width="70" />
        <el-table-column prop="dh114_id" label="商户ID" width="90" />
        <el-table-column label="图片" width="80">
          <template #default="{ row }">
            <el-image v-if="row.image" :src="row.image" fit="cover" class="thumb" preview-teleported :preview-src-list="[row.image]" />
            <div v-else class="thumb thumb-empty">无图</div>
          </template>
        </el-table-column>
        <el-table-column label="菜名" min-width="160">
          <template #default="{ row }">
            <span>{{ row.name }}</span>
            <el-tag v-if="row.is_signature" type="warning" size="small" effect="dark" style="margin-left: 6px">招牌</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="价格" width="100">
          <template #default="{ row }">
            <span class="price">¥{{ Number(row.price || 0).toFixed(2) }}</span>
          </template>
        </el-table-column>
        <el-table-column label="原价" width="100">
          <template #default="{ row }">
            <span v-if="row.original_price > 0" class="original-price">¥{{ Number(row.original_price).toFixed(2) }}</span>
            <span v-else class="text-muted">-</span>
          </template>
        </el-table-column>
        <el-table-column label="销量" width="80" prop="sold_count" sortable />
        <el-table-column label="状态" width="90">
          <template #default="{ row }">
            <el-switch :model-value="row.status === 1" @change="(val) => onToggle(row, val)" />
          </template>
        </el-table-column>
        <el-table-column label="描述" min-width="180" show-overflow-tooltip>
          <template #default="{ row }">{{ row.description || '-' }}</template>
        </el-table-column>
        <el-table-column label="创建时间" width="160">
          <template #default="{ row }">{{ formatTime(row.created_at) }}</template>
        </el-table-column>
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
    <el-dialog v-model="formVisible" :title="formTitle" width="640px" destroy-on-close>
      <el-form ref="formRef" :model="form" :rules="rules" label-width="100px">
        <el-form-item label="商户ID" prop="dh114_id">
          <el-input-number v-model="form.dh114_id" :min="1" :controls="false" style="width: 100%" />
        </el-form-item>
        <el-form-item label="菜名" prop="name">
          <el-input v-model="form.name" maxlength="128" />
        </el-form-item>
        <el-form-item label="图片">
          <el-input v-model="form.image" placeholder="图片 URL" />
        </el-form-item>
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="价格" prop="price">
              <el-input-number v-model="form.price" :min="0" :precision="2" :controls="false" style="width: 100%" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="原价">
              <el-input-number v-model="form.original_price" :min="0" :precision="2" :controls="false" style="width: 100%" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="16">
          <el-col :span="8">
            <el-form-item label="招牌">
              <el-switch v-model="form.is_signature" :active-value="1" :inactive-value="0" />
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item label="状态">
              <el-select v-model="form.status" style="width: 100%">
                <el-option label="在售" :value="1" />
                <el-option label="下架" :value="0" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item label="排序">
              <el-input-number v-model="form.sort" :min="0" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-form-item label="描述">
          <el-input v-model="form.description" type="textarea" :rows="3" maxlength="500" show-word-limit />
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
import request from '@/utils/request'
import { formatTime } from '@/utils/format'

const loading = ref(false)
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)
const list = ref([])
const selection = ref([])

const stats = reactive({ total: 0, signature: 0, onSale: 0, offShelf: 0, soldTotal: 0 })
const avgPrice = ref(0)

const filters = reactive({ dh114_id: '', keyword: '', is_signature: null, min_price: undefined, max_price: undefined })

const onSearch = () => { page.value = 1; loadList() }
const onReset = () => {
  filters.dh114_id = ''
  filters.keyword = ''
  filters.is_signature = null
  filters.min_price = undefined
  filters.max_price = undefined
  page.value = 1
  loadList()
}

const loadList = async () => {
  loading.value = true
  try {
    const params = {
      page: page.value,
      page_size: pageSize.value,
      dh114_id: filters.dh114_id || undefined,
      keyword: filters.keyword || undefined,
      is_signature: filters.is_signature === null || filters.is_signature === '' ? undefined : filters.is_signature,
      min_price: filters.min_price || undefined,
      max_price: filters.max_price || undefined
    }
    const res = await request.get('/dh114/menus', { params })
    const data = res.data || {}
    list.value = data.list || []
    total.value = data.total || 0
    calcStats()
  } catch (e) {
    list.value = []
    total.value = 0
  } finally {
    loading.value = false
  }
}

const calcStats = () => {
  const total = list.value.length
  const signature = list.value.filter((r) => r.is_signature === 1).length
  const onSale = list.value.filter((r) => r.status === 1).length
  const offShelf = list.value.filter((r) => r.status === 0).length
  const soldTotal = list.value.reduce((s, r) => s + (r.sold_count || 0), 0)
  const sum = list.value.reduce((s, r) => s + Number(r.price || 0), 0)
  Object.assign(stats, { total, signature, onSale, offShelf, soldTotal })
  avgPrice.value = total ? sum / total : 0
}

// ===== 表单 =====
const formVisible = ref(false)
const formLoading = ref(false)
const formRef = ref(null)
const isEdit = ref(false)
const formTitle = computed(() => isEdit.value ? '编辑菜单' : '新建菜单')
const form = reactive({
  id: null, dh114_id: undefined, name: '', image: '',
  price: 0, original_price: 0, is_signature: 0,
  status: 1, sort: 0, description: ''
})
const rules = {
  dh114_id: [{ required: true, message: '请输入商户ID', trigger: 'blur' }],
  name: [{ required: true, message: '请输入菜名', trigger: 'blur' }],
  price: [{ required: true, message: '请输入价格', trigger: 'blur' }]
}

const resetForm = () => {
  Object.assign(form, {
    id: null, dh114_id: undefined, name: '', image: '',
    price: 0, original_price: 0, is_signature: 0,
    status: 1, sort: 0, description: ''
  })
}

const openCreate = () => {
  isEdit.value = false
  resetForm()
  formVisible.value = true
}

const openEdit = (row) => {
  isEdit.value = true
  Object.assign(form, {
    id: row.id, dh114_id: row.dh114_id, name: row.name || '',
    image: row.image || '', price: row.price || 0,
    original_price: row.original_price || 0, is_signature: row.is_signature || 0,
    status: row.status || 1, sort: row.sort || 0, description: row.description || ''
  })
  formVisible.value = true
}

const onSubmit = async () => {
  try {
    await formRef.value.validate()
    formLoading.value = true
    const payload = { ...form }
    if (isEdit.value) {
      await request.put(`/dh114/menus/${form.id}`, payload)
    } else {
      await request.post('/dh114/menus', payload)
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
    await ElMessageBox.confirm(`确认删除菜单 "${row.name}"？`, '提示', { type: 'warning' })
    await request.delete(`/dh114/menus/${row.id}`)
    ElMessage.success('删除成功')
    await loadList()
  } catch (e) { /* cancel */ }
}

const onToggle = async (row, val) => {
  try {
    await request.put(`/dh114/menus/${row.id}`, { status: val ? 1 : 0 })
    row.status = val ? 1 : 0
    ElMessage.success('状态已更新')
  } catch (e) { /* ignore */ }
}

const onBatchDelete = async () => {
  try {
    await ElMessageBox.confirm(`确认批量删除 ${selection.value.length} 个菜单？`, '危险操作', {
      type: 'error', confirmButtonText: '确认删除', cancelButtonText: '取消'
    })
    for (const row of selection.value) {
      await request.delete(`/dh114/menus/${row.id}`)
    }
    ElMessage.success('批量删除完成')
    await loadList()
  } catch (e) { /* cancel */ }
}

onMounted(() => loadList())
</script>

<style scoped>
.stat-row { margin-bottom: 16px; }
.stat-item { text-align: center; padding: 16px; }
.stat-value { font-size: 22px; font-weight: 600; color: #409eff; }
.text-success { color: #67c23a; }
.text-warning { color: #e6a23c; }
.text-danger { color: #f56c6c; }
.text-primary { color: #409eff; }
.filter-form { margin-bottom: 12px; padding: 12px 16px; background: #fafafa; border-radius: 4px; }
.filter-form :deep(.el-form-item) { margin-bottom: 8px; margin-right: 12px; }
.toolbar { display: flex; justify-content: space-between; margin-bottom: 12px; }
.toolbar-left { display: flex; gap: 8px; }
.text-muted { color: #909399; }
.thumb { width: 50px; height: 50px; border-radius: 4px; border: 1px solid #ebeef5; }
.thumb-empty { display: flex; align-items: center; justify-content: center; background: #fafafa; color: #909399; font-size: 12px; }
.price { color: #f56c6c; font-weight: 600; }
.original-price { color: #c0c4cc; text-decoration: line-through; font-size: 12px; }
.pagination-wrap { margin-top: 16px; display: flex; justify-content: flex-end; }
</style>
