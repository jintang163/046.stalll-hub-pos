import request from '@/utils/request'

export function getSupplierEnums() {
  return request({
    url: '/suppliers/enums',
    method: 'get'
  })
}

export function getSuppliers(params) {
  return request({
    url: '/suppliers',
    method: 'get',
    params
  })
}

export function getSupplier(id) {
  return request({
    url: `/suppliers/${id}`,
    method: 'get'
  })
}

export function createSupplier(data) {
  return request({
    url: '/suppliers',
    method: 'post',
    data
  })
}

export function updateSupplier(id, data) {
  return request({
    url: `/suppliers/${id}`,
    method: 'put',
    data
  })
}

export function deleteSupplier(id) {
  return request({
    url: `/suppliers/${id}`,
    method: 'delete'
  })
}

export function getSupplierCategories(params) {
  return request({
    url: '/suppliers/categories',
    method: 'get',
    params
  })
}

export function getSupplierStats(params) {
  return request({
    url: '/suppliers/stats',
    method: 'get',
    params
  })
}

export function notifySupplier(id, data) {
  return request({
    url: `/suppliers/${id}/notify`,
    method: 'post',
    data
  })
}

export function getPurchaseOrders(params) {
  return request({
    url: '/purchase-orders',
    method: 'get',
    params
  })
}

export function getPurchaseOrder(id) {
  return request({
    url: `/purchase-orders/${id}`,
    method: 'get'
  })
}

export function createPurchaseOrder(data) {
  return request({
    url: '/purchase-orders',
    method: 'post',
    data
  })
}

export function sendPurchaseOrder(id, data) {
  return request({
    url: `/purchase-orders/${id}/send`,
    method: 'post',
    data
  })
}

export function completePurchaseOrder(id) {
  return request({
    url: `/purchase-orders/${id}/complete`,
    method: 'post'
  })
}

export function cancelPurchaseOrder(id, data) {
  return request({
    url: `/purchase-orders/${id}/cancel`,
    method: 'post',
    data
  })
}

export function getPurchaseReceives(params) {
  return request({
    url: '/purchase-receives',
    method: 'get',
    params
  })
}

export function getPurchaseReceive(id) {
  return request({
    url: `/purchase-receives/${id}`,
    method: 'get'
  })
}

export function createPurchaseReceive(data) {
  return request({
    url: '/purchase-receives',
    method: 'post',
    data
  })
}

export function getAccountsPayable(params) {
  return request({
    url: '/accounts-payable',
    method: 'get',
    params
  })
}

export function getAccountsPayableItem(id) {
  return request({
    url: `/accounts-payable/${id}`,
    method: 'get'
  })
}

export function updatePayableOverdue(data) {
  return request({
    url: '/accounts-payable/update-overdue',
    method: 'post',
    data
  })
}

export function getPayableStats(params) {
  return request({
    url: '/accounts-payable/stats',
    method: 'get',
    params
  })
}

export function getPayablePayments(params) {
  return request({
    url: '/payable-payments',
    method: 'get',
    params
  })
}

export function createPayablePayment(data) {
  return request({
    url: '/payable-payments',
    method: 'post',
    data
  })
}

export function getReconciliations(params) {
  return request({
    url: '/reconciliations',
    method: 'get',
    params
  })
}

export function getReconciliation(id) {
  return request({
    url: `/reconciliations/${id}`,
    method: 'get'
  })
}

export function createReconciliation(data) {
  return request({
    url: '/reconciliations',
    method: 'post',
    data
  })
}

export function confirmReconciliation(id, data) {
  return request({
    url: `/reconciliations/${id}/confirm`,
    method: 'post',
    data
  })
}

export function inputSupplierReconAmount(id, data) {
  return request({
    url: `/reconciliations/${id}/supplier-amount`,
    method: 'post',
    data
  })
}
