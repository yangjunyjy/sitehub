// Each page is identified by its path under views in both Vite and RBAC.
export function pagePath(file, options = {}) {
  const match = /^\.\.\/views\/(.+)\/index\.vue$/.exec(file);
  if (!match || match[1].split('/').includes('components')) return null;
  return options.path ?? (match[1] === 'home' ? '/' : '/' + match[1]);
}

export function publicNavigation(items) {
  return items.flatMap(item => {
    if (item.isVisible === false || item.isVisible === 0) return [];
    const children = publicNavigation(item.children || []);
    const segments = typeof item.path === 'string' ? item.path.split('/').slice(1) : [];
    const valid = item.path === '/' || (item.path?.startsWith('/') && segments.every(segment =>
      /^[\p{L}\p{N}_.~-]+$/u.test(segment) && segment !== '.' && segment !== '..'));
    return valid ? [item, ...children] : children;
  });
}

export function createPageRoutes(pages, layout, menus = [], options = {}) {
  const records = new Map();
  const byComponent = new Map();
  let missing;
  for (const [file, component] of Object.entries(pages)) {
    const settings = options[file.replace(/index\.vue$/, 'route.js')] || {};
    const path = pagePath(file, settings);
    if (path === null) continue;
    if (records.has(path)) throw new Error(`Duplicate page route: ${path}`);
    const page = file.slice('../views/'.length, -'/index.vue'.length);
    const meta = { page, notFound: path.includes('(.*)') };
    const record = { path: path.slice(1), name: page, component, meta };
    records.set(path, record);
    if (meta.notFound) missing = record;
    else byComponent.set(page + '/index', record);
  }
  for (const menu of menus) {
    if (!menu.component) continue;
    const page = byComponent.get(menu.component.replace(/\.vue$/, '')) || missing;
    // Preserve an optional document slug when a menu changes its URL.
    const suffix = page.path.match(/(?:\/:\w+\?)+$/)?.[0] || '';
    const path = (menu.path === '/' && suffix ? '' : menu.path) + suffix;
    if (records.get(path) === page) continue;
    records.set(path, { ...page, path: path.slice(1), name: 'menu:' + path });
    if (suffix) records.delete(menu.path);
  }
  return [{ path: '/', component: layout, children: [...records.values()] }];
}
