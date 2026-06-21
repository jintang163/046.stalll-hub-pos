import React, { useState, useEffect } from 'react'
import { View, Text, Image, ScrollView } from '@tarojs/components'
import Taro, { useDidShow } from '@tarojs/taro'
import { Tabs, TabPane, Loading, Empty, PullRefresh, Dialog } from '@nutui/nutui-react-taro'
import { getOrders, cancelOrder, refundOrder } from '../../services/order'
import type { Order } from '../../services/order'
import { isLogin } from '../../services/auth'
import styles from './list.module.scss'

const statusMap: Record<number, { text: string; color: string }> = {
  0: { text: '待支付', color: '#f56c6c' },
  1: { text: '待接单', color: '#e6a23c' },
  2: { text: '制作中', color: '#409eff' },
  3: { text: '已完成', color: '#67c23a' },
  4: { text: '已取消', color: '#909399' },
  5: { text: '退款中', color: '#f56c6c' },
  6: { text: '已退款', color: '#909399' }
}

const reservationStatusMap: Record<number, { text: string; color: string }> = {
  0: { text: '待确认', color: '#e6a23c' },
  1: { text: '已确认', color: '#67c23a' },
  2: { text: '已取消', color: '#909399' },
  3: { text: '已完成', color: '#67c23a' }
}

const statusTabList = [
  { value: -1, title: '全部' },
  { value: 0, title: '待支付' },
  { value: 1, title: '待接单' },
  { value: 2, title: '制作中' },
  { value: 3, title: '已完成' }
]

const orderTypeTabList = [
  { value: 'all', title: '全部' },
  { value: 'instant', title: '即时订单' },
  { value: 'reservation', title: '预约订单' }
]

const OrderList: React.FC = () => {
  const [statusTabValue, setStatusTabValue] = useState(-1)
  const [orderTypeTabValue, setOrderTypeTabValue] = useState<'all' | 'instant' | 'reservation'>('all')
  const [orders, setOrders] = useState<Order[]>([])
  const [loading, setLoading] = useState(false)
  const [refreshing, setRefreshing] = useState(false)
  const [page, setPage] = useState(1)
  const [hasMore, setHasMore] = useState(true)
  const [showRefundDialog, setShowRefundDialog] = useState(false)
  const [refundOrderNo, setRefundOrderNo] = useState('')
  const [refundReason, setRefundReason] = useState('')

  useDidShow(() => {
    if (isLogin()) {
      loadOrders(1, true)
    } else {
      Taro.navigateTo({ url: '/pages/user/index' })
    }
  })

  const loadOrders = async (pageNum = 1, isRefresh = false) => {
    setLoading(true)
    if (isRefresh) setRefreshing(true)
    try {
      const status = statusTabValue === -1 ? undefined : statusTabValue
      const result = await getOrders(status, pageNum, 10, orderTypeTabValue)
      
      if (pageNum === 1) {
        setOrders(result.list)
      } else {
        setOrders(prev => [...prev, ...result.list])
      }
      setHasMore(result.list.length === 10)
      setPage(pageNum)
    } catch (e: any) {
      if (e.message?.includes('401')) {
        Taro.removeStorageSync('token')
        Taro.navigateTo({ url: '/pages/user/index' })
      }
    } finally {
      setLoading(false)
      setRefreshing(false)
    }
  }

  const handleStatusTabChange = (value: number) => {
    setStatusTabValue(value)
    loadOrders(1, true)
  }

  const handleOrderTypeTabChange = (value: string) => {
    setOrderTypeTabValue(value as 'all' | 'instant' | 'reservation')
    loadOrders(1, true)
  }

  const handleRefresh = async () => {
    await loadOrders(1, true)
  }

  const handleLoadMore = () => {
    if (hasMore && !loading) {
      loadOrders(page + 1)
    }
  }

  const handleOrderClick = (order: Order) => {
    Taro.navigateTo({ url: `/pages/order/detail?order_no=${order.order_no}` })
  }

  const handlePay = (order: Order) => {
    Taro.navigateTo({ url: `/pages/order/detail?order_no=${order.order_no}` })
  }

  const handleCancel = (order: Order) => {
    Dialog.show({
      title: '确认取消',
      content: '确定要取消该订单吗？',
      okText: '确定取消',
      cancelText: '再想想',
      onOk: async () => {
        try {
          await cancelOrder(order.order_no, '用户取消')
          Taro.showToast({ title: '已取消', icon: 'success' })
          loadOrders(1, true)
        } catch (e: any) {
          Taro.showToast({ title: e.message || '取消失败', icon: 'none' })
        }
      }
    })
  }

  const handleRefund = (order: Order) => {
    setRefundOrderNo(order.order_no)
    setRefundReason('')
    setShowRefundDialog(true)
  }

  const confirmRefund = async () => {
    if (!refundReason.trim()) {
      Taro.showToast({ title: '请填写退款原因', icon: 'none' })
      return
    }
    try {
      await refundOrder(refundOrderNo, refundReason.trim())
      setShowRefundDialog(false)
      Taro.showToast({ title: '退款申请已提交', icon: 'success' })
      loadOrders(1, true)
    } catch (e: any) {
      Taro.showToast({ title: e.message || '申请失败', icon: 'none' })
    }
  }

  if (!isLogin()) {
    return (
      <View className={styles.empty}>
        <Text className={styles.emptyText}>请先登录</Text>
      </View>
    )
  }

  return (
    <View className={styles.container}>
      <Tabs
        value={orderTypeTabValue}
        onChange={handleOrderTypeTabChange}
        tabTitleStyle={{ fontSize: '26rpx' }}
        className={styles.orderTypeTabs}
      >
        {orderTypeTabList.map(tab => (
          <TabPane key={tab.value} title={tab.title} value={tab.value} />
        ))}
      </Tabs>

      <Tabs
        value={statusTabValue}
        onChange={handleStatusTabChange}
        tabTitleStyle={{ fontSize: '28rpx' }}
      >
        {statusTabList.map(tab => (
          <TabPane key={tab.value} title={tab.title} value={tab.value}>
            <PullRefresh
              onRefresh={handleRefresh}
              refreshing={refreshing}
            >
              <ScrollView
                scrollY
                onScrollToLower={handleLoadMore}
                className={styles.orderList}
              >
                {orders.length === 0 && !loading ? (
                  <Empty description='暂无订单' />
                ) : (
                  orders.map(order => (
                    <View
                      key={order.id}
                      className={styles.orderCard}
                      onClick={() => handleOrderClick(order)}
                    >
                      <View className={styles.orderHeader}>
                        <View className={styles.orderHeaderLeft}>
                          <Text className={styles.orderNo}>订单号：{order.order_no}</Text>
                          {order.is_reservation && (
                            <View className={styles.reservationTag}>
                              <Text className={styles.reservationTagText}>预约</Text>
                            </View>
                          )}
                        </View>
                        <Text
                          className={styles.orderStatus}
                          style={{ color: statusMap[order.status]?.color }}
                        >
                          {statusMap[order.status]?.text}
                        </Text>
                      </View>

                      {order.is_reservation && order.reservation_status !== undefined && (
                        <View className={styles.orderInfo}>
                          <Text className={styles.infoLabel}>预约状态：</Text>
                          <Text
                            className={styles.infoValue}
                            style={{ color: reservationStatusMap[order.reservation_status]?.color }}
                          >
                            {reservationStatusMap[order.reservation_status]?.text}
                          </Text>
                        </View>
                      )}

                      {order.is_reservation && order.reservation_time && (
                        <View className={styles.orderInfo}>
                          <Text className={styles.infoLabel}>预约时间：</Text>
                          <Text className={styles.infoValue}>{order.reservation_time}</Text>
                        </View>
                      )}

                      {order.is_reservation && order.time_slot_name && (
                        <View className={styles.orderInfo}>
                          <Text className={styles.infoLabel}>时段：</Text>
                          <Text className={styles.infoValue}>{order.time_slot_name}</Text>
                        </View>
                      )}

                      {order.time_slot_discount && (
                        <View className={styles.discountInfo}>
                          <Text className={styles.discountIcon}>🎉</Text>
                          <Text className={styles.discountText}>{order.time_slot_discount}</Text>
                        </View>
                      )}

                      {order.table_no && (
                        <View className={styles.orderInfo}>
                          <Text className={styles.infoLabel}>桌号：</Text>
                          <Text className={styles.infoValue}>{order.table_no}</Text>
                        </View>
                      )}

                      <View className={styles.orderItems}>
                        {order.items.slice(0, 2).map(item => (
                          <View key={item.id} className={styles.orderItem}>
                            <Text className={styles.itemName}>
                              {item.product_name} · {item.sku_name}
                            </Text>
                            <Text className={styles.itemQty}>×{item.quantity}</Text>
                            <Text className={styles.itemPrice}>¥{item.price.toFixed(2)}</Text>
                          </View>
                        ))}
                        {order.items.length > 2 && (
                          <Text className={styles.moreItems}>共{order.items.length}件商品</Text>
                        )}
                      </View>

                      <View className={styles.orderFooter}>
                        <View className={styles.orderTotal}>
                          <Text className={styles.totalLabel}>实付：</Text>
                          <Text className={styles.totalSymbol}>¥</Text>
                          <Text className={styles.totalValue}>{order.actual_amount.toFixed(2)}</Text>
                        </View>
                        <View className={styles.orderActions}>
                          {order.status === 0 && (
                            <>
                              <View
                                className={styles.actionBtnSecondary}
                                onClick={(e) => { e.stopPropagation(); handleCancel(order); }}
                              >
                                <Text className={styles.actionTextSecondary}>取消订单</Text>
                              </View>
                              <View
                                className={styles.actionBtnPrimary}
                                onClick={(e) => { e.stopPropagation(); handlePay(order); }}
                              >
                                <Text className={styles.actionTextPrimary}>去支付</Text>
                              </View>
                            </>
                          )}
                          {order.status === 1 && (
                            <View
                              className={styles.actionBtnSecondary}
                              onClick={(e) => { e.stopPropagation(); handleCancel(order); }}
                            >
                              <Text className={styles.actionTextSecondary}>取消订单</Text>
                            </View>
                          )}
                          {order.status === 3 && (
                            <View
                              className={styles.actionBtnSecondary}
                              onClick={(e) => { e.stopPropagation(); handleRefund(order); }}
                            >
                              <Text className={styles.actionTextSecondary}>申请退款</Text>
                            </View>
                          )}
                          <View
                            className={styles.actionBtnSecondary}
                            onClick={(e) => { e.stopPropagation(); handleOrderClick(order); }}
                          >
                            <Text className={styles.actionTextSecondary}>查看详情</Text>
                          </View>
                        </View>
                      </View>
                    </View>
                  ))
                )}

                {loading && orders.length > 0 && (
                  <View className={styles.loadingMore}>
                    <Loading type='spinner' size='16px' />
                    <Text className={styles.loadingText}>加载中...</Text>
                  </View>
                )}

                {!hasMore && orders.length > 0 && (
                  <View className={styles.noMore}>
                    <Text className={styles.noMoreText}>— 没有更多了 —</Text>
                  </View>
                )}
              </ScrollView>
            </PullRefresh>
          </TabPane>
        ))}
      </Tabs>

      <Dialog
        visible={showRefundDialog}
        title='申请退款'
        footer={
          <View className={styles.dialogFooter}>
            <View
              className={styles.dialogBtn}
              onClick={() => setShowRefundDialog(false)}
            >
              <Text>取消</Text>
            </View>
            <View
              className={styles.dialogBtnPrimary}
              onClick={confirmRefund}
            >
              <Text style={{ color: '#fff' }}>提交</Text>
            </View>
          </View>
        }
        onClose={() => setShowRefundDialog(false)}
      >
        <View className={styles.refundForm}>
          <Text className={styles.formLabel}>退款原因</Text>
          <textarea
            className={styles.formTextarea}
            placeholder='请填写退款原因'
            value={refundReason}
            onInput={(e: any) => setRefundReason(e.target.value)}
          />
        </View>
      </Dialog>
    </View>
  )
}

export default OrderList
