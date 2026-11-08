<script setup>
import ProductIcon from "./ProductIcon.vue";
defineProps({
  title: String,
  message: String,
  actionLabel: String,
  error: Boolean,
  loading: Boolean,
});
defineEmits(["retry"]);
</script>
<template>
  <section
    class="content-state"
    :role="error ? 'alert' : 'status'"
    :aria-busy="loading || undefined"
  >
    <span class="state-icon"><ProductIcon name="playbook" /></span>
    <h2>{{ title }}</h2>
    <p>{{ message }}</p>
    <div class="state-actions">
      <button
        v-if="actionLabel"
        type="button"
        class="site-button site-button--primary"
        :disabled="loading"
        @click="$emit('retry')"
      >
        {{ loading ? "正在加载…" : actionLabel }}
      </button>
      <slot />
    </div>
  </section>
</template>
<style scoped>
.content-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 60px 24px;
  text-align: center;
}
.state-icon {
  display: grid;
  place-items: center;
  width: 58px;
  height: 58px;
  margin-bottom: 20px;
  border-radius: 8px;
  background: var(--deploy-soft);
  color: var(--deploy-primary);
}
.state-icon :deep(svg) {
  width: 27px;
  height: 27px;
  stroke: currentColor;
  stroke-width: 1.5;
}
.content-state h2 {
  margin: 0 0 12px;
  font-size: 22px;
}
.content-state p {
  max-width: 420px;
  margin: 0;
  color: var(--deploy-muted);
  font-size: 14px;
  line-height: 1.9;
}
.state-actions {
  display: flex;
  justify-content: center;
  flex-wrap: wrap;
  gap: 12px;
  margin-top: 24px;
}
</style>
