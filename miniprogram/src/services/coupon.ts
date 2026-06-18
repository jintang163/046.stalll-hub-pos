import request from './request'

export interface Coupon {
  id: number
  name: string
  type: number
  value: number
  min_amount: number
  start_time: string
  end_time: string
  status: number
  description: string
}

export interface MemberCoupon {
  id: number
  coupon_id: number
  member_id: number
  status: number
  used_time?: string
  coupon: Coupon
}

export const getAvailableCoupons = (amount: number) => {
  return request<MemberCoupon[]>({
    url: `/coupons/available?amount=${amount}`,
    method: 'GET'
  })
}

export const getMyCoupons = (status?: number) => {
  const url = status !== undefined 
    ? `/coupons/my?status=${status}` 
    : '/coupons/my'
  return request<MemberCoupon[]>({
    url,
    method: 'GET'
  })
}
