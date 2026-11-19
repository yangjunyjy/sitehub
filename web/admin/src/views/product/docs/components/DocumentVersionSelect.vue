<template>
  <el-form-item label="产品版本">
    <el-select
      :model-value="modelValue"
      aria-label="产品文档版本"
      :disabled="disabled"
      @update:model-value="$emit('update:modelValue', $event)"
    >
      <el-option
        v-for="item in versions"
        :key="item.id"
        :value="item.id"
        :label="`${item.modeLabel ? item.modeLabel + ' / ' : ''}${item.name}${item.status === 'draft' ? ' · 草稿' : ''}${item.id === defaultVersion ? ' · 默认' : ''}`"
      />
    </el-select>
  </el-form-item>
</template>
<script setup>
defineProps({ modelValue: String, versions: Array, defaultVersion: String, disabled: Boolean })
defineEmits(['update:modelValue'])
</script>
