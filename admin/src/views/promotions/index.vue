<template>
  <div class="promotions-page">
    <div class="page-header">
      <h2 class="page-title">营销活动管理</h2>
      <div class="header-actions">
        <el-button type="success" @click="openDialog">
          <el-icon><Plus /></el-icon>新增活动
        </el-button>
      </div>
    </div>

    <div class="card-wrapper">
      <div class="search-bar">
        <el-input
          v-model="query.name"
          placeholder="搜索活动名称"
          clearable
          style="width: 240px"
          @keyup.enter="fetchList" />
        <el-select v-model="query.type" placeholder="活动类型" clearable style="width: 140px">
          <el-option label="满减活动" value="full_reduction" />
          <el-option label="折扣活动" value="discount" />
          <el-option label="阶梯满减" value="tiered" />
        </el-select>
        <el-select v-model="query.status" placeholder="状态" clearable style="width: 140px">
          <el-option label="禁用" :value="0" />
          <el-option label="启用" :value="1" />
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
        <el-table-column label="活动类型" width="120">
          <template #default="{ row }">
            <el-tag :type="getTypeTagType(row.type)">{{ getTypeName(row.type) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="活动规则" min-width="200">
          <template #default="{ row }">
            <div v-if="row.type === 'full_reduction'" class="rule-text">
              满 ¥{{ row.min_amount }} 减 ¥{{ row.discount_value }}
            </div>
            <div v-else-if="row.type === 'discount'" class="rule-text">
              打 {{ row.discount_value }} 折
            </div>
            <div v-else-if="row.type === 'tiered'" class="rule-text">
              阶梯优惠 ({{ row.tiers?.length || 0 }} 档)
            </div>
          </template>
        </el-table-column>
        <el-table-column label="活动时间" width="240">
          <template #default="{ row }">
            <div class="date-range">
              <div>{{ formatDate(row.start_time) }}</div>
              <div class="divider">至</div>
              <div>{{ formatDate(row.end_time) }}</div>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="适用范围" width="100">
          <template #default="{ row }">
            <el-tag size="small">{{ getApplicableName(row.applicable_type) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="priority" label="优先级" width="80" align="center" />
        <el-table-column label="是否可叠加" width="100" align="center">
          <template #default="{ row }">
            <el-tag :type="row.stackable ? 'success' : 'info'" size="small">
              {{ row.stackable ? '可叠加' : '不可叠加' }}
            </el-tag>
          </template>
        </el-table-column>
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
      :title="isEdit ? '编辑活动' : '新增活动'"
      width="700px"
      :close-on-click-modal="false">
      <el-form
        ref="promotionFormRef"
        :model="promotionForm"
        :rules="promotionRules"
        label-width="120px">
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="活动名称" prop="name">
              <el-input v-model="promotionForm.name" placeholder="请输入活动名称" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="活动类型" prop="type">
              <el-select v-model="promotionForm.type" placeholder="请选择类型" style="width: 100%" @change="handleTypeChange">
                <el-option label="满减活动" value="full_reduction" />
                <el-option label="折扣活动" value="discount" />
                <el-option label="阶梯满减" value="tiered" />
              </el-select>
            </el-form-item>
          </el-col>
        </el-row>

        <el-form-item v-if="promotionForm.type === 'full_reduction'" label="满减规则">
          <el-row :gutter="20">
            <el-col :span="12">
              <el-form-item prop="min_amount" label-width="0">
                <el-input-number
                  v-model="promotionForm.min_amount"
                  :min="0"
                  :precision="2"
                  style="width: 100%" />
                <span style="margin-left: 8px;">元</span>
              </el-form-item>
            </el-col>
            <el-col :span="12">
              <el-form-item prop="discount_value" label-width="0">
                <span style="margin-right: 8px;">减</span>
                <el-input-number
                  v-model="promotionForm.discount_value"
                  :min="0"
                  :precision="2"
                  style="width: 100%" />
                <span style="margin-left: 8px;">元</span>
              </el-form-item>
            </el-col>
          </el-row>
        </el-form-item>

        <el-form-item v-if="promotionForm.type === 'discount'" label="折扣值" prop="discount_value">
          <el-input-number
            v-model="promotionForm.discount_value"
            :min="0.1"
            :max="9.9"
            :precision="1"
            :step="0.1" />
          <span style="margin-left: 8px;">折</span>
        </el-form-item>

        <el-form-item v-if="promotionForm.type === 'discount'" label="最高优惠">
          <el-input-number
            v-model="promotionForm.max_discount"
            :min="0"
            :precision="2" />
          <span style="margin-left: 8px;">元（0为不限制）</span>
        </el-form-item>

        <el-form-item v-if="promotionForm.type === 'tiered'" label="阶梯优惠">
          <div class="tier-list">
            <div v-for="(tier, index) in promotionForm.tiers" :key="index" class="tier-item">
              <span>第 {{ index + 1 }} 档：</span>
              <el-input-number
                v-model="tier.min_amount"
                :min="0"
                :precision="2"
                size="small" />
              <span style="margin: 0 4px;">元减</span>
              <el-input-number
                v-model="tier.discount_value"
                :min="0"
                :precision="2"
                size="small" />
              <span style="margin: 0 4px;">元</span>
              <el-button
                v-if="promotionForm.tiers.length > 1"
                type="danger"
                link
                size="small"
                @click="removeTier(index)">
                删除
              </el-button>
            </div>
            <el-button type="primary" link size="small" @click="addTier">+ 添加档位</el-button>
          </div>
        </el-form-item>

        <el-form-item label="活动时间" prop="time_range">
          <el-date-picker
            v-model="promotionForm.time_range"
            type="datetimerange"
            range-separator="至"
            start-placeholder="开始时间"
            end-placeholder="结束时间"
            value-format="YYYY-MM-DD HH:mm:ss"
            style="width: 100%" />
        </el-form-item>

        <el-form-item label="适用范围" prop="applicable_type">
          <el-radio-group v-model="promotionForm.applicable_type">
            <el-radio value="all">全部商品</el-radio>
            <el-radio value="category">指定分类</el-radio>
            <el-radio value="product">指定商品</el-radio>
          </el-radio-group>
        </el-form-item>

        <el-form-item v-if="promotionForm.applicable_type !== 'all'" label="适用ID列表">
          <el-select
            v-model="promotionForm.applicable_ids"
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

        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="优先级">
              <el-input-number
                v-model="promotionForm.priority"
                :min="1"
                :max="999" />
              <span style="margin-left: 8px; color: #909399;">数字越小优先级越高</span>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="是否可叠加">
              <el-switch v-model="promotionForm.stackable" active-text="可叠加" inactive-text="不可叠加" />
            </el-form-item>
          </el-col>
        </el-row>

        <el-form-item label="活动描述">
          <el-input
            v-model="promotionForm.description"
            type="textarea"
            :rows="2"
            placeholder="请输入活动描述" />
        </el-form-item>

        <el-form-item label="状态">
          <el-switch v-model="promotionForm.status" :active-value="1" :inactive-value="0" active-text="启用" inactive-text="禁用" />
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
  getPromotionList,
  getPromotion,
  createPromotion,
  updatePromotion,
  deletePromotion
} from '@/api/promotions'
import { getProductList } from '@/api/product'

const loading = ref(false)
const submitLoading = ref(false)
const list = ref([])
const total = ref(0)
const productList = ref([])

const query = reactive({
  name: '',
  type: '',
  status: null,
  page: 1,
  page_size: 10
})

const dialogVisible = ref(false)
const isEdit = ref(false)
const editId = ref(null)
const promotionFormRef = ref()

const defaultForm = () => ({
  name: '',
  type: 'full_reduction',
  min_amount: 0,
  discount_value: 0,
  max_discount: 0,
  tiers: [{ min_amount: 0, discount_value: 0 }],
  time_range: [],
  start_time: '',
  end_time: '',
  applicable_type: 'all',
  applicable_ids: [],
  priority: 100,
  stackable: false,
  description: '',
  status: 1
})

const promotionForm = reactive(defaultForm())

const promotionRules = {
  name: [{ required: true, message: '请输入活动名称', trigger: 'blur' }],
  type: [{ required: true, message: '请选择活动类型', trigger: 'change' }],
  time_range: [{ required: true, message: '请选择活动时间', trigger: 'change' }],
  applicable_type: [{ required: true, message: '请选择适用范围', trigger: 'change' }]
}

const typeMap = {
  full_reduction: { name: '满减活动', type: 'danger' },
  discount: { name: '折扣活动', type: 'success' },
  tiered: { name: '阶梯满减', type: 'warning' }
}

const statusMap = {
  0: { name: '禁用', type: 'info' },
  1: { name: '启用', type: 'success' },
  2: { name: '已结束', type: 'info' }
}

const applicableMap = {
  all: '全部商品',
  category: '指定分类',
  product: '指定商品'
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

function getApplicableName(type) {
  return applicableMap[type] || '未知'
}

function formatDate(date) {
  if (!date) return '-'
  return new Date(date).toLocaleString()
}

function handleTypeChange() {
  if (promotionForm.type === 'full_reduction') {
    if (promotionForm.tiers?.length) {
      promotionForm.min_amount = promotionForm.tiers[0]?.min_amount || 0
      promotionForm.discount_value = promotionForm.tiers[0]?.discount_value || 0
    }
  }
}

function addTier() {
  promotionForm.tiers.push({ min_amount: 0, discount_value: 0 })
}

function removeTier(index) {
  promotionForm.tiers.splice(index, 1)
}

function parseApplicableIds(ids) {
  if (!ids) return []
  if (Array.isArray(ids)) return ids
  return String(ids).split(',').map(id => parseInt(id)).filter(id => !isNaN(id))
}

async function fetchList() {
  loading.value = true
  try {
    const res = await getPromotionList(query)
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
  query.type = ''
  query.status = null
  query.page = 1
  fetchList()
}

function openDialog(row = null) {
  isEdit.value = !!row
  editId.value = row?.id || null
  fetchProducts()

  Object.assign(promotionForm, defaultForm())

  if (row) {
    Object.assign(promotionForm, {
      name: row.name,
      type: row.type,
      min_amount: row.min_amount || 0,
      discount_value: row.discount_value || 0,
      max_discount: row.max_discount || 0,
      tiers: row.tiers?.length ? row.tiers : [{ min_amount: 0, discount_value: 0 }],
      time_range: row.start_time && row.end_time ? [row.start_time, row.end_time] : [],
      start_time: row.start_time,
      end_time: row.end_time,
      applicable_type: row.applicable_type || 'all',
      applicable_ids: parseApplicableIds(row.applicable_ids),
      priority: row.priority || 100,
      stackable: row.stackable || false,
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
  ElMessageBox.confirm(`确定删除活动"${row.name}"吗？`, '提示', {
    type: 'warning',
    confirmButtonText: '确定',
    cancelButtonText: '取消'
  }).then(async () => {
    await deletePromotion(row.id)
    ElMessage.success('删除成功')
    fetchList()
  })
}

async function handleSubmit() {
  try {
    await promotionFormRef.value.validate()
    submitLoading.value = true

    const data = {
      ...promotionForm
    }

    if (promotionForm.time_range?.length === 2) {
      data.start_time = promotionForm.time_range[0]
      data.end_time = promotionForm.time_range[1]
    }
    delete data.time_range

    if (isEdit.value) {
      await updatePromotion(editId.value, data)
      ElMessage.success('更新成功')
    } else {
      await createPromotion(data)
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
.promotions-page {
  .header-actions {
    display: flex;
    gap: 12px;
  }

  .rule-text {
    font-size: 14px;
    color: #303133;
    font-weight: 500;
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

  .tier-list {
    width: 100%;

    .tier-item {
      display: flex;
      align-items: center;
      gap: 8px;
      margin-bottom: 12px;
    }
  }
}
</style>
