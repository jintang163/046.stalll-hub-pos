import React, { useState, useEffect } from 'react'
import { View, Text, Image, ScrollView } from '@tarojs/components'
import Taro, { useRouter } from '@tarojs/taro'
import { Loading, Swiper, SwiperItem, Button } from '@nutui/nutui-react-taro'
import { getProductDetail } from '../../services/product'
import type { Product, SKU, AttributeValue } from '../../services/product'
import { useCartStore } from '../../store/cart'
import styles from './detail.module.scss'

const ProductDetail: React.FC = () => {
  const router = useRouter()
  const productId = parseInt(router.params.id as string)
  
  const [product, setProduct] = useState<Product | null>(null)
  const [loading, setLoading] = useState(true)
  const [selectedSku, setSelectedSku] = useState<SKU | null>(null)
  const [selectedAttrs, setSelectedAttrs] = useState<Map<number, { id: number; value: string; price: number }>>(new Map())
  const [quantity, setQuantity] = useState(1)
  
  const addItem = useCartStore(state => state.addItem)
  const itemCount = useCartStore(state => state.itemCount())
  const totalAmount = useCartStore(state => state.total())

  useEffect(() => {
    loadProduct()
  }, [productId])

  const loadProduct = async () => {
    setLoading(true)
    try {
      const data = await getProductDetail(productId)
      setProduct(data)
      const availableSkus = data.skus.filter(s => s.status === 1)
      if (availableSkus.length === 1) {
        setSelectedSku(availableSkus[0])
      }
    } catch {
      Taro.showToast({ title: '加载失败', icon: 'none' })
    } finally {
      setLoading(false)
    }
  }

  const handleSkuSelect = (sku: SKU) => {
    if (sku.status !== 1 || sku.stock <= 0) return
    setSelectedSku(sku)
  }

  const handleAttrSelect = (attrId: number, attrName: string, value: AttributeValue) => {
    if (value.status !== 1 || value.stock <= 0) return
    
    const newAttrs = new Map(selectedAttrs)
    const existing = newAttrs.get(attrId)
    
    if (existing && existing.id === value.id) {
      newAttrs.delete(attrId)
    } else {
      newAttrs.set(attrId, {
        id: value.id,
        value: value.value,
        price: value.extra_price || 0
      })
    }
    setSelectedAttrs(newAttrs)
  }

  const currentPrice = (() => {
    if (!selectedSku) return 0
    let price = selectedSku.price || 0
    selectedAttrs.forEach(attr => {
      price += attr.price
    })
    return price
  })()

  const canAddToCart = (() => {
    if (!product || !selectedSku) return false
    
    const requiredAttrs = product.attributes.filter(a => a.status === 1)
    for (const attr of requiredAttrs) {
      if (!selectedAttrs.has(attr.id)) return false
    }
    return true
  })()

  const handleAddToCart = () => {
    if (!product || !selectedSku) return
    
    const attrs: { attr_id: number; attr_name: string; value: AttributeValue }[] = []
    selectedAttrs.forEach((val, attrId) => {
      const attr = product.attributes.find(a => a.id === attrId)
      const value = attr?.values.find(v => v.id === val.id)
      if (attr && value) {
        attrs.push({ attr_id: attrId, attr_name: attr.name, value })
      }
    })
    
    addItem(product, selectedSku, attrs, quantity)
    Taro.showToast({ title: '已加入购物车', icon: 'success' })
  }

  const handleBuyNow = () => {
    handleAddToCart()
    setTimeout(() => {
      Taro.navigateTo({ url: '/pages/order/confirm' })
    }, 500)
  }

  const handleCartClick = () => {
    Taro.switchTab({ url: '/pages/cart/index' })
  }

  if (loading) {
    return (
      <View className={styles.loadingPage}>
        <Loading type='spinner' />
        <Text className={styles.loadingText}>加载中...</Text>
      </View>
    )
  }

  if (!product) {
    return (
      <View className={styles.error}>
        <Text>商品不存在</Text>
      </View>
    )
  }

  return (
    <View className={styles.container}>
      <ScrollView scrollY className={styles.scrollContent}>
        <Swiper className={styles.banner} autoPlay>
          <SwiperItem>
            <Image
              className={styles.bannerImage}
              src={product.image || 'https://trae-api-cn.mchost.guru/api/ide/v1/text_to_image?prompt=delicious%20food%20dish&image_size=landscape_16_9'}
            />
          </SwiperItem>
        </Swiper>

        <View className={styles.info}>
          <View className={styles.priceRow}>
            <View className={styles.price}>
              <Text className={styles.priceSymbol}>¥</Text>
              <Text className={styles.priceValue}>{currentPrice.toFixed(2)}</Text>
            </View>
            {selectedSku && (
              <Text className={styles.stock}>库存：{selectedSku.stock} 份</Text>
            )}
          </View>
          <Text className={styles.name}>{product.name}</Text>
          <Text className={styles.desc}>{product.description}</Text>
          <View className={styles.tags}>
            {product.is_hot && <View className={styles.hotTag}>热销</View>}
            {product.is_recommend && <View className={styles.recommendTag}>推荐</View>}
          </View>
        </View>

        <View className={styles.section}>
          <Text className={styles.sectionTitle}>规格</Text>
          <View className={styles.options}>
            {product.skus.map(sku => (
              <View
                key={sku.id}
                className={`${styles.option} ${selectedSku?.id === sku.id ? styles.optionActive : ''} ${sku.status !== 1 || sku.stock <= 0 ? styles.optionDisabled : ''}`}
                onClick={() => handleSkuSelect(sku)}
              >
                <Text>{sku.spec_name}</Text>
                <Text className={styles.optionPrice}>¥{sku.price.toFixed(2)}</Text>
              </View>
            ))}
          </View>
        </View>

        {product.attributes.map(attr => (
          attr.status === 1 && (
            <View key={attr.id} className={styles.section}>
              <Text className={styles.sectionTitle}>{attr.name}</Text>
              <View className={styles.options}>
                {attr.values.map(val => (
                  <View
                    key={val.id}
                    className={`${styles.option} ${selectedAttrs.get(attr.id)?.id === val.id ? styles.optionActive : ''} ${val.status !== 1 || val.stock <= 0 ? styles.optionDisabled : ''}`}
                    onClick={() => handleAttrSelect(attr.id, attr.name, val)}
                  >
                    <Text>{val.value}</Text>
                    {val.extra_price > 0 && (
                      <Text className={styles.optionPrice}>+¥{val.extra_price.toFixed(2)}</Text>
                    )}
                  </View>
                ))}
              </View>
            </View>
          )
        ))}

        <View className={styles.section}>
          <Text className={styles.sectionTitle}>数量</Text>
          <View className={styles.quantitySelector}>
            <View
              className={styles.quantityBtn}
              onClick={() => setQuantity(Math.max(1, quantity - 1))}
            >
              <Text>−</Text>
            </View>
            <Text className={styles.quantityValue}>{quantity}</Text>
            <View
              className={styles.quantityBtn}
              onClick={() => {
                if (selectedSku && quantity < selectedSku.stock) {
                  setQuantity(quantity + 1)
                }
              }}
            >
              <Text>+</Text>
            </View>
          </View>
        </View>
      </ScrollView>

      <View className={styles.footer}>
        <View className={styles.cartBtn} onClick={handleCartClick}>
          <Text className={styles.cartIcon}>🛒</Text>
          {itemCount > 0 && <View className={styles.cartBadge}>{itemCount}</View>}
          <Text className={styles.cartText}>购物车</Text>
        </View>
        <View
          className={`${styles.addBtn} ${!canAddToCart ? styles.btnDisabled : ''}`}
          onClick={handleAddToCart}
        >
          <Text className={styles.btnText}>加入购物车</Text>
        </View>
        <View
          className={`${styles.buyBtn} ${!canAddToCart ? styles.btnDisabled : ''}`}
          onClick={handleBuyNow}
        >
          <Text className={styles.btnText}>立即购买</Text>
        </View>
      </View>
    </View>
  )
}

export default ProductDetail
