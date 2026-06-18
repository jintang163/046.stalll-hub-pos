<template>
  <div class="sync-page">
    <header class="page-header">
      <h1>数据同步</h1>
      <div class="header-actions">
        <el-button @click="goBack">
          <el-icon><ArrowLeft /></el-icon>
          返回
        </el-button>
      </div>
    </header>

    <div class="page-body">
      <el-row :gutter="20">
        <el-col :span="12">
          <div class="card p-20">
            <h2 class="card-title">同步状态</h2>
            <el-descriptions :column="1" border>
              <el-descriptions-item label="网络状态">
                <el-tag :type="orderStore.isOnline ? 'success' : 'danger'">
                  {{ orderStore.isOnline ? '已连接' : '已断开' }}
                </el-tag>
              </el-descriptions-item>
              <el-descriptions-item label="NSQ连接">
                <el-tag :type="nsqStore.connected ? 'success' : 'warning'">
                  {{ nsqStore.connected ? '已连接' : '未连接' }}
                </el-tag>
              </el-descriptions-item>
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
              <el-descriptions-item label="待同步订单">
                <span class="text-warning">{{ orderStore.pendingOrders.length }}</span> 单
              </el-descriptions-item>
            </el-descriptions>

            <div class="sync-actions">
              <el-button 
                type="primary" 
                size="large"
                :loading="syncStore.isSyncing"
                @click="handleFullSync"
                style="width: 100%; margin-bottom: 12px;"
              >
                <el-icon><RefreshRight /></el-icon>
                全量同步
              </el-button>
              <el-button 
                size="large"
                :loading="syncStore.isSyncing"
                @click="handleIncrementalSync"
                style="width: 100%; margin-bottom: 12px;"
              >
                <el-icon><Top /></el-icon>
                增量同步
              </el-button>
              <el-button 
                size="large"
                type="success"
                :disabled="orderStore.pendingOrders.length === 0"
                @click="handleSyncOrders"
                style="width: 100%;"
              >
                <el-icon><Upload /></el-icon>
                上传订单 ({{ orderStore.pendingOrders.length }})
              </el-button>
            </div>
          </div>
        </el-col>

        <el-col :span="12">
          <div class="card p-20">
            <h2 class="card-title">同步进度</h2>
            
            <div class="progress-display">
              <el-progress 
                :percentage="syncStore.progress.percent" 
                :status="progressStatus"
                :stroke-width="12"
              />
              <div class="progress-message">
                {{ syncStore.progress.message || '等待同步...' }}
              </div>
            </div>

            <div class="sync-log">
              <h3>同步日志</h3>
              <div class="log-list" ref="logListRef">
                <div 
                  v-for="(log, index) in syncLogs" 
                  :key="index"
                  class="log-item"
                  :class="log.type"
                >
                  <span class="log-time">{{ log.time }}</span>
                  <span class="log-text">{{ log.message }}</span>
                </div>
                <div v-if="syncLogs.length === 0" class="empty-log">
                  暂无同步记录
                </div>
              </div>
            </div>
          </div>

          <div class="card p-20 mt-20">
            <h2 class="card-title">待同步订单</h2>
            <div class="pending-list">
              <div 
                v-for="order in orderStore.pendingOrders.slice(0, 10)" 
                :key="order.order_no"
                class="pending-item"
              >
                <div class="order-no">{{ order.order_no }}</div>
                <div class="order-info">
                  <span>{{ order.items.length }}件商品</span>
                  <span class="amount">¥{{ order.actual_amount.toFixed(2) }}</span>
                </div>
                <el-tag size="small" type="warning">待同步</el-tag>
              </div>
              <div v-if="orderStore.pendingOrders.length === 0" class="empty">
                没有待同步的订单
              </div>
              <div v-if="orderStore.pendingOrders.length > 10" class="more">
                还有 {{ orderStore.pendingOrders.length - 10 }} 个订单...
              </div>
            </div>
          </div>
        </el-col>
      </el-row>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, watch, onMounted, nextTick } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { ArrowLeft, RefreshRight, Top, Upload } from '@element-plus/icons-vue'
import { useSyncStore } from '@/store/sync'
import { useProductStore } from '@/store/product'
import { useOrderStore } from '@/store/order'
import { useNSQStore } from '@/store/nsq'
import dayjs from 'dayjs'

const router = useRouter()
const syncStore = useSyncStore()
const productStore = useProductStore()
const orderStore = useOrderStore()
const nsqStore = useNSQStore()

const syncLogs = ref([])
const logListRef = ref(null)

const progressStatus = computed(() => {
  switch (syncStore.progress.status) {
    case 'completed': return 'success'
    case 'error': return 'exception'
    default: return null
  }
})

const addLog = (message, type = 'info') => {
  syncLogs.value.unshift({
    time: dayjs().format('HH:mm:ss'),
    message,
    type
  })
  
  if (syncLogs.value.length > 100) {
    syncLogs.value.pop()
  }
  
  nextTick(() => {
    if (logListRef.value) {
      logListRef.value.scrollTop = 0
    }
  })
}

watch(() => syncStore.progress, (newVal) => {
  if (newVal.message) {
    const type = newVal.status === 'error' ? 'error' : 
                 newVal.status === 'completed' ? 'success' : 'info'
    addLog(newVal.message, type)
  }
}, { deep: true })

const handleFullSync = () => {
  addLog('开始全量同步...', 'info')
  syncStore.fullSync()
}

const handleIncrementalSync = () => {
  addLog('开始增量同步...', 'info')
  syncStore.incrementalSync()
}

const handleSyncOrders = async () => {
  addLog('开始上传订单...', 'info')
  try {
    await orderStore.forceSync()
    addLog(`上传完成，共 ${orderStore.pendingOrders.length} 个待同步订单`, 'success')
  } catch (e) {
    addLog('上传失败: ' + e.message, 'error')
    ElMessage.error('上传失败')
  }
}

const formatDate = (date) => {
  return dayjs(date).format('YYYY-MM-DD HH:mm:ss')
}

const goBack = () => {
  router.push('/')
}

onMounted(() => {
  addLog('页面加载完成', 'info')
})
</script>

<style lang="scss" scoped>
.sync-page {
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
}

.page-body {
  flex: 1;
  padding: 20px 24px;
  overflow-y: auto;
}

.card-title {
  margin: 0 0 20px;
  font-size: 16px;
  font-weight: 600;
  padding-bottom: 12px;
  border-bottom: 1px solid #e4e7ed;
}

.sync-actions {
  margin-top: 24px;
}

.progress-display {
  text-align: center;
  padding: 20px 0;
  
  .progress-message {
    margin-top: 16px;
    color: #606266;
    font-size: 14px;
  }
}

.sync-log {
  margin-top: 24px;
  
  h3 {
    margin: 0 0 12px;
    font-size: 14px;
    font-weight: 600;
  }
  
  .log-list {
    max-height: 300px;
    overflow-y: auto;
    background: #1e1e1e;
    border-radius: 6px;
    padding: 12px;
    font-family: 'Consolas', 'Monaco', monospace;
  }
  
  .log-item {
    display: flex;
    gap: 12px;
    padding: 4px 0;
    font-size: 13px;
    
    &.info { color: #d4d4d4; }
    &.success { color: #4ec9b0; }
    &.error { color: #f48771; }
    
    .log-time {
      color: #858585;
      flex-shrink: 0;
    }
  }
  
  .empty-log {
    color: #858585;
    text-align: center;
    padding: 20px;
  }
}

.pending-list {
  .pending-item {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 12px;
    background: #f5f7fa;
    border-radius: 6px;
    margin-bottom: 8px;
    
    .order-no {
      font-family: 'Consolas', monospace;
      font-size: 13px;
      color: #303133;
    }
    
    .order-info {
      flex: 1;
      text-align: right;
      margin-right: 12px;
      
      span {
        margin-right: 12px;
        font-size: 13px;
        color: #606266;
      }
      
      .amount {
        color: #f56c6c;
        font-weight: 600;
      }
    }
  }
  
  .empty,
  .more {
    text-align: center;
    padding: 16px;
    color: #909399;
    font-size: 13px;
  }
  
  .more {
    padding: 8px;
  }
}
</style>
