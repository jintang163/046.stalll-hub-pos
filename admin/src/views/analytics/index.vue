<template>
  <div class="analytics-page">
    <div class="page-header">
      <h2 class="page-title">营业报表分析</h2>
      <div class="header-actions">
        <el-radio-group v-model="reportType" @change="fetchAllData">
          <el-radio-button value="daily">日报</el-radio-button>
          <el-radio-button value="monthly">月报</el-radio-button>
        </el-radio-group>
        <el-date-picker
          v-model="dateRange"
          type="daterange"
          range-separator="至"
          start-placeholder="开始日期"
          end-placeholder="结束日期"
          value-format="YYYY-MM-DD"
          style="width: 280px" />
        <el-select v-model="storeId" placeholder="选择门店" clearable style="width: 160px" @change="fetchAllData">
          <el-option label="全部门店" :value="0" />
          <el-option v-for="store in storeList" :key="store.id" :label="store.name" :value="store.id" />
        </el-select>
        <el-button type="primary" @click="fetchAllData">
          <el-icon><Search /></el-icon>查询
        </el-button>
        <el-button type="warning" @click="handleBackfill" v-if="!backfilling">
          <el-icon><Refresh /></el-icon>全量回填
        </el-button>
        <el-button type="warning" disabled v-else>
          <el-icon class="is-loading" :size="14"><Loading /></el-icon>回填中...
        </el-button>
      </div>
    </div>

    <div class="summary-cards">
      <el-row :gutter="20">
        <el-col :span="6">
          <div class="summary-card revenue">
            <div class="card-icon">
              <el-icon :size="28"><Money /></el-icon>
            </div>
            <div class="card-info">
              <div class="card-label">总营业额</div>
              <div class="card-value">¥{{ formatAmount(summary.totalRevenue) }}</div>
            </div>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="summary-card orders">
            <div class="card-icon">
              <el-icon :size="28"><List /></el-icon>
            </div>
            <div class="card-info">
              <div class="card-label">订单数</div>
              <div class="card-value">{{ summary.orderCount }}</div>
            </div>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="summary-card avg">
            <div class="card-icon">
              <el-icon :size="28"><User /></el-icon>
            </div>
            <div class="card-info">
              <div class="card-label">客单价</div>
              <div class="card-value">¥{{ formatAmount(summary.avgOrderAmount) }}</div>
            </div>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="summary-card products">
            <div class="card-icon">
              <el-icon :size="28"><Goods /></el-icon>
            </div>
            <div class="card-info">
              <div class="card-label">热门菜品</div>
              <div class="card-value">{{ topProducts.length }}</div>
            </div>
          </div>
        </el-col>
      </el-row>
    </div>

    <el-row :gutter="20" class="chart-row">
      <el-col :span="16">
        <div class="chart-container">
          <div class="chart-title">时段营业趋势</div>
          <div ref="hourlyChartRef" style="height: 400px;"></div>
        </div>
      </el-col>
      <el-col :span="8">
        <div class="chart-container">
          <div class="chart-title">热门菜品 TOP10</div>
          <div ref="topProductsChartRef" style="height: 400px;"></div>
        </div>
      </el-col>
    </el-row>

    <el-row :gutter="20" class="chart-row">
      <el-col :span="24">
        <div class="chart-container">
          <div class="chart-title">{{ reportType === 'monthly' ? '月度' : '每日' }}营业明细</div>
          <el-table :data="revenueList" v-loading="loading" border stripe>
            <el-table-column prop="report_date" label="日期" width="140" align="center" />
            <el-table-column prop="store_id" label="门店ID" width="100" align="center">
              <template #default="{ row }">
                {{ getStoreName(row.store_id) }}
              </template>
            </el-table-column>
            <el-table-column prop="total_revenue" label="总营业额(元)" width="180" align="center">
              <template #default="{ row }">
                <span class="amount">¥{{ formatAmount(row.total_revenue) }}</span>
              </template>
            </el-table-column>
            <el-table-column prop="order_count" label="订单数" width="120" align="center" />
            <el-table-column prop="avg_order_amount" label="客单价(元)" width="160" align="center">
              <template #default="{ row }">
                ¥{{ formatAmount(row.avg_order_amount) }}
              </template>
            </el-table-column>
          </el-table>
        </div>
      </el-col>
    </el-row>

    <el-row :gutter="20" class="chart-row">
      <el-col :span="24">
        <div class="chart-container">
          <div class="chart-title">热门菜品排行</div>
          <el-table :data="topProducts" v-loading="loading" border stripe>
            <el-table-column type="index" label="排名" width="80" align="center">
              <template #default="{ $index }">
                <span :class="['rank-badge', $index < 3 ? 'top3' : '']">{{ $index + 1 }}</span>
              </template>
            </el-table-column>
            <el-table-column prop="product_name" label="菜品名称" min-width="200" />
            <el-table-column prop="quantity" label="销量" width="120" align="center">
              <template #default="{ row }">
                <span class="highlight">{{ row.quantity || 0 }}</span>
              </template>
            </el-table-column>
            <el-table-column prop="revenue" label="销售额(元)" width="160" align="center">
              <template #default="{ row }">
                <span class="amount">¥{{ formatAmount(row.revenue) }}</span>
              </template>
            </el-table-column>
          </el-table>
        </div>
      </el-col>
    </el-row>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, onBeforeUnmount } from 'vue'
import { ElMessage } from 'element-plus'
import { Search, Money, List, User, Goods, Refresh, Loading } from '@element-plus/icons-vue'
import { getRevenueReport, getHourlyTrend, getTopProducts, triggerBackfill } from '@/api/analytics'
import { storeApi } from '@/api/stores'
import * as echarts from 'echarts'

const loading = ref(false)
const backfilling = ref(false)
const reportType = ref('daily')
const storeId = ref(0)
const storeList = ref([])
const revenueList = ref([])
const topProducts = ref([])
const hourlyData = ref([])

const summary = reactive({
  totalRevenue: 0,
  orderCount: 0,
  avgOrderAmount: 0
})

const today = new Date()
const formatDate = (d) => {
  const year = d.getFullYear()
  const month = String(d.getMonth() + 1).padStart(2, '0')
  const day = String(d.getDate()).padStart(2, '0')
  return `${year}-${month}-${day}`
}

const dateRange = ref([
  formatDate(new Date(today.getFullYear(), today.getMonth(), 1)),
  formatDate(today)
])

const hourlyChartRef = ref(null)
const topProductsChartRef = ref(null)
let hourlyChart = null
let topProductsChart = null

function formatAmount(val) {
  if (!val && val !== 0) return '0.00'
  return Number(val).toFixed(2)
}

function getStoreName(id) {
  const store = storeList.value.find(s => s.id === id)
  return store ? store.name : `门店${id}`
}

function getQueryParams() {
  return {
    store_id: storeId.value || 0,
    start_date: dateRange.value?.[0] || '',
    end_date: dateRange.value?.[1] || '',
    report_type: reportType.value
  }
}

async function fetchStores() {
  try {
    const res = await storeApi.list({ page: 1, page_size: 100 })
    storeList.value = res?.list || res?.data || []
  } catch (e) {
    console.error('Failed to fetch stores:', e)
  }
}

async function fetchAllData() {
  loading.value = true
  try {
    await Promise.all([
      fetchRevenueReport(),
      fetchHourlyTrend(),
      fetchTopProducts()
    ])
  } finally {
    loading.value = false
  }
}

async function fetchRevenueReport() {
  try {
    const res = await getRevenueReport(getQueryParams())
    revenueList.value = res?.data?.list || res?.data || res || []
    let totalRev = 0
    let totalOrd = 0
    for (const r of revenueList.value) {
      totalRev += Number(r.total_revenue || 0)
      totalOrd += Number(r.order_count || 0)
    }
    summary.totalRevenue = totalRev
    summary.orderCount = totalOrd
    summary.avgOrderAmount = totalOrd > 0 ? totalRev / totalOrd : 0
  } catch (e) {
    console.error('fetchRevenueReport error:', e)
  }
}

async function fetchHourlyTrend() {
  try {
    const res = await getHourlyTrend(getQueryParams())
    hourlyData.value = res?.data || res || []
    renderHourlyChart()
  } catch (e) {
    console.error('fetchHourlyTrend error:', e)
  }
}

async function fetchTopProducts() {
  try {
    const params = { ...getQueryParams(), top_n: 10 }
    const res = await getTopProducts(params)
    topProducts.value = res?.data?.list || res?.data || res || []
    renderTopProductsChart()
  } catch (e) {
    console.error('fetchTopProducts error:', e)
  }
}

async function handleBackfill() {
  try {
    backfilling.value = true
    await triggerBackfill()
    ElMessage.success('全量回填任务已启动，请稍后刷新查看')
    setTimeout(() => {
      backfilling.value = false
    }, 30000)
  } catch (e) {
    backfilling.value = false
    ElMessage.error('启动回填失败')
  }
}

function renderHourlyChart() {
  if (!hourlyChartRef.value) return
  if (!hourlyChart) {
    hourlyChart = echarts.init(hourlyChartRef.value)
  }

  const hours = Array.from({ length: 24 }, (_, i) => `${String(i).padStart(2, '0')}:00`)
  const orderCounts = Array(24).fill(0)
  const revenues = Array(24).fill(0)

  for (const d of hourlyData.value) {
    const h = Number(d.hour)
    if (h >= 0 && h < 24) {
      orderCounts[h] = d.order_count || 0
      revenues[h] = Number(d.revenue || 0)
    }
  }

  hourlyChart.setOption({
    tooltip: {
      trigger: 'axis',
      axisPointer: { type: 'cross' }
    },
    legend: {
      data: ['营业额', '订单数']
    },
    grid: {
      left: '3%', right: '4%', bottom: '3%', containLabel: true
    },
    xAxis: {
      type: 'category',
      data: hours,
      axisLabel: { interval: 2 }
    },
    yAxis: [
      {
        type: 'value',
        name: '营业额(元)',
        position: 'left'
      },
      {
        type: 'value',
        name: '订单数',
        position: 'right'
      }
    ],
    series: [
      {
        name: '营业额',
        type: 'bar',
        data: revenues,
        itemStyle: {
          color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
            { offset: 0, color: '#409EFF' },
            { offset: 1, color: '#79bbff' }
          ])
        }
      },
      {
        name: '订单数',
        type: 'line',
        yAxisIndex: 1,
        data: orderCounts,
        smooth: true,
        lineStyle: { color: '#67c23a', width: 2 },
        itemStyle: { color: '#67c23a' },
        areaStyle: {
          color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
            { offset: 0, color: 'rgba(103, 194, 58, 0.3)' },
            { offset: 1, color: 'rgba(103, 194, 58, 0.05)' }
          ])
        }
      }
    ]
  })
}

function renderTopProductsChart() {
  if (!topProductsChartRef.value) return
  if (!topProductsChart) {
    topProductsChart = echarts.init(topProductsChartRef.value)
  }

  const names = topProducts.value.map(p => p.product_name).reverse()
  const revenues = topProducts.value.map(p => Number(p.revenue || 0)).reverse()

  topProductsChart.setOption({
    tooltip: {
      trigger: 'axis',
      axisPointer: { type: 'shadow' },
      formatter: (params) => {
        const p = params[0]
        return `${p.name}<br/>销售额: ¥${Number(p.value).toFixed(2)}`
      }
    },
    grid: {
      left: '3%', right: '15%', bottom: '3%', top: '3%', containLabel: true
    },
    xAxis: {
      type: 'value',
      axisLabel: {
        formatter: (val) => `¥${val}`
      }
    },
    yAxis: {
      type: 'category',
      data: names,
      axisLabel: {
        width: 80,
        overflow: 'truncate'
      }
    },
    series: [
      {
        type: 'bar',
        data: revenues,
        itemStyle: {
          color: (params) => {
            const colors = ['#f56c6c', '#e6a23c', '#5cb87a', '#409EFF', '#909399', '#b37feb', '#36cfc9', '#ff85c0', '#ffc53d', '#597ef7']
            return colors[params.dataIndex % colors.length]
          },
          borderRadius: [0, 4, 4, 0]
        },
        label: {
          show: true,
          position: 'right',
          formatter: (params) => `¥${Number(params.value).toFixed(0)}`
        }
      }
    ]
  })
}

function handleResize() {
  hourlyChart?.resize()
  topProductsChart?.resize()
}

onMounted(async () => {
  await fetchStores()
  await fetchAllData()
  window.addEventListener('resize', handleResize)
})

onBeforeUnmount(() => {
  window.removeEventListener('resize', handleResize)
  hourlyChart?.dispose()
  topProductsChart?.dispose()
})
</script>

<style scoped lang="scss">
.analytics-page {
  .page-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 20px;
    flex-wrap: wrap;
    gap: 12px;
  }

  .header-actions {
    display: flex;
    gap: 12px;
    align-items: center;
    flex-wrap: wrap;
  }

  .summary-cards {
    margin-bottom: 20px;
  }

  .summary-card {
    background: #fff;
    border-radius: 8px;
    padding: 20px;
    display: flex;
    align-items: center;
    gap: 16px;
    border-left: 4px solid #409eff;
    box-shadow: 0 1px 4px rgba(0, 0, 0, 0.06);

    &.revenue { border-left-color: #409eff; }
    &.orders { border-left-color: #67c23a; }
    &.avg { border-left-color: #e6a23c; }
    &.products { border-left-color: #f56c6c; }

    .card-icon {
      width: 56px;
      height: 56px;
      border-radius: 50%;
      display: flex;
      align-items: center;
      justify-content: center;
      color: #fff;
    }

    &.revenue .card-icon { background: linear-gradient(135deg, #409eff, #79bbff); }
    &.orders .card-icon { background: linear-gradient(135deg, #67c23a, #95d475); }
    &.avg .card-icon { background: linear-gradient(135deg, #e6a23c, #eebe77); }
    &.products .card-icon { background: linear-gradient(135deg, #f56c6c, #f89898); }

    .card-info {
      flex: 1;

      .card-label {
        color: #909399;
        font-size: 14px;
        margin-bottom: 6px;
      }

      .card-value {
        color: #303133;
        font-size: 24px;
        font-weight: 600;
      }
    }
  }

  .chart-row {
    margin-bottom: 20px;
  }

  .chart-container {
    background: #fff;
    border-radius: 8px;
    padding: 20px;
    box-shadow: 0 1px 4px rgba(0, 0, 0, 0.06);

    .chart-title {
      font-size: 16px;
      font-weight: 600;
      margin-bottom: 16px;
      color: #303133;
      padding-left: 10px;
      border-left: 3px solid #409eff;
    }
  }

  .amount {
    color: #f56c6c;
    font-weight: 600;
  }

  .highlight {
    color: #409eff;
    font-weight: 600;
  }

  .rank-badge {
    display: inline-block;
    width: 24px;
    height: 24px;
    line-height: 24px;
    text-align: center;
    border-radius: 4px;
    background: #f0f2f5;
    color: #606266;
    font-size: 12px;

    &.top3 {
      background: linear-gradient(135deg, #e6a23c, #f5c961);
      color: #fff;
      font-weight: 600;
    }
  }
}
</style>
