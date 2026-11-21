<script setup>
import ReviewQueue from '@/components/community/ReviewQueue.vue'

const fields = [
  { key: 'name', prop: 'name', label: '网站名称', minWidth: 110 },
  { key: 'url', prop: 'url', label: '网站地址', minWidth: 190 },
  { key: 'content', prop: 'content', label: '网站描述', minWidth: 190 },
]
</script>
<template><ReviewQueue title="友链管理" kind="friends" :fields="fields" /></template>
