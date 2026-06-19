<template>
  <div class="recharge-activities-page">
    <div class="page-header">
      <h2 class="page-title">充值活动管理</h2>
      <div class="header-actions">
        <el-button type="success" @click="openDialog">
          <el-icon><Plus /></el-icon>新增活动
        </el-button>
      </div>
    </div>

    <div class="card-wrapper">
      <div class="search-bar">
        <el-select v-model="query.status" placeholder="活动状态" clearable style="width: 140px">
          <el-option label="待生效" :value="0" />
          <el-option label="进行中" :value="1" />
          <el-option label="已结束" :value="2" />
        </el-select>
        <el-button type="primary" @click="fetchList">
          <el-icon><Search /></el-icon>搜索
        </el-button>
        <el-button @click="resetQuery">
          <el-icon><Refresh /></el-icon>重置
        </el-button>
      </div>

      <el-table :data="list" v-loading="loading">
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="name" label="活动名称" min-width="180" />
        <el-table-column label="充值门槛" width="120" align="center">
          <template #default="{ row }">
            <span class="amount">¥{{ row.min_amount }}</span>
          </template>
        </el-table-column>
        <el-table-column label="赠送金额" width="120" align="center">
          <template #default="{ row }">
            <span v-if="row.bonus_amount > 0" class="bonus">¥{{ row.bonus_amount }}</span>
            <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column label="赠送积分" width="120" align="center">
          <template #default="{ row }">
            <span v-if="row.bonus_points > 0" class="points">{{ row.bonus_points }}</span>
            <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column label="活动时间" width="300">
          <template #default="{ row }">
            <div class="date-range">
              <div>{{ formatTime(row.start_time) }}</div>
              <div class="divider">至</div>
              <div>{{ formatTime(row.end_time) }}</div>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="自动生效" width="100" align="center">
          <template #default="{ row }">
            <el-tag :type="row.auto_activate ? 'success' : 'info'" size="small">
              {{ row.auto_activate ? '是' : '否' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="100" align="center">
          <template #default="{ row }">
            <el-tag :type="getStatusTagType(row.status)">{{ getStatusName(row.status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="handleEdit(row)">编辑</el-button>
            <el-button type="danger" link size="small" @click="handleDelete(row)">删除</el-button>
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
      :title="isEdit ? '编辑充值活动' : '新增充值活动'"
      width="650px"
      :close-on-click-modal="false">
      <el-form
        ref="formRef"
        :model="form"
        :rules="formRules"
        label-width="120px">
        <el-form-item label="活动名称" prop="name">
          <el-input v-model="form.name" placeholder="请输入活动名称" />
        </el-form-item>
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="充值门槛" prop="min_amount">
              <el-input-number v-model="form.min_amount" :min="0.01" :precision="2" style="width: 100%" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="赠送金额">
              <el-input-number v-model="form.bonus_amount" :min="0" :precision="2" style="width: 100%" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="赠送积分">
              <el-input-number v-model="form.bonus_points" :min="0" style="width: 100%" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="状态">
              <el-select v-model="form.status" style="width: 100%">
                <el-option label="待生效" :value="0" />
                <el-option label="进行中" :value="1" />
                <el-option label="已结束" :value="2" />
              </el-select>
            </el-form-item>
          </el-col>
        </el-row>
        <el-form-item label="活动时间" prop="date_range">
          <el-date-picker
            v-model="form.date_range"
            type="datetimerange"
            range-separator="至"
            start-placeholder="开始时间"
            end-placeholder="结束时间"
            value-format="YYYY-MM-DDTHH:mm:ssZ"
            style="width: 100%" />
        </el-form-item>
        <el-form-item label="自动生效">
          <el-switch v-model="form.auto_activate" active-text="到期自动生效" inactive-text="手动生效" />
          <div style="margin-top: 4px; color: #909399; font-size: 12px;">
            开启后，到开始时间时系统将自动将活动状态变更为进行中
          </div>
        </el-form-item>
        <el-form-item label="活动说明">
          <el-input
            v-model="form.description"
            type="textarea"
            :rows="2"
            placeholder="请输入活动说明" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitLoading" @click="handleSubmit">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Search, Refresh } from '@element-plus/icons-vue'
import {
  getRechargeActivityList,
  createRechargeActivity,
  updateRechargeActivity,
  deleteRechargeActivity
} from '@/api/recharge-activities'
import dayjs from 'dayjs'

const loading = ref(false)
const submitLoading = ref(false)
const list = ref([])
const total = ref(0)

const query = reactive({
  status: null,
  page: 1,
  page_size: 10
})

const dialogVisible = ref(false)
const isEdit = ref(false)
const editId = ref(null)
const formRef = ref()

const form = reactive({
  name: '',
  min_amount: 100,
  bonus_amount: 0,
  bonus_points: 0,
  date_range: [],
  auto_activate: true,
  status: 0,
  description: ''
})

const formRules = {
  name: [{ required: true, message: '请输入活动名称', trigger: 'blur' }],
  min_amount: [{ required: true, message: '请输入充值门槛', trigger: 'blur' }],
  date_range: [{ required: true, message: '请选择活动时间', trigger: 'change' }]
}

const statusMap = {
  0: { name: '待生效', type: 'info' },
  1: { name: '进行中', type: 'success' },
  2: { name: '已结束', type: 'info' }
}

function getStatusName(status) {
  return statusMap[status]?.name || '未知'
}

function getStatusTagType(status) {
  return statusMap[status]?.type || 'info'
}

function formatTime(time) {
  if (!time) return '-'
  return dayjs(time).format('YYYY-MM-DD HH:mm')
}

async function fetchList() {
  loading.value = true
  try {
    const res = await getRechargeActivityList(query)
    list.value = res.list || []
    total.value = res.total || 0
  } catch (e) {
    console.error(e)
  } finally {
    loading.value = false
  }
}

function resetQuery() {
  query.status = null
  query.page = 1
  fetchList()
}

function openDialog(row = null) {
  isEdit.value = !!row
  editId.value = row?.id || null

  if (row) {
    Object.assign(form, {
      name: row.name,
      min_amount: row.min_amount,
      bonus_amount: row.bonus_amount || 0,
      bonus_points: row.bonus_points || 0,
      date_range: [row.start_time, row.end_time],
      auto_activate: row.auto_activate || false,
      status: row.status,
      description: row.description || ''
    })
  } else {
    form.name = ''
    form.min_amount = 100
    form.bonus_amount = 0
    form.bonus_points = 0
    form.date_range = []
    form.auto_activate = true
    form.status = 0
    form.description = ''
  }

  dialogVisible.value = true
}

function handleEdit(row) {
  openDialog(row)
}

function handleDelete(row) {
  ElMessageBox.confirm(`确定删除活动"${row.name}"吗？`, '提示', {
    type: 'warning',
    confirmButtonText: '确定',
    cancelButtonText: '取消'
  }).then(async () => {
    await deleteRechargeActivity(row.id)
    ElMessage.success('删除成功')
    fetchList()
  })
}

async function handleSubmit() {
  try {
    await formRef.value.validate()
    submitLoading.value = true

    const data = {
      name: form.name,
      min_amount: form.min_amount,
      bonus_amount: form.bonus_amount,
      bonus_points: form.bonus_points,
      start_time: form.date_range[0],
      end_time: form.date_range[1],
      auto_activate: form.auto_activate,
      status: form.status,
      description: form.description
    }

    if (isEdit.value) {
      await updateRechargeActivity(editId.value, data)
      ElMessage.success('更新成功')
    } else {
      await createRechargeActivity(data)
      ElMessage.success('创建成功')
    }

    dialogVisible.value = false
    fetchList()
  } catch (e) {
    console.error(e)
  } finally {
    submitLoading.value = false
  }
}

onMounted(() => {
  fetchList()
})
</script>

<style scoped lang="scss">
.recharge-activities-page {
  .header-actions {
    display: flex;
    gap: 12px;
  }

  .amount {
    color: #409eff;
    font-weight: 600;
  }

  .bonus {
    color: #f56c6c;
    font-weight: 600;
  }

  .points {
    color: #e6a23c;
    font-weight: 600;
  }

  .date-range {
    display: flex;
    flex-direction: column;
    gap: 4px;
    font-size: 13px;

    .divider {
      color: #909399;
    }
  }

  .pagination {
    margin-top: 20px;
    display: flex;
    justify-content: flex-end;
  }
}
</style>
