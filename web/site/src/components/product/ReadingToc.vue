<script setup>
import {
  computed,
  nextTick,
  onMounted,
  onBeforeUnmount,
  ref,
  watch,
} from "vue";
const props = defineProps({
  items: { type: Array, default: () => [] },
  title: { type: String, default: "本页目录" },
});
const active = ref("");
const list = ref(null);
const baseLevel = computed(() =>
  Math.min(...props.items.map((item) => item.level), 6),
);
let frame = 0,
  disposed = false;
let resizeObserver;
const scrollBehavior = () =>
  matchMedia("(prefers-reduced-motion: reduce)").matches ? "instant" : "smooth";

const revealActive = () => {
  const container = list.value;
  const button = container?.querySelector('[aria-current="location"]');
  if (!button || !container.clientHeight) return;
  const viewport = container.getBoundingClientRect();
  const item = button.getBoundingClientRect();
  const padding = 8;
  if (
    item.top >= viewport.top + padding &&
    item.bottom <= viewport.bottom - padding
  )
    return;

  // Scroll only the list: scrollIntoView would also move the article viewport.
  container.scrollTo({
    top:
      container.scrollTop +
      item.top -
      viewport.top -
      (container.clientHeight - item.height) / 2,
    behavior: scrollBehavior(),
  });
};
watch(
  active,
  () => {
    // A new chapter supersedes any ongoing animation, even if already visible.
    const container = list.value;
    container?.scrollTo({ top: container.scrollTop, behavior: "instant" });
    revealActive();
  },
  { flush: "post" },
);
const update = () => {
  frame = 0;
  const visible = props.items
    .map((item) => ({ item, element: document.getElementById(item.slug) }))
    .filter((entry) => entry.element);
  if (!visible.length) {
    active.value = "";
    return;
  }
  const passed = visible.filter(
    (entry) => entry.element.getBoundingClientRect().top <= 156,
  );
  active.value = (passed.at(-1) || visible[0]).item.slug;
};
const schedule = () => {
  if (!frame && !disposed) frame = requestAnimationFrame(update);
};
watch(
  () => props.items,
  async () => {
    await nextTick();
    schedule();
  },
  { immediate: true },
);
const navigate = (item) => {
  active.value = item.slug;
  document.getElementById(item.slug)?.scrollIntoView({
    block: "start",
    behavior: scrollBehavior(),
  });
};
onMounted(() => {
  window.addEventListener("scroll", schedule, { passive: true });
  window.addEventListener("resize", schedule);
  // Header transitions and viewport resizing can reduce the available list height.
  resizeObserver = new ResizeObserver(revealActive);
  resizeObserver.observe(list.value);
  schedule();
});
onBeforeUnmount(() => {
  disposed = true;
  resizeObserver?.disconnect();
  cancelAnimationFrame(frame);
  window.removeEventListener("scroll", schedule);
  window.removeEventListener("resize", schedule);
});
</script>
<template>
  <nav class="reading-toc" :aria-label="title">
    <p class="toc-title">
      {{ title }}<span>{{ items.length }}</span>
    </p>
    <div ref="list" class="toc-items">
      <button
        v-for="item in items"
        :key="item.slug"
        type="button"
        :class="{ active: active === item.slug }"
        :style="{ '--toc-depth': Math.min(2, item.level - baseLevel) }"
        :aria-current="active === item.slug ? 'location' : undefined"
        @click="navigate(item)"
      >
        {{ item.text }}
      </button>
    </div>
  </nav>
</template>
<style scoped>
.reading-toc {
  display: flex;
  flex-direction: column;
  min-height: 0;
  overflow: hidden;
  border-left: 1px solid var(--deploy-line);
  padding-left: 16px;
}
.toc-title {
  font-size: 12px;
  font-weight: 600;
  margin: 0 0 10px;
  color: var(--deploy-ink);
  flex-shrink: 0;
}
.toc-title span {
  display: none;
}
.toc-items {
  display: flex;
  flex-direction: column;
  min-height: 0;
  overflow-y: auto;
  overscroll-behavior-y: contain;
  scrollbar-width: thin;
  padding: 1px;
  scrollbar-gutter: stable;
}
.reading-toc button {
  flex-shrink: 0;
  width: 100%;
  min-height: 32px;
  border: 0;
  background: transparent;
  text-align: left;
  color: var(--deploy-muted);
  padding: 5px 0 5px calc(var(--toc-depth) * 12px);
  font-size: 12px;
  line-height: 1.65;
  overflow-wrap: anywhere;
}
.reading-toc button:hover,
.reading-toc button.active {
  color: var(--deploy-primary);
}
.reading-toc button.active {
  font-weight: 500;
}
</style>
