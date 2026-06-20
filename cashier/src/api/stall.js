import request from '@/utils/request'

export const getStallList = (params) => {
  return request.get('/stalls', { params })
}

export const getAllStalls = (storeId) => {
  return request.get('/stalls/all', { params: { store_id: storeId } })
}

export const getStallDetail = (id) => {
  return request.get(`/stalls/${id}`)
}

export const createStall = (data) => {
  return request.post('/stalls', data)
}

export const updateStall = (id, data) => {
  return request.put(`/stalls/${id}`, data)
}

export const deleteStall = (id) => {
  return request.delete(`/stalls/${id}`)
}

export const getStallDevices = (params) => {
  return request.get('/stall-devices', { params })
}

export const registerStallDevice = (data) => {
  return request.post('/stall-devices', data)
}

export const deleteStallDevice = (id) => {
  return request.delete(`/stall-devices/${id}`)
}

export const stallHeartbeat = (deviceId, appVersion) => {
  return request.post('/stall/heartbeat', { device_id: deviceId, app_version: appVersion })
}

export const getStallUsers = (params) => {
  return request.get('/stall-users', { params })
}

export const createStallUser = (data) => {
  return request.post('/stall-users', data)
}

export const updateStallUser = (id, data) => {
  return request.put(`/stall-users/${id}`, data)
}

export const deleteStallUser = (id) => {
  return request.delete(`/stall-users/${id}`)
}

export const stallLogin = (username, password) => {
  return request.post('/stall/auth/login', { username, password })
}

export const getStallSettlements = (params) => {
  return request.get('/stall-settlements', { params })
}

export const createStallSettlement = (data) => {
  return request.post('/stall-settlements', data)
}

export const getStallDailyReport = (params) => {
  return request.get('/stall-reports/daily', { params })
}

export const generateStallDailyReport = (params) => {
  return request.post('/stall-reports/daily/generate', null, { params })
}
