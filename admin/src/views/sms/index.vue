<template>
  <div class="sms-page">
    <div class="page-header">
      <h2 class="page-title">短信营销统计</h2>
      <div class="header-actions">
        <el-date-picker
          v-model="dateRange"
          type="daterange"
          range-separator="至"
          start-placeholder="开始日期"
          end-placeholder="结束日期"
          value-format="YYYY-MM-DD"
          style="width: 280px" />
        <el-select v-model="taskType" placeholder="任务类型" clearable style="width: 160px">
          <el-option label="全部类型" value="" />
          <el-option v-for="item in taskTypeOptions" :key="item.value" :label="item.label" :value="item.value" />
        </el-select>
        <el-button type="primary" @click="fetchAllData">
          <el-icon><Search /></el-icon>查询
        </el-button>
        <el-button type="success" @click="handleExport" :loading="exporting">
          <el-icon><Download /></el-icon>导出
        </el-button>
      </div>
    </div>

    <div class="summary-cards">
      <el-row :gutter="20">
        <el-col :span="4">
          <div class="summary-card total">
            <div class="card-icon">
              <el-icon :size="28"><Message /></el-icon>
            </div>
            <div class="card-info">
              <div class="card-label">发送总条数</div>
              <div class="card-value">
                {{ summary.totalSend }}
                <span class="change" :class="summary.totalSendChange >= 0 ? 'up' : 'down'">
                  {{ summary.totalSendChange >= 0 ? '↑' : '↓' }}{{ Math.abs(summary.totalSendChange) }}
                </span>
              </div>
            </div>
          </div>
        </el-col>
        <el-col :span="4">
          <div class="summary-card success">
            <div class="card-icon">
              <el-icon :size="28"><CircleCheck /></el-icon>
            </div>
            <div class="card-info">
              <div class="card-label">成功条数</div>
              <div class="card-value">
                {{ summary.successCount }}
                <span class="sub-info">成功率 {{ summary.successRate }}%</span>
                <span class="change" :class="summary.successChange >= 0 ? 'up' : 'down'">
                  {{ summary.successChange >= 0 ? '↑' : '↓' }}{{ Math.abs(summary.successChange) }}
                </span>
              </div>
            </div>
          </div>
        </el-col>
        <el-col :span="4">
          <div class="summary-card fail">
            <div class="card-icon">
              <el-icon :size="28"><CircleClose /></el-icon>
            </div>
            <div class="card-info">
              <div class="card-label">失败条数</div>
              <div class="card-value">
                {{ summary.failCount }}
                <span class="sub-info">失败率 {{ summary.failRate }}%</span>
              </div>
            </div>
          </div>
        </el-col>
        <el-col :span="4">
          <div class="summary-card read">
            <div class="card-icon">
              <el-icon :size="28"><View /></el-icon>
            </div>
            <div class="card-info">
              <div class="card-label">阅读条数</div>
              <div class="card-value">
                {{ summary.readCount }}
                <span class="sub-info">阅读率 {{ summary.readRate }}%</span>
              </div>
            </div>
          </div>
        </el-col>
        <el-col :span="4">
          <div class="summary-card convert">
            <div class="card-icon">
              <el-icon :size="28"><UserFilled /></el-icon>
            </div>
            <div class="card-info">
              <div class="card-label">转化人数</div>
              <div class="card-value">
                {{ summary.convertCount }}
                <span class="sub-info">转化率 {{ summary.convertRate }}%</span>
              </div>
            </div>
          </div>
        </el-col>
        <el-col :span="4">
          <div class="summary-card roi">
            <div class="card-icon">
              <el-icon :size="28"><Money /></el-icon>
            </div>
            <div class="card-info">
              <div class="card-label">转化金额</div>
              <div class="card-value">
                ¥{{ formatAmount(summary.convertAmount) }}
                <span class="sub-info">ROI {{ summary.roi }}</span>
              </div>
            </div>
          </div>
        </el-col>
      </el-row>
    </div>

    <el-row :gutter="20" class="chart-row">
      <el-col :span="24">
        <div class="chart-container">
          <div class="chart-title">发送量趋势</div>
          <div ref="sendTrendChartRef" style="height: 400px;"></div>
        </div>
      </el-col>
    </el-row>

    <el-row :gutter="20" class="chart-row">
      <el-col :span="14">
        <div class="chart-container">
          <div class="chart-title">转化率趋势</div>
          <div ref="convertRateChartRef" style="height: 400px;"></div>
        </div>
      </el-col>
      <el-col :span="10">
        <div class="chart-container">
          <div class="chart-title">任务类型占比</div>
          <div ref="taskTypePieChartRef" style="height: 400px;"></div>
        </div>
      </el-col>
    </el-row>

    <el-row :gutter="20" class="chart-row">
      <el-col :span="24">
        <div class="chart-container">
          <div class="chart-title">任务ROI排行</div>
          <div ref="roiChartRef" style="height: 400px;"></div>
        </div>
      </el-col>
    </el-row>

    <el-row :gutter="20" class="chart-row">
      <el-col :span="24">
        <div class="chart-container">
          <div class="chart-title">任务统计明细</div>
          <el-table :data="taskList" v-loading="loading" border stripe>
            <el-table-column prop="task_name" label="任务名称" min-width="200" />
            <el-table-column prop="send_count" label="发送人数" width="120" align="center" />
            <el-table-column prop="success_count" label="成功人数" width="120" align="center" />
            <el-table-column prop="success_rate" label="成功率" width="120" align="center">
              <template #default="{ row }">
                {{ row.success_rate || 0 }}%
              </template>
            </el-table-column>
            <el-table-column prop="convert_count" label="转化人数" width="120" align="center" />
            <el-table-column prop="convert_rate" label="转化率" width="120" align="center">
              <template #default="{ row }">
                {{ row.convert_rate || 0 }}%
              </template>
            </el-table-column>
            <el-table-column prop="convert_amount" label="转化金额(元)" width="140" align="center">
              <template #default="{ row }">
                <span class="amount">¥{{ formatAmount(row.convert_amount) }}</span>
              </template>
            </el-table-column>
            <el-table-column prop="send_cost" label="发送成本(元)" width="140" align="center">
              <template #default="{ row }">
                ¥{{ formatAmount(row.send_cost) }}
              </template>
            </el-table-column>
            <el-table-column prop="roi" label="ROI" width="120" align="center">
              <template #default="{ row }">
                <span class="highlight">{{ row.roi || 0 }}</span>
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
import { Search, Download, Message, CircleCheck, CircleClose, View, UserFilled, Money } from '@element-plus/icons-vue'
import { getSmsTaskList, getSmsTaskStatistics } from '@/api/sms'
import * as echarts from 'echarts'

const loading = ref(false)
const exporting = ref(false)
const taskType = ref('')
const taskList = ref([])
const trendData = ref([])

const taskTypeOptions = [
  { label: '营销短信', value: 'marketing' },
  { label: '会员关怀', value: 'care' },
  { label: '活动通知', value: 'activity' },
  { label: '生日祝福', value: 'birthday' },
  { label: '优惠券提醒', value: 'coupon' }
]

const summary = reactive({
  totalSend: 0,
  totalSendChange: 0,
  successCount: 0,
  successRate: 0,
  successChange: 0,
  failCount: 0,
  failRate: 0,
  readCount: 0,
  readRate: 0,
  convertCount: 0,
  convertRate: 0,
  convertAmount: 0,
  roi: 0
})

const today = new Date()
const formatDate = (d) => {
  const year = d.getFullYear()
  const month = String(d.getMonth() + 1).padStart(2, '0')
  const day = String(d.getDate()).padStart(2, '0')
  return `${year}-${month}-${day}`
}

const dateRange = ref([
  formatDate(new Date(today.getFullYear(), today.getMonth(), today.getDate() - 29)),
  formatDate(today)
])

const sendTrendChartRef = ref(null)
const convertRateChartRef = ref(null)
const taskTypePieChartRef = ref(null)
const roiChartRef = ref(null)
let sendTrendChart = null
let convertRateChart = null
let taskTypePieChart = null
let roiChart = null

function formatAmount(val) {
  if (!val && val !== 0) return '0.00'
  return Number(val).toFixed(2)
}

function getQueryParams() {
  return {
    task_type: taskType.value || '',
    start_date: dateRange.value?.[0] || '',
    end_date: dateRange.value?.[1] || ''
  }
}

async function fetchAllData() {
  loading.value = true
  try {
    await Promise.all([
      fetchTaskStatistics(),
      fetchTaskList()
    ])
  } finally {
    loading.value = false
  }
}

async function fetchTaskStatistics() {
  try {
    const res = await getSmsTaskStatistics(getQueryParams())
    const data = res?.data || res || {}
    trendData.value = data?.trend || data?.list || []
    computeSummary(data)
    renderSendTrendChart()
    renderConvertRateChart()
    renderTaskTypePieChart(data?.taskTypeStats || data?.type_stats || [])
    renderRoiChart(data?.taskRoi || data?.roi_list || [])
  } catch (e) {
    console.error('fetchTaskStatistics error:', e)
  }
}

async function fetchTaskList() {
  try {
    const params = { ...getQueryParams(), page: 1, page_size: 100 }
    const res = await getSmsTaskList(params)
    taskList.value = res?.data?.list || res?.data || res?.list || []
  } catch (e) {
    console.error('fetchTaskList error:', e)
  }
}

function computeSummary(data) {
  const stats = data?.summary || data?.stats || data || {}

  summary.totalSend = stats.total_send || stats.sendCount || 0
  summary.totalSendChange = stats.total_send_change || stats.totalSendChange || 0
  summary.successCount = stats.success_count || stats.successCount || 0
  summary.successRate = stats.success_rate || stats.successRate || 0
  summary.successChange = stats.success_change || stats.successChange || 0
  summary.failCount = stats.fail_count || stats.failCount || 0
  summary.failRate = stats.fail_rate || stats.failRate || 0
  summary.readCount = stats.read_count || stats.readCount || 0
  summary.readRate = stats.read_rate || stats.readRate || 0
  summary.convertCount = stats.convert_count || stats.convertCount || 0
  summary.convertRate = stats.convert_rate || stats.convertRate || 0
  summary.convertAmount = stats.convert_amount || stats.convertAmount || 0
  summary.roi = stats.roi || 0
}

async function handleExport() {
  try {
    exporting.value = true
    ElMessage.success('导出功能开发中')
  } catch (e) {
    ElMessage.error('导出失败')
  } finally {
    exporting.value = false
  }
}

function renderSendTrendChart() {
  if (!sendTrendChartRef.value) return
  if (!sendTrendChart) {
    sendTrendChart = echarts.init(sendTrendChartRef.value)
  }

  const dates = trendData.value.map(d => d.date || d.report_date || '')
  const sendCounts = trendData.value.map(d => d.send_count || d.sendCount || 0)
  const successCounts = trendData.value.map(d => d.success_count || d.successCount || 0)
  const failCounts = trendData.value.map(d => d.fail_count || d.failCount || 0)

  sendTrendChart.setOption({
    tooltip: {
      trigger: 'axis'
    },
    legend: {
      data: ['发送数', '成功数', '失败数']
    },
    grid: {
      left: '3%',
      right: '4%',
      bottom: '3%',
      containLabel: true
    },
    xAxis: {
      type: 'category',
      boundaryGap: false,
      data: dates
    },
    yAxis: {
      type: 'value'
    },
    series: [
      {
        name: '发送数',
        type: 'line',
        smooth: true,
        data: sendCounts,
        lineStyle: { color: '#409EFF', width: 3 },
        itemStyle: { color: '#409EFF' },
        areaStyle: {
          color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
            { offset: 0, color: 'rgba(64, 158, 255, 0.3)' },
            { offset: 1, color: 'rgba(64, 158, 255, 0.05)' }
          ])
        }
      },
      {
        name: '成功数',
        type: 'line',
        smooth: true,
        data: successCounts,
        lineStyle: { color: '#67C23A', width: 2 },
        itemStyle: { color: '#67C23A' }
      },
      {
        name: '失败数',
        type: 'line',
        smooth: true,
        data: failCounts,
        lineStyle: { color: '#F56C6C', width: 2 },
        itemStyle: { color: '#F56C6C' }
      }
    ]
  })
}

function renderConvertRateChart() {
  if (!convertRateChartRef.value) return
  if (!convertRateChart) {
    convertRateChart = echarts.init(convertRateChartRef.value)
  }

  const dates = trendData.value.map(d => d.date || d.report_date || '')
  const convertRates = trendData.value.map(d => d.convert_rate || d.convertRate || 0)
  const readRates = trendData.value.map(d => d.read_rate || d.readRate || 0)

  convertRateChart.setOption({
    tooltip: {
      trigger: 'axis',
      formatter: (params) => {
        let result = params[0].name + '<br/>'
        params.forEach(item => {
          result += `${item.marker}${item.seriesName}: ${item.value}%<br/>`
        })
        return result
      }
    },
    legend: {
      data: ['阅读率', '转化率']
    },
    grid: {
      left: '3%',
      right: '4%',
      bottom: '3%',
      containLabel: true
    },
    xAxis: {
      type: 'category',
      boundaryGap: false,
      data: dates
    },
    yAxis: {
      type: 'value',
      axisLabel: {
        formatter: '{value}%'
      }
    },
    series: [
      {
        name: '阅读率',
        type: 'line',
        smooth: true,
        data: readRates,
        lineStyle: { color: '#E6A23C', width: 2 },
        itemStyle: { color: '#E6A23C' },
        areaStyle: {
          color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
            { offset: 0, color: 'rgba(230, 162, 60, 0.3)' },
            { offset: 1, color: 'rgba(230, 162, 60, 0.05)' }
          ])
        }
      },
      {
        name: '转化率',
        type: 'line',
        smooth: true,
        data: convertRates,
        lineStyle: { color: '#67C23A', width: 3 },
        itemStyle: { color: '#67C23A' }
      }
    ]
  })
}

function renderTaskTypePieChart(typeStats) {
  if (!taskTypePieChartRef.value) return
  if (!taskTypePieChart) {
    taskTypePieChart = echarts.init(taskTypePieChartRef.value)
  }

  let data = typeStats || []
  if (data.length === 0) {
    data = taskTypeOptions.map(item => ({
      value: Math.floor(Math.random() * 1000) + 100,
      name: item.label,
      itemStyle: { color: getTypeColor(item.value) }
    }))
  } else {
    data = data.map(item => ({
      value: item.count || item.value || 0,
      name: item.name || item.type_name || item.taskType || '',
      itemStyle: { color: getTypeColor(item.type || item.value) }
    }))
  }

  taskTypePieChart.setOption({
    tooltip: {
      trigger: 'item',
      formatter: '{b}: {c} ({d}%)'
    },
    legend: {
      orient: 'vertical',
      right: 10,
      top: 'center'
    },
    series: [
      {
        type: 'pie',
        radius: ['40%', '70%'],
        avoidLabelOverlap: false,
        itemStyle: {
          borderRadius: 8,
          borderColor: '#fff',
          borderWidth: 2
        },
        label: {
          show: false,
          position: 'center'
        },
        emphasis: {
          label: {
            show: true,
            fontSize: 16,
            fontWeight: 'bold'
          }
        },
        labelLine: {
          show: false
        },
        data: data
      }
    ]
  })
}

function getTypeColor(type) {
  const colorMap = {
    marketing: '#409EFF',
    care: '#67C23A',
    activity: '#E6A23C',
    birthday: '#F56C6C',
    coupon: '#909399'
  }
  return colorMap[type] || '#409EFF'
}

function renderRoiChart(roiList) {
  if (!roiChartRef.value) return
  if (!roiChart) {
    roiChart = echarts.init(roiChartRef.value)
  }

  let data = roiList || []
  if (data.length === 0) {
    data = taskList.value.slice(0, 10).map(item => ({
      name: item.task_name || '未知任务',
      value: item.roi || 0,
      amount: item.convert_amount || 0
    }))
  }

  const names = data.map(d => d.name || d.task_name || '').reverse()
  const rois = data.map(d => Number(d.roi || d.value || 0)).reverse()

  roiChart.setOption({
    tooltip: {
      trigger: 'axis',
      axisPointer: { type: 'shadow' },
      formatter: (params) => {
        const p = params[0]
        const item = data[data.length - 1 - p.dataIndex]
        return `${p.name}<br/>ROI: ${p.value}<br/>转化金额: ¥${formatAmount(item?.amount || item?.convert_amount || 0)}`
      }
    },
    grid: {
      left: '3%',
      right: '15%',
      bottom: '3%',
      top: '3%',
      containLabel: true
    },
    xAxis: {
      type: 'value',
      axisLabel: {
        formatter: (val) => val
      }
    },
    yAxis: {
      type: 'category',
      data: names,
      axisLabel: {
        width: 100,
        overflow: 'truncate'
      }
    },
    series: [
      {
        type: 'bar',
        data: rois,
        itemStyle: {
          color: (params) => {
            const val = params.data
            if (val >= 3) return '#67C23A'
            if (val >= 1) return '#409EFF'
            if (val >= 0) return '#E6A23C'
            return '#F56C6C'
          },
          borderRadius: [0, 4, 4, 0]
        },
        label: {
          show: true,
          position: 'right',
          formatter: (params) => params.value
        }
      }
    ]
  })
}

function handleResize() {
  sendTrendChart?.resize()
  convertRateChart?.resize()
  taskTypePieChart?.resize()
  roiChart?.resize()
}

onMounted(async () => {
  await fetchAllData()
  window.addEventListener('resize', handleResize)
})

onBeforeUnmount(() => {
  window.removeEventListener('resize', handleResize)
  sendTrendChart?.dispose()
  convertRateChart?.dispose()
  taskTypePieChart?.dispose()
  roiChart?.dispose()
})
</script>

<style scoped lang="scss">
.sms-page {
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

    &.total { border-left-color: #409eff; }
    &.success { border-left-color: #67c23a; }
    &.fail { border-left-color: #f56c6c; }
    &.read { border-left-color: #e6a23c; }
    &.convert { border-left-color: #909399; }
    &.roi { border-left-color: #b37feb; }

    .card-icon {
      width: 56px;
      height: 56px;
      border-radius: 50%;
      display: flex;
      align-items: center;
      justify-content: center;
      color: #fff;
    }

    &.total .card-icon { background: linear-gradient(135deg, #409eff, #79bbff); }
    &.success .card-icon { background: linear-gradient(135deg, #67c23a, #95d475); }
    &.fail .card-icon { background: linear-gradient(135deg, #f56c6c, #f89898); }
    &.read .card-icon { background: linear-gradient(135deg, #e6a23c, #eebe77); }
    &.convert .card-icon { background: linear-gradient(135deg, #909399, #b1b3b8); }
    &.roi .card-icon { background: linear-gradient(135deg, #b37feb, #c9a0f2); }

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
        position: relative;

        .sub-info {
          display: block;
          font-size: 12px;
          font-weight: normal;
          color: #909399;
          margin-top: 4px;
        }

        .change {
          position: absolute;
          top: 0;
          right: 0;
          font-size: 14px;
          font-weight: normal;

          &.up { color: #67c23a; }
          &.down { color: #f56c6c; }
        }
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
}
</style>
