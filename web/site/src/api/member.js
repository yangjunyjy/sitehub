import request from '@/utils/request'
export const sendMemberCode = email => request.post('/rbac/members/code', { email }).then(r => r.data)
export const registerMember = values => request.post('/rbac/members/register', values).then(r => r.data)
export const loginMember = async values => {
  const { data } = await request.post('/rbac/members/login', values)
  sessionStorage.setItem('memberToken', data.token)
  return data.user
}
export const getMemberProfile = () => request.get('/rbac/members/me').then(r => r.data)
export const logoutMember = async () => { await request.post('/rbac/members/logout', {}); sessionStorage.removeItem('memberToken') }
export const checkMemberDownload = id => request.get('/product/downloads/resolve', { params: { id } }).then(r => r.data)
