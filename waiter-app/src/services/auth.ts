import request from './request'
import type { LoginResponse, UserInfo } from '../types'

export const authApi = {
  login: (username: string, password: string): Promise<LoginResponse> =>
    request({
      url: '/auth/login',
      method: 'POST',
      data: { username, password },
      needLogin: false
    }),

  logout: (): Promise<void> =>
    request({
      url: '/auth/logout',
      method: 'POST'
    }),

  getCurrentUser: (): Promise<UserInfo> =>
    request({
      url: '/auth/user',
      method: 'GET'
    })
}
