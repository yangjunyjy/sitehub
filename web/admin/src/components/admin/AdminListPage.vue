<template>
  <main class="admin-list-page" :aria-label="title">
    <h1 class="admin-sr-only">{{ title }}</h1>
    <slot name="summary" />
    <AdminSearchForm
      v-if="$slots.filters"
      v-show="showSearch"
      :model="filterModel"
      :loading="loading"
      @search="$emit('search')"
      @reset="$emit('reset')"
    >
      <slot name="filters" />
      <template v-if="$slots.advanced" #advanced><slot name="advanced" /></template>
    </AdminSearchForm>
    <el-alert
      v-if="error"
      :title="error"
      type="error"
      show-icon
      :closable="false"
      class="admin-list-error"
    >
      <el-button link type="primary" :loading="loading" @click="$emit('refresh')"
        >重新加载</el-button
      >
    </el-alert>
    <AdminTablePanel
      :title="title"
      :loading="loading"
      :searchable="!!$slots.filters"
      :show-search="showSearch"
      v-model:current-page="currentPage"
      v-model:page-size="pageSize"
      :total="total"
      :page-sizes="pageSizes"
      :pageable="pageable"
      :columns="columns"
      :preference-key="preferenceKey"
      @toggle-search="showSearch = !showSearch"
      @refresh="$emit('refresh')"
      @page-change="$emit('page-change', $event)"
    >
      <template #actions><slot name="actions" /></template>
      <template v-if="$slots.tabs" #tabs><slot name="tabs" /></template>
      <template v-if="$slots.notice" #notice><slot name="notice" /></template>
      <template #default="{ tableHeight, density, tableProps, fullscreen }">
        <slot
          :tableHeight="tableHeight"
          :density="density"
          :tableProps="tableProps"
          :fullscreen="fullscreen"
        />
      </template>
    </AdminTablePanel>
    <slot name="footer" />
    <slot name="dialogs" />
  </main>
</template>
<script setup>
import { ref } from 'vue'
import AdminSearchForm from './AdminSearchForm.vue'
import AdminTablePanel from './AdminTablePanel.vue'
defineProps({
  title: { type: String, required: true },
  filterModel: Object,
  loading: Boolean,
  error: String,
  total: Number,
  pageSizes: Array,
  pageable: { type: Boolean, default: true },
  columns: { type: Array, default: () => [] },
  preferenceKey: String,
})
defineEmits(['search', 'reset', 'refresh', 'page-change'])
const currentPage = defineModel('currentPage', { default: 1 })
const pageSize = defineModel('pageSize', { default: 10 })
const showSearch = ref(true)
</script>
