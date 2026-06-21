<template>
  <view class="menu-page">
    <view class="menu-header">
      <view class="table-info">
        <text class="table-label">桌号：</text>
        <text class="table-no">{{ tableNo }}</text>
        <text class="action-text" v-if="action === 'add'">（加菜）</text>
        <text class="action-text" v-else-if="action === 'checkin'">（入座点餐）</text>
        <text class="action-text" v-else>（新点餐）</text>
      </view>
    </view>

    <view class="menu-body">
      <scroll-view scroll-y class="category-list">
        <view 
          class="category-item" 
          v-for="(cat, index) in categories" 
          :key="cat.id"
          :class="{ active: currentCategoryIndex === index }"
          @click="selectCategory(index)">
          {{ cat.name }}
        </view>
      </scroll-view>

      <scroll-view scroll-y class="product-list">
        <view class="product-item" v-for="product in categoryProducts" :key="product.id">
          <image class="product-image" :src="product.main_image || '/static/default-food.png'" mode="aspectFill" />
          <view class="product-info">
            <view class="product-name">{{ product.name }}</view>
            <view class="product-desc">{{ product.description }}</view>
            <view class="product-bottom">
              <view class="product-price">
                <text class="price-symbol">¥</text>
                <text class="price-value">{{ product.min_price }}</text>
              </view>
              <view class="cart-actions">
                <view class="cart-btn minus" v-if="getCartCount(product) > 0" @click="decreaseProduct(product)">
                  -
                </view>
                <text class="cart-count" v-if="getCartCount(product) > 0">{{ getCartCount(product) }}</text>
                <view class="cart-btn plus" @click="increaseProduct(product)">+</view>
              </view>
            </view>
          </view>
        </view>
        <view class="empty" v-if="categoryProducts.length === 0 && !loading">
          暂无菜品
        </view>
      </scroll-view>
    </view>

    <view class="menu-footer">
      <view class="cart-info" @click="toggleCartPopup">
        <view class="cart-icon-wrapper">
          <text class="cart-icon">🛒</text>
          <view class="cart-badge" v-if="cartStore.totalCount > 0">{{ cartStore.totalCount }}</view>
        </view>
        <view class="cart-amount">
          <text class="amount-symbol">¥</text>
          <text class="amount-value">{{ cartStore.totalAmount.toFixed(2) }}</text>
        </view>
      </view>
      <view class="footer-btn" :class="{ disabled: cartStore.totalCount === 0 }" @click="submitOrder">
        {{ action === 'add' ? '确认加菜' : '确认下单' }}
      </view>
    </view>

    <view class="cart-popup" v-if="showCartPopup" @click.self="showCartPopup = false">
      <view class="cart-popup-content">
        <view class="cart-popup-header">
          <text class="cart-popup-title">已选菜品</text>
          <text class="cart-clear" @click="clearCart">清空</text>
        </view>
        <scroll-view scroll-y class="cart-items-list">
          <view class="cart-item" v-for="item in cartStore.items" :key="item.sku_id">
            <view class="cart-item-info">
              <view class="cart-item-name">{{ item.product_name }}</view>
              <view class="cart-item-spec">{{ item.sku_name }}</view>
              <view class="cart-item-price">¥{{ item.price }}</view>
            </view>
            <view class="cart-item-actions">
              <view class="cart-btn minus" @click="decreaseItem(item)">-</view>
              <text class="cart-count">{{ item.quantity }}</text>
              <view class="cart-btn plus" @click="increaseItem(item)">+</view>
            </view>
          </view>
        </scroll-view>
      </view>
    </view>

    <view class="sku-popup" v-if="showSkuPopup && selectedProduct" @click.self="closeSkuPopup">
      <view class="sku-popup-content">
        <view class="sku-header">
          <image class="sku-image" :src="selectedProduct.main_image || '/static/default-food.png'" mode="aspectFill" />
          <view class="sku-info">
            <view class="sku-name">{{ selectedProduct.name }}</view>
            <view class="sku-price">¥{{ selectedProduct.min_price }}</view>
          </view>
          <view class="sku-close" @click="closeSkuPopup">×</view>
        </view>
        <scroll-view scroll-y class="sku-body">
          <view class="sku-list">
            <view 
              class="sku-item" 
              v-for="sku in selectedProduct.skus" 
              :key="sku.id"
              :class="{ active: selectedSkuId === sku.id }"
              @click="selectSku(sku)">
              <view class="sku-spec-name">{{ sku.spec_name }}</view>
              <view class="sku-spec-price">¥{{ sku.price }}</view>
              <view class="sku-stock" v-if="sku.stock >= 0">库存：{{ sku.stock }}</view>
            </view>
          </view>
        </scroll-view>
        <view class="sku-footer">
          <view class="sku-actions">
            <view class="cart-btn minus" @click="decreaseSku">-</view>
            <text class="cart-count">{{ skuQuantity }}</text>
            <view class="cart-btn plus" @click="increaseSku">+</view>
          </view>
          <view class="sku-confirm" @click="confirmSku">加入购物车</view>
        </view>
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import { useUserStore } from '../../store/user'
import { useCartStore, CartItem } from '../../store/cart'
import { productApi } from '../../services/product'
import { orderApi } from '../../services/order'
import { waiterApi } from '../../services/waiter'
import { tableApi } from '../../services/table'
import type { Category, Product, ProductDetail, ProductSKU } from '../../types'

const userStore = useUserStore()
const cartStore = useCartStore()

const tableId = ref<number>(0)
const tableNo = ref<string>('')
const orderId = ref<number>(0)
const action = ref<string>('new')

const categories = ref<Category[]>([])
const products = ref<Product[]>([])
const currentCategoryIndex = ref(0)
const loading = ref(false)

const showCartPopup = ref(false)
const showSkuPopup = ref(false)
const selectedProduct = ref<ProductDetail | null>(null)
const selectedSkuId = ref<number>(0)
const skuQuantity = ref(1)
const submitting = ref(false)

const categoryProducts = computed(() => {
  if (categories.value.length === 0) return []
  const currentCat = categories.value[currentCategoryIndex.value]
  if (!currentCat) return []
  return products.value.filter(p => p.category_id === currentCat.id)
})

const selectCategory = (index: number) => {
  currentCategoryIndex.value = index
}

const getCartCount = (product: Product) => {
  return cartStore.items
    .filter(item => item.product_id === product.id)
    .reduce((sum, item) => sum + item.quantity, 0)
}

const loadCategoriesAndProducts = async () => {
  if (!userStore.userInfo?.store_id) return
  loading.value = true
  try {
    const [catsData, prodsData] = await Promise.all([
      productApi.listCategories(userStore.userInfo.store_id),
      productApi.listProducts({
        store_id: userStore.userInfo.store_id,
        status: 1,
        page_size: 500
      })
    ])
    categories.value = catsData.length > 0 ? catsData : [{ id: 0, name: '全部', store_id: 0, sort_order: 0, status: 1, description: '' }]
    products.value = prodsData.list
  } catch (e: any) {
    console.error('Load products failed:', e)
  } finally {
    loading.value = false
  }
}

const toggleCartPopup = () => {
  showCartPopup.value = !showCartPopup.value
}

const clearCart = () => {
  uni.showModal({
    title: '提示',
    content: '确定清空购物车吗？',
    success: (res) => {
      if (res.confirm) {
        cartStore.clearCart()
      }
    }
  })
}

const increaseProduct = async (product: Product) => {
  try {
    const detail = await productApi.getProduct(product.id)
    selectedProduct.value = detail
    selectedSkuId.value = detail.skus.length > 0 ? detail.skus[0].id : 0
    skuQuantity.value = 1
    showSkuPopup.value = true
  } catch (e) {
    console.error('Get product detail failed:', e)
  }
}

const decreaseProduct = (product: Product) => {
  const items = cartStore.items.filter(item => item.product_id === product.id)
  if (items.length > 0) {
    cartStore.updateQuantity(items[0].sku_id, items[0].quantity - 1)
  }
}

const increaseItem = (item: CartItem) => {
  cartStore.updateQuantity(item.sku_id, item.quantity + 1)
}

const decreaseItem = (item: CartItem) => {
  cartStore.updateQuantity(item.sku_id, item.quantity - 1)
}

const closeSkuPopup = () => {
  showSkuPopup.value = false
  selectedProduct.value = null
}

const selectSku = (sku: ProductSKU) => {
  selectedSkuId.value = sku.id
}

const increaseSku = () => {
  skuQuantity.value++
}

const decreaseSku = () => {
  if (skuQuantity.value > 1) {
    skuQuantity.value--
  }
}

const confirmSku = () => {
  if (!selectedProduct.value || selectedSkuId.value === 0) {
    uni.showToast({ title: '请选择规格', icon: 'none' })
    return
  }
  const sku = selectedProduct.value.skus.find(s => s.id === selectedSkuId.value)
  if (sku) {
    cartStore.addItem(selectedProduct.value, sku, skuQuantity.value)
    closeSkuPopup()
  }
}

const submitOrder = async () => {
  if (cartStore.totalCount === 0) {
    uni.showToast({ title: '请选择菜品', icon: 'none' })
    return
  }

  if (action.value === 'checkin') {
    try {
      await tableApi.checkin({
        table_id: tableId.value,
        people_count: 1
      })
    } catch (e) {}
  }

  submitting.value = true
  try {
    const orderItems = cartStore.items.map(item => ({
      product_id: item.product_id,
      sku_id: item.sku_id,
      product_name: item.product_name,
      sku_name: item.sku_name,
      price: item.price,
      quantity: item.quantity,
      attribute_values: item.attribute_values || []
    }))

    if (action.value === 'add' && orderId.value > 0) {
      await waiterApi.addOrderItems(orderId.value, orderItems)
      uni.showToast({ title: '加菜成功', icon: 'success' })
      setTimeout(() => {
        uni.navigateBack()
      }, 1000)
    } else {
      const result = await orderApi.createOrder({
        store_id: userStore.userInfo!.store_id,
        table_no: tableNo.value,
        order_type: 'dine_in',
        items: orderItems,
        remark: cartStore.remark,
        source: 'waiter_app'
      })
      uni.showToast({ title: '下单成功', icon: 'success' })
      cartStore.clearCart()
      setTimeout(() => {
        uni.redirectTo({
          url: `/pages/orders/detail?id=${result.order_id}`
        })
      }, 1000)
    }
  } catch (e: any) {
    console.error('Submit order failed:', e)
  } finally {
    submitting.value = false
  }
}

onLoad((options: any) => {
  if (options?.tableId) tableId.value = parseInt(options.tableId)
  if (options?.tableNo) tableNo.value = options.tableNo
  if (options?.orderId) orderId.value = parseInt(options.orderId)
  if (options?.action) action.value = options.action

  if (!userStore.isLoggedIn()) {
    uni.reLaunch({ url: '/pages/login/index' })
    return
  }
  loadCategoriesAndProducts()
})
</script>

<style lang="scss" scoped>
.menu-page {
  min-height: 100vh;
  background: #f5f5f5;
  display: flex;
  flex-direction: column;
}

.menu-header {
  background: #fff;
  padding: 24rpx;
  border-bottom: 1rpx solid #ebedf0;

  .table-info {
    font-size: 30rpx;

    .table-label {
      color: #646566;
    }

    .table-no {
      font-weight: bold;
      color: #323233;
    }

    .action-text {
      color: #1989fa;
      margin-left: 8rpx;
    }
  }
}

.menu-body {
  flex: 1;
  display: flex;
  height: 0;
}

.category-list {
  width: 180rpx;
  background: #f7f8fa;
  height: 100%;

  .category-item {
    padding: 32rpx 16rpx;
    text-align: center;
    font-size: 26rpx;
    color: #646566;
    border-left: 6rpx solid transparent;

    &.active {
      background: #fff;
      color: #1989fa;
      border-left-color: #1989fa;
      font-weight: bold;
    }
  }
}

.product-list {
  flex: 1;
  height: 100%;
  padding: 16rpx;
}

.product-item {
  display: flex;
  background: #fff;
  border-radius: 12rpx;
  padding: 16rpx;
  margin-bottom: 16rpx;
}

.product-image {
  width: 160rpx;
  height: 160rpx;
  border-radius: 12rpx;
  background: #f7f8fa;
  flex-shrink: 0;
}

.product-info {
  flex: 1;
  margin-left: 16rpx;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
}

.product-name {
  font-size: 28rpx;
  font-weight: bold;
  color: #323233;
}

.product-desc {
  font-size: 22rpx;
  color: #969799;
  margin-top: 4rpx;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.product-bottom {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.product-price {
  .price-symbol {
    font-size: 22rpx;
    color: #ee0a24;
  }

  .price-value {
    font-size: 32rpx;
    color: #ee0a24;
    font-weight: bold;
  }
}

.cart-actions {
  display: flex;
  align-items: center;
}

.cart-btn {
  width: 48rpx;
  height: 48rpx;
  line-height: 48rpx;
  text-align: center;
  border-radius: 50%;
  font-size: 32rpx;

  &.plus {
    background: #1989fa;
    color: #fff;
  }

  &.minus {
    background: #fff;
    border: 2rpx solid #ebedf0;
    color: #646566;
  }
}

.cart-count {
  margin: 0 16rpx;
  font-size: 28rpx;
  color: #323233;
  min-width: 40rpx;
  text-align: center;
}

.menu-footer {
  display: flex;
  align-items: center;
  background: #fff;
  padding: 16rpx 24rpx;
  padding-bottom: calc(16rpx + env(safe-area-inset-bottom));
  box-shadow: 0 -4rpx 12rpx rgba(0, 0, 0, 0.04);
}

.cart-info {
  display: flex;
  align-items: center;
  flex: 1;
}

.cart-icon-wrapper {
  position: relative;
  width: 80rpx;
  height: 80rpx;

  .cart-icon {
    font-size: 48rpx;
  }

  .cart-badge {
    position: absolute;
    top: -4rpx;
    right: -4rpx;
    min-width: 36rpx;
    height: 36rpx;
    line-height: 36rpx;
    padding: 0 8rpx;
    border-radius: 18rpx;
    background: #ee0a24;
    color: #fff;
    font-size: 20rpx;
    text-align: center;
  }
}

.cart-amount {
  margin-left: 16rpx;

  .amount-symbol {
    font-size: 24rpx;
    color: #ee0a24;
  }

  .amount-value {
    font-size: 36rpx;
    color: #ee0a24;
    font-weight: bold;
  }
}

.footer-btn {
  background: linear-gradient(135deg, #1989fa 0%, #07c160 100%);
  color: #fff;
  padding: 20rpx 48rpx;
  border-radius: 44rpx;
  font-size: 28rpx;
  font-weight: bold;

  &.disabled {
    background: #c8c9cc;
  }
}

.cart-popup {
  position: fixed;
  left: 0;
  top: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  z-index: 100;
  display: flex;
  align-items: flex-end;
}

.cart-popup-content {
  width: 100%;
  max-height: 70vh;
  background: #fff;
  border-radius: 24rpx 24rpx 0 0;
  display: flex;
  flex-direction: column;
}

.cart-popup-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 32rpx;
  border-bottom: 1rpx solid #ebedf0;

  .cart-popup-title {
    font-size: 32rpx;
    font-weight: bold;
    color: #323233;
  }

  .cart-clear {
    font-size: 26rpx;
    color: #969799;
  }
}

.cart-items-list {
  flex: 1;
  padding: 16rpx 32rpx;
}

.cart-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 24rpx 0;
  border-bottom: 1rpx solid #f2f3f5;
}

.cart-item-info {
  .cart-item-name {
    font-size: 28rpx;
    color: #323233;
    font-weight: bold;
  }

  .cart-item-spec {
    font-size: 22rpx;
    color: #969799;
    margin-top: 4rpx;
  }

  .cart-item-price {
    font-size: 26rpx;
    color: #ee0a24;
    margin-top: 8rpx;
  }
}

.cart-item-actions {
  display: flex;
  align-items: center;
}

.sku-popup {
  position: fixed;
  left: 0;
  top: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  z-index: 200;
  display: flex;
  align-items: flex-end;
}

.sku-popup-content {
  width: 100%;
  max-height: 80vh;
  background: #fff;
  border-radius: 24rpx 24rpx 0 0;
  display: flex;
  flex-direction: column;
}

.sku-header {
  display: flex;
  padding: 32rpx;
  border-bottom: 1rpx solid #ebedf0;
  position: relative;
}

.sku-image {
  width: 160rpx;
  height: 160rpx;
  border-radius: 12rpx;
  background: #f7f8fa;
}

.sku-info {
  margin-left: 24rpx;
  flex: 1;

  .sku-name {
    font-size: 30rpx;
    font-weight: bold;
    color: #323233;
  }

  .sku-price {
    font-size: 36rpx;
    color: #ee0a24;
    margin-top: 16rpx;
    font-weight: bold;
  }
}

.sku-close {
  position: absolute;
  right: 24rpx;
  top: 16rpx;
  font-size: 48rpx;
  color: #969799;
}

.sku-body {
  flex: 1;
  padding: 16rpx 32rpx;
}

.sku-list {
  display: flex;
  flex-wrap: wrap;
  gap: 16rpx;
}

.sku-item {
  width: calc(50% - 8rpx);
  padding: 20rpx;
  background: #f7f8fa;
  border-radius: 12rpx;
  border: 2rpx solid transparent;
  box-sizing: border-box;

  &.active {
    background: #e6f7ff;
    border-color: #1989fa;
  }

  .sku-spec-name {
    font-size: 26rpx;
    color: #323233;
    font-weight: bold;
  }

  .sku-spec-price {
    font-size: 24rpx;
    color: #ee0a24;
    margin-top: 8rpx;
  }

  .sku-stock {
    font-size: 22rpx;
    color: #969799;
    margin-top: 4rpx;
  }
}

.sku-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 24rpx 32rpx;
  padding-bottom: calc(24rpx + env(safe-area-inset-bottom));
  border-top: 1rpx solid #ebedf0;
}

.sku-actions {
  display: flex;
  align-items: center;
}

.sku-confirm {
  background: linear-gradient(135deg, #1989fa 0%, #07c160 100%);
  color: #fff;
  padding: 20rpx 64rpx;
  border-radius: 44rpx;
  font-size: 28rpx;
  font-weight: bold;
}

.empty {
  padding: 80rpx 0;
  text-align: center;
  color: #969799;
  font-size: 28rpx;
}
</style>
