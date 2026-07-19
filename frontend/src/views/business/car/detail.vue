<template>
  <div class="app-container">
    <el-page-header @back="goBack" content="车源详情" class="page-header" />
    <el-row :gutter="16" class="detail-row">
      <el-col :xs="24" :md="16">
        <el-card shadow="never" class="info-card">
          <template #header><div class="card-title">基本信息</div></template>
          <el-descriptions v-if="detail" :column="2" border>
            <el-descriptions-item label="ID">{{ detail.id }}</el-descriptions-item>
            <el-descriptions-item label="标题">{{ detail.title }}</el-descriptions-item>
            <el-descriptions-item label="品牌">{{ detail.brand_name || detail.brand }}</el-descriptions-item>
            <el-descriptions-item label="车系">{{ detail.series || detail.model_name }}</el-descriptions-item>
            <el-descriptions-item label="车型">{{ detail.model }}</el-descriptions-item>
            <el-descriptions-item label="年份">{{ detail.year }}</el-descriptions-item>
            <el-descriptions-item label="里程">{{ detail.mileage }}万公里</el-descriptions-item>
            <el-descriptions-item label="价格">¥{{ Number(detail.price || 0).toFixed(2) }}万</el-descriptions-item>
            <el-descriptions-item label="颜色">{{ detail.color || '-' }}</el-descriptions-item>
            <el-descriptions-item label="排量">{{ detail.displacement || '-' }}</el-descriptions-item>
            <el-descriptions-item label="变速箱">{{ detail.gearbox || '-' }}</el-descriptions-item>
            <el-descriptions-item label="燃油类型">{{ detail.fuel_type || '-' }}</el-descriptions-item>
            <el-descriptions-item label="上牌时间">{{ formatTime(detail.first_plate_date, 'YYYY-MM-DD') }}</el-descriptions-item>
            <el-descriptions-item label="排放标准">{{ detail.emission_standard || '-' }}</el-descriptions-item>
            <el-descriptions-item label="城市">{{ detail.city || '-' }}</el-descriptions-item>
            <el-descriptions-item label="地址">{{ detail.address || '-' }}</el-descriptions-item>
            <el-descriptions-item label="发布者">{{ detail.user_name || `#${detail.user_id}` }}</el-descriptions-item>
            <el-descriptions-item label="联系电话">{{ detail.contact_phone || '-' }}</el-descriptions-item>
            <el-descriptions-item label="真车验证">{{ detail.is_real_car ? '已验证' : '未验证' }}</el-descriptions-item>
            <el-descriptions-item label="描述" :span="2">{{ detail.description || '-' }}</el-descriptions-item>
          </el-descriptions>
          <el-skeleton v-else :rows="8" animated />
        </el-card>

        <el-card shadow="never" class="info-card" v-if="reports.length">
          <template #header><div class="card-title">举报历史</div></template>
          <el-table :data="reports" border size="small">
            <el-table-column prop="id" label="ID" width="70" />
            <el-table-column prop="report_type" label="类型" width="100" />
            <el-table-column prop="reason" label="原因" show-overflow-tooltip />
            <el-table-column label="状态" width="100">
              <template #default="{ row }">
                <el-tag size="small">{{ reportStatusText(row.status) }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column label="时间" width="160">
              <template #default="{ row }">{{ formatTime(row.created_at) }}</template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-col>

      <el-col :xs="24" :md="8">
        <el-card shadow="never" class="info-card">
          <template #header><div class="card-title">状态信息</div></template>
          <el-descriptions v-if="detail" :column="1" border>
            <el-descriptions-item label="审核状态">
              <el-tag :type="auditTagType(detail.audit_status)" size="small">{{ auditText(detail.audit_status) }}</el-tag>
            </el-descriptions-item>
            <el-descriptions-item label="发布状态">
              <el-tag :type="statusTagType(detail.status)" size="small">{{ statusText(detail.status) }}</el-tag>
            </el-descriptions-item>
            <el-descriptions-item label="浏览数">{{ detail.view_count || 0 }}</el-descriptions-item>
            <el-descriptions-item label="收藏数">{{ detail.fav_count || 0 }}</el-descriptions-item>
            <el-descriptions-item label="联系数">{{ detail.contact_count || 0 }}</el-descriptions-item>
            <el-descriptions-item label="分享数">{{ detail.share_count || 0 }}</el-descriptions-item>
            <el-descriptions-item label="发布时间">{{ formatTime(detail.published_at) }}</el-descriptions-item>
            <el-descriptions-item label="创建时间">{{ formatTime(detail.created_at) }}</el-descriptions-item>
            <el-descriptions-item label="更新时间">{{ formatTime(detail.updated_at) }}</el-descriptions-item>
          </el-descriptions>
          <el-skeleton v-else :rows="6" animated />
        </el-card>
      </el-col>
    </el-row>

    <div class="footer-actions" v-if="detail">
      <el-button v-if="detail.audit_status === 0 || detail.audit_status === 2" type="success" @click="handleAudit(1)">审核通过</el-button>
      <el-button v-if="detail.audit_status === 0 || detail.audit_status === 1" type="danger" @click="handleAudit(2)">审核拒绝</el-button>
      <el-button v-if="detail.status === 0 || detail.status === 2" type="primary" @click="handleStatus(1)">设为已发布</el-button>
      <el-button v-if="detail.status === 1" type="warning" @click="handleStatus(2)">下架</el-button>
      <el-button v-if="!detail.is_real_car" type="info" @click="handleRealVerify">真车验证</el-button>
      <el-popconfirm title="确认删除该车源？" @confirm="handleDelete">
        <template #reference><el-button type="danger">删除</el-button></template>
      </el-popconfirm>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { adminGetCar, auditCar, adminUpdateCarStatus, realCarVerify, deleteCar, adminListReports } from '@/api/car'
import { formatTime } from '@/utils/format'

const route = useRoute()
const router = useRouter()

const detail = ref(null)
const reports = ref([])

const auditText = (s) => ({ 0: '待审核', 1: '已通过', 2: '已拒绝' }[s] || '-')
const auditTagType = (s) => ({ 0: 'warning', 1: 'success', 2: 'danger' }[s] || 'info')
const statusText = (s) => ({ 0: '草稿', 1: '已发布', 2: '已下架', 3: '已售出' }[s] || '-')
const statusTagType = (s) => ({ 0: 'info', 1: 'success', 2: 'warning', 3: 'primary' }[s] || 'info')
const reportStatusText = (s) => ({ 0: '待处理', 1: '已处理', 2: '已驳回', 3: '申诉中' }[s] || '-')

const goBack = () => router.push('/business/car')

const loadDetail = async () => {
  const id = route.params.id
  if (!id) return
  try {
    const [detailRes, reportsRes] = await Promise.all([
      adminGetCar(id),
      adminListReports({ car_id: id, page: 1, page_size: 20 }).catch(() => null)
    ])
    detail.value = detailRes.data || null
    if (reportsRes) reports.value = reportsRes.data?.list || []
  } catch (e) {
    if (e?.message) ElMessage.error(e.message)
  }
}

const handleAudit = async (status) => {
  try {
    if (status === 2) {
      const { value } = await ElMessageBox.prompt('请输入拒绝原因', '审核拒绝', { type: 'warning', inputType: 'textarea' })
      await auditCar(detail.value.id, { audit_status: status, audit_reason: value })
    } else {
      await auditCar(detail.value.id, { audit_status: status })
    }
    ElMessage.success('操作成功')
    loadDetail()
  } catch (e) {
    if (e !== 'cancel' && e?.message) ElMessage.error(e.message)
  }
}

const handleStatus = async (status) => {
  try {
    await adminUpdateCarStatus(detail.value.id, { status })
    ElMessage.success('操作成功')
    loadDetail()
  } catch (e) {
    if (e?.message) ElMessage.error(e.message)
  }
}

const handleRealVerify = async () => {
  try {
    const { value } = await ElMessageBox.prompt('请输入真车验证备注', '真车验证', { inputType: 'textarea' })
    await realCarVerify(detail.value.id, { remark: value })
    ElMessage.success('已标记为真车验证')
    loadDetail()
  } catch (e) {
    if (e !== 'cancel' && e?.message) ElMessage.error(e.message)
  }
}

const handleDelete = async () => {
  try {
    await deleteCar(detail.value.id)
    ElMessage.success('删除成功')
    router.push('/business/car')
  } catch (e) {
    if (e?.message) ElMessage.error(e.message)
  }
}

onMounted(() => loadDetail())
</script>

<style scoped>
.page-header { margin-bottom: 16px; }
.detail-row { margin-bottom: 16px; }
.info-card { margin-bottom: 16px; }
.card-title { font-weight: 600; }
.footer-actions { padding: 12px 16px; background: #fff; border-radius: 4px; display: flex; gap: 8px; flex-wrap: wrap; }
</style>
