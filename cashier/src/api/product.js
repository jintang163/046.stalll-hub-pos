import request from '@/utils/request'

export const batchSoldOut = (data) => {
  return request.post('/products/sold-out', data)
}

export const batchRestoreSoldOut = (data) => {
  return request.post('/products/sold-out/restore', data)
}

export const getSoldOutRecords = (params) => {
  return request.get('/products/sold-out/records', { params })
}
