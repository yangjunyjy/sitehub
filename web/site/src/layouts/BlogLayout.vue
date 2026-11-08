<script setup>
import { computed, watch, ref, nextTick } from "vue";
import { useRoute, useRouter } from "vue-router";
import BrandMark from "@/components/product/BrandMark.vue";
import ThemeToggle from "@/components/product/ThemeToggle.vue";
const route = useRoute(),
  router = useRouter(),
  search = ref("");
const navigation = [
  { to: "/blog", label: "文章", code: "01" },
  { to: "/blog/archives", label: "归档", code: "02" },
  { to: "/blog/categories", label: "分类", code: "03" },
  { to: "/blog/tags", label: "标签", code: "04" },
  { to: "/blog/reading", label: "阅读空间", code: "05" },
];
const current = computed(() =>
  route.path.startsWith("/article/") ? "/blog" : route.path,
);
const links = ref();
watch(
  () => route.fullPath,
  async () => {
    search.value =
      typeof route.query.q === "string" && route.path === "/blog"
        ? route.query.q
        : "";
    await nextTick();
    const active = links.value?.querySelector('[aria-current="page"]');
    if (active) {
      const rect = active.getBoundingClientRect(),
        nav = links.value.getBoundingClientRect();
      if (rect.left < nav.left) links.value.scrollLeft += rect.left - nav.left;
      if (rect.right > nav.right)
        links.value.scrollLeft += rect.right - nav.right;
    }
  },
  { immediate: true },
);
function submit() {
  router.push({
    path: "/blog",
    query: search.value.trim() ? { q: search.value.trim() } : {},
  });
}
</script>
<template>
  <div class="blog-app">
    <header class="journal-header">
      <div class="site-shell journal-masthead">
        <router-link
          to="/blog"
          class="journal-brand"
          aria-label="Deploy Blog 首页"
        >
          <BrandMark /><span>Deploy <b>Blog</b></span
          ><small>工程与思考</small>
        </router-link>
        <div class="journal-actions">
          <form role="search" aria-label="博客搜索" @submit.prevent="submit">
            <input
              v-model="search"
              aria-label="搜索博客"
              placeholder="搜索文章…"
              maxlength="200"
              type="search"
            />
            <button type="submit" aria-label="提交博客搜索">
              <svg viewBox="0 0 24 24" aria-hidden="true">
                <circle cx="10" cy="10" r="6" />
                <path d="m15 15 5 5" />
              </svg>
            </button>
          </form>
          <ThemeToggle />
          <router-link to="/" class="journal-website"
            >返回官网 <span>↗</span></router-link
          >
        </div>
      </div>
      <div class="journal-nav-wrap">
        <nav ref="links" class="site-shell journal-nav" aria-label="博客导航">
          <router-link
            v-for="item in navigation"
            :key="item.to"
            :to="item.to"
            :aria-current="current === item.to ? 'page' : undefined"
            ><small>{{ item.code }}</small
            >{{ item.label }}</router-link
          >
        </nav>
      </div>
    </header>
    <main id="main-content" class="journal-main" tabindex="-1"><slot /></main>
    <footer class="journal-footer site-shell">
      <div>
        <router-link to="/blog"
          >Deploy Blog<span>记录、连接、持续构建。</span></router-link
        >
        <p>每一篇实践，都值得被找到。</p>
      </div>
      <nav aria-label="博客页脚">
        <router-link to="/blog/archives">文章归档</router-link
        ><router-link to="/blog/reading">阅读空间</router-link
        ><router-link to="/">Deploy 官网 ↗</router-link>
      </nav>
    </footer>
  </div>
</template>
<style scoped>
.blog-app {
  min-height: 100dvh;
  display: flex;
  flex-direction: column;
  --site-shell: min(1160px, calc(100% - 64px));
}
.journal-header {
  background: var(--header-bg);
  border-bottom: 1px solid var(--deploy-line);
}
.journal-masthead {
  min-height: 86px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 24px;
}
.journal-brand {
  display: flex;
  align-items: center;
  gap: 10px;
  white-space: nowrap;
}
.journal-brand :deep(svg) {
  width: 31px;
  height: 31px;
}
.journal-brand > span {
  font-size: 23px;
  font-weight: 700;
  letter-spacing: -1px;
}
.journal-brand b {
  color: var(--deploy-primary);
  font-weight: 500;
}
.journal-brand small {
  border-left: 1px solid var(--deploy-line);
  padding-left: 16px;
  margin-left: 10px;
  color: var(--deploy-faint);
  font-size: 12px;
}
.journal-actions {
  display: flex;
  gap: 22px;
  align-items: center;
}
.journal-actions form {
  display: flex;
  align-items: center;
  width: 210px;
  border: 1px solid var(--deploy-line);
  border-radius: 6px;
  padding: 4px 6px 4px 12px;
  background: var(--deploy-paper);
}
.journal-actions input {
  min-width: 0;
  width: 100%;
  border: 0;
  outline: 0;
  background: transparent;
  color: var(--deploy-ink);
  font-size: 12px;
}
.journal-actions form:focus-within {
  outline: 2px solid var(--deploy-primary);
  outline-offset: 2px;
}
.journal-actions form button {
  border: 0;
  background: transparent;
  color: var(--deploy-muted);
  padding: 6px;
}
.journal-actions svg {
  width: 16px;
  height: 16px;
  fill: none;
  stroke: currentColor;
  stroke-width: 1.5;
}
.journal-website {
  font-size: 12px;
  color: var(--deploy-muted);
  white-space: nowrap;
}
.journal-website span {
  margin-left: 8px;
}
.journal-nav-wrap {
  border-top: 1px solid var(--deploy-line);
}
.journal-nav {
  display: flex;
  gap: 36px;
  overflow-x: auto;
  scrollbar-width: none;
}
.journal-nav a {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 14px 0 12px;
  border-bottom: 2px solid transparent;
  color: var(--deploy-muted);
  white-space: nowrap;
  font-size: 14px;
}
.journal-nav small {
  font: 10px var(--site-mono);
  color: var(--deploy-faint);
}
.journal-nav a[aria-current] {
  border-color: var(--deploy-primary);
  color: var(--deploy-primary);
  font-weight: 600;
}
.journal-main {
  flex: 1;
  min-width: 0;
  outline: none;
}
.journal-footer {
  border-top: 1px solid var(--deploy-line);
  padding: 32px 0 40px;
  display: flex;
  justify-content: space-between;
  gap: 30px;
}
.journal-footer > div > a {
  font-size: 15px;
  font-weight: 600;
}
.journal-footer span {
  margin-left: 16px;
  font-size: 12px;
  font-weight: 400;
  color: var(--deploy-muted);
}
.journal-footer p {
  font-size: 11px;
  color: var(--deploy-faint);
  margin: 8px 0 0;
}
.journal-footer nav {
  display: flex;
  gap: 24px;
  font-size: 12px;
  color: var(--deploy-muted);
}
@media (max-width: 760px) {
  .blog-app {
    --site-shell: calc(100% - 40px);
  }
  .journal-masthead {
    min-height: 72px;
    gap: 12px;
  }
  .journal-brand > span {
    font-size: 21px;
  }
  .journal-brand small,
  .journal-actions form {
    display: none;
  }
  .journal-actions {
    gap: 14px;
  }
  .journal-nav {
    gap: 26px;
  }
  .journal-nav a {
    font-size: 13px;
    padding-block: 12px;
  }
  .journal-nav small {
    display: none;
  }
  .journal-footer {
    flex-direction: column;
    gap: 20px;
  }
}
@media (max-width: 390px) {
  .journal-brand :deep(svg) {
    width: 25px;
    height: 25px;
  }
  .journal-website {
    font-size: 11px;
  }
  .journal-actions {
    gap: 10px;
  }
  .journal-nav {
    gap: 23px;
  }
}
</style>
