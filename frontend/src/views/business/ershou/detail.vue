<template>
  <div class="app-container" v-loading="loading">
    <!-- 返回按钮 -->
    <div class="back-bar">
      <el-button :icon="ArrowLeft" @click="goBack">返回列表</el-button>
      <span class="page-title">商品详情 #{{ detail?.id || id }}</span>
    </div>

    <div v-if="detail" class="detail-wrap">
      <el-row :gutter="16">
        <!-- 左侧：图集 + 基本信息 -->
        <el-col :xs="24" :md="16">
          <el-card class="section-card">
            <template #header>
              <span class="section-title">基本信息</span>
            </template>
            <el-descriptions :column="3" border>
              <el-descriptions-item label="标题" :span="3">{{ detail.title }}</el-descriptions-item>
              <el-descriptions-item label="分类">{{ categoryName(detail.category_id) }}</el-descriptions-item>
              <el-descriptions-item label="品牌">{{ detail.brand || '-' }}</el-descriptions-item>
              <el-descriptions-item label="成色">{{ conditionText(detail.condition) }}</el-descriptions-item>
              <el-descriptions-item label="售价">
                <span class="price">¥{{ Number(detail.price || 0).toFixed(2) }}</span>
              </el-descriptions-item>
              <el-descriptions-item label="原价">
                <span v-if="detail.original_price > 0" class="original-price">¥{{ Number(detail.original_price).toFixed(2) }}</span>
                <span v-else>-</span>
              </el-descriptions-item>
              <el-descriptions-item label="价格单位">{{ detail.price_unit || '-' }}</el-descriptions-item>
              <el-descriptions-item label="加急">
                <el-tag v-if="detail.is_urgent" type="danger" size="small">急转</el-tag>
                <span v-else>否</span>
              </el-descriptions-item>
              <el-descriptions-item label="交易方式">{{ deliveryText(detail.delivery_method) }}</el-descriptions-item>
              <el-descriptions-item label="审核状态">
                <el-tag :type="auditTagType(detail.audit_status)" size="small">{{ auditText(detail.audit_status) }}</el-tag>
              </el-descriptions-item>
              <el-descriptions-item label="物品状态">
                <el-tag :type="statusTagType(detail.status)" size="small" effect="plain">{{ statusText(detail.status) }}</el-tag>
              </el-descriptions-item>
              <el-descriptions-item v-if="detail.summary" label="摘要" :span="3">{{ detail.summary }}</el-descriptions-item>
              <el-descriptions-item v-if="detail.content" label="详细描述" :span="3">
                <div class="content-box">{{ detail.content }}</div>
              </el-descriptions-item>
              <el-descriptions-item v-if="detail.audit_reason" label="审核原因" :span="3">{{ detail.audit_reason }}</el-descriptions-item>
            </el-descriptions>
          </el-card>

          <el-card class="section-card">
            <template #header>
              <span class="section-title">图集（{{ allImages.length }}）</span>
            </template>
            <div v-if="!allImages.length" class="empty-text">暂无图片</div>
            <div v-else class="images-grid">
              <el-image
                v-for="(img, idx) in allImages"
                :key="idx"
                :src="img"
                fit="cover"
                class="image-item"
                :preview-src-list="allImages"
                :initial-index="idx"
                preview-teleported
              />
            </div>
          </el-card>

          <el-card class="section-card">
            <template #header>
              <div class="card-header-flex">
                <span class="section-title">SKU 规格（{{ skus.length }}）</span>
              </div>
            </template>
            <el-table :data="skus" border size="small">
              <el-table-column prop="sku_code" label="SKU编码" width="140" />
              <el-table-column prop="name" label="名称" min-width="160" />
              <el-table-column label="规格" width="180">
                <template #default="{ row }">
                  {{ [row.color, row.size, row.version].filter(Boolean).join(' / ') || '-' }}
                </template>
              </el-table-column>
              <el-table-column label="价格" width="100">
                <template #default="{ row }">¥{{ Number(row.price || 0).toFixed(2) }}</template>
              </el-table-column>
              <el-table-column prop="stock" label="库存" width="80" />
              <el-table-column prop="sold_count" label="已售" width="80" />
              <el-table-column prop="barcode" label="条码" width="120" />
              <el-table-column label="状态" width="90">
                <template #default="{ row }">
                  <el-tag :type="row.status === 1 ? 'success' : 'info'" size="small">
                    {{ { 0: '禁用', 1: '在售', 2: '售罄' }[row.status] || '-' }}
                  </el-tag>
                </template>
              </el-table-column>
            </el-table>
          </el-card>

          <el-card class="section-card">
            <template #header>
              <span class="section-title">位置信息</span>
            </template>
            <el-descriptions :column="2" border>
              <el-descriptions-item label="地址" :span="2">{{ detail.address || '-' }}</el-descriptions-item>
              <el-descriptions-item label="经度">{{ detail.longitude || '-' }}</el-descriptions-item>
              <el-descriptions-item label="纬度">{{ detail.latitude || '-' }}</el-descriptions-item>
              <el-descriptions-item label="地区ID">{{ detail.region_id || '-' }}</el-descriptions-item>
              <el-descriptions-item label="距离">{{ detail.distance ? detail.distance + ' km' : '-' }}</el-descriptions-item>
            </el-descriptions>
          </el-card>

          <el-card class="section-card">
            <template #header>
              <span class="section-title">联系方式</span>
            </template>
            <el-descriptions :column="2" border>
              <el-descriptions-item label="联系人">{{ detail.user_name || `用户#${detail.user_id}` }}</el-descriptions-item>
              <el-descriptions-item label="联系电话">{{ maskPhone(detail.contact_phone || detail.user_phone) }}</el-descriptions-item>
              <el-descriptions-item label="微信号">{{ detail.contact_wechat || '-' }}</el-descriptions-item>
              <el-descriptions-item label="用户ID">{{ detail.user_id }}</el-descriptions-item>
            </el-descriptions>
          </el-card>
        </el-col>

        <!-- 右侧：状态/统计/推广/拍卖/担保 -->
        <el-col :xs="24" :md="8">
          <el-card class="section-card">
            <template #header><span class="section-title">状态信息</span></template>
            <el-descriptions :column="1" border>
              <el-descriptions-item label="审核状态">
                <el-tag :type="auditTagType(detail.audit_status)" size="small">{{ auditText(detail.audit_status) }}</el-tag>
              </el-descriptions-item>
              <el-descriptions-item label="物品状态">
                <el-tag :type="statusTagType(detail.status)" size="small" effect="plain">{{ statusText(detail.status) }}</el-tag>
              </el-descriptions-item>
              <el-descriptions-item label="过期时间">{{ formatTime(detail.expiry_time) }}</el-descriptions-item>
              <el-descriptions-item label="发布时间">{{ formatTime(detail.published_at) }}</el-descriptions-item>
              <el-descriptions-item label="创建时间">{{ formatTime(detail.created_at) }}</el-descriptions-item>
              <el-descriptions-item label="更新时间">{{ formatTime(detail.updated_at) }}</el-descriptions-item>
            </el-descriptions>
          </el-card>

          <el-card class="section-card">
            <template #header><span class="section-title">交互数据</span></template>
            <div class="stats-grid">
              <div class="stats-item">
                <div class="stats-value">{{ detail.view_count || 0 }}</div>
                <div class="stats-label">浏览</div>
              </div>
              <div class="stats-item">
                <div class="stats-value">{{ detail.fav_count || 0 }}</div>
                <div class="stats-label">收藏</div>
              </div>
              <div class="stats-item">
                <div class="stats-value">{{ detail.message_count || 0 }}</div>
                <div class="stats-label">留言</div>
              </div>
            </div>
          </el-card>

          <el-card v-if="detail.auction" class="section-card">
            <template #header><span class="section-title">拍卖信息</span></template>
            <el-descriptions :column="1" border>
              <el-descriptions-item label="起拍价">¥{{ Number(detail.auction.start_price || 0).toFixed(2) }}</el-descriptions-item>
              <el-descriptions-item label="当前价">¥{{ Number(detail.auction.current_bid_price || 0).toFixed(2) }}</el-descriptions-item>
              <el-descriptions-item label="加价幅度">¥{{ Number(detail.auction.step_price || 0).toFixed(2) }}</el-descriptions-item>
              <el-descriptions-item label="保留价">¥{{ Number(detail.auction.reserve_price || 0).toFixed(2) }}</el-descriptions-item>
              <el-descriptions-item label="保证金">¥{{ Number(detail.auction.bond_amount || 0).toFixed(2) }}</el-descriptions-item>
              <el-descriptions-item label="出价次数">{{ detail.auction.bid_count }}</el-descriptions-item>
              <el-descriptions-item label="围观人数">{{ detail.auction.watcher_count }}</el-descriptions-item>
              <el-descriptions-item label="开始时间">{{ formatTime(detail.auction.start_time) }}</el-descriptions-item>
              <el-descriptions-item label="截拍时间">{{ formatTime(detail.auction.end_time) }}</el-descriptions-item>
              <el-descriptions-item label="状态">{{ detail.auction.status_text }}</el-descriptions-item>
            </el-descriptions>
          </el-card>

          <el-card v-if="detail.promotion" class="section-card">
            <template #header><span class="section-title">推广状态</span></template>
            <el-descriptions :column="1" border>
              <el-descriptions-item label="推广类型">{{ detail.promotion.promotion_type }}</el-descriptions-item>
              <el-descriptions-item label="状态">{{ detail.promotion.status_text }}</el-descriptions-item>
              <el-descriptions-item label="开始时间">{{ formatTime(detail.promotion.start_time) }}</el-descriptions-item>
              <el-descriptions-item label="结束时间">{{ formatTime(detail.promotion.end_time) }}</el-descriptions-item>
              <el-descriptions-item label="花费">¥{{ Number(detail.promotion.amount || 0).toFixed(2) }}</el-descriptions-item>
              <el-descriptions-item label="曝光">{{ detail.promotion.impression_count }}</el-descriptions-item>
              <el-descriptions-item label="点击">{{ detail.promotion.click_count }}</el-descriptions-item>
              <el-descriptions-item label="收藏">{{ detail.promotion.fav_count }}</el-descriptions-item>
              <el-descriptions-item label="咨询">{{ detail.promotion.consult_count }}</el-descriptions-item>
              <el-descriptions-item label="下单">{{ detail.promotion.order_count }}</el-descriptions-item>
              <el-descriptions-item label="ROI">{{ Number(detail.promotion.roi || 0).toFixed(2) }}</el-descriptions-item>
            </el-descriptions>
          </el-card>

          <el-card v-if="detail.escrow" class="section-card">
            <template #header><span class="section-title">担保交易</span></template>
            <el-descriptions :column="1" border>
              <el-descriptions-item label="担保金额">¥{{ Number(detail.escrow.escrow_amount || 0).toFixed(2) }}</el-descriptions-item>
              <el-descriptions-item label="平台手续费">¥{{ Number(detail.escrow.platform_fee || 0).toFixed(2) }}</el-descriptions-item>
              <el-descriptions-item label="卖家所得">¥{{ Number(detail.escrow.seller_amount || 0).toFixed(2) }}</el-descriptions-item>
              <el-descriptions-item label="状态">{{ detail.escrow.status_text }}</el-descriptions-item>
              <el-descriptions-item label="冻结时间">{{ formatTime(detail.escrow.frozen_at) }}</el-descriptions-item>
              <el-descriptions-item label="放款时间">{{ formatTime(detail.escrow.release_at) }}</el-descriptions-item>
              <el-descriptions-item label="自动放款">{{ formatTime(detail.escrow.auto_release_at) }}</el-descriptions-item>
            </el-descriptions>
          </el-card>

          <el-card v-if="detail.shop" class="section-card">
            <template #header><span class="section-title">所属店铺</span></template>
            <el-descriptions :column="1" border>
              <el-descriptions-item label="店铺名">{{ detail.shop.shop_name }}</el-descriptions-item>
              <el-descriptions-item label="等级">{{ detail.shop.level_text }}</el-descriptions-item>
              <el-descriptions-item label="粉丝数">{{ detail.shop.follower_count }}</el-descriptions-item>
              <el-descriptions-item label="商品数">{{ detail.shop.item_count }}</el-descriptions-item>
              <el-descriptions-item label="好评率">{{ (detail.shop.good_rate * 100).toFixed(1) }}%</el-descriptions-item>
            </el-descriptions>
          </el-card>
        </el-col>
      </el-row>
    </div>

    <!-- 底部操作 -->
    <div v-if="detail" class="action-bar">
      <el-button :icon="ArrowLeft" @click="goBack">返回</el-button>
      <el-button
        v-if="detail.audit_status === 0 || detail.audit_status === 2"
        type="success"
        @click="onAudit(1)"
      >审核通过</el-button>
      <el-button
        v-if="detail.audit_status === 0 || detail.audit_status === 1"
        type="danger"
        @click="onAudit(2)"
      >审核拒绝</el-button>
      <el-button v-if="detail.status === 1" type="warning" @click="onUpdateStatus(3)">下架</el-button>
      <el-button v-if="detail.status === 3" type="primary" @click="onUpdateStatus(1)">重新发布</el-button>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { ArrowLeft } from '@element-plus/icons-vue'
import { adminGetErshou, auditErshou, adminUpdateErshouStatus, listErshouSKUs } from '@/api/ershou'
import { getCategoryTree } from '@/api/category'
import { formatTime } from '@/utils/format'

const route = useRoute()
const router = useRouter()
const id = computed(() => route.params.id)

const loading = ref(false)
const detail = ref(null)
const skus = ref([])
const categoryMap = ref({})

const allImages = computed(() => {
  const arr = []
  if (detail.value?.cover_image) arr.push(detail.value.cover_image)
  if (Array.isArray(detail.value?.images)) arr.push(...detail.value.images)
  return arr
})

const categoryName = (cid) => categoryMap.value[cid] || '-'

const loadCategoryMap = async () => {
  try {
    const res = await getCategoryTree()
    const map = {}
    const walk = (nodes) => {
      if (!Array.isArray(nodes)) return
      nodes.forEach((n) => {
        map[n.id] = n.name
        if (n.children?.length) walk(n.children)
      })
    }
    walk(res.data || [])
    categoryMap.value = map
  } catch (e) {
    // ignore
  }
}

const loadDetail = async () => {
  loading.value = true
  try {
    const [dRes, skuRes] = await Promise.all([
      adminGetErshou(id.value),
      listErshouSKUs(id.value).catch(() => ({ data: { list: [] } }))
    ])
    detail.value = dRes.data || null
    skus.value = skuRes.data?.list || []
  } catch (e) {
    ElMessage.error('加载详情失败')
  } finally {
    loading.value = false
  }
}

const goBack = () => {
  router.push('/module-center/business/ershou/list')
}

// ===== 格式化 =====
const auditText = (s) => ({ 0: '待审核', 1: '已通过', 2: '已拒绝' }[s] || '-')
const auditTagType = (s) => ({ 0: 'warning', 1: 'success', 2: 'danger' }[s] || 'info')
const statusText = (s) => ({ 0: '草稿', 1: '已发布', 2: '已售出', 3: '已下架', 4: '已过期' }[s] || '-')
const statusTagType = (s) => ({ 0: 'info', 1: 'success', 2: 'primary', 3: 'warning', 4: 'danger' }[s] || 'info')
const conditionText = (c) => ({ new: '全新', almost_new: '几乎全新', used: '有使用痕迹', broken: '有瑕疵' }[c] || '-')
const deliveryText = (d) => ({ face: '当面交易', self: '自提', express: '快递' }[d] || '-')

const maskPhone = (p) => {
  if (!p) return '-'
  const s = String(p)
  if (s.length < 7) return s
  return s.slice(0, 3) + '****' + s.slice(-4)
}

const onAudit = async (auditStatus) => {
  try {
    if (auditStatus === 2) {
      const { value } = await ElMessageBox.prompt('请输入拒绝原因', '拒绝审核', {
        inputType: 'textarea',
        inputPlaceholder: '拒绝原因'
      })
      await auditErshou(id.value, { audit_status: auditStatus, audit_reason: value || '' })
    } else {
      await ElMessageBox.confirm('确定审核通过？', '提示', { type: 'warning' })
      await auditErshou(id.value, { audit_status: auditStatus })
    }
    ElMessage.success('操作成功')
    await loadDetail()
  } catch (e) {
    // 取消
  }
}

const onUpdateStatus = async (status) => {
  try {
    const label = statusText(status)
    await ElMessageBox.confirm(`确定将商品设为「${label}」吗？`, '提示', { type: 'warning' })
    await adminUpdateErshouStatus(id.value, status)
    ElMessage.success('操作成功')
    await loadDetail()
  } catch (e) {
    // 取消
  }
}

onMounted(async () => {
  await Promise.all([loadCategoryMap(), loadDetail()])
})
</script>

<style scoped>
.back-bar {
  display: flex; align-items: center; gap: 16px;
  margin-bottom: 16px;
}
.page-title { font-size: 18px; font-weight: 600; color: #303133; }
.section-card { margin-bottom: 16px; }
.section-title { font-weight: 600; color: #303133; }
.card-header-flex {
  display: flex; justify-content: space-between; align-items: center;
}
.images-grid { display: flex; flex-wrap: wrap; gap: 8px; }
.image-item {
  width: 120px; height: 120px; border-radius: 4px; border: 1px solid #ebeef5;
}
.content-box {
  white-space: pre-wrap; word-break: break-all;
  max-height: 300px; overflow-y: auto;
}
.stats-grid {
  display: grid; grid-template-columns: repeat(3, 1fr); gap: 12px;
}
.stats-item {
  text-align: center; padding: 12px;
  background: #fafafa; border-radius: 4px;
}
.stats-value { font-size: 24px; font-weight: 600; color: #409eff; }
.stats-label { font-size: 12px; color: #909399; margin-top: 4px; }
.empty-text { color: #909399; text-align: center; padding: 24px 0; }
.action-bar {
  margin-top: 16px; padding: 12px 16px;
  background: #fff; border-radius: 4px;
  display: flex; gap: 8px; flex-wrap: wrap;
}
.price { color: #f56c6c; font-weight: 600; font-size: 16px; }
.original-price { color: #c0c4cc; text-decoration: line-through; }
</style>
