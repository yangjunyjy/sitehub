<script setup>
import MenuBranch from '@shared/navigation/MenuBranch.vue'
import ThemeToggle from "@/components/product/ThemeToggle.vue";
import BrandMark from "@/components/product/BrandMark.vue";
import SiteSearch from "@/components/product/SiteSearch.vue";
import { inject, onBeforeUnmount, onMounted, ref, watch } from "vue";
import { useRoute } from "vue-router";
import { useWebsiteContent } from "@/composables/useWebsiteContent";
import { useMember } from "@/composables/useMember";

const { member, restoreMember } = useMember();

const route = useRoute();
const navigation = inject('navigation');
const reload = () => window.location.reload();
const menuOpen = ref(false);
const menuButton = ref(null);
const header = ref(null);
const headerHidden = ref(false);
const searchOpen = ref(false);
let lastScroll = 0,
  scrollDirection = 0,
  scrollDistance = 0,
  scrollFrame = 0;

// Clamp elastic overscroll so bouncing at either end does not reverse the header.
const scrollPosition = () =>
  Math.max(
    0,
    Math.min(
      window.scrollY,
      document.documentElement.scrollHeight - window.innerHeight,
    ),
  );
const resetScroll = () => {
  lastScroll = scrollPosition();
  scrollDirection = 0;
  scrollDistance = 0;
};
const revealHeader = () => {
  headerHidden.value = false;
  cancelAnimationFrame(scrollFrame);
  scrollFrame = 0;
  resetScroll();
};
const updateHeader = () => {
  scrollFrame = 0;
  const position = scrollPosition();
  const delta = position - lastScroll;
  lastScroll = position;
  const keyboardFocus =
    header.value?.contains(document.activeElement) &&
    document.activeElement?.matches(":focus-visible");
  if (position <= 120 || menuOpen.value || searchOpen.value || keyboardFocus) {
    revealHeader();
    return;
  }
  if (!delta) return;
  const direction = Math.sign(delta);
  if (direction !== scrollDirection) {
    scrollDirection = direction;
    scrollDistance = 0;
  }
  scrollDistance += Math.abs(delta);
  // Different thresholds make upward intent responsive without reacting to small jitter.
  if (direction > 0 && scrollDistance >= 36) headerHidden.value = true;
  if (direction < 0 && scrollDistance >= 12) headerHidden.value = false;
};
const onScroll = () => {
  if (!scrollFrame) scrollFrame = requestAnimationFrame(updateHeader);
};
const openSearch = () => {
  menuOpen.value = false;
  searchOpen.value = true;
};
const closeMenu = (event) => {
  if (event.key === "Escape" && menuOpen.value) {
    menuOpen.value = false;
    menuButton.value?.focus();
  }
};
const desktop = window.matchMedia("(min-width:1081px)");
const resetDesktop = () => {
  if (desktop.matches) menuOpen.value = false;
};
onMounted(() => {
  restoreMember().catch(() => {});
  resetScroll();
  window.addEventListener("scroll", onScroll, { passive: true });
  window.addEventListener("resize", revealHeader);
  window.addEventListener("keydown", closeMenu);
  desktop.addEventListener("change", resetDesktop);
});
onBeforeUnmount(() => {
  cancelAnimationFrame(scrollFrame);
  window.removeEventListener("scroll", onScroll);
  window.removeEventListener("resize", revealHeader);
  document.documentElement.classList.remove("site-header-hidden");
  window.removeEventListener("keydown", closeMenu);
  desktop.removeEventListener("change", resetDesktop);
});
const { website } = useWebsiteContent();
watch(
  () => route.fullPath,
  () => {
    menuOpen.value = false;
    revealHeader();
  },
);
watch(menuOpen, (open) =>
  document.documentElement.classList.toggle("menu-open", open),
);
watch([menuOpen, searchOpen], revealHeader);
watch(headerHidden, (hidden) =>
  document.documentElement.classList.toggle("site-header-hidden", hidden),
);
onBeforeUnmount(() => document.documentElement.classList.remove("menu-open"));
</script>

<template>
  <header
    ref="header"
    class="product-header"
    :class="{
      'is-hidden': headerHidden,
      'is-docs': route.meta.page === 'docs',
    }"
    @focusin="revealHeader"
  >
    <div class="site-shell product-header__inner">
      <router-link
        class="brand"
        :to="{ name: 'home' }"
        :aria-label="`${website.content.brand.name} 首页`"
      >
        <BrandMark />
        <strong>{{ website.content.brand.name }}</strong
        >
      </router-link>
      <div class="header-search">
        <SiteSearch @open="openSearch" @close="searchOpen = false" />
      </div>
      <nav class="desktop-nav" aria-label="主导航">
        <button v-if="navigation.failed" type="button" @click="reload">导航暂不可用 · 重试</button>
        <MenuBranch v-for="item in navigation.links" :key="item.id" :item="item" />
      </nav>
      <div class="header-actions">
        <ThemeToggle />
        <router-link
          class="header-contact"
          :to="{ name: member.user ? 'account' : 'login' }"
          ><svg viewBox="0 0 24 24" aria-hidden="true">
            <circle cx="12" cy="8" r="3.5" />
            <path d="M5 21v-3a7 7 0 0 1 14 0v3" /></svg
          ><span>{{ member.user ? "我的账号" : "登录" }}</span></router-link
        >
      </div>
      <button
        ref="menuButton"
        class="menu-button"
        type="button"
        :aria-expanded="menuOpen"
        :aria-label="menuOpen ? '关闭导航' : '打开导航'"
        aria-controls="mobile-navigation"
        @click="menuOpen = !menuOpen"
      >
        <span></span><span></span>
      </button>
    </div>
    <div
      id="mobile-navigation"
      class="mobile-navigation"
      :class="{ open: menuOpen }"
      :inert="!menuOpen"
    >
      <nav class="site-shell">
        <button v-if="navigation.failed" type="button" @click="reload">导航暂不可用 · 重试</button>
        <MenuBranch v-for="item in navigation.links" :key="item.id" :item="item" mobile /><a
          v-if="website.content.brand.contactEmail"
          :href="`mailto:${website.content.brand.contactEmail}?subject=Deploy 企业版咨询`"
          >企业咨询 </a
        ><router-link :to="{ name: member.user ? 'account' : 'login' }"
          >{{ member.user ? "我的账号" : "登录" }}
        </router-link>
        <router-link
          class="site-button site-button--primary"
          :to="{ name: member.user ? 'download' : 'register' }"
          >{{ member.user ? "立即下载" : "注册账号" }}</router-link
        >
      </nav>
    </div>
  </header>
</template>

<style lang="scss" scoped>
.product-header {
  position: sticky;
  top: 0;
  z-index: 100;
  height: 64px;
  background: var(--header-bg);
  backdrop-filter: blur(12px);
  border-bottom: 1px solid var(--deploy-line);
  transition:
    transform 0.25s,
    opacity 0.2s;
}
.product-header.is-hidden {
  transform: translateY(-100%);
  opacity: 0;
  pointer-events: none;
}
.product-header__inner {
  width: 100%;
  max-width: 1600px;
  padding: 0 32px;
  display: flex;
  align-items: center;
  gap: 28px;
  height: 100%;
}
.brand {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-shrink: 0;
}
.brand :deep(svg) {
  width: 26px;
  height: 26px;
}
.brand strong {
  font-size: 16px;
  font-weight: 650;
}
.header-search {
  margin-right: auto;
  min-width: 0;
}
.desktop-nav {
  display: flex;
  align-items: center;
  gap: 25px;
}
.desktop-nav a {
  padding: 20px 0;
  font-size: 14px;
  color: var(--deploy-ink);
  white-space: nowrap;
}
.desktop-nav a:hover,
.desktop-nav a.active {
  color: var(--deploy-primary);
}
.header-actions {
  display: flex;
  align-items: center;
  gap: 20px;
  font-size: 13px;
}
.header-contact {
  display: flex;
  gap: 6px;
  align-items: center;
  border-left: 1px solid var(--deploy-line);
  padding-left: 20px;
  white-space: nowrap;
  color: var(--deploy-muted);
}
.header-contact:hover {
  color: var(--deploy-primary);
}
.header-contact svg {
  display: none;
}
.menu-button,
.mobile-navigation {
  display: none;
}
@media (min-width: 1081px) {
  .product-header.is-docs .brand {
    width: 212px;
  }
  .header-search {
    max-width: 200px;
  }
}
@media (max-width: 1080px) {
  .product-header__inner {
    gap: 18px;
    padding: 0 24px;
  }
  .desktop-nav,
  .header-contact {
    display: none;
  }
  .header-search {
    margin-left: auto;
    margin-right: 0;
  }
  .header-actions {
    gap: 0;
  }
  .menu-button {
    display: grid;
    place-content: center;
    gap: 6px;
    width: 30px;
    height: 34px;
    border: 0;
    background: transparent;
  }
  .menu-button span {
    width: 18px;
    height: 1.5px;
    background: var(--deploy-muted);
  }
  .mobile-navigation {
    position: fixed;
    inset: 64px 0 auto;
    height: calc(100dvh - 64px);
    overflow: auto;
    display: block;
    visibility: hidden;
    background: var(--deploy-white);
    opacity: 0;
    transition: opacity 0.2s;
  }
  .mobile-navigation.open {
    visibility: visible;
    opacity: 1;
  }
  .mobile-navigation nav {
    display: grid;
    padding: 24px 16px 40px;
    max-width: 620px;
    gap: 0;
  }
  .mobile-navigation nav > a:not(.site-button) {
    padding: 15px 0;
    border-bottom: 1px solid var(--deploy-line);
    font-size: 14px;
  }
  .mobile-navigation .router-link-exact-active {
    color: var(--deploy-primary);
  }
  .mobile-navigation .site-button {
    margin-top: 24px;
  }
}
@media (max-width: 540px) {
  .product-header__inner {
    padding: 0 20px;
    gap: 14px;
  }
  .brand {
    gap: 6px;
  }
  .header-search {
    max-width: 40px;
  }
  .header-actions {
    margin-left: 0;
  }
}
</style>
