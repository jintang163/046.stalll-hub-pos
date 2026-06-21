<template>
  <div class="profit-page">
    <div class="page-header">
      <h2 class="page-title">利润分析</h2>
      <div class="header-actions">
        <el-date-picker
          v-model="dateRange"
          type="daterange"
          range-separator="至"
          start-placeholder="开始日期"
          end-placeholder="结束日期"
          value-format="YYYY-MM-DD"
          style="width: 280px" />
        <el-select v-model="storeId" placeholder="选择门店" clearable style="width: 160px" @change="fetchData">
          <el-option label="全部门店" :value="0" />
          <el-option v-for="store in storeList" :key="store.id" :label="store.name" :value="store.id" />
        </el-select>
        <el-select v-model="costMode" placeholder="成本模式" style="width: 140px" @change="fetchData">
          <el-option label="静态成本" value="static" />
          <el-option label="BOM动态成本" value="bom" />
        </el-select>
        <el-button type="primary" @click="fetchData">
          <el-icon><Search /></el-icon>查询
        </el-button>
        <el-button type="success" @click="showImportDialog">
          <el-icon><Upload /></el-icon>导入成本
        </el-button>
      </div>
    </div>

    <div class="summary-cards">
      <el-row :gutter="20">
        <el-col :span="4">
          <div class="summary-card revenue">
            <div class="card-icon"><el-icon :size="24"><Money /></el-icon></div>
            <div class="card-info">
              <div class="card-label">总营业额</div>
              <div class="card-value">¥{{ formatAmount(profitSummary.total_revenue) }}</div>
            </div>
          </div>
        </el-col>
        <el-col :span="4">
          <div class="summary-card cost">
            <div class="card-icon"><el-icon :size="24"><ShoppingCart /></el-icon></div>
            <div class="card-info">
              <div class="card-label">食材成本</div>
              <div class="card-value">¥{{ formatAmount(profitSummary.total_material_cost || profitSummary.total_cost) }}</div>
            </div>
          </div>
        </el-col>
        <el-col :span="4">
          <div class="summary-card profit">
            <div class="card-icon"><el-icon :size="24"><TrendCharts /></el-icon></div>
            <div class="card-info">
              <div class="card-label">毛利润</div>
              <div class="card-value">¥{{ formatAmount(profitSummary.gross_profit) }}</div>
            </div>
          </div>
        </el-col>
        <el-col :span="4">
          <div class="summary-card margin">
            <div class="card-icon"><el-icon :size="24"><DataAnalysis /></el-icon></div>
            <div class="card-info">
              <div class="card-label">毛利率</div>
              <div class="card-value">{{ formatMargin(profitSummary.gross_margin) }}%</div>
            </div>
          </div>
        </el-col>
        <el-col :span="4">
          <div class="summary-card net-profit">
            <div class="card-icon"><el-icon :size="24"><GoldMedal /></el-icon></div>
            <div class="card-info">
              <div class="card-label">净利润</div>
              <div class="card-value">¥{{ formatAmount(profitSummary.net_profit || profitSummary.gross_profit) }}</div>
            </div>
          </div>
        </el-col>
        <el-col :span="4">
          <div class="summary-card net-margin">
            <div class="card-icon"><el-icon :size="24"><Histogram /></el-icon></div>
            <div class="card-info">
              <div class="card-label">净利率</div>
              <div class="card-value">{{ formatMargin(profitSummary.net_margin || profitSummary.gross_margin) }}%</div>
            </div>
          </div>
        </el-col>
      </el-row>
    </div>

    <el-row :gutter="20" class="chart-row">
      <el-col :span="12">
        <div class="chart-container">
          <div class="chart-title">营收与成本对比 TOP10</div>
          <div ref="compareChartRef" style="height: 380px;"></div>
        </div>
      </el-col>
      <el-col :span="12">
        <div class="chart-container">
          <div class="chart-title">毛利率分布 TOP10</div>
          <div ref="marginChartRef" style="height: 380px;"></div>
        </div>
      </el-col>
    </el-row>

    <div class="chart-container">
      <div class="chart-title">商品利润明细</div>
      <el-table :data="profitReport" v-loading="loading" border stripe>
        <el-table-column type="index" label="#" width="60" align="center" />
        <el-table-column prop="product_name" label="菜品名称" min-width="180" />
        <el-table-column prop="quantity" label="销量" width="90" align="center">
          <template #default="{ row }">
            <span class="highlight">{{ row.quantity || 0 }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="unit_price" label="单价" width="100" align="center">
          <template #default="{ row }">¥{{ formatAmount(row.unit_price) }}</template>
        </el-table-column>
        <el-table-column prop="unit_cost" label="单位成本" width="100" align="center">
          <template #default="{ row }">
            ¥{{ formatAmount(row.unit_cost) }}
          </template>
        </el-table-column>
        <el-table-column prop="revenue" label="营收(元)" width="120" align="center">
          <template #default="{ row }">
            <span class="amount">¥{{ formatAmount(row.revenue) }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="material_cost" label="食材成本(元)" width="120" align="center">
          <template #default="{ row }">
            <span class="cost-text">¥{{ formatAmount(row.material_cost || row.total_cost) }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="gross_profit" label="毛利(元)" width="110" align="center">
          <template #default="{ row }">
            <span :class="Number(row.gross_profit) >= 0 ? 'profit-text' : 'loss-text'">
              ¥{{ formatAmount(row.gross_profit) }}
            </span>
          </template>
        </el-table-column>
        <el-table-column prop="gross_margin" label="毛利率" width="160" align="center">
          <template #default="{ row }">
            <el-progress
              :percentage="clampPercent(Number(row.gross_margin))"
              :stroke-width="10"
              :color="getMarginColor(Number(row.gross_margin))"
              :format="() => formatMargin(row.gross_margin) + '%'" />
          </template>
        </el-table-column>
      </el-table>
    </div>

    <el-dialog v-model="importDialogVisible" title="导入成本数据" width="500px">
      <el-form label-width="100px">
        <el-form-item label="生效日期">
          <el-date-picker
            v-model="importDate"
            type="date"
            placeholder="选择生效日期"
            value-format="YYYY-MM-DD"
            style="width: 100%" />
        </el-form-item>
        <el-form-item label="成本文件">
          <el-upload
            ref="uploadRef"
            :auto-upload="false"
            :limit="1"
            accept=".xlsx,.xls"
            :on-change="handleFileChange">
            <el-button type="primary">选择Excel文件</el-button>
            <template #tip>
              <div class="upload-tip">
                Excel格式: 第1列=菜品名称, 第2列=单位成本, 第3列=售价(可选)<br/>
                支持表头跳过、自动识别列，列数不足会自动补默认值
              </div>
            </template>
          </el-upload>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="importDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="submitImport" :loading="importLoading">确认导入</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, onBeforeUnmount } from 'vue'
import { ElMessage } from 'element-plus'
import { Search, Upload, Money, ShoppingCart, TrendCharts, DataAnalysis, GoldMedal, Histogram } from '@element-plus/icons-vue'
import { getProfitReport, getProfitSummary, getProfitReportV2, getProfitSummaryV2, importCostExcel } from '@/api/analytics'
import { storeApi } from '@/api/stores'
import * as echarts from 'echarts'

const loading = ref(false)
const importLoading = ref(false)
const importDialogVisible = ref(false)
const importDate = ref('')
const storeId = ref(0)
const storeList = ref([])
const profitReport = ref([])
const uploadFile = ref(null)
const costMode = ref('static')

const profitSummary = reactive({
  total_revenue: 0,
  total_cost: 0,
  total_material_cost: 0,
  gross_profit: 0,
  gross_margin: 0,
  net_profit: 0,
  net_margin: 0,
  product_count: 0,
  order_count: 0
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

const compareChartRef = ref(null)
const marginChartRef = ref(null)
let compareChart = null
let marginChart = null

function formatAmount(val) {
  if (!val && val !== 0) return '0.00'
  return Number(val).toFixed(2)
}

function formatMargin(val) {
  if (!val && val !== 0) return '0.0'
  return Number(val).toFixed(1)
}

function clampPercent(val) {
  if (!val && val !== 0) return 0
  const v = Number(val)
  if (v < 0) return 0
  if (v > 100) return 100
  return v
}

function getMarginColor(margin) {
  if (margin >= 60) return '#67c23a'
  if (margin >= 40) return '#409eff'
  if (margin >= 20) return '#e6a23c'
  return '#f56c6c'
}

function showImportDialog() {
  importDate.value = formatDate(new Date())
  importDialogVisible.value = true
}

function handleFileChange(file) {
  uploadFile.value = file.raw
}

async function submitImport() {
  if (!uploadFile.value) {
    ElMessage.warning('请选择文件')
    return
  }
  if (!importDate.value) {
    ElMessage.warning('请选择生效日期')
    return
  }

  importLoading.value = true
  try {
    const formData = new FormData()
    formData.append('file', uploadFile.value)
    formData.append('effective_date', importDate.value)
    const res = await importCostExcel(formData)
    const data = res?.data || res
    ElMessage.success(`导入成功: 成功${data.success_count || 0}条, 失败${data.fail_count || 0}条`)
    importDialogVisible.value = false
    fetchData()
  } catch (e) {
    ElMessage.error('导入失败')
  } finally {
    importLoading.value = false
  }
}

function getQueryParams() {
  return {
    store_id: storeId.value || 0,
    start_date: dateRange.value?.[0] || '',
    end_date: dateRange.value?.[1] || ''
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

async function fetchData() {
  loading.value = true
  try {
    let summaryRes, reportRes
    if (costMode.value === 'bom') {
      [summaryRes, reportRes] = await Promise.all([
        getProfitSummaryV2(getQueryParams()),
        getProfitReportV2(getQueryParams())
      ])
    } else {
      [summaryRes, reportRes] = await Promise.all([
        getProfitSummary(getQueryParams()),
        getProfitReport(getQueryParams())
      ])
    }

    const s = summaryRes?.data || summaryRes || {}
    Object.assign(profitSummary, s)
    profitReport.value = reportRes?.data?.list || reportRes?.data || reportRes || []

    renderCompareChart()
    renderMarginChart()
  } catch (e) {
    console.error(e)
  } finally {
    loading.value = false
  }
}

function renderCompareChart() {
  if (!compareChartRef.value) return
  if (!compareChart) {
    compareChart = echarts.init(compareChartRef.value)
  }

  const top10 = profitReport.value.slice(0, 10)
  const names = top10.map(p => p.product_name)
  const revenues = top10.map(p => Number(p.revenue || 0))
  const costs = top10.map(p => Number(p.material_cost || p.total_cost || 0))

  compareChart.setOption({
    tooltip: {
      trigger: 'axis',
      axisPointer: { type: 'shadow' }
    },
    legend: {
      data: ['营收', '成本']
    },
    grid: {
      left: '3%', right: '4%', bottom: '3%', containLabel: true
    },
    xAxis: {
      type: 'category',
      data: names,
      axisLabel: {
        rotate: 30,
        width: 60,
        overflow: 'truncate'
      }
    },
    yAxis: {
      type: 'value',
      name: '金额(元)'
    },
    series: [
      {
        name: '营收',
        type: 'bar',
        data: revenues,
        itemStyle: {
          color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
            { offset: 0, color: '#409EFF' },
            { offset: 1, color: '#79bbff' }
          ]),
          borderRadius: [4, 4, 0, 0]
        }
      },
      {
        name: '成本',
        type: 'bar',
        data: costs,
        itemStyle: {
          color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
            { offset: 0, color: '#e6a23c' },
            { offset: 1, color: '#eebe77' }
          ]),
          borderRadius: [4, 4, 0, 0]
        }
      }
    ]
  })
}

function renderMarginChart() {
  if (!marginChartRef.value) return
  if (!marginChart) {
    marginChart = echarts.init(marginChartRef.value)
  }

  const data = profitReport.value.slice(0, 10).map(p => ({
    name: p.product_name,
    value: Number(p.gross_margin || 0)
  }))

  marginChart.setOption({
    tooltip: {
      trigger: 'item',
      formatter: '{b}: {c}%'
    },
    series: [
      {
        type: 'pie',
        radius: ['40%', '70%'],
        avoidLabelOverlap: true,
        itemStyle: {
          borderRadius: 6,
          borderColor: '#fff',
          borderWidth: 2
        },
        label: {
          show: true,
          formatter: '{b}\n{c}%',
          fontSize: 11
        },
        data: data,
        color: ['#67c23a', '#409eff', '#e6a23c', '#f56c6c', '#909399', '#b37feb', '#36cfc9', '#ff85c0', '#ffc53d', '#597ef7']
      }
    ]
  })
}

function handleResize() {
  compareChart?.resize()
  marginChart?.resize()
}

onMounted(async () => {
  await fetchStores()
  fetchData()
  window.addEventListener('resize', handleResize)
})

onBeforeUnmount(() => {
  window.removeEventListener('resize', handleResize)
  compareChart?.dispose()
  marginChart?.dispose()
})
</script>

<style scoped lang="scss">
.profit-page {
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
    &.cost { border-left-color: #e6a23c; }
    &.profit { border-left-color: #67c23a; }
    &.margin { border-left-color: #f56c6c; }
    &.net-profit { border-left-color: #722ed1; }
    &.net-margin { border-left-color: #13c2c2; }

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
    &.cost .card-icon { background: linear-gradient(135deg, #e6a23c, #eebe77); }
    &.profit .card-icon { background: linear-gradient(135deg, #67c23a, #95d475); }
    &.margin .card-icon { background: linear-gradient(135deg, #f56c6c, #f89898); }
    &.net-profit .card-icon { background: linear-gradient(135deg, #722ed1, #b37feb); }
    &.net-margin .card-icon { background: linear-gradient(135deg, #13c2c2, #5cd3d3); }

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
    margin-bottom: 20px;
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

  .cost-text {
    color: #e6a23c;
    font-weight: 600;
  }

  .profit-text {
    color: #67c23a;
    font-weight: 600;
  }

  .loss-text {
    color: #f56c6c;
    font-weight: 600;
  }

  .upload-tip {
    color: #909399;
    font-size: 12px;
    margin-top: 4px;
    line-height: 1.6;
  }
}
</style>
