import React, { useState, useEffect } from 'react'
import { View, Text, Image, ScrollView, Input } from '@tarojs/components'
import Taro from '@tarojs/taro'
import { Loading, Dialog, Cell, Button } from '@nutui/nutui-react-taro'
import { useCartStore } from '../../store/cart'
import { useAppStore } from '../../store/app'
import { getAvailableCoupons, MemberCoupon, calculateBestCombination, BestPromotionResponse } from '../../services/coupon'
import { createOrder, getPaymentParams } from '../../services/order'
import type { OrderCreateDTO, OrderItem } from '../../services/order'
import { orderTypeMap, planRoute } from '../../services/delivery'
import type { OrderType } from '../../services/delivery'
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
  const [bestPromo, setBestPromo] = useState<BestPromotionResponse | null>(null)

  const [orderType, setOrderType] = useState<OrderType>('dine_in')
  const [deliveryAddress, setDeliveryAddress] = useState('')
  const [deliveryContact, setDeliveryContact] = useState('')
  const [deliveryPhone, setDeliveryPhone] = useState('')
  const [deliveryFee, setDeliveryFee] = useState(0)

  const actualTotal = bestPromo
    ? bestPromo.final_amount + deliveryFee
    : Math.max(0, total - couponDiscount) + deliveryFee

  const productIds = Array.from(new Set(items.map(item => item.product_id)))

  useEffect(() => {
    loadPromotions()
    loadCoupons()
  }, [total, productIds.join(',')])

  useEffect(() => {
    if (orderType === 'delivery' && deliveryAddress && currentStore) {
      estimateDeliveryFee()
    }
  }, [orderType, deliveryAddress])

  const estimateDeliveryFee = async () => {
    try {
      if (!currentStore) return
      const result = await planRoute(
        116.397428, 39.90923,
        116.407, 39.919
      )
      setDeliveryFee(result.fee)
    } catch {
      setDeliveryFee(5)
    }
  }

  const loadPromotions = async () => {
    if (!currentStore || productIds.length === 0) return
    try {
      const result = await calculateBestCombination({
        store_id: currentStore.id,
        amount: total,
        product_ids: productIds,
        member_coupon_id: couponId || undefined,
        member_id: user?.id
      })
      setBestPromo(result)
    } catch {}
  }

  const loadCoupons = async () => {
    if (!isLogin()) return
    setLoading(true)
    try {
      const list = await getAvailableCoupons(total, productIds)
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
      const discount = coupon.coupon.type === 'fixed'
        ? Number(coupon.coupon.value)
        : coupon.coupon.type === 'percentage'
          ? Math.min(Number(total) * Number(coupon.coupon.value) / 10, Number(coupon.coupon.max_discount) || Infinity)
          : 0
      setCoupon(coupon.id, Math.min(discount, Number(total)))
    }
    setShowCouponDialog(false)
    setTimeout(loadPromotions, 100)
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

    if (orderType === 'dine_in' && !tableNo) {
      Taro.showToast({ title: '请输入桌号', icon: 'none' })
      return
    }

    if (orderType === 'delivery') {
      if (!deliveryAddress) {
        Taro.showToast({ title: '请填写配送地址', icon: 'none' })
        return
      }
      if (!deliveryContact || !deliveryPhone) {
        Taro.showToast({ title: '请填写联系人信息', icon: 'none' })
        return
      }
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
        table_no: orderType === 'dine_in' ? tableNo : undefined,
        remark: remark,
        coupon_id: couponId || undefined,
        member_coupon_id: couponId || undefined,
        member_id: user?.id,
        order_type: orderType,
        source: 'miniprogram',
        delivery_address: orderType === 'delivery' ? deliveryAddress : undefined,
        delivery_contact: orderType === 'delivery' ? deliveryContact : undefined,
        delivery_phone: orderType === 'delivery' ? deliveryPhone : undefined,
        delivery_fee: orderType === 'delivery' ? deliveryFee : undefined,
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

  const getCouponTypeLabel = (type: string) => {
    const map: Record<string, string> = {
      fixed: '满减券',
      percentage: '折扣券',
      exchange: '兑换券'
    }
    return map[type] || type
  }

  const orderTypes: OrderType[] = ['dine_in', 'pickup', 'delivery', 'takeout']

  return (
    <View className={styles.container}>
      <ScrollView scrollY className={styles.scrollContent}>
        <View className={styles.section}>
          <View className={styles.sectionHeader}>
            <Text className={styles.sectionIcon}>🛎️</Text>
            <Text className={styles.sectionTitle}>用餐方式</Text>
          </View>
          <View className={styles.orderTypeRow}>
            {orderTypes.map(type => (
              <View
                key={type}
                className={`${styles.orderTypeItem} ${orderType === type ? styles.orderTypeActive : ''}`}
                onClick={() => setOrderType(type)}
              >
                <Text className={styles.orderTypeIcon}>{orderTypeMap[type].icon}</Text>
                <Text className={styles.orderTypeLabel}>{orderTypeMap[type].label}</Text>
              </View>
            ))}
          </View>
        </View>

        {orderType === 'dine_in' && currentStore && (
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

        {(orderType === 'pickup' || orderType === 'takeout') && currentStore && (
          <View className={styles.section}>
            <View className={styles.sectionHeader}>
              <Text className={styles.sectionIcon}>📍</Text>
              <Text className={styles.sectionTitle}>取餐信息</Text>
            </View>
            <Cell
              title={currentStore.name}
              description={currentStore.address}
            />
            <View className={styles.pickupHint}>
              <Text className={styles.pickupHintText}>
                {orderType === 'pickup' ? '备餐完成后将推送取餐码，请凭码取餐' : '备餐完成后通知您，请到店取餐'}
              </Text>
            </View>
          </View>
        )}

        {orderType === 'delivery' && (
          <View className={styles.section}>
            <View className={styles.sectionHeader}>
              <Text className={styles.sectionIcon}>🛵</Text>
              <Text className={styles.sectionTitle}>配送信息</Text>
            </View>
            <View className={styles.formField}>
              <Text className={styles.formLabel}>联系人</Text>
              <Input
                className={styles.formInput}
                placeholder='请输入联系人姓名'
                value={deliveryContact}
                onInput={(e) => setDeliveryContact(e.detail.value)}
              />
            </View>
            <View className={styles.formField}>
              <Text className={styles.formLabel}>手机号</Text>
              <Input
                className={styles.formInput}
                placeholder='请输入联系电话'
                type='number'
                value={deliveryPhone}
                onInput={(e) => setDeliveryPhone(e.detail.value)}
              />
            </View>
            <View className={styles.formField}>
              <Text className={styles.formLabel}>配送地址</Text>
              <Input
                className={styles.formInput}
                placeholder='请输入详细配送地址'
                value={deliveryAddress}
                onInput={(e) => setDeliveryAddress(e.detail.value)}
              />
            </View>
            {deliveryFee > 0 && (
              <View className={styles.deliveryFeeRow}>
                <Text className={styles.deliveryFeeLabel}>配送费</Text>
                <Text className={styles.deliveryFeeValue}>¥{deliveryFee.toFixed(2)}</Text>
              </View>
            )}
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

        {bestPromo && bestPromo.promotions.length > 0 && (
          <View className={styles.section}>
            <View className={styles.sectionHeader}>
              <Text className={styles.sectionIcon}>🏷️</Text>
              <Text className={styles.sectionTitle}>已享优惠</Text>
            </View>
            {bestPromo.promotions.map((p, idx) => (
              <Cell
                key={idx}
                title={p.name}
                description={p.promotion_id ? '营销活动' : p.coupon_id ? '优惠券' : ''}
                extra={`-¥${Number(p.discount).toFixed(2)}`}
              />
            ))}
          </View>
        )}

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
          {deliveryFee > 0 && (
            <Cell title='配送费' extra={`¥${deliveryFee.toFixed(2)}`} />
          )}
          {bestPromo && bestPromo.total_discount > 0 && (
            <Cell title='活动优惠' extra={`-¥${Number(bestPromo.total_discount).toFixed(2)}`} />
          )}
          {!bestPromo && couponDiscount > 0 && (
            <Cell title='优惠券抵扣' extra={`-¥${couponDiscount.toFixed(2)}`} />
          )}
          <Cell title='实付金额' extra={`¥${Number(actualTotal).toFixed(2)}`} />
        </View>
      </ScrollView>

      <View className={styles.footer}>
        <View className={styles.footerTotal}>
          <Text className={styles.totalLabel}>实付：</Text>
          <Text className={styles.totalSymbol}>¥</Text>
          <Text className={styles.totalValue}>{Number(actualTotal).toFixed(2)}</Text>
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
                    {coupon.coupon.type === 'fixed' ? '¥' : ''}
                    {coupon.coupon.value}
                    {coupon.coupon.type === 'percentage' ? '折' : ''}
                  </Text>
                  <Text className={styles.couponCondition}>满{coupon.coupon.min_amount}可用</Text>
                </View>
                <View className={styles.couponInfo}>
                  <Text className={styles.couponName}>{coupon.coupon.name}</Text>
                  <Text className={styles.couponType}>{getCouponTypeLabel(coupon.coupon.type)}</Text>
                  <Text className={styles.couponTime}>
                    {coupon.expire_at ? `有效期至 ${coupon.expire_at.slice(0, 10)}` : '长期有效'}
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
        content={`需支付 ¥${Number(actualTotal).toFixed(2)}`}
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
