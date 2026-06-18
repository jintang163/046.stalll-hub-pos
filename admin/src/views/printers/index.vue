<template>
  <div class="printers-page">
    <div class="page-header">
      <h2 class="page-title">打印机管理</h2>
      <div class="header-actions">
        <el-button type="success" @click="openDialog">
          <el-icon><Plus /></el-icon>新增打印机
        </el-button>
      </div>
    </div>

    <div class="card-wrapper">
      <div class="search-bar">
        <el-input
          v-model="query.name"
          placeholder="搜索打印机名称"
          clearable
          style="width: 240px"
          @keyup.enter="fetchList" />
        <el-select v-model="query.type" placeholder="打印机类型" clearable style="width: 160px">
          <el-option label="前台小票" :value="1" />
          <el-option label="后厨小票" :value="2" />
          <el-option label="标签打印机" :value="3" />
          <el-option label="A4打印机" :value="4" />
        </el-select>
        <el-select v-model="query.status" placeholder="连接状态" clearable style="width: 140px">
          <el-option label="在线" :value="1" />
          <el-option label="离线" :value="0" />
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
        <el-table-column prop="name" label="打印机名称" width="180" />
        <el-table-column label="打印机类型" width="140">
          <template #default="{ row }">
            <el-tag :type="getTypeTagType(row.type)">{{ getTypeName(row.type) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="brand" label="品牌" width="120" />
        <el-table-column prop="model" label="型号" width="120" />
        <el-table-column prop="ip_address" label="IP地址" width="160" />
        <el-table-column prop="port" label="端口" width="100" align="center" />
        <el-table-column prop="store_name" label="所属门店" width="140" />
        <el-table-column label="连接状态" width="100" align="center">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'info'">
              <span class="status-dot" :class="{ online: row.status === 1 }"></span>
              {{ row.status === 1 ? '在线' : '离线' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="paper_width" label="纸宽(mm)" width="120" align="center" />
        <el-table-column label="打印类型" width="140">
          <template #default="{ row }">
            <el-tag v-if="row.print_order" type="primary" size="small">订单</el-tag>
            <el-tag v-if="row.print_receipt" type="success" size="small" style="margin-left: 4px;">小票</el-tag>
            <el-tag v-if="row.print_label" type="warning" size="small" style="margin-left: 4px;">标签</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="last_print_time" label="最后打印" width="160" />
        <el-table-column label="操作" width="260" fixed="right">
          <template #default="{ row }">
            <el-button type="success" link size="small" @click="handleTestPrint(row)" :loading="testLoadingId === row.id">测试打印</el-button>
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
      :title="isEdit ? '编辑打印机' : '新增打印机'"
      width="700px"
      :close-on-click-modal="false">
      <el-form
        ref="printerFormRef"
        :model="printerForm"
        :rules="printerRules"
        label-width="120px">
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="打印机名称" prop="name">
              <el-input v-model="printerForm.name" placeholder="请输入打印机名称" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="打印机类型" prop="type">
              <el-select v-model="printerForm.type" placeholder="请选择类型" style="width: 100%">
                <el-option label="前台小票" :value="1" />
                <el-option label="后厨小票" :value="2" />
                <el-option label="标签打印机" :value="3" />
                <el-option label="A4打印机" :value="4" />
              </el-select>
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="品牌">
              <el-input v-model="printerForm.brand" placeholder="请输入品牌" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="型号">
              <el-input v-model="printerForm.model" placeholder="请输入型号" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="20">
          <el-col :span="16">
            <el-form-item label="IP地址" prop="ip_address">
              <el-input v-model="printerForm.ip_address" placeholder="请输入IP地址，如：192.168.1.100" />
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item label="端口" prop="port">
              <el-input-number v-model="printerForm.port" :min="1" :max="65535" style="width: 100%" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="所属门店" prop="store_id">
              <el-select v-model="printerForm.store_id" placeholder="请选择门店" style="width: 100%">
                <el-option v-for="store in storeList" :key="store.id" :label="store.name" :value="store.id" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="纸张宽度" prop="paper_width">
              <el-select v-model="printerForm.paper_width" placeholder="请选择纸张宽度" style="width: 100%">
                <el-option label="58mm" :value="58" />
                <el-option label="80mm" :value="80" />
                <el-option label="110mm" :value="110" />
                <el-option label="A4" :value="210" />
              </el-select>
            </el-form-item>
          </el-col>
        </el-row>
        <el-form-item label="打印内容">
          <div class="checkbox-group">
            <el-checkbox v-model="printerForm.print_order">打印订单</el-checkbox>
            <el-checkbox v-model="printerForm.print_receipt">打印小票</el-checkbox>
            <el-checkbox v-model="printerForm.print_label">打印标签</el-checkbox>
          </div>
        </el-form-item>
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="打印份数">
              <el-input-number v-model="printerForm.copies" :min="1" :max="10" style="width: 100%" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="编码格式">
              <el-select v-model="printerForm.encoding" placeholder="请选择编码格式" style="width: 100%">
                <el-option label="UTF-8" value="UTF-8" />
                <el-option label="GBK" value="GBK" />
                <el-option label="GB2312" value="GB2312" />
              </el-select>
            </el-form-item>
          </el-col>
        </el-row>
        <el-form-item label="打印机状态">
          <el-switch v-model="printerForm.status" :active-value="1" :inactive-value="0" active-text="启用" inactive-text="禁用" />
        </el-form-item>
        <el-form-item label="备注">
          <el-input
            v-model="printerForm.remark"
            type="textarea"
            :rows="2"
            placeholder="请输入备注" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitLoading" @click="handleTestPrint(printerForm)" v-if="isEdit">测试打印</el-button>
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
  getPrinterList,
  createPrinter,
  updatePrinter,
  deletePrinter,
  testPrint
} from '@/api/printers'

const loading = ref(false)
const submitLoading = ref(false)
const testLoadingId = ref(null)
const list = ref([])
const total = ref(0)

const storeList = ref([
  { id: 1, name: '总店' },
  { id: 2, name: '分店A' },
  { id: 3, name: '分店B' }
])

const query = reactive({
  name: '',
  type: null,
  status: null,
  page: 1,
  page_size: 10
})

const dialogVisible = ref(false)
const isEdit = ref(false)
const editId = ref(null)
const printerFormRef = ref()

const printerForm = reactive({
  name: '',
  type: 1,
  brand: '',
  model: '',
  ip_address: '',
  port: 9100,
  store_id: 1,
  paper_width: 80,
  print_order: true,
  print_receipt: true,
  print_label: false,
  copies: 1,
  encoding: 'UTF-8',
  status: 1,
  remark: ''
})

const printerRules = {
  name: [{ required: true, message: '请输入打印机名称', trigger: 'blur' }],
  type: [{ required: true, message: '请选择打印机类型', trigger: 'change' }],
  ip_address: [{ required: true, message: '请输入IP地址', trigger: 'blur' }],
  port: [{ required: true, message: '请输入端口', trigger: 'blur' }],
  store_id: [{ required: true, message: '请选择门店', trigger: 'change' }],
  paper_width: [{ required: true, message: '请选择纸张宽度', trigger: 'change' }]
}

const typeMap = {
  1: { name: '前台小票', type: 'primary' },
  2: { name: '后厨小票', type: 'success' },
  3: { name: '标签打印机', type: 'warning' },
  4: { name: 'A4打印机', type: 'info' }
}

function getTypeName(type) {
  return typeMap[type]?.name || '未知'
}

function getTypeTagType(type) {
  return typeMap[type]?.type || 'info'
}

async function fetchList() {
  loading.value = true
  try {
    const res = await getPrinterList(query)
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
  query.type = null
  query.status = null
  query.page = 1
  fetchList()
}

function openDialog(row = null) {
  isEdit.value = !!row
  editId.value = row?.id || null

  if (row) {
    Object.assign(printerForm, {
      name: row.name,
      type: row.type,
      brand: row.brand || '',
      model: row.model || '',
      ip_address: row.ip_address,
      port: row.port,
      store_id: row.store_id,
      paper_width: row.paper_width,
      print_order: row.print_order || false,
      print_receipt: row.print_receipt || false,
      print_label: row.print_label || false,
      copies: row.copies || 1,
      encoding: row.encoding || 'UTF-8',
      status: row.status,
      remark: row.remark || ''
    })
  } else {
    printerForm.name = ''
    printerForm.type = 1
    printerForm.brand = ''
    printerForm.model = ''
    printerForm.ip_address = ''
    printerForm.port = 9100
    printerForm.store_id = 1
    printerForm.paper_width = 80
    printerForm.print_order = true
    printerForm.print_receipt = true
    printerForm.print_label = false
    printerForm.copies = 1
    printerForm.encoding = 'UTF-8'
    printerForm.status = 1
    printerForm.remark = ''
  }

  dialogVisible.value = true
}

function handleEdit(row) {
  openDialog(row)
}

async function handleTestPrint(row) {
  try {
    testLoadingId.value = row.id
    await testPrint(row.id || editId.value)
    ElMessage.success('测试打印指令已发送')
  } catch (e) {
    console.error(e)
    ElMessage.error('测试打印失败，请检查打印机连接')
  } finally {
    testLoadingId.value = null
  }
}

function handleDelete(row) {
  ElMessageBox.confirm(`确定删除打印机"${row.name}"吗？`, '提示', {
    type: 'warning',
    confirmButtonText: '确定',
    cancelButtonText: '取消'
  }).then(async () => {
    await deletePrinter(row.id)
    ElMessage.success('删除成功')
    fetchList()
  })
}

async function handleSubmit() {
  try {
    await printerFormRef.value.validate()
    submitLoading.value = true

    if (isEdit.value) {
      await updatePrinter(editId.value, printerForm)
      ElMessage.success('更新成功')
    } else {
      await createPrinter(printerForm)
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
.printers-page {
  .header-actions {
    display: flex;
    gap: 12px;
  }

  .status-dot {
    display: inline-block;
    width: 8px;
    height: 8px;
    border-radius: 50%;
    background: #909399;
    margin-right: 6px;

    &.online {
      background: #67c23a;
      animation: pulse 2s infinite;
    }
  }

  @keyframes pulse {
    0% {
      opacity: 1;
    }
    50% {
      opacity: 0.5;
    }
    100% {
      opacity: 1;
    }
  }

  .checkbox-group {
    display: flex;
    gap: 24px;
  }

  .pagination {
    margin-top: 20px;
    display: flex;
    justify-content: flex-end;
  }
}
</style>
