<template>
  <el-tooltip :content="reason || label" placement="top">
    <el-button
      :aria-label="label"
      :aria-disabled="!!reason || busy"
      :disabled="busy"
      :icon="Delete"
      :class="{ unavailable: reason }"
      text
      type="danger"
      @click="activate"
    >
      <slot />
    </el-button>
  </el-tooltip>
</template>
<script setup>
import { Delete } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'

const props = defineProps({ label: String, reason: String, busy: Boolean })
const emit = defineEmits(['remove'])
function activate() {
  if (props.reason) ElMessage.info(props.reason)
  else emit('remove')
}
</script>
<style scoped lang="scss">
.unavailable {
  color: var(--el-text-color-secondary);
}
</style>
