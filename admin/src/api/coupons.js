import request from '@/utils/request'

export function getCouponList(params) {
  return request({
    url: '/coupons',
    method: 'get',
    params
  })
}

export function getCoupon(id) {
  return request({
    url: `/coupons/${id}`,
    method: 'get'
  })
}

export function createCoupon(data) {
  return request({
    url: '/coupons',
    method: 'post',
    data
  })
}

export function updateCoupon(id, data) {
  return request({
    url: `/coupons/${id}`,
    method: 'put',
    data
  })
}

export function deleteCoupon(id) {
  return request({
    url: `/coupons/${id}`,
    method: 'delete'
  })
}

export function issueCoupon(data) {
  return request({
    url: '/coupons/issue',
    method: 'post',
    data
  })
}
