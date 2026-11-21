import { createRouter, createWebHistory } from "vue-router";

export function createBlogRouter(routes) {
  const router = createRouter({
    history: createWebHistory(import.meta.env.BASE_URL),
    routes,
    scrollBehavior(to, from, saved) {
      if (saved) return saved;
      if (to.hash)
        return {
          el: to.hash,
          top: 90,
          behavior: matchMedia("(prefers-reduced-motion: reduce)").matches
            ? "instant"
            : "smooth",
        };
      if (to.path === from.path) return false;
      return { top: 0 };
    },
  });
  router.afterEach(() => { document.title = "云间小记 · SiteHub"; });
  return router;
}
