export function pageRoutes(pages, paths = {}) {
  return Object.entries(pages)
    .filter(([file]) => !file.split("/").includes("components"))
    .map(([file, component]) => {
      const directory = file.slice("../views/".length, -"/index.vue".length);
      const path = paths[`../views/${directory}/route.mjs`] ?? `/${directory}`;
      return { name: `${directory}/index`, path, component };
    });
}
