<template>
  <div class="app-container">
    <div class="page-card">
      <div class="toolbar">
        <div class="toolbar-left">
          <el-button type="primary" :icon="Plus" @click="openCreate">新建区域</el-button>
          <el-button :icon="Refresh" @click="loadList">刷新</el-button>
          <el-button :icon="Search" @click="locVisible = true">按经纬度查分站</el-button>
        </div>
        <div class="toolbar-right">
          <el-input
            v-model="search"
            placeholder="名称"
            clearable
            style="width: 200px"
            :prefix-icon="Search"
            @keyup.enter="onSearch"
            @clear="onSearch"
          />
          <el-input
            v-model="cityCodeFilter"
            placeholder="城市编码"
            clearable
            style="width: 140px; margin-left: 8px"
            @keyup.enter="onSearch"
            @clear="onSearch"
          />
          <el-select
            v-model="levelFilter"
            placeholder="层级"
            clearable
            style="width: 120px; margin-left: 8px"
            @change="onSearch"
          >
            <el-option label="省" :value="1" />
            <el-option label="市" :value="2" />
            <el-option label="区" :value="3" />
            <el-option label="乡镇" :value="4" />
          </el-select>
          <el-button type="primary" :icon="Search" style="margin-left: 8px" @click="onSearch">搜索</el-button>
        </div>
      </div>

      <el-table v-loading="loading" :data="list" border stripe>
        <el-table-column prop="id" label="ID" width="70" />
        <el-table-column prop="name" label="区域名称" min-width="160" />
        <el-table-column label="层级" width="80">
          <template #default="{ row }">
            <el-tag size="small">{{ levelText(row.level) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="city_code" label="城市编码" width="120" />
        <el-table-column prop="ad_code" label="行政区划代码" width="130" />
        <el-table-column label="中心点" width="200">
          <template #default="{ row }">
            <span v-if="row.center_lat && row.center_lng">
              {{ row.center_lat.toFixed(7) }}, {{ row.center_lng.toFixed(7) }}
            </span>
            <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column prop="parent_id" label="父级ID" width="80" />
        <el-table-column label="状态" width="90">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'info'" size="small">
              {{ row.status === 1 ? '启用' : '禁用' }}
            </el-tag>
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
              {{ row.status === 1 ? '禁用' : '启用' }}
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
    <el-dialog v-model="formVisible" :title="isEdit ? '编辑区域' : '新建区域'" width="680px" @closed="resetForm">
      <el-form ref="formRef" :model="form" :rules="rules" label-width="110px">
        <el-form-item label="区域名称" prop="name">
          <el-input v-model="form.name" maxlength="100" placeholder="如：五常市" />
        </el-form-item>
        <el-form-item label="城市编码" prop="city_code">
          <el-input v-model="form.city_code" maxlength="20" placeholder="城市编码" />
        </el-form-item>
        <el-form-item label="行政区划代码" prop="ad_code">
          <el-input v-model="form.ad_code" maxlength="20" placeholder="行政区划代码" />
        </el-form-item>
        <el-form-item label="父级ID" prop="parent_id">
          <el-input-number v-model="form.parent_id" :min="0" style="width: 100%" />
        </el-form-item>
        <el-form-item label="层级" prop="level">
          <el-radio-group v-model="form.level">
            <el-radio :value="1">省</el-radio>
            <el-radio :value="2">市</el-radio>
            <el-radio :value="3">区</el-radio>
            <el-radio :value="4">乡镇</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="中心纬度" prop="center_lat">
          <el-input-number v-model="form.center_lat" :precision="7" :step="0.0000001" :min="-90" :max="90" style="width: 100%" />
        </el-form-item>
        <el-form-item label="中心经度" prop="center_lng">
          <el-input-number v-model="form.center_lng" :precision="7" :step="0.0000001" :min="-180" :max="180" style="width: 100%" />
        </el-form-item>
        <el-form-item label="边界多边形" prop="boundary_text">
          <el-input
            v-model="form.boundary_text"
            type="textarea"
            :rows="4"
            placeholder='JSON 数组，如 [{"lat":44.93,"lng":127.15},{"lat":44.95,"lng":127.17},{"lat":44.91,"lng":127.17}]'
          />
          <div class="tip">多边形顶点（至少 3 个），按顺序连接形成区域边界</div>
        </el-form-item>
        <el-form-item label="排序" prop="sort">
          <el-input-number v-model="form.sort" :min="0" style="width: 100%" />
        </el-form-item>
        <el-form-item label="状态" prop="status">
          <el-radio-group v-model="form.status">
            <el-radio :value="1">启用</el-radio>
            <el-radio :value="0">禁用</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="邮编" prop="zip_code">
          <el-input v-model="form.zip_code" maxlength="20" />
        </el-form-item>
        <el-form-item label="描述" prop="description">
          <el-input v-model="form.description" type="textarea" :rows="2" maxlength="500" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="formVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="onSubmit">确定</el-button>
      </template>
    </el-dialog>

    <!-- 经纬度查分站弹窗 -->
    <el-dialog v-model="locVisible" title="按经纬度判断分站" width="480px">
      <el-form :model="locQuery" label-width="80px">
        <el-form-item label="纬度">
          <el-input-number v-model="locQuery.latitude" :precision="7" :step="0.0000001" style="width: 100%" />
        </el-form-item>
        <el-form-item label="经度">
          <el-input-number v-model="locQuery.longitude" :precision="7" :step="0.0000001" style="width: 100%" />
        </el-form-item>
      </el-form>
      <el-descriptions v-if="locResult" :column="1" border style="margin-top: 12px">
        <el-descriptions-item label="分站ID">{{ locResult.region_id }}</el-descriptions-item>
        <el-descriptions-item label="名称">{{ locResult.name }}</el-descriptions-item>
        <el-descriptions-item label="城市编码">{{ locResult.city_code }}</el-descriptions-item>
        <el-descriptions-item label="行政区划">{{ locResult.ad_code }}</el-descriptions-item>
        <el-descriptions-item label="是否在区域内">
          <el-tag :type="locResult.inside ? 'success' : 'info'">{{ locResult.inside ? '是' : '否' }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="数据来源">{{ locResult.source }}</el-descriptions-item>
      </el-descriptions>
      <template #footer>
        <el-button @click="locVisible = false">关闭</el-button>
        <el-button type="primary" @click="onLocQuery">查询</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Refresh, Search } from '@element-plus/icons-vue'
import {
  getLbsRegionList,
  createLbsRegion,
  updateLbsRegion,
  deleteLbsRegion,
  updateLbsAdminRegionStatus,
  getLbsRegionByLocation
} from '@/api/lbs'

const list = ref([])
const loading = ref(false)
const page = ref(1)
const pageSize = ref(10)
const total = ref(0)
const search = ref('')
const cityCodeFilter = ref('')
const levelFilter = ref(null)

const formVisible = ref(false)
const isEdit = ref(false)
const submitting = ref(false)
const formRef = ref(null)
const form = reactive({
  id: 0, name: '', city_code: '', ad_code: '', parent_id: 0, level: 3,
  center_lat: 0, center_lng: 0, boundary_text: '', sort: 0, status: 1,
  zip_code: '', description: ''
})

const rules = {
  name: [{ required: true, message: '请输入区域名称', trigger: 'blur' }],
  level: [{ required: true, message: '请选择层级', trigger: 'change' }]
}

const locVisible = ref(false)
const locQuery = reactive({ latitude: 44.9350, longitude: 127.1500 })
const locResult = ref(null)

function levelText(l) {
  return { 1: '省', 2: '市', 3: '区', 4: '乡镇' }[l] || '-'
}

async function loadList() {
  loading.value = true
  try {
    const params = {
      page: page.value,
      page_size: pageSize.value,
      keyword: search.value,
      city_code: cityCodeFilter.value
    }
    if (levelFilter.value) params.level = levelFilter.value
    const res = await getLbsRegionList(params)
    list.value = res.data?.list || []
    total.value = res.data?.total || 0
  } catch (e) {
    ElMessage.error('列表加载失败')
  } finally {
    loading.value = false
  }
}

function onSearch() {
  page.value = 1
  loadList()
}

function openCreate() {
  isEdit.value = false
  Object.assign(form, {
    id: 0, name: '', city_code: '', ad_code: '', parent_id: 0, level: 3,
    center_lat: 0, center_lng: 0, boundary_text: '', sort: 0, status: 1,
    zip_code: '', description: ''
  })
  formVisible.value = true
}

function openEdit(row) {
  isEdit.value = true
  Object.assign(form, {
    id: row.id, name: row.name, city_code: row.city_code, ad_code: row.ad_code,
    parent_id: row.parent_id, level: row.level,
    center_lat: row.center_lat, center_lng: row.center_lng,
    boundary_text: row.boundary ? JSON.stringify(row.boundary) : '',
    sort: row.sort, status: row.status,
    zip_code: row.zip_code, description: row.description
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
      city_code: form.city_code,
      ad_code: form.ad_code,
      parent_id: form.parent_id,
      level: form.level,
      center_lat: form.center_lat,
      center_lng: form.center_lng,
      sort: form.sort,
      status: form.status,
      zip_code: form.zip_code,
      description: form.description
    }
    if (form.boundary_text) {
      try { data.boundary = JSON.parse(form.boundary_text) } catch (e) { data.boundary = null }
    }
    if (isEdit.value) {
      await updateLbsRegion(form.id, data)
      ElMessage.success('更新成功')
    } else {
      await createLbsRegion(data)
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
    await updateLbsAdminRegionStatus(row.id, newStatus)
    ElMessage.success('状态已更新')
    loadList()
  } catch (e) {
    ElMessage.error('更新失败')
  }
}

async function onDelete(row) {
  try {
    await ElMessageBox.confirm(`确定删除区域 "${row.name}" 吗？`, '提示', { type: 'warning' })
    await deleteLbsRegion(row.id)
    ElMessage.success('删除成功')
    loadList()
  } catch (e) {
    if (e !== 'cancel') ElMessage.error('删除失败')
  }
}

async function onLocQuery() {
  try {
    const res = await getLbsRegionByLocation(locQuery.latitude, locQuery.longitude)
    locResult.value = res.data
    if (!locResult.value || !locResult.value.region_id) {
      ElMessage.warning('未找到匹配的分站')
    }
  } catch (e) {
    ElMessage.error('查询失败')
  }
}

onMounted(() => {
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
.tip {
  font-size: 12px;
  color: #909399;
  margin-top: 4px;
}
</style>
