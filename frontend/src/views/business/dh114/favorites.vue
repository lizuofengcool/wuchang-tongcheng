<template>
  <div class="app-container">
    <!-- 统计卡片 -->
    <el-row :gutter="16" class="stat-row">
      <el-col :xs="12" :sm="8" :md="6">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value">{{ stats.total }}</div><div class="stat-label">总收藏数</div></div></el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="6">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value text-success">{{ stats.today }}</div><div class="stat-label">今日新增</div></div></el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="6">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value text-warning">{{ stats.uniqueUsers }}</div><div class="stat-label">独立用户</div></div></el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="6">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value text-primary">{{ stats.uniqueBusiness }}</div><div class="stat-label">被收藏商户</div></div></el-card>
      </el-col>
    </el-row>

    <div class="page-card">
      <el-form :inline="true" :model="filters" class="filter-form" @submit.prevent>
        <el-form-item label="商户ID">
          <el-input v-model="filters.dh114_id" placeholder="商户ID" clearable style="width: 140px" @keyup.enter="onSearch" />
        </el-form-item>
        <el-form-item label="用户ID">
          <el-input v-model="filters.user_id" placeholder="用户ID" clearable style="width: 140px" @keyup.enter="onSearch" />
        </el-form-item>
        <el-form-item label="收藏类型">
          <el-select v-model="filters.target_type" placeholder="全部" clearable style="width: 140px" @change="onSearch">
            <el-option label="商户" value="business" />
            <el-option label="优惠券" value="coupon" />
            <el-option label="团购" value="groupbuy" />
          </el-select>
        </el-form-item>
        <el-form-item label="分组">
          <el-input v-model="filters.group_name" placeholder="分组名" clearable style="width: 140px" @keyup.enter="onSearch" />
        </el-form-item>
        <el-form-item label="时间范围">
          <el-date-picker
            v-model="filters.dateRange"
            type="daterange"
            range-separator="至"
            start-placeholder="开始"
            end-placeholder="结束"
            value-format="YYYY-MM-DD"
            style="width: 240px"
          />
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
        <el-button type="primary" :icon="Plus" @click="openCreate">新建收藏</el-button>
      </div>

      <el-table v-loading="loading" :data="list" border stripe @selection-change="(rows) => selection = rows">
        <el-table-column type="selection" width="44" fixed="left" />
        <el-table-column prop="id" label="ID" width="70" />
        <el-table-column prop="user_id" label="用户ID" width="90" />
        <el-table-column prop="dh114_id" label="商户ID" width="90" />
        <el-table-column label="商户名" min-width="160">
          <template #default="{ row }">{{ row.business_name || `商户#${row.dh114_id}` }}</template>
        </el-table-column>
        <el-table-column label="收藏对象" width="120">
          <template #default="{ row }">
            <el-tag :type="targetTypeTag(row.target_type)" size="small">{{ targetText(row.target_type) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="target_id" label="对象ID" width="90" />
        <el-table-column label="分组" width="120">
          <template #default="{ row }">{{ row.group_name || '-' }}</template>
        </el-table-column>
        <el-table-column label="备注" min-width="160" show-overflow-tooltip>
          <template #default="{ row }">{{ row.remark || '-' }}</template>
        </el-table-column>
        <el-table-column label="收藏时间" width="160">
          <template #default="{ row }">{{ formatTime(row.created_at) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="120" fixed="right">
          <template #default="{ row }">
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
    <el-dialog v-model="formVisible" title="新建收藏" width="500px" destroy-on-close>
      <el-form ref="formRef" :model="form" :rules="rules" label-width="100px">
        <el-form-item label="用户ID" prop="user_id">
          <el-input-number v-model="form.user_id" :min="1" :controls="false" style="width: 100%" />
        </el-form-item>
        <el-form-item label="商户ID" prop="dh114_id">
          <el-input-number v-model="form.dh114_id" :min="1" :controls="false" style="width: 100%" />
        </el-form-item>
        <el-form-item label="收藏类型">
          <el-select v-model="form.target_type" style="width: 100%">
            <el-option label="商户" value="business" />
            <el-option label="优惠券" value="coupon" />
            <el-option label="团购" value="groupbuy" />
          </el-select>
        </el-form-item>
        <el-form-item label="对象ID">
          <el-input-number v-model="form.target_id" :min="1" :controls="false" style="width: 100%" />
        </el-form-item>
        <el-form-item label="分组名">
          <el-input v-model="form.group_name" maxlength="64" />
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="form.remark" type="textarea" :rows="2" maxlength="200" />
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
import { ref, reactive, onMounted } from 'vue'
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

const stats = reactive({ total: 0, today: 0, uniqueUsers: 0, uniqueBusiness: 0 })

const filters = reactive({
  dh114_id: '', user_id: '', target_type: '', group_name: '', dateRange: null
})

const onSearch = () => { page.value = 1; loadList() }
const onReset = () => {
  filters.dh114_id = ''
  filters.user_id = ''
  filters.target_type = ''
  filters.group_name = ''
  filters.dateRange = null
  page.value = 1
  loadList()
}

const targetText = (t) => ({ business: '商户', coupon: '优惠券', groupbuy: '团购' }[t] || '-')
const targetTypeTag = (t) => ({ business: 'primary', coupon: 'warning', groupbuy: 'success' }[t] || 'info')

const loadList = async () => {
  loading.value = true
  try {
    const params = {
      page: page.value,
      page_size: pageSize.value,
      dh114_id: filters.dh114_id || undefined,
      user_id: filters.user_id || undefined,
      target_type: filters.target_type || undefined,
      group_name: filters.group_name || undefined
    }
    if (filters.dateRange && filters.dateRange.length === 2) {
      params.start_date = filters.dateRange[0]
      params.end_date = filters.dateRange[1]
    }
    const res = await request.get('/dh114/favorites', { params })
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
  const today = new Date().toISOString().slice(0, 10)
  const total = list.value.length
  const todayCount = list.value.filter((r) => r.created_at && r.created_at.slice(0, 10) === today).length
  const users = new Set(list.value.map((r) => r.user_id).filter(Boolean)).size
  const business = new Set(list.value.map((r) => r.dh114_id).filter(Boolean)).size
  Object.assign(stats, { total, today: todayCount, uniqueUsers: users, uniqueBusiness: business })
}

const onDelete = async (row) => {
  try {
    await ElMessageBox.confirm('确认删除该收藏？', '提示', { type: 'warning' })
    await request.delete(`/dh114/favorites/${row.id}`)
    ElMessage.success('删除成功')
    await loadList()
  } catch (e) { /* cancel */ }
}

const onBatchDelete = async () => {
  try {
    await ElMessageBox.confirm(`确认批量删除 ${selection.value.length} 个收藏？`, '危险操作', {
      type: 'error', confirmButtonText: '确认删除', cancelButtonText: '取消'
    })
    for (const row of selection.value) {
      await request.delete(`/dh114/favorites/${row.id}`)
    }
    ElMessage.success('批量删除完成')
    await loadList()
  } catch (e) { /* cancel */ }
}

// ===== 新建 =====
const formVisible = ref(false)
const formLoading = ref(false)
const formRef = ref(null)
const form = reactive({
  user_id: undefined, dh114_id: undefined, target_type: 'business',
  target_id: undefined, group_name: '', remark: ''
})
const rules = {
  user_id: [{ required: true, message: '请输入用户ID', trigger: 'blur' }],
  dh114_id: [{ required: true, message: '请输入商户ID', trigger: 'blur' }]
}

const openCreate = () => {
  Object.assign(form, {
    user_id: undefined, dh114_id: undefined, target_type: 'business',
    target_id: undefined, group_name: '', remark: ''
  })
  formVisible.value = true
}

const onSubmit = async () => {
  try {
    await formRef.value.validate()
    formLoading.value = true
    await request.post('/dh114/favorites', { ...form })
    ElMessage.success('创建成功')
    formVisible.value = false
    await loadList()
  } catch (e) {
    // 校验或接口失败
  } finally {
    formLoading.value = false
  }
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
.pagination-wrap { margin-top: 16px; display: flex; justify-content: flex-end; }
</style>
