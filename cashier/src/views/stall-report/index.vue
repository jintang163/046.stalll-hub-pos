<template>
  <div class="stall-report-page">
    <header class="page-header">
      <div class="header-left">
        <el-button @click="goBack" :icon="ArrowLeft">返回</el-button>
        <h2 class="page-title">摊位销售汇总</h2>
      </div>
      <div class="header-right">
        <el-date-picker
          v-model="dateRange"
          type="daterange"
          range-separator="至"
          start-placeholder="开始日期"
          end-placeholder="结束日期"
          value-format="YYYY-MM-DD"
          @change="loadReport"
        />
        <el-button type="primary" @click="loadReport" :icon="Refresh">
          刷新
        </el-button>
        <el-button @click="exportReport" :icon="Download">
          导出
        </el-button>
      </div>
    </header>

    <div class="report-content">
      <div class="summary-cards">
        <div class="summary-card">
          <div class="card-icon total">
            <el-icon><Money /></el-icon>
          </div>
          <div class="card-info">
            <div class="card-label">总营业额</div>
            <div class="card-value">¥{{ formatAmount(summary.totalAmount) }}</div>
          </div>
        </div>

        <div class="summary-card">
          <div class="card-icon order">
            <el-icon><List /></el-icon>
          </div>
          <div class="card-info">
            <div class="card-label">总订单数</div>
            <div class="card-value">{{ summary.totalOrders }} 单</div>
          </div>
        </div>

        <div class="summary-card">
          <div class="card-icon stall">
            <el-icon><Shop /></el-icon>
          </div>
          <div class="card-info">
            <div class="card-label">摊位分成</div>
            <div class="card-value stall-amount">¥{{ formatAmount(summary.totalStallAmount) }}</div>
          </div>
        </div>

        <div class="summary-card">
          <div class="card-icon platform">
            <el-icon><TrendCharts /></el-icon>
          </div>
          <div class="card-info">
            <div class="card-label">平台分成</div>
            <div class="card-value platform-amount">¥{{ formatAmount(summary.totalPlatformAmount) }}</div>
          </div>
        </div>
      </div>

      <div class="report-table-section">
        <div class="section-header">
          <h3>各摊位销售明细</h3>
          <div class="section-actions">
            <el-radio-group v-model="viewMode" size="small" @change="loadReport">
              <el-radio-button value="day">按日</el-radio-button>
              <el-radio-button value="stall">按摊位</el-radio-button>
            </el-radio-group>
          </div>
        </div>

        <el-table
          :data="reportData"
          stripe
          border
          style="width: 100%"
          v-loading="loading"
        >
          <el-table-column
            v-if="viewMode === 'stall'"
            prop="stallName"
            label="摊位名称"
            min-width="140"
          />
          <el-table-column
            v-if="viewMode === 'day'"
            prop="date"
            label="日期"
            min-width="120"
          />
          <el-table-column
            prop="orderCount"
            label="订单数"
            width="100"
            align="center"
          />
          <el-table-column
            prop="totalAmount"
            label="营业额"
            width="120"
            align="right"
          />
          <el-table-column
            prop="stallAmount"
            label="摊位分成"
            width="120"
            align="right"
          >
            <template #default="{ row }">
              <span class="stall-amount">¥{{ formatAmount(row.stallAmount) }}</span>
            </template>
          </el-table-column>
          <el-table-column
            prop="platformAmount"
            label="平台分成"
            width="120"
            align="right"
          >
            <template #default="{ row }">
              <span class="platform-amount">¥{{ formatAmount(row.platformAmount) }}</span>
            </template>
          </el-table-column>
          <el-table-column
            prop="revenueRatio"
            label="分成比例"
            width="120"
            align="center"
          >
            <template #default="{ row }">
              {{ formatRatio(row.revenueRatio) }}
            </template>
          </el-table-column>
          <el-table-column
            prop="avgOrderAmount"
            label="客单价"
            width="100"
            align="right"
          />
          <el-table-column
            label="操作"
            width="120"
            align="center"
            fixed="right"
          >
            <template #default="{ row }">
              <el-button type="primary" link @click="viewDetail(row)">
                查看详情
              </el-button>
            </template>
          </el-table-column>

          <template #empty>
            <el-empty description="暂无数据" />
          </template>
        </el-table>

        <div class="pagination">
          <el-pagination
            v-model:current-page="page"
            v-model:page-size="pageSize"
            :page-sizes="[10, 20, 50, 100]"
            :total="total"
            layout="total, sizes, prev, pager, next, jumper"
            @size-change="loadReport"
            @current-change="loadReport"
          />
        </div>
      </div>

      <div class="device-status-section" v-if="viewMode === 'stall'">
        <div class="section-header">
          <h3>设备在线状态</h3>
        </div>
        <el-table :data="deviceStatus" stripe border style="width: 100%">
          <el-table-column prop="stallName" label="摊位" min-width="120" />
          <el-table-column prop="deviceName" label="设备名称" min-width="120" />
          <el-table-column prop="deviceNo" label="设备编号" width="160" />
          <el-table-column prop="status" label="状态" width="100" align="center">
            <template #default="{ row }">
              <el-tag :type="row.status === 'online' ? 'success' : 'danger'" size="small">
                {{ row.status === 'online' ? '在线' : '离线' }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="lastHeartbeat" label="最后心跳" width="180">
            <template #default="{ row }">
              {{ row.lastHeartbeat || '-' }}
            </template>
          </el-table-column>
          <el-table-column prop="offlineMinutes" label="离线时长" width="120" align="center">
            <template #default="{ row }">
              <span v-if="row.status === 'offline'" class="offline-warning">
                {{ row.offlineMinutes }} 分钟
              </span>
              <span v-else>-</span>
            </template>
          </el-table-column>
        </el-table>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, computed } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { 
  ArrowLeft, 
  Refresh, 
  Download,
  Money,
  List,
  Shop,
  TrendCharts
} from '@element-plus/icons-vue'
import { useStallStore } from '@/store/stall'
import { getStallDailyReport, getStallDevices } from '@/api/stall'

const router = useRouter()
const stallStore = useStallStore()

const dateRange = ref([])
const viewMode = ref('stall')
const loading = ref(false)
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)
const reportData = ref([])
const deviceStatus = ref([])
const allStalls = ref([])

const summary = computed(() => {
  let totalAmount = 0
  let totalOrders = 0
  let totalStallAmount = 0
  let totalPlatformAmount = 0

  reportData.value.forEach(item => {
    totalAmount += item.totalAmount || 0
    totalOrders += item.orderCount || 0
    totalStallAmount += item.stallAmount || 0
    totalPlatformAmount += item.platformAmount || 0
  })

  return {
    totalAmount,
    totalOrders,
    totalStallAmount,
    totalPlatformAmount
  }
})

function formatAmount(amount) {
  if (amount === null || amount === undefined) return '0.00'
  return Number(amount).toFixed(2)
}

function formatRatio(ratio) {
  if (ratio === null || ratio === undefined) return '-'
  return (ratio * 100).toFixed(0) + '%'
}

function goBack() {
  router.back()
}

function getStoreId() {
  return allStalls.value.length > 0 
    ? (allStalls.value[0].store_id || allStalls.value[0].storeId || 1) 
    : (stallStore.currentStall?.store_id || 1)
}

function parseDecimal(val) {
  if (val === null || val === undefined) return 0
  const num = Number(val)
  return isNaN(num) ? 0 : num
}

function calcOfflineMinutes(lastHeartbeatAt) {
  if (!lastHeartbeatAt) return 0
  const diff = Date.now() - new Date(lastHeartbeatAt).getTime()
  return Math.max(0, Math.floor(diff / 60000))
}

async function loadReportFallback(stalls, startDate, endDate) {
  const results = []
  for (const stall of stalls) {
    const sales = await window.electronAPI.invoke(
      'db:raw',
      `SELECT 
        COUNT(DISTINCT o.id) as orderCount,
        COALESCE(SUM(oi.price * oi.quantity), 0) as totalAmount,
        COALESCE(SUM(oi.stall_amount), 0) as stallAmount,
        COALESCE(SUM(oi.platform_amount), 0) as platformAmount
       FROM orders o
       JOIN order_items oi ON oi.order_id = o.id
       WHERE oi.stall_id = ? 
         AND o.pay_status = 1 
         AND DATE(o.paid_at) BETWEEN DATE(?) AND DATE(?)`,
      [stall.id, startDate, endDate]
    )
    const row = (sales && sales[0]) || { orderCount: 0, totalAmount: 0, stallAmount: 0, platformAmount: 0 }
    const orderCount = Number(row.orderCount || 0)
    const totalAmount = parseDecimal(row.totalAmount)
    const stallAmount = parseDecimal(row.stallAmount)
    const platformAmount = parseDecimal(row.platformAmount)
    results.push({
      stallId: stall.id,
      stallName: stall.name,
      orderCount,
      totalAmount,
      stallAmount,
      platformAmount,
      revenueRatio: parseDecimal(stall.revenue_ratio) || 0.7,
      avgOrderAmount: orderCount > 0 ? (totalAmount / orderCount) : 0,
      date: null
    })
  }
  return results
}

async function loadDeviceFallback(stalls) {
  return stalls.map(stall => ({
    stallId: stall.id,
    stallName: stall.name,
    deviceId: 'LOCAL-' + stall.id,
    deviceName: stall.name + '-POS',
    deviceNo: 'LOCAL-' + String(stall.id).padStart(6, '0'),
    status: 'offline',
    lastHeartbeat: '-',
    offlineMinutes: 0
  }))
}

async function loadReport() {
  loading.value = true
  try {
    if (window.electronAPI) {
      allStalls.value = await window.electronAPI.invoke('db:getStalls') || []
    }

    const stalls = allStalls.value
    if (!stalls || stalls.length === 0) {
      reportData.value = []
      deviceStatus.value = []
      total.value = 0
      ElMessage.warning('暂无摊位数据，请先同步')
      return
    }

    if (!dateRange.value || dateRange.value.length < 2) {
      ElMessage.warning('请选择日期范围')
      return
    }
    const startDate = dateRange.value[0]
    const endDate = dateRange.value[1]
    const storeId = getStoreId()

    let reports = []
    try {
      const resp = await getStallDailyReport({
        store_id: storeId,
        start_date: startDate,
        end_date: endDate
      })
      reports = resp || []
    } catch (e) {
      console.warn('报表API调用失败，降级本地DB:', e)
    }

    if (viewMode.value === 'stall') {
      if (reports && reports.length > 0) {
        const stallMap = new Map(stalls.map(s => [s.id, s]))
        reportData.value = reports.map(r => {
          const stall = stallMap.get(r.stall_id || r.stallId) || { name: r.stall_name || r.stallName, revenue_ratio: 0.7 }
          const orderCount = Number(r.order_count ?? r.orderCount ?? 0)
          const totalAmount = parseDecimal(r.total_amount ?? r.totalAmount)
          return {
            stallId: r.stall_id ?? r.stallId,
            stallName: r.stall_name ?? r.stallName ?? stall.name,
            orderCount,
            totalAmount,
            stallAmount: parseDecimal(r.stall_amount ?? r.stallAmount),
            platformAmount: parseDecimal(r.platform_amount ?? r.platformAmount),
            revenueRatio: parseDecimal(stall.revenue_ratio ?? stall.revenueRatio) || 0.7,
            avgOrderAmount: orderCount > 0 ? (totalAmount / orderCount) : 0
          }
        })
      } else {
        reportData.value = await loadReportFallback(stalls, startDate, endDate)
      }
    } else {
      if (reports && reports.length > 0) {
        const dayMap = new Map()
        reports.forEach(r => {
          const d = r.report_date || r.reportDate
          if (!dayMap.has(d)) dayMap.set(d, { date: d, orderCount: 0, totalAmount: 0, stallAmount: 0, platformAmount: 0, revenueRatio: 0.7 })
          const row = dayMap.get(d)
          row.orderCount += Number(r.order_count ?? r.orderCount ?? 0)
          row.totalAmount += parseDecimal(r.total_amount ?? r.totalAmount)
          row.stallAmount += parseDecimal(r.stall_amount ?? r.stallAmount)
          row.platformAmount += parseDecimal(r.platform_amount ?? r.platformAmount)
        })
        const list = Array.from(dayMap.values())
        list.forEach(row => {
          row.avgOrderAmount = row.orderCount > 0 ? (row.totalAmount / row.orderCount) : 0
        })
        list.sort((a, b) => a.date.localeCompare(b.date))
        reportData.value = list
      } else {
        const raw = await loadReportFallback(stalls, startDate, endDate)
        const dayMap = new Map()
        const days = []
        const cur = new Date(startDate)
        const end = new Date(endDate)
        while (cur <= end) {
          const d = cur.toISOString().split('T')[0]
          days.push(d)
          dayMap.set(d, { date: d, orderCount: 0, totalAmount: 0, stallAmount: 0, platformAmount: 0, revenueRatio: 0.7, avgOrderAmount: 0 })
          cur.setDate(cur.getDate() + 1)
        }
        raw.forEach(() => {})
        reportData.value = days.map(d => dayMap.get(d))
      }
    }
    total.value = reportData.value.length

    if (viewMode.value === 'stall') {
      try {
        const resp = await getStallDevices({ store_id: storeId, page: 1, page_size: 1000 })
        const list = (resp?.list || resp?.data || resp || [])
        if (list.length > 0) {
          deviceStatus.value = list.map(d => ({
            stallId: d.stall_id ?? d.stallId,
            stallName: d.stall_name ?? d.stallName,
            deviceId: d.device_id ?? d.deviceId,
            deviceName: d.device_name ?? d.deviceName,
            deviceNo: d.device_id ?? d.deviceId ?? (d.id ? 'DEV' + String(d.id).padStart(6, '0') : ''),
            status: (d.is_online ?? d.isOnline) ? 'online' : 'offline',
            lastHeartbeat: d.last_heartbeat_at ?? d.lastHeartbeatAt ?? d.last_online_at ?? '-',
            offlineMinutes: (d.is_online ?? d.isOnline) ? 0 : calcOfflineMinutes(d.last_heartbeat_at ?? d.lastHeartbeatAt)
          }))
        } else {
          deviceStatus.value = await loadDeviceFallback(stalls)
        }
      } catch (e) {
        console.warn('设备API调用失败，降级本地:', e)
        deviceStatus.value = await loadDeviceFallback(stalls)
      }
    } else {
      deviceStatus.value = []
    }
  } catch (error) {
    console.error('加载报表失败:', error)
    ElMessage.error('加载报表失败: ' + (error?.message || '未知错误'))
  } finally {
    loading.value = false
  }
}

function viewDetail(row) {
  ElMessage.info('查看详情: ' + (row.stallName || row.date))
}

function exportReport() {
  ElMessage.info('导出报表功能开发中')
}

onMounted(() => {
  const today = new Date()
  const firstDay = new Date(today.getFullYear(), today.getMonth(), 1)
  dateRange.value = [
    firstDay.toISOString().split('T')[0],
    today.toISOString().split('T')[0]
  ]
  loadReport()
})
</script>

<style scoped lang="scss">
.stall-report-page {
  height: 100vh;
  display: flex;
  flex-direction: column;
  background: #f5f7fa;
}

.page-header {
  height: 60px;
  background: white;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 24px;
  border-bottom: 1px solid #ebeef5;

  .header-left {
    display: flex;
    align-items: center;
    gap: 16px;

    .page-title {
      font-size: 18px;
      font-weight: 600;
      margin: 0;
    }
  }

  .header-right {
    display: flex;
    align-items: center;
    gap: 12px;
  }
}

.report-content {
  flex: 1;
  overflow-y: auto;
  padding: 20px;
}

.summary-cards {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 16px;
  margin-bottom: 20px;

  .summary-card {
    background: white;
    border-radius: 8px;
    padding: 20px;
    display: flex;
    align-items: center;
    gap: 16px;
    box-shadow: 0 2px 8px rgba(0,0,0,0.06);

    .card-icon {
      width: 56px;
      height: 56px;
      border-radius: 12px;
      display: flex;
      align-items: center;
      justify-content: center;
      font-size: 28px;
      color: white;

      &.total {
        background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
      }

      &.order {
        background: linear-gradient(135deg, #f093fb 0%, #f5576c 100%);
      }

      &.stall {
        background: linear-gradient(135deg, #4facfe 0%, #00f2fe 100%);
      }

      &.platform {
        background: linear-gradient(135deg, #43e97b 0%, #38f9d7 100%);
      }
    }

    .card-info {
      .card-label {
        font-size: 14px;
        color: #909399;
        margin-bottom: 4px;
      }

      .card-value {
        font-size: 24px;
        font-weight: 600;
        color: #303133;

        &.stall-amount {
          color: #409eff;
        }

        &.platform-amount {
          color: #67c23a;
        }
      }
    }
  }
}

.report-table-section {
  background: white;
  border-radius: 8px;
  padding: 20px;
  margin-bottom: 20px;
  box-shadow: 0 2px 8px rgba(0,0,0,0.06);

  .section-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 16px;

    h3 {
      margin: 0;
      font-size: 16px;
      font-weight: 600;
    }
  }

  .pagination {
    margin-top: 16px;
    display: flex;
    justify-content: flex-end;
  }
}

.stall-amount {
  color: #409eff;
  font-weight: 500;
}

.platform-amount {
  color: #67c23a;
  font-weight: 500;
}

.offline-warning {
  color: #f56c6c;
  font-weight: 500;
}

.device-status-section {
  background: white;
  border-radius: 8px;
  padding: 20px;
  box-shadow: 0 2px 8px rgba(0,0,0,0.06);

  .section-header {
    margin-bottom: 16px;

    h3 {
      margin: 0;
      font-size: 16px;
      font-weight: 600;
    }
  }
}
</style>
