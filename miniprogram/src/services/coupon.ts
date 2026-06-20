import request from './request'

export interface Coupon {
  id: number
  store_id: number
  rule_key: string
  name: string
  type: 'fixed' | 'percentage' | 'exchange'
  value: number
  min_amount: number
  discount_rate: number
  max_discount: number
  total_count: number
  used_count: number
  per_user_limit: number
  validity_type: string
  validity_days: number
  start_time?: string
  end_time?: string
  applicable_type: string
  applicable_ids: number[]
  stackable: boolean
  description: string
  status: number
  created_at: string
  exchange_product_id?: number
}

export interface ClaimableCoupon extends Coupon {
  remaining_count: number
  claimed_count: number
  can_claim: boolean
}

export interface MemberCoupon {
  id: number
  store_id: number
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
  min_amount: number
  discount_amount: number
}

export interface Promotion {
  id: number
  store_id: number
  rule_key: string
  name: string
  type: 'full_reduction' | 'discount' | 'tiered'
  min_amount: number
  discount_amount: number
  discount_rate: number
  max_discount: number
  applicable_type: string
  applicable_ids: number[]
  start_time?: string
  end_time?: string
  tiers: PromotionTier[]
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

export const getClaimableCoupons = (store_id?: number) => {
  const params = store_id ? `?store_id=${store_id}` : ''
  return request<ClaimableCoupon[]>({
    url: `/mini/coupons/claimable${params}`,
    method: 'GET'
  })
}

export const getAvailableCoupons = (amount: number, product_ids?: number[], store_id?: number) => {
  const params = new URLSearchParams()
  params.append('amount', String(amount))
  if (store_id) params.append('store_id', String(store_id))
  if (product_ids?.length) params.append('product_ids', product_ids.join(','))
  return request<MemberCoupon[]>({
    url: `/mini/coupons/available?${params.toString()}`,
    method: 'GET'
  })
}

export const getMyCoupons = (status?: number) => {
  const params = status !== undefined ? `?status=${status}` : ''
  return request<{ list: MemberCoupon[]; total: number }>({
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

export const getActivePromotions = (store_id?: number) => {
  const params = store_id ? `?store_id=${store_id}` : ''
  return request<Promotion[]>({
    url: `/mini/promotions/active${params}`,
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
