<template>
  <div class="app-container" v-loading="loading">
    <div class="back-bar">
      <el-button :icon="ArrowLeft" @click="goBack">返回列表</el-button>
      <span class="page-title">店铺详情 #{{ detail?.id || id }}</span>
    </div>

    <div v-if="detail" class="detail-wrap">
      <el-row :gutter="16">
        <el-col :xs="24" :md="16">
          <el-card class="section-card">
            <template #header><span class="section-title">基本信息</span></template>
            <el-descriptions :column="3" border>
              <el-descriptions-item label="店铺名" :span="3">{{ detail.name }}</el-descriptions-item>
              <el-descriptions-item label="简称">{{ detail.short_name || '-' }}</el-descriptions-item>
              <el-descriptions-item label="店主">{{ detail.owner_name || '-' }}</el-descriptions-item>
              <el-descriptions-item label="联系电话">{{ detail.contact_phone || '-' }}</el-descriptions-item>
              <el-descriptions-item label="分类">{{ detail.category_name || detail.category_id || '-' }}</el-descriptions-item>
              <el-descriptions-item label="地区">{{ detail.region_id || '-' }}</el-descriptions-item>
              <el-descriptions-item label="用户ID">{{ detail.user_id || '-' }}</el-descriptions-item>
              <el-descriptions-item label="地址" :span="3">{{ detail.address || '-' }}</el-descriptions-item>
              <el-descriptions-item label="经度">{{ detail.longitude || '-' }}</el-descriptions-item>
              <el-descriptions-item label="纬度">{{ detail.latitude || '-' }}</el-descriptions-item>
              <el-descriptions-item label="排序">{{ detail.sort || 0 }}</el-descriptions-item>
              <el-descriptions-item v-if="detail.summary" label="简介" :span="3">{{ detail.summary }}</el-descriptions-item>
              <el-descriptions-item v-if="detail.description" label="详细描述" :span="3">
                <div class="content-box">{{ detail.description }}</div>
              </el-descriptions-item>
              <el-descriptions-item v-if="detail.audit_reason" label="审核原因" :span="3">{{ detail.audit_reason }}</el-descriptions-item>
            </el-descriptions>
          </el-card>

          <el-card v-if="detail.images && detail.images.length" class="section-card">
            <template #header><span class="section-title">店铺图集（{{ detail.images.length }}）</span></template>
            <div class="images-grid">
              <el-image v-for="(img, idx) in detail.images" :key="idx" :src="img" fit="cover" class="image-item" :preview-src-list="detail.images" :initial-index="idx" preview-teleported />
            </div>
          </el-card>
        </el-col>

        <el-col :xs="24" :md="8">
          <el-card class="section-card">
            <template #header><span class="section-title">状态信息</span></template>
            <el-descriptions :column="1" border>
              <el-descriptions-item label="审核状态">
                <el-tag :type="auditTagType(detail.audit_status)" size="small">{{ auditText(detail.audit_status) }}</el-tag>
              </el-descriptions-item>
              <el-descriptions-item label="店铺状态">
                <el-tag :type="statusTagType(detail.status)" size="small" effect="plain">{{ statusText(detail.status) }}</el-tag>
              </el-descriptions-item>
              <el-descriptions-item label="是否推荐">
                <el-tag v-if="detail.is_recommended" type="warning" size="small">推荐</el-tag>
                <span v-else>否</span>
              </el-descriptions-item>
              <el-descriptions-item label="是否置顶">
                <el-tag v-if="detail.is_pinned" type="danger" size="small">置顶</el-tag>
                <span v-else>否</span>
              </el-descriptions-item>
              <el-descriptions-item label="推广权重">{{ detail.promotion_weight || 0 }}</el-descriptions-item>
              <el-descriptions-item label="创建时间">{{ formatTime(detail.created_at) }}</el-descriptions-item>
              <el-descriptions-item label="更新时间">{{ formatTime(detail.updated_at) }}</el-descriptions-item>
            </el-descriptions>
          </el-card>

          <el-card class="section-card">
            <template #header><span class="section-title">交互数据</span></template>
            <div class="stats-grid">
              <div class="stats-item"><div class="stats-value">{{ detail.view_count || 0 }}</div><div class="stats-label">浏览</div></div>
              <div class="stats-item"><div class="stats-value">{{ detail.fav_count || 0 }}</div><div class="stats-label">收藏</div></div>
              <div class="stats-item"><div class="stats-value">{{ detail.product_count || 0 }}</div><div class="stats-label">商品数</div></div>
              <div class="stats-item"><div class="stats-value">{{ detail.sale_count || 0 }}</div><div class="stats-label">销量</div></div>
              <div class="stats-item"><div class="stats-value">{{ detail.review_count || 0 }}</div><div class="stats-label">评价</div></div>
              <div class="stats-item"><div class="stats-value">{{ Number(detail.rating || 0).toFixed(1) }}</div><div class="stats-label">评分</div></div>
            </div>
          </el-card>
        </el-col>
      </el-row>
    </div>

    <div v-if="detail" class="action-bar">
      <el-button :icon="ArrowLeft" @click="goBack">返回</el-button>
      <el-button v-if="detail.audit_status === 0 || detail.audit_status === 2" type="success" @click="onAudit(1)">审核通过</el-button>
      <el-button v-if="detail.audit_status === 0 || detail.audit_status === 1" type="danger" @click="onAudit(2)">审核拒绝</el-button>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { ArrowLeft } from '@element-plus/icons-vue'
import { formatTime } from '@/utils/format'
import { getMallShopDetail, auditMallShop } from '@/api/mall'

const route = useRoute()
const router = useRouter()
const id = computed(() => route.params.id)

const loading = ref(false)
const detail = ref(null)

const loadDetail = async () => {
  loading.value = true
  try {
    const res = await getMallShopDetail(id.value)
    detail.value = res.data || null
  } catch (e) {
    ElMessage.error('加载详情失败')
  } finally {
    loading.value = false
  }
}

const goBack = () => router.push('/business/mall/shop')

const auditText = (s) => ({ 0: '待审核', 1: '已通过', 2: '已拒绝' }[s] || '-')
const auditTagType = (s) => ({ 0: 'warning', 1: 'success', 2: 'danger' }[s] || 'info')
const statusText = (s) => ({ 0: '关闭', 1: '营业中', 2: '休息中' }[s] || '-')
const statusTagType = (s) => ({ 0: 'danger', 1: 'success', 2: 'warning' }[s] || 'info')

const onAudit = async (auditStatus) => {
  try {
    if (auditStatus === 2) {
      const { value } = await ElMessageBox.prompt('请输入拒绝原因', '拒绝审核', { inputType: 'textarea', inputPlaceholder: '拒绝原因' })
      await auditMallShop(id.value, { audit_status: auditStatus, audit_reason: value || '' })
    } else {
      await ElMessageBox.confirm('确定审核通过？', '提示', { type: 'warning' })
      await auditMallShop(id.value, { audit_status: auditStatus })
    }
    ElMessage.success('操作成功')
    await loadDetail()
  } catch (e) { /* cancel */ }
}

onMounted(() => loadDetail())
</script>

<style scoped>
.back-bar { display: flex; align-items: center; gap: 16px; margin-bottom: 16px; }
.page-title { font-size: 18px; font-weight: 600; color: #303133; }
.section-card { margin-bottom: 16px; }
.section-title { font-weight: 600; color: #303133; }
.content-box { white-space: pre-wrap; word-break: break-all; max-height: 300px; overflow-y: auto; }
.images-grid { display: flex; flex-wrap: wrap; gap: 8px; }
.image-item { width: 120px; height: 120px; border-radius: 4px; border: 1px solid #ebeef5; }
.stats-grid { display: grid; grid-template-columns: repeat(3, 1fr); gap: 12px; }
.stats-item { text-align: center; padding: 12px; background: #fafafa; border-radius: 4px; }
.stats-value { font-size: 22px; font-weight: 600; color: #409eff; }
.stats-label { font-size: 12px; color: #909399; margin-top: 4px; }
.action-bar { margin-top: 16px; padding: 12px 16px; background: #fff; border-radius: 4px; display: flex; gap: 8px; flex-wrap: wrap; }
</style>
