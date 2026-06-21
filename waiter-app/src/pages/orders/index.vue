<template>
  <view class="orders-page">
    <view class="filter-tabs">
      <view 
        class="filter-tab" 
        v-for="tab in tabs" 
        :key="tab.key"
        :class="{ active: currentTab === tab.key }"
        @click="switchTab(tab.key)">
        {{ tab.label }}
        <view class="tab-count" v-if="tab.count > 0">{{ tab.count }}</view>
      </view>
    </view>

    <scroll-view scroll-y class="orders-scroll" @refresherrefresh="onRefresh" :refresher-triggered="refreshing">
      <view class="order-card" v-for="order in orders" :key="order.id" @click="viewDetail(order)">
        <view class="order-header">
          <view class="order-no">订单号：{{ order.order_no }}</view>
          <view class="order-status" :class="`status-${order.order_status}`">
            {{ getStatusText(order.order_status, order.pay_status) }}
          </view>
        </view>

        <view class="order-body">
          <view class="order-table" v-if="order.table_no">
            <text class="label">桌号：</text>
            <text class="value">{{ order.table_no }}</text>
          </view>
          <view class="order-items">
            <view class="order-item" v-for="item in order.items.slice(0, 3)" :key="item.id">
              <image class="item-image" :src="item.image || '/static/default-food.png'" mode="aspectFill" />
              <view class="item-info">
                <view class="item-name">{{ item.product_name }}</view>
                <view class="item-spec">{{ item.sku_name }}</view>
              </view>
              <view class="item-status" :class="`cook-${item.cook_status}`">
                {{ getCookStatusText(item.cook_status) }}
              </view>
              <view class="item-qty">x{{ item.quantity }}</view>
            </view>
            <view class="more-items" v-if="order.items.length > 3">
              等{{ order.items.length }}件商品
            </view>
          </view>
        </view>

        <view class="order-footer">
          <view class="order-time">{{ formatTime(order.created_at) }}</view>
          <view class="order-amount">
            <text class="amount-label">合计：</text>
            <text class="amount-value">¥{{ order.pay_amount }}</text>
          </view>
        </view>

        <view class="order-actions" v-if="order.order_status <= 3">
          <view class="action-btn btn-serve" @click.stop="batchServe(order)" v-if="getPendingServeCount(order) > 0">
            上菜({{ getPendingServeCount(order) }})
          </view>
          <view class="action-btn btn-detail" @click.stop="viewDetail(order)">
            查看详情
          </view>
        </view>
      </view>

      <view class="empty" v-if="orders.length === 0 && !loading">
        暂无订单
      </view>
    </scroll-view>
  </view>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import { useUserStore } from '../../store/user'
import { orderApi } from '../../services/order'
import { waiterApi } from '../../services/waiter'
import type { OrderDetail } from '../../types'

const userStore = useUserStore()

const orders = ref<OrderDetail[]>([])
const currentTab = ref<number>(0)
const loading = ref(false)
const refreshing = ref(false)

const tabs = computed(() => [
  { key: 0, label: '全部', count: 0 },
  { key: 1, label: '待接单', count: 0 },
  { key: 2, label: '制作中', count: 0 },
  { key: 3, label: '待上菜', count: 0 },
  { key: 4, label: '已完成', count: 0 }
])

const getStatusText = (orderStatus: number, payStatus: number) => {
  if (orderStatus === -1) return '已取消'
  if (payStatus === 1) return '已支付'
  const map: Record<number, string> = {
    1: '待接单',
    2: '制作中',
    3: '待上菜',
    4: '已完成',
    5: '已结账'
  }
  return map[orderStatus] || '未知'
}

const getCookStatusText = (status: number) => {
  const map: Record<number, string> = {
    0: '待制作',
    1: '制作中',
    2: '已完成',
    3: '已上菜'
  }
  return map[status] || '未知'
}

const getPendingServeCount = (order: OrderDetail) => {
  return order.items.filter(item => item.cook_status === 2).length
}

const formatTime = (time: string) => {
  if (!time) return ''
  const t = new Date(time)
  return `${String(t.getMonth() + 1).padStart(2, '0')}-${String(t.getDate()).padStart(2, '0')} ${String(t.getHours()).padStart(2, '0')}:${String(t.getMinutes()).padStart(2, '0')}`
}

const loadOrders = async () => {
  if (!userStore.userInfo?.store_id) return
  loading.value = true
  try {
    const params: any = {
      store_id: userStore.userInfo.store_id,
      page_size: 50
    }
    if (currentTab.value >= 1 && currentTab.value <= 4) {
      params.order_status = currentTab.value
    }
    const result = await orderApi.getOrders(params)
    orders.value = result.list as OrderDetail[]
  } catch (e: any) {
    console.error('Load orders failed:', e)
  } finally {
    loading.value = false
    refreshing.value = false
  }
}

const onRefresh = () => {
  refreshing.value = true
  loadOrders()
}

const switchTab = (tab: number) => {
  currentTab.value = tab
  loadOrders()
}

const viewDetail = (order: OrderDetail) => {
  uni.navigateTo({
    url: `/pages/orders/detail?id=${order.id}`
  })
}

const batchServe = async (order: OrderDetail) => {
  const readyItems = order.items.filter(item => item.cook_status === 2)
  if (readyItems.length === 0) {
    uni.showToast({ title: '暂无可上菜菜品', icon: 'none' })
    return
  }

  uni.showModal({
    title: '确认上菜',
    content: `确定将 ${readyItems.length} 道菜品标记为已上菜？`,
    success: async (res) => {
      if (res.confirm) {
        try {
          await waiterApi.markItemsServed(readyItems.map(item => item.id))
          uni.showToast({ title: '标记成功', icon: 'success' })
          loadOrders()
        } catch (e) {
          console.error('Mark served failed:', e)
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
    loadOrders()
  }
})
</script>

<style lang="scss" scoped>
.orders-page {
  min-height: 100vh;
  background: #f5f5f5;
  display: flex;
  flex-direction: column;
}

.filter-tabs {
  display: flex;
  background: #fff;
  padding: 0 8rpx;
  box-shadow: 0 2rpx 8rpx rgba(0, 0, 0, 0.04);
  position: sticky;
  top: 0;
  z-index: 10;
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

.tab-count {
  display: inline-block;
  min-width: 32rpx;
  height: 32rpx;
  line-height: 32rpx;
  padding: 0 8rpx;
  margin-left: 4rpx;
  background: #ee0a24;
  color: #fff;
  font-size: 20rpx;
  border-radius: 16rpx;
}

.orders-scroll {
  flex: 1;
  height: 0;
  padding: 16rpx;
}

.order-card {
  background: #fff;
  border-radius: 16rpx;
  margin-bottom: 20rpx;
  overflow: hidden;
}

.order-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 24rpx;
  border-bottom: 1rpx solid #f2f3f5;

  .order-no {
    font-size: 26rpx;
    color: #646566;
  }

  .order-status {
    font-size: 24rpx;
    padding: 6rpx 16rpx;
    border-radius: 8rpx;

    &.status-1 {
      background: #fff7e6;
      color: #ff976a;
    }

    &.status-2 {
      background: #e6f7ff;
      color: #1989fa;
    }

    &.status-3 {
      background: #f3eaff;
      color: #7232dd;
    }

    &.status-4,
    &.status-5 {
      background: #e8f5e9;
      color: #07c160;
    }

    &.status--1 {
      background: #f0f0f0;
      color: #969799;
    }
  }
}

.order-body {
  padding: 16rpx 24rpx;
}

.order-table {
  margin-bottom: 16rpx;

  .label {
    font-size: 26rpx;
    color: #646566;
  }

  .value {
    font-size: 26rpx;
    color: #323233;
    font-weight: bold;
  }
}

.order-items {
  .order-item {
    display: flex;
    align-items: center;
    padding: 12rpx 0;
  }

  .item-image {
    width: 80rpx;
    height: 80rpx;
    border-radius: 8rpx;
    background: #f7f8fa;
    flex-shrink: 0;
  }

  .item-info {
    flex: 1;
    margin-left: 16rpx;
    overflow: hidden;

    .item-name {
      font-size: 26rpx;
      color: #323233;
      overflow: hidden;
      text-overflow: ellipsis;
      white-space: nowrap;
    }

    .item-spec {
      font-size: 22rpx;
      color: #969799;
      margin-top: 4rpx;
    }
  }

  .item-status {
    font-size: 20rpx;
    padding: 4rpx 12rpx;
    border-radius: 6rpx;
    margin-right: 16rpx;

    &.cook-0 {
      background: #fff7e6;
      color: #ff976a;
    }

    &.cook-1 {
      background: #e6f7ff;
      color: #1989fa;
    }

    &.cook-2 {
      background: #f3eaff;
      color: #7232dd;
    }

    &.cook-3 {
      background: #e8f5e9;
      color: #07c160;
    }
  }

  .item-qty {
    font-size: 24rpx;
    color: #646566;
    flex-shrink: 0;
  }

  .more-items {
    font-size: 24rpx;
    color: #969799;
    text-align: center;
    padding: 12rpx 0;
  }
}

.order-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16rpx 24rpx;
  border-top: 1rpx solid #f2f3f5;

  .order-time {
    font-size: 22rpx;
    color: #969799;
  }

  .order-amount {
    .amount-label {
      font-size: 24rpx;
      color: #646566;
    }

    .amount-value {
      font-size: 32rpx;
      color: #ee0a24;
      font-weight: bold;
    }
  }
}

.order-actions {
  display: flex;
  gap: 16rpx;
  padding: 16rpx 24rpx 24rpx;

  .action-btn {
    flex: 1;
    text-align: center;
    padding: 16rpx 0;
    border-radius: 8rpx;
    font-size: 26rpx;

    &.btn-serve {
      background: #07c160;
      color: #fff;
    }

    &.btn-detail {
      background: #f7f8fa;
      color: #323233;
      border: 1rpx solid #ebedf0;
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
