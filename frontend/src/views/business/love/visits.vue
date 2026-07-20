<template>
  <div class="app-container">
    <!-- 顶部统计卡片 -->
    <el-row :gutter="16" class="stat-row">
      <el-col :xs="12" :sm="8" :md="6">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value">{{ stats.total }}</div><div class="stat-label">总访问数</div></div></el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="6">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value text-success">{{ stats.today }}</div><div class="stat-label">今日访问</div></div></el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="6">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value text-primary">{{ stats.activeUsers }}</div><div class="stat-label">活跃访客</div></div></el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="6">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value text-warning">{{ stats.unread }}</div><div class="stat-label">未读访客</div></div></el-card>
      </el-col>
    </el-row>

    <div class="page-card">
      <el-form :inline="true" :model="filters" class="filter-form" @submit.prevent>
        <el-form-item label="关键词">
          <el-input v-model="filters.keyword" placeholder="会员ID/昵称" clearable style="width: 200px" :prefix-icon="Search" @keyup.enter="onSearch" />
        </el-form-item>
        <el-form-item label="访问来源">
          <el-select v-model="filters.source" placeholder="全部" clearable style="width: 140px" @change="onSearch">
            <el-option label="列表浏览" value="list" />
            <el-option label="匹配页" value="match" />
            <el-option label="搜索" value="search" />
            <el-option label="分享" value="share" />
            <el-option label="推送" value="push" />
          </el-select>
        </el-form-item>
        <el-form-item label="是否已读">
          <el-select v-model="filters.is_read" placeholder="全部" clearable style="width: 120px" @change="onSearch">
            <el-option label="已读" :value="1" />
            <el-option label="未读" :value="0" />
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

      <div class="toolbar">
        <el-button :icon="Refresh" @click="loadList">刷新</el-button>
        <el-button type="success" :icon="Check" @click="onMarkAllRead">全部已读</el-button>
      </div>

      <el-table v-loading="loading" :data="list" border stripe @sort-change="onSortChange">
        <el-table-column prop="id" label="ID" width="70" sortable="custom" />
        <el-table-column label="访客" min-width="160">
          <template #default="{ row }">
            <div class="user-cell">
              <el-image v-if="row.visitor_avatar" :src="row.visitor_avatar" fit="cover" class="user-avatar" />
              <div class="user-info">
                <div class="user-name">{{ row.visitor_name || `#${row.visitor_id}` }}</div>
                <div class="user-meta">{{ row.visitor_gender === 'male' ? '男' : '女' }} · {{ row.visitor_age || '?' }}岁</div>
              </div>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="被访问者" min-width="160">
          <template #default="{ row }">
            <div class="user-cell">
              <el-image v-if="row.target_avatar" :src="row.target_avatar" fit="cover" class="user-avatar" />
              <div class="user-info">
                <div class="user-name">{{ row.target_name || `#${row.target_user_id}` }}</div>
                <div class="user-meta">{{ row.target_gender === 'male' ? '男' : '女' }} · {{ row.target_age || '?' }}岁</div>
              </div>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="来源" width="110">
          <template #default="{ row }">
            <el-tag :type="sourceTagType(row.source)" size="small">{{ sourceText(row.source) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="停留" width="90">
          <template #default="{ row }">
            <span v-if="row.duration">{{ row.duration }}s</span>
            <span v-else class="text-muted">-</span>
          </template>
        </el-table-column>
        <el-table-column label="已读" width="80">
          <template #default="{ row }">
            <el-tag v-if="row.is_read" type="success" size="small">已读</el-tag>
            <el-tag v-else type="warning" size="small">未读</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="访问时间" width="160" prop="created_at" sortable="custom">
          <template #default="{ row }">{{ formatTime(row.created_at) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="120" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="openDetail(row)">详情</el-button>
            <el-button v-if="!row.is_read" type="success" link size="small" @click="onMarkRead(row)">标记已读</el-button>
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

    <!-- 详情弹窗 -->
    <el-dialog v-model="detailVisible" title="访客详情" width="600px">
      <el-descriptions v-if="detail" :column="2" border>
        <el-descriptions-item label="ID">{{ detail.id }}</el-descriptions-item>
        <el-descriptions-item label="访客ID">{{ detail.visitor_id }}</el-descriptions-item>
        <el-descriptions-item label="访客">{{ detail.visitor_name || `#${detail.visitor_id}` }}</el-descriptions-item>
        <el-descriptions-item label="被访问者">{{ detail.target_name || `#${detail.target_user_id}` }}</el-descriptions-item>
        <el-descriptions-item label="被访问ID">{{ detail.target_user_id }}</el-descriptions-item>
        <el-descriptions-item label="来源">
          <el-tag size="small">{{ sourceText(detail.source) }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="停留时长">{{ detail.duration || 0 }} 秒</el-descriptions-item>
        <el-descriptions-item label="是否已读">
          <el-tag v-if="detail.is_read" type="success" size="small">已读</el-tag>
          <el-tag v-else type="warning" size="small">未读</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="设备">{{ detail.device || '-' }}</el-descriptions-item>
        <el-descriptions-item label="IP">{{ detail.ip || '-' }}</el-descriptions-item>
        <el-descriptions-item label="访问时间" :span="2">{{ formatTime(detail.created_at) }}</el-descriptions-item>
      </el-descriptions>
      <template #footer><el-button @click="detailVisible = false">关闭</el-button></template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Refresh, RefreshLeft, Search, Check } from '@element-plus/icons-vue'
import request from '@/utils/request'
import { formatTime } from '@/utils/format'

const loading = ref(false)
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)
const list = ref([])
const sortField = ref('created_at')
const sortOrder = ref('descending')

const filters = reactive({ keyword: '', source: '', is_read: null, dateRange: null })

const stats = reactive({ total: 0, today: 0, activeUsers: 0, unread: 0 })

const sourceText = (s) => ({ list: '列表浏览', match: '匹配页', search: '搜索', share: '分享', push: '推送' }[s] || s || '-')
const sourceTagType = (s) => ({ list: '', match: 'success', search: 'warning', share: 'info', push: 'danger' }[s] || 'info')

const onSearch = () => { page.value = 1; loadList() }
const onReset = () => {
  filters.keyword = ''; filters.source = ''; filters.is_read = null; filters.dateRange = null
  page.value = 1; loadList()
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
      source: filters.source || undefined,
      is_read: filters.is_read === null || filters.is_read === '' ? undefined : filters.is_read,
      sort: sortField.value,
      order: sortOrder.value
    }
    if (filters.dateRange && filters.dateRange.length === 2) {
      params.start_date = filters.dateRange[0]
      params.end_date = filters.dateRange[1]
    }
    const res = await request.get('/love/visits', { params })
    const data = res.data || {}
    list.value = data.list || []
    total.value = data.total || 0
    calcStats()
    if (data.stats) {
      stats.unread = data.stats.unread || 0
      stats.activeUsers = data.stats.active_users || 0
    }
  } catch (e) {
    list.value = []
    total.value = 0
  } finally {
    loading.value = false
  }
}

const calcStats = () => {
  const today = new Date().toISOString().slice(0, 10)
  stats.total = list.value.length
  stats.today = list.value.filter((r) => (r.created_at || '').startsWith(today)).length
  stats.unread = list.value.filter((r) => !r.is_read).length
  const visitorIds = new Set(list.value.map((r) => r.visitor_id))
  stats.activeUsers = visitorIds.size
}

const detailVisible = ref(false)
const detail = ref(null)
const openDetail = (row) => { detail.value = row; detailVisible.value = true }

const onMarkRead = async (row) => {
  try {
    await request.post(`/love/visits/${row.id}/read`)
    ElMessage.success('已标记已读')
    await loadList()
  } catch (e) { /* fail */ }
}

const onMarkAllRead = async () => {
  try {
    await ElMessageBox.confirm('确认将所有未读访客标记为已读？', '提示', { type: 'warning' })
    await request.post('/love/visits/read-all')
    ElMessage.success('操作成功')
    await loadList()
  } catch (e) { /* cancel */ }
}

onMounted(() => loadList())
</script>

<style scoped>
.stat-row { margin-bottom: 16px; }
.stat-item { text-align: center; padding: 16px; }
.stat-value { font-size: 22px; font-weight: 600; color: #303133; }
.text-success { color: #67c23a; }
.text-warning { color: #e6a23c; }
.text-primary { color: #409eff; }
.filter-form { margin-bottom: 12px; padding: 12px 16px; background: #fafafa; border-radius: 4px; }
.filter-form :deep(.el-form-item) { margin-bottom: 8px; margin-right: 12px; }
.toolbar { display: flex; justify-content: flex-start; gap: 8px; margin-bottom: 12px; }
.text-muted { color: #909399; }
.user-cell { display: flex; align-items: center; gap: 8px; }
.user-avatar { width: 36px; height: 36px; border-radius: 50%; border: 1px solid #ebeef5; }
.user-info { display: flex; flex-direction: column; }
.user-name { color: #303133; font-size: 13px; }
.user-meta { color: #909399; font-size: 12px; margin-top: 2px; }
.pagination-wrap { margin-top: 16px; display: flex; justify-content: flex-end; }
</style>
