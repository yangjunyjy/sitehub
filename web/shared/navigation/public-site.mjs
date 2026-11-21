import { externalAddress } from './external-address.mjs';

function menuPath(value, parent) {
  if (!value || typeof value !== 'string' || value.startsWith('//')) return '';
  const path = value.startsWith('/') ? value : parent.replace(/\/$/, '') + '/' + value;
  if (path === '/') return path;
  return path.slice(1).split('/').every(segment =>
    (/^[\p{L}\p{N}_~.-]+$/u.test(segment) && segment !== '.' && segment !== '..') || /^:\w+\??$/.test(segment)) ? path : '';
}

export function publicSite(pages, settings, menus) {
  const routes = [];
  const bindings = [];
  const claimed = new Set();

  function bind(node, base) {
    const file = './views/' + node.component + '.vue';
    if (!base || !pages[file] || file.split('/').includes('components')) return;
    const options = settings[file.replace(/index\.vue$/, 'menu.json')] || {};
    if (options.path?.includes('(.*)')) return;
    const params = options.params || '';
    const path = base.includes('/:') ? base : (base === '/' && params ? '' : base) + params;
    if (claimed.has(path)) return;
    claimed.add(path);
    const binding = { file, path };
    bindings.push(binding);
    return binding;
  }

  function visit(items, parent = '') {
    const nodes = [];
    for (const menu of items || []) {
      if (!menu || menu.isVisible === false || menu.isVisible === 0) continue;
      const external = menu.pageType === 'external';
      const path = external ? externalAddress(menu.externalUrl) : menuPath(menu.path, parent);
      nodes.push({
        menu, external, address: external ? path : '', binding: external ? null : bind(menu, path),
        children: visit(menu.children, external ? parent : path || parent),
      });
    }
    return nodes;
  }
  const tree = visit(menus);
  const views = Object.entries(pages).flatMap(([file, component]) => {
    const match = /^\.\/views\/(.+)\/index\.vue$/.exec(file);
    if (!match || match[1].split('/').includes('components')) return [];
    const page = match[1];
    const options = settings[file.replace(/index\.vue$/, 'menu.json')] || {};
    const base = options.path ?? '/' + page;
    const path = (base === '/' && options.params ? '' : base) + (options.params || '');
    return [{ file, page, component, path, bindings: bindings.filter(binding => binding.file === file) }];
  });

  // Reserve defaults of unbound views (including login) before assigning menu URLs.
  // Resolving the whole snapshot also allows two views to exchange their URLs.
  const reserved = new Set(views.filter(view => !view.bindings.length).map(view => view.path));
  for (const path of reserved) {
    for (const view of views) {
      if (!view.bindings.some(binding => binding.path === path)) continue;
      view.bindings = view.bindings.filter(binding => binding.path !== path);
      if (!view.bindings.length) reserved.add(view.path);
    }
  }
  for (const { page, component, path, bindings: configured } of views) {
    const locations = configured.length ? configured : [{path}];
    for (const [index, location] of locations.entries()) {
      routes.push({ path: location.path, name: index ? 'menu:' + location.path : page, component,
        meta: { page, notFound: path.includes('(.*)') } });
      const link = location.path.replace(/\/:\w+\?/g, '');
      location.address = link.includes('/:') ? '' : link || '/';
    }
  }

  function navigation(nodes) {
    return nodes.flatMap(({ menu, address, binding, external, children }) => {
      const path = binding?.address || address;
      const nested = navigation(children);
      if (!menu.name || (!path && !nested.length)) return [];
      return [{ id: menu.id || path || menu.name, label: menu.name, icon: menu.icon, path, external, children: nested }];
    });
  }
  const links = navigation(tree);
  const home = views.find(view => view.path === '/');
  if (home && !routes.some(record => record.path === '/')) routes.push({path: '/', redirect: {name: home.page}});
  return { routes, links };
}

export async function loadPublicSite(fetchMenus, pages, settings) {
  try {
    const { data } = await fetchMenus();
    if (!Array.isArray(data?.items)) throw new Error('菜单格式无效');
    return publicSite(pages, settings, data.items);
  } catch {
    return { ...publicSite(pages, settings, []), failed: true };
  }
}
