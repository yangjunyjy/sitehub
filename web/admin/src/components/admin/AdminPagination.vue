<template>
  <div class="admin-pagination">
    <el-pagination
      :current-page="currentPage"
      :page-size="pageSize"
      :total="total"
      :disabled="disabled"
      :page-sizes="pageSizes"
      :pager-count="5"
      layout="total, prev, pager, next, sizes, jumper"
      background
      @update:current-page="changePage"
      @update:page-size="changeSize"
    />
  </div>
</template>
<script setup>
import { nextTick } from 'vue'
const props = defineProps({
  currentPage: { type: Number, default: 1 },
  pageSize: { type: Number, default: 10 },
  total: { type: Number, default: 0 },
  disabled: Boolean,
  pageSizes: { type: Array, default: () => [10, 20, 50, 100] },
})
const emit = defineEmits(['update:currentPage', 'update:pageSize', 'change'])
let scheduled = false
function notify() {
  if (scheduled) return
  scheduled = true
  nextTick(() => {
    scheduled = false
    emit('change', { page: props.currentPage, pageSize: props.pageSize })
  })
}
function changePage(page) {
  if (page === props.currentPage) return
  emit('update:currentPage', page)
  notify()
}
function changeSize(size) {
  if (size === props.pageSize) return
  emit('update:pageSize', size)
  emit('update:currentPage', 1)
  notify()
}
</script>
