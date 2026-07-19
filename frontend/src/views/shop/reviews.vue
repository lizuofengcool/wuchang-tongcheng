<template>
  <div class="app-container">
    <div class="page-card">
      <div class="toolbar">
        <div class="toolbar-left">
          <el-button :icon="Refresh" @click="loadReviews">刷新</el-button>
        </div>
        <div class="toolbar-right">
          <el-input
            v-model.number="shopIdFilter"
            placeholder="店铺ID"
            clearable
            style="width: 120px"
            @keyup.enter="onSearch"
            @clear="onSearch"
          />
          <el-select
            v-model="statusFilter"
            placeholder="评价状态"
            clearable
            style="width: 120px; margin-left: 8px"
            @change="onSearch"
          >
            <el-option label="待审核" :value="0" />
            <el-option label="已通过" :value="1" />
            <el-option label="已拒绝" :value="2" />
          </el-select>
          <el-button type="primary" :icon="Search" style="margin-left: 8px" @click="onSearch">搜索</el-button>
        </div>
      </div>

      <el-table v-loading="loading" :data="list" border stripe>
        <el-table-column prop="id" label="ID" width="70" />
        <el-table-column prop="shop_id" label="店铺ID" width="80" />
        <el-table-column prop="user_id" label="用户ID" width="80" />
        <el-table-column label="评分" width="160">
          <template #default="{ row }">
            <el-rate :model-value="row.rating" disabled size="small" />
          </template>
        </el-table-column>
        <el-table-column prop="content" label="评价内容" min-width="240" show-overflow-tooltip />
        <el-table-column prop="reply" label="商家回复" min-width="180" show-overflow-tooltip>
          <template #default="{ row }">{{ row.reply || '-' }}</template>
        </el-table-column>
        <el-table-column label="状态" width="90">
          <template #default="{ row }">
            <el-tag :type="statusTagType(row.status)" size="small">{{ statusText(row.status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="评价时间" width="160">
          <template #default="{ row }">{{ formatTime(row.created_at) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="170" fixed="right">
          <template #default="{ row }">
            <el-button
              v-if="row.status !== 1"
              type="success"
              link
              size="small"
              @click="handleAudit(row, 1)"
            >通过</el-button>
            <el-button
              v-if="row.status !== 2"
              type="danger"
              link
              size="small"
              @click="handleAudit(row, 2)"
            >拒绝</el-button>
            <span v-if="row.status === 1 && row.status === 2" class="text-muted">-</span>
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
          @current-change="loadReviews"
          @size-change="loadReviews"
        />
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Refresh, Search } from '@element-plus/icons-vue'
import { adminListShopReviews, auditShopReview } from '@/api/shop'
import { formatTime } from '@/utils/format'

const loading = ref(false)
const shopIdFilter = ref(null)
const statusFilter = ref(null)
const page = ref(1)
const pageSize = ref(10)
const total = ref(0)
const list = ref([])

const statusText = (s) => ({ 0: '待审核', 1: '已通过', 2: '已拒绝' }[s] || '-')
const statusTagType = (s) => ({ 0: 'warning', 1: 'success', 2: 'danger' }[s] || 'info')

const onSearch = () => {
  page.value = 1
  loadReviews()
}

const loadReviews = async () => {
  loading.value = true
  try {
    const res = await adminListShopReviews({
      page: page.value,
      page_size: pageSize.value,
      shop_id: shopIdFilter.value || undefined,
      status: statusFilter.value === null || statusFilter.value === '' ? -1 : statusFilter.value
    })
    const data = res.data || {}
    list.value = data.list || []
    total.value = data.total || 0
  } finally {
    loading.value = false
  }
}

const handleAudit = async (row, status) => {
  const action = status === 1 ? '通过' : '拒绝'
  try {
    await ElMessageBox.confirm(`确定${action}该条评价吗？`, '提示', { type: 'warning' })
    await auditShopReview(row.id, status)
    ElMessage.success(`已${action}`)
    await loadReviews()
  } catch (e) {
    // 取消
  }
}

onMounted(() => {
  loadReviews()
})
</script>

<style scoped>
.toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex-wrap: wrap;
  gap: 8px;
}
.toolbar-left {
  display: flex;
  gap: 8px;
}
.toolbar-right {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
}
.text-muted {
  color: #c0c4cc;
}
</style>
