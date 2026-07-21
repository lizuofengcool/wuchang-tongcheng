<template>
  <div class="app-container">
    <div class="page-card">
      <el-form :inline="true" :model="filters" class="filter-form" @submit.prevent>
        <el-form-item label="关键词">
          <el-input v-model="filters.keyword" placeholder="优惠券标题" clearable style="width: 200px" :prefix-icon="Search" @keyup.enter="onSearch" />
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="filters.status" placeholder="全部" clearable style="width: 120px" @change="onSearch">
            <el-option label="启用" :value="1" />
            <el-option label="禁用" :value="0" />
          </el-select>
        </el-form-item>
        <el-form-item label="店铺ID">
          <el-input-number v-model="filters.shop_id" :controls="false" placeholder="店铺ID" style="width: 140px" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" :icon="Search" @click="onSearch">搜索</el-button>
          <el-button :icon="RefreshLeft" @click="onReset">重置</el-button>
        </el-form-item>
      </el-form>

      <div class="toolbar">
        <div class="toolbar-left">
          <el-button :icon="Refresh" @click="loadList">刷新</el-button>
          <el-button type="primary" :icon="Plus" @click="openCreate">新增优惠券</el-button>
        </div>
      </div>

      <el-table v-loading="loading" :data="list" border stripe @sort-change="onSortChange">
        <el-table-column prop="id" label="ID" width="70" sortable="custom" fixed="left" />
        <el-table-column label="标题" min-width="200">
          <template #default="{ row }">{{ row.title || '-' }}</template>
        </el-table-column>
        <el-table-column label="金额" width="100">
          <template #default="{ row }">¥{{ Number(row.amount || 0).toFixed(2) }}</template>
        </el-table-column>
        <el-table-column label="门槛" width="100">
          <template #default="{ row }">满 ¥{{ Number(row.threshold || 0).toFixed(2) }}</template>
        </el-table-column>
        <el-table-column label="适用范围" width="160">
          <template #default="{ row }">{{ row.applicable_text || '-' }}</template>
        </el-table-column>
        <el-table-column label="库存" width="80" prop="stock" />
        <el-table-column label="已领" width="80" prop="received_count" />
        <el-table-column label="已用" width="80" prop="used_count" />
        <el-table-column label="有效期" width="320">
          <template #default="{ row }">
            <div class="text-xs">{{ formatTime(row.start_time) }}</div>
            <div class="text-xs text-muted">至 {{ formatTime(row.end_time) }}</div>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="90">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'info'" size="small">{{ row.status === 1 ? '启用' : '禁用' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="创建时间" width="160" prop="created_at" sortable="custom">
          <template #default="{ row }">{{ formatTime(row.created_at) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="180" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="openEdit(row)">编辑</el-button>
            <el-button v-if="row.status === 1" type="warning" link size="small" @click="handleStatus(row, 0)">禁用</el-button>
            <el-button v-else type="success" link size="small" @click="handleStatus(row, 1)">启用</el-button>
            <el-button type="danger" link size="small" @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>

      <div class="pagination-wrap">
        <el-pagination v-model:current-page="page" v-model:page-size="pageSize" :total="total" :page-sizes="[10, 20, 50, 100]" layout="total, sizes, prev, pager, next, jumper" background @current-change="loadList" @size-change="loadList" />
      </div>
    </div>

    <!-- 新增/编辑弹窗 -->
    <el-dialog v-model="formVisible" :title="formType === 'create' ? '新增优惠券' : '编辑优惠券'" width="600px" destroy-on-close>
      <el-form ref="formRef" :model="form" :rules="formRules" label-width="100px">
        <el-form-item label="标题" prop="title">
          <el-input v-model="form.title" placeholder="请输入优惠券标题" maxlength="100" show-word-limit />
        </el-form-item>
        <el-form-item label="金额" prop="amount">
          <el-input-number v-model="form.amount" :min="0" :precision="2" :controls="false" style="width: 200px" />
          <span class="text-muted text-xs" style="margin-left: 8px">元</span>
        </el-form-item>
        <el-form-item label="使用门槛" prop="threshold">
          <el-input-number v-model="form.threshold" :min="0" :precision="2" :controls="false" style="width: 200px" />
          <span class="text-muted text-xs" style="margin-left: 8px">满此金额可用，0 表示无门槛</span>
        </el-form-item>
        <el-form-item label="店铺ID">
          <el-input-number v-model="form.shop_id" :min="0" :controls="false" style="width: 200px" />
          <span class="text-muted text-xs" style="margin-left: 8px">0 表示全场通用</span>
        </el-form-item>
        <el-form-item label="库存" prop="stock">
          <el-input-number v-model="form.stock" :min="0" :controls="false" style="width: 200px" />
        </el-form-item>
        <el-form-item label="适用范围">
          <el-input v-model="form.applicable_text" placeholder="如：全场通用 / 指定团购" maxlength="100" />
        </el-form-item>
        <el-form-item label="开始时间" prop="start_time">
          <el-date-picker v-model="form.start_time" type="datetime" value-format="YYYY-MM-DD HH:mm:ss" style="width: 240px" />
        </el-form-item>
        <el-form-item label="结束时间" prop="end_time">
          <el-date-picker v-model="form.end_time" type="datetime" value-format="YYYY-MM-DD HH:mm:ss" style="width: 240px" />
        </el-form-item>
        <el-form-item label="状态">
          <el-switch v-model="form.status" :active-value="1" :inactive-value="0" />
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="form.description" type="textarea" :rows="3" placeholder="请输入备注" maxlength="500" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="formVisible = false">取消</el-button>
        <el-button type="primary" :loading="formLoading" @click="onFormSubmit">确认</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Refresh, RefreshLeft, Search, Plus } from '@element-plus/icons-vue'
import { formatTime } from '@/utils/format'
import { getGroupbuyCouponList, createGroupbuyCoupon, updateGroupbuyCoupon, deleteGroupbuyCoupon } from '@/api/groupbuy'

const loading = ref(false)
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)
const list = ref([])
const sortField = ref('created_at')
const sortOrder = ref('descending')

const filters = reactive({ keyword: '', status: null, shop_id: null })

const onSearch = () => { page.value = 1; loadList() }
const onReset = () => {
  filters.keyword = ''
  filters.status = null
  filters.shop_id = null
  page.value = 1
  loadList()
}
const onSortChange = ({ prop, order }) => {
  sortField.value = prop || 'created_at'
  sortOrder.value = order || 'descending'
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
      shop_id: filters.shop_id || undefined,
      sort: sortField.value,
      order: sortOrder.value
    }
    const res = await getGroupbuyCouponList(params)
    const data = res.data || {}
    list.value = data.list || []
    total.value = data.total || 0
  } catch (e) {
    ElMessage.error('加载优惠券列表失败')
    list.value = []
    total.value = 0
  } finally {
    loading.value = false
  }
}

// 新增/编辑
const formVisible = ref(false)
const formLoading = ref(false)
const formType = ref('create')
const formRef = ref(null)
const form = reactive({
  id: null, title: '', amount: 0, threshold: 0, shop_id: 0, stock: 0,
  applicable_text: '', start_time: null, end_time: null, status: 1, description: ''
})
const formRules = {
  title: [{ required: true, message: '请输入标题', trigger: 'blur' }],
  amount: [{ required: true, message: '请输入金额', trigger: 'blur' }],
  stock: [{ required: true, message: '请输入库存', trigger: 'blur' }],
  start_time: [{ required: true, message: '请选择开始时间', trigger: 'change' }],
  end_time: [{ required: true, message: '请选择结束时间', trigger: 'change' }]
}

const openCreate = () => {
  formType.value = 'create'
  Object.assign(form, {
    id: null, title: '', amount: 0, threshold: 0, shop_id: 0, stock: 0,
    applicable_text: '', start_time: null, end_time: null, status: 1, description: ''
  })
  formVisible.value = true
}

const openEdit = (row) => {
  formType.value = 'edit'
  Object.assign(form, {
    id: row.id, title: row.title || '', amount: Number(row.amount || 0), threshold: Number(row.threshold || 0),
    shop_id: Number(row.shop_id || 0), stock: Number(row.stock || 0),
    applicable_text: row.applicable_text || '', start_time: row.start_time, end_time: row.end_time,
    status: row.status, description: row.description || ''
  })
  formVisible.value = true
}

const onFormSubmit = async () => {
  if (!formRef.value) return
  try {
    await formRef.value.validate()
    formLoading.value = true
    const payload = { ...form }
    delete payload.id
    if (formType.value === 'create') {
      await createGroupbuyCoupon(payload)
      ElMessage.success('创建成功')
    } else {
      await updateGroupbuyCoupon(form.id, payload)
      ElMessage.success('更新成功')
    }
    formVisible.value = false
    await loadList()
  } catch (e) {
    if (e && e.message) ElMessage.error(e.message)
  } finally {
    formLoading.value = false
  }
}

const handleStatus = async (row, status) => {
  try {
    const action = status === 1 ? '启用' : '禁用'
    await ElMessageBox.confirm(`确定${action}优惠券 "${row.title}" 吗？`, '提示', { type: 'warning' })
    await updateGroupbuyCoupon(row.id, { status })
    ElMessage.success(`${action}成功`)
    await loadList()
  } catch (e) { /* cancel */ }
}

const handleDelete = async (row) => {
  try {
    await ElMessageBox.confirm(`确定删除优惠券 "${row.title}" 吗？此操作不可恢复！`, '危险操作', { type: 'warning', confirmButtonText: '确认删除', cancelButtonText: '取消' })
    await deleteGroupbuyCoupon(row.id)
    ElMessage.success('删除成功')
    await loadList()
  } catch (e) { /* cancel */ }
}

onMounted(() => loadList())
</script>

<style scoped>
.filter-form { margin-bottom: 12px; padding: 12px 16px; background: #fafafa; border-radius: 4px; }
.filter-form :deep(.el-form-item) { margin-bottom: 8px; margin-right: 12px; }
.toolbar { display: flex; justify-content: space-between; align-items: center; flex-wrap: wrap; gap: 8px; margin-bottom: 12px; }
.toolbar-left { display: flex; gap: 8px; flex-wrap: wrap; }
.pagination-wrap { margin-top: 16px; display: flex; justify-content: flex-end; }
.text-muted { color: #909399; }
.text-xs { font-size: 12px; }
</style>
