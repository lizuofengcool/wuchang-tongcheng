<template>
  <div class="app-container">
    <div class="page-card">
      <div class="toolbar">
        <div class="toolbar-left">
          <el-button :icon="Refresh" @click="loadList">刷新</el-button>
        </div>
        <div class="toolbar-right">
          <el-input-number
            v-model="partnerFilter"
            :min="0"
            placeholder="合伙人ID"
            style="width: 140px"
            @change="onSearch"
          />
          <el-input
            v-model="codeFilter"
            placeholder="推广码"
            clearable
            style="width: 160px; margin-left: 8px"
            @keyup.enter="onSearch"
            @clear="onSearch"
          />
          <el-button type="primary" :icon="Search" style="margin-left: 8px" @click="onSearch">搜索</el-button>
        </div>
      </div>

      <el-table v-loading="loading" :data="list" border stripe>
        <el-table-column prop="id" label="ID" width="70" />
        <el-table-column prop="partner_id" label="合伙人ID" width="100" />
        <el-table-column prop="code" label="推广码" width="140" />
        <el-table-column prop="name" label="渠道名称" min-width="160" />
        <el-table-column label="点击数" width="100" align="right">
          <template #default="{ row }">{{ row.click_count || 0 }}</template>
        </el-table-column>
        <el-table-column label="注册数" width="100" align="right">
          <template #default="{ row }">{{ row.register_count || 0 }}</template>
        </el-table-column>
        <el-table-column label="订单数" width="100" align="right">
          <template #default="{ row }">{{ row.order_count || 0 }}</template>
        </el-table-column>
        <el-table-column label="累计佣金" width="130" align="right">
          <template #default="{ row }">
            <span class="amount">¥{{ formatAmount(row.commission_amount) }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="创建时间" width="170" />
        <el-table-column label="操作" width="120" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" size="small" @click="openDetail(row)">详情</el-button>
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
    <el-dialog v-model="detailVisible" title="渠道详情" width="520px">
      <el-descriptions :column="1" border>
        <el-descriptions-item label="ID">{{ detail.id }}</el-descriptions-item>
        <el-descriptions-item label="合伙人ID">{{ detail.partner_id }}</el-descriptions-item>
        <el-descriptions-item label="推广码">{{ detail.code }}</el-descriptions-item>
        <el-descriptions-item label="渠道名称">{{ detail.name || '-' }}</el-descriptions-item>
        <el-descriptions-item label="点击数">{{ detail.click_count || 0 }}</el-descriptions-item>
        <el-descriptions-item label="注册数">{{ detail.register_count || 0 }}</el-descriptions-item>
        <el-descriptions-item label="订单数">{{ detail.order_count || 0 }}</el-descriptions-item>
        <el-descriptions-item label="累计佣金">¥{{ formatAmount(detail.commission_amount) }}</el-descriptions-item>
        <el-descriptions-item label="创建时间">{{ detail.created_at || '-' }}</el-descriptions-item>
        <el-descriptions-item label="更新时间">{{ detail.updated_at || '-' }}</el-descriptions-item>
      </el-descriptions>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { Refresh, Search } from '@element-plus/icons-vue'
import { getDistributionChannelList } from '@/api/distribution'

const loading = ref(false)
const list = ref([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(10)
const partnerFilter = ref(0)
const codeFilter = ref('')

const detailVisible = ref(false)
const detail = reactive({})

function formatAmount(n) {
  if (n === undefined || n === null) return '0.00'
  return Number(n).toFixed(2)
}

async function loadList() {
  loading.value = true
  try {
    const params = { page: page.value, page_size: pageSize.value }
    if (partnerFilter.value > 0) params.partner_id = partnerFilter.value
    if (codeFilter.value) params.code = codeFilter.value
    const res = await getDistributionChannelList(params)
    list.value = res.data?.list || []
    total.value = res.data?.total || 0
  } catch (e) {
    ElMessage.error('加载渠道列表失败')
  } finally {
    loading.value = false
  }
}

function onSearch() {
  page.value = 1
  loadList()
}

function openDetail(row) {
  Object.assign(detail, row)
  detailVisible.value = true
}

onMounted(() => loadList())
</script>

<style scoped>
.pagination-wrap { margin-top: 16px; display: flex; justify-content: flex-end; }
.amount { font-weight: 600; color: #e6a23c; }
</style>
