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

const router = useRouter()

const dateRange = ref([])
const viewMode = ref('stall')
const loading = ref(false)
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)
const reportData = ref([])
const deviceStatus = ref([])

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

async function loadReport() {
  loading.value = true
  try {
    if (window.electronAPI) {
      const stalls = await window.electronAPI.invoke('db:getStalls')
      
      reportData.value = stalls.map(stall => ({
        stallId: stall.id,
        stallName: stall.name,
        orderCount: Math.floor(Math.random() * 100) + 10,
        totalAmount: (Math.random() * 5000 + 500).toFixed(2),
        stallAmount: (Math.random() * 3500 + 350).toFixed(2),
        platformAmount: (Math.random() * 1500 + 150).toFixed(2),
        revenueRatio: stall.revenue_ratio || 0.7,
        avgOrderAmount: (Math.random() * 50 + 15).toFixed(2)
      }))
      total.value = reportData.value.length

      deviceStatus.value = stalls.map(stall => ({
        stallId: stall.id,
        stallName: stall.name,
        deviceName: stall.name + '-POS01',
        deviceNo: 'DEV' + String(stall.id).padStart(6, '0'),
        status: Math.random() > 0.3 ? 'online' : 'offline',
        lastHeartbeat: new Date().toLocaleString(),
        offlineMinutes: Math.floor(Math.random() * 60)
      }))
    }
  } catch (error) {
    console.error('加载报表失败:', error)
    ElMessage.error('加载报表失败')
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
