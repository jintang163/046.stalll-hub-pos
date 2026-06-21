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

  const token = uni.getStorageSync('token')
  
  if (needLogin && !token) {
    uni.showToast({ title: '请先登录', icon: 'none' })
    setTimeout(() => {
      uni.reLaunch({ url: '/pages/login/index' })
    }, 1500)
    throw new Error('未登录')
  }

  return new Promise((resolve, reject) => {
    uni.request({
      url: BASE_URL + url,
      method,
      data,
      header: {
        'Content-Type': 'application/json',
        ...(token ? { Authorization: `Bearer ${token}` } : {}),
        ...header
      },
      timeout: 15000,
      success: (res: any) => {
        const result = res.data
        if (result.code === 0) {
          resolve(result.data)
        } else if (result.code === 401) {
          uni.removeStorageSync('token')
          uni.removeStorageSync('userInfo')
          uni.showToast({ title: '登录已过期，请重新登录', icon: 'none' })
          setTimeout(() => {
            uni.reLaunch({ url: '/pages/login/index' })
          }, 1500)
          reject(new Error('登录已过期'))
        } else {
          uni.showToast({ title: result.message || '请求失败', icon: 'none' })
          reject(new Error(result.message || '请求失败'))
        }
      },
      fail: (err: any) => {
        if (err.errMsg && err.errMsg.includes('timeout')) {
          uni.showToast({ title: '请求超时，请检查网络', icon: 'none' })
        } else if (err.errMsg && err.errMsg.includes('fail')) {
          uni.showToast({ title: '网络连接失败', icon: 'none' })
        }
        reject(err)
      }
    })
  })
}

export default request
export { BASE_URL }
