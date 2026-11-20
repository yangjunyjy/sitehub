import { externalAddress } from './external-address.mjs';
export function navigationLinks(items, routes) {
  const available = new Map(routes.map(({ name, path }) => [name, path]));
  const seen = new Set(),
    result = [];
  function visit(nodes, depth = 0) {
    if (depth > 12 || !Array.isArray(nodes)) return;
    for (const node of nodes) {
      if (
        !node ||
        node.isVisible === false ||
        node.is_visible === false ||
        Number(node.status) === 2
      )
        continue;
      const component = String(node.component || "")
        .replace(/\.vue$/, "")
        .replace(/^\/+/, "");
      const external = node.pageType === 'external';
      const path = external ? externalAddress(node.externalUrl) : available.get(component);
      const label = node.menuName || node.menu_name || node.name || node.title;
      if (
        label &&
        path &&
        (external || !path.includes(":")) &&
        !seen.has(path)
      ) {
        seen.add(path);
        result.push({ path, label: String(label), ...(external && { external: true }) });
      }
      visit(node.children, depth + 1);
    }
  }
  visit(items);
  return result;
}
