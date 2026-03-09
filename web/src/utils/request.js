import axios from 'axios'
import { useUserStore } from '../stores/user'
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

service.interceptors.request.use(
  config => {
    const userStore = useUserStore()
    if (userStore.token) {
      config.headers['Authorization'] = `Bearer ${userStore.token}`
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
        const userStore = useUserStore()
        userStore.logout()
        const currentPath = window.location.pathname
        const portal = currentPath.startsWith('/enterprise')
          ? 'enterprise'
          : currentPath.startsWith('/university')
          ? 'university'
          : 'student'
        window.location.href = `/${portal}/login`
      }
    } else if (error?.code === 'ECONNABORTED') {
      error.message = '请求超时，请稍后重试（AI处理较慢时属于正常现象）'
    }
    return Promise.reject(error)
  }
)

export default service
