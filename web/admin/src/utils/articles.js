export function articleListParams(filters, page = 1, pageSize = 10) {
  const params = {
    page: Math.max(1, Number(page) || 1),
    pageSize: Math.min(50, Math.max(1, Number(pageSize) || 10)),
  }
  if (filters.status === 'trash') params.trash = true
  if (filters.title?.trim()) params.title = filters.title.trim()
  if (['0', '1', '2'].includes(String(filters.status))) params.status = Number(filters.status)
  if (filters.categoryId) params.categoryID = String(filters.categoryId)
  if (filters.tagId) params.tagIDs = [String(filters.tagId)]
  return params
}
export function normalizeArticle(item) {
  return {
    ...item,
    title: item.title || '未命名文章',
    tags: Array.isArray(item.tags) ? item.tags.filter(Boolean) : [],
    isTop: Boolean(item.isTop ?? item.is_top),
    status: Number(item.status),
    updatedAt: item.updatedAt || item.updated_at || item.createdAt || item.created_at,
  }
}
export const articleError = (error, fallback = '请求失败，请重试') =>
  error.response?.data?.message || error.message || fallback
export function articleMarkdown(article) {
  if (typeof article.content !== 'string') throw new Error('文章正文未加载，请重试')
  return `# ${article.title || '未命名文章'}\n\n${article.summary ? `> ${article.summary}\n\n` : ''}${article.content}`
}
export const articleFilename = (article) =>
  `${String(article.title || 'article')
    .replace(/[\\/:*?"<>|\x00-\x1f]/g, '_')
    .slice(0, 120)}-${article.id}.md`
export function saveBlob(blob, name) {
  const url = URL.createObjectURL(blob),
    link = document.createElement('a')
  link.href = url
  link.download = name
  document.body.append(link)
  link.click()
  link.remove()
  setTimeout(() => URL.revokeObjectURL(url), 1000)
}
