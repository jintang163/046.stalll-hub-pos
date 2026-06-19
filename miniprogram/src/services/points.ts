import request from './request'

export interface PointsRecord {
  id: number
  member_id: number
  member_name: string
  type: string
  points: number
  balance: number
  order_id: number
  order_no: string
  remark: string
  created_at: string
}

export interface RechargeActivity {
  id: number
  name: string
  min_amount: number
  bonus_amount: number
  bonus_points: number
  start_time: string
  end_time: string
  status: number
  description: string
}

export const getPointsRecords = (params?: { page?: number; page_size?: number; type?: string }) => {
  return request<{ list: PointsRecord[]; total: number }>({
    url: '/points-records',
    method: 'GET',
    data: params
  })
}

export const getAvailableRechargeActivities = () => {
  return request<RechargeActivity[]>({
    url: '/recharge-activities',
    method: 'GET',
    data: { status: 1, page: 1, page_size: 100 }
  })
}

export const processRecharge = (data: { member_id: number; amount: number; activity_id?: number }) => {
  return request({
    url: '/member-recharges',
    method: 'POST',
    data
  })
}

export const getRechargeRecords = (params?: { page?: number; page_size?: number }) => {
  return request({
    url: '/member-recharges',
    method: 'GET',
    data: params
  })
}
