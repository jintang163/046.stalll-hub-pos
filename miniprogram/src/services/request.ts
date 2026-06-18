import Taro from '@tarojs/taro'

const BASE_URL = 'http://localhost:8080/api/v1'

interface RequestOptions {
  url: string
  method?: 'GET' | 'POST' | 'PUT' | 'DELETE'
  data?: any
  header?: any
  needLogin?: boolean
}

const request = async <T = any>(options: RequestOptions): Promise<T> => {
  const { url, method = 'GET', data, header = {}, needLogin = true } = options

  const token = Taro.getStorageSync('token')
  
  if (needLogin && !token) {
    Taro.showToast({ title: '请先登录', icon: 'none' })
    throw new Error('未登录')
  }

  try {
    const res = await Taro.request({
      url: BASE_URL + url,
      method,
      data,
      header: {
        'Content-Type': 'application/json',
        ...(token ? { Authorization: `Bearer ${token}` } : {}),
        ...header
      },
      timeout: 15000
    })

    const result = res.data as any
    
    if (result.code === 0) {
      return result.data
    } else if (result.code === 401) {
      Taro.removeStorageSync('token')
      Taro.removeStorageSync('userInfo')
      Taro.showToast({ title: '登录已过期，请重新登录', icon: 'none' })
      throw new Error('登录已过期')
    } else {
      Taro.showToast({ title: result.message || '请求失败', icon: 'none' })
      throw new Error(result.message || '请求失败')
    }
  } catch (error: any) {
    if (error.errMsg && error.errMsg.includes('timeout')) {
      Taro.showToast({ title: '请求超时，请检查网络', icon: 'none' })
    } else if (error.errMsg && error.errMsg.includes('fail')) {
      Taro.showToast({ title: '网络连接失败', icon: 'none' })
    }
    throw error
  }
}

export default request
