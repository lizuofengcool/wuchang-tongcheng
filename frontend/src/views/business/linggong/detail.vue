<template>
  <div class="app-container" v-loading="loading">
    <div class="back-bar">
      <el-button :icon="ArrowLeft" @click="goBack">返回列表</el-button>
      <span class="page-title">零工岗位详情 #{{ detail?.id || id }}</span>
    </div>

    <div v-if="detail" class="detail-wrap">
      <el-row :gutter="16">
        <!-- 左侧：基本信息 + 图集 -->
        <el-col :xs="24" :md="16">
          <el-card class="section-card">
            <template #header><span class="section-title">基本信息</span></template>
            <el-descriptions :column="3" border>
              <el-descriptions-item label="标题" :span="3">{{ detail.title }}</el-descriptions-item>
              <el-descriptions-item label="岗位类型">
                <el-tag size="small" :type="typeTagType(detail.linggong_type)">{{ typeMap[detail.linggong_type] || detail.linggong_type }}</el-tag>
              </el-descriptions-item>
              <el-descriptions-item label="发布者类型">{{ publisherMap[detail.publisher_type] || detail.publisher_type }}</el-descriptions-item>
              <el-descriptions-item label="公司名称">{{ detail.company_name || '-' }}</el-descriptions-item>
              <el-descriptions-item label="联系人">{{ detail.contact_name || detail.user_name || '-' }}</el-descriptions-item>
              <el-descriptions-item label="联系电话">{{ maskPhone(detail.contact_phone || detail.user_phone) }}</el-descriptions-item>
              <el-descriptions-item label="微信号">{{ detail.contact_wechat || '-' }}</el-descriptions-item>
              <el-descriptions-item label="计费方式">
                <el-tag size="small">{{ billingMap[detail.billing_type] || detail.billing_type }}</el-tag>
              </el-descriptions-item>
              <el-descriptions-item label="薪资">
                <span class="price">¥{{ formatSalary(detail) }}</span>
                <span class="price-unit">{{ detail.salary_unit }}</span>
              </el-descriptions-item>
              <el-descriptions-item label="结算周期">{{ settlementMap[detail.settlement] || detail.settlement }}</el-descriptions-item>
              <el-descriptions-item label="招募人数">{{ detail.recruit_count }}</el-descriptions-item>
              <el-descriptions-item label="已报名">{{ detail.applied_count }}</el-descriptions-item>
              <el-descriptions-item label="已确认">{{ detail.confirmed_count }}</el-descriptions-item>
              <el-descriptions-item label="工作日">{{ detail.work_days }} 天</el-descriptions-item>
              <el-descriptions-item label="每日工时">{{ detail.work_hours }} 小时</el-descriptions-item>
              <el-descriptions-item label="工作时间">{{ detail.work_time_start || '-' }} - {{ detail.work_time_end || '-' }}</el-descriptions-item>
              <el-descriptions-item label="性别要求">{{ genderText(detail.need_gender) }}</el-descriptions-item>
              <el-descriptions-item label="年龄要求">{{ detail.min_age || 0 }} - {{ detail.max_age || 0 }}</el-descriptions-item>
              <el-descriptions-item label="学历要求">{{ detail.education || '-' }}</el-descriptions-item>
              <el-descriptions-item label="经验要求">{{ detail.experience || '-' }}</el-descriptions-item>
              <el-descriptions-item label="需健康证">{{ detail.need_health_cert ? '是' : '否' }}</el-descriptions-item>
              <el-descriptions-item label="需身份证">{{ detail.need_id_card ? '是' : '否' }}</el-descriptions-item>
              <el-descriptions-item label="最低信用分">{{ detail.min_credit_score || 0 }}</el-descriptions-item>
              <el-descriptions-item label="工作地点" :span="3">{{ formatAddress(detail) }}</el-descriptions-item>
              <el-descriptions-item label="工作方式">{{ workLocationText(detail.work_location_type) }}</el-descriptions-item>
              <el-descriptions-item label="工作强度">{{ intensityText(detail.work_intensity) }}</el-descriptions-item>
              <el-descriptions-item label="审核状态">
                <el-tag :type="auditTagType(detail.audit_status)" size="small">{{ auditText(detail.audit_status) }}</el-tag>
              </el-descriptions-item>
              <el-descriptions-item label="岗位状态">
                <el-tag :type="statusTagType(detail.status)" size="small" effect="plain">{{ statusMap[detail.status] || '-' }}</el-tag>
              </el-descriptions-item>
              <el-descriptions-item v-if="detail.audit_reason" label="审核原因" :span="3">{{ detail.audit_reason }}</el-descriptions-item>
              <el-descriptions-item v-if="detail.content" label="详细描述" :span="3">
                <div class="content-box">{{ detail.content }}</div>
              </el-descriptions-item>
            </el-descriptions>
          </el-card>

          <el-card class="section-card">
            <template #header><span class="section-title">图集（{{ allImages.length }}）</span></template>
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
        </el-col>

        <!-- 右侧：状态/统计 -->
        <el-col :xs="24" :md="8">
          <el-card class="section-card">
            <template #header><span class="section-title">状态信息</span></template>
            <el-descriptions :column="1" border>
              <el-descriptions-item label="审核状态">
                <el-tag :type="auditTagType(detail.audit_status)" size="small">{{ auditText(detail.audit_status) }}</el-tag>
              </el-descriptions-item>
              <el-descriptions-item label="岗位状态">
                <el-tag :type="statusTagType(detail.status)" size="small" effect="plain">{{ statusMap[detail.status] || '-' }}</el-tag>
              </el-descriptions-item>
              <el-descriptions-item label="精选">{{ detail.featured ? '是' : '否' }}</el-descriptions-item>
              <el-descriptions-item label="运营甄选">{{ detail.picked ? '是' : '否' }}</el-descriptions-item>
              <el-descriptions-item label="官方验真">{{ detail.verified ? '是' : '否' }}</el-descriptions-item>
              <el-descriptions-item label="推广等级">{{ detail.promotion_level || 0 }}</el-descriptions-item>
              <el-descriptions-item label="流量权重">{{ Number(detail.traffic_weight || 1).toFixed(2) }}</el-descriptions-item>
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
                <div class="stats-value">{{ detail.contact_count || 0 }}</div>
                <div class="stats-label">联系</div>
              </div>
              <div class="stats-item">
                <div class="stats-value">{{ detail.share_count || 0 }}</div>
                <div class="stats-label">分享</div>
              </div>
              <div class="stats-item">
                <div class="stats-value">{{ detail.application_count || 0 }}</div>
                <div class="stats-label">报名</div>
              </div>
              <div class="stats-item">
                <div class="stats-value">{{ detail.risk_score || 0 }}</div>
                <div class="stats-label">风险分</div>
              </div>
            </div>
          </el-card>

          <el-card v-if="detail.employer_verified" class="section-card">
            <template #header><span class="section-title">雇主认证</span></template>
            <el-descriptions :column="1" border>
              <el-descriptions-item label="雇主ID">{{ detail.employer_id }}</el-descriptions-item>
              <el-descriptions-item label="认证时间">{{ formatTime(detail.employer_verified_at) }}</el-descriptions-item>
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
      <el-button v-if="detail.status === 1" type="warning" @click="onUpdateStatus(2)">下架</el-button>
      <el-button v-if="detail.status === 2" type="primary" @click="onUpdateStatus(1)">重新发布</el-button>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { ArrowLeft } from '@element-plus/icons-vue'
import { adminGetLinggong, auditLinggong, adminUpdateLinggongStatus } from '@/api/linggong'
import { formatTime } from '@/utils/format'

const route = useRoute()
const router = useRouter()
const id = computed(() => route.params.id)

const loading = ref(false)
const detail = ref(null)

const allImages = computed(() => {
  const arr = []
  if (detail.value?.cover_image) arr.push(detail.value.cover_image)
  if (Array.isArray(detail.value?.images)) arr.push(...detail.value.images)
  return arr
})

const loadDetail = async () => {
  loading.value = true
  try {
    const dRes = await adminGetLinggong(id.value)
    detail.value = dRes.data || null
  } catch (e) {
    ElMessage.error('加载详情失败')
  } finally {
    loading.value = false
  }
}

const goBack = () => {
  router.push('/business/linggong/list')
}

// ===== 格式化 =====
const typeMap = {
  short_term: '短期', long_term: '长期', task: '任务制',
  hourly: '小时工', daily: '日结', temp: '临时工'
}
const typeTagType = (t) => ({
  short_term: 'primary', long_term: 'success', task: 'warning',
  hourly: 'info', daily: 'danger', temp: 'info'
}[t] || 'info')

const publisherMap = { personal: '个人', company: '企业', agent: '中介', headhunter: '猎头' }
const billingMap = {
  by_piece: '按件', by_hour: '按时', by_day: '按日', by_week: '按周',
  by_month: '按月', fixed: '固定', negotiable: '面议'
}
const settlementMap = {
  'T+0': '当日结', 'T+1': '次日结', 'T+3': '三日结',
  'T+7': '周结', 'M+1': '月结', project: '项目结'
}
const statusMap = {
  0: '草稿', 1: '已发布', 2: '已下架', 3: '已过期',
  4: '已删除', 5: '已满员', 6: '已关闭', 7: '已完成'
}
const statusTagType = (s) => ({
  0: 'info', 1: 'success', 2: 'warning', 3: 'danger',
  4: 'info', 5: 'primary', 6: 'info', 7: 'success'
}[s] || 'info')
const auditText = (s) => ({ 0: '待审核', 1: '已通过', 2: '已拒绝' }[s] || '-')
const auditTagType = (s) => ({ 0: 'warning', 1: 'success', 2: 'danger' }[s] || 'info')
const genderText = (g) => ({ any: '不限', male: '男', female: '女' }[g] || '不限')
const workLocationText = (w) => ({ onsite: '现场', remote: '远程', hybrid: '混合' }[w] || '-')
const intensityText = (i) => ({ light: '轻松', medium: '中等', heavy: '繁重', extreme: '极重' }[i] || '-')

const formatSalary = (row) => {
  if (!row) return '0'
  if (row.salary_negotiable) return '面议'
  if (row.salary_min && row.salary_max) return `${Number(row.salary_min).toFixed(0)}-${Number(row.salary_max).toFixed(0)}`
  if (row.salary_min) return Number(row.salary_min).toFixed(0)
  if (row.salary_max) return Number(row.salary_max).toFixed(0)
  return '0'
}

const formatAddress = (row) => {
  if (!row) return '-'
  return [row.province, row.city, row.district, row.business_district, row.address].filter(Boolean).join(' ') || '-'
}

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
      await auditLinggong(id.value, { audit_status: auditStatus, audit_reason: value || '' })
    } else {
      await ElMessageBox.confirm('确定审核通过？', '提示', { type: 'warning' })
      await auditLinggong(id.value, { audit_status: auditStatus })
    }
    ElMessage.success('操作成功')
    await loadDetail()
  } catch (e) {
    // 取消
  }
}

const onUpdateStatus = async (status) => {
  try {
    const label = statusMap[status]
    await ElMessageBox.confirm(`确定将岗位设为「${label}」吗？`, '提示', { type: 'warning' })
    await adminUpdateLinggongStatus(id.value, status)
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
.price-unit { color: #909399; margin-left: 4px; }
</style>
