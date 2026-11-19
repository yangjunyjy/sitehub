<script setup>
import { computed, onBeforeUnmount, ref, watch } from "vue";
import { useRoute, useRouter } from "vue-router";
import { articles, searchArticles } from "@/api";
import ArticleList from "./ArticleList.vue";
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
const pages = computed(() => Math.max(1, Math.ceil(total.value / 8)));
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
      pageSize: 8,
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
watch(() => route.fullPath, load, { immediate: true });
onBeforeUnmount(() => {
  generation++;
  controller?.abort();
});
</script>
<template>
  <section id="articles" class="content-shell feed-shell">
    <div class="feed-main glass-panel">
      <header class="section-heading">
        <div>
          <h2>
            {{ heading }}
            <small v-if="!loading && !error">{{ total }} 篇</small>
          </h2>
        </div>
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
        v-if="loading || error"
        :loading="loading"
        :error="error"
        @retry="load"
      /><ArticleList v-else-if="items.length" :items="items" /><ContentState
        v-else
        :title="search ? '没有找到相关的文章' : undefined"
        :description="search ? '换一个关键词，再试试看吧。' : undefined"
      />
      <nav v-if="pages > 1 && !error" class="pagination" aria-label="文章分页">
        <button :disabled="page <= 1 || loading" @click="changePage(page - 1)">
          ←</button
        ><span>{{ page }} / {{ pages }}</span
        ><button
          :disabled="page >= pages || loading"
          @click="changePage(page + 1)"
        >
          →
        </button>
      </nav>
    </div>
    <BlogSidebar v-if="sidebar" />
  </section>
</template>
<style lang="scss" scoped>
.feed-shell {
  display: flex;
  gap: 28px;
  align-items: flex-start;
  padding-top: 32px;
  padding-bottom: 80px;
}
.feed-main {
  flex: 1;
  min-width: 0;
  padding: clamp(16px, 3vw, 38px);
}
.section-heading {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding-bottom: 4px;
  h2 {
    margin: 8px 0;
    font-size: 23px;
    font-weight: 500;
    small {
      margin-left: 10px;
      color: var(--muted);
      font-size: 10px;
      font-weight: 400;
    }
  }
}
.section-eyebrow {
  font-size: 8px;
  letter-spacing: 2.5px;
  color: #aaa2b7;
}
.sort-button {
  font-size: 10px;
  border: 0;
  background: none;
  color: var(--muted);
  cursor: pointer;
  padding: 8px 0;
  &.active {
    color: var(--accent);
  }
}
@media (max-width: 1050px) {
  .feed-shell {
    gap: 35px;
  }
}
@media (max-width: 640px) {
  .feed-shell {
    padding-top: 35px;
    padding-bottom: 45px;
  }
  .section-heading h2 {
    font-size: 20px;
  }
}
</style>
