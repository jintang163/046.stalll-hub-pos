<template>
  <view class="order-detail-page" v-if="order">
    <scroll-view scroll-y class="detail-scroll">
      <view class="section order-info-section">
        <view class="section-title">订单信息</view>
        <view class="info-row">
          <text class="info-label">订单号</text>
          <text class="info-value">{{ order.order_no }}</text>
          <text class="copy-btn" @click="copyOrderNo">复制</text>
        </view>
        <view class="info-row">
          <text class="info-label">桌号</text>
          <text class="info-value">{{ order.table_no || '-' }}</text>
        </view>
        <view class="info-row">
          <text class="info-label">订单状态</text>
          <text class="info-value" :class="`status-${order.order_status}`">
            {{ getStatusText(order.order_status, order.pay_status) }}
          </text>
        </view>
        <view class="info-row">
          <text class="info-label">下单时间</text>
          <text class="info-value">{{ formatTime(order.created_at) }}</text>
        </view>
        <view class="info-row" v-if="order.remark">
          <text class="info-label">备注</text>
          <text class="info-value">{{ order.remark }}</text>
        </view>
      </view>

      <view class="section items-section">
        <view class="section-header">
          <view class="section-title">菜品明细</view>
          <view class="section-actions" v-if="pendingReadyItems.length > 0">
            <view class="batch-serve-btn" @click="batchServeReady">
              全部上菜({{ pendingReadyItems.length }})
            </view>
          </view>
        </view>

        <view class="item-card" v-for="item in order.items" :key="item.id">
          <view class="item-main">
            <image class="item-image" :src="item.image || '/static/default-food.png'" mode="aspectFill" />
            <view class="item-info">
              <view class="item-name">{{ item.product_name }}</view>
              <view class="item-spec">{{ item.sku_name }}</view>
              <view class="item-price-row">
                <text class="item-price">¥{{ item.price }}</text>
                <text class="item-qty">x{{ item.quantity }}</text>
              </view>
            </view>
          </view>

          <view class="item-status-row">
            <view class="item-cook-status" :class="`cook-${item.cook_status}`">
              {{ getCookStatusText(item.cook_status) }}
            </view>
            <view class="item-actions">
              <view 
                class="action-btn serve-btn" 
                v-if="item.cook_status === 2"
                @click="serveItem(item)">
                标记上菜
              </view>
              <view 
                class="action-btn cooking-btn" 
                v-if="item.cook_status === 0"
                @click="startCooking(item)">
                开始制作
              </view>
              <view 
                class="action-btn refund-btn" 
                v-if="item.cook_status < 2 && order.order_status < 4"
                @click="refundItem(item)">
                退菜
              </view>
            </view>
          </view>
        </view>
      </view>

      <view class="section amount-section">
        <view class="section-title">金额明细</view>
        <view class="amount-row">
          <text class="amount-label">商品金额</text>
          <text class="amount-value">¥{{ order.total_amount }}</text>
        </view>
        <view class="amount-row" v-if="parseFloat(order.discount_amount) > 0">
          <text class="amount-label">优惠金额</text>
          <text class="amount-value discount">-¥{{ order.discount_amount }}</text>
        </view>
        <view class="amount-row" v-if="parseFloat(order.coupon_amount) > 0">
          <text class="amount-label">优惠券</text>
          <text class="amount-value discount">-¥{{ order.coupon_amount }}</text>
        </view>
        <view class="amount-row total">
          <text class="amount-label">实付金额</text>
          <text class="amount-value">¥{{ order.pay_amount }}</text>
        </view>
        <view class="amount-row" v-if="order.pay_status === 1">
          <text class="amount-label">支付方式</text>
          <text class="amount-value">{{ getPayMethodText(order.pay_method) }}</text>
        </view>
      </view>
    </scroll-view>

    <view class="footer-actions">
      <view class="footer-btn btn-refresh" @click="loadDetail">刷新</view>
      <view 
        class="footer-btn btn-add" 
        v-if="order.order_status < 4"
        @click="addDish">加菜</view>
      <view 
        class="footer-btn btn-complete" 
        v-if="allServed && order.order_status === 3"
        @click="completeOrder">订单完成</view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { onLoad, onShow } from '@dcloudio/uni-app'
import { useUserStore } from '../../store/user'
import { useCartStore } from '../../store/cart'
import { orderApi } from '../../services/order'
import { waiterApi } from '../../services/waiter'
import type { OrderDetail, OrderItem } from '../../types'

const userStore = useUserStore()
const cartStore = useCartStore()

const orderId = ref<number>(0)
const order = ref<OrderDetail | null>(null)
const loading = ref(false)

const pendingReadyItems = computed(() => {
  if (!order.value) return []
  return order.value.items.filter(item => item.cook_status === 2)
})

const allServed = computed(() => {
  if (!order.value || order.value.items.length === 0) return false
  return order.value.items.every(item => item.cook_status >= 3)
})

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
    2: '待上菜',
    3: '已上菜'
  }
  return map[status] || '未知'
}

const getPayMethodText = (method: string) => {
  const map: Record<string, string> = {
    wechat: '微信支付',
    alipay: '支付宝',
    cash: '现金'
  }
  return map[method] || method || '未支付'
}

const formatTime = (time: string) => {
  if (!time) return ''
  const t = new Date(time)
  return `${t.getFullYear()}-${String(t.getMonth() + 1).padStart(2, '0')}-${String(t.getDate()).padStart(2, '0')} ${String(t.getHours()).padStart(2, '0')}:${String(t.getMinutes()).padStart(2, '0')}:${String(t.getSeconds()).padStart(2, '0')}`
}

const loadDetail = async () => {
  if (!orderId.value) return
  loading.value = true
  try {
    const result = await orderApi.getOrderDetail(orderId.value)
    order.value = result
  } catch (e: any) {
    console.error('Load order detail failed:', e)
  } finally {
    loading.value = false
  }
}

const copyOrderNo = () => {
  if (!order.value) return
  uni.setClipboardData({
    data: order.value.order_no,
    success: () => {
      uni.showToast({ title: '已复制', icon: 'success' })
    }
  })
}

const serveItem = async (item: OrderItem) => {
  uni.showModal({
    title: '确认上菜',
    content: `确定将「${item.product_name}」标记为已上菜？`,
    success: async (res) => {
      if (res.confirm) {
        try {
          await waiterApi.markItemsServed([item.id])
          uni.showToast({ title: '上菜成功', icon: 'success' })
          loadDetail()
        } catch (e) {
          console.error('Serve item failed:', e)
        }
      }
    }
  })
}

const batchServeReady = async () => {
  if (pendingReadyItems.value.length === 0) return
  uni.showModal({
    title: '批量上菜',
    content: `确定将 ${pendingReadyItems.value.length} 道菜品标记为已上菜？`,
    success: async (res) => {
      if (res.confirm) {
        try {
          await waiterApi.markItemsServed(pendingReadyItems.value.map(item => item.id))
          uni.showToast({ title: '批量上菜成功', icon: 'success' })
          loadDetail()
        } catch (e) {
          console.error('Batch serve failed:', e)
        }
      }
    }
  })
}

const startCooking = async (item: OrderItem) => {
  try {
    await waiterApi.updateItemCookStatus([item.id], 1)
    uni.showToast({ title: '已开始制作', icon: 'success' })
    loadDetail()
  } catch (e) {
    console.error('Start cooking failed:', e)
  }
}

const refundItem = async (item: OrderItem) => {
  uni.showModal({
    title: '退菜确认',
    content: `确定要退掉「${item.product_name}」吗？`,
    success: async (res) => {
      if (res.confirm && order.value) {
        try {
          await orderApi.refundOrder(order.value.id, {
            refund_type: 'partial',
            refund_amount: (parseFloat(item.price) * item.quantity).toFixed(2),
            refund_reason: '服务员退菜',
            items: [{ order_item_id: item.id, quantity: item.quantity }]
          })
          uni.showToast({ title: '退菜成功', icon: 'success' })
          loadDetail()
        } catch (e) {
          console.error('Refund item failed:', e)
        }
      }
    }
  })
}

const completeOrder = async () => {
  if (!order.value) return
  uni.showModal({
    title: '确认完成',
    content: '确定将该订单标记为已完成？',
    success: async (res) => {
      if (res.confirm) {
        try {
          await orderApi.updateStatus(order.value!.id, 4)
          uni.showToast({ title: '订单已完成', icon: 'success' })
          loadDetail()
        } catch (e) {
          console.error('Complete order failed:', e)
        }
      }
    }
  })
}

const addDish = () => {
  if (!order.value) return
  cartStore.clearCart()
  cartStore.setTable(0, order.value.table_no || '')
  uni.navigateTo({
    url: `/pages/menu/index?tableId=0&tableNo=${order.value.table_no}&orderId=${order.value.id}&action=add`
  })
}

onLoad((options: any) => {
  if (options?.id) {
    orderId.value = parseInt(options.id)
  }
})

onShow(() => {
  if (!userStore.isLoggedIn()) {
    uni.reLaunch({ url: '/pages/login/index' })
    return
  }
  if (orderId.value > 0) {
    loadDetail()
  }
})
</script>

<style lang="scss" scoped>
.order-detail-page {
  min-height: 100vh;
  background: #f5f5f5;
  display: flex;
  flex-direction: column;
}

.detail-scroll {
  flex: 1;
  height: 0;
  padding-bottom: 140rpx;
}

.section {
  background: #fff;
  margin: 20rpx;
  border-radius: 16rpx;
  padding: 24rpx;
}

.section-title {
  font-size: 30rpx;
  font-weight: bold;
  color: #323233;
  margin-bottom: 20rpx;
}

.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16rpx;

  .section-actions {
    .batch-serve-btn {
      background: #07c160;
      color: #fff;
      padding: 10rpx 24rpx;
      border-radius: 32rpx;
      font-size: 24rpx;
    }
  }
}

.info-row {
  display: flex;
  align-items: center;
  padding: 12rpx 0;
  border-bottom: 1rpx solid #f7f8fa;

  &:last-child {
    border-bottom: none;
  }

  .info-label {
    font-size: 26rpx;
    color: #969799;
    width: 160rpx;
    flex-shrink: 0;
  }

  .info-value {
    flex: 1;
    font-size: 26rpx;
    color: #323233;

    &.status-1 { color: #ff976a; }
    &.status-2 { color: #1989fa; }
    &.status-3 { color: #7232dd; }
    &.status-4,
    &.status-5 { color: #07c160; }
    &.status--1 { color: #969799; }
  }

  .copy-btn {
    font-size: 24rpx;
    color: #1989fa;
    padding: 4rpx 16rpx;
    border: 1rpx solid #1989fa;
    border-radius: 8rpx;
  }
}

.item-card {
  padding: 20rpx 0;
  border-bottom: 1rpx solid #f2f3f5;

  &:last-child {
    border-bottom: none;
  }
}

.item-main {
  display: flex;
}

.item-image {
  width: 120rpx;
  height: 120rpx;
  border-radius: 12rpx;
  background: #f7f8fa;
  flex-shrink: 0;
}

.item-info {
  flex: 1;
  margin-left: 20rpx;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
}

.item-name {
  font-size: 28rpx;
  color: #323233;
  font-weight: bold;
}

.item-spec {
  font-size: 22rpx;
  color: #969799;
  margin-top: 4rpx;
}

.item-price-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: 8rpx;

  .item-price {
    font-size: 28rpx;
    color: #ee0a24;
    font-weight: bold;
  }

  .item-qty {
    font-size: 24rpx;
    color: #646566;
  }
}

.item-status-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: 16rpx;
  padding-top: 16rpx;
  border-top: 1rpx dashed #ebedf0;
}

.item-cook-status {
  font-size: 24rpx;
  padding: 8rpx 20rpx;
  border-radius: 8rpx;

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

.item-actions {
  display: flex;
  gap: 16rpx;

  .action-btn {
    font-size: 24rpx;
    padding: 8rpx 20rpx;
    border-radius: 8rpx;

    &.serve-btn {
      background: #07c160;
      color: #fff;
    }

    &.cooking-btn {
      background: #1989fa;
      color: #fff;
    }

    &.refund-btn {
      background: #fff;
      color: #ee0a24;
      border: 1rpx solid #ee0a24;
    }
  }
}

.amount-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12rpx 0;

  &.total {
    padding-top: 20rpx;
    margin-top: 12rpx;
    border-top: 1rpx dashed #ebedf0;

    .amount-label,
    .amount-value {
      font-size: 30rpx;
      font-weight: bold;
    }

    .amount-value {
      color: #ee0a24;
    }
  }

  .amount-label {
    font-size: 26rpx;
    color: #646566;
  }

  .amount-value {
    font-size: 26rpx;
    color: #323233;

    &.discount {
      color: #07c160;
    }
  }
}

.footer-actions {
  display: flex;
  gap: 16rpx;
  padding: 20rpx 24rpx;
  padding-bottom: calc(20rpx + env(safe-area-inset-bottom));
  background: #fff;
  box-shadow: 0 -4rpx 12rpx rgba(0, 0, 0, 0.04);
  position: fixed;
  left: 0;
  right: 0;
  bottom: 0;

  .footer-btn {
    flex: 1;
    text-align: center;
    padding: 20rpx 0;
    border-radius: 44rpx;
    font-size: 28rpx;
    font-weight: bold;

    &.btn-refresh {
      background: #f7f8fa;
      color: #323233;
    }

    &.btn-add {
      background: #ff976a;
      color: #fff;
    }

    &.btn-complete {
      background: linear-gradient(135deg, #1989fa 0%, #07c160 100%);
      color: #fff;
    }
  }
}
</style>
