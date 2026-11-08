import request from '@/utils/request'
export const logout = () => request.post('/rbac/auth/logout')
export const getUserInfo = () => request.get('/rbac/auth/me')
export const getPermissions = () => request.get('/rbac/auth/permissions')
export const rbacLogin = data => request.post('/rbac/auth/login', data)
export const sendEmailCode = (email, captchaKey, config = {}) => request.post('/rbac/auth/code', { email, captchaKey }, config)
export const changePassword = data => request.post('/rbac/auth/password', data)
export const getRBACMenu = () => request.get('/rbac/auth/menus').then(r => ({ ...r, data: r.data.items || [] }))
