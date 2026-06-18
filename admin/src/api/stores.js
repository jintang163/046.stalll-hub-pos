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
