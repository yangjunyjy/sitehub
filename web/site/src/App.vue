<script setup>
import { computed, watch, defineAsyncComponent } from "vue";
import { useRoute, useRouter } from "vue-router";
import Header from "@/components/Header.vue";
import Footer from "@/components/Footer.vue";
import SiteNavigationStatus from "@/components/product/SiteNavigationStatus.vue";
import ContentState from "@/components/product/ContentState.vue";
import { useWebsiteContent } from "@/composables/useWebsiteContent";
const route = useRoute();
const router = useRouter();
// The first async route has not resolved when App mounts. Choose its layout now,
// so opening the blog never briefly mounts product navigation or its API requests.
const initialRoute = router.resolve(router.options.history.location);
const pageMeta = computed(() =>
  route.matched.length ? route.meta : initialRoute.meta,
);
const hideLayout = computed(() => Boolean(pageMeta.value.hideLayout));
const isBlog = computed(() => pageMeta.value.layout === "blog");
const BlogLayout = defineAsyncComponent(
  () => import("@/layouts/BlogLayout.vue"),
);
const { website, loadWebsiteContent } = useWebsiteContent();
const retry = () => loadWebsiteContent(true).catch(() => {});
watch(
  isBlog,
  (blog) => {
    if (!blog) loadWebsiteContent().catch(() => {});
  },
  { immediate: true },
);
</script>

<template>
  <div id="app" :class="{ 'site-docs': route.path.startsWith('/docs') }">
    <SiteNavigationStatus />
    <a class="skip-link" href="#main-content">跳至主要内容</a>
    <BlogLayout v-if="isBlog"><router-view /></BlogLayout>
    <Header v-if="!hideLayout && !isBlog" />
    <main v-if="!isBlog" id="main-content" class="app-main" tabindex="-1">
      <router-view v-if="hideLayout" />
      <ContentState
        v-else-if="website.error && !website.loaded"
        class="site-shell site-status"
        error
        title="暂时无法加载页面"
        message="连接未能完成，请稍后重试。"
        action-label="重新加载"
        :loading="website.loading"
        @retry="retry"
      />
      <div
        v-else-if="!website.loaded"
        class="site-shell site-loading"
        role="status"
        aria-label="正在加载页面"
      >
        <span class="site-kicker">Deploy / 正在加载</span><i></i><i></i><i></i>
        <div></div>
      </div>
      <router-view v-else />
    </main>
    <Footer v-if="!hideLayout && !isBlog && !route.path.startsWith('/docs')" />
  </div>
</template>

<style>
#app {
  display: flex;
  flex-direction: column;
  min-height: 100vh;
}
.app-main {
  flex: 1;
}
</style>
