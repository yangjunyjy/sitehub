<script setup>
import { onMounted } from "vue";
import ContentState from "@/components/ContentState.vue";
import ArticleList from "@/components/ArticleList.vue";
import AboutCard from "./components/AboutCard.vue";
import { useResource } from "@/composables/useResource";
import { about, articles } from "@/api";
const profile = useResource(about);
const posts = useResource(() =>
  articles({ page: 1, pageSize: 3, sort: "latest" }),
);
onMounted(() => {
  profile.load();
  posts.load();
});
</script>
<template>
  <section class="about-page" aria-label="关于">
    <div class="about-shell">
      <ContentState
        v-if="profile.loading.value || profile.error.value"
        :loading="profile.loading.value"
        :error="profile.error.value"
        @retry="profile.load()"
      />
      <AboutCard v-else :profile="profile.data.value || {}" />
      <section
        class="about-writing"
        v-if="posts.data.value?.items?.length || posts.error.value"
      >
        <header>
          <h2>最近的文章</h2>
          <RouterLink :to="{ name: 'home', hash: '#articles' }"
            >全部文章 ↗</RouterLink
          >
        </header>
        <ContentState
          v-if="posts.error.value"
          :error="posts.error.value"
          @retry="posts.load()"
        />
        <ArticleList v-else :items="posts.data.value.items" />
      </section>
    </div>
  </section>
</template>
<style scoped>
.about-page {
  background: var(--journal-paper);
  color: var(--journal-ink);
  padding: clamp(92px, 8vw, 120px) 20px 64px;
}
.about-shell {
  max-width: 960px;
  margin: auto;
}
.about-writing {
  margin-top: 36px;
}
.about-writing header {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  gap: 12px;
  border-bottom: 2px solid var(--journal-ink);
  padding-bottom: 10px;
}
h2 {
  font: 600 23px/1.3 var(--journal-serif);
  margin: 0;
}
.about-writing header a {
  color: var(--journal-accent);
  font-size: 13px;
}
</style>
