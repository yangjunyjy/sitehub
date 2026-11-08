<script setup>
import { computed, ref, watch } from "vue";
import { formatDate, truncateText } from "@/utils/format";
import ProductIcon from "./ProductIcon.vue";
import { useReadingLibrary } from "@/composables/useReadingLibrary";
const { isSaved, toggle } = useReadingLibrary();
const props = defineProps({ article: { type: Object, required: true } });
const imageFailed = ref(false);
watch(
  () => props.article.cover,
  () => {
    imageFailed.value = false;
  },
);
const date = computed(
  () =>
    props.article.published_at ||
    props.article.publishedAt ||
    props.article.created_at ||
    props.article.createdAt,
);
const hasCover = computed(
  () => Boolean(props.article.cover) && !imageFailed.value,
);
</script>
<template>
  <article class="blog-article-row" :class="{ 'has-cover': hasCover }">
    <router-link
      v-if="hasCover"
      class="article-thumbnail"
      :to="`/article/${article.id}`"
      tabindex="-1"
      aria-hidden="true"
      ><img
        :src="article.cover"
        alt=""
        loading="lazy"
        @error="imageFailed = true"
    /></router-link>
    <div v-else class="article-mark" aria-hidden="true">
      <ProductIcon name="playbook" />
    </div>
    <div class="article-row-body">
      <div class="article-labels">
        <span>{{ article.category?.name || "未分类" }}</span
        ><span v-if="article.is_top || article.isTop" class="pinned">置顶</span>
      </div>
      <h2>
        <router-link :to="`/article/${article.id}`">{{
          article.title
        }}</router-link>
      </h2>
      <p v-if="article.summary" class="row-summary">
        {{ truncateText(article.summary, 160) }}
      </p>
      <div v-if="article.tags?.length" class="row-tags">
        <template v-for="tag in article.tags" :key="tag.id || tag.name"
          ><router-link
            v-if="tag.id"
            :to="{ path: '/blog', query: { tags: String(tag.id) } }"
            ># {{ tag.name }}</router-link
          ><span v-else># {{ tag.name }}</span></template
        >
      </div>
      <div class="row-meta">
        <time v-if="date" :datetime="date">{{ formatDate(date) }}</time
        ><span v-if="article.author?.name">{{ article.author.name }}</span
        ><span>{{ article.views || 0 }} 次阅读</span>
      </div>
    </div>
    <div class="row-actions">
      <button
        type="button"
        class="row-save"
        :aria-pressed="isSaved(article.id)"
        :aria-label="`${isSaved(article.id) ? '取消稍后阅读' : '稍后阅读'}：${article.title}`"
        @click="toggle(article)"
      >
        <svg viewBox="0 0 24 24" aria-hidden="true">
          <path d="M6 4h12v17l-6-4-6 4Z" /></svg
        >{{ isSaved(article.id) ? "已保存" : "稍后阅读" }}</button
      ><router-link
        class="row-read"
        :to="`/article/${article.id}`"
        :aria-label="`阅读：${article.title}`"
        >阅读全文
      </router-link>
    </div>
  </article>
</template>
<style scoped>
.row-actions {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  gap: 16px;
  align-self: center;
}
.row-save {
  display: flex;
  align-items: center;
  gap: 5px;
  white-space: nowrap;
  border: 0;
  background: transparent;
  color: var(--deploy-muted);
  font-size: 11px;
  padding: 4px 0;
}
.row-save[aria-pressed="true"] {
  color: var(--deploy-primary);
}
.row-save svg {
  width: 14px;
  height: 14px;
  fill: none;
  stroke: currentColor;
  stroke-width: 1.6;
}
.row-save[aria-pressed="true"] svg {
  fill: var(--deploy-soft);
}
.row-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 6px 14px;
  color: var(--deploy-muted);
  font-size: 12px;
  margin: 10px 0;
}
.row-tags a:hover {
  color: var(--deploy-primary);
  text-decoration: underline;
}
.blog-article-row {
  display: grid;
  grid-template-columns: 48px minmax(0, 1fr) auto;
  align-items: start;
  gap: 20px;
  padding: 28px;
  transition: background 0.18s;
}
.blog-article-row + .blog-article-row {
  border-top: 1px solid var(--deploy-line);
}
.blog-article-row:hover {
  background: var(--deploy-white);
}
.article-mark {
  display: grid;
  place-items: center;
  width: 46px;
  height: 52px;
  border-radius: 12px;
  background: var(--deploy-paper);
  color: var(--deploy-primary);
}
.article-mark :deep(svg) {
  width: 24px;
  height: 24px;
  stroke: currentColor;
  stroke-width: 1.5;
}
.article-labels {
  display: flex;
  gap: 8px;
  margin-bottom: 8px;
  color: var(--deploy-muted);
  font-size: 12px;
}
.article-labels .pinned {
  color: var(--deploy-primary);
}
.article-row-body h2 {
  margin: 0;
  font-size: 20px;
  line-height: 1.55;
  overflow-wrap: anywhere;
}
.article-row-body h2 a:hover {
  color: var(--deploy-primary);
}
.row-summary {
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
  margin: 10px 0 0;
  color: var(--deploy-muted);
  font-size: 14px;
  line-height: 1.85;
  overflow-wrap: anywhere;
}
.row-meta {
  display: flex;
  flex-wrap: wrap;
  gap: 8px 16px;
  margin-top: 18px;
  color: var(--deploy-muted);
  font-size: 12px;
}
.row-read {
  display: inline-flex;
  align-items: center;
  gap: 10px;
  align-self: end;
  padding-block: 10px;
  color: var(--deploy-primary);
  font-size: 12px;
  white-space: nowrap;
}
.row-read span {
  font-size: 17px;
}
.has-cover {
  grid-template-columns: 164px minmax(0, 1fr) auto;
}
.article-thumbnail {
  grid-row: 1 / 3;
  overflow: hidden;
  width: 164px;
  height: 116px;
  border-radius: 10px;
  background: var(--deploy-paper);
}
.article-thumbnail img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}
.has-cover .row-read {
  display: inline-flex;
}
@media (max-width: 760px) {
  .row-actions {
    grid-column: 2;
    flex-direction: row;
    align-items: center;
    justify-content: space-between;
  }
  .has-cover .row-actions {
    grid-column: 1/-1;
  }
  .blog-article-row {
    padding: 22px;
    gap: 16px;
    grid-template-columns: 40px minmax(0, 1fr);
  }
  .article-mark {
    width: 40px;
    height: 46px;
  }
  .article-row-body h2 {
    font-size: 18px;
  }
  .row-read {
    grid-column: 2;
    padding: 0;
    justify-self: start;
  }
  .has-cover {
    grid-template-columns: minmax(0, 1fr) 104px;
  }
  .has-cover .article-row-body {
    grid-column: 1;
    grid-row: 1;
  }
  .article-thumbnail {
    grid-column: 2;
    grid-row: 1;
    width: 104px;
    height: 80px;
  }
}
@media (max-width: 420px) {
  .blog-article-row {
    padding: 20px 16px;
    column-gap: 12px;
  }
  .has-cover {
    grid-template-columns: minmax(0, 1fr);
  }
  .article-thumbnail {
    grid-column: 1;
    grid-row: auto;
    width: 100%;
    height: 150px;
  }
  .has-cover .article-row-body {
    grid-row: auto;
  }
}
</style>
