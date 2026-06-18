<template>
  <div class="reports-page">
    <div class="page-header">
      <h2 class="page-title">营业报表</h2>
      <div class="header-actions">
        <el-date-picker
          v-model="dateRange"
          type="daterange"
          range-separator="至"
          start-placeholder="开始日期"
          end-placeholder="结束日期"
          value-format="YYYY-MM-DD"
          style="width: 280px" />
        <el-select v-model="storeId" placeholder="选择门店" clearable style="width: 160px">
          <el-option label="全部门店" :value="0" />
          <el-option v-for="store in storeList" :key="store.id" :label="store.name" :value="store.id" />
        </el-select>
        <el-button type="primary" @click="fetchCurrentReport">
          <el-icon><Search /></el-icon>查询
        </el-button>
        <el-button @click="exportReport">
          <el-icon><Download /></el-icon>导出
        </el-button>
      </div>
    </div>

    <div class="card-wrapper">
      <el-tabs v-model="activeTab" @tab-change="handleTabChange">
        <el-tab-pane label="营业日报" name="daily">
          <div class="summary-cards">
            <el-row :gutter="20">
              <el-col :span="6">
                <div class="summary-card">
                  <div class="card-label">营业额</div>
                  <div class="card-value">¥{{ (dailySummary?.total_amount || 0).toFixed(2) }}</div>
                  <div class="card-change" :class="dailySummary?.amount_change >= 0 ? 'up' : 'down'">
                    {{ dailySummary?.amount_change >= 0 ? '+' : '' }}{{ dailySummary?.amount_change || 0 }}%
                  </div>
                </div>
              </el-col>
              <el-col :span="6">
                <div class="summary-card">
                  <div class="card-label">订单数</div>
                  <div class="card-value">{{ dailySummary?.order_count || 0 }}</div>
                  <div class="card-change" :class="dailySummary?.order_change >= 0 ? 'up' : 'down'">
                    {{ dailySummary?.order_change >= 0 ? '+' : '' }}{{ dailySummary?.order_change || 0 }}%
                  </div>
                </div>
              </el-col>
              <el-col :span="6">
                <div class="summary-card">
                  <div class="card-label">客单价</div>
                  <div class="card-value">¥{{ (dailySummary?.avg_amount || 0).toFixed(2) }}</div>
                  <div class="card-change" :class="dailySummary?.avg_change >= 0 ? 'up' : 'down'">
                    {{ dailySummary?.avg_change >= 0 ? '+' : '' }}{{ dailySummary?.avg_change || 0 }}%
                  </div>
                </div>
              </el-col>
              <el-col :span="6">
                <div class="summary-card">
                  <div class="card-label">退款金额</div>
                  <div class="card-value refund">¥{{ (dailySummary?.refund_amount || 0).toFixed(2) }}</div>
                  <div class="card-change">退款 {{ dailySummary?.refund_count || 0 }} 笔</div>
                </div>
              </el-col>
            </el-row>
          </div>

          <el-table :data="dailyList" v-loading="loading" border>
            <el-table-column prop="date" label="日期" width="120" />
            <el-table-column prop="order_count" label="订单数" width="100" align="center" />
            <el-table-column prop="total_amount" label="营业额(元)" width="140" align="center">
              <template #default="{ row }">
                <span class="amount">¥{{ row.total_amount?.toFixed(2) }}</span>
              </template>
            </el-table-column>
            <el-table-column prop="avg_amount" label="客单价(元)" width="140" align="center">
              <template #default="{ row }">
                ¥{{ row.avg_amount?.toFixed(2) }}
              </template>
            </el-table-column>
            <el-table-column prop="refund_count" label="退款数" width="100" align="center" />
            <el-table-column prop="refund_amount" label="退款金额(元)" width="140" align="center">
              <template #default="{ row }">
                <span class="refund">-¥{{ row.refund_amount?.toFixed(2) }}</span>
              </template>
            </el-table-column>
            <el-table-column prop="net_amount" label="实收金额(元)" width="140" align="center">
              <template #default="{ row }">
                <span class="amount">¥{{ row.net_amount?.toFixed(2) }}</span>
              </template>
            </el-table-column>
            <el-table-column label="同比" width="100" align="center">
              <template #default="{ row }">
                <span :class="row.yoy >= 0 ? 'text-green' : 'text-red'">
                  {{ row.yoy >= 0 ? '+' : '' }}{{ row.yoy || 0 }}%
                </span>
              </template>
            </el-table-column>
            <el-table-column label="环比" width="100" align="center">
              <template #default="{ row }">
                <span :class="row.mom >= 0 ? 'text-green' : 'text-red'">
                  {{ row.mom >= 0 ? '+' : '' }}{{ row.mom || 0 }}%
                </span>
              </template>
            </el-table-column>
          </el-table>
        </el-tab-pane>

        <el-tab-pane label="商品销售" name="product">
          <el-table :data="productSalesList" v-loading="loading" border>
            <el-table-column type="index" label="排名" width="80" align="center" />
            <el-table-column prop="product_name" label="商品名称" min-width="180" />
            <el-table-column prop="category_name" label="分类" width="120" />
            <el-table-column prop="quantity" label="销量" width="100" align="center">
              <template #default="{ row }">
                <span class="highlight">{{ row.quantity || 0 }}</span>
              </template>
            </el-table-column>
            <el-table-column prop="amount" label="销售额(元)" width="140" align="center">
              <template #default="{ row }">
                <span class="amount">¥{{ row.amount?.toFixed(2) }}</span>
              </template>
            </el-table-column>
            <el-table-column prop="refund_quantity" label="退款量" width="100" align="center" />
            <el-table-column prop="refund_amount" label="退款金额(元)" width="140" align="center">
              <template #default="{ row }">
                <span class="refund">-¥{{ row.refund_amount?.toFixed(2) }}</span>
              </template>
            </el-table-column>
            <el-table-column prop="net_amount" label="实收金额(元)" width="140" align="center">
              <template #default="{ row }">
                <span class="amount">¥{{ row.net_amount?.toFixed(2) }}</span>
              </template>
            </el-table-column>
            <el-table-column label="占比" width="120">
              <template #default="{ row }">
                <el-progress
                  :percentage="row.percentage || 0"
                  :stroke-width="10"
                  :show-text="true" />
              </template>
            </el-table-column>
          </el-table>
        </el-tab-pane>

        <el-tab-pane label="分类销售" name="category">
          <el-row :gutter="20" style="margin-bottom: 20px;">
            <el-col :span="12">
              <div class="chart-container">
                <div class="chart-title">分类销售占比</div>
                <el-table :data="categorySalesList" v-loading="loading" border>
                  <el-table-column type="index" label="#" width="60" align="center" />
                  <el-table-column prop="category_name" label="分类名称" />
                  <el-table-column prop="amount" label="销售额(元)" align="center">
                    <template #default="{ row }">
                      <span class="amount">¥{{ row.amount?.toFixed(2) }}</span>
                    </template>
                  </el-table-column>
                  <el-table-column prop="percentage" label="占比" width="120">
                    <template #default="{ row }">
                      <el-progress
                        :percentage="row.percentage || 0"
                        :stroke-width="8"
                        :color="getCategoryColor(row.category_name)" />
                    </template>
                  </el-table-column>
                </el-table>
              </div>
            </el-col>
            <el-col :span="12">
              <div class="chart-container">
                <div class="chart-title">分类销量排行</div>
                <el-table :data="categorySalesList" v-loading="loading" border>
                  <el-table-column type="index" label="#" width="60" align="center" />
                  <el-table-column prop="category_name" label="分类名称" />
                  <el-table-column prop="quantity" label="销量" align="center">
                    <template #default="{ row }">
                      <span class="highlight">{{ row.quantity || 0 }}</span>
                    </template>
                  </el-table-column>
                  <el-table-column prop="order_count" label="订单数" align="center" />
                </el-table>
              </div>
            </el-col>
          </el-row>
        </el-tab-pane>

        <el-tab-pane label="时段销售" name="timeslot">
          <el-table :data="timeslotList" v-loading="loading" border>
            <el-table-column prop="time_slot" label="时段" width="140" />
            <el-table-column prop="order_count" label="订单数" width="100" align="center" />
            <el-table-column prop="quantity" label="商品销量" width="120" align="center">
              <template #default="{ row }">
                <span class="highlight">{{ row.quantity || 0 }}</span>
              </template>
            </el-table-column>
            <el-table-column prop="amount" label="销售额(元)" width="140" align="center">
              <template #default="{ row }">
                <span class="amount">¥{{ row.amount?.toFixed(2) }}</span>
              </template>
            </el-table-column>
            <el-table-column prop="avg_amount" label="客单价(元)" width="140" align="center">
              <template #default="{ row }">
                ¥{{ row.avg_amount?.toFixed(2) }}
              </template>
            </el-table-column>
            <el-table-column label="销售趋势" min-width="200">
              <template #default="{ row }">
                <el-progress
                  :percentage="row.percentage || 0"
                  :stroke-width="12"
                  :color="getTimeSlotColor(row.percentage)"
                  :show-text="false" />
              </template>
            </el-table-column>
          </el-table>
        </el-tab-pane>

        <el-tab-pane label="支付统计" name="payment">
          <div class="payment-summary">
            <el-row :gutter="20">
              <el-col v-for="item in paymentSummary" :key="item.payment_type" :span="6">
                <div class="payment-card" :style="{ borderColor: item.color }">
                  <div class="payment-icon" :style="{ background: item.color }">
                    <el-icon><Money /></el-icon>
                  </div>
                  <div class="payment-info">
                    <div class="payment-name">{{ item.payment_name }}</div>
                    <div class="payment-count">{{ item.order_count || 0 }} 笔</div>
                    <div class="payment-amount">¥{{ item.amount?.toFixed(2) }}</div>
                  </div>
                </div>
              </el-col>
            </el-row>
          </div>

          <el-table :data="paymentList" v-loading="loading" border>
            <el-table-column prop="payment_name" label="支付方式" width="140" />
            <el-table-column prop="order_count" label="订单数" width="120" align="center" />
            <el-table-column prop="amount" label="金额(元)" width="140" align="center">
              <template #default="{ row }">
                <span class="amount">¥{{ row.amount?.toFixed(2) }}</span>
              </template>
            </el-table-column>
            <el-table-column prop="refund_count" label="退款数" width="100" align="center" />
            <el-table-column prop="refund_amount" label="退款金额(元)" width="140" align="center">
              <template #default="{ row }">
                <span class="refund">-¥{{ row.refund_amount?.toFixed(2) }}</span>
              </template>
            </el-table-column>
            <el-table-column prop="net_amount" label="实收金额(元)" width="140" align="center">
              <template #default="{ row }">
                <span class="amount">¥{{ row.net_amount?.toFixed(2) }}</span>
              </template>
            </el-table-column>
            <el-table-column label="占比" width="150">
              <template #default="{ row }">
                <el-progress
                  :percentage="row.percentage || 0"
                  :stroke-width="10"
                  :show-text="true" />
              </template>
            </el-table-column>
          </el-table>
        </el-tab-pane>
      </el-tabs>

      <div class="pagination">
        <el-pagination
          v-model:current-page="query.page"
          v-model:page-size="query.page_size"
          :total="total"
          :page-sizes="[10, 20, 50, 100]"
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="fetchCurrentReport"
          @current-change="fetchCurrentReport" />
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { Search, Download, Money } from '@element-plus/icons-vue'
import {
  getDailyReport,
  getProductSalesReport,
  getCategorySalesReport,
  getTimeSlotReport,
  getPaymentStatsReport
} from '@/api/reports'

const loading = ref(false)
const activeTab = ref('daily')
const total = ref(0)

const today = new Date()
const lastMonth = new Date(today.getTime() - 30 * 24 * 60 * 60 * 1000)
const formatDate = (d) => {
  const year = d.getFullYear()
  const month = String(d.getMonth() + 1).padStart(2, '0')
  const day = String(d.getDate()).padStart(2, '0')
  return `${year}-${month}-${day}`
}

const dateRange = ref([formatDate(lastMonth), formatDate(today)])
const storeId = ref(0)

const storeList = ref([
  { id: 1, name: '总店' },
  { id: 2, name: '分店A' },
  { id: 3, name: '分店B' }
])

const query = reactive({
  page: 1,
  page_size: 10,
  start_date: '',
  end_date: '',
  store_id: 0
})

const dailySummary = ref({})
const dailyList = ref([])
const productSalesList = ref([])
const categorySalesList = ref([])
const timeslotList = ref([])
const paymentList = ref([])

const paymentSummary = computed(() => {
  return paymentList.value.slice(0, 4).map((item, index) => ({
    ...item,
    color: ['#409eff', '#67c23a', '#e6a23c', '#f56c6c'][index]
  }))
})

const categoryColors = ['#409eff', '#67c23a', '#e6a23c', '#f56c6c', '#909399', '#0693e3']

function getCategoryColor(name) {
  const index = categorySalesList.value.findIndex(c => c.category_name === name)
  return categoryColors[index % categoryColors.length]
}

function getTimeSlotColor(percentage) {
  if (percentage >= 20) return '#67c23a'
  if (percentage >= 10) return '#409eff'
  if (percentage >= 5) return '#e6a23c'
  return '#909399'
}

function handleTabChange(tab) {
  query.page = 1
  fetchCurrentReport(tab)
}

function fetchCurrentReport(tab = activeTab.value) {
  if (dateRange.value && dateRange.value.length === 2) {
    query.start_date = dateRange.value[0]
    query.end_date = dateRange.value[1]
  }
  query.store_id = storeId.value

  switch (tab) {
    case 'daily':
      fetchDailyReport()
      break
    case 'product':
      fetchProductSales()
      break
    case 'category':
      fetchCategorySales()
      break
    case 'timeslot':
      fetchTimeSlot()
      break
    case 'payment':
      fetchPaymentStats()
      break
  }
}

async function fetchDailyReport() {
  loading.value = true
  try {
    const res = await getDailyReport(query)
    dailyList.value = res.list || []
    total.value = res.total || 0
    dailySummary.value = res.summary || {}
  } catch (e) {
    console.error(e)
  } finally {
    loading.value = false
  }
}

async function fetchProductSales() {
  loading.value = true
  try {
    const res = await getProductSalesReport(query)
    productSalesList.value = res.list || []
    total.value = res.total || 0
  } catch (e) {
    console.error(e)
  } finally {
    loading.value = false
  }
}

async function fetchCategorySales() {
  loading.value = true
  try {
    const res = await getCategorySalesReport(query)
    categorySalesList.value = res.list || []
    total.value = res.total || 0
  } catch (e) {
    console.error(e)
  } finally {
    loading.value = false
  }
}

async function fetchTimeSlot() {
  loading.value = true
  try {
    const res = await getTimeSlotReport(query)
    timeslotList.value = res.list || []
    total.value = res.total || 0
  } catch (e) {
    console.error(e)
  } finally {
    loading.value = false
  }
}

async function fetchPaymentStats() {
  loading.value = true
  try {
    const res = await getPaymentStatsReport(query)
    paymentList.value = res.list || []
    total.value = res.total || 0
  } catch (e) {
    console.error(e)
  } finally {
    loading.value = false
  }
}

function exportReport() {
  ElMessage.info('导出功能开发中')
}

onMounted(() => {
  fetchCurrentReport()
})
</script>

<style scoped lang="scss">
.reports-page {
  .header-actions {
    display: flex;
    gap: 12px;
    align-items: center;
  }

  .summary-cards {
    margin-bottom: 20px;
  }

  .summary-card {
    background: #fff;
    border: 1px solid #ebeef5;
    border-radius: 8px;
    padding: 20px;
    text-align: center;

    .card-label {
      color: #909399;
      font-size: 14px;
      margin-bottom: 8px;
    }

    .card-value {
      color: #303133;
      font-size: 28px;
      font-weight: 600;
      margin-bottom: 8px;

      &.refund {
        color: #f56c6c;
      }
    }

    .card-change {
      font-size: 13px;
      color: #909399;

      &.up {
        color: #67c23a;
      }

      &.down {
        color: #f56c6c;
      }
    }
  }

  .chart-container {
    background: #fff;
    border: 1px solid #ebeef5;
    border-radius: 8px;
    padding: 16px;

    .chart-title {
      font-size: 16px;
      font-weight: 600;
      margin-bottom: 16px;
      color: #303133;
    }
  }

  .payment-summary {
    margin-bottom: 20px;
  }

  .payment-card {
    display: flex;
    align-items: center;
    background: #fff;
    border: 2px solid #ebeef5;
    border-radius: 8px;
    padding: 20px;

    .payment-icon {
      width: 48px;
      height: 48px;
      border-radius: 50%;
      display: flex;
      align-items: center;
      justify-content: center;
      color: #fff;
      font-size: 24px;
      margin-right: 16px;
    }

    .payment-info {
      flex: 1;

      .payment-name {
        font-size: 14px;
        color: #606266;
        margin-bottom: 4px;
      }

      .payment-count {
        font-size: 12px;
        color: #909399;
        margin-bottom: 4px;
      }

      .payment-amount {
        font-size: 20px;
        font-weight: 600;
        color: #303133;
      }
    }
  }

  .amount {
    color: #f56c6c;
    font-weight: 600;
  }

  .refund {
    color: #67c23a;
  }

  .highlight {
    color: #409eff;
    font-weight: 600;
  }

  .text-green {
    color: #67c23a;
  }

  .text-red {
    color: #f56c6c;
  }

  .pagination {
    margin-top: 20px;
    display: flex;
    justify-content: flex-end;
  }
}
</style>
