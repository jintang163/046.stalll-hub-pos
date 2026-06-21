import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { WSMessage, WaiterCall, TableInfo, OrderDetail } from '../types'
import { useUserStore } from './user'
import { BASE_URL } from '../services/request'

export const useWebSocketStore = defineStore('websocket', () => {
  const isConnected = ref(false)
  const pendingCalls = ref<WaiterCall[]>([])
  const lastMessage = ref<WSMessage | null>(null)
  let socketTask: UniApp.SocketTask | null = null
  let reconnectTimer: number | null = null
  let reconnectCount = 0

  const connect = () => {
    const userStore = useUserStore()
    if (!userStore.token || !userStore.userInfo?.store_id) {
      console.error('[WS] No token or store info, cannot connect')
      return
    }

    const wsUrl = BASE_URL.replace('http://', 'ws://').replace('https://', 'wss://')
      + '/waiter/ws?store_id=' + userStore.userInfo.store_id
      + '&user_id=' + userStore.userInfo.id

    console.log('[WS] Connecting to:', wsUrl)
    
    socketTask = uni.connectSocket({
      url: wsUrl,
      complete: () => {}
    })

    socketTask.onOpen(() => {
      console.log('[WS] Connected')
      isConnected.value = true
      reconnectCount = 0
      if (reconnectTimer) {
        clearTimeout(reconnectTimer)
        reconnectTimer = null
      }
    })

    socketTask.onMessage((res: any) => {
      try {
        const msg: WSMessage = JSON.parse(res.data)
        console.log('[WS] Received:', msg)
        lastMessage.value = msg
        handleMessage(msg)
      } catch (e) {
        console.error('[WS] Parse message error:', e)
      }
    })

    socketTask.onClose(() => {
      console.log('[WS] Disconnected')
      isConnected.value = false
      scheduleReconnect()
    })

    socketTask.onError((err: any) => {
      console.error('[WS] Error:', err)
      isConnected.value = false
    })
  }

  const handleMessage = (msg: WSMessage) => {
    switch (msg.type) {
      case 'connected':
        console.log('[WS] Server confirmed connection')
        break
      case 'call_waiter':
        handleCallWaiter(msg)
        break
      case 'order_update':
        handleOrderUpdate(msg)
        break
    }
  }

  const handleCallWaiter = (msg: WSMessage) => {
    const call: WaiterCall = {
      id: msg.call_id!,
      store_id: msg.store_id!,
      table_id: msg.table_id!,
      table_no: msg.table_no!,
      content: msg.content || '',
      call_type: msg.call_type as any || 'service',
      status: 1,
      handler_id: 0,
      handle_time: null,
      created_at: msg.created_at || new Date().toISOString()
    }

    pendingCalls.value.unshift(call)

    uni.showModal({
      title: '新呼叫',
      content: `${call.table_no} 桌呼叫服务\n${call.content || getCallTypeText(call.call_type)}`,
      confirmText: '去处理',
      cancelText: '知道了',
      success: (res) => {
        if (res.confirm) {
          uni.switchTab({ url: '/pages/calls/index' })
        }
      }
    })

    uni.vibrateLong({ complete: () => {} })
  }

  const handleOrderUpdate = (msg: WSMessage) => {
    uni.showToast({
      title: `订单 ${msg.order_no} 已更新`,
      icon: 'none',
      duration: 2000
    })
  }

  const getCallTypeText = (type: string) => {
    const map: Record<string, string> = {
      service: '需要服务',
      water: '需要加水',
      pay: '需要结账',
      other: '其他服务'
    }
    return map[type] || '需要服务'
  }

  const scheduleReconnect = () => {
    if (reconnectTimer) return
    reconnectCount++
    const delay = Math.min(1000 * reconnectCount, 10000)
    console.log(`[WS] Reconnecting in ${delay}ms...`)
    reconnectTimer = setTimeout(() => {
      reconnectTimer = null
      connect()
    }, delay) as unknown as number
  }

  const disconnect = () => {
    if (reconnectTimer) {
      clearTimeout(reconnectTimer)
      reconnectTimer = null
    }
    if (socketTask) {
      socketTask.close({ complete: () => {} })
      socketTask = null
    }
    isConnected.value = false
  }

  const clearCall = (callId: number) => {
    pendingCalls.value = pendingCalls.value.filter(c => c.id !== callId)
  }

  const clearAllCalls = () => {
    pendingCalls.value = []
  }

  return {
    isConnected,
    pendingCalls,
    lastMessage,
    connect,
    disconnect,
    clearCall,
    clearAllCalls
  }
})
