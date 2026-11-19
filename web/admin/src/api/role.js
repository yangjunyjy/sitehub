import request from '@/utils/request'
import { uint64 } from '../utils/native-api.mjs'
export const getRoleList = (params, config = {}) => request.get('/rbac/roles', { params: { page: 1, pageSize: 100, ...params }, ...config })
export const createRole = data => request.post('/rbac/roles/create', data)
export const updateRole = data => request.post('/rbac/roles/update', { ...data, id: uint64(data.id) })
export const deleteRole = id => request.post('/rbac/roles/delete', { id: uint64(id) })
export const assignRoleMenus = data => request.post('/rbac/roles/menus', { role_id: uint64(data.role_id), menu_ids: data.menu_ids.map(uint64) })
export const GetMenusByRoleID = id => request.get('/rbac/roles/menus', { params: { id } }).then(r => ({ ...r, data: r.data.items || [] }))
