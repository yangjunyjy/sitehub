// An empty group preserves files created before the catalog was introduced.
export function filterDownloadCatalog(packages, filters = {}) {
  const keyword = String(filters.q || "")
    .trim()
    .toLocaleLowerCase();
  return packages
    .filter(
      (item) =>
        (!filters.group || (item.groupId || "ungrouped") === filters.group) &&
        ["product", "edition", "platform", "arch", "version"].every(
          (field) => !filters[field] || item[field] === filters[field],
        ) &&
        (!keyword ||
          [
            item.product,
            item.edition,
            item.description,
            item.fileName,
            item.version,
            item.platform,
            item.arch,
          ]
            .join(" ")
            .toLocaleLowerCase()
            .includes(keyword)),
    )
    .sort((a, b) => (Number(a.sortOrder) || 0) - (Number(b.sortOrder) || 0));
}

export function downloadGroups(groups = [], packages = []) {
  const result = groups.map((group) => ({
    ...group,
    packages: packages.filter((item) => item.groupId === group.id),
  }));
  const ungrouped = packages.filter(
    (item) =>
      !item.groupId || !groups.some((group) => group.id === item.groupId),
  );
  if (ungrouped.length)
    result.push({
      id: "ungrouped",
      name: groups.length ? "其他下载" : "全部文件",
      description: "",
      packages: ungrouped,
    });
  return result.filter((group) => group.packages.length);
}
