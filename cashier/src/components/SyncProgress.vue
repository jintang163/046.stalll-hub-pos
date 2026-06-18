<template>
  <div v-if="visible" class="sync-progress-container">
    <div class="sync-progress-card">
      <div class="sync-header">
        <el-icon :class="statusIcon"><component :is="statusIcon" /></el-icon>
        <span>{{ progress.message || '数据同步中' }}</span>
      </div>
      <el-progress 
        :percentage="progress.percent" 
        :status="progressStatus"
        :stroke-width="8"
      />
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { useSyncStore } from '@/store/sync'
import { Loading, CircleCheck, CircleClose } from '@element-plus/icons-vue'

const syncStore = useSyncStore()

const visible = computed(() => syncStore.progress.status === 'syncing')
const progress = computed(() => syncStore.progress)

const progressStatus = computed(() => {
  switch (syncStore.progress.status) {
    case 'completed': return 'success'
    case 'error': return 'exception'
    default: return null
  }
})

const statusIcon = computed(() => {
  switch (syncStore.progress.status) {
    case 'completed': return CircleCheck
    case 'error': return CircleClose
    default: return Loading
  }
})
</script>

<style lang="scss" scoped>
.sync-progress-container {
  position: fixed;
  top: 20px;
  right: 20px;
  z-index: 9999;
}

.sync-progress-card {
  background: #fff;
  border-radius: 8px;
  padding: 16px 20px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
  min-width: 320px;
}

.sync-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 12px;
  font-weight: 500;
  color: #303133;
  
  .el-icon {
    font-size: 20px;
    color: #409eff;
    animation: spin 2s linear infinite;
  }
  
  @keyframes spin {
    from { transform: rotate(0deg); }
    to { transform: rotate(360deg); }
  }
}
</style>
