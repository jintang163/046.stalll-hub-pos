import request from './request'

export interface TimeSlotPricing {
  id: number
  store_id: number
  name: string
  start_time: string
  end_time: string
  discount_type: 'percentage' | 'fixed'
  discount_value: number
  min_amount: number
  max_discount?: number
  applicable_days: number[]
  priority: number
  status: number
  created_at?: string
  updated_at?: string
}

export const getActiveTimeSlots = (storeId: number) => {
  return request<TimeSlotPricing[]>({
    url: `/time-slot-pricing/active/${storeId}`,
    method: 'GET'
  })
}

export const calculateTimeSlotPrice = (data: {
  store_id: number
  time: string
  amount: number
}) => {
  return request<{
    time_slot: TimeSlotPricing | null
    discount: number
  }>({
    url: '/time-slot-pricing/calculate',
    method: 'POST',
    data
  })
}
