import request from '@/utils/request'

export function getPointsRuleList(params) {
  return request({
    url: '/points-rules',
    method: 'get',
    params
  })
}

export function getPointsRule(id) {
  return request({
    url: `/points-rules/${id}`,
    method: 'get'
  })
}

export function createPointsRule(data) {
  return request({
    url: '/points-rules',
    method: 'post',
    data
  })
}

export function updatePointsRule(id, data) {
  return request({
    url: `/points-rules/${id}`,
    method: 'put',
    data
  })
}

export function deletePointsRule(id) {
  return request({
    url: `/points-rules/${id}`,
    method: 'delete'
  })
}

export function calculateEarnedPoints(data) {
  return request({
    url: '/points-rules/calculate-earn',
    method: 'post',
    data
  })
}

export function calculateRedeemDiscount(data) {
  return request({
    url: '/points-rules/calculate-redeem',
    method: 'post',
    data
  })
}
