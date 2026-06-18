const axios = require('axios')

let syncProgress = {
  percent: 0,
  status: 'idle',
  message: '',
  currentStep: '',
  totalCount: 0,
  successCount: 0,
  failCount: 0
}

let syncCancelled = false
let mainWindowRef = null

class IPC {
  static init(ipcMain, db, mainWindow, nsq, store, app) {
    mainWindowRef = mainWindow

    IPC.initDBHandlers(ipcMain, db)
    IPC.initNSQHandlers(ipcMain, nsq, store)
    IPC.initSyncHandlers(ipcMain, db, store)
    IPC.initOrdersHandlers(ipcMain, db)
    IPC.initAppHandlers(ipcMain, app, store, mainWindow)
  }

  static initDBHandlers(ipcMain, db) {
    ipcMain.handle('db:insert', (_, table, data) => {
      try {
        return db.insert(table, data)
      } catch (error) {
        console.error('[IPC] db:insert error:', error)
        return { success: false, error: error.message }
      }
    })

    ipcMain.handle('db:update', (_, table, data, where, whereParams = []) => {
      try {
        return db.update(table, data, where, whereParams)
      } catch (error) {
        console.error('[IPC] db:update error:', error)
        return { success: false, error: error.message }
      }
    })

    ipcMain.handle('db:delete', (_, table, where, whereParams = []) => {
      try {
        return db.delete(table, where, whereParams)
      } catch (error) {
        console.error('[IPC] db:delete error:', error)
        return { success: false, error: error.message }
      }
    })

    ipcMain.handle('db:query', (_, sql, params = []) => {
      try {
        return db.query(sql, params)
      } catch (error) {
        console.error('[IPC] db:query error:', error)
        return { success: false, error: error.message }
      }
    })

    ipcMain.handle('db:queryOne', (_, sql, params = []) => {
      try {
        return db.queryOne(sql, params)
      } catch (error) {
        console.error('[IPC] db:queryOne error:', error)
        return { success: false, error: error.message }
      }
    })

    ipcMain.handle('db:transaction', (_, operations) => {
      try {
        return db.transaction(() => {
          const results = []
          for (const op of operations) {
            const { type, table, data, where, whereParams, sql, params } = op
            let result
            switch (type) {
              case 'insert':
                result = db.insert(table, data)
                break
              case 'update':
                result = db.update(table, data, where, whereParams)
                break
              case 'delete':
                result = db.delete(table, where, whereParams)
                break
              case 'query':
                result = db.query(sql, params)
                break
              case 'queryOne':
                result = db.queryOne(sql, params)
                break
              default:
                throw new Error(`Unknown operation type: ${type}`)
            }
            results.push(result)
          }
          return { success: true, results }
        })
      } catch (error) {
        console.error('[IPC] db:transaction error:', error)
        return { success: false, error: error.message }
      }
    })

    ipcMain.handle('db:batchInsert', (_, table, dataList) => {
      try {
        return db.batchInsert(table, dataList)
      } catch (error) {
        console.error('[IPC] db:batchInsert error:', error)
        return { success: false, error: error.message }
      }
    })

    ipcMain.handle('db:batchUpdate', (_, table, dataList, whereFields) => {
      try {
        return db.batchUpdate(table, dataList, whereFields)
      } catch (error) {
        console.error('[IPC] db:batchUpdate error:', error)
        return { success: false, error: error.message }
      }
    })

    ipcMain.handle('db:batchDelete', (_, table, whereField, ids) => {
      try {
        return db.batchDelete(table, whereField, ids)
      } catch (error) {
        console.error('[IPC] db:batchDelete error:', error)
        return { success: false, error: error.message }
      }
    })

    ipcMain.handle('db:getProducts', () => {
      try {
        return db.getProducts()
      } catch (error) {
        console.error('[IPC] db:getProducts error:', error)
        return []
      }
    })

    ipcMain.handle('db:getProductById', (_, id) => {
      try {
        return db.getProductById(id)
      } catch (error) {
        console.error('[IPC] db:getProductById error:', error)
        return null
      }
    })

    ipcMain.handle('db:getCategories', () => {
      try {
        return db.getCategories()
      } catch (error) {
        console.error('[IPC] db:getCategories error:', error)
        return []
      }
    })

    ipcMain.handle('db:getSKUs', (_, productId) => {
      try {
        return db.getSKUs(productId)
      } catch (error) {
        console.error('[IPC] db:getSKUs error:', error)
        return []
      }
    })

    ipcMain.handle('db:getAttributes', (_, productId) => {
      try {
        return db.getAttributes(productId)
      } catch (error) {
        console.error('[IPC] db:getAttributes error:', error)
        return []
      }
    })

    ipcMain.handle('db:saveProducts', (_, products) => {
      try {
        return db.saveProducts(products)
      } catch (error) {
        console.error('[IPC] db:saveProducts error:', error)
        return { success: false, error: error.message }
      }
    })

    ipcMain.handle('db:saveCategories', (_, categories) => {
      try {
        return db.saveCategories(categories)
      } catch (error) {
        console.error('[IPC] db:saveCategories error:', error)
        return { success: false, error: error.message }
      }
    })

    ipcMain.handle('db:updateStock', (_, skuId, stock) => {
      try {
        return db.updateStock(skuId, stock)
      } catch (error) {
        console.error('[IPC] db:updateStock error:', error)
        return { success: false, error: error.message }
      }
    })

    ipcMain.handle('db:updateProductStatus', (_, productId, status) => {
      try {
        return db.updateProductStatus(productId, status)
      } catch (error) {
        console.error('[IPC] db:updateProductStatus error:', error)
        return { success: false, error: error.message }
      }
    })

    ipcMain.handle('db:deleteProduct', (_, productId) => {
      try {
        return db.deleteProduct(productId)
      } catch (error) {
        console.error('[IPC] db:deleteProduct error:', error)
        return { success: false, error: error.message }
      }
    })

    ipcMain.handle('db:deleteCategory', (_, categoryId) => {
      try {
        return db.deleteCategory(categoryId)
      } catch (error) {
        console.error('[IPC] db:deleteCategory error:', error)
        return { success: false, error: error.message }
      }
    })

    ipcMain.handle('db:clearAllProducts', () => {
      try {
        return db.clearAllProducts()
      } catch (error) {
        console.error('[IPC] db:clearAllProducts error:', error)
        return { success: false, error: error.message }
      }
    })

    ipcMain.handle('db:getLastSyncTime', () => {
      try {
        return db.getLastSyncTime()
      } catch (error) {
        console.error('[IPC] db:getLastSyncTime error:', error)
        return null
      }
    })

    ipcMain.handle('db:setLastSyncTime', (_, time) => {
      try {
        return db.setLastSyncTime(time)
      } catch (error) {
        console.error('[IPC] db:setLastSyncTime error:', error)
        return { success: false, error: error.message }
      }
    })

    ipcMain.handle('db:getLastSyncID', () => {
      try {
        return db.getLastSyncID()
      } catch (error) {
        console.error('[IPC] db:getLastSyncID error:', error)
        return 0
      }
    })

    ipcMain.handle('db:setLastSyncID', (_, id) => {
      try {
        return db.setLastSyncID(id)
      } catch (error) {
        console.error('[IPC] db:setLastSyncID error:', error)
        return { success: false, error: error.message }
      }
    })

    ipcMain.handle('db:createSyncRecord', (_, record) => {
      try {
        return db.createSyncRecord(record)
      } catch (error) {
        console.error('[IPC] db:createSyncRecord error:', error)
        return { success: false, error: error.message }
      }
    })

    ipcMain.handle('db:updateSyncRecord', (_, id, updates) => {
      try {
        return db.updateSyncRecord(id, updates)
      } catch (error) {
        console.error('[IPC] db:updateSyncRecord error:', error)
        return { success: false, error: error.message }
      }
    })

    ipcMain.handle('db:getSyncRecords', (_, limit) => {
      try {
        return db.getSyncRecords(limit)
      } catch (error) {
        console.error('[IPC] db:getSyncRecords error:', error)
        return []
      }
    })

    ipcMain.handle('db:raw', (_, sql, params) => {
      try {
        return db.raw(sql, params)
      } catch (error) {
        console.error('[IPC] db:raw error:', error)
        return { success: false, error: error.message }
      }
    })
  }

  static initNSQHandlers(ipcMain, nsq, store) {
    ipcMain.handle('nsq:publish', async (_, topic, data) => {
      if (!nsq) {
        return { success: false, error: 'NSQ client not initialized' }
      }
      try {
        return await nsq.publish(topic, data)
      } catch (error) {
        console.error('[IPC] nsq:publish error:', error)
        return { success: false, error: error.message }
      }
    })

    ipcMain.handle('nsq:subscribe', async (_, topic, channel) => {
      if (!nsq) {
        return { success: false, error: 'NSQ client not initialized' }
      }
      try {
        return await nsq.subscribe(topic, channel, (data, message) => {
          if (mainWindowRef && !mainWindowRef.isDestroyed()) {
            mainWindowRef.webContents.send('nsq:message', { topic, channel, data })
          }
        })
      } catch (error) {
        console.error('[IPC] nsq:subscribe error:', error)
        return { success: false, error: error.message }
      }
    })

    ipcMain.handle('nsq:unsubscribe', async (_, topic, channel) => {
      if (!nsq) {
        return { success: false, error: 'NSQ client not initialized' }
      }
      try {
        return await nsq.unsubscribe(topic, channel)
      } catch (error) {
        console.error('[IPC] nsq:unsubscribe error:', error)
        return { success: false, error: error.message }
      }
    })

    ipcMain.handle('nsq:getStatus', () => {
      if (!nsq) {
        return { isConnected: false, error: 'NSQ client not initialized' }
      }
      try {
        return nsq.getConnectionStatus()
      } catch (error) {
        console.error('[IPC] nsq:getStatus error:', error)
        return { isConnected: false, error: error.message }
      }
    })

    ipcMain.handle('nsq:reconnect', async () => {
      if (!nsq) {
        return { success: false, error: 'NSQ client not initialized' }
      }
      try {
        await nsq.close()
        await nsq.connect()
        return { success: true }
      } catch (error) {
        console.error('[IPC] nsq:reconnect error:', error)
        return { success: false, error: error.message }
      }
    })
  }

  static initSyncHandlers(ipcMain, db, store) {
    const updateProgress = (progress) => {
      syncProgress = { ...syncProgress, ...progress }
      if (mainWindowRef && !mainWindowRef.isDestroyed()) {
        mainWindowRef.webContents.send('sync:progress', syncProgress)
      }
    }

    ipcMain.handle('sync:getProgress', () => {
      return syncProgress
    })

    ipcMain.handle('sync:startFullSync', async () => {
      if (syncProgress.status === 'syncing') {
        return { success: false, error: 'Sync already in progress' }
      }

      syncCancelled = false
      const startTime = new Date().toISOString()
      let syncRecordId = null

      try {
        const recordResult = db.createSyncRecord({
          sync_type: 'full',
          status: 'running',
          start_time: startTime
        })
        syncRecordId = recordResult.id

        updateProgress({
          percent: 0,
          status: 'syncing',
          message: '开始全量同步...',
          currentStep: 'initializing',
          totalCount: 0,
          successCount: 0,
          failCount: 0
        })

        const apiBaseURL = store.get('apiBaseURL')
        const storeID = store.get('storeID')

        updateProgress({ percent: 5, message: '同步分类数据...', currentStep: 'categories' })
        await IPC.syncCategories(db, apiBaseURL, storeID, updateProgress)

        if (syncCancelled) {
          throw new Error('Sync cancelled by user')
        }

        updateProgress({ percent: 20, message: '同步商品数据...', currentStep: 'products' })
        await IPC.syncProducts(db, apiBaseURL, storeID, updateProgress)

        if (syncCancelled) {
          throw new Error('Sync cancelled by user')
        }

        updateProgress({ percent: 70, message: '同步订单数据...', currentStep: 'orders' })
        await IPC.syncOrders(db, apiBaseURL, storeID, updateProgress)

        if (syncCancelled) {
          throw new Error('Sync cancelled by user')
        }

        updateProgress({ percent: 95, message: '更新同步记录...', currentStep: 'finalizing' })
        db.setLastSyncTime(new Date().toISOString())

        updateProgress({ percent: 100, status: 'completed', message: '同步完成', currentStep: 'done' })

        db.updateSyncRecord(syncRecordId, {
          status: 'completed',
          total_count: syncProgress.totalCount,
          success_count: syncProgress.successCount,
          fail_count: syncProgress.failCount,
          end_time: new Date().toISOString()
        })

        return { success: true, ...syncProgress }
      } catch (error) {
        console.error('[IPC] Full sync error:', error)
        updateProgress({
          status: 'failed',
          message: `同步失败: ${error.message}`,
          currentStep: 'error'
        })

        if (syncRecordId) {
          db.updateSyncRecord(syncRecordId, {
            status: 'failed',
            total_count: syncProgress.totalCount,
            success_count: syncProgress.successCount,
            fail_count: syncProgress.failCount,
            end_time: new Date().toISOString(),
            error_message: error.message
          })
        }

        return { success: false, error: error.message }
      }
    })

    ipcMain.handle('sync:startIncrementalSync', async () => {
      if (syncProgress.status === 'syncing') {
        return { success: false, error: 'Sync already in progress' }
      }

      syncCancelled = false
      const startTime = new Date().toISOString()
      let syncRecordId = null

      try {
        const recordResult = db.createSyncRecord({
          sync_type: 'incremental',
          status: 'running',
          start_time: startTime
        })
        syncRecordId = recordResult.id

        const lastSyncID = db.getLastSyncID()
        const lastSyncTime = db.getLastSyncTime()

        updateProgress({
          percent: 0,
          status: 'syncing',
          message: '开始增量同步...',
          currentStep: 'initializing'
        })

        const apiBaseURL = store.get('apiBaseURL')
        const storeID = store.get('storeID')

        updateProgress({ percent: 10, message: '同步增量商品数据...', currentStep: 'products' })
        await IPC.syncIncrementalProducts(db, apiBaseURL, storeID, lastSyncID, lastSyncTime, updateProgress)

        if (syncCancelled) {
          throw new Error('Sync cancelled by user')
        }

        updateProgress({ percent: 50, message: '同步增量订单数据...', currentStep: 'orders' })
        await IPC.syncIncrementalOrders(db, apiBaseURL, storeID, lastSyncTime, updateProgress)

        if (syncCancelled) {
          throw new Error('Sync cancelled by user')
        }

        updateProgress({ percent: 90, message: '更新同步记录...', currentStep: 'finalizing' })
        db.setLastSyncTime(new Date().toISOString())

        updateProgress({ percent: 100, status: 'completed', message: '同步完成', currentStep: 'done' })

        db.updateSyncRecord(syncRecordId, {
          status: 'completed',
          total_count: syncProgress.totalCount,
          success_count: syncProgress.successCount,
          fail_count: syncProgress.failCount,
          end_time: new Date().toISOString()
        })

        return { success: true, ...syncProgress }
      } catch (error) {
        console.error('[IPC] Incremental sync error:', error)
        updateProgress({
          status: 'failed',
          message: `同步失败: ${error.message}`,
          currentStep: 'error'
        })

        if (syncRecordId) {
          db.updateSyncRecord(syncRecordId, {
            status: 'failed',
            total_count: syncProgress.totalCount,
            success_count: syncProgress.successCount,
            fail_count: syncProgress.failCount,
            end_time: new Date().toISOString(),
            error_message: error.message
          })
        }

        return { success: false, error: error.message }
      }
    })

    ipcMain.handle('sync:cancelSync', () => {
      if (syncProgress.status === 'syncing') {
        syncCancelled = true
        updateProgress({ status: 'cancelling', message: '正在取消同步...' })
        return { success: true }
      }
      return { success: false, message: 'No sync in progress' }
    })
  }

  static async syncCategories(db, apiBaseURL, storeID, updateProgress) {
    try {
      const response = await axios.get(`${apiBaseURL}/categories`, {
        params: { store_id: storeID, page_size: 1000 }
      })
      const categories = response.data.data || []
      
      updateProgress({ totalCount: categories.length })
      
      if (categories.length > 0) {
        db.clearAllProducts()
        const result = db.saveCategories(categories)
        updateProgress({ successCount: result.count })
      }
    } catch (error) {
      console.error('[Sync] Categories sync error:', error)
      updateProgress({ failCount: (syncProgress.failCount || 0) + 1 })
      throw error
    }
  }

  static async syncProducts(db, apiBaseURL, storeID, updateProgress) {
    try {
      const response = await axios.get(`${apiBaseURL}/products`, {
        params: { store_id: storeID, page_size: 1000, include: 'skus,attributes' }
      })
      const products = response.data.data || []
      
      updateProgress({ totalCount: syncProgress.totalCount + products.length })
      
      const batchSize = 50
      for (let i = 0; i < products.length; i += batchSize) {
        if (syncCancelled) break
        
        const batch = products.slice(i, i + batchSize)
        const result = db.saveProducts(batch)
        
        const currentSuccess = (syncProgress.successCount || 0) + result.count
        const percent = Math.min(65, 20 + Math.floor((i + batch.length) / products.length * 50))
        
        updateProgress({
          successCount: currentSuccess,
          percent,
          message: `同步商品数据... ${i + batch.length}/${products.length}`
        })
        
        await new Promise(resolve => setTimeout(resolve, 50))
      }
    } catch (error) {
      console.error('[Sync] Products sync error:', error)
      updateProgress({ failCount: (syncProgress.failCount || 0) + 1 })
      throw error
    }
  }

  static async syncOrders(db, apiBaseURL, storeID, updateProgress) {
    try {
      const response = await axios.get(`${apiBaseURL}/orders`, {
        params: { store_id: storeID, page_size: 500, start_date: new Date(Date.now() - 7 * 24 * 60 * 60 * 1000).toISOString() }
      })
      const orders = response.data.data || []
      
      updateProgress({ totalCount: syncProgress.totalCount + orders.length })
      
      for (let i = 0; i < orders.length; i++) {
        if (syncCancelled) break
        
        try {
          db.saveOrder(orders[i])
          updateProgress({
            successCount: (syncProgress.successCount || 0) + 1,
            percent: Math.min(90, 70 + Math.floor((i + 1) / orders.length * 25)),
            message: `同步订单数据... ${i + 1}/${orders.length}`
          })
        } catch (orderError) {
          console.error('[Sync] Order save error:', orderError)
          updateProgress({ failCount: (syncProgress.failCount || 0) + 1 })
        }
      }
    } catch (error) {
      console.error('[Sync] Orders sync error:', error)
      updateProgress({ failCount: (syncProgress.failCount || 0) + 1 })
      throw error
    }
  }

  static async syncIncrementalProducts(db, apiBaseURL, storeID, lastSyncID, lastSyncTime, updateProgress) {
    try {
      const params = {
        store_id: storeID,
        page_size: 1000,
        include: 'skus,attributes'
      }
      if (lastSyncTime) {
        params.updated_after = lastSyncTime
      }
      
      const response = await axios.get(`${apiBaseURL}/products/incremental`, { params })
      const products = response.data.data || []
      
      updateProgress({ totalCount: products.length })
      
      const batchSize = 50
      for (let i = 0; i < products.length; i += batchSize) {
        if (syncCancelled) break
        
        const batch = products.slice(i, i + batchSize)
        const result = db.saveProducts(batch)
        
        const currentSuccess = (syncProgress.successCount || 0) + result.count
        const percent = Math.min(45, 10 + Math.floor((i + batch.length) / products.length * 40))
        
        updateProgress({
          successCount: currentSuccess,
          percent,
          message: `同步增量商品数据... ${i + batch.length}/${products.length}`
        })
        
        await new Promise(resolve => setTimeout(resolve, 50))
      }
    } catch (error) {
      console.error('[Sync] Incremental products sync error:', error)
      updateProgress({ failCount: (syncProgress.failCount || 0) + 1 })
      throw error
    }
  }

  static async syncIncrementalOrders(db, apiBaseURL, storeID, lastSyncTime, updateProgress) {
    try {
      const params = {
        store_id: storeID,
        page_size: 500
      }
      if (lastSyncTime) {
        params.updated_after = lastSyncTime
      }
      
      const response = await axios.get(`${apiBaseURL}/orders/incremental`, { params })
      const orders = response.data.data || []
      
      updateProgress({ totalCount: syncProgress.totalCount + orders.length })
      
      for (let i = 0; i < orders.length; i++) {
        if (syncCancelled) break
        
        try {
          db.saveOrder(orders[i])
          updateProgress({
            successCount: (syncProgress.successCount || 0) + 1,
            percent: Math.min(85, 50 + Math.floor((i + 1) / orders.length * 40)),
            message: `同步增量订单数据... ${i + 1}/${orders.length}`
          })
        } catch (orderError) {
          console.error('[Sync] Incremental order save error:', orderError)
          updateProgress({ failCount: (syncProgress.failCount || 0) + 1 })
        }
      }
    } catch (error) {
      console.error('[Sync] Incremental orders sync error:', error)
      updateProgress({ failCount: (syncProgress.failCount || 0) + 1 })
      throw error
    }
  }

  static initOrdersHandlers(ipcMain, db) {
    ipcMain.handle('orders:saveOrder', (_, order) => {
      try {
        return db.saveOrder(order)
      } catch (error) {
        console.error('[IPC] orders:saveOrder error:', error)
        return { success: false, error: error.message }
      }
    })

    ipcMain.handle('orders:getPendingOrders', () => {
      try {
        return db.getPendingOrders()
      } catch (error) {
        console.error('[IPC] orders:getPendingOrders error:', error)
        return []
      }
    })

    ipcMain.handle('orders:updateOrderStatus', (_, orderNo, status) => {
      try {
        return db.updateOrderStatus(orderNo, status)
      } catch (error) {
        console.error('[IPC] orders:updateOrderStatus error:', error)
        return { success: false, error: error.message }
      }
    })

    ipcMain.handle('orders:getOrdersByDate', (_, date) => {
      try {
        return db.getOrdersByDate(date)
      } catch (error) {
        console.error('[IPC] orders:getOrdersByDate error:', error)
        return []
      }
    })

    ipcMain.handle('orders:getOrderByNo', (_, orderNo) => {
      try {
        return db.getOrderByNo(orderNo)
      } catch (error) {
        console.error('[IPC] orders:getOrderByNo error:', error)
        return null
      }
    })

    ipcMain.handle('orders:deleteOrder', (_, orderNo) => {
      try {
        return db.deleteOrder(orderNo)
      } catch (error) {
        console.error('[IPC] orders:deleteOrder error:', error)
        return { success: false, error: error.message }
      }
    })
  }

  static initAppHandlers(ipcMain, app, store, mainWindow) {
    ipcMain.handle('app:getAppInfo', () => {
      return {
        version: app.getVersion(),
        name: app.getName(),
        platform: process.platform,
        electronVersion: process.versions.electron,
        nodeVersion: process.versions.node,
        userDataPath: app.getPath('userData')
      }
    })

    ipcMain.handle('app:getVersion', () => {
      return app.getVersion()
    })

    ipcMain.handle('app:getStoreID', () => {
      return store.get('storeID')
    })

    ipcMain.handle('app:setStoreID', (_, id) => {
      store.set('storeID', id)
      return true
    })

    ipcMain.handle('app:getConfig', () => {
      return store.store
    })

    ipcMain.handle('app:setConfig', (_, config) => {
      store.set(config)
      return true
    })

    ipcMain.handle('app:quit', () => {
      app.quit()
      return true
    })

    ipcMain.handle('app:reload', () => {
      if (mainWindow) {
        mainWindow.reload()
      }
      return true
    })

    ipcMain.handle('app:openDevTools', () => {
      if (mainWindow) {
        mainWindow.webContents.openDevTools()
      }
      return true
    })
  }
}

module.exports = IPC
