<template>
  <div class="app-container">
    <div class="page-card">
      <div class="toolbar">
        <div class="toolbar-left">
          <el-button :icon="Refresh" @click="loadList">刷新</el-button>
        </div>
        <div class="toolbar-right">
          <el-input
            v-model="search"
            placeholder="标题/品牌/发布者"
            clearable
            style="width: 200px"
            :prefix-icon="Search"
            @keyup.enter="onSearch"
            @clear="onSearch"
          />
          <el-select
            v-model="categoryFilter"
            placeholder="分类"
            clearable
            style="width: 140px; margin-left: 8px"
            @change="onSearch"
          >
            <el-option v-for="c in flatCategories" :key="c.id" :label="c.name" :value="c.id" />
          </el-select>
          <el-select
            v-model="auditFilter"
            placeholder="审核状态"
            clearable
            style="width: 120px; margin-left: 8px"
            @change="onSearch"
          >
            <el-option label="待审核" :value="0" />
            <el-option label="已通过" :value="1" />
            <el-option label="已拒绝" :value="2" />
          </el-select>
          <el-select
            v-model="statusFilter"
            placeholder="物品状态"
            clearable
            style="width: 120px; margin-left: 8px"
            @change="onSearch"
          >
            <el-option label="草稿" :value="0" />
            <el-option label="已发布" :value="1" />
            <el-option label="已售出" :value="2" />
            <el-option label="已下架" :value="3" />
            <el-option label="已过期" :value="4" />
          </el-select>
          <el-button type="primary" :icon="Search" style="margin-left: 8px" @click="onSearch">搜索</el-button>
        </div>
      </div>

      <el-table v-loading="loading" :data="list" border stripe>
        <el-table-column prop="id" label="ID" width="70" />
        <el-table-column label="物品" min-width="240">
          <template #default="{ row }">
            <div class="item-cell">
              <el-image
                v-if="row.cover_image"
                :src="row.cover_image"
                fit="cover"
                class="item-cover"
                :preview-src-list="[row.cover_image]"
                preview-teleported
              />
              <div v-else class="item-cover item-cover-empty">无图</div>
              <div class="item-info">
                <div class="item-title">
                  {{ row.title }}
                  <el-tag v-if="row.is_urgent" type="danger" size="small" effect="dark">急</el-tag>
                </div>
                <div class="item-desc">
                  <span class="price">¥{{ Number(row.price || 0).toFixed(2) }}</span>
                  <span v-if="row.original_price > 0" class="original-price">
                    ¥{{ Number(row.original_price).toFixed(2) }}
                  </span>
                  <span class="cond">{{ conditionText(row.condition) }}</span>
                </div>
              </div>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="分类" width="110">
          <template #default="{ row }">{{ categoryName(row.category_id) }}</template>
        </el-table-column>
        <el-table-column label="发布者" width="140">
          <template #default="{ row }">
            <div class="publisher">
              <div>{{ row.user_name || `用户#${row.user_id}` }}</div>
              <div v-if="row.user_phone" class="phone">{{ row.user_phone }}</div>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="浏览" width="70" prop="view_count" />
        <el-table-column label="收藏" width="70" prop="fav_count" />
        <el-table-column label="留言" width="70" prop="message_count" />
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
        <el-table-column label="发布时间" width="160">
          <template #default="{ row }">{{ formatTime(row.published_at || row.created_at) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="280" fixed="right">
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
              <el-button type="warning" link size="small">状态<el-icon class="el-icon--right"><ArrowDown /></el-icon></el-button>
              <template #dropdown>
                <el-dropdown-menu>
                  <el-dropdown-item :command="1" :disabled="row.status === 1">发布</el-dropdown-item>
                  <el-dropdown-item :command="3" :disabled="row.status === 3">下架</el-dropdown-item>
                  <el-dropdown-item :command="4" :disabled="row.status === 4">设为过期</el-dropdown-item>
                </el-dropdown-menu>
              </template>
            </el-dropdown>
            <el-button type="info" link size="small" @click="openMessages(row)">留言</el-button>
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
    <el-dialog v-model="detailVisible" title="二手物品详情" width="760px" @close="onDetailClose">
      <el-descriptions v-if="detail" :column="2" border>
        <el-descriptions-item label="ID">{{ detail.id }}</el-descriptions-item>
        <el-descriptions-item label="标题">{{ detail.title }}</el-descriptions-item>
        <el-descriptions-item label="分类">{{ categoryName(detail.category_id) }}</el-descriptions-item>
        <el-descriptions-item label="品牌">{{ detail.brand || '-' }}</el-descriptions-item>
        <el-descriptions-item label="价格">
          <span class="price">¥{{ Number(detail.price || 0).toFixed(2) }}</span>
          <span v-if="detail.original_price > 0" class="original-price">
            ¥{{ Number(detail.original_price).toFixed(2) }}
          </span>
          <span class="price-unit">{{ detail.price_unit }}</span>
        </el-descriptions-item>
        <el-descriptions-item label="成色">{{ conditionText(detail.condition) }}</el-descriptions-item>
        <el-descriptions-item label="审核状态">
          <el-tag :type="auditTagType(detail.audit_status)" size="small">{{ auditText(detail.audit_status) }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="物品状态">
          <el-tag :type="statusTagType(detail.status)" size="small" effect="plain">{{ statusText(detail.status) }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="加急">
          <el-tag v-if="detail.is_urgent" type="danger" size="small">急转</el-tag>
          <span v-else>否</span>
        </el-descriptions-item>
        <el-descriptions-item label="交易方式">{{ deliveryText(detail.delivery_method) }}</el-descriptions-item>
        <el-descriptions-item label="联系电话">{{ detail.contact_phone || detail.user_phone || '-' }}</el-descriptions-item>
        <el-descriptions-item label="微信号">{{ detail.contact_wechat || '-' }}</el-descriptions-item>
        <el-descriptions-item label="地址" :span="2">{{ detail.address || '-' }}</el-descriptions-item>
        <el-descriptions-item v-if="detail.summary" label="摘要" :span="2">{{ detail.summary }}</el-descriptions-item>
        <el-descriptions-item v-if="detail.content" label="详细描述" :span="2">
          <div class="content-box">{{ detail.content }}</div>
        </el-descriptions-item>
        <el-descriptions-item v-if="detail.audit_reason" label="审核原因" :span="2">
          {{ detail.audit_reason }}
        </el-descriptions-item>
        <el-descriptions-item label="发布者">{{ detail.user_name || `用户#${detail.user_id}` }}</el-descriptions-item>
        <el-descriptions-item label="发布者ID">{{ detail.user_id }}</el-descriptions-item>
        <el-descriptions-item label="浏览量">{{ detail.view_count }}</el-descriptions-item>
        <el-descriptions-item label="收藏量">{{ detail.fav_count }}</el-descriptions-item>
        <el-descriptions-item label="留言数">{{ detail.message_count }}</el-descriptions-item>
        <el-descriptions-item label="过期时间">{{ formatTime(detail.expiry_time) }}</el-descriptions-item>
        <el-descriptions-item label="发布时间">{{ formatTime(detail.published_at) }}</el-descriptions-item>
        <el-descriptions-item label="创建时间">{{ formatTime(detail.created_at) }}</el-descriptions-item>
        <el-descriptions-item label="更新时间">{{ formatTime(detail.updated_at) }}</el-descriptions-item>
        <el-descriptions-item label="地区ID">{{ detail.region_id }}</el-descriptions-item>
        <el-descriptions-item label="封面图" :span="2">
          <el-image
            v-if="detail.cover_image"
            :src="detail.cover_image"
            fit="cover"
            style="width: 160px; height: 120px; border-radius: 4px"
            :preview-src-list="allImages"
            preview-teleported
          />
          <span v-else>未上传</span>
        </el-descriptions-item>
      </el-descriptions>
      <div v-if="images.length" class="detail-images">
        <div class="detail-images-title">物品图集（{{ images.length }}）</div>
        <div class="detail-images-grid">
          <el-image
            v-for="(img, idx) in images"
            :key="idx"
            :src="img"
            fit="cover"
            class="detail-image-item"
            :preview-src-list="allImages"
            :initial-index="idx"
            preview-teleported
          />
        </div>
      </div>
      <template #footer>
        <el-button @click="detailVisible = false">关闭</el-button>
      </template>
    </el-dialog>

    <!-- 留言弹窗 -->
    <el-dialog v-model="messagesVisible" title="留言列表" width="680px">
      <div v-loading="messagesLoading">
        <el-empty v-if="!messages.length" description="暂无留言" />
        <div v-else class="messages-list">
          <div v-for="msg in messages" :key="msg.id" class="message-item">
            <div class="message-header">
              <span class="message-from">{{ msg.from_name || `用户#${msg.from_user_id}` }}</span>
              <span class="message-time">{{ formatTime(msg.created_at) }}</span>
            </div>
            <div class="message-content">{{ msg.content }}</div>
          </div>
        </div>
      </div>
      <template #footer>
        <el-button @click="messagesVisible = false">关闭</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Refresh, Search, ArrowDown } from '@element-plus/icons-vue'
import {
  adminListErshous,
  adminGetErshou,
  auditErshou,
  adminUpdateErshouStatus,
  listErshouMessages
} from '@/api/ershou'
import { getCategoryTree } from '@/api/category'
import { formatTime } from '@/utils/format'

// ===== 列表 =====
const loading = ref(false)
const search = ref('')
const categoryFilter = ref(null)
const auditFilter = ref(null)
const statusFilter = ref(null)
const page = ref(1)
const pageSize = ref(10)
const total = ref(0)
const list = ref([])

// ===== 分类 =====
const categoryTree = ref([])
const flatCategories = computed(() => {
  const result = []
  const walk = (nodes) => {
    if (!Array.isArray(nodes)) return
    nodes.forEach((n) => {
      result.push({ id: n.id, name: n.name })
      if (n.children?.length) walk(n.children)
    })
  }
  walk(categoryTree.value)
  return result
})
const categoryName = (id) => flatCategories.value.find((c) => c.id === id)?.name || '-'

const loadCategories = async () => {
  try {
    const res = await getCategoryTree()
    categoryTree.value = res.data || []
  } catch (e) {
    categoryTree.value = []
  }
}

// ===== 状态格式化 =====
const auditText = (s) => ({ 0: '待审核', 1: '已通过', 2: '已拒绝' }[s] || '-')
const auditTagType = (s) => ({ 0: 'warning', 1: 'success', 2: 'danger' }[s] || 'info')
const statusText = (s) => ({ 0: '草稿', 1: '已发布', 2: '已售出', 3: '已下架', 4: '已过期' }[s] || '-')
const statusTagType = (s) => ({ 0: 'info', 1: 'success', 2: 'primary', 3: 'warning', 4: 'danger' }[s] || 'info')
const conditionText = (c) => ({ new: '全新', almost_new: '几乎全新', used: '有使用痕迹', broken: '有瑕疵' }[c] || '-')
const deliveryText = (d) => ({ face: '当面交易', self: '自提', express: '快递' }[d] || '-')

// ===== 查询 =====
const onSearch = () => {
  page.value = 1
  loadList()
}

const loadList = async () => {
  loading.value = true
  try {
    const res = await adminListErshous({
      page: page.value,
      page_size: pageSize.value,
      keyword: search.value.trim() || undefined,
      category_id: categoryFilter.value || undefined,
      audit_status: auditFilter.value === null || auditFilter.value === '' ? undefined : auditFilter.value,
      status: statusFilter.value === null || statusFilter.value === '' ? undefined : statusFilter.value
    })
    const data = res.data || {}
    list.value = data.list || []
    total.value = data.total || 0
  } finally {
    loading.value = false
  }
}

// ===== 详情弹窗 =====
const detailVisible = ref(false)
const detail = ref(null)
const images = ref([])
const allImages = computed(() => {
  const arr = []
  if (detail.value?.cover_image) arr.push(detail.value.cover_image)
  arr.push(...images.value)
  return arr
})

const openDetail = async (row) => {
  detail.value = row
  images.value = []
  detailVisible.value = true
  try {
    const res = await adminGetErshou(row.id)
    if (res.data) {
      detail.value = res.data
      images.value = res.data.images || []
    }
  } catch (e) {
    // ignore
  }
}

const onDetailClose = () => {
  detail.value = null
  images.value = []
}

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
      await auditErshou(row.id, { audit_status: auditStatus, audit_reason: value || '' })
    } else {
      await ElMessageBox.confirm(`确定${action}物品 "${row.title}" 的审核吗？`, '提示', { type: 'warning' })
      await auditErshou(row.id, { audit_status: auditStatus })
    }
    ElMessage.success(`已${action}`)
    await loadList()
    // 如果详情弹窗打开，同步刷新
    if (detailVisible.value && detail.value?.id === row.id) {
      const res = await adminGetErshou(row.id)
      if (res.data) detail.value = res.data
    }
  } catch (e) {
    // 取消
  }
}

// ===== 强制状态变更 =====
const handleStatusCommand = async (row, status) => {
  try {
    const label = statusText(status)
    await ElMessageBox.confirm(`确定将物品 "${row.title}" 设为「${label}」吗？`, '提示', { type: 'warning' })
    await adminUpdateErshouStatus(row.id, status)
    ElMessage.success('状态更新成功')
    await loadList()
    if (detailVisible.value && detail.value?.id === row.id) {
      const res = await adminGetErshou(row.id)
      if (res.data) detail.value = res.data
    }
  } catch (e) {
    // 取消
  }
}

// ===== 留言弹窗 =====
const messagesVisible = ref(false)
const messagesLoading = ref(false)
const messages = ref([])

const openMessages = async (row) => {
  messages.value = []
  messagesVisible.value = true
  messagesLoading.value = true
  try {
    const res = await listErshouMessages(row.id, { page: 1, page_size: 50 })
    messages.value = res.data?.list || []
  } catch (e) {
    messages.value = []
  } finally {
    messagesLoading.value = false
  }
}

onMounted(async () => {
  await loadCategories()
  await loadList()
})
</script>

<style scoped>
.toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex-wrap: wrap;
  gap: 8px;
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
.item-cell {
  display: flex;
  align-items: center;
  gap: 10px;
}
.item-cover {
  width: 56px;
  height: 56px;
  border-radius: 6px;
  flex-shrink: 0;
  border: 1px solid #ebeef5;
}
.item-cover-empty {
  display: flex;
  align-items: center;
  justify-content: center;
  background: #fafafa;
  color: #909399;
  font-size: 12px;
}
.item-info {
  flex: 1;
  min-width: 0;
}
.item-title {
  font-weight: 500;
  color: #303133;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  display: flex;
  align-items: center;
  gap: 6px;
}
.item-desc {
  font-size: 12px;
  color: #909399;
  margin-top: 2px;
  display: flex;
  align-items: center;
  gap: 6px;
}
.price {
  color: #f56c6c;
  font-weight: 600;
}
.original-price {
  color: #c0c4cc;
  text-decoration: line-through;
  font-size: 12px;
}
.cond {
  color: #909399;
  font-size: 12px;
}
.price-unit {
  color: #909399;
  font-size: 12px;
  margin-left: 2px;
}
.publisher {
  font-size: 13px;
}
.publisher .phone {
  font-size: 12px;
  color: #909399;
  margin-top: 2px;
}
.detail-images {
  margin-top: 16px;
  border-top: 1px solid #ebeef5;
  padding-top: 12px;
}
.detail-images-title {
  font-size: 14px;
  font-weight: 500;
  margin-bottom: 8px;
  color: #303133;
}
.detail-images-grid {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}
.detail-image-item {
  width: 110px;
  height: 110px;
  border-radius: 4px;
  border: 1px solid #ebeef5;
}
.content-box {
  white-space: pre-wrap;
  word-break: break-all;
  max-height: 200px;
  overflow-y: auto;
}
.messages-list {
  max-height: 480px;
  overflow-y: auto;
}
.message-item {
  padding: 10px 0;
  border-bottom: 1px solid #f0f0f0;
}
.message-item:last-child {
  border-bottom: none;
}
.message-header {
  display: flex;
  justify-content: space-between;
  margin-bottom: 4px;
}
.message-from {
  font-weight: 500;
  color: #303133;
}
.message-time {
  font-size: 12px;
  color: #909399;
}
.message-content {
  color: #606266;
  word-break: break-all;
}
</style>
