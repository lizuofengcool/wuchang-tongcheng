<template>
  <div class="app-container">
    <div class="page-card">
      <div class="toolbar">
        <el-input v-model="keyword" placeholder="搜索福利标签" clearable style="width: 240px" :prefix-icon="Search" />
        <el-button type="primary" :icon="Plus">新建福利</el-button>
      </div>
      <el-empty description="福利标签管理功能开发中">
        <el-button type="primary" @click="loadList">刷新</el-button>
      </el-empty>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { Search, Plus } from '@element-plus/icons-vue'

const keyword = ref('')
const loadList = () => { /* 预留：后续接入福利标签 API */ }
</script>

<style scoped>
.toolbar { display: flex; justify-content: space-between; margin-bottom: 16px; }
</style>
