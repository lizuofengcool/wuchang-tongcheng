<template>
  <div class="app-container">
    <!-- 顶部统计卡片 -->
    <el-row :gutter="16" class="stat-row">
      <el-col :xs="12" :sm="8" :md="4">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value">{{ stats.total }}</div><div class="stat-label">评价总数</div></div></el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="4">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value text-success">{{ stats.avgScore.toFixed(1) }}</div><div class="stat-label">平均分</div></div></el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="4">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value text-primary">{{ stats.good }}</div><div class="stat-label">好评数</div></div></el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="4">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value text-warning">{{ stats.medium }}</div><div class="stat-label">中评数</div></div></el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="4">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value text-danger">{{ stats.bad }}</div><div class="stat-label">差评数</div></div></el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="4">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value">{{ stats.pending }}</div><div class="stat-label">待审核</div></div></el-card>
      </el-col>
    </el-row>

    <div class="page-card">
      <el-form :inline="true" :model="filters" class="filter-form" @submit.prevent>
        <el-form-item label="关键词">
          <el-input v-model="filters.keyword" placeholder="评价编号/内容" clearable style="width: 200px" @keyup.enter="onSearch" />
        </el-form-item>
        <el-form-item label="评分">
          <el-select v-model="filters.rating" placeholder="全部" clearable style="width: 120px" @change="onSearch">
            <el-option label="5分" :value="5" />
            <el-option label="4分" :value="4" />
            <el-option label="3分" :value="3" />
            <el-option label="2分" :value="2" />
            <el-option label="1分" :value="1" />
          </el-select>
        </el-form-item>
        <el-form-item label="类型">
          <el-select v-model="filters.target_type" placeholder="全部" clearable style="width: 140px" @change="onSearch">
            <el-option v-for="(label, val) in typeMap" :key="val" :label="label" :value="val" />
          </el-select>
        </el-form-item>
        <el-form-item label="审核状态">
          <el-select v-model="filters.audit_status" placeholder="全部" clearable style="width: 120px" @change="onSearch">
            <el-option label="待审核" :value="0" />
            <el-option label="已通过" :value="1" />
            <el-option label="已拒绝" :value="2" />
          </el-select>
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
        <el-table-column prop="id" label="ID" width="70" fixed="left" />
        <el-table-column prop="rating_no" label="评价编号" width="180" />
        <el-table-column label="评价对象" min-width="180">
          <template #default="{ row }">
            <div>{{ row.target_name || `#${row.target_id}` }}</div>
            <el-tag size="small">{{ typeMap[row.target_type] || row.target_type }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="评价人" width="120">
          <template #default="{ row }">{{ row.rater_name || `#${row.rater_id}` }}</template>
        </el-table-column>
        <el-table-column label="评分" width="160">
          <template #default="{ row }">
            <el-rate :model-value="row.score || 0" disabled show-score />
          </template>
        </el-table-column>
        <el-table-column label="内容" min-width="200" show-overflow-tooltip>
          <template #default="{ row }">{{ row.content || '-' }}</template>
        </el-table-column>
        <el-table-column label="回复" width="80">
          <template #default="{ row }">
            <el-badge v-if="row.reply" value="回" type="primary" />
            <span v-else class="text-muted">-</span>
          </template>
        </el-table-column>
        <el-table-column label="点赞" width="70" prop="like_count" />
        <el-table-column label="审核" width="100">
          <template #default="{ row }">
            <el-tag :type="auditTagType(row.audit_status)" size="small">{{ auditText(row.audit_status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="创建时间" width="160">
          <template #default="{ row }">{{ formatTime(row.created_at) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="openDetail(row)">详情</el-button>
            <el-button v-if="row.audit_status === 0" type="success" link size="small" @click="onAudit(row, 1)">通过</el-button>
            <el-button v-if="row.audit_status === 0 || row.audit_status === 1" type="danger" link size="small" @click="onAudit(row, 2)">拒绝</el-button>
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

    <!-- 详情弹窗 -->
    <el-dialog v-model="detailVisible" title="评价详情" width="700px">
      <el-descriptions v-if="detail" :column="2" border>
        <el-descriptions-item label="ID">{{ detail.id }}</el-descriptions-item>
        <el-descriptions-item label="评价编号">{{ detail.rating_no }}</el-descriptions-item>
        <el-descriptions-item label="评价对象" :span="2">{{ detail.target_name || `#${detail.target_id}` }}</el-descriptions-item>
        <el-descriptions-item label="类型">
          <el-tag size="small">{{ typeMap[detail.target_type] || detail.target_type }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="评价人">{{ detail.rater_name || `#${detail.rater_id}` }}</el-descriptions-item>
        <el-descriptions-item label="评分" :span="2">
          <el-rate :model-value="detail.score || 0" disabled show-score />
        </el-descriptions-item>
        <el-descriptions-item label="内容" :span="2">{{ detail.content || '-' }}</el-descriptions-item>
        <el-descriptions-item v-if="detail.images && detail.images.length" label="图片" :span="2">
          <div class="images-grid">
            <el-image v-for="(img, idx) in detail.images" :key="idx" :src="img" fit="cover" class="image-item" :preview-src-list="detail.images" :initial-index="idx" preview-teleported />
          </div>
        </el-descriptions-item>
        <el-descriptions-item v-if="detail.reply" label="回复" :span="2">{{ detail.reply }}</el-descriptions-item>
        <el-descriptions-item v-if="detail.append" label="追评" :span="2">{{ detail.append }}</el-descriptions-item>
        <el-descriptions-item label="点赞数">{{ detail.like_count || 0 }}</el-descriptions-item>
        <el-descriptions-item label="审核状态">
          <el-tag :type="auditTagType(detail.audit_status)" size="small">{{ auditText(detail.audit_status) }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="创建时间">{{ formatTime(detail.created_at) }}</el-descriptions-item>
        <el-descriptions-item label="更新时间">{{ formatTime(detail.updated_at) }}</el-descriptions-item>
      </el-descriptions>
      <template #footer><el-button @click="detailVisible = false">关闭</el-button></template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Refresh, RefreshLeft, Search } from '@element-plus/icons-vue'
import { adminListLinggongRatings, getLinggongRating, auditLinggongRating } from '@/api/linggong'
import { formatTime } from '@/utils/format'

const loading = ref(false)
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)
const list = ref([])

const filters = reactive({ keyword: '', rating: null, target_type: '', audit_status: null })

const typeMap = {
  linggong: '岗位', employer: '雇主', worker: '工人', task: '任务'
}
const auditText = (s) => ({ 0: '待审核', 1: '已通过', 2: '已拒绝' }[s] || '-')
const auditTagType = (s) => ({ 0: 'warning', 1: 'success', 2: 'danger' }[s] || 'info')

const stats = computed(() => {
  const total = list.value.length
  let sumScore = 0
  let good = 0
  let medium = 0
  let bad = 0
  let pending = 0
  list.value.forEach((r) => {
    sumScore += Number(r.score || 0)
    if (r.score >= 4) good++
    else if (r.score === 3) medium++
    else if (r.score > 0) bad++
    if (r.audit_status === 0) pending++
  })
  return {
    total,
    avgScore: total > 0 ? sumScore / total : 0,
    good,
    medium,
    bad,
    pending
  }
})

const onSearch = () => { page.value = 1; loadList() }
const onReset = () => {
  filters.keyword = ''
  filters.rating = null
  filters.target_type = ''
  filters.audit_status = null
  page.value = 1
  loadList()
}

const loadList = async () => {
  loading.value = true
  try {
    const res = await adminListLinggongRatings({
      page: page.value,
      page_size: pageSize.value,
      keyword: filters.keyword || undefined,
      rating: filters.rating || undefined,
      target_type: filters.target_type || undefined,
      audit_status: filters.audit_status === null || filters.audit_status === '' ? undefined : filters.audit_status
    })
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

const detailVisible = ref(false)
const detail = ref(null)

const openDetail = async (row) => {
  try {
    const res = await getLinggongRating(row.id)
    detail.value = res.data || row
  } catch (e) {
    detail.value = row
  }
  detailVisible.value = true
}

const onAudit = async (row, auditStatus) => {
  try {
    if (auditStatus === 2) {
      const { value } = await ElMessageBox.prompt('请输入拒绝原因', '拒绝评价', {
        inputType: 'textarea',
        inputPlaceholder: '拒绝原因'
      })
      await auditLinggongRating(row.id, { audit_status: auditStatus, audit_reason: value || '' })
    } else {
      await ElMessageBox.confirm('确定通过该评价吗？', '提示', { type: 'warning' })
      await auditLinggongRating(row.id, { audit_status: auditStatus })
    }
    ElMessage.success('操作成功')
    await loadList()
  } catch (e) {
    // 取消
  }
}

onMounted(() => loadList())
</script>

<style scoped>
.stat-row { margin-bottom: 16px; }
.stat-item { text-align: center; padding: 16px; }
.stat-value { font-size: 22px; font-weight: 600; color: #303133; }
.text-primary { color: #409eff; }
.text-warning { color: #e6a23c; }
.text-success { color: #67c23a; }
.text-danger { color: #f56c6c; }
.text-muted { color: #909399; }
.filter-form { margin-bottom: 12px; padding: 12px 16px; background: #fafafa; border-radius: 4px; }
.filter-form :deep(.el-form-item) { margin-bottom: 8px; margin-right: 12px; }
.toolbar { display: flex; justify-content: flex-start; margin-bottom: 12px; }
.images-grid { display: flex; flex-wrap: wrap; gap: 8px; }
.image-item { width: 100px; height: 100px; border-radius: 4px; border: 1px solid #ebeef5; }
.pagination-wrap { margin-top: 16px; display: flex; justify-content: flex-end; }
</style>
