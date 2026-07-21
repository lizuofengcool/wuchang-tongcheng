<template>
  <div class="app-container">
    <!-- 购物车统计 -->
    <el-row :gutter="16" class="stat-row">
      <el-col :xs="12" :sm="8" :md="6">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value">{{ stats.total_count }}</div><div class="stat-label">购物车项总数</div></div></el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="6">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value text-primary">{{ stats.selected_count }}</div><div class="stat-label">已选数量</div></div></el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="6">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value text-success">¥{{ Number(stats.total_amount || 0).toFixed(2) }}</div><div class="stat-label">合计金额</div></div></el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="6">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value text-warning">{{ stats.shop_count }}</div><div class="stat-label">涉及店铺数</div></div></el-card>
      </el-col>
    </el-row>

    <div class="page-card">
      <el-form :inline="true" :model="filters" class="filter-form" @submit.prevent>
        <el-form-item label="查询方式">
          <el-radio-group v-model="filters.mode" @change="onModeChange">
            <el-radio-button label="mine">当前用户</el-radio-button>
            <el-radio-button label="by_shop">按店铺</el-radio-button>
            <el-radio-button label="detail">按购物车项ID</el-radio-button>
          </el-radio-group>
        </el-form-item>
        <el-form-item v-if="filters.mode === 'by_shop'" label="店铺ID">
          <el-input-number v-model="filters.shop_id" :controls="false" :min="1" placeholder="店铺ID" style="width: 140px" />
        </el-form-item>
        <el-form-item v-if="filters.mode === 'detail'" label="购物车项ID">
          <el-input-number v-model="filters.cart_id" :controls="false" :min="1" placeholder="ID" style="width: 140px" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" :icon="Search" @click="onSearch">查询</el-button>
          <el-button :icon="RefreshLeft" @click="onReset">重置</el-button>
        </el-form-item>
      </el-form>

      <div class="toolbar">
        <el-button :icon="Refresh" @click="loadList">刷新</el-button>
        <span v-if="summary" class="toolbar-summary">
          合计：¥{{ Number(summary.total_amount || 0).toFixed(2) }} · 共 {{ summary.total_count || 0 }} 项
        </span>
      </div>

      <el-table v-loading="loading" :data="list" border stripe>
        <el-table-column prop="id" label="ID" width="70" fixed="left" />
        <el-table-column label="商品" min-width="200">
          <template #default="{ row }">
            <div class="product-cell">
              <el-image v-if="row.product_image" :src="row.product_image" fit="cover" class="product-thumb" />
              <div class="product-info">
                <div class="product-name">{{ row.product_name || `#${row.product_id}` }}</div>
                <div class="product-sku">{{ row.sku_name || row.sku_specs || '-' }}</div>
              </div>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="店铺ID" width="90" prop="shop_id" />
        <el-table-column label="用户ID" width="90" prop="user_id" />
        <el-table-column label="单价" width="110">
          <template #default="{ row }">¥{{ Number(row.unit_price || 0).toFixed(2) }}</template>
        </el-table-column>
        <el-table-column label="数量" width="90" prop="quantity" />
        <el-table-column label="小计" width="120">
          <template #default="{ row }">¥{{ Number(row.subtotal || (row.unit_price * row.quantity) || 0).toFixed(2) }}</template>
        </el-table-column>
        <el-table-column label="已选" width="80">
          <template #default="{ row }">
            <el-tag :type="row.is_selected ? 'success' : 'info'" size="small">{{ row.is_selected ? '已选' : '未选' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="加入时间" width="160">
          <template #default="{ row }">{{ formatTime(row.created_at) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="160" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="openDetail(row)">详情</el-button>
            <el-button type="danger" link size="small" @click="onDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>

      <div v-if="filters.mode !== 'mine'" class="pagination-wrap">
        <el-pagination v-model:current-page="page" v-model:page-size="pageSize" :total="total" :page-sizes="[10, 20, 50, 100]" layout="total, sizes, prev, pager, next, jumper" background @current-change="loadList" @size-change="loadList" />
      </div>
    </div>

    <!-- 详情弹窗 -->
    <el-dialog v-model="detailVisible" title="购物车项详情" width="640px" destroy-on-close>
      <el-descriptions v-if="detail" :column="2" border>
        <el-descriptions-item label="ID">{{ detail.id }}</el-descriptions-item>
        <el-descriptions-item label="用户ID">{{ detail.user_id }}</el-descriptions-item>
        <el-descriptions-item label="店铺ID">{{ detail.shop_id }}</el-descriptions-item>
        <el-descriptions-item label="商品ID">{{ detail.product_id }}</el-descriptions-item>
        <el-descriptions-item label="SKU ID">{{ detail.sku_id || '-' }}</el-descriptions-item>
        <el-descriptions-item label="数量">{{ detail.quantity }}</el-descriptions-item>
        <el-descriptions-item label="单价">¥{{ Number(detail.unit_price || 0).toFixed(2) }}</el-descriptions-item>
        <el-descriptions-item label="小计">¥{{ Number(detail.subtotal || (detail.unit_price * detail.quantity) || 0).toFixed(2) }}</el-descriptions-item>
        <el-descriptions-item label="商品名" :span="2">{{ detail.product_name || '-' }}</el-descriptions-item>
        <el-descriptions-item label="SKU 名" :span="2">{{ detail.sku_name || detail.sku_specs || '-' }}</el-descriptions-item>
        <el-descriptions-item label="商品图片" :span="2">
          <el-image v-if="detail.product_image" :src="detail.product_image" fit="cover" style="width: 80px; height: 80px" />
          <span v-else>-</span>
        </el-descriptions-item>
        <el-descriptions-item label="是否选中">
          <el-tag :type="detail.is_selected ? 'success' : 'info'" size="small">{{ detail.is_selected ? '已选' : '未选' }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="状态">{{ detail.status === 1 ? '有效' : (detail.status === 0 ? '失效' : '-') }}</el-descriptions-item>
        <el-descriptions-item label="加入时间">{{ formatTime(detail.created_at) }}</el-descriptions-item>
        <el-descriptions-item label="更新时间">{{ formatTime(detail.updated_at) }}</el-descriptions-item>
      </el-descriptions>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Refresh, RefreshLeft, Search } from '@element-plus/icons-vue'
import { formatTime } from '@/utils/format'
import {
  getMallCartList, getMallCartByShop, getMallCartDetail, deleteMallCartItem,
  getMallCartSummary, getMallCartCount, getMallCartSelectedCount
} from '@/api/mall'

const loading = ref(false)
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)
const list = ref([])

const stats = reactive({ total_count: 0, selected_count: 0, total_amount: 0, shop_count: 0 })
const summary = ref(null)

const filters = reactive({ mode: 'mine', shop_id: null, cart_id: null })

const onModeChange = () => { page.value = 1; loadList() }
const onSearch = () => { page.value = 1; loadList() }
const onReset = () => {
  filters.mode = 'mine'
  filters.shop_id = null
  filters.cart_id = null
  page.value = 1
  loadList()
}

const loadList = async () => {
  loading.value = true
  try {
    let data = {}
    if (filters.mode === 'mine') {
      const res = await getMallCartList()
      data = res.data || {}
      list.value = data.list || data || []
      total.value = data.total || list.value.length
    } else if (filters.mode === 'by_shop') {
      if (!filters.shop_id) {
        ElMessage.warning('请输入店铺ID')
        list.value = []
        total.value = 0
        return
      }
      const res = await getMallCartByShop(filters.shop_id)
      data = res.data || {}
      list.value = data.list || data || []
      total.value = data.total || list.value.length
    } else if (filters.mode === 'detail') {
      if (!filters.cart_id) {
        ElMessage.warning('请输入购物车项ID')
        list.value = []
        total.value = 0
        return
      }
      const res = await getMallCartDetail(filters.cart_id)
      const item = res.data || null
      list.value = item ? [item] : []
      total.value = list.value.length
    }
    computeStats()
    await loadSummary()
  } catch (e) {
    ElMessage.error('加载购物车列表失败')
    list.value = []
    total.value = 0
  } finally {
    loading.value = false
  }
}

const computeStats = () => {
  const totalCount = list.value.length
  const selectedCount = list.value.filter((r) => r.is_selected).length
  const totalAmount = list.value.reduce((sum, r) => sum + Number(r.subtotal || (r.unit_price * r.quantity) || 0), 0)
  const shopCount = new Set(list.value.map((r) => r.shop_id).filter(Boolean)).size
  Object.assign(stats, {
    total_count: totalCount,
    selected_count: selectedCount,
    total_amount: totalAmount,
    shop_count: shopCount
  })
}

const loadSummary = async () => {
  try {
    const res = await getMallCartSummary()
    summary.value = res.data || null
  } catch (e) { /* ignore */ }
}

const detailVisible = ref(false)
const detail = ref(null)

const openDetail = async (row) => {
  try {
    const res = await getMallCartDetail(row.id)
    detail.value = res.data || null
    detailVisible.value = true
  } catch (e) {
    ElMessage.error('加载详情失败')
  }
}

const onDelete = async (row) => {
  try {
    await ElMessageBox.confirm(`确认删除购物车项 #${row.id}？`, '提示', { type: 'warning' })
    await deleteMallCartItem(row.id)
    ElMessage.success('删除成功')
    await loadList()
  } catch (e) { /* cancel */ }
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
.toolbar { display: flex; justify-content: space-between; align-items: center; margin-bottom: 12px; }
.toolbar-summary { color: #606266; font-size: 13px; }
.pagination-wrap { margin-top: 16px; display: flex; justify-content: flex-end; }
.product-cell { display: flex; align-items: center; gap: 10px; }
.product-thumb { width: 48px; height: 48px; border-radius: 4px; flex-shrink: 0; }
.product-info { flex: 1; min-width: 0; }
.product-name { font-size: 13px; color: #303133; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.product-sku { font-size: 12px; color: #909399; margin-top: 2px; }
</style>
