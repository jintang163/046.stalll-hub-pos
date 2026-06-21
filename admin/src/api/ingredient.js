import request from '@/utils/request'

export function getIngredients(params) {
  return request({
    url: '/ingredients',
    method: 'get',
    params
  })
}

export function getIngredient(id) {
  return request({
    url: `/ingredients/${id}`,
    method: 'get'
  })
}

export function createIngredient(data) {
  return request({
    url: '/ingredients',
    method: 'post',
    data
  })
}

export function updateIngredient(id, data) {
  return request({
    url: `/ingredients/${id}`,
    method: 'put',
    data
  })
}

export function deleteIngredient(id) {
  return request({
    url: `/ingredients/${id}`,
    method: 'delete'
  })
}

export function getIngredientCategories(params) {
  return request({
    url: '/ingredients/categories',
    method: 'get',
    params
  })
}

export function getPriceHistory(params) {
  return request({
    url: `/ingredients/${params.ingredient_id}/price-history`,
    method: 'get',
    params
  })
}

export function getProductBOM(productId, params) {
  return request({
    url: `/bom/${productId}`,
    method: 'get',
    params
  })
}

export function saveProductBOM(data) {
  return request({
    url: '/bom/save',
    method: 'post',
    data
  })
}

export function getProductCostDetail(productId, params) {
  return request({
    url: `/bom/${productId}/cost-detail`,
    method: 'get',
    params
  })
}

export function getCostAlerts(params) {
  return request({
    url: '/cost-alerts',
    method: 'get',
    params
  })
}

export function handleCostAlert(data) {
  return request({
    url: '/cost-alerts/handle',
    method: 'post',
    data
  })
}

export function triggerInventorySync() {
  return request({
    url: '/inventory/sync',
    method: 'post'
  })
}
