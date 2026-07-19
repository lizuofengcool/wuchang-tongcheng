<template>
  <div class="app-container">
    <div class="page-card">
      <el-form :inline="true" :model="filters" class="filter-form" @submit.prevent>
        <el-form-item label="退款单号">
          <el-input v-model="filters.refund_no" placeholder="退款单号" clearable style="width: 180px" @keyup.enter="onSearch" />
        </el-form-item>
        <el-form-item label="订单ID">
          <el-input v-model="filters.order_id" placeholder="订单ID" clearable style="width: 140px" @keyup.enter="onSearch" />
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="filters.status" placeholder="全部" clearable style="width: 140px" @change="onSearch">
            <el-option v-for="(label, val) in statusMap" :key="val" :label="label" :value="Number(val)" />
          </el-select>
        </el-form-item>
        <el-form-item label="退款类型">
          <el-select v-model="filters.refund_type" placeholder="全部" clearable style="width: 140px" @change="onSearch">
            <el-option label="退货退款" value="return" />
            <el-option label="换货" value="exchange" />
            <el-option label="维修" value="repair" />
            <el-option label="仅退款" value="refund" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" :icon="Search" @click="onSearch">搜索</el-button>
          <el-button :icon="RefreshLeft" @click="onReset">重置</el-button>
        </el-form-item>
      </el-form>

      <div class="toolbar"><el-button :icon="Refresh" @click="loadList">刷新</el-button></div>

      <el-table v-loading="loading" :data="list" border stripe>
        <el-table-column prop="refund_no" label="退款单号" width="180" fixed="left" />
        <el-table-column prop="order_id" label="订单ID" width="90" />
        <el-table-column prop="ershou_id" label="商品ID" width="90" />
        <el-table-column label="买家" width="90">
          <template #default="{ row }">#{{ row.buyer_id }}</template>
        </el-table-column>
        <el-table-column label="卖家" width="90">
          <template #default="{ row }">#{{ row.seller_id }}</template>
        </el-table-column>
        <el-table-column label="退款金额" width="120">
          <template #default="{ row }">
            <span class="price">¥{{ Number(row.refund_amount || 0).toFixed(2) }}</span>
          </template>
        </el-table-column>
        <el-table-column label="类型" width="100">
          <template #default="{ row }">{{ typeMap[row.refund_type] || row.refund_type }}</template>
        </el-table-column>
        <el-table-column label="原因" min-width="180" show-overflow-tooltip>
          <template #default="{ row }">{{ row.reason }}</template>
        </el-table-column>
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="statusTagType(row.status)" size="small">{{ row.status_text || statusMap[row.status] || '-' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="凭证" width="80">
          <template #default="{ row }">
            <el-badge v-if="row.evidence_images && row.evidence_images.length" :value="row.evidence_images.length" type="primary">
              <el-icon><Picture /></el-icon>
            </el-badge>
            <span v-else class="text-muted">-</span>
          </template>
        </el-table-column>
        <el-table-column label="SLA截止" width="160">
          <template #default="{ row }">
            <span :class="{ 'text-danger': isSLAExpired(row.sla_deadline) }">{{ formatTime(row.sla_deadline) }}</span>
          </template>
        </el-table-column>
        <el-table-column label="申请时间" width="160">
          <template #default="{ row }">{{ formatTime(row.created_at) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="160" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="openDetail(row)">详情</el-button>
            <el-button v-if="row.status === 0" type="success" link size="small" @click="openProcess(row)">处理</el-button>
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
    <el-dialog v-model="detailVisible" title="退款详情" width="700px">
      <el-descriptions v-if="detail" :column="2" border>
        <el-descriptions-item label="退款单号">{{ detail.refund_no }}</el-descriptions-item>
        <el-descriptions-item label="订单ID">{{ detail.order_id }}</el-descriptions-item>
        <el-descriptions-item label="商品ID">{{ detail.ershou_id }}</el-descriptions-item>
        <el-descriptions-item label="退款金额"><span class="price">¥{{ Number(detail.refund_amount || 0).toFixed(2) }}</span></el-descriptions-item>
        <el-descriptions-item label="退款类型">{{ typeMap[detail.refund_type] || detail.refund_type }}</el-descriptions-item>
        <el-descriptions-item label="状态">
          <el-tag :type="statusTagType(detail.status)" size="small">{{ detail.status_text || statusMap[detail.status] }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="买家ID">{{ detail.buyer_id }}</el-descriptions-item>
        <el-descriptions-item label="卖家ID">{{ detail.seller_id }}</el-descriptions-item>
        <el-descriptions-item label="原因" :span="2">{{ detail.reason }}</el-descriptions-item>
        <el-descriptions-item v-if="detail.description" label="详细描述" :span="2">{{ detail.description }}</el-descriptions-item>
        <el-descriptions-item v-if="detail.seller_reason" label="卖家说明" :span="2">{{ detail.seller_reason }}</el-descriptions-item>
        <el-descriptions-item v-if="detail.arbitration_result" label="仲裁结果" :span="2">{{ detail.arbitration_result }}</el-descriptions-item>
        <el-descriptions-item label="SLA截止">{{ formatTime(detail.sla_deadline) }}</el-descriptions-item>
        <el-descriptions-item label="仲裁人ID">{{ detail.arbitrator_id || '-' }}</el-descriptions-item>
        <el-descriptions-item label="完成时间">{{ formatTime(detail.completed_at) }}</el-descriptions-item>
        <el-descriptions-item label="创建时间">{{ formatTime(detail.created_at) }}</el-descriptions-item>
        <el-descriptions-item v-if="detail.evidence_images && detail.evidence_images.length" label="凭证" :span="2">
          <div class="images-grid">
            <el-image v-for="(img, idx) in detail.evidence_images" :key="idx" :src="img" fit="cover" class="image-item" :preview-src-list="detail.evidence_images" :initial-index="idx" preview-teleported />
          </div>
        </el-descriptions-item>
      </el-descriptions>
      <template #footer><el-button @click="detailVisible = false">关闭</el-button></template>
    </el-dialog>

    <!-- 处理退款弹窗 -->
    <el-dialog v-model="processVisible" title="处理退款" width="500px">
      <el-form :model="processForm" label-width="100px">
        <el-form-item label="处理操作">
          <el-radio-group v-model="processForm.action">
            <el-radio value="approve">同意退款</el-radio>
            <el-radio value="reject">拒绝退款</el-radio>
            <el-radio value="arbitrate">仲裁</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item v-if="processForm.action === 'reject'" label="拒绝原因">
          <el-input v-model="processForm.seller_reason" type="textarea" :rows="3" placeholder="拒绝原因" />
        </el-form-item>
        <el-form-item v-if="processForm.action === 'arbitrate'" label="仲裁结果">
          <el-input v-model="processForm.arbitration_result" type="textarea" :rows="3" placeholder="仲裁结果" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="processVisible = false">取消</el-button>
        <el-button type="primary" :loading="processLoading" @click="onProcess">确认处理</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { Refresh, RefreshLeft, Search, Picture } from '@element-plus/icons-vue'
import { listErshouRefunds, processErshouRefund } from '@/api/ershou'
import { formatTime } from '@/utils/format'

const loading = ref(false)
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)
const list = ref([])

const filters = reactive({ refund_no: '', order_id: '', status: null, refund_type: '' })

const statusMap = { 0: '待处理', 1: '已同意', 2: '已退款', 3: '已拒绝', 4: '仲裁中' }
const typeMap = { return: '退货退款', exchange: '换货', repair: '维修', refund: '仅退款' }

const statusTagType = (s) => ({ 0: 'warning', 1: 'primary', 2: 'success', 3: 'danger', 4: 'info' }[s] || 'info')

const isSLAExpired = (t) => t && new Date(t).getTime() < Date.now()

const onSearch = () => { page.value = 1; loadList() }
const onReset = () => { filters.refund_no = ''; filters.order_id = ''; filters.status = null; filters.refund_type = ''; page.value = 1; loadList() }

const loadList = async () => {
  loading.value = true
  try {
    const res = await listErshouRefunds({
      page: page.value,
      page_size: pageSize.value,
      refund_no: filters.refund_no || undefined,
      order_id: filters.order_id || undefined,
      status: filters.status === null || filters.status === '' ? undefined : filters.status,
      refund_type: filters.refund_type || undefined
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

const detailVisible = ref(false)
const detail = ref(null)
const openDetail = (row) => { detail.value = row; detailVisible.value = true }

const processVisible = ref(false)
const processLoading = ref(false)
const processForm = reactive({ id: null, action: 'approve', seller_reason: '', arbitration_result: '' })

const openProcess = (row) => {
  processForm.id = row.id
  processForm.action = 'approve'
  processForm.seller_reason = ''
  processForm.arbitration_result = ''
  processVisible.value = true
}

const onProcess = async () => {
  processLoading.value = true
  try {
    await processErshouRefund(processForm.id, {
      action: processForm.action,
      seller_reason: processForm.seller_reason,
      arbitration_result: processForm.arbitration_result
    })
    ElMessage.success('处理成功')
    processVisible.value = false
    await loadList()
  } catch (e) {
    // 失败已提示
  } finally {
    processLoading.value = false
  }
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
.images-grid { display: flex; flex-wrap: wrap; gap: 8px; }
.image-item { width: 100px; height: 100px; border-radius: 4px; border: 1px solid #ebeef5; }
.pagination-wrap { margin-top: 16px; display: flex; justify-content: flex-end; }
</style>
