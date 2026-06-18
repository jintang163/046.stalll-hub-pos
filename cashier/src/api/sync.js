import request from '@/utils/request'

export const syncProducts = (storeID, lastSyncID = 0, limit = 100) => {
  return request.get('/sync/products', {
    params: { store_id: storeID, last_sync_id: lastSyncID, limit }
  })
}

export const syncCategories = (storeID) => {
  return request.get('/sync/categories', {
    params: { store_id: storeID }
  })
}

export const syncAllProducts = (storeID) => {
  return request.get('/sync/products/all', {
    params: { store_id: storeID }
  })
}

export const getSyncCount = (storeID) => {
  return request.get('/sync/count', {
    params: { store_id: storeID }
  })
}
