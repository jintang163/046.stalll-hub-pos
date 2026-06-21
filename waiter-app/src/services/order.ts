import request from './request'
import type { OrderDetail, OrderListResponse, CreateOrderRequest } from '../types'

export const orderApi = {
  createOrder: (data: any): Promise<{ order_id: number; order_no: string; pay_amount: string }> =>
    request({
      url: '/orders',
      method: 'POST',
      data
    }),

  getOrders: (params: {
    store_id?: number
    order_status?: number
    pay_status?: number
    order_type?: string
    page?: number
    page_size?: number
  } = {}): Promise<OrderListResponse> =>
    request({
      url: `/orders?${new URLSearchParams(params as any).toString()}`,
      method: 'GET'
    }),

  getOrderDetail: (id: number): Promise<OrderDetail> =>
    request({
      url: `/orders/${id}`,
      method: 'GET'
    }),

  getOrderByNo: (orderNo: string): Promise<OrderDetail> =>
    request({
      url: `/orders/no/${orderNo}`,
      method: 'GET'
    }),

  updateStatus: (id: number, orderStatus: number): Promise<void> =>
    request({
      url: `/orders/${id}/status`,
      method: 'PUT',
      data: { order_status: orderStatus }
    }),

  cancelOrder: (id: number, reason: string): Promise<void> =>
    request({
      url: `/orders/${id}/cancel`,
      method: 'POST',
      data: { reason }
    }),

  refundOrder: (id: number, data: {
    refund_type: 'full' | 'partial'
    refund_amount: string
    refund_reason: string
    items?: { order_item_id: number; quantity: number }[]
  }): Promise<{ refund_id: number }> =>
    request({
      url: `/orders/${id}/refund`,
      method: 'POST',
      data
    })
}
