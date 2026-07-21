<template>
  <div class="app-container">
    <div class="page-card">
      <el-form :inline="true" :model="filters" class="filter-form" @submit.prevent>
        <el-form-item label="订单ID">
          <el-input-number v-model="filters.order_id" :controls="false" placeholder="按订单查询" style="width: 160px" />
        </el-form-item>
        <el-form-item label="商品ID">
          <el-input-number v-model="filters.product_id" :controls="false" placeholder="商品ID" style="width: 140px" />
        </el-form-item>
        <el-form-item label="店铺ID">
          <el-input-number v-model="filters.shop_id" :controls="false" placeholder="店铺ID" style="width: 140px" />
        </el-form-item>
        <el-form-item label="评价状态">
          <el-select v-model="filters.has_review" placeholder="全部" clearable style="width: 120px" @change="onSearch">
            <el-option label="已评价" :value="1" />
            <el-option label="未评价" :value="0" />
          </el-select>
        </el-form-item>
        <el-form-item label="退款状态">
          <el-select v-model="filters.refund_status" placeholder="全部" clearable style="width: 120px" @change="onSearch">
            <el-option label="未退款" :value="0" />
            <el-option label="退款中" :value="1" />
            <el-option label="已退款" :value="2" />
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
        <el-table-column prop="order_id" label="订单ID" width="90" />
        <el-table-column label="商品" min-width="220">
          <template #default="{ row }">
            <div class="title-text">{{ row.product_name || `商品#${row.product_id}` }}</div>
            <div class="text-muted text-xs">{{ row.sku_name || row.sku_no || '-' }}</div>
          </template>
        </el-table-column>
        <el-table-column prop="shop_id" label="店铺ID" width="90" />
        <el-table-column label="单价" width="100">
          <template #default="{ row }">¥{{ Number(row.unit_price || 0).toFixed(2) }}</template>
        </el-table-column>
        <el-table-column prop="quantity" label="数量" width="80" />
        <el-table-column label="小计" width="120">
          <template #default="{ row }">¥{{ Number(row.total_price || 0).toFixed(2) }}</template>
        </el-table-column>
        <el-table-column label="实付" width="120">
          <template #default="{ row }">¥{{ Number(row.pay_amount || 0).toFixed(2) }}</template>
        </el-table-column>
        <el-table-column label="评价状态" width="100">
          <template #default="{ row }">
            <el-tag :type="row.has_review ? 'success' : 'info'" size="small">{{ row.has_review ? '已评价' : '未评价' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="退款状态" width="100">
          <template #default="{ row }">
            <el-tag :type="refundTagType(row.refund_status)" size="small">{{ refundText(row.refund_status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="创建时间" width="160" prop="created_at" sortable="custom">
          <template #default="{ row }">{{ formatTime(row.created_at) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="160" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="openDetail(row)">详情</el-button>
            <el-button type="warning" link size="small" @click="openReviewStatus(row)">评价状态</el-button>
          </template>
        </el-table-column>
      </el-table>

      <div class="pagination-wrap">
        <el-pagination v-model:current-page="page" v-model:page-size="pageSize" :total="total" :page-sizes="[10, 20, 50, 100]" layout="total, sizes, prev, pager, next, jumper" background @current-change="loadList" @size-change="loadList" />
      </div>
    </div>

    <!-- 详情弹窗 -->
    <el-dialog v-model="detailVisible" title="订单项详情" width="640px" destroy-on-close>
      <el-descriptions v-if="detail" :column="2" border>
        <el-descriptions-item label="ID">{{ detail.id }}</el-descriptions-item>
        <el-descriptions-item label="订单ID">{{ detail.order_id }}</el-descriptions-item>
        <el-descriptions-item label="商品ID">{{ detail.product_id }}</el-descriptions-item>
        <el-descriptions-item label="SKU ID">{{ detail.sku_id || '-' }}</el-descriptions-item>
        <el-descriptions-item label="商品名" :span="2">{{ detail.product_name || '-' }}</el-descriptions-item>
        <el-descriptions-item label="SKU名" :span="2">{{ detail.sku_name || detail.sku_no || '-' }}</el-descriptions-item>
        <el-descriptions-item label="单价">¥{{ Number(detail.unit_price || 0).toFixed(2) }}</el-descriptions-item>
        <el-descriptions-item label="数量">{{ detail.quantity }}</el-descriptions-item>
        <el-descriptions-item label="小计">¥{{ Number(detail.total_price || 0).toFixed(2) }}</el-descriptions-item>
        <el-descriptions-item label="实付">¥{{ Number(detail.pay_amount || 0).toFixed(2) }}</el-descriptions-item>
        <el-descriptions-item label="评价状态">
          <el-tag :type="detail.has_review ? 'success' : 'info'" size="small">{{ detail.has_review ? '已评价' : '未评价' }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="退款状态">
          <el-tag :type="refundTagType(detail.refund_status)" size="small">{{ refundText(detail.refund_status) }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="创建时间">{{ formatTime(detail.created_at) }}</el-descriptions-item>
        <el-descriptions-item label="更新时间">{{ formatTime(detail.updated_at) }}</el-descriptions-item>
      </el-descriptions>
    </el-dialog>

    <!-- 评价状态更新弹窗 -->
    <el-dialog v-model="reviewStatusVisible" title="更新评价状态" width="420px" destroy-on-close>
      <el-form :model="reviewStatusForm" label-width="100px">
        <el-form-item label="订单项ID">{{ reviewStatusForm.id }}</el-form-item>
        <el-form-item label="是否评价">
          <el-switch v-model="reviewStatusForm.has_review" />
        </el-form-item>
        <el-form-item label="评价ID">
          <el-input-number v-model="reviewStatusForm.review_id" :controls="false" :min="0" style="width: 100%" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="reviewStatusVisible = false">取消</el-button>
        <el-button type="primary" :loading="reviewStatusLoading" @click="onReviewStatusSubmit">确认</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { Refresh, RefreshLeft, Search } from '@element-plus/icons-vue'
import { formatTime } from '@/utils/format'
import { getMallOrderItemList, getMallOrderItemDetail, updateMallOrderItemReviewStatus } from '@/api/mall'

const loading = ref(false)
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)
const list = ref([])
const sortField = ref('created_at')
const sortOrder = ref('descending')

const filters = reactive({ order_id: null, product_id: null, shop_id: null, has_review: null, refund_status: null })

const onSearch = () => { page.value = 1; loadList() }
const onReset = () => {
  filters.order_id = null
  filters.product_id = null
  filters.shop_id = null
  filters.has_review = null
  filters.refund_status = null
  page.value = 1
  loadList()
}
const onSortChange = ({ prop, order }) => {
  sortField.value = prop || 'created_at'
  sortOrder.value = order || 'descending'
  loadList()
}

const refundText = (s) => ({ 0: '未退款', 1: '退款中', 2: '已退款' }[s] || '-')
const refundTagType = (s) => ({ 0: 'info', 1: 'warning', 2: 'success' }[s] || 'info')

const loadList = async () => {
  loading.value = true
  try {
    const params = {
      page: page.value,
      page_size: pageSize.value,
      order_id: filters.order_id || undefined,
      product_id: filters.product_id || undefined,
      shop_id: filters.shop_id || undefined,
      has_review: filters.has_review === null || filters.has_review === '' ? undefined : filters.has_review,
      refund_status: filters.refund_status === null || filters.refund_status === '' ? undefined : filters.refund_status,
      sort: sortField.value,
      order: sortOrder.value
    }
    const res = await getMallOrderItemList(params)
    const data = res.data || {}
    list.value = data.list || []
    total.value = data.total || 0
  } catch (e) {
    ElMessage.error('加载订单项列表失败')
    list.value = []
    total.value = 0
  } finally {
    loading.value = false
  }
}

const detailVisible = ref(false)
const detail = ref(null)

const openDetail = async (row) => {
  try {
    const res = await getMallOrderItemDetail(row.id)
    detail.value = res.data || null
    detailVisible.value = true
  } catch (e) {
    ElMessage.error('加载详情失败')
  }
}

const reviewStatusVisible = ref(false)
const reviewStatusLoading = ref(false)
const reviewStatusForm = reactive({ id: null, has_review: false, review_id: 0 })

const openReviewStatus = (row) => {
  Object.assign(reviewStatusForm, {
    id: row.id,
    has_review: !!row.has_review,
    review_id: row.review_id || 0
  })
  reviewStatusVisible.value = true
}

const onReviewStatusSubmit = async () => {
  try {
    reviewStatusLoading.value = true
    await updateMallOrderItemReviewStatus(reviewStatusForm.id, {
      has_review: reviewStatusForm.has_review,
      review_id: reviewStatusForm.review_id
    })
    ElMessage.success('更新成功')
    reviewStatusVisible.value = false
    await loadList()
  } catch (e) {
    ElMessage.error('更新失败')
  } finally {
    reviewStatusLoading.value = false
  }
}

onMounted(() => loadList())
</script>

<style scoped>
.filter-form { margin-bottom: 12px; padding: 12px 16px; background: #fafafa; border-radius: 4px; }
.filter-form :deep(.el-form-item) { margin-bottom: 8px; margin-right: 12px; }
.toolbar { display: flex; justify-content: space-between; margin-bottom: 12px; }
.pagination-wrap { margin-top: 16px; display: flex; justify-content: flex-end; }
.title-text { font-weight: 500; color: #303133; }
.text-muted { color: #909399; }
.text-xs { font-size: 12px; }
</style>
