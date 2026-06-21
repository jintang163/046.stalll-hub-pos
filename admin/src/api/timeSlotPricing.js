import request from '@/utils/request'

export function getTimeSlotPricingList(params) {
  return request({
    url: '/time-slot-pricing',
    method: 'get',
    params
  })
}

export function getTimeSlotPricing(id) {
  return request({
    url: `/time-slot-pricing/${id}`,
    method: 'get'
  })
}

export function createTimeSlotPricing(data) {
  return request({
    url: '/time-slot-pricing',
    method: 'post',
    data
  })
}

export function updateTimeSlotPricing(id, data) {
  return request({
    url: `/time-slot-pricing/${id}`,
    method: 'put',
    data
  })
}

export function deleteTimeSlotPricing(id) {
  return request({
    url: `/time-slot-pricing/${id}`,
    method: 'delete'
  })
}

export function getActiveTimeSlots(storeId) {
  return request({
    url: `/time-slot-pricing/active/${storeId}`,
    method: 'get'
  })
}

export function calculateTimeSlotPrice(data) {
  return request({
    url: '/time-slot-pricing/calculate',
    method: 'post',
    data
  })
}
