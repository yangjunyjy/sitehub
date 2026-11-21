<script setup>
defineProps({
  title: String,
  loading: Boolean,
  confirmText: { type: String, default: "保存" },
});
const open = defineModel({ type: Boolean });
const emit = defineEmits(["submit"]);
</script>
<template>
  <el-dialog
    v-model="open"
    :title="title"
    destroy-on-close
    :close-on-click-modal="false"
    :close-on-press-escape="!loading"
    :show-close="!loading"
    append-to-body
  >
    <slot />
    <template #footer>
      <el-button :disabled="loading" @click="open = false">取消</el-button>
      <el-button type="primary" :loading="loading" @click="emit('submit')">{{
        confirmText
      }}</el-button>
    </template>
  </el-dialog>
</template>
