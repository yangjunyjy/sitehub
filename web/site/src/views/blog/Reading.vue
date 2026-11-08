<script setup>
import { computed, ref } from "vue";
import { useRoute } from "vue-router";
import { useReadingLibrary } from "@/composables/useReadingLibrary";
import BlogPageHeading from "@/components/blog/BlogPageHeading.vue";
import { archiveDate } from "@/utils/blog-index";
const { saved, history, persistent, remove, clearHistory, toggle, isSaved } =
  useReadingLibrary();
const route = useRoute(),
  keyword = ref("");
const showHistory = computed(() => route.query.tab === "history");
const items = computed(() =>
  (showHistory.value ? history.value : saved.value).filter((item) =>
    item.title
      .toLocaleLowerCase()
      .includes(keyword.value.trim().toLocaleLowerCase()),
  ),
);
</script>
<template>
  <div class="site-shell reading-page">
    <BlogPageHeading
      eyebrow="YOUR LIBRARY / 留给自己的阅读时间"
      title="阅读空间"
      description="收好想读的文章，接着上一次的思考。"
    />
    <div class="reading-workspace">
      <aside>
        <nav aria-label="阅读空间导航">
          <router-link
            to="/blog/reading"
            :aria-current="!showHistory ? 'page' : undefined"
            >稍后阅读 <span>{{ saved.length }}</span></router-link
          ><router-link
            to="/blog/reading?tab=history"
            :aria-current="showHistory ? 'page' : undefined"
            >最近阅读 <span>{{ history.length }}</span></router-link
          >
        </nav>
        <p>
          {{
            persistent
              ? "记录仅保存在当前浏览器，不随账号同步。"
              : "当前浏览器存储不可用，记录仅在本次阅读期间保留。"
          }}
        </p>
        <small>稍后阅读最多 50 篇<br />最近阅读保留 12 篇</small>
      </aside>
      <section aria-label="我的文章记录">
        <div class="reading-toolbar">
          <input
            v-model="keyword"
            type="search"
            aria-label="查找阅读记录"
            placeholder="查找阅读记录…"
            maxlength="200"
          /><button
            v-if="showHistory && history.length"
            type="button"
            @click="clearHistory"
          >
            清空阅读历史
          </button>
        </div>
        <div v-if="!items.length" class="reading-empty">
          <span aria-hidden="true">⌑</span>
          <h2>
            {{
              keyword
                ? "没有匹配的文章"
                : showHistory
                  ? "还没有阅读记录"
                  : "给下一次阅读留个位置"
            }}
          </h2>
          <p>
            {{
              keyword
                ? "换个关键词试试。"
                : showHistory
                  ? "读过的文章会出现在这里。"
                  : "在文章旁点击“稍后阅读”，想读的时候随时回来。"
            }}
          </p>
          <router-link to="/blog">发现文章 →</router-link>
        </div>
        <ol v-else>
          <li v-for="(item, index) in items" :key="item.id">
            <span class="reading-number">{{
              String(index + 1).padStart(2, "0")
            }}</span>
            <div>
              <router-link :to="`/article/${item.id}`">{{
                item.title
              }}</router-link>
              <p v-if="item.summary">{{ item.summary }}</p>
              <time v-if="archiveDate(item.date)">{{
                archiveDate(item.date)
              }}</time>
            </div>
            <button
              v-if="!showHistory"
              type="button"
              :aria-label="`移除 ${item.title}`"
              @click="remove(item.id)"
            >
              移除</button
            ><button
              v-else
              type="button"
              :aria-pressed="isSaved(item.id)"
              @click="toggle(item)"
            >
              {{ isSaved(item.id) ? "已保存" : "稍后阅读" }}
            </button>
          </li>
        </ol>
      </section>
    </div>
  </div>
</template>
<style scoped>
.reading-page {
  padding-bottom: 64px;
}
.reading-workspace {
  display: grid;
  grid-template-columns: 200px minmax(0, 1fr);
  gap: 64px;
}
aside nav {
  display: grid;
  gap: 8px;
}
aside nav a {
  display: flex;
  justify-content: space-between;
  gap: 20px;
  padding: 12px;
  border-radius: 5px;
  font-size: 13px;
}
aside nav a[aria-current] {
  background: var(--deploy-soft);
  color: var(--deploy-primary);
}
aside p {
  margin: 24px 12px 12px;
  font-size: 12px;
  color: var(--deploy-muted);
}
aside small {
  display: block;
  margin: 0 12px;
  color: var(--deploy-faint);
  font-size: 11px;
  line-height: 1.9;
}
section {
  min-width: 0;
}
.reading-toolbar {
  display: flex;
  gap: 16px;
  justify-content: space-between;
  align-items: center;
  border-bottom: 1px solid var(--deploy-line);
  padding-bottom: 24px;
}
input {
  border: 1px solid var(--deploy-line);
  border-radius: 6px;
  color: var(--deploy-ink);
  background: var(--deploy-white);
  padding: 8px 12px;
  min-width: 0;
  width: 260px;
  font-size: 13px;
}
button {
  border: 0;
  background: transparent;
  color: var(--deploy-muted);
  font-size: 12px;
  white-space: nowrap;
  padding: 8px;
}
button:hover {
  color: var(--deploy-primary);
}
ol {
  list-style: none;
  padding: 0;
  margin: 0;
}
li {
  display: flex;
  align-items: start;
  gap: 24px;
  border-bottom: 1px solid var(--deploy-line);
  padding: 28px 0;
}
.reading-number {
  font: 13px var(--site-mono);
  color: var(--deploy-faint);
  padding-top: 6px;
}
li > div {
  flex: 1;
  min-width: 0;
}
li a {
  font-size: 18px;
  font-weight: 500;
  overflow-wrap: anywhere;
}
li p {
  font-size: 13px;
  color: var(--deploy-muted);
  margin: 8px 0;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
  overflow-wrap: anywhere;
}
time {
  font: 11px var(--site-mono);
  color: var(--deploy-faint);
}
.reading-empty {
  text-align: center;
  padding: 60px 12px;
}
.reading-empty > span {
  font-size: 48px;
  color: var(--deploy-primary);
}
.reading-empty h2 {
  font-size: 23px;
}
.reading-empty p {
  color: var(--deploy-muted);
  font-size: 13px;
}
.reading-empty a {
  color: var(--deploy-primary);
  font-size: 13px;
}
@media (max-width: 760px) {
  .reading-workspace {
    grid-template-columns: minmax(0, 1fr);
    gap: 24px;
  }
  aside nav {
    grid-template-columns: 1fr 1fr;
  }
  aside p {
    margin-top: 16px;
  }
  aside small {
    display: none;
  }
  li {
    gap: 12px;
  }
  .reading-number {
    display: none;
  }
  li a {
    font-size: 16px;
  }
}
</style>
