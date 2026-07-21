<template>
  <div class="app-container">
    <div class="page-card">
      <el-form :inline="true" :model="filters" class="filter-form" @submit.prevent>
        <el-form-item label="关键词">
          <el-input v-model="filters.keyword" placeholder="商品名/编号" clearable style="width: 200px" :prefix-icon="Search" @keyup.enter="onSearch" />
        </el-form-item>
        <el-form-item label="店铺ID">
          <el-input-number v-model="filters.shop_id" :controls="false" placeholder="店铺ID" style="width: 140px" />
        </el-form-item>
        <el-form-item label="分类ID">
          <el-input-number v-model="filters.category_id" :controls="false" placeholder="分类ID" style="width: 140px" />
        </el-form-item>
        <el-form-item label="审核状态">
          <el-select v-model="filters.audit_status" placeholder="全部" clearable style="width: 120px" @change="onSearch">
            <el-option label="待审核" :value="0" />
            <el-option label="已通过" :value="1" />
            <el-option label="已拒绝" :value="2" />
          </el-select>
        </el-form-item>
        <el-form-item label="商品状态">
          <el-select v-model="filters.status" placeholder="全部" clearable style="width: 120px" @change="onSearch">
            <el-option label="下架" :value="0" />
            <el-option label="上架" :value="1" />
            <el-option label="草稿" :value="2" />
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
        <el-table-column label="主图" width="70">
          <template #default="{ row }">
            <el-image v-if="row.main_image" :src="row.main_image" fit="cover" class="cover-thumb" :preview-src-list="[row.main_image]" preview-teleported />
            <div v-else class="cover-thumb cover-empty">无</div>
          </template>
        </el-table-column>
        <el-table-column label="商品名/编号" min-width="220">
          <template #default="{ row }">
            <div class="title-text">{{ row.title }}</div>
            <div class="text-muted text-xs">{{ row.product_no || '-' }}</div>
          </template>
        </el-table-column>
        <el-table-column label="店铺" width="140">
          <template #default="{ row }">
            <div>{{ row.shop_name || `#${row.shop_id}` }}</div>
          </template>
        </el-table-column>
        <el-table-column label="分类" width="100">
          <template #default="{ row }">{{ row.category_name || row.category_id || '-' }}</template>
        </el-table-column>
        <el-table-column label="价格" width="120">
          <template #default="{ row }">
            <div>¥{{ Number(row.price || 0).toFixed(2) }}</div>
            <div v-if="row.original_price" class="text-muted text-xs line-through">¥{{ Number(row.original_price).toFixed(2) }}</div>
          </template>
        </el-table-column>
        <el-table-column label="库存" width="80" prop="stock" />
        <el-table-column label="销量" width="80" prop="sale_count" />
        <el-table-column label="浏览" width="80" prop="view_count" />
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
        <el-table-column label="创建时间" width="160" prop="created_at" sortable="custom">
          <template #default="{ row }">{{ formatTime(row.created_at) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="240" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="openDetail(row)">详情</el-button>
            <el-button v-if="row.audit_status === 0 || row.audit_status === 2" type="success" link size="small" @click="handleAudit(row, 1)">通过</el-button>
            <el-button v-if="row.audit_status === 0 || row.audit_status === 1" type="danger" link size="small" @click="handleAudit(row, 2)">拒绝</el-button>
            <el-button type="warning" link size="small" @click="openPromotion(row)">推广</el-button>
          </template>
        </el-table-column>
      </el-table>

      <div class="pagination-wrap">
        <el-pagination v-model:current-page="page" v-model:page-size="pageSize" :total="total" :page-sizes="[10, 20, 50, 100]" layout="total, sizes, prev, pager, next, jumper" background @current-change="loadList" @size-change="loadList" />
      </div>
    </div>

    <!-- 详情弹窗 -->
    <el-dialog v-model="detailVisible" title="商品详情" width="780px" destroy-on-close>
      <el-descriptions v-if="detail" :column="2" border>
        <el-descriptions-item label="ID">{{ detail.id }}</el-descriptions-item>
        <el-descriptions-item label="商品名">{{ detail.title }}</el-descriptions-item>
        <el-descriptions-item label="编号">{{ detail.product_no || '-' }}</el-descriptions-item>
        <el-descriptions-item label="店铺">{{ detail.shop_name || `#${detail.shop_id}` }}</el-descriptions-item>
        <el-descriptions-item label="分类">{{ detail.category_name || detail.category_id || '-' }}</el-descriptions-item>
        <el-descriptions-item label="品牌">{{ detail.brand || '-' }}</el-descriptions-item>
        <el-descriptions-item label="价格">¥{{ Number(detail.price || 0).toFixed(2) }}</el-descriptions-item>
        <el-descriptions-item label="原价">¥{{ Number(detail.original_price || 0).toFixed(2) }}</el-descriptions-item>
        <el-descriptions-item label="库存">{{ detail.stock || 0 }}</el-descriptions-item>
        <el-descriptions-item label="销量">{{ detail.sale_count || 0 }}</el-descriptions-item>
        <el-descriptions-item label="浏览">{{ detail.view_count || 0 }}</el-descriptions-item>
        <el-descriptions-item label="评分">{{ Number(detail.rating || 0).toFixed(1) }}</el-descriptions-item>
        <el-descriptions-item label="审核状态">
          <el-tag :type="auditTagType(detail.audit_status)" size="small">{{ auditText(detail.audit_status) }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="商品状态">
          <el-tag :type="statusTagType(detail.status)" size="small" effect="plain">{{ statusText(detail.status) }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="简介" :span="2">{{ detail.summary || '-' }}</el-descriptions-item>
        <el-descriptions-item label="详细描述" :span="2">
          <div class="content-box">{{ detail.description || '-' }}</div>
        </el-descriptions-item>
        <el-descriptions-item v-if="detail.audit_reason" label="审核原因" :span="2">{{ detail.audit_reason }}</el-descriptions-item>
        <el-descriptions-item label="创建时间">{{ formatTime(detail.created_at) }}</el-descriptions-item>
        <el-descriptions-item label="更新时间">{{ formatTime(detail.updated_at) }}</el-descriptions-item>
      </el-descriptions>
      <div v-if="detail && detail.images && detail.images.length" class="images-grid">
        <el-image v-for="(img, idx) in detail.images" :key="idx" :src="img" fit="cover" class="image-item" :preview-src-list="detail.images" :initial-index="idx" preview-teleported />
      </div>
    </el-dialog>

    <!-- 推广弹窗 -->
    <el-dialog v-model="promotionVisible" title="推广配置" width="520px" destroy-on-close>
      <el-form :model="promotionForm" label-width="110px">
        <el-form-item label="是否推荐">
          <el-switch v-model="promotionForm.is_recommended" :active-value="1" :inactive-value="0" />
        </el-form-item>
        <el-form-item label="是否置顶">
          <el-switch v-model="promotionForm.is_pinned" :active-value="1" :inactive-value="0" />
        </el-form-item>
        <el-form-item label="推广权重">
          <el-input-number v-model="promotionForm.promotion_weight" :min="0" :max="9999" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="promotionVisible = false">取消</el-button>
        <el-button type="primary" :loading="promotionLoading" @click="onPromotionSubmit">确认</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Refresh, RefreshLeft, Search } from '@element-plus/icons-vue'
import { formatTime } from '@/utils/format'
import { getMallProductList, getMallProductDetail, auditMallProduct, updateMallProductPromotion } from '@/api/mall'

const loading = ref(false)
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)
const list = ref([])
const sortField = ref('created_at')
const sortOrder = ref('descending')

const filters = reactive({ keyword: '', shop_id: null, category_id: null, audit_status: null, status: null })

const onSearch = () => { page.value = 1; loadList() }
const onReset = () => {
  filters.keyword = ''
  filters.shop_id = null
  filters.category_id = null
  filters.audit_status = null
  filters.status = null
  page.value = 1
  loadList()
}
const onSortChange = ({ prop, order }) => {
  sortField.value = prop || 'created_at'
  sortOrder.value = order || 'descending'
  loadList()
}

const auditText = (s) => ({ 0: '待审核', 1: '已通过', 2: '已拒绝' }[s] || '-')
const auditTagType = (s) => ({ 0: 'warning', 1: 'success', 2: 'danger' }[s] || 'info')
const statusText = (s) => ({ 0: '下架', 1: '上架', 2: '草稿' }[s] || '-')
const statusTagType = (s) => ({ 0: 'info', 1: 'success', 2: 'warning' }[s] || 'info')

const loadList = async () => {
  loading.value = true
  try {
    const params = {
      page: page.value,
      page_size: pageSize.value,
      keyword: filters.keyword.trim() || undefined,
      shop_id: filters.shop_id || undefined,
      category_id: filters.category_id || undefined,
      audit_status: filters.audit_status === null || filters.audit_status === '' ? undefined : filters.audit_status,
      status: filters.status === null || filters.status === '' ? undefined : filters.status,
      sort: sortField.value,
      order: sortOrder.value
    }
    const res = await getMallProductList(params)
    const data = res.data || {}
    list.value = data.list || []
    total.value = data.total || 0
  } catch (e) {
    ElMessage.error('加载商品列表失败')
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
    const res = await getMallProductDetail(row.id)
    detail.value = res.data || null
    detailVisible.value = true
  } catch (e) {
    ElMessage.error('加载详情失败')
  }
}

const handleAudit = async (row, auditStatus) => {
  try {
    if (auditStatus === 2) {
      const { value } = await ElMessageBox.prompt('请输入拒绝原因（可选）', '拒绝审核', { inputType: 'textarea', inputPlaceholder: '拒绝原因' })
      await auditMallProduct(row.id, { audit_status: auditStatus, audit_reason: value || '' })
    } else {
      await ElMessageBox.confirm(`确定通过商品 "${row.title}" 的审核吗？`, '提示', { type: 'warning' })
      await auditMallProduct(row.id, { audit_status: auditStatus })
    }
    ElMessage.success('操作成功')
    await loadList()
  } catch (e) { /* cancel */ }
}

const promotionVisible = ref(false)
const promotionLoading = ref(false)
const promotionForm = reactive({ id: null, is_recommended: 0, is_pinned: 0, promotion_weight: 0 })

const openPromotion = (row) => {
  Object.assign(promotionForm, {
    id: row.id,
    is_recommended: row.is_recommended || 0,
    is_pinned: row.is_pinned || 0,
    promotion_weight: row.promotion_weight || 0
  })
  promotionVisible.value = true
}

const onPromotionSubmit = async () => {
  try {
    promotionLoading.value = true
    await updateMallProductPromotion(promotionForm.id, { ...promotionForm })
    ElMessage.success('推广配置已更新')
    promotionVisible.value = false
    await loadList()
  } catch (e) {
    ElMessage.error('更新失败')
  } finally {
    promotionLoading.value = false
  }
}

onMounted(() => loadList())
</script>

<style scoped>
.filter-form { margin-bottom: 12px; padding: 12px 16px; background: #fafafa; border-radius: 4px; }
.filter-form :deep(.el-form-item) { margin-bottom: 8px; margin-right: 12px; }
.toolbar { display: flex; justify-content: space-between; margin-bottom: 12px; }
.pagination-wrap { margin-top: 16px; display: flex; justify-content: flex-end; }
.cover-thumb { width: 50px; height: 50px; border-radius: 4px; border: 1px solid #ebeef5; }
.cover-empty { display: flex; align-items: center; justify-content: center; background: #fafafa; color: #909399; font-size: 12px; }
.title-text { font-weight: 500; color: #303133; }
.text-muted { color: #909399; }
.text-xs { font-size: 12px; }
.line-through { text-decoration: line-through; }
.content-box { white-space: pre-wrap; word-break: break-all; max-height: 200px; overflow-y: auto; }
.images-grid { display: flex; flex-wrap: wrap; gap: 8px; margin-top: 12px; }
.image-item { width: 100px; height: 100px; border-radius: 4px; border: 1px solid #ebeef5; }
</style>
