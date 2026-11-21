<script setup>
import * as icons from "@element-plus/icons-vue";
defineProps({ name: String, size: Number });
</script>
<template>
  <component
    :is="icons[name] || icons.Document"
    :width="size"
    :height="size"
    aria-hidden="true"
  />
</template>
