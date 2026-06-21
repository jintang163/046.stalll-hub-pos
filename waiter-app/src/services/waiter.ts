import request from './request'
import type { TableInfo, WaiterStats, WaiterCall } from '../types'

export const waiterApi = {
  getStats: (storeId: number): Promise<WaiterStats> =>
    request({
      url: `/waiter/stats?store_id=${storeId}`,
      method: 'GET'
    }),

  getTables: (storeId: number, floor?: number, area?: string): Promise<TableInfo[]> =>
    request({
      url: `/waiter/tables?store_id=${storeId}${floor ? `&floor=${floor}` : ''}${area ? `&area=${area}` : ''}`,
      method: 'GET'
    }),

  updateItemCookStatus: (orderItemIds: number[], cookStatus: number): Promise<void> =>
    request({
      url: '/waiter/order-items/cook-status',
      method: 'PUT',
      data: { order_item_ids: orderItemIds, cook_status: cookStatus }
    }),

  markItemsServed: (orderItemIds: number[]): Promise<void> =>
    request({
      url: '/waiter/order-items/serve',
      method: 'POST',
      data: { order_item_ids: orderItemIds }
    }),

  addOrderItems: (orderId: number, items: any[]): Promise<void> =>
    request({
      url: `/waiter/orders/${orderId}/items`,
      method: 'POST',
      data: { order_id: orderId, items }
    }),

  getCalls: (storeId: number, status?: number): Promise<WaiterCall[]> =>
    request({
      url: `/waiter/calls?store_id=${storeId}${status ? `&status=${status}` : ''}`,
      method: 'GET'
    }),

  handleCall: (callId: number): Promise<void> =>
    request({
      url: `/waiter/calls/${callId}/handle`,
      method: 'POST'
    }),

  callWaiter: (data: {
    store_id: number
    table_id: number
    table_no: string
    content?: string
    call_type: 'service' | 'water' | 'pay' | 'other'
  }): Promise<{ call_id: number }> =>
    request({
      url: '/waiter/call',
      method: 'POST',
      data,
      needLogin: false
    })
}
