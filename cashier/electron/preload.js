const { contextBridge, ipcRenderer } = require('electron')

contextBridge.exposeInMainWorld('ipcRenderer', {
  on: (channel, callback) => {
    const validChannels = ['sync:progress']
    if (validChannels.includes(channel)) {
      ipcRenderer.on(channel, callback)
    }
  },
  removeAllListeners: (channel) => {
    ipcRenderer.removeAllListeners(channel)
  }
})

contextBridge.exposeInMainWorld('electronAPI', {
  db: {
    getProducts: () => ipcRenderer.invoke('db:getProducts'),
    getProductById: (id) => ipcRenderer.invoke('db:getProductById', id),
    getCategories: () => ipcRenderer.invoke('db:getCategories'),
    getSKUs: (productId) => ipcRenderer.invoke('db:getSKUs', productId),
    getAttributes: (productId) => ipcRenderer.invoke('db:getAttributes', productId),
    saveProducts: (products) => ipcRenderer.invoke('db:saveProducts', products),
    saveCategories: (categories) => ipcRenderer.invoke('db:saveCategories', categories),
    updateStock: (skuId, stock) => ipcRenderer.invoke('db:updateStock', skuId),
    updateProductStatus: (productId, status) => ipcRenderer.invoke('db:updateProductStatus', productId),
    deleteProduct: (productId) => ipcRenderer.invoke('db:deleteProduct', productId),
    clearAllProducts: () => ipcRenderer.invoke('db:clearAllProducts'),
    getLastSyncTime: () => ipcRenderer.invoke('db:getLastSyncTime'),
    setLastSyncTime: (time) => ipcRenderer.invoke('db:setLastSyncTime', time),
    getLastSyncID: () => ipcRenderer.invoke('db:getLastSyncID'),
    setLastSyncID: (id) => ipcRenderer.invoke('db:setLastSyncID', id)
  },
  
  orders: {
    saveOrder: (order) => ipcRenderer.invoke('orders:saveOrder', order),
    getPendingOrders: () => ipcRenderer.invoke('orders:getPendingOrders'),
    updateOrderStatus: (orderNo, status) => ipcRenderer.invoke('orders:updateOrderStatus', orderNo),
    getOrdersByDate: (date) => ipcRenderer.invoke('orders:getOrdersByDate', date),
    getOrderByNo: (orderNo) => ipcRenderer.invoke('orders:getOrderByNo', orderNo),
    deleteOrder: (orderNo) => ipcRenderer.invoke('orders:deleteOrder', orderNo)
  },
  
  sync: {
    getProgress: () => ipcRenderer.invoke('sync:getProgress'),
    setProgress: (progress) => ipcRenderer.invoke('sync:setProgress', progress)
  },
  
  app: {
    getVersion: () => ipcRenderer.invoke('app:getVersion'),
    getStoreID: () => ipcRenderer.invoke('app:getStoreID'),
    setStoreID: (id) => ipcRenderer.invoke('app:setStoreID', id),
    getConfig: () => ipcRenderer.invoke('app:getConfig'),
    setConfig: (config) => ipcRenderer.invoke('app:setConfig', config)
  }
})
