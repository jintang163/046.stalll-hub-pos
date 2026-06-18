import { create } from 'zustand'
import Taro from '@tarojs/taro'
import type { Store } from '../services/store'
import type { UserInfo } from '../services/auth'
import { getCurrentStore } from '../services/store'
import { getCurrentUser, isLogin, logout as authLogout } from '../services/auth'

interface AppStore {
  currentStore: Store | null
  user: UserInfo | null
  loading: boolean
  theme: 'light' | 'dark'

  init: () => Promise<void>
  setStore: (store: Store) => void
  setUser: (user: UserInfo) => void
  loadUser: () => Promise<void>
  logout: () => void
}

const getInitialStore = () => {
  try {
    return Taro.getStorageSync('currentStore') || null
  } catch {
    return null
  }
}

const getInitialUser = () => {
  try {
    return Taro.getStorageSync('userInfo') || null
  } catch {
    return null
  }
}

export const useAppStore = create<AppStore>((set, get) => ({
  currentStore: getInitialStore(),
  user: getInitialUser(),
  loading: false,
  theme: 'light',

  init: async () => {
    set({ loading: true })
    try {
      let store = get().currentStore
      if (!store) {
        try {
          store = await getCurrentStore()
          if (store) {
            Taro.setStorageSync('currentStore', store)
            set({ currentStore: store })
          }
        } catch {}
      }

      if (isLogin()) {
        try {
          const user = await getCurrentUser()
          Taro.setStorageSync('userInfo', user)
          set({ user })
        } catch {}
      }
    } finally {
      set({ loading: false })
    }
  },

  setStore: (store) => {
    Taro.setStorageSync('currentStore', store)
    set({ currentStore: store })
  },

  setUser: (user) => {
    Taro.setStorageSync('userInfo', user)
    set({ user })
  },

  loadUser: async () => {
    if (!isLogin()) return
    try {
      const user = await getCurrentUser()
      Taro.setStorageSync('userInfo', user)
      set({ user })
    } catch {}
  },

  logout: () => {
    authLogout()
    set({ user: null })
  }
}))
