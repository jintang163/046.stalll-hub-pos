import request from '@/utils/request'

export function getRevenueReport(params) {
  return request({
    url: '/analytics/revenue',
    method: 'get',
    params
  })
}

export function getHourlyTrend(params) {
  return request({
    url: '/analytics/hourly-trend',
    method: 'get',
    params
  })
}

export function getTopProducts(params) {
  return request({
    url: '/analytics/top-products',
    method: 'get',
    params
  })
}

export function importCostExcel(formData) {
  return request({
    url: '/analytics/cost/import',
    method: 'post',
    data: formData,
    headers: { 'Content-Type': 'multipart/form-data' }
  })
}

export function getCostList(params) {
  return request({
    url: '/analytics/cost/list',
    method: 'get',
    params
  })
}

export function getProfitReport(params) {
  return request({
    url: '/analytics/profit/report',
    method: 'get',
    params
  })
}

export function getProfitSummary(params) {
  return request({
    url: '/analytics/profit/summary',
    method: 'get',
    params
  })
}
