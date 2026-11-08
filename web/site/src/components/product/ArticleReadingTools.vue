<script setup>
import {
  computed,
  ref,
  watch,
  nextTick,
  onMounted,
  onBeforeUnmount,
} from "vue";
import { useReadingLibrary } from "@/composables/useReadingLibrary";
import ReadingLibrary from "./ReadingLibrary.vue";
const props = defineProps({ article: { type: Object, required: true } });
const emit = defineEmits(["font-change"]);
const { isSaved, toggle, remember } = useReadingLibrary();
const size = ref(16),
  progress = ref(0);
try {
  const value = Number(localStorage.getItem("deploy-reading-font"));
  if ([16, 18, 20].includes(value)) size.value = value;
} catch {}
const minutes = computed(() => {
  const text = props.article.content || "";
  return Math.max(
    1,
    Math.ceil(
      (text.match(/[\u3400-\u9fff]/g)?.length || 0) / 350 +
        (text.match(/[a-zA-Z0-9]+/g)?.length || 0) / 220,
    ),
  );
});
let frame,
  observer,
  disposed = false;
function update() {
  frame = undefined;
  const element = document.querySelector(".article-page .article-content");
  if (!element) return;
  const bounds = element.getBoundingClientRect(),
    start = scrollY + bounds.top,
    length = bounds.height - innerHeight + 120;
  progress.value = Math.round(
    Math.max(
      0,
      Math.min(
        100,
        length <= 0
          ? scrollY + 120 >= start
            ? 100
            : 0
          : ((scrollY + 120 - start) / length) * 100,
      ),
    ),
  );
}
function schedule() {
  if (!frame) frame = requestAnimationFrame(update);
}
function setSize(value) {
  size.value = value;
  emit("font-change", value);
  try {
    localStorage.setItem("deploy-reading-font", String(value));
  } catch {}
  nextTick(schedule);
}
watch(
  () => props.article.id,
  async () => {
    remember(props.article);
    await nextTick();
    if (disposed) return;
    observer?.disconnect();
    const element = document.querySelector(".article-page .article-content");
    if (element && observer) observer.observe(element);
    schedule();
  },
  { immediate: true },
);
onMounted(() => {
  observer = new ResizeObserver(schedule);
  const element = document.querySelector(".article-page .article-content");
  if (element) observer.observe(element);
  emit("font-change", size.value);
  window.addEventListener("scroll", schedule, { passive: true });
  window.addEventListener("resize", schedule);
  schedule();
});
onBeforeUnmount(() => {
  disposed = true;
  observer?.disconnect();
  cancelAnimationFrame(frame);
  window.removeEventListener("scroll", schedule);
  window.removeEventListener("resize", schedule);
});
</script>
<template>
  <div
    class="reading-progress"
    role="progressbar"
    aria-label="阅读进度"
    :aria-valuenow="progress"
    :aria-valuemin="0"
    :aria-valuemax="100"
  >
    <i :style="{ transform: `scaleX(${progress / 100})` }" />
  </div>
  <div class="article-reading-tools">
    <span class="reading-duration">约 {{ minutes }} 分钟阅读</span>
    <div class="reading-fonts" role="group" aria-label="正文字号">
      <button
        v-for="(value, index) in [16, 18, 20]"
        :key="value"
        type="button"
        :aria-label="['标准字号', '较大字号', '大字号'][index]"
        :aria-pressed="size === value"
        @click="setSize(value)"
        :style="{ fontSize: value + 'px' }"
      >
        A
      </button>
    </div>
    <button
      type="button"
      class="reading-save"
      :aria-pressed="isSaved(article.id)"
      @click="toggle(article)"
    >
      {{ isSaved(article.id) ? "已加入阅读列表" : "稍后阅读" }}</button
    ><ReadingLibrary />
  </div>
</template>
<style scoped>
.reading-progress {
  position: fixed;
  left: 0;
  right: 0;
  top: 0;
  height: 3px;
  z-index: 210;
  pointer-events: none;
}
.reading-progress i {
  display: block;
  width: 100%;
  height: 100%;
  background: var(--deploy-primary);
  transform-origin: left;
  transition: transform 0.1s linear;
}
.article-reading-tools {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 16px;
  margin-top: 24px;
  padding-top: 16px;
  border-top: 1px solid var(--deploy-line);
  font-size: 12px;
  color: var(--deploy-muted);
}
.reading-duration {
  margin-right: auto;
}
.reading-fonts {
  display: flex;
  gap: 2px;
}
.reading-fonts button {
  height: 32px;
  width: 30px;
  padding: 0;
  border: 0;
  background: transparent;
  color: var(--deploy-muted);
  border-radius: 4px;
}
.reading-fonts button[aria-pressed="true"] {
  background: var(--deploy-soft);
  color: var(--deploy-primary);
}
.reading-save {
  border: 0;
  background: transparent;
  color: var(--deploy-primary);
  padding: 6px 0;
  font-size: 12px;
  white-space: nowrap;
}
.article-reading-tools :deep(.library-trigger) {
  font-size: 12px;
  padding: 6px 10px;
}
@media (max-width: 500px) {
  .article-reading-tools {
    gap: 10px;
  }
  .reading-duration {
    width: 100%;
  }
}
</style>
