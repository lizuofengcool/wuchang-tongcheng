<template>
  <div class="app-container" v-loading="loading">
    <div class="back-bar">
      <el-button :icon="ArrowLeft" @click="goBack">返回列表</el-button>
      <span class="page-title">拼车详情 #{{ detail?.id || id }}</span>
    </div>

    <div v-if="detail" class="detail-wrap">
      <el-row :gutter="16">
        <!-- 左侧：基本信息 -->
        <el-col :xs="24" :md="16">
          <el-card class="section-card">
            <template #header><span class="section-title">基本信息</span></template>
            <el-descriptions :column="3" border>
              <el-descriptions-item label="出发地" :span="3">{{ detail.origin }}</el-descriptions-item>
              <el-descriptions-item label="目的地" :span="3">{{ detail.destination }}</el-descriptions-item>
              <el-descriptions-item label="途经" :span="3">
                <span v-if="detail.waypoints && detail.waypoints.length">{{ detail.waypoints.join(' → ') }}</span>
                <span v-else class="text-muted">-</span>
              </el-descriptions-item>
              <el-descriptions-item label="拼车类型">
                <el-tag :type="typeTagType(detail.type)" size="small" effect="plain">{{ typeText(detail.type) }}</el-tag>
              </el-descriptions-item>
              <el-descriptions-item label="价格">
                <span class="price">¥{{ Number(detail.price || 0).toFixed(2) }}</span>
                <span class="price-unit">/ {{ detail.price_unit || '座' }}</span>
              </el-descriptions-item>
              <el-descriptions-item label="加急">
                <el-tag v-if="detail.is_urgent" type="danger" size="small">急</el-tag>
                <span v-else>否</span>
              </el-descriptions-item>
              <el-descriptions-item label="总座位">{{ detail.seats_total }}</el-descriptions-item>
              <el-descriptions-item label="已预订">{{ detail.seats_taken }}</el-descriptions-item>
              <el-descriptions-item label="剩余空位">{{ detail.seats_total - detail.seats_taken }}</el-descriptions-item>
              <el-descriptions-item label="出发时间">{{ formatTime(detail.departure_time) }}</el-descriptions-item>
              <el-descriptions-item label="集合地点">{{ detail.meeting_point || '-' }}</el-descriptions-item>
              <el-descriptions-item label="车型">{{ detail.vehicle_type || '-' }}</el-descriptions-item>
              <el-descriptions-item v-if="detail.summary" label="摘要" :span="3">{{ detail.summary }}</el-descriptions-item>
              <el-descriptions-item v-if="detail.content" label="详细描述" :span="3">
                <div class="content-box">{{ detail.content }}</div>
              </el-descriptions-item>
              <el-descriptions-item v-if="detail.audit_reason" label="审核原因" :span="3">{{ detail.audit_reason }}</el-descriptions-item>
            </el-descriptions>
          </el-card>

          <el-card class="section-card">
            <template #header><span class="section-title">行程路线</span></template>
            <el-descriptions :column="2" border>
              <el-descriptions-item label="出发地经纬度">{{ detail.origin_lng || '-' }}, {{ detail.origin_lat || '-' }}</el-descriptions-item>
              <el-descriptions-item label="目的地经纬度">{{ detail.dest_lng || '-' }}, {{ detail.dest_lat || '-' }}</el-descriptions-item>
              <el-descriptions-item label="预估距离">{{ detail.estimated_distance ? detail.estimated_distance + ' km' : '-' }}</el-descriptions-item>
              <el-descriptions-item label="预估时长">{{ detail.estimated_duration ? detail.estimated_duration + ' 分钟' : '-' }}</el-descriptions-item>
              <el-descriptions-item label="高速费">{{ detail.has_toll ? '是' : '否' }}</el-descriptions-item>
              <el-descriptions-item label="拼车规则">{{ detail.share_rule || '-' }}</el-descriptions-item>
            </el-descriptions>
          </el-card>

          <el-card class="section-card">
            <template #header><span class="section-title">联系方式</span></template>
            <el-descriptions :column="2" border>
              <el-descriptions-item label="联系人">{{ detail.user_name || `用户#${detail.user_id}` }}</el-descriptions-item>
              <el-descriptions-item label="联系电话">{{ maskPhone(detail.contact_phone || detail.user_phone) }}</el-descriptions-item>
              <el-descriptions-item label="微信号">{{ detail.contact_wechat || '-' }}</el-descriptions-item>
              <el-descriptions-item label="用户ID">{{ detail.user_id }}</el-descriptions-item>
            </el-descriptions>
          </el-card>
        </el-col>

        <!-- 右侧：状态/统计 -->
        <el-col :xs="24" :md="8">
          <el-card class="section-card">
            <template #header><span class="section-title">状态信息</span></template>
            <el-descriptions :column="1" border>
              <el-descriptions-item label="审核状态">
                <el-tag :type="auditTagType(detail.audit_status)" size="small">{{ auditText(detail.audit_status) }}</el-tag>
              </el-descriptions-item>
              <el-descriptions-item label="拼车状态">
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
                <div class="stats-value">{{ detail.booking_count || 0 }}</div>
                <div class="stats-label">预订</div>
              </div>
            </div>
          </el-card>

          <el-card class="section-card">
            <template #header><span class="section-title">关联信息</span></template>
            <el-descriptions :column="1" border>
              <el-descriptions-item label="行程ID">{{ detail.trip_id || '-' }}</el-descriptions-item>
              <el-descriptions-item label="车主ID">{{ detail.driver_id || '-' }}</el-descriptions-item>
              <el-descriptions-item label="车辆ID">{{ detail.vehicle_id || '-' }}</el-descriptions-item>
              <el-descriptions-item label="地区ID">{{ detail.region_id || '-' }}</el-descriptions-item>
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
import request from '@/utils/request'
import { formatTime } from '@/utils/format'

const route = useRoute()
const router = useRouter()
const id = computed(() => route.params.id)

const loading = ref(false)
const detail = ref(null)

const loadDetail = async () => {
  loading.value = true
  try {
    const res = await request.get(`/pinche/admin/pinches/${id.value}`)
    detail.value = res.data || null
  } catch (e) {
    ElMessage.error('加载详情失败')
  } finally {
    loading.value = false
  }
}

const goBack = () => {
  router.push('/business/pinche/list')
}

// ===== 格式化 =====
const auditText = (s) => ({ 0: '待审核', 1: '已通过', 2: '已拒绝' }[s] || '-')
const auditTagType = (s) => ({ 0: 'warning', 1: 'success', 2: 'danger' }[s] || 'info')
const statusText = (s) => ({ 0: '草稿', 1: '已发布', 2: '已满员', 3: '已下架', 4: '已过期', 5: '已取消' }[s] || '-')
const statusTagType = (s) => ({ 0: 'info', 1: 'success', 2: 'primary', 3: 'warning', 4: 'danger', 5: 'info' }[s] || 'info')
const typeText = (t) => ({ passenger: '人找车', driver: '车找人', cargo: '车找货', cargo_need: '货找车' }[t] || '-')
const typeTagType = (t) => ({ passenger: 'primary', driver: 'success', cargo: 'warning', cargo_need: 'danger' }[t] || 'info')

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
      await request.put(`/pinche/admin/pinches/${id.value}/audit`, { audit_status: auditStatus, audit_reason: value || '' })
    } else {
      await ElMessageBox.confirm('确定审核通过？', '提示', { type: 'warning' })
      await request.put(`/pinche/admin/pinches/${id.value}/audit`, { audit_status: auditStatus })
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
    await ElMessageBox.confirm(`确定将拼车设为「${label}」吗？`, '提示', { type: 'warning' })
    await request.put(`/pinche/admin/pinches/${id.value}/status`, { status })
    ElMessage.success('操作成功')
    await loadDetail()
  } catch (e) {
    // 取消
  }
}

onMounted(() => loadDetail())
</script>

<style scoped>
.back-bar {
  display: flex; align-items: center; gap: 16px;
  margin-bottom: 16px;
}
.page-title { font-size: 18px; font-weight: 600; color: #303133; }
.section-card { margin-bottom: 16px; }
.section-title { font-weight: 600; color: #303133; }
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
.text-muted { color: #909399; }
.action-bar {
  margin-top: 16px; padding: 12px 16px;
  background: #fff; border-radius: 4px;
  display: flex; gap: 8px; flex-wrap: wrap;
}
.price { color: #f56c6c; font-weight: 600; font-size: 16px; }
.price-unit { color: #909399; font-size: 12px; margin-left: 4px; }
</style>
