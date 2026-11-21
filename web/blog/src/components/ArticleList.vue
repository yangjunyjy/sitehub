<script setup>
import { ref } from "vue";
import BlogIcon from "./BlogIcon.vue";
import { articleDate, safeURL } from "@/utils/format";
defineProps({ items: { type: Array, default: () => [] } });
const brokenImages = ref(new Set());
const failImage = (id) => {
  brokenImages.value = new Set([...brokenImages.value, String(id)]);
};
</script>
<template>
  <div class="article-list">
    <article
      v-for="article in items"
      :key="String(article.id)"
      class="article-row"
    >
      <div class="article-row__body">
        <div class="article-row__meta">
          <span>{{ articleDate(article) }}</span
          ><span v-if="article.category?.name" class="article-row__category">{{
            article.category.name
          }}</span
          ><span v-if="article.is_top || article.isTop" class="pinned"
            >置顶</span
          >
        </div>
        <h2>
          <router-link :to="{ name: 'article', params: { id: article.id } }">{{
            article.title
          }}</router-link>
        </h2>
        <p v-if="article.summary" class="article-row__summary">{{ article.summary }}</p>
        <div class="article-row__bottom">
          <span
            ><BlogIcon name="clock" :size="13" />
            {{ article.views || 0 }} 次阅读</span
          >
          <div class="article-row__tags">
            <router-link
              v-for="tag in (article.tags || []).slice(0, 2)"
              :key="tag.id"
              :to="{ name: 'tags', query: { tags: String(tag.id) } }"
              ># {{ tag.name }}</router-link
            >
          </div>
        </div>
      </div>
      <router-link
        class="article-row__cover"
        :to="{ name: 'article', params: { id: article.id } }"
        :aria-label="'阅读：' + article.title"
        ><img
          v-if="safeURL(article.cover) && !brokenImages.has(String(article.id))"
          :src="safeURL(article.cover)"
          alt=""
          loading="lazy"
          @error="failImage(article.id)" />
        <div v-else class="cover-placeholder">
          <BlogIcon name="feather" :size="34" /><i></i></div
      ></router-link>
    </article>
  </div>
</template>
<style lang="scss" scoped>
.article-row {
  display: flex;
  gap: 35px;
  align-items: center;
  padding: 32px 0;
  border-bottom: 1px solid var(--line);
  &__body {
    flex: 1;
    min-width: 0;
  }
  &__meta {
    display: flex;
    gap: 14px;
    font-size: 11px;
    color: var(--muted);
    align-items: center;
  }
  &__category {
    color: var(--accent);
  }
  .pinned {
    color: #bf8f75;
    background: #f8efe8;
    border-radius: 3px;
    padding: 1px 5px;
  }
  h2 {
    font-size: 22px;
    font-weight: 500;
    line-height: 1.7;
    margin: 11px 0 10px;
    a:hover {
      color: var(--accent);
    }
  }
  &__summary {
    font-size: 13px;
    line-height: 2;
    color: #676775;
    margin: 0 0 18px;
    display: -webkit-box;
    -webkit-line-clamp: 2;
    -webkit-box-orient: vertical;
    overflow: hidden;
  }
  &__bottom {
    display: flex;
    flex-wrap: wrap;
    justify-content: space-between;
    gap: 12px;
    font-size: 10px;
    color: var(--muted);
    > span {
      display: flex;
      align-items: center;
      gap: 5px;
    }
  }
  &__tags {
    display: flex;
    gap: 12px;
  }
  &__cover {
    display: block;
    width: 215px;
    height: 143px;
    flex-shrink: 0;
    overflow: hidden;
    border-radius: 5px;
    img {
      width: 100%;
      height: 100%;
      object-fit: cover;
      transition: transform 0.6s;
    }
    &:hover img {
      transform: scale(1.06);
    }
  }
}
.cover-placeholder {
  position: relative;
  height: 100%;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 15px;
  color: #ffffffc9;
  background: linear-gradient(145deg, #b4c5d6, #d9cbdc 55%, #eed6d7);
  overflow: hidden;
  span {
    font-size: 8px;
    letter-spacing: 0.3em;
  }
  i {
    position: absolute;
    bottom: -45px;
    width: 280px;
    height: 90px;
    border-radius: 50%;
    background: #ffffff55;
    transform: rotate(-8deg);
  }
}
@media (max-width: 640px) {
  .article-row {
    gap: 18px;
    padding: 25px 0;
    h2 {
      font-size: 18px;
      margin: 8px 0;
    }
    &__cover {
      width: 110px;
      height: 110px;
    }
    &__summary {
      -webkit-line-clamp: 2;
      font-size: 12px;
      margin-bottom: 10px;
    }
    &__meta {
      gap: 8px;
      font-size: 10px;
    }
    &__tags {
      display: none;
    }
  }
}
</style>
