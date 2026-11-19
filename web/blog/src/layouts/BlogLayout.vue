<script setup>
import LiquidGlass from "@/components/inspira/LiquidGlass.vue";
import { computed, onBeforeUnmount, onMounted, ref, watch } from "vue";
import { useRoute, useRouter } from "vue-router";
import BlogIcon from "@/components/BlogIcon.vue";
import { useMember } from "@/composables/useMember";
import { useResource } from "@/composables/useResource";
import { publicMenus } from "@/api";
import { navigationLinks } from "@/utils/navigation.mjs";
const route = useRoute(),
  menu = ref(false),
  scrolled = ref(false);
const { member, restore } = useMember();
const navigation = useResource(publicMenus);
const router = useRouter();
const links = computed(() =>
  navigationLinks(navigation.data.value?.items, router.getRoutes()),
);
const solid = computed(() => scrolled.value || menu.value);
const onScroll = () => {
  scrolled.value = window.scrollY > 55;
};
const scrollToTop = () => window.scrollTo({ top: 0, behavior: "smooth" });
watch(
  () => route.fullPath,
  () => {
    menu.value = false;
  },
);
onMounted(() => {
  window.addEventListener("scroll", onScroll, { passive: true });
  onScroll();
  restore().catch(() => {});
  navigation.load();
});
onBeforeUnmount(() => window.removeEventListener("scroll", onScroll));
</script>
<template>
  <div class="blog-app">
    <a class="skip-link" href="#main-content">跳至主要内容</a>
    <header class="blog-header" :class="{ 'is-solid': solid }">
      <LiquidGlass :radius="22" />
      <router-link class="blog-brand" to="/"
        ><BlogIcon name="feather" :size="25" /><span
          >云间小记</span
        ></router-link
      >
      <nav class="desktop-nav" aria-label="主要导航">
        <button
          v-if="navigation.error.value"
          class="navigation-retry"
          @click="navigation.load()"
        >
          导航暂不可用 · 重试
        </button>
        <router-link
          v-for="link in links"
          :key="link.path"
          :to="link.path"
          :class="{ 'is-current': route.path === link.path }"
          >{{ link.label }}</router-link
        >
      </nav>
      <div class="header-actions">
        <router-link to="/search" aria-label="搜索文章"
          ><BlogIcon name="search" :size="18" /></router-link
        ><router-link
          :to="member.user ? '/account' : '/login'"
          aria-label="会员中心"
          ><BlogIcon name="user" :size="19" /></router-link
        ><button
          class="menu-toggle"
          :aria-expanded="menu"
          aria-controls="mobile-navigation"
          aria-label="切换导航"
          @click="menu = !menu"
        >
          <BlogIcon :name="menu ? 'close' : 'menu'" />
        </button>
      </div>
      <nav
        v-if="menu"
        id="mobile-navigation"
        class="mobile-nav"
        aria-label="移动导航"
      >
        <button
          v-if="navigation.error.value"
          class="navigation-retry"
          @click="navigation.load()"
        >
          导航暂不可用 · 重试
        </button>
        <router-link v-for="link in links" :key="link.path" :to="link.path">{{
          link.label
        }}</router-link>
      </nav>
    </header>
    <main id="main-content"><slot /></main>
    <footer v-if="route.path !== '/messages'" class="blog-footer">
      <BlogIcon name="feather" :size="25" />
      <div>
        <router-link to="/about">关于</router-link><span>·</span
        ><router-link to="/friends">友链</router-link><span>·</span
        ><router-link to="/messages">留言</router-link>
      </div>
      <small
        >© {{ new Date().getFullYear() }} 云间小记 · Built with SiteHub</small
      >
    </footer>
    <button
      v-if="scrolled"
      class="back-top"
      aria-label="返回顶部"
      @click="scrollToTop"
    >
      ↑
    </button>
  </div>
</template>
<style lang="scss" scoped>
.navigation-retry {
  border: 0;
  background: none;
  color: inherit;
  font-size: 10px;
  padding: 8px 0;
  opacity: 0.8;
}
.blog-header {
  height: 60px;
  padding: 0 clamp(14px, 2.4vw, 34px);
  width: min(1120px, 92%);
  margin-inline: auto;
  isolation: isolate;
  display: flex;
  align-items: center;
  gap: 25px;
  position: absolute;
  inset: 18px 0 auto;
  z-index: 40;
  color: white;
  transition:
    background 0.3s,
    box-shadow 0.3s;
  &.is-solid {
    position: fixed;
    color: var(--ink);
    --glass-tint: rgb(255 255 255 / 92%);
  }
}
.blog-brand {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-right: auto;
  > span {
    font-size: 18px;
    letter-spacing: 3px;
    font-weight: 500;
  }
  small {
    display: block;
    font-size: 6px;
    letter-spacing: 2px;
    margin-top: 5px;
    opacity: 0.7;
  }
}
.desktop-nav {
  display: flex;
  gap: 27px;
  font-size: 14px;
  > a {
    padding: 9px 0;
    position: relative;
    &:after {
      content: "";
      position: absolute;
      bottom: 0;
      left: 50%;
      width: 0;
      height: 1px;
      background: currentColor;
      transition: 0.2s;
    }
    &:hover:after,
    &.is-current:after {
      left: 15%;
      width: 70%;
    }
  }
}
.header-actions {
  display: flex;
  align-items: center;
  gap: 23px;
  margin-left: 18px;
  > a {
    display: flex;
  }
  button {
    color: inherit;
    background: none;
    padding: 0;
    border: 0;
  }
}
.menu-toggle {
  display: none;
}
.mobile-nav {
  display: none;
}
.blog-footer {
  background: #f8f8fa;
  padding: 46px 20px 32px;
  text-align: center;
  color: #aaa3b4;
  > p {
    font-size: 11px;
    letter-spacing: 2px;
    margin: 16px 0 23px;
  }
  > div {
    margin-top: 16px;
    display: flex;
    justify-content: center;
    gap: 20px;
    font-size: 10px;
  }
  small {
    display: block;
    font-size: 9px;
    margin-top: 24px;
    letter-spacing: 0.5px;
  }
}
.back-top {
  position: fixed;
  bottom: 25px;
  right: 23px;
  z-index: 20;
  border: 1px solid var(--line);
  background: white;
  box-shadow: 0 4px 15px #3a335711;
  width: 35px;
  height: 35px;
  border-radius: 4px;
  color: var(--accent);
  cursor: pointer;
}
@media (max-width: 900px) {
  .desktop-nav {
    gap: 18px;
  }
  .blog-header {
    padding: 0 4%;
  }
  .header-actions {
    gap: 16px;
    margin-left: 0;
  }
}
@media (max-width: 730px) {
  .desktop-nav {
    display: none;
  }
  .menu-toggle {
    display: block;
  }
  .blog-header {
    height: 56px;
    inset-block-start: 12px;
    width: calc(100% - 24px);
  }
  .blog-brand > span {
    font-size: 17px;
  }
  .mobile-nav {
    display: grid;
    grid-template-columns: repeat(4, 1fr);
    position: absolute;
    top: 64px;
    border-radius: 18px;
    border: 1px solid var(--line);
    left: 0;
    right: 0;
    background: #fffffff5;
    color: var(--ink);
    padding: 15px 20px 22px;
    box-shadow: 0 12px 18px #3431480a;
    gap: 8px;
    a {
      padding: 13px;
      text-align: center;
      font-size: 12px;
    }
  }
}
</style>
