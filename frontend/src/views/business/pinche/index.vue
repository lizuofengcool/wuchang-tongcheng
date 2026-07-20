<template>
  <div class="app-container">
    <!-- 顶部统计卡片 -->
    <el-row :gutter="16" class="stat-row">
      <el-col :xs="12" :sm="8" :md="4">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-icon" style="background: #409eff">
            <el-icon :size="22"><Van /></el-icon>
          </div>
          <div class="stat-content">
            <div class="stat-value">{{ stats.total }}</div>
            <div class="stat-label">总拼车数</div>
          </div>
        </el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="4">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-icon" style="background: #67c23a">
            <el-icon :size="22"><CirclePlus /></el-icon>
          </div>
          <div class="stat-content">
            <div class="stat-value">{{ stats.todayNew }}</div>
            <div class="stat-label">今日新增</div>
          </div>
        </el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="4">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-icon" style="background: #e6a23c">
            <el-icon :size="22"><Clock /></el-icon>
          </div>
          <div class="stat-content">
            <div class="stat-value">{{ stats.pendingAudit }}</div>
            <div class="stat-label">待审核</div>
          </div>
        </el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="4">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-icon" style="background: #13c2c2">
            <el-icon :size="22"><Promotion /></el-icon>
          </div>
          <div class="stat-content">
            <div class="stat-value">{{ stats.published }}</div>
            <div class="stat-label">已发布</div>
          </div>
        </el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="4">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-icon" style="background: #722ed1">
            <el-icon :size="22"><Finished /></el-icon>
          </div>
          <div class="stat-content">
            <div class="stat-value">{{ stats.completed }}</div>
            <div class="stat-label">已成行</div>
          </div>
        </el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="4">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-icon" style="background: #f56c6c">
            <el-icon :size="22"><Warning /></el-icon>
          </div>
          <div class="stat-content">
            <div class="stat-value">{{ stats.violation }}</div>
            <div class="stat-label">违规数</div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 高级筛选区 -->
    <div class="page-card">
      <el-form :inline="true" :model="filters" class="filter-form" @submit.prevent>
        <el-form-item label="关键词">
          <el-input
            v-model="filters.keyword"
            placeholder="出发地/目的地/发布者"
            clearable
            style="width: 220px"
            :prefix-icon="Search"
            @keyup.enter="onSearch"
          />
        </el-form-item>
        <el-form-item label="拼车类型">
          <el-select v-model="filters.type" placeholder="全部" clearable style="width: 130px" @change="onSearch">
            <el-option label="人找车" value="passenger" />
            <el-option label="车找人" value="driver" />
            <el-option label="车找货" value="cargo" />
            <el-option label="货找车" value="cargo_need" />
          </el-select>
        </el-form-item>
        <el-form-item label="座位/空位">
          <el-input-number v-model="filters.min_seats" :min="0" :controls="false" placeholder="最少" style="width: 90px" />
          <span style="margin: 0 4px">-</span>
          <el-input-number v-model="filters.max_seats" :min="0" :controls="false" placeholder="最多" style="width: 90px" />
        </el-form-item>
        <el-form-item label="价格区间">
          <el-input-number v-model="filters.min_price" :min="0" :controls="false" placeholder="最低" style="width: 100px" />
          <span style="margin: 0 4px">-</span>
          <el-input-number v-model="filters.max_price" :min="0" :controls="false" placeholder="最高" style="width: 100px" />
        </el-form-item>
        <el-form-item label="审核状态">
          <el-select v-model="filters.audit_status" placeholder="全部" clearable style="width: 120px" @change="onSearch">
            <el-option label="待审核" :value="0" />
            <el-option label="已通过" :value="1" />
            <el-option label="已拒绝" :value="2" />
          </el-select>
        </el-form-item>
        <el-form-item label="拼车状态">
          <el-select v-model="filters.status" placeholder="全部" clearable style="width: 120px" @change="onSearch">
            <el-option label="草稿" :value="0" />
            <el-option label="已发布" :value="1" />
            <el-option label="已满员" :value="2" />
            <el-option label="已下架" :value="3" />
            <el-option label="已过期" :value="4" />
            <el-option label="已取消" :value="5" />
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
          <el-button type="success" :icon="Check" :disabled="!selection.length" @click="onBatchAudit(1)">批量通过</el-button>
          <el-button type="warning" :icon="Bottom" :disabled="!selection.length" @click="onBatchStatus(3)">批量下架</el-button>
          <el-button type="danger" :icon="Delete" :disabled="!selection.length" @click="onBatchDelete">批量删除</el-button>
        </div>
        <div class="toolbar-right">
          <el-button type="primary" :icon="Plus" @click="onCreate">新建拼车</el-button>
        </div>
      </div>

      <!-- 表格 -->
      <el-table
        v-loading="loading"
        :data="list"
        border
        stripe
        @selection-change="onSelectionChange"
        @sort-change="onSortChange"
      >
        <el-table-column type="selection" width="44" fixed="left" />
        <el-table-column prop="id" label="ID" width="70" sortable="custom" fixed="left" />
        <el-table-column label="路线" min-width="240">
          <template #default="{ row }">
            <div class="route-cell">
              <div class="route-text">
                <el-link type="primary" :underline="'never'" @click="openDetail(row)">
                  <span class="from">{{ row.origin }}</span>
                  <el-icon class="arrow"><Right /></el-icon>
                  <span class="to">{{ row.destination }}</span>
                </el-link>
                <el-tag v-if="row.is_urgent" type="danger" size="small" effect="dark">急</el-tag>
              </div>
              <div class="route-desc">
                <span class="price">¥{{ Number(row.price || 0).toFixed(2) }}</span>
                <span class="text-muted">/ 座</span>
                <span class="divider">·</span>
                <span>{{ row.seats_total - row.seats_taken }}/{{ row.seats_total }} 空座</span>
              </div>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="类型" width="100">
          <template #default="{ row }">
            <el-tag :type="typeTagType(row.type)" size="small" effect="plain">{{ typeText(row.type) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="出发时间" width="160">
          <template #default="{ row }">{{ formatTime(row.departure_time) }}</template>
        </el-table-column>
        <el-table-column label="发布者" width="150">
          <template #default="{ row }">
            <div class="publisher">
              <div class="publisher-name">{{ row.user_name || `用户#${row.user_id}` }}</div>
              <div class="publisher-phone">{{ maskPhone(row.user_phone) }}</div>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="地区" width="100">
          <template #default="{ row }">
            <span v-if="row.region_id">区#{{ row.region_id }}</span>
            <span v-else class="text-muted">-</span>
          </template>
        </el-table-column>
        <el-table-column label="浏览" width="70" prop="view_count" sortable="custom" />
        <el-table-column label="收藏" width="70" prop="fav_count" sortable="custom" />
        <el-table-column label="预订" width="70" prop="booking_count" />
        <el-table-column label="审核" width="90">
          <template #default="{ row }">
            <el-tag :type="auditTagType(row.audit_status)" size="small">{{ auditText(row.audit_status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="90">
          <template #default="{ row }">
            <el-tag :type="statusTagType(row.status)" size="small" effect="plain">{{ statusText(row.status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="发布时间" width="160" prop="created_at" sortable="custom">
          <template #default="{ row }">{{ formatTime(row.published_at || row.created_at) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="240" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="openDetail(row)">详情</el-button>
            <el-button
              v-if="row.audit_status === 0 || row.audit_status === 2"
              type="success"
              link
              size="small"
              @click="handleAudit(row, 1)"
            >通过</el-button>
            <el-button
              v-if="row.audit_status === 0 || row.audit_status === 1"
              type="danger"
              link
              size="small"
              @click="handleAudit(row, 2)"
            >拒绝</el-button>
            <el-dropdown trigger="click" @command="(cmd) => handleStatusCommand(row, cmd)">
              <el-button type="warning" link size="small">
                更多<el-icon class="el-icon--right"><ArrowDown /></el-icon>
              </el-button>
              <template #dropdown>
                <el-dropdown-menu>
                  <el-dropdown-item :command="1" :disabled="row.status === 1">发布</el-dropdown-item>
                  <el-dropdown-item :command="3" :disabled="row.status === 3">下架</el-dropdown-item>
                  <el-dropdown-item :command="4" :disabled="row.status === 4">设为过期</el-dropdown-item>
                  <el-dropdown-item :command="5" :disabled="row.status === 5">取消</el-dropdown-item>
                  <el-dropdown-item :command="'delete'" divided>删除</el-dropdown-item>
                </el-dropdown-menu>
              </template>
            </el-dropdown>
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
    <el-dialog v-model="detailVisible" title="拼车详情" width="900px" @close="onDetailClose" destroy-on-close>
      <div v-loading="detailLoading">
        <el-tabs v-if="detail" v-model="detailTab">
          <el-tab-pane label="基本信息" name="basic">
            <el-descriptions :column="3" border>
              <el-descriptions-item label="ID">{{ detail.id }}</el-descriptions-item>
              <el-descriptions-item label="出发地">{{ detail.origin }}</el-descriptions-item>
              <el-descriptions-item label="目的地">{{ detail.destination }}</el-descriptions-item>
              <el-descriptions-item label="途经">
                <span v-if="detail.waypoints && detail.waypoints.length">{{ detail.waypoints.join(' → ') }}</span>
                <span v-else class="text-muted">-</span>
              </el-descriptions-item>
              <el-descriptions-item label="拼车类型">
                <el-tag :type="typeTagType(detail.type)" size="small" effect="plain">{{ typeText(detail.type) }}</el-tag>
              </el-descriptions-item>
              <el-descriptions-item label="价格">
                <span class="price">¥{{ Number(detail.price || 0).toFixed(2) }}</span>
                <span class="price-unit">/ {{ detail.price_unit || '座' }}</span>
              </el-descriptions-item>
              <el-descriptions-item label="总座位">{{ detail.seats_total }}</el-descriptions-item>
              <el-descriptions-item label="已预订">{{ detail.seats_taken }}</el-descriptions-item>
              <el-descriptions-item label="剩余空位">{{ detail.seats_total - detail.seats_taken }}</el-descriptions-item>
              <el-descriptions-item label="出发时间">{{ formatTime(detail.departure_time) }}</el-descriptions-item>
              <el-descriptions-item label="集合地点">{{ detail.meeting_point || '-' }}</el-descriptions-item>
              <el-descriptions-item label="车型">{{ detail.vehicle_type || '-' }}</el-descriptions-item>
              <el-descriptions-item label="审核状态">
                <el-tag :type="auditTagType(detail.audit_status)" size="small">{{ auditText(detail.audit_status) }}</el-tag>
              </el-descriptions-item>
              <el-descriptions-item label="拼车状态">
                <el-tag :type="statusTagType(detail.status)" size="small" effect="plain">{{ statusText(detail.status) }}</el-tag>
              </el-descriptions-item>
              <el-descriptions-item label="加急">
                <el-tag v-if="detail.is_urgent" type="danger" size="small">急</el-tag>
                <span v-else>否</span>
              </el-descriptions-item>
              <el-descriptions-item label="联系电话">{{ maskPhone(detail.contact_phone || detail.user_phone) }}</el-descriptions-item>
              <el-descriptions-item label="微信号">{{ detail.contact_wechat || '-' }}</el-descriptions-item>
              <el-descriptions-item v-if="detail.summary" label="摘要" :span="3">{{ detail.summary }}</el-descriptions-item>
              <el-descriptions-item v-if="detail.content" label="详细描述" :span="3">
                <div class="content-box">{{ detail.content }}</div>
              </el-descriptions-item>
              <el-descriptions-item v-if="detail.audit_reason" label="审核原因" :span="3">{{ detail.audit_reason }}</el-descriptions-item>
              <el-descriptions-item label="发布者">{{ detail.user_name || `用户#${detail.user_id}` }}</el-descriptions-item>
              <el-descriptions-item label="发布者ID">{{ detail.user_id }}</el-descriptions-item>
              <el-descriptions-item label="地区ID">{{ detail.region_id }}</el-descriptions-item>
              <el-descriptions-item label="浏览量">{{ detail.view_count }}</el-descriptions-item>
              <el-descriptions-item label="收藏量">{{ detail.fav_count }}</el-descriptions-item>
              <el-descriptions-item label="预订数">{{ detail.booking_count }}</el-descriptions-item>
              <el-descriptions-item label="过期时间">{{ formatTime(detail.expiry_time) }}</el-descriptions-item>
              <el-descriptions-item label="发布时间">{{ formatTime(detail.published_at) }}</el-descriptions-item>
              <el-descriptions-item label="创建时间">{{ formatTime(detail.created_at) }}</el-descriptions-item>
            </el-descriptions>
          </el-tab-pane>
          <el-tab-pane label="行程路线" name="route">
            <el-descriptions :column="2" border>
              <el-descriptions-item label="出发地经纬度">{{ detail.origin_lng }}, {{ detail.origin_lat }}</el-descriptions-item>
              <el-descriptions-item label="目的地经纬度">{{ detail.dest_lng }}, {{ detail.dest_lat }}</el-descriptions-item>
              <el-descriptions-item label="预估距离">{{ detail.estimated_distance ? detail.estimated_distance + ' km' : '-' }}</el-descriptions-item>
              <el-descriptions-item label="预估时长">{{ detail.estimated_duration ? detail.estimated_duration + ' 分钟' : '-' }}</el-descriptions-item>
              <el-descriptions-item label="高速费">{{ detail.has_toll ? '是' : '否' }}</el-descriptions-item>
              <el-descriptions-item label="拼车规则">{{ detail.share_rule || '-' }}</el-descriptions-item>
            </el-descriptions>
          </el-tab-pane>
          <el-tab-pane label="操作日志" name="logs">
            <div class="empty-text">操作日志功能开发中</div>
          </el-tab-pane>
        </el-tabs>
      </div>
      <template #footer>
        <el-button @click="detailVisible = false">关闭</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import request from '@/utils/request'
import {
  Refresh, RefreshLeft, Search, ArrowDown, Check, Bottom, Delete, Plus,
  Van, CirclePlus, Clock, Promotion, Finished, Warning, Right
} from '@element-plus/icons-vue'
import { formatTime } from '@/utils/format'

// ===== 列表 =====
const loading = ref(false)
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)
const list = ref([])
const selection = ref([])
const sortField = ref('created_at')
const sortOrder = ref('descending')

// ===== 统计卡片 =====
const stats = reactive({
  total: 0, todayNew: 0, pendingAudit: 0, published: 0, completed: 0, violation: 0
})

const loadStats = async () => {
  try {
    const res = await request.get('/pinche/statistics/overview')
    const d = res.data || {}
    stats.total = d.total_pinches || d.total || 0
    stats.todayNew = d.today_new || 0
    stats.pendingAudit = d.pending_audit || 0
    stats.published = d.published || 0
    stats.completed = d.completed || 0
    stats.violation = d.violation || 0
  } catch (e) {
    // 接口失败时统计保持默认值 0
  }
}

// ===== 筛选 =====
const filters = reactive({
  keyword: '', type: '', min_seats: undefined, max_seats: undefined,
  min_price: undefined, max_price: undefined,
  audit_status: null, status: null, dateRange: null
})

const onSearch = () => {
  page.value = 1
  loadList()
}

const onReset = () => {
  filters.keyword = ''
  filters.type = ''
  filters.min_seats = undefined
  filters.max_seats = undefined
  filters.min_price = undefined
  filters.max_price = undefined
  filters.audit_status = null
  filters.status = null
  filters.dateRange = null
  page.value = 1
  loadList()
}

const onSortChange = ({ prop, order }) => {
  sortField.value = prop || 'created_at'
  sortOrder.value = order || 'descending'
  loadList()
}

const onSelectionChange = (rows) => {
  selection.value = rows
}

// ===== 状态格式化 =====
const auditText = (s) => ({ 0: '待审核', 1: '已通过', 2: '已拒绝' }[s] || '-')
const auditTagType = (s) => ({ 0: 'warning', 1: 'success', 2: 'danger' }[s] || 'info')
const statusText = (s) => ({ 0: '草稿', 1: '已发布', 2: '已满员', 3: '已下架', 4: '已过期', 5: '已取消' }[s] || '-')
const statusTagType = (s) => ({ 0: 'info', 1: 'success', 2: 'primary', 3: 'warning', 4: 'danger', 5: 'info' }[s] || 'info')
const typeText = (t) => ({ passenger: '人找车', driver: '车找人', cargo: '车找货', cargo_need: '货找车' }[t] || '-')
const typeTagType = (t) => ({ passenger: 'primary', driver: 'success', cargo: 'warning', cargo_need: 'danger' }[t] || 'info')

const maskPhone = (p) => {
  if (!p) return '-'
  const s = String(p)
  if (s.length < 7) return s
  return s.slice(0, 3) + '****' + s.slice(-4)
}

// ===== 查询 =====
const loadList = async () => {
  loading.value = true
  try {
    const params = {
      page: page.value,
      page_size: pageSize.value,
      keyword: filters.keyword.trim() || undefined,
      type: filters.type || undefined,
      min_seats: filters.min_seats || undefined,
      max_seats: filters.max_seats || undefined,
      min_price: filters.min_price || undefined,
      max_price: filters.max_price || undefined,
      audit_status: filters.audit_status === null || filters.audit_status === '' ? undefined : filters.audit_status,
      status: filters.status === null || filters.status === '' ? undefined : filters.status,
      sort: sortField.value,
      order: sortOrder.value
    }
    if (filters.dateRange && filters.dateRange.length === 2) {
      params.start_date = filters.dateRange[0]
      params.end_date = filters.dateRange[1]
    }
    const res = await request.get('/pinche/admin/pinches', { params })
    const data = res.data || {}
    list.value = data.list || []
    total.value = data.total || 0
    if (data.stats) {
      Object.assign(stats, data.stats)
    }
  } catch (e) {
    list.value = []
    total.value = 0
  } finally {
    loading.value = false
  }
}

// ===== 详情弹窗 =====
const detailVisible = ref(false)
const detailLoading = ref(false)
const detail = ref(null)
const detailTab = ref('basic')

const openDetail = async (row) => {
  detail.value = row
  detailTab.value = 'basic'
  detailVisible.value = true
  detailLoading.value = true
  try {
    const res = await request.get(`/pinche/admin/pinches/${row.id}`)
    if (res.data) {
      detail.value = res.data
    }
  } catch (e) {
    // 接口失败保持现状
  } finally {
    detailLoading.value = false
  }
}

const onDetailClose = () => {
  detail.value = null
}

// ===== 审核 =====
const handleAudit = async (row, auditStatus) => {
  const action = auditStatus === 1 ? '通过' : '拒绝'
  try {
    if (auditStatus === 2) {
      const { value } = await ElMessageBox.prompt('请输入拒绝原因（可选）', '拒绝审核', {
        confirmButtonText: '确定拒绝',
        cancelButtonText: '取消',
        inputType: 'textarea',
        inputPlaceholder: '拒绝原因（可不填）'
      })
      await request.put(`/pinche/admin/pinches/${row.id}/audit`, { audit_status: auditStatus, audit_reason: value || '' })
    } else {
      await ElMessageBox.confirm(`确定${action}拼车 "${row.origin} → ${row.destination}" 的审核吗？`, '提示', { type: 'warning' })
      await request.put(`/pinche/admin/pinches/${row.id}/audit`, { audit_status: auditStatus })
    }
    ElMessage.success(`已${action}`)
    await loadList()
    if (detailVisible.value && detail.value?.id === row.id) {
      const res = await request.get(`/pinche/admin/pinches/${row.id}`)
      if (res.data) detail.value = res.data
    }
  } catch (e) {
    // 取消
  }
}

// ===== 状态变更 =====
const handleStatusCommand = async (row, cmd) => {
  try {
    if (cmd === 'delete') {
      await ElMessageBox.confirm(`确定删除拼车 "${row.origin} → ${row.destination}" 吗？删除后不可恢复！`, '危险操作', {
        type: 'error',
        confirmButtonText: '确认删除',
        cancelButtonText: '取消'
      })
      await request.delete(`/pinche/${row.id}`)
      ElMessage.success('已删除')
      await loadList()
      return
    }
    const label = statusText(cmd)
    await ElMessageBox.confirm(`确定将拼车设为「${label}」吗？`, '提示', { type: 'warning' })
    await request.put(`/pinche/admin/pinches/${row.id}/status`, { status: cmd })
    ElMessage.success('状态更新成功')
    await loadList()
    if (detailVisible.value && detail.value?.id === row.id) {
      const res = await request.get(`/pinche/admin/pinches/${row.id}`)
      if (res.data) detail.value = res.data
    }
  } catch (e) {
    // 取消
  }
}

// ===== 批量操作 =====
const onBatchAudit = async (auditStatus) => {
  try {
    await ElMessageBox.confirm(`确认批量审核通过 ${selection.value.length} 个拼车？`, '批量审核', { type: 'warning' })
    await request.post('/pinche/admin/pinches/batch-audit', {
      ids: selection.value.map((r) => r.id),
      audit_status: auditStatus
    })
    ElMessage.success('批量审核完成')
    await loadList()
  } catch (e) {
    // 取消
  }
}

const onBatchStatus = async (status) => {
  try {
    const label = statusText(status)
    await ElMessageBox.confirm(`确认批量将 ${selection.value.length} 个拼车设为「${label}」？`, '批量状态变更', { type: 'warning' })
    await request.post('/pinche/admin/pinches/batch-status', {
      ids: selection.value.map((r) => r.id),
      status
    })
    ElMessage.success('批量操作完成')
    await loadList()
  } catch (e) {
    // 取消
  }
}

const onBatchDelete = async () => {
  try {
    await ElMessageBox.confirm(`确认批量删除 ${selection.value.length} 个拼车？删除后不可恢复！`, '危险操作', {
      type: 'error',
      confirmButtonText: '确认删除',
      cancelButtonText: '取消'
    })
    await request.post('/pinche/admin/pinches/batch-delete', { ids: selection.value.map((r) => r.id) })
    ElMessage.success('批量删除完成')
    await loadList()
  } catch (e) {
    // 取消
  }
}

// ===== 新建 =====
const onCreate = () => {
  ElMessage.info('新建拼车功能开发中，请使用 C 端发布')
}

onMounted(async () => {
  await Promise.all([loadStats()])
  await loadList()
})
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

.filter-form {
  margin-bottom: 12px;
  padding: 12px 16px;
  background: #fafafa;
  border-radius: 4px;
}
.filter-form :deep(.el-form-item) { margin-bottom: 8px; margin-right: 12px; }

.toolbar {
  display: flex; justify-content: space-between; align-items: center;
  flex-wrap: wrap; gap: 8px; margin-bottom: 12px;
}
.toolbar-left { display: flex; gap: 8px; flex-wrap: wrap; }
.toolbar-right { display: flex; gap: 8px; }

.route-cell { display: flex; flex-direction: column; gap: 4px; min-width: 0; }
.route-text {
  font-weight: 500; color: #303133;
  display: flex; align-items: center; gap: 6px;
}
.route-text .el-link { display: inline-flex; align-items: center; gap: 4px; }
.from { color: #303133; }
.to { color: #303133; }
.arrow { color: #909399; }
.route-desc {
  font-size: 12px; color: #909399;
  display: flex; align-items: center; gap: 6px;
}
.price { color: #f56c6c; font-weight: 600; }
.price-unit { color: #909399; font-size: 12px; margin-left: 2px; }
.divider { color: #dcdfe6; margin: 0 2px; }
.publisher { font-size: 13px; }
.publisher-name { color: #303133; }
.publisher-phone { font-size: 12px; color: #909399; margin-top: 2px; }
.text-muted { color: #909399; }

.content-box {
  white-space: pre-wrap; word-break: break-all;
  max-height: 200px; overflow-y: auto;
}
.empty-text { color: #909399; text-align: center; padding: 32px 0; }
.pagination-wrap {
  margin-top: 16px; display: flex; justify-content: flex-end;
}

@media (max-width: 1200px) {
  .filter-form :deep(.el-form-item) { margin-right: 8px; }
}
</style>
