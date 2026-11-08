<script setup>
import { ref, onMounted, onBeforeUnmount } from "vue";
import { getPopularArticles } from "@/api";
const articles = ref([]),
  loading = ref(true),
  error = ref(false);
let controller,
  generation = 0;
async function load() {
  const current = ++generation;
  controller?.abort();
  controller = new AbortController();
  loading.value = true;
  error.value = false;
  try {
    const response = await getPopularArticles({ signal: controller.signal });
    if (current === generation)
      articles.value = Array.isArray(response.data)
        ? response.data
        : response.data?.items || [];
  } catch {
    if (current === generation) error.value = true;
  } finally {
    if (current === generation) loading.value = false;
  }
}
onMounted(load);
onBeforeUnmount(() => {
  generation++;
  controller?.abort();
});
</script>
<template>
  <section class="blog-discovery" aria-label="热门文章">
    <header>
      <h2>大家都在读</h2>
      <span>按阅读量</span>
    </header>
    <p v-if="loading" class="discovery-state">正在加载热门文章…</p>
    <p v-else-if="error" class="discovery-state">
      热门文章暂不可用 <button type="button" @click="load">重试</button>
    </p>
    <ol v-else-if="articles.length">
      <li v-for="(article, index) in articles" :key="article.id">
        <span>{{ String(index + 1).padStart(2, "0") }}</span>
        <div>
          <router-link :to="`/article/${article.id}`">{{
            article.title
          }}</router-link
          ><small
            >{{ Number(article.views || 0).toLocaleString() }} 次阅读</small
          >
        </div>
      </li>
    </ol>
    <p v-else class="discovery-state">文章发布后会出现在这里。</p>
  </section>
</template>
<style scoped>
.blog-discovery {
  padding: 22px 0 28px;
  margin-bottom: 24px;
  border-bottom: 1px solid var(--deploy-line);
}
header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 12px;
}
h2 {
  font-size: 14px;
  font-weight: 600;
  margin: 0;
}
header > span {
  font-size: 10px;
  color: var(--deploy-faint);
}
ol {
  list-style: none;
  padding: 0;
  margin: 22px 0 0;
}
li {
  display: flex;
  gap: 14px;
  margin-top: 20px;
}
li > span {
  font: 17px var(--site-mono);
  color: var(--deploy-line);
  padding-top: 3px;
}
li:first-child > span {
  color: var(--deploy-primary);
}
li a {
  font-size: 13px;
  line-height: 1.7;
  overflow-wrap: anywhere;
}
li a:hover {
  color: var(--deploy-primary);
}
small {
  display: block;
  color: var(--deploy-faint);
  font-size: 10px;
  margin-top: 6px;
}
.discovery-state {
  font-size: 12px;
  color: var(--deploy-muted);
}
button {
  color: var(--deploy-primary);
  background: none;
  border: 0;
}
</style>
