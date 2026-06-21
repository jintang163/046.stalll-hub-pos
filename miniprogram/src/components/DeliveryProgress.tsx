import React, { useState, useEffect, useRef } from 'react'
import { View, Text } from '@tarojs/components'
import Taro from '@tarojs/taro'
import { getDeliveryTracking, getRiderLocation, deliveryStatusMap } from '../../services/delivery'
import type { DeliveryTracking } from '../../services/delivery'
import styles from './DeliveryProgress.module.scss'

interface DeliveryProgressProps {
  orderId: number
  orderType: string
  pickupCode?: string
}

const DeliveryProgress: React.FC<DeliveryProgressProps> = ({ orderId, orderType, pickupCode }) => {
  const [tracking, setTracking] = useState<DeliveryTracking | null>(null)
  const [refreshing, setRefreshing] = useState(false)
  const timerRef = useRef<any>(null)

  useEffect(() => {
    if (orderType === 'delivery') {
      loadTracking()
      timerRef.current = setInterval(loadTracking, 15000)
    }
    return () => {
      if (timerRef.current) {
        clearInterval(timerRef.current)
      }
    }
  }, [orderId, orderType])

  const loadTracking = async () => {
    setRefreshing(true)
    try {
      const data = await getDeliveryTracking(orderId)
      setTracking(data)
    } catch {}
    setRefreshing(false)
  }

  if (orderType === 'pickup') {
    return (
      <View className={styles.container}>
        <View className={styles.pickupCard}>
          <View className={styles.pickupHeader}>
            <Text className={styles.pickupEmoji}>🏪</Text>
            <Text className={styles.pickupTitle}>到店自提</Text>
          </View>
          {pickupCode ? (
            <View className={styles.pickupCodeSection}>
              <Text className={styles.pickupCodeLabel}>取餐码</Text>
              <View className={styles.pickupCodeBox}>
                <Text className={styles.pickupCodeValue}>{pickupCode}</Text>
              </View>
              <Text className={styles.pickupCodeHint}>备餐完成后凭此码取餐</Text>
            </View>
          ) : (
            <View className={styles.pickupWaiting}>
              <Text className={styles.waitingText}>等待备餐中，取餐码即将生成...</Text>
            </View>
          )}
        </View>
      </View>
    )
  }

  if (orderType === 'takeout') {
    return (
      <View className={styles.container}>
        <View className={styles.pickupCard}>
          <View className={styles.pickupHeader}>
            <Text className={styles.pickupEmoji}>🥡</Text>
            <Text className={styles.pickupTitle}>到店自取</Text>
          </View>
          <View className={styles.pickupWaiting}>
            <Text className={styles.waitingText}>备餐完成后通知您到店取餐</Text>
          </View>
        </View>
      </View>
    )
  }

  if (orderType !== 'delivery' || !tracking) {
    return null
  }

  const statusInfo = deliveryStatusMap[tracking.delivery_status] || { text: '未知', color: '#999' }

  const getStepStatus = (step: number) => {
    if (tracking.delivery_status >= step) return 'done'
    if (tracking.delivery_status === step - 1) return 'current'
    return 'pending'
  }

  const steps = [
    { key: 0, title: '商家接单', icon: '📋' },
    { key: 1, title: '骑手取餐', icon: '🛵' },
    { key: 2, title: '配送中', icon: '🚀' },
    { key: 3, title: '已送达', icon: '✅' },
  ]

  return (
    <View className={styles.container}>
      <View className={styles.statusCard}>
        <View className={styles.statusHeader}>
          <Text className={styles.statusEmoji}>
            {tracking.delivery_status === 3 ? '✅' : '🛵'}
          </Text>
          <View className={styles.statusInfo}>
            <Text className={styles.statusText} style={{ color: statusInfo.color }}>
              {statusInfo.text}
            </Text>
            {tracking.estimated_time && tracking.delivery_status < 3 && (
              <Text className={styles.estimatedTime}>
                预计 {new Date(tracking.estimated_time).toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' })} 送达
              </Text>
            )}
          </View>
        </View>

        <View className={styles.stepsRow}>
          {steps.map((step, idx) => (
            <View key={step.key} className={styles.stepWrapper}>
              <View
                className={`${styles.stepDot} ${getStepStatus(step.key) === 'done' ? styles.stepDone : ''} ${getStepStatus(step.key) === 'current' ? styles.stepCurrent : ''}`}
              >
                <Text className={styles.stepDotIcon}>{step.icon}</Text>
              </View>
              {idx < steps.length - 1 && (
                <View className={`${styles.stepLine} ${getStepStatus(step.key) === 'done' ? styles.stepLineDone : ''}`} />
              )}
              <Text className={`${styles.stepTitle} ${getStepStatus(step.key) === 'done' || getStepStatus(step.key) === 'current' ? styles.stepTitleActive : ''}`}>
                {step.title}
              </Text>
            </View>
          ))}
        </View>
      </View>

      {tracking.rider_id > 0 && tracking.delivery_status >= 1 && tracking.delivery_status < 3 && (
        <View className={styles.riderCard}>
          <View className={styles.riderInfo}>
            <View className={styles.riderAvatar}>
              <Text className={styles.riderAvatarText}>🛵</Text>
            </View>
            <View className={styles.riderDetail}>
              <Text className={styles.riderName}>{tracking.rider_name || '骑手'}</Text>
              <Text className={styles.riderPhone}>{tracking.rider_phone}</Text>
            </View>
            <View className={styles.riderActions}>
              <View
                className={styles.callBtn}
                onClick={() => {
                  if (tracking.rider_phone) {
                    Taro.makePhoneCall({ phoneNumber: tracking.rider_phone })
                  }
                }}
              >
                <Text className={styles.callBtnText}>📞 联系骑手</Text>
              </View>
            </View>
          </View>

          {(tracking.rider_lng > 0 || tracking.rider_lat > 0) && (
            <View className={styles.locationCard}>
              <View className={styles.locationRow}>
                <Text className={styles.locationIcon}>📍</Text>
                <View className={styles.locationInfo}>
                  <Text className={styles.locationLabel}>骑手位置</Text>
                  <Text className={styles.locationCoords}>
                    {tracking.rider_lat.toFixed(4)}, {tracking.rider_lng.toFixed(4)}
                  </Text>
                </View>
              </View>
              {tracking.distance > 0 && (
                <View className={styles.distanceRow}>
                  <Text className={styles.distanceText}>
                    距您约 {tracking.distance.toFixed(1)} km · 约 {tracking.duration} 分钟
                  </Text>
                </View>
              )}
            </View>
          )}
        </View>
      )}

      {tracking.trackings && tracking.trackings.length > 0 && (
        <View className={styles.trackCard}>
          <Text className={styles.trackTitle}>配送轨迹</Text>
          {tracking.trackings.slice(0, 5).map((point, idx) => (
            <View key={idx} className={styles.trackItem}>
              <View className={styles.trackDot} />
              <Text className={styles.trackTime}>
                {new Date(point.timestamp * 1000).toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' })}
              </Text>
              <Text className={styles.trackSpeed}>
                {point.speed > 0 ? `${(point.speed * 3.6).toFixed(0)} km/h` : '等待中'}
              </Text>
            </View>
          ))}
        </View>
      )}

      <View className={styles.refreshRow}>
        <Text className={styles.refreshHint}>
          {refreshing ? '刷新中...' : '每15秒自动刷新骑手位置'}
        </Text>
      </View>
    </View>
  )
}

export default DeliveryProgress
