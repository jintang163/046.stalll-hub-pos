import request from '@/utils/request'

export function getStoreList(params) {
  return request({
    url: '/stores',
    method: 'get',
    params
  })
}

export function getStore(id) {
  return request({
    url: `/stores/${id}`,
    method: 'get'
  })
}

export function createStore(data) {
  return request({
    url: '/stores',
    method: 'post',
    data
  })
}

export function updateStore(id, data) {
  return request({
    url: `/stores/${id}`,
    method: 'put',
    data
  })
}

export function deleteStore(id) {
  return request({
    url: `/stores/${id}`,
    method: 'delete'
  })
}

export function getStores(params) {
  return request({
    url: '/store-map',
    method: 'get',
    params
  })
}

export const storeApi = {
  list: (params) => request.get('/stores', { params }),
  create: (data) => request.post('/stores', data),
  update: (id, data) => request.put(`/stores/${id}`, data),
  delete: (id) => request.delete(`/stores/${id}`),
  get: (id) => request.get(`/stores/${id}`)
}

export const storeMapApi = {
  getStores: (params) => request.get('/store-map', { params })
}
