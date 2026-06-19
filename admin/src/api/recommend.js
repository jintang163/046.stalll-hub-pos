import request from '@/utils/request'

export const getRecommendConfig = () => {
  return request({
    url: '/recommendations/config',
    method: 'GET'
  })
}

export const getRecommendConfigMeta = () => {
  return request({
    url: '/recommendations/config/meta',
    method: 'GET'
  })
}

export const updateRecommendConfig = (data) => {
  return request({
    url: '/recommendations/config',
    method: 'PUT',
    data
  })
}

export const triggerRecommendRefresh = () => {
  return request({
    url: '/recommendations/refresh',
    method: 'POST'
  })
}

export const getRecommendRefreshStatus = () => {
  return request({
    url: '/recommendations/refresh/status',
    method: 'GET'
  })
}
