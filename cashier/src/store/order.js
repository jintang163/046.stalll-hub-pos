import { defineStore } from 'pinia'
import { ref } from 'vue'
import { ElMessage } from 'element-plus'
import { generateOrder } from '@/utils/order'
import { submitOrdersBatch } from '@/api/order'
import { checkNetwork } from '@/utils/request'

export const useOrderStore = defineStore('order', () => {
  const pendingOrders = ref([])
  const todayOrders = ref([])
  const isOnline = ref(true)
  const syncTimer = ref(null)

  const init = async () => {
    await loadPendingOrders()
    await loadTodayOrders()
    checkOnlineStatus()
    
    if (syncTimer.value) clearInterval(syncTimer.value)
    syncTimer.value = setInterval(() => {
      syncPendingOrders()
    }, 30000)
  }

  const checkOnlineStatus = async () => {
    isOnline.value = await checkNetwork()
  }

  const loadPendingOrders = async () => {
    if (!window.electronAPI) return
    try {
      pendingOrders.value = await window.electronAPI.orders.getPendingOrders()
    } catch (e) {
      console.error('加载待同步订单失败:', e)
    }
  }

  const loadTodayOrders = async () => {
    if (!window.electronAPI) return
    try {
      const today = new Date().toISOString().split('T')[0]
      todayOrders.value = await window.electronAPI.orders.getOrdersByDate(today)
    } catch (e) {
      console.error('加载今日订单失败:', e)
    }
  }

  const createOrder = async (cart, options = {}) => {
    if (!window.electronAPI) throw new Error('Electron API不可用')

    const order = generateOrder(cart, options)
    await window.electronAPI.orders.saveOrder(order)
    
    await loadPendingOrders()
    await loadTodayOrders()

    if (isOnline.value) {
      setTimeout(() => syncPendingOrders(), 1000)
    }

    return order
  }

  const syncPendingOrders = async () => {
    if (!isOnline.value || pendingOrders.value.length === 0) return

    try {
      const ordersToSync = pendingOrders.value.slice(0, 50)
      const result = await submitOrdersBatch(ordersToSync)
      
      for (const order of result.synced || []) {
        if (window.electronAPI) {
          await window.electronAPI.orders.updateOrderStatus(order.order_no, 1)
        }
      }
      
      await loadPendingOrders()
      ElMessage.success(`成功同步 ${result.synced?.length || 0} 个订单`)
    } catch (e) {
      console.error('同步订单失败:', e)
    }
  }

  const forceSync = async () => {
    await checkOnlineStatus()
    if (!isOnline.value) {
      ElMessage.error('网络不可用，请检查连接')
      return
    }
    await syncPendingOrders()
  }

  const getOrderByNo = async (orderNo) => {
    if (!window.electronAPI) return null
    return window.electronAPI.orders.getOrderByNo(orderNo)
  }

  return {
    pendingOrders,
    todayOrders,
    isOnline,
    init,
    checkOnlineStatus,
    loadPendingOrders,
    loadTodayOrders,
    createOrder,
    syncPendingOrders,
    forceSync,
    getOrderByNo
  }
})
