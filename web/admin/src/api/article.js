import request from '@/utils/request'
import { createUUID } from '@/utils/uuid'
import { uint64 } from '../utils/native-api.mjs'
export const getArticleList = (params, config = {}) => request.get('/blog/admin/articles', { params, ...config })
export const getArticleDetail = (id, config = {}) => request.get('/blog/admin/articles/detail', { params: { id }, ...config })
export const getArticleOptions = () => request.get('/blog/admin/articles/options')
const saveDraft = ({ id, version, request_id, ...input }) => request.post('/blog/admin/articles/draft', {
  id: uint64(id), version: uint64(version), requestID: request_id,
  input: { ...input, category_id: uint64(input.category_id), tag_ids: (input.tag_ids || []).map(uint64) },
})
export const createArticle = saveDraft
export const updateArticle = saveDraft
export const articleAction = (action, data) => request.post('/blog/admin/articles/action', {
  id: uint64(data.id), version: uint64(data.version), requestID: data.request_id, action,
  ...(data.scheduled_at ? { scheduledAt: data.scheduled_at } : {}),
  ...(data.revision_id ? { revisionID: uint64(data.revision_id) } : {}),
  ...(data.is_top !== undefined ? { isTop: data.is_top } : {}),
})
export const deleteArticle = (id, version) => articleAction('delete', { id, version, request_id: createUUID() })
export const toggleArticleTop = (id, isTop, version) => articleAction('top', { id, version, is_top: isTop, request_id: createUUID() })
export const getArticleHistory = id => request.get('/blog/admin/articles/history', { params: { id } }).then(r => ({ ...r, data: r.data.items || [] }))
export const getArticleRevision = (id, revisionId) => request.get('/blog/admin/articles/revision', { params: { id, revisionId } })
export const uploadArticleAsset = (file, onProgress) => {
  const data = new FormData(); data.append('file', file)
  return request.post('/blog/assets/upload', data, { timeout: 180000, onUploadProgress: e => onProgress?.(e.total ? Math.round(e.loaded / e.total * 100) : 0) })
}
