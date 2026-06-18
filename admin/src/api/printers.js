import request from '@/utils/request'

export function getPrinterList(params) {
  return request({
    url: '/printers',
    method: 'get',
    params
  })
}

export function getPrinter(id) {
  return request({
    url: `/printers/${id}`,
    method: 'get'
  })
}

export function createPrinter(data) {
  return request({
    url: '/printers',
    method: 'post',
    data
  })
}

export function updatePrinter(id, data) {
  return request({
    url: `/printers/${id}`,
    method: 'put',
    data
  })
}

export function deletePrinter(id) {
  return request({
    url: `/printers/${id}`,
    method: 'delete'
  })
}

export function testPrint(id) {
  return request({
    url: `/printers/${id}/test`,
    method: 'post'
  })
}
