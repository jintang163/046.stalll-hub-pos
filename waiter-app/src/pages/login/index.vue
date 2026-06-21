<template>
  <view class="login-page">
    <view class="login-header">
      <view class="logo">🍽️</view>
      <view class="app-name">餐厅服务员端</view>
      <view class="app-desc">高效管理 · 贴心服务</view>
    </view>

    <view class="login-form">
      <view class="form-item">
        <view class="form-label">账号</view>
        <input 
          class="form-input" 
          v-model="username" 
          placeholder="请输入账号" 
          placeholder-class="input-placeholder"
        />
      </view>

      <view class="form-item">
        <view class="form-label">密码</view>
        <input 
          class="form-input" 
          v-model="password" 
          type="password"
          placeholder="请输入密码" 
          placeholder-class="input-placeholder"
        />
      </view>

      <view class="btn-login" @click="handleLogin" :class="{ disabled: loading }">
        {{ loading ? '登录中...' : '登录' }}
      </view>
    </view>

    <view class="login-footer">
      <view class="tip">请使用管理员分配的账号登录</view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useUserStore } from '../../store/user'
import { useWebSocketService } from '../../services/websocket'

const username = ref('')
const password = ref('')
const loading = ref(false)

const userStore = useUserStore()
const wsService = useWebSocketService()

const handleLogin = async () => {
  if (!username.value.trim()) {
    uni.showToast({ title: '请输入账号', icon: 'none' })
    return
  }
  if (!password.value.trim()) {
    uni.showToast({ title: '请输入密码', icon: 'none' })
    return
  }

  loading.value = true
  try {
    await userStore.login(username.value.trim(), password.value.trim())
    wsService.connect()
    uni.showToast({ title: '登录成功', icon: 'success' })
    setTimeout(() => {
      uni.switchTab({ url: '/pages/tables/index' })
    }, 1000)
  } catch (e: any) {
    console.error('Login failed:', e)
  } finally {
    loading.value = false
  }
}
</script>

<style lang="scss" scoped>
.login-page {
  min-height: 100vh;
  background: linear-gradient(135deg, #1989fa 0%, #07c160 100%);
  padding: 0 60rpx;
  display: flex;
  flex-direction: column;
}

.login-header {
  padding-top: 160rpx;
  padding-bottom: 100rpx;
  text-align: center;

  .logo {
    font-size: 120rpx;
    margin-bottom: 32rpx;
  }

  .app-name {
    font-size: 52rpx;
    color: #fff;
    font-weight: bold;
    margin-bottom: 16rpx;
  }

  .app-desc {
    font-size: 28rpx;
    color: rgba(255, 255, 255, 0.8);
  }
}

.login-form {
  background: #fff;
  border-radius: 24rpx;
  padding: 60rpx 48rpx;
  box-shadow: 0 8rpx 32rpx rgba(0, 0, 0, 0.1);

  .form-item {
    margin-bottom: 40rpx;

    .form-label {
      font-size: 28rpx;
      color: #646566;
      margin-bottom: 16rpx;
    }

    .form-input {
      height: 88rpx;
      background: #f7f8fa;
      border-radius: 12rpx;
      padding: 0 24rpx;
      font-size: 30rpx;
      color: #323233;
    }

    .input-placeholder {
      color: #c8c9cc;
    }
  }

  .btn-login {
    margin-top: 40rpx;
    height: 88rpx;
    line-height: 88rpx;
    background: linear-gradient(135deg, #1989fa 0%, #07c160 100%);
    color: #fff;
    text-align: center;
    border-radius: 44rpx;
    font-size: 32rpx;
    font-weight: bold;

    &.disabled {
      opacity: 0.6;
    }
  }
}

.login-footer {
  margin-top: auto;
  padding-bottom: 80rpx;
  text-align: center;

  .tip {
    font-size: 24rpx;
    color: rgba(255, 255, 255, 0.7);
  }
}
</style>
