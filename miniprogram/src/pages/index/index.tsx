import React, { useState, useEffect, useMemo, useCallback } from 'react'
import { View, Text, Image, ScrollView, Input } from '@tarojs/components'
import Taro, { useDidShow, useReachBottom } from '@tarojs/taro'
import { SearchBar, Loading, Badge, Popup } from '@nutui/nutui-react-taro'
import { getCategories, getProducts } from '../../services/product'
import type { Category, Product, SKU, AttributeValue } from '../../services/product'
import { getScanOrderRecommendations, type RecommendItem } from '../../services/recommend'
import { useAppStore } from '../../store/app'
import { useCartStore } from '../../store/cart'
import styles from './index.module.scss'

const Index: React.FC = () => {
  const [categories, setCategories] = useState<Category[]>([])
  const [products, setProducts] = useState<Product[]>([])
  const [activeCategory, setActiveCategory] = useState<number | null>(null)
  const [loading, setLoading] = useState(false)
  const [searchKeyword, setSearchKeyword] = useState('')
  const [page, setPage] = useState(1)
  const [hasMore, setHasMore] = useState(true)
  const [showCart, setShowCart] = useState(false)
  const [selectedProduct, setSelectedProduct] = useState<Product | null>(null)
  const [selectedSku, setSelectedSku] = useState<SKU | null>(null)
  const [selectedAttrs, setSelectedAttrs] = useState<Map<number, { id: number; value: string; price: number }>>(new Map())
  const [quantity, setQuantity] = useState(1)
  const [showSkuPopup, setShowSkuPopup] = useState(false)
  const [recommendItems, setRecommendItems] = useState<RecommendItem[]>([])
  const [recommendLoading, setRecommendLoading] = useState(false)

  const currentStore = useAppStore(state => state.currentStore)
  const tableNo = useCartStore(state => state.tableNo)
  const cartItems = useCartStore(state => state.items)
  const totalAmount = useCartStore(state => state.total())
  const itemCount = useCartStore(state => state.itemCount())
  const addItem = useCartStore(state => state.addItem)
  const initApp = useAppStore(state => state.init)

  useEffect(() => {
    initApp()
  }, [])

  const loadCategories = useCallback(async (storeId: number) => {
    try {
      const list = await getCategories(storeId)
      setCategories(list)
      if (list.length > 0 && activeCategory === null) {
        setActiveCategory(list[0].id)
      }
    } catch {}
  }, [activeCategory])

  const loadProducts = useCallback(async (storeId: number, catId: number | null, pageNum = 1) => {
    if (!storeId) return
    setLoading(true)
    try {
      const result = await getProducts(storeId, catId || undefined, pageNum, 20)
      if (pageNum === 1) {
        setProducts(result.list)
      } else {
        setProducts(prev => [...prev, ...result.list])
      }
      setHasMore(result.list.length === 20)
      setPage(pageNum)
    } finally {
      setLoading(false)
    }
  }, [])

  const loadRecommendations = useCallback(async (storeId: number, tableNumber: string) => {
    if (!storeId || !tableNumber) {
      setRecommendItems([])
      return
    }
    setRecommendLoading(true)
    try {
      const result = await getScanOrderRecommendations(storeId, tableNumber, 4)
      setRecommendItems(result.items || [])
    } catch (e) {
      setRecommendItems([])
    } finally {
      setRecommendLoading(false)
    }
  }, [])

  useDidShow(() => {
    if (currentStore) {
      loadCategories(currentStore.id)
      loadProducts(currentStore.id, activeCategory, 1)
      if (tableNo) {
        loadRecommendations(currentStore.id, tableNo)
      }
    } else {
      Taro.navigateTo({ url: '/pages/store/select' })
    }
  })

  useEffect(() => {
    if (currentStore && tableNo) {
      loadRecommendations(currentStore.id, tableNo)
    }
  }, [currentStore, tableNo, loadRecommendations])

  useEffect(() => {
    if (currentStore && activeCategory !== null) {
      loadProducts(currentStore.id, activeCategory, 1)
    }
  }, [activeCategory, currentStore])

  useReachBottom(() => {
    if (hasMore && !loading && currentStore) {
      loadProducts(currentStore.id, activeCategory, page + 1)
    }
  })

  const handleStoreClick = () => {
    Taro.navigateTo({ url: '/pages/store/select' })
  }

  const handleProductClick = (product: Product) => {
    Taro.navigateTo({ url: `/pages/product/detail?id=${product.id}` })
  }

  const handleAddClick = (e: React.MouseEvent, product: Product) => {
    e.stopPropagation()
    
    const soldOutSkus = product.skus.filter(s => s.is_sold_out)
    if (soldOutSkus.length === product.skus.length) {
      Taro.showToast({ title: '该商品已沽清', icon: 'none' })
      return
    }
    
    const availableSkus = product.skus.filter(s => s.status === 1 && !s.is_sold_out)
    const availableAttrs = product.attributes.filter(a => a.status === 1 && a.values.some(v => v.status === 1))
    
    if (availableSkus.length === 1 && availableAttrs.length === 0) {
      const attrs: { attr_id: number; attr_name: string; value: AttributeValue }[] = []
      addItem(product, availableSkus[0], attrs, 1)
      Taro.showToast({ title: '已加入购物车', icon: 'success' })
    } else {
      setSelectedProduct(product)
      setSelectedSku(null)
      setSelectedAttrs(new Map())
      setQuantity(1)
      setShowSkuPopup(true)
    }
  }

  const handleRecommendAddClick = (e: React.MouseEvent, item: RecommendItem) => {
    e.stopPropagation()
    
    const product = products.find(p => p.id === item.product_id)
    if (!product) {
      Taro.showToast({ title: '商品信息获取失败', icon: 'none' })
      return
    }
    
    const sku = product.skus.find(s => s.id === item.sku_id)
    if (!sku || sku.status !== 1 || sku.is_sold_out || sku.stock <= 0) {
      const availableSkus = product.skus.filter(s => s.status === 1 && !s.is_sold_out)
      if (availableSkus.length === 0) {
        Taro.showToast({ title: '该商品已沽清', icon: 'none' })
        return
      }
      handleAddClick(e, product)
      return
    }
    
    const availableAttrs = product.attributes.filter(a => a.status === 1 && a.values.some(v => v.status === 1))
    if (availableAttrs.length === 0) {
      const attrs: { attr_id: number; attr_name: string; value: AttributeValue }[] = []
      addItem(product, sku, attrs, 1)
      Taro.showToast({ title: '已加入购物车', icon: 'success' })
    } else {
      setSelectedProduct(product)
      setSelectedSku(sku)
      setSelectedAttrs(new Map())
      setQuantity(1)
      setShowSkuPopup(true)
    }
  }

  const getReasonTagStyle = (reasonType: string) => {
    switch (reasonType) {
      case 'table_history':
        return styles.reasonTagHistory
      case 'time_hot':
        return styles.reasonTagTime
      case 'hot':
        return styles.reasonTagHot
      default:
        return styles.reasonTagDefault
    }
  }

  const handleSkuSelect = (sku: SKU) => {
    if (sku.status !== 1 || sku.stock <= 0 || sku.is_sold_out) return
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

  const canAddToCart = useMemo(() => {
    if (!selectedProduct || !selectedSku) return false
    
    const requiredAttrs = selectedProduct.attributes.filter(a => a.status === 1)
    for (const attr of requiredAttrs) {
      if (!selectedAttrs.has(attr.id)) return false
    }
    return true
  }, [selectedProduct, selectedSku, selectedAttrs])

  const currentPrice = useMemo(() => {
    if (!selectedSku) return 0
    let price = selectedSku.price || 0
    selectedAttrs.forEach(attr => {
      price += attr.price
    })
    return price
  }, [selectedSku, selectedAttrs])

  const handleAddToCart = () => {
    if (!selectedProduct || !selectedSku) return
    
    const attrs: { attr_id: number; attr_name: string; value: AttributeValue }[] = []
    selectedAttrs.forEach((val, attrId) => {
      const attr = selectedProduct.attributes.find(a => a.id === attrId)
      const value = attr?.values.find(v => v.id === val.id)
      if (attr && value) {
        attrs.push({ attr_id: attrId, attr_name: attr.name, value })
      }
    })
    
    addItem(selectedProduct, selectedSku, attrs, quantity)
    setShowSkuPopup(false)
    Taro.showToast({ title: '已加入购物车', icon: 'success' })
  }

  const handleCartClick = () => {
    if (itemCount > 0) {
      Taro.navigateTo({ url: '/pages/cart/index' })
    }
  }

  const handleCheckout = () => {
    if (itemCount === 0) return
    Taro.navigateTo({ url: '/pages/order/confirm' })
  }

  if (!currentStore) {
    return (
      <View className={styles.loadingPage}>
        <Loading type='spinner' />
        <Text className={styles.loadingText}>加载中...</Text>
      </View>
    )
  }

  return (
    <View className={styles.container}>
      <View className={styles.header} onClick={handleStoreClick}>
        <View className={styles.storeInfo}>
          <Text className={styles.storeName}>{currentStore.name}</Text>
          <Text className={styles.storeAddress}>
            {currentStore.address} ▼
          </Text>
        </View>
      </View>

      <View className={styles.searchBar}>
        <View className={styles.searchWrap}>
          <SearchBar
            placeholder='搜索菜品'
            value={searchKeyword}
            onChange={v => setSearchKeyword(v)}
            onSearch={() => {}}
          />
        </View>
        <View className={styles.voiceBtn} onClick={() => Taro.navigateTo({ url: '/pages/voice/index' })}>
          <Text className={styles.voiceIcon}>🎤</Text>
        </View>
      </View>

      {recommendItems.length > 0 && (
        <View className={styles.recommendSection}>
          <View className={styles.recommendHeader}>
            <Text className={styles.recommendTitle}>✨ 为你推荐</Text>
            {tableNo && <Text className={styles.recommendSubtitle}>（{tableNo}号桌专属）</Text>}
          </View>
          <ScrollView scrollX className={styles.recommendScroll}>
            <View className={styles.recommendList}>
              {recommendItems.map(item => (
                <View key={item.product_id} className={styles.recommendCard}>
                  <View className={styles.recommendImageWrap}>
                    <Image
                      className={styles.recommendImage}
                      src={item.main_image || 'https://trae-api-cn.mchost.guru/api/ide/v1/text_to_image?prompt=delicious%20food%20dish&image_size=square'}
                    />
                    <View className={`${styles.reasonTag} ${getReasonTagStyle(item.reason_type)}`}>
                      <Text className={styles.reasonTagText}>{item.reason}</Text>
                    </View>
                  </View>
                  <View className={styles.recommendInfo}>
                    <Text className={styles.recommendName} numberOfLines={1}>{item.product_name}</Text>
                    <View className={styles.recommendFooter}>
                      <View className={styles.recommendPrice}>
                        <Text className={styles.priceSymbol}>¥</Text>
                        <Text className={styles.priceValue}>{item.price}</Text>
                      </View>
                      <View className={styles.recommendAddBtn} onClick={(e) => handleRecommendAddClick(e, item)}>
                        <Text className={styles.recommendAddText}>+</Text>
                      </View>
                    </View>
                  </View>
                </View>
              ))}
            </View>
          </ScrollView>
        </View>
      )}

      <View className={styles.content}>
        <ScrollView scrollY className={styles.categorySidebar}>
          {categories.map(cat => (
            <View
              key={cat.id}
              className={`${styles.categoryItem} ${activeCategory === cat.id ? styles.categoryActive : ''}`}
              onClick={() => setActiveCategory(cat.id)}
            >
              <Text className={styles.categoryName}>{cat.name}</Text>
              {activeCategory === cat.id && <View className={styles.categoryBar} />}
            </View>
          ))}
        </ScrollView>

        <ScrollView scrollY className={styles.productList}>
          {products.map(product => {
            const isAllSoldOut = product.skus.length > 0 && product.skus.every(s => s.is_sold_out)
            return (
            <View
              key={product.id}
              className={`${styles.productCard} ${isAllSoldOut ? styles.productCardSoldOut : ''}`}
              onClick={() => handleProductClick(product)}
            >
              <Image
                className={styles.productImage}
                src={product.image || 'https://trae-api-cn.mchost.guru/api/ide/v1/text_to_image?prompt=delicious%20food%20dish&image_size=square'}
              />
              <View className={styles.productInfo}>
                <View className={styles.productHeader}>
                  <Text className={styles.productName}>{product.name}</Text>
                  {product.is_hot && <View className={styles.hotTag}>热销</View>}
                  {product.is_recommend && <View className={styles.recommendTag}>推荐</View>}
                  {isAllSoldOut && <View className={styles.soldOutTag}>沽清</View>}
                </View>
                <Text className={styles.productDesc} numberOfLines={2}>{product.description}</Text>
                <View className={styles.productFooter}>
                  <View className={styles.priceInfo}>
                    <Text className={styles.priceSymbol}>¥</Text>
                    <Text className={styles.priceValue}>
                      {product.min_price?.toFixed(2)}
                      {product.max_price && product.max_price !== product.min_price && `~${product.max_price.toFixed(2)}`}
                    </Text>
                  </View>
                  <View className={styles.addButton} onClick={(e) => handleAddClick(e, product)}>
                    <Text className={styles.addButtonText}>+</Text>
                  </View>
                </View>
              </View>
            </View>
            )
          })}

          {loading && products.length > 0 && (
            <View className={styles.loadingMore}>
              <Loading type='spinner' size='16px' />
              <Text className={styles.loadingMoreText}>加载中...</Text>
            </View>
          )}

          {!loading && products.length === 0 && (
            <View className={styles.empty}>
              <Text className={styles.emptyText}>暂无商品</Text>
            </View>
          )}

          {!hasMore && products.length > 0 && (
            <View className={styles.noMore}>
              <Text className={styles.noMoreText}>— 没有更多了 —</Text>
            </View>
          )}
        </ScrollView>
      </View>

      <View className={styles.cartBar} onClick={handleCartClick}>
        <View className={styles.cartIcon}>
          <Text className={styles.cartEmoji}>🛒</Text>
          {itemCount > 0 && (
            <Badge value={itemCount} className={styles.cartBadge} />
          )}
        </View>
        <View className={styles.cartInfo}>
          <View className={styles.cartTotal}>
            <Text className={styles.totalSymbol}>¥</Text>
            <Text className={styles.totalValue}>{totalAmount.toFixed(2)}</Text>
          </View>
          <Text className={styles.cartHint}>已选 {itemCount} 件</Text>
        </View>
        <View
          className={`${styles.checkoutButton} ${itemCount === 0 ? styles.checkoutDisabled : ''}`}
          onClick={(e) => { e.stopPropagation(); handleCheckout(); }}
        >
          <Text className={styles.checkoutText}>{itemCount === 0 ? '去选购' : '去结算'}</Text>
        </View>
      </View>

      <Popup
        visible={showSkuPopup}
        position='bottom'
        round
        onClose={() => setShowSkuPopup(false)}
        style={{ height: '80vh' }}
      >
        {selectedProduct && (
          <View className={styles.skuPopup}>
            <View className={styles.skuHeader}>
              <Image
                className={styles.skuImage}
                src={selectedProduct.image || 'https://trae-api-cn.mchost.guru/api/ide/v1/text_to_image?prompt=delicious%20food%20dish&image_size=square'}
              />
              <View className={styles.skuInfo}>
                <Text className={styles.skuName}>{selectedProduct.name}</Text>
                <View className={styles.skuPrice}>
                  <Text className={styles.priceSymbol}>¥</Text>
                  <Text className={styles.priceBig}>{currentPrice.toFixed(2)}</Text>
                </View>
                {selectedSku && (
                  <Text className={styles.skuStock}>库存：{selectedSku.stock} 份</Text>
                )}
              </View>
              <View className={styles.skuClose} onClick={() => setShowSkuPopup(false)}>
                <Text>✕</Text>
              </View>
            </View>

            <ScrollView scrollY className={styles.skuContent}>
              <View className={styles.skuSection}>
                <Text className={styles.sectionTitle}>规格</Text>
                <View className={styles.skuOptions}>
                  {selectedProduct.skus.map(sku => (
                    <View
                      key={sku.id}
                      className={`${styles.skuOption} ${selectedSku?.id === sku.id ? styles.skuOptionActive : ''} ${sku.status !== 1 || sku.stock <= 0 || sku.is_sold_out ? styles.skuOptionDisabled : ''}`}
                      onClick={() => handleSkuSelect(sku)}
                    >
                      <Text>{sku.spec_name}</Text>
                      {sku.is_sold_out && <Text className={styles.skuOptionPrice}>（沽清）</Text>}
                      <Text className={styles.skuOptionPrice}>¥{sku.price.toFixed(2)}</Text>
                    </View>
                  ))}
                </View>
              </View>

              {selectedProduct.attributes.map(attr => (
                attr.status === 1 && (
                  <View key={attr.id} className={styles.skuSection}>
                    <Text className={styles.sectionTitle}>{attr.name}</Text>
                    <View className={styles.skuOptions}>
                      {attr.values.map(val => (
                        <View
                          key={val.id}
                          className={`${styles.skuOption} ${selectedAttrs.get(attr.id)?.id === val.id ? styles.skuOptionActive : ''} ${val.status !== 1 || val.stock <= 0 ? styles.skuOptionDisabled : ''}`}
                          onClick={() => handleAttrSelect(attr.id, attr.name, val)}
                        >
                          <Text>{val.value}</Text>
                          {val.extra_price > 0 && (
                            <Text className={styles.skuOptionPrice}>+¥{val.extra_price.toFixed(2)}</Text>
                          )}
                        </View>
                      ))}
                    </View>
                  </View>
                )
              ))}

              <View className={styles.skuSection}>
                <Text className={styles.sectionTitle}>数量</View>
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

            <View className={styles.skuFooter}>
              <View
                className={`${styles.skuAddButton} ${!canAddToCart ? styles.skuAddDisabled : ''}`}
                onClick={handleAddToCart}
              >
                <Text className={styles.skuAddText}>
                  {canAddToCart ? '加入购物车' : '请选择规格'}
                </Text>
              </View>
            </View>
          </View>
        )}
      </Popup>
    </View>
  )
}

export default Index
