<script setup>
import { onMounted, onBeforeUnmount, ref } from "vue";
import { getLatestArticles } from "@/api";
import { formatDate, truncateText } from "@/utils/format";
defineProps({ home: { type: Object, required: true } });
const articles = ref([]);
const loadingArticles = ref(true);
const articlesError = ref(false);
let requestId = 0;
let controller;

const loadArticles = async () => {
  const current = ++requestId;
  controller?.abort();
  controller = new AbortController();
  loadingArticles.value = true;
  articlesError.value = false;
  try {
    const response = await getLatestArticles({ signal: controller.signal });
    if (current !== requestId) return;
    const value = response.data?.items || response.data || [];
    articles.value = Array.isArray(value) ? value.slice(0, 3) : [];
  } catch {
    if (current !== requestId) return;
    articlesError.value = true;
    articles.value = [];
  } finally {
    if (current === requestId) loadingArticles.value = false;
  }
};
onMounted(loadArticles);
onBeforeUnmount(() => {
  requestId++;
  controller?.abort();
});
</script>
<template>
  <section
    id="insights"
    class="site-section insight-section"
    :aria-busy="loadingArticles"
  >
    <div class="site-shell">
      <header class="insight-heading">
        <div>
          <span class="site-index">04</span>
          <p class="site-kicker">Insights</p>
          <h2 class="site-heading">{{ home.blogTitle }}</h2>
        </div>
        <router-link to="/blog" target="_blank" rel="noopener"
          >查看全部文章 ↗</router-link
        >
      </header>
      <div v-if="articles.length" class="insight-grid">
        <router-link
          v-for="article in articles"
          :key="article.id"
          :to="`/article/${article.id}`"
          ><span>{{ article.category?.name || "技术实践" }}</span>
          <h3>{{ article.title }}</h3>
          <p>{{ truncateText(article.summary, 90) }}</p>
          <footer>
            <time>{{ formatDate(article.createdAt) }}</time
            ><b>阅读文章 </b>
          </footer></router-link
        >
      </div>
      <div v-else-if="articlesError" class="empty-state" role="status">
        <h3>文章暂时无法加载</h3>
        <p>请稍后重试，或前往产品文档继续了解 Deploy。</p>
        <button
          class="site-button site-button--outline"
          :disabled="loadingArticles"
          @click="loadArticles"
        >
          重新加载
        </button>
      </div>
      <div v-else-if="!loadingArticles" class="empty-state">
        <h3>产品文章正在准备中</h3>
        <p>内容发布后，最新文章将在这里自动展示。</p>
      </div>
      <div
        v-else
        class="insight-loading"
        role="status"
        aria-label="正在加载最新文章"
      >
        <i></i><i></i><i></i>
      </div>
    </div>
  </section>
</template>
<style scoped>
.insight-heading {
  display: flex;
  align-items: end;
  justify-content: space-between;
  gap: 24px;
  margin-bottom: 32px;
}
.insight-heading > a {
  min-height: 44px;
  color: var(--deploy-primary);
  font-size: 14px;
}
.insight-grid,
.insight-loading {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 20px;
}
.insight-grid > a {
  display: flex;
  flex-direction: column;
  min-height: 280px;
  padding: 28px;
  border: 1px solid var(--deploy-line);
  border-radius: 8px;
  background: var(--deploy-white);
  transition: box-shadow 0.2s;
}
.insight-grid > a:hover {
  box-shadow: var(--deploy-shadow);
}
.insight-grid > a > span {
  align-self: flex-start;
  padding: 3px 10px;
  border-radius: 6px;
  background: var(--deploy-soft);
  color: var(--deploy-primary);
  font-size: 12px;
}
.insight-grid h3 {
  margin: 22px 0 12px;
  font-size: 20px;
}
.insight-grid p {
  margin: 0;
  color: var(--deploy-muted);
  font-size: 14px;
}
.insight-grid footer {
  display: flex;
  justify-content: space-between;
  gap: 12px;
  margin-top: auto;
  padding-top: 24px;
  color: var(--deploy-muted);
  font-size: 12px;
}
.insight-grid footer b {
  color: var(--deploy-primary);
  font-weight: 500;
}
.insight-loading i {
  height: 280px;
  border-radius: 8px;
  background: var(--deploy-paper);
  animation: site-pulse 1.4s infinite;
}

@media (max-width: 760px) {
  .insight-grid,
  .insight-loading {
    grid-template-columns: 1fr;
  }
  .insight-grid > a {
    min-height: 240px;
  }
}
@media (max-width: 460px) {
  .insight-heading {
    align-items: flex-start;
    flex-direction: column;
    gap: 16px;
  }
}
</style>
