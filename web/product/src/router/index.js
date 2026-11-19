import { createRouter, createWebHistory } from 'vue-router';
import ProductLayout from '@/layouts/ProductLayout.vue';
import { createPageRoutes } from './pages.js';
import { beginNavigation, finishNavigation, failNavigation } from '@/composables/useNavigationState';

const pages = import.meta.glob('../views/**/index.vue');

const options = import.meta.glob('../views/**/route.js', { eager: true, import: 'default' });

export function createProductRouter(menuItems) {
  const router = createRouter({
    history: createWebHistory(import.meta.env.BASE_URL),
    routes: createPageRoutes(pages, ProductLayout, menuItems, options),
    scrollBehavior(to, from, savedPosition) {
      if (savedPosition) return savedPosition;
      if (to.path === from.path && to.hash === from.hash) return false;
      if (to.hash) return { el: to.hash, behavior: matchMedia('(prefers-reduced-motion: reduce)').matches ? 'instant' : 'smooth' };
      return { top: 0 };
    },
  });
  router.beforeEach(to => beginNavigation(to.fullPath));
  router.afterEach((to, _from, failure) => {
    finishNavigation(to.fullPath);
    if (!failure) document.title = to.meta.notFound ? '页面不存在｜Deploy' : 'Deploy｜产品、下载与文档';
  });
  router.onError((_error, to) => failNavigation(to.fullPath));
  return router;
}
