<script setup>
import ReviewQueue from "@/components/community/ReviewQueue.vue";

const fields = [
  { key: "name", prop: "name", label: "昵称", minWidth: 110 },
  { key: "content", prop: "content", label: "留言内容", minWidth: 190 },
];
</script>
<template>
  <ReviewQueue title="留言管理" kind="messages" :fields="fields" />
</template>
