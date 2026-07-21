<template>
  <div class="app-container">
    <div class="page-card">
      <div class="toolbar">
        <div class="toolbar-left">
          <el-button :icon="Refresh" @click="loadList">刷新</el-button>
          <el-button :icon="Share" @click="openTree">上下级树</el-button>
        </div>
        <div class="toolbar-right">
          <el-input
            v-model="keywordFilter"
            placeholder="关键字"
            clearable
            style="width: 160px"
            @keyup.enter="onSearch"
            @clear="onSearch"
          />
          <el-select
            v-model="levelFilter"
            placeholder="等级"
            clearable
            style="width: 140px; margin-left: 8px"
            @change="onSearch"
          >
            <el-option label="普通合伙人" :value="1" />
            <el-option label="高级合伙人" :value="2" />
            <el-option label="城市合伙人" :value="3" />
          </el-select>
          <el-select
            v-model="statusFilter"
            placeholder="状态"
            clearable
            style="width: 140px; margin-left: 8px"
            @change="onSearch"
          >
            <el-option label="待审核" :value="0" />
            <el-option label="正常" :value="1" />
            <el-option label="冻结" :value="2" />
            <el-option label="拒绝" :value="3" />
            <el-option label="退出" :value="4" />
          </el-select>
          <el-button type="primary" :icon="Search" style="margin-left: 8px" @click="onSearch">搜索</el-button>
        </div>
      </div>

      <el-table v-loading="loading" :data="list" border stripe>
        <el-table-column prop="id" label="ID" width="70" />
        <el-table-column prop="user_id" label="用户ID" width="90" />
        <el-table-column label="等级" width="110">
          <template #default="{ row }">
            <el-tag :type="levelTagType(row.level)" size="small">
              {{ row.level_text || levelText(row.level) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="佣金率" width="100" align="right">
          <template #default="{ row }">{{ formatRate(row.commission_rate) }}</template>
        </el-table-column>
        <el-table-column label="累计佣金" width="120" align="right">
          <template #default="{ row }">
            <span class="amount-total">¥{{ formatAmount(row.total_commission) }}</span>
          </template>
        </el-table-column>
        <el-table-column label="已结算" width="120" align="right">
          <template #default="{ row }">¥{{ formatAmount(row.settled_commission) }}</template>
        </el-table-column>
        <el-table-column label="待结算" width="120" align="right">
          <template #default="{ row }">¥{{ formatAmount(row.pending_commission) }}</template>
        </el-table-column>
        <el-table-column label="上级ID" width="80">
          <template #default="{ row }">{{ row.parent_id || '-' }}</template>
        </el-table-column>
        <el-table-column label="下级数" width="80" align="center">
          <template #default="{ row }">{{ row.child_count || 0 }}</template>
        </el-table-column>
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="statusTagType(row.status)" size="small">
              {{ row.status_text || statusText(row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="joined_at" label="加入时间" width="170">
          <template #default="{ row }">{{ row.joined_at || '-' }}</template>
        </el-table-column>
        <el-table-column label="操作" width="280" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" size="small" @click="openEdit(row)">编辑</el-button>
            <el-button
              v-if="row.status === 0 || row.status === 2"
              link
              type="success"
              size="small"
              @click="onChangeStatus(row, 1)"
            >激活</el-button>
            <el-button
              v-if="row.status === 1"
              link
              type="warning"
              size="small"
              @click="onChangeStatus(row, 2)"
            >冻结</el-button>
            <el-button link type="primary" size="small" @click="openUpgrade(row)">升级</el-button>
            <el-button link type="primary" size="small" @click="openRate(row)">调佣金率</el-button>
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

    <!-- 编辑弹窗 -->
    <el-dialog v-model="editVisible" title="编辑合伙人" width="520px">
      <el-form :model="editForm" label-width="110px">
        <el-form-item label="合伙人ID">
          <span>#{{ editForm.id }}（用户 #{{ editForm.user_id }}）</span>
        </el-form-item>
        <el-form-item label="等级">
          <el-select v-model="editForm.level" style="width: 100%">
            <el-option label="普通合伙人" :value="1" />
            <el-option label="高级合伙人" :value="2" />
            <el-option label="城市合伙人" :value="3" />
          </el-select>
        </el-form-item>
        <el-form-item label="佣金比例">
          <el-input-number v-model="editForm.commission_rate" :min="0" :max="1" :step="0.01" :precision="4" />
          <span class="tip">0-1 之间，如 0.10 表示 10%</span>
        </el-form-item>
        <el-form-item label="上级合伙人ID">
          <el-input-number v-model="editForm.parent_id" :min="0" />
          <span class="tip">0=无上级</span>
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="editForm.status" style="width: 100%">
            <el-option label="待审核" :value="0" />
            <el-option label="正常" :value="1" />
            <el-option label="冻结" :value="2" />
            <el-option label="拒绝" :value="3" />
            <el-option label="退出" :value="4" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="editVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="onSubmitEdit">保存</el-button>
      </template>
    </el-dialog>

    <!-- 升级弹窗 -->
    <el-dialog v-model="upgradeVisible" title="升级合伙人等级" width="460px">
      <el-form :model="upgradeForm" label-width="100px">
        <el-form-item label="合伙人">
          <span>#{{ upgradeForm.id }}（当前等级：{{ levelText(upgradeForm.current_level) }}）</span>
        </el-form-item>
        <el-form-item label="目标等级">
          <el-select v-model="upgradeForm.target_level" style="width: 100%">
            <el-option label="普通合伙人" :value="1" :disabled="upgradeForm.current_level >= 1" />
            <el-option label="高级合伙人" :value="2" :disabled="upgradeForm.current_level >= 2" />
            <el-option label="城市合伙人" :value="3" :disabled="upgradeForm.current_level >= 3" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="upgradeVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="onSubmitUpgrade">升级</el-button>
      </template>
    </el-dialog>

    <!-- 调整佣金率弹窗 -->
    <el-dialog v-model="rateVisible" title="调整佣金比例" width="460px">
      <el-form :model="rateForm" label-width="100px">
        <el-form-item label="合伙人">
          <span>#{{ rateForm.id }}</span>
        </el-form-item>
        <el-form-item label="当前佣金率">
          <span>{{ formatRate(rateForm.current_rate) }}</span>
        </el-form-item>
        <el-form-item label="新佣金率">
          <el-input-number v-model="rateForm.commission_rate" :min="0" :max="1" :step="0.01" :precision="4" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="rateVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="onSubmitRate">保存</el-button>
      </template>
    </el-dialog>

    <!-- 上下级树弹窗 -->
    <el-dialog v-model="treeVisible" title="合伙人上下级树" width="640px">
      <el-form :model="treeForm" label-width="100px" inline>
        <el-form-item label="根合伙人ID">
          <el-input-number v-model="treeForm.parent_id" :min="0" placeholder="0=全部顶级" />
        </el-form-item>
        <el-form-item label="深度">
          <el-input-number v-model="treeForm.depth" :min="1" :max="5" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" :loading="treeLoading" @click="loadTree">加载</el-button>
        </el-form-item>
      </el-form>
      <el-tree
        v-loading="treeLoading"
        :data="treeData"
        :props="treeProps"
        default-expand-all
        node-key="id"
      >
        <template #default="{ data }">
          <span>
            <el-tag :type="levelTagType(data.level)" size="small" style="margin-right: 6px">
              {{ data.level_text || levelText(data.level) }}
            </el-tag>
            <span>ID#{{ data.id }} 用户#{{ data.user_id }}</span>
            <span style="margin-left: 8px; color: #909399">下级 {{ data.child_count || 0 }}</span>
          </span>
        </template>
      </el-tree>
      <div v-if="!treeData.length && !treeLoading" class="empty-tip">暂无数据</div>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Refresh, Search, Share } from '@element-plus/icons-vue'
import {
  getDistributionPartnerList,
  updateDistributionPartner,
  updateDistributionPartnerStatus,
  upgradeDistributionPartner,
  adjustDistributionPartnerRate,
  getDistributionPartnerTree
} from '@/api/distribution'

const loading = ref(false)
const list = ref([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(10)
const keywordFilter = ref('')
const levelFilter = ref(null)
const statusFilter = ref(null)

const editVisible = ref(false)
const upgradeVisible = ref(false)
const rateVisible = ref(false)
const treeVisible = ref(false)
const submitting = ref(false)
const treeLoading = ref(false)

const editForm = reactive({
  id: 0, user_id: 0, level: 1, commission_rate: 0, parent_id: 0, status: 0
})
const upgradeForm = reactive({ id: 0, current_level: 1, target_level: 2 })
const rateForm = reactive({ id: 0, current_rate: 0, commission_rate: 0 })
const treeForm = reactive({ parent_id: 0, depth: 2 })
const treeData = ref([])
const treeProps = { children: 'children', label: 'id' }

function levelText(l) {
  return l === 3 ? '城市合伙人' : l === 2 ? '高级合伙人' : '普通合伙人'
}
function levelTagType(l) {
  return l === 3 ? 'danger' : l === 2 ? 'warning' : 'info'
}
function statusText(s) {
  return ['待审核', '正常', '冻结', '拒绝', '退出'][s] || ''
}
function statusTagType(s) {
  return s === 1 ? 'success' : s === 2 ? 'warning' : s === 3 ? 'danger' : s === 4 ? 'info' : ''
}
function formatAmount(n) {
  if (n === undefined || n === null) return '0.00'
  return Number(n).toFixed(2)
}
function formatRate(r) {
  if (r === undefined || r === null) return '0.00%'
  return (Number(r) * 100).toFixed(2) + '%'
}

async function loadList() {
  loading.value = true
  try {
    const params = { page: page.value, page_size: pageSize.value }
    if (keywordFilter.value) params.keyword = keywordFilter.value
    if (levelFilter.value !== null && levelFilter.value !== '') params.level = levelFilter.value
    if (statusFilter.value !== null && statusFilter.value !== '') params.status = statusFilter.value
    const res = await getDistributionPartnerList(params)
    list.value = res.data?.list || []
    total.value = res.data?.total || 0
  } catch (e) {
    ElMessage.error('加载合伙人列表失败')
  } finally {
    loading.value = false
  }
}

function onSearch() {
  page.value = 1
  loadList()
}

function openEdit(row) {
  Object.assign(editForm, {
    id: row.id,
    user_id: row.user_id,
    level: row.level,
    commission_rate: row.commission_rate,
    parent_id: row.parent_id || 0,
    status: row.status
  })
  editVisible.value = true
}

async function onSubmitEdit() {
  submitting.value = true
  try {
    await updateDistributionPartner(editForm.id, {
      level: editForm.level,
      commission_rate: editForm.commission_rate,
      parent_id: editForm.parent_id,
      status: editForm.status
    })
    ElMessage.success('保存成功')
    editVisible.value = false
    loadList()
  } catch (e) {
    ElMessage.error(e.message || '操作失败')
  } finally {
    submitting.value = false
  }
}

async function onChangeStatus(row, status) {
  try {
    await ElMessageBox.confirm(`确定将合伙人 #${row.id} 状态改为「${statusText(status)}」吗？`, '提示', { type: 'warning' })
    await updateDistributionPartnerStatus(row.id, { status })
    ElMessage.success('状态已更新')
    loadList()
  } catch (e) {
    if (e !== 'cancel') ElMessage.error(e.message || '操作失败')
  }
}

function openUpgrade(row) {
  upgradeForm.id = row.id
  upgradeForm.current_level = row.level
  upgradeForm.target_level = Math.min(row.level + 1, 3)
  upgradeVisible.value = true
}

async function onSubmitUpgrade() {
  submitting.value = true
  try {
    await upgradeDistributionPartner(upgradeForm.id, { level: upgradeForm.target_level })
    ElMessage.success('升级成功')
    upgradeVisible.value = false
    loadList()
  } catch (e) {
    ElMessage.error(e.message || '操作失败')
  } finally {
    submitting.value = false
  }
}

function openRate(row) {
  rateForm.id = row.id
  rateForm.current_rate = row.commission_rate
  rateForm.commission_rate = row.commission_rate
  rateVisible.value = true
}

async function onSubmitRate() {
  submitting.value = true
  try {
    await adjustDistributionPartnerRate(rateForm.id, { commission_rate: rateForm.commission_rate })
    ElMessage.success('佣金率已更新')
    rateVisible.value = false
    loadList()
  } catch (e) {
    ElMessage.error(e.message || '操作失败')
  } finally {
    submitting.value = false
  }
}

function openTree() {
  treeVisible.value = true
  treeData.value = []
}

async function loadTree() {
  treeLoading.value = true
  try {
    const res = await getDistributionPartnerTree({
      parent_id: treeForm.parent_id,
      depth: treeForm.depth
    })
    const data = res.data
    treeData.value = Array.isArray(data) ? data : (data ? [data] : [])
  } catch (e) {
    ElMessage.error('加载上下级树失败')
  } finally {
    treeLoading.value = false
  }
}

onMounted(() => loadList())
</script>

<style scoped>
.pagination-wrap { margin-top: 16px; display: flex; justify-content: flex-end; }
.amount-total { font-weight: 600; color: #e6a23c; }
.tip { margin-left: 8px; color: #909399; font-size: 12px; }
.empty-tip { padding: 16px; text-align: center; color: #c0c4cc; }
</style>
