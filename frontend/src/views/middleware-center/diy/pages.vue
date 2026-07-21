<template>
  <div class="app-container">
    <div class="page-card">
      <el-form :inline="true" :model="filters" class="filter-form" @submit.prevent>
        <el-form-item label="页面类型">
          <el-select v-model="filters.type" placeholder="全部" clearable style="width: 140px" @change="onSearch">
            <el-option v-for="(label, val) in typeMap" :key="val" :label="label" :value="val" />
          </el-select>
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="filters.status" placeholder="全部" clearable style="width: 120px" @change="onSearch">
            <el-option v-for="(label, val) in statusMap" :key="val" :label="label" :value="Number(val)" />
          </el-select>
        </el-form-item>
        <el-form-item label="关键词">
          <el-input v-model="filters.keyword" placeholder="标题" clearable style="width: 200px" @keyup.enter="onSearch" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" :icon="Search" @click="onSearch">搜索</el-button>
          <el-button :icon="RefreshLeft" @click="onReset">重置</el-button>
        </el-form-item>
      </el-form>

      <div class="toolbar">
        <el-button :icon="Refresh" @click="loadList">刷新</el-button>
      </div>

      <el-table v-loading="loading" :data="list" border stripe>
        <el-table-column prop="id" label="ID" width="70" />
        <el-table-column prop="title" label="标题" min-width="180" show-overflow-tooltip />
        <el-table-column label="类型" width="100">
          <template #default="{ row }">
            <el-tag size="small" :type="typeTagType(row.type)">{{ row.type_text || typeMap[row.type] || row.type }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="slug" label="Slug" min-width="150" show-overflow-tooltip />
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="statusTagType(row.status)" size="small">{{ row.status_text || statusMap[row.status] || '未知' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="user_id" label="创建者" width="90" />
        <el-table-column prop="biz_id" label="业务ID" width="90" />
        <el-table-column label="发布时间" width="160">
          <template #default="{ row }">{{ formatTime(row.published_at) }}</template>
        </el-table-column>
        <el-table-column label="更新时间" width="160">
          <template #default="{ row }">{{ formatTime(row.updated_at) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="220" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="openEditor(row)">编辑</el-button>
            <el-button v-if="row.status === 1" type="warning" link size="small" @click="onOffline(row)">强制下线</el-button>
            <el-button v-if="row.status === 2" type="success" link size="small" @click="onRestore(row)">恢复发布</el-button>
            <el-button type="info" link size="small" @click="onViewStats(row)">统计</el-button>
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

    <!-- 统计弹窗 -->
    <el-dialog v-model="statVisible" :title="`页面统计 - #${statPageId}`" width="640px" destroy-on-close>
      <div v-loading="statLoading">
        <el-descriptions :column="3" border>
          <el-descriptions-item label="总浏览">{{ statSummary.total_view || 0 }}</el-descriptions-item>
          <el-descriptions-item label="总点击">{{ statSummary.total_click || 0 }}</el-descriptions-item>
          <el-descriptions-item label="总转化">{{ statSummary.total_conversion || 0 }}</el-descriptions-item>
        </el-descriptions>
        <div class="stat-list-title">明细（按日期）</div>
        <el-table :data="statList" border size="small" max-height="320">
          <el-table-column prop="stat_date" label="日期" width="120">
            <template #default="{ row }">{{ formatDate(row.stat_date) }}</template>
          </el-table-column>
          <el-table-column prop="view_count" label="浏览" width="90" />
          <el-table-column prop="click_count" label="点击" width="90" />
          <el-table-column prop="conversion_count" label="转化" width="90" />
        </el-table>
      </div>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Refresh, RefreshLeft, Search } from '@element-plus/icons-vue'
import {
  getDiyPageList,
  updateDiyPageStatus,
  getDiyStatSummaryByPage,
  getDiyStatListByPage
} from '@/api/diy'

const router = useRouter()

const loading = ref(false)
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)
const list = ref([])

const filters = reactive({ type: '', status: null, keyword: '' })

const typeMap = {
  home: '首页',
  topic: '专题页',
  shop: '店铺页',
  activity: '活动页'
}
const typeTagType = (t) => ({
  home: 'danger',
  topic: 'warning',
  shop: 'success',
  activity: 'primary'
}[t] || 'info')

const statusMap = { 0: '草稿', 1: '已发布', 2: '已下线' }
const statusTagType = (s) => ({ 0: 'info', 1: 'success', 2: 'danger' }[s] || 'info')

const formatTime = (t) => {
  if (!t) return '-'
  return String(t).replace('T', ' ').slice(0, 16)
}
const formatDate = (t) => {
  if (!t) return '-'
  return String(t).slice(0, 10)
}

const onSearch = () => { page.value = 1; loadList() }
const onReset = () => {
  filters.type = ''
  filters.status = null
  filters.keyword = ''
  page.value = 1
  loadList()
}

const loadList = async () => {
  loading.value = true
  try {
    const params = {
      page: page.value,
      page_size: pageSize.value,
      type: filters.type || undefined,
      status: filters.status === null || filters.status === '' ? undefined : filters.status,
      keyword: filters.keyword || undefined
    }
    const res = await getDiyPageList(params)
    const data = res.data || {}
    list.value = data.list || []
    total.value = data.total || 0
  } catch (e) {
    list.value = []
    total.value = 0
  } finally {
    loading.value = false
  }
}

const openEditor = (row) => {
  router.push({ name: 'DiyEditor', query: { id: row.id } })
}

const onOffline = (row) => {
  ElMessageBox.confirm(`确认强制下线页面「${row.title}」？`, '提示', {
    type: 'warning',
    confirmButtonText: '确认',
    cancelButtonText: '取消'
  }).then(async () => {
    try {
      await updateDiyPageStatus(row.id, { status: 2 })
      ElMessage.success('已下线')
      loadList()
    } catch (e) {}
  }).catch(() => {})
}

const onRestore = (row) => {
  ElMessageBox.confirm(`确认恢复发布页面「${row.title}」？`, '提示', {
    type: 'warning',
    confirmButtonText: '确认',
    cancelButtonText: '取消'
  }).then(async () => {
    try {
      await updateDiyPageStatus(row.id, { status: 1 })
      ElMessage.success('已恢复发布')
      loadList()
    } catch (e) {}
  }).catch(() => {})
}

// ===== 统计 =====
const statVisible = ref(false)
const statLoading = ref(false)
const statPageId = ref(null)
const statSummary = ref({})
const statList = ref([])

const onViewStats = async (row) => {
  statPageId.value = row.id
  statVisible.value = true
  statLoading.value = true
  statSummary.value = {}
  statList.value = []
  try {
    const [sumRes, listRes] = await Promise.all([
      getDiyStatSummaryByPage(row.id),
      getDiyStatListByPage(row.id, { page: 1, page_size: 50 })
    ])
    statSummary.value = sumRes.data || {}
    const ld = listRes.data || {}
    statList.value = ld.list || []
  } catch (e) {
  } finally {
    statLoading.value = false
  }
}

onMounted(() => {
  loadList()
})
</script>

<style scoped>
.page-card { background: #fff; padding: 16px; border-radius: 8px; box-shadow: 0 1px 4px rgba(0,0,0,0.04); }
.filter-form { margin-bottom: 12px; }
.toolbar { margin-bottom: 12px; display: flex; gap: 8px; }
.pagination-wrap { margin-top: 16px; display: flex; justify-content: flex-end; }
.stat-list-title { margin: 16px 0 8px; font-weight: 600; color: #303133; }
</style>
