import axios from 'axios'
import { ElMessage } from 'element-plus'
import router from '../router/index.js'
import { useUserStore, resolveRoleFromPath } from '../stores/user'
import { API_BASE_URL } from './backend'

const normalizeBackendErrorMessage = (msg = '') => {
  const text = String(msg || '')
  if (!text) return text
  if (/field\s+validation.*answer.*required/i.test(text) || /key:\s*'answer'/i.test(text)) {
    return '您似乎没有做出任何回答'
  }
  if (/failed\s+to\s+transcribe\s+audio/i.test(text) || /empty\s+transcription\s+result/i.test(text)) {
    return '未识别到有效语音，请靠近麦克风并清晰作答后重试'
  }
  return text
}

const service = axios.create({
  baseURL: API_BASE_URL,
  timeout: 60000
})

const ROLE_AUTH_STRATEGY = new Map([
  ['enterprise', { loginPath: '/enterprise/login' }],
  ['university', { loginPath: '/university/login' }],
  ['student', { loginPath: '/student/login' }]
])

const resolveRoleContext = (path = '') => {
  const role = resolveRoleFromPath(path)
  const strategy = ROLE_AUTH_STRATEGY.get(role) || ROLE_AUTH_STRATEGY.get('student')
  return {
    role,
    loginPath: strategy.loginPath
  }
}

const getCurrentPathWithQuery = () => {
  if (typeof window === 'undefined') return '/'
  return `${window.location.pathname}${window.location.search || ''}${window.location.hash || ''}`
}

const normalizeRequestUrl = (url = '') => {
  const raw = String(url || '').trim()
  if (!raw) return ''
  return raw.startsWith('/') ? raw : `/${raw}`
}

const PUBLIC_ENDPOINTS = new Set(['/login', '/register'])

const isPublicEndpoint = (url = '') => {
  const path = normalizeRequestUrl(url)
  return PUBLIC_ENDPOINTS.has(path)
}

const shouldBypassAuthEnforcement = (config = {}) => {
  const currentPath = typeof window === 'undefined' ? '/' : window.location.pathname
  if (currentPath.endsWith('/login') || currentPath === '/') return true
  if (config?.headers?.['X-Skip-Auth'] === 'true') return true
  return isPublicEndpoint(config?.url)
}

const redirectToRoleLogin = (loginPath, fromPath) => {
  const redirect = encodeURIComponent(fromPath || '/')
  router.replace(`${loginPath}?redirect=${redirect}`).catch(() => {})
}

const enforceRoleLogoutAndRedirect = ({ role, loginPath }) => {
  const userStore = useUserStore()
  userStore.logout(role)
  redirectToRoleLogin(loginPath, getCurrentPathWithQuery())
}

service.interceptors.request.use(
  config => {
    const userStore = useUserStore()
    const currentPath = typeof window === 'undefined' ? '/' : window.location.pathname
    const { role, loginPath } = resolveRoleContext(currentPath)
    const token = userStore.getTokenByRole(role)

    if (token && !userStore.isTokenExpired(token)) {
      config.headers = config.headers || {}
      config.headers['Authorization'] = `Bearer ${token}`
      return config
    }

    if (!shouldBypassAuthEnforcement(config)) {
      enforceRoleLogoutAndRedirect({ role, loginPath })
      return Promise.reject(new Error('登录状态已失效，请重新登录'))
    }

    return config
  },
  error => {
    return Promise.reject(error)
  }
)

service.interceptors.response.use(
  response => {
    return response.data
  },
  error => {
    const res = error?.response
    if (res?.data?.error) {
      res.data.error = normalizeBackendErrorMessage(res.data.error)
    }
    if (error?.message) {
      error.message = normalizeBackendErrorMessage(error.message)
    }

    if (res?.status === 401) {
      const msg = (res.data && res.data.error) || ''
      if (/invalid token/i.test(msg) || /authorization/i.test(msg)) {
        const currentPath = typeof window === 'undefined' ? '/' : window.location.pathname
        const context = resolveRoleContext(currentPath)
        enforceRoleLogoutAndRedirect(context)
      }
    } else if (error?.code === 'ECONNABORTED') {
      error.message = '请求超时：服务响应超过 15 秒，请稍后重试。'
    }

    const isCanceled = axios.isCancel(error) || error?.code === 'ERR_CANCELED'
    if (res?.status !== 401 && !isCanceled) {
      ElMessage.error(error?.message || '请求失败，请稍后重试')
    }

    return Promise.reject(error)
  }
)

export default service
