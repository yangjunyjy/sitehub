<script setup>
import { ref, onMounted, onBeforeUnmount } from "vue";
import {
  Connection,
  Fold,
  Expand,
  SwitchButton,
  User,
} from "@element-plus/icons-vue";
import { pages, currentPage } from "./pages";
defineProps({ actor: String, role: String, busy: Boolean });
const roleNames = { viewer: "只读", publisher: "发布员", admin: "管理员" };
defineEmits(["logout"]);
const page = ref(currentPage()),
  collapsed = ref(false),
  mobile = ref(false);
function navigate() {
  page.value = currentPage();
  mobile.value = false;
}
const media = window.matchMedia("(max-width: 760px)");
function toggle() {
  if (media.matches) mobile.value = !mobile.value;
  else collapsed.value = !collapsed.value;
}
function escape(e) {
  if (e.key === "Escape") mobile.value = false;
}
function resize() {
  mobile.value = false;
}
onMounted(() => {
  window.addEventListener("hashchange", navigate);
  window.addEventListener("keydown", escape);
  media.addEventListener("change", resize);
});
onBeforeUnmount(() => {
  window.removeEventListener("hashchange", navigate);
  window.removeEventListener("keydown", escape);
  media.removeEventListener("change", resize);
});
</script>
<template>
  <div class="shell" :class="{ collapsed, 'mobile-open': mobile }">
    <button
      v-if="mobile"
      class="scrim"
      aria-label="关闭菜单"
      @click="mobile = false"
    ></button>
    <aside class="sidebar">
      <a class="brand" href="#routes" aria-label="网关控制台首页"
        ><span class="brand-mark"><Connection /></span
        ><strong>网关控制台</strong></a
      >
      <nav aria-label="主导航">
        <div v-for="group in ['流量管理', '变更管理']" :key="group">
          <small>{{ group }}</small>
          <a
            v-for="item in pages.filter((p) => p.group === group)"
            :key="item.id"
            :href="'#' + item.id"
            :title="item.name"
            :aria-label="item.name"
            :aria-current="page.id === item.id ? 'page' : undefined"
            @click="mobile = false"
            ><component :is="item.icon" /><span>{{ item.name }}</span></a
          >
        </div>
      </nav>
      <button
        class="sidebar-toggle"
        @click="toggle"
        :aria-label="collapsed ? '展开菜单' : '收起菜单'"
      >
        <component :is="collapsed ? Expand : Fold" /><span>收起菜单</span>
      </button>
    </aside>
    <section class="main-column">
      <header class="topbar">
        <button class="icon" aria-label="切换导航" @click="toggle">
          <component :is="collapsed ? Expand : Fold" /></button
        ><span class="crumb"
          >{{ page.group }} <b>/</b> <strong>{{ page.name }}</strong></span
        >
        <div class="account">
          <span class="avatar"><User /></span>{{ actor
          }}<small>{{ roleNames[role] }}</small>
        </div>
        <button class="exit" :disabled="busy" @click="$emit('logout')">
          <SwitchButton />退出登录
        </button>
      </header>
      <main class="work-area"><slot :section="page.id" /></main>
    </section>
  </div>
</template>
