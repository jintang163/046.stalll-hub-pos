import request from './request'

export interface RecommendItem {
  product_id: number
  product_name: string
  category_id: number
  main_image: string
  price: string
  sku_id: number
  score: number
  reason: string
  reason_type: 'cf' | 'hot' | 'user_history' | 'user_favorite' | 'cf_hot' | 'table_history' | 'time_hot'
}

export interface ScanOrderRecommendResponse {
  items: RecommendItem[]
  table_no: string
  store_id: number
  count: number
  timestamp: number
}

export const getCartRecommendations = (
  storeId: number,
  productIds: number[],
  count = 8,
  memberId?: number,
  userId?: number
) => {
  const params = new URLSearchParams()
  params.append('store_id', storeId.toString())
  params.append('count', count.toString())
  productIds.forEach(id => params.append('product_ids', id.toString()))
  if (memberId && memberId > 0) {
    params.append('member_id', memberId.toString())
  }
  if (userId && userId > 0) {
    params.append('user_id', userId.toString())
  }
  return request<RecommendItem[]>({
    url: `/recommendations/cart?${params.toString()}`,
    method: 'GET',
    needLogin: false
  })
}

export const getScanOrderRecommendations = (
  storeId: number,
  tableNo: string,
  count = 4
) => {
  const params = new URLSearchParams()
  params.append('store_id', storeId.toString())
  params.append('table_no', tableNo)
  params.append('count', count.toString())
  return request<ScanOrderRecommendResponse>({
    url: `/recommendations/scan-order?${params.toString()}`,
    method: 'GET',
    needLogin: false
  })
}
