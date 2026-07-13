<template>
  <div class="app-container">
    <div class="page-card">
      <div class="toolbar">
        <div class="toolbar-left">
          <el-button :icon="Refresh" @click="loadFavorites(true)">刷新</el-button>
        </div>
        <div class="toolbar-right">
          <span class="muted">共 {{ total }} 条收藏</span>
        </div>
      </div>

      <el-empty v-if="!loading && list.length === 0" description="还没有收藏，去头条列表找找感兴趣的内容吧" />

      <el-table v-else v-loading="loading" :data="list" border stripe>
        <el-table-column prop="id" label="ID" width="70" />
        <el-table-column label="标题" min-width="240" show-overflow-tooltip>
          <template #default="{ row }">
            <el-link type="primary" @click="goDetail(row)">{{ row.title || '（无标题）' }}</el-link>
          </template>
        </el-table-column>
        <el-table-column prop="author_name" label="作者" width="120" />
        <el-table-column label="类型" width="100">
          <template #default="{ row }">
            <el-tag size="small" :type="listingTypeTag(row.listing_type)">
              {{ listingTypeText(row.listing_type) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="价格" width="120">
          <template #default="{ row }">
            <span v-if="row.price > 0">{{ formatPrice(row.price, row.price_unit) }}</span>
            <span v-else class="muted">面议</span>
          </template>
        </el-table-column>
        <el-table-column prop="view_count" label="浏览" width="80" />
        <el-table-column prop="fav_count" label="收藏数" width="90" />
        <el-table-column label="状态" width="90">
          <template #default="{ row }">
            <el-tag :type="statusTag(row.status)" size="small">{{ statusText(row.status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="发布时间" width="170">
          <template #default="{ row }">{{ formatTime(row.published_at) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="160" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="goDetail(row)">查看</el-button>
            <el-button type="danger" link size="small" @click="handleUnfav(row)">取消收藏</el-button>
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
          @current-change="loadFavorites(false)"
          @size-change="loadFavorites(true)"
        />
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Refresh } from '@element-plus/icons-vue'
import { listFavorites, favNews } from '@/api/news'
import { formatTime } from '@/utils/format'

const router = useRouter()

const loading = ref(false)
const list = ref([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(10)

const listingTypeText = (t) => {
  switch (t) {
    case 'sell': return '出售'
    case 'buy': return '求购'
    case 'rent': return '出租'
    case 'service': return '服务'
    case 'job': return '招聘'
    default: return '其他'
  }
}

const listingTypeTag = (t) => {
  switch (t) {
    case 'sell': return 'success'
    case 'buy': return 'warning'
    case 'rent': return 'primary'
    case 'service': return 'info'
    case 'job': return 'danger'
    default: return ''
  }
}

const statusText = (s) => {
  switch (s) {
    case 0: return '草稿'
    case 1: return '已发布'
    case 2: return '已下架'
    case 3: return '已过期'
    default: return '未知'
  }
}

const statusTag = (s) => {
  switch (s) {
    case 1: return 'success'
    case 0: return 'info'
    case 2: return 'warning'
    case 3: return 'danger'
    default: return ''
  }
}

const formatPrice = (price, unit) => {
  if (!unit) unit = '元'
  return `${price} ${unit}`
}

const loadFavorites = async (reset = false) => {
  if (reset) {
    page.value = 1
  }
  loading.value = true
  try {
    const res = await listFavorites({
      page: page.value,
      page_size: pageSize.value
    })
    const data = res.data || {}
    list.value = data.list || []
    total.value = data.total || 0
  } catch (e) {
    // 错误已由 request 拦截器提示
  } finally {
    loading.value = false
  }
}

const goDetail = (row) => {
  router.push({ name: 'NewsDetail', params: { id: row.id } })
}

const handleUnfav = async (row) => {
  try {
    await ElMessageBox.confirm(
      `确定取消收藏「${row.title || '该头条'}」吗？`,
      '取消收藏',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )
  } catch (e) {
    return // 用户取消
  }

  try {
    const res = await favNews(row.id)
    // toggle 接口返回 faved=false 表示已取消收藏
    if (res.data && res.data.faved === false) {
      ElMessage.success('已取消收藏')
      // 若当前页只剩这一条且不是第一页，回退一页避免空列表
      if (list.value.length === 1 && page.value > 1) {
        page.value -= 1
      }
      await loadFavorites(false)
    } else {
      // 状态异常（理论上收藏列表里 toggle 一定返回 faved=false）
      ElMessage.warning('操作未生效，请刷新后重试')
      await loadFavorites(false)
    }
  } catch (e) {
    // 错误已由 request 拦截器提示
  }
}

onMounted(() => {
  loadFavorites(true)
})
</script>

<style scoped>
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
  flex-wrap: wrap;
}
.toolbar-right {
  display: flex;
  align-items: center;
}
.muted {
  color: #909399;
  font-size: 13px;
}
.pagination-wrap {
  margin-top: 16px;
  display: flex;
  justify-content: flex-end;
}
</style>
