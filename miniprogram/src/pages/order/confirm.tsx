import React, { useState, useEffect } from 'react'
import { View, Text, Image, ScrollView } from '@tarojs/components'
import Taro from '@tarojs/taro'
import { Loading, Dialog, Cell, Button } from '@nutui/nutui-react-taro'
import { useCartStore } from '../../store/cart'
import { useAppStore } from '../../store/app'
import { getAvailableCoupons, MemberCoupon } from '../../services/coupon'
import { createOrder, getPaymentParams } from '../../services/order'
import type { OrderCreateDTO, OrderItem } from '../../services/order'
import { isLogin, loginByCode } from '../../services/auth'
import styles from './confirm.module.scss'

const OrderConfirm: React.FC = () => {
  const items = useCartStore(state => state.items)
  const total = useCartStore(state => state.total())
  const couponId = useCartStore(state => state.couponId)
  const couponDiscount = useCartStore(state => state.couponDiscount)
  const tableNo = useCartStore(state => state.tableNo)
  const remark = useCartStore(state => state.remark)
  const setCoupon = useCartStore(state => state.setCoupon)
  const clear = useCartStore(state => state.clear)
  const currentStore = useAppStore(state => state.currentStore)
  const user = useAppStore(state => state.user)

  const [loading, setLoading] = useState(false)
  const [submitting, setSubmitting] = useState(false)
  const [coupons, setCoupons] = useState<MemberCoupon[]>([])
  const [showCouponDialog, setShowCouponDialog] = useState(false)
  const [showPayDialog, setShowPayDialog] = useState(false)
  const [payLoading, setPayLoading] = useState(false)
  const [createdOrder, setCreatedOrder] = useState<any>(null)

  const actualTotal = Math.max(0, total - couponDiscount)

  useEffect(() => {
    loadCoupons()
  }, [total])

  const loadCoupons = async () => {
    if (!isLogin()) return
    setLoading(true)
    try {
      const list = await getAvailableCoupons(total)
      setCoupons(list)
    } catch {}
    finally {
      setLoading(false)
    }
  }

  const handleSelectCoupon = (coupon: MemberCoupon | null) => {
    if (!coupon) {
      setCoupon(null, 0)
    } else {
      const discount = coupon.coupon.type === 1 
        ? coupon.coupon.value 
        : (total * coupon.coupon.value / 100)
      setCoupon(coupon.id, Math.min(discount, total))
    }
    setShowCouponDialog(false)
  }

  const handleSubmit = async () => {
    if (!currentStore) {
      Taro.showToast({ title: '请先选择门店', icon: 'none' })
      return
    }

    if (items.length === 0) {
      Taro.showToast({ title: '购物车为空', icon: 'none' })
      return
    }

    if (!tableNo) {
      Taro.showToast({ title: '请输入桌号', icon: 'none' })
      return
    }

    if (!isLogin()) {
      await handleWxLogin()
      if (!isLogin()) return
    }

    setSubmitting(true)
    try {
      const orderItems: OrderItem[] = items.map(item => ({
        product_id: item.product_id,
        product_name: item.product_name,
        sku_id: item.sku_id,
        sku_name: item.sku_name,
        attribute_ids: item.attribute_ids,
        attribute_names: item.attribute_names,
        price: item.price,
        quantity: item.quantity,
        subtotal: item.subtotal,
        remark: remark
      }))

      const orderData: OrderCreateDTO = {
        store_id: currentStore.id,
        items: orderItems,
        table_no: tableNo,
        remark: remark,
        coupon_id: couponId || undefined
      }

      const order = await createOrder(orderData)
      setCreatedOrder(order)
      setShowPayDialog(true)
    } catch (e: any) {
      Taro.showToast({ title: e.message || '下单失败', icon: 'none' })
    } finally {
      setSubmitting(false)
    }
  }

  const handleWxLogin = async () => {
    try {
      const res = await Taro.login()
      if (res.code) {
        const result = await loginByCode(res.code)
        Taro.setStorageSync('token', result.token)
        Taro.setStorageSync('userInfo', result.user)
        useAppStore.getState().setUser(result.user)
      }
    } catch (e: any) {
      Taro.showToast({ title: '登录失败，请重试', icon: 'none' })
    }
  }

  const handlePay = async () => {
    if (!createdOrder) return
    setPayLoading(true)
    try {
      const params = await getPaymentParams(createdOrder.order_no)
      
      await Taro.requestPayment({
        timeStamp: params.timeStamp,
        nonceStr: params.nonceStr,
        package: params.package,
        signType: params.signType || 'MD5',
        paySign: params.paySign
      })

      clear()
      Taro.showToast({ title: '支付成功', icon: 'success' })
      setTimeout(() => {
        Taro.redirectTo({ url: `/pages/order/detail?order_no=${createdOrder.order_no}` })
      }, 1500)
    } catch (e: any) {
      if (e.errMsg?.includes('cancel')) {
        Taro.showToast({ title: '已取消支付', icon: 'none' })
      } else {
        Taro.showToast({ title: '支付失败，请重新支付', icon: 'none' })
      }
    } finally {
      setPayLoading(false)
      setShowPayDialog(false)
    }
  }

  const selectedCoupon = coupons.find(c => c.id === couponId)

  return (
    <View className={styles.container}>
      <ScrollView scrollY className={styles.scrollContent}>
        {currentStore && (
          <View className={styles.section}>
            <View className={styles.sectionHeader}>
              <Text className={styles.sectionIcon}>📍</Text>
              <Text className={styles.sectionTitle}>门店信息</Text>
            </View>
            <Cell
              title={currentStore.name}
              description={currentStore.address}
            />
            <Cell
              title='桌号'
              description={tableNo || '请选择桌号'}
              onClick={() => Taro.navigateTo({ url: '/pages/cart/index' })}
            />
          </View>
        )}

        <View className={styles.section}>
          <View className={styles.sectionHeader}>
            <Text className={styles.sectionIcon}>🍜</Text>
            <Text className={styles.sectionTitle}>商品清单</Text>
          </View>
          {items.map(item => (
            <View key={item.id} className={styles.orderItem}>
              <Image
                className={styles.itemImage}
                src={item.product_image || 'https://trae-api-cn.mchost.guru/api/ide/v1/text_to_image?prompt=food&image_size=square'}
              />
              <View className={styles.itemInfo}>
                <Text className={styles.itemName} numberOfLines={1}>{item.product_name}</Text>
                <Text className={styles.itemSpec}>
                  {item.sku_name}
                  {item.attribute_names.map((name, idx) => (
                    <Text key={idx}> · {name}</Text>
                  ))}
                </Text>
                <View className={styles.itemBottom}>
                  <View className={styles.itemPrice}>
                    <Text className={styles.priceSymbol}>¥</Text>
                    <Text className={styles.priceValue}>{item.price.toFixed(2)}</Text>
                  </View>
                  <Text className={styles.itemQuantity}>× {item.quantity}</Text>
                </View>
              </View>
            </View>
          ))}
        </View>

        <View className={styles.section}>
          <View className={styles.sectionHeader}>
            <Text className={styles.sectionIcon}>🎟️</Text>
            <Text className={styles.sectionTitle}>优惠券</Text>
          </View>
          <Cell
            title={selectedCoupon ? selectedCoupon.coupon.name : '选择优惠券'}
            description={selectedCoupon ? `已优惠 ¥${couponDiscount.toFixed(2)}` : `${coupons.length} 张可用`}
            extra='▶'
            onClick={() => isLogin() && setShowCouponDialog(true)}
          />
          {selectedCoupon && (
            <Cell
              title='不使用优惠券'
              description='点击取消优惠券使用'
              onClick={() => handleSelectCoupon(null)}
            />
          )}
        </View>

        {remark && (
          <View className={styles.section}>
            <View className={styles.sectionHeader}>
              <Text className={styles.sectionIcon}>📝</Text>
              <Text className={styles.sectionTitle}>备注</Text>
            </View>
            <Cell description={remark} />
          </View>
        )}

        <View className={styles.section}>
          <View className={styles.sectionHeader}>
            <Text className={styles.sectionIcon}>💰</Text>
            <Text className={styles.sectionTitle}>费用明细</Text>
          </View>
          <Cell title='商品合计' extra={`¥${total.toFixed(2)}`} />
          {couponDiscount > 0 && (
            <Cell title='优惠券抵扣' extra={`-¥${couponDiscount.toFixed(2)}`} />
          )}
          <Cell title='实付金额' extra={`¥${actualTotal.toFixed(2)}`} />
        </View>
      </ScrollView>

      <View className={styles.footer}>
        <View className={styles.footerTotal}>
          <Text className={styles.totalLabel}>实付：</Text>
          <Text className={styles.totalSymbol}>¥</Text>
          <Text className={styles.totalValue}>{actualTotal.toFixed(2)}</Text>
        </View>
        <View
          className={`${styles.submitBtn} ${submitting ? styles.btnDisabled : ''}`}
          onClick={handleSubmit}
        >
          {submitting ? (
            <Loading type='spinner' size='16px' color='#fff' />
          ) : (
            <Text className={styles.submitText}>提交订单</Text>
          )}
        </View>
      </View>

      <Dialog
        visible={showCouponDialog}
        title='选择优惠券'
        footer={null}
        onClose={() => setShowCouponDialog(false)}
        style={{ height: '70vh' }}
      >
        <ScrollView scrollY className={styles.couponList}>
          {coupons.length === 0 ? (
            <View className={styles.emptyCoupon}>
              <Text className={styles.emptyText}>暂无可用优惠券</Text>
            </View>
          ) : (
            coupons.map(coupon => (
              <View
                key={coupon.id}
                className={`${styles.couponItem} ${selectedCoupon?.id === coupon.id ? styles.couponSelected : ''}`}
                onClick={() => handleSelectCoupon(coupon)}
              >
                <View className={styles.couponLeft}>
                  <Text className={styles.couponValue}>
                    {coupon.coupon.type === 1 ? '¥' : ''}
                    {coupon.coupon.value}
                    {coupon.coupon.type === 2 ? '折' : ''}
                  </Text>
                  <Text className={styles.couponCondition}>满{coupon.coupon.min_amount}可用</Text>
                </View>
                <View className={styles.couponInfo}>
                  <Text className={styles.couponName}>{coupon.coupon.name}</Text>
                  <Text className={styles.couponTime}>
                    {coupon.coupon.start_time} 至 {coupon.coupon.end_time}
                  </Text>
                </View>
              </View>
            ))
          )}
        </ScrollView>
      </Dialog>

      <Dialog
        visible={showPayDialog}
        title='确认支付'
        content={`需支付 ¥${actualTotal.toFixed(2)}`}
        okText='立即支付'
        cancelText='稍后支付'
        onOk={handlePay}
        onCancel={() => {
          setShowPayDialog(false)
          clear()
          Taro.redirectTo({ url: '/pages/order/list' })
        }}
      />
    </View>
  )
}

export default OrderConfirm
