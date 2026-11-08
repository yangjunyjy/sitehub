<template>
  <el-popover
    v-if="columns.length"
    v-model:visible="columnsOpen"
    placement="bottom-end"
    trigger="click"
    :width="286"
    :teleported="false"
    :popper-options="popperOptions"
    popper-class="admin-table-settings"
    @show="appearanceOpen = false"
  >
    <template #reference>
      <el-button
        class="admin-tool-button"
        :class="{ 'is-active': columnsOpen }"
        :icon="Operation"
        title="列设置"
        aria-label="列设置"
        :aria-expanded="columnsOpen"
      />
    </template>
    <section
      class="column-settings"
      aria-label="列设置选项"
      @keydown.esc.stop="columnsOpen = false"
    >
      <header>
        <strong>列设置</strong
        ><el-button link type="primary" @click="$emit('reset-columns')">恢复默认</el-button>
      </header>
      <p>勾选显示，拖动调整顺序</p>
      <ul class="column-settings-list">
        <li
          v-for="column in columns"
          :key="column.key"
          :data-column-key="column.key"
          :class="{ 'is-dragging': dragged === column.key }"
          @dragover="dragOver($event, column)"
          @drop.prevent="drop(column)"
        >
          <button
            class="column-drag-handle"
            :draggable="movable(column)"
            :disabled="!movable(column)"
            :aria-label="`拖动排序 ${column.label}`"
            title="拖动排序，也可用上下方向键"
            @dragstart="dragStart($event, column)"
            @dragend="dragged = ''"
            @keydown.up.prevent="step(column, -1)"
            @keydown.down.prevent="step(column, 1)"
          >
            <el-icon><DCaret /></el-icon>
          </button>
          <el-checkbox
            :model-value="!hidden.includes(column.key)"
            :disabled="!canHide(column)"
            @change="$emit('visibility', column.key, $event)"
            >{{ column.label }}</el-checkbox
          >
          <span v-if="column.locked" class="column-fixed-note">固定</span>
          <span v-else-if="column.fixed === 'right'" class="column-fixed-note">右侧</span>
          <div v-if="movable(column)" class="column-move-buttons">
            <button
              :disabled="!neighbor(column.key, -1)"
              :aria-label="`上移 ${column.label}`"
              @click="step(column, -1)"
            >
              <el-icon><ArrowUp /></el-icon>
            </button>
            <button
              :disabled="!neighbor(column.key, 1)"
              :aria-label="`下移 ${column.label}`"
              @click="step(column, 1)"
            >
              <el-icon><ArrowDown /></el-icon>
            </button>
          </div>
        </li>
      </ul>
      <small>至少保留一列数据</small>
      <span class="admin-sr-only" aria-live="polite">{{ announcement }}</span>
    </section>
  </el-popover>
  <el-popover
    v-model:visible="appearanceOpen"
    placement="bottom-end"
    trigger="click"
    :width="220"
    :teleported="false"
    :popper-options="popperOptions"
    popper-class="admin-table-settings"
    @show="columnsOpen = false"
  >
    <template #reference>
      <el-button
        class="admin-tool-button"
        :class="{ 'is-active': appearanceOpen }"
        :icon="Setting"
        title="表格设置"
        aria-label="表格设置"
        :aria-expanded="appearanceOpen"
      />
    </template>
    <section
      class="appearance-settings"
      aria-label="表格外观选项"
      @keydown.esc.stop="appearanceOpen = false"
    >
      <header>
        <strong>表格设置</strong
        ><el-button link type="primary" @click="$emit('reset-appearance')">恢复默认</el-button>
      </header>
      <el-checkbox
        v-for="item in appearanceOptions"
        :key="item.key"
        :model-value="appearance[item.key]"
        @change="$emit('appearance', item.key, $event)"
        >{{ item.label }}</el-checkbox
      >
    </section>
  </el-popover>
</template>
<script setup>
import { ref } from 'vue'
import { Operation, Setting, DCaret, ArrowUp, ArrowDown } from '@element-plus/icons-vue'
import { columnGroup } from '@/utils/admin-table'
const props = defineProps({
  columns: Array,
  hidden: Array,
  appearance: Object,
  canHide: Function,
  neighbor: Function,
})
const emit = defineEmits(['visibility', 'move', 'appearance', 'reset-columns', 'reset-appearance'])
// Position against the viewport even inside a scrolling or fullscreen table panel.
const popperOptions = {
  strategy: 'fixed',
  modifiers: [
    { name: 'preventOverflow', options: { rootBoundary: 'viewport', tether: false, padding: 12 } },
  ],
}
const columnsOpen = ref(false),
  appearanceOpen = ref(false),
  dragged = ref(''),
  announcement = ref('')
const appearanceOptions = [
  { key: 'stripe', label: '斑马纹' },
  { key: 'border', label: '边框' },
  { key: 'headerBackground', label: '表头背景' },
]
function movable(column) {
  return !!(props.neighbor(column.key, -1) || props.neighbor(column.key, 1))
}
function step(column, direction) {
  const target = props.neighbor(column.key, direction)
  if (target) {
    emit('move', column.key, target)
    announcement.value = `${column.label}已${direction < 0 ? '上移' : '下移'}`
  }
}
function dragStart(event, column) {
  if (!movable(column)) {
    event.preventDefault()
    return
  }
  dragged.value = column.key
  event.dataTransfer.effectAllowed = 'move'
  event.dataTransfer.setData('text/plain', column.key)
}
function canDrop(column) {
  const from = props.columns.find((item) => item.key === dragged.value)
  return from && !column.locked && columnGroup(from) === columnGroup(column)
}
function dragOver(event, column) {
  if (canDrop(column)) {
    event.preventDefault()
    event.dataTransfer.dropEffect = 'move'
  }
}
function drop(column) {
  if (canDrop(column)) {
    emit('move', dragged.value, column.key)
    announcement.value = '列顺序已更新'
  }
  dragged.value = ''
}
</script>
