<script setup>
import { computed, onMounted, onBeforeUnmount, ref } from "vue";
import { Marked } from "marked";
import DOMPurify from "dompurify";
import { createArticleRenderer } from "@/utils/article-markdown";
const props = defineProps({ article: { type: Object, required: true } });
const render = createArticleRenderer(Marked, DOMPurify);
const rendered = computed(() =>
  render(
    props.article.content || "",
    props.article.content_format || "markdown",
  ),
);
const notice = ref("");
const body = ref(null), progress = ref(0);
let resizeObserver, scrollFrame = 0;
function measure() {
  scrollFrame = 0;
  const end = body.value.getBoundingClientRect().bottom + window.scrollY - window.innerHeight;
  progress.value = Math.round(Math.max(0, Math.min(1, end > 0 ? window.scrollY / end : 1)) * 100);
}
function schedule() { if (!scrollFrame) scrollFrame = requestAnimationFrame(measure); }
onMounted(() => {
  resizeObserver = new ResizeObserver(schedule);
  resizeObserver.observe(body.value);
  window.addEventListener('scroll', schedule, { passive: true });
  window.addEventListener('resize', schedule);
  measure();
});
onBeforeUnmount(() => {
  resizeObserver?.disconnect();
  window.removeEventListener('scroll', schedule);
  window.removeEventListener('resize', schedule);
  cancelAnimationFrame(scrollFrame);
});
async function copyCode(event) {
  const button = event.target.closest(".code-copy-btn");
  if (!button) return;
  const code = button.closest(".code-block-wrapper")?.querySelector("code");
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
    <Teleport to="body"><div class="reading-progress" role="progressbar" aria-label="文章阅读进度" :aria-valuenow="progress" aria-valuemin="0" aria-valuemax="100"><span :style="{ transform: `scaleX(${progress/100})` }" /></div></Teleport>
    <article ref="body" class="article-body" @click="copyCode">
      <div v-if="article.summary" class="article-intro">
        {{ article.summary }}
      </div>
      <div class="markdown-body" v-html="rendered.html"></div>
      <p v-if="notice" role="status" class="copy-notice">{{ notice }}</p>
    </article>
    <aside v-if="rendered.headings.length" class="article-toc">
      <h3>文章目录</h3>
      <a
        v-for="heading in rendered.headings"
        :key="heading.slug"
        :href="'#' + heading.slug"
        :class="{ 'is-nested': heading.level > 2 }"
        >{{ heading.text }}</a
      >
    </aside>
  </div>
</template>
<style lang="scss">
.reading-progress { position: fixed; z-index: 2000; inset: 0 0 auto; height: 4px; pointer-events: none; background: #a299b84a; }
.reading-progress span { display: block; height: 100%; transform-origin: left; background: #508260; }
.article-reading {
  display: flex;
  gap: 65px;
  padding-top: 28px;
  padding-bottom: 65px;
}
.article-body {
  flex: 1;
  min-width: 0;
}
.article-intro {
  padding: 20px 25px;
  margin-bottom: 35px;
  background: #faf8fc;
  border-left: 2px solid #c8b7da;
  font-size: 13px;
  line-height: 2.2;
  color: #4e5e51;
}
.article-toc {
  position: sticky;
  top: 105px;
  width: 220px;
  flex-shrink: 0;
  align-self: flex-start;
  border-left: 1px solid var(--line);
  padding-left: 23px;
  max-height: 70vh;
  overflow: auto;
  h3 {
    font-size: 12px;
    font-weight: 500;
    margin: 0 0 21px;
    small {
      font-size: 7px;
      letter-spacing: 1.5px;
      display: block;
      color: var(--muted);
      margin-top: 7px;
    }
  }
  a {
    font-size: 11px;
    display: block;
    line-height: 1.8;
    padding: 6px 0;
    color: #546257;
    &:hover {
      color: var(--accent);
    }
    &.is-nested {
      padding-left: 12px;
      font-size: 12px;
    }
  }
}
.markdown-body {
  font-size: clamp(16px, 1.5vw, 18px);
  color: #303931;
  line-height: 1.95;
  overflow-wrap: anywhere;
  h1,
  h2,
  h3,
  h4,
  h5,
  h6 {
    font-weight: 500;
    color: #514657;
    line-height: 1.7;
    scroll-margin-top: 100px;
  }
  h1 {
    font-size: 27px;
    margin: 38px 0 22px;
  }
  h2 {
    font-size: 23px;
    margin: 35px 0 20px;
    padding-bottom: 12px;
    border-bottom: 1px solid var(--line);
  }
  h3 {
    font-size: 19px;
    margin: 30px 0 18px;
  }
  p {
    margin: 18px 0;
  }
  a {
    color: #315b40;
    border-bottom: 1px solid #d7c9e2;
  }
  img {
    display: block;
    max-width: 100%;
    height: auto;
    margin: 25px auto;
    border-radius: 5px;
  }
  blockquote {
    padding: 4px 22px;
    margin: 24px 0;
    background: #f9f7fb;
    border-left: 3px solid #d8c9e2;
    color: #4b5a50;
  }
  ul,
  ol {
    padding-left: 23px;
  }
  li {
    margin: 6px 0;
  }
  code {
    padding: 3px 6px;
    background: #f5f0f8;
    border-radius: 3px;
    color: #814b65;
    font-size: 0.88em;
  }
  pre {
    overflow: auto;
    padding: 20px 23px;
    margin: 0;
    background: #f7f7fa;
    line-height: 1.85;
    code {
      padding: 0;
      background: none;
      color: #46574b;
      font-size: 12px;
    }
  }
  .code-block-wrapper {
    margin: 25px 0;
    border: 1px solid #eeebf2;
    border-radius: 5px;
    overflow: hidden;
  }
  .code-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 8px 15px;
    background: #f1edf5;
    font-size: 10px;
    color: #5c7563;
    button {
      font-size: 10px;
      color: #5c7563;
      border: 0;
      background: none;
      padding: 4px 7px;
    }
  }
  table {
    display: block;
    overflow: auto;
    border-collapse: collapse;
    max-width: 100%;
    margin: 25px 0;
    font-size: 12px;
  }
  th,
  td {
    padding: 11px 16px;
    border: 1px solid #ece7f0;
    text-align: left;
  }
  th {
    background: #f7f4fa;
  }
  hr {
    border: 0;
    border-top: 1px solid var(--line);
    margin: 35px 0;
  }
}
.copy-notice {
  font-size: 11px;
  color: #b48e9e;
}
@media (max-width: 850px) {
  .article-toc {
    display: none;
  }
  .article-reading {
    padding-top: 35px;
  }
  .markdown-body {
    font-size: 16px;
    h2 {
      font-size: 21px;
    }
  }
}
</style>
