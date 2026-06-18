import React, { useState, useEffect } from 'react'
import { View, Text, Image } from '@tarojs/components'
import Taro, { useDidShow } from '@tarojs/taro'
import { Cell, Loading } from '@nutui/nutui-react-taro'
import { getStoreList } from '../../services/store'
import type { Store } from '../../services/store'
import { useAppStore } from '../../store/app'
import styles from './select.module.scss'

const StoreSelect: React.FC = () => {
  const [stores, setStores] = useState<Store[]>([])
  const [loading, setLoading] = useState(true)
  const setStore = useAppStore(state => state.setStore)
  const currentStore = useAppStore(state => state.currentStore)

  const loadStores = async () => {
    setLoading(true)
    try {
      const list = await getStoreList()
      setStores(list)
    } catch {
      Taro.showToast({ title: '加载门店失败', icon: 'none' })
    } finally {
      setLoading(false)
    }
  }

  const handleSelect = (store: Store) => {
    setStore(store)
    Taro.showToast({ title: `已选择：${store.name}`, icon: 'success' })
    setTimeout(() => {
      Taro.switchTab({ url: '/pages/index/index' })
    }, 1000)
  }

  useDidShow(() => {
    loadStores()
  })

  if (loading) {
    return (
      <View className={styles.loading}>
        <Loading type='spinner' />
        <Text className={styles.loadingText}>加载中...</Text>
      </View>
    )
  }

  return (
    <View className={styles.container}>
      <View className={styles.header}>
        <Text className={styles.title}>请选择门店</Text>
        <Text className={styles.subtitle}>选择您附近的门店点餐</Text>
      </View>

      <View className={styles.list}>
        {stores.map(store => (
          <Cell
            key={store.id}
            className={styles.storeItem}
            onClick={() => handleSelect(store)}
          >
            <View className={styles.storeCard}>
              <View className={styles.storeLeft}>
                <Image
                  className={styles.storeImage}
                  src={store.image || 'https://trae-api-cn.mchost.guru/api/ide/v1/text_to_image?prompt=restaurant%20store%20front%20modern&image_size=square'}
                />
              </View>
              <View className={styles.storeInfo}>
                <View className={styles.storeHeader}>
                  <Text className={styles.storeName}>{store.name}</Text>
                  {currentStore?.id === store.id && (
                    <View className={styles.currentTag}>当前</View>
                  )}
                </View>
                <Text className={styles.storeAddress}>{store.address}</Text>
                <View className={styles.storeMeta}>
                  <Text className={styles.storePhone}>{store.phone}</Text>
                  <Text className={styles.storeHours}>营业时间：{store.business_hours}</Text>
                </View>
              </View>
            </View>
          </Cell>
        ))}
      </View>

      {stores.length === 0 && (
        <View className={styles.empty}>
          <Text className={styles.emptyText}>暂无门店</Text>
        </View>
      )}
    </View>
  )
}

export default StoreSelect
