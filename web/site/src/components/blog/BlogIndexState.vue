<script setup>
import ContentState from "@/components/product/ContentState.vue";
defineProps({
  loading: Boolean,
  error: String,
  progress: Number,
  total: Number,
});
defineEmits(["retry"]);
</script>
<template>
  <div v-if="loading" class="index-loading" role="status">
    <span></span>
    <h2>正在整理文章目录</h2>
    <p>
      {{
        total ? `已整理 ${progress} / ${total} 篇文章` : "正在读取已发布文章…"
      }}
    </p>
  </div>
  <ContentState
    v-else-if="error"
    error
    title="文章目录暂不可用"
    :message="error"
    action-label="重新加载"
    @retry="$emit('retry')"
  />
  <slot v-else />
</template>
<style scoped>
.index-loading {
  padding: 72px 24px;
  text-align: center;
  color: var(--deploy-muted);
}
h2 {
  font-size: 18px;
  color: var(--deploy-ink);
}
p {
  font-size: 13px;
}
span {
  display: block;
  width: 120px;
  height: 3px;
  border-radius: 2px;
  margin: auto;
  background: var(--deploy-soft);
  animation: site-pulse 1.4s infinite;
}
</style>
