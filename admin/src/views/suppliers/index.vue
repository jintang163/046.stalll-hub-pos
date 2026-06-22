<template>
  <div class="suppliers-page">
    <div class="page-header">
      <h2>供应商管理</h2>
      <div class="header-actions">
        <el-button type="primary" @click="handleAdd">
          <el-icon><Plus /></el-icon>
          新增供应商
        </el-button>
      </div>
    </div>

    <el-row :gutter="16" class="stats-row" v-if="statsLoaded">
      <el-col :span="6">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-item">
            <div class="stat-label">供应商总数</div>
            <div class="stat-value primary">{{ supplierStats.total_supplier }}</div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-item">
            <div class="stat-label">合作中</div>
            <div class="stat-value success">{{ supplierStats.active_supplier }}</div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-item">
            <div class="stat-label">应付总额</div>
            <div class="stat-value warning">¥{{ formatAmount(supplierStats.total_payable) }}</div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-item">
            <div class="stat-label">逾期应付</div>
            <div class="stat-value danger">¥{{ formatAmount(supplierStats.overdue_payable) }}</div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <el-card class="search-card">
      <el-form :inline="true" :model="searchForm">
        <el-form-item label="门店">
          <el-select v-model="searchForm.store_id" placeholder="全部门店" style="width: 150px">
            <el-option v-for="store in storeList" :key="store.id" :label="store.name" :value="store.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="分类">
          <el-select v-model="searchForm.category" placeholder="全部分类" clearable style="width: 150px">
            <el-option v-for="cat in categories" :key="cat" :label="cat" :value="cat" />
          </el-select>
        </el-form-item>
        <el-form-item label="关键词">
          <el-input v-model="searchForm.keyword" placeholder="供应商名称/编号/联系人/电话" clearable style="width: 220px" />
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="searchForm.status" placeholder="全部" style="width: 120px">
            <el-option label="全部" :value="-1" />
            <el-option label="启用" :value="1" />
            <el-option label="禁用" :value="0" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="handleSearch">搜索</el-button>
          <el-button @click="handleReset">重置</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <el-card class="table-card">
      <el-table :data="tableData" v-loading="loading" border stripe>
        <el-table-column prop="supplier_no" label="供应商编号" width="140" />
        <el-table-column prop="name" label="供应商名称" min-width="180" show-overflow-tooltip />
        <el-table-column prop="category" label="分类" width="100" />
        <el-table-column prop="contact_person" label="联系人" width="100" />
        <el-table-column label="联系方式" width="180">
          <template #default="{ row }">
            <div>{{ row.mobile || row.phone }}</div>
            <div style="color: #909399; font-size: 12px">{{ row.email }}</div>
          </template>
        </el-table-column>
        <el-table-column prop="payment_term_text" label="账期" width="100" />
        <el-table-column prop="settlement_method" label="结算方式" width="110">
          <template #default="{ row }">
            {{ settlementMethodMap[row.settlement_method] || row.settlement_method }}
          </template>
        </el-table-column>
        <el-table-column label="应付金额" width="120">
          <template #default="{ row }">
            <span style="color: #e6a23c; font-weight: bold">¥{{ formatAmount(row.current_payable) }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="status_text" label="状态" width="80">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'info'">
              {{ row.status_text }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="260" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="handleView(row)">详情</el-button>
            <el-button link type="primary" @click="handleNotify(row)">通知</el-button>
            <el-button link type="primary" @click="handleEdit(row)">编辑</el-button>
            <el-button link type="danger" @click="handleDelete(row)">删除</el-button>
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

    <el-dialog v-model="dialogVisible" :title="dialogTitle" width="720px" top="5vh">
      <el-form :model="form" label-width="110px" v-if="dialogVisible">
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="门店">
              <el-select v-model="form.store_id" placeholder="请选择门店" style="width: 100%">
                <el-option v-for="store in storeList" :key="store.id" :label="store.name" :value="store.id" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="供应商名称" required>
              <el-input v-model="form.name" placeholder="请输入供应商名称" maxlength="100" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="简称">
              <el-input v-model="form.short_name" placeholder="请输入简称" maxlength="50" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="分类">
              <el-input v-model="form.category" placeholder="如：蔬菜、肉类、水产" maxlength="50" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="联系人">
              <el-input v-model="form.contact_person" placeholder="请输入联系人" maxlength="50" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="手机号">
              <el-input v-model="form.mobile" placeholder="请输入手机号" maxlength="20" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="电话">
              <el-input v-model="form.phone" placeholder="请输入固定电话" maxlength="20" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="邮箱">
              <el-input v-model="form.email" placeholder="请输入邮箱" maxlength="100" />
            </el-form-item>
          </el-col>
          <el-col :span="24">
            <el-form-item label="地址">
              <el-input v-model="form.address" placeholder="请输入详细地址" maxlength="255" />
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item label="省">
              <el-input v-model="form.province" placeholder="省份" maxlength="50" />
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item label="市">
              <el-input v-model="form.city" placeholder="城市" maxlength="50" />
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item label="区/县">
              <el-input v-model="form.district" placeholder="区/县" maxlength="50" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="账期">
              <el-select v-model="form.payment_term" placeholder="请选择账期" style="width: 100%">
                <el-option v-for="(label, days) in paymentTermMap" :key="days" :label="label" :value="days" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="结算方式">
              <el-select v-model="form.settlement_method" placeholder="请选择结算方式" style="width: 100%">
                <el-option v-for="(label, key) in settlementMethodMap" :key="key" :label="label" :value="key" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="信用额度">
              <el-input-number v-model="form.credit_limit" :precision="2" :min="0" style="width: 100%" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="状态">
              <el-radio-group v-model="form.status">
                <el-radio :value="1">启用</el-radio>
                <el-radio :value="0">禁用</el-radio>
              </el-radio-group>
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item label="开户银行">
              <el-input v-model="form.bank_name" placeholder="银行名称" maxlength="100" />
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item label="银行账号">
              <el-input v-model="form.bank_account" placeholder="银行账号" maxlength="50" />
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item label="户名">
              <el-input v-model="form.bank_account_name" placeholder="账户名称" maxlength="100" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="税号">
              <el-input v-model="form.tax_no" placeholder="统一社会信用代码" maxlength="50" />
            </el-form-item>
          </el-col>
          <el-col :span="24">
            <el-form-item label="备注">
              <el-input v-model="form.remark" type="textarea" :rows="2" maxlength="255" />
            </el-form-item>
          </el-col>
        </el-row>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSubmit">确定</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="detailVisible" title="供应商详情" width="640px">
      <div v-if="currentSupplier" class="supplier-detail">
        <el-descriptions :column="2" border>
          <el-descriptions-item label="供应商编号">{{ currentSupplier.supplier_no }}</el-descriptions-item>
          <el-descriptions-item label="名称">{{ currentSupplier.name }}</el-descriptions-item>
          <el-descriptions-item label="简称">{{ currentSupplier.short_name || '-' }}</el-descriptions-item>
          <el-descriptions-item label="分类">{{ currentSupplier.category || '-' }}</el-descriptions-item>
          <el-descriptions-item label="联系人">{{ currentSupplier.contact_person || '-' }}</el-descriptions-item>
          <el-descriptions-item label="手机">{{ currentSupplier.mobile || '-' }}</el-descriptions-item>
          <el-descriptions-item label="电话">{{ currentSupplier.phone || '-' }}</el-descriptions-item>
          <el-descriptions-item label="邮箱">{{ currentSupplier.email || '-' }}</el-descriptions-item>
          <el-descriptions-item label="账期">{{ currentSupplier.payment_term_desc || paymentTermMap[currentSupplier.payment_term] || '-' }}</el-descriptions-item>
          <el-descriptions-item label="结算方式">{{ settlementMethodMap[currentSupplier.settlement_method] || '-' }}</el-descriptions-item>
          <el-descriptions-item label="信用额度">¥{{ formatAmount(currentSupplier.credit_limit) }}</el-descriptions-item>
          <el-descriptions-item label="状态">
            <el-tag :type="currentSupplier.status === 1 ? 'success' : 'info'">
              {{ currentSupplier.status_text }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="地址" :span="2">{{ currentSupplier.address || '-' }}</el-descriptions-item>
          <el-descriptions-item label="开户行">{{ currentSupplier.bank_name || '-' }}</el-descriptions-item>
          <el-descriptions-item label="账号">{{ currentSupplier.bank_account || '-' }}</el-descriptions-item>
          <el-descriptions-item label="开户名" :span="2">{{ currentSupplier.bank_account_name || '-' }}</el-descriptions-item>
          <el-descriptions-item label="税号" :span="2">{{ currentSupplier.tax_no || '-' }}</el-descriptions-item>
          <el-descriptions-item label="累计采购">
            <span style="color: #409eff; font-weight: bold">¥{{ formatAmount(currentSupplier.total_purchase) }}</span>
          </el-descriptions-item>
          <el-descriptions-item label="已付款">
            <span style="color: #67c23a; font-weight: bold">¥{{ formatAmount(currentSupplier.total_paid) }}</span>
          </el-descriptions-item>
          <el-descriptions-item label="当前应付" :span="2">
            <span style="color: #e6a23c; font-weight: bold; font-size: 16px">¥{{ formatAmount(currentSupplier.current_payable) }}</span>
          </el-descriptions-item>
          <el-descriptions-item label="备注" :span="2">{{ currentSupplier.remark || '-' }}</el-descriptions-item>
        </el-descriptions>
      </div>
    </el-dialog>

    <el-dialog v-model="notifyVisible" title="通知供应商" width="480px">
      <el-form label-width="90px">
        <el-form-item label="通知对象">
          <span>{{ currentSupplier?.name }}</span>
        </el-form-item>
        <el-form-item label="通知方式">
          <el-checkbox-group v-model="notifyForm.notify_type">
            <el-checkbox value="sms">短信</el-checkbox>
            <el-checkbox value="email">邮件</el-checkbox>
          </el-checkbox-group>
        </el-form-item>
        <el-form-item label="通知内容">
          <el-input v-model="notifyForm.content" type="textarea" :rows="4" placeholder="请输入通知内容" maxlength="500" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="notifyVisible = false">取消</el-button>
        <el-button type="primary" @click="submitNotify">发送</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import * as supplierApi from '@/api/supplier'
import { storeApi } from '@/api/stores'

const loading = ref(false)
const statsLoaded = ref(false)
const dialogVisible = ref(false)
const dialogTitle = ref('新增供应商')
const detailVisible = ref(false)
const notifyVisible = ref(false)

const currentSupplier = ref(null)

const searchForm = reactive({
  store_id: 0,
  category: '',
  keyword: '',
  status: -1
})

const pagination = reactive({
  page: 1,
  pageSize: 20,
  total: 0
})

const tableData = ref([])
const storeList = ref([])
const categories = ref([])
const supplierStats = reactive({
  total_supplier: 0,
  active_supplier: 0,
  total_payable: 0,
  overdue_payable: 0
})

const enums = ref({
  payment_terms: {},
  settlement_methods: {}
})

const paymentTermMap = computed(() => enums.value.payment_terms || {})
const settlementMethodMap = computed(() => enums.value.settlement_methods || {})

const form = reactive({
  id: 0,
  store_id: 0,
  name: '',
  short_name: '',
  category: '',
  contact_person: '',
  phone: '',
  mobile: '',
  email: '',
  fax: '',
  address: '',
  province: '',
  city: '',
  district: '',
  bank_name: '',
  bank_account: '',
  bank_account_name: '',
  tax_no: '',
  payment_term: 0,
  payment_term_desc: '',
  settlement_method: 'bank_transfer',
  credit_limit: 0,
  status: 1,
  remark: ''
})

const notifyForm = reactive({
  notify_type: ['sms'],
  content: ''
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
    if (storeList.value.length > 0 && !searchForm.store_id) {
      searchForm.store_id = storeList.value[0].id
    }
    loadCategories()
    fetchList()
    fetchStats()
  }).catch(() => {
    storeList.value = [{ id: 1, name: '默认门店' }]
    searchForm.store_id = 1
    loadCategories()
    fetchList()
    fetchStats()
  })
}

function loadCategories() {
  supplierApi.getSupplierCategories({ store_id: searchForm.store_id }).then(res => {
    categories.value = res.data || []
  })
}

function fetchStats() {
  supplierApi.getSupplierStats({ store_id: searchForm.store_id }).then(res => {
    Object.assign(supplierStats, res.data || {})
    statsLoaded.value = true
  })
}

function fetchList() {
  loading.value = true
  const params = {
    store_id: searchForm.store_id,
    category: searchForm.category,
    keyword: searchForm.keyword,
    status: searchForm.status,
    page: pagination.page,
    page_size: pagination.pageSize
  }
  supplierApi.getSuppliers(params).then(res => {
    const data = res.data
    tableData.value = data?.list || data || []
    pagination.total = data?.total || 0
    loading.value = false
  }).catch(() => {
    ElMessage.error('获取供应商列表失败')
    loading.value = false
  })
}

function formatAmount(val) {
  if (val == null) return '0.00'
  return Number(val).toFixed(2)
}

function handleSearch() {
  pagination.page = 1
  fetchList()
  fetchStats()
}

function handleReset() {
  searchForm.category = ''
  searchForm.keyword = ''
  searchForm.status = -1
  pagination.page = 1
  fetchList()
}

function handleAdd() {
  dialogTitle.value = '新增供应商'
  Object.assign(form, {
    id: 0,
    store_id: searchForm.store_id,
    name: '',
    short_name: '',
    category: '',
    contact_person: '',
    phone: '',
    mobile: '',
    email: '',
    fax: '',
    address: '',
    province: '',
    city: '',
    district: '',
    bank_name: '',
    bank_account: '',
    bank_account_name: '',
    tax_no: '',
    payment_term: 0,
    payment_term_desc: '',
    settlement_method: 'bank_transfer',
    credit_limit: 0,
    status: 1,
    remark: ''
  })
  dialogVisible.value = true
}

function handleEdit(row) {
  dialogTitle.value = '编辑供应商'
  Object.assign(form, row)
  dialogVisible.value = true
}

function handleView(row) {
  currentSupplier.value = row
  detailVisible.value = true
}

function handleNotify(row) {
  currentSupplier.value = row
  notifyForm.notify_type = ['sms']
  notifyForm.content = ''
  notifyVisible.value = true
}

function submitNotify() {
  if (!notifyForm.notify_type.length) {
    ElMessage.warning('请选择通知方式')
    return
  }
  if (!notifyForm.content) {
    ElMessage.warning('请输入通知内容')
    return
  }
  supplierApi.notifySupplier(currentSupplier.value.id, notifyForm).then(() => {
    ElMessage.success('通知已发送')
    notifyVisible.value = false
  }).catch(err => {
    ElMessage.error(err.message || '发送失败')
  })
}

function handleSubmit() {
  if (!form.name) {
    ElMessage.warning('请输入供应商名称')
    return
  }
  const api = form.id ? supplierApi.updateSupplier(form.id, form) : supplierApi.createSupplier(form)
  api.then(() => {
    ElMessage.success('保存成功')
    dialogVisible.value = false
    fetchList()
    fetchStats()
    loadCategories()
  }).catch(err => {
    ElMessage.error(err.message || '保存失败')
  })
}

function handleDelete(row) {
  ElMessageBox.confirm('确定要删除该供应商吗？删除后将无法恢复。', '提示', {
    type: 'warning'
  }).then(() => {
    supplierApi.deleteSupplier(row.id).then(() => {
      ElMessage.success('删除成功')
      fetchList()
      fetchStats()
    })
  })
}
</script>

<style scoped>
.suppliers-page {
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

.stats-row {
  margin-bottom: 20px;
}

.stat-card {
  height: 100px;
}

.stat-item {
  display: flex;
  flex-direction: column;
  justify-content: center;
  height: 100%;
}

.stat-label {
  color: #909399;
  font-size: 14px;
  margin-bottom: 8px;
}

.stat-value {
  font-size: 22px;
  font-weight: bold;
}

.stat-value.primary { color: #409eff; }
.stat-value.success { color: #67c23a; }
.stat-value.warning { color: #e6a23c; }
.stat-value.danger { color: #f56c6c; }

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

.supplier-detail {
  padding: 10px 0;
}
</style>
