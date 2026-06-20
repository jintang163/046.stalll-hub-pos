<template>
  <div class="stall-cashier-page">
    <header class="stall-header">
      <div class="header-left">
        <div class="stall-info" v-if="currentStall">
          <div class="stall-logo" v-if="currentStall.logo">
            <el-image :src="currentStall.logo" fit="cover" />
          </div>
          <div class="stall-detail">
            <h2 class="stall-name">{{ currentStall.name }}</h2>
            <p class="stall-no">编号: {{ currentStall.stall_no }}</p>
          </div>
        </div>
        <el-tag :type="isOnline ? 'success' : 'danger'" size="large">
          <el-icon><component :is="isOnline ? 'Connection' : 'Cpu'" /></el-icon>
          {{ isOnline ? '在线' : '离线' }}
        </el-tag>
      </div>
      <div class="header-right">
        <div class="today-summary">
          <span class="label">今日订单:</span>
          <span class="value">{{ dailySales.orderCount }} 单</span>
          <span class="divider">|</span>
          <span class="label">营业额:</span>
          <span class="value amount">¥{{ dailySales.totalAmount.toFixed(2) }}</span>
          <span class="divider">|</span>
          <span class="label">摊位分成:</span>
          <span class="value amount stall">¥{{ dailySales.stallAmount.toFixed(2) }}</span>
        </div>
        <el-button @click="exitStallMode">
          <el-icon><Switch /></el-icon>
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
            @click="openSKUSelector(product)"
          >
            <div class="product-image">
              <el-image 
                :src="product.image || defaultImg" 
                fit="cover"
                :preview-src-list="[product.image || defaultImg]"
              />
              <div v-if="product.is_hot" class="tag hot">热销</div>
              <div v-if="product.is_recommend" class="tag recommend">推荐</div>
              <div v-if="!product.status" class="tag offline">已下架</div>
            </div>
            <div class="product-info">
              <div class="product-name">{{ product.name }}</div>
              <div class="product-price">
                <span class="price">¥{{ getMinPrice(product) }}</span>
              </div>
            </div>
          </div>
        </div>
        <el-empty v-if="!loading && filteredProducts.length === 0" description="暂无商品" />
      </main>

      <aside class="cart-panel">
        <div class="cart-header">
          <span class="cart-title">购物车</span>
          <span class="cart-count">{{ cartItemCount }} 件</span>
          <el-button text type="danger" @click="clearCart">清空</el-button>
        </div>

        <div class="cart-info">
          <el-input 
            v-model="tableNo" 
            placeholder="桌号/取餐号" 
            size="large"
          />
        </div>

        <div class="cart-list">
          <div 
            v-for="item in cartItems" 
            :key="item.id"
            class="cart-item"
          >
            <div class="item-info">
              <div class="item-name">{{ item.name }}</div>
              <div class="item-spec">{{ item.spec }}</div>
              <div class="item-price">¥{{ item.price.toFixed(2) }}</div>
            </div>
            <div class="item-qty">
              <el-button size="small" circle @click="decreaseQty(item)">-</el-button>
              <span class="qty">{{ item.quantity }}</span>
              <el-button size="small" circle type="primary" @click="increaseQty(item)">+</el-button>
            </div>
            <div class="item-subtotal">¥{{ item.subtotal.toFixed(2) }}</div>
          </div>
        </div>

        <div class="cart-summary">
          <div class="summary-row">
            <span>商品小计</span>
            <span>¥{{ subtotal.toFixed(2) }}</span>
          </div>
          <div class="summary-row">
            <span>摊位分成 ({{ (currentStall?.revenue_ratio * 100).toFixed(0) }}%)</span>
            <span class="stall-amount">¥{{ stallSubtotal.toFixed(2) }}</span>
          </div>
          <div class="summary-row total">
            <span>应付金额</span>
            <span class="total-amount">¥{{ subtotal.toFixed(2) }}</span>
          </div>
        </div>

        <div class="cart-actions">
          <el-button size="large" type="primary" @click="checkout" :disabled="cartItems.length === 0">
            收款结算
          </el-button>
        </div>
      </aside>
    </div>

    <sku-selector 
      v-model="skuSelectorVisible" 
      :product="selectedProduct"
      @confirm="addToCart"
    />
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Refresh, List, Setting, Switch, Connection, Cpu } from '@element-plus/icons-vue'
import { useStallStore } from '@/store/stall'
import { useProductStore } from '@/store/product'
import SKUSelector from '@/components/SKUSelector.vue'

const router = useRouter()
const stallStore = useStallStore()
const productStore = useProductStore()

const currentStall = computed(() => stallStore.currentStall)
const stallProducts = ref([])
const loading = ref(false)
const currentCategory = ref('all')
const cartItems = ref([])
const tableNo = ref('')
const skuSelectorVisible = ref(false)
const selectedProduct = ref(null)
const isOnline = ref(true)
const dailySales = ref({
  orderCount: 0,
  totalAmount: 0,
  stallAmount: 0,
  platformAmount: 0
})

const defaultImg = '/default-product.png'

const categories = computed(() => {
  const catMap = new Map()
  stallProducts.value.forEach(p => {
    if (p.category_id && p.category_name) {
      if (!catMap.has(p.category_id)) {
        catMap.set(p.category_id, { id: p.category_id, name: p.category_name })
      }
    }
  })
  return Array.from(catMap.values())
})

const filteredProducts = computed(() => {
  if (currentCategory.value === 'all') {
    return stallProducts.value
  }
  return stallProducts.value.filter(p => p.category_id === currentCategory.value)
})

const cartItemCount = computed(() => {
  return cartItems.value.reduce((sum, item) => sum + item.quantity, 0)
})

const subtotal = computed(() => {
  return cartItems.value.reduce((sum, item) => sum + item.subtotal, 0)
})

const stallSubtotal = computed(() => {
  const ratio = currentStall.value?.revenue_ratio || 0.7
  return subtotal.value * ratio
})

function getCategoryCount(categoryId) {
  return stallProducts.value.filter(p => p.category_id === categoryId).length
}

function getMinPrice(product) {
  if (product.skus && product.skus.length > 0) {
    return Math.min(...product.skus.map(s => s.price)).toFixed(2)
  }
  return product.price?.toFixed(2) || '0.00'
}

function openSKUSelector(product) {
  if (!product.status) return
  selectedProduct.value = product
  skuSelectorVisible.value = true
}

function addToCart(item) {
  const existing = cartItems.value.find(c => c.id === item.id && c.spec === item.spec)
  if (existing) {
    existing.quantity += item.quantity
    existing.subtotal = existing.price * existing.quantity
  } else {
    cartItems.value.push({
      ...item,
      subtotal: item.price * item.quantity
    })
  }
  skuSelectorVisible.value = false
  ElMessage.success('已添加到购物车')
}

function increaseQty(item) {
  item.quantity++
  item.subtotal = item.price * item.quantity
}

function decreaseQty(item) {
  if (item.quantity > 1) {
    item.quantity--
    item.subtotal = item.price * item.quantity
  } else {
    const index = cartItems.value.indexOf(item)
    if (index > -1) {
      cartItems.value.splice(index, 1)
    }
  }
}

function clearCart() {
  ElMessageBox.confirm('确定要清空购物车吗？', '提示', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning'
  }).then(() => {
    cartItems.value = []
    ElMessage.success('购物车已清空')
  }).catch(() => {})
}

async function checkout() {
  if (cartItems.value.length === 0) {
    ElMessage.warning('购物车为空')
    return
  }

  try {
    const orderData = {
      store_id: currentStall.value.store_id,
      stall_id: currentStall.value.id,
      table_no: tableNo.value,
      order_type: 'stall',
      source: 'stall_pos',
      items: cartItems.value.map(item => ({
        product_id: item.product_id,
        sku_id: item.sku_id,
        price: item.price,
        quantity: item.quantity,
        attribute_values: item.attributes || []
      }))
    }

    console.log('提交订单:', orderData)
    ElMessage.success('订单提交成功')
    cartItems.value = []
    tableNo.value = ''
    
    loadDailySales()
  } catch (error) {
    console.error('结算失败:', error)
    ElMessage.error('结算失败: ' + error.message)
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
    const products = await stallStore.getProductsByStall(currentStall.value.id)
    stallProducts.value = products
  } catch (error) {
    console.error('加载摊位商品失败:', error)
    ElMessage.error('加载商品失败')
  } finally {
    loading.value = false
  }
}

async function loadDailySales() {
  if (!currentStall.value) return
  
  try {
    const today = new Date().toISOString().split('T')[0]
    const sales = await stallStore.getStallDailySales(currentStall.value.id, today)
    dailySales.value = sales
  } catch (error) {
    console.error('加载日销售数据失败:', error)
  }
}

onMounted(async () => {
  if (!currentStall.value) {
    router.push('/')
    return
  }
  
  await loadStallProducts()
  await loadDailySales()
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
  height: 70px;
  background: linear-gradient(135deg, #409eff 0%, #66b1ff 100%);
  color: white;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 24px;
  box-shadow: 0 2px 8px rgba(0,0,0,0.1);

  .header-left {
    display: flex;
    align-items: center;
    gap: 20px;
  }

  .stall-info {
    display: flex;
    align-items: center;
    gap: 12px;

    .stall-logo {
      width: 50px;
      height: 50px;
      border-radius: 8px;
      overflow: hidden;
      background: white;
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
    gap: 20px;
  }

  .today-summary {
    display: flex;
    align-items: center;
    gap: 8px;
    font-size: 14px;

    .label {
      opacity: 0.9;
    }

    .value {
      font-weight: 600;

      &.amount {
        font-size: 16px;

        &.stall {
          color: #67c23a;
        }
      }
    }

    .divider {
      opacity: 0.5;
      margin: 0 4px;
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
    font-size: 16px;
    font-weight: 600;
    border-bottom: 1px solid #ebeef5;
  }

  .category-list {
    .category-item {
      display: flex;
      justify-content: space-between;
      align-items: center;
      padding: 12px 16px;
      cursor: pointer;
      transition: all 0.2s;
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
    grid-template-columns: repeat(auto-fill, minmax(160px, 1fr));
    gap: 16px;
  }

  .product-card {
    background: white;
    border-radius: 8px;
    overflow: hidden;
    cursor: pointer;
    transition: all 0.2s;
    box-shadow: 0 2px 8px rgba(0,0,0,0.06);

    &:hover {
      transform: translateY(-2px);
      box-shadow: 0 4px 12px rgba(0,0,0,0.1);
    }

    .product-image {
      position: relative;
      width: 100%;
      padding-top: 100%;

      :deep(.el-image) {
        position: absolute;
        top: 0;
        left: 0;
        width: 100%;
        height: 100%;
      }

      .tag {
        position: absolute;
        top: 8px;
        left: 8px;
        padding: 2px 8px;
        font-size: 12px;
        border-radius: 4px;
        color: white;

        &.hot {
          background: #f56c6c;
        }

        &.recommend {
          background: #e6a23c;
        }

        &.offline {
          background: #909399;
        }
      }
    }

    .product-info {
      padding: 12px;

      .product-name {
        font-size: 14px;
        font-weight: 500;
        margin-bottom: 8px;
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
    }
  }
}

.cart-panel {
  width: 360px;
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
      font-size: 14px;
      color: #909399;
      margin-left: 8px;
    }
  }

  .cart-info {
    padding: 12px 16px;
    border-bottom: 1px solid #f5f7fa;
  }

  .cart-list {
    flex: 1;
    overflow-y: auto;
    padding: 8px 0;

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
          margin-top: 4px;
        }

        .item-price {
          font-size: 14px;
          color: #f56c6c;
          margin-top: 4px;
        }
      }

      .item-qty {
        display: flex;
        align-items: center;
        gap: 8px;
        margin: 0 12px;

        .qty {
          min-width: 24px;
          text-align: center;
          font-size: 14px;
        }
      }

      .item-subtotal {
        width: 70px;
        text-align: right;
        font-size: 14px;
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

      &.total {
        margin-top: 12px;
        padding-top: 12px;
        border-top: 1px dashed #dcdfe6;
        font-size: 16px;
        font-weight: 600;

        .total-amount {
          color: #f56c6c;
          font-size: 20px;
        }
      }

      .stall-amount {
        color: #67c23a;
      }
    }
  }

  .cart-actions {
    padding: 16px;
    border-top: 1px solid #ebeef5;

    .el-button {
      width: 100%;
      height: 48px;
      font-size: 16px;
    }
  }
}
</style>
