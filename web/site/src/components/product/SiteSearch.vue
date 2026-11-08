<script setup>
import {
  computed,
  nextTick,
  onBeforeUnmount,
  onMounted,
  ref,
  watch,
} from "vue";
import { useRoute } from "vue-router";
import { getArticles, getDocsNavigation } from "@/api";
import ProductIcon from "./ProductIcon.vue";

const emit = defineEmits(["open", "close"]);
const route = useRoute();
const trigger = ref(null),
  dialog = ref(null),
  input = ref(null),
  open = ref(false);
const keyword = ref(""),
  scope = ref("all"),
  composing = ref(false);
const documents = ref([]),
  articles = ref([]),
  articleTotal = ref(0);
const docsLoading = ref(false),
  articlesLoading = ref(false);
const docsError = ref(false),
  articlesError = ref(false),
  documentLimit = ref(8);
const query = computed(() => keyword.value.trim().slice(0, 200));
const matchingDocuments = computed(() => {
  const value = query.value.toLocaleLowerCase();
  return documents.value.filter(
    (item) =>
      !value ||
      `${item.title} ${item.summary || ""}`.toLocaleLowerCase().includes(value),
  );
});
const visibleDocuments = computed(() =>
  matchingDocuments.value.slice(0, query.value ? documentLimit.value : 4),
);
const visibleArticles = computed(() =>
  query.value ? articles.value : articles.value.slice(0, 4),
);
const articleLocation = computed(() => ({
  path: "/blog",
  query: query.value ? { q: query.value } : { sort: "hot" },
}));
let timer,
  articleController,
  docsController,
  articleRequest = 0,
  docsRequest = 0;
let previousFocus,
  backdropPointer = false,
  disposed = false;

// Each source has its own status: an unavailable article API must not hide matching documents.
const loadDocuments = async () => {
  const current = ++docsRequest;
  docsController?.abort();
  docsController = new AbortController();
  docsLoading.value = true;
  docsError.value = false;
  try {
    const response = await getDocsNavigation({
      signal: docsController.signal,
      version: route.path.startsWith("/docs") ? route.query.version : undefined,
      mode: route.path.startsWith("/docs") ? route.query.mode : undefined,
    });
    if (current !== docsRequest || !open.value) return;
    if (!Array.isArray(response.data?.groups))
      throw new Error("Invalid navigation response");
    const seen = new Set();
    documents.value = response.data.groups
      .flatMap((group) =>
        (group.items || []).map((item) => ({
          ...item,
          group: group.title,
          versionId: response.data.versionId,
          modeId:
            response.data.modeId ||
            (route.path.startsWith("/docs") ? route.query.mode : undefined),
        })),
      )
      .filter((item) => {
        if (item.id == null || !item.title || seen.has(String(item.id)))
          return false;
        seen.add(String(item.id));
        return true;
      });
  } catch {
    if (current === docsRequest && open.value) {
      documents.value = [];
      docsError.value = true;
    }
  } finally {
    if (current === docsRequest) docsLoading.value = false;
  }
};
const loadArticles = async () => {
  clearTimeout(timer);
  const current = ++articleRequest;
  articleController?.abort();
  articleController = new AbortController();
  articlesLoading.value = true;
  articlesError.value = false;
  articles.value = [];
  articleTotal.value = 0;
  try {
    const response = await getArticles(
      {
        page: 1,
        pageSize: 6,
        sort: query.value ? "latest" : "hot",
        ...(query.value ? { keyword: query.value } : {}),
      },
      { signal: articleController.signal },
    );
    if (current !== articleRequest || !open.value) return;
    if (!Array.isArray(response.data?.items))
      throw new Error("Invalid article response");
    articles.value = response.data.items;
    articleTotal.value = Math.max(
      0,
      Number(response.data.total) || articles.value.length,
    );
  } catch {
    if (current === articleRequest && open.value) articlesError.value = true;
  } finally {
    if (current === articleRequest) articlesLoading.value = false;
  }
};
const scheduleSearch = () => {
  clearTimeout(timer);
  articleRequest++;
  articleController?.abort();
  articles.value = [];
  articleTotal.value = 0;
  articlesError.value = false;
  documentLimit.value = 8;
  if (!open.value) return;
  articlesLoading.value = true;
  if (!composing.value) timer = setTimeout(loadArticles, 250);
};
watch(query, scheduleSearch);
const finishComposition = () => {
  composing.value = false;
  scheduleSearch();
};
const submit = () => {
  if (!composing.value) loadArticles();
};
const openSearch = async () => {
  if (open.value || disposed) return;
  previousFocus = document.activeElement;
  emit("open");
  open.value = true;
  scope.value = "all";
  document.documentElement.classList.add("search-open");
  dialog.value.showModal();
  await nextTick();
  if (!open.value || disposed) return;
  input.value?.focus();
  loadDocuments();
  loadArticles();
};
const stopRequests = () => {
  clearTimeout(timer);
  articleRequest++;
  docsRequest++;
  articleController?.abort();
  docsController?.abort();
};
const closeSearch = (restoreFocus = true) => {
  if (!open.value) return;
  open.value = false;
  emit("close");
  composing.value = false;
  stopRequests();
  document.documentElement.classList.remove("search-open");
  dialog.value?.close();
  if (restoreFocus && previousFocus?.isConnected) previousFocus.focus();
};
const shortcut = (event) => {
  if (
    (event.ctrlKey || event.metaKey) &&
    event.key.toLowerCase() === "k" &&
    !event.isComposing
  ) {
    event.preventDefault();
    if (open.value) input.value?.focus();
    else openSearch();
  }
};
const isOutside = (event) => {
  const rect = dialog.value.getBoundingClientRect();
  return (
    event.target === dialog.value &&
    (event.clientY < rect.top ||
      event.clientY > rect.bottom ||
      event.clientX < rect.left ||
      event.clientX > rect.right)
  );
};
const closeFromBackdrop = (event) => {
  if (backdropPointer && isOutside(event)) closeSearch();
  backdropPointer = false;
};
const navigateResults = (event) => {
  if (event.key === "Tab") {
    const elements = [
      ...dialog.value.querySelectorAll(
        "a[href],button:not(:disabled),input:not(:disabled)",
      ),
    ].filter((element) => element.getClientRects().length);
    const first = elements[0],
      last = elements.at(-1);
    if (event.shiftKey && document.activeElement === first) {
      event.preventDefault();
      last?.focus();
    } else if (!event.shiftKey && document.activeElement === last) {
      event.preventDefault();
      first?.focus();
    }
    return;
  }
  if (
    event.isComposing ||
    composing.value ||
    !["ArrowDown", "ArrowUp"].includes(event.key)
  )
    return;
  const links = [...dialog.value.querySelectorAll(".search-result-link")];
  const position = links.indexOf(document.activeElement);
  if (document.activeElement !== input.value && position === -1) return;
  event.preventDefault();
  if (event.key === "ArrowDown") (links[position + 1] || links[0])?.focus();
  else if (position <= 0) input.value?.focus();
  else links[position - 1]?.focus();
};
watch(
  () => route.fullPath,
  () => closeSearch(false),
);
onMounted(() => window.addEventListener("keydown", shortcut));
onBeforeUnmount(() => {
  disposed = true;
  closeSearch(false);
  stopRequests();
  window.removeEventListener("keydown", shortcut);
});
</script>

<template>
  <button
    ref="trigger"
    class="site-search-trigger"
    type="button"
    aria-label="搜索文档和文章"
    aria-haspopup="dialog"
    :aria-expanded="open"
    aria-controls="site-search-dialog"
    @click="openSearch"
  >
    <svg viewBox="0 0 24 24" aria-hidden="true">
      <circle cx="10.5" cy="10.5" r="6.5" />
      <path d="m16 16 4 4" />
    </svg>
    <span>搜索</span><kbd>Ctrl K</kbd>
  </button>
  <Teleport to="body">
    <dialog
      id="site-search-dialog"
      ref="dialog"
      class="site-search-dialog"
      aria-labelledby="site-search-title"
      @cancel.prevent="!composing && closeSearch()"
      @pointerdown="backdropPointer = isOutside($event)"
      @click="closeFromBackdrop"
      @keydown="navigateResults"
    >
      <div class="site-search-shell">
        <div class="site-search-heading">
          <h2 id="site-search-title">站内搜索</h2>
          <button
            class="search-close"
            type="button"
            aria-label="关闭搜索"
            @click="closeSearch()"
          >
            <span>Esc</span
            ><svg viewBox="0 0 24 24" aria-hidden="true">
              <path d="m6 6 12 12M6 18 18 6" />
            </svg>
          </button>
        </div>
        <form class="site-search-form" role="search" @submit.prevent="submit">
          <svg viewBox="0 0 24 24" aria-hidden="true">
            <circle cx="10.5" cy="10.5" r="6.5" />
            <path d="m16 16 4 4" />
          </svg>
          <input
            ref="input"
            v-model="keyword"
            type="search"
            aria-label="站内搜索关键词"
            placeholder="搜索文档、文章…"
            maxlength="200"
            autocomplete="off"
            autofocus
            @compositionstart="
              composing = true;
              scheduleSearch();
            "
            @compositionend="finishComposition"
          />
          <button
            v-if="keyword"
            class="search-clear"
            type="button"
            aria-label="清空搜索关键词"
            @click="
              keyword = '';
              input?.focus();
            "
          >
            ×</button
          ><button class="search-submit" type="submit">搜索</button>
        </form>
        <div class="search-controls">
          <div class="search-scopes" role="group" aria-label="搜索范围">
            <button
              v-for="item in [
                { value: 'all', label: '全部' },
                { value: 'docs', label: '文档' },
                { value: 'articles', label: '文章' },
              ]"
              :key="item.value"
              type="button"
              :aria-pressed="scope === item.value"
              :class="{ active: scope === item.value }"
              @click="scope = item.value"
            >
              {{ item.label }}
            </button>
          </div>
          <span>{{
            query ? `“${query}” 的搜索结果` : "发现文档与产品实践"
          }}</span>
        </div>
        <div class="search-result-viewport">
          <div class="search-columns" :class="{ single: scope !== 'all' }">
            <section
              v-if="scope !== 'articles'"
              class="search-section"
              aria-label="文档搜索结果"
              :aria-busy="docsLoading"
            >
              <div class="search-section-heading" aria-live="polite">
                <h3>{{ query ? "相关文档" : "文档速查" }}</h3>
                <span v-if="!docsLoading && !docsError">{{
                  query ? `${matchingDocuments.length} 篇` : "快速上手"
                }}</span>
              </div>
              <div
                v-if="docsLoading"
                class="search-skeleton"
                role="status"
                aria-label="正在加载文档"
              >
                <i /><i /><i />
              </div>
              <div v-else-if="docsError" class="search-feedback" role="status">
                <p>文档暂时无法加载</p>
                <button type="button" @click="loadDocuments">
                  重试文档搜索
                </button>
              </div>
              <template v-else-if="visibleDocuments.length"
                ><router-link
                  v-for="item in visibleDocuments"
                  :key="item.id"
                  class="search-result-link"
                  :to="{
                    path: `/docs/${encodeURIComponent(item.slug || item.id)}`,
                    query: { version: item.versionId, mode: item.modeId },
                  }"
                  @click="closeSearch(false)"
                  ><span class="search-result-icon"
                    ><ProductIcon name="playbook" /></span
                  ><span class="search-result-copy"
                    ><strong>{{ item.title }}</strong
                    ><small>{{ item.summary || item.group }}</small></span
                  ></router-link
                ><button
                  v-if="
                    query && matchingDocuments.length > visibleDocuments.length
                  "
                  class="search-more"
                  type="button"
                  @click="documentLimit += 10"
                >
                  显示更多文档</button
                ><router-link
                  v-else-if="!query"
                  class="search-more"
                  to="/docs"
                  @click="closeSearch(false)"
                  >进入文档中心
                </router-link></template
              >
              <div v-else class="search-feedback" role="status">
                <p>{{ query ? "没有匹配的文档" : "文档正在准备中" }}</p>
                <span>{{
                  query
                    ? "试试文档标题或简介中的关键词"
                    : "发布后的文档会显示在这里"
                }}</span>
              </div>
            </section>
            <section
              v-if="scope !== 'docs'"
              class="search-section"
              aria-label="文章搜索结果"
              :aria-busy="articlesLoading"
            >
              <div class="search-section-heading" aria-live="polite">
                <h3>{{ query ? "相关文章" : "推荐阅读" }}</h3>
                <span v-if="!articlesLoading && !articlesError">{{
                  query ? `${articleTotal} 篇` : "按阅读量排序"
                }}</span>
              </div>
              <div
                v-if="articlesLoading"
                class="search-skeleton"
                role="status"
                aria-label="正在搜索文章"
              >
                <i /><i /><i />
              </div>
              <div
                v-else-if="articlesError"
                class="search-feedback"
                role="status"
              >
                <p>文章暂时无法加载</p>
                <button type="button" @click="loadArticles">
                  重试文章搜索
                </button>
              </div>
              <template v-else-if="articles.length"
                ><router-link
                  v-for="item in visibleArticles"
                  :key="item.id"
                  class="search-result-link"
                  :to="`/article/${item.id}`"
                  @click="closeSearch(false)"
                  ><span class="search-result-icon article-icon"
                    ><ProductIcon name="playbook" /></span
                  ><span class="search-result-copy"
                    ><strong>{{ item.title }}</strong
                    ><small>{{
                      item.summary || item.category?.name || "产品博客"
                    }}</small></span
                  ></router-link
                ><router-link
                  class="search-more"
                  :to="articleLocation"
                  @click="closeSearch(false)"
                  >{{
                    query ? `查看全部 ${articleTotal} 篇文章` : "进入产品博客"
                  }}
                </router-link></template
              >
              <div v-else class="search-feedback" role="status">
                <p>{{ query ? "没有匹配的文章" : "文章正在准备中" }}</p>
                <span>{{
                  query
                    ? "换个标题或正文关键词试试"
                    : "发布后的文章会显示在这里"
                }}</span>
              </div>
            </section>
          </div>
        </div>
        <!-- <footer class="search-footer">
          <span>搜索文档标题、简介与文章标题、正文</span
          ><span class="search-keyboard-hint"
            ><kbd>方向键</kbd> 切换 <kbd>Enter</kbd> 打开
            <kbd>Esc</kbd> 关闭</span
          >
        </footer> -->
      </div>
    </dialog>
  </Teleport>
</template>

<style>
html.search-open {
  overflow: hidden;
}
.site-search-trigger {
  display: flex;
  align-items: center;
  gap: 9px;
  width: 154px;
  min-height: 40px;
  padding: 8px 12px;
  border: 1px solid var(--deploy-line);
  border-radius: 9px;
  background: var(--deploy-paper);
  color: var(--deploy-muted);
  text-align: left;
}
.site-search-trigger:hover,
.site-search-trigger[aria-expanded="true"] {
  border-color: var(--deploy-line);
  background: var(--deploy-paper);
  color: var(--deploy-primary);
}
.site-search-trigger svg,
.site-search-form > svg,
.search-close svg {
  flex-shrink: 0;
  width: 18px;
  height: 18px;
  fill: none;
  stroke: currentColor;
  stroke-width: 1.7;
}
.site-search-trigger > span {
  font-size: 12px;
  white-space: nowrap;
}
.site-search-trigger kbd {
  margin-left: auto;
  padding: 1px 4px;
  border: 1px solid var(--deploy-line);
  border-radius: 4px;
  background: var(--deploy-white);
  font-family: inherit;
  font-size: 10px;
  white-space: nowrap;
}
.site-search-dialog {
  position: fixed;
  inset: 80px 0 auto;
  width: min(760px, calc(100% - 40px));
  max-width: 760px;
  max-height: calc(100dvh - 110px);
  margin: 0 auto;
  padding: 0;
  border: 0;
  border-top: 1px solid var(--deploy-line);
  border-radius: 10px;
  background: var(--deploy-white);
  color: var(--deploy-ink);
  box-shadow: 0 22px 50px #00000012;
}
.site-search-dialog::backdrop {
  background: #0006;
}
.site-search-shell {
  width: calc(100% - 40px);
  margin-inline: auto;
  padding: 24px 0 18px;
}
.site-search-heading {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 16px;
  margin-bottom: 14px;
}
.site-search-heading h2 {
  margin: 0;
  font-size: 15px;
  font-weight: 600;
  color: var(--deploy-muted);
}
.search-close {
  display: inline-flex;
  align-items: center;
  gap: 9px;
  min-height: 34px;
  padding: 6px 8px;
  border: 0;
  border-radius: 7px;
  background: transparent;
  color: var(--deploy-muted);
}
.search-close:hover {
  background: var(--deploy-paper);
  color: var(--deploy-primary);
}
.search-close span {
  font-size: 11px;
}
.site-search-form {
  display: flex;
  align-items: center;
  gap: 16px;
  min-height: 62px;
  padding: 7px 8px 7px 20px;
  border: 1px solid var(--deploy-line);
  border-radius: 12px;
  background: var(--deploy-white);
  transition: box-shadow 0.15s;
}
.site-search-form:focus-within {
  border-color: var(--deploy-line);
  box-shadow: 0 0 0 3px #00000012;
}
.site-search-form > svg {
  width: 24px;
  height: 24px;
  color: var(--deploy-muted);
}
.site-search-form input {
  width: 100%;
  min-width: 0;
  padding: 6px 0;
  border: 0;
  outline: none;
  background: transparent;
  color: var(--deploy-ink);
  font-size: 17px;
}
.site-search-form input::placeholder {
  color: var(--deploy-faint);
}
.site-search-form input::-webkit-search-cancel-button {
  display: none;
}
.search-clear {
  border: 0;
  background: none;
  color: var(--deploy-faint);
  min-width: 32px;
  min-height: 40px;
  font-size: 24px;
}
.search-submit {
  flex-shrink: 0;
  min-width: 80px;
  min-height: 44px;
  border: 0;
  border-radius: 8px;
  background: var(--deploy-primary);
  color: white;
  font-size: 14px;
}
.search-submit:hover {
  background: var(--deploy-primary-hover);
}
.search-controls {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 18px;
  padding: 20px 0 16px;
}
.search-scopes {
  display: flex;
  gap: 6px;
  flex-shrink: 0;
}
.search-scopes button {
  min-height: 34px;
  padding: 6px 16px;
  border: 0;
  border-radius: 7px;
  color: var(--deploy-muted);
  background: transparent;
  font-size: 13px;
}
.search-scopes button.active {
  background: var(--deploy-paper);
  color: var(--deploy-primary);
  font-weight: 600;
}
.search-controls > span {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-size: 12px;
  color: var(--deploy-muted);
}
.search-result-viewport {
  max-height: min(380px, calc(100dvh - 342px));
  overflow-y: auto;
  overscroll-behavior: contain;
  scrollbar-width: thin;
}
.search-columns {
  display: grid;
  grid-template-columns: minmax(0, 1fr) minmax(0, 1fr);
  gap: 40px;
  padding: 4px 4px 10px;
}
.search-columns.single {
  grid-template-columns: minmax(0, 1fr);
}
.search-section {
  min-width: 0;
}
.search-section-heading {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 10px;
}
.search-section-heading h3 {
  margin: 0;
  color: var(--deploy-muted);
  font-size: 12px;
  font-weight: 500;
}
.search-section-heading > span {
  color: var(--deploy-faint);
  font-size: 11px;
}
.search-result-link {
  display: flex;
  align-items: center;
  gap: 12px;
  min-height: 66px;
  padding: 10px;
  border-radius: 9px;
}
.search-result-link:hover,
.search-result-link:focus-visible {
  background: var(--deploy-paper);
}
.search-result-icon {
  display: grid;
  place-items: center;
  width: 36px;
  height: 38px;
  flex-shrink: 0;
  border: 1px solid var(--deploy-line);
  border-radius: 8px;
  background: var(--deploy-white);
  color: var(--deploy-primary);
}
.search-result-icon svg {
  width: 20px;
  height: 20px;
  stroke: currentColor;
  stroke-width: 1.5;
}
.article-icon {
  color: var(--deploy-primary);
}
.search-result-copy {
  display: grid;
  gap: 3px;
  min-width: 0;
}
.search-result-copy strong {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  color: var(--deploy-ink);
  font-size: 14px;
  font-weight: 500;
}
.search-result-copy small {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  color: var(--deploy-muted);
  font-size: 12px;
}
.search-result-arrow {
  margin-left: auto;
  color: var(--deploy-faint);
  font-size: 14px;
}
.search-more {
  display: inline-flex;
  min-height: 36px;
  align-items: center;
  margin: 8px 0 0 10px;
  padding: 4px 0;
  border: 0;
  background: none;
  color: var(--deploy-primary);
  font-size: 12px;
}
.search-more:hover {
  color: var(--deploy-primary);
  text-decoration: underline;
}
.search-feedback {
  padding: 26px 12px 34px;
  color: var(--deploy-muted);
  font-size: 13px;
}
.search-feedback p {
  margin: 0 0 8px;
  color: var(--deploy-muted);
}
.search-feedback span {
  font-size: 12px;
}
.search-feedback button {
  min-height: 34px;
  padding: 5px 10px;
  border: 1px solid var(--deploy-line);
  border-radius: 7px;
  background: var(--deploy-white);
  color: var(--deploy-primary);
  font-size: 12px;
}
.search-skeleton {
  display: grid;
  gap: 12px;
  padding: 8px 10px;
}
.search-skeleton i {
  display: block;
  height: 46px;
  border-radius: 8px;
  background: var(--deploy-paper);
  animation: site-pulse 1.4s infinite;
}
.search-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 18px;
  margin-top: 16px;
  padding-top: 14px;
  border-top: 1px solid var(--deploy-line);
  color: var(--deploy-faint);
  font-size: 11px;
}
.search-keyboard-hint {
  display: flex;
  align-items: center;
  gap: 6px;
  white-space: nowrap;
}
.search-keyboard-hint kbd {
  padding: 0 4px;
  border: 1px solid var(--deploy-line);
  border-radius: 4px;
  background: var(--deploy-white);
  font: inherit;
}
@media (max-width: 1280px) {
  .site-search-trigger {
    width: 140px;
  }
  .site-search-trigger kbd {
    display: none;
  }
}
@media (max-width: 760px) {
  .site-search-trigger {
    width: 42px;
    min-height: 42px;
    justify-content: center;
    padding: 8px;
  }
  .site-search-trigger > span {
    display: none;
  }
  .site-search-dialog {
    top: 20px;
    width: calc(100% - 24px);
    max-height: calc(100dvh - 40px);
    border-radius: 8px;
  }
  .site-search-dialog::backdrop {
    background: #0006;
  }
  .site-search-shell {
    width: calc(100% - 32px);
    padding: 16px 0;
  }
  .site-search-form {
    gap: 10px;
    min-height: 54px;
    padding-left: 12px;
  }
  .site-search-form > svg {
    width: 20px;
    height: 20px;
  }
  .site-search-form input {
    font-size: 16px;
  }
  .search-submit {
    min-width: 58px;
    min-height: 38px;
  }
  .search-clear {
    min-width: 26px;
    padding: 0;
  }
  .search-columns {
    grid-template-columns: minmax(0, 1fr);
    gap: 22px;
  }
  .search-controls {
    gap: 8px;
    padding: 16px 0;
  }
  .search-scopes button {
    padding-inline: 13px;
  }
  .search-result-viewport {
    max-height: min(480px, calc(100dvh - 310px));
  }
  .search-keyboard-hint {
    display: none;
  }
}
@media (max-height: 500px) {
  .site-search-heading {
    margin-bottom: 8px;
  }
  .site-search-shell {
    padding-top: 8px;
  }
  .search-footer {
    display: none;
  }
  .search-controls {
    padding-block: 8px;
  }
  .search-result-viewport {
    max-height: calc(100dvh - 255px);
  }
}
</style>
