export interface UserInfo {
  id: number
  username: string
  real_name: string
  phone: string
  store_id: number
  role: string
}

export interface LoginResponse {
  token: string
  user: UserInfo
}

export interface TableInfo {
  id: number
  store_id: number
  table_no: string
  name: string
  type: string
  capacity: number
  floor: number
  area: string
  status: number
  current_order_id: number
  current_customer_count: number
  checkin_time: string | null
  order_no: string
  order_amount: number
  item_count: number
  served_count: number
  display_status: 'idle' | 'occupied' | 'ordered' | 'all_served' | 'paid'
}

export interface OrderItem {
  id: number
  product_id: number
  sku_id: number
  category_id: number
  category_name: string
  product_name: string
  sku_name: string
  attribute_values: string
  image: string
  price: string
  quantity: number
  subtotal: string
  status: number
  print_status: number
  cook_status: number
}

export interface OrderDetail {
  id: number
  order_no: string
  store_id: number
  store_name: string
  member_id: number
  member_name: string
  table_no: string
  order_type: string
  total_amount: string
  discount_amount: string
  coupon_amount: string
  pay_amount: string
  pay_method: string
  pay_status: number
  pay_time: string | null
  order_status: number
  print_status: number
  points_earned: number
  points_used: number
  remark: string
  source: string
  items: OrderItem[]
  created_at: string
}

export interface OrderListResponse {
  list: OrderDetail[]
  total: number
  page: number
  size: number
}

export interface Product {
  id: number
  category_id: number
  name: string
  description: string
  main_image: string
  status: number
  is_hot: boolean
  is_recommend: boolean
  sort_order: number
  min_price: string
  max_price: string
  total_stock: number
  sku_count: number
}

export interface ProductDetail {
  id: number
  store_id: number
  category_id: number
  name: string
  description: string
  main_image: string
  images: string
  status: number
  is_hot: boolean
  is_recommend: boolean
  skus: ProductSKU[]
}

export interface ProductSKU {
  id: number
  product_id: number
  sku_code: string
  spec_name: string
  price: string
  original_price: string
  stock: number
  sold_count: number
  image: string
  status: number
}

export interface Category {
  id: number
  store_id: number
  name: string
  sort_order: number
  status: number
  description: string
}

export interface WaiterCall {
  id: number
  store_id: number
  table_id: number
  table_no: string
  content: string
  call_type: 'service' | 'water' | 'pay' | 'other'
  status: number
  handler_id: number
  handle_time: string | null
  created_at: string
}

export interface WaiterStats {
  total_tables: number
  idle_tables: number
  occupied_tables: number
  ordered_tables: number
  pending_calls: number
  pending_orders: number
}

export interface WSMessage {
  type: 'connected' | 'call_waiter' | 'order_update'
  call_id?: number
  store_id?: number
  table_id?: number
  table_no?: string
  content?: string
  call_type?: string
  order_id?: number
  order_no?: string
  message?: string
  created_at?: string
}
