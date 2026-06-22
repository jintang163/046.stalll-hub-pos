import request from '@/utils/request'

export function getReceiptAdList(params) {
  return request({
    url: '/receipt-ads',
    method: 'get',
    params
  })
}

export function getReceiptAd(id) {
  return request({
    url: `/receipt-ads/${id}`,
    method: 'get'
  })
}

export function createReceiptAd(data) {
  return request({
    url: '/receipt-ads',
    method: 'post',
    data
  })
}

export function updateReceiptAd(id, data) {
  return request({
    url: `/receipt-ads/${id}`,
    method: 'put',
    data
  })
}

export function deleteReceiptAd(id) {
  return request({
    url: `/receipt-ads/${id}`,
    method: 'delete'
  })
}

export function updateReceiptAdStatus(id, status) {
  return request({
    url: `/receipt-ads/${id}/status`,
    method: 'put',
    data: { status }
  })
}

export function getActiveReceiptAds(position) {
  return request({
    url: '/receipt-ads/active',
    method: 'get',
    params: { position }
  })
}

export function recordAdClick(data) {
  return request({
    url: '/receipt-ads/clicks',
    method: 'post',
    data
  })
}

export function getAdStats(params) {
  return request({
    url: '/receipt-ads/stats',
    method: 'get',
    params
  })
}
