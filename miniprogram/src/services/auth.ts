import request from './request'
import Taro from '@tarojs/taro'

export interface UserInfo {
  id: number
  nickname: string
  avatar: string
  phone: string
  points: number
  level: number
  level_name: string
}

export const loginByCode = (code: string) => {
  return request<{ token: string; user: UserInfo }>({
    url: '/auth/wx-login',
    method: 'POST',
    data: { code },
    needLogin: false
  })
}

export const login = (phone: string, code: string) => {
  return request<{ token: string; user: UserInfo }>({
    url: '/auth/login',
    method: 'POST',
    data: { phone, code },
    needLogin: false
  })
}

export const getCurrentUser = () => {
  return request<UserInfo>({
    url: '/auth/me',
    method: 'GET'
  })
}

export const saveUserInfo = (user: UserInfo) => {
  Taro.setStorageSync('userInfo', user)
}

export const getUserInfo = (): UserInfo | null => {
  return Taro.getStorageSync('userInfo') || null
}

export const saveToken = (token: string) => {
  Taro.setStorageSync('token', token)
}

export const getToken = (): string => {
  return Taro.getStorageSync('token') || ''
}

export const isLogin = (): boolean => {
  return !!getToken()
}

export const logout = () => {
  Taro.removeStorageSync('token')
  Taro.removeStorageSync('userInfo')
}
