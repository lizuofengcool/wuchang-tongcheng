<template>
  <div class="app-container">
    <!-- 评价统计 -->
    <el-row :gutter="16" class="stat-row">
      <el-col :xs="12" :sm="8" :md="6">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value">{{ stats.total }}</div><div class="stat-label">评价总数</div></div></el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="6">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value text-warning">{{ stats.pending }}</div><div class="stat-label">待审核</div></div></el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="6">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value text-success">{{ stats.approved }}</div><div class="stat-label">已通过</div></div></el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="6">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value text-primary">{{ stats.avg_rating }}</div><div class="stat-label">平均评分</div></div></el-card>
      </el-col>
    </el-row>

    <div class="page-card">
      <el-form :inline="true" :model="filters" class="filter-form" @submit.prevent>
        <el-form-item label="商品ID">
          <el-input-number v-model="filters.product_id" :controls="false" :min="1" placeholder="商品ID" style="width: 140px" />
        </el-form-item>
        <el-form-item label="店铺ID">
          <el-input-number v-model="filters.shop_id" :controls="false" :min="1" placeholder="店铺ID" style="width: 140px" />
        </el-form-item>
        <el-form-item label="用户ID">
          <el-input-number v-model="filters.user_id" :controls="false" :min="1" placeholder="用户ID" style="width: 140px" />
        </el-form-item>
        <el-form-item label="评分">
          <el-select v-model="filters.rating" placeholder="全部" clearable style="width: 110px" @change="onSearch">
            <el-option label="5 星" :value="5" />
            <el-option label="4 星" :value="4" />
            <el-option label="3 星" :value="3" />
            <el-option label="2 星" :value="2" />
            <el-option label="1 星" :value="1" />
          </el-select>
        </el-form-item>
        <el-form-item label="审核状态">
          <el-select v-model="filters.audit_status" placeholder="全部" clearable style="width: 130px" @change="onSearch">
            <el-option label="待审核" :value="0" />
            <el-option label="已通过" :value="1" />
            <el-option label="已拒绝" :value="2" />
          </el-select>
        </el-form-item>
        <el-form-item label="是否有图">
          <el-select v-model="filters.has_images" placeholder="全部" clearable style="width: 110px" @change="onSearch">
            <el-option label="有图" :value="1" />
            <el-option label="无图" :value="0" />
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
        <el-table-column label="商品" min-width="200">
          <template #default="{ row }">
            <el-link type="primary" :underline="false" @click="openDetail(row)">{{ row.product_name || `商品#${row.product_id}` }}</el-link>
          </template>
        </el-table-column>
        <el-table-column label="店铺ID" width="90" prop="shop_id" />
        <el-table-column label="用户ID" width="90" prop="user_id" />
        <el-table-column label="评分" width="140">
          <template #default="{ row }">
            <el-rate :model-value="row.rating || 0" disabled size="small" />
          </template>
        </el-table-column>
        <el-table-column label="内容" min-width="220" show-overflow-tooltip>
          <template #default="{ row }">{{ row.content || '-' }}</template>
        </el-table-column>
        <el-table-column label="有图" width="70">
          <template #default="{ row }">
            <el-tag v-if="row.images && row.images.length" type="success" size="small">{{ row.images.length }}图</el-tag>
            <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column label="审核" width="100">
          <template #default="{ row }">
            <el-tag :type="auditTagType(row.audit_status)" size="small">{{ auditText(row.audit_status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="点赞/踩" width="100">
          <template #default="{ row }">
            <span class="text-success">👍{{ row.like_count || 0 }}</span>
            <span class="text-danger" style="margin-left: 6px">👎{{ row.dislike_count || 0 }}</span>
          </template>
        </el-table-column>
        <el-table-column label="创建时间" width="160" prop="created_at" sortable="custom">
          <template #default="{ row }">{{ formatTime(row.created_at) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="220" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="openDetail(row)">详情</el-button>
            <el-button v-if="row.audit_status === 0" type="success" link size="small" @click="openAudit(row, 1)">通过</el-button>
            <el-button v-if="row.audit_status === 0" type="danger" link size="small" @click="openAudit(row, 2)">拒绝</el-button>
            <el-button type="warning" link size="small" @click="openReply(row)">回复</el-button>
          </template>
        </el-table-column>
      </el-table>

      <div class="pagination-wrap">
        <el-pagination v-model:current-page="page" v-model:page-size="pageSize" :total="total" :page-sizes="[10, 20, 50, 100]" layout="total, sizes, prev, pager, next, jumper" background @current-change="loadList" @size-change="loadList" />
      </div>
    </div>

    <!-- 详情弹窗 -->
    <el-dialog v-model="detailVisible" title="评价详情" width="760px" destroy-on-close>
      <el-descriptions v-if="detail" :column="2" border>
        <el-descriptions-item label="ID">{{ detail.id }}</el-descriptions-item>
        <el-descriptions-item label="评分">
          <el-rate :model-value="detail.rating || 0" disabled />
        </el-descriptions-item>
        <el-descriptions-item label="商品ID">{{ detail.product_id }}</el-descriptions-item>
        <el-descriptions-item label="商品名">{{ detail.product_name || '-' }}</el-descriptions-item>
        <el-descriptions-item label="订单ID">{{ detail.order_id || '-' }}</el-descriptions-item>
        <el-descriptions-item label="订单项ID">{{ detail.order_item_id || '-' }}</el-descriptions-item>
        <el-descriptions-item label="店铺ID">{{ detail.shop_id }}</el-descriptions-item>
        <el-descriptions-item label="用户ID">{{ detail.user_id }}</el-descriptions-item>
        <el-descriptions-item label="用户名">{{ detail.username || '-' }}</el-descriptions-item>
        <el-descriptions-item label="SKU ID">{{ detail.sku_id || '-' }}</el-descriptions-item>
        <el-descriptions-item label="评价内容" :span="2">{{ detail.content || '-' }}</el-descriptions-item>
        <el-descriptions-item label="追加评价" :span="2">{{ detail.append_content || '-' }}</el-descriptions-item>
        <el-descriptions-item label="评价图片" :span="2">
          <div v-if="detail.images && detail.images.length" class="image-list">
            <el-image v-for="(img, i) in detail.images" :key="i" :src="img" :preview-src-list="detail.images" fit="cover" class="review-thumb" />
          </div>
          <span v-else>-</span>
        </el-descriptions-item>
        <el-descriptions-item label="视频">
          <video v-if="detail.video" :src="detail.video" controls style="max-width: 200px; max-height: 120px" />
          <span v-else>-</span>
        </el-descriptions-item>
        <el-descriptions-item label="审核状态">
          <el-tag :type="auditTagType(detail.audit_status)" size="small">{{ auditText(detail.audit_status) }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="审核原因" :span="2">{{ detail.audit_reason || '-' }}</el-descriptions-item>
        <el-descriptions-item label="商家回复" :span="2">{{ detail.reply_content || '-' }}</el-descriptions-item>
        <el-descriptions-item label="回复时间">{{ formatTime(detail.reply_at) }}</el-descriptions-item>
        <el-descriptions-item label="点赞">{{ detail.like_count || 0 }}</el-descriptions-item>
        <el-descriptions-item label="踩">{{ detail.dislike_count || 0 }}</el-descriptions-item>
        <el-descriptions-item label="匿名">
          <el-tag :type="detail.is_anonymous ? 'warning' : 'info'" size="small">{{ detail.is_anonymous ? '匿名' : '实名' }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="状态">{{ detail.status === 1 ? '显示' : (detail.status === 0 ? '隐藏' : '-') }}</el-descriptions-item>
        <el-descriptions-item label="创建时间">{{ formatTime(detail.created_at) }}</el-descriptions-item>
        <el-descriptions-item label="更新时间">{{ formatTime(detail.updated_at) }}</el-descriptions-item>
      </el-descriptions>
    </el-dialog>

    <!-- 审核弹窗 -->
    <el-dialog v-model="auditVisible" title="评价审核" width="480px" destroy-on-close>
      <el-form :model="auditForm" label-width="100px">
        <el-form-item label="评价ID">{{ auditForm.id }}</el-form-item>
        <el-form-item label="评分">
          <el-rate :model-value="auditForm.rating || 0" disabled />
        </el-form-item>
        <el-form-item label="评价内容">{{ auditForm.content }}</el-form-item>
        <el-form-item label="审核结果">
          <el-tag :type="auditForm.audit_status === 1 ? 'success' : 'danger'">{{ auditForm.audit_status === 1 ? '通过' : '拒绝' }}</el-tag>
        </el-form-item>
        <el-form-item label="审核原因">
          <el-input v-model="auditForm.audit_reason" type="textarea" :rows="3" placeholder="审核原因（拒绝时必填）" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="auditVisible = false">取消</el-button>
        <el-button type="primary" :loading="auditLoading" @click="onAuditSubmit">确认</el-button>
      </template>
    </el-dialog>

    <!-- 回复弹窗 -->
    <el-dialog v-model="replyVisible" title="商家回复" width="480px" destroy-on-close>
      <el-form :model="replyForm" label-width="100px">
        <el-form-item label="评价ID">{{ replyForm.id }}</el-form-item>
        <el-form-item label="评价内容">{{ replyForm.content }}</el-form-item>
        <el-form-item v-if="replyForm.reply_content" label="原回复">{{ replyForm.reply_content }}</el-form-item>
        <el-form-item label="回复内容">
          <el-input v-model="replyForm.reply_content" type="textarea" :rows="4" maxlength="500" show-word-limit placeholder="请输入回复内容" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="replyVisible = false">取消</el-button>
        <el-button type="primary" :loading="replyLoading" @click="onReplySubmit">确认回复</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { Refresh, RefreshLeft, Search } from '@element-plus/icons-vue'
import { formatTime } from '@/utils/format'
import { getMallReviewList, getMallReviewDetail, updateMallReviewStatus, replyMallReview } from '@/api/mall'

const loading = ref(false)
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)
const list = ref([])
const sortField = ref('created_at')
const sortOrder = ref('descending')

const stats = reactive({ total: 0, pending: 0, approved: 0, avg_rating: '0.0' })

const filters = reactive({
  product_id: null, shop_id: null, user_id: null,
  rating: null, audit_status: null, has_images: null
})

const onSearch = () => { page.value = 1; loadList() }
const onReset = () => {
  filters.product_id = null
  filters.shop_id = null
  filters.user_id = null
  filters.rating = null
  filters.audit_status = null
  filters.has_images = null
  page.value = 1
  loadList()
}
const onSortChange = ({ prop, order }) => {
  sortField.value = prop || 'created_at'
  sortOrder.value = order || 'descending'
  loadList()
}

const auditText = (s) => ({ 0: '待审核', 1: '已通过', 2: '已拒绝' }[s] ?? '-')
const auditTagType = (s) => ({ 0: 'warning', 1: 'success', 2: 'danger' }[s] || 'info')

const loadList = async () => {
  loading.value = true
  try {
    const params = {
      page: page.value,
      page_size: pageSize.value,
      product_id: filters.product_id || undefined,
      shop_id: filters.shop_id || undefined,
      user_id: filters.user_id || undefined,
      rating: filters.rating || undefined,
      audit_status: filters.audit_status === null || filters.audit_status === '' ? undefined : filters.audit_status,
      has_images: filters.has_images === null || filters.has_images === '' ? undefined : filters.has_images,
      sort: sortField.value,
      order: sortOrder.value
    }
    const res = await getMallReviewList(params)
    const data = res.data || {}
    list.value = data.list || []
    total.value = data.total || 0
    computeStats()
  } catch (e) {
    ElMessage.error('加载评价列表失败')
    list.value = []
    total.value = 0
  } finally {
    loading.value = false
  }
}

const computeStats = () => {
  const total = list.value.length
  const pending = list.value.filter((r) => r.audit_status === 0).length
  const approved = list.value.filter((r) => r.audit_status === 1).length
  const sum = list.value.reduce((s, r) => s + Number(r.rating || 0), 0)
  const avg = total > 0 ? (sum / total).toFixed(1) : '0.0'
  Object.assign(stats, { total, pending, approved, avg_rating: avg })
}

const detailVisible = ref(false)
const detail = ref(null)

const openDetail = async (row) => {
  try {
    const res = await getMallReviewDetail(row.id)
    detail.value = res.data || null
    detailVisible.value = true
  } catch (e) {
    ElMessage.error('加载详情失败')
  }
}

// ===== 审核 =====
const auditVisible = ref(false)
const auditLoading = ref(false)
const auditForm = reactive({ id: null, rating: 0, content: '', audit_status: 1, audit_reason: '' })

const openAudit = (row, status) => {
  Object.assign(auditForm, {
    id: row.id,
    rating: row.rating,
    content: row.content || '-',
    audit_status: status,
    audit_reason: ''
  })
  auditVisible.value = true
}

const onAuditSubmit = async () => {
  if (auditForm.audit_status === 2 && !auditForm.audit_reason.trim()) {
    ElMessage.warning('拒绝时请填写原因')
    return
  }
  try {
    auditLoading.value = true
    await updateMallReviewStatus(auditForm.id, {
      audit_status: auditForm.audit_status,
      audit_reason: auditForm.audit_reason
    })
    ElMessage.success('审核成功')
    auditVisible.value = false
    await loadList()
  } catch (e) {
    ElMessage.error('审核失败')
  } finally {
    auditLoading.value = false
  }
}

// ===== 回复 =====
const replyVisible = ref(false)
const replyLoading = ref(false)
const replyForm = reactive({ id: null, content: '', reply_content: '' })

const openReply = async (row) => {
  try {
    const res = await getMallReviewDetail(row.id)
    const detail = res.data || {}
    Object.assign(replyForm, {
      id: row.id,
      content: row.content || '-',
      reply_content: detail.reply_content || ''
    })
    replyVisible.value = true
  } catch (e) {
    ElMessage.error('加载评价详情失败')
  }
}

const onReplySubmit = async () => {
  if (!replyForm.reply_content.trim()) {
    ElMessage.warning('请输入回复内容')
    return
  }
  try {
    replyLoading.value = true
    await replyMallReview(replyForm.id, { reply_content: replyForm.reply_content })
    ElMessage.success('回复成功')
    replyVisible.value = false
    await loadList()
  } catch (e) {
    ElMessage.error('回复失败')
  } finally {
    replyLoading.value = false
  }
}

onMounted(() => loadList())
</script>

<style scoped>
.stat-row { margin-bottom: 16px; }
.stat-item { text-align: center; padding: 16px; }
.stat-value { font-size: 20px; font-weight: 600; color: #303133; }
.text-success { color: #67c23a; }
.text-warning { color: #e6a23c; }
.text-danger { color: #f56c6c; }
.text-primary { color: #409eff; }
.filter-form { margin-bottom: 12px; padding: 12px 16px; background: #fafafa; border-radius: 4px; }
.filter-form :deep(.el-form-item) { margin-bottom: 8px; margin-right: 12px; }
.toolbar { display: flex; justify-content: space-between; margin-bottom: 12px; }
.pagination-wrap { margin-top: 16px; display: flex; justify-content: flex-end; }
.image-list { display: flex; gap: 8px; flex-wrap: wrap; }
.review-thumb { width: 80px; height: 80px; border-radius: 4px; }
</style>
