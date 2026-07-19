<template>
  <div class="app-container">
    <!-- 顶部统计卡片 -->
    <el-row :gutter="20" class="stat-row">
      <el-col :xs="12" :sm="8" :md="4">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-icon" style="background: #409eff">
            <el-icon :size="24"><Menu /></el-icon>
          </div>
          <div class="stat-content">
            <div class="stat-value">{{ stats.total }}</div>
            <div class="stat-label">总模块数</div>
          </div>
        </el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="4">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-icon" style="background: #67c23a">
            <el-icon :size="24"><CircleCheck /></el-icon>
          </div>
          <div class="stat-content">
            <div class="stat-value">{{ stats.enabled }}</div>
            <div class="stat-label">已启用</div>
          </div>
        </el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="4">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-icon" style="background: #f56c6c">
            <el-icon :size="24"><CircleClose /></el-icon>
          </div>
          <div class="stat-content">
            <div class="stat-value">{{ stats.disabled }}</div>
            <div class="stat-label">已禁用</div>
          </div>
        </el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="4">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-icon" style="background: #e6a23c">
            <el-icon :size="24"><Coin /></el-icon>
          </div>
          <div class="stat-content">
            <div class="stat-value">{{ stats.middleware }}</div>
            <div class="stat-label">中台数</div>
          </div>
        </el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="4">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-icon" style="background: #909399">
            <el-icon :size="24"><ShoppingBag /></el-icon>
          </div>
          <div class="stat-content">
            <div class="stat-value">{{ stats.business }}</div>
            <div class="stat-label">垂直业务数</div>
          </div>
        </el-card>
      </el-col>
      <el-col :xs="12" :sm="8" :md="4">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-icon" style="background: #9c27b0">
            <el-icon :size="24"><Setting /></el-icon>
          </div>
          <div class="stat-content">
            <div class="stat-value">{{ stats.system }}</div>
            <div class="stat-label">系统模块数</div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 筛选 + 表格 -->
    <div class="page-card">
      <div class="toolbar">
        <div class="toolbar-left">
          <el-button type="primary" :icon="Refresh" @click="loadModules">刷新</el-button>
        </div>
        <div class="toolbar-right">
          <el-select
            v-model="filters.category"
            placeholder="分类"
            clearable
            style="width: 160px"
          >
            <el-option label="系统模块" value="system" />
            <el-option label="业务模块" value="business" />
            <el-option label="营销模块" value="marketing" />
            <el-option label="用户模块" value="user" />
            <el-option label="社区模块" value="community" />
            <el-option label="中台模块" value="middleware" />
          </el-select>
          <el-select
            v-model="filters.status"
            placeholder="状态"
            clearable
            style="width: 140px; margin-left: 8px"
          >
            <el-option label="全部" value="" />
            <el-option label="启用" value="1" />
            <el-option label="禁用" value="0" />
          </el-select>
          <el-input
            v-model="filters.keyword"
            placeholder="模块名/展示名"
            clearable
            style="width: 220px; margin-left: 8px"
            :prefix-icon="Search"
            @keyup.enter="loadModules"
          />
          <el-button
            type="primary"
            :icon="Search"
            style="margin-left: 8px"
            @click="loadModules"
          >搜索</el-button>
        </div>
      </div>

      <el-table v-loading="loading" :data="filteredList" border stripe>
        <el-table-column prop="name" label="模块名" min-width="160" show-overflow-tooltip />
        <el-table-column prop="display_name" label="展示名" min-width="140" show-overflow-tooltip>
          <template #default="{ row }">
            <span>{{ row.display_name || row.name }}</span>
          </template>
        </el-table-column>
        <el-table-column label="分类" width="120">
          <template #default="{ row }">
            <el-tag :type="categoryTag(row.category)" size="small">
              {{ categoryLabel(row.category) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="version" label="版本" width="100">
          <template #default="{ row }">
            <span>v{{ row.version || '1.0.0' }}</span>
          </template>
        </el-table-column>
        <el-table-column label="依赖" min-width="180" show-overflow-tooltip>
          <template #default="{ row }">
            <span v-if="!row.dependencies || !row.dependencies.length" class="text-muted">无</span>
            <el-tag
              v-for="dep in (row.dependencies || [])"
              :key="dep"
              size="small"
              type="info"
              class="dep-tag"
            >{{ dep }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-switch
              :model-value="row.enabled"
              :loading="row._switching"
              @change="(val) => handleToggle(row, val)"
            />
          </template>
        </el-table-column>
        <el-table-column label="操作" width="100" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="openDetail(row)">详情</el-button>
          </template>
        </el-table-column>
      </el-table>

      <div class="pagination-wrap">
        <el-pagination
          v-model:current-page="pagination.page"
          v-model:page-size="pagination.pageSize"
          :total="pagination.total"
          :page-sizes="[10, 20, 50, 100]"
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="loadModules"
          @current-change="loadModules"
        />
      </div>
    </div>

    <!-- 模块详情弹窗 -->
    <el-dialog v-model="detailVisible" title="模块详情" width="640px">
      <el-descriptions v-if="current" :column="2" border>
        <el-descriptions-item label="模块名">{{ current.name }}</el-descriptions-item>
        <el-descriptions-item label="展示名">{{ current.display_name || current.name }}</el-descriptions-item>
        <el-descriptions-item label="分类">
          <el-tag :type="categoryTag(current.category)" size="small">
            {{ categoryLabel(current.category) }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="版本">v{{ current.version || '1.0.0' }}</el-descriptions-item>
        <el-descriptions-item label="状态">
          <el-tag :type="current.enabled ? 'success' : 'danger'" size="small">
            {{ current.enabled ? '启用' : '禁用' }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="作者">{{ current.author || '-' }}</el-descriptions-item>
        <el-descriptions-item label="主页" :span="2">
          <a v-if="current.homepage" :href="current.homepage" target="_blank" class="link">
            {{ current.homepage }}
          </a>
          <span v-else class="text-muted">-</span>
        </el-descriptions-item>
        <el-descriptions-item label="描述" :span="2">
          {{ current.description || '无描述' }}
        </el-descriptions-item>
        <el-descriptions-item label="依赖模块" :span="2">
          <span v-if="!current.dependencies || !current.dependencies.length" class="text-muted">无</span>
          <el-tag
            v-for="dep in (current.dependencies || [])"
            :key="dep"
            size="small"
            type="info"
            class="dep-tag"
          >{{ dep }}</el-tag>
        </el-descriptions-item>
      </el-descriptions>
      <template #footer>
        <el-button @click="detailVisible = false">关闭</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Refresh, Search } from '@element-plus/icons-vue'
import { getModules, enableModule, disableModule } from '@/api/module'

const loading = ref(false)
const list = ref([])

const filters = reactive({
  category: '',
  status: '',
  keyword: ''
})

const pagination = reactive({
  page: 1,
  pageSize: 20,
  total: 0
})

// ====== 统计卡片 ======
const stats = computed(() => {
  const total = list.value.length
  const enabled = list.value.filter((m) => m.enabled).length
  const disabled = total - enabled
  const middleware = list.value.filter((m) => m.category === 'middleware').length
  const business = list.value.filter((m) => m.category === 'business').length
  const system = list.value.filter((m) => m.category === 'system').length
  return { total, enabled, disabled, middleware, business, system }
})

// ====== 前端筛选（接口若已支持筛选则后端会返回筛选后结果）======
const filteredList = computed(() => {
  let arr = list.value
  if (filters.category) {
    arr = arr.filter((m) => m.category === filters.category)
  }
  if (filters.status !== '') {
    const enabled = filters.status === '1'
    arr = arr.filter((m) => Boolean(m.enabled) === enabled)
  }
  const kw = filters.keyword.trim().toLowerCase()
  if (kw) {
    arr = arr.filter(
      (m) =>
        (m.name || '').toLowerCase().includes(kw) ||
        (m.display_name || '').toLowerCase().includes(kw)
    )
  }
  return arr
})

// ====== 加载模块列表 ======
const loadModules = async () => {
  loading.value = true
  try {
    const res = await getModules({
      page: pagination.page,
      page_size: pagination.pageSize,
      category: filters.category || undefined,
      enabled: filters.status !== '' ? filters.status : undefined,
      keyword: filters.keyword || undefined
    })
    // 兼容两种返回格式：分页 { list, total } 或纯数组
    const data = res.data
    if (Array.isArray(data)) {
      list.value = data.map((m) => ({ ...m, _switching: false }))
      pagination.total = data.length
    } else if (data && Array.isArray(data.list)) {
      list.value = data.list.map((m) => ({ ...m, _switching: false }))
      pagination.total = data.total || data.list.length
    } else {
      list.value = []
      pagination.total = 0
    }
  } catch (e) {
    // 后端不可达时清空列表（Agent 3 不依赖后端可用性）
    list.value = []
    pagination.total = 0
  } finally {
    loading.value = false
  }
}

// ====== 启停切换 ======
const handleToggle = async (row, val) => {
  const action = val ? '启用' : '禁用'
  try {
    await ElMessageBox.confirm(
      `${action}模块会影响全端入口和接口，确认操作？`,
      `${action}模块`,
      {
        confirmButtonText: `确认${action}`,
        cancelButtonText: '取消',
        type: 'warning'
      }
    )
  } catch (e) {
    // 用户取消，不做任何变更
    return
  }

  row._switching = true
  try {
    if (val) {
      await enableModule(row.name)
    } else {
      await disableModule(row.name)
    }
    row.enabled = val
    ElMessage.success(`已${action}模块：${row.display_name || row.name}`)
    // 启停后刷新列表，保证状态一致
    await loadModules()
  } catch (e) {
    // 错误信息已由 request 拦截器统一提示
  } finally {
    row._switching = false
  }
}

// ====== 详情弹窗 ======
const detailVisible = ref(false)
const current = ref(null)
const openDetail = (row) => {
  current.value = row
  detailVisible.value = true
}

// ====== 分类辅助 ======
const categoryLabel = (c) => {
  const map = {
    system: '系统模块',
    business: '业务模块',
    marketing: '营销模块',
    user: '用户模块',
    community: '社区模块',
    middleware: '中台模块'
  }
  return map[c] || c || '-'
}
const categoryTag = (c) => {
  const map = {
    system: '',
    business: 'success',
    marketing: 'warning',
    user: 'info',
    community: 'info',
    middleware: 'danger'
  }
  return map[c] || ''
}

onMounted(() => {
  loadModules()
})
</script>

<style scoped>
.stat-row {
  margin-bottom: 20px;
}
.stat-card {
  display: flex;
  align-items: center;
}
.stat-card :deep(.el-card__body) {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 18px;
  width: 100%;
}
.stat-icon {
  width: 48px;
  height: 48px;
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #fff;
  flex-shrink: 0;
}
.stat-content {
  flex: 1;
  min-width: 0;
}
.stat-value {
  font-size: 22px;
  font-weight: 600;
  color: #303133;
}
.stat-label {
  font-size: 12px;
  color: #909399;
  margin-top: 4px;
}
.toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex-wrap: wrap;
  gap: 8px;
  margin-bottom: 16px;
}
.toolbar-left {
  display: flex;
  gap: 8px;
}
.toolbar-right {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
}
.dep-tag {
  margin-right: 4px;
  margin-bottom: 4px;
}
.text-muted {
  color: #909399;
}
.link {
  color: #409eff;
  text-decoration: none;
  word-break: break-all;
}
.link:hover {
  text-decoration: underline;
}
.pagination-wrap {
  margin-top: 16px;
  display: flex;
  justify-content: flex-end;
}
</style>
