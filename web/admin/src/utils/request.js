import axios from 'axios'
import { ElMessage, ElMessageBox } from 'element-plus'
import router from '@/router'
import { queryString, stringifyJSON, viewValues } from '../../../shared/native-api.mjs'

const request = axios.create({ baseURL: '/v1', timeout: 30000, paramsSerializer: { serialize: queryString } })
let totpStatus
export const initRequestInterceptors = () => {
  totpStatus = request.get('/rbac/auth/totp', { silentError: true }).then(response => response.data.enabled)
  return totpStatus
}
const publicPosts = new Set(['/rbac/auth/login', '/rbac/auth/code', '/rbac/captcha/verify', '/rbac/auth/logout'])
request.interceptors.request.use(async config => {
  const token = sessionStorage.getItem('token')
  if (token) config.headers.Authorization = `Bearer ${token}`
  if (token && config.method !== 'get' && !publicPosts.has(config.url)) {
    const required = await (totpStatus || initRequestInterceptors())
    if (required) {
      const { value } = await ElMessageBox.prompt('请输入验证器中的六位动态验证码', '确认操作', {
        inputPattern: /^\d{6}$/, inputErrorMessage: '请输入六位数字', inputType: 'text',
      })
      config.headers['X-TOTP-Token'] = value
    }
  }
  if (config.data && !(config.data instanceof FormData) && typeof config.data === 'object') {
    config.data = stringifyJSON(config.data)
    config.headers['Content-Type'] = 'application/json'
  }
  return config
})
request.interceptors.response.use(response => {
  if (response.config.responseType === 'blob') return response.data
  const result = response.data
  if (result.code !== 0) return Promise.reject(new Error(result.message || '请求失败'))
  return { ...result, data: viewValues(result.data) }
}, error => {
  if (axios.isCancel(error)) return Promise.reject(error)
  const status = error.response?.status
  const message = status >= 500 ? '服务暂时不可用，请稍后重试' : error.response?.data?.message || error.message || '请求失败'
  error.message = message
  if (error.response?.data && status >= 500) error.response.data.message = message
  if (status === 401 && sessionStorage.getItem('token')) {
    sessionStorage.removeItem('token'); sessionStorage.removeItem('userInfo')
    sessionStorage.setItem('routeInitialized', 'false'); totpStatus = undefined
    router.replace('/login')
  }
  if (!error.config?.silentError) ElMessage.error(message)
  return Promise.reject(error)
})
export const downloadRequest = request
export default request
