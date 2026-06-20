import request from './request'

export interface OrderItem {
  id?: number
  order_no?: string
  product_id: number
  product_name: string
  sku_id: number
  sku_name: string
  attribute_ids: number[]
  attribute_names: string[]
  price: number
  quantity: number
  subtotal: number
  remark?: string
}

export interface Order {
  id?: number
  order_no: string
  store_id: number
  total_amount: number
  discount_amount: number
  actual_amount: number
  member_id?: number
  member_name?: string
  table_no?: string
  remark?: string
  status: number
  pay_status: number
  pay_method?: string
  created_at?: string
  updated_at?: string
  paid_at?: string
  items: OrderItem[]
}

export interface OrderCreateDTO {
  store_id: number
  items: OrderItem[]
  table_no?: string
  remark?: string
  coupon_id?: number
  member_coupon_id?: number
  member_id?: number
  order_type?: 'dine_in' | 'takeout' | 'delivery'
  points_used?: number
  source?: string
}

export const createOrder = (data: OrderCreateDTO) => {
  return request<Order>({
    url: '/orders',
    method: 'POST',
    data
  })
}

export const getOrders = (status?: number, page = 1, pageSize = 20) => {
  const params = new URLSearchParams()
  params.append('page', page.toString())
  params.append('page_size', pageSize.toString())
  if (status !== undefined) {
    params.append('status', status.toString())
  }
  
  return request<{ list: Order[]; total: number }>({
    url: `/orders?${params.toString()}`,
    method: 'GET'
  })
}

export const getOrderDetail = (orderNo: string) => {
  return request<Order>({
    url: `/orders/${orderNo}`,
    method: 'GET'
  })
}

export const cancelOrder = (orderNo: string, reason: string) => {
  return request({
    url: `/orders/${orderNo}/cancel`,
    method: 'POST',
    data: { reason }
  })
}

export const refundOrder = (orderNo: string, reason: string) => {
  return request({
    url: `/orders/${orderNo}/refund`,
    method: 'POST',
    data: { reason }
  })
}

export const getPaymentParams = (orderNo: string) => {
  return request<{
    appId: string
    timeStamp: string
    nonceStr: string
    package: string
    signType: string
    paySign: string
  }>({
    url: `/orders/${orderNo}/pay`,
    method: 'POST'
  })
}
