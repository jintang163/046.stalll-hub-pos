import request from './request'

export interface Store {
  id: number
  name: string
  address: string
  phone: string
  image: string
  business_hours: string
  status: number
  longitude: number
  latitude: number
  distance?: number
}

export const getStoreList = () => {
  return request<Store[]>({
    url: '/stores',
    method: 'GET',
    needLogin: false
  })
}

export const getStoreDetail = (id: number) => {
  return request<Store>({
    url: `/stores/${id}`,
    method: 'GET',
    needLogin: false
  })
}

export const getCurrentStore = () => {
  return request<Store>({
    url: '/stores/current',
    method: 'GET',
    needLogin: false
  })
}
