const { Reader, Writer } = require('nsqjs')
const EventEmitter = require('events')

class NSQClient extends EventEmitter {
  constructor(config) {
    super()
    this.nsqdHost = config.nsqdHost || 'localhost'
    this.nsqdPort = config.nsqdPort || 4150
    this.lookupdHTTPAddresses = config.lookupdHTTPAddresses || ['http://localhost:4161']
    this.storeID = config.storeID || 1

    this.reader = null
    this.writer = null
    this.subscriptions = new Map()
    this.isConnected = false
    this.reconnectInterval = config.reconnectInterval || 5000
    this.maxReconnectAttempts = config.maxReconnectAttempts || 10
    this.reconnectAttempts = 0
    this.shouldReconnect = true
    this.mainWindow = null
    this.db = null
  }

  setMainWindow(window) {
    this.mainWindow = window
  }

  setDatabase(db) {
    this.db = db
  }

  async connect() {
    try {
      await this.connectWriter()
      await this.connectReader()
      this.isConnected = true
      this.reconnectAttempts = 0
      this.emit('connected')
      console.log('[NSQ] 连接成功')
    } catch (error) {
      console.error('[NSQ] 连接失败:', error.message)
      this.isConnected = false
      this.scheduleReconnect()
    }
  }

  async connectWriter() {
    return new Promise((resolve, reject) => {
      this.writer = new Writer(this.nsqdHost, this.nsqdPort)

      this.writer.on('ready', () => {
        console.log('[NSQ] Writer 已就绪')
        resolve()
      })

      this.writer.on('error', (err) => {
        console.error('[NSQ] Writer 错误:', err.message)
        reject(err)
      })

      this.writer.on('closed', () => {
        console.log('[NSQ] Writer 已关闭')
        if (this.shouldReconnect) {
          this.scheduleReconnect()
        }
      })

      this.writer.connect()
    })
  }

  async connectReader() {
    return new Promise((resolve, reject) => {
      const channel = `cashier_${this.storeID}`

      this.reader = new Reader('product_updates', channel, {
        lookupdHTTPAddresses: this.lookupdHTTPAddresses,
        maxInFlight: 10,
        maxAttempts: 5
      })

      this.reader.on('message', (message) => {
        this.handleMessage(message)
        message.finish()
      })

      this.reader.on('error', (err) => {
        console.error('[NSQ] Reader 错误:', err.message)
      })

      this.reader.on('nsqd_connected', () => {
        console.log('[NSQ] Reader 已连接到 nsqd')
        resolve()
      })

      this.reader.on('nsqd_closed', () => {
        console.log('[NSQ] Reader 与 nsqd 断开连接')
        if (this.shouldReconnect) {
          this.scheduleReconnect()
        }
      })

      this.reader.on('discard', (message) => {
        console.warn('[NSQ] 消息已丢弃:', message.id)
      })

      this.reader.connect()
    })
  }

  scheduleReconnect() {
    if (this.reconnectAttempts >= this.maxReconnectAttempts) {
      console.error('[NSQ] 达到最大重连次数，停止重连')
      this.emit('maxReconnectAttemptsReached')
      return
    }

    this.reconnectAttempts++
    console.log(`[NSQ] ${this.reconnectInterval / 1000}秒后进行第${this.reconnectAttempts}次重连`)

    setTimeout(() => {
      if (this.shouldReconnect) {
        this.connect()
      }
    }, this.reconnectInterval)
  }

  async handleMessage(message) {
    try {
      const data = JSON.parse(message.body.toString())
      const { type, payload } = data

      console.log(`[NSQ] 收到消息: type=${type}, id=${message.id}`)

      switch (type) {
        case 'product_created':
        case 'product_updated':
          await this.handleProductUpdate(payload)
          this.broadcastToRenderer('nsq:product:updated', payload)
          break
        case 'product_deleted':
          await this.handleProductDelete(payload)
          this.broadcastToRenderer('nsq:product:deleted', payload)
          break
        case 'category_created':
        case 'category_updated':
          await this.handleCategoryUpdate(payload)
          this.broadcastToRenderer('nsq:category:updated', payload)
          break
        case 'category_deleted':
          await this.handleCategoryDelete(payload)
          this.broadcastToRenderer('nsq:category:deleted', payload)
          break
        case 'stock_updated':
          await this.handleStockUpdate(payload)
          this.broadcastToRenderer('nsq:stock:updated', payload)
          break
        case 'order_created':
        case 'order_updated':
          await this.handleOrderUpdate(payload)
          this.broadcastToRenderer('nsq:order:updated', payload)
          break
        default:
          console.log(`[NSQ] 未知消息类型: ${type}`)
      }

      this.emit('message', { type, payload, messageId: message.id })
    } catch (error) {
      console.error('[NSQ] 处理消息失败:', error)
      this.emit('messageError', { error, messageId: message.id })
    }
  }

  async handleProductUpdate(product) {
    if (!this.db) return
    try {
      await this.db.saveProducts([product])
      console.log(`[NSQ] 商品已更新: id=${product.id}`)
    } catch (error) {
      console.error('[NSQ] 更新商品失败:', error)
    }
  }

  async handleProductDelete(payload) {
    if (!this.db) return
    try {
      const { id } = payload
      await this.db.deleteProduct(id)
      console.log(`[NSQ] 商品已删除: id=${id}`)
    } catch (error) {
      console.error('[NSQ] 删除商品失败:', error)
    }
  }

  async handleCategoryUpdate(category) {
    if (!this.db) return
    try {
      await this.db.saveCategories([category])
      console.log(`[NSQ] 分类已更新: id=${category.id}`)
    } catch (error) {
      console.error('[NSQ] 更新分类失败:', error)
    }
  }

  async handleCategoryDelete(payload) {
    if (!this.db) return
    try {
      const { id } = payload
      await this.db.deleteCategory(id)
      console.log(`[NSQ] 分类已删除: id=${id}`)
    } catch (error) {
      console.error('[NSQ] 删除分类失败:', error)
    }
  }

  async handleStockUpdate(payload) {
    if (!this.db) return
    try {
      const { sku_id, stock } = payload
      await this.db.updateStock(sku_id, stock)
      console.log(`[NSQ] 库存已更新: sku_id=${sku_id}, stock=${stock}`)
    } catch (error) {
      console.error('[NSQ] 更新库存失败:', error)
    }
  }

  async handleOrderUpdate(order) {
    if (!this.db) return
    try {
      await this.db.saveOrder(order)
      console.log(`[NSQ] 订单已更新: order_no=${order.order_no}`)
    } catch (error) {
      console.error('[NSQ] 更新订单失败:', error)
    }
  }

  broadcastToRenderer(channel, data) {
    if (this.mainWindow && !this.mainWindow.isDestroyed()) {
      this.mainWindow.webContents.send(channel, data)
    }
  }

  async publish(topic, data) {
    return new Promise((resolve, reject) => {
      if (!this.writer || !this.isConnected) {
        reject(new Error('NSQ Writer 未连接'))
        return
      }

      const message = typeof data === 'string' ? data : JSON.stringify(data)

      this.writer.publish(topic, message, (err) => {
        if (err) {
          console.error(`[NSQ] 发布消息失败: topic=${topic}`, err)
          reject(err)
        } else {
          console.log(`[NSQ] 消息已发布: topic=${topic}`)
          resolve({ success: true })
        }
      })
    })
  }

  async subscribe(topic, channel, handler) {
    return new Promise((resolve, reject) => {
      const key = `${topic}:${channel}`

      if (this.subscriptions.has(key)) {
        console.warn(`[NSQ] 已存在订阅: ${key}`)
        resolve({ success: true, alreadySubscribed: true })
        return
      }

      const reader = new Reader(topic, channel, {
        lookupdHTTPAddresses: this.lookupdHTTPAddresses,
        maxInFlight: 10
      })

      reader.on('message', (message) => {
        try {
          const data = JSON.parse(message.body.toString())
          if (handler) {
            handler(data, message)
          }
          message.finish()
        } catch (error) {
          console.error('[NSQ] 处理订阅消息失败:', error)
          message.requeue()
        }
      })

      reader.on('error', (err) => {
        console.error(`[NSQ] 订阅错误 [${key}]:`, err.message)
      })

      reader.on('nsqd_connected', () => {
        console.log(`[NSQ] 订阅成功: ${key}`)
        this.subscriptions.set(key, reader)
        resolve({ success: true, topic, channel })
      })

      reader.connect()
    })
  }

  async unsubscribe(topic, channel) {
    const key = `${topic}:${channel}`
    const reader = this.subscriptions.get(key)

    if (reader) {
      reader.close()
      this.subscriptions.delete(key)
      console.log(`[NSQ] 已取消订阅: ${key}`)
      return { success: true }
    }

    return { success: false, message: '订阅不存在' }
  }

  async close() {
    this.shouldReconnect = false

    for (const [key, reader] of this.subscriptions) {
      reader.close()
      console.log(`[NSQ] 关闭订阅: ${key}`)
    }
    this.subscriptions.clear()

    if (this.writer) {
      this.writer.close()
      this.writer = null
    }

    if (this.reader) {
      this.reader.close()
      this.reader = null
    }

    this.isConnected = false
    console.log('[NSQ] 连接已关闭')
    this.emit('closed')
  }

  getConnectionStatus() {
    return {
      isConnected: this.isConnected,
      reconnectAttempts: this.reconnectAttempts,
      subscriptions: Array.from(this.subscriptions.keys())
    }
  }
}

module.exports = NSQClient
