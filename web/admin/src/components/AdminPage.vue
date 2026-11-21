<template>
  <div class="admin-page">
    <h1 class="admin-sr-only">{{ title }}</h1>
    <div v-if="$slots.actions" class="page-actions" role="group" :aria-label="title + '操作'">
      <slot name="actions" />
    </div>
    <el-alert
      v-if="error"
      :title="error"
      type="error"
      show-icon
      :closable="false"
      class="page-error"
    />
    <div v-loading="loading" class="page-body"><slot /></div>
  </div>
</template>
<script setup>
defineProps({
  title: String,
  loading: Boolean,
  error: String,
})
</script>
