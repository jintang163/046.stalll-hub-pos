import { defineStore } from 'pinia'
import { login, logout, getCurrentUser } from '@/api/auth'

export const useUserStore = defineStore('user', {
  state: () => ({
    token: localStorage.getItem('token') || '',
    userInfo: null
  }),

  actions: {
    async login(loginForm) {
      const res = await login(loginForm)
      this.token = res.token
      this.userInfo = res.user
      localStorage.setItem('token', res.token)
      return res
    },

    async logout() {
      try {
        await logout()
      } catch (e) {
        console.error(e)
      }
      this.token = ''
      this.userInfo = null
      localStorage.removeItem('token')
    },

    async getCurrentUser() {
      try {
        const res = await getCurrentUser()
        this.userInfo = res
        return res
      } catch (e) {
        this.token = ''
        localStorage.removeItem('token')
        throw e
      }
    }
  }
})
