<script setup>
import { formatDate } from "@/utils/format";
import ReadingToc from "@/components/product/ReadingToc.vue";
import ContentState from "@/components/product/ContentState.vue";
import MarkdownContent from "@/components/product/MarkdownContent.vue";
import { useArticleReader } from "@/composables/useArticleReader";
import { ref } from "vue";
import ArticleReadingTools from "@/components/product/ArticleReadingTools.vue";
import ArticleRecommendations from "@/components/product/ArticleRecommendations.vue";
const fontSize = ref(16);

const {
  article,
  loading,
  error,
  notFound,
  copyStatus,
  toc,
  renderedContent,
  loadArticle,
  copyLink,
} = useArticleReader();
</script>

<template>
  <div class="article-page">
    <div v-if="loading" class="site-shell article-loading">
      <span></span><span></span><span></span>
    </div>
    <section v-else-if="error" class="site-shell article-state">
      <ContentState
        error
        :title="notFound ? '没有找到这篇文章' : '文章暂时无法加载'"
        :message="error"
        action-label="重新加载"
        @retry="loadArticle"
        ><router-link class="site-button site-button--outline" to="/blog"
          >返回博客</router-link
        ></ContentState
      >
    </section>
    <template v-else-if="article">
      <div
        class="site-shell article-layout"
        :class="{ 'article-layout--without-toc': !toc.length }"
      >
        <header class="article-hero">
          <div class="article-context">
            <router-link class="article-back" to="/blog"> 返回博客</router-link>
            <p class="article-category">
              <router-link
                v-if="article.category?.id"
                :to="{
                  path: '/blog',
                  query: { category: String(article.category.id) },
                }"
                >{{ article.category.name }}</router-link
              ><span v-else>未分类</span>
            </p>
          </div>
          <h1>{{ article.title }}</h1>
          <p v-if="article.summary" class="article-summary">
            {{ article.summary }}
          </p>
          <div class="article-meta">
            <div class="article-meta__details">
              <span v-if="article.author">{{
                article.author.name || article.author.nickname
              }}</span>
              <span>{{
                formatDate(
                  article.published_at ||
                    article.publishedAt ||
                    article.createdAt ||
                    article.created_at,
                )
              }}</span>
              <span>{{ article.views || 0 }} 次阅读</span>
            </div>
            <div class="article-share-wrap">
              <button class="article-share" type="button" @click="copyLink">
                <svg viewBox="0 0 24 24" fill="none" aria-hidden="true">
                  <rect x="8" y="8" width="12" height="12" rx="2" />
                  <path
                    d="M16 8V4a1 1 0 0 0-1-1H4a1 1 0 0 0-1 1v11a1 1 0 0 0 1 1h4"
                  />
                </svg>
                复制链接
              </button>
              <p v-if="copyStatus" class="share-status" role="status">
                {{ copyStatus }}
              </p>
            </div>
          </div>
          <div v-if="article.tags?.length" class="article-topic-links">
            <router-link
              v-for="tag in article.tags"
              :key="tag.id"
              :to="{ path: '/blog', query: { tags: String(tag.id) } }"
              ># {{ tag.name }}</router-link
            >
          </div>
          <ArticleReadingTools
            :article="article"
            @font-change="fontSize = $event"
          />
        </header>
        <MarkdownContent
          class="article-content"
          :style="{ fontSize: fontSize + 'px' }"
          :html="renderedContent"
        />
        <aside v-if="toc.length" class="article-toc-nav">
          <ReadingToc :items="toc" title="本文目录" />
        </aside>
      </div>

      <ArticleRecommendations :article="article" />

      <section class="article-next">
        <div class="site-shell article-next__inner">
          <div>
            <p class="site-kicker">KEEP READING</p>
            <h2>下一篇，换个视角</h2>
          </div>
          <div>
            <router-link to="/blog">更多文章</router-link
            ><router-link to="/blog/archives">文章归档</router-link>
          </div>
        </div>
      </section>
    </template>
  </div>
</template>

<style scoped>
.article-topic-links {
  display: flex;
  gap: 8px 16px;
  flex-wrap: wrap;
  margin-top: 18px;
  color: var(--deploy-primary);
  font-size: 12px;
}
.article-layout {
  display: grid;
  grid-template-columns: minmax(0, 760px) 210px;
  grid-template-areas: "hero toc" "body toc";
  align-items: start;
  justify-content: center;
  gap: 0 64px;
  max-width: 1100px;
  padding: 48px 0 72px;
}
.article-layout--without-toc {
  grid-template-columns: minmax(0, 800px);
  grid-template-areas: "hero" "body";
}
.article-hero {
  grid-area: hero;
  min-width: 0;
  padding-bottom: 32px;
  margin-bottom: 8px;
  border-bottom: 1px solid var(--deploy-line);
}
.article-context {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  font-size: 12px;
  color: var(--deploy-muted);
}
.article-category {
  color: var(--deploy-primary);
  margin: 0;
}
.article-hero h1 {
  font-size: 36px;
  font-weight: 600;
  line-height: 1.45;
  letter-spacing: -0.025em;
  margin: 24px 0 18px;
  overflow-wrap: anywhere;
}
.article-summary {
  font-size: 16px;
  color: var(--deploy-muted);
  line-height: 1.9;
  margin: 0 0 24px;
}
.article-meta {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
}
.article-meta__details {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
  font-size: 12px;
  color: var(--deploy-faint);
}
.article-share-wrap {
  position: relative;
}
.article-share {
  display: flex;
  align-items: center;
  gap: 6px;
  border: 0;
  background: transparent;
  color: var(--deploy-muted);
  font-size: 11px;
  white-space: nowrap;
}
.article-share svg {
  width: 14px;
  height: 14px;
  stroke: currentColor;
  stroke-width: 1.5;
}
.share-status {
  position: absolute;
  font-size: 11px;
  right: 0;
  white-space: nowrap;
  color: var(--deploy-primary);
}
.article-content {
  grid-area: body;
  padding: 0;
}
.article-toc-nav {
  grid-area: toc;
  position: sticky;
  top: 96px;
  max-height: calc(100dvh - 128px);
  display: flex;
  min-height: 0;
}
.article-toc-nav :deep(.reading-toc) {
  width: 100%;
}
.article-next {
  border-top: 1px solid var(--deploy-line);
  padding: 36px 0;
  background: var(--deploy-paper);
}
.article-next__inner {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 24px;
}
.article-next h2 {
  font-size: 22px;
  margin: 10px 0 0;
}
.article-next__inner > div:last-child {
  display: flex;
  gap: 24px;
  color: var(--deploy-primary);
  font-size: 13px;
}
.article-loading {
  display: grid;
  gap: 20px;
  padding-block: 100px;
}
.article-loading span {
  height: 24px;
  background: var(--deploy-paper);
  animation: site-pulse 1.4s infinite;
}
.article-loading span:first-child {
  height: 48px;
  width: 65%;
}
.article-state {
  min-height: 65vh;
  padding-block: 64px;
}
@media (max-width: 1000px) {
  .article-layout {
    gap: 0 36px;
    grid-template-columns: minmax(0, 1fr) 180px;
  }
  .article-layout--without-toc {
    grid-template-columns: minmax(0, 1fr);
  }
}
@media (max-width: 760px) {
  .article-layout,
  .article-layout--without-toc {
    display: block;
    padding: 32px 0 48px;
  }
  .article-toc-nav {
    display: none;
  }
  .article-hero h1 {
    font-size: 28px;
  }
  .article-meta {
    align-items: flex-start;
  }
  .article-next__inner {
    align-items: flex-start;
    flex-direction: column;
  }
  .article-next h2 {
    font-size: 20px;
  }
}
</style>
