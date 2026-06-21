<template>
  <div class="reservations-page">
    <el-tabs v-model="activeTab" class="main-tabs">
      <el-tab-pane label="桌位预约" name="reservation">
        <div class="page-header">
          <h2 class="page-title">预约管理</h2>
          <div class="header-actions">
            <el-button
              type="warning"
              :disabled="selectedIds.length === 0"
              @click="handleBatchConfirm">
              <el-icon><Check /></el-icon>批量确认
            </el-button>
            <el-button
              type="danger"
              :disabled="selectedIds.length === 0"
              @click="handleBatchCancel">
              <el-icon><Close /></el-icon>批量取消
            </el-button>
            <el-button type="success" @click="openDialog">
              <el-icon><Plus /></el-icon>新增预约
            </el-button>
          </div>
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
            <el-select v-model="query.status" placeholder="预约状态" clearable style="width: 140px">
              <el-option label="待确认" :value="1" />
              <el-option label="已确认" :value="2" />
              <el-option label="已取消" :value="3" />
              <el-option label="已完成" :value="4" />
            </el-select>
            <el-date-picker
              v-model="query.reservation_date"
              type="date"
              placeholder="预约日期"
              value-format="YYYY-MM-DD"
              style="width: 160px" />
            <el-input
              v-model="query.keyword"
              placeholder="搜索姓名/电话/桌号"
              clearable
              style="width: 240px"
              @keyup.enter="fetchList" />
            <el-button type="primary" @click="fetchList">
              <el-icon><Search /></el-icon>搜索
            </el-button>
            <el-button @click="resetQuery">
              <el-icon><Refresh /></el-icon>重置
            </el-button>
          </div>

          <el-table
            :data="list"
            v-loading="loading"
            @selection-change="handleSelectionChange">
            <el-table-column type="selection" width="55" />
            <el-table-column prop="id" label="预约单号" width="100" />
            <el-table-column prop="member_name" label="会员姓名" width="120" />
            <el-table-column prop="member_phone" label="会员电话" width="140" />
            <el-table-column prop="table_name" label="桌号" width="100" />
            <el-table-column prop="reservation_date" label="预约日期" width="120" />
            <el-table-column prop="reservation_time" label="预约时间" width="100" />
            <el-table-column prop="people_count" label="人数" width="80" align="center" />
            <el-table-column label="状态" width="100" align="center">
              <template #default="{ row }">
                <el-tag :type="getStatusTagType(row.status)" effect="light">
                  {{ getStatusName(row.status) }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="created_at" label="创建时间" width="160" />
            <el-table-column label="操作" width="200" fixed="right">
              <template #default="{ row }">
                <el-button
                  v-if="row.status === 1"
                  type="success"
                  link
                  size="small"
                  @click="handleConfirm(row)">
                  确认
                </el-button>
                <el-button
                  v-if="row.status === 1 || row.status === 2"
                  type="danger"
                  link
                  size="small"
                  @click="handleCancel(row)">
                  取消
                </el-button>
                <el-button
                  v-if="row.status === 2"
                  type="primary"
                  link
                  size="small"
                  @click="handleCheckin(row)">
                  核销
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
          title="新增预约"
          width="600px"
          :close-on-click-modal="false">
          <el-form
            ref="reservationFormRef"
            :model="reservationForm"
            :rules="reservationRules"
            label-width="100px">
            <el-row :gutter="20">
              <el-col :span="12">
                <el-form-item label="门店" prop="store_id">
                  <el-select
                    v-model="reservationForm.store_id"
                    placeholder="请选择门店"
                    style="width: 100%"
                    @change="handleStoreChange">
                    <el-option
                      v-for="store in storeList"
                      :key="store.id"
                      :label="store.name"
                      :value="store.id" />
                  </el-select>
                </el-form-item>
              </el-col>
              <el-col :span="12">
                <el-form-item label="桌号" prop="table_id">
                  <el-select
                    v-model="reservationForm.table_id"
                    placeholder="请选择桌号"
                    style="width: 100%">
                    <el-option
                      v-for="table in tableList"
                      :key="table.id"
                      :label="table.name"
                      :value="table.id" />
                  </el-select>
                </el-form-item>
              </el-col>
            </el-row>
            <el-row :gutter="20">
              <el-col :span="12">
                <el-form-item label="姓名" prop="member_name">
                  <el-input v-model="reservationForm.member_name" placeholder="请输入姓名" />
                </el-form-item>
              </el-col>
              <el-col :span="12">
                <el-form-item label="电话" prop="member_phone">
                  <el-input v-model="reservationForm.member_phone" placeholder="请输入电话" />
                </el-form-item>
              </el-col>
            </el-row>
            <el-row :gutter="20">
              <el-col :span="12">
                <el-form-item label="预约日期" prop="reservation_date">
                  <el-date-picker
                    v-model="reservationForm.reservation_date"
                    type="date"
                    placeholder="选择日期"
                    value-format="YYYY-MM-DD"
                    style="width: 100%" />
                </el-form-item>
              </el-col>
              <el-col :span="12">
                <el-form-item label="预约时间" prop="reservation_time">
                  <el-time-picker
                    v-model="reservationForm.reservation_time"
                    placeholder="选择时间"
                    value-format="HH:mm"
                    style="width: 100%" />
                </el-form-item>
              </el-col>
            </el-row>
            <el-form-item label="人数" prop="people_count">
              <el-input-number
                v-model="reservationForm.people_count"
                :min="1"
                :max="50"
                style="width: 200px" />
            </el-form-item>
            <el-form-item label="备注">
              <el-input
                v-model="reservationForm.remark"
                type="textarea"
                :rows="2"
                placeholder="请输入备注" />
            </el-form-item>
          </el-form>
          <template #footer>
            <el-button @click="dialogVisible = false">取消</el-button>
            <el-button type="primary" :loading="submitLoading" @click="handleSubmit">确定</el-button>
          </template>
        </el-dialog>
      </el-tab-pane>

      <el-tab-pane label="预约订单" name="order">
        <div class="page-header">
          <h2 class="page-title">预约订单管理</h2>
        </div>

        <div class="card-wrapper">
          <div class="search-bar">
            <el-date-picker
              v-model="orderQuery.reservation_time_range"
              type="daterange"
              range-separator="至"
              start-placeholder="开始日期"
              end-placeholder="结束日期"
              value-format="YYYY-MM-DD"
              style="width: 280px" />
            <el-select v-model="orderQuery.status" placeholder="订单状态" clearable style="width: 140px">
              <el-option label="待支付" :value="1" />
              <el-option label="已支付" :value="2" />
              <el-option label="已完成" :value="3" />
              <el-option label="已取消" :value="4" />
              <el-option label="已退款" :value="5" />
            </el-select>
            <el-select v-model="orderQuery.is_reservation" placeholder="是否预约订单" clearable style="width: 160px">
              <el-option label="是" :value="1" />
              <el-option label="否" :value="0" />
            </el-select>
            <el-input
              v-model="orderQuery.keyword"
              placeholder="搜索订单号/会员名称"
              clearable
              style="width: 240px"
              @keyup.enter="fetchOrderList" />
            <el-button type="primary" @click="fetchOrderList">
              <el-icon><Search /></el-icon>搜索
            </el-button>
            <el-button @click="resetOrderQuery">
              <el-icon><Refresh /></el-icon>重置
            </el-button>
          </div>

          <el-table
            :data="orderList"
            v-loading="orderLoading">
            <el-table-column prop="order_no" label="订单号" width="160" />
            <el-table-column prop="store_name" label="门店名称" width="140" />
            <el-table-column prop="member_name" label="会员名称" width="120" />
            <el-table-column prop="reservation_time" label="预约时间" width="160" />
            <el-table-column label="订单金额" width="120" align="center">
              <template #default="{ row }">
                <span class="price">¥{{ (row.total_amount || 0).toFixed(2) }}</span>
              </template>
            </el-table-column>
            <el-table-column label="订单状态" width="100" align="center">
              <template #default="{ row }">
                <el-tag :type="getOrderStatusTagType(row.status)" effect="light">
                  {{ getOrderStatusName(row.status) }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column label="支付状态" width="100" align="center">
              <template #default="{ row }">
                <el-tag :type="getPayStatusTagType(row.pay_status)" effect="light">
                  {{ getPayStatusName(row.pay_status) }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="created_at" label="创建时间" width="160" />
            <el-table-column label="操作" width="240" fixed="right">
              <template #default="{ row }">
                <el-button type="primary" link size="small" @click="handleViewOrderDetail(row)">
                  查看详情
                </el-button>
                <el-button
                  v-if="row.status === 1 || row.status === 2"
                  type="danger"
                  link
                  size="small"
                  @click="handleCancelOrder(row)">
                  取消预约
                </el-button>
                <el-button
                  v-if="row.status === 2"
                  type="success"
                  link
                  size="small"
                  @click="handleMarkOrderComplete(row)">
                  标记已完成
                </el-button>
              </template>
            </el-table-column>
          </el-table>

          <div class="pagination">
            <el-pagination
              v-model:current-page="orderQuery.page"
              v-model:page-size="orderQuery.page_size"
              :total="orderTotal"
              :page-sizes="[10, 20, 50, 100]"
              layout="total, sizes, prev, pager, next, jumper"
              @size-change="fetchOrderList"
              @current-change="fetchOrderList" />
          </div>
        </div>

        <el-dialog
          v-model="orderDetailVisible"
          title="订单详情"
          width="700px"
          :close-on-click-modal="false">
          <div v-if="currentOrderDetail" class="order-detail">
            <el-descriptions :column="2" border>
              <el-descriptions-item label="订单号">
                {{ currentOrderDetail.order_no }}
              </el-descriptions-item>
              <el-descriptions-item label="门店名称">
                {{ currentOrderDetail.store_name }}
              </el-descriptions-item>
              <el-descriptions-item label="会员名称">
                {{ currentOrderDetail.member_name }}
              </el-descriptions-item>
              <el-descriptions-item label="预约时间">
                {{ currentOrderDetail.reservation_time }}
              </el-descriptions-item>
              <el-descriptions-item label="订单金额">
                <span class="price">¥{{ (currentOrderDetail.total_amount || 0).toFixed(2) }}</span>
              </el-descriptions-item>
              <el-descriptions-item label="订单状态">
                <el-tag :type="getOrderStatusTagType(currentOrderDetail.status)" effect="light">
                  {{ getOrderStatusName(currentOrderDetail.status) }}
                </el-tag>
              </el-descriptions-item>
              <el-descriptions-item label="支付状态">
                <el-tag :type="getPayStatusTagType(currentOrderDetail.pay_status)" effect="light">
                  {{ getPayStatusName(currentOrderDetail.pay_status) }}
                </el-tag>
              </el-descriptions-item>
              <el-descriptions-item label="创建时间">
                {{ currentOrderDetail.created_at }}
              </el-descriptions-item>
            </el-descriptions>

            <div class="detail-section">
              <h4 class="section-title">商品明细</h4>
              <el-table :data="currentOrderDetail.items || []" size="small">
                <el-table-column prop="product_name" label="商品名称" />
                <el-table-column prop="quantity" label="数量" width="80" align="center" />
                <el-table-column label="单价" width="100" align="center">
                  <template #default="{ row }">
                    ¥{{ (row.price || 0).toFixed(2) }}
                  </template>
                </el-table-column>
                <el-table-column label="小计" width="100" align="center">
                  <template #default="{ row }">
                    ¥{{ ((row.price || 0) * (row.quantity || 0)).toFixed(2) }}
                  </template>
                </el-table-column>
              </el-table>
            </div>

            <div v-if="currentOrderDetail.remark" class="detail-section">
              <h4 class="section-title">备注</h4>
              <p>{{ currentOrderDetail.remark }}</p>
            </div>
          </div>
          <template #footer>
            <el-button @click="orderDetailVisible = false">关闭</el-button>
          </template>
        </el-dialog>
      </el-tab-pane>
    </el-tabs>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Search, Refresh, Check, Close } from '@element-plus/icons-vue'
import { reservationApi, tableApi } from '@/api/tables'
import { storeApi } from '@/api/stores'
import { getOrderList, getOrderDetail, updateOrderStatus, cancelOrder } from '@/api/orders'

const activeTab = ref('reservation')

const loading = ref(false)
const submitLoading = ref(false)
const list = ref([])
const total = ref(0)
const storeList = ref([])
const tableList = ref([])
const selectedIds = ref([])

const query = reactive({
  store_id: null,
  status: null,
  reservation_date: '',
  keyword: '',
  page: 1,
  page_size: 10
})

const dialogVisible = ref(false)
const reservationFormRef = ref()

const reservationForm = reactive({
  store_id: null,
  table_id: null,
  member_name: '',
  member_phone: '',
  reservation_date: '',
  reservation_time: '',
  people_count: 2,
  remark: ''
})

const reservationRules = {
  store_id: [{ required: true, message: '请选择门店', trigger: 'change' }],
  table_id: [{ required: true, message: '请选择桌号', trigger: 'change' }],
  member_name: [{ required: true, message: '请输入姓名', trigger: 'blur' }],
  member_phone: [{ required: true, message: '请输入电话', trigger: 'blur' }],
  reservation_date: [{ required: true, message: '请选择预约日期', trigger: 'change' }],
  reservation_time: [{ required: true, message: '请选择预约时间', trigger: 'change' }],
  people_count: [{ required: true, message: '请输入人数', trigger: 'blur' }]
}

const statusMap = {
  1: { name: '待确认', type: 'warning' },
  2: { name: '已确认', type: 'success' },
  3: { name: '已取消', type: 'info' },
  4: { name: '已完成', type: 'primary' }
}

function getStatusName(status) {
  return statusMap[status]?.name || '未知'
}

function getStatusTagType(status) {
  return statusMap[status]?.type || 'info'
}

async function fetchStoreList() {
  try {
    const res = await storeApi.list({ page: 1, page_size: 100 })
    storeList.value = res.list || []
  } catch (e) {
    console.error(e)
  }
}

async function fetchTableList(storeId = null) {
  try {
    const params = { page: 1, page_size: 100 }
    if (storeId) {
      params.store_id = storeId
    }
    const res = await tableApi.list(params)
    tableList.value = res.list || []
  } catch (e) {
    console.error(e)
  }
}

async function fetchList() {
  loading.value = true
  try {
    const params = { ...query }
    if (!params.reservation_date) {
      delete params.reservation_date
    }
    const res = await reservationApi.list(params)
    list.value = res.list || []
    total.value = res.total || 0
  } catch (e) {
    console.error(e)
  } finally {
    loading.value = false
  }
}

function resetQuery() {
  query.store_id = null
  query.status = null
  query.reservation_date = ''
  query.keyword = ''
  query.page = 1
  fetchList()
}

function handleSelectionChange(selection) {
  selectedIds.value = selection
    .filter(item => item.status === 1 || item.status === 2)
    .map(item => item.id)
}

function handleStoreChange(storeId) {
  reservationForm.table_id = null
  fetchTableList(storeId)
}

function openDialog() {
  reservationForm.store_id = null
  reservationForm.table_id = null
  reservationForm.member_name = ''
  reservationForm.member_phone = ''
  reservationForm.reservation_date = ''
  reservationForm.reservation_time = ''
  reservationForm.people_count = 2
  reservationForm.remark = ''
  tableList.value = []
  dialogVisible.value = true
}

async function handleSubmit() {
  try {
    await reservationFormRef.value.validate()
    submitLoading.value = true
    await reservationApi.create(reservationForm)
    ElMessage.success('创建成功')
    dialogVisible.value = false
    fetchList()
  } catch (e) {
    console.error(e)
  } finally {
    submitLoading.value = false
  }
}

async function handleConfirm(row) {
  ElMessageBox.confirm(`确定确认预约"${row.id}"吗？`, '提示', {
    type: 'warning',
    confirmButtonText: '确定',
    cancelButtonText: '取消'
  }).then(async () => {
    try {
      await reservationApi.update(row.id, { status: 2 })
      ElMessage.success('确认成功')
      fetchList()
    } catch (e) {
      console.error(e)
    }
  })
}

async function handleCancel(row) {
  ElMessageBox.confirm(`确定取消预约"${row.id}"吗？`, '提示', {
    type: 'warning',
    confirmButtonText: '确定',
    cancelButtonText: '取消'
  }).then(async () => {
    try {
      await reservationApi.cancel(row.id)
      ElMessage.success('取消成功')
      fetchList()
    } catch (e) {
      console.error(e)
    }
  })
}

async function handleCheckin(row) {
  ElMessageBox.confirm(`确定核销预约"${row.id}"吗？`, '提示', {
    type: 'warning',
    confirmButtonText: '确定',
    cancelButtonText: '取消'
  }).then(async () => {
    try {
      await reservationApi.checkin(row.id)
      ElMessage.success('核销成功')
      fetchList()
    } catch (e) {
      console.error(e)
    }
  })
}

async function handleBatchConfirm() {
  const confirmIds = selectedIds.value.filter(id => {
    const item = list.value.find(i => i.id === id)
    return item && item.status === 1
  })
  if (confirmIds.length === 0) {
    ElMessage.warning('请选择待确认的预约')
    return
  }
  ElMessageBox.confirm(`确定批量确认选中的 ${confirmIds.length} 条预约吗？`, '提示', {
    type: 'warning',
    confirmButtonText: '确定',
    cancelButtonText: '取消'
  }).then(async () => {
    try {
      loading.value = true
      await Promise.all(confirmIds.map(id => reservationApi.update(id, { status: 2 })))
      ElMessage.success('批量确认成功')
      selectedIds.value = []
      fetchList()
    } catch (e) {
      console.error(e)
    } finally {
      loading.value = false
    }
  })
}

async function handleBatchCancel() {
  if (selectedIds.value.length === 0) {
    ElMessage.warning('请选择要取消的预约')
    return
  }
  ElMessageBox.confirm(`确定批量取消选中的 ${selectedIds.value.length} 条预约吗？`, '提示', {
    type: 'warning',
    confirmButtonText: '确定',
    cancelButtonText: '取消'
  }).then(async () => {
    try {
      loading.value = true
      await Promise.all(selectedIds.value.map(id => reservationApi.cancel(id)))
      ElMessage.success('批量取消成功')
      selectedIds.value = []
      fetchList()
    } catch (e) {
      console.error(e)
    } finally {
      loading.value = false
    }
  })
}

const orderLoading = ref(false)
const orderList = ref([])
const orderTotal = ref(0)
const orderDetailVisible = ref(false)
const currentOrderDetail = ref(null)

const orderQuery = reactive({
  reservation_time_range: [],
  status: null,
  is_reservation: 1,
  keyword: '',
  page: 1,
  page_size: 10
})

const orderStatusMap = {
  1: { name: '待支付', type: 'warning' },
  2: { name: '已支付', type: 'primary' },
  3: { name: '已完成', type: 'success' },
  4: { name: '已取消', type: 'info' },
  5: { name: '已退款', type: 'danger' }
}

const payStatusMap = {
  0: { name: '未支付', type: 'info' },
  1: { name: '已支付', type: 'success' },
  2: { name: '已退款', type: 'danger' }
}

function getOrderStatusName(status) {
  return orderStatusMap[status]?.name || '未知'
}

function getOrderStatusTagType(status) {
  return orderStatusMap[status]?.type || 'info'
}

function getPayStatusName(status) {
  return payStatusMap[status]?.name || '未知'
}

function getPayStatusTagType(status) {
  return payStatusMap[status]?.type || 'info'
}

async function fetchOrderList() {
  orderLoading.value = true
  try {
    const params = { ...orderQuery }
    if (params.reservation_time_range && params.reservation_time_range.length === 2) {
      params.reservation_start = params.reservation_time_range[0]
      params.reservation_end = params.reservation_time_range[1]
    }
    delete params.reservation_time_range
    const res = await getOrderList(params)
    orderList.value = res.list || []
    orderTotal.value = res.total || 0
  } catch (e) {
    console.error(e)
  } finally {
    orderLoading.value = false
  }
}

function resetOrderQuery() {
  orderQuery.reservation_time_range = []
  orderQuery.status = null
  orderQuery.is_reservation = 1
  orderQuery.keyword = ''
  orderQuery.page = 1
  fetchOrderList()
}

async function handleViewOrderDetail(row) {
  try {
    const res = await getOrderDetail(row.id)
    currentOrderDetail.value = res
    orderDetailVisible.value = true
  } catch (e) {
    console.error(e)
  }
}

async function handleCancelOrder(row) {
  ElMessageBox.confirm(`确定取消订单"${row.order_no}"吗？`, '提示', {
    type: 'warning',
    confirmButtonText: '确定',
    cancelButtonText: '取消'
  }).then(async () => {
    try {
      await cancelOrder(row.id, '管理员取消')
      ElMessage.success('取消成功')
      fetchOrderList()
    } catch (e) {
      console.error(e)
    }
  })
}

async function handleMarkOrderComplete(row) {
  ElMessageBox.confirm(`确定将订单"${row.order_no}"标记为已完成吗？`, '提示', {
    type: 'warning',
    confirmButtonText: '确定',
    cancelButtonText: '取消'
  }).then(async () => {
    try {
      await updateOrderStatus(row.id, 3)
      ElMessage.success('标记成功')
      fetchOrderList()
    } catch (e) {
      console.error(e)
    }
  })
}

onMounted(() => {
  fetchStoreList()
  fetchList()
})
</script>

<style scoped lang="scss">
.reservations-page {
  .main-tabs {
    :deep(.el-tabs__header) {
      margin-bottom: 0;
    }
  }

  .header-actions {
    display: flex;
    gap: 12px;
  }

  .pagination {
    margin-top: 20px;
    display: flex;
    justify-content: flex-end;
  }

  .price {
    color: #f56c6c;
    font-weight: 600;
  }

  .order-detail {
    .detail-section {
      margin-top: 20px;

      .section-title {
        margin: 0 0 12px 0;
        font-size: 14px;
        font-weight: 600;
        color: #303133;
      }
    }
  }
}
</style>
