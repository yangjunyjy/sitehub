<script setup>
import ReviewQueue from "@/components/community/ReviewQueue.vue";

const fields = [
  { key: "name", prop: "name", label: "昵称", minWidth: 110 },
  { key: "articleId", prop: "articleId", label: "文章 ID", minWidth: 110 },
  { key: "content", prop: "content", label: "评论内容", minWidth: 190 },
];
</script>
<template>
  <ReviewQueue title="评论管理" kind="comments" :fields="fields" />
</template>
