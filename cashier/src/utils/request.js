import axios from 'axios'
import { ElMessage } from 'element-plus'

const getBaseURL = async () => {
  if (window.electronAPI) {
    const config = await window.electronAPI.app.getConfig()
    return config.apiBaseURL || 'http://localhost:8080/api/v1'
  }
  return 'http://localhost:8080/api/v1'
}

const request = axios.create({
  timeout: 15000
})

request.interceptors.request.use(
  async (config) => {
    config.baseURL = await getBaseURL()
    const token = localStorage.getItem('token')
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)

request.interceptors.response.use(
  (response) => {
    const res = response.data
    if (res.code !== 0) {
      ElMessage.error(res.message || '请求失败')
      return Promise.reject(new Error(res.message || '请求失败'))
    }
    return res.data
  },
  (error) => {
    if (!navigator.onLine) {
      ElMessage.error('网络连接已断开，请检查网络')
    } else if (error.response) {
      ElMessage.error(`服务器错误: ${error.response.status}`)
    } else {
      ElMessage.error(error.message || '请求超时')
    }
    return Promise.reject(error)
  }
)

export const checkNetwork = async () => {
  try {
    const baseURL = await getBaseURL()
    await axios.get(`${baseURL}/health`, { timeout: 3000 })
    return true
  } catch {
    return false
  }
}

export default request
