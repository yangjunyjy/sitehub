<template>
  <el-table
    ref="table"
    v-bind="$attrs"
    :stripe="stripe"
    :border="border"
    :class="{ 'admin-table--header-background': headerBackground }"
  >
    <el-table-column v-for="column in columns" :key="column.key" v-bind="columnProps(column)">
      <template v-if="$slots[column.key]" #default="scope"
        ><slot :name="column.key" v-bind="scope"
      /></template>
    </el-table-column>
    <template v-if="$slots.empty" #empty><slot name="empty" /></template>
    <template v-if="$slots.append" #append><slot name="append" /></template>
  </el-table>
</template>
<script setup>
import { ref, onMounted, onBeforeUnmount } from 'vue'
defineOptions({ inheritAttrs: false })
defineProps({
  columns: { type: Array, default: () => [] },
  stripe: Boolean,
  border: Boolean,
  headerBackground: Boolean,
})
const table = ref()
const narrow = ref(false)
let observer
onMounted(() => {
  // Fixed action columns can cover the entire title on narrow containers.
  // Observe the table itself so drawers and a collapsed sidebar work too.
  observer = new ResizeObserver(([entry]) => {
    if (entry.contentRect.width > 0) narrow.value = entry.contentRect.width < 720
  })
  if (table.value?.$el) observer.observe(table.value.$el)
})
onBeforeUnmount(() => observer?.disconnect())
function columnProps(column) {
  const { key, locked, utility, ...rest } = column
  return { ...rest, fixed: narrow.value ? false : rest.fixed, columnKey: key }
}
// Use public Element Plus methods; preserve selection and tree state when columns move.
defineExpose({
  clearSelection: () => table.value?.clearSelection(),
  toggleRowExpansion: (row, expanded) => table.value?.toggleRowExpansion(row, expanded),
})
</script>
