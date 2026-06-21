import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { OrderItem, WSMessage, KitchenOrderItem, AppConfig } from '../types'
import { orderApi } from '../services/order'
import request from '../services/request'

export const useKitchenStore = defineStore('kitchen', () => {
  const config = ref<AppConfig>({
    storeId: parseInt(localStorage.getItem('kd_storeId') || '1'),
    userId: parseInt(localStorage.getItem('kd_userId') || '999'),
    token: localStorage.getItem('kd_token') || '',
    apiBaseUrl: localStorage.getItem('kd_apiBaseUrl') || 'http://localhost:8080/api/v1',
    wsUrl: localStorage.getItem('kd_wsUrl') || '',
    overdueMinutes: parseInt(localStorage.getItem('kd_overdueMinutes') || '15'),
    voiceAlert: localStorage.getItem('kd_voiceAlert') !== 'false',
    flashAlert: localStorage.getItem('kd_flashAlert') !== 'false'
  })

  const isConnected = ref(false)
  const pendingItems = ref<OrderItem[]>([])
  const cookingItems = ref<OrderItem[]>([])
  const completedItems = ref<OrderItem[]>([])
  const lastMessage = ref<WSMessage | null>(null)
  const isLoading = ref(false)
  const overdueItems = ref<Set<number>>(new Set())

  let ws: WebSocket | null = null
  let reconnectTimer: number | null = null
  let reconnectCount = 0
  let tickTimer: number | null = null

  const pendingItemsWithMeta = computed<KitchenOrderItem[]>(() => {
    const now = Date.now()
    return pendingItems.value
      .map(item => {
        const createdAt = new Date(item.created_at).getTime()
        const waitingSeconds = Math.floor((now - createdAt) / 1000)
        const isOverdue = waitingSeconds > config.value.overdueMinutes * 60
        if (isOverdue) {
          overdueItems.value.add(item.id)
        }
        return {
          ...item,
          waitingSeconds,
          isOverdue
        }
      })
      .sort((a, b) => new Date(a.created_at).getTime() - new Date(b.created_at).getTime())
  })

  const cookingItemsWithMeta = computed<KitchenOrderItem[]>(() => {
    const now = Date.now()
    return cookingItems.value
      .map(item => {
        const createdAt = new Date(item.created_at).getTime()
        const waitingSeconds = Math.floor((now - createdAt) / 1000)
        const isOverdue = waitingSeconds > config.value.overdueMinutes * 60
        return {
          ...item,
          waitingSeconds,
          isOverdue
        }
      })
      .sort((a, b) => new Date(a.created_at).getTime() - new Date(b.created_at).getTime())
  })

  const overdueCount = computed(() => {
    const pendingOverdue = pendingItemsWithMeta.value.filter(i => i.isOverdue).length
    const cookingOverdue = cookingItemsWithMeta.value.filter(i => i.isOverdue).length
    return pendingOverdue + cookingOverdue
  })

  const totalPending = computed(() => pendingItems.value.length)
  const totalCooking = computed(() => cookingItems.value.length)
  const totalCompleted = computed(() => completedItems.value.length)

  const saveConfig = () => {
    localStorage.setItem('kd_storeId', String(config.value.storeId))
    localStorage.setItem('kd_userId', String(config.value.userId))
    localStorage.setItem('kd_token', config.value.token)
    localStorage.setItem('kd_apiBaseUrl', config.value.apiBaseUrl)
    localStorage.setItem('kd_wsUrl', config.value.wsUrl)
    localStorage.setItem('kd_overdueMinutes', String(config.value.overdueMinutes))
    localStorage.setItem('kd_voiceAlert', String(config.value.voiceAlert))
    localStorage.setItem('kd_flashAlert', String(config.value.flashAlert))
    request.setToken(config.value.token)
  }

  const getWsUrl = (): string => {
    if (config.value.wsUrl) return config.value.wsUrl
    const httpUrl = config.value.apiBaseUrl.replace('/api/v1', '')
    return httpUrl.replace('http://', 'ws://').replace('https://', 'wss://')
      + '/waiter/ws?store_id=' + config.value.storeId
      + '&user_id=' + config.value.userId
  }

  const connect = () => {
    if (!config.value.storeId || !config.value.userId) {
      console.error('[WS] No store_id or user_id')
      return
    }

    const wsUrl = getWsUrl()
    console.log('[WS] Connecting to:', wsUrl)

    try {
      ws = new WebSocket(wsUrl)

      ws.onopen = () => {
        console.log('[WS] Connected')
        isConnected.value = true
        reconnectCount = 0
        if (reconnectTimer) {
          clearTimeout(reconnectTimer)
          reconnectTimer = null
        }
        loadInitialData()
        startTick()
      }

      ws.onmessage = (event) => {
        try {
          const msg: WSMessage = JSON.parse(event.data)
          console.log('[WS] Received:', msg)
          lastMessage.value = msg
          handleMessage(msg)
        } catch (e) {
          console.error('[WS] Parse message error:', e)
        }
      }

      ws.onclose = () => {
        console.log('[WS] Disconnected')
        isConnected.value = false
        stopTick()
        scheduleReconnect()
      }

      ws.onerror = (err) => {
        console.error('[WS] Error:', err)
        isConnected.value = false
      }
    } catch (e) {
      console.error('[WS] Connection error:', e)
      scheduleReconnect()
    }
  }

  const disconnect = () => {
    if (reconnectTimer) {
      clearTimeout(reconnectTimer)
      reconnectTimer = null
    }
    stopTick()
    if (ws) {
      ws.close()
      ws = null
    }
    isConnected.value = false
  }

  const scheduleReconnect = () => {
    if (reconnectTimer) return
    reconnectCount++
    const delay = Math.min(1000 * reconnectCount, 10000)
    console.log(`[WS] Reconnecting in ${delay}ms...`)
    reconnectTimer = window.setTimeout(() => {
      reconnectTimer = null
      connect()
    }, delay)
  }

  const handleMessage = (msg: WSMessage) => {
    switch (msg.type) {
      case 'connected':
        console.log('[WS] Server confirmed connection')
        break
      case 'ping':
        break
      case 'new_order':
        handleNewOrder(msg)
        break
      case 'order_update':
        handleOrderUpdate(msg)
        break
    }
  }

  const handleNewOrder = (msg: WSMessage) => {
    if (msg.items && msg.items.length > 0) {
      msg.items.forEach(item => {
        if (item.cook_status === 0 || item.cook_status === 1) {
          addOrUpdateItem(item)
        }
      })
    } else {
      loadInitialData()
    }
    playAlert('new')
  }

  const handleOrderUpdate = (msg: WSMessage) => {
    if (msg.items && msg.items.length > 0) {
      msg.items.forEach(item => {
        addOrUpdateItem(item)
      })
    } else {
      loadInitialData()
    }
  }

  const addOrUpdateItem = (item: OrderItem) => {
    const existingIndex = pendingItems.value.findIndex(i => i.id === item.id)
    const cookingIndex = cookingItems.value.findIndex(i => i.id === item.id)

    if (item.cook_status === 0) {
      if (existingIndex >= 0) {
        pendingItems.value[existingIndex] = item
      } else if (cookingIndex < 0) {
        pendingItems.value.push(item)
      }
      cookingItems.value = cookingItems.value.filter(i => i.id !== item.id)
      completedItems.value = completedItems.value.filter(i => i.id !== item.id)
    } else if (item.cook_status === 1) {
      if (cookingIndex >= 0) {
        cookingItems.value[cookingIndex] = item
      } else {
        cookingItems.value.push(item)
      }
      pendingItems.value = pendingItems.value.filter(i => i.id !== item.id)
      completedItems.value = completedItems.value.filter(i => i.id !== item.id)
    } else if (item.cook_status >= 2) {
      if (cookingIndex >= 0) {
        completedItems.value.unshift(item)
      }
      pendingItems.value = pendingItems.value.filter(i => i.id !== item.id)
      cookingItems.value = cookingItems.value.filter(i => i.id !== item.id)
      while (completedItems.value.length > 50) {
        completedItems.value.pop()
      }
    }
  }

  const loadInitialData = async () => {
    isLoading.value = true
    try {
      const [pending, cooking] = await Promise.all([
        orderApi.getOrderItemsByCookStatus(config.value.storeId, 0),
        orderApi.getOrderItemsByCookStatus(config.value.storeId, 1)
      ])

      pendingItems.value = pending
      cookingItems.value = cooking
      console.log('[Kitchen] Loaded:', pending.length, 'pending,', cooking.length, 'cooking')
    } catch (e) {
      console.error('[Kitchen] Load initial data error:', e)
    } finally {
      isLoading.value = false
    }
  }

  const startCooking = async (itemId: number) => {
    try {
      await orderApi.updateCookStatus({
        order_item_ids: [itemId],
        cook_status: 1
      })
      const item = pendingItems.value.find(i => i.id === itemId)
      if (item) {
        addOrUpdateItem({ ...item, cook_status: 1 })
      }
    } catch (e) {
      console.error('[Kitchen] Start cooking error:', e)
      throw e
    }
  }

  const markCompleted = async (itemId: number) => {
    try {
      await orderApi.updateCookStatus({
        order_item_ids: [itemId],
        cook_status: 2
      })
      const item = cookingItems.value.find(i => i.id === itemId)
        || pendingItems.value.find(i => i.id === itemId)
      if (item) {
        addOrUpdateItem({ ...item, cook_status: 2 })
      }
    } catch (e) {
      console.error('[Kitchen] Mark completed error:', e)
      throw e
    }
  }

  const markServed = async (itemId: number) => {
    try {
      await orderApi.markItemsServed([itemId])
      const item = cookingItems.value.find(i => i.id === itemId)
        || pendingItems.value.find(i => i.id === itemId)
      if (item) {
        addOrUpdateItem({ ...item, cook_status: 3 })
      }
    } catch (e) {
      console.error('[Kitchen] Mark served error:', e)
      throw e
    }
  }

  const startTick = () => {
    if (tickTimer) return
    tickTimer = window.setInterval(() => {
      if (overdueCount.value > 0 && config.value.voiceAlert) {
        playAlert('overdue')
      }
    }, 15000)
  }

  const stopTick = () => {
    if (tickTimer) {
      clearInterval(tickTimer)
      tickTimer = null
    }
  }

  const playAlert = (type: 'new' | 'overdue') => {
    if (!config.value.voiceAlert) return

    try {
      let text = ''
      if (type === 'new') {
        text = '新订单请注意'
      } else if (type === 'overdue') {
        text = `有${overdueCount.value}个菜品已超时，请尽快处理`
      }

      if ('speechSynthesis' in window && text) {
        const utterance = new SpeechSynthesisUtterance(text)
        utterance.lang = 'zh-CN'
        utterance.rate = 1
        utterance.volume = 1
        window.speechSynthesis.speak(utterance)
      }
    } catch (e) {
      console.error('[Alert] Play alert error:', e)
    }
  }

  const formatWaitingTime = (seconds: number): string => {
    const mins = Math.floor(seconds / 60)
    const secs = seconds % 60
    return `${mins.toString().padStart(2, '0')}:${secs.toString().padStart(2, '0')}`
  }

  request.setToken(config.value.token)

  return {
    config,
    isConnected,
    isLoading,
    pendingItems,
    cookingItems,
    completedItems,
    pendingItemsWithMeta,
    cookingItemsWithMeta,
    overdueCount,
    totalPending,
    totalCooking,
    totalCompleted,
    saveConfig,
    connect,
    disconnect,
    loadInitialData,
    startCooking,
    markCompleted,
    markServed,
    formatWaitingTime,
    playAlert
  }
})
