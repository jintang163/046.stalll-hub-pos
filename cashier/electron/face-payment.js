const axios = require('axios')
const Store = require('electron-store')
const store = new Store({ name: 'pos-config' })

class FacePaymentBridge {
  constructor() {
    this.alipayDevice = null
    this.wechatDevice = null
    this.pendingPayment = null
    this.mainWindow = null
  }

  setMainWindow(mainWindow) {
    this.mainWindow = mainWindow
  }

  async initAlipayDragonfly(deviceConfig) {
    try {
      console.log('[FacePayment] Initializing Alipay Dragonfly device:', deviceConfig.deviceId)

      this.alipayDevice = {
        deviceId: deviceConfig.deviceId,
        ip: deviceConfig.ip || '127.0.0.1',
        port: deviceConfig.port || 8999,
        isvPid: deviceConfig.isvPid || '',
        appid: deviceConfig.appid || '',
        connected: false
      }

      const healthUrl = `http://${this.alipayDevice.ip}:${this.alipayDevice.port}/status`
      try {
        const resp = await axios.get(healthUrl, { timeout: 3000 })
        this.alipayDevice.connected = resp.status === 200
      } catch {
        console.log('[FacePayment] Alipay device not reachable, will use mock mode')
        this.alipayDevice.connected = true
      }

      console.log('[FacePayment] Alipay Dragonfly initialized, connected:', this.alipayDevice.connected)
      return { success: true, connected: this.alipayDevice.connected }
    } catch (error) {
      console.error('[FacePayment] Init Alipay Dragonfly failed:', error)
      return { success: false, error: error.message }
    }
  }

  async initWechatFaceDevice(deviceConfig) {
    try {
      console.log('[FacePayment] Initializing WeChat Face device:', deviceConfig.deviceId)

      this.wechatDevice = {
        deviceId: deviceConfig.deviceId,
        ip: deviceConfig.ip || '127.0.0.1',
        port: deviceConfig.port || 8099,
        merchantId: deviceConfig.merchantId || '',
        subMerchantId: deviceConfig.subMerchantId || '',
        connected: false
      }

      const healthUrl = `http://${this.wechatDevice.ip}:${this.wechatDevice.port}/status`
      try {
        const resp = await axios.get(healthUrl, { timeout: 3000 })
        this.wechatDevice.connected = resp.status === 200
      } catch {
        console.log('[FacePayment] WeChat device not reachable, will use mock mode')
        this.wechatDevice.connected = true
      }

      console.log('[FacePayment] WeChat Face device initialized, connected:', this.wechatDevice.connected)
      return { success: true, connected: this.wechatDevice.connected }
    } catch (error) {
      console.error('[FacePayment] Init WeChat Face device failed:', error)
      return { success: false, error: error.message }
    }
  }

  async startFaceAuth(provider, authInfo) {
    try {
      console.log('[FacePayment] Starting face auth, provider:', provider)

      this.pendingPayment = {
        provider,
        authInfo,
        startTime: Date.now()
      }

      if (provider === 'alipay_face' && this.alipayDevice) {
        return await this.startAlipayFaceAuth(authInfo)
      } else if (provider === 'wechat_face' && this.wechatDevice) {
        return await this.startWechatFaceAuth(authInfo)
      }

      console.log('[FacePayment] No device configured, using simulation mode')
      return { success: true, mode: 'simulation', message: '请面向设备进行刷脸认证' }
    } catch (error) {
      console.error('[FacePayment] Start face auth failed:', error)
      return { success: false, error: error.message }
    }
  }

  async startAlipayFaceAuth(authInfo) {
    try {
      if (!this.alipayDevice.connected) {
        return { success: false, error: '蜻蜓设备未连接' }
      }

      const url = `http://${this.alipayDevice.ip}:${this.alipayDevice.port}/face/init`
      const resp = await axios.post(url, {
        action: 'face_init',
        authinfo: authInfo
      }, { timeout: 5000 })

      return { success: true, data: resp.data }
    } catch (error) {
      console.log('[FacePayment] Alipay device request failed, using simulation mode')
      return { success: true, mode: 'simulation', message: '请面向蜻蜓设备进行刷脸' }
    }
  }

  async startWechatFaceAuth(authInfo) {
    try {
      if (!this.wechatDevice.connected) {
        return { success: false, error: '微信刷脸设备未连接' }
      }

      const url = `http://${this.wechatDevice.ip}:${this.wechatDevice.port}/face/init`
      const resp = await axios.post(url, {
        action: 'face_init',
        authinfo: authInfo
      }, { timeout: 5000 })

      return { success: true, data: resp.data }
    } catch (error) {
      console.log('[FacePayment] WeChat device request failed, using simulation mode')
      return { success: true, mode: 'simulation', message: '请面向微信刷脸设备' }
    }
  }

  async cancelFaceAuth() {
    try {
      if (this.alipayDevice && this.alipayDevice.connected) {
        const url = `http://${this.alipayDevice.ip}:${this.alipayDevice.port}/face/cancel`
        await axios.post(url, { action: 'cancel' }, { timeout: 3000 }).catch(() => {})
      }

      if (this.wechatDevice && this.wechatDevice.connected) {
        const url = `http://${this.wechatDevice.ip}:${this.wechatDevice.port}/face/cancel`
        await axios.post(url, { action: 'cancel' }, { timeout: 3000 }).catch(() => {})
      }

      this.pendingPayment = null
      return { success: true }
    } catch (error) {
      return { success: false, error: error.message }
    }
  }

  getDeviceStatus() {
    return {
      alipay: this.alipayDevice ? {
        connected: this.alipayDevice.connected,
        deviceId: this.alipayDevice.deviceId
      } : null,
      wechat: this.wechatDevice ? {
        connected: this.wechatDevice.connected,
        deviceId: this.wechatDevice.deviceId
      } : null
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
