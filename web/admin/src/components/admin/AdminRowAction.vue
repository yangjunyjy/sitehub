<template>
  <button
    ref="button"
    v-bind="$attrs"
    type="button"
    class="admin-row-action"
    :class="`admin-row-action--${definition.tone}`"
    :aria-label="resolvedLabel"
    :aria-busy="loading || undefined"
    :title="action === 'more' ? resolvedLabel : undefined"
    :disabled="disabled || loading"
    @click="$emit('click', $event)"
  >
    <el-icon :class="{ 'is-loading': loading }" aria-hidden="true">
      <component :is="loading ? Loading : definition.icon" />
    </el-icon>
    <el-tooltip
      v-if="action !== 'more'"
      virtual-triggering
      :virtual-ref="button"
      :content="resolvedLabel"
      :trigger="['hover', 'focus']"
      :trigger-keys="[]"
      :show-after="250"
      :popper-options="tooltipOptions"
      placement="top"
    />
  </button>
</template>

<script setup>
import { computed, ref } from 'vue'
import {
  Edit,
  CopyDocument,
  Delete,
  View,
  Download,
  User,
  Key,
  ArrowUp,
  ArrowDown,
  Upload,
  SoldOut,
  RefreshLeft,
  MoreFilled,
  Loading,
} from '@element-plus/icons-vue'

defineOptions({ inheritAttrs: false })
const props = defineProps({
  action: { type: String, required: true },
  label: String,
  disabled: Boolean,
  loading: Boolean,
})
defineEmits(['click'])
// Keep a native root so dropdowns can bind their own trigger and keyboard events.
const button = ref()
// A focused action may move outside the horizontal table viewport after resize.
// Keep its teleported tooltip inside the screen instead of widening the page.
const tooltipOptions = {
  modifiers: [{ name: 'preventOverflow', options: { padding: 8, altAxis: true, tether: false } }],
}
const actions = {
  copy: { icon: CopyDocument, label: '复制', tone: 'neutral' },
  edit: { icon: Edit, label: '编辑', tone: 'edit' },
  rename: { icon: Edit, label: '重命名', tone: 'edit' },
  delete: { icon: Delete, label: '删除', tone: 'danger' },
  view: { icon: View, label: '查看', tone: 'neutral' },
  export: { icon: Download, label: '导出', tone: 'neutral' },
  assignRoles: { icon: User, label: '分配角色', tone: 'access' },
  assignPermissions: { icon: Key, label: '分配权限', tone: 'access' },
  moveUp: { icon: ArrowUp, label: '上移', tone: 'neutral' },
  moveDown: { icon: ArrowDown, label: '下移', tone: 'neutral' },
  publish: { icon: Upload, label: '发布', tone: 'success' },
  unpublish: { icon: SoldOut, label: '下架', tone: 'warning' },
  restore: { icon: RefreshLeft, label: '恢复', tone: 'success' },
  more: { icon: MoreFilled, label: '更多操作', tone: 'neutral' },
}
const definition = computed(() => actions[props.action] || actions.more)
const resolvedLabel = computed(() => props.label || definition.value.label)
</script>

<style scoped>
.admin-row-action {
  --action-color: #409eff;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  flex: 0 0 auto;
  width: 30px;
  height: 30px;
  padding: 0;
  margin: 0;
  border: 1px solid transparent;
  border-radius: 4px;
  color: var(--action-color);
  background: transparent;
  font-size: 16px;
  cursor: pointer;
  transition: background 0.15s;
}
.admin-row-action--danger {
  --action-color: #f56c6c;
}
.admin-row-action--neutral {
  --action-color: #909399;
}
.admin-row-action--access {
  --action-color: #409eff;
}
.admin-row-action--success {
  --action-color: #67c23a;
}
.admin-row-action--warning {
  --action-color: #e6a23c;
}
.admin-row-action:hover:not(:disabled) {
  background: #f5f7fa;
}
.admin-row-action:focus-visible {
  outline: 2px solid var(--action-color);
  outline-offset: 1px;
}
.admin-row-action:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}
.admin-table-panel[data-density='compact'] .admin-row-action {
  width: 28px;
  height: 28px;
  font-size: 15px;
}
@media (pointer: coarse) {
  .admin-row-action,
  .admin-table-panel[data-density='compact'] .admin-row-action {
    width: 40px;
    height: 40px;
  }
}
</style>
