import request from '@/utils/request'

export function savePlatformAuth(data) {
  return request({
    url: '/review/auth',
    method: 'post',
    data
  })
}

export function getPlatformAuth(storeId, platform) {
  return request({
    url: '/review/auth',
    method: 'get',
    params: { store_id: storeId, platform }
  })
}

export function listPlatformAuths(storeId) {
  return request({
    url: '/review/auths',
    method: 'get',
    params: { store_id: storeId }
  })
}

export function syncReviews(data) {
  return request({
    url: '/review/sync',
    method: 'post',
    data
  })
}

export function syncAllReviews() {
  return request({
    url: '/review/sync-all',
    method: 'post'
  })
}

export function getRatingList(params) {
  return request({
    url: '/review/ratings',
    method: 'get',
    params
  })
}

export function getRatingTrend(params) {
  return request({
    url: '/review/ratings/trend',
    method: 'get',
    params
  })
}

export function getReviewList(params) {
  return request({
    url: '/review/reviews',
    method: 'get',
    params
  })
}

export function getReviewDetail(id) {
  return request({
    url: `/review/reviews/${id}`,
    method: 'get'
  })
}

export function replyReview(id, data) {
  return request({
    url: `/review/reviews/${id}/reply`,
    method: 'post',
    data
  })
}

export function createWorkOrder(data) {
  return request({
    url: '/review/work-orders',
    method: 'post',
    data
  })
}

export function getWorkOrderList(params) {
  return request({
    url: '/review/work-orders',
    method: 'get',
    params
  })
}

export function getWorkOrderDetail(id) {
  return request({
    url: `/review/work-orders/${id}`,
    method: 'get'
  })
}

export function handleWorkOrder(id, data) {
  return request({
    url: `/review/work-orders/${id}/handle`,
    method: 'post',
    data
  })
}

export function getAlertList(params) {
  return request({
    url: '/review/alerts',
    method: 'get',
    params
  })
}

export function handleAlert(id, data) {
  return request({
    url: `/review/alerts/${id}/handle`,
    method: 'post',
    data
  })
}

export function checkAlerts() {
  return request({
    url: '/review/alerts/check',
    method: 'post'
  })
}
