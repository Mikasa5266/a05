import { defineStore } from 'pinia'
import { login, register, getUserProfile } from '../api/auth'

const ROLE_KEYS = {
  student: { token: 'student_token', userInfo: 'student_user_info' },
  enterprise: { token: 'enterprise_token', userInfo: 'enterprise_user_info' },
  university: { token: 'university_token', userInfo: 'university_user_info' }
}

const ROLE_NAMES = Object.keys(ROLE_KEYS)

const resolveRoleFromPath = (path = '') => {
  const pathname = String(path || '')
  if (pathname.startsWith('/enterprise')) return 'enterprise'
  if (pathname.startsWith('/university')) return 'university'
  return 'student'
}

const resolveCurrentRole = () => {
  if (typeof window === 'undefined') return 'student'
  return resolveRoleFromPath(window.location.pathname)
}

const safeParse = (value) => {
  try {
    return JSON.parse(value)
  } catch {
    return null
  }
}

const readRoleAuthFromStorage = (role) => {
  const keys = ROLE_KEYS[role] || ROLE_KEYS.student
  return {
    token: localStorage.getItem(keys.token) || '',
    userInfo: safeParse(localStorage.getItem(keys.userInfo) || 'null')
  }
}

const readAllRoleAuth = () => {
  return ROLE_NAMES.reduce((acc, role) => {
    acc[role] = readRoleAuthFromStorage(role)
    return acc
  }, {})
}

const decodeJwtPayload = (token = '') => {
  const parts = String(token || '').split('.')
  if (parts.length < 2) return null
  try {
    const normalized = parts[1].replace(/-/g, '+').replace(/_/g, '/')
    const decoded = atob(normalized)
    return JSON.parse(decoded)
  } catch {
    return null
  }
}

export const useUserStore = defineStore('user', {
  state: () => ({
    roleAuth: readAllRoleAuth()
  }),
  getters: {
    currentRole: () => resolveCurrentRole(),
    token(state) {
      return state.roleAuth[this.currentRole]?.token || ''
    },
    userInfo(state) {
      return state.roleAuth[this.currentRole]?.userInfo || null
    }
  },
  actions: {
    getRoleAuth(role = resolveCurrentRole()) {
      const safeRole = ROLE_NAMES.includes(role) ? role : 'student'
      const found = this.roleAuth[safeRole]
      if (found) return found
      this.roleAuth[safeRole] = { token: '', userInfo: null }
      return this.roleAuth[safeRole]
    },
    getTokenByRole(role = resolveCurrentRole()) {
      return this.getRoleAuth(role).token || ''
    },
    getUserInfoByRole(role = resolveCurrentRole()) {
      return this.getRoleAuth(role).userInfo || null
    },
    isTokenExpired(token = '') {
      if (!token) return true
      const payload = decodeJwtPayload(token)
      if (!payload?.exp) return false
      const nowSeconds = Math.floor(Date.now() / 1000)
      return nowSeconds >= Number(payload.exp)
    },
    hasValidTokenByRole(role = resolveCurrentRole()) {
      const token = this.getTokenByRole(role)
      return !!token && !this.isTokenExpired(token)
    },
    persistRoleAuth(role) {
      const safeRole = ROLE_NAMES.includes(role) ? role : 'student'
      const keys = ROLE_KEYS[safeRole]
      const auth = this.getRoleAuth(safeRole)
      if (auth.token) {
        localStorage.setItem(keys.token, auth.token)
      } else {
        localStorage.removeItem(keys.token)
      }
      if (auth.userInfo) {
        localStorage.setItem(keys.userInfo, JSON.stringify(auth.userInfo))
      } else {
        localStorage.removeItem(keys.userInfo)
      }
    },
    clearLegacyStorage() {
      localStorage.removeItem('token')
      localStorage.removeItem('userInfo')
    },
    setRoleAuth(role, payload = {}) {
      const safeRole = ROLE_NAMES.includes(role) ? role : 'student'
      this.roleAuth[safeRole] = {
        token: payload.token || '',
        userInfo: payload.userInfo || null
      }
      this.persistRoleAuth(safeRole)
    },
    async login(data) {
      const res = await login(data)
      const role = data?.role || res?.user?.role || resolveCurrentRole()
      this.setRoleAuth(role, {
        token: res.token,
        userInfo: res.user
      })
      this.clearLegacyStorage()
      return res
    },
    async register(data) {
      return register(data)
    },
    async getUserInfo(role = resolveCurrentRole()) {
      const safeRole = ROLE_NAMES.includes(role) ? role : 'student'
      const token = this.getTokenByRole(safeRole)
      if (!token || this.isTokenExpired(token)) {
        this.logout(safeRole)
        return
      }
      try {
        const res = await getUserProfile()
        this.setRoleAuth(safeRole, {
          token,
          userInfo: res.user
        })
      } catch (error) {
        this.logout(safeRole)
        throw error
      }
    },
    logout(role = resolveCurrentRole()) {
      const safeRole = ROLE_NAMES.includes(role) ? role : 'student'
      this.setRoleAuth(safeRole, {
        token: '',
        userInfo: null
      })
      this.clearLegacyStorage()
    },
    logoutAll() {
      ROLE_NAMES.forEach((role) => {
        this.setRoleAuth(role, {
          token: '',
          userInfo: null
        })
      })
      this.clearLegacyStorage()
    }
  }
})

export { ROLE_KEYS, ROLE_NAMES, resolveRoleFromPath, resolveCurrentRole }