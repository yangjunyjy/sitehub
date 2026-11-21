import request from '@/utils/request'
export const getWebsiteContent = () => request.get('/product/content').then(r => ({ ...r, data: { content: r.data } }))
export const getWebsiteStorage = () => request.get('/file/settings').then(r => ({ ...r, data: { storage: r.data } }))
export const updateWebsiteContent = value => request.post('/product/content', value)
export const updateWebsiteStorage = value => request.post('/file/settings', value)
