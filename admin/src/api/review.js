import request from '@/utils/request'

export function getReviewList(params) {
  return request({
    url: '/reviews',
    method: 'get',
    params
  })
}

export function getReviewDetail(id) {
  return request({
    url: `/reviews/${id}`,
    method: 'get'
  })
}

export function replyReview(id, data) {
  return request({
    url: `/reviews/${id}/reply`,
    method: 'post',
    data
  })
}

export function createWorkOrder(data) {
  return request({
    url: '/work-orders',
    method: 'post',
    data
  })
}

export function getWorkOrderList(params) {
  return request({
    url: '/work-orders',
    method: 'get',
    params
  })
}

export function getWorkOrderDetail(id) {
  return request({
    url: `/work-orders/${id}`,
    method: 'get'
  })
}

export function handleWorkOrder(id, data) {
  return request({
    url: `/work-orders/${id}/handle`,
    method: 'post',
    data
  })
}

export function getAlertList(params) {
  return request({
    url: '/review-alerts',
    method: 'get',
    params
  })
}

export function handleAlert(id, data) {
  return request({
    url: `/review-alerts/${id}/handle`,
    method: 'post',
    data
  })
}

export function getRatingList(params) {
  return request({
    url: '/review-ratings',
    method: 'get',
    params
  })
}

export function getRatingTrend(params) {
  return request({
    url: '/review-ratings/trend',
    method: 'get',
    params
  })
}

export function syncAllReviews(params) {
  return request({
    url: '/reviews/sync',
    method: 'post',
    params
  })
}

export function checkAlerts(params) {
  return request({
    url: '/review-alerts/check',
    method: 'post',
    params
  })
}
