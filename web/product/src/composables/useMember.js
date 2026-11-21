import { reactive, readonly } from 'vue'
import { getMemberProfile, loginMember, logoutMember } from '@/api/member'

const state = reactive({ user: null, ready: false })
let pending
export const restoreMember = async () => {
  if (pending) return pending
  pending = getMemberProfile().then((user) => { state.user = user; return user }).catch((error) => {
    if (error.status === 401) { state.user = null; return null }
    throw error
  }).finally(() => { state.ready = true; pending = null })
  return pending
}
export const useMember = () => ({
  member: readonly(state),
  restoreMember,
  clearMember() { state.user = null; state.ready = true },
  async login(values) { const user = await loginMember(values); state.user = user; state.ready = true; return user },
  async logout() { await logoutMember(); state.user = null; state.ready = true },
})

// Only allow local product routes as post-login destinations.
export const memberDestination = (value, router) => {
  const account = { name: 'account' }
  if (typeof value !== 'string' || !value.startsWith('/') || value.startsWith('//') || /[\\\s]/.test(value)) return account
  const destination = router.resolve(value)
  if (destination.meta.notFound || !destination.name || ['login', 'register'].includes(destination.name)) return account
  return destination.fullPath
}
