<template>
  <section class="admin-search-card" aria-label="查询条件">
    <el-form
      :model="model"
      class="admin-search-form"
      label-position="left"
      @submit.prevent="submit"
      @keydown.enter.capture="guardComposition"
    >
      <div class="admin-search-fields"><slot /></div>
      <div
        v-if="$slots.advanced"
        v-show="expanded"
        :id="advancedId"
        class="admin-search-fields admin-search-advanced"
      >
        <slot name="advanced" />
      </div>
      <div class="admin-search-actions">
        <el-button :disabled="loading" @click="$emit('reset')">重置</el-button>
        <el-button type="primary" native-type="submit" :loading="loading">查询</el-button>
        <el-button
          v-if="$slots.advanced"
          link
          type="primary"
          :aria-expanded="expanded"
          :aria-controls="advancedId"
          @click="expanded = !expanded"
        >
          {{ expanded ? '收起' : '展开'
          }}<el-icon><ArrowUp v-if="expanded" /><ArrowDown v-else /></el-icon>
        </el-button>
      </div>
    </el-form>
  </section>
</template>
<script setup>
import { ref, useId } from 'vue'
import { ArrowDown, ArrowUp } from '@element-plus/icons-vue'
const props = defineProps({ model: Object, loading: Boolean })
const emit = defineEmits(['search', 'reset'])
const expanded = ref(false)
const advancedId = `admin-search-${useId()}`
function submit() {
  if (!props.loading) emit('search')
}
function guardComposition(event) {
  if (event.isComposing || event.keyCode === 229) event.preventDefault()
}
</script>
