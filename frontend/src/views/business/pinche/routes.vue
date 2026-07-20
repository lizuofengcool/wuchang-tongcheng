<template>
  <div class="app-container">
    <!-- 顶部统计卡片 -->
    <el-row :gutter="16" class="stat-row">
      <el-col :xs="12" :sm="8" :md="6">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-icon" style="background: #409eff">
            <el-icon :size="22"><Position /></el-icon>
          </div>
          <div class="stat-content">
            <div class="stat-value">{{ stats.total }}</div>
            <div class="stat-label">总路线数</div>
          </div>
        </el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="6">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-icon" style="background: #67c23a">
            <el-icon :size="22"><Star /></el-icon>
          </div>
          <div class="stat-content">
            <div class="stat-value">{{ stats.favorites }}</div>
            <div class="stat-label">收藏路线</div>
          </div>
        </el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="6">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-icon" style="background: #e6a23c">
            <el-icon :size="22"><TrendCharts /></el-icon>
          </div>
          <div class="stat-content">
            <div class="stat-value">{{ stats.hotRoutes }}</div>
            <div class="stat-label">热门路线</div>
          </div>
        </el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="6">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-icon" style="background: #722ed1">
            <el-icon :size="22"><Clock /></el-icon>
          </div>
          <div class="stat-content">
            <div class="stat-value">{{ stats.todayUsed }}</div>
            <div class="stat-label">今日使用</div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <div class="page-card">
      <!-- 筛选区 -->
      <el-form :inline="true" :model="filters" class="filter-form" @submit.prevent>
        <el-form-item label="关键词">
          <el-input
            v-model="filters.keyword"
            placeholder="出发地/目的地"
            clearable
            style="width: 220px"
            :prefix-icon="Search"
            @keyup.enter="onSearch"
          />
        </el-form-item>
        <el-form-item label="路线类型">
          <el-select v-model="filters.route_type" placeholder="全部" clearable style="width: 140px" @change="onSearch">
            <el-option label="常用路线" value="common" />
            <el-option label="自定义路线" value="custom" />
            <el-option label="推荐路线" value="recommend" />
          </el-select>
        </el-form-item>
        <el-form-item label="收藏">
          <el-select v-model="filters.is_favorite" placeholder="全部" clearable style="width: 120px" @change="onSearch">
            <el-option label="仅收藏" :value="1" />
            <el-option label="未收藏" :value="0" />
          </el-select>
        </el-form-item>
        <el-form-item label="时间范围">
          <el-date-picker
            v-model="filters.dateRange"
            type="daterange"
            range-separator="至"
            start-placeholder="开始日期"
            end-placeholder="结束日期"
            value-format="YYYY-MM-DD"
            style="width: 240px"
          />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" :icon="Search" @click="onSearch">搜索</el-button>
          <el-button :icon="RefreshLeft" @click="onReset">重置</el-button>
        </el-form-item>
      </el-form>

      <!-- 工具栏 -->
      <div class="toolbar">
        <div class="toolbar-left">
          <el-button :icon="Refresh" @click="loadList">刷新</el-button>
          <el-button type="danger" :icon="Delete" :disabled="!selection.length" @click="onBatchDelete">批量删除</el-button>
        </div>
        <div class="toolbar-right">
          <el-button type="primary" :icon="Plus" @click="openCreate">新建路线</el-button>
        </div>
      </div>

      <!-- 表格 -->
      <el-table
        v-loading="loading"
        :data="list"
        border
        stripe
        @selection-change="onSelectionChange"
      >
        <el-table-column type="selection" width="44" fixed="left" />
        <el-table-column prop="id" label="ID" width="70" fixed="left" />
        <el-table-column label="路线" min-width="220">
          <template #default="{ row }">
            <div class="route-cell">
              <div class="route-text">
                <span class="from">{{ row.origin }}</span>
                <el-icon class="arrow"><Right /></el-icon>
                <span class="to">{{ row.destination }}</span>
              </div>
              <div v-if="row.waypoints && row.waypoints.length" class="route-way">
                途经：{{ row.waypoints.join(' → ') }}
              </div>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="类型" width="120">
          <template #default="{ row }">
            <el-tag :type="routeTypeTagType(row.route_type)" size="small">{{ routeTypeText(row.route_type) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="distance" label="距离(km)" width="110" />
        <el-table-column prop="duration" label="时长(分)" width="110" />
        <el-table-column prop="use_count" label="使用次数" width="100" sortable />
        <el-table-column prop="fav_count" label="收藏数" width="90" sortable />
        <el-table-column label="收藏" width="80">
          <template #default="{ row }">
            <el-icon v-if="row.is_favorite" color="#e6a23c"><Star /></el-icon>
            <span v-else class="text-muted">-</span>
          </template>
        </el-table-column>
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
            <el-button type="primary" link size="small" @click="openDetail(row)">详情</el-button>
            <el-button type="warning" link size="small" @click="openEdit(row)">编辑</el-button>
            <el-button type="danger" link size="small" @click="onDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>

      <!-- 分页 -->
      <div class="pagination-wrap">
        <el-pagination
          v-model:current-page="page"
          v-model:page-size="pageSize"
          :total="total"
          :page-sizes="[10, 20, 50, 100]"
          layout="total, sizes, prev, pager, next, jumper"
          background
          @current-change="loadList"
          @size-change="loadList"
        />
      </div>
    </div>

    <!-- 详情弹窗 -->
    <el-dialog v-model="detailVisible" title="路线详情" width="700px">
      <el-descriptions v-if="detail" :column="2" border>
        <el-descriptions-item label="ID">{{ detail.id }}</el-descriptions-item>
        <el-descriptions-item label="类型">
          <el-tag :type="routeTypeTagType(detail.route_type)" size="small">{{ routeTypeText(detail.route_type) }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="出发地" :span="2">{{ detail.origin }}</el-descriptions-item>
        <el-descriptions-item label="目的地" :span="2">{{ detail.destination }}</el-descriptions-item>
        <el-descriptions-item label="途经" :span="2">
          <span v-if="detail.waypoints && detail.waypoints.length">{{ detail.waypoints.join(' → ') }}</span>
          <span v-else class="text-muted">-</span>
        </el-descriptions-item>
        <el-descriptions-item label="距离">{{ detail.distance ? detail.distance + ' km' : '-' }}</el-descriptions-item>
        <el-descriptions-item label="时长">{{ detail.duration ? detail.duration + ' 分钟' : '-' }}</el-descriptions-item>
        <el-descriptions-item label="出发地经纬度">{{ detail.origin_lng || '-' }}, {{ detail.origin_lat || '-' }}</el-descriptions-item>
        <el-descriptions-item label="目的地经纬度">{{ detail.dest_lng || '-' }}, {{ detail.dest_lat || '-' }}</el-descriptions-item>
        <el-descriptions-item label="使用次数">{{ detail.use_count }}</el-descriptions-item>
        <el-descriptions-item label="收藏数">{{ detail.fav_count }}</el-descriptions-item>
        <el-descriptions-item label="状态">
          <el-tag :type="detail.status === 1 ? 'success' : 'info'" size="small">{{ detail.status === 1 ? '启用' : '禁用' }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="创建时间">{{ formatTime(detail.created_at) }}</el-descriptions-item>
        <el-descriptions-item label="用户ID" :span="2">{{ detail.user_id || '-' }}</el-descriptions-item>
      </el-descriptions>
      <template #footer><el-button @click="detailVisible = false">关闭</el-button></template>
    </el-dialog>

    <!-- 新建/编辑弹窗 -->
    <el-dialog v-model="formVisible" :title="formTitle" width="640px">
      <el-form ref="formRef" :model="form" :rules="rules" label-width="100px">
        <el-form-item label="出发地" prop="origin">
          <el-input v-model="form.origin" maxlength="128" placeholder="出发地" />
        </el-form-item>
        <el-form-item label="目的地" prop="destination">
          <el-input v-model="form.destination" maxlength="128" placeholder="目的地" />
        </el-form-item>
        <el-form-item label="途经点">
          <el-input v-model="form.waypointsStr" type="textarea" :rows="2" placeholder="多个途经点用逗号分隔" />
        </el-form-item>
        <el-form-item label="路线类型" prop="route_type">
          <el-select v-model="form.route_type" style="width: 100%">
            <el-option label="常用路线" value="common" />
            <el-option label="自定义路线" value="custom" />
            <el-option label="推荐路线" value="recommend" />
          </el-select>
        </el-form-item>
        <el-form-item label="距离(km)">
          <el-input-number v-model="form.distance" :min="0" :precision="2" style="width: 100%" />
        </el-form-item>
        <el-form-item label="时长(分钟)">
          <el-input-number v-model="form.duration" :min="0" style="width: 100%" />
        </el-form-item>
        <el-form-item label="出发地经纬度">
          <el-input v-model="form.origin_lng" placeholder="经度" style="width: 50%" />
          <el-input v-model="form.origin_lat" placeholder="纬度" style="width: 50%" />
        </el-form-item>
        <el-form-item label="目的地经纬度">
          <el-input v-model="form.dest_lng" placeholder="经度" style="width: 50%" />
          <el-input v-model="form.dest_lat" placeholder="纬度" style="width: 50%" />
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
import request from '@/utils/request'
import {
  Refresh, RefreshLeft, Search, Delete, Plus,
  Position, Star, TrendCharts, Clock, Right
} from '@element-plus/icons-vue'
import { formatTime } from '@/utils/format'

const loading = ref(false)
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)
const list = ref([])
const selection = ref([])

const stats = reactive({ total: 0, favorites: 0, hotRoutes: 0, todayUsed: 0 })

const filters = reactive({
  keyword: '', route_type: '', is_favorite: null, dateRange: null
})

const routeTypeText = (t) => ({ common: '常用路线', custom: '自定义路线', recommend: '推荐路线' }[t] || '-')
const routeTypeTagType = (t) => ({ common: 'primary', custom: 'info', recommend: 'success' }[t] || 'info')

const onSearch = () => { page.value = 1; loadList() }
const onReset = () => {
  filters.keyword = ''; filters.route_type = ''; filters.is_favorite = null; filters.dateRange = null
  page.value = 1; loadList()
}

const onSelectionChange = (rows) => { selection.value = rows }

const loadList = async () => {
  loading.value = true
  try {
    const params = {
      page: page.value,
      page_size: pageSize.value,
      keyword: filters.keyword.trim() || undefined,
      route_type: filters.route_type || undefined,
      is_favorite: filters.is_favorite === null || filters.is_favorite === '' ? undefined : filters.is_favorite
    }
    if (filters.dateRange && filters.dateRange.length === 2) {
      params.start_date = filters.dateRange[0]
      params.end_date = filters.dateRange[1]
    }
    const res = await request.get('/pinche/routes', { params })
    const data = res.data || {}
    list.value = data.list || []
    total.value = data.total || 0
    stats.total = data.total || stats.total
  } catch (e) {
    list.value = []; total.value = 0
  } finally {
    loading.value = false
  }
}

// 详情弹窗
const detailVisible = ref(false)
const detail = ref(null)
const openDetail = (row) => { detail.value = row; detailVisible.value = true }

// 新建/编辑
const formVisible = ref(false)
const formLoading = ref(false)
const formRef = ref(null)
const isEdit = ref(false)
const formTitle = computed(() => isEdit.value ? '编辑路线' : '新建路线')
const form = reactive({
  id: null, origin: '', destination: '', waypointsStr: '',
  route_type: 'common', distance: 0, duration: 0,
  origin_lng: '', origin_lat: '', dest_lng: '', dest_lat: '',
  status: 1
})
const rules = {
  origin: [{ required: true, message: '请输入出发地', trigger: 'blur' }],
  destination: [{ required: true, message: '请输入目的地', trigger: 'blur' }],
  route_type: [{ required: true, message: '请选择类型', trigger: 'change' }]
}

const openCreate = () => {
  isEdit.value = false
  Object.assign(form, {
    id: null, origin: '', destination: '', waypointsStr: '',
    route_type: 'common', distance: 0, duration: 0,
    origin_lng: '', origin_lat: '', dest_lng: '', dest_lat: '',
    status: 1
  })
  formVisible.value = true
}

const openEdit = (row) => {
  isEdit.value = true
  Object.assign(form, {
    ...row,
    waypointsStr: row.waypoints ? row.waypoints.join(', ') : ''
  })
  formVisible.value = true
}

const onSubmit = async () => {
  try {
    await formRef.value.validate()
    const payload = {
      ...form,
      waypoints: form.waypointsStr ? form.waypointsStr.split(/[,，\n]/).map((s) => s.trim()).filter(Boolean) : []
    }
    delete payload.waypointsStr
    formLoading.value = true
    if (isEdit.value) {
      await request.put(`/pinche/routes/${form.id}`, payload)
    } else {
      await request.post('/pinche/routes', payload)
    }
    ElMessage.success(isEdit.value ? '更新成功' : '创建成功')
    formVisible.value = false
    await loadList()
  } catch (e) {
    // 校验失败或接口错误
  } finally {
    formLoading.value = false
  }
}

const onDelete = async (row) => {
  try {
    await ElMessageBox.confirm(`确认删除路线 "${row.origin} → ${row.destination}"？`, '提示', { type: 'warning' })
    await request.delete(`/pinche/routes/${row.id}`)
    ElMessage.success('删除成功')
    await loadList()
  } catch (e) { /* cancel */ }
}

const onBatchDelete = async () => {
  try {
    await ElMessageBox.confirm(`确认批量删除 ${selection.value.length} 个路线？`, '危险操作', {
      type: 'error', confirmButtonText: '确认删除', cancelButtonText: '取消'
    })
    await Promise.all(selection.value.map((r) => request.delete(`/pinche/routes/${r.id}`)))
    ElMessage.success('批量删除完成')
    await loadList()
  } catch (e) { /* cancel */ }
}

onMounted(() => loadList())
</script>

<style scoped>
.stat-row { margin-bottom: 16px; }
.stat-card { display: flex; align-items: center; }
.stat-card :deep(.el-card__body) {
  display: flex; align-items: center; gap: 14px; padding: 16px; width: 100%;
}
.stat-icon {
  width: 44px; height: 44px; border-radius: 8px;
  display: flex; align-items: center; justify-content: center;
  color: #fff; flex-shrink: 0;
}
.stat-content { flex: 1; min-width: 0; }
.stat-value { font-size: 22px; font-weight: 600; color: #303133; }
.stat-label { font-size: 12px; color: #909399; margin-top: 2px; }

.filter-form { margin-bottom: 12px; padding: 12px 16px; background: #fafafa; border-radius: 4px; }
.filter-form :deep(.el-form-item) { margin-bottom: 8px; margin-right: 12px; }

.toolbar { display: flex; justify-content: space-between; align-items: center; flex-wrap: wrap; gap: 8px; margin-bottom: 12px; }
.toolbar-left { display: flex; gap: 8px; flex-wrap: wrap; }
.toolbar-right { display: flex; gap: 8px; }

.route-cell { display: flex; flex-direction: column; gap: 4px; min-width: 0; }
.route-text { font-weight: 500; color: #303133; display: flex; align-items: center; gap: 6px; }
.from, .to { color: #303133; }
.arrow { color: #909399; }
.route-way { font-size: 12px; color: #909399; }
.text-muted { color: #909399; }
.pagination-wrap { margin-top: 16px; display: flex; justify-content: flex-end; }
</style>
