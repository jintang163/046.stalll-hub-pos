import axios, { type AxiosInstance } from 'axios'

const DEFAULT_API_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080/api/v1'

class HttpRequest {
  private instance: AxiosInstance
  private token: string = ''

  constructor(baseURL: string = DEFAULT_API_URL) {
    this.instance = axios.create({
      baseURL,
      timeout: 15000,
      headers: {
        'Content-Type': 'application/json'
      }
    })

    this.instance.interceptors.response.use(
      (response) => {
        const result = response.data
        if (result.code === 0) {
          return result.data
        } else {
          console.error('[API] Error:', result.message)
          return Promise.reject(new Error(result.message || '请求失败'))
        }
      },
      (error) => {
        console.error('[API] Network error:', error.message)
        return Promise.reject(error)
      }
    )
  }

  setToken(token: string) {
    this.token = token
    this.instance.defaults.headers.common['Authorization'] = `Bearer ${token}`
  }

  get<T = any>(url: string, params?: any): Promise<T> {
    return this.instance.get(url, { params })
  }

  post<T = any>(url: string, data?: any): Promise<T> {
    return this.instance.post(url, data)
  }

  put<T = any>(url: string, data?: any): Promise<T> {
    return this.instance.put(url, data)
  }

  delete<T = any>(url: string, data?: any): Promise<T> {
    return this.instance.delete(url, { data })
  }
}

const request = new HttpRequest()

export default request
export { HttpRequest }
