<template>
  <div class="coupons-page">
    <div class="page-header">
      <h2 class="page-title">优惠券管理</h2>
      <div class="header-actions">
        <el-button type="success" @click="openDialog">
          <el-icon><Plus /></el-icon>新增优惠券
        </el-button>
      </div>
    </div>

    <div class="card-wrapper">
      <div class="search-bar">
        <el-input
          v-model="query.name"
          placeholder="搜索优惠券名称"
          clearable
          style="width: 240px"
          @keyup.enter="fetchList" />
        <el-select v-model="query.type" placeholder="优惠券类型" clearable style="width: 140px">
          <el-option label="满减券" :value="1" />
          <el-option label="折扣券" :value="2" />
          <el-option label="代金券" :value="3" />
        </el-select>
        <el-select v-model="query.status" placeholder="状态" clearable style="width: 140px">
          <el-option label="未开始" :value="0" />
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
        <el-table-column prop="rule_key" label="规则标识" width="120">
          <template #default="{ row }">
            <el-tag v-if="row.rule_key" size="small" :type="row.rule_key === 'birthday' ? 'danger' : 'info'">
              {{ row.rule_key }}
            </el-tag>
            <span v-else style="color: #c0c4cc;">-</span>
          </template>
        </el-table-column>
        <el-table-column prop="name" label="优惠券名称" min-width="180" />
        <el-table-column label="类型" width="100">
          <template #default="{ row }">
            <el-tag :type="getTypeTagType(row.type)">{{ getTypeName(row.type) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="面值" width="140">
          <template #default="{ row }">
            <span v-if="row.type === 2" class="discount-value">{{ row.value }}折</span>
            <span v-else class="amount-value">¥{{ row.value }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="min_amount" label="最低消费" width="120">
          <template #default="{ row }">
            ¥{{ row.min_amount || 0 }}
          </template>
        </el-table-column>
        <el-table-column prop="total_quantity" label="发放总量" width="100" align="center" />
        <el-table-column prop="used_quantity" label="已使用" width="100" align="center" />
        <el-table-column label="有效期" width="240">
          <template #default="{ row }">
            <div class="date-range">
              <div>{{ row.start_date }}</div>
              <div class="divider">至</div>
              <div>{{ row.end_date }}</div>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="100" align="center">
          <template #default="{ row }">
            <el-tag :type="getStatusTagType(row.status)">{{ getStatusName(row.status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="创建时间" width="160" />
        <el-table-column label="操作" width="240" fixed="right">
          <template #default="{ row }">
            <el-button type="success" link size="small" @click="openIssueDialog(row)" :disabled="row.status === 2">发放</el-button>
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
      :title="isEdit ? '编辑优惠券' : '新增优惠券'"
      width="700px"
      :close-on-click-modal="false">
      <el-form
        ref="couponFormRef"
        :model="couponForm"
        :rules="couponRules"
        label-width="120px">
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="规则标识">
              <el-input v-model="couponForm.rule_key" placeholder="请输入规则标识，如 birthday">
                <template #append>
                  <el-button @click="couponForm.rule_key = 'birthday'" :type="couponForm.rule_key === 'birthday' ? 'primary' : ''">生日券</el-button>
                </template>
              </el-input>
              <div style="margin-top: 4px; font-size: 12px; color: #909399;">
                设置为 birthday 可作为生日自动发放优惠券
              </div>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="优惠券名称" prop="name">
              <el-input v-model="couponForm.name" placeholder="请输入优惠券名称" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="优惠券类型" prop="type">
              <el-select v-model="couponForm.type" placeholder="请选择类型" style="width: 100%" @change="handleTypeChange">
                <el-option label="满减券" :value="1" />
                <el-option label="折扣券" :value="2" />
                <el-option label="代金券" :value="3" />
              </el-select>
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="面值" prop="value">
              <el-input-number
                v-model="couponForm.value"
                :min="0.01"
                :max="couponForm.type === 2 ? 9.99 : 9999"
                :precision="couponForm.type === 2 ? 1 : 2"
                :step="couponForm.type === 2 ? 0.1 : 1"
                style="width: 100%" />
              <span v-if="couponForm.type === 2" style="margin-left: 8px; color: #909399;">折</span>
              <span v-else style="margin-left: 8px; color: #909399;">元</span>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="最低消费">
              <el-input-number
                v-model="couponForm.min_amount"
                :min="0"
                :precision="2"
                style="width: 100%" />
              <span style="margin-left: 8px; color: #909399;">元</span>
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="发放总量" prop="total_quantity">
              <el-input-number
                v-model="couponForm.total_quantity"
                :min="-1"
                style="width: 100%" />
              <span style="margin-left: 8px; color: #909399;">张 (-1为不限量)</span>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="每人限领">
              <el-input-number
                v-model="couponForm.limit_per_user"
                :min="1"
                style="width: 100%" />
              <span style="margin-left: 8px; color: #909399;">张</span>
            </el-form-item>
          </el-col>
        </el-row>
        <el-form-item label="有效期" prop="date_range">
          <el-date-picker
            v-model="couponForm.date_range"
            type="daterange"
            range-separator="至"
            start-placeholder="开始日期"
            end-placeholder="结束日期"
            value-format="YYYY-MM-DD"
            style="width: 100%" />
        </el-form-item>
        <el-form-item label="使用说明">
          <el-input
            v-model="couponForm.description"
            type="textarea"
            :rows="2"
            placeholder="请输入使用说明" />
        </el-form-item>
        <el-form-item label="状态">
          <el-switch v-model="couponForm.status" :active-value="1" :inactive-value="0" active-text="启用" inactive-text="禁用" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitLoading" @click="handleSubmit">确定</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="issueDialogVisible" title="发放优惠券" width="600px">
      <el-form label-width="100px">
        <el-form-item label="优惠券">
          <span>{{ currentCoupon?.name }}</span>
        </el-form-item>
        <el-form-item label="面值">
          <span v-if="currentCoupon?.type === 2">{{ currentCoupon?.value }}折</span>
          <span v-else>¥{{ currentCoupon?.value }}</span>
        </el-form-item>
        <el-form-item label="发放方式">
          <el-radio-group v-model="issueForm.type">
            <el-radio value="all">全体会员</el-radio>
            <el-radio value="指定">指定会员</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item v-if="issueForm.type === '指定'" label="选择会员">
          <el-select
            v-model="issueForm.member_ids"
            multiple
            filterable
            placeholder="请选择会员"
            style="width: 100%">
            <el-option
              v-for="member in memberList"
              :key="member.id"
              :label="`${member.name} - ${member.phone}`"
              :value="member.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="发放数量">
          <el-input-number v-model="issueForm.quantity" :min="1" style="width: 200px" />
          <span style="margin-left: 8px; color: #909399;">张/人</span>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="issueDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="issueLoading" @click="handleIssue">确定发放</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Search, Refresh } from '@element-plus/icons-vue'
import {
  getCouponList,
  createCoupon,
  updateCoupon,
  deleteCoupon,
  issueCoupon
} from '@/api/coupons'
import { getMemberList } from '@/api/members'

const loading = ref(false)
const submitLoading = ref(false)
const issueLoading = ref(false)
const list = ref([])
const total = ref(0)
const currentCoupon = ref(null)
const memberList = ref([])

const query = reactive({
  name: '',
  type: null,
  status: null,
  page: 1,
  page_size: 10
})

const dialogVisible = ref(false)
const issueDialogVisible = ref(false)
const isEdit = ref(false)
const editId = ref(null)
const couponFormRef = ref()

const couponForm = reactive({
  rule_key: '',
  name: '',
  type: 1,
  value: 0,
  min_amount: 0,
  total_quantity: -1,
  limit_per_user: 1,
  date_range: [],
  start_date: '',
  end_date: '',
  description: '',
  status: 1
})

const issueForm = reactive({
  coupon_id: null,
  type: 'all',
  member_ids: [],
  quantity: 1
})

const couponRules = {
  name: [{ required: true, message: '请输入优惠券名称', trigger: 'blur' }],
  type: [{ required: true, message: '请选择优惠券类型', trigger: 'change' }],
  value: [{ required: true, message: '请输入面值', trigger: 'blur' }],
  total_quantity: [{ required: true, message: '请输入发放总量', trigger: 'blur' }],
  date_range: [{ required: true, message: '请选择有效期', trigger: 'change' }]
}

const typeMap = {
  1: { name: '满减券', type: 'danger' },
  2: { name: '折扣券', type: 'success' },
  3: { name: '代金券', type: 'warning' }
}

const statusMap = {
  0: { name: '未开始', type: 'info' },
  1: { name: '进行中', type: 'success' },
  2: { name: '已结束', type: 'info' }
}

function getTypeName(type) {
  return typeMap[type]?.name || '未知'
}

function getTypeTagType(type) {
  return typeMap[type]?.type || 'info'
}

function getStatusName(status) {
  return statusMap[status]?.name || '未知'
}

function getStatusTagType(status) {
  return statusMap[status]?.type || 'info'
}

function handleTypeChange() {
  if (couponForm.type === 2) {
    couponForm.value = couponForm.value > 9.9 ? 9 : couponForm.value
  }
}

async function fetchList() {
  loading.value = true
  try {
    const res = await getCouponList(query)
    list.value = res.list || []
    total.value = res.total || 0
  } catch (e) {
    console.error(e)
  } finally {
    loading.value = false
  }
}

async function fetchMembers() {
  try {
    const res = await getMemberList({ page: 1, page_size: 1000 })
    memberList.value = res.list || []
  } catch (e) {
    console.error(e)
  }
}

function resetQuery() {
  query.name = ''
  query.type = null
  query.status = null
  query.page = 1
  fetchList()
}

function openDialog(row = null) {
  isEdit.value = !!row
  editId.value = row?.id || null

  if (row) {
    Object.assign(couponForm, {
      rule_key: row.rule_key || '',
      name: row.name,
      type: row.type,
      value: row.value,
      min_amount: row.min_amount || 0,
      total_quantity: row.total_quantity,
      limit_per_user: row.limit_per_user || 1,
      date_range: [row.start_date, row.end_date],
      start_date: row.start_date,
      end_date: row.end_date,
      description: row.description || '',
      status: row.status
    })
  } else {
    couponForm.rule_key = ''
    couponForm.name = ''
    couponForm.type = 1
    couponForm.value = 0
    couponForm.min_amount = 0
    couponForm.total_quantity = -1
    couponForm.limit_per_user = 1
    couponForm.date_range = []
    couponForm.start_date = ''
    couponForm.end_date = ''
    couponForm.description = ''
    couponForm.status = 1
  }

  dialogVisible.value = true
}

function handleEdit(row) {
  openDialog(row)
}

function openIssueDialog(row) {
  currentCoupon.value = row
  issueForm.coupon_id = row.id
  issueForm.type = 'all'
  issueForm.member_ids = []
  issueForm.quantity = 1
  fetchMembers()
  issueDialogVisible.value = true
}

async function handleIssue() {
  if (issueForm.type === '指定' && issueForm.member_ids.length === 0) {
    ElMessage.warning('请选择会员')
    return
  }

  try {
    issueLoading.value = true
    await issueCoupon({
      coupon_id: issueForm.coupon_id,
      type: issueForm.type,
      member_ids: issueForm.type === '指定' ? issueForm.member_ids : [],
      quantity: issueForm.quantity
    })
    ElMessage.success('发放成功')
    issueDialogVisible.value = false
    fetchList()
  } catch (e) {
    console.error(e)
  } finally {
    issueLoading.value = false
  }
}

function handleDelete(row) {
  ElMessageBox.confirm(`确定删除优惠券"${row.name}"吗？`, '提示', {
    type: 'warning',
    confirmButtonText: '确定',
    cancelButtonText: '取消'
  }).then(async () => {
    await deleteCoupon(row.id)
    ElMessage.success('删除成功')
    fetchList()
  })
}

async function handleSubmit() {
  try {
    await couponFormRef.value.validate()
    submitLoading.value = true

    const data = {
      ...couponForm,
      start_date: couponForm.date_range[0],
      end_date: couponForm.date_range[1]
    }
    delete data.date_range

    if (isEdit.value) {
      await updateCoupon(editId.value, data)
      ElMessage.success('更新成功')
    } else {
      await createCoupon(data)
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
.coupons-page {
  .header-actions {
    display: flex;
    gap: 12px;
  }

  .amount-value {
    color: #f56c6c;
    font-weight: 600;
    font-size: 16px;
  }

  .discount-value {
    color: #67c23a;
    font-weight: 600;
    font-size: 16px;
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
