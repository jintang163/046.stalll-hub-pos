<template>
  <div class="sms-tasks-page">
    <div class="page-header">
      <h2 class="page-title">短信任务管理</h2>
      <div class="header-actions">
        <el-button type="success" @click="openDialog">
          <el-icon><Plus /></el-icon>新建任务
        </el-button>
      </div>
    </div>

    <div class="card-wrapper">
      <div class="search-bar">
        <el-input
          v-model="query.name"
          placeholder="搜索任务名称"
          clearable
          style="width: 240px"
          @keyup.enter="fetchList" />
        <el-select v-model="query.status" placeholder="任务状态" clearable style="width: 140px">
          <el-option label="全部" value="" />
          <el-option label="草稿" :value="0" />
          <el-option label="待发送" :value="1" />
          <el-option label="发送中" :value="2" />
          <el-option label="已完成" :value="3" />
          <el-option label="已暂停" :value="4" />
          <el-option label="已取消" :value="5" />
        </el-select>
        <el-select v-model="query.schedule_type" placeholder="调度类型" clearable style="width: 140px">
          <el-option label="全部" value="" />
          <el-option label="立即发送" :value="1" />
          <el-option label="定时发送" :value="2" />
        </el-select>
        <el-button type="primary" @click="fetchList">
          <el-icon><Search /></el-icon>搜索
        </el-button>
        <el-button @click="resetQuery">
          <el-icon><Refresh /></el-icon>重置
        </el-button>
      </div>

      <el-table :data="list" v-loading="loading">
        <el-table-column prop="id" label="任务ID" width="90" />
        <el-table-column prop="name" label="任务名称" min-width="180" />
        <el-table-column label="任务类型" width="120">
          <template #default="{ row }">
            <el-tag :type="getTaskTypeTagType(row.task_type)">
              {{ getTaskTypeName(row.task_type) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="template_name" label="使用模板" min-width="160" />
        <el-table-column prop="target_count" label="目标人数" width="100" align="center" />
        <el-table-column label="成功/失败/总数" width="160" align="center">
          <template #default="{ row }">
            <div class="stat-text">
              <span class="success">{{ row.success_count || 0 }}</span>
              <span class="divider">/</span>
              <span class="fail">{{ row.fail_count || 0 }}</span>
              <span class="divider">/</span>
              <span>{{ row.total_count || 0 }}</span>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="成功率" width="100" align="center">
          <template #default="{ row }">
            <span :class="['success-rate', getSuccessRateClass(row.success_rate)]">
              {{ formatSuccessRate(row.success_rate) }}
            </span>
          </template>
        </el-table-column>
        <el-table-column label="调度类型" width="100" align="center">
          <template #default="{ row }">
            <el-tag size="small" :type="row.schedule_type === 1 ? 'primary' : 'warning'">
              {{ row.schedule_type === 1 ? '立即发送' : '定时发送' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="schedule_time" label="定时发送时间" width="160" />
        <el-table-column label="状态" width="100" align="center">
          <template #default="{ row }">
            <el-tag :type="getStatusTagType(row.status)" effect="light">
              {{ getStatusName(row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="创建时间" width="160" />
        <el-table-column label="操作" width="260" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="handleViewDetail(row)">详情</el-button>
            <el-button
              v-if="row.status === 0 || row.status === 4"
              type="success"
              link
              size="small"
              @click="handleStart(row)">
              启动
            </el-button>
            <el-button
              v-if="row.status === 2"
              type="warning"
              link
              size="small"
              @click="handlePause(row)">
              暂停
            </el-button>
            <el-button
              v-if="row.status === 0"
              type="primary"
              link
              size="small"
              @click="handleEdit(row)">
              编辑
            </el-button>
            <el-button
              v-if="row.status === 0 || row.status === 3 || row.status === 5"
              type="danger"
              link
              size="small"
              @click="handleDelete(row)">
              删除
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <div class="pagination">
        <el-pagination
          v-model:current-page="query.page"
          v-model:page-size="query.page_size"
          :total="total"
          :page-sizes="[10, 20, 50, 100]"
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="fetchList"
          @current-change="fetchList" />
      </div>
    </div>

    <el-dialog
      v-model="dialogVisible"
      :title="isEdit ? '编辑任务' : '新建任务'"
      width="780px"
      :close-on-click-modal="false"
      class="task-dialog">
      <el-form
        ref="taskFormRef"
        :model="taskForm"
        :rules="taskRules"
        label-width="120px">

        <div class="form-section">
          <h4 class="section-title">基本信息</h4>
          <el-form-item label="任务名称" prop="name">
            <el-input v-model="taskForm.name" placeholder="请输入任务名称" maxlength="50" show-word-limit />
          </el-form-item>
          <el-row :gutter="20">
            <el-col :span="12">
              <el-form-item label="任务类型" prop="task_type">
                <el-select v-model="taskForm.task_type" placeholder="请选择任务类型" style="width: 100%">
                  <el-option label="营销短信" :value="1" />
                  <el-option label="通知短信" :value="2" />
                  <el-option label="会员关怀" :value="3" />
                  <el-option label="其他" :value="99" />
                </el-select>
              </el-form-item>
            </el-col>
          </el-row>
          <el-form-item label="备注">
            <el-input
              v-model="taskForm.remark"
              type="textarea"
              :rows="2"
              placeholder="请输入备注信息"
              maxlength="200"
              show-word-limit />
          </el-form-item>
        </div>

        <div class="form-section">
          <h4 class="section-title">模板选择</h4>
          <el-form-item label="选择模板" prop="template_id">
            <el-select
              v-model="taskForm.template_id"
              placeholder="请选择已审核通过的模板"
              style="width: 100%"
              @change="handleTemplateChange">
              <el-option
                v-for="tpl in templateList"
                :key="tpl.id"
                :label="tpl.name"
                :value="tpl.id" />
            </el-select>
          </el-form-item>
          <el-form-item label="模板内容">
            <div class="template-preview">
              {{ selectedTemplate?.content || '请选择模板' }}
            </div>
          </el-form-item>
        </div>

        <div class="form-section">
          <h4 class="section-title">
            目标人群筛选
            <span class="target-count">预估目标人数：<b>{{ targetCount }}</b> 人</span>
          </h4>
          <el-form-item label="会员等级">
            <el-select
              v-model="taskForm.filter_levels"
              multiple
              placeholder="请选择会员等级（不选则为全部）"
              style="width: 100%"
              @change="handleFilterChange">
              <el-option
                v-for="level in memberLevelList"
                :key="level.value"
                :label="level.label"
                :value="level.value" />
            </el-select>
          </el-form-item>
          <el-row :gutter="20">
            <el-col :span="12">
              <el-form-item label="消费次数">
                <div class="range-input">
                  <el-input-number
                    v-model="taskForm.filter_consume_count_min"
                    :min="0"
                    :controls="false"
                    placeholder="最小值"
                    style="width: 48%"
                    @change="handleFilterChange" />
                  <span class="range-separator">~</span>
                  <el-input-number
                    v-model="taskForm.filter_consume_count_max"
                    :min="0"
                    :controls="false"
                    placeholder="最大值"
                    style="width: 48%"
                    @change="handleFilterChange" />
                </div>
              </el-form-item>
            </el-col>
            <el-col :span="12">
              <el-form-item label="消费金额(元)">
                <div class="range-input">
                  <el-input-number
                    v-model="taskForm.filter_consume_amount_min"
                    :min="0"
                    :precision="2"
                    :controls="false"
                    placeholder="最小值"
                    style="width: 48%"
                    @change="handleFilterChange" />
                  <span class="range-separator">~</span>
                  <el-input-number
                    v-model="taskForm.filter_consume_amount_max"
                    :min="0"
                    :precision="2"
                    :controls="false"
                    placeholder="最大值"
                    style="width: 48%"
                    @change="handleFilterChange" />
                </div>
              </el-form-item>
            </el-col>
          </el-row>
          <el-row :gutter="20">
            <el-col :span="12">
              <el-form-item label="积分">
                <div class="range-input">
                  <el-input-number
                    v-model="taskForm.filter_points_min"
                    :min="0"
                    :controls="false"
                    placeholder="最小值"
                    style="width: 48%"
                    @change="handleFilterChange" />
                  <span class="range-separator">~</span>
                  <el-input-number
                    v-model="taskForm.filter_points_max"
                    :min="0"
                    :controls="false"
                    placeholder="最大值"
                    style="width: 48%"
                    @change="handleFilterChange" />
                </div>
              </el-form-item>
            </el-col>
          </el-row>
        </div>

        <div class="form-section">
          <h4 class="section-title">发送设置</h4>
          <el-form-item label="发送类型" prop="schedule_type">
            <el-radio-group v-model="taskForm.schedule_type">
              <el-radio :value="1">立即发送</el-radio>
              <el-radio :value="2">定时发送</el-radio>
            </el-radio-group>
          </el-form-item>
          <el-form-item v-if="taskForm.schedule_type === 2" label="定时发送时间" prop="schedule_time">
            <el-date-picker
              v-model="taskForm.schedule_time"
              type="datetime"
              placeholder="选择发送时间"
              value-format="YYYY-MM-DD HH:mm:ss"
              style="width: 100%" />
          </el-form-item>
        </div>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button :loading="submitLoading" @click="handleSaveDraft">保存草稿</el-button>
        <el-button type="primary" :loading="submitLoading" @click="handleSubmit">
          {{ isEdit ? '保存并发送' : '立即发送' }}
        </el-button>
      </template>
    </el-dialog>

    <el-dialog
      v-model="detailVisible"
      title="任务详情"
      width="700px"
      :close-on-click-modal="false">
      <div v-if="currentDetail" class="task-detail">
        <el-descriptions :column="2" border>
          <el-descriptions-item label="任务ID">
            {{ currentDetail.id }}
          </el-descriptions-item>
          <el-descriptions-item label="任务名称">
            {{ currentDetail.name }}
          </el-descriptions-item>
          <el-descriptions-item label="任务类型">
            <el-tag :type="getTaskTypeTagType(currentDetail.task_type)" size="small">
              {{ getTaskTypeName(currentDetail.task_type) }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="使用模板">
            {{ currentDetail.template_name || '-' }}
          </el-descriptions-item>
          <el-descriptions-item label="调度类型">
            {{ currentDetail.schedule_type === 1 ? '立即发送' : '定时发送' }}
          </el-descriptions-item>
          <el-descriptions-item label="定时发送时间">
            {{ currentDetail.schedule_time || '-' }}
          </el-descriptions-item>
          <el-descriptions-item label="状态">
            <el-tag :type="getStatusTagType(currentDetail.status)" effect="light">
              {{ getStatusName(currentDetail.status) }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="创建时间">
            {{ currentDetail.created_at }}
          </el-descriptions-item>
        </el-descriptions>

        <div class="detail-section">
          <h4 class="section-title">发送统计</h4>
          <div class="stat-cards">
            <div class="stat-card">
              <div class="stat-label">目标人数</div>
              <div class="stat-value">{{ currentDetail.target_count || 0 }}</div>
            </div>
            <div class="stat-card">
              <div class="stat-label">发送总数</div>
              <div class="stat-value">{{ currentDetail.total_count || 0 }}</div>
            </div>
            <div class="stat-card success">
              <div class="stat-label">成功</div>
              <div class="stat-value">{{ currentDetail.success_count || 0 }}</div>
            </div>
            <div class="stat-card fail">
              <div class="stat-label">失败</div>
              <div class="stat-value">{{ currentDetail.fail_count || 0 }}</div>
            </div>
            <div class="stat-card">
              <div class="stat-label">成功率</div>
              <div class="stat-value">{{ formatSuccessRate(currentDetail.success_rate) }}</div>
            </div>
          </div>
        </div>

        <div v-if="currentDetail.remark" class="detail-section">
          <h4 class="section-title">备注</h4>
          <p>{{ currentDetail.remark }}</p>
        </div>
      </div>
      <template #footer>
        <el-button @click="detailVisible = false">关闭</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Search, Refresh } from '@element-plus/icons-vue'
import {
  createSmsTask,
  updateSmsTask,
  deleteSmsTask,
  getSmsTask,
  getSmsTaskList,
  startSmsTask,
  pauseSmsTask,
  calculateTargetCount,
  getActiveTemplates
} from '@/api/sms'

const loading = ref(false)
const submitLoading = ref(false)
const list = ref([])
const total = ref(0)
const templateList = ref([])

const memberLevelList = [
  { value: 1, label: '普通会员' },
  { value: 2, label: '银卡会员' },
  { value: 3, label: '金卡会员' },
  { value: 4, label: '钻石会员' }
]

const query = reactive({
  name: '',
  status: '',
  schedule_type: '',
  page: 1,
  page_size: 10
})

const dialogVisible = ref(false)
const detailVisible = ref(false)
const isEdit = ref(false)
const editId = ref(null)
const taskFormRef = ref()
const currentDetail = ref(null)
const targetCount = ref(0)
const calculatingTarget = ref(false)

const defaultForm = () => ({
  name: '',
  task_type: 1,
  template_id: null,
  remark: '',
  schedule_type: 1,
  schedule_time: '',
  filter_levels: [],
  filter_consume_count_min: null,
  filter_consume_count_max: null,
  filter_consume_amount_min: null,
  filter_consume_amount_max: null,
  filter_points_min: null,
  filter_points_max: null
})

const taskForm = reactive(defaultForm())

const taskRules = {
  name: [{ required: true, message: '请输入任务名称', trigger: 'blur' }],
  task_type: [{ required: true, message: '请选择任务类型', trigger: 'change' }],
  template_id: [{ required: true, message: '请选择短信模板', trigger: 'change' }],
  schedule_type: [{ required: true, message: '请选择发送类型', trigger: 'change' }],
  schedule_time: [{ required: true, message: '请选择定时发送时间', trigger: 'change' }]
}

const selectedTemplate = computed(() => {
  if (!taskForm.template_id) return null
  return templateList.value.find(t => t.id === taskForm.template_id)
})

const statusMap = {
  0: { name: '草稿', type: 'info' },
  1: { name: '待发送', type: 'warning' },
  2: { name: '发送中', type: 'primary' },
  3: { name: '已完成', type: 'success' },
  4: { name: '已暂停', type: 'warning' },
  5: { name: '已取消', type: 'danger' }
}

const taskTypeMap = {
  1: { name: '营销短信', type: 'danger' },
  2: { name: '通知短信', type: 'primary' },
  3: { name: '会员关怀', type: 'success' },
  99: { name: '其他', type: 'info' }
}

function getStatusName(status) {
  return statusMap[status]?.name || '未知'
}

function getStatusTagType(status) {
  return statusMap[status]?.type || 'info'
}

function getTaskTypeName(type) {
  return taskTypeMap[type]?.name || '未知'
}

function getTaskTypeTagType(type) {
  return taskTypeMap[type]?.type || 'info'
}

function formatSuccessRate(rate) {
  if (rate === null || rate === undefined) return '-'
  return (rate * 100).toFixed(2) + '%'
}

function getSuccessRateClass(rate) {
  if (rate === null || rate === undefined) return ''
  if (rate >= 0.9) return 'high'
  if (rate >= 0.7) return 'medium'
  return 'low'
}

async function fetchList() {
  loading.value = true
  try {
    const params = { ...query }
    if (params.status === '') delete params.status
    if (params.schedule_type === '') delete params.schedule_type
    const res = await getSmsTaskList(params)
    list.value = res.list || []
    total.value = res.total || 0
  } catch (e) {
    console.error(e)
  } finally {
    loading.value = false
  }
}

async function fetchTemplates() {
  try {
    const res = await getActiveTemplates()
    templateList.value = res || []
  } catch (e) {
    console.error(e)
  }
}

function resetQuery() {
  query.name = ''
  query.status = ''
  query.schedule_type = ''
  query.page = 1
  fetchList()
}

function handleTemplateChange() {
}

let filterTimer = null
function handleFilterChange() {
  if (filterTimer) clearTimeout(filterTimer)
  filterTimer = setTimeout(() => {
    calculateTarget()
  }, 500)
}

async function calculateTarget() {
  try {
    calculatingTarget.value = true
    const params = {
      filter_levels: taskForm.filter_levels,
      filter_consume_count_min: taskForm.filter_consume_count_min,
      filter_consume_count_max: taskForm.filter_consume_count_max,
      filter_consume_amount_min: taskForm.filter_consume_amount_min,
      filter_consume_amount_max: taskForm.filter_consume_amount_max,
      filter_points_min: taskForm.filter_points_min,
      filter_points_max: taskForm.filter_points_max
    }
    const res = await calculateTargetCount(params)
    targetCount.value = res?.count || 0
  } catch (e) {
    console.error(e)
    targetCount.value = 0
  } finally {
    calculatingTarget.value = false
  }
}

function openDialog(row = null) {
  isEdit.value = !!row
  editId.value = row?.id || null

  Object.assign(taskForm, defaultForm())
  targetCount.value = 0

  if (row) {
    Object.assign(taskForm, {
      name: row.name,
      task_type: row.task_type,
      template_id: row.template_id,
      remark: row.remark || '',
      schedule_type: row.schedule_type,
      schedule_time: row.schedule_time || '',
      filter_levels: row.filter_levels || [],
      filter_consume_count_min: row.filter_consume_count_min || null,
      filter_consume_count_max: row.filter_consume_count_max || null,
      filter_consume_amount_min: row.filter_consume_amount_min || null,
      filter_consume_amount_max: row.filter_consume_amount_max || null,
      filter_points_min: row.filter_points_min || null,
      filter_points_max: row.filter_points_max || null
    })
  }

  fetchTemplates()
  dialogVisible.value = true

  setTimeout(() => {
    calculateTarget()
  }, 300)
}

function handleEdit(row) {
  openDialog(row)
}

function handleDelete(row) {
  ElMessageBox.confirm(`确定删除任务"${row.name}"吗？`, '提示', {
    type: 'warning',
    confirmButtonText: '确定',
    cancelButtonText: '取消'
  }).then(async () => {
    try {
      await deleteSmsTask(row.id)
      ElMessage.success('删除成功')
      fetchList()
    } catch (e) {
      console.error(e)
    }
  })
}

async function handleViewDetail(row) {
  try {
    const res = await getSmsTask(row.id)
    currentDetail.value = res
    detailVisible.value = true
  } catch (e) {
    console.error(e)
  }
}

async function handleStart(row) {
  ElMessageBox.confirm(`确定启动任务"${row.name}"吗？`, '提示', {
    type: 'warning',
    confirmButtonText: '确定',
    cancelButtonText: '取消'
  }).then(async () => {
    try {
      await startSmsTask(row.id)
      ElMessage.success('启动成功')
      fetchList()
    } catch (e) {
      console.error(e)
    }
  })
}

async function handlePause(row) {
  ElMessageBox.confirm(`确定暂停任务"${row.name}"吗？`, '提示', {
    type: 'warning',
    confirmButtonText: '确定',
    cancelButtonText: '取消'
  }).then(async () => {
    try {
      await pauseSmsTask(row.id)
      ElMessage.success('暂停成功')
      fetchList()
    } catch (e) {
      console.error(e)
    }
  })
}

async function handleSaveDraft() {
  try {
    await taskFormRef.value.validateField(['name', 'task_type', 'template_id'])
    submitLoading.value = true

    const data = buildSubmitData()
    data.status = 0

    if (isEdit.value) {
      await updateSmsTask(editId.value, data)
      ElMessage.success('保存成功')
    } else {
      await createSmsTask(data)
      ElMessage.success('创建草稿成功')
    }

    dialogVisible.value = false
    fetchList()
  } catch (e) {
    console.error(e)
  } finally {
    submitLoading.value = false
  }
}

async function handleSubmit() {
  try {
    await taskFormRef.value.validate()
    submitLoading.value = true

    const data = buildSubmitData()
    data.status = 1

    if (isEdit.value) {
      await updateSmsTask(editId.value, data)
      ElMessage.success('保存并发送成功')
    } else {
      await createSmsTask(data)
      ElMessage.success('任务已提交发送')
    }

    dialogVisible.value = false
    fetchList()
  } catch (e) {
    console.error(e)
  } finally {
    submitLoading.value = false
  }
}

function buildSubmitData() {
  const data = {
    name: taskForm.name,
    task_type: taskForm.task_type,
    template_id: taskForm.template_id,
    remark: taskForm.remark,
    schedule_type: taskForm.schedule_type,
    schedule_time: taskForm.schedule_type === 2 ? taskForm.schedule_time : null,
    filter_levels: taskForm.filter_levels,
    filter_consume_count_min: taskForm.filter_consume_count_min,
    filter_consume_count_max: taskForm.filter_consume_count_max,
    filter_consume_amount_min: taskForm.filter_consume_amount_min,
    filter_consume_amount_max: taskForm.filter_consume_amount_max,
    filter_points_min: taskForm.filter_points_min,
    filter_points_max: taskForm.filter_points_max
  }
  return data
}

onMounted(() => {
  fetchList()
  fetchTemplates()
})
</script>

<style scoped lang="scss">
.sms-tasks-page {
  .header-actions {
    display: flex;
    gap: 12px;
  }

  .pagination {
    margin-top: 20px;
    display: flex;
    justify-content: flex-end;
  }

  .stat-text {
    font-size: 13px;

    .success {
      color: #67c23a;
      font-weight: 600;
    }

    .fail {
      color: #f56c6c;
      font-weight: 600;
    }

    .divider {
      color: #c0c4cc;
      margin: 0 4px;
    }
  }

  .success-rate {
    font-weight: 600;

    &.high {
      color: #67c23a;
    }

    &.medium {
      color: #e6a23c;
    }

    &.low {
      color: #f56c6c;
    }
  }

  .task-dialog {
    :deep(.el-dialog__body) {
      max-height: 600px;
      overflow-y: auto;
    }
  }

  .form-section {
    margin-bottom: 24px;
    padding-bottom: 20px;
    border-bottom: 1px solid #ebeef5;

    &:last-child {
      border-bottom: none;
      margin-bottom: 0;
      padding-bottom: 0;
    }

    .section-title {
      margin: 0 0 16px 0;
      font-size: 15px;
      font-weight: 600;
      color: #303133;
      display: flex;
      align-items: center;
      justify-content: space-between;

      .target-count {
        font-size: 13px;
        font-weight: 400;
        color: #606266;

        b {
          color: #409eff;
          font-size: 16px;
        }
      }
    }
  }

  .template-preview {
    width: 100%;
    min-height: 60px;
    padding: 12px;
    background: #f5f7fa;
    border: 1px solid #dcdfe6;
    border-radius: 4px;
    font-size: 13px;
    color: #606266;
    line-height: 1.6;
  }

  .range-input {
    display: flex;
    align-items: center;
    gap: 8px;

    .range-separator {
      color: #909399;
      flex-shrink: 0;
    }
  }

  .task-detail {
    .detail-section {
      margin-top: 20px;

      .section-title {
        margin: 0 0 12px 0;
        font-size: 14px;
        font-weight: 600;
        color: #303133;
      }
    }

    .stat-cards {
      display: flex;
      gap: 12px;
      flex-wrap: wrap;

      .stat-card {
        flex: 1;
        min-width: 100px;
        padding: 16px;
        background: #f5f7fa;
        border-radius: 8px;
        text-align: center;

        .stat-label {
          font-size: 12px;
          color: #909399;
          margin-bottom: 8px;
        }

        .stat-value {
          font-size: 20px;
          font-weight: 600;
          color: #303133;
        }

        &.success .stat-value {
          color: #67c23a;
        }

        &.fail .stat-value {
          color: #f56c6c;
        }
      }
    }
  }
}
</style>
