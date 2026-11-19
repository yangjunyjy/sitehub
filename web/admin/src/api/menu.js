import request from '@/utils/request'
import { uint64 } from '../utils/native-api.mjs'
export const getMenuTree = (config = {}) => request.get('/rbac/menus/tree', config).then(r => ({ ...r, data: r.data.items || [] }))
const flatten = items => items.flatMap(item => [item, ...flatten(item.children || [])])
export const getAllMenus = config => getMenuTree(config).then(r => ({ ...r, data: flatten(r.data) }))
const input = data => ({ ...data, ...(data.id ? { id: uint64(data.id) } : {}), parent_id: uint64(data.parent_id) })
export const createMenu = data => request.post('/rbac/menus/create', input(data))
export const updateMenu = data => request.post('/rbac/menus/update', input(data))
export const deleteMenu = id => request.post('/rbac/menus/delete', { id: uint64(id) })
