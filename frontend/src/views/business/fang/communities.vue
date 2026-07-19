<template>
  <div class="app-container">
    <div class="page-card">
      <el-form :inline="true" :model="filters" class="filter-form" @submit.prevent>
        <el-form-item label="关键词">
          <el-input v-model="filters.keyword" placeholder="小区名/区域" clearable style="width: 200px" :prefix-icon="Search" @keyup.enter="onSearch" />
        </el-form-item>
        <el-form-item label="城市">
          <el-input v-model="filters.city" placeholder="城市" clearable style="width: 120px" @keyup.enter="onSearch" />
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
        <el-button type="primary" :icon="Plus" @click="openCreate">新建小区</el-button>
      </div>

      <el-table v-loading="loading" :data="list" border stripe>
        <el-table-column prop="id" label="ID" width="70" fixed="left" />
        <el-table-column prop="name" label="小区名称" min-width="180" show-overflow-tooltip />
        <el-table-column prop="city" label="城市" width="100" />
        <el-table-column prop="district" label="区域" width="100" />
        <el-table-column prop="address" label="地址" min-width="200" show-overflow-tooltip />
        <el-table-column label="均价" width="120">
          <template #default="{ row }">{{ row.avg_price ? `${row.avg_price}元/㎡` : '-' }}</template>
        </el-table-column>
        <el-table-column prop="house_count" label="房源数" width="90" />
        <el-table-column label="状态" width="90">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'info'" size="small">{{ row.status === 1 ? '启用' : '禁用' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="创建时间" width="160">
          <template #default="{ row }">{{ formatTime(row.created_at) }}</template>
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

    <el-dialog v-model="formVisible" :title="formMode === 'create' ? '新建小区' : '编辑小区'" width="600px" destroy-on-close>
      <el-form ref="formRef" :model="form" :rules="rules" label-width="100px">
        <el-form-item label="小区名称" prop="name">
          <el-input v-model="form.name" placeholder="请输入小区名称" />
        </el-form-item>
        <el-form-item label="城市" prop="city">
          <el-input v-model="form.city" placeholder="如：北京" />
        </el-form-item>
        <el-form-item label="区域" prop="district">
          <el-input v-model="form.district" placeholder="如：朝阳区" />
        </el-form-item>
        <el-form-item label="地址" prop="address">
          <el-input v-model="form.address" placeholder="详细地址" />
        </el-form-item>
        <el-form-item label="均价(元/㎡)">
          <el-input-number v-model="form.avg_price" :min="0" :controls="false" style="width: 200px" />
        </el-form-item>
        <el-form-item label="经度">
          <el-input-number v-model="form.longitude" :controls="false" :precision="6" style="width: 200px" />
        </el-form-item>
        <el-form-item label="纬度">
          <el-input-number v-model="form.latitude" :controls="false" :precision="6" style="width: 200px" />
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
import { ElMessage, ElMessageBox } from 'element-plus'
import { Search, RefreshLeft, Refresh, Plus } from '@element-plus/icons-vue'
import { adminListCommunities, adminUpdateCommunityStatus, createCommunity, updateCommunity, getCommunity } from '@/api/house'
import { formatTime } from '@/utils/format'

const loading = ref(false)
const list = ref([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const filters = reactive({ keyword: '', city: '', status: '' })

const formVisible = ref(false)
const formMode = ref('create')
const formRef = ref(null)
const submitting = ref(false)
const form = reactive({ id: null, name: '', city: '', district: '', address: '', avg_price: 0, longitude: null, latitude: null, description: '' })
const rules = {
  name: [{ required: true, message: '请输入小区名称', trigger: 'blur' }],
  city: [{ required: true, message: '请输入城市', trigger: 'blur' }]
}

const buildParams = () => {
  const p = { page: page.value, page_size: pageSize.value }
  if (filters.keyword) p.keyword = filters.keyword
  if (filters.city) p.city = filters.city
  if (filters.status !== '' && filters.status !== null) p.status = filters.status
  return p
}

const loadList = async () => {
  loading.value = true
  try {
    const res = await adminListCommunities(buildParams())
    const d = res.data || {}
    list.value = d.list || []
    total.value = d.total || 0
  } catch (e) {
    list.value = []
  } finally {
    loading.value = false
  }
}

const onSearch = () => { page.value = 1; loadList() }
const onReset = () => {
  Object.assign(filters, { keyword: '', city: '', status: '' })
  onSearch()
}

const resetForm = () => {
  Object.assign(form, { id: null, name: '', city: '', district: '', address: '', avg_price: 0, longitude: null, latitude: null, description: '' })
}

const openCreate = () => {
  formMode.value = 'create'
  resetForm()
  formVisible.value = true
}

const openEdit = async (row) => {
  formMode.value = 'edit'
  resetForm()
  try {
    const res = await getCommunity(row.id)
    Object.assign(form, res.data || row)
  } catch (e) {
    Object.assign(form, row)
  }
  formVisible.value = true
}

const handleSubmit = async () => {
  try {
    await formRef.value.validate()
    submitting.value = true
    if (formMode.value === 'create') {
      await createCommunity({ ...form })
      ElMessage.success('创建成功')
    } else {
      await updateCommunity(form.id, { ...form })
      ElMessage.success('更新成功')
    }
    formVisible.value = false
    loadList()
  } catch (e) {
    if (e?.message) ElMessage.error(e.message)
  } finally {
    submitting.value = false
  }
}

const handleStatus = async (row, status) => {
  try {
    await ElMessageBox.confirm(`确定${status === 1 ? '启用' : '禁用'}该小区吗？`, '提示', { type: 'warning' })
    await adminUpdateCommunityStatus(row.id, { status })
    ElMessage.success('操作成功')
    loadList()
  } catch (e) { /* cancel */ }
}

const handleDelete = async (row) => {
  try {
    await ElMessageBox.confirm('确定删除该小区吗？', '危险操作', { type: 'error' })
    await updateCommunity(row.id, { deleted: true })
    ElMessage.success('已删除')
    loadList()
  } catch (e) { /* cancel */ }
}

onMounted(() => loadList())
</script>
