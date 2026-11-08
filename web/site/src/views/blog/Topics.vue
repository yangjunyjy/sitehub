<script setup>
import { computed, ref, watch } from "vue";
import { useRoute, useRouter } from "vue-router";
import { useBlogIndex } from "@/composables/useBlogIndex";
import { blogTopics } from "@/utils/blog-index";
import BlogPageHeading from "@/components/blog/BlogPageHeading.vue";
import BlogIndexState from "@/components/blog/BlogIndexState.vue";
const props = defineProps({ kind: { type: String, required: true } });
const { items, loading, error, progress, total, reload } = useBlogIndex();
const route = useRoute(),
  router = useRouter(),
  keyword = ref("");
watch(
  () => route.query.q,
  (value) => (keyword.value = typeof value === "string" ? value : ""),
  { immediate: true },
);
const categories = computed(() => props.kind === "categories");
const name = computed(() => (categories.value ? "分类" : "标签"));
const topics = computed(() => blogTopics(items.value, props.kind));
const filtered = computed(() =>
  topics.value.filter((item) =>
    item.name
      .toLocaleLowerCase()
      .includes(String(route.query.q || "").toLocaleLowerCase()),
  ),
);
const toTopic = (item) => ({
  path: "/blog",
  query: { [categories.value ? "category" : "tags"]: item.id },
});
const unclassified = computed(
  () => items.value.filter((item) => !item.category).length,
);
function search() {
  router.push({
    path: route.path,
    query: keyword.value.trim() ? { q: keyword.value.trim() } : {},
  });
}
</script>
<template>
  <div class="site-shell topics-page">
    <BlogPageHeading
      :eyebrow="categories ? 'COLLECTIONS / 按主题阅读' : 'TOPICS / 发现关联'"
      :title="`文章${name}`"
      :description="
        categories
          ? '从一个感兴趣的主题，开始系统地阅读。'
          : '技术、工具与方法，让零散的记录重新连接。'
      "
    />
    <BlogIndexState
      :loading="loading"
      :error="error"
      :progress="progress"
      :total="total"
      @retry="reload"
    >
      <div class="topics-toolbar">
        <p>
          {{ topics.length }} 个{{ name }}
          <span
            >· {{ items.length }} 篇文章{{
              categories && unclassified ? ` · ${unclassified} 篇未分类` : ""
            }}</span
          >
        </p>
        <form role="search" @submit.prevent="search">
          <input
            v-model="keyword"
            :aria-label="`查找${name}`"
            :placeholder="`查找${name}…`"
            type="search"
            maxlength="100"
          /><button type="submit">查找</button>
        </form>
      </div>
      <div v-if="categories && filtered.length" class="category-collections">
        <router-link
          v-for="(item, index) in filtered"
          :key="item.id"
          :to="toTopic(item)"
          class="category-collection"
          ><div>
            <span>{{ String(index + 1).padStart(2, "0") }}</span
            ><small>{{ item.count }} 篇文章</small>
          </div>
          <h2>{{ item.name }}</h2>
          <p>最近记录</p>
          <h3>{{ item.latest }}</h3>
          <footer>浏览专题<span>↗</span></footer></router-link
        >
      </div>
      <div
        v-else-if="!categories && filtered.length"
        class="topic-cloud"
        aria-label="标签索引"
      >
        <router-link
          v-for="item in filtered"
          :key="item.id"
          :to="toTopic(item)"
          :style="{
            '--topic-weight':
              Math.min(24, 14 + Math.log2(item.count + 1) * 2) + 'px',
          }"
          ><span>#</span>{{ item.name
          }}<small>{{ item.count }}</small></router-link
        >
      </div>
      <div v-else class="topics-empty">
        <h2>
          {{
            route.query.q ? `没有找到这个${name}` : `还没有公开的${name}文章`
          }}
        </h2>
        <p>仅统计已发布文章使用的{{ name }}。</p>
        <router-link :to="route.query.q ? route.path : '/blog'"
          >{{
            route.query.q ? "查看全部" + name : "返回文章列表"
          }}
          →</router-link
        >
      </div>
      <div class="topics-note">
        <span>换一种方式发现内容</span
        ><router-link to="/blog/archives">浏览时间归档 →</router-link
        ><router-link :to="categories ? '/blog/tags' : '/blog/categories'"
          >{{ categories ? "探索全部标签" : "探索文章分类" }} →</router-link
        >
      </div>
    </BlogIndexState>
  </div>
</template>
<style scoped>
.topics-page {
  padding-bottom: 64px;
}
.topics-toolbar {
  display: flex;
  justify-content: space-between;
  gap: 20px;
  align-items: center;
  margin-bottom: 30px;
}
.topics-toolbar p {
  font-size: 14px;
  margin: 0;
}
.topics-toolbar p span {
  color: var(--deploy-faint);
  font-size: 12px;
  margin-left: 8px;
}
form {
  display: flex;
  border: 1px solid var(--deploy-line);
  border-radius: 6px;
  padding: 4px;
}
input {
  border: 0;
  background: transparent;
  padding: 5px 10px;
  width: 180px;
  min-width: 0;
  color: var(--deploy-ink);
  font-size: 13px;
}
button {
  background: var(--deploy-paper);
  color: var(--deploy-muted);
  border: 0;
  border-radius: 4px;
  padding: 4px 12px;
  font-size: 12px;
  white-space: nowrap;
}
.category-collections {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 22px;
}
.category-collection {
  border: 1px solid var(--deploy-line);
  padding: 26px;
  border-radius: 8px;
  display: flex;
  flex-direction: column;
  transition:
    border-color 0.15s,
    transform 0.15s;
}
.category-collection:hover {
  border-color: var(--deploy-primary);
  transform: translateY(-3px);
}
.category-collection > div {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.category-collection > div > span {
  font: 22px var(--site-mono);
  color: var(--deploy-primary);
}
.category-collection small {
  font-size: 11px;
  color: var(--deploy-faint);
}
h2 {
  font-size: 23px;
  margin: 28px 0 24px;
  overflow-wrap: anywhere;
}
.category-collection p {
  font-size: 10px;
  color: var(--deploy-faint);
  margin: 0 0 6px;
}
h3 {
  font-size: 13px;
  font-weight: 400;
  margin: 0 0 28px;
  color: var(--deploy-muted);
  overflow-wrap: anywhere;
}
.category-collection footer {
  border-top: 1px solid var(--deploy-line);
  padding-top: 16px;
  margin-top: auto;
  display: flex;
  justify-content: space-between;
  color: var(--deploy-primary);
  font-size: 12px;
}
.topic-cloud {
  display: flex;
  flex-wrap: wrap;
  gap: 16px;
  padding: 30px 0 50px;
  align-items: center;
}
.topic-cloud a {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 14px 22px;
  border: 1px solid var(--deploy-line);
  border-radius: 7px;
  font-size: var(--topic-weight);
  max-width: 100%;
  overflow-wrap: anywhere;
}
.topic-cloud a:hover {
  border-color: var(--deploy-primary);
  background: var(--deploy-soft);
}
.topic-cloud a > span {
  color: var(--deploy-primary);
  font-size: 14px;
}
.topic-cloud small {
  font: 11px var(--site-mono);
  color: var(--deploy-faint);
}
.topics-note {
  display: flex;
  flex-wrap: wrap;
  gap: 12px 28px;
  border-top: 1px solid var(--deploy-line);
  padding-top: 24px;
  margin-top: 48px;
  font-size: 12px;
  color: var(--deploy-primary);
}
.topics-note > span {
  margin-right: auto;
  color: var(--deploy-faint);
}
.topics-empty {
  text-align: center;
  padding: 50px 12px;
}
.topics-empty p {
  font-size: 13px;
  color: var(--deploy-muted);
}
.topics-empty a {
  color: var(--deploy-primary);
  font-size: 13px;
}
@media (max-width: 900px) {
  .category-collections {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}
@media (max-width: 600px) {
  .category-collections {
    grid-template-columns: minmax(0, 1fr);
  }
  .topics-toolbar {
    flex-wrap: wrap;
  }
  .topics-toolbar form {
    width: 100%;
  }
  input {
    width: 100%;
  }
  .topic-cloud {
    gap: 10px;
    padding-top: 12px;
  }
  .topic-cloud a {
    padding: 10px 14px;
  }
}
@media (prefers-reduced-motion: reduce) {
  .category-collection {
    transition: none;
  }
  .category-collection:hover {
    transform: none;
  }
}
</style>
