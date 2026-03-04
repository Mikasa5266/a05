import axios from 'axios'
import { useUserStore } from '../stores/user'

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
    if (error?.code === 'ECONNABORTED') {
      error.message = '请求超时，请稍后重试（AI处理较慢时属于正常现象）'
    }
    return Promise.reject(error)
  }
)

export default service