import React, { useState, useEffect } from 'react'
import { View, Text, Image, ScrollView } from '@tarojs/components'
import Taro, { useRouter } from '@tarojs/taro'
import { Loading, Dialog, Button, Cell } from '@nutui/nutui-react-taro'
import { getOrderDetail, cancelOrder, refundOrder, getPaymentParams } from '../../services/order'
import type { Order } from '../../services/order'
import { orderTypeMap } from '../../services/delivery'
import DeliveryProgress from '../../components/DeliveryProgress'
import styles from './detail.module.scss'

const statusMap: Record<number, { text: string; color: string }> = {
  0: { text: '待支付', color: '#f56c6c' },
  1: { text: '待接单', color: '#e6a23c' },
  2: { text: '制作中', color: '#409eff' },
  3: { text: '已完成', color: '#67c23a' },
  4: { text: '已取消', color: '#909399' },
  5: { text: '退款中', color: '#f56c6c' },
  6: { text: '已退款', color: '#909399' }
}

const OrderDetail: React.FC = () => {
  const router = useRouter()
  const orderNo = router.params.order_no as string

  const [order, setOrder] = useState<Order | null>(null)
  const [loading, setLoading] = useState(true)
  const [showCancelDialog, setShowCancelDialog] = useState(false)
  const [showRefundDialog, setShowRefundDialog] = useState(false)
  const [refundReason, setRefundReason] = useState('')
  const [payLoading, setPayLoading] = useState(false)

  useEffect(() => {
    loadOrder()
  }, [orderNo])

  const loadOrder = async () => {
    setLoading(true)
    try {
      const data = await getOrderDetail(orderNo)
      setOrder(data)
    } catch (e: any) {
      Taro.showToast({ title: e.message || '加载失败', icon: 'none' })
    } finally {
      setLoading(false)
    }
  }

  const getStatusSteps = () => {
    if (!order) return []
    const steps = [
      { title: '提交订单', done: order.status >= 0 },
      { title: '支付成功', done: order.status >= 1 || order.pay_status === 1 },
      { title: '商家接单', done: order.status >= 2 },
      { title: '制作完成', done: order.status >= 3 }
    ]
    return steps
  }

  const handlePay = async () => {
    if (!order) return
    setPayLoading(true)
    try {
      const params = await getPaymentParams(order.order_no)
      
      await Taro.requestPayment({
        timeStamp: params.timeStamp,
        nonceStr: params.nonceStr,
        package: params.package,
        signType: params.signType || 'MD5',
        paySign: params.paySign
      })

      Taro.showToast({ title: '支付成功', icon: 'success' })
      loadOrder()
    } catch (e: any) {
      if (e.errMsg?.includes('cancel')) {
        Taro.showToast({ title: '已取消支付', icon: 'none' })
      } else {
        Taro.showToast({ title: '支付失败，请重新支付', icon: 'none' })
      }
    } finally {
      setPayLoading(false)
    }
  }

  const handleCancel = async () => {
    try {
      await cancelOrder(orderNo, '用户取消')
      setShowCancelDialog(false)
      Taro.showToast({ title: '已取消', icon: 'success' })
      loadOrder()
    } catch (e: any) {
      Taro.showToast({ title: e.message || '取消失败', icon: 'none' })
    }
  }

  const handleRefund = async () => {
    if (!refundReason.trim()) {
      Taro.showToast({ title: '请填写退款原因', icon: 'none' })
      return
    }
    try {
      await refundOrder(orderNo, refundReason.trim())
      setShowRefundDialog(false)
      Taro.showToast({ title: '退款申请已提交', icon: 'success' })
      loadOrder()
    } catch (e: any) {
      Taro.showToast({ title: e.message || '申请失败', icon: 'none' })
    }
  }

  if (loading) {
    return (
      <View className={styles.loadingPage}>
        <Loading type='spinner' />
        <Text className={styles.loadingText}>加载中...</Text>
      </View>
    )
  }

  if (!order) {
    return (
      <View className={styles.empty}>
        <Text className={styles.emptyText}>订单不存在</Text>
      </View>
    )
  }

  const statusInfo = statusMap[order.status] || { text: '未知状态', color: '#999' }
  const orderTypeLabel = (orderTypeMap as any)[order.order_type || 'dine_in']?.label || '堂食'

  return (
    <View className={styles.container}>
      <ScrollView scrollY className={styles.scrollContent}>
        <View className={styles.statusCard}>
          <View className={styles.statusIcon}>
            <Text className={styles.statusEmoji}>
              {order.status === 3 ? '✅' : order.status === 0 ? '💳' : order.status === 4 ? '❌' : '🍳'}
            </Text>
          </View>
          <View className={styles.statusInfo}>
            <Text className={styles.statusText} style={{ color: statusInfo.color }}>
              {statusInfo.text}
            </Text>
            {order.order_type && (
              <Text className={styles.orderTypeTag}>{orderTypeLabel}</Text>
            )}
            {order.status === 2 && (
              <Text className={styles.statusDesc}>商家正在为您准备餐品，请稍候</Text>
            )}
            {order.status === 3 && order.paid_at && (
              <Text className={styles.statusDesc}>完成时间：{order.paid_at}</Text>
            )}
          </View>
        </View>

        {(order.order_type === 'delivery' || order.order_type === 'pickup' || order.order_type === 'takeout') && order.id && (
          <DeliveryProgress
            orderId={order.id}
            orderType={order.order_type || 'dine_in'}
            pickupCode={(order as any).pickup_code}
            orderStatus={order.status}
          />
        )}

        {order.order_type === 'dine_in' && order.status >= 1 && order.status <= 3 && (
          <View className={styles.section}>
            <View className={styles.sectionHeader}>
              <Text className={styles.sectionTitle}>订单进度</Text>
            </View>
            <View className={styles.steps}>
              {getStatusSteps().map((step, idx) => (
                <View key={idx} className={`${styles.step} ${step.done ? styles.stepDone : ''}`}>
                  <View className={styles.stepDot} />
                  {idx < getStatusSteps().length - 1 && <View className={styles.stepLine} />}
                  <Text className={styles.stepTitle}>{step.title}</Text>
                </View>
              ))}
            </View>
          </View>
        )}

        <View className={styles.section}>
          <View className={styles.sectionHeader}>
            <Text className={styles.sectionIcon}>📍</Text>
            <Text className={styles.sectionTitle}>门店信息</Text>
          </View>
          <Cell title='门店' description={order.store_id?.toString() || '门店名称'} />
          {order.table_no && (
            <Cell title='桌号' description={order.table_no} />
          )}
          {(order as any).delivery_address && (
            <Cell title='配送地址' description={(order as any).delivery_address} />
          )}
          {(order as any).delivery_contact && (
            <Cell title='联系人' description={`${(order as any).delivery_contact} ${(order as any).delivery_phone || ''}`} />
          )}
        </View>

        <View className={styles.section}>
          <View className={styles.sectionHeader}>
            <Text className={styles.sectionIcon}>🍜</Text>
            <Text className={styles.sectionTitle}>商品清单</Text>
          </View>
          {order.items.map(item => (
            <View key={item.id} className={styles.orderItem}>
              <View className={styles.itemInfo}>
                <Text className={styles.itemName}>{item.product_name}</Text>
                <Text className={styles.itemSpec}>
                  {item.sku_name}
                  {item.attribute_names.map((name, idx) => (
                    <Text key={idx}> · {name}</Text>
                  ))}
                </Text>
                {item.remark && (
                  <Text className={styles.itemRemark}>备注：{item.remark}</Text>
                )}
              </View>
              <View className={styles.itemRight}>
                <Text className={styles.itemPrice}>¥{item.price.toFixed(2)}</Text>
                <Text className={styles.itemQty}>×{item.quantity}</Text>
                <Text className={styles.itemSubtotal}>¥{item.subtotal.toFixed(2)}</Text>
              </View>
            </View>
          ))}
        </View>

        <View className={styles.section}>
          <View className={styles.sectionHeader}>
            <Text className={styles.sectionIcon}>📝</Text>
            <Text className={styles.sectionTitle}>订单信息</Text>
          </View>
          <Cell title='订单号' description={order.order_no} />
          <Cell title='下单时间' description={order.created_at} />
          {order.paid_at && (
            <Cell title='支付时间' description={order.paid_at} />
          )}
          {order.pay_method && (
            <Cell title='支付方式' description={order.pay_method} />
          )}
          {order.remark && (
            <Cell title='备注' description={order.remark} />
          )}
        </View>

        <View className={styles.section}>
          <View className={styles.sectionHeader}>
            <Text className={styles.sectionIcon}>💰</Text>
            <Text className={styles.sectionTitle}>费用明细</Text>
          </View>
          <Cell title='商品合计' extra={`¥${order.total_amount.toFixed(2)}`} />
          {order.discount_amount > 0 && (
            <Cell title='优惠折扣' extra={`-¥${order.discount_amount.toFixed(2)}`} />
          )}
          {(order as any).delivery_fee > 0 && (
            <Cell title='配送费' extra={`¥${Number((order as any).delivery_fee).toFixed(2)}`} />
          )}
          <View className={styles.totalRow}>
            <Text className={styles.totalLabel}>实付金额</Text>
            <View className={styles.totalPrice}>
              <Text className={styles.totalSymbol}>¥</Text>
              <Text className={styles.totalValue}>{order.actual_amount.toFixed(2)}</Text>
            </View>
          </View>
        </View>
      </ScrollView>

      {(order.status === 0 || order.status === 1 || order.status === 3) && (
        <View className={styles.footer}>
          {order.status === 0 && (
            <>
              <View
                className={styles.footerBtnSecondary}
                onClick={() => setShowCancelDialog(true)}
              >
                <Text className={styles.btnSecondaryText}>取消订单</Text>
              </View>
              <View
                className={`${styles.footerBtnPrimary} ${payLoading ? styles.btnDisabled : ''}`}
                onClick={handlePay}
              >
                {payLoading ? (
                  <Loading type='spinner' size='16px' color='#fff' />
                ) : (
                  <Text className={styles.btnPrimaryText}>立即支付</Text>
                )}
              </View>
            </>
          )}
          {order.status === 1 && (
            <View
              className={styles.footerBtnSecondary}
              onClick={() => setShowCancelDialog(true)}
            >
              <Text className={styles.btnSecondaryText}>取消订单</Text>
            </View>
          )}
          {order.status === 3 && (
            <View
              className={styles.footerBtnPrimary}
              onClick={() => setShowRefundDialog(true)}
            >
              <Text className={styles.btnPrimaryText}>申请退款</Text>
            </View>
          )}
        </View>
      )}

      <Dialog
        visible={showCancelDialog}
        title='确认取消'
        content='确定要取消该订单吗？'
        okText='确定取消'
        cancelText='再想想'
        onOk={handleCancel}
        onCancel={() => setShowCancelDialog(false)}
      />

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
              onClick={handleRefund}
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

export default OrderDetail
