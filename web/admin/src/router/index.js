import { createWebHistory } from "vue-router";
import { useRbacStore } from "../store/rbac";
import { createAdminRouter } from "./runtime.js";

const modules = import.meta.glob("../views/**/index.vue");
const router = createAdminRouter({
  history: createWebHistory(),
  modules,
  layout: () => import("../layout/Layout.vue"),
  store: useRbacStore,
});

export const addRoutes = (menus) => router.installMenus(menus);
export const resetRoutes = () => router.resetMenus();
export default router;
