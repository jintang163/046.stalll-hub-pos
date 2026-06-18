<template>
  <div class="orders-page">
    <header class="page-header">
      <h1>订单管理</h1>
      <div class="header-actions">
        <el-button @click="goBack">
          <el-icon><ArrowLeft /></el-icon>
          返回收银台
        </el-button>
        <el-button type="primary" @click="refresh">
          <el-icon><Refresh /></el-icon>
          刷新
        </el-button>
      </div>
    </header>

    <div class="page-body">
      <div class="filter-bar">
        <el-date-picker
          v-model="filterDate"
          type="date"
          placeholder="选择日期"
          format="YYYY-MM-DD"
          value-format="YYYY-MM-DD"
          @change="loadOrders"
        />
        <el-select v-model="filterStatus" placeholder="订单状态" clearable @change="loadOrders">
          <el-option label="全部" value="" />
          <el-option label="待接单" :value="0" />
          <el-option label="已接单" :value="1" />
          <el-option label="制作中" :value="2" />
          <el-option label="已完成" :value="3" />
          <el-option label="已取消" :value="4" />
        </el-select>
        <el-select v-model="filterPayStatus" placeholder="支付状态" clearable @change="loadOrders">
          <el-option label="全部" value="" />
          <el-option label="未支付" :value="0" />
          <el-option label="已支付" :value="1" />
          <el-option label="已退款" :value="2" />
        </el-select>
        <el-input 
          v-model="searchKeyword" 
          placeholder="搜索订单号/桌号"
          clearable
          @keyup.enter="loadOrders"
          style="width: 200px;"
        >
          <template #prefix>
            <el-icon><Search /></el-icon>
          </template>
        </el-input>
      </div>

      <div class="orders-stats">
        <el-row :gutter="16">
          <el-col :span="6">
            <div class="stat-card">
              <div class="stat-label">今日订单</div>
              <div class="stat-value">{{ orders.length }}</div>
            </div>
          </el-col>
          <el-col :span="6">
            <div class="stat-card success">
              <div class="stat-label">已完成</div>
              <div class="stat-value">{{ completedCount }}</div>
            </div>
          </el-col>
          <el-col :span="6">
            <div class="stat-card warning">
              <div class="stat-label">待同步</div>
              <div class="stat-value">{{ pendingCount }}</div>
            </div>
          </el-col>
          <el-col :span="6">
            <div class="stat-card primary">
              <div class="stat-label">今日营收</div>
              <div class="stat-value">¥{{ todayRevenue.toFixed(2) }}</div>
            </div>
          </el-col>
        </el-row>
      </div>

      <div class="orders-list">
        <el-table 
          :data="filteredOrders" 
          v-loading="loading"
          stripe
          border
        >
          <el-table-column prop="order_no" label="订单号" width="200" />
          <el-table-column prop="table_no" label="桌号" width="80" />
          <el-table-column label="商品" min-width="200">
            <template #default="{ row }">
              <div v-for="item in row.items.slice(0, 2)" :key="item.id" class="order-item-mini">
                {{ item.product_name }} x{{ item.quantity }}
              </div>
              <div v-if="row.items.length > 2" class="more-items">
                等{{ row.items.length }}件商品
              </div>
            </template>
          </el-table-column>
          <el-table-column label="金额" width="120">
            <template #default="{ row }">
              <div class="amount">¥{{ row.actual_amount.toFixed(2) }}</div>
              <div v-if="row.discount_amount > 0" class="discount">
                优惠 ¥{{ row.discount_amount.toFixed(2) }}
              </div>
            </template>
          </el-table-column>
          <el-table-column prop="pay_method" label="支付方式" width="100">
            <template #default="{ row }">
              <el-tag v-if="row.pay_method" size="small">
                {{ payMethodText(row.pay_method) }}
              </el-tag>
              <span v-else>-</span>
            </template>
          </el-table-column>
          <el-table-column label="状态" width="100">
            <template #default="{ row }">
              <el-tag :type="statusType(row.status)" size="small">
                {{ statusText(row.status) }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column label="同步" width="80">
            <template #default="{ row }">
              <el-tag :type="row.synced ? 'success' : 'warning'" size="small">
                {{ row.synced ? '已同步' : '待同步' }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="created_at" label="时间" width="160" />
          <el-table-column label="操作" width="180" fixed="right">
            <template #default="{ row }">
              <el-button size="small" @click="viewDetail(row)">
                详情
              </el-button>
              <el-button 
                size="small" 
                type="primary"
                :disabled="row.synced"
                @click="syncOrder(row)"
              >
                同步
              </el-button>
            </template>
          </el-table-column>
        </el-table>

        <el-empty v-if="!loading && filteredOrders.length === 0" description="暂无订单" />
      </div>
    </div>

    <el-dialog v-model="detailVisible" title="订单详情" width="600px">
      <div v-if="currentOrder" class="order-detail">
        <el-descriptions :column="2" border>
          <el-descriptions-item label="订单号">
            {{ currentOrder.order_no }}
          </el-descriptions-item>
          <el-descriptions-item label="状态">
            <el-tag :type="statusType(currentOrder.status)">
              {{ statusText(currentOrder.status) }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="桌号">
            {{ currentOrder.table_no || '-' }}
          </el-descriptions-item>
          <el-descriptions-item label="支付方式">
            {{ payMethodText(currentOrder.pay_method) }}
          </el-descriptions-item>
          <el-descriptions-item label="支付状态">
            <el-tag :type="currentOrder.pay_status ? 'success' : 'warning'">
              {{ currentOrder.pay_status ? '已支付' : '未支付' }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="同步状态">
            <el-tag :type="currentOrder.synced ? 'success' : 'warning'">
              {{ currentOrder.synced ? '已同步' : '待同步' }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="创建时间">
            {{ currentOrder.created_at }}
          </el-descriptions-item>
          <el-descriptions-item label="支付时间">
            {{ currentOrder.paid_at || '-' }}
          </el-descriptions-item>
          <el-descriptions-item label="备注" :span="2">
            {{ currentOrder.remark || '-' }}
          </el-descriptions-item>
        </el-descriptions>

        <div class="detail-section">
          <h3>商品明细</h3>
          <el-table :data="currentOrder.items" border size="small">
            <el-table-column prop="product_name" label="商品" />
            <el-table-column prop="sku_name" label="规格" width="100" />
            <el-table-column label="属性" width="120">
              <template #default="{ row }">
                {{ row.attribute_names || '-' }}
              </template>
            </el-table-column>
            <el-table-column prop="price" label="单价" width="80">
              <template #default="{ row }">
                ¥{{ row.price.toFixed(2) }}
              </template>
            </el-table-column>
            <el-table-column prop="quantity" label="数量" width="60" />
            <el-table-column prop="subtotal" label="小计" width="80">
              <template #default="{ row }">
                ¥{{ row.subtotal.toFixed(2) }}
              </template>
            </el-table-column>
          </el-table>
        </div>

        <div class="detail-section">
          <h3>费用明细</h3>
          <div class="fee-list">
            <div class="fee-row">
              <span>商品合计:</span>
              <span>¥{{ currentOrder.total_amount.toFixed(2) }}</span>
            </div>
            <div v-if="currentOrder.discount_amount > 0" class="fee-row discount">
              <span>优惠:</span>
              <span>-¥{{ currentOrder.discount_amount.toFixed(2) }}</span>
            </div>
            <div class="fee-row total">
              <span>实收:</span>
              <span>¥{{ currentOrder.actual_amount.toFixed(2) }}</span>
            </div>
          </div>
        </div>
      </div>
      
      <template #footer>
        <el-button @click="detailVisible = false">关闭</el-button>
        <el-button 
          v-if="currentOrder && !currentOrder.synced"
          type="primary"
          @click="syncOrder(currentOrder)"
        >
          同步订单
        </el-button>
        <el-button 
          v-if="currentOrder && currentOrder.status < 3"
          type="warning"
          @click="updateStatus(currentOrder)"
        >
          更新状态
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { ArrowLeft, Refresh, Search } from '@element-plus/icons-vue'
import { useOrderStore } from '@/store/order'

const router = useRouter()
const orderStore = useOrderStore()

const loading = ref(false)
const orders = ref([])
const filterDate = ref(new Date().toISOString().split('T')[0])
const filterStatus = ref('')
const filterPayStatus = ref('')
const searchKeyword = ref('')
const detailVisible = ref(false)
const currentOrder = ref(null)

const filteredOrders = computed(() => {
  let result = [...orders.value]
  
  if (filterStatus.value !== '') {
    result = result.filter(o => o.status === filterStatus.value)
  }
  
  if (filterPayStatus.value !== '') {
    result = result.filter(o => o.pay_status === filterPayStatus.value)
  }
  
  if (searchKeyword.value) {
    const kw = searchKeyword.value.toLowerCase()
    result = result.filter(o => 
      o.order_no.toLowerCase().includes(kw) ||
      (o.table_no && o.table_no.toLowerCase().includes(kw))
    )
  }
  
  return result.sort((a, b) => new Date(b.created_at) - new Date(a.created_at))
})

const completedCount = computed(() => 
  orders.value.filter(o => o.status === 3).length
)

const pendingCount = computed(() => 
  orders.value.filter(o => !o.synced).length
)

const todayRevenue = computed(() => 
  orders.value
    .filter(o => o.status !== 4 && o.pay_status === 1)
    .reduce((sum, o) => sum + o.actual_amount, 0)
)

const statusText = (status) => {
  const texts = ['待接单', '已接单', '制作中', '已完成', '已取消']
  return texts[status] || '未知'
}

const statusType = (status) => {
  const types = ['warning', 'primary', 'info', 'success', 'danger']
  return types[status] || 'info'
}

const payMethodText = (method) => {
  const texts = {
    cash: '现金',
    wechat: '微信支付',
    alipay: '支付宝',
    card: '刷卡'
  }
  return texts[method] || method || '-'
}

const loadOrders = async () => {
  if (!window.electronAPI) return
  loading.value = true
  try {
    orders.value = await window.electronAPI.orders.getOrdersByDate(filterDate.value)
  } finally {
    loading.value = false
  }
}

const refresh = () => {
  loadOrders()
  orderStore.forceSync()
}

const viewDetail = async (order) => {
  currentOrder.value = await window.electronAPI.orders.getOrderByNo(order.order_no)
  detailVisible.value = true
}

const syncOrder = async (order) => {
  try {
    await orderStore.syncPendingOrders()
    ElMessage.success('同步成功')
    loadOrders()
    if (currentOrder.value) {
      currentOrder.value = await window.electronAPI.orders.getOrderByNo(order.order_no)
    }
  } catch (e) {
    ElMessage.error('同步失败: ' + e.message)
  }
}

const updateStatus = async (order) => {
  const newStatus = order.status < 3 ? order.status + 1 : 3
  if (window.electronAPI) {
    await window.electronAPI.orders.updateOrderStatus(order.order_no, newStatus)
    ElMessage.success(`状态已更新为: ${statusText(newStatus)}`)
    loadOrders()
    viewDetail(order)
  }
}

const goBack = () => {
  router.push('/')
}

onMounted(() => {
  loadOrders()
})
</script>

<style lang="scss" scoped>
.orders-page {
  display: flex;
  flex-direction: column;
  height: 100vh;
  background: #f5f7fa;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 24px;
  background: #fff;
  border-bottom: 1px solid #e4e7ed;
  
  h1 {
    margin: 0;
    font-size: 20px;
    font-weight: 600;
  }
  
  .header-actions {
    display: flex;
    gap: 12px;
  }
}

.page-body {
  flex: 1;
  padding: 20px 24px;
  overflow-y: auto;
}

.filter-bar {
  display: flex;
  gap: 16px;
  margin-bottom: 20px;
  padding: 16px;
  background: #fff;
  border-radius: 8px;
}

.orders-stats {
  margin-bottom: 20px;
  
  .stat-card {
    background: #fff;
    padding: 20px;
    border-radius: 8px;
    text-align: center;
    
    .stat-label {
      color: #909399;
      font-size: 14px;
      margin-bottom: 8px;
    }
    
    .stat-value {
      font-size: 28px;
      font-weight: 700;
      color: #303133;
    }
    
    &.success .stat-value { color: #67c23a; }
    &.warning .stat-value { color: #e6a23c; }
    &.primary .stat-value { color: #409eff; }
  }
}

.orders-list {
  background: #fff;
  border-radius: 8px;
  padding: 16px;
}

.order-item-mini {
  font-size: 13px;
  color: #606266;
  line-height: 1.6;
}

.more-items {
  font-size: 12px;
  color: #909399;
}

.amount {
  font-weight: 600;
  color: #f56c6c;
}

.discount {
  font-size: 12px;
  color: #67c23a;
}

.order-detail {
  .detail-section {
    margin-top: 24px;
    
    h3 {
      margin: 0 0 12px;
      font-size: 16px;
      font-weight: 600;
    }
  }
  
  .fee-list {
    background: #f5f7fa;
    padding: 16px;
    border-radius: 8px;
    
    .fee-row {
      display: flex;
      justify-content: space-between;
      margin-bottom: 8px;
      
      &.discount {
        color: #67c23a;
      }
      
      &.total {
        font-weight: 600;
        padding-top: 8px;
        border-top: 1px dashed #dcdfe6;
        font-size: 16px;
        
        span:last-child {
          color: #f56c6c;
        }
      }
    }
  }
}
</style>
