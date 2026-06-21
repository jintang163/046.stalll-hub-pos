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
    return request.get('/orders', {
      store_id: storeId,
      order_status: 1,
      cook_status: cookStatus,
      page_size: 200
    }).then(res => {
      const items: OrderItem[] = []
      res.list.forEach((order: any) => {
        if (order.items) {
          order.items.forEach((item: OrderItem) => {
            if (item.cook_status === cookStatus) {
              items.push({
                ...item,
                order_id: order.id,
                order_no: order.order_no,
                table_no: order.table_no,
                remark: order.remark
              })
            }
          })
        }
      })
      return items
    })
  },

  updateCookStatus: (data: CookStatusUpdateRequest): Promise<void> => {
    return request.put('/waiter/order-items/cook-status', data)
  },

  markItemsServed: (orderItemIds: number[]): Promise<void> => {
    return request.post('/waiter/order-items/serve', { order_item_ids: orderItemIds })
  }
}
