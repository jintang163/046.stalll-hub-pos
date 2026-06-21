<template>
  <div class="kitchen-display" :class="{ 'flash-alert': kitchenStore.overdueCount > 0 && kitchenStore.config.flashAlert }">
    <header class="header">
      <div class="header-left">
        <div class="title">
          <span class="icon">👨‍🍳</span>
          <span>厨师端显示屏</span>
        </div>
        <div class="status">
          <el-tag :type="kitchenStore.isConnected ? 'success' : 'danger'" effect="dark">
            {{ kitchenStore.isConnected ? '● 已连接' : '● 连接中断' }}
          </el-tag>
        </div>
      </div>
      <div class="header-center">
        <div class="stats">
          <div class="stat-item pending">
            <span class="stat-label">待制作</span>
            <span class="stat-value">{{ kitchenStore.totalPending }}</span>
          </div>
          <div class="stat-item cooking">
            <span class="stat-label">制作中</span>
            <span class="stat-value">{{ kitchenStore.totalCooking }}</span>
          </div>
          <div class="stat-item completed">
            <span class="stat-label">已完成</span>
            <span class="stat-value">{{ kitchenStore.totalCompleted }}</span>
          </div>
          <div class="stat-item overdue" v-if="kitchenStore.overdueCount > 0">
            <span class="stat-label">已超时</span>
            <span class="stat-value">{{ kitchenStore.overdueCount }}</span>
          </div>
        </div>
      </div>
      <div class="header-right">
        <div class="current-time">{{ currentTime }}</div>
        <el-button type="primary" @click="showSettings = true" size="large">
          ⚙️ 设置
        </el-button>
      </div>
    </header>

    <main class="main-content" v-loading="kitchenStore.isLoading">
      <section class="column pending-column">
        <div class="column-header">
          <h2>📋 待制作</h2>
          <span class="count-badge">{{ kitchenStore.pendingItemsWithMeta.length }}</span>
        </div>
        <div class="items-list" ref="pendingListRef">
          <div
            v-for="item in kitchenStore.pendingItemsWithMeta"
            :key="item.id"
            class="item-card"
            :class="{ 'overdue-card': item.isOverdue }"
          >
            <div class="item-header">
              <div class="item-table">
                <el-tag :type="item.isOverdue ? 'danger' : 'warning'" effect="dark">
                  {{ item.table_no || '外带' }}桌
                </el-tag>
              </div>
              <div class="item-order-no">#{{ item.order_no }}</div>
            </div>

            <div class="item-body">
              <div class="item-name" :class="{ 'overdue-text': item.isOverdue }">
                {{ item.product_name }}
              </div>
              <div class="item-sku" v-if="item.sku_name && item.sku_name !== item.product_name">
                {{ item.sku_name }}
              </div>
              <div class="item-spec" v-if="item.attribute_values">
                {{ item.attribute_values }}
              </div>
              <div class="item-remark" v-if="item.remark">
                💬 {{ item.remark }}
              </div>
            </div>

            <div class="item-footer">
              <div class="item-meta">
                <span class="item-quantity">x{{ item.quantity }}</span>
                <span class="item-time" :class="{ 'overdue-text': item.isOverdue }">
                  ⏱️ {{ kitchenStore.formatWaitingTime(item.waitingSeconds) }}
                </span>
              </div>
              <div class="item-actions">
                <el-button
                  type="primary"
                  @click="handleStartCooking(item.id)"
                  size="large"
                >
                  开始制作
                </el-button>
                <el-button
                  type="success"
                  @click="handleMarkCompleted(item.id)"
                  size="large"
                >
                  ✅ 完成
                </el-button>
              </div>
            </div>
          </div>

          <div class="empty-state" v-if="kitchenStore.pendingItemsWithMeta.length === 0">
            <div class="empty-icon">🎉</div>
            <div class="empty-text">暂无待制作菜品</div>
          </div>
        </div>
      </section>

      <section class="column cooking-column">
        <div class="column-header">
          <h2>🔥 制作中</h2>
          <span class="count-badge cooking">{{ kitchenStore.cookingItemsWithMeta.length }}</span>
        </div>
        <div class="items-list" ref="cookingListRef">
          <div
            v-for="item in kitchenStore.cookingItemsWithMeta"
            :key="item.id"
            class="item-card cooking-card"
            :class="{ 'overdue-card': item.isOverdue }"
          >
            <div class="item-header">
              <div class="item-table">
                <el-tag :type="item.isOverdue ? 'danger' : 'primary'" effect="dark">
                  {{ item.table_no || '外带' }}桌
                </el-tag>
              </div>
              <div class="item-order-no">#{{ item.order_no }}</div>
            </div>

            <div class="item-body">
              <div class="item-name" :class="{ 'overdue-text': item.isOverdue }">
                {{ item.product_name }}
              </div>
              <div class="item-sku" v-if="item.sku_name && item.sku_name !== item.product_name">
                {{ item.sku_name }}
              </div>
              <div class="item-spec" v-if="item.attribute_values">
                {{ item.attribute_values }}
              </div>
            </div>

            <div class="item-footer">
              <div class="item-meta">
                <span class="item-quantity">x{{ item.quantity }}</span>
                <span class="item-time" :class="{ 'overdue-text': item.isOverdue }">
                  ⏱️ {{ kitchenStore.formatWaitingTime(item.waitingSeconds) }}
                </span>
              </div>
              <div class="item-actions">
                <el-button
                  type="success"
                  @click="handleMarkCompleted(item.id)"
                  size="large"
                >
                  ✅ 制作完成
                </el-button>
              </div>
            </div>
          </div>

          <div class="empty-state" v-if="kitchenStore.cookingItemsWithMeta.length === 0">
            <div class="empty-icon">🔥</div>
            <div class="empty-text">暂无制作中菜品</div>
          </div>
        </div>
      </section>
    </main>

    <el-dialog
      v-model="showSettings"
      title="⚙️ 系统设置"
      width="600px"
      :close-on-click-modal="false"
    >
      <el-form label-width="140px" label-position="left">
        <el-form-item label="门店 ID">
          <el-input-number
            v-model="kitchenStore.config.storeId"
            :min="1"
            size="large"
            style="width: 100%"
          />
        </el-form-item>
        <el-form-item label="用户 ID">
          <el-input-number
            v-model="kitchenStore.config.userId"
            :min="1"
            size="large"
            style="width: 100%"
          />
        </el-form-item>
        <el-form-item label="API Token">
          <el-input
            v-model="kitchenStore.config.token"
            type="textarea"
            :rows="2"
            placeholder="请输入登录后的 Token"
            size="large"
          />
        </el-form-item>
        <el-form-item label="API 地址">
          <el-input
            v-model="kitchenStore.config.apiBaseUrl"
            placeholder="http://localhost:8080/api/v1"
            size="large"
          />
        </el-form-item>
        <el-form-item label="WebSocket 地址">
          <el-input
            v-model="kitchenStore.config.wsUrl"
            placeholder="留空则根据 API 地址自动推导"
            size="large"
          />
        </el-form-item>
        <el-form-item label="超时时长（分钟）">
          <el-input-number
            v-model="kitchenStore.config.overdueMinutes"
            :min="1"
            :max="120"
            size="large"
            style="width: 100%"
          />
        </el-form-item>
        <el-form-item label="语音提醒">
          <el-switch
            v-model="kitchenStore.config.voiceAlert"
            active-text="开启"
            inactive-text="关闭"
            size="large"
          />
        </el-form-item>
        <el-form-item label="闪烁提醒">
          <el-switch
            v-model="kitchenStore.config.flashAlert"
            active-text="开启"
            inactive-text="关闭"
            size="large"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showSettings = false" size="large">取消</el-button>
        <el-button type="primary" @click="saveAndReconnect" size="large">保存并重连</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useKitchenStore } from './store/kitchen'
import type { KitchenOrderItem } from './types'

const kitchenStore = useKitchenStore()

const showSettings = ref(false)
const currentTime = ref('')
const pendingListRef = ref<HTMLElement | null>(null)
const cookingListRef = ref<HTMLElement | null>(null)

let timeTimer: number | null = null
let tickTimer: number | null = null

const updateTime = () => {
  const now = new Date()
  currentTime.value = now.toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
    hour12: false
  })
}

const handleStartCooking = async (itemId: number) => {
  try {
    await ElMessageBox.confirm('确认开始制作此菜品？', '提示', {
      confirmButtonText: '开始制作',
      cancelButtonText: '取消',
      type: 'info'
    })
    await kitchenStore.startCooking(itemId)
    ElMessage.success('已开始制作')
  } catch (e: any) {
    if (e !== 'cancel') {
      ElMessage.error(e.message || '操作失败')
    }
  }
}

const handleMarkCompleted = async (itemId: number) => {
  try {
    await ElMessageBox.confirm('确认此菜品已制作完成？', '提示', {
      confirmButtonText: '完成',
      cancelButtonText: '取消',
      type: 'success'
    })
    await kitchenStore.markCompleted(itemId)
    ElMessage.success('已标记完成')
  } catch (e: any) {
    if (e !== 'cancel') {
      ElMessage.error(e.message || '操作失败')
    }
  }
}

const saveAndReconnect = () => {
  kitchenStore.saveConfig()
  kitchenStore.disconnect()
  showSettings.value = false
  setTimeout(() => {
    kitchenStore.connect()
  }, 500)
  ElMessage.success('设置已保存，正在重新连接...')
}

onMounted(() => {
  updateTime()
  timeTimer = window.setInterval(updateTime, 1000)

  tickTimer = window.setInterval(() => {
    const pendingEl = pendingListRef.value
    if (pendingEl && pendingEl.scrollHeight > pendingEl.clientHeight) {
      pendingEl.scrollTop += 1
      if (pendingEl.scrollTop >= pendingEl.scrollHeight - pendingEl.clientHeight) {
        pendingEl.scrollTop = 0
      }
    }

    const cookingEl = cookingListRef.value
    if (cookingEl && cookingEl.scrollHeight > cookingEl.clientHeight) {
      cookingEl.scrollTop += 1
      if (cookingEl.scrollTop >= cookingEl.scrollHeight - cookingEl.clientHeight) {
        cookingEl.scrollTop = 0
      }
    }
  }, 50)

  if (!kitchenStore.config.token || !kitchenStore.config.storeId) {
    showSettings.value = true
  } else {
    kitchenStore.connect()
  }
})

onUnmounted(() => {
  if (timeTimer) clearInterval(timeTimer)
  if (tickTimer) clearInterval(tickTimer)
  kitchenStore.disconnect()
})
</script>

<style scoped lang="scss">
.kitchen-display {
  width: 100%;
  height: 100%;
  display: flex;
  flex-direction: column;
  background: #0a0a0a;
  transition: background-color 0.3s ease;
}

.header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 20px 40px;
  background: linear-gradient(135deg, #1a1a2e 0%, #16213e 100%);
  border-bottom: 3px solid #333;
  flex-shrink: 0;

  .header-left {
    display: flex;
    align-items: center;
    gap: 24px;

    .title {
      font-size: 42px;
      font-weight: 700;
      color: #fff;
      display: flex;
      align-items: center;
      gap: 12px;

      .icon {
        font-size: 48px;
      }
    }
  }

  .header-center {
    flex: 1;
    display: flex;
    justify-content: center;
  }

  .stats {
    display: flex;
    gap: 40px;

    .stat-item {
      display: flex;
      flex-direction: column;
      align-items: center;
      padding: 12px 32px;
      background: rgba(255, 255, 255, 0.05);
      border-radius: 12px;
      min-width: 140px;

      .stat-label {
        font-size: 20px;
        color: #999;
        margin-bottom: 4px;
      }

      .stat-value {
        font-size: 56px;
        font-weight: 700;
        line-height: 1;
      }

      &.pending .stat-value { color: #e6a23c; }
      &.cooking .stat-value { color: #409eff; }
      &.completed .stat-value { color: #67c23a; }
      &.overdue .stat-value { color: #f56c6c; animation: pulse 1s ease-in-out infinite; }
    }
  }

  .header-right {
    display: flex;
    align-items: center;
    gap: 24px;

    .current-time {
      font-size: 32px;
      font-weight: 600;
      color: #fff;
      font-family: 'Courier New', monospace;
      letter-spacing: 2px;
    }
  }
}

.main-content {
  flex: 1;
  display: flex;
  overflow: hidden;
  padding: 24px;
  gap: 24px;

  .column {
    flex: 1;
    display: flex;
    flex-direction: column;
    background: #111;
    border-radius: 16px;
    border: 2px solid #222;
    overflow: hidden;

    .column-header {
      display: flex;
      align-items: center;
      justify-content: space-between;
      padding: 20px 28px;
      background: #1a1a1a;
      border-bottom: 2px solid #333;
      flex-shrink: 0;

      h2 {
        font-size: 36px;
        font-weight: 700;
        color: #fff;
        margin: 0;
      }

      .count-badge {
        font-size: 32px;
        font-weight: 700;
        padding: 4px 20px;
        background: #e6a23c;
        color: #fff;
        border-radius: 24px;
        min-width: 60px;
        text-align: center;

        &.cooking {
          background: #409eff;
        }
      }
    }

    .items-list {
      flex: 1;
      overflow-y: auto;
      padding: 20px;
      display: flex;
      flex-direction: column;
      gap: 20px;

      .item-card {
        background: #1a1a1a;
        border: 3px solid #444;
        border-radius: 16px;
        padding: 24px;
        transition: all 0.3s ease;
        flex-shrink: 0;

        &:hover {
          border-color: #666;
          transform: translateY(-2px);
        }

        &.cooking-card {
          border-color: #409eff;
          background: linear-gradient(135deg, #1a2a3a 0%, #1a1a2a 100%);
        }

        .item-header {
          display: flex;
          justify-content: space-between;
          align-items: center;
          margin-bottom: 16px;

          .item-order-no {
            font-size: 20px;
            color: #666;
            font-family: 'Courier New', monospace;
          }
        }

        .item-body {
          margin-bottom: 20px;

          .item-name {
            font-size: 42px;
            font-weight: 700;
            color: #fff;
            line-height: 1.2;
            margin-bottom: 8px;
            word-break: break-all;

            &.overdue-text {
              color: #ff4444;
            }
          }

          .item-sku,
          .item-spec {
            font-size: 24px;
            color: #aaa;
            margin-bottom: 4px;
          }

          .item-remark {
            font-size: 22px;
            color: #e6a23c;
            margin-top: 8px;
            padding: 8px 12px;
            background: rgba(230, 162, 60, 0.1);
            border-radius: 8px;
            border-left: 4px solid #e6a23c;
          }
        }

        .item-footer {
          .item-meta {
            display: flex;
            justify-content: space-between;
            align-items: center;
            margin-bottom: 16px;

            .item-quantity {
              font-size: 36px;
              font-weight: 700;
              color: #409eff;
            }

            .item-time {
              font-size: 28px;
              font-weight: 600;
              color: #999;
              font-family: 'Courier New', monospace;

              &.overdue-text {
                color: #ff4444;
              }
            }
          }

          .item-actions {
            display: flex;
            gap: 12px;
          }
        }

        &.overdue-card {
          border-color: #ff0000;
          background: linear-gradient(135deg, #2a0a0a 0%, #1a0a0a 100%);
        }
      }

      .empty-state {
        flex: 1;
        display: flex;
        flex-direction: column;
        align-items: center;
        justify-content: center;
        color: #555;

        .empty-icon {
          font-size: 80px;
          margin-bottom: 16px;
        }

        .empty-text {
          font-size: 28px;
        }
      }
    }
  }
}

@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.6; }
}

:deep(.el-dialog) {
  background: #1a1a1a !important;
  border: 2px solid #333 !important;

  .el-dialog__header {
    border-bottom: 1px solid #333;
  }

  .el-dialog__title {
    color: #fff !important;
    font-size: 28px !important;
  }

  .el-dialog__body {
    color: #fff !important;
  }

  .el-form-item__label {
    color: #ccc !important;
    font-size: 20px !important;
  }

  .el-input__wrapper,
  .el-textarea__inner,
  .el-input-number {
    background: #2a2a2a !important;
    border-color: #444 !important;
    color: #fff !important;
    font-size: 20px !important;
  }

  .el-dialog__footer {
    border-top: 1px solid #333;
  }
}
</style>
