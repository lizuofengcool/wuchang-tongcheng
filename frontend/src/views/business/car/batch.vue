<template>
  <div class="app-container">
    <div class="page-card">
      <el-tabs v-model="activeTab">
        <el-tab-pane label="批量审核" name="audit">
          <div class="toolbar">
            <el-button type="success" :icon="Check" :disabled="!selection.length" @click="onBatchAudit(1)">批量通过</el-button>
            <el-button type="danger" :icon="Close" :disabled="!selection.length" @click="onBatchAudit(2)">批量拒绝</el-button>
            <el-button :icon="Refresh" @click="loadList">刷新</el-button>
          </div>
          <el-table v-loading="loading" :data="list" border stripe @selection-change="onSelectionChange">
            <el-table-column type="selection" width="44" fixed="left" />
            <el-table-column prop="id" label="ID" width="70" />
            <el-table-column prop="title" label="标题" min-width="160" show-overflow-tooltip />
            <el-table-column label="价格" width="120">
              <template #default="{ row }">¥{{ Number(row.price || 0).toFixed(2) }}万</template>
            </el-table-column>
            <el-table-column label="审核状态" width="100">
              <template #default="{ row }">
                <el-tag :type="auditTagType(row.audit_status)" size="small">{{ auditText(row.audit_status) }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column label="创建时间" width="160">
              <template #default="{ row }">{{ formatTime(row.created_at) }}</template>
            </el-table-column>
          </el-table>
        </el-tab-pane>

        <el-tab-pane label="批量状态变更" name="status">
          <div class="toolbar">
            <el-button type="primary" :icon="Promotion" :disabled="!selection.length" @click="onBatchStatus(1)">批量发布</el-button>
            <el-button type="warning" :icon="Bottom" :disabled="!selection.length" @click="onBatchStatus(2)">批量下架</el-button>
            <el-button :icon="Refresh" @click="loadList">刷新</el-button>
          </div>
          <el-table v-loading="loading" :data="list" border stripe @selection-change="onSelectionChange">
            <el-table-column type="selection" width="44" fixed="left" />
            <el-table-column prop="id" label="ID" width="70" />
            <el-table-column prop="title" label="标题" min-width="160" show-overflow-tooltip />
            <el-table-column label="价格" width="120">
              <template #default="{ row }">¥{{ Number(row.price || 0).toFixed(2) }}万</template>
            </el-table-column>
            <el-table-column label="状态" width="100">
              <template #default="{ row }">
                <el-tag :type="statusTagType(row.status)" size="small">{{ statusText(row.status) }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column label="创建时间" width="160">
              <template #default="{ row }">{{ formatTime(row.created_at) }}</template>
            </el-table-column>
          </el-table>
        </el-tab-pane>

        <el-tab-pane label="批量删除" name="delete">
          <div class="toolbar">
            <el-popconfirm title="确认批量删除所选车源？此操作不可撤销" @confirm="onBatchDelete">
              <template #reference>
                <el-button type="danger" :icon="Delete" :disabled="!selection.length">批量删除</el-button>
              </template>
            </el-popconfirm>
            <el-button :icon="Refresh" @click="loadList">刷新</el-button>
          </div>
          <el-table v-loading="loading" :data="list" border stripe @selection-change="onSelectionChange">
            <el-table-column type="selection" width="44" fixed="left" />
            <el-table-column prop="id" label="ID" width="70" />
            <el-table-column prop="title" label="标题" min-width="160" show-overflow-tooltip />
            <el-table-column label="价格" width="120">
              <template #default="{ row }">¥{{ Number(row.price || 0).toFixed(2) }}万</template>
            </el-table-column>
            <el-table-column label="状态" width="100">
              <template #default="{ row }">
                <el-tag :type="statusTagType(row.status)" size="small">{{ statusText(row.status) }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column label="创建时间" width="160">
              <template #default="{ row }">{{ formatTime(row.created_at) }}</template>
            </el-table-column>
          </el-table>
        </el-tab-pane>
      </el-tabs>

      <div class="pagination-wrap">
        <el-pagination v-model:current-page="page" v-model:page-size="pageSize" :total="total" :page-sizes="[10, 20, 50]" layout="total, sizes, prev, pager, next, jumper" background @current-change="loadList" @size-change="loadList" />
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Check, Close, Promotion, Bottom, Delete, Refresh } from '@element-plus/icons-vue'
import { adminListCars, auditCar, adminUpdateCarStatus, deleteCar } from '@/api/car'
import { formatTime } from '@/utils/format'

const loading = ref(false)
const list = ref([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const selection = ref([])
const activeTab = ref('audit')
const filters = reactive({})

const auditText = (s) => ({ 0: '待审核', 1: '已通过', 2: '已拒绝' }[s] || '-')
const auditTagType = (s) => ({ 0: 'warning', 1: 'success', 2: 'danger' }[s] || 'info')
const statusText = (s) => ({ 0: '草稿', 1: '已发布', 2: '已下架', 3: '已售出' }[s] || '-')
const statusTagType = (s) => ({ 0: 'info', 1: 'success', 2: 'warning', 3: 'primary' }[s] || 'info')

const buildParams = () => ({ page: page.value, page_size: pageSize.value })

const loadList = async () => {
  loading.value = true
  try {
    const res = await adminListCars(buildParams())
    const d = res.data || {}
    list.value = d.list || []
    total.value = d.total || 0
  } catch (e) { list.value = [] } finally { loading.value = false }
}

const onSelectionChange = (rows) => { selection.value = rows }

const onBatchAudit = async (status) => {
  try {
    if (status === 2) {
      const { value } = await ElMessageBox.prompt('请输入批量拒绝原因', '批量拒绝', { type: 'warning', inputType: 'textarea' })
      await Promise.all(selection.value.map((r) => auditCar(r.id, { audit_status: status, audit_reason: value })))
    } else {
      await Promise.all(selection.value.map((r) => auditCar(r.id, { audit_status: status })))
    }
    ElMessage.success('批量审核完成')
    loadList()
  } catch (e) {
    if (e !== 'cancel' && e?.message) ElMessage.error(e.message)
  }
}

const onBatchStatus = async (status) => {
  try {
    await ElMessageBox.confirm(`确认批量${status === 1 ? '发布' : '下架'} ${selection.value.length} 条车源？`, '提示', { type: 'warning' })
    await Promise.all(selection.value.map((r) => adminUpdateCarStatus(r.id, { status })))
    ElMessage.success('批量操作完成')
    loadList()
  } catch (e) {
    if (e !== 'cancel' && e?.message) ElMessage.error(e.message)
  }
}

const onBatchDelete = async () => {
  try {
    await Promise.all(selection.value.map((r) => deleteCar(r.id)))
    ElMessage.success('批量删除完成')
    loadList()
  } catch (e) {
    if (e?.message) ElMessage.error(e.message)
  }
}

onMounted(() => loadList())
</script>

<style scoped>
.page-card { background: #fff; padding: 16px; border-radius: 4px; box-shadow: 0 1px 4px rgba(0,0,0,0.04); }
.toolbar { display: flex; justify-content: flex-start; gap: 8px; margin-bottom: 12px; }
.pagination-wrap { margin-top: 16px; display: flex; justify-content: flex-end; }
</style>
