import request from '@/utils/request'
export function uploadImage(data) {
  data.set('kind', 'image'); data.delete('type')
  return request.post('/file/upload', data, { timeout: 180000 })
}
export function uploadAttachment(data, onProgress) {
  data.set('kind', 'attachment')
  return request.post('/file/upload', data, { timeout: 180000, onUploadProgress: e => onProgress?.(e.total ? Math.round(e.loaded / e.total * 100) : 0) })
}
