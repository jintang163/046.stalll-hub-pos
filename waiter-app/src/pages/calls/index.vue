<template>
  <view class="calls-page">
    <view class="ws-status-bar" :class="{ connected: wsStore.isConnected }">
      <view class="ws-dot"></view>
      <text class="ws-text">{{ wsStore.isConnected ? '实时连接中' : '连接断开，正在重连...' }}</text>
    </view>

    <view class="filter-tabs">
      <view 
        class="filter-tab" 
        :class="{ active: currentStatus === status.key }"
        v-for="status in statusFilters"
        :key="status.key"
        @click="currentStatus = status.key">
        {{ status.label }}
      </view>
    </view>

    <scroll-view scroll-y class="calls-scroll" @refresherrefresh="onRefresh" :refresher-triggered="refreshing">
      <view class="call-card" v-for="call in filteredCalls" :key="call.id" :class="{ pending: call.status === 1 }">
        <view class="call-header">
          <view class="call-type" :class="`type-${call.call_type}`">
            {{ getCallTypeText(call.call_type) }}
          </view>
          <view class="call-time">{{ formatTime(call.created_at) }}</view>
        </view>

        <view class="call-body">
          <view class="call-table">
            <text class="table-label">桌号</text>
            <text class="table-no">{{ call.table_no }}</text>
          </view>
          <view class="call-content" v-if="call.content">
            {{ call.content }}
          </view>
        </view>

        <view class="call-footer">
          <view class="call-status" :class="`status-${call.status}`">
            {{ call.status === 1 ? '待处理' : '已处理' }}
          </view>
          <view class="call-actions" v-if="call.status === 1">
            <view class="action-btn btn-handle" @click="handleCall(call)">
              前往处理
            </view>
          </view>
        </view>
      </view>

      <view class="empty" v-if="filteredCalls.length === 0 && !loading">
        暂无呼叫记录
      </view>
    </scroll-view>
  </view>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import { useUserStore } from '../../store/user'
import { useWebSocketStore } from '../../store/websocket'
import { waiterApi } from '../../services/waiter'
import type { WaiterCall } from '../../types'

const userStore = useUserStore()
const wsStore = useWebSocketStore()

const calls = ref<WaiterCall[]>([])
const currentStatus = ref<number>(0)
const loading = ref(false)
const refreshing = ref(false)

const statusFilters = [
  { key: 0, label: '全部' },
  { key: 1, label: '待处理' },
  { key: 2, label: '已处理' }
]

const filteredCalls = computed(() => {
  if (currentStatus.value === 0) return calls.value
  return calls.value.filter(c => c.status === currentStatus.value)
})

const getCallTypeText = (type: string) => {
  const map: Record<string, string> = {
    service: '呼叫服务',
    water: '需要加水',
    pay: '需要结账',
    other: '其他'
  }
  return map[type] || '呼叫服务'
}

const formatTime = (time: string) => {
  if (!time) return ''
  const t = new Date(time)
  return `${String(t.getHours()).padStart(2, '0')}:${String(t.getMinutes()).padStart(2, '0')}`
}

const loadCalls = async () => {
  if (!userStore.userInfo?.store_id) return
  loading.value = true
  try {
    const result = await waiterApi.getCalls(userStore.userInfo.store_id, 0)
    calls.value = result
  } catch (e: any) {
    console.error('Load calls failed:', e)
  } finally {
    loading.value = false
    refreshing.value = false
  }
}

const onRefresh = () => {
  refreshing.value = true
  loadCalls()
}

const handleCall = async (call: WaiterCall) => {
  uni.showModal({
    title: '确认处理',
    content: `确定开始处理 ${call.table_no} 桌的呼叫？`,
    success: async (res) => {
      if (res.confirm) {
        try {
          await waiterApi.handleCall(call.id)
          wsStore.clearCall(call.id)
          uni.showToast({ title: '已处理', icon: 'success' })
          loadCalls()
        } catch (e) {
          console.error('Handle call failed:', e)
        }
      }
    }
  })
}

onMounted(() => {
  if (!userStore.isLoggedIn()) {
    uni.reLaunch({ url: '/pages/login/index' })
    return
  }
})

onShow(() => {
  if (userStore.isLoggedIn()) {
    loadCalls()
  }
})
</script>

<style lang="scss" scoped>
.calls-page {
  min-height: 100vh;
  background: #f5f5f5;
  display: flex;
  flex-direction: column;
}

.ws-status-bar {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 12rpx 24rpx;
  background: #fff7e6;

  &.connected {
    background: #e8f5e9;

    .ws-dot {
      background: #07c160;
    }

    .ws-text {
      color: #07c160;
    }
  }

  .ws-dot {
    width: 16rpx;
    height: 16rpx;
    border-radius: 50%;
    background: #ff976a;
    margin-right: 12rpx;
    animation: pulse 2s infinite;
  }

  .ws-text {
    font-size: 24rpx;
    color: #ff976a;
  }
}

@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.5; }
}

.filter-tabs {
  display: flex;
  background: #fff;
  padding: 0 8rpx;
  box-shadow: 0 2rpx 8rpx rgba(0, 0, 0, 0.04);
}

.filter-tab {
  flex: 1;
  text-align: center;
  padding: 28rpx 0;
  font-size: 26rpx;
  color: #646566;
  position: relative;

  &.active {
    color: #1989fa;
    font-weight: bold;

    &::after {
      content: '';
      position: absolute;
      bottom: 0;
      left: 50%;
      transform: translateX(-50%);
      width: 48rpx;
      height: 6rpx;
      background: #1989fa;
      border-radius: 3rpx;
    }
  }
}

.calls-scroll {
  flex: 1;
  height: 0;
  padding: 16rpx;
}

.call-card {
  background: #fff;
  border-radius: 16rpx;
  margin-bottom: 20rpx;
  overflow: hidden;
  border-left: 8rpx solid #ebedf0;

  &.pending {
    border-left-color: #ee0a24;
  }
}

.call-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 20rpx 24rpx;
  border-bottom: 1rpx solid #f2f3f5;
}

.call-type {
  font-size: 24rpx;
  padding: 6rpx 16rpx;
  border-radius: 8rpx;

  &.type-service {
    background: #e6f7ff;
    color: #1989fa;
  }

  &.type-water {
    background: #e6fffb;
    color: #08979c;
  }

  &.type-pay {
    background: #f9f0ff;
    color: #722ed1;
  }

  &.type-other {
    background: #fff7e6;
    color: #fa8c16;
  }
}

.call-time {
  font-size: 22rpx;
  color: #969799;
}

.call-body {
  padding: 20rpx 24rpx;
}

.call-table {
  display: flex;
  align-items: center;

  .table-label {
    font-size: 26rpx;
    color: #969799;
    margin-right: 12rpx;
  }

  .table-no {
    font-size: 36rpx;
    font-weight: bold;
    color: #323233;
  }
}

.call-content {
  font-size: 26rpx;
  color: #646566;
  margin-top: 12rpx;
  padding: 16rpx;
  background: #f7f8fa;
  border-radius: 8rpx;
}

.call-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16rpx 24rpx;
  border-top: 1rpx solid #f2f3f5;
  background: #fafafa;
}

.call-status {
  font-size: 24rpx;
  padding: 6rpx 16rpx;
  border-radius: 8rpx;

  &.status-1 {
    background: #fff1f0;
    color: #ee0a24;
  }

  &.status-2 {
    background: #e8f5e9;
    color: #07c160;
  }
}

.call-actions {
  .action-btn {
    font-size: 24rpx;
    padding: 10rpx 28rpx;
    border-radius: 32rpx;

    &.btn-handle {
      background: linear-gradient(135deg, #1989fa 0%, #07c160 100%);
      color: #fff;
    }
  }
}

.empty {
  padding: 120rpx 0;
  text-align: center;
  color: #969799;
  font-size: 28rpx;
}
</style>
