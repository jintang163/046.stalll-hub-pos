<template>
  <div class="transfers-page">
    <div class="page-header">
      <h2>库存调拨</h2>
      <div class="header-actions">
        <el-button type="primary" @click="handleCreate">
          <el-icon><Plus /></el-icon>
          发起调拨
        </el-button>
      </div>
    </div>

    <el-card class="search-card">
      <el-form :inline="true" :model="searchForm">
        <el-form-item label="调出门店">
          <el-select v-model="searchForm.from_store_id" placeholder="全部门店" clearable style="width: 150px">
            <el-option v-for="store in storeList" :key="store.id" :label="store.name" :value="store.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="调入门店">
          <el-select v-model="searchForm.to_store_id" placeholder="全部门店" clearable style="width: 150px">
            <el-option v-for="store in storeList" :key="store.id" :label="store.name" :value="store.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="searchForm.status" placeholder="全部" style="width: 130px">
            <el-option label="全部" :value="-1" />
            <el-option label="待接单" :value="0" />
            <el-option label="待出库" :value="1" />
            <el-option label="已出库" :value="2" />
            <el-option label="运输中" :value="3" />
            <el-option label="已收货" :value="4" />
            <el-option label="已完成" :value="5" />
            <el-option label="已取消" :value="6" />
          </el-select>
        </el-form-item>
        <el-form-item label="调拨单号">
          <el-input v-model="searchForm.keyword" placeholder="输入单号" clearable style="width: 180px" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="handleSearch">搜索</el-button>
          <el-button @click="handleReset">重置</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <el-card class="table-card">
      <el-table :data="tableData" v-loading="loading" border stripe>
        <el-table-column prop="transfer_no" label="调拨单号" width="180" />
        <el-table-column label="调出门店" width="120">
          <template #default="{ row }">
            {{ row.from_store?.name || '-' }}
          </template>
        </el-table-column>
        <el-table-column label="调入门店" width="120">
          <template #default="{ row }">
            {{ row.to_store?.name || '-' }}
          </template>
        </el-table-column>
        <el-table-column prop="total_qty" label="总数量" width="100">
          <template #default="{ row }">
          {{ row.total_qty }}
          </template>
        </el-table-column>
        <el-table-column prop="total_amount" label="总金额" width="120">
          <template #default="{ row }">
            ¥{{ row.total_amount }}
          </template>
        </el-table-column>
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="getStatusType(row.status)">
              {{ getStatusText(row.status) }}
            </el-tag>
            <el-tag v-if="row.has_diff" type="warning" size="small" style="margin-left: 4px">差异</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="物流信息" width="160">
          <template #default="{ row }">
            <div v-if="row.tracking_no">
              <div>{{ row.logistics_company }}</div>
              <div style="font-size: 12px; color: #909399">{{ row.tracking_no }}</div>
            </div>
            <span v-else style="color: #c0c4cc">-</span>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="创建时间" width="160" />
        <el-table-column label="操作" width="340" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="viewDetail(row)">详情</el-button>
            <el-button link type="primary" @click="viewLogistics(row)" v-if="row.tracking_no">物流</el-button>
            <el-button link type="success" @click="handleAccept(row)" v-if="row.status === 0">接单</el-button>
            <el-button link type="danger" @click="handleReject(row)" v-if="row.status === 0">拒单</el-button>
            <el-button link type="primary" @click="handleOutbound(row)" v-if="row.status === 1">出库</el-button>
            <el-button link type="primary" @click="handleShip(row)" v-if="row.status === 2">发货</el-button>
            <el-button link type="primary" @click="handleReceive(row)" v-if="row.status === 2 || row.status === 3">收货</el-button>
            <el-button link type="primary" @click="handleComplete(row)" v-if="row.status === 4">完成</el-button>
            <el-button link type="danger" @click="handleCancel(row)" v-if="row.status === 0 || row.status === 1">取消</el-button>
          </template>
        </el-table-column>
      </el-table>

      <el-pagination
        v-model:current-page="pagination.page"
        v-model:page-size="pagination.pageSize"
        :total="pagination.total"
        :page-sizes="[10, 20, 50, 100]"
        layout="total, sizes, prev, pager, next, jumper"
        class="pagination"
        @size-change="fetchList"
        @current-change="fetchList"
      />
    </el-card>

    <el-dialog v-model="createVisible" title="发起调拨" width="800px" top="5vh">
      <el-form :model="transferForm" label-width="100px">
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="调出门店">
              <el-select v-model="transferForm.from_store_id" placeholder="请选择调出门店" style="width: 100%" @change="loadIngredients">
                <el-option v-for="store in storeList" :key="store.id" :label="store.name" :value="store.id" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="调入门店">
              <el-select v-model="transferForm.to_store_id" placeholder="请选择调入门店" style="width: 100%">
                <el-option v-for="store in storeList" :key="store.id" :label="store.name" :value="store.id" />
              </el-select>
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="优先级">
              <el-select v-model="transferForm.priority" style="width: 100%">
                <el-option label="普通" value="normal" />
                <el-option label="紧急" value="urgent" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="调拨类型">
              <el-select v-model="transferForm.transfer_type" style="width: 100%">
                <el-option label="常规调拨" value="normal" />
                <el-option label="紧急调拨" value="emergency" />
              </el-select>
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="发货人">
              <el-input v-model="transferForm.sender_name" placeholder="请输入发货人姓名" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="发货电话">
              <el-input v-model="transferForm.sender_phone" placeholder="请输入发货人电话" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="收货人">
              <el-input v-model="transferForm.receiver_name" placeholder="请输入收货人姓名" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="收货电话">
              <el-input v-model="transferForm.receiver_phone" placeholder="请输入收货人电话" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-form-item label="收货地址">
          <el-input v-model="transferForm.receiver_address" placeholder="请输入收货地址" />
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="transferForm.remark" type="textarea" :rows="2" placeholder="请输入备注" />
        </el-form-item>

        <el-form-item label="调拨明细">
          <div class="item-list-wrapper">
            <el-table :data="transferForm.items" border size="small">
              <el-table-column prop="ingredient_name" label="食材名称" min-width="120" />
              <el-table-column prop="unit" label="单位" width="70" />
              <el-table-column label="调出数量" width="140">
                <template #default="{ row, $index }">
                  <el-input-number v-model="row.out_qty" :precision="2" :min="0" size="small" style="width: 100%" />
                </template>
              </el-table-column>
              <el-table-column prop="unit_price" label="单价" width="100">
                <template #default="{ row }">
                  ¥{{ row.unit_price }}
                </template>
              </el-table-column>
              <el-table-column label="操作" width="80">
                <template #default="{ $index }">
                  <el-button link type="danger" @click="removeItem($index)">删除</el-button>
                </template>
              </el-table-column>
            </el-table>
            <div class="add-item-btn">
              <el-select v-model="selectedIngredient" placeholder="添加食材" filterable style="width: 300px" @change="addIngredient">
                <el-option
                  v-for="ing in availableIngredients"
                  :key="ing.id"
                  :label="`${ing.name} (库存: ${ing.current_stock}${ing.unit})`"
                  :value="ing.id"
                />
              </el-select>
            </div>
          </div>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="createVisible = false">取消</el-button>
        <el-button type="primary" @click="submitTransfer">提交</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="detailVisible" title="调拨单详情" width="900px" top="5vh">
      <div v-if="currentDetail" class="transfer-detail">
        <el-descriptions :column="2" border>
          <el-descriptions-item label="调拨单号">{{ currentDetail.transfer_no }}</el-descriptions-item>
          <el-descriptions-item label="状态">
            <el-tag :type="getStatusType(currentDetail.status)">{{ getStatusText(currentDetail.status) }}</el-tag>
            <el-tag v-if="currentDetail.has_diff" type="warning" size="small" style="margin-left: 8px">有差异</el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="调出门店">{{ currentDetail.from_store?.name || '-' }}</el-descriptions-item>
          <el-descriptions-item label="调入门店">{{ currentDetail.to_store?.name || '-' }}</el-descriptions-item>
          <el-descriptions-item label="总数量">{{ currentDetail.total_qty }}</el-descriptions-item>
          <el-descriptions-item label="总金额">¥{{ currentDetail.total_amount }}</el-descriptions-item>
          <el-descriptions-item label="创建时间">{{ currentDetail.created_at }}</el-descriptions-item>
          <el-descriptions-item label="接单时间">{{ currentDetail.accepted_at || '-' }}</el-descriptions-item>
          <el-descriptions-item label="出库时间">{{ currentDetail.out_confirmed_at || '-' }}</el-descriptions-item>
          <el-descriptions-item label="收货时间">{{ currentDetail.received_at || '-' }}</el-descriptions-item>
          <el-descriptions-item label="完成时间">{{ currentDetail.completed_at || '-' }}</el-descriptions-item>
          <el-descriptions-item label="接单人" v-if="currentDetail.accept_operator_name">
            {{ currentDetail.accept_operator_name }}
          </el-descriptions-item>
          <el-descriptions-item label="出库人" v-if="currentDetail.out_operator_name">
            {{ currentDetail.out_operator_name }}
          </el-descriptions-item>
          <el-descriptions-item label="收货人" v-if="currentDetail.in_operator_name">
            {{ currentDetail.in_operator_name }}
          </el-descriptions-item>
          <el-descriptions-item label="物流公司" v-if="currentDetail.logistics_company">
            {{ currentDetail.logistics_company }}
          </el-descriptions-item>
          <el-descriptions-item label="运单号" v-if="currentDetail.tracking_no">
            {{ currentDetail.tracking_no }}
          </el-descriptions-item>
          <el-descriptions-item label="备注" :span="2">
            {{ currentDetail.remark || '-' }}
          </el-descriptions-item>
          <el-descriptions-item label="差异说明" v-if="currentDetail.has_diff" :span="2">
            {{ currentDetail.diff_remark || '-' }}
          </el-descriptions-item>
        </el-descriptions>

        <h4 style="margin: 20px 0 10px">调拨明细</h4>
        <el-table :data="currentDetail.items" border size="small">
          <el-table-column prop="ingredient_name" label="食材名称" min-width="120" />
          <el-table-column prop="unit" label="单位" width="70" />
          <el-table-column prop="out_qty" label="调出数量" width="100" />
          <el-table-column prop="in_qty" label="实收数量" width="100" />
          <el-table-column prop="diff_qty" label="差异数量" width="100">
            <template #default="{ row }">
              <span :style="{ color: row.diff_qty > 0 ? '#67c23a' : row.diff_qty < 0 ? '#f56c6c' : '' }">
                {{ row.diff_qty > 0 ? '+' : '' }}{{ row.diff_qty }}
              </span>
            </template>
          </el-table-column>
          <el-table-column prop="unit_price" label="单价" width="100">
            <template #default="{ row }">¥{{ row.unit_price }}</template>
          </el-table-column>
          <el-table-column prop="amount" label="金额" width="100">
            <template #default="{ row }">¥{{ row.amount }}</template>
          </el-table-column>
          <el-table-column prop="remark" label="备注" min-width="100" />
        </el-table>
      </div>
      <template #footer>
        <el-button @click="detailVisible = false">关闭</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="logisticsVisible" title="物流跟踪" width="600px">
      <div v-if="logisticsData" class="logistics-tracking">
        <div class="logistics-header">
          <span>运单号：{{ logisticsData.tracking_no }}</span>
          <el-button size="small" type="primary" link @click="handleRefreshLogistics">刷新</el-button>
        </div>
        <el-timeline v-if="logisticsData.tracks && logisticsData.tracks.length > 0">
          <el-timeline-item
            v-for="(track, index) in logisticsData.tracks"
            :key="index"
            :timestamp="formatTrackTime(track)"
            :type="index === 0 ? 'primary' : ''">
            {{ track.description || track.accept_station }}
            <div v-if="track.location" style="font-size: 12px; color: #909399; margin-top: 4px">
              {{ track.location }}
            </div>
          </el-timeline-item>
        </el-timeline>
        <el-empty v-else description="暂无物流信息" />
      </div>
    </el-dialog>

    <el-dialog v-model="receiveVisible" title="确认收货" width="800px" top="5vh">
      <div v-if="receiveItems.length > 0">
        <el-alert
          title="请核对实收数量，如有差异请填写实际数量和备注"
          type="warning"
          :closable="false"
          style="margin-bottom: 20px"
        />
        <el-table :data="receiveItems" border>
          <el-table-column prop="ingredient_name" label="食材名称" min-width="120" />
          <el-table-column prop="unit" label="单位" width="70" />
          <el-table-column prop="out_qty" label="调出数量" width="100" />
          <el-table-column label="实收数量" width="150">
            <template #default="{ row }">
              <el-input-number v-model="row.in_qty" :precision="2" :min="0" size="small" style="width: 100%" />
            </template>
          </el-table-column>
          <el-table-column label="差异" width="100">
            <template #default="{ row }">
              <span :style="{ color: (row.in_qty - row.out_qty) > 0 ? '#67c23a' : (row.in_qty - row.out_qty) < 0 ? '#f56c6c' : '' }">
                {{ (row.in_qty - row.out_qty) > 0 ? '+' : '' }}{{ (row.in_qty - row.out_qty).toFixed(2) }}
              </span>
            </template>
          </el-table-column>
          <el-table-column label="备注" min-width="150">
            <template #default="{ row }">
              <el-input v-model="row.remark" size="small" placeholder="差异说明" />
            </template>
          </el-table-column>
        </el-table>
        <el-form style="margin-top: 20px">
          <el-form-item label="收货备注">
            <el-input v-model="receiveRemark" type="textarea" :rows="2" placeholder="请输入收货备注" />
          </el-form-item>
        </el-form>
      </div>
      <template #footer>
        <el-button @click="receiveVisible = false">取消</el-button>
        <el-button type="primary" @click="submitReceive">确认收货</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="shipVisible" title="发货" width="500px">
      <el-form :model="shipForm" label-width="100px">
        <el-form-item label="物流公司">
          <el-select v-model="shipForm.logistics_company" placeholder="请选择物流公司" style="width: 100%" filterable>
            <el-option label="顺丰速运" value="顺丰速运" />
            <el-option label="圆通速递" value="圆通速递" />
            <el-option label="中通快递" value="中通快递" />
            <el-option label="申通快递" value="申通快递" />
            <el-option label="韵达快递" value="韵达快递" />
            <el-option label="百世快递" value="百世快递" />
            <el-option label="EMS" value="EMS" />
            <el-option label="京东物流" value="京东物流" />
            <el-option label="德邦物流" value="德邦物流" />
          </el-select>
        </el-form-item>
        <el-form-item label="运单号">
          <el-input v-model="shipForm.tracking_no" placeholder="请输入运单号" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="shipVisible = false">取消</el-button>
        <el-button type="primary" @click="submitShip">确认发货</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import * as transferApi from '@/api/transfer'
import { storeApi } from '@/api/stores'
import * as ingredientApi from '@/api/ingredient'

const loading = ref(false)
const createVisible = ref(false)
const detailVisible = ref(false)
const logisticsVisible = ref(false)
const receiveVisible = ref(false)
const shipVisible = ref(false)

const currentDetail = ref(null)
const logisticsData = ref(null)
const receiveItems = ref([])
const receiveRemark = ref('')

const storeList = ref([])
const availableIngredients = ref([])
const selectedIngredient = ref(null)

const searchForm = reactive({
  from_store_id: '',
  to_store_id: '',
  status: -1,
  keyword: ''
})

const pagination = reactive({
  page: 1,
  pageSize: 20,
  total: 0
})

const tableData = ref([])

const transferForm = reactive({
  from_store_id: 0,
  to_store_id: 0,
  transfer_type: 'normal',
  priority: 'normal',
  sender_name: '',
  sender_phone: '',
  receiver_name: '',
  receiver_phone: '',
  receiver_address: '',
  remark: '',
  items: []
})

const shipForm = reactive({
  logistics_company: '',
  tracking_no: '',
  logistics_code: ''
})

onMounted(() => {
  loadStores()
  fetchList()
})

function loadStores() {
  storeApi.list({ page: 1, page_size: 100 }).then(res => {
    storeList.value = res.data?.list || res.data || []
  }).catch(() => {
    storeList.value = [{ id: 1, name: '默认门店' }]
  })
}

function fetchList() {
  loading.value = true
  const params = {
    from_store_id: searchForm.from_store_id || undefined,
    to_store_id: searchForm.to_store_id || undefined,
    status: searchForm.status >= 0 ? searchForm.status : undefined,
    keyword: searchForm.keyword || undefined,
    page: pagination.page,
    page_size: pagination.pageSize
  }
  transferApi.getTransfers(params).then(res => {
    const data = res.data
    tableData.value = data?.list || data || []
    pagination.total = data?.total || 0
    loading.value = false
  }).catch(err => {
    ElMessage.error('获取调拨单列表失败')
    loading.value = false
  })
}

function handleSearch() {
  pagination.page = 1
  fetchList()
}

function handleReset() {
  searchForm.from_store_id = ''
  searchForm.to_store_id = ''
  searchForm.status = -1
  searchForm.keyword = ''
  pagination.page = 1
  fetchList()
}

function getStatusText(status) {
  const map = {
    0: '待接单',
    1: '待出库',
    2: '已出库',
    3: '运输中',
    4: '已收货',
    5: '已完成',
    6: '已取消'
  }
  return map[status] || '未知'
}

function getStatusType(status) {
  const map = {
    0: 'warning',
    1: 'info',
    2: 'primary',
    3: 'info',
    4: 'success',
    5: 'success',
    6: 'info'
  }
  return map[status] || 'info'
}

function formatTrackTime(track) {
  if (track.track_time) return track.track_time
  if (track.accept_time) return track.accept_time
  return ''
}

function handleCreate() {
  Object.assign(transferForm, {
    from_store_id: storeList.value.length > 0 ? storeList.value[0].id : 0,
    to_store_id: 0,
    transfer_type: 'normal',
    priority: 'normal',
    sender_name: '',
    sender_phone: '',
    receiver_name: '',
    receiver_phone: '',
    receiver_address: '',
    remark: '',
    items: []
  })
  selectedIngredient.value = null
  if (transferForm.from_store_id) {
    loadIngredients()
  }
  createVisible.value = true
}

function loadIngredients() {
  if (!transferForm.from_store_id) return
  ingredientApi.getIngredients({ store_id: transferForm.from_store_id, page_size: 100, status: 1 }).then(res => {
    const data = res.data
    availableIngredients.value = data?.list || data || []
  })
}

function addIngredient(ingredientId) {
  const ing = availableIngredients.value.find(i => i.id === ingredientId)
  if (!ing) return

  const exist = transferForm.items.find(item => item.ingredient_id === ingredientId)
  if (exist) {
    ElMessage.warning('该食材已添加')
    return
  }

  transferForm.items.push({
    ingredient_id: ing.id,
    ingredient_no: ing.ingredient_no,
    ingredient_name: ing.name,
    unit: ing.unit,
    out_qty: 1,
    unit_price: ing.current_price,
    remark: ''
  })
  selectedIngredient.value = null
}

function removeItem(index) {
  transferForm.items.splice(index, 1)
}

function submitTransfer() {
  if (!transferForm.from_store_id) {
    ElMessage.warning('请选择调出门店')
    return
  }
  if (!transferForm.to_store_id) {
    ElMessage.warning('请选择调入门店')
    return
  }
  if (transferForm.from_store_id === transferForm.to_store_id) {
    ElMessage.warning('调出门店和调入门店不能相同')
    return
  }
  if (transferForm.items.length === 0) {
    ElMessage.warning('请添加调拨食材')
    return
  }

  transferApi.createTransfer(transferForm).then(() => {
    ElMessage.success('调拨单创建成功')
    createVisible.value = false
    fetchList()
  }).catch(err => {
    ElMessage.error(err.message || '创建失败')
  })
}

function viewDetail(row) {
  transferApi.getTransfer(row.id).then(res => {
    currentDetail.value = res.data
    detailVisible.value = true
  })
}

function viewLogistics(row) {
  transferApi.getLogisticsTrack(row.id).then(res => {
    logisticsData.value = res.data
    logisticsVisible.value = true
  }).catch(err => {
    ElMessage.error('获取物流信息失败')
  })
}

function handleRefreshLogistics() {
  if (!logisticsData.value?.tracking_no) return
  transferApi.refreshLogistics(logisticsData.value.transfer_id || currentDetail.value?.id).then(res => {
    logisticsData.value = {
      ...logisticsData.value,
      tracks: res.data
    }
    ElMessage.success('刷新成功')
  }).catch(err => {
    ElMessage.error(err.message || '刷新失败')
  })
}

function handleAccept(row) {
  ElMessageBox.confirm('确认接单吗？接单后请及时安排出库。', '确认接单', {
    type: 'info'
  }).then(() => {
    transferApi.acceptTransfer(row.id, {}).then(() => {
      ElMessage.success('接单成功')
      fetchList()
    }).catch(err => {
      ElMessage.error(err.message || '接单失败')
    })
  }).catch(() => {})
}

function handleReject(row) {
  ElMessageBox.prompt('请输入拒单原因', '拒单', {
    confirmButtonText: '确认拒单',
    cancelButtonText: '取消',
    inputPlaceholder: '请输入拒单原因',
    type: 'warning'
  }).then(({ value }) => {
    transferApi.rejectTransfer(row.id, { reason: value || '' }).then(() => {
      ElMessage.success('已拒单')
      fetchList()
    }).catch(err => {
      ElMessage.error(err.message || '拒单失败')
    })
  }).catch(() => {})
}

function handleOutbound(row) {
  ElMessageBox.confirm('确认出库吗？出库后库存将从调出门店扣除。', '确认出库', {
    type: 'warning'
  }).then(() => {
    transferApi.confirmOutbound(row.id, {}).then(() => {
      ElMessage.success('出库成功')
      fetchList()
    }).catch(err => {
      ElMessage.error(err.message || '出库失败')
    })
  }).catch(() => {})
}

function handleShip(row) {
  shipForm.logistics_company = ''
  shipForm.tracking_no = ''
  shipVisible.value = true
  shipForm.currentId = row.id
}

function submitShip() {
  if (!shipForm.logistics_company) {
    ElMessage.warning('请选择物流公司')
    return
  }
  if (!shipForm.tracking_no) {
    ElMessage.warning('请输入运单号')
    return
  }
  transferApi.startShipping(shipForm.currentId, shipForm).then(() => {
    ElMessage.success('发货成功')
    shipVisible.value = false
    fetchList()
  }).catch(err => {
    ElMessage.error(err.message || '发货失败')
  })
}

function handleReceive(row) {
  transferApi.getTransfer(row.id).then(res => {
    const transfer = res.data
    receiveItems.value = transfer.items.map(item => ({
      id: item.id,
      item_id: item.id,
      ingredient_id: item.ingredient_id,
      ingredient_name: item.ingredient_name,
      unit: item.unit,
      out_qty: item.out_qty,
      in_qty: item.out_qty,
      remark: ''
    }))
    receiveRemark.value = ''
    receiveVisible.value = true
    receiveVisible.currentId = row.id
  })
}

function submitReceive() {
  const items = receiveItems.value.map(item => ({
    item_id: item.item_id,
    in_qty: item.in_qty,
    remark: item.remark
  }))
  transferApi.receiveTransfer(receiveVisible.currentId, {
    items,
    remark: receiveRemark.value
  }).then(() => {
    ElMessage.success('收货成功')
    receiveVisible.value = false
    fetchList()
  }).catch(err => {
    ElMessage.error(err.message || '收货失败')
    })
}

function handleComplete(row) {
  ElMessageBox.prompt('请输入完成备注（可选）', '完成调拨', {
    confirmButtonText: '确认',
    cancelButtonText: '取消',
    inputPlaceholder: '请输入备注',
    type: 'info'
  }).then(({ value }) => {
    transferApi.completeTransfer(row.id, { diff_remark: value || '' }).then(() => {
      ElMessage.success('调拨完成')
      fetchList()
    }).catch(err => {
      ElMessage.error(err.message || '操作失败')
    })
  }).catch(() => {})
}

function handleCancel(row) {
  ElMessageBox.prompt('请输入取消原因', '取消调拨', {
    confirmButtonText: '确认取消',
    cancelButtonText: '取消',
    inputPlaceholder: '请输入取消原因',
    type: 'warning'
  }).then(({ value }) => {
    transferApi.cancelTransfer(row.id, { remark: value || '' }).then(() => {
      ElMessage.success('已取消')
      fetchList()
    }).catch(err => {
      ElMessage.error(err.message || '取消失败')
    })
  }).catch(() => {})
}
</script>

<style scoped>
.transfers-page {
  padding: 20px;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.page-header h2 {
  margin: 0;
}

.search-card {
  margin-bottom: 20px;
}

.table-card {
  margin-bottom: 20px;
}

.pagination {
  margin-top: 20px;
  justify-content: flex-end;
  display: flex;
}

.item-list-wrapper {
  width: 100%;
}

.add-item-btn {
  margin-top: 10px;
  text-align: left;
}

.transfer-detail {
  padding: 10px 0;
}

.transfer-detail h4 {
  margin: 0 0 10px 0;
  padding: 0;
}

.logistics-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
  padding-bottom: 10px;
  border-bottom: 1px solid #ebeef5;
}

.logistics-tracking {
  max-height: 400px;
  overflow-y: auto;
}
</style>
