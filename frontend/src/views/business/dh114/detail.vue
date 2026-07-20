<template>
  <div class="app-container" v-loading="loading">
    <!-- 返回按钮 -->
    <div class="back-bar">
      <el-button :icon="ArrowLeft" @click="goBack">返回列表</el-button>
      <span class="page-title">商户详情 #{{ detail?.id || id }}</span>
    </div>

    <div v-if="detail" class="detail-wrap">
      <el-row :gutter="16">
        <!-- 左侧 -->
        <el-col :xs="24" :md="16">
          <el-card class="section-card">
            <template #header><span class="section-title">基本信息</span></template>
            <el-descriptions :column="3" border>
              <el-descriptions-item label="商户名" :span="3">{{ detail.name }}</el-descriptions-item>
              <el-descriptions-item label="商户类型">{{ detail.business_type_text || detail.business_type || '-' }}</el-descriptions-item>
              <el-descriptions-item label="分类">{{ categoryName(detail.category_id) }}</el-descriptions-item>
              <el-descriptions-item label="联系电话">{{ maskPhone(detail.contact_phone) }}</el-descriptions-item>
              <el-descriptions-item label="地址" :span="3">{{ detail.address || '-' }}</el-descriptions-item>
              <el-descriptions-item label="经度">{{ detail.longitude || '-' }}</el-descriptions-item>
              <el-descriptions-item label="纬度">{{ detail.latitude || '-' }}</el-descriptions-item>
              <el-descriptions-item label="地区ID">{{ detail.region_id || '-' }}</el-descriptions-item>
              <el-descriptions-item v-if="detail.summary" label="简介" :span="3">{{ detail.summary }}</el-descriptions-item>
              <el-descriptions-item v-if="detail.content" label="详细描述" :span="3">
                <div class="content-box">{{ detail.content }}</div>
              </el-descriptions-item>
              <el-descriptions-item v-if="detail.audit_reason" label="审核原因" :span="3">{{ detail.audit_reason }}</el-descriptions-item>
            </el-descriptions>
          </el-card>

          <el-card class="section-card">
            <template #header><span class="section-title">封面/图集（{{ allImages.length }}）</span></template>
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
                <span class="section-title">营业时间（{{ businessHours.length }}）</span>
                <el-button type="primary" link size="small" @click="loadBusinessHours">刷新</el-button>
              </div>
            </template>
            <el-table :data="businessHours" border size="small">
              <el-table-column prop="day_of_week" label="星期" width="100">
                <template #default="{ row }">{{ weekText(row.day_of_week) }}</template>
              </el-table-column>
              <el-table-column prop="open_time" label="开门时间" width="120" />
              <el-table-column prop="close_time" label="关门时间" width="120" />
              <el-table-column label="营业状态" width="100">
                <template #default="{ row }">
                  <el-tag :type="row.is_closed ? 'danger' : 'success'" size="small">
                    {{ row.is_closed ? '休息' : '营业' }}
                  </el-tag>
                </template>
              </el-table-column>
              <el-table-column prop="remark" label="备注" min-width="160" />
            </el-table>
          </el-card>

          <el-card class="section-card">
            <template #header>
              <div class="card-header-flex">
                <span class="section-title">菜单（{{ menus.length }}）</span>
                <el-button type="primary" link size="small" @click="loadMenus">刷新</el-button>
              </div>
            </template>
            <el-table :data="menus" border size="small">
              <el-table-column prop="id" label="ID" width="70" />
              <el-table-column label="图片" width="70">
                <template #default="{ row }">
                  <el-image v-if="row.image" :src="row.image" fit="cover" class="menu-thumb" preview-teleported :preview-src-list="[row.image]" />
                  <span v-else class="text-muted">-</span>
                </template>
              </el-table-column>
              <el-table-column prop="name" label="菜名" min-width="160" />
              <el-table-column label="价格" width="100">
                <template #default="{ row }">¥{{ Number(row.price || 0).toFixed(2) }}</template>
              </el-table-column>
              <el-table-column label="招牌" width="80">
                <template #default="{ row }">
                  <el-tag v-if="row.is_signature" type="warning" size="small">招牌</el-tag>
                  <span v-else class="text-muted">-</span>
                </template>
              </el-table-column>
              <el-table-column prop="sold_count" label="已售" width="80" />
            </el-table>
          </el-card>
        </el-col>

        <!-- 右侧 -->
        <el-col :xs="24" :md="8">
          <el-card class="section-card">
            <template #header><span class="section-title">状态信息</span></template>
            <el-descriptions :column="1" border>
              <el-descriptions-item label="审核状态">
                <el-tag :type="auditTagType(detail.audit_status)" size="small">{{ auditText(detail.audit_status) }}</el-tag>
              </el-descriptions-item>
              <el-descriptions-item label="商户状态">
                <el-tag :type="statusTagType(detail.status)" size="small" effect="plain">{{ statusText(detail.status) }}</el-tag>
              </el-descriptions-item>
              <el-descriptions-item label="认证状态">
                <el-tag :type="verificationTagType(detail.verification_status)" size="small">{{ verificationText(detail.verification_status) }}</el-tag>
              </el-descriptions-item>
              <el-descriptions-item label="是否推荐">
                <el-tag v-if="detail.is_recommended" type="warning" size="small">推荐</el-tag>
                <span v-else>否</span>
              </el-descriptions-item>
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
                <div class="stats-value">{{ detail.call_count || 0 }}</div>
                <div class="stats-label">电话</div>
              </div>
              <div class="stats-item">
                <div class="stats-value">{{ detail.share_count || 0 }}</div>
                <div class="stats-label">分享</div>
              </div>
              <div class="stats-item">
                <div class="stats-value">{{ detail.review_count || 0 }}</div>
                <div class="stats-label">评价</div>
              </div>
              <div class="stats-item">
                <div class="stats-value">{{ Number(detail.rating || 0).toFixed(1) }}</div>
                <div class="stats-label">评分</div>
              </div>
            </div>
          </el-card>

          <el-card v-if="detail.business" class="section-card">
            <template #header><span class="section-title">商户资料</span></template>
            <el-descriptions :column="1" border>
              <el-descriptions-item label="法人">{{ detail.business.legal_person || '-' }}</el-descriptions-item>
              <el-descriptions-item label="营业执照号">{{ detail.business.license_no || '-' }}</el-descriptions-item>
              <el-descriptions-item label="营业面积">{{ detail.business.area ? detail.business.area + ' ㎡' : '-' }}</el-descriptions-item>
              <el-descriptions-item label="员工数">{{ detail.business.employee_count || '-' }}</el-descriptions-item>
              <el-descriptions-item label="简介">{{ detail.business.description || '-' }}</el-descriptions-item>
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
const businessHours = ref([])
const menus = ref([])
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
    const res = await request.get('/dh114/categories', { params: { page: 1, page_size: 200 } })
    const data = res.data || {}
    const list = data.list || data || []
    const map = {}
    list.forEach((c) => { map[c.id] = c.name })
    categoryMap.value = map
  } catch (e) {
    // ignore
  }
}

const loadDetail = async () => {
  loading.value = true
  try {
    const res = await request.get(`/dh114/admin/dh114s/${id.value}`)
    detail.value = res.data || null
  } catch (e) {
    ElMessage.error('加载详情失败')
  } finally {
    loading.value = false
  }
}

const loadBusinessHours = async () => {
  try {
    const res = await request.get(`/dh114/${id.value}/business-hours`)
    const data = res.data || {}
    businessHours.value = data.list || data || []
  } catch (e) {
    businessHours.value = []
  }
}

const loadMenus = async () => {
  try {
    const res = await request.get(`/dh114/${id.value}/menus`)
    const data = res.data || {}
    menus.value = data.list || data || []
  } catch (e) {
    menus.value = []
  }
}

const goBack = () => {
  router.push('/business/dh114/list')
}

// ===== 格式化 =====
const auditText = (s) => ({ 0: '待审核', 1: '已通过', 2: '已拒绝' }[s] || '-')
const auditTagType = (s) => ({ 0: 'warning', 1: 'success', 2: 'danger' }[s] || 'info')
const statusText = (s) => ({ 0: '草稿', 1: '已发布', 3: '已下架', 4: '已关闭' }[s] || '-')
const statusTagType = (s) => ({ 0: 'info', 1: 'success', 3: 'warning', 4: 'danger' }[s] || 'info')
const verificationText = (s) => ({ 0: '未认证', 1: '待审核', 2: '已认证', 3: '认证失败' }[s] || '-')
const verificationTagType = (s) => ({ 0: 'info', 1: 'warning', 2: 'success', 3: 'danger' }[s] || 'info')
const weekText = (d) => ({ 1: '周一', 2: '周二', 3: '周三', 4: '周四', 5: '周五', 6: '周六', 0: '周日' }[d] || '-')

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
      await request.put(`/dh114/admin/dh114s/${id.value}/audit`, { audit_status: auditStatus, audit_reason: value || '' })
    } else {
      await ElMessageBox.confirm('确定审核通过？', '提示', { type: 'warning' })
      await request.put(`/dh114/admin/dh114s/${id.value}/audit`, { audit_status: auditStatus })
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
    await ElMessageBox.confirm(`确定将商户设为「${label}」吗？`, '提示', { type: 'warning' })
    await request.put(`/dh114/admin/dh114s/${id.value}/status`, { status })
    ElMessage.success('操作成功')
    await loadDetail()
  } catch (e) {
    // 取消
  }
}

onMounted(async () => {
  await Promise.all([loadCategoryMap(), loadDetail(), loadBusinessHours(), loadMenus()])
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
.menu-thumb {
  width: 40px; height: 40px; border-radius: 4px; border: 1px solid #ebeef5;
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
.stats-value { font-size: 22px; font-weight: 600; color: #409eff; }
.stats-label { font-size: 12px; color: #909399; margin-top: 4px; }
.empty-text { color: #909399; text-align: center; padding: 24px 0; }
.action-bar {
  margin-top: 16px; padding: 12px 16px;
  background: #fff; border-radius: 4px;
  display: flex; gap: 8px; flex-wrap: wrap;
}
.text-muted { color: #909399; }
</style>
