<template>
  <div class="app-container">
    <div class="page-card">
      <div class="toolbar">
        <div class="toolbar-left">
          <el-button type="primary" :icon="Plus" @click="openCreate">新建围栏</el-button>
          <el-button :icon="Refresh" @click="loadList">刷新</el-button>
          <el-button :icon="Search" @click="checkVisible = true">围栏判断</el-button>
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
          <el-select
            v-model="typeFilter"
            placeholder="类型"
            clearable
            style="width: 120px; margin-left: 8px"
            @change="onSearch"
          >
            <el-option label="圆形" value="circle" />
            <el-option label="多边形" value="polygon" />
          </el-select>
          <el-select
            v-model="statusFilter"
            placeholder="状态"
            clearable
            style="width: 120px; margin-left: 8px"
            @change="onSearch"
          >
            <el-option label="启用" :value="1" />
            <el-option label="禁用" :value="0" />
          </el-select>
          <el-button type="primary" :icon="Search" style="margin-left: 8px" @click="onSearch">搜索</el-button>
        </div>
      </div>

      <el-table v-loading="loading" :data="list" border stripe>
        <el-table-column prop="id" label="ID" width="70" />
        <el-table-column prop="region_id" label="区域ID" width="80" />
        <el-table-column prop="name" label="围栏名称" min-width="160" />
        <el-table-column label="类型" width="100">
          <template #default="{ row }">
            <el-tag :type="row.type === 'circle' ? 'primary' : 'success'" size="small">
              {{ row.type === 'circle' ? '圆形' : '多边形' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="圆心" width="200">
          <template #default="{ row }">
            <span v-if="row.type === 'circle' && row.center_lat && row.center_lng">
              {{ row.center_lat.toFixed(7) }}, {{ row.center_lng.toFixed(7) }}
            </span>
            <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column label="半径(m)" width="100">
          <template #default="{ row }">
            <span v-if="row.type === 'circle'">{{ row.radius }}</span>
            <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column label="多边形顶点" width="120">
          <template #default="{ row }">
            <span v-if="row.type === 'polygon'">{{ polygonCount(row.points) }} 个</span>
            <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column prop="owner_type" label="所有者类型" width="120" />
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
    <el-dialog v-model="formVisible" :title="isEdit ? '编辑围栏' : '新建围栏'" width="700px" @closed="resetForm">
      <el-form ref="formRef" :model="form" :rules="rules" label-width="110px">
        <el-form-item label="围栏名称" prop="name">
          <el-input v-model="form.name" maxlength="100" placeholder="如：五常市中心配送范围" />
        </el-form-item>
        <el-form-item label="关联区域ID" prop="region_id">
          <el-input-number v-model="form.region_id" :min="0" style="width: 100%" />
        </el-form-item>
        <el-form-item label="围栏类型" prop="type">
          <el-radio-group v-model="form.type" @change="onTypeChange">
            <el-radio value="circle">圆形</el-radio>
            <el-radio value="polygon">多边形</el-radio>
          </el-radio-group>
        </el-form-item>
        <template v-if="form.type === 'circle'">
          <el-form-item label="圆心纬度" prop="center_lat">
            <el-input-number v-model="form.center_lat" :precision="7" :step="0.0000001" :min="-90" :max="90" style="width: 100%" />
          </el-form-item>
          <el-form-item label="圆心经度" prop="center_lng">
            <el-input-number v-model="form.center_lng" :precision="7" :step="0.0000001" :min="-180" :max="180" style="width: 100%" />
          </el-form-item>
          <el-form-item label="半径(米)" prop="radius">
            <el-input-number v-model="form.radius" :min="1" :max="1000000" :step="100" style="width: 100%" />
          </el-form-item>
        </template>
        <template v-else>
          <el-form-item label="多边形顶点" prop="points_text">
            <el-input
              v-model="form.points_text"
              type="textarea"
              :rows="6"
              placeholder='JSON 数组，至少 3 个点，如 [{"lat":44.93,"lng":127.15},{"lat":44.95,"lng":127.17},{"lat":44.91,"lng":127.17}]'
            />
            <div class="tip">顶点按顺序连接形成多边形围栏，至少 3 个点</div>
          </el-form-item>
        </template>
        <el-form-item label="所有者ID" prop="owner_id">
          <el-input-number v-model="form.owner_id" :min="0" style="width: 100%" />
        </el-form-item>
        <el-form-item label="所有者类型" prop="owner_type">
          <el-select v-model="form.owner_type" placeholder="选择类型" clearable style="width: 100%">
            <el-option label="商家 shop" value="shop" />
            <el-option label="代理商 agent" value="agent" />
            <el-option label="到家服务 daojia" value="daojia" />
          </el-select>
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
        <el-form-item label="描述" prop="description">
          <el-input v-model="form.description" type="textarea" :rows="2" maxlength="500" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="formVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="onSubmit">确定</el-button>
      </template>
    </el-dialog>

    <!-- 围栏判断弹窗 -->
    <el-dialog v-model="checkVisible" title="围栏判断" width="520px">
      <el-form :model="checkQuery" label-width="100px">
        <el-form-item label="围栏ID">
          <el-input-number v-model="checkQuery.id" :min="1" style="width: 100%" />
        </el-form-item>
        <el-form-item label="纬度">
          <el-input-number v-model="checkQuery.latitude" :precision="7" :step="0.0000001" style="width: 100%" />
        </el-form-item>
        <el-form-item label="经度">
          <el-input-number v-model="checkQuery.longitude" :precision="7" :step="0.0000001" style="width: 100%" />
        </el-form-item>
      </el-form>
      <el-descriptions v-if="checkResult" :column="1" border style="margin-top: 12px">
        <el-descriptions-item label="是否在围栏内">
          <el-tag :type="checkResult.inside ? 'success' : 'info'">{{ checkResult.inside ? '是' : '否' }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item v-if="checkResult.geofence_id" label="匹配围栏ID">{{ checkResult.geofence_id }}</el-descriptions-item>
        <el-descriptions-item v-if="checkResult.name" label="围栏名称">{{ checkResult.name }}</el-descriptions-item>
      </el-descriptions>
      <template #footer>
        <el-button @click="checkVisible = false">关闭</el-button>
        <el-button type="primary" @click="onCheck">判断</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Refresh, Search } from '@element-plus/icons-vue'
import {
  getLbsGeofenceList,
  createLbsGeofence,
  updateLbsGeofence,
  deleteLbsGeofence,
  updateLbsAdminGeofenceStatus,
  checkLbsPointInGeofence
} from '@/api/lbs'

const list = ref([])
const loading = ref(false)
const page = ref(1)
const pageSize = ref(10)
const total = ref(0)
const search = ref('')
const typeFilter = ref('')
const statusFilter = ref(null)

const formVisible = ref(false)
const isEdit = ref(false)
const submitting = ref(false)
const formRef = ref(null)
const form = reactive({
  id: 0, region_id: 0, name: '', type: 'circle', status: 1, sort: 0,
  description: '', center_lat: 0, center_lng: 0, radius: 1000,
  points_text: '', owner_id: 0, owner_type: ''
})

const rules = {
  name: [{ required: true, message: '请输入围栏名称', trigger: 'blur' }],
  type: [{ required: true, message: '请选择类型', trigger: 'change' }]
}

const checkVisible = ref(false)
const checkQuery = reactive({ id: 1, latitude: 44.9350, longitude: 127.1500 })
const checkResult = ref(null)

function polygonCount(points) {
  if (!points) return 0
  if (Array.isArray(points)) return points.length
  try {
    const arr = typeof points === 'string' ? JSON.parse(points) : points
    return Array.isArray(arr) ? arr.length : 0
  } catch (e) {
    return 0
  }
}

function onTypeChange() {
  // 切换类型时清空不相关字段
  if (form.type === 'circle') {
    form.points_text = ''
  } else {
    form.center_lat = 0
    form.center_lng = 0
    form.radius = 0
  }
}

async function loadList() {
  loading.value = true
  try {
    const params = {
      page: page.value,
      page_size: pageSize.value,
      keyword: search.value,
      type: typeFilter.value
    }
    if (statusFilter.value !== null && statusFilter.value !== '') {
      params.status = statusFilter.value
    }
    const res = await getLbsGeofenceList(params)
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
    id: 0, region_id: 0, name: '', type: 'circle', status: 1, sort: 0,
    description: '', center_lat: 0, center_lng: 0, radius: 1000,
    points_text: '', owner_id: 0, owner_type: ''
  })
  formVisible.value = true
}

function openEdit(row) {
  isEdit.value = true
  Object.assign(form, {
    id: row.id, region_id: row.region_id, name: row.name, type: row.type,
    status: row.status, sort: row.sort, description: row.description,
    center_lat: row.center_lat, center_lng: row.center_lng, radius: row.radius,
    points_text: row.points ? JSON.stringify(row.points) : '',
    owner_id: row.owner_id, owner_type: row.owner_type
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
      region_id: form.region_id,
      name: form.name,
      type: form.type,
      status: form.status,
      sort: form.sort,
      description: form.description,
      center_lat: form.center_lat,
      center_lng: form.center_lng,
      radius: form.radius,
      owner_id: form.owner_id,
      owner_type: form.owner_type
    }
    if (form.type === 'polygon' && form.points_text) {
      try {
        data.points = JSON.parse(form.points_text)
      } catch (e) {
        ElMessage.error('多边形顶点 JSON 格式错误')
        submitting.value = false
        return
      }
    }
    if (isEdit.value) {
      await updateLbsGeofence(form.id, data)
      ElMessage.success('更新成功')
    } else {
      await createLbsGeofence(data)
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
    await updateLbsAdminGeofenceStatus(row.id, newStatus)
    ElMessage.success('状态已更新')
    loadList()
  } catch (e) {
    ElMessage.error('更新失败')
  }
}

async function onDelete(row) {
  try {
    await ElMessageBox.confirm(`确定删除围栏 "${row.name}" 吗？`, '提示', { type: 'warning' })
    await deleteLbsGeofence(row.id)
    ElMessage.success('删除成功')
    loadList()
  } catch (e) {
    if (e !== 'cancel') ElMessage.error('删除失败')
  }
}

async function onCheck() {
  try {
    const res = await checkLbsPointInGeofence(checkQuery.id, {
      latitude: checkQuery.latitude,
      longitude: checkQuery.longitude
    })
    checkResult.value = res.data
  } catch (e) {
    ElMessage.error('判断失败')
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
