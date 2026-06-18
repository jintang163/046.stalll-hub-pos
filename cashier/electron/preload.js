const { contextBridge, ipcRenderer } = require('electron')

contextBridge.exposeInMainWorld('ipcRenderer', {
  on: (channel, callback) => {
    const validChannels = [
      'sync:progress',
      'nsq:status',
      'nsq:product:updated',
      'nsq:product:deleted',
      'nsq:category:updated',
      'nsq:category:deleted',
      'nsq:stock:updated',
      'nsq:order:updated',
      'nsq:message'
    ]
    if (validChannels.includes(channel)) {
      ipcRenderer.on(channel, (event, ...args) => callback(...args))
    }
  },
  removeAllListeners: (channel) => {
    ipcRenderer.removeAllListeners(channel)
  },
  removeListener: (channel, callback) => {
    ipcRenderer.removeListener(channel, callback)
  }
})

contextBridge.exposeInMainWorld('electronAPI', {
  db: {
    insert: (table, data) => ipcRenderer.invoke('db:insert', table, data),
    update: (table, data, where, whereParams) => ipcRenderer.invoke('db:update', table, data, where, whereParams),
    delete: (table, where, whereParams) => ipcRenderer.invoke('db:delete', table, where, whereParams),
    query: (sql, params) => ipcRenderer.invoke('db:query', sql, params),
    queryOne: (sql, params) => ipcRenderer.invoke('db:queryOne', sql, params),
    transaction: (operations) => ipcRenderer.invoke('db:transaction', operations),
    batchInsert: (table, dataList) => ipcRenderer.invoke('db:batchInsert', table, dataList),
    batchUpdate: (table, dataList, whereFields) => ipcRenderer.invoke('db:batchUpdate', table, dataList, whereFields),
    batchDelete: (table, whereField, ids) => ipcRenderer.invoke('db:batchDelete', table, whereField, ids),
    getProducts: () => ipcRenderer.invoke('db:getProducts'),
    getProductById: (id) => ipcRenderer.invoke('db:getProductById', id),
    getCategories: () => ipcRenderer.invoke('db:getCategories'),
    getSKUs: (productId) => ipcRenderer.invoke('db:getSKUs', productId),
    getAttributes: (productId) => ipcRenderer.invoke('db:getAttributes', productId),
    saveProducts: (products) => ipcRenderer.invoke('db:saveProducts', products),
    saveCategories: (categories) => ipcRenderer.invoke('db:saveCategories', categories),
    updateStock: (skuId, stock) => ipcRenderer.invoke('db:updateStock', skuId, stock),
    updateProductStatus: (productId, status) => ipcRenderer.invoke('db:updateProductStatus', productId, status),
    deleteProduct: (productId) => ipcRenderer.invoke('db:deleteProduct', productId),
    deleteCategory: (categoryId) => ipcRenderer.invoke('db:deleteCategory', categoryId),
    clearAllProducts: () => ipcRenderer.invoke('db:clearAllProducts'),
    getLastSyncTime: () => ipcRenderer.invoke('db:getLastSyncTime'),
    setLastSyncTime: (time) => ipcRenderer.invoke('db:setLastSyncTime', time),
    getLastSyncID: () => ipcRenderer.invoke('db:getLastSyncID'),
    setLastSyncID: (id) => ipcRenderer.invoke('db:setLastSyncID', id),
    createSyncRecord: (record) => ipcRenderer.invoke('db:createSyncRecord', record),
    updateSyncRecord: (id, updates) => ipcRenderer.invoke('db:updateSyncRecord', id, updates),
    getSyncRecords: (limit) => ipcRenderer.invoke('db:getSyncRecords', limit),
    raw: (sql, params) => ipcRenderer.invoke('db:raw', sql, params)
  },
  
  nsq: {
    publish: (topic, data) => ipcRenderer.invoke('nsq:publish', topic, data),
    subscribe: (topic, channel) => ipcRenderer.invoke('nsq:subscribe', topic, channel),
    unsubscribe: (topic, channel) => ipcRenderer.invoke('nsq:unsubscribe', topic, channel),
    getStatus: () => ipcRenderer.invoke('nsq:getStatus'),
    reconnect: () => ipcRenderer.invoke('nsq:reconnect')
  },
  
  sync: {
    startFullSync: () => ipcRenderer.invoke('sync:startFullSync'),
    getSyncProgress: () => ipcRenderer.invoke('sync:getSyncProgress'),
    startIncrementalSync: () => ipcRenderer.invoke('sync:startIncrementalSync'),
    cancelSync: () => ipcRenderer.invoke('sync:cancelSync')
  },
  
  orders: {
    saveOrder: (order) => ipcRenderer.invoke('orders:saveOrder', order),
    getPendingOrders: () => ipcRenderer.invoke('orders:getPendingOrders'),
    updateOrderStatus: (orderNo, status) => ipcRenderer.invoke('orders:updateOrderStatus', orderNo, status),
    getOrdersByDate: (date) => ipcRenderer.invoke('orders:getOrdersByDate', date),
    getOrderByNo: (orderNo) => ipcRenderer.invoke('orders:getOrderByNo', orderNo),
    deleteOrder: (orderNo) => ipcRenderer.invoke('orders:deleteOrder', orderNo)
  },
  
  app: {
    getAppInfo: () => ipcRenderer.invoke('app:getAppInfo'),
    getVersion: () => ipcRenderer.invoke('app:getVersion'),
    getStoreID: () => ipcRenderer.invoke('app:getStoreID'),
    setStoreID: (id) => ipcRenderer.invoke('app:setStoreID', id),
    getConfig: () => ipcRenderer.invoke('app:getConfig'),
    setConfig: (config) => ipcRenderer.invoke('app:setConfig', config),
    quit: () => ipcRenderer.invoke('app:quit'),
    reload: () => ipcRenderer.invoke('app:reload'),
    openDevTools: () => ipcRenderer.invoke('app:openDevTools')
  }
})
