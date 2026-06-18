import request from '@/utils/request'

export function getProductList(params) {
  return request({
    url: '/products',
    method: 'get',
    params
  })
}

export function getProduct(id) {
  return request({
    url: `/products/${id}`,
    method: 'get'
  })
}

export function createProduct(data) {
  return request({
    url: '/products',
    method: 'post',
    data
  })
}

export function updateProduct(id, data) {
  return request({
    url: `/products/${id}`,
    method: 'put',
    data
  })
}

export function deleteProduct(id) {
  return request({
    url: `/products/${id}`,
    method: 'delete'
  })
}

export function copyProduct(data) {
  return request({
    url: '/products/copy',
    method: 'post',
    data
  })
}

export function batchUpdatePrice(data) {
  return request({
    url: '/products/batch-price',
    method: 'post',
    data
  })
}

export function updateStock(data) {
  return request({
    url: '/products/stock',
    method: 'put',
    data
  })
}

export function syncProducts(params) {
  return request({
    url: '/sync/products',
    method: 'get',
    params
  })
}

export function getStockWarnings(params) {
  return request({
    url: '/products/stock-warnings',
    method: 'get',
    params
  })
}
