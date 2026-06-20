<template>
  <div class="stall-cashier-page">
    <header class="stall-header">
      <div class="header-left">
        <div class="stall-info" v-if="currentStall">
          <div class="stall-logo" v-if="currentStall.logo">
            <el-image :src="currentStall.logo" fit="cover" />
          </div>
          <div v-else class="stall-logo placeholder">
            <el-icon :size="28"><Shop /></el-icon>
          </div>
          <div class="stall-detail">
            <h2 class="stall-name">{{ currentStall.name }}</h2>
            <p class="stall-no">编号: {{ currentStall.stall_no }}</p>
          </div>
        </div>
        <el-tag :type="isOnline ? 'success' : 'danger'" size="large" class="online-tag">
          <el-icon><component :is="isOnline ? 'Connection' : 'Cpu'" /></el-icon>
          {{ isOnline ? '在线' : '离线' }}
        </el-tag>
        <el-tag size="large" type="info" v-if="pendingOrderCount > 0">
          待同步: {{ pendingOrderCount }} 单
        </el-tag>
      </div>
      <div class="header-right">
        <div class="today-summary">
          <span class="label">今日订单:</span>
          <span class="value">{{ dailySales.orderCount }} 单</span>
          <span class="divider">|</span>
          <span class="label">营业额:</span>
          <span class="value amount">¥{{ formatMoney(dailySales.totalAmount) }}</span>
          <span class="divider">|</span>
          <span class="label">摊位分成:</span>
          <span class="value amount stall">¥{{ formatMoney(dailySales.stallAmount) }}</span>
        </div>
        <el-button @click="refreshData" :icon="Refresh" :loading="loading">
          刷新
        </el-button>
        <el-button @click="exitStallMode" :icon="Switch">
          退出摊位模式
        </el-button>
      </div>
    </header>

    <div class="stall-body">
      <aside class="category-sidebar">
        <div class="category-title">商品分类</div>
        <div class="category-list">
          <div
            class="category-item"
            :class="{ active: currentCategory === 'all' }"
            @click="currentCategory = 'all'"
          >
            <span>全部商品</span>
            <span class="count">{{ stallProducts.length }}</span>
          </div>
          <div
            v-for="cat in categories"
            :key="cat.id"
            class="category-item"
            :class="{ active: currentCategory === cat.id }"
            @click="currentCategory = cat.id"
          >
            <span>{{ cat.name }}</span>
            <span class="count">{{ getCategoryCount(cat.id) }}</span>
          </div>
        </div>
      </aside>

      <main class="product-panel">
        <div class="product-list" v-loading="loading">
          <div
            v-for="product in filteredProducts"
            :key="product.id"
            class="product-card"
            :class="{ disabled: !product.status || (product.stock != null && product.stock <= 0) }"
            @click="openSKUSelector(product)"
          >
            <div class="product-image">
              <el-image
                :src="product.image || defaultImg"
                fit="cover"
                :preview-src-list="product.image ? [product.image] : []"
              />
              <div v-if="product.is_hot" class="tag hot">热销</div>
              <div v-if="product.is_recommend" class="tag recommend">推荐</div>
              <div v-if="!product.status" class="tag offline">已下架</div>
              <div v-if="product.stock != null && product.stock <= 0" class="tag soldout">售罄</div>
            </div>
            <div class="product-info">
              <div class="product-name">{{ product.name }}</div>
              <div class="product-price">
                <span class="price">¥{{ getMinPrice(product) }}</span>
                <span v-if="hasMultiSKU(product)" class="multi-sku">多规格</span>
              </div>
              <div class="product-stock" :class="{ 'low-stock': isLowStock(product) }">
                库存: {{ getTotalStock(product) }}
              </div>
            </div>
          </div>
        </div>
        <el-empty v-if="!loading && filteredProducts.length === 0" description="暂无商品，点击右上角刷新数据" />
      </main>

      <aside class="cart-panel">
        <div class="cart-header">
          <span class="cart-title">购物车</span>
          <span class="cart-count">{{ cartItemCount }} 件</span>
          <el-button text type="danger" @click="clearCart" :disabled="cartItems.length === 0">清空</el-button>
        </div>

        <div class="cart-info">
          <el-input
            v-model="tableNo"
            placeholder="桌号/取餐号"
            size="large"
            clearable
          />
          <el-input
            v-model="remark"
            placeholder="订单备注(可选)"
            class="mt-10"
            type="textarea"
            :rows="2"
            resize="none"
          />
        </div>

        <div class="cart-list">
          <div
            v-for="item in cartItems"
            :key="item._key"
            class="cart-item"
          >
            <div class="item-info">
              <div class="item-name">{{ item.product_name }}</div>
              <div class="item-spec">{{ item.sku_name || item.spec }}</div>
              <div class="item-price">¥{{ Number(item.price).toFixed(2) }}</div>
            </div>
            <div class="item-qty">
              <el-button size="small" circle @click="updateItemQuantity(item, item.quantity - 1)" :disabled="submittingOrder">-</el-button>
              <span class="qty">{{ item.quantity }}</span>
              <el-button size="small" circle type="primary" @click="updateItemQuantity(item, item.quantity + 1)" :disabled="submittingOrder">+</el-button>
            </div>
            <div class="item-subtotal">¥{{ (item.price * item.quantity).toFixed(2) }}</div>
          </div>
        </div>
        <el-empty v-if="cartItems.length === 0" description="购物车为空，点击左侧商品添加" :image-size="60" />

        <div class="cart-summary">
          <div class="summary-row">
            <span>商品小计</span>
            <span>¥{{ formatMoney(subtotal) }}</span>
          </div>
          <div class="summary-row discount-row" v-if="discountAmount > 0">
            <span>优惠</span>
            <span class="discount">-¥{{ formatMoney(discountAmount) }}</span>
          </div>
          <div class="summary-row">
            <span>摊位分成 ({{ formatRatio(currentStall?.revenue_ratio) }}%)</span>
            <span class="stall-amount">¥{{ formatMoney(stallSubtotal) }}</span>
          </div>
          <div class="summary-row total">
            <span>应付金额</span>
            <span class="total-amount">¥{{ formatMoney(payAmount) }}</span>
          </div>
        </div>

        <div class="cart-actions">
          <el-alert
            v-if="!isOnline"
            type="warning"
            :closable="false"
            show-icon
            title="当前离线，订单将保存本地，网络恢复后自动同步"
            class="mb-10"
          />
          <el-button
            size="large"
            type="primary"
            @click="openPayDialog"
            :disabled="cartItems.length === 0 || submittingOrder"
            :loading="submittingOrder"
          >
            {{ isOnline ? '收款结算' : '离线下单' }}
          </el-button>
        </div>
      </aside>
    </div>

    <sku-selector
      v-model="skuSelectorVisible"
      :product="selectedProduct"
      @confirm="addToCart"
    />

    <el-dialog v-model="payDialogVisible" title="选择支付方式" width="420px" :close-on-click-modal="false">
      <div class="pay-method-list">
        <div
          v-for="method in payMethods"
          :key="method.value"
          class="pay-method-item"
          :class="{ active: selectedPayMethod === method.value }"
          @click="selectedPayMethod = method.value"
        >
          <el-icon><component :is="method.icon" /></el-icon>
          <span>{{ method.label }}</span>
        </div>
      </div>

      <div class="pay-amount-section">
        <div class="pay-amount-row">
          <span>应收金额</span>
          <span class="value">¥{{ formatMoney(payAmount) }}</span>
        </div>
        <div class="pay-amount-row discount-input" v-if="selectedPayMethod === 'cash'">
          <span>收款金额</span>
          <el-input-number v-model="receivedAmount" :min="payAmount" :precision="2" :step="1" size="large" />
        </div>
        <div class="pay-amount-row change-row" v-if="selectedPayMethod === 'cash' && receivedAmount > payAmount">
          <span>找零</span>
          <span class="value change">¥{{ formatMoney(receivedAmount - payAmount) }}</span>
        </div>
      </div>

      <template #footer>
        <el-button @click="payDialogVisible = false" :disabled="submittingOrder">取消</el-button>
        <el-button type="primary" @click="checkout" :loading="submittingOrder">确认收款</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, watch, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  Refresh, List, Setting, Switch, Connection, Cpu, Shop,
  Money, Wallet, CreditCard, Postcard
} from '@element-plus/icons-vue'
import { useStallStore } from '@/store/stall'
import SKUSelector from '@/components/SKUSelector.vue'
import { submitOrder } from '@/api/order'
import { getStallDailyReport, getStallDevices } from '@/api/stall'

const router = useRouter()
const stallStore = useStallStore()

const currentStall = computed(() => stallStore.currentStall)
const stallProducts = ref([])
const categories = ref([])
const loading = ref(false)
const currentCategory = ref('all')
const cartItems = ref([])
const tableNo = ref('')
const remark = ref('')
const skuSelectorVisible = ref(false)
const selectedProduct = ref(null)
const isOnline = ref(navigator.onLine)
const pendingOrderCount = ref(0)
const submittingOrder = ref(false)
const payDialogVisible = ref(false)
const selectedPayMethod = ref('cash')
const receivedAmount = ref(0)
const discountAmount = ref(0)

const payMethods = reactive([
  { value: 'cash', label: '现金', icon: Money },
  { value: 'wechat', label: '微信支付', icon: Wallet },
  { value: 'alipay', label: '支付宝', icon: Postcard },
  { value: 'card', label: '会员卡', icon: CreditCard },
])

const defaultImg = 'data:image/svg+xml,%3Csvg xmlns="http://www.w3.org/2000/svg" width="100" height="100"%3E%3Crect fill="%23f0f0f0" width="100" height="100"/%3E%3Ctext x="50" y="55" text-anchor="middle" fill="%23999"%3E暂无图%3C/text%3E%3C/svg%3E'

const dailySales = reactive({
  orderCount: 0,
  totalAmount: 0,
  paidAmount: 0,
  stallAmount: 0,
  platformAmount: 0
})

const filteredProducts = computed(() => {
  if (currentCategory.value === 'all') {
    return stallProducts.value
  }
  return stallProducts.value.filter(p => String(p.category_id) === String(currentCategory.value))
})

const cartItemCount = computed(() => {
  return cartItems.value.reduce((sum, item) => sum + item.quantity, 0)
})

const subtotal = computed(() => {
  return cartItems.value.reduce((sum, item) => sum + (item.price * item.quantity), 0)
})

const payAmount = computed(() => {
  return Math.max(0, subtotal.value - discountAmount.value)
})

const stallSubtotal = computed(() => {
  const ratio = currentStall.value?.revenue_ratio || 0.7
  return payAmount.value * ratio
})

function formatMoney(amount) {
  const num = Number(amount || 0)
  return num.toFixed(2)
}

function formatRatio(ratio) {
  if (ratio == null) return '70'
  return (Number(ratio) * 100).toFixed(0)
}

function getCategoryCount(categoryId) {
  return stallProducts.value.filter(p => String(p.category_id) === String(categoryId)).length
}

function getMinPrice(product) {
  if (product.skus && product.skus.length > 0) {
    return Math.min(...product.skus.map(s => Number(s.price))).toFixed(2)
  }
  return Number(product.price || 0).toFixed(2)
}

function hasMultiSKU(product) {
  return product.skus && product.skus.length > 1
}

function isLowStock(product) {
  if (product.stock == null) return false
  return product.stock > 0 && product.stock <= 10
}

function getTotalStock(product) {
  if (product.stock != null) return product.stock
  if (product.skus && product.skus.length > 0) {
    return product.skus.reduce((sum, s) => sum + (Number(s.stock) || 0), 0)
  }
  return '--'
}

function openSKUSelector(product) {
  if (!product.status) {
    ElMessage.warning('该商品已下架')
    return
  }
  if (product.stock != null && product.stock <= 0) {
    ElMessage.warning('该商品已售罄')
    return
  }
  selectedProduct.value = product
  skuSelectorVisible.value = true
}

function addToCart(item) {
  const key = item.product_id + '_' + (item.sku_id || 'default') + '_' + (item.attribute_values?.join(',') || '')
  const existing = cartItems.value.find(c => c._key === key)

  if (existing) {
    existing.quantity += item.quantity
  } else {
    cartItems.value.push({
      ...item,
      _key: key
    })
  }
  skuSelectorVisible.value = false
  ElMessage.success('已添加到购物车')
}

function updateItemQuantity(item, newQty) {
  if (newQty <= 0) {
    const index = cartItems.value.indexOf(item)
    if (index > -1) cartItems.value.splice(index, 1)
    return
  }
  item.quantity = newQty
}

function clearCart() {
  if (cartItems.value.length === 0) return
  ElMessageBox.confirm('确定要清空购物车吗？', '提示', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning'
  }).then(() => {
    cartItems.value = []
    tableNo.value = ''
    remark.value = ''
    discountAmount.value = 0
    ElMessage.success('购物车已清空')
  }).catch(() => {})
}

function openPayDialog() {
  if (cartItems.value.length === 0) {
    ElMessage.warning('购物车为空')
    return
  }
  selectedPayMethod.value = 'cash'
  receivedAmount.value = payAmount.value
  payDialogVisible.value = true
}

async function checkout() {
  submittingOrder.value = true
  try {
    const storeId = currentStall.value?.store_id || 1
    const orderItems = cartItems.value.map(item => ({
      product_id: item.product_id,
      sku_id: item.sku_id || null,
      price: Number(item.price),
      quantity: item.quantity,
      attribute_values: item.attribute_values || []
    }))

    const orderData = {
      store_id: Number(storeId),
      stall_id: Number(currentStall.value?.id),
      table_no: tableNo.value,
      order_type: 'dine_in',
      source: 'stall_pos',
      pay_method: selectedPayMethod.value,
      pay_amount: Number(payAmount.value.toFixed(2)),
      items: orderItems,
      remark: remark.value
    }

    let orderResult
    if (isOnline.value) {
      orderResult = await submitOrder(orderData)
      ElMessage.success('订单提交成功：' + (orderResult?.order_no || ''))
    } else {
      await saveOfflineOrder(orderData)
      ElMessage.success('订单已保存本地，网络恢复后自动同步')
    }

    await saveOrderToLocalDB(orderData, orderResult)

    payDialogVisible.value = false
    cartItems.value = []
    tableNo.value = ''
    remark.value = ''
    discountAmount.value = 0

    await loadDailySales()

  } catch (error) {
    console.error('结算失败:', error)
    if (isOnline.value) {
      try {
        await saveOfflineOrder({
          ...orderData,
          pay_status: 1
        })
        ElMessage.warning('服务器提交失败，已保存本地订单')
        payDialogVisible.value = false
        cartItems.value = []
      } catch (saveErr) {
        ElMessage.error('提交失败，且本地保存也失败：' + saveErr.message)
      }
    } else {
      ElMessage.error('本地保存失败：' + error.message)
    }
  } finally {
    submittingOrder.value = false
  }
}

async function saveOfflineOrder(orderData) {
  if (window.electronAPI) {
    const localOrder = {
      order_no: generateOrderNo(),
      ...orderData,
      created_at: new Date().toISOString(),
      sync_status: 0
    }
    await window.electronAPI.invoke('db:insertOrder', localOrder)
    pendingOrderCount.value++
  }
}

function generateOrderNo() {
  const now = new Date()
  const stamp = now.getFullYear().toString() +
    (now.getMonth() + 1).toString().padStart(2, '0') +
    now.getDate().toString().padStart(2, '0') +
    now.getHours().toString().padStart(2, '0') +
    now.getMinutes().toString().padStart(2, '0') +
    now.getSeconds().toString().padStart(2, '0')
  const rand = Math.floor(Math.random() * 1000).toString().padStart(3, '0')
  return 'SO' + currentStall.value?.id + stamp + rand
}

async function saveOrderToLocalDB(orderData, orderResult) {
  if (!window.electronAPI) return
  try {
    const items = cartItems.value.map(i => ({
      ...i,
      stall_id: currentStall.value?.id,
      stall_amount: (i.price * i.quantity) * (currentStall.value?.revenue_ratio || 0.7),
      platform_amount: (i.price * i.quantity) * (currentStall.value?.platform_ratio || 0.3)
    }))
    const order = {
      order_no: orderResult?.order_no || generateOrderNo(),
      store_id: orderData.store_id,
      stall_id: orderData.stall_id,
      table_no: orderData.table_no,
      order_type: orderData.order_type,
      source: orderData.source,
      total_amount: subtotal.value,
      pay_amount: payAmount.value,
      discount_amount: discountAmount.value,
      pay_method: orderData.pay_method,
      status: 'paid',
      remark: orderData.remark,
      sync_status: 1,
      items: items
    }
    await window.electronAPI.invoke('db:insertOrder', order)
  } catch (e) {
    console.warn('本地订单记录失败', e)
  }
}

function exitStallMode() {
  ElMessageBox.confirm('确定要退出摊位模式吗？', '提示', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning'
  }).then(() => {
    stallStore.setCurrentStall(null)
    router.push('/')
  }).catch(() => {})
}

async function loadStallProducts() {
  if (!currentStall.value) return
  loading.value = true
  try {
    if (window.electronAPI) {
      const products = await stallStore.getProductsByStall(currentStall.value.id)
      stallProducts.value = products || []
      buildCategories()
      if (stallProducts.value.length === 0) {
        ElMessage.info('暂无本地商品，请先在总收银台完成数据同步')
      }
    }
  } catch (error) {
    console.error('加载摊位商品失败:', error)
    ElMessage.error('加载商品失败')
  } finally {
    loading.value = false
  }
}

function buildCategories() {
  const catMap = new Map()
  stallProducts.value.forEach(p => {
    if (p.category_id) {
      if (!catMap.has(p.category_id)) {
        catMap.set(p.category_id, {
          id: p.category_id,
          name: p.category_name || ('分类' + p.category_id)
        })
      }
    }
  })
  categories.value = Array.from(catMap.values())
}

async function loadDailySales() {
  if (!currentStall.value) return
  try {
    const today = new Date().toISOString().split('T')[0]
    const params = {
      stall_id: currentStall.value.id,
      start_date: today,
      end_date: today
    }

    try {
      if (isOnline.value) {
        const reports = await getStallDailyReport(params)
        if (reports && reports.length > 0) {
          const report = reports[0]
          dailySales.orderCount = report.order_count || 0
          dailySales.totalAmount = Number(report.total_amount || 0)
          dailySales.stallAmount = Number(report.stall_amount || 0)
          dailySales.platformAmount = Number(report.platform_amount || 0)
          return
        }
      }
    } catch (e) {
      console.warn('从后端加载日报失败，改用本地数据', e)
    }

    if (window.electronAPI) {
      const local = await stallStore.getStallDailySales(currentStall.value.id, today)
      dailySales.orderCount = local.orderCount || 0
      dailySales.totalAmount = Number(local.totalAmount || 0)
      dailySales.stallAmount = Number(local.stallAmount || 0)
      dailySales.platformAmount = Number(local.platformAmount || 0)
    }
  } catch (error) {
    console.error('加载日销售数据失败:', error)
  }
}

async function loadPendingCount() {
  if (!window.electronAPI) return
  try {
    const pending = await window.electronAPI.invoke('db:getPendingOrderCount')
    pendingOrderCount.value = pending || 0
  } catch (e) {}
}

async function refreshData() {
  await Promise.all([
    loadStallProducts(),
    loadDailySales(),
    loadPendingCount()
  ])
  ElMessage.success('数据已刷新')
}

function handleOnline() {
  isOnline.value = true
}

function handleOffline() {
  isOnline.value = false
}

onMounted(async () => {
  if (!currentStall.value) {
    router.push('/')
    return
  }

  window.addEventListener('online', handleOnline)
  window.addEventListener('offline', handleOffline)

  await refreshData()
})

import { onBeforeUnmount } from 'vue'
onBeforeUnmount(() => {
  window.removeEventListener('online', handleOnline)
  window.removeEventListener('offline', handleOffline)
})
</script>

<style scoped lang="scss">
.stall-cashier-page {
  height: 100vh;
  display: flex;
  flex-direction: column;
  background: #f5f7fa;
}

.stall-header {
  height: 76px;
  background: linear-gradient(135deg, #667eea 0%, #409eff 100%);
  color: white;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 24px;
  box-shadow: 0 2px 8px rgba(0,0,0,0.15);

  .header-left {
    display: flex;
    align-items: center;
    gap: 20px;
  }

  .online-tag {
    margin-left: 8px;
  }

  .stall-info {
    display: flex;
    align-items: center;
    gap: 12px;

    .stall-logo {
      width: 52px;
      height: 52px;
      border-radius: 10px;
      overflow: hidden;
      background: white;
      display: flex;
      align-items: center;
      justify-content: center;
      color: #409eff;

      &.placeholder {
        background: rgba(255,255,255,0.9);
      }
    }

    .stall-detail {
      .stall-name {
        font-size: 20px;
        font-weight: 600;
        margin: 0;
      }
      .stall-no {
        font-size: 13px;
        opacity: 0.9;
        margin: 2px 0 0 0;
      }
    }
  }

  .header-right {
    display: flex;
    align-items: center;
    gap: 16px;
  }

  .today-summary {
    display: flex;
    align-items: center;
    gap: 8px;
    font-size: 13px;
    padding: 8px 16px;
    background: rgba(255,255,255,0.15);
    border-radius: 8px;

    .label {
      opacity: 0.9;
    }
    .value {
      font-weight: 600;
      &.amount {
        font-size: 15px;
        &.stall {
          color: #67f2a0;
        }
      }
    }
    .divider {
      opacity: 0.4;
      margin: 0 6px;
    }
  }
}

.stall-body {
  flex: 1;
  display: flex;
  overflow: hidden;
}

.category-sidebar {
  width: 180px;
  background: white;
  border-right: 1px solid #ebeef5;
  overflow-y: auto;

  .category-title {
    padding: 16px;
    font-size: 15px;
    font-weight: 600;
    border-bottom: 1px solid #ebeef5;
  }

  .category-list {
    .category-item {
      display: flex;
      justify-content: space-between;
      align-items: center;
      padding: 14px 16px;
      cursor: pointer;
      transition: all 0.15s;
      border-left: 3px solid transparent;

      &:hover {
        background: #f5f7fa;
      }
      &.active {
        background: #ecf5ff;
        border-left-color: #409eff;
        color: #409eff;
        font-weight: 500;
      }

      .count {
        font-size: 12px;
        color: #909399;
        background: #f5f7fa;
        padding: 2px 8px;
        border-radius: 10px;
      }
    }
  }
}

.product-panel {
  flex: 1;
  overflow-y: auto;
  padding: 20px;

  .product-list {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(170px, 1fr));
    gap: 16px;
  }

  .product-card {
    background: white;
    border-radius: 10px;
    overflow: hidden;
    cursor: pointer;
    transition: all 0.2s;
    box-shadow: 0 2px 10px rgba(0,0,0,0.06);

    &:hover:not(.disabled) {
      transform: translateY(-3px);
      box-shadow: 0 6px 18px rgba(0,0,0,0.12);
    }

    &.disabled {
      opacity: 0.6;
      cursor: not-allowed;
    }

    .product-image {
      position: relative;
      width: 100%;
      padding-top: 100%;
      background: #f5f7fa;

      :deep(.el-image) {
        position: absolute;
        top: 0; left: 0;
        width: 100%; height: 100%;
      }

      .tag {
        position: absolute;
        top: 8px;
        left: 8px;
        padding: 2px 8px;
        font-size: 11px;
        border-radius: 4px;
        color: white;

        &.hot { background: #f56c6c; }
        &.recommend { background: #e6a23c; }
        &.offline { background: #909399; }
        &.soldout { background: #606266; top: auto; bottom: 8px; left: 8px; right: 8px; text-align: center; }
      }
    }

    .product-info {
      padding: 12px;

      .product-name {
        font-size: 14px;
        font-weight: 500;
        margin-bottom: 6px;
        overflow: hidden;
        text-overflow: ellipsis;
        white-space: nowrap;
      }

      .product-price {
        display: flex;
        align-items: center;
        gap: 8px;

        .price {
          font-size: 18px;
          font-weight: 600;
          color: #f56c6c;
        }
        .multi-sku {
          font-size: 12px;
          color: #909399;
          background: #f5f7fa;
          padding: 2px 6px;
          border-radius: 4px;
        }
      }

      .product-stock {
        margin-top: 6px;
        font-size: 12px;
        color: #909399;

        &.low-stock { color: #e6a23c; }
      }
    }
  }
}

.cart-panel {
  width: 380px;
  background: white;
  border-left: 1px solid #ebeef5;
  display: flex;
  flex-direction: column;

  .cart-header {
    display: flex;
    align-items: center;
    padding: 16px;
    border-bottom: 1px solid #ebeef5;

    .cart-title {
      font-size: 16px;
      font-weight: 600;
    }
    .cart-count {
      flex: 1;
      font-size: 13px;
      color: #909399;
      margin-left: 8px;
    }
  }

  .cart-info {
    padding: 12px 16px;
    border-bottom: 1px solid #f5f7fa;

    :deep(.el-textarea) {
      margin-top: 10px;
    }
  }

  .cart-list {
    flex: 1;
    overflow-y: auto;
    padding: 4px 0;

    .cart-item {
      display: flex;
      align-items: center;
      padding: 12px 16px;
      border-bottom: 1px solid #f5f7fa;

      .item-info {
        flex: 1;
        min-width: 0;

        .item-name {
          font-size: 14px;
          font-weight: 500;
          overflow: hidden;
          text-overflow: ellipsis;
          white-space: nowrap;
        }
        .item-spec {
          font-size: 12px;
          color: #909399;
          margin-top: 3px;
        }
        .item-price {
          font-size: 13px;
          color: #f56c6c;
          margin-top: 3px;
        }
      }

      .item-qty {
        display: flex;
        align-items: center;
        gap: 6px;
        margin: 0 12px;

        .qty {
          min-width: 26px;
          text-align: center;
          font-size: 14px;
          font-weight: 500;
        }
      }

      .item-subtotal {
        width: 76px;
        text-align: right;
        font-size: 15px;
        font-weight: 600;
        color: #f56c6c;
      }
    }
  }

  .cart-summary {
    padding: 16px;
    border-top: 1px solid #ebeef5;
    background: #fafafa;

    .summary-row {
      display: flex;
      justify-content: space-between;
      margin-bottom: 10px;
      font-size: 14px;
      color: #606266;

      &.discount-row {
        .discount { color: #67c23a; }
      }

      &.total {
        margin-top: 12px;
        padding-top: 12px;
        border-top: 1px dashed #dcdfe6;
        font-size: 15px;
        font-weight: 600;
        color: #303133;

        .total-amount {
          color: #f56c6c;
          font-size: 22px;
        }
      }

      .stall-amount {
        color: #409eff;
        font-weight: 500;
      }
    }
  }

  .cart-actions {
    padding: 16px;
    border-top: 1px solid #ebeef5;

    .el-button {
      width: 100%;
      height: 52px;
      font-size: 17px;
      font-weight: 600;
    }
  }
}

.pay-method-list {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 12px;

  .pay-method-item {
    display: flex;
    align-items: center;
    gap: 10px;
    padding: 16px;
    border: 2px solid #ebeef5;
    border-radius: 8px;
    cursor: pointer;
    transition: all 0.15s;
    font-size: 15px;
    font-weight: 500;

    &:hover {
      border-color: #409eff;
    }

    &.active {
      border-color: #409eff;
      background: #ecf5ff;
      color: #409eff;
    }

    .el-icon {
      font-size: 24px;
    }
  }
}

.pay-amount-section {
  margin-top: 20px;
  padding: 16px;
  background: #f5f7fa;
  border-radius: 8px;

  .pay-amount-row {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 12px;
    font-size: 14px;

    &:last-child {
      margin-bottom: 0;
    }

    &.discount-input {
      align-items: center;
    }

    &.change-row {
      .value.change {
        color: #67c23a;
        font-weight: 600;
      }
    }

    .value {
      font-size: 18px;
      font-weight: 600;
      color: #f56c6c;
    }
  }
}

.mb-10 { margin-bottom: 10px; }
.mt-10 { margin-top: 10px; }
</style>
