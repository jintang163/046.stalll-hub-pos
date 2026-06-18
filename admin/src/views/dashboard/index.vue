<template>
  <div class="dashboard">
    <el-row :gutter="20" class="stats-row">
      <el-col :span="6">
        <div class="stat-card" style="background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);">
          <div class="stat-icon">
            <el-icon><Wallet /></el-icon>
          </div>
          <div class="stat-content">
            <p class="stat-label">今日营业额</p>
            <p class="stat-value">¥{{ stats.todayRevenue }}</p>
          </div>
        </div>
      </el-col>
      <el-col :span="6">
        <div class="stat-card" style="background: linear-gradient(135deg, #f093fb 0%, #f5576c 100%);">
          <div class="stat-icon">
            <el-icon><List /></el-icon>
          </div>
          <div class="stat-content">
            <p class="stat-label">今日订单数</p>
            <p class="stat-value">{{ stats.todayOrders }}</p>
          </div>
        </div>
      </el-col>
      <el-col :span="6">
        <div class="stat-card" style="background: linear-gradient(135deg, #4facfe 0%, #00f2fe 100%);">
          <div class="stat-icon">
            <el-icon><Goods /></el-icon>
          </div>
          <div class="stat-content">
            <p class="stat-label">商品总数</p>
            <p class="stat-value">{{ stats.productCount }}</p>
          </div>
        </div>
      </el-col>
      <el-col :span="6">
        <div class="stat-card" style="background: linear-gradient(135deg, #43e97b 0%, #38f9d7 100%);">
          <div class="stat-icon">
            <el-icon><Warning /></el-icon>
          </div>
          <div class="stat-content">
            <p class="stat-label">库存预警</p>
            <p class="stat-value">{{ stats.stockWarning }}</p>
          </div>
        </div>
      </el-col>
    </el-row>

    <el-row :gutter="20" class="content-row">
      <el-col :span="16">
        <div class="card-wrapper">
          <div class="card-header">
            <h3>最近订单</h3>
            <el-button type="primary" link @click="$router.push('/orders')">查看全部</el-button>
          </div>
          <el-table :data="recentOrders" style="width: 100%">
            <el-table-column prop="order_no" label="订单号" />
            <el-table-column prop="table_no" label="桌号" />
            <el-table-column prop="total_amount" label="金额">
              <template #default="{ row }">¥{{ row.total_amount }}</template>
            </el-table-column>
            <el-table-column prop="pay_method" label="支付方式" />
            <el-table-column prop="status" label="状态">
              <template #default="{ row }">
                <el-tag :type="getStatusType(row.status)">{{ getStatusText(row.status) }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="created_at" label="时间" />
          </el-table>
        </div>
      </el-col>
      <el-col :span="8">
        <div class="card-wrapper">
          <div class="card-header">
            <h3>库存预警</h3>
          </div>
          <el-table :data="stockWarnings" style="width: 100%">
            <el-table-column prop="product.name" label="商品名称" />
            <el-table-column prop="sku.spec_name" label="规格" />
            <el-table-column prop="current_stock" label="当前库存">
              <template #default="{ row }">
                <span style="color: #f56c6c; font-weight: 600;">{{ row.current_stock }}</span>
              </template>
            </el-table-column>
            <el-table-column prop="threshold" label="预警值" />
          </el-table>
        </div>
      </el-col>
    </el-row>
  </div>
</template>

<script setup>
import { reactive, onMounted, ref } from 'vue'

const stats = reactive({
  todayRevenue: '0.00',
  todayOrders: 0,
  productCount: 0,
  stockWarning: 0
})

const recentOrders = ref([])
const stockWarnings = ref([])

onMounted(() => {
  stats.todayRevenue = '2,580.00'
  stats.todayOrders = 86
  stats.productCount = 128
  stats.stockWarning = 5

  recentOrders.value = [
    { order_no: '202401150001', table_no: 'A1', total_amount: '88.00', pay_method: '微信', status: 1, created_at: '12:30:25' },
    { order_no: '202401150002', table_no: 'B3', total_amount: '156.00', pay_method: '支付宝', status: 2, created_at: '12:28:10' },
    { order_no: '202401150003', table_no: 'C2', total_amount: '68.00', pay_method: '现金', status: 3, created_at: '12:25:33' },
    { order_no: '202401150004', table_no: 'A5', total_amount: '234.00', pay_method: '微信', status: 1, created_at: '12:20:15' },
    { order_no: '202401150005', table_no: 'B1', total_amount: '45.00', pay_method: '微信', status: 2, created_at: '12:15:42' }
  ]

  stockWarnings.value = [
    { product: { name: '香辣小龙虾' }, sku: { spec_name: '大份' }, current_stock: 3, threshold: 10 },
    { product: { name: '青岛啤酒' }, sku: { spec_name: '500ml' }, current_stock: 5, threshold: 20 },
    { product: { name: '烤羊肉串' }, sku: { spec_name: '10串' }, current_stock: 2, threshold: 15 }
  ]
})

function getStatusType(status) {
  const types = { 1: 'primary', 2: 'success', 3: 'warning' }
  return types[status] || 'info'
}

function getStatusText(status) {
  const texts = { 1: '待接单', 2: '制作中', 3: '已完成' }
  return texts[status] || '未知'
}
</script>

<style scoped lang="scss">
.dashboard {
  .stats-row {
    margin-bottom: 20px;
  }

  .stat-card {
    border-radius: 12px;
    padding: 24px;
    display: flex;
    align-items: center;
    gap: 20px;
    color: #fff;

    .stat-icon {
      width: 60px;
      height: 60px;
      background: rgba(255, 255, 255, 0.2);
      border-radius: 12px;
      display: flex;
      align-items: center;
      justify-content: center;
      font-size: 30px;
    }

    .stat-label {
      font-size: 14px;
      opacity: 0.9;
      margin-bottom: 4px;
    }

    .stat-value {
      font-size: 28px;
      font-weight: 700;
    }
  }

  .card-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 16px;

    h3 {
      font-size: 16px;
      font-weight: 600;
      color: #303133;
    }
  }
}
</style>
