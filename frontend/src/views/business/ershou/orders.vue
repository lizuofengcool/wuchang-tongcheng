<template>
  <div class="app-container">
    <!-- 筛选 -->
    <div class="page-card">
      <el-form :inline="true" :model="filters" class="filter-form" @submit.prevent>
        <el-form-item label="订单号">
          <el-input v-model="filters.order_no" placeholder="订单号" clearable style="width: 180px" @keyup.enter="onSearch" />
        </el-form-item>
        <el-form-item label="关键词">
          <el-input v-model="filters.keyword" placeholder="商品/买家/卖家" clearable style="width: 200px" @keyup.enter="onSearch" />
        </el-form-item>
        <el-form-item label="角色">
          <el-select v-model="filters.role" placeholder="全部" clearable style="width: 120px" @change="onSearch">
            <el-option label="买家" value="buyer" />
            <el-option label="卖家" value="seller" />
            <el-option label="全部" value="all" />
          </el-select>
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="filters.status" placeholder="全部" clearable style="width: 140px" @change="onSearch">
            <el-option v-for="(label, val) in statusMap" :key="val" :label="label" :value="Number(val)" />
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

      <div class="toolbar">
        <el-button :icon="Refresh" @click="loadList">刷新</el-button>
      </div>

      <el-table v-loading="loading" :data="list" border stripe>
        <el-table-column prop="order_no" label="订单号" width="180" fixed="left" />
        <el-table-column label="商品" min-width="220">
          <template #default="{ row }">
            <div v-if="row.items && row.items.length" class="order-items">
              <div v-for="it in row.items" :key="it.id" class="order-item">
                <el-image v-if="it.cover_image" :src="it.cover_image" fit="cover" class="item-thumb" />
                <div class="item-info">
                  <div class="item-title">{{ it.title }}</div>
                  <div class="item-meta">
                    ¥{{ Number(it.unit_price || 0).toFixed(2) }} × {{ it.quantity }}
                  </div>
                </div>
              </div>
            </div>
            <span v-else class="text-muted">-</span>
          </template>
        </el-table-column>
        <el-table-column label="买家" width="120">
          <template #default="{ row }">#{{ row.buyer_id }}</template>
        </el-table-column>
        <el-table-column label="卖家" width="120">
          <template #default="{ row }">#{{ row.seller_id }}</template>
        </el-table-column>
        <el-table-column label="金额" width="120">
          <template #default="{ row }">
            <span class="price">¥{{ Number(row.total_amount || 0).toFixed(2) }}</span>
          </template>
        </el-table-column>
        <el-table-column label="支付方式" width="100">
          <template #default="{ row }">{{ payMethodText(row.pay_method) }}</template>
        </el-table-column>
        <el-table-column label="状态" width="110">
          <template #default="{ row }">
            <el-tag :type="statusTagType(row.status)" size="small">{{ statusMap[row.status] || '-' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="交付方式" width="100">
          <template #default="{ row }">{{ deliveryText(row.delivery_method) }}</template>
        </el-table-column>
        <el-table-column label="担保" width="70">
          <template #default="{ row }">
            <el-tag v-if="row.escrow_enabled" type="warning" size="small">担保</el-tag>
            <span v-else class="text-muted">-</span>
          </template>
        </el-table-column>
        <el-table-column label="创建时间" width="160">
          <template #default="{ row }">{{ formatTime(row.created_at) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="openDetail(row)">详情</el-button>
            <el-button v-if="canShip(row)" type="success" link size="small" @click="onShip(row)">发货</el-button>
            <el-button v-if="canReceive(row)" type="warning" link size="small" @click="onReceive(row)">确认收货</el-button>
            <el-button v-if="canCancel(row)" type="danger" link size="small" @click="onCancel(row)">取消</el-button>
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

    <!-- 订单详情弹窗 -->
    <el-dialog v-model="detailVisible" title="订单详情" width="900px" destroy-on-close>
      <div v-loading="detailLoading">
        <el-tabs v-if="detail">
          <el-tab-pane label="基本信息">
            <el-descriptions :column="3" border>
              <el-descriptions-item label="订单号" :span="3">{{ detail.order_no }}</el-descriptions-item>
              <el-descriptions-item label="买家ID">{{ detail.buyer_id }}</el-descriptions-item>
              <el-descriptions-item label="卖家ID">{{ detail.seller_id }}</el-descriptions-item>
              <el-descriptions-item label="店铺ID">{{ detail.shop_id || '-' }}</el-descriptions-item>
              <el-descriptions-item label="商品金额">¥{{ Number(detail.item_amount || 0).toFixed(2) }}</el-descriptions-item>
              <el-descriptions-item label="运费">¥{{ Number(detail.delivery_fee || 0).toFixed(2) }}</el-descriptions-item>
              <el-descriptions-item label="优惠">¥{{ Number(detail.discount_amount || 0).toFixed(2) }}</el-descriptions-item>
              <el-descriptions-item label="总金额"><span class="price">¥{{ Number(detail.total_amount || 0).toFixed(2) }}</span></el-descriptions-item>
              <el-descriptions-item label="状态">
                <el-tag :type="statusTagType(detail.status)" size="small">{{ statusMap[detail.status] || '-' }}</el-tag>
              </el-descriptions-item>
              <el-descriptions-item label="支付方式">{{ payMethodText(detail.pay_method) }}</el-descriptions-item>
              <el-descriptions-item label="支付交易号">{{ detail.pay_trade_no || '-' }}</el-descriptions-item>
              <el-descriptions-item label="交付方式">{{ deliveryText(detail.delivery_method) }}</el-descriptions-item>
              <el-descriptions-item label="担保">{{ detail.escrow_enabled ? '已启用' : '未启用' }}</el-descriptions-item>
              <el-descriptions-item label="分期">{{ detail.installment_enabled ? `${detail.installment_periods}期` : '未启用' }}</el-descriptions-item>
              <el-descriptions-item label="联系人">{{ detail.contact_name || '-' }}</el-descriptions-item>
              <el-descriptions-item label="联系电话">{{ maskPhone(detail.contact_phone) }}</el-descriptions-item>
              <el-descriptions-item label="收货地址" :span="3">{{ detail.contact_address || '-' }}</el-descriptions-item>
              <el-descriptions-item label="备注" :span="3">{{ detail.remark || '-' }}</el-descriptions-item>
              <el-descriptions-item label="付款时间">{{ formatTime(detail.paid_at) }}</el-descriptions-item>
              <el-descriptions-item label="发货时间">{{ formatTime(detail.shipped_at) }}</el-descriptions-item>
              <el-descriptions-item label="收货时间">{{ formatTime(detail.received_at) }}</el-descriptions-item>
              <el-descriptions-item label="结算时间">{{ formatTime(detail.settled_at) }}</el-descriptions-item>
              <el-descriptions-item label="关闭时间">{{ formatTime(detail.closed_at) }}</el-descriptions-item>
              <el-descriptions-item label="自动关闭">{{ formatTime(detail.auto_close_at) }}</el-descriptions-item>
              <el-descriptions-item label="自动收货">{{ formatTime(detail.auto_receive_at) }}</el-descriptions-item>
              <el-descriptions-item label="创建时间">{{ formatTime(detail.created_at) }}</el-descriptions-item>
            </el-descriptions>
          </el-tab-pane>

          <el-tab-pane label="商品明细">
            <el-table :data="detail.items || []" border size="small">
              <el-table-column prop="title" label="商品" min-width="200" />
              <el-table-column prop="sku_code" label="SKU" width="140" />
              <el-table-column label="单价" width="100">
                <template #default="{ row }">¥{{ Number(row.unit_price || 0).toFixed(2) }}</template>
              </el-table-column>
              <el-table-column prop="quantity" label="数量" width="80" />
              <el-table-column label="小计" width="100">
                <template #default="{ row }">¥{{ Number(row.subtotal || 0).toFixed(2) }}</template>
              </el-table-column>
              <el-table-column prop="remark" label="备注" min-width="120" />
            </el-table>
          </el-tab-pane>

          <el-tab-pane label="物流跟踪">
            <el-descriptions v-if="logistics" :column="2" border>
              <el-descriptions-item label="物流公司">{{ logistics.express_company }}</el-descriptions-item>
              <el-descriptions-item label="运单号">{{ logistics.tracking_no }}</el-descriptions-item>
              <el-descriptions-item label="状态">{{ logistics.status_text }}</el-descriptions-item>
              <el-descriptions-item label="运费">¥{{ Number(logistics.freight || 0).toFixed(2) }}</el-descriptions-item>
              <el-descriptions-item label="发货人">{{ logistics.shipper_name }}</el-descriptions-item>
              <el-descriptions-item label="发货电话">{{ maskPhone(logistics.shipper_phone) }}</el-descriptions-item>
              <el-descriptions-item label="收货人">{{ logistics.receiver_name }}</el-descriptions-item>
              <el-descriptions-item label="收货电话">{{ maskPhone(logistics.receiver_phone) }}</el-descriptions-item>
              <el-descriptions-item label="发货时间">{{ formatTime(logistics.shipped_at) }}</el-descriptions-item>
              <el-descriptions-item label="签收时间">{{ formatTime(logistics.delivered_at) }}</el-descriptions-item>
            </el-descriptions>
            <div v-else class="empty-text">暂无物流信息</div>
          </el-tab-pane>

          <el-tab-pane label="担保状态">
            <el-descriptions v-if="escrow" :column="2" border>
              <el-descriptions-item label="担保金额">¥{{ Number(escrow.escrow_amount || 0).toFixed(2) }}</el-descriptions-item>
              <el-descriptions-item label="平台手续费">¥{{ Number(escrow.platform_fee || 0).toFixed(2) }}</el-descriptions-item>
              <el-descriptions-item label="卖家所得">¥{{ Number(escrow.seller_amount || 0).toFixed(2) }}</el-descriptions-item>
              <el-descriptions-item label="状态">{{ escrow.status_text }}</el-descriptions-item>
              <el-descriptions-item label="冻结时间">{{ formatTime(escrow.frozen_at) }}</el-descriptions-item>
              <el-descriptions-item label="放款时间">{{ formatTime(escrow.release_at) }}</el-descriptions-item>
            </el-descriptions>
            <div v-else class="empty-text">未启用担保交易</div>
          </el-tab-pane>

          <el-tab-pane label="退款记录">
            <el-descriptions v-if="refund" :column="2" border>
              <el-descriptions-item label="退款单号">{{ refund.refund_no }}</el-descriptions-item>
              <el-descriptions-item label="退款类型">{{ refund.refund_type }}</el-descriptions-item>
              <el-descriptions-item label="退款金额">¥{{ Number(refund.refund_amount || 0).toFixed(2) }}</el-descriptions-item>
              <el-descriptions-item label="状态">{{ refund.status_text }}</el-descriptions-item>
              <el-descriptions-item label="原因" :span="2">{{ refund.reason }}</el-descriptions-item>
              <el-descriptions-item label="申请时间">{{ formatTime(refund.created_at) }}</el-descriptions-item>
              <el-descriptions-item label="完成时间">{{ formatTime(refund.completed_at) }}</el-descriptions-item>
            </el-descriptions>
            <div v-else class="empty-text">暂无退款记录</div>
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
import { Refresh, RefreshLeft, Search } from '@element-plus/icons-vue'
import {
  listErshouOrders, getErshouOrder, updateErshouOrderStatus,
  getErshouLogistics, getErshouEscrow, getErshouOrderRefund
} from '@/api/ershou'
import { formatTime } from '@/utils/format'

const loading = ref(false)
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)
const list = ref([])

const filters = reactive({
  order_no: '', keyword: '', role: '', status: null, dateRange: null
})

const statusMap = {
  0: '待付款', 1: '待发货', 2: '已发货', 3: '待收货',
  4: '已完成', 5: '已取消', 6: '已退款', 7: '已关闭'
}
const statusTagType = (s) => ({
  0: 'info', 1: 'warning', 2: 'primary', 3: 'warning',
  4: 'success', 5: 'info', 6: 'danger', 7: 'info'
}[s] || 'info')

const payMethodText = (m) => ({ wechat: '微信', alipay: '支付宝', balance: '余额', installment: '分期' }[m] || '-')
const deliveryText = (d) => ({ face: '当面交易', self: '自提', express: '快递' }[d] || '-')

const maskPhone = (p) => {
  if (!p) return '-'
  const s = String(p)
  if (s.length < 7) return s
  return s.slice(0, 3) + '****' + s.slice(-4)
}

const canShip = (row) => row.status === 1
const canReceive = (row) => row.status === 2 || row.status === 3
const canCancel = (row) => row.status === 0 || row.status === 1

const onSearch = () => { page.value = 1; loadList() }
const onReset = () => {
  filters.order_no = ''
  filters.keyword = ''
  filters.role = ''
  filters.status = null
  filters.dateRange = null
  page.value = 1
  loadList()
}

const loadList = async () => {
  loading.value = true
  try {
    const params = {
      page: page.value,
      page_size: pageSize.value,
      order_no: filters.order_no || undefined,
      keyword: filters.keyword || undefined,
      role: filters.role || undefined,
      status: filters.status === null || filters.status === '' ? undefined : filters.status
    }
    if (filters.dateRange && filters.dateRange.length === 2) {
      params.start_date = filters.dateRange[0]
      params.end_date = filters.dateRange[1]
    }
    const res = await listErshouOrders(params)
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

// ===== 详情 =====
const detailVisible = ref(false)
const detailLoading = ref(false)
const detail = ref(null)
const logistics = ref(null)
const escrow = ref(null)
const refund = ref(null)

const openDetail = async (row) => {
  detail.value = row
  logistics.value = null
  escrow.value = null
  refund.value = null
  detailVisible.value = true
  detailLoading.value = true
  try {
    const [dRes, logRes, escRes, refRes] = await Promise.all([
      getErshouOrder(row.id),
      getErshouLogistics(row.id).catch(() => ({ data: null })),
      getErshouEscrow(row.id).catch(() => ({ data: null })),
      getErshouOrderRefund(row.id).catch(() => ({ data: null }))
    ])
    if (dRes.data) detail.value = dRes.data
    logistics.value = logRes.data || null
    escrow.value = escRes.data || null
    refund.value = refRes.data || null
  } catch (e) {
    // ignore
  } finally {
    detailLoading.value = false
  }
}

// ===== 操作 =====
const onShip = async (row) => {
  try {
    const { value } = await ElMessageBox.prompt('请输入物流单号（可选）', '发货', {
      inputPlaceholder: '物流单号'
    })
    await updateErshouOrderStatus(row.id, { action: 'ship', remark: value || '' })
    ElMessage.success('已发货')
    await loadList()
  } catch (e) { /* cancel */ }
}

const onReceive = async (row) => {
  try {
    await ElMessageBox.confirm('确认收货？', '提示', { type: 'warning' })
    await updateErshouOrderStatus(row.id, { action: 'receive' })
    ElMessage.success('已确认收货')
    await loadList()
  } catch (e) { /* cancel */ }
}

const onCancel = async (row) => {
  try {
    await ElMessageBox.confirm('确认取消该订单？', '提示', { type: 'warning' })
    await updateErshouOrderStatus(row.id, { action: 'cancel' })
    ElMessage.success('已取消')
    await loadList()
  } catch (e) { /* cancel */ }
}

onMounted(() => {
  loadList()
})
</script>

<style scoped>
.filter-form {
  margin-bottom: 12px; padding: 12px 16px;
  background: #fafafa; border-radius: 4px;
}
.filter-form :deep(.el-form-item) { margin-bottom: 8px; margin-right: 12px; }
.toolbar { display: flex; justify-content: space-between; margin-bottom: 12px; }
.order-items { display: flex; flex-direction: column; gap: 6px; }
.order-item { display: flex; align-items: center; gap: 8px; }
.item-thumb { width: 40px; height: 40px; border-radius: 4px; border: 1px solid #ebeef5; flex-shrink: 0; }
.item-info { flex: 1; min-width: 0; }
.item-title { font-size: 13px; color: #303133; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.item-meta { font-size: 12px; color: #909399; margin-top: 2px; }
.price { color: #f56c6c; font-weight: 600; }
.text-muted { color: #909399; }
.empty-text { color: #909399; text-align: center; padding: 24px 0; }
.pagination-wrap { margin-top: 16px; display: flex; justify-content: flex-end; }
</style>
