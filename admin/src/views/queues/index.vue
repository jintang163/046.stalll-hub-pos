<template>
  <div class="queues-page">
    <div class="page-header">
      <h2 class="page-title">排队管理</h2>
      <div class="header-actions">
        <el-button type="primary" @click="openCallDialog">
          <el-icon><Bell /></el-icon>叫号
        </el-button>
        <el-button type="success" @click="handleCallNext">
          <el-icon><Right /></el-icon>下一位
        </el-button>
        <el-button type="warning" @click="openConfigDialog">
          <el-icon><Setting /></el-icon>排队配置
        </el-button>
        <el-button @click="fetchData">
          <el-icon><Refresh /></el-icon>刷新
        </el-button>
      </div>
    </div>

    <div class="card-wrapper">
      <el-tabs v-model="activeTab" type="card" @tab-change="handleTabChange">
        <el-tab-pane label="小桌(A)" name="A">
          <span class="tab-badge" :class="{ active: activeTab === 'A' }">
            等待 {{ waitingCounts.A || 0 }} 桌
          </span>
        </el-tab-pane>
        <el-tab-pane label="中桌(B)" name="B">
          <span class="tab-badge" :class="{ active: activeTab === 'B' }">
            等待 {{ waitingCounts.B || 0 }} 桌
          </span>
        </el-tab-pane>
        <el-tab-pane label="大桌(C)" name="C">
          <span class="tab-badge" :class="{ active: activeTab === 'C' }">
            等待 {{ waitingCounts.C || 0 }} 桌
          </span>
        </el-tab-pane>
      </el-tabs>

      <div class="queue-groups" v-loading="loading">
        <div v-for="group in groupedQueues" :key="group.status" class="queue-group">
          <div class="group-header">
            <span class="group-title">{{ group.title }}</span>
            <el-tag :type="group.tagType" size="small">{{ group.list.length }}</el-tag>
          </div>
          <div v-if="group.list.length === 0" class="empty-state">
            <el-empty description="暂无数据" :image-size="80" />
          </div>
          <div v-else class="queue-list">
            <div v-for="item in group.list" :key="item.id" class="queue-card">
              <div class="queue-info">
                <div class="queue-number">{{ item.queue_number }}</div>
                <div class="queue-details">
                  <div class="name-row">
                    <span class="name">{{ item.name }}</span>
                    <el-tag size="small" type="info">{{ item.people_count }}人</el-tag>
                  </div>
                  <div class="phone">{{ item.phone }}</div>
                  <div class="meta-row">
                    <span>等待时长: {{ formatWaitTime(item.created_at) }}</span>
                    <span v-if="group.status === 'waiting'" class="ahead">前面还有 {{ getAheadCount(item) }} 桌</span>
                  </div>
                </div>
              </div>
              <div class="queue-actions">
                <el-button
                  v-if="group.status === 'waiting'"
                  type="primary"
                  size="small"
                  @click="handleCall(item)">
                  叫号
                </el-button>
                <el-button
                  v-if="group.status === 'called'"
                  type="success"
                  size="small"
                  @click="handleArrive(item)">
                  安排入座
                </el-button>
                <el-button
                  v-if="group.status === 'waiting' || group.status === 'called'"
                  type="danger"
                  size="small"
                  @click="handleCancel(item)">
                  取消
                </el-button>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <el-dialog
      v-model="callDialogVisible"
      title="叫号确认"
      width="500px"
      :close-on-click-modal="false">
      <div class="call-dialog-content">
        <div class="call-queue-info">
          <div class="queue-number-large">{{ currentCallItem?.queue_number }}</div>
          <div class="call-details">
            <div class="call-name">{{ currentCallItem?.name }}</div>
            <div class="call-phone">{{ currentCallItem?.phone }}</div>
            <div class="call-people">{{ currentCallItem?.people_count }} 人用餐</div>
          </div>
        </div>
        <div class="voice-section">
          <el-switch
            v-model="voiceEnabled"
            active-text="语音播报"
            inactive-text="关闭语音" />
        </div>
      </div>
      <template #footer>
        <el-button @click="callDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="callLoading" @click="confirmCall">
          确认叫号
        </el-button>
      </template>
    </el-dialog>

    <el-dialog
      v-model="configDialogVisible"
      title="排队配置"
      width="700px"
      :close-on-click-modal="false">
      <el-form
        ref="configFormRef"
        :model="configForm"
        :rules="configRules"
        label-width="120px">
        <el-divider content-position="left">桌型配置</el-divider>
        <div class="table-config-grid">
          <div v-for="tableType in ['A', 'B', 'C']" :key="tableType" class="table-config-card">
            <div class="table-type-title">{{ getTableTypeName(tableType) }}</div>
            <el-form-item :label="'前缀'" :prop="`table_types.${tableType}.prefix`">
              <el-input v-model="configForm.table_types[tableType].prefix" :placeholder="`如${tableType.toLowerCase()}`" />
            </el-form-item>
            <el-form-item :label="'容纳人数'" :prop="`table_types.${tableType}.capacity`">
              <el-input-number v-model="configForm.table_types[tableType].capacity" :min="1" :max="20" style="width: 100%" />
            </el-form-item>
          </div>
        </div>

        <el-divider content-position="left">叫号设置</el-divider>
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="自动叫号" prop="auto_call">
              <el-switch v-model="configForm.auto_call" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="叫号间隔(秒)" prop="call_interval">
              <el-input-number v-model="configForm.call_interval" :min="5" :max="300" style="width: 100%" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="最大叫号次数" prop="max_call_times">
              <el-input-number v-model="configForm.max_call_times" :min="1" :max="10" style="width: 100%" />
            </el-form-item>
          </el-col>
        </el-row>

        <el-divider content-position="left">过期设置</el-divider>
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="自动过期" prop="auto_expire">
              <el-switch v-model="configForm.auto_expire" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="过期时间(分钟)" prop="expire_time">
              <el-input-number v-model="configForm.expire_time" :min="1" :max="120" style="width: 100%" />
            </el-form-item>
          </el-col>
        </el-row>

        <el-divider content-position="left">通知设置</el-divider>
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="语音通知" prop="voice_notify">
              <el-switch v-model="configForm.voice_notify" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="短信通知" prop="sms_notify">
              <el-switch v-model="configForm.sms_notify" />
            </el-form-item>
          </el-col>
        </el-row>
      </el-form>
      <template #footer>
        <el-button @click="configDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="configLoading" @click="saveConfig">
          保存配置
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted, onUnmounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Bell, Right, Setting, Refresh } from '@element-plus/icons-vue'
import { queueApi, queueConfigApi } from '@/api/tables'

const loading = ref(false)
const callLoading = ref(false)
const configLoading = ref(false)

const activeTab = ref('A')
const queueList = ref([])
const waitingCounts = ref({ A: 0, B: 0, C: 0 })

const callDialogVisible = ref(false)
const configDialogVisible = ref(false)
const currentCallItem = ref(null)
const voiceEnabled = ref(true)

const configFormRef = ref()
const configForm = reactive({
  table_types: {
    A: { prefix: 'a', capacity: 2 },
    B: { prefix: 'b', capacity: 4 },
    C: { prefix: 'c', capacity: 8 }
  },
  auto_call: false,
  call_interval: 30,
  max_call_times: 3,
  auto_expire: true,
  expire_time: 15,
  voice_notify: true,
  sms_notify: false
})

const configRules = {
  auto_call: [{ required: true, message: '请选择是否自动叫号', trigger: 'change' }],
  call_interval: [{ required: true, message: '请输入叫号间隔', trigger: 'blur' }],
  max_call_times: [{ required: true, message: '请输入最大叫号次数', trigger: 'blur' }],
  auto_expire: [{ required: true, message: '请选择是否自动过期', trigger: 'change' }],
  expire_time: [{ required: true, message: '请输入过期时间', trigger: 'blur' }],
  voice_notify: [{ required: true, message: '请选择是否语音通知', trigger: 'change' }],
  sms_notify: [{ required: true, message: '请选择是否短信通知', trigger: 'change' }]
}

const statusGroups = [
  { status: 'waiting', title: '等待中', tagType: 'primary' },
  { status: 'called', title: '已叫号', tagType: 'warning' },
  { status: 'seated', title: '已入座', tagType: 'success' },
  { status: 'cancelled', title: '已取消', tagType: 'info' },
  { status: 'expired', title: '已过号', tagType: 'danger' }
]

const groupedQueues = computed(() => {
  const filtered = queueList.value.filter(item => item.queue_type === activeTab.value)
  return statusGroups.map(group => ({
    ...group,
    list: filtered.filter(item => item.status === group.status)
  }))
})

function getTableTypeName(type) {
  const names = { A: '小桌(A)', B: '中桌(B)', C: '大桌(C)' }
  return names[type] || type
}

function formatWaitTime(createdAt) {
  if (!createdAt) return '--'
  const now = new Date()
  const created = new Date(createdAt)
  const diff = Math.floor((now - created) / 1000)
  const hours = Math.floor(diff / 3600)
  const minutes = Math.floor((diff % 3600) / 60)
  const seconds = diff % 60
  if (hours > 0) {
    return `${hours}小时${minutes}分钟`
  } else if (minutes > 0) {
    return `${minutes}分钟${seconds}秒`
  } else {
    return `${seconds}秒`
  }
}

function getAheadCount(item) {
  const waitingList = queueList.value.filter(
    q => q.queue_type === item.queue_type &&
    q.status === 'waiting' &&
    new Date(q.created_at) < new Date(item.created_at)
  )
  return waitingList.length
}

async function fetchData() {
  loading.value = true
  try {
    await Promise.all([
      fetchQueueList(),
      fetchWaitingCounts()
    ])
  } catch (e) {
    console.error(e)
  } finally {
    loading.value = false
  }
}

async function fetchQueueList() {
  try {
    const res = await queueApi.list({ page_size: 1000 })
    queueList.value = res.list || []
  } catch (e) {
    console.error(e)
  }
}

async function fetchWaitingCounts() {
  try {
    const res = await queueApi.getWaitingCount()
    waitingCounts.value = res || { A: 0, B: 0, C: 0 }
  } catch (e) {
    console.error(e)
  }
}

async function handleTabChange() {
  await fetchQueueList()
}

function openCallDialog(item = null) {
  if (item) {
    currentCallItem.value = item
  } else {
    const waitingList = queueList.value
      .filter(q => q.queue_type === activeTab.value && q.status === 'waiting')
      .sort((a, b) => new Date(a.created_at) - new Date(b.created_at))
    if (waitingList.length === 0) {
      ElMessage.warning('当前没有等待中的排队')
      return
    }
    currentCallItem.value = waitingList[0]
  }
  voiceEnabled.value = true
  callDialogVisible.value = true
}

async function handleCallNext() {
  try {
    await ElMessageBox.confirm(
      `确定叫下一位${getTableTypeName(activeTab.value)}吗？`,
      '提示',
      { type: 'warning', confirmButtonText: '确定', cancelButtonText: '取消' }
    )
    callLoading.value = true
    await queueApi.callNext(null, activeTab.value)
    ElMessage.success('叫号成功')
    playVoice(currentCallItem.value)
    await fetchData()
  } catch (e) {
    if (e !== 'cancel') {
      console.error(e)
    }
  } finally {
    callLoading.value = false
  }
}

function handleCall(item) {
  openCallDialog(item)
}

async function confirmCall() {
  if (!currentCallItem.value) return
  try {
    callLoading.value = true
    await queueApi.call({ id: currentCallItem.value.id })
    ElMessage.success('叫号成功')
    callDialogVisible.value = false
    if (voiceEnabled.value) {
      playVoice(currentCallItem.value)
    }
    await fetchData()
  } catch (e) {
    console.error(e)
  } finally {
    callLoading.value = false
  }
}

function playVoice(item) {
  if (!item) return
  const text = `请${item.queue_number}号，${item.name}，到${getTableTypeName(item.queue_type)}用餐`
  if ('speechSynthesis' in window) {
    const utterance = new SpeechSynthesisUtterance(text)
    utterance.lang = 'zh-CN'
    utterance.rate = 0.9
    utterance.pitch = 1
    window.speechSynthesis.speak(utterance)
  }
}

async function handleCancel(item) {
  try {
    await ElMessageBox.confirm(
      `确定取消排队"${item.queue_number}"吗？`,
      '提示',
      { type: 'warning', confirmButtonText: '确定', cancelButtonText: '取消' }
    )
    await queueApi.cancel({ id: item.id })
    ElMessage.success('取消成功')
    await fetchData()
  } catch (e) {
    if (e !== 'cancel') {
      console.error(e)
    }
  }
}

async function handleArrive(item) {
  try {
    await ElMessageBox.confirm(
      `确定安排"${item.queue_number}"入座吗？`,
      '提示',
      { type: 'warning', confirmButtonText: '确定', cancelButtonText: '取消' }
    )
    await queueApi.arrive({ id: item.id })
    ElMessage.success('安排入座成功')
    await fetchData()
  } catch (e) {
    if (e !== 'cancel') {
      console.error(e)
    }
  }
}

async function openConfigDialog() {
  try {
    const res = await queueConfigApi.get()
    if (res) {
      Object.assign(configForm, {
        table_types: res.table_types || {
          A: { prefix: 'a', capacity: 2 },
          B: { prefix: 'b', capacity: 4 },
          C: { prefix: 'c', capacity: 8 }
        },
        auto_call: res.auto_call ?? false,
        call_interval: res.call_interval ?? 30,
        max_call_times: res.max_call_times ?? 3,
        auto_expire: res.auto_expire ?? true,
        expire_time: res.expire_time ?? 15,
        voice_notify: res.voice_notify ?? true,
        sms_notify: res.sms_notify ?? false
      })
    }
    configDialogVisible.value = true
  } catch (e) {
    console.error(e)
  }
}

async function saveConfig() {
  try {
    await configFormRef.value.validate()
    configLoading.value = true
    await queueConfigApi.save(configForm)
    ElMessage.success('配置保存成功')
    configDialogVisible.value = false
  } catch (e) {
    console.error(e)
  } finally {
    configLoading.value = false
  }
}

let refreshTimer = null

onMounted(() => {
  fetchData()
  refreshTimer = setInterval(() => {
    fetchWaitingCounts()
  }, 30000)
})

onUnmounted(() => {
  if (refreshTimer) {
    clearInterval(refreshTimer)
  }
})
</script>

<style scoped lang="scss">
.queues-page {
  .header-actions {
    display: flex;
    gap: 12px;
  }

  .tab-badge {
    margin-left: 8px;
    font-size: 12px;
    color: #909399;

    &.active {
      color: #409eff;
      font-weight: 600;
    }
  }

  .queue-groups {
    margin-top: 20px;
    display: flex;
    flex-wrap: wrap;
    gap: 20px;
  }

  .queue-group {
    flex: 1;
    min-width: 280px;
    background: #fafafa;
    border-radius: 8px;
    padding: 16px;

    .group-header {
      display: flex;
      align-items: center;
      justify-content: space-between;
      margin-bottom: 16px;
      padding-bottom: 12px;
      border-bottom: 1px solid #ebeef5;

      .group-title {
        font-size: 16px;
        font-weight: 600;
        color: #303133;
      }
    }

    .empty-state {
      padding: 40px 0;
    }
  }

  .queue-list {
    display: flex;
    flex-direction: column;
    gap: 12px;
    max-height: 500px;
    overflow-y: auto;
  }

  .queue-card {
    background: #fff;
    border-radius: 8px;
    padding: 16px;
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);
    display: flex;
    align-items: center;
    justify-content: space-between;

    .queue-info {
      display: flex;
      align-items: center;
      gap: 16px;
      flex: 1;

      .queue-number {
        font-size: 28px;
        font-weight: 700;
        color: #409eff;
        min-width: 80px;
        text-align: center;
      }

      .queue-details {
        flex: 1;

        .name-row {
          display: flex;
          align-items: center;
          gap: 8px;
          margin-bottom: 4px;

          .name {
            font-size: 16px;
            font-weight: 600;
            color: #303133;
          }
        }

        .phone {
          color: #606266;
          margin-bottom: 4px;
        }

        .meta-row {
          display: flex;
          gap: 16px;
          font-size: 12px;
          color: #909399;

          .ahead {
            color: #e6a23c;
            font-weight: 500;
          }
        }
      }
    }

    .queue-actions {
      display: flex;
      gap: 8px;
    }
  }

  .call-dialog-content {
    text-align: center;
    padding: 20px 0;

    .call-queue-info {
      display: flex;
      align-items: center;
      justify-content: center;
      gap: 30px;
      margin-bottom: 30px;

      .queue-number-large {
        font-size: 64px;
        font-weight: 700;
        color: #409eff;
      }

      .call-details {
        text-align: left;

        .call-name {
          font-size: 24px;
          font-weight: 600;
          margin-bottom: 8px;
        }

        .call-phone {
          font-size: 16px;
          color: #606266;
          margin-bottom: 8px;
        }

        .call-people {
          font-size: 14px;
          color: #909399;
        }
      }
    }

    .voice-section {
      display: flex;
      justify-content: center;
      padding-top: 20px;
      border-top: 1px solid #ebeef5;
    }
  }

  .table-config-grid {
    display: grid;
    grid-template-columns: repeat(3, 1fr);
    gap: 20px;
    margin-bottom: 20px;

    .table-config-card {
      background: #f5f7fa;
      border-radius: 8px;
      padding: 16px;

      .table-type-title {
        font-size: 16px;
        font-weight: 600;
        margin-bottom: 16px;
        text-align: center;
        color: #303133;
      }
    }
  }
}
</style>
