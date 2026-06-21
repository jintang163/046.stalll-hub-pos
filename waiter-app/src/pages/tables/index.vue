<template>
  <view class="tables-page">
    <view class="stats-bar">
      <view class="stat-item" v-for="stat in statsList" :key="stat.key">
        <view class="stat-value" :class="stat.color">{{ stat.value }}</view>
        <view class="stat-label">{{ stat.label }}</view>
      </view>
    </view>

    <view class="filter-bar">
      <scroll-view scroll-x class="filter-scroll">
        <view class="filter-item" 
          :class="{ active: currentStatus === status.key }"
          v-for="status in statusFilters"
          :key="status.key"
          @click="filterByStatus(status.key)">
          {{ status.label }}
        </view>
      </scroll-view>
    </view>

    <scroll-view scroll-y class="tables-scroll" @refresherrefresh="onRefresh" :refresher-triggered="refreshing">
      <view class="tables-grid">
        <view 
          class="table-card" 
          v-for="table in filteredTables" 
          :key="table.id"
          :class="`table-${table.display_status}`"
          @click="onTableClick(table)">
          <view class="table-header">
            <view class="table-no">{{ table.table_no }}</view>
            <view class="table-status" :class="`status-${table.display_status}`">
              {{ getStatusText(table.display_status) }}
            </view>
          </view>
          <view class="table-info">
            <view class="info-row" v-if="table.current_customer_count > 0">
              <text class="info-icon">👥</text>
              <text>{{ table.current_customer_count }}人</text>
            </view>
            <view class="info-row" v-if="table.order_no">
              <text class="info-icon">📋</text>
              <text>{{ table.order_no }}</text>
            </view>
            <view class="info-row" v-if="table.item_count > 0">
              <text class="info-icon">🍽️</text>
              <text>{{ table.served_count }}/{{ table.item_count }} 道</text>
            </view>
            <view class="info-row" v-if="table.order_amount > 0">
              <text class="info-icon">💰</text>
              <text>¥{{ table.order_amount.toFixed(2) }}</text>
            </view>
          </view>
          <view class="table-footer" v-if="table.display_status === 'ordered' || table.display_status === 'all_served'">
            <view class="footer-btn btn-order" @click.stop="viewOrder(table)">查看订单</view>
            <view class="footer-btn btn-add" @click.stop="addDish(table)">加菜</view>
          </view>
          <view class="table-footer" v-else-if="table.display_status === 'occupied'">
            <view class="footer-btn btn-order" @click.stop="orderNow(table)">点餐</view>
          </view>
        </view>
      </view>

      <view class="empty" v-if="filteredTables.length === 0 && !loading">
        暂无桌位数据
      </view>
    </scroll-view>
  </view>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import { useUserStore } from '../../store/user'
import { useCartStore } from '../../store/cart'
import { waiterApi } from '../../services/waiter'
import type { TableInfo, WaiterStats } from '../../types'

const userStore = useUserStore()
const cartStore = useCartStore()

const tables = ref<TableInfo[]>([])
const stats = ref<WaiterStats>({
  total_tables: 0,
  idle_tables: 0,
  occupied_tables: 0,
  ordered_tables: 0,
  pending_calls: 0,
  pending_orders: 0
})
const currentStatus = ref<string>('all')
const loading = ref(false)
const refreshing = ref(false)

const statusFilters = [
  { key: 'all', label: '全部' },
  { key: 'idle', label: '空闲' },
  { key: 'occupied', label: '已入座' },
  { key: 'ordered', label: '已下单' },
  { key: 'all_served', label: '已上菜' },
  { key: 'paid', label: '已结账' }
]

const statsList = computed(() => [
  { key: 'total', label: '总桌数', value: stats.value.total_tables, color: 'text-primary' },
  { key: 'idle', label: '空闲', value: stats.value.idle_tables, color: 'text-success' },
  { key: 'occupied', label: '入座', value: stats.value.occupied_tables, color: 'text-warning' },
  { key: 'ordered', label: '下单', value: stats.value.ordered_tables, color: 'text-primary' }
])

const filteredTables = computed(() => {
  if (currentStatus.value === 'all') {
    return tables.value
  }
  return tables.value.filter(t => t.display_status === currentStatus.value)
})

const getStatusText = (status: string) => {
  const map: Record<string, string> = {
    idle: '空闲',
    occupied: '已入座',
    ordered: '已下单',
    all_served: '已上菜',
    paid: '已结账'
  }
  return map[status] || status
}

const filterByStatus = (status: string) => {
  currentStatus.value = status
}

const loadData = async () => {
  if (!userStore.userInfo?.store_id) return
  loading.value = true
  try {
    const [tablesData, statsData] = await Promise.all([
      waiterApi.getTables(userStore.userInfo.store_id),
      waiterApi.getStats(userStore.userInfo.store_id)
    ])
    tables.value = tablesData
    stats.value = statsData
  } catch (e: any) {
    console.error('Load tables failed:', e)
  } finally {
    loading.value = false
    refreshing.value = false
  }
}

const onRefresh = () => {
  refreshing.value = true
  loadData()
}

const onTableClick = (table: TableInfo) => {
  if (table.display_status === 'idle') {
    uni.showActionSheet({
      itemList: ['客人入座'],
      success: () => {
        uni.navigateTo({
          url: `/pages/menu/index?tableId=${table.id}&tableNo=${table.table_no}&action=checkin`
        })
      }
    })
  } else if (table.current_order_id > 0) {
    uni.navigateTo({
      url: `/pages/orders/detail?id=${table.current_order_id}`
    })
  }
}

const viewOrder = (table: TableInfo) => {
  if (table.current_order_id > 0) {
    uni.navigateTo({
      url: `/pages/orders/detail?id=${table.current_order_id}`
    })
  }
}

const addDish = (table: TableInfo) => {
  cartStore.clearCart()
  cartStore.setTable(table.id, table.table_no)
  uni.navigateTo({
    url: `/pages/menu/index?tableId=${table.id}&tableNo=${table.table_no}&orderId=${table.current_order_id}&action=add`
  })
}

const orderNow = (table: TableInfo) => {
  cartStore.clearCart()
  cartStore.setTable(table.id, table.table_no)
  uni.navigateTo({
    url: `/pages/menu/index?tableId=${table.id}&tableNo=${table.table_no}&action=new`
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
    loadData()
  }
})
</script>

<style lang="scss" scoped>
.tables-page {
  min-height: 100vh;
  background: #f5f5f5;
  display: flex;
  flex-direction: column;
}

.stats-bar {
  display: flex;
  background: #fff;
  padding: 24rpx 0;
  box-shadow: 0 2rpx 8rpx rgba(0, 0, 0, 0.04);

  .stat-item {
    flex: 1;
    text-align: center;

    .stat-value {
      font-size: 44rpx;
      font-weight: bold;
      margin-bottom: 8rpx;
    }

    .stat-label {
      font-size: 24rpx;
      color: #969799;
    }
  }
}

.filter-bar {
  background: #fff;
  border-bottom: 1rpx solid #ebedf0;
  padding: 16rpx 0;

  .filter-scroll {
    white-space: nowrap;
    padding: 0 24rpx;
  }

  .filter-item {
    display: inline-block;
    padding: 12rpx 32rpx;
    margin-right: 16rpx;
    background: #f7f8fa;
    border-radius: 32rpx;
    font-size: 26rpx;
    color: #646566;

    &.active {
      background: #1989fa;
      color: #fff;
    }
  }
}

.tables-scroll {
  flex: 1;
  height: 0;
}

.tables-grid {
  display: flex;
  flex-wrap: wrap;
  padding: 16rpx;
}

.table-card {
  width: calc(50% - 16rpx);
  margin: 8rpx;
  background: #fff;
  border-radius: 16rpx;
  padding: 24rpx;
  box-shadow: 0 2rpx 8rpx rgba(0, 0, 0, 0.04);
  border-left: 8rpx solid #ebedf0;
  box-sizing: border-box;

  &.table-idle {
    border-left-color: #07c160;
  }

  &.table-occupied {
    border-left-color: #ff976a;
  }

  &.table-ordered {
    border-left-color: #1989fa;
  }

  &.table-all_served {
    border-left-color: #7232dd;
  }

  &.table-paid {
    border-left-color: #969799;
  }
}

.table-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16rpx;

  .table-no {
    font-size: 32rpx;
    font-weight: bold;
    color: #323233;
  }

  .table-status {
    font-size: 22rpx;
    padding: 4rpx 12rpx;
    border-radius: 8rpx;

    &.status-idle {
      background: #e8f5e9;
      color: #07c160;
    }

    &.status-occupied {
      background: #fff7e6;
      color: #ff976a;
    }

    &.status-ordered {
      background: #e6f7ff;
      color: #1989fa;
    }

    &.status-all_served {
      background: #f3eaff;
      color: #7232dd;
    }

    &.status-paid {
      background: #f0f0f0;
      color: #969799;
    }
  }
}

.table-info {
  .info-row {
    display: flex;
    align-items: center;
    font-size: 24rpx;
    color: #646566;
    margin-bottom: 8rpx;

    .info-icon {
      margin-right: 8rpx;
      font-size: 24rpx;
    }
  }
}

.table-footer {
  display: flex;
  gap: 12rpx;
  margin-top: 16rpx;
  padding-top: 16rpx;
  border-top: 1rpx solid #f2f3f5;

  .footer-btn {
    flex: 1;
    text-align: center;
    padding: 12rpx 0;
    border-radius: 8rpx;
    font-size: 24rpx;

    &.btn-order {
      background: #e6f7ff;
      color: #1989fa;
    }

    &.btn-add {
      background: #e8f5e9;
      color: #07c160;
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
