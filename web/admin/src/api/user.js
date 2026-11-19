import request from '@/utils/request'
import { uint64, allPages } from '../utils/native-api.mjs'
export const getUserList = (params, config = {}) => request.get('/rbac/users', { params: { page: 1, pageSize: 20, ...params }, ...config })
export const getUserDetail = id => request.get('/rbac/users/detail', { params: { id } })
export const createUser = data => request.post('/rbac/users/create', data)
export const updateUser = data => request.post('/rbac/users/update', { ...data, id: uint64(data.id) })
export const deleteUser = id => request.post('/rbac/users/delete', { id: uint64(id) })
export const assignUserRoles = data => request.post('/rbac/users/roles', { userId: uint64(data.userId), roleIds: data.roleIds.map(uint64) })
export const getallUser = () => allPages(getUserList)
