import request from '@/utils/request'
import { uint64 } from '../../../shared/native-api.mjs'
export const getCategoryList = (params, config = {}) => request.get('/blog/categories', { params: { page: 1, pageSize: 100, ...params }, ...config })
export const getCategoryTree = () => request.get('/blog/categories/tree').then(r => ({ ...r, data: r.data.items || [] }))
export const getCategoryDetail = id => request.get('/blog/categories/detail', { params: { id } })
const input = data => ({ ...data, ...(data.id ? { id: uint64(data.id) } : {}), parentId: uint64(data.parentId ?? data.parent_id), sortOrder: Number(data.sortOrder ?? data.sort_order ?? 0) })
export const createCategory = data => request.post('/blog/categories/create', input(data))
export const updateCategory = (id, data) => request.post('/blog/categories/update', input({ ...data, id }))
export const deleteCategory = id => request.post('/blog/categories/delete', { id: uint64(id) })
