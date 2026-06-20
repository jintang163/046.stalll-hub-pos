import request from './request'

export interface Coupon {
  id: number
  store_id: number
  rule_key: string
  name: string
  type: string
  value: number
  min_amount: number
  max_discount: number
  total_count: number
  used_count: number
  per_user_limit: number
  validity_type: string
  validity_days: number
  start_time: string
  end_time: string
  applicable_type: string
  applicable_ids: string
  exclude_products: string
  stackable: boolean
  description: string
  status: number
  created_at: string
  exchange_product_id?: number
}

export interface MemberCoupon {
  id: number
  member_id: number
  coupon_id: number
  coupon: Coupon
  code: string
  status: number
  used_at?: string
  expire_at?: string
  order_id?: number
  created_at: string
}

export interface PromotionTier {
  id: number
  min_amount: number
  discount_value: number
}

export interface Promotion {
  id: number
  store_id: number
  name: string
  type: string
  min_amount: number
  discount_value: number
  max_discount: number
  tiers: PromotionTier[]
  start_time: string
  end_time: string
  applicable_type: string
  applicable_ids: string
  priority: number
  stackable: boolean
  description: string
  status: number
  created_at: string
}

export interface PromotionCalcResult {
  promotion_id?: number
  coupon_id?: number
  name: string
  type: string
  discount: number
}

export interface BestPromotionResponse {
  promotions: PromotionCalcResult[]
  total_discount: number
  final_amount: number
}

export interface ClaimCouponRequest {
  coupon_id: number
}

export const getAvailableCoupons = (amount: number, product_ids?: number[]) => {
  const params = new URLSearchParams()
  params.append('amount', String(amount))
  if (product_ids?.length) {
    params.append('product_ids', product_ids.join(','))
  }
  return request<MemberCoupon[]>({
    url: `/mini/coupons/available?${params.toString()}`,
    method: 'GET'
  })
}

export const getMyCoupons = (status?: number) => {
  const params = status !== undefined ? `?status=${status}` : ''
  return request<MemberCoupon[]>({
    url: `/mini/coupons/my${params}`,
    method: 'GET'
  })
}

export const claimCoupon = (data: ClaimCouponRequest) => {
  return request<MemberCoupon>({
    url: '/mini/coupons/claim',
    method: 'POST',
    data
  })
}

export const getActivePromotions = () => {
  return request<Promotion[]>({
    url: '/mini/promotions/active',
    method: 'GET'
  })
}

export const calculateBestCombination = (data: {
  store_id: number
  amount: number
  product_ids: number[]
  member_coupon_id?: number
  member_id?: number
}) => {
  return request<BestPromotionResponse>({
    url: '/mini/promotions/calculate',
    method: 'POST',
    data
  })
}
