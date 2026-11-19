<script setup>
import { ref, watch } from "vue";
import { useRoute, useRouter } from "vue-router";
import BlogIcon from "@/components/BlogIcon.vue";
const route = useRoute(),
  router = useRouter(),
  keyword = ref("");
watch(
  () => route.query.q,
  (value) => {
    keyword.value = String(value || "");
  },
  { immediate: true },
);
const submit = () =>
  router.push({
    path: "/search",
    query: { q: keyword.value.trim().slice(0, 200) || undefined },
  });
</script>
<template>
  <form class="search-form" role="search" @submit.prevent="submit">
    <BlogIcon name="search" :size="19" /><input
      v-model="keyword"
      maxlength="200"
      aria-label="搜索关键词"
      placeholder="想找点什么？输入关键词…"
    /><button type="submit">搜索 ↗</button>
  </form>
</template>
<style lang="scss" scoped>
.search-form {
  max-width: 620px;
  margin: 35px auto 0;
  display: flex;
  align-items: center;
  gap: 17px;
  padding: 8px 12px 8px 22px;
  border-radius: 5px;
  background: #fffffff0;
  color: #9d91b0;
  box-shadow: 0 5px 35px #5a477513;
  input {
    flex: 1;
    min-width: 0;
    background: none;
    border: 0;
    outline: none;
    font-size: 13px;
    padding: 10px 0;
    color: var(--ink);
  }
  button {
    border: 0;
    background: #ab99be;
    color: white;
    border-radius: 3px;
    font-size: 11px;
    padding: 11px 18px;
  }
}
</style>
