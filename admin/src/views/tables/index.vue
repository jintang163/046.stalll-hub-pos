<template>
  <div class="tables-page">
    <div class="page-header">
      <h2 class="page-title">桌位管理</h2>
      <div class="header-actions">
        <el-button type="success" @click="openDialog">
          <el-icon><Plus /></el-icon>新增桌位
        </el-button>
        <el-button type="primary" @click="openBatchDialog">
          <el-icon><DocumentAdd /></el-icon>批量创建
        </el-button>
        <el-button @click="fetchList">
          <el-icon><Refresh /></el-icon>刷新
        </el-button>
      </div>
    </div>

    <div class="card-wrapper">
      <div class="search-bar">
        <el-select
          v-model="query.store_id"
          placeholder="门店选择"
          clearable
          style="width: 180px"
          @change="handleStoreChange">
          <el-option
            v-for="store in storeList"
            :key="store.id"
            :label="store.name"
            :value="store.id" />
        </el-select>
        <el-select v-model="query.floor" placeholder="楼层筛选" clearable style="width: 140px">
          <el-option
            v-for="floor in floorList"
            :key="floor"
            :label="floor + '楼'"
            :value="floor" />
        </el-select>
        <el-select v-model="query.area_id" placeholder="区域筛选" clearable style="width: 140px">
          <el-option
            v-for="area in areaList"
            :key="area.id"
            :label="area.name"
            :value="area.id" />
        </el-select>
        <el-select v-model="query.status" placeholder="状态筛选" clearable style="width: 140px">
          <el-option label="空闲" :value="0" />
          <el-option label="占用" :value="1" />
          <el-option label="停用" :value="2" />
        </el-select>
        <el-input
          v-model="query.keyword"
          placeholder="搜索桌号/名称"
          clearable
          style="width: 200px"
          @keyup.enter="fetchList" />
        <el-button type="primary" @click="fetchList">
          <el-icon><Search /></el-icon>搜索
        </el-button>
        <el-button @click="resetQuery">
          <el-icon><Refresh /></el-icon>重置
        </el-button>
      </div>

      <el-table :data="list" v-loading="loading">
        <el-table-column prop="table_no" label="桌号" width="100" />
        <el-table-column prop="name" label="名称" min-width="140" />
        <el-table-column label="类型" width="100">
          <template #default="{ row }">
            <el-tag :type="getTypeTagType(row.type)">{{ getTypeName(row.type) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="capacity" label="容纳人数" width="100" align="center" />
        <el-table-column prop="floor" label="楼层" width="80" align="center">
          <template #default="{ row }">{{ row.floor }}楼</template>
        </el-table-column>
        <el-table-column prop="area_name" label="区域" width="120" />
        <el-table-column label="二维码" width="100" align="center">
          <template #default="{ row }">
            <el-button
              v-if="row.qr_code"
              type="primary"
              link
              size="small"
              @click="previewQRCode(row)">
              查看
            </el-button>
            <span v-else style="color: #c0c4cc;">未生成</span>
          </template>
        </el-table-column>
        <el-table-column label="当前状态" width="180" align="center">
          <template #default="{ row }">
            <div class="status-wrapper">
              <el-tag :type="getStatusTagType(row.status)" effect="dark">
                {{ getStatusName(row.status) }}
              </el-tag>
              <div v-if="row.status === 1 && row.occupied_duration" class="occupied-time">
                已消费 {{ formatDuration(row.occupied_duration) }}
              </div>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="创建时间" width="160" />
        <el-table-column label="操作" width="300" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="handleView(row)">查看详情</el-button>
            <el-button type="warning" link size="small" @click="handleGenerateQR(row)">生成二维码</el-button>
            <el-button type="success" link size="small" @click="handleEdit(row)">编辑</el-button>
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
      :title="isEdit ? '编辑桌位' : '新增桌位'"
      width="600px"
      :close-on-click-modal="false">
      <el-form
        ref="tableFormRef"
        :model="tableForm"
        :rules="tableRules"
        label-width="100px">
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="桌号" prop="table_no">
              <el-input v-model="tableForm.table_no" placeholder="请输入桌号" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="名称" prop="name">
              <el-input v-model="tableForm.name" placeholder="请输入桌位名称" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="类型" prop="type">
              <el-select v-model="tableForm.type" placeholder="请选择桌位类型" style="width: 100%">
                <el-option label="普通桌" value="normal" />
                <el-option label="卡座" value="booth" />
                <el-option label="圆桌" value="round" />
                <el-option label="方桌" value="square" />
                <el-option label="包厢" value="private" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="容纳人数" prop="capacity">
              <el-input-number v-model="tableForm.capacity" :min="1" :max="50" style="width: 100%" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="门店" prop="store_id">
              <el-select v-model="tableForm.store_id" placeholder="请选择门店" style="width: 100%">
                <el-option
                  v-for="store in storeList"
                  :key="store.id"
                  :label="store.name"
                  :value="store.id" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="楼层" prop="floor">
              <el-input-number v-model="tableForm.floor" :min="1" :max="20" style="width: 100%" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-form-item label="区域" prop="area_id">
          <el-select v-model="tableForm.area_id" placeholder="请选择区域" style="width: 100%">
            <el-option
              v-for="area in areaList"
              :key="area.id"
              :label="area.name"
              :value="area.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="状态">
          <el-radio-group v-model="tableForm.status">
            <el-radio :value="0">空闲</el-radio>
            <el-radio :value="2">停用</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="备注">
          <el-input
            v-model="tableForm.remark"
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

    <el-dialog
      v-model="batchDialogVisible"
      title="批量创建桌位"
      width="600px"
      :close-on-click-modal="false">
      <el-form
        ref="batchFormRef"
        :model="batchForm"
        :rules="batchRules"
        label-width="100px">
        <el-form-item label="门店" prop="store_id">
          <el-select v-model="batchForm.store_id" placeholder="请选择门店" style="width: 100%">
            <el-option
              v-for="store in storeList"
              :key="store.id"
              :label="store.name"
              :value="store.id" />
          </el-select>
        </el-form-item>
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="创建数量" prop="quantity">
              <el-input-number v-model="batchForm.quantity" :min="1" :max="100" style="width: 100%" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="桌号前缀" prop="prefix">
              <el-input v-model="batchForm.prefix" placeholder="如：A、B、T" maxlength="5" show-word-limit />
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="起始编号" prop="start_no">
              <el-input-number v-model="batchForm.start_no" :min="1" style="width: 100%" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="容纳人数" prop="capacity">
              <el-input-number v-model="batchForm.capacity" :min="1" :max="50" style="width: 100%" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="楼层" prop="floor">
              <el-input-number v-model="batchForm.floor" :min="1" :max="20" style="width: 100%" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="区域" prop="area_id">
              <el-select v-model="batchForm.area_id" placeholder="请选择区域" style="width: 100%">
                <el-option
                  v-for="area in areaList"
                  :key="area.id"
                  :label="area.name"
                  :value="area.id" />
              </el-select>
            </el-form-item>
          </el-col>
        </el-row>
        <el-form-item label="桌位类型" prop="type">
          <el-select v-model="batchForm.type" placeholder="请选择桌位类型" style="width: 100%">
            <el-option label="普通桌" value="normal" />
            <el-option label="卡座" value="booth" />
            <el-option label="圆桌" value="round" />
            <el-option label="方桌" value="square" />
            <el-option label="包厢" value="private" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <div class="batch-preview">
            <span class="preview-label">预览：</span>
            <span class="preview-content">{{ getBatchPreview() }}</span>
          </div>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="batchDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="batchSubmitLoading" @click="handleBatchSubmit">确定创建</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="detailDialogVisible" title="桌位详情" width="500px">
      <div v-if="currentTable" class="detail-content">
        <div class="detail-item">
          <span class="label">桌号：</span>
          <span class="value">{{ currentTable.table_no }}</span>
        </div>
        <div class="detail-item">
          <span class="label">名称：</span>
          <span class="value">{{ currentTable.name }}</span>
        </div>
        <div class="detail-item">
          <span class="label">类型：</span>
          <el-tag :type="getTypeTagType(currentTable.type)">{{ getTypeName(currentTable.type) }}</el-tag>
        </div>
        <div class="detail-item">
          <span class="label">容纳人数：</span>
          <span class="value">{{ currentTable.capacity }}人</span>
        </div>
        <div class="detail-item">
          <span class="label">门店：</span>
          <span class="value">{{ getStoreName(currentTable.store_id) }}</span>
        </div>
        <div class="detail-item">
          <span class="label">楼层：</span>
          <span class="value">{{ currentTable.floor }}楼</span>
        </div>
        <div class="detail-item">
          <span class="label">区域：</span>
          <span class="value">{{ currentTable.area_name || '-' }}</span>
        </div>
        <div class="detail-item">
          <span class="label">状态：</span>
          <el-tag :type="getStatusTagType(currentTable.status)" effect="dark">
            {{ getStatusName(currentTable.status) }}
          </el-tag>
        </div>
        <div v-if="currentTable.status === 1 && currentTable.occupied_duration" class="detail-item">
          <span class="label">消费时长：</span>
          <span class="value occupied">{{ formatDuration(currentTable.occupied_duration) }}</span>
        </div>
        <div v-if="currentTable.remark" class="detail-item">
          <span class="label">备注：</span>
          <span class="value">{{ currentTable.remark }}</span>
        </div>
        <div class="detail-item">
          <span class="label">创建时间：</span>
          <span class="value">{{ currentTable.created_at }}</span>
        </div>
      </div>
    </el-dialog>

    <el-dialog v-model="qrCodeDialogVisible" title="桌位二维码" width="400px">
      <div class="qrcode-wrapper" v-if="currentQRCode">
        <img :src="currentQRCode" alt="桌位二维码" class="qrcode-image" />
        <p class="qrcode-tip">微信扫码即可点餐</p>
      </div>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Refresh, Search, DocumentAdd } from '@element-plus/icons-vue'
import { tableApi, tableAreaApi } from '@/api/tables'
import { getStoreList } from '@/api/stores'

const loading = ref(false)
const submitLoading = ref(false)
const batchSubmitLoading = ref(false)
const list = ref([])
const total = ref(0)
const storeList = ref([])
const areaList = ref([])
const floorList = ref([1, 2, 3, 4, 5])
const currentTable = ref(null)
const currentQRCode = ref('')

const query = reactive({
  store_id: null,
  floor: null,
  area_id: null,
  status: null,
  keyword: '',
  page: 1,
  page_size: 10
})

const dialogVisible = ref(false)
const batchDialogVisible = ref(false)
const detailDialogVisible = ref(false)
const qrCodeDialogVisible = ref(false)
const isEdit = ref(false)
const editId = ref(null)
const tableFormRef = ref()
const batchFormRef = ref()

const tableForm = reactive({
  table_no: '',
  name: '',
  type: 'normal',
  capacity: 4,
  store_id: null,
  floor: 1,
  area_id: null,
  status: 0,
  remark: ''
})

const batchForm = reactive({
  store_id: null,
  quantity: 10,
  prefix: '',
  start_no: 1,
  capacity: 4,
  floor: 1,
  area_id: null,
  type: 'normal'
})

const tableRules = {
  table_no: [{ required: true, message: '请输入桌号', trigger: 'blur' }],
  name: [{ required: true, message: '请输入桌位名称', trigger: 'blur' }],
  type: [{ required: true, message: '请选择桌位类型', trigger: 'change' }],
  capacity: [{ required: true, message: '请输入容纳人数', trigger: 'blur' }],
  store_id: [{ required: true, message: '请选择门店', trigger: 'change' }],
  floor: [{ required: true, message: '请输入楼层', trigger: 'blur' }],
  area_id: [{ required: true, message: '请选择区域', trigger: 'change' }]
}

const batchRules = {
  store_id: [{ required: true, message: '请选择门店', trigger: 'change' }],
  quantity: [{ required: true, message: '请输入创建数量', trigger: 'blur' }],
  prefix: [{ required: true, message: '请输入桌号前缀', trigger: 'blur' }],
  start_no: [{ required: true, message: '请输入起始编号', trigger: 'blur' }],
  capacity: [{ required: true, message: '请输入容纳人数', trigger: 'blur' }],
  floor: [{ required: true, message: '请输入楼层', trigger: 'blur' }],
  area_id: [{ required: true, message: '请选择区域', trigger: 'change' }],
  type: [{ required: true, message: '请选择桌位类型', trigger: 'change' }]
}

const typeMap = {
  normal: { name: '普通桌', type: 'info' },
  booth: { name: '卡座', type: '' },
  round: { name: '圆桌', type: 'warning' },
  square: { name: '方桌', type: 'primary' },
  private: { name: '包厢', type: 'danger' }
}

const statusMap = {
  0: { name: '空闲', type: 'success' },
  1: { name: '占用', type: 'warning' },
  2: { name: '停用', type: 'info' }
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

function formatDuration(minutes) {
  if (!minutes) return '-'
  const hours = Math.floor(minutes / 60)
  const mins = minutes % 60
  if (hours > 0) {
    return `${hours}小时${mins}分钟`
  }
  return `${mins}分钟`
}

function getStoreName(storeId) {
  const store = storeList.value.find(s => s.id === storeId)
  return store?.name || '-'
}

function getBatchPreview() {
  const { prefix, start_no, quantity } = batchForm
  if (!prefix) return '请输入前缀'
  const start = start_no
  const end = start_no + quantity - 1
  return `${prefix}${start} ~ ${prefix}${end}，共 ${quantity} 个桌位`
}

async function fetchStores() {
  try {
    const res = await getStoreList({ page: 1, page_size: 100 })
    storeList.value = res.list || []
    if (storeList.value.length > 0 && !tableForm.store_id) {
      tableForm.store_id = storeList.value[0].id
      batchForm.store_id = storeList.value[0].id
    }
  } catch (e) {
    console.error(e)
  }
}

async function fetchAreas() {
  try {
    const params = {}
    if (query.store_id) {
      params.store_id = query.store_id
    } else if (storeList.value.length > 0) {
      params.store_id = storeList.value[0].id
    }
    const res = await tableAreaApi.list(params)
    areaList.value = res.list || []
  } catch (e) {
    console.error(e)
  }
}

async function fetchList() {
  loading.value = true
  try {
    const res = await tableApi.list(query)
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
  query.floor = null
  query.area_id = null
  query.status = null
  query.keyword = ''
  query.page = 1
  fetchList()
}

function handleStoreChange() {
  query.area_id = null
  tableForm.area_id = null
  batchForm.area_id = null
  fetchAreas()
}

function openDialog(row = null) {
  isEdit.value = !!row
  editId.value = row?.id || null

  if (row) {
    Object.assign(tableForm, {
      table_no: row.table_no,
      name: row.name,
      type: row.type,
      capacity: row.capacity,
      store_id: row.store_id,
      floor: row.floor,
      area_id: row.area_id,
      status: row.status === 1 ? 0 : row.status,
      remark: row.remark || ''
    })
  } else {
    tableForm.table_no = ''
    tableForm.name = ''
    tableForm.type = 'normal'
    tableForm.capacity = 4
    tableForm.floor = 1
    tableForm.area_id = areaList.value[0]?.id || null
    tableForm.status = 0
    tableForm.remark = ''
  }

  dialogVisible.value = true
}

function openBatchDialog() {
  batchForm.store_id = storeList.value[0]?.id || null
  batchForm.quantity = 10
  batchForm.prefix = ''
  batchForm.start_no = 1
  batchForm.capacity = 4
  batchForm.floor = 1
  batchForm.area_id = areaList.value[0]?.id || null
  batchForm.type = 'normal'
  batchDialogVisible.value = true
}

function handleEdit(row) {
  openDialog(row)
}

function handleView(row) {
  currentTable.value = row
  detailDialogVisible.value = true
}

function previewQRCode(row) {
  currentQRCode.value = row.qr_code
  qrCodeDialogVisible.value = true
}

async function handleGenerateQR(row) {
  try {
    const res = await tableApi.generateQRCode(row.id)
    ElMessage.success('二维码生成成功')
    row.qr_code = res.qr_code
    currentQRCode.value = res.qr_code
    qrCodeDialogVisible.value = true
  } catch (e) {
    console.error(e)
  }
}

function handleDelete(row) {
  ElMessageBox.confirm(`确定删除桌位"${row.name}"吗？`, '提示', {
    type: 'warning',
    confirmButtonText: '确定',
    cancelButtonText: '取消'
  }).then(async () => {
    await tableApi.delete(row.id)
    ElMessage.success('删除成功')
    fetchList()
  })
}

async function handleSubmit() {
  try {
    await tableFormRef.value.validate()
    submitLoading.value = true

    if (isEdit.value) {
      await tableApi.update(editId.value, tableForm)
      ElMessage.success('更新成功')
    } else {
      await tableApi.create(tableForm)
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

async function handleBatchSubmit() {
  try {
    await batchFormRef.value.validate()
    batchSubmitLoading.value = true

    await tableApi.batchCreate(batchForm)
    ElMessage.success(`成功创建 ${batchForm.quantity} 个桌位`)

    batchDialogVisible.value = false
    fetchList()
  } catch (e) {
    console.error(e)
  } finally {
    batchSubmitLoading.value = false
  }
}

onMounted(() => {
  fetchStores().then(() => {
    fetchAreas()
  })
  fetchList()
})
</script>

<style scoped lang="scss">
.tables-page {
  .header-actions {
    display: flex;
    gap: 12px;
  }

  .search-bar {
    display: flex;
    flex-wrap: wrap;
    gap: 12px;
    margin-bottom: 20px;
  }

  .status-wrapper {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 4px;

    .occupied-time {
      font-size: 12px;
      color: #e6a23c;
    }
  }

  .pagination {
    margin-top: 20px;
    display: flex;
    justify-content: flex-end;
  }

  .batch-preview {
    display: flex;
    align-items: center;
    padding: 12px 16px;
    background: #f5f7fa;
    border-radius: 4px;

    .preview-label {
      color: #909399;
      margin-right: 8px;
    }

    .preview-content {
      color: #409eff;
      font-weight: 600;
    }
  }

  .detail-content {
    display: flex;
    flex-direction: column;
    gap: 16px;

    .detail-item {
      display: flex;
      align-items: center;

      .label {
        width: 100px;
        color: #909399;
        flex-shrink: 0;
      }

      .value {
        color: #303133;

        &.occupied {
          color: #e6a23c;
          font-weight: 600;
        }
      }
    }
  }

  .qrcode-wrapper {
    display: flex;
    flex-direction: column;
    align-items: center;
    padding: 20px;

    .qrcode-image {
      width: 240px;
      height: 240px;
      border: 1px solid #ebeef5;
      padding: 10px;
      border-radius: 8px;
      margin-bottom: 16px;
    }

    .qrcode-tip {
      color: #909399;
      font-size: 14px;
      margin: 0;
    }
  }
}
</style>
