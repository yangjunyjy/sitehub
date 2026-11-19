<script setup>
import { computed, onMounted } from "vue";
import { useRoute, useRouter } from "vue-router";
import { categories, tags } from "@/api";
import { useResource } from "@/composables/useResource";
const props = defineProps({ kind: String });
const route = useRoute(),
  router = useRouter();
const { data, loading, error, load } = useResource(
  props.kind === "categories" ? categories : tags,
  [],
);
const key = computed(() => (props.kind === "categories" ? "category" : "tags"));
const current = computed(() => String(route.query[key.value] || ""));
function choose(id) {
  router.push({
    path: route.path,
    query: { [key.value]: id ? String(id) : undefined },
    hash: "#topic-list",
  });
}
onMounted(load);
</script>
<template>
  <div id="topic-list" class="topic-selector content-shell glass-panel">
    <h2 class="topic-selector__caption">{{ kind === "categories" ? "文章分类" : "文章标签" }} <small v-if="!loading && !error">{{ data?.length || 0 }} 个</small></h2>
    <div v-if="loading" class="topic-selector__hint" role="status">
      正在加载…
    </div>
    <p v-else-if="error" class="topic-selector__hint">
      {{ error }} <button class="text-button" @click="load">重试</button>
    </p>
    <div v-else class="topic-selector__chips">
      <button :aria-pressed="!current" :class="{ active: !current }" @click="choose('')">
        全部{{ kind === "categories" ? "分类" : "标签" }}</button
      ><button
        v-for="item in data"
        :key="item.id"
        :aria-pressed="current === String(item.id)" :class="{ active: current === String(item.id) }"
        @click="choose(item.id)"
      >
        <span v-if="kind === 'tags'"># </span>{{ item.name }}
        <small class="topic-count">{{ item.articleCount ?? item.article_count ?? 0 }}</small>
      </button>
    </div>
  </div>
</template>
<style lang="scss" scoped>
.topic-selector {
  margin-top: 32px;
  padding: 28px 20px;
  border-bottom: 1px solid var(--line);
  text-align: center;
  &__caption {
    font-size: 19px;
    color: var(--ink);
    small { font-size: 12px; font-weight: 400; color: var(--muted); }
    letter-spacing: 1px;
    margin: 0 0 25px;
  }
  &__chips {
    max-width: 800px;
    margin: auto;
    display: flex;
    justify-content: center;
    flex-wrap: wrap;
    gap: 10px;
    button {
      .topic-count { border-inline-start: 1px solid currentColor; margin-left: 8px; padding-left: 8px; }
      font-size: 12px;
      padding: 8px 18px;
      border: 1px solid #e9e4ef;
      background: white;
      color: #91899d;
      border-radius: 30px;
      transition: 0.2s;
      &.active,
      &:hover {
        background: var(--accent);
        color: white;
        border-color: #a394bb;
      }
    }
  }
  &__hint {
    font-size: 12px;
    color: var(--muted);
  }
}
</style>
