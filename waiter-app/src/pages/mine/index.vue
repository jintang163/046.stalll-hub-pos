<template>
  <view class="mine-page">
    <view class="user-header">
      <view class="avatar">{{ userStore.userInfo?.real_name?.charAt(0) || userStore.userInfo?.username?.charAt(0) || 'W' }}</view>
      <view class="user-info">
        <view class="user-name">{{ userStore.userInfo?.real_name || userStore.userInfo?.username }}</view>
        <view class="user-role">{{ getRoleText(userStore.userInfo?.role) }}</view>
      </view>
    </view>

    <view class="stats-card">
      <view class="stat-item">
        <view class="stat-value">{{ stats.total_tables }}</view>
        <view class="stat-label">总桌数</view>
      </view>
      <view class="stat-divider"></view>
      <view class="stat-item">
        <view class="stat-value text-warning">{{ stats.occupied_tables + stats.ordered_tables }}</view>
        <view class="stat-label">使用中</view>
      </view>
      <view class="stat-divider"></view>
      <view class="stat-item">
        <view class="stat-value text-danger">{{ stats.pending_calls }}</view>
        <view class="stat-label">待处理</view>
      </view>
    </view>

    <view class="menu-list">
      <view class="menu-item" @click="goTo('/pages/tables/index')">
        <view class="menu-icon">🪑</view>
        <view class="menu-text">桌位管理</view>
        <view class="menu-arrow">›</view>
      </view>
      <view class="menu-item" @click="goTo('/pages/orders/index')">
        <view class="menu-icon">📋</view>
        <view class="menu-text">订单管理</view>
        <view class="menu-arrow">›</view>
      </view>
      <view class="menu-item" @click="goTo('/pages/calls/index')">
        <view class="menu-icon">🔔</view>
        <view class="menu-text">呼叫记录</view>
        <view class="menu-badge" v-if="wsStore.pendingCalls.length > 0">{{ wsStore.pendingCalls.length }}</view>
        <view class="menu-arrow">›</view>
      </view>
    </view>

    <view class="menu-list">
      <view class="menu-item" @click="checkConnection">
        <view class="menu-icon">📡</view>
        <view class="menu-text">连接状态</view>
        <view class="menu-value" :class="{ connected: wsStore.isConnected }">
          {{ wsStore.isConnected ? '已连接' : '未连接' }}
        </view>
      </view>
      <view class="menu-item" @click="refreshData">
        <view class="menu-icon">🔄</view>
        <view class="menu-text">刷新数据</view>
        <view class="menu-arrow">›</view>
      </view>
    </view>

    <view class="menu-list">
      <view class="menu-item danger" @click="handleLogout">
        <view class="menu-icon">🚪</view>
        <view class="menu-text">退出登录</view>
      </view>
    </view>

    <view class="app-info">
      <text>服务员端 v1.0.0</text>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import { useUserStore } from '../../store/user'
import { useWebSocketStore } from '../../store/websocket'
import { waiterApi } from '../../services/waiter'
import { useWebSocketService } from '../../services/websocket'
import type { WaiterStats } from '../../types'

const userStore = useUserStore()
const wsStore = useWebSocketStore()
const wsService = useWebSocketService()

const stats = ref<WaiterStats>({
  total_tables: 0,
  idle_tables: 0,
  occupied_tables: 0,
  ordered_tables: 0,
  pending_calls: 0,
  pending_orders: 0
})

const getRoleText = (role?: string) => {
  const map: Record<string, string> = {
    admin: '管理员',
    manager: '店长',
    staff: '服务员',
    waiter: '服务员'
  }
  return map[role || ''] || '服务员'
}

const loadStats = async () => {
  if (!userStore.userInfo?.store_id) return
  try {
    stats.value = await waiterApi.getStats(userStore.userInfo.store_id)
  } catch (e) {
    console.error('Load stats failed:', e)
  }
}

const goTo = (url: string) => {
  uni.switchTab({ url })
}

const checkConnection = () => {
  if (wsStore.isConnected) {
    uni.showToast({ title: 'WebSocket已连接', icon: 'success' })
  } else {
    uni.showModal({
      title: '连接断开',
      content: 'WebSocket连接已断开，是否重新连接？',
      success: (res) => {
        if (res.confirm) {
          wsService.connect()
        }
      }
    })
  }
}

const refreshData = () => {
  loadStats()
  uni.showToast({ title: '已刷新', icon: 'success' })
}

const handleLogout = () => {
  uni.showModal({
    title: '确认退出',
    content: '确定要退出登录吗？',
    success: async (res) => {
      if (res.confirm) {
        wsService.disconnect()
        await userStore.logout()
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
    loadStats()
  }
})
</script>

<style lang="scss" scoped>
.mine-page {
  min-height: 100vh;
  background: #f5f5f5;
  padding-bottom: 40rpx;
}

.user-header {
  display: flex;
  align-items: center;
  padding: 60rpx 40rpx;
  background: linear-gradient(135deg, #1989fa 0%, #07c160 100%);

  .avatar {
    width: 120rpx;
    height: 120rpx;
    line-height: 120rpx;
    text-align: center;
    background: rgba(255, 255, 255, 0.3);
    border-radius: 50%;
    font-size: 48rpx;
    color: #fff;
    font-weight: bold;
  }

  .user-info {
    margin-left: 24rpx;

    .user-name {
      font-size: 36rpx;
      font-weight: bold;
      color: #fff;
    }

    .user-role {
      font-size: 24rpx;
      color: rgba(255, 255, 255, 0.8);
      margin-top: 8rpx;
    }
  }
}

.stats-card {
  display: flex;
  background: #fff;
  margin: -30rpx 24rpx 24rpx;
  border-radius: 16rpx;
  padding: 32rpx 0;
  box-shadow: 0 4rpx 16rpx rgba(0, 0, 0, 0.08);

  .stat-item {
    flex: 1;
    text-align: center;

    .stat-value {
      font-size: 40rpx;
      font-weight: bold;
      color: #323233;

      &.text-warning {
        color: #ff976a;
      }

      &.text-danger {
        color: #ee0a24;
      }
    }

    .stat-label {
      font-size: 24rpx;
      color: #969799;
      margin-top: 8rpx;
    }
  }

  .stat-divider {
    width: 1rpx;
    background: #ebedf0;
    margin: 8rpx 0;
  }
}

.menu-list {
  background: #fff;
  margin: 24rpx;
  border-radius: 16rpx;
  overflow: hidden;
}

.menu-item {
  display: flex;
  align-items: center;
  padding: 32rpx 24rpx;
  border-bottom: 1rpx solid #f2f3f5;
  position: relative;

  &:last-child {
    border-bottom: none;
  }

  &.danger {
    .menu-text {
      color: #ee0a24;
    }
  }

  .menu-icon {
    font-size: 40rpx;
    margin-right: 20rpx;
  }

  .menu-text {
    flex: 1;
    font-size: 28rpx;
    color: #323233;
  }

  .menu-value {
    font-size: 26rpx;
    color: #969799;
    margin-right: 12rpx;

    &.connected {
      color: #07c160;
    }
  }

  .menu-badge {
    position: absolute;
    right: 60rpx;
    top: 50%;
    transform: translateY(-50%);
    min-width: 36rpx;
    height: 36rpx;
    line-height: 36rpx;
    padding: 0 10rpx;
    border-radius: 18rpx;
    background: #ee0a24;
    color: #fff;
    font-size: 20rpx;
    text-align: center;
  }

  .menu-arrow {
    font-size: 32rpx;
    color: #c8c9cc;
  }
}

.app-info {
  text-align: center;
  padding: 40rpx;
  font-size: 24rpx;
  color: #c8c9cc;
}
</style>
