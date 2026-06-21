<template>
  <el-dialog
    v-model="visible"
    title="刷脸支付"
    width="480px"
    :close-on-click-modal="false"
    :close-on-press-escape="false"
    @close="handleClose"
  >
    <div class="face-payment-dialog">
      <div class="payment-amount-section">
        <div class="amount-label">支付金额</div>
        <div class="amount-value">¥{{ amount.toFixed(2) }}</div>
        <div class="order-info">订单号: {{ orderNo }}</div>
      </div>

      <div v-if="step === 'select'" class="step-select">
        <div class="provider-grid">
          <div 
            class="provider-card" 
            :class="{ active: selectedProvider === 'alipay_face' }"
            @click="selectedProvider = 'alipay_face'"
          >
            <div class="provider-icon alipay-icon">🔵</div>
            <div class="provider-name">支付宝蜻蜓</div>
            <div class="provider-desc">刷脸支付</div>
          </div>
          <div 
            class="provider-card" 
            :class="{ active: selectedProvider === 'wechat_face' }"
            @click="selectedProvider = 'wechat_face'"
          >
            <div class="provider-icon wechat-icon">💚</div>
            <div class="provider-name">微信刷脸</div>
            <div class="provider-desc">刷脸支付</div>
          </div>
        </div>
      </div>

      <div v-if="step === 'auth'" class="step-auth">
        <div class="face-scanner">
          <div class="scanner-frame">
            <div class="scanner-ring" :class="{ scanning: !authCompleted }">
              <div class="face-icon">👤</div>
            </div>
            <div class="scanner-pulse" v-if="!authCompleted"></div>
          </div>
          <div class="scanner-status">
            <template v-if="!authCompleted">
              <div class="status-text scanning-text">请面向设备进行刷脸认证</div>
              <div class="status-hint">等待用户刷脸中...</div>
            </template>
            <template v-else-if="authSuccess">
              <div class="status-text success-text">✓ 刷脸认证成功</div>
              <div class="status-hint">正在确认支付...</div>
            </template>
            <template v-else>
              <div class="status-text fail-text">✗ 刷脸认证失败</div>
              <div class="status-hint">{{ authError || '请重新尝试' }}</div>
            </template>
          </div>
        </div>
      </div>

      <div v-if="step === 'result'" class="step-result">
        <div class="result-icon" :class="{ success: paymentSuccess, fail: !paymentSuccess }">
          {{ paymentSuccess ? '✓' : '✗' }}
        </div>
        <div class="result-title">{{ paymentSuccess ? '支付成功' : '支付失败' }}</div>
        <div class="result-amount" v-if="paymentSuccess">¥{{ amount.toFixed(2) }}</div>
        <div class="result-detail" v-if="paymentSuccess">
          <div class="detail-row">
            <span>支付方式</span>
            <span>{{ getProviderName(selectedProvider) }}</span>
          </div>
          <div class="detail-row">
            <span>交易流水号</span>
            <span>{{ transactionId }}</span>
          </div>
          <div class="detail-row">
            <span>支付时间</span>
            <span>{{ payTime }}</span>
          </div>
        </div>
        <div class="result-error" v-if="!paymentSuccess">
          {{ paymentError }}
        </div>
      </div>
    </div>

    <template #footer>
      <div class="dialog-footer">
        <el-button v-if="step === 'select'" @click="handleClose">取消</el-button>
        <el-button 
          v-if="step === 'select'" 
          type="primary" 
          :disabled="!selectedProvider"
          @click="startFacePayment"
        >
          开始刷脸
        </el-button>
        <el-button v-if="step === 'auth'" @click="cancelFaceAuth">取消</el-button>
        <el-button 
          v-if="step === 'auth' && !authCompleted" 
          type="warning"
          @click="simulateAuth"
        >
          模拟刷脸成功
        </el-button>
        <el-button 
          v-if="step === 'result'" 
          :type="paymentSuccess ? 'primary' : 'danger'"
          @click="handleClose"
        >
          {{ paymentSuccess ? '完成' : '关闭' }}
        </el-button>
        <el-button 
          v-if="step === 'result' && !paymentSuccess" 
          type="primary"
          @click="retry"
        >
          重试
        </el-button>
      </div>
    </template>
  </el-dialog>
</template>

<script setup>
import { ref, computed, onMounted, onBeforeUnmount } from 'vue'
import { ElMessage } from 'element-plus'
import { initFacePayment, confirmFacePayment, cancelFacePayment as cancelFacePaymentApi, queryFacePaymentStatus } from '@/api/face-payment'

const props = defineProps({
  modelValue: Boolean,
  orderId: { type: Number, default: 0 },
  orderNo: { type: String, default: '' },
  amount: { type: Number, default: 0 },
  storeId: { type: Number, default: 0 }
})

const emit = defineEmits(['update:modelValue', 'success', 'fail'])

const visible = computed({
  get: () => props.modelValue,
  set: (val) => emit('update:modelValue', val)
})

const step = ref('select')
const selectedProvider = ref('')
const facePaymentId = ref('')
const authCompleted = ref(false)
const authSuccess = ref(false)
const authError = ref('')
const paymentSuccess = ref(false)
const paymentError = ref('')
const transactionId = ref('')
const payTime = ref('')
const pollingTimer = ref(null)

const getProviderName = (provider) => {
  const map = {
    alipay_face: '支付宝刷脸',
    wechat_face: '微信刷脸'
  }
  return map[provider] || provider
}

const startFacePayment = async () => {
  if (!selectedProvider.value) return
  
  step.value = 'auth'
  authCompleted.value = false
  authSuccess.value = false
  authError.value = ''

  try {
    const result = await initFacePayment({
      store_id: props.storeId,
      order_id: props.orderId,
      provider: selectedProvider.value,
      device_id: 'default'
    })
    
    facePaymentId.value = result.face_payment_id

    if (window.electronAPI?.facePayment) {
      const authResult = await window.electronAPI.facePayment.startAuth(
        selectedProvider.value,
        result.authinfo
      )
      console.log('[FacePayment] Device auth result:', authResult)
    }

    startPolling()
  } catch (e) {
    console.error('Init face payment failed:', e)
    authCompleted.value = true
    authSuccess.value = false
    authError.value = e.message || '初始化失败'
  }
}

const startPolling = () => {
  stopPolling()
  pollingTimer.value = setInterval(async () => {
    if (!facePaymentId.value) return
    try {
      const result = await queryFacePaymentStatus(facePaymentId.value)
      if (result.status === 2) {
        stopPolling()
        authCompleted.value = true
        authSuccess.value = true
        onPaymentSuccess(result)
      } else if (result.status === 3) {
        stopPolling()
        authCompleted.value = true
        authSuccess.value = false
        authError.value = result.err_msg || '支付失败'
        onPaymentFail(result.err_msg)
      }
    } catch (e) {
      console.error('Polling face payment status failed:', e)
    }
  }, 2000)
}

const stopPolling = () => {
  if (pollingTimer.value) {
    clearInterval(pollingTimer.value)
    pollingTimer.value = null
  }
}

const simulateAuth = async () => {
  if (!facePaymentId.value) return

  authCompleted.value = true
  authSuccess.value = true

  try {
    const result = await confirmFacePayment({
      face_payment_id: facePaymentId.value,
      provider: selectedProvider.value,
      auth_code: 'SIM_FACE_AUTH_' + Date.now(),
      open_id: ''
    })

    onPaymentSuccess(result)
  } catch (e) {
    authSuccess.value = false
    authError.value = e.message || '支付确认失败'
    onPaymentFail(e.message)
  }
}

const onPaymentSuccess = (result) => {
  paymentSuccess.value = true
  transactionId.value = result.transaction_id || ''
  payTime.value = result.pay_time || new Date().toLocaleString()
  step.value = 'result'

  if (window.electronAPI?.voice) {
    window.electronAPI.voice.speakPaymentSuccess(props.amount, selectedProvider.value)
  }

  emit('success', {
    order_no: props.orderNo,
    transaction_id: transactionId.value,
    pay_method: selectedProvider.value
  })
}

const onPaymentFail = (error) => {
  paymentSuccess.value = false
  paymentError.value = error || '支付失败'
  step.value = 'result'

  if (window.electronAPI?.voice) {
    window.electronAPI.voice.speakPaymentFailed()
  }

  emit('fail', { order_no: props.orderNo, error })
}

const cancelFaceAuth = async () => {
  stopPolling()
  
  if (facePaymentId.value) {
    try {
      await cancelFacePaymentApi(facePaymentId.value)
    } catch (e) {
      console.error('Cancel face payment failed:', e)
    }
  }

  if (window.electronAPI?.facePayment) {
    await window.electronAPI.facePayment.cancelAuth()
  }

  handleClose()
}

const retry = () => {
  step.value = 'select'
  facePaymentId.value = ''
  authCompleted.value = false
  authSuccess.value = false
  authError.value = ''
  paymentSuccess.value = false
  paymentError.value = ''
  transactionId.value = ''
  payTime.value = ''
}

const handleClose = () => {
  stopPolling()
  step.value = 'select'
  selectedProvider.value = ''
  facePaymentId.value = ''
  authCompleted.value = false
  authSuccess.value = false
  authError.value = ''
  paymentSuccess.value = false
  paymentError.value = ''
  visible.value = false
}

onBeforeUnmount(() => {
  stopPolling()
})
</script>

<style lang="scss" scoped>
.face-payment-dialog {
  .payment-amount-section {
    text-align: center;
    padding: 20px;
    background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
    border-radius: 12px;
    margin-bottom: 24px;
    color: #fff;

    .amount-label {
      font-size: 14px;
      opacity: 0.9;
      margin-bottom: 8px;
    }

    .amount-value {
      font-size: 48px;
      font-weight: 700;
      margin-bottom: 8px;
    }

    .order-info {
      font-size: 12px;
      opacity: 0.7;
    }
  }

  .step-select {
    .provider-grid {
      display: grid;
      grid-template-columns: 1fr 1fr;
      gap: 16px;
    }

    .provider-card {
      padding: 24px 16px;
      border: 2px solid #e4e7ed;
      border-radius: 12px;
      text-align: center;
      cursor: pointer;
      transition: all 0.3s;

      &:hover {
        border-color: #409eff;
        box-shadow: 0 4px 12px rgba(64, 158, 255, 0.15);
      }

      &.active {
        border-color: #409eff;
        background: #ecf5ff;
      }

      .provider-icon {
        font-size: 48px;
        margin-bottom: 12px;
      }

      .provider-name {
        font-size: 16px;
        font-weight: 600;
        margin-bottom: 4px;
      }

      .provider-desc {
        font-size: 12px;
        color: #909399;
      }
    }
  }

  .step-auth {
    .face-scanner {
      text-align: center;
      padding: 24px 0;
    }

    .scanner-frame {
      position: relative;
      width: 200px;
      height: 200px;
      margin: 0 auto 24px;
      display: flex;
      align-items: center;
      justify-content: center;
    }

    .scanner-ring {
      width: 160px;
      height: 160px;
      border-radius: 50%;
      border: 4px solid #dcdfe6;
      display: flex;
      align-items: center;
      justify-content: center;
      transition: all 0.3s;

      &.scanning {
        border-color: #409eff;
        animation: pulse-ring 2s infinite;
      }

      .face-icon {
        font-size: 64px;
      }
    }

    .scanner-pulse {
      position: absolute;
      top: 50%;
      left: 50%;
      transform: translate(-50%, -50%);
      width: 160px;
      height: 160px;
      border-radius: 50%;
      border: 2px solid #409eff;
      animation: pulse-expand 2s infinite;
      opacity: 0;
    }

    @keyframes pulse-ring {
      0%, 100% { box-shadow: 0 0 0 0 rgba(64, 158, 255, 0.4); }
      50% { box-shadow: 0 0 0 15px rgba(64, 158, 255, 0); }
    }

    @keyframes pulse-expand {
      0% { transform: translate(-50%, -50%) scale(1); opacity: 0.5; }
      100% { transform: translate(-50%, -50%) scale(1.5); opacity: 0; }
    }

    .scanner-status {
      .status-text {
        font-size: 18px;
        font-weight: 600;
        margin-bottom: 8px;

        &.scanning-text { color: #409eff; }
        &.success-text { color: #67c23a; }
        &.fail-text { color: #f56c6c; }
      }

      .status-hint {
        font-size: 14px;
        color: #909399;
      }
    }
  }

  .step-result {
    text-align: center;
    padding: 20px 0;

    .result-icon {
      width: 80px;
      height: 80px;
      line-height: 80px;
      border-radius: 50%;
      font-size: 40px;
      color: #fff;
      margin: 0 auto 16px;

      &.success { background: #67c23a; }
      &.fail { background: #f56c6c; }
    }

    .result-title {
      font-size: 24px;
      font-weight: 700;
      margin-bottom: 12px;
    }

    .result-amount {
      font-size: 36px;
      font-weight: 700;
      color: #f56c6c;
      margin-bottom: 20px;
    }

    .result-detail {
      background: #f5f7fa;
      border-radius: 8px;
      padding: 16px;
      text-align: left;

      .detail-row {
        display: flex;
        justify-content: space-between;
        padding: 8px 0;
        font-size: 14px;

        span:first-child { color: #909399; }
        span:last-child { color: #303133; }
      }
    }

    .result-error {
      color: #f56c6c;
      font-size: 14px;
      margin-top: 12px;
    }
  }
}

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
}
</style>
