<template>
  <div class="cashier-page">
    <header class="cashier-header">
      <div class="header-left">
        <h1 class="logo">大排档收银系统</h1>
        <el-tag :type="orderStore.isOnline ? 'success' : 'danger'" size="large">
          <el-icon><component :is="orderStore.isOnline ? 'Connection' : 'Cpu'" /></el-icon>
          {{ orderStore.isOnline ? '在线' : '离线' }}
        </el-tag>
        <el-tag size="large" type="info">
          待同步: {{ orderStore.pendingOrders.length }} 单
        </el-tag>
      </div>
      <div class="header-right">
        <el-button @click="handleSyncOrders">
          <el-icon><Refresh /></el-icon>
          同步订单
        </el-button>
        <el-button @click="showSync = true">
          <el-icon><Download /></el-icon>
          数据同步
        </el-button>
        <el-button @click="goOrders">
          <el-icon><List /></el-icon>
          订单管理
        </el-button>
        <el-button @click="goStallReport">
          <el-icon><DataAnalysis /></el-icon>
          摊位报表
        </el-button>
        <el-button @click="goSettings">
          <el-icon><Setting /></el-icon>
          设置
        </el-button>
      </div>
    </header>

    <div class="cashier-body">
      <aside class="category-sidebar">
        <div class="category-title">商品分类</div>
        <div class="category-list">
          <div 
            class="category-item"
            :class="{ active: productStore.currentCategory === 'all' }"
            @click="productStore.setCategory('all')"
          >
            <span>全部商品</span>
            <span class="count">{{ productStore.products.filter(p => p.status).length }}</span>
          </div>
          <div 
            v-for="cat in productStore.categories" 
            :key="cat.id"
            class="category-item"
            :class="{ active: productStore.currentCategory === cat.id }"
            @click="productStore.setCategory(cat.id)"
          >
            <span>{{ cat.name }}</span>
            <span class="count">{{ getCategoryCount(cat.id) }}</span>
          </div>
        </div>
      </aside>

      <main class="product-panel">
        <div class="product-list" v-loading="productStore.loading">
          <div 
            v-for="product in productStore.filteredProducts" 
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
                <span v-if="hasMultiSKU(product)" class="multi-sku">多规格</span>
              </div>
              <div class="product-stock" :class="{ 'low-stock': isLowStock(product) }">
                库存: {{ getTotalStock(product) }}
              </div>
            </div>
          </div>
        </div>
        <el-empty v-if="!productStore.loading && productStore.filteredProducts.length === 0" description="暂无商品" />
      </main>

      <aside class="cart-panel">
        <div class="cart-header">
          <span class="cart-title">购物车</span>
          <span class="cart-count">{{ cartStore.itemCount }} 件</span>
          <el-button text type="danger" @click="clearCart">清空</el-button>
        </div>

        <div class="cart-info">
          <el-input 
            v-model="cartStore.tableNo" 
            placeholder="桌号(可选)"
            class="mb-10"
            clearable
          />
          <el-input 
            v-model="cartStore.remark" 
            placeholder="订单备注(可选)"
            type="textarea"
            :rows="2"
            clearable
          />
        </div>

        <div class="cart-items" v-if="cartStore.items.length > 0">
          <div v-for="(item, index) in cartStore.items" :key="item._key" class="cart-item">
            <div class="item-info">
              <div class="item-name">{{ item.product_name }}</div>
              <div class="item-spec">
                {{ item.sku_name }}
                <span v-if="item.attribute_names && item.attribute_names.length">
                  / {{ item.attribute_names.join(' / ') }}
                </span>
              </div>
              <div class="item-price">¥{{ item.price.toFixed(2) }} × {{ item.quantity }}</div>
            </div>
            <div class="item-actions">
              <el-input-number 
                v-model="item.quantity" 
                :min="1" 
                :max="99"
                size="small"
                @change="updateItemQuantity(index, item.quantity)"
              />
              <div class="item-subtotal">¥{{ item.subtotal.toFixed(2) }}</div>
              <el-button 
                text 
                type="danger" 
                size="small"
                @click="removeItem(index)"
              >
                删除
              </el-button>
            </div>
          </div>
        </div>
        <el-empty v-else description="购物车是空的" :image-size="80" />

        <div class="cart-footer">
          <div class="cart-summary">
            <div class="summary-row">
              <span>合计:</span>
              <span class="total">¥{{ cartStore.total.toFixed(2) }}</span>
            </div>
            <div v-if="cartStore.discountAmount > 0" class="summary-row discount">
              <span>优惠:</span>
              <span>-¥{{ cartStore.discountAmount.toFixed(2) }}</span>
            </div>
            <div class="summary-row actual">
              <span>应收:</span>
              <span class="actual-total">¥{{ cartStore.actualTotal.toFixed(2) }}</span>
            </div>
          </div>
          <div class="cart-actions">
            <el-button 
              size="large" 
              type="primary" 
              :disabled="cartStore.items.length === 0"
              @click="openPayment"
            >
              立即下单 ({{ cartStore.itemCount }})
            </el-button>
          </div>
        </div>
      </aside>
    </div>

    <sku-selector 
      v-model="skuSelectorVisible" 
      :product="selectedProduct"
      @add="handleAddToCart"
    />

    <el-dialog v-model="paymentVisible" title="收款" width="500px">
      <div class="payment-dialog">
        <div class="payment-amount">
          <div class="label">应收金额</div>
          <div class="amount">¥{{ cartStore.actualTotal.toFixed(2) }}</div>
        </div>
        
        <el-form label-width="80px">
          <el-form-item label="收款方式">
            <el-radio-group v-model="payMethod" size="large">
              <el-radio-button value="cash">现金</el-radio-button>
              <el-radio-button value="wechat">微信</el-radio-button>
              <el-radio-button value="alipay">支付宝</el-radio-button>
              <el-radio-button value="card">刷卡</el-radio-button>
            </el-radio-group>
          </el-form-item>
          
          <el-form-item label="优惠金额">
            <el-input-number 
              v-model="discountInput" 
              :min="0" 
              :max="cartStore.total"
              :precision="2"
              size="large"
              style="width: 100%;"
            />
          </el-form-item>
          
          <el-form-item label="备注">
            <el-input 
              v-model="cartStore.remark" 
              type="textarea" 
              :rows="2"
              placeholder="订单备注"
            />
          </el-form-item>
        </el-form>

        <div class="payment-summary">
          <div class="row">
            <span>原价:</span>
            <span>¥{{ cartStore.total.toFixed(2) }}</span>
          </div>
          <div class="row discount">
            <span>优惠:</span>
            <span>-¥{{ discountInput.toFixed(2) }}</span>
          </div>
          <div class="row actual">
            <span>实收:</span>
            <span class="actual">¥{{ (cartStore.total - discountInput).toFixed(2) }}</span>
          </div>
        </div>
      </div>
      
      <template #footer>
        <el-button @click="paymentVisible = false">取消</el-button>
        <el-button 
          type="primary" 
          size="large"
          @click="handleSubmitOrder"
        >
          确认收款
        </el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="showSync" title="数据同步" width="500px">
      <div class="sync-dialog">
        <div class="sync-info">
          <el-descriptions :column="1" border>
            <el-descriptions-item label="上次同步时间">
              {{ syncStore.lastSyncTime ? formatDate(syncStore.lastSyncTime) : '从未同步' }}
            </el-descriptions-item>
            <el-descriptions-item label="上次同步ID">
              {{ syncStore.lastSyncID }}
            </el-descriptions-item>
            <el-descriptions-item label="本地商品数">
              {{ productStore.products.length }}
            </el-descriptions-item>
            <el-descriptions-item label="本地分类数">
              {{ productStore.categories.length }}
            </el-descriptions-item>
          </el-descriptions>
        </div>
        
        <div class="sync-actions">
          <el-button 
            type="primary" 
            size="large"
            :loading="syncStore.isSyncing"
            @click="handleFullSync"
          >
            <el-icon><RefreshRight /></el-icon>
            全量同步
          </el-button>
          <el-button 
            size="large"
            :loading="syncStore.isSyncing"
            @click="handleIncrementalSync"
          >
            <el-icon><Top /></el-icon>
            增量同步
          </el-button>
        </div>
      </div>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { 
  Refresh, Download, List, Setting, Connection, Cpu,
  RefreshRight, Top, DataAnalysis
} from '@element-plus/icons-vue'
import { useProductStore } from '@/store/product'
import { useCartStore } from '@/store/cart'
import { useOrderStore } from '@/store/order'
import { useSyncStore } from '@/store/sync'
import SKUSelector from '@/components/SKUSelector.vue'
import dayjs from 'dayjs'

const router = useRouter()
const productStore = useProductStore()
const cartStore = useCartStore()
const orderStore = useOrderStore()
const syncStore = useSyncStore()

const defaultImg = 'data:image/svg+xml,%3Csvg xmlns="http://www.w3.org/2000/svg" width="100" height="100"%3E%3Crect fill="%23f0f0f0" width="100" height="100"/%3E%3Ctext x="50" y="55" text-anchor="middle" fill="%23999"%3E暂无图%3C/text%3E%3C/svg%3E'

const skuSelectorVisible = ref(false)
const selectedProduct = ref(null)
const paymentVisible = ref(false)
const showSync = ref(false)
const payMethod = ref('cash')
const discountInput = ref(0)

const loadData = async () => {
  await productStore.loadCategories()
  await productStore.loadProducts()
}

const getCategoryCount = (catId) => {
  return productStore.products.filter(p => p.category_id === catId && p.status).length
}

const getMinPrice = (product) => {
  if (!product.skus || product.skus.length === 0) return '0.00'
  const prices = product.skus.filter(s => s.status).map(s => s.price)
  if (prices.length === 0) return '0.00'
  return Math.min(...prices).toFixed(2)
}

const hasMultiSKU = (product) => {
  return product.skus && product.skus.length > 1
}

const getTotalStock = (product) => {
  if (!product.skus || product.skus.length === 0) return 0
  return product.skus.reduce((sum, s) => sum + (s.stock > 0 ? s.stock : 0), 0)
}

const isLowStock = (product) => {
  const total = getTotalStock(product)
  return total > 0 && total <= (product.warning_threshold || 10)
}

const openSKUSelector = (product) => {
  if (!product.status) {
    ElMessage.warning('该商品已下架')
    return
  }
  selectedProduct.value = product
  skuSelectorVisible.value = true
}

const handleAddToCart = ({ product, sku, attributes, quantity }) => {
  cartStore.addItem(product, sku, attributes, quantity)
}

const updateItemQuantity = (index, quantity) => {
  cartStore.updateQuantity(index, quantity)
}

const removeItem = (index) => {
  cartStore.removeItem(index)
}

const clearCart = async () => {
  try {
    await ElMessageBox.confirm('确定要清空购物车吗？', '确认', {
      type: 'warning'
    })
    cartStore.clear()
    ElMessage.success('已清空')
  } catch {}
}

const openPayment = () => {
  discountInput.value = 0
  paymentVisible.value = true
}

const handleSubmitOrder = async () => {
  try {
    cartStore.applyDiscount(discountInput.value)
    
    const order = await orderStore.createOrder(cartStore, {
      pay_method: payMethod.value,
      paid: true,
      table_no: cartStore.tableNo,
      remark: cartStore.remark
    })
    
    ElMessage.success(`订单 ${order.order_no} 创建成功`)
    paymentVisible.value = false
    cartStore.clear()
    
    if (!orderStore.isOnline.value) {
      ElMessage.info('当前离线，订单已本地保存，联网后自动同步')
    }
  } catch (e) {
    console.error('创建订单失败:', e)
    ElMessage.error('创建订单失败: ' + e.message)
  }
}

const handleSyncOrders = () => {
  orderStore.forceSync()
}

const handleFullSync = () => {
  syncStore.fullSync()
}

const handleIncrementalSync = () => {
  syncStore.incrementalSync()
}

const goOrders = () => {
  router.push('/orders')
}

const goStallReport = () => {
  router.push('/stall-report')
}

const goSettings = () => {
  router.push('/settings')
}

const formatDate = (date) => {
  return dayjs(date).format('YYYY-MM-DD HH:mm:ss')
}

onMounted(() => {
  loadData()
})
</script>

<style lang="scss" scoped>
.cashier-page {
  display: flex;
  flex-direction: column;
  height: 100vh;
  background: #f5f7fa;
}

.cashier-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 20px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: #fff;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  
  .header-left {
    display: flex;
    align-items: center;
    gap: 16px;
    
    .logo {
      margin: 0;
      font-size: 22px;
      font-weight: 700;
      letter-spacing: 2px;
    }
  }
  
  .header-right {
    display: flex;
    gap: 12px;
  }
}

.cashier-body {
  flex: 1;
  display: flex;
  overflow: hidden;
}

.category-sidebar {
  width: 160px;
  background: #fff;
  border-right: 1px solid #e4e7ed;
  overflow-y: auto;
  
  .category-title {
    padding: 16px;
    font-weight: 600;
    font-size: 15px;
    border-bottom: 1px solid #e4e7ed;
    background: #fafafa;
  }
  
  .category-item {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 14px 16px;
    cursor: pointer;
    border-left: 3px solid transparent;
    transition: all 0.2s;
    
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
      background: #f0f0f0;
      padding: 2px 8px;
      border-radius: 10px;
    }
  }
}

.product-panel {
  flex: 1;
  padding: 16px;
  overflow-y: auto;
  
  .product-list {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(180px, 1fr));
    gap: 16px;
  }
  
  .product-card {
    background: #fff;
    border-radius: 8px;
    overflow: hidden;
    cursor: pointer;
    transition: all 0.2s;
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);
    
    &:hover {
      transform: translateY(-2px);
      box-shadow: 0 4px 16px rgba(0, 0, 0, 0.12);
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
        padding: 2px 8px;
        font-size: 12px;
        border-radius: 4px;
        color: #fff;
        
        &.hot {
          left: 8px;
          background: #f56c6c;
        }
        
        &.recommend {
          left: 8px;
          background: #e6a23c;
        }
        
        &.offline {
          right: 8px;
          background: #909399;
        }
      }
    }
    
    .product-info {
      padding: 12px;
      
      .product-name {
        font-size: 15px;
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
        margin-bottom: 6px;
        
        .price {
          color: #f56c6c;
          font-size: 18px;
          font-weight: 600;
        }
        
        .multi-sku {
          font-size: 12px;
          color: #909399;
          background: #f0f0f0;
          padding: 2px 6px;
          border-radius: 4px;
        }
      }
      
      .product-stock {
        font-size: 12px;
        color: #67c23a;
        
        &.low-stock {
          color: #e6a23c;
        }
      }
    }
  }
}

.cart-panel {
  width: 380px;
  background: #fff;
  border-left: 1px solid #e4e7ed;
  display: flex;
  flex-direction: column;
  
  .cart-header {
    display: flex;
    align-items: center;
    padding: 16px;
    border-bottom: 1px solid #e4e7ed;
    background: #fafafa;
    
    .cart-title {
      font-size: 16px;
      font-weight: 600;
      flex: 1;
    }
    
    .cart-count {
      margin-right: 16px;
      color: #909399;
    }
  }
  
  .cart-info {
    padding: 12px 16px;
    border-bottom: 1px solid #f0f0f0;
  }
  
  .cart-items {
    flex: 1;
    overflow-y: auto;
    padding: 8px 0;
    
    .cart-item {
      padding: 12px 16px;
      border-bottom: 1px solid #f5f5f5;
      
      .item-info {
        margin-bottom: 8px;
        
        .item-name {
          font-weight: 500;
          margin-bottom: 4px;
        }
        
        .item-spec {
          font-size: 12px;
          color: #909399;
          margin-bottom: 4px;
        }
        
        .item-price {
          font-size: 13px;
          color: #606266;
        }
      }
      
      .item-actions {
        display: flex;
        align-items: center;
        gap: 12px;
        
        .item-subtotal {
          flex: 1;
          text-align: right;
          color: #f56c6c;
          font-weight: 600;
        }
      }
    }
  }
  
  .cart-footer {
    border-top: 1px solid #e4e7ed;
    padding: 16px;
    background: #fafafa;
    
    .cart-summary {
      margin-bottom: 16px;
      
      .summary-row {
        display: flex;
        justify-content: space-between;
        margin-bottom: 8px;
        
        .total {
          font-size: 16px;
          color: #f56c6c;
          text-decoration: line-through;
        }
        
        &.discount {
          color: #67c23a;
        }
        
        &.actual {
          font-weight: 600;
          padding-top: 8px;
          border-top: 1px dashed #e4e7ed;
          
          .actual-total {
            font-size: 24px;
            color: #f56c6c;
          }
        }
      }
    }
    
    .cart-actions {
      .el-button {
        width: 100%;
        height: 48px;
        font-size: 16px;
      }
    }
  }
}

.payment-dialog {
  .payment-amount {
    text-align: center;
    padding: 24px;
    background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
    border-radius: 8px;
    margin-bottom: 24px;
    color: #fff;
    
    .label {
      font-size: 14px;
      opacity: 0.9;
      margin-bottom: 8px;
    }
    
    .amount {
      font-size: 42px;
      font-weight: 700;
    }
  }
  
  .payment-summary {
    background: #f5f7fa;
    padding: 16px;
    border-radius: 8px;
    margin-top: 16px;
    
    .row {
      display: flex;
      justify-content: space-between;
      margin-bottom: 8px;
      
      &.discount {
        color: #67c23a;
      }
      
      &.actual {
        font-weight: 600;
        padding-top: 8px;
        border-top: 1px dashed #dcdfe6;
        
        .actual {
          font-size: 20px;
          color: #f56c6c;
        }
      }
    }
  }
}

.sync-dialog {
  .sync-info {
    margin-bottom: 24px;
  }
  
  .sync-actions {
    display: flex;
    gap: 16px;
    justify-content: center;
    
    .el-button {
      min-width: 140px;
    }
  }
}
</style>
