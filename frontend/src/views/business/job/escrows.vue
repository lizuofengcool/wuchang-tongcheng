<template>
  <div class="app-container">
    <div class="page-card">
      <div class="toolbar">
        <el-button :icon="Refresh" @click="loadList">刷新</el-button>
      </div>
      <el-empty description="招聘模块担保交易功能开发中">
        <el-button type="primary" @click="loadList">刷新</el-button>
      </el-empty>
    </div>
  </div>
</template>

<script setup>
import { Refresh } from '@element-plus/icons-vue'
const loadList = () => { /* 预留：后续接入担保交易 API */ }
</script>

<style scoped>
.toolbar { display: flex; justify-content: space-between; margin-bottom: 16px; }
</style>
