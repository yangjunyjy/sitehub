export const documentStatus = { draft: '草稿', published: '已发布', modified: '有未发布修改' }

// Always serialize only the backend document contract, excluding table metadata.
export function documentPayload(document) {
  return {
    id: document.id,
    slug: document.slug.trim(),
    title: document.title.trim(),
    summary: (document.summary || '').trim(),
    content: document.content || '',
  }
}

export function validateDocument(document) {
  const item = documentPayload(document)
  if (!item.title || !item.slug) throw new Error('请填写文档标题和访问标识')
  if (!/^[a-zA-Z0-9][a-zA-Z0-9_-]*$/.test(item.slug))
    throw new Error('访问标识仅支持英文字母、数字、短横线和下划线')
  // 占用状态由后端在保存事务中检查，避免旧列表或内部 ID 误报访问标识冲突。
  return item
}
