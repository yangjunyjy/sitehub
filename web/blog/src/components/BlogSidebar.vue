<script setup>
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
    <section class="profile-card glass-panel">
      <div class="profile-avatar">
        <img
          v-if="safeURL(profile.data.value?.avatar)"
          :src="safeURL(profile.data.value.avatar)"
          alt="博主头像"
        /><BlogIcon v-else name="feather" :size="32" />
      </div>
      <h3>{{ profile.data.value?.name || "云间小记" }}</h3>
      <p v-if="profile.data.value?.title">{{ profile.data.value.title }}</p>
      <div class="profile-card__line"></div>
      <router-link to="/about">关于这里 <span>↗</span></router-link>
    </section>
    <section class="sidebar-section glass-panel">
      <h3>
        <BlogIcon name="folder" :size="15" /> 文章分类
      </h3>
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
        :to="{ path: '/categories', query: { category: String(category.id) } }"
        ><span>{{ category.name }}</span
        ><span>{{ category.articleCount ?? category.article_count ?? 0 }}</span></router-link
      >
    </section>
    <section class="sidebar-section glass-panel">
      <h3><BlogIcon name="tag" :size="15" /> 标签云</h3>
      <p v-if="tagData.error.value" class="sidebar-hint">
        标签暂时无法加载 <button @click="tagData.load()">重试</button>
      </p>
      <div class="tag-cloud">
        <router-link
          v-for="tag in (tagData.data.value || []).slice(0, 18)"
          :key="tag.id"
          :to="{ path: '/tags', query: { tags: String(tag.id) } }"
          >{{ tag.name }} · {{ tag.articleCount ?? tag.article_count ?? 0 }}</router-link
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
  width: 245px;
  flex-shrink: 0;
  display: grid;
  gap: 16px;
}
.profile-card {
  text-align: center;
  padding: 30px 20px;
  border: 1px solid var(--line);
  border-radius: 6px;
  h3 {
    font-size: 18px;
    font-weight: 500;
    margin: 17px 0 9px;
  }
  p {
    font-size: 12px;
    color: var(--muted);
    line-height: 1.9;
  }
  &__line {
    height: 1px;
    background: var(--line);
    margin: 24px 15px 18px;
  }
  > a {
    font-size: 11px;
    letter-spacing: 1px;
    span {
      margin-left: 15px;
    }
  }
}
.profile-avatar {
  display: grid;
  place-items: center;
  width: 72px;
  height: 72px;
  margin: auto;
  border-radius: 50%;
  background: linear-gradient(135deg, #b9c7e0, #edd2df);
  color: white;
  overflow: hidden;
  img {
    width: 100%;
    height: 100%;
    object-fit: cover;
  }
}
.sidebar-section {
  padding: 20px;
  h3 {
    font-size: 13px;
    font-weight: 500;
    display: flex;
    align-items: center;
    gap: 8px;
    margin: 0 0 18px;
    small {
      margin-left: auto;
      font-size: 8px;
      letter-spacing: 1.2px;
      color: var(--muted);
    }
  }
}
.category-link {
  display: flex;
  justify-content: space-between;
  font-size: 12px;
  padding: 11px 0;
  border-bottom: 1px dashed var(--line);
  color: #777581;
  &:hover {
    color: var(--accent);
  }
}
.tag-cloud {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  a {
    font-size: 10px;
    padding: 5px 9px;
    background: #f5f4f9;
    border-radius: 3px;
    color: #8d859f;
    &:nth-child(3n) {
      background: #f8f1ed;
      color: #b39a88;
    }
    &:nth-child(3n + 1) {
      background: #eef5f4;
      color: #83a29d;
    }
  }
}
.sidebar-hint {
  font-size: 11px;
  color: var(--muted);
  button {
    color: var(--accent);
    border: 0;
    background: none;
  }
}
@media (max-width: 850px) {
  .blog-sidebar {
    display: none;
  }
}
</style>
