<template>
  <div class="recommendations-page">
    <div class="page-header">
      <h2 class="page-title">智能推荐配置</h2>
      <div class="header-actions">
        <el-button @click="handleReset">
          <el-icon><Refresh /></el-icon>重置
        </el-button>
        <el-button type="primary" @click="handleSave" :loading="saving">
          <el-icon><Check /></el-icon>保存配置
        </el-button>
      </div>
    </div>

    <div class="card-wrapper">
      <el-row :gutter="20">
        <el-col :span="16">
          <el-card class="config-card" shadow="never">
            <template #header>
              <div class="card-header">
                <el-icon color="#667eea" :size="20"><Setting /></el-icon>
                <span class="card-title">推荐权重与参数配置</span>
                <el-tag size="small" type="success" effect="plain" class="header-tag">
                  调整后即时生效
                </el-tag>
              </div>
            </template>

            <el-form
              ref="configFormRef"
              :model="configData"
              label-width="200px"
              label-position="right"
              size="default"
            >
              <el-divider content-position="left"><span class="divider-text">权重配置</span></el-divider>

              <template v-for="item in weightMetaList" :key="item.key">
                <el-form-item :label="itemLabel(item)" :prop="item.key">
                  <el-tooltip :content="item.description" placement="top" :show-after="300">
                    <div class="slider-wrapper">
                      <el-slider
                        v-if="item.type === 'slider'"
                        v-model="configData[item.key]"
                        :min="item.min"
                        :max="item.max"
                        :step="item.step"
                        show-stops
                        show-tooltip
                        :marks="getSliderMarks(item)"
                        class="config-slider"
                      />
                      <el-input-number
                        v-else-if="item.type === 'number'"
                        v-model="configData[item.key]"
                        :min="item.min"
                        :max="item.max"
                        :step="item.step"
                        :controls="false"
                        size="default"
                      />
                      <el-switch
                        v-else-if="item.type === 'switch'"
                        v-model="configData[item.key]"
                      />
                      <span v-if="item.unit" class="unit-label">{{ item.unit }}</span>
                      <span class="value-badge">
                        {{ formatValue(item) }}
                      </span>
                    </div>
                  </el-tooltip>
                </el-form-item>
              </template>

              <el-divider content-position="left"><span class="divider-text">算法参数</span></el-divider>

              <template v-for="item in paramMetaList" :key="item.key">
                <el-form-item :label="itemLabel(item)" :prop="item.key">
                  <el-tooltip :content="item.description" placement="top" :show-after="300">
                    <div class="slider-wrapper">
                      <el-slider
                        v-if="item.type === 'slider'"
                        v-model="configData[item.key]"
                        :min="item.min"
                        :max="item.max"
                        :step="item.step"
                        show-stops
                        :marks="getSliderMarks(item)"
                        class="config-slider"
                      />
                      <el-input-number
                        v-else-if="item.type === 'number'"
                        v-model="configData[item.key]"
                        :min="item.min"
                        :max="item.max"
                        :step="item.step"
                        size="default"
                      />
                      <el-switch
                        v-else-if="item.type === 'switch'"
                        v-model="configData[item.key]"
                      />
                      <span v-if="item.unit" class="unit-label">{{ item.unit }}</span>
                      <span class="value-badge">
                        {{ formatValue(item) }}
                      </span>
                    </div>
                  </el-tooltip>
                </el-form-item>
              </template>
            </el-form>
          </el-card>
        </el-col>

        <el-col :span="8">
          <el-card class="status-card" shadow="never">
            <template #header>
              <div class="card-header">
                <el-icon color="#67c23a" :size="20"><DataBoard /></el-icon>
                <span class="card-title">推荐计算状态</span>
              </div>
            </template>

            <el-descriptions :column="1" border size="small" class="status-descriptions">
              <el-descriptions-item label="服务状态">
                <el-tag :type="configData.enabled ? 'success' : 'info'" effect="dark" round>
                  {{ configData.enabled ? '已启用' : '已关闭' }}
                </el-tag>
              </el-descriptions-item>
              <el-descriptions-item label="刷新任务">
                <template v-if="statusData.is_running">
                  <el-tag type="warning" effect="dark" round>
                    <span class="status-dot running"></span>正在计算
                  </el-tag>
                </template>
                <template v-else>
                  <el-tag type="success" effect="plain" round>
                    <span class="status-dot"></span>空闲
                  </el-tag>
                </template>
              </el-descriptions-item>
              <el-descriptions-item label="上次刷新">
                <span v-if="statusData.last_refreshed_at" class="last-refresh">
                  {{ statusData.last_refreshed_at }}
                </span>
                <span v-else class="text-muted">尚未执行</span>
              </el-descriptions-item>
              <el-descriptions-item label="已覆盖商品">
                <span class="num-emphasis">{{ statusData.total_products || 0 }}</span> 件
              </el-descriptions-item>
              <el-descriptions-item label="推荐规则数">
                <span class="num-emphasis">{{ statusData.total_pairs || 0 }}</span> 对
              </el-descriptions-item>
            </el-descriptions>

            <el-divider />

            <div class="action-box">
              <el-button
                type="success"
                size="large"
                style="width: 100%"
                @click="handleTriggerRefresh"
                :loading="refreshing"
              >
                <el-icon><Promotion /></el-icon>
                {{ refreshing ? '正在刷新推荐...' : '立即刷新推荐' }}
              </el-button>
              <p class="tip-text">
                小提示：首次使用请先点击"立即刷新推荐"。
                后续系统每 {{ configData.refresh_interval_hours || 6 }} 小时自动刷新一次。
              </p>
            </div>
          </el-card>

          <el-card class="preview-card" shadow="never" style="margin-top: 20px">
            <template #header>
              <div class="card-header">
                <el-icon color="#e6a23c" :size="20"><Histogram /></el-icon>
                <span class="card-title">权重占比预览</span>
              </div>
            </template>
            <div class="weight-preview">
              <div class="weight-bar">
                <div
                  class="weight-segment cf-weight"
                  :style="{ width: cfPct + '%' }"
                  :title="`协同过滤: ${configData.cf_weight}`"
                ></div>
                <div
                  class="weight-segment hot-weight"
                  :style="{ width: hotPct + '%' }"
                  :title="`热门榜: ${configData.hot_weight}`"
                ></div>
                <div
                  class="weight-segment user-weight"
                  :style="{ width: userPct + '%' }"
                  :title="`用户历史: ${configData.user_history_weight}`"
                ></div>
                <div
                  class="weight-segment div-weight"
                  :style="{ width: divPct + '%' }"
                  :title="`多样性: ${configData.category_diversity_weight}`"
                ></div>
              </div>
              <div class="weight-legend">
                <div class="legend-item">
                  <span class="legend-dot cf-weight"></span>
                  协同过滤 {{ cfPct }}%
                </div>
                <div class="legend-item">
                  <span class="legend-dot hot-weight"></span>
                  热门榜 {{ hotPct }}%
                </div>
                <div class="legend-item">
                  <span class="legend-dot user-weight"></span>
                  用户历史 {{ userPct }}%
                </div>
                <div class="legend-item">
                  <span class="legend-dot div-weight"></span>
                  多样性 {{ divPct }}%
                </div>
              </div>
            </div>
          </el-card>
        </el-col>
      </el-row>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  Check, Refresh, Setting, DataBoard, Promotion, Histogram
} from '@element-plus/icons-vue'
import {
  getRecommendConfig,
  getRecommendConfigMeta,
  updateRecommendConfig,
  triggerRecommendRefresh,
  getRecommendRefreshStatus
} from '@/api/recommend'

const configFormRef = ref(null)
const loading = ref(false)
const saving = ref(false)
const refreshing = ref(false)

const configData = reactive({})
const metaList = ref([])
const statusData = reactive({})

const weightKeys = ['cf_weight', 'hot_weight', 'user_history_weight', 'category_diversity_weight']
const paramKeys = [
  'recommend_count', 'min_order_pairs', 'min_similarity',
  'cf_days', 'hot_days', 'user_history_days', 'user_history_top_k',
  'enabled', 'auto_refresh', 'refresh_interval_hours'
]

const weightMetaList = computed(() =>
  metaList.value.filter(m => weightKeys.includes(m.key))
)
const paramMetaList = computed(() =>
  metaList.value.filter(m => paramKeys.includes(m.key))
)

const totalWeight = computed(() => {
  let total = 0
  for (const k of weightKeys) {
    total += (Number(configData[k]) || 0)
  }
  return total || 1
})
const cfPct = computed(() => Math.round((Number(configData.cf_weight || 0) / totalWeight.value) * 100))
const hotPct = computed(() => Math.round((Number(configData.hot_weight || 0) / totalWeight.value) * 100))
const userPct = computed(() => Math.round((Number(configData.user_history_weight || 0) / totalWeight.value) * 100))
const divPct = computed(() => {
  const remaining = 100 - cfPct.value - hotPct.value - userPct.value
  return remaining >= 0 ? remaining : 0
})

const itemLabel = (item) => `${item.label}${item.unit ? `（${item.unit}）` : ''}`

const formatValue = (item) => {
  const v = configData[item.key]
  if (item.type === 'switch') return v ? '开' : '关'
  if (item.type === 'slider') {
    return typeof v === 'number' ? v.toFixed(2) : v
  }
  return v
}

const getSliderMarks = (item) => {
  const marks = {}
  const range = item.max - item.min
  const stepCount = 4
  for (let i = 0; i <= stepCount; i++) {
    const val = item.min + (range * i / stepCount)
    const display = Number.isInteger(item.step)
      ? Math.round(val)
      : Number(val.toFixed(2))
    marks[display] = String(display)
  }
  return marks
}

const loadMeta = async () => {
  try {
    metaList.value = await getRecommendConfigMeta()
    metaList.value.forEach(m => {
      if (!(m.key in configData)) {
        configData[m.key] = m.default
      }
    })
  } catch (e) {
    console.error('加载配置元数据失败', e)
  }
}

const loadConfig = async () => {
  loading.value = true
  try {
    const data = await getRecommendConfig()
    Object.keys(data).forEach(k => {
      if (k !== 'id' && k !== 'store_id' && k !== 'last_refreshed_at' && !k.endsWith('_at')) {
        configData[k] = data[k]
      }
    })
  } catch (e) {
    ElMessage.error('加载推荐配置失败')
  } finally {
    loading.value = false
  }
}

const loadStatus = async () => {
  try {
    const data = await getRecommendRefreshStatus()
    Object.assign(statusData, data)
  } catch (e) {}
}

const handleSave = async () => {
  saving.value = true
  try {
    const payload = {}
    metaList.value.forEach(m => {
      payload[m.key] = configData[m.key]
    })
    await updateRecommendConfig(payload)
    ElMessage.success('推荐配置保存成功')
    setTimeout(() => loadStatus(), 500)
  } catch (e) {
    ElMessage.error('保存失败：' + (e.message || '未知错误'))
  } finally {
    saving.value = false
  }
}

const handleReset = async () => {
  await ElMessageBox.confirm(
    '确定要重置为系统默认值吗？',
    '重置确认',
    { type: 'warning' }
  )
  metaList.value.forEach(m => {
    configData[m.key] = m.default
  })
  ElMessage.success('已重置为默认值，记得点"保存"提交哦')
}

const handleTriggerRefresh = async () => {
  refreshing.value = true
  try {
    await triggerRecommendRefresh()
    ElMessage.success('刷新任务已启动，约需 10~60 秒完成')
    statusData.is_running = true
    let attempts = 0
    const poller = setInterval(async () => {
      attempts++
      await loadStatus()
      if (!statusData.is_running || attempts > 30) {
        clearInterval(poller)
        if (statusData.is_running) ElMessage.info('计算在后台继续进行，请稍后刷新页面查看')
        else ElMessage.success('推荐刷新完成')
      }
    }, 2000)
  } catch (e) {
    ElMessage.error('刷新任务启动失败')
  } finally {
    refreshing.value = false
  }
}

onMounted(async () => {
  await loadMeta()
  await Promise.all([loadConfig(), loadStatus()])
  setInterval(loadStatus, 30000)
})
</script>

<style lang="scss" scoped>
.recommendations-page {
  padding: 24px;
  background: #f5f7fa;
  min-height: 100vh;
}

.page-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 24px;

  .page-title {
    margin: 0;
    font-size: 24px;
    font-weight: 700;
    color: #1f2937;
  }

  .header-actions {
    display: flex;
    gap: 12px;
  }
}

.card-wrapper {
  .config-card,
  .status-card,
  .preview-card {
    border-radius: 12px;
    border: 1px solid #eef2f7;
  }

  .card-header {
    display: flex;
    align-items: center;
    gap: 8px;

    .card-title {
      font-size: 16px;
      font-weight: 600;
      color: #1f2937;
      flex: 1;
    }

    .header-tag {
      margin-left: auto;
    }
  }
}

.divider-text {
  font-size: 14px;
  font-weight: 600;
  color: #667eea;
}

.slider-wrapper {
  display: flex;
  align-items: center;
  gap: 16px;
  width: 100%;

  .config-slider {
    flex: 1;
    max-width: 480px;
  }

  .unit-label {
    font-size: 13px;
    color: #909399;
    min-width: 24px;
  }

  .value-badge {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    min-width: 72px;
    padding: 4px 12px;
    background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
    color: #fff;
    border-radius: 16px;
    font-size: 13px;
    font-weight: 600;
  }
}

.status-descriptions {
  :deep(.el-descriptions__label) {
    width: 110px;
    color: #909399;
    font-weight: 500;
  }

  :deep(.el-descriptions__content) {
    font-weight: 500;
  }

  .num-emphasis {
    color: #667eea;
    font-weight: 700;
    font-size: 16px;
  }

  .last-refresh {
    color: #67c23a;
  }

  .text-muted {
    color: #c0c4cc;
  }

  .status-dot {
    display: inline-block;
    width: 8px;
    height: 8px;
    border-radius: 50%;
    background: #67c23a;
    margin-right: 6px;

    &.running {
      background: #e6a23c;
      animation: pulse 1.2s infinite;
    }
  }
}

@keyframes pulse {
  0%, 100% { opacity: 1; transform: scale(1); }
  50% { opacity: 0.5; transform: scale(0.85); }
}

.action-box {
  .tip-text {
    margin: 12px 4px 0;
    font-size: 12px;
    color: #909399;
    line-height: 1.6;
  }
}

.weight-preview {
  .weight-bar {
    height: 28px;
    border-radius: 14px;
    overflow: hidden;
    display: flex;
    background: #f0f2f5;
  }

  .weight-segment {
    height: 100%;
    transition: width 0.3s;

    &.cf-weight { background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); }
    &.hot-weight { background: linear-gradient(135deg, #f56c6c 0%, #ff9966 100%); }
    &.user-weight { background: linear-gradient(135deg, #67c23a 0%, #95d475 100%); }
    &.div-weight { background: linear-gradient(135deg, #e6a23c 0%, #f0c78a 100%); }
  }

  .weight-legend {
    margin-top: 16px;
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 12px;

    .legend-item {
      display: flex;
      align-items: center;
      gap: 8px;
      font-size: 13px;
      color: #606266;
    }

    .legend-dot {
      width: 12px;
      height: 12px;
      border-radius: 4px;
      flex-shrink: 0;

      &.cf-weight { background: #667eea; }
      &.hot-weight { background: #f56c6c; }
      &.user-weight { background: #67c23a; }
      &.div-weight { background: #e6a23c; }
    }
  }
}
</style>
