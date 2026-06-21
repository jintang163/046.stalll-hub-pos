import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { UserInfo } from '../types'
import { authApi } from '../services/auth'

export const useUserStore = defineStore('user', () => {
  const token = ref<string>('')
  const userInfo = ref<UserInfo | null>(null)

  const login = async (username: string, password: string) => {
    const res = await authApi.login(username, password)
    token.value = res.token
    userInfo.value = res.user
    uni.setStorageSync('token', res.token)
    uni.setStorageSync('userInfo', JSON.stringify(res.user))
    return res
  }

  const logout = async () => {
    try {
      await authApi.logout()
    } catch (e) {
    }
    token.value = ''
    userInfo.value = null
    uni.removeStorageSync('token')
    uni.removeStorageSync('userInfo')
    uni.reLaunch({ url: '/pages/login/index' })
  }

  const restoreFromStorage = () => {
    const savedToken = uni.getStorageSync('token')
    const savedUserInfo = uni.getStorageSync('userInfo')
    if (savedToken) {
      token.value = savedToken
    }
    if (savedUserInfo) {
      try {
        userInfo.value = JSON.parse(savedUserInfo)
      } catch (e) {}
    }
  }

  const isLoggedIn = () => {
    return !!token.value
  }

  return {
    token,
    userInfo,
    login,
    logout,
    restoreFromStorage,
    isLoggedIn
  }
})
