// Menu components are exact paths relative to views, without extensions.
import { externalAddress } from '../utils/external-address.mjs';
export function componentKey(component) {
  if (
    typeof component !== "string" ||
    !/^(?:[a-zA-Z0-9_-]+|\[[a-zA-Z0-9_]+\])(?:\/(?:[a-zA-Z0-9_-]+|\[[a-zA-Z0-9_]+\]))*\/index$/.test(
      component,
    )
  )
    return "";
  return `../views/${component}.vue`;
}

export const isSuperAdmin = (user) =>
  (user?.roleIds || []).some((id) => Number(id) === 1) ||
  (user?.roles || []).some((role) => Number(role.id) === 1);

export function filterMenuTree(menus = [], visibleOnly = false) {
  return menus.flatMap((menu) => {
    if (menu.status !== undefined && Number(menu.status) !== 1) return [];
    if (visibleOnly && !menu.isVisible) return [];
    if (menu.type === "button") return visibleOnly ? [] : [{ ...menu }];
    const children = filterMenuTree(menu.children || [], visibleOnly);
    if (menu.type === "directory")
      return children.length ? [{ ...menu, children }] : [];
    if (menu.type !== "menu") return [];
    return [{ ...menu, children }];
  });
}

function menuPath(value, parentPath) {
  if (typeof value !== "string" || !value || value.startsWith("//")) return "";
  return value.startsWith("/")
    ? value
    : `${parentPath || ""}/${value}`.replace(/\/+/g, "/");
}

export function navigationMenus(
  menus = [],
  superAdmin = false,
  parentPath = "",
) {
  return filterMenuTree(menus, true)
    .flatMap((menu) => {
      if (!superAdmin && menu.component === "admin/settings/email/index")
        return [];
      if (menu.pageType === 'external') {
        const address = externalAddress(menu.externalUrl);
        return address ? [{ ...menu, externalUrl: address }] : [];
      }
      const path = menuPath(menu.path, parentPath);
      const children = navigationMenus(menu.children || [], superAdmin, path);
      if ((menu.type === "directory" && !children.length) || !path) return [];
      return [{ ...menu, path, children }];
    })
    .sort((a, b) => (Number(a.sort) || 0) - (Number(b.sort) || 0));
}

export function menuRoutes(menus, modules, notFound) {
  const routes = [];
  const usedPaths = new Set([
    "/login",
    "/access-unavailable",
    "/dashboard",
    "/",
    "/403",
  ]);
  const visit = (items, parentPath = "") => {
    for (const menu of filterMenuTree(items)) {
      if (menu.type === "button" || menu.pageType === 'external') continue;
      const path = menuPath(menu.path, parentPath);
      if (menu.type === "menu" && path && !usedPaths.has(path)) {
        const component = modules[componentKey(menu.component)];
        usedPaths.add(path);
        routes.push({
          path,
          name: `menu-${menu.id}`,
          component: component || notFound,
          meta: {
            title: menu.title,
            icon: menu.icon,
            permission: menu.permission,
            menuCode: menu.menuCode,
            componentPath: menu.component,
            missingComponent: !component,
            superAdmin: menu.component === "admin/settings/email/index",
          },
        });
      }
      visit(menu.children || [], path || parentPath);
    }
  };
  visit(menus || []);
  return routes;
}
