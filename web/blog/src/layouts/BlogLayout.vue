<script setup>
import { computed, onBeforeUnmount, onMounted, ref, watch } from "vue";
import { useRoute, useRouter } from "vue-router";
import BlogIcon from "@/components/BlogIcon.vue";
import { useMember } from "@/composables/useMember";
import { useResource } from "@/composables/useResource";
import { publicMenus } from "@/api";
import { navigationLinks } from "@/utils/navigation.mjs";
const route = useRoute(),
  menu = ref(false),
  scrolled = ref(false),
  hidden = ref(false),
  contentOpen = ref(false),
  contentGroup = ref(null),
  contentTrigger = ref(null);
const { member, restore } = useMember();
const navigation = useResource(publicMenus);
const router = useRouter();
const links = computed(() =>
  navigationLinks(navigation.data.value?.items, router.getRoutes()),
);
const contentPages = new Map([
  ["/categories", { icon: "folder", hint: "按主题浏览文章" }],
  ["/tags", { icon: "tag", hint: "按标签发现文章" }],
  ["/archives", { icon: "clock", hint: "按时间回顾文章" }],
]);
const icons = { "/": "home", "/friends": "link", "/about": "info", "/messages": "message" };
const desktopLinks = computed(() => {
  const items = [];
  const group = { label: "内容", children: [] };
  for (const link of links.value) {
    if (contentPages.has(link.path)) {
      if (!group.children.length) items.push(group);
      group.children.push({ ...link, ...contentPages.get(link.path) });
    } else items.push(link);
  }
  return items;
});
const overContent = ref(false);
const solid = computed(() => overContent.value || menu.value);
let anchor = 0, frame = 0, compactLayout, closeTimer;
function updateHeader() {
  frame = 0;
  const position = Math.max(0, Math.min(window.scrollY, document.documentElement.scrollHeight - window.innerHeight));
  scrolled.value = position > 55;
  const hero = document.querySelector("main .sky-hero, main .message-universe");
  overContent.value = hero ? hero.getBoundingClientRect().bottom <= 60 : position > 55;
  if (position <= 64 || menu.value || contentOpen.value) {
    hidden.value = false;
    anchor = position;
  } else if (hidden.value) {
    anchor = Math.max(anchor, position);
    if (anchor - position >= 12) { hidden.value = false; anchor = position; }
  } else {
    anchor = Math.min(anchor, position);
    if (position - anchor >= 48) { hidden.value = true; anchor = position; }
  }
}
const onScroll = () => { if (!frame) frame = requestAnimationFrame(updateHeader); };
const scrollToTop = () => window.scrollTo({ top: 0, behavior: matchMedia("(prefers-reduced-motion: reduce)").matches ? "instant" : "smooth" });
const closeOutside = (event) => {
  if (!contentGroup.value?.contains(event.target)) contentOpen.value = false;
};
function enterContent(event) {
  if (event.pointerType !== "mouse") return;
  clearTimeout(closeTimer);
  contentOpen.value = true;
}
function leaveContent(event) {
  if (event.pointerType !== "mouse") return;
  clearTimeout(closeTimer);
  closeTimer = setTimeout(() => {
    if (!contentGroup.value?.querySelector(":focus-visible")) contentOpen.value = false;
  }, 160);
}
function activateContent(event) {
  clearTimeout(closeTimer);
  contentOpen.value = event.detail && event.pointerType === "mouse" ? true : !contentOpen.value;
}
function resetNavigation() {
  clearTimeout(closeTimer);
  menu.value = false;
  contentOpen.value = false;
  hidden.value = false;
  anchor = window.scrollY;
}
function closeContent() {
  contentOpen.value = false;
  contentTrigger.value?.focus();
}
watch(() => route.fullPath, () => { resetNavigation(); onScroll(); }, { flush: 'post' });
onMounted(() => {
  compactLayout = matchMedia("(max-width: 730px)");
  compactLayout.addEventListener("change", resetNavigation);
  window.addEventListener("scroll", onScroll, { passive: true });
  window.addEventListener("resize", onScroll);
  document.addEventListener("pointerdown", closeOutside);
  anchor = window.scrollY;
  updateHeader();
  restore().catch(() => {});
  navigation.load();
});
onBeforeUnmount(() => {
  clearTimeout(closeTimer);
  compactLayout?.removeEventListener("change", resetNavigation);
  window.removeEventListener("scroll", onScroll);
  window.removeEventListener("resize", onScroll);
  document.removeEventListener("pointerdown", closeOutside);
  cancelAnimationFrame(frame);
});
</script>
<template>
  <div class="blog-app">
    <a class="skip-link" href="#main-content">跳至主要内容</a>
    <header class="blog-header" :class="{ 'is-solid': solid, 'is-hidden': hidden && !menu && !contentOpen }">
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
        <template v-for="link in desktopLinks" :key="link.path || 'content'">
          <div v-if="link.children" :ref="element => contentGroup = element" class="nav-group"
            @pointerenter="enterContent" @pointerleave="leaveContent"
            @focusout="!$event.currentTarget.contains($event.relatedTarget) && (contentOpen = false)"
            @keydown.esc.stop="closeContent">
            <button :ref="element => contentTrigger = element" type="button" class="nav-item" aria-controls="content-links"
              :class="{ 'is-current': contentPages.has(route.path) }" :aria-expanded="contentOpen" @click="activateContent">
              <BlogIcon name="folder" :size="18" />{{ link.label }}<BlogIcon class="nav-caret" name="chevron" :size="12" />
            </button>
            <Transition name="nav-drop">
              <div v-show="contentOpen" id="content-links" class="content-links" :inert="!contentOpen">
                <div class="content-sheet">
                  <router-link v-for="child in link.children" :key="child.path" :to="child.path">
                    <BlogIcon class="content-icon" :name="child.icon" :size="19" />
                    <span>{{ child.label }}<small>{{ child.hint }}</small></span>
                    <BlogIcon class="row-arrow" name="chevron" :size="14" />
                  </router-link>
                </div>
              </div>
            </Transition>
          </div>
          <a v-else-if="link.external" class="nav-item" :href="link.path" target="_blank" rel="noopener noreferrer">{{ link.label }} ↗</a>
          <router-link v-else class="nav-item" :to="link.path" :class="{ 'is-current': route.path === link.path }">
            <BlogIcon v-if="icons[link.path]" :name="icons[link.path]" :size="18" />{{ link.label }}
          </router-link>
        </template>
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
        <template v-for="link in links" :key="link.path">
          <a v-if="link.external" :href="link.path" target="_blank" rel="noopener noreferrer" @click="menu = false">{{ link.label }} ↗</a>
          <router-link v-else :to="link.path">{{ link.label }}</router-link>
        </template>
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
  --menu-surface: linear-gradient(145deg, #252f58f0, #373957ee);
  --menu-ink: #fffbf5;
  --menu-muted: #d0d5e7;
  --menu-edge: #eef0ff36;
  --menu-row: #bfbdec29;
  --menu-icon: #e6ddff;
  height: 60px;
  padding: 0 clamp(14px, 2.4vw, 34px);
  display: flex;
  align-items: center;
  gap: 25px;
  position: fixed;
  inset: 0 0 auto;
  z-index: 40;
  color: white;
  background: rgb(27 32 65 / 40%);
  backdrop-filter: blur(10px);
  transition:
    transform 340ms cubic-bezier(.22,.75,.2,1),
    color .25s, background-color .3s,
    box-shadow 0.3s;
  &.is-hidden:not(:has(:focus-visible)) {
    transform: translateY(calc(-100% - 8px));
    pointer-events: none;
  }
  &.is-solid {
    --menu-surface: linear-gradient(145deg, #fffdf9fa, #f0ecf5fa);
    --menu-ink: #423d52;
    --menu-muted: #696378;
    --menu-edge: #807c982e;
    --menu-row: #9790b624;
    --menu-icon: #6c6486;
    color: var(--ink);
    background: rgb(252 250 248 / 96%);
    box-shadow: 0 2px 12px #362d4512;
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
  align-items: center;
  gap: clamp(8px, 1.8vw, 28px);
  font-size: 14px;
}
.nav-item {
  display: flex;
  align-items: center;
  gap: 7px;
  padding: 9px 2px;
  color: inherit;
  background: transparent;
  border: 0;
  border-bottom: 1px solid transparent;
  &:hover, &.is-current { border-bottom-color: currentColor; }
}
.nav-caret { transform: rotate(90deg); transition: transform .2s; }
[aria-expanded="true"] .nav-caret { transform: rotate(-90deg); }
.nav-group { position: relative; }
.nav-group > button[aria-expanded="true"] { border-bottom-color: currentColor; }
.content-links {
  position: absolute;
  top: 100%;
  left: -15px;
  width: 224px;
  padding-top: 9px;
}
.nav-drop-enter-active, .nav-drop-leave-active { transition: opacity .15s ease, transform .18s ease-out; }
.nav-drop-enter-from, .nav-drop-leave-to { opacity: 0; transform: translateY(-4px); }
.content-sheet {
  padding: 6px;
  color: var(--menu-ink);
  background: var(--menu-surface);
  backdrop-filter: blur(14px);
  border: 1px solid var(--menu-edge);
  border-radius: 15px;
  box-shadow: 0 12px 30px #19213933;
  a {
    display: grid;
    grid-template-columns: 30px 1fr 14px;
    align-items: center;
    gap: 9px;
    padding: 10px 11px;
    border-radius: 9px;
    line-height: 1.4;
    font-weight: 550;
    transition: background-color .15s;
    &:hover, &:focus-visible, &.router-link-active { background: var(--menu-row); }
    &:hover .row-arrow, &:focus-visible .row-arrow { opacity: 1; transform: translateX(2px); }
  }
  small { display: block; margin-top: 3px; color: var(--menu-muted); font-size: 11px; font-weight: 400; }
}
.content-icon { justify-self: center; color: var(--menu-icon); }
.row-arrow { opacity: .55; transition: opacity .15s, transform .15s; }
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
  }
  .blog-brand > span {
    font-size: 17px;
  }
  .mobile-nav {
    display: grid;
    grid-template-columns: repeat(4, 1fr);
    position: absolute;
    top: calc(100% + 6px);
    border-radius: 18px;
    border: 1px solid var(--line);
    left: 12px;
    right: 12px;
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
