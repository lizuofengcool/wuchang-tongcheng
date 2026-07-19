<template>
  <PlaceholderView
    title="招聘求职"
    description="五常同城招聘求职模块，提供职位发布、简历投递、企业认证等功能"
    module="job"
    icon="Briefcase"
    :features="['职位发布', '简历投递', '职位搜索', '企业认证', '面试邀请', '薪资范围']"
  />
</template>

<script setup>
import PlaceholderView from '@/components/PlaceholderView.vue'
</script>
