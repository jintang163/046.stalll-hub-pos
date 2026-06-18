<template>
  <div class="stores-page">
    <div class="page-header">
      <h2 class="page-title">门店管理</h2>
      <div class="header-actions">
        <el-button type="success" @click="openDialog">
          <el-icon><Plus /></el-icon>新增门店
        </el-button>
      </div>
    </div>

    <div class="card-wrapper">
      <div class="search-bar">
        <el-input
          v-model="query.name"
          placeholder="搜索门店名称"
          clearable
          style="width: 240px"
          @keyup.enter="fetchList" />
        <el-select v-model="query.status" placeholder="门店状态" clearable style="width: 140px">
          <el-option label="营业中" :value="1" />
          <el-option label="已停业" :value="0" />
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
        <el-table-column prop="name" label="门店名称" min-width="180" />
        <el-table-column prop="code" label="门店编码" width="140" />
        <el-table-column prop="address" label="门店地址" min-width="240" />
        <el-table-column prop="contact" label="联系人" width="120" />
        <el-table-column prop="phone" label="联系电话" width="140" />
        <el-table-column prop="business_hours" label="营业时间" width="200" />
        <el-table-column prop="status" label="状态" width="100" align="center">
          <template #default="{ row }">
            <el-switch
              v-model="row.status"
              :active-value="1"
              :inactive-value="0"
              @change="handleStatusChange(row)" />
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
      :title="isEdit ? '编辑门店' : '新增门店'"
      width="700px"
      :close-on-click-modal="false">
      <el-form
        ref="storeFormRef"
        :model="storeForm"
        :rules="storeRules"
        label-width="100px">
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="门店名称" prop="name">
              <el-input v-model="storeForm.name" placeholder="请输入门店名称" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="门店编码" prop="code">
              <el-input v-model="storeForm.code" placeholder="请输入门店编码" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-form-item label="门店地址" prop="address">
          <el-input v-model="storeForm.address" placeholder="请输入门店地址" />
        </el-form-item>
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="联系人" prop="contact">
              <el-input v-model="storeForm.contact" placeholder="请输入联系人" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="联系电话" prop="phone">
              <el-input v-model="storeForm.phone" placeholder="请输入联系电话" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="营业开始时间" prop="open_time">
              <el-time-picker
                v-model="storeForm.open_time"
                format="HH:mm"
                value-format="HH:mm"
                placeholder="选择开始时间"
                style="width: 100%" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="营业结束时间" prop="close_time">
              <el-time-picker
                v-model="storeForm.close_time"
                format="HH:mm"
                value-format="HH:mm"
                placeholder="选择结束时间"
                style="width: 100%" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-form-item label="门店状态">
          <el-switch v-model="storeForm.status" :active-value="1" :inactive-value="0" active-text="营业中" inactive-text="已停业" />
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
  getStoreList,
  createStore,
  updateStore,
  deleteStore
} from '@/api/stores'

const loading = ref(false)
const submitLoading = ref(false)
const list = ref([])
const total = ref(0)

const query = reactive({
  name: '',
  status: null,
  page: 1,
  page_size: 10
})

const dialogVisible = ref(false)
const isEdit = ref(false)
const editId = ref(null)
const storeFormRef = ref()

const storeForm = reactive({
  name: '',
  code: '',
  address: '',
  contact: '',
  phone: '',
  open_time: '08:00',
  close_time: '22:00',
  status: 1
})

const storeRules = {
  name: [{ required: true, message: '请输入门店名称', trigger: 'blur' }],
  code: [{ required: true, message: '请输入门店编码', trigger: 'blur' }],
  address: [{ required: true, message: '请输入门店地址', trigger: 'blur' }],
  contact: [{ required: true, message: '请输入联系人', trigger: 'blur' }],
  phone: [{ required: true, message: '请输入联系电话', trigger: 'blur' }]
}

async function fetchList() {
  loading.value = true
  try {
    const res = await getStoreList(query)
    list.value = res.list || []
    total.value = res.total || 0
  } catch (e) {
    console.error(e)
  } finally {
    loading.value = false
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

  if (row) {
    Object.assign(storeForm, {
      name: row.name,
      code: row.code,
      address: row.address,
      contact: row.contact,
      phone: row.phone,
      open_time: row.open_time || '08:00',
      close_time: row.close_time || '22:00',
      status: row.status
    })
  } else {
    storeForm.name = ''
    storeForm.code = ''
    storeForm.address = ''
    storeForm.contact = ''
    storeForm.phone = ''
    storeForm.open_time = '08:00'
    storeForm.close_time = '22:00'
    storeForm.status = 1
  }

  dialogVisible.value = true
}

function handleEdit(row) {
  openDialog(row)
}

async function handleStatusChange(row) {
  try {
    await updateStore(row.id, { status: row.status })
    ElMessage.success(row.status === 1 ? '门店已营业' : '门店已停业')
  } catch (e) {
    row.status = row.status === 1 ? 0 : 1
  }
}

function handleDelete(row) {
  ElMessageBox.confirm(`确定删除门店"${row.name}"吗？`, '提示', {
    type: 'warning',
    confirmButtonText: '确定',
    cancelButtonText: '取消'
  }).then(async () => {
    await deleteStore(row.id)
    ElMessage.success('删除成功')
    fetchList()
  })
}

async function handleSubmit() {
  try {
    await storeFormRef.value.validate()
    submitLoading.value = true

    const data = {
      ...storeForm,
      business_hours: `${storeForm.open_time}-${storeForm.close_time}`
    }

    if (isEdit.value) {
      await updateStore(editId.value, data)
      ElMessage.success('更新成功')
    } else {
      await createStore(data)
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
.stores-page {
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
