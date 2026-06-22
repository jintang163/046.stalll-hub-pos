<template>
  <div class="purchase-orders-page">
    <div class="page-header">
      <h2>采购订单管理</h2>
      <div class="header-actions">
        <el-button @click="activeTab = 'orders'">
          <el-icon><Document /></el-icon>
          采购订单
        </el-button>
        <el-button @click="activeTab = 'receives'">
          <el-icon><Box /></el-icon>
          收货入库
        </el-button>
        <el-button type="primary" @click="handleAddOrder">
          <el-icon><Plus /></el-icon>
          新建采购订单
        </el-button>
      </div>
    </div>

    <el-tabs v-model="activeTab">
      <el-tab-pane label="采购订单" name="orders">
        <el-card class="search-card">
          <el-form :inline="true" :model="orderSearch">
            <el-form-item label="门店">
              <el-select v-model="orderSearch.store_id" placeholder="全部门店" style="width: 150px">
                <el-option v-for="store in storeList" :key="store.id" :label="store.name" :value="store.id" />
              </el-select>
            </el-form-item>
            <el-form-item label="状态">
              <el-select v-model="orderSearch.status" placeholder="全部" style="width: 140px">
                <el-option label="全部" :value="0" />
                <el-option v-for="(label, status) in purchaseStatusMap" :key="status" :label="label" :value="Number(status)" />
              </el-select>
            </el-form-item>
            <el-form-item label="供应商">
              <el-select v-model="orderSearch.supplier_id" placeholder="全部供应商" clearable filterable style="width: 200px">
                <el-option v-for="sup in supplierList" :key="sup.id" :label="sup.name" :value="sup.id" />
              </el-select>
            </el-form-item>
            <el-form-item label="关键词">
              <el-input v-model="orderSearch.keyword" placeholder="订单号/供应商" clearable style="width: 200px" />
            </el-form-item>
            <el-form-item>
              <el-button type="primary" @click="fetchOrders">搜索</el-button>
              <el-button @click="resetOrderSearch">重置</el-button>
            </el-form-item>
          </el-form>
        </el-card>

        <el-card class="table-card">
          <el-table :data="orderList" v-loading="orderLoading" border stripe>
            <el-table-column prop="purchase_no" label="采购单号" width="160" />
            <el-table-column prop="supplier_name" label="供应商" min-width="160" show-overflow-tooltip />
            <el-table-column label="商品数/总件数" width="140">
              <template #default="{ row }">
                <div>{{ row.item_count }} 种</div>
                <div style="color: #909399; font-size: 12px">{{ row.total_quantity }} 件</div>
              </template>
            </el-table-column>
            <el-table-column label="金额" width="140">
              <template #default="{ row }">
                <div style="color: #409eff; font-weight: bold">¥{{ formatAmount(row.total_amount) }}</div>
                <div v-if="row.received_amount > 0" style="color: #67c23a; font-size: 12px">
                  已收 ¥{{ formatAmount(row.received_amount) }}
                </div>
              </template>
            </el-table-column>
            <el-table-column prop="payment_term_text" label="账期" width="100" />
            <el-table-column prop="expected_date" label="预计到货" width="120" />
            <el-table-column prop="status_text" label="状态" width="100">
              <template #default="{ row }">
                <el-tag :type="getStatusTagType(row.status)">
                  {{ row.status_text }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="created_at" label="创建时间" width="160">
              <template #default="{ row }">
                {{ formatDateTime(row.created_at) }}
              </template>
            </el-table-column>
            <el-table-column label="操作" width="280" fixed="right">
              <template #default="{ row }">
                <el-button link type="primary" @click="viewOrderDetail(row)">详情</el-button>
                <el-button link type="primary" v-if="row.status === 0 || row.status === 5" @click="sendOrder(row)">发送</el-button>
                <el-button link type="primary" v-if="row.status >= 1 && row.status < 4" @click="handleReceive(row)">收货</el-button>
                <el-button link type="success" v-if="row.status >= 1 && row.status < 4" @click="completeOrder(row)">完成</el-button>
                <el-button link type="danger" v-if="row.status < 3" @click="cancelOrder(row)">取消</el-button>
              </template>
            </el-table-column>
          </el-table>

          <el-pagination
            v-model:current-page="orderPagination.page"
            v-model:page-size="orderPagination.pageSize"
            :total="orderPagination.total"
            :page-sizes="[10, 20, 50, 100]"
            layout="total, sizes, prev, pager, next, jumper"
            class="pagination"
            @size-change="fetchOrders"
            @current-change="fetchOrders"
          />
        </el-card>
      </el-tab-pane>

      <el-tab-pane label="收货入库记录" name="receives">
        <el-card class="search-card">
          <el-form :inline="true" :model="receiveSearch">
            <el-form-item label="门店">
              <el-select v-model="receiveSearch.store_id" placeholder="全部门店" style="width: 150px">
                <el-option v-for="store in storeList" :key="store.id" :label="store.name" :value="store.id" />
              </el-select>
            </el-form-item>
            <el-form-item label="供应商">
              <el-select v-model="receiveSearch.supplier_id" placeholder="全部供应商" clearable filterable style="width: 200px">
                <el-option v-for="sup in supplierList" :key="sup.id" :label="sup.name" :value="sup.id" />
              </el-select>
            </el-form-item>
            <el-form-item label="日期">
              <el-date-picker
                v-model="dateRange"
                type="daterange"
                range-separator="至"
                start-placeholder="开始日期"
                end-placeholder="结束日期"
                value-format="YYYY-MM-DD"
                style="width: 260px"
              />
            </el-form-item>
            <el-form-item label="关键词">
              <el-input v-model="receiveSearch.keyword" placeholder="收货单号/采购单号" clearable style="width: 200px" />
            </el-form-item>
            <el-form-item>
              <el-button type="primary" @click="fetchReceives">搜索</el-button>
              <el-button @click="resetReceiveSearch">重置</el-button>
            </el-form-item>
          </el-form>
        </el-card>

        <el-card class="table-card">
          <el-table :data="receiveList" v-loading="receiveLoading" border stripe>
            <el-table-column prop="receive_no" label="收货单号" width="180" />
            <el-table-column prop="purchase_no" label="关联采购单" width="160" />
            <el-table-column prop="supplier_name" label="供应商" min-width="160" show-overflow-tooltip />
            <el-table-column prop="receive_type" label="收货类型" width="100">
              <template #default="{ row }">
                <el-tag :type="row.receive_type === 'full' ? 'success' : 'warning'">
                  {{ row.receive_type === 'full' ? '全部' : '部分' }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="total_qty" label="收货数量" width="120">
              <template #default="{ row }">
                <span style="color: #67c23a; font-weight: bold">{{ row.total_qty }}</span>
              </template>
            </el-table-column>
            <el-table-column prop="total_amount" label="收货金额" width="130">
              <template #default="{ row }">
                <span style="color: #409eff; font-weight: bold">¥{{ formatAmount(row.total_amount) }}</span>
              </template>
            </el-table-column>
            <el-table-column prop="operator_name" label="操作人" width="100" />
            <el-table-column prop="received_at" label="收货时间" width="160">
              <template #default="{ row }">
                {{ formatDateTime(row.received_at || row.created_at) }}
              </template>
            </el-table-column>
            <el-table-column label="操作" width="100" fixed="right">
              <template #default="{ row }">
                <el-button link type="primary" @click="viewReceiveDetail(row)">详情</el-button>
              </template>
            </el-table-column>
          </el-table>

          <el-pagination
            v-model:current-page="receivePagination.page"
            v-model:page-size="receivePagination.pageSize"
            :total="receivePagination.total"
            :page-sizes="[10, 20, 50, 100]"
            layout="total, sizes, prev, pager, next, jumper"
            class="pagination"
            @size-change="fetchReceives"
            @current-change="fetchReceives"
          />
        </el-card>
      </el-tab-pane>
    </el-tabs>

    <el-dialog v-model="orderDialogVisible" :title="orderDialogTitle" width="800px" top="3vh">
      <el-form :model="orderForm" label-width="100px">
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="门店" required>
              <el-select v-model="orderForm.store_id" placeholder="请选择门店" style="width: 100%">
                <el-option v-for="store in storeList" :key="store.id" :label="store.name" :value="store.id" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="供应商" required>
              <el-select
                v-model="orderForm.supplier_id"
                placeholder="请选择供应商"
                filterable
                style="width: 100%"
                @change="onSupplierChange"
              >
                <el-option v-for="sup in supplierList" :key="sup.id" :label="sup.name" :value="sup.id" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item label="账期">
              <el-select v-model="orderForm.payment_term" placeholder="请选择账期" style="width: 100%">
                <el-option v-for="(label, days) in paymentTermMap" :key="days" :label="label" :value="Number(days)" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item label="预计到货">
              <el-date-picker v-model="orderForm.expected_date" type="date" value-format="YYYY-MM-DD" style="width: 100%" />
            </el-form-item>
          </el-col>
          <el-col :span="24">
            <el-form-item label="备注">
              <el-input v-model="orderForm.remark" type="textarea" :rows="1" maxlength="255" />
            </el-form-item>
          </el-col>
        </el-row>

        <el-divider content-position="left">采购明细</el-divider>

        <div class="items-toolbar">
          <el-button size="small" type="primary" @click="addOrderItem">
            <el-icon><Plus /></el-icon>
            添加食材
          </el-button>
          <el-button size="small" @click="openIngredientSelect">
            <el-icon><Search /></el-icon>
            从食材库选择
          </el-button>
        </div>

        <el-table :data="orderForm.items" border size="small" class="items-table">
          <el-table-column label="食材名称" min-width="160">
            <template #default="{ row, $index }">
              <el-input v-model="row.ingredient_name" size="small" placeholder="食材名称" />
            </template>
          </el-table-column>
          <el-table-column label="分类" width="110">
            <template #default="{ row }">
              <el-input v-model="row.category" size="small" placeholder="分类" />
            </template>
          </el-table-column>
          <el-table-column label="单位" width="80">
            <template #default="{ row }">
              <el-input v-model="row.unit" size="small" placeholder="单位" />
            </template>
          </el-table-column>
          <el-table-column label="采购数量" width="120">
            <template #default="{ row }">
              <el-input-number v-model="row.purchase_qty" :precision="2" :min="0" size="small" controls-position="right" style="width: 100%" />
            </template>
          </el-table-column>
          <el-table-column label="单价(元)" width="120">
            <template #default="{ row }">
              <el-input-number v-model="row.unit_price" :precision="2" :min="0" size="small" controls-position="right" style="width: 100%" />
            </template>
          </el-table-column>
          <el-table-column label="小计" width="120">
            <template #default="{ row }">
              <span style="color: #409eff; font-weight: bold">¥{{ formatAmount(row.purchase_qty * row.unit_price) }}</span>
            </template>
          </el-table-column>
          <el-table-column label="操作" width="60">
            <template #default="{ $index }">
              <el-button link type="danger" size="small" @click="removeOrderItem($index)">删除</el-button>
            </template>
          </el-table-column>
        </el-table>

        <div class="order-total">
          <span>共 <b>{{ orderForm.items.length }}</b> 种食材，</span>
          <span>合计 <b style="color: #f56c6c; font-size: 18px">¥{{ orderTotalAmount }}</b></span>
        </div>
      </el-form>
      <template #footer>
        <el-button @click="orderDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="submitOrder">保存</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="receiveDialogVisible" title="采购收货入库" width="860px" top="3vh">
      <div v-if="currentOrder" class="receive-info">
        <el-descriptions :column="3" border size="small">
          <el-descriptions-item label="采购单号">{{ currentOrder.purchase_no }}</el-descriptions-item>
          <el-descriptions-item label="供应商">{{ currentOrder.supplier_name }}</el-descriptions-item>
          <el-descriptions-item label="账期">{{ currentOrder.payment_term_text || '-' }}</el-descriptions-item>
        </el-descriptions>
      </div>

      <el-divider content-position="left">收货明细（请填写实际收货数量）</el-divider>

      <div class="receive-toolbar">
        <el-alert type="info" :closable="false" size="small" show-icon>
          填写收货数量，系统将自动增加对应食材库存。支持分批收货。
        </el-alert>
      </div>

      <el-table :data="receiveItems" border size="small" class="items-table">
        <el-table-column label="食材名称" min-width="140">
          <template #default="{ row }">{{ row.ingredient_name }}</template>
        </el-table-column>
        <el-table-column label="分类" width="80">
          <template #default="{ row }">{{ row.category }}</template>
        </el-table-column>
        <el-table-column label="单位" width="60">
          <template #default="{ row }">{{ row.unit }}</template>
        </el-table-column>
        <el-table-column label="采购数量" width="100" align="right">
          <template #default="{ row }">
            <span style="color: #909399">{{ row.purchase_qty }}</span>
          </template>
        </el-table-column>
        <el-table-column label="已收数量" width="100" align="right">
          <template #default="{ row }">
            <span style="color: #67c23a">{{ row.received_qty || 0 }}</span>
          </template>
        </el-table-column>
        <el-table-column label="待收数量" width="100" align="right">
          <template #default="{ row }">
            <span style="color: #e6a23c">{{ Math.max(0, row.purchase_qty - (row.received_qty || 0)) }}</span>
          </template>
        </el-table-column>
        <el-table-column label="本次收货" width="130">
          <template #default="{ row }">
            <el-input-number
              v-model="row.actual_received_qty"
              :precision="2"
              :min="0"
              :max="row.purchase_qty - (row.received_qty || 0) + 100"
              size="small"
              controls-position="right"
              style="width: 100%"
            />
          </template>
        </el-table-column>
        <el-table-column label="合格数量" width="120">
          <template #default="{ row }">
            <el-input-number v-model="row.qualified_qty" :precision="2" :min="0" size="small" controls-position="right" style="width: 100%" />
          </template>
        </el-table-column>
        <el-table-column label="单价" width="90" align="right">
          <template #default="{ row }">¥{{ formatAmount(row.unit_price) }}</template>
        </el-table-column>
        <el-table-column label="批次号" width="120">
          <template #default="{ row }">
            <el-input v-model="row.batch_no" size="small" placeholder="批次" />
          </template>
        </el-table-column>
        <el-table-column label="保质期" width="110">
          <template #default="{ row }">
            <el-date-picker v-model="row.expiry_date" type="date" value-format="YYYY-MM-DD" size="small" style="width: 100%" />
          </template>
        </el-table-column>
      </el-table>

      <el-form label-width="80px" class="receive-extra">
        <el-form-item label="备注">
          <el-input v-model="receiveForm.remark" type="textarea" :rows="2" maxlength="255" />
        </el-form-item>
        <el-form-item label="操作人">
          <el-input v-model="receiveForm.operator_name" maxlength="50" placeholder="请输入操作人姓名" style="width: 200px" />
        </el-form-item>
      </el-form>

      <template #footer>
        <el-button @click="receiveDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="submitReceive">确认入库</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="orderDetailVisible" title="采购订单详情" width="760px">
      <div v-if="currentOrderDetail">
        <el-descriptions :column="3" border>
          <el-descriptions-item label="采购单号">{{ currentOrderDetail.purchase_no }}</el-descriptions-item>
          <el-descriptions-item label="供应商">{{ currentOrderDetail.supplier_name }}</el-descriptions-item>
          <el-descriptions-item label="状态">
            <el-tag :type="getStatusTagType(currentOrderDetail.status)">{{ currentOrderDetail.status_text }}</el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="门店">{{ currentOrderDetail.store_name }}</el-descriptions-item>
          <el-descriptions-item label="账期">{{ currentOrderDetail.payment_term_text || '-' }}</el-descriptions-item>
          <el-descriptions-item label="预计到货">{{ currentOrderDetail.expected_date || '-' }}</el-descriptions-item>
          <el-descriptions-item label="联系电话">{{ currentOrderDetail.supplier_phone || '-' }}</el-descriptions-item>
          <el-descriptions-item label="邮箱">{{ currentOrderDetail.supplier_email || '-' }}</el-descriptions-item>
          <el-descriptions-item label="创建时间">{{ formatDateTime(currentOrderDetail.created_at) }}</el-descriptions-item>
        </el-descriptions>

        <el-divider content-position="left">采购明细</el-divider>

        <el-table :data="currentOrderDetail.items" border size="small">
          <el-table-column prop="ingredient_name" label="食材名称" min-width="140" />
          <el-table-column prop="category" label="分类" width="80" />
          <el-table-column prop="unit" label="单位" width="60" />
          <el-table-column prop="purchase_qty" label="采购数量" width="100" align="right" />
          <el-table-column label="已收数量" width="100" align="right">
            <template #default="{ row }">
              <span style="color: #67c23a">{{ row.received_qty || 0 }}</span>
            </template>
          </el-table-column>
          <el-table-column prop="unit_price" label="单价" width="100" align="right">
            <template #default="{ row }">¥{{ formatAmount(row.unit_price) }}</template>
          </el-table-column>
          <el-table-column prop="subtotal" label="小计" width="120" align="right">
            <template #default="{ row }">
              <span style="color: #409eff; font-weight: bold">¥{{ formatAmount(row.subtotal) }}</span>
            </template>
          </el-table-column>
        </el-table>

        <div class="detail-summary">
          <div>合计：<b style="color: #f56c6c; font-size: 18px">¥{{ formatAmount(currentOrderDetail.total_amount) }}</b></div>
          <div v-if="currentOrderDetail.remark" style="color: #909399; margin-top: 8px">备注：{{ currentOrderDetail.remark }}</div>
        </div>
      </div>
    </el-dialog>

    <el-dialog v-model="ingredientSelectVisible" title="选择食材" width="720px" top="8vh">
      <div class="ingredient-search">
        <el-input
          v-model="ingredientKeyword"
          placeholder="搜索食材名称/编号"
          clearable
          style="width: 280px"
          @keyup.enter="searchIngredients"
        >
          <template #prefix>
            <el-icon><Search /></el-icon>
          </template>
        </el-input>
        <el-button type="primary" @click="searchIngredients">搜索</el-button>
      </div>
      <el-table
        ref="ingredientTableRef"
        :data="ingredientList"
        border
        height="380"
        @selection-change="onIngredientSelectionChange"
        v-loading="ingredientLoading"
      >
        <el-table-column type="selection" width="50" />
        <el-table-column prop="ingredient_no" label="编号" width="100" />
        <el-table-column prop="name" label="食材名称" min-width="160" />
        <el-table-column prop="category" label="分类" width="90" />
        <el-table-column prop="unit" label="单位" width="60" />
        <el-table-column prop="current_price" label="当前单价" width="110" align="right">
          <template #default="{ row }">¥{{ formatAmount(row.current_price) }}</template>
        </el-table-column>
        <el-table-column prop="current_stock" label="当前库存" width="100" align="right" />
      </el-table>
      <template #footer>
        <el-button @click="ingredientSelectVisible = false">取消</el-button>
        <el-button type="primary" @click="confirmIngredientSelect">确定添加</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Document, Box, Search } from '@element-plus/icons-vue'
import * as supplierApi from '@/api/supplier'
import * as ingredientApi from '@/api/ingredient'
import { storeApi } from '@/api/stores'

const activeTab = ref('orders')

const orderLoading = ref(false)
const receiveLoading = ref(false)
const ingredientLoading = ref(false)

const orderDialogVisible = ref(false)
const receiveDialogVisible = ref(false)
const orderDetailVisible = ref(false)
const ingredientSelectVisible = ref(false)

const orderDialogTitle = ref('新建采购订单')

const currentOrder = ref(null)
const currentOrderDetail = ref(null)

const ingredientTableRef = ref(null)
const selectedIngredients = ref([])
const ingredientKeyword = ref('')
const ingredientList = ref([])

const purchaseStatusMap = {
  0: '待发送',
  1: '已发送',
  2: '已确认',
  3: '部分入库',
  4: '已完成',
  5: '已取消'
}

const enums = ref({
  payment_terms: {}
})
const paymentTermMap = computed(() => enums.value.payment_terms || {})

const dateRange = ref([])

const storeList = ref([])
const supplierList = ref([])

const orderSearch = reactive({
  store_id: 0,
  status: 0,
  supplier_id: 0,
  keyword: ''
})

const receiveSearch = reactive({
  store_id: 0,
  supplier_id: 0,
  keyword: ''
})

const orderPagination = reactive({ page: 1, pageSize: 20, total: 0 })
const receivePagination = reactive({ page: 1, pageSize: 20, total: 0 })

const orderList = ref([])
const receiveList = ref([])

const orderForm = reactive({
  id: 0,
  store_id: 0,
  supplier_id: 0,
  payment_term: 0,
  expected_date: '',
  remark: '',
  items: []
})

const receiveForm = reactive({
  remark: '',
  operator_name: ''
})

const receiveItems = ref([])

const orderTotalAmount = computed(() => {
  return orderForm.items.reduce((sum, item) => sum + (item.purchase_qty * item.unit_price || 0), 0)
})

onMounted(() => {
  loadEnums()
  loadStores()
})

function loadEnums() {
  supplierApi.getSupplierEnums().then(res => {
    enums.value = res.data || {}
  })
}

function loadStores() {
  storeApi.list({ page: 1, page_size: 100 }).then(res => {
    storeList.value = res.data?.list || res.data || []
    if (storeList.value.length > 0) {
      orderSearch.store_id = storeList.value[0].id
      receiveSearch.store_id = storeList.value[0].id
      orderForm.store_id = storeList.value[0].id
    }
    loadSuppliers()
    fetchOrders()
    fetchReceives()
  }).catch(() => {
    storeList.value = [{ id: 1, name: '默认门店' }]
    orderSearch.store_id = 1
    receiveSearch.store_id = 1
    orderForm.store_id = 1
    loadSuppliers()
    fetchOrders()
    fetchReceives()
  })
}

function loadSuppliers() {
  supplierApi.getSuppliers({ page: 1, page_size: 500, store_id: orderSearch.store_id, status: 1 }).then(res => {
    supplierList.value = res.data?.list || res.data || []
  })
}

function formatAmount(val) {
  if (val == null) return '0.00'
  return Number(val).toFixed(2)
}

function formatDateTime(val) {
  if (!val) return '-'
  const d = new Date(val)
  if (isNaN(d.getTime())) return String(val).slice(0, 16).replace('T', ' ')
  const pad = n => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth()+1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}`
}

function getStatusTagType(status) {
  const map = { 0: 'info', 1: 'warning', 2: 'primary', 3: 'warning', 4: 'success', 5: 'info' }
  return map[status] || ''
}

function fetchOrders() {
  orderLoading.value = true
  const params = {
    store_id: orderSearch.store_id,
    status: orderSearch.status,
    keyword: orderSearch.keyword,
    page: orderPagination.page,
    page_size: orderPagination.pageSize
  }
  supplierApi.getPurchaseOrders(params).then(res => {
    const data = res.data
    orderList.value = (data?.list || data || []).map(o => ({
      ...o,
      status_text: purchaseStatusMap[o.status] || ''
    }))
    orderPagination.total = data?.total || 0
    orderLoading.value = false
  }).catch(() => {
    ElMessage.error('获取采购订单失败')
    orderLoading.value = false
  })
}

function resetOrderSearch() {
  orderSearch.status = 0
  orderSearch.supplier_id = 0
  orderSearch.keyword = ''
  orderPagination.page = 1
  fetchOrders()
}

function fetchReceives() {
  receiveLoading.value = true
  const params = {
    store_id: receiveSearch.store_id,
    supplier_id: receiveSearch.supplier_id,
    keyword: receiveSearch.keyword,
    start_date: dateRange.value?.[0] || '',
    end_date: dateRange.value?.[1] || '',
    page: receivePagination.page,
    page_size: receivePagination.pageSize
  }
  supplierApi.getPurchaseReceives(params).then(res => {
    const data = res.data
    receiveList.value = data?.list || data || []
    receivePagination.total = data?.total || 0
    receiveLoading.value = false
  }).catch(() => {
    ElMessage.error('获取收货记录失败')
    receiveLoading.value = false
  })
}

function resetReceiveSearch() {
  receiveSearch.supplier_id = 0
  receiveSearch.keyword = ''
  dateRange.value = []
  receivePagination.page = 1
  fetchReceives()
}

function handleAddOrder() {
  orderDialogTitle.value = '新建采购订单'
  Object.assign(orderForm, {
    id: 0,
    store_id: orderSearch.store_id,
    supplier_id: 0,
    payment_term: 0,
    expected_date: '',
    remark: '',
    items: []
  })
  orderDialogVisible.value = true
}

function addOrderItem() {
  orderForm.items.push({
    ingredient_id: 0,
    ingredient_name: '',
    category: '',
    unit: '',
    forecast_qty: 0,
    safety_stock_qty: 0,
    current_stock: 0,
    purchase_qty: 1,
    unit_price: 0
  })
}

function removeOrderItem(index) {
  orderForm.items.splice(index, 1)
}

function onSupplierChange(id) {
  const sup = supplierList.value.find(s => s.id === id)
  if (sup && sup.payment_term != null && orderForm.payment_term === 0) {
    orderForm.payment_term = sup.payment_term
  }
}

function openIngredientSelect() {
  ingredientKeyword.value = ''
  ingredientList.value = []
  searchIngredients()
  ingredientSelectVisible.value = true
}

function searchIngredients() {
  ingredientLoading.value = true
  ingredientApi.getIngredients({
    store_id: orderForm.store_id,
    keyword: ingredientKeyword.value,
    status: 1,
    page: 1,
    page_size: 100
  }).then(res => {
    ingredientList.value = res.data?.list || res.data || []
    ingredientLoading.value = false
  }).catch(() => {
    ingredientLoading.value = false
  })
}

function onIngredientSelectionChange(rows) {
  selectedIngredients.value = rows
}

function confirmIngredientSelect() {
  selectedIngredients.value.forEach(ing => {
    if (!orderForm.items.find(i => i.ingredient_id === ing.id)) {
      orderForm.items.push({
        ingredient_id: ing.id,
        ingredient_name: ing.name,
        category: ing.category,
        unit: ing.unit,
        forecast_qty: 0,
        safety_stock_qty: 0,
        current_stock: ing.current_stock || 0,
        purchase_qty: 1,
        unit_price: ing.current_price || 0
      })
    }
  })
  ingredientSelectVisible.value = false
}

function submitOrder() {
  if (!orderForm.store_id) return ElMessage.warning('请选择门店')
  if (!orderForm.supplier_id) return ElMessage.warning('请选择供应商')
  if (!orderForm.items.length) return ElMessage.warning('请添加采购明细')
  for (const item of orderForm.items) {
    if (!item.ingredient_name) return ElMessage.warning('请填写食材名称')
    if (!item.purchase_qty || item.purchase_qty <= 0) return ElMessage.warning('请填写采购数量')
  }
  const sup = supplierList.value.find(s => s.id === orderForm.supplier_id) || {}
  const data = {
    ...orderForm,
    supplier_name: sup.name || '',
    supplier_phone: sup.mobile || sup.phone || '',
    supplier_email: sup.email || ''
  }
  supplierApi.createPurchaseOrder(data).then(() => {
    ElMessage.success('采购订单创建成功')
    orderDialogVisible.value = false
    fetchOrders()
  }).catch(err => ElMessage.error(err.message || '创建失败'))
}

function viewOrderDetail(row) {
  supplierApi.getPurchaseOrder(row.id).then(res => {
    currentOrderDetail.value = {
      ...res.data,
      status_text: purchaseStatusMap[res.data.status] || ''
    }
    orderDetailVisible.value = true
  })
}

function viewReceiveDetail(row) {
  supplierApi.getPurchaseReceive(row.id).then(res => {
    ElMessageBox.alert(
      `收货单号: ${res.data.receive_no}\n采购单: ${res.data.purchase_no}\n供应商: ${res.data.supplier_name}\n总数量: ${res.data.total_qty}\n总金额: ¥${formatAmount(res.data.total_amount)}\n操作人: ${res.data.operator_name || '-'}\n时间: ${formatDateTime(res.data.received_at || res.data.created_at)}`,
      '收货单详情',
      { confirmButtonText: '确定', customClass: 'wide-message-box' }
    )
  })
}

function sendOrder(row) {
  ElMessageBox.confirm(`确定要将采购单 ${row.purchase_no} 发送给供应商吗？系统将通过短信/邮件通知。`, '提示', {
    type: 'info'
  }).then(() => {
    supplierApi.sendPurchaseOrder(row.id, { notify_type: ['sms', 'email'], content: '' }).then(() => {
      ElMessage.success('已发送给供应商')
      fetchOrders()
    }).catch(err => ElMessage.error(err.message || '发送失败'))
  })
}

function completeOrder(row) {
  ElMessageBox.confirm(`确定要将采购单 ${row.purchase_no} 标记为已完成吗？将自动生成应付账款。`, '提示', {
    type: 'warning'
  }).then(() => {
    supplierApi.completePurchaseOrder(row.id).then(() => {
      ElMessage.success('订单已完成，应付账款已生成')
      fetchOrders()
    }).catch(err => ElMessage.error(err.message || '操作失败'))
  })
}

function cancelOrder(row) {
  ElMessageBox.prompt('请输入取消原因（可选）', '取消采购订单', {
    confirmButtonText: '确定取消',
    cancelButtonText: '返回',
    inputPlaceholder: '请输入取消原因',
    type: 'warning'
  }).then(({ value }) => {
    supplierApi.cancelPurchaseOrder(row.id, { remark: value || '' }).then(() => {
      ElMessage.success('已取消')
      fetchOrders()
    }).catch(err => ElMessage.error(err.message || '取消失败'))
  }).catch(() => {})
}

function handleReceive(row) {
  supplierApi.getPurchaseOrder(row.id).then(res => {
    currentOrder.value = {
      ...res.data,
      status_text: purchaseStatusMap[res.data.status] || ''
    }
    receiveItems.value = res.data.items.map(item => ({
      ...item,
      actual_received_qty: Math.max(0, item.purchase_qty - (item.received_qty || 0)),
      qualified_qty: Math.max(0, item.purchase_qty - (item.received_qty || 0)),
      batch_no: '',
      expiry_date: ''
    }))
    receiveForm.remark = ''
    receiveForm.operator_name = ''
    receiveDialogVisible.value = true
  })
}

function submitReceive() {
  const hasReceive = receiveItems.value.some(i => i.actual_received_qty > 0)
  if (!hasReceive) return ElMessage.warning('请填写至少一项收货数量')
  if (!receiveForm.operator_name) return ElMessage.warning('请输入操作人姓名')

  const items = receiveItems.value
    .filter(i => i.actual_received_qty > 0)
    .map(i => ({
      purchase_item_id: i.id,
      ingredient_id: i.ingredient_id,
      ingredient_name: i.ingredient_name,
      category: i.category,
      unit: i.unit,
      purchase_qty: i.purchase_qty,
      received_qty: i.actual_received_qty,
      qualified_qty: i.qualified_qty || i.actual_received_qty,
      rejected_qty: Math.max(0, i.actual_received_qty - (i.qualified_qty || i.actual_received_qty)),
      unit_price: i.unit_price,
      batch_no: i.batch_no,
      expiry_date: i.expiry_date,
      reject_reason: ''
    }))

  const data = {
    store_id: currentOrder.value.store_id,
    purchase_id: currentOrder.value.id,
    receive_type: 'partial',
    remark: receiveForm.remark,
    operator_name: receiveForm.operator_name,
    items
  }

  supplierApi.createPurchaseReceive(data).then(() => {
    ElMessage.success('收货入库成功，库存已更新')
    receiveDialogVisible.value = false
    fetchOrders()
    fetchReceives()
  }).catch(err => ElMessage.error(err.message || '入库失败'))
}
</script>

<style scoped>
.purchase-orders-page { padding: 20px; }
.page-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 20px; }
.page-header h2 { margin: 0; }
.header-actions { display: flex; gap: 10px; }
.search-card { margin-bottom: 20px; }
.table-card { margin-bottom: 20px; }
.pagination { margin-top: 20px; justify-content: flex-end; display: flex; }
.items-toolbar { display: flex; gap: 10px; margin-bottom: 12px; }
.items-table { margin-bottom: 16px; }
.order-total { text-align: right; padding: 12px 6px; font-size: 14px; color: #606266; }
.detail-summary { padding: 16px 6px; text-align: right; }
.receive-info { margin-bottom: 16px; }
.receive-toolbar { margin-bottom: 12px; }
.receive-extra { margin-top: 16px; }
.ingredient-search { display: flex; gap: 10px; margin-bottom: 16px; }
</style>
