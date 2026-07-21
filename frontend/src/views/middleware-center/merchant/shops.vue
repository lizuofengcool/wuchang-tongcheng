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
            placeholder="商户名"
            clearable
            style="width: 200px"
            :prefix-icon="Search"
            @keyup.enter="onSearch"
            @clear="onSearch"
          />
          <el-select
            v-model="statusFilter"
            placeholder="状态"
            clearable
            style="width: 140px; margin-left: 8px"
            @change="onSearch"
          >
            <el-option label="审核中" :value="0" />
            <el-option label="正常" :value="1" />
            <el-option label="停用" :value="2" />
          </el-select>
          <el-input-number
            v-model="ownerFilter"
            :min="0"
            placeholder="店主ID"
            style="width: 130px; margin-left: 8px"
            @change="onSearch"
          />
          <el-button type="primary" :icon="Search" style="margin-left: 8px" @click="onSearch">搜索</el-button>
        </div>
      </div>

      <el-table v-loading="loading" :data="list" border stripe>
        <el-table-column prop="id" label="ID" width="70" />
        <el-table-column prop="region_id" label="地区ID" width="90" />
        <el-table-column label="商户名" min-width="180">
          <template #default="{ row }">
            <div class="shop-name">
              <el-image v-if="row.logo" :src="row.logo" fit="cover" class="shop-logo" />
              <el-icon v-else class="shop-logo-placeholder"><Shop /></el-icon>
              <span>{{ row.name || '-' }}</span>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="owner_id" label="店主ID" width="100" />
        <el-table-column label="主营类目" width="110">
          <template #default="{ row }">{{ row.category_id ? `#${row.category_id}` : '-' }}</template>
        </el-table-column>
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="statusTagType(row.status)" size="small">{{ row.status_text || statusText(row.status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="信用分" width="100">
          <template #default="{ row }">
            <span :class="row.credit_score < 60 ? 'credit-low' : ''">{{ row.credit_score }}</span>
          </template>
        </el-table-column>
        <el-table-column label="等级" width="90">
          <template #default="{ row }">
            <el-tag type="warning" size="small">Lv{{ row.level }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="settled_at" label="入驻时间" width="170">
          <template #default="{ row }">{{ row.settled_at || '-' }}</template>
        </el-table-column>
        <el-table-column prop="created_at" label="创建时间" width="170" />
        <el-table-column label="操作" width="280" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" size="small" @click="openDetail(row)">详情</el-button>
            <el-button link type="warning" size="small" @click="openStatus(row)">状态</el-button>
            <el-button link type="primary" size="small" @click="openCredit(row)">信用</el-button>
            <el-button link type="success" size="small" @click="openLevel(row)">等级</el-button>
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
          @size-change="loadList"
          @current-change="loadList"
        />
      </div>
    </div>

    <!-- 详情弹窗 -->
    <el-dialog v-model="detailVisible" title="店铺详情" width="640px">
      <el-descriptions :column="2" border v-if="detail">
        <el-descriptions-item label="ID">{{ detail.id }}</el-descriptions-item>
        <el-descriptions-item label="地区ID">{{ detail.region_id }}</el-descriptions-item>
        <el-descriptions-item label="商户名">{{ detail.name || '-' }}</el-descriptions-item>
        <el-descriptions-item label="店主ID">{{ detail.owner_id }}</el-descriptions-item>
        <el-descriptions-item label="主营类目">{{ detail.category_id ? `#${detail.category_id}` : '-' }}</el-descriptions-item>
        <el-descriptions-item label="状态">{{ detail.status_text || statusText(detail.status) }}</el-descriptions-item>
        <el-descriptions-item label="信用分">{{ detail.credit_score }}</el-descriptions-item>
        <el-descriptions-item label="等级">Lv{{ detail.level }}</el-descriptions-item>
        <el-descriptions-item label="入驻时间">{{ detail.settled_at || '-' }}</el-descriptions-item>
        <el-descriptions-item label="创建时间">{{ detail.created_at }}</el-descriptions-item>
        <el-descriptions-item label="LOGO" :span="2">
          <el-image v-if="detail.logo" :src="detail.logo" fit="cover" class="detail-logo" />
          <span v-else class="text-muted">-</span>
        </el-descriptions-item>
        <el-descriptions-item label="简介" :span="2">{{ detail.intro || '-' }}</el-descriptions-item>
      </el-descriptions>
    </el-dialog>

    <!-- 状态变更弹窗 -->
    <el-dialog v-model="statusVisible" title="更新店铺状态" width="420px">
      <el-form :model="statusForm" label-width="80px">
        <el-form-item label="店铺">
          <span>{{ statusForm.name }} (#{{ statusForm.id }})</span>
        </el-form-item>
        <el-form-item label="状态">
          <el-radio-group v-model="statusForm.status">
            <el-radio :value="0">审核中</el-radio>
            <el-radio :value="1">正常</el-radio>
            <el-radio :value="2">停用</el-radio>
          </el-radio-group>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="statusVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="onSubmitStatus">确定</el-button>
      </template>
    </el-dialog>

    <!-- 信用分调整弹窗 -->
    <el-dialog v-model="creditVisible" title="信用分调整" width="420px">
      <el-form :model="creditForm" label-width="80px">
        <el-form-item label="店铺">
          <span>{{ creditForm.name }} (#{{ creditForm.id }})</span>
        </el-form-item>
        <el-form-item label="当前分数">
          <span>{{ creditForm.current }}</span>
        </el-form-item>
        <el-form-item label="调整值">
          <el-input-number v-model="creditForm.delta" :min="-100" :max="100" />
          <span class="tip">正数加 / 负数减</span>
        </el-form-item>
        <el-form-item label="原因">
          <el-input v-model="creditForm.reason" type="textarea" :rows="2" maxlength="500" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="creditVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="onSubmitCredit">确定</el-button>
      </template>
    </el-dialog>

    <!-- 等级调整弹窗 -->
    <el-dialog v-model="levelVisible" title="等级调整" width="420px">
      <el-form :model="levelForm" label-width="80px">
        <el-form-item label="店铺">
          <span>{{ levelForm.name }} (#{{ levelForm.id }})</span>
        </el-form-item>
        <el-form-item label="当前等级">
          <el-tag type="warning">Lv{{ levelForm.current }}</el-tag>
        </el-form-item>
        <el-form-item label="新等级">
          <el-input-number v-model="levelForm.level" :min="1" :max="10" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="levelVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="onSubmitLevel">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Refresh, Search, Shop } from '@element-plus/icons-vue'
import {
  getMerchantShopList,
  getMerchantShopDetail,
  updateMerchantShopStatus,
  updateMerchantShopCredit,
  updateMerchantShopLevel
} from '@/api/merchant'

const loading = ref(false)
const list = ref([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(10)
const search = ref('')
const statusFilter = ref(null)
const ownerFilter = ref(0)

const detailVisible = ref(false)
const detail = ref(null)

const statusVisible = ref(false)
const creditVisible = ref(false)
const levelVisible = ref(false)
const submitting = ref(false)

const statusForm = reactive({ id: 0, name: '', status: 0 })
const creditForm = reactive({ id: 0, name: '', current: 100, delta: 0, reason: '' })
const levelForm = reactive({ id: 0, name: '', current: 1, level: 1 })

function statusText(s) {
  return s === 1 ? '正常' : s === 2 ? '停用' : '审核中'
}

function statusTagType(s) {
  return s === 1 ? 'success' : s === 2 ? 'danger' : 'warning'
}

async function loadList() {
  loading.value = true
  try {
    const params = { page: page.value, page_size: pageSize.value, keyword: search.value }
    if (statusFilter.value !== null && statusFilter.value !== '') params.status = statusFilter.value
    if (ownerFilter.value > 0) params.owner_id = ownerFilter.value
    const res = await getMerchantShopList(params)
    list.value = res.data?.list || []
    total.value = res.data?.total || 0
  } catch (e) {
    ElMessage.error('加载店铺列表失败')
  } finally {
    loading.value = false
  }
}

function onSearch() {
  page.value = 1
  loadList()
}

async function openDetail(row) {
  try {
    const res = await getMerchantShopDetail(row.id)
    detail.value = res.data
    detailVisible.value = true
  } catch (e) {
    ElMessage.error('加载详情失败')
  }
}

function openStatus(row) {
  statusForm.id = row.id
  statusForm.name = row.name
  statusForm.status = row.status
  statusVisible.value = true
}

async function onSubmitStatus() {
  submitting.value = true
  try {
    await updateMerchantShopStatus(statusForm.id, { status: statusForm.status })
    ElMessage.success('状态更新成功')
    statusVisible.value = false
    loadList()
  } catch (e) {
    ElMessage.error(e.message || '操作失败')
  } finally {
    submitting.value = false
  }
}

function openCredit(row) {
  creditForm.id = row.id
  creditForm.name = row.name
  creditForm.current = row.credit_score
  creditForm.delta = 0
  creditForm.reason = ''
  creditVisible.value = true
}

async function onSubmitCredit() {
  if (creditForm.delta === 0) {
    ElMessage.warning('调整值不能为 0')
    return
  }
  submitting.value = true
  try {
    await updateMerchantShopCredit(creditForm.id, {
      delta: creditForm.delta,
      reason: creditForm.reason
    })
    ElMessage.success('信用分调整成功')
    creditVisible.value = false
    loadList()
  } catch (e) {
    ElMessage.error(e.message || '操作失败')
  } finally {
    submitting.value = false
  }
}

function openLevel(row) {
  levelForm.id = row.id
  levelForm.name = row.name
  levelForm.current = row.level
  levelForm.level = row.level
  levelVisible.value = true
}

async function onSubmitLevel() {
  submitting.value = true
  try {
    await updateMerchantShopLevel(levelForm.id, { level: levelForm.level })
    ElMessage.success('等级调整成功')
    levelVisible.value = false
    loadList()
  } catch (e) {
    ElMessage.error(e.message || '操作失败')
  } finally {
    submitting.value = false
  }
}

onMounted(() => loadList())
</script>

<style scoped>
.shop-name { display: flex; align-items: center; gap: 8px; }
.shop-logo { width: 32px; height: 32px; border-radius: 4px; }
.shop-logo-placeholder { font-size: 28px; color: #c0c4cc; }
.detail-logo { width: 100px; height: 100px; border-radius: 4px; }
.pagination-wrap { margin-top: 16px; display: flex; justify-content: flex-end; }
.credit-low { color: #f56c6c; font-weight: 600; }
.tip { margin-left: 8px; color: #909399; font-size: 12px; }
.text-muted { color: #c0c4cc; }
</style>
