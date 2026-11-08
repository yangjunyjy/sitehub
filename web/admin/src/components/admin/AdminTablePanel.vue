<template>
  <section
    ref="panel"
    class="admin-table-panel"
    :class="{ 'is-fullscreen': fullscreen }"
    :data-density="density"
    :aria-label="title"
    :role="fullscreen ? 'dialog' : 'region'"
    :aria-modal="fullscreen || undefined"
    tabindex="-1"
    @keydown.esc.stop="fullscreen = false"
    @keydown.tab="trapFocus"
  >
    <div ref="toolbar" class="admin-table-toolbar">
      <div class="admin-table-actions"><slot name="actions" /></div>
      <div class="admin-table-tools">
        <el-tooltip
          v-if="searchable"
          :content="showSearch ? '隐藏查询条件' : '显示查询条件'"
          :show-after="400"
        >
          <el-button
            class="admin-tool-button"
            :class="{ 'is-active': showSearch }"
            :icon="Search"
            :aria-label="showSearch ? '隐藏查询条件' : '显示查询条件'"
            :aria-pressed="showSearch"
            @click="$emit('toggle-search')"
          />
        </el-tooltip>
        <el-tooltip content="刷新列表" :show-after="400"
          ><el-button
            class="admin-tool-button"
            :icon="Refresh"
            :loading="loading"
            aria-label="刷新列表"
            @click="$emit('refresh')"
        /></el-tooltip>
        <el-dropdown trigger="click" @command="setDensity">
          <el-button class="admin-tool-button" :icon="Rank" aria-label="表格密度" />
          <template #dropdown
            ><el-dropdown-menu>
              <el-dropdown-item
                v-for="item in densities"
                :key="item.value"
                :command="item.value"
                :class="{ 'is-active': density === item.value }"
                >{{ item.label }}{{ density === item.value ? ' ✓' : '' }}</el-dropdown-item
              >
            </el-dropdown-menu></template
          >
        </el-dropdown>
        <el-tooltip :content="fullscreen ? '退出表格全屏' : '表格全屏'" :show-after="400"
          ><el-button
            class="admin-tool-button"
            :icon="fullscreen ? Close : FullScreen"
            :aria-label="fullscreen ? '退出表格全屏' : '表格全屏'"
            :aria-pressed="fullscreen"
            @click="fullscreen = !fullscreen"
        /></el-tooltip>
        <AdminTableSettings
          :columns="orderedColumns"
          :hidden="preferences.hidden"
          :appearance="preferences.appearance"
          :can-hide="canHide"
          :neighbor="neighbor"
          @visibility="setVisible"
          @move="move"
          @appearance="setAppearance"
          @reset-columns="resetColumns"
          @reset-appearance="resetAppearance"
        />
      </div>
    </div>
    <div v-if="$slots.tabs" class="admin-table-tabs"><slot name="tabs" /></div>
    <div v-if="$slots.notice" class="admin-table-notice"><slot name="notice" /></div>
    <div ref="body" v-loading="loading" class="admin-table-body">
      <slot
        :tableHeight="tableHeight"
        :density="density"
        :tableProps="tableProps"
        :fullscreen="fullscreen"
      />
    </div>
    <footer v-if="pageable" ref="footer">
      <AdminPagination
        v-model:current-page="currentPage"
        v-model:page-size="pageSize"
        :total="total"
        :disabled="loading"
        :page-sizes="pageSizes"
        @change="$emit('page-change', $event)"
      />
    </footer>
  </section>
</template>
<script setup>
import {
  computed,
  ref,
  watch,
  nextTick,
  onMounted,
  onBeforeUnmount,
  onActivated,
  onDeactivated,
} from 'vue'
import { Search, Refresh, Rank, FullScreen, Close } from '@element-plus/icons-vue'
import AdminPagination from './AdminPagination.vue'
import AdminTableSettings from './AdminTableSettings.vue'
import { useAdminTablePreferences } from '@/composables/useAdminTablePreferences'
const props = defineProps({
  title: String,
  loading: Boolean,
  searchable: Boolean,
  showSearch: Boolean,
  total: Number,
  pageSizes: Array,
  pageable: { type: Boolean, default: true },
  columns: { type: Array, default: () => [] },
  preferenceKey: String,
})
const {
  preferences,
  orderedColumns,
  tableProps,
  canHide,
  setVisible,
  move,
  neighbor,
  resetColumns,
  setAppearance,
  resetAppearance,
  setDensity,
} = useAdminTablePreferences(props)
const density = computed(() => preferences.value.density)
defineEmits(['refresh', 'toggle-search', 'page-change'])
const currentPage = defineModel('currentPage', { default: 1 })
const pageSize = defineModel('pageSize', { default: 10 })
const panel = ref(),
  toolbar = ref(),
  body = ref(),
  footer = ref()
const fullscreen = ref(false),
  tableHeight = ref(480)
const densities = [
  { label: '舒适', value: 'comfortable' },
  { label: '标准', value: 'default' },
  { label: '紧凑', value: 'compact' },
]
let observer, measureFrame, previousFocus
function scheduleMeasure() {
  if (measureFrame) return
  measureFrame = requestAnimationFrame(() => {
    measureFrame = undefined
    measure()
  })
}
function measure() {
  if (!body.value?.getClientRects().length) return
  // Use the actual card position: search wrapping, summary cards and sidebar changes vary by page.
  tableHeight.value = Math.max(
    280,
    Math.floor(
      window.innerHeight -
        body.value.getBoundingClientRect().top -
        (footer.value?.offsetHeight || 0) -
        24,
    ),
  )
}
function trapFocus(event) {
  if (!fullscreen.value) return
  const elements = [
    ...panel.value.querySelectorAll(
      'button:not([disabled]), input:not([disabled]), [tabindex="0"], a[href]',
    ),
  ].filter((element) => element.getClientRects().length)
  const first = elements[0],
    last = elements.at(-1)
  if (
    event.shiftKey &&
    (document.activeElement === first || document.activeElement === panel.value)
  ) {
    event.preventDefault()
    last?.focus()
  } else if (!event.shiftKey && document.activeElement === last) {
    event.preventDefault()
    first?.focus()
  }
}
watch(fullscreen, async (value) => {
  if (value) previousFocus = document.activeElement
  await nextTick()
  if (value) panel.value?.focus()
  else if (previousFocus?.isConnected) previousFocus.focus()
})
watch([fullscreen, () => props.showSearch], async () => {
  await nextTick()
  scheduleMeasure()
})
onMounted(async () => {
  // Element Plus defers its initial max-height assignment. Measure after that
  // flush so its initial value cannot overwrite the measured height.
  await nextTick()
  if (!panel.value) return
  observer = new ResizeObserver(scheduleMeasure)
  for (const element of [panel.value.parentElement, toolbar.value, footer.value])
    if (element) observer.observe(element)
  window.addEventListener('resize', scheduleMeasure)
  measure()
})
onActivated(() => nextTick(scheduleMeasure))
onDeactivated(() => {
  fullscreen.value = false
})
onBeforeUnmount(() => {
  observer?.disconnect()
  cancelAnimationFrame(measureFrame)
  window.removeEventListener('resize', scheduleMeasure)
})
</script>
