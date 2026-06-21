<template>
  <div class="workorders-page">
    <el-tabs v-model="activeTab" class="main-tabs">
      <el-tab-pane label="差评工单" name="workorder">
        <div class="page-header">
          <h2 class="page-title">差评工单管理</h2>
        </div>

        <div class="card-wrapper">
          <div class="search-bar">
            <el-select
              v-model="query.store_id"
              placeholder="选择门店"
              clearable
              style="width: 160px">
              <el-option
                v-for="store in storeList"
                :key="store.id"
                :label="store.name"
                :value="store.id" />
            </el-select>
            <el-select v-model="query.status" placeholder="工单状态" clearable style="width: 140px">
              <el-option label="待处理" :value="1" />
              <el-option label="处理中" :value="2" />
              <el-option label="已完成" :value="3" />
              <el-option label="已取消" :value="4" />
            </el-select>
            <el-select v-model="query.priority" placeholder="优先级" clearable style="width: 120px">
              <el-option label="低" :value="1" />
              <el-option label="中" :value="2" />
              <el-option label="高" :value="3" />
              <el-option label="紧急" :value="4" />
            </el-select>
            <el-select
              v-model="query.assignee_id"
              placeholder="指派店长"
              clearable
              filterable
              style="width: 160px">
              <el-option
                v-for="manager in managerList"
                :key="manager.id"
                :label="manager.name"
                :value="manager.id" />
            </el-select>
            <el-button type="primary" @click="fetchWorkOrderList">
              <el-icon><Search /></el-icon>搜索
            </el-button>
            <el-button @click="resetWorkOrderQuery">
              <el-icon><Refresh /></el-icon>重置
            </el-button>
          </div>

          <el-table :data="workOrderList" v-loading="workOrderLoading">
            <el-table-column prop="order_no" label="工单号" width="140" />
            <el-table-column prop="store_name" label="门店" width="140" />
            <el-table-column prop="title" label="标题" min-width="180" show-overflow-tooltip />
            <el-table-column label="优先级" width="100" align="center">
              <template #default="{ row }">
                <el-tag :type="getPriorityTagType(row.priority)" size="small">
                  {{ getPriorityName(row.priority) }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column label="状态" width="100" align="center">
              <template #default="{ row }">
                <el-tag :type="getWorkOrderStatusTagType(row.status)" effect="light">
                  {{ getWorkOrderStatusName(row.status) }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="assignee_name" label="指派人" width="100" />
            <el-table-column prop="created_at" label="创建时间" width="170" />
            <el-table-column prop="deadline" label="截止时间" width="170" />
            <el-table-column label="操作" width="180" fixed="right">
              <template #default="{ row }">
                <el-button type="primary" link size="small" @click="handleViewWorkOrder(row)">查看详情</el-button>
                <el-button
                  v-if="row.status === 1 || row.status === 2"
                  type="success"
                  link
                  size="small"
                  @click="handleProcessWorkOrder(row)">
                  处理
                </el-button>
              </template>
            </el-table-column>
          </el-table>

          <div class="pagination">
            <el-pagination
              v-model:current-page="query.page"
              v-model:page-size="query.page_size"
              :total="workOrderTotal"
              :page-sizes="[10, 20, 50, 100]"
              layout="total, sizes, prev, pager, next, jumper"
              @size-change="fetchWorkOrderList"
              @current-change="fetchWorkOrderList" />
          </div>
        </div>

        <el-dialog
          v-model="workOrderDetailVisible"
          title="工单详情"
          width="750px"
          :close-on-click-modal="false">
          <div v-if="currentWorkOrder" class="workorder-detail">
            <el-descriptions :column="2" border>
              <el-descriptions-item label="工单号">{{ currentWorkOrder.order_no }}</el-descriptions-item>
              <el-descriptions-item label="门店">{{ currentWorkOrder.store_name }}</el-descriptions-item>
              <el-descriptions-item label="标题">{{ currentWorkOrder.title }}</el-descriptions-item>
              <el-descriptions-item label="优先级">
                <el-tag :type="getPriorityTagType(currentWorkOrder.priority)" size="small">
                  {{ getPriorityName(currentWorkOrder.priority) }}
                </el-tag>
              </el-descriptions-item>
              <el-descriptions-item label="状态">
                <el-tag :type="getWorkOrderStatusTagType(currentWorkOrder.status)" effect="light">
                  {{ getWorkOrderStatusName(currentWorkOrder.status) }}
                </el-tag>
              </el-descriptions-item>
              <el-descriptions-item label="指派人">{{ currentWorkOrder.assignee_name || '-' }}</el-descriptions-item>
              <el-descriptions-item label="创建时间">{{ currentWorkOrder.created_at }}</el-descriptions-item>
              <el-descriptions-item label="截止时间">{{ currentWorkOrder.deadline || '-' }}</el-descriptions-item>
            </el-descriptions>

            <div v-if="currentWorkOrder.review" class="detail-section">
              <h4 class="section-title">关联差评内容</h4>
              <div class="review-content">
                <div class="review-header">
                  <span class="review-user">{{ currentWorkOrder.review.user_nickname }}</span>
                  <el-rate v-model="currentWorkOrder.review.rating" disabled show-score text-color="#ff9900" size="small" />
                </div>
                <p class="review-text">{{ currentWorkOrder.review.content }}</p>
                <div v-if="currentWorkOrder.review.images && currentWorkOrder.review.images.length" class="image-list">
                  <el-image
                    v-for="(img, idx) in currentWorkOrder.review.images"
                    :key="idx"
                    :src="img"
                    :preview-src-list="currentWorkOrder.review.images"
                    fit="cover"
                    class="review-image" />
                </div>
              </div>
            </div>

            <div v-if="currentWorkOrder.remark" class="detail-section">
              <h4 class="section-title">工单备注</h4>
              <p class="content-text">{{ currentWorkOrder.remark }}</p>
            </div>

            <div v-if="currentWorkOrder.handle_result" class="detail-section">
              <h4 class="section-title">处理结果</h4>
              <p class="content-text">{{ currentWorkOrder.handle_result }}</p>
              <p class="handle-time">处理时间：{{ currentWorkOrder.handled_at }}</p>
              <p class="handle-time">处理人：{{ currentWorkOrder.handler_name }}</p>
            </div>

            <div v-if="(currentWorkOrder.status === 1 || currentWorkOrder.status === 2) && !currentWorkOrder.handle_result" class="detail-section">
              <h4 class="section-title">处理工单</h4>
              <el-form ref="handleFormRef" :model="handleForm" :rules="handleRules" label-width="100px">
                <el-form-item label="处理结果" prop="handle_result">
                  <el-input
                    v-model="handleForm.handle_result"
                    type="textarea"
                    :rows="4"
                    placeholder="请输入处理结果" />
                </el-form-item>
                <el-form-item label="状态变更" prop="status">
                  <el-radio-group v-model="handleForm.status">
                    <el-radio :value="2">处理中</el-radio>
                    <el-radio :value="3">已完成</el-radio>
                    <el-radio :value="4">已取消</el-radio>
                  </el-radio-group>
                </el-form-item>
              </el-form>
            </div>
          </div>
          <template #footer>
            <el-button @click="workOrderDetailVisible = false">关闭</el-button>
            <el-button
              v-if="currentWorkOrder && (currentWorkOrder.status === 1 || currentWorkOrder.status === 2) && !currentWorkOrder.handle_result"
              type="primary"
              :loading="submitLoading"
              @click="handleSubmitWorkOrder">
              提交处理
            </el-button>
          </template>
        </el-dialog>
      </el-tab-pane>

      <el-tab-pane label="评分告警" name="alert">
        <div class="page-header">
          <h2 class="page-title">评分告警管理</h2>
        </div>

        <div class="card-wrapper">
          <div class="search-bar">
            <el-select
              v-model="alertQuery.store_id"
              placeholder="选择门店"
              clearable
              style="width: 160px">
              <el-option
                v-for="store in storeList"
                :key="store.id"
                :label="store.name"
                :value="store.id" />
            </el-select>
            <el-select v-model="alertQuery.status" placeholder="告警状态" clearable style="width: 140px">
              <el-option label="未处理" :value="0" />
              <el-option label="已处理" :value="1" />
            </el-select>
            <el-select v-model="alertQuery.alert_type" placeholder="告警类型" clearable style="width: 160px">
              <el-option label="周评分下降" value="weekly_drop" />
              <el-option label="月评分下降" value="monthly_drop" />
              <el-option label="差评激增" value="bad_surge" />
              <el-option label="评分低于阈值" value="below_threshold" />
            </el-select>
            <el-date-picker
              v-model="alertQuery.time_range"
              type="daterange"
              range-separator="至"
              start-placeholder="开始日期"
              end-placeholder="结束日期"
              value-format="YYYY-MM-DD"
              style="width: 280px" />
            <el-button type="primary" @click="fetchAlertList">
              <el-icon><Search /></el-icon>搜索
            </el-button>
            <el-button @click="resetAlertQuery">
              <el-icon><Refresh /></el-icon>重置
            </el-button>
          </div>

          <el-table :data="alertList" v-loading="alertLoading">
            <el-table-column prop="id" label="ID" width="80" />
            <el-table-column prop="store_name" label="门店" width="140" />
            <el-table-column label="平台" width="100" align="center">
              <template #default="{ row }">
                <el-tag size="small" :type="getPlatformTagType(row.platform)">
                  {{ getPlatformName(row.platform) }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column label="告警类型" width="130" align="center">
              <template #default="{ row }">
                <el-tag :type="getAlertTypeTagType(row.alert_type)" size="small">
                  {{ getAlertTypeName(row.alert_type) }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="title" label="标题" min-width="180" show-overflow-tooltip />
            <el-table-column label="评分变化" width="160" align="center">
              <template #default="{ row }">
                <div class="rating-change">
                  <span class="rating-before">{{ row.rating_before }}</span>
                  <el-icon class="arrow-icon"><ArrowRight /></el-icon>
                  <span class="rating-after">{{ row.rating_after }}</span>
                  <span v-if="row.rating_drop" class="rating-drop">-{{ row.rating_drop }}</span>
                </div>
              </template>
            </el-table-column>
            <el-table-column label="状态" width="100" align="center">
              <template #default="{ row }">
                <el-tag :type="row.status === 1 ? 'success' : 'warning'" effect="light" size="small">
                  {{ row.status === 1 ? '已处理' : '未处理' }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="created_at" label="告警时间" width="170" />
            <el-table-column label="操作" width="140" fixed="right">
              <template #default="{ row }">
                <el-button
                  v-if="row.status === 0"
                  type="primary"
                  link
                  size="small"
                  @click="handleProcessAlert(row)">
                  处理
                </el-button>
                <el-button type="info" link size="small" @click="handleViewAlert(row)">查看</el-button>
              </template>
            </el-table-column>
          </el-table>

          <div class="pagination">
            <el-pagination
              v-model:current-page="alertQuery.page"
              v-model:page-size="alertQuery.page_size"
              :total="alertTotal"
              :page-sizes="[10, 20, 50, 100]"
              layout="total, sizes, prev, pager, next, jumper"
              @size-change="fetchAlertList"
              @current-change="fetchAlertList" />
          </div>
        </div>

        <el-dialog
          v-model="alertVisible"
          :title="isProcessAlert ? '处理告警' : '告警详情'"
          width="600px"
          :close-on-click-modal="false">
          <div v-if="currentAlert" class="alert-detail">
            <el-descriptions :column="2" border>
              <el-descriptions-item label="ID">{{ currentAlert.id }}</el-descriptions-item>
              <el-descriptions-item label="门店">{{ currentAlert.store_name }}</el-descriptions-item>
              <el-descriptions-item label="平台">
                <el-tag size="small" :type="getPlatformTagType(currentAlert.platform)">
                  {{ getPlatformName(currentAlert.platform) }}
                </el-tag>
              </el-descriptions-item>
              <el-descriptions-item label="告警类型">
                <el-tag :type="getAlertTypeTagType(currentAlert.alert_type)" size="small">
                  {{ getAlertTypeName(currentAlert.alert_type) }}
                </el-tag>
              </el-descriptions-item>
              <el-descriptions-item label="标题" :span="2">{{ currentAlert.title }}</el-descriptions-item>
              <el-descriptions-item label="评分变化" :span="2">
                <div class="rating-change large">
                  <span class="rating-before">{{ currentAlert.rating_before }}</span>
                  <el-icon class="arrow-icon"><ArrowRight /></el-icon>
                  <span class="rating-after">{{ currentAlert.rating_after }}</span>
                  <span v-if="currentAlert.rating_drop" class="rating-drop">-{{ currentAlert.rating_drop }}</span>
                </div>
              </el-descriptions-item>
              <el-descriptions-item label="状态">
                <el-tag :type="currentAlert.status === 1 ? 'success' : 'warning'" effect="light" size="small">
                  {{ currentAlert.status === 1 ? '已处理' : '未处理' }}
                </el-tag>
              </el-descriptions-item>
              <el-descriptions-item label="告警时间">{{ currentAlert.created_at }}</el-descriptions-item>
            </el-descriptions>

            <div v-if="currentAlert.content" class="detail-section">
              <h4 class="section-title">告警详情</h4>
              <p class="content-text">{{ currentAlert.content }}</p>
            </div>

            <div v-if="currentAlert.handle_remark" class="detail-section">
              <h4 class="section-title">处理备注</h4>
              <p class="content-text">{{ currentAlert.handle_remark }}</p>
              <p class="handle-time">处理时间：{{ currentAlert.handled_at }}</p>
            </div>

            <div v-if="isProcessAlert && currentAlert.status === 0" class="detail-section">
              <h4 class="section-title">处理告警</h4>
              <el-form ref="alertFormRef" :model="alertForm" :rules="alertRules" label-width="100px">
                <el-form-item label="处理备注" prop="handle_remark">
                  <el-input
                    v-model="alertForm.handle_remark"
                    type="textarea"
                    :rows="4"
                    placeholder="请输入处理备注" />
                </el-form-item>
              </el-form>
            </div>
          </div>
          <template #footer>
            <el-button @click="alertVisible = false">关闭</el-button>
            <el-button
              v-if="isProcessAlert && currentAlert && currentAlert.status === 0"
              type="primary"
              :loading="submitLoading"
              @click="handleSubmitAlert">
              确认处理
            </el-button>
          </template>
        </el-dialog>
      </el-tab-pane>
    </el-tabs>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { Search, Refresh, ArrowRight } from '@element-plus/icons-vue'
import {
  getWorkOrderList,
  getWorkOrderDetail,
  handleWorkOrder,
  getAlertList,
  handleAlert
} from '@/api/review'
import { getStoreList } from '@/api/store'

const activeTab = ref('workorder')
const submitLoading = ref(false)
const storeList = ref([])
const managerList = ref([
  { id: 1, name: '张三' },
  { id: 2, name: '李四' },
  { id: 3, name: '王五' }
])

const workOrderLoading = ref(false)
const workOrderList = ref([])
const workOrderTotal = ref(0)

const query = reactive({
  store_id: null,
  status: null,
  priority: null,
  assignee_id: null,
  page: 1,
  page_size: 10
})

const workOrderDetailVisible = ref(false)
const currentWorkOrder = ref(null)
const handleFormRef = ref()

const handleForm = reactive({
  handle_result: '',
  status: 3
})

const handleRules = {
  handle_result: [{ required: true, message: '请输入处理结果', trigger: 'blur' }],
  status: [{ required: true, message: '请选择状态', trigger: 'change' }]
}

const priorityMap = {
  1: { name: '低', type: 'info' },
  2: { name: '中', type: 'primary' },
  3: { name: '高', type: 'warning' },
  4: { name: '紧急', type: 'danger' }
}

const workOrderStatusMap = {
  1: { name: '待处理', type: 'warning' },
  2: { name: '处理中', type: 'primary' },
  3: { name: '已完成', type: 'success' },
  4: { name: '已取消', type: 'info' }
}

const platformMap = {
  meituan: { name: '美团', type: 'primary' },
  eleme: { name: '饿了么', type: 'success' },
  dianping: { name: '大众点评', type: 'warning' },
  douyin: { name: '抖音', type: 'danger' },
  miniprogram: { name: '小程序', type: 'info' }
}

const alertTypeMap = {
  weekly_drop: { name: '周评分下降', type: 'warning' },
  monthly_drop: { name: '月评分下降', type: 'danger' },
  bad_surge: { name: '差评激增', type: 'danger' },
  below_threshold: { name: '评分低于阈值', type: 'warning' }
}

function getPriorityName(priority) {
  return priorityMap[priority]?.name || '未知'
}

function getPriorityTagType(priority) {
  return priorityMap[priority]?.type || 'info'
}

function getWorkOrderStatusName(status) {
  return workOrderStatusMap[status]?.name || '未知'
}

function getWorkOrderStatusTagType(status) {
  return workOrderStatusMap[status]?.type || 'info'
}

function getPlatformName(platform) {
  return platformMap[platform]?.name || '未知'
}

function getPlatformTagType(platform) {
  return platformMap[platform]?.type || 'info'
}

function getAlertTypeName(type) {
  return alertTypeMap[type]?.name || '未知'
}

function getAlertTypeTagType(type) {
  return alertTypeMap[type]?.type || 'info'
}

async function fetchStoreList() {
  try {
    const res = await getStoreList({ page: 1, page_size: 100 })
    storeList.value = res.list || []
  } catch (e) {
    console.error(e)
  }
}

async function fetchWorkOrderList() {
  workOrderLoading.value = true
  try {
    const res = await getWorkOrderList(query)
    workOrderList.value = res.list || []
    workOrderTotal.value = res.total || 0
  } catch (e) {
    console.error(e)
  } finally {
    workOrderLoading.value = false
  }
}

function resetWorkOrderQuery() {
  query.store_id = null
  query.status = null
  query.priority = null
  query.assignee_id = null
  query.page = 1
  fetchWorkOrderList()
}

async function handleViewWorkOrder(row) {
  try {
    const res = await getWorkOrderDetail(row.id)
    currentWorkOrder.value = res
    handleForm.handle_result = ''
    handleForm.status = 3
    workOrderDetailVisible.value = true
  } catch (e) {
    console.error(e)
  }
}

function handleProcessWorkOrder(row) {
  handleViewWorkOrder(row)
}

async function handleSubmitWorkOrder() {
  try {
    await handleFormRef.value.validate()
    submitLoading.value = true
    await handleWorkOrder(currentWorkOrder.value.id, {
      handle_result: handleForm.handle_result,
      status: handleForm.status
    })
    ElMessage.success('处理成功')
    workOrderDetailVisible.value = false
    fetchWorkOrderList()
  } catch (e) {
    console.error(e)
  } finally {
    submitLoading.value = false
  }
}

const alertLoading = ref(false)
const alertList = ref([])
const alertTotal = ref(0)
const alertVisible = ref(false)
const currentAlert = ref(null)
const isProcessAlert = ref(false)
const alertFormRef = ref()

const alertQuery = reactive({
  store_id: null,
  status: null,
  alert_type: '',
  time_range: [],
  page: 1,
  page_size: 10
})

const alertForm = reactive({
  handle_remark: ''
})

const alertRules = {
  handle_remark: [{ required: true, message: '请输入处理备注', trigger: 'blur' }]
}

async function fetchAlertList() {
  alertLoading.value = true
  try {
    const params = { ...alertQuery }
    if (params.time_range && params.time_range.length === 2) {
      params.start_time = params.time_range[0]
      params.end_time = params.time_range[1]
    }
    delete params.time_range
    const res = await getAlertList(params)
    alertList.value = res.list || []
    alertTotal.value = res.total || 0
  } catch (e) {
    console.error(e)
  } finally {
    alertLoading.value = false
  }
}

function resetAlertQuery() {
  alertQuery.store_id = null
  alertQuery.status = null
  alertQuery.alert_type = ''
  alertQuery.time_range = []
  alertQuery.page = 1
  fetchAlertList()
}

function handleViewAlert(row) {
  currentAlert.value = row
  isProcessAlert.value = false
  alertForm.handle_remark = ''
  alertVisible.value = true
}

function handleProcessAlert(row) {
  currentAlert.value = row
  isProcessAlert.value = true
  alertForm.handle_remark = ''
  alertVisible.value = true
}

async function handleSubmitAlert() {
  try {
    await alertFormRef.value.validate()
    submitLoading.value = true
    await handleAlert(currentAlert.value.id, {
      handle_remark: alertForm.handle_remark
    })
    ElMessage.success('处理成功')
    alertVisible.value = false
    fetchAlertList()
  } catch (e) {
    console.error(e)
  } finally {
    submitLoading.value = false
  }
}

onMounted(() => {
  fetchStoreList()
  fetchWorkOrderList()
})
</script>

<style scoped lang="scss">
.workorders-page {
  .main-tabs {
    :deep(.el-tabs__header) {
      margin-bottom: 0;
    }
  }

  .search-bar {
    display: flex;
    flex-wrap: wrap;
    gap: 10px;
    margin-bottom: 20px;
    align-items: center;
  }

  .pagination {
    margin-top: 20px;
    display: flex;
    justify-content: flex-end;
  }

  .rating-change {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 6px;
    font-size: 14px;

    &.large {
      justify-content: flex-start;
      font-size: 16px;
    }

    .rating-before {
      color: #67c23a;
      font-weight: 600;
    }

    .rating-after {
      color: #f56c6c;
      font-weight: 600;
    }

    .arrow-icon {
      color: #909399;
    }

    .rating-drop {
      color: #f56c6c;
      font-weight: 700;
      background: #fef0f0;
      padding: 2px 8px;
      border-radius: 4px;
      font-size: 12px;
    }
  }

  .workorder-detail,
  .alert-detail {
    .detail-section {
      margin-top: 20px;

      .section-title {
        margin: 0 0 12px 0;
        font-size: 14px;
        font-weight: 600;
        color: #303133;
      }
    }

    .content-text {
      margin: 0;
      padding: 12px;
      background: #f5f7fa;
      border-radius: 4px;
      line-height: 1.6;
      color: #303133;
    }

    .handle-time {
      margin: 8px 0 0 0;
      font-size: 12px;
      color: #909399;
    }

    .review-content {
      padding: 12px;
      background: #fef0f0;
      border-radius: 4px;

      .review-header {
        display: flex;
        align-items: center;
        justify-content: space-between;
        margin-bottom: 8px;

        .review-user {
          font-weight: 600;
          color: #303133;
        }
      }

      .review-text {
        margin: 0 0 12px 0;
        line-height: 1.6;
        color: #303133;
      }

      .image-list {
        display: flex;
        gap: 10px;
        flex-wrap: wrap;
      }

      .review-image {
        width: 80px;
        height: 80px;
        border-radius: 4px;
        cursor: pointer;
      }
    }
  }
}
</style>
