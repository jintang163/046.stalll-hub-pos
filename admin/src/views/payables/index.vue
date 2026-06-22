<template>
  <div class="payables-page">
    <div class="page-header">
      <h2>应付账款与对账</h2>
      <div class="header-actions">
        <el-button type="primary" @click="createReconciliation" :disabled="!search.store_id">
          <el-icon><DocumentAdd /></el-icon>
          生成对账单
        </el-button>
        <el-button @click="refreshAll">
          <el-icon><Refresh /></el-icon>
          刷新
        </el-button>
      </div>
    </div>

    <el-row :gutter="20" class="stats-row">
      <el-col :span="6">
        <el-card class="stat-card" shadow="hover">
          <div class="stat-label">应付总额</div>
          <div class="stat-value primary">¥{{ formatAmount(stats.total_payable) }}</div>
          <div class="stat-footer">
            <el-tag size="small">总笔数 {{ stats.payable_count }}</el-tag>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card class="stat-card" shadow="hover">
          <div class="stat-label">已付金额</div>
          <div class="stat-value success">¥{{ formatAmount(stats.total_paid) }}</div>
          <div class="stat-footer">
            <el-tag size="small" type="success">付款笔数 {{ stats.payment_count }}</el-tag>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card class="stat-card" shadow="hover">
          <div class="stat-label">未付金额</div>
          <div class="stat-value warning">¥{{ formatAmount(stats.total_unpaid) }}</div>
          <div class="stat-footer">
            <el-tag size="small" type="warning">待付款 {{ stats.unpaid_count }}</el-tag>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card class="stat-card overdue" shadow="hover">
          <div class="stat-label">逾期应付</div>
          <div class="stat-value danger">¥{{ formatAmount(stats.overdue_amount) }}</div>
          <div class="stat-footer">
            <el-tag size="small" type="danger">逾期 {{ stats.overdue_count }} 笔</el-tag>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <el-tabs v-model="activeTab">
      <el-tab-pane label="应付账款明细" name="payables">
        <el-card class="search-card">
          <el-form :inline="true" :model="search">
            <el-form-item label="门店">
              <el-select v-model="search.store_id" placeholder="全部门店" style="width: 150px" @change="onStoreChange">
                <el-option v-for="s in storeList" :key="s.id" :label="s.name" :value="s.id" />
              </el-select>
            </el-form-item>
            <el-form-item label="供应商">
              <el-select v-model="search.supplier_id" placeholder="全部供应商" clearable filterable style="width: 200px">
                <el-option v-for="s in supplierList" :key="s.id" :label="s.name" :value="s.id" />
              </el-select>
            </el-form-item>
            <el-form-item label="状态">
              <el-select v-model="search.status" placeholder="全部" style="width: 140px">
                <el-option label="全部" :value="0" />
                <el-option v-for="(label, s) in payableStatusMap" :key="s" :label="label" :value="Number(s)" />
              </el-select>
            </el-form-item>
            <el-form-item label="到期日">
              <el-date-picker
                v-model="dueDateRange"
                type="daterange"
                range-separator="至"
                start-placeholder="开始日期"
                end-placeholder="结束日期"
                value-format="YYYY-MM-DD"
                style="width: 260px"
              />
            </el-form-item>
            <el-form-item label="关键词">
              <el-input v-model="search.keyword" placeholder="单据号/供应商" clearable style="width: 180px" />
            </el-form-item>
            <el-form-item>
              <el-button type="primary" @click="fetchPayables">搜索</el-button>
              <el-button @click="resetSearch">重置</el-button>
            </el-form-item>
          </el-form>
        </el-card>

        <el-card class="table-card">
          <template #header>
            <div class="card-header">
              <span>应付账款列表</span>
              <el-button size="small" @click="updateOverdue">更新逾期状态</el-button>
            </div>
          </template>
          <el-table :data="payableList" v-loading="payableLoading" border stripe>
            <el-table-column prop="bill_no" label="单号" width="170" />
            <el-table-column prop="supplier_name" label="供应商" min-width="160" show-overflow-tooltip />
            <el-table-column prop="biz_type_text" label="业务类型" width="100" />
            <el-table-column prop="related_bill_no" label="关联单号" width="160" />
            <el-table-column label="应付金额" width="120">
              <template #default="{ row }">
                <span style="color: #409eff; font-weight: bold">¥{{ formatAmount(row.total_amount) }}</span>
              </template>
            </el-table-column>
            <el-table-column label="已付金额" width="120">
              <template #default="{ row }">
                <span style="color: #67c23a">¥{{ formatAmount(row.paid_amount) }}</span>
              </template>
            </el-table-column>
            <el-table-column label="未付金额" width="120">
              <template #default="{ row }">
                <span :style="{ color: row.unpaid_amount > 0 ? '#e6a23c' : '#67c23a', fontWeight: 'bold' }">
                  ¥{{ formatAmount(row.unpaid_amount) }}
                </span>
              </template>
            </el-table-column>
            <el-table-column prop="bill_date" label="账单日期" width="110" />
            <el-table-column label="到期日" width="110">
              <template #default="{ row }">
                <span :style="{ color: isOverdue(row) ? '#f56c6c' : '' }">{{ row.due_date }}</span>
              </template>
            </el-table-column>
            <el-table-column label="状态" width="110">
              <template #default="{ row }">
                <el-tag :type="row.is_overdue === 1 ? 'danger' : (row.status === 2 ? 'success' : (row.status === 1 ? 'warning' : 'info'))">
                  {{ row.is_overdue === 1 ? '已逾期' : payableStatusMap[row.status] }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column label="操作" width="180" fixed="right">
              <template #default="{ row }">
                <el-button link type="primary" @click="viewPayableDetail(row)">详情</el-button>
                <el-button link type="primary" v-if="row.status !== 2" @click="openPayment(row)">
                  <el-icon><Money /></el-icon>
                  付款
                </el-button>
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
            @size-change="fetchPayables"
            @current-change="fetchPayables"
          />
        </el-card>
      </el-tab-pane>

      <el-tab-pane label="付款记录" name="payments">
        <el-card class="search-card">
          <el-form :inline="true" :model="paymentSearch">
            <el-form-item label="门店">
              <el-select v-model="paymentSearch.store_id" placeholder="全部门店" style="width: 150px">
                <el-option v-for="s in storeList" :key="s.id" :label="s.name" :value="s.id" />
              </el-select>
            </el-form-item>
            <el-form-item label="供应商">
              <el-select v-model="paymentSearch.supplier_id" placeholder="全部供应商" clearable filterable style="width: 200px">
                <el-option v-for="s in supplierList" :key="s.id" :label="s.name" :value="s.id" />
              </el-select>
            </el-form-item>
            <el-form-item label="付款方式">
              <el-select v-model="paymentSearch.payment_method" placeholder="全部" style="width: 140px">
                <el-option label="全部" value="" />
                <el-option v-for="(label, m) in paymentMethodMap" :key="m" :label="label" :value="String(m)" />
              </el-select>
            </el-form-item>
            <el-form-item label="付款日期">
              <el-date-picker
                v-model="paymentDateRange"
                type="daterange"
                range-separator="至"
                start-placeholder="开始日期"
                end-placeholder="结束日期"
                value-format="YYYY-MM-DD"
                style="width: 260px"
              />
            </el-form-item>
            <el-form-item>
              <el-button type="primary" @click="fetchPayments">搜索</el-button>
              <el-button @click="resetPaymentSearch">重置</el-button>
            </el-form-item>
          </el-form>
        </el-card>

        <el-card class="table-card">
          <el-table :data="paymentList" v-loading="paymentLoading" border stripe>
            <el-table-column prop="payment_no" label="付款单号" width="180" />
            <el-table-column prop="supplier_name" label="供应商" min-width="160" show-overflow-tooltip />
            <el-table-column label="付款金额" width="130">
              <template #default="{ row }">
                <span style="color: #67c23a; font-weight: bold">¥{{ formatAmount(row.amount) }}</span>
              </template>
            </el-table-column>
            <el-table-column prop="payment_method_text" label="付款方式" width="100" />
            <el-table-column prop="payment_date" label="付款日期" width="110" />
            <el-table-column prop="transaction_no" label="交易号" width="160" show-overflow-tooltip />
            <el-table-column prop="operator_name" label="操作人" width="100" />
            <el-table-column prop="remark" label="备注" min-width="150" show-overflow-tooltip />
            <el-table-column prop="created_at" label="创建时间" width="160">
              <template #default="{ row }">{{ formatDateTime(row.created_at) }}</template>
            </el-table-column>
            <el-table-column label="操作" width="100">
              <template #default="{ row }">
                <el-button link type="primary" @click="viewPaymentDetail(row)">详情</el-button>
              </template>
            </el-table-column>
          </el-table>

          <el-pagination
            v-model:current-page="paymentPagination.page"
            v-model:page-size="paymentPagination.pageSize"
            :total="paymentPagination.total"
            :page-sizes="[10, 20, 50, 100]"
            layout="total, sizes, prev, pager, next, jumper"
            class="pagination"
            @size-change="fetchPayments"
            @current-change="fetchPayments"
          />
        </el-card>
      </el-tab-pane>

      <el-tab-pane label="对账单" name="reconciliations">
        <el-card class="search-card">
          <el-form :inline="true" :model="reconSearch">
            <el-form-item label="门店">
              <el-select v-model="reconSearch.store_id" placeholder="全部门店" style="width: 150px">
                <el-option v-for="s in storeList" :key="s.id" :label="s.name" :value="s.id" />
              </el-select>
            </el-form-item>
            <el-form-item label="供应商">
              <el-select v-model="reconSearch.supplier_id" placeholder="全部供应商" clearable filterable style="width: 200px">
                <el-option v-for="s in supplierList" :key="s.id" :label="s.name" :value="s.id" />
              </el-select>
            </el-form-item>
            <el-form-item label="状态">
              <el-select v-model="reconSearch.status" placeholder="全部" style="width: 140px">
                <el-option label="全部" :value="0" />
                <el-option v-for="(label, s) in reconStatusMap" :key="s" :label="label" :value="Number(s)" />
              </el-select>
            </el-form-item>
            <el-form-item label="对账期间">
              <el-date-picker
                v-model="reconPeriodRange"
                type="monthrange"
                range-separator="至"
                start-placeholder="开始月"
                end-placeholder="结束月"
                value-format="YYYY-MM"
                style="width: 260px"
              />
            </el-form-item>
            <el-form-item>
              <el-button type="primary" @click="fetchReconciliations">搜索</el-button>
              <el-button @click="resetReconSearch">重置</el-button>
            </el-form-item>
          </el-form>
        </el-card>

        <el-card class="table-card">
          <el-table :data="reconList" v-loading="reconLoading" border stripe>
            <el-table-column prop="recon_no" label="对账单号" width="180" />
            <el-table-column prop="supplier_name" label="供应商" min-width="160" show-overflow-tooltip />
            <el-table-column label="对账期间" width="160">
              <template #default="{ row }">{{ row.period_start }} ~ {{ row.period_end }}</template>
            </el-table-column>
            <el-table-column label="系统应付" width="120">
              <template #default="{ row }">
                <span style="color: #409eff">¥{{ formatAmount(row.system_amount) }}</span>
              </template>
            </el-table-column>
            <el-table-column label="供应商金额" width="120">
              <template #default="{ row }">
                <span v-if="row.supplier_amount != null" style="color: #909399">¥{{ formatAmount(row.supplier_amount) }}</span>
                <span v-else style="color: #c0c4cc">-</span>
              </template>
            </el-table-column>
            <el-table-column label="差额" width="120">
              <template #default="{ row }">
                <span v-if="row.supplier_amount != null" :style="{ color: diffAmount(row) === 0 ? '#67c23a' : '#f56c6c', fontWeight: 'bold' }">
                  {{ diffAmount(row) >= 0 ? '+' : '' }}¥{{ formatAmount(diffAmount(row)) }}
                </span>
                <span v-else style="color: #c0c4cc">-</span>
              </template>
            </el-table-column>
            <el-table-column label="状态" width="100">
              <template #default="{ row }">
                <el-tag :type="reconStatusTagType(row.status)">
                  {{ reconStatusMap[row.status] }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="recon_date" label="对账日期" width="110" />
            <el-table-column prop="operator_name" label="操作人" width="100" />
            <el-table-column label="操作" width="200" fixed="right">
              <template #default="{ row }">
                <el-button link type="primary" @click="viewReconDetail(row)">详情</el-button>
                <el-button link type="primary" v-if="row.status === 0" @click="confirmRecon(row)">确认</el-button>
                <el-button link type="warning" v-if="row.status === 0 || row.status === 1" @click="inputSupplierAmount(row)">供应商金额</el-button>
              </template>
            </el-table-column>
          </el-table>

          <el-pagination
            v-model:current-page="reconPagination.page"
            v-model:page-size="reconPagination.pageSize"
            :total="reconPagination.total"
            :page-sizes="[10, 20, 50, 100]"
            layout="total, sizes, prev, pager, next, jumper"
            class="pagination"
            @size-change="fetchReconciliations"
            @current-change="fetchReconciliations"
          />
        </el-card>
      </el-tab-pane>
    </el-tabs>

    <el-dialog v-model="paymentDialogVisible" title="创建付款记录" width="640px">
      <div v-if="currentPayable" class="payable-info">
        <el-alert :closable="false" type="info" show-icon>
          <template #title>
            当前账单：{{ currentPayable.bill_no }}｜供应商：{{ currentPayable.supplier_name }}｜
            应付金额：<b>¥{{ formatAmount(currentPayable.total_amount) }}</b>｜
            未付：<b style="color:#f56c6c">¥{{ formatAmount(currentPayable.unpaid_amount) }}</b>
          </template>
        </el-alert>
      </div>

      <el-form :model="paymentForm" label-width="100px" style="margin-top: 20px">
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="付款金额" required>
              <el-input-number
                v-model="paymentForm.amount"
                :precision="2"
                :min="0.01"
                :max="currentPayable?.unpaid_amount || 999999"
                style="width: 100%"
              />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="付款方式" required>
              <el-select v-model="paymentForm.payment_method" placeholder="请选择" style="width: 100%">
                <el-option v-for="(label, m) in paymentMethodMap" :key="m" :label="label" :value="Number(m)" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="付款日期" required>
              <el-date-picker v-model="paymentForm.payment_date" type="date" value-format="YYYY-MM-DD" style="width: 100%" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="交易号">
              <el-input v-model="paymentForm.transaction_no" maxlength="100" placeholder="银行/支付流水号" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="操作人" required>
              <el-input v-model="paymentForm.operator_name" maxlength="50" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="凭证附件">
              <el-upload
                action="#"
                :auto-upload="false"
                :limit="1"
                accept="image/*,.pdf"
              >
                <el-button type="primary" size="small">
                  <el-icon><Upload /></el-icon>
                  选择文件
                </el-button>
              </el-upload>
            </el-form-item>
          </el-col>
          <el-col :span="24">
            <el-form-item label="备注">
              <el-input v-model="paymentForm.remark" type="textarea" :rows="2" maxlength="255" />
            </el-form-item>
          </el-col>
        </el-row>
      </el-form>
      <template #footer>
        <el-button @click="paymentDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="submitPayment">确认付款</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="reconDialogVisible" :title="reconDialogTitle" width="560px">
      <el-form :model="reconForm" label-width="100px">
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="供应商" required>
              <el-select v-model="reconForm.supplier_id" placeholder="请选择供应商" filterable style="width: 100%">
                <el-option v-for="s in supplierList" :key="s.id" :label="s.name" :value="s.id" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="期间" required>
              <el-date-picker
                v-model="reconPeriodRange2"
                type="monthrange"
                range-separator="至"
                start-placeholder="开始月"
                end-placeholder="结束月"
                value-format="YYYY-MM"
                style="width: 100%"
              />
            </el-form-item>
          </el-col>
          <el-col :span="24">
            <el-form-item label="操作人" required>
              <el-input v-model="reconForm.operator_name" maxlength="50" />
            </el-form-item>
          </el-col>
        </el-row>
      </el-form>
      <template #footer>
        <el-button @click="reconDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="submitCreateRecon">生成对账单</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="supplierAmountDialogVisible" title="录入供应商对账金额" width="480px">
      <div v-if="currentRecon" class="recon-info">
        <el-descriptions :column="1" border size="small">
          <el-descriptions-item label="对账单号">{{ currentRecon.recon_no }}</el-descriptions-item>
          <el-descriptions-item label="系统应付金额">
            <span style="color: #409eff; font-weight: bold">¥{{ formatAmount(currentRecon.system_amount) }}</span>
          </el-descriptions-item>
        </el-descriptions>
      </div>
      <el-form :model="supplierAmountForm" label-width="120px" style="margin-top: 16px">
        <el-form-item label="供应商对账金额" required>
          <el-input-number v-model="supplierAmountForm.supplier_amount" :precision="2" :min="0" style="width: 100%" />
        </el-form-item>
        <el-form-item label="差异说明">
          <el-input v-model="supplierAmountForm.difference_remark" type="textarea" :rows="3" maxlength="500" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="supplierAmountDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="submitSupplierAmount">确认</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { DocumentAdd, Refresh, Money, Upload } from '@element-plus/icons-vue'
import * as supplierApi from '@/api/supplier'
import { storeApi } from '@/api/stores'

const activeTab = ref('payables')

const payableLoading = ref(false)
const paymentLoading = ref(false)
const reconLoading = ref(false)

const paymentDialogVisible = ref(false)
const reconDialogVisible = ref(false)
const supplierAmountDialogVisible = ref(false)

const reconDialogTitle = ref('生成对账单')

const currentPayable = ref(null)
const currentRecon = ref(null)

const payableStatusMap = {
  0: '未付',
  1: '部分支付',
  2: '已付清',
  3: '坏账'
}

const paymentMethodMap = {
  1: '银行转账',
  2: '微信',
  3: '支付宝',
  4: '现金',
  5: '承兑',
  6: '其他'
}

const reconStatusMap = {
  0: '待对账',
  1: '差异',
  2: '已对账',
  3: '已取消'
}

const storeList = ref([])
const supplierList = ref([])

const dueDateRange = ref([])
const paymentDateRange = ref([])
const reconPeriodRange = ref([])
const reconPeriodRange2 = ref([])

const search = reactive({
  store_id: 0,
  supplier_id: 0,
  status: 0,
  keyword: ''
})

const paymentSearch = reactive({
  store_id: 0,
  supplier_id: 0,
  payment_method: '',
  keyword: ''
})

const reconSearch = reactive({
  store_id: 0,
  supplier_id: 0,
  status: 0,
  period_start: '',
  period_end: ''
})

const pagination = reactive({ page: 1, pageSize: 20, total: 0 })
const paymentPagination = reactive({ page: 1, pageSize: 20, total: 0 })
const reconPagination = reactive({ page: 1, pageSize: 20, total: 0 })

const stats = reactive({
  total_payable: 0,
  total_paid: 0,
  total_unpaid: 0,
  overdue_amount: 0,
  payable_count: 0,
  payment_count: 0,
  unpaid_count: 0,
  overdue_count: 0
})

const payableList = ref([])
const paymentList = ref([])
const reconList = ref([])

const paymentForm = reactive({
  store_id: 0,
  payable_ids: [],
  supplier_id: 0,
  amount: 0,
  payment_method: 1,
  payment_date: '',
  transaction_no: '',
  operator_name: '',
  receipt_url: '',
  remark: ''
})

const reconForm = reactive({
  supplier_id: 0,
  period_start: '',
  period_end: '',
  operator_name: ''
})

const supplierAmountForm = reactive({
  supplier_amount: 0,
  difference_remark: ''
})

onMounted(() => {
  loadStores()
})

function loadStores() {
  storeApi.list({ page: 1, page_size: 100 }).then(res => {
    storeList.value = res.data?.list || res.data || []
    if (storeList.value.length > 0) {
      search.store_id = storeList.value[0].id
      paymentSearch.store_id = storeList.value[0].id
      reconSearch.store_id = storeList.value[0].id
      paymentForm.store_id = storeList.value[0].id
    }
    loadSuppliers()
    fetchStats()
    fetchPayables()
    fetchPayments()
    fetchReconciliations()
  }).catch(() => {
    storeList.value = [{ id: 1, name: '默认门店' }]
    search.store_id = 1
    paymentSearch.store_id = 1
    reconSearch.store_id = 1
    paymentForm.store_id = 1
    loadSuppliers()
    fetchStats()
    fetchPayables()
    fetchPayments()
    fetchReconciliations()
  })
}

function loadSuppliers() {
  supplierApi.getSuppliers({ page: 1, page_size: 500, store_id: search.store_id, status: 1 }).then(res => {
    supplierList.value = res.data?.list || res.data || []
  })
}

function onStoreChange() {
  loadSuppliers()
  refreshAll()
}

function refreshAll() {
  fetchStats()
  fetchPayables()
  fetchPayments()
  fetchReconciliations()
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

function todayStr() {
  const d = new Date()
  const pad = n => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth()+1)}-${pad(d.getDate())}`
}

function isOverdue(row) {
  if (row.is_overdue === 1) return true
  if (row.status === 2) return false
  if (!row.due_date) return false
  return new Date(row.due_date) < new Date(todayStr())
}

function diffAmount(row) {
  if (row.supplier_amount == null) return 0
  return Number(row.system_amount) - Number(row.supplier_amount)
}

function reconStatusTagType(status) {
  const map = { 0: 'warning', 1: 'danger', 2: 'success', 3: 'info' }
  return map[status] || ''
}

function fetchStats() {
  supplierApi.getSupplierStats({ store_id: search.store_id }).then(res => {
    const d = res.data || {}
    Object.assign(stats, d)
  })
}

function fetchPayables() {
  payableLoading.value = true
  supplierApi.getAccountsPayable({
    store_id: search.store_id,
    supplier_id: search.supplier_id,
    status: search.status,
    due_start: dueDateRange.value?.[0] || '',
    due_end: dueDateRange.value?.[1] || '',
    keyword: search.keyword,
    page: pagination.page,
    page_size: pagination.pageSize
  }).then(res => {
    const data = res.data
    payableList.value = (data?.list || data || []).map(p => ({
      ...p,
      biz_type_text: p.biz_type === 1 ? '采购入库' : (p.biz_type === 2 ? '退货' : '其他')
    }))
    pagination.total = data?.total || 0
    payableLoading.value = false
  }).catch(() => {
    payableLoading.value = false
  })
}

function resetSearch() {
  search.supplier_id = 0
  search.status = 0
  search.keyword = ''
  dueDateRange.value = []
  pagination.page = 1
  fetchPayables()
}

function fetchPayments() {
  paymentLoading.value = true
  supplierApi.getPayablePayments({
    store_id: paymentSearch.store_id,
    supplier_id: paymentSearch.supplier_id,
    payment_method: Number(paymentSearch.payment_method) || 0,
    start_date: paymentDateRange.value?.[0] || '',
    end_date: paymentDateRange.value?.[1] || '',
    page: paymentPagination.page,
    page_size: paymentPagination.pageSize
  }).then(res => {
    const data = res.data
    paymentList.value = (data?.list || data || []).map(p => ({
      ...p,
      payment_method_text: paymentMethodMap[p.payment_method] || '-'
    }))
    paymentPagination.total = data?.total || 0
    paymentLoading.value = false
  }).catch(() => {
    paymentLoading.value = false
  })
}

function resetPaymentSearch() {
  paymentSearch.supplier_id = 0
  paymentSearch.payment_method = ''
  paymentDateRange.value = []
  paymentPagination.page = 1
  fetchPayments()
}

function fetchReconciliations() {
  reconLoading.value = true
  supplierApi.getReconciliations({
    store_id: reconSearch.store_id,
    supplier_id: reconSearch.supplier_id,
    status: reconSearch.status,
    period_start: reconPeriodRange.value?.[0] || '',
    period_end: reconPeriodRange.value?.[1] || '',
    page: reconPagination.page,
    page_size: reconPagination.pageSize
  }).then(res => {
    const data = res.data
    reconList.value = data?.list || data || []
    reconPagination.total = data?.total || 0
    reconLoading.value = false
  }).catch(() => {
    reconLoading.value = false
  })
}

function resetReconSearch() {
  reconSearch.supplier_id = 0
  reconSearch.status = 0
  reconPeriodRange.value = []
  reconPagination.page = 1
  fetchReconciliations()
}

function updateOverdue() {
  supplierApi.updatePayableOverdue({ store_id: search.store_id }).then(() => {
    ElMessage.success('逾期状态已更新')
    fetchStats()
    fetchPayables()
  })
}

function viewPayableDetail(row) {
  supplierApi.getAccountsPayableItem(row.id).then(res => {
    const d = res.data || row
    const itemsHtml = (d.items || []).map((it, idx) => {
      return `${idx + 1}. ${it.bill_no || '子单'}｜${it.biz_type === 1 ? '采购' : (it.biz_type === 2 ? '退货' : '其他')}｜金额 ¥${formatAmount(it.amount)}｜日期 ${it.bill_date || '-'}`
    }).join('\n')
    ElMessageBox.alert(
      `单号: ${d.bill_no}\n供应商: ${d.supplier_name}\n业务类型: ${d.biz_type === 1 ? '采购入库' : (d.biz_type === 2 ? '退货' : '其他')}\n关联单号: ${d.related_bill_no || '-'}\n\n应付金额: ¥${formatAmount(d.total_amount)}\n已付金额: ¥${formatAmount(d.paid_amount)}\n未付金额: ¥${formatAmount(d.unpaid_amount)}\n\n账单日期: ${d.bill_date}\n到期日: ${d.due_date}\n状态: ${d.is_overdue === 1 ? '已逾期' : payableStatusMap[d.status]}\n\n${itemsHtml ? '明细:\n' + itemsHtml : ''}\n${d.remark ? '\n备注: ' + d.remark : ''}`,
      '应付账款详情',
      { confirmButtonText: '确定', customClass: 'wide-message-box' }
    )
  })
}

function viewPaymentDetail(row) {
  ElMessageBox.alert(
    `付款单号: ${row.payment_no}\n供应商: ${row.supplier_name}\n付款金额: ¥${formatAmount(row.amount)}\n付款方式: ${paymentMethodMap[row.payment_method] || '-'}\n付款日期: ${row.payment_date}\n交易号: ${row.transaction_no || '-'}\n操作人: ${row.operator_name || '-'}\n${row.remark ? '备注: ' + row.remark : ''}`,
    '付款详情',
    { confirmButtonText: '确定' }
  )
}

function openPayment(row) {
  currentPayable.value = row
  Object.assign(paymentForm, {
    store_id: search.store_id,
    payable_ids: [row.id],
    supplier_id: row.supplier_id,
    amount: Number(row.unpaid_amount) || 0,
    payment_method: 1,
    payment_date: todayStr(),
    transaction_no: '',
    operator_name: '',
    receipt_url: '',
    remark: ''
  })
  paymentDialogVisible.value = true
}

function submitPayment() {
  if (!paymentForm.amount || paymentForm.amount <= 0) return ElMessage.warning('请填写付款金额')
  if (!paymentForm.payment_method) return ElMessage.warning('请选择付款方式')
  if (!paymentForm.payment_date) return ElMessage.warning('请选择付款日期')
  if (!paymentForm.operator_name) return ElMessage.warning('请填写操作人')
  supplierApi.createPayablePayment(paymentForm).then(() => {
    ElMessage.success('付款记录创建成功')
    paymentDialogVisible.value = false
    refreshAll()
  }).catch(err => ElMessage.error(err.message || '创建失败'))
}

function createReconciliation() {
  reconForm = Object.assign(reconForm, {
    supplier_id: 0,
    operator_name: ''
  })
  reconPeriodRange2.value = []
  reconDialogVisible.value = true
}

function submitCreateRecon() {
  if (!reconForm.supplier_id) return ElMessage.warning('请选择供应商')
  if (!reconPeriodRange2.value?.length || reconPeriodRange2.value.length < 2) return ElMessage.warning('请选择对账期间')
  if (!reconForm.operator_name) return ElMessage.warning('请填写操作人')
  Object.assign(reconForm, {
    period_start: reconPeriodRange2.value[0],
    period_end: reconPeriodRange2.value[1]
  })
  supplierApi.createReconciliation({ ...reconForm, store_id: reconSearch.store_id }).then(() => {
    ElMessage.success('对账单生成成功')
    reconDialogVisible.value = false
    fetchReconciliations()
  }).catch(err => ElMessage.error(err.message || '生成失败'))
}

function viewReconDetail(row) {
  supplierApi.getReconciliation(row.id).then(res => {
    const d = res.data || row
    const itemsHtml = (d.items || []).map((it, idx) => {
      return `${idx + 1}. ${it.bill_no || '单号'}｜${it.biz_type_text || it.biz_type === 1 ? '采购' : '其他'}｜系统 ¥${formatAmount(it.system_amount)}｜供应商 ${it.supplier_amount != null ? '¥' + formatAmount(it.supplier_amount) : '-'}｜备注 ${it.remark || '-'}`
    }).join('\n')
    ElMessageBox.alert(
      `对账单号: ${d.recon_no}\n供应商: ${d.supplier_name}\n对账期间: ${d.period_start} ~ ${d.period_end}\n\n系统应付: ¥${formatAmount(d.system_amount)}\n供应商金额: ${d.supplier_amount != null ? '¥' + formatAmount(d.supplier_amount) : '未录入'}\n差额: ${d.supplier_amount != null ? '¥' + formatAmount(Number(d.system_amount) - Number(d.supplier_amount)) : '-'}\n\n状态: ${reconStatusMap[d.status]}\n对账日期: ${d.recon_date || '-'}\n操作人: ${d.operator_name || '-'}\n${d.difference_remark ? '差异说明: ' + d.difference_remark : ''}\n\n${itemsHtml ? '明细:\n' + itemsHtml : ''}`,
      '对账单详情',
      { confirmButtonText: '确定', customClass: 'wide-message-box' }
    )
  })
}

function confirmRecon(row) {
  ElMessageBox.confirm(`确认对账单 ${row.recon_no} 已对账完成？`, '提示', { type: 'warning' }).then(() => {
    supplierApi.confirmReconciliation(row.id).then(() => {
      ElMessage.success('对账已确认')
      fetchReconciliations()
    }).catch(err => ElMessage.error(err.message || '确认失败'))
  })
}

function inputSupplierAmount(row) {
  currentRecon.value = row
  supplierAmountForm.supplier_amount = row.supplier_amount || 0
  supplierAmountForm.difference_remark = row.difference_remark || ''
  supplierAmountDialogVisible.value = true
}

function submitSupplierAmount() {
  if (!currentRecon.value) return
  supplierApi.inputSupplierReconAmount(currentRecon.value.id, { ...supplierAmountForm }).then(() => {
    ElMessage.success('供应商金额已录入')
    supplierAmountDialogVisible.value = false
    fetchReconciliations()
  }).catch(err => ElMessage.error(err.message || '录入失败'))
}
</script>

<style scoped>
.payables-page { padding: 20px; }
.page-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 20px; }
.page-header h2 { margin: 0; }
.header-actions { display: flex; gap: 10px; }
.stats-row { margin-bottom: 20px; }
.stat-card { border-left: 4px solid #409eff; }
.stat-card.overdue { border-left: 4px solid #f56c6c; }
.stat-label { font-size: 13px; color: #909399; margin-bottom: 8px; }
.stat-value { font-size: 26px; font-weight: bold; margin-bottom: 8px; }
.stat-value.primary { color: #409eff; }
.stat-value.success { color: #67c23a; }
.stat-value.warning { color: #e6a23c; }
.stat-value.danger { color: #f56c6c; }
.stat-footer { border-top: 1px solid #ebeef5; padding-top: 8px; }
.search-card { margin-bottom: 20px; }
.table-card { margin-bottom: 20px; }
.card-header { display: flex; justify-content: space-between; align-items: center; }
.pagination { margin-top: 20px; justify-content: flex-end; display: flex; }
.payable-info { margin-bottom: 10px; }
.recon-info { margin-bottom: 10px; }
</style>
