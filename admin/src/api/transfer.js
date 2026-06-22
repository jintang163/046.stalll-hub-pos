import request from '@/utils/request'

export function getTransfers(params) {
  return request({
    url: '/transfers',
    method: 'get',
    params
  })
}

export function getTransfer(id) {
  return request({
    url: `/transfers/${id}`,
    method: 'get'
  })
}

export function createTransfer(data) {
  return request({
    url: '/transfers',
    method: 'post',
    data
  })
}

export function confirmOutbound(id, data) {
  return request({
    url: `/transfers/${id}/confirm-outbound`,
    method: 'post',
    data
  })
}

export function startShipping(id, data) {
  return request({
    url: `/transfers/${id}/ship`,
    method: 'post',
    data
  })
}

export function receiveTransfer(id, data) {
  return request({
    url: `/transfers/${id}/receive`,
    method: 'post',
    data
  })
}

export function completeTransfer(id, data) {
  return request({
    url: `/transfers/${id}/complete`,
    method: 'post',
    data
  })
}

export function cancelTransfer(id, data) {
  return request({
    url: `/transfers/${id}/cancel`,
    method: 'post',
    data
  })
}

export function getLogisticsTrack(id) {
  return request({
    url: `/transfers/${id}/logistics`,
    method: 'get'
  })
}

export function getTransferItems(id) {
  return request({
    url: `/transfers/${id}/items`,
    method: 'get'
  })
}
