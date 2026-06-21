import request from '@/utils/request'

export function getOrderList(params) {
  return request({
    url: '/orders',
    method: 'get',
    params
  })
}

export function getOrderDetail(id) {
  return request({
    url: `/orders/${id}`,
    method: 'get'
  })
}

export function updateOrderStatus(id, status) {
  return request({
    url: `/orders/${id}/status`,
    method: 'put',
    data: { status }
  })
}

export function cancelOrder(id, reason) {
  return request({
    url: `/orders/${id}/cancel`,
    method: 'post',
    data: { reason }
  })
}
