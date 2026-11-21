import { createRouter } from "vue-router";
import { isSuperAdmin, menuRoutes, navigationMenus } from "./pages.js";

// Exercise the real navigation lifecycle with either web or memory history.
export function createAdminRouter({
  history,
  modules,
  settings = {},
  layout,
  store,
  storage = globalThis.sessionStorage,
  document = globalThis.document,
}) {
  const standalone = [], children = [];
  for (const [file, options] of Object.entries(settings)) {
    const component = modules[file.replace(/menu\.json$/, 'index.vue')];
    for (const record of options.routes || []) {
      (options.layout ? children : standalone).push({ ...record, component });
    }
  }
  const notFound = children.find(record => record.name === 'NotFound')?.component;
  const router = createRouter({
    history,
    routes: [
      ...standalone,
      { path: "/", redirect: "/dashboard" },
      { path: "/dashboard", name: "dashboard", component: layout, children },
    ],
  });
  const reservedPaths = router.getRoutes().map(record => record.path);
  router.navigation = (menus, user) => navigationMenus(menus, isSuperAdmin(user), settings);
  let removers = [];
  let signature = null;
  let initialized = false;
  router.resetMenus = () => {
    for (const remove of removers) remove();
    removers = [];
    signature = null;
    initialized = false;
  };
  router.installMenus = (menus) => {
    router.resetMenus();
    for (const route of menuRoutes(menus, modules, notFound, settings, reservedPaths))
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
    const session = storage?.getItem("token");
    if (!session) {
      router.resetMenus();
      return { path: "/login", query: { redirect: to.fullPath } };
    }
    if (to.name === 'AccessUnavailable') return true;
    const rbac = store();
    try {
      if (!rbac.menusLoaded) await rbac.loadAllDataFromBackend();
      if (!rbac.menusLoaded) throw new Error("菜单加载失败");
      const menus = rbac.currentMenus || [];
      if (!initialized || signature !== JSON.stringify(menus)) {
        router.installMenus(menus);
        return { path: to.path, query: to.query, hash: to.hash, replace: true };
      }
    } catch (error) {
      if (session !== storage?.getItem('token') && storage?.getItem('token')) return false;
      router.resetMenus();
      rbac.resetAllData();
      if (storage?.getItem('token') && error.response?.status !== 401) {
        return { name: 'AccessUnavailable', query: { redirect: to.fullPath }, replace: true };
      }
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
