<template>
  <div class="app-container">
    <!-- 顶部统计卡片 -->
    <el-row :gutter="16" class="stat-row">
      <el-col :xs="12" :sm="8" :md="4">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value">{{ stats.total }}</div><div class="stat-label">故事总数</div></div></el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="4">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value text-success">{{ stats.featured }}</div><div class="stat-label">精选故事</div></div></el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="4">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value text-warning">{{ stats.pendingAudit }}</div><div class="stat-label">待审核</div></div></el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="4">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value text-primary">{{ stats.today }}</div><div class="stat-label">今日新增</div></div></el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="4">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value">{{ stats.totalView }}</div><div class="stat-label">总浏览</div></div></el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="4">
        <el-card shadow="hover"><div class="stat-item"><div class="stat-value text-danger">{{ stats.totalLike }}</div><div class="stat-label">总点赞</div></div></el-card>
      </el-col>
    </el-row>

    <div class="page-card">
      <el-form :inline="true" :model="filters" class="filter-form" @submit.prevent>
        <el-form-item label="关键词">
          <el-input v-model="filters.keyword" placeholder="标题/作者/内容" clearable style="width: 220px" :prefix-icon="Search" @keyup.enter="onSearch" />
        </el-form-item>
        <el-form-item label="话题">
          <el-select v-model="filters.topic" placeholder="全部" clearable filterable style="width: 140px" @change="onSearch">
            <el-option label="相亲故事" value="match" />
            <el-option label="恋爱心得" value="love" />
            <el-option label="约会技巧" value="dating" />
            <el-option label="婚姻感悟" value="marriage" />
            <el-option label="单身生活" value="single" />
          </el-select>
        </el-form-item>
        <el-form-item label="审核状态">
          <el-select v-model="filters.audit_status" placeholder="全部" clearable style="width: 120px" @change="onSearch">
            <el-option label="待审核" :value="0" />
            <el-option label="已通过" :value="1" />
            <el-option label="已拒绝" :value="2" />
          </el-select>
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="filters.status" placeholder="全部" clearable style="width: 120px" @change="onSearch">
            <el-option label="正常" :value="1" />
            <el-option label="下架" :value="3" />
          </el-select>
        </el-form-item>
        <el-form-item label="精选">
          <el-select v-model="filters.is_featured" placeholder="全部" clearable style="width: 100px" @change="onSearch">
            <el-option label="是" :value="1" />
            <el-option label="否" :value="0" />
          </el-select>
        </el-form-item>
        <el-form-item label="时间范围">
          <el-date-picker
            v-model="filters.dateRange"
            type="daterange"
            range-separator="至"
            start-placeholder="开始日期"
            end-placeholder="结束日期"
            value-format="YYYY-MM-DD"
            style="width: 240px"
          />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" :icon="Search" @click="onSearch">搜索</el-button>
          <el-button :icon="RefreshLeft" @click="onReset">重置</el-button>
        </el-form-item>
      </el-form>

      <div class="toolbar">
        <div class="toolbar-left">
          <el-button :icon="Refresh" @click="loadList">刷新</el-button>
          <el-button type="success" :icon="Check" :disabled="!selection.length" @click="onBatchAudit(1)">批量通过</el-button>
        </div>
        <div class="toolbar-right">
          <el-button type="primary" :icon="Plus" @click="openCreate">新建故事</el-button>
        </div>
      </div>

      <el-table v-loading="loading" :data="list" border stripe @selection-change="onSelectionChange" @sort-change="onSortChange">
        <el-table-column type="selection" width="44" fixed="left" />
        <el-table-column prop="id" label="ID" width="70" sortable="custom" fixed="left" />
        <el-table-column label="封面" width="80">
          <template #default="{ row }">
            <el-image v-if="row.cover_image" :src="row.cover_image" fit="cover" class="cover-thumb" :preview-src-list="[row.cover_image]" preview-teleported />
            <div v-else class="cover-thumb cover-empty">无图</div>
          </template>
        </el-table-column>
        <el-table-column label="标题/作者" min-width="220">
          <template #default="{ row }">
            <div class="title-cell">
              <div class="title-text">
                <el-link type="primary" :underline="'never'" @click="openDetail(row)">{{ row.title }}</el-link>
                <el-tag v-if="row.is_featured" type="danger" size="small" effect="dark">精</el-tag>
              </div>
              <div class="title-desc">
                <span>{{ row.author_name || `#${row.user_id}` }}</span>
                <el-tag v-if="row.topic" size="small" type="info">{{ topicText(row.topic) }}</el-tag>
              </div>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="浏览" width="70" prop="view_count" sortable="custom" />
        <el-table-column label="点赞" width="70" prop="like_count" sortable="custom" />
        <el-table-column label="评论" width="70" prop="comment_count" />
        <el-table-column label="分享" width="70" prop="share_count" />
        <el-table-column label="审核" width="90">
          <template #default="{ row }">
            <el-tag :type="auditTagType(row.audit_status)" size="small">{{ auditText(row.audit_status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="90">
          <template #default="{ row }">
            <el-tag :type="statusTagType(row.status)" size="small" effect="plain">{{ statusText(row.status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="发布时间" width="160" prop="created_at" sortable="custom">
          <template #default="{ row }">{{ formatTime(row.created_at) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="220" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="openDetail(row)">详情</el-button>
            <el-button
              v-if="row.audit_status === 0 || row.audit_status === 2"
              type="success"
              link
              size="small"
              @click="handleAudit(row, 1)"
            >通过</el-button>
            <el-button
              v-if="row.audit_status === 0 || row.audit_status === 1"
              type="danger"
              link
              size="small"
              @click="handleAudit(row, 2)"
            >拒绝</el-button>
            <el-dropdown trigger="click" @command="(cmd) => handleCommand(row, cmd)">
              <el-button type="warning" link size="small">
                更多<el-icon class="el-icon--right"><ArrowDown /></el-icon>
              </el-button>
              <template #dropdown>
                <el-dropdown-menu>
                  <el-dropdown-item :command="'edit'">编辑</el-dropdown-item>
                  <el-dropdown-item :command="'featured'">{{ row.is_featured ? '取消精选' : '设为精选' }}</el-dropdown-item>
                  <el-dropdown-item :command="3" :disabled="row.status === 3">下架</el-dropdown-item>
                  <el-dropdown-item :command="1" :disabled="row.status === 1">恢复</el-dropdown-item>
                  <el-dropdown-item :command="'delete'" divided>删除</el-dropdown-item>
                </el-dropdown-menu>
              </template>
            </el-dropdown>
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
    <el-dialog v-model="detailVisible" title="故事详情" width="800px" destroy-on-close>
      <div v-loading="detailLoading">
        <el-tabs v-if="detail" v-model="detailTab">
          <el-tab-pane label="基本信息" name="basic">
            <el-descriptions :column="2" border>
              <el-descriptions-item label="ID">{{ detail.id }}</el-descriptions-item>
              <el-descriptions-item label="标题">{{ detail.title }}</el-descriptions-item>
              <el-descriptions-item label="作者">{{ detail.author_name || `#${detail.user_id}` }}</el-descriptions-item>
              <el-descriptions-item label="用户ID">{{ detail.user_id }}</el-descriptions-item>
              <el-descriptions-item label="话题">
                <el-tag size="small">{{ topicText(detail.topic) }}</el-tag>
              </el-descriptions-item>
              <el-descriptions-item label="精选">
                <el-tag v-if="detail.is_featured" type="danger" size="small">精选</el-tag>
                <span v-else>否</span>
              </el-descriptions-item>
              <el-descriptions-item label="审核状态">
                <el-tag :type="auditTagType(detail.audit_status)" size="small">{{ auditText(detail.audit_status) }}</el-tag>
              </el-descriptions-item>
              <el-descriptions-item label="状态">
                <el-tag :type="statusTagType(detail.status)" size="small" effect="plain">{{ statusText(detail.status) }}</el-tag>
              </el-descriptions-item>
              <el-descriptions-item label="浏览">{{ detail.view_count }}</el-descriptions-item>
              <el-descriptions-item label="点赞">{{ detail.like_count }}</el-descriptions-item>
              <el-descriptions-item label="评论">{{ detail.comment_count }}</el-descriptions-item>
              <el-descriptions-item label="分享">{{ detail.share_count }}</el-descriptions-item>
              <el-descriptions-item v-if="detail.summary" label="摘要" :span="2">{{ detail.summary }}</el-descriptions-item>
              <el-descriptions-item v-if="detail.content" label="内容" :span="2">
                <div class="content-box">{{ detail.content }}</div>
              </el-descriptions-item>
              <el-descriptions-item v-if="detail.audit_reason" label="审核原因" :span="2">{{ detail.audit_reason }}</el-descriptions-item>
              <el-descriptions-item label="发布时间">{{ formatTime(detail.created_at) }}</el-descriptions-item>
              <el-descriptions-item label="更新时间">{{ formatTime(detail.updated_at) }}</el-descriptions-item>
            </el-descriptions>
          </el-tab-pane>
          <el-tab-pane :label="`图片 (${images.length})`" name="images">
            <div v-if="!images.length" class="empty-text">暂无图片</div>
            <div v-else class="images-grid">
              <el-image v-for="(img, idx) in images" :key="idx" :src="img" fit="cover" class="image-item" :preview-src-list="images" :initial-index="idx" preview-teleported />
            </div>
          </el-tab-pane>
        </el-tabs>
      </div>
      <template #footer><el-button @click="detailVisible = false">关闭</el-button></template>
    </el-dialog>

    <!-- 编辑弹窗 -->
    <el-dialog v-model="formVisible" :title="formTitle" width="640px">
      <el-form ref="formRef" :model="form" :rules="rules" label-width="100px">
        <el-form-item label="标题" prop="title">
          <el-input v-model="form.title" maxlength="128" />
        </el-form-item>
        <el-form-item label="话题">
          <el-select v-model="form.topic" style="width: 100%">
            <el-option v-for="(label, val) in topicMap" :key="val" :label="label" :value="val" />
          </el-select>
        </el-form-item>
        <el-form-item label="封面">
          <el-input v-model="form.cover_image" placeholder="封面图URL" />
        </el-form-item>
        <el-form-item label="摘要">
          <el-input v-model="form.summary" type="textarea" :rows="2" maxlength="200" />
        </el-form-item>
        <el-form-item label="内容">
          <el-input v-model="form.content" type="textarea" :rows="6" maxlength="5000" />
        </el-form-item>
        <el-form-item label="精选">
          <el-switch v-model="form.is_featured" />
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="form.status" style="width: 100%">
            <el-option label="正常" :value="1" />
            <el-option label="下架" :value="3" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="formVisible = false">取消</el-button>
        <el-button type="primary" :loading="formLoading" @click="onSubmit">确认</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Refresh, RefreshLeft, Search, ArrowDown, Check, Plus } from '@element-plus/icons-vue'
import request from '@/utils/request'
import { formatTime } from '@/utils/format'

const loading = ref(false)
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)
const list = ref([])
const selection = ref([])
const sortField = ref('created_at')
const sortOrder = ref('descending')

const filters = reactive({
  keyword: '', topic: '', audit_status: null, status: null, is_featured: null, dateRange: null
})

const stats = reactive({ total: 0, featured: 0, pendingAudit: 0, today: 0, totalView: 0, totalLike: 0 })

const topicMap = {
  match: '相亲故事', love: '恋爱心得', dating: '约会技巧', marriage: '婚姻感悟', single: '单身生活'
}
const topicText = (t) => topicMap[t] || t || '-'

const auditText = (s) => ({ 0: '待审核', 1: '已通过', 2: '已拒绝' }[s] || '-')
const auditTagType = (s) => ({ 0: 'warning', 1: 'success', 2: 'danger' }[s] || 'info')
const statusText = (s) => ({ 0: '禁用', 1: '正常', 3: '下架' }[s] || '-')
const statusTagType = (s) => ({ 0: 'danger', 1: 'success', 3: 'warning' }[s] || 'info')

const onSearch = () => { page.value = 1; loadList() }
const onReset = () => {
  filters.keyword = ''; filters.topic = ''; filters.audit_status = null
  filters.status = null; filters.is_featured = null; filters.dateRange = null
  page.value = 1; loadList()
}
const onSortChange = ({ prop, order }) => {
  sortField.value = prop || 'created_at'
  sortOrder.value = order || 'descending'
  loadList()
}
const onSelectionChange = (rows) => { selection.value = rows }

const loadList = async () => {
  loading.value = true
  try {
    const params = {
      page: page.value,
      page_size: pageSize.value,
      keyword: filters.keyword.trim() || undefined,
      topic: filters.topic || undefined,
      audit_status: filters.audit_status === null || filters.audit_status === '' ? undefined : filters.audit_status,
      status: filters.status === null || filters.status === '' ? undefined : filters.status,
      is_featured: filters.is_featured === null || filters.is_featured === '' ? undefined : filters.is_featured,
      sort: sortField.value,
      order: sortOrder.value
    }
    if (filters.dateRange && filters.dateRange.length === 2) {
      params.start_date = filters.dateRange[0]
      params.end_date = filters.dateRange[1]
    }
    const res = await request.get('/love/admin/stories', { params })
    const data = res.data || {}
    list.value = data.list || []
    total.value = data.total || 0
    calcStats()
  } catch (e) {
    list.value = []
    total.value = 0
  } finally {
    loading.value = false
  }
}

const calcStats = () => {
  const today = new Date().toISOString().slice(0, 10)
  stats.total = list.value.length
  stats.featured = list.value.filter((r) => r.is_featured).length
  stats.pendingAudit = list.value.filter((r) => r.audit_status === 0).length
  stats.today = list.value.filter((r) => (r.created_at || '').startsWith(today)).length
  stats.totalView = list.value.reduce((s, r) => s + (r.view_count || 0), 0)
  stats.totalLike = list.value.reduce((s, r) => s + (r.like_count || 0), 0)
}

// 详情
const detailVisible = ref(false)
const detailLoading = ref(false)
const detail = ref(null)
const detailTab = ref('basic')
const images = computed(() => {
  if (!detail.value) return []
  return Array.isArray(detail.value.images) ? detail.value.images : []
})

const openDetail = async (row) => {
  detail.value = row
  detailTab.value = 'basic'
  detailVisible.value = true
  detailLoading.value = true
  try {
    const res = await request.get(`/love/stories/${row.id}`)
    if (res.data) detail.value = res.data
  } catch (e) { /* keep */ }
  finally { detailLoading.value = false }
}

// 审核
const handleAudit = async (row, auditStatus) => {
  try {
    if (auditStatus === 2) {
      const { value } = await ElMessageBox.prompt('请输入拒绝原因', '拒绝审核', {
        inputType: 'textarea',
        inputPlaceholder: '拒绝原因'
      })
      await request.put(`/love/admin/stories/${row.id}/audit`, { audit_status: auditStatus, audit_reason: value || '' })
    } else {
      await ElMessageBox.confirm(`确定通过故事 "${row.title}" 的审核？`, '提示', { type: 'warning' })
      await request.put(`/love/admin/stories/${row.id}/audit`, { audit_status: auditStatus })
    }
    ElMessage.success('操作成功')
    await loadList()
  } catch (e) { /* cancel */ }
}

// 状态/精选/删除
const handleCommand = async (row, cmd) => {
  try {
    if (cmd === 'delete') {
      await ElMessageBox.confirm(`确认删除故事 "${row.title}"？删除后不可恢复！`, '危险操作', {
        type: 'error', confirmButtonText: '确认删除', cancelButtonText: '取消'
      })
      await request.delete(`/love/stories/${row.id}`)
      ElMessage.success('已删除')
      await loadList()
      return
    }
    if (cmd === 'edit') { openEdit(row); return }
    if (cmd === 'featured') {
      await request.put(`/love/admin/stories/${row.id}/featured`, { is_featured: !row.is_featured })
      ElMessage.success(row.is_featured ? '已取消精选' : '已设为精选')
      await loadList()
      return
    }
    const label = statusText(cmd)
    await ElMessageBox.confirm(`确定将故事设为「${label}」吗？`, '提示', { type: 'warning' })
    await request.put(`/love/admin/stories/${row.id}/status`, { status: cmd })
    ElMessage.success('操作成功')
    await loadList()
  } catch (e) { /* cancel */ }
}

// 批量审核
const onBatchAudit = async (auditStatus) => {
  try {
    await ElMessageBox.confirm(`确认批量审核通过 ${selection.value.length} 个故事？`, '批量审核', { type: 'warning' })
    await request.put('/love/admin/stories/batch-audit', {
      ids: selection.value.map((r) => r.id),
      audit_status: auditStatus
    })
    ElMessage.success('批量审核完成')
    await loadList()
  } catch (e) { /* cancel */ }
}

// 新建/编辑
const formVisible = ref(false)
const formLoading = ref(false)
const formRef = ref(null)
const isEdit = ref(false)
const formTitle = computed(() => isEdit.value ? '编辑故事' : '新建故事')
const form = reactive({
  id: null, title: '', topic: 'match', cover_image: '', summary: '', content: '',
  is_featured: false, status: 1
})
const rules = {
  title: [{ required: true, message: '请输入标题', trigger: 'blur' }]
}

const openCreate = () => {
  isEdit.value = false
  Object.assign(form, {
    id: null, title: '', topic: 'match', cover_image: '', summary: '', content: '',
    is_featured: false, status: 1
  })
  formVisible.value = true
}

const openEdit = (row) => {
  isEdit.value = true
  Object.assign(form, {
    id: row.id, title: row.title || '', topic: row.topic || 'match', cover_image: row.cover_image || '',
    summary: row.summary || '', content: row.content || '',
    is_featured: !!row.is_featured, status: row.status ?? 1
  })
  formVisible.value = true
}

const onSubmit = async () => {
  try {
    await formRef.value.validate()
    formLoading.value = true
    if (isEdit.value) {
      await request.put(`/love/stories/${form.id}`, form)
    } else {
      await request.post('/love/stories', form)
    }
    ElMessage.success(isEdit.value ? '更新成功' : '创建成功')
    formVisible.value = false
    await loadList()
  } catch (e) { /* ignore */ }
  finally { formLoading.value = false }
}

onMounted(() => loadList())
</script>

<style scoped>
.stat-row { margin-bottom: 16px; }
.stat-item { text-align: center; padding: 16px; }
.stat-value { font-size: 22px; font-weight: 600; color: #303133; }
.text-success { color: #67c23a; }
.text-warning { color: #e6a23c; }
.text-primary { color: #409eff; }
.text-danger { color: #f56c6c; }
.filter-form { margin-bottom: 12px; padding: 12px 16px; background: #fafafa; border-radius: 4px; }
.filter-form :deep(.el-form-item) { margin-bottom: 8px; margin-right: 12px; }
.toolbar { display: flex; justify-content: space-between; align-items: center; flex-wrap: wrap; gap: 8px; margin-bottom: 12px; }
.toolbar-left { display: flex; gap: 8px; flex-wrap: wrap; }
.toolbar-right { display: flex; gap: 8px; }
.cover-thumb { width: 50px; height: 50px; border-radius: 4px; border: 1px solid #ebeef5; }
.cover-empty { display: flex; align-items: center; justify-content: center; background: #fafafa; color: #909399; font-size: 12px; }
.title-cell { display: flex; flex-direction: column; gap: 4px; min-width: 0; }
.title-text { font-weight: 500; color: #303133; display: flex; align-items: center; gap: 6px; }
.title-desc { font-size: 12px; color: #909399; display: flex; align-items: center; gap: 6px; }
.text-muted { color: #909399; }
.images-grid { display: flex; flex-wrap: wrap; gap: 8px; }
.image-item { width: 120px; height: 120px; border-radius: 4px; border: 1px solid #ebeef5; }
.content-box { white-space: pre-wrap; word-break: break-all; max-height: 300px; overflow-y: auto; }
.empty-text { color: #909399; text-align: center; padding: 32px 0; }
.pagination-wrap { margin-top: 16px; display: flex; justify-content: flex-end; }
</style>
