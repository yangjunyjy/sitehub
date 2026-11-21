<script setup>
import { computed, nextTick, onBeforeUnmount, ref, watch } from 'vue';
import { useRoute } from 'vue-router';
import { readCatalog } from '@/api/documents';
import ResourceIcon from '@shared/icons/ResourceIcon.vue';

const emit = defineEmits(['open', 'close']);
const route = useRoute();
const dialog = ref(null), input = ref(null), trigger = ref(null);
const open = ref(false), keyword = ref(''), documents = ref([]), loading = ref(false), error = ref('');
let controller, previousFocus, disposed = false, revision = 0;
const results = computed(() => {
  const value = keyword.value.trim().toLocaleLowerCase();
  return documents.value.filter(item => !value || `${item.title} ${item.summary || ''}`.toLocaleLowerCase().includes(value));
});
const location = item => ({ name: 'docs', params: { id: item.slug || item.id },
  query: { version: item.versionId, mode: item.modeId } });
async function load() {
  const current = ++revision;
  controller?.abort();
  controller = new AbortController();
  loading.value = true;
  error.value = '';
  try {
    const catalog = await readCatalog({ signal: controller.signal,
      version: route.meta.page === 'docs' ? route.query.version : undefined,
      mode: route.meta.page === 'docs' ? route.query.mode : undefined });
    if (current !== revision || !open.value) return;
    const seen = new Set();
    documents.value = (catalog.navigation?.groups || []).flatMap(group => (group.items || []).map(item => ({ ...item,
      group: group.title, versionId: catalog.navigation.versionId, modeId: catalog.modeId })))
      .filter(item => item.id != null && item.title && !seen.has(String(item.id)) && seen.add(String(item.id)));
  } catch {
    if (current === revision && open.value) error.value = '暂时无法加载文档，请重试。';
  } finally { if (current === revision) loading.value = false; }
}
async function show() {
  if (open.value || disposed) return;
  previousFocus = document.activeElement;
  open.value = true;
  emit('open');
  dialog.value.showModal();
  document.documentElement.classList.add('search-open');
  await nextTick();
  input.value?.focus();
  load();
}
function close() {
  if (!open.value) return;
  open.value = false;
  revision++;
  controller?.abort();
  dialog.value?.close();
  document.documentElement.classList.remove('search-open');
  emit('close');
  if (!disposed) previousFocus?.focus();
}
function shortcut(event) {
  if ((event.ctrlKey || event.metaKey) && event.key.toLowerCase() === 'k') { event.preventDefault(); show(); }
}
if (typeof window !== 'undefined') window.addEventListener('keydown', shortcut);
watch(() => route.fullPath, close);
onBeforeUnmount(() => { disposed = true; close(); controller?.abort(); window.removeEventListener('keydown', shortcut); });
</script>

<template>
  <button ref="trigger" class="search-trigger" type="button" aria-label="搜索产品文档" aria-keyshortcuts="Control+k Meta+k" title="搜索文档 (Ctrl / ⌘ K)" @click="show">
    <ResourceIcon name="operations/search" :size="18" />
  </button>
  <dialog ref="dialog" class="site-search" aria-labelledby="product-search-title" @cancel.prevent="close"
    @click="event => event.target === dialog && close()">
    <div class="search-dialog">
      <header><h2 id="product-search-title">搜索产品文档</h2><button type="button" aria-label="关闭搜索" @click="close">Esc</button></header>
      <label class="search-input"><ResourceIcon name="operations/search" :size="18" /><input ref="input" v-model="keyword" type="search" placeholder="搜索安装、配置与使用指南" /></label>
      <p v-if="loading" class="search-state" role="status">正在加载文档…</p>
      <div v-else-if="error" class="search-state" role="alert"><p>{{ error }}</p><button type="button" @click="load">重新加载</button></div>
      <div v-else class="search-results">
        <p class="search-summary">{{ results.length }} 篇文档</p>
        <router-link v-for="item in results" :key="item.id" :to="location(item)" class="search-result" @click="close">
          <small>{{ item.group }}</small><strong>{{ item.title }}</strong><span v-if="item.summary">{{ item.summary }}</span>
        </router-link>
        <p v-if="!results.length" class="search-state">没有找到相关文档，试试其他关键词。</p>
      </div>
      <footer><span>产品文档</span><span>按 Esc 关闭</span></footer>
    </div>
  </dialog>
</template>

<style lang="scss" scoped>
.search-trigger {
  display: grid; place-items: center; width: 36px; height: 36px; padding: 0;
  border: 0; border-radius: 50%; color: var(--deploy-muted); background: transparent;
  transition: color .16s, background .16s;
  &:hover, &:focus-visible { color: var(--deploy-ink); background: var(--deploy-paper); }
}
.site-search { width: min(680px, calc(100vw - 32px)); max-height: min(720px, calc(100vh - 80px)); margin: 64px auto auto; padding: 0; border: 1px solid var(--deploy-line); border-radius: 12px; background: var(--deploy-white); color: var(--deploy-ink); box-shadow: 0 20px 80px #0003;
  &::backdrop { background: #0006; backdrop-filter: blur(3px); }
}
.search-dialog { display: flex; flex-direction: column; max-height: inherit;
  > header, > footer { display: flex; align-items: center; justify-content: space-between; padding: 16px 22px; }
  h2 { font-size: 16px; margin: 0; } button { background: transparent; color: inherit; border: 1px solid var(--deploy-line); border-radius: 4px; padding: 3px 7px; }
  > footer { border-top: 1px solid var(--deploy-line); color: var(--deploy-muted); font-size: 12px; }
}
.search-input { display: flex; align-items: center; gap: 10px; margin: 0 22px 12px; padding: 12px 14px; border: 1px solid var(--deploy-line); border-radius: 6px;
  :deep(svg) { width: 20px; height: 20px; } input { width: 100%; color: inherit; background: transparent; border: none; outline: none; font: inherit; }
}
.search-results { overflow-y: auto; padding: 0 14px 16px; }.search-summary { color: var(--deploy-muted); font-size: 12px; padding: 0 10px; }
.search-result { display: flex; flex-direction: column; gap: 4px; padding: 13px 12px; border-radius: 6px; text-decoration: none; color: inherit;
  &:hover, &:focus-visible { background: var(--deploy-paper); } small, span { color: var(--deploy-muted); font-size: 12px; } strong { font-size: 14px; }
}
.search-state { text-align: center; padding: 32px; color: var(--deploy-muted); }
@media (max-width: 760px) { .site-search { margin-top: 24px; } }
</style>
