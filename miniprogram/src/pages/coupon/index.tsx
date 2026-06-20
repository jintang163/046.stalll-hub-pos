import React, { useState, useEffect, useCallback } from 'react'
import { View, Text, ScrollView } from '@tarojs/components'
import Taro, { useDidShow } from '@tarojs/taro'
import { Loading, Toast } from '@nutui/nutui-react-taro'
import {
  Coupon,
  Promotion,
  getCouponList as getAdminCouponList,
  getActivePromotions,
  claimCoupon,
  getMyCoupons,
  MemberCoupon
} from '../../services/coupon'
import { isLogin, loginByCode } from '../../services/auth'
import { useAppStore } from '../../store/app'
import styles from './index.module.scss'

const getCouponList = async (): Promise<Coupon[]> => {
  try {
    const res = await Taro.request({
      url: '/api/v1/coupons',
      method: 'GET',
      data: { status: 1, page: 1, page_size: 100 }
    })
    return (res.data as any)?.list || []
  } catch {
    return []
  }
}

const CouponCenter: React.FC = () => {
  const user = useAppStore(state => state.user)
  const [activeTab, setActiveTab] = useState<'coupons' | 'promotions'>('coupons')
  const [coupons, setCoupons] = useState<Coupon[]>([])
  const [promotions, setPromotions] = useState<Promotion[]>([])
  const [myCoupons, setMyCoupons] = useState<MemberCoupon[]>([])
  const [loading, setLoading] = useState(false)
  const [claimingId, setClaimingId] = useState<number | null>(null)

  const loadData = useCallback(async () => {
    setLoading(true)
    try {
      const [couponList, promoList, mineList] = await Promise.all([
        getCouponList(),
        getActivePromotions().catch(() => [] as Promotion[]),
        isLogin() ? getMyCoupons().catch(() => [] as MemberCoupon[]) : Promise.resolve([] as MemberCoupon[])
      ])
      setCoupons(couponList.filter(c => c.status === 1))
      setPromotions(promoList)
      setMyCoupons(mineList)
    } finally {
      setLoading(false)
    }
  }, [])

  useDidShow(() => {
    loadData()
  })

  const handleClaim = async (couponId: number) => {
    if (!isLogin()) {
      const res = await Taro.login()
      if (res.code) {
        try {
          const result = await loginByCode(res.code)
          Taro.setStorageSync('token', result.token)
          Taro.setStorageSync('userInfo', result.user)
          useAppStore.getState().setUser(result.user)
        } catch {
          Toast.show('请先登录')
          return
        }
      } else {
        Toast.show('请先登录')
        return
      }
    }

    setClaimingId(couponId)
    try {
      await claimCoupon({ coupon_id: couponId })
      Toast.show('领取成功')
      loadData()
    } catch (e: any) {
      Toast.show(e.message || '领取失败')
    } finally {
      setClaimingId(null)
    }
  }

  const isClaimed = (couponId: number) => {
    return myCoupons.some(mc => mc.coupon_id === couponId)
  }

  const getCouponTypeClass = (type: string) => {
    if (type === 'percentage') return styles.discount
    if (type === 'exchange') return styles.exchange
    return ''
  }

  const getPromoTagClass = (type: string) => {
    if (type === 'discount') return styles.discount
    if (type === 'tiered') return styles.tiered
    return ''
  }

  const getPromoTypeName = (type: string) => {
    const map: Record<string, string> = {
      full_reduction: '满减',
      discount: '折扣',
      tiered: '阶梯'
    }
    return map[type] || '活动'
  }

  const formatDate = (date: string) => {
    if (!date) return ''
    return new Date(date).toLocaleDateString()
  }

  const renderCouponCard = (coupon: Coupon) => {
    const claimed = isClaimed(coupon.id)
    const stockLeft = coupon.total_count > 0 ? coupon.total_count - coupon.used_count : -1
    const outOfStock = stockLeft === 0

    return (
      <View key={coupon.id} className={styles.couponCard}>
        <View className={styles.couponInner}>
          <View className={`${styles.couponLeft} ${getCouponTypeClass(coupon.type)}`}>
            <View className={styles.couponValue}>
              {coupon.type === 'fixed' && <Text className={styles.symbol}>¥</Text>}
              {coupon.value}
              {coupon.type === 'percentage' && <Text className={styles.unit}>折</Text>}
            </View>
            <View className={styles.couponCondition}>
              {coupon.min_amount > 0 ? `满${coupon.min_amount}可用` : '无门槛'}
            </View>
          </View>
          <View className={styles.couponRight}>
            <View>
              <View className={styles.couponName}>{coupon.name}</View>
              <View className={styles.couponDesc}>
                {coupon.description || (coupon.applicable_type === 'all' ? '全场通用' : '指定商品可用')}
              </View>
            </View>
            <View className={styles.couponMeta}>
              <View className={styles.couponTime}>
                {coupon.validity_type === 'fixed'
                  ? `${formatDate(coupon.start_time)} 至 ${formatDate(coupon.end_time)}`
                  : `领取后${coupon.validity_days}天有效`}
              </View>
              {stockLeft >= 0 && (
                <View className={styles.couponStock}>剩余 {stockLeft} 张</View>
              )}
            </View>
          </View>
        </View>
        <View
          className={`${styles.claimBtn} ${claimed ? styles.claimed : ''} ${outOfStock ? styles.disabled : ''}`}
          onClick={() => !claimed && !outOfStock && handleClaim(coupon.id)}
        >
          {claimingId === coupon.id ? '领取中...' : claimed ? '已领取' : outOfStock ? '已领完' : '立即领取'}
        </View>
      </View>
    )
  }

  const renderPromotionCard = (promo: Promotion) => {
    return (
      <View key={promo.id} className={styles.promotionCard}>
        <View className={styles.promoHeader}>
          <View className={styles.promoName}>{promo.name}</View>
          <View className={`${styles.promoTag} ${getPromoTagClass(promo.type)}`}>
            {getPromoTypeName(promo.type)}
          </View>
        </View>

        {promo.type === 'full_reduction' && (
          <View className={styles.promoRule}>
            满 ¥{promo.min_amount} 减 ¥{promo.discount_value}
          </View>
        )}

        {promo.type === 'discount' && (
          <View className={styles.promoRule}>
            全场 {promo.discount_value} 折
            {promo.max_discount > 0 && ` （最高减¥${promo.max_discount}）`}
          </View>
        )}

        {promo.type === 'tiered' && promo.tiers && (
          <View className={styles.tierList}>
            {promo.tiers
              .sort((a, b) => a.min_amount - b.min_amount)
              .map((tier, idx) => (
                <View key={idx} className={styles.tierItem}>
                  第{idx + 1}档：满 ¥{tier.min_amount} 减 ¥{tier.discount_value}
                </View>
              ))}
          </View>
        )}

        {promo.description && (
          <View className={styles.promoDesc}>{promo.description}</View>
        )}

        <View className={styles.promoTime}>
          活动时间：{formatDate(promo.start_time)} 至 {formatDate(promo.end_time)}
        </View>
      </View>
    )
  }

  return (
    <View className={styles.container}>
      <View className={styles.header}>
        <View className={styles.title}>🎫 领券中心</View>
        <View className={styles.subtitle}>优惠多多，先到先得</View>
      </View>

      <View className={styles.tabs}>
        <View
          className={`${styles.tab} ${activeTab === 'coupons' ? styles.active : ''}`}
          onClick={() => setActiveTab('coupons')}
        >
          优惠券
        </View>
        <View
          className={`${styles.tab} ${activeTab === 'promotions' ? styles.active : ''}`}
          onClick={() => setActiveTab('promotions')}
        >
          营销活动
        </View>
      </View>

      <ScrollView scrollY className={styles.couponList}>
        {loading ? (
          <View className={styles.loading}>
            <Loading type='spinner' size='24px' />
          </View>
        ) : activeTab === 'coupons' ? (
          coupons.length > 0 ? (
            coupons.map(renderCouponCard)
          ) : (
            <View className={styles.emptyState}>
              <View className={styles.emptyIcon}>🎟️</View>
              <View className={styles.emptyText}>暂无可用优惠券</View>
            </View>
          )
        ) : (
          promotions.length > 0 ? (
            promotions.map(renderPromotionCard)
          ) : (
            <View className={styles.emptyState}>
              <View className={styles.emptyIcon}>🎉</View>
              <View className={styles.emptyText}>暂无进行中的活动</View>
            </View>
          )
        )}
      </ScrollView>
    </View>
  )
}

export default CouponCenter
