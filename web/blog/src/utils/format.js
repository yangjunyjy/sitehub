export const displayDate = (value) => {
  const date = String(value || "").slice(0, 10);
  return /^\d{4}-\d{2}-\d{2}$/.test(date) && !date.startsWith("0000")
    ? date
    : "日期待补充";
};
export const articleDate = (article) =>
  displayDate(
    article.published_at ||
      article.publishedAt ||
      article.created_at ||
      article.createdAt,
  );
export const safeURL = (value) => {
  if (typeof value !== "string" || !value.trim()) return "";
  try {
    const parsed = new URL(value, window.location.origin);
    return ["http:", "https:"].includes(parsed.protocol) &&
      !parsed.username &&
      !parsed.password
      ? parsed.href
      : "";
  } catch {
    return "";
  }
};
