import request from '@/utils/request'
export const getWebsiteContent = () => request.get('/product/content').then(r => ({ ...r, data: { content: r.data } }))
export const getWebsiteDownloads = () => request.get('/product/admin/downloads').then(r => ({ ...r, data: { downloads: r.data } }))
export const getWebsiteStorage = () => request.get('/file/settings').then(r => ({ ...r, data: { storage: r.data } }))
export const updateWebsiteContent = value => request.post('/product/content', value)
export const updateWebsiteDownloads = value => request.post('/product/downloads', value)
export const updateWebsiteStorage = value => request.post('/file/settings', value)
export const uploadDownloadPackage = (file, onProgress) => {
  const data = new FormData(); data.append('file', file); data.append('kind', 'release')
  return request.post('/file/upload', data, { timeout: 600000, onUploadProgress: e => onProgress?.(e.total ? Math.round(e.loaded / e.total * 100) : 0) })
}
