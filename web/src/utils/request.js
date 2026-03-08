import axios from 'axios'
import { useUserStore } from '../stores/user'

const normalizeBackendErrorMessage = (msg = '') => {
  const text = String(msg || '')
  if (!text) return text
  if (/audio\s+data\s+is\s+empty/i.test(text) || /empty\s+transcription\s+result/i.test(text)) {
    return '未识别到有效语音，请靠近麦克风并清晰作答后重试'
  }
  if (/status:\s*401|authentication_error|authorization|unauthorized/i.test(text)) {
    return '语音转写鉴权失败（ASR Key 无效或无权限），请联系管理员检查 server/config.yaml 的 asr.api_key'
  }
  if (/maximum\s+content\s+size\s+limit|status:\s*413|payload\s+too\s+large/i.test(text)) {
    return '语音文件过大，请缩短单次语音回答时长后重试'
  }
  if (/invalid\s+file\s+format|unsupported\s+audio|invalid\s+audio|decode\s+audio/i.test(text)) {
    return '语音格式暂不受支持，请使用 Chrome/Edge 重试并确保允许麦克风权限'
  }
  if (/model\s+field|model\s+is\s+required|unknown\s+model|invalid\s+model/i.test(text)) {
    return '语音转写服务配置异常（模型参数无效），请联系管理员检查 ASR 配置'
  }
  if (/field\s+validation.*answer.*required/i.test(text) || /key:\s*'answer'/i.test(text)) {
    return '您似乎没有做出任何回答'
  }
  if (/failed\s+to\s+transcribe\s+audio/i.test(text)) {
    return '语音转写失败，请稍后重试；若持续失败，请检查后端 ASR 配置与日志'
  }
  return text
}

const service = axios.create({
  baseURL: import.meta.env.VITE_API_URL || 'http://localhost:8080/api/v1',
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
