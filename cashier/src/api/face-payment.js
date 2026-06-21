import request from '@/utils/request'

export const initFacePayment = (data) => {
  return request.post('/face-payment/init', data)
}

export const confirmFacePayment = (data) => {
  return request.post('/face-payment/confirm', data)
}

export const queryFacePaymentStatus = (facePaymentId) => {
  return request.get(`/face-payment/${facePaymentId}/status`)
}

export const cancelFacePayment = (facePaymentId) => {
  return request.post(`/face-payment/${facePaymentId}/cancel`)
}

export const voiceBroadcast = (data) => {
  return request.post('/voice/broadcast', data)
}
