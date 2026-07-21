<template>
  <div class="app-container" v-loading="loading">
    <el-page-header @back="goBack" class="page-header">
      <template #content>
        <span class="header-title">团购详情 #{{ detail?.id || '-' }}</span>
      </template>
    </el-page-header>

    <div v-if="detail" class="page-card">
      <el-descriptions :column="3" border title="基本信息">
        <el-descriptions-item label="ID">{{ detail.id }}</el-descriptions-item>
        <el-descriptions-item label="标题" :span="2">{{ detail.title }}</el-descriptions-item>
        <el-descriptions-item label="副标题" :span="3">{{ detail.sub_title || '-' }}</el-descriptions-item>
        <el-descriptions-item label="店铺">{{ detail.shop_name || `#${detail.shop_id}` }}</el-descriptions-item>
        <el-descriptions-item label="分类ID">{{ detail.category_id || '-' }}</el-descriptions-item>
        <el-descriptions-item label="状态">
          <el-tag :type="statusTagType(detail.status)" size="small" effect="plain">{{ statusText(detail.status) }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="审核状态">
          <el-tag :type="auditTagType(detail.audit_status)" size="small">{{ auditText(detail.audit_status) }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="审核原因" v-if="detail.audit_reason">{{ detail.audit_reason }}</el-descriptions-item>
      </el-descriptions>

      <el-descriptions :column="3" border title="价格与库存" class="sub-section">
        <el-descriptions-item label="团购价">¥{{ Number(detail.price || 0).toFixed(2) }}</el-descriptions-item>
        <el-descriptions-item label="原价">¥{{ Number(detail.origin_price || 0).toFixed(2) }}</el-descriptions-item>
        <el-descriptions-item label="折扣">{{ discountText(detail.price, detail.origin_price) }}</el-descriptions-item>
        <el-descriptions-item label="库存">{{ detail.stock || 0 }}</el-descriptions-item>
        <el-descriptions-item label="销量">{{ detail.sales_count || 0 }}</el-descriptions-item>
        <el-descriptions-item label="限购">{{ detail.limit_per_user || 0 }} 件/人</el-descriptions-item>
      </el-descriptions>

      <el-descriptions :column="3" border title="时间信息" class="sub-section">
        <el-descriptions-item label="开始时间">{{ formatTime(detail.start_time) }}</el-descriptions-item>
        <el-descriptions-item label="结束时间">{{ formatTime(detail.end_time) }}</el-descriptions-item>
        <el-descriptions-item label="创建时间">{{ formatTime(detail.created_at) }}</el-descriptions-item>
        <el-descriptions-item label="更新时间" :span="3">{{ formatTime(detail.updated_at) }}</el-descriptions-item>
      </el-descriptions>

      <el-descriptions :column="3" border title="推广设置" class="sub-section">
        <el-descriptions-item label="精选">{{ detail.is_featured ? '是' : '否' }}</el-descriptions-item>
        <el-descriptions-item label="热门">{{ detail.is_hot ? '是' : '否' }}</el-descriptions-item>
        <el-descriptions-item label="浏览量">{{ detail.view_count || 0 }}</el-descriptions-item>
      </el-descriptions>

      <div v-if="detail.cover_image" class="cover-wrap">
        <h4 class="sub-title">封面图</h4>
        <el-image :src="detail.cover_image" fit="cover" style="width: 320px; height: 180px; border-radius: 6px" :preview-src-list="[detail.cover_image]" />
      </div>

      <div v-if="detail.description" class="desc-wrap">
        <h4 class="sub-title">详情描述</h4>
        <div class="desc-content">{{ detail.description }}</div>
      </div>

      <div class="action-bar">
        <el-button v-if="detail.audit_status === 0 || detail.audit_status === 2" type="success" @click="handleAudit(1)">审核通过</el-button>
        <el-button v-if="detail.audit_status === 0 || detail.audit_status === 1" type="danger" @click="handleAudit(2)">审核拒绝</el-button>
        <el-button v-if="detail.status === 0" type="primary" @click="handleStatus(1)">上架</el-button>
        <el-button v-if="detail.status === 1" type="warning" @click="handleStatus(0)">下架</el-button>
        <el-button type="danger" plain @click="handleDelete">删除</el-button>
      </div>
    </div>

    <el-empty v-else-if="!loading" description="未找到团购信息" />
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { formatTime } from '@/utils/format'
import { getGroupbuyDetail, deleteGroupbuy, auditGroupbuy, updateGroupbuyStatus } from '@/api/groupbuy'

const route = useRoute()
const router = useRouter()

const loading = ref(false)
const detail = ref(null)

const statusText = (s) => ({ 0: '下架', 1: '上架', 2: '售罄' }[s] || '-')
const statusTagType = (s) => ({ 0: 'info', 1: 'success', 2: 'warning' }[s] || 'info')
const auditText = (s) => ({ 0: '待审核', 1: '已通过', 2: '已拒绝' }[s] || '-')
const auditTagType = (s) => ({ 0: 'warning', 1: 'success', 2: 'danger' }[s] || 'info')

const discountText = (price, origin) => {
  if (!origin || origin <= 0 || !price) return '-'
  const d = (price / origin * 10).toFixed(1)
  return d + ' 折'
}

const loadDetail = async () => {
  const id = route.params.id
  if (!id) return
  loading.value = true
  try {
    const res = await getGroupbuyDetail(id)
    detail.value = res.data || null
  } catch (e) {
    ElMessage.error('加载详情失败')
  } finally {
    loading.value = false
  }
}

const handleAudit = async (auditStatus) => {
  if (!detail.value) return
  try {
    if (auditStatus === 2) {
      const { value } = await ElMessageBox.prompt('请输入拒绝原因（可选）', '拒绝审核', {
        confirmButtonText: '确定', cancelButtonText: '取消', inputType: 'textarea', inputPlaceholder: '拒绝原因'
      })
      await auditGroupbuy(detail.value.id, { audit_status: auditStatus, audit_reason: value || '' })
    } else {
      await ElMessageBox.confirm('确定通过审核吗？', '提示', { type: 'warning' })
      await auditGroupbuy(detail.value.id, { audit_status: auditStatus })
    }
    ElMessage.success('操作成功')
    await loadDetail()
  } catch (e) { /* cancel */ }
}

const handleStatus = async (status) => {
  if (!detail.value) return
  try {
    const action = status === 1 ? '上架' : '下架'
    await ElMessageBox.confirm(`确定${action}此团购吗？`, '提示', { type: 'warning' })
    await updateGroupbuyStatus(detail.value.id, status)
    ElMessage.success(`${action}成功`)
    await loadDetail()
  } catch (e) { /* cancel */ }
}

const handleDelete = async () => {
  if (!detail.value) return
  try {
    await ElMessageBox.confirm('确定删除此团购吗？此操作不可恢复！', '危险操作', { type: 'warning', confirmButtonText: '确认删除', cancelButtonText: '取消' })
    await deleteGroupbuy(detail.value.id)
    ElMessage.success('删除成功')
    router.push('/marketing/groupbuy/list')
  } catch (e) { /* cancel */ }
}

const goBack = () => {
  router.push('/marketing/groupbuy/list')
}

onMounted(() => loadDetail())
</script>

<style scoped>
.page-header { margin-bottom: 16px; }
.header-title { font-size: 16px; font-weight: 600; }
.sub-section { margin-top: 16px; }
.sub-title { margin: 16px 0 8px; font-weight: 600; color: #303133; }
.cover-wrap { margin-top: 16px; }
.desc-wrap { margin-top: 16px; }
.desc-content { padding: 12px 16px; background: #fafafa; border-radius: 4px; color: #606266; line-height: 1.6; white-space: pre-wrap; }
.action-bar { margin-top: 24px; padding-top: 16px; border-top: 1px solid #ebeef5; display: flex; gap: 8px; flex-wrap: wrap; }
</style>
