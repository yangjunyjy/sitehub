import request from '@/utils/request'
import { uint64 } from '../../../shared/native-api.mjs'
const revision = data => ({ ...data, revision: uint64(data.revision) })
export const changeDocMode = data => request.post('/product/docs/modes', revision(data))
export const getDocumentWorkspace = () => request.get('/product/admin/docs')
export const getAdminDocument = (id, version) => request.get('/product/admin/docs/document', { params: { id, version } })
export const getDocumentHistory = (id, version, historyId) => request.get('/product/admin/docs/history', { params: { id, version, historyId } })
export const saveDocument = data => request.post('/product/docs/save', { ...revision(data), groupID: uint64(data.groupID ?? data.groupId) })
export const deleteDocument = data => request.post('/product/docs/delete', revision(data))
export const changeDocVersion = data => request.post('/product/docs/versions', { ...revision(data), modeID: data.modeID ?? data.modeId })
export const changeDocGroup = data => request.post('/product/docs/groups', { ...revision(data), id: uint64(data.id) })
export const saveDocSettings = data => request.post('/product/docs/settings', revision(data))
