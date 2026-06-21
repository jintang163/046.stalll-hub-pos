import request from './request'
import type { OrderItem, CookStatusUpdateRequest } from '../types'

export const orderApi = {
  getPendingOrders: (params: {
    store_id: number
    cook_status?: number
    page?: number
    page_size?: number
  } = {}): Promise<{ list: OrderItem[]; total: number }> => {
    return request.get('/orders', {
      ...params,
      order_status: 1,
      page_size: params.page_size || 100
    })
  },

  getOrderItemsByCookStatus: (storeId: number, cookStatus: number): Promise<OrderItem[]> => {
    return request.get<{ list: OrderItem[]; total: number }>('/waiter/order-items/by-cook-status', {
      store_id: storeId,
      cook_status: cookStatus
    }).then(res => res.list)
  },

  updateCookStatus: (data: CookStatusUpdateRequest): Promise<void> => {
    return request.put('/waiter/order-items/cook-status', data)
  },

  markItemsServed: (orderItemIds: number[]): Promise<void> => {
    return request.post('/waiter/order-items/serve', { order_item_ids: orderItemIds })
  }
}
