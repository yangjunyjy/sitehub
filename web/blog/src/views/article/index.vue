<script setup>
import { onBeforeUnmount, ref, watch } from "vue";
import { useRoute } from "vue-router";
import ArticleActions from "./components/ArticleActions.vue";
import ContentState from "@/components/ContentState.vue";
import ArticleBody from "./components/ArticleBody.vue";
import DiscussionFeed from "@/components/DiscussionFeed.vue";
import { articleDetail, recordView } from "@/api";
import { articleDate } from "@/utils/format";
const route = useRoute(),
  article = ref(null),
  loading = ref(true),
  error = ref(""),
  notFound = ref(false),
  copyNotice = ref("");
let generation = 0,
  controller;
async function load() {
  const id = ++generation;
  controller?.abort();
  controller = new AbortController();
  loading.value = true;
  error.value = "";
  notFound.value = false;
  article.value = null;
  try {
    if (!/^[1-9]\d*$/.test(String(route.params.id))) {
      notFound.value = true;
      return;
    }
    const result = await articleDetail(route.params.id, {
      signal: controller.signal,
    });
    if (id !== generation) return;
    if (!result.data || typeof result.data.content !== "string")
      throw new Error("文章响应无效，请稍后重试");
    if (result.data.content_format && result.data.content_format !== "markdown")
      throw new Error("这篇文章的格式暂不支持阅读，请联系博主");
    article.value = result.data;
    document.title = article.value.title + " · 云间小记";
    recordView(route.params.id).catch(() => {});
  } catch (failure) {
    if (id === generation) {
      notFound.value = failure.status === 404;
      error.value = failure.message;
    }
  } finally {
    if (id === generation) loading.value = false;
  }
}
async function share() {
  try {
    await navigator.clipboard.writeText(window.location.href);
    copyNotice.value = "文章链接已复制";
  } catch {
    copyNotice.value = "请复制浏览器地址栏中的链接";
  }
}
watch(() => route.params.id, load, { immediate: true });
onBeforeUnmount(() => {
  generation++;
  controller?.abort();
});
</script>
<template>
  <div class="reader-page">
    <section class="reader-shell">
      <header v-if="article" class="reader-title">
        <router-link :to="{ name: 'home' }">云间小记</router-link>
        <span v-if="article.category?.name">
          / {{ article.category.name }}</span
        >
        <h1>{{ article.title }}</h1>
        <p>{{ articleDate(article) }}</p>
      </header>
      <ContentState v-if="loading" loading /><ContentState
        v-else-if="notFound"
        title="这篇文章暂时不在这里"
        description="文章可能已下线，回到首页看看其他故事吧。"
      /><ContentState
        v-else-if="error"
        :error="error"
        @retry="load"
      /><ArticleBody v-else-if="article" :article="article" />
      <div v-if="article" class="article-bottom">
        <div>
          <router-link
            v-for="tag in article.tags || []"
            :key="tag.id"
            :to="{ name: 'tags', query: { tags: String(tag.id) } }"
            ># {{ tag.name }}</router-link
          >
        </div>
        <ArticleActions
          :key="String(route.params.id)"
          :article-id="String(route.params.id)"
        />
        <button class="text-button" @click="share">分享这篇文章 ↗</button>
        <p v-if="copyNotice" role="status">{{ copyNotice }}</p>
      </div>
      <DiscussionFeed
        v-if="article"
        :key="route.params.id"
        :article-id="String(route.params.id)"
      />
      <div class="article-back">
        <router-link :to="{ name: 'home' }">← 回到文章列表</router-link>
      </div>
    </section>
  </div>
</template>
<style lang="scss" scoped>
.reader-page {
  min-height: 80vh;
  background: #fff;
  color: #25312b;
  padding: 100px 20px 32px;
}
.reader-shell {
  width: min(100%, 1020px);
  margin: auto;
}
.reader-title {
  border-bottom: 1px solid #dce2d9;
  padding-bottom: 24px;
  font-size: 13px;
  color: #607365;
}
.reader-title h1 {
  margin: 15px 0 8px;
  color: #202a24;
  font:
    650 clamp(32px, 5vw, 52px) / 1.25 Georgia,
    "Songti SC",
    SimSun,
    serif;
  letter-spacing: -0.04em;
}
.reader-title p {
  margin: 0;
  font-size: 13px;
}
.article-bottom {
  display: flex;
  justify-content: space-between;
  gap: 20px;
  flex-wrap: wrap;
  border-top: 1px solid var(--line);
  padding: 25px 0 32px;
  > div {
    display: flex;
    gap: 10px;
    flex-wrap: wrap;
    align-items: center;
    a {
      font-size: 11px;
      color: #4a6852;
      background: #f2f5ee;
      padding: 5px 12px;
      border-radius: 3px;
    }
  }
  p {
    width: 100%;
    font-size: 11px;
    color: var(--muted);
    margin: 0;
  }
}
.article-back {
  text-align: center;
  font-size: 12px;
  padding: 35px 0 60px;
  color: #aa9ab8;
}
</style>
