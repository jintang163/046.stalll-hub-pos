import React, { useState } from 'react'
import { View, Text, Image, ScrollView } from '@tarojs/components'
import Taro, { useDidShow } from '@tarojs/taro'
import { Button, Cell, Dialog } from '@nutui/nutui-react-taro'
import { useAppStore } from '../../store/app'
import { isLogin, logout as authLogout, loginByCode, getUserInfo } from '../../services/auth'
import { getMyCoupons } from '../../services/coupon'
import type { MemberCoupon } from '../../services/coupon'
import styles from './index.module.scss'

const menuList = [
  { icon: '📦', title: '我的订单', path: '/pages/order/list' },
  { icon: '🎟️', title: '优惠券', path: '/pages/coupon/list' },
  { icon: '💰', title: '充值', path: '/pages/user/recharge' },
  { icon: '🏪', title: '切换门店', path: '/pages/store/select' },
  { icon: '📍', title: '收货地址', path: '/pages/address/list' },
  { icon: '💬', title: '联系客服', path: '' },
  { icon: '⚙️', title: '设置', path: '/pages/settings/index' }
]

const User: React.FC = () => {
  const user = useAppStore(state => state.user)
  const setUser = useAppStore(state => state.setUser)
  const currentStore = useAppStore(state => state.currentStore)
  
  const [coupons, setCoupons] = useState<MemberCoupon[]>([])
  const [showLogoutDialog, setShowLogoutDialog] = useState(false)
  const [loginLoading, setLoginLoading] = useState(false)

  useDidShow(() => {
    if (isLogin()) {
      loadUserData()
    }
  })

  const loadUserData = async () => {
    try {
      const [userInfo, couponList] = await Promise.all([
        getUserInfo(),
        getMyCoupons(1)
      ])
      if (userInfo) {
        Taro.setStorageSync('userInfo', userInfo)
        setUser(userInfo)
      }
      setCoupons(couponList)
    } catch {}
  }

  const handleWxLogin = async () => {
    setLoginLoading(true)
    try {
      const res = await Taro.login()
      if (res.code) {
        const result = await loginByCode(res.code)
        Taro.setStorageSync('token', result.token)
        Taro.setStorageSync('userInfo', result.user)
        setUser(result.user)
        Taro.showToast({ title: '登录成功', icon: 'success' })
        loadUserData()
      }
    } catch (e: any) {
      Taro.showToast({ title: e.message || '登录失败，请重试', icon: 'none' })
    } finally {
      setLoginLoading(false)
    }
  }

  const handlePhoneLogin = () => {
    Taro.navigateTo({ url: '/pages/user/login' })
  }

  const handleMenuClick = (item: typeof menuList[0]) => {
    if (!item.path) {
      Taro.showToast({ title: '功能开发中', icon: 'none' })
      return
    }
    if (item.path.includes('order') || item.path.includes('coupon') || item.path.includes('address')) {
      if (!isLogin()) {
        Taro.showToast({ title: '请先登录', icon: 'none' })
        return
      }
    }
    if (item.path.includes('tab')) {
      Taro.switchTab({ url: item.path })
    } else {
      Taro.navigateTo({ url: item.path })
    }
  }

  const handleLogout = () => {
    setShowLogoutDialog(true)
  }

  const confirmLogout = () => {
    authLogout()
    setUser(null)
    setCoupons([])
    setShowLogoutDialog(false)
    Taro.showToast({ title: '已退出登录', icon: 'success' })
  }

  const availableCoupons = coupons.filter(c => c.status === 1).length

  if (!isLogin() || !user) {
    return (
      <View className={styles.loginContainer}>
        <View className={styles.loginHeader}>
          <Text className={styles.loginIcon}>🍜</Text>
          <Text className={styles.loginTitle}>欢迎来到大排档</Text>
          <Text className={styles.loginDesc}>登录后享受更多会员权益</Text>
        </View>

        <View className={styles.loginButtons}>
          <Button
            className={styles.wxLoginBtn}
            type='primary'
            size='large'
            loading={loginLoading}
            onClick={handleWxLogin}
          >
            <Text className={styles.wxLoginText}>微信一键登录</Text>
          </Button>
          
          <View className={styles.phoneLoginBtn} onClick={handlePhoneLogin}>
            <Text className={styles.phoneLoginText}>手机号登录</Text>
          </View>
        </View>

        <View className={styles.loginAgreement}>
          <Text className={styles.agreementText}>
            登录即表示同意
            <Text className={styles.agreementLink}> 《用户协议》</Text>
            和
            <Text className={styles.agreementLink}> 《隐私政策》</Text>
          </Text>
        </View>
      </View>
    )
  }

  return (
    <View className={styles.container}>
      <ScrollView scrollY className={styles.scrollContent}>
        <View className={styles.userHeader}>
          <View className={styles.userInfo}>
            <Image
              className={styles.avatar}
              src={user.avatar || 'https://trae-api-cn.mchost.guru/api/ide/v1/text_to_image?prompt=user%20avatar%20portrait&image_size=square'}
            />
            <View className={styles.userInfoText}>
              <Text className={styles.nickname}>{user.nickname || '用户'}</Text>
              <View className={styles.memberInfo}>
                <Text className={styles.level}>{user.level_name}</Text>
                <Text className={styles.points}>{user.points} 积分</Text>
              </View>
            </View>
          </View>
        </View>

        <View className={styles.statsCard}>
          <View className={styles.statItem}>
            <Text className={styles.statValue}>{user.balance || '0.00'}</Text>
            <Text className={styles.statLabel}>余额(元)</Text>
          </View>
          <View className={styles.statDivider} />
          <View className={styles.statItem}>
            <Text className={styles.statValue}>{user.points}</Text>
            <Text className={styles.statLabel}>积分</Text>
          </View>
          <View className={styles.statDivider} />
          <View className={styles.statItem}>
            <Text className={styles.statValue}>{availableCoupons}</Text>
            <Text className={styles.statLabel}>可用券</Text>
          </View>
        </View>

        {currentStore && (
          <View className={styles.storeCard} onClick={() => Taro.navigateTo({ url: '/pages/store/select' })}>
            <View className={styles.storeInfo}>
              <Text className={styles.storeLabel}>当前门店</Text>
              <Text className={styles.storeName}>{currentStore.name}</Text>
              <Text className={styles.storeAddress}>{currentStore.address}</Text>
            </View>
            <Text className={styles.storeArrow}>▶</Text>
          </View>
        )}

        <View className={styles.menuList}>
          {menuList.map((item, index) => (
            <Cell
              key={index}
              onClick={() => handleMenuClick(item)}
            >
              <View className={styles.menuItem}>
                <Text className={styles.menuIcon}>{item.icon}</Text>
                <Text className={styles.menuTitle}>{item.title}</Text>
              </View>
            </Cell>
          ))}
        </View>

        <View className={styles.logoutBtn} onClick={handleLogout}>
          <Text className={styles.logoutText}>退出登录</Text>
        </View>
      </ScrollView>

      <Dialog
        visible={showLogoutDialog}
        title='确认退出'
        content='确定要退出登录吗？'
        okText='退出登录'
        cancelText='取消'
        onOk={confirmLogout}
        onCancel={() => setShowLogoutDialog(false)}
      />
    </View>
  )
}

export default User
