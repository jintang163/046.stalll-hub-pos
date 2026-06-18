const Store = require('electron-store')
const { app } = require('electron')

const store = new Store({
  name: 'pos-config',
  defaults: {
    storeID: 1,
    apiBaseURL: 'http://localhost:8080/api/v1',
    nsqLookupd: 'http://localhost:4161',
    nsqd: 'localhost:4150'
  }
})

class IPC {
  static init(ipcMain, db, mainWindow) {
    ipcMain.handle('db:getProducts', () => db.getProducts())
    ipcMain.handle('db:getProductById', (_, id) => db.getProductById(id))
    ipcMain.handle('db:getCategories', () => db.getCategories())
    ipcMain.handle('db:getSKUs', (_, productId) => db.getSKUs(productId))
    ipcMain.handle('db:getAttributes', (_, productId) => db.getAttributes(productId))
    ipcMain.handle('db:saveProducts', (_, products) => db.saveProducts(products))
    ipcMain.handle('db:saveCategories', (_, categories) => db.saveCategories(categories))
    ipcMain.handle('db:updateStock', (_, skuId, stock) => db.updateStock(skuId, stock))
    ipcMain.handle('db:updateProductStatus', (_, productId, status) => db.updateProductStatus(productId, status))
    ipcMain.handle('db:deleteProduct', (_, productId) => db.deleteProduct(productId))
    ipcMain.handle('db:clearAllProducts', () => db.clearAllProducts())
    ipcMain.handle('db:getLastSyncTime', () => db.getLastSyncTime())
    ipcMain.handle('db:setLastSyncTime', (_, time) => db.setLastSyncTime(time))
    ipcMain.handle('db:getLastSyncID', () => db.getLastSyncID())
    ipcMain.handle('db:setLastSyncID', (_, id) => db.setLastSyncID(id))

    ipcMain.handle('orders:saveOrder', (_, order) => db.saveOrder(order))
    ipcMain.handle('orders:getPendingOrders', () => db.getPendingOrders())
    ipcMain.handle('orders:updateOrderStatus', (_, orderNo, status) => db.updateOrderStatus(orderNo, status))
    ipcMain.handle('orders:getOrdersByDate', (_, date) => db.getOrdersByDate(date))
    ipcMain.handle('orders:getOrderByNo', (_, orderNo) => db.getOrderByNo(orderNo))
    ipcMain.handle('orders:deleteOrder', (_, orderNo) => db.deleteOrder(orderNo))

    let syncProgress = { percent: 0, status: 'idle', message: '' }
    ipcMain.handle('sync:getProgress', () => syncProgress)
    ipcMain.handle('sync:setProgress', (_, progress) => {
      syncProgress = { ...syncProgress, ...progress }
      if (mainWindow) {
        mainWindow.webContents.send('sync:progress', syncProgress)
      }
      return syncProgress
    })

    ipcMain.handle('app:getVersion', () => app.getVersion())
    ipcMain.handle('app:getStoreID', () => store.get('storeID'))
    ipcMain.handle('app:setStoreID', (_, id) => {
      store.set('storeID', id)
      return true
    })
    ipcMain.handle('app:getConfig', () => store.store)
    ipcMain.handle('app:setConfig', (_, config) => {
      store.set(config)
      return true
    })
  }
}

module.exports = IPC
