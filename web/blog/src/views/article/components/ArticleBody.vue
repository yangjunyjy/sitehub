<script setup>
import { computed, onMounted, onBeforeUnmount, ref } from "vue";
import ArticleOutline from "./ArticleOutline.vue";
import { Marked } from "marked";
import DOMPurify from "dompurify";
import { createArticleRenderer } from "@/utils/article-markdown";
const props = defineProps({ article: { type: Object, required: true } });
const render = createArticleRenderer(Marked, DOMPurify);
const rendered = computed(() =>
  render(
    props.article.content || "",
    props.article.content_format || props.article.contentFormat || "markdown",
  ),
);
const notice = ref("");
const body = ref(null),
  progress = ref(0);
const activeHeading = ref("");
let resizeObserver,
  scrollFrame = 0,
  offsets = [],
  articleEnd = 0;
function measure() {
  scrollFrame = 0;
  const y = Math.max(0, window.scrollY);
  progress.value = Math.round(
    Math.max(0, Math.min(1, articleEnd > 0 ? y / articleEnd : 1)) * 100,
  );
  let left = 0,
    right = offsets.length;
  while (left < right) {
    const middle = Math.floor((left + right) / 2);
    if (offsets[middle].top <= y + 110) left = middle + 1;
    else right = middle;
  }
  activeHeading.value = offsets[Math.max(0, left - 1)]?.id || "";
}
function positions() {
  if (!body.value) return;
  articleEnd =
    body.value.getBoundingClientRect().bottom + window.scrollY - innerHeight;
  offsets = Array.from(
    body.value.querySelectorAll("h1[id],h2[id],h3[id],h4[id],h5[id],h6[id]"),
    (heading) => ({
      id: heading.id,
      top: heading.getBoundingClientRect().top + window.scrollY,
    }),
  );
  measure();
}
function schedule() {
  if (!scrollFrame) scrollFrame = requestAnimationFrame(measure);
}
onMounted(() => {
  resizeObserver = new ResizeObserver(positions);
  resizeObserver.observe(body.value);
  window.addEventListener("scroll", schedule, { passive: true });
  window.addEventListener("resize", positions);
  positions();
  try {
    const anchor = document.getElementById(
      decodeURIComponent(location.hash.slice(1)),
    );
    if (anchor && body.value.contains(anchor)) anchor.scrollIntoView();
  } catch {
    /* Ignore malformed fragment identifiers in shared links. */
  }
});
onBeforeUnmount(() => {
  resizeObserver?.disconnect();
  window.removeEventListener("scroll", schedule);
  window.removeEventListener("resize", positions);
  cancelAnimationFrame(scrollFrame);
});
async function copyCode(event) {
  const button = event.target.closest(".code-copy-btn");
  if (!button) return;
  const code = button.closest(".code-block-wrapper")?.querySelector("pre");
  if (!code) return;
  try {
    await navigator.clipboard.writeText(code.textContent);
    button.textContent = "已复制";
  } catch {
    notice.value = "复制未完成，请选择代码手动复制";
  }
}
</script>
<template>
  <div class="article-reading">
    <Teleport to="body"
      ><div
        class="reading-progress"
        role="progressbar"
        aria-label="文章阅读进度"
        :aria-valuenow="progress"
        aria-valuemin="0"
        aria-valuemax="100"
      >
        <span :style="{ transform: `scaleX(${progress / 100})` }" /></div
    ></Teleport>
    <ArticleOutline
      v-if="rendered.headings.length"
      :headings="rendered.headings"
      :active="activeHeading"
    />
    <article ref="body" class="article-body" @click="copyCode">
      <div class="markdown-body" v-html="rendered.html"></div>
      <p v-if="notice" role="status" class="copy-notice">{{ notice }}</p>
    </article>
  </div>
</template>
<style lang="scss" src="../typography.scss"></style>
