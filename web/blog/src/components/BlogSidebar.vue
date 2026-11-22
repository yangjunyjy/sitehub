<script setup>
import ResourceIcon from "@shared/icons/ResourceIcon.vue";
import { onMounted } from "vue";
import { useResource } from "@/composables/useResource";
import { about, categories, tags } from "@/api";
import { safeURL } from "@/utils/format";
import BlogIcon from "./BlogIcon.vue";
const profile = useResource(about);
const categoryData = useResource(categories, []);
const tagData = useResource(tags, []);
onMounted(() => {
  profile.load();
  categoryData.load();
  tagData.load();
});
</script>
<template>
  <aside class="blog-sidebar">
    <section class="profile-card">
      <div class="profile-avatar">
        <img
          v-if="safeURL(profile.data.value?.avatar)"
          :src="safeURL(profile.data.value.avatar)"
          alt="博主头像"
        /><BlogIcon v-else name="feather" :size="32" />
      </div>
      <h3>{{ profile.data.value?.name || "云间小记" }}</h3>
      <p v-if="profile.data.value?.title">{{ profile.data.value.title }}</p>
      <router-link :to="{ name: 'about' }">关于这里 <span>↗</span></router-link>
    </section>
    <section class="sidebar-section">
      <h3><BlogIcon name="folder" :size="15" /> 文章分类</h3>
      <p v-if="categoryData.error.value" class="sidebar-hint">
        分类暂时无法加载 <button @click="categoryData.load()">重试</button>
      </p>
      <p
        v-else-if="
          !categoryData.loading.value && !categoryData.data.value?.length
        "
        class="sidebar-hint"
      >
        暂未添加分类
      </p>
      <router-link
        v-for="category in (categoryData.data.value || []).slice(0, 8)"
        :key="category.id"
        class="category-link"
        :to="{ name: 'categories', query: { category: String(category.id) } }"
        ><span
          ><ResourceIcon
            v-if="category.icon"
            :name="category.icon"
            :size="15"
          />
          {{ category.name }}</span
        ><span>{{
          category.articleCount ?? category.article_count ?? 0
        }}</span></router-link
      >
    </section>
    <section class="sidebar-section">
      <h3><BlogIcon name="tag" :size="15" /> 标签云</h3>
      <p v-if="tagData.error.value" class="sidebar-hint">
        标签暂时无法加载 <button @click="tagData.load()">重试</button>
      </p>
      <div class="tag-cloud">
        <router-link
          v-for="tag in (tagData.data.value || []).slice(0, 18)"
          :key="tag.id"
          :to="{ name: 'tags', query: { tags: String(tag.id) } }"
          >{{ tag.name }} ·
          {{ tag.articleCount ?? tag.article_count ?? 0 }}</router-link
        >
      </div>
      <p
        v-if="!tagData.loading.value && !tagData.data.value?.length"
        class="sidebar-hint"
      >
        暂未添加标签
      </p>
    </section>
  </aside>
</template>
<style lang="scss" scoped>
.blog-sidebar {
  flex: 0 0 clamp(195px, 19vw, 248px);
  display: grid;
  gap: 24px;
}
.profile-card {
  display: grid;
  grid-template-columns: auto 1fr;
  align-items: center;
  column-gap: 13px;
  padding: 14px 0 18px;
  border-block: 1px solid var(--journal-rule);
}
.profile-avatar {
  grid-row: 1 / span 3;
  width: 48px;
  height: 48px;
  display: grid;
  place-items: center;
  border-radius: 50%;
  background: #e9e2d9;
  color: #68574d;
  overflow: hidden;
}
.profile-avatar img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}
.profile-card h3 {
  margin: 0;
  font: 600 18px/1.5 var(--journal-serif);
}
.profile-card p {
  margin: 4px 0;
  font-size: 12px;
  color: var(--journal-muted);
}
.profile-card > a {
  font-size: 12px;
  color: var(--journal-accent);
}
.sidebar-section h3 {
  margin: 0 0 8px;
  display: flex;
  gap: 7px;
  align-items: center;
  font-size: 14px;
}
.category-link {
  display: flex;
  justify-content: space-between;
  padding: 6px 0;
  border-bottom: 1px solid var(--journal-rule);
  font-size: 13px;
}
.category-link span:first-child {
  display: inline-flex;
  gap: 6px;
  align-items: center;
}
.tag-cloud {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}
.tag-cloud a {
  border: 1px solid var(--journal-rule);
  border-radius: 3px;
  padding: 2px 7px;
  font-size: 12px;
  color: var(--journal-muted);
}
a:hover,
.sidebar-hint button {
  color: var(--journal-accent);
}
.sidebar-hint {
  color: var(--journal-muted);
  font-size: 12px;
}
.sidebar-hint button {
  border: 0;
  background: transparent;
}
@media (max-width: 850px) {
  .blog-sidebar {
    display: none;
  }
}
</style>
