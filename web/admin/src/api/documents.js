import request from '@/utils/request'
import { uint64 } from '../utils/native-api.mjs'

const base = '/product/admin/manuals'
export const getDocumentWorkspace = (params = {}) => request.get(base + '/workspace', { params })
export const getAdminDocument = (id, version) =>
  request.get(base + '/document', { params: { id, version } })
export const getDocumentHistory = (id, version) =>
  request.get(base + '/history', { params: { id, version } })
export const getDocumentRevision = (id, version, historyId) =>
  request.get(base + '/history/detail', { params: { id, version, historyId } })

const write = (path, data) =>
  request.post(base + path, { ...data, revision: uint64(data.revision) })
export const saveDocSettings = (data) => write('/settings/save', data)
export const saveMode = (data) => write('/modes/save', data)
export const defaultMode = (data) => write('/modes/default', data)
export const deleteMode = (data) => write('/modes/delete', data)
export const saveVersion = (data) => write('/versions/save', data)
export const publishVersion = (data) => write('/versions/publish', data)
export const withdrawVersion = (data) => write('/versions/withdraw', data)
export const defaultVersion = (data) => write('/versions/default', data)
export const deleteVersion = (data) => write('/versions/delete', data)
export const saveGroup = (data) => write('/groups/save', { ...data, id: uint64(data.id) })
export const moveGroup = (data) => write('/groups/move', { ...data, id: uint64(data.id) })
export const deleteGroup = (data) => write('/groups/delete', { ...data, id: uint64(data.id) })
export const saveDocument = (data) =>
  write('/documents/save', { ...data, groupId: uint64(data.groupId) })
export const publishDocument = (data) => write('/documents/publish', data)
export const withdrawDocument = (data) => write('/documents/withdraw', data)
export const deleteDocument = (data) => write('/documents/delete', data)
