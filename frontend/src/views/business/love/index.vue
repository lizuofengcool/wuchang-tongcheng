<template>
  <div class="app-container">
    <!-- 顶部统计卡片 -->
    <el-row :gutter="16" class="stat-row">
      <el-col :xs="12" :sm="8" :md="4">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-icon" style="background: #409eff">
            <el-icon :size="22"><User /></el-icon>
          </div>
          <div class="stat-content">
            <div class="stat-value">{{ stats.total }}</div>
            <div class="stat-label">总会员数</div>
          </div>
        </el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="4">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-icon" style="background: #67c23a">
            <el-icon :size="22"><CirclePlus /></el-icon>
          </div>
          <div class="stat-content">
            <div class="stat-value">{{ stats.todayNew }}</div>
            <div class="stat-label">今日新增</div>
          </div>
        </el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="4">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-icon" style="background: #e6a23c">
            <el-icon :size="22"><Clock /></el-icon>
          </div>
          <div class="stat-content">
            <div class="stat-value">{{ stats.pendingAudit }}</div>
            <div class="stat-label">待审核</div>
          </div>
        </el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="4">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-icon" style="background: #13c2c2">
            <el-icon :size="22"><Star /></el-icon>
          </div>
          <div class="stat-content">
            <div class="stat-value">{{ stats.verified }}</div>
            <div class="stat-label">实名认证</div>
          </div>
        </el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="4">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-icon" style="background: #722ed1">
            <el-icon :size="22"><Promotion /></el-icon>
          </div>
          <div class="stat-content">
            <div class="stat-value">{{ stats.featured }}</div>
            <div class="stat-label">精选会员</div>
          </div>
        </el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="4">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-icon" style="background: #f56c6c">
            <el-icon :size="22"><Warning /></el-icon>
          </div>
          <div class="stat-content">
            <div class="stat-value">{{ stats.violation }}</div>
            <div class="stat-label">违规数</div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 高级筛选区 -->
    <div class="page-card">
      <el-form :inline="true" :model="filters" class="filter-form" @submit.prevent>
        <el-form-item label="关键词">
          <el-input
            v-model="filters.keyword"
            placeholder="昵称/ID/手机号"
            clearable
            style="width: 200px"
            :prefix-icon="Search"
            @keyup.enter="onSearch"
          />
        </el-form-item>
        <el-form-item label="性别">
          <el-select v-model="filters.gender" placeholder="全部" clearable style="width: 110px" @change="onSearch">
            <el-option label="男" value="male" />
            <el-option label="女" value="female" />
          </el-select>
        </el-form-item>
        <el-form-item label="年龄">
          <el-input-number v-model="filters.min_age" :min="0" :max="100" :controls="false" placeholder="最小" style="width: 90px" />
          <span style="margin: 0 4px">-</span>
          <el-input-number v-model="filters.max_age" :min="0" :max="100" :controls="false" placeholder="最大" style="width: 90px" />
        </el-form-item>
        <el-form-item label="实名认证">
          <el-select v-model="filters.id_verified" placeholder="全部" clearable style="width: 110px" @change="onSearch">
            <el-option label="已认证" :value="1" />
            <el-option label="未认证" :value="0" />
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
            <el-option label="禁用" :value="0" />
            <el-option label="下架" :value="3" />
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

      <!-- 工具栏 -->
      <div class="toolbar">
        <div class="toolbar-left">
          <el-button :icon="Refresh" @click="loadList">刷新</el-button>
          <el-button type="success" :icon="Check" :disabled="!selection.length" @click="onBatchAudit(1)">批量通过</el-button>
          <el-button type="warning" :icon="Bottom" :disabled="!selection.length" @click="onBatchStatus(3)">批量下架</el-button>
          <el-button type="danger" :icon="Delete" :disabled="!selection.length" @click="onBatchStatus(0)">批量禁用</el-button>
        </div>
        <div class="toolbar-right">
          <el-button type="primary" :icon="Plus" @click="onCreate">新建会员</el-button>
        </div>
      </div>

      <!-- 表格 -->
      <el-table
        v-loading="loading"
        :data="list"
        border
        stripe
        @selection-change="onSelectionChange"
        @sort-change="onSortChange"
      >
        <el-table-column type="selection" width="44" fixed="left" />
        <el-table-column prop="id" label="ID" width="70" sortable="custom" fixed="left" />
        <el-table-column label="头像" width="70">
          <template #default="{ row }">
            <el-image
              v-if="row.avatar"
              :src="row.avatar"
              fit="cover"
              class="cover-thumb"
              :preview-src-list="[row.avatar]"
              preview-teleported
            />
            <div v-else class="cover-thumb cover-empty">无图</div>
          </template>
        </el-table-column>
        <el-table-column label="昵称/标签" min-width="220">
          <template #default="{ row }">
            <div class="title-cell">
              <div class="title-text">
                <el-link type="primary" :underline="'never'" @click="openDetail(row)">{{ row.nickname || `会员#${row.id}` }}</el-link>
                <el-tag v-if="row.is_featured" type="danger" size="small" effect="dark">精</el-tag>
                <el-tag v-if="row.is_picked" type="success" size="small" effect="dark">荐</el-tag>
                <el-tag v-if="row.id_verified" type="primary" size="small">实名</el-tag>
              </div>
              <div class="title-desc">
                <span>{{ genderText(row.gender) }}</span>
                <span>{{ row.age }}岁</span>
                <span v-if="row.height">{{ row.height }}cm</span>
                <span v-if="row.education">{{ educationText(row.education) }}</span>
              </div>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="地区" width="100">
          <template #default="{ row }">
            <span v-if="row.location_city">{{ row.location_city }}</span>
            <span v-else class="text-muted">-</span>
          </template>
        </el-table-column>
        <el-table-column label="职业" width="120">
          <template #default="{ row }">
            <span v-if="row.occupation">{{ row.occupation }}</span>
            <span v-else class="text-muted">-</span>
          </template>
        </el-table-column>
        <el-table-column label="浏览" width="70" prop="view_count" sortable="custom" />
        <el-table-column label="喜欢" width="70" prop="like_count" sortable="custom" />
        <el-table-column label="匹配" width="70" prop="match_count" />
        <el-table-column label="访客" width="70" prop="visitor_count" />
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
        <el-table-column label="注册时间" width="160" prop="created_at" sortable="custom">
          <template #default="{ row }">{{ formatTime(row.created_at) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="260" fixed="right">
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
            <el-dropdown trigger="click" @command="(cmd) => handleStatusCommand(row, cmd)">
              <el-button type="warning" link size="small">
                更多<el-icon class="el-icon--right"><ArrowDown /></el-icon>
              </el-button>
              <template #dropdown>
                <el-dropdown-menu>
                  <el-dropdown-item :command="'edit'">编辑</el-dropdown-item>
                  <el-dropdown-item :command="1" :disabled="row.status === 1">设为正常</el-dropdown-item>
                  <el-dropdown-item :command="3" :disabled="row.status === 3">下架</el-dropdown-item>
                  <el-dropdown-item :command="0" :disabled="row.status === 0">禁用</el-dropdown-item>
                  <el-dropdown-item :command="'featured'">设为精选</el-dropdown-item>
                  <el-dropdown-item :command="'picked'">设为推荐</el-dropdown-item>
                  <el-dropdown-item :command="'delete'" divided>删除</el-dropdown-item>
                </el-dropdown-menu>
              </template>
            </el-dropdown>
          </template>
        </el-table-column>
      </el-table>

      <!-- 分页 -->
      <div class="pagination-wrap">
        <el-pagination
          v-model:current-page="page"
          v-model:page-size="pageSize"
          :total="total"
          :page-sizes="[10, 20, 50, 100]"
          layout="total, sizes, prev, pager, next, jumper"
          background
          @current-change="loadList"
          @size-change="loadList"
        />
      </div>
    </div>

    <!-- 详情弹窗 -->
    <el-dialog v-model="detailVisible" title="会员详情" width="900px" @close="onDetailClose" destroy-on-close>
      <div v-loading="detailLoading">
        <el-tabs v-if="detail" v-model="detailTab">
          <el-tab-pane label="基本信息" name="basic">
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
              <el-descriptions-item label="年龄要求">{{ detail.looking_for_age_min }}-{{ detail.looking_for_age_max }} 岁</el-descriptions-item>
              <el-descriptions-item label="实名认证">
                <el-tag v-if="detail.id_verified" type="success" size="small">已认证</el-tag>
                <el-tag v-else type="info" size="small">未认证</el-tag>
              </el-descriptions-item>
              <el-descriptions-item label="视频认证">
                <el-tag v-if="detail.video_verified" type="success" size="small">已认证</el-tag>
                <el-tag v-else type="info" size="small">未认证</el-tag>
              </el-descriptions-item>
              <el-descriptions-item label="精选">
                <el-tag v-if="detail.is_featured" type="danger" size="small">精选</el-tag>
                <span v-else>否</span>
              </el-descriptions-item>
              <el-descriptions-item label="推荐">
                <el-tag v-if="detail.is_picked" type="success" size="small">推荐</el-tag>
                <span v-else>否</span>
              </el-descriptions-item>
              <el-descriptions-item label="审核状态">
                <el-tag :type="auditTagType(detail.audit_status)" size="small">{{ auditText(detail.audit_status) }}</el-tag>
              </el-descriptions-item>
              <el-descriptions-item label="状态">
                <el-tag :type="statusTagType(detail.status)" size="small" effect="plain">{{ statusText(detail.status) }}</el-tag>
              </el-descriptions-item>
              <el-descriptions-item v-if="detail.signature" label="个性签名" :span="3">{{ detail.signature }}</el-descriptions-item>
              <el-descriptions-item v-if="detail.about_me" label="关于我" :span="3">
                <div class="content-box">{{ detail.about_me }}</div>
              </el-descriptions-item>
              <el-descriptions-item v-if="detail.looking_for" label="期望对象" :span="3">
                <div class="content-box">{{ detail.looking_for }}</div>
              </el-descriptions-item>
              <el-descriptions-item v-if="detail.audit_reason" label="审核原因" :span="3">{{ detail.audit_reason }}</el-descriptions-item>
              <el-descriptions-item label="浏览量">{{ detail.view_count }}</el-descriptions-item>
              <el-descriptions-item label="喜欢数">{{ detail.like_count }}</el-descriptions-item>
              <el-descriptions-item label="匹配数">{{ detail.match_count }}</el-descriptions-item>
              <el-descriptions-item label="访客数">{{ detail.visitor_count }}</el-descriptions-item>
              <el-descriptions-item label="注册时间">{{ formatTime(detail.created_at) }}</el-descriptions-item>
              <el-descriptions-item label="更新时间">{{ formatTime(detail.updated_at) }}</el-descriptions-item>
            </el-descriptions>
          </el-tab-pane>

          <el-tab-pane :label="`相册 (${photos.length})`" name="photos">
            <div v-if="!photos.length" class="empty-text">暂无照片</div>
            <div v-else class="images-grid">
              <el-image
                v-for="(img, idx) in photos"
                :key="idx"
                :src="img"
                fit="cover"
                class="image-item"
                :preview-src-list="photos"
                :initial-index="idx"
                preview-teleported
              />
            </div>
          </el-tab-pane>

          <el-tab-pane label="操作日志" name="logs">
            <div class="empty-text">操作日志功能开发中</div>
          </el-tab-pane>
        </el-tabs>
      </div>
      <template #footer>
        <el-button @click="detailVisible = false">关闭</el-button>
      </template>
    </el-dialog>

    <!-- 编辑弹窗 -->
    <el-dialog v-model="formVisible" :title="formTitle" width="640px">
      <el-form ref="formRef" :model="form" :rules="rules" label-width="100px">
        <el-form-item label="昵称" prop="nickname">
          <el-input v-model="form.nickname" maxlength="32" placeholder="昵称" />
        </el-form-item>
        <el-form-item label="性别">
          <el-radio-group v-model="form.gender">
            <el-radio value="male">男</el-radio>
            <el-radio value="female">女</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="年龄">
          <el-input-number v-model="form.age" :min="18" :max="100" />
        </el-form-item>
        <el-form-item label="身高">
          <el-input-number v-model="form.height" :min="140" :max="220" />
        </el-form-item>
        <el-form-item label="学历">
          <el-select v-model="form.education" style="width: 100%">
            <el-option v-for="(label, val) in educationMap" :key="val" :label="label" :value="val" />
          </el-select>
        </el-form-item>
        <el-form-item label="收入">
          <el-select v-model="form.income" style="width: 100%">
            <el-option v-for="(label, val) in incomeMap" :key="val" :label="label" :value="val" />
          </el-select>
        </el-form-item>
        <el-form-item label="职业">
          <el-input v-model="form.occupation" maxlength="64" />
        </el-form-item>
        <el-form-item label="所在城市">
          <el-input v-model="form.location_city" maxlength="64" />
        </el-form-item>
        <el-form-item label="个性签名">
          <el-input v-model="form.signature" maxlength="200" />
        </el-form-item>
        <el-form-item label="关于我">
          <el-input v-model="form.about_me" type="textarea" :rows="3" maxlength="1000" />
        </el-form-item>
        <el-form-item label="精选">
          <el-switch v-model="form.is_featured" />
        </el-form-item>
        <el-form-item label="推荐">
          <el-switch v-model="form.is_picked" />
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="form.status" style="width: 100%">
            <el-option label="正常" :value="1" />
            <el-option label="禁用" :value="0" />
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
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  Refresh, RefreshLeft, Search, ArrowDown, Check, Bottom, Delete, Plus,
  User, CirclePlus, Clock, Star, Promotion, Warning
} from '@element-plus/icons-vue'
import request from '@/utils/request'
import { formatTime } from '@/utils/format'

const router = useRouter()

// ===== 列表 =====
const loading = ref(false)
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)
const list = ref([])
const selection = ref([])
const sortField = ref('created_at')
const sortOrder = ref('descending')

// ===== 统计卡片 =====
const stats = reactive({
  total: 0, todayNew: 0, pendingAudit: 0, verified: 0, featured: 0, violation: 0
})

const loadStats = async () => {
  try {
    const res = await request.get('/love/admin/loves', { params: { page: 1, page_size: 1 } })
    const d = (res.data && res.data.stats) || {}
    stats.total = d.total || res.data?.total || 0
    stats.pendingAudit = d.pending_audit || 0
    stats.featured = d.featured || 0
    stats.verified = d.verified || 0
    stats.violation = d.violation || 0
    stats.todayNew = d.today_new || 0
  } catch (e) {
    // 静默
  }
}

// ===== 筛选 =====
const filters = reactive({
  keyword: '', gender: '', min_age: undefined, max_age: undefined,
  id_verified: null, audit_status: null, status: null, dateRange: null
})

const onSearch = () => { page.value = 1; loadList() }

const onReset = () => {
  filters.keyword = ''
  filters.gender = ''
  filters.min_age = undefined
  filters.max_age = undefined
  filters.id_verified = null
  filters.audit_status = null
  filters.status = null
  filters.dateRange = null
  page.value = 1
  loadList()
}

const onSortChange = ({ prop, order }) => {
  sortField.value = prop || 'created_at'
  sortOrder.value = order || 'descending'
  loadList()
}

const onSelectionChange = (rows) => { selection.value = rows }

// ===== 状态格式化 =====
const auditText = (s) => ({ 0: '待审核', 1: '已通过', 2: '已拒绝' }[s] || '-')
const auditTagType = (s) => ({ 0: 'warning', 1: 'success', 2: 'danger' }[s] || 'info')
const statusText = (s) => ({ 0: '禁用', 1: '正常', 3: '下架' }[s] || '-')
const statusTagType = (s) => ({ 0: 'danger', 1: 'success', 3: 'warning' }[s] || 'info')
const genderText = (g) => ({ male: '男', female: '女' }[g] || '未知')
const educationMap = { high_school: '高中', college: '大专', bachelor: '本科', master: '硕士', doctor: '博士' }
const educationText = (e) => educationMap[e] || e || '-'
const incomeMap = {
  low: '10万以下', medium: '10-30万', high: '30-50万', very_high: '50万以上'
}
const incomeText = (i) => incomeMap[i] || i || '-'
const relationshipMap = { single: '单身', divorced: '离异', widowed: '丧偶' }
const relationshipText = (r) => relationshipMap[r] || r || '-'

// ===== 查询 =====
const loadList = async () => {
  loading.value = true
  try {
    const params = {
      page: page.value,
      page_size: pageSize.value,
      keyword: filters.keyword.trim() || undefined,
      gender: filters.gender || undefined,
      min_age: filters.min_age || undefined,
      max_age: filters.max_age || undefined,
      id_verified: filters.id_verified === null || filters.id_verified === '' ? undefined : filters.id_verified,
      audit_status: filters.audit_status === null || filters.audit_status === '' ? undefined : filters.audit_status,
      status: filters.status === null || filters.status === '' ? undefined : filters.status,
      sort: sortField.value,
      order: sortOrder.value
    }
    if (filters.dateRange && filters.dateRange.length === 2) {
      params.start_date = filters.dateRange[0]
      params.end_date = filters.dateRange[1]
    }
    const res = await request.get('/love/admin/loves', { params })
    const data = res.data || {}
    list.value = data.list || []
    total.value = data.total || 0
    if (data.stats) {
      Object.assign(stats, data.stats)
    }
  } catch (e) {
    list.value = []
    total.value = 0
  } finally {
    loading.value = false
  }
}

// ===== 详情弹窗 =====
const detailVisible = ref(false)
const detailLoading = ref(false)
const detail = ref(null)
const detailTab = ref('basic')

const photos = computed(() => {
  if (!detail.value) return []
  return Array.isArray(detail.value.photos) ? detail.value.photos : (detail.value.avatar ? [detail.value.avatar] : [])
})

const openDetail = async (row) => {
  detail.value = row
  detailTab.value = 'basic'
  detailVisible.value = true
  detailLoading.value = true
  try {
    const res = await request.get(`/love/admin/loves/${row.id}`)
    if (res.data) detail.value = res.data
  } catch (e) {
    // 接口失败保持现状
  } finally {
    detailLoading.value = false
  }
}

const onDetailClose = () => { detail.value = null }

// ===== 审核 =====
const handleAudit = async (row, auditStatus) => {
  const action = auditStatus === 1 ? '通过' : '拒绝'
  try {
    if (auditStatus === 2) {
      const { value } = await ElMessageBox.prompt('请输入拒绝原因（可选）', '拒绝审核', {
        confirmButtonText: '确定拒绝',
        cancelButtonText: '取消',
        inputType: 'textarea',
        inputPlaceholder: '拒绝原因（可不填）'
      })
      await request.put(`/love/admin/loves/${row.id}/audit`, { audit_status: auditStatus, audit_reason: value || '' })
    } else {
      await ElMessageBox.confirm(`确定${action}会员 "${row.nickname}" 的审核吗？`, '提示', { type: 'warning' })
      await request.put(`/love/admin/loves/${row.id}/audit`, { audit_status: auditStatus })
    }
    ElMessage.success(`已${action}`)
    await loadList()
  } catch (e) {
    // 取消
  }
}

// ===== 状态变更 =====
const handleStatusCommand = async (row, cmd) => {
  try {
    if (cmd === 'delete') {
      await ElMessageBox.confirm(`确定删除会员 "${row.nickname}" 吗？删除后不可恢复！`, '危险操作', {
        type: 'error',
        confirmButtonText: '确认删除',
        cancelButtonText: '取消'
      })
      await request.delete(`/love/admin/loves/${row.id}`)
      ElMessage.success('已删除')
      await loadList()
      return
    }
    if (cmd === 'edit') {
      openEdit(row)
      return
    }
    if (cmd === 'featured') {
      await request.put(`/love/admin/loves/${row.id}/featured`, { is_featured: !row.is_featured })
      ElMessage.success(row.is_featured ? '已取消精选' : '已设为精选')
      await loadList()
      return
    }
    if (cmd === 'picked') {
      await request.put(`/love/admin/loves/${row.id}/picked`, { is_picked: !row.is_picked })
      ElMessage.success(row.is_picked ? '已取消推荐' : '已设为推荐')
      await loadList()
      return
    }
    const label = statusText(cmd)
    await ElMessageBox.confirm(`确定将会员 "${row.nickname}" 设为「${label}」吗？`, '提示', { type: 'warning' })
    await request.put(`/love/admin/loves/${row.id}/status`, { status: cmd })
    ElMessage.success('状态更新成功')
    await loadList()
  } catch (e) {
    // 取消
  }
}

// ===== 批量操作 =====
const onBatchAudit = async (auditStatus) => {
  try {
    await ElMessageBox.confirm(`确认批量审核通过 ${selection.value.length} 个会员？`, '批量审核', { type: 'warning' })
    await request.put('/love/admin/loves/batch-audit', {
      ids: selection.value.map((r) => r.id),
      audit_status: auditStatus
    })
    ElMessage.success('批量审核完成')
    await loadList()
  } catch (e) {
    // 取消
  }
}

const onBatchStatus = async (status) => {
  try {
    const label = statusText(status)
    await ElMessageBox.confirm(`确认批量将 ${selection.value.length} 个会员设为「${label}」？`, '批量状态变更', { type: 'warning' })
    await request.put('/love/admin/loves/batch-status', {
      ids: selection.value.map((r) => r.id),
      status
    })
    ElMessage.success('批量操作完成')
    await loadList()
  } catch (e) {
    // 取消
  }
}

// ===== 新建/编辑 =====
const formVisible = ref(false)
const formLoading = ref(false)
const formRef = ref(null)
const isEdit = ref(false)
const formTitle = computed(() => isEdit.value ? '编辑会员' : '新建会员')
const form = reactive({
  id: null, nickname: '', gender: 'male', age: 25, height: 170,
  education: 'bachelor', income: 'medium', occupation: '', location_city: '',
  signature: '', about_me: '', is_featured: false, is_picked: false, status: 1
})
const rules = {
  nickname: [{ required: true, message: '请输入昵称', trigger: 'blur' }]
}

const onCreate = () => {
  isEdit.value = false
  Object.assign(form, {
    id: null, nickname: '', gender: 'male', age: 25, height: 170,
    education: 'bachelor', income: 'medium', occupation: '', location_city: '',
    signature: '', about_me: '', is_featured: false, is_picked: false, status: 1
  })
  formVisible.value = true
}

const openEdit = (row) => {
  isEdit.value = true
  Object.assign(form, {
    id: row.id, nickname: row.nickname || '', gender: row.gender || 'male',
    age: row.age || 25, height: row.height || 170, education: row.education || 'bachelor',
    income: row.income || 'medium', occupation: row.occupation || '',
    location_city: row.location_city || '', signature: row.signature || '',
    about_me: row.about_me || '', is_featured: !!row.is_featured,
    is_picked: !!row.is_picked, status: row.status ?? 1
  })
  formVisible.value = true
}

const onSubmit = async () => {
  try {
    await formRef.value.validate()
    formLoading.value = true
    if (isEdit.value) {
      await request.put(`/love/${form.id}`, form)
    } else {
      await request.post('/love', form)
    }
    ElMessage.success(isEdit.value ? '更新成功' : '创建成功')
    formVisible.value = false
    await loadList()
  } catch (e) {
    // 校验失败
  } finally {
    formLoading.value = false
  }
}

onMounted(async () => {
  await Promise.all([loadStats()])
  await loadList()
})
</script>

<style scoped>
.stat-row { margin-bottom: 16px; }
.stat-card { display: flex; align-items: center; }
.stat-card :deep(.el-card__body) {
  display: flex; align-items: center; gap: 14px; padding: 16px; width: 100%;
}
.stat-icon {
  width: 44px; height: 44px; border-radius: 8px;
  display: flex; align-items: center; justify-content: center;
  color: #fff; flex-shrink: 0;
}
.stat-content { flex: 1; min-width: 0; }
.stat-value { font-size: 22px; font-weight: 600; color: #303133; }
.stat-label { font-size: 12px; color: #909399; margin-top: 2px; }

.filter-form {
  margin-bottom: 12px; padding: 12px 16px;
  background: #fafafa; border-radius: 4px;
}
.filter-form :deep(.el-form-item) { margin-bottom: 8px; margin-right: 12px; }

.toolbar {
  display: flex; justify-content: space-between; align-items: center;
  flex-wrap: wrap; gap: 8px; margin-bottom: 12px;
}
.toolbar-left { display: flex; gap: 8px; flex-wrap: wrap; }
.toolbar-right { display: flex; gap: 8px; }

.cover-thumb {
  width: 50px; height: 50px; border-radius: 4px; border: 1px solid #ebeef5;
}
.cover-empty {
  display: flex; align-items: center; justify-content: center;
  background: #fafafa; color: #909399; font-size: 12px;
}
.title-cell { display: flex; flex-direction: column; gap: 4px; min-width: 0; }
.title-text {
  font-weight: 500; color: #303133;
  display: flex; align-items: center; gap: 6px;
}
.title-text .el-link { max-width: 100%; }
.title-desc {
  font-size: 12px; color: #909399;
  display: flex; align-items: center; gap: 6px;
}
.text-muted { color: #909399; }

.images-grid { display: flex; flex-wrap: wrap; gap: 8px; }
.image-item {
  width: 120px; height: 120px; border-radius: 4px; border: 1px solid #ebeef5;
}
.content-box {
  white-space: pre-wrap; word-break: break-all;
  max-height: 200px; overflow-y: auto;
}
.empty-text { color: #909399; text-align: center; padding: 32px 0; }
.pagination-wrap {
  margin-top: 16px; display: flex; justify-content: flex-end;
}

@media (max-width: 1200px) {
  .filter-form :deep(.el-form-item) { margin-right: 8px; }
}
</style>
