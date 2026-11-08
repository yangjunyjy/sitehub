<script setup>
defineProps({
  modes: { type: Array, default: () => [] },
  current: { type: Object, required: true },
  location: { type: Function, required: true },
});
</script>
<template>
  <nav class="docs-modes" aria-label="文档工作模式">
    <router-link
      v-for="mode in modes"
      :key="mode.id"
      :to="location(mode)"
      class="docs-mode"
      :class="{ active: current.id === mode.id }"
      :aria-current="current.id === mode.id ? 'page' : undefined"
      :title="mode.description"
      :aria-description="mode.description"
      ><strong>{{ mode.label }}</strong
      ><small v-if="!mode.available">待推出</small></router-link
    >
  </nav>
</template>
<style scoped>
.docs-modes {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
  padding: 4px;
  border: 1px solid var(--deploy-line);
  border-radius: 6px;
  background: var(--deploy-white);
}
.docs-mode {
  display: flex;
  align-items: center;
  justify-content: center;
  flex: 1;
  gap: 5px;
  min-height: 30px;
  padding: 4px 6px;
  border-radius: 4px;
  color: var(--deploy-muted);
  font-size: 11px;
  white-space: nowrap;
}
.docs-mode strong {
  font-weight: 500;
}
.docs-mode small {
  font-size: 9px;
  color: var(--deploy-faint);
}
.docs-mode:hover {
  color: var(--deploy-primary);
}
.docs-mode.active {
  background: var(--deploy-soft);
  color: var(--deploy-primary);
}
</style>
