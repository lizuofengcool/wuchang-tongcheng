<template>
  <div class="app-container">
    <div class="page-card">
      <div class="toolbar">
        <el-button :icon="Refresh" @click="loadList">刷新</el-button>
      </div>
      <el-table v-loading="loading" :data="list" border stripe>
        <el-table-column type="index" label="#" width="60" />
        <el-table-column label="日期" width="140">
          <template #default="{ row }">{{ row.date || row.label || '-' }}</template>
        </el-table-column>
        <el-table-column label="平均薪资(K)" width="140">
          <template #default="{ row }">{{ row.value || row.avg_salary || row.salary || 0 }}</template>
        </el-table-column>
        <el-table-column label="最低薪资(K)" width="140">
          <template #default="{ row }">{{ row.min || row.min_salary || 0 }}</template>
        </el-table-column>
        <el-table-column label="最高薪资(K)" width="140">
          <template #default="{ row }">{{ row.max || row.max_salary || 0 }}</template>
        </el-table-column>
        <el-table-column label="中位数(K)" width="140">
          <template #default="{ row }">{{ row.median || 0 }}</template>
        </el-table-column>
      </el-table>
      <div v-if="!loading && !list.length" class="empty-text">暂无薪资数据</div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { Refresh } from '@element-plus/icons-vue'
import { getSalaryTrend } from '@/api/job'

const loading = ref(false)
const list = ref([])

const loadList = async () => {
  loading.value = true
  try {
    const res = await getSalaryTrend({ days: 30 })
    const d = res.data || {}
    const dates = d.dates || d.x_axis || []
    const values = d.values || d.avg_salaries || d.data || []
    if (Array.isArray(dates) && dates.length) {
      list.value = dates.map((date, i) => ({ date, value: values[i] || 0 }))
    } else if (Array.isArray(d) && d.length) {
      list.value = d
    } else {
      list.value = []
    }
  } catch (e) { list.value = [] } finally { loading.value = false }
}

onMounted(() => loadList())
</script>

<style scoped>
.toolbar { display: flex; justify-content: space-between; margin-bottom: 12px; }
.empty-text { text-align: center; color: #999; padding: 40px; }
</style>
