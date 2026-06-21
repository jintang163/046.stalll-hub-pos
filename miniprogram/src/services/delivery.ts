import request from './request'

export type DeliveryType = 'self' | 'meituan' | 'eleme'
export type OrderType = 'dine_in' | 'takeout' | 'delivery' | 'pickup'

export interface DeliveryOrder {
  id: number
  order_id: number
  order_no: string
  store_id: number
  delivery_type: DeliveryType
  delivery_status: number
  rider_id: number
  rider_name: string
  rider_phone: string
  rider_lng: number
  rider_lat: number
  delivery_fee: number
  distance: number
  duration: number
  sender_name: string
  sender_phone: string
  sender_address: string
  receiver_name: string
  receiver_phone: string
  receiver_address: string
  platform_order_id: string
  platform_type: string
  estimated_time: string | null
  picked_up_at: string | null
  delivered_at: string | null
  created_at: string
}

export interface DeliveryTracking {
  order_no: string
  delivery_type: DeliveryType
  delivery_status: number
  rider_id: number
  rider_name: string
  rider_phone: string
  rider_lng: number
  rider_lat: number
  distance: number
  duration: number
  receiver_address: string
  sender_address: string
  estimated_time: string | null
  trackings: TrackingPoint[]
}

export interface TrackingPoint {
  lng: number
  lat: number
  speed: number
  timestamp: number
}

export interface RiderLocation {
  rider_id: number
  rider_name: string
  lng: number
  lat: number
  speed: number
  heading: number
  updated_at: string
}

export interface PickupCodeInfo {
  order_id: number
  code: string
  status: number
  expired_at?: string
}

export interface RoutePlanResult {
  distance: number
  duration: number
  route: string
  fee: number
}

export interface GeocodeResult {
  lng: number
  lat: number
  formatted: string
}

export const deliveryStatusMap: Record<number, { text: string; color: string }> = {
  0: { text: '待接单', color: '#e6a23c' },
  1: { text: '骑手已接单，取餐中', color: '#409eff' },
  2: { text: '配送中', color: '#409eff' },
  3: { text: '已送达', color: '#67c23a' },
  4: { text: '已取消', color: '#909399' },
}

export const orderTypeMap: Record<OrderType, { label: string; icon: string }> = {
  dine_in: { label: '堂食', icon: '🍽️' },
  takeout: { label: '外带', icon: '🥡' },
  delivery: { label: '外卖配送', icon: '🛵' },
  pickup: { label: '到店自提', icon: '🏪' },
}

export const createDeliveryOrder = (data: {
  order_id: number
  delivery_type: DeliveryType
  receiver_name: string
  receiver_phone: string
  receiver_address: string
  receiver_lng?: number
  receiver_lat?: number
  sender_name?: string
  sender_phone?: string
  sender_address?: string
  sender_lng?: number
  sender_lat?: number
}) => {
  return request<DeliveryOrder>({
    url: '/delivery',
    method: 'POST',
    data
  })
}

export const getDeliveryByOrder = (orderId: number) => {
  return request<DeliveryOrder>({
    url: `/delivery/order/${orderId}`,
    method: 'GET',
    needLogin: false
  })
}

export const getDeliveryTracking = (orderId: number) => {
  return request<DeliveryTracking>({
    url: `/delivery/tracking/${orderId}`,
    method: 'GET',
    needLogin: false
  })
}

export const getRiderLocation = (riderId: number) => {
  return request<RiderLocation>({
    url: `/riders/${riderId}/location`,
    method: 'GET'
  })
}

export const generatePickupCode = (orderId: number, storeId: number) => {
  return request<PickupCodeInfo>({
    url: '/pickup/code',
    method: 'POST',
    data: { order_id: orderId, store_id: storeId }
  })
}

export const verifyPickupCode = (code: string, storeId: number) => {
  return request<PickupCodeInfo>({
    url: '/pickup/verify',
    method: 'POST',
    data: { code, store_id: storeId }
  })
}

export const getPickupCodeByOrder = (orderId: number) => {
  return request<PickupCodeInfo>({
    url: `/pickup/order/${orderId}`,
    method: 'GET',
    needLogin: false
  })
}

export const planRoute = (originLng: number, originLat: number, destLng: number, destLat: number) => {
  return request<RoutePlanResult>({
    url: '/amap/route',
    method: 'POST',
    data: { origin_lng: originLng, origin_lat: originLat, dest_lng: destLng, dest_lat: destLat }
  })
}

export const geocode = (address: string, city?: string) => {
  return request<GeocodeResult>({
    url: '/amap/geocode',
    method: 'POST',
    data: { address, city }
  })
}

export const simulateRiderLocation = (deliveryId: number) => {
  return request<{ message: string }>({
    url: `/delivery/${deliveryId}/simulate-location`,
    method: 'POST'
  })
}
