<template>
  <div class="app-container">
    <div class="page-card">
      <div class="toolbar">
        <div class="toolbar-left">
          <el-button :icon="Refresh" @click="loadShops">刷新</el-button>
        </div>
        <div class="toolbar-right">
          <el-input
            v-model="search"
            placeholder="店铺名称关键词"
            clearable
            style="width: 180px"
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
            placeholder="营业状态"
            clearable
            style="width: 120px; margin-left: 8px"
            @change="onSearch"
          >
            <el-option label="歇业" :value="0" />
            <el-option label="营业中" :value="1" />
            <el-option label="休息中" :value="2" />
          </el-select>
          <el-select
            v-model="recommendFilter"
            placeholder="推荐"
            clearable
            style="width: 100px; margin-left: 8px"
            @change="onSearch"
          >
            <el-option label="推荐" :value="1" />
            <el-option label="未推荐" :value="0" />
          </el-select>
          <el-button type="primary" :icon="Search" style="margin-left: 8px" @click="onSearch">搜索</el-button>
        </div>
      </div>

      <el-table v-loading="loading" :data="list" border stripe>
        <el-table-column prop="id" label="ID" width="70" />
        <el-table-column label="店铺" min-width="220">
          <template #default="{ row }">
            <div class="shop-cell">
              <el-image
                v-if="row.logo"
                :src="row.logo"
                fit="cover"
                class="shop-logo"
                :preview-src-list="[row.logo]"
                preview-teleported
              />
              <div v-else class="shop-logo shop-logo-empty">无图</div>
              <div class="shop-info">
                <div class="shop-name">{{ row.name }}</div>
                <div class="shop-desc">{{ row.description || '-' }}</div>
              </div>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="分类" width="110">
          <template #default="{ row }">{{ categoryName(row.category_id) }}</template>
        </el-table-column>
        <el-table-column prop="phone" label="电话" width="130" />
        <el-table-column prop="address" label="地址" min-width="160" show-overflow-tooltip />
        <el-table-column label="评分" width="80">
          <template #default="{ row }">
            <span :class="['rating', row.rating >= 4 ? 'rating-good' : row.rating >= 3 ? 'rating-normal' : 'rating-low']">
              {{ Number(row.rating || 0).toFixed(1) }}
            </span>
          </template>
        </el-table-column>
        <el-table-column prop="views" label="浏览" width="70" />
        <el-table-column label="审核" width="90">
          <template #default="{ row }">
            <el-tag :type="auditTagType(row.audit_status)" size="small">{{ auditText(row.audit_status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="营业" width="90">
          <template #default="{ row }">
            <el-tag :type="statusTagType(row.status)" size="small" effect="plain">{{ statusText(row.status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="推荐" width="70">
          <template #default="{ row }">
            <el-tag v-if="row.is_recommend === 1" type="warning" size="small">推荐</el-tag>
            <span v-else class="text-muted">-</span>
          </template>
        </el-table-column>
        <el-table-column label="创建时间" width="160">
          <template #default="{ row }">{{ formatTime(row.created_at) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="280" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="openDetail(row)">详情</el-button>
            <el-button
              v-if="row.audit_status === 0"
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
              <el-button type="warning" link size="small">营业<el-icon class="el-icon--right"><ArrowDown /></el-icon></el-button>
              <template #dropdown>
                <el-dropdown-menu>
                  <el-dropdown-item :command="1" :disabled="row.status === 1">营业中</el-dropdown-item>
                  <el-dropdown-item :command="2" :disabled="row.status === 2">休息中</el-dropdown-item>
                  <el-dropdown-item :command="0" :disabled="row.status === 0">歇业</el-dropdown-item>
                </el-dropdown-menu>
              </template>
            </el-dropdown>
            <el-button
              :type="row.is_recommend === 1 ? 'info' : 'warning'"
              link
              size="small"
              @click="handleToggleRecommend(row)"
            >{{ row.is_recommend === 1 ? '取消推荐' : '推荐' }}</el-button>
            <el-button type="danger" link size="small" @click="handleDelete(row)">删除</el-button>
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
          @current-change="loadShops"
          @size-change="loadShops"
        />
      </div>
    </div>

    <!-- 详情弹窗 -->
    <el-dialog v-model="detailVisible" title="商家详情" width="680px">
      <el-descriptions v-if="detail" :column="2" border>
        <el-descriptions-item label="ID">{{ detail.id }}</el-descriptions-item>
        <el-descriptions-item label="店铺名称">{{ detail.name }}</el-descriptions-item>
        <el-descriptions-item label="分类">{{ categoryName(detail.category_id) }}</el-descriptions-item>
        <el-descriptions-item label="审核状态">
          <el-tag :type="auditTagType(detail.audit_status)" size="small">{{ auditText(detail.audit_status) }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="营业状态">
          <el-tag :type="statusTagType(detail.status)" size="small" effect="plain">{{ statusText(detail.status) }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="是否推荐">
          <el-tag v-if="detail.is_recommend === 1" type="warning" size="small">推荐</el-tag>
          <span v-else>否</span>
        </el-descriptions-item>
        <el-descriptions-item label="联系电话">{{ detail.phone || '-' }}</el-descriptions-item>
        <el-descriptions-item label="营业时间">{{ detail.business_hours || '-' }}</el-descriptions-item>
        <el-descriptions-item label="评分">{{ Number(detail.rating || 0).toFixed(1) }}</el-descriptions-item>
        <el-descriptions-item label="浏览量">{{ detail.views }}</el-descriptions-item>
        <el-descriptions-item label="地址" :span="2">{{ detail.address || '-' }}</el-descriptions-item>
        <el-descriptions-item label="简介" :span="2">{{ detail.description || '-' }}</el-descriptions-item>
        <el-descriptions-item label="Logo" :span="2">
          <el-image
            v-if="detail.logo"
            :src="detail.logo"
            fit="cover"
            style="width: 120px; height: 120px; border-radius: 4px"
            :preview-src-list="[detail.logo]"
            preview-teleported
          />
          <span v-else>未上传</span>
        </el-descriptions-item>
        <el-descriptions-item label="所有者用户ID">{{ detail.user_id }}</el-descriptions-item>
        <el-descriptions-item label="地区ID">{{ detail.region_id }}</el-descriptions-item>
        <el-descriptions-item label="创建时间">{{ formatTime(detail.created_at) }}</el-descriptions-item>
        <el-descriptions-item label="更新时间">{{ formatTime(detail.updated_at) }}</el-descriptions-item>
      </el-descriptions>
      <!-- 相册 -->
      <div v-if="images.length" class="detail-images">
        <div class="detail-images-title">店铺相册（{{ images.length }}）</div>
        <div class="detail-images-grid">
          <el-image
            v-for="img in images"
            :key="img.id"
            :src="img.image_url"
            fit="cover"
            class="detail-image-item"
            :preview-src-list="images.map((i) => i.image_url)"
            preview-teleported
          />
        </div>
      </div>
      <template #footer>
        <el-button @click="detailVisible = false">关闭</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Refresh, Search, ArrowDown } from '@element-plus/icons-vue'
import {
  adminListShops,
  auditShop,
  updateShopStatus,
  setShopRecommend,
  deleteShop,
  getShop,
  getShopImages
} from '@/api/shop'
import { getCategoryTree } from '@/api/category'
import { formatTime } from '@/utils/format'

// ===== 列表 =====
const loading = ref(false)
const search = ref('')
const categoryFilter = ref(null)
const auditFilter = ref(null)
const statusFilter = ref(null)
const recommendFilter = ref(null)
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
const statusText = (s) => ({ 0: '歇业', 1: '营业中', 2: '休息中' }[s] || '-')
const statusTagType = (s) => ({ 0: 'info', 1: 'success', 2: 'warning' }[s] || 'info')

// ===== 查询 =====
const onSearch = () => {
  page.value = 1
  loadShops()
}

const loadShops = async () => {
  loading.value = true
  try {
    const res = await adminListShops({
      page: page.value,
      page_size: pageSize.value,
      keyword: search.value.trim() || undefined,
      category_id: categoryFilter.value || undefined,
      audit_status: auditFilter.value === null || auditFilter.value === '' ? -1 : auditFilter.value,
      status: statusFilter.value === null || statusFilter.value === '' ? -1 : statusFilter.value,
      is_recommend: recommendFilter.value === null || recommendFilter.value === '' ? -1 : recommendFilter.value
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

const openDetail = async (row) => {
  detail.value = row
  images.value = []
  detailVisible.value = true
  try {
    const [infoRes, imgRes] = await Promise.all([getShop(row.id), getShopImages(row.id)])
    if (infoRes.data) detail.value = infoRes.data
    images.value = imgRes.data || []
  } catch (e) {
    // ignore
  }
}

// ===== 审核 =====
const handleAudit = async (row, auditStatus) => {
  const action = auditStatus === 1 ? '通过' : '拒绝'
  try {
    if (auditStatus === 2) {
      const { value } = await ElMessageBox.prompt(`请输入拒绝原因（可选）`, '拒绝商家审核', {
        confirmButtonText: '确定拒绝',
        cancelButtonText: '取消',
        inputType: 'textarea',
        inputPlaceholder: '拒绝原因（可不填）'
      })
      await auditShop(row.id, { audit_status: auditStatus, reason: value || '' })
    } else {
      await ElMessageBox.confirm(`确定${action}商家 "${row.name}" 的审核吗？`, '提示', { type: 'warning' })
      await auditShop(row.id, { audit_status: auditStatus })
    }
    ElMessage.success(`已${action}`)
    await loadShops()
  } catch (e) {
    // 取消
  }
}

// ===== 营业状态 =====
const handleStatusCommand = async (row, status) => {
  try {
    await updateShopStatus(row.id, status)
    ElMessage.success('状态更新成功')
    await loadShops()
  } catch (e) {
    // ignore
  }
}

// ===== 推荐 =====
const handleToggleRecommend = async (row) => {
  const next = row.is_recommend === 1 ? 0 : 1
  try {
    await setShopRecommend(row.id, next)
    ElMessage.success(next === 1 ? '已设为推荐' : '已取消推荐')
    await loadShops()
  } catch (e) {
    // ignore
  }
}

// ===== 删除 =====
const handleDelete = async (row) => {
  try {
    await ElMessageBox.confirm(`确定删除商家 "${row.name}" 吗？删除后不可恢复！`, '危险操作', {
      type: 'error',
      confirmButtonText: '确定删除',
      cancelButtonText: '取消'
    })
    await deleteShop(row.id)
    ElMessage.success('删除成功')
    await loadShops()
  } catch (e) {
    // 取消
  }
}

onMounted(async () => {
  await loadCategories()
  await loadShops()
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
.shop-cell {
  display: flex;
  align-items: center;
  gap: 10px;
}
.shop-logo {
  width: 48px;
  height: 48px;
  border-radius: 6px;
  flex-shrink: 0;
  border: 1px solid #ebeef5;
}
.shop-logo-empty {
  display: flex;
  align-items: center;
  justify-content: center;
  background: #fafafa;
  color: #909399;
  font-size: 12px;
}
.shop-info {
  flex: 1;
  min-width: 0;
}
.shop-name {
  font-weight: 500;
  color: #303133;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.shop-desc {
  font-size: 12px;
  color: #909399;
  margin-top: 2px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.rating {
  font-weight: 600;
}
.rating-good {
  color: #67c23a;
}
.rating-normal {
  color: #e6a23c;
}
.rating-low {
  color: #f56c6c;
}
.text-muted {
  color: #c0c4cc;
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
  width: 100px;
  height: 100px;
  border-radius: 4px;
  border: 1px solid #ebeef5;
}
</style>
