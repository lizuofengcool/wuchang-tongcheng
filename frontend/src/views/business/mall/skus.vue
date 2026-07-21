<template>
  <div class="app-container">
    <div class="page-card">
      <el-form :inline="true" :model="filters" class="filter-form" @submit.prevent>
        <el-form-item label="商品ID">
          <el-input-number v-model="filters.product_id" :controls="false" placeholder="按商品查询" style="width: 160px" />
        </el-form-item>
        <el-form-item label="店铺ID">
          <el-input-number v-model="filters.shop_id" :controls="false" placeholder="按店铺查询" style="width: 160px" />
        </el-form-item>
        <el-form-item label="关键词">
          <el-input v-model="filters.keyword" placeholder="SKU名/规格" clearable style="width: 200px" :prefix-icon="Search" @keyup.enter="onSearch" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" :icon="Search" @click="onSearch">搜索</el-button>
          <el-button :icon="RefreshLeft" @click="onReset">重置</el-button>
        </el-form-item>
      </el-form>

      <div class="toolbar">
        <el-button :icon="Refresh" @click="loadList">刷新</el-button>
        <div class="text-muted text-xs">提示：按商品查询时调用 SKU 列表接口；按店铺查询时调用分页接口</div>
      </div>

      <el-table v-loading="loading" :data="list" border stripe>
        <el-table-column prop="id" label="ID" width="70" fixed="left" />
        <el-table-column label="SKU名" min-width="180">
          <template #default="{ row }">
            <div>{{ row.sku_name || row.name || '-' }}</div>
            <div class="text-muted text-xs">{{ row.sku_no || '-' }}</div>
          </template>
        </el-table-column>
        <el-table-column label="规格" min-width="160">
          <template #default="{ row }">
            <el-tag v-for="(v, k) in row.specs" :key="k" size="small" class="spec-tag">{{ k }}:{{ v }}</el-tag>
            <span v-if="!row.specs || !Object.keys(row.specs).length" class="text-muted">-</span>
          </template>
        </el-table-column>
        <el-table-column label="商品ID" width="100" prop="product_id" />
        <el-table-column label="店铺ID" width="100" prop="shop_id" />
        <el-table-column label="价格" width="120">
          <template #default="{ row }">¥{{ Number(row.price || 0).toFixed(2) }}</template>
        </el-table-column>
        <el-table-column label="原价" width="100">
          <template #default="{ row }">¥{{ Number(row.original_price || 0).toFixed(2) }}</template>
        </el-table-column>
        <el-table-column label="库存" width="100">
          <template #default="{ row }">
            <span :class="{ 'text-danger': row.stock <= 0 }">{{ row.stock || 0 }}</span>
          </template>
        </el-table-column>
        <el-table-column label="销量" width="80" prop="sale_count" />
        <el-table-column label="状态" width="90">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'info'" size="small">{{ row.status === 1 ? '启用' : '禁用' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="创建时间" width="160">
          <template #default="{ row }">{{ formatTime(row.created_at) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="160" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="openDetail(row)">详情</el-button>
            <el-button type="warning" link size="small" @click="openStock(row)">调库存</el-button>
          </template>
        </el-table-column>
      </el-table>

      <div v-if="filters.shop_id" class="pagination-wrap">
        <el-pagination v-model:current-page="page" v-model:page-size="pageSize" :total="total" :page-sizes="[10, 20, 50, 100]" layout="total, sizes, prev, pager, next, jumper" background @current-change="loadList" @size-change="loadList" />
      </div>
    </div>

    <!-- 详情弹窗 -->
    <el-dialog v-model="detailVisible" title="SKU 详情" width="640px" destroy-on-close>
      <el-descriptions v-if="detail" :column="2" border>
        <el-descriptions-item label="ID">{{ detail.id }}</el-descriptions-item>
        <el-descriptions-item label="SKU名">{{ detail.sku_name || detail.name || '-' }}</el-descriptions-item>
        <el-descriptions-item label="SKU编号">{{ detail.sku_no || '-' }}</el-descriptions-item>
        <el-descriptions-item label="商品ID">{{ detail.product_id }}</el-descriptions-item>
        <el-descriptions-item label="店铺ID">{{ detail.shop_id }}</el-descriptions-item>
        <el-descriptions-item label="价格">¥{{ Number(detail.price || 0).toFixed(2) }}</el-descriptions-item>
        <el-descriptions-item label="原价">¥{{ Number(detail.original_price || 0).toFixed(2) }}</el-descriptions-item>
        <el-descriptions-item label="库存">{{ detail.stock || 0 }}</el-descriptions-item>
        <el-descriptions-item label="销量">{{ detail.sale_count || 0 }}</el-descriptions-item>
        <el-descriptions-item label="限购">{{ detail.limit_per_user || '-' }}</el-descriptions-item>
        <el-descriptions-item label="状态">
          <el-tag :type="detail.status === 1 ? 'success' : 'info'" size="small">{{ detail.status === 1 ? '启用' : '禁用' }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="创建时间">{{ formatTime(detail.created_at) }}</el-descriptions-item>
        <el-descriptions-item label="规格" :span="2">
          <el-tag v-for="(v, k) in detail.specs" :key="k" size="small" class="spec-tag">{{ k }}:{{ v }}</el-tag>
          <span v-if="!detail.specs || !Object.keys(detail.specs).length" class="text-muted">-</span>
        </el-descriptions-item>
      </el-descriptions>
    </el-dialog>

    <!-- 调库存弹窗 -->
    <el-dialog v-model="stockVisible" title="调整库存" width="420px" destroy-on-close>
      <el-form :model="stockForm" label-width="100px">
        <el-form-item label="SKU">{{ stockForm.sku_name || stockForm.id }}</el-form-item>
        <el-form-item label="当前库存">{{ stockForm.old_stock }}</el-form-item>
        <el-form-item label="新库存">
          <el-input-number v-model="stockForm.stock" :min="0" :max="999999" style="width: 100%" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="stockVisible = false">取消</el-button>
        <el-button type="primary" :loading="stockLoading" @click="onStockSubmit">确认</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { Refresh, RefreshLeft, Search } from '@element-plus/icons-vue'
import { formatTime } from '@/utils/format'
import { getMallSkusByProduct, getMallSkusByShop, getMallSkuDetail, updateMallSkuStock } from '@/api/mall'

const loading = ref(false)
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)
const list = ref([])

const filters = reactive({ product_id: null, shop_id: null, keyword: '' })

const onSearch = () => { page.value = 1; loadList() }
const onReset = () => {
  filters.product_id = null
  filters.shop_id = null
  filters.keyword = ''
  page.value = 1
  loadList()
}

const loadList = async () => {
  if (!filters.product_id && !filters.shop_id) {
    ElMessage.info('请输入商品ID或店铺ID进行查询')
    list.value = []
    total.value = 0
    return
  }
  loading.value = true
  try {
    if (filters.product_id) {
      const res = await getMallSkusByProduct(filters.product_id)
      const data = res.data
      list.value = Array.isArray(data) ? data : (data?.list || [])
      total.value = list.value.length
    } else {
      const params = { page: page.value, page_size: pageSize.value, keyword: filters.keyword || undefined }
      const res = await getMallSkusByShop(filters.shop_id, params)
      const data = res.data || {}
      list.value = data.list || []
      total.value = data.total || 0
    }
    if (filters.keyword) {
      const kw = filters.keyword.toLowerCase()
      list.value = list.value.filter((r) => (r.sku_name || r.name || '').toLowerCase().includes(kw) || (r.sku_no || '').toLowerCase().includes(kw))
    }
  } catch (e) {
    ElMessage.error('加载 SKU 列表失败')
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
    const res = await getMallSkuDetail(row.id)
    detail.value = res.data || null
    detailVisible.value = true
  } catch (e) {
    ElMessage.error('加载详情失败')
  }
}

const stockVisible = ref(false)
const stockLoading = ref(false)
const stockForm = reactive({ id: null, sku_name: '', old_stock: 0, stock: 0 })

const openStock = (row) => {
  Object.assign(stockForm, {
    id: row.id,
    sku_name: row.sku_name || row.name || '',
    old_stock: row.stock || 0,
    stock: row.stock || 0
  })
  stockVisible.value = true
}

const onStockSubmit = async () => {
  try {
    stockLoading.value = true
    await updateMallSkuStock(stockForm.id, { stock: stockForm.stock })
    ElMessage.success('库存已更新')
    stockVisible.value = false
    await loadList()
  } catch (e) {
    ElMessage.error('更新失败')
  } finally {
    stockLoading.value = false
  }
}

onMounted(() => loadList())
</script>

<style scoped>
.filter-form { margin-bottom: 12px; padding: 12px 16px; background: #fafafa; border-radius: 4px; }
.filter-form :deep(.el-form-item) { margin-bottom: 8px; margin-right: 12px; }
.toolbar { display: flex; justify-content: space-between; align-items: center; margin-bottom: 12px; flex-wrap: wrap; gap: 8px; }
.pagination-wrap { margin-top: 16px; display: flex; justify-content: flex-end; }
.text-muted { color: #909399; }
.text-xs { font-size: 12px; }
.text-danger { color: #f56c6c; }
.spec-tag { margin-right: 4px; margin-bottom: 2px; }
</style>
