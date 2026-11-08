<template>
  <div class="admin-page">
    <h1 class="admin-sr-only">{{ title }}</h1>
    <header v-if="$slots.actions" class="admin-page-head">
      <div class="page-caption">
        <h2>{{ title }}</h2>
        <p v-if="description">{{ description }}</p>
      </div>
      <div class="page-actions"><slot name="actions" /></div>
    </header>
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
  description: String,
  loading: Boolean,
  error: String,
})
</script>
