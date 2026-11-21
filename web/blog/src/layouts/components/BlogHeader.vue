<script setup>
import MenuBranch from "@shared/navigation/MenuBranch.vue";
import { computed, inject, onBeforeUnmount, onMounted, ref, watch } from "vue";
import { useRoute } from "vue-router";
import BlogIcon from "@/components/BlogIcon.vue";
import { useMember } from "@/composables/useMember";
const navigation = inject('navigation');
const reload = () => window.location.reload();
const route = useRoute(), menu = ref(false), scrolled = ref(false), hidden = ref(false);
const header = ref(null);
const { member, restore } = useMember();
const overContent = ref(false);
const solid = computed(() => ['article','login','register','account'].includes(route.meta.page) || overContent.value || menu.value);
let anchor = 0,
  frame = 0,
  compactLayout;
function updateHeader() {
  frame = 0;
  const position = Math.max(
    0,
    Math.min(
      window.scrollY,
      document.documentElement.scrollHeight - window.innerHeight,
    ),
  );
  scrolled.value = position > 55;
  const hero = document.querySelector("main .sky-hero, main .message-universe");
  overContent.value = hero
    ? hero.getBoundingClientRect().bottom <= 60
    : position > 55;
  if (position <= 64 || menu.value || header.value?.querySelector('[aria-expanded="true"]')) {
    hidden.value = false;
    anchor = position;
  } else if (hidden.value) {
    anchor = Math.max(anchor, position);
    if (anchor - position >= 12) {
      hidden.value = false;
      anchor = position;
    }
  } else {
    anchor = Math.min(anchor, position);
    if (position - anchor >= 48) {
      hidden.value = true;
      anchor = position;
    }
  }
}
const onScroll = () => {
  if (!frame) frame = requestAnimationFrame(updateHeader);
};
const scrollToTop = () =>
  window.scrollTo({
    top: 0,
    behavior: matchMedia("(prefers-reduced-motion: reduce)").matches
      ? "instant"
      : "smooth",
  });
function resetNavigation() {
  menu.value = false;
  hidden.value = false;
  anchor = window.scrollY;
}
watch(
  () => route.fullPath,
  () => {
    resetNavigation();
    onScroll();
  },
  { flush: "post" },
);
onMounted(() => {
  compactLayout = matchMedia("(max-width: 730px)");
  compactLayout.addEventListener("change", resetNavigation);
  window.addEventListener("scroll", onScroll, { passive: true });
  window.addEventListener("resize", onScroll);
  anchor = window.scrollY;
  updateHeader();
  restore().catch(() => {});
});
onBeforeUnmount(() => {
  compactLayout?.removeEventListener("change", resetNavigation);
  window.removeEventListener("scroll", onScroll);
  window.removeEventListener("resize", onScroll);
  cancelAnimationFrame(frame);
});
</script>
<template>
  <header
    ref="header" class="blog-header"
    :class="{ 'is-solid': solid, 'is-hidden': hidden && !menu }"
  >
    <router-link class="blog-brand" :to="{ name: 'home' }"
      ><BlogIcon name="feather" :size="25" /><span>云间小记</span></router-link
    >
    <nav class="desktop-nav" aria-label="主要导航">
      <button
        v-if="navigation.failed"
        class="navigation-retry"
        @click="reload()"
      >
        导航暂不可用 · 重试
      </button>
      <MenuBranch v-for="item in navigation.links" :key="item.id" :item="item" />
    </nav>
    <div class="header-actions">
      <router-link :to="{name:'search'}" aria-label="搜索文章"
        ><BlogIcon name="search" :size="18" /></router-link
      ><router-link
        :to="member.user ? {name:'account'} : {name:'login', query:['login','register'].includes(route.meta.page) ? route.query : {redirect:route.fullPath}}"
        :aria-label="member.user ? '个人中心' : '会员登录'"
        ><BlogIcon name="user" :size="18" /><span class="member-label">{{ member.user ? '个人中心' : '登录' }}</span></router-link
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
        v-if="navigation.failed"
        class="navigation-retry"
        @click="reload()"
      >
        导航暂不可用 · 重试
      </button>
      <MenuBranch v-for="item in navigation.links" :key="item.id" :item="item" mobile />
    </nav>
  </header>
  <button
    v-if="scrolled"
    class="back-top"
    aria-label="返回顶部"
    @click="scrollToTop"
  >
    ↑
  </button>
</template>
<style lang="scss" scoped src="./header.scss"></style>
