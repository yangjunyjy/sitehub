// The public list API omits article bodies and exposes published snapshots only.
// Collect every page before presenting archive counts; page one is never a complete archive.
export async function collectBlogIndex(
  fetchPage,
  { signal, onProgress = () => {} } = {},
) {
  const items = new Map();
  let expected;
  for (let page = 1; ; page++) {
    signal?.throwIfAborted();
    const response = await fetchPage(
      { page, pageSize: 50, sort: "latest" },
      { signal },
    );
    signal?.throwIfAborted();
    const data = response.data;
    if (
      !Array.isArray(data?.items) ||
      !Number.isSafeInteger(data.total) ||
      data.total < 0
    )
      throw new Error("文章索引响应无效");
    if (expected !== undefined && expected !== data.total)
      throw new Error("文章数量已更新，请重新加载归档");
    expected = data.total;
    for (const article of data.items) {
      const id = String(article.id);
      if (
        !/^[1-9]\d*$/.test(id) ||
        typeof article.title !== "string" ||
        items.has(id)
      )
        throw new Error("文章索引已变化，请重新加载归档");
      const date =
        article.published_at ||
        article.publishedAt ||
        article.created_at ||
        article.createdAt;
      items.set(id, {
        id,
        title: article.title,
        date: archiveDate(date),
        category: article.category?.id
          ? { id: String(article.category.id), name: article.category.name }
          : null,
        tags: (article.tags || [])
          .filter((tag) => tag.id && tag.name)
          .map((tag) => ({ id: String(tag.id), name: tag.name })),
      });
    }
    onProgress(items.size, expected);
    if (items.size === expected)
      return [...items.values()].sort(
        (a, b) =>
          b.date.localeCompare(a.date) ||
          b.id.localeCompare(a.id, undefined, { numeric: true }),
      );
    if (!data.items.length || items.size > expected)
      throw new Error("文章索引不完整，请重新加载归档");
  }
}

export function archiveDate(value) {
  const date = String(value || "").slice(0, 10);
  if (!/^\d{4}-\d{2}-\d{2}$/.test(date) || date.startsWith("0000")) return "";
  const parsed = new Date(date + "T00:00:00Z");
  return Number.isFinite(parsed.getTime()) &&
    parsed.toISOString().slice(0, 10) === date
    ? date
    : "";
}

export function archiveGroups(items) {
  const years = new Map();
  for (const article of items) {
    const year = article.date.slice(0, 4) || "undated",
      month = article.date.slice(5, 7) || "undated";
    if (!years.has(year))
      years.set(year, { year, count: 0, months: new Map() });
    const group = years.get(year);
    if (!group.months.has(month)) group.months.set(month, { month, items: [] });
    group.count++;
    group.months.get(month).items.push(article);
  }
  return [...years.values()].map((group) => ({
    ...group,
    months: [...group.months.values()],
  }));
}

export function blogTopics(items, kind) {
  const topics = new Map();
  for (const article of items) {
    const labels =
      kind === "categories"
        ? article.category
          ? [article.category]
          : []
        : article.tags;
    for (const label of new Map(
      labels.map((label) => [label.id, label]),
    ).values()) {
      if (!topics.has(label.id))
        topics.set(label.id, { ...label, count: 0, latest: article.title });
      topics.get(label.id).count++;
    }
  }
  return [...topics.values()].sort(
    (a, b) => b.count - a.count || a.name.localeCompare(b.name, "zh-CN"),
  );
}
