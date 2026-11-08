import request from '@/utils/request'
import { uint64, allPages } from '../../../shared/native-api.mjs'
export const getTagList = (params, config = {}) => request.get('/blog/tags', { params: { page: 1, pageSize: 100, ...params }, ...config })
export const getAllTag = () => allPages(getTagList)
export const getTagDetail = id => request.get('/blog/tags/detail', { params: { id } })
export const createTag = data => request.post('/blog/tags/create', data)
export const updateTag = data => request.post('/blog/tags/update', { ...data, id: uint64(data.id) })
export const deleteTag = id => request.post('/blog/tags/delete', { id: uint64(id) })
