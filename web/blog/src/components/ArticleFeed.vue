<script setup>
import { computed, onBeforeUnmount, ref, watch } from "vue";
import { useRoute, useRouter } from "vue-router";
import { articles, searchArticles } from "@/api";
import ArticleList from "./ArticleList.vue";
import BlogPagination from "./BlogPagination.vue";
import BlogSidebar from "./BlogSidebar.vue";
import ContentState from "./ContentState.vue";
const props = defineProps({
  search: Boolean,
  sidebar: Boolean,
  heading: { type: String, default: "最新文章" },
});
const route = useRoute(),
  router = useRouter();
const items = ref([]),
  total = ref(0),
  loading = ref(false),
  error = ref("");
const page = computed(() =>
  /^\d+$/.test(String(route.query.page))
    ? Math.max(1, Math.min(100000, Number(route.query.page)))
    : 1,
);
const pageSize = 8;
const pages = computed(() => Math.max(1, Math.ceil(total.value / pageSize)));
let generation = 0,
  controller;
async function load() {
  const id = ++generation;
  controller?.abort();
  controller = new AbortController();
  loading.value = true;
  error.value = "";
  try {
    const params = {
      page: page.value,
      pageSize,
      sort: route.query.sort === "hot" ? "hot" : "latest",
    };
    if (/^[1-9]\d*$/.test(String(route.query.category)))
      params.categoryID = String(route.query.category);
    if (route.query.tags)
      params.tagIDs = String(route.query.tags)
        .split(",")
        .filter((value) => /^[1-9]\d*$/.test(value));
    if (props.search)
      params.keyword = String(route.query.q || "")
        .trim()
        .slice(0, 200);
    if (props.search && !params.keyword) {
      items.value = [];
      total.value = 0;
      return;
    }
    const result = await (props.search ? searchArticles : articles)(params, {
      signal: controller.signal,
    });
    if (id !== generation) return;
    if (!Array.isArray(result.data?.items))
      throw new Error("文章列表响应无效，请稍后重试");
    items.value = result.data.items;
    total.value = Number(result.data.total || 0);
    if (page.value > pages.value)
      await router.replace({
        path: route.path,
        hash: route.hash,
        query: {
          ...route.query,
          page: pages.value === 1 ? undefined : pages.value,
        },
      });
  } catch (failure) {
    if (id === generation) {
      error.value = failure.message;
      items.value = [];
    }
  } finally {
    if (id === generation) loading.value = false;
  }
}
const changePage = (next) => {
  if (next < 1 || next > pages.value || loading.value) return;
  router.push({
    path: route.path,
    query: { ...route.query, page: next === 1 ? undefined : next },
    hash: "#articles",
  });
};
watch(() => route.query, load, { immediate: true });
onBeforeUnmount(() => {
  generation++;
  controller?.abort();
});
</script>
<template>
  <section id="articles" class="feed-paper">
    <div class="content-shell feed-shell">
      <div class="feed-main" :aria-busy="loading">
        <header class="section-heading">
          <h2>
            {{ heading }}
            <small v-if="!loading && !error">{{ total }} 篇</small>
          </h2>
          <button
            v-if="!search"
            class="sort-button"
            :class="{ active: route.query.sort === 'hot' }"
            @click="
              router.push({
                path: route.path,
                query: {
                  ...route.query,
                  sort: route.query.sort === 'hot' ? undefined : 'hot',
                  page: undefined,
                },
              })
            "
          >
            {{ route.query.sort === "hot" ? "热门优先 ↓" : "最新发布 ↓" }}
          </button>
        </header>
        <ContentState
          v-if="error || (loading && !items.length)"
          :loading="loading"
          :error="error"
          @retry="load"
        />
        <ArticleList v-else-if="items.length" :items="items" />
        <ContentState
          v-else
          :title="search ? '没有找到相关的文章' : undefined"
          :description="search ? '换一个关键词，再试试看吧。' : undefined"
        />
        <BlogPagination
          v-if="total && !error"
          label="文章分页"
          :page="page"
          :pages="pages"
          :busy="loading"
          @change="changePage"
        />
      </div>
      <BlogSidebar v-if="sidebar" />
    </div>
  </section>
</template>
<style lang="scss" scoped>
.feed-paper {
  background: var(--journal-paper);
  color: var(--journal-ink);
}
.feed-shell {
  display: flex;
  align-items: flex-start;
  gap: clamp(36px, 5.5vw, 86px);
  padding-block: clamp(32px, 5vw, 68px);
}
.feed-main {
  flex: 1;
  min-width: 0;
}
.section-heading {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  gap: 12px;
  border-bottom: 2px solid var(--journal-ink);
  padding-bottom: 12px;
}
h2 {
  margin: 0;
  font: 600 clamp(24px, 3vw, 34px)/1.3 var(--journal-serif);
}
h2 small {
  color: var(--journal-muted);
  font: 12px var(--font-sans);
  display: inline-block;
  margin-left: 6px;
}
.sort-button {
  color: var(--journal-muted);
  font-size: 12px;
  border: 0;
  background: transparent;
  padding: 5px 0 5px 8px;
}
.sort-button:hover,
.sort-button.active {
  color: var(--journal-accent);
}
</style>
