<template>
  <div class="reservations-page">
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
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Search, Refresh, Check, Close } from '@element-plus/icons-vue'
import { reservationApi, tableApi } from '@/api/tables'
import { storeApi } from '@/api/stores'

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

onMounted(() => {
  fetchStoreList()
  fetchList()
})
</script>

<style scoped lang="scss">
.reservations-page {
  .header-actions {
    display: flex;
    gap: 12px;
  }

  .pagination {
    margin-top: 20px;
    display: flex;
    justify-content: flex-end;
  }
}
</style>
