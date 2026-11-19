import { createRouter } from "vue-router";
import { isSuperAdmin, menuRoutes } from "./pages.js";

// Exercise the real navigation lifecycle with either web or memory history.
export function createAdminRouter({
  history,
  modules,
  layout,
  store,
  initializeRequests = () => Promise.resolve(),
  storage = globalThis.sessionStorage,
  document = globalThis.document,
}) {
  const notFound = modules["../views/not-found/index.vue"];
  const router = createRouter({
    history,
    routes: [
      {
        path: "/login",
        name: "Login",
        component: modules["../views/login/index.vue"],
        meta: { title: "登录" },
      },
      { path: "/", redirect: "/dashboard" },
      {
        path: "/dashboard",
        name: "dashboard",
        component: layout,
        children: [
          {
            path: "/403",
            name: "Forbidden",
            component: notFound,
            props: { status: 403 },
            meta: { title: "无权访问", hidden: true },
          },
          {
            path: "/:pathMatch(.*)*",
            name: "NotFound",
            component: notFound,
            meta: { title: "页面不存在", hidden: true },
          },
        ],
      },
    ],
  });
  let removers = [];
  let signature = null;
  let initialized = false;
  let initializingRequests = false;
  router.resetMenus = () => {
    for (const remove of removers) remove();
    removers = [];
    signature = null;
    initialized = false;
  };
  router.installMenus = (menus) => {
    router.resetMenus();
    for (const route of menuRoutes(menus, modules, notFound))
      removers.push(router.addRoute("dashboard", route));
    signature = JSON.stringify(menus || []);
    initialized = true;
  };
  // Resolve links through the records granted by RBAC, including hidden editor pages.
  router.page = (component, location = {}) => {
    const record = router.getRoutes().find(route => route.meta.componentPath === component);
    return record ? { ...location, name: record.name } : { name: "Forbidden" };
  };
  router.beforeEach(async (to) => {
    if (document) document.title = `${to.meta.title || "管理后台"} · SiteHub`;
    if (to.path === "/login") {
      if (!storage?.getItem("token")) router.resetMenus();
      return true;
    }
    if (!storage?.getItem("token")) {
      router.resetMenus();
      return { path: "/login", query: { redirect: to.fullPath } };
    }
    if (!initializingRequests) {
      initializingRequests = true;
      Promise.resolve()
        .then(initializeRequests)
        .catch(() => {});
    }
    const rbac = store();
    try {
      if (!rbac.menusLoaded) await rbac.loadAllDataFromBackend();
      if (!rbac.menusLoaded) throw new Error("菜单加载失败");
      const menus = rbac.currentMenus || [];
      if (!initialized || signature !== JSON.stringify(menus)) {
        router.installMenus(menus);
        return { path: to.path, query: to.query, hash: to.hash, replace: true };
      }
    } catch {
      router.resetMenus();
      rbac.resetAllData();
      storage?.removeItem("token");
      storage?.removeItem("userInfo");
      return { path: "/login", query: { redirect: to.fullPath } };
    }
    if (to.name === "dashboard")
      return router.page("dashboard/index", { query: to.query, hash: to.hash });
    if (
      (to.meta.superAdmin && !isSuperAdmin(rbac.currentUser)) ||
      (to.meta.permission && !rbac.hasPermission(to.meta.permission))
    )
      return "/403";
    return true;
  });
  return router;
}
