import request from '@/utils/request'

export const submitOrder = (order) => {
  return request.post('/orders', order)
}

export const submitOrdersBatch = (orders) => {
  return request.post('/orders/batch', { orders })
}

export const getOrderList = (params) => {
  return request.get('/orders', { params })
}

export const getOrderDetail = (orderNo) => {
  return request.get(`/orders/${orderNo}`)
}

export const updateOrderStatus = (orderNo, status) => {
  return request.put(`/orders/${orderNo}/status`, { status })
}

export const refundOrder = (orderNo, reason) => {
  return request.post(`/orders/${orderNo}/refund`, { reason })
}
