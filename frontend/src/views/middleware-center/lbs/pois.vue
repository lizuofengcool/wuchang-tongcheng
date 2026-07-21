<template>
  <div class="app-container">
    <div class="page-card">
      <div class="toolbar">
        <div class="toolbar-left">
          <el-button type="primary" :icon="Plus" @click="openCreate">新建POI</el-button>
          <el-button :icon="Refresh" @click="loadList">刷新</el-button>
          <el-button :icon="Search" @click="nearbyVisible = true">附近检索</el-button>
        </div>
        <div class="toolbar-right">
          <el-input
            v-model="search"
            placeholder="名称/地址"
            clearable
            style="width: 200px"
            :prefix-icon="Search"
            @keyup.enter="onSearch"
            @clear="onSearch"
          />
          <el-select
            v-model="categoryFilter"
            placeholder="分类"
            clearable
            style="width: 160px; margin-left: 8px"
            @change="onSearch"
          >
            <el-option v-for="c in categories" :key="c" :label="c" :value="c" />
          </el-select>
          <el-select
            v-model="statusFilter"
            placeholder="状态"
            clearable
            style="width: 120px; margin-left: 8px"
            @change="onSearch"
          >
            <el-option label="上线" :value="1" />
            <el-option label="下架" :value="0" />
            <el-option label="待审" :value="2" />
            <el-option label="拒绝" :value="3" />
          </el-select>
          <el-button type="primary" :icon="Search" style="margin-left: 8px" @click="onSearch">搜索</el-button>
        </div>
      </div>

      <el-table v-loading="loading" :data="list" border stripe>
        <el-table-column prop="id" label="ID" width="70" />
        <el-table-column prop="region_id" label="地区" width="80" />
        <el-table-column label="名称" min-width="160">
          <template #default="{ row }">
            <div class="poi-name">
              <el-image v-if="row.icon" :src="row.icon" fit="cover" class="poi-icon" />
              <span>{{ row.name || '-' }}</span>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="category" label="分类" width="120" />
        <el-table-column prop="address" label="地址" min-width="180" show-overflow-tooltip />
        <el-table-column label="经纬度" width="180">
          <template #default="{ row }">
            <span v-if="row.latitude && row.longitude">{{ row.latitude.toFixed(7) }}, {{ row.longitude.toFixed(7) }}</span>
            <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column prop="phone" label="电话" width="130" />
        <el-table-column label="状态" width="90">
          <template #default="{ row }">
            <el-tag :type="statusTagType(row.status)" size="small">{{ row.status_text || statusText(row.status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="创建时间" width="170" />
        <el-table-column label="操作" width="220" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" size="small" @click="openEdit(row)">编辑</el-button>
            <el-button
              link
              :type="row.status === 1 ? 'warning' : 'success'"
              size="small"
              @click="toggleStatus(row)"
            >
              {{ row.status === 1 ? '下架' : '上线' }}
            </el-button>
            <el-button link type="danger" size="small" @click="onDelete(row)">删除</el-button>
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
          @size-change="loadList"
          @current-change="loadList"
        />
      </div>
    </div>

    <!-- 新建/编辑弹窗 -->
    <el-dialog v-model="formVisible" :title="isEdit ? '编辑POI' : '新建POI'" width="640px" @closed="resetForm">
      <el-form ref="formRef" :model="form" :rules="rules" label-width="90px">
        <el-form-item label="名称" prop="name">
          <el-input v-model="form.name" maxlength="200" placeholder="请输入POI名称" />
        </el-form-item>
        <el-form-item label="分类" prop="category">
          <el-select v-model="form.category" placeholder="选择分类" filterable allow-create style="width: 100%">
            <el-option v-for="c in categories" :key="c" :label="c" :value="c" />
          </el-select>
        </el-form-item>
        <el-form-item label="地址" prop="address">
          <el-input v-model="form.address" maxlength="500" placeholder="详细地址" />
        </el-form-item>
        <el-form-item label="经纬度" prop="latitude">
          <el-input-number v-model="form.latitude" :precision="7" :step="0.0000001" :min="-90" :max="90" style="width: 200px" />
          <el-input-number v-model="form.longitude" :precision="7" :step="0.0000001" :min="-180" :max="180" style="width: 200px; margin-left: 8px" />
        </el-form-item>
        <el-form-item label="电话" prop="phone">
          <el-input v-model="form.phone" maxlength="32" />
        </el-form-item>
        <el-form-item label="图标URL" prop="icon">
          <el-input v-model="form.icon" maxlength="255" placeholder="图标 URL" />
        </el-form-item>
        <el-form-item label="来源" prop="source">
          <el-radio-group v-model="form.source">
            <el-radio value="manual">手动</el-radio>
            <el-radio value="amap">高德</el-radio>
            <el-radio value="import">导入</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="状态" prop="status">
          <el-radio-group v-model="form.status">
            <el-radio :value="1">上线</el-radio>
            <el-radio :value="0">下架</el-radio>
            <el-radio :value="2">待审</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="标签" prop="tags_text">
          <el-input v-model="form.tags_text" type="textarea" :rows="2" placeholder='JSON 数组，如 ["wifi","停车位"]' />
        </el-form-item>
        <el-form-item label="扩展" prop="extra_text">
          <el-input v-model="form.extra_text" type="textarea" :rows="2" placeholder='JSON 对象' />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="formVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="onSubmit">确定</el-button>
      </template>
    </el-dialog>

    <!-- 附近检索弹窗 -->
    <el-dialog v-model="nearbyVisible" title="附近 POI 检索" width="600px">
      <el-form :model="nearby" label-width="100px">
        <el-form-item label="纬度">
          <el-input-number v-model="nearby.latitude" :precision="7" :step="0.0000001" style="width: 100%" />
        </el-form-item>
        <el-form-item label="经度">
          <el-input-number v-model="nearby.longitude" :precision="7" :step="0.0000001" style="width: 100%" />
        </el-form-item>
        <el-form-item label="半径(km)">
          <el-input-number v-model="nearby.radius_km" :min="0.1" :max="100" :step="0.5" style="width: 100%" />
        </el-form-item>
        <el-form-item label="分类">
          <el-select v-model="nearby.category" placeholder="全部" clearable style="width: 100%">
            <el-option v-for="c in categories" :key="c" :label="c" :value="c" />
          </el-select>
        </el-form-item>
        <el-form-item label="关键字">
          <el-input v-model="nearby.keyword" placeholder="名称/地址" />
        </el-form-item>
      </el-form>
      <el-table :data="nearbyList" border stripe style="margin-top: 12px">
        <el-table-column prop="id" label="ID" width="60" />
        <el-table-column prop="name" label="名称" min-width="120" />
        <el-table-column prop="category" label="分类" width="100" />
        <el-table-column label="距离(km)" width="100">
          <template #default="{ row }">
            <el-tag type="success" size="small">{{ row.distance ? row.distance.toFixed(2) : '-' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="address" label="地址" min-width="140" show-overflow-tooltip />
      </el-table>
      <template #footer>
        <el-button @click="nearbyVisible = false">关闭</el-button>
        <el-button type="primary" @click="onNearbySearch">检索</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Refresh, Search } from '@element-plus/icons-vue'
import {
  getLbsAdminPoiList,
  getLbsAdminPoiCategories,
  createLbsPoi,
  updateLbsPoi,
  deleteLbsPoi,
  updateLbsAdminPoiStatus,
  deleteLbsAdminPoi,
  getLbsPoiNearby
} from '@/api/lbs'

const list = ref([])
const loading = ref(false)
const page = ref(1)
const pageSize = ref(10)
const total = ref(0)
const search = ref('')
const categoryFilter = ref('')
const statusFilter = ref(null)
const categories = ref([])

const formVisible = ref(false)
const isEdit = ref(false)
const submitting = ref(false)
const formRef = ref(null)
const form = reactive({
  id: 0,
  name: '',
  category: '',
  address: '',
  latitude: 0,
  longitude: 0,
  phone: '',
  icon: '',
  source: 'manual',
  status: 1,
  tags_text: '',
  extra_text: ''
})

const rules = {
  name: [{ required: true, message: '请输入名称', trigger: 'blur' }],
  latitude: [{ required: true, message: '请输入纬度', trigger: 'blur' }],
  longitude: [{ required: true, message: '请输入经度', trigger: 'blur' }]
}

const nearbyVisible = ref(false)
const nearby = reactive({
  latitude: 39.9042,
  longitude: 116.4074,
  radius_km: 5,
  category: '',
  keyword: ''
})
const nearbyList = ref([])

function statusText(s) {
  const map = { 0: '下架', 1: '上线', 2: '待审', 3: '拒绝', 4: '删除' }
  return map[s] || '-'
}
function statusTagType(s) {
  return s === 1 ? 'success' : s === 2 ? 'warning' : s === 3 ? 'danger' : 'info'
}

async function loadList() {
  loading.value = true
  try {
    const params = {
      page: page.value,
      page_size: pageSize.value,
      keyword: search.value,
      category: categoryFilter.value
    }
    if (statusFilter.value !== null && statusFilter.value !== '') {
      params.status = statusFilter.value
    }
    const res = await getLbsAdminPoiList(params)
    list.value = res.data?.list || []
    total.value = res.data?.total || 0
  } catch (e) {
    ElMessage.error('列表加载失败')
  } finally {
    loading.value = false
  }
}

async function loadCategories() {
  try {
    const res = await getLbsAdminPoiCategories()
    categories.value = res.data || []
  } catch (e) {
    categories.value = []
  }
}

function onSearch() {
  page.value = 1
  loadList()
}

function openCreate() {
  isEdit.value = false
  Object.assign(form, {
    id: 0, name: '', category: '', address: '', latitude: 0, longitude: 0,
    phone: '', icon: '', source: 'manual', status: 1, tags_text: '', extra_text: ''
  })
  formVisible.value = true
}

function openEdit(row) {
  isEdit.value = true
  Object.assign(form, {
    id: row.id, name: row.name, category: row.category, address: row.address,
    latitude: row.latitude, longitude: row.longitude, phone: row.phone, icon: row.icon,
    source: row.source || 'manual', status: row.status,
    tags_text: row.tags ? JSON.stringify(row.tags) : '',
    extra_text: row.extra ? JSON.stringify(row.extra) : ''
  })
  formVisible.value = true
}

function resetForm() {
  formRef.value?.resetFields()
}

async function onSubmit() {
  await formRef.value?.validate()
  submitting.value = true
  try {
    const data = {
      name: form.name,
      category: form.category,
      address: form.address,
      latitude: form.latitude,
      longitude: form.longitude,
      phone: form.phone,
      icon: form.icon,
      source: form.source,
      status: form.status
    }
    if (form.tags_text) {
      try { data.tags = JSON.parse(form.tags_text) } catch (e) { data.tags = null }
    }
    if (form.extra_text) {
      try { data.extra = JSON.parse(form.extra_text) } catch (e) { data.extra = null }
    }
    if (isEdit.value) {
      await updateLbsPoi(form.id, data)
      ElMessage.success('更新成功')
    } else {
      await createLbsPoi(data)
      ElMessage.success('创建成功')
    }
    formVisible.value = false
    loadList()
  } catch (e) {
    ElMessage.error(e?.message || '操作失败')
  } finally {
    submitting.value = false
  }
}

async function toggleStatus(row) {
  const newStatus = row.status === 1 ? 0 : 1
  try {
    await updateLbsAdminPoiStatus(row.id, newStatus)
    ElMessage.success('状态已更新')
    loadList()
  } catch (e) {
    ElMessage.error('更新失败')
  }
}

async function onDelete(row) {
  try {
    await ElMessageBox.confirm(`确定删除 POI "${row.name}" 吗？`, '提示', { type: 'warning' })
    await deleteLbsAdminPoi(row.id)
    ElMessage.success('删除成功')
    loadList()
  } catch (e) {
    if (e !== 'cancel') ElMessage.error('删除失败')
  }
}

async function onNearbySearch() {
  try {
    const res = await getLbsPoiNearby({
      latitude: nearby.latitude,
      longitude: nearby.longitude,
      radius_km: nearby.radius_km,
      category: nearby.category,
      keyword: nearby.keyword,
      page: 1,
      page_size: 20
    })
    nearbyList.value = res.data?.list || []
    if (nearbyList.value.length === 0) ElMessage.info('未找到附近 POI')
  } catch (e) {
    ElMessage.error('检索失败')
  }
}

onMounted(() => {
  loadCategories()
  loadList()
})
</script>

<style scoped>
.page-card {
  background: #fff;
  padding: 16px;
  border-radius: 4px;
}
.toolbar {
  display: flex;
  justify-content: space-between;
  margin-bottom: 12px;
}
.toolbar-right {
  display: flex;
  align-items: center;
}
.pagination-wrap {
  margin-top: 12px;
  display: flex;
  justify-content: flex-end;
}
.poi-name {
  display: flex;
  align-items: center;
  gap: 8px;
}
.poi-icon {
  width: 32px;
  height: 32px;
  border-radius: 4px;
}
</style>
