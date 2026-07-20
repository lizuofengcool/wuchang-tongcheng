<template>
  <div class="app-container">
    <div class="page-card">
      <el-form :inline="true" :model="filters" class="filter-form" @submit.prevent>
        <el-form-item label="店铺名">
          <el-input v-model="filters.keyword" placeholder="店铺名" clearable style="width: 200px" @keyup.enter="onSearch" />
        </el-form-item>
        <el-form-item label="等级">
          <el-select v-model="filters.level" placeholder="全部" clearable style="width: 140px" @change="onSearch">
            <el-option v-for="(label, val) in levelMap" :key="val" :label="label" :value="Number(val)" />
          </el-select>
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="filters.status" placeholder="全部" clearable style="width: 140px" @change="onSearch">
            <el-option v-for="(label, val) in statusMap" :key="val" :label="label" :value="Number(val)" />
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
        <el-table-column prop="id" label="ID" width="70" />
        <el-table-column label="Logo" width="70">
          <template #default="{ row }">
            <el-image v-if="row.logo" :src="row.logo" fit="cover" class="logo-thumb" />
            <div v-else class="logo-thumb logo-empty">无</div>
          </template>
        </el-table-column>
        <el-table-column label="店铺名" min-width="180">
          <template #default="{ row }">
            <el-link type="primary" :underline="'never'" @click="openDetail(row)">{{ row.shop_name }}</el-link>
          </template>
        </el-table-column>
        <el-table-column label="Banner" width="100">
          <template #default="{ row }">
            <el-image v-if="row.banner" :src="row.banner" fit="cover" class="banner-thumb" :preview-src-list="[row.banner]" preview-teleported />
            <span v-else class="text-muted">-</span>
          </template>
        </el-table-column>
        <el-table-column label="等级" width="100">
          <template #default="{ row }">
            <el-tag :type="levelTagType(row.level)" size="small">{{ row.level_text || levelMap[row.level] }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="follower_count" label="粉丝数" width="100" />
        <el-table-column prop="item_count" label="商品数" width="100" />
        <el-table-column prop="sold_count" label="已售" width="80" />
        <el-table-column label="好评率" width="100">
          <template #default="{ row }">{{ (row.good_rate * 100).toFixed(1) }}%</template>
        </el-table-column>
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="statusTagType(row.status)" size="small">{{ row.status_text || statusMap[row.status] }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="入驻时间" width="160">
          <template #default="{ row }">{{ formatTime(row.created_at) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="openDetail(row)">详情</el-button>
            <el-button v-if="row.status === 0" type="success" link size="small" @click="onAudit(row, 1)">通过</el-button>
            <el-button v-if="row.status === 0" type="danger" link size="small" @click="onAudit(row, 2)">拒绝</el-button>
            <el-dropdown trigger="click" @command="(cmd) => onLevelCommand(row, cmd)">
              <el-button type="warning" link size="small">等级<el-icon class="el-icon--right"><ArrowDown /></el-icon></el-button>
              <template #dropdown>
                <el-dropdown-menu>
                  <el-dropdown-item v-for="(label, val) in levelMap" :key="val" :command="Number(val)">设为{{ label }}</el-dropdown-item>
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
    <el-dialog v-model="detailVisible" title="店铺详情" width="800px">
      <el-tabs v-if="detail">
        <el-tab-pane label="基础信息">
          <el-descriptions :column="2" border>
            <el-descriptions-item label="店铺名">{{ detail.shop_name }}</el-descriptions-item>
            <el-descriptions-item label="店铺ID">{{ detail.id }}</el-descriptions-item>
            <el-descriptions-item label="用户ID">{{ detail.user_id }}</el-descriptions-item>
            <el-descriptions-item label="等级">
              <el-tag :type="levelTagType(detail.level)" size="small">{{ detail.level_text || levelMap[detail.level] }}</el-tag>
            </el-descriptions-item>
            <el-descriptions-item label="状态">
              <el-tag :type="statusTagType(detail.status)" size="small">{{ detail.status_text || statusMap[detail.status] }}</el-tag>
            </el-descriptions-item>
            <el-descriptions-item label="联系人">{{ detail.contact_name || '-' }}</el-descriptions-item>
            <el-descriptions-item label="联系电话">{{ maskPhone(detail.contact_phone) }}</el-descriptions-item>
            <el-descriptions-item label="微信号">{{ detail.contact_wechat || '-' }}</el-descriptions-item>
            <el-descriptions-item label="地址" :span="2">{{ detail.address || '-' }}</el-descriptions-item>
            <el-descriptions-item label="描述" :span="2">{{ detail.description || '-' }}</el-descriptions-item>
            <el-descriptions-item label="营业执照">{{ detail.business_license || '-' }}</el-descriptions-item>
            <el-descriptions-item label="执照编号">{{ detail.license_no || '-' }}</el-descriptions-item>
            <el-descriptions-item label="保证金">¥{{ Number(detail.deposit || 0).toFixed(2) }}</el-descriptions-item>
            <el-descriptions-item label="认证时间">{{ formatTime(detail.verified_at) }}</el-descriptions-item>
            <el-descriptions-item v-if="detail.rejected_reason" label="拒绝原因" :span="2">{{ detail.rejected_reason }}</el-descriptions-item>
          </el-descriptions>
        </el-tab-pane>
        <el-tab-pane label="数据统计">
          <el-descriptions :column="2" border>
            <el-descriptions-item label="粉丝数">{{ detail.follower_count }}</el-descriptions-item>
            <el-descriptions-item label="商品数">{{ detail.item_count }}</el-descriptions-item>
            <el-descriptions-item label="已售">{{ detail.sold_count }}</el-descriptions-item>
            <el-descriptions-item label="总成交额">¥{{ Number(detail.total_amount || 0).toFixed(2) }}</el-descriptions-item>
            <el-descriptions-item label="好评率">{{ (detail.good_rate * 100).toFixed(1) }}%</el-descriptions-item>
            <el-descriptions-item label="审批时间">{{ formatTime(detail.approved_at) }}</el-descriptions-item>
          </el-descriptions>
        </el-tab-pane>
        <el-tab-pane label="Logo/Banner">
          <div v-if="detail.logo" class="image-block">
            <div class="image-label">Logo</div>
            <el-image :src="detail.logo" fit="cover" class="logo-large" :preview-src-list="[detail.logo]" preview-teleported />
          </div>
          <div v-if="detail.banner" class="image-block">
            <div class="image-label">Banner</div>
            <el-image :src="detail.banner" fit="cover" class="banner-large" :preview-src-list="[detail.banner]" preview-teleported />
          </div>
        </el-tab-pane>
      </el-tabs>
      <template #footer><el-button @click="detailVisible = false">关闭</el-button></template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Refresh, RefreshLeft, Search, ArrowDown } from '@element-plus/icons-vue'
import { listErshouShops, getErshouShop, auditErshouShop, updateErshouShopStatus } from '@/api/ershou'
import { formatTime } from '@/utils/format'

const loading = ref(false)
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)
const list = ref([])

const filters = reactive({ keyword: '', level: null, status: null })

const levelMap = { 0: '普通', 1: '认证', 2: '金牌', 3: '钻石' }
const statusMap = { 0: '待审核', 1: '正常', 2: '已禁用', 3: '已拒绝' }

const levelTagType = (l) => ({ 0: 'info', 1: 'success', 2: 'warning', 3: 'danger' }[l] || 'info')
const statusTagType = (s) => ({ 0: 'warning', 1: 'success', 2: 'danger', 3: 'info' }[s] || 'info')

const maskPhone = (p) => {
  if (!p) return '-'
  const s = String(p)
  if (s.length < 7) return s
  return s.slice(0, 3) + '****' + s.slice(-4)
}

const onSearch = () => { page.value = 1; loadList() }
const onReset = () => { filters.keyword = ''; filters.level = null; filters.status = null; page.value = 1; loadList() }

const loadList = async () => {
  loading.value = true
  try {
    const res = await listErshouShops({
      page: page.value,
      page_size: pageSize.value,
      keyword: filters.keyword || undefined,
      level: filters.level === null || filters.level === '' ? undefined : filters.level,
      status: filters.status === null || filters.status === '' ? undefined : filters.status
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
  detail.value = row
  detailVisible.value = true
  try {
    const res = await getErshouShop(row.id)
    if (res.data) detail.value = res.data
  } catch (e) { /* ignore */ }
}

const onAudit = async (row, status) => {
  try {
    const action = status === 1 ? '通过' : '拒绝'
    let reason = ''
    if (status === 2) {
      const { value } = await ElMessageBox.prompt('请输入拒绝原因', '拒绝店铺', { inputType: 'textarea' })
      reason = value || ''
    } else {
      await ElMessageBox.confirm(`确认${action}店铺 "${row.shop_name}" 的审核？`, '提示', { type: 'warning' })
    }
    await auditErshouShop(row.id, { status, rejected_reason: reason })
    ElMessage.success(`已${action}`)
    await loadList()
  } catch (e) { /* cancel */ }
}

const onLevelCommand = async (row, level) => {
  try {
    const label = levelMap[level]
    await ElMessageBox.confirm(`确认将店铺 "${row.shop_name}" 设为「${label}」等级？`, '提示', { type: 'warning' })
    // 等级调整通过 status 接口或专用接口（此处复用 update 接口）
    const { updateErshouShop } = await import('@/api/ershou')
    await updateErshouShop(row.id, { level })
    ElMessage.success('等级更新成功')
    await loadList()
  } catch (e) { /* cancel */ }
}

onMounted(() => loadList())
</script>

<style scoped>
.filter-form { margin-bottom: 12px; padding: 12px 16px; background: #fafafa; border-radius: 4px; }
.filter-form :deep(.el-form-item) { margin-bottom: 8px; margin-right: 12px; }
.toolbar { display: flex; justify-content: flex-start; margin-bottom: 12px; }
.logo-thumb { width: 50px; height: 50px; border-radius: 4px; border: 1px solid #ebeef5; }
.logo-empty { display: flex; align-items: center; justify-content: center; background: #fafafa; color: #909399; font-size: 12px; }
.banner-thumb { width: 80px; height: 40px; border-radius: 4px; border: 1px solid #ebeef5; }
.text-muted { color: #909399; }
.image-block { margin-bottom: 16px; }
.image-label { font-weight: 500; margin-bottom: 8px; color: #303133; }
.logo-large { width: 120px; height: 120px; border-radius: 4px; border: 1px solid #ebeef5; }
.banner-large { width: 100%; max-height: 200px; border-radius: 4px; border: 1px solid #ebeef5; }
.pagination-wrap { margin-top: 16px; display: flex; justify-content: flex-end; }
</style>
