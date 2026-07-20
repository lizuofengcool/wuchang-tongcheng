<template>
  <div class="app-container">
    <!-- 顶部统计卡片 -->
    <el-row :gutter="16" class="stat-row">
      <el-col :xs="12" :sm="8" :md="4">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value">{{ stats.total }}</div><div class="stat-label">订阅总数</div></div></el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="4">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value text-success">{{ stats.active }}</div><div class="stat-label">生效中</div></div></el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="4">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value text-warning">{{ stats.expiringSoon }}</div><div class="stat-label">即将到期</div></div></el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="4">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value text-primary">¥{{ stats.totalRevenue }}</div><div class="stat-label">总收入</div></div></el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="4">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value">{{ stats.todayNew }}</div><div class="stat-label">今日新订</div></div></el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="4">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value text-danger">{{ stats.refunded }}</div><div class="stat-label">已退款</div></div></el-card>
      </el-col>
    </el-row>

    <div class="page-card">
      <el-form :inline="true" :model="filters" class="filter-form" @submit.prevent>
        <el-form-item label="关键词">
          <el-input v-model="filters.keyword" placeholder="会员ID/昵称/订单号" clearable style="width: 220px" :prefix-icon="Search" @keyup.enter="onSearch" />
        </el-form-item>
        <el-form-item label="等级">
          <el-select v-model="filters.level_id" placeholder="全部" clearable style="width: 160px" @change="onSearch">
            <el-option v-for="l in levels" :key="l.id" :label="l.level_name" :value="l.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="filters.status" placeholder="全部" clearable style="width: 130px" @change="onSearch">
            <el-option label="待支付" value="pending" />
            <el-option label="已支付" value="active" />
            <el-option label="已取消" value="cancelled" />
            <el-option label="已退款" value="refunded" />
            <el-option label="已过期" value="expired" />
          </el-select>
        </el-form-item>
        <el-form-item label="自动续费">
          <el-select v-model="filters.auto_renew" placeholder="全部" clearable style="width: 120px" @change="onSearch">
            <el-option label="是" :value="1" />
            <el-option label="否" :value="0" />
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

      <div class="toolbar"><el-button :icon="Refresh" @click="loadList">刷新</el-button></div>

      <el-table v-loading="loading" :data="list" border stripe @sort-change="onSortChange">
        <el-table-column prop="id" label="ID" width="70" sortable="custom" />
        <el-table-column prop="order_no" label="订单号" width="180" fixed="left" />
        <el-table-column label="会员" min-width="160">
          <template #default="{ row }">
            <div class="user-cell">
              <el-image v-if="row.user_avatar" :src="row.user_avatar" fit="cover" class="user-avatar" />
              <div class="user-info">
                <div class="user-name">{{ row.user_name || `#${row.user_id}` }}</div>
                <div class="user-meta">ID: {{ row.user_id }}</div>
              </div>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="等级" width="120">
          <template #default="{ row }">
            <el-tag type="warning" size="small">{{ row.level_name || `#${row.level_id}` }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="金额" width="100">
          <template #default="{ row }">
            <span class="price">¥{{ Number(row.amount || 0).toFixed(2) }}</span>
          </template>
        </el-table-column>
        <el-table-column label="支付方式" width="100">
          <template #default="{ row }">{{ paymentText(row.payment_method) }}</template>
        </el-table-column>
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="statusTagType(row.status)" size="small">{{ statusText(row.status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="自动续费" width="90">
          <template #default="{ row }">
            <el-tag v-if="row.auto_renew" type="success" size="small">是</el-tag>
            <span v-else class="text-muted">否</span>
          </template>
        </el-table-column>
        <el-table-column label="有效期" width="320">
          <template #default="{ row }">
            <div>{{ formatTime(row.start_at) }} 至 {{ formatTime(row.end_at) }}</div>
          </template>
        </el-table-column>
        <el-table-column label="订阅时间" width="160" prop="created_at" sortable="custom">
          <template #default="{ row }">{{ formatTime(row.created_at) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="openDetail(row)">详情</el-button>
            <el-button
              v-if="row.status === 'pending'"
              type="success"
              link
              size="small"
              @click="onMarkPaid(row)"
            >标记已付</el-button>
            <el-button
              v-if="row.status === 'active'"
              type="warning"
              link
              size="small"
              @click="onToggleAutoRenew(row)"
            >{{ row.auto_renew ? '关闭续费' : '开启续费' }}</el-button>
            <el-button
              v-if="row.status === 'active' || row.status === 'pending'"
              type="danger"
              link
              size="small"
              @click="onRefund(row)"
            >退款</el-button>
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
    <el-dialog v-model="detailVisible" title="订阅详情" width="700px">
      <el-descriptions v-if="detail" :column="2" border>
        <el-descriptions-item label="ID">{{ detail.id }}</el-descriptions-item>
        <el-descriptions-item label="订单号">{{ detail.order_no }}</el-descriptions-item>
        <el-descriptions-item label="会员">{{ detail.user_name || `#${detail.user_id}` }}</el-descriptions-item>
        <el-descriptions-item label="用户ID">{{ detail.user_id }}</el-descriptions-item>
        <el-descriptions-item label="等级">{{ detail.level_name || `#${detail.level_id}` }}</el-descriptions-item>
        <el-descriptions-item label="等级ID">{{ detail.level_id }}</el-descriptions-item>
        <el-descriptions-item label="金额">¥{{ Number(detail.amount || 0).toFixed(2) }}</el-descriptions-item>
        <el-descriptions-item label="支付方式">{{ paymentText(detail.payment_method) }}</el-descriptions-item>
        <el-descriptions-item label="状态">
          <el-tag :type="statusTagType(detail.status)" size="small">{{ statusText(detail.status) }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="自动续费">
          <el-tag v-if="detail.auto_renew" type="success" size="small">是</el-tag>
          <span v-else>否</span>
        </el-descriptions-item>
        <el-descriptions-item label="开始时间">{{ formatTime(detail.start_at) }}</el-descriptions-item>
        <el-descriptions-item label="结束时间">{{ formatTime(detail.end_at) }}</el-descriptions-item>
        <el-descriptions-item v-if="detail.refund_amount > 0" label="退款金额">¥{{ Number(detail.refund_amount).toFixed(2) }}</el-descriptions-item>
        <el-descriptions-item v-if="detail.refunded_at" label="退款时间">{{ formatTime(detail.refunded_at) }}</el-descriptions-item>
        <el-descriptions-item v-if="detail.cancelled_at" label="取消时间">{{ formatTime(detail.cancelled_at) }}</el-descriptions-item>
        <el-descriptions-item v-if="detail.paid_at" label="支付时间">{{ formatTime(detail.paid_at) }}</el-descriptions-item>
        <el-descriptions-item label="创建时间">{{ formatTime(detail.created_at) }}</el-descriptions-item>
        <el-descriptions-item label="更新时间">{{ formatTime(detail.updated_at) }}</el-descriptions-item>
      </el-descriptions>
      <template #footer><el-button @click="detailVisible = false">关闭</el-button></template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Refresh, RefreshLeft, Search } from '@element-plus/icons-vue'
import request from '@/utils/request'
import { formatTime } from '@/utils/format'

const loading = ref(false)
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)
const list = ref([])
const sortField = ref('created_at')
const sortOrder = ref('descending')

const filters = reactive({
  keyword: '', level_id: null, status: '', auto_renew: null, dateRange: null
})

const stats = reactive({ total: 0, active: 0, expiringSoon: 0, totalRevenue: '0.00', todayNew: 0, refunded: 0 })

const levels = ref([])

const statusText = (s) => ({ pending: '待支付', active: '已支付', cancelled: '已取消', refunded: '已退款', expired: '已过期' }[s] || '-')
const statusTagType = (s) => ({ pending: 'warning', active: 'success', cancelled: 'info', refunded: 'danger', expired: 'info' }[s] || 'info')
const paymentText = (p) => ({ wechat: '微信', alipay: '支付宝', apple: 'Apple Pay', balance: '余额' }[p] || p || '-')

const onSearch = () => { page.value = 1; loadList() }
const onReset = () => {
  filters.keyword = ''; filters.level_id = null; filters.status = ''
  filters.auto_renew = null; filters.dateRange = null
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
      level_id: filters.level_id || undefined,
      status: filters.status || undefined,
      auto_renew: filters.auto_renew === null || filters.auto_renew === '' ? undefined : filters.auto_renew,
      sort: sortField.value,
      order: sortOrder.value
    }
    if (filters.dateRange && filters.dateRange.length === 2) {
      params.start_date = filters.dateRange[0]
      params.end_date = filters.dateRange[1]
    }
    const res = await request.get('/love/admin/memberships', { params })
    const data = res.data || {}
    list.value = data.list || []
    total.value = data.total || 0
    calcStats()
  } catch (e) {
    list.value = []
    total.value = 0
  } finally {
    loading.value = false
  }
}

const calcStats = () => {
  const today = new Date().toISOString().slice(0, 10)
  const now = Date.now()
  const threeDaysLater = now + 3 * 24 * 3600 * 1000
  stats.total = list.value.length
  stats.active = list.value.filter((r) => r.status === 'active').length
  stats.expiringSoon = list.value.filter((r) => r.status === 'active' && r.end_at && new Date(r.end_at).getTime() < threeDaysLater && new Date(r.end_at).getTime() > now).length
  stats.todayNew = list.value.filter((r) => (r.created_at || '').startsWith(today)).length
  stats.refunded = list.value.filter((r) => r.status === 'refunded').length
  const totalRevenue = list.value.filter((r) => r.status === 'active' || r.status === 'expired').reduce((s, r) => s + Number(r.amount || 0), 0)
  stats.totalRevenue = totalRevenue.toFixed(2)
}

const loadLevels = async () => {
  try {
    const res = await request.get('/love/admin/member-levels', { params: { page: 1, page_size: 100 } })
    const data = res.data || {}
    levels.value = data.list || data || []
  } catch (e) {
    levels.value = []
  }
}

const detailVisible = ref(false)
const detail = ref(null)
const openDetail = (row) => { detail.value = row; detailVisible.value = true }

const onMarkPaid = async (row) => {
  try {
    await ElMessageBox.confirm(`确认将订单 "${row.order_no}" 标记为已支付？`, '提示', { type: 'warning' })
    await request.post(`/love/memberships/${row.id}/mark-paid`)
    ElMessage.success('已标记已支付')
    await loadList()
  } catch (e) { /* cancel */ }
}

const onToggleAutoRenew = async (row) => {
  try {
    await ElMessageBox.confirm(`确认${row.auto_renew ? '关闭' : '开启'}自动续费？`, '提示', { type: 'warning' })
    await request.put(`/love/admin/memberships/${row.id}/auto-renew`, { auto_renew: !row.auto_renew })
    ElMessage.success('操作成功')
    await loadList()
  } catch (e) { /* cancel */ }
}

const onRefund = async (row) => {
  try {
    const { value } = await ElMessageBox.prompt('请输入退款原因', '退款', {
      inputType: 'textarea',
      inputPlaceholder: '退款原因',
      confirmButtonText: '确认退款',
      cancelButtonText: '取消'
    })
    await request.post(`/love/admin/memberships/${row.id}/refund`, { reason: value || '' })
    ElMessage.success('退款成功')
    await loadList()
  } catch (e) { /* cancel */ }
}

onMounted(async () => {
  await loadLevels()
  await loadList()
})
</script>

<style scoped>
.stat-row { margin-bottom: 16px; }
.stat-item { text-align: center; padding: 16px; }
.stat-value { font-size: 22px; font-weight: 600; color: #303133; }
.text-success { color: #67c23a; }
.text-warning { color: #e6a23c; }
.text-primary { color: #409eff; }
.text-danger { color: #f56c6c; }
.filter-form { margin-bottom: 12px; padding: 12px 16px; background: #fafafa; border-radius: 4px; }
.filter-form :deep(.el-form-item) { margin-bottom: 8px; margin-right: 12px; }
.toolbar { display: flex; justify-content: flex-start; margin-bottom: 12px; }
.text-muted { color: #909399; }
.price { color: #f56c6c; font-weight: 600; }
.user-cell { display: flex; align-items: center; gap: 8px; }
.user-avatar { width: 36px; height: 36px; border-radius: 50%; border: 1px solid #ebeef5; }
.user-info { display: flex; flex-direction: column; }
.user-name { color: #303133; font-size: 13px; }
.user-meta { color: #909399; font-size: 12px; margin-top: 2px; }
.pagination-wrap { margin-top: 16px; display: flex; justify-content: flex-end; }
</style>
