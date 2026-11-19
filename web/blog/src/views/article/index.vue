<script setup>
import { onBeforeUnmount, ref, watch } from "vue";
import { useRoute } from "vue-router";
import SkyHero from "@/components/SkyHero.vue";
import ContentState from "@/components/ContentState.vue";
import ArticleBody from "./components/ArticleBody.vue";
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
    if (result.data.content_format && result.data.content_format !== 'markdown')
      throw new Error('这篇文章的格式暂不支持阅读，请联系博主');
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
  <SkyHero
    :title="article?.title || '云间的一段文字'"
    :subtitle="
      article
        ? articleDate(article) +
          (article.category?.name ? ' · ' + article.category.name : '')
        : 'THE JOURNAL'
    "
    eyebrow="WORDS, THOUGHTS & LITTLE DISCOVERIES"
  />
  <section class="content-shell">
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
          :to="{ path: '/tags', query: { tags: String(tag.id) } }"
          ># {{ tag.name }}</router-link
        >
      </div>
      <button class="text-button" @click="share">分享这篇文章 ↗</button>
      <p v-if="copyNotice" role="status">{{ copyNotice }}</p>
    </div>
    <div class="article-back">
      <router-link to="/">← 回到文章列表</router-link>
    </div>
  </section>
</template>
<style lang="scss" scoped>
.article-bottom {
  display: flex;
  justify-content: space-between;
  gap: 20px;
  flex-wrap: wrap;
  border-top: 1px solid var(--line);
  padding: 25px 0 10px;
  > div {
    display: flex;
    gap: 10px;
    flex-wrap: wrap;
    align-items: center;
    a {
      font-size: 11px;
      color: #a68db6;
      background: #f6f1fa;
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
