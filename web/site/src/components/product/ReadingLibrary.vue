<script setup>
import { ref, computed, watch, onBeforeUnmount } from "vue";
import { useRoute } from "vue-router";
import { useReadingLibrary } from "@/composables/useReadingLibrary";
const { saved, history, persistent, remove, clearHistory } =
  useReadingLibrary();
const dialog = ref(),
  tab = ref("saved"),
  opened = ref(false),
  route = useRoute();
const items = computed(() =>
  tab.value === "saved" ? saved.value : history.value,
);
function open() {
  dialog.value.showModal();
  opened.value = true;
}
function close() {
  dialog.value?.close();
  opened.value = false;
}
watch(() => route.fullPath, close);
watch(opened, (value) =>
  document.documentElement.classList.toggle("reading-library-open", value),
);
onBeforeUnmount(() =>
  document.documentElement.classList.remove("reading-library-open"),
);
</script>
<template>
  <button type="button" class="library-trigger" @click="open">
    <svg viewBox="0 0 24 24" aria-hidden="true">
      <path d="M6 4h12v17l-6-4-6 4Z" /></svg
    >阅读列表<span v-if="saved.length">{{ saved.length }}</span>
  </button>
  <dialog
    ref="dialog"
    class="reading-library"
    aria-labelledby="reading-library-title"
    @close="opened = false"
    @click="
      (event) => {
        if (event.target === dialog) close();
      }
    "
  >
    <header>
      <div>
        <h2 id="reading-library-title">你的阅读空间</h2>
        <p>
          {{
            persistent
              ? "保存在此浏览器，随时接着读。"
              : "浏览器存储不可用，本次阅读期间有效。"
          }}
        </p>
      </div>
      <button type="button" aria-label="关闭阅读列表" @click="close">×</button>
    </header>
    <nav aria-label="阅读记录类型">
      <button
        type="button"
        :aria-pressed="tab === 'saved'"
        @click="tab = 'saved'"
      >
        稍后阅读 · {{ saved.length }}</button
      ><button
        type="button"
        :aria-pressed="tab === 'history'"
        @click="tab = 'history'"
      >
        最近阅读 · {{ history.length }}
      </button>
    </nav>
    <div class="library-items">
      <article v-for="item in items" :key="item.id">
        <div>
          <router-link :to="`/article/${item.id}`">{{
            item.title
          }}</router-link>
          <p v-if="item.summary">{{ item.summary }}</p>
        </div>
        <button
          v-if="tab === 'saved'"
          type="button"
          :aria-label="`移除 ${item.title}`"
          @click="remove(item.id)"
        >
          移除
        </button>
      </article>
      <div v-if="!items.length" class="library-empty">
        <strong>{{
          tab === "saved" ? "还没有保存文章" : "还没有阅读记录"
        }}</strong>
        <p>
          {{
            tab === "saved"
              ? "点击文章旁的“稍后阅读”，收集想看的内容。"
              : "打开文章后，会在这里留下最近阅读的记录。"
          }}
        </p>
      </div>
    </div>
    <footer>
      <span>稍后阅读最多保存 50 篇，最近阅读保留 12 篇。</span
      ><button
        v-if="tab === 'history' && history.length"
        type="button"
        @click="clearHistory"
      >
        清空阅读记录
      </button>
    </footer>
  </dialog>
</template>
<style scoped>
.library-trigger {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 9px 14px;
  border: 1px solid var(--deploy-line);
  background: var(--deploy-white);
  color: var(--deploy-muted);
  border-radius: 6px;
  font-size: 13px;
  white-space: nowrap;
}
.library-trigger svg {
  width: 15px;
  height: 15px;
  fill: none;
  stroke: currentColor;
  stroke-width: 1.5;
}
.library-trigger > span {
  color: var(--deploy-primary);
}
.reading-library {
  width: min(640px, calc(100% - 32px));
  max-height: 80dvh;
  border: 1px solid var(--deploy-line);
  border-radius: 12px;
  background: var(--deploy-white);
  color: var(--deploy-ink);
  padding: 0;
  box-shadow: 0 16px 60px #0003;
}
.reading-library::backdrop {
  background: #0006;
}
.reading-library header {
  display: flex;
  justify-content: space-between;
  align-items: start;
  padding: 24px 24px 16px;
  gap: 16px;
}
.reading-library h2 {
  font-size: 22px;
  margin: 0 0 8px;
}
.reading-library header p {
  font-size: 12px;
  color: var(--deploy-muted);
  margin: 0;
}
.reading-library header button {
  border: 0;
  background: transparent;
  color: var(--deploy-muted);
  font-size: 24px;
  line-height: 1;
  padding: 6px;
}
.reading-library nav {
  display: flex;
  gap: 20px;
  padding: 0 24px;
  border-bottom: 1px solid var(--deploy-line);
}
.reading-library nav button {
  background: transparent;
  color: var(--deploy-muted);
  border: 0;
  border-bottom: 2px solid transparent;
  font-size: 13px;
  padding: 12px 0;
}
.reading-library nav button[aria-pressed="true"] {
  color: var(--deploy-primary);
  border-bottom-color: var(--deploy-primary);
}
.library-items {
  max-height: 48dvh;
  overflow: auto;
  padding: 0 24px;
}
.library-items article {
  display: flex;
  gap: 16px;
  justify-content: space-between;
  padding: 20px 0;
  border-bottom: 1px solid var(--deploy-line);
}
.library-items article > div {
  min-width: 0;
}
.library-items a {
  font-size: 15px;
  font-weight: 500;
  overflow-wrap: anywhere;
}
.library-items a:hover {
  color: var(--deploy-primary);
}
.library-items p {
  font-size: 12px;
  color: var(--deploy-muted);
  margin: 8px 0 0;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}
.library-items button,
.reading-library footer button {
  border: 0;
  background: transparent;
  color: var(--deploy-primary);
  white-space: nowrap;
  font-size: 12px;
  align-self: start;
}
.library-empty {
  padding: 48px 0;
  text-align: center;
}
.library-empty strong {
  font-size: 15px;
  font-weight: 500;
}
.reading-library footer {
  display: flex;
  justify-content: space-between;
  gap: 12px;
  padding: 16px 24px;
  font-size: 11px;
  color: var(--deploy-muted);
}
@media (max-width: 500px) {
  .reading-library footer {
    flex-direction: column;
  }
  .reading-library header,
  .library-items {
    padding-inline: 18px;
  }
}
</style>
<style>
html.reading-library-open {
  overflow: hidden;
}
</style>
