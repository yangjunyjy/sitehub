<script setup>
import { computed } from "vue";
const props = defineProps({
  page: { type: Number, required: true },
  pages: { type: Number, required: true },
  label: { type: String, required: true },
  busy: Boolean,
});
defineEmits(["change"]);
const numbers = computed(() => {
  const numbers = new Set([1, props.pages]);
  const start = Math.max(1, Math.min(props.page - 1, props.pages - 2));
  for (let n = start; n <= Math.min(props.pages, start + 2); n++)
    numbers.add(n);
  return [...numbers].sort((a, b) => a - b);
});
</script>
<template>
  <nav class="blog-pages" :aria-label="label">
    <span class="page-count">第 {{ page }} / {{ pages }} 页</span>
    <div>
      <button
        type="button"
        aria-label="上一页"
        :disabled="page <= 1 || busy"
        @click="$emit('change', page - 1)"
      >
        ‹
      </button>
      <template v-for="(n, i) in numbers" :key="n">
        <span v-if="i && n - numbers[i - 1] > 1" aria-hidden="true">…</span>
        <button
          type="button"
          :aria-label="'第 ' + n + ' 页'"
          :aria-current="n === page ? 'page' : undefined"
          :disabled="busy"
          @click="n !== page && $emit('change', n)"
        >
          {{ n }}
        </button>
      </template>
      <button
        type="button"
        aria-label="下一页"
        :disabled="page >= pages || busy"
        @click="$emit('change', page + 1)"
      >
        ›
      </button>
    </div>
  </nav>
</template>
<style scoped>
.blog-pages,
.blog-pages div {
  display: flex;
  align-items: center;
  gap: 4px;
}
.blog-pages {
  justify-content: space-between;
  gap: 8px;
  flex-wrap: wrap;
  padding-top: 16px;
  color: var(--journal-muted);
  font-size: 12px;
  font-variant-numeric: tabular-nums;
}
button {
  min-width: 30px;
  height: 32px;
  padding: 0 4px;
  border: 0;
  border-radius: 3px;
  background: transparent;
  color: inherit;
}
button[aria-current] {
  background: var(--journal-ink);
  color: #fff;
}
button:not([aria-current]):enabled:hover {
  background: var(--journal-wash);
}
button:first-child,
button:last-child {
  font-size: 20px;
}
</style>
