<script setup>
import { onBeforeUnmount, ref, watch } from "vue";
import { useRoute } from "vue-router";
import ArticleActions from "./components/ArticleActions.vue";
import ContentState from "@/components/ContentState.vue";
import ArticleBody from "./components/ArticleBody.vue";
import DiscussionFeed from "@/components/DiscussionFeed.vue";
import { articleDetail, recordView } from "@/api";
import { useMember } from "@/composables/useMember";
import { articleDate } from "@/utils/format";
const { member } = useMember();
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
  copyNotice.value = "";
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
    if (
      !["markdown", "html"].includes(
        result.data.content_format || result.data.contentFormat || "markdown",
      )
    )
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
        <nav class="reader-crumb" aria-label="文章位置">
          <router-link :to="{ name: 'home' }">云间小记</router-link>
          <template v-if="article.category?.name"
            ><span aria-hidden="true">/</span
            ><router-link
              :to="{
                name: 'categories',
                query: { category: String(article.category.id) },
              }"
              >{{ article.category.name }}</router-link
            ></template
          >
        </nav>
        <h1>{{ article.title }}</h1>
        <p v-if="article.summary" class="reader-deck">{{ article.summary }}</p>
        <div class="reader-byline">
          <strong v-if="article.author?.name">{{ article.author.name }}</strong>
          <time
            v-if="articleDate(article) !== '日期待补充'"
            :datetime="articleDate(article)"
            >{{ articleDate(article) }}</time
          >
        </div>
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
        <div v-if="article.tags?.length">
          <router-link
            v-for="tag in article.tags || []"
            :key="tag.id"
            :to="{ name: 'tags', query: { tags: String(tag.id) } }"
            ># {{ tag.name }}</router-link
          >
        </div>
        <button class="text-button" @click="share">分享这篇文章 ↗</button>
        <p v-if="copyNotice" role="status">{{ copyNotice }}</p>
      </div>
      <template v-if="article">
        <ArticleActions
          :key="[route.params.id, member.user?.id].join(':')"
          :article-id="String(route.params.id)"
        />
      </template>
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
<style scoped lang="scss">
.reader-page {
  --discussion-ink: var(--journal-ink);
  --discussion-rule: var(--journal-rule);
  --discussion-muted: var(--journal-muted);
  --discussion-accent: var(--journal-accent);
  --discussion-wash: var(--journal-wash);
  --reader-measure: 710px;
  --reader-gap: clamp(24px, 3.4vw, 55px);
  --reader-grid: minmax(0, 1fr) minmax(0, var(--reader-measure)) minmax(0, 1fr);
  min-height: 80vh;
  background: var(--journal-paper);
  color: var(--journal-ink);
  padding: clamp(90px, 9vw, 128px) clamp(20px, 3vw, 46px) 32px;
}
.reader-shell {
  width: min(100%, 1450px);
  margin: auto;
}
.reader-title,
.article-bottom,
.article-back {
  width: min(100%, var(--reader-measure));
  margin-inline: auto;
}
.reader-title {
  border-bottom: 1px solid var(--journal-rule);
  padding: 0 0 clamp(24px, 3.3vw, 41px);
}
.reader-crumb,
.reader-byline {
  display: flex;
  align-items: center;
  gap: 10px;
  flex-wrap: wrap;
  font-size: 13px;
  color: var(--journal-muted);
}
.reader-crumb a:hover,
.article-back a:hover {
  text-decoration: underline;
}
.reader-crumb span {
  color: #99919c;
}
.reader-title h1 {
  margin: 18px 0 14px;
  color: var(--journal-ink);
  font:
    650 clamp(32px, 4.1vw, 54px) / 1.26 Georgia,
    "Songti SC",
    SimSun,
    serif;
  letter-spacing: -0.035em;
  text-wrap: pretty;
}
.reader-deck {
  margin: 0;
  line-height: 1.75;
  font-size: clamp(16px, 1.65vw, 20px);
  color: var(--journal-muted);
}
.reader-byline:not(:empty) {
  margin-top: 20px;
}
.reader-byline strong {
  color: var(--journal-ink);
}
.reader-byline time {
  font-variant-numeric: tabular-nums;
}
.reader-byline strong + time::before {
  content: "·";
  margin-right: 10px;
}
.article-bottom {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 14px;
  border-top: 1px solid var(--journal-rule);
  padding: 18px 0 30px;
}
.article-bottom > div {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}
.article-bottom a {
  font-size: 13px;
  color: var(--journal-accent);
  text-decoration: underline;
  text-underline-offset: 0.22em;
}
.article-bottom button {
  margin-left: auto;
  font-size: 13px;
  color: var(--journal-accent);
}
.article-bottom p {
  width: 100%;
  margin: 0;
  text-align: right;
  font-size: 12px;
}
.reader-shell :deep(.discussion) {
  width: min(100%, var(--reader-measure));
}
.article-back {
  text-align: center;
  font-size: 13px;
  padding: 30px 0 45px;
  color: var(--journal-accent);
}
@media (max-width: 640px) {
  .reader-page {
    padding-top: 86px;
  }
  .reader-byline:not(:empty) {
    margin-top: 14px;
  }
}
</style>
