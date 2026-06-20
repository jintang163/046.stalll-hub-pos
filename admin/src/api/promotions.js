import request from '@/utils/request'

export function getPromotionList(params) {
  return request({
    url: '/promotions',
    method: 'get',
    params
  })
}

export function getPromotion(id) {
  return request({
    url: `/promotions/${id}`,
    method: 'get'
  })
}

export function createPromotion(data) {
  return request({
    url: '/promotions',
    method: 'post',
    data
  })
}

export function updatePromotion(id, data) {
  return request({
    url: `/promotions/${id}`,
    method: 'put',
    data
  })
}

export function deletePromotion(id) {
  return request({
    url: `/promotions/${id}`,
    method: 'delete'
  })
}

export function calculateBestCombination(data) {
  return request({
    url: '/promotions/calculate',
    method: 'post',
    data
  })
}
