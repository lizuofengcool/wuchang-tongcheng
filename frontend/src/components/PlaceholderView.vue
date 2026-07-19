<template>
  <div class="placeholder-container">
    <div class="placeholder-card">
      <el-icon :size="64" class="placeholder-icon">
        <component :is="icon || 'InfoFilled'" />
      </el-icon>
      <h2 class="placeholder-title">{{ title }}</h2>
      <p class="placeholder-desc">{{ description || '该模块正在开发中，敬请期待' }}</p>
      <div class="placeholder-meta">
        <el-tag size="small" type="info">模块：{{ module }}</el-tag>
        <el-tag size="small" type="warning">状态：开发中</el-tag>
        <el-tag size="small" type="success">路由：{{ $route.path }}</el-tag>
      </div>
      <div class="placeholder-plan">
        <h4>规划功能：</h4>
        <ul>
          <li v-for="(f, i) in features" :key="i">{{ f }}</li>
        </ul>
      </div>
    </div>
  </div>
</template>

<script setup>
defineProps({
  title: { type: String, default: '模块开发中' },
  description: { type: String, default: '' },
  module: { type: String, default: '' },
  icon: { type: String, default: 'InfoFilled' },
  features: { type: Array, default: () => [] }
})
</script>

<style scoped>
.placeholder-container {
  display: flex;
  justify-content: center;
  align-items: flex-start;
  padding: 40px 20px;
}

.placeholder-card {
  background: #fff;
  border-radius: 8px;
  padding: 48px 40px;
  max-width: 640px;
  width: 100%;
  text-align: center;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.08);
}

.placeholder-icon {
  color: #409eff;
  margin-bottom: 16px;
}

.placeholder-title {
  font-size: 24px;
  color: #303133;
  margin: 0 0 12px;
}

.placeholder-desc {
  font-size: 14px;
  color: #909399;
  margin: 0 0 24px;
}

.placeholder-meta {
  display: flex;
  justify-content: center;
  gap: 8px;
  flex-wrap: wrap;
  margin-bottom: 24px;
}

.placeholder-plan {
  text-align: left;
  background: #f5f7fa;
  border-radius: 4px;
  padding: 16px 20px;
}

.placeholder-plan h4 {
  margin: 0 0 12px;
  color: #303133;
  font-size: 14px;
}

.placeholder-plan ul {
  margin: 0;
  padding-left: 20px;
}

.placeholder-plan li {
  font-size: 13px;
  color: #606266;
  line-height: 1.8;
}
</style>
