import request from '@/utils/request'

export const tableApi = {
  create: (data) => request.post('/tables', data),
  update: (id, data) => request.put(`/tables/${id}`, data),
  delete: (id) => request.delete(`/tables/${id}`),
  get: (id) => request.get(`/tables/${id}`),
  list: (params) => request.get('/tables', { params }),
  batchCreate: (data) => request.post('/tables/batch', data),
  generateQRCode: (id) => request.post(`/tables/${id}/qrcode`),
  getOccupied: (params) => request.get('/tables/occupied', { params }),
  checkin: (data) => request.post('/tables/checkin', data),
  checkout: (data) => request.post('/tables/checkout', data),
}

export const tableAreaApi = {
  create: (data) => request.post('/table-areas', data),
  update: (id, data) => request.put(`/table-areas/${id}`, data),
  delete: (id) => request.delete(`/table-areas/${id}`),
  list: (params) => request.get('/table-areas', { params }),
}

export const reservationApi = {
  create: (data) => request.post('/reservations', data),
  update: (id, data) => request.put(`/reservations/${id}`, data),
  cancel: (id) => request.post(`/reservations/${id}/cancel`),
  checkin: (id) => request.post(`/reservations/${id}/checkin`),
  get: (id) => request.get(`/reservations/${id}`),
  list: (params) => request.get('/reservations', { params }),
  getTimeSlots: (params) => request.get('/reservations/timeslots', { params }),
}

export const queueApi = {
  create: (data) => request.post('/queues', data),
  call: (data) => request.post('/queues/call', data),
  callNext: (storeId, queueType) => request.post(`/queues/call-next/${storeId}?queue_type=${queueType}`),
  cancel: (data) => request.post('/queues/cancel', data),
  arrive: (data) => request.post('/queues/arrive', data),
  get: (id) => request.get(`/queues/${id}`),
  list: (params) => request.get('/queues', { params }),
  getStatus: (params) => request.get('/queues/status', { params }),
  getMy: (params) => request.get('/queues/my', { params }),
  getWaitingCount: (params) => request.get('/queues/waiting-count', { params }),
}

export const queueConfigApi = {
  get: (params) => request.get('/queue-config', { params }),
  save: (data) => request.post('/queue-config', data),
}

export const storeMapApi = {
  getStores: () => request.get('/store-map'),
}
