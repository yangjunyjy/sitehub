<script setup>
import BlogArticleRow from "@/components/product/BlogArticleRow.vue";
import ContentState from "@/components/product/ContentState.vue";
import ReadingLibrary from "@/components/product/ReadingLibrary.vue";
import BlogDiscovery from "@/components/product/BlogDiscovery.vue";
import { useBlogFeed } from "@/composables/useBlogFeed";
const {
  articles,
  categories,
  loading,
  error,
  categoryError,
  categoryLoading,
  total,
  keywordInput,
  tags,
  tagError,
  tagLoading,
  initialized,
  filters,
  totalPages,
  categoryName,
  selectedTags,
  tagName,
  toggleTag,
  loadTags,
  updateFilters,
  load,
  loadCategories,
  search,
  reset,
  changePage,
  view,
  pageNumbers,
} = useBlogFeed();
</script>

<template>
  <div class="blog-page">
    <header class="blog-heading site-shell">
      <div>
        <p class="site-kicker">THE JOURNAL / 记录与实践</p>
        <h1>技术博客</h1>
        <p class="blog-description">
          关于工程、工具与构建过程。让经验留下来，让思考继续发生。
        </p>
      </div>
      <div class="blog-heading-actions">
        <router-link to="/blog/archives">沿时间线阅读 ↗</router-link
        ><ReadingLibrary />
      </div>
    </header>
    <div class="site-shell blog-layout">
      <section class="blog-results" aria-label="博客文章">
        <div class="blog-filters">
          <div class="blog-controls">
            <nav class="category-tabs" aria-label="文章分类">
              <button
                type="button"
                :class="{ active: !filters.category }"
                :aria-pressed="!filters.category"
                @click="updateFilters({ category: '', page: 1 })"
              >
                全部文章
              </button>
              <button
                v-for="category in categories"
                :key="category.id"
                type="button"
                :class="{ active: filters.category === String(category.id) }"
                :aria-pressed="filters.category === String(category.id)"
                @click="
                  updateFilters({ category: String(category.id), page: 1 })
                "
              >
                {{ category.name }}
              </button>
            </nav>
            <form class="blog-search" role="search" @submit.prevent="search">
              <label for="blog-keyword" class="visually-hidden">查找文章</label>
              <div>
                <svg viewBox="0 0 24 24" aria-hidden="true">
                  <circle cx="10.5" cy="10.5" r="6.5" />
                  <path d="m16 16 4 4" /></svg
                ><input
                  id="blog-keyword"
                  v-model="keywordInput"
                  type="search"
                  maxlength="200"
                  placeholder="搜索文章"
                /><button type="submit">搜索</button>
              </div>
            </form>
          </div>
          <p v-if="categoryError" class="category-feedback" role="status">
            分类暂时无法加载
            <button
              type="button"
              :disabled="categoryLoading"
              @click="loadCategories"
            >
              重试
            </button>
          </p>
        </div>
        <div class="result-toolbar">
          <p aria-live="polite">
            {{
              loading
                ? "正在查找文章…"
                : error
                  ? "文章加载失败"
                  : filters.keyword || filters.category || filters.tags
                    ? `找到 ${total} 篇文章`
                    : `共 ${total} 篇文章`
            }}
          </p>
          <div class="result-options">
            <div
              class="blog-view-switch"
              role="group"
              aria-label="文章显示方式"
            >
              <button
                type="button"
                aria-label="列表视图"
                :aria-pressed="view === 'list'"
                @click="view = 'list'"
              >
                <svg viewBox="0 0 20 20" aria-hidden="true">
                  <path d="M3 5h14M3 10h14M3 15h14" />
                </svg></button
              ><button
                type="button"
                aria-label="卡片视图"
                :aria-pressed="view === 'grid'"
                @click="view = 'grid'"
              >
                <svg viewBox="0 0 20 20" aria-hidden="true">
                  <path d="M3 3h5v5H3Zm9 0h5v5h-5ZM3 12h5v5H3Zm9 0h5v5h-5Z" />
                </svg>
              </button>
            </div>
            <label
              >排序<select
                aria-label="文章排序"
                :value="filters.sort"
                @change="updateFilters({ sort: $event.target.value, page: 1 })"
              >
                <option value="latest">最新发布</option>
                <option value="hot">最多阅读</option>
              </select></label
            >
          </div>
        </div>
        <div class="active-filters">
          <span v-if="filters.keyword"
            >关键词：{{ filters.keyword
            }}<button
              type="button"
              aria-label="清除搜索关键词"
              @click="updateFilters({ keyword: '', page: 1 })"
            >
              ×
            </button></span
          >
          <span v-if="filters.category"
            >{{ categoryName
            }}<button
              type="button"
              aria-label="清除分类筛选"
              @click="updateFilters({ category: '', page: 1 })"
            >
              ×
            </button></span
          >
          <span v-for="id in selectedTags" :key="id"
            >{{ tagName(id)
            }}<button
              type="button"
              :aria-label="`移除标签 ${tagName(id)}`"
              @click="toggleTag(id)"
            >
              &times;
            </button></span
          >
          <button
            v-if="filters.keyword || filters.category || filters.tags"
            type="button"
            class="reset-filters"
            @click="reset"
          >
            清除筛选
          </button>
        </div>
        <div class="results-content" :aria-busy="loading">
          <div
            v-if="loading && !initialized"
            class="blog-loading"
            role="status"
            aria-label="正在加载文章"
          >
            <div v-for="i in 3" :key="i"><i /><i /><i /></div>
          </div>
          <ContentState
            v-else-if="error"
            class="blog-error"
            error
            title="内容服务暂不可用"
            :message="error"
            action-label="重新加载"
            @retry="load"
          />
          <ContentState
            v-else-if="!articles.length"
            :title="
              filters.keyword
                ? '没有找到相关文章'
                : filters.category
                  ? '这个分类还没有文章'
                  : '文章正在准备中'
            "
            :message="
              filters.keyword
                ? '换个关键词试试，或清除筛选查看全部文章。'
                : '新的产品动态和工程实践发布后，会展示在这里。'
            "
            :action-label="
              filters.keyword || filters.category || filters.tags
                ? '查看全部文章'
                : undefined
            "
            @retry="reset"
            ><router-link to="/docs" class="site-button site-button--outline"
              >阅读产品文档</router-link
            ></ContentState
          >
          <div
            v-else
            class="blog-article-list"
            :class="{ 'is-grid': view === 'grid' }"
          >
            <BlogArticleRow
              v-for="article in articles"
              :key="article.id"
              :article="article"
            />
          </div>
        </div>
        <nav
          v-if="!error && totalPages > 1"
          class="blog-pagination"
          aria-label="文章分页"
        >
          <button
            type="button"
            :disabled="loading || filters.page === 1"
            @click="changePage(filters.page - 1)"
          >
            上一页
          </button>
          <span class="page-count" aria-live="polite"
            >{{ filters.page }} / {{ totalPages }}</span
          >
          <button
            v-for="number in pageNumbers"
            :key="number"
            class="page-number"
            type="button"
            :aria-label="`第 ${number} 页`"
            :aria-current="filters.page === number ? 'page' : undefined"
            :disabled="loading"
            @click="changePage(number)"
          >
            {{ number }}
          </button>
          <button
            type="button"
            :disabled="loading || filters.page === totalPages"
            @click="changePage(filters.page + 1)"
          >
            下一页
          </button>
        </nav>
      </section>
      <aside class="blog-aside">
        <section class="journal-note">
          <p class="site-kicker">HELLO, READER.</p>
          <h2>从一个问题开始，<br />把实践写成答案。</h2>
          <p>
            这里记录开发中的探索、踩坑与复盘。按主题深入，也可以沿着时间慢慢读。
          </p>
          <div>
            <router-link to="/blog/archives"
              >文章归档 <span>↗</span></router-link
            ><router-link to="/blog/categories"
              >专题分类 <span>↗</span></router-link
            >
          </div>
        </section>
        <BlogDiscovery />
        <section class="blog-resources tag-filter" aria-label="文章标签">
          <h2>按标签浏览</h2>
          <p>可多选，匹配任一标签</p>
          <div class="tag-options">
            <button
              type="button"
              :class="{ active: !selectedTags.length }"
              :aria-pressed="!selectedTags.length"
              @click="updateFilters({ tags: '', page: 1 })"
            >
              全部标签
            </button>
            <button
              v-for="tag in tags"
              :key="tag.id"
              type="button"
              :class="{ active: selectedTags.includes(String(tag.id)) }"
              :aria-pressed="selectedTags.includes(String(tag.id))"
              @click="toggleTag(tag.id)"
            >
              {{ tag.name }}
            </button>
          </div>
          <p v-if="tagError" role="status">
            标签加载失败
            <button type="button" :disabled="tagLoading" @click="loadTags">
              重试
            </button>
          </p>
          <p v-else-if="tagLoading">正在加载标签…</p>
          <p v-else-if="!tags.length">暂未设置标签</p>
        </section>
        <nav class="blog-resources" aria-label="继续发现">
          <h2>继续发现</h2>
          <router-link to="/blog/tags">全部标签 <span>→</span></router-link>
          <router-link to="/blog/reading"
            >我的阅读空间 <span>→</span></router-link
          >
        </nav>
      </aside>
    </div>
  </div>
</template>

<style scoped>
.blog-heading {
  padding: 40px 0 36px;
  border-bottom: 1px solid var(--deploy-line);
  margin-bottom: 32px;
}
.blog-heading h1 {
  font-size: 42px;
  font-weight: 600;
  letter-spacing: -0.03em;
  margin: 12px 0;
}
.blog-heading > p:last-child {
  font-size: 15px;
  color: var(--deploy-muted);
  margin: 0;
}
.results-content {
  min-height: 420px;
}
.active-filters {
  min-height: 12px;
}
.tag-filter {
  margin-bottom: 18px;
}
.tag-filter p {
  color: var(--deploy-muted);
  font-size: 12px;
  margin: 8px 0 14px;
}
.tag-options {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}
.tag-options button {
  border: 1px solid var(--deploy-line);
  background: var(--deploy-white);
  color: var(--deploy-muted);
  border-radius: 6px;
  padding: 5px 10px;
  font-size: 13px;
  min-height: 34px;
}
.tag-options button.active {
  background: var(--deploy-soft);
  border-color: var(--deploy-line);
  color: var(--deploy-primary);
}
@media (max-width: 1080px) {
  .blog-aside {
    grid-row: auto;
  }
  .tag-filter {
    display: block !important;
  }
}

.blog-page {
  padding-block: 28px 64px;
}
.visually-hidden {
  position: absolute;
  width: 1px;
  height: 1px;
  padding: 0;
  margin: -1px;
  overflow: hidden;
  clip-path: inset(50%);
  white-space: nowrap;
  border: 0;
}
.blog-search {
  width: 240px;
  flex-shrink: 0;
}
.blog-search > div {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 3px 3px 3px 11px;
  border: 1px solid var(--deploy-line);
  border-radius: 10px;
  background: var(--deploy-white);
}
.blog-search > div:focus-within {
  border-color: var(--deploy-line);
  box-shadow: 0 0 0 3px #00000012;
}
.blog-search svg {
  width: 18px;
  height: 18px;
  flex-shrink: 0;
  fill: none;
  stroke: var(--deploy-muted);
  stroke-width: 1.6;
}
.blog-search input {
  width: 100%;
  min-width: 0;
  border: 0;
  outline: 0;
  background: transparent;
  color: var(--deploy-ink);
  font-size: 13px;
}
.blog-search input::placeholder {
  color: var(--deploy-faint);
}
.blog-search button {
  min-width: 48px;
  min-height: 32px;
  border: 0;
  border-radius: 7px;
  background: var(--deploy-primary);
  color: white;
  font-size: 13px;
}
.blog-layout {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 256px;
  gap: 40px;
  align-items: start;
}
.blog-results {
  min-width: 0;
  overflow: hidden;
  background: var(--deploy-white);
}
.blog-filters {
  padding: 0 0 0;
}
.blog-controls {
  display: flex;
  align-items: center;
  gap: 20px;
  padding-bottom: 16px;
  border-bottom: 1px solid var(--deploy-line);
}
.category-tabs {
  display: flex;
  flex: 1;
  min-width: 0;
  gap: 6px;
  overflow-x: auto;
  scrollbar-width: thin;
}
.category-tabs button {
  min-height: 40px;
  flex-shrink: 0;
  padding: 8px 14px;
  border: 0;
  border-radius: 8px;
  background: transparent;
  color: var(--deploy-muted);
  font-size: 14px;
}
.category-tabs button:hover {
  background: var(--deploy-paper);
}
.category-tabs button.active {
  background: var(--deploy-soft);
  color: var(--deploy-primary);
  font-weight: 600;
}
.category-feedback {
  margin: 12px 0 0;
  color: #8e795f;
  font-size: 12px;
}
.category-feedback button {
  border: 0;
  background: none;
  color: var(--deploy-primary);
  text-decoration: underline;
}
.result-toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 16px;
  padding: 14px 28px;
}
.result-toolbar p {
  margin: 0;
  color: var(--deploy-muted);
  font-size: 12px;
}
.result-toolbar label {
  display: flex;
  align-items: center;
  gap: 10px;
  color: var(--deploy-muted);
  font-size: 12px;
}
.result-toolbar select {
  min-height: 34px;
  padding: 5px 9px;
  border: 1px solid var(--deploy-line);
  border-radius: 7px;
  background: var(--deploy-white);
  color: var(--deploy-muted);
  font-size: 12px;
}
.active-filters {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 8px;
  padding: 0 28px 12px;
}
.active-filters > span {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  max-width: 100%;
  padding: 3px 8px;
  border-radius: 6px;
  background: var(--deploy-paper);
  color: var(--deploy-primary);
  font-size: 12px;
  overflow-wrap: anywhere;
}
.active-filters button {
  min-width: 24px;
  min-height: 28px;
  border: 0;
  background: none;
  color: var(--deploy-primary);
  font-size: 16px;
}
.active-filters .reset-filters {
  font-size: 12px;
}
.blog-article-list {
  padding-bottom: 4px;
}
.blog-pagination {
  display: flex;
  justify-content: center;
  align-items: center;
  gap: 24px;
  padding: 24px;
  border-top: 1px solid var(--deploy-line);
  color: var(--deploy-muted);
  font-size: 13px;
}
.blog-pagination button {
  min-height: 40px;
  padding: 8px 14px;
  border: 1px solid var(--deploy-line);
  border-radius: 8px;
  background: var(--deploy-white);
  color: var(--deploy-primary);
  font-size: 13px;
}
.blog-loading > div {
  display: grid;
  gap: 12px;
  padding: 30px;
  border-top: 1px solid var(--deploy-line);
}
.blog-loading i {
  width: 100%;
  height: 14px;
  border-radius: 5px;
  background: var(--deploy-paper);
  animation: site-pulse 1.4s infinite;
}
.blog-loading i:first-child {
  width: 65%;
  height: 24px;
}
.blog-loading i:last-child {
  width: 30%;
}
.blog-aside {
  align-self: start;
}
.blog-resources {
  padding: 20px 0;
  border-bottom: 1px solid var(--deploy-line);
  background: var(--deploy-white);
}
.blog-resources h2 {
  margin: 0 0 8px;
  font-size: 14px;
}
.blog-resources a {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 16px;
  padding: 12px 0;
  color: var(--deploy-muted);
  font-size: 14px;
}
.blog-resources a + a {
  border-top: 1px solid var(--deploy-line);
}
@media (max-width: 1080px) {
  .blog-layout {
    grid-template-columns: minmax(0, 1fr);
  }
  .blog-aside {
    position: static;
  }
  .blog-resources {
    display: flex;
    align-items: center;
    flex-wrap: wrap;
    gap: 8px 28px;
    padding: 12px 20px;
  }
  .blog-resources h2 {
    margin: 0;
  }
  .blog-resources a {
    gap: 12px;
    font-size: 13px;
  }
  .blog-resources a + a {
    border-top: 0;
  }
}
@media (max-width: 760px) {
  .blog-controls {
    flex-direction: column;
    align-items: stretch;
    gap: 14px;
  }
  .category-tabs {
    flex: auto;
    padding-bottom: 2px;
  }
  .blog-search {
    width: 100%;
  }
  .blog-filters {
    padding: 16px 16px 0;
  }
  .result-toolbar {
    padding: 14px 16px;
  }
  .active-filters {
    padding-inline: 16px;
  }
  .blog-layout {
    gap: 20px;
  }
  .blog-pagination {
    gap: 18px;
    padding: 20px 12px;
  }
}
@media (max-width: 540px) {
  .blog-resources {
    gap: 4px 24px;
  }
  .blog-resources h2 {
    width: 100%;
  }
  .blog-page {
    padding-block: 20px 40px;
  }
}
</style>
<style scoped>
.blog-heading-actions {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  gap: 16px;
  flex-shrink: 0;
}
.blog-heading-actions > a {
  color: var(--deploy-primary);
  font-size: 12px;
}
.journal-note {
  padding: 24px;
  border: 1px solid var(--deploy-line);
  background: var(--deploy-paper);
  border-radius: 8px;
}
.journal-note .site-kicker {
  font-size: 10px;
}
.journal-note h2 {
  font-size: 19px;
  line-height: 1.7;
  font-weight: 500;
  margin: 16px 0 12px;
}
.journal-note > p:not(.site-kicker) {
  font-size: 12px;
  line-height: 1.9;
  color: var(--deploy-muted);
}
.journal-note > div {
  margin-top: 20px;
  padding-top: 16px;
  border-top: 1px solid var(--deploy-line);
  display: grid;
  gap: 12px;
}
.journal-note a {
  display: flex;
  justify-content: space-between;
  font-size: 12px;
  color: var(--deploy-primary);
}
@media (max-width: 1080px) {
  .blog-aside {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 24px;
    border-top: 1px solid var(--deploy-line);
    padding-top: 24px;
  }
  .journal-note {
    grid-row: span 2;
  }
  .blog-resources {
    display: block;
  }
}
@media (max-width: 600px) {
  .blog-aside {
    grid-template-columns: minmax(0, 1fr);
  }
  .blog-heading-actions {
    flex-direction: row;
    align-items: center;
  }
  .blog-heading {
    margin-bottom: 24px;
  }
  .blog-heading h1 {
    font-size: 34px;
  }
  .blog-description {
    max-width: 30em;
  }
}
.blog-heading {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 24px;
}
.blog-description {
  font-size: 15px;
  color: var(--deploy-muted);
  margin: 0;
}
.result-options {
  display: flex;
  align-items: center;
  gap: 16px;
}
.blog-view-switch {
  display: flex;
  gap: 2px;
  border: 1px solid var(--deploy-line);
  padding: 2px;
  border-radius: 5px;
}
.blog-view-switch button {
  display: grid;
  place-items: center;
  width: 28px;
  height: 26px;
  border: 0;
  border-radius: 3px;
  background: transparent;
  color: var(--deploy-faint);
}
.blog-view-switch button[aria-pressed="true"] {
  background: var(--deploy-soft);
  color: var(--deploy-primary);
}
.blog-view-switch svg {
  width: 16px;
  height: 16px;
  fill: none;
  stroke: currentColor;
  stroke-width: 1.3;
}
.blog-article-list.is-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 16px;
  padding: 0 16px 16px;
}
.is-grid :deep(.blog-article-row) {
  display: flex;
  flex-direction: column;
  gap: 16px;
  padding: 20px;
  border: 1px solid var(--deploy-line);
  border-radius: 8px;
  min-width: 0;
}
.is-grid :deep(.article-row-body) {
  flex: 1;
  width: 100%;
}
.is-grid :deep(.article-row-body h2) {
  font-size: 18px;
}
.is-grid :deep(.article-thumbnail) {
  width: 100%;
  height: 160px;
}
.is-grid :deep(.article-mark) {
  display: none;
}
.is-grid :deep(.row-actions) {
  width: 100%;
  align-items: center;
  flex-direction: row;
  justify-content: space-between;
  border-top: 1px solid var(--deploy-line);
  padding-top: 12px;
}
.is-grid :deep(.row-read) {
  padding: 0;
}
.blog-pagination {
  gap: 6px;
  flex-wrap: wrap;
}
.blog-pagination .page-number {
  padding: 6px 10px;
  min-width: 32px;
}
.page-number[aria-current="page"] {
  color: var(--deploy-primary);
  border-color: var(--deploy-primary);
  background: var(--deploy-soft);
}
.page-count {
  padding-inline: 6px;
  font-size: 11px;
  color: var(--deploy-faint);
}
@media (max-width: 760px) {
  .blog-heading {
    align-items: flex-start;
    flex-direction: column;
    gap: 20px;
  }
  .blog-article-list.is-grid {
    grid-template-columns: 1fr;
  }
  .result-options {
    gap: 8px;
  }
  .blog-pagination {
    gap: 5px;
    padding-inline: 8px;
  }
  .blog-pagination .page-count {
    display: none;
  }
  .blog-pagination button {
    font-size: 11px;
    min-height: 32px;
  }
  .blog-pagination .page-number {
    padding: 4px 7px;
    min-width: 26px;
  }
}
</style>
