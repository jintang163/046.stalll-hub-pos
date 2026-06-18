import React, { useState } from 'react'
import { View, Text, Image, ScrollView, Checkbox } from '@tarojs/components'
import Taro from '@tarojs/taro'
import { Button, Dialog, Toast } from '@nutui/nutui-react-taro'
import { useCartStore, CartItem } from '../../store/cart'
import { useAppStore } from '../../store/app'
import styles from './index.module.scss'

const Cart: React.FC = () => {
  const items = useCartStore(state => state.items)
  const total = useCartStore(state => state.total())
  const updateQuantity = useCartStore(state => state.updateQuantity)
  const removeItem = useCartStore(state => state.removeItem)
  const clear = useCartStore(state => state.clear)
  const setTableNo = useCartStore(state => state.setTableNo)
  const setRemark = useCartStore(state => state.setRemark)
  const tableNo = useCartStore(state => state.tableNo)
  const remark = useCartStore(state => state.remark)
  const currentStore = useAppStore(state => state.currentStore)

  const [selectedItems, setSelectedItems] = useState<Set<string>>(new Set(items.map(i => i.id)))
  const [showClearDialog, setShowClearDialog] = useState(false)
  const [showTableDialog, setShowTableDialog] = useState(false)
  const [tableInput, setTableInput] = useState(tableNo)
  const [remarkInput, setRemarkInput] = useState(remark)

  const toggleSelect = (id: string) => {
    const newSelected = new Set(selectedItems)
    if (newSelected.has(id)) {
      newSelected.delete(id)
    } else {
      newSelected.add(id)
    }
    setSelectedItems(newSelected)
  }

  const toggleSelectAll = () => {
    if (selectedItems.size === items.length) {
      setSelectedItems(new Set())
    } else {
      setSelectedItems(new Set(items.map(i => i.id)))
    }
  }

  const selectedTotal = () => {
    return items
      .filter(item => selectedItems.has(item.id))
      .reduce((sum, item) => sum + item.subtotal, 0)
  }

  const selectedCount = () => {
    return items
      .filter(item => selectedItems.has(item.id))
      .reduce((sum, item) => sum + item.quantity, 0)
  }

  const handleCheckout = () => {
    if (selectedItems.size === 0) {
      Taro.showToast({ title: '请选择商品', icon: 'none' })
      return
    }
    if (!tableNo) {
      setShowTableDialog(true)
      return
    }
    
    Taro.navigateTo({ url: '/pages/order/confirm' })
  }

  const handleClear = () => {
    setShowClearDialog(true)
  }

  const confirmClear = () => {
    clear()
    setSelectedItems(new Set())
    setShowClearDialog(false)
    Taro.showToast({ title: '已清空购物车', icon: 'success' })
  }

  const confirmTable = () => {
    if (!tableInput.trim()) {
      Taro.showToast({ title: '请输入桌号', icon: 'none' })
      return
    }
    setTableNo(tableInput.trim())
    setRemark(remarkInput.trim())
    setShowTableDialog(false)
  }

  const handleEditTable = () => {
    setTableInput(tableNo)
    setRemarkInput(remark)
    setShowTableDialog(true)
  }

  if (items.length === 0) {
    return (
      <View className={styles.empty}>
        <Text className={styles.emptyIcon}>🛒</Text>
        <Text className={styles.emptyText}>购物车空空如也</Text>
        <Button
          className={styles.emptyBtn}
          type='primary'
          onClick={() => Taro.switchTab({ url: '/pages/index/index' })}
        >
          去逛逛
        </Button>
      </View>
    )
  }

  return (
    <View className={styles.container}>
      {currentStore && (
        <View className={styles.storeBar}>
          <Text className={styles.storeName}>{currentStore.name}</Text>
          <View className={styles.tableInfo} onClick={handleEditTable}>
            <Text className={styles.tableLabel}>桌号：</Text>
            <Text className={styles.tableValue}>{tableNo || '请选择'}</Text>
            <Text className={styles.tableArrow}>▶</Text>
          </View>
        </View>
      )}

      <ScrollView scrollY className={styles.cartList}>
        {items.map(item => (
          <View key={item.id} className={styles.cartItem}>
            <Checkbox
              className={styles.checkbox}
              checked={selectedItems.has(item.id)}
              onChange={() => toggleSelect(item.id)}
            />
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
                <View className={styles.quantityControl}>
                  <View
                    className={styles.quantityBtn}
                    onClick={() => updateQuantity(item.id, item.quantity - 1)}
                  >
                    <Text>−</Text>
                  </View>
                  <Text className={styles.quantityValue}>{item.quantity}</Text>
                  <View
                    className={styles.quantityBtn}
                    onClick={() => updateQuantity(item.id, item.quantity + 1)}
                  >
                    <Text>+</Text>
                  </View>
                </View>
              </View>
            </View>
            <View
              className={styles.deleteBtn}
              onClick={() => removeItem(item.id)}
            >
              <Text className={styles.deleteIcon}>🗑️</Text>
            </View>
          </View>
        ))}

        <View className={styles.clearRow} onClick={handleClear}>
          <Text className={styles.clearText}>清空购物车</Text>
        </View>
      </ScrollView>

      <View className={styles.footer}>
        <View className={styles.selectAll} onClick={toggleSelectAll}>
          <Checkbox
            checked={selectedItems.size === items.length && items.length > 0}
          />
          <Text className={styles.selectAllText}>全选</Text>
        </View>
        <View className={styles.totalInfo}>
          <Text className={styles.totalLabel}>合计：</Text>
          <Text className={styles.totalSymbol}>¥</Text>
          <Text className={styles.totalValue}>{selectedTotal().toFixed(2)}</Text>
        </View>
        <View className={styles.checkoutBtn} onClick={handleCheckout}>
          <Text className={styles.checkoutText}>
            去结算({selectedCount()})
          </Text>
        </View>
      </View>

      <Dialog
        visible={showClearDialog}
        title='确认清空'
        content='确定要清空购物车吗？'
        okText='确定'
        cancelText='取消'
        onOk={confirmClear}
        onCancel={() => setShowClearDialog(false)}
      />

      <Dialog
        visible={showTableDialog}
        title='设置桌号'
        footer={
          <View className={styles.dialogFooter}>
            <Button
              className={styles.dialogBtn}
              onClick={() => setShowTableDialog(false)}
            >
              取消
            </Button>
            <Button
              className={styles.dialogBtn}
              type='primary'
              onClick={confirmTable}
            >
              确定
            </Button>
          </View>
        }
        onClose={() => setShowTableDialog(false)}
      >
        <View className={styles.tableDialog}>
          <View className={styles.inputRow}>
            <Text className={styles.inputLabel}>桌号</Text>
            <input
              className={styles.inputField}
              placeholder='请输入桌号'
              value={tableInput}
              onInput={(e: any) => setTableInput(e.target.value)}
            />
          </View>
          <View className={styles.inputRow}>
            <Text className={styles.inputLabel}>备注</Text>
            <textarea
              className={styles.textareaField}
              placeholder='口味偏好、忌口等'
              value={remarkInput}
              onInput={(e: any) => setRemarkInput(e.target.value)}
            />
          </View>
        </View>
      </Dialog>
    </View>
  )
}

export default Cart
