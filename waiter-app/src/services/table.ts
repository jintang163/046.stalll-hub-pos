import request from './request'

export const tableApi = {
  listTables: (params: {
    store_id: number
    status?: number
    floor?: number
    area?: string
    page_num?: number
    page_size?: number
  }): Promise<{ list: any[]; total: number }> =>
    request({
      url: `/tables?${new URLSearchParams(params as any).toString()}`,
      method: 'GET'
    }),

  getOccupiedTables: (storeId: number): Promise<any[]> =>
    request({
      url: `/tables/occupied?store_id=${storeId}`,
      method: 'GET'
    }),

  getAvailableTables: (storeId: number, peopleCount = 2): Promise<any[]> =>
    request({
      url: `/tables/available?store_id=${storeId}&people_count=${peopleCount}`,
      method: 'GET',
      needLogin: false
    }),

  checkin: (data: {
    table_id: number
    people_count: number
    reservation_id?: number
    member_id?: number
  }): Promise<void> =>
    request({
      url: '/tables/checkin',
      method: 'POST',
      data
    }),

  checkout: (data: { table_id: number; order_id: number }): Promise<void> =>
    request({
      url: '/tables/checkout',
      method: 'POST',
      data
    }),

  listAreas: (storeId: number): Promise<any[]> =>
    request({
      url: `/table-areas?store_id=${storeId}`,
      method: 'GET'
    })
}
