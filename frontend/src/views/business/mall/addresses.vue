<template>
  <div class="app-container">
    <!-- 地址统计 -->
    <el-row :gutter="16" class="stat-row">
      <el-col :xs="12" :sm="8" :md="6">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value">{{ stats.total }}</div><div class="stat-label">地址总数</div></div></el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="6">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value text-primary">{{ stats.distinct_users }}</div><div class="stat-label">用户数</div></div></el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="6">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value text-success">{{ stats.default_count }}</div><div class="stat-label">默认地址数</div></div></el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="6">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value text-warning">{{ stats.distinct_cities }}</div><div class="stat-label">涉及城市数</div></div></el-card>
      </el-col>
    </el-row>

    <div class="page-card">
      <el-form :inline="true" :model="filters" class="filter-form" @submit.prevent>
        <el-form-item label="用户ID">
          <el-input-number v-model="filters.user_id" :controls="false" :min="1" placeholder="用户ID" style="width: 140px" @keyup.enter="onSearch" />
        </el-form-item>
        <el-form-item label="收货人">
          <el-input v-model="filters.consignee" placeholder="收货人姓名" clearable style="width: 160px" @keyup.enter="onSearch" />
        </el-form-item>
        <el-form-item label="手机号">
          <el-input v-model="filters.phone" placeholder="手机号" clearable style="width: 160px" @keyup.enter="onSearch" />
        </el-form-item>
        <el-form-item label="是否默认">
          <el-select v-model="filters.is_default" placeholder="全部" clearable style="width: 120px" @change="onSearch">
            <el-option label="默认地址" :value="1" />
            <el-option label="非默认" :value="0" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" :icon="Search" @click="onSearch">搜索</el-button>
          <el-button :icon="RefreshLeft" @click="onReset">重置</el-button>
        </el-form-item>
      </el-form>

      <div class="toolbar">
        <el-button :icon="Refresh" @click="loadList">刷新</el-button>
      </div>

      <el-table v-loading="loading" :data="list" border stripe @sort-change="onSortChange">
        <el-table-column prop="id" label="ID" width="70" sortable="custom" fixed="left" />
        <el-table-column label="收货人" width="120">
          <template #default="{ row }">
            <el-link type="primary" :underline="false" @click="openDetail(row)">{{ row.consignee || '-' }}</el-link>
            <el-tag v-if="row.is_default" type="success" size="small" style="margin-left: 6px">默认</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="手机号" width="130" prop="phone" />
        <el-table-column label="省份" width="100" prop="province" />
        <el-table-column label="城市" width="100" prop="city" />
        <el-table-column label="区县" width="100" prop="district" />
        <el-table-column label="详细地址" min-width="220" show-overflow-tooltip>
          <template #default="{ row }">{{ row.address || row.detail || '-' }}</template>
        </el-table-column>
        <el-table-column label="用户ID" width="90" prop="user_id" />
        <el-table-column label="创建时间" width="160" prop="created_at" sortable="custom">
          <template #default="{ row }">{{ formatTime(row.created_at) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="100" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="openDetail(row)">详情</el-button>
          </template>
        </el-table-column>
      </el-table>

      <div class="pagination-wrap">
        <el-pagination v-model:current-page="page" v-model:page-size="pageSize" :total="total" :page-sizes="[10, 20, 50, 100]" layout="total, sizes, prev, pager, next, jumper" background @current-change="loadList" @size-change="loadList" />
      </div>
    </div>

    <!-- 详情弹窗 -->
    <el-dialog v-model="detailVisible" title="收货地址详情" width="640px" destroy-on-close>
      <el-descriptions v-if="detail" :column="2" border>
        <el-descriptions-item label="ID">{{ detail.id }}</el-descriptions-item>
        <el-descriptions-item label="用户ID">{{ detail.user_id }}</el-descriptions-item>
        <el-descriptions-item label="收货人">{{ detail.consignee || '-' }}</el-descriptions-item>
        <el-descriptions-item label="手机号">{{ detail.phone || '-' }}</el-descriptions-item>
        <el-descriptions-item label="省份">{{ detail.province || '-' }}</el-descriptions-item>
        <el-descriptions-item label="城市">{{ detail.city || '-' }}</el-descriptions-item>
        <el-descriptions-item label="区县">{{ detail.district || '-' }}</el-descriptions-item>
        <el-descriptions-item label="邮编">{{ detail.zipcode || detail.zip_code || '-' }}</el-descriptions-item>
        <el-descriptions-item label="详细地址" :span="2">{{ detail.address || detail.detail || '-' }}</el-descriptions-item>
        <el-descriptions-item label="完整地址" :span="2">{{ fullAddress(detail) }}</el-descriptions-item>
        <el-descriptions-item label="是否默认">
          <el-tag :type="detail.is_default ? 'success' : 'info'" size="small">{{ detail.is_default ? '默认' : '非默认' }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="状态">{{ detail.status === 1 ? '启用' : (detail.status === 0 ? '禁用' : '-') }}</el-descriptions-item>
        <el-descriptions-item label="创建时间">{{ formatTime(detail.created_at) }}</el-descriptions-item>
        <el-descriptions-item label="更新时间">{{ formatTime(detail.updated_at) }}</el-descriptions-item>
      </el-descriptions>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { Refresh, RefreshLeft, Search } from '@element-plus/icons-vue'
import { formatTime } from '@/utils/format'
import { getMallAddressList, getMallAddressDetail } from '@/api/mall'

const loading = ref(false)
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)
const list = ref([])
const sortField = ref('created_at')
const sortOrder = ref('descending')

const stats = reactive({ total: 0, distinct_users: 0, default_count: 0, distinct_cities: 0 })

const filters = reactive({ user_id: null, consignee: '', phone: '', is_default: null })

const onSearch = () => { page.value = 1; loadList() }
const onReset = () => {
  filters.user_id = null
  filters.consignee = ''
  filters.phone = ''
  filters.is_default = null
  page.value = 1
  loadList()
}
const onSortChange = ({ prop, order }) => {
  sortField.value = prop || 'created_at'
  sortOrder.value = order || 'descending'
  loadList()
}

const fullAddress = (d) => {
  return [d.province, d.city, d.district, d.address || d.detail].filter(Boolean).join('')
}

const loadList = async () => {
  loading.value = true
  try {
    const params = {
      page: page.value,
      page_size: pageSize.value,
      user_id: filters.user_id || undefined,
      consignee: filters.consignee.trim() || undefined,
      phone: filters.phone.trim() || undefined,
      is_default: filters.is_default === null || filters.is_default === '' ? undefined : filters.is_default,
      sort: sortField.value,
      order: sortOrder.value
    }
    const res = await getMallAddressList(params)
    const data = res.data || {}
    list.value = data.list || []
    total.value = data.total || 0
    computeStats()
  } catch (e) {
    ElMessage.error('加载地址列表失败')
    list.value = []
    total.value = 0
  } finally {
    loading.value = false
  }
}

const computeStats = () => {
  const total = list.value.length
  const distinctUsers = new Set(list.value.map((r) => r.user_id).filter(Boolean)).size
  const defaultCount = list.value.filter((r) => r.is_default).length
  const distinctCities = new Set(list.value.map((r) => r.city).filter(Boolean)).size
  Object.assign(stats, { total, distinct_users: distinctUsers, default_count: defaultCount, distinct_cities: distinctCities })
}

const detailVisible = ref(false)
const detail = ref(null)

const openDetail = async (row) => {
  try {
    const res = await getMallAddressDetail(row.id)
    detail.value = res.data || null
    detailVisible.value = true
  } catch (e) {
    ElMessage.error('加载详情失败')
  }
}

onMounted(() => loadList())
</script>

<style scoped>
.stat-row { margin-bottom: 16px; }
.stat-item { text-align: center; padding: 16px; }
.stat-value { font-size: 20px; font-weight: 600; color: #303133; }
.text-success { color: #67c23a; }
.text-warning { color: #e6a23c; }
.text-primary { color: #409eff; }
.filter-form { margin-bottom: 12px; padding: 12px 16px; background: #fafafa; border-radius: 4px; }
.filter-form :deep(.el-form-item) { margin-bottom: 8px; margin-right: 12px; }
.toolbar { display: flex; justify-content: space-between; margin-bottom: 12px; }
.pagination-wrap { margin-top: 16px; display: flex; justify-content: flex-end; }
</style>
