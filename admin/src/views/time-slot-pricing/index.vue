<template>
  <div class="time-slot-pricing-page">
    <div class="page-header">
      <h2 class="page-title">时段定价配置</h2>
      <div class="header-actions">
        <el-button type="success" @click="openDialog">
          <el-icon><Plus /></el-icon>新增时段
        </el-button>
      </div>
    </div>

    <div class="card-wrapper">
      <div class="search-bar">
        <el-input
          v-model="query.name"
          placeholder="搜索时段名称"
          clearable
          style="width: 240px"
          @keyup.enter="fetchList" />
        <el-select v-model="query.status" placeholder="状态" clearable style="width: 140px">
          <el-option label="禁用" :value="0" />
          <el-option label="启用" :value="1" />
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
        <el-table-column prop="name" label="名称" min-width="180" />
        <el-table-column label="时段" width="180">
          <template #default="{ row }">
            <div class="time-range">
              <span>{{ row.start_time }}</span>
              <span class="divider">-</span>
              <span>{{ row.end_time }}</span>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="定价类型" width="120">
          <template #default="{ row }">
            <el-tag :type="getPricingTypeTagType(row.pricing_type)">{{ getPricingTypeName(row.pricing_type) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="优惠规则" min-width="200">
          <template #default="{ row }">
            <div v-if="row.pricing_type === 'discount'" class="rule-text">
              {{ row.discount_rate }} 折
            </div>
            <div v-else-if="row.pricing_type === 'full_reduction'" class="rule-text">
              满 ¥{{ row.min_amount }} 减 ¥{{ row.reduction_amount }}
            </div>
          </template>
        </el-table-column>
        <el-table-column label="适用范围" width="100">
          <template #default="{ row }">
            <el-tag size="small">{{ getApplicableName(row.applicable_type) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="星期" width="180">
          <template #default="{ row }">
            <div class="weekday-tags">
              <el-tag
                v-for="day in parseWeekdays(row.weekdays)"
                :key="day"
                size="small"
                type="info"
                style="margin-right: 4px; margin-bottom: 4px;">
                {{ getWeekdayName(day) }}
              </el-tag>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="priority" label="优先级" width="80" align="center" />
        <el-table-column label="状态" width="100" align="center">
          <template #default="{ row }">
            <el-tag :type="getStatusTagType(row.status)">{{ getStatusName(row.status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="创建时间" width="160" />
        <el-table-column label="操作" width="180" fixed="right">
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
      :title="isEdit ? '编辑时段' : '新增时段'"
      width="700px"
      :close-on-click-modal="false">
      <el-form
        ref="timeSlotFormRef"
        :model="timeSlotForm"
        :rules="timeSlotRules"
        label-width="120px">
        <el-row :gutter="20">
          <el-col :span="24">
            <el-form-item label="时段名称" prop="name">
              <el-input v-model="timeSlotForm.name" placeholder="请输入时段名称" />
            </el-form-item>
          </el-col>
        </el-row>

        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="开始时间" prop="start_time">
              <el-time-picker
                v-model="timeSlotForm.start_time"
                value-format="HH:mm"
                placeholder="选择开始时间"
                style="width: 100%" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="结束时间" prop="end_time">
              <el-time-picker
                v-model="timeSlotForm.end_time"
                value-format="HH:mm"
                placeholder="选择结束时间"
                style="width: 100%" />
            </el-form-item>
          </el-col>
        </el-row>

        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="定价类型" prop="pricing_type">
              <el-select v-model="timeSlotForm.pricing_type" placeholder="请选择类型" style="width: 100%">
                <el-option label="折扣" value="discount" />
                <el-option label="满减" value="full_reduction" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="优先级" prop="priority">
              <el-input-number
                v-model="timeSlotForm.priority"
                :min="1"
                :max="999"
                style="width: 100%" />
              <span style="margin-left: 8px; color: #909399;">数字越小优先级越高</span>
            </el-form-item>
          </el-col>
        </el-row>

        <el-form-item v-if="timeSlotForm.pricing_type === 'discount'" label="折扣率" prop="discount_rate">
          <el-input-number
            v-model="timeSlotForm.discount_rate"
            :min="1"
            :max="100"
            :step="1" />
          <span style="margin-left: 8px;">%（如50表示5折）</span>
        </el-form-item>

        <el-form-item v-if="timeSlotForm.pricing_type === 'full_reduction'" label="满减规则">
          <el-row :gutter="20">
            <el-col :span="12">
              <el-form-item prop="min_amount" label-width="0">
                <el-input-number
                  v-model="timeSlotForm.min_amount"
                  :min="0"
                  :precision="2"
                  style="width: 100%" />
                <span style="margin-left: 8px;">元</span>
              </el-form-item>
            </el-col>
            <el-col :span="12">
              <el-form-item prop="reduction_amount" label-width="0">
                <span style="margin-right: 8px;">减</span>
                <el-input-number
                  v-model="timeSlotForm.reduction_amount"
                  :min="0"
                  :precision="2"
                  style="width: 100%" />
                <span style="margin-left: 8px;">元</span>
              </el-form-item>
            </el-col>
          </el-row>
        </el-form-item>

        <el-form-item label="适用范围" prop="applicable_type">
          <el-radio-group v-model="timeSlotForm.applicable_type">
            <el-radio value="all">全部商品</el-radio>
            <el-radio value="category">指定分类</el-radio>
            <el-radio value="product">指定商品</el-radio>
          </el-radio-group>
        </el-form-item>

        <el-form-item v-if="timeSlotForm.applicable_type !== 'all'" label="适用ID列表" prop="applicable_ids">
          <el-select
            v-model="timeSlotForm.applicable_ids"
            multiple
            filterable
            placeholder="请选择适用的分类/商品"
            style="width: 100%">
            <el-option
              v-for="p in productList"
              :key="p.id"
              :label="p.name"
              :value="p.id" />
          </el-select>
        </el-form-item>

        <el-form-item label="星期选择" prop="weekdays">
          <el-checkbox-group v-model="timeSlotForm.weekdays">
            <el-checkbox :value="1">周一</el-checkbox>
            <el-checkbox :value="2">周二</el-checkbox>
            <el-checkbox :value="3">周三</el-checkbox>
            <el-checkbox :value="4">周四</el-checkbox>
            <el-checkbox :value="5">周五</el-checkbox>
            <el-checkbox :value="6">周六</el-checkbox>
            <el-checkbox :value="7">周日</el-checkbox>
          </el-checkbox-group>
        </el-form-item>

        <el-form-item label="状态">
          <el-switch v-model="timeSlotForm.status" :active-value="1" :inactive-value="0" active-text="启用" inactive-text="禁用" />
        </el-form-item>

        <el-form-item label="描述">
          <el-input
            v-model="timeSlotForm.description"
            type="textarea"
            :rows="2"
            placeholder="请输入描述" />
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
  getTimeSlotPricingList,
  getTimeSlotPricing,
  createTimeSlotPricing,
  updateTimeSlotPricing,
  deleteTimeSlotPricing
} from '@/api/timeSlotPricing'
import { getProductList } from '@/api/product'

const loading = ref(false)
const submitLoading = ref(false)
const list = ref([])
const total = ref(0)
const productList = ref([])

const query = reactive({
  name: '',
  status: null,
  page: 1,
  page_size: 10
})

const dialogVisible = ref(false)
const isEdit = ref(false)
const editId = ref(null)
const timeSlotFormRef = ref()

const defaultForm = () => ({
  name: '',
  start_time: '',
  end_time: '',
  pricing_type: 'discount',
  discount_rate: 100,
  min_amount: 0,
  reduction_amount: 0,
  applicable_type: 'all',
  applicable_ids: [],
  weekdays: [1, 2, 3, 4, 5, 6, 7],
  priority: 100,
  description: '',
  status: 1
})

const timeSlotForm = reactive(defaultForm())

const timeSlotRules = {
  name: [{ required: true, message: '请输入时段名称', trigger: 'blur' }],
  start_time: [{ required: true, message: '请选择开始时间', trigger: 'change' }],
  end_time: [{ required: true, message: '请选择结束时间', trigger: 'change' }],
  pricing_type: [{ required: true, message: '请选择定价类型', trigger: 'change' }],
  discount_rate: [{ required: true, message: '请输入折扣率', trigger: 'blur' }],
  applicable_type: [{ required: true, message: '请选择适用范围', trigger: 'change' }],
  weekdays: [{ required: true, message: '请选择星期', trigger: 'change' }]
}

const pricingTypeMap = {
  discount: { name: '折扣', type: 'success' },
  full_reduction: { name: '满减', type: 'danger' }
}

const statusMap = {
  0: { name: '禁用', type: 'info' },
  1: { name: '启用', type: 'success' }
}

const applicableMap = {
  all: '全部商品',
  category: '指定分类',
  product: '指定商品'
}

const weekdayMap = {
  1: '周一',
  2: '周二',
  3: '周三',
  4: '周四',
  5: '周五',
  6: '周六',
  7: '周日'
}

function getPricingTypeName(type) {
  return pricingTypeMap[type]?.name || '未知'
}

function getPricingTypeTagType(type) {
  return pricingTypeMap[type]?.type || 'info'
}

function getStatusName(status) {
  return statusMap[status]?.name || '未知'
}

function getStatusTagType(status) {
  return statusMap[status]?.type || 'info'
}

function getApplicableName(type) {
  return applicableMap[type] || '未知'
}

function getWeekdayName(day) {
  return weekdayMap[day] || '未知'
}

function parseWeekdays(weekdays) {
  if (!weekdays) return []
  if (Array.isArray(weekdays)) return weekdays
  return String(weekdays).split(',').map(Number).filter(n => !isNaN(n))
}

function parseApplicableIds(ids) {
  if (!ids) return []
  if (Array.isArray(ids)) return ids
  return String(ids).split(',').map(id => parseInt(id)).filter(id => !isNaN(id))
}

async function fetchList() {
  loading.value = true
  try {
    const res = await getTimeSlotPricingList(query)
    list.value = res.list || []
    total.value = res.total || 0
  } catch (e) {
    console.error(e)
  } finally {
    loading.value = false
  }
}

async function fetchProducts() {
  try {
    const res = await getProductList({ page: 1, page_size: 1000 })
    productList.value = res.list || []
  } catch (e) {
    console.error(e)
  }
}

function resetQuery() {
  query.name = ''
  query.status = null
  query.page = 1
  fetchList()
}

function openDialog(row = null) {
  isEdit.value = !!row
  editId.value = row?.id || null
  fetchProducts()

  Object.assign(timeSlotForm, defaultForm())

  if (row) {
    Object.assign(timeSlotForm, {
      name: row.name,
      start_time: row.start_time,
      end_time: row.end_time,
      pricing_type: row.pricing_type,
      discount_rate: row.discount_rate || 100,
      min_amount: row.min_amount || 0,
      reduction_amount: row.reduction_amount || 0,
      applicable_type: row.applicable_type || 'all',
      applicable_ids: parseApplicableIds(row.applicable_ids),
      weekdays: parseWeekdays(row.weekdays),
      priority: row.priority || 100,
      description: row.description || '',
      status: row.status
    })
  }

  dialogVisible.value = true
}

function handleEdit(row) {
  openDialog(row)
}

function handleDelete(row) {
  ElMessageBox.confirm(`确定删除时段"${row.name}"吗？`, '提示', {
    type: 'warning',
    confirmButtonText: '确定',
    cancelButtonText: '取消'
  }).then(async () => {
    await deleteTimeSlotPricing(row.id)
    ElMessage.success('删除成功')
    fetchList()
  })
}

async function handleSubmit() {
  try {
    await timeSlotFormRef.value.validate()
    submitLoading.value = true

    const data = {
      ...timeSlotForm
    }

    if (isEdit.value) {
      await updateTimeSlotPricing(editId.value, data)
      ElMessage.success('更新成功')
    } else {
      await createTimeSlotPricing(data)
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
.time-slot-pricing-page {
  .header-actions {
    display: flex;
    gap: 12px;
  }

  .rule-text {
    font-size: 14px;
    color: #303133;
    font-weight: 500;
  }

  .time-range {
    display: flex;
    align-items: center;
    gap: 8px;
    font-size: 14px;
    font-weight: 500;

    .divider {
      color: #909399;
    }
  }

  .weekday-tags {
    display: flex;
    flex-wrap: wrap;
  }

  .pagination {
    margin-top: 20px;
    display: flex;
    justify-content: flex-end;
  }
}
</style>
