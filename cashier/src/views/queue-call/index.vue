<template>
  <div class="queue-call-page">
    <header class="page-header">
      <div class="header-left">
        <h1 class="page-title">排队叫号</h1>
        <el-tag :type="wsConnected ? 'success' : 'danger'" size="large" class="ws-tag">
          <el-icon><component :is="wsConnected ? 'Connection' : 'Warning'" /></el-icon>
          {{ wsConnected ? '实时连接' : '连接断开' }}
        </el-tag>
      </div>
      <div class="header-right">
        <el-select v-model="selectedStore" size="default" style="width: 160px" @change="onStoreChange">
          <el-option v-for="s in stores" :key="s.id" :label="s.name" :value="s.id" />
        </el-select>
        <el-button @click="refreshQueues" :icon="Refresh" :loading="loading">
          刷新
        </el-button>
      </div>
    </header>

    <div class="page-body">
      <section class="current-call-section">
        <div class="current-call-card">
          <div class="current-label">当前叫号</div>
          <div class="current-number" :class="currentQueueType">
            {{ currentNumber || '--' }}
          </div>
          <div class="current-info">
            <span>{{ queueTypeName[currentQueueType] }}</span>
            <span v-if="currentPeopleCount">· {{ currentPeopleCount }}人</span>
            <span v-if="callCount > 0">· 第{{ callCount }}次叫号</span>
          </div>
        </div>

        <div class="call-actions">
          <el-button type="primary" size="large" :icon="Microphone" @click="callNext('small')">
            下一位 (小桌)
          </el-button>
          <el-button type="success" size="large" :icon="Microphone" @click="callNext('medium')">
            下一位 (中桌)
          </el-button>
          <el-button type="warning" size="large" :icon="Microphone" @click="callNext('large')">
            下一位 (大桌)
          </el-button>
          <el-button type="info" size="large" :icon="RefreshRight" @click="recall">
            重叫
          </el-button>
        </div>

        <div class="quick-actions">
          <el-button :icon="Check" type="success" size="default" @click="markArrived">
            已入座
          </el-button>
          <el-button :icon="Close" type="danger" size="default" @click="skipNumber">
            过号
          </el-button>
          <el-button :icon="VideoPlay" type="primary" size="default" @click="speakCurrent">
            再次播报
          </el-button>
        </div>
      </section>

      <section class="queue-lists-section">
        <div class="queue-panel" v-for="type in queueTypes" :key="type.value">
          <div class="queue-panel-header">
            <span class="panel-title">{{ type.name }}</span>
            <el-tag :type="type.tagType" size="small">
              等待 {{ waitingCounts[type.value] || 0 }} 桌
            </el-tag>
          </div>
          <div class="queue-list" v-loading="loading">
            <div
              v-for="(item, idx) in queueLists[type.value]"
              :key="item.queue_number || item.queueNumber"
              class="queue-item"
              :class="{
                current: (currentNumber === item.queue_number || currentNumber === item.queueNumber) && type.value === currentQueueType,
                called: (item.call_count || item.callCount) > 0
              }"
            >
              <div class="queue-number">{{ item.queue_number || item.queueNumber }}</div>
              <div class="queue-meta">
                <span>{{ item.people_count || item.peopleCount }}人</span>
                <span v-if="item.call_count || item.callCount" class="call-tag">
                  叫号{{ item.call_count || item.callCount }}次
                </span>
              </div>
              <div class="queue-actions">
                <el-button size="small" type="primary" text @click="callSpecific(type.value, item)">
                  叫号
                </el-button>
                <el-button size="small" type="success" text @click="markArrivedItem(type.value, item)">
                  入座
                </el-button>
              </div>
            </div>
            <div v-if="!queueLists[type.value] || queueLists[type.value].length === 0" class="empty-queue">
              <el-empty description="暂无排队" :image-size="60" />
            </div>
          </div>
        </div>
      </section>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, onBeforeUnmount, computed } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  Refresh,
  Microphone,
  RefreshRight,
  Check,
  Close,
  VideoPlay,
  Connection,
  Warning
} from '@element-plus/icons-vue'
import request from '@/utils/request'
import { useSettingsStore } from '@/store/settings'

const router = useRouter()
const settingsStore = useSettingsStore()

const loading = ref(false)
const wsConnected = ref(false)
const selectedStore = ref(1)
const stores = ref([{ id: 1, name: '总店' }])

const queueTypes = [
  { value: 'small', name: '小桌 (1-4人)', tagType: 'primary' },
  { value: 'medium', name: '中桌 (5-6人)', tagType: 'success' },
  { value: 'large', name: '大桌 (7-10人)', tagType: 'warning' }
]
const queueTypeName = { small: '小桌', medium: '中桌', large: '大桌' }

const queueLists = reactive({ small: [], medium: [], large: [] })
const waitingCounts = reactive({ small: 0, medium: 0, large: 0 })

const currentNumber = ref('')
const currentQueueType = ref('small')
const currentPeopleCount = ref(0)
const callCount = ref(0)

let ws = null
let reconnectTimer = null
let reconnectAttempts = 0

const refreshQueues = async () => {
  loading.value = true
  try {
    const res = await request.get('/queue2/all-waiting', {
      params: { store_id: selectedStore.value }
    })
    const data = res.data?.data || res.data || {}
    queueLists.small = data.small || []
    queueLists.medium = data.medium || []
    queueLists.large = data.large || []
    waitingCounts.small = data.small?.length || 0
    waitingCounts.medium = data.medium?.length || 0
    waitingCounts.large = data.large?.length || 0

    if (!currentNumber.value) {
      for (const t of ['small', 'medium', 'large']) {
        if (queueLists[t]?.length > 0) {
          currentNumber.value = queueLists[t][0].queue_number || queueLists[t][0].queueNumber
          currentQueueType.value = t
          currentPeopleCount.value = queueLists[t][0].people_count || queueLists[t][0].peopleCount
          callCount.value = queueLists[t][0].call_count || queueLists[t][0].callCount || 0
          break
        }
      }
    }
  } catch (e) {
    console.error('Load queues error:', e)
    ElMessage.error('加载排队列表失败')
  } finally {
    loading.value = false
  }
}

const callNext = async (type) => {
  try {
    const res = await request.post('/queue2/call', {
      store_id: selectedStore.value,
      queue_type: type
    })
    const data = res.data?.data || res.data
    if (data?.queue_number || data?.queueNumber) {
      const num = data.queue_number || data.queueNumber
      currentNumber.value = num
      currentQueueType.value = type
      currentPeopleCount.value = data.people_count || data.peopleCount || 0
      callCount.value = data.call_count || data.callCount || 1

      speakCall(num, queueTypeName[type])
      refreshQueues()
      ElMessage.success(`已叫号: ${num}`)
    }
  } catch (e) {
    ElMessage.error(e.message || '叫号失败')
  }
}

const callSpecific = (type, item) => {
  const num = item.queue_number || item.queueNumber
  ElMessageBox.confirm(`确定要叫号 ${num} 吗？`, '提示', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'info'
  }).then(() => {
    callNext(type)
  }).catch(() => {})
}

const recall = () => {
  if (!currentNumber.value) {
    ElMessage.warning('暂无正在叫号的号码')
    return
  }
  speakCall(currentNumber.value, queueTypeName[currentQueueType.value])
  ElMessage.success('已重叫')
}

const speakCurrent = () => {
  recall()
}

const speakCall = (number, typeName) => {
  try {
    const text = `请${number}号顾客，${typeName}，请到前台就餐`
    if ('speechSynthesis' in window) {
      window.speechSynthesis.cancel()
      const utterance = new SpeechSynthesisUtterance(text)
      utterance.lang = 'zh-CN'
      utterance.rate = 0.85
      utterance.volume = 1
      utterance.pitch = 1
      window.speechSynthesis.speak(utterance)
    }

    if (window.electronAPI) {
      window.electronAPI.speak?.(text)
    }
  } catch (e) {
    console.error('Speak error:', e)
  }
}

const markArrived = async () => {
  if (!currentNumber.value) {
    ElMessage.warning('请先选择号码')
    return
  }
  const queueId = `${selectedStore.value}:${currentNumber.value}`
  try {
    await request.post('/queue2/arrive', {
      store_id: selectedStore.value,
      queue_id: queueId,
      table_no: ''
    })
    ElMessage.success('已标记入座')
    refreshQueues()
  } catch (e) {
    ElMessage.error(e.message || '操作失败')
  }
}

const markArrivedItem = (type, item) => {
  const num = item.queue_number || item.queueNumber
  const queueId = `${selectedStore.value}:${num}`
  ElMessageBox.prompt('请输入安排的桌号', '入座确认', {
    confirmButtonText: '确认入座',
    cancelButtonText: '取消',
    inputPattern: /.+/,
    inputErrorMessage: '请输入桌号'
  }).then(async ({ value: tableNo }) => {
    try {
      await request.post('/queue2/arrive', {
        store_id: selectedStore.value,
        queue_id: queueId,
        table_no: tableNo
      })
      if (num === currentNumber.value) {
        currentNumber.value = ''
        currentPeopleCount.value = 0
        callCount.value = 0
      }
      ElMessage.success('已标记入座')
      refreshQueues()
    } catch (e) {
      ElMessage.error(e.message || '操作失败')
    }
  }).catch(() => {})
}

const skipNumber = async () => {
  if (!currentNumber.value) {
    ElMessage.warning('请先选择号码')
    return
  }
  ElMessageBox.confirm(`确定要将 ${currentNumber.value} 过号吗？`, '过号确认', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning'
  }).then(async () => {
    const queueId = `${selectedStore.value}:${currentNumber.value}`
    try {
      await request.post('/queue2/cancel', {
        store_id: selectedStore.value,
        queue_id: queueId
      })
      currentNumber.value = ''
      currentPeopleCount.value = 0
      callCount.value = 0
      ElMessage.success('已过号')
      refreshQueues()
    } catch (e) {
      ElMessage.error(e.message || '操作失败')
    }
  }).catch(() => {})
}

const connectWebSocket = () => {
  if (ws) {
    try { ws.close() } catch (e) {}
    ws = null
  }

  const baseUrl = settingsStore.serverUrl || 'http://localhost:8080'
  const wsUrl = baseUrl.replace(/^http/, 'ws') + '/api/v1/queue/ws?store_id=' + selectedStore.value

  ws = new WebSocket(wsUrl)

  ws.onopen = () => {
    console.log('[WS] Queue connected')
    wsConnected.value = true
    reconnectAttempts = 0
  }

  ws.onmessage = (event) => {
    try {
      const msg = JSON.parse(event.data)
      console.log('[WS] Queue message:', msg)

      if (msg.type === 'call') {
        const num = msg.queue_number || msg.queueNumber
        const type = msg.queue_type || 'small'
        if (num) {
          currentNumber.value = num
          currentQueueType.value = type
          callCount.value = msg.call_count || msg.callCount || 1
          speakCall(num, queueTypeName[type])
          ElMessage({
            message: `叫号通知: ${num}`,
            type: 'info',
            duration: 3000
          })
        }
        refreshQueues()
      }
      if (msg.type === 'arrive' || msg.type === 'cancel') {
        refreshQueues()
      }
    } catch (e) {
      console.error('Parse WS message error:', e)
    }
  }

  ws.onclose = () => {
    console.log('[WS] Queue disconnected')
    wsConnected.value = false
    scheduleReconnect()
  }

  ws.onerror = (e) => {
    console.error('[WS] Queue error:', e)
    wsConnected.value = false
  }
}

const scheduleReconnect = () => {
  if (reconnectTimer) {
    clearTimeout(reconnectTimer)
    reconnectTimer = null
  }
  if (reconnectAttempts >= 10) return

  reconnectAttempts++
  const delay = Math.min(1000 * Math.pow(2, reconnectAttempts), 30000)
  reconnectTimer = setTimeout(() => {
    connectWebSocket()
  }, delay)
}

const onStoreChange = () => {
  refreshQueues()
  connectWebSocket()
}

onMounted(() => {
  refreshQueues()
  connectWebSocket()
})

onBeforeUnmount(() => {
  if (ws) {
    try { ws.close() } catch (e) {}
    ws = null
  }
  if (reconnectTimer) {
    clearTimeout(reconnectTimer)
    reconnectTimer = null
  }
  if ('speechSynthesis' in window) {
    window.speechSynthesis.cancel()
  }
})
</script>

<style lang="scss" scoped>
.queue-call-page {
  height: 100vh;
  display: flex;
  flex-direction: column;
  background: #f0f2f5;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 24px;
  background: #fff;
  border-bottom: 1px solid #e4e7ed;

  .header-left {
    display: flex;
    align-items: center;
    gap: 16px;
  }

  .page-title {
    margin: 0;
    font-size: 20px;
    font-weight: 600;
    color: #303133;
  }

  .ws-tag {
    margin-left: 8px;
  }

  .header-right {
    display: flex;
    align-items: center;
    gap: 12px;
  }
}

.page-body {
  flex: 1;
  display: flex;
  gap: 20px;
  padding: 20px;
  overflow: hidden;
}

.current-call-section {
  width: 360px;
  display: flex;
  flex-direction: column;
  gap: 16px;

  .current-call-card {
    background: #fff;
    border-radius: 12px;
    padding: 32px 24px;
    text-align: center;
    box-shadow: 0 2px 12px rgba(0, 0, 0, 0.08);

    .current-label {
      font-size: 16px;
      color: #909399;
      margin-bottom: 16px;
    }

    .current-number {
      font-size: 72px;
      font-weight: 700;
      line-height: 1.2;
      margin-bottom: 12px;
      font-family: 'DIN', monospace;

      &.small { color: #409eff; }
      &.medium { color: #67c23a; }
      &.large { color: #e6a23c; }
    }

    .current-info {
      font-size: 14px;
      color: #606266;
      display: flex;
      justify-content: center;
      gap: 8px;
    }
  }

  .call-actions {
    display: flex;
    flex-direction: column;
    gap: 10px;

    .el-button {
      width: 100%;
    }
  }

  .quick-actions {
    display: flex;
    gap: 8px;

    .el-button {
      flex: 1;
    }
  }
}

.queue-lists-section {
  flex: 1;
  display: flex;
  gap: 16px;
  overflow: hidden;

  .queue-panel {
    flex: 1;
    display: flex;
    flex-direction: column;
    background: #fff;
    border-radius: 12px;
    overflow: hidden;
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);

    .queue-panel-header {
      display: flex;
      justify-content: space-between;
      align-items: center;
      padding: 16px 20px;
      border-bottom: 1px solid #ebeef5;
      background: #fafafa;

      .panel-title {
        font-size: 16px;
        font-weight: 600;
        color: #303133;
      }
    }

    .queue-list {
      flex: 1;
      overflow-y: auto;
      padding: 8px;

      .queue-item {
        display: flex;
        align-items: center;
        padding: 12px 16px;
        border-radius: 8px;
        margin-bottom: 8px;
        background: #f5f7fa;
        transition: all 0.2s;

        &:hover {
          background: #ecf5ff;
        }

        &.current {
          background: #ecf5ff;
          border: 2px solid #409eff;
        }

        &.called {
          background: #f0f9eb;
        }

        .queue-number {
          font-size: 22px;
          font-weight: 700;
          color: #303133;
          min-width: 70px;
        }

        .queue-meta {
          flex: 1;
          display: flex;
          flex-direction: column;
          gap: 4px;

          .call-tag {
            font-size: 12px;
            color: #e6a23c;
          }
        }

        .queue-actions {
          display: flex;
          flex-direction: column;
          gap: 4px;
        }
      }

      .empty-queue {
        display: flex;
        justify-content: center;
        align-items: center;
        height: 200px;
      }
    }
  }
}
</style>
