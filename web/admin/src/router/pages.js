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

export function filterMenuTree(menus = []) {
  return menus.flatMap((menu) => {
    if (menu.status !== undefined && Number(menu.status) !== 1) return [];
    if (menu.menu_type === 3) return [menu];
    const children = filterMenuTree(menu.children || []);
    if (menu.menu_type === 1)
      return children.length ? [{ ...menu, children }] : [];
    if (menu.menu_type !== 2) return [];
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
  settings = {},
  parentPath = "",
) {
  return menus
    .flatMap((menu) => {
      if (menu.status !== undefined && Number(menu.status) !== 1 || !menu.is_visible || ![1, 2].includes(menu.menu_type)) return [];
      if (!superAdmin && settings[componentKey(menu.component).replace(/index\.vue$/, "menu.json")]?.superAdmin)
        return [];
      if (menu.pageType === 'external') {
        const address = externalAddress(menu.externalUrl);
        return address ? [{ ...menu, externalUrl: address }] : [];
      }
      const path = menuPath(menu.path, parentPath);
      const children = navigationMenus(menu.children || [], superAdmin, settings, path || parentPath);
      if (menu.menu_type === 1 ? !children.length : !path) return [];
      return [{ ...menu, path, children }];
    })
    .sort((a, b) => (Number(a.sort) || 0) - (Number(b.sort) || 0));
}

export function menuRoutes(menus, modules, notFound, settings, reservedPaths) {
  const routes = [];
  const usedPaths = new Set(reservedPaths);
  const visit = (items, parentPath = "") => {
    for (const menu of items) {
      if (menu.menu_type === 3 || menu.pageType === 'external') continue;
      const path = menuPath(menu.path, parentPath);
      if (menu.menu_type === 2 && path && !usedPaths.has(path)) {
        const component = modules[componentKey(menu.component)];
        usedPaths.add(path);
        routes.push({
          path,
          name: `menu-${menu.id}`,
          component: component || notFound,
          meta: {
            title: menu.menu_name,
            icon: menu.icon,
            permission: menu.permission,
            menuCode: menu.menu_code,
            componentPath: menu.component,
            missingComponent: !component,
            superAdmin: settings[componentKey(menu.component).replace(/index\.vue$/, "menu.json")]?.superAdmin,
          },
        });
      }
      visit(menu.children || [], path || parentPath);
    }
  };
  visit(filterMenuTree(menus || []));
  return routes;
}
