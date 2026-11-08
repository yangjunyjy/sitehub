import request from '@/utils/request'
export const getEmailSettings = () => request({ url: '/rbac/settings/email', method: 'get' })
export const updateEmailSettings = (data) =>
  request({ url: '/rbac/settings/email', method: 'post', data })
