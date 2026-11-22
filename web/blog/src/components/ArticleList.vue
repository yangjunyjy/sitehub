<script setup>
import { onBeforeUnmount, ref } from "vue";
import { articleDate, safeURL } from "@/utils/format";
defineProps({ items: { type: Array, default: () => [] } });

const brokenCovers = ref(new Set());
let observer;
const vReveal = {
  mounted(row) {
    if (
      !("IntersectionObserver" in window) ||
      matchMedia("(prefers-reduced-motion: reduce)").matches
    )
      return;
    observer ??= new IntersectionObserver(
      (entries) => {
        const entering = entries
          .filter((entry) => entry.isIntersecting)
          .sort((a, b) => a.boundingClientRect.top - b.boundingClientRect.top);
        entering.forEach(({ target }, index) => {
          target.style.setProperty("--delay", Math.min(index, 4) * 55 + "ms");
          target.classList.add("arrived");
          observer.unobserve(target);
        });
      },
      { threshold: 0.06 },
    );
    row.classList.add("awaiting-entry");
    observer.observe(row);
  },
  beforeUnmount(row) {
    observer?.unobserve(row);
  },
};
onBeforeUnmount(() => observer?.disconnect());
</script>

<template>
  <div class="article-list">
    <article
      v-for="article in items"
      :key="article.id"
      v-reveal
      class="article-row"
    >
      <time
        v-if="articleDate(article) !== '日期待补充'"
        :datetime="articleDate(article)"
      >
        <span>{{ articleDate(article).slice(5).replace("-", ".") }}</span>
        <small>{{ articleDate(article).slice(0, 4) }}</small>
      </time>
      <div class="entry">
        <h3>
          <RouterLink :to="{ name: 'article', params: { id: article.id } }">{{
            article.title
          }}</RouterLink>
          <span v-if="article.is_top || article.isTop" class="pin">置顶</span>
        </h3>
        <p v-if="article.summary">{{ article.summary }}</p>
        <div class="entry-meta">
          <RouterLink
            v-if="article.category?.id"
            :to="{
              name: 'categories',
              query: { category: String(article.category.id) },
            }"
            >{{ article.category.name }}</RouterLink
          >
          <span>{{ article.views || 0 }} 次阅读</span>
          <RouterLink
            v-for="tag in (article.tags || []).slice(0, 2)"
            :key="tag.id"
            :to="{ name: 'tags', query: { tags: String(tag.id) } }"
            >#{{ tag.name }}</RouterLink
          >
        </div>
      </div>
      <RouterLink
        v-if="safeURL(article.cover) && !brokenCovers.has(article.cover)"
        class="thumbnail"
        :to="{ name: 'article', params: { id: article.id } }"
        tabindex="-1"
        aria-hidden="true"
      >
        <img
          :src="safeURL(article.cover)"
          alt=""
          loading="lazy"
          @error="brokenCovers.add(article.cover)"
        />
      </RouterLink>
    </article>
  </div>
</template>

<style scoped lang="scss">
.article-row {
  display: grid;
  grid-template-columns: 4.4rem minmax(0, 1fr) auto;
  align-items: center;
  column-gap: clamp(14px, 2vw, 28px);
  border-bottom: 1px solid var(--journal-rule);
  padding: 14px 0;
}
@media (prefers-reduced-motion: no-preference) {
  .awaiting-entry:not(.arrived):not(:focus-within) {
    opacity: 0;
  }
  .arrived {
    animation: article-rise 540ms cubic-bezier(0.2, 0.75, 0.22, 1)
      var(--delay, 0ms) backwards;
  }
  @keyframes article-rise {
    from {
      opacity: 0;
      transform: perspective(800px) translateY(22px) rotateX(-5deg);
    }
    to {
      opacity: 1;
      transform: none;
    }
  }
}
.article-row:not(:has(time)) .entry {
  grid-column: 1 / 3;
}
time {
  display: grid;
  font:
    22px/1.15 Georgia,
    serif;
  color: var(--journal-ink);
  font-variant-numeric: tabular-nums;
  small {
    margin-top: 4px;
    color: var(--journal-muted);
    font: 11px/1.2 var(--font-sans);
    letter-spacing: 0.08em;
  }
}
.entry {
  grid-column: 2;
  min-width: 0;
  h3 {
    margin: 0;
    font: 600 clamp(18px, 2vw, 23px)/1.35 var(--journal-serif);
  }
  h3 a:hover {
    color: var(--journal-accent);
  }
  p {
    margin: 4px 0 0;
    overflow: hidden;
    white-space: nowrap;
    text-overflow: ellipsis;
    font-size: 13px;
    color: var(--journal-muted);
  }
}
.pin {
  display: inline-block;
  vertical-align: middle;
  margin-left: 7px;
  border: 1px solid #d6b6a7;
  border-radius: 3px;
  padding: 1px 4px;
  font: 11px/1.25 var(--font-sans);
  color: var(--journal-accent);
}
.entry-meta {
  display: flex;
  flex-wrap: wrap;
  gap: 0 12px;
  align-items: center;
  margin-top: 4px;
  color: var(--journal-muted);
  font-size: 12px;
}
.entry-meta a:hover {
  color: var(--journal-accent);
  text-decoration: underline;
}
.thumbnail {
  display: block;
  width: 88px;
  height: 58px;
}
.thumbnail img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  border-radius: 3px;
}
@media (max-width: 600px) {
  .article-row {
    grid-template-columns: 44px minmax(0, 1fr) auto;
    gap: 9px;
    padding: 12px 0;
  }
  time {
    font-size: 16px;
  }
  .thumbnail {
    width: 54px;
    height: 48px;
  }
  .entry-meta {
    font-size: 11px;
    gap: 0 7px;
  }
}
</style>
