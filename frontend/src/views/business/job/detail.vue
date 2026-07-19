<template>
  <div class="app-container" v-loading="loading">
    <div v-if="detail">
      <el-page-header @back="goBack" :content="`职位详情 #${detail.id}`" style="margin-bottom: 16px" />

      <el-row :gutter="16">
        <el-col :xs="24" :md="16">
          <el-card shadow="never" class="detail-card">
            <template #header>
              <div class="card-header">
                <span class="title">{{ detail.title }}</span>
                <div>
                  <el-tag :type="auditTagType(detail.audit_status)" size="small">{{ auditText(detail.audit_status) }}</el-tag>
                  <el-tag :type="statusTagType(detail.status)" size="small" effect="plain" style="margin-left: 6px">{{ statusText(detail.status) }}</el-tag>
                </div>
              </div>
            </template>
            <el-descriptions :column="3" border>
              <el-descriptions-item label="薪资">{{ formatSalary(detail) }}</el-descriptions-item>
              <el-descriptions-item label="公司">{{ detail.company_name || `公司#${detail.company_id}` }}</el-descriptions-item>
              <el-descriptions-item label="学历">{{ educationText(detail.education) }}</el-descriptions-item>
              <el-descriptions-item label="经验">{{ detail.experience || '-' }}</el-descriptions-item>
              <el-descriptions-item label="城市">{{ detail.city || '不限' }}</el-descriptions-item>
              <el-descriptions-item label="地址">{{ detail.address || '-' }}</el-descriptions-item>
              <el-descriptions-item label="招聘人数">{{ detail.headcount || '-' }}</el-descriptions-item>
              <el-descriptions-item label="加急">
                <el-tag v-if="detail.is_urgent" type="danger" size="small">急招</el-tag>
                <span v-else>否</span>
              </el-descriptions-item>
              <el-descriptions-item label="发布者">{{ detail.user_name || `用户#${detail.user_id}` }}</el-descriptions-item>
              <el-descriptions-item label="福利" :span="3">{{ Array.isArray(detail.benefits) ? detail.benefits.join('、') : (detail.benefits || '-') }}</el-descriptions-item>
              <el-descriptions-item label="技能要求" :span="3">{{ Array.isArray(detail.skills) ? detail.skills.join('、') : (detail.skills || '-') }}</el-descriptions-item>
              <el-descriptions-item v-if="detail.description" label="职位描述" :span="3">
                <div class="content-box">{{ detail.description }}</div>
              </el-descriptions-item>
              <el-descriptions-item v-if="detail.audit_reason" label="审核原因" :span="3">{{ detail.audit_reason }}</el-descriptions-item>
            </el-descriptions>
          </el-card>

          <el-card shadow="never" class="detail-card" style="margin-top: 16px">
            <template #header>举报历史</template>
            <el-table :data="reports" border size="small">
              <el-table-column prop="report_no" label="举报单号" width="160" />
              <el-table-column prop="reason" label="原因" min-width="180" show-overflow-tooltip />
              <el-table-column label="状态" width="100">
                <template #default="{ row }">
                  <el-tag :type="row.status === 0 ? 'warning' : 'info'" size="small">{{ row.status === 0 ? '待处理' : '已处理' }}</el-tag>
                </template>
              </el-table-column>
              <el-table-column label="时间" width="160">
                <template #default="{ row }">{{ formatTime(row.created_at) }}</template>
              </el-table-column>
            </el-table>
            <div v-if="!reports.length" class="empty-text">暂无举报记录</div>
          </el-card>
        </el-col>

        <el-col :xs="24" :md="8">
          <el-card shadow="never" class="detail-card">
            <template #header>状态信息</template>
            <el-descriptions :column="1" border>
              <el-descriptions-item label="浏览量">{{ detail.view_count || 0 }}</el-descriptions-item>
              <el-descriptions-item label="投递数">{{ detail.application_count || 0 }}</el-descriptions-item>
              <el-descriptions-item label="收藏量">{{ detail.fav_count || 0 }}</el-descriptions-item>
              <el-descriptions-item label="发布时间">{{ formatTime(detail.published_at) }}</el-descriptions-item>
              <el-descriptions-item label="创建时间">{{ formatTime(detail.created_at) }}</el-descriptions-item>
              <el-descriptions-item label="过期时间">{{ formatTime(detail.expiry_time) }}</el-descriptions-item>
            </el-descriptions>
          </el-card>
        </el-col>
      </el-row>

      <el-card shadow="never" class="detail-card" style="margin-top: 16px">
        <div class="action-bar">
          <el-button v-if="detail.audit_status === 0 || detail.audit_status === 2" type="success" :icon="Check" @click="handleAudit(1)">审核通过</el-button>
          <el-button v-if="detail.audit_status === 0 || detail.audit_status === 1" type="danger" :icon="Close" @click="handleAudit(2)">审核拒绝</el-button>
          <el-button v-if="detail.status !== 1" type="primary" @click="handleStatus(1)">设为招聘中</el-button>
          <el-button v-if="detail.status !== 3" type="warning" @click="handleStatus(3)">下架</el-button>
          <el-button v-if="detail.status !== 4" @click="handleStatus(4)">设为过期</el-button>
          <el-button type="danger" :icon="Delete" @click="handleDelete">删除</el-button>
        </div>
      </el-card>
    </div>
    <el-empty v-else-if="!loading" description="未找到职位" />
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Check, Close, Delete } from '@element-plus/icons-vue'
import { adminGetJob, auditJob, adminUpdateJobStatus, deleteJob, adminListReports } from '@/api/job'
import { formatTime } from '@/utils/format'

const route = useRoute()
const router = useRouter()
const loading = ref(false)
const detail = ref(null)
const reports = ref([])

const auditText = (s) => ({ 0: '待审核', 1: '已通过', 2: '已拒绝' }[s] || '-')
const auditTagType = (s) => ({ 0: 'warning', 1: 'success', 2: 'danger' }[s] || 'info')
const statusText = (s) => ({ 0: '草稿', 1: '招聘中', 2: '已停招', 3: '已下架', 4: '已过期' }[s] || '-')
const statusTagType = (s) => ({ 0: 'info', 1: 'success', 2: 'warning', 3: 'warning', 4: 'danger' }[s] || 'info')
const educationText = (e) => ({ unlimited: '不限', college: '大专', bachelor: '本科', master: '硕士', doctor: '博士' }[e] || '-')

const formatSalary = (row) => {
  const min = Number(row.salary_min || 0)
  const max = Number(row.salary_max || 0)
  if (!min && !max) return '面议'
  if (min && max) return `${min}-${max}K`
  return min ? `${min}K起` : `${max}K以内`
}

const loadDetail = async () => {
  const id = route.params.id
  if (!id) return
  loading.value = true
  try {
    const [dRes, rRes] = await Promise.all([
      adminGetJob(id),
      adminListReports({ target_id: id, page: 1, page_size: 20 }).catch(() => ({ data: { list: [] } }))
    ])
    detail.value = dRes.data || null
    reports.value = rRes.data?.list || []
  } catch (e) {
    detail.value = null
  } finally {
    loading.value = false
  }
}

const handleAudit = async (auditStatus) => {
  try {
    if (auditStatus === 2) {
      const { value } = await ElMessageBox.prompt('请输入拒绝原因（可选）', '拒绝审核', {
        confirmButtonText: '确定拒绝', cancelButtonText: '取消', inputType: 'textarea'
      })
      await auditJob(detail.value.id, { audit_status: auditStatus, audit_reason: value || '' })
    } else {
      await ElMessageBox.confirm('确定审核通过吗？', '提示', { type: 'warning' })
      await auditJob(detail.value.id, { audit_status: auditStatus })
    }
    ElMessage.success('操作成功')
    await loadDetail()
  } catch (e) { /* cancel */ }
}

const handleStatus = async (status) => {
  try {
    const label = statusText(status)
    await ElMessageBox.confirm(`确定设为「${label}」吗？`, '提示', { type: 'warning' })
    await adminUpdateJobStatus(detail.value.id, status)
    ElMessage.success('状态更新成功')
    await loadDetail()
  } catch (e) { /* cancel */ }
}

const handleDelete = async () => {
  try {
    await ElMessageBox.confirm('确定删除该职位吗？删除后不可恢复！', '危险操作', { type: 'error' })
    await deleteJob(detail.value.id)
    ElMessage.success('已删除')
    router.push('/business/job/list')
  } catch (e) { /* cancel */ }
}

const goBack = () => router.push('/business/job/list')

onMounted(() => loadDetail())
</script>

<style scoped>
.detail-card { margin-bottom: 0; }
.card-header { display: flex; justify-content: space-between; align-items: center; }
.card-header .title { font-size: 16px; font-weight: 600; }
.content-box { white-space: pre-wrap; line-height: 1.6; }
.empty-text { text-align: center; color: #999; padding: 20px; }
.action-bar { display: flex; flex-wrap: wrap; gap: 8px; }
</style>
