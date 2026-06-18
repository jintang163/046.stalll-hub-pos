import request from '@/utils/request'

export function getDailyReport(params) {
  return request({
    url: '/reports/daily',
    method: 'get',
    params
  })
}

export function getProductSalesReport(params) {
  return request({
    url: '/reports/product-sales',
    method: 'get',
    params
  })
}

export function getCategorySalesReport(params) {
  return request({
    url: '/reports/category-sales',
    method: 'get',
    params
  })
}

export function getTimeSlotReport(params) {
  return request({
    url: '/reports/time-slot',
    method: 'get',
    params
  })
}

export function getPaymentStatsReport(params) {
  return request({
    url: '/reports/payment-stats',
    method: 'get',
    params
  })
}
