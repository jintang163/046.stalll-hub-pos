<template>
  <div class="settings-page">
    <header class="page-header">
      <h1>系统设置</h1>
      <div class="header-actions">
        <el-button @click="goBack">
          <el-icon><ArrowLeft /></el-icon>
          返回
        </el-button>
      </div>
    </header>

    <div class="page-body">
      <el-card class="settings-card">
        <template #header>
          <div class="card-header">
            <span>基础配置</span>
          </div>
        </template>

        <el-form 
          ref="formRef" 
          :model="formData" 
          label-width="120px"
          label-position="left"
        >
          <el-form-item label="门店ID">
            <el-input-number v-model="formData.storeID" :min="1" />
          </el-form-item>

          <el-divider />

          <el-form-item label="API地址">
            <el-input v-model="formData.apiBaseURL" placeholder="http://localhost:8080/api/v1" />
          </el-form-item>

          <el-form-item label="NSQD地址">
            <el-input v-model="formData.nsqd" placeholder="localhost:4150" />
          </el-form-item>

          <el-form-item label="NSQ Lookupd">
            <el-input v-model="formData.nsqLookupd" placeholder="http://localhost:4161" />
          </el-form-item>

          <el-divider />

          <el-form-item label="打印机配置">
            <el-select v-model="formData.printerType" placeholder="选择打印机类型">
              <el-option label="无" value="" />
              <el-option label="USB打印机" value="usb" />
              <el-option label="网络打印机" value="network" />
              <el-option label="蓝牙打印机" value="bluetooth" />
            </el-select>
          </el-form-item>

          <el-form-item v-if="formData.printerType === 'network'" label="打印机地址">
            <el-input v-model="formData.printerAddress" placeholder="192.168.1.100:9100" />
          </el-form-item>

          <el-form-item label="自动打印">
            <el-switch v-model="formData.autoPrint" />
          </el-form-item>

          <el-divider />

          <el-form-item label="自动同步间隔">
            <el-select v-model="formData.syncInterval">
              <el-option label="30秒" :value="30" />
              <el-option label="1分钟" :value="60" />
              <el-option label="5分钟" :value="300" />
              <el-option label="10分钟" :value="600" />
            </el-select>
          </el-form-item>

          <el-form-item label="订单自动上传">
            <el-switch v-model="formData.autoUpload" />
          </el-form-item>

          <el-form-item>
            <el-button type="primary" @click="handleSave">保存设置</el-button>
            <el-button @click="handleReset">重置</el-button>
          </el-form-item>
        </el-form>
      </el-card>

      <el-card class="settings-card mt-20">
        <template #header>
          <div class="card-header">
            <span>系统信息</span>
          </div>
        </template>

        <el-descriptions :column="1" border>
          <el-descriptions-item label="软件版本">
            {{ appVersion }}
          </el-descriptions-item>
          <el-descriptions-item label="Electron版本">
            {{ process.versions.electron }}
          </el-descriptions-item>
          <el-descriptions-item label="Node版本">
            {{ process.versions.node }}
          </el-descriptions-item>
          <el-descriptions-item label="Chrome版本">
            {{ process.versions.chrome }}
          </el-descriptions-item>
          <el-descriptions-item label="数据存储路径">
            <span class="path-text">{{ userDataPath }}</span>
          </el-descriptions-item>
        </el-descriptions>
      </el-card>

      <el-card class="settings-card mt-20">
        <template #header>
          <div class="card-header">
            <span>数据管理</span>
          </div>
        </template>

        <div class="data-actions">
          <el-button type="warning" @click="clearLocalData">
            <el-icon><Delete /></el-icon>
            清空本地商品数据
          </el-button>
          <el-button type="danger" @click="clearAllData">
            <el-icon><Warning /></el-icon>
            重置所有数据
          </el-button>
        </div>

        <el-alert 
          type="warning" 
          title="注意" 
          description="清空数据会删除本地所有缓存，请谨慎操作！建议先同步所有订单。"
          :closable="false"
          show-icon
          style="margin-top: 16px;"
        />
      </el-card>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { ArrowLeft, Delete, Warning } from '@element-plus/icons-vue'
import { useSyncStore } from '@/store/sync'

const router = useRouter()
const syncStore = useSyncStore()

const formRef = ref(null)
const appVersion = ref('1.0.0')
const userDataPath = ref('')

const defaultForm = {
  storeID: 1,
  apiBaseURL: 'http://localhost:8080/api/v1',
  nsqd: 'localhost:4150',
  nsqLookupd: 'http://localhost:4161',
  printerType: '',
  printerAddress: '',
  autoPrint: false,
  syncInterval: 60,
  autoUpload: true
}

const formData = reactive({ ...defaultForm })

const loadConfig = async () => {
  if (!window.electronAPI) return
  
  try {
    const config = await window.electronAPI.app.getConfig()
    Object.assign(formData, {
      storeID: config.storeID || 1,
      apiBaseURL: config.apiBaseURL,
      nsqd: config.nsqd,
      nsqLookupd: config.nsqLookupd,
      printerType: config.printerType || '',
      printerAddress: config.printerAddress || '',
      autoPrint: config.autoPrint || false,
      syncInterval: config.syncInterval || 60,
      autoUpload: config.autoUpload !== false
    })
    
    appVersion.value = await window.electronAPI.app.getVersion()
  } catch (e) {
    console.error('加载配置失败:', e)
  }
}

const handleSave = async () => {
  if (!window.electronAPI) return
  
  try {
    await window.electronAPI.app.setConfig({
      storeID: formData.storeID,
      apiBaseURL: formData.apiBaseURL,
      nsqd: formData.nsqd,
      nsqLookupd: formData.nsqLookupd,
      printerType: formData.printerType,
      printerAddress: formData.printerAddress,
      autoPrint: formData.autoPrint,
      syncInterval: formData.syncInterval,
      autoUpload: formData.autoUpload
    })
    
    syncStore.storeID = formData.storeID
    
    ElMessage.success('设置已保存')
  } catch (e) {
    ElMessage.error('保存失败: ' + e.message)
  }
}

const handleReset = () => {
  Object.assign(formData, defaultForm)
}

const clearLocalData = async () => {
  try {
    await ElMessageBox.confirm(
      '确定要清空本地商品数据吗？下次启动时需要重新同步。',
      '确认操作',
      { type: 'warning' }
    )
    
    if (window.electronAPI) {
      await window.electronAPI.db.clearAllProducts()
      await window.electronAPI.db.setLastSyncTime(null)
      await window.electronAPI.db.setLastSyncID(0)
    }
    
    ElMessage.success('本地商品数据已清空')
  } catch {}
}

const clearAllData = async () => {
  try {
    await ElMessageBox.confirm(
      '确定要重置所有数据吗？这将删除所有本地数据，包括订单！此操作不可恢复！',
      '危险操作',
      { type: 'error' }
    )
    
    if (window.electronAPI) {
      await window.electronAPI.db.clearAllProducts()
      await window.electronAPI.db.setLastSyncTime(null)
      await window.electronAPI.db.setLastSyncID(0)
    }
    
    ElMessage.success('所有数据已重置')
  } catch {}
}

const goBack = () => {
  router.push('/')
}

onMounted(() => {
  loadConfig()
})
</script>

<style lang="scss" scoped>
.settings-page {
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

.settings-card {
  max-width: 800px;
  margin: 0 auto;
}

.card-header {
  font-weight: 600;
  font-size: 16px;
}

.path-text {
  font-family: 'Consolas', monospace;
  font-size: 13px;
  color: #606266;
  word-break: break-all;
}

.data-actions {
  display: flex;
  gap: 16px;
}
</style>
