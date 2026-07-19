<template>
  <div class="app-container">
    <div class="page-card">
      <el-form :inline="true" :model="filters" class="filter-form" @submit.prevent>
        <el-form-item label="状态">
          <el-select v-model="filters.status" placeholder="全部" clearable style="width: 140px" @change="onSearch">
            <el-option v-for="(label, val) in statusMap" :key="val" :label="label" :value="Number(val)" />
          </el-select>
        </el-form-item>
        <el-form-item label="关键词">
          <el-input v-model="filters.keyword" placeholder="商品标题" clearable style="width: 200px" @keyup.enter="onSearch" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" :icon="Search" @click="onSearch">搜索</el-button>
          <el-button :icon="RefreshLeft" @click="onReset">重置</el-button>
        </el-form-item>
      </el-form>

      <div class="toolbar"><el-button :icon="Refresh" @click="loadList">刷新</el-button></div>

      <el-table v-loading="loading" :data="list" border stripe>
        <el-table-column prop="id" label="ID" width="70" />
        <el-table-column prop="ershou_id" label="商品ID" width="90" />
        <el-table-column label="起拍价" width="110">
          <template #default="{ row }">¥{{ Number(row.start_price || 0).toFixed(2) }}</template>
        </el-table-column>
        <el-table-column label="当前价" width="120">
          <template #default="{ row }">
            <span class="price">¥{{ Number(row.current_bid_price || 0).toFixed(2) }}</span>
          </template>
        </el-table-column>
        <el-table-column label="加价幅度" width="100">
          <template #default="{ row }">¥{{ Number(row.step_price || 0).toFixed(2) }}</template>
        </el-table-column>
        <el-table-column label="保留价" width="100">
          <template #default="{ row }">¥{{ Number(row.reserve_price || 0).toFixed(2) }}</template>
        </el-table-column>
        <el-table-column label="保证金" width="100">
          <template #default="{ row }">¥{{ Number(row.bond_amount || 0).toFixed(2) }}</template>
        </el-table-column>
        <el-table-column prop="bid_count" label="出价数" width="80" />
        <el-table-column prop="watcher_count" label="围观" width="80" />
        <el-table-column label="开始时间" width="160">
          <template #default="{ row }">{{ formatTime(row.start_time) }}</template>
        </el-table-column>
        <el-table-column label="截拍时间" width="160">
          <template #default="{ row }">
            <span :class="{ 'text-danger': isUrgent(row.end_time) }">{{ formatTime(row.end_time) }}</span>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="statusTagType(row.status)" size="small">{{ row.status_text || statusMap[row.status] || '-' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="成交价" width="110">
          <template #default="{ row }">
            <span v-if="row.winner_price">¥{{ Number(row.winner_price).toFixed(2) }}</span>
            <span v-else class="text-muted">-</span>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="180" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="openBids(row)">出价记录</el-button>
            <el-button v-if="row.status === 0 || row.status === 1" type="danger" link size="small" @click="onEnd(row)">手动截拍</el-button>
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

    <!-- 出价记录弹窗 -->
    <el-dialog v-model="bidsVisible" title="出价记录" width="700px">
      <el-table v-loading="bidsLoading" :data="bids" border size="small">
        <el-table-column prop="id" label="ID" width="70" />
        <el-table-column prop="user_id" label="用户ID" width="100" />
        <el-table-column label="出价" width="120">
          <template #default="{ row }">¥{{ Number(row.bid_price || 0).toFixed(2) }}</template>
        </el-table-column>
        <el-table-column label="出价时间" width="180">
          <template #default="{ row }">{{ formatTime(row.created_at) }}</template>
        </el-table-column>
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag v-if="row.is_winning" type="success" size="small">领先</el-tag>
            <el-tag v-else type="info" size="small">出局</el-tag>
          </template>
        </el-table-column>
      </el-table>
      <template #footer><el-button @click="bidsVisible = false">关闭</el-button></template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Refresh, RefreshLeft, Search } from '@element-plus/icons-vue'
import { listErshouAuctions, endErshouAuction } from '@/api/ershou'
import { formatTime } from '@/utils/format'

const loading = ref(false)
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)
const list = ref([])

const filters = reactive({ status: null, keyword: '' })

const statusMap = {
  0: '待开始', 1: '进行中', 2: '已结束', 3: '已成交', 4: '流拍'
}
const statusTagType = (s) => ({
  0: 'info', 1: 'success', 2: 'warning', 3: 'primary', 4: 'danger'
}[s] || 'info')

const isUrgent = (endTime) => {
  if (!endTime) return false
  const diff = new Date(endTime).getTime() - Date.now()
  return diff > 0 && diff < 60 * 60 * 1000
}

const onSearch = () => { page.value = 1; loadList() }
const onReset = () => { filters.status = null; filters.keyword = ''; page.value = 1; loadList() }

const loadList = async () => {
  loading.value = true
  try {
    const res = await listErshouAuctions({
      page: page.value,
      page_size: pageSize.value,
      status: filters.status === null || filters.status === '' ? undefined : filters.status,
      keyword: filters.keyword || undefined
    })
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

// ===== 出价记录 =====
const bidsVisible = ref(false)
const bidsLoading = ref(false)
const bids = ref([])

const openBids = async (row) => {
  bids.value = []
  bidsVisible.value = true
  bidsLoading.value = true
  try {
    // 出价记录接口若未提供，从拍卖详情中获取
    const { getErshouAuctionByErshouId } = await import('@/api/ershou')
    const res = await getErshouAuctionByErshouId(row.ershou_id)
    bids.value = res.data?.bids || []
  } catch (e) {
    bids.value = []
  } finally {
    bidsLoading.value = false
  }
}

const onEnd = async (row) => {
  try {
    await ElMessageBox.confirm('确认手动截拍？截拍后将不能继续出价。', '提示', { type: 'warning' })
    await endErshouAuction(row.ershou_id)
    ElMessage.success('已截拍')
    await loadList()
  } catch (e) { /* cancel */ }
}

onMounted(() => loadList())
</script>

<style scoped>
.filter-form { margin-bottom: 12px; padding: 12px 16px; background: #fafafa; border-radius: 4px; }
.filter-form :deep(.el-form-item) { margin-bottom: 8px; margin-right: 12px; }
.toolbar { display: flex; justify-content: flex-start; margin-bottom: 12px; }
.price { color: #f56c6c; font-weight: 600; }
.text-muted { color: #909399; }
.text-danger { color: #f56c6c; font-weight: 600; }
.pagination-wrap { margin-top: 16px; display: flex; justify-content: flex-end; }
</style>
