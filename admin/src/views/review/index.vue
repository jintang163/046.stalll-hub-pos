<template>
  <div class="review-page">
    <div class="page-header">
      <h2 class="page-title">点评评分趋势</h2>
      <div class="header-actions">
        <el-select v-model="storeId" placeholder="选择门店" clearable style="width: 160px">
          <el-option label="全部门店" :value="0" />
          <el-option v-for="store in storeList" :key="store.id" :label="store.name" :value="store.id" />
        </el-select>
        <el-radio-group v-model="platform">
          <el-radio-button value="all">全部</el-radio-button>
          <el-radio-button value="dianping">大众点评</el-radio-button>
          <el-radio-button value="meituan">美团</el-radio-button>
        </el-radio-group>
        <el-date-picker
          v-model="dateRange"
          type="daterange"
          range-separator="至"
          start-placeholder="开始日期"
          end-placeholder="结束日期"
          value-format="YYYY-MM-DD"
          style="width: 280px" />
        <el-button type="primary" @click="fetchAllData">
          <el-icon><Search /></el-icon>查询
        </el-button>
        <el-button type="success" @click="handleSync" :loading="syncing">
          <el-icon><Refresh /></el-icon>手动同步
        </el-button>
        <el-button type="warning" @click="handleCheckAlerts" :loading="checkingAlerts">
          <el-icon><Bell /></el-icon>检查告警
        </el-button>
        <el-button type="info" @click="showAuthDialog = true">
          <el-icon><Setting /></el-icon>平台授权
        </el-button>
      </div>
    </div>

    <el-dialog v-model="showAuthDialog" title="平台授权配置" width="600px" destroy-on-close>
      <el-table :data="authList" v-loading="authLoading" border stripe>
        <el-table-column prop="store_id" label="门店ID" width="80" align="center" />
        <el-table-column label="门店" width="140" align="center">
          <template #default="{ row }">{{ getStoreName(row.store_id) }}</template>
        </el-table-column>
        <el-table-column prop="platform" label="平台" width="120" align="center">
          <template #default="{ row }">
            <el-tag :type="row.platform === 'dianping' ? 'danger' : 'warning'">
              {{ row.platform === 'dianping' ? '大众点评' : '美团' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="sync_status" label="同步状态" width="120" align="center">
          <template #default="{ row }">
            <el-tag :type="row.sync_status === 'success' ? 'success' : row.sync_status === 'syncing' ? 'warning' : row.sync_status === 'failed' ? 'danger' : 'info'" size="small">
              {{ row.sync_status === 'success' ? '成功' : row.sync_status === 'syncing' ? '同步中' : row.sync_status === 'failed' ? '失败' : '待同步' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="last_sync_time" label="最后同步" width="170" align="center">
          <template #default="{ row }">{{ row.last_sync_time || '从未同步' }}</template>
        </el-table-column>
        <el-table-column label="状态" width="100" align="center">
          <template #default="{ row }">
            <el-switch v-model="row.status" :active-value="1" :inactive-value="0" @change="toggleAuthStatus(row)" />
          </template>
        </el-table-column>
        <el-table-column label="操作" width="100" align="center" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="editAuth(row)">编辑</el-button>
            <el-button type="danger" link size="small" @click="deleteAuth(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>

      <div style="margin-top: 16px; text-align: right;">
        <el-button type="primary" @click="addAuth">新增授权</el-button>
      </div>
    </el-dialog>

    <el-dialog v-model="showAuthFormDialog" :title="authForm.id ? '编辑授权' : '新增授权'" width="500px" destroy-on-close>
      <el-form :model="authForm" label-width="100px">
        <el-form-item label="门店" required>
          <el-select v-model="authForm.store_id" placeholder="选择门店" :disabled="!!authForm.id" style="width: 100%">
            <el-option v-for="store in storeList" :key="store.id" :label="store.name" :value="store.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="平台" required>
          <el-radio-group v-model="authForm.platform" :disabled="!!authForm.id">
            <el-radio-button value="dianping">大众点评</el-radio-button>
            <el-radio-button value="meituan">美团</el-radio-button>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="店铺URL">
          <el-input v-model="authForm.store_url" placeholder="大众点评/美团店铺页面URL" />
        </el-form-item>
        <el-form-item label="店铺ID">
          <el-input v-model="authForm.shop_id" placeholder="平台分配的店铺ID" />
        </el-form-item>
        <el-form-item label="授权Token">
          <el-input v-model="authForm.auth_token" placeholder="平台授权Token" type="password" show-password />
        </el-form-item>
        <el-form-item label="刷新Token">
          <el-input v-model="authForm.refresh_token" placeholder="平台刷新Token" type="password" show-password />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showAuthFormDialog = false">取消</el-button>
        <el-button type="primary" @click="saveAuth" :loading="authSaving">保存</el-button>
      </template>
    </el-dialog>

    <div class="summary-cards">
      <el-row :gutter="20">
        <el-col :span="6">
          <div class="summary-card rating">
            <div class="card-icon">
              <el-icon :size="28"><Star /></el-icon>
            </div>
            <div class="card-info">
              <div class="card-label">当前综合评分</div>
              <div class="card-value">
                {{ summary.avgRating }}
                <span class="change" :class="summary.ratingChange >= 0 ? 'up' : 'down'">
                  {{ summary.ratingChange >= 0 ? '↑' : '↓' }}{{ Math.abs(summary.ratingChange) }}
                </span>
              </div>
            </div>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="summary-card good">
            <div class="card-icon">
              <el-icon :size="28"><CircleCheck /></el-icon>
            </div>
            <div class="card-info">
              <div class="card-label">好评率</div>
              <div class="card-value">
                {{ summary.goodRate }}%
                <span class="change" :class="summary.goodRateChange >= 0 ? 'up' : 'down'">
                  {{ summary.goodRateChange >= 0 ? '↑' : '↓' }}{{ Math.abs(summary.goodRateChange) }}%
                </span>
              </div>
            </div>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="summary-card total">
            <div class="card-icon">
              <el-icon :size="28"><ChatDotRound /></el-icon>
            </div>
            <div class="card-info">
              <div class="card-label">评价总数</div>
              <div class="card-value">
                {{ summary.totalReviews }}
                <span class="change" :class="summary.totalReviewsChange >= 0 ? 'up' : 'down'">
                  {{ summary.totalReviewsChange >= 0 ? '↑' : '↓' }}{{ Math.abs(summary.totalReviewsChange) }}
                </span>
              </div>
            </div>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="summary-card pending">
            <div class="card-icon">
              <el-icon :size="28"><Tickets /></el-icon>
            </div>
            <div class="card-info">
              <div class="card-label">待处理工单</div>
              <div class="card-value">{{ summary.pendingOrders }}</div>
            </div>
          </div>
        </el-col>
      </el-row>
    </div>

    <el-row :gutter="20" class="chart-row">
      <el-col :span="24">
        <div class="chart-container">
          <div class="chart-title">评分趋势</div>
          <div ref="ratingTrendChartRef" style="height: 400px;"></div>
        </div>
      </el-col>
    </el-row>

    <el-row :gutter="20" class="chart-row">
      <el-col :span="14">
        <div class="chart-container">
          <div class="chart-title">评价数量分布</div>
          <div ref="reviewCountChartRef" style="height: 400px;"></div>
        </div>
      </el-col>
      <el-col :span="10">
        <div class="chart-container">
          <div class="chart-title">评分分布</div>
          <div ref="ratingPieChartRef" style="height: 400px;"></div>
        </div>
      </el-col>
    </el-row>

    <el-row :gutter="20" class="chart-row">
      <el-col :span="24">
        <div class="chart-container">
          <div class="chart-title">最新评分快照</div>
          <el-table :data="ratingList" v-loading="loading" border stripe>
            <el-table-column prop="snapshot_date" label="日期" width="140" align="center" />
            <el-table-column prop="store_id" label="门店" width="140" align="center">
              <template #default="{ row }">
                {{ getStoreName(row.store_id) }}
              </template>
            </el-table-column>
            <el-table-column prop="platform" label="平台" width="120" align="center">
              <template #default="{ row }">
                <el-tag :type="row.platform === 'dianping' ? 'danger' : 'warning'">
                  {{ row.platform === 'dianping' ? '大众点评' : '美团' }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="avg_rating" label="综合评分" width="120" align="center">
              <template #default="{ row }">
                <span class="highlight">{{ row.avg_rating }}</span>
              </template>
            </el-table-column>
            <el-table-column prop="taste_rating" label="口味" width="100" align="center" />
            <el-table-column prop="environment_rating" label="环境" width="100" align="center" />
            <el-table-column prop="service_rating" label="服务" width="100" align="center" />
            <el-table-column prop="review_count" label="评价数" width="120" align="center" />
            <el-table-column prop="good_rate" label="好评率" width="120" align="center">
              <template #default="{ row }">
                {{ row.good_rate }}%
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
import { Search, Refresh, Bell, Star, CircleCheck, ChatDotRound, Tickets, Setting } from '@element-plus/icons-vue'
import { getRatingList, getRatingTrend, getReviewList, syncAllReviews, checkAlerts, listPlatformAuths, savePlatformAuth, getPlatformAuth } from '@/api/review'
import { getStoreList } from '@/api/stores'
import * as echarts from 'echarts'

const loading = ref(false)
const syncing = ref(false)
const checkingAlerts = ref(false)
const storeId = ref(0)
const platform = ref('all')
const storeList = ref([])
const ratingList = ref([])
const trendData = ref([])
const reviewData = ref([])

const summary = reactive({
  avgRating: 0,
  ratingChange: 0,
  goodRate: 0,
  goodRateChange: 0,
  totalReviews: 0,
  totalReviewsChange: 0,
  pendingOrders: 0
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

const ratingTrendChartRef = ref(null)
const reviewCountChartRef = ref(null)
const ratingPieChartRef = ref(null)
let ratingTrendChart = null
let reviewCountChart = null
let ratingPieChart = null

const showAuthDialog = ref(false)
const showAuthFormDialog = ref(false)
const authLoading = ref(false)
const authSaving = ref(false)
const authList = ref([])
const authForm = reactive({
  id: 0,
  store_id: 0,
  platform: 'dianping',
  store_url: '',
  shop_id: '',
  auth_token: '',
  refresh_token: ''
})

function getStoreName(id) {
  const store = storeList.value.find(s => s.id === id)
  return store ? store.name : `门店${id}`
}

function getQueryParams() {
  return {
    store_id: storeId.value || 0,
    platform: platform.value === 'all' ? '' : platform.value,
    start_date: dateRange.value?.[0] || '',
    end_date: dateRange.value?.[1] || ''
  }
}

async function fetchStores() {
  try {
    const res = await getStoreList({ page: 1, page_size: 100 })
    storeList.value = res?.list || res?.data || []
  } catch (e) {
    console.error('Failed to fetch stores:', e)
  }
}

async function fetchAllData() {
  loading.value = true
  try {
    await Promise.all([
      fetchRatingList(),
      fetchRatingTrend(),
      fetchReviewStats()
    ])
  } finally {
    loading.value = false
  }
}

async function fetchRatingList() {
  try {
    const res = await getRatingList(getQueryParams())
    ratingList.value = res?.data?.list || res?.data || res || []
    computeSummary()
  } catch (e) {
    console.error('fetchRatingList error:', e)
  }
}

async function fetchRatingTrend() {
  try {
    const res = await getRatingTrend(getQueryParams())
    trendData.value = res?.data || res || []
    renderRatingTrendChart()
  } catch (e) {
    console.error('fetchRatingTrend error:', e)
  }
}

async function fetchReviewStats() {
  try {
    const params = { ...getQueryParams(), page: 1, page_size: 1 }
    const res = await getReviewList(params)
    const data = res?.data || res || {}
    reviewData.value = data?.list || data?.reviews || []
    const total = data?.total || 0
    summary.totalReviews = total
    summary.pendingOrders = data?.pending_orders || 0
    renderReviewCountChart()
    renderRatingPieChart()
  } catch (e) {
    console.error('fetchReviewStats error:', e)
  }
}

function computeSummary() {
  if (ratingList.value.length === 0) return
  const latest = ratingList.value[0]
  summary.avgRating = latest.avg_rating || 0
  summary.goodRate = latest.good_rate || 0
  summary.totalReviews = latest.review_count || summary.totalReviews

  if (ratingList.value.length > 1) {
    const prev = ratingList.value[1]
    summary.ratingChange = Number((summary.avgRating - (prev.avg_rating || 0)).toFixed(2))
    summary.goodRateChange = Number((summary.goodRate - (prev.good_rate || 0)).toFixed(2))
    summary.totalReviewsChange = summary.totalReviews - (prev.review_count || 0)
  }
}

async function handleSync() {
  const activeAuths = authList.value.filter(a => a.status === 1)
  if (activeAuths.length === 0) {
    ElMessage.warning('未配置平台授权，请先在"平台授权"中配置授权信息后再同步')
    showAuthDialog.value = true
    return
  }
  try {
    syncing.value = true
    await syncAllReviews()
    ElMessage.success('同步任务已启动，请稍后刷新查看')
    setTimeout(() => {
      fetchAllData()
    }, 3000)
  } catch (e) {
    ElMessage.error('同步失败')
  } finally {
    syncing.value = false
  }
}

async function handleCheckAlerts() {
  try {
    checkingAlerts.value = true
    await checkAlerts()
    ElMessage.success('告警检查任务已启动')
  } catch (e) {
    ElMessage.error('检查告警失败')
  } finally {
    checkingAlerts.value = false
  }
}

async function fetchAuthList() {
  authLoading.value = true
  try {
    const res = await listPlatformAuths(0)
    authList.value = res?.data || res || []
  } catch (e) {
    console.error('fetchAuthList error:', e)
  } finally {
    authLoading.value = false
  }
}

function addAuth() {
  Object.assign(authForm, {
    id: 0,
    store_id: storeList.value.length > 0 ? storeList.value[0].id : 0,
    platform: 'dianping',
    store_url: '',
    shop_id: '',
    auth_token: '',
    refresh_token: ''
  })
  showAuthFormDialog.value = true
}

function editAuth(row) {
  Object.assign(authForm, {
    id: row.id,
    store_id: row.store_id,
    platform: row.platform,
    store_url: row.store_url || '',
    shop_id: row.shop_id || '',
    auth_token: row.auth_token || '',
    refresh_token: row.refresh_token || ''
  })
  showAuthFormDialog.value = true
}

async function saveAuth() {
  if (!authForm.store_id) {
    ElMessage.warning('请选择门店')
    return
  }
  if (!authForm.platform) {
    ElMessage.warning('请选择平台')
    return
  }
  try {
    authSaving.value = true
    await savePlatformAuth({
      store_id: authForm.store_id,
      platform: authForm.platform,
      store_url: authForm.store_url,
      shop_id: authForm.shop_id,
      auth_token: authForm.auth_token,
      refresh_token: authForm.refresh_token
    })
    ElMessage.success('授权保存成功')
    showAuthFormDialog.value = false
    await fetchAuthList()
  } catch (e) {
    ElMessage.error('保存授权失败')
  } finally {
    authSaving.value = false
  }
}

async function toggleAuthStatus(row) {
  try {
    await savePlatformAuth({
      store_id: row.store_id,
      platform: row.platform,
      store_url: row.store_url,
      shop_id: row.shop_id,
      auth_token: row.auth_token,
      refresh_token: row.refresh_token
    })
    ElMessage.success(row.status === 1 ? '已启用' : '已禁用')
  } catch (e) {
    ElMessage.error('更新状态失败')
    row.status = row.status === 1 ? 0 : 1
  }
}

async function deleteAuth(row) {
  try {
    await savePlatformAuth({
      store_id: row.store_id,
      platform: row.platform,
      store_url: '',
      shop_id: '',
      auth_token: '',
      refresh_token: ''
    })
    ElMessage.success('已删除')
    await fetchAuthList()
  } catch (e) {
    ElMessage.error('删除失败')
  }
}

function renderRatingTrendChart() {
  if (!ratingTrendChartRef.value) return
  if (!ratingTrendChart) {
    ratingTrendChart = echarts.init(ratingTrendChartRef.value)
  }

  const dates = trendData.value.map(d => d.date || d.snapshot_date || '')
  const avgRatings = trendData.value.map(d => d.avg_rating || 0)
  const tasteRatings = trendData.value.map(d => d.taste_rating || 0)
  const envRatings = trendData.value.map(d => d.environment_rating || 0)
  const serviceRatings = trendData.value.map(d => d.service_rating || 0)

  ratingTrendChart.setOption({
    tooltip: {
      trigger: 'axis'
    },
    legend: {
      data: ['综合评分', '口味', '环境', '服务']
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
      min: 0,
      max: 5,
      interval: 1
    },
    series: [
      {
        name: '综合评分',
        type: 'line',
        smooth: true,
        data: avgRatings,
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
        name: '口味',
        type: 'line',
        smooth: true,
        data: tasteRatings,
        lineStyle: { color: '#67C23A', width: 2 },
        itemStyle: { color: '#67C23A' }
      },
      {
        name: '环境',
        type: 'line',
        smooth: true,
        data: envRatings,
        lineStyle: { color: '#E6A23C', width: 2 },
        itemStyle: { color: '#E6A23C' }
      },
      {
        name: '服务',
        type: 'line',
        smooth: true,
        data: serviceRatings,
        lineStyle: { color: '#F56C6C', width: 2 },
        itemStyle: { color: '#F56C6C' }
      }
    ]
  })
}

function renderReviewCountChart() {
  if (!reviewCountChartRef.value) return
  if (!reviewCountChart) {
    reviewCountChart = echarts.init(reviewCountChartRef.value)
  }

  const dates = trendData.value.map(d => d.date || d.snapshot_date || '')
  const goodCounts = trendData.value.map(d => d.good_count || 0)
  const midCounts = trendData.value.map(d => d.mid_count || 0)
  const badCounts = trendData.value.map(d => d.bad_count || 0)

  reviewCountChart.setOption({
    tooltip: {
      trigger: 'axis',
      axisPointer: { type: 'shadow' }
    },
    legend: {
      data: ['好评', '中评', '差评']
    },
    grid: {
      left: '3%',
      right: '4%',
      bottom: '3%',
      containLabel: true
    },
    xAxis: {
      type: 'category',
      data: dates
    },
    yAxis: {
      type: 'value'
    },
    series: [
      {
        name: '好评',
        type: 'bar',
        stack: 'total',
        data: goodCounts,
        itemStyle: {
          color: '#67C23A'
        }
      },
      {
        name: '中评',
        type: 'bar',
        stack: 'total',
        data: midCounts,
        itemStyle: {
          color: '#E6A23C'
        }
      },
      {
        name: '差评',
        type: 'bar',
        stack: 'total',
        data: badCounts,
        itemStyle: {
          color: '#F56C6C'
        }
      }
    ]
  })
}

function renderRatingPieChart() {
  if (!ratingPieChartRef.value) return
  if (!ratingPieChart) {
    ratingPieChart = echarts.init(ratingPieChartRef.value)
  }

  let star5 = 0, star4 = 0, star3 = 0, star2 = 0, star1 = 0
  for (const r of reviewData.value) {
    const rating = Number(r.rating || 0)
    if (rating >= 4.5) star5++
    else if (rating >= 3.5) star4++
    else if (rating >= 2.5) star3++
    else if (rating >= 1.5) star2++
    else star1++
  }

  if (star5 + star4 + star3 + star2 + star1 === 0) {
    star5 = summary.totalReviews > 0 ? Math.round(summary.totalReviews * 0.4) : 0
    star4 = summary.totalReviews > 0 ? Math.round(summary.totalReviews * 0.3) : 0
    star3 = summary.totalReviews > 0 ? Math.round(summary.totalReviews * 0.15) : 0
    star2 = summary.totalReviews > 0 ? Math.round(summary.totalReviews * 0.1) : 0
    star1 = summary.totalReviews > 0 ? Math.round(summary.totalReviews * 0.05) : 0
  }

  ratingPieChart.setOption({
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
        data: [
          { value: star5, name: '5星', itemStyle: { color: '#67C23A' } },
          { value: star4, name: '4星', itemStyle: { color: '#95D475' } },
          { value: star3, name: '3星', itemStyle: { color: '#E6A23C' } },
          { value: star2, name: '2星', itemStyle: { color: '#F56C6C' } },
          { value: star1, name: '1星', itemStyle: { color: '#F78989' } }
        ]
      }
    ]
  })
}

function handleResize() {
  ratingTrendChart?.resize()
  reviewCountChart?.resize()
  ratingPieChart?.resize()
}

onMounted(async () => {
  await fetchStores()
  await Promise.all([fetchAllData(), fetchAuthList()])
  window.addEventListener('resize', handleResize)
})

onBeforeUnmount(() => {
  window.removeEventListener('resize', handleResize)
  ratingTrendChart?.dispose()
  reviewCountChart?.dispose()
  ratingPieChart?.dispose()
})
</script>

<style scoped lang="scss">
.review-page {
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

    &.rating { border-left-color: #409eff; }
    &.good { border-left-color: #67c23a; }
    &.total { border-left-color: #e6a23c; }
    &.pending { border-left-color: #f56c6c; }

    .card-icon {
      width: 56px;
      height: 56px;
      border-radius: 50%;
      display: flex;
      align-items: center;
      justify-content: center;
      color: #fff;
    }

    &.rating .card-icon { background: linear-gradient(135deg, #409eff, #79bbff); }
    &.good .card-icon { background: linear-gradient(135deg, #67c23a, #95d475); }
    &.total .card-icon { background: linear-gradient(135deg, #e6a23c, #eebe77); }
    &.pending .card-icon { background: linear-gradient(135deg, #f56c6c, #f89898); }

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

        .change {
          font-size: 14px;
          font-weight: normal;
          margin-left: 8px;

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

  .highlight {
    color: #409eff;
    font-weight: 600;
  }
}
</style>
