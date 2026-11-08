import { createRouter, createWebHistory } from "vue-router";
import {
  beginNavigation,
  finishNavigation,
  failNavigation,
} from "@/composables/useNavigationState";

const routes = [
  {
    path: "/login",
    name: "login",
    component: () => import("@/views/auth/index.vue"),
    meta: { title: "登录", hideLayout: true },
  },
  {
    path: "/register",
    name: "register",
    component: () => import("@/views/auth/index.vue"),
    meta: { title: "邮箱注册", hideLayout: true },
  },
  {
    path: "/account",
    name: "account",
    component: () => import("@/views/auth/Account.vue"),
    meta: { title: "我的账号" },
  },
  {
    path: "/",
    name: "home",
    component: () => import("@/views/Home.vue"),
    meta: { title: "企业一体化 DevOps 运维平台" },
  },
  {
    path: "/product",
    redirect: { path: "/", hash: "#capabilities" },
  },
  {
    path: "/download",
    name: "download",
    component: () => import("@/views/download/index.vue"),
    meta: { title: "下载" },
  },
  {
    path: "/docs/:id?",
    name: "docs",
    component: () => import("@/views/docs/index.vue"),
    meta: { title: "产品文档" },
  },
  {
    path: "/blog",
    name: "blog",
    component: () => import("@/views/blog/index.vue"),
    meta: { title: "技术博客", layout: "blog" },
  },
  {
    path: "/blog/archives",
    component: () => import("@/views/blog/Archives.vue"),
    meta: { title: "文章归档", layout: "blog" },
  },
  {
    path: "/blog/categories",
    component: () => import("@/views/blog/Topics.vue"),
    props: { kind: "categories" },
    meta: { title: "文章分类", layout: "blog" },
  },
  {
    path: "/blog/tags",
    component: () => import("@/views/blog/Topics.vue"),
    props: { kind: "tags" },
    meta: { title: "文章标签", layout: "blog" },
  },
  {
    path: "/blog/reading",
    component: () => import("@/views/blog/Reading.vue"),
    meta: { title: "阅读空间", layout: "blog" },
  },
  {
    path: "/articles",
    redirect: "/blog",
  },
  {
    path: "/article/:id",
    name: "article",
    component: () => import("@/views/ArticleDetail.vue"),
    meta: { title: "文章详情", layout: "blog" },
  },
  {
    path: "/about",
    redirect: "/",
  },
  {
    path: "/:pathMatch(.*)*",
    name: "not-found",
    component: () => import("@/views/NotFound.vue"),
    meta: { title: "页面不存在" },
  },
];

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes,
  scrollBehavior(to, from, savedPosition) {
    if (savedPosition) return savedPosition;
    if (to.path === from.path && to.hash === from.hash) return false;
    if (to.hash)
      return {
        el: to.hash,
        behavior: matchMedia("(prefers-reduced-motion: reduce)").matches
          ? "instant"
          : "smooth",
      };
    return { top: 0 };
  },
});

router.beforeEach((to) => beginNavigation(to.fullPath));
router.afterEach((to, _from, failure) => {
  finishNavigation(to.fullPath);
  if (failure) return;
  document.title = `${to.meta.title || "企业一体化 DevOps 运维平台"}｜Deploy`;
});
router.onError((_error, to) => failNavigation(to.fullPath));

export default router;
