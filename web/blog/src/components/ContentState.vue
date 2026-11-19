<script setup>
defineProps({
  loading: Boolean,
  error: String,
  title: { type: String, default: "这里还没有新的故事" },
  description: { type: String, default: "等下一次灵感到来，再一起慢慢阅读。" },
});
defineEmits(["retry"]);
</script>
<template>
  <div
    class="content-state"
    role="status"
    :aria-live="error ? 'assertive' : 'polite'"
  >
    <div class="content-state__mark">
      {{ loading ? "◌" : error ? "☁" : "✧" }}
    </div>
    <h3>
      {{ loading ? "正在收集云间的文字…" : error ? "暂时无法加载" : title }}
    </h3>
    <p>{{ loading ? "请稍等片刻" : error || description }}</p>
    <button
      v-if="error"
      class="text-button"
      type="button"
      @click="$emit('retry')"
    >
      重新加载 ↗
    </button>
  </div>
</template>
<style lang="scss" scoped>
.content-state {
  padding: 65px 24px;
  text-align: center;
  color: var(--muted);
  &__mark {
    font-size: 36px;
    color: var(--accent);
    margin-bottom: 15px;
  }
  h3 {
    color: var(--ink);
    font-size: 17px;
    font-weight: 500;
  }
  p {
    font-size: 13px;
    line-height: 1.9;
  }
}
</style>
