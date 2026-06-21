const axios = require('axios')
const Store = require('electron-store')
const store = new Store({ name: 'pos-config' })

class FacePaymentBridge {
  constructor() {
    this.alipayDevice = null
    this.wechatDevice = null
    this.pendingPayment = null
    this.mainWindow = null
    this.faceAuthPollingTimer = null
    this.alipayBaseURL = null
    this.wechatBaseURL = null
  }

  setMainWindow(mainWindow) {
    this.mainWindow = mainWindow
  }

  _sendToRenderer(channel, data) {
    if (this.mainWindow && !this.mainWindow.isDestroyed()) {
      this.mainWindow.webContents.send(channel, data)
    }
  }

  async initAlipayDragonfly(deviceConfig) {
    try {
      console.log('[FacePayment] Initializing Alipay Dragonfly device:', deviceConfig)

      const ip = deviceConfig.ip || '127.0.0.1'
      const port = deviceConfig.port || 8999
      this.alipayBaseURL = `http://${ip}:${port}`

      this.alipayDevice = {
        deviceId: deviceConfig.deviceId || '',
        ip,
        port,
        appId: deviceConfig.appId || '',
        merchantId: deviceConfig.merchantId || '',
        isvPid: deviceConfig.isvPid || '',
        connected: false,
        deviceInfo: null
      }

      try {
        const resp = await axios.get(`${this.alipayBaseURL}/rpc/getDevInfo`, { timeout: 3000 })
        if (resp.data && resp.data.result_code === '200') {
          this.alipayDevice.connected = true
          this.alipayDevice.deviceInfo = resp.data.data || resp.data
        } else {
          throw new Error('getDevInfo failed')
        }
      } catch (err) {
        console.warn('[FacePayment] Alipay Dragonfly RPC unreachable, falling back to standard /status:', err.message)
        try {
          const resp2 = await axios.get(`${this.alipayBaseURL}/status`, { timeout: 3000 })
          this.alipayDevice.connected = resp2.status === 200
        } catch {
          console.warn('[FacePayment] Alipay device not reachable, will use simulation mode')
          this.alipayDevice.connected = false
        }
      }

      console.log('[FacePayment] Alipay Dragonfly initialized, connected:', this.alipayDevice.connected)
      return { success: true, connected: this.alipayDevice.connected, deviceInfo: this.alipayDevice.deviceInfo }
    } catch (error) {
      console.error('[FacePayment] Init Alipay Dragonfly failed:', error)
      return { success: false, error: error.message }
    }
  }

  async initWechatFaceDevice(deviceConfig) {
    try {
      console.log('[FacePayment] Initializing WeChat Face device:', deviceConfig)

      const ip = deviceConfig.ip || '127.0.0.1'
      const port = deviceConfig.port || 8099
      this.wechatBaseURL = `http://${ip}:${port}`

      this.wechatDevice = {
        deviceId: deviceConfig.deviceId || '',
        ip,
        port,
        appId: deviceConfig.appId || '',
        merchantId: deviceConfig.merchantId || '',
        subMerchantId: deviceConfig.subMerchantId || '',
        connected: false,
        deviceInfo: null
      }

      try {
        const resp = await axios.get(`${this.wechatBaseURL}/getdevinfo`, { timeout: 3000 })
        if (resp.data && (resp.data.return_code === 'SUCCESS' || resp.data.code === 200)) {
          this.wechatDevice.connected = true
          this.wechatDevice.deviceInfo = resp.data
        } else {
          throw new Error('getdevinfo failed')
        }
      } catch (err) {
        console.warn('[FacePayment] WeChat face device RPC unreachable, falling back /status:', err.message)
        try {
          const resp2 = await axios.get(`${this.wechatBaseURL}/status`, { timeout: 3000 })
          this.wechatDevice.connected = resp2.status === 200
        } catch {
          console.warn('[FacePayment] WeChat device not reachable, will use simulation mode')
          this.wechatDevice.connected = false
        }
      }

      console.log('[FacePayment] WeChat Face device initialized, connected:', this.wechatDevice.connected)
      return { success: true, connected: this.wechatDevice.connected, deviceInfo: this.wechatDevice.deviceInfo }
    } catch (error) {
      console.error('[FacePayment] Init WeChat Face device failed:', error)
      return { success: false, error: error.message }
    }
  }

  async startFaceAuth(provider, authInfo) {
    try {
      console.log('[FacePayment] Starting face auth, provider:', provider, 'authInfo:', authInfo)

      this.pendingPayment = {
        provider,
        authInfo,
        startTime: Date.now(),
        authCode: null,
        openId: null,
        faceCode: null
      }

      this._clearPollingTimer()

      if (provider === 'alipay_face') {
        return await this.startAlipayFaceAuth(authInfo)
      } else if (provider === 'wechat_face') {
        return await this.startWechatFaceAuth(authInfo)
      }

      return { success: true, mode: 'simulation', message: '请面向设备进行刷脸认证' }
    } catch (error) {
      console.error('[FacePayment] Start face auth failed:', error)
      return { success: false, error: error.message }
    }
  }

  async startAlipayFaceAuth(authInfo) {
    const authData = this._safeParseAuthInfo(authInfo)

    if (this.alipayDevice && this.alipayDevice.connected) {
      try {
        const params = {
          app_id: this.alipayDevice.appId || '',
          isv_pid: this.alipayDevice.isvPid || '',
          merchant_id: this.alipayDevice.merchantId || '',
          device_id: this.alipayDevice.deviceId || '',
          out_trade_no: authData.order_no || '',
          total_amount: authData.amount || '',
          scene: 'SECURITY_CODE',
          face_payment_id: authData.face_payment_id || '',
          order_detail: authData.orderDetail || JSON.stringify({ subject: '刷脸支付' })
        }

        let resp
        try {
          resp = await axios.post(`${this.alipayBaseURL}/rpc/smkStartPayFace`, params, { timeout: 5000 })
        } catch {
          resp = await axios.post(`${this.alipayBaseURL}/face/start`, params, { timeout: 5000 })
        }

        const result = resp.data || {}
        if (result.result_code === '200' || result.code === 200 || result.rtn_code === 0) {
          this._sendToRenderer('face-payment:alipay:started', {
            face_payment_id: authData.face_payment_id
          })
          this._startAlipayFaceResultPolling(authData)
          return { success: true, mode: 'device', message: '请面向蜻蜓设备进行刷脸', face_payment_id: authData.face_payment_id }
        }

        return { success: false, error: result.rtn_msg || result.message || '刷脸启动失败' }
      } catch (error) {
        console.warn('[FacePayment] Alipay Dragonfly face auth request failed:', error.message)
      }
    }

    console.log('[FacePayment] Alipay device not connected, starting simulation mode + polling')
    this._startAlipayFaceResultPolling(authData)
    return { success: true, mode: 'simulation', message: '请面向蜻蜓设备进行刷脸（模拟模式）' }
  }

  async startWechatFaceAuth(authInfo) {
    const authData = this._safeParseAuthInfo(authInfo)

    if (this.wechatDevice && this.wechatDevice.connected) {
      try {
        const params = {
          appid: this.wechatDevice.appId || '',
          mch_id: this.wechatDevice.merchantId || '',
          sub_mch_id: this.wechatDevice.subMerchantId || '',
          device_id: this.wechatDevice.deviceId || '',
          out_trade_no: authData.order_no || '',
          total_fee: Math.round(parseFloat(authData.amount || '0') * 100),
          body: authData.body || '微信刷脸支付',
          face_payment_id: authData.face_payment_id || '',
          nonce_str: Date.now().toString()
        }

        let resp
        try {
          resp = await axios.post(`${this.wechatBaseURL}/facepay`, params, { timeout: 5000 })
        } catch {
          resp = await axios.post(`${this.wechatBaseURL}/face/start`, params, { timeout: 5000 })
        }

        const result = resp.data || {}
        if (result.return_code === 'SUCCESS' || result.result_code === 'SUCCESS' || result.code === 200) {
          this._sendToRenderer('face-payment:wechat:started', {
            face_payment_id: authData.face_payment_id
          })
          this._startWechatFaceResultPolling(authData)
          return { success: true, mode: 'device', message: '请面向微信刷脸设备', face_payment_id: authData.face_payment_id }
        }

        return { success: false, error: result.return_msg || result.err_msg || result.message || '刷脸启动失败' }
      } catch (error) {
        console.warn('[FacePayment] WeChat face auth request failed:', error.message)
      }
    }

    console.log('[FacePayment] WeChat device not connected, starting simulation mode + polling')
    this._startWechatFaceResultPolling(authData)
    return { success: true, mode: 'simulation', message: '请面向微信刷脸设备（模拟模式）' }
  }

  _startAlipayFaceResultPolling(authData) {
    if (!this.alipayDevice || !this.alipayDevice.connected) return

    this._clearPollingTimer()
    const timeoutMs = 60000
    const startTs = Date.now()

    this.faceAuthPollingTimer = setInterval(async () => {
      if (Date.now() - startTs > timeoutMs) {
        this._clearPollingTimer()
        this._sendToRenderer('face-payment:timeout', {
          provider: 'alipay_face',
          face_payment_id: authData.face_payment_id
        })
        return
      }

      try {
        const params = {
          out_trade_no: authData.order_no || '',
          face_payment_id: authData.face_payment_id || ''
        }

        let resp
        try {
          resp = await axios.post(`${this.alipayBaseURL}/rpc/smkQueryPayFaceResult`, params, { timeout: 3000 })
        } catch {
          resp = await axios.post(`${this.alipayBaseURL}/face/query`, params, { timeout: 3000 })
        }

        const result = resp.data || {}

        if (result.rtn_code === 0 || result.result_code === '200') {
          const data = result.data || result
          if (data.pay_status === 'SUCCESS' || data.buyer_user_id || data.auth_code || data.security_code) {
            this._clearPollingTimer()
            const authCode = data.auth_code || data.security_code || data.face_code || ''
            const userId = data.buyer_user_id || data.user_id || data.open_id || ''

            this.pendingPayment.authCode = authCode
            this.pendingPayment.openId = userId

            this._sendToRenderer('face-payment:auth-success', {
              provider: 'alipay_face',
              face_payment_id: authData.face_payment_id,
              order_no: authData.order_no,
              auth_code: authCode,
              open_id: userId,
              buyer_logon_id: data.buyer_logon_id || '',
              raw: data
            })
          }
        }

        if (result.rtn_code && result.rtn_code !== 0 && result.rtn_code !== 1000 && result.rtn_code !== 1003) {
          this._clearPollingTimer()
          this._sendToRenderer('face-payment:auth-fail', {
            provider: 'alipay_face',
            face_payment_id: authData.face_payment_id,
            error: result.rtn_msg || '刷脸失败'
          })
        }
      } catch (err) {
        console.debug('[FacePayment] Alipay face poll failed, retrying...:', err.message)
      }
    }, 1500)
  }

  _startWechatFaceResultPolling(authData) {
    if (!this.wechatDevice || !this.wechatDevice.connected) return

    this._clearPollingTimer()
    const timeoutMs = 60000
    const startTs = Date.now()

    this.faceAuthPollingTimer = setInterval(async () => {
      if (Date.now() - startTs > timeoutMs) {
        this._clearPollingTimer()
        this._sendToRenderer('face-payment:timeout', {
          provider: 'wechat_face',
          face_payment_id: authData.face_payment_id
        })
        return
      }

      try {
        const params = {
          out_trade_no: authData.order_no || '',
          face_payment_id: authData.face_payment_id || ''
        }

        let resp
        try {
          resp = await axios.post(`${this.wechatBaseURL}/facepayquery`, params, { timeout: 3000 })
        } catch {
          resp = await axios.post(`${this.wechatBaseURL}/face/query`, params, { timeout: 3000 })
        }

        const result = resp.data || {}

        if (result.return_code === 'SUCCESS' && result.result_code === 'SUCCESS') {
          this._clearPollingTimer()
          const authCode = result.auth_code || result.face_code || result.openid || ''
          const openId = result.openid || result.sub_openid || ''

          this.pendingPayment.authCode = authCode
          this.pendingPayment.openId = openId

          this._sendToRenderer('face-payment:auth-success', {
            provider: 'wechat_face',
            face_payment_id: authData.face_payment_id,
            order_no: authData.order_no,
            auth_code: authCode,
            open_id: openId,
            face_code: result.face_code || '',
            raw: result
          })
        }

        if (result.return_code === 'SUCCESS' && result.result_code && result.result_code !== 'SUCCESS' && result.err_code !== 'USERPAYING') {
          this._clearPollingTimer()
          this._sendToRenderer('face-payment:auth-fail', {
            provider: 'wechat_face',
            face_payment_id: authData.face_payment_id,
            error: result.err_code_des || '刷脸失败'
          })
        }
      } catch (err) {
        console.debug('[FacePayment] WeChat face poll failed, retrying...:', err.message)
      }
    }, 1500)
  }

  _clearPollingTimer() {
    if (this.faceAuthPollingTimer) {
      clearInterval(this.faceAuthPollingTimer)
      this.faceAuthPollingTimer = null
    }
  }

  async cancelFaceAuth() {
    this._clearPollingTimer()

    try {
      if (this.alipayDevice && this.alipayDevice.connected) {
        try {
          await axios.post(`${this.alipayBaseURL}/rpc/smkCancel`, {
            face_payment_id: this.pendingPayment?.authInfo ? this._safeParseAuthInfo(this.pendingPayment.authInfo).face_payment_id : ''
          }, { timeout: 2000 }).catch(() => {})
        } catch {
          await axios.post(`${this.alipayBaseURL}/face/cancel`, {}, { timeout: 2000 }).catch(() => {})
        }
      }

      if (this.wechatDevice && this.wechatDevice.connected) {
        try {
          await axios.post(`${this.wechatBaseURL}/facepaycancel`, {}, { timeout: 2000 }).catch(() => {})
        } catch {
          await axios.post(`${this.wechatBaseURL}/face/cancel`, {}, { timeout: 2000 }).catch(() => {})
        }
      }
    } catch (e) {
      console.warn('[FacePayment] cancelFaceAuth graceful error:', e.message)
    }

    this.pendingPayment = null
    return { success: true }
  }

  getDeviceStatus() {
    return {
      alipay: this.alipayDevice ? {
        connected: this.alipayDevice.connected,
        deviceId: this.alipayDevice.deviceId,
        baseURL: this.alipayBaseURL
      } : null,
      wechat: this.wechatDevice ? {
        connected: this.wechatDevice.connected,
        deviceId: this.wechatDevice.deviceId,
        baseURL: this.wechatBaseURL
      } : null,
      pending: this.pendingPayment ? {
        provider: this.pendingPayment.provider,
        startTime: this.pendingPayment.startTime,
        hasAuthCode: !!this.pendingPayment.authCode
      } : null
    }
  }

  _safeParseAuthInfo(authInfo) {
    if (!authInfo) return {}
    if (typeof authInfo === 'object') return authInfo
    try {
      return JSON.parse(authInfo)
    } catch {
      return {}
    }
  }
}

class VoiceBroadcastService {
  constructor() {
    this.enabled = true
    this.volume = 1.0
    this.mainWindow = null
  }

  setMainWindow(mainWindow) {
    this.mainWindow = mainWindow
  }

  async speak(text, options = {}) {
    if (!this.enabled) return { success: false, error: 'Voice broadcast disabled' }

    try {
      if (this.mainWindow && !this.mainWindow.isDestroyed()) {
        this.mainWindow.webContents.send('voice:speak', {
          text,
          volume: options.volume || this.volume,
          rate: options.rate || 1.0,
          lang: options.lang || 'zh-CN'
        })
      }
      return { success: true }
    } catch (error) {
      console.error('[VoiceBroadcast] Speak failed:', error)
      return { success: false, error: error.message }
    }
  }

  async speakPaymentSuccess(amount, payMethod) {
    const methodText = payMethod === 'alipay_face' ? '支付宝刷脸' :
                       payMethod === 'wechat_face' ? '微信刷脸' :
                       payMethod === 'wechat' ? '微信' :
                       payMethod === 'alipay' ? '支付宝' : payMethod
    const amountStr = parseFloat(amount).toFixed(2)
    const text = `${methodText}支付成功，收款${amountStr}元`
    return this.speak(text, { rate: 0.9 })
  }

  async speakPaymentFailed() {
    return this.speak('支付失败，请重试', { rate: 0.9 })
  }

  async speakRefund(amount) {
    const amountStr = parseFloat(amount).toFixed(2)
    return this.speak(`退款${amountStr}元`, { rate: 0.9 })
  }

  setEnabled(enabled) {
    this.enabled = enabled
    return { success: true }
  }

  setVolume(volume) {
    this.volume = Math.max(0, Math.min(1, volume))
    return { success: true }
  }

  getStatus() {
    return {
      enabled: this.enabled,
      volume: this.volume
    }
  }
}

const facePaymentBridge = new FacePaymentBridge()
const voiceBroadcastService = new VoiceBroadcastService()

function initFacePaymentIPC(ipcMain, mainWindow) {
  facePaymentBridge.setMainWindow(mainWindow)
  voiceBroadcastService.setMainWindow(mainWindow)

  ipcMain.handle('face-payment:initAlipay', (_, config) => {
    return facePaymentBridge.initAlipayDragonfly(config)
  })

  ipcMain.handle('face-payment:initWechat', (_, config) => {
    return facePaymentBridge.initWechatFaceDevice(config)
  })

  ipcMain.handle('face-payment:startAuth', (_, provider, authInfo) => {
    return facePaymentBridge.startFaceAuth(provider, authInfo)
  })

  ipcMain.handle('face-payment:cancelAuth', () => {
    return facePaymentBridge.cancelFaceAuth()
  })

  ipcMain.handle('face-payment:getDeviceStatus', () => {
    return facePaymentBridge.getDeviceStatus()
  })

  ipcMain.handle('voice:speak', (_, text, options) => {
    return voiceBroadcastService.speak(text, options)
  })

  ipcMain.handle('voice:speakPaymentSuccess', (_, amount, payMethod) => {
    return voiceBroadcastService.speakPaymentSuccess(amount, payMethod)
  })

  ipcMain.handle('voice:speakPaymentFailed', () => {
    return voiceBroadcastService.speakPaymentFailed()
  })

  ipcMain.handle('voice:speakRefund', (_, amount) => {
    return voiceBroadcastService.speakRefund(amount)
  })

  ipcMain.handle('voice:setEnabled', (_, enabled) => {
    return voiceBroadcastService.setEnabled(enabled)
  })

  ipcMain.handle('voice:setVolume', (_, volume) => {
    return voiceBroadcastService.setVolume(volume)
  })

  ipcMain.handle('voice:getStatus', () => {
    return voiceBroadcastService.getStatus()
  })
}

module.exports = {
  FacePaymentBridge,
  VoiceBroadcastService,
  facePaymentBridge,
  voiceBroadcastService,
  initFacePaymentIPC
}
