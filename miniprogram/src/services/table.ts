import request from './request'

export interface TableInfo {
  id: number
  storeId: number
  storeName: string
  tableNo: string
  tableType: string
  capacity: number
  area: string
  floor: number
  status: number
  message: string
}

export interface TableItem {
  id: number
  storeId: number
  tableNo: string
  name: string
  type: string
  capacity: number
  floor: number
  area: string
  qrCode: string
  qrCodeUrl: string
  status: number
  currentOrderId: number
  currentCustomerCount: number
  checkinTime: string
}

export interface ReservationItem {
  id: number
  storeId: number
  tableId: number
  memberId: number
  memberName: string
  memberPhone: string
  tableNo: string
  reserveDate: string
  reserveTime: string
  peopleCount: number
  status: number
  checkinStatus: number
  checkinTime: string
  cancelTime: string
  remark: string
  source: string
  orderId: number
  createdAt: string
}

export interface TimeSlot {
  time: string
  available: number
  total: number
  status: number
}

export interface QueueItem {
  id: number
  storeId: number
  queueType: string
  queueNumber: string
  sequence: number
  memberId: number
  memberName: string
  memberPhone: string
  peopleCount: number
  status: number
  callCount: number
  lastCallTime: string
  callTime: string
  arriveTime: string
  cancelTime: string
  aheadCount: number
  waitDuration: number
  remark: string
  tableId: number
  tableNo: string
  createdAt: string
}

export interface QueueStatus {
  queueNumber: string
  queueType: string
  status: number
  sequence: number
  aheadCount: number
  waitTime: number
  peopleCount: number
  createdAt: string
}

export interface QueueConfig {
  storeId: number
  smallPrefix: string
  smallCapacity: number
  mediumPrefix: string
  mediumCapacity: number
  largePrefix: string
  largeCapacity: number
  autoCall: boolean
  callInterval: number
  maxCallCount: number
  autoExpire: boolean
  expireMinutes: number
  voiceNotify: boolean
  smsNotify: boolean
}

export const tableApi = {
  scanQRCode: (scene: string): Promise<TableInfo> =>
    request.post('/tables/scan', { scene }),

  getAvailableTables: (storeId: number, peopleCount = 2): Promise<TableItem[]> =>
    request.get('/tables/available', { params: { store_id: storeId, people_count: peopleCount } }),

  checkin: (data: { tableId: number; peopleCount: number; reservationId?: number; memberId?: number }) =>
    request.post('/tables/checkin', data),

  checkout: (data: { tableId: number; orderId: number }) =>
    request.post('/tables/checkout', data),
}

export const reservationApi = {
  create: (data: {
    storeId: number
    tableId?: number
    memberId?: number
    memberName: string
    memberPhone: string
    tableNo?: string
    reserveDate: string
    reserveTime: string
    peopleCount: number
    remark?: string
    source?: string
  }): Promise<ReservationItem> => request.post('/reservations', data),

  cancel: (id: number) =>
    request.post(`/reservations/${id}/cancel`),

  get: (id: number): Promise<ReservationItem> =>
    request.get(`/reservations/${id}`),

  list: (params: {
    storeId?: number
    memberId?: number
    status?: number
    reserveDate?: string
    checkinStatus?: number
    keyword?: string
    pageNum?: number
    pageSize?: number
  }) => request.get('/reservations', { params }),

  getTimeSlots: (params: {
    storeId: number
    reserveDate: string
    peopleCount?: number
  }): Promise<TimeSlot[]> => request.get('/reservations/timeslots', { params }),
}

export const queueApi = {
  create: (data: {
    storeId: number
    queueType?: string
    memberId?: number
    memberName: string
    memberPhone: string
    peopleCount: number
    remark?: string
  }): Promise<QueueItem> => request.post('/queues', data),

  cancel: (data: { queueId: number; reason?: string }) =>
    request.post('/queues/cancel', data),

  getStatus: (params: {
    storeId: number
    queueType?: string
    memberId?: number
    queueNumber?: string
  }): Promise<QueueStatus> => request.get('/queues/status', { params }),

  getMy: (params: { memberId: number; storeId: number }): Promise<QueueItem[]> =>
    request.get('/queues/my', { params }),

  getWaitingCount: (storeId: number): Promise<{ small: number; medium: number; large: number }> =>
    request.get('/queues/waiting-count', { params: { store_id: storeId } }),

  getConfig: (storeId: number): Promise<QueueConfig> =>
    request.get('/queue-config', { params: { store_id: storeId } }),
}
