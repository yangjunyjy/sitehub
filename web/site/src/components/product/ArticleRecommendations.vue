<script setup>
import { ref, watch, onBeforeUnmount } from "vue";
import { getArticles } from "@/api";
const props = defineProps({ article: { type: Object, required: true } });
const articles = ref([]),
  error = ref(false),
  loading = ref(false);
let requestId = 0,
  controller;
async function load() {
  const current = ++requestId;
  controller?.abort();
  controller = new AbortController();
  articles.value = [];
  error.value = false;
  loading.value = true;
  try {
    const response = await getArticles(
      {
        page: 1,
        pageSize: 4,
        sort: "latest",
        ...(props.article.category?.id
          ? { categoryId: props.article.category.id }
          : {}),
      },
      { signal: controller.signal },
    );
    if (current === requestId)
      articles.value = (response.data?.items || [])
        .filter((item) => String(item.id) !== String(props.article.id))
        .slice(0, 3);
  } catch {
    if (current === requestId) error.value = true;
  } finally {
    if (current === requestId) loading.value = false;
  }
}
watch(() => props.article.id, load, { immediate: true });
onBeforeUnmount(() => {
  requestId++;
  controller?.abort();
});
</script>
<template>
  <section
    v-if="articles.length || loading || error"
    class="article-recommendations site-shell"
    aria-label="继续阅读"
  >
    <header>
      <h2>继续阅读</h2>
      <router-link to="/blog">浏览全部文章 →</router-link>
    </header>
    <p v-if="loading">正在查找相关文章…</p>
    <p v-else-if="error">
      相关文章暂时无法加载。<button type="button" @click="load">重试</button>
    </p>
    <div v-else class="recommendation-grid">
      <router-link
        v-for="item in articles"
        :key="item.id"
        :to="`/article/${item.id}`"
        ><small>{{ item.category?.name || "工程实践" }}</small>
        <h3>{{ item.title }}</h3>
        <p>{{ item.summary }}</p>
        <span>阅读全文 →</span></router-link
      >
    </div>
  </section>
</template>
<style scoped>
.article-recommendations {
  max-width: 1100px;
  padding-block: 40px 56px;
  border-top: 1px solid var(--deploy-line);
}
header {
  display: flex;
  justify-content: space-between;
  gap: 20px;
  align-items: center;
  margin-bottom: 24px;
}
h2 {
  font-size: 22px;
  margin: 0;
  font-weight: 600;
}
header a {
  color: var(--deploy-primary);
  font-size: 13px;
}
.recommendation-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 20px;
}
.recommendation-grid > a {
  display: flex;
  flex-direction: column;
  border: 1px solid var(--deploy-line);
  border-radius: 8px;
  padding: 24px;
  min-width: 0;
}
.recommendation-grid > a:hover {
  border-color: var(--deploy-primary);
}
small {
  color: var(--deploy-primary);
  font-size: 11px;
}
h3 {
  font-size: 17px;
  line-height: 1.7;
  margin: 12px 0;
  overflow-wrap: anywhere;
}
p {
  font-size: 13px;
  line-height: 1.8;
  color: var(--deploy-muted);
}
.recommendation-grid p {
  display: -webkit-box;
  -webkit-line-clamp: 3;
  -webkit-box-orient: vertical;
  overflow: hidden;
  margin: 0 0 20px;
}
span {
  font-size: 12px;
  color: var(--deploy-primary);
  margin-top: auto;
}
button {
  color: var(--deploy-primary);
  border: 0;
  background: transparent;
}
@media (max-width: 760px) {
  .recommendation-grid {
    grid-template-columns: 1fr;
  }
  .recommendation-grid > a {
    padding: 20px;
  }
}
</style>
