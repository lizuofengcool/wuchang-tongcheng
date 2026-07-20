<template>
  <div class="app-container" v-loading="loading">
    <!-- 返回按钮 -->
    <div class="back-bar">
      <el-button :icon="ArrowLeft" @click="goBack">返回列表</el-button>
      <span class="page-title">会员详情 #{{ detail?.id || id }}</span>
    </div>

    <div v-if="detail" class="detail-wrap">
      <el-row :gutter="16">
        <!-- 左侧：基本资料 + 相册 + 关于我 -->
        <el-col :xs="24" :md="16">
          <el-card class="section-card">
            <template #header><span class="section-title">基本资料</span></template>
            <el-descriptions :column="3" border>
              <el-descriptions-item label="ID">{{ detail.id }}</el-descriptions-item>
              <el-descriptions-item label="昵称">{{ detail.nickname || '-' }}</el-descriptions-item>
              <el-descriptions-item label="用户ID">{{ detail.user_id }}</el-descriptions-item>
              <el-descriptions-item label="性别">{{ genderText(detail.gender) }}</el-descriptions-item>
              <el-descriptions-item label="年龄">{{ detail.age || '-' }}</el-descriptions-item>
              <el-descriptions-item label="生日">{{ formatTime(detail.birthday, 'YYYY-MM-DD') }}</el-descriptions-item>
              <el-descriptions-item label="身高">{{ detail.height || '-' }} cm</el-descriptions-item>
              <el-descriptions-item label="体重">{{ detail.weight || '-' }} kg</el-descriptions-item>
              <el-descriptions-item label="学历">{{ educationText(detail.education) }}</el-descriptions-item>
              <el-descriptions-item label="收入">{{ incomeText(detail.income) }}</el-descriptions-item>
              <el-descriptions-item label="职业">{{ detail.occupation || '-' }}</el-descriptions-item>
              <el-descriptions-item label="感情状态">{{ relationshipText(detail.relationship_status) }}</el-descriptions-item>
              <el-descriptions-item label="所在城市">{{ detail.location_city || '-' }}</el-descriptions-item>
              <el-descriptions-item label="地区ID">{{ detail.region_id || '-' }}</el-descriptions-item>
              <el-descriptions-item label="想找">{{ genderText(detail.looking_for_gender) }}</el-descriptions-item>
              <el-descriptions-item label="年龄要求">{{ detail.looking_for_age_min || '-' }}-{{ detail.looking_for_age_max || '-' }} 岁</el-descriptions-item>
              <el-descriptions-item v-if="detail.signature" label="个性签名" :span="3">{{ detail.signature }}</el-descriptions-item>
              <el-descriptions-item v-if="detail.about_me" label="关于我" :span="3">
                <div class="content-box">{{ detail.about_me }}</div>
              </el-descriptions-item>
              <el-descriptions-item v-if="detail.looking_for" label="期望对象" :span="3">
                <div class="content-box">{{ detail.looking_for }}</div>
              </el-descriptions-item>
              <el-descriptions-item v-if="detail.tags" label="标签" :span="3">
                <el-tag v-for="(t, i) in parseTags(detail.tags)" :key="i" size="small" class="tag-item">{{ t }}</el-tag>
              </el-descriptions-item>
            </el-descriptions>
          </el-card>

          <el-card class="section-card">
            <template #header>
              <span class="section-title">相册（{{ allImages.length }}）</span>
            </template>
            <div v-if="!allImages.length" class="empty-text">暂无照片</div>
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

          <el-card v-if="detail.voice_intro" class="section-card">
            <template #header><span class="section-title">语音介绍</span></template>
            <el-descriptions :column="2" border>
              <el-descriptions-item label="语音URL" :span="2">{{ detail.voice_intro }}</el-descriptions-item>
              <el-descriptions-item label="时长">{{ detail.voice_intro_duration || 0 }} 秒</el-descriptions-item>
            </el-descriptions>
          </el-card>
        </el-col>

        <!-- 右侧：状态/认证/统计 -->
        <el-col :xs="24" :md="8">
          <el-card class="section-card">
            <template #header><span class="section-title">状态信息</span></template>
            <el-descriptions :column="1" border>
              <el-descriptions-item label="审核状态">
                <el-tag :type="auditTagType(detail.audit_status)" size="small">{{ auditText(detail.audit_status) }}</el-tag>
              </el-descriptions-item>
              <el-descriptions-item label="会员状态">
                <el-tag :type="statusTagType(detail.status)" size="small" effect="plain">{{ statusText(detail.status) }}</el-tag>
              </el-descriptions-item>
              <el-descriptions-item label="精选">
                <el-tag v-if="detail.is_featured" type="danger" size="small">精选</el-tag>
                <span v-else>否</span>
              </el-descriptions-item>
              <el-descriptions-item label="推荐">
                <el-tag v-if="detail.is_picked" type="success" size="small">推荐</el-tag>
                <span v-else>否</span>
              </el-descriptions-item>
              <el-descriptions-item v-if="detail.audit_reason" label="审核原因">{{ detail.audit_reason }}</el-descriptions-item>
              <el-descriptions-item label="注册时间">{{ formatTime(detail.created_at) }}</el-descriptions-item>
              <el-descriptions-item label="更新时间">{{ formatTime(detail.updated_at) }}</el-descriptions-item>
            </el-descriptions>
          </el-card>

          <el-card class="section-card">
            <template #header><span class="section-title">认证信息</span></template>
            <el-descriptions :column="1" border>
              <el-descriptions-item label="实名认证">
                <el-tag v-if="detail.id_verified" type="success" size="small">已认证</el-tag>
                <el-tag v-else type="info" size="small">未认证</el-tag>
              </el-descriptions-item>
              <el-descriptions-item label="视频认证">
                <el-tag v-if="detail.video_verified" type="success" size="small">已认证</el-tag>
                <el-tag v-else type="info" size="small">未认证</el-tag>
              </el-descriptions-item>
              <el-descriptions-item label="手机认证">
                <el-tag v-if="detail.phone_verified" type="success" size="small">已认证</el-tag>
                <el-tag v-else type="info" size="small">未认证</el-tag>
              </el-descriptions-item>
              <el-descriptions-item label="房产认证">
                <el-tag v-if="detail.house_verified" type="success" size="small">已认证</el-tag>
                <el-tag v-else type="info" size="small">未认证</el-tag>
              </el-descriptions-item>
              <el-descriptions-item label="车辆认证">
                <el-tag v-if="detail.car_verified" type="success" size="small">已认证</el-tag>
                <el-tag v-else type="info" size="small">未认证</el-tag>
              </el-descriptions-item>
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
                <div class="stats-value">{{ detail.like_count || 0 }}</div>
                <div class="stats-label">喜欢</div>
              </div>
              <div class="stats-item">
                <div class="stats-value">{{ detail.match_count || 0 }}</div>
                <div class="stats-label">匹配</div>
              </div>
              <div class="stats-item">
                <div class="stats-value">{{ detail.visitor_count || 0 }}</div>
                <div class="stats-label">访客</div>
              </div>
            </div>
          </el-card>

          <el-card v-if="detail.profile" class="section-card">
            <template #header><span class="section-title">详细资料</span></template>
            <el-descriptions :column="1" border>
              <el-descriptions-item label="籍贯">{{ detail.profile.hometown || '-' }}</el-descriptions-item>
              <el-descriptions-item label="行业">{{ detail.profile.industry || '-' }}</el-descriptions-item>
              <el-descriptions-item label="公司">{{ detail.profile.company || '-' }}</el-descriptions-item>
              <el-descriptions-item label="房产情况">{{ detail.profile.house_status || '-' }}</el-descriptions-item>
              <el-descriptions-item label="车辆情况">{{ detail.profile.car_status || '-' }}</el-descriptions-item>
              <el-descriptions-item label="是否吸烟">{{ profileBool(detail.profile.smoking) }}</el-descriptions-item>
              <el-descriptions-item label="是否饮酒">{{ profileBool(detail.profile.drinking) }}</el-descriptions-item>
              <el-descriptions-item label="是否要小孩">{{ profileBool(detail.profile.want_children) }}</el-descriptions-item>
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
      <el-button v-if="detail.status === 3" type="primary" @click="onUpdateStatus(1)">恢复正常</el-button>
      <el-button v-if="detail.status !== 0" type="danger" @click="onUpdateStatus(0)">禁用</el-button>
      <el-button type="info" @click="onToggleFeatured">{{ detail.is_featured ? '取消精选' : '设为精选' }}</el-button>
      <el-button type="info" @click="onTogglePicked">{{ detail.is_picked ? '取消推荐' : '设为推荐' }}</el-button>
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

const allImages = computed(() => {
  if (!detail.value) return []
  const arr = []
  if (detail.value.avatar) arr.push(detail.value.avatar)
  if (Array.isArray(detail.value.photos)) arr.push(...detail.value.photos)
  return arr
})

const parseTags = (tags) => {
  if (!tags) return []
  if (Array.isArray(tags)) return tags
  try { return JSON.parse(tags) || [] } catch (e) { return String(tags).split(/[,，]/).filter(Boolean) }
}

const profileBool = (v) => {
  if (v === true || v === 1) return '是'
  if (v === false || v === 0) return '否'
  return '-'
}

// ===== 格式化 =====
const auditText = (s) => ({ 0: '待审核', 1: '已通过', 2: '已拒绝' }[s] || '-')
const auditTagType = (s) => ({ 0: 'warning', 1: 'success', 2: 'danger' }[s] || 'info')
const statusText = (s) => ({ 0: '禁用', 1: '正常', 3: '下架' }[s] || '-')
const statusTagType = (s) => ({ 0: 'danger', 1: 'success', 3: 'warning' }[s] || 'info')
const genderText = (g) => ({ male: '男', female: '女' }[g] || '未知')
const educationMap = { high_school: '高中', college: '大专', bachelor: '本科', master: '硕士', doctor: '博士' }
const educationText = (e) => educationMap[e] || e || '-'
const incomeMap = { low: '10万以下', medium: '10-30万', high: '30-50万', very_high: '50万以上' }
const incomeText = (i) => incomeMap[i] || i || '-'
const relationshipMap = { single: '单身', divorced: '离异', widowed: '丧偶' }
const relationshipText = (r) => relationshipMap[r] || r || '-'

const loadDetail = async () => {
  loading.value = true
  try {
    const res = await request.get(`/love/admin/loves/${id.value}`)
    detail.value = res.data || null
  } catch (e) {
    ElMessage.error('加载详情失败')
  } finally {
    loading.value = false
  }
}

const goBack = () => {
  router.push('/business/love/list')
}

const onAudit = async (auditStatus) => {
  try {
    if (auditStatus === 2) {
      const { value } = await ElMessageBox.prompt('请输入拒绝原因', '拒绝审核', {
        inputType: 'textarea',
        inputPlaceholder: '拒绝原因'
      })
      await request.put(`/love/admin/loves/${id.value}/audit`, { audit_status: auditStatus, audit_reason: value || '' })
    } else {
      await ElMessageBox.confirm('确定审核通过？', '提示', { type: 'warning' })
      await request.put(`/love/admin/loves/${id.value}/audit`, { audit_status: auditStatus })
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
    await ElMessageBox.confirm(`确定将会员设为「${label}」吗？`, '提示', { type: 'warning' })
    await request.put(`/love/admin/loves/${id.value}/status`, { status })
    ElMessage.success('操作成功')
    await loadDetail()
  } catch (e) {
    // 取消
  }
}

const onToggleFeatured = async () => {
  try {
    await request.put(`/love/admin/loves/${id.value}/featured`, { is_featured: !detail.value.is_featured })
    ElMessage.success('操作成功')
    await loadDetail()
  } catch (e) { /* ignore */ }
}

const onTogglePicked = async () => {
  try {
    await request.put(`/love/admin/loves/${id.value}/picked`, { is_picked: !detail.value.is_picked })
    ElMessage.success('操作成功')
    await loadDetail()
  } catch (e) { /* ignore */ }
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
  display: grid; grid-template-columns: repeat(2, 1fr); gap: 12px;
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
.tag-item { margin-right: 4px; margin-bottom: 4px; }
</style>
