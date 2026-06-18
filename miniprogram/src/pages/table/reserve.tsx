import React, { useState, useEffect, useMemo } from 'react'
import { View, Text, ScrollView, Input, Textarea } from '@tarojs/components'
import Taro, { useDidShow } from '@tarojs/taro'
import { Tabs, TabPane, Loading, Dialog } from '@nutui/nutui-react-taro'
import dayjs from 'dayjs'
import { useAppStore } from '../../store/app'
import { reservationApi, TimeSlot, ReservationItem } from '../../services/table'
import { isLogin, loginByCode } from '../../services/auth'
import styles from './reserve.module.scss'

const weekMap = ['周日', '周一', '周二', '周三', '周四', '周五', '周六']

const statusMap: Record<number, { text: string; className: string }> = {
  0: { text: '待确认', className: styles.statusPending },
  1: { text: '已确认', className: styles.statusConfirmed },
  2: { text: '已取消', className: styles.statusCancelled },
  3: { text: '已完成', className: styles.statusConfirmed }
}

const tabList = [
  { value: 0, title: '预约订桌' },
  { value: 1, title: '我的预约' }
]

interface DateItem {
  date: string
  week: string
  day: string
  isToday: boolean
}

const TableReserve: React.FC = () => {
  const currentStore = useAppStore(state => state.currentStore)
  const user = useAppStore(state => state.user)
  const setUser = useAppStore(state => state.setUser)

  const [tabValue, setTabValue] = useState(0)
  const [loading, setLoading] = useState(false)
  const [submitting, setSubmitting] = useState(false)
  const [success, setSuccess] = useState(false)
  const [createdReservation, setCreatedReservation] = useState<ReservationItem | null>(null)

  const [dates, setDates] = useState<DateItem[]>([])
  const [selectedDate, setSelectedDate] = useState('')
  const [timeSlots, setTimeSlots] = useState<TimeSlot[]>([])
  const [selectedTime, setSelectedTime] = useState('')
  const [peopleCount, setPeopleCount] = useState(2)
  const [name, setName] = useState('')
  const [phone, setPhone] = useState('')
  const [remark, setRemark] = useState('')

  const [myReservations, setMyReservations] = useState<ReservationItem[]>([])
  const [listLoading, setListLoading] = useState(false)

  useEffect(() => {
    initDates()
  }, [])

  useEffect(() => {
    if (selectedDate && currentStore) {
      loadTimeSlots()
    }
  }, [selectedDate, peopleCount, currentStore])

  useDidShow(() => {
    if (tabValue === 1 && isLogin()) {
      loadMyReservations()
    }
  })

  const initDates = () => {
    const list: DateItem[] = []
    const today = dayjs()
    for (let i = 0; i < 7; i++) {
      const d = today.add(i, 'day')
      list.push({
        date: d.format('YYYY-MM-DD'),
        week: i === 0 ? '今天' : weekMap[d.day()],
        day: d.format('MM/DD'),
        isToday: i === 0
      })
    }
    setDates(list)
    setSelectedDate(list[0].date)
  }

  const loadTimeSlots = async () => {
    if (!currentStore) return
    setLoading(true)
    try {
      const list = await reservationApi.getTimeSlots({
        storeId: currentStore.id,
        reserveDate: selectedDate,
        peopleCount
      })
      setTimeSlots(list)
    } catch (e: any) {
      Taro.showToast({ title: e.message || '获取时段失败', icon: 'none' })
    } finally {
      setLoading(false)
    }
  }

  const loadMyReservations = async () => {
    if (!isLogin()) return
    setListLoading(true)
    try {
      const result: any = await reservationApi.list({
        pageNum: 1,
        pageSize: 20
      })
      setMyReservations(result.list || result || [])
    } catch (e: any) {
      Taro.showToast({ title: e.message || '获取预约列表失败', icon: 'none' })
    } finally {
      setListLoading(false)
    }
  }

  const handleDateSelect = (date: string) => {
    setSelectedDate(date)
    setSelectedTime('')
  }

  const handleTimeSelect = (slot: TimeSlot) => {
    if (slot.available <= 0) return
    setSelectedTime(slot.time)
  }

  const handlePeopleChange = (delta: number) => {
    const newValue = peopleCount + delta
    if (newValue >= 1 && newValue <= 10) {
      setPeopleCount(newValue)
      setSelectedTime('')
    }
  }

  const recommendedTable = useMemo(() => {
    if (peopleCount <= 2) {
      return { type: '小桌', desc: '适合1-2人用餐，安静舒适' }
    } else if (peopleCount <= 4) {
      return { type: '中桌', desc: '适合3-4人用餐，朋友小聚首选' }
    } else if (peopleCount <= 6) {
      return { type: '大桌', desc: '适合5-6人用餐，家庭聚会推荐' }
    } else {
      return { type: '包厢', desc: '适合7-10人用餐，独立私密空间' }
    }
  }, [peopleCount])

  const handleWxLogin = async () => {
    try {
      const res = await Taro.login()
      if (res.code) {
        const result = await loginByCode(res.code)
        Taro.setStorageSync('token', result.token)
        Taro.setStorageSync('userInfo', result.user)
        setUser(result.user)
        return true
      }
    } catch (e: any) {
      Taro.showToast({ title: '登录失败，请重试', icon: 'none' })
    }
    return false
  }

  const validateForm = (): boolean => {
    if (!currentStore) {
      Taro.showToast({ title: '请先选择门店', icon: 'none' })
      return false
    }
    if (!selectedDate) {
      Taro.showToast({ title: '请选择预约日期', icon: 'none' })
      return false
    }
    if (!selectedTime) {
      Taro.showToast({ title: '请选择预约时段', icon: 'none' })
      return false
    }
    if (!name.trim()) {
      Taro.showToast({ title: '请输入联系人姓名', icon: 'none' })
      return false
    }
    if (!phone.trim()) {
      Taro.showToast({ title: '请输入联系电话', icon: 'none' })
      return false
    }
    if (!/^1[3-9]\d{9}$/.test(phone.trim())) {
      Taro.showToast({ title: '请输入正确的手机号', icon: 'none' })
      return false
    }
    return true
  }

  const handleSubmit = async () => {
    if (!validateForm()) return

    if (!isLogin()) {
      const loginSuccess = await handleWxLogin()
      if (!loginSuccess) return
    }

    setSubmitting(true)
    try {
      const reservation = await reservationApi.create({
        storeId: currentStore!.id,
        memberId: user?.id,
        memberName: name.trim(),
        memberPhone: phone.trim(),
        reserveDate: selectedDate,
        reserveTime: selectedTime,
        peopleCount,
        remark: remark.trim() || undefined,
        source: 'miniprogram'
      })
      setCreatedReservation(reservation)
      setSuccess(true)
    } catch (e: any) {
      Taro.showToast({ title: e.message || '预约失败，请重试', icon: 'none' })
    } finally {
      setSubmitting(false)
    }
  }

  const handleStoreSelect = () => {
    Taro.navigateTo({ url: '/pages/store/select' })
  }

  const handleTabChange = async (value: number) => {
    setTabValue(value)
    if (value === 1) {
      if (!isLogin()) {
        const loginSuccess = await handleWxLogin()
        if (!loginSuccess) {
          setTabValue(0)
          return
        }
      }
      loadMyReservations()
    }
  }

  const handleViewDetail = (id: number) => {
    Taro.showToast({ title: '查看详情功能开发中', icon: 'none' })
  }

  const handleCancel = (id: number) => {
    Dialog.show({
      title: '确认取消',
      content: '确定要取消该预约吗？',
      okText: '确定取消',
      cancelText: '再想想',
      onOk: async () => {
        try {
          await reservationApi.cancel(id)
          Taro.showToast({ title: '已取消预约', icon: 'success' })
          loadMyReservations()
        } catch (e: any) {
          Taro.showToast({ title: e.message || '取消失败', icon: 'none' })
        }
      }
    })
  }

  const handleBackToHome = () => {
    Taro.switchTab({ url: '/pages/index/index' })
  }

  const handleViewMyReservation = () => {
    setSuccess(false)
    setTabValue(1)
    loadMyReservations()
  }

  const handleMakeAnother = () => {
    setSuccess(false)
    setCreatedReservation(null)
    setSelectedTime('')
    setName('')
    setPhone('')
    setRemark('')
    setTabValue(0)
  }

  if (success && createdReservation) {
    return (
      <View className={styles.successPage}>
        <View className={styles.successIcon}>✓</View>
        <Text className={styles.successTitle}>预约成功</Text>
        <Text className={styles.successSubtitle}>我们已为您保留桌位，请按时到店</Text>

        <View className={styles.reservationDetail}>
          <View className={styles.detailRow}>
            <Text className={styles.detailLabel}>门店</Text>
            <Text className={styles.detailValue}>{currentStore?.name}</Text>
          </View>
          <View className={styles.detailRow}>
            <Text className={styles.detailLabel}>预约日期</Text>
            <Text className={styles.detailValue}>{createdReservation.reserveDate}</Text>
          </View>
          <View className={styles.detailRow}>
            <Text className={styles.detailLabel}>预约时间</Text>
            <Text className={styles.detailValue}>{createdReservation.reserveTime}</Text>
          </View>
          <View className={styles.detailRow}>
            <Text className={styles.detailLabel}>用餐人数</Text>
            <Text className={styles.detailValue}>{createdReservation.peopleCount}人</Text>
          </View>
          <View className={styles.detailRow}>
            <Text className={styles.detailLabel}>联系人</Text>
            <Text className={styles.detailValue}>{createdReservation.memberName}</Text>
          </View>
          <View className={styles.detailRow}>
            <Text className={styles.detailLabel}>联系电话</Text>
            <Text className={styles.detailValue}>{createdReservation.memberPhone}</Text>
          </View>
          {createdReservation.remark && (
            <View className={styles.detailRow}>
              <Text className={styles.detailLabel}>备注</Text>
              <Text className={styles.detailValue}>{createdReservation.remark}</Text>
            </View>
          )}
        </View>

        <View className={styles.successActions}>
          <View className={styles.actionBtn} onClick={handleBackToHome}>
            <Text className={styles.actionBtnText}>返回首页</Text>
          </View>
          <View className={`${styles.actionBtn} ${styles.actionBtnPrimary}`} onClick={handleViewMyReservation}>
            <Text className={styles.actionBtnText}>查看预约</Text>
          </View>
        </View>
      </View>
    )
  }

  return (
    <View className={styles.container}>
      <Tabs
        value={tabValue}
        onChange={handleTabChange}
        tabTitleStyle={{ fontSize: '28rpx' }}
        className={styles.tabs}
      >
        <TabPane key={0} title={tabList[0].title} value={0}>
          <ScrollView scrollY className={styles.scrollContent}>
            {currentStore && (
              <View className={styles.section}>
                <View className={styles.storeSelector} onClick={handleStoreSelect}>
                  <View className={styles.storeInfo}>
                    <Text className={styles.storeName}>{currentStore.name}</Text>
                    <Text className={styles.storeAddress}>{currentStore.address}</Text>
                  </View>
                  <Text className={styles.storeArrow}>▶</Text>
                </View>
              </View>
            )}

            <View className={styles.section}>
              <View className={styles.sectionHeader}>
                <Text className={styles.sectionIcon}>📅</Text>
                <Text className={styles.sectionTitle}>选择日期</Text>
              </View>
              <ScrollView scrollX className={styles.dateList}>
                {dates.map(item => (
                  <View
                    key={item.date}
                    className={`${styles.dateItem} ${selectedDate === item.date ? styles.dateActive : ''}`}
                    onClick={() => handleDateSelect(item.date)}
                  >
                    <Text className={styles.dateWeek}>{item.week}</Text>
                    <Text className={styles.dateDay}>{item.day}</Text>
                  </View>
                ))}
              </ScrollView>
            </View>

            <View className={styles.section}>
              <View className={styles.sectionHeader}>
                <Text className={styles.sectionIcon}>⏰</Text>
                <Text className={styles.sectionTitle}>选择时段</Text>
              </View>
              {loading ? (
                <View style={{ padding: '40rpx', textAlign: 'center' }}>
                  <Loading type='spinner' size='16px' />
                </View>
              ) : (
                <View className={styles.timeSlots}>
                  {timeSlots.map(slot => {
                    const isAvailable = slot.available > 0
                    const isSelected = selectedTime === slot.time
                    let slotClass = styles.timeSlot
                    if (isSelected) {
                      slotClass += ` ${styles.timeSlotActive}`
                    } else if (isAvailable) {
                      slotClass += ` ${styles.timeSlotAvailable}`
                    } else {
                      slotClass += ` ${styles.timeSlotFull}`
                    }
                    return (
                      <View
                        key={slot.time}
                        className={slotClass}
                        onClick={() => handleTimeSelect(slot)}
                      >
                        <Text className={styles.timeText}>{slot.time}</Text>
                        <Text className={styles.availableText}>
                          {isAvailable ? `剩${slot.available}桌` : '已满'}
                        </Text>
                      </View>
                    )
                  })}
                </View>
              )}
            </View>

            <View className={styles.section}>
              <View className={styles.sectionHeader}>
                <Text className={styles.sectionIcon}>👥</Text>
                <Text className={styles.sectionTitle}>用餐人数</Text>
              </View>
              <View className={styles.peopleSection}>
                <View className={styles.peopleRow}>
                  <Text className={styles.peopleLabel}>请选择用餐人数</Text>
                  <View className={styles.stepper}>
                    <View
                      className={`${styles.stepperBtn} ${peopleCount <= 1 ? styles.stepperBtnDisabled : ''}`}
                      onClick={() => handlePeopleChange(-1)}
                    >
                      <Text>−</Text>
                    </View>
                    <Text className={styles.stepperValue}>{peopleCount}</Text>
                    <View
                      className={`${styles.stepperBtn} ${peopleCount >= 10 ? styles.stepperBtnDisabled : ''}`}
                      onClick={() => handlePeopleChange(1)}
                    >
                      <Text>+</Text>
                    </View>
                  </View>
                </View>
              </View>
            </View>

            <View className={styles.section}>
              <View className={styles.sectionHeader}>
                <Text className={styles.sectionIcon}>🪑</Text>
                <Text className={styles.sectionTitle}>推荐桌位</Text>
              </View>
              <View className={styles.tableRecommend}>
                <View className={styles.tableCard}>
                  <Text className={styles.tableType}>{recommendedTable.type}</Text>
                  <Text className={styles.tableDesc}>{recommendedTable.desc}</Text>
                </View>
              </View>
            </View>

            <View className={styles.section}>
              <View className={styles.sectionHeader}>
                <Text className={styles.sectionIcon}>📞</Text>
                <Text className={styles.sectionTitle}>联系信息</Text>
              </View>
              <View className={styles.formSection}>
                <View className={styles.inputRow}>
                  <Text className={styles.inputLabel}>姓名</Text>
                  <Input
                    className={styles.inputField}
                    placeholder='请输入联系人姓名'
                    value={name}
                    onInput={(e: any) => setName(e.detail.value)}
                    maxlength={20}
                  />
                </View>
                <View className={styles.inputRow}>
                  <Text className={styles.inputLabel}>手机号</Text>
                  <Input
                    className={styles.inputField}
                    type='number'
                    placeholder='请输入联系电话'
                    value={phone}
                    onInput={(e: any) => setPhone(e.detail.value)}
                    maxlength={11}
                  />
                </View>
                <View className={`${styles.inputRow} ${styles.remarkRow}`}>
                  <Text className={styles.inputLabel}>备注</Text>
                  <Textarea
                    className={styles.textareaField}
                    placeholder='选填，如有特殊需求请备注'
                    value={remark}
                    onInput={(e: any) => setRemark(e.detail.value)}
                    maxlength={200}
                  />
                </View>
              </View>
            </View>
          </ScrollView>

          <View className={styles.footer}>
            <Text className={styles.rulesText}>
              * 请提前15分钟到店，超过预约时间15分钟未到店将自动取消
            </Text>
            <View className={styles.footerActions}>
              <View
                className={`${styles.submitBtn} ${submitting ? styles.btnDisabled : ''}`}
                onClick={handleSubmit}
              >
                {submitting ? (
                  <Loading type='spinner' size='16px' color='#fff' />
                ) : (
                  <Text className={styles.submitText}>提交预约</Text>
                )}
              </View>
            </View>
          </View>
        </TabPane>

        <TabPane key={1} title={tabList[1].title} value={1}>
          {listLoading && myReservations.length === 0 ? (
            <View style={{ padding: '80rpx', textAlign: 'center' }}>
              <Loading type='spinner' size='16px' />
            </View>
          ) : myReservations.length === 0 ? (
            <View className={styles.emptyReservations}>
              <Text className={styles.emptyIcon}>📋</Text>
              <Text className={styles.emptyText}>暂无预约记录</Text>
            </View>
          ) : (
            <ScrollView scrollY className={styles.scrollContent}>
              <View className={styles.reservationList}>
                {myReservations.map(item => (
                  <View key={item.id} className={styles.reservationCard}>
                    <View className={styles.cardHeader}>
                      <Text className={styles.cardStore}>{item.tableNo || '预约订桌'}</Text>
                      <Text className={`${styles.cardStatus} ${statusMap[item.status]?.className}`}>
                        {statusMap[item.status]?.text || '未知状态'}
                      </Text>
                    </View>
                    <View className={styles.cardBody}>
                      <View className={styles.cardRow}>
                        <Text className={styles.cardIcon}>📅</Text>
                        <Text className={styles.cardText}>{item.reserveDate} {item.reserveTime}</Text>
                      </View>
                      <View className={styles.cardRow}>
                        <Text className={styles.cardIcon}>👥</Text>
                        <Text className={styles.cardText}>{item.peopleCount}人</Text>
                      </View>
                      <View className={styles.cardRow}>
                        <Text className={styles.cardIcon}>👤</Text>
                        <Text className={styles.cardText}>{item.memberName} {item.memberPhone}</Text>
                      </View>
                      {item.remark && (
                        <View className={styles.cardRow}>
                          <Text className={styles.cardIcon}>📝</Text>
                          <Text className={styles.cardText}>{item.remark}</Text>
                        </View>
                      )}
                    </View>
                    <View className={styles.cardFooter}>
                      <View
                        className={`${styles.cardBtn} ${styles.cardBtnSecondary}`}
                        onClick={() => handleViewDetail(item.id)}
                      >
                        <Text>查看详情</Text>
                      </View>
                      {(item.status === 0 || item.status === 1) && (
                        <View
                          className={`${styles.cardBtn} ${styles.cardBtnPrimary}`}
                          onClick={() => handleCancel(item.id)}
                        >
                          <Text>取消预约</Text>
                        </View>
                      )}
                    </View>
                  </View>
                ))}
              </View>
            </ScrollView>
          )}
        </TabPane>
      </Tabs>
    </View>
  )
}

export default TableReserve
