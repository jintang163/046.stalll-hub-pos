import request from './request'

export interface QueueNumberInfo {
  queueNumber: string
  sequence: number
  aheadCount: number
  peopleCount: number
  status: number
  callCount: number
  createdAt: string
  tableNo?: string
}

export interface QueueInfo {
  storeId: number
  queueType: string
  queuePrefix: string
  waitCount: number
  currentNum: string
  latestNumbers: string[]
  myNumber?: QueueNumberInfo
}

export interface PreOrderItem {
  productId: number
  productName: string
  skuId: number
  skuName: string
  price: number
  quantity: number
  subtotal: number
}

export interface PreOrder {
  queueId: string
  queueNumber: string
  storeId: number
  memberId: number
  items: PreOrderItem[]
  totalAmount: number
  remark: string
  updatedAt: string
}

export interface AllWaitingResult {
  small: QueueNumberInfo[]
  medium: QueueNumberInfo[]
  large: QueueNumberInfo[]
}

export interface QueueConfig {
  storeId: number
  smallPrefix: string
  smallCapacity: number
  mediumPrefix: string
  mediumCapacity: number
  largePrefix: string
  largeCapacity: number
  autoCall: boolean
  callInterval: number
  maxCallCount: number
  autoExpire: boolean
  expireMinutes: number
  voiceNotify: boolean
  smsNotify: boolean
}

export interface QueueMessage {
  type: 'call' | 'arrive' | 'cancel' | 'connected'
  storeId: number
  queueType?: string
  queueId?: string
  queueNumber?: string
  callCount?: number
  tableNo?: string
  timestamp: number
}

type WsMessageCallback = (msg: QueueMessage) => void

class QueueService {
  private wsTask: any = null
  private reconnectTimer: any = null
  private listeners: Set<WsMessageCallback> = new Set()
  private storeId: number = 1
  private reconnectAttempts: number = 0
  private maxReconnectAttempts: number = 10

  async takeNumber(params: {
    storeId: number
    memberId?: number
    memberName: string
    memberPhone: string
    peopleCount: number
    remark?: string
  }): Promise<QueueNumberInfo> {
    return request({
      url: '/queue2/take',
      method: 'POST',
      data: params,
      needLogin: false,
    })
  }

  async getQueueInfo(storeId: number, queueType: string, queueId?: string): Promise<QueueInfo> {
    const params: any = { store_id: storeId, queue_type: queueType }
    if (queueId) params.queue_id = queueId
    return request({
      url: '/queue2/info',
      method: 'GET',
      data: params,
      needLogin: false,
    })
  }

  async getAllWaiting(storeId: number): Promise<AllWaitingResult> {
    return request({
      url: '/queue2/all-waiting',
      method: 'GET',
      data: { store_id: storeId },
    })
  }

  async callNumber(storeId: number, queueType: string): Promise<QueueNumberInfo> {
    return request({
      url: '/queue2/call',
      method: 'POST',
      data: { store_id: storeId, queue_type: queueType },
    })
  }

  async arrive(storeId: number, queueId: string, tableNo: string): Promise<void> {
    return request({
      url: '/queue2/arrive',
      method: 'POST',
      data: { store_id: storeId, queue_id: queueId, table_no: tableNo },
    })
  }

  async cancel(storeId: number, queueId: string): Promise<void> {
    return request({
      url: '/queue2/cancel',
      method: 'POST',
      data: { store_id: storeId, queue_id: queueId },
      needLogin: false,
    })
  }

  async getConfig(storeId: number): Promise<QueueConfig> {
    return request({
      url: '/queue2/config',
      method: 'GET',
      data: { store_id: storeId },
      needLogin: false,
    })
  }

  async savePreOrder(params: {
    queueId: string
    storeId: number
    memberId?: number
    items: PreOrderItem[]
    totalAmount: number
    remark?: string
  }): Promise<void> {
    return request({
      url: '/queue2/preorder',
      method: 'POST',
      data: params,
      needLogin: false,
    })
  }

  async getPreOrder(queueId: string): Promise<PreOrder> {
    return request({
      url: '/queue2/preorder',
      method: 'GET',
      data: { queue_id: queueId },
      needLogin: false,
    })
  }

  connectWebSocket(storeId: number): void {
    this.storeId = storeId
    this.reconnectAttempts = 0
    this.doConnect()
  }

  private doConnect(): void {
    if (this.wsTask) {
      try {
        this.wsTask.close()
      } catch (e) {}
      this.wsTask = null
    }

    const baseUrl = 'ws://localhost:8080/api/v1/queue/ws'
    const url = `${baseUrl}?store_id=${this.storeId}`

    // #ifdef MP-WEIXIN
    this.wsTask = wx.connectSocket({
      url,
      fail: () => {
        console.log('[WS] connectSocket failed')
        this.scheduleReconnect()
      }
    })
    // #endif

    // #ifdef H5
    if (typeof WebSocket !== 'undefined') {
      this.wsTask = new WebSocket(url)
    }
    // #endif

    if (!this.wsTask) return

    const handleOpen = () => {
      console.log('[WS] connected')
      this.reconnectAttempts = 0
      this.notifyListeners({
        type: 'connected',
        storeId: this.storeId,
        timestamp: Date.now(),
      })
    }

    const handleMessage = (res: any) => {
      try {
        const data = typeof res.data === 'string' ? res.data : JSON.stringify(res.data)
        const msg: QueueMessage = JSON.parse(data)
        this.notifyListeners(msg)
      } catch (e) {
        console.error('[WS] parse message error:', e)
      }
    }

    const handleClose = () => {
      console.log('[WS] closed')
      this.scheduleReconnect()
    }

    const handleError = (err: any) => {
      console.error('[WS] error:', err)
      this.scheduleReconnect()
    }

    // #ifdef MP-WEIXIN
    this.wsTask.onOpen(handleOpen)
    this.wsTask.onMessage(handleMessage)
    this.wsTask.onClose(handleClose)
    this.wsTask.onError(handleError)
    // #endif

    // #ifdef H5
    if (this.wsTask instanceof WebSocket) {
      this.wsTask.onopen = handleOpen
      this.wsTask.onmessage = handleMessage
      this.wsTask.onclose = handleClose
      this.wsTask.onerror = handleError
    }
    // #endif
  }

  private scheduleReconnect(): void {
    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer)
      this.reconnectTimer = null
    }

    if (this.reconnectAttempts >= this.maxReconnectAttempts) {
      console.log('[WS] max reconnect attempts reached')
      return
    }

    this.reconnectAttempts++
    const delay = Math.min(1000 * Math.pow(2, this.reconnectAttempts), 30000)
    console.log(`[WS] reconnect in ${delay}ms (attempt ${this.reconnectAttempts})`)

    this.reconnectTimer = setTimeout(() => {
      this.doConnect()
    }, delay)
  }

  disconnectWebSocket(): void {
    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer)
      this.reconnectTimer = null
    }
    if (this.wsTask) {
      try {
        // #ifdef MP-WEIXIN
        this.wsTask.close()
        // #endif
        // #ifdef H5
        if (this.wsTask instanceof WebSocket) {
          this.wsTask.close()
        }
        // #endif
      } catch (e) {}
      this.wsTask = null
    }
    this.reconnectAttempts = 0
  }

  addListener(callback: WsMessageCallback): void {
    this.listeners.add(callback)
  }

  removeListener(callback: WsMessageCallback): void {
    this.listeners.delete(callback)
  }

  private notifyListeners(msg: QueueMessage): void {
    this.listeners.forEach(cb => {
      try {
        cb(msg)
      } catch (e) {
        console.error('[WS] listener error:', e)
      }
    })
  }
}

export const queue2Service = new QueueService()
export default queue2Service
