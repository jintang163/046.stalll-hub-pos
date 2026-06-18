import { defineStore } from 'pinia'
import { ref } from 'vue'
import { ElMessage } from 'element-plus'
import { useProductStore } from './product'

export const useNSQStore = defineStore('nsq', () => {
  const connected = ref(false)
  const config = ref({
    nsqd: 'localhost:4150',
    lookupd: 'http://localhost:4161'
  })
  const subscribers = ref([])

  const productStore = useProductStore()

  const init = async () => {
    if (!window.electronAPI) return
    try {
      const appConfig = await window.electronAPI.app.getConfig()
      config.value.nsqd = appConfig.nsqd || config.value.nsqd
      config.value.lookupd = appConfig.nsqLookupd || config.value.lookupd
      
      loadNSQLibrary()
    } catch (e) {
      console.error('NSQ初始化失败:', e)
    }
  }

  const loadNSQLibrary = () => {
    if (typeof require === 'undefined') return
    
    try {
      const nsq = require('nsqjs')
      if (!nsq) return
      
      subscribeTopics(nsq)
    } catch (e) {
      console.error('加载NSQ库失败:', e)
    }
  }

  const subscribeTopics = (nsq) => {
    const topics = [
      { topic: 'product_change', channel: 'cashier', handler: handleProductChange },
      { topic: 'stock_change', channel: 'cashier', handler: handleStockChange },
      { topic: 'order_status', channel: 'cashier', handler: handleOrderStatus }
    ]

    topics.forEach(({ topic, channel, handler }) => {
      try {
        const [host, port] = config.value.nsqd.split(':')
        const reader = new nsq.Reader(topic, channel, {
          lookupdHTTPAddresses: config.value.lookupd,
          nsqdTCPAddresses: config.value.nsqd
        })

        reader.connect()

        reader.on('ready', () => {
          connected.value = true
          console.log(`NSQ 已连接到 ${topic}`)
        })

        reader.on('message', (msg) => {
          try {
            const data = JSON.parse(msg.body.toString())
            handler(data)
            msg.finish()
          } catch (e) {
            console.error(`处理NSQ消息失败 ${topic}:`, e)
            msg.requeue()
          }
        })

        reader.on('error', (err) => {
          console.error(`NSQ ${topic} 错误:`, err)
        })

        reader.on('nsqd_closed', () => {
          connected.value = false
        })

        subscribers.value.push(reader)
      } catch (e) {
        console.error(`订阅 ${topic} 失败:`, e)
      }
    })
  }

  const handleProductChange = async (data) => {
    console.log('收到商品变更:', data)
    
    if (!window.electronAPI) return
    
    const { action, store_id, product_id, data: productData } = data
    
    switch (action) {
      case 'create':
      case 'update':
        if (productData) {
          await window.electronAPI.db.saveProducts([productData])
          productStore.updateProduct(productData)
          ElMessage.info(`商品「${productData.name}」已更新`)
        }
        break
      case 'delete':
        await window.electronAPI.db.deleteProduct(product_id)
        productStore.removeProduct(product_id)
        ElMessage.info('商品已删除')
        break
      case 'status':
        if (productData) {
          await window.electronAPI.db.updateProductStatus(product_id, productData.status)
          productStore.updateProduct({ id: product_id, status: productData.status })
        }
        break
    }
  }

  const handleStockChange = async (data) => {
    console.log('收到库存变更:', data)
    
    if (!window.electronAPI) return
    
    const { sku_id, stock } = data
    await window.electronAPI.db.updateStock(sku_id, stock)
    productStore.updateStock(sku_id, stock)
  }

  const handleOrderStatus = (data) => {
    console.log('收到订单状态变更:', data)
    const { order_no, status } = data
    ElMessage.info(`订单 ${order_no} 状态已更新`)
  }

  const destroy = () => {
    subscribers.value.forEach(reader => {
      try {
        reader.close()
      } catch (e) {
        console.error('关闭NSQ订阅失败:', e)
      }
    })
    subscribers.value = []
    connected.value = false
  }

  return {
    connected,
    config,
    init,
    destroy
  }
})
