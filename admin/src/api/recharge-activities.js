import request from '@/utils/request'

export function getRechargeActivityList(params) {
  return request({
    url: '/recharge-activities',
    method: 'get',
    params
  })
}

export function getRechargeActivity(id) {
  return request({
    url: `/recharge-activities/${id}`,
    method: 'get'
  })
}

export function createRechargeActivity(data) {
  return request({
    url: '/recharge-activities',
    method: 'post',
    data
  })
}

export function updateRechargeActivity(id, data) {
  return request({
    url: `/recharge-activities/${id}`,
    method: 'put',
    data
  })
}

export function deleteRechargeActivity(id) {
  return request({
    url: `/recharge-activities/${id}`,
    method: 'delete'
  })
}

export function processRecharge(data) {
  return request({
    url: '/member-recharges',
    method: 'post',
    data
  })
}

export function getRechargeList(params) {
  return request({
    url: '/member-recharges',
    method: 'get',
    params
  })
}
