export interface OrderItem {
  id: number
  order_id: number
  product_id: number
  sku_id: number
  category_id: number
  category_name: string
  product_name: string
  sku_name: string
  spec_name: string
  attribute_values: string
  image: string
  price: string
  quantity: number
  subtotal: string
  status: number
  print_status: number
  cook_status: number
  table_no: string
  order_no: string
  remark: string
  created_at: string
}

export interface WSMessage {
  type: 'connected' | 'new_order' | 'order_update' | 'ping'
  order_id?: number
  order_no?: string
  store_id?: number
  items?: OrderItem[]
  message?: string
  created_at?: string
}

export interface CookStatusUpdateRequest {
  order_item_ids: number[]
  cook_status: number
}

export interface KitchenOrderItem extends OrderItem {
  waitingSeconds: number
  isOverdue: boolean
}

export interface AppConfig {
  storeId: number
  userId: number
  token: string
  apiBaseUrl: string
  wsUrl: string
  overdueMinutes: number
  voiceAlert: boolean
  flashAlert: boolean
}
