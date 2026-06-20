import React, { useState, useEffect, useRef } from 'react'
import { View, Text, Input, ScrollView } from '@tarojs/components'
import Taro, { useDidShow, usePullDownRefresh, useDidHide } from '@tarojs/taro'
import { Tabs, TabPane, Loading, Dialog } from '@nutui/nutui-react-taro'
import { queueApi } from '../../services/table'
import type { QueueItem } from '../../services/table'
import { getCurrentStore } from '../../services/store'
import type { Store } from '../../services/store'
import { isLogin } from '../../services/auth'
import { queue2Service, type QueueMessage } from '../../services/queue2'
import classNames from 'classnames'
import dayjs from 'dayjs'
import styles from './queue.module.scss'

const TABLE_TYPES = [
  { value: 'small', name: '小桌', desc: '1-4人', minPeople: 1, maxPeople: 4 },
  { value: 'medium', name: '中桌', desc: '5-6人', minPeople: 5, maxPeople: 6 },
  { value: 'large', name: '大桌', desc: '7-10人', minPeople: 7, maxPeople: 10 }
]

const STATUS_MAP: Record<number, { text: string; className: string }> = {
  0: { text: '排队中', className: styles.statusQueuing },
  1: { text: '已叫号', className: styles.statusCalled },
  2: { text: '已入座', className: styles.statusSeated },
  3: { text: '已取消', className: styles.statusCancelled },
  4: { text: '已过号', className: styles.statusCancelled }
}

const QueuePage: React.FC = () => {
  const [tabValue, setTabValue] = useState(0)
  const [loading, setLoading] = useState(false)
  const [store, setStore] = useState<Store | null>(null)
  const [waitingCount, setWaitingCount] = useState({ small: 0, medium: 0, large: 0 })
  const [tableType, setTableType] = useState('small')
  const [peopleCount, setPeopleCount] = useState(2)
  const [memberName, setMemberName] = useState('')
  const [memberPhone, setMemberPhone] = useState('')
  const [myQueues, setMyQueues] = useState<QueueItem[]>([])
  const [showSuccessPopup, setShowSuccessPopup] = useState(false)
  const [successQueue, setSuccessQueue] = useState<QueueItem | null>(null)
  const [showCallingAlert, setShowCallingAlert] = useState(false)
  const [callingQueue, setCallingQueue] = useState<QueueItem | null>(null)
  const [waitedTime, setWaitedTime] = useState(0)
  const timerRef = useRef<NodeJS.Timeout | null>(null)
  const refreshTimerRef = useRef<NodeJS.Timeout | null>(null)
  const previousStatusRef = useRef<number | null>(null)

  useDidShow(() => {
    if (isLogin()) {
      loadData()
      startWebSocket()
    } else {
      Taro.navigateTo({ url: '/pages/user/index' })
    }
  })

  useDidHide(() => {
    stopWebSocket()
  })

  usePullDownRefresh(() => {
    loadData(true)
  })

  useEffect(() => {
    return () => {
      if (timerRef.current) clearInterval(timerRef.current)
      if (refreshTimerRef.current) clearInterval(refreshTimerRef.current)
    }
  }, [])

  useEffect(() => {
    if (myQueues.length > 0) {
      const activeQueue = myQueues.find(q => q.status === 0 || q.status === 1)
      if (activeQueue) {
        startWaitedTimeTimer(activeQueue.createdAt)
        startAutoRefresh()
        checkStatusChange(activeQueue)
      } else {
        stopTimers()
      }
    } else {
      stopTimers()
    }
  }, [myQueues])

  const startWaitedTimeTimer = (createdAt: string) => {
    if (timerRef.current) clearInterval(timerRef.current)
    const updateWaitedTime = () => {
      const now = dayjs()
      const created = dayjs(createdAt)
      const diff = now.diff(created, 'second')
      setWaitedTime(diff)
    }
    updateWaitedTime()
    timerRef.current = setInterval(updateWaitedTime, 1000)
  }

  const startAutoRefresh = () => {
    if (refreshTimerRef.current) clearInterval(refreshTimerRef.current)
    refreshTimerRef.current = setInterval(() => {
      loadMyQueues()
    }, 30000)
  }

  const startWebSocket = () => {
    const storeId = store?.id || 1
    queue2Service.connectWebSocket(storeId)
    queue2Service.addListener(handleWsMessage)
  }

  const stopWebSocket = () => {
    queue2Service.removeListener(handleWsMessage)
    queue2Service.disconnectWebSocket()
  }

  const handleWsMessage = (msg: QueueMessage) => {
    console.log('[WS] received:', msg)
    if (msg.type === 'call' && msg.queueNumber) {
      const activeQueue = myQueues.find(q => q.status === 0 || q.status === 1)
      if (activeQueue && activeQueue.queueNumber === msg.queueNumber) {
        triggerCallingNotification(activeQueue)
        loadMyQueues()
      }
    }
    if (msg.type === 'arrive') {
      loadMyQueues()
    }
    if (msg.type === 'cancel') {
      loadMyQueues()
    }
  }

  const stopTimers = () => {
    if (timerRef.current) {
      clearInterval(timerRef.current)
      timerRef.current = null
    }
    if (refreshTimerRef.current) {
      clearInterval(refreshTimerRef.current)
      refreshTimerRef.current = null
    }
  }

  const checkStatusChange = (queue: QueueItem) => {
    if (previousStatusRef.current === 0 && queue.status === 1) {
      triggerCallingNotification(queue)
    }
    previousStatusRef.current = queue.status
  }

  const triggerCallingNotification = (queue: QueueItem) => {
    setCallingQueue(queue)
    setShowCallingAlert(true)
    Taro.vibrateLong()
    speak(`请${queue.queueNumber}号顾客到${queue.tableNo || '前台'}入座`)
    setTimeout(() => {
      setShowCallingAlert(false)
    }, 5000)
  }

  const speak = (text: string) => {
    try {
      if (typeof window !== 'undefined' && 'speechSynthesis' in window) {
        const utterance = new SpeechSynthesisUtterance(text)
        utterance.lang = 'zh-CN'
        utterance.rate = 0.9
        utterance.volume = 1
        window.speechSynthesis.speak(utterance)
      }
    } catch (e) {
      console.log('Speech not supported')
    }
  }

  const loadData = async (isRefresh = false) => {
    setLoading(true)
    try {
      const [storeData, countData] = await Promise.all([
        getCurrentStore(),
        queueApi.getWaitingCount(1)
      ])
      setStore(storeData)
      setWaitingCount(countData)
      await loadMyQueues()
    } catch (e: any) {
      if (e.message?.includes('401')) {
        Taro.removeStorageSync('token')
        Taro.navigateTo({ url: '/pages/user/index' })
      }
    } finally {
      setLoading(false)
      if (isRefresh) Taro.stopPullDownRefresh()
    }
  }

  const loadMyQueues = async () => {
    try {
      const userInfo = Taro.getStorageSync('userInfo')
      if (userInfo?.id) {
        const result = await queueApi.getMy({ memberId: userInfo.id, storeId: 1 })
        setMyQueues(result)
      }
    } catch (e) {
      console.error('Load my queues error:', e)
    }
  }

  const handleTableTypeChange = (type: string) => {
    setTableType(type)
    const tableTypeConfig = TABLE_TYPES.find(t => t.value === type)
    if (tableTypeConfig) {
      if (peopleCount < tableTypeConfig.minPeople) {
        setPeopleCount(tableTypeConfig.minPeople)
      } else if (peopleCount > tableTypeConfig.maxPeople) {
        setPeopleCount(tableTypeConfig.maxPeople)
      }
    }
  }

  const handlePeopleCountChange = (delta: number) => {
    const tableTypeConfig = TABLE_TYPES.find(t => t.value === tableType)
    if (!tableTypeConfig) return
    const newCount = peopleCount + delta
    if (newCount >= tableTypeConfig.minPeople && newCount <= tableTypeConfig.maxPeople) {
      setPeopleCount(newCount)
    }
  }

  const validateForm = (): boolean => {
    if (!memberName.trim()) {
      Taro.showToast({ title: '请输入姓名', icon: 'none' })
      return false
    }
    if (!memberPhone.trim()) {
      Taro.showToast({ title: '请输入手机号', icon: 'none' })
      return false
    }
    if (!/^1[3-9]\d{9}$/.test(memberPhone)) {
      Taro.showToast({ title: '请输入正确的手机号', icon: 'none' })
      return false
    }
    return true
  }

  const handleCreateQueue = async () => {
    if (!validateForm()) return
    setLoading(true)
    try {
      const userInfo = Taro.getStorageSync('userInfo')
      const result = await queueApi.create({
        storeId: store?.id || 1,
        queueType: tableType,
        memberId: userInfo?.id,
        memberName: memberName.trim(),
        memberPhone: memberPhone.trim(),
        peopleCount
      })
      setSuccessQueue(result)
      setShowSuccessPopup(true)
      await loadMyQueues()
    } catch (e: any) {
      Taro.showToast({ title: e.message || '取号失败', icon: 'none' })
    } finally {
      setLoading(false)
    }
  }

  const handleCancelQueue = (queue: QueueItem) => {
    Dialog.show({
      title: '确认取消',
      content: '确定要取消排队吗？',
      okText: '确定取消',
      cancelText: '再想想',
      onOk: async () => {
        try {
          await queueApi.cancel({ queueId: queue.id, reason: '用户取消' })
          Taro.showToast({ title: '已取消排队', icon: 'success' })
          loadMyQueues()
        } catch (e: any) {
          Taro.showToast({ title: e.message || '取消失败', icon: 'none' })
        }
      }
    })
  }

  const handleRequeue = () => {
    setTabValue(0)
    setShowSuccessPopup(false)
  }

  const handlePreOrder = () => {
    const activeQueue = myQueues.find(q => q.status === 0 || q.status === 1)
    if (!activeQueue) return
    Taro.showToast({ title: '预点餐功能开发中', icon: 'none' })
  }

  const formatWaitedTime = (seconds: number): string => {
    const mins = Math.floor(seconds / 60)
    const secs = seconds % 60
    if (mins > 60) {
      const hours = Math.floor(mins / 60)
      const remainMins = mins % 60
      return `${hours}小时${remainMins}分钟`
    }
    return `${mins}分${secs}秒`
  }

  const getEstimatedWaitTime = (): string => {
    const count = waitingCount[tableType as keyof typeof waitingCount]
    const minutes = count * 8
    if (minutes === 0) return '预计无需等待'
    if (minutes < 60) return `预计等待约 ${minutes} 分钟`
    const hours = Math.floor(minutes / 60)
    const remainMins = minutes % 60
    return `预计等待约 ${hours}小时${remainMins}分钟`
  }

  const handleTabChange = (value: number) => {
    setTabValue(value)
    if (value === 1) {
      loadMyQueues()
    }
  }

  const activeQueue = myQueues.find(q => q.status === 0 || q.status === 1)
  const historicalQueues = myQueues.filter(q => q.status !== 0 && q.status !== 1)

  if (!isLogin()) {
    return (
      <View className={styles.empty}>
        <Text className={styles.emptyText}>请先登录</Text>
      </View>
    )
  }

  return (
    <View className={styles.container}>
      <View className={styles.storeHeader}>
        <Text className={styles.storeName}>{store?.name || '加载中...'}</Text>
        <Text className={styles.storeAddress}>{store?.address || ''}</Text>
        <View className={styles.waitingStats}>
          <View className={styles.statItem}>
            <Text className={styles.statNumber}>{waitingCount.small}</Text>
            <Text className={styles.statLabel}>小桌等待</Text>
          </View>
          <View className={styles.statItem}>
            <Text className={styles.statNumber}>{waitingCount.medium}</Text>
            <Text className={styles.statLabel}>中桌等待</Text>
          </View>
          <View className={styles.statItem}>
            <Text className={styles.statNumber}>{waitingCount.large}</Text>
            <Text className={styles.statLabel}>大桌等待</Text>
          </View>
        </View>
      </View>

      <Tabs
        value={tabValue}
        onChange={handleTabChange}
        tabTitleStyle={{ fontSize: '28rpx' }}
        className={styles.tabsWrapper}
      >
        <TabPane key={0} title='取号' value={0}>
          <ScrollView scrollY>
            <View className={styles.formSection}>
              <Text className={styles.sectionTitle}>选择桌型</Text>
              <View className={styles.tableTypeGrid}>
                {TABLE_TYPES.map(type => (
                  <View
                    key={type.value}
                    className={classNames(styles.tableTypeItem, {
                      [styles.tableTypeActive]: tableType === type.value
                    })}
                    onClick={() => handleTableTypeChange(type.value)}
                  >
                    <Text className={styles.tableTypeName}>{type.name}</Text>
                    <Text className={styles.tableTypeDesc}>{type.desc}</Text>
                  </View>
                ))}
              </View>
            </View>

            <View className={styles.formSection}>
              <Text className={styles.sectionTitle}>用餐人数</Text>
              <View className={styles.stepperRow}>
                <Text className={styles.stepperLabel}>人数</Text>
                <View style={{ display: 'flex', alignItems: 'center', gap: '24rpx' }}>
                  <View
                    style={{
                      width: '60rpx',
                      height: '60rpx',
                      display: 'flex',
                      alignItems: 'center',
                      justifyContent: 'center',
                      background: '#f5f7fa',
                      borderRadius: '50%',
                      fontSize: '36rpx',
                      color: '#666'
                    }}
                    onClick={() => handlePeopleCountChange(-1)}
                  >
                    -
                  </View>
                  <Text style={{ fontSize: '36rpx', fontWeight: '600', minWidth: '60rpx', textAlign: 'center' }}>
                    {peopleCount}
                  </Text>
                  <View
                    style={{
                      width: '60rpx',
                      height: '60rpx',
                      display: 'flex',
                      alignItems: 'center',
                      justifyContent: 'center',
                      background: '#f5f7fa',
                      borderRadius: '50%',
                      fontSize: '36rpx',
                      color: '#666'
                    }}
                    onClick={() => handlePeopleCountChange(1)}
                  >
                    +
                  </View>
                </View>
              </View>
            </View>

            <View className={styles.formSection}>
              <Text className={styles.sectionTitle}>联系信息</Text>
              <View className={styles.formItem}>
                <Text className={styles.formLabel}>姓名</Text>
                <Input
                  className={styles.formInput}
                  placeholder='请输入姓名'
                  value={memberName}
                  onInput={(e: any) => setMemberName(e.detail.value)}
                  maxlength={20}
                />
              </View>
              <View className={styles.formItem}>
                <Text className={styles.formLabel}>手机号</Text>
                <Input
                  className={styles.formInput}
                  type='number'
                  placeholder='请输入手机号'
                  value={memberPhone}
                  onInput={(e: any) => setMemberPhone(e.detail.value)}
                  maxlength={11}
                />
              </View>
            </View>

            <View
              className={classNames(styles.submitBtn, {
                [styles.submitBtnDisabled]: loading
              })}
              onClick={!loading ? handleCreateQueue : undefined}
            >
              {loading ? (
                <Loading type='spinner' size='24px' color='#fff' />
              ) : (
                <Text className={styles.submitBtnText}>立即取号</Text>
              )}
            </View>
            <Text className={styles.estimatedTime}>{getEstimatedWaitTime()}</Text>
          </ScrollView>
        </TabPane>

        <TabPane key={1} title='我的排队' value={1}>
          <ScrollView scrollY className={styles.myQueueSection}>
            {loading && myQueues.length === 0 ? (
              <View className={styles.loadingWrapper}>
                <Loading type='spinner' size='24px' />
                <Text className={styles.loadingText}>加载中...</Text>
              </View>
            ) : myQueues.length === 0 ? (
              <View className={styles.emptyWrapper}>
                <Text className={styles.emptyIcon}>📋</Text>
                <Text className={styles.emptyText}>暂无排队记录</Text>
                <View className={styles.emptyBtn} onClick={() => setTabValue(0)}>
                  <Text>去取号</Text>
                </View>
              </View>
            ) : (
              <>
                {activeQueue && (
                  <View className={styles.queueCard} key={activeQueue.id}>
                    <View className={styles.queueNumberWrapper}>
                      <View className={classNames(styles.statusTag, STATUS_MAP[activeQueue.status]?.className)}>
                        {STATUS_MAP[activeQueue.status]?.text}
                      </View>
                      <Text className={styles.queueNumberLabel}>您的排队号</Text>
                      <Text className={styles.queueNumber}>{activeQueue.queueNumber}</Text>
                    </View>
                    <View className={styles.queueInfoGrid}>
                      <View className={styles.queueInfoItem}>
                        <Text className={classNames(styles.queueInfoValue, styles.aheadCount)}>
                          {activeQueue.aheadCount}
                        </Text>
                        <Text className={styles.queueInfoLabel}>前面还有(桌)</Text>
                      </View>
                      <View className={styles.queueInfoItem}>
                        <Text className={classNames(styles.queueInfoValue, styles.waitTime)}>
                          {formatWaitedTime(waitedTime)}
                        </Text>
                        <Text className={styles.queueInfoLabel}>已等待</Text>
                      </View>
                    </View>
                    <View className={styles.queueDetail}>
                      <View className={styles.detailRow}>
                        <Text className={styles.detailLabel}>桌型</Text>
                        <Text className={styles.detailValue}>
                          {TABLE_TYPES.find(t => t.value === activeQueue.queueType)?.name || '-'}
                        </Text>
                      </View>
                      <View className={styles.detailRow}>
                        <Text className={styles.detailLabel}>人数</Text>
                        <Text className={styles.detailValue}>{activeQueue.peopleCount}人</Text>
                      </View>
                      <View className={styles.detailRow}>
                        <Text className={styles.detailLabel}>取号时间</Text>
                        <Text className={styles.detailValue}>
                          {dayjs(activeQueue.createdAt).format('YYYY-MM-DD HH:mm')}
                        </Text>
                      </View>
                      {activeQueue.tableNo && (
                        <View className={styles.detailRow}>
                          <Text className={styles.detailLabel}>安排桌号</Text>
                          <Text className={styles.detailValue}>{activeQueue.tableNo}</Text>
                        </View>
                      )}
                    </View>
                    <View className={styles.queueActions}>
                      {activeQueue.status === 0 && (
                        <View
                          className={classNames(styles.actionBtn, styles.actionBtnDanger)}
                          onClick={() => handleCancelQueue(activeQueue)}
                        >
                          <Text>取消排队</Text>
                        </View>
                      )}
                      {activeQueue.status === 0 && (
                        <View
                          className={classNames(styles.actionBtn, styles.actionBtnWarning)}
                          onClick={handlePreOrder}
                        >
                          <Text>预点餐</Text>
                        </View>
                      )}
                      <View
                        className={classNames(styles.actionBtn, styles.actionBtnPrimary)}
                        onClick={handleRequeue}
                      >
                        <Text>重新取号</Text>
                      </View>
                    </View>
                  </View>
                )}

                {historicalQueues.map(queue => (
                  <View className={styles.queueCard} key={queue.id}>
                    <View style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '16rpx' }}>
                      <Text style={{ fontSize: '32rpx', fontWeight: '600' }}>{queue.queueNumber}</Text>
                      <View className={classNames(styles.statusTag, STATUS_MAP[queue.status]?.className)}>
                        {STATUS_MAP[queue.status]?.text}
                      </View>
                    </View>
                    <View className={styles.queueDetail}>
                      <View className={styles.detailRow}>
                        <Text className={styles.detailLabel}>桌型</Text>
                        <Text className={styles.detailValue}>
                          {TABLE_TYPES.find(t => t.value === queue.queueType)?.name || '-'}
                        </Text>
                      </View>
                      <View className={styles.detailRow}>
                        <Text className={styles.detailLabel}>人数</Text>
                        <Text className={styles.detailValue}>{queue.peopleCount}人</Text>
                      </View>
                      <View className={styles.detailRow}>
                        <Text className={styles.detailLabel}>取号时间</Text>
                        <Text className={styles.detailValue}>
                          {dayjs(queue.createdAt).format('YYYY-MM-DD HH:mm')}
                        </Text>
                      </View>
                    </View>
                  </View>
                ))}
              </>
            )}
          </ScrollView>
        </TabPane>
      </Tabs>

      {showSuccessPopup && successQueue && (
        <View className={styles.successPopup} onClick={() => setShowSuccessPopup(false)}>
          <View className={styles.successContent} onClick={(e) => e.stopPropagation()}>
            <View className={styles.successIcon}>
              <Text className={styles.successIconText}>✓</Text>
            </View>
            <Text className={styles.successTitle}>取号成功</Text>
            <Text className={styles.successQueueNumber}>{successQueue.queueNumber}</Text>
            <View className={styles.successInfo}>
              <View className={styles.successInfoItem}>
                <Text className={styles.successInfoValue}>{successQueue.aheadCount}</Text>
                <Text className={styles.successInfoLabel}>前面桌数</Text>
              </View>
              <View className={styles.successInfoItem}>
                <Text className={styles.successInfoValue}>
                  {TABLE_TYPES.find(t => t.value === successQueue.queueType)?.name}
                </Text>
                <Text className={styles.successInfoLabel}>桌型</Text>
              </View>
              <View className={styles.successInfoItem}>
                <Text className={styles.successInfoValue}>
                  {waitingCount[successQueue.queueType as keyof typeof waitingCount] * 8}分钟
                </Text>
                <Text className={styles.successInfoLabel}>预计等待</Text>
              </View>
            </View>
            <View
              className={styles.successBtn}
              onClick={() => {
                setShowSuccessPopup(false)
                setTabValue(1)
              }}
            >
              <Text>查看我的排队</Text>
            </View>
          </View>
        </View>
      )}

      {showCallingAlert && callingQueue && (
        <View className={styles.callingAlert}>
          <Text className={styles.callingText}>🎉 叫号通知 🎉</Text>
          <Text className={classNames(styles.callingText, styles.callingNumber)}>{callingQueue.queueNumber}</Text>
          <Text className={classNames(styles.callingText, styles.calledTable)}>
            请到 {callingQueue.tableNo || '前台'} 入座
          </Text>
        </View>
      )}
    </View>
  )
}

export default QueuePage
