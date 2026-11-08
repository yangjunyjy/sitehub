<script setup>
import { computed, ref, watch } from "vue";
import { useRoute, useRouter } from "vue-router";
import { useBlogIndex } from "@/composables/useBlogIndex";
import { archiveGroups } from "@/utils/blog-index";
import BlogPageHeading from "@/components/blog/BlogPageHeading.vue";
import BlogIndexState from "@/components/blog/BlogIndexState.vue";
const { items, loading, error, progress, total, reload } = useBlogIndex();
const route = useRoute(),
  router = useRouter(),
  keyword = ref("");
const year = computed(() =>
  /^\d{4}$/.test(route.query.year) || route.query.year === "undated"
    ? String(route.query.year)
    : "",
);
const month = computed(() =>
  /^(0[1-9]|1[0-2])$/.test(route.query.month) ? String(route.query.month) : "",
);
const query = computed(() =>
  typeof route.query.q === "string" ? route.query.q.trim().slice(0, 200) : "",
);
watch(query, (value) => (keyword.value = value), { immediate: true });
const years = computed(() => archiveGroups(items.value));
const yearItems = computed(() =>
  items.value.filter(
    (item) =>
      !year.value || (item.date.slice(0, 4) || "undated") === year.value,
  ),
);
const months = computed(() =>
  Array.from({ length: 12 }, (_, i) => {
    const value = String(i + 1).padStart(2, "0");
    return {
      value,
      count: yearItems.value.filter((item) => item.date.slice(5, 7) === value)
        .length,
    };
  }),
);
const filtered = computed(() =>
  yearItems.value.filter(
    (item) =>
      (!month.value || item.date.slice(5, 7) === month.value) &&
      (!query.value ||
        item.title
          .toLocaleLowerCase()
          .includes(query.value.toLocaleLowerCase())),
  ),
);
const pageCount = computed(() =>
  Math.max(1, Math.ceil(filtered.value.length / 40)),
);
const page = computed(() =>
  Math.min(
    pageCount.value,
    Math.max(1, Number.parseInt(route.query.page, 10) || 1),
  ),
);
const groups = computed(() =>
  archiveGroups(filtered.value.slice((page.value - 1) * 40, page.value * 40)),
);
function location(patch) {
  const values = {
    year: year.value,
    month: month.value,
    q: query.value,
    page: 1,
    ...patch,
  };
  return {
    path: "/blog/archives",
    query: Object.fromEntries(
      Object.entries(values).filter(
        ([key, value]) => value && !(key === "page" && value === 1),
      ),
    ),
  };
}
const title = (value) => (value === "undated" ? "日期待补充" : value + " 年");
</script>
<template>
  <div class="site-shell archive-page">
    <BlogPageHeading
      eyebrow="THE ARCHIVE / 时间里的记录"
      title="文章归档"
      description="沿着时间线，找回每一次记录。"
    >
      <button
        class="index-refresh"
        type="button"
        :disabled="loading"
        @click="reload"
      >
        刷新归档 ↻
      </button>
    </BlogPageHeading>
    <BlogIndexState
      :loading="loading"
      :error="error"
      :progress="progress"
      :total="total"
      @retry="reload"
    >
      <div class="archive-layout">
        <aside class="archive-sidebar">
          <div class="archive-overview">
            <strong>{{ items.length }}</strong
            ><span>篇公开文章</span>
            <p>按发布日期归档<br />从最新记录开始阅读</p>
          </div>
          <nav aria-label="归档年份">
            <router-link
              :to="location({ year: '', month: '' })"
              :aria-current="!year ? 'page' : undefined"
              >全部年份 <span>{{ items.length }}</span></router-link
            >
            <router-link
              v-for="group in years"
              :key="group.year"
              :to="location({ year: group.year, month: '' })"
              :aria-current="year === group.year ? 'page' : undefined"
              >{{ title(group.year) }}
              <span>{{ group.count }}</span></router-link
            >
          </nav>
          <router-link class="archive-side-link" to="/blog/categories"
            >按主题发现文章 →</router-link
          >
        </aside>
        <section class="archive-content" aria-label="归档时间线">
          <form
            class="archive-search"
            role="search"
            @submit.prevent="router.push(location({ q: keyword.trim() }))"
          >
            <label for="archive-keyword">在归档中查找</label>
            <div>
              <input
                id="archive-keyword"
                v-model="keyword"
                type="search"
                maxlength="200"
                placeholder="输入文章标题"
              /><button type="submit">搜索</button>
            </div>
          </form>
          <nav class="archive-months" aria-label="归档月份">
            <router-link
              :to="location({ month: '' })"
              :aria-current="!month ? 'page' : undefined"
              >全年</router-link
            >
            <router-link
              v-for="item in months.filter(
                (item) => item.count || item.value === month,
              )"
              :key="item.value"
              :to="location({ month: item.value })"
              :aria-current="month === item.value ? 'page' : undefined"
              >{{ Number(item.value) }} 月
              <span>{{ item.count }}</span></router-link
            >
          </nav>
          <div class="archive-results-summary" aria-live="polite">
            <span
              >{{ filtered.length }} 篇文章{{ year ? ` · ${title(year)}` : ""
              }}{{ month ? ` · ${Number(month)} 月` : ""
              }}{{ query ? ` · “${query}”` : "" }}</span
            ><router-link v-if="year || month || query" to="/blog/archives"
              >清除筛选</router-link
            >
          </div>
          <div v-if="!filtered.length" class="archive-empty">
            <h2>
              {{ items.length ? "没有匹配的记录" : "第一篇记录，即将开始" }}
            </h2>
            <p>
              {{
                items.length
                  ? "试试其他年份、月份或关键词。"
                  : "已发布文章会自动出现在归档中。"
              }}
            </p>
            <router-link to="/blog">返回文章列表 →</router-link>
          </div>
          <section
            v-for="group in groups"
            :key="group.year"
            class="archive-year-group"
          >
            <h2>
              {{ title(group.year) }}
              <span
                >{{ pageCount > 1 ? "本页 " : "" }}{{ group.count }} 篇</span
              >
            </h2>
            <section
              v-for="entry in group.months"
              :key="entry.month"
              class="archive-month-group"
            >
              <h3>
                {{
                  entry.month === "undated"
                    ? "未记录日期"
                    : `${Number(entry.month)} 月`
                }}
              </h3>
              <ol>
                <li v-for="article in entry.items" :key="article.id">
                  <time :datetime="article.date || undefined">{{
                    article.date ? article.date.slice(5).replace("-", ".") : "—"
                  }}</time>
                  <div>
                    <router-link :to="`/article/${article.id}`">{{
                      article.title
                    }}</router-link
                    ><router-link
                      v-if="article.category"
                      class="archive-category"
                      :to="{
                        path: '/blog',
                        query: { category: article.category.id },
                      }"
                      >{{ article.category.name }}</router-link
                    >
                  </div>
                  <span aria-hidden="true">↗</span>
                </li>
              </ol>
            </section>
          </section>
          <nav
            v-if="pageCount > 1"
            class="archive-pagination"
            aria-label="归档分页"
          >
            <button
              type="button"
              :disabled="page === 1"
              @click="router.push(location({ page: page - 1 }))"
            >
              上一页</button
            ><span>{{ page }} / {{ pageCount }}</span
            ><button
              type="button"
              :disabled="page === pageCount"
              @click="router.push(location({ page: page + 1 }))"
            >
              下一页
            </button>
          </nav>
        </section>
      </div>
    </BlogIndexState>
  </div>
</template>
<style scoped>
.archive-page {
  padding-bottom: 64px;
}
.index-refresh {
  color: var(--deploy-muted);
  border: 1px solid var(--deploy-line);
  background: var(--deploy-white);
  padding: 8px 14px;
  border-radius: 6px;
  font-size: 12px;
}
.archive-layout {
  display: grid;
  grid-template-columns: 200px minmax(0, 1fr);
  gap: 64px;
}
.archive-sidebar {
  position: sticky;
  top: 24px;
  align-self: start;
}
.archive-overview {
  padding: 20px 0 24px;
  border-bottom: 1px solid var(--deploy-line);
}
.archive-overview strong {
  display: block;
  font-size: 60px;
  font-weight: 500;
  line-height: 1.1;
  letter-spacing: -3px;
}
.archive-overview > span {
  font-size: 12px;
  color: var(--deploy-primary);
}
.archive-overview p {
  color: var(--deploy-faint);
  font-size: 12px;
  line-height: 1.9;
  margin: 16px 0 0;
}
.archive-sidebar nav {
  display: grid;
  gap: 6px;
  margin: 20px 0;
}
.archive-sidebar nav a {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 9px 12px;
  border-radius: 5px;
  font-size: 13px;
  color: var(--deploy-muted);
}
.archive-sidebar nav span {
  font: 11px var(--site-mono);
}
.archive-sidebar a[aria-current] {
  background: var(--deploy-soft);
  color: var(--deploy-primary);
}
.archive-side-link {
  color: var(--deploy-faint);
  font-size: 12px;
  margin-left: 12px;
}
.archive-content {
  min-width: 0;
}
.archive-search {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
}
.archive-search label {
  font-size: 13px;
  font-weight: 500;
}
.archive-search > div {
  display: flex;
  border: 1px solid var(--deploy-line);
  border-radius: 6px;
  padding: 4px;
}
.archive-search input {
  min-width: 0;
  width: 220px;
  background: transparent;
  border: 0;
  padding: 4px 8px;
  font-size: 13px;
  color: var(--deploy-ink);
}
.archive-search button {
  background: var(--deploy-primary);
  color: var(--deploy-white);
  border: 0;
  border-radius: 4px;
  padding: 5px 14px;
  font-size: 12px;
  white-space: nowrap;
}
.archive-months {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin: 22px 0 18px;
}
.archive-months a {
  padding: 5px 10px;
  font-size: 12px;
  color: var(--deploy-muted);
  border: 1px solid transparent;
  border-radius: 5px;
}
.archive-months span {
  margin-left: 5px;
  font-size: 10px;
  color: var(--deploy-faint);
}
.archive-months a[aria-current] {
  border-color: var(--deploy-line);
  background: var(--deploy-paper);
  color: var(--deploy-primary);
}
.archive-results-summary {
  display: flex;
  gap: 16px;
  justify-content: space-between;
  font-size: 12px;
  color: var(--deploy-faint);
  border-bottom: 1px solid var(--deploy-line);
  padding-bottom: 20px;
  overflow-wrap: anywhere;
}
.archive-results-summary a {
  color: var(--deploy-primary);
  white-space: nowrap;
}
.archive-year-group > h2 {
  display: flex;
  gap: 16px;
  align-items: baseline;
  font-size: 28px;
  margin: 30px 0 24px;
  font-weight: 600;
}
.archive-year-group > h2 span {
  font-size: 11px;
  color: var(--deploy-faint);
  font-weight: 400;
}
.archive-month-group {
  display: grid;
  grid-template-columns: 50px minmax(0, 1fr);
  gap: 24px;
}
.archive-month-group h3 {
  font-size: 12px;
  color: var(--deploy-faint);
  font-weight: 400;
  padding-top: 14px;
  margin: 0;
}
ol {
  list-style: none;
  margin: 0;
  padding: 0 0 20px 24px;
  border-left: 1px solid var(--deploy-line);
}
li {
  display: flex;
  align-items: baseline;
  gap: 20px;
  position: relative;
  padding: 14px 0;
  border-bottom: 1px dashed var(--deploy-line);
}
li::before {
  content: "";
  position: absolute;
  width: 5px;
  height: 5px;
  border-radius: 50%;
  background: var(--deploy-primary);
  left: -27px;
  top: 24px;
}
time {
  color: var(--deploy-faint);
  font: 12px var(--site-mono);
  white-space: nowrap;
}
li > div {
  flex: 1;
  min-width: 0;
}
li > div > a:first-child {
  font-size: 15px;
  overflow-wrap: anywhere;
  display: block;
}
li a:hover {
  color: var(--deploy-primary);
}
.archive-category {
  display: inline-block;
  font-size: 11px;
  color: var(--deploy-faint);
  margin-top: 4px;
}
li > span {
  color: var(--deploy-faint);
  font-size: 12px;
}
.archive-empty {
  text-align: center;
  padding: 50px 12px;
}
.archive-empty h2 {
  font-size: 21px;
}
.archive-empty p {
  color: var(--deploy-muted);
  font-size: 13px;
}
.archive-empty a {
  color: var(--deploy-primary);
  font-size: 13px;
}
.archive-pagination {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 24px;
  font-size: 12px;
  margin-top: 24px;
}
.archive-pagination button {
  border: 1px solid var(--deploy-line);
  background: var(--deploy-white);
  color: var(--deploy-muted);
  padding: 8px 16px;
  border-radius: 5px;
}
@media (max-width: 760px) {
  .archive-layout {
    grid-template-columns: minmax(0, 1fr);
    gap: 24px;
  }
  .archive-sidebar {
    position: static;
  }
  .archive-overview {
    display: none;
  }
  .archive-sidebar nav {
    display: flex;
    flex-wrap: wrap;
    margin: 0;
    gap: 8px;
  }
  .archive-sidebar nav a {
    gap: 12px;
    border: 1px solid var(--deploy-line);
  }
  .archive-side-link {
    display: none;
  }
  .archive-search {
    flex-wrap: wrap;
    gap: 10px;
  }
  .archive-search > div {
    flex: 1;
    min-width: 220px;
  }
  .archive-search input {
    width: 100%;
  }
  .archive-month-group {
    grid-template-columns: 32px minmax(0, 1fr);
    gap: 12px;
  }
  ol {
    padding-left: 16px;
  }
  li {
    gap: 10px;
  }
  li::before {
    left: -19px;
  }
  li > div > a:first-child {
    font-size: 14px;
  }
}
</style>
