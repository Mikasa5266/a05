import { defineStore } from 'pinia'
import { login, register, getUserProfile } from '../api/auth'

export const useUserStore = defineStore('user', {
  state: () => ({
    token: localStorage.getItem('token') || '',
    userInfo: null
  }),
  actions: {
    async login(data) {
      const res = await login(data)
      this.token = res.token
      this.userInfo = res.user
      localStorage.setItem('token', res.token)
    },
    async register(data) {
      return register(data)
    },
    async getUserInfo() {
      if (!this.token) return
      try {
        const res = await getUserProfile()
        this.userInfo = res.user
      } catch (error) {
        this.logout()
      }
    },
    logout() {
      this.token = ''
      this.userInfo = null
      localStorage.removeItem('token')
    }
  }
})