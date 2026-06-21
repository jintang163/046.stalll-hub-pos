import request from './request'

export interface CallWaiterRequest {
  store_id: number
  table_id: number
  table_no: string
  call_type?: string
  content?: string
}

export interface WaiterCallItem {
  id: number
  store_id: number
  table_id: number
  table_no: string
  content: string
  call_type: string
  status: number
  handler_id: number
  handle_time: string
  created_at: string
}

export const waiterApi = {
  callWaiter: (data: CallWaiterRequest): Promise<WaiterCallItem> =>
    request<WaiterCallItem>({
      url: '/waiter/call',
      method: 'POST',
      data,
      needLogin: false
    }),

  getCalls: (storeId: number, status = 0): Promise<WaiterCallItem[]> =>
    request<WaiterCallItem[]>({
      url: '/waiter/calls',
      method: 'GET',
      data: { store_id: storeId, status }
    }),

  handleCall: (callId: number): Promise<void> =>
    request<void>({
      url: `/waiter/calls/${callId}/handle`,
      method: 'POST'
    }),
}
